package no_duplicate_logical_operands

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// typeScriptWrappers mirrors upstream unwrapTypeScriptExpression (as,
// satisfies, <T>, ! and foo<T>) plus tsgo's explicit parentheses, which ESTree
// does not expose.
const typeScriptWrappers = ast.OEKParentheses | ast.OEKAssertions | ast.OEKExpressionsWithTypeArguments

var (
	duplicateMessage = rule.RuleMessage{
		Id:          "no-duplicate-logical-operands",
		Description: "This operand duplicates the left operand.",
	}
	removeDuplicateMessage = rule.RuleMessage{
		Id:          "no-duplicate-logical-operands/suggestion",
		Description: "Remove the duplicate operand.",
	}
)

// isSimpleComputedProperty matches ESTree Identifier, ThisExpression and
// Literal keys, plus static strings: upstream getStaticStringValue also
// accepts template literals without expressions.
func isSimpleComputedProperty(node *ast.Node) bool {
	switch ast.SkipOuterExpressions(node, typeScriptWrappers).Kind {
	case ast.KindIdentifier, ast.KindThisKeyword,
		ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral,
		ast.KindNumericLiteral, ast.KindBigIntLiteral, ast.KindRegularExpressionLiteral,
		ast.KindTrueKeyword, ast.KindFalseKeyword, ast.KindNullKeyword:
		return true
	}
	return false
}

func isSimpleReference(node *ast.Node) bool {
	node = ast.SkipOuterExpressions(node, typeScriptWrappers)
	switch node.Kind {
	case ast.KindIdentifier, ast.KindSuperKeyword, ast.KindThisKeyword:
		return true
	// Optional chaining is checked per member, as upstream does. tsgo flags
	// every member of an optional chain, which ESTree wraps in a
	// ChainExpression that is never a simple reference either.
	case ast.KindPropertyAccessExpression:
		return !ast.IsOptionalChain(node) && isSimpleReference(node.Expression())
	case ast.KindElementAccessExpression:
		return !ast.IsOptionalChain(node) && isSimpleReference(node.Expression()) &&
			isSimpleComputedProperty(node.AsElementAccessExpression().ArgumentExpression)
	}
	return false
}

// isSafelyAutofixableReference inspects the ESTree operand itself, so an
// authored TypeScript wrapper downgrades the edit to a suggestion.
func isSafelyAutofixableReference(node *ast.Node) bool {
	node = utils.ESTreeRuntimeExpression(node)
	return node.Kind == ast.KindIdentifier || node.Kind == ast.KindThisKeyword
}

// adjacentLeftOperand returns the operand textually next to the right one.
// ESTree flattens a parenthesized same-operator left operand the same way.
func adjacentLeftOperand(expression *ast.BinaryExpression) *ast.Node {
	left := utils.ESTreeRuntimeExpression(expression.Left)
	if left.Kind == ast.KindBinaryExpression &&
		left.AsBinaryExpression().OperatorToken.Kind == expression.OperatorToken.Kind {
		return left.AsBinaryExpression().Right
	}
	return expression.Left
}

// exposesIdentifierToTypeof reports whether fixing node would leave a bare
// identifier as the operand of `typeof`. `typeof (missing || missing)` throws a
// ReferenceError for an undeclared global, while `typeof missing` evaluates to
// "undefined", so that edit changes behavior. Upstream v80.0.0 does not guard
// this; member operands are unaffected because `typeof a.b` still throws.
func exposesIdentifierToTypeof(node *ast.Node, retained *ast.Node) bool {
	if ast.SkipOuterExpressions(retained, typeScriptWrappers).Kind != ast.KindIdentifier {
		return false
	}
	for node.Parent != nil && ast.IsOuterExpression(node.Parent, typeScriptWrappers) {
		node = node.Parent
	}
	return node.Parent != nil && node.Parent.Kind == ast.KindTypeOfExpression
}

// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v80.0.0/rules/no-duplicate-logical-operands.js
var NoDuplicateLogicalOperandsRule = rule.Rule{
	Name:   "unicorn/no-duplicate-logical-operands",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindBinaryExpression: func(node *ast.Node) {
				expression := node.AsBinaryExpression()
				operator := expression.OperatorToken.Kind
				if operator != ast.KindAmpersandAmpersandToken && operator != ast.KindBarBarToken {
					return
				}

				adjacentLeft := adjacentLeftOperand(expression)
				// Identifier reads inside `with` can trigger proxy traps or getters.
				if ast.FindAncestorKind(node, ast.KindWithStatement) != nil ||
					!isSimpleReference(adjacentLeft) ||
					!isSimpleReference(expression.Right) ||
					!unicornutil.IsSameComparisonOperand(adjacentLeft, expression.Right) {
					return
				}

				// The retained left range includes its parentheses, as upstream's
				// getParenthesizedRange does; the expression starts with it.
				nodeRange := utils.TrimNodeTextRange(ctx.SourceFile, node)
				leftRange := utils.TrimNodeTextRange(ctx.SourceFile, expression.Left)
				buildFixes := func() []rule.RuleFix {
					// In a call, tag or delete position the logical expression is a
					// value, while the retained operand would be a reference.
					if unicornutil.IsReferenceContext(node) ||
						exposesIdentifierToTypeof(node, expression.Left) ||
						utils.HasCommentInSpan(ctx.Comments.All(), leftRange.End(), nodeRange.End()) {
						return nil
					}
					replacement := ctx.SourceFile.Text()[leftRange.Pos():leftRange.End()]
					return []rule.RuleFix{rule.RuleFixReplaceRange(nodeRange, replacement)}
				}

				right := utils.ESTreeRuntimeExpression(expression.Right)
				if isSafelyAutofixableReference(adjacentLeft) && isSafelyAutofixableReference(expression.Right) {
					ctx.ReportNodeWithDeferredFixes(right, duplicateMessage, buildFixes)
					return
				}
				ctx.ReportNodeWithDeferredSuggestions(right, duplicateMessage, func() []rule.RuleSuggestion {
					fixes := buildFixes()
					if fixes == nil {
						return nil
					}
					return []rule.RuleSuggestion{{Message: removeDuplicateMessage, FixesArr: fixes}}
				})
			},
		}
	},
}
