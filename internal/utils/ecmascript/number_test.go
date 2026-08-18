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
