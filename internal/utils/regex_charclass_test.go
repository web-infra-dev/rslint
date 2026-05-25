package utils

import (
	"reflect"
	"testing"
)

// ---------------------------------------------------------------------------
// Layer 1: pattern scanner
// ---------------------------------------------------------------------------

func TestIterateRegexCharacterClasses_Basic(t *testing.T) {
	cases := []struct {
		pattern string
		flags   RegexFlags
		want    [][2]int // (start, end) pairs
	}{
		// Single class
		{"[abc]", RegexFlags{}, [][2]int{{0, 5}}},
		// Multiple classes
		{"[a][b]", RegexFlags{}, [][2]int{{0, 3}, {3, 6}}},
		// Class with content
		{"foo[abc]bar", RegexFlags{}, [][2]int{{3, 8}}},
		// Negated
		{"[^abc]", RegexFlags{}, [][2]int{{0, 6}}},
		// Empty class
		{"[]", RegexFlags{}, [][2]int{{0, 2}}},
		// Negated empty
		{"[^]", RegexFlags{}, [][2]int{{0, 3}}},
		// `\]` inside class — does not close
		{`[\]]`, RegexFlags{}, [][2]int{{0, 4}}},
		// `[` inside class is literal in non-v mode
		{"[[]", RegexFlags{}, [][2]int{{0, 3}}},
		// `\\` is escaped backslash
		{`[\\]`, RegexFlags{}, [][2]int{{0, 4}}},
		// Class outside escape paren
		{`(\[abc\])`, RegexFlags{}, nil},
		// Escaped `[` outside class
		{`\[abc\]`, RegexFlags{}, nil},
	}
	for _, c := range cases {
		t.Run(c.pattern, func(t *testing.T) {
			var got [][2]int
			ok := IterateRegexCharacterClasses(c.pattern, c.flags, func(s, e int) {
				got = append(got, [2]int{s, e})
			})
			if !ok {
				t.Fatalf("scan failed for %q", c.pattern)
			}
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("%q: got %v, want %v", c.pattern, got, c.want)
			}
		})
	}
}

func TestIterateRegexCharacterClasses_EscapedBracketsInVFlag(t *testing.T) {
	// v flag + `\]` inside nested class: the escape must not close the outer
	// class prematurely.
	var got [][2]int
	ok := IterateRegexCharacterClasses(`[[\]]]`, RegexFlags{UnicodeSets: true}, func(s, e int) {
		got = append(got, [2]int{s, e})
	})
	if !ok {
		t.Fatal("scan failed")
	}
	want := [][2]int{{1, 5}, {0, 6}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestIterateRegexCharacterClasses_DeepNesting(t *testing.T) {
	// Five levels of nesting under v flag.
	var got [][2]int
	ok := IterateRegexCharacterClasses(`[[[[[x]]]]]`, RegexFlags{UnicodeSets: true}, func(s, e int) {
		got = append(got, [2]int{s, e})
	})
	if !ok {
		t.Fatal("scan failed")
	}
	// Innermost first: 4..7, then 3..8, 2..9, 1..10, 0..11.
	want := [][2]int{{4, 7}, {3, 8}, {2, 9}, {1, 10}, {0, 11}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestIterateRegexCharacterClasses_QBraceContainsBracket(t *testing.T) {
	// `\q{ab|c]d}` inside v class — the `]` inside `\q{}` is literal and
	// must NOT close the outer class.
	var got [][2]int
	ok := IterateRegexCharacterClasses(`[\q{ab|c]d}]`, RegexFlags{UnicodeSets: true}, func(s, e int) {
		got = append(got, [2]int{s, e})
	})
	if !ok {
		t.Fatal("scan failed")
	}
	want := [][2]int{{0, 12}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestIterateRegexCharacterClasses_PropertyEscapeContainsBracket(t *testing.T) {
	// `\p{...}` can't contain `]` in practice (property names don't), but
	// exercise the general skip behavior with an adjacent class.
	var got [][2]int
	ok := IterateRegexCharacterClasses(`\p{Letter}[a]`, RegexFlags{Unicode: true}, func(s, e int) {
		got = append(got, [2]int{s, e})
	})
	if !ok {
		t.Fatal("scan failed")
	}
	want := [][2]int{{10, 13}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestParseRegexCharacterClass_SetDifferenceWithClass(t *testing.T) {
	els := parseClass(t, `[a--[b]]`, RegexFlags{UnicodeSets: true})
	// Elements: 'a', breaker(--), breaker(nested [b])
	if len(els) != 3 {
		t.Fatalf("len=%d (got %+v)", len(els), els)
	}
	if els[0].Kind != RegexCharSingle || els[0].Value != 'a' {
		t.Errorf("[0] = %+v", els[0])
	}
	if els[1].Kind != RegexCharBreaker {
		t.Errorf("[1] (--) = %+v", els[1])
	}
	if els[2].Kind != RegexCharBreaker {
		t.Errorf("[2] (nested [b]) = %+v", els[2])
	}
}

func TestParseRegexCharacterClass_MultipleRangesAndSets(t *testing.T) {
	// `[\dA-Z\sa-z]` — alternating sets and ranges.
	els := parseClass(t, `[\dA-Z\sa-z]`, RegexFlags{})
	wantKinds := []RegexCharElementKind{
		RegexCharBreaker,            // \d
		RegexCharRange,              // A-Z
		RegexCharBreaker,            // \s
		RegexCharRange,              // a-z
	}
	if len(els) != len(wantKinds) {
		t.Fatalf("len=%d, want %d (got %+v)", len(els), len(wantKinds), els)
	}
	for i, k := range wantKinds {
		if els[i].Kind != k {
			t.Errorf("[%d].Kind = %v, want %v", i, els[i].Kind, k)
		}
	}
}

func TestIterateRegexCharacterClasses_VFlagNested(t *testing.T) {
	// v-flag enables nested classes; we expect the callback to fire for
	// EACH nesting level (innermost first thanks to recursion).
	cases := []struct {
		pattern string
		want    [][2]int
	}{
		// Single level nesting
		{"[[a]]", [][2]int{{1, 4}, {0, 5}}},
		// Two-level nesting
		{"[[[x]]]", [][2]int{{2, 5}, {1, 6}, {0, 7}}},
		// Sibling nested classes
		{"[[a][b]]", [][2]int{{1, 4}, {4, 7}, {0, 8}}},
		// With set op
		{"[a--[b]]", [][2]int{{4, 7}, {0, 8}}},
		// With \q{}
		{`[a\q{xy}]`, [][2]int{{0, 9}}},
	}
	for _, c := range cases {
		t.Run(c.pattern, func(t *testing.T) {
			var got [][2]int
			ok := IterateRegexCharacterClasses(c.pattern, RegexFlags{UnicodeSets: true}, func(s, e int) {
				got = append(got, [2]int{s, e})
			})
			if !ok {
				t.Fatalf("scan failed for %q", c.pattern)
			}
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("%q: got %v, want %v", c.pattern, got, c.want)
			}
		})
	}
}

func TestIterateRegexCharacterClasses_EscapeBoundaries(t *testing.T) {
	// Escapes that span > 2 bytes must be skipped correctly so subsequent
	// `[` is recognized.
	cases := []struct {
		pattern string
		flags   RegexFlags
		want    [][2]int
	}{
		// `\xHH` followed by class
		{`\x41[ab]`, RegexFlags{}, [][2]int{{4, 8}}},
		// `\uHHHH` followed by class
		{`\u0041[ab]`, RegexFlags{}, [][2]int{{6, 10}}},
		// `\u{H}` under u flag
		{`\u{41}[ab]`, RegexFlags{Unicode: true}, [][2]int{{6, 10}}},
		// `\u{H}` under non-u: only `\u` consumed → `{41}` then `[ab]`
		{`\u{41}[ab]`, RegexFlags{}, [][2]int{{6, 10}}},
		// `\p{Letter}` under u flag
		{`\p{Letter}[ab]`, RegexFlags{Unicode: true}, [][2]int{{10, 14}}},
		// `\cX`
		{`\cI[ab]`, RegexFlags{}, [][2]int{{3, 7}}},
		// `\d` inside group then class
		{`(\d)[ab]`, RegexFlags{}, [][2]int{{4, 8}}},
	}
	for _, c := range cases {
		t.Run(c.pattern, func(t *testing.T) {
			var got [][2]int
			ok := IterateRegexCharacterClasses(c.pattern, c.flags, func(s, e int) {
				got = append(got, [2]int{s, e})
			})
			if !ok {
				t.Fatalf("scan failed for %q", c.pattern)
			}
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("%q: got %v, want %v", c.pattern, got, c.want)
			}
		})
	}
}

func TestIterateRegexCharacterClasses_Malformed(t *testing.T) {
	// Unterminated class at EOF
	for _, p := range []string{"[abc", "[", `[\`, `\`} {
		ok := IterateRegexCharacterClasses(p, RegexFlags{}, func(s, e int) {})
		if ok {
			t.Errorf("expected ok=false for %q", p)
		}
	}
}

// ---------------------------------------------------------------------------
// Layer 2: character class element parser
// ---------------------------------------------------------------------------

func parseClass(t *testing.T, pattern string, flags RegexFlags) []RegexCharElement {
	t.Helper()
	els, _, ok := ParseRegexCharacterClass(pattern, 0, flags)
	if !ok {
		t.Fatalf("ParseRegexCharacterClass(%q) failed", pattern)
	}
	return els
}

func TestParseRegexCharacterClass_BasicChars(t *testing.T) {
	els := parseClass(t, "[abc]", RegexFlags{})
	if len(els) != 3 {
		t.Fatalf("len=%d, want 3", len(els))
	}
	for i, want := range []uint32{'a', 'b', 'c'} {
		if els[i].Kind != RegexCharSingle || els[i].Value != want {
			t.Errorf("[%d] = %+v", i, els[i])
		}
	}
}

func TestParseRegexCharacterClass_Negated(t *testing.T) {
	// `[^abc]` — the `^` is consumed but does not appear as an element.
	els := parseClass(t, "[^abc]", RegexFlags{})
	if len(els) != 3 {
		t.Fatalf("len=%d, want 3", len(els))
	}
	if els[0].Value != 'a' || els[0].Start != 2 {
		t.Errorf("[0] = %+v", els[0])
	}
}

func TestParseRegexCharacterClass_Empty(t *testing.T) {
	els := parseClass(t, "[]", RegexFlags{})
	if len(els) != 0 {
		t.Errorf("got %v, want empty", els)
	}
}

func TestParseRegexCharacterClass_Range(t *testing.T) {
	els := parseClass(t, "[a-z]", RegexFlags{})
	if len(els) != 1 {
		t.Fatalf("len=%d, want 1", len(els))
	}
	r := els[0]
	if r.Kind != RegexCharRange || r.Value != 'a' || r.Max != 'z' {
		t.Errorf("range = %+v", r)
	}
	if r.Start != 1 || r.End != 4 {
		t.Errorf("source span = %d..%d, want 1..4", r.Start, r.End)
	}
}

func TestParseRegexCharacterClass_RangeWithEscape(t *testing.T) {
	// `[\u0041-\u005A]` — range from A to Z via \uHHHH
	els := parseClass(t, `[\u0041-\u005A]`, RegexFlags{})
	if len(els) != 1 {
		t.Fatalf("len=%d, want 1", len(els))
	}
	if els[0].Kind != RegexCharRange || els[0].Value != 0x41 || els[0].Max != 0x5A {
		t.Errorf("got %+v", els[0])
	}
}

func TestParseRegexCharacterClass_DashAtBoundary(t *testing.T) {
	// `[-a]` — leading `-` is literal
	els := parseClass(t, "[-a]", RegexFlags{})
	if len(els) != 2 || els[0].Value != '-' || els[1].Value != 'a' {
		t.Errorf("got %+v", els)
	}
	// `[a-]` — trailing `-` is literal
	els = parseClass(t, "[a-]", RegexFlags{})
	if len(els) != 2 || els[0].Value != 'a' || els[1].Value != '-' {
		t.Errorf("got %+v", els)
	}
}

func TestParseRegexCharacterClass_HexEscapes(t *testing.T) {
	els := parseClass(t, `[\x41\u0042\cI\n\t\\]`, RegexFlags{})
	wants := []uint32{0x41, 0x42, 0x09, '\n', '\t', '\\'}
	if len(els) != len(wants) {
		t.Fatalf("len=%d, want %d (got %+v)", len(els), len(wants), els)
	}
	for i, w := range wants {
		if els[i].Value != w {
			t.Errorf("[%d] value=%x want %x", i, els[i].Value, w)
		}
	}
}

func TestParseRegexCharacterClass_UBraceUnderUFlag(t *testing.T) {
	els := parseClass(t, `[\u{41}\u{1F44D}]`, RegexFlags{Unicode: true})
	if len(els) != 2 {
		t.Fatalf("len=%d, want 2", len(els))
	}
	if !els[0].IsUBrace || els[0].Value != 0x41 {
		t.Errorf("[0] = %+v", els[0])
	}
	if !els[1].IsUBrace || els[1].Value != 0x1F44D {
		t.Errorf("[1] = %+v", els[1])
	}
}

func TestParseRegexCharacterClass_UBraceUnderNonUFlag(t *testing.T) {
	// Non-u mode: `\u{41}` is treated as identity `u` then literal `{41}`.
	els := parseClass(t, `[\u{41}]`, RegexFlags{})
	if len(els) < 4 {
		t.Fatalf("len=%d, want >=4 (got %+v)", len(els), els)
	}
	if els[0].Value != 'u' {
		t.Errorf("[0] = %+v, want value='u'", els[0])
	}
}

func TestParseRegexCharacterClass_SurrogatePairCollapse(t *testing.T) {
	// Under u/v flag, `\uD83D\uDC4D` collapses to one element with astral value.
	els := parseClass(t, `[\uD83D\uDC4D]`, RegexFlags{Unicode: true})
	if len(els) != 1 {
		t.Fatalf("len=%d, want 1 (got %+v)", len(els), els)
	}
	if els[0].Value != 0x1F44D {
		t.Errorf("collapsed value = %x, want 1F44D", els[0].Value)
	}

	// Under non-u, they remain two units.
	els = parseClass(t, `[\uD83D\uDC4D]`, RegexFlags{})
	if len(els) != 2 {
		t.Fatalf("len=%d, want 2 (got %+v)", len(els), els)
	}
	if els[0].Value != 0xD83D || els[1].Value != 0xDC4D {
		t.Errorf("got %v", els)
	}
}

func TestParseRegexCharacterClass_CharacterSetBreaker(t *testing.T) {
	els := parseClass(t, `[a\db]`, RegexFlags{})
	wantKinds := []RegexCharElementKind{
		RegexCharSingle, RegexCharBreaker, RegexCharSingle,
	}
	if len(els) != len(wantKinds) {
		t.Fatalf("len=%d, want %d", len(els), len(wantKinds))
	}
	for i, k := range wantKinds {
		if els[i].Kind != k {
			t.Errorf("[%d].Kind = %v, want %v", i, els[i].Kind, k)
		}
	}
}

func TestParseRegexCharacterClass_PropertyEscape(t *testing.T) {
	els := parseClass(t, `[\p{Letter}a]`, RegexFlags{Unicode: true})
	if len(els) != 2 || els[0].Kind != RegexCharBreaker || els[1].Value != 'a' {
		t.Errorf("got %+v", els)
	}
	if els[0].Start != 1 || els[0].End != 11 {
		t.Errorf("breaker span = %d..%d, want 1..11", els[0].Start, els[0].End)
	}
}

func TestParseRegexCharacterClass_QuotedDisjunction(t *testing.T) {
	els := parseClass(t, `[\q{abc}]`, RegexFlags{UnicodeSets: true})
	if len(els) != 1 || els[0].Kind != RegexCharBreaker {
		t.Errorf("got %+v", els)
	}
}

func TestParseRegexCharacterClass_VFlagSetOperator(t *testing.T) {
	els := parseClass(t, `[a--b]`, RegexFlags{UnicodeSets: true})
	wantKinds := []RegexCharElementKind{
		RegexCharSingle, RegexCharBreaker, RegexCharSingle,
	}
	if len(els) != len(wantKinds) {
		t.Fatalf("len=%d, want %d (got %+v)", len(els), len(wantKinds), els)
	}
	for i, k := range wantKinds {
		if els[i].Kind != k {
			t.Errorf("[%d].Kind = %v want %v", i, els[i].Kind, k)
		}
	}

	els = parseClass(t, `[a&&b]`, RegexFlags{UnicodeSets: true})
	if len(els) != 3 || els[1].Kind != RegexCharBreaker {
		t.Errorf("&& got %+v", els)
	}
}

func TestParseRegexCharacterClass_VFlagNestedAsBreaker(t *testing.T) {
	// `[a[b]c]` under v flag: nested `[b]` appears as a breaker
	els := parseClass(t, "[a[b]c]", RegexFlags{UnicodeSets: true})
	if len(els) != 3 {
		t.Fatalf("len=%d (got %+v)", len(els), els)
	}
	if els[1].Kind != RegexCharBreaker {
		t.Errorf("[1] should be breaker: %+v", els[1])
	}
}

func TestParseRegexCharacterClass_BackspaceBInClass(t *testing.T) {
	// Inside class, `\b` is U+0008.
	els := parseClass(t, `[\b]`, RegexFlags{})
	if len(els) != 1 || els[0].Value != 0x08 {
		t.Errorf("got %+v", els)
	}
}

func TestParseRegexCharacterClass_RawAstral(t *testing.T) {
	// Raw 👍 in a class. Under non-u, this is a single-element single-rune
	// pattern (4 UTF-8 bytes); the regex engine in JS would see it as 2
	// surrogate code units, but we leave that responsibility to the caller
	// (e.g. the misleading-character-class rule passes string-literal data
	// already split into surrogates by ParseJSStringLiteralSource).
	els := parseClass(t, "[\U0001F44D]", RegexFlags{})
	if len(els) != 1 {
		t.Fatalf("len=%d", len(els))
	}
	if els[0].Value != 0x1F44D {
		t.Errorf("value=%x", els[0].Value)
	}
}

func TestParseRegexCharacterClass_Position(t *testing.T) {
	// Verify Start/End are byte offsets within the original pattern.
	els, end, ok := ParseRegexCharacterClass("foo[ab]bar", 3, RegexFlags{})
	if !ok {
		t.Fatal("parse failed")
	}
	if end != 7 {
		t.Errorf("end=%d, want 7", end)
	}
	if len(els) != 2 {
		t.Fatalf("len=%d, want 2", len(els))
	}
	if els[0].Start != 4 || els[0].End != 5 {
		t.Errorf("[0] span = %d..%d, want 4..5", els[0].Start, els[0].End)
	}
	if els[1].Start != 5 || els[1].End != 6 {
		t.Errorf("[1] span = %d..%d, want 5..6", els[1].Start, els[1].End)
	}
}

func TestParseRegexCharacterClass_RangeWithDashChain(t *testing.T) {
	// `[a-b-c]` — `a-b` range, then `-`, then `c`
	els := parseClass(t, "[a-b-c]", RegexFlags{})
	if len(els) != 3 {
		t.Fatalf("len=%d (got %+v)", len(els), els)
	}
	if els[0].Kind != RegexCharRange || els[0].Value != 'a' || els[0].Max != 'b' {
		t.Errorf("[0] = %+v", els[0])
	}
	if els[1].Kind != RegexCharSingle || els[1].Value != '-' {
		t.Errorf("[1] = %+v", els[1])
	}
	if els[2].Kind != RegexCharSingle || els[2].Value != 'c' {
		t.Errorf("[2] = %+v", els[2])
	}
}
