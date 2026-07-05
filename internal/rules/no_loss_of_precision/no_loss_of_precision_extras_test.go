package no_loss_of_precision

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestNoLossOfPrecisionExtras locks in branches and edge shapes that the upstream test suite doesn't exercise. Each case carries an inline comment pointing at the specific branch / Dimension 4 row / tsgo AST quirk it covers, so future refactors can't silently regress them without breaking a named lock-in.
func TestNoLossOfPrecisionExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoLossOfPrecisionRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: parenthesized safe literal ----
			{Code: `var x = ((123.456));`},

			// ---- Dimension 4: string and BigInt literals are not number literals ----
			{Code: `const x = "9007199254740993";`},
			{Code: `const x = 9007199254740993n;`},

			// ---- Dimension 4: string-literal keys are outside the numeric-literal listener ----
			{Code: `const x = { "9007199254740993": "value" };`},

			// ---- Dimension 4: nested safe literals remain valid in TS/JSX containers ----
			{Code: `enum E { A = 9007199254740991 }`},
			{Code: `interface I { value: 9007199254740991; }`},
			{Code: `type T = 9007199254740991;`},
			{Code: `declare const x: 9007199254740991;`},
			{Code: `const x = <Widget value={9007199254740991} />;`, Tsx: true},

			// ---- Dimension 4: graceful degradation on empty/rest shapes ----
			{Code: `const {} = source;`},
			{Code: `const { ...rest } = source;`},

			// ---- Dimension 4: IEEE-754 boundary values that remain representable ----
			{Code: `const min = 5e-324;`},
			{Code: `const max = 1.7976931348623157e308;`},
			{Code: `const carry = 9999999999999998;`},

			// ---- Dimension 4: non-decimal leading-zero spellings stay exact when the value is exact ----
			{Code: `const exactHex = 0x0001fffffffffffff;`},
			{Code: `const exactOctal = 0o000377777777777777777;`},

			// ---- Real-user: eslint/eslint#19957 plus-signed exponent remains valid ----
			{Code: `const test = 9.000e+3;`},
			{Code: `const baz = 9.0000e+5;`},

			// ---- Real-user: eslint/eslint#19957 underflow-to-zero remains valid ----
			{Code: `const tiny = 1e-9999;`},

			// N/A: optional chaining has no numeric literal receiver/key behavior for this rule.
			// N/A: access/key equivalence classes do not compare names; every NumericLiteral node is checked independently.
			// N/A: declaration/container forms do not change state; numeric literal nodes are inspected wherever they appear.
			// N/A: autofix boundaries do not apply because the rule has no fixer.
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: parenthesized literal ----
			noLossInvalid(`var x = (9007199254740993);`, `9007199254740993`),
			noLossInvalid(`var x = ((9007199254740993));`, `9007199254740993`),

			// ---- Dimension 4: TS expression wrappers on inspected literal ----
			noLossInvalid(`const x = 9007199254740993!;`, `9007199254740993`),
			noLossInvalid(`const x = 9007199254740993 as number;`, `9007199254740993`),
			noLossInvalid(`const x = 9007199254740993 satisfies number;`, `9007199254740993`),
			noLossInvalid(`const x = +9007199254740993;`, `9007199254740993`),

			// ---- Dimension 4: comments/trivia before a literal do not affect raw token comparison ----
			noLossInvalid(`const x = /* leading trivia */ 9007199254740993;`, `9007199254740993`),

			// ---- Dimension 4: numeric property/key forms ----
			noLossInvalid(`const x = { 9007199254740993: "value" };`, `9007199254740993`),
			noLossInvalid(`const x = { [9007199254740993]: "value" };`, `9007199254740993`),
			noLossInvalid(`class C { 9007199254740993() {} }`, `9007199254740993`),
			noLossInvalid(`class C { [9007199254740993]() {} }`, `9007199254740993`),
			noLossInvalid(`foo[9007199254740993];`, `9007199254740993`),
			noLossInvalid(`foo?.[9007199254740993];`, `9007199254740993`),

			// ---- Dimension 4: nested containers do not affect reporting ----
			noLossInvalid(`function f(value = 9007199254740993) { return value; }`, `9007199254740993`),
			noLossInvalid(`const [value = 9007199254740993] = values;`, `9007199254740993`),
			noLossInvalid(`const { value = 9007199254740993 } = source;`, `9007199254740993`),
			noLossInvalid(`enum E { A = 9007199254740993 }`, `9007199254740993`),
			noLossInvalid(`interface I { value: 9007199254740993; }`, `9007199254740993`),
			noLossInvalid(`type T = 9007199254740993;`, `9007199254740993`),
			noLossInvalid(`declare const x: 9007199254740993;`, `9007199254740993`),
			noLossInvalidWithTsx(`const x = <Widget value={9007199254740993} />;`, `9007199254740993`),

			// ---- Dimension 4: traversal reports each nested numeric literal independently ----
			noLossInvalidWithTargets(
				`function f() { return [9007199254740993, { value: 5123000000000000000000000000001 }]; }`,
				`9007199254740993`,
				`5123000000000000000000000000001`,
			),
			noLossInvalidWithTargets(
				`class C {
  field = 9007199254740993;
  #private = 5123000000000000000000000000001;
}`,
				`9007199254740993`,
				`5123000000000000000000000000001`,
			),
			noLossInvalidWithTargets(
				`class C {
  method() {
    return 9007199254740993;
  }
  static {
    const value = 5123000000000000000000000000001;
  }
}`,
				`9007199254740993`,
				`5123000000000000000000000000001`,
			),

			// ---- Dimension 4: spread sibling does not mask numeric literal checks ----
			noLossInvalid(`const x = { ...source, value: 9007199254740993 };`, `9007199254740993`),

			// ---- Dimension 4: template and sequence-expression containers still visit numeric literals ----
			noLossInvalid("const text = `value ${9007199254740993}`;", `9007199254740993`),
			noLossInvalid(`const value = (0, 9007199254740993);`, `9007199254740993`),

			// ---- Dimension 4: exponent signs and small-magnitude decimals ----
			noLossInvalid(`const x = 9.007199254740993e+15;`, `9.007199254740993e+15`),
			noLossInvalid(`const x = .9007199254740993e+16;`, `.9007199254740993e+16`),

			// ---- Dimension 4: IEEE-754 subnormal, max-value, and rounding-carry boundaries ----
			noLossInvalid(`const under = 4e-324;`, `4e-324`),
			noLossInvalid(`const over = 1.7976931348623159e308;`, `1.7976931348623159e308`),
			noLossInvalid(`const carry = 9999999999999999;`, `9999999999999999`),
			noLossInvalid(`const rounded = 99.99999999999998;`, `99.99999999999998`),

			// ---- Dimension 4: non-decimal leading-zero spellings still compare the intended value ----
			noLossInvalid(`const lossyHex = 0x00020000000000001;`, `0x00020000000000001`),
			noLossInvalid(`const lossyOctal = 0o000400000000000000001;`, `0o000400000000000000001`),

			// Locks in upstream baseTenLosesPrecision() arm 1: requested precision greater than 100 reports before formatting.
			noLossInvalid(`const x = 1.00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001;`, `1.00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001`),

			// Locks in upstream notBaseTenLosesPrecision() arm 1: uppercase binary prefix chooses base 2.
			noLossInvalid(`const x = 0B100000000000000000000000000000000000000000000000000001;`, `0B100000000000000000000000000000000000000000000000000001`),

			// Locks in upstream notBaseTenLosesPrecision() arm 2: uppercase hex prefix chooses base 16.
			noLossInvalid(`const x = 0X20000000000001;`, `0X20000000000001`),

			// Locks in upstream notBaseTenLosesPrecision() arm 3: octal prefixes use base 8.
			noLossInvalid(`const x = 0o400000000000000001;`, `0o400000000000000001`),

			// ---- Real-user: eslint/eslint#15767 working-as-intended precision report ----
			noLossInvalid(`const test = 555.9771118164062;`, `555.9771118164062`),

			// ---- Real-user: eslint/eslint#16989 working-as-intended precision report ----
			noLossInvalid(`const a = -9726.622680664062;`, `9726.622680664062`),

			// ---- Real-user: eslint/eslint#17492 working-as-intended precision report ----
			noLossInvalid(`const value = 255.10000610351562;`, `255.10000610351562`),
		},
	)
}

func TestNoLossOfPrecisionRawLiteralBranches(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
		want bool
	}{
		{
			name: "Locks in upstream losesPrecision() arm 1: binary prefix dispatches to base 2",
			raw:  "0b100000000000000000000000000000000000000000000000000001",
			want: true,
		},
		{
			name: "Locks in upstream losesPrecision() arm 2: leading-zero decimal with 8 stays base ten",
			raw:  "0009007199254740993",
			want: true,
		},
		{
			name: "Locks in upstream valid leading-zero decimal with fractional part",
			raw:  "019.5",
			want: false,
		},
		{
			name: "Locks in upstream valid leading-zero decimal integer",
			raw:  "00195",
			want: false,
		},
		{
			name: "Locks in upstream valid leading-zero decimal containing 8",
			raw:  "0008",
			want: false,
		},
		{
			name: "Locks in upstream losesPrecision() arm 3: exact legacy octal remains valid",
			raw:  "0377777777777777777",
			want: false,
		},
		{
			name: "Locks in upstream losesPrecision() arm 4: lossy legacy octal reports",
			raw:  "0400000000000000001",
			want: true,
		},
		{
			name: "Locks in non-decimal leading-zero exact hexadecimal spelling",
			raw:  "0x0001fffffffffffff",
			want: false,
		},
		{
			name: "Locks in non-decimal leading-zero lossy hexadecimal spelling",
			raw:  "0x00020000000000001",
			want: true,
		},
		{
			name: "Locks in non-decimal leading-zero exact octal spelling",
			raw:  "0o000377777777777777777",
			want: false,
		},
		{
			name: "Locks in non-decimal leading-zero lossy octal spelling",
			raw:  "0o000400000000000000001",
			want: true,
		},
		{
			name: "Locks in upstream baseTenLosesPrecision() arm 1: zero is skipped by the listener guard",
			raw:  "0.000000000000000000000000000000000000000000000000000000000000000000000000000000",
			want: false,
		},
		{
			name: "Locks in base-ten underflow-to-zero guard",
			raw:  "1e-9999",
			want: false,
		},
		{
			name: "Locks in smallest subnormal value",
			raw:  "5e-324",
			want: false,
		},
		{
			name: "Locks in subnormal rounding up to a different literal",
			raw:  "4e-324",
			want: true,
		},
		{
			name: "Locks in maximum finite value",
			raw:  "1.7976931348623157e308",
			want: false,
		},
		{
			name: "Locks in overflow to infinity",
			raw:  "1.7976931348623159e308",
			want: true,
		},
		{
			name: "Locks in rounding carry mismatch",
			raw:  "9999999999999999",
			want: true,
		},
		{
			name: "Locks in rounding carry exact neighbor",
			raw:  "9999999999999998",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := losesPrecision(tt.raw); got != tt.want {
				t.Fatalf("losesPrecision(%q) = %v, want %v", tt.raw, got, tt.want)
			}
		})
	}
}
