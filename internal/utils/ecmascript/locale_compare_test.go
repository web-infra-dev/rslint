// cspell:ignore aaland arevela heploc

package ecmascript

import (
	"encoding/json"
	"math/rand/v2"
	"os"
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
		{name: "root punctuation before digits before letters", locale: "", input: []string{"a1", "a-b", "a_b", "a"}, want: []string{"a", "a_b", "a-b", "a1"}},
		{name: "auto is the root order", locale: "auto", input: []string{"B", "a", "A", "b"}, want: []string{"a", "A", "b", "B"}},
		{name: "unknown locale is the root order", locale: "xx-YY", input: []string{"ch", "d", "cz"}, want: []string{"ch", "cz", "d"}},
		{name: "Danish contraction", locale: "da", input: []string{"aa", "b"}, want: []string{"b", "aa"}},
		{name: "Danish default upper first", locale: "da", input: []string{"ch", "cH", "Ch", "CH"}, want: []string{"CH", "Ch", "cH", "ch"}},
		{name: "Danish contraction case weights", locale: "da", input: []string{"aa", "Aa", "AA", "aA"}, want: []string{"aA", "AA", "Aa", "aa"}},
		{name: "Norwegian", locale: "nb", input: []string{"aa", "b"}, want: []string{"b", "aa"}},
		{name: "Norwegian with region", locale: "no-NO", input: []string{"aa", "b"}, want: []string{"b", "aa"}},
		{name: "Norwegian Nynorsk", locale: "nn-NO", input: []string{"aa", "b"}, want: []string{"b", "aa"}},
		{name: "Norwegian default lower first", locale: "nb", input: []string{"CH", "Ch", "cH", "ch"}, want: []string{"ch", "cH", "Ch", "CH"}},
		{name: "Norwegian contraction case weights", locale: "nb", input: []string{"AA", "Aa", "aa", "aA"}, want: []string{"aA", "aa", "Aa", "AA"}},
		{name: "Czech ch follows h", locale: "cs", input: []string{"i", "ch", "hz"}, want: []string{"hz", "ch", "i"}},
		{name: "Slovak ch follows h", locale: "sk-SK", input: []string{"onChange", "onFileRemove"}, want: []string{"onFileRemove", "onChange"}},
		{name: "traditional Spanish collation", locale: "es-u-co-trad", input: []string{"d", "ch", "cz"}, want: []string{"cz", "ch", "d"}},
		{name: "modern Spanish has no ch contraction", locale: "es", input: []string{"d", "cz", "ch"}, want: []string{"ch", "cz", "d"}},
		{name: "Lithuanian y between i and j", locale: "lt", input: []string{"j", "y", "i"}, want: []string{"i", "y", "j"}},
		{name: "explicit upper first", locale: "nb-u-kf-upper", input: []string{"ch", "cH", "Ch", "CH"}, want: []string{"CH", "Ch", "cH", "ch"}},
		{name: "explicit lower first", locale: "da-u-kf-lower", input: []string{"CH", "Ch", "cH", "ch"}, want: []string{"ch", "cH", "Ch", "CH"}},
		{name: "unsupported case first uses locale default", locale: "da-u-kf-foobar", input: []string{"a", "A"}, want: []string{"A", "a"}},
		{name: "valueless case first uses locale default", locale: "da-u-kf", input: []string{"a", "A"}, want: []string{"A", "a"}},
		{name: "Maltese default upper first", locale: "mt", input: []string{"a", "A"}, want: []string{"A", "a"}},
		{name: "numeric", locale: "nb-u-kn-true", input: []string{"a10", "a2"}, want: []string{"a2", "a10"}},
		{name: "valueless numeric is true", locale: "en-u-kn", input: []string{"a10", "a2"}, want: []string{"a2", "a10"}},
		{name: "numeric ignores leading zeros", locale: "en-u-kn", input: []string{"a010", "a9", "a0"}, want: []string{"a0", "a9", "a010"}},
		{name: "numeric false", locale: "en-u-kn-false", input: []string{"a2", "a10"}, want: []string{"a10", "a2"}},
		{name: "unsupported Nordic collation falls back to default", locale: "nb-u-co-phonebk", input: []string{"aa", "b"}, want: []string{"b", "aa"}},
		{name: "irrelevant strength key", locale: "en-u-ks-level1", input: []string{"A", "a"}, want: []string{"a", "A"}},
		// Locales select tables by truncation, without likely subtags.
		{name: "region does not imply Serbian Latin", locale: "sr-ME", input: []string{"lz", "lj"}, want: []string{"lj", "lz"}},
		{name: "explicit Serbian Latin", locale: "sr-Latn-ME", input: []string{"lj", "lz"}, want: []string{"lz", "lj"}},
		{name: "region does not imply Arabic script", locale: "az-IR", input: []string{"I", "x"}, want: []string{"x", "I"}},
		{name: "script without data keeps the language", locale: "da-Cyrl", input: []string{"aa", "b"}, want: []string{"b", "aa"}},
		{name: "alias with script", locale: "sh", input: []string{"lj", "lz"}, want: []string{"lz", "lj"}},
		{name: "alias loses its script to an explicit one", locale: "sh-Cyrl", input: []string{"lz", "lj"}, want: []string{"lj", "lz"}},
		{name: "individual language code", locale: "azj", input: []string{"I", "x"}, want: []string{"x", "I"}},
		{name: "deprecated language code", locale: "tl", input: []string{"ng", "nz"}, want: []string{"nz", "ng"}},
		{name: "collation type with region", locale: "es-MX-u-co-trad", input: []string{"d", "ch", "cz"}, want: []string{"cz", "ch", "d"}},
		{name: "unsupported collation type", locale: "hr-u-co-trad", input: []string{"lz", "lj"}, want: []string{"lz", "lj"}},
		{name: "emoji collation replaces tailoring", locale: "cs-u-co-emoji", input: []string{"i", "ch"}, want: []string{"ch", "i"}},
		{name: "POSIX variant", locale: "en-US-posix", input: []string{"a", "_", "B"}, want: []string{"B", "_", "a"}},
		{name: "POSIX extension", locale: "en-US-u-va-posix", input: []string{"a", "_", "B"}, want: []string{"B", "_", "a"}},
		{name: "explicit va overrides the POSIX variant", locale: "en-US-posix-u-va-other", input: []string{"B", "a"}, want: []string{"a", "B"}},
		{name: "valueless va overrides the POSIX variant", locale: "en-US-posix-u-va", input: []string{"B", "a"}, want: []string{"a", "B"}},
		{name: "later va overrides a leading va-posix", locale: "en-US-u-va-posix-va-other", input: []string{"B", "a"}, want: []string{"a", "B"}},
		{name: "first va wins with another variant", locale: "en-US-fonipa-u-va-posix-va-other", input: []string{"a", "B"}, want: []string{"B", "a"}},
		{name: "POSIX variant needs to be the only variant", locale: "en-US-posix-fonipa", input: []string{"B", "a"}, want: []string{"a", "B"}},
		{name: "removed alias variant does not count", locale: "en-US-aaland-posix", input: []string{"a", "B"}, want: []string{"B", "a"}},
		{name: "removed alias variant does not count again", locale: "en-US-arevela-posix", input: []string{"a", "B"}, want: []string{"B", "a"}},
		{name: "replaced alias variant still counts", locale: "en-US-heploc-posix", input: []string{"B", "a"}, want: []string{"a", "B"}},
		{name: "variant rule changes the language", locale: "zh-hakka-u-kn", input: []string{"a9", "a10"}, want: []string{"a10", "a9"}},
		{name: "numeric type alias", locale: "en-u-kn-yes", input: []string{"a10", "a9"}, want: []string{"a9", "a10"}},
		{name: "repeated key keeps the first type", locale: "en-u-kn-false-kn-true", input: []string{"a9", "a10"}, want: []string{"a10", "a9"}},
		{name: "unsupported language ignores extension keys", locale: "xx-u-kn", input: []string{"a9", "a10"}, want: []string{"a10", "a9"}},
		{name: "POSIX needs the US region", locale: "en-GB-posix", input: []string{"B", "_", "a"}, want: []string{"_", "a", "B"}},
		{name: "POSIX numbers precede punctuation", locale: "en-US-u-va-posix-kn", input: []string{"!", "a10", "a9"}, want: []string{"!", "a9", "a10"}},
		{name: "case-insensitive tag", locale: "SR-latn", input: []string{"lj", "lz"}, want: []string{"lz", "lj"}},
		{name: "private use is ignored", locale: "en-x-cs", input: []string{"i", "ch"}, want: []string{"ch", "i"}},
		{name: "Thai ignores punctuation", locale: "th", input: []string{"a_c", "ab"}, want: []string{"ab", "a_c"}},
		{name: "Thai keeps symbols", locale: "th", input: []string{"ab", "a$c"}, want: []string{"a$c", "ab"}},
		{name: "Thai space is ignorable", locale: "th-TH", input: []string{"a c", "ab"}, want: []string{"ab", "a c"}},
		// Combining marks are secondary differences after NFD.
		{name: "accent is secondary", locale: "", input: []string{"É", "f", "é", "E", "e"}, want: []string{"e", "E", "é", "É", "f"}},
		{name: "accented letter sorts with its base", locale: "", input: []string{"zeta", "bé", "alpha"}, want: []string{"alpha", "bé", "zeta"}},
		{name: "accented case is tertiary", locale: "", input: []string{"Éb", "éa", "éb"}, want: []string{"éa", "éb", "Éb"}},
		// Other characters outside printable ASCII use a stable fallback order:
		// after ASCII, by simple lowercase mapping, with case as a tertiary
		// difference.
		{name: "other characters sort after ASCII", locale: "", input: []string{"ж", "z", "ⓐ"}, want: []string{"z", "ж", "ⓐ"}},
		{name: "Roman numerals upper first", locale: "en-u-kf-upper", input: []string{"ⅰ", "Ⅰ"}, want: []string{"Ⅰ", "ⅰ"}},
		{name: "Roman numerals Danish", locale: "da", input: []string{"ⅰ", "Ⅰ"}, want: []string{"Ⅰ", "ⅰ"}},
		{name: "Roman numerals Maltese", locale: "mt", input: []string{"ⅰ", "Ⅰ"}, want: []string{"Ⅰ", "ⅰ"}},
		{name: "circled letters upper first", locale: "en-u-kf-upper", input: []string{"ⓐ", "Ⓐ"}, want: []string{"Ⓐ", "ⓐ"}},
		{name: "Roman numerals and letters", locale: "en-u-kf-upper", input: []string{"Ⅰ", "I", "ⅰ", "i"}, want: []string{"I", "i", "Ⅰ", "ⅰ"}},
		{name: "Roman numeral case precedes suffix case", locale: "en-u-kf-upper", input: []string{"ⅰA", "Ⅰa"}, want: []string{"Ⅰa", "ⅰA"}},
		{name: "Roman numerals explicit lower first", locale: "da-u-kf-lower", input: []string{"Ⅰ", "ⅰ"}, want: []string{"ⅰ", "Ⅰ"}},
		{name: "Roman numerals case first disabled", locale: "en-u-kf-false", input: []string{"Ⅰ", "ⅰ"}, want: []string{"ⅰ", "Ⅰ"}},
		{name: "title-case letters precede lowercase", locale: "en-u-kf-upper", input: []string{"ǆ", "ǅ"}, want: []string{"ǅ", "ǆ"}},
		{name: "control characters are ignorable", locale: "", input: []string{"a\x01c", "ab"}, want: []string{"ab", "a\x01c"}},
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

// TestLocaleComparerMatchesICU checks the ASCII tables against corpora sorted
// by Intl.Collator when the tables were generated.
func TestLocaleComparerMatchesICU(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("testdata/locale_collation_golden.json")
	if err != nil {
		t.Fatal(err)
	}
	var golden []struct {
		Locale string   `json:"locale"`
		Sorted []string `json:"sorted"`
	}
	if err := json.Unmarshal(data, &golden); err != nil {
		t.Fatal(err)
	}
	if len(golden) == 0 {
		t.Fatal("empty golden data")
	}
	for _, entry := range golden {
		t.Run(entry.Locale, func(t *testing.T) {
			t.Parallel()

			if len(entry.Sorted) < 2 {
				t.Fatalf("golden corpus for %s has %d strings", entry.Locale, len(entry.Sorted))
			}
			comparer := NewLocaleComparer(entry.Locale)
			for i := 1; i < len(entry.Sorted); i++ {
				left, right := entry.Sorted[i-1], entry.Sorted[i]
				if comparer.Compare(left, right) >= 0 || comparer.Compare(right, left) <= 0 {
					t.Fatalf("ICU %s orders %q before %q; comparer disagrees", collationICUVersion, left, right)
				}
			}
			shuffled := slices.Clone(entry.Sorted)
			rand.New(rand.NewPCG(1, 2)).Shuffle(len(shuffled), func(i, j int) {
				shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
			})
			slices.SortStableFunc(shuffled, comparer.Compare)
			if !slices.Equal(shuffled, entry.Sorted) {
				t.Fatalf("sorted corpus differs from ICU %s", collationICUVersion)
			}
		})
	}
}

// TestLocaleComparerResolvesLikeICU checks the table each tag selects against
// the order Intl.Collator used for it when the tables were generated.
func TestLocaleComparerResolvesLikeICU(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("testdata/locale_collation_resolution.json")
	if err != nil {
		t.Fatal(err)
	}
	var resolution []struct {
		Locale     string `json:"locale"`
		Table      int    `json:"table"`
		UpperFirst bool   `json:"upperFirst"`
		Numeric    bool   `json:"numeric"`
	}
	if err := json.Unmarshal(data, &resolution); err != nil {
		t.Fatal(err)
	}
	if len(resolution) == 0 {
		t.Fatal("empty resolution data")
	}
	for _, entry := range resolution {
		comparer := NewLocaleComparer(entry.Locale)
		table := -1
		for i := range collationTables {
			if comparer.table == &collationTables[i] {
				table = i
			}
		}
		if table != entry.Table || comparer.upperFirst != entry.UpperFirst || comparer.numeric != entry.Numeric {
			t.Errorf("%s: got table %d (upper first %t, numeric %t), ICU %s uses table %d (upper first %t, numeric %t)",
				entry.Locale, table, comparer.upperFirst, comparer.numeric, collationICUVersion, entry.Table, entry.UpperFirst, entry.Numeric)
		}
	}
}

func TestLocaleComparerIsTotalOrder(t *testing.T) {
	t.Parallel()

	values := []string{"A", "Ａ", "Ⓐ", "a", "ａ", "ⓐ", "I", "Ⅰ", "i", "ⅰ", "ch", "Ch", "cH", "CH", "c", "h", "a1", "a10", "a01", "a_", "é", "E"}
	for _, locale := range []string{"", "en-u-kf-upper", "cs", "da", "cy-u-kn", "th", "en-US-u-va-posix-kn"} {
		comparer := NewLocaleComparer(locale)
		for _, left := range values {
			for _, right := range values {
				if got, reverse := comparer.Compare(left, right), comparer.Compare(right, left); got != -reverse {
					t.Fatalf("%s: Compare(%q, %q) = %d but reverse = %d", locale, left, right, got, reverse)
				}
				for _, third := range values {
					if comparer.Compare(left, right) < 0 && comparer.Compare(right, third) < 0 && comparer.Compare(left, third) >= 0 {
						t.Fatalf("%s: non-transitive order: %q < %q < %q", locale, left, right, third)
					}
				}
			}
		}
	}
}
