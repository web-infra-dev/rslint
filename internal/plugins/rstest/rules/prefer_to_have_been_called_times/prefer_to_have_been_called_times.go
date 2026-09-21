package prefer_to_have_been_called_times

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
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
					// expect.element asserts on a browser locator, which has no mock
					// context, and expect.poll takes a callback rather than the value
					// this rule reads, so `mock.calls` can only reach it as an argument
					// that already throws at runtime.
					(parsed.Entry != rstestUtils.RstestExpectEntryCall &&
						parsed.Entry != rstestUtils.RstestExpectEntrySoft) {
					return
				}
				arguments := parsed.Head.AsCallExpression().Arguments
				if arguments == nil || len(arguments.Nodes) == 0 {
					return
				}
				receiver, mockCalls := parseMockCallsAccess(arguments.Nodes[0])
				if mockCalls == nil {
					return
				}
				for _, matcher := range parsed.Matchers {
					if testFramework.IsSubjectMutatingChaiMatcher(matcher.Name) {
						break
					}
					if matcher.Kind != rstestUtils.RstestExpectMatcherCall || matcher.Name != "toHaveLength" {
						continue
					}
					matcherCall := testFramework.InvokedAccessorCall(&matcher.Entry)
					ctx.ReportNodeWithDeferredFixes(matcher.Entry.Node, rule.RuleMessage{
						Id: "preferMatcher", Description: "Prefer `toHaveBeenCalledTimes`",
					}, func() []rule.RuleFix {
						// Rewriting the factory argument changes the subject that the
						// returned Chai assertion carries, so it must not be reused.
						if len(parsed.Matchers) != 1 ||
							matcherCall == nil ||
							parsed.Expression != matcherCall ||
							!isDiscardedAssertion(parsed.Expression) {
							return nil
						}
						if !isFixableRewrite(receiver, matcherCall, arguments.Nodes) {
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
						// expect<T>() pins the subject's type and the matcher's type
						// arguments describe toHaveLength, so both stop type-checking
						// once the subject and the matcher change.
						for _, call := range []*ast.Node{parsed.Head, matcherCall} {
							if call.AsCallExpression().TypeArguments == nil {
								continue
							}
							typeRange, ok := testFramework.CallTypeArgumentListRange(ctx.SourceFile, call)
							if !ok || utils.HasCommentInSpan(ctx.Comments.All(), typeRange.Pos(), typeRange.End()) {
								return nil
							}
							fixes = append(fixes, rule.RuleFixRemoveRange(typeRange))
						}
						return fixes
					})
				}
			},
		}
	},
}

// isFixableRewrite reports whether replacing the asserted mock.calls array with
// the mock itself preserves the assertion's result.
//
// expect() captures the calls array when it runs, while toHaveBeenCalledTimes
// reads mock.calls when the matcher runs, so anything evaluated in between that
// can reset the mock changes the count the rewritten assertion sees. The two
// matchers also disagree on non-numbers: toHaveLength compares loosely and
// toHaveBeenCalledTimes strictly, so a rewritten `toHaveLength('1')` fails and a
// rewritten `not.toHaveLength('1')` starts passing.
func isFixableRewrite(receiver, matcherCall *ast.Node, headArguments []*ast.Node) bool {
	// A bare `super` is not a value, so `expect(super)` would not parse.
	if receiver == nil || ast.SkipParentheses(receiver).Kind == ast.KindSuperKeyword {
		return false
	}
	matcherArguments := matcherCall.Arguments()
	if len(matcherArguments) != 1 {
		return false
	}
	if _, _, ok := testFramework.StaticNumericLiteral(matcherArguments[0]); !ok {
		return false
	}
	for _, argument := range headArguments[1:] {
		if !testFramework.IsSideEffectFreeLiteral(argument) {
			return false
		}
	}
	return true
}

// parseMockCallsAccess returns the receiver and the `mock` and `calls` key
// nodes of a `<value>.mock.calls` access, or nil when node is any other
// expression.
func parseMockCallsAccess(node *ast.Node) (*ast.Node, []*ast.Node) {
	receiver, calls := unwrapStaticMember(node, "calls")
	if calls == nil {
		return nil, nil
	}
	if inner, mock := unwrapStaticMember(receiver, "mock"); mock != nil {
		return inner, []*ast.Node{mock, calls}
	}
	return nil, nil
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
