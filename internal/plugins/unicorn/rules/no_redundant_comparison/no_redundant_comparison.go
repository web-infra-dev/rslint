// Ported from eslint-plugin-unicorn v79.0.0; see LICENSE.
package no_redundant_comparison

import (
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/jsnum"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const expressionWrappers = ast.OEKParentheses | ast.OEKAssertions | ast.OEKExpressionsWithTypeArguments

var NoRedundantComparisonRule = rule.Rule{
	Name:   "unicorn/no-redundant-comparison",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		var evaluator *utils.StaticStringEvaluator
		return rule.RuleListeners{ast.KindBinaryExpression: func(node *ast.Node) {
			if !isLogicalAnd(node) || isLogicalAnd(utils.ESTreeParent(node)) {
				return
			}
			var operands []*ast.Node
			collectOperands(node, &operands)
			conditions := classifyOperands(operands)
			if len(conditions.comparisons) == 0 || (len(conditions.equalities) == 0 && len(conditions.inequalities) == 0) {
				return
			}
			if evaluator == nil {
				evaluator = utils.NewStaticStringEvaluatorWithReferenceResolver(ctx.TypeChecker, ctx.SourceFile, ctx.Refs)
				evaluator.GlobalAccess = ctx.Globals.Access
			}
			for _, operand := range conditions.analyzedOperands {
				if evaluator.HasSideEffect(operand, false) || containsTaggedTemplate(operand) {
					return
				}
			}
			redundant := findRedundantComparison(conditions, evaluator)
			if redundant == nil {
				return
			}

			message := rule.RuleMessage{
				Id: "no-redundant-comparison", Description: "This comparison is redundant because it is already implied by the other conditions.",
			}
			buildFixes := func() []rule.RuleFix {
				parentNode := utils.ESTreeParent(redundant)
				parent := parentNode.AsBinaryExpression()
				isLeft := utils.ESTreeRuntimeExpression(parent.Left) == redundant
				left := utils.TrimNodeTextRange(ctx.SourceFile, parent.Left)
				right := utils.TrimNodeTextRange(ctx.SourceFile, parent.Right)
				start, end := left.End(), right.End()
				if isLeft {
					start, end = left.Pos(), right.Pos()
				}
				if utils.HasCommentInSpan(ctx.Comments.All(), start, end) {
					return nil
				}
				replacement := ""
				if isLeft && unicornutil.NeedsSemicolonBefore(ctx.SourceFile, parentNode, ctx.SourceFile.Text()[end:]) {
					replacement = ";"
				}
				return []rule.RuleFix{rule.RuleFixReplaceRange(core.NewTextRange(start, end), replacement)}
			}
			// Every ordering predicate has a reference operand, so dropping it can
			// drop an implicit coercion. Strict comparisons only need getter checks.
			operator := ast.SkipOuterExpressions(redundant, expressionWrappers).AsBinaryExpression().OperatorToken.Kind
			if isOrdering(operator) || evaluator.HasSideEffect(redundant, true) {
				ctx.ReportNodeWithDeferredSuggestions(redundant, message, func() []rule.RuleSuggestion {
					fixes := buildFixes()
					if len(fixes) == 0 {
						return nil
					}
					return []rule.RuleSuggestion{{
						Message:  rule.RuleMessage{Id: "no-redundant-comparison/suggestion", Description: "Remove this redundant comparison."},
						FixesArr: fixes,
					}}
				})
			} else {
				ctx.ReportNodeWithDeferredFixes(redundant, message, buildFixes)
			}
		}}
	},
}

type comparison struct {
	operator    ast.Kind
	left, right *ast.Node
	node        *ast.Node
}

type conditions struct {
	equalities, inequalities, comparisons []comparison
	analyzedOperands                      []*ast.Node
}

type predicate struct {
	operator    ast.Kind
	value, node *ast.Node
}

type numericBound struct {
	value jsnum.Number
	valid bool
}

func isLogicalAnd(node *ast.Node) bool {
	return node != nil && node.Kind == ast.KindBinaryExpression && node.AsBinaryExpression().OperatorToken.Kind == ast.KindAmpersandAmpersandToken
}

func collectOperands(node *ast.Node, operands *[]*ast.Node) {
	node = utils.ESTreeRuntimeExpression(node)
	if !isLogicalAnd(node) {
		*operands = append(*operands, node)
		return
	}
	// TypeScript wrappers stop flattening, as they do in the upstream ESTree.
	collectOperands(node.AsBinaryExpression().Left, operands)
	collectOperands(node.AsBinaryExpression().Right, operands)
}

func isReference(node *ast.Node) bool {
	node = ast.SkipOuterExpressions(node, expressionWrappers)
	return ast.IsIdentifier(node) || ast.IsAccessExpression(node) || node.Kind == ast.KindThisKeyword || node.Kind == ast.KindSuperKeyword
}

func containsTaggedTemplate(node *ast.Node) bool {
	return node.Kind == ast.KindTaggedTemplateExpression || node.ForEachChild(containsTaggedTemplate)
}

func classifyOperands(operands []*ast.Node) conditions {
	var result conditions
	first, last := -1, -1
	for index, operand := range operands {
		expression := ast.SkipOuterExpressions(operand, expressionWrappers)
		if expression.Kind != ast.KindBinaryExpression || unicornutil.ContainsOptionalChain(expression) {
			continue
		}
		binary := expression.AsBinaryExpression()
		operator := binary.OperatorToken.Kind
		if !unicornutil.IsComparisonOperator(operator) {
			continue
		}
		leftReference, rightReference := isReference(binary.Left), isReference(binary.Right)
		item := comparison{operator: operator, left: binary.Left, right: binary.Right, node: operand}
		if leftReference && rightReference && !isOrdering(operator) {
			if unicornutil.IsSameComparisonOperand(binary.Left, binary.Right) {
				continue
			}
			if operator == ast.KindEqualsEqualsEqualsToken {
				result.equalities = append(result.equalities, item)
			} else {
				result.inequalities = append(result.inequalities, item)
			}
		} else if leftReference || rightReference {
			result.comparisons = append(result.comparisons, item)
		} else {
			continue
		}
		if first == -1 {
			first = index
		}
		last = index
	}
	if first != -1 {
		result.analyzedOperands = operands[first : last+1]
	}
	return result
}

func buildEqualityClasses(equalities []comparison) [][]*ast.Node {
	var classes [][]*ast.Node
	findClass := func(reference *ast.Node) int {
		return slices.IndexFunc(classes, func(members []*ast.Node) bool {
			return slices.ContainsFunc(members, func(member *ast.Node) bool { return unicornutil.IsSameComparisonOperand(member, reference) })
		})
	}
	for _, equality := range equalities {
		left, right := findClass(equality.left), findClass(equality.right)
		switch {
		case left >= 0 && right >= 0:
			if left != right {
				classes[left] = append(classes[left], classes[right]...)
				classes = slices.Delete(classes, right, right+1)
			}
		case left >= 0:
			classes[left] = append(classes[left], equality.right)
		case right >= 0:
			classes[right] = append(classes[right], equality.left)
		default:
			classes = append(classes, []*ast.Node{equality.left, equality.right})
		}
	}
	return classes
}

func predicateForSubject(item comparison, reference *ast.Node) (predicate, bool) {
	if unicornutil.IsSameComparisonOperand(item.left, reference) {
		return predicate{operator: item.operator, value: item.right, node: item.node}, true
	}
	if unicornutil.IsSameComparisonOperand(item.right, reference) {
		return predicate{operator: unicornutil.FlipComparisonOperator(item.operator), value: item.left, node: item.node}, true
	}
	return predicate{}, false
}

func predicateForClass(item comparison, members []*ast.Node) (predicate, bool) {
	left := slices.ContainsFunc(members, func(member *ast.Node) bool { return unicornutil.IsSameComparisonOperand(member, item.left) })
	right := slices.ContainsFunc(members, func(member *ast.Node) bool { return unicornutil.IsSameComparisonOperand(member, item.right) })
	if left == right {
		return predicate{}, false
	}
	if left {
		return predicate{operator: item.operator, value: item.right, node: item.node}, true
	}
	return predicate{operator: unicornutil.FlipComparisonOperator(item.operator), value: item.left, node: item.node}, true
}

func findRedundantComparison(conditions conditions, evaluator *utils.StaticStringEvaluator) *ast.Node {
	var numericBounds map[*ast.Node]numericBound
	for _, members := range buildEqualityClasses(conditions.equalities) {
		var predicates []predicate
		for _, item := range conditions.comparisons {
			if predicate, ok := predicateForClass(item, members); ok {
				predicates = append(predicates, predicate)
			}
		}
		for index, first := range predicates {
			for _, second := range predicates[index+1:] {
				// Source order breaks ties between equivalent predicates.
				if entails(first, second, evaluator, &numericBounds) {
					return second.node
				}
				if entails(second, first, evaluator, &numericBounds) {
					return first.node
				}
			}
		}
	}
	for _, inequality := range conditions.inequalities {
		for _, pair := range [][2]*ast.Node{{inequality.left, inequality.right}, {inequality.right, inequality.left}} {
			var value *ast.Node
			for _, item := range conditions.comparisons {
				if predicate, ok := predicateForSubject(item, pair[0]); ok && predicate.operator == ast.KindEqualsEqualsEqualsToken {
					value = predicate.value
					break
				}
			}
			if value == nil {
				continue
			}
			for _, item := range conditions.comparisons {
				if predicate, ok := predicateForSubject(item, pair[1]); ok && predicate.operator == ast.KindExclamationEqualsEqualsToken && unicornutil.IsSameComparisonOperand(predicate.value, value) {
					return item.node
				}
			}
		}
	}
	return nil
}

func isOrdering(operator ast.Kind) bool {
	return operator == ast.KindGreaterThanToken || operator == ast.KindGreaterThanEqualsToken || operator == ast.KindLessThanToken || operator == ast.KindLessThanEqualsToken
}

// getNumericBound evaluates each bound at most once within the analyzed chain,
// including unknown values. Strict comparisons never allocate this cache.
func getNumericBound(node *ast.Node, evaluator *utils.StaticStringEvaluator, cache *map[*ast.Node]numericBound) numericBound {
	if bound, ok := (*cache)[node]; ok {
		return bound
	}
	value, _ := evaluator.EvalValueIfNoSideEffects(ast.SkipOuterExpressions(node, expressionWrappers))
	number, numeric := value.(jsnum.Number)
	bound := numericBound{value: number, valid: numeric && !number.IsNaN()}
	if *cache == nil {
		*cache = make(map[*ast.Node]numericBound)
	}
	(*cache)[node] = bound
	return bound
}

func entails(first, second predicate, evaluator *utils.StaticStringEvaluator, numericBounds *map[*ast.Node]numericBound) bool {
	if first.operator == second.operator && unicornutil.IsSameComparisonOperand(first.value, second.value) {
		return true
	}
	if !isOrdering(first.operator) || !isOrdering(second.operator) {
		return false
	}
	firstLower := first.operator == ast.KindGreaterThanToken || first.operator == ast.KindGreaterThanEqualsToken
	secondLower := second.operator == ast.KindGreaterThanToken || second.operator == ast.KindGreaterThanEqualsToken
	if firstLower != secondLower {
		return false
	}
	firstBound := getNumericBound(first.value, evaluator, numericBounds)
	secondBound := getNumericBound(second.value, evaluator, numericBounds)
	if !firstBound.valid || !secondBound.valid {
		return false
	}
	a, b := firstBound.value, secondBound.value
	if a != b {
		if firstLower {
			return a > b
		}
		return a < b
	}
	firstInclusive := first.operator == ast.KindGreaterThanEqualsToken || first.operator == ast.KindLessThanEqualsToken
	secondInclusive := second.operator == ast.KindGreaterThanEqualsToken || second.operator == ast.KindLessThanEqualsToken
	return !firstInclusive || secondInclusive
}
