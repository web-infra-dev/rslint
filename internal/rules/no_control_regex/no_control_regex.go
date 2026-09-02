package no_control_regex

import (
	"fmt"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

var controlCharNames = func() [0x20]string {
	var names [0x20]string
	for value := range names {
		names[value] = fmt.Sprintf(`\x%02x`, value)
	}
	return names
}()

// https://eslint.org/docs/latest/rules/no-control-regex
var NoControlRegexRule = rule.Rule{
	Name:   "no-control-regex",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		report := func(node *ast.Node, controlChars []string) {
			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "unexpected",
				Description: fmt.Sprintf("Unexpected control character(s) in regular expression: %s.", strings.Join(controlChars, ", ")),
			})
		}

		return rule.RuleListeners{
			ast.KindRegularExpressionLiteral: func(node *ast.Node) {
				pattern, flags := utils.ExtractRegexPatternAndFlags(node.Text())
				if chars := collectControlChars(pattern, flags); len(chars) > 0 {
					report(node, chars)
				}
			},
			ast.KindCallExpression: func(node *ast.Node) {
				callExpr := node.AsCallExpression()
				checkRegExpConstructor(ctx, callExpr.Expression, callExpr.Arguments, report)
			},
			ast.KindNewExpression: func(node *ast.Node) {
				newExpr := node.AsNewExpression()
				checkRegExpConstructor(ctx, newExpr.Expression, newExpr.Arguments, report)
			},
		}
	},
}

func checkRegExpConstructor(
	ctx rule.RuleContext,
	callee *ast.Node,
	args *ast.NodeList,
	report func(*ast.Node, []string),
) {
	callee = ast.SkipParentheses(callee)
	if callee == nil || callee.Kind != ast.KindIdentifier || callee.AsIdentifier().Text != "RegExp" {
		return
	}
	if !ctx.Globals.Access("RegExp").IsDeclared() {
		return
	}
	if ctx.Refs != nil {
		if !ctx.Refs.IsGlobalReference(callee) {
			return
		}
	} else if utils.IsShadowed(callee, "RegExp") {
		return
	}
	if args == nil || len(args.Nodes) == 0 {
		return
	}

	patternNode := ast.SkipParentheses(args.Nodes[0])
	if patternNode == nil || patternNode.Kind != ast.KindStringLiteral {
		return
	}
	pattern, ok := stringLiteralValue(ctx, patternNode)
	if !ok {
		return
	}

	// ESLint only treats flags as known when the second argument is a string
	// literal; otherwise flags default to "" (neither u nor v).
	flags := ""
	if len(args.Nodes) >= 2 {
		flagsNode := ast.SkipParentheses(args.Nodes[1])
		if flagsNode != nil && flagsNode.Kind == ast.KindStringLiteral {
			if value, valueOK := stringLiteralValue(ctx, flagsNode); valueOK {
				flags = value
			}
		}
	}

	if chars := collectControlChars(pattern, flags); len(chars) > 0 {
		report(patternNode, chars)
	}
}

func stringLiteralValue(ctx rule.RuleContext, node *ast.Node) (string, bool) {
	if utils.IsInStrictModeWithSourceType(node, ctx.SourceFile, ctx.LanguageOptions.EffectiveSourceType()) {
		return node.AsStringLiteral().Text, true
	}
	r := utils.TrimNodeTextRange(ctx.SourceFile, node)
	units := utils.ParseJSStringLiteralSource(ctx.SourceFile.Text()[r.Pos():r.End()])
	if units == nil {
		return "", false
	}
	values := make([]uint16, len(units))
	for i, unit := range units {
		values[i] = uint16(unit.Value)
	}
	return ecmascript.StringFromCodeUnits(values), true
}

// collectControlChars scans a regex pattern and returns each code point in
// U+0000..U+001F that appears as:
//   - a raw character,
//   - a \xHH escape,
//   - a \uHHHH escape, or
//   - a \u{H...} escape (only under the u or v flag).
//
// Each hit is formatted as `\xHH` (lowercase, 2 digits). Symbolic control
// escapes (\t, \n, \r, \v, \f, \0, \cX) are NOT reported — matching ESLint.
func collectControlChars(pattern, flagsStr string) []string {
	pattern = ecmascript.CombineSurrogatePairs(pattern)
	flags := utils.ParseRegexFlags(flagsStr)
	uvMode := flags.UV()
	results, uncertainEscape := scanControlChars(pattern, uvMode)
	if len(results) == 0 {
		return nil
	}
	if (!flags.Unicode || !flags.UnicodeSets) && !uncertainEscape && !strings.ContainsAny(pattern, `^$.*+?()[]{|}`) {
		return results
	}
	if errorPosition, invalid := utils.RegexPatternCharacterEventCutoff(pattern, flags); invalid {
		if errorPosition <= 0 {
			return nil
		}
		if errorPosition < len(pattern) {
			results, _ = scanControlChars(pattern[:errorPosition], uvMode)
		}
	}
	return results
}

// scanControlChars performs the character-callback part of regexpp's walk.
func scanControlChars(pattern string, uvMode bool) (results []string, uncertainEscape bool) {
	record := func(cp uint32) {
		results = append(results, controlCharNames[cp])
	}

	i := 0
	for i < len(pattern) {
		c := pattern[i]

		if c == '\\' && i+1 < len(pattern) {
			next := pattern[i+1]
			switch next {
			case 'x':
				// \xHH — 2 hex digits required.
				if i+3 < len(pattern) && utils.IsHexDigit(pattern[i+2]) && utils.IsHexDigit(pattern[i+3]) {
					cp := utils.ParseHexUint(pattern[i+2 : i+4])
					if cp <= 0x1f {
						record(cp)
					}
					i += 4
					continue
				}
				uncertainEscape = true
			case 'u':
				// \u{H...} — only recognized under u / v flag.
				if uvMode && i+2 < len(pattern) && pattern[i+2] == '{' {
					if closeRel := strings.IndexByte(pattern[i+3:], '}'); closeRel > 0 {
						hex := pattern[i+3 : i+3+closeRel]
						if utils.AllHexDigits(hex) {
							significant := strings.TrimLeft(hex, "0")
							if len(significant) <= 2 {
								cp := utils.ParseHexUint(hex)
								if cp <= 0x1f {
									record(cp)
								}
							}
							i += 3 + closeRel + 1
							continue
						}
					}
				}
				// \uHHHH — 4 hex digits.
				if i+5 < len(pattern) && utils.AllHexDigits(pattern[i+2:i+6]) {
					cp := utils.ParseHexUint(pattern[i+2 : i+6])
					if cp <= 0x1f {
						record(cp)
					}
					i += 6
					continue
				}
				uncertainEscape = true
			default:
				switch next {
				case 'f', 'n', 'r', 't', 'v', 'd', 'D', 's', 'S', 'w', 'W',
					'^', '$', '\\', '.', '*', '+', '?', '(', ')', '[', ']', '{', '}', '|', '/':
				case '0':
					uncertainEscape = uncertainEscape ||
						uvMode && i+2 < len(pattern) && pattern[i+2] >= '0' && pattern[i+2] <= '9'
				case 'c':
					valid := i+2 < len(pattern) &&
						((pattern[i+2] >= 'a' && pattern[i+2] <= 'z') || (pattern[i+2] >= 'A' && pattern[i+2] <= 'Z'))
					uncertainEscape = uncertainEscape || uvMode && !valid
				default:
					uncertainEscape = uncertainEscape || uvMode
				}
			}
			// Any other escape (\t, \n, \\, \cI, \0, \d, \p{…}, \q{…}, etc.):
			// consume 2 bytes. Multi-byte escape bodies (e.g. `\q{…}` under v
			// flag) are NOT skipped as a unit — their contents may legitimately
			// contain nested escapes (`\u{1F}` etc.) that should be detected by
			// the subsequent loop iterations.
			i += 2
			continue
		}
		if c <= 0x1f {
			record(uint32(c))
		}
		if c == '\\' {
			uncertainEscape = true
		}
		i++
	}
	return results, uncertainEscape
}
