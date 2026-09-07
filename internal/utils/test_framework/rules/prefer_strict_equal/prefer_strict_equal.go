// Package prefer_strict_equal implements the matcher-independent body shared
// by test-framework rules that prefer toStrictEqual over toEqual.
package prefer_strict_equal

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

const strictMatcher = "toStrictEqual"

type ExpectCall struct {
	Matcher      string
	MatcherEntry testFramework.MemberEntry
}

type Runtime struct {
	Parse func(*ast.Node) *ExpectCall
}

type Config struct {
	Name    string
	Prepare func(rule.RuleContext) Runtime
}

var (
	useToStrictEqualMessage = rule.RuleMessage{
		Id:          "useToStrictEqual",
		Description: "Use `toStrictEqual()` instead",
	}
	suggestReplaceWithStrictEqualMessage = rule.RuleMessage{
		Id:          "suggestReplaceWithStrictEqual",
		Description: "Replace with `toStrictEqual()`",
	}
)

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
					parsed := runtime.Parse(node)
					if parsed == nil || parsed.Matcher != "toEqual" || parsed.MatcherEntry.Node == nil {
						return
					}

					matcherNode := parsed.MatcherEntry.Node
					ctx.ReportNodeWithDeferredSuggestions(
						matcherNode,
						useToStrictEqualMessage,
						func() []rule.RuleSuggestion {
							fixRange, fixText, ok := testFramework.AccessorReplacement(
								ctx.SourceFile,
								matcherNode,
								strictMatcher,
							)
							if !ok {
								return nil
							}
							return []rule.RuleSuggestion{{
								Message: suggestReplaceWithStrictEqualMessage,
								FixesArr: []rule.RuleFix{{
									Range: fixRange,
									Text:  fixText,
								}},
							}}
						},
					)
				},
			}
		},
	}
}
