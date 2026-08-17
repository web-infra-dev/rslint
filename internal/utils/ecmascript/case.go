package ecmascript

import (
	"slices"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"

	"github.com/microsoft/typescript-go/shim/stringutil"
)

// StringToUppercase ports String.prototype.toUpperCase, which maps every
// character of s to its full uppercase — the one Unicode's Default Case
// Conversion gives it, rather than the single-character mapping unicode.ToUpper
// stops at. `ß` becoming `SS` is the familiar case of the difference; `ﬁ`
// becoming `FI`, and `ΐ` becoming three characters, are the same rule.
//
// A lone surrogate is carried across untouched, as are bytes that are not
// UTF-8, which is what happens to any character with no uppercase anyway.
//
// https://tc39.es/ecma262/2024/multipage/text-processing.html#sec-string.prototype.touppercase
func StringToUppercase(s string) string {
	return unicode17Remap(stringutil.ToUpperJS(s), unicode17ToUpper)
}

// StringToLowercase ports String.prototype.toLowerCase, the other half of what
// [StringToUppercase] describes. Only one character lowercases to more than one
// — U+0130 LATIN CAPITAL LETTER I WITH DOT ABOVE keeps its dot as a combining
// mark — but the mapping still reads its surroundings once: a capital sigma
// lowercases to U+03C2 ς rather than U+03C3 σ when it ends a word.
//
// https://tc39.es/ecma262/2024/multipage/text-processing.html#sec-string.prototype.tolowercase
func StringToLowercase(s string) string {
	return unicode17Remap(stringutil.ToLowerJS(s), unicode17ToLower)
}

// StringToLocaleUppercase ports String.prototype.toLocaleUpperCase.
//
// The locale-sensitive mapping departs from the plain one for three languages
// — Turkish, Azeri and Lithuanian — and only ever for the host's own locale,
// which a linter has no business reading: the same source has to draw the same
// diagnostics on every machine it runs on. So this answers with the mapping
// [StringToUppercase] gives, which is what the locale-sensitive one comes to
// everywhere else.
//
// https://tc39.es/ecma262/2024/multipage/text-processing.html#sec-string.prototype.tolocaleuppercase
func StringToLocaleUppercase(s string) string {
	return StringToUppercase(s)
}

// StringToLocaleLowercase ports String.prototype.toLocaleLowerCase, and reads
// its locale the way [StringToLocaleUppercase] describes.
//
// https://tc39.es/ecma262/2024/multipage/text-processing.html#sec-string.prototype.tolocalelowercase
func StringToLocaleLowercase(s string) string {
	return StringToLowercase(s)
}

// unicode17Remap carries a mapped string the rest of the way to the edition of
// Unicode Node reads; see unicode17.go. The characters the delta names have no
// case mapping in the edition the caser was generated from, so it hands them
// back as they came in and a second pass can finish the job.
func unicode17Remap(mapped string, delta func(rune) (rune, bool)) string {
	// Every character the delta names lies outside ASCII, which is what most
	// strings hold nothing but.
	if !stringutil.ContainsNonASCII(mapped) {
		return mapped
	}
	// Only the characters the delta names are rewritten. Everything else is
	// spliced across as the bytes it already was, so a lone surrogate — which
	// the caser carries in an encoding of its own — comes through unharmed.
	var out strings.Builder
	spliced := 0
	for i, r := range mapped {
		remapped, ok := delta(r)
		if !ok {
			continue
		}
		if spliced == 0 {
			out.Grow(len(mapped))
		}
		out.WriteString(mapped[spliced:i])
		out.WriteRune(remapped)
		spliced = i + utf8.RuneLen(r)
	}
	if spliced == 0 {
		return mapped
	}
	out.WriteString(mapped[spliced:])
	return out.String()
}

// Canonicalize maps a character to the one JavaScript compares it as when case
// does not matter. Which reading applies turns on whether the regexp carries
// `u` or `v`, and the two disagree:
//
//   - With one of those flags, ECMAScript folds by Unicode's simple case
//     folding, so `k`, `K` and U+212A KELVIN SIGN are one character, as are
//     `s`, `S` and U+017F LATIN SMALL LETTER LONG S.
//   - Without, ECMAScript uppercases the character, with the one rule that a
//     character outside ASCII never canonicalizes into ASCII. That splits the
//     two groups above back apart, while leaving U+03A3 Σ, U+03C3 σ and
//     U+03C2 ς as one character under either reading.
//
// Go's simple uppercase stands in for String.prototype.toUpperCase, once the
// characters whose full uppercase spans more than one UTF-16 code unit are
// held back — `ß` is the familiar one — because JavaScript keeps the original
// character there.
//
// https://tc39.es/ecma262/2024/multipage/text-processing.html#sec-runtime-semantics-canonicalize-ch
func Canonicalize(r rune, unicodeMode bool) rune {
	if unicodeMode {
		return simpleFold(r)
	}
	// Canonicalizing one UTF-16 code unit at a time is what makes this the
	// answer for a regexp without the `u` flag, and it leaves a
	// supplementary-plane character to stand for itself.
	if r > 0xFFFF || expandsOnUppercase(r) {
		return r
	}
	// Go's tables are an older Unicode than the one JavaScript reads; see
	// unicode17.go.
	if upper, ok := unicode17ToUpper(r); ok {
		return upper
	}
	upper := unicode.ToUpper(r)
	if r >= utf8.RuneSelf && upper < utf8.RuneSelf {
		return r
	}
	return upper
}

// simpleFold maps a character to the least of the characters it folds together
// with, which serves as the canonical form under `u` and `v`. Simple case
// folding carries no rule about ASCII and no rule about the supplementary
// plane, so neither guard above applies here.
func simpleFold(r rune) rune {
	least := r
	for folded := unicode.SimpleFold(r); folded != r; folded = unicode.SimpleFold(folded) {
		if folded < least {
			least = folded
		}
	}
	// Go's tables are an older Unicode than the one JavaScript reads; see
	// unicode17.go. A pair it adds folds together, so the lower of the two is
	// the canonical form for both.
	if other, ok := unicode17Fold(r); ok && other < least {
		least = other
	}
	return least
}

// expandsOnUppercase reports the characters whose simple uppercase is one
// character but whose full uppercase is several.
func expandsOnUppercase(r rune) bool {
	return r >= 0x1F80 && r <= 0x1F87 ||
		r >= 0x1F90 && r <= 0x1F97 ||
		r >= 0x1FA0 && r <= 0x1FA7 ||
		r == 0x1FB3 || r == 0x1FC3 || r == 0x1FF3
}

// CaseEquivalents returns every character that Canonicalize maps onto the same
// character as r, r included, or nil when r stands alone. The result is shared
// and must not be modified.
//
// A caller comparing one character against another can just canonicalize both.
// This is for a caller that has to widen a set — the members of a regexp
// character class, say — to cover everything a case-insensitive comparison
// would accept.
func CaseEquivalents(r rune, unicodeMode bool) []rune {
	byMember, _ := caseTables(unicodeMode)
	return byMember[r]
}

// CaseEquivalenceGroups returns every group of two or more characters that
// canonicalize alike. The groups are shared and must not be modified.
//
// Widening a range rather than a single character means asking which groups
// reach into it, which needs the groups enumerated rather than looked up.
func CaseEquivalenceGroups(unicodeMode bool) [][]rune {
	_, groups := caseTables(unicodeMode)
	return groups
}

// caseTables groups the characters that canonicalize alike, under whichever
// reading the caller's flags select. Each table is built on first use, so a
// caller that never compares without regard to case pays for neither, and one
// that only ever sees a `u` pattern pays for one.
func caseTables(unicodeMode bool) (map[rune][]rune, [][]rune) {
	if unicodeMode {
		return foldTables()
	}
	return uppercaseTables()
}

var (
	uppercaseTables = sync.OnceValues(func() (map[rune][]rune, [][]rune) { return buildCaseTables(false) })
	foldTables      = sync.OnceValues(func() (map[rune][]rune, [][]rune) { return buildCaseTables(true) })
)

// buildCaseTables collects the characters that canonicalize alike. A group of
// one has nothing to widen, so only the rest are kept: as a lookup by member,
// and as a list to walk.
func buildCaseTables(unicodeMode bool) (map[rune][]rune, [][]rune) {
	grouped := map[rune][]rune{}
	record := func(r rune) {
		canonical := Canonicalize(r, unicodeMode)
		if !slices.Contains(grouped[canonical], r) {
			grouped[canonical] = append(grouped[canonical], r)
		}
	}
	// Every character that canonicalizes onto another one has a case mapping,
	// so the case ranges name them all. Folding reaches further than
	// the uppercase mapping does, though — U+212A KELVIN SIGN joins `k` and
	// `K` without any of the three having the other as its uppercase — so each
	// orbit is walked out rather than assumed to be a pair.
	for _, caseRange := range unicode.CaseRanges {
		for r := rune(caseRange.Lo); r <= rune(caseRange.Hi); r++ {
			record(r)
			record(Canonicalize(r, unicodeMode))
			if unicodeMode {
				for folded := unicode.SimpleFold(r); folded != r; folded = unicode.SimpleFold(folded) {
					record(folded)
				}
			}
		}
	}
	// Except the ones Go has no case mapping for at all.
	for _, r := range unicode17CaseAdditions() {
		record(r)
	}

	byMember := map[rune][]rune{}
	groups := [][]rune{}
	for _, members := range grouped {
		if len(members) < 2 {
			continue
		}
		slices.Sort(members)
		groups = append(groups, members)
		for _, member := range members {
			byMember[member] = members
		}
	}
	return byMember, groups
}
