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
