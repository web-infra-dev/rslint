package no_await_in_loop

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func buildUnexpectedAwaitMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unexpectedAwait",
		Description: "Unexpected `await` inside a loop.",
	}
}

// isLoopNode checks if a node is a loop statement
func isLoopNode(node *ast.Node) bool {
	if node == nil {
		return false
	}
	kind := node.Kind
	return kind == ast.KindWhileStatement ||
		kind == ast.KindDoStatement ||
		kind == ast.KindForStatement ||
		kind == ast.KindForInStatement ||
		kind == ast.KindForOfStatement
}

// isForAwaitOfNode checks if a node is a for-await-of loop
func isForAwaitOfNode(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindForOfStatement {
		return false
	}
	// Check for await modifier on ForOfStatement
	stmt := node.AsForInOrOfStatement()
	if stmt == nil {
		return false
	}
	return stmt.AwaitModifier != nil
}

// isInLoop checks if we're currently inside a loop (excluding for-await-of and loop initializers)
func isInLoop(node *ast.Node) bool {
	current := node.Parent
	for current != nil {
		// If we hit a for-await-of, it's allowed
		if isForAwaitOfNode(current) {
			return false
		}

		// Check if we're in a loop initializer position (which is allowed)
		if isLoopNode(current) {
			return !isInLoopInitializer(node, current)
		}

		// Stop if we hit a function boundary (functions create new async contexts)
		kind := current.Kind
		if kind == ast.KindFunctionDeclaration ||
			kind == ast.KindFunctionExpression ||
			kind == ast.KindArrowFunction ||
			kind == ast.KindMethodDeclaration ||
			kind == ast.KindConstructor ||
			kind == ast.KindGetAccessor ||
			kind == ast.KindSetAccessor {
			return false
		}

		current = current.Parent
	}
	return false
}

// isInLoopInitializer checks if a node is in a loop initializer position
func isInLoopInitializer(node *ast.Node, loop *ast.Node) bool {
	if loop == nil || node == nil {
		return false
	}

	switch loop.Kind {
	case ast.KindForInStatement, ast.KindForOfStatement:
		// For for-in/for-of, the initializer is the expression being iterated over
		stmt := loop.AsForInOrOfStatement()
		if stmt == nil {
			return false
		}
		// Check if the node is in the iterable expression
		expr := stmt.Expression
		return isNodeInSubtree(node, expr)

	case ast.KindForStatement:
		// For regular for loops, the initializer is the first part
		stmt := loop.AsForStatement()
		if stmt == nil {
			return false
		}
		// Check if the node is in the initializer
		init := stmt.Initializer
		return isNodeInSubtree(node, init)

	default:
		return false
	}
}

// isNodeInSubtree checks if a node is within a subtree rooted at root
func isNodeInSubtree(node *ast.Node, root *ast.Node) bool {
	if root == nil || node == nil {
		return false
	}
	current := node
	for current != nil {
		if current == root {
			return true
		}
		// Stop at certain boundaries
		if isLoopNode(current) || isFunctionNode(current) {
			return false
		}
		current = current.Parent
	}
	return false
}

// isFunctionNode checks if a node is a function
func isFunctionNode(node *ast.Node) bool {
	if node == nil {
		return false
	}
	kind := node.Kind
	return kind == ast.KindFunctionDeclaration ||
		kind == ast.KindFunctionExpression ||
		kind == ast.KindArrowFunction ||
		kind == ast.KindMethodDeclaration ||
		kind == ast.KindConstructor ||
		kind == ast.KindGetAccessor ||
		kind == ast.KindSetAccessor
}

// NoAwaitInLoopRule disallows await inside of loops
var NoAwaitInLoopRule = rule.CreateRule(rule.Rule{
	Name: "no-await-in-loop",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindAwaitExpression: func(node *ast.Node) {
				if node == nil {
					return
				}

				// Check if this await is inside a loop
				if isInLoop(node) {
					ctx.ReportNode(node, buildUnexpectedAwaitMessage())
				}
			},

			// Handle for-await-of in loops
			ast.KindForOfStatement: func(node *ast.Node) {
				if node == nil {
					return
				}

				// Check if this is a for-await-of
				if !isForAwaitOfNode(node) {
					return
				}

				// Check if the for-await-of is nested inside another loop
				if isInLoop(node) {
					// Report the await modifier
					stmt := node.AsForInOrOfStatement()
					if stmt != nil && stmt.AwaitModifier != nil {
						ctx.ReportNode(stmt.AwaitModifier, buildUnexpectedAwaitMessage())
					}
				}
			},
		}
	},
})
