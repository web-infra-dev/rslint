package no_await_in_promise_methods

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

const (
	messageIDError      = "no-await-in-promise-methods/error"
	messageIDSuggestion = "no-await-in-promise-methods/suggestion"
)

var promiseMethods = []string{"all", "allSettled", "any", "race"}

func errorMessage(method string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          messageIDError,
		Description: fmt.Sprintf("Promise in `Promise.%s()` should not be awaited.", method),
		Data:        map[string]string{"method": method},
	}
}

var suggestionMessage = rule.RuleMessage{
	Id:          messageIDSuggestion,
	Description: "Remove `await`.",
}

func removeAwaitSuggestion(sourceFile *ast.SourceFile, awaitExpression *ast.Node) []rule.RuleSuggestion {
	text := sourceFile.Text()
	awaitRange := utils.TrimNodeTextRange(sourceFile, awaitExpression)
	awaitEnd := awaitRange.Pos() + len("await")
	whitespaceEnd := ecmascript.SkipLeadingWhitespace(text, awaitEnd, len(text))

	return []rule.RuleSuggestion{{
		Message: suggestionMessage,
		FixesArr: []rule.RuleFix{
			rule.RuleFixRemoveRange(core.NewTextRange(awaitRange.Pos(), whitespaceEnd)),
		},
	}}
}

// NoAwaitInPromiseMethodsRule disallows awaiting direct elements passed to
// Promise combinator methods.
//
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v73.0.0/rules/no-await-in-promise-methods.js
var NoAwaitInPromiseMethodsRule = rule.Rule{
	Name:   "unicorn/no-await-in-promise-methods",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		oneArgument := 1

		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				call, ok := unicornutil.MatchDotMethodCall(node, unicornutil.DotMethodCallOptions{
					Methods:             promiseMethods,
					ArgumentsLength:     &oneArgument,
					RejectSpreadElement: true,
				})
				if !ok {
					return
				}

				object := ast.SkipParentheses(call.Object)
				if object == nil || !ast.IsIdentifier(object) ||
					object.AsIdentifier().Text != "Promise" {
					return
				}

				argument := ast.SkipParentheses(node.Arguments()[0])
				if argument == nil || argument.Kind != ast.KindArrayLiteralExpression {
					return
				}

				method := call.Property.AsIdentifier().Text
				for _, element := range argument.AsArrayLiteralExpression().Elements.Nodes {
					element = ast.SkipParentheses(element)
					if element == nil || element.Kind != ast.KindAwaitExpression {
						continue
					}

					ctx.ReportNodeWithDeferredSuggestions(
						element,
						errorMessage(method),
						func() []rule.RuleSuggestion {
							return removeAwaitSuggestion(ctx.SourceFile, element)
						},
					)
				}
			},
		}
	},
}
