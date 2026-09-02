package no_cond_assign

import (
	_ "embed"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

//go:embed no_cond_assign.schema.json
var schemaJSON []byte

var (
	missingMessage = rule.RuleMessage{
		Id:          "missing",
		Description: "Expected a conditional expression and instead saw an assignment.",
	}
	unexpectedIfMessage = rule.RuleMessage{
		Id:          "unexpected",
		Description: "Unexpected assignment within an 'if' statement.",
		Data:        map[string]string{"type": "an 'if' statement"},
	}
	unexpectedWhileMessage = rule.RuleMessage{
		Id:          "unexpected",
		Description: "Unexpected assignment within a 'while' statement.",
		Data:        map[string]string{"type": "a 'while' statement"},
	}
	unexpectedDoWhileMessage = rule.RuleMessage{
		Id:          "unexpected",
		Description: "Unexpected assignment within a 'do...while' statement.",
		Data:        map[string]string{"type": "a 'do...while' statement"},
	}
	unexpectedForMessage = rule.RuleMessage{
		Id:          "unexpected",
		Description: "Unexpected assignment within a 'for' statement.",
		Data:        map[string]string{"type": "a 'for' statement"},
	}
	unexpectedConditionalMessage = rule.RuleMessage{
		Id:          "unexpected",
		Description: "Unexpected assignment within ConditionalExpression.",
		Data:        map[string]string{"type": "ConditionalExpression"},
	}
)

func unexpectedMessage(conditional *ast.Node) rule.RuleMessage {
	switch conditional.Kind {
	case ast.KindIfStatement:
		return unexpectedIfMessage
	case ast.KindWhileStatement:
		return unexpectedWhileMessage
	case ast.KindDoStatement:
		return unexpectedDoWhileMessage
	case ast.KindForStatement:
		return unexpectedForMessage
	case ast.KindConditionalExpression:
		return unexpectedConditionalMessage
	default:
		return unexpectedConditionalMessage
	}
}

func isAssignmentExpression(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindBinaryExpression {
		return false
	}
	binary := node.AsBinaryExpression()
	return binary != nil && binary.OperatorToken != nil && ast.IsAssignmentOperator(binary.OperatorToken.Kind)
}

// conditionalTest returns a conditional node's test expression and the number
// of explicit parenthesis wrappers needed to permit a top-level assignment in
// the default mode. Statement grammar supplies ESLint's outer parenthesis pair,
// while a ConditionalExpression has no grammar parentheses and therefore needs
// two explicit pairs.
func conditionalTest(node *ast.Node) (test *ast.Node, requiredParenDepth int) {
	if node == nil {
		return nil, 0
	}
	switch node.Kind {
	case ast.KindIfStatement:
		if statement := node.AsIfStatement(); statement != nil {
			return statement.Expression, 1
		}
	case ast.KindWhileStatement:
		if statement := node.AsWhileStatement(); statement != nil {
			return statement.Expression, 1
		}
	case ast.KindDoStatement:
		if statement := node.AsDoStatement(); statement != nil {
			return statement.Expression, 1
		}
	case ast.KindForStatement:
		if statement := node.AsForStatement(); statement != nil {
			return statement.Condition, 1
		}
	case ast.KindConditionalExpression:
		if expression := node.AsConditionalExpression(); expression != nil {
			return expression.Condition, 2
		}
	}
	return nil, 0
}

func unwrapParentheses(node *ast.Node) (expression *ast.Node, depth int) {
	expression = node
	for expression != nil && expression.Kind == ast.KindParenthesizedExpression {
		parenthesized := expression.AsParenthesizedExpression()
		if parenthesized == nil || parenthesized.Expression == nil {
			break
		}
		expression = parenthesized.Expression
		depth++
	}
	return expression, depth
}

// findConditionalAncestor mirrors ESLint's bottom-up search: a conditional is
// selected only when the current child is exactly its test slot. In particular,
// assignments in ternary branches are not attributed to that ternary, though
// the whole ternary can itself be the test of an outer conditional.
func findConditionalAncestor(node *ast.Node) *ast.Node {
	current := node
	for current != nil {
		parent := current.Parent
		if parent == nil {
			return nil
		}
		if test, _ := conditionalTest(parent); test == current {
			return parent
		}
		// ts-go combines ESTree's MethodDefinition and FunctionExpression into
		// one node. Its computed name and member decorators remain outside the
		// function boundary; parameters, types, and the body remain inside it.
		if ast.IsFunctionLikeDeclaration(parent) &&
			current != parent.Name() && current.Kind != ast.KindDecorator {
			return nil
		}
		current = parent
	}
	return nil
}

// parseOptions returns the configured mode, defaulting to "except-parens".
func parseOptions(options []any) string {
	if len(options) == 0 {
		return "except-parens"
	}
	if mode, ok := options[0].(string); ok {
		return mode
	}
	return "except-parens"
}

// NoCondAssignRule disallows assignment operators in conditional test expressions.
var NoCondAssignRule = rule.Rule{
	Name:   "no-cond-assign",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		if parseOptions(options) == "always" {
			return rule.RuleListeners{
				ast.KindBinaryExpression: func(node *ast.Node) {
					if !isAssignmentExpression(node) {
						return
					}
					conditional := findConditionalAncestor(node)
					if conditional != nil {
						ctx.ReportNode(node, unexpectedMessage(conditional))
					}
				},
			}
		}

		checkConditional := func(node *ast.Node) {
			test, requiredParenDepth := conditionalTest(node)
			test, parenDepth := unwrapParentheses(test)
			if !isAssignmentExpression(test) || parenDepth >= requiredParenDepth {
				return
			}
			ctx.ReportNode(test, missingMessage)
		}

		return rule.RuleListeners{
			ast.KindIfStatement:           checkConditional,
			ast.KindWhileStatement:        checkConditional,
			ast.KindDoStatement:           checkConditional,
			ast.KindForStatement:          checkConditional,
			ast.KindConditionalExpression: checkConditional,
		}
	},
}
