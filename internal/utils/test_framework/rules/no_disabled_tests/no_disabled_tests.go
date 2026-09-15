package no_disabled_tests

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

type ParsedCall struct {
	Call    *testFramework.ParsedCall
	HasSkip bool
	HasTodo bool
}

type Runtime struct {
	Parse                   func(*ast.Node) *ParsedCall
	IsStandaloneSkippedCall func(*ast.Node) bool
	Skip                    bool
}

type Config struct {
	Name    string
	Prepare func(rule.RuleContext) Runtime
	// HasOptionsOverload marks frameworks whose test API accepts a
	// `(description, options, fn?)` overload. For those, a two-argument call
	// passing an options object registers a test without a body.
	HasOptionsOverload bool
}

// isMissingFunctionArgument reports whether a test registration omits its
// callback.
func isMissingFunctionArgument(node *ast.Node, hasOptionsOverload bool) bool {
	arguments := node.Arguments()
	if len(arguments) < 2 {
		return true
	}
	if !hasOptionsOverload || len(arguments) > 2 {
		return false
	}
	options := ast.SkipParentheses(arguments[1])
	return options != nil && options.Kind == ast.KindObjectLiteralExpression
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
			runtime := config.Prepare(ctx)
			if runtime.Skip {
				return rule.RuleListeners{}
			}
			return rule.RuleListeners{
				ast.KindCallExpression: func(node *ast.Node) {
					if runtime.IsStandaloneSkippedCall != nil &&
						runtime.IsStandaloneSkippedCall(node) {
						ctx.ReportNode(node, buildErrorSkippedTestMessage())
						return
					}

					parsed := runtime.Parse(node)
					if parsed == nil || parsed.Call == nil ||
						parsed.Call.Kind != testFramework.FnKindDescribe &&
							parsed.Call.Kind != testFramework.FnKindTest {
						return
					}

					if parsed.HasSkip {
						ctx.ReportNode(node, buildErrorSkippedTestMessage())
					}

					if parsed.Call.Kind == testFramework.FnKindTest &&
						!parsed.HasTodo &&
						isMissingFunctionArgument(node, config.HasOptionsOverload) {
						ctx.ReportNode(node, buildErrorMissingFunctionMessage())
					}
				},
			}
		},
	}
}
