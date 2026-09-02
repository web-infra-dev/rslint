package prefer_strict_boolean_matchers

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

// strictMatcher is the one matcher this rule ever writes; which literal it is
// given is what distinguishes the two diagnostics.
const strictMatcher = "toBe"

// coercingMatchers maps each truthiness matcher to the literal that asserts the
// same thing strictly, and to the diagnostic that says so.
var coercingMatchers = map[string]struct {
	literal string
	message rule.RuleMessage
}{
	"toBeTruthy": {
		literal: "true",
		message: rule.RuleMessage{
			Id:          "preferToBeTrue",
			Description: "Prefer using `toBe(true)` to test value is `true`",
		},
	},
	"toBeFalsy": {
		literal: "false",
		message: rule.RuleMessage{
			Id:          "preferToBeFalse",
			Description: "Prefer using `toBe(false)` to test value is `false`",
		},
	},
}

var PreferStrictBooleanMatchersRule = rule.Rule{
	Name:   "rstest/prefer-strict-boolean-matchers",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		analysis := rstestUtils.GetRstestCallAnalysis(ctx)
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				matcher, matcherCall, matcherName, ok := coercingMatcherEntry(analysis.ParseExpectCall(node))
				if !ok {
					return
				}
				replacement := coercingMatchers[matcherName]

				ctx.ReportNodeWithDeferredFixes(matcher.Node, replacement.message, func() []rule.RuleFix {
					nameRange, nameText, ok := testFramework.AccessorReplacement(
						ctx.SourceFile,
						matcher.Node,
						strictMatcher,
					)
					if !ok {
						return nil
					}

					argumentsRange, ok := rstestUtils.MatcherArgumentListRange(ctx.SourceFile, matcherCall)
					if !ok {
						return nil
					}
					// The literal is inserted at the start of the argument list
					// rather than written over it, so a comment the author put
					// between the parentheses stays where it is.
					argumentStart := argumentsRange.Pos()

					return []rule.RuleFix{
						rule.RuleFixReplaceRange(nameRange, nameText),
						rule.RuleFixReplaceRange(core.NewTextRange(argumentStart, argumentStart), replacement.literal),
					}
				})
			},
		}
	},
}

// coercingMatcherEntry returns the matcher entry of a truthiness assertion
// together with the call expression that invokes that matcher.
//
// The matcher must be invoked: `expect(value).toBeTruthy` asserts nothing, and
// Chai's truthiness assertions are property getters (`expect(value).to.be.ok`)
// whose strict form is `equal(true)` rather than `toBe(true)`. Both are
// excluded by requiring RstestExpectMatcherCall.
//
// An assertion factory must also have run. Head is nil for
// `expect.toBeTruthy()` and friends, where the author never passed a value to
// assert on; the assertion asserts nothing whichever matcher it names, so
// rewriting the matcher would only make a broken assertion look right.
func coercingMatcherEntry(
	parsed *rstestUtils.ParsedRstestExpectCall,
) (rstestUtils.ParsedRstestFnMemberEntry, *ast.Node, string, bool) {
	if parsed == nil ||
		parsed.Reason != rstestUtils.RstestExpectParseReasonNone ||
		parsed.Head == nil ||
		parsed.MatcherEntry == nil ||
		len(parsed.Matchers) == 0 ||
		parsed.Matchers[0].Kind != rstestUtils.RstestExpectMatcherCall {
		return rstestUtils.ParsedRstestFnMemberEntry{}, nil, "", false
	}
	if _, ok := coercingMatchers[parsed.Matcher]; !ok {
		return rstestUtils.ParsedRstestFnMemberEntry{}, nil, "", false
	}

	// A computed identifier key names its matcher at runtime, so `toBeTruthy`
	// here is the name of a variable rather than of a matcher. Reporting it
	// would flag whatever assertion that variable holds.
	if parsed.MatcherEntry.Node == nil || testFramework.IsComputedIdentifierAccessor(parsed.MatcherEntry.Node) {
		return rstestUtils.ParsedRstestFnMemberEntry{}, nil, "", false
	}

	call := rstestUtils.MatcherCall(parsed.MatcherEntry)
	if call == nil {
		return rstestUtils.ParsedRstestFnMemberEntry{}, nil, "", false
	}
	// `toBeTruthy` and `toBeFalsy` take no arguments. Upstream inserts the
	// literal without looking, which welds it onto whatever was already there:
	// `toBeTruthy(flag)` becomes a `toBe` call on one run-together identifier,
	// code that still compiles and asserts something else entirely. Anything
	// between the parentheses means this is not the shape the rule rewrites.
	if arguments := call.AsCallExpression().Arguments; arguments != nil && len(arguments.Nodes) > 0 {
		return rstestUtils.ParsedRstestFnMemberEntry{}, nil, "", false
	}

	return *parsed.MatcherEntry, call, parsed.Matcher, true
}
