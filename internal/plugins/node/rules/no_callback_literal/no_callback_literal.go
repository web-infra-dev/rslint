package no_callback_literal

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/rules/no-callback-literal.js
var NoCallbackLiteralRule = rule.Rule{
	Name:   "node/no-callback-literal",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				call := node.AsCallExpression()
				callee := utils.ESTreeCallCallee(call.Expression)
				if callee == nil || callee.Kind != ast.KindIdentifier ||
					(callee.Text() != "callback" && callee.Text() != "cb") ||
					len(call.Arguments.Nodes) == 0 || couldBeError(call.Arguments.Nodes[0]) {
					return
				}
				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "unexpectedLiteral",
					Description: "Unexpected literal in error position of callback.",
				})
			},
		}
	},
}

// Unlike ESLint core's utils.CouldBeError, this rule accepts unknown expressions
// and null, follows every assignment's right side, and checks both logical arms.
// Keep that rule-specific policy here rather than changing the shared predicate.
func couldBeError(node *ast.Node) bool {
	node = utils.ESTreeRuntimeExpression(node)
	if node == nil {
		return true
	}
	switch node.Kind {
	case ast.KindRegularExpressionLiteral:
		// Upstream can accept newer regex syntax when its Node.js runtime cannot
		// instantiate the literal (ESTree value is null). Always treat regexes as
		// non-errors, independently of the runtime used to launch the linter.
		return false
	case ast.KindStringLiteral, ast.KindNumericLiteral, ast.KindBigIntLiteral,
		ast.KindTrueKeyword, ast.KindFalseKeyword,
		ast.KindArrayLiteralExpression, ast.KindObjectLiteralExpression,
		ast.KindNoSubstitutionTemplateLiteral, ast.KindTemplateExpression:
		return false
	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		operator := binary.OperatorToken.Kind
		if ast.IsAssignmentOperator(operator) || operator == ast.KindCommaToken {
			return couldBeError(binary.Right)
		}
		switch operator {
		case ast.KindAmpersandAmpersandToken, ast.KindBarBarToken, ast.KindQuestionQuestionToken:
			return couldBeError(binary.Left) || couldBeError(binary.Right)
		}
	case ast.KindConditionalExpression:
		conditional := node.AsConditionalExpression()
		return couldBeError(conditional.WhenTrue) || couldBeError(conditional.WhenFalse)
	}
	return true
}
