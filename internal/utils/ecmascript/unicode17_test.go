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
			"Canonicalize and the unicode17CaseAdditions loop in caseTables", unicode.Version)
	}

	for _, pair := range unicode17CasePairs {
		if upper := unicode.ToUpper(pair[0]); upper != pair[0] {
			t.Errorf("unicode.ToUpper(%U) = %U on Unicode %s, so the delta no longer "+
				"has to carry this pair", pair[0], upper, unicode.Version)
		}
	}
}

// TestUnicode17Canonicalize covers what the delta is for: JavaScript compares
// each of these pairs equal under a regexp's `i` flag, and Go's own tables
// would compare them unequal. Both readings of that flag have to reach the
// delta — the uppercase reading lands on the pair's uppercase, the folding
// one on the lower of the two — so both are checked.
func TestUnicode17Canonicalize(t *testing.T) {
	for _, pair := range unicode17CasePairs {
		lower, upper := pair[0], pair[1]
		group := []rune{min(lower, upper), max(lower, upper)}

		if got := Canonicalize(lower, false); got != upper {
			t.Errorf("Canonicalize(%U, false) = %U, want %U", lower, got, upper)
		}
		if got := Canonicalize(upper, false); got != upper {
			t.Errorf("Canonicalize(%U, false) = %U, want it to stand for itself", upper, got)
		}
		if got := Canonicalize(lower, true); got != group[0] {
			t.Errorf("Canonicalize(%U, true) = %U, want %U", lower, got, group[0])
		}
		if got := Canonicalize(upper, true); got != group[0] {
			t.Errorf("Canonicalize(%U, true) = %U, want %U", upper, got, group[0])
		}

		for _, unicodeMode := range []bool{false, true} {
			if got := CaseEquivalents(lower, unicodeMode); !slices.Equal(got, group) {
				t.Errorf("CaseEquivalents(%U, %v) = %U, want %U", lower, unicodeMode, got, group)
			}
		}
	}
}
