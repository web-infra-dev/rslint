package unicode17

import (
	"testing"
	"unicode"
)

// toolchainEdition is the edition of Unicode every table in this package is
// written against. The data is a difference, not a set: it says what 17.0 has
// that this edition does not, so it is only correct while the toolchain is
// still on this one.
const toolchainEdition = "15.0.0"

// TestDeltaStillNeeded is the marker on this package: it fails the moment the
// toolchain stops being the edition the data was written against, which is
// either the moment the package should be deleted or the moment its tables
// have to be worked out again.
func TestDeltaStillNeeded(t *testing.T) {
	if unicode.Version != toolchainEdition {
		t.Fatalf("the standard library is on Unicode %s rather than %s, so this "+
			"package no longer says what it means. On 17.0.0 or later it is obsolete: "+
			"delete it, then drop the two lookups in ecmascript's isCased and "+
			"isCaseIgnorable, the unicode17Uppercase and toLower calls beside them, "+
			"the four in ecmascript/regexp's Canonicalize, simpleFold and "+
			"buildCaseTables, and the category questions the rules ask it. On an "+
			"edition in between, every table has to be recomputed against the new "+
			"one and this constant moved with them",
			unicode.Version, toolchainEdition)
	}

	for _, pair := range mappings() {
		if upper := unicode.ToUpper(pair[0]); upper != pair[0] {
			t.Errorf("unicode.ToUpper(%U) = %U on Unicode %s, so the delta no longer "+
				"has to carry this pair", pair[0], upper, unicode.Version)
		}
		if lower := unicode.ToLower(pair[1]); lower != pair[1] {
			t.Errorf("unicode.ToLower(%U) = %U on Unicode %s, so the delta no longer "+
				"has to carry this pair", pair[1], lower, unicode.Version)
		}
	}

	for _, pair := range foldPairs {
		if folded := unicode.SimpleFold(pair[0]); folded != pair[0] {
			t.Errorf("unicode.SimpleFold(%U) = %U on Unicode %s, so the delta no longer "+
				"has to carry this pair", pair[0], folded, unicode.Version)
		}
	}
}

// TestMappedCharactersAreCased covers the seam between the two halves of the
// data: a character the mapping or folding half names has a case, so the
// property half has to agree it is cased — either by naming it too, or by
// leaving it to a toolchain that already knows.
func TestMappedCharactersAreCased(t *testing.T) {
	for _, r := range append(CaseAdditions(), FoldAdditions()...) {
		cased, ok := Cased(r)
		if !ok {
			cased = casedByToolchain(r)
		}
		if !cased {
			t.Errorf("%U has a case mapping or folding but is not cased", r)
		}
	}
}

// TestRangeTablesAreWellFormed covers what unicode.Is takes on trust in a
// hand-written table: the ranges of each half rise, they do not touch, and
// none of them straddles the boundary between the two halves.
func TestRangeTablesAreWellFormed(t *testing.T) {
	tables := map[string]*unicode.RangeTable{
		"casedAdded":           casedAdded,
		"casedRemoved":         casedRemoved,
		"caseIgnorableAdded":   caseIgnorableAdded,
		"caseIgnorableRemoved": caseIgnorableRemoved,
		"upperAdded":           upperAdded,
		"lowerAdded":           lowerAdded,
		"lowerRemoved":         lowerRemoved,
		"letterAdded":          letterAdded,
		"markAdded":            markAdded,
	}
	for name, table := range tables {
		t.Run(name, func(t *testing.T) {
			previous := rune(-1)
			for _, r := range table.R16 {
				if rune(r.Hi) > 0xFFFF {
					t.Errorf("%U..%U belongs in R32", r.Lo, r.Hi)
				}
				previous = checkRange(t, rune(r.Lo), rune(r.Hi), rune(r.Stride), previous)
			}
			for _, r := range table.R32 {
				if rune(r.Lo) <= 0xFFFF {
					t.Errorf("%U..%U belongs in R16", r.Lo, r.Hi)
				}
				previous = checkRange(t, rune(r.Lo), rune(r.Hi), rune(r.Stride), previous)
			}
		})
	}
}

// checkRange reports on one range of a table and returns the highest character
// named so far, which is what the next range has to clear.
func checkRange(t *testing.T, lo, hi, stride, previous rune) rune {
	t.Helper()
	switch {
	case lo > hi:
		t.Errorf("%U..%U runs backwards", lo, hi)
	case lo <= previous:
		t.Errorf("%U..%U does not clear %U", lo, hi, previous)
	case stride != 1:
		t.Errorf("%U..%U has a stride of %d, which the data here does not use", lo, hi, stride)
	}
	return max(hi, previous)
}

// casedByToolchain is the Cased property as the standard library's own tables
// answer it, which is the derivation [Cased] stands in front of.
func casedByToolchain(r rune) bool {
	return unicode.In(r, unicode.Ll, unicode.Lu, unicode.Lt, unicode.Other_Lowercase, unicode.Other_Uppercase)
}

// mappings reads the two shapes the mapping data is written in back out as the
// {lower, upper} pairs the tests here want.
func mappings() [][2]rune {
	pairs := append([][2]rune(nil), casePairs[:]...)
	for _, run := range caseRuns {
		for lower := run.lower; lower <= run.lastLower; lower++ {
			pairs = append(pairs, [2]rune{lower, lower + run.toUpper})
		}
	}
	return pairs
}

// TestAddedCharactersAreStillUnknown covers every character the tables add: the
// toolchain has to still be answering no for it, or the entry has outlived the
// edition it was written for.
func TestAddedCharactersAreStillUnknown(t *testing.T) {
	tables := []struct {
		name  string
		table *unicode.RangeTable
		known func(rune) bool
	}{
		{"casedAdded", casedAdded, casedByToolchain},
		{"upperAdded", upperAdded, func(r rune) bool { return unicode.Is(unicode.Lu, r) }},
		{"lowerAdded", lowerAdded, func(r rune) bool { return unicode.Is(unicode.Ll, r) }},
		{"letterAdded", letterAdded, unicode.IsLetter},
		{"markAdded", markAdded, func(r rune) bool { return unicode.Is(unicode.M, r) }},
	}
	for _, table := range tables {
		t.Run(table.name, func(t *testing.T) {
			for _, r := range runesOf(table.table) {
				if table.known(r) {
					t.Errorf("the standard library already knows %U on Unicode %s", r, unicode.Version)
				}
			}
		})
	}
}

// TestCasedLettersAreLetters covers the seam between the tables: a character
// one of them makes an uppercase or lowercase letter has to be a letter in the
// one that answers for the category as a whole.
func TestCasedLettersAreLetters(t *testing.T) {
	for _, table := range []*unicode.RangeTable{upperAdded, lowerAdded} {
		for _, r := range runesOf(table) {
			if !IsLetter(r) {
				t.Errorf("%U is a cased letter but letterAdded does not name it", r)
			}
		}
	}
}

// runesOf walks out every character a table names.
func runesOf(table *unicode.RangeTable) []rune {
	var runes []rune
	for _, r := range table.R16 {
		for c := rune(r.Lo); c <= rune(r.Hi); c += rune(r.Stride) {
			runes = append(runes, c)
		}
	}
	for _, r := range table.R32 {
		for c := rune(r.Lo); c <= rune(r.Hi); c += rune(r.Stride) {
			runes = append(runes, c)
		}
	}
	return runes
}
