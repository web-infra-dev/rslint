package ecmascript

import (
	"math"
	"strconv"
	"strings"
)

// NumberToString writes a number the way JavaScript writes one, which is what
// `String(n)` and string concatenation produce: the shortest run of digits that
// reads back as the same value, no sign on a zero, and exponential notation
// once the decimal point sits past the twenty-first place.
//
// Go's strconv agrees on the digits — both pick the shortest representation
// that round-trips — but not on when to leave fixed notation, nor on how to
// spell the infinities.
//
// Only whole numbers are covered: a value carrying a fraction is written in
// fixed notation whatever its size, where JavaScript switches to exponential
// notation below 1e-6.
//
// https://tc39.es/ecma262/2024/multipage/ecmascript-data-types-and-values.html#sec-numeric-types-number-tostring
func NumberToString(value float64) string {
	switch {
	case value == 0:
		return "0"
	case math.IsNaN(value):
		return "NaN"
	case math.IsInf(value, 1):
		return "Infinity"
	case math.IsInf(value, -1):
		return "-Infinity"
	}

	sign := ""
	digits := strconv.FormatFloat(value, 'f', -1, 64)
	if strings.HasPrefix(digits, "-") {
		sign, digits = "-", digits[1:]
	}
	if len(digits) <= 21 || strings.ContainsRune(digits, '.') {
		return sign + digits
	}

	// The digits hold no decimal point by now, so the exponent is however many
	// of them follow the first.
	exponent := strconv.Itoa(len(digits) - 1)
	significant := strings.TrimRight(digits, "0")
	if len(significant) == 1 {
		return sign + significant + "e+" + exponent
	}
	return sign + significant[:1] + "." + significant[1:] + "e+" + exponent
}
