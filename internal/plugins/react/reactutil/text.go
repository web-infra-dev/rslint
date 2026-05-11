package reactutil

import (
	"strings"
	"unicode"
)

// ApplyData expands `{{key}}` placeholders in a message template using the
// given data map, mirroring ESLint's `RuleMessage.data` interpolation. Keys
// not present in `data` are left untouched, matching ESLint's behavior of
// passing through unknown placeholders verbatim. Use this whenever a rule
// emits a templated `RuleMessage.Description` so the in-rule code reads like
// the upstream `messages` table instead of hand-rolled `strings.ReplaceAll`
// loops.
func ApplyData(template string, data map[string]string) string {
	if len(data) == 0 {
		return template
	}
	out := template
	for k, v := range data {
		out = strings.ReplaceAll(out, "{{"+k+"}}", v)
	}
	return out
}

// HorizontalWhitespacePrefix returns the longest prefix of s consisting of
// ECMA WhiteSpace characters (excluding LineTerminators). It matches the
// behavior of `/^\s*/` applied to a single line — used by JSX layout rules
// that compute "indent of line N" without crossing into the next line.
//
// LineTerminators (\n, \r,  ,  ) are NOT consumed, so passing a
// multi-line string only ever returns the indent of the first line.
func HorizontalWhitespacePrefix(s string) string {
	for i, r := range s {
		if !isHorizontalWhitespace(r) {
			return s[:i]
		}
	}
	return s
}

func isHorizontalWhitespace(r rune) bool {
	switch r {
	case ' ', '\t', '\v', '\f', 0xFEFF:
		return true
	}
	return unicode.Is(unicode.Zs, r)
}

// UTF16Length returns the number of UTF-16 code units required to encode s.
// ASCII / BMP runes count as 1; runes outside the BMP (>= U+10000) count as
// 2 (surrogate pair). Use this when computing column positions that must
// agree with ESLint's `loc.column` semantics — Go's `len(s)` is byte length
// (UTF-8), which can over-count for multi-byte characters.
func UTF16Length(s string) int {
	n := 0
	for _, r := range s {
		if r < 0x10000 {
			n++
		} else {
			n += 2
		}
	}
	return n
}
