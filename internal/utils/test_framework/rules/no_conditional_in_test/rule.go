// Package no_conditional_in holds the framework-neutral traversal shared
// by jest/no-conditional-in-test and rstest/no-conditional-in-test.
//
// Known deliberate divergences from eslint-plugin-jest's no-conditional-in-test:
//
//  1. Test-case tracking is a depth counter, not a bool. Upstream sets
//     `inTestCase = true` and clears it on any test call's exit, so a nested
//     test registration ends the enclosing one early and a conditional written
//     after it goes unreported. Do not "restore parity" by turning it back
//     into a bool.
//  2. Optional chains are matched on tsgo's representation. tsgo has no
//     `ChainExpression` wrapper — the optional flag sits on individual
//     property, element, call and non-null links — so each outermost chain is
//     reported exactly once, with the range extended over the trailing `!` of
//     `a?.b!` the way ESLint's ChainExpression range covers it.
//
// The scope is the test registration call itself, so every argument of the
// call is inside the test and a function declared outside the call is not,
// even when the call names it as its callback.
package no_conditional_in

import (
	_ "embed"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	internalUtils "github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed no_conditional_in_test.schema.json
var schemaJSON []byte

// Runtime is the framework adapter the shared traversal drives.
type Runtime struct {
	// IsTestCall reports whether a CallExpression registers a test case.
	IsTestCall func(*ast.Node) bool
	// Skip turns the rule off for this file without walking it.
	Skip bool
}

type Config struct {
	Name string
	// Prepare creates the framework adapter once per file.
	Prepare func(ctx rule.RuleContext) Runtime
}

type options struct {
	allowOptionalChaining bool
}

func parseOptions(rawOptions []any) options {
	opts := options{allowOptionalChaining: true}
	if len(rawOptions) == 0 {
		return opts
	}

	optionMap, _ := rawOptions[0].(map[string]any)
	if allow, ok := optionMap["allowOptionalChaining"].(bool); ok {
		opts.allowOptionalChaining = allow
	}
	return opts
}

func buildConditionalInTestMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "conditionalInTest",
		Description: "Avoid having conditionals in tests",
	}
}

func isOutermostOptionalChain(node *ast.Node) bool {
	if !ast.IsOptionalChain(node) || !ast.IsOutermostOptionalChain(node) {
		return false
	}

	parent := node.Parent
	if parent == nil || !ast.IsOptionalChain(parent) {
		return true
	}

	switch parent.Kind {
	case ast.KindPropertyAccessExpression:
		return parent.AsPropertyAccessExpression().Expression != node
	case ast.KindElementAccessExpression:
		return parent.AsElementAccessExpression().Expression != node
	case ast.KindCallExpression:
		return parent.AsCallExpression().Expression != node
	default:
		return true
	}
}

func NewRule(config Config) rule.Rule {
	return rule.Rule{
		Name:   config.Name,
		Schema: rule.NewSchema(schemaJSON),
		Run: func(ctx rule.RuleContext, rawOptions []any) rule.RuleListeners {
			runtime := config.Prepare(ctx)
			if runtime.Skip || runtime.IsTestCall == nil {
				return rule.RuleListeners{}
			}
			opts := parseOptions(rawOptions)
			// The scope is the registration call, so every argument of the call
			// is inside the test. The counter, rather than a flag, keeps an
			// inner registration exiting from clearing the outer test's state.
			testCallDepth := 0
			testCalls := map[*ast.Node]bool{}

			reportConditional := func(node *ast.Node) {
				if testCallDepth > 0 {
					ctx.ReportNode(node, buildConditionalInTestMessage())
				}
			}

			reportOptionalChain := func(node *ast.Node) {
				if testCallDepth > 0 &&
					!opts.allowOptionalChaining &&
					isOutermostOptionalChain(node) {
					reportRange := internalUtils.TrimNodeTextRange(ctx.SourceFile, node)
					for parent := node.Parent; parent != nil &&
						parent.Kind == ast.KindNonNullExpression &&
						parent.AsNonNullExpression().Expression == node; parent = node.Parent {
						node = parent
						reportRange = reportRange.WithEnd(node.End())
					}
					ctx.ReportRange(reportRange, buildConditionalInTestMessage())
				}
			}

			return rule.RuleListeners{
				ast.KindIfStatement:           reportConditional,
				ast.KindSwitchStatement:       reportConditional,
				ast.KindConditionalExpression: reportConditional,
				ast.KindBinaryExpression: func(node *ast.Node) {
					if ast.IsLogicalExpression(node) {
						reportConditional(node)
					}
				},

				ast.KindPropertyAccessExpression: reportOptionalChain,
				ast.KindElementAccessExpression:  reportOptionalChain,
				ast.KindNonNullExpression:        reportOptionalChain,

				ast.KindCallExpression: func(node *ast.Node) {
					reportOptionalChain(node)
					if runtime.IsTestCall(node) {
						testCalls[node] = true
						testCallDepth++
					}
				},
				rule.ListenerOnExit(ast.KindCallExpression): func(node *ast.Node) {
					if !testCalls[node] {
						return
					}
					delete(testCalls, node)
					testCallDepth--
				},
			}
		},
	}
}
