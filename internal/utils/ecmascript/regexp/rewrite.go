package regexp

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"
)

// lineTerminators spells the characters ECMAScript reads as ending a line, in
// the form regexp2 parses inside a character class. It has to be the `\uXXXX`
// form: regexp2 reads neither `\x{2028}` nor a bare `\u{2028}`.
const lineTerminators = `\n\r\u2028\u2029`

const (
	// nonTerminator stands in for a `.`, which in JavaScript matches anything
	// but a line terminator. regexp2 follows .NET, where only `\n` is left out.
	nonTerminator = `[^` + lineTerminators + `]`
	// anyCharacter stands in for a `.` under the `s` flag, and for `[^]`.
	anyCharacter = `[\s\S]`
	// neverMatches stands in for `[]`, the class JavaScript reads as matching
	// nothing at all.
	neverMatches = `(?!)`
	// inputStart and inputEnd are the anchors a pattern without `m` wants:
	// the very start and the very end, where regexp2's own `^` and `$` would
	// also stop at a newline.
	inputStart = `\A`
	inputEnd   = `\z`
	// lineStart and lineEnd are the anchors a pattern with `m` wants.
	// regexp2's Multiline breaks on `\n` alone, where JavaScript breaks on all
	// four terminators, so the break is spelled out as a lookaround.
	lineStart = `(?:\A|(?<=[` + lineTerminators + `]))`
	lineEnd   = `(?:\z|(?=[` + lineTerminators + `]))`
)

// wordCharacters is the set ECMAScript builds `\w` and a word boundary out of.
// It is ASCII whatever the flags say, where .NET — and so regexp2 — reaches
// for a Unicode word set, which is why a boundary has to be spelled out.
//
// Under `u` and `i` together the set gains the two characters that fold into
// it: U+017F LATIN SMALL LETTER LONG S onto `s`, U+212A KELVIN SIGN onto `k`.
func wordCharacters(options rewriteOptions) string {
	if options.ignoreCase && options.unicode {
		return `[A-Za-z0-9_\u017f\u212a]`
	}
	return `[A-Za-z0-9_]`
}

// wordBoundary writes out `\b`, or `\B` when negated, as the pair of
// lookarounds ECMAScript defines it as, so the boundary lands where JavaScript
// puts it rather than where .NET's wider word set would.
func wordBoundary(negated bool, options rewriteOptions) string {
	word := wordCharacters(options)
	if negated {
		return `(?:(?<=` + word + `)(?=` + word + `)|(?<!` + word + `)(?!` + word + `))`
	}
	return `(?:(?<!` + word + `)(?=` + word + `)|(?<=` + word + `)(?!` + word + `))`
}

// hasControlLetter reports whether rest opens with the character a `\c` needs
// to be a control escape. A character class takes a digit or an underscore as
// well as a letter, which is the one place the two productions differ.
func hasControlLetter(rest string, inClass bool) bool {
	if rest == "" {
		return false
	}
	c := rest[0]
	if c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' {
		return true
	}
	return inClass && (c >= '0' && c <= '9' || c == '_')
}

// literalRune writes r so that regexp2 reads it as the character itself,
// whatever meaning that character carries on its own. A letter or a digit is
// written as it stands; everything else takes the `\uXXXX` form, which is what
// regexp2 parses both inside a character class and out.
func literalRune(r rune) string {
	switch {
	case r >= '0' && r <= '9', r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z':
		return string(r)
	case r > 0xFFFF:
		// Past what `\uXXXX` names, and nothing up there carries a meaning of
		// its own to escape.
		return string(r)
	}
	return fmt.Sprintf(`\u%04x`, r)
}

// rewriteOptions says which of a JavaScript regexp's flags the rewrite has to
// make up for, because regexp2 reads them differently or not at all.
type rewriteOptions struct {
	multiline  bool
	dotAll     bool
	ignoreCase bool
	unicode    bool
}

// rewrite translates a JavaScript regexp source into one regexp2 reads the
// same way. It walks the source rather than running a regexp over it, because
// what a character means depends on whether it is escaped and whether it sits
// inside a character class.
//
// The second result is false when the source uses a backreference under `i`.
// JavaScript compares a backreference by the same canonicalization it compares
// a literal by, and a widened pattern cannot say that — the caller falls back
// to regexp2's own case-insensitivity there, which is close but not exact.
func rewrite(source string, options rewriteOptions) (string, bool, error) {
	var out strings.Builder
	out.Grow(len(source))

	exact := true
	inClass := false
	classBodyStart := 0

	for i := 0; i < len(source); {
		r, size := utf8.DecodeRuneInString(source[i:])

		switch {
		case r == '\\':
			// `\k<name>` names a group. The name is not text to be matched,
			// so widening the letters in it would rewrite the reference.
			if name, nameSize, ok := namedBackreference(source[i:]); ok {
				if options.ignoreCase {
					exact = false
				}
				out.WriteString(name)
				i += nameSize
				continue
			}
			next, nextSize, ok := decodeEscape(source, i+size, options.unicode)
			if !ok {
				// A trailing backslash, which regexp2 can judge for itself.
				out.WriteString(source[i:])
				i = len(source)
				continue
			}
			end := i + size + nextSize
			if options.ignoreCase && isBackreference(source[i+size:i+size+nextSize]) {
				exact = false
			}
			// An escape that names one character is widened like a literal;
			// one that names a set — `\d`, `\w`, `\s` — is left alone.
			if options.ignoreCase && !inClass && next != utf8.RuneError {
				if class, widened := CaseClass(next, options.unicode); widened {
					out.WriteString(class)
					i = end
					continue
				}
			}
			// Annex B: a `\c` that no control letter follows is not an escape
			// at all, and the backslash and the `c` each stand for themselves.
			// .NET refuses the same text outright, which would drop whatever
			// option carried it.
			if nextSize == 1 && source[i+size] == 'c' && !hasControlLetter(source[end:], inClass) {
				out.WriteString(literalRune('\\'))
				out.WriteString("c")
				i = end
				continue
			}
			// `\b` and `\B` are boundaries, and only out here — inside a class
			// a `\b` is a backspace, which regexp2 already reads that way.
			if !inClass && nextSize == 1 && (source[i+size] == 'b' || source[i+size] == 'B') {
				out.WriteString(wordBoundary(source[i+size] == 'B', options))
				i = end
				continue
			}
			// An escape that resolves to a character is written as that
			// character. Passed through as written, one .NET reads differently
			// — `\A`, `\a`, `\e` — would keep its .NET meaning, and one .NET
			// does not know at all would refuse to compile.
			if next != utf8.RuneError {
				out.WriteString(literalRune(next))
				i = end
				continue
			}
			out.WriteString(source[i:end])
			i = end
			continue

		case inClass:
			if r == ']' {
				body := out.String()[classBodyStart:]
				if options.ignoreCase {
					rest := CaseCloseClass(body, options.unicode)
					trimmed := out.String()[:classBodyStart]
					out.Reset()
					out.WriteString(trimmed)
					out.WriteString(rest)
				}
				out.WriteByte(']')
				inClass = false
				i += size
				continue
			}
			out.WriteRune(r)
			i += size
			continue

		case r == '(':
			// `(?<name>` opens a named group. Same as `\k<name>`: the name is
			// not text, so it goes through untouched.
			if opener, openerSize, ok := namedGroupOpener(source[i:]); ok {
				out.WriteString(opener)
				i += openerSize
				continue
			}
			if err := checkGroupConstruct(source[i:]); err != nil {
				return "", false, err
			}
			out.WriteByte('(')
			i += size
			continue

		case r == '[':
			// `[]` never matches and `[^]` matches anything: syntax only
			// JavaScript reads, and regexp2 would take the brackets literally.
			if strings.HasPrefix(source[i:], "[]") {
				out.WriteString(neverMatches)
				i += 2
				continue
			}
			if strings.HasPrefix(source[i:], "[^]") {
				out.WriteString(anyCharacter)
				i += 3
				continue
			}
			out.WriteByte('[')
			inClass = true
			i += size
			if i < len(source) && source[i] == '^' {
				out.WriteByte('^')
				i++
			}
			classBodyStart = out.Len()
			continue

		case r == '.':
			if options.dotAll {
				out.WriteString(anyCharacter)
			} else {
				out.WriteString(nonTerminator)
			}
			i += size
			continue

		case r == '^':
			if options.multiline {
				out.WriteString(lineStart)
			} else {
				out.WriteString(inputStart)
			}
			i += size
			continue

		case r == '$':
			if options.multiline {
				out.WriteString(lineEnd)
			} else {
				out.WriteString(inputEnd)
			}
			i += size
			continue

		default:
			if options.ignoreCase {
				if class, widened := CaseClass(r, options.unicode); widened {
					out.WriteString(class)
					i += size
					continue
				}
			}
			out.WriteRune(r)
			i += size
		}
	}

	return out.String(), exact, nil
}

// decodeEscape reads the body of a backslash escape starting at i, returning
// the character it names and how many bytes it spans. The rune is
// utf8.RuneError for an escape that names a set or a group rather than one
// character, which the caller leaves alone.
func decodeEscape(source string, i int, unicode bool) (rune, int, bool) {
	if i >= len(source) {
		return 0, 0, false
	}
	r, size := utf8.DecodeRuneInString(source[i:])

	switch r {
	case 'u':
		// `\uXXXX`, or `\u{X...}` under the `u` flag.
		if unicode && strings.HasPrefix(source[i+size:], "{") {
			end := strings.IndexByte(source[i+size:], '}')
			if end < 0 {
				return utf8.RuneError, size, true
			}
			value, err := strconv.ParseUint(source[i+size+1:i+size+end], 16, 32)
			if err != nil {
				return utf8.RuneError, size + end + 1, true
			}
			return rune(value), size + end + 1, true
		}
		return decodeFixedHex(source, i, size, 4)
	case 'x':
		return decodeFixedHex(source, i, size, 2)
	case 'n':
		return '\n', size, true
	case 'r':
		return '\r', size, true
	case 't':
		return '\t', size, true
	case 'f':
		return '\f', size, true
	case 'v':
		return '\v', size, true
	case '0':
		return 0, size, true
	case 'd', 'D', 'w', 'W', 's', 'S', 'b', 'B', 'k', 'c':
		// Names a set, a boundary or a group rather than a character.
		return utf8.RuneError, size, true
	case 'p', 'P':
		// A property class, but only under `u`. Without it JavaScript reads
		// `\p` as the letter, where .NET reads a property either way.
		if unicode {
			return utf8.RuneError, size, true
		}
		return r, size, true
	}
	if r >= '1' && r <= '9' {
		// A backreference.
		return utf8.RuneError, size, true
	}
	// Any other escape stands for the character itself.
	return r, size, true
}

func decodeFixedHex(source string, i int, size int, digits int) (rune, int, bool) {
	start := i + size
	if start+digits > len(source) {
		return utf8.RuneError, size, true
	}
	value, err := strconv.ParseUint(source[start:start+digits], 16, 32)
	if err != nil {
		return utf8.RuneError, size, true
	}
	return rune(value), size + digits, true
}

// namedGroupOpener matches `(?<name>` at the start of source, returning it
// whole. A `(?<=` or `(?<!` is a lookbehind rather than a name, and carries
// text that does get widened, so it is left to the ordinary walk.
func namedGroupOpener(source string) (string, int, bool) {
	if !strings.HasPrefix(source, "(?<") {
		return "", 0, false
	}
	rest := source[len("(?<"):]
	if strings.HasPrefix(rest, "=") || strings.HasPrefix(rest, "!") {
		return "", 0, false
	}
	end := strings.IndexByte(rest, '>')
	if end < 0 {
		return "", 0, false
	}
	width := len("(?<") + end + 1
	return source[:width], width, true
}

// namedBackreference matches `\k<name>` at the start of source.
func namedBackreference(source string) (string, int, bool) {
	if !strings.HasPrefix(source, `\k<`) {
		return "", 0, false
	}
	end := strings.IndexByte(source[len(`\k<`):], '>')
	if end < 0 {
		return "", 0, false
	}
	width := len(`\k<`) + end + 1
	return source[:width], width, true
}

// isBackreference reports whether an escape body names an earlier group, which
// is the one thing a widened pattern cannot reproduce under `i`.
func isBackreference(body string) bool {
	if body == "" {
		return false
	}
	return (body[0] >= '1' && body[0] <= '9') || body[0] == 'k'
}

// checkGroupConstruct rejects a `(?…` opener JavaScript has no syntax for.
// regexp2 follows .NET, which reads several more — `(?i)` sets a flag there
// and is a syntax error in JavaScript — so accepting them would quietly give
// a pattern a meaning its author never wrote.
func checkGroupConstruct(source string) error {
	if !strings.HasPrefix(source, "(?") {
		return nil
	}
	rest := source[len("(?"):]
	for _, valid := range []string{":", "=", "!", "<="} {
		if strings.HasPrefix(rest, valid) {
			return nil
		}
	}
	if strings.HasPrefix(rest, "<!") {
		return nil
	}
	return fmt.Errorf("%w: (?%.1s", ErrUnsupportedSyntax, rest)
}
