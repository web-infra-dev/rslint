package no_unused_expressions

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type options struct {
	allowShortCircuit   bool
	allowTernary        bool
	allowTaggedTemplates bool
	enforceForJSX       bool
	ignoreDirectives    bool
}

func parseOptions(rawOptions any) options {
	opts := options{}
	optsMap := utils.GetOptionsMap(rawOptions)
	if optsMap == nil {
		return opts
	}
	if v, ok := optsMap["allowShortCircuit"].(bool); ok {
		opts.allowShortCircuit = v
	}
	if v, ok := optsMap["allowTernary"].(bool); ok {
		opts.allowTernary = v
	}
	if v, ok := optsMap["allowTaggedTemplates"].(bool); ok {
		opts.allowTaggedTemplates = v
	}
	if v, ok := optsMap["enforceForJSX"].(bool); ok {
		opts.enforceForJSX = v
	}
	if v, ok := optsMap["ignoreDirectives"].(bool); ok {
		opts.ignoreDirectives = v
	}
	return opts
}

func unusedExpressionMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unusedExpression",
		Description: "Expected an assignment or function call and instead saw an expression.",
	}
}

// isDisallowed returns true if the expression has no side effects and should be reported.
// This mirrors ESLint's Checker pattern.
func isDisallowed(node *ast.Node, opts *options) bool {
	if node == nil {
		return false
	}

	switch node.Kind {
	// Expressions that never have side effects
	case ast.KindArrayLiteralExpression,
		ast.KindObjectLiteralExpression,
		ast.KindArrowFunction,
		ast.KindFunctionExpression,
		ast.KindClassExpression,
		ast.KindIdentifier,
		ast.KindPropertyAccessExpression,
		ast.KindElementAccessExpression,
		ast.KindMetaProperty,
		ast.KindThisKeyword,
		ast.KindSuperKeyword:
		return true

	// Literals
	case ast.KindStringLiteral,
		ast.KindNumericLiteral,
		ast.KindBigIntLiteral,
		ast.KindRegularExpressionLiteral,
		ast.KindNoSubstitutionTemplateLiteral,
		ast.KindTrueKeyword,
		ast.KindFalseKeyword,
		ast.KindNullKeyword:
		return true

	// Template literals (untagged)
	case ast.KindTemplateExpression:
		return true

	// Sequence expressions (comma-separated)
	case ast.KindCommaListExpression:
		return true

	// Tagged template expressions
	case ast.KindTaggedTemplateExpression:
		return !opts.allowTaggedTemplates

	// JSX
	case ast.KindJsxElement, ast.KindJsxFragment, ast.KindJsxSelfClosingElement:
		return opts.enforceForJSX

	// Unary expressions: void and delete have side effects, others don't
	case ast.KindPrefixUnaryExpression:
		unary := node.AsPrefixUnaryExpression()
		// ++/-- have side effects
		if unary.Operator == ast.KindPlusPlusToken || unary.Operator == ast.KindMinusMinusToken {
			return false
		}
		// !, +, -, ~ do not
		return true

	// typeof has no side effects
	case ast.KindTypeOfExpression:
		return true

	// Binary expressions
	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		op := binary.OperatorToken.Kind
		// Assignment operators have side effects
		if ast.IsAssignmentOperator(op) {
			return false
		}
		// Logical operators: &&, ||, ??
		if op == ast.KindAmpersandAmpersandToken || op == ast.KindBarBarToken || op == ast.KindQuestionQuestionToken {
			if opts.allowShortCircuit {
				return isDisallowed(binary.Right, opts)
			}
			return true
		}
		// All other binary operators (arithmetic, comparison, bitwise) have no side effects
		return true

	// Conditional (ternary) expression
	case ast.KindConditionalExpression:
		if opts.allowTernary {
			cond := node.AsConditionalExpression()
			return isDisallowed(cond.WhenTrue, opts) || isDisallowed(cond.WhenFalse, opts)
		}
		return true

	// Parenthesized expression: unwrap
	case ast.KindParenthesizedExpression:
		return isDisallowed(node.AsParenthesizedExpression().Expression, opts)

	// TypeScript-specific: unwrap type assertions and check inner expression
	case ast.KindAsExpression:
		return isDisallowed(node.AsAsExpression().Expression, opts)
	case ast.KindTypeAssertionExpression:
		return isDisallowed(node.AsTypeAssertion().Expression, opts)
	case ast.KindNonNullExpression:
		return isDisallowed(node.AsNonNullExpression().Expression, opts)
	// NOTE: SatisfiesExpression is NOT unwrapped to match @typescript-eslint behavior.
	// The original rule does not handle TSSatisfiesExpression, so it defaults to
	// "not disallowed" (i.e., `foo satisfies T;` is NOT flagged).

	// Instantiation expression: Foo<string> as a standalone expression
	case ast.KindExpressionWithTypeArguments:
		expr := node.AsExpressionWithTypeArguments()
		return isDisallowed(expr.Expression, opts)

	// Expressions with side effects (NOT disallowed)
	case ast.KindCallExpression,
		ast.KindNewExpression,
		ast.KindDeleteExpression,
		ast.KindVoidExpression,
		ast.KindAwaitExpression,
		ast.KindYieldExpression,
		ast.KindPostfixUnaryExpression:
		return false

	default:
		// Unknown node types default to not disallowed (safe)
		return false
	}
}

// isDirectiveContext returns true if the parent node is a valid context for
// directive prologues: SourceFile, function body (Block), class static block
// body (Block), or ModuleBlock. Arbitrary blocks (e.g., if-blocks) are NOT valid.
// NOTE: Per spec, class static blocks do NOT have directive prologues, but the
// @typescript-eslint parser treats leading string literals in static blocks as
// directives, so we match that behavior for alignment.
func isDirectiveContext(parent *ast.Node) bool {
	switch parent.Kind {
	case ast.KindSourceFile, ast.KindModuleBlock:
		return true
	case ast.KindBlock:
		// Block is a directive context if it's the body of a function-like or class static block
		return parent.Parent != nil && ast.IsFunctionLikeOrClassStaticBlockDeclaration(parent.Parent)
	default:
		return false
	}
}

// isDirective checks if a node is a directive prologue (e.g., 'use strict').
// Uses ast.IsPrologueDirective from tsgo to identify string literal expression
// statements, then validates the node appears in the leading directive sequence
// of a valid context (SourceFile, function body, or module/namespace block).
func isDirective(node *ast.Node) bool {
	if !ast.IsPrologueDirective(node) {
		return false
	}
	parent := node.Parent
	if parent == nil || !isDirectiveContext(parent) {
		return false
	}
	// Check that this node appears in the leading sequence of directive-like statements.
	// node.Parent.Statements() covers SourceFile, Block, and ModuleBlock.
	for _, stmt := range parent.Statements() {
		if stmt == node {
			return true
		}
		if !ast.IsPrologueDirective(stmt) {
			return false
		}
	}
	return false
}

// https://typescript-eslint.io/rules/no-unused-expressions
var NoUnusedExpressionsRule = rule.CreateRule(rule.Rule{
	Name: "no-unused-expressions",
	Run: func(ctx rule.RuleContext, rawOptions any) rule.RuleListeners {
		opts := parseOptions(rawOptions)

		return rule.RuleListeners{
			ast.KindExpressionStatement: func(node *ast.Node) {
				exprStmt := node.AsExpressionStatement()
				expr := exprStmt.Expression

				if !isDisallowed(expr, &opts) {
					return
				}

				// Always allow directives (unless ignoreDirectives is irrelevant here;
				// the base behavior always skips real directives)
				if isDirective(node) {
					return
				}

				ctx.ReportNode(node, unusedExpressionMessage())
			},
		}
	},
})
