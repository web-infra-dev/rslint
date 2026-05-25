package no_octal_escape

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestFindOctalEscape(t *testing.T) {
	tests := []struct {
		name     string
		raw      string
		expected string
	}{
		// ================================================================
		// Empty / minimal strings
		// ================================================================
		{name: "empty single-quoted", raw: `''`, expected: ""},
		{name: "empty double-quoted", raw: `""`, expected: ""},
		{name: "single char", raw: `'a'`, expected: ""},

		// ================================================================
		// Hex escapes — \xNN is not octal
		// ================================================================
		{name: "hex escape \\x51", raw: `"\x51"`, expected: ""},
		{name: "hex escape \\x01", raw: `'\x01'`, expected: ""},
		{name: "hex escape \\xFF", raw: `'\xFF'`, expected: ""},

		// ================================================================
		// Unicode escapes — \uNNNN, \u{N} are not octal
		// ================================================================
		{name: "unicode escape \\u0041", raw: `'\u0041'`, expected: ""},
		{name: "unicode escape \\u0001", raw: `'\u0001'`, expected: ""},
		{name: "unicode escape braces \\u{41}", raw: `'\u{41}'`, expected: ""},
		{name: "unicode escape braces \\u{1F600}", raw: `'\u{1F600}'`, expected: ""},

		// ================================================================
		// Hex/unicode escape followed by octal — scanner must not confuse
		// the digits after \x/\u with octal sequences
		// ================================================================
		{name: "hex then octal \\x41\\1", raw: `'\x41\1'`, expected: "1"},
		{name: "unicode then octal \\u0041\\1", raw: `'\u0041\1'`, expected: "1"},
		{name: "unicode braces then octal", raw: `'\u{41}\1'`, expected: "1"},
		{name: "hex then two-digit octal", raw: `'\x41\01'`, expected: "01"},
		{name: "incomplete hex then octal \\x\\1", raw: `'\x\1'`, expected: "1"},
		{name: "incomplete hex digit then octal \\x1\\1", raw: `'\x1\1'`, expected: "1"},

		// ================================================================
		// \0 (NULL) — valid when not followed by a digit
		// ================================================================
		{name: "null alone", raw: `'\0'`, expected: ""},
		{name: "null space", raw: `'\0 '`, expected: ""},
		{name: "space null", raw: `' \0'`, expected: ""},
		{name: "a null", raw: `'a\0'`, expected: ""},
		{name: "null a", raw: `'\0a'`, expected: ""},
		{name: "null at end before quote", raw: `"x\0"`, expected: ""},

		// ================================================================
		// \0 chained — multiple NULLs without triggering octal
		// ================================================================
		{name: "null null", raw: `'\0\0'`, expected: ""},
		{name: "null null null", raw: `'\0\0\0'`, expected: ""},
		{name: "null a null", raw: `'\0a\0'`, expected: ""},
		{name: "null escape-backslash null", raw: `'\0\\\0'`, expected: ""},

		// ================================================================
		// \0 chained — then octal deeper in the string
		// ================================================================
		{name: "null null then \\1", raw: `'\0\0\1'`, expected: "1"},
		{name: "null a null b \\1", raw: `'\0a\0b\1'`, expected: "1"},
		{name: "null escape-backslash then \\1", raw: `'\0\\\1'`, expected: "1"},
		{name: "null null then \\01", raw: `'\0\0\01'`, expected: "01"},

		// ================================================================
		// \0 followed by octal digit → octal escape
		// ================================================================
		{name: "\\00", raw: `'\00'`, expected: "00"},
		{name: "\\01", raw: `'\01'`, expected: "01"},
		{name: "\\02", raw: `'\02'`, expected: "02"},
		{name: "\\07", raw: `'\07'`, expected: "07"},
		{name: "null then \\01", raw: `'\0\01'`, expected: "01"},
		{name: "null space \\01", raw: `'\0 \01'`, expected: "01"},
		{name: "null a \\01", raw: `'\0a\01'`, expected: "01"},
		{name: "null then \\00", raw: `'\0\00'`, expected: "00"},

		// ================================================================
		// \0 followed by 8 or 9 → octal "0"
		// ================================================================
		{name: "\\08", raw: `'\08'`, expected: "0"},
		{name: "\\09", raw: `'\09'`, expected: "0"},
		{name: "\\089", raw: `'\089'`, expected: "0"},
		{name: "null then \\08", raw: `'\0\08'`, expected: "0"},
		{name: "null then \\09", raw: `'\0\09'`, expected: "0"},
		{name: "a \\08 a", raw: `'a\08a'`, expected: "0"},

		// ================================================================
		// \8 and \9 — NOT octal (non-octal digits after \)
		// ================================================================
		{name: "\\8 alone", raw: `'\8'`, expected: ""},
		{name: "\\9 alone", raw: `'\9'`, expected: ""},
		{name: "\\80", raw: `'\80'`, expected: ""},
		{name: "\\81", raw: `'\81'`, expected: ""},
		{name: "\\8 \\0", raw: `'\8\0'`, expected: ""},
		{name: "\\0 \\8", raw: `'\0\8'`, expected: ""},
		{name: "a \\8 a", raw: `'a\8a'`, expected: ""},

		// ================================================================
		// Single-digit octal: \1 through \7
		// ================================================================
		{name: "\\1", raw: `'\1'`, expected: "1"},
		{name: "\\2", raw: `'\2'`, expected: "2"},
		{name: "\\3", raw: `'\3'`, expected: "3"},
		{name: "\\4", raw: `'\4'`, expected: "4"},
		{name: "\\5", raw: `'\5'`, expected: "5"},
		{name: "\\6", raw: `'\6'`, expected: "6"},
		{name: "\\7", raw: `'\7'`, expected: "7"},

		// ================================================================
		// Two-digit octal
		// ================================================================
		{name: "\\10", raw: `'\10'`, expected: "10"},
		{name: "\\12", raw: `'\12'`, expected: "12"},
		{name: "\\37", raw: `'\37'`, expected: "37"},
		{name: "\\40", raw: `'\40'`, expected: "40"},
		{name: "\\77", raw: `'\77'`, expected: "77"},

		// ================================================================
		// Three-digit octal (first digit 0-3 allows 3 digits total)
		// ================================================================
		{name: "\\000", raw: `'\000'`, expected: "000"},
		{name: "\\001", raw: `'\001'`, expected: "001"},
		{name: "\\377", raw: `'\377'`, expected: "377"},
		{name: "\\251", raw: `'\251'`, expected: "251"},

		// ================================================================
		// Octal sequence length boundary — non-octal char stops the sequence
		// ================================================================
		{name: "\\378 → 37 (8 not octal)", raw: `"\378"`, expected: "37"},
		{name: "\\37a → 37 (a not octal)", raw: `"\37a"`, expected: "37"},
		{name: "\\381 → 3 (8 not octal)", raw: `"\381"`, expected: "3"},
		{name: "\\3a1 → 3 (a not octal)", raw: `"\3a1"`, expected: "3"},
		{name: "\\258 → 25 (8 not octal)", raw: `"\258"`, expected: "25"},
		{name: "\\25a → 25 (a not octal)", raw: `"\25a"`, expected: "25"},
		{name: "\\3s51 → 3", raw: `"\3s51"`, expected: "3"},
		{name: "\\78 → 7 (4-7 range, max 2 digits)", raw: `"\78"`, expected: "7"},
		{name: "\\5a → 5", raw: `"\5a"`, expected: "5"},

		// ================================================================
		// Octal sequence length boundary — max digits consumed
		// First digit 0-3: up to 3 digits; 4-7: up to 2 digits
		// ================================================================
		{name: "\\0377 → 037 (3 digits max for 0-3)", raw: `'\0377'`, expected: "037"},
		{name: "\\0000 → 000 (3 digits max)", raw: `'\0000'`, expected: "000"},
		{name: "\\777 → 77 (2 digits max for 4-7)", raw: `'\777'`, expected: "77"},
		{name: "\\4777 → 47 (2 digits max)", raw: `'\4777'`, expected: "47"},
		{name: "\\751 → 75 (2 digits max for 7)", raw: `"\751"`, expected: "75"},
		{name: "\\400 → 40 (2 digits max for 4)", raw: `"\400"`, expected: "40"},

		// ================================================================
		// Escaped backslash (\\) — two chars consumed as pair, NOT octal
		// ================================================================
		{name: "escaped backslash alone", raw: `'\\'`, expected: ""},
		{name: "escaped backslash 0", raw: `'\\0'`, expected: ""},
		{name: "escaped backslash 1", raw: `'\\1'`, expected: ""},
		{name: "escaped backslash 01", raw: `'\\01'`, expected: ""},
		{name: "escaped backslash 08", raw: `'\\08'`, expected: ""},
		{name: "escaped backslash 12", raw: `'\\12'`, expected: ""},
		{name: "escaped backslash null", raw: `'\\\0'`, expected: ""},
		{name: "escaped backslash \\8", raw: `'\\\8'`, expected: ""},
		{name: "null then escaped backslash", raw: `'\0\\'`, expected: ""},

		// ================================================================
		// Backslash parity chains — even count = all paired (no octal),
		// odd count = last one starts a new escape
		// ================================================================
		// 2 backslashes (1 pair) + digit → escaped backslash then plain digit
		{name: "2 backslashes + 1", raw: `'\\1'`, expected: ""},
		// 3 backslashes (1 pair + lone \\) + digit → octal
		{name: "3 backslashes + 1", raw: `'\\\1'`, expected: "1"},
		// 4 backslashes (2 pairs) + digit → no octal
		{name: "4 backslashes + 1", raw: `'\\\\1'`, expected: ""},
		// 5 backslashes (2 pairs + lone \\) + digit → octal
		{name: "5 backslashes + 1", raw: `'\\\\\1'`, expected: "1"},
		// 6 backslashes (3 pairs) + digit → no octal
		{name: "6 backslashes + 1", raw: `'\\\\\\1'`, expected: ""},
		// Same pattern with \0 instead of \1
		{name: "2 backslashes + 0 + 1", raw: `'\\01'`, expected: ""},
		{name: "3 backslashes + 0 + 1", raw: `'\\\01'`, expected: "01"},
		{name: "4 backslashes + 0 + 1", raw: `'\\\\01'`, expected: ""},
		{name: "5 backslashes + 0 + 1", raw: `'\\\\\01'`, expected: "01"},
		// With \08 (octal "0")
		{name: "3 backslashes + 08", raw: `'\\\08'`, expected: "0"},
		{name: "4 backslashes + 08", raw: `'\\\\08'`, expected: ""},
		{name: "5 backslashes + 08", raw: `'\\\\\08'`, expected: "0"},

		// ================================================================
		// Standard escape sequences before octal — all skip correctly
		// ================================================================
		{name: "\\n then \\1", raw: `'\n\1'`, expected: "1"},
		{name: "\\t then \\1", raw: `'\t\1'`, expected: "1"},
		{name: "\\r then \\1", raw: `'\r\1'`, expected: "1"},
		{name: "\\v then \\1", raw: `'\v\1'`, expected: "1"},
		{name: "\\f then \\1", raw: `'\f\1'`, expected: "1"},
		{name: "\\b then \\1", raw: `'\b\1'`, expected: "1"},
		{name: "\\n then \\01", raw: `'\n\01'`, expected: "01"},
		{name: "\\n then \\08", raw: `'\n\08'`, expected: "0"},

		// ================================================================
		// Non-standard escape letters (\a, \c, \d, etc.) — also skip correctly
		// ================================================================
		{name: "\\a alone", raw: `'\a'`, expected: ""},
		{name: "\\a then \\1", raw: `'\a\1'`, expected: "1"},
		{name: "\\c then \\1", raw: `'\c\1'`, expected: "1"},
		{name: "\\d then \\1", raw: `'\d\1'`, expected: "1"},
		{name: "\\e then \\1", raw: `'\e\1'`, expected: "1"},
		{name: "\\w then \\1", raw: `'\w\1'`, expected: "1"},
		{name: "\\z then \\1", raw: `'\z\1'`, expected: "1"},

		// ================================================================
		// Mixed complex sequences — many escapes then octal deep inside
		// ================================================================
		{name: "many escapes then octal at end", raw: `'\n\t\r\\\0\x41\1'`, expected: "1"},
		{name: "many escapes no octal", raw: `'\n\t\r\\\0\x41'`, expected: ""},
		{name: "interleaved null and escaped-backslash then octal", raw: `'\0\\\0\\\1'`, expected: "1"},
		{name: "interleaved null and escaped-backslash no octal", raw: `'\0\\\0\\\0'`, expected: ""},

		// ================================================================
		// Surrounding content — octal in various positions
		// ================================================================
		{name: "octal after content", raw: `"foo \01 bar"`, expected: "01"},
		{name: "space before \\1", raw: `' \1'`, expected: "1"},
		{name: "\\1 then space", raw: `'\1 '`, expected: "1"},
		{name: "a before \\1", raw: `'a\1'`, expected: "1"},
		{name: "\\1 then a", raw: `'\1a'`, expected: "1"},
		{name: "a \\1 a", raw: `'a\1a'`, expected: "1"},
		{name: "space before \\01", raw: `' \01'`, expected: "01"},
		{name: "\\01 then space", raw: `'\01 '`, expected: "01"},
		{name: "a \\01 a", raw: `'a\01a'`, expected: "01"},

		// ================================================================
		// Only first octal escape is reported
		// ================================================================
		{name: "\\01\\02 → first", raw: `'\01\02'`, expected: "01"},
		{name: "\\02\\01 → first", raw: `'\02\01'`, expected: "02"},
		{name: "\\01\\2 → first", raw: `'\01\2'`, expected: "01"},
		{name: "\\2\\01 → first", raw: `'\2\01'`, expected: "2"},
		{name: "\\08\\1 → first (\\08 = 0)", raw: `'\08\1'`, expected: "0"},
		{name: "foo \\1 bar \\2 → first", raw: `'foo \1 bar \2'`, expected: "1"},
		{name: "\\1\\1\\1 → first", raw: `'\1\1\1'`, expected: "1"},

		// ================================================================
		// Plain digits — no backslash, not octal
		// ================================================================
		{name: "plain 0", raw: `'0'`, expected: ""},
		{name: "plain 1", raw: `'1'`, expected: ""},
		{name: "plain 8", raw: `'8'`, expected: ""},
		{name: "plain 01", raw: `'01'`, expected: ""},
		{name: "plain 08", raw: `'08'`, expected: ""},
		{name: "plain 12", raw: `'12'`, expected: ""},
		{name: "plain 377", raw: `'377'`, expected: ""},

		// ================================================================
		// Escaped backslash at end followed by nothing / closing quote
		// ================================================================
		{name: "escaped backslash at end", raw: `'abc\\'`, expected: ""},
		{name: "escaped backslash then null at end", raw: `'abc\\\0'`, expected: ""},

		// ================================================================
		// Double-quoted strings (same behavior as single-quoted)
		// ================================================================
		{name: "double-quoted \\1", raw: `"\1"`, expected: "1"},
		{name: "double-quoted \\01", raw: `"\01"`, expected: "01"},
		{name: "double-quoted \\377", raw: `"\377"`, expected: "377"},
		{name: "double-quoted escaped backslash", raw: `"\\1"`, expected: ""},
		{name: "double-quoted null", raw: `"\0"`, expected: ""},

		// ================================================================
		// Line continuation — \<newline> is "other escape", skip pair
		// ================================================================
		{name: "backslash-LF then \\1", raw: "'\\\n\\1'", expected: "1"},
		{name: "backslash-CR then \\1", raw: "'\\\r\\1'", expected: "1"},
		// \<CR><LF>: scanner skips \ and CR as pair, LF is regular char
		{name: "backslash-CRLF then \\1", raw: "'\\\r\n\\1'", expected: "1"},
		{name: "backslash-LF alone (no octal)", raw: "'\\\n'", expected: ""},
		{name: "backslash-LF then null", raw: "'\\\n\\0'", expected: ""},

		// ================================================================
		// Escaped quote characters — \' and \" before octal
		// ================================================================
		{name: "escaped single-quote then \\1", raw: `'\'\1'`, expected: "1"},
		{name: "escaped double-quote then \\1", raw: `"\"\1"`, expected: "1"},
		{name: "escaped quote alone (no octal)", raw: `'\''`, expected: ""},
		{name: "escaped quote then null", raw: `'\'\0'`, expected: ""},

		// ================================================================
		// Multi-byte UTF-8 after backslash — scanner skips 2 bytes,
		// may land mid-code-point; must not false-match continuation bytes
		// ================================================================
		{name: "backslash-multi-byte then \\1", raw: "'\\\xc3\xa9\\1'", expected: "1"},
		{name: "backslash-multi-byte alone", raw: "'\\\xc3\xa9'", expected: ""},
		// 3-byte UTF-8 (e.g. U+2028 line separator) after backslash
		{name: "backslash-3byte-utf8 then \\1", raw: "'\\\xe2\x80\xa8\\1'", expected: "1"},
		// Regular multi-byte char (not after \) before octal
		{name: "utf8-char then \\1", raw: "'\xc3\xa9\\1'", expected: "1"},
		{name: "utf8-char no octal", raw: "'\xc3\xa9'", expected: ""},

		// ================================================================
		// Boundary: \0 exactly at end (i+2 == n, no char after 0)
		// ================================================================
		{name: "\\0 right before closing quote", raw: `'\0'`, expected: ""},
		{name: "\\0 right before closing double-quote", raw: `"\0"`, expected: ""},
		// a\0 at end — \0 is last escape before closing quote
		{name: "content then \\0 at end", raw: `'abc\0'`, expected: ""},

		// ================================================================
		// Boundary: backslash as very last byte (truncated/malformed)
		// ================================================================
		{name: "trailing backslash (malformed)", raw: `'\`, expected: ""},
		{name: "content then trailing backslash", raw: `'abc\`, expected: ""},

		// ================================================================
		// Boundary: single character raw strings
		// ================================================================
		{name: "just backslash", raw: `\`, expected: ""},
		{name: "just zero", raw: `0`, expected: ""},
		{name: "just one", raw: `1`, expected: ""},
		{name: "empty raw string", raw: ``, expected: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := findOctalEscape(tt.raw)
			if result != tt.expected {
				t.Errorf("findOctalEscape(%q) = %q, want %q", tt.raw, result, tt.expected)
			}
		})
	}
}

func TestNoOctalEscapeRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoOctalEscapeRule,
		// Valid cases — code without octal escapes (parseable by TS)
		[]rule_tester.ValidTestCase{
			// Hex / unicode escapes
			{Code: `var foo = "\x51";`},
			{Code: `var foo = "\u0041";`},
			{Code: `var foo = "\u{1F600}";`},
			// Escaped backslash before digits
			{Code: `var foo = "foo \\251 bar";`},
			// Regex backreference (not a string literal)
			{Code: `var foo = /([abc]) \1/g;`},
			// \0 (NULL) variants
			{Code: `var foo = '\0';`},
			{Code: `'\0'`},
			{Code: `'\0 '`},
			{Code: `' \0'`},
			{Code: `'a\0'`},
			{Code: `'\0a'`},
			{Code: `'\0\0'`},
			// Escaped backslashes
			{Code: `'\\'`},
			{Code: `'\\0'`},
			{Code: `'\\1'`},
			{Code: `'\\01'`},
			{Code: `'\\08'`},
			{Code: `'\\12'`},
			{Code: `'\\\0'`},
			{Code: `'\0\\'`},
			{Code: `'\\\\1'`},
			{Code: `'\\\\\\1'`},
			// Plain digits
			{Code: `'0'`},
			{Code: `'1'`},
			{Code: `'8'`},
			{Code: `'01'`},
			{Code: `'08'`},
			{Code: `'80'`},
			{Code: `'12'`},
			// Standard escapes
			{Code: `'\n'`},
			{Code: `'\t'`},
			// Empty string
			{Code: `''`},
			{Code: `""`},
		},
		// Invalid cases are tested via TestFindOctalEscape above because the
		// TypeScript parser reports octal escapes as syntax errors (TS1487/TS1488),
		// preventing program creation. The rule still works in production on files
		// parsed through the lenient fallback path (gap files without tsconfig).
		[]rule_tester.InvalidTestCase{},
	)
}
