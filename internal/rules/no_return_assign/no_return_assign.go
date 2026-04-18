package no_return_assign

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// isSentinel mirrors ESLint's SENTINEL_TYPE regex
// (/^(?:[a-zA-Z]+?Statement|ArrowFunctionExpression|FunctionExpression|ClassExpression)$/):
// any statement, plus function / class expressions that introduce a new scope
// boundary where the walk-up for the enclosing return/arrow must stop.
func isSentinel(node *ast.Node) bool {
	if node == nil {
		return false
	}
	if ast.IsStatement(node) {
		return true
	}
	return ast.IsArrowFunction(node) || ast.IsFunctionExpression(node) || ast.IsClassExpression(node)
}

// parseMode extracts the rule mode: "except-parens" (default) or "always".
// Options come in as a raw string (Go tests) or a []interface{} holding the
// string at index 0 (JS tests, ESLint option-array convention).
func parseMode(options any) string {
	switch v := options.(type) {
	case string:
		if v != "" {
			return v
		}
	case []interface{}:
		if len(v) > 0 {
			if s, ok := v[0].(string); ok && s != "" {
				return s
			}
		}
	}
	return "except-parens"
}

// https://eslint.org/docs/latest/rules/no-return-assign
var NoReturnAssignRule = rule.Rule{
	Name: "no-return-assign",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		always := parseMode(options) == "always"

		return rule.RuleListeners{
			ast.KindBinaryExpression: func(node *ast.Node) {
				if !ast.IsAssignmentExpression(node, false /*excludeCompoundAssignment*/) {
					return
				}
				// except-parens: a directly parenthesised assignment
				// (`return (a = b)`, `() => (a = b)`) is allowed.
				// tsgo materializes parens as a ParenthesizedExpression node,
				// which replaces ESLint's token-level isParenthesised check.
				if !always && node.Parent != nil && ast.IsParenthesizedExpression(node.Parent) {
					return
				}

				currentChild := node
				parent := node.Parent
				for parent != nil && !isSentinel(parent) {
					currentChild = parent
					parent = parent.Parent
				}
				if parent == nil {
					return
				}
				switch {
				case ast.IsReturnStatement(parent):
					ctx.ReportNode(parent, rule.RuleMessage{
						Id:          "returnAssignment",
						Description: "Return statement should not contain assignment.",
					})
				case ast.IsArrowFunction(parent):
					if arrow := parent.AsArrowFunction(); arrow != nil && arrow.Body == currentChild {
						ctx.ReportNode(parent, rule.RuleMessage{
							Id:          "arrowAssignment",
							Description: "Arrow function should not return assignment.",
						})
					}
				}
			},
		}
	},
}
