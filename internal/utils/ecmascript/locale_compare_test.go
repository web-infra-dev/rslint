package ecmascript

import (
	"slices"
	"sort"
	"testing"
)

func TestLocaleComparer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name, locale string
		input        []string
		want         []string
	}{
		{name: "Danish contraction", locale: "da", input: []string{"aa", "b"}, want: []string{"b", "aa"}},
		{name: "Danish default upper first", locale: "da", input: []string{"ch", "cH", "Ch", "CH"}, want: []string{"CH", "Ch", "cH", "ch"}},
		{name: "Danish contraction case weights", locale: "da", input: []string{"aa", "Aa", "AA", "aA"}, want: []string{"aA", "AA", "Aa", "aa"}},
		{name: "Norwegian", locale: "nb", input: []string{"aa", "b"}, want: []string{"b", "aa"}},
		{name: "Norwegian", locale: "no-NO", input: []string{"aa", "b"}, want: []string{"b", "aa"}},
		{name: "Norwegian Nynorsk", locale: "nn-NO", input: []string{"aa", "b"}, want: []string{"b", "aa"}},
		{name: "Norwegian default lower first", locale: "nb", input: []string{"CH", "Ch", "cH", "ch"}, want: []string{"ch", "cH", "Ch", "CH"}},
		{name: "Norwegian contraction case weights", locale: "nb", input: []string{"AA", "Aa", "aa", "aA"}, want: []string{"aA", "aa", "Aa", "AA"}},
		{name: "explicit upper first", locale: "nb-u-kf-upper", input: []string{"ch", "cH", "Ch", "CH"}, want: []string{"CH", "Ch", "cH", "ch"}},
		{name: "explicit lower first", locale: "da-u-kf-lower", input: []string{"CH", "Ch", "cH", "ch"}, want: []string{"ch", "cH", "Ch", "CH"}},
		{name: "unsupported case first uses locale default", locale: "da-u-kf-foobar", input: []string{"a", "A"}, want: []string{"A", "a"}},
		{name: "valueless case first uses locale default", locale: "da-u-kf", input: []string{"a", "A"}, want: []string{"A", "a"}},
		{name: "case first preserves non-case tertiary order", locale: "en-u-kf-upper", input: []string{"ａ", "a", "Ａ", "A"}, want: []string{"A", "Ａ", "a", "ａ"}},
		{name: "Maltese default upper first", locale: "mt", input: []string{"a", "A"}, want: []string{"A", "a"}},
		{name: "numeric", locale: "nb-u-kn-true", input: []string{"a10", "a2"}, want: []string{"a2", "a10"}},
		{name: "unsupported Nordic collation falls back to default", locale: "nb-u-co-phonebk", input: []string{"aa", "b"}, want: []string{"b", "aa"}},
		{name: "irrelevant strength key", locale: "en-u-ks-level1", input: []string{"A", "a"}, want: []string{"a", "A"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			comparer := NewLocaleComparer(test.locale)
			got := slices.Clone(test.input)
			sort.SliceStable(got, func(i, j int) bool {
				return comparer.Compare(got[i], got[j]) < 0
			})
			if !slices.Equal(got, test.want) {
				t.Fatalf("locale %q sorted %q, want %q", test.locale, got, test.want)
			}
		})
	}
}

func TestLocaleComparerUpperFirstIsTransitive(t *testing.T) {
	t.Parallel()

	comparer := NewLocaleComparer("en-u-kf-upper")
	values := []string{"A", "Ａ", "a", "ａ"}
	for _, left := range values {
		for _, middle := range values {
			for _, right := range values {
				if comparer.Compare(left, middle) < 0 && comparer.Compare(middle, right) < 0 && comparer.Compare(left, right) >= 0 {
					t.Fatalf("non-transitive order: %q < %q < %q", left, middle, right)
				}
			}
		}
	}
}
