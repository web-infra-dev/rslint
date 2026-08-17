package ecmascript

import (
	"math"
	"strconv"
	"strings"
)

// NumberToString writes a number the way JavaScript writes one, which is what
// `String(n)` and string concatenation produce: the shortest run of digits that
// reads back as the same value, no sign on a zero, and exponential notation
// once the decimal point sits past the twenty-first place ahead of those digits
// or more than six places behind them.
//
// Go's strconv agrees on the digits — both pick the shortest representation
// that round-trips — but not on when to leave fixed notation, nor on how to
// spell the infinities.
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
	if value < 0 {
		sign, value = "-", -value
	}

	// The spec names the digits themselves k, and the place the decimal point
	// sits in front of them n, so that the value is the digits times ten to the
	// n minus k. Exponential notation hands back both.
	mantissa, exponent, _ := strings.Cut(strconv.FormatFloat(value, 'e', -1, 64), "e")
	digits := strings.Replace(mantissa, ".", "", 1)
	k := len(digits)
	n, _ := strconv.Atoi(exponent)
	n++

	switch {
	case k <= n && n <= 21:
		return sign + digits + strings.Repeat("0", n-k)
	case 0 < n && n <= 21:
		return sign + digits[:n] + "." + digits[n:]
	case -6 < n && n <= 0:
		return sign + "0." + strings.Repeat("0", -n) + digits
	}

	// A negative exponent carries its own sign out of strconv; a positive one
	// is written with a plus in front of it.
	magnitude := strconv.Itoa(n - 1)
	if n > 1 {
		magnitude = "+" + magnitude
	}
	if k == 1 {
		return sign + digits + "e" + magnitude
	}
	return sign + digits[:1] + "." + digits[1:] + "e" + magnitude
}
