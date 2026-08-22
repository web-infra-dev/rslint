package utils

import (
	"strings"
	"unicode/utf8"
)

// RegexCapturingGroup is one capturing group found by RegexCapturingGroups:
// either a plain `(...)` group or a named `(?<name>...)` group. Start is the
// byte offset of the group's opening `(` within the pattern; End is one past
// the matching `)`.
type RegexCapturingGroup struct {
	Start int
	End   int
	Name  string // "" for an unnamed group
}

// RegexCapturingGroups walks an ECMAScript regex pattern and returns every
// capturing group it contains, in source order. It is not a full regex
// parser — it recognizes just enough syntax (groups, character classes,
// escapes, quantifiers, ES2025 modifier groups) to correctly skip past
// constructs it doesn't need while finding every `(` that opens a capturing
// group. ok is false when the pattern fails to parse, mirroring how
// regexpp/the host engine would reject it; callers should treat that the same
// as "no capturing groups" (skip the pattern entirely).
//
// A handful of checks that need tables or a second model of the pattern are
// left out, so those patterns come back ok even though the host engine rejects
// them: unknown `\p{...}` property names, a group name repeated within one
// alternative, and v-mode's rule that a class `-` outside a range be escaped.
func RegexCapturingGroups(pattern string, flags RegexFlags) (groups []RegexCapturingGroup, ok bool) {
	pos := 0
	if !scanRegexAlternatives(pattern, flags, &pos, &groups, false) {
		return nil, false
	}
	if pos != len(pattern) {
		return nil, false
	}
	if flags.UV() && !unicodeBackrefsResolve(pattern, flags, groups) {
		return nil, false
	}
	return groups, true
}

// unicodeBackrefsResolve reports whether every backreference in an already
// well-formed u/v-mode pattern points at a group that exists. Outside u/v the
// check doesn't apply: a numeric escape with no matching group falls back to an
// octal or identity escape, and `\k` isn't a backreference at all unless the
// pattern has named groups.
func unicodeBackrefsResolve(pattern string, flags RegexFlags, groups []RegexCapturingGroup) bool {
	for i := 0; i < len(pattern); {
		switch pattern[i] {
		case '[':
			end, ok := ClassEnd(pattern, i, flags)
			if !ok {
				return false
			}
			i = end
		case '\\':
			step, ok := skipPatternLevelEscape(pattern, i, flags)
			if !ok {
				return false
			}
			switch next := pattern[i+1]; {
			case next >= '1' && next <= '9':
				n := 0
				for _, digit := range []byte(pattern[i+1 : i+step]) {
					n = n*10 + int(digit-'0')
					if n > len(groups) {
						return false
					}
				}
			case next == 'k':
				// skipPatternLevelEscape only accepts `\k` as `\k<name>`.
				if !hasGroupNamed(groups, pattern[i+3:i+step-1]) {
					return false
				}
			}
			i += step
		default:
			_, w := utf8.DecodeRuneInString(pattern[i:])
			if w == 0 {
				return false
			}
			i += w
		}
	}
	return true
}

func hasGroupNamed(groups []RegexCapturingGroup, name string) bool {
	for _, group := range groups {
		if group.Name == name {
			return true
		}
	}
	return false
}

func scanRegexAlternatives(pattern string, flags RegexFlags, pos *int, groups *[]RegexCapturingGroup, expectClose bool) bool {
	for *pos < len(pattern) {
		c := pattern[*pos]
		if c == ')' {
			return expectClose
		}
		if c == '|' {
			*pos++
			continue
		}
		if !scanRegexTerm(pattern, flags, pos, groups) {
			return false
		}
	}
	return !expectClose
}

func scanRegexTerm(pattern string, flags RegexFlags, pos *int, groups *[]RegexCapturingGroup) bool {
	// Assertions can't carry a quantifier. Annex B makes an exception for
	// lookahead, which u/v mode takes back.
	quantifiable := true
	c := pattern[*pos]
	switch c {
	case '(':
		groupQuantifiable, ok := scanRegexGroup(pattern, flags, pos, groups)
		if !ok {
			return false
		}
		quantifiable = groupQuantifiable
	case '^', '$':
		quantifiable = false
		*pos++
	case '[':
		end, ok := ClassEnd(pattern, *pos, flags)
		if !ok {
			return false
		}
		// `\q` inside a class is a SyntaxError, not an identity escape: under v
		// unless it opens a `\q{...}` string literal, and under u always, since
		// string literals are v-only. ClassEnd's own escape-skipping doesn't
		// distinguish this — check it separately so a "how do you spell an
		// invalid class" pattern isn't silently accepted as valid.
		if flags.UV() && classHasInvalidQEscape(pattern, *pos, end, flags.UnicodeSets) {
			return false
		}
		*pos = end
	case '\\':
		step, ok := skipPatternLevelEscape(pattern, *pos, flags)
		if !ok {
			return false
		}
		if next := pattern[*pos+1]; step > 1 && (next == 'b' || next == 'B') {
			quantifiable = false
		}
		*pos += step
	case '*', '+', '?':
		// Standalone quantifier without an operand. Always a syntax error
		// here (Annex B's lenient legacy-quantifier fallback is not
		// modeled); the caller bails out on the whole pattern.
		return false
	case '{':
		// A `{n}`/`{n,m}` reached at term position has no operand to quantify,
		// which is a syntax error in every mode. A `{` that doesn't spell a
		// quantifier is a syntax error under u/v and a literal outside it.
		probe := *pos
		skipRegexQuantifier(pattern, &probe)
		if probe != *pos || flags.UV() {
			return false
		}
		*pos++
	case '}', ']':
		// A closer with no opener: a literal outside u/v, rejected under it.
		if flags.UV() {
			return false
		}
		*pos++
	default:
		_, w := utf8.DecodeRuneInString(pattern[*pos:])
		if w == 0 {
			return false
		}
		*pos += w
	}
	if quantifiable {
		skipRegexQuantifier(pattern, pos)
		return true
	}
	probe := *pos
	skipRegexQuantifier(pattern, &probe)
	return probe == *pos
}

// skipPatternLevelEscape returns how many bytes the `\`-prefixed escape at
// pattern[i] consumes when it appears outside a character class.
//
// Under u/v the escape grammar is closed — every form is spelled out and
// anything else is a SyntaxError — so those patterns get their own strict
// reading. Outside u/v the escapes are the permissive Annex B set that
// SkipPatternEscape models, with one exception: `\c` opens a control escape
// only when a ControlLetter follows, and otherwise the backslash stands alone
// as a literal `\`, leaving the `c` to be scanned as an ordinary character.
func skipPatternLevelEscape(pattern string, i int, flags RegexFlags) (int, bool) {
	if i+1 >= len(pattern) {
		return 0, false
	}
	if flags.UV() {
		return skipUnicodePatternEscape(pattern, i)
	}
	if pattern[i+1] == 'c' && (i+2 >= len(pattern) || !isASCIILetter(pattern[i+2])) {
		return 1, true
	}
	return SkipPatternEscape(pattern, i, flags)
}

// skipUnicodePatternEscape is skipPatternLevelEscape's u/v-mode reading: it
// accepts only the escapes the ES grammar lists for an AtomEscape and rejects
// everything else, where Annex B would have fallen back to an identity escape.
// `\q{...}` lands in that rejected set — it's a v-mode string disjunction, and
// those exist only inside a character class.
//
// Property names inside `\p{...}` aren't validated, so an unknown one is still
// accepted here.
func skipUnicodePatternEscape(pattern string, i int) (int, bool) {
	rest := pattern[i+2:]
	switch c := pattern[i+1]; c {
	case 'b', 'B', 'd', 'D', 's', 'S', 'w', 'W', 'f', 'n', 'r', 't', 'v',
		'^', '$', '\\', '.', '*', '+', '?', '(', ')', '[', ']', '{', '}', '|', '/':
		return 2, true
	case 'c':
		if len(rest) > 0 && isASCIILetter(rest[0]) {
			return 3, true
		}
	case '0':
		// `\0` is the NUL escape only when no digit follows; `\01` is a legacy
		// octal escape, which u/v mode doesn't have.
		if len(rest) == 0 || !isRegexDigit(rest[0]) {
			return 2, true
		}
	case '1', '2', '3', '4', '5', '6', '7', '8', '9':
		n := 0
		for n < len(rest) && isRegexDigit(rest[n]) {
			n++
		}
		return 2 + n, true
	case 'x':
		if len(rest) >= 2 && isHex(rest[0]) && isHex(rest[1]) {
			return 4, true
		}
	case 'u':
		if len(rest) > 0 && rest[0] == '{' {
			n := 1
			for n < len(rest) && isHex(rest[n]) {
				n++
			}
			if n > 1 && n < len(rest) && rest[n] == '}' {
				return 3 + n, true
			}
			break
		}
		if len(rest) >= 4 && allHexStr(rest[:4]) {
			return 6, true
		}
	case 'p', 'P':
		if len(rest) > 0 && rest[0] == '{' {
			if closeRel := strings.IndexByte(rest[1:], '}'); closeRel >= 0 {
				return 4 + closeRel, true
			}
		}
	case 'k':
		if len(rest) > 0 && rest[0] == '<' {
			if closeRel := strings.IndexByte(rest[1:], '>'); closeRel > 0 {
				return 4 + closeRel, true
			}
		}
	}
	return 0, false
}

func skipRegexQuantifier(pattern string, pos *int) {
	if *pos >= len(pattern) {
		return
	}
	switch pattern[*pos] {
	case '?', '*', '+':
		*pos++
	case '{':
		save := *pos
		i := *pos + 1
		nStart := i
		for i < len(pattern) && isRegexDigit(pattern[i]) {
			i++
		}
		if i == nStart {
			return
		}
		if i < len(pattern) && pattern[i] == ',' {
			i++
			for i < len(pattern) && isRegexDigit(pattern[i]) {
				i++
			}
		}
		if i < len(pattern) && pattern[i] == '}' {
			*pos = i + 1
		} else {
			*pos = save
			return
		}
	default:
		return
	}
	if *pos < len(pattern) && pattern[*pos] == '?' {
		*pos++
	}
}

// scanRegexGroup consumes the `(...)` starting at pattern[*pos], recording it
// when it captures. quantifiable is false for the assertion forms a quantifier
// may not follow.
func scanRegexGroup(pattern string, flags RegexFlags, pos *int, groups *[]RegexCapturingGroup) (quantifiable, ok bool) {
	start := *pos
	*pos++ // consume '('

	isCapturing := false
	quantifiable = true
	name := ""

	if *pos < len(pattern) && pattern[*pos] == '?' {
		if *pos+1 >= len(pattern) {
			return false, false
		}
		switch pattern[*pos+1] {
		case ':':
			*pos += 2
		case '=', '!':
			// Annex B allows a quantifier on lookahead; u/v mode doesn't.
			quantifiable = !flags.UV()
			*pos += 2
		case '<':
			if *pos+2 >= len(pattern) {
				return false, false
			}
			switch pattern[*pos+2] {
			case '=', '!':
				quantifiable = false
				*pos += 3
			default:
				// (?<name>...)
				*pos += 2 // consume "?<"
				nameEnd := *pos
				for nameEnd < len(pattern) && pattern[nameEnd] != '>' {
					nameEnd++
				}
				if nameEnd >= len(pattern) {
					return false, false
				}
				name = pattern[*pos:nameEnd]
				if !isRegexGroupName(name) {
					return false, false
				}
				*pos = nameEnd + 1
				isCapturing = true
			}
		default:
			// ES2025 modifier group: (?ims-ims:...) / (?i:...) / (?-i:...)
			if !skipRegexModifierGroupHeader(pattern, pos) {
				return false, false
			}
		}
	} else {
		isCapturing = true
	}

	groupIdx := -1
	if isCapturing {
		*groups = append(*groups, RegexCapturingGroup{Start: start, Name: name})
		groupIdx = len(*groups) - 1
	}

	if !scanRegexAlternatives(pattern, flags, pos, groups, true) {
		return false, false
	}
	if *pos >= len(pattern) || pattern[*pos] != ')' {
		return false, false
	}
	*pos++ // consume ')'

	if groupIdx >= 0 {
		(*groups)[groupIdx].End = *pos
	}
	return quantifiable, true
}

// isRegexGroupName reports whether name could be an IdentifierName, the grammar
// a `(?<name>...)` group name has to follow. Only the ASCII range is checked —
// every non-ASCII rune is accepted rather than run against the ID_Start /
// ID_Continue tables, so an invalid name there slips through instead of
// suppressing a diagnostic on a pattern that's actually fine.
func isRegexGroupName(name string) bool {
	if name == "" {
		return false
	}
	for i := 0; i < len(name); {
		c := name[i]
		switch {
		case c >= utf8.RuneSelf:
			_, w := utf8.DecodeRuneInString(name[i:])
			if w == 0 {
				return false
			}
			i += w
			continue
		case c == '\\':
			// A `\u` escape spells the code point out; its value isn't checked.
			if i+1 >= len(name) || name[i+1] != 'u' {
				return false
			}
			i += 2
			continue
		case isASCIILetter(c), c == '$', c == '_':
		case isRegexDigit(c):
			if i == 0 {
				return false
			}
		default:
			return false
		}
		i++
	}
	return true
}

// skipRegexModifierGroupHeader consumes an ES2025 regex-modifier group header
// `?ims-ims:` / `?ims:` / `?-ims:` starting at pattern[*pos]=='?', leaving
// *pos positioned right after the ':'. Only `i`, `m`, and `s` are valid
// modifier flags, no flag may repeat within a set, the added and removed sets
// must be disjoint, and at least one of them must be non-empty.
func skipRegexModifierGroupHeader(pattern string, pos *int) bool {
	i := *pos + 1 // skip '?'
	added, ok := scanRegexModifierFlags(pattern, &i)
	if !ok {
		return false
	}
	removed := uint(0)
	if i < len(pattern) && pattern[i] == '-' {
		i++
		if removed, ok = scanRegexModifierFlags(pattern, &i); !ok {
			return false
		}
		if added|removed == 0 || added&removed != 0 {
			return false
		}
	}
	if i >= len(pattern) || pattern[i] != ':' {
		return false
	}
	*pos = i + 1
	return true
}

// scanRegexModifierFlags consumes the run of modifier flags at pattern[*pos]
// and returns them as a bit set, rejecting a flag that appears twice.
func scanRegexModifierFlags(pattern string, pos *int) (uint, bool) {
	seen := uint(0)
	for *pos < len(pattern) && isRegexModifierFlag(pattern[*pos]) {
		bit := uint(1) << (pattern[*pos] - 'a')
		if seen&bit != 0 {
			return 0, false
		}
		seen |= bit
		*pos++
	}
	return seen, true
}

// classHasInvalidQEscape reports whether pattern[start:end] (a `[...]` class,
// as bounded by ClassEnd) contains a `\q` that isn't a valid escape there:
// under v one that doesn't open a `\q{...}` string literal, and under u any at
// all. `\\q` (an escaped backslash followed by a literal `q`) doesn't count —
// this walks escapes the same way ClassEnd/SkipPatternEscape do so it isn't
// misled by one.
func classHasInvalidQEscape(pattern string, start, end int, unicodeSets bool) bool {
	for i := start; i < end; {
		if pattern[i] != '\\' {
			_, w := utf8.DecodeRuneInString(pattern[i:])
			if w == 0 {
				i++
			} else {
				i += w
			}
			continue
		}
		if i+1 < end && pattern[i+1] == 'q' && (!unicodeSets || i+2 >= end || pattern[i+2] != '{') {
			return true
		}
		step, ok := SkipPatternEscape(pattern, i, RegexFlags{UnicodeSets: true})
		if !ok {
			return false
		}
		i += step
	}
	return false
}

func isRegexModifierFlag(c byte) bool { return c == 'i' || c == 'm' || c == 's' }

// isASCIILetter matches a regex ControlLetter, and the ASCII half of an
// IdentifierName's character set.
func isASCIILetter(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z'
}

func isRegexDigit(c byte) bool { return c >= '0' && c <= '9' }
