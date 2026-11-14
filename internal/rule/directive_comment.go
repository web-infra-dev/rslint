package rule

import (
	"strings"
	"unicode/utf8"

	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

// These helpers scan inline directive comments. Each directive owns its value
// syntax, while descriptions and whitespace follow the same ECMAScript rules.

// mayContainDirective reports whether text could contain a block comment whose
// content starts with one of keywords. It is a cheap source-text scan that lets
// a parser skip asking the shared comment store for the files that cannot carry
// the directive at all; a false positive only costs the real parse.
func mayContainDirective(text string, keywords []string) bool {
	for searchStart := 0; searchStart < len(text); {
		markerOffset := strings.Index(text[searchStart:], "/*")
		if markerOffset < 0 {
			return false
		}

		contentStart := searchStart + markerOffset + len("/*")
		contentStart, _ = trimECMAScriptWhitespaceRange(text, contentStart, len(text))
		for _, keyword := range keywords {
			if !strings.HasPrefix(text[contentStart:], keyword) {
				continue
			}
			restStart := contentStart + len(keyword)
			if restStart == len(text) || strings.HasPrefix(text[restStart:], "*/") {
				return true
			}
			r, _ := utf8.DecodeRuneInString(text[restStart:])
			if ecmascript.IsWhiteSpaceOrLineTerminator(r) {
				return true
			}
		}

		searchStart = contentStart
	}
	return false
}

// matchDirectiveLabelRange reports the offset just past the first of keywords
// that text[start:end] begins with. ESLint's directivesPattern anchors the label
// at the start of the comment content and requires ECMAScript whitespace or the
// end of the content after it, so a `/* globally */` comment is not a directive.
func matchDirectiveLabelRange(text string, start int, end int, keywords []string) (int, bool) {
	for _, keyword := range keywords {
		if _, ok := cutDirectiveLabel(text[start:end], keyword); ok {
			return start + len(keyword), true
		}
	}
	return 0, false
}

func cutDirectiveLabel(text string, label string) (string, bool) {
	value, ok := strings.CutPrefix(text, label)
	if !ok || value == "" {
		return value, ok
	}
	if value[0] < utf8.RuneSelf {
		return value, ecmascript.IsTriviaWhitespaceByte(value[0])
	}
	next, _ := utf8.DecodeRuneInString(value)
	return value, ecmascript.IsTriviaWhitespaceRune(next)
}

// findDirectiveJustification returns the offset where a ` -- justification `
// tail begins within text[start:end], or -1 when there is none. ESLint splits
// on /\s-{2,}\s/u and keeps only the part before the match.
func findDirectiveJustification(text string, start int, end int) int {
	for from := start; from < end; {
		offset := strings.Index(text[from:end], "--")
		if offset < 0 {
			break
		}
		hyphenStart := from + offset
		afterHyphens := hyphenStart + 2
		for afterHyphens < end && text[afterHyphens] == '-' {
			afterHyphens++
		}
		before, size := utf8.DecodeLastRuneInString(text[start:hyphenStart])
		after, _ := utf8.DecodeRuneInString(text[afterHyphens:end])
		if ecmascript.IsWhiteSpaceOrLineTerminator(before) && ecmascript.IsWhiteSpaceOrLineTerminator(after) {
			return hyphenStart - size
		}
		from = afterHyphens
	}
	return -1
}

func trimECMAScriptWhitespaceRange(text string, start int, end int) (int, int) {
	start += ecmascript.SkipLeadingWhitespace(text[start:end], 0, end-start)
	return start, start + ecmascript.SkipTrailingWhitespace(text[start:end], 0, end-start)
}

// stripMatchingQuoteRange removes one matching layer of single or double
// quotes, the unwrapping parseListConfig performs on every entry. An unbalanced
// or lone quote is part of the name.
func stripMatchingQuoteRange(text string, start int, end int) (int, int) {
	if end-start >= 2 {
		first, last := text[start], text[end-1]
		if (first == '\'' || first == '"') && first == last {
			return start + 1, end - 1
		}
	}
	return start, end
}
