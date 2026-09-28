package prefer_array_last_methods

import (
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const (
	messageID    = "prefer-array-last-methods"
	suggestionID = "replace"
)

var replacements = map[string]string{
	"find":      "findLast",
	"findIndex": "findLastIndex",
	"indexOf":   "lastIndexOf",
	"reduce":    "reduceRight",
}

var PreferArrayLastMethodsRule = rule.Rule{
	Name:   "unicorn/prefer-array-last-methods",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				outer, ok := unicornutil.MatchDotMethodCall(node, unicornutil.DotMethodCallOptions{
					Methods: []string{"find", "findIndex", "indexOf", "reduce"},
				})
				if !ok {
					return
				}

				reversingCall := utils.ESTreeRuntimeExpression(outer.Object)
				zeroArguments := 0
				inner, ok := unicornutil.MatchDotMethodCall(reversingCall, unicornutil.DotMethodCallOptions{
					Methods:         []string{"reverse", "toReversed"},
					ArgumentsLength: &zeroArguments,
				})
				if !ok {
					return
				}

				method := outer.Property.Text()
				replacement := replacements[method]
				reversingMethod := inner.Property.Text()
				message := rule.RuleMessage{
					Id:          messageID,
					Description: fmt.Sprintf("Prefer `Array#%s()` over `Array#%s().%s()`.", replacement, reversingMethod, method),
					Data: map[string]string{
						"method":          method,
						"replacement":     replacement,
						"reversingMethod": reversingMethod,
					},
				}

				reportRange := utils.TrimNodeTextRange(ctx.SourceFile, outer.Property)
				callRange := utils.TrimNodeTextRange(ctx.SourceFile, node)
				ctx.ReportRangeWithDeferredSuggestions(reportRange, message, func() []rule.RuleSuggestion {
					if utils.HasCommentInSpan(ctx.Comments.All(), callRange.Pos(), callRange.End()) {
						return nil
					}
					fixes := []rule.RuleFix{
						rule.RuleFixReplaceRange(reportRange, replacement),
					}
					fixes = append(fixes, unicornutil.RemoveMethodCallFixes(inner)...)
					return []rule.RuleSuggestion{{
						Message: rule.RuleMessage{
							Id:          suggestionID,
							Description: fmt.Sprintf("Replace `.%s().%s()` with `.%s()`.", reversingMethod, method, replacement),
							Data: map[string]string{
								"method":          method,
								"replacement":     replacement,
								"reversingMethod": reversingMethod,
							},
						},
						FixesArr: fixes,
					}}
				})
			},
		}
	},
}
