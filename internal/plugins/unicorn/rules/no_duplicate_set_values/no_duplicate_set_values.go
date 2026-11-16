package no_duplicate_set_values

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// element is one checked array element. node is the ESTree element (nil for
// a hole); value holds its comparable static value when known.
type element struct {
	node  *ast.Node
	value any
	known bool
}

// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v80.0.0/rules/no-duplicate-set-values.js
var NoDuplicateSetValuesRule = rule.Rule{
	Name:   "unicorn/no-duplicate-set-values",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		var evaluator *utils.StaticStringEvaluator
		staticValue := func(node *ast.Node) (any, bool) {
			if evaluator == nil {
				evaluator = utils.NewStaticStringEvaluatorWithReferenceResolver(ctx.TypeChecker, ctx.SourceFile, ctx.Refs)
				evaluator.GlobalAccess = ctx.Globals.Access
			}
			value, known := evaluator.EvalValueIfNoSideEffects(node)
			return value, known && utils.IsPrimitiveStaticValue(value)
		}

		return rule.RuleListeners{
			ast.KindNewExpression: func(node *ast.Node) {
				expression := node.AsNewExpression()
				callee := utils.ESTreeRuntimeExpression(expression.Expression)
				if !ast.IsIdentifier(callee) || callee.Text() != "Set" ||
					expression.Arguments == nil || len(expression.Arguments.Nodes) != 1 {
					return
				}
				// A duplicate hole is reported on the authored argument, TypeScript
				// wrappers included; the elements come from the unwrapped array.
				iterable := utils.ESTreeRuntimeExpression(expression.Arguments.Nodes[0])
				array := utils.SkipAssertionsAndParens(iterable)
				if !ast.IsArrayLiteralExpression(array) {
					return
				}

				var checked []element
				for _, child := range array.AsArrayLiteralExpression().Elements.Nodes {
					if ast.IsSpreadElement(child) {
						continue
					}
					current := element{known: true, value: utils.StaticUndefinedValue()}
					if !ast.IsOmittedExpression(child) {
						current.node = utils.ESTreeRuntimeExpression(child)
						current.value, current.known = staticValue(current.node)
					}
					for _, previous := range checked {
						if isDuplicate(previous, current) {
							report(ctx, iterable, current.node)
							break
						}
					}
					checked = append(checked, current)
				}
			},
		}
	},
}

func isDuplicate(left, right element) bool {
	if left.known && right.known {
		equal, _ := utils.StaticValuesSameValueZero(left.value, right.value)
		return equal
	}
	if left.node == nil || right.node == nil || isLiteral(left.node) || isLiteral(right.node) {
		return false
	}
	return unicornutil.IsSameReference(left.node, right.node)
}

// isLiteral reports whether node is an ESTree Literal.
func isLiteral(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindNumericLiteral, ast.KindBigIntLiteral, ast.KindStringLiteral,
		ast.KindRegularExpressionLiteral, ast.KindNullKeyword,
		ast.KindTrueKeyword, ast.KindFalseKeyword:
		return true
	}
	return false
}

func report(ctx rule.RuleContext, iterable, duplicate *ast.Node) {
	target, value := iterable, "undefined"
	if duplicate != nil {
		target = duplicate
		textRange := utils.TrimNodeTextRange(ctx.SourceFile, duplicate)
		value = ctx.SourceFile.Text()[textRange.Pos():textRange.End()]
	}
	ctx.ReportNode(target, rule.RuleMessage{
		Id:          "no-duplicate-set-values",
		Description: "Remove duplicate value `" + value + "` from the Set.",
	})
}
