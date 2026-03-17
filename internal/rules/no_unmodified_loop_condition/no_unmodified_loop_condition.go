package no_unmodified_loop_condition

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// buildLoopConditionNotModifiedMessage creates the diagnostic message
func buildLoopConditionNotModifiedMessage(name string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "loopConditionNotModified",
		Description: fmt.Sprintf("'%s' is not modified in this loop.", name),
	}
}

// hasDynamicExpression checks if an expression contains any dynamic sub-expression
// (call, member access, new, tagged template) that could have side effects.
// Returns true if a dynamic expression is found, meaning we should skip checking.
func hasDynamicExpression(node *ast.Node) bool {
	if node == nil {
		return false
	}

	switch node.Kind {
	case ast.KindCallExpression,
		ast.KindPropertyAccessExpression,
		ast.KindElementAccessExpression,
		ast.KindNewExpression,
		ast.KindTaggedTemplateExpression:
		return true
	}

	found := false
	node.ForEachChild(func(child *ast.Node) bool {
		if hasDynamicExpression(child) {
			found = true
			return true // stop iteration
		}
		return false
	})
	return found
}

// collectConditionIdentifiers recursively walks the condition expression and
// collects all identifier names. Returns nil if any dynamic expression is found.
func collectConditionIdentifiers(node *ast.Node) []string {
	if node == nil {
		return nil
	}

	// If the condition contains a dynamic expression, skip the whole condition
	if hasDynamicExpression(node) {
		return nil
	}

	var names []string
	collectIdentifiers(node, &names)
	return names
}

// collectIdentifiers recursively collects identifier names from an expression
func collectIdentifiers(node *ast.Node, names *[]string) {
	if node == nil {
		return
	}

	if node.Kind == ast.KindIdentifier {
		text := node.Text()
		if text != "" {
			// Avoid duplicates
			found := false
			for _, n := range *names {
				if n == text {
					found = true
					break
				}
			}
			if !found {
				*names = append(*names, text)
			}
		}
		return
	}

	node.ForEachChild(func(child *ast.Node) bool {
		collectIdentifiers(child, names)
		return false
	})
}

// isAssignmentOperator checks if a token kind is an assignment operator
func isAssignmentOperator(kind ast.Kind) bool {
	switch kind {
	case ast.KindEqualsToken,
		ast.KindPlusEqualsToken,
		ast.KindMinusEqualsToken,
		ast.KindAsteriskEqualsToken,
		ast.KindAsteriskAsteriskEqualsToken,
		ast.KindSlashEqualsToken,
		ast.KindPercentEqualsToken,
		ast.KindLessThanLessThanEqualsToken,
		ast.KindGreaterThanGreaterThanEqualsToken,
		ast.KindGreaterThanGreaterThanGreaterThanEqualsToken,
		ast.KindAmpersandEqualsToken,
		ast.KindBarEqualsToken,
		ast.KindCaretEqualsToken,
		ast.KindBarBarEqualsToken,
		ast.KindAmpersandAmpersandEqualsToken,
		ast.KindQuestionQuestionEqualsToken:
		return true
	}
	return false
}

// isFunctionBoundary returns true for nodes that represent function boundaries.
// We should not look inside functions for modifications because inner functions
// might not execute during the loop iteration.
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

// isIdentifierModified checks if a specific identifier name is the target of
// an assignment or increment/decrement in the given node.
func isIdentifierModified(node *ast.Node, name string) bool {
	if node == nil {
		return false
	}

	switch node.Kind {
	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		if binary != nil && binary.OperatorToken != nil && isAssignmentOperator(binary.OperatorToken.Kind) {
			// Check if the left side is the identifier we're looking for
			if isIdentifierWithName(binary.Left, name) {
				return true
			}
		}

	case ast.KindPrefixUnaryExpression:
		prefix := node.AsPrefixUnaryExpression()
		if prefix != nil {
			if prefix.Operator == ast.KindPlusPlusToken || prefix.Operator == ast.KindMinusMinusToken {
				if isIdentifierWithName(prefix.Operand, name) {
					return true
				}
			}
		}

	case ast.KindPostfixUnaryExpression:
		postfix := node.AsPostfixUnaryExpression()
		if postfix != nil {
			if postfix.Operator == ast.KindPlusPlusToken || postfix.Operator == ast.KindMinusMinusToken {
				if isIdentifierWithName(postfix.Operand, name) {
					return true
				}
			}
		}
	}

	return false
}

// isIdentifierWithName checks if a node is an Identifier with the given name
func isIdentifierWithName(node *ast.Node, name string) bool {
	if node == nil {
		return false
	}
	return node.Kind == ast.KindIdentifier && node.Text() == name
}

// isModifiedInBody walks the body (and optionally the incrementor for ForStatement)
// looking for any modification to the given identifier name.
func isModifiedInBody(body *ast.Node, name string) bool {
	if body == nil {
		return false
	}

	if isIdentifierModified(body, name) {
		return true
	}

	// Don't recurse into function boundaries
	if isFunctionBoundary(body.Kind) {
		return false
	}

	found := false
	body.ForEachChild(func(child *ast.Node) bool {
		if isModifiedInBody(child, name) {
			found = true
			return true // stop iteration
		}
		return false
	})
	return found
}

// checkLoopCondition checks identifiers in a loop condition and reports those
// that are not modified in the loop body (or incrementor for for-statements).
func checkLoopCondition(ctx rule.RuleContext, condition *ast.Node, body *ast.Node, incrementor *ast.Node) {
	if condition == nil || body == nil {
		return
	}

	names := collectConditionIdentifiers(condition)
	if names == nil {
		return // dynamic expression found, skip
	}

	for _, name := range names {
		modified := isModifiedInBody(body, name)
		if !modified && incrementor != nil {
			modified = isModifiedInBody(incrementor, name)
		}
		if !modified {
			ctx.ReportNode(condition, buildLoopConditionNotModifiedMessage(name))
		}
	}
}

// NoUnmodifiedLoopConditionRule disallows variables in loop conditions that are not modified in the loop
var NoUnmodifiedLoopConditionRule = rule.Rule{
	Name: "no-unmodified-loop-condition",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindWhileStatement: func(node *ast.Node) {
				whileStmt := node.AsWhileStatement()
				if whileStmt == nil {
					return
				}
				checkLoopCondition(ctx, whileStmt.Expression, whileStmt.Statement, nil)
			},
			ast.KindDoStatement: func(node *ast.Node) {
				doStmt := node.AsDoStatement()
				if doStmt == nil {
					return
				}
				checkLoopCondition(ctx, doStmt.Expression, doStmt.Statement, nil)
			},
			ast.KindForStatement: func(node *ast.Node) {
				forStmt := node.AsForStatement()
				if forStmt == nil {
					return
				}
				checkLoopCondition(ctx, forStmt.Condition, forStmt.Statement, forStmt.Incrementor)
			},
		}
	},
}
