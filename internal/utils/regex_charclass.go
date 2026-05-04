// cspell:ignore octals
package utils

import (
	"strings"
	"unicode/utf8"
)

// ---------------------------------------------------------------------------
// Public types
// ---------------------------------------------------------------------------

// RegexFlags captures the subset of ECMAScript regex flags that affects how
// patterns (and in particular character classes) are parsed.
type RegexFlags struct {
	Unicode     bool // u flag
	UnicodeSets bool // v flag
}

// UV reports whether unicode/unicodeSets mode is active (either u or v).
func (f RegexFlags) UV() bool { return f.Unicode || f.UnicodeSets }

// ParseRegexFlags returns a RegexFlags from a flag string (e.g. "gui").
func ParseRegexFlags(flags string) RegexFlags {
	return RegexFlags{
		Unicode:     strings.ContainsRune(flags, 'u'),
		UnicodeSets: strings.ContainsRune(flags, 'v'),
	}
}

// RegexCharElementKind classifies one element of a character class body.
type RegexCharElementKind int

const (
	// RegexCharSingle is a single character element (literal, escape, …).
	RegexCharSingle RegexCharElementKind = iota
	// RegexCharRange is a character range `a-b` (inclusive on both ends).
	// Value is the `a` side; Max is the `b` side. Only the min/max endpoints
	// appear in the resulting character sequence — the inner characters are
	// implicit.
	RegexCharRange
	// RegexCharBreaker is any element that interrupts a contiguous character
	// sequence within a class: `\d`, `\D`, `\w`, `\W`, `\s`, `\S`, `\b`, `\B`,
	// `\p{...}`, `\P{...}`, `\q{...}` (v-flag), nested `[...]` (v-flag), or a
	// v-flag set operator `--` / `&&`.
	RegexCharBreaker
)

// RegexCharElement is one element of a parsed character class body.
//
// For RegexCharSingle:
//   - Value holds the element's effective value (UTF-16 code unit or the
//     astral code point when combined via `\uHHHH\uHHHH` under u/v).
//   - IsUBrace is true iff the element was written as `\u{H...}`.
//   - IsLoneSurrogate is true iff the element is a lone surrogate from a raw
//     astral character under non-u mode (where regexpp sees it as two units).
//
// For RegexCharRange:
//   - Value / IsUBrace describe the `a` endpoint.
//   - Max / MaxIsUBrace describe the `b` endpoint.
//
// For RegexCharBreaker:
//   - Value, Max, IsUBrace are unused.
//
// Start / End are byte offsets within the pattern text covering the element's
// source extent.
type RegexCharElement struct {
	Kind            RegexCharElementKind
	Value           uint32
	IsUBrace        bool
	IsLoneSurrogate bool

	Max         uint32
	MaxIsUBrace bool

	Start int
	End   int

	// RawStart/RawEnd optionally narrow to just the element's own source
	// extent, distinct from Start/End which always span the element. For
	// ranges this pair equals Start/End.
}

// ---------------------------------------------------------------------------
// Pattern scanner (layer 1)
// ---------------------------------------------------------------------------

// IterateRegexCharacterClasses walks a regex pattern and invokes cb once per
// top-level character class (including any v-flag nested classes, at each
// nesting level). cb receives the byte range [start, end) covering `[`..`]`.
//
// Returns false if the pattern is malformed (unterminated class, unterminated
// escape at EOF). When false, cb may have been invoked for classes encountered
// before the error.
func IterateRegexCharacterClasses(pattern string, flags RegexFlags, cb func(start, end int)) bool {
	i := 0
	for i < len(pattern) {
		switch pattern[i] {
		case '\\':
			step, ok := SkipPatternEscape(pattern, i, flags)
			if !ok {
				return false
			}
			i += step
		case '[':
			end, ok := iterateClassFromLBracket(pattern, i, flags, cb)
			if !ok {
				return false
			}
			i = end
		case '(':
			// Parens don't affect character-class scanning; consume and continue.
			i++
		default:
			_, w := utf8.DecodeRuneInString(pattern[i:])
			if w == 0 {
				i++
			} else {
				i += w
			}
		}
	}
	return true
}

// iterateClassFromLBracket recursively scans a character class starting at
// pattern[start] (which must be `[`). Invokes cb for the class (and, under v
// flag, for any nested classes). Returns the index just past `]`, or ok=false
// on malformed input.
func iterateClassFromLBracket(pattern string, start int, flags RegexFlags, cb func(start, end int)) (int, bool) {
	end, ok := ClassEnd(pattern, start, flags)
	if !ok {
		return start, false
	}
	// Recurse into nested classes first (under v flag), then invoke cb on
	// the outermost.
	if flags.UnicodeSets {
		i := start + 1
		if i < end-1 && pattern[i] == '^' {
			i++
		}
		for i < end-1 {
			c := pattern[i]
			if c == '\\' {
				step, ok := SkipPatternEscape(pattern, i, flags)
				if !ok {
					return start, false
				}
				i += step
				continue
			}
			if c == '[' {
				nestedEnd, ok := iterateClassFromLBracket(pattern, i, flags, cb)
				if !ok {
					return start, false
				}
				i = nestedEnd
				continue
			}
			_, w := utf8.DecodeRuneInString(pattern[i:])
			if w == 0 {
				i++
			} else {
				i += w
			}
		}
	}
	cb(start, end)
	return end, true
}

// ClassEnd returns the byte index just past the matching `]` for a class
// starting at `[` at pattern[start]. Handles escaped `]`, v-flag nested
// classes, and `\q{...}` which contains a literal `]` inside braces.
func ClassEnd(pattern string, start int, flags RegexFlags) (int, bool) {
	if start >= len(pattern) || pattern[start] != '[' {
		return start, false
	}
	i := start + 1
	// A leading `^` is not an element by itself.
	if i < len(pattern) && pattern[i] == '^' {
		i++
	}
	depth := 1
	for i < len(pattern) {
		c := pattern[i]
		switch {
		case c == '\\':
			step, ok := SkipPatternEscape(pattern, i, flags)
			if !ok {
				return start, false
			}
			i += step
		case c == '[' && flags.UnicodeSets:
			depth++
			i++
		case c == ']':
			depth--
			i++
			if depth == 0 {
				return i, true
			}
		default:
			_, w := utf8.DecodeRuneInString(pattern[i:])
			if w == 0 {
				i++
			} else {
				i += w
			}
		}
	}
	return start, false
}

// SkipPatternEscape returns how many bytes a `\`-prefixed escape consumes at
// pattern[i] (including the leading `\`) for the purposes of class-boundary
// scanning. Returns ok=false at EOF on `\`.
func SkipPatternEscape(pattern string, i int, flags RegexFlags) (int, bool) {
	if i+1 >= len(pattern) {
		return 0, false
	}
	next := pattern[i+1]
	switch next {
	case 'x':
		if i+3 < len(pattern) && isHex(pattern[i+2]) && isHex(pattern[i+3]) {
			return 4, true
		}
		return 2, true
	case 'u':
		if i+2 < len(pattern) && pattern[i+2] == '{' {
			if !flags.UV() {
				return 2, true
			}
			closeRel := strings.IndexByte(pattern[i+3:], '}')
			if closeRel < 0 {
				return 2, true // best-effort recover
			}
			return 3 + closeRel + 1, true
		}
		if i+5 < len(pattern) && allHexStr(pattern[i+2:i+6]) {
			return 6, true
		}
		return 2, true
	case 'c':
		if i+2 < len(pattern) {
			return 3, true
		}
		return 2, true
	case 'p', 'P':
		if flags.UV() && i+2 < len(pattern) && pattern[i+2] == '{' {
			closeRel := strings.IndexByte(pattern[i+3:], '}')
			if closeRel >= 0 {
				return 3 + closeRel + 1, true
			}
		}
		return 2, true
	case 'q':
		if flags.UV() && i+2 < len(pattern) && pattern[i+2] == '{' {
			closeRel := strings.IndexByte(pattern[i+3:], '}')
			if closeRel >= 0 {
				return 3 + closeRel + 1, true
			}
		}
		return 2, true
	}
	// Generic two-byte escape (covers identity, \d, \w, ...).
	// If the next is a multi-byte UTF-8 rune, consume its full width.
	_, w := utf8.DecodeRuneInString(pattern[i+1:])
	if w == 0 {
		return 2, true
	}
	return 1 + w, true
}

// ---------------------------------------------------------------------------
// Character class body parser (layer 2)
// ---------------------------------------------------------------------------

// ParseRegexCharacterClass parses a character class starting at pattern[start]
// (which must be `[`). It returns the flat element list (ranges expanded into
// min/max; nested classes and set operators emitted as breakers at the
// position where they appear) and the byte index just past the closing `]`.
//
// The input range [start, end) is trusted to be well-formed (e.g. from a
// prior call to IterateRegexCharacterClasses).
//
// Note: nested v-flag classes themselves are not recursed into by this
// function — they appear only as RegexCharBreaker. Callers that want to scan
// nested classes should iterate them via IterateRegexCharacterClasses and
// call ParseRegexCharacterClass on each separately.
func ParseRegexCharacterClass(pattern string, start int, flags RegexFlags) ([]RegexCharElement, int, bool) {
	if start >= len(pattern) || pattern[start] != '[' {
		return nil, start, false
	}
	i := start + 1
	if i < len(pattern) && pattern[i] == '^' {
		i++
	}

	var out []RegexCharElement
	pendingIdx := -1 // index into `out` of the last RegexCharSingle (candidate for range start)

	for i < len(pattern) {
		c := pattern[i]
		switch {
		case c == ']':
			return out, i + 1, true

		case c == '\\':
			el, step, ok := readClassEscape(pattern, i, flags)
			if !ok {
				return nil, start, false
			}
			if el.Kind == RegexCharBreaker {
				out = append(out, el)
				pendingIdx = -1
			} else {
				// RegexCharSingle (possibly with companion astral surrogate entry below)
				out = append(out, el)
				pendingIdx = len(out) - 1
				// Under non-u/v mode, a raw astral character or a `\u{H}` with H>0xFFFF
				// should emit TWO units. But class escapes only produce 1 unit per call;
				// astral raw chars are handled in the `default` case below, and `\u{H}`
				// under u/v produces the combined code point (one element). Under non-u,
				// `\u{H}` is not recognized at all (treated as identity `u`).
			}
			i += step

		case c == '[' && flags.UnicodeSets:
			// v-flag nested class — emit as breaker; caller recurses via
			// IterateRegexCharacterClasses to parse nested contents.
			nestedEnd, ok := ClassEnd(pattern, i, flags)
			if !ok {
				return nil, start, false
			}
			out = append(out, RegexCharElement{Kind: RegexCharBreaker, Start: i, End: nestedEnd})
			i = nestedEnd
			pendingIdx = -1

		case (c == '-' || c == '&') && flags.UnicodeSets && i+1 < len(pattern) && pattern[i+1] == c:
			// v-flag set operator `--` or `&&` — breaker
			out = append(out, RegexCharElement{Kind: RegexCharBreaker, Start: i, End: i + 2})
			i += 2
			pendingIdx = -1

		case c == '-' && pendingIdx >= 0 && i+1 < len(pattern) && pattern[i+1] != ']':
			// Range: combine pending with next element.
			i++
			var nextEl RegexCharElement
			var step int
			var ok bool
			if pattern[i] == '\\' {
				nextEl, step, ok = readClassEscape(pattern, i, flags)
				if !ok {
					return nil, start, false
				}
			} else {
				nextEl, step, ok = readRawClassChar(pattern, i, flags)
				if !ok {
					return nil, start, false
				}
			}
			if nextEl.Kind != RegexCharSingle {
				// Under u flag, `set-to-X` or `X-to-set` is a syntax error;
				// ESLint's regexpp would throw. We treat this as malformed
				// and abort class parsing (caller will skip the regex).
				return nil, start, false
			}
			minEl := out[pendingIdx]
			rangeEl := RegexCharElement{
				Kind:        RegexCharRange,
				Value:       minEl.Value,
				IsUBrace:    minEl.IsUBrace,
				Max:         nextEl.Value,
				MaxIsUBrace: nextEl.IsUBrace,
				Start:       minEl.Start,
				End:         nextEl.End,
			}
			out[pendingIdx] = rangeEl
			// If the min endpoint came from a raw astral in non-u mode
			// (which would logically expand to a surrogate pair), the range
			// representation is already degenerate; callers in this project
			// only inspect endpoints, so we don't split here.
			i += step
			pendingIdx = -1

		default:
			el, step, ok := readRawClassChar(pattern, i, flags)
			if !ok {
				return nil, start, false
			}
			if el.Kind == RegexCharSingle {
				out = append(out, el)
				pendingIdx = len(out) - 1
				// Note: we do NOT split raw astrals into surrogate pairs here
				// — callers (e.g. the misleading-character-class rule) decide
				// per-flag whether to expand them, based on the rule's own
				// semantic needs. See IsLoneSurrogate on the element.
			}
			i += step
		}
	}
	return nil, start, false // unterminated
}

// readClassEscape parses a `\`-prefixed token inside a character class. It
// emits one RegexCharElement: either RegexCharSingle or RegexCharBreaker.
func readClassEscape(pattern string, i int, flags RegexFlags) (RegexCharElement, int, bool) {
	if i+1 >= len(pattern) {
		return RegexCharElement{}, 0, false
	}
	next := pattern[i+1]
	switch next {
	case 'd', 'D', 'w', 'W', 's', 'S', 'b', 'B':
		// Note: inside a class, `\b` means U+0008 (backspace), not word
		// boundary. ESLint treats it as a single character. `\B` inside a
		// class is a syntax error under u mode; under non-u it's identity.
		// For our purposes, we treat `\b` as a RegexCharSingle with value 8
		// (backspace) and `\B` ... hmm.
		//
		// Actually: regexpp treats `\b` in class as Character(0x08) and `\B`
		// in class as a syntax error in u mode but identity in non-u.
		// `\d`/`\D`/`\w`/`\W`/`\s`/`\S` are CharacterSet (breaker).
		if next == 'b' {
			return RegexCharElement{
				Kind: RegexCharSingle, Value: 0x08, Start: i, End: i + 2,
			}, 2, true
		}
		if next == 'B' {
			if flags.UV() {
				return RegexCharElement{}, 0, false // u mode syntax error
			}
			// Under non-u, identity escape.
			return RegexCharElement{
				Kind: RegexCharSingle, Value: 'B', Start: i, End: i + 2,
			}, 2, true
		}
		return RegexCharElement{Kind: RegexCharBreaker, Start: i, End: i + 2}, 2, true
	case 'p', 'P':
		if flags.UV() && i+2 < len(pattern) && pattern[i+2] == '{' {
			closeRel := strings.IndexByte(pattern[i+3:], '}')
			if closeRel < 0 {
				return RegexCharElement{}, 0, false
			}
			end := i + 3 + closeRel + 1
			return RegexCharElement{Kind: RegexCharBreaker, Start: i, End: end}, end - i, true
		}
		// Non-u mode: identity escape of `p`/`P`.
		return RegexCharElement{
			Kind: RegexCharSingle, Value: uint32(next), Start: i, End: i + 2,
		}, 2, true
	case 'q':
		if flags.UV() && i+2 < len(pattern) && pattern[i+2] == '{' {
			closeRel := strings.IndexByte(pattern[i+3:], '}')
			if closeRel < 0 {
				return RegexCharElement{}, 0, false
			}
			end := i + 3 + closeRel + 1
			return RegexCharElement{Kind: RegexCharBreaker, Start: i, End: end}, end - i, true
		}
		return RegexCharElement{
			Kind: RegexCharSingle, Value: 'q', Start: i, End: i + 2,
		}, 2, true
	case 'x':
		if i+3 < len(pattern) && isHex(pattern[i+2]) && isHex(pattern[i+3]) {
			v := parseHexStr(pattern[i+2 : i+4])
			return RegexCharElement{
				Kind: RegexCharSingle, Value: v, Start: i, End: i + 4,
			}, 4, true
		}
		return RegexCharElement{
			Kind: RegexCharSingle, Value: 'x', Start: i, End: i + 2,
		}, 2, true
	case 'u':
		if i+2 < len(pattern) && pattern[i+2] == '{' {
			if !flags.UV() {
				// Non-u mode: treat `\u` as identity `u`.
				return RegexCharElement{
					Kind: RegexCharSingle, Value: 'u', Start: i, End: i + 2,
				}, 2, true
			}
			closeRel := strings.IndexByte(pattern[i+3:], '}')
			if closeRel < 0 {
				return RegexCharElement{}, 0, false
			}
			hex := pattern[i+3 : i+3+closeRel]
			end := i + 3 + closeRel + 1
			if hex == "" || !allHexStr(hex) {
				return RegexCharElement{}, 0, false
			}
			return RegexCharElement{
				Kind: RegexCharSingle, Value: parseHexStr(hex), IsUBrace: true,
				Start: i, End: end,
			}, end - i, true
		}
		if i+5 < len(pattern) && allHexStr(pattern[i+2:i+6]) {
			hi := parseHexStr(pattern[i+2 : i+6])
			// Surrogate pair `\uHHHH\uHHHH` under u/v collapses to one element
			// with an astral value.
			if flags.UV() && hi >= 0xD800 && hi <= 0xDBFF && i+11 < len(pattern) &&
				pattern[i+6] == '\\' && pattern[i+7] == 'u' && allHexStr(pattern[i+8:i+12]) {
				lo := parseHexStr(pattern[i+8 : i+12])
				if lo >= 0xDC00 && lo <= 0xDFFF {
					cp := 0x10000 + (hi-0xD800)*0x400 + (lo - 0xDC00)
					return RegexCharElement{
						Kind: RegexCharSingle, Value: cp, Start: i, End: i + 12,
					}, 12, true
				}
			}
			return RegexCharElement{
				Kind: RegexCharSingle, Value: hi, Start: i, End: i + 6,
			}, 6, true
		}
		return RegexCharElement{
			Kind: RegexCharSingle, Value: 'u', Start: i, End: i + 2,
		}, 2, true
	case 'n':
		return RegexCharElement{Kind: RegexCharSingle, Value: '\n', Start: i, End: i + 2}, 2, true
	case 't':
		return RegexCharElement{Kind: RegexCharSingle, Value: '\t', Start: i, End: i + 2}, 2, true
	case 'r':
		return RegexCharElement{Kind: RegexCharSingle, Value: '\r', Start: i, End: i + 2}, 2, true
	case 'v':
		return RegexCharElement{Kind: RegexCharSingle, Value: '\v', Start: i, End: i + 2}, 2, true
	case 'f':
		return RegexCharElement{Kind: RegexCharSingle, Value: '\f', Start: i, End: i + 2}, 2, true
	case '0':
		// In u/v mode `\0` is always NUL (no octal extension). In non-u
		// mode, if followed by a digit it's a legacy octal — but legacy
		// octals aren't meaningful for this project's detectors, so treat
		// as NUL for simplicity.
		return RegexCharElement{Kind: RegexCharSingle, Value: 0, Start: i, End: i + 2}, 2, true
	case 'c':
		if i+2 < len(pattern) {
			return RegexCharElement{
				Kind: RegexCharSingle, Value: uint32(pattern[i+2]) & 0x1F,
				Start: i, End: i + 3,
			}, 3, true
		}
		return RegexCharElement{}, 0, false
	}
	// Identity escape — value is the (possibly multi-byte) character after `\`.
	r, w := utf8.DecodeRuneInString(pattern[i+1:])
	if w == 0 {
		return RegexCharElement{}, 0, false
	}
	// Under u/v mode, identity escape of a letter/digit is a syntax error.
	// We relax this: let caller see the element and skip the regex if it
	// becomes problematic. (Most production code won't hit this.)
	return RegexCharElement{
		Kind: RegexCharSingle, Value: uint32(r), Start: i, End: i + 1 + w,
	}, 1 + w, true
}

// readRawClassChar reads a non-`\`-prefixed character inside a class body
// (i.e. a raw character at pattern[i]). Raw astral in non-u/v mode emits
// the HIGH surrogate only — callers handle the low-surrogate companion by
// inspecting IsLoneSurrogate and stepping via a subsequent call.
//
// To keep the protocol simple, ParseRegexCharacterClass calls this once and
// trusts the single element returned; we do NOT split raw astrals here. The
// misleading-character-class rule accepts this because it gets raw astrals
// only in string-literal contexts, where the layer-3 parser already does the
// UTF-16 surrogate split.
func readRawClassChar(pattern string, i int, flags RegexFlags) (RegexCharElement, int, bool) {
	r, w := utf8.DecodeRuneInString(pattern[i:])
	if w == 0 {
		return RegexCharElement{}, 0, false
	}
	return RegexCharElement{
		Kind: RegexCharSingle, Value: uint32(r), Start: i, End: i + w,
	}, w, true
}
