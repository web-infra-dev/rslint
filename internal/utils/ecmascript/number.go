package ecmascript

import (
	"errors"
	"math"
	"math/big"
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

// StringToNumber reads a string the way JavaScript reads one wherever a number
// is wanted — `Number(s)`, `+s`, or either operand of a relational comparison —
// and answers NaN when the string is not a numeric literal at all.
//
// Go's strconv answers a nearby question and disagrees at both edges. It reads
// spellings ECMAScript has no grammar for: digit separators (`1_000`), the
// infinities and NaN under other names (`inf`, `nan`), and hexadecimal floats
// (`0x1p2`). It also refuses the `0x`/`0o`/`0b` integers ECMAScript does accept
// here, which is the one place a numeric string is read in another base — a
// leading zero on its own never reads as octal. So the string is measured
// against the StringNumericLiteral grammar first, and only a string that
// grammar accepts is handed to strconv.
//
// https://tc39.es/ecma262/2024/multipage/abstract-operations.html#sec-stringtonumber
func StringToNumber(s string) float64 {
	s = StringTrim(s)

	switch s {
	case "":
		return 0
	case "Infinity", "+Infinity":
		return math.Inf(1)
	case "-Infinity":
		return math.Inf(-1)
	}

	if len(s) > 2 && s[0] == '0' {
		if base, ok := nonDecimalIntegerBase(s[1]); ok {
			return nonDecimalIntegerValue(s[2:], base)
		}
	}

	if !isStrDecimalLiteral(s) {
		return math.NaN()
	}
	// A number too large or too small to hold is an infinity or a zero in
	// JavaScript too, which is what strconv answers alongside ErrRange.
	value, err := strconv.ParseFloat(s, 64)
	if err != nil && !errors.Is(err, strconv.ErrRange) {
		return math.NaN()
	}
	return value
}

// nonDecimalIntegerBase maps the letter of a `0x`/`0o`/`0b` prefix to the base
// it introduces.
func nonDecimalIntegerBase(letter byte) (base int, ok bool) {
	switch letter {
	case 'x', 'X':
		return 16, true
	case 'o', 'O':
		return 8, true
	case 'b', 'B':
		return 2, true
	}
	return 0, false
}

// nonDecimalIntegerValue reads the digits following a `0x`/`0o`/`0b` prefix.
// The grammar allows nothing but digits of that base there — no sign, no
// separator — and however many of them: a value past what a float64 holds
// rounds the same way JavaScript rounds it, to the nearest representable
// number and to an infinity beyond the largest.
func nonDecimalIntegerValue(digits string, base int) float64 {
	for _, digit := range []byte(digits) {
		if digitValue(digit) >= base {
			return math.NaN()
		}
	}
	value, ok := new(big.Int).SetString(digits, base)
	if !ok {
		return math.NaN()
	}
	float, _ := new(big.Float).SetInt(value).Float64()
	return float
}

// digitValue returns the value digit carries as a base-16 digit, or 16 for a
// byte that is no digit at all.
func digitValue(digit byte) int {
	switch {
	case '0' <= digit && digit <= '9':
		return int(digit - '0')
	case 'a' <= digit && digit <= 'f':
		return int(digit-'a') + 10
	case 'A' <= digit && digit <= 'F':
		return int(digit-'A') + 10
	}
	return 16
}

// isStrDecimalLiteral reports whether s is a StrDecimalLiteral: an optional
// sign, then digits carrying an optional fraction, or a fraction on its own,
// followed by an optional exponent. Either the whole part or the fractional
// part may be left out, but not both.
//
// https://tc39.es/ecma262/2024/multipage/abstract-operations.html#prod-StrDecimalLiteral
func isStrDecimalLiteral(s string) bool {
	if s != "" && (s[0] == '+' || s[0] == '-') {
		s = s[1:]
	}

	if index := strings.IndexAny(s, "eE"); index >= 0 {
		exponent := s[index+1:]
		if exponent != "" && (exponent[0] == '+' || exponent[0] == '-') {
			exponent = exponent[1:]
		}
		if exponent == "" || !isAllDigits(exponent) {
			return false
		}
		s = s[:index]
	}

	whole, fraction, hasPoint := strings.Cut(s, ".")
	if !isAllDigits(whole) || !isAllDigits(fraction) {
		return false
	}
	if hasPoint {
		return whole != "" || fraction != ""
	}
	return whole != ""
}

func isAllDigits(s string) bool {
	for _, digit := range []byte(s) {
		if digit < '0' || digit > '9' {
			return false
		}
	}
	return true
}
