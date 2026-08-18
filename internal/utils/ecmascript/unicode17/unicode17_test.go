package unicode17

import (
	"strconv"
	"strings"
	"testing"
	"unicode"
)

// TestDeltaStillNeeded is the marker on this package: it fails once the
// toolchain's own tables cover what the data here stands in for, which is the
// moment the package should be deleted rather than maintained.
func TestDeltaStillNeeded(t *testing.T) {
	major, _, _ := strings.Cut(unicode.Version, ".")
	edition, err := strconv.Atoi(major)
	if err != nil {
		t.Fatalf("cannot read the edition out of unicode.Version %q", unicode.Version)
	}
	if edition >= 17 {
		t.Fatalf("the standard library is on Unicode %s, so this package is obsolete: "+
			"delete it, then drop the two lookups in ecmascript's isCased and "+
			"isCaseIgnorable, the unicode17Uppercase and toLower calls beside them, "+
			"and the three in ecmascript/regexp's Canonicalize, simpleFold and "+
			"buildCaseTables", unicode.Version)
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
}

// TestMappedCharactersAreCased covers the seam between the two halves of the
// data: a character the mapping half names has a case, so the property half has
// to agree it is cased — either by naming it too, or by leaving it to a
// toolchain that already knows.
func TestMappedCharactersAreCased(t *testing.T) {
	for _, r := range CaseAdditions() {
		cased, ok := Cased(r)
		if !ok {
			cased = casedByToolchain(r)
		}
		if !cased {
			t.Errorf("%U has a case mapping but is not cased", r)
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
