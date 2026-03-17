package no_extra_bind

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// https://eslint.org/docs/latest/rules/no-extra-bind
var NoExtraBindRule = rule.Rule{
	Name: "no-extra-bind",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		type scopeInfo struct {
			isBound   bool
			thisFound bool
			upper     *scopeInfo
		}

		var scope *scopeInfo

		// getBindCallNode checks if funcNode is the callee of a .bind() call
		// and returns the CallExpression node if so, otherwise nil.
		// It handles parenthesized expressions like (function(){}).bind(x).
		getBindCallNode := func(funcNode *ast.Node) *ast.Node {
			current := funcNode.Parent
			// Skip parenthesized expressions
			for current != nil && current.Kind == ast.KindParenthesizedExpression {
				current = current.Parent
			}

			if current == nil || current.Kind != ast.KindPropertyAccessExpression {
				return nil
			}

			propAccess := current.AsPropertyAccessExpression()
			if propAccess == nil || propAccess.Name() == nil {
				return nil
			}
			if propAccess.Name().Text() != "bind" {
				return nil
			}

			callParent := current.Parent
			if callParent == nil || callParent.Kind != ast.KindCallExpression {
				return nil
			}

			callExpr := callParent.AsCallExpression()
			if callExpr == nil || callExpr.Expression != current {
				return nil
			}

			// .bind() with 2+ arguments is partial application, which is valid
			argCount := 0
			if callExpr.Arguments != nil {
				argCount = len(callExpr.Arguments.Nodes)
			}
			if argCount >= 2 {
				return nil
			}

			// Check that the single argument (if any) is not a spread element
			if argCount == 1 {
				firstArg := callExpr.Arguments.Nodes[0]
				if firstArg.Kind == ast.KindSpreadElement {
					return nil
				}
			}

			return callParent
		}

		enterFunction := func(node *ast.Node) {
			scope = &scopeInfo{
				isBound:   getBindCallNode(node) != nil,
				thisFound: false,
				upper:     scope,
			}
		}

		exitFunction := func(node *ast.Node) {
			if scope != nil {
				if scope.isBound && !scope.thisFound {
					callNode := getBindCallNode(node)
					if callNode != nil {
						ctx.ReportNode(callNode, rule.RuleMessage{
							Id:          "unexpected",
							Description: "The function binding is unnecessary.",
						})
					}
				}
				scope = scope.upper
			}
		}

		return rule.RuleListeners{
			ast.KindFunctionExpression:                       enterFunction,
			rule.ListenerOnExit(ast.KindFunctionExpression):  exitFunction,
			ast.KindFunctionDeclaration:                      enterFunction,
			rule.ListenerOnExit(ast.KindFunctionDeclaration): exitFunction,

			// Arrow functions with .bind() are always unnecessary
			ast.KindArrowFunction: func(node *ast.Node) {
				callNode := getBindCallNode(node)
				if callNode != nil {
					ctx.ReportNode(callNode, rule.RuleMessage{
						Id:          "unexpected",
						Description: "The function binding is unnecessary.",
					})
				}
			},

			ast.KindThisKeyword: func(node *ast.Node) {
				if scope != nil {
					scope.thisFound = true
				}
			},
		}
	},
}
