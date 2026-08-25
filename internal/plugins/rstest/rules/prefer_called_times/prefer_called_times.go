package prefer_called_times

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

// calledOnceMatcher is the only matcher this rule rewrites.
//
// Upstream also matches `toBeCalledOnce` and derives the replacement with
// matcherName.replace('Once', 'Times'). Rstest's assertion library does not
// expose `toBeCalledOnce` at all (@vitest/expect@4.1.10 dist/index.d.ts has
// `toHaveBeenCalledOnce` at line 650 and no `toBeCalledOnce` anywhere), so
// there is exactly one source name and exactly one target name, and the
// derivation collapses into two constants.
const (
	calledOnceMatcher  = "toHaveBeenCalledOnce"
	calledTimesMatcher = "toHaveBeenCalledTimes"
)

var preferCalledTimesMessage = rule.RuleMessage{
	Id:          "preferCalledTimes",
	Description: "Prefer " + calledTimesMatcher + "(1)",
}

var PreferCalledTimesRule = rule.Rule{
	Name:   "rstest/prefer-called-times",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		analysis := rstestUtils.GetRstestCallAnalysis(ctx)
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				matcher, ok := calledOnceMatcherEntry(analysis.ParseExpectCall(node))
				if !ok {
					return
				}

				ctx.ReportNodeWithDeferredFixes(matcher.Node, preferCalledTimesMessage, func() []rule.RuleFix {
					nameRange, nameText, ok := testFramework.AccessorReplacement(
						ctx.SourceFile,
						matcher.Node,
						calledTimesMatcher,
					)
					if !ok {
						return nil
					}

					argumentStart, ok := argumentListStart(ctx.SourceFile, matcher.Call)
					if !ok {
						return nil
					}

					return []rule.RuleFix{
						rule.RuleFixReplaceRange(nameRange, nameText),
						rule.RuleFixReplaceRange(core.NewTextRange(argumentStart, argumentStart), "1"),
					}
				})
			},
		}
	},
}

// calledOnceMatcherEntry returns the matcher entry of a `toHaveBeenCalledOnce()`
// assertion.
//
// The matcher must be invoked: `expect(fn).toHaveBeenCalledOnce` asserts
// nothing, and Chai's `calledOnce` is a property getter whose call-count
// equivalent is `callCount(1)` rather than `toHaveBeenCalledTimes(1)`. Both are
// excluded by requiring RstestExpectMatcherCall.
//
// An assertion factory must also have run. Head is nil for `expect.toBe(1)`
// and friends, where the author never passed a value to assert on; the
// assertion is broken whichever call-count matcher it names, so rewriting the
// matcher would only make a wrong assertion look right.
func calledOnceMatcherEntry(parsed *rstestUtils.ParsedRstestExpectCall) (rstestUtils.ParsedRstestFnMemberEntry, bool) {
	if parsed == nil ||
		parsed.Reason != rstestUtils.RstestExpectParseReasonNone ||
		parsed.Head == nil ||
		parsed.MatcherEntry == nil ||
		parsed.Matcher != calledOnceMatcher ||
		len(parsed.Matchers) == 0 ||
		parsed.Matchers[0].Kind != rstestUtils.RstestExpectMatcherCall {
		return rstestUtils.ParsedRstestFnMemberEntry{}, false
	}

	call := parsed.MatcherEntry.Call
	if call == nil || call.Kind != ast.KindCallExpression {
		return rstestUtils.ParsedRstestFnMemberEntry{}, false
	}
	// `toHaveBeenCalledOnce` takes no arguments, so anything already written
	// between the parentheses means the assertion is not the shape this rule
	// rewrites.
	if arguments := call.AsCallExpression().Arguments; arguments != nil && len(arguments.Nodes) > 0 {
		return rstestUtils.ParsedRstestFnMemberEntry{}, false
	}

	return *parsed.MatcherEntry, true
}

// argumentListStart returns the offset just after the matcher call's opening
// parenthesis, which is where the `1` argument is inserted.
//
// Upstream computes this as `matcher.range[1] + 1`, which assumes the `(`
// immediately follows the matcher name. That holds for `x.toBeCalledOnce()`
// only: `x['toBeCalledOnce']()`, `x.toBeCalledOnce /* c */ ()` and a call split
// across lines all put other characters in between. The parenthesis is scanned
// for instead, starting after the callee so that the parentheses of the
// receiver — `expect(fn)` — are never mistaken for the matcher's own.
func argumentListStart(sourceFile *ast.SourceFile, call *ast.Node) (int, bool) {
	callee := call.AsCallExpression().Expression
	if callee == nil {
		return 0, false
	}

	start := callee.End()
	// A type argument list may hold parentheses of its own, as in
	// `toHaveBeenCalledOnce<(a: string) => void>()`.
	if typeArguments := call.AsCallExpression().TypeArguments; typeArguments != nil {
		start = max(start, typeArguments.End())
	}

	for _, token := range utils.TokensOfNode(sourceFile, call) {
		if token.Start < start {
			continue
		}
		if token.Kind == ast.KindOpenParenToken {
			return token.End, true
		}
	}
	return 0, false
}
