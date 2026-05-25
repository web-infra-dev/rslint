package no_nonoctal_decimal_escape

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// https://eslint.org/docs/latest/rules/no-nonoctal-decimal-escape
var NoNonoctalDecimalEscapeRule = rule.Rule{
	Name: "no-nonoctal-decimal-escape",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindStringLiteral: func(node *ast.Node) {
				trimmedRange := utils.TrimNodeTextRange(ctx.SourceFile, node)
				rawStart := trimmedRange.Pos()
				raw := ctx.SourceFile.Text()[rawStart:trimmedRange.End()]

				if !containsBackslashDigit89(raw) {
					return
				}

				for _, hit := range scanDecimalEscapes(raw) {
					reportDecimalEscape(ctx, rawStart, hit)
				}
			},
		}
	},
}

// decimalEscapeHit describes a single \8 or \9 occurrence found in the raw
// source text of a string literal, plus the immediately preceding `\X` escape
// pair (used to detect the special "\0\8" / "\0\9" case).
type decimalEscapeHit struct {
	previousEscape      string
	previousEscapeStart int
	decimalEscapeStart  int
	decimalEscapeEnd    int
	decimalEscape       string
}

func containsBackslashDigit89(raw string) bool {
	for i := 0; i+1 < len(raw); i++ {
		if raw[i] == '\\' && (raw[i+1] == '8' || raw[i+1] == '9') {
			return true
		}
	}
	return false
}

// scanDecimalEscapes walks raw source text and returns every \8 / \9 hit.
// Mirrors ESLint's regex
//
//	(?:[^\\]|(?<previousEscape>\\.))*?(?<decimalEscape>\\[89])
//
// where previousEscape captures the LAST `\X` pair before the decimal escape.
// In ESLint's lazy match, that capture only survives when the preceding `\X`
// is immediately adjacent to the decimal escape — any unescaped character
// between them clears it. We replicate that by resetting `previousEscape`
// whenever a non-backslash byte is consumed.
func scanDecimalEscapes(raw string) []decimalEscapeHit {
	var hits []decimalEscapeHit
	previousEscape := ""
	previousEscapeStart := -1
	n := len(raw)
	for i := 0; i < n; {
		if raw[i] != '\\' {
			i++
			previousEscape = ""
			previousEscapeStart = -1
			continue
		}
		if i+1 >= n {
			break
		}
		next := raw[i+1]
		if next == '8' || next == '9' {
			hits = append(hits, decimalEscapeHit{
				previousEscape:      previousEscape,
				previousEscapeStart: previousEscapeStart,
				decimalEscapeStart:  i,
				decimalEscapeEnd:    i + 2,
				decimalEscape:       raw[i : i+2],
			})
			i += 2
			previousEscape = ""
			previousEscapeStart = -1
			continue
		}
		previousEscape = raw[i : i+2]
		previousEscapeStart = i
		i += 2
	}
	return hits
}

func reportDecimalEscape(ctx rule.RuleContext, rawStart int, hit decimalEscapeHit) {
	decimalEscapeStartAbs := rawStart + hit.decimalEscapeStart
	decimalEscapeEndAbs := rawStart + hit.decimalEscapeEnd
	digit := hit.decimalEscape[1:]

	suggestions := make([]rule.RuleSuggestion, 0, 3)

	if hit.previousEscape == "\\0" {
		// "\0\X" — replacing with "\0X" would create a legacy octal escape, so
		// the rule offers two alternative refactors instead of the single one.
		previousEscapeStartAbs := rawStart + hit.previousEscapeStart
		nullEscape := unicodeEscape(0)
		combined := nullEscape + digit
		suggestions = append(suggestions, rule.RuleSuggestion{
			Message: rule.RuleMessage{
				Id:          "refactor",
				Description: refactorMessage("\\0"+hit.decimalEscape, combined),
			},
			FixesArr: []rule.RuleFix{
				rule.RuleFixReplaceRange(
					core.NewTextRange(previousEscapeStartAbs, decimalEscapeEndAbs),
					combined,
				),
			},
		})
		digitUnicode := unicodeEscape(rune(digit[0]))
		suggestions = append(suggestions, rule.RuleSuggestion{
			Message: rule.RuleMessage{
				Id:          "refactor",
				Description: refactorMessage(hit.decimalEscape, digitUnicode),
			},
			FixesArr: []rule.RuleFix{
				rule.RuleFixReplaceRange(
					core.NewTextRange(decimalEscapeStartAbs, decimalEscapeEndAbs),
					digitUnicode,
				),
			},
		})
	} else {
		suggestions = append(suggestions, rule.RuleSuggestion{
			Message: rule.RuleMessage{
				Id:          "refactor",
				Description: refactorMessage(hit.decimalEscape, digit),
			},
			FixesArr: []rule.RuleFix{
				rule.RuleFixReplaceRange(
					core.NewTextRange(decimalEscapeStartAbs, decimalEscapeEndAbs),
					digit,
				),
			},
		})
	}

	escaped := "\\" + hit.decimalEscape
	suggestions = append(suggestions, rule.RuleSuggestion{
		Message: rule.RuleMessage{
			Id:          "escapeBackslash",
			Description: escapeBackslashMessage(hit.decimalEscape, escaped),
		},
		FixesArr: []rule.RuleFix{
			rule.RuleFixReplaceRange(
				core.NewTextRange(decimalEscapeStartAbs, decimalEscapeEndAbs),
				escaped,
			),
		},
	})

	ctx.ReportRangeWithSuggestions(
		core.NewTextRange(decimalEscapeStartAbs, decimalEscapeEndAbs),
		rule.RuleMessage{
			Id:          "decimalEscape",
			Description: fmt.Sprintf("Don't use '%s' escape sequence.", hit.decimalEscape),
		},
		suggestions...,
	)
}

func unicodeEscape(ch rune) string {
	return fmt.Sprintf("\\u%04x", ch)
}

func refactorMessage(original, replacement string) string {
	return fmt.Sprintf("Replace '%s' with '%s'. This maintains the current functionality.", original, replacement)
}

func escapeBackslashMessage(original, replacement string) string {
	return fmt.Sprintf("Replace '%s' with '%s' to include the actual backslash character.", original, replacement)
}
