package no_native

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const promiseName = "Promise"

func buildNameMessage(name string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "name",
		Description: `"` + name + `" is not defined.`,
		Data: map[string]string{
			"name": name,
		},
	}
}

// isPromiseReference reports whether node is an identifier `Promise` used as a
// reference (value or type), as opposed to a declaration name, property key, or
// label.
func isPromiseReference(node *ast.Node) bool {
	if node == nil || !ast.IsIdentifier(node) || node.AsIdentifier().Text != promiseName {
		return false
	}
	return !utils.IsNonReferenceIdentifier(node)
}

// isTypeReference reports whether the identifier appears in a type position. The
// `typeof` operand (a type query) syntactically lives in a type but references
// the *value* binding, so it is treated as a value reference.
func isTypeReference(node *ast.Node) bool {
	return ast.IsPartOfTypeNode(node) && !ast.IsPartOfTypeQuery(node)
}

func shouldReportPromiseReference(ctx rule.RuleContext, node *ast.Node) bool {
	if isTypeReference(node) {
		// Type references must be resolved against the type namespace only, and
		// syntactically: a merged symbol (e.g. `type Promise = string`) points
		// back at the lib, so the checker cannot tell a user type from native.
		return !utils.IsTypeShadowed(node, promiseName)
	}

	// Value reference.
	if ctx.TypeChecker != nil {
		if symbol := utils.GetReferenceSymbol(node, ctx.TypeChecker); symbol != nil &&
			hasNonDefaultLibraryValueDeclaration(ctx, symbol) {
			return false
		}
	}
	return !utils.IsShadowed(node, promiseName)
}

// hasNonDefaultLibraryValueDeclaration reports whether the symbol has a
// user-authored declaration that binds a *value*. Interface and type-alias
// declarations merge into the global `Promise` symbol but bind no value, so they
// must not count when validating a value reference.
func hasNonDefaultLibraryValueDeclaration(ctx rule.RuleContext, symbol *ast.Symbol) bool {
	if symbol == nil || ctx.Program == nil {
		return false
	}
	for _, declaration := range symbol.Declarations {
		if ast.IsInterfaceDeclaration(declaration) || ast.IsTypeAliasDeclaration(declaration) {
			continue
		}
		sourceFile := ast.GetSourceFileOfNode(declaration)
		if sourceFile != nil && !utils.IsSourceFileDefaultLibrary(ctx.Program, sourceFile) {
			return true
		}
	}
	return false
}

var NoNativeRule = rule.Rule{
	Name: "promise/no-native",
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindIdentifier: func(node *ast.Node) {
				if !isPromiseReference(node) || !shouldReportPromiseReference(ctx, node) {
					return
				}
				ctx.ReportNode(node, buildNameMessage(promiseName))
			},
		}
	},
}
