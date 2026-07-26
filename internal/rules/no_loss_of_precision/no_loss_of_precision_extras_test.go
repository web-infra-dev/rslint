package no_loss_of_precision

import (
	"math"
	"math/big"
	"sync"
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
			{Code: `const safeIntegerBoundary = 9007199254740992;`},
			{Code: `const min = 5e-324;`},
			{Code: `const max = 1.7976931348623157e308;`},
			{Code: `const carry = 9999999999999998;`},

			// ---- Dimension 4: non-decimal leading-zero spellings stay exact when the value is exact ----
			{Code: `const exactHex = 0x0001fffffffffffff;`},
			{Code: `const exactOctal = 0o000377777777777777777;`},
			{Code: `const exactShiftedHex = 0x40000000000000;`},

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
			noLossInvalid("const x =\n// leading line trivia\n9007199254740993;", `9007199254740993`),

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
			name: "Locks in the largest integer guaranteed exact by the decimal fast path",
			raw:  "9007199254740992",
			want: false,
		},
		{
			name: "Locks in the first integer above the exact decimal fast-path boundary",
			raw:  "9007199254740993",
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
			name: "Locks in a hexadecimal integer wider than 53 bits but exact after removing powers of two",
			raw:  "0x40000000000000",
			want: false,
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

func TestNoLossOfPrecisionConcurrent(t *testing.T) {
	tests := []struct {
		raw  string
		want bool
	}{
		{raw: "123.456", want: false},
		{raw: "255.10000610351562", want: true},
		{raw: "5e-324", want: false},
		{raw: "4e-324", want: true},
		{raw: "1.7976931348623157e308", want: false},
		{raw: "1.7976931348623159e308", want: true},
	}
	type failure struct {
		raw       string
		got, want bool
	}

	start := make(chan struct{})
	failures := make(chan failure, 64)
	var waitGroup sync.WaitGroup
	for range 64 {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			<-start
			for range 32 {
				for _, test := range tests {
					if got := losesPrecision(test.raw); got != test.want {
						failures <- failure{raw: test.raw, got: got, want: test.want}
						return
					}
				}
			}
		}()
	}
	close(start)
	waitGroup.Wait()
	close(failures)

	for failure := range failures {
		t.Errorf("losesPrecision(%q) = %v, want %v", failure.raw, failure.got, failure.want)
	}
}

func TestNotBaseTenLosesPrecisionMatchesFloat64(t *testing.T) {
	t.Parallel()

	for bitLength := 1; bitLength <= 1_050; bitLength++ {
		for _, significantBits := range []int{1, 2, 52, 53, 54, 60} {
			if significantBits > bitLength {
				continue
			}

			significantValue := new(big.Int).Lsh(big.NewInt(1), uint(significantBits-1))
			if significantBits > 1 {
				significantValue.Add(significantValue, big.NewInt(1))
			}
			original := new(big.Int).Lsh(significantValue, uint(bitLength-significantBits))
			want := referenceIntegerLosesPrecision(original)

			for _, base := range []int{2, 8, 16} {
				digits := original.Text(base)
				if got := notBaseTenLosesPrecision(digits, base); got != want {
					t.Fatalf(
						"notBaseTenLosesPrecision(%q, %d) = %v, want %v (bit length %d, significant bits %d)",
						digits,
						base,
						got,
						want,
						bitLength,
						significantBits,
					)
				}
			}
		}
	}
}

func referenceIntegerLosesPrecision(original *big.Int) bool {
	value, _ := new(big.Float).SetInt(original).Float64()
	if math.IsInf(value, 0) {
		return true
	}

	reconstructed := new(big.Int)
	new(big.Float).SetFloat64(value).Int(reconstructed)
	return original.Cmp(reconstructed) != 0
}

func TestNumberToPrecisionScientificMatchesUncached(t *testing.T) {
	t.Parallel()

	const mantissaMask = uint64(1<<52 - 1)
	mantissas := []uint64{0, 1, 1 << 51, mantissaMask}
	precisions := []int{1, 2, 6, 17, 100}

	for exponentBits := uint64(0); exponentBits < 0x7ff; exponentBits += 17 {
		for _, mantissa := range mantissas {
			value := math.Float64frombits(exponentBits<<52 | mantissa)
			if value == 0 || math.IsInf(value, 0) || math.IsNaN(value) {
				continue
			}
			for _, precision := range precisions {
				got := numberToPrecisionScientific(value, precision)
				want := uncachedNumberToPrecisionScientific(value, precision)
				if got != want {
					t.Fatalf(
						"numberToPrecisionScientific(%g, %d) = %#v, want %#v",
						value,
						precision,
						got,
						want,
					)
				}
			}
		}
	}
}

func uncachedNumberToPrecisionScientific(value float64, precision int) scientificNotation {
	rat := new(big.Rat).SetFloat64(math.Abs(value))
	if rat == nil {
		return scientificNotation{}
	}

	magnitude := int(math.Floor(math.Log10(math.Abs(value))))
	for rat.Cmp(uncachedDecimalPower(magnitude)) < 0 {
		magnitude--
	}
	for rat.Cmp(uncachedDecimalPower(magnitude+1)) >= 0 {
		magnitude++
	}

	scaled := new(big.Rat).Set(rat)
	scaleExponent := precision - 1 - magnitude
	if scaleExponent >= 0 {
		scaled.Mul(scaled, uncachedDecimalPower(scaleExponent))
	} else {
		scaled.Quo(scaled, uncachedDecimalPower(-scaleExponent))
	}

	rounded := uncachedRoundRatHalfUp(scaled)
	coefficient := rounded.String()
	if len(coefficient) > precision {
		magnitude += len(coefficient) - precision
		coefficient = coefficient[:precision]
	}
	for len(coefficient) < precision {
		coefficient = "0" + coefficient
	}

	return scientificNotation{
		coefficient: coefficient,
		magnitude:   magnitude,
	}
}

func uncachedRoundRatHalfUp(rat *big.Rat) *big.Int {
	quotient := new(big.Int)
	remainder := new(big.Int)
	quotient.QuoRem(rat.Num(), rat.Denom(), remainder)

	doubleRemainder := new(big.Int).Mul(remainder, big.NewInt(2))
	if doubleRemainder.Cmp(rat.Denom()) >= 0 {
		quotient.Add(quotient, big.NewInt(1))
	}
	return quotient
}

func uncachedDecimalPower(exponent int) *big.Rat {
	absoluteExponent := exponent
	if absoluteExponent < 0 {
		absoluteExponent = -absoluteExponent
	}
	power := new(big.Int).Exp(
		big.NewInt(10),
		big.NewInt(int64(absoluteExponent)),
		nil,
	)
	if exponent >= 0 {
		return new(big.Rat).SetInt(power)
	}
	return new(big.Rat).SetFrac(big.NewInt(1), power)
}
