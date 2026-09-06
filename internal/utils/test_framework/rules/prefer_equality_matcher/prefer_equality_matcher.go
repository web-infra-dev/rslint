// Package prefer_equality_matcher holds the framework-neutral matching and
// suggestion planning shared by jest/prefer-equality-matcher and
// rstest/prefer-equality-matcher. Framework adapters own expect parsing and
// source edits; this package owns the strict-comparison contract and truth
// table.
package prefer_equality_matcher

import (
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

var equalityMatchers = []string{"toBe", "toEqual", "toStrictEqual"}

// ExpectCall is the framework-neutral projection of one parsed expect
// assertion. Adapters deliberately project only the matcher shape their
// framework wants this rule to inspect.
type ExpectCall struct {
	HeadCall     *ast.Node
	MatcherCall  *ast.Node
	MatcherEntry testFramework.MemberEntry
	Matcher      string
	Modifiers    []testFramework.MemberEntry
}

type Runtime struct {
	Parse func(*ast.Node) *ExpectCall
}

// Match contains the source nodes and normalized modifier decision needed by
// a framework-specific suggestion builder.
type Match struct {
	Expect          *ExpectCall
	Comparison      *ast.Node
	Left            *ast.Node
	Right           *ast.Node
	MatcherArgument *ast.Node
	ShouldHaveNot   bool
	ModifierText    string
}

type Config struct {
	Name       string
	Prepare    func(rule.RuleContext) Runtime
	BuildFixes func(rule.RuleContext, Match, string) []rule.RuleFix
}

func buildUseEqualityMatcherMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "useEqualityMatcher",
		Description: "Prefer using one of the equality matchers instead",
	}
}

func buildSuggestEqualityMatcherMessage(equalityMatcher string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "suggestEqualityMatcher",
		Description: "Use `" + equalityMatcher + "`",
	}
}

func parseStrictEqualityComparison(node *ast.Node) (left, right *ast.Node, negated bool, ok bool) {
	node = ast.SkipParentheses(node)
	if node == nil || node.Kind != ast.KindBinaryExpression {
		return nil, nil, false, false
	}

	binary := node.AsBinaryExpression()
	switch binary.OperatorToken.Kind {
	case ast.KindEqualsEqualsEqualsToken:
		return binary.Left, binary.Right, false, true
	case ast.KindExclamationEqualsEqualsToken:
		return binary.Left, binary.Right, true, true
	default:
		return nil, nil, false, false
	}
}

func unwrapBooleanLiteral(node *ast.Node) (literal *ast.Node, value bool, ok bool) {
	for node != nil {
		node = ast.SkipParentheses(node)
		switch node.Kind {
		case ast.KindAsExpression:
			node = node.AsAsExpression().Expression
		case ast.KindTypeAssertionExpression:
			node = node.AsTypeAssertion().Expression
		case ast.KindTrueKeyword:
			return node, true, true
		case ast.KindFalseKeyword:
			return node, false, true
		default:
			return nil, false, false
		}
	}
	return nil, false, false
}

func modifierText(modifiers []testFramework.MemberEntry, addNot bool) string {
	text := ""
	for _, modifier := range modifiers {
		if modifier.Name != "not" {
			text = "." + modifier.Name
			break
		}
	}
	if addNot {
		text += ".not"
	}
	return text
}

func shouldAddNot(comparisonNegated, matcherValue, hasNot bool) bool {
	return (comparisonNegated != matcherValue) == hasNot
}

func NewRule(config Config) rule.Rule {
	return rule.Rule{
		Name:   config.Name,
		Schema: rule.EmptyArraySchema,
		Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
			runtime := config.Prepare(ctx)
			return rule.RuleListeners{
				ast.KindCallExpression: func(node *ast.Node) {
					if runtime.Parse == nil || config.BuildFixes == nil {
						return
					}
					expectCall := runtime.Parse(node)
					if expectCall == nil ||
						expectCall.HeadCall == nil ||
						expectCall.MatcherCall == nil ||
						expectCall.MatcherEntry.Node == nil ||
						!slices.Contains(equalityMatchers, expectCall.Matcher) {
						return
					}

					expectArguments := expectCall.HeadCall.Arguments()
					if len(expectArguments) == 0 {
						return
					}
					comparison := ast.SkipParentheses(expectArguments[0])
					left, right, comparisonNegated, ok := parseStrictEqualityComparison(comparison)
					if !ok {
						return
					}

					matcherArguments := expectCall.MatcherCall.Arguments()
					if len(matcherArguments) == 0 {
						return
					}
					matcherArgument, matcherValue, ok := unwrapBooleanLiteral(matcherArguments[0])
					if !ok {
						return
					}

					hasNot := slices.ContainsFunc(expectCall.Modifiers, func(entry testFramework.MemberEntry) bool {
						return entry.Name == "not"
					})
					shouldHaveNot := shouldAddNot(comparisonNegated, matcherValue, hasNot)
					match := Match{
						Expect:          expectCall,
						Comparison:      comparison,
						Left:            left,
						Right:           right,
						MatcherArgument: matcherArgument,
						ShouldHaveNot:   shouldHaveNot,
						ModifierText:    modifierText(expectCall.Modifiers, shouldHaveNot),
					}

					ctx.ReportNodeWithDeferredSuggestions(
						expectCall.MatcherEntry.Node,
						buildUseEqualityMatcherMessage(),
						func() []rule.RuleSuggestion {
							suggestions := make([]rule.RuleSuggestion, 0, len(equalityMatchers))
							for _, equalityMatcher := range equalityMatchers {
								fixes := config.BuildFixes(ctx, match, equalityMatcher)
								if len(fixes) == 0 {
									return nil
								}
								suggestions = append(suggestions, rule.RuleSuggestion{
									Message:  buildSuggestEqualityMatcherMessage(equalityMatcher),
									FixesArr: fixes,
								})
							}
							return suggestions
						},
					)
				},
			}
		},
	}
}
