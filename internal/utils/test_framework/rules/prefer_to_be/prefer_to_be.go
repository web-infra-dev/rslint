// Package prefer_to_be implements the matcher-independent decision tree shared
// by Jest- and Rstest-flavoured prefer-to-be rules. Framework adapters own
// expect-call parsing, diagnostic wording and source edits.
package prefer_to_be

import (
	"math"
	"strconv"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

type Kind uint8

const (
	KindToBe Kind = iota
	KindNull
	KindNaN
	KindUndefined
	KindDefined
)

type ExpectCall struct {
	Matcher      string
	MatcherEntry testFramework.MemberEntry
	MatcherCall  *ast.Node
	Modifiers    []testFramework.MemberEntry
}

type Runtime struct {
	Parse func(*ast.Node) *ExpectCall
}

type Match struct {
	Expect   *ExpectCall
	Kind     Kind
	NotEntry *testFramework.MemberEntry
}

type Config struct {
	Name string

	Prepare func(rule.RuleContext) Runtime
	Message func(Kind) rule.RuleMessage
	// BuildFixes is called only when autofixes are requested. Returning nil
	// leaves the eager diagnostic intact without attaching an unsafe edit.
	BuildFixes func(rule.RuleContext, Match) []rule.RuleFix

	// AllowFractionalNumbers selects Jest's policy. Rstest leaves it false to
	// follow Vitest and Rstest's recommendation to use toBeCloseTo for decimal
	// literals. Negative decimals use the same policy as positive decimals.
	AllowFractionalNumbers bool
	// IsSpecialIdentifier can reject locally shadowed undefined/NaN names. A
	// nil callback preserves eslint-plugin-jest's text-only upstream behavior.
	IsSpecialIdentifier func(rule.RuleContext, *ast.Node, string) bool
}

var equalityMatchers = map[string]bool{
	"toBe":          true,
	"toEqual":       true,
	"toStrictEqual": true,
}

func firstArgument(call *ExpectCall) *ast.Node {
	if call == nil || call.MatcherCall == nil {
		return nil
	}
	arguments := call.MatcherCall.Arguments()
	if len(arguments) == 0 {
		return nil
	}
	return testFramework.FollowTypeAssertionChain(arguments[0])
}

func findNot(modifiers []testFramework.MemberEntry) *testFramework.MemberEntry {
	for index := range modifiers {
		if modifiers[index].Name == "not" {
			return &modifiers[index]
		}
	}
	return nil
}

func isIdentifier(node *ast.Node, name string) bool {
	return node != nil && node.Kind == ast.KindIdentifier && node.AsIdentifier().Text == name
}

func isFractionalNumericLiteral(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindNumericLiteral {
		return false
	}
	value, err := strconv.ParseFloat(node.AsNumericLiteral().Text, 64)
	return err == nil && math.Floor(value) != math.Ceil(value)
}

func shouldUseToBe(argument *ast.Node, allowFractionalNumbers bool) bool {
	if argument == nil {
		return false
	}
	if argument.Kind == ast.KindPrefixUnaryExpression {
		unary := argument.AsPrefixUnaryExpression()
		if unary.Operator == ast.KindMinusToken {
			// Upstream only peels the unary operator here. A type assertion
			// nested inside it remains a non-literal; an assertion around the
			// entire negative number was already removed by firstArgument.
			argument = ast.SkipParentheses(unary.Operand)
		}
	}
	if argument == nil || (!allowFractionalNumbers && isFractionalNumericLiteral(argument)) {
		return false
	}

	switch argument.Kind {
	case ast.KindStringLiteral,
		ast.KindNumericLiteral,
		ast.KindBigIntLiteral,
		ast.KindTrueKeyword,
		ast.KindFalseKeyword,
		ast.KindNullKeyword,
		ast.KindNoSubstitutionTemplateLiteral,
		ast.KindTemplateExpression:
		return true
	default:
		return false
	}
}

func match(ctx rule.RuleContext, call *ExpectCall, config Config) (Match, bool) {
	if call == nil || call.MatcherEntry.Node == nil || call.MatcherCall == nil {
		return Match{}, false
	}
	notEntry := findNot(call.Modifiers)
	if notEntry != nil {
		switch call.Matcher {
		case "toBeUndefined":
			return Match{Expect: call, Kind: KindDefined, NotEntry: notEntry}, true
		case "toBeDefined":
			return Match{Expect: call, Kind: KindUndefined, NotEntry: notEntry}, true
		}
	}

	if !equalityMatchers[call.Matcher] {
		return Match{}, false
	}
	argument := firstArgument(call)
	if argument == nil {
		return Match{}, false
	}

	result := Match{Expect: call}
	switch {
	case argument.Kind == ast.KindNullKeyword:
		result.Kind = KindNull
	case isIdentifier(argument, "undefined") &&
		(config.IsSpecialIdentifier == nil || config.IsSpecialIdentifier(ctx, argument, "undefined")):
		if notEntry != nil {
			result.Kind = KindDefined
			result.NotEntry = notEntry
		} else {
			result.Kind = KindUndefined
		}
	case isIdentifier(argument, "NaN") &&
		(config.IsSpecialIdentifier == nil || config.IsSpecialIdentifier(ctx, argument, "NaN")):
		result.Kind = KindNaN
	case call.Matcher != "toBe" && shouldUseToBe(argument, config.AllowFractionalNumbers):
		result.Kind = KindToBe
	default:
		return Match{}, false
	}
	return result, true
}

func NewRule(config Config) rule.Rule {
	return rule.Rule{
		Name:   config.Name,
		Schema: rule.EmptyArraySchema,
		Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
			runtime := config.Prepare(ctx)
			return rule.RuleListeners{
				ast.KindCallExpression: func(node *ast.Node) {
					if runtime.Parse == nil || config.Message == nil {
						return
					}
					matched, ok := match(ctx, runtime.Parse(node), config)
					if !ok {
						return
					}

					message := config.Message(matched.Kind)
					ctx.ReportNodeWithDeferredFixes(
						matched.Expect.MatcherEntry.Node,
						message,
						func() []rule.RuleFix {
							if config.BuildFixes == nil {
								return nil
							}
							return config.BuildFixes(ctx, matched)
						},
					)
				},
			}
		},
	}
}
