// Package ecmascript holds the pieces of JavaScript's own semantics that a
// port has to reproduce exactly, where Go's standard library answers a nearby
// but different question. Trimming a string, deciding what counts as blank,
// mapping a string's case, writing a number back out — each of those is
// specified by ECMAScript, and each of them differs from the Go equivalent in
// ways a rule ported from ESLint can observe.
//
// A helper standing in for something JavaScript exposes carries that thing's
// name with its receiver in front — [StringTrim] for String.prototype.trim,
// [NumberToString] for Number::toString — so that a reader checking the port
// against the language has one name to look for. Everything else names what
// ECMAScript names it: a grammar production ([IsWhiteSpace],
// [IsLineTerminator]), the two of them together ([IsWhiteSpaceOrLineTerminator])
// or, where the language spells out no name at all, the question being asked
// ([IsBlank]).
//
// Everything here is a pure function of its arguments. Only
// [IsValidRegexLiteral] reaches past the standard library, for the scanner that
// reads a regexp literal, and a rule carries that dependency already.
//
// The one larger piece of the language a rule has to reproduce lives in
// [ecmascript/regexp], which compiles a JavaScript RegExp and closes the gaps
// between what ECMAScript specifies and what a Go engine does. It is a package
// of its own so that only a caller who needs a backtracking engine carries the
// dependency on one.
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

// IsWhiteSpace reports whether r is an ECMAScript WhiteSpace: a tab, a vertical
// tab, a form feed, U+FEFF ZWNBSP, or a Space_Separator, which is where the
// ordinary space and U+00A0 come from. A LineTerminator is not one of these —
// the language reads the two together often enough to have a name for that too,
// and this package spells it [IsWhiteSpaceOrLineTerminator].
//
// https://tc39.es/ecma262/2024/multipage/ecmascript-language-lexical-grammar.html#prod-WhiteSpace
func IsWhiteSpace(r rune) bool {
	switch r {
	case '\t', '\v', '\f', 0xFEFF:
		return true
	}
	// Space_Separator, which covers the ordinary space and U+00A0.
	return unicode.Is(unicode.Zs, r)
}

// IsWhiteSpaceOrLineTerminator reports whether r is a character JavaScript
// trims from the edge of a string, which is also the set a regexp's `\s`
// matches: an ECMAScript WhiteSpace or a LineTerminator.
//
// Go's unicode.IsSpace answers a different question, and so does the set
// strings.TrimSpace works from. They disagree in both directions: U+0085 NEL
// is Unicode whitespace but not ECMAScript whitespace, and U+FEFF is
// ECMAScript whitespace but not Unicode whitespace.
//
// Source: https://github.com/microsoft/typescript-go/blob/5652e65d5ae944375676d3955f9755e554576d41/internal/jsnum/string.go#L99
func IsWhiteSpaceOrLineTerminator(r rune) bool {
	return IsWhiteSpace(r) || IsLineTerminator(r)
}

// StringTrim trims s the way String.prototype.trim does.
func StringTrim(s string) string {
	return strings.TrimFunc(s, IsWhiteSpaceOrLineTerminator)
}

// StringTrimStart trims the leading edge the way
// String.prototype.trimStart does.
func StringTrimStart(s string) string {
	return strings.TrimLeftFunc(s, IsWhiteSpaceOrLineTerminator)
}

// StringTrimEnd trims the trailing edge the way String.prototype.trimEnd
// does.
func StringTrimEnd(s string) string {
	return strings.TrimRightFunc(s, IsWhiteSpaceOrLineTerminator)
}

// IsBlank reports whether s holds nothing but whitespace, which is what
// JavaScript's `s.trim() === ""` asks.
func IsBlank(s string) bool {
	for _, r := range s {
		if !IsWhiteSpaceOrLineTerminator(r) {
			return false
		}
	}
	return true
}
