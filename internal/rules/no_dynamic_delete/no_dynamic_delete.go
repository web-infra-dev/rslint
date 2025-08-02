package no_dynamic_delete

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

var NoDynamicDeleteRule = rule.Rule{
	Name: "no-dynamic-delete",
	Run: func(ctx rule.RuleContext, _ any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindDeleteExpression: func(node *ast.Node) {
				deleteExpr := node.AsDeleteExpression()
				expression := deleteExpr.Expression
				
				// Check if the expression is a MemberExpression with computed property
				if !ast.IsElementAccessExpression(expression) {
					return
				}
				
				elementAccess := expression.AsElementAccessExpression()
				argumentExpression := elementAccess.ArgumentExpression
				
				if argumentExpression == nil {
					return
				}
				
				// Check if the index expression is acceptable
				if isAcceptableIndexExpression(argumentExpression) {
					return
				}
				
				// Report the error on the property/index expression
				ctx.ReportNode(argumentExpression, rule.RuleMessage{
					Id:          "dynamicDelete",
					Description: "Do not delete dynamically computed property keys.",
				})
			},
		}
	},
}

// isAcceptableIndexExpression checks if the property expression is a literal string/number
// or a negative number literal
func isAcceptableIndexExpression(property *ast.Node) bool {
	switch property.Kind {
	case ast.KindStringLiteral:
		// String literals are acceptable
		return true
	case ast.KindNumericLiteral:
		// Number literals are acceptable
		return true
	case ast.KindPrefixUnaryExpression:
		// Check for negative number literals (-7)
		unary := property.AsPrefixUnaryExpression()
		if unary.Operator == ast.KindMinusToken &&
			ast.IsNumericLiteral(unary.Operand) {
			return true
		}
		return false
	default:
		return false
	}
}