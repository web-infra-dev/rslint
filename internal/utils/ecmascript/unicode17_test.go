package ecmascript

import (
	"slices"
	"strconv"
	"strings"
	"testing"
	"unicode"
)

// TestUnicode17DeltaStillNeeded is the marker on unicode17.go: it fails once
// the toolchain's own tables cover what the delta stands in for, which is the
// moment the file should be deleted rather than maintained.
func TestUnicode17DeltaStillNeeded(t *testing.T) {
	major, _, _ := strings.Cut(unicode.Version, ".")
	edition, err := strconv.Atoi(major)
	if err != nil {
		t.Fatalf("cannot read the edition out of unicode.Version %q", unicode.Version)
	}
	if edition >= 17 {
		t.Fatalf("the standard library is on Unicode %s, so unicode17.go is obsolete: "+
			"delete it and unicode17_test.go, then drop the unicode17ToUpper call in "+
			"Canonicalize, the unicode17Remap calls in StringToUppercase and "+
			"StringToLowercase, and the unicode17CaseAdditions loop in caseTables",
			unicode.Version)
	}

	for _, pair := range unicode17Delta() {
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

// TestUnicode17Canonicalize covers what the delta is for: JavaScript compares
// each of these pairs equal under a regexp's `iu` flags, and Go's own tables
// would compare them unequal. The `i` flag on its own canonicalizes one UTF-16
// code unit at a time, so it joins the pairs inside the basic plane and leaves
// the two supplementary-plane scripts apart — which is checked too, because the
// delta must not reach further than JavaScript does.
func TestUnicode17Canonicalize(t *testing.T) {
	for _, pair := range unicode17Delta() {
		lower, upper := pair[0], pair[1]
		group := []rune{min(lower, upper), max(lower, upper)}
		joinsWithoutU := lower <= 0xFFFF

		if got := Canonicalize(lower, true); got != group[0] {
			t.Errorf("Canonicalize(%U, true) = %U, want %U", lower, got, group[0])
		}
		if got := Canonicalize(upper, true); got != group[0] {
			t.Errorf("Canonicalize(%U, true) = %U, want %U", upper, got, group[0])
		}
		if got := CaseEquivalents(lower, true); !slices.Equal(got, group) {
			t.Errorf("CaseEquivalents(%U, true) = %U, want %U", lower, got, group)
		}

		wantCanonical, wantEquivalents := lower, []rune(nil)
		if joinsWithoutU {
			wantCanonical, wantEquivalents = upper, group
		}
		if got := Canonicalize(lower, false); got != wantCanonical {
			t.Errorf("Canonicalize(%U, false) = %U, want %U", lower, got, wantCanonical)
		}
		if got := Canonicalize(upper, false); got != upper {
			t.Errorf("Canonicalize(%U, false) = %U, want it to stand for itself", upper, got)
		}
		if got := CaseEquivalents(lower, false); !slices.Equal(got, wantEquivalents) {
			t.Errorf("CaseEquivalents(%U, false) = %U, want %U", lower, got, wantEquivalents)
		}
	}
}

// TestUnicode17Mappings covers the other side of the delta: the caser the
// string functions are built on is an older edition too, so both directions
// have to reach past it.
func TestUnicode17Mappings(t *testing.T) {
	for _, pair := range unicode17Delta() {
		lower, upper := string(pair[0]), string(pair[1])
		if got := StringToUppercase(lower); got != upper {
			t.Errorf("StringToUppercase(%U) = %U, want %U", pair[0], []rune(got), pair[1])
		}
		if got := StringToLowercase(upper); got != lower {
			t.Errorf("StringToLowercase(%U) = %U, want %U", pair[1], []rune(got), pair[0])
		}
		// In a longer string too, so the splice that carries the rest of it
		// across is exercised as well.
		if got := StringToUppercase("a" + lower + "ß"); got != "A"+upper+"SS" {
			t.Errorf("StringToUppercase(a%Uß) = %q, want %q", pair[0], got, "A"+upper+"SS")
		}
	}
}

// unicode17Delta flattens the two shapes the delta is written in into the
// {lower, upper} pairs every test here wants.
func unicode17Delta() [][2]rune {
	pairs := append([][2]rune(nil), unicode17CasePairs[:]...)
	for _, run := range unicode17CaseRuns {
		for lower := run.lower; lower <= run.lastLower; lower++ {
			pairs = append(pairs, [2]rune{lower, lower + run.toUpper})
		}
	}
	return pairs
}
