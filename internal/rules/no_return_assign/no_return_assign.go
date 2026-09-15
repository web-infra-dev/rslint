package no_return_assign

import (
	_ "embed"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

//go:embed no_return_assign.schema.json
var schemaJSON []byte

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
func parseMode(options []any) string {
	mode := "except-parens"
	if len(options) == 0 {
		return mode
	}
	if s, ok := options[0].(string); ok {
		mode = s
	}
	return mode
}

// https://eslint.org/docs/latest/rules/no-return-assign
var NoReturnAssignRule = rule.Rule{
	Name:   "no-return-assign",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
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
