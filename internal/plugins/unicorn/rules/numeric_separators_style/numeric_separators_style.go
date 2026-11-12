// Ported from eslint-plugin-unicorn v79.0.0. See LICENSE.
package numeric_separators_style

import (
	_ "embed"
	"math"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed numeric_separators_style.schema.json
var schemaJSON []byte

var invalidGroupLength = rule.RuleMessage{
	Id:          "numeric-separators-style",
	Description: "Invalid group length in numeric value.",
}

var NumericSeparatorsStyleRule = rule.Rule{
	Name:   "unicorn/numeric-separators-style",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		formats := parseOptions(options)
		source := ctx.SourceFile.Text()
		allowsSeparators := ctx.LanguageOptions.EffectiveECMAVersion() >= 2021
		checkLiteral := func(node *ast.Node, flags ast.TokenFlags) {
			// The scanner marks legacy octal and leading-zero decimals as invalid.
			// Inserting separators into those spellings can produce invalid JavaScript.
			if flags&ast.TokenFlagsIsInvalid != 0 {
				return
			}
			format := formats.groupingFor(flags)
			if format.onlyIfContainsSeparator && flags&ast.TokenFlagsContainsSeparator == 0 {
				return
			}
			start, end := scanner.GetTokenPosOfNode(node, ctx.SourceFile, false), node.End()
			raw := source[start:end]
			formatted := formatLiteral(raw, format, flags&ast.TokenFlagsWithSpecifier != 0)
			if formatted == raw {
				return
			}
			span := core.NewTextRange(start, end)
			ctx.ReportRangeWithDeferredFixes(span, invalidGroupLength, func() []rule.RuleFix {
				if !allowsSeparators && strings.Contains(formatted, "_") {
					return nil
				}
				return []rule.RuleFix{rule.RuleFixReplaceRange(span, formatted)}
			})
		}
		return rule.RuleListeners{
			ast.KindNumericLiteral: func(node *ast.Node) {
				checkLiteral(node, node.AsNumericLiteral().TokenFlags)
			},
			ast.KindBigIntLiteral: func(node *ast.Node) {
				checkLiteral(node, node.AsBigIntLiteral().TokenFlags)
			},
		}
	},
}

type grouping struct {
	onlyIfContainsSeparator bool
	minimumDigits           int
	groupLength             int
	fractionGroupLength     int
}

type numberFormats struct {
	binary, octal, hexadecimal, decimal grouping
}

func parseOptions(options []any) numberFormats {
	var config map[string]any
	if len(options) > 0 {
		config, _ = options[0].(map[string]any)
	}
	onlyIfContainsSeparator, _ := config["onlyIfContainsSeparator"].(bool)
	return numberFormats{
		binary:      parseGrouping(config["binary"], onlyIfContainsSeparator, 0, 4),
		octal:       parseGrouping(config["octal"], onlyIfContainsSeparator, 0, 4),
		hexadecimal: parseGrouping(config["hexadecimal"], onlyIfContainsSeparator, 0, 2),
		decimal:     parseGrouping(config["number"], onlyIfContainsSeparator, 5, 3),
	}
}

func parseGrouping(value any, onlyIfContainsSeparator bool, minimumDigits, groupLength int) grouping {
	config, _ := value.(map[string]any)
	if override, ok := config["onlyIfContainsSeparator"].(bool); ok {
		onlyIfContainsSeparator = override
	}
	return grouping{
		onlyIfContainsSeparator: onlyIfContainsSeparator,
		minimumDigits:           digitLimit(config["minimumDigits"], minimumDigits),
		groupLength:             digitLimit(config["groupLength"], groupLength),
		fractionGroupLength:     digitLimit(config["fractionGroupLength"], math.MaxInt),
	}
}

func digitLimit(value any, fallback int) int {
	// The schema permits limits larger than any Go string.
	if number, ok := value.(float32); ok {
		value = float64(number)
	}
	if number, ok := value.(float64); ok && number >= float64(math.MaxInt) {
		return math.MaxInt
	}
	if number, ok := utils.CoerceIntegral(value); ok {
		return number
	}
	return fallback
}

func (formats numberFormats) groupingFor(flags ast.TokenFlags) grouping {
	switch flags & ast.TokenFlagsWithSpecifier {
	case ast.TokenFlagsBinarySpecifier:
		return formats.binary
	case ast.TokenFlagsOctalSpecifier:
		return formats.octal
	case ast.TokenFlagsHexSpecifier:
		return formats.hexadecimal
	default:
		return formats.decimal
	}
}

func formatLiteral(raw string, format grouping, hasRadixPrefix bool) string {
	value := strings.TrimSuffix(raw, "n")
	suffix := raw[len(value):]
	if hasRadixPrefix {
		digits := value[2:]
		grouped := groupDigits(strings.ReplaceAll(digits, "_", ""), format.minimumDigits, format.groupLength, false)
		if grouped == digits {
			return raw
		}
		return value[:2] + grouped + suffix
	}
	return formatDecimal(strings.ReplaceAll(value, "_", ""), format) + suffix
}

func formatDecimal(value string, format grouping) string {
	mantissa, exponent := value, ""
	if mark := strings.IndexAny(value, "eE"); mark >= 0 {
		power := mark + 1
		if power < len(value) && (value[power] == '+' || value[power] == '-') {
			power++
		}
		mantissa = value[:mark]
		exponent = value[mark:power] + groupDigits(value[power:], format.minimumDigits, format.groupLength, false)
	}
	integer, fractional, hasDot := strings.Cut(mantissa, ".")
	integer = groupDigits(integer, format.minimumDigits, format.groupLength, false)
	if hasDot {
		integer += "." + groupDigits(fractional, format.minimumDigits, format.fractionGroupLength, true)
	}
	return integer + exponent
}

func groupDigits(digits string, minimumDigits, groupLength int, fromLeft bool) string {
	if len(digits) < minimumDigits || len(digits) <= groupLength {
		return digits
	}
	firstGroup := groupLength
	if remainder := len(digits) % groupLength; !fromLeft && remainder != 0 {
		firstGroup = remainder
	}
	var grouped strings.Builder
	grouped.Grow(len(digits) + (len(digits)-1)/groupLength)
	grouped.WriteString(digits[:firstGroup])
	for start := firstGroup; start < len(digits); start += groupLength {
		grouped.WriteByte('_')
		grouped.WriteString(digits[start:min(start+groupLength, len(digits))])
	}
	return grouped.String()
}
