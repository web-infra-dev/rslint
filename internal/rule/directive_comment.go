package rule

import (
	"strings"
	"unicode/utf8"

	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

// Scanning shared by ESLint's block directive comments (`/* global */`,
// `/* exported */`). Each directive owns its own value syntax; what they share
// is locating the label and dropping a `-- justification` tail.

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
		if !strings.HasPrefix(text[start:end], keyword) {
			continue
		}
		restStart := start + len(keyword)
		if restStart == end {
			return restStart, true
		}
		r, _ := utf8.DecodeRuneInString(text[restStart:end])
		if ecmascript.IsWhiteSpaceOrLineTerminator(r) {
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
		if !ecmascript.IsWhiteSpaceOrLineTerminator(r) {
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
			if ecmascript.IsWhiteSpaceOrLineTerminator(next) {
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
		if !ecmascript.IsWhiteSpaceOrLineTerminator(r) {
			break
		}
		start += size
	}
	for end > start {
		r, size := utf8.DecodeLastRuneInString(text[start:end])
		if !ecmascript.IsWhiteSpaceOrLineTerminator(r) {
			break
		}
		end -= size
	}
	return start, end
}
