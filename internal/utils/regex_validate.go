package utils

import (
	esregexp "github.com/web-infra-dev/rslint/internal/utils/ecmascript/regexp"
)

// IsValidRegexPattern reports whether pattern parses cleanly under flags, the
// way JavaScript's RegExp constructor would (a try/catch around parsing, not
// a match attempt). Under the u/v flags this combines three checks:
//
//   - IterateRegexCharacterClasses catches unterminated classes, including
//     v-flag nested classes such as `[[a]` (valid `u` syntax, a SyntaxError
//     under `v`).
//   - regexp2 compile catches other structural errors (unclosed `(`, unmatched
//     quantifier under Unicode mode, bad hex escapes, etc.). esregexp has no
//     `v`-flag token, so `v` validation reuses the `u` compile — the two
//     flags share the same non-class grammar.
//   - A narrow u-flag identity-escape check for the handful of escapes
//     regexp2 accepts but ES-u-mode rejects (`\a`, `\9`, …), plus a scan for
//     the syntax characters regexp2 reads as literals.
//
// If ANY check fails the pattern is treated as unparsable — matching
// JavaScript's own parse-or-reject behavior.
func IsValidRegexPattern(pattern string, flags RegexFlags) bool {
	if flags.UV() {
		if !IterateRegexCharacterClasses(pattern, flags, func(int, int) {}) {
			return false
		}
	}

	compileFlags := ""
	if flags.UV() {
		compileFlags = "u"
	}
	if _, err := esregexp.Compile(pattern, compileFlags); err != nil {
		return false
	}

	if flags.UV() {
		if hasInvalidIdentityEscapeForUFlag(pattern) {
			return false
		}
		if hasInvalidSyntaxCharForUFlag(pattern, flags) {
			return false
		}
	}
	return true
}

// hasInvalidSyntaxCharForUFlag reports whether the pattern contains an
// unescaped `{`, `}` or `]` outside a character class. Under the u/v flag all
// three are SyntaxCharacters and legal only in their structural roles — `{`
// and `}` as a `{n}` / `{n,}` / `{n,m}` quantifier, `]` as a class terminator
// — but regexp2 accepts each as a literal (its .NET lineage). Escapes and
// character classes are skipped wholesale through the flag-aware scanners, so
// v-flag nested classes are not mistaken for a stray `]`.
func hasInvalidSyntaxCharForUFlag(pattern string, flags RegexFlags) bool {
	i := 0
	for i < len(pattern) {
		switch pattern[i] {
		case '\\':
			step, ok := SkipPatternEscape(pattern, i, flags)
			if !ok {
				// Unterminated escape — let compile/scan report it.
				return false
			}
			i += step
		case '[':
			end, ok := ClassEnd(pattern, i, flags)
			if !ok {
				// Unterminated class — let compile/scan report it.
				return false
			}
			i = end
		case '{':
			end, ok := quantifierEnd(pattern, i)
			if !ok {
				return true
			}
			i = end
		case '}', ']':
			return true
		default:
			i++
		}
	}
	return false
}

// quantifierEnd returns the byte index just past the `}` of the
// `{n}` / `{n,}` / `{n,m}` quantifier opening at pattern[start], or ok=false
// when the brace does not open one.
func quantifierEnd(pattern string, start int) (int, bool) {
	i := start + 1
	digits := 0
	for i < len(pattern) && pattern[i] >= '0' && pattern[i] <= '9' {
		i++
		digits++
	}
	if digits == 0 {
		return start, false
	}
	if i < len(pattern) && pattern[i] == ',' {
		i++
		for i < len(pattern) && pattern[i] >= '0' && pattern[i] <= '9' {
			i++
		}
	}
	if i < len(pattern) && pattern[i] == '}' {
		return i + 1, true
	}
	return start, false
}

// hasInvalidIdentityEscapeForUFlag scans for escapes that ECMAScript u/v mode
// rejects but regexp2 accepts. Conservative: on malformed input it returns
// false (i.e. defers to regexp2's verdict) rather than inventing an error.
func hasInvalidIdentityEscapeForUFlag(pattern string) bool {
	i := 0
	for i < len(pattern) {
		c := pattern[i]
		if c != '\\' || i+1 >= len(pattern) {
			i++
			continue
		}
		next := pattern[i+1]
		switch next {
		// Recognized single-letter escapes.
		case 'd', 'D', 'w', 'W', 's', 'S', 'b', 'B', 'n', 't', 'r', 'v', 'f', '0',
			'x', 'u', 'c', 'p', 'P', 'k', 'q',
			// SyntaxCharacter / `/` — legal identity escapes under u.
			'^', '$', '.', '*', '+', '?', '(', ')', '[', ']', '{', '}', '|', '\\', '/':
			i += 2
			continue
		}
		// Decimal backreference (\1..\9) is legal.
		if next >= '1' && next <= '9' {
			i += 2
			continue
		}
		// Any letter/digit identity escape not recognized above is illegal under u.
		if (next >= 'a' && next <= 'z') || (next >= 'A' && next <= 'Z') {
			return true
		}
		i += 2
	}
	return false
}
