package rule

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// Scanning shared by ESLint's block directive comments (`/* global */`,
// `/* exported */`). Each directive owns its own value syntax; what they share
// is locating the label, dropping a `-- justification` tail, and ESLint's exact
// whitespace class.

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
			if isECMAScriptWhitespace(r) {
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
		if !strings.HasPrefix(text[start:end], keyword) {
			continue
		}
		restStart := start + len(keyword)
		if restStart == end {
			return restStart, true
		}
		r, _ := utf8.DecodeRuneInString(text[restStart:end])
		if isECMAScriptWhitespace(r) {
			return restStart, true
		}
	}
	return 0, false
}

// findDirectiveJustification returns the offset where a ` -- justification `
// tail begins within text[start:end], or -1 when there is none. ESLint splits
// on /\s-{2,}\s/u and keeps only the part before the match.
func findDirectiveJustification(text string, start int, end int) int {
	for index := start; index < end; {
		r, size := utf8.DecodeRuneInString(text[index:end])
		if !isECMAScriptWhitespace(r) {
			index += size
			continue
		}

		hyphenStart := index + size
		afterHyphens := hyphenStart
		for afterHyphens < end && text[afterHyphens] == '-' {
			afterHyphens++
		}
		if afterHyphens-hyphenStart >= 2 && afterHyphens < end {
			next, _ := utf8.DecodeRuneInString(text[afterHyphens:end])
			if isECMAScriptWhitespace(next) {
				return index
			}
		}
		index += size
	}
	return -1
}

func trimECMAScriptWhitespaceRange(text string, start int, end int) (int, int) {
	for start < end {
		r, size := utf8.DecodeRuneInString(text[start:end])
		if !isECMAScriptWhitespace(r) {
			break
		}
		start += size
	}
	for end > start {
		r, size := utf8.DecodeLastRuneInString(text[start:end])
		if !isECMAScriptWhitespace(r) {
			break
		}
		end -= size
	}
	return start, end
}

// ECMAScript's \s set is Unicode Zs plus ASCII spacing/line terminators,
// U+2028/U+2029, and BOM. unicode.IsSpace is not exact: it includes U+0085 and
// excludes BOM. TypeScript's internal stringutil helper also accepts U+0085
// and U+200B, so it cannot model ESLint's JavaScript regexp semantics here.
func isECMAScriptWhitespace(r rune) bool {
	if unicode.Is(unicode.Zs, r) {
		return true
	}
	switch r {
	case '\t', '\v', '\f', '\n', '\r', '\u2028', '\u2029', '\uFEFF':
		return true
	default:
		return false
	}
}
