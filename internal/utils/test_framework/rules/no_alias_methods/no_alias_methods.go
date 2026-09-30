// Package no_alias_methods implements the body shared by test-framework rules
// that replace matcher aliases with their canonical names.
package no_alias_methods

import (
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

// MatcherAliases maps alias matchers to their canonical names. Jest and
// Vitest-compatible expect APIs define the same 11 pairs.
//
// Sources: eslint-plugin-jest v29.16.6 src/rules/no-alias-methods.ts;
// @vitest/expect 4.1.10 dist/index.d.ts (JestAssertion).
var MatcherAliases = map[string]string{
	"toBeCalled":       "toHaveBeenCalled",
	"toBeCalledTimes":  "toHaveBeenCalledTimes",
	"toBeCalledWith":   "toHaveBeenCalledWith",
	"lastCalledWith":   "toHaveBeenLastCalledWith",
	"nthCalledWith":    "toHaveBeenNthCalledWith",
	"toReturn":         "toHaveReturned",
	"toReturnTimes":    "toHaveReturnedTimes",
	"toReturnWith":     "toHaveReturnedWith",
	"lastReturnedWith": "toHaveLastReturnedWith",
	"nthReturnedWith":  "toHaveNthReturnedWith",
	"toThrowError":     "toThrow",
}

// Runtime.Parse returns the matcher that the expect chain completed by node
// calls, or nil. Each matcher must be returned for at most one node.
type Runtime struct {
	Parse func(*ast.Node) *testFramework.MemberEntry
}

type Config struct {
	Name    string
	Prepare func(rule.RuleContext) Runtime
}

func buildReplaceAliasMessage(alias string, canonical string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "replaceAlias",
		Description: fmt.Sprintf("Replace %s() with its canonical name of %s()", alias, canonical),
	}
}

func NewRule(config Config) rule.Rule {
	return rule.Rule{
		Name:   config.Name,
		Schema: rule.EmptyArraySchema,
		Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
			runtime := config.Prepare(ctx)
			return rule.RuleListeners{
				ast.KindCallExpression: func(node *ast.Node) {
					if runtime.Parse == nil {
						return
					}
					matcher := runtime.Parse(node)
					// A computed identifier key such as `expect(a)[toBeCalled]()`
					// names the matcher by the variable's runtime value, not by
					// its text, so it proves nothing about which matcher runs.
					if matcher == nil || matcher.Node == nil ||
						testFramework.IsComputedIdentifierAccessor(matcher.Node) {
						return
					}
					canonical, ok := MatcherAliases[matcher.Name]
					if !ok {
						return
					}

					matcherNode := matcher.Node
					ctx.ReportNodeWithDeferredFixes(
						matcherNode,
						buildReplaceAliasMessage(matcher.Name, canonical),
						func() []rule.RuleFix {
							fixRange, fixText, ok := testFramework.AccessorReplacement(
								ctx.SourceFile,
								matcherNode,
								canonical,
							)
							if !ok {
								return nil
							}
							return []rule.RuleFix{{
								Text:  fixText,
								Range: fixRange,
							}}
						},
					)
				},
			}
		},
	}
}
