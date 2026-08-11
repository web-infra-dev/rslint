// Package ecmascript holds the pieces of JavaScript's own semantics that a
// port has to reproduce exactly, where Go's standard library answers a nearby
// but different question. Trimming a string, deciding what counts as blank,
// comparing two characters without regard to case, writing a number back out —
// each of those is specified by ECMAScript, and each of them differs from the
// Go equivalent in ways a rule ported from ESLint can observe.
//
// Everything here is a pure function of its arguments. Only
// [IsValidRegexLiteral] reaches past the standard library, for the scanner that
// reads a regexp literal, and a rule carries that dependency already.
package ecmascript

import (
	"strings"
	"unicode"
)

// LineTerminators are the characters ECMAScript reads as ending a line. They
// are what a JavaScript `.` in a regexp refuses to match, and what a template
// literal counts newlines by.
//
// https://tc39.es/ecma262/2024/multipage/ecmascript-language-lexical-grammar.html#prod-LineTerminator
const LineTerminators = "\n\r\u2028\u2029"

// IsLineTerminator reports whether r ends a line.
func IsLineTerminator(r rune) bool {
	switch r {
	case '\n', '\r', 0x2028, 0x2029:
		return true
	}
	return false
}

// IsWhiteSpace reports whether r is a character JavaScript trims from the edge
// of a string: an ECMAScript WhiteSpace or a LineTerminator.
//
// Go's unicode.IsSpace answers a different question, and so does the set
// strings.TrimSpace works from. They disagree in both directions: U+0085 NEL
// is Unicode whitespace but not ECMAScript whitespace, and U+FEFF is
// ECMAScript whitespace but not Unicode whitespace.
//
// Source: https://github.com/microsoft/typescript-go/blob/5652e65d5ae944375676d3955f9755e554576d41/internal/jsnum/string.go#L99
//
// https://tc39.es/ecma262/2024/multipage/ecmascript-language-lexical-grammar.html#prod-WhiteSpace
func IsWhiteSpace(r rune) bool {
	switch r {
	// LineTerminator
	case '\n', '\r', 0x2028, 0x2029:
		return true
	// WhiteSpace
	case '\t', '\v', '\f', 0xFEFF:
		return true
	}
	// WhiteSpace, which covers the ordinary space and U+00A0
	return unicode.Is(unicode.Zs, r)
}

// Trim trims s the way String.prototype.trim does.
func Trim(s string) string {
	return strings.TrimFunc(s, IsWhiteSpace)
}

// IsBlank reports whether s holds nothing but whitespace, which is what
// JavaScript's `s.trim() === ""` asks.
func IsBlank(s string) bool {
	for _, r := range s {
		if !IsWhiteSpace(r) {
			return false
		}
	}
	return true
}
