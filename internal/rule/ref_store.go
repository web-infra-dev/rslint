package rule

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/binder"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// RefStore is the single owner of per-file reference, binding-declaration, and
// import-provenance facts for one linted source file. Its reference index maps
// a declaration symbol to every explicit identifier that references it, in
// source order.
//
// References are resolved with the binder's NameResolver — the same scope walk
// the checker performs for identifiers. Resolve tries this scope walk first;
// when it can't place an identifier and a TypeChecker was supplied to
// NewRefStore, it falls back to the checker. This is usually needed for
// cross-file, .d.ts, and standard-library globals; export specifiers also need
// it when a declaration-space-specific lookup continues past a same-named
// local binding. Without a checker, that fallback is a no-op and Resolve
// returns nil for those identifiers, the same as if they didn't resolve at
// all.
//
// References resolves symbols the same way internally, so it can return
// identifiers referencing a symbol obtained from the checker fallback too; for
// a symbol declared in this file the binder scope walk already answers for
// every identifier sharing its name. Global-script top-level symbols and raw
// declarations inside an external module's `declare global` augmentation cost
// one checker call when queried to reconcile their merged identity (see
// mergedSymbol and globalAugmentationSymbol).
//
// The store deals in raw binder symbols: query it with the symbol attached to
// a declaration node (node.Symbol()), not with checker.GetSymbolAtLocation
// results, which may be checker-merged and compare unequal — except for
// symbols Resolve obtained through the checker fallback itself, and for a
// global script top-level / `declare global` declarations, whose raw and
// merged identities the store reconciles.
//
// Collection is demand-driven. Existing callers retain the lazy full-file
// prepass on first complete query. Rules that declare and request RefNeeds can
// instead have the linter collect the same neutral facts during its ordinary
// traversal. A complete query made before that traversal finishes falls back
// to a full prepass and never exposes partial data. Name resolution remains
// lazy per queried symbol name, so asking about `x` does not resolve unrelated
// identifiers. A RefStore belongs to one lintFile invocation and is therefore
// used serially by that file's rule initializers and listeners.
//
// Every node passed to a node-to-fact or node-to-symbol method must belong to
// this store's source file. References deliberately accepts symbols originating
// outside the file because checker fallback can resolve ambient/cross-file
// declarations. The hot query path does not repeat an ancestor walk merely to
// validate this file-ownership precondition.
type RefStore struct {
	sourceFile *ast.SourceFile
	resolver   binder.NameResolver
	tc         *checker.Checker

	requestsOpen bool
	state        refCollectionState
	requested    RefNeeds
	streaming    RefNeeds
	built        RefNeeds
	// candidates maps identifier text to the reference-position identifiers
	// still awaiting resolution; entries move into refs on first query.
	candidates map[string][]*ast.Node
	refs       map[*ast.Symbol][]*ast.Node
	// merged maps a raw binder symbol onto the checker-merged symbol that
	// refs is keyed by for it; see mergedSymbol.
	merged map[*ast.Symbol]*ast.Symbol
	// facts is allocated only when a declaration/import candidate is actually
	// observed. Keeping these new collections behind one pointer preserves the
	// zero-consumer RefStore allocation class and the existing References maps.
	facts *refFactsIndex
}

type refCollectionState uint8

const (
	refCollectionCold refCollectionState = iota
	refCollectionStreaming
	refCollectionPrepass
	refCollectionCompleteByStream
	refCollectionCompleteByPrepass
)

type refFactsIndex struct {
	declarationNames  []*ast.Node
	importNames       []*ast.Node
	materialized      *materializedRefFacts
	materializedNeeds RefNeeds
}

type materializedRefFacts struct {
	declarations []BindingDeclarationFacts
	imports      []ImportBinding
}

// NewRefStore creates the shared RefStore facts for one source file. options must
// be the file's program options (the resolver consults script target and
// module settings during scope walks). tc is the file's TypeChecker, used as
// a fallback by Resolve for identifiers the binder scope walk can't place,
// and by References when asked about a symbol that fallback produced; nil
// disables both (Resolve then never falls back, and References behaves as if
// every such symbol were never queried). Nodes passed to its node-to-fact or
// node-to-symbol methods must belong to sourceFile.
func NewRefStore(sourceFile *ast.SourceFile, options *core.CompilerOptions, tc *checker.Checker) *RefStore {
	return newRefStore(sourceFile, options, tc)
}

// RefCollector is the linter-only traversal control plane for a RefStore. It
// is intentionally not reachable from RuleContext: rules can request and
// query facts, but cannot seal an in-progress collection and accidentally
// publish partial data as complete. The collector owns no analysis data; its
// one pointer refers to the same RefStore rules query.
type RefCollector struct {
	store *RefStore
}

// RefObserver is the allocation-free Identifier observation token returned to
// the linter by RefCollector.Start. Its fields are private so rules cannot
// manufacture an observer for the RefStore held by RuleContext.
type RefObserver struct {
	store *RefStore
	needs RefNeeds
}

// Active reports whether this file requested an incremental collection.
func (observer RefObserver) Active() bool {
	return observer.store != nil
}

// Observe records one Identifier dispatched by the linter. After an early
// complete-query fallback, the state guard makes all remaining calls no-ops.
func (observer RefObserver) Observe(node *ast.Node) {
	if observer.store.state != refCollectionStreaming {
		return
	}
	s := observer.store
	switch observer.needs {
	case RefNeedReferences:
		s.collectReferenceIdentifier(node)
	case RefNeedBindingDeclarations:
		s.collectDeclarationIdentifier(node)
	case RefNeedImportBindings:
		s.collectImportIdentifier(node)
	case RefNeedReferences | RefNeedBindingDeclarations:
		s.collectReferenceIdentifier(node)
		s.collectDeclarationIdentifier(node)
	case RefNeedReferences | RefNeedImportBindings:
		s.collectReferenceIdentifier(node)
		s.collectImportIdentifier(node)
	case RefNeedBindingDeclarations | RefNeedImportBindings:
		s.collectDeclarationIdentifier(node)
		s.collectImportIdentifier(node)
	case RefNeedsAll:
		s.collectReferenceIdentifier(node)
		s.collectDeclarationIdentifier(node)
		s.collectImportIdentifier(node)
	}
}

// NewRefStoreWithCollector creates a RefStore together with the traversal
// controller the linter must retain privately for that store.
func NewRefStoreWithCollector(
	sourceFile *ast.SourceFile,
	options *core.CompilerOptions,
	tc *checker.Checker,
) (*RefStore, RefCollector) {
	store := newRefStore(sourceFile, options, tc)
	return store, RefCollector{store: store}
}

func newRefStore(sourceFile *ast.SourceFile, options *core.CompilerOptions, tc *checker.Checker) *RefStore {
	resolver := binder.NameResolver{CompilerOptions: options}
	if ast.IsGlobalSourceFile(sourceFile.AsNode()) {
		// A script file's own top-level locals are never consulted by the
		// scope walk (they're conceptually merged into the global symbol
		// table), so the resolver must be handed this file's locals as its
		// globals table or a top-level `var`/function declaration never
		// resolves.
		resolver.Globals = sourceFile.Locals
	}
	return &RefStore{
		sourceFile:   sourceFile,
		resolver:     resolver,
		tc:           tc,
		requestsOpen: true,
	}
}

// References returns every identifier in the file that references sym, in
// source order. Declaration names are not references: the `a` of `var a = 1`
// is excluded while the `a` of a later `a = 2` is included. Treat the
// returned slice as read-only.
func (s *RefStore) References(sym *ast.Symbol) []*ast.Node {
	if s == nil || sym == nil {
		return nil
	}
	s.ensureBuilt(RefNeedReferences)
	s.resolvePending(sym.Name)
	// A symbol's name is not always the name its references are written with:
	// `export default function foo() {}` binds the declaration to an export
	// symbol named "default", while every reference still spells `foo`. Drain
	// the bucket of each declared name too, or those references stay pending
	// forever.
	for _, decl := range sym.Declarations {
		if name := decl.Name(); name != nil && name.Kind == ast.KindIdentifier && name.Text() != sym.Name {
			s.resolvePending(name.Text())
		}
	}
	if merged, ok := s.merged[sym]; ok {
		sym = merged
	} else if merged := s.globalAugmentationSymbol(sym); merged != sym {
		sym = merged
	}
	return s.refs[sym]
}

// globalAugmentationSymbol reconciles a raw binder symbol declared inside an
// external module's `declare global` block with the checker-merged global
// identity used by references outside that block. Unlike ordinary module-local
// symbols, the binder resolver cannot reach the augmentation export from the
// file's outer lexical scope. Restricting the checker lookup to this syntax
// keeps the normal module References path binder-only.
func (s *RefStore) globalAugmentationSymbol(sym *ast.Symbol) *ast.Symbol {
	if s.tc == nil || !isGlobalAugmentationBinderSymbol(sym) {
		return sym
	}
	var name *ast.Node
	for _, declaration := range sym.Declarations {
		name = declaration.Name()
		if name != nil {
			break
		}
	}
	if name == nil {
		return sym
	}
	merged := s.tc.GetSymbolAtLocation(name)
	if merged == nil || merged == sym || !sharesDeclaration(merged, sym) {
		merged = sym
	}
	if s.merged == nil {
		s.merged = make(map[*ast.Symbol]*ast.Symbol)
	}
	s.merged[sym] = merged
	return merged
}

func isGlobalAugmentationBinderSymbol(sym *ast.Symbol) bool {
	parent := sym.Parent
	if parent == nil || parent.Name != ast.InternalSymbolNameGlobal {
		return false
	}
	for _, declaration := range parent.Declarations {
		if ast.IsGlobalScopeAugmentation(declaration) {
			return true
		}
	}
	return false
}

// resolvePending resolves every not-yet-resolved candidate identifier spelled
// name and files each one under the symbol it references. Candidates the
// binder scope walk can't place fall back to the TypeChecker when one is
// available.
func (s *RefStore) resolvePending(name string) {
	pending, ok := s.candidates[name]
	if !ok {
		return
	}
	delete(s.candidates, name)
	for _, id := range pending {
		meaning := referenceMeaning(id)
		target := s.binderReferenceSymbol(id, meaning)
		if target == nil {
			target = s.checkerReferenceSymbol(id, meaning)
		} else {
			target = s.mergedSymbol(target, id)
		}
		if target != nil {
			if s.refs == nil {
				s.refs = make(map[*ast.Symbol][]*ast.Node)
			}
			s.refs[target] = append(s.refs[target], id)
		}
	}
}

// mergedSymbol returns the identity references to sym must be filed under,
// given that sym was resolved from the identifier at.
//
// The checker merges a global script file's top-level declarations with the
// same-named lib.d.ts and cross-file globals into one symbol object, distinct
// from the file's raw binder symbol. Both identities can reach this index for
// a single name: in `Map; type Alias = Map<string, string>; interface Map<K,
// V> {}` the type reference binds to the file's own symbol, while the value
// reference has no local value meaning to bind to and only resolves through
// the checker fallback. Filing them apart would make References return
// whichever half happens to match the queried identity, so normalize onto the
// checker's identity here and record the mapping for References to follow.
//
// Only a symbol in a global script's file-globals table or inside `declare
// global` can be merged this way, so ordinary module/local symbols skip the
// checker round trip; the exceptional symbols pay it once each.
func (s *RefStore) mergedSymbol(sym *ast.Symbol, at *ast.Node) *ast.Symbol {
	if s.tc == nil {
		return sym
	}
	isGlobalScriptTopLevel := s.resolver.Globals != nil && s.resolver.Globals[sym.Name] == sym
	isGlobalAugmentation := isGlobalAugmentationBinderSymbol(sym)
	if !isGlobalScriptTopLevel && !isGlobalAugmentation {
		return sym
	}
	if merged, ok := s.merged[sym]; ok {
		return merged
	}
	merged := utils.GetReferenceSymbol(at, s.tc)
	if merged == nil || merged == sym || !sharesDeclaration(merged, sym) {
		merged = sym
	}
	if s.merged == nil {
		s.merged = make(map[*ast.Symbol]*ast.Symbol)
	}
	s.merged[sym] = merged
	return merged
}

// sharesDeclaration reports whether a and b have a declaration node in
// common, the mark of one being the merged form of the other.
func sharesDeclaration(a, b *ast.Symbol) bool {
	for _, decl := range b.Declarations {
		for _, other := range a.Declarations {
			if other == decl {
				return true
			}
		}
	}
	return false
}

// Resolve returns the symbol that a reference-position identifier resolves
// to — the forward counterpart to References, for rules that need "what
// declaration does this identifier refer to" rather than "what identifiers
// refer to this declaration." It uses the binder's scope walk first; when
// that can't place the identifier and a TypeChecker was supplied to
// NewRefStore, it falls back to the checker. That fallback is a real
// TypeChecker round-trip, unlike the binder-only common path.
//
// node must belong to this store's source file. Returns nil for identifiers
// that aren't reference positions (declaration
// names, property keys, import bindings, re-export/alias-label names, labels,
// intrinsic JSX tags — see isReferencePosition) and for names that don't
// resolve to any symbol.
func (s *RefStore) Resolve(node *ast.Node) *ast.Symbol {
	if s == nil || node == nil || node.Kind != ast.KindIdentifier || !isReferencePosition(node) {
		return nil
	}
	meaning := referenceMeaning(node)
	if sym := s.binderReferenceSymbol(node, meaning); sym != nil {
		return sym
	}
	return s.checkerReferenceSymbol(node, meaning)
}

// ResolveInFile returns the symbol a reference-position identifier resolves to
// using only the binder's lexical scope walk. Unlike Resolve, it never asks the
// TypeChecker to supply a declaration from lib files, ambient .d.ts files, or
// another source file.
//
// node must belong to this store's source file. Imported bindings and
// declarations authored in this source file are still
// visible because their binder symbols live in the file's own scope graph.
// This distinction is required by ESLint-compatible rules such as no-undef:
// their answer must not change merely because rslint happened to construct a
// TypeChecker for the file.
func (s *RefStore) ResolveInFile(node *ast.Node) *ast.Symbol {
	if s == nil || node == nil || node.Kind != ast.KindIdentifier || !isReferencePosition(node) {
		return nil
	}
	return s.binderReferenceSymbol(node, referenceMeaning(node))
}

// binderReferenceSymbol resolves one reference without checker work. A named
// export nested in a namespace needs one extra guard: the binder exposes that
// export's own alias in the namespace's Exports table. If no local declaration
// matches the requested meaning, NameResolver can otherwise return the alias
// being declared instead of continuing to an outer lexical declaration.
func (s *RefStore) binderReferenceSymbol(node *ast.Node, meaning ast.SymbolFlags) *ast.Symbol {
	target := s.resolver.Resolve(node, node.Text(), meaning, nil, true /*isUse*/, false /*excludeGlobals*/)
	if !isOwnExportAlias(node, target) {
		return target
	}
	outer := outerExportScopeLocation(node)
	if outer == nil {
		return nil
	}
	return s.resolver.Resolve(outer, node.Text(), meaning, nil, true /*isUse*/, false /*excludeGlobals*/)
}

// checkerReferenceSymbol resolves a reference the binder-only scope walk
// could not place. Most identifiers can use GetReferenceSymbol, whose
// GetSymbolAtLocation fallback understands their syntactic context.
//
// Export specifiers are different: GetSymbolAtLocation returns the export
// alias symbol rather than the local binding, and
// GetExportSpecifierLocalTargetSymbol searches all declaration spaces at
// once. The latter is also insufficient for a type-only export: a local
// value-only binding could win that broad lookup even when an outer/global
// type binding is the actual target. ResolveName applies meaning during the
// scope walk, preserving both declaration-space selection and shadowing.
func (s *RefStore) checkerReferenceSymbol(node *ast.Node, meaning ast.SymbolFlags) *ast.Symbol {
	if s.tc == nil {
		return nil
	}
	if node.Parent != nil && node.Parent.Kind == ast.KindExportSpecifier {
		target := s.tc.ResolveName(node.Text(), node, meaning, false /*excludeGlobals*/)
		if !isOwnExportAlias(node, target) {
			return target
		}
		outer := outerExportScopeLocation(node)
		if outer == nil {
			return nil
		}
		return s.tc.ResolveName(node.Text(), outer, meaning, false /*excludeGlobals*/)
	}
	return utils.GetReferenceSymbol(node, s.tc)
}

// isOwnExportAlias identifies the alias symbol declared by node's own export
// specifier. It is a declaration result, not the local target that the
// specifier references.
func isOwnExportAlias(node *ast.Node, symbol *ast.Symbol) bool {
	if node == nil || node.Parent == nil || node.Parent.Kind != ast.KindExportSpecifier || symbol == nil {
		return false
	}
	for _, declaration := range symbol.Declarations {
		if declaration == node.Parent {
			return true
		}
	}
	return false
}

// outerExportScopeLocation returns the first location outside the namespace
// containing a local export specifier. Top-level module exports have no outer
// lexical scope to retry.
func outerExportScopeLocation(node *ast.Node) *ast.Node {
	for current := node.Parent; current != nil; current = current.Parent {
		if current.Kind == ast.KindModuleDeclaration {
			return current.Parent
		}
		if current.Kind == ast.KindSourceFile {
			return nil
		}
	}
	return nil
}

// request records a per-file dynamic collection request. Rule.Run calls are
// accumulated before traversal. A late request cannot install a listener into
// an active registry safely, so it takes the correctness-preserving full-scan
// fallback instead.
func (s *RefStore) request(needs RefNeeds) {
	if s == nil || needs == 0 {
		return
	}
	if s.requestsOpen {
		s.requested |= needs
		return
	}
	s.ensureBuilt(needs)
}

// Start closes the Rule.Run request window and returns the private Identifier
// observer the linter should use for this file. An inactive observer means no
// rule requested a complete-file collection.
func (collector RefCollector) Start() RefObserver {
	s := collector.store
	if s == nil {
		return RefObserver{}
	}
	s.requestsOpen = false
	missing := s.requested &^ s.built
	if missing == 0 {
		return RefObserver{}
	}
	return s.startCollection(missing)
}

// startCollection is the requested-only slow path, split out so the common
// zero-request Start remains small enough for the compiler to inline.
func (s *RefStore) startCollection(missing RefNeeds) RefObserver {
	if s.state == refCollectionPrepass {
		panic("rule: RefStore collection started during a prepass")
	}
	s.clearFeatures(missing)
	s.streaming = missing
	s.state = refCollectionStreaming

	if !missing.IsValid() {
		panic("rule: invalid RefStore collection needs")
	}
	return RefObserver{store: s, needs: missing}
}

// Complete seals incremental collection. Complete-file queries made by
// ListenerOnFileFinalize callbacks can therefore never observe partial state.
func (collector RefCollector) Complete() {
	s := collector.store
	if s == nil {
		return
	}
	s.requestsOpen = false
	if s.state != refCollectionStreaming {
		return
	}
	s.built |= s.streaming
	s.streaming = 0
	s.state = refCollectionCompleteByStream
}

// ensureBuilt guarantees that every requested collection is complete. If an
// incremental traversal is still in progress, partial buckets are cleared in
// place and repopulated from the file root so no query can return a partial or
// later-mutating result.
func (s *RefStore) ensureBuilt(needs RefNeeds) {
	if needs == 0 || s.built&needs == needs {
		return
	}
	if s.state == refCollectionPrepass {
		panic("rule: reentrant RefStore prepass")
	}

	missing := needs &^ s.built
	if s.state == refCollectionStreaming {
		missing |= s.streaming
	}
	s.clearFeatures(missing)
	s.state = refCollectionPrepass
	s.collectFeatures(missing)
	s.built |= missing
	s.streaming &^= missing
	s.state = refCollectionCompleteByPrepass
}

// collectFeatures walks the file once and derives all requested feature sets
// from Identifier nodes. It is used only by the lazy/early-query fallback.
func (s *RefStore) collectFeatures(needs RefNeeds) {
	var visit func(n *ast.Node) bool
	visit = func(n *ast.Node) bool {
		if n.Kind == ast.KindIdentifier {
			s.collectIdentifier(n, needs)
		}
		n.ForEachChild(visit)
		return false
	}
	s.sourceFile.AsNode().ForEachChild(visit)
}

func (s *RefStore) collectIdentifier(node *ast.Node, needs RefNeeds) {
	if needs&RefNeedReferences != 0 {
		s.collectReferenceIdentifier(node)
	}
	if needs&RefNeedBindingDeclarations != 0 {
		s.collectDeclarationIdentifier(node)
	}
	if needs&RefNeedImportBindings != 0 {
		s.collectImportIdentifier(node)
	}
}

func (s *RefStore) collectReferenceIdentifier(node *ast.Node) {
	if !isReferencePosition(node) {
		return
	}
	if s.candidates == nil {
		s.candidates = make(map[string][]*ast.Node)
	}
	text := node.Text()
	s.candidates[text] = append(s.candidates[text], node)
}

func (s *RefStore) collectDeclarationIdentifier(node *ast.Node) {
	if isBindingDeclarationIdentifier(node) {
		facts := s.ensureFacts()
		facts.declarationNames = append(facts.declarationNames, node)
	}
}

func (s *RefStore) collectImportIdentifier(node *ast.Node) {
	if isImportBindingIdentifier(node) {
		facts := s.ensureFacts()
		facts.importNames = append(facts.importNames, node)
	}
}

func isImportBindingIdentifier(node *ast.Node) bool {
	if node.Parent == nil || node.Parent.Name() != node {
		return false
	}
	switch node.Parent.Kind {
	case ast.KindImportClause, ast.KindImportSpecifier, ast.KindNamespaceImport, ast.KindImportEqualsDeclaration:
		return true
	default:
		return false
	}
}

func (s *RefStore) ensureFacts() *refFactsIndex {
	if s.facts == nil {
		s.facts = &refFactsIndex{}
	}
	return s.facts
}

// clearFeatures resets only incomplete collections and retains their backing
// storage for an early-query fallback. Completed feature slices/maps are never
// modified after they may have been returned to callers.
func (s *RefStore) clearFeatures(needs RefNeeds) {
	if needs&RefNeedReferences != 0 {
		if s.candidates != nil {
			for name, candidates := range s.candidates {
				s.candidates[name] = candidates[:0]
			}
		}
		if s.refs != nil {
			clear(s.refs)
		}
		if s.merged != nil {
			clear(s.merged)
		}
	}
	if s.facts == nil {
		return
	}
	if needs&RefNeedBindingDeclarations != 0 {
		s.facts.declarationNames = s.facts.declarationNames[:0]
		if materialized := s.facts.materialized; materialized != nil {
			materialized.declarations = nil
		}
		s.facts.materializedNeeds &^= RefNeedBindingDeclarations
	}
	if needs&RefNeedImportBindings != 0 {
		s.facts.importNames = s.facts.importNames[:0]
		if materialized := s.facts.materialized; materialized != nil {
			materialized.imports = nil
		}
		s.facts.materializedNeeds &^= RefNeedImportBindings
	}
}

// isReferencePosition reports whether an identifier can reference a symbol
// declared elsewhere: not a declaration name, and not a position that names
// something non-local (property names, import bindings, labels, intrinsic JSX
// tags). The one exception is a local named export's real target — see the
// ExportSpecifier case below.
func isReferencePosition(n *ast.Node) bool {
	p := n.Parent
	if n.Text() == "const" && p != nil && p.Parent != nil && ast.IsConstAssertion(p.Parent) {
		// The identifier-shaped type in `expr as const` is a parser sentinel,
		// not a lexical lookup (NameResolver has the same explicit exclusion).
		return false
	}
	if p != nil && p.Kind == ast.KindShorthandPropertyAssignment && p.AsShorthandPropertyAssignment().Name() == n {
		// `{x}` reads x as an expression, and `({x} = obj)` writes to x once
		// the object literal is reinterpreted as an assignment pattern;
		// IsDeclarationName treats this name as a declaration (it also
		// declares the object's property), which would otherwise discard
		// both real uses.
		return true
	}
	if p != nil && p.Kind == ast.KindExportSpecifier {
		// `export { foo }` / `export { foo as bar }` reads its local binding
		// through PropertyName when present (the rename target, `foo`) or
		// Name when it isn't (no rename) — ESLint's scope-manager marks this
		// as a read of that binding. IsDeclarationName treats this node as a
		// declaration either way (Name() equals the node itself when there's
		// no PropertyName, and equals the alias label `bar` when there is),
		// which would otherwise discard the real local reference the same
		// way it would for a shorthand property name above. A re-export
		// (`export { foo } from 'mod'`, aliased or not) references the
		// other module's binding instead, so neither part is a reference
		// there.
		if utils.IsReExportSpecifier(p) {
			return false
		}
		if propertyName := p.PropertyName(); propertyName != nil {
			return propertyName == n
		}
		return true
	}
	if ast.IsDeclarationName(n) {
		return false
	}
	if p == nil {
		return false
	}
	switch p.Kind {
	case ast.KindPropertyAccessExpression:
		return p.AsPropertyAccessExpression().Name() != n
	case ast.KindQualifiedName:
		return p.AsQualifiedName().Right != n
	case ast.KindPropertyAssignment:
		return p.AsPropertyAssignment().Name() != n
	case ast.KindBindingElement:
		return p.AsBindingElement().PropertyName != n
	case ast.KindImportAttribute:
		// Import attribute keys (`type` in `with { type: "json" }`) are
		// syntactic names, not references to same-named variables.
		return p.AsImportAttribute().Name() != n
	case ast.KindImportSpecifier, ast.KindNamespaceImport,
		ast.KindImportClause, ast.KindNamespaceExport:
		// Import bindings and the export half of `export * as NS` resolve
		// through module/alias machinery the checker owns; the current
		// consumers never treat them as references.
		return false
	case ast.KindLabeledStatement, ast.KindBreakStatement, ast.KindContinueStatement:
		// Labels live in their own namespace and never reference variables.
		return false
	case ast.KindMetaProperty:
		// `target`/`meta`/`defer` in `new.target`, `import.meta`, and
		// `import.defer` are syntactic, not identifiers that can reference a
		// variable.
		return false
	case ast.KindJsxNamespacedName:
		// The namespace and name pieces of a namespaced JSX tag or attribute
		// (`<foo:bar attr:name="v" />`) are syntactic, never variable
		// references; the plain identifier tag name case is still handled
		// by IsJsxTagName below.
		return false
	}
	if ast.IsJsxTagName(n) {
		// Intrinsics include lowercase HTML names and custom-element names
		// containing a dash, regardless of their first character.
		return !scanner.IsIntrinsicJsxName(n.Text())
	}
	return true
}

// referenceMeaning mirrors the meaning the checker's
// getSymbolOfNameOrPropertyAccessExpression would resolve this identifier
// with. SymbolFlagsAlias is always included: the plain binder lookup cannot
// resolve alias targets, so import aliases must match on the alias symbol
// itself (which is also what the reference consumers track).
func referenceMeaning(n *ast.Node) ast.SymbolFlags {
	p := n.Parent
	if p != nil && p.Kind == ast.KindTypePredicate {
		// `x is T` — the x names a parameter.
		return ast.SymbolFlagsFunctionScopedVariable
	}
	if p != nil && p.Kind == ast.KindExportAssignment {
		// `export = Foo` and `export default Foo` resolve entity names in all
		// declaration spaces. In particular, TypeScript permits type-only
		// declarations such as an interface to be consumed by `export =`.
		// Mirror the checker's getSymbolOfNameOrPropertyAccessExpression path
		// instead of treating the identifier as an ordinary value expression.
		return ast.SymbolFlagsValue | ast.SymbolFlagsType | ast.SymbolFlagsNamespace | ast.SymbolFlagsAlias
	}
	if p != nil && p.Kind == ast.KindExportSpecifier {
		// A type-only export creates a type reference. It may bind to a type,
		// namespace, or import alias, but never to a value-only declaration.
		// Namespace is separate from Type in the TypeScript binder even though
		// typescript-eslint scope-manager treats namespace definitions as
		// type-capable variables.
		if ast.IsTypeOnlyImportOrExportDeclaration(p) {
			return ast.SymbolFlagsType | ast.SymbolFlagsNamespace | ast.SymbolFlagsAlias
		}
		// A regular local export creates a dual value/type reference, so it
		// can bind in any declaration space.
		return ast.SymbolFlagsValue | ast.SymbolFlagsType | ast.SymbolFlagsNamespace | ast.SymbolFlagsAlias
	}
	// `import a = b.c` — the right-hand side may name a value, type, or
	// namespace.
	entity := n
	for entity.Parent != nil && entity.Parent.Kind == ast.KindQualifiedName {
		entity = entity.Parent
	}
	if entity.Parent != nil && entity.Parent.Kind == ast.KindImportEqualsDeclaration &&
		entity.Parent.AsImportEqualsDeclaration().ModuleReference == entity {
		return ast.SymbolFlagsValue | ast.SymbolFlagsType | ast.SymbolFlagsNamespace | ast.SymbolFlagsAlias
	}
	// Heritage clauses parse dotted names as PropertyAccessExpressions rather
	// than QualifiedNames. In a type-only heritage position, the root of that
	// property-access chain still names a namespace and must skip a same-named
	// value binding. A class `extends` expression is not part of a type node, so
	// it deliberately continues to the value-space branch below.
	if isTypeOnlyPropertyAccessQualifier(n) {
		return ast.SymbolFlagsNamespace | ast.SymbolFlagsAlias
	}
	if ast.IsExpressionNode(n) {
		return ast.SymbolFlagsValue | ast.SymbolFlagsAlias
	}
	// Leftmost qualifier of a non-expression entity name (`A` in `let x: A.B`)
	// names a namespace. Expression and typeof chains were handled above.
	if p != nil && p.Kind == ast.KindQualifiedName && p.AsQualifiedName().Left == n {
		return ast.SymbolFlagsNamespace | ast.SymbolFlagsAlias
	}
	if ast.IsPartOfTypeNode(n) {
		return ast.SymbolFlagsType | ast.SymbolFlagsAlias
	}
	return ast.SymbolFlagsValue | ast.SymbolFlagsAlias
}

// isTypeOnlyPropertyAccessQualifier reports whether n is the lexical root of
// an expression-shaped qualified name used as a type. TypeScript represents
// `N.T` in `class C implements N.T` and `interface I extends N.T` with a
// PropertyAccessExpression, even though N has namespace meaning there.
func isTypeOnlyPropertyAccessQualifier(n *ast.Node) bool {
	entity := n
	for entity.Parent != nil && entity.Parent.Kind == ast.KindPropertyAccessExpression {
		access := entity.Parent.AsPropertyAccessExpression()
		if access == nil || access.Expression != entity {
			break
		}
		entity = entity.Parent
	}
	return entity != n && ast.IsPartOfTypeNode(entity)
}
