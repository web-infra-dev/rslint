package no_dynamic_delete

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func buildDynamicDeleteMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "dynamicDelete",
		Description: "Do not delete dynamically computed property keys.",
	}
}

func isAllowedDeleteArgument(argument *ast.Node) bool {
	argument = ast.SkipParentheses(argument)

	switch argument.Kind {
	case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral, ast.KindNumericLiteral:
		return true
	case ast.KindPrefixUnaryExpression:
		unary := argument.AsPrefixUnaryExpression()
		return unary != nil && unary.Operator == ast.KindMinusToken && unary.Operand.Kind == ast.KindNumericLiteral
	default:
		return false
	}
}

var NoDynamicDeleteRule = rule.CreateRule(rule.Rule{
	Name: "no-dynamic-delete",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindDeleteExpression: func(node *ast.Node) {
				deleteExpr := node.AsDeleteExpression()
				if deleteExpr == nil {
					return
				}

				expression := ast.SkipParentheses(deleteExpr.Expression)
				if !ast.IsElementAccessExpression(expression) {
					return
				}

				elementAccess := expression.AsElementAccessExpression()
				if elementAccess == nil || elementAccess.ArgumentExpression == nil {
					return
				}

				if isAllowedDeleteArgument(elementAccess.ArgumentExpression) {
					return
				}

				ctx.ReportNode(elementAccess.ArgumentExpression, buildDynamicDeleteMessage())
			},
		}
	},
})
