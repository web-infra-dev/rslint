package prefer_comparison_matcher

import (
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

type ExpectCall struct {
	HeadCall     *ast.Node
	MatcherCall  *ast.Node
	MatcherEntry testFramework.MemberEntry
	Matcher      string
	Modifiers    []testFramework.MemberEntry
	Editable     bool
}

type Match struct {
	Expect          *ExpectCall
	Comparison      *ast.Node
	Left            *ast.Node
	Right           *ast.Node
	MatcherArgument *ast.Node
	Matcher         string
	Negated         bool
}

type Config struct {
	Name             string
	Prepare          func(rule.RuleContext) func(*ast.Node) *ExpectCall
	PreserveNegation bool
	Suggestions      bool
	IgnoreOperand    func(*ast.Node) bool
	BuildFixes       func(rule.RuleContext, Match) []rule.RuleFix
}

func comparisonMatchers(operator ast.Kind) (string, string) {
	switch operator {
	case ast.KindGreaterThanToken:
		return "toBeGreaterThan", "toBeLessThanOrEqual"
	case ast.KindLessThanToken:
		return "toBeLessThan", "toBeGreaterThanOrEqual"
	case ast.KindGreaterThanEqualsToken:
		return "toBeGreaterThanOrEqual", "toBeLessThan"
	case ast.KindLessThanEqualsToken:
		return "toBeLessThanOrEqual", "toBeGreaterThan"
	default:
		return "", ""
	}
}

func NewRule(config Config) rule.Rule {
	return rule.Rule{
		Name:   config.Name,
		Schema: rule.EmptyArraySchema,
		Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
			parse := config.Prepare(ctx)
			return rule.RuleListeners{
				ast.KindCallExpression: func(node *ast.Node) {
					expect := parse(node)
					if expect == nil || !slices.Contains([]string{"toBe", "toEqual", "toStrictEqual"}, expect.Matcher) {
						return
					}
					expectArgs, matcherArgs := expect.HeadCall.Arguments(), expect.MatcherCall.Arguments()
					if len(expectArgs) == 0 || len(matcherArgs) == 0 {
						return
					}
					comparison := ast.SkipParentheses(expectArgs[0])
					if comparison.Kind != ast.KindBinaryExpression {
						return
					}
					binary := comparison.AsBinaryExpression()
					matcher, inverse := comparisonMatchers(binary.OperatorToken.Kind)
					if matcher == "" {
						return
					}
					for _, operand := range []*ast.Node{binary.Left, binary.Right} {
						inner := ast.SkipOuterExpressions(operand, ast.OEKParentheses|ast.OEKTypeAssertions)
						if utils.IsStringLiteralOrTemplate(inner) || (config.IgnoreOperand != nil && config.IgnoreOperand(operand)) {
							return
						}
					}
					boolean := ast.SkipOuterExpressions(matcherArgs[0], ast.OEKParentheses|ast.OEKTypeAssertions)
					if boolean.Kind != ast.KindTrueKeyword && boolean.Kind != ast.KindFalseKeyword {
						return
					}
					hasNot := slices.ContainsFunc(expect.Modifiers, func(entry testFramework.MemberEntry) bool { return entry.Name == "not" })
					negated := (boolean.Kind == ast.KindTrueKeyword) == hasNot
					if negated && !config.PreserveNegation {
						matcher, negated = inverse, false
					}
					preferred := matcher
					if negated {
						preferred = "not." + matcher
					}
					match := Match{expect, comparison, binary.Left, binary.Right, ast.SkipParentheses(matcherArgs[0]), matcher, negated}
					message := rule.RuleMessage{
						Id:          "useToBeComparison",
						Description: "Prefer using `" + preferred + "` instead",
						Data:        map[string]string{"preferredMatcher": preferred},
					}
					if config.Suggestions {
						ctx.ReportNodeWithDeferredSuggestions(expect.MatcherEntry.Node, message, func() []rule.RuleSuggestion {
							fixes := config.BuildFixes(ctx, match)
							if len(fixes) == 0 {
								return nil
							}
							return []rule.RuleSuggestion{{
								Message:  rule.RuleMessage{Id: "suggestComparisonMatcher", Description: "Use `" + preferred + "`"},
								FixesArr: fixes,
							}}
						})
					} else {
						ctx.ReportNodeWithDeferredFixes(expect.MatcherEntry.Node, message, func() []rule.RuleFix {
							return config.BuildFixes(ctx, match)
						})
					}
				},
			}
		},
	}
}

func OperandFixes(ctx rule.RuleContext, match Match) []rule.RuleFix {
	return []rule.RuleFix{
		rule.RuleFixReplace(ctx.SourceFile, match.Comparison, operandText(ctx.SourceFile, match.Left)),
		rule.RuleFixReplace(ctx.SourceFile, match.MatcherArgument, operandText(ctx.SourceFile, match.Right)),
	}
}

func operandText(source *ast.SourceFile, node *ast.Node) string {
	inner := ast.SkipParentheses(node)
	// Comma-expression parentheses prevent an operand from becoming several call arguments.
	if inner.Kind == ast.KindBinaryExpression && inner.AsBinaryExpression().OperatorToken.Kind == ast.KindCommaToken {
		return scanner.GetSourceTextOfNodeFromSourceFile(source, node, false)
	}
	return scanner.GetSourceTextOfNodeFromSourceFile(source, inner, false)
}
