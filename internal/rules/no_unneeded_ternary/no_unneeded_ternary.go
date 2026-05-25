package no_unneeded_ternary

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// https://eslint.org/docs/latest/rules/no-unneeded-ternary

type options struct {
	defaultAssignment bool
}

func parseOptions(opts any) options {
	result := options{defaultAssignment: true}
	optsMap := utils.GetOptionsMap(opts)
	if optsMap == nil {
		return result
	}
	if v, ok := optsMap["defaultAssignment"].(bool); ok {
		result.defaultAssignment = v
	}
	return result
}

// boolKind reports whether node (after peeling parens) is a `true`/`false`
// keyword and returns its boolean value.
func boolKind(node *ast.Node) (bool, bool) {
	inner := ast.SkipParentheses(node)
	switch inner.Kind {
	case ast.KindTrueKeyword:
		return true, true
	case ast.KindFalseKeyword:
		return false, true
	}
	return false, false
}

// isBooleanExpression mirrors ESLint's helper: a node is a guaranteed boolean
// when it is a comparison BinaryExpression or `!expr`. tsgo uses
// PrefixUnaryExpression for `!`.
func isBooleanExpression(node *ast.Node) bool {
	inner := ast.SkipParentheses(node)
	switch inner.Kind {
	case ast.KindBinaryExpression:
		op := inner.AsBinaryExpression().OperatorToken
		if op == nil {
			return false
		}
		switch op.Kind {
		case ast.KindEqualsEqualsToken, ast.KindEqualsEqualsEqualsToken,
			ast.KindExclamationEqualsToken, ast.KindExclamationEqualsEqualsToken,
			ast.KindGreaterThanToken, ast.KindGreaterThanEqualsToken,
			ast.KindLessThanToken, ast.KindLessThanEqualsToken,
			ast.KindInKeyword, ast.KindInstanceOfKeyword:
			return true
		}
	case ast.KindPrefixUnaryExpression:
		return inner.AsPrefixUnaryExpression().Operator == ast.KindExclamationToken
	}
	return false
}

// inverseOperatorString returns the swapped equality operator text and true
// when the operator has an inverse, mirroring ESLint's OPERATOR_INVERSES map.
// (Only ==/!= and ===/!== have safe inverses; relational operators like
// `<`/`>=` do not, since both sides return false for NaN.)
func inverseOperatorString(op ast.Kind) (string, bool) {
	switch op {
	case ast.KindEqualsEqualsToken:
		return "!=", true
	case ast.KindExclamationEqualsToken:
		return "==", true
	case ast.KindEqualsEqualsEqualsToken:
		return "!==", true
	case ast.KindExclamationEqualsEqualsToken:
		return "===", true
	}
	return "", false
}

// eslintLikePrecedence is a package-local alias for utils.EslintLikePrecedence.
func eslintLikePrecedence(node *ast.Node) int {
	return utils.EslintLikePrecedence(node)
}

// isCoalesceExpression reports whether node (after parens) is a `??`
// BinaryExpression. Mirrors ESLint's astUtils.isCoalesceExpression.
func isCoalesceExpression(node *ast.Node) bool {
	if node.Kind != ast.KindBinaryExpression {
		return false
	}
	op := node.AsBinaryExpression().OperatorToken
	return op != nil && op.Kind == ast.KindQuestionQuestionToken
}

// invertExpression returns source text for the boolean inverse of testNode.
// testNode is the raw `cond.Condition` (still wrapped in any parens). Mirrors
// ESLint's invertExpression: swap ==/=== style operators in place when
// possible, otherwise prefix with `!` and parenthesize when precedence
// requires it.
func invertExpression(sf *ast.SourceFile, testNode *ast.Node) string {
	inner := ast.SkipParentheses(testNode)
	if inner.Kind == ast.KindBinaryExpression {
		bin := inner.AsBinaryExpression()
		if bin.OperatorToken != nil {
			if invOp, ok := inverseOperatorString(bin.OperatorToken.Kind); ok {
				innerRange := utils.TrimNodeTextRange(sf, inner)
				opRange := utils.TrimNodeTextRange(sf, bin.OperatorToken)
				text := sf.Text()
				return text[innerRange.Pos():opRange.Pos()] + invOp + text[opRange.End():innerRange.End()]
			}
		}
	}
	testText := utils.TrimmedNodeText(sf, testNode)
	if eslintLikePrecedence(inner) < 16 {
		return "!(" + testText + ")"
	}
	return "!" + testText
}

// matchesDefaultAssignment reports whether the conditional matches the
// pattern `id ? id : expression` after peeling parens on test and consequent.
func matchesDefaultAssignment(test, consequent *ast.Node) bool {
	innerTest := ast.SkipParentheses(test)
	innerCons := ast.SkipParentheses(consequent)
	if innerTest.Kind != ast.KindIdentifier || innerCons.Kind != ast.KindIdentifier {
		return false
	}
	return innerTest.AsIdentifier().Text == innerCons.AsIdentifier().Text
}

// buildBooleanLiteralFix returns the fix for the
// `unnecessaryConditionalExpression` case. Returns nil when no safe fix
// exists (e.g. both sides equal but test has side effects).
func buildBooleanLiteralFix(sf *ast.SourceFile, cond *ast.ConditionalExpression, consVal, altVal bool) []rule.RuleFix {
	condNode := cond.AsNode()
	if consVal == altVal {
		// `foo ? true : true` → `true`, but only when test is a bare
		// identifier (no side effects).
		if ast.SkipParentheses(cond.Condition).Kind != ast.KindIdentifier {
			return nil
		}
		text := "false"
		if consVal {
			text = "true"
		}
		return []rule.RuleFix{rule.RuleFixReplace(sf, condNode, text)}
	}
	if altVal {
		// `foo ? false : true` → `!foo` (or operator inversion).
		return []rule.RuleFix{rule.RuleFixReplace(sf, condNode, invertExpression(sf, cond.Condition))}
	}
	// `foo ? true : false` → `foo` when test is already a boolean
	// expression, else `!!foo`.
	var replacement string
	if isBooleanExpression(cond.Condition) {
		replacement = utils.TrimmedNodeText(sf, cond.Condition)
	} else {
		replacement = "!" + invertExpression(sf, cond.Condition)
	}
	return []rule.RuleFix{rule.RuleFixReplace(sf, condNode, replacement)}
}

// buildDefaultAssignmentFix returns the fix for the
// `unnecessaryConditionalAssignment` case (`a ? a : b` → `a || b`).
// The alternate gets wrapped in parens when its precedence is below `||` or
// it's a `??` expression — but only when not already parenthesized.
func buildDefaultAssignmentFix(sf *ast.SourceFile, cond *ast.ConditionalExpression) []rule.RuleFix {
	condNode := cond.AsNode()
	innerAlt := ast.SkipParentheses(cond.WhenFalse)
	alreadyParenthesised := cond.WhenFalse.Kind == ast.KindParenthesizedExpression

	shouldWrap := (eslintLikePrecedence(innerAlt) < 4 || isCoalesceExpression(innerAlt)) && !alreadyParenthesised

	var alternateText string
	if shouldWrap {
		alternateText = "(" + utils.TrimmedNodeText(sf, innerAlt) + ")"
	} else {
		alternateText = utils.TrimmedNodeText(sf, cond.WhenFalse)
	}

	testText := utils.TrimmedNodeText(sf, cond.Condition)
	return []rule.RuleFix{rule.RuleFixReplace(sf, condNode, testText+" || "+alternateText)}
}

// NoUnneededTernaryRule disallows ternary operators when simpler alternatives
// exist: `x ? true : false` → `x` / `!!x`, and (with `defaultAssignment:
// false`) `a ? a : b` → `a || b`.
var NoUnneededTernaryRule = rule.Rule{
	Name: "no-unneeded-ternary",
	Run: func(ctx rule.RuleContext, ruleOptions any) rule.RuleListeners {
		opts := parseOptions(ruleOptions)

		condExprMsg := rule.RuleMessage{
			Id:          "unnecessaryConditionalExpression",
			Description: "Unnecessary use of boolean literals in conditional expression.",
		}
		condAssignMsg := rule.RuleMessage{
			Id:          "unnecessaryConditionalAssignment",
			Description: "Unnecessary use of conditional expression for default assignment.",
		}

		return rule.RuleListeners{
			ast.KindConditionalExpression: func(node *ast.Node) {
				cond := node.AsConditionalExpression()
				if cond == nil {
					return
				}

				consVal, consOk := boolKind(cond.WhenTrue)
				altVal, altOk := boolKind(cond.WhenFalse)
				if consOk && altOk {
					if fixes := buildBooleanLiteralFix(ctx.SourceFile, cond, consVal, altVal); fixes != nil {
						ctx.ReportNodeWithFixes(node, condExprMsg, fixes...)
					} else {
						ctx.ReportNode(node, condExprMsg)
					}
					return
				}

				if !opts.defaultAssignment && matchesDefaultAssignment(cond.Condition, cond.WhenTrue) {
					ctx.ReportNodeWithFixes(node, condAssignMsg, buildDefaultAssignmentFix(ctx.SourceFile, cond)...)
				}
			},
		}
	},
}
