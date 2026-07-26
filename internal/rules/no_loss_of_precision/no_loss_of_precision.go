package no_loss_of_precision

import (
	"math"
	"math/big"
	"strconv"
	"strings"
	"sync"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
)

var noLossOfPrecisionMessage = rule.RuleMessage{
	Id:          "noLossOfPrecision",
	Description: "This number literal will lose precision at runtime.",
}

const maxExactDecimalInteger = "9007199254740992"

const (
	minCachedDecimalExponent = -324
	maxCachedDecimalExponent = 423
)

type decimalPowerCacheEntry struct {
	once  sync.Once
	value *big.Rat
}

// decimalPowerCache initializes only the exponents a process actually uses.
// Each value is immutable after publication, so rule listeners can safely
// share it across files without imposing a full-table cold-start cost.
var decimalPowerCache [maxCachedDecimalExponent - minCachedDecimalExponent + 1]decimalPowerCacheEntry

// https://eslint.org/docs/latest/rules/no-loss-of-precision
var NoLossOfPrecisionRule = rule.Rule{
	Name: "no-loss-of-precision",
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindNumericLiteral: func(node *ast.Node) {
				// tsgo normalizes NumericLiteral.Text at parse time, so ESLint parity
				// requires reading the raw token text to preserve prefixes, separators,
				// exponent spelling, and trailing fractional zeros.
				sourceText := ctx.SourceFile.Text()
				start := scanner.SkipTrivia(sourceText, node.Pos())
				raw := sourceText[start:node.End()]
				if losesPrecision(raw) {
					// Reuse the range already found for the raw token. ReportNode
					// would scan the same trivia a second time.
					ctx.ReportRange(core.NewTextRange(start, node.End()), noLossOfPrecisionMessage)
				}
			},
		}
	},
}

func removeNumericSeparators(s string) string {
	return strings.ReplaceAll(s, "_", "")
}

func losesPrecision(raw string) bool {
	normalized := removeNumericSeparators(raw)

	if len(normalized) >= 2 && normalized[0] == '0' {
		switch normalized[1] {
		case 'b', 'B':
			return notBaseTenLosesPrecision(normalized[2:], 2)
		case 'o', 'O':
			return notBaseTenLosesPrecision(normalized[2:], 8)
		case 'x', 'X':
			return notBaseTenLosesPrecision(normalized[2:], 16)
		}
	}
	if isLegacyOctal(normalized) {
		return notBaseTenLosesPrecision(normalized[1:], 8)
	}
	if isExactDecimalInteger(normalized) {
		return false
	}

	return baseTenLosesPrecision(normalized)
}

func isLegacyOctal(raw string) bool {
	if len(raw) < 2 || raw[0] != '0' {
		return false
	}
	for i := 1; i < len(raw); i++ {
		if raw[i] < '0' || raw[i] > '7' {
			return false
		}
	}
	return true
}

// isExactDecimalInteger recognizes the overwhelmingly common integer fast
// path. Every non-negative integer through 2^53 is exactly representable as a
// float64; larger integers continue through the exact comparison below.
func isExactDecimalInteger(raw string) bool {
	firstSignificantDigit := len(raw)
	for i := range len(raw) {
		if raw[i] < '0' || raw[i] > '9' {
			return false
		}
		if firstSignificantDigit == len(raw) && raw[i] != '0' {
			firstSignificantDigit = i
		}
	}
	if firstSignificantDigit == len(raw) {
		return len(raw) > 0
	}

	significantDigits := raw[firstSignificantDigit:]
	return len(significantDigits) < len(maxExactDecimalInteger) ||
		len(significantDigits) == len(maxExactDecimalInteger) &&
			significantDigits <= maxExactDecimalInteger
}

func notBaseTenLosesPrecision(digits string, base int) bool {
	// Prefix signs are not part of a NumericLiteral token. Reject them here as
	// the old full-token regular expressions did, rather than letting big.Int
	// accept malformed direct inputs.
	if len(digits) == 0 || digits[0] == '+' || digits[0] == '-' {
		return false
	}
	original := new(big.Int)
	_, ok := original.SetString(digits, base)
	if !ok {
		return false
	}
	if original.Sign() == 0 {
		return false
	}

	// An integer is exactly representable as a finite float64 iff its highest
	// set bit fits the finite exponent range and at most 53 significant bits
	// remain after removing factors of two.
	return original.BitLen() > 1024 ||
		original.BitLen()-int(original.TrailingZeroBits()) > 53
}

func baseTenLosesPrecision(raw string) bool {
	value, err := strconv.ParseFloat(raw, 64)
	if err != nil && !math.IsInf(value, 0) {
		return false
	}
	if value == 0 {
		return false
	}
	if math.IsInf(value, 0) {
		return true
	}

	normalizedRawNumber := convertNumberToScientificNotation(raw, false)
	requestedPrecision := len(normalizedRawNumber.coefficient)
	if requestedPrecision > 100 {
		return true
	}

	if requestedPrecision < 1 {
		requestedPrecision = 1
	}
	normalizedStoredNumber := numberToPrecisionScientific(value, requestedPrecision)

	return normalizedRawNumber.magnitude != normalizedStoredNumber.magnitude ||
		normalizedRawNumber.coefficient != normalizedStoredNumber.coefficient
}

// scientificNotation matches the upstream rule's comparison shape: coefficient
// digits with an implied decimal point after the first digit, plus magnitude.
type scientificNotation struct {
	coefficient string
	magnitude   int
}

func convertNumberToScientificNotation(stringNumber string, parseAsFloat bool) scientificNotation {
	exponentIndex := strings.IndexAny(stringNumber, "eE")
	originalCoefficient := stringNumber
	if exponentIndex >= 0 {
		originalCoefficient = stringNumber[:exponentIndex]
	}
	var normalizedNumber scientificNotation
	if parseAsFloat || strings.Contains(stringNumber, ".") {
		normalizedNumber = normalizeFloat(originalCoefficient)
	} else {
		normalizedNumber = normalizeInteger(originalCoefficient)
	}
	if exponentIndex >= 0 {
		exponent, _ := strconv.Atoi(stringNumber[exponentIndex+1:])
		normalizedNumber.magnitude += exponent
	}
	return normalizedNumber
}

func normalizeInteger(stringInteger string) scientificNotation {
	trimmedInteger := removeLeadingZeros(stringInteger)
	significantDigits := removeTrailingZeros(trimmedInteger)
	return scientificNotation{
		coefficient: significantDigits,
		magnitude:   len(trimmedInteger) - 1,
	}
}

func normalizeFloat(stringFloat string) scientificNotation {
	trimmedFloat := removeLeadingZeros(stringFloat)
	indexOfDecimalPoint := strings.Index(trimmedFloat, ".")

	switch indexOfDecimalPoint {
	case 0:
		significantDigits := removeLeadingZeros(trimmedFloat[1:])
		return scientificNotation{
			coefficient: significantDigits,
			magnitude:   len(significantDigits) - len(trimmedFloat),
		}
	case -1:
		return scientificNotation{
			coefficient: trimmedFloat,
			magnitude:   len(trimmedFloat) - 1,
		}
	default:
		return scientificNotation{
			coefficient: strings.ReplaceAll(trimmedFloat, ".", ""),
			magnitude:   indexOfDecimalPoint - 1,
		}
	}
}

func removeLeadingZeros(numberAsString string) string {
	for i := range len(numberAsString) {
		if numberAsString[i] != '0' {
			return numberAsString[i:]
		}
	}
	return numberAsString
}

func removeTrailingZeros(numberAsString string) string {
	for i := len(numberAsString) - 1; i >= 0; i-- {
		if numberAsString[i] != '0' {
			return numberAsString[:i+1]
		}
	}
	return numberAsString
}

// numberToPrecisionScientific mirrors the part of Number#toPrecision() the
// rule compares against. strconv.FormatFloat is close, but it doesn't preserve
// JS toPrecision's observable rounding on literals such as 255.10000610351562,
// so this rounds the exact float64 rational to the requested significant digit
// count before normalizing.
func numberToPrecisionScientific(value float64, precision int) scientificNotation {
	rat := new(big.Rat).SetFloat64(math.Abs(value))
	if rat == nil {
		return scientificNotation{}
	}

	magnitude := int(math.Floor(math.Log10(math.Abs(value))))
	for rat.Cmp(decimalPower(magnitude)) < 0 {
		magnitude--
	}
	for rat.Cmp(decimalPower(magnitude+1)) >= 0 {
		magnitude++
	}

	scaled := rat
	scaleExponent := precision - 1 - magnitude
	if scaleExponent >= 0 {
		scaled.Mul(scaled, decimalPower(scaleExponent))
	} else {
		scaled.Quo(scaled, decimalPower(-scaleExponent))
	}

	rounded := roundRatHalfUp(scaled)
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

func roundRatHalfUp(r *big.Rat) *big.Int {
	quotient := new(big.Int)
	remainder := new(big.Int)
	quotient.QuoRem(r.Num(), r.Denom(), remainder)

	remainder.Lsh(remainder, 1)
	if remainder.Cmp(r.Denom()) >= 0 {
		quotient.Add(quotient, big.NewInt(1))
	}
	return quotient
}

// decimalPower returns a shared, immutable power of ten for every exponent
// reachable by a finite float64 at precisions 1 through 100.
func decimalPower(exponent int) *big.Rat {
	if exponent >= minCachedDecimalExponent && exponent <= maxCachedDecimalExponent {
		entry := &decimalPowerCache[exponent-minCachedDecimalExponent]
		entry.once.Do(func() {
			entry.value = newDecimalPower(exponent)
		})
		return entry.value
	}

	// Keep the helper total for defensive direct callers, even though the rule's
	// finite-float path cannot reach this branch.
	return newDecimalPower(exponent)
}

func newDecimalPower(exponent int) *big.Rat {
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
