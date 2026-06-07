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

func isPromiseValueReference(node *ast.Node) bool {
	if node == nil || !ast.IsIdentifier(node) || node.AsIdentifier().Text != promiseName {
		return false
	}
	if utils.IsNonReferenceIdentifier(node) {
		return false
	}
	if ast.IsPartOfTypeNode(node) || ast.IsPartOfTypeQuery(node) {
		return false
	}
	return true
}

func shouldReportPromiseReference(ctx rule.RuleContext, node *ast.Node) bool {
	if ctx.TypeChecker != nil {
		symbol := utils.GetReferenceSymbol(node, ctx.TypeChecker)
		if symbol != nil && hasNonDefaultLibraryDeclaration(ctx, symbol) {
			return false
		}
		if utils.IsShadowed(node, promiseName) {
			return false
		}
		return true
	}

	return !utils.IsShadowed(node, promiseName)
}

func hasNonDefaultLibraryDeclaration(ctx rule.RuleContext, symbol *ast.Symbol) bool {
	if symbol == nil || ctx.Program == nil {
		return false
	}
	for _, declaration := range symbol.Declarations {
		sourceFile := ast.GetSourceFileOfNode(declaration)
		if sourceFile != nil && !utils.IsSourceFileDefaultLibrary(ctx.Program, sourceFile) {
			return true
		}
	}
	return false
}

var NoNativeRule = rule.Rule{
	Name: "promise/no-native",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindIdentifier: func(node *ast.Node) {
				if !isPromiseValueReference(node) || !shouldReportPromiseReference(ctx, node) {
					return
				}
				ctx.ReportNode(node, buildNameMessage(promiseName))
			},
		}
	},
}
