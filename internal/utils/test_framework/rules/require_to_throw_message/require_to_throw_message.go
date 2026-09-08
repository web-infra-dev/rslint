// Package require_to_throw_message holds the framework-neutral rule body shared
// by jest/require-to-throw-message and rstest/require-to-throw-message.
package require_to_throw_message

import (
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

// ExpectCall is the part of a parsed expect assertion this rule needs.
// Framework adapters retain ownership of call-chain parsing and provenance.
type ExpectCall struct {
	Matcher      string
	MatcherEntry *testFramework.MemberEntry
	Modifiers    []string
	MatcherArgs  []*ast.Node
}

type Runtime struct {
	ParseExpectCall func(node *ast.Node) *ExpectCall
}

type Config struct {
	Name    string
	Prepare func(rule.RuleContext) Runtime
}

var throwMatchers = map[string]bool{
	"toThrow":      true,
	"toThrowError": true,
}

func buildAddErrorMessage(matcherName string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "addErrorMessage",
		Description: "Add an error message to " + matcherName + "()",
		Data: map[string]string{
			"matcherName": matcherName,
		},
	}
}

func shouldReport(call *ExpectCall) bool {
	return call != nil &&
		call.MatcherEntry != nil &&
		call.MatcherEntry.Node != nil &&
		throwMatchers[call.Matcher] &&
		len(call.MatcherArgs) == 0 &&
		!slices.Contains(call.Modifiers, "not")
}

func NewRule(config Config) rule.Rule {
	return rule.Rule{
		Name:   config.Name,
		Schema: rule.EmptyArraySchema,
		Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
			runtime := config.Prepare(ctx)
			return rule.RuleListeners{
				ast.KindCallExpression: func(node *ast.Node) {
					call := runtime.ParseExpectCall(node)
					if !shouldReport(call) {
						return
					}
					ctx.ReportNode(call.MatcherEntry.Node, buildAddErrorMessage(call.Matcher))
				},
			}
		},
	}
}
