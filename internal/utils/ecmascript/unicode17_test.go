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
// would compare them unequal.
func TestUnicode17Canonicalize(t *testing.T) {
	for _, pair := range unicode17CasePairs {
		lower, upper := pair[0], pair[1]
		if got := Canonicalize(lower); got != upper {
			t.Errorf("Canonicalize(%U) = %U, want %U", lower, got, upper)
		}
		if got := Canonicalize(upper); got != upper {
			t.Errorf("Canonicalize(%U) = %U, want it to stand for itself", upper, got)
		}
		if want := []rune{min(lower, upper), max(lower, upper)}; !slices.Equal(CaseEquivalents(lower), want) {
			t.Errorf("CaseEquivalents(%U) = %U, want %U", lower, CaseEquivalents(lower), want)
		}
	}
}
