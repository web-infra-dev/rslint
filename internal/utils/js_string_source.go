package utils

import (
	"strings"
	"unicode/utf8"
)

// StringCodeUnit represents one UTF-16 code unit produced by evaluating a JS
// string literal or no-substitution template literal, paired with the byte
// range of the source text that produced it.
//
// For most code units one entry maps 1:1 to a slice of the literal's source
// text. For astral escapes like `\u{1F44D}` and for raw astral characters (4
// UTF-8 bytes) two consecutive entries are emitted whose `Start` and `End`
// are identical — they represent the high and low surrogate halves of the
// same source span.
//
// Line continuations (`\<LF>`, `\<CR>`, `\<CR><LF>`, `\<LS>`, `\<PS>`) do not
// produce any code unit.
//
// This mirrors ESLint's `utils/char-source.js` `CodeUnit` class, with the
// value added so callers don't need to re-parse the source text.
type StringCodeUnit struct {
	// Value is the UTF-16 code unit value (0..0xFFFF). For astral entries the
	// two consecutive units carry the high and low surrogate values.
	Value uint32
	// Start is the inclusive byte offset within the literal source text.
	Start int
	// End is the exclusive byte offset within the literal source text.
	End int
}

// ParseJSStringLiteralSource parses a JS string literal source text (including
// its surrounding quotes) and returns the per-code-unit source mapping.
//
// Returns nil if the input is not a well-formed string literal prologue
// (leading quote, closing quote) — the caller should treat this as "give up".
// The parser is permissive on legacy-octal and identity escapes (matching
// ESLint's `char-source.js`) and does not enforce strict-mode restrictions.
func ParseJSStringLiteralSource(source string) []StringCodeUnit {
	if len(source) < 2 {
		return nil
	}
	quote := source[0]
	if quote != '\'' && quote != '"' {
		return nil
	}
	reader := &jsSrcReader{src: source, pos: 1}
	out := []StringCodeUnit{}
	for reader.pos < len(source) {
		ch := reader.src[reader.pos]
		if ch == quote {
			return out
		}
		if ch == '\\' {
			out = appendEscape(out, reader)
			continue
		}
		// Raw character — UTF-8 aware.
		start := reader.pos
		r, w := utf8.DecodeRuneInString(source[reader.pos:])
		if w == 0 {
			return nil
		}
		reader.pos += w
		if r > 0xFFFF {
			hi, lo := encodeSurrogatePair(uint32(r))
			out = append(out,
				StringCodeUnit{Value: hi, Start: start, End: reader.pos},
				StringCodeUnit{Value: lo, Start: start, End: reader.pos},
			)
		} else {
			out = append(out, StringCodeUnit{Value: uint32(r), Start: start, End: reader.pos})
		}
	}
	// Unterminated — caller should treat as malformed.
	return nil
}

// ParseJSTemplateLiteralSource parses a no-substitution template literal
// source (including surrounding backticks) similarly. For templates with
// substitution `${...}`, pass just one segment at a time is unsupported —
// callers should treat substituted templates as dynamic.
func ParseJSTemplateLiteralSource(source string) []StringCodeUnit {
	if len(source) < 2 || source[0] != '`' {
		return nil
	}
	reader := &jsSrcReader{src: source, pos: 1}
	out := []StringCodeUnit{}
	for reader.pos < len(source) {
		ch := reader.src[reader.pos]
		if ch == '`' {
			return out
		}
		if ch == '$' && reader.pos+1 < len(source) && reader.src[reader.pos+1] == '{' {
			// Start of substitution — end of this span.
			return out
		}
		if ch == '\\' {
			out = appendEscape(out, reader)
			continue
		}
		// Raw CR/LF/CRLF in template literals → cooked LF, keep source extent.
		if ch == '\r' {
			start := reader.pos
			reader.pos++
			if reader.pos < len(source) && reader.src[reader.pos] == '\n' {
				reader.pos++
			}
			out = append(out, StringCodeUnit{Value: '\n', Start: start, End: reader.pos})
			continue
		}
		// Other raw character — UTF-8 aware.
		start := reader.pos
		r, w := utf8.DecodeRuneInString(source[reader.pos:])
		if w == 0 {
			return nil
		}
		reader.pos += w
		if r > 0xFFFF {
			hi, lo := encodeSurrogatePair(uint32(r))
			out = append(out,
				StringCodeUnit{Value: hi, Start: start, End: reader.pos},
				StringCodeUnit{Value: lo, Start: start, End: reader.pos},
			)
		} else {
			out = append(out, StringCodeUnit{Value: uint32(r), Start: start, End: reader.pos})
		}
	}
	// Unterminated.
	return nil
}

// jsSrcReader tracks the current scan position within the literal source.
type jsSrcReader struct {
	src string
	pos int
}

// appendEscape reads one backslash escape sequence (or line continuation)
// starting at `reader.pos` (which must point at `\`) and appends 0, 1 or 2
// code units to `out`.
func appendEscape(out []StringCodeUnit, reader *jsSrcReader) []StringCodeUnit {
	start := reader.pos
	reader.pos++ // consume `\`
	if reader.pos >= len(reader.src) {
		return out
	}
	next, w := utf8.DecodeRuneInString(reader.src[reader.pos:])
	reader.pos += w

	// Simple escapes.
	if v, ok := simpleEscapeValue(next); ok {
		return append(out, StringCodeUnit{Value: v, Start: start, End: reader.pos})
	}

	switch next {
	case 'x':
		// `\xHH`
		if reader.pos+1 < len(reader.src) &&
			isHex(reader.src[reader.pos]) && isHex(reader.src[reader.pos+1]) {
			v := parseHexByte(reader.src[reader.pos]) << 4
			v |= parseHexByte(reader.src[reader.pos+1])
			reader.pos += 2
			return append(out, StringCodeUnit{Value: v, Start: start, End: reader.pos})
		}
		// Malformed — best effort: treat as identity `x`.
		return append(out, StringCodeUnit{Value: 'x', Start: start, End: reader.pos})
	case 'u':
		return appendUnicodeEscape(out, reader, start)
	case '\r':
		// Line continuation `\<CR>` or `\<CR><LF>`.
		if reader.pos < len(reader.src) && reader.src[reader.pos] == '\n' {
			reader.pos++
		}
		return out
	case '\n', '\u2028', '\u2029':
		// Line continuation. (The LS/PS rune was already consumed by DecodeRuneInString.)
		return out
	case '0', '1', '2', '3':
		return appendOctalEscape(out, reader, start, 3)
	case '4', '5', '6', '7':
		return appendOctalEscape(out, reader, start, 2)
	default:
		// Identity escape — the character itself becomes the code unit value.
		v := uint32(next)
		if v > 0xFFFF {
			hi, lo := encodeSurrogatePair(v)
			return append(out,
				StringCodeUnit{Value: hi, Start: start, End: reader.pos},
				StringCodeUnit{Value: lo, Start: start, End: reader.pos},
			)
		}
		return append(out, StringCodeUnit{Value: v, Start: start, End: reader.pos})
	}
}

// appendUnicodeEscape reads `\uHHHH` or `\u{H...}` (the `\u` was already
// consumed). `start` is the byte offset of the leading `\`.
func appendUnicodeEscape(out []StringCodeUnit, reader *jsSrcReader, start int) []StringCodeUnit {
	// `\u{H...}` variant.
	if reader.pos < len(reader.src) && reader.src[reader.pos] == '{' {
		closeRel := strings.IndexByte(reader.src[reader.pos+1:], '}')
		if closeRel < 0 {
			return out // unterminated — give up on this escape
		}
		hex := reader.src[reader.pos+1 : reader.pos+1+closeRel]
		reader.pos = reader.pos + 1 + closeRel + 1 // past `}`
		if hex == "" || !allHexStr(hex) {
			return out
		}
		cp := parseHexStr(hex)
		if cp > 0xFFFF {
			hi, lo := encodeSurrogatePair(cp)
			return append(out,
				StringCodeUnit{Value: hi, Start: start, End: reader.pos},
				StringCodeUnit{Value: lo, Start: start, End: reader.pos},
			)
		}
		return append(out, StringCodeUnit{Value: cp, Start: start, End: reader.pos})
	}
	// `\uHHHH` fixed-length variant.
	if reader.pos+3 < len(reader.src) &&
		isHex(reader.src[reader.pos]) && isHex(reader.src[reader.pos+1]) &&
		isHex(reader.src[reader.pos+2]) && isHex(reader.src[reader.pos+3]) {
		v := parseHexStr(reader.src[reader.pos : reader.pos+4])
		reader.pos += 4
		return append(out, StringCodeUnit{Value: v, Start: start, End: reader.pos})
	}
	// Malformed — best effort: treat as identity `u`.
	return append(out, StringCodeUnit{Value: 'u', Start: start, End: reader.pos})
}

// appendOctalEscape reads up to `maxExtra` additional octal digits after a
// first octal digit that has already been consumed into reader.pos.
// `start` is the byte offset of the leading `\`.
func appendOctalEscape(out []StringCodeUnit, reader *jsSrcReader, start int, maxLen int) []StringCodeUnit {
	// reader.pos points to the char AFTER the first octal digit we consumed.
	// Source-wise, the first digit is at start+1. Collect contiguous octal
	// digits up to (maxLen - 1) more after the first.
	firstDigit := reader.src[start+1]
	digits := []byte{firstDigit}
	extra := maxLen - 1
	for i := 0; i < extra && reader.pos < len(reader.src); i++ {
		c := reader.src[reader.pos]
		if c < '0' || c > '7' {
			break
		}
		digits = append(digits, c)
		reader.pos++
	}
	v := uint32(0)
	for _, d := range digits {
		v = v*8 + uint32(d-'0')
	}
	return append(out, StringCodeUnit{Value: v, Start: start, End: reader.pos})
}

// simpleEscapeValue returns the value of one of the simple escapes ('b', 'f',
// 'n', 'r', 't', 'v') or ok=false otherwise.
func simpleEscapeValue(c rune) (uint32, bool) {
	switch c {
	case 'b':
		return '\b', true
	case 'f':
		return '\f', true
	case 'n':
		return '\n', true
	case 'r':
		return '\r', true
	case 't':
		return '\t', true
	case 'v':
		return '\v', true
	}
	return 0, false
}

// encodeSurrogatePair splits an astral code point (> 0xFFFF) into its UTF-16
// high/low surrogate components.
func encodeSurrogatePair(cp uint32) (hi, lo uint32) {
	cp -= 0x10000
	hi = 0xD800 + (cp >> 10)
	lo = 0xDC00 + (cp & 0x3FF)
	return
}

// IsHexDigit reports whether b is an ASCII hex digit (0-9, a-f, A-F).
// Exposed for rules that do their own pattern scanning.
func IsHexDigit(b byte) bool {
	return (b >= '0' && b <= '9') || (b >= 'a' && b <= 'f') || (b >= 'A' && b <= 'F')
}

// AllHexDigits reports whether s is non-empty and every byte is a hex digit.
func AllHexDigits(s string) bool { return allHexStr(s) }

// ParseHexUint parses s as a hex number, returning 0 on empty/non-hex input.
func ParseHexUint(s string) uint32 { return parseHexStr(s) }

func isHex(b byte) bool {
	return IsHexDigit(b)
}

func allHexStr(s string) bool {
	if s == "" {
		return false
	}
	for i := range len(s) {
		if !isHex(s[i]) {
			return false
		}
	}
	return true
}

func parseHexByte(b byte) uint32 {
	switch {
	case b >= '0' && b <= '9':
		return uint32(b - '0')
	case b >= 'a' && b <= 'f':
		return uint32(b-'a') + 10
	case b >= 'A' && b <= 'F':
		return uint32(b-'A') + 10
	}
	return 0
}

func parseHexStr(s string) uint32 {
	v := uint32(0)
	for i := range len(s) {
		v = (v << 4) | parseHexByte(s[i])
	}
	return v
}
