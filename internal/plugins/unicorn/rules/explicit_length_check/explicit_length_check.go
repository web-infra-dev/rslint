// Ported from eslint-plugin-unicorn v76.0.0; see LICENSE.
package explicit_length_check

import (
	_ "embed"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/jsnum"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed explicit_length_check.schema.json
var schemaJSON []byte

var ExplicitLengthCheckRule = rule.Rule{
	Name:   "unicorn/explicit-length-check",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		nonZeroOperator := ast.KindGreaterThanToken
		if len(options) > 0 {
			if option, ok := options[0].(map[string]any); ok && option["non-zero"] == "not-equal" {
				nonZeroOperator = ast.KindExclamationEqualsEqualsToken
			}
		}
		values := lengthValues{ctx: ctx}
		return rule.RuleListeners{
			ast.KindPropertyAccessExpression: func(length *ast.Node) {
				if !isLengthOrSize(length) || utils.ESTreeRuntimeExpression(length.Expression()).Kind == ast.KindThisKeyword {
					return
				}
				var node *ast.Node
				zero, autoFix := false, true
				if comparison, isZero := lengthComparison(length, false); comparison != nil {
					ancestor, negative := unicornutil.BooleanAncestor(ctx, comparison)
					node, zero = ancestor, isZero != negative
				} else {
					ancestor, negative := unicornutil.BooleanAncestor(ctx, length)
					if unicornutil.IsBooleanExpression(ctx, ancestor) || unicornutil.IsControlFlowTest(ancestor) {
						node, zero = ancestor, negative
					} else if isAnd(utils.ESTreeParent(length)) {
						node, zero, autoFix = length, negative, false
					}
				}
				if node == nil {
					return
				}
				operator, code, id, description := nonZeroOperator, "> 0", "non-zero", "is not zero."
				if nonZeroOperator == ast.KindExclamationEqualsEqualsToken {
					code = "!== 0"
				}
				if zero {
					operator, code, id, description = ast.KindEqualsEqualsEqualsToken, "=== 0", "zero", "is zero."
				}
				if compareRight(node, operator, 0) {
					return
				}
				if isGuarded(ctx, node, length) || values.knownNonCollection(length) {
					return
				}
				property := length.Name().Text()
				message := rule.RuleMessage{
					Id: id, Description: "Use `." + property + " " + code + "` when checking " + property + " " + description,
					Data: map[string]string{"property": property, "code": code},
				}
				parent := utils.ESTreeParent(node)
				// Replacing the left operand of `!length > 0` would change the
				// meaning through precedence. Upstream offers no edit here.
				if unicornutil.IsLogicalNot(node) && isBinaryComparisonParent(parent, node) {
					ctx.ReportNode(node, message)
					return
				}
				fix := func() []rule.RuleFix {
					fixed := ctx.SourceFile.Text()[utils.TrimNodeTextRange(ctx.SourceFile, length).Pos():length.End()] + " " + code
					if comparisonNeedsParentheses(node, operator) {
						fixed = "(" + fixed + ")"
					}
					if unicornutil.NeedsSemicolonBefore(ctx.SourceFile, node, fixed) {
						fixed = ";" + fixed
					}
					fixes := []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, node, fixed)}
					return append(fixes, unicornutil.SpaceAroundKeywordFixes(ctx.SourceFile, node)...)
				}
				if autoFix {
					ctx.ReportNodeWithDeferredFixes(node, message, fix)
				} else {
					ctx.ReportNodeWithDeferredSuggestions(node, message, func() []rule.RuleSuggestion {
						return []rule.RuleSuggestion{{
							Message:  rule.RuleMessage{Id: "suggestion", Description: "Replace `." + property + "` with `." + property + " " + code + "`.", Data: message.Data},
							FixesArr: fix(),
						}}
					})
				}
			},
		}
	},
}

func isLengthOrSize(node *ast.Node) bool {
	return node != nil && node.Kind == ast.KindPropertyAccessExpression && !ast.IsOptionalChain(node) &&
		ast.IsIdentifier(node.Name()) && (node.Name().Text() == "length" || node.Name().Text() == "size")
}

func isAnd(node *ast.Node) bool {
	return node != nil && node.Kind == ast.KindBinaryExpression && node.AsBinaryExpression().OperatorToken.Kind == ast.KindAmpersandAmpersandToken
}

func isNumber(node *ast.Node, value jsnum.Number) bool {
	node = utils.ESTreeRuntimeExpression(node)
	return node != nil && node.Kind == ast.KindNumericLiteral && jsnum.FromString(node.Text()) == value
}

func compareRight(node *ast.Node, operator ast.Kind, value jsnum.Number) bool {
	return node != nil && node.Kind == ast.KindBinaryExpression &&
		node.AsBinaryExpression().OperatorToken.Kind == operator && isNumber(node.AsBinaryExpression().Right, value)
}

func lengthComparison(length *ast.Node, allowTypeScript bool) (*ast.Node, bool) {
	parent := utils.ESTreeParent(length)
	if allowTypeScript {
		for {
			if _, ok := utils.TransparentExpression(parent); !ok {
				break
			}
			parent = utils.ESTreeParent(parent)
		}
	}
	if parent == nil || parent.Kind != ast.KindBinaryExpression {
		return nil, false
	}
	comparison := parent.AsBinaryExpression()
	op := comparison.OperatorToken.Kind
	// Normalize Yoda comparisons so only the right-hand literal is classified.
	value := comparison.Right
	if isNumber(comparison.Left, 0) || isNumber(comparison.Left, 1) {
		value = comparison.Left
		switch op {
		case ast.KindGreaterThanToken:
			op = ast.KindLessThanToken
		case ast.KindGreaterThanEqualsToken:
			op = ast.KindLessThanEqualsToken
		case ast.KindLessThanToken:
			op = ast.KindGreaterThanToken
		case ast.KindLessThanEqualsToken:
			op = ast.KindGreaterThanEqualsToken
		}
	}
	if isNumber(value, 0) {
		switch op {
		case ast.KindEqualsEqualsEqualsToken, ast.KindEqualsEqualsToken, ast.KindLessThanEqualsToken:
			return parent, true
		case ast.KindExclamationEqualsEqualsToken, ast.KindExclamationEqualsToken, ast.KindGreaterThanToken:
			return parent, false
		}
	}
	if isNumber(value, 1) {
		switch op {
		case ast.KindLessThanToken:
			return parent, true
		case ast.KindGreaterThanEqualsToken:
			return parent, false
		}
	}
	return nil, false
}

// Only && siblings can establish these guards. Flatten without allocating an
// operand list, and retain explicit coercions as distinct operands.
func someAndOperand(node *ast.Node, predicate func(*ast.Node) bool) bool {
	node = utils.ESTreeRuntimeExpression(node)
	if isAnd(node) {
		binary := node.AsBinaryExpression()
		return someAndOperand(binary.Left, predicate) || someAndOperand(binary.Right, predicate)
	}
	return predicate(node)
}

func isGuarded(ctx rule.RuleContext, node, length *ast.Node) bool {
	root := node
	for isAnd(utils.ESTreeParent(root)) {
		root = utils.ESTreeParent(root)
	}
	if !isAnd(root) {
		return false
	}
	return someAndOperand(root, func(operand *ast.Node) bool {
		if operand == node {
			return false
		}
		if operand.Kind == ast.KindPropertyAccessExpression && !ast.IsOptionalChain(operand) && ast.IsIdentifier(operand.Name()) {
			switch operand.Name().Text() {
			case "depth", "height", "width":
				return unicornutil.IsSameReference(operand.Expression(), length.Expression())
			}
		}
		if node != length {
			return false
		}
		comparison := operand
		for unicornutil.IsLogicalNot(comparison) || unicornutil.IsGlobalBooleanCall(ctx, comparison) {
			if unicornutil.IsLogicalNot(comparison) {
				comparison = utils.ESTreeRuntimeExpression(comparison.AsPrefixUnaryExpression().Operand)
			} else {
				comparison = utils.ESTreeRuntimeExpression(comparison.Arguments()[0])
			}
		}
		if comparison.Kind != ast.KindBinaryExpression {
			return false
		}
		binary := comparison.AsBinaryExpression()
		member := utils.SkipAssertionsAndParens(binary.Left)
		if !isLengthOrSize(member) {
			member = utils.SkipAssertionsAndParens(binary.Right)
		}
		if !isLengthOrSize(member) || !unicornutil.IsSameReference(member, length) {
			return false
		}
		check, zero := lengthComparison(member, true)
		if check == nil {
			return false
		}
		ancestor, negative := unicornutil.BooleanAncestor(ctx, check)
		return ancestor == operand && zero == negative
	})
}

// The replacement is a comparison, even when the original node was a call
// or a negation. Preserve its grouping in the surrounding expression.
func comparisonNeedsParentheses(node *ast.Node, operator ast.Kind) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindBinaryExpression:
		precedence := ast.GetOperatorPrecedence(ast.KindBinaryExpression, operator, ast.OperatorPrecedenceFlagsNone)
		parentPrecedence := ast.GetExpressionPrecedence(parent)
		return precedence < parentPrecedence || (precedence == parentPrecedence && parent.AsBinaryExpression().Right == node)
	case ast.KindPrefixUnaryExpression, ast.KindDeleteExpression, ast.KindTypeOfExpression, ast.KindVoidExpression, ast.KindAwaitExpression,
		ast.KindTypeAssertionExpression, ast.KindNonNullExpression:
		return true
	case ast.KindTaggedTemplateExpression:
		return parent.AsTaggedTemplateExpression().Tag == node
	case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression, ast.KindCallExpression, ast.KindNewExpression,
		ast.KindExpressionWithTypeArguments:
		return parent.Expression() == node
	}
	return false
}

func isBinaryComparisonParent(parent, node *ast.Node) bool {
	if parent == nil || parent.Kind != ast.KindBinaryExpression || utils.ESTreeRuntimeExpression(parent.AsBinaryExpression().Left) != node {
		return false
	}
	op := parent.AsBinaryExpression().OperatorToken.Kind
	return !unicornutil.IsLogicalExpression(parent) && op != ast.KindQuestionQuestionToken && op != ast.KindCommaToken && !ast.IsAssignmentOperator(op)
}
