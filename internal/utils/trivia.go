package utils

import (
	"unicode/utf8"
)

// IsTriviaWhitespaceByte reports whether `b` is one of the ASCII whitespace
// or line-terminator bytes recognized by ECMAScript §12.2 / §12.3:
//
//	U+0009 HT, U+000A LF, U+000B VT, U+000C FF, U+000D CR, U+0020 SP.
//
// Used as the fast path for forward/reverse trivia scans on byte streams.
// For non-ASCII bytes (>= 0x80), the caller MUST decode the rune and fall
// back to IsTriviaWhitespaceRune — multi-byte UTF-8 sequences encode runes
// like NBSP (U+00A0) whose individual bytes are >= 0x80 and would be misread
// by a byte-only test.
func IsTriviaWhitespaceByte(b byte) bool {
	switch b {
	case ' ', '\t', '\n', '\r', '\f', '\v':
		return true
	}
	return false
}

// IsTriviaWhitespaceRune reports whether `r` is a non-ASCII whitespace or
// line-terminator rune recognized by ECMAScript:
//
//   - WhiteSpace (§12.2): U+00A0 NBSP, U+FEFF ZWNBSP, plus any rune in
//     Unicode `Space_Separator` (Zs) — U+1680 Ogham, U+2000–U+200A
//     (En Quad through Hair Space), U+202F Narrow NBSP, U+205F Medium
//     Mathematical Space, U+3000 Ideographic Space.
//   - LineTerminator (§12.3): U+2028 LS, U+2029 PS.
//
// Should ONLY be called for runes >= U+0080. The ASCII subset is handled by
// IsTriviaWhitespaceByte on the fast path.
//
// Deliberately excluded — they are NOT matched by JS `\s` / ECMAScript
// WhiteSpace (verified empirically against V8 and ESLint with `/\s/`):
//
//   - U+0085 Next Line: in Unicode `\p{White_Space}` but NOT in ES
//     WhiteSpace (only specific allow-listed runes count).
//   - U+200B Zero Width Space: category Cf (Format), not Zs; tsgo's scanner
//     treats it as whitespace internally, but that's a TS-compiler quirk
//     not ESLint behavior.
func IsTriviaWhitespaceRune(r rune) bool {
	switch r {
	case 0x00A0, // <No-Break Space>
		0x1680, // <Ogham Space Mark>
		0x2000, // <En Quad>
		0x2001, // <Em Quad>
		0x2002, // <En Space>
		0x2003, // <Em Space>
		0x2004, // <Three-Per-Em Space>
		0x2005, // <Four-Per-Em Space>
		0x2006, // <Six-Per-Em Space>
		0x2007, // <Figure Space>
		0x2008, // <Punctuation Space>
		0x2009, // <Thin Space>
		0x200A, // <Hair Space>
		0x2028, // <Line Separator>
		0x2029, // <Paragraph Separator>
		0x202F, // <Narrow No-Break Space>
		0x205F, // <Medium Mathematical Space>
		0x3000, // <Ideographic Space>
		0xFEFF: // <Byte Order Mark / Zero Width No-Break Space>
		return true
	}
	return false
}

// ContainsLineTerminator reports whether the byte range `[low, high)` of
// `text` contains any ECMAScript LineTerminator (§12.3): LF (`\n`), CR (`\r`),
// LS (U+2028), or PS (U+2029).
//
// Used by spacing rules to short-circuit "tokens on different lines"
// decisions equivalent to ESLint's `isTokenOnSameLine` helper. Out-of-range
// indices are clamped, and an empty range yields `false`.
//
// LS and PS are encoded in UTF-8 as `E2 80 A8` and `E2 80 A9` respectively;
// detecting them via byte-level prefix `E2 80 A8`/`A9` is exactly as
// reliable as a rune scan and significantly faster on cold cache.
func ContainsLineTerminator(text string, low, high int) bool {
	if low < 0 {
		low = 0
	}
	if high > len(text) {
		high = len(text)
	}
	for i := low; i < high; i++ {
		c := text[i]
		if c == '\n' || c == '\r' {
			return true
		}
		// Look for the LS/PS UTF-8 prefix `E2 80 A8` / `E2 80 A9`.
		if c == 0xE2 && i+2 < high && text[i+1] == 0x80 && (text[i+2] == 0xA8 || text[i+2] == 0xA9) {
			return true
		}
	}
	return false
}

// SkipLeadingWhitespace walks forward from `low` through ECMAScript trivia
// whitespace and line terminators (the union of §12.2 WhiteSpace and §12.3
// LineTerminator), returning the position of the first non-whitespace byte
// (or `high` if nothing but whitespace remains).
//
// Counterpart to SkipTrailingWhitespace. tsgo's `scanner.SkipTriviaEx` also
// skips comments and conflict markers; this helper is purely whitespace +
// line-terminator, useful when callers need to handle comments / other
// tokens separately. ASCII bytes take the fast path; non-ASCII bytes
// decode via `utf8.DecodeRune` and dispatch to `IsTriviaWhitespaceRune`.
func SkipLeadingWhitespace(text string, low, high int) int {
	if low < 0 {
		low = 0
	}
	if high > len(text) {
		high = len(text)
	}
	p := low
	for p < high {
		if text[p] < 0x80 {
			if !IsTriviaWhitespaceByte(text[p]) {
				return p
			}
			p++
			continue
		}
		r, size := utf8.DecodeRuneInString(text[p:])
		if size == 0 || r == utf8.RuneError || !IsTriviaWhitespaceRune(r) {
			return p
		}
		p += size
	}
	return high
}

// SkipTrailingWhitespace walks back from `high` through ECMAScript trivia
// whitespace and line terminators (the union of §12.2 WhiteSpace and §12.3
// LineTerminator), returning the position one past the last non-whitespace
// byte (i.e. the end-exclusive of the last token in `[low, high)`). Returns
// `low` when nothing but whitespace remains.
//
// tsgo exposes `scanner.SkipTriviaEx` for forward scanning only; this
// function is the missing reverse counterpart. ASCII bytes are handled on
// the fast path; non-ASCII bytes decode the rune via `utf8.DecodeLastRune`
// and dispatch to `IsTriviaWhitespaceRune`.
//
// Does NOT skip block or line comments — callers that need to skip those
// (or recognize comment boundaries) must do it themselves. This mirrors
// `SkipTriviaEx`'s `StopAtComments: true` mode, where the forward scan also
// stops at the start of a comment.
func SkipTrailingWhitespace(text string, low, high int) int {
	if low < 0 {
		low = 0
	}
	if high > len(text) {
		high = len(text)
	}
	p := high
	for p > low {
		// Fast path: ASCII byte.
		if text[p-1] < 0x80 {
			if !IsTriviaWhitespaceByte(text[p-1]) {
				return p
			}
			p--
			continue
		}
		// Slow path: decode the previous rune.
		r, size := utf8.DecodeLastRuneInString(text[:p])
		if size == 0 || r == utf8.RuneError || !IsTriviaWhitespaceRune(r) {
			return p
		}
		p -= size
	}
	return low
}
