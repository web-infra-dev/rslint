package utils

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

// PreferBooleanMatcherConfig configures the rule produced by
// [MakePreferBooleanMatcherRule].
//
// `rstest/prefer-to-be-truthy` and `rstest/prefer-to-be-falsy` are the same
// rule with the boolean flipped, so both are thin wrappers around the factory;
// do not inline the matching or the fix in either of them — extend the factory
// instead.
type PreferBooleanMatcherConfig struct {
	// RuleName is the registered rule name, e.g. "rstest/prefer-to-be-truthy".
	RuleName string

	// MessageId identifies the single diagnostic the rule reports.
	MessageId string

	// Description is that diagnostic's text.
	Description string

	// ExpectedLiteral is the boolean literal the equality matcher must be
	// given for the assertion to be rewritable: ast.KindTrueKeyword or
	// ast.KindFalseKeyword.
	ExpectedLiteral ast.Kind

	// ReplacementMatcher is the matcher the assertion is rewritten to, e.g.
	// "toBeTruthy". It takes no arguments, so the fix removes the literal.
	ReplacementMatcher string
}

// MakePreferBooleanMatcherRule produces a rule that reports an equality
// assertion against a boolean literal — `expect(value).toBe(true)` and its
// `toEqual` / `toStrictEqual` spellings — and rewrites it to the dedicated
// truthiness matcher.
//
// The diagnostic is reported on the matcher accessor rather than the whole
// assertion, because that is the only part the fix rewrites.
func MakePreferBooleanMatcherRule(cfg PreferBooleanMatcherConfig) rule.Rule {
	message := rule.RuleMessage{Id: cfg.MessageId, Description: cfg.Description}

	return rule.Rule{
		Name:   cfg.RuleName,
		Schema: rule.EmptyArraySchema,
		Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
			analysis := GetRstestCallAnalysis(ctx)
			return rule.RuleListeners{
				ast.KindCallExpression: func(node *ast.Node) {
					matcher, matcherCall, ok := booleanEqualityMatcherEntry(
						analysis.ParseExpectCall(node),
						cfg.ExpectedLiteral,
					)
					if !ok {
						return
					}

					ctx.ReportNodeWithDeferredFixes(matcher.Node, message, func() []rule.RuleFix {
						nameRange, nameText, ok := testFramework.AccessorReplacement(
							ctx.SourceFile,
							matcher.Node,
							cfg.ReplacementMatcher,
						)
						if !ok {
							return nil
						}

						argumentsRange, ok := MatcherArgumentListRange(ctx.SourceFile, matcherCall)
						if !ok {
							return nil
						}
						// The replacement matcher takes no arguments, so the
						// literal and everything written alongside it goes. A
						// comment between the parentheses would go with it, so
						// the rename is withheld too and the diagnostic stands
						// alone rather than the fix silently dropping what the
						// author wrote.
						if utils.HasCommentInSpan(ctx.Comments.All(), argumentsRange.Pos(), argumentsRange.End()) {
							return nil
						}

						return []rule.RuleFix{
							rule.RuleFixReplaceRange(nameRange, nameText),
							rule.RuleFixRemoveRange(argumentsRange),
						}
					})
				},
			}
		},
	}
}

// booleanEqualityMatcherEntry returns the matcher entry of an equality
// assertion whose single argument is the given boolean literal, together with
// the call expression that invokes that matcher.
//
// The matcher must be invoked: `expect(value).toBe` asserts nothing, and Chai's
// truthiness assertions are property getters (`expect(value).to.be.true`) whose
// equality forms are named `equal` rather than `toBe`. Both are excluded by
// requiring RstestExpectMatcherCall.
//
// An assertion factory must also have run. Head is nil for `expect.toBe(true)`
// and friends, where the author never passed a value to assert on; the
// assertion asserts nothing whichever matcher it names, so renaming the matcher
// would only make a broken assertion look right.
func booleanEqualityMatcherEntry(
	parsed *ParsedRstestExpectCall,
	expectedLiteral ast.Kind,
) (ParsedRstestFnMemberEntry, *ast.Node, bool) {
	if parsed == nil ||
		parsed.Reason != RstestExpectParseReasonNone ||
		parsed.Head == nil ||
		parsed.MatcherEntry == nil ||
		!RSTEST_EQUALITY_MATCHERS[parsed.Matcher] ||
		len(parsed.Matchers) == 0 ||
		parsed.Matchers[0].Kind != RstestExpectMatcherCall {
		return ParsedRstestFnMemberEntry{}, nil, false
	}

	// A computed identifier key names its matcher at runtime, so `toBe` here is
	// the name of a variable rather than of a matcher. Reporting it would flag
	// whatever assertion that variable holds.
	if parsed.MatcherEntry.Node == nil || testFramework.IsComputedIdentifierAccessor(parsed.MatcherEntry.Node) {
		return ParsedRstestFnMemberEntry{}, nil, false
	}

	call := MatcherCall(parsed.MatcherEntry)
	if call == nil {
		return ParsedRstestFnMemberEntry{}, nil, false
	}
	arguments := call.AsCallExpression().Arguments
	if arguments == nil || len(arguments.Nodes) != 1 {
		return ParsedRstestFnMemberEntry{}, nil, false
	}
	// Type assertions are transparent, so `toBe(true as boolean)` is the same
	// assertion as `toBe(true)`. A spread argument is not a literal and never
	// matches; neither does `!0`, which is an operator applied to a literal
	// rather than a literal.
	argument := FollowTypeAssertionChain(arguments.Nodes[0])
	if argument == nil || argument.Kind != expectedLiteral {
		return ParsedRstestFnMemberEntry{}, nil, false
	}

	return *parsed.MatcherEntry, call, true
}
