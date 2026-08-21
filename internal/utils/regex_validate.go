package utils

import (
	"strings"

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
//     regexp2 accepts but ES-u-mode rejects (`\a`, `\9`, …), plus an
//     unmatched-`{` check for a brace regexp2 reads as literal.
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
		if hasUnmatchedBraceForUFlag(pattern) {
			return false
		}
	}
	return true
}

// hasUnmatchedBraceForUFlag reports whether the pattern contains a literal `{`
// that is not part of a valid `{n}` / `{n,}` / `{n,m}` quantifier or a
// recognized `\u{...}` / `\p{...}` / `\q{...}` escape. Under the u/v flag this
// is a SyntaxError per ECMAScript, but regexp2 accepts it (its .NET lineage
// treats a bare `{` as literal). The scan is outside character classes only —
// inside a class, `{` is always literal.
func hasUnmatchedBraceForUFlag(pattern string) bool {
	inClass := false
	i := 0
	for i < len(pattern) {
		c := pattern[i]
		if c == '\\' {
			if i+1 >= len(pattern) {
				return false
			}
			next := pattern[i+1]
			if (next == 'u' || next == 'p' || next == 'P' || next == 'q') && i+2 < len(pattern) && pattern[i+2] == '{' {
				if end := strings.IndexByte(pattern[i+2:], '}'); end != -1 {
					i += 2 + end + 1
					continue
				}
				// Unterminated brace escape — let compile/scan report it.
				return false
			}
			i += 2
			continue
		}
		if c == '[' && !inClass {
			inClass = true
			i++
			continue
		}
		if c == ']' && inClass {
			inClass = false
			i++
			continue
		}
		if inClass {
			i++
			continue
		}
		if c == '{' && !looksLikeQuantifier(pattern, i) {
			return true
		}
		i++
	}
	return false
}

// looksLikeQuantifier reports whether pattern[start] opens a valid
// `{n}` / `{n,}` / `{n,m}` quantifier.
func looksLikeQuantifier(pattern string, start int) bool {
	i := start + 1
	digits := 0
	for i < len(pattern) && pattern[i] >= '0' && pattern[i] <= '9' {
		i++
		digits++
	}
	if digits == 0 {
		return false
	}
	if i < len(pattern) && pattern[i] == ',' {
		i++
		for i < len(pattern) && pattern[i] >= '0' && pattern[i] <= '9' {
			i++
		}
	}
	return i < len(pattern) && pattern[i] == '}'
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
