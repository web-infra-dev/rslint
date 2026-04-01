package no_disabled_tests

import (
	"slices"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// Message Builder

func buildErrorMissingFunctionMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "missingFunction",
		Description: "Test is missing function argument",
	}
}

func buildErrorSkippedTestMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "skippedTest",
		Description: "Tests should not be skipped",
	}
}

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

var NoDisabledTestsRule = rule.Rule{
	Name: "jest/no-disabled-tests",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				if isPendingCall(node, ctx) {
					ctx.ReportNode(node, buildErrorSkippedTestMessage())
					return
				}

				jestFnCall := utils.ParseJestFnCall(node, ctx)
				if jestFnCall == nil ||
					jestFnCall.Kind != utils.JestFnTypeDescribe &&
						jestFnCall.Kind != utils.JestFnTypeTest {
					return
				}

				if strings.HasPrefix(jestFnCall.Name, "x") ||
					slices.Contains(jestFnCall.Members, "skip") {
					ctx.ReportNode(node, buildErrorSkippedTestMessage())
				}

				if jestFnCall.Kind == utils.JestFnTypeTest {
					if len(node.Arguments()) < 2 && !slices.Contains(jestFnCall.Members, "todo") {
						ctx.ReportNode(node, buildErrorMissingFunctionMessage())
					}
				}
			},
		}
	},
}
