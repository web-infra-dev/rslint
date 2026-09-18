// Package prefer_to_have_length implements the matcher-independent body shared
// by test-framework rules that prefer toHaveLength over equality checks of a
// value's length property.
package prefer_to_have_length

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

type ExpectCall struct {
	HeadCall     *ast.Node
	MatcherCall  *ast.Node
	MatcherEntry testFramework.MemberEntry
	Matcher      string
	CanFix       bool
}

type Runtime struct {
	Parse func(*ast.Node) []*ExpectCall
}

type Match struct {
	Expect      *ExpectCall
	Receiver    *ast.Node
	LengthEntry testFramework.MemberEntry
}

type Config struct {
	Name                string
	Prepare             func(rule.RuleContext) Runtime
	ParseLengthAccessor func(*ast.Node) (*ast.Node, testFramework.MemberEntry, bool)
	BuildFixes          func(rule.RuleContext, Match) []rule.RuleFix
}

var useToHaveLengthMessage = rule.RuleMessage{
	Id:          "useToHaveLength",
	Description: "Use `toHaveLength()` instead",
}

func isEqualityMatcher(name string) bool {
	return name == "toBe" || name == "toEqual" || name == "toStrictEqual"
}

func lengthAccess(node *ast.Node) (receiver *ast.Node, entry testFramework.MemberEntry, ok bool) {
	node = ast.SkipParentheses(node)
	if node == nil || ast.IsOptionalChain(node) {
		return nil, testFramework.MemberEntry{}, false
	}

	var name *ast.Node
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		property := node.AsPropertyAccessExpression()
		receiver, name = property.Expression, property.Name()
	case ast.KindElementAccessExpression:
		element := node.AsElementAccessExpression()
		receiver, name = element.Expression, ast.SkipParentheses(element.ArgumentExpression)
	default:
		return nil, testFramework.MemberEntry{}, false
	}

	staticName, static := utils.AccessExpressionStaticName(node)
	if !static || staticName != "length" || name == nil || receiver == nil {
		return nil, testFramework.MemberEntry{}, false
	}
	return receiver, testFramework.MemberEntry{Name: "length", Node: name}, true
}

func NewRule(config Config) rule.Rule {
	return rule.Rule{
		Name:   config.Name,
		Schema: rule.EmptyArraySchema,
		Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
			runtime := config.Prepare(ctx)
			parseLengthAccessor := config.ParseLengthAccessor
			if parseLengthAccessor == nil {
				parseLengthAccessor = lengthAccess
			}
			return rule.RuleListeners{
				ast.KindCallExpression: func(node *ast.Node) {
					if runtime.Parse == nil {
						return
					}
					for _, parsed := range runtime.Parse(node) {
						if parsed == nil || parsed.HeadCall == nil || parsed.MatcherCall == nil ||
							parsed.MatcherEntry.Node == nil || !isEqualityMatcher(parsed.Matcher) {
							continue
						}
						headArguments := parsed.HeadCall.Arguments()
						if len(headArguments) == 0 {
							continue
						}
						receiver, lengthEntry, ok := parseLengthAccessor(headArguments[0])
						if !ok {
							continue
						}

						match := Match{Expect: parsed, Receiver: receiver, LengthEntry: lengthEntry}
						ctx.ReportNodeWithDeferredFixes(
							parsed.MatcherEntry.Node,
							useToHaveLengthMessage,
							func() []rule.RuleFix {
								if config.BuildFixes == nil {
									return nil
								}
								return config.BuildFixes(ctx, match)
							},
						)
					}
				},
			}
		},
	}
}
