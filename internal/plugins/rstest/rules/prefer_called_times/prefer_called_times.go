package prefer_called_times

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
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
				matcher, matcherCall, ok := calledOnceMatcherEntry(analysis.ParseExpectCall(node))
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

					argumentsRange, ok := rstestUtils.MatcherArgumentListRange(ctx.SourceFile, matcherCall)
					if !ok {
						return nil
					}
					// The count is inserted at the start of the argument list
					// rather than written over it, so a comment the author put
					// between the parentheses stays where it is.
					argumentStart := argumentsRange.Pos()

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
// assertion together with the call expression that invokes that matcher.
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
func calledOnceMatcherEntry(
	parsed *rstestUtils.ParsedRstestExpectCall,
) (rstestUtils.ParsedRstestFnMemberEntry, *ast.Node, bool) {
	if parsed == nil ||
		parsed.Reason != rstestUtils.RstestExpectParseReasonNone ||
		parsed.Head == nil ||
		parsed.MatcherEntry == nil ||
		parsed.Matcher != calledOnceMatcher ||
		len(parsed.Matchers) == 0 ||
		parsed.Matchers[0].Kind != rstestUtils.RstestExpectMatcherCall {
		return rstestUtils.ParsedRstestFnMemberEntry{}, nil, false
	}

	call := rstestUtils.MatcherCall(parsed.MatcherEntry)
	if call == nil {
		return rstestUtils.ParsedRstestFnMemberEntry{}, nil, false
	}
	// `toHaveBeenCalledOnce` takes no arguments, so anything already written
	// between the parentheses means the assertion is not the shape this rule
	// rewrites.
	if arguments := call.AsCallExpression().Arguments; arguments != nil && len(arguments.Nodes) > 0 {
		return rstestUtils.ParsedRstestFnMemberEntry{}, nil, false
	}

	return *parsed.MatcherEntry, call, true
}
