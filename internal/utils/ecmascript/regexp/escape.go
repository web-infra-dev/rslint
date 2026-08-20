package regexp

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// escapeKind says what a backslash escape names, which is what decides how the
// rewrite may use it: a character may be widened under `i` and may bound a
// character range, a set may be neither, and an assertion is not a character at
// all.
type escapeKind uint8

const (
	// escapeRune names one character.
	escapeRune escapeKind = iota
	// escapeSet names a set of characters: `\d`, `\w`, `\s` and their negations,
	// and a `\p{…}` property escape.
	escapeSet
	// escapeAssertion names a position rather than a character: `\b` and `\B`,
	// and only outside a character class.
	escapeAssertion
	// escapeBackreference names an earlier group: `\1`, `\k<name>`.
	escapeBackreference
)

// setKind tells apart the sets whose ECMAScript membership the rewrite has to
// spell out from the ones regexp2 already reads the same way.
type setKind uint8

const (
	setOther setKind = iota
	setWord
	setNonWord
	setProperty
)

// decodedEscape is one backslash escape read the way JavaScript reads it.
//
// width counts the bytes the escape spans after the backslash, which is not
// always the whole of what follows: a `\c` that no control letter completes is
// no escape at all under Annex B, and comes back as a backslash of width none,
// leaving the `c` to be read as the character it is.
type decodedEscape struct {
	kind    escapeKind
	set     setKind
	r       rune
	width   int
	negated bool
}

// escapeContext is what an escape's meaning depends on besides its own text.
// A `\b` is a boundary out in the pattern and a backspace inside a character
// class; a `\1` is a backreference where a first group exists and an octal
// escape where none does; and `u` decides both which syntax is legal and, for
// several escapes, what the legal syntax means.
type escapeContext struct {
	unicode bool
	inClass bool
	groups  int
	named   bool
}

// syntaxCharacters is what `u` narrows an identity escape down to, alongside
// `/`. Without `u`, Annex B lets a backslash stand before nearly anything.
const syntaxCharacters = `^$\.*+?()[]{}|`

// decodeEscape reads the escape whose backslash ends at i, returning what it
// names. An escape JavaScript itself rejects comes back as an error rather than
// as text for regexp2 to judge, which reads a different syntax.
func decodeEscape(source string, i int, ctx escapeContext) (decodedEscape, error) {
	if i >= len(source) {
		return decodedEscape{}, fmt.Errorf(`%w: \ at the end of the pattern`, ErrUnsupportedSyntax)
	}
	r, size := utf8.DecodeRuneInString(source[i:])

	switch r {
	case 'u':
		return decodeUnicodeEscape(source, i, size, ctx)
	case 'x':
		return decodeFixedHex(source, i, size, 2, ctx)
	case 'n':
		return decodedEscape{r: '\n', width: size}, nil
	case 'r':
		return decodedEscape{r: '\r', width: size}, nil
	case 't':
		return decodedEscape{r: '\t', width: size}, nil
	case 'f':
		return decodedEscape{r: '\f', width: size}, nil
	case 'v':
		return decodedEscape{r: '\v', width: size}, nil
	case 'c':
		return decodeControlEscape(source, i, size, ctx)
	case 'd', 'D', 's', 'S':
		return decodedEscape{kind: escapeSet, width: size}, nil
	case 'w':
		return decodedEscape{kind: escapeSet, set: setWord, width: size}, nil
	case 'W':
		return decodedEscape{kind: escapeSet, set: setNonWord, width: size}, nil
	case 'p', 'P':
		return decodePropertyEscape(source, i, size, ctx)
	case 'b':
		if ctx.inClass {
			// Inside a class a `\b` is a backspace, which is the one place the
			// two readings of the escape part ways.
			return decodedEscape{r: '\b', width: size}, nil
		}
		return decodedEscape{kind: escapeAssertion, width: size}, nil
	case 'B':
		if ctx.inClass {
			return identityEscape(r, size, ctx)
		}
		return decodedEscape{kind: escapeAssertion, width: size, negated: true}, nil
	case 'k':
		// `\k<name>` is read whole before this. What is left is a `\k` naming
		// nothing, which Annex B reads as the letter — but only in a pattern
		// that opens no named group, since one that does makes the reference
		// mandatory.
		if ctx.named {
			return decodedEscape{}, fmt.Errorf(`%w: \k naming no group`, ErrUnsupportedSyntax)
		}
		return identityEscape(r, size, ctx)
	}

	if r >= '0' && r <= '9' {
		return decodeNumericEscape(source, i, ctx)
	}
	return identityEscape(r, size, ctx)
}

// decodeUnicodeEscape reads `\uXXXX`, or `\u{X…}` where the `u` flag spells it.
func decodeUnicodeEscape(source string, i int, size int, ctx escapeContext) (decodedEscape, error) {
	if ctx.unicode && strings.HasPrefix(source[i+size:], "{") {
		end := strings.IndexByte(source[i+size:], '}')
		if end < 0 {
			return decodedEscape{}, fmt.Errorf(`%w: \u{ that no } closes`, ErrUnsupportedSyntax)
		}
		value, err := strconv.ParseUint(source[i+size+1:i+size+end], 16, 32)
		if err != nil || value > unicode.MaxRune {
			return decodedEscape{}, fmt.Errorf(`%w: \u{…} naming no character`, ErrUnsupportedSyntax)
		}
		return decodedEscape{r: rune(value), width: size + end + 1}, nil
	}
	return decodeFixedHex(source, i, size, 4, ctx)
}

// decodeFixedHex reads the fixed-width hexadecimal escapes, `\xXX` and
// `\uXXXX`. Without `u`, one that no hex digits complete is not an error but an
// identity escape, so `/\x/` matches an `x` — and under `i` an `X` as well.
func decodeFixedHex(source string, i int, size int, digits int, ctx escapeContext) (decodedEscape, error) {
	start := i + size
	if start+digits <= len(source) {
		if value, err := strconv.ParseUint(source[start:start+digits], 16, 32); err == nil {
			return decodedEscape{r: rune(value), width: size + digits}, nil
		}
	}
	if ctx.unicode {
		return decodedEscape{}, fmt.Errorf(`%w: \%.1s naming no character`, ErrUnsupportedSyntax, source[i:])
	}
	r, _ := utf8.DecodeRuneInString(source[i:])
	return decodedEscape{r: r, width: size}, nil
}

// decodeControlEscape reads `\cX`, the control character an ASCII letter names.
func decodeControlEscape(source string, i int, size int, ctx escapeContext) (decodedEscape, error) {
	if letter := i + size; letter < len(source) {
		c := source[letter]
		// Annex B widens the letter a character class takes to a digit or an
		// underscore, which is the one place the two productions differ.
		if c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' ||
			!ctx.unicode && ctx.inClass && (c >= '0' && c <= '9' || c == '_') {
			return decodedEscape{r: rune(c % 32), width: size + 1}, nil
		}
	}
	if ctx.unicode {
		return decodedEscape{}, fmt.Errorf(`%w: \c naming no control character`, ErrUnsupportedSyntax)
	}
	// Annex B: with no control letter after it the backslash is not an escape
	// at all and stands for itself, leaving the `c` to be read on its own.
	return decodedEscape{r: '\\', width: 0}, nil
}

// decodePropertyEscape reads `\p{…}` and `\P{…}`, which name a set out of
// Unicode's tables. Only the `u` flag spells them; without it .NET would still
// read a property where JavaScript reads the letter.
func decodePropertyEscape(source string, i int, size int, ctx escapeContext) (decodedEscape, error) {
	if !ctx.unicode {
		return identityEscape(rune(source[i]), size, ctx)
	}
	if !strings.HasPrefix(source[i+size:], "{") {
		return decodedEscape{}, fmt.Errorf(`%w: \%.1s naming no property`, ErrUnsupportedSyntax, source[i:])
	}
	end := strings.IndexByte(source[i+size:], '}')
	if end < 0 {
		return decodedEscape{}, fmt.Errorf(`%w: \%.1s{ that no } closes`, ErrUnsupportedSyntax, source[i:])
	}
	// The name belongs to the escape: `\p{Script=Greek}`. Taking it along keeps
	// the walk from reading the letters in it as text and widening them.
	return decodedEscape{kind: escapeSet, set: setProperty, width: size + end + 1}, nil
}

// decodeNumericEscape reads an escape a digit opens, which is a backreference,
// a NUL, or one of Annex B's legacy octal escapes depending on the digits, on
// where it sits, and on how many groups the pattern opens.
func decodeNumericEscape(source string, i int, ctx escapeContext) (decodedEscape, error) {
	if source[i] == '0' && !isDecimalDigit(source, i+1) {
		return decodedEscape{r: 0, width: 1}, nil
	}
	// A number no group answers to is not a backreference at all, and Annex B
	// reads the same text as an escaped character instead. Inside a class there
	// is no backreference to read in the first place.
	if source[i] != '0' && !ctx.inClass {
		if value, width := decimalEscape(source, i); value <= ctx.groups {
			return decodedEscape{kind: escapeBackreference, width: width}, nil
		}
	}
	if ctx.unicode {
		return decodedEscape{}, fmt.Errorf(`%w: \%.1s`, ErrUnsupportedSyntax, source[i:])
	}
	if r, width := decodeLegacyOctal(source, i); width > 0 {
		return decodedEscape{r: r, width: width}, nil
	}
	// `\8` and `\9`, which no octal escape spells.
	return decodedEscape{r: rune(source[i]), width: 1}, nil
}

// decimalEscape reads the whole run of digits a backreference is numbered with,
// counting past what any pattern could hold rather than overflowing.
func decimalEscape(source string, i int) (int, int) {
	const tooMany = 1 << 20
	value, width := 0, 0
	for isDecimalDigit(source, i+width) {
		if value < tooMany {
			value = value*10 + int(source[i+width]-'0')
		}
		width++
	}
	return value, width
}

// decodeLegacyOctal reads the octal escape Annex B keeps for a pattern without
// `u`, which spans up to three digits and never overruns one byte.
func decodeLegacyOctal(source string, i int) (rune, int) {
	if source[i] < '0' || source[i] > '7' {
		return 0, 0
	}
	digits := 3
	if source[i] > '3' {
		digits = 2
	}
	value, width := rune(source[i]-'0'), 1
	for width < digits && i+width < len(source) && source[i+width] >= '0' && source[i+width] <= '7' {
		value = value*8 + rune(source[i+width]-'0')
		width++
	}
	return value, width
}

// identityEscape reads a backslash standing before a character that carries no
// escape of its own. Annex B lets it stand before nearly anything; `u` narrows
// it to the characters a pattern would otherwise read as syntax.
func identityEscape(r rune, size int, ctx escapeContext) (decodedEscape, error) {
	if ctx.unicode && !strings.ContainsRune(syntaxCharacters, r) && r != '/' && (!ctx.inClass || r != '-') {
		return decodedEscape{}, fmt.Errorf(`%w: \%c`, ErrUnsupportedSyntax, r)
	}
	return decodedEscape{r: r, width: size}, nil
}

func isDecimalDigit(source string, i int) bool {
	return i < len(source) && source[i] >= '0' && source[i] <= '9'
}

// countGroups reads how many capturing groups a pattern opens and whether any
// of them is named, which is what a `\1` and a `\k` are read against.
func countGroups(source string) (int, bool) {
	count, named := 0, false
	inClass := false
	for i := 0; i < len(source); {
		switch source[i] {
		case '\\':
			_, size := utf8.DecodeRuneInString(source[i+1:])
			i += 1 + size
			continue
		case '[':
			inClass = true
		case ']':
			inClass = false
		case '(':
			if inClass {
				break
			}
			rest := source[i+1:]
			switch {
			case !strings.HasPrefix(rest, "?"):
				count++
			case strings.HasPrefix(rest, "?<") && !strings.HasPrefix(rest, "?<=") && !strings.HasPrefix(rest, "?<!"):
				count++
				named = true
			}
		}
		_, size := utf8.DecodeRuneInString(source[i:])
		i += size
	}
	return count, named
}
