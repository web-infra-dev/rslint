package prefer_flat_math_min_max

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const messageID = "prefer-flat-math-min-max"

var PreferFlatMathMinMaxRule = rule.Rule{
	Name:   "unicorn/prefer-flat-math-min-max",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				method, call, ok := mathMinMaxCall(node, "")
				if !ok || !hasNestedMathMinMaxCall(call, method) ||
					isNestedInSameMathMinMaxCall(node, method) {
					return
				}

				ctx.ReportNodeWithDeferredFixes(node, rule.RuleMessage{
					Id:          messageID,
					Description: "Prefer a flat `Math." + method + "()` call instead of nested calls.",
					Data:        map[string]string{"method": method},
				}, func() []rule.RuleFix {
					callRange := utils.TrimNodeTextRange(ctx.SourceFile, node)
					if utils.HasCommentInSpan(ctx.Comments.All(), callRange.Pos(), callRange.End()) {
						return nil
					}

					arguments := flattenedArguments(call, method)
					argumentTexts := make([]string, 0, len(arguments))
					for _, argument := range arguments {
						argumentTexts = append(argumentTexts, utils.TrimmedNodeText(ctx.SourceFile, argument))
					}
					replacement := utils.TrimmedNodeText(ctx.SourceFile, call.Callee) +
						"(" + strings.Join(argumentTexts, ", ") + ")"
					return []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, node, replacement)}
				})
			},
		}
	},
}

func mathMinMaxCall(node *ast.Node, expectedMethod string) (string, unicornutil.DotMethodCall, bool) {
	node = utils.ESTreeRuntimeExpression(node)
	call, ok := unicornutil.MatchDotMethodCall(node, unicornutil.DotMethodCallOptions{
		Methods: []string{"min", "max"},
	})
	if !ok {
		return "", unicornutil.DotMethodCall{}, false
	}

	object := utils.ESTreeRuntimeExpression(call.Object)
	if object == nil || !ast.IsIdentifier(object) || object.Text() != "Math" {
		return "", unicornutil.DotMethodCall{}, false
	}

	method := call.Property.Text()
	if expectedMethod != "" && method != expectedMethod {
		return "", unicornutil.DotMethodCall{}, false
	}
	return method, call, true
}

func hasNestedMathMinMaxCall(call unicornutil.DotMethodCall, method string) bool {
	for _, argument := range call.Call.Arguments() {
		if _, _, ok := mathMinMaxCall(argument, method); ok {
			return true
		}
	}
	return false
}

func flattenedArguments(call unicornutil.DotMethodCall, method string) []*ast.Node {
	var result []*ast.Node
	for _, argument := range call.Call.Arguments() {
		if _, nested, ok := mathMinMaxCall(argument, method); ok {
			result = append(result, flattenedArguments(nested, method)...)
			continue
		}
		result = append(result, utils.ESTreeRuntimeExpression(argument))
	}
	return result
}

func isNestedInSameMathMinMaxCall(node *ast.Node, method string) bool {
	current := node
	for current.Parent != nil && current.Parent.Kind == ast.KindParenthesizedExpression {
		current = current.Parent
	}

	_, parentCall, ok := mathMinMaxCall(current.Parent, method)
	if !ok {
		return false
	}
	for _, argument := range parentCall.Call.Arguments() {
		if argument == current {
			return true
		}
	}
	return false
}
