// Package prefer_string_trim_start_end ports eslint-plugin-unicorn's
// `prefer-string-trim-start-end` rule.
package prefer_string_trim_start_end

import (
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
)

const messageID = "prefer-string-trim-start-end"

var trimMethods = []string{"trimLeft", "trimRight"}

func preferTrimStartEndMessage(method, replacement string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          messageID,
		Description: fmt.Sprintf("Prefer `String#%s()` over `String#%s()`.", replacement, method),
		Data: map[string]string{
			"method":      method,
			"replacement": replacement,
		},
	}
}

// PreferStringTrimStartEndRule prefers the direction-independent names
// String#trimStart and String#trimEnd over their legacy aliases.
//
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v74.0.0/rules/prefer-string-trim-start-end.js
var PreferStringTrimStartEndRule = rule.Rule{
	Name:   "unicorn/prefer-string-trim-start-end",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		zeroArguments := 0
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				call, ok := unicornutil.MatchDotMethodCall(node, unicornutil.DotMethodCallOptions{
					Methods:             trimMethods,
					ArgumentsLength:     &zeroArguments,
					RejectSpreadElement: true,
					AllowOptionalMember: true,
				})
				if !ok || unicornutil.IsKnownNonString(ctx, call.Object) {
					return
				}

				method := call.Property.AsIdentifier().Text
				replacement := "trimStart"
				if method == "trimRight" {
					replacement = "trimEnd"
				}

				ctx.ReportNodeWithDeferredFixes(
					call.Property,
					preferTrimStartEndMessage(method, replacement),
					func() []rule.RuleFix {
						return []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, call.Property, replacement)}
					},
				)
			},
		}
	},
}
