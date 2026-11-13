// Ported from eslint-plugin-unicorn v79.0.0; see LICENSE.
package no_unnecessary_splice

import (
	"math"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

var messages = map[string]string{
	"no-op":   "This `splice()` call does not change the array.",
	"shift":   "Prefer `.shift()` over `.splice()`.",
	"unshift": "Prefer `.unshift()` over `.splice()`.",
	"pop":     "Prefer `.pop()` over `.splice()`.",
	"push":    "Prefer `.push()` over `.splice()`.",
	"empty":   "Prefer setting `.length = 0` over `.splice()` to empty an array.",
}

var NoUnnecessarySpliceRule = rule.Rule{
	Name:   "unicorn/no-unnecessary-splice",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		var effects *utils.StaticStringEvaluator
		return rule.RuleListeners{ast.KindCallExpression: func(node *ast.Node) {
			call, ok := unicornutil.MatchDotMethodCall(node, unicornutil.DotMethodCallOptions{Method: "splice"})
			if !ok || unicornutil.HasOptionalChainElementThroughInstantiation(call.Object) || hasOptionalComputedKey(call.Object) {
				return
			}
			replacement := spliceReplacement(call.Object, node.Arguments())
			if replacement == "" || shouldSkipReceiver(ctx, call.Object) {
				return
			}
			if replacement == "empty" && !isSpliceValueDiscarded(node) {
				return
			}
			message := rule.RuleMessage{Id: replacement, Description: messages[replacement]}
			ctx.ReportNodeWithDeferredFixes(call.Property, message, func() []rule.RuleFix {
				if replacement != "empty" && !isSpliceValueDiscarded(node) {
					return nil
				}
				if effects == nil {
					effects = utils.NewStaticStringEvaluatorWithReferenceResolver(ctx.TypeChecker, ctx.SourceFile, ctx.Refs)
					effects.GlobalAccess = ctx.Globals.Access
				}
				return spliceFixes(ctx, call, replacement, effects)
			})
		}}
	},
}

func shouldSkipReceiver(ctx rule.RuleContext, receiver *ast.Node) bool {
	if !ast.IsInJSFile(receiver) {
		return unicornutil.ShouldSkipKnownNonArrayReceiver(ctx, receiver)
	}
	// Keep Unicorn's diagnostics for receivers written directly as literals.
	node := utils.ESTreeRuntimeExpression(receiver)
	switch node.Kind {
	case ast.KindArrayLiteralExpression, ast.KindObjectLiteralExpression, ast.KindFunctionExpression,
		ast.KindTemplateExpression, ast.KindNoSubstitutionTemplateLiteral:
		return false
	}
	return !utils.IsESTreeLiteralKind(node.Kind) &&
		unicornutil.IsKnownNonIndexedCollectionWithOptions(ctx, receiver, unicornutil.ArrayReceiverOptions{ReadJSDoc: true})
}

const expressionWrappers = ast.OEKParentheses | ast.OEKAssertions | ast.OEKExpressionsWithTypeArguments

func spliceReplacement(receiver *ast.Node, args []*ast.Node) string {
	if len(args) == 0 {
		return "no-op"
	}
	// A spread here can supply the delete count. For example, ...[0, 1]
	// removes an element in splice(...[0, 1], 0).
	if args[0].Kind == ast.KindSpreadElement {
		return ""
	}
	start := staticNumber(args[0])
	startsAtZero := start == 0 && !math.Signbit(start)
	if len(args) == 1 {
		if startsAtZero {
			return "empty"
		}
		return ""
	}
	count := staticNumber(args[1])
	if len(args) == 2 {
		switch {
		case count <= 0:
			return "no-op"
		case startsAtZero && isLengthOf(args[1], receiver):
			return "empty"
		case count == 1:
			if startsAtZero {
				return "shift"
			}
			if isLengthMinusOneOf(args[0], receiver) {
				return "pop"
			}
		}
	} else if count <= 0 {
		if startsAtZero {
			return "unshift"
		}
		if isLengthOf(args[0], receiver) {
			return "push"
		}
	}
	return ""
}

// Unicorn recognizes numeric literals and unary signs, without evaluating
// identifiers or general constant expressions.
func staticNumber(node *ast.Node) float64 {
	node = ast.SkipOuterExpressions(node, expressionWrappers)
	switch node.Kind {
	case ast.KindNumericLiteral:
		if value, ok := ecmascript.StringToNumber(node.Text()); ok {
			return value
		}
	case ast.KindPrefixUnaryExpression:
		unary := node.AsPrefixUnaryExpression()
		switch unary.Operator {
		case ast.KindPlusToken:
			return staticNumber(unary.Operand)
		case ast.KindMinusToken:
			return -staticNumber(unary.Operand)
		}
	}
	return math.NaN()
}

func isLengthOf(node, receiver *ast.Node) bool {
	node = ast.SkipOuterExpressions(node, expressionWrappers)
	return ast.IsPropertyAccessExpression(node) && ast.IsIdentifier(node.Name()) && node.Name().Text() == "length" &&
		!unicornutil.HasOptionalChainElementThroughInstantiation(node) &&
		!hasOptionalComputedKey(node.Expression()) &&
		unicornutil.IsSameComparisonOperand(node.Expression(), receiver)
}

func isLengthMinusOneOf(node, receiver *ast.Node) bool {
	node = ast.SkipOuterExpressions(node, expressionWrappers)
	if !ast.IsBinaryExpression(node) {
		return false
	}
	binary := node.AsBinaryExpression()
	return binary.OperatorToken.Kind == ast.KindMinusToken &&
		staticNumber(binary.Right) == 1 && isLengthOf(binary.Left, receiver)
}

func hasOptionalComputedKey(node *ast.Node) bool {
	for {
		node = ast.SkipOuterExpressions(node, expressionWrappers)
		if !ast.IsAccessExpression(node) && !ast.IsCallExpression(node) {
			return false
		}
		if ast.IsElementAccessExpression(node) &&
			unicornutil.HasOptionalChainElementThroughInstantiation(node.AsElementAccessExpression().ArgumentExpression) {
			return true
		}
		node = node.Expression()
	}
}

func isSpliceValueDiscarded(node *ast.Node) bool {
	for parent := utils.ESTreeParent(node); parent != nil; parent = utils.ESTreeParent(node) {
		if ast.IsExpressionStatement(parent) {
			return true
		}
		if !ast.IsCommaExpression(parent) {
			return false
		}
		if utils.ESTreeRuntimeExpression(parent.AsBinaryExpression().Left) == node {
			return true
		}
		node = parent
	}
	return false
}
