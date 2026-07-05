package no_loss_of_precision

import (
	"math"
	"math/big"
	"regexp"
	"strconv"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var noLossOfPrecisionMessage = rule.RuleMessage{
	Id:          "noLossOfPrecision",
	Description: "This number literal will lose precision at runtime.",
}

var (
	binaryPattern      = regexp.MustCompile(`(?i)^0b[01]+$`)
	octalPattern       = regexp.MustCompile(`(?i)^0o[0-7]+$`)
	hexPattern         = regexp.MustCompile(`(?i)^0x[0-9a-f]+$`)
	legacyOctalPattern = regexp.MustCompile(`^0[0-7]+$`)
)

// https://eslint.org/docs/latest/rules/no-loss-of-precision
var NoLossOfPrecisionRule = rule.Rule{
	Name: "no-loss-of-precision",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindNumericLiteral: func(node *ast.Node) {
				// tsgo normalizes NumericLiteral.Text at parse time, so ESLint parity
				// requires reading the raw token text to preserve prefixes, separators,
				// exponent spelling, and trailing fractional zeros.
				raw := utils.TrimmedNodeText(ctx.SourceFile, node)
				if losesPrecision(raw) {
					ctx.ReportNode(node, noLossOfPrecisionMessage)
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

	if binaryPattern.MatchString(normalized) {
		return notBaseTenLosesPrecision(normalized[2:], 2)
	}
	if octalPattern.MatchString(normalized) {
		return notBaseTenLosesPrecision(normalized[2:], 8)
	}
	if hexPattern.MatchString(normalized) {
		return notBaseTenLosesPrecision(normalized[2:], 16)
	}
	if legacyOctalPattern.MatchString(normalized) {
		return notBaseTenLosesPrecision(normalized[1:], 8)
	}

	return baseTenLosesPrecision(normalized)
}

func notBaseTenLosesPrecision(digits string, base int) bool {
	original := new(big.Int)
	_, ok := original.SetString(strings.ToLower(digits), base)
	if !ok {
		return false
	}

	f, _ := new(big.Float).SetInt(original).Float64()
	if math.IsInf(f, 0) {
		return true
	}

	reconstructed := new(big.Int)
	bf := new(big.Float).SetFloat64(f)
	bf.Int(reconstructed)

	return original.Cmp(reconstructed) != 0
}

func baseTenLosesPrecision(raw string) bool {
	rawNumber := strings.ToLower(raw)
	value, err := strconv.ParseFloat(rawNumber, 64)
	if err != nil && !math.IsInf(value, 0) {
		return false
	}
	if value == 0 {
		return false
	}
	if math.IsInf(value, 0) {
		return true
	}

	normalizedRawNumber := convertNumberToScientificNotation(rawNumber, false)
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
	splitNumber := strings.Split(stringNumber, "e")
	originalCoefficient := splitNumber[0]
	var normalizedNumber scientificNotation
	if parseAsFloat || strings.Contains(stringNumber, ".") {
		normalizedNumber = normalizeFloat(originalCoefficient)
	} else {
		normalizedNumber = normalizeInteger(originalCoefficient)
	}
	if len(splitNumber) > 1 {
		exponent, _ := strconv.Atoi(splitNumber[1])
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
	for rat.Cmp(pow10Rat(magnitude)) < 0 {
		magnitude--
	}
	for rat.Cmp(pow10Rat(magnitude+1)) >= 0 {
		magnitude++
	}

	scaled := new(big.Rat).Set(rat)
	scaleExponent := precision - 1 - magnitude
	if scaleExponent >= 0 {
		scaled.Mul(scaled, new(big.Rat).SetInt(pow10Int(scaleExponent)))
	} else {
		scaled.Quo(scaled, new(big.Rat).SetInt(pow10Int(-scaleExponent)))
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

	doubleRemainder := new(big.Int).Mul(remainder, big.NewInt(2))
	if doubleRemainder.Cmp(r.Denom()) >= 0 {
		quotient.Add(quotient, big.NewInt(1))
	}
	return quotient
}

func pow10Rat(exponent int) *big.Rat {
	if exponent >= 0 {
		return new(big.Rat).SetInt(pow10Int(exponent))
	}
	return new(big.Rat).SetFrac(big.NewInt(1), pow10Int(-exponent))
}

func pow10Int(exponent int) *big.Int {
	return new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(exponent)), nil)
}
