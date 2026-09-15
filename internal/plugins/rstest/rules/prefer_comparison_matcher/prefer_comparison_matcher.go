package prefer_comparison_matcher

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
	shared "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/prefer_comparison_matcher"
)

var PreferComparisonMatcherRule = shared.NewRule(shared.Config{
	Name: "rstest/prefer-comparison-matcher",
	// !(a > b) is not a <= b when either operand is NaN.
	PreserveNegation: true,
	// Numeric matchers reject coercion, which cannot be ruled out without type information.
	Suggestions:   true,
	IgnoreOperand: isNonNumericOperand,
	Prepare: func(ctx rule.RuleContext) func(*ast.Node) *shared.ExpectCall {
		analysis := rstestUtils.GetRstestCallAnalysis(ctx)
		return func(node *ast.Node) *shared.ExpectCall {
			parsed := analysis.ParseExpectCall(node)
			if parsed == nil || parsed.Reason != rstestUtils.RstestExpectParseReasonNone || parsed.MatcherEntry == nil || parsed.Head == nil {
				return nil
			}
			// Poll consumes a callback and element consumes a locator, not a boolean subject.
			if parsed.Entry != rstestUtils.RstestExpectEntryCall && parsed.Entry != rstestUtils.RstestExpectEntrySoft {
				return nil
			}
			if len(parsed.Matchers) == 0 || parsed.Matchers[0].Kind != rstestUtils.RstestExpectMatcherCall || testFramework.IsComputedIdentifierAccessor(parsed.MatcherEntry.Node) {
				return nil
			}
			for _, modifier := range parsed.Modifiers {
				// A relational expression produces a boolean, never a promise to resolve or reject.
				if modifier == "resolves" || modifier == "rejects" {
					return nil
				}
			}
			matcherCall := rstestUtils.MatcherCall(parsed.MatcherEntry)
			if matcherCall == nil {
				return nil
			}
			return &shared.ExpectCall{
				HeadCall: parsed.Head, MatcherCall: matcherCall, MatcherEntry: *parsed.MatcherEntry,
				Matcher: parsed.Matcher, Modifiers: parsed.ModifierEntries,
				Editable: parsed.Expression == matcherCall,
			}
		}
	},
	BuildFixes: buildSuggestion,
})

func isNonNumericOperand(node *ast.Node) bool {
	node = utils.SkipAssertionsAndParens(node)
	if utils.IsStringLiteralOrTemplate(node) {
		return true
	}
	switch node.Kind {
	case ast.KindTrueKeyword, ast.KindFalseKeyword, ast.KindNullKeyword, ast.KindRegularExpressionLiteral,
		ast.KindArrayLiteralExpression, ast.KindObjectLiteralExpression, ast.KindArrowFunction,
		ast.KindFunctionExpression, ast.KindClassExpression, ast.KindVoidExpression:
		return true
	default:
		return false
	}
}

func buildSuggestion(ctx rule.RuleContext, match shared.Match) []rule.RuleFix {
	if !match.Expect.Editable {
		return nil
	}
	headCall := match.Expect.HeadCall.AsCallExpression()
	matcherCall := match.Expect.MatcherCall.AsCallExpression()
	if headCall == nil || matcherCall == nil || hasTypeArguments(headCall) || hasTypeArguments(matcherCall) ||
		!isStaticNumericOperand(match.Left) || !isStaticNumericOperand(match.Right) {
		return nil
	}
	for _, node := range []*ast.Node{match.Comparison, match.MatcherArgument} {
		span := utils.TrimNodeTextRange(ctx.SourceFile, node)
		if utils.HasCommentInSpan(ctx.Comments.All(), span.Pos(), span.End()) {
			return nil
		}
	}
	matcherRange, matcherText, ok := testFramework.AccessorReplacement(ctx.SourceFile, match.Expect.MatcherEntry.Node, match.Matcher)
	if !ok {
		return nil
	}
	fixes := append(shared.OperandFixes(ctx, match), rule.RuleFixReplaceRange(matcherRange, matcherText))
	var not *testFramework.MemberEntry
	for i := range match.Expect.Modifiers {
		if match.Expect.Modifiers[i].Name == "not" {
			not = &match.Expect.Modifiers[i]
			break
		}
	}
	switch {
	case match.Negated && not == nil:
		span, text, ok := testFramework.InsertMemberBeforeAccessor(&match.Expect.MatcherEntry, "not")
		if !ok {
			return nil
		}
		fixes = append(fixes, rule.RuleFixReplaceRange(span, text))
	case !match.Negated && not != nil:
		ranges, ok := testFramework.RemoveAccessorEntryRanges(ctx.SourceFile, ctx.Comments.All(), not)
		if !ok {
			return nil
		}
		for _, span := range ranges {
			fixes = append(fixes, rule.RuleFixRemoveRange(span))
		}
	}
	return fixes
}

func hasTypeArguments(call *ast.CallExpression) bool {
	return call.TypeArguments != nil && len(call.TypeArguments.Nodes) > 0
}

func isStaticNumericOperand(node *ast.Node) bool {
	node = utils.SkipAssertionsAndParens(node)
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindNumericLiteral, ast.KindBigIntLiteral:
		return true
	case ast.KindPrefixUnaryExpression:
		unary := node.AsPrefixUnaryExpression()
		operand := utils.SkipAssertionsAndParens(unary.Operand)
		if operand == nil {
			return false
		}
		if unary.Operator == ast.KindMinusToken {
			return operand.Kind == ast.KindNumericLiteral || operand.Kind == ast.KindBigIntLiteral
		}
		return unary.Operator == ast.KindPlusToken && operand.Kind == ast.KindNumericLiteral
	default:
		return false
	}
}
