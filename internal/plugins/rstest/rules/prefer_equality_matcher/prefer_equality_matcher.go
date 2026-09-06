package prefer_equality_matcher

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
	shared "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/prefer_equality_matcher"
)

var PreferEqualityMatcherRule = shared.NewRule(shared.Config{
	Name: "rstest/prefer-equality-matcher",
	Prepare: func(ctx rule.RuleContext) shared.Runtime {
		analysis := rstestUtils.GetRstestCallAnalysis(ctx)
		return shared.Runtime{
			Parse: func(node *ast.Node) *shared.ExpectCall {
				parsed := analysis.ParseExpectCall(node)
				if parsed == nil ||
					parsed.Reason != rstestUtils.RstestExpectParseReasonNone ||
					parsed.Head == nil ||
					parsed.MatcherEntry == nil ||
					len(parsed.Matchers) == 0 ||
					parsed.Matchers[0].Kind != rstestUtils.RstestExpectMatcherCall ||
					testFramework.IsComputedIdentifierAccessor(parsed.MatcherEntry.Node) {
					return nil
				}

				matcherCall := rstestUtils.MatcherCall(parsed.MatcherEntry)
				if matcherCall == nil {
					return nil
				}

				return &shared.ExpectCall{
					HeadCall:     parsed.Head,
					MatcherCall:  matcherCall,
					MatcherEntry: *parsed.MatcherEntry,
					Matcher:      parsed.Matcher,
					Modifiers:    parsed.ModifierEntries,
				}
			},
		}
	},
	BuildFixes: buildSuggestionFixes,
})

func buildSuggestionFixes(ctx rule.RuleContext, match shared.Match, equalityMatcher string) []rule.RuleFix {
	comparisonRange := utils.TrimNodeTextRange(ctx.SourceFile, match.Comparison)
	if ctx.Comments != nil && utils.HasCommentInSpan(
		ctx.Comments.All(),
		comparisonRange.Pos(),
		comparisonRange.End(),
	) {
		return nil
	}

	matcherRange, matcherText, ok := testFramework.AccessorReplacement(
		ctx.SourceFile,
		match.Expect.MatcherEntry.Node,
		equalityMatcher,
	)
	if !ok {
		return nil
	}

	fixes := []rule.RuleFix{
		rule.RuleFixReplace(
			ctx.SourceFile,
			match.Comparison,
			scanner.GetSourceTextOfNodeFromSourceFile(ctx.SourceFile, ast.SkipParentheses(match.Left), false),
		),
		rule.RuleFixReplaceRange(matcherRange, matcherText),
		rule.RuleFixReplace(
			ctx.SourceFile,
			match.MatcherArgument,
			scanner.GetSourceTextOfNodeFromSourceFile(ctx.SourceFile, ast.SkipParentheses(match.Right), false),
		),
	}

	var notModifier *testFramework.MemberEntry
	for index := range match.Expect.Modifiers {
		if match.Expect.Modifiers[index].Name == "not" {
			notModifier = &match.Expect.Modifiers[index]
			break
		}
	}

	switch {
	case match.ShouldHaveNot && notModifier == nil:
		textRange, text, ok := testFramework.InsertMemberBeforeAccessor(
			&match.Expect.MatcherEntry,
			"not",
		)
		if !ok {
			return nil
		}
		fixes = append(fixes, rule.RuleFixReplaceRange(textRange, text))
	case !match.ShouldHaveNot && notModifier != nil:
		ranges, ok := testFramework.RemoveAccessorEntryRanges(
			ctx.SourceFile,
			ctx.Comments.All(),
			notModifier,
		)
		if !ok {
			return nil
		}
		for _, textRange := range ranges {
			fixes = append(fixes, rule.RuleFixRemoveRange(textRange))
		}
	}

	return fixes
}
