package scope

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

// IsReferenceIdentifier reports whether id reads or writes a binding, as
// opposed to naming one (declarations), naming a member (`a.b`, `{ b: 1 }`),
// or naming a label.
//
// eslint-scope models a declaration identifier as a reference with `init:
// true` and every consumer filters those out, so treating them as non-
// references here is equivalent — and it keeps the two representations of the
// same identifier from disagreeing.
func IsReferenceIdentifier(id *ast.Node) bool {
	if id == nil || id.Kind != ast.KindIdentifier {
		return false
	}
	parent := id.Parent
	if parent == nil {
		return false
	}

	// The argument, attributes, and qualifier of `import('m').a` are module
	// syntax rather than bindings in this file. Type arguments remain normal
	// references.
	if utils.IsImportTypeSyntax(id) {
		return false
	}
	// eslint-scope's JSXElement visitor only visits the opening tag. The
	// typescript-eslint visitor deliberately visits both opening and closing
	// tags, so preserve the parser split for every direct, member, and
	// namespaced JSX name shape.
	if ast.IsInJSFile(id) && isPartOfJsxClosingTagName(id) {
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
		if qn == nil || qn.Left != id {
			return false
		}
		// TSESTree represents the root of `typeof this.member` as a
		// TSThisType, which its type visitor deliberately does not reference.
		return id.Text() != "this" || !ast.IsPartOfTypeQuery(id)

	case ast.KindShorthandPropertyAssignment:
		// `{ a }` and `({ a = b } = c)` — the name is the value.
		return true

	case ast.KindExportSpecifier:
		// The local binding is `local` in `export { local as exported }`, and
		// the sole name in `export { local }`. Re-exports name another module's
		// binding and therefore have no local reference.
		if utils.IsReExportSpecifier(parent) {
			return false
		}
		spec := parent.AsExportSpecifier()
		if spec == nil {
			return false
		}
		local := spec.PropertyName
		if local == nil {
			local = spec.Name()
		}
		return local == id

	case ast.KindImportSpecifier, ast.KindNamespaceImport,
		ast.KindImportClause, ast.KindNamespaceExport:
		// Import bindings and namespace-export labels do not read a local
		// binding.
		return false

	case ast.KindPropertyAssignment:
		return parent.AsPropertyAssignment().Name() != id

	case ast.KindBindingElement:
		// `{ prop: local = init }` — the property and binding names do not read;
		// an identifier in the default initializer does.
		binding := parent.AsBindingElement()
		return binding != nil && binding.Initializer == id

	case ast.KindImportAttribute:
		return parent.AsImportAttribute().Name() != id

	case ast.KindLabeledStatement, ast.KindBreakStatement, ast.KindContinueStatement:
		return false

	case ast.KindMetaProperty:
		// `new.target`, `import.meta`.
		return false

	case ast.KindTypePredicate:
		// `x is T` / `asserts x is T` references the value binding named by
		// the predicate. `this is T` uses a ThisType node, not an Identifier.
		return true

	case ast.KindTypeQuery:
		// TSESTree exposes bare `typeof this` as TSThisType rather than an
		// Identifier, so scope-manager creates no reference for it.
		return id.Text() != "this"

	case ast.KindNamespaceExportDeclaration:
		// `export as namespace Name` references the exported value. This is
		// distinct from `export * as Name from`, whose NamespaceExport is only
		// an external export label.
		return true

	case ast.KindJsxAttribute:
		return false

	case ast.KindJsxNamespacedName:
		// typescript-eslint visits both pieces of a namespaced TSX tag as
		// value references; Espree visits neither piece in JSX. Namespaced
		// attribute names are excluded above by their JsxAttribute owner.
		return !ast.IsInJSFile(id) && ast.IsJsxTagName(parent)

	case ast.KindJsxOpeningElement, ast.KindJsxSelfClosingElement, ast.KindJsxClosingElement:
		if ast.IsInJSFile(id) && id.Text() == "this" {
			return false
		}
		return IsJsxComponentName(id.Text())
	}

	return !ast.IsDeclarationName(id)
}

// IsTypeScriptJsxThisReference reports the one scope-manager reference that
// tsgo cannot represent as an Identifier: bare `this` in a TSX tag name.
// TSESTree emits a JSXIdentifier there. It deliberately excludes
// `<this.Member>`, whose `this` object is skipped by scope-manager, and every
// JavaScript/JSX file, where Espree does not reference `this` tag names.
func IsTypeScriptJsxThisReference(node *ast.Node) bool {
	return node != nil && node.Kind == ast.KindThisKeyword &&
		!ast.IsInJSFile(node) && ast.IsJsxTagName(node)
}

// isPartOfJsxClosingTagName climbs a tag-name chain from any identifier piece
// to its JSX closing element. Property-access member names are included here
// even though IsReferenceIdentifier later rejects every piece except the
// root; the parser-level opening/closing decision applies to the whole chain.
func isPartOfJsxClosingTagName(id *ast.Node) bool {
	current := id
	for current != nil && current.Parent != nil {
		parent := current.Parent
		switch parent.Kind {
		case ast.KindPropertyAccessExpression:
			access := parent.AsPropertyAccessExpression()
			if access == nil || (access.Expression != current && access.Name() != current) {
				return false
			}
			current = parent
		case ast.KindJsxNamespacedName:
			current = parent
		case ast.KindJsxClosingElement:
			return ast.IsJsxTagName(current)
		default:
			return false
		}
	}
	return false
}

// ReferenceSpace is the pair of independent reference capabilities exposed by
// typescript-eslint's scope manager. A reference can be value-only, type-only,
// or dual-capable.
type ReferenceSpace uint8

const (
	ReferenceValue ReferenceSpace = 1 << iota
	ReferenceType

	ReferenceDual = ReferenceValue | ReferenceType
)

// IncludesValue reports whether the reference can resolve a value definition.
func (space ReferenceSpace) IncludesValue() bool {
	return space&ReferenceValue != 0
}

// IncludesType reports whether the reference can resolve a type definition.
func (space ReferenceSpace) IncludesType() bool {
	return space&ReferenceType != 0
}

// DeclarationMeaning translates typescript-eslint's reference capabilities
// to TypeScript binder declaration spaces. Namespace declarations are both
// value- and type-capable in scope-manager, while import aliases can satisfy
// either capability.
func (space ReferenceSpace) DeclarationMeaning() ast.SymbolFlags {
	meaning := ast.SymbolFlagsAlias
	if space.IncludesValue() {
		meaning |= ast.SymbolFlagsValue | ast.SymbolFlagsNamespace
	}
	if space.IncludesType() {
		meaning |= ast.SymbolFlagsType | ast.SymbolFlagsNamespace
	}
	return meaning
}

// ReferenceSpaces keeps typescript-eslint's scope capabilities independent
// from TypeScript's exact semantic meaning. They deliberately differ for some
// syntax: ESTree erases parentheses around a bare export assignment, while the
// TypeScript checker treats a parenthesized identifier as a value expression;
// import-equals is value-only to scope-manager but its lexical root has
// namespace meaning to the checker.
type ReferenceSpaces struct {
	ESLint     ReferenceSpace
	TypeScript ast.SemanticMeaning
}

// ClassifyReferenceSpaces returns the scope-manager and checker spaces of id.
// Callers should first use IsReferenceIdentifier when the AST position itself
// may be a declaration, member name, label, or other non-reference.
func ClassifyReferenceSpaces(id *ast.Node) ReferenceSpaces {
	return ReferenceSpaces{
		ESLint:     ESLintReferenceSpace(id),
		TypeScript: TypeScriptReferenceMeaning(id),
	}
}

// ESLintReferenceSpace reports which binding spaces id can resolve to in
// typescript-eslint's scope manager. Most identifiers name either a value or a
// type; export forms can name whichever space the exported binding occupies.
func ESLintReferenceSpace(id *ast.Node) ReferenceSpace {
	if id == nil || id.Parent == nil {
		return ReferenceValue
	}
	parent := id.Parent
	// TSESTree erases parentheses, so a parenthesized bare identifier remains
	// the direct declaration of `export default (X)` / `export = (X)` there.
	// Recover that shape before switching on the immediate tsgo parent.
	if parent.Kind == ast.KindExportAssignment || parent.Kind == ast.KindParenthesizedExpression {
		if exportParent := ast.WalkUpParenthesizedExpressions(parent); exportParent != nil &&
			exportParent.Kind == ast.KindExportAssignment {
			if assignment := exportParent.AsExportAssignment(); assignment != nil &&
				ast.SkipParentheses(assignment.Expression) == id {
				return ReferenceDual
			}
		}
	}
	switch parent.Kind {
	case ast.KindExportSpecifier:
		// `export type { T }` and `export { type T }` export types only.
		if spec := parent.AsExportSpecifier(); spec != nil && spec.IsTypeOnly {
			return ReferenceType
		}
		if named := parent.Parent; named != nil && named.Parent != nil {
			if decl := named.Parent.AsExportDeclaration(); decl != nil && decl.IsTypeOnly {
				return ReferenceType
			}
		}
		return ReferenceDual
	case ast.KindTypePredicate:
		// The parameter name is a value reference even though the surrounding
		// predicate is a type node.
		return ReferenceValue
	}

	if InTypePosition(id) {
		return ReferenceType
	}
	return ReferenceValue
}

// TypeScriptReferenceMeaning mirrors the declaration space TypeScript uses to
// resolve a lexical identifier reference. It stays separate from the ESLint
// space above so neither consumer has to approximate the other's semantics.
func TypeScriptReferenceMeaning(id *ast.Node) ast.SemanticMeaning {
	if id == nil || id.Parent == nil {
		return ast.SemanticMeaningValue
	}
	parent := id.Parent

	switch parent.Kind {
	// Only the raw direct child receives export-assignment All meaning.
	// Parentheses turn the identifier into an ordinary value expression in the
	// TypeScript AST even though ESTree erases them.
	case ast.KindExportAssignment:
		return ast.SemanticMeaningAll
	case ast.KindExportSpecifier:
		if ast.IsTypeOnlyImportOrExportDeclaration(parent) {
			return ast.SemanticMeaningType | ast.SemanticMeaningNamespace
		}
		return ast.SemanticMeaningAll
	case ast.KindTypePredicate:
		// A type-predicate parameter reads a value.
		return ast.SemanticMeaningValue
	case ast.KindQualifiedName:
		// Dotted type names resolve their leftmost lexical identifier as a
		// namespace.
		if parent.AsQualifiedName().Left == id {
			if ast.IsPartOfTypeQuery(id) {
				return ast.SemanticMeaningValue
			}
			// This is also the root of an internal import-equals module name such
			// as `A` in `import X = A.B`. The final qualified entity can have All
			// meaning there, but its member name is not a lexical reference.
			return ast.SemanticMeaningNamespace
		}
	case ast.KindPropertyAccessExpression:
		// Type-only heritage clauses use PropertyAccessExpression for their
		// qualified target, while class extends and ordinary accesses are values.
		if ast.IsPartOfTypeQuery(id) {
			return ast.SemanticMeaningValue
		}
		if isTypeOnlyPropertyAccessQualifier(id) {
			return ast.SemanticMeaningNamespace
		}
		return ast.SemanticMeaningValue
	case ast.KindImportEqualsDeclaration:
		if parent.AsImportEqualsDeclaration().ModuleReference == id {
			return ast.SemanticMeaningNamespace
		}
	}

	// Most identifiers are ordinary values. Only ask the more specific
	// type-query question after establishing that the identifier is nested in
	// type syntax; a type-query operand still reads a value there.
	if InTypePosition(id) {
		if ast.IsPartOfTypeQuery(id) {
			return ast.SemanticMeaningValue
		}
		return ast.SemanticMeaningType
	}
	return ast.SemanticMeaningValue
}

// InTypePosition reports whether `id` names a type rather than a value.
// eslint-scope marks every identifier its type visitor reaches as a type
// reference; this is the same question asked of the tsgo AST.
//
// A dotted name is a type node as a whole, never in its parts, so the
// outermost one carries the answer. A heritage clause spells its dotted name
// with property accesses even in type position (`implements A.B.C`), so both
// shapes have to be climbed.
func InTypePosition(id *ast.Node) bool {
	if id == nil {
		return false
	}
	entity := id
	for entity.Parent != nil &&
		(entity.Parent.Kind == ast.KindQualifiedName || entity.Parent.Kind == ast.KindPropertyAccessExpression) {
		entity = entity.Parent
	}
	return ast.IsPartOfTypeNode(entity)
}

// isTypeOnlyPropertyAccessQualifier reports whether id is the lexical root of
// an expression-shaped qualified name used as a type. TypeScript represents
// `N.T` in `class C implements N.T` and `interface I extends N.T` with a
// PropertyAccessExpression, while class `extends N.T` remains value syntax.
func isTypeOnlyPropertyAccessQualifier(id *ast.Node) bool {
	entity := id
	for entity.Parent != nil && entity.Parent.Kind == ast.KindPropertyAccessExpression {
		access := entity.Parent.AsPropertyAccessExpression()
		if access == nil || access.Expression != entity {
			break
		}
		entity = entity.Parent
	}
	return entity != id && ast.IsPartOfTypeNode(entity)
}

// IsJsxComponentName reports whether a bare JSXIdentifier passes the shared
// JavaScript first-UTF-16-code-unit casing test. Astral letters therefore pass
// because their first code unit is an unchanged high surrogate. The special
// `this` result is consumed by typescript-eslint; the AST-aware caller excludes
// it for Espree, whose JSX visitor never creates a `this` reference.
func IsJsxComponentName(name string) bool {
	if name == "" {
		return false
	}
	if name == "this" {
		return true
	}
	firstByte := name[0]
	if firstByte < 0x80 {
		return firstByte < 'a' || firstByte > 'z'
	}
	firstRune, width := ecmascript.DecodeStringRune(name)
	if firstRune > 0xffff {
		return true
	}
	first := name[:width]
	return ecmascript.StringToUpperCase(first) == first
}
