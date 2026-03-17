package no_unsafe_finally

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// buildUnsafeUsageMessage creates the diagnostic message for unsafe finally usage
func buildUnsafeUsageMessage(nodeType string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unsafeUsage",
		Description: "Unsafe usage of " + nodeType + ".",
	}
}

// isFunctionBoundary returns true for nodes that are function boundaries
// (these stop return/throw from being flagged, and also stop break/continue)
func isFunctionBoundary(kind ast.Kind) bool {
	switch kind {
	case ast.KindFunctionDeclaration,
		ast.KindFunctionExpression,
		ast.KindArrowFunction,
		ast.KindMethodDeclaration,
		ast.KindGetAccessor,
		ast.KindSetAccessor,
		ast.KindConstructor:
		return true
	}
	return false
}

// isClassBoundary returns true for class declaration/expression nodes
func isClassBoundary(kind ast.Kind) bool {
	return kind == ast.KindClassDeclaration || kind == ast.KindClassExpression
}

// isLoopNode returns true for loop statement nodes
func isLoopNode(kind ast.Kind) bool {
	switch kind {
	case ast.KindForStatement,
		ast.KindForInStatement,
		ast.KindForOfStatement,
		ast.KindWhileStatement,
		ast.KindDoStatement:
		return true
	}
	return false
}

// isReturnThrowSentinel checks if a node kind is a sentinel for return/throw statements.
// Return and throw are only stopped by function boundaries and class boundaries.
func isReturnThrowSentinel(kind ast.Kind) bool {
	return isFunctionBoundary(kind) || isClassBoundary(kind)
}

// isBreakSentinel checks if a node kind is a sentinel for unlabeled break statements.
// Unlabeled break is stopped by function boundaries, class boundaries, loops, and switch.
func isBreakSentinel(kind ast.Kind) bool {
	return isFunctionBoundary(kind) || isClassBoundary(kind) || isLoopNode(kind) || kind == ast.KindSwitchStatement
}

// isContinueSentinel checks if a node kind is a sentinel for unlabeled continue statements.
// Unlabeled continue is stopped by function boundaries, class boundaries, and loops (NOT switch).
func isContinueSentinel(kind ast.Kind) bool {
	return isFunctionBoundary(kind) || isClassBoundary(kind) || isLoopNode(kind)
}

// isInFinally walks up the parent chain and checks if the node is inside a finally block
// of a try statement. If a sentinel node is encountered first, the node is considered safe.
// For labeled break/continue, the label target must be outside the finally block for it to be unsafe.
func isInFinally(node *ast.Node, isSentinel func(ast.Kind) bool, label *ast.Node) bool {
	current := node.Parent
	for current != nil {
		// Check if this is a sentinel node that makes the statement safe
		if isSentinel(current.Kind) {
			return false
		}

		// For labeled break/continue: if we find the matching label inside the
		// finally block, the break/continue is safe (targets something inside finally)
		if label != nil && current.Kind == ast.KindLabeledStatement {
			labeledStmt := current.AsLabeledStatement()
			if labeledStmt.Label != nil && labeledStmt.Label.Text() == label.Text() {
				// The label is between the statement and any finally block,
				// so the break/continue targets something inside the finally block
				return false
			}
		}

		// Check if we're in a finally block of a try statement
		if current.Parent != nil && current.Parent.Kind == ast.KindTryStatement {
			tryStmt := current.Parent.AsTryStatement()
			if tryStmt.FinallyBlock != nil && tryStmt.FinallyBlock == current {
				return true
			}
		}

		current = current.Parent
	}
	return false
}

// NoUnsafeFinallyRule disallows control flow statements in finally blocks
var NoUnsafeFinallyRule = rule.Rule{
	Name: "no-unsafe-finally",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindReturnStatement: func(node *ast.Node) {
				if isInFinally(node, isReturnThrowSentinel, nil) {
					ctx.ReportNode(node, buildUnsafeUsageMessage("ReturnStatement"))
				}
			},
			ast.KindThrowStatement: func(node *ast.Node) {
				if isInFinally(node, isReturnThrowSentinel, nil) {
					ctx.ReportNode(node, buildUnsafeUsageMessage("ThrowStatement"))
				}
			},
			ast.KindBreakStatement: func(node *ast.Node) {
				breakStmt := node.AsBreakStatement()
				if breakStmt.Label != nil {
					// Labeled break: only sentinel is function/class boundary
					// The label check is handled inside isInFinally
					if isInFinally(node, isReturnThrowSentinel, breakStmt.Label) {
						ctx.ReportNode(node, buildUnsafeUsageMessage("BreakStatement"))
					}
				} else {
					// Unlabeled break: loops and switch are sentinels
					if isInFinally(node, isBreakSentinel, nil) {
						ctx.ReportNode(node, buildUnsafeUsageMessage("BreakStatement"))
					}
				}
			},
			ast.KindContinueStatement: func(node *ast.Node) {
				continueStmt := node.AsContinueStatement()
				if continueStmt.Label != nil {
					// Labeled continue: only sentinel is function/class boundary
					// The label check is handled inside isInFinally
					if isInFinally(node, isReturnThrowSentinel, continueStmt.Label) {
						ctx.ReportNode(node, buildUnsafeUsageMessage("ContinueStatement"))
					}
				} else {
					// Unlabeled continue: loops are sentinels (NOT switch)
					if isInFinally(node, isContinueSentinel, nil) {
						ctx.ReportNode(node, buildUnsafeUsageMessage("ContinueStatement"))
					}
				}
			},
		}
	},
}
