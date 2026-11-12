package no_await_in_loop

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func buildUnexpectedAwaitMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unexpectedAwait",
		Description: "Unexpected `await` inside a loop.",
	}
}

// isBoundary distinguishes function execution from computed method names, which
// are evaluated in the enclosing context. Async iteration is intentionally serial.
func isBoundary(node, child *ast.Node) bool {
	switch node.Kind {
	case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction, ast.KindConstructor:
		return true
	case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
		return child != node.Name()
	case ast.KindForOfStatement:
		return node.AsForInOrOfStatement().AwaitModifier != nil
	default:
		return false
	}
}

// isLooped checks the direct child visited while walking ancestors. Initializers
// and iterables run once for this loop, but may still repeat in an outer loop.
func isLooped(node, parent *ast.Node) bool {
	switch parent.Kind {
	case ast.KindForStatement:
		stmt := parent.AsForStatement()
		return node == stmt.Condition || node == stmt.Incrementor || node == stmt.Statement
	case ast.KindForInStatement, ast.KindForOfStatement:
		stmt := parent.AsForInOrOfStatement()
		return node == stmt.Statement || (node == stmt.Initializer && ast.IsVarAwaitUsing(node))
	case ast.KindWhileStatement:
		stmt := parent.AsWhileStatement()
		return node == stmt.Expression || node == stmt.Statement
	case ast.KindDoStatement:
		stmt := parent.AsDoStatement()
		return node == stmt.Expression || node == stmt.Statement
	default:
		return false
	}
}

// NoAwaitInLoopRule disallows await inside of loops, matching ESLint v10.8.1.
var NoAwaitInLoopRule = rule.Rule{
	Name:   "no-await-in-loop",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		validate := func(awaitNode *ast.Node) {
			switch awaitNode.Kind {
			case ast.KindForOfStatement:
				if awaitNode.AsForInOrOfStatement().AwaitModifier == nil {
					return
				}
			case ast.KindVariableDeclarationList:
				if !ast.IsVarAwaitUsing(awaitNode) {
					return
				}
			}

			for node, parent := awaitNode, awaitNode.Parent; parent != nil; node, parent = parent, parent.Parent {
				if isBoundary(parent, node) {
					return
				}
				if isLooped(node, parent) {
					// ESTree VariableDeclaration includes a statement's semicolon;
					// ts-go stores it on the surrounding VariableStatement instead.
					reportNode := awaitNode
					if awaitNode.Kind == ast.KindVariableDeclarationList && awaitNode.Parent.Kind == ast.KindVariableStatement {
						reportNode = awaitNode.Parent
					}
					ctx.ReportNode(reportNode, buildUnexpectedAwaitMessage())
					return
				}
			}
		}
		return rule.RuleListeners{
			ast.KindAwaitExpression:         validate,
			ast.KindForOfStatement:          validate,
			ast.KindVariableDeclarationList: validate,
		}
	},
}
