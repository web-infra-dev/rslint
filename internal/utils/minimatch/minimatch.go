// Package minimatch ports minimatch 3.1.5 and the brace-expansion 1.1.16 it
// pulls in, the glob matcher an ESLint plugin compares paths with. It covers
// the extended glob syntax — `!(a)`, `@(a|b)`, `+(a)`, `?(a)`, `*(a)` — that
// the general-purpose Go glob libraries leave out, so a rule ported from an
// ESLint plugin can accept the same patterns its upstream does.
//
// Plugins are what this is for, so 3.x is the version to answer like. ESLint
// itself moved on to minimatch 10 for the paths it matches on its own behalf,
// which reads a handful of patterns differently.
//
// Patterns compile to regexp2 rather than the standard library's regexp,
// because a negated list needs the lookahead RE2 has no syntax for.
//
// The port covers matching only: brace expansion, extended glob syntax, `**`,
// character classes and negated patterns. Partial matching and the filesystem
// traversal helpers are left out.
//
// A `**` walks the path by recursion, the way minimatch did before it traded
// that for a head/body/tail decomposition and a depth cap. Both answer the
// same, and a pattern of the shape rules write costs microseconds either way,
// but the recursion is exponential in how many `**` one pattern carries: past
// roughly eight of them it stops being cheap.
//
// A `?` and a character class each match one whole character, where minimatch
// counts the UTF-16 units JavaScript strings are made of and so needs `??` for
// a character outside the basic multilingual plane.
//
// A class that names nothing, `[!]` or `[^]`, matches nothing. JavaScript
// alone reads it as any character, and later minimatch dropped that too.
//
// A `/` separates the parts of a path and a `\` escapes the character after
// it, on every platform. minimatch 3 rewrote a `\` to a `/` when it ran on
// Windows, which spends the escape a pattern was written with — a rule option
// naming `src/\*.ts` would pick up every `.ts` file there. Paths reach a rule
// separated by `/` whatever the platform, so there is nothing to rewrite, and
// later minimatch stopped rewriting by default too.
//
// See the LICENSE file in this directory for the upstream copyright notices.
package minimatch

import (
	"strings"

	"github.com/dlclark/regexp2"
)

// Options mirrors the minimatch options a caller can pass. The zero value is
// minimatch's own default behavior.
type Options struct {
	// Dot lets a wildcard match a name starting with a period. Without it,
	// `a/**/b` does not match `a/.d/b`.
	Dot bool
	// NoBrace turns off `{a,b}` and `{1..3}` expansion.
	NoBrace bool
	// NoGlobStar makes `**` behave like a single `*`.
	NoGlobStar bool
	// NoExt turns off extended glob syntax such as `+(a|b)`.
	NoExt bool
	// NoCase matches case-insensitively.
	NoCase bool
	// NoNegate turns off the leading `!` that negates a whole pattern.
	NoNegate bool
	// NoComment turns off treating a leading `#` as starting a comment.
	NoComment bool
	// NoNull is unused by this port and kept for option parity.
	NoNull bool
	// MatchBase matches a pattern without slashes against the basename of a
	// path that has them, so `a?b` matches `/xyz/123/acb`.
	MatchBase bool
	// FlipNegate returns the result of a negated pattern unnegated.
	FlipNegate bool
}

var plTypes = map[byte]struct{ open, close string }{
	'!': {open: "(?:(?!(?:", close: "))[^/]*?)"},
	'?': {open: "(?:", close: ")?"},
	'+': {open: "(?:", close: ")+"},
	'*': {open: "(?:", close: ")*"},
	'@': {open: "(?:", close: ")"},
}

const (
	// qmark matches any single thing other than a path separator.
	qmark = "[^/]"
	// star matches any number of characters within one path part.
	star = qmark + "*?"
	// reSpecials are the characters that need escaping in a regexp.
	reSpecials = `().*{}+?[]^$\!`
	// maxPatternLength bounds a pattern the way minimatch does.
	maxPatternLength = 1024 * 64
)

// braceShortcut detects the brace set that makes expansion worth running. The
// lookahead rules out a nested `{`, which is what keeps the scan linear.
var braceShortcut = regexp2.MustCompile(`\{(?:(?!\{).)*\}`, regexp2.None)

// neverMatches stands in for a pattern part that failed to compile, looking for
// a character past the end of the string.
var neverMatches = regexp2.MustCompile(`$.`, regexp2.None)

// patternPart is one path part of a compiled pattern: a `**`, a literal name,
// or a regexp when the part carries wildcards.
type patternPart struct {
	globstar bool
	literal  string
	re       *regexp2.Regexp
}

// Matcher is a compiled pattern, ready to match any number of paths.
type Matcher struct {
	options Options
	pattern string
	negate  bool
	comment bool
	empty   bool
	// set holds one row per brace expansion, each row one part per path part.
	set [][]patternPart
}

// New compiles pattern. A pattern that cannot be compiled yields a Matcher that
// matches nothing rather than an error, which is how minimatch behaves.
func New(pattern string, options Options) *Matcher {
	pattern = strings.TrimSpace(pattern)

	m := &Matcher{options: options, pattern: pattern}
	m.make()
	return m
}

// Match reports whether path matches pattern. Compile the pattern with New
// instead when the same one is matched against more than one path.
func Match(pattern string, path string, options Options) bool {
	return New(pattern, options).Match(path)
}

func (m *Matcher) make() {
	if !m.options.NoComment && strings.HasPrefix(m.pattern, "#") {
		m.comment = true
		return
	}
	if m.pattern == "" {
		m.empty = true
		return
	}

	// step 1: figure out negation, etc.
	m.parseNegate()

	// step 2: expand braces, then turn each expansion into a series of
	// path-part matching patterns.
	for _, expansion := range m.braceExpand() {
		parts := splitSlashes(expansion)
		row := make([]patternPart, 0, len(parts))
		compiled := true
		for _, part := range parts {
			p, ok := m.parsePart(part)
			if !ok {
				// filter out everything that didn't compile properly.
				compiled = false
				break
			}
			row = append(row, p)
		}
		if compiled {
			m.set = append(m.set, row)
		}
	}
}

func (m *Matcher) parseNegate() {
	if m.options.NoNegate {
		return
	}
	offset := 0
	for offset < len(m.pattern) && m.pattern[offset] == '!' {
		m.negate = !m.negate
		offset++
	}
	m.pattern = m.pattern[offset:]
}

func (m *Matcher) braceExpand() []string {
	if m.options.NoBrace || len(m.pattern) > maxPatternLength {
		return []string{m.pattern}
	}
	if hasBraces, err := braceShortcut.MatchString(m.pattern); err != nil || !hasBraces {
		// shortcut. no need to expand.
		return []string{m.pattern}
	}
	return BraceExpand(m.pattern)
}

// parsePart compiles one path part. Following the lead of Bash 4.1, `**` only
// has special meaning when it is the only thing in a path part; anywhere else
// any series of `*` is equivalent to a single `*`.
func (m *Matcher) parsePart(pattern string) (patternPart, bool) {
	if pattern == "**" {
		if !m.options.NoGlobStar {
			return patternPart{globstar: true}, true
		}
		pattern = "*"
	}
	if pattern == "" {
		return patternPart{}, true
	}

	source, hasMagic, ok := m.parseSource(pattern, false)
	if !ok {
		return patternPart{}, false
	}

	// Skip the regexp for non-magical patterns, unescaping the pattern so it
	// compares exactly against a name.
	if !hasMagic {
		return patternPart{literal: globUnescape(pattern)}, true
	}

	flags := regexp2.None
	if m.options.NoCase {
		flags = regexp2.IgnoreCase
	}
	re, err := regexp2.Compile("^"+source+"$", flags)
	if err != nil {
		// An invalid regular expression can't match anything.
		re = neverMatches
	}
	return patternPart{re: re}, true
}

// patternListItem tracks one open extended glob list while a part is parsed.
type patternListItem struct {
	kind    byte
	reStart int
	reEnd   int
	open    string
	close   string
}

// parseSource translates one path part into regexp source. isSub marks the
// re-walk of a would-be character class, whose result is spliced into the
// caller's source rather than compiled on its own. The third result is false
// when the part cannot be translated at all, which drops the whole expansion.
func (m *Matcher) parseSource(pattern string, isSub bool) (string, bool, bool) {
	if len(pattern) > maxPatternLength {
		return "", false, false
	}
	if pattern == "" {
		return "", m.options.NoCase, true
	}

	re := ""
	hasMagic := m.options.NoCase
	escaping := false
	patternListStack := []patternListItem{}
	negativeLists := []patternListItem{}
	stateChar := byte(0)
	inClass := false
	reClassStart := -1
	classStart := -1

	// `.` and `..` never match anything that doesn't start with a period, even
	// when Dot is set.
	patternStart := ""
	switch {
	case pattern[0] == '.':
		patternStart = ""
	case m.options.Dot:
		patternStart = `(?!(?:^|/)\.{1,2}(?:$|/))`
	default:
		patternStart = `(?!\.)`
	}

	// clearStateChar emits a state-tracking character that turned out not to
	// open an extended glob list.
	clearStateChar := func() {
		if stateChar == 0 {
			return
		}
		switch stateChar {
		case '*':
			re += star
			hasMagic = true
		case '?':
			re += qmark
			hasMagic = true
		default:
			re += `\` + string(stateChar)
		}
		stateChar = 0
	}

	// The index a rune is reported at is its offset in bytes, which is what the
	// character class slicing below wants: every character the parser gives a
	// meaning to is ASCII, so an offset always lands on a rune boundary.
	for i, c := range pattern {
		// skip over any that are escaped.
		if escaping && strings.ContainsRune(reSpecials, c) {
			re += `\` + string(c)
			escaping = false
			continue
		}

		switch c {
		case '/':
			// completely not allowed, even escaped: the pattern is already
			// split on path separators by now.
			return "", false, false

		case '\\':
			clearStateChar()
			escaping = true

		case '?', '*', '+', '@', '!':
			// All of those are literals inside a class, except that the glob
			// `[!a]` means `[^a]` in a regexp.
			if inClass {
				if c == '!' && i == classStart+1 {
					c = '^'
				}
				re += string(c)
				continue
			}
			// coalesce consecutive non-globstar `*` characters
			if c == '*' && stateChar == '*' {
				continue
			}
			// If we already have a state character then there was something
			// like `**` or `+?`: handle that one first, then hold on to this.
			clearStateChar()
			stateChar = byte(c)
			// Without extended glob syntax `+(asdf|foo)` isn't a thing, so
			// release the state character now rather than opening a list.
			if m.options.NoExt {
				clearStateChar()
			}

		case '(':
			if inClass {
				re += "("
				continue
			}
			if stateChar == 0 {
				re += `\(`
				continue
			}
			plType := plTypes[stateChar]
			patternListStack = append(patternListStack, patternListItem{
				kind:    stateChar,
				reStart: len(re),
				open:    plType.open,
				close:   plType.close,
			})
			re += plType.open
			stateChar = 0

		case ')':
			if inClass || len(patternListStack) == 0 {
				re += `\)`
				continue
			}
			clearStateChar()
			hasMagic = true
			pl := patternListStack[len(patternListStack)-1]
			patternListStack = patternListStack[:len(patternListStack)-1]
			re += pl.close
			pl.reEnd = len(re)
			if pl.kind == '!' {
				negativeLists = append(negativeLists, pl)
			}

		case '|':
			if inClass || len(patternListStack) == 0 || escaping {
				re += `\|`
				escaping = false
				continue
			}
			clearStateChar()
			re += "|"

		case '[':
			// swallow any state-tracking char before the `[`
			clearStateChar()
			if inClass {
				re += `\` + string(c)
				continue
			}
			inClass = true
			classStart = i
			reClassStart = len(re)
			re += string(c)

		case ']':
			// A right bracket loses its special meaning and represents itself
			// when it occurs first in a bracket expression. -- POSIX.2 2.8.3.2
			if i == classStart+1 || !inClass {
				re += `\` + string(c)
				escaping = false
				continue
			}
			// Handle the case where we left a class open: `[z-a]` is valid and
			// equivalent to `\[z-a\]`. Split where the last `[` was and re-walk
			// the contents so any character that was passed through as-is gets
			// translated.
			//
			// What decides that is whether the class compiles, so ask about the
			// translated source rather than the pattern text it came from: a
			// character such as `[` or `\` that a glob passes straight through
			// carries a meaning of its own in a regexp character class.
			if _, err := regexp2.Compile(re[reClassStart:]+"]", regexp2.None); err != nil {
				class := pattern[classStart+1 : i]
				source, sourceMagic, _ := m.parseSource(class, true)
				re = re[:reClassStart] + `\[` + source + `\]`
				hasMagic = hasMagic || sourceMagic
				inClass = false
				continue
			}
			// finish up the class.
			hasMagic = true
			inClass = false
			re += string(c)

		default:
			// swallow any state character that wasn't consumed
			clearStateChar()
			if escaping {
				escaping = false
			} else if strings.ContainsRune(reSpecials, c) && (c != '^' || !inClass) {
				re += `\`
			}
			re += string(c)
		}
	}

	// Handle the case where we left a class open: `[abc` is valid and
	// equivalent to `\[abc`.
	if inClass {
		class := pattern[classStart+1:]
		source, sourceMagic, _ := m.parseSource(class, true)
		re = re[:reClassStart] + `\[` + source
		hasMagic = hasMagic || sourceMagic
	}

	// Handle the case where we had a `+(` thing at the very end of the part,
	// whose list never closed.
	for len(patternListStack) > 0 {
		pl := patternListStack[len(patternListStack)-1]
		patternListStack = patternListStack[:len(patternListStack)-1]

		tail := escapeTailPipes(re[pl.reStart+len(pl.open):])
		var opener string
		switch pl.kind {
		case '*':
			opener = star
		case '?':
			opener = qmark
		default:
			opener = `\` + string(pl.kind)
		}

		hasMagic = true
		re = re[:pl.reStart] + opener + `\(` + tail
	}

	// handle trailing things that only matter at the very end.
	clearStateChar()
	if escaping {
		// trailing backslash
		re += `\\`
	}

	// Only apply the no-dot start if the source starts with something that
	// could conceivably capture a period.
	addPatternStart := false
	if re != "" {
		switch re[0] {
		case '.', '[', '(':
			addPatternStart = true
		}
	}

	// Work around the lack of negative lookbehind: a pattern like
	// `*.!(x).!(y|z)` has to keep a name like `a.xyz.yz` from matching, so the
	// first negative lookahead has to look all the way to the end.
	for n := len(negativeLists) - 1; n >= 0; n-- {
		nl := negativeLists[n]

		before := jsSlice(re, 0, nl.reStart)
		first := jsSlice(re, nl.reStart, nl.reEnd-8)
		last := jsSlice(re, nl.reEnd-8, nl.reEnd)
		after := jsSlice(re, nl.reEnd, len(re))

		last += after

		// Handle nested lists such as `*(*.js|!(*.json))`, where an open paren
		// means the `)` does not belong to the part considered "after" the
		// negated section.
		for openParens := strings.Count(before, "("); openParens > 0; openParens-- {
			after = dropFirstCloseParen(after)
		}

		dollar := ""
		if after == "" && !isSub {
			dollar = "$"
		}
		re = before + first + after + dollar + last
	}

	// A non-empty source has to be kept from matching an empty path part, so
	// that `a/*` does not match `a/`.
	if re != "" && hasMagic {
		re = "(?=.)" + re
	}

	if addPatternStart {
		re = patternStart + re
	}

	return re, hasMagic, true
}

// Match reports whether path matches the compiled pattern.
func (m *Matcher) Match(path string) bool {
	if m.comment {
		return false
	}
	if m.empty {
		return path == ""
	}

	parts := splitSlashes(path)

	// the basename is the last non-empty part
	basename := ""
	for i := len(parts) - 1; i >= 0; i-- {
		if parts[i] != "" {
			basename = parts[i]
			break
		}
	}

	// Just one of the pattern rows needs to match. If negating, one match
	// means we have failed. Either way, return on the first hit.
	for _, row := range m.set {
		file := parts
		if m.options.MatchBase && len(row) == 1 {
			file = []string{basename}
		}
		if m.matchOne(file, row) {
			if m.options.FlipNegate {
				return true
			}
			return !m.negate
		}
	}

	// No hits. That is success for a negated pattern, failure otherwise.
	if m.options.FlipNegate {
		return false
	}
	return m.negate
}

func (m *Matcher) matchOne(file []string, row []patternPart) bool {
	fi, pi := 0, 0
	fl, pl := len(file), len(row)

	for ; fi < fl && pi < pl; fi, pi = fi+1, pi+1 {
		part := row[pi]

		if part.globstar {
			// `a/**/b/**/c` matches `a/b/x/y/z/c`, `a/x/y/z/b/c`, `a/b/c` and
			// so on: take the rest of the pattern after the `**` and see
			// whether it matches the remainder of the file. If not, the `**`
			// swallows one part and we try again.
			if pi+1 == pl {
				// A trailing `**` swallows the rest, except that it never
				// swallows `.`, `..` or, without Dot, a dot-name.
				for _, remaining := range file[fi:] {
					if m.isDotPart(remaining) {
						return false
					}
				}
				return true
			}
			for fr := fi; fr < fl; fr++ {
				if m.matchOne(file[fr:], row[pi+1:]) {
					return true
				}
				// `.` and `..` are never swallowed, and a dot-name only when
				// explicitly asked for.
				if m.isDotPart(file[fr]) {
					break
				}
			}
			return false
		}

		if !part.match(file[fi]) {
			return false
		}
	}

	switch {
	case fi == fl && pi == pl:
		// ran out of pattern and file at the same time: an exact hit.
		return true
	case fi == fl:
		// ran out of file with pattern left over.
		return false
	default:
		// Ran out of pattern with file left over. That is only acceptable on
		// the very last empty part of a path written with a trailing slash, so
		// that `a/*` still matches `a/b/`.
		return fi == fl-1 && file[fi] == ""
	}
}

func (p patternPart) match(name string) bool {
	if p.re == nil {
		// A part without wildcards has to match exactly.
		return name == p.literal
	}
	matched, err := p.re.MatchString(name)
	return err == nil && matched
}

func (m *Matcher) isDotPart(part string) bool {
	return part == "." || part == ".." || (!m.options.Dot && strings.HasPrefix(part, "."))
}

// splitSlashes splits a path on runs of `/`, the way minimatch's `/\/+/` does.
func splitSlashes(path string) []string {
	parts := []string{}
	start := 0
	for i := 0; i <= len(path); i++ {
		if i < len(path) && path[i] != '/' {
			continue
		}
		parts = append(parts, path[start:i])
		for i+1 < len(path) && path[i+1] == '/' {
			i++
		}
		start = i + 1
	}
	return parts
}

// globUnescape drops the backslashes a pattern used to escape a character, so
// a part without wildcards compares against a name as written.
func globUnescape(pattern string) string {
	var unescaped strings.Builder
	for i := 0; i < len(pattern); i++ {
		if pattern[i] == '\\' && i+1 < len(pattern) {
			i++
		}
		unescaped.WriteByte(pattern[i])
	}
	return unescaped.String()
}

// escapeTailPipes escapes every `|` that an unclosed extended glob list left
// behind, doubling the backslashes in front of it. Doubling an even run of
// backslashes is what keeps the one backslash that escapes the `|` from being
// read as escaping a backslash.
func escapeTailPipes(tail string) string {
	var escaped strings.Builder
	for i := 0; i < len(tail); {
		start := i
		for i < len(tail) && tail[i] == '\\' {
			i++
		}
		if i < len(tail) && tail[i] == '|' {
			backslashes := i - start
			escaped.WriteString(strings.Repeat(`\`, (backslashes-backslashes%2)*2))
			escaped.WriteString(`\|`)
			i++
			continue
		}
		escaped.WriteString(tail[start:i])
		if i < len(tail) {
			escaped.WriteByte(tail[i])
			i++
		}
	}
	return escaped.String()
}

// jsSlice takes a substring the way JavaScript's String.prototype.slice does,
// clamping an offset that runs past the end instead of failing. Rewriting an
// unclosed extended glob list shortens the source after a negated list already
// recorded where it ended, so those offsets can point past the end.
func jsSlice(source string, start int, end int) string {
	start = min(max(start, 0), len(source))
	end = min(max(end, start), len(source))
	return source[start:end]
}

// dropFirstCloseParen removes the first `)` and any quantifier bound to it.
func dropFirstCloseParen(source string) string {
	index := strings.IndexByte(source, ')')
	if index < 0 {
		return source
	}
	end := index + 1
	if end < len(source) && strings.IndexByte("+*?", source[end]) >= 0 {
		end++
	}
	return source[:index] + source[end:]
}
