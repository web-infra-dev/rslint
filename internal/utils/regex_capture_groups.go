package utils

import "unicode/utf8"

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
func RegexCapturingGroups(pattern string, flags RegexFlags) (groups []RegexCapturingGroup, ok bool) {
	pos := 0
	if !scanRegexAlternatives(pattern, flags, &pos, &groups, false) {
		return nil, false
	}
	if pos != len(pattern) {
		return nil, false
	}
	return groups, true
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
	c := pattern[*pos]
	switch c {
	case '(':
		if !scanRegexGroup(pattern, flags, pos, groups) {
			return false
		}
	case '[':
		end, ok := ClassEnd(pattern, *pos, flags)
		if !ok {
			return false
		}
		// Under v-mode, a bare `\q` inside a class (not starting a `\q{...}`
		// string-literal escape) is a SyntaxError, not an identity escape.
		// ClassEnd's own escape-skipping doesn't distinguish this — check it
		// separately so a "how do you spell an invalid v-mode class" pattern
		// isn't silently accepted as valid.
		if flags.UnicodeSets && classHasBareQEscape(pattern, *pos, end) {
			return false
		}
		*pos = end
	case '\\':
		step, ok := skipPatternLevelEscape(pattern, *pos, flags)
		if !ok {
			return false
		}
		*pos += step
	case '*', '+', '?':
		// Standalone quantifier without an operand. Always a syntax error
		// here (Annex B's lenient legacy-quantifier fallback is not
		// modeled); the caller bails out on the whole pattern.
		return false
	case '{':
		// Under u/v mode a standalone `{` (one not part of a `{n}`/`{n,m}`
		// quantifier) is a syntax error. Outside u/v mode it's a literal.
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
	skipRegexQuantifier(pattern, pos)
	return true
}

// skipPatternLevelEscape returns how many bytes the `\`-prefixed escape at
// pattern[i] consumes when it appears outside a character class. Two escapes
// read differently there than they do inside one, so they can't go through
// SkipPatternEscape:
//
//   - `\c` opens a control escape only when a ControlLetter follows. Under u/v
//     anything else is a SyntaxError; outside u/v, Annex B reads the backslash
//     alone as a literal `\` and leaves the `c` to be scanned as an ordinary
//     character, so only the backslash is consumed here.
//   - `\q{...}` is a v-mode string disjunction, which exists only inside a
//     class. Outside one, `\q` under u/v isn't a valid escape at all.
func skipPatternLevelEscape(pattern string, i int, flags RegexFlags) (int, bool) {
	if i+1 >= len(pattern) {
		return 0, false
	}
	switch pattern[i+1] {
	case 'c':
		if i+2 < len(pattern) && isRegexControlLetter(pattern[i+2]) {
			return 3, true
		}
		if flags.UV() {
			return 0, false
		}
		return 1, true
	case 'q':
		if flags.UV() {
			return 0, false
		}
	}
	return SkipPatternEscape(pattern, i, flags)
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

func scanRegexGroup(pattern string, flags RegexFlags, pos *int, groups *[]RegexCapturingGroup) bool {
	start := *pos
	*pos++ // consume '('

	isCapturing := false
	name := ""

	if *pos < len(pattern) && pattern[*pos] == '?' {
		if *pos+1 >= len(pattern) {
			return false
		}
		switch pattern[*pos+1] {
		case ':':
			*pos += 2
		case '=', '!':
			*pos += 2
		case '<':
			if *pos+2 >= len(pattern) {
				return false
			}
			switch pattern[*pos+2] {
			case '=', '!':
				*pos += 3
			default:
				// (?<name>...)
				*pos += 2 // consume "?<"
				nameEnd := *pos
				for nameEnd < len(pattern) && pattern[nameEnd] != '>' {
					nameEnd++
				}
				if nameEnd >= len(pattern) || nameEnd == *pos {
					return false
				}
				name = pattern[*pos:nameEnd]
				*pos = nameEnd + 1
				isCapturing = true
			}
		default:
			// ES2025 modifier group: (?ims-ims:...) / (?i:...) / (?-i:...)
			if !skipRegexModifierGroupHeader(pattern, pos) {
				return false
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
		return false
	}
	if *pos >= len(pattern) || pattern[*pos] != ')' {
		return false
	}
	*pos++ // consume ')'

	if groupIdx >= 0 {
		(*groups)[groupIdx].End = *pos
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

// classHasBareQEscape reports whether pattern[start:end] (a `[...]` class,
// as bounded by ClassEnd) contains a `\q` that doesn't open a `\q{...}`
// string-literal escape. `\\q` (an escaped backslash followed by a literal
// `q`) doesn't count — this walks escapes the same way ClassEnd/
// SkipPatternEscape do so it isn't misled by one.
func classHasBareQEscape(pattern string, start, end int) bool {
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
		if i+1 < end && pattern[i+1] == 'q' && (i+2 >= end || pattern[i+2] != '{') {
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

func isRegexControlLetter(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z'
}

func isRegexDigit(c byte) bool { return c >= '0' && c <= '9' }
