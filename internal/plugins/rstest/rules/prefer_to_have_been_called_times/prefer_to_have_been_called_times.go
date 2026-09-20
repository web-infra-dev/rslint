package prefer_to_have_been_called_times

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

var PreferToHaveBeenCalledTimesRule = rule.Rule{
	Name:   "rstest/prefer-to-have-been-called-times",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		analysis := rstestUtils.GetRstestCallAnalysis(ctx)
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				parsed := analysis.ParseExpectCall(node)
				if parsed == nil ||
					parsed.Reason != rstestUtils.RstestExpectParseReasonNone ||
					parsed.Head == nil ||
					// expect.element asserts on a browser locator, which has no mock context.
					parsed.Entry == rstestUtils.RstestExpectEntryElement {
					return
				}
				arguments := parsed.Head.AsCallExpression().Arguments
				if arguments == nil || len(arguments.Nodes) == 0 {
					return
				}
				mockCalls := parseMockCallsAccess(arguments.Nodes[0])
				if mockCalls == nil {
					return
				}
				for _, matcher := range parsed.Matchers {
					if matcher.Kind != rstestUtils.RstestExpectMatcherCall || matcher.Name != "toHaveLength" {
						continue
					}
					ctx.ReportNodeWithDeferredFixes(matcher.Entry.Node, rule.RuleMessage{
						Id: "preferMatcher", Description: "Prefer `toHaveBeenCalledTimes`",
					}, func() []rule.RuleFix {
						// Rewriting the factory argument changes the subject that the
						// returned Chai assertion carries, so it must not be reused.
						if parsed.Entry == rstestUtils.RstestExpectEntryPoll ||
							len(parsed.Matchers) != 1 ||
							parsed.Expression != testFramework.InvokedAccessorCall(&matcher.Entry) ||
							!isDiscardedAssertion(parsed.Expression) {
							return nil
						}
						nameRange, name, ok := testFramework.AccessorReplacement(ctx.SourceFile, matcher.Entry.Node, "toHaveBeenCalledTimes")
						if !ok {
							return nil
						}
						fixes := []rule.RuleFix{rule.RuleFixReplaceRange(nameRange, name)}
						for _, key := range mockCalls {
							ranges, ok := testFramework.RemoveAccessorEntryRanges(ctx.SourceFile, ctx.Comments.All(), &testFramework.MemberEntry{Node: key})
							if !ok {
								return nil
							}
							for _, textRange := range ranges {
								fixes = append(fixes, rule.RuleFixRemoveRange(textRange))
							}
						}
						return fixes
					})
				}
			},
		}
	},
}

// parseMockCallsAccess returns the `mock` and `calls` key nodes of a
// `<value>.mock.calls` access, or nil when node is any other expression.
func parseMockCallsAccess(node *ast.Node) []*ast.Node {
	receiver, calls := unwrapStaticMember(node, "calls")
	if calls == nil {
		return nil
	}
	if _, mock := unwrapStaticMember(receiver, "mock"); mock != nil {
		return []*ast.Node{mock, calls}
	}
	return nil
}

// unwrapStaticMember returns the receiver and key node of a non-optional
// `.name` or `["name"]` access. An optional link is rejected because removing
// the accessor would drop the guard that produced the asserted value.
func unwrapStaticMember(node *ast.Node, name string) (*ast.Node, *ast.Node) {
	if node == nil {
		return nil, nil
	}
	node = ast.SkipParentheses(node)
	if node == nil || ast.IsOptionalChain(node) {
		return nil, nil
	}
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		access := node.AsPropertyAccessExpression()
		key := access.Name()
		if key == nil || key.Kind != ast.KindIdentifier || key.AsIdentifier().Text != name {
			return nil, nil
		}
		return access.Expression, key
	case ast.KindElementAccessExpression:
		access := node.AsElementAccessExpression()
		key := ast.SkipParentheses(access.ArgumentExpression)
		if !isStaticKey(key, name) {
			return nil, nil
		}
		return access.Expression, key
	default:
		return nil, nil
	}
}

// isStaticKey accepts only literal bracket keys. An identifier key such as the
// `mock` in `fn[mock]` is a variable read, not the member name it spells.
func isStaticKey(node *ast.Node, name string) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindStringLiteral:
		return node.AsStringLiteral().Text == name
	case ast.KindNoSubstitutionTemplateLiteral:
		return node.AsNoSubstitutionTemplateLiteral().Text == name
	default:
		return false
	}
}

func isDiscardedAssertion(node *ast.Node) bool {
	for parent := node.Parent; parent != nil; parent = parent.Parent {
		switch parent.Kind {
		case ast.KindParenthesizedExpression, ast.KindAwaitExpression:
			continue
		case ast.KindExpressionStatement:
			return true
		default:
			return false
		}
	}
	return false
}
