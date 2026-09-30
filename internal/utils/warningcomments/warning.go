// Package warningcomments implements ESLint warning-term matching and display.
package warningcomments

import (
	"strings"

	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
	esregexp "github.com/web-infra-dev/rslint/internal/utils/ecmascript/regexp"
)

const charLimit = 40

// escapeRegExp mirrors the `escape-string-regexp` npm package upstream uses:
// backslash-escape the regexp metacharacters, and encode `-` as `\x2d` so a
// decoration/term string embedded inside a `[...]` character class can never
// be misread as forming a range.
func escapeRegExp(s string) string {
	var b strings.Builder
	for i := range len(s) {
		c := s[i]
		switch c {
		case '|', '\\', '{', '}', '(', ')', '[', ']', '^', '$', '+', '*', '?', '.':
			b.WriteByte('\\')
			b.WriteByte(c)
		case '-':
			b.WriteString(`\x2d`)
		default:
			b.WriteByte(c)
		}
	}
	return b.String()
}

// termStartsWithWordChar / termEndsWithWordChar port upstream's `/^\w/u` and
// `/\w$/u` term probes, which decide whether a word boundary is required on
// that side of the term.
var (
	termStartsWithWordChar = esregexp.MustCompile(`^\w`, "u")
	termEndsWithWordChar   = esregexp.MustCompile(`\w$`, "u")
)

// Pattern ports the upstream rule's convertToRegExp 1:1: build a
// case-insensitive regexp that matches `term` as a whole word in the
// configured location. See the source comment in ESLint's
// lib/rules/no-warning-comments.js for the rationale behind each prefix /
// suffix case. A term that cannot be compiled yields a nil regexp, which
// never matches.
func Pattern(term, location, decoration string) *esregexp.RegExp {
	escaped := escapeRegExp(term)
	escapedDecoration := escapeRegExp(decoration)

	var prefix string
	if location == "start" {
		prefix = `^[\s` + escapedDecoration + `]*`
	} else if termStartsWithWordChar.Test(term) {
		prefix = `\b`
	}

	suffix := ""
	if termEndsWithWordChar.Test(term) {
		suffix = `\b`
	}

	// Case-insensitive with Unicode case folding, as upstream compiles it —
	// which is also what puts U+017F and U+212A inside a word boundary.
	re, err := esregexp.Compile(prefix+escaped+suffix, "iu")
	if err != nil {
		return nil
	}
	return re
}

// Display ports upstream's display-truncation loop: rebuild the
// trimmed, whitespace-collapsed comment word by word, stopping (with a
// trailing "...") the moment adding the next word would exceed charLimit
// UTF-16 code units — the same unit JS's String#length counts, so an astral
// character (outside the BMP) counts as 2.
func Display(comment string) string {
	var b strings.Builder
	truncated := false
	first := true

	// Splitting on runs of ECMAScript WhiteSpace/LineTerminator, and dropping
	// the empty leading/trailing pieces, is what `comment.trim().split(/\s+/u)`
	// comes to: JS regex `\s` matches those two productions together.
	// strings.Fields can't stand in for it — Go's unicode.IsSpace counts U+0085
	// NEL, which JS does not, and skips U+FEFF BOM, which JS counts.
	for _, word := range strings.FieldsFunc(comment, ecmascript.IsWhiteSpaceOrLineTerminator) {
		var candidate string
		if first {
			candidate = word
		} else {
			candidate = b.String() + " " + word
		}

		if ecmascript.StringCodeUnitCount(candidate) <= charLimit {
			if first {
				b.WriteString(word)
				first = false
			} else {
				b.WriteByte(' ')
				b.WriteString(word)
			}
		} else {
			truncated = true
			break
		}
	}

	if truncated {
		return b.String() + "..."
	}
	return b.String()
}
