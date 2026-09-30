package prefer_unicode_code_point_escapes

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf16"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var message = rule.RuleMessage{
	Id:          "prefer-unicode-code-point-escapes",
	Description: "Prefer Unicode code point escapes.",
}

var suggestionMessage = rule.RuleMessage{
	Id:          "prefer-unicode-code-point-escapes/add-unicode-flag",
	Description: "Use Unicode code point escapes and add the `u` flag.",
}

// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v76.0.0/rules/prefer-unicode-code-point-escapes.js
var PreferUnicodeCodePointEscapesRule = rule.Rule{
	Name:   "unicorn/prefer-unicode-code-point-escapes",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		checkString := func(node *ast.Node) {
			isTemplate := ast.IsTemplateLiteralKind(node.Kind)
			if isTemplate {
				template := ast.FindAncestor(node, func(parent *ast.Node) bool {
					return parent.Kind == ast.KindNoSubstitutionTemplateLiteral || parent.Kind == ast.KindTemplateExpression || parent.Kind == ast.KindTemplateLiteralType
				})
				if template != nil && template.Parent != nil && ast.IsTaggedTemplateExpression(template.Parent) && template.Parent.AsTaggedTemplateExpression().Template == template {
					return
				}
			}
			var raw string
			switch node.Kind {
			case ast.KindTemplateHead, ast.KindTemplateMiddle, ast.KindTemplateTail:
				raw = node.RawText()
			case ast.KindNoSubstitutionTemplateLiteral:
				// Unlike template heads and spans, these nodes do not store RawText.
				raw = utils.TrimmedNodeText(ctx.SourceFile, node)
				raw = raw[1 : len(raw)-1]
			default:
				raw = utils.TrimmedNodeText(ctx.SourceFile, node)
			}
			if !strings.ContainsRune(raw, '\\') || !scanEscapes(raw, false, false, nil) {
				return
			}
			ctx.ReportNodeWithDeferredFixes(node, message, func() []rule.RuleFix {
				fixed := replaceEscapes(raw, false, false)
				if isTemplate {
					// Espree normalizes TemplateElement.value.raw line endings;
					// typescript-eslint preserves them in TypeScript sources.
					if ast.IsSourceFileJS(ctx.SourceFile) {
						fixed = strings.ReplaceAll(strings.ReplaceAll(fixed, "\r\n", "\n"), "\r", "\n")
					}
					end := node.End() - 1
					if node.Kind == ast.KindTemplateHead || node.Kind == ast.KindTemplateMiddle {
						end-- // Heads and middles end with ${.
					}
					contentRange := core.NewTextRange(end-len(raw), end)
					return []rule.RuleFix{rule.RuleFixReplaceRange(contentRange, fixed)}
				}
				return []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, node, fixed)}
			})
		}
		return rule.RuleListeners{
			ast.KindStringLiteral:                 checkString,
			ast.KindNoSubstitutionTemplateLiteral: checkString,
			ast.KindTemplateHead:                  checkString,
			ast.KindTemplateMiddle:                checkString,
			ast.KindTemplateTail:                  checkString,
			ast.KindRegularExpressionLiteral: func(node *ast.Node) {
				raw := utils.TrimmedNodeText(ctx.SourceFile, node)
				if !strings.ContainsRune(raw, '\\') {
					return
				}
				pattern, flags := utils.ExtractRegexPatternAndFlags(raw)
				unicodeFlags := utils.ParseRegexFlags(flags)
				changed := scanEscapes(pattern, true, unicodeFlags.UnicodeSets, nil)
				if !changed && (unicodeFlags.UV() || !hasCodePointEscape(pattern)) {
					return
				}
				if unicodeFlags.UV() {
					ctx.ReportNodeWithDeferredFixes(node, message, func() []rule.RuleFix {
						fixed := "/" + replaceEscapes(pattern, true, unicodeFlags.UnicodeSets) + "/" + flags
						return []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, node, fixed)}
					})
					return
				}
				ctx.ReportNodeWithDeferredSuggestions(node, message, func() []rule.RuleSuggestion {
					fixed := replaceEscapes(pattern, true, false)
					if !utils.IsValidRegexPattern(fixed, utils.RegexFlags{Unicode: true}) {
						return nil
					}
					return []rule.RuleSuggestion{{
						Message: suggestionMessage,
						FixesArr: []rule.RuleFix{
							rule.RuleFixReplace(ctx.SourceFile, node, "/"+fixed+"/"+flags+"u"),
						},
					}}
				})
			},
		}
	},
}

// scanEscapes walks ASCII escape syntax without decoding the surrounding
// source. A nil visitor only detects the first replacement, so diagnostics
// alone do not construct replacement text or edit records.
func scanEscapes(text string, isRegex, nestedClasses bool, visit func(start, end int, value rune)) bool {
	depth := 0
	found := false
	for i := 0; i < len(text); i++ {
		if isRegex {
			switch text[i] {
			case '[':
				if depth == 0 || nestedClasses {
					depth++
				}
			case ']':
				if depth > 0 {
					depth--
				}
			}
		}
		if text[i] != '\\' || i+1 == len(text) {
			continue
		}
		value, end, ok := escapeValue(text, i, isRegex, depth > 0)
		if !ok {
			// Consume the escaped character too: escaped backslashes and
			// brackets must not start escapes or change character-class depth.
			i++
			continue
		}
		if visit == nil {
			return true
		}
		visit(i, end, value)
		found = true
		i = end - 1
	}
	return found
}

func replaceEscapes(text string, isRegex, nestedClasses bool) string {
	var result strings.Builder
	previous := 0
	scanEscapes(text, isRegex, nestedClasses, func(start, end int, value rune) {
		result.WriteString(text[previous:start])
		fmt.Fprintf(&result, `\u{%X}`, value)
		previous = end
	})
	result.WriteString(text[previous:])
	return result.String()
}

func escapeValue(text string, start int, isRegex, inClass bool) (rune, int, bool) {
	switch text[start+1] {
	case 'x', 'u':
		digits := 2
		if text[start+1] == 'u' {
			digits = 4
		}
		value, ok := parseHex(text, start+2, digits)
		if !ok || isRegex && inClass && utf16.IsSurrogate(value) {
			return 0, 0, false
		}
		end := start + 2 + digits
		if !inClass && value >= 0xD800 && value <= 0xDBFF && strings.HasPrefix(text[end:], `\u`) {
			if low, ok := parseHex(text, end+2, 4); ok && low >= 0xDC00 && low <= 0xDFFF {
				return utf16.DecodeRune(value, low), end + 6, true
			}
		}
		return value, end, true
	case 'c':
		if isRegex && start+2 < len(text) {
			letter := text[start+2]
			if letter >= 'a' && letter <= 'z' || letter >= 'A' && letter <= 'Z' {
				return rune(letter % 32), start + 3, true
			}
		}
	}
	if isRegex || text[start+1] < '0' || text[start+1] > '7' {
		return 0, 0, false
	}
	if text[start+1] == '0' && (start+2 == len(text) || text[start+2] < '0' || text[start+2] > '7') {
		return 0, 0, false
	}
	maximumLength := 2
	if text[start+1] <= '3' {
		maximumLength = 3
	}
	end := start + 1
	var value rune
	for end < len(text) && end < start+1+maximumLength && text[end] >= '0' && text[end] <= '7' {
		value = value*8 + rune(text[end]-'0')
		end++
	}
	return value, end, true
}

func parseHex(text string, start, length int) (rune, bool) {
	if start+length > len(text) || !utils.AllHexDigits(text[start:start+length]) {
		return 0, false
	}
	return rune(utils.ParseHexUint(text[start : start+length])), true
}

func hasCodePointEscape(text string) bool {
	for i := 0; i < len(text); i++ {
		if text[i] != '\\' {
			continue
		}
		if strings.HasPrefix(text[i:], `\u{`) {
			end := i + 3
			for end < len(text) && utils.IsHexDigit(text[end]) {
				end++
			}
			if end > i+3 && end < len(text) && text[end] == '}' {
				if value, err := strconv.ParseUint(text[i+3:end], 16, 32); err == nil && value <= 0x10FFFF {
					return true
				}
			}
		}
		i++
	}
	return false
}
