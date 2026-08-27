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
		},
	)
}
