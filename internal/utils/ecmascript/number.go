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

// NumberToExponential implements Number.prototype.toExponential's finite
// number formatting. A negative fractionDigits value requests the shortest
// round-tripping representation used when the argument is omitted.
func NumberToExponential(value float64, fractionDigits int) string {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return NumberToString(value)
	}
	sign := ""
	if value < 0 {
		sign = "-"
		value = -value
	} else if value == 0 {
		// JavaScript suppresses the sign of negative zero in decimal formats.
		value = 0
	}
	if fractionDigits < 0 {
		scientific := strconv.FormatFloat(value, 'e', -1, 64)
		exponentMarker := strings.LastIndexByte(scientific, 'e')
		exponent, _ := strconv.Atoi(scientific[exponentMarker+1:])
		return sign + scientific[:exponentMarker] + "e" + signedDecimalExponent(exponent)
	}
	digits, exponent := roundedSignificantDecimal(value, fractionDigits+1)
	mantissa := digits[:1]
	if fractionDigits > 0 {
		mantissa += "." + digits[1:]
	}
	return sign + mantissa + "e" + signedDecimalExponent(exponent)
}

// NumberToPrecision implements Number.prototype.toPrecision's finite number
// formatting for precisions in the ECMAScript range [1, 100].
func NumberToPrecision(value float64, precision int) string {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return NumberToString(value)
	}
	sign := ""
	if value < 0 {
		sign = "-"
		value = -value
	}
	if value == 0 {
		if precision == 1 {
			return "0"
		}
		return "0." + strings.Repeat("0", precision-1)
	}

	digits, exponent := roundedSignificantDecimal(value, precision)
	if exponent < -6 || exponent >= precision {
		if len(digits) == 1 {
			return sign + digits + "e" + signedDecimalExponent(exponent)
		}
		return sign + digits[:1] + "." + digits[1:] + "e" + signedDecimalExponent(exponent)
	}

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

// roundedSignificantDecimal implements the finite, positive-number rounding
// step shared by ECMAScript's ToRawPrecision and ToRawExponential operations.
// The specification chooses the larger decimal integer on an exact tie; Go's
// strconv fixed-precision formatting uses ties-to-even and therefore cannot be
// used for values such as 1.25 at two significant digits.
func roundedSignificantDecimal(value float64, precision int) (digits string, exponent int) {
	if value == 0 {
		return strings.Repeat("0", precision), 0
	}
	rational := new(big.Rat).SetFloat64(value)
	numerator := new(big.Int).Set(rational.Num())
	denominator := new(big.Int).Set(rational.Denom())
	scientific := strconv.FormatFloat(value, 'e', -1, 64)
	exponentMarker := strings.LastIndexByte(scientific, 'e')
	exponent, _ = strconv.Atoi(scientific[exponentMarker+1:])
	// The shortest round-tripping decimal can cross a power-of-ten boundary
	// (for example, float64(1e-7) is slightly smaller than exact 1e-7). Correct
	// the estimate against the exact binary rational before scaling.
	for compareRationalToPowerOfTen(numerator, denominator, exponent) < 0 {
		exponent--
	}
	for compareRationalToPowerOfTen(numerator, denominator, exponent+1) >= 0 {
		exponent++
	}
	scale := precision - 1 - exponent
	if scale >= 0 {
		numerator.Mul(numerator, decimalPowerOfTen(scale))
	} else {
		denominator.Mul(denominator, decimalPowerOfTen(-scale))
	}

	remainder := new(big.Int)
	rounded := new(big.Int)
	rounded.QuoRem(numerator, denominator, remainder)
	if remainder.Lsh(remainder, 1).Cmp(denominator) >= 0 {
		rounded.Add(rounded, big.NewInt(1))
	}
	digits = rounded.String()
	if len(digits) > precision {
		rounded.Quo(rounded, big.NewInt(10))
		digits = rounded.String()
		exponent++
	}
	if len(digits) < precision {
		digits = strings.Repeat("0", precision-len(digits)) + digits
	}
	return digits, exponent
}

func decimalPowerOfTen(exponent int) *big.Int {
	return new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(exponent)), nil)
}

func compareRationalToPowerOfTen(numerator, denominator *big.Int, exponent int) int {
	if exponent >= 0 {
		right := new(big.Int).Mul(denominator, decimalPowerOfTen(exponent))
		return numerator.Cmp(right)
	}
	left := new(big.Int).Mul(numerator, decimalPowerOfTen(-exponent))
	return left.Cmp(denominator)
}

func signedDecimalExponent(exponent int) string {
	if exponent >= 0 {
		return "+" + strconv.Itoa(exponent)
	}
	return strconv.Itoa(exponent)
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
