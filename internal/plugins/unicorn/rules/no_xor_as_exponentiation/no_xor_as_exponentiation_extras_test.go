// TestNoXorAsExponentiationExtras locks in branches and edge shapes that the
// upstream test suite doesn't exercise. Each case carries an inline comment
// pointing at the specific branch it covers, so future refactors can't silently
// regress them without breaking a named lock-in.
package no_xor_as_exponentiation_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_xor_as_exponentiation"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// tsValid is the in-package alias for jsValid — the rule is fully syntactic
// and has no type-aware branches, so .ts and .mjs fixtures exercise the same
// path.
func tsValid(code string) rule_tester.ValidTestCase {
	return rule_tester.ValidTestCase{Code: code, FileName: "file.mjs"}
}

func TestNoXorAsExponentiationExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_xor_as_exponentiation.NoXorAsExponentiationRule,
		[]rule_tester.ValidTestCase{
			// ---- Branch: not a BinaryExpression ----
			tsValid(`2;`),
			tsValid(`let x = 2;`),

			// ---- Branch: BinaryExpression with a non-CaretToken operator ----
			tsValid(`2 + 8`),
			tsValid(`2 - 8`),
			tsValid(`2 * 8`),
			tsValid(`2 / 8`),
			tsValid(`2 % 8`),
			tsValid(`2 & 8`),
			tsValid(`2 | 8`),
			tsValid(`2 << 8`),
			tsValid(`2 >> 8`),
			tsValid(`2 >>> 8`),

			// ---- Branch: only one operand is a decimal-integer literal ----
			// Right side is not a literal.
			tsValid(`2 ^ identifier`),
			tsValid(`2 ^ functionCall()`),
			// Left side is not a literal.
			tsValid(`identifier ^ 2`),
			tsValid(`(a + b) ^ 2`),

			// ---- Branch: left is decimal-integer, right is some other numeric literal shape ----
			// Hex
			tsValid(`2 ^ 0xFF`),
			// Octal
			tsValid(`2 ^ 0o10`),
			// Binary
			tsValid(`2 ^ 0b10`),
			// Float
			tsValid(`2 ^ 1.5`),
			// Scientific
			tsValid(`2 ^ 1e2`),
			// BigInt — different literal kind entirely.
			tsValid(`2 ^ 8n`),
			// TypeScript rejects legacy octal (`01`, `007`) and decimal-with-leading-zero
			// (`08`, `09`) at parse time, so they don't reach the rule. The decimal-integer
			// regex intentionally accepts `0[0-7]*[89]\d*` to mirror ESLint's
			// legacy-octal-but-decimal classification; see the upstream test suite's
			// "Non-decimal literals are likely intentional bitwise XOR" block.

			// ---- Branch: literal is wrapped in a non-parenthesis expression ----
			// In tsgo (and ESTree) `-2` and `+2` are PrefixUnaryExpression, not NumericLiteral;
			// the raw-text check rejects them. Parens don't change the AST shape here.
			tsValid(`-2 ^ 3`),
			tsValid(`2 ^ -3`),
			tsValid(`+2 ^ 3`),
			tsValid(`2 ^ +3`),
			tsValid(`(-2) ^ 8`),
			tsValid(`2 ^ (-8)`),

			// ---- Branch: empty / comment-only files ----
			tsValid(``),
			tsValid(`// just a comment`),

			// ---- Branch: `^` inside a non-binary context (e.g. template literal, regex) ----
			tsValid("const r = /a^b/;"),
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: ESTree-transparent parentheses around operands ----
			// SkipParentheses unwraps to the inner NumericLiteral, so these still fire.
			xorInvalid(`(2) ^ 8`),
			xorInvalid(`2 ^ (8)`),
			xorInvalid(`((2)) ^ ((8))`),
			xorInvalid(`(((2 ^ 8)))`),

			// ---- Dimension 4: comments on either side of the operator survive the suggestion ----
			// The suggestion replaces only the `^` token, leaving surrounding whitespace and
			// comments in place.
			xorInvalid(`2/**/^ 8`),
			xorInvalid(`2 ^/**/ 8`),
			xorInvalid(`2/*a*/^/*b*/8`),

			// ---- Branch lock-in: single-digit pairs at both ends of the regex ----
			xorInvalid(`0 ^ 1`),
			xorInvalid(`9 ^ 9`),
			// Numeric separators — the regex `(?:_?\d)*` allows them.
			xorInvalid(`1_2 ^ 3_4`),
			xorInvalid(`1_000_000 ^ 2`),

			// ---- Branch lock-in: deeply nested binary expressions ----
			// Only the innermost `2 ^ 8` fires; the outer `^ a` operands are not both literals.
			// For `((2 ^ 8)) ^ a` the only caret the rule reports is the inner one.
			// `((2 ^ 8)) ^ a`:
			//   char 0: (  1: (  2: 2  3:   4: ^  5:   6: 8  7: )  8: )  9:   10: ^  11:   12: a
			xorAtInvalid(`((2 ^ 8)) ^ a`, 4),
			// `a ^ (2 ^ 8)`:
			//   char 0: a  1:   2: ^  3:   4: (  5: 2  6:   7: ^  8:   9: 8  10: )
			xorAtInvalid(`a ^ (2 ^ 8)`, 7),
			// `1 ^ (2 ^ 3) ^ 4` (left-associative: `(1 ^ (2 ^ 3)) ^ 4`):
			//   char 0: 1  1:   2: ^  3:   4: (  5: 2  6:   7: ^  8:   9: 3  10: )  11:   12: ^  13:   14: 4
			// Only the inner `2 ^ 3` matches; the outer `^` operands include a non-literal.
			xorAtInvalid(`1 ^ (2 ^ 3) ^ 4`, 7),

			// ---- Branch lock-in: XOR at unusual positions in an expression ----
			xorInvalid(`2 ^ 8 ? 1 : 0`),
			xorInvalid(`[2 ^ 8]`),
			xorInvalid(`({x: 2 ^ 8})`),
			xorInvalid(`function f() { return 2 ^ 8; }`),
			xorInvalid(`2 ^ 8;`),
		},
	)
}
