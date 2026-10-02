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
		primitiveValues := utils.NewStaticStringEvaluatorWithReferenceResolver(ctx.TypeChecker, ctx.SourceFile, ctx.Refs)
		primitiveValues.GlobalAccess = ctx.Globals.Access

		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				method, call, ok := mathMinMaxCall(ctx, node, "")
				if !ok || !hasNestedMathMinMaxCall(ctx, call, method) ||
					isNestedInSameMathMinMaxCall(ctx, node, method) {
					return
				}

				ctx.ReportNodeWithDeferredFixes(node, rule.RuleMessage{
					Id:          messageID,
					Description: "Prefer a flat `Math." + method + "()` call instead of nested calls.",
					Data:        map[string]string{"method": method},
				}, func() []rule.RuleFix {
					callRange := utils.TrimNodeTextRange(ctx.SourceFile, node)
					if utils.HasCommentInSpan(ctx.Comments.All(), callRange.Pos(), callRange.End()) ||
						!nestedArgumentsHaveSafeCoercion(ctx, primitiveValues, call, method) {
						return nil
					}

					arguments := flattenedArguments(ctx, call, method)
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

func mathMinMaxCall(ctx rule.RuleContext, node *ast.Node, expectedMethod string) (string, unicornutil.DotMethodCall, bool) {
	node = utils.ESTreeRuntimeExpression(node)
	call, ok := unicornutil.MatchDotMethodCall(node, unicornutil.DotMethodCallOptions{
		Methods: []string{"min", "max"},
	})
	if !ok {
		return "", unicornutil.DotMethodCall{}, false
	}

	object := utils.ESTreeRuntimeExpression(call.Object)
	if object == nil || !ast.IsIdentifier(object) || object.Text() != "Math" ||
		!ctx.Globals.Access("Math").IsDeclared() || !unicornutil.IsGlobalReference(ctx, object) {
		return "", unicornutil.DotMethodCall{}, false
	}

	method := call.Property.Text()
	if expectedMethod != "" && method != expectedMethod {
		return "", unicornutil.DotMethodCall{}, false
	}
	return method, call, true
}

func hasNestedMathMinMaxCall(ctx rule.RuleContext, call unicornutil.DotMethodCall, method string) bool {
	for _, argument := range call.Call.Arguments() {
		if _, _, ok := mathMinMaxCall(ctx, argument, method); ok {
			return true
		}
	}
	return false
}

func flattenedArguments(ctx rule.RuleContext, call unicornutil.DotMethodCall, method string) []*ast.Node {
	var result []*ast.Node
	for _, argument := range call.Call.Arguments() {
		if _, nested, ok := mathMinMaxCall(ctx, argument, method); ok {
			result = append(result, flattenedArguments(ctx, nested, method)...)
			continue
		}
		result = append(result, argument)
	}
	return result
}

func nestedArgumentsHaveSafeCoercion(
	ctx rule.RuleContext,
	evaluator *utils.StaticStringEvaluator,
	call unicornutil.DotMethodCall,
	method string,
) bool {
	for _, argument := range call.Call.Arguments() {
		if _, nested, ok := mathMinMaxCall(ctx, argument, method); ok &&
			!flattenedCallArgumentsHaveSafeCoercion(ctx, evaluator, nested, method) {
			return false
		}
	}
	return true
}

func flattenedCallArgumentsHaveSafeCoercion(
	ctx rule.RuleContext,
	evaluator *utils.StaticStringEvaluator,
	call unicornutil.DotMethodCall,
	method string,
) bool {
	for _, argument := range call.Call.Arguments() {
		if _, nested, ok := mathMinMaxCall(ctx, argument, method); ok {
			if !flattenedCallArgumentsHaveSafeCoercion(ctx, evaluator, nested, method) {
				return false
			}
			continue
		}
		if _, ok := evaluator.EvalControlFlowValue(argument); !ok ||
			!unicornutil.IsStaticPrimitiveArgument(evaluator, argument) {
			return false
		}
	}
	return true
}

func isNestedInSameMathMinMaxCall(ctx rule.RuleContext, node *ast.Node, method string) bool {
	current := node
	for current.Parent != nil && current.Parent.Kind == ast.KindParenthesizedExpression {
		current = current.Parent
	}

	_, _, ok := mathMinMaxCall(ctx, current.Parent, method)
	return ok
}
