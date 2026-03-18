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

// identifierRef holds an identifier's name and its AST node for precise reporting.
type identifierRef struct {
	name string
	node *ast.Node
}

// collectConditionIdentifierRefs collects identifier references with their nodes.
// Returns nil if any dynamic expression is found.
func collectConditionIdentifierRefs(node *ast.Node) []identifierRef {
	if node == nil {
		return nil
	}
	if hasDynamicExpression(node) {
		return nil
	}
	var refs []identifierRef
	collectIdentifierRefs(node, &refs)
	return refs
}

func collectIdentifierRefs(node *ast.Node, refs *[]identifierRef) {
	if node == nil {
		return
	}
	if node.Kind == ast.KindIdentifier {
		text := node.Text()
		if text != "" {
			// Avoid duplicates by name
			for _, r := range *refs {
				if r.name == text {
					return
				}
			}
			*refs = append(*refs, identifierRef{name: text, node: node})
		}
		return
	}
	node.ForEachChild(func(child *ast.Node) bool {
		collectIdentifierRefs(child, refs)
		return false
	})
}

// splitConditionGroups splits a condition by top-level || into groups.
// Within each group, identifiers are treated as a unit: if any one is modified,
// the whole group is considered modified (ESLint's group semantics).
func splitConditionGroups(condition *ast.Node) []*ast.Node {
	if condition == nil {
		return nil
	}
	if condition.Kind == ast.KindParenthesizedExpression {
		pe := condition.AsParenthesizedExpression()
		if pe != nil && pe.Expression != nil {
			return splitConditionGroups(pe.Expression)
		}
	}
	if condition.Kind == ast.KindBinaryExpression {
		bin := condition.AsBinaryExpression()
		if bin != nil && bin.OperatorToken != nil && bin.OperatorToken.Kind == ast.KindBarBarToken {
			left := splitConditionGroups(bin.Left)
			right := splitConditionGroups(bin.Right)
			return append(left, right...)
		}
	}
	return []*ast.Node{condition}
}

// checkLoopCondition checks identifiers in a loop condition and reports those
// that are not modified in the loop body (or incrementor for for-statements).
// Uses ESLint's group semantics: conditions split by || form groups;
// within each group, if any identifier is modified, the whole group is OK.
func checkLoopCondition(ctx rule.RuleContext, condition *ast.Node, body *ast.Node, incrementor *ast.Node) {
	if condition == nil || body == nil {
		return
	}

	groups := splitConditionGroups(condition)
	for _, group := range groups {
		refs := collectConditionIdentifierRefs(group)
		if refs == nil {
			continue // dynamic expression found, skip this group
		}

		// Check if any identifier in this group is modified
		anyModified := false
		for _, ref := range refs {
			if isModifiedInBody(body, ref.name) || (incrementor != nil && isModifiedInBody(incrementor, ref.name)) {
				anyModified = true
				break
			}
		}

		// If no identifier in the group is modified, report each unmodified one
		if !anyModified {
			for _, ref := range refs {
				ctx.ReportNode(ref.node, buildLoopConditionNotModifiedMessage(ref.name))
			}
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
