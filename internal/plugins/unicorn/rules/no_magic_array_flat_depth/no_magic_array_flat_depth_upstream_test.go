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

			// Arrow / class expressions are deliberately NOT in the
			// `directlyReportableReceiverTypes` set; they fall through to
			// `isKnownNonIndexedCollection`, which classifies both as
			// non-array expressions and skips them. Lock in every shape
			// upstream marks as "skip" so future refactors don't promote
			// them to the reportable set.
			jsValid(`(() => {}).flat(2)`),
			jsValid(`(async () => {}).flat(2)`),
			jsValid(`(class {}).flat(2)`),
			jsValid(`(class extends Array {}).flat(2)`),
			jsValid(`(class { flat() {} }).flat(2)`),
			jsValid(`(() => {})?.flat(2)`),
			jsValid(`(class {})?.flat(2)`),
			jsValid(`(((() => {}))).flat(2)`),
			jsValid(`(((class {}))).flat(2)`),

			// Constructor calls that the static evaluator resolves to a
			// known non-array value — upstream (and rslint) legitimately
			// skip these.
			jsValid(`Number(1).flat(2)`),
			jsValid(`String("x").flat(2)`),
			jsValid(`Boolean(0).flat(2)`),
			jsValid(`BigInt(1).flat(2)`),
			// Parenthesized form of a coercion constructor must still
			// be skipped.
			jsValid(`(Number(1)).flat(2)`),

			// `String.fromCharCode` is folded by the shared static
			// evaluator to a string value, so the rule skips it.
			jsValid(`String.fromCharCode(65).flat(2)`),
			jsValid(`Math.abs(-2).flat(2)`),
			jsValid(`Math.max(1, 2).flat(2)`),
			jsValid(`Array.isArray([]).flat(2)`),
			jsValid(`parseInt("2", 10).flat(2)`),
			jsValid(`parseFloat("2").flat(2)`),
			jsValid(`Object().flat(2)`),
			jsValid(`const value = 1; value.flat(2);`),

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
			depthInvalid(`Array.of(1).flat(2)`),
			depthInvalid(`Object.freeze([]).flat(2)`),
			depthInvalid(`Number(value).flat(2)`),
			depthInvalid(`String(value).flat(2)`),
			depthInvalid(`BigInt().flat(2)`),
			depthInvalid(`(flag ? Math.PI : 1).flat(2)`),
			{
				Code:     `array.flat /* reason */ (2)`,
				FileName: "file.mjs",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: messageID,
					Message:   messageString,
				}},
			},

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

			// Source-only unknown receivers — upstream's `getStaticType`
			// returns "unknown" for bare identifiers and member expressions
			// without an evaluable static value, so the rule reports. These
			// exercise the source-only branch of
			// `ShouldSkipKnownNonArrayReceiver`, which would otherwise lean
			// on the type checker and over-classify globals.
			{
				Code:     `undefined.flat(2)`,
				FileName: "file.mjs",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: messageID,
					Message:   messageString,
				}},
			},
			{
				Code:     `NaN.flat(2)`,
				FileName: "file.mjs",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: messageID,
					Message:   messageString,
				}},
			},
			{
				Code:     `Infinity.flat(2)`,
				FileName: "file.mjs",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: messageID,
					Message:   messageString,
				}},
			},
			{
				Code:     `Number.NaN.flat(2)`,
				FileName: "file.mjs",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: messageID,
					Message:   messageString,
				}},
			},
			{
				Code:     `Math.PI.flat(2)`,
				FileName: "file.mjs",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: messageID,
					Message:   messageString,
				}},
			},
			{
				Code:     `Symbol.iterator.flat(2)`,
				FileName: "file.mjs",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: messageID,
					Message:   messageString,
				}},
			},
			{
				Code:     `Symbol().flat(2)`,
				FileName: "file.mjs",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: messageID,
					Message:   messageString,
				}},
			},

			// Local bindings where upstream keeps the source-only
			// "unknown" classification: `let` / `var` / destructuring
			// don't follow the `const IDENT = EXPR` resolution path, and
			// `const IDENT = Math.PI` is still unknown because Math.PI
			// itself is a MemberExpression. Lock in that rslint matches.
			{
				Code:     `let value = 1; value.flat(2);`,
				FileName: "file.mjs",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: messageID,
					Message:   messageString,
				}},
			},
			{
				Code:     `var value = 1; value.flat(2);`,
				FileName: "file.mjs",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: messageID,
					Message:   messageString,
				}},
			},
			{
				Code:     `const {value = 1} = {}; value.flat(2);`,
				FileName: "file.mjs",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: messageID,
					Message:   messageString,
				}},
			},
			{
				Code:     `const value = Math.PI; value.flat(2);`,
				FileName: "file.mjs",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: messageID,
					Message:   messageString,
				}},
			},

			// Source-only unknown receivers: calls the rslint static
			// evaluator cannot fold. The rslint type checker would
			// over-classify their return type as a known non-array
			// primitive, so the source-only path bypasses it.
			{
				Code:     `Math.random().flat(2)`,
				FileName: "file.mjs",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: messageID,
					Message:   messageString,
				}},
			},
			{
				Code:     `Array.from([1]).flat(2)`,
				FileName: "file.mjs",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: messageID,
					Message:   messageString,
				}},
			},
			{
				Code:     `Object.keys({a: 1}).flat(2)`,
				FileName: "file.mjs",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: messageID,
					Message:   messageString,
				}},
			},
			{
				Code:     `JSON.parse("[]").flat(2)`,
				FileName: "file.mjs",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: messageID,
					Message:   messageString,
				}},
			},
		},
	)
}
