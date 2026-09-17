// Ported from eslint-plugin-unicorn v75.0.0; see LICENSE.
package unicornutil

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// AllowArrayMutationStatement reads the shared sort/reverse option's default.
func AllowArrayMutationStatement(options []any) bool {
	if len(options) == 0 {
		return true
	}
	object, ok := options[0].(map[string]any)
	return !ok || object["allowExpressionStatement"] != false
}

// ReportArrayMutation offers explicit, non-automatic alternatives to mutation.
func ReportArrayMutation(ctx rule.RuleContext, call DotMethodCall, replacement string, allowStatement bool) {
	array := utils.ESTreeRuntimeExpression(call.Object)
	var spread *ast.Node
	if array.Kind == ast.KindArrayLiteralExpression {
		elements := array.AsArrayLiteralExpression().Elements
		if elements != nil && len(elements.Nodes) == 1 && elements.Nodes[0].Kind == ast.KindSpreadElement {
			spread = elements.Nodes[0]
		}
	}
	if allowStatement && spread == nil {
		if parent := utils.ESTreeParent(call.Call); parent != nil && ast.IsExpressionStatement(parent) {
			return
		}
	}
	if ShouldSkipKnownNonArrayReceiver(ctx, call.Object) {
		return
	}
	method := call.Property.AsIdentifier().Text
	ctx.ReportNodeWithDeferredSuggestions(call.Property, rule.RuleMessage{
		Id:          "error",
		Description: "Use `Array#" + replacement + "()` instead of `Array#" + method + "()`.",
	}, func() []rule.RuleSuggestion {
		methodFix := rule.RuleFixReplace(ctx.SourceFile, call.Property, replacement)
		var suggestions []rule.RuleSuggestion
		if spread != nil {
			rawArgument := spread.AsSpreadElement().Expression
			argument := utils.ESTreeRuntimeExpression(rawArgument)
			text := utils.TrimmedNodeText(ctx.SourceFile, rawArgument)
			if rawArgument == argument && ShouldAddParenthesesToMemberExpressionObject(ctx.SourceFile, argument) {
				text = "(" + text + ")"
			}
			suggestions = append(suggestions, rule.RuleSuggestion{
				Message:  rule.RuleMessage{Id: "suggestion-spreading-array", Description: "The spreading object is an array."},
				FixesArr: []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, array, text), methodFix},
			})
		}
		message := rule.RuleMessage{
			Id:          "suggestion-apply-replacement",
			Description: "Switch to `." + replacement + "()`.",
		}
		if spread != nil {
			message = rule.RuleMessage{
				Id:          "suggestion-not-spreading-array",
				Description: "The spreading object is NOT an array.",
			}
		}
		return append(suggestions, rule.RuleSuggestion{Message: message, FixesArr: []rule.RuleFix{methodFix}})
	})
}
