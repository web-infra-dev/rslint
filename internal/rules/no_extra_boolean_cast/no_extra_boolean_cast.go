package no_extra_boolean_cast

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// https://eslint.org/docs/latest/rules/no-extra-boolean-cast

type options struct {
	enforceForLogicalOperands  bool
	enforceForInnerExpressions bool
}

func parseOptions(opts any) options {
	result := options{}
	optsMap := utils.GetOptionsMap(opts)
	if optsMap == nil {
		return result
	}
	if v, ok := optsMap["enforceForLogicalOperands"].(bool); ok {
		result.enforceForLogicalOperands = v
	}
	if v, ok := optsMap["enforceForInnerExpressions"].(bool); ok {
		result.enforceForInnerExpressions = v
	}
	return result
}

// effectiveParent returns the first non-parenthesized ancestor of node,
// along with the direct child of that ancestor (either node itself or the
// outermost ParenthesizedExpression wrapping it).
func effectiveParent(node *ast.Node) (parent, child *ast.Node) {
	child = node
	for child.Parent != nil && ast.IsParenthesizedExpression(child.Parent) {
		child = child.Parent
	}
	return child.Parent, child
}

// isBooleanFunctionOrConstructorCall reports whether node is a
// Boolean(...) call or new Boolean(...) construction.
func isBooleanFunctionOrConstructorCall(node *ast.Node) bool {
	if node == nil {
		return false
	}
	var callee *ast.Node
	switch node.Kind {
	case ast.KindCallExpression:
		callee = node.AsCallExpression().Expression
	case ast.KindNewExpression:
		callee = node.AsNewExpression().Expression
	default:
		return false
	}
	if callee == nil || callee.Kind != ast.KindIdentifier {
		return false
	}
	return callee.AsIdentifier().Text == "Boolean"
}

// firstArgument returns the first argument node of a Boolean(...) /
// new Boolean(...) call, or nil when absent.
func firstArgument(node *ast.Node) *ast.Node {
	var args *ast.NodeList
	switch node.Kind {
	case ast.KindCallExpression:
		args = node.AsCallExpression().Arguments
	case ast.KindNewExpression:
		args = node.AsNewExpression().Arguments
	default:
		return nil
	}
	if args == nil || len(args.Nodes) == 0 {
		return nil
	}
	return args.Nodes[0]
}

// isInBooleanContext reports whether node sits in a position that
// already coerces to boolean: the test of if / while / do-while / for,
// the condition of a ternary, the operand of `!`, or the first argument
// to Boolean() / new Boolean().
func isInBooleanContext(node *ast.Node) bool {
	parent, child := effectiveParent(node)
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindIfStatement:
		return parent.AsIfStatement().Expression == child
	case ast.KindWhileStatement:
		return parent.AsWhileStatement().Expression == child
	case ast.KindDoStatement:
		return parent.AsDoStatement().Expression == child
	case ast.KindForStatement:
		return parent.AsForStatement().Condition == child
	case ast.KindConditionalExpression:
		return parent.AsConditionalExpression().Condition == child
	case ast.KindPrefixUnaryExpression:
		return parent.AsPrefixUnaryExpression().Operator == ast.KindExclamationToken
	case ast.KindCallExpression, ast.KindNewExpression:
		if !isBooleanFunctionOrConstructorCall(parent) {
			return false
		}
		return firstArgument(parent) == child
	}
	return false
}

// isInFlaggedContext reports whether node is in a context where a
// redundant boolean cast should be flagged. When the legacy
// `enforceForLogicalOperands` or the current `enforceForInnerExpressions`
// option is set, recurses through logical / nullish / conditional /
// sequence (comma) parents to reach an outer boolean context.
func isInFlaggedContext(node *ast.Node, opts options) bool {
	parent, child := effectiveParent(node)
	if parent == nil {
		return false
	}

	if opts.enforceForLogicalOperands || opts.enforceForInnerExpressions {
		if parent.Kind == ast.KindBinaryExpression {
			bin := parent.AsBinaryExpression()
			if bin.OperatorToken != nil {
				op := bin.OperatorToken.Kind
				if op == ast.KindBarBarToken || op == ast.KindAmpersandAmpersandToken {
					return isInFlaggedContext(parent, opts)
				}
				if opts.enforceForInnerExpressions && op == ast.KindQuestionQuestionToken && bin.Right == child {
					return isInFlaggedContext(parent, opts)
				}
			}
		}
	}

	if opts.enforceForInnerExpressions {
		if parent.Kind == ast.KindConditionalExpression {
			cond := parent.AsConditionalExpression()
			if cond.WhenTrue == child || cond.WhenFalse == child {
				return isInFlaggedContext(parent, opts)
			}
		}
		// TypeScript parses `a, b, c` as a left-associative nested
		// BinaryExpression tree. The "last expression" — the only one
		// whose value propagates — is always the Right operand of the
		// outermost comma binary.
		if parent.Kind == ast.KindBinaryExpression {
			bin := parent.AsBinaryExpression()
			if bin.OperatorToken != nil && bin.OperatorToken.Kind == ast.KindCommaToken && bin.Right == child {
				return isInFlaggedContext(parent, opts)
			}
		}
	}

	return isInBooleanContext(node)
}

// ---------------------------------------------------------------------
// Fixer helpers
// ---------------------------------------------------------------------

// isCommaSequence reports whether node is a `a, b` comma binary.
func isCommaSequence(node *ast.Node) bool {
	if !ast.IsBinaryExpression(node) {
		return false
	}
	op := node.AsBinaryExpression().OperatorToken
	return op != nil && op.Kind == ast.KindCommaToken
}

// isLogicalOrCoalesceOp reports whether op is `||`, `&&`, or `??`.
func isLogicalOrCoalesceOp(op ast.Kind) bool {
	return op == ast.KindBarBarToken || op == ast.KindAmpersandAmpersandToken || op == ast.KindQuestionQuestionToken
}

// isMixedLogicalAndCoalesce returns true when a replacement node and its
// new binary parent combine `??` with `||`/`&&` — a combination that is
// a syntax error without grouping parens.
func isMixedLogicalAndCoalesce(node, parent *ast.Node) bool {
	if !ast.IsBinaryExpression(node) || !ast.IsBinaryExpression(parent) {
		return false
	}
	nOp := node.AsBinaryExpression().OperatorToken
	pOp := parent.AsBinaryExpression().OperatorToken
	if nOp == nil || pOp == nil {
		return false
	}
	nodeCoalesce := nOp.Kind == ast.KindQuestionQuestionToken
	nodeLogical := nOp.Kind == ast.KindBarBarToken || nOp.Kind == ast.KindAmpersandAmpersandToken
	parentCoalesce := pOp.Kind == ast.KindQuestionQuestionToken
	parentLogical := pOp.Kind == ast.KindBarBarToken || pOp.Kind == ast.KindAmpersandAmpersandToken
	return (nodeCoalesce && parentLogical) || (nodeLogical && parentCoalesce)
}

// needsParens reports whether inserting `replacement`'s text in place of
// `previousNode` would require wrapping it in grouping parentheses to
// preserve semantics. Mirrors ESLint's `needsParens` helper.
func needsParens(previousNode, replacement *ast.Node) bool {
	// If previousNode is already wrapped in parens, those parens stay
	// around the replacement text, so no additional parens are needed.
	if previousNode.Parent != nil && ast.IsParenthesizedExpression(previousNode.Parent) {
		return false
	}
	parent := previousNode.Parent
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindCallExpression, ast.KindNewExpression:
		// A bare comma expression as an argument would be read as
		// separate positional arguments — wrap it.
		return isCommaSequence(replacement)
	case ast.KindIfStatement, ast.KindDoStatement, ast.KindWhileStatement, ast.KindForStatement:
		return false
	case ast.KindConditionalExpression:
		cond := parent.AsConditionalExpression()
		if cond.Condition == previousNode {
			return ast.GetExpressionPrecedence(replacement) <= ast.GetExpressionPrecedence(parent)
		}
		if cond.WhenTrue == previousNode || cond.WhenFalse == previousNode {
			// ESLint compares against AssignmentExpression precedence:
			// a comma expression is the only thing that needs parens
			// here.
			return ast.GetExpressionPrecedence(replacement) < ast.OperatorPrecedenceAssignment
		}
		return false
	case ast.KindPrefixUnaryExpression:
		return ast.GetExpressionPrecedence(replacement) < ast.GetExpressionPrecedence(parent)
	case ast.KindBinaryExpression:
		bin := parent.AsBinaryExpression()
		if bin.OperatorToken == nil {
			return false
		}
		op := bin.OperatorToken.Kind
		if op == ast.KindCommaToken {
			// ESLint's SequenceExpression branch — no extra parens.
			return false
		}
		if isLogicalOrCoalesceOp(op) && isMixedLogicalAndCoalesce(replacement, parent) {
			return true
		}
		if bin.Left == previousNode {
			return ast.GetExpressionPrecedence(replacement) < ast.GetExpressionPrecedence(parent)
		}
		return ast.GetExpressionPrecedence(replacement) <= ast.GetExpressionPrecedence(parent)
	}
	return false
}

// maybePrefixSpace returns replacement with a single leading space
// prepended when inserting it at trimStart would merge with the
// preceding token into a single identifier/keyword.
func maybePrefixSpace(sourceFile *ast.SourceFile, replaceRange core.TextRange, replacement string) string {
	if utils.NeedsLeadingSpaceForReplacement(sourceFile.Text(), replaceRange.Pos(), replacement) {
		return " " + replacement
	}
	return replacement
}

// hasCommentsInSpan reports whether any `//` or `/*` sequence appears in
// the half-open source range [start, end). `utils.HasCommentsInRange`
// only surfaces comments anchored at the range start, so a raw text
// scan is required to catch comments sprinkled *between* children
// (e.g. `!!/* keep */foo`).
func hasCommentsInSpan(src string, start, end int) bool {
	if start < 0 {
		start = 0
	}
	if end > len(src) {
		end = len(src)
	}
	for i := start; i+1 < end; i++ {
		if src[i] == '/' && (src[i+1] == '/' || src[i+1] == '*') {
			return true
		}
	}
	return false
}

// buildNegationFix returns the fix for a redundant `!!expr`. outerUnary
// is the outer `!` of `!!expr`.
func buildNegationFix(ctx rule.RuleContext, outerUnary *ast.Node) []rule.RuleFix {
	innerUnary := outerUnary.AsPrefixUnaryExpression().Operand
	if innerUnary == nil {
		return nil
	}
	argument := innerUnary.AsPrefixUnaryExpression().Operand
	if argument == nil {
		return nil
	}

	replaceRange := utils.TrimNodeTextRange(ctx.SourceFile, outerUnary)
	if hasCommentsInSpan(ctx.SourceFile.Text(), replaceRange.Pos(), replaceRange.End()) {
		return nil
	}

	// ESLint's AST has no parenthesized-expression nodes — the argument
	// of `!!` is whatever the parens wrap. Match that by peeling any
	// ParenthesizedExpression wrappers before computing the replacement
	// text so `!!(a19)` → `a19` (not `(a19)`).
	inner := ast.SkipParentheses(argument)
	argText := utils.TrimmedNodeText(ctx.SourceFile, inner)

	if needsParens(outerUnary, inner) {
		return []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, outerUnary, "("+argText+")")}
	}

	return []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, outerUnary, maybePrefixSpace(ctx.SourceFile, replaceRange, argText))}
}

// buildCallFix returns the fix for a redundant `Boolean(...)` call.
func buildCallFix(ctx rule.RuleContext, callNode *ast.Node) []rule.RuleFix {
	call := callNode.AsCallExpression()
	args := call.Arguments

	argCount := 0
	if args != nil {
		argCount = len(args.Nodes)
	}

	callRange := utils.TrimNodeTextRange(ctx.SourceFile, callNode)

	if argCount == 0 {
		// `Boolean()` alone is `false`. `!Boolean()` collapses to `true`
		// — we replace the parent unary in that case so the `!` goes
		// away together.
		parent, _ := effectiveParent(callNode)
		if parent != nil && parent.Kind == ast.KindPrefixUnaryExpression &&
			parent.AsPrefixUnaryExpression().Operator == ast.KindExclamationToken {
			parentRange := utils.TrimNodeTextRange(ctx.SourceFile, parent)
			if hasCommentsInSpan(ctx.SourceFile.Text(), parentRange.Pos(), parentRange.End()) {
				return nil
			}
			return []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, parent, maybePrefixSpace(ctx.SourceFile, parentRange, "true"))}
		}

		if hasCommentsInSpan(ctx.SourceFile.Text(), callRange.Pos(), callRange.End()) {
			return nil
		}
		return []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, callNode, maybePrefixSpace(ctx.SourceFile, callRange, "false"))}
	}

	if argCount == 1 {
		arg := args.Nodes[0]
		if ast.IsSpreadElement(arg) {
			return nil
		}
		if hasCommentsInSpan(ctx.SourceFile.Text(), callRange.Pos(), callRange.End()) {
			return nil
		}
		// Peel parens to match ESLint's paren-less AST; see buildNegationFix.
		inner := ast.SkipParentheses(arg)
		argText := utils.TrimmedNodeText(ctx.SourceFile, inner)
		if needsParens(callNode, inner) {
			return []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, callNode, "("+argText+")")}
		}
		return []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, callNode, maybePrefixSpace(ctx.SourceFile, callRange, argText))}
	}

	// Two or more arguments: unsafe to drop, skip the fix.
	return nil
}

// NoExtraBooleanCastRule disallows unnecessary boolean casts.
// Reports `!!expr` (double negation) and `Boolean(expr)` calls in
// contexts that already coerce to boolean.
var NoExtraBooleanCastRule = rule.Rule{
	Name: "no-extra-boolean-cast",
	Run: func(ctx rule.RuleContext, ruleOptions any) rule.RuleListeners {
		opts := parseOptions(ruleOptions)

		negationMsg := rule.RuleMessage{
			Id:          "unexpectedNegation",
			Description: "Redundant double negation.",
		}
		callMsg := rule.RuleMessage{
			Id:          "unexpectedCall",
			Description: "Redundant Boolean call.",
		}

		return rule.RuleListeners{
			// Detect !!expr (double negation) in a flagged context.
			ast.KindPrefixUnaryExpression: func(node *ast.Node) {
				prefix := node.AsPrefixUnaryExpression()
				if prefix == nil || prefix.Operator != ast.KindExclamationToken {
					return
				}
				operand := prefix.Operand
				if operand == nil || !ast.IsPrefixUnaryExpression(operand) {
					return
				}
				if operand.AsPrefixUnaryExpression().Operator != ast.KindExclamationToken {
					return
				}
				if !isInFlaggedContext(node, opts) {
					return
				}
				if fixes := buildNegationFix(ctx, node); fixes != nil {
					ctx.ReportNodeWithFixes(node, negationMsg, fixes...)
				} else {
					ctx.ReportNode(node, negationMsg)
				}
			},

			// Detect Boolean(expr) in a flagged context.
			//
			// `new Boolean(...)` is intentionally NOT listened to here —
			// matching ESLint. `new Boolean(x)` creates a Boolean object,
			// which is always truthy in a boolean context, so it is not
			// equivalent to a plain `x` and cannot be called "redundant".
			// `isBooleanFunctionOrConstructorCall` is still used when
			// determining whether an inner expression is in a boolean
			// context (argument to `new Boolean(...)`).
			ast.KindCallExpression: func(node *ast.Node) {
				if !isBooleanFunctionOrConstructorCall(node) {
					return
				}
				if !isInFlaggedContext(node, opts) {
					return
				}
				if fixes := buildCallFix(ctx, node); fixes != nil {
					ctx.ReportNodeWithFixes(node, callMsg, fixes...)
				} else {
					ctx.ReportNode(node, callMsg)
				}
			},
		}
	},
}
