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

// isReturnThrowSentinel checks if a node is a sentinel for return/throw statements
// and labeled break statements. These are only stopped by function and class boundaries.
func isReturnThrowSentinel(node *ast.Node) bool {
	return ast.IsFunctionLikeDeclaration(node) || ast.IsClassLike(node)
}

// isBreakSentinel checks if a node is a sentinel for unlabeled break statements.
// Unlabeled break is stopped by function boundaries, class boundaries, loops, and switch.
func isBreakSentinel(node *ast.Node) bool {
	return ast.IsFunctionLikeDeclaration(node) || ast.IsClassLike(node) ||
		ast.IsIterationStatement(node, false) || ast.IsSwitchStatement(node)
}

// isContinueSentinel checks if a node is a sentinel for continue statements (labeled or not).
// Continue is stopped by function boundaries, class boundaries, and loops (NOT switch).
func isContinueSentinel(node *ast.Node) bool {
	return ast.IsFunctionLikeDeclaration(node) || ast.IsClassLike(node) ||
		ast.IsIterationStatement(node, false)
}

// isInFinally walks up the parent chain and checks if the node is inside a finally block
// of a try statement. If a sentinel node is encountered first, the node is considered safe.
// For labeled break/continue, the label target must be outside the finally block for it to be unsafe.
func isInFinally(node *ast.Node, isSentinel func(*ast.Node) bool, label *ast.Node) bool {
	return ast.FindAncestorOrQuit(node.Parent, func(current *ast.Node) ast.FindAncestorResult {
		if isSentinel(current) {
			return ast.FindAncestorQuit
		}

		// For labeled break/continue: if we find the matching label inside the
		// finally block, the break/continue is safe (targets something inside finally)
		if label != nil && ast.IsLabeledStatement(current) {
			labeledStmt := current.AsLabeledStatement()
			if labeledStmt.Label != nil && labeledStmt.Label.Text() == label.Text() {
				return ast.FindAncestorQuit
			}
		}

		// Check if we're in a finally block of a try statement
		if current.Parent != nil && ast.IsTryStatement(current.Parent) {
			tryStmt := current.Parent.AsTryStatement()
			if tryStmt.FinallyBlock != nil && tryStmt.FinallyBlock == current {
				return ast.FindAncestorTrue
			}
		}

		return ast.FindAncestorFalse
	}) != nil
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
					if isInFinally(node, isReturnThrowSentinel, breakStmt.Label) {
						ctx.ReportNode(node, buildUnsafeUsageMessage("BreakStatement"))
					}
				} else {
					if isInFinally(node, isBreakSentinel, nil) {
						ctx.ReportNode(node, buildUnsafeUsageMessage("BreakStatement"))
					}
				}
			},
			ast.KindContinueStatement: func(node *ast.Node) {
				continueStmt := node.AsContinueStatement()
				if continueStmt.Label != nil {
					if isInFinally(node, isContinueSentinel, continueStmt.Label) {
						ctx.ReportNode(node, buildUnsafeUsageMessage("ContinueStatement"))
					}
				} else {
					if isInFinally(node, isContinueSentinel, nil) {
						ctx.ReportNode(node, buildUnsafeUsageMessage("ContinueStatement"))
					}
				}
			},
		}
	},
}
