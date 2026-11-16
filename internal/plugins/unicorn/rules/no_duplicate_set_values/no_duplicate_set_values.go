package no_duplicate_set_values

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// typeScriptWrappers mirrors upstream unwrapTypeScriptExpression: parentheses,
// `as`, `<T>x`, `!`, `satisfies` and instantiation expressions such as `foo<T>`.
const typeScriptWrappers = ast.OEKParentheses | ast.OEKAssertions | ast.OEKExpressionsWithTypeArguments

// element is one checked array element. node is the ESTree element (nil for
// a hole), reported as authored; value is node without TypeScript wrappers,
// which upstream compares through; key is the SameValueZero key of its comparable static value when
// known.
type element struct {
	node, value *ast.Node
	key         any
	known       bool
}

// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v80.0.0/rules/no-duplicate-set-values.js
var NoDuplicateSetValuesRule = rule.Rule{
	Name:   "unicorn/no-duplicate-set-values",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		var evaluator *utils.StaticStringEvaluator
		// staticValue returns the SameValueZero key of a side-effect-free
		// comparable primitive (upstream isComparableStaticValue).
		staticValue := func(node *ast.Node) (any, bool) {
			if evaluator == nil {
				evaluator = utils.NewStaticStringEvaluatorWithReferenceResolver(ctx.TypeChecker, ctx.SourceFile, ctx.Refs)
				evaluator.GlobalAccess = ctx.Globals.Access
			}
			value, known := evaluator.EvalValueIfNoSideEffects(node)
			if !known {
				return nil, false
			}
			return utils.StaticValueSameValueZeroKey(value)
		}
		undefinedKey, _ := utils.StaticValueSameValueZeroKey(utils.StaticUndefinedValue())

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
				array := ast.SkipOuterExpressions(iterable, typeScriptWrappers)
				if !ast.IsArrayLiteralExpression(array) {
					return
				}

				// Known primitives are indexed by their SameValueZero key, so a
				// distinct static value costs O(1). Same-reference comparison only
				// runs when one side has no static value: a known element checks
				// the unresolved references, an unresolved one checks all.
				var checked comparisons
				for _, child := range array.AsArrayLiteralExpression().Elements.Nodes {
					if ast.IsSpreadElement(child) {
						continue
					}
					current := element{known: true, key: undefinedKey}
					if !ast.IsOmittedExpression(child) {
						current.node = utils.ESTreeRuntimeExpression(child)
						current.value = ast.SkipOuterExpressions(current.node, typeScriptWrappers)
						current.key, current.known = staticValue(current.value)
					}
					if checked.isDuplicate(current) {
						report(ctx, iterable, current.node)
					}
					checked.add(current)
				}
			},
		}
	},
}

// comparisons holds the earlier elements of one array literal.
type comparisons struct {
	keys map[any]struct{}
	// references are the earlier non-Literal expressions; unresolved is the
	// subset without a static value.
	references, unresolved []*ast.Node
}

func (c *comparisons) add(current element) {
	if current.known {
		if c.keys == nil {
			c.keys = make(map[any]struct{})
		}
		c.keys[current.key] = struct{}{}
	}
	if current.node == nil || isLiteral(current.node) {
		return
	}
	c.references = append(c.references, current.value)
	if !current.known {
		c.unresolved = append(c.unresolved, current.value)
	}
}

// isDuplicate mirrors upstream isDuplicateValue against every earlier element:
// two static values compare with SameValueZero, otherwise two non-Literal
// expressions compare by reference.
func (c *comparisons) isDuplicate(current element) bool {
	candidates := c.references
	if current.known {
		if _, ok := c.keys[current.key]; ok {
			return true
		}
		candidates = c.unresolved
	}
	if current.node == nil || isLiteral(current.node) {
		return false
	}
	for _, previous := range candidates {
		if unicornutil.IsSameReference(previous, current.value) {
			return true
		}
	}
	return false
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
