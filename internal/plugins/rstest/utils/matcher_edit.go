package utils

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

// Shared primitives for rules that rewrite a matcher in place: locating the
// call the matcher name is the callee of, reading through the type-assertion
// wrappers ESTree does not have, and finding the span the matcher's arguments
// occupy. Every rewrite rule in this plugin needs at least two of the three, so
// they live here rather than being copied per rule.

// MatcherCall returns the call expression that directly invokes entry's
// accessor, or nil when the accessor is not called.
//
// MemberEntry.Call cannot be used for this. GetMemberEntries marks the last
// entry of a callee chain as invoked, so each enclosing call overwrites the
// field: for `expect(fn).toBe(true)()` the matcher entry ends up carrying the
// outer call, whose parentheses hold no matcher argument list at all. Editing
// that call's arguments would touch the wrong text. Walking up from the
// accessor instead always lands on the call the matcher name is the callee of,
// and stops at it. Parentheses around the callee are transparent, matching how
// the chain was parsed in the first place.
func MatcherCall(entry *ParsedRstestFnMemberEntry) *ast.Node {
	return testFramework.InvokedAccessorCall(entry)
}

// FollowTypeAssertionChain unwraps `x as T` and `<T>x` wrappers, along with the
// parentheses that ESTree does not represent at all.
//
// This is the Go form of the `followTypeAssertionChain` the upstream plugins
// apply to a matcher argument before inspecting it. `satisfies` and the
// non-null assertion are deliberately not unwrapped: upstream stops at both,
// so a value written `1 satisfies number` or `x!` keeps whatever wrapper it
// was given.
func FollowTypeAssertionChain(node *ast.Node) *ast.Node {
	return testFramework.FollowTypeAssertionChain(node)
}

// MatcherArgumentListRange returns the span between the matcher call's
// parentheses — the text its arguments occupy, empty for a call written `()`.
//
// Removing an argument node on its own leaves behind whatever was written
// around it: `toBe(true,)` keeps its comma and becomes `toBeTruthy(,)`, a
// syntax error. Taking the whole span between the parentheses removes the
// trailing comma and the whitespace of a multi-line argument list along with
// the argument. The same range read from its start is where a rule that adds
// an argument to an empty list inserts one.
//
// The opening parenthesis is scanned for rather than assumed to follow the
// matcher name, because `x['toBe'](true)`, `x.toBe /* c */ (true)` and a call
// split across lines all put other characters in between; the scan starts
// after the callee so that the parentheses of the receiver — `expect(value)` —
// are never mistaken for the matcher's own. A call expression always ends on
// its own closing parenthesis, so the last token of the call is the other end
// of the span.
func MatcherArgumentListRange(sourceFile *ast.SourceFile, call *ast.Node) (core.TextRange, bool) {
	return testFramework.CallArgumentListRange(sourceFile, call)
}
