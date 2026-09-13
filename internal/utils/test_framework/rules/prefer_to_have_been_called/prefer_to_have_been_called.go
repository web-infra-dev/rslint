package prefer_to_have_been_called

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

type Assertion struct {
	Matcher testFramework.MemberEntry
	Not     *testFramework.MemberEntry
	CanFix  bool
}

type Config struct {
	Name    string
	Prepare func(rule.RuleContext) func(*ast.Node) []Assertion
}

func NewRule(config Config) rule.Rule {
	return rule.Rule{
		Name:   config.Name,
		Schema: rule.EmptyArraySchema,
		Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
			parse := config.Prepare(ctx)
			return rule.RuleListeners{
				ast.KindCallExpression: func(node *ast.Node) {
					for _, assertion := range parse(node) {
						if assertion.Matcher.Name != "toHaveBeenCalledTimes" && assertion.Matcher.Name != "toBeCalledTimes" {
							continue
						}
						call := testFramework.InvokedAccessorCall(&assertion.Matcher)
						if call == nil || call.AsCallExpression().Arguments == nil || len(call.AsCallExpression().Arguments.Nodes) == 0 {
							continue
						}
						argument := testFramework.FollowTypeAssertionChain(call.AsCallExpression().Arguments.Nodes[0])
						if argument == nil || argument.Kind != ast.KindNumericLiteral || utils.NormalizeNumericLiteral(argument.Text()) != "0" {
							continue
						}
						ctx.ReportNodeWithDeferredFixes(assertion.Matcher.Node, rule.RuleMessage{
							Id: "preferMatcher", Description: "Use `toHaveBeenCalled`",
						}, func() []rule.RuleFix {
							if !assertion.CanFix || len(call.AsCallExpression().Arguments.Nodes) != 1 {
								return nil
							}
							nameRange, name, ok := testFramework.AccessorReplacement(ctx.SourceFile, assertion.Matcher.Node, "toHaveBeenCalled")
							if !ok {
								return nil
							}
							argumentsRange, ok := testFramework.CallArgumentListRange(ctx.SourceFile, call)
							if !ok || utils.HasCommentInSpan(ctx.Comments.All(), argumentsRange.Pos(), argumentsRange.End()) {
								return nil
							}
							fixes := []rule.RuleFix{rule.RuleFixReplaceRange(nameRange, name), rule.RuleFixRemoveRange(argumentsRange)}
							if call.AsCallExpression().TypeArguments != nil {
								typeRange, ok := testFramework.CallTypeArgumentListRange(ctx.SourceFile, call)
								if !ok || utils.HasCommentInSpan(ctx.Comments.All(), typeRange.Pos(), typeRange.End()) {
									return nil
								}
								fixes = append(fixes, rule.RuleFixRemoveRange(typeRange))
							}
							if assertion.Not != nil {
								ranges, ok := testFramework.RemoveAccessorEntryRanges(ctx.SourceFile, ctx.Comments.All(), assertion.Not)
								if !ok {
									return nil
								}
								for _, textRange := range ranges {
									fixes = append(fixes, rule.RuleFixRemoveRange(textRange))
								}
							} else {
								textRange, text, ok := testFramework.InsertMemberBeforeAccessor(&assertion.Matcher, "not")
								if !ok {
									return nil
								}
								fixes = append(fixes, rule.RuleFixReplaceRange(textRange, text))
							}
							return fixes
						})
					}
				},
			}
		},
	}
}
