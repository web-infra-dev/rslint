package no_setter_return

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// buildSetterMessage returns the diagnostic message for a return-with-value in a setter.
func buildSetterMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "setter",
		Description: "Setter cannot return a value.",
	}
}

// findEnclosingSetter walks up the parent chain to find if the node is inside a setter.
// Returns the setter node if found, or nil if a different function boundary is hit first.
func findEnclosingSetter(node *ast.Node) *ast.Node {
	current := node.Parent
	for current != nil {
		switch current.Kind {
		case ast.KindSetAccessor:
			return current
		case ast.KindFunctionDeclaration,
			ast.KindFunctionExpression,
			ast.KindArrowFunction,
			ast.KindMethodDeclaration,
			ast.KindGetAccessor,
			ast.KindConstructor:
			// Hit a different function boundary; stop searching.
			return nil
		}
		current = current.Parent
	}
	return nil
}

// NoSetterReturnRule disallows returning a value from a setter.
// Setters cannot meaningfully return values; any return value is silently ignored.
// A bare `return;` (without a value) is allowed for control flow.
var NoSetterReturnRule = rule.Rule{
	Name: "no-setter-return",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindReturnStatement: func(node *ast.Node) {
				ret := node.AsReturnStatement()
				// Allow bare return (no expression)
				if ret.Expression == nil {
					return
				}
				// Check if the return statement is inside a setter
				if findEnclosingSetter(node) != nil {
					ctx.ReportNode(node, buildSetterMessage())
				}
			},
		}
	},
}
