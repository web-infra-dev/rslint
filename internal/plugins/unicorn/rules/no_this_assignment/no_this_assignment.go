package no_this_assignment

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const messageID = "no-this-assignment"

func assignmentMessage(name string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          messageID,
		Description: fmt.Sprintf("Do not assign `this` to `%s`.", name),
		Data:        map[string]string{"name": name},
	}
}

// NoThisAssignmentRule disallows assigning this directly to an identifier.
//
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v73.0.0/rules/no-this-assignment.js
var NoThisAssignmentRule = rule.Rule{
	Name:   "unicorn/no-this-assignment",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		checkAssignment := func(reportNode, variableNode, valueNode *ast.Node) {
			if variableNode == nil || valueNode == nil {
				return
			}
			variableNode = ast.SkipParentheses(variableNode)
			valueNode = ast.SkipParentheses(valueNode)
			if variableNode.Kind != ast.KindIdentifier || valueNode.Kind != ast.KindThisKeyword {
				return
			}

			ctx.ReportNode(reportNode, assignmentMessage(variableNode.AsIdentifier().Text))
		}

		return rule.RuleListeners{
			ast.KindVariableDeclaration: func(node *ast.Node) {
				declaration := node.AsVariableDeclaration()
				checkAssignment(node, declaration.Name(), declaration.Initializer)
			},
			ast.KindBinaryExpression: func(node *ast.Node) {
				binary := node.AsBinaryExpression()
				if binary.OperatorToken == nil || !ast.IsAssignmentOperator(binary.OperatorToken.Kind) {
					return
				}
				if utils.IsDefaultValueInDestructuringAssignment(node) {
					return
				}

				checkAssignment(node, binary.Left, binary.Right)
			},
		}
	},
}
