package no_nested_ternary

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

const (
	messageIDTooDeep     = "no-nested-ternary/too-deep"
	messageIDShouldParen = "no-nested-ternary/should-parenthesized"
)

func messageTooDeep() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          messageIDTooDeep,
		Description: "Do not nest ternary expressions.",
	}
}

func messageShouldParen() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          messageIDShouldParen,
		Description: "Nested ternary expression should be parenthesized.",
	}
}

// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v73.0.0/rules/no-nested-ternary.js
var NoNestedTernaryRule = rule.Rule{
	Name:   "unicorn/no-nested-ternary",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindConditionalExpression: func(node *ast.Node) {
				checkConditionalExpression(ctx, node)
			},
		}
	},
}

func checkConditionalExpression(ctx rule.RuleContext, node *ast.Node) {
	cond := node.AsConditionalExpression()
	if cond == nil {
		return
	}

	// Skip when any direct child is a ConditionalExpression (after paren skip):
	// the outer will report and double-reporting this node would be a duplicate.
	if isConditionalExpression(cond.Condition) ||
		isConditionalExpression(cond.WhenTrue) ||
		isConditionalExpression(cond.WhenFalse) {
		return
	}

	// Count consecutive ConditionalExpression ancestors (skipping paren wrappers,
	// which ESTree flattens out of the AST but tsgo keeps as explicit nodes).
	nestLevel := 0
	for ancestor := effectiveParent(node); ancestor != nil && ancestor.Kind == ast.KindConditionalExpression; ancestor = effectiveParent(ancestor) {
		nestLevel++
	}

	switch {
	case nestLevel == 1 && !isSourceParenthesized(node):
		// One level of unparenthesized nesting → wrap it in parens (autofix).
		ctx.ReportNodeWithDeferredFixes(node, messageShouldParen(), func() []rule.RuleFix {
			return []rule.RuleFix{
				rule.RuleFixInsertBefore(ctx.SourceFile, node, "("),
				rule.RuleFixInsertAfter(node, ")"),
			}
		})
	case nestLevel > 1:
		// Deep nesting: report "too-deep" on the appropriate node. When the
		// nesting chain has more than two ternary levels, point at the upper
		// boundary (ancestors[nestLevel-3] in the upstream array, i.e. the
		// (nestLevel-2)th effective parent); otherwise report on the innermost
		// node itself.
		target := node
		for i := 0; i < nestLevel-2; i++ {
			target = effectiveParent(target)
		}
		ctx.ReportNode(target, messageTooDeep())
	}
}

// isConditionalExpression reports whether the (paren-skipped) node is a ConditionalExpression.
func isConditionalExpression(node *ast.Node) bool {
	if node == nil {
		return false
	}
	return ast.SkipParentheses(node).Kind == ast.KindConditionalExpression
}

// isSourceParenthesized reports whether the node is wrapped in a paren in the source.
// One or more paren wrappers are equivalent for this rule; the test's first non-paren
// ancestor is what matters for the `nestLevel == 1` parenthesized-OK branch.
func isSourceParenthesized(node *ast.Node) bool {
	return node.Parent != nil && node.Parent.Kind == ast.KindParenthesizedExpression
}

// effectiveParent returns node's parent, treating any ParenthesizedExpression
// wrappers as transparent — they don't change the AST-nesting level that the
// upstream rule computes by walking the source-flattened ancestor chain.
func effectiveParent(node *ast.Node) *ast.Node {
	current := node
	for current.Parent != nil && current.Parent.Kind == ast.KindParenthesizedExpression {
		current = current.Parent
	}
	return current.Parent
}
