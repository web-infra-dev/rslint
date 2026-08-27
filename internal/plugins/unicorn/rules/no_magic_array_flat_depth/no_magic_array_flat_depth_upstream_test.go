// TestNoMagicArrayFlatDepthUpstream migrates the full valid/invalid suite from
// upstream test/no-magic-array-flat-depth.js 1:1. Position assertions cover
// line/column for the depth argument reported in every invalid case.
// rslint-specific lock-in cases live in no_magic_array_flat_depth_extras_test.go.
package no_magic_array_flat_depth_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_magic_array_flat_depth"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoMagicArrayFlatDepthUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_magic_array_flat_depth.NoMagicArrayFlatDepthRule,
		[]rule_tester.ValidTestCase{
			// ---- Known non-array receiver (type information) ----
			{
				Code:     `function f(foo: {flat(depth: number): void}) { foo.flat(2); }`,
				FileName: "file.ts",
			},
			// `shouldSkipKnownNonArrayReceiver` skip case — a typed `Set` is
			// known not to be an array, so the rule legitimately skips it.
			{
				Code:     `function f(foo: Set<number>) { foo.flat(2); }`,
				FileName: "file.ts",
			},

			// ---- depth is 1 (the default) ----
			jsValid(`array.flat(1)`),
			jsValid(`array.flat(1.0)`),
			jsValid(`array.flat(0x01)`),

			// ---- depth is not a numeric literal ----
			jsValid(`array.flat(unknown)`),
			jsValid(`array.flat(Number.POSITIVE_INFINITY)`),
			jsValid(`array.flat(Infinity)`),

			// ---- comments around the depth explain the magic number ----
			jsValid(`array.flat(/* explanation */2)`),
			jsValid(`array.flat(2/* explanation */)`),

			// ---- no argument or wrong number of arguments ----
			jsValid(`array.flat()`),
			jsValid(`array.flat(2, extraArgument)`),

			// ---- not a CallExpression of `.flat` ----
			jsValid(`new array.flat(2)`),
			jsValid(`array.flat?.(2)`),
			jsValid(`array.notFlat(2)`),
			jsValid(`flat(2)`),
		},
		[]rule_tester.InvalidTestCase{
			depthInvalid(`array.flat(2)`),
			depthInvalid(`array?.flat(2)`),
			depthInvalid(`array.flat(99,)`),
			depthInvalid(`array.flat(0b10,)`),

			// A receiver that is known to be an array must still be reported.
			{
				Code:     `function f(foo: number[][]) { foo.flat(2); }`,
				FileName: "file.ts",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: messageID,
					Message:   messageString,
				}},
			},

			// `shouldSkipKnownNonArrayReceiver` regression cases — these
			// receivers are directly visible (mismatch is at the call site)
			// so the rule MUST still report them, even though
			// `isKnownNonIndexedCollection` would otherwise skip them.
			{
				Code:     `({flat(){}}).flat(2)`,
				FileName: "file.mjs",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: messageID,
					Message:   messageString,
				}},
			},
			{
				Code:     `"x".flat(2)`,
				FileName: "file.mjs",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: messageID,
					Message:   messageString,
				}},
			},
			// Typed array — shares most of Array's method surface, but
			// `flat()` isn't on its prototype, so reporting the broken call
			// is the right call.
			{
				Code:     `new Uint8Array().flat(2)`,
				FileName: "file.mjs",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: messageID,
					Message:   messageString,
				}},
			},
		},
	)
}
