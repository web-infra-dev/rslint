// TestPreferCodePointUpstream migrates the valid/invalid suite from upstream
// test/prefer-code-point.js 1:1, including the TypeScript wrapper cases that
// assert whether a suggestion is offered. Upstream's final node:test executes
// the transpiled output and checks that the suggested rename would change a
// string hash; that runtime assertion corresponds to the "no suggestion"
// cases below and is not reproducible as a rule diagnostic.
package prefer_code_point_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_code_point"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferCodePointUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&prefer_code_point.PreferCodePointRule,
		[]rule_tester.ValidTestCase{
			// Already using code points.
			valid(`"🦄".codePointAt(0)`),
			valid(`String.fromCodePoint(0x1f984)`),
			valid(`String.fromCodePoint`),

			// `charCodeAt` is only reported when it is actually called.
			valid(`foo.charCodeAt`),
			valid(`new foo.charCodeAt`),
			valid(`charCodeAt(0)`),
			valid(`foo.charCodeAt?.(0)`),
			valid(`foo[charCodeAt](0)`),
			valid(`foo["charCodeAt"](0)`),
			valid(`foo.notCharCodeAt(0)`),

			// The `fromCharCode` member must be a dot access on the `String` identifier.
			valid(`new String.fromCodePoint`),
			valid(`fromCodePoint(foo)`),
			valid(`String.fromCodePoint?.(foo)`),
			valid(`String?.fromCodePoint(foo)`),
			valid(`window.String.fromCodePoint(foo)`),
			valid(`String[fromCodePoint](foo)`),
			valid(`String["fromCodePoint"](foo)`),
			valid(`String.notFromCodePoint(foo)`),
			valid(`NotString.fromCodePoint(foo)`),
		},
		[]rule_tester.InvalidTestCase{
			// ---- `charCodeAt` with a suggestion ----
			invalid(`string.charCodeAt(index)`, "charCodeAt", "codePointAt", true),
			invalid(`string?.charCodeAt(index)`, "charCodeAt", "codePointAt", true),
			invalid(`(( (( string )).charCodeAt( ((index)), )))`, "charCodeAt", "codePointAt", true),
			invalid(`string.charCodeAt(index).toString(16)`, "charCodeAt", "codePointAt", true),
			invalid(`string.charCodeAt(index) > 127`, "charCodeAt", "codePointAt", true),
			invalid(`foo(string.charCodeAt(index))`, "charCodeAt", "codePointAt", true),
			invalid(`const x = string.charCodeAt(index)`, "charCodeAt", "codePointAt", true),
			invalid(`string.charCodeAt(index) === 65`, "charCodeAt", "codePointAt", true),
			invalid(`string.charCodeAt(index) || 0`, "charCodeAt", "codePointAt", true),
			invalid(`!string.charCodeAt(index)`, "charCodeAt", "codePointAt", true),

			// ---- `charCodeAt` used as a number: reported without a suggestion ----
			invalid(`hash = (((hash * 31) | 0) + string.charCodeAt(offset)) | 0`, "charCodeAt", "codePointAt", false),
			invalid(`string.charCodeAt(index) + 1`, "charCodeAt", "codePointAt", false),
			invalid(`string.charCodeAt(index) * 31`, "charCodeAt", "codePointAt", false),
			invalid(`string.charCodeAt(index) - 48`, "charCodeAt", "codePointAt", false),
			invalid(`(string.charCodeAt(index) | 0)`, "charCodeAt", "codePointAt", false),
			invalid(`string.charCodeAt(index) ^ hash`, "charCodeAt", "codePointAt", false),
			invalid(`(hash << 5) - hash + string.charCodeAt(index)`, "charCodeAt", "codePointAt", false),
			invalid(`sum += string.charCodeAt(index)`, "charCodeAt", "codePointAt", false),
			invalid(`sum += condition ? string.charCodeAt(index) : 0`, "charCodeAt", "codePointAt", false),
			invalid(`sum += string.charCodeAt(index) || 0`, "charCodeAt", "codePointAt", false),
			invalid(`string?.charCodeAt(index) + 1`, "charCodeAt", "codePointAt", false),
			invalid(`~string.charCodeAt(index)`, "charCodeAt", "codePointAt", false),
			invalid(`-string.charCodeAt(index)`, "charCodeAt", "codePointAt", false),
			invalid(`sum += condition ? 0 : string.charCodeAt(index)`, "charCodeAt", "codePointAt", false),
			// The call is the condition, not a branch, so it is not numeric context.
			invalid(`sum += string.charCodeAt(index) ? 1 : 0`, "charCodeAt", "codePointAt", true),

			// ---- `String.fromCharCode` ----
			invalid(`String.fromCharCode( code )`, "fromCharCode", "fromCodePoint", true),
			invalid(`(( (( String )).fromCharCode( ((code)), ) ))`, "fromCharCode", "fromCodePoint", true),
			invalid(`String.fromCharCode.bind(String)`, "fromCharCode", "fromCodePoint", true),
			invalid(`const x = String.fromCharCode`, "fromCharCode", "fromCodePoint", true),
			invalid(`String.fromCharCode(65, 66, 67)`, "fromCharCode", "fromCodePoint", true),

			// ---- TypeScript wrappers on the receiver ----
			invalidTS(`foo!.charCodeAt(0)`, "charCodeAt", "codePointAt", true),
			invalidTS(`(str as string).charCodeAt(0)`, "charCodeAt", "codePointAt", true),
		},
	)
}

// TestPreferCodePointUpstreamTypeScriptWrappers covers the upstream cases
// where a TypeScript expression wrapper around the result must not hide
// numeric use. A wrapper is transparent for the diagnostic but still counts as
// numeric context for deciding whether the rename suggestion is safe.
func TestPreferCodePointUpstreamTypeScriptWrappers(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&prefer_code_point.PreferCodePointRule,
		nil,
		[]rule_tester.InvalidTestCase{
			invalidTS(`text.charCodeAt(0)! ^ 0;`, "charCodeAt", "codePointAt", false),
			invalidTS(`(text.charCodeAt(0) as number) ^ 0;`, "charCodeAt", "codePointAt", false),
			invalidTS(`(text.charCodeAt(0) satisfies number) ^ 0;`, "charCodeAt", "codePointAt", false),
			invalidTS(`(<number>text.charCodeAt(0)) ^ 0;`, "charCodeAt", "codePointAt", false),
			invalidTS(`sum += (text?.charCodeAt(0)! || 0) as number;`, "charCodeAt", "codePointAt", false),
			invalidTS(`sum += ready ? (text.charCodeAt(0) as number) : 0;`, "charCodeAt", "codePointAt", false),
			invalidTS(`-(text.charCodeAt(0) as number);`, "charCodeAt", "codePointAt", false),
			// Not used as a number: the suggestion survives the wrapper.
			invalidTS(`const value = text.charCodeAt(0)!;`, "charCodeAt", "codePointAt", true),
		},
	)
}
