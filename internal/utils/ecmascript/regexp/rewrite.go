package regexp

import (
	"fmt"
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
// for a Unicode word set, which is why both have to be spelled out.
func wordCharacters(options rewriteOptions) string {
	return writeClass(wordClassAtoms(options), false, rewriteOptions{})
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

// groupKind tells the groups whose closing parenthesis a quantifier may follow
// apart from the ones it may not.
type groupKind uint8

const (
	groupPlain groupKind = iota
	groupLookahead
	groupLookbehind
)

// openGroup is one `(` the walk has not yet seen the `)` of. options is what to
// restore as it closes; only a modifier group makes an entry differ from the
// one below.
type openGroup struct {
	options rewriteOptions
	kind    groupKind
}

// rewrite translates a JavaScript regexp source into one regexp2 reads the
// same way. It walks the source rather than running a regexp over it, because
// what a character means depends on whether it is escaped and whether it sits
// inside a character class.
//
// The second result is false when the source uses a backreference or a
// property escape under `i`. JavaScript compares a backreference by the same
// canonicalization it compares a literal by, and it draws `\p{…}` from
// Unicode's own tables; a widened pattern can name neither, so the caller falls
// back to regexp2's own case-insensitivity there, which is close but not exact.
func rewrite(source string, options rewriteOptions) (string, bool, error) {
	var out strings.Builder
	out.Grow(len(source))

	exact := true
	// current is what the flags say here, which a modifier group changes for
	// the span of one group.
	current := options
	groups := []openGroup{}
	// How a `\1` and a `\k` read is settled by the pattern as a whole, so the
	// groups are counted before the walk rather than as it goes.
	groupCount, named := countGroups(source)
	context := func() escapeContext {
		return escapeContext{unicode: current.unicode, groups: groupCount, named: named}
	}

	for i := 0; i < len(source); {
		r, size := utf8.DecodeRuneInString(source[i:])

		switch r {
		case '\\':
			// `\k<name>` names a group. The name is not text to be matched,
			// so widening the letters in it would rewrite the reference.
			if name, nameSize, ok := namedBackreference(source[i:]); ok {
				if current.ignoreCase {
					exact = false
				}
				out.WriteString(name)
				i += nameSize
				continue
			}
			escape, err := decodeEscape(source, i+size, context())
			if err != nil {
				return "", false, err
			}
			end := i + size + escape.width

			switch escape.kind {
			case escapeRune:
				// An escape that resolves to a character is written as that
				// character. Passed through as written, one .NET reads
				// differently — `\A`, `\a`, `\e` — would keep its .NET meaning,
				// and one .NET does not know at all would refuse to compile.
				if current.ignoreCase {
					if class, widened := CaseClass(escape.r, current.unicode); widened {
						out.WriteString(class)
						i = end
						continue
					}
				}
				out.WriteString(literalRune(escape.r))

			case escapeBackreference:
				if current.ignoreCase {
					exact = false
				}
				out.WriteString(source[i:end])

			case escapeAssertion:
				// A pattern cannot repeat a position, and lowering the
				// assertion first would leave regexp2 reading syntax
				// JavaScript rejects as something it accepts.
				if quantifierWidth(source[end:]) > 0 {
					return "", false, errNothingToRepeat(source[end:])
				}
				out.WriteString(wordBoundary(escape.negated, current))

			case escapeSet:
				switch {
				// `\d` and `\s` regexp2 already reads as ECMAScript does, and
				// `\w` too until `u` and `i` together widen the set past ASCII.
				case escape.set == setWord && current.ignoreCase && current.unicode:
					out.WriteString(writeClass(wordClassAtoms(current), false, rewriteOptions{}))
				case escape.set == setNonWord && current.ignoreCase && current.unicode:
					out.WriteString(writeClass(wordClassAtoms(current), true, rewriteOptions{}))
				default:
					// A property escape names a set out of Unicode's tables,
					// which a widened pattern has no way to name back.
					if escape.set == setProperty && current.ignoreCase {
						exact = false
					}
					out.WriteString(source[i:end])
				}
			}
			i = end
			continue

		case '[':
			body, negated, width, err := readClass(source[i:])
			if err != nil {
				return "", false, err
			}
			atoms, classExact, err := classAtoms(body, current, context())
			if err != nil {
				return "", false, err
			}
			if !classExact {
				exact = false
			}
			out.WriteString(writeClass(atoms, negated, current))
			i += width
			continue

		case '(':
			// `(?<name>` opens a named group. Same as `\k<name>`: the name is
			// not text, so it goes through untouched.
			if opener, openerSize, ok := namedGroupOpener(source[i:]); ok {
				groups = append(groups, openGroup{options: current})
				out.WriteString(opener)
				i += openerSize
				continue
			}
			// `(?i-m:…)` turns a flag on or off over one group. The rewrite
			// answers to the flags as they stand here, so the group opens as a
			// plain one and the walk goes on under what it says.
			if inside, openerSize, ok := modifierGroup(source[i:], current); ok {
				groups = append(groups, openGroup{options: current})
				current = inside
				if options.ignoreCase && !inside.ignoreCase {
					// A pattern that falls back hands ignoring case to regexp2
					// for the whole of itself, and this group is the one place
					// it must not reach.
					out.WriteString(`(?-i:`)
				} else {
					out.WriteString(`(?:`)
				}
				i += openerSize
				continue
			}
			if err := checkGroupConstruct(source[i:]); err != nil {
				return "", false, err
			}
			groups = append(groups, openGroup{options: current, kind: groupKindOf(source[i:])})
			out.WriteByte('(')
			i += size
			continue

		case ')':
			if last := len(groups) - 1; last >= 0 {
				group := groups[last]
				groups = groups[:last]
				current = group.options
				if err := checkGroupQuantifier(group.kind, source[i+size:], options); err != nil {
					return "", false, err
				}
			}
			out.WriteByte(')')
			i += size
			continue

		case '.':
			if current.dotAll {
				out.WriteString(anyCharacter)
			} else {
				out.WriteString(nonTerminator)
			}
			i += size
			continue

		case '^':
			if quantifierWidth(source[i+size:]) > 0 {
				return "", false, errNothingToRepeat(source[i+size:])
			}
			if current.multiline {
				out.WriteString(lineStart)
			} else {
				out.WriteString(inputStart)
			}
			i += size
			continue

		case '$':
			if quantifierWidth(source[i+size:]) > 0 {
				return "", false, errNothingToRepeat(source[i+size:])
			}
			if current.multiline {
				out.WriteString(lineEnd)
			} else {
				out.WriteString(inputEnd)
			}
			i += size
			continue

		case ']', '{', '}':
			// Annex B lets these stand for themselves when no `u` flag is
			// present. Unicode mode instead treats an unmatched `]` or `}` and
			// a `{` that does not open a valid quantifier as syntax errors.
			if current.unicode {
				if r == '{' {
					if width := boundedQuantifierWidth(source[i:]); width > 0 {
						out.WriteString(source[i : i+width])
						i += width
						continue
					}
				}
				return "", false, fmt.Errorf("%w: lone %c", ErrUnsupportedSyntax, r)
			}
			out.WriteRune(r)
			i += size
			continue

		default:
			if current.ignoreCase {
				if class, widened := CaseClass(r, current.unicode); widened {
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

func errNothingToRepeat(quantifier string) error {
	return fmt.Errorf("%w: %s repeats an assertion", ErrUnsupportedSyntax, quantifier[:quantifierWidth(quantifier)])
}

// quantifierWidth reports how many bytes of source a quantifier spans, or none
// where no quantifier opens it. A `{` that no bound closes is a `{` standing
// for itself, which Annex B allows and which repeats nothing.
func quantifierWidth(source string) int {
	width := 0
	switch {
	case source == "":
		return 0
	case source[0] == '*', source[0] == '+', source[0] == '?':
		width = 1
	case source[0] == '{':
		if width = boundedQuantifierWidth(source); width == 0 {
			return 0
		}
	default:
		return 0
	}
	// A lazy quantifier is still a quantifier.
	if width < len(source) && source[width] == '?' {
		width++
	}
	return width
}

// boundedQuantifierWidth reads `{m}`, `{m,}` or `{m,n}`.
func boundedQuantifierWidth(source string) int {
	i := len("{")
	digits := i
	for isDecimalDigit(source, i) {
		i++
	}
	if i == digits {
		return 0
	}
	if i < len(source) && source[i] == ',' {
		i++
		for isDecimalDigit(source, i) {
			i++
		}
	}
	if i < len(source) && source[i] == '}' {
		return i + 1
	}
	return 0
}

// groupKindOf reads whether a group is a lookaround, which is what decides
// whether a quantifier may follow its close.
func groupKindOf(source string) groupKind {
	switch {
	case strings.HasPrefix(source, "(?="), strings.HasPrefix(source, "(?!"):
		return groupLookahead
	case strings.HasPrefix(source, "(?<="), strings.HasPrefix(source, "(?<!"):
		return groupLookbehind
	}
	return groupPlain
}

// checkGroupQuantifier rejects a quantifier repeating a group that matches no
// characters. Annex B keeps one exception: without `u`, a lookahead may be
// quantified, and matches its body once or not at all. A lookbehind never may.
func checkGroupQuantifier(kind groupKind, rest string, options rewriteOptions) error {
	if kind == groupPlain || quantifierWidth(rest) == 0 {
		return nil
	}
	if kind == groupLookahead && !options.unicode {
		return nil
	}
	return errNothingToRepeat(rest)
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

// modifierGroup reads a `(?flags:` or `(?flags-flags:` opener — the syntax
// JavaScript spells turning `i`, `m` or `s` on or off over one group with — and
// returns the flags that hold inside it. A modifier may be named once across
// both sides, and naming none at all is `(?:`, an ordinary group.
func modifierGroup(source string, current rewriteOptions) (rewriteOptions, int, bool) {
	if !strings.HasPrefix(source, "(?") {
		return current, 0, false
	}
	rest := source[len("(?"):]
	end := strings.IndexByte(rest, ':')
	if end <= 0 {
		return current, 0, false
	}

	inside := current
	named := map[byte]bool{}
	enable := true
	for i := range end {
		flag := rest[i]
		if flag == '-' {
			if !enable {
				return current, 0, false
			}
			enable = false
			continue
		}
		if named[flag] {
			return current, 0, false
		}
		named[flag] = true
		switch flag {
		case 'i':
			inside.ignoreCase = enable
		case 'm':
			inside.multiline = enable
		case 's':
			inside.dotAll = enable
		default:
			return current, 0, false
		}
	}
	if len(named) == 0 {
		return current, 0, false
	}
	return inside, len("(?") + end + 1, true
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
