package no_undef

import (
	_ "embed"
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed no_undef.schema.json
var schemaJSON []byte

// https://eslint.org/docs/latest/rules/no-undef
//
// Like ESLint, this rule is deliberately independent of TypeScript's checker
// environment. It resolves declarations in the current file, then applies the
// versioned globals selected by languageOptions and the config/inline globals
// exposed on RuleContext. For TypeScript syntax, it also mirrors the parser's
// default esnext type-space globals. DOM, Node, ambient .d.ts, and cross-file
// declarations are not silently supplied by a TypeChecker.

type options struct {
	checkTypeof bool
}

func parseOptions(rawOptions []any) options {
	result := options{checkTypeof: false}
	if len(rawOptions) == 0 {
		return result
	}
	optsMap, _ := rawOptions[0].(map[string]any)
	if v, ok := optsMap["typeof"].(bool); ok {
		result.checkTypeof = v
	}
	return result
}

var NoUndefRule = rule.Rule{
	Name:   "no-undef",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, rawOptions []any) rule.RuleListeners {
		opts := parseOptions(rawOptions)

		if ctx.Refs == nil {
			return rule.RuleListeners{}
		}

		return rule.RuleListeners{
			ast.KindIdentifier: func(node *ast.Node) {
				// Skip identifiers that do not create scope references.
				if shouldSkip(node, opts.checkTypeof) {
					return
				}

				name := node.Text()
				if ctx.Globals.Access(name).IsDeclared() {
					return
				}
				referenceSpace := typescriptESLintReferenceSpace(node)
				// An explicit `off` removes only implicit globals. A declaration or
				// import authored in this file still defines its references.
				if ctx.Refs.ResolveInFileWithMeaning(node, referenceSpace.symbolMeaning()) != nil {
					return
				}
				// CommonJS's wrapper-level `arguments` binding is lexical rather than
				// a configurable global, so it remains defined even when a same-named
				// global is disabled.
				if ctx.Refs.HasImplicitWrapperBinding(name) {
					return
				}
				// typescript-eslint's scope manager seeds its default `esnext`
				// library in a separate type namespace. A name such as `Record`
				// therefore resolves in `type T = Record<string, unknown>`, but the
				// same spelling in `Record;` remains undefined. Keep this check after
				// same-file resolution so authored declarations retain normal scope
				// behavior, and do not consult the TypeChecker: ambient, cross-file,
				// and host-library symbols are outside core no-undef's inputs.
				if referenceSpace&eslintReferenceType != 0 && rule.IsDefaultTypeScriptTypeGlobal(name) {
					return
				}

				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "undef",
					Description: fmt.Sprintf("'%s' is not defined.", name),
				})
			},
		}
	},
}

// shouldSkip returns true if the identifier does not create a scope reference.
// This includes property/declaration names, labels, ImportType metadata,
// synthetic JSDoc syntax, and typeof operands (when checkTypeof is false).
func shouldSkip(node *ast.Node, checkTypeof bool) bool {
	if node.Parent == nil {
		return true
	}

	parent := node.Parent

	// Skip declaration names (var x, function x, class x, import x, etc.)
	// Exceptions: ShorthandPropertyAssignment names ({x}) and the name of a
	// non-aliased local export (export { x }) are declaration names but also
	// value references, so they must NOT be skipped. Re-exports
	// (export { x } from 'mod') reference the other module's binding instead,
	// and the alias label of export { x as y } is only a label.
	if ast.IsDeclarationName(node) && parent.Kind != ast.KindShorthandPropertyAssignment {
		if parent.Kind != ast.KindExportSpecifier || parent.PropertyName() != nil || utils.IsReExportSpecifier(parent) {
			return true
		}
	}

	// Skip property names in member access (obj.prop -> skip "prop")
	if parent.Kind == ast.KindPropertyAccessExpression &&
		parent.AsPropertyAccessExpression().Name() == node {
		return true
	}

	// Skip property names in object literals ({key: value} -> skip "key")
	if parent.Kind == ast.KindPropertyAssignment && parent.Name() == node {
		return true
	}

	// Skip property names in binding patterns ({key: newName} destructuring -> skip "key")
	if parent.Kind == ast.KindBindingElement && parent.PropertyName() == node {
		return true
	}

	// Skip the original name in aliased imports (import { Original as Alias } -> skip "Original")
	// The propertyName of an ImportSpecifier is the module's export name, not a local reference.
	if parent.Kind == ast.KindImportSpecifier && parent.PropertyName() == node {
		return true
	}

	// Skip the original name in re-export aliases (export { Original as Alias } from 'module')
	// The propertyName of an ExportSpecifier in a re-export is the source module's export name.
	// Note: without `from`, export { X as Y } refers to local X, so only skip when moduleSpecifier exists.
	if parent.Kind == ast.KindExportSpecifier && parent.PropertyName() == node {
		if utils.IsReExportSpecifier(parent) {
			return true
		}
	}

	// Skip label identifiers
	if parent.Kind == ast.KindLabeledStatement ||
		parent.Kind == ast.KindBreakStatement ||
		parent.Kind == ast.KindContinueStatement {
		return true
	}

	// Skip typeof operands (typeof x -> skip "x" unless checkTypeof is true).
	// Walk through ParenthesizedExpression nodes because typeof (x) and
	// typeof ((x)) are semantically identical to typeof x, but TypeScript's
	// AST inserts ParenthesizedExpression nodes unlike ESTree.
	if !checkTypeof && isTypeOfOperand(node) {
		return true
	}

	// `import("pkg").Name` treats both the module argument and qualifier as
	// module syntax rather than references. Its type arguments remain normal
	// type references and are deliberately not skipped.
	if utils.IsImportTypeSyntax(node) {
		return true
	}

	// In a qualified type name (`Namespace.Name`), only the leftmost root is a
	// reference. Every right-hand identifier names a member of that root.
	if parent.Kind == ast.KindQualifiedName && parent.AsQualifiedName().Right == node {
		return true
	}

	// Skip enum member names
	if parent.Kind == ast.KindEnumMember && parent.Name() == node {
		return true
	}

	// Skip meta-property names (`target` in `new.target`, `meta` in
	// `import.meta`, `defer` in `import.defer`) — syntactic, not identifiers
	// that reference a variable.
	if parent.Kind == ast.KindMetaProperty {
		return true
	}

	// Skip import attribute keys (`type` in `with { type: "json" }`) —
	// syntactic names, not references to same-named variables.
	if parent.Kind == ast.KindImportAttribute && parent.AsImportAttribute().Name() == node {
		return true
	}

	// Skip the namespace/name parts of a namespaced JSX tag or attribute
	// (`<foo:bar attr:name="v" />`) — syntactic, never variable references.
	if parent.Kind == ast.KindJsxNamespacedName {
		return true
	}

	// Skip lowercase JSX tag names (`<div>`): these are JSX intrinsics, not
	// identifier references. Uppercase tag names (`<Foo>`) reference a
	// component value and are checked normally.
	if ast.IsJsxTagName(node) {
		text := node.Text()
		if len(text) == 0 || (text[0] >= 'a' && text[0] <= 'z') {
			return true
		}
	}

	return false
}

type eslintReferenceSpace uint8

const (
	eslintReferenceValue eslintReferenceSpace = 1 << iota
	eslintReferenceType

	eslintReferenceDual = eslintReferenceValue | eslintReferenceType
)

// symbolMeaning translates typescript-eslint's two reference capabilities to
// TypeScript binder declaration spaces. Namespace declarations are considered
// both type- and value-capable by scope-manager, and import aliases likewise
// resolve in either capability regardless of `import type` spelling.
func (space eslintReferenceSpace) symbolMeaning() ast.SymbolFlags {
	meaning := ast.SymbolFlagsAlias
	if space&eslintReferenceValue != 0 {
		meaning |= ast.SymbolFlagsValue | ast.SymbolFlagsNamespace
	}
	if space&eslintReferenceType != 0 {
		meaning |= ast.SymbolFlagsType | ast.SymbolFlagsNamespace
	}
	return meaning
}

// typescriptESLintReferenceSpace mirrors the reference kind assigned by
// typescript-eslint's scope manager. Most type references are identified by
// TypeScript-Go's IsPartOfTypeNode helper; the explicit cases below cover
// dual-space exports, qualified heritage names, and value-only constructs
// nested inside type syntax.
func typescriptESLintReferenceSpace(node *ast.Node) eslintReferenceSpace {
	if node == nil || node.Parent == nil {
		return eslintReferenceValue
	}

	parent := node.Parent

	// `typeof T` and the parameter name in `x is T` are value references.
	if ast.IsPartOfTypeQuery(node) || parent.Kind == ast.KindTypePredicate {
		return eslintReferenceValue
	}

	// TypeScript import-equals visits its module reference as a value, even
	// though the TypeScript binder can also resolve namespaces there.
	if isImportEqualsModuleReference(node) {
		return eslintReferenceValue
	}

	// A bare identifier in an export assignment/default export, or in a local
	// export specifier, can resolve either a value or a type upstream.
	if parent.Kind == ast.KindExportAssignment {
		return eslintReferenceDual
	}
	if parent.Kind == ast.KindExportSpecifier && !utils.IsReExportSpecifier(parent) {
		if ast.IsTypeOnlyImportOrExportDeclaration(parent) {
			return eslintReferenceType
		}
		return eslintReferenceDual
	}

	// Only the leftmost identifier of a qualified type name is a reference.
	// shouldSkip has already removed every right-hand member name.
	if parent.Kind == ast.KindQualifiedName && parent.AsQualifiedName().Left == node {
		return eslintReferenceType
	}

	if ast.IsPartOfTypeNode(node) {
		return eslintReferenceType
	}

	// TypeScript represents dotted heritage targets as property-access chains.
	// Their lexical root is a type reference in `implements` and interface
	// `extends`, but a value reference in class `extends`.
	entity := node
	for entity.Parent != nil && entity.Parent.Kind == ast.KindPropertyAccessExpression {
		access := entity.Parent.AsPropertyAccessExpression()
		if access == nil || access.Expression != entity {
			break
		}
		entity = entity.Parent
	}
	if entity != node && ast.IsPartOfTypeNode(entity) {
		return eslintReferenceType
	}
	return eslintReferenceValue
}

func isImportEqualsModuleReference(node *ast.Node) bool {
	entity := node
	for entity.Parent != nil && entity.Parent.Kind == ast.KindQualifiedName {
		entity = entity.Parent
	}
	return entity.Parent != nil &&
		entity.Parent.Kind == ast.KindImportEqualsDeclaration &&
		entity.Parent.AsImportEqualsDeclaration().ModuleReference == entity
}

// isTypeOfOperand checks if the identifier is the sole operand of a typeof
// expression, walking through any ParenthesizedExpression wrappers.
// typeof x      → Identifier.Parent = TypeOfExpression → true
// typeof (x)    → Identifier.Parent = ParenExpr.Parent = TypeOfExpression → true
// typeof (a+b)  → Identifier.Parent = BinaryExpression → false
func isTypeOfOperand(node *ast.Node) bool {
	current := node.Parent
	for current != nil && current.Kind == ast.KindParenthesizedExpression {
		current = current.Parent
	}
	return current != nil && current.Kind == ast.KindTypeOfExpression
}
