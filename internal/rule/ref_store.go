package rule

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/binder"
	"github.com/microsoft/typescript-go/shim/core"
)

// RefStore owns the per-file identifier-reference index for one linted source
// file: given a symbol declared in the file, it returns every identifier that
// references it, in source order.
//
// References are resolved with the binder's NameResolver — the same scope walk
// the checker performs for identifiers — so lookups never touch the
// TypeChecker and never trigger lazy type computation. The store consequently
// deals in raw binder symbols: query it with the symbol attached to a
// declaration node (node.Symbol()), not with checker.GetSymbolAtLocation
// results, which may be checker-merged and compare unequal.
//
// Collection is lazy twice over: the single AST walk that gathers candidate
// identifiers runs on first use, and name resolution runs once per queried
// symbol name, so a rule asking about `x` never pays for resolving unrelated
// identifiers. A RefStore belongs to one lintFile invocation and is therefore
// used serially by that file's rule initializers and listeners.
type RefStore struct {
	sourceFile *ast.SourceFile
	resolver   binder.NameResolver
	walked     bool
	// candidates maps identifier text to the reference-position identifiers
	// still awaiting resolution; entries move into refs on first query.
	candidates map[string][]*ast.Node
	refs       map[*ast.Symbol][]*ast.Node
}

// NewRefStore creates the reference index for one source file. options must
// be the file's program options (the resolver consults script target and
// module settings during scope walks).
func NewRefStore(sourceFile *ast.SourceFile, options *core.CompilerOptions) *RefStore {
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
	if pending, ok := s.candidates[sym.Name]; ok {
		delete(s.candidates, sym.Name)
		for _, id := range pending {
			target := s.resolver.Resolve(id, id.Text(), referenceMeaning(id), nil, true /*isUse*/, false /*excludeGlobals*/)
			if target != nil {
				s.refs[target] = append(s.refs[target], id)
			}
		}
	}
	return s.refs[sym]
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
// something non-local (property names, import/export bindings, labels,
// intrinsic JSX tags).
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
	case ast.KindImportSpecifier, ast.KindExportSpecifier, ast.KindNamespaceImport,
		ast.KindImportClause, ast.KindNamespaceExport:
		// Import/export bindings resolve through module/alias machinery the
		// checker owns; the current consumers never treat them as references.
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
