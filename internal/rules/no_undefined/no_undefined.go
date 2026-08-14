package no_undefined

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// https://eslint.org/docs/latest/rules/no-undefined
//
// ESLint's implementation walks every scope and flags each declaration and
// non-initializing reference of a variable literally named "undefined". A
// variable resolves purely by name for this rule, so every occurrence of the
// identifier "undefined" in a value-space declaration, binding, or reference
// position is reportable regardless of which scope it resolves in — there is
// no case where a name-only match fails to be flagged for a scope reason.
// This lets the port skip scope resolution entirely and instead classify
// each `undefined` identifier occurrence by its syntactic position: report
// it unless the position is a pure syntactic key/label (property access
// name, non-computed object/class/interface member name, import/export
// remote name, etc.) that never participates in variable binding.
var NoUndefinedRule = rule.Rule{
	Name:   "no-undefined",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindIdentifier: func(node *ast.Node) {
				if node.Text() != "undefined" {
					return
				}
				if isNonBindingUndefinedPosition(node) {
					return
				}
				ctx.ReportRange(undefinedReportRange(ctx.SourceFile, node), rule.RuleMessage{
					Id:          "unexpectedUndefined",
					Description: "Unexpected use of undefined.",
				})
			},
		}
	},
}

// isNonBindingUndefinedPosition reports whether node sits in a syntactic
// position that never creates or reads a variable binding — a property/member
// name, an import/export specifier's remote name, a label, or similar — and
// therefore is not a use of the `undefined` variable regardless of its text.
func isNonBindingUndefinedPosition(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return true
	}

	switch parent.Kind {
	case ast.KindPropertyAccessExpression:
		// obj.undefined -> "undefined" only names the property, not a variable.
		return parent.AsPropertyAccessExpression().Name() == node

	case ast.KindPropertySignature, ast.KindMethodSignature, ast.KindMethodDeclaration,
		ast.KindPropertyDeclaration, ast.KindGetAccessor, ast.KindSetAccessor,
		ast.KindPropertyAssignment, ast.KindNamedTupleMember, ast.KindJsxAttribute:
		// Non-computed member/key names ({ undefined: 1 }, class C { undefined() {} },
		// interface I { undefined: string }, [undefined: string] tuple label,
		// <div undefined="x" />) name a slot, not a variable. Computed forms
		// ([undefined]: 1) put an intervening ComputedPropertyName node between
		// node and parent, so they never match here and remain reportable.
		return parent.Name() == node

	case ast.KindBindingElement, ast.KindImportSpecifier:
		// { key: newName } destructuring and `import { key as newName }` both
		// use PropertyName for the remote/source name, which is not a local
		// binding; Name() is the actual local binding and stays reportable.
		return parent.PropertyName() == node

	case ast.KindExportSpecifier:
		return isNonBindingExportSpecifierName(node, parent)

	case ast.KindLabeledStatement, ast.KindBreakStatement, ast.KindContinueStatement:
		return true

	case ast.KindImportAttribute:
		// import ... with { undefined: "json" } — an attribute key, not a variable.
		return parent.AsImportAttribute().Name() == node

	case ast.KindParameter:
		// An index signature's parameter (`interface I { [undefined: string]: any }`)
		// names the key slot rather than a variable. Every other parameter name is
		// a real binding, and an initializer identifier is a reference either way.
		return parent.Name() == node && parent.Parent != nil && parent.Parent.Kind == ast.KindIndexSignature

	case ast.KindModuleDeclaration:
		// Only a standalone `namespace undefined {}` binds a variable; the
		// segments of a dotted name (`namespace undefined.A {}`) do not.
		return parent.Name() == node && isDottedModuleNameSegment(parent)

	case ast.KindJsxNamespacedName:
		// A namespaced tag name (`<undefined:foo />`) references both of its
		// segments, while a namespaced attribute name (`<X undefined:attr />`)
		// names a slot on the element.
		return !ast.IsJsxTagName(parent)

	case ast.KindQualifiedName:
		// In `A.undefined` (a type-space qualified name), only the leftmost
		// root is ever a reference; every right-hand name is a member name.
		return parent.AsQualifiedName().Right == node

	case ast.KindNamespaceExport:
		// `export * as undefined from 'mod'` only ever appears in a re-export;
		// it never creates a binding visible in this file.
		return true
	}

	if utils.IsImportTypeSyntax(node) {
		return true
	}
	if utils.IsInJSDocSyntax(node) {
		return true
	}
	if ast.IsJsxTagName(node) {
		return true
	}

	return false
}

// isDottedModuleNameSegment reports whether decl is one link of a dotted
// namespace name. tsgo lowers `namespace A.B {}` to nested module declarations
// — the outer one holds another module declaration as its body, and the inner
// one is that body — so both shapes identify a segment that only qualifies the
// namespace path instead of declaring a name of its own.
func isDottedModuleNameSegment(decl *ast.Node) bool {
	if body := decl.Body(); body != nil && ast.IsModuleDeclaration(body) {
		return true
	}
	return decl.Parent != nil && ast.IsModuleDeclaration(decl.Parent) && decl.Parent.Body() == decl
}

// undefinedReportRange spans what typescript-eslint hands ESLint as the
// reported node. A declaration name there carries its own optional marker and
// type annotation inside the identifier, so `function f(undefined: number) {}`
// reports through the annotation; every other occurrence spans the bare
// identifier.
func undefinedReportRange(sourceFile *ast.SourceFile, node *ast.Node) core.TextRange {
	identifierRange := utils.TrimNodeTextRange(sourceFile, node)
	if end := declarationNameEnd(node); end > identifierRange.End() {
		return identifierRange.WithEnd(end)
	}
	return identifierRange
}

// declarationNameEnd returns the end of the trailing marker or annotation that
// belongs to node's declaration name, or 0 when node names nothing that can
// carry one.
func declarationNameEnd(node *ast.Node) int {
	parent := node.Parent
	if parent == nil || parent.Name() != node {
		return 0
	}

	switch parent.Kind {
	case ast.KindParameter:
		decl := parent.AsParameterDeclaration()
		// A rest parameter's annotation belongs to the rest element upstream,
		// which leaves the identifier itself bare.
		if decl.DotDotDotToken != nil {
			return 0
		}
		if decl.Type != nil {
			return decl.Type.End()
		}
		if decl.QuestionToken != nil {
			return decl.QuestionToken.End()
		}

	case ast.KindVariableDeclaration:
		if declType := parent.AsVariableDeclaration().Type; declType != nil {
			return declType.End()
		}
	}

	return 0
}

// isNonBindingExportSpecifierName classifies the two name slots of an
// ExportSpecifier. A re-export (`export { x } from 'mod'`) never touches a
// local binding, so both slots are non-binding. A local export's PropertyName
// (when aliased: `export { local as label }`) is the real local reference; its
// Name() is a pure public label. A local, non-aliased export's Name() (no
// PropertyName) is both the label and the local reference simultaneously.
func isNonBindingExportSpecifierName(node *ast.Node, parent *ast.Node) bool {
	if utils.IsReExportSpecifier(parent) {
		return true
	}
	if parent.PropertyName() != nil {
		return parent.Name() == node
	}
	return false
}
