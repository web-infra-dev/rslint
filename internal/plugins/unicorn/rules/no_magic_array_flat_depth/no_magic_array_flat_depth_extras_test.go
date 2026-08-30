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

			// ---- Static-call parity: known, safely evaluated non-arrays ----
			jsValid(`Boolean({__proto__: null}).flat(2)`),
			jsValid(`Number(1n).flat(2)`),
			jsValid(`String(1n).flat(2)`),
			jsValid(`parseInt(1n).flat(2)`),
			jsValid(`Array.isArray(1n).flat(2)`),
			jsValid(`/** @type {Array<number>} */ const value = 1; value.flat(2)`),
			{
				Code:     `array.flat<number>(/* explanation */ 2)`,
				FileName: "file.ts",
			},
			{
				Code:     `array.flat<number>(2 /* explanation */)`,
				FileName: "file.ts",
			},
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

			// ---- Static-call parity: unknown or throwing evaluations report ----
			depthInvalid(`Math.abs({__proto__: null}).flat(2)`),
			depthInvalid(`Number({__proto__: null}).flat(2)`),
			depthInvalid(`String({__proto__: null}).flat(2)`),
			depthInvalid(`parseInt({__proto__: null}, 10).flat(2)`),
			depthInvalid(`String.fromCharCode(1n).flat(2)`),
			depthInvalid(`Object.freeze({}, value).flat(2)`),
			depthInvalid(`/* global value */ Object.freeze({}, value).flat(2)`),
			depthInvalid(`/* global value */ (true ? Object.freeze({}, value) : 1).flat(2)`),
			depthInvalid(`/** @type {number} */ const value = []; value.flat(2)`),
			depthInvalid(`/** @type {number} */ const value = unknown; value.flat(2)`),
			depthInvalid(`(/** @type {number} */ ([])).flat(2)`),
			depthInvalid(`(/** @type {number} */ (unknown)).flat(2)`),
			depthInvalid(`(/** @type {number[]} */ (1)).flat(2)`),
			depthInvalid(`(/** @type {number} */ ('x')).flat(2)`),
			depthInvalid(`(/** @type {number} */ ({})).flat(2)`),
			depthInvalid("(/** @type {number} */ (`x`)).flat(2)"),
			depthInvalid(`let value = {}; value.x = 1; Boolean(value).flat(2)`),
			depthInvalid(`await using value = {}; value.flat(2)`),
			depthInvalid(`const Math = {abs() { return 1 }}; Math.abs(-2).flat(2)`),
			{
				Code:     `array.flat<number>(2)`,
				FileName: "file.ts",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: messageID,
					Message:   messageString,
				}},
			},
		},
	)
}
