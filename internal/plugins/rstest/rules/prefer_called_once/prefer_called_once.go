package prefer_called_once

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

// calledOnceMatcher is the only name this rule ever writes.
//
// NOTE: Unlike ESLint, the replacement is not derived from the matcher that was
// written. Upstream computes it with matcherName.replace('Times', 'Once'), so
// `expect(fn).toBeCalledTimes(1)` becomes `expect(fn).toBeCalledOnce()`.
// Rstest's assertion library has no `toBeCalledOnce`: @vitest/expect@4.1.10
// declares `toHaveBeenCalledTimes` and its `toBeCalledTimes` alias, and
// `toHaveBeenCalledOnce`, but neither dist/index.d.ts nor dist/index.js
// mentions `toBeCalledOnce` anywhere. Deriving the name would turn a passing
// assertion into a call to an undefined matcher, so both spellings of the
// count assertion are rewritten to the one name that exists.
const calledOnceMatcher = "toHaveBeenCalledOnce"

// calledTimesMatchers are the count assertions that can be spelled with
// `toHaveBeenCalledOnce()` instead. Both exist in @vitest/expect@4.1.10, the
// second as a deprecated alias of the first.
var calledTimesMatchers = map[string]bool{
	"toHaveBeenCalledTimes": true,
	"toBeCalledTimes":       true,
}

var preferCalledOnceMessage = rule.RuleMessage{
	Id:          "preferCalledOnce",
	Description: "Prefer " + calledOnceMatcher + "()",
}

var PreferCalledOnceRule = rule.Rule{
	Name:   "rstest/prefer-called-once",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		analysis := rstestUtils.GetRstestCallAnalysis(ctx)
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				matcher, matcherCall, ok := calledOnceMatcherEntry(analysis.ParseExpectCall(node))
				if !ok {
					return
				}

				ctx.ReportNodeWithDeferredFixes(matcher.Node, preferCalledOnceMessage, func() []rule.RuleFix {
					nameRange, nameText, ok := testFramework.AccessorReplacement(
						ctx.SourceFile,
						matcher.Node,
						calledOnceMatcher,
					)
					if !ok {
						return nil
					}

					argumentsRange, ok := argumentListRange(ctx.SourceFile, matcherCall)
					if !ok {
						return nil
					}
					// `toHaveBeenCalledOnce` takes no arguments, so the count
					// and everything written alongside it goes. A comment
					// between the parentheses would go with it, so the rename
					// is withheld too and the diagnostic stands alone rather
					// than the fix silently dropping what the author wrote.
					if utils.HasCommentInSpan(ctx.Comments.All(), argumentsRange.Pos(), argumentsRange.End()) {
						return nil
					}

					return []rule.RuleFix{
						rule.RuleFixReplaceRange(nameRange, nameText),
						rule.RuleFixRemoveRange(argumentsRange),
					}
				})
			},
		}
	},
}

// calledOnceMatcherEntry returns the matcher entry of a call-count assertion
// that asserts exactly one call, together with the call expression that invokes
// that matcher.
//
// The matcher must be invoked: `expect(fn).toHaveBeenCalledTimes` asserts
// nothing, and Chai's call-count assertions are property getters whose
// `toHaveBeenCalledOnce()` equivalent is `calledOnce` rather than a matcher
// call. Both are excluded by requiring RstestExpectMatcherCall.
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
		!calledTimesMatchers[parsed.Matcher] ||
		len(parsed.Matchers) == 0 ||
		parsed.Matchers[0].Kind != rstestUtils.RstestExpectMatcherCall {
		return rstestUtils.ParsedRstestFnMemberEntry{}, nil, false
	}

	call := matcherCall(parsed.MatcherEntry)
	if call == nil {
		return rstestUtils.ParsedRstestFnMemberEntry{}, nil, false
	}
	arguments := call.AsCallExpression().Arguments
	if arguments == nil || len(arguments.Nodes) != 1 || !isOneLiteral(arguments.Nodes[0]) {
		return rstestUtils.ParsedRstestFnMemberEntry{}, nil, false
	}

	return *parsed.MatcherEntry, call, true
}

// isOneLiteral reports whether node is a numeric literal whose value is 1.
//
// Type assertions are transparent, matching upstream's followTypeAssertionChain,
// and so are parentheses, which ESTree does not represent at all and upstream
// therefore never sees. `satisfies` and the non-null assertion are not: upstream
// stops at both, so `toHaveBeenCalledTimes(1 satisfies number)` keeps its count.
//
// A BigInt `1n` is a different literal kind carrying a BigInt value, and `+1` is
// an operator applied to a literal rather than a literal; upstream reports
// neither, because ESTree's Literal `value` is `1n` for the first and there is no
// Literal at all for the second.
func isOneLiteral(node *ast.Node) bool {
	node = followTypeAssertionChain(node)
	return node != nil &&
		node.Kind == ast.KindNumericLiteral &&
		utils.NormalizeNumericLiteral(node.Text()) == "1"
}

// followTypeAssertionChain unwraps `x as T` and `<T>x` wrappers, along with the
// parentheses that ESTree drops before upstream's own unwrapping runs.
func followTypeAssertionChain(node *ast.Node) *ast.Node {
	for node != nil {
		node = ast.SkipParentheses(node)
		switch node.Kind {
		case ast.KindAsExpression:
			node = node.AsAsExpression().Expression
		case ast.KindTypeAssertionExpression:
			node = node.AsTypeAssertion().Expression
		default:
			return node
		}
	}
	return nil
}

// matcherCall returns the call expression that directly invokes entry's
// accessor.
//
// MemberEntry.Call cannot be used for this. GetMemberEntries marks the last
// entry of a callee chain as invoked, so each enclosing call overwrites the
// field: for `expect(fn).toHaveBeenCalledTimes(1)()` the matcher entry ends up
// carrying the outer call, whose parentheses hold no matcher argument list at
// all. Removing that call's arguments would delete the wrong text. Walking up
// from the accessor instead always lands on the call the matcher name is the
// callee of, and stops at it. Parentheses around the callee are transparent,
// matching how the chain was parsed in the first place.
func matcherCall(entry *rstestUtils.ParsedRstestFnMemberEntry) *ast.Node {
	_, accessor := testFramework.AccessorReceiverAndParent(entry)
	if accessor == nil {
		return nil
	}

	child := accessor
	parent := accessor.Parent
	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		child = parent
		parent = parent.Parent
	}
	if parent == nil ||
		parent.Kind != ast.KindCallExpression ||
		parent.AsCallExpression().Expression != child {
		return nil
	}
	return parent
}

// argumentListRange returns the span between the matcher call's parentheses,
// which is the text the count occupies.
//
// Upstream removes the argument node itself, which leaves behind whatever was
// written around it: `toBeCalledTimes(1,)` keeps its comma and becomes
// `toBeCalledOnce(,)`, a syntax error. Taking the whole span between the
// parentheses removes the trailing comma and the whitespace of a multi-line
// argument list along with the count.
//
// The opening parenthesis is scanned for rather than assumed to follow the
// matcher name, because `x['toBeCalledTimes'](1)`,
// `x.toBeCalledTimes /* c */ (1)` and a call split across lines all put other
// characters in between; the scan starts after the callee so that the
// parentheses of the receiver — `expect(fn)` — are never mistaken for the
// matcher's own. A call expression always ends on its own closing parenthesis,
// so the last token of the call is the other end of the span.
func argumentListRange(sourceFile *ast.SourceFile, call *ast.Node) (core.TextRange, bool) {
	callee := call.AsCallExpression().Expression
	if callee == nil {
		return core.TextRange{}, false
	}

	start := callee.End()
	// A type argument list may hold parentheses of its own, as in
	// `toHaveBeenCalledTimes<(a: string) => void>(1)`.
	if typeArguments := call.AsCallExpression().TypeArguments; typeArguments != nil {
		start = max(start, typeArguments.End())
	}

	tokens := utils.TokensOfNode(sourceFile, call)
	if len(tokens) == 0 {
		return core.TextRange{}, false
	}
	closeParen := tokens[len(tokens)-1]
	if closeParen.Kind != ast.KindCloseParenToken {
		return core.TextRange{}, false
	}

	for _, token := range tokens {
		if token.Start < start {
			continue
		}
		if token.Kind == ast.KindOpenParenToken {
			if token.End > closeParen.Start {
				return core.TextRange{}, false
			}
			return core.NewTextRange(token.End, closeParen.Start), true
		}
	}
	return core.TextRange{}, false
}
