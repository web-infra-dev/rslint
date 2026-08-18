package ecmascript

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/utils/ecmascript/unicode17"
)

// TestUnicode17Mappings walks every character the delta names, which is what
// case_test.go's table can only sample. Both directions have to reach past the
// caser, since it is built on an older edition of Unicode than JavaScript
// reads.
func TestUnicode17Mappings(t *testing.T) {
	for _, pair := range unicode17Pairs(t) {
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

// TestUnicode17Cased walks the same characters through the Final_Sigma
// condition, where what matters is not their mapping but that they are cased at
// all: a capital sigma is final when a cased character comes before it and none
// comes after.
func TestUnicode17Cased(t *testing.T) {
	for _, pair := range unicode17Pairs(t) {
		for _, r := range pair {
			before := string(r) + "Σ"
			if got := StringToLowercase(before); got != StringToLowercase(string(r))+"ς" {
				t.Errorf("StringToLowercase(%UΣ) = %U, want a final sigma", r, []rune(got))
			}
			after := "ΑΣ" + string(r)
			if got := StringToLowercase(after); got != "ασ"+StringToLowercase(string(r)) {
				t.Errorf("StringToLowercase(ΑΣ%U) = %U, want a plain sigma", r, []rune(got))
			}
		}
	}
}

// unicode17Pairs reads the delta's mapping data back out as the {lower, upper}
// pairs the tests here want. Only the lower half of a pair has an uppercase the
// delta knows, so the walk names each pair once.
func unicode17Pairs(t *testing.T) [][2]rune {
	t.Helper()
	var pairs [][2]rune
	for _, r := range unicode17.CaseAdditions() {
		if upper, ok := unicode17.ToUpper(r); ok {
			pairs = append(pairs, [2]rune{r, upper})
		}
	}
	if len(pairs) == 0 {
		t.Fatal("the delta names no case mapping at all")
	}
	return pairs
}
