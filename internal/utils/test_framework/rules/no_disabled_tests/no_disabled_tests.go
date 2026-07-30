package no_disabled_tests

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

type ParsedCall struct {
	Call    *testFramework.ParsedCall
	HasSkip bool
	HasTodo bool
}

type Config struct {
	Name                    string
	Parse                   func(*ast.Node, rule.RuleContext) *ParsedCall
	IsStandaloneSkippedCall func(*ast.Node, rule.RuleContext) bool
}

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

// NewRule creates a no-disabled-tests rule for a test framework.
func NewRule(config Config) rule.Rule {
	return rule.Rule{
		Name:   config.Name,
		Schema: rule.EmptyArraySchema,
		Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
			return rule.RuleListeners{
				ast.KindCallExpression: func(node *ast.Node) {
					if config.IsStandaloneSkippedCall != nil &&
						config.IsStandaloneSkippedCall(node, ctx) {
						ctx.ReportNode(node, buildErrorSkippedTestMessage())
						return
					}

					parsed := config.Parse(node, ctx)
					if parsed == nil || parsed.Call == nil ||
						parsed.Call.Kind != testFramework.FnKindDescribe &&
							parsed.Call.Kind != testFramework.FnKindTest {
						return
					}

					if parsed.HasSkip {
						ctx.ReportNode(node, buildErrorSkippedTestMessage())
					}

					if parsed.Call.Kind == testFramework.FnKindTest &&
						len(node.Arguments()) < 2 &&
						!parsed.HasTodo {
						ctx.ReportNode(node, buildErrorMissingFunctionMessage())
					}
				},
			}
		},
	}
}
