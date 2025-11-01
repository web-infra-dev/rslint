package no_constructor_return

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// Message builder
func buildUnexpectedMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unexpected",
		Description: "Unexpected return statement in constructor.",
	}
}

// findEnclosingConstructor walks up the tree to find if we're inside a constructor
func findEnclosingConstructor(node *ast.Node) *ast.Node {
	if node == nil {
		return nil
	}

	current := node.Parent
	for current != nil {
		// Check if this is a constructor method
		if current.Kind == ast.KindConstructor {
			return current
		}

		// Stop at function boundaries - don't traverse into nested functions
		// Return statements in nested functions don't apply to the outer constructor
		switch current.Kind {
		case ast.KindFunctionDeclaration,
			ast.KindFunctionExpression,
			ast.KindArrowFunction,
			ast.KindMethodDeclaration:
			// We've hit a function boundary, stop searching
			return nil
		}

		current = current.Parent
	}

	return nil
}

// isReturnStatementInConstructor checks if a return statement with a value is inside a constructor
func isReturnStatementInConstructor(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindReturnStatement {
		return false
	}

	returnStmt := node.AsReturnStatement()
	if returnStmt == nil {
		return false
	}

	// Only flag return statements that have an expression (return with a value)
	// Bare return statements (return;) are allowed for flow control
	if returnStmt.Expression == nil {
		return false
	}

	// Check if this return statement is inside a constructor
	constructor := findEnclosingConstructor(node)
	return constructor != nil
}

// NoConstructorReturnRule disallows returning values in constructors
var NoConstructorReturnRule = rule.CreateRule(rule.Rule{
	Name: "no-constructor-return",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			// Check return statements
			ast.KindReturnStatement: func(node *ast.Node) {
				if isReturnStatementInConstructor(node) {
					ctx.ReportNode(node, buildUnexpectedMessage())
				}
			},
		}
	},
})
