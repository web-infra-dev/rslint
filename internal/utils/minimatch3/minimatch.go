// Package minimatch3 ports minimatch 3.1.5 and the brace-expansion 1.1.16 it
// pulls in, the glob matcher an ESLint plugin compares paths with. It covers
// the extended glob syntax — `!(a)`, `@(a|b)`, `+(a)`, `?(a)`, `*(a)` — that
// the general-purpose Go glob libraries leave out, so a rule ported from an
// ESLint plugin can accept the same patterns its upstream does.
//
// The major version is in the name because a single ESLint run reads globs by
// two of them. A plugin declares its own dependency and the ecosystem sits on
// 3.x — eslint-plugin-import and eslint-plugin-react both pin `^3.1.2` — while
// ESLint matches its own flat config `files` and `ignores` with minimatch 10.
// The two disagree on more than the version bump suggests, so a port that did
// not say which one it answers like would be read as either.
//
// Patterns compile to regexp2 rather than the standard library's regexp,
// because a negated list needs the lookahead RE2 has no syntax for.
//
// The port covers matching only: brace expansion, extended glob syntax, `**`,
// character classes, negated patterns, and the partial-prefix mode exposed by
// matcher options. Filesystem traversal and list-expansion helpers are left
// out.
//
// A `**` walks the path by recursion, remembering the pairs of path part and
// pattern part it has already failed at so that a pattern carrying several of
// them stays cheap. minimatch instead caps how deep it recurses, and answers
// that a path does not match once it hits the cap.
//
// A `?` and a character class each match one whole character, where minimatch
// counts the UTF-16 units JavaScript strings are made of and so needs `??` for
// a character outside the basic multilingual plane.
//
// A `/` separates the parts of a path and a `\` escapes the character after
// it, on every platform. minimatch 3 rewrote a `\` to a `/` when it ran on
// Windows, which spends the escape a pattern was written with — a rule option
// naming `src/\*.ts` would pick up every `.ts` file there. Paths reach a rule
// separated by `/` whatever the platform, so there is nothing to rewrite, and
// later minimatch stopped rewriting by default too.
//
// See the LICENSE file in this directory for the upstream copyright notices.
package minimatch3

import (
	"strings"
	"time"

	"github.com/dlclark/regexp2"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
	esregexp "github.com/web-infra-dev/rslint/internal/utils/ecmascript/regexp"
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
	// Partial accepts a path that matches the beginning of a pattern even when
	// the path runs out first. minimatch uses this while walking a filesystem;
	// eslint-plugin-import also exposes it through pathGroups.patternOptions.
	Partial bool
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
	// anyChar matches any single character, the way JavaScript's `[^]` does.
	anyChar = `[\s\S]`
	// nonTerminator matches any single character a JavaScript `.` does, which
	// is every character but a line terminator. regexp2 reads a `.` the way
	// .NET does, where only a `\n` is left out.
	nonTerminator = `[^\n\r\u2028\u2029]`
	// reSpecials are the characters that need escaping in a regexp.
	reSpecials = `().*{}+?[]^$\!`
	// maxPatternLength bounds a pattern the way minimatch does, in the UTF-16
	// code units JavaScript measures a string in.
	maxPatternLength = 1024 * 64
)

// matchTimeout bounds how long one name may spend in one pattern part. The
// extended glob syntax compiles to the overlapping alternations a backtracking
// engine explores exponentially — `+(a|aa)` against a run of `a` that ends in
// anything else is the shape of it — and regexp2 backtracks. A name that
// matches settles within microseconds, so a match still running after this has
// found nothing and is not about to.
const matchTimeout = time.Second

// braceShortcut detects the brace set that makes expansion worth running. The
// lookahead rules out a nested `{`, which is what keeps the scan linear.
var braceShortcut = boundMatching(regexp2.MustCompile(`\{(?:(?!\{)`+nonTerminator+`)*\}`, regexp2.None))

// neverMatches stands in for a pattern part that failed to compile, looking for
// a character past the end of the string.
var neverMatches = regexp2.MustCompile(`\z.`, regexp2.None)

// boundMatching holds a compiled pattern to matchTimeout.
func boundMatching(re *regexp2.Regexp) *regexp2.Regexp {
	re.MatchTimeout = matchTimeout
	return re
}

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
	invalid bool
	// set holds one row per brace expansion, each row one part per path part.
	set [][]patternPart
}

// New compiles pattern. A pattern that cannot be compiled yields a Matcher that
// matches nothing rather than an error, which is how minimatch behaves.
func New(pattern string, options Options) *Matcher {
	m := &Matcher{options: options, pattern: pattern}

	// minimatch measures the pattern it was handed before it reads anything
	// else about it, and refuses to compile one this long. Nothing about the
	// pattern is honored past that point, a leading `!` included, so a matcher
	// that says no to every path stands in for the refusal.
	if overMaxPatternLength(pattern) {
		m.invalid = true
		return m
	}

	m.pattern = ecmascript.StringTrim(pattern)
	m.make()
	return m
}

// overMaxPatternLength reports whether a pattern runs past the length
// minimatch refuses to compile one at, measuring it in the UTF-16 code units
// String.prototype.length counts rather than in bytes. A character below
// U+10000 is one code unit and the rest are two, so the count never runs above
// the length in bytes: a pattern that fits the limit in bytes fits it in code
// units too, which is what keeps the scan off the ordinary pattern.
func overMaxPatternLength(pattern string) bool {
	if len(pattern) <= maxPatternLength {
		return false
	}
	units := 0
	for _, r := range pattern {
		units++
		if r > 0xFFFF {
			units++
		}
		if units > maxPatternLength {
			return true
		}
	}
	return false
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
	if m.options.NoBrace {
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

	re, err := regexp2.Compile("^"+endAnchors(source)+`\z`, regexp2.None)
	if err != nil {
		// An invalid regular expression can't match anything.
		re = neverMatches
	}
	return patternPart{re: boundMatching(re)}, true
}

// endAnchors rewrites the end-of-input anchors a source was built with, from
// the `$` JavaScript reads as the very end of the input to the `\z` regexp2
// reads that way. A bare `$` also matches ahead of a `\n` that ends the input
// under the .NET rules regexp2 follows.
//
// A `$` a pattern asked for itself was escaped on the way in, and one that
// rewriting a negated list swept into a character class stands for itself
// there, so every other one is an anchor.
//
// The anchors stay a single character until here so that the source keeps the
// offsets a negated list recorded while it was being written.
func endAnchors(source string) string {
	if !strings.ContainsRune(source, '$') {
		return source
	}
	var anchored strings.Builder
	escaping := false
	inClass := false
	for i := range len(source) {
		c := source[i]
		switch {
		case escaping:
			escaping = false
		case c == '\\':
			escaping = true
		case inClass:
			inClass = c != ']'
		case c == '[':
			inClass = true
		case c == '$':
			anchored.WriteString(`\z`)
			continue
		}
		anchored.WriteByte(c)
	}
	return anchored.String()
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
	if overMaxPatternLength(pattern) {
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
			// A class naming nothing, `[!]` or `[^]`, is syntax JavaScript
			// alone reads, and it matches any one character. Spell that out:
			// a regexp character class cannot be empty, so the fallback below
			// would otherwise read the whole thing as a literal.
			if re[reClassStart:] == "[^" {
				re = re[:reClassStart] + anyChar
				hasMagic = true
				inClass = false
				continue
			}
			// Handle the case where we left a class open: `[z-a]` is valid and
			// equivalent to `\[z-a\]`. Split where the last `[` was and re-walk
			// the contents so any character that was passed through as-is gets
			// translated.
			//
			// What decides that is whether the original class compiles as an
			// ECMAScript regexp. regexp2 follows .NET character-class syntax,
			// which accepts and rejects a different set of classes.
			if !isValidJSClass(pattern[classStart : i+1]) {
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
			if m.options.NoCase {
				re = re[:reClassStart+1] + esregexp.CaseCloseClass(re[reClassStart+1:], false)
			}
			re += string(c)

		default:
			// swallow any state character that wasn't consumed
			clearStateChar()
			if escaping {
				escaping = false
			} else if strings.ContainsRune(reSpecials, c) && (c != '^' || !inClass) {
				re += `\`
			}
			// A character class is widened as a whole once it closes, so only
			// a literal standing on its own is widened here.
			// minimatch builds `new RegExp(re, "i")`, with no `u`, so the
			// comparison is the one that never crosses into ASCII.
			if m.options.NoCase && !inClass {
				if class, widened := esregexp.CaseClass(c, false); widened {
					re += class
					continue
				}
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
		re = "(?=" + nonTerminator + ")" + re
	}

	if addPatternStart {
		re = patternStart + re
	}

	return re, hasMagic, true
}

// isValidJSClass asks the ECMAScript scanner whether class is a complete
// character class. minimatch uses the RegExp constructor for this check; the
// scanner accepts a literal instead, so delimiters and raw line terminators
// are escaped without changing the pattern the regexp parser sees.
func isValidJSClass(class string) bool {
	var literal strings.Builder
	literal.Grow(len(class) + 2)
	literal.WriteByte('/')
	backslashes := 0
	for _, r := range class {
		switch r {
		case '\\':
			literal.WriteRune(r)
			backslashes++
			continue
		case '/':
			if backslashes%2 == 0 {
				literal.WriteByte('\\')
			}
			literal.WriteRune(r)
		case '\n':
			if backslashes%2 == 0 {
				literal.WriteByte('\\')
			}
			literal.WriteByte('n')
		case '\r':
			if backslashes%2 == 0 {
				literal.WriteByte('\\')
			}
			literal.WriteByte('r')
		case '\u2028':
			if backslashes%2 == 0 {
				literal.WriteByte('\\')
			}
			literal.WriteString("u2028")
		case '\u2029':
			if backslashes%2 == 0 {
				literal.WriteByte('\\')
			}
			literal.WriteString("u2029")
		default:
			literal.WriteRune(r)
		}
		backslashes = 0
	}
	literal.WriteByte('/')
	return ecmascript.IsValidRegexLiteral(literal.String())
}

// Match reports whether path matches the compiled pattern.
func (m *Matcher) Match(path string) bool {
	if m.invalid || m.comment {
		return false
	}
	if m.empty {
		return path == ""
	}
	if m.options.Partial && path == "/" {
		return true
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
	if m.options.Partial {
		for firstGlobStar, part := range row {
			if part.globstar {
				return m.matchPartialGlobStar(file, row, firstGlobStar)
			}
		}
	}
	return m.matchOneFrom(file, row, 0, 0, nil)
}

// partialGlobStarSection is a run of ordinary pattern parts between two
// `**` tokens. after is the last file offset where minimatch 3.1.5 tries that
// run before accepting the file as a viable prefix.
type partialGlobStarSection struct {
	parts []patternPart
	after int
}

// matchPartialGlobStar follows minimatch 3.1.5's separate partial-mode
// globstar algorithm. Its cutoffs are deliberately not the same as an exact
// recursive globstar match: once too little of the current file remains to
// place a later section, it accepts the file as a prefix without inspecting
// every remaining part. That counter-intuitive cutoff is observable for
// patterns with more than one `**`, including a dot-name near the cutoff.
func (m *Matcher) matchPartialGlobStar(file []string, row []patternPart, firstGlobStar int) bool {
	if !matchPartialPartsAt(file, row[:firstGlobStar], 0) {
		return false
	}
	fileIndex := firstGlobStar
	body := row[firstGlobStar+1:]
	if len(body) == 0 {
		for index := fileIndex; index < len(file); index++ {
			if m.isDotPart(file[index]) {
				return false
			}
		}
		return true
	}

	sections := []partialGlobStarSection{{}}
	nonGlobStarParts := 0
	nonGlobStarSums := []int{0}
	for _, part := range body {
		if part.globstar {
			nonGlobStarSums = append(nonGlobStarSums, nonGlobStarParts)
			sections = append(sections, partialGlobStarSection{})
			continue
		}
		last := len(sections) - 1
		sections[last].parts = append(sections[last].parts, part)
		nonGlobStarParts++
	}
	for sectionIndex, sumIndex := 0, len(nonGlobStarSums)-1; sectionIndex < len(sections); sectionIndex, sumIndex = sectionIndex+1, sumIndex-1 {
		sections[sectionIndex].after = len(file) - (nonGlobStarSums[sumIndex] + len(sections[sectionIndex].parts))
	}

	failed := make(map[[2]int]struct{})
	return m.matchPartialGlobStarSections(file, sections, fileIndex, 0, failed)
}

func (m *Matcher) matchPartialGlobStarSections(file []string, sections []partialGlobStarSection, fileIndex int, sectionIndex int, failed map[[2]int]struct{}) bool {
	state := [2]int{fileIndex, sectionIndex}
	if _, found := failed[state]; found {
		return false
	}
	if sectionIndex == len(sections) {
		sawPart := false
		for ; fileIndex < len(file); fileIndex++ {
			sawPart = true
			if m.isDotPart(file[fileIndex]) {
				failed[state] = struct{}{}
				return false
			}
		}
		if !sawPart {
			failed[state] = struct{}{}
		}
		return sawPart
	}

	section := sections[sectionIndex]
	for fileIndex <= section.after {
		if matchPartialPartsAt(file, section.parts, fileIndex) &&
			m.matchPartialGlobStarSections(file, sections, fileIndex+len(section.parts), sectionIndex+1, failed) {
			return true
		}
		// minimatch 3.1.5 throws if its section cursor has already run past
		// the file. A malformed option must not crash the linter, so this port
		// treats that otherwise unreachable branch as a failed match. With
		// Dot enabled upstream skips the unsafe dot-name check and advances.
		if fileIndex >= len(file) {
			if m.options.Dot {
				fileIndex++
				continue
			}
			failed[state] = struct{}{}
			return false
		}
		if m.isDotPart(file[fileIndex]) {
			failed[state] = struct{}{}
			return false
		}
		fileIndex++
	}
	return true
}

// matchPartialPartsAt mirrors the bounded slice minimatch hands its ordinary
// matcher while placing one section. Running out of file is success in partial
// mode; a cursor beyond the file is the unsafe upstream branch handled above.
func matchPartialPartsAt(file []string, parts []patternPart, fileIndex int) bool {
	if fileIndex > len(file) {
		return false
	}
	for partIndex, part := range parts {
		index := fileIndex + partIndex
		if index >= len(file) {
			return true
		}
		if part.globstar || !part.match(file[index]) {
			return false
		}
	}
	return true
}

// matchOneFrom matches the parts of a path from fi on against the parts of one
// pattern row from pi on.
//
// Every recursion resumes at some pair of those two offsets and reads nothing
// else, so a pair that has already failed cannot succeed a second time: failed
// records the pairs that did, which is what keeps a pattern carrying several
// `**` from exploring exponentially many ways to divide the path between them.
// It is allocated only once a `**` actually has something to backtrack over.
func (m *Matcher) matchOneFrom(file []string, row []patternPart, fi int, pi int, failed []bool) bool {
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
			if failed == nil {
				failed = make([]bool, fl*(pl+1))
			}
			for fr := fi; fr < fl; fr++ {
				if state := fr*(pl+1) + pi + 1; !failed[state] {
					if m.matchOneFrom(file, row, fr, pi+1, failed) {
						return true
					}
					failed[state] = true
				}
				// `.` and `..` are never swallowed, and a dot-name only when
				// explicitly asked for.
				if m.isDotPart(file[fr]) {
					return false
				}
			}
			// In partial mode, consuming every remaining safe path part means
			// the file can still grow into the rest of the pattern. minimatch 3
			// uses this while walking a filesystem (for example, `a/x` is a
			// partial match for `a/**/b`). A dot part returns above because `**`
			// is not allowed to consume it under the current options.
			return m.options.Partial
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
		return m.options.Partial
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
