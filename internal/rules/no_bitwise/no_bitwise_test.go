package no_bitwise

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoBitwiseRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoBitwiseRule,
		[]rule_tester.ValidTestCase{
			// ── Non-bitwise operators are ignored ────────────────────────
			{Code: `a + b`},
			{Code: `!a`},
			{Code: `a && b`},
			{Code: `a || b`},
			{Code: `a += b`},
			{Code: `a &&= b`},
			{Code: `a ||= b`},
			{Code: `a ??= b`},

			// ── `allow` option ────────────────────────────────────────────
			{
				Code:    `~[1, 2, 3].indexOf(1)`,
				Options: map[string]interface{}{"allow": []interface{}{"~"}},
			},
			{
				Code:    `~1<<2 === -8`,
				Options: map[string]interface{}{"allow": []interface{}{"~", "<<"}},
			},

			// ── `int32Hint` option: canonical + alternative zero forms ────
			{
				Code:    `a|0`,
				Options: map[string]interface{}{"int32Hint": true},
			},
			{
				Code:    `a | (0)`, // Parenthesized zero; ESLint treats parens as transparent.
				Options: map[string]interface{}{"int32Hint": true},
			},
			{
				Code:    `a | 0.0`, // Float zero — Number("0.0") === 0.
				Options: map[string]interface{}{"int32Hint": true},
			},
			{
				Code:    `a | 0x0`, // Hex zero.
				Options: map[string]interface{}{"int32Hint": true},
			},
			{
				Code:    `a | 0b0`, // Binary zero.
				Options: map[string]interface{}{"int32Hint": true},
			},
			{
				Code:    `a | 0o0`, // Octal zero.
				Options: map[string]interface{}{"int32Hint": true},
			},
			{
				Code:    `a | 0e0`, // Scientific zero.
				Options: map[string]interface{}{"int32Hint": true},
			},
			{
				Code:    `a | 0E0`, // Uppercase exponent.
				Options: map[string]interface{}{"int32Hint": true},
			},
			{
				Code:    `a | 0e+0`, // Signed positive exponent.
				Options: map[string]interface{}{"int32Hint": true},
			},
			{
				Code:    `a | 0e-0`, // Signed negative exponent.
				Options: map[string]interface{}{"int32Hint": true},
			},
			{
				Code:    `a | 0.`, // Trailing decimal point.
				Options: map[string]interface{}{"int32Hint": true},
			},
			{
				Code:    `a | .0`, // Leading decimal point.
				Options: map[string]interface{}{"int32Hint": true},
			},
			{
				Code:    `a | 0.0e0`, // Combined float + exponent.
				Options: map[string]interface{}{"int32Hint": true},
			},
			{
				Code:    `a | 0X0`, // Uppercase hex prefix.
				Options: map[string]interface{}{"int32Hint": true},
			},

			// ── `allow` takes precedence; int32Hint may be off explicitly ─
			{
				Code: `a|0`,
				Options: map[string]interface{}{
					"allow":     []interface{}{"|"},
					"int32Hint": false,
				},
			},

			// ── Type-level unions/intersections are NOT BinaryExpression ─
			// Ensures the rule does not fire on TypeScript type syntax.
			{Code: `type T = A | B`},
			{Code: `type T = A & B`},
			{Code: `let x: string | number = 1`},
			{Code: `let x: { a: 1 } & { b: 2 } = { a: 1, b: 2 }`},
			{Code: `function f<T extends A | B>(x: T): T { return x; }`},
		},
		[]rule_tester.InvalidTestCase{
			// ── All 7 binary + 6 compound assignment + unary ~ operators ──
			{
				Code: `a ^ b`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code: `a | b`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code: `a & b`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code: `a << b`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code: `a >> b`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code: `a >>> b`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code: `a|0`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code: `~a`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code: `a ^= b`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code: `a |= b`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code: `a &= b`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code: `a <<= b`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code: `a >>= b`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code: `a >>>= b`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},

			// ── Nested / chained uses: each operator reports separately ──
			{
				Code: `a | b | c`, // Parses as (a | b) | c — inner + outer.
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code: `~~a`, // Double bitwise NOT — outer + inner.
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
					{MessageId: "unexpected", Line: 1, Column: 2},
				},
			},
			{
				Code: `~(a | b)`, // Unary over parenthesized binary — both report.
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
					{MessageId: "unexpected", Line: 1, Column: 3},
				},
			},

			// ── Parenthesized wrapping does not suppress the report ──────
			{
				Code: `(a | b)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 2},
				},
			},

			// ── int32Hint boundaries ──────────────────────────────────────
			{
				Code:    `a | 1`, // int32Hint is only for `| 0`.
				Options: map[string]interface{}{"int32Hint": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code:    `a | 0x10`, // Non-zero hex must still report (sanity).
				Options: map[string]interface{}{"int32Hint": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code:    `a | 1.0`, // Non-zero float must still report.
				Options: map[string]interface{}{"int32Hint": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code:    `0 | a`, // int32Hint requires zero on the RIGHT.
				Options: map[string]interface{}{"int32Hint": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code:    `a & 0`, // int32Hint is `|` only, not `&`.
				Options: map[string]interface{}{"int32Hint": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code:    `a | -0`, // Unary minus → not a Literal node in ESTree.
				Options: map[string]interface{}{"int32Hint": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code:    `a | 0n`, // BigInt zero is distinct from numeric 0.
				Options: map[string]interface{}{"int32Hint": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},

			// ── Bitwise inside TypeScript value-level constructs ─────────
			{
				Code: `enum E { A = 1 << 0, B = 1 << 1 }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 14},
					{MessageId: "unexpected", Line: 1, Column: 26},
				},
			},
			{
				Code: `const o = { [a | b]: 1 }`, // Computed property name.
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 14},
				},
			},
			{
				Code: `class C { m() { return a | b; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 24},
				},
			},

			// ── `~` not listed in `allow` still reports ──────────────────
			{
				Code:    `~a`,
				Options: map[string]interface{}{"allow": []interface{}{"|"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},

			// ── Mixed precedence: arithmetic binds tighter than bitwise ──
			// Parses as `(a + b) & c` — only `&` is bitwise.
			{
				Code: `a + b & c`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},

			// ── Right-associative compound bitwise assignment ────────────
			// `a |= b |= c` parses as `a |= (b |= c)` → both fire.
			{
				Code: `a |= b |= c`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
					{MessageId: "unexpected", Line: 1, Column: 6},
				},
			},

			// ── Unary `~` stacked on binary result ───────────────────────
			{
				Code: `~(a & b)`, // outer `~` + inner `&`.
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
					{MessageId: "unexpected", Line: 1, Column: 3},
				},
			},

			// ── Bitwise in Identifier-looking-but-computed positions ─────
			{
				Code: `obj[a | b]`, // computed element access.
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 5},
				},
			},

			// ── Bitwise inside template literal span ─────────────────────
			{
				Code: "`${a | b}`",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 4},
				},
			},
		},
	)
}
