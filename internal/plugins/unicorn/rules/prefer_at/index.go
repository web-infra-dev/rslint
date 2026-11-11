package prefer_at

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const runtimeWrappers = ast.OEKParentheses | ast.OEKAssertions | ast.OEKExpressionsWithTypeArguments

func unwrapped(node *ast.Node) *ast.Node {
	if node == nil {
		return nil
	}
	return ast.SkipOuterExpressions(node, runtimeWrappers)
}

func negativeIndexLength(index, receiver *ast.Node) *ast.Node {
	for expression := unwrapped(index); expression != nil && ast.IsBinaryExpression(expression); {
		binary := expression.AsBinaryExpression()
		if binary.OperatorToken.Kind != ast.KindMinusToken {
			break
		}
		value, ok := unicornutil.SafeIntegerLiteral(unwrapped(binary.Right))
		if !ok || value == 0 {
			break
		}
		expression = unwrapped(binary.Left)
		if expression != nil && ast.IsPropertyAccessExpression(expression) && !ast.IsOptionalChain(expression) &&
			ast.IsIdentifier(expression.Name()) && expression.Name().Text() == "length" &&
			unicornutil.IsSameComparisonOperand(expression.Expression(), receiver) {
			return binary.Left
		}
	}
	return nil
}

func negativeIntegerLiteral(node *ast.Node) (float64, bool) {
	node = utils.ESTreeRuntimeExpression(node)
	if node == nil || !ast.IsPrefixUnaryExpression(node) || node.AsPrefixUnaryExpression().Operator != ast.KindMinusToken {
		return 0, false
	}
	value, ok := unicornutil.SafeIntegerLiteral(node.AsPrefixUnaryExpression().Operand)
	return -value, ok && value > 0
}

func isArguments(node *ast.Node) bool {
	node = unwrapped(node)
	return node != nil && ast.IsIdentifier(node) && node.Text() == "arguments"
}

func isReferenceContext(node *ast.Node) bool {
	// A call or tag uses the access as its receiver, and an assignment,
	// update or delete requires a reference. TS wrappers are erased.
	for node.Parent != nil && ast.IsOuterExpression(node.Parent, runtimeWrappers) && node.Parent.Expression() == node {
		node = node.Parent
	}
	if utils.IsWriteReference(node) {
		return true
	}
	parent := node.Parent
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindCallExpression:
		return ast.IsCallExpressionTarget(node, false, false)
	case ast.KindDeleteExpression:
		return parent.Expression() == node
	case ast.KindTaggedTemplateExpression:
		return ast.IsTaggedTemplateTag(node, false, false)
	}
	return false
}

func inUnparenthesizedConstructorCallee(node *ast.Node) bool {
	for node.Parent != nil {
		parent := node.Parent
		switch {
		case ast.IsNewExpression(parent):
			return ast.IsNewExpressionTarget(node, false, false)
		case ast.IsTaggedTemplateExpression(parent):
			if !ast.IsTaggedTemplateTag(node, false, false) {
				return false
			}
		case ast.IsAccessExpression(parent), ast.IsOuterExpression(parent, runtimeWrappers&^ast.OEKParentheses):
			if parent.Expression() != node {
				return false
			}
		default:
			return false
		}
		node = parent
	}
	return false
}
