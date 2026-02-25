package no_loss_of_precision

import (
	"math"
	"math/big"
	"regexp"
	"strconv"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

var ruleMessage = rule.RuleMessage{
	Id:          "no-loss-of-precision",
	Description: "This number literal will lose precision at runtime.",
}

// Regex patterns for parsing numbers
var (
	binaryPattern      = regexp.MustCompile(`(?i)^0b[01]+$`)
	octalPattern       = regexp.MustCompile(`(?i)^0o[0-7]+$`)
	hexPattern         = regexp.MustCompile(`(?i)^0x[0-9a-f]+$`)
	legacyOctalPattern = regexp.MustCompile(`^0[0-7]+$`)
)

// NoLossOfPrecisionRule disallows literal numbers that lose precision
// https://eslint.org/docs/latest/rules/no-loss-of-precision
var NoLossOfPrecisionRule = rule.Rule{
	Name: "no-loss-of-precision",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindNumericLiteral: func(node *ast.Node) {
				// Get the raw text directly from source
				// We can't use numLiteral.Text because the parser normalizes numbers
				start := node.Pos()
				end := node.End()
				text := ctx.SourceFile.Text()

				if end <= start || end > len(text) {
					return
				}

				raw := strings.TrimSpace(text[start:end])
				if raw == "" {
					return
				}

				if losesPrecision(raw) {
					ctx.ReportNode(node, ruleMessage)
				}
			},
		}
	},
}

// removeNumericSeparators removes underscore separators from numeric literals
func removeNumericSeparators(s string) string {
	return strings.ReplaceAll(s, "_", "")
}

// losesPrecision checks if the given numeric literal loses precision
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

// notBaseTenLosesPrecision checks if a non-base-10 number loses precision
func notBaseTenLosesPrecision(digits string, base int) bool {
	// Parse as big.Int for arbitrary precision
	original := new(big.Int)
	_, ok := original.SetString(strings.ToLower(digits), base)
	if !ok {
		return false
	}

	// Convert to float64 (JavaScript Number)
	f, _ := new(big.Float).SetInt(original).Float64()

	// Check for infinity
	if math.IsInf(f, 0) {
		return true
	}

	// Convert float64 back to big.Int
	reconstructed := new(big.Int)
	bf := new(big.Float).SetFloat64(f)
	bf.Int(reconstructed)

	// Compare: if they differ, precision was lost
	return original.Cmp(reconstructed) != 0
}

// baseTenLosesPrecision checks if a base-10 number loses precision
func baseTenLosesPrecision(raw string) bool {
	// Parse using big.Float for arbitrary precision
	rawBigFloat, _, err := new(big.Float).SetPrec(256).Parse(raw, 10)
	if err != nil {
		return false
	}

	// Convert to float64 (this is what JavaScript does)
	jsValue, _ := rawBigFloat.Float64()

	// Check for infinity
	if math.IsInf(jsValue, 0) {
		return true
	}

	// Get significant info from raw
	rawSigDigits, rawExp, rawPrecision := getSignificantInfo(raw)

	// If raw precision exceeds JavaScript's ~17 significant digits, it's precision loss
	if rawPrecision > 17 {
		return true
	}

	// Get JS representation
	precision := rawPrecision
	if precision < 1 {
		precision = 1
	}
	jsStr := strconv.FormatFloat(jsValue, 'e', precision-1, 64)
	jsSigDigits, jsExp, _ := getSignificantInfo(jsStr)

	rawAbs := strings.TrimPrefix(rawSigDigits, "-")
	jsAbs := strings.TrimPrefix(jsSigDigits, "-")

	// Check sign
	if strings.HasPrefix(rawSigDigits, "-") != strings.HasPrefix(jsSigDigits, "-") {
		return true
	}

	// Handle zero
	if rawAbs == "0" {
		return jsAbs != "0"
	}

	// Check exponent
	if rawExp != jsExp {
		return true
	}

	// Compare significant digits (trimmed)
	if len(rawAbs) <= len(jsAbs) {
		return !strings.HasPrefix(jsAbs, rawAbs)
	}

	// rawAbs is longer than jsAbs
	if !strings.HasPrefix(rawAbs, jsAbs) {
		return true
	}
	extra := rawAbs[len(jsAbs):]
	return strings.Trim(extra, "0") != ""
}

// getSignificantInfo extracts significant digits, exponent, and precision from a number string
func getSignificantInfo(raw string) (sigDigits string, exp int, rawPrecision int) {
	negative := false
	if strings.HasPrefix(raw, "-") {
		negative = true
		raw = raw[1:]
	} else if strings.HasPrefix(raw, "+") {
		raw = raw[1:]
	}

	// Handle scientific notation
	var mantissa string
	expOffset := 0
	if idx := strings.IndexAny(raw, "eE"); idx >= 0 {
		mantissa = raw[:idx]
		expOffset, _ = strconv.Atoi(raw[idx+1:])
	} else {
		mantissa = raw
	}

	// Split by decimal point
	var intPart, fracPart string
	if dotIdx := strings.Index(mantissa, "."); dotIdx >= 0 {
		intPart = mantissa[:dotIdx]
		fracPart = mantissa[dotIdx+1:]
	} else {
		intPart = mantissa
		fracPart = ""
	}

	// Combine all digits
	allDigits := intPart + fracPart

	// Find first non-zero digit
	firstNonZero := 0
	for firstNonZero < len(allDigits) && allDigits[firstNonZero] == '0' {
		firstNonZero++
	}

	if firstNonZero == len(allDigits) {
		return "0", 0, 1
	}

	// Calculate exponent
	exp = len(intPart) - firstNonZero - 1 + expOffset

	// Get significant digits
	sigDigitsWithZeros := allDigits[firstNonZero:]
	sigDigits = strings.TrimRight(sigDigitsWithZeros, "0")
	if sigDigits == "" {
		sigDigits = "0"
	}

	// Check if trailing zeros in fractional part represent precision
	// This happens when there's a non-zero digit in the fractional part
	// followed by trailing zeros (e.g., ".1230000...")
	fracPartTrimmed := strings.TrimRight(fracPart, "0")
	fracHasNonZero := len(strings.TrimLeft(fracPart, "0")) > 0

	if fracHasNonZero && len(fracPart) > len(fracPartTrimmed) {
		// Trailing zeros after non-zero digit in fractional part = precision intent
		rawPrecision = len(sigDigitsWithZeros)
	} else {
		// No precision intent from trailing zeros
		rawPrecision = len(sigDigits)
	}

	if negative {
		sigDigits = "-" + sigDigits
	}

	return sigDigits, exp, rawPrecision
}
