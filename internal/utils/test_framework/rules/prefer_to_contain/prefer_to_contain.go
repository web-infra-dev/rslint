// Package prefer_to_contain implements the matcher-independent decision tree
// shared by Jest- and Rstest-flavoured prefer-to-contain rules. Framework
// adapters own expect-call parsing and framework-specific semantic exclusions.
package prefer_to_contain

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
	Modifiers    []testFramework.MemberEntry
}

type Runtime struct {
	Parse func(*ast.Node) *ExpectCall
}

type Match struct {
	Expect        *ExpectCall
	IncludesCall  *ast.Node
	Receiver      *ast.Node
	Item          *ast.Node
	MatcherValue  *ast.Node
	ShouldHaveNot bool
	NotEntry      *testFramework.MemberEntry
}

type Config struct {
	Name string

	Prepare func(rule.RuleContext) Runtime
	// IgnoreMatch applies framework-specific semantic exclusions after the
	// syntax has been recognized.
	IgnoreMatch func(rule.RuleContext, Match) bool
	// BuildFixes lets an adapter preserve an established framework-specific
	// edit contract. A nil callback uses the shared safe edit builder.
	BuildFixes func(rule.RuleContext, Match) []rule.RuleFix
}

var equalityMatchers = map[string]bool{
	"toBe":          true,
	"toEqual":       true,
	"toStrictEqual": true,
}

func message() rule.RuleMessage {
	return rule.RuleMessage{Id: "useToContain", Description: "Use toContain() instead"}
}

func findNot(modifiers []testFramework.MemberEntry) *testFramework.MemberEntry {
	for index := range modifiers {
		if modifiers[index].Name == "not" {
			return &modifiers[index]
		}
	}
	return nil
}

func staticAccessor(node *ast.Node, name string) (*ast.Node, bool) {
	node = ast.SkipParentheses(node)
	if node == nil {
		return nil, false
	}
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		property := node.AsPropertyAccessExpression()
		member := property.Name()
		return property.Expression, member != nil && member.Kind == ast.KindIdentifier && member.AsIdentifier().Text == name
	case ast.KindElementAccessExpression:
		element := node.AsElementAccessExpression()
		member := ast.SkipParentheses(element.ArgumentExpression)
		if member == nil {
			return nil, false
		}
		switch member.Kind {
		case ast.KindStringLiteral:
			return element.Expression, member.AsStringLiteral().Text == name
		case ast.KindNoSubstitutionTemplateLiteral:
			return element.Expression, member.AsNoSubstitutionTemplateLiteral().Text == name
		default:
			return nil, false
		}
	default:
		return nil, false
	}
}

func match(call *ExpectCall) (Match, bool) {
	if call == nil || call.HeadCall == nil || call.MatcherCall == nil ||
		call.MatcherEntry.Node == nil || !equalityMatchers[call.Matcher] {
		return Match{}, false
	}

	headArguments := call.HeadCall.Arguments()
	matcherArguments := call.MatcherCall.Arguments()
	if len(headArguments) == 0 || len(matcherArguments) != 1 {
		return Match{}, false
	}

	includesCall := ast.SkipParentheses(headArguments[0])
	if includesCall == nil || includesCall.Kind != ast.KindCallExpression || ast.IsOptionalChain(includesCall) {
		return Match{}, false
	}
	includesExpression := includesCall.AsCallExpression()
	receiver, ok := staticAccessor(includesExpression.Expression, "includes")
	if !ok || ast.IsOptionalChain(ast.SkipParentheses(includesExpression.Expression)) {
		return Match{}, false
	}
	includesArguments := includesCall.Arguments()
	if len(includesArguments) != 1 || includesArguments[0].Kind == ast.KindSpreadElement {
		return Match{}, false
	}

	matcherValue := testFramework.FollowTypeAssertionChain(matcherArguments[0])
	if matcherValue == nil || matcherArguments[0].Kind == ast.KindSpreadElement {
		return Match{}, false
	}
	value := false
	switch matcherValue.Kind {
	case ast.KindTrueKeyword:
		value = true
	case ast.KindFalseKeyword:
	default:
		return Match{}, false
	}

	notEntry := findNot(call.Modifiers)
	return Match{
		Expect: call, IncludesCall: includesCall, Receiver: receiver,
		Item: includesArguments[0], MatcherValue: matcherArguments[0],
		ShouldHaveNot: value == (notEntry != nil), NotEntry: notEntry,
	}, true
}

func hasComment(ctx rule.RuleContext, node *ast.Node) bool {
	if ctx.Comments == nil || node == nil {
		return false
	}
	span := utils.TrimNodeTextRange(ctx.SourceFile, node)
	return utils.HasCommentInSpan(ctx.Comments.All(), span.Pos(), span.End())
}

func callArgumentsHaveComment(ctx rule.RuleContext, call *ast.Node) bool {
	if ctx.Comments == nil {
		return false
	}
	span, ok := testFramework.CallArgumentListRange(ctx.SourceFile, call)
	return ok && utils.HasCommentInSpan(ctx.Comments.All(), span.Pos(), span.End())
}

func BuildFixes(ctx rule.RuleContext, matched Match) []rule.RuleFix {
	// Moving the item from inside expect(...) to the matcher call changes its
	// position relative to a custom message expression. Report the style issue
	// but do not reorder potentially observable evaluations.
	if len(matched.Expect.HeadCall.Arguments()) != 1 ||
		hasComment(ctx, matched.IncludesCall) ||
		callArgumentsHaveComment(ctx, matched.IncludesCall) ||
		callArgumentsHaveComment(ctx, matched.Expect.MatcherCall) {
		return nil
	}

	matcherRange, matcherText, ok := testFramework.AccessorReplacement(
		ctx.SourceFile,
		matched.Expect.MatcherEntry.Node,
		"toContain",
	)
	if !ok {
		return nil
	}

	fixes := []rule.RuleFix{
		rule.RuleFixReplace(ctx.SourceFile, matched.IncludesCall, utils.TrimmedNodeText(ctx.SourceFile, matched.Receiver)),
		rule.RuleFixReplaceRange(matcherRange, matcherText),
		rule.RuleFixReplace(ctx.SourceFile, matched.MatcherValue, utils.TrimmedNodeText(ctx.SourceFile, matched.Item)),
	}

	switch {
	case matched.ShouldHaveNot && matched.NotEntry == nil:
		span, text, ok := testFramework.InsertMemberBeforeAccessor(&matched.Expect.MatcherEntry, "not")
		if !ok {
			return nil
		}
		fixes = append(fixes, rule.RuleFixReplaceRange(span, text))
	case !matched.ShouldHaveNot && matched.NotEntry != nil:
		ranges, ok := testFramework.RemoveAccessorEntryRanges(ctx.SourceFile, ctx.Comments.All(), matched.NotEntry)
		if !ok {
			return nil
		}
		for _, span := range ranges {
			fixes = append(fixes, rule.RuleFixRemoveRange(span))
		}
	}
	return fixes
}

func NewRule(config Config) rule.Rule {
	return rule.Rule{
		Name:   config.Name,
		Schema: rule.EmptyArraySchema,
		Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
			runtime := config.Prepare(ctx)
			fixBuilder := config.BuildFixes
			if fixBuilder == nil {
				fixBuilder = BuildFixes
			}
			reported := map[*ast.Node]struct{}{}
			return rule.RuleListeners{
				ast.KindCallExpression: func(node *ast.Node) {
					if runtime.Parse == nil {
						return
					}
					matched, ok := match(runtime.Parse(node))
					if !ok || (config.IgnoreMatch != nil && config.IgnoreMatch(ctx, matched)) {
						return
					}
					if _, seen := reported[matched.Expect.MatcherCall]; seen {
						return
					}
					reported[matched.Expect.MatcherCall] = struct{}{}
					ctx.ReportNodeWithDeferredFixes(
						matched.Expect.MatcherEntry.Node,
						message(),
						func() []rule.RuleFix { return fixBuilder(ctx, matched) },
					)
				},
			}
		},
	}
}
