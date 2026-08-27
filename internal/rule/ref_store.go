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
	init       RefStoreInit
	walked     bool
	// candidates maps identifier text to the reference-position identifiers
	// still awaiting resolution; entries move into refs on first query.
	candidates map[string][]*ast.Node
	refs       map[*ast.Symbol][]*ast.Node
	// authoredImports holds the local names bound by the file's own import
	// declarations; nil until a JSDoc-synthesized binding forces the question.
	authoredImports map[string]struct{}
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
// every such symbol were never queried). init supplies only declaration-less
// wrapper bindings and top-level scope facts resolved by the linter.
func NewRefStore(sourceFile *ast.SourceFile, options *core.CompilerOptions, tc *checker.Checker, init RefStoreInit) *RefStore {
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
		init:       init,
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
// that can't place the identifier and a TypeChecker was supplied to
// NewRefStore, it falls back to the checker. That fallback is a real
// TypeChecker round-trip, unlike the binder-only common path.
//
// Returns nil for identifiers that aren't reference positions (declaration
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
// Imported bindings and declarations authored in this source file are still
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

// ResolveInFileWithMeaning is ResolveInFile with an explicit declaration-space
// meaning. It is for consumers whose reference model intentionally differs
// from TypeScript's checker semantics, such as ESLint scope variables that can
// independently be type-capable, value-capable, or both. Like ResolveInFile,
// it never consults the TypeChecker or declarations outside this source file.
func (s *RefStore) ResolveInFileWithMeaning(node *ast.Node, meaning ast.SymbolFlags) *ast.Symbol {
	if s == nil || node == nil || node.Kind != ast.KindIdentifier || !isReferencePosition(node) {
		return nil
	}
	return s.binderReferenceSymbol(node, meaning)
}

// HasImplicitWrapperBinding reports whether the resolved file wrapper supplies
// a declaration-less, non-global binding with name.
func (s *RefStore) HasImplicitWrapperBinding(name string) bool {
	return s != nil && name != "" && s.init.hasImplicitWrapperBinding(name)
}

// IsDefinedInFile reports whether a reference-position identifier is defined
// by the file's lexical scope graph or by a non-global binding supplied by its
// resolved file wrapper. It never consults the TypeChecker and does not
// manufacture an ast.Symbol for an implicit binding.
func (s *RefStore) IsDefinedInFile(node *ast.Node) bool {
	if s == nil || node == nil || node.Kind != ast.KindIdentifier || !isReferencePosition(node) {
		return false
	}
	if s.binderReferenceSymbol(node, referenceMeaning(node)) != nil {
		return true
	}
	return s.HasImplicitWrapperBinding(node.Text())
}

// IsNameDefinedInFile reports whether name is visible in the value scope at
// location through a declaration in this file or a binding supplied by the
// resolved file wrapper. Unlike IsDefinedInFile, location need not itself be a
// reference-position identifier. The TypeChecker is never consulted.
func (s *RefStore) IsNameDefinedInFile(location *ast.Node, name string) bool {
	return s.IsNameDefinedInFileWithMeaning(location, name, ast.SymbolFlagsValue)
}

// IsNameDefinedInFileWithMeaning is IsNameDefinedInFile for an explicit set of
// declaration spaces. Implicit wrapper bindings participate only in value
// lookups.
func (s *RefStore) IsNameDefinedInFileWithMeaning(location *ast.Node, name string, meaning ast.SymbolFlags) bool {
	if s == nil || location == nil || name == "" {
		return false
	}
	if s.resolveName(location, name, meaning) != nil {
		return true
	}
	return meaning&ast.SymbolFlagsValue != 0 && s.HasImplicitWrapperBinding(name)
}

// resolveName runs the binder scope walk for name at location and reconciles
// its departures from the scopes the source text actually spells.
//
// The first is a JavaScript scoping rule NameResolver deliberately leaves out:
// a parameter initializer cannot see a declaration made in the function body.
// TypeScript keeps that resolution so the checker can report "used before its
// declaration" against the right symbol, but scope-shaped consumers need the
// language answer — the parameter environment is the body environment's
// parent, so the name belongs to whatever encloses the function.
//
// The second is the declarations the parser synthesizes from JSDoc tags, which
// join a scope's symbol table without being written there.
//
// The rest are bindings the binder's symbol tables carry that the language
// declares nowhere the walk passes: a namespace re-export, a segment of a
// dotted namespace name, and a class expression's self-reference seen from the
// class's own decorators. Every skip resumes the walk from the scope that held
// the rejected binding, so an outer declaration of the same name is still
// found.
func (s *RefStore) resolveName(location *ast.Node, name string, meaning ast.SymbolFlags) *ast.Symbol {
	for location != nil {
		result := s.resolver.Resolve(location, name, meaning, nil, true /*isUse*/, false /*excludeGlobals*/)
		if result == nil {
			return nil
		}
		if scope := jsdocOnlyScope(result, name); scope != nil && !s.hasAuthoredImportBinding(name) {
			location = scope.Parent
			continue
		}
		if isExportAliasOnly(result) {
			location = outerExportScopeLocation(result.Declarations[0])
			continue
		}
		if isQualifiedNamespaceNameOnly(result) {
			if scope := bindingScope(result, name); scope != nil {
				location = scope.Parent
				continue
			}
		}
		if class := classExpressionNameInOwnDecorator(location, result); class != nil {
			location = class.Parent
			continue
		}
		fn := bodyOnlyParameterBarrier(location, result)
		if fn == nil {
			return result
		}
		// A named function expression binds its own name outside the body, so
		// that binding survives the barrier; everything else resolves from the
		// scope holding the function.
		if own := functionExpressionSelfBinding(fn, name, meaning); own != nil {
			return own
		}
		location = fn.Parent
	}
	return nil
}

// isExportAliasOnly reports whether every declaration of symbol is an export
// specifier or a namespace export. A namespace's `export { X } from 'y'`,
// `export * as X from 'y'`, and bare `export { X }` add X to what the namespace
// exports without declaring it inside — ESLint's scope manager creates no
// variable for them, so the scope walk has to step over the binding and resume
// outside the namespace.
func isExportAliasOnly(symbol *ast.Symbol) bool {
	if len(symbol.Declarations) == 0 {
		return false
	}
	for _, declaration := range symbol.Declarations {
		if declaration.Kind != ast.KindExportSpecifier && declaration.Kind != ast.KindNamespaceExport {
			return false
		}
	}
	return true
}

// isQualifiedNamespaceNameOnly reports whether every declaration of symbol is a
// segment of a dotted namespace name. `namespace N.M {}` opens one scope for
// the whole declaration and declares neither segment in it — ESLint's scope
// manager only creates a variable for a namespace named by a plain identifier
// — so the scope walk has to step over the binding and resume outside.
func isQualifiedNamespaceNameOnly(symbol *ast.Symbol) bool {
	if len(symbol.Declarations) == 0 {
		return false
	}
	for _, declaration := range symbol.Declarations {
		if !utils.IsQualifiedNamespaceSegment(declaration) {
			return false
		}
	}
	return true
}

// bindingScope returns the node whose symbol table binds name to symbol —
// either as a local or, for a namespace member, as an export. It returns nil
// when no enclosing table claims it, leaving the caller to keep the symbol
// rather than guess where the walk should resume.
func bindingScope(symbol *ast.Symbol, name string) *ast.Node {
	if len(symbol.Declarations) == 0 {
		return nil
	}
	for node := symbol.Declarations[0].Parent; node != nil; node = node.Parent {
		if locals := node.Locals(); locals != nil && locals[name] == symbol {
			return node
		}
		if owner := node.Symbol(); owner != nil && owner.Exports != nil && owner.Exports[name] == symbol {
			return node
		}
	}
	return nil
}

// classExpressionNameInOwnDecorator returns the class expression that binds
// result to its own name while location sits in one of that class's
// decorators. A class's decorators are evaluated in the scope holding the
// class, so the name a class expression binds for itself does not reach them,
// but TypeScript's resolver keeps that self-reference visible throughout the
// class node — decorators included.
func classExpressionNameInOwnDecorator(location *ast.Node, result *ast.Symbol) *ast.Node {
	if len(result.Declarations) != 1 || result.Declarations[0].Kind != ast.KindClassExpression {
		return nil
	}
	class := result.Declarations[0]
	for node := location; node != nil; node = node.Parent {
		if node.Kind == ast.KindDecorator && node.Parent == class {
			return class
		}
	}
	return nil
}

// jsdocOnlyScope returns the scope whose symbol table binds name to symbol when
// every declaration of symbol was synthesized from a JSDoc tag. ESLint reads
// that text as a comment and declares nothing, so the binding has to be stepped
// over. It returns nil when the symbol has a declaration the file spells, and
// when no enclosing scope table claims it — leaving that symbol in place rather
// than guessing where the walk should resume.
func jsdocOnlyScope(symbol *ast.Symbol, name string) *ast.Node {
	if len(symbol.Declarations) == 0 || hasAuthoredDeclaration(symbol) {
		return nil
	}
	for node := symbol.Declarations[0].Parent; node != nil; node = node.Parent {
		if locals := node.Locals(); locals != nil && locals[name] == symbol {
			return node
		}
	}
	// JSDoc template parameters on classes are resolved from the class's type
	// parameter environment rather than a Locals table. A reparsed clone is
	// attached directly to the class, so leave the class before looking for an
	// authored outer declaration.
	declaration := symbol.Declarations[0]
	if declaration.Kind == ast.KindTypeParameter && declaration.Parent != nil && ast.IsClassLike(declaration.Parent) {
		return declaration.Parent
	}
	return nil
}

// hasAuthoredImportBinding reports whether an import declaration the file
// spells introduces name. Two aliases cannot share one table slot, so when a
// JSDoc `@import` tag binds a name a real import already binds, the symbol the
// binder keeps carries only one of them — and it can be the synthesized one.
// The authored syntax declares the name either way.
func (s *RefStore) hasAuthoredImportBinding(name string) bool {
	if s.authoredImports == nil {
		s.authoredImports = collectAuthoredImportBindings(s.sourceFile)
	}
	_, ok := s.authoredImports[name]
	return ok
}

// collectAuthoredImportBindings gathers every local name bound by an import
// declaration written in the file. The declarations reparsed from `@import`
// tags carry their own node kind, so they never enter this set.
func collectAuthoredImportBindings(sourceFile *ast.SourceFile) map[string]struct{} {
	bindings := make(map[string]struct{})
	if sourceFile == nil || sourceFile.Statements == nil {
		return bindings
	}
	var visit func(*ast.Node) bool
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindIdentifier && ast.IsDeclarationName(node) {
			bindings[node.Text()] = struct{}{}
		}
		node.ForEachChild(visit)
		return false
	}
	for _, statement := range sourceFile.Statements.Nodes {
		if statement.Kind == ast.KindImportDeclaration ||
			statement.Kind == ast.KindImportEqualsDeclaration {
			statement.ForEachChild(visit)
		}
	}
	return bindings
}

// bodyOnlyParameterBarrier returns the function whose body holds every
// declaration of result while location sits in that function's parameter list,
// which is the shape where the scope walk crossed a boundary the language does
// not. It returns nil when no such function encloses location.
func bodyOnlyParameterBarrier(location *ast.Node, result *ast.Symbol) *ast.Node {
	if len(result.Declarations) == 0 {
		return nil
	}
	for node := location; node != nil; node = node.Parent {
		if node.Kind != ast.KindParameter || node.Parent == nil || !ast.IsFunctionLike(node.Parent) {
			continue
		}
		fn := node.Parent
		if body := fn.Body(); body != nil && declaredOnlyWithin(result, body) {
			return fn
		}
	}
	return nil
}

// declaredOnlyWithin reports whether every declaration of symbol lies inside
// body. A symbol that also has a declaration outside it — a parameter of the
// same name, a type parameter, a merged outer declaration — stays visible from
// the parameter list.
func declaredOnlyWithin(symbol *ast.Symbol, body *ast.Node) bool {
	for _, decl := range symbol.Declarations {
		if decl.Pos() < body.Pos() || decl.End() > body.End() {
			return false
		}
	}
	return true
}

// functionExpressionSelfBinding returns the symbol a named function expression
// binds for its own name. The binder keeps that symbol off the function's
// locals table, so a scope walk restarted outside the function would otherwise
// lose it.
func functionExpressionSelfBinding(fn *ast.Node, name string, meaning ast.SymbolFlags) *ast.Symbol {
	if meaning&ast.SymbolFlagsFunction == 0 || !ast.IsFunctionExpression(fn) {
		return nil
	}
	if fnName := fn.Name(); fnName != nil && fnName.Text() == name {
		return fn.Symbol()
	}
	return nil
}

// IsGlobalNameReference reports whether name at location refers to an
// environment global rather than to a declaration in this file, mirroring
// ESLint's sourceCode.isGlobalReference for the given declaration spaces.
//
// A global program scope needs its own answer: ESLint keeps that file's
// top-level declarations and the configured globals in one global-scope
// variable, so any definition of the name there — a type-only `interface` or
// `type` included — clears the global reference. Inner scopes and module
// program scopes hold separate variables, and a value reference only resolves
// to one declared in a requested space, so a type-only declaration there
// leaves it global.
func (s *RefStore) IsGlobalNameReference(location *ast.Node, name string, meaning ast.SymbolFlags) bool {
	if s == nil || location == nil || name == "" {
		return false
	}
	if !s.HasNonGlobalProgramScope() && s.hasAuthoredProgramDefinition(name) {
		return false
	}
	// When a JSDoc import and an authored import bind the same local name,
	// ts-go can retain only the synthesized declaration on the resolved symbol.
	// The source scan is the authoritative evidence that the local exists.
	if s.hasAuthoredImportBinding(name) {
		return false
	}
	if hasAuthoredDeclaration(s.resolveName(location, name, meaning)) {
		return false
	}
	return meaning&ast.SymbolFlagsValue == 0 || !s.HasImplicitWrapperBinding(name)
}

// hasAuthoredProgramDefinition reports whether the file spells a definition
// that typescript-eslint places in its global Program scope. Dotted namespace
// segments are present in ts-go's file locals for namespace resolution, but
// typescript-eslint creates no variable definition for them.
func (s *RefStore) hasAuthoredProgramDefinition(name string) bool {
	if s.sourceFile == nil {
		return false
	}
	symbol := s.sourceFile.Locals[name]
	return hasAuthoredDeclaration(symbol) && !isQualifiedNamespaceNameOnly(symbol)
}

// IsGlobalReference is IsGlobalNameReference using the declaration-space
// meaning implied by node's syntactic reference position.
func (s *RefStore) IsGlobalReference(node *ast.Node) bool {
	if s == nil || node == nil || node.Kind != ast.KindIdentifier || !isReferencePosition(node) {
		return false
	}
	return s.IsGlobalNameReference(node, node.Text(), referenceMeaning(node))
}

// hasAuthoredDeclaration reports whether symbol has a declaration the file
// actually spells in syntax. A JSDoc tag such as `@typedef` or `@import` is
// reparsed into synthesized declaration nodes that join the file's symbol
// table; ESLint reads that same text as a comment and creates no scope variable
// for it, so a symbol declared only that way defines nothing. A SourceFile
// declaration is likewise synthesized for the CommonJS `exports` wrapper.
func hasAuthoredDeclaration(symbol *ast.Symbol) bool {
	if symbol == nil {
		return false
	}
	for _, decl := range symbol.Declarations {
		if decl.Kind != ast.KindSourceFile && !isJSDocSynthesized(decl) {
			return true
		}
	}
	return false
}

// isJSDocSynthesized reports whether the parser produced node from a JSDoc
// tag. Only the root of a reparsed subtree carries the flag — the import
// specifier a `@import` tag binds is a deep clone inside a flagged clause — so
// the answer comes from the ancestor chain.
func isJSDocSynthesized(node *ast.Node) bool {
	for current := node; current != nil; current = current.Parent {
		if current.Flags&ast.NodeFlagsReparsed != 0 {
			return true
		}
	}
	return false
}

// HasNonGlobalProgramScope reports whether program-level declarations should
// be treated as living outside the global scope. Module syntax and resolved
// language defaults can both contribute; an authored global program scope
// (script, or commonjs on TypeScript-flavoured extensions) forces global
// program scope instead.
func (s *RefStore) HasNonGlobalProgramScope() bool {
	if s == nil || s.init.globalTopLevelScope {
		return false
	}
	return ast.IsExternalModule(s.sourceFile) || s.init.nonGlobalTopLevelScope
}

// binderReferenceSymbol resolves one reference without checker work. A named
// export nested in a namespace needs one extra guard: the binder exposes that
// export's own alias in the namespace's Exports table. If no local declaration
// matches the requested meaning, NameResolver can otherwise return the alias
// being declared instead of continuing to an outer lexical declaration.
func (s *RefStore) binderReferenceSymbol(node *ast.Node, meaning ast.SymbolFlags) *ast.Symbol {
	target := s.resolveName(node, node.Text(), meaning)
	if !isOwnExportAlias(node, target) {
		return target
	}
	outer := outerExportScopeLocation(node)
	if outer == nil {
		return nil
	}
	return s.resolveName(outer, node.Text(), meaning)
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
