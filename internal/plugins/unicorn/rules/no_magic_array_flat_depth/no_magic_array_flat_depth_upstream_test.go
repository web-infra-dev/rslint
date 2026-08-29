// TestNoMagicArrayFlatDepthUpstream mirrors Unicorn v73.0.0's complete
// test/no-magic-array-flat-depth.js valid/invalid suite. rslint-only regression
// coverage lives in the neighboring extras files.
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
			{
				Code:     `function f(foo: {flat(depth: number): void}) { foo.flat(2); }`,
				FileName: "file.ts",
			},
			jsValid(`array.flat(1)`),
			jsValid(`array.flat(1.0)`),
			jsValid(`array.flat(0x01)`),
			jsValid(`array.flat(unknown)`),
			jsValid(`array.flat(Number.POSITIVE_INFINITY)`),
			jsValid(`array.flat(Infinity)`),
			jsValid(`array.flat(/* explanation */2)`),
			jsValid(`array.flat(2/* explanation */)`),
			jsValid(`array.flat()`),
			jsValid(`array.flat(2, extraArgument)`),
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
			{
				Code:     `function f(foo: number[][]) { foo.flat(2); }`,
				FileName: "file.ts",
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: messageID, Message: messageString}},
			},
		},
	)
}
