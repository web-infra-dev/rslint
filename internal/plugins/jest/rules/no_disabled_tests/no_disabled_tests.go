package no_disabled_tests

import (
	"slices"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	shared "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/no_disabled_tests"
)

func isPendingCall(node *ast.Node, ctx rule.RuleContext) bool {
	if node == nil || node.Kind != ast.KindCallExpression {
		return false
	}

	callExpr := node.AsCallExpression()
	if callExpr == nil || callExpr.Expression == nil || callExpr.Expression.Kind != ast.KindIdentifier {
		return false
	}

	identifier := callExpr.Expression.AsIdentifier()
	if identifier == nil || identifier.Text != "pending" {
		return false
	}

	if ctx.TypeChecker == nil {
		return true
	}

	symbol := ctx.TypeChecker.GetSymbolAtLocation(callExpr.Expression)
	if symbol == nil {
		return true
	}

	for _, decl := range symbol.Declarations {
		if decl == nil {
			continue
		}
		if decl.Kind != ast.KindImportSpecifier {
			return false
		}

		importDecl := utils.FindImportDeclaration(decl)
		if importDecl == nil || importDecl.ModuleSpecifier == nil {
			return false
		}

		return importDecl.ModuleSpecifier.Text() == "@jest/globals"
	}

	return true
}

var NoDisabledTestsRule = shared.NewRule(shared.Config{
	Name: "jest/no-disabled-tests",
	Prepare: func(ctx rule.RuleContext) shared.Runtime {
		return shared.Runtime{
			Parse: func(node *ast.Node) *shared.ParsedCall {
				parsed := utils.ParseJestFnCall(node, ctx)
				if parsed == nil {
					return nil
				}
				return &shared.ParsedCall{
					Call: &parsed.ParsedCall,
					HasSkip: strings.HasPrefix(parsed.Name, "x") ||
						slices.Contains(parsed.Members, "skip"),
					HasTodo: slices.Contains(parsed.Members, "todo"),
				}
			},
			IsStandaloneSkippedCall: func(node *ast.Node) bool {
				return isPendingCall(node, ctx)
			},
		}
	},
})
