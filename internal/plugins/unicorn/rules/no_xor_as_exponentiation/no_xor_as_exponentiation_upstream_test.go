// TestNoXorAsExponentiationUpstream migrates the full valid/invalid suite from
// upstream test/no-xor-as-exponentiation.js 1:1. Position assertions cover
// line/column for the operator token reported in every invalid case.
// rslint-specific lock-in cases live in no_xor_as_exponentiation_extras_test.go.
package no_xor_as_exponentiation_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_xor_as_exponentiation"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoXorAsExponentiationUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_xor_as_exponentiation.NoXorAsExponentiationRule,
		[]rule_tester.ValidTestCase{
			// ---- Already correct exponentiation ----
			jsValid(`2 ** 32`),

			// ---- Non-decimal literals are likely intentional bitwise XOR ----
			jsValid(`0xFF ^ 8`),
			jsValid(`2 ^ 0x10`),
			jsValid(`0b100 ^ 2`),
			jsValid(`0o20 ^ 2`),
			jsValid(`2 ^ 0o20`),

			// ---- Non-literal operands ----
			jsValid(`a ^ b`),
			jsValid(`x ^ 2`),
			jsValid(`2 ^ y`),
			jsValid(`flags ^ MASK`),

			// ---- Floats ----
			jsValid(`2.5 ^ 3`),
			jsValid(`2 ^ 3.5`),
			jsValid(`2e3 ^ 2`),

			// ---- BigInt ----
			jsValid(`2n ^ 32n`),

			// ---- Unary operand is not a literal ----
			jsValid(`2 ^ -3`),
			jsValid(`-2 ^ 3`),
			jsValid(`2 ^ +3`),

			// ---- Other operators ----
			jsValid(`2 | 8`),
			jsValid(`2 & 8`),
			jsValid(`2 << 8`),
			jsValid(`2 ** 8`),

			// ---- TypeScript: operand is a `TSAsExpression`, not a literal ----
			// rslint does not differentiate AsExpression from its operand at the
			// rule layer; Skip: the upstream parser boundary is not relevant here.
			{Code: `(2 as number) ^ 8`, Skip: true},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Plain decimal-integer pairs ----
			xorInvalid(`2 ^ 32`),
			xorInvalid(`3 ^ 3`),
			xorInvalid(`10 ^ 6`),
			xorInvalid(`0 ^ 0`),
			xorInvalid(`2 ^ 8`),

			// ---- Surrounding whitespace preserved by the suggestion ----
			xorInvalid(`2  ^  8`),

			// ---- Inside an expression ----
			xorInvalid(`const x = 2 ^ 8;`),
			xorInvalid(`foo(2 ^ 8)`),

			// ---- Numeric separators are still decimal integers ----
			xorInvalid(`10 ^ 1_000`),

			// ---- Nested: only the inner literal pair fires (the outer ^ has a non-literal operand) ----
			xorInvalid(`2 ^ 8 ^ 2`),

			// ---- Comment between operands must be preserved by the suggestion ----
			xorInvalid(`2 /* comment */ ^ 8`),

			// ---- TypeScript parser, plain literal pair still fires ----
			xorInvalid(`2 ^ 8`),
		},
	)
}
