package no_magic_array_flat_depth_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_magic_array_flat_depth"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoMagicArrayFlatDepthCommentBoundary(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_magic_array_flat_depth.NoMagicArrayFlatDepthRule,
		[]rule_tester.ValidTestCase{
			jsValid(`array.flat(/* inside */ 2)`),
			jsValid(`array.flat((/* inside */ 2))`),
			jsValid(`array.flat(2 /* inside */)`),
			jsValid("array.flat((2 // inside\n))"),
			{Code: `array.flat<number>(/* inside */ 2)`, FileName: "file.ts"},
		},
		[]rule_tester.InvalidTestCase{
			depthInvalid(`array.flat /* outside */ (2)`),
			depthInvalid("array.flat // outside\n(2)"),
			depthInvalid(`array.flat(((2)))`),
			depthInvalid(`array.flat(2) /* outside */`),
			{
				Code:     `array.flat<number> /* outside */ (2)`,
				FileName: "file.ts",
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: messageID, Message: messageString}},
			},
		},
	)
}
