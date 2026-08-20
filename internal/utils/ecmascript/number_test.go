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

// Every expectation here is what `String(Number(s))` answers in JavaScript.
func TestStringToNumber(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{input: "", want: "0"},
		{input: " ", want: "0"},
		{input: "\t\n ", want: "0"},
		{input: "\ufeff5", want: "5"},
		{input: "5\ufeff", want: "5"},
		{input: "0x10", want: "16"},
		{input: "0X10", want: "16"},
		{input: "0b11", want: "3"},
		{input: "0B11", want: "3"},
		{input: "0o17", want: "15"},
		{input: "0O17", want: "15"},
		{input: "017", want: "17"},
		{input: "08", want: "8"},
		{input: "0x", want: "NaN"},
		{input: "0b", want: "NaN"},
		{input: "0x+1", want: "NaN"},
		{input: "0x-1", want: "NaN"},
		{input: "0x1p2", want: "NaN"},
		{input: "1_000", want: "NaN"},
		{input: "1_0.5", want: "NaN"},
		{input: "inf", want: "NaN"},
		{input: "Infinity", want: "Infinity"},
		{input: "+Infinity", want: "Infinity"},
		{input: "-Infinity", want: "-Infinity"},
		{input: "infinity", want: "NaN"},
		{input: "INFINITY", want: "NaN"},
		{input: "NaN", want: "NaN"},
		{input: "nan", want: "NaN"},
		{input: ".5", want: "0.5"},
		{input: "5.", want: "5"},
		{input: "1.e5", want: "100000"},
		{input: "5e", want: "NaN"},
		{input: "5e+", want: "NaN"},
		{input: "5e3", want: "5000"},
		{input: "5E-3", want: "0.005"},
		{input: "+5", want: "5"},
		{input: "-5", want: "-5"},
		{input: "- 5", want: "NaN"},
		{input: "1e400", want: "Infinity"},
		{input: "1e-400", want: "0"},
		{input: "0x1FFFFFFFFFFFFFFFFF", want: "590295810358705700000"},
		{input: "0xffffffffffffffffffffffffffffffffff", want: "8.711228593176025e+40"},
		{input: "9007199254740993", want: "9007199254740992"},
		{input: ".", want: "NaN"},
		{input: "..5", want: "NaN"},
		{input: "1.2.3", want: "NaN"},
		{input: "0o8", want: "NaN"},
		{input: "0b2", want: "NaN"},
		{input: "0xg", want: "NaN"},
		{input: " 12 ", want: "12"},
		{input: "\n12\t", want: "12"},
		{input: "12\u00a0", want: "12"},
		{input: "\u00a012", want: "12"},
		{input: "1,000", want: "NaN"},
		{input: "\u0665", want: "NaN"},
		{input: "0x1_0", want: "NaN"},
		{input: "+0x10", want: "NaN"},
		{input: "-0", want: "0"},
		{input: "0.0", want: "0"},
		{input: "00.5", want: "0.5"},
		{input: "0e0", want: "0"},
		{input: ".e5", want: "NaN"},
		{input: "1e", want: "NaN"},
		{input: "e5", want: "NaN"},
		{input: "0xA", want: "10"},
		{input: "0XaB", want: "171"},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			if got := NumberToString(StringToNumber(test.input)); got != test.want {
				t.Errorf("StringToNumber(%q) = %v, want %v", test.input, got, test.want)
			}
		})
	}
}
