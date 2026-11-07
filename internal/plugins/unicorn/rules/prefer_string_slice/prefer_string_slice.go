package prefer_string_slice

import (
	"math"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/jsnum"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

var (
	substrMessage    = rule.RuleMessage{Id: "substr", Description: "Prefer `String#slice()` over `String#substr()`."}
	substringMessage = rule.RuleMessage{Id: "substring", Description: "Prefer `String#slice()` over `String#substring()`."}
)

// PreferStringSliceRule follows eslint-plugin-unicorn v79.0.0.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/rules/prefer-string-slice.js
var PreferStringSliceRule = rule.Rule{
	Name:   "unicorn/prefer-string-slice",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		var evaluator *utils.StaticStringEvaluator
		return rule.RuleListeners{ast.KindCallExpression: func(node *ast.Node) {
			call, ok := unicornutil.MatchDotMethodCall(node, unicornutil.DotMethodCallOptions{
				Methods: []string{"substr", "substring"}, AllowOptionalCall: true, AllowOptionalMember: true,
			})
			if !ok || unicornutil.IsKnownNonStringSyntax(ctx, call.Object) {
				return
			}
			method := call.Property.Text()
			if method == "substring" && isSingleCharacterSubstring(node.Arguments()) {
				return
			}
			message := substrMessage
			if method == "substring" {
				message = substringMessage
			}
			ctx.ReportNodeWithDeferredFixes(node, message, func() []rule.RuleFix {
				fixes := []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, call.Property, "slice")}
				args := node.Arguments()
				if len(args) == 0 {
					return fixes
				}
				if len(args) > 2 {
					return nil
				}
				for _, arg := range args {
					if arg.Kind == ast.KindSpreadElement {
						return nil
					}
				}
				if evaluator == nil {
					evaluator = utils.NewStaticStringEvaluatorWithReferenceResolver(ctx.TypeChecker, ctx.SourceFile, ctx.Refs)
					evaluator.GlobalAccess = ctx.Globals.Access
				}

				// Replacing an argument includes its parentheses; abort the whole
				// fix if that replacement would discard a comment.
				aborted := false
				replace := func(arg *ast.Node, text string) {
					span := utils.TrimNodeTextRange(ctx.SourceFile, arg)
					if utils.HasCommentInSpan(ctx.Comments.All(), span.Pos(), span.End()) {
						aborted = true
						return
					}
					fixes = append(fixes, rule.RuleFixReplace(ctx.SourceFile, arg, text))
				}
				if method == "substr" {
					if !fixSubstrArguments(ctx, call, evaluator, replace, &fixes) {
						return nil
					}
				} else if !fixSubstringArguments(ctx, args, evaluator, replace, &fixes) {
					return nil
				}
				if aborted {
					return nil
				}
				return fixes
			})
		}}
	},
}

func fixSubstrArguments(ctx rule.RuleContext, call unicornutil.DotMethodCall, evaluator *utils.StaticStringEvaluator, replace func(*ast.Node, string), fixes *[]rule.RuleFix) bool {
	args := call.Call.Arguments()
	if len(args) == 1 {
		return true
	}
	first, second := utils.ESTreeRuntimeExpression(args[0]), utils.ESTreeRuntimeExpression(args[1])
	if value, known := evaluator.EvalValueIfNoSideEffects(first); known && value == jsnum.Number(0) {
		if negative := negativeIndex(ctx, args[1], call.Object); negative != "" {
			replace(args[1], negative)
			return true
		}
		if second.Kind == ast.KindNumericLiteral || isLengthProperty(second) {
			return true
		}
		if number, known := numericValue(second); known {
			replace(args[1], ecmascript.NumberToString(math.Max(0, number)))
			return true
		}
		text, ok := coalescedEndText(ctx, args[1], evaluator)
		if !ok {
			return false
		}
		// The original expression, including its comments, is retained twice.
		*fixes = append(*fixes, rule.RuleFixReplace(ctx.SourceFile, args[1], "Math.max(0, "+text+")"))
		return true
	}
	if first.Kind != ast.KindNumericLiteral || second.Kind != ast.KindNumericLiteral {
		return false
	}
	// substr truncates each argument before adding the start and length.
	start, _ := numericValue(first)
	length, _ := numericValue(second)
	replace(args[1], ecmascript.NumberToString(math.Trunc(start)+math.Trunc(length)))
	return true
}

func fixSubstringArguments(ctx rule.RuleContext, args []*ast.Node, evaluator *utils.StaticStringEvaluator, replace func(*ast.Node, string), fixes *[]rule.RuleFix) bool {
	first := utils.ESTreeRuntimeExpression(args[0])
	start, knownStart := numericValue(first)
	if len(args) == 1 {
		if isLengthProperty(first) {
			return true
		}
		if knownStart {
			replace(args[0], ecmascript.NumberToString(math.Max(0, start)))
		} else {
			// Wrapping preserves comments inside the existing parentheses.
			*fixes = append(*fixes, rule.RuleFixReplace(ctx.SourceFile, args[0], "Math.max(0, "+utils.TrimmedNodeText(ctx.SourceFile, args[0])+")"))
		}
		return true
	}
	end, knownEnd := numericValue(utils.ESTreeRuntimeExpression(args[1]))
	if knownStart && knownEnd {
		newStart, newEnd := math.Max(0, start), math.Max(0, end)
		if start > end {
			newStart, newEnd = newEnd, newStart
		}
		if newStart != start {
			replace(args[0], ecmascript.NumberToString(newStart))
		}
		if newEnd != end {
			replace(args[1], ecmascript.NumberToString(newEnd))
		}
		return true
	}
	if (knownStart && start == 0) || (knownEnd && end == 0) {
		var text string
		if knownEnd && end == 0 {
			// A moved start is evaluated once and keeps undefined -> 0.
			text = utils.TrimmedNodeText(ctx.SourceFile, args[0])
		} else {
			var ok bool
			text, ok = coalescedEndText(ctx, args[1], evaluator)
			if !ok {
				return false
			}
		}
		replace(args[0], "0")
		replace(args[1], "Math.max(0, "+text+")")
		return true
	}
	return false
}

func coalescedEndText(ctx rule.RuleContext, arg *ast.Node, evaluator *utils.StaticStringEvaluator) (string, bool) {
	if evaluator.HasSideEffect(arg, false) || containsSpreadOrTag(arg) {
		return "", false
	}
	text := utils.TrimmedNodeText(ctx.SourceFile, arg)
	node := utils.ESTreeRuntimeExpression(arg)
	if arg == node && ast.GetExpressionPrecedence(node) < ast.OperatorPrecedenceAdditive {
		text = "(" + text + ")"
	}
	// Missing/undefined ends mean "to the end"; null still means zero.
	return text + " === undefined ? Infinity : " + text, true
}

func containsSpreadOrTag(node *ast.Node) bool {
	if node.Kind == ast.KindSpreadElement || node.Kind == ast.KindSpreadAssignment || node.Kind == ast.KindTaggedTemplateExpression {
		return true
	}
	return node.ForEachChild(containsSpreadOrTag)
}

func numericValue(node *ast.Node) (float64, bool) {
	node = utils.ESTreeRuntimeExpression(node)
	if node.Kind == ast.KindNumericLiteral {
		return float64(jsnum.FromString(node.Text())), true
	}
	if node.Kind == ast.KindPrefixUnaryExpression && node.AsPrefixUnaryExpression().Operator == ast.KindMinusToken {
		value, known := numericValue(node.AsPrefixUnaryExpression().Operand)
		return -value, known
	}
	return 0, false
}

func isLengthProperty(node *ast.Node) bool {
	node = utils.ESTreeRuntimeExpression(node)
	return ast.IsPropertyAccessExpression(node) && !ast.IsOptionalChain(node) &&
		ast.IsIdentifier(node.Name()) && node.Name().Text() == "length"
}

func isSafeIntegerLiteral(node *ast.Node) bool {
	node = utils.ESTreeRuntimeExpression(node)
	if node.Kind != ast.KindNumericLiteral {
		return false
	}
	value, _ := numericValue(node)
	return value >= 0 && value <= 9007199254740991 && math.Trunc(value) == value
}

func negativeIndex(ctx rule.RuleContext, arg, receiver *ast.Node) string {
	node := utils.ESTreeRuntimeExpression(arg)
	receiver = utils.ESTreeRuntimeExpression(receiver)
	if node.Kind != ast.KindBinaryExpression {
		return ""
	}
	binary := node.AsBinaryExpression()
	right := utils.ESTreeRuntimeExpression(binary.Right)
	value, _ := numericValue(right)
	if binary.OperatorToken.Kind != ast.KindMinusToken || !isSafeIntegerLiteral(right) || value == 0 || !isLengthProperty(binary.Left) {
		return ""
	}
	_, staticString := utils.GetStaticStringLiteralValue(receiver)
	if (!ast.IsIdentifier(receiver) && !staticString) || !unicornutil.IsSameComparisonOperand(utils.ESTreeRuntimeExpression(binary.Left).Expression(), receiver) {
		return ""
	}
	// arg.Pos starts after the previous token, before its leading comments.
	if utils.HasCommentInSpan(ctx.Comments.All(), arg.Pos(), arg.End()) {
		return ""
	}
	return "-" + utils.TrimmedNodeText(ctx.SourceFile, right)
}

func isSingleCharacterSubstring(args []*ast.Node) bool {
	if len(args) != 2 {
		return false
	}
	first, second := utils.ESTreeRuntimeExpression(args[0]), utils.ESTreeRuntimeExpression(args[1])
	if isSafeIntegerLiteral(first) && isSafeIntegerLiteral(second) {
		start, _ := numericValue(first)
		end, _ := numericValue(second)
		return math.Abs(start-end) == 1
	}
	plusOne := func(node, index *ast.Node) bool {
		if node.Kind != ast.KindBinaryExpression || node.AsBinaryExpression().OperatorToken.Kind != ast.KindPlusToken {
			return false
		}
		binary := node.AsBinaryExpression()
		isOne := func(node *ast.Node) bool {
			node = utils.ESTreeRuntimeExpression(node)
			value, _ := numericValue(node)
			return node.Kind == ast.KindNumericLiteral && value == 1
		}
		return (unicornutil.IsSameComparisonOperand(binary.Left, index) && isOne(binary.Right)) ||
			(isOne(binary.Left) && unicornutil.IsSameComparisonOperand(binary.Right, index))
	}
	if plusOne(second, first) || plusOne(first, second) {
		return true
	}
	if first.Kind == ast.KindBinaryExpression {
		binary := first.AsBinaryExpression()
		right := utils.ESTreeRuntimeExpression(binary.Right)
		value, _ := numericValue(right)
		return binary.OperatorToken.Kind == ast.KindMinusToken && right.Kind == ast.KindNumericLiteral && value == 1 && unicornutil.IsSameComparisonOperand(binary.Left, second)
	}
	return false
}
