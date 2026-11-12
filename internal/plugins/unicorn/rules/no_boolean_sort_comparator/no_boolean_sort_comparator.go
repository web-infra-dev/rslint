// Ported from eslint-plugin-unicorn v79.0.0; see LICENSE.
package no_boolean_sort_comparator

import (
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const typeScriptWrappers = ast.OEKAssertions | ast.OEKExpressionsWithTypeArguments

var comparatorMessage = rule.RuleMessage{
	Id:          "no-boolean-sort-comparator",
	Description: "Do not return a boolean from a sort comparator.",
}

var NoBooleanSortComparatorRule = rule.Rule{
	Name:   "unicorn/no-boolean-sort-comparator",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		booleans := unicornutil.NewBooleanAnalyzer(&ctx)
		return rule.RuleListeners{ast.KindCallExpression: func(node *ast.Node) {
			args := node.Arguments()
			if len(args) == 0 || ast.IsSpreadElement(args[0]) {
				return
			}
			callee := utils.ESTreeCallCallee(node.Expression())
			if callee == nil || !ast.IsAccessExpression(callee) {
				return
			}
			name, known := utils.AccessExpressionStaticName(callee)
			if !known && ast.IsElementAccessExpression(callee) {
				name, _ = booleans.StaticEvaluator().EvalAccessExpressionName(callee)
			}
			if name != "sort" && name != "toSorted" {
				return
			}
			comparator := utils.ESTreeRuntimeExpression(args[0])
			value, assertedBoolean := unwrapComparator(booleans, comparator)
			if value == nil {
				return
			}
			var result *ast.Node
			inline := ast.IsArrowFunction(value) || ast.IsFunctionExpression(value)
			if inline {
				result = unicornutil.SingleReturnExpression(value)
				if !assertedBoolean && !booleans.IsFunction(value, result) {
					return
				}
			} else if !assertedBoolean && !booleans.IsFunctionReference(value) {
				return
			}
			receiver := callee.Expression()
			if unicornutil.IsKnownNonIndexedCollectionWithOptions(ctx, receiver, unicornutil.ArrayReceiverOptions{ReadJSDoc: true}) {
				return
			}
			if !inline || comparator != value || result == nil {
				ctx.ReportNode(comparator, comparatorMessage)
				return
			}
			ctx.ReportNodeWithDeferredSuggestions(comparator, comparatorMessage, func() []rule.RuleSuggestion {
				return comparatorSuggestion(ctx, comparator, result, receiver)
			})
		}}
	},
}

func unwrapComparator(booleans *unicornutil.BooleanAnalyzer, node *ast.Node) (*ast.Node, bool) {
	assertedBoolean := false
	for node != nil && ast.IsOuterExpression(node, typeScriptWrappers) {
		if annotation := node.Type(); !assertedBoolean && annotation != nil {
			assertedBoolean = booleans.IsFunctionType(annotation)
		}
		node = utils.ESTreeRuntimeExpression(node.Expression())
	}
	return node, assertedBoolean
}

func comparatorSuggestion(ctx rule.RuleContext, comparator, result, receiver *ast.Node) []rule.RuleSuggestion {
	result = utils.ESTreeRuntimeExpression(result)
	if result == nil || !ast.IsBinaryExpression(result) {
		// A replacement would remove an authored assertion around the result.
		return nil
	}
	comparison := result.AsBinaryExpression()
	operator := comparison.OperatorToken.Kind
	if operator != ast.KindGreaterThanToken && operator != ast.KindGreaterThanEqualsToken &&
		operator != ast.KindLessThanToken && operator != ast.KindLessThanEqualsToken {
		return nil
	}
	parameters := comparator.Parameters()
	if len(parameters) != 2 || len(comparator.TypeParameters()) > 0 || comparator.Type() != nil ||
		unicornutil.CallableReturnTypeAnnotation(ctx, comparator) != nil {
		return nil
	}
	first, second := unicornutil.PlainParameterIdentifier(parameters[0]), unicornutil.PlainParameterIdentifier(parameters[1])
	if first == nil || second == nil || first.Text() == second.Text() ||
		parameters[0].Type() != nil || parameters[1].Type() != nil ||
		ast.HasQuestionToken(parameters[0]) || ast.HasQuestionToken(parameters[1]) {
		return nil
	}
	left, right := utils.ESTreeRuntimeExpression(comparison.Left), utils.ESTreeRuntimeExpression(comparison.Right)
	if left == nil || right == nil || ast.IsOuterExpression(left, typeScriptWrappers) || ast.IsOuterExpression(right, typeScriptWrappers) ||
		!mirrorParameters(left, right, first, second) {
		return nil
	}
	if ast.IsFunctionExpression(comparator) {
		name := comparator.Name()
		if name != nil && (name.Text() == first.Text() || name.Text() == second.Text()) {
			name = nil // The parameter shadows the function name.
		}
		if usesFunctionBinding(result, name) {
			// Turning the function into an arrow would change these bindings.
			return nil
		}
	}
	if hasStringOrBigIntOperand(ctx, left, right, receiver) {
		return nil
	}
	span := utils.TrimNodeTextRange(ctx.SourceFile, comparator)
	if utils.HasCommentInSpan(ctx.Comments.All(), span.Pos(), span.End()) {
		return nil
	}
	if operator == ast.KindLessThanToken || operator == ast.KindLessThanEqualsToken {
		left, right = right, left
	}
	replacement := "(" + first.Text() + ", " + second.Text() + ") => " +
		utils.TrimmedNodeText(ctx.SourceFile, left) + " - " + utils.TrimmedNodeText(ctx.SourceFile, right)
	return []rule.RuleSuggestion{{
		Message: rule.RuleMessage{
			Id:          "no-boolean-sort-comparator/suggestion",
			Description: "Replace with `" + replacement + "`.",
			Data:        map[string]string{"replacement": replacement},
		},
		FixesArr: []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, comparator, replacement)},
	}}
}

// Subtracting strings changes their ordering. BigInt subtraction returns a
// BigInt, which sort cannot convert to a Number.
func hasStringOrBigIntOperand(ctx rule.RuleContext, left, right, receiver *ast.Node) bool {
	if types := ctx.TypeChecker; types != nil {
		const nonNumeric = checker.TypeFlagsStringLike | checker.TypeFlagsBigIntLike
		if utils.IsTypeFlagSetWithUnion(types.GetTypeAtLocation(left), nonNumeric) ||
			utils.IsTypeFlagSetWithUnion(types.GetTypeAtLocation(right), nonNumeric) {
			return true
		}
	}
	if !ast.IsIdentifier(left) || !ast.IsIdentifier(right) {
		return false
	}
	receiver = utils.ESTreeRuntimeExpression(receiver)
	if receiver == nil {
		return false
	}
	if ast.IsArrayLiteralExpression(receiver) {
		return slices.ContainsFunc(receiver.Elements(), func(element *ast.Node) bool {
			element = utils.ESTreeRuntimeExpression(element)
			return element != nil && (ast.IsStringLiteralLike(element) || ast.IsTemplateExpression(element) || ast.IsBigIntLiteral(element))
		})
	}
	if ast.IsNewExpression(receiver) {
		constructor := utils.ESTreeCallCallee(receiver.Expression())
		return constructor != nil && ast.IsIdentifier(constructor) &&
			(constructor.Text() == "BigInt64Array" || constructor.Text() == "BigUint64Array") &&
			unicornutil.IsGlobalReference(ctx, constructor)
	}
	return false
}

func mirrorParameters(left, right, first, second *ast.Node) bool {
	for left != nil && right != nil {
		left = ast.SkipOuterExpressions(left, ast.OEKParentheses|typeScriptWrappers)
		right = ast.SkipOuterExpressions(right, ast.OEKParentheses|typeScriptWrappers)
		if ast.IsIdentifier(left) && ast.IsIdentifier(right) {
			return (left.Text() == first.Text() && right.Text() == second.Text()) ||
				(left.Text() == second.Text() && right.Text() == first.Text())
		}
		if left.Kind != right.Kind || !ast.IsAccessExpression(left) ||
			ast.IsOptionalChain(left) || ast.IsOptionalChain(right) {
			return false
		}
		leftObject, leftKey := utils.MemberExpressionParts(left)
		rightObject, rightKey := utils.MemberExpressionParts(right)
		if !unicornutil.IsSameComparisonOperand(leftKey, rightKey) {
			return false
		}
		left, right = leftObject, rightObject
	}
	return false
}

func usesFunctionBinding(node, functionName *ast.Node) bool {
	if node.Kind == ast.KindThisKeyword || node.Kind == ast.KindMetaProperty {
		return true
	}
	if ast.IsIdentifier(node) && (node.Text() == "arguments" ||
		functionName != nil && node.Text() == functionName.Text()) && !utils.IsNonReferenceIdentifier(node) {
		return true
	}
	return node.ForEachChild(func(child *ast.Node) bool {
		return usesFunctionBinding(child, functionName)
	})
}
