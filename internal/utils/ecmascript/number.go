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
// outside the [1e-6, 1e21) range.
//
// Go's strconv agrees on the significant digits — both pick the shortest
// representation that round-trips — but not on when to leave fixed notation,
// how to pad exponents, or how to spell the infinities.
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
		sign = "-"
		value = -value
	}

	scientific := strconv.FormatFloat(value, 'e', -1, 64)
	exponentMarker := strings.LastIndexByte(scientific, 'e')
	mantissa := scientific[:exponentMarker]
	exponent, _ := strconv.Atoi(scientific[exponentMarker+1:])
	digits := strings.Replace(mantissa, ".", "", 1)

	if exponent >= -6 && exponent < 21 {
		decimalPosition := exponent + 1
		switch {
		case decimalPosition <= 0:
			return sign + "0." + strings.Repeat("0", -decimalPosition) + digits
		case decimalPosition >= len(digits):
			return sign + digits + strings.Repeat("0", decimalPosition-len(digits))
		default:
			return sign + digits[:decimalPosition] + "." + digits[decimalPosition:]
		}
	}

	exponentText := strconv.Itoa(exponent)
	if exponent >= 0 {
		exponentText = "+" + exponentText
	}
	if len(digits) == 1 {
		return sign + digits + "e" + exponentText
	}
	return sign + digits[:1] + "." + digits[1:] + "e" + exponentText
}

// StringToNumber applies JavaScript's StringToNumber operation. The boolean is
// false exactly when JavaScript would produce NaN; infinities and signed zero
// are valid results.
func StringToNumber(value string) (float64, bool) {
	value = StringTrim(value)
	if value == "" {
		return 0, true
	}

	switch value {
	case "Infinity", "+Infinity":
		return math.Inf(1), true
	case "-Infinity":
		return math.Inf(-1), true
	}

	if strings.ContainsRune(value, '_') {
		return 0, false
	}

	prefixStart := 0
	if value[0] == '+' || value[0] == '-' {
		prefixStart = 1
	}
	if len(value) > prefixStart+2 && value[prefixStart] == '0' {
		base := 0
		switch value[prefixStart+1] {
		case 'x', 'X':
			base = 16
		case 'o', 'O':
			base = 8
		case 'b', 'B':
			base = 2
		}
		if base != 0 {
			// StringToNumber accepts only unsigned non-decimal prefixes.
			if prefixStart != 0 {
				return 0, false
			}
			digits := value[2:]
			if digits[0] == '+' || digits[0] == '-' {
				return 0, false
			}
			integer, ok := new(big.Int).SetString(digits, base)
			if !ok {
				return 0, false
			}
			number, _ := new(big.Float).SetInt(integer).Float64()
			return number, true
		}
	}

	number, err := strconv.ParseFloat(value, 64)
	if math.IsNaN(number) {
		return 0, false
	}
	if err != nil {
		// Decimal overflow and underflow are valid infinities and zeroes.
		// Other parse failures are NaN.
		if errors.Is(err, strconv.ErrRange) {
			return number, true
		}
		return 0, false
	}
	// ParseFloat additionally accepts Go's Inf spellings; JavaScript does not.
	if math.IsInf(number, 0) {
		return 0, false
	}
	return number, true
}

// StringToBigInt applies JavaScript's StringToBigInt operation. Decimal
// strings, including those with leading zeroes, are decimal; only unsigned
// 0x, 0o, and 0b prefixes select another radix. The boolean is false exactly
// when JavaScript would produce a syntax error while converting the string.
func StringToBigInt(value string) (*big.Int, bool) {
	value = StringTrim(value)
	if value == "" {
		return new(big.Int), true
	}

	sign := 1
	if value[0] == '+' || value[0] == '-' {
		if value[0] == '-' {
			sign = -1
		}
		value = value[1:]
		if value == "" {
			return nil, false
		}
	}

	base := 10
	digits := value
	if len(value) >= 2 && value[0] == '0' {
		switch value[1] {
		case 'x', 'X':
			if sign != 1 {
				return nil, false
			}
			base, digits = 16, value[2:]
		case 'o', 'O':
			if sign != 1 {
				return nil, false
			}
			base, digits = 8, value[2:]
		case 'b', 'B':
			if sign != 1 {
				return nil, false
			}
			base, digits = 2, value[2:]
		}
	}
	if digits == "" {
		return nil, false
	}

	n, ok := new(big.Int).SetString(digits, base)
	if !ok {
		return nil, false
	}
	if sign < 0 {
		n.Neg(n)
	}
	return n, true
}
