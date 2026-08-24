package filename_case

import (
	"reflect"
	"testing"
)

// TestSplitWordsParity locks in change-case@5.4 `split()` behaviour for
// shapes the rule_tester suite cannot directly exercise: empty inputs,
// single characters, all-caps, all-digits, ALL-CAPS + Title nesting (Pass 2
// firing multiple times), and consecutive non-alphanumeric runs collapsing
// into a single delimiter.
func TestSplitWordsParity(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		// Empty and single-char (no transitions, no delimiters).
		{"", nil},
		{"a", []string{"a"}},
		{"A", []string{"A"}},
		{"5", []string{"5"}},

		// All one class (no transitions).
		{"foo", []string{"foo"}},
		{"FOO", []string{"FOO"}},
		{"123", []string{"123"}},
		{"foo123", []string{"foo123"}},

		// Pass 1 only: lower/digit → upper.
		{"fooBar", []string{"foo", "Bar"}},
		{"iss47Spec", []string{"iss47", "Spec"}},
		{"5Foo", []string{"5", "Foo"}},

		// Pass 2 only: ALL-CAPS + Title boundary.
		{"FOOBar", []string{"FOO", "Bar"}},

		// Pass 1 + Pass 2 nested (the classic XMLHttpRequest case).
		{"XMLHttpRequest", []string{"XML", "Http", "Request"}},
		{"HTTPSConnection", []string{"HTTPS", "Connection"}},
		{"IOError", []string{"IO", "Error"}},
		{"parseHTTPSURL", []string{"parse", "HTTPSURL"}},

		// Pass 3: non-alphanumeric runs collapse into one delimiter.
		{"foo-bar", []string{"foo", "bar"}},
		{"foo--bar", []string{"foo", "bar"}},
		{"foo___bar", []string{"foo", "bar"}},
		{"foo-_-bar", []string{"foo", "bar"}},
		{"foo.bar.baz", []string{"foo", "bar", "baz"}},

		// Leading / trailing delimiters get trimmed.
		{"---foo---", []string{"foo"}},
		{"___foo___", []string{"foo"}},

		// Pure non-alphanumeric → no words.
		{"---", nil},
		{"...", nil},

		// All three passes together.
		{"foo-BarBaz", []string{"foo", "Bar", "Baz"}},
		{"FOO-barBaz", []string{"FOO", "bar", "Baz"}},

		// Unicode category boundaries mirror the upstream property escapes.
		{"éÉclair", []string{"é", "Éclair"}},
		{"ÉÉclair", []string{"É", "Éclair"}},
		// Title-case letters are \p{Lt}, not \p{Lu} or \p{Ll}, so they do not
		// participate in either camel/acronym boundary expression.
		{"\u01C5Foo", []string{"\u01C5Foo"}},
		// Combining marks and non-ASCII decimal digits are delimiters: the
		// upstream separator accepts \p{L} plus ECMAScript's ASCII-only \d.
		{"e\u0301Mail", []string{"e", "Mail"}},
		{"foo١Bar", []string{"foo", "Bar"}},
	}
	for _, c := range cases {
		got := splitWords(c.in)
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("splitWords(%q) = %#v, want %#v", c.in, got, c.want)
		}
	}
}

// TestFixFilenameDecoratedCartesianProduct locks in candidate de-duplication
// and left-to-right cartesian-product order when ignored decoration splits a
// filename into multiple independently converted chunks.
func TestFixFilenameDecoratedCartesianProduct(t *testing.T) {
	t.Run("order", func(t *testing.T) {
		leading, words := splitFilename("FOO_BAR[BAZ_QUX]")
		cases := []caseStyle{allCases[0], allCases[2]} // camel, kebab
		valid, invalidWord, candidates := validateFilename(words, cases)
		if valid {
			t.Fatal("expected filename to be invalid")
		}
		got := fixFilename(words, cases, invalidWord, candidates, leading, ".js")
		want := []string{
			"fooBar[bazQux].js",
			"fooBar[baz-qux].js",
			"foo-bar[bazQux].js",
			"foo-bar[baz-qux].js",
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("fixFilename() = %#v, want %#v", got, want)
		}
	})

	t.Run("deduplicate each chunk", func(t *testing.T) {
		leading, words := splitFilename("1_[1_]")
		cases := []caseStyle{allCases[0], allCases[2], allCases[4]} // camel, kebab, pascal
		valid, invalidWord, candidates := validateFilename(words, cases)
		if valid {
			t.Fatal("expected filename to be invalid")
		}
		got := fixFilename(words, cases, invalidWord, candidates, leading, ".js")
		want := []string{"1[1].js"}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("fixFilename() = %#v, want %#v", got, want)
		}
	})
}

func TestSplitFilenameNormalizesInvalidUTF8(t *testing.T) {
	input := string([]byte{'a', 0xff, 'B'})
	leading, got := splitFilename(input)
	want := []filenameWord{
		{word: "a"},
		{word: "�", ignored: true},
		{word: "B"},
	}
	if leading != "" || !reflect.DeepEqual(got, want) {
		t.Fatalf("splitFilename(%q) = (%q, %#v), want (%q, %#v)", input, leading, got, "", want)
	}
}

// TestPascalLikeTransformDigitBranch locks in the `_<digit>` branch:
// non-first words starting with a digit get a `_` prefix, but a first word
// starting with a digit does not.
func TestPascalLikeTransformDigitBranch(t *testing.T) {
	cases := []struct {
		word  string
		index int
		want  string
	}{
		// First-word digit start: NO leading `_`.
		{"123", 0, "123"},
		{"1foo", 0, "1foo"},
		{"5", 0, "5"},
		// Non-first-word digit start: leading `_`.
		{"123", 1, "_123"},
		{"1foo", 1, "_1foo"},
		{"5", 2, "_5"},
		// Non-first-word letter start: regular pascal-style capitalize.
		{"foo", 1, "Foo"},
		{"bar", 2, "Bar"},
		// First-word letter start: regular pascal-style capitalize.
		{"foo", 0, "Foo"},
		// Empty word stays empty regardless of index.
		{"", 0, ""},
		{"", 1, ""},
	}
	for _, c := range cases {
		got := pascalLikeTransform(c.word, c.index)
		if got != c.want {
			t.Errorf("pascalLikeTransform(%q, %d) = %q, want %q", c.word, c.index, got, c.want)
		}
	}
}

// TestEnglishishJoinOxford locks in oxford-comma + `or` formatting for 0/1/2/3/4
// items. The 4-item case is reachable when four `cases` are enabled and the
// filename violates all four (rare in practice but covered here so
// future ListFormat-style refactors can't silently change wording).
func TestEnglishishJoinOxford(t *testing.T) {
	cases := []struct {
		in   []string
		want string
	}{
		{nil, ""},
		{[]string{"a"}, "a"},
		{[]string{"a", "b"}, "a or b"},
		{[]string{"a", "b", "c"}, "a, b, or c"},
		{[]string{"a", "b", "c", "d"}, "a, b, c, or d"},
		{[]string{"a", "b", "c", "d", "e"}, "a, b, c, d, or e"},
	}
	for _, c := range cases {
		got := englishishJoin(c.in)
		if got != c.want {
			t.Errorf("englishishJoin(%#v) = %q, want %q", c.in, got, c.want)
		}
	}
}
