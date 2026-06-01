package utils

import (
	"testing"
)

func TestIsTriviaWhitespaceByte(t *testing.T) {
	for _, c := range []byte{' ', '\t', '\n', '\r', '\f', '\v'} {
		if !IsTriviaWhitespaceByte(c) {
			t.Errorf("expected byte %q to be whitespace", c)
		}
	}
	for _, c := range []byte{'a', '0', '_', '/', '*', 0x00, 0x7F, 0x80} {
		if IsTriviaWhitespaceByte(c) {
			t.Errorf("expected byte %q to NOT be whitespace", c)
		}
	}
}

func TestIsTriviaWhitespaceRune(t *testing.T) {
	// All ECMAScript-recognized non-ASCII WhiteSpace + LineTerminator runes.
	whitespace := []rune{
		0x00A0, // NBSP
		0x1680, // Ogham
		0x2000, 0x2001, 0x2002, 0x2003, 0x2004, 0x2005, 0x2006, 0x2007, 0x2008, 0x2009, 0x200A,
		0x2028, // LS
		0x2029, // PS
		0x202F, 0x205F, 0x3000,
		0xFEFF, // ZWNBSP/BOM
	}
	for _, r := range whitespace {
		if !IsTriviaWhitespaceRune(r) {
			t.Errorf("expected rune U+%04X to be whitespace", r)
		}
	}

	// Runes that look whitespace-ish but are NOT ECMAScript WhiteSpace.
	nonWhitespace := []rune{
		0x0085, // NEL — in \p{White_Space} but not in ES WhiteSpace
		0x200B, // ZWSP — category Cf, not Zs
		0x200C, // ZWNJ
		0x200D, // ZWJ
		'a', '0', '/', 0x4E2D, // CJK "中" — NOT whitespace
	}
	for _, r := range nonWhitespace {
		if IsTriviaWhitespaceRune(r) {
			t.Errorf("expected rune U+%04X to NOT be whitespace", r)
		}
	}
}

func TestContainsLineTerminator(t *testing.T) {
	cases := []struct {
		name string
		text string
		want bool
	}{
		{"empty", "", false},
		{"plain ascii", "abc", false},
		{"only spaces", "   \t   ", false},
		{"LF", "a\nb", true},
		{"CR", "a\rb", true},
		{"CRLF", "a\r\nb", true},
		{"LS", "a b", true},
		{"PS", "a b", true},
		{"NBSP only — NOT a line terminator", "a b", false},
		{"IDEO only", "a\u3000b", false},
		{"LS at start", " b", true},
		{"LS at end", "a ", true},
		{"PS at end", "a ", true},
		// CJK "啊" is E5 95 8A — must NOT trigger.
		{"CJK no false-positive", "\u554A\u554A", false},
		// U+2030 = E2 80 B0 (per mille) — has E2 80 prefix but third byte
		// is B0, not A8/A9, so must NOT trigger.
		{"0xE2-0x80 prefix but not LS/PS", "a\u2030b", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := ContainsLineTerminator(c.text, 0, len(c.text)); got != c.want {
				t.Errorf("ContainsLineTerminator(%q) = %v, want %v", c.text, got, c.want)
			}
		})
	}
}

func TestContainsLineTerminator_RangeClamping(t *testing.T) {
	text := "a\nb"
	if !ContainsLineTerminator(text, -10, 100) {
		t.Error("expected clamping to find LF")
	}
	if ContainsLineTerminator(text, 0, 0) {
		t.Error("expected empty range to be false")
	}
	if ContainsLineTerminator(text, 2, 1) {
		t.Error("expected reversed range to be false")
	}
}

func TestSkipTrailingWhitespace(t *testing.T) {
	cases := []struct {
		name string
		text string
		want int // expected returned position (relative to start; high = len(text))
	}{
		{"empty", "", 0},
		{"no whitespace", "abc", 3},
		{"single trailing space", "ab ", 2},
		{"multi ASCII whitespace", "ab \t\n", 2},
		{"only whitespace returns low", "   ", 0},
		// NBSP is 2 bytes; trailing NBSP must be fully skipped.
		{"NBSP trailing", "ab ", 2},
		{"NBSP + ASCII trailing", "ab  \t", 2},
		// LS is 3 bytes.
		{"LS trailing", "ab ", 2},
		{"PS + space trailing", "ab  ", 2},
		{"IDEO trailing", "ab\u3000", 2},
		{"BOM trailing", "ab\uFEFF", 2},
		// Non-whitespace non-ASCII must NOT be skipped.
		{"CJK trailing", "ab\u4E2D", 5},
		// Whitespace then non-WS non-ASCII = stop AT the non-WS end.
		{"WS then CJK from right", "a\u4E2D ", 4},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := SkipTrailingWhitespace(c.text, 0, len(c.text))
			if got != c.want {
				t.Errorf("SkipTrailingWhitespace(%q) = %d, want %d", c.text, got, c.want)
			}
		})
	}
}

func TestSkipTrailingWhitespace_RangeClamping(t *testing.T) {
	text := "abc   "
	if got := SkipTrailingWhitespace(text, -5, 100); got != 3 {
		t.Errorf("expected clamped scan to return 3, got %d", got)
	}
	if got := SkipTrailingWhitespace(text, 0, 0); got != 0 {
		t.Errorf("expected empty range to return low (0), got %d", got)
	}
}

func TestSkipLeadingWhitespace(t *testing.T) {
	cases := []struct {
		name string
		text string
		want int
	}{
		{"empty", "", 0},
		{"no whitespace", "abc", 0},
		{"single leading space", " ab", 1},
		{"multi ASCII whitespace", " \t\nab", 3},
		{"only whitespace returns high", "   ", 3},
		// NBSP is 2 bytes.
		{"NBSP leading", "\u00A0ab", 2},
		// LS is 3 bytes.
		{"LS leading", "\u2028ab", 3},
		{"IDEO leading", "\u3000ab", 3},
		// Non-whitespace non-ASCII must NOT be skipped.
		{"CJK leading", "\u4E2Dab", 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := SkipLeadingWhitespace(c.text, 0, len(c.text))
			if got != c.want {
				t.Errorf("SkipLeadingWhitespace(%q) = %d, want %d", c.text, got, c.want)
			}
		})
	}
}
