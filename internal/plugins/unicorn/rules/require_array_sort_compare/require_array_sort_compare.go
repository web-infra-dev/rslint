package require_array_sort_compare

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const (
	messageID           = "require-array-sort-compare"
	numericSuggestionID = "require-array-sort-compare/numeric"
	stringSuggestionID  = "require-array-sort-compare/string"
	numericCompare      = "(a, b) => a - b"
	stringCompare       = "(a, b) => a.localeCompare(b)"
)

var missingCompareMessage = rule.RuleMessage{
	Id:          messageID,
	Description: "Pass a compare function to avoid sorting elements as strings.",
}

func hasCommentsInside(ctx rule.RuleContext, node *ast.Node) bool {
	textRange := utils.TrimNodeTextRange(ctx.SourceFile, node)
	return utils.HasCommentInSpan(ctx.Comments.All(), textRange.Pos(), textRange.End())
}

func compareSuggestion(
	ctx rule.RuleContext,
	call *ast.Node,
	argument *ast.Node,
	messageID string,
	message string,
	compare string,
) rule.RuleSuggestion {
	var fix rule.RuleFix
	if argument != nil {
		fix = rule.RuleFixReplace(ctx.SourceFile, argument, compare)
	} else {
		closingParenthesis := call.End() - 1
		fix = rule.RuleFixReplaceRange(
			core.NewTextRange(closingParenthesis, closingParenthesis),
			compare,
		)
	}
	return rule.RuleSuggestion{
		Message:  rule.RuleMessage{Id: messageID, Description: message},
		FixesArr: []rule.RuleFix{fix},
	}
}

// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/rules/require-array-sort-compare.js
var RequireArraySortCompareRule = rule.Rule{
	Name:   "unicorn/require-array-sort-compare",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		maximumArguments := 1
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				call, ok := unicornutil.MatchDotMethodCall(node, unicornutil.DotMethodCallOptions{
					Methods:             []string{"sort", "toSorted"},
					MaximumArguments:    &maximumArguments,
					RejectSpreadElement: true,
					AllowOptionalMember: true,
				})
				if !ok {
					return
				}

				var compareArgument *ast.Node
				arguments := node.Arguments()
				if len(arguments) == 1 {
					compareArgument = utils.ESTreeRuntimeExpression(arguments[0])
					if !utils.IsUndefinedIdentifier(compareArgument) {
						return
					}
				}

				if unicornutil.IsKnownNonArray(ctx, call.Object) {
					return
				}

				ctx.ReportNodeWithDeferredSuggestions(call.Property, missingCompareMessage, func() []rule.RuleSuggestion {
					if hasCommentsInside(ctx, node) {
						return nil
					}
					return []rule.RuleSuggestion{
						compareSuggestion(
							ctx,
							node,
							compareArgument,
							numericSuggestionID,
							"Sort numerically.",
							numericCompare,
						),
						compareSuggestion(
							ctx,
							node,
							compareArgument,
							stringSuggestionID,
							"Sort strings with `String#localeCompare()`.",
							stringCompare,
						),
					}
				})
			},
		}
	},
}
