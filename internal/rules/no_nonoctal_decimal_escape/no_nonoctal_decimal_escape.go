package no_nonoctal_decimal_escape

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// https://eslint.org/docs/latest/rules/no-nonoctal-decimal-escape
var NoNonoctalDecimalEscapeRule = rule.Rule{
	Name: "no-nonoctal-decimal-escape",
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		sourceText := ctx.SourceFile.Text()
		return rule.RuleListeners{
			ast.KindStringLiteral: func(node *ast.Node) {
				// Ordinary string literals containing \8 or \9 carry this flag.
				// JSX attribute strings are scanned in verbatim mode, so their token
				// flags do not describe escapes and they need the raw-text fallback.
				literal := node.AsStringLiteral()
				if literal.TokenFlags&ast.TokenFlagsContainsInvalidEscape == 0 &&
					(node.Parent == nil || node.Parent.Kind != ast.KindJsxAttribute) {
					return
				}

				rawStart := scanner.SkipTrivia(sourceText, node.Pos())
				rawEnd := node.End()
				if rawStart < 0 || rawStart >= rawEnd || rawEnd > len(sourceText) {
					return
				}

				decimalEscapes := newDecimalEscapeIterator(sourceText[rawStart:rawEnd])
				for hit, ok := decimalEscapes.next(); ok; hit, ok = decimalEscapes.next() {
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

type decimalEscapeIterator struct {
	raw                 string
	offset              int
	previousEscapeStart int
}

func newDecimalEscapeIterator(raw string) decimalEscapeIterator {
	return decimalEscapeIterator{raw: raw, previousEscapeStart: -1}
}

// next walks raw source text and returns the next \8 / \9 hit.
// Mirrors ESLint's regex
//
//	(?:[^\\]|(?<previousEscape>\\.))*?(?<decimalEscape>\\[89])
//
// where previousEscape captures the LAST `\X` pair before the decimal escape.
// In ESLint's lazy match, that capture only survives when the preceding `\X`
// is immediately adjacent to the decimal escape — any unescaped character
// between them clears it. We replicate that by resetting `previousEscape`
// whenever a non-backslash byte is consumed.
func (s *decimalEscapeIterator) next() (decimalEscapeHit, bool) {
	for s.offset < len(s.raw) {
		if s.raw[s.offset] != '\\' {
			s.offset++
			s.previousEscapeStart = -1
			continue
		}
		if s.offset+1 >= len(s.raw) {
			break
		}

		escapeStart := s.offset
		next := s.raw[escapeStart+1]
		if next == '8' || next == '9' {
			previousEscape := ""
			if s.previousEscapeStart >= 0 {
				previousEscape = s.raw[s.previousEscapeStart : s.previousEscapeStart+2]
			}
			hit := decimalEscapeHit{
				previousEscape:      previousEscape,
				previousEscapeStart: s.previousEscapeStart,
				decimalEscapeStart:  escapeStart,
				decimalEscapeEnd:    escapeStart + 2,
				decimalEscape:       s.raw[escapeStart : escapeStart+2],
			}
			s.offset = escapeStart + 2
			s.previousEscapeStart = -1
			return hit, true
		}

		s.previousEscapeStart = escapeStart
		s.offset = escapeStart + 2
	}
	return decimalEscapeHit{}, false
}

func reportDecimalEscape(ctx rule.RuleContext, rawStart int, hit decimalEscapeHit) {
	decimalEscapeStartAbs := rawStart + hit.decimalEscapeStart
	decimalEscapeEndAbs := rawStart + hit.decimalEscapeEnd
	text := decimalEscapeTextFor(hit.decimalEscape[1])
	decimalEscapeRange := core.NewTextRange(decimalEscapeStartAbs, decimalEscapeEndAbs)

	ctx.ReportRangeWithDeferredSuggestions(
		decimalEscapeRange,
		text.diagnostic,
		func() []rule.RuleSuggestion {
			return buildDecimalEscapeSuggestions(rawStart, decimalEscapeRange, hit, text)
		},
	)
}

type decimalEscapeText struct {
	digit                  string
	digitUnicode           string
	nullAndDigitUnicode    string
	escaped                string
	diagnostic             rule.RuleMessage
	refactor               rule.RuleMessage
	refactorNullAndDigit   rule.RuleMessage
	refactorDigitAsUnicode rule.RuleMessage
	escapeBackslash        rule.RuleMessage
}

var decimalEscapeTexts = [...]decimalEscapeText{
	{
		digit:               "8",
		digitUnicode:        `\u0038`,
		nullAndDigitUnicode: `\u00008`,
		escaped:             `\\8`,
		diagnostic: rule.RuleMessage{
			Id:          "decimalEscape",
			Description: `Don't use '\8' escape sequence.`,
		},
		refactor: rule.RuleMessage{
			Id:          "refactor",
			Description: `Replace '\8' with '8'. This maintains the current functionality.`,
		},
		refactorNullAndDigit: rule.RuleMessage{
			Id:          "refactor",
			Description: `Replace '\0\8' with '\u00008'. This maintains the current functionality.`,
		},
		refactorDigitAsUnicode: rule.RuleMessage{
			Id:          "refactor",
			Description: `Replace '\8' with '\u0038'. This maintains the current functionality.`,
		},
		escapeBackslash: rule.RuleMessage{
			Id:          "escapeBackslash",
			Description: `Replace '\8' with '\\8' to include the actual backslash character.`,
		},
	},
	{
		digit:               "9",
		digitUnicode:        `\u0039`,
		nullAndDigitUnicode: `\u00009`,
		escaped:             `\\9`,
		diagnostic: rule.RuleMessage{
			Id:          "decimalEscape",
			Description: `Don't use '\9' escape sequence.`,
		},
		refactor: rule.RuleMessage{
			Id:          "refactor",
			Description: `Replace '\9' with '9'. This maintains the current functionality.`,
		},
		refactorNullAndDigit: rule.RuleMessage{
			Id:          "refactor",
			Description: `Replace '\0\9' with '\u00009'. This maintains the current functionality.`,
		},
		refactorDigitAsUnicode: rule.RuleMessage{
			Id:          "refactor",
			Description: `Replace '\9' with '\u0039'. This maintains the current functionality.`,
		},
		escapeBackslash: rule.RuleMessage{
			Id:          "escapeBackslash",
			Description: `Replace '\9' with '\\9' to include the actual backslash character.`,
		},
	},
}

func decimalEscapeTextFor(digit byte) *decimalEscapeText {
	if digit == '8' {
		return &decimalEscapeTexts[0]
	}
	return &decimalEscapeTexts[1]
}

type decimalEscapeSuggestionBundle struct {
	fixes       [3]rule.RuleFix
	suggestions [3]rule.RuleSuggestion
}

func (b *decimalEscapeSuggestionBundle) set(index int, message rule.RuleMessage, textRange core.TextRange, text string) {
	b.fixes[index] = rule.RuleFixReplaceRange(textRange, text)
	b.suggestions[index] = rule.RuleSuggestion{
		Message:  message,
		FixesArr: b.fixes[index : index+1 : index+1],
	}
}

func buildDecimalEscapeSuggestions(
	rawStart int,
	decimalEscapeRange core.TextRange,
	hit decimalEscapeHit,
	text *decimalEscapeText,
) []rule.RuleSuggestion {
	bundle := &decimalEscapeSuggestionBundle{}
	if hit.previousEscape == "\\0" {
		// "\0\X" — replacing with "\0X" would create a legacy octal escape, so
		// the rule offers two alternative refactors instead of the single one.
		bundle.set(
			0,
			text.refactorNullAndDigit,
			core.NewTextRange(rawStart+hit.previousEscapeStart, decimalEscapeRange.End()),
			text.nullAndDigitUnicode,
		)
		bundle.set(1, text.refactorDigitAsUnicode, decimalEscapeRange, text.digitUnicode)
		bundle.set(2, text.escapeBackslash, decimalEscapeRange, text.escaped)
		return bundle.suggestions[:3:3]
	}

	bundle.set(0, text.refactor, decimalEscapeRange, text.digit)
	bundle.set(1, text.escapeBackslash, decimalEscapeRange, text.escaped)
	return bundle.suggestions[:2:2]
}
