package ecmascript

import (
	"math"
	"testing"
)

// Every expectation here is what `String(n)` answers in JavaScript.
func TestNumberToString(t *testing.T) {
	tests := []struct {
		name  string
		value float64
		want  string
	}{
		{name: "zero", value: 0, want: "0"},
		// JavaScript writes a negative zero without its sign; Go writes "-0".
		{name: "negative zero", value: math.Copysign(0, -1), want: "0"},
		{name: "one", value: 1, want: "1"},
		{name: "minus one", value: -1, want: "-1"},
		{name: "fraction", value: 1.25, want: "1.25"},
		{name: "last small fixed", value: 1e-6, want: "0.000001"},
		{name: "first small exponential", value: 1e-7, want: "1e-7"},
		{name: "small fraction", value: 1.2e-7, want: "1.2e-7"},
		{name: "large", value: 1e15, want: "1000000000000000"},
		// The decimal point sits in the twenty-first place here and the
		// twenty-second one below, which is where JavaScript switches.
		{name: "last fixed", value: 1e20, want: "100000000000000000000"},
		{name: "first exponential", value: 1e21, want: "1e+21"},
		{name: "first exponential negative", value: -1e21, want: "-1e+21"},
		{name: "many significant digits", value: 1.2345678901234569e23, want: "1.2345678901234569e+23"},
		// Past 2^53 the shortest run of digits that reads back the same is
		// shorter than the whole number.
		{name: "past 2^53", value: 9223372036854774096, want: "9223372036854774000"},
		{name: "fraction", value: 1.5, want: "1.5"},
		{name: "fraction with several digits", value: 123.456, want: "123.456"},
		{name: "below one", value: 0.5, want: "0.5"},
		// The decimal point sits in the sixth place behind the digits here and
		// the seventh one below, which is where JavaScript switches.
		{name: "last fixed fraction", value: 1e-6, want: "0.000001"},
		{name: "first small exponential", value: 1e-7, want: "1e-7"},
		{name: "first small exponential negative", value: -1e-7, want: "-1e-7"},
		{name: "small with several digits", value: 1.25e-7, want: "1.25e-7"},
		{name: "not a number", value: math.NaN(), want: "NaN"},
		{name: "infinity", value: math.Inf(1), want: "Infinity"},
		{name: "negative infinity", value: math.Inf(-1), want: "-Infinity"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := NumberToString(test.value); got != test.want {
				t.Errorf("NumberToString(%v) = %q, want %q", test.value, got, test.want)
			}
		})
	}
}

func TestStringToNumber(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  float64
		ok    bool
	}{
		{name: "empty", value: "", want: 0, ok: true},
		{name: "ECMAScript whitespace", value: "\uFEFF\u00A01\u2029", want: 1, ok: true},
		{name: "negative zero", value: "-0", want: math.Copysign(0, -1), ok: true},
		{name: "decimal", value: "1.25", want: 1.25, ok: true},
		{name: "exponent", value: "1e2", want: 100, ok: true},
		{name: "hex", value: "0x10", want: 16, ok: true},
		{name: "octal", value: "0o10", want: 8, ok: true},
		{name: "binary", value: "0b10", want: 2, ok: true},
		{name: "positive infinity", value: "+Infinity", want: math.Inf(1), ok: true},
		{name: "overflow", value: "1e9999", want: math.Inf(1), ok: true},
		{name: "underflow", value: "1e-400", want: 0, ok: true},
		{name: "malformed overflow", value: "1e9999x", ok: false},
		{name: "not a number", value: "NaN", ok: false},
		{name: "Go infinity", value: "Inf", ok: false},
		{name: "case-folded infinity", value: "infinity", ok: false},
		{name: "signed hex", value: "-0x10", ok: false},
		{name: "positive signed hex", value: "+0x10", ok: false},
		{name: "sign after hex prefix", value: "0x+10", ok: false},
		{name: "hex float", value: "0x1p2", ok: false},
		{name: "numeric separator", value: "1_000", ok: false},
		{name: "text", value: "value", ok: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, ok := StringToNumber(test.value)
			if ok != test.ok {
				t.Fatalf("StringToNumber(%q) ok = %v, want %v", test.value, ok, test.ok)
			}
			if ok && math.Float64bits(got) != math.Float64bits(test.want) {
				t.Errorf("StringToNumber(%q) = %v, want %v", test.value, got, test.want)
			}
		})
	}
}

func TestStringToBigInt(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
		ok    bool
	}{
		{name: "empty", value: "", want: "0", ok: true},
		{name: "whitespace", value: " \t\n", want: "0", ok: true},
		{name: "decimal leading zero", value: "010", want: "10", ok: true},
		{name: "decimal invalid octal digit", value: "08", want: "8", ok: true},
		{name: "signed decimal", value: "+10", want: "10", ok: true},
		{name: "negative decimal", value: "-10", want: "-10", ok: true},
		{name: "hex", value: "0x10", want: "16", ok: true},
		{name: "octal", value: "0o10", want: "8", ok: true},
		{name: "binary", value: "0b10", want: "2", ok: true},
		{name: "positive signed hex", value: "+0x10", ok: false},
		{name: "negative signed hex", value: "-0x10", ok: false},
		{name: "positive signed octal", value: "+0o10", ok: false},
		{name: "negative signed octal", value: "-0o10", ok: false},
		{name: "positive signed binary", value: "+0b10", ok: false},
		{name: "negative signed binary", value: "-0b10", ok: false},
		{name: "positive sign after hex prefix", value: "0x+10", ok: false},
		{name: "negative sign after hex prefix", value: "0x-10", ok: false},
		{name: "positive sign after octal prefix", value: "0o+10", ok: false},
		{name: "negative sign after binary prefix", value: "0b-10", ok: false},
		{name: "double positive sign", value: "++10", ok: false},
		{name: "positive negative signs", value: "+-10", ok: false},
		{name: "negative positive signs", value: "-+10", ok: false},
		{name: "double negative sign", value: "--10", ok: false},
		{name: "sign only", value: "+", ok: false},
		{name: "bare prefix", value: "0x", ok: false},
		{name: "numeric separator", value: "1_000", ok: false},
		{name: "BigInt suffix", value: "10n", ok: false},
		{name: "decimal point", value: "1.0", ok: false},
		{name: "exponent", value: "1e2", ok: false},
		{name: "text", value: "value", ok: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, ok := StringToBigInt(test.value)
			if ok != test.ok {
				t.Fatalf("StringToBigInt(%q) ok = %v, want %v", test.value, ok, test.ok)
			}
			if ok && got.String() != test.want {
				t.Errorf("StringToBigInt(%q) = %s, want %s", test.value, got, test.want)
			}
		})
	}
}
