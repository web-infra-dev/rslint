package no_untyped_mock_factory

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	jestUtils "github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	shared "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/no_untyped_mock_factory"
)

// Source: eslint-plugin-jest v29.16.6, no-untyped-mock-factory.
var NoUntypedMockFactoryRule = shared.NewRule(shared.Config{
	Name: "jest/no-untyped-mock-factory",
	Candidates: func(ctx rule.RuleContext) func(*ast.Node) bool {
		analysis := jestUtils.GetJestCallAnalysis(ctx)
		return func(node *ast.Node) bool {
			// The upstream parser reports only the outer call, including when a
			// mock call is an argument to another call. Keep this rule's contract
			// local rather than changing the parser for its other consumers.
			parent := ast.WalkUpParenthesizedExpressions(node.Parent)
			if parent != nil && (parent.Kind == ast.KindCallExpression || parent.Kind == ast.KindPropertyAccessExpression || parent.Kind == ast.KindElementAccessExpression) {
				return false
			}

			callee := ast.SkipParentheses(node.AsCallExpression().Expression)
			if callee == nil || (callee.Kind != ast.KindPropertyAccessExpression && callee.Kind != ast.KindElementAccessExpression) {
				return false
			}
			parsed := analysis.ParseFnCall(node)
			if parsed == nil || parsed.Kind != jestUtils.JestFnTypeJest || len(parsed.Members) == 0 {
				return false
			}
			member := parsed.Members[len(parsed.Members)-1]
			return member == "mock" || member == "doMock"
		}
	},
})
