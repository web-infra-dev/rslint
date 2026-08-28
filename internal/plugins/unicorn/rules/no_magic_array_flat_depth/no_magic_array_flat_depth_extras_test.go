// TestNoMagicArrayFlatDepthExtras locks in branches and edge shapes that the
// upstream test suite doesn't exercise. Each case carries an inline comment
// pointing at the specific branch it covers, so future refactors can't
// silently regress them without breaking a named lock-in.
package no_magic_array_flat_depth_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_magic_array_flat_depth"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// tsValid is the in-package alias for jsValid — the rule is mostly
// syntactic and the TS-only valid case is exercised in upstream.
func tsValid(code string) rule_tester.ValidTestCase {
	return rule_tester.ValidTestCase{Code: code, FileName: "file.mjs"}
}

// tsDepthInvalid mirrors the upstream helper, but routes through the same
// depthInvalid builder.
func tsDepthInvalid(code string) rule_tester.InvalidTestCase {
	return depthInvalid(code)
}

func TestNoMagicArrayFlatDepthExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_magic_array_flat_depth.NoMagicArrayFlatDepthRule,
		[]rule_tester.ValidTestCase{
			// ---- Branch: not a CallExpression ----
			tsValid(`2;`),
			tsValid(`let x = 2;`),

			// ---- Branch: CallExpression with a non-`flat` method name ----
			tsValid(`array.map(2)`),
			tsValid(`array.forEach(2)`),
			tsValid(`array.slice(2)`),
			tsValid(`array.flatMap(2)`),
			// Computed member access disqualifies the call.
			tsValid(`array["flat"](2)`),
			tsValid(`arr()["flat"](2)`),
			// Bare `flat(2)` is not a dot-method call.
			tsValid(`flat(2)`),

			// ---- Branch: special non-numeric tokens that look numeric-ish ----
			tsValid(`array.flat(NaN)`),
			tsValid(`array.flat(Number.MAX_SAFE_INTEGER)`),
			tsValid(`array.flat(BigInt(1))`),
			tsValid(`array.flat(depthVar)`),
			// BigInt literal — different kind in tsgo.
			tsValid(`array.flat(2n)`),

			// ---- Branch: depth is `1` in any normalized shape ----
			tsValid(`array.flat(1.0)`),
			tsValid(`array.flat(0x01)`),
			tsValid(`array.flat((1))`),

			// ---- Branch: optional chain on the call disqualifies it ----
			tsValid(`array.flat?.(2)`),
		},
		[]rule_tester.InvalidTestCase{
			// ---- Branch lock-in: every non-1 numeric literal shape ----
			tsDepthInvalid(`array.flat(0)`),
			tsDepthInvalid(`array.flat(2)`),
			tsDepthInvalid(`array.flat(99)`),
			tsDepthInvalid(`array.flat(0xff)`),
			tsDepthInvalid(`array.flat(0b10)`),
			tsDepthInvalid(`array.flat(0o10)`),
			tsDepthInvalid(`array.flat(1e2)`),
			tsDepthInvalid(`array.flat(1.5)`),
			tsDepthInvalid(`array.flat(0b10_0000)`), // numeric separators

			// ---- Branch lock-in: parenthesized depth ----
			// `(2)` is a ParenthesizedExpression wrapping a NumericLiteral;
			// MatchDotMethodCall returns the NumericLiteral as the argument.
			{
				Code:     `array.flat((2))`,
				FileName: "file.mjs",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: messageID,
					Message:   messageString,
				}},
			},

			// ---- Branch lock-in: receiver in unusual positions ----
			tsDepthInvalid(`arr.map(x).flat(2)`),
			// `arr()?.flat(2)` — optional member on a parenthesized call
			// still matches the dot-method-call shape.
			{
				Code:     `arr()?.flat(2)`,
				FileName: "file.mjs",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: messageID,
					Message:   messageString,
				}},
			},
		},
	)
}
