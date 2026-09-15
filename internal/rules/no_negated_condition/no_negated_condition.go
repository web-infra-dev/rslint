package no_negated_condition

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// https://eslint.org/docs/latest/rules/no-negated-condition

var unexpectedNegatedMsg = rule.RuleMessage{
	Id:          "unexpectedNegated",
	Description: "Unexpected negated condition.",
}

// isNegatedTest reports whether test (after peeling parens) is a negated
// unary expression (`!x`) or a negated equality binary expression
// (`x != y` / `x !== y`). Mirrors upstream's isNegatedUnaryExpression /
// isNegatedBinaryExpression.
func isNegatedTest(test *ast.Node) bool {
	if test == nil {
		return false
	}
	inner := ast.SkipParentheses(test)
	switch inner.Kind {
	case ast.KindPrefixUnaryExpression:
		return inner.AsPrefixUnaryExpression().Operator == ast.KindExclamationToken
	case ast.KindBinaryExpression:
		op := inner.AsBinaryExpression().OperatorToken
		if op == nil {
			return false
		}
		return op.Kind == ast.KindExclamationEqualsToken || op.Kind == ast.KindExclamationEqualsEqualsToken
	}
	return false
}

// hasElseWithoutCondition reports whether node has a plain `else` branch,
// i.e. an ElseStatement that is not itself an `else if` IfStatement link.
func hasElseWithoutCondition(ifStmt *ast.IfStatement) bool {
	return ifStmt.ElseStatement != nil && ifStmt.ElseStatement.Kind != ast.KindIfStatement
}

var NoNegatedConditionRule = rule.Rule{
	Name:   "no-negated-condition",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindIfStatement: func(node *ast.Node) {
				ifStmt := node.AsIfStatement()
				if ifStmt == nil || !hasElseWithoutCondition(ifStmt) {
					return
				}
				if isNegatedTest(ifStmt.Expression) {
					ctx.ReportNode(node, unexpectedNegatedMsg)
				}
			},
			ast.KindConditionalExpression: func(node *ast.Node) {
				cond := node.AsConditionalExpression()
				if cond == nil {
					return
				}
				if isNegatedTest(cond.Condition) {
					ctx.ReportNode(node, unexpectedNegatedMsg)
				}
			},
		}
	},
}
