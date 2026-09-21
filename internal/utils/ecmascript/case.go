package ecmascript

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// StringToUpperCase ports String.prototype.toUpperCase, which maps every
// character of s to its full uppercase — the one Unicode's Default Case
// Conversion gives it, rather than the single-character mapping unicode.ToUpper
// stops at. `ß` becoming `SS` is the familiar case of the difference; `ﬁ`
// becoming `FI`, and `ΐ` becoming three characters, are the same rule.
//
// The full mapping lives in Unicode's SpecialCasing table, which the standard
// library does not ship, so this reads it out of golang.org/x/text. That
// package's tables carry the same edition of Unicode the standard library's
// do, and move to the next one under the same build constraint, so the two
// never answer from different editions.
//
// A lone surrogate is carried across untouched, as are bytes that are not
// UTF-8, which is what happens to any character with no uppercase anyway.
//
// https://tc39.es/ecma262/2024/multipage/text-processing.html#sec-string.prototype.touppercase
func StringToUpperCase(s string) string {
	if mapped, ok := asciiCase(s, 'a', 'z', 'A'-'a'); ok {
		return mapped
	}
	// A Caser may be stateful, so it is not shared across calls.
	return cases.Upper(language.Und).String(s)
}

// StringToLowerCase ports String.prototype.toLowerCase, the other half of what
// [StringToUpperCase] describes. Only one character lowercases to more than one
// — U+0130 LATIN CAPITAL LETTER I WITH DOT ABOVE keeps its dot as a combining
// mark — so the single-character mapping carries the rest, and the one place
// the answer turns on a character's surroundings is written out below: a
// capital sigma lowercases to U+03C2 ς rather than U+03C3 σ when it ends a
// word.
//
// https://tc39.es/ecma262/2024/multipage/text-processing.html#sec-string.prototype.tolowercase
func StringToLowerCase(s string) string {
	if mapped, ok := asciiCase(s, 'A', 'Z', 'a'-'A'); ok {
		return mapped
	}
	var out strings.Builder
	out.Grow(len(s))
	// Final_Sigma is the only condition the walk carries a context for, and it
	// asks about one character. Without that character there is nothing to
	// accumulate, and the two derived properties below cost more per character
	// than the mapping does.
	tracksFinalSigma := strings.ContainsRune(s, 0x03A3)
	// Whether the most recent character that Final_Sigma does not ignore was a
	// cased one. Accumulating it while walking forwards is what saves the
	// condition from having to scan backwards.
	casedBefore := false
	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		if r == utf8.RuneError && size == 1 {
			// Not UTF-8: a lone surrogate, which the compiler carries in an
			// encoding of its own, or a stray byte. Neither has a case, and
			// neither is case-ignorable, so the byte goes across as it came
			// and the Final_Sigma context starts over.
			out.WriteByte(s[i])
			casedBefore = false
			i++
			continue
		}
		switch {
		case r == 0x0130:
			out.WriteString("i̇")
		case r == 0x03A3 && casedBefore && !casedFollows(s[i+size:]):
			out.WriteRune(0x03C2)
		default:
			out.WriteRune(unicode.ToLower(r))
		}
		if tracksFinalSigma && !isCaseIgnorable(r) {
			casedBefore = isCased(r)
		}
		i += size
	}
	return out.String()
}

// StringToLocaleUpperCase ports String.prototype.toLocaleUpperCase.
//
// The locale-sensitive mapping departs from the plain one for three languages
// — Turkish, Azeri and Lithuanian — and only ever for the host's own locale,
// which a linter has no business reading: the same source has to draw the same
// diagnostics on every machine it runs on. So this answers with the mapping
// [StringToUpperCase] gives, which is what the locale-sensitive one comes to
// everywhere else.
//
// https://tc39.es/ecma262/2024/multipage/text-processing.html#sec-string.prototype.tolocaleuppercase
func StringToLocaleUpperCase(s string) string {
	return StringToUpperCase(s)
}

// StringToLocaleLowerCase ports String.prototype.toLocaleLowerCase, and reads
// its locale the way [StringToLocaleUpperCase] describes.
//
// https://tc39.es/ecma262/2024/multipage/text-processing.html#sec-string.prototype.tolocalelowercase
func StringToLocaleLowerCase(s string) string {
	return StringToLowerCase(s)
}

// EqualsWhenLowercased ports the comparison a rule makes by lowercasing both
// sides — `a.toLowerCase() === b.toLowerCase()`.
//
// It is not strings.EqualFold, which asks Go's own question. Go folds U+017F
// LATIN SMALL LETTER LONG S together with `s`, and JavaScript keeps the two
// apart here, because neither lowercases onto the other.
func EqualsWhenLowercased(a, b string) bool {
	if equal, settled := asciiEquals(a, b); settled {
		return equal
	}
	return StringToLowerCase(a) == StringToLowerCase(b)
}

// EqualsWhenUppercased is [EqualsWhenLowercased] for the other half of the
// pair — `a.toUpperCase() === b.toUpperCase()` — which is a different question
// and not always the same answer. A dotless `ı` uppercases onto `I` while it
// lowercases onto itself; a Kelvin sign lowercases onto `k` while it
// uppercases onto itself.
//
// Which one a rule asks is whichever one it is ported from asks: jsx-ast-utils
// looks a prop up by uppercasing, and the rules that read a value rather than
// a name mostly lowercase.
func EqualsWhenUppercased(a, b string) bool {
	if equal, settled := asciiEquals(a, b); settled {
		return equal
	}
	return StringToUpperCase(a) == StringToUpperCase(b)
}

// asciiEquals compares two strings that hold nothing but ASCII, where a case
// mapping is one byte for one byte and both directions come to the same
// answer. The second result is false when one of them holds anything else, so
// the caller has to map in full. Nearly every comparison a rule makes ends
// here.
func asciiEquals(a, b string) (equal, settled bool) {
	if len(a) != len(b) {
		// Past ASCII a lowercase mapping can change a string's length, so an
		// unequal length only settles the question while both sides are ASCII.
		return false, isASCII(a) && isASCII(b)
	}
	for i := range len(a) {
		left, right := a[i], b[i]
		if left >= utf8.RuneSelf || right >= utf8.RuneSelf {
			return false, false
		}
		if left != right && asciiLower(left) != asciiLower(right) {
			return false, true
		}
	}
	return true, true
}

// isASCII reports whether s holds nothing a case mapping reads as more than
// one byte.
func isASCII(s string) bool {
	for i := range len(s) {
		if s[i] >= utf8.RuneSelf {
			return false
		}
	}
	return true
}

// asciiLower is the lowercase mapping of one ASCII byte, which is the whole of
// String.prototype.toLowerCase for a string that holds nothing else.
func asciiLower(c byte) byte {
	if c >= 'A' && c <= 'Z' {
		return c + 'a' - 'A'
	}
	return c
}

// asciiCase maps the letters of s between lo and hi by delta, reporting false
// when s holds a character outside ASCII — which is where a case mapping stops
// being a per-byte one. Most strings a rule reads hold nothing else.
func asciiCase(s string, lo, hi byte, delta int) (string, bool) {
	mapped := false
	for i := range len(s) {
		c := s[i]
		if c >= utf8.RuneSelf {
			return "", false
		}
		mapped = mapped || (c >= lo && c <= hi)
	}
	if !mapped {
		return s, true
	}
	buf := []byte(s)
	for i, c := range buf {
		if c >= lo && c <= hi {
			buf[i] = byte(int(c) + delta)
		}
	}
	return string(buf), true
}

// casedFollows reports whether the forward half of the Final_Sigma condition
// holds against s, which starts just past the sigma: a cased character, once
// the ones the condition ignores are stepped over.
func casedFollows(s string) bool {
	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		if r == utf8.RuneError && size == 1 {
			return false
		}
		if !isCaseIgnorable(r) {
			return isCased(r)
		}
		i += size
	}
	return false
}

// isCased reports Unicode's Cased property, which is a derived one the standard
// library ships no table for. It is exactly the union of the five tables the
// standard library does ship that its definition names.
func isCased(r rune) bool {
	return unicode.In(r, unicode.Ll, unicode.Lu, unicode.Lt, unicode.Other_Lowercase, unicode.Other_Uppercase)
}

// isCaseIgnorable reports Unicode's Case_Ignorable property, derived the way
// [isCased] describes. Its definition names one thing the standard library has
// no table for, so [isWordBreakCaseIgnorable] states that part.
func isCaseIgnorable(r rune) bool {
	return unicode.In(r, unicode.Mn, unicode.Me, unicode.Cf, unicode.Lm, unicode.Sk) ||
		isWordBreakCaseIgnorable(r)
}

// isWordBreakCaseIgnorable reports the characters Case_Ignorable draws from the
// Word_Break property — the values MidLetter, MidNumLet and Single_Quote. They
// are what keeps an apostrophe or a full stop from ending the word a sigma sits
// at the end of.
func isWordBreakCaseIgnorable(r rune) bool {
	switch r {
	case '\'', '.', ':', 0x00B7, 0x0387, 0x055F, 0x05F4,
		0x2018, 0x2019, 0x2024, 0x2027,
		0xFE13, 0xFE52, 0xFE55, 0xFF07, 0xFF0E, 0xFF1A:
		return true
	}
	return false
}
