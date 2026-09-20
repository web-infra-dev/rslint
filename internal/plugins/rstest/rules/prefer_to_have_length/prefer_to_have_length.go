package prefer_to_have_length

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
	shared "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/prefer_to_have_length"
)

func endsAtOnlyMatcher(parsed *rstestUtils.ParsedRstestExpectCall) bool {
	return len(parsed.Matchers) == 1 &&
		len(parsed.MemberEntries) > 0 &&
		parsed.MemberEntries[len(parsed.MemberEntries)-1].Node == parsed.MatcherEntry.Node
}

// These Chai matchers replace the assertion object's current value. A later
// equality matcher therefore no longer compares the value originally passed
// to expect(), so prefer-to-have-length must stop at this boundary.
var subjectMutatingChaiMatchers = map[string]bool{
	"property":                  true,
	"ownProperty":               true,
	"haveOwnProperty":           true,
	"ownPropertyDescriptor":     true,
	"haveOwnPropertyDescriptor": true,
	"toContain":                 true,
	"toThrow":                   true,
	"toThrowError":              true,
	"throw":                     true,
	"throws":                    true,
	"Throw":                     true,
}

func isDiscardedAssertion(node *ast.Node) bool {
	for parent := node.Parent; parent != nil; parent = parent.Parent {
		if parent.Kind == ast.KindParenthesizedExpression {
			continue
		}
		return parent.Kind == ast.KindExpressionStatement
	}
	return false
}

func isSafeExpectedLength(node *ast.Node) bool {
	_, negativeZero, ok := staticNumericLiteral(node)
	return ok && !negativeZero
}

func staticNumericLiteral(node *ast.Node) (zero, negativeZero, ok bool) {
	for node != nil {
		node = ast.SkipParentheses(node)
		switch node.Kind {
		case ast.KindAsExpression:
			node = node.AsAsExpression().Expression
		case ast.KindTypeAssertionExpression:
			node = node.AsTypeAssertion().Expression
		case ast.KindSatisfiesExpression:
			node = node.AsSatisfiesExpression().Expression
		case ast.KindNonNullExpression:
			node = node.AsNonNullExpression().Expression
		case ast.KindNumericLiteral:
			return utils.NormalizeNumericLiteral(node.AsNumericLiteral().Text) == "0", false, true
		case ast.KindPrefixUnaryExpression:
			unary := node.AsPrefixUnaryExpression()
			if unary == nil || (unary.Operator != ast.KindPlusToken && unary.Operator != ast.KindMinusToken) {
				return false, false, false
			}
			zero, negativeZero, ok := staticNumericLiteral(unary.Operand)
			if !ok {
				return false, false, false
			}
			if zero && unary.Operator == ast.KindMinusToken {
				negativeZero = !negativeZero
			}
			return zero, negativeZero, true
		default:
			return false, false, false
		}
	}
	return false, false, false
}

func isSafeLengthReceiver(node *ast.Node) bool {
	for node != nil {
		node = ast.SkipParentheses(node)
		switch node.Kind {
		case ast.KindAsExpression:
			node = node.AsAsExpression().Expression
		case ast.KindTypeAssertionExpression:
			node = node.AsTypeAssertion().Expression
		case ast.KindSatisfiesExpression:
			node = node.AsSatisfiesExpression().Expression
		case ast.KindNonNullExpression:
			node = node.AsNonNullExpression().Expression
		case ast.KindArrayLiteralExpression, ast.KindStringLiteral,
			ast.KindNoSubstitutionTemplateLiteral, ast.KindFunctionExpression,
			ast.KindArrowFunction:
			return true
		default:
			return false
		}
	}
	return false
}

func buildFixes(ctx rule.RuleContext, match shared.Match) []rule.RuleFix {
	parsed := match.Expect
	if !parsed.CanFix || parsed.Matcher != "toBe" {
		return nil
	}
	headArguments := parsed.HeadCall.Arguments()
	matcherArguments := parsed.MatcherCall.Arguments()
	if len(headArguments) != 1 || len(matcherArguments) != 1 || !isSafeLengthReceiver(match.Receiver) ||
		!isSafeExpectedLength(matcherArguments[0]) {
		return nil
	}

	lengthRanges, ok := testFramework.RemoveAccessorEntryRanges(ctx.SourceFile, ctx.Comments.All(), &match.LengthEntry)
	if !ok {
		return nil
	}
	matcherRange, matcherText, ok := testFramework.AccessorReplacement(ctx.SourceFile, parsed.MatcherEntry.Node, "toHaveLength")
	if !ok {
		return nil
	}

	fixes := make([]rule.RuleFix, 0, len(lengthRanges)+2)
	for _, textRange := range lengthRanges {
		fixes = append(fixes, rule.RuleFixRemoveRange(textRange))
	}
	fixes = append(fixes, rule.RuleFixReplaceRange(matcherRange, matcherText))

	for _, call := range []*ast.Node{parsed.HeadCall, parsed.MatcherCall} {
		if call.AsCallExpression().TypeArguments != nil {
			typeRange, ok := testFramework.CallTypeArgumentListRange(ctx.SourceFile, call)
			if !ok || utils.HasCommentInSpan(ctx.Comments.All(), typeRange.Pos(), typeRange.End()) {
				return nil
			}
			fixes = append(fixes, rule.RuleFixRemoveRange(typeRange))
		}
	}
	return fixes
}

var PreferToHaveLengthRule = shared.NewRule(shared.Config{
	Name:       "rstest/prefer-to-have-length",
	BuildFixes: buildFixes,
	Prepare: func(ctx rule.RuleContext) shared.Runtime {
		analysis := rstestUtils.GetRstestCallAnalysis(ctx)
		return shared.Runtime{Parse: func(node *ast.Node) []*shared.ExpectCall {
			parsed := analysis.ParseExpectCall(node)
			if parsed == nil ||
				parsed.Reason != rstestUtils.RstestExpectParseReasonNone ||
				parsed.Head == nil ||
				(parsed.Entry != rstestUtils.RstestExpectEntryCall && parsed.Entry != rstestUtils.RstestExpectEntrySoft) ||
				parsed.MatcherEntry == nil {
				return nil
			}
			for _, modifier := range parsed.Modifiers {
				if modifier != "not" {
					return nil
				}
			}

			onlyTerminalMatcher := endsAtOnlyMatcher(parsed)
			matches := make([]*shared.ExpectCall, 0, len(parsed.Matchers))
			for index := range parsed.Matchers {
				matcher := &parsed.Matchers[index]
				if subjectMutatingChaiMatchers[matcher.Name] {
					break
				}
				if matcher.Kind != rstestUtils.RstestExpectMatcherCall ||
					testFramework.IsComputedIdentifierAccessor(matcher.Entry.Node) {
					continue
				}
				matcherCall := rstestUtils.MatcherCall(&matcher.Entry)
				if matcherCall == nil {
					continue
				}
				matches = append(matches, &shared.ExpectCall{
					HeadCall:     parsed.Head,
					MatcherCall:  matcherCall,
					MatcherEntry: matcher.Entry,
					Matcher:      matcher.Name,
					CanFix:       onlyTerminalMatcher && parsed.Expression == matcherCall && isDiscardedAssertion(matcherCall),
				})
			}
			return matches
		}}
	},
})
