package scope

import (
	"github.com/microsoft/typescript-go/shim/ast"
)

// isReferenceIdentifier reports whether `id` reads or writes a binding, as
// opposed to naming one (declarations), naming a member (`a.b`, `{ b: 1 }`),
// or naming a label.
//
// eslint-scope models a declaration identifier as a reference with `init:
// true` and every consumer filters those out, so treating them as non-
// references here is equivalent — and it keeps the two representations of the
// same identifier from disagreeing.
func isReferenceIdentifier(id *ast.Node) bool {
	if id == nil || id.Kind != ast.KindIdentifier {
		return false
	}
	parent := id.Parent
	if parent == nil {
		return false
	}

	// `typeof import('m').a.b` names exports of another module, not bindings
	// in this file.
	if isImportTypeQualifier(id) {
		return false
	}

	switch parent.Kind {
	case ast.KindPropertyAccessExpression:
		// Only the object is a binding: `a.b` references `a`, not `b`.
		pae := parent.AsPropertyAccessExpression()
		return pae != nil && pae.Expression == id

	case ast.KindQualifiedName:
		// Same shape in type position: `A.B.C` references `A`.
		qn := parent.AsQualifiedName()
		return qn != nil && qn.Left == id

	case ast.KindShorthandPropertyAssignment:
		// `{ a }` and `({ a = b } = c)` — the name is the value.
		return true

	case ast.KindExportSpecifier:
		// The local binding is `local` in `export { local as exported }`, and
		// the sole name in `export { local }`.
		spec := parent.AsExportSpecifier()
		if spec == nil {
			return false
		}
		local := spec.PropertyName
		if local == nil {
			local = spec.Name()
		}
		return local == id

	case ast.KindImportSpecifier:
		// Neither the imported name nor the local alias reads a local binding.
		return false

	case ast.KindBindingElement:
		// `{ prop: local = init }` — only the default initializer reads.
		be := parent.AsBindingElement()
		return be != nil && be.Initializer == id

	case ast.KindLabeledStatement, ast.KindBreakStatement, ast.KindContinueStatement:
		return false

	case ast.KindMetaProperty:
		// `new.target`, `import.meta`.
		return false

	case ast.KindTypePredicate:
		// `x is T` — `x` names a parameter, it does not read it.
		return false

	case ast.KindJsxAttribute, ast.KindJsxNamespacedName:
		return false

	case ast.KindJsxOpeningElement, ast.KindJsxSelfClosingElement, ast.KindJsxClosingElement:
		return isJsxComponentName(id)
	}

	return !ast.IsDeclarationName(id)
}

// referenceSpaces reports which binding spaces `id` can resolve to. Most
// identifiers name either a value or a type depending on where they sit; the
// export forms name whichever space the exported binding happens to live in.
func referenceSpaces(id *ast.Node) (value bool, isType bool) {
	parent := id.Parent
	switch parent.Kind {
	case ast.KindExportSpecifier:
		// `export type { T }` and `export { type T }` export types only.
		if spec := parent.AsExportSpecifier(); spec != nil && spec.IsTypeOnly {
			return false, true
		}
		if named := parent.Parent; named != nil && named.Parent != nil {
			if decl := named.Parent.AsExportDeclaration(); decl != nil && decl.IsTypeOnly {
				return false, true
			}
		}
		return true, true
	case ast.KindExportAssignment:
		// `export = X` exports a value or a type; `export default X` is an
		// expression and names a value.
		if assignment := parent.AsExportAssignment(); assignment != nil && assignment.IsExportEquals {
			return true, true
		}
	}

	// `A.B.C` is a type node as a whole, never in its parts, so ask about the
	// outermost qualified name rather than the identifier itself.
	entity := id
	for entity.Parent != nil && entity.Parent.Kind == ast.KindQualifiedName {
		entity = entity.Parent
	}
	if ast.IsPartOfTypeNode(entity) {
		return false, true
	}
	return true, false
}

// isImportTypeQualifier reports whether `id` is part of the qualifier of an
// `import(...)` type — the `a` or `b` in `typeof import('m').a.b`.
func isImportTypeQualifier(id *ast.Node) bool {
	current := id
	for current.Parent != nil && current.Parent.Kind == ast.KindQualifiedName {
		current = current.Parent
	}
	parent := current.Parent
	if parent == nil || parent.Kind != ast.KindImportType {
		return false
	}
	importType := parent.AsImportTypeNode()
	return importType != nil && importType.Qualifier == current
}

// isJsxComponentName reports whether a JSX tag name refers to a binding.
// A tag name starting with a lower-case letter is an intrinsic element
// (`<div/>` is the string "div"), matching how JSX transforms resolve names.
func isJsxComponentName(id *ast.Node) bool {
	text := id.Text()
	if text == "" {
		return false
	}
	first := text[0]
	return first < 'a' || first > 'z'
}
