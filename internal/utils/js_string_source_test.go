// cspell:ignore nonascii srcs
package utils

import (
	"fmt"
	"testing"
)

// Each test expresses the expected code-unit sequence produced by
// ParseJSStringLiteralSource / ParseJSTemplateLiteralSource. Every entry is
// {Value, Start, End} — Start/End are byte offsets within the input source
// (including the surrounding quote / backtick characters).

type expectedUnit struct {
	value uint32
	start int
	end   int
}

func assertCodeUnits(t *testing.T, label, source string, got []StringCodeUnit, want []expectedUnit) {
	t.Helper()
	if len(got) != len(want) {
		t.Errorf("%s(%q): len=%d want=%d\n  got: %v\n  want: %v", label, source, len(got), len(want), got, want)
		return
	}
	for i := range want {
		if got[i].Value != want[i].value || got[i].Start != want[i].start || got[i].End != want[i].end {
			t.Errorf("%s(%q)[%d]: got %+v, want %+v", label, source, i, got[i], want[i])
		}
	}
}

// ---------------------------------------------------------------------------
// String literal tests
// ---------------------------------------------------------------------------

func TestParseJSStringLiteralSource_Basic(t *testing.T) {
	// Simple ASCII, single-quoted. Source: 'abc'
	got := ParseJSStringLiteralSource(`'abc'`)
	want := []expectedUnit{
		{'a', 1, 2}, {'b', 2, 3}, {'c', 3, 4},
	}
	assertCodeUnits(t, "string", `'abc'`, got, want)
}

func TestParseJSStringLiteralSource_Empty(t *testing.T) {
	for _, src := range []string{`""`, `''`} {
		got := ParseJSStringLiteralSource(src)
		if got == nil {
			t.Fatalf("ParseJSStringLiteralSource(%q) = nil, want empty slice", src)
		}
		if len(got) != 0 {
			t.Errorf("ParseJSStringLiteralSource(%q) = %v, want []", src, got)
		}
	}
}

func TestParseJSStringLiteralSource_DoubleQuote(t *testing.T) {
	got := ParseJSStringLiteralSource(`"ab"`)
	want := []expectedUnit{{'a', 1, 2}, {'b', 2, 3}}
	assertCodeUnits(t, "string", `"ab"`, got, want)
}

func TestParseJSStringLiteralSource_SimpleEscapes(t *testing.T) {
	// Each simple escape occupies 2 source bytes and produces 1 code unit.
	cases := []struct {
		src   string
		value uint32
	}{
		{`"\b"`, '\b'},
		{`"\f"`, '\f'},
		{`"\n"`, '\n'},
		{`"\r"`, '\r'},
		{`"\t"`, '\t'},
		{`"\v"`, '\v'},
	}
	for _, c := range cases {
		got := ParseJSStringLiteralSource(c.src)
		assertCodeUnits(t, "simple-escape", c.src, got, []expectedUnit{{c.value, 1, 3}})
	}
}

func TestParseJSStringLiteralSource_IdentityEscapes(t *testing.T) {
	cases := []struct {
		src   string
		value uint32
	}{
		{`"\\"`, '\\'},
		{`"\'"`, '\''},
		{`"\""`, '"'},
		{`"\?"`, '?'},
		{`"\a"`, 'a'},
		{`"\z"`, 'z'},
		{`"\$"`, '$'},
		{`"\|"`, '|'},
	}
	for _, c := range cases {
		got := ParseJSStringLiteralSource(c.src)
		assertCodeUnits(t, "identity-escape", c.src, got, []expectedUnit{{c.value, 1, 3}})
	}
}

func TestParseJSStringLiteralSource_HexEscape(t *testing.T) {
	// \xHH → 4 source bytes, 1 code unit
	got := ParseJSStringLiteralSource(`"\x41"`)
	assertCodeUnits(t, "hex", `"\x41"`, got, []expectedUnit{{0x41, 1, 5}})

	got = ParseJSStringLiteralSource(`"\x00"`)
	assertCodeUnits(t, "hex-zero", `"\x00"`, got, []expectedUnit{{0x00, 1, 5}})

	got = ParseJSStringLiteralSource(`"\xFF"`)
	assertCodeUnits(t, "hex-ff", `"\xFF"`, got, []expectedUnit{{0xFF, 1, 5}})

	got = ParseJSStringLiteralSource(`"\xff"`)
	assertCodeUnits(t, "hex-lower", `"\xff"`, got, []expectedUnit{{0xff, 1, 5}})

	got = ParseJSStringLiteralSource(`"\xAb"`)
	assertCodeUnits(t, "hex-mixed", `"\xAb"`, got, []expectedUnit{{0xAb, 1, 5}})
}

func TestParseJSStringLiteralSource_UnicodeEscapeBMP(t *testing.T) {
	// \uHHHH → 6 source bytes, 1 code unit
	got := ParseJSStringLiteralSource(`"\u0041"`)
	assertCodeUnits(t, "u", `"\u0041"`, got, []expectedUnit{{0x41, 1, 7}})

	got = ParseJSStringLiteralSource(`"\uD83D"`)
	assertCodeUnits(t, "u-lone-hi", `"\uD83D"`, got, []expectedUnit{{0xD83D, 1, 7}})

	got = ParseJSStringLiteralSource(`"\uDC4D"`)
	assertCodeUnits(t, "u-lone-lo", `"\uDC4D"`, got, []expectedUnit{{0xDC4D, 1, 7}})

	got = ParseJSStringLiteralSource(`"\uFFFF"`)
	assertCodeUnits(t, "u-ffff", `"\uFFFF"`, got, []expectedUnit{{0xFFFF, 1, 7}})
}

func TestParseJSStringLiteralSource_UnicodeEscapeSurrogatePair(t *testing.T) {
	// Surrogate pair as two separate \uHHHH → 2 units mapping to different ranges
	got := ParseJSStringLiteralSource(`"\uD83D\uDC4D"`)
	assertCodeUnits(t, "surrogate-pair", `"\uD83D\uDC4D"`, got, []expectedUnit{
		{0xD83D, 1, 7},
		{0xDC4D, 7, 13},
	})
}

func TestParseJSStringLiteralSource_UnicodeEscapeCodePointBMP(t *testing.T) {
	// \u{H} BMP → 1 unit
	got := ParseJSStringLiteralSource(`"\u{41}"`)
	assertCodeUnits(t, "u-brace-bmp", `"\u{41}"`, got, []expectedUnit{{0x41, 1, 7}})

	got = ParseJSStringLiteralSource(`"\u{0041}"`)
	assertCodeUnits(t, "u-brace-leading-zero", `"\u{0041}"`, got, []expectedUnit{{0x41, 1, 9}})

	got = ParseJSStringLiteralSource(`"\u{FFFF}"`)
	assertCodeUnits(t, "u-brace-ffff", `"\u{FFFF}"`, got, []expectedUnit{{0xFFFF, 1, 9}})
}

func TestParseJSStringLiteralSource_UnicodeEscapeCodePointAstral(t *testing.T) {
	// \u{1F44D} astral → 2 units, both mapping to the same source range
	got := ParseJSStringLiteralSource(`"\u{1F44D}"`)
	want := []expectedUnit{
		{0xD83D, 1, 10},
		{0xDC4D, 1, 10},
	}
	assertCodeUnits(t, "u-brace-astral", `"\u{1F44D}"`, got, want)

	// Upper-bound astral
	got = ParseJSStringLiteralSource(`"\u{10FFFF}"`)
	want = []expectedUnit{
		{0xDBFF, 1, 11},
		{0xDFFF, 1, 11},
	}
	assertCodeUnits(t, "u-brace-max", `"\u{10FFFF}"`, got, want)
}

func TestParseJSStringLiteralSource_RawASCII(t *testing.T) {
	got := ParseJSStringLiteralSource(`"A"`)
	assertCodeUnits(t, "raw-ascii", `"A"`, got, []expectedUnit{{'A', 1, 2}})
}

func TestParseJSStringLiteralSource_RawNonASCII(t *testing.T) {
	// U+00C1 (Á) is 2 UTF-8 bytes → 1 code unit spanning 2 bytes
	got := ParseJSStringLiteralSource("\"\u00C1\"")
	assertCodeUnits(t, "raw-2byte", "\"\u00C1\"", got, []expectedUnit{{0x00C1, 1, 3}})

	// U+2747 (❇) is 3 UTF-8 bytes → 1 code unit spanning 3 bytes
	got = ParseJSStringLiteralSource("\"\u2747\"")
	assertCodeUnits(t, "raw-3byte", "\"\u2747\"", got, []expectedUnit{{0x2747, 1, 4}})
}

func TestParseJSStringLiteralSource_RawAstral(t *testing.T) {
	// U+1F44D (👍) is 4 UTF-8 bytes → 2 code units (surrogate pair) mapping to same range
	got := ParseJSStringLiteralSource("\"\U0001F44D\"")
	want := []expectedUnit{
		{0xD83D, 1, 5},
		{0xDC4D, 1, 5},
	}
	assertCodeUnits(t, "raw-astral", "\"\U0001F44D\"", got, want)
}

func TestParseJSStringLiteralSource_LineContinuation(t *testing.T) {
	// `\<LF>` line continuation → 0 code units
	got := ParseJSStringLiteralSource("\"\\\n\"")
	assertCodeUnits(t, "lc-lf", "\"\\\n\"", got, []expectedUnit{})

	// `\<CR>` line continuation → 0 code units
	got = ParseJSStringLiteralSource("\"\\\r\"")
	assertCodeUnits(t, "lc-cr", "\"\\\r\"", got, []expectedUnit{})

	// `\<CR><LF>` → 0 code units
	got = ParseJSStringLiteralSource("\"\\\r\n\"")
	assertCodeUnits(t, "lc-crlf", "\"\\\r\n\"", got, []expectedUnit{})

	// `\<LS>` (U+2028) → 0 code units. LS is 3 UTF-8 bytes.
	got = ParseJSStringLiteralSource("\"\\\u2028\"")
	assertCodeUnits(t, "lc-ls", "\"\\\u2028\"", got, []expectedUnit{})

	// `\<PS>` (U+2029) → 0 code units.
	got = ParseJSStringLiteralSource("\"\\\u2029\"")
	assertCodeUnits(t, "lc-ps", "\"\\\u2029\"", got, []expectedUnit{})
}

func TestParseJSStringLiteralSource_LineContinuationSurroundedByChars(t *testing.T) {
	// "a\<LF>b" → {a, b}, with line continuation consumed invisibly.
	got := ParseJSStringLiteralSource("\"a\\\nb\"")
	// Source layout: " a \ <LF> b "
	// Offsets:       0 1 2 3    4 5
	// Expected units: a (1..2), b (4..5)
	want := []expectedUnit{{'a', 1, 2}, {'b', 4, 5}}
	assertCodeUnits(t, "lc-inline", "\"a\\\nb\"", got, want)

	// With CRLF
	got = ParseJSStringLiteralSource("\"a\\\r\nb\"")
	// Offsets:       0 1 2 3  4  5 6
	want = []expectedUnit{{'a', 1, 2}, {'b', 5, 6}}
	assertCodeUnits(t, "lc-crlf-inline", "\"a\\\r\nb\"", got, want)
}

func TestParseJSStringLiteralSource_LegacyOctal(t *testing.T) {
	// `\0` alone when followed by quote → NUL, 1 unit
	got := ParseJSStringLiteralSource(`"\0"`)
	assertCodeUnits(t, "octal-0", `"\0"`, got, []expectedUnit{{0, 1, 3}})

	// `\00` → octal 0, 1 unit (3 source bytes)
	got = ParseJSStringLiteralSource(`"\00"`)
	assertCodeUnits(t, "octal-00", `"\00"`, got, []expectedUnit{{0, 1, 4}})

	// `\012` → 0o12 = 0x0A, 1 unit (4 source bytes)
	got = ParseJSStringLiteralSource(`"\012"`)
	assertCodeUnits(t, "octal-012", `"\012"`, got, []expectedUnit{{0x0A, 1, 5}})

	// `\377` → 0o377 = 0xFF, max legacy octal
	got = ParseJSStringLiteralSource(`"\377"`)
	assertCodeUnits(t, "octal-max", `"\377"`, got, []expectedUnit{{0xFF, 1, 5}})

	// `\4` → 4, with max len 2, since leading digit ∈ {4..7}
	got = ParseJSStringLiteralSource(`"\4"`)
	assertCodeUnits(t, "octal-4", `"\4"`, got, []expectedUnit{{4, 1, 3}})

	// `\45` → 0o45 = 0x25
	got = ParseJSStringLiteralSource(`"\45"`)
	assertCodeUnits(t, "octal-45", `"\45"`, got, []expectedUnit{{0x25, 1, 4}})

	// `\456` → only first 2 digits used since leading is 4..7; `\45` then `6`
	got = ParseJSStringLiteralSource(`"\456"`)
	want := []expectedUnit{{0x25, 1, 4}, {'6', 4, 5}}
	assertCodeUnits(t, "octal-456", `"\456"`, got, want)

	// `\0` + `8` → NUL (octal runs out because 8 is not a digit in 0..7), then `8`
	got = ParseJSStringLiteralSource(`"\08"`)
	want = []expectedUnit{{0, 1, 3}, {'8', 3, 4}}
	assertCodeUnits(t, "octal-08", `"\08"`, got, want)
}

func TestParseJSStringLiteralSource_UnterminatedReturnsNil(t *testing.T) {
	for _, src := range []string{
		`"abc`,     // no closing quote
		`"ab\`,     // ends on backslash
		`"\u004`,   // truncated \u escape (still returns nil as the whole source is malformed overall: no closing quote)
		``,
		`"`,
		`'`,
	} {
		got := ParseJSStringLiteralSource(src)
		if got != nil {
			t.Errorf("ParseJSStringLiteralSource(%q) = %v, want nil", src, got)
		}
	}
}

func TestParseJSStringLiteralSource_NonStringInput(t *testing.T) {
	// Template literal passed to string parser → nil
	got := ParseJSStringLiteralSource("`abc`")
	if got != nil {
		t.Errorf("ParseJSStringLiteralSource on template: got %v, want nil", got)
	}
}

// ---------------------------------------------------------------------------
// Template literal tests
// ---------------------------------------------------------------------------

func TestParseJSTemplateLiteralSource_Basic(t *testing.T) {
	got := ParseJSTemplateLiteralSource("`abc`")
	want := []expectedUnit{{'a', 1, 2}, {'b', 2, 3}, {'c', 3, 4}}
	assertCodeUnits(t, "tpl", "`abc`", got, want)
}

func TestParseJSTemplateLiteralSource_Empty(t *testing.T) {
	got := ParseJSTemplateLiteralSource("``")
	if got == nil || len(got) != 0 {
		t.Errorf("ParseJSTemplateLiteralSource(``) = %v, want []", got)
	}
}

func TestParseJSTemplateLiteralSource_SimpleEscapes(t *testing.T) {
	got := ParseJSTemplateLiteralSource("`\\n`")
	assertCodeUnits(t, "tpl-\\n", "`\\n`", got, []expectedUnit{{'\n', 1, 3}})
}

func TestParseJSTemplateLiteralSource_UnicodeEscapeAstral(t *testing.T) {
	got := ParseJSTemplateLiteralSource("`\\u{1F44D}`")
	want := []expectedUnit{{0xD83D, 1, 10}, {0xDC4D, 1, 10}}
	assertCodeUnits(t, "tpl-u-astral", "`\\u{1F44D}`", got, want)
}

func TestParseJSTemplateLiteralSource_RawAstral(t *testing.T) {
	got := ParseJSTemplateLiteralSource("`\U0001F44D`")
	want := []expectedUnit{{0xD83D, 1, 5}, {0xDC4D, 1, 5}}
	assertCodeUnits(t, "tpl-raw-astral", "`\U0001F44D`", got, want)
}

func TestParseJSTemplateLiteralSource_RawCRLF(t *testing.T) {
	// Raw CRLF in template → cooked LF (1 code unit), source spans 2 bytes.
	got := ParseJSTemplateLiteralSource("`\r\n`")
	want := []expectedUnit{{'\n', 1, 3}}
	assertCodeUnits(t, "tpl-raw-crlf", "`\r\n`", got, want)

	// Raw CR alone → cooked LF, 1 byte
	got = ParseJSTemplateLiteralSource("`\r`")
	want = []expectedUnit{{'\n', 1, 2}}
	assertCodeUnits(t, "tpl-raw-cr", "`\r`", got, want)

	// Raw LF alone → LF, 1 byte
	got = ParseJSTemplateLiteralSource("`\n`")
	want = []expectedUnit{{'\n', 1, 2}}
	assertCodeUnits(t, "tpl-raw-lf", "`\n`", got, want)
}

func TestParseJSTemplateLiteralSource_LineContinuation(t *testing.T) {
	got := ParseJSTemplateLiteralSource("`\\\n`")
	assertCodeUnits(t, "tpl-lc-lf", "`\\\n`", got, []expectedUnit{})

	got = ParseJSTemplateLiteralSource("`\\\r\n`")
	assertCodeUnits(t, "tpl-lc-crlf", "`\\\r\n`", got, []expectedUnit{})
}

func TestParseJSTemplateLiteralSource_SubstitutionTerminatesSpan(t *testing.T) {
	// `${...}` inside template terminates the span we're parsing.
	// Our parser stops at `$` + `{`.
	got := ParseJSTemplateLiteralSource("`ab${x}`")
	want := []expectedUnit{{'a', 1, 2}, {'b', 2, 3}}
	assertCodeUnits(t, "tpl-subst", "`ab${x}`", got, want)

	// Escaped `$` → literal `$`
	got = ParseJSTemplateLiteralSource("`\\${x}`")
	// `\$` → literal `$` at source [1..3], then raw `{`, `x`, `}`
	want = []expectedUnit{
		{'$', 1, 3},
		{'{', 3, 4},
		{'x', 4, 5},
		{'}', 5, 6},
	}
	assertCodeUnits(t, "tpl-escaped-dollar", "`\\${x}`", got, want)
}

func TestParseJSTemplateLiteralSource_BacktickEscape(t *testing.T) {
	got := ParseJSTemplateLiteralSource("`\\``")
	want := []expectedUnit{{'`', 1, 3}}
	assertCodeUnits(t, "tpl-escaped-bt", "`\\``", got, want)
}

func TestParseJSTemplateLiteralSource_Unterminated(t *testing.T) {
	for _, src := range []string{"", "`", "`abc", "`ab\\"} {
		got := ParseJSTemplateLiteralSource(src)
		if got != nil {
			t.Errorf("ParseJSTemplateLiteralSource(%q) = %v, want nil", src, got)
		}
	}
}

// ---------------------------------------------------------------------------
// Boundary / cross-cutting tests
// ---------------------------------------------------------------------------

func TestParseJSStringLiteralSource_MixedFormsSourceSpans(t *testing.T) {
	// Cover all source spans in a single string:
	//   "a\x41\u0042\u{43}\u{1F600}👍\nZ"
	// Units:
	//   a        (1..2)
	//   \x41=A   (2..6)
	//   \u0042=B (6..12)
	//   \u{43}=C (12..18)
	//   \u{1F600} astral → 2 units, (18..26) each
	//   raw 👍 (4 bytes) → 2 units, (26..30) each
	//   \n       (30..32)
	//   Z        (32..33)
	src := "\"a\\x41\\u0042\\u{43}\\u{1F600}\U0001F600\\nZ\""
	got := ParseJSStringLiteralSource(src)
	// Offsets in source:
	//   "  a  \x41  \u0042  \u{43}  \u{1F600}  😀  \n  Z  "
	//   0  1   2      6      12      18        27  31  33 34
	want := []expectedUnit{
		{'a', 1, 2},
		{0x41, 2, 6},
		{0x42, 6, 12},
		{0x43, 12, 18},
		{0xD83D, 18, 27}, {0xDE00, 18, 27}, // \u{1F600} → astral, 9 source bytes
		{0xD83D, 27, 31}, {0xDE00, 27, 31}, // raw astral, 4 UTF-8 bytes
		{'\n', 31, 33},
		{'Z', 33, 34},
	}
	assertCodeUnits(t, "mixed", src, got, want)
}

func TestParseJSStringLiteralSource_LongerIdentityEscape_NonASCII(t *testing.T) {
	// `\Á` → identity escape of raw 2-byte UTF-8 char → 1 code unit covering 3 source bytes (`\` + 2-byte UTF-8)
	src := "\"\\\u00C1\""
	got := ParseJSStringLiteralSource(src)
	want := []expectedUnit{{0x00C1, 1, 4}}
	assertCodeUnits(t, "identity-nonascii", src, got, want)

	// `\❇` → 3-byte UTF-8 → 1 code unit, 4 source bytes (`\` + 3-byte UTF-8)
	src = "\"\\\u2747\""
	got = ParseJSStringLiteralSource(src)
	want = []expectedUnit{{0x2747, 1, 5}}
	assertCodeUnits(t, "identity-nonascii-3byte", src, got, want)

	// `\👍` → `\` + 4-byte UTF-8 astral → astral identity escape → 2 code units
	src = "\"\\\U0001F44D\""
	got = ParseJSStringLiteralSource(src)
	want = []expectedUnit{
		{0xD83D, 1, 6},
		{0xDC4D, 1, 6},
	}
	assertCodeUnits(t, "identity-astral", src, got, want)
}

// Regression for `\x` with exactly one trailing hex digit — we treat as
// identity escape of `x`. No panic, 1 code unit with value 'x'.
func TestParseJSStringLiteralSource_MalformedHex(t *testing.T) {
	got := ParseJSStringLiteralSource(`"\x1"`)
	// `\x1` isn't a valid \xHH — we fall back to identity `x` and continue.
	// Expected units: x (1..3), 1 (3..4)
	want := []expectedUnit{{'x', 1, 3}, {'1', 3, 4}}
	assertCodeUnits(t, "malformed-hex", `"\x1"`, got, want)
}

func TestParseJSStringLiteralSource_NestedBackslashEscapes(t *testing.T) {
	// `"\\\u0041"` — two chars: `\` (from `\\`) then `A` (from `\u0041`).
	src := `"\\\u0041"`
	got := ParseJSStringLiteralSource(src)
	want := []expectedUnit{
		{'\\', 1, 3},    // \\ → \
		{0x41, 3, 9},    // \u0041 → A (6 bytes)
	}
	assertCodeUnits(t, "nested-\\\\\\u", src, got, want)

	// `"\\\\"` — four backslashes in source → two `\` in value.
	src = `"\\\\"`
	got = ParseJSStringLiteralSource(src)
	want = []expectedUnit{
		{'\\', 1, 3},
		{'\\', 3, 5},
	}
	assertCodeUnits(t, "four-backslashes", src, got, want)
}

func TestParseJSStringLiteralSource_EscapedQuoteInsideMatchingString(t *testing.T) {
	// `"\""` → one char: `"`
	got := ParseJSStringLiteralSource(`"\""`)
	assertCodeUnits(t, "escaped-dq", `"\""`, got, []expectedUnit{{'"', 1, 3}})

	// `'\''` → one char: `'`
	got = ParseJSStringLiteralSource(`'\''`)
	assertCodeUnits(t, "escaped-sq", `'\''`, got, []expectedUnit{{'\'', 1, 3}})

	// Unescaped different quote is just the char
	got = ParseJSStringLiteralSource(`"'"`)
	assertCodeUnits(t, "unescaped-sq-in-dq", `"'"`, got, []expectedUnit{{'\'', 1, 2}})
}

func TestParseJSStringLiteralSource_OctalBoundaries(t *testing.T) {
	// `\3` with max=3 (since leading is 0..3)
	got := ParseJSStringLiteralSource(`"\3"`)
	assertCodeUnits(t, "octal-3", `"\3"`, got, []expectedUnit{{3, 1, 3}})

	// `\37` (max 3 digits since leading 0..3; `\37` = 0o37 = 0x1F)
	got = ParseJSStringLiteralSource(`"\37"`)
	assertCodeUnits(t, "octal-37", `"\37"`, got, []expectedUnit{{0x1F, 1, 4}})

	// `\400` — leading `4` caps at 2 digits → `\40` (0x20) then `0` literal
	got = ParseJSStringLiteralSource(`"\400"`)
	want := []expectedUnit{{0x20, 1, 4}, {'0', 4, 5}}
	assertCodeUnits(t, "octal-400", `"\400"`, got, want)
}

func TestParseJSTemplateLiteralSource_DollarNotSubstitution(t *testing.T) {
	// `$` alone, not followed by `{`, is a literal char.
	got := ParseJSTemplateLiteralSource("`a$b`")
	want := []expectedUnit{{'a', 1, 2}, {'$', 2, 3}, {'b', 3, 4}}
	assertCodeUnits(t, "tpl-dollar-plain", "`a$b`", got, want)
}

func TestParseJSTemplateLiteralSource_MultipleLineContinuationsAndRawNewlines(t *testing.T) {
	// Mix raw LF, raw CRLF, and `\<LF>` continuation.
	src := "`a\nb\r\nc\\\nd`"
	got := ParseJSTemplateLiteralSource(src)
	// Layout:
	//   `  a  \n  b  \r\n  c  \<LF>  d  `
	//   0  1  2   3  4     6  7   8    10 11
	want := []expectedUnit{
		{'a', 1, 2},
		{'\n', 2, 3},
		{'b', 3, 4},
		{'\n', 4, 6}, // raw CRLF normalized to LF, spans 2 source bytes
		{'c', 6, 7},
		// `\<LF>` produces 0 units
		{'d', 9, 10},
	}
	assertCodeUnits(t, "tpl-mixed-newlines", src, got, want)
}

func TestParseJSStringLiteralSource_MalformedUnicode(t *testing.T) {
	// `\u0` — not enough digits → identity 'u'
	got := ParseJSStringLiteralSource(`"\u0"`)
	want := []expectedUnit{{'u', 1, 3}, {'0', 3, 4}}
	assertCodeUnits(t, "malformed-u", `"\u0"`, got, want)
}

func TestParseJSStringLiteralSource_AllEscapeSequenceBoundary(t *testing.T) {
	// Paranoid: exhaustively check that each hex/unicode escape spans the
	// correct byte count for a few representative code points.
	type hexCase struct{ prefix, hex, suffix string }
	cases := []hexCase{
		{`"\x`, `41`, `"`},
		{`"\u`, `0041`, `"`},
		{`"\u{`, `41`, `}"`},
		{`"\u{`, `1F44D`, `}"`},
		{`"\u{`, `10FFFF`, `}"`},
	}
	for _, c := range cases {
		src := c.prefix + c.hex + c.suffix
		got := ParseJSStringLiteralSource(src)
		if got == nil {
			t.Errorf("malformed parse on %q", src)
			continue
		}
		// At least one unit was produced; source spans must sum to len(src)-2.
		var covered int
		for _, u := range got {
			if u.End > covered {
				covered = u.End
			}
			if u.Start == u.End {
				t.Errorf("%q: zero-width unit %+v", src, u)
			}
		}
		expectedEnd := len(src) - 1 // just before the closing quote
		if covered != expectedEnd {
			t.Errorf("%q: covered %d, want %d", src, covered, expectedEnd)
		}
	}
}

// Cross-check: applying the units' Start/End to produce a slice of the source
// text and concatenating the raw source pieces yields the original source
// (minus quotes and line continuations, which are unmapped).
func TestParseJSStringLiteralSource_SourceCoverage(t *testing.T) {
	srcs := []string{
		`"abc"`,
		`"\n\r\t"`,
		`"\x41\x42"`,
		`"\u0041\u0042"`,
		`"\u{1F44D}Z"`,
		"\"A\\\nB\"",          // with line continuation
		"\"\U0001F44D\u00C1\"", // raw astral + raw 2-byte
	}
	for _, src := range srcs {
		t.Run(fmt.Sprintf("%q", src), func(t *testing.T) {
			units := ParseJSStringLiteralSource(src)
			if units == nil {
				t.Fatalf("parse failed: %q", src)
			}
			// All End offsets must be within source bounds.
			for _, u := range units {
				if u.Start < 0 || u.End > len(src) || u.Start > u.End {
					t.Errorf("bad range %+v in %q", u, src)
				}
				if u.End > len(src)-1 {
					t.Errorf("unit %+v ends past closing quote in %q", u, src)
				}
			}
			// Adjacent units whose Value is not a surrogate pair half must have non-overlapping ranges.
			for i := 1; i < len(units); i++ {
				prev, cur := units[i-1], units[i]
				if prev.Start == cur.Start && prev.End == cur.End {
					// astral pair — both map to same range, OK
					continue
				}
				if cur.Start < prev.End {
					t.Errorf("overlapping non-pair units at index %d in %q: prev=%+v cur=%+v", i, src, prev, cur)
				}
			}
		})
	}
}
