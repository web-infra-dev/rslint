package no_extra_boolean_cast

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// https://eslint.org/docs/latest/rules/no-extra-boolean-cast

// isInBooleanContext checks whether the given node is in a position that
// already coerces to boolean:
//   - test of if / while / do-while / for
//   - condition of a ternary (ConditionalExpression)
//   - operand of the logical-not operator (!)
func isInBooleanContext(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}

	// Skip over parenthesized expressions to reach the real parent
	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		node = parent
		parent = parent.Parent
	}
	if parent == nil {
		return false
	}

	switch parent.Kind {
	case ast.KindIfStatement:
		return parent.AsIfStatement().Expression == node
	case ast.KindWhileStatement:
		return parent.AsWhileStatement().Expression == node
	case ast.KindDoStatement:
		return parent.AsDoStatement().Expression == node
	case ast.KindForStatement:
		return parent.AsForStatement().Condition == node
	case ast.KindConditionalExpression:
		return parent.AsConditionalExpression().Condition == node
	case ast.KindPrefixUnaryExpression:
		prefix := parent.AsPrefixUnaryExpression()
		return prefix.Operator == ast.KindExclamationToken
	case ast.KindCallExpression:
		// Boolean(!!expr) — argument to Boolean() is a boolean context
		return isBooleanCall(parent)
	case ast.KindNewExpression:
		// new Boolean(!!expr) — argument to new Boolean() is a boolean context
		return isBooleanNewExpr(parent)
	}

	return false
}

// isBooleanCall checks if a CallExpression is a Boolean() call.
func isBooleanCall(node *ast.Node) bool {
	call := node.AsCallExpression()
	if call == nil || call.Expression == nil {
		return false
	}
	callee := call.Expression
	return callee.Kind == ast.KindIdentifier && callee.AsIdentifier().Text == "Boolean"
}

// isBooleanNewExpr checks if a NewExpression is a new Boolean() call.
func isBooleanNewExpr(node *ast.Node) bool {
	newExpr := node.AsNewExpression()
	if newExpr == nil || newExpr.Expression == nil {
		return false
	}
	callee := newExpr.Expression
	return callee.Kind == ast.KindIdentifier && callee.AsIdentifier().Text == "Boolean"
}

// NoExtraBooleanCastRule disallows unnecessary boolean casts.
// Reports !!expr (double negation) and Boolean(expr) calls in contexts
// that already coerce to boolean.
var NoExtraBooleanCastRule = rule.Rule{
	Name: "no-extra-boolean-cast",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			// Detect !!expr (double negation) in boolean context
			ast.KindPrefixUnaryExpression: func(node *ast.Node) {
				prefix := node.AsPrefixUnaryExpression()
				if prefix == nil || prefix.Operator != ast.KindExclamationToken {
					return
				}

				// Check if the operand is also a ! expression => !!expr
				operand := prefix.Operand
				if operand == nil || operand.Kind != ast.KindPrefixUnaryExpression {
					return
				}
				innerPrefix := operand.AsPrefixUnaryExpression()
				if innerPrefix == nil || innerPrefix.Operator != ast.KindExclamationToken {
					return
				}

				// We have !!expr. Check if the outer !! is in a boolean context.
				if isInBooleanContext(node) {
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "unexpectedNegation",
						Description: "Redundant double negation.",
					})
				}
			},

			// Detect new Boolean(expr) in boolean context
			ast.KindNewExpression: func(node *ast.Node) {
				newExpr := node.AsNewExpression()
				if newExpr == nil || newExpr.Expression == nil {
					return
				}
				callee := newExpr.Expression
				if callee.Kind != ast.KindIdentifier || callee.AsIdentifier().Text != "Boolean" {
					return
				}
				if isInBooleanContext(node) {
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "unexpectedCall",
						Description: "Redundant Boolean call.",
					})
				}
			},

			// Detect Boolean(expr) calls in boolean context
			ast.KindCallExpression: func(node *ast.Node) {
				callExpr := node.AsCallExpression()
				if callExpr == nil || callExpr.Expression == nil {
					return
				}

				// Check callee is identifier "Boolean"
				callee := callExpr.Expression
				if callee.Kind != ast.KindIdentifier {
					return
				}
				if callee.AsIdentifier().Text != "Boolean" {
					return
				}

				// Ensure this is not a `new Boolean(...)` expression.
				// In that case, the CallExpression's parent would be a NewExpression.
				// Actually, `new Boolean(...)` is parsed as KindNewExpression, not
				// KindCallExpression, so we don't need an extra check here.

				// Check if the Boolean() call is in a boolean context
				if isInBooleanContext(node) {
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "unexpectedCall",
						Description: "Redundant Boolean call.",
					})
				}
			},
		}
	},
}
