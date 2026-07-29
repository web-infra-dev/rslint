package rule

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/binder"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// RefStore owns the per-file identifier-reference index for one linted source
// file: given a symbol declared in the file, it returns every identifier that
// references it, in source order.
//
// References are resolved with the binder's NameResolver — the same scope walk
// the checker performs for identifiers. Resolve tries this scope walk first;
// when it can't place an identifier (a symbol declared outside this file —
// cross-file, .d.ts, and standard-library globals) and a TypeChecker was
// supplied to NewRefStore, it falls back to the checker for that identifier.
// Without a checker, that fallback is a no-op and Resolve returns nil for
// those identifiers, the same as if they didn't resolve at all.
//
// References resolves symbols the same way internally, so it can return
// identifiers referencing a symbol obtained from the checker fallback too; for
// a symbol declared in this file the binder scope walk already answers for
// every identifier sharing its name, the one exception being a global script
// file's top-level symbols, which cost one checker call each to normalize
// onto their merged identity (see mergedSymbol).
//
// The store deals in raw binder symbols: query it with the symbol attached to
// a declaration node (node.Symbol()), not with checker.GetSymbolAtLocation
// results, which may be checker-merged and compare unequal — except for
// symbols Resolve obtained through the checker fallback itself, and for a
// global script file's top-level declarations, whose raw and merged
// identities the store reconciles (see mergedSymbol).
//
// Collection is lazy twice over: the single AST walk that gathers candidate
// identifiers runs on first use, and name resolution runs once per queried
// symbol name, so a rule asking about `x` never pays for resolving unrelated
// identifiers. A RefStore belongs to one lintFile invocation and is therefore
// used serially by that file's rule initializers and listeners.
type RefStore struct {
	sourceFile *ast.SourceFile
	resolver   binder.NameResolver
	tc         *checker.Checker
	walked     bool
	// candidates maps identifier text to the reference-position identifiers
	// still awaiting resolution; entries move into refs on first query.
	candidates map[string][]*ast.Node
	refs       map[*ast.Symbol][]*ast.Node
	// merged maps a raw binder symbol onto the checker-merged symbol that
	// refs is keyed by for it; see mergedSymbol.
	merged map[*ast.Symbol]*ast.Symbol
}

// NewRefStore creates the reference index for one source file. options must
// be the file's program options (the resolver consults script target and
// module settings during scope walks). tc is the file's TypeChecker, used as
// a fallback by Resolve for identifiers the binder scope walk can't place,
// and by References when asked about a symbol that fallback produced; nil
// disables both (Resolve then never falls back, and References behaves as if
// every such symbol were never queried).
func NewRefStore(sourceFile *ast.SourceFile, options *core.CompilerOptions, tc *checker.Checker) *RefStore {
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
		sourceFile: sourceFile,
		resolver:   resolver,
		tc:         tc,
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
	if !s.walked {
		s.walked = true
		s.collectCandidates()
	}
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
	}
	return s.refs[sym]
}

// resolvePending resolves every not-yet-resolved candidate identifier spelled
// name and files each one under the symbol it references. Candidates the
// binder scope walk can't place (declared outside this file) fall back to
// the TypeChecker when one is available.
func (s *RefStore) resolvePending(name string) {
	pending, ok := s.candidates[name]
	if !ok {
		return
	}
	delete(s.candidates, name)
	for _, id := range pending {
		target := s.resolver.Resolve(id, name, referenceMeaning(id), nil, true /*isUse*/, false /*excludeGlobals*/)
		if target == nil {
			if s.tc != nil {
				target = utils.GetReferenceSymbol(id, s.tc)
			}
		} else {
			target = s.mergedSymbol(target, id)
		}
		if target != nil {
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
// Only a symbol the resolver found in the file's own globals table can be
// merged this way, so every symbol of a module file and every nested scope
// skips the checker round trip; the rest pay it once per symbol.
func (s *RefStore) mergedSymbol(sym *ast.Symbol, at *ast.Node) *ast.Symbol {
	if s.tc == nil || s.resolver.Globals == nil || s.resolver.Globals[sym.Name] != sym {
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
// that can't place the identifier — a symbol declared outside this file
// (cross-file, .d.ts, and standard-library globals) — and a TypeChecker was
// supplied to NewRefStore, it falls back to the checker. That fallback is a
// real TypeChecker round-trip, unlike the binder-only common path.
//
// Returns nil for identifiers that aren't reference positions (declaration
// names, property keys, import bindings, re-export/alias-label names, labels,
// intrinsic JSX tags — see isReferencePosition) and for names that don't
// resolve to any symbol.
func (s *RefStore) Resolve(node *ast.Node) *ast.Symbol {
	if s == nil || node == nil || node.Kind != ast.KindIdentifier || !isReferencePosition(node) {
		return nil
	}
	if sym := s.resolver.Resolve(node, node.Text(), referenceMeaning(node), nil, true /*isUse*/, false /*excludeGlobals*/); sym != nil {
		return sym
	}
	if s.tc == nil {
		return nil
	}
	return utils.GetReferenceSymbol(node, s.tc)
}

// collectCandidates walks the file once and buckets by name every identifier
// that occupies a reference position.
func (s *RefStore) collectCandidates() {
	s.candidates = make(map[string][]*ast.Node)
	s.refs = make(map[*ast.Symbol][]*ast.Node)
	var visit func(n *ast.Node) bool
	visit = func(n *ast.Node) bool {
		if n.Kind == ast.KindIdentifier && isReferencePosition(n) {
			text := n.Text()
			s.candidates[text] = append(s.candidates[text], n)
		}
		n.ForEachChild(visit)
		return false
	}
	s.sourceFile.AsNode().ForEachChild(visit)
}

// isReferencePosition reports whether an identifier can reference a symbol
// declared elsewhere: not a declaration name, and not a position that names
// something non-local (property names, import bindings, labels, intrinsic JSX
// tags). The one exception is a local named export's real target — see the
// ExportSpecifier case below.
func isReferencePosition(n *ast.Node) bool {
	p := n.Parent
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
		// Lowercase tag names are JSX intrinsics (`<div>`), not identifier
		// references; uppercase ones reference a component value.
		text := n.Text()
		return len(text) > 0 && (text[0] < 'a' || text[0] > 'z')
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
		// `export { X }` can re-export a value, a type, or a namespace —
		// same declaration-space breadth as `export =` above.
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
