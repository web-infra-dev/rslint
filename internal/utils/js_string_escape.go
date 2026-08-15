package utils

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// EscapeJSSingleQuotedString escapes text for the contents of an ECMAScript
// single-quoted string literal. It preserves printable Unicode while escaping
// the delimiter, backslash, control characters, and every line terminator.
func EscapeJSSingleQuotedString(text string) string {
	var escaped strings.Builder
	last := 0
	for offset := 0; offset < len(text); {
		r, size := decodeJSStringRune(text[offset:])
		replacement := ""
		switch r {
		case '\\':
			replacement = `\\`
		case '\'':
			replacement = `\'`
		case '\b':
			replacement = `\b`
		case '\f':
			replacement = `\f`
		case '\n':
			replacement = `\n`
		case '\r':
			replacement = `\r`
		case '\t':
			replacement = `\t`
		case '\v':
			replacement = `\v`
		case '\u2028':
			replacement = `\u2028`
		case '\u2029':
			replacement = `\u2029`
		default:
			if r >= 0xD800 && r <= 0xDFFF {
				replacement = fmt.Sprintf(`\u%04X`, r)
			} else if r == utf8.RuneError && size == 1 {
				// This byte is neither UTF-8 nor tsgo's three-byte lone-
				// surrogate sentinel. Escaping it is conservative and, most
				// importantly, cannot emit malformed source or panic.
				replacement = fmt.Sprintf(`\x%02X`, text[offset])
			} else if r >= 0 && r < 0x20 {
				replacement = fmt.Sprintf(`\x%02X`, r)
			}
		}
		if replacement == "" {
			offset += size
			continue
		}
		if escaped.Len() == 0 {
			escaped.Grow(len(text) + 8)
		}
		escaped.WriteString(text[last:offset])
		escaped.WriteString(replacement)
		offset += size
		last = offset
	}
	if escaped.Len() == 0 {
		return text
	}
	escaped.WriteString(text[last:])
	return escaped.String()
}

// decodeJSStringRune mirrors TypeScript Go's stringutil.DecodeJSStringRune.
// tsgo preserves a lone UTF-16 surrogate in a Go string as its three-byte
// CESU-8/WTF-8 sentinel, which unicode/utf8 intentionally rejects.
func decodeJSStringRune(text string) (rune, int) {
	if len(text) >= 3 && text[0] == 0xED && text[1] >= 0xA0 && text[1] <= 0xBF && text[2] >= 0x80 && text[2] <= 0xBF {
		return 0xD000 | rune(text[1]&0x3F)<<6 | rune(text[2]&0x3F), 3
	}
	return utf8.DecodeRuneInString(text)
}
