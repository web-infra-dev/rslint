package ecmascript

import (
	"slices"
	"sync"
	"unicode"
	"unicode/utf8"
)

// Canonicalize maps a character to the one JavaScript compares it as when case
// does not matter — a regexp carrying the `i` flag without the `u` flag, which
// is what a plain `/x/i` is.
//
// This is neither Go's case folding nor the Unicode simple folding a regexp
// engine reaches for when told to ignore case. JavaScript uppercases the
// character, with the one rule that a character outside ASCII never
// canonicalizes into ASCII. The two familiar consequences: U+03A3 Σ, U+03C3 σ
// and U+03C2 ς are one character to it, while U+212A K and `k` are two.
//
// Go's simple uppercase stands in for String.prototype.toUpperCase, once the
// characters whose full uppercase spans more than one UTF-16 code unit are
// held back — `ß` is the familiar one — because JavaScript keeps the original
// character there.
//
// https://tc39.es/ecma262/2024/multipage/text-processing.html#sec-runtime-semantics-canonicalize-ch
func Canonicalize(r rune) rune {
	// Canonicalizing one UTF-16 code unit at a time is what makes this the
	// answer for a regexp without the `u` flag, and it leaves a
	// supplementary-plane character to stand for itself.
	if r > 0xFFFF || expandsOnUppercase(r) {
		return r
	}
	for _, pair := range unicode17CasePairs {
		if r == pair[0] {
			return pair[1]
		}
	}
	upper := unicode.ToUpper(r)
	if r >= utf8.RuneSelf && upper < utf8.RuneSelf {
		return r
	}
	return upper
}

// expandsOnUppercase reports the characters whose simple uppercase is one
// character but whose full uppercase is several.
func expandsOnUppercase(r rune) bool {
	return r >= 0x1F80 && r <= 0x1F87 ||
		r >= 0x1F90 && r <= 0x1F97 ||
		r >= 0x1FA0 && r <= 0x1FA7 ||
		r == 0x1FB3 || r == 0x1FC3 || r == 0x1FF3
}

// unicode17CasePairs are the simple uppercase mappings Unicode 16 and 17 added
// that Go 1.26's Unicode 15 tables do not have. Node 26 canonicalizes by
// Unicode 17, and naming the delta here states that target without depending on
// a Node installation or a generated table.
var unicode17CasePairs = [...][2]rune{
	{0x019B, 0xA7DC},
	{0x0264, 0xA7CB},
	{0x1C8A, 0x1C89},
	{0xA7CD, 0xA7CC},
	{0xA7CF, 0xA7CE},
	{0xA7D3, 0xA7D2},
	{0xA7D5, 0xA7D4},
	{0xA7DB, 0xA7DA},
}

// CaseEquivalents returns every character that Canonicalize maps onto the same
// character as r, r included, or nil when r stands alone. The result is shared
// and must not be modified.
//
// A caller comparing one character against another can just canonicalize both.
// This is for a caller that has to widen a set — the members of a regexp
// character class, say — to cover everything a case-insensitive comparison
// would accept.
func CaseEquivalents(r rune) []rune {
	byMember, _ := caseTables()
	return byMember[r]
}

// CaseEquivalenceGroups returns every group of two or more characters that
// canonicalize alike. The groups are shared and must not be modified.
//
// Widening a range rather than a single character means asking which groups
// reach into it, which needs the groups enumerated rather than looked up.
func CaseEquivalenceGroups() [][]rune {
	_, groups := caseTables()
	return groups
}

// caseTables groups the characters that canonicalize alike. A group of one has
// nothing to widen, so only the rest are kept: as a lookup by member, and as a
// list to walk. Built on first use — a caller that never compares without
// regard to case never pays for it.
var caseTables = sync.OnceValues(func() (map[rune][]rune, [][]rune) {
	grouped := map[rune][]rune{}
	record := func(r rune) {
		canonical := Canonicalize(r)
		if !slices.Contains(grouped[canonical], r) {
			grouped[canonical] = append(grouped[canonical], r)
		}
	}
	// Every character that canonicalizes onto another one has a case mapping,
	// so the case ranges name them all.
	for _, caseRange := range unicode.CaseRanges {
		for r := rune(caseRange.Lo); r <= rune(caseRange.Hi); r++ {
			record(r)
			record(Canonicalize(r))
		}
	}
	// Except the ones Go has no case mapping for at all.
	for _, pair := range unicode17CasePairs {
		record(pair[0])
		record(pair[1])
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
})
