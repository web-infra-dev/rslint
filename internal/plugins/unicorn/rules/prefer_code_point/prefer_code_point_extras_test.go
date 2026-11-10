// TestPreferCodePointExtras locks in branches and AST shapes that the upstream
// suite does not exercise: optional-call rejection, computed/optional
// `fromCharCode` receivers, every numeric operator and compound assignment,
// and numeric-context wrappers around optional chains.
package prefer_code_point_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_code_point"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferCodePointExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&prefer_code_point.PreferCodePointRule,
		[]rule_tester.ValidTestCase{
			// TypeScript wrappers on the receiver do not change matching.
			validTS(`(str as string).charCodeAt?.(0)`),

			// Optional calls are not `charCodeAt` calls in the upstream matcher.
			valid(`string.charCodeAt?.(index)`),
			valid(`string?.charCodeAt?.(index)`),
			valid(`foo?.charCodeAt?.(0)`),

			// Non-call uses of the method reference.
			valid(`foo.charCodeAt.call(foo, 0)`),
			valid(`const method = foo.charCodeAt`),

			// `fromCharCode` needs a non-optional dot access on the `String` identifier.
			valid(`String?.fromCharCode(code)`),
			valid(`String["fromCharCode"](code)`),
			valid(`String[fromCharCode](code)`),
			valid(`window.String.fromCharCode(code)`),
			valid(`NotString.fromCharCode(code)`),
		},
		[]rule_tester.InvalidTestCase{
			// No argument-count constraint is applied.
			invalid(`string.charCodeAt()`, "charCodeAt", "codePointAt", true),

			// Parenthesized callees are transparent.
			invalid(`(string.charCodeAt)(index)`, "charCodeAt", "codePointAt", true),

			// The `fromCharCode` member itself is reported, even when called optionally
			// or used as a value.
			invalid(`String.fromCharCode?.(code)`, "fromCharCode", "fromCodePoint", true),
			invalid(`String.fromCharCode.length`, "fromCharCode", "fromCodePoint", true),
			invalid(`function f(String) { return String.fromCharCode(1); }`, "fromCharCode", "fromCodePoint", true),

			// ---- Numeric binary operators keep the diagnostic but drop the suggestion ----
			invalid(`string.charCodeAt(0) / 2`, "charCodeAt", "codePointAt", false),
			invalid(`string.charCodeAt(0) % 2`, "charCodeAt", "codePointAt", false),
			invalid(`string.charCodeAt(0) ** 2`, "charCodeAt", "codePointAt", false),
			invalid(`string.charCodeAt(0) & mask`, "charCodeAt", "codePointAt", false),
			invalid(`string.charCodeAt(0) | mask`, "charCodeAt", "codePointAt", false),
			invalid(`string.charCodeAt(0) << 1`, "charCodeAt", "codePointAt", false),
			invalid(`string.charCodeAt(0) >> 1`, "charCodeAt", "codePointAt", false),
			invalid(`string.charCodeAt(0) >>> 1`, "charCodeAt", "codePointAt", false),
			invalid(`1 + string.charCodeAt(0)`, "charCodeAt", "codePointAt", false),

			// ---- Numeric compound assignments require the call on the right ----
			invalid(`sum -= string.charCodeAt(0)`, "charCodeAt", "codePointAt", false),
			invalid(`sum *= string.charCodeAt(0)`, "charCodeAt", "codePointAt", false),
			invalid(`sum /= string.charCodeAt(0)`, "charCodeAt", "codePointAt", false),
			invalid(`sum %= string.charCodeAt(0)`, "charCodeAt", "codePointAt", false),
			invalid(`sum **= string.charCodeAt(0)`, "charCodeAt", "codePointAt", false),
			invalid(`sum &= string.charCodeAt(0)`, "charCodeAt", "codePointAt", false),
			invalid(`sum |= string.charCodeAt(0)`, "charCodeAt", "codePointAt", false),
			invalid(`sum ^= string.charCodeAt(0)`, "charCodeAt", "codePointAt", false),
			invalid(`sum <<= string.charCodeAt(0)`, "charCodeAt", "codePointAt", false),
			invalid(`sum >>= string.charCodeAt(0)`, "charCodeAt", "codePointAt", false),
			invalid(`sum >>>= string.charCodeAt(0)`, "charCodeAt", "codePointAt", false),

			// Logical assignments are not numeric contexts.
			invalid(`sum &&= string.charCodeAt(0)`, "charCodeAt", "codePointAt", true),
			invalid(`sum ||= string.charCodeAt(0)`, "charCodeAt", "codePointAt", true),
			invalid(`sum ??= string.charCodeAt(0)`, "charCodeAt", "codePointAt", true),

			// Plain assignment is not a numeric context either.
			invalid(`x = string.charCodeAt(0)`, "charCodeAt", "codePointAt", true),

			// ---- Unary operators ----
			invalid(`+string.charCodeAt(0)`, "charCodeAt", "codePointAt", false),
			invalid(`~string.charCodeAt(0)`, "charCodeAt", "codePointAt", false),
			invalid(`typeof string.charCodeAt(0)`, "charCodeAt", "codePointAt", true),
			invalid(`void string.charCodeAt(0)`, "charCodeAt", "codePointAt", true),

			// ---- Wrappers around optional chains ----
			invalid(`(string?.charCodeAt(0)) + 1`, "charCodeAt", "codePointAt", false),
			invalid(`sum += string.charCodeAt(0) && 1`, "charCodeAt", "codePointAt", false),
			invalidTS(`string?.charCodeAt(0)! + 1`, "charCodeAt", "codePointAt", false),
			invalidTS(`string?.charCodeAt(0)!`, "charCodeAt", "codePointAt", true),
		},
	)
}
