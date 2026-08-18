package ecmascript

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/utils/unicode17"
)

// Every expectation here is what Node 26 answers for the same call. The cases
// are the ones where the answer is more than a character-by-character walk of
// Go's own tables: a character whose case runs to several characters, the
// Final_Sigma context, and the characters Node knows a case for and Go does
// not.
func TestStringToUppercase(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "empty", in: "", want: ""},
		{name: "ascii", in: "aBcD", want: "ABCD"},
		{name: "ascii untouched", in: "42!", want: "42!"},
		{name: "eszett grows", in: "ß", want: "SS"},
		{name: "ligature grows", in: "ﬁ", want: "FI"},
		{name: "grows onto a combining mark", in: "ǰ", want: "J̌"},
		{name: "greek with iota subscript", in: "ᾀ", want: "ἈΙ"},
		{name: "title case", in: "ǅ", want: "Ǆ"},
		{name: "dotted capital i stands still", in: "İ", want: "İ"},
		{name: "unicode 16 pair", in: "ƛ", want: "Ƛ"},
		{name: "garay", in: "\U00010D70", want: "\U00010D50"},
		{name: "beria erfe", in: "\U00016EBB", want: "\U00016EA0"},
		{name: "in a longer string", in: "aß\U00010D70z", want: "ASS\U00010D50Z"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := StringToUppercase(test.in); got != test.want {
				t.Errorf("StringToUppercase(%U) = %U, want %U", []rune(test.in), []rune(got), []rune(test.want))
			}
			if got := StringToLocaleUppercase(test.in); got != test.want {
				t.Errorf("StringToLocaleUppercase(%U) = %U, want %U", []rune(test.in), []rune(got), []rune(test.want))
			}
		})
	}
}

func TestStringToLowercase(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "empty", in: "", want: ""},
		{name: "ascii", in: "ABCdef", want: "abcdef"},
		{name: "ascii untouched", in: "42!", want: "42!"},
		{name: "dotted capital i keeps its dot", in: "İ", want: "i̇"},
		{name: "title case", in: "ǅ", want: "ǆ"},
		{name: "unicode 16 pair", in: "Ƛ", want: "ƛ"},
		{name: "garay", in: "\U00010D50", want: "\U00010D70"},
		// A sigma is final when a cased character comes before it and none
		// comes after.
		{name: "sigma alone", in: "Σ", want: "σ"},
		{name: "sigma ends the word", in: "ΑΣ", want: "ας"},
		{name: "sigma starts the word", in: "ΣΑ", want: "σα"},
		{name: "sigma inside the word", in: "ΑΣΒ", want: "ασβ"},
		{name: "punctuation ends the word", in: "ΑΣ,Α", want: "ας,α"},
		// An apostrophe or an accent is ignored by the condition rather than
		// counted as the end of the word.
		{name: "apostrophe does not end the word", in: "ΑΣ'Β", want: "ασ'β"},
		{name: "apostrophe at the end", in: "ΑΣ'", want: "ας'"},
		{name: "accent before the sigma", in: "ΆΣ", want: "άς"},
		{name: "accent after the sigma", in: "ΑΣ́", want: "ας́"},
		// The properties that decide a final sigma moved on in Unicode 16 and
		// 17 too: these characters became cased, stopped being cased, became
		// case-ignorable and stopped being case-ignorable, in that order.
		{name: "character that became cased before the sigma", in: "\U00010D50Σ", want: "\U00010D70ς"},
		{name: "character that became cased after the sigma", in: "ΑΣ\U00010D50", want: "ασ\U00010D70"},
		{name: "character that stopped being cased before the sigma", in: "ʕΣ", want: "ʕσ"},
		{name: "character that stopped being cased after the sigma", in: "ΑΣʕ", want: "ας\u0295"},
		{name: "character that became case-ignorable before the sigma", in: "Α\u0897Σ", want: "α\u0897ς"},
		{name: "character that became case-ignorable after the sigma", in: "ΑΣ\u0897Α", want: "ασ\u0897α"},
		{name: "character that stopped being case-ignorable before the sigma", in: "Α\U0001171EΣ", want: "α\U0001171Eσ"},
		{name: "character that stopped being case-ignorable after the sigma", in: "ΑΣ\U0001171EΑ", want: "ας\U0001171Eα"},
		{name: "character that became both before the sigma", in: "Α\uA7F1Σ", want: "α\uA7F1ς"},
		{name: "character that became both after the sigma", in: "ΑΣ\uA7F1Α", want: "ασ\uA7F1α"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := StringToLowercase(test.in); got != test.want {
				t.Errorf("StringToLowercase(%U) = %U, want %U", []rune(test.in), []rune(got), []rune(test.want))
			}
			if got := StringToLocaleLowercase(test.in); got != test.want {
				t.Errorf("StringToLocaleLowercase(%U) = %U, want %U", []rune(test.in), []rune(got), []rune(test.want))
			}
		})
	}
}

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
