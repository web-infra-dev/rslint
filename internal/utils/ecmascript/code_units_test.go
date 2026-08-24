package ecmascript

import (
	"slices"
	"testing"
)

// loneSurrogate writes one surrogate the way the compiler carries it in a
// string value: the three bytes UTF-8 would spell the code point with if
// surrogates were encodable.
func loneSurrogate(code rune) string {
	return string([]byte{
		byte(0xE0 | code>>12),
		byte(0x80 | (code>>6)&0x3F),
		byte(0x80 | code&0x3F),
	})
}

// Every expectation here is what indexing the string in JavaScript answers.
func TestStringCodeUnits(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  []uint16
	}{
		{name: "empty", value: "", want: []uint16{}},
		{name: "ascii", value: "ab", want: []uint16{'a', 'b'}},
		{name: "basic plane", value: "中", want: []uint16{0x4E2D}},
		{name: "last of the basic plane", value: "\uFFFF", want: []uint16{0xFFFF}},
		{name: "astral", value: "\U0001F600", want: []uint16{0xD83D, 0xDE00}},
		{name: "first astral", value: "\U00010000", want: []uint16{0xD800, 0xDC00}},
		{name: "lone high surrogate", value: loneSurrogate(0xD800), want: []uint16{0xD800}},
		{name: "lone low surrogate", value: loneSurrogate(0xDFFF), want: []uint16{0xDFFF}},
		{name: "lone surrogate between letters", value: "a" + loneSurrogate(0xD801) + "b", want: []uint16{'a', 0xD801, 'b'}},
		// U+D7FF is the character just below the surrogate block, so it shares
		// the lead byte a surrogate is carried with but is UTF-8 of its own.
		{name: "below the surrogate block", value: "\uD7FF", want: []uint16{0xD7FF}},
		{name: "not utf-8", value: "\xFF", want: []uint16{0xFFFD}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StringCodeUnits(tt.value); !slices.Equal(got, tt.want) {
				t.Errorf("StringCodeUnits(%q) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

// Every expectation here is what `a < b` and `a > b` answer in JavaScript.
func TestCompareStrings(t *testing.T) {
	tests := []struct {
		name string
		a, b string
		want int
	}{
		{name: "equal", a: "a", b: "a", want: 0},
		{name: "both empty", a: "", b: "", want: 0},
		{name: "empty against a letter", a: "", b: "a", want: -1},
		{name: "letters", a: "a", b: "b", want: -1},
		{name: "letters reversed", a: "b", b: "a", want: 1},
		{name: "prefix", a: "ab", b: "abc", want: -1},
		{name: "case", a: "B", b: "a", want: -1},
		{name: "basic plane", a: "中", b: "文", want: -1},
		// A character outside the basic plane is a pair of surrogates, which
		// rank below the end of the basic plane although the character does not.
		{name: "astral against the end of the basic plane", a: "\U0001F600", b: "\uFFFF", want: -1},
		{name: "end of the basic plane against astral", a: "\uFFFF", b: "\U0001F600", want: 1},
		{name: "astral against astral", a: "\U0001F600", b: "\U0001F601", want: -1},
		{name: "astral against a letter", a: "\U0001F600", b: "a", want: 1},
		{name: "lone surrogates", a: loneSurrogate(0xD800), b: loneSurrogate(0xD801), want: -1},
		{name: "lone surrogates reversed", a: loneSurrogate(0xD801), b: loneSurrogate(0xD800), want: 1},
		// The pair the astral character is written with starts at U+D800, so
		// the lone low surrogate outranks the whole character.
		{name: "lone low surrogate against astral", a: loneSurrogate(0xDFFF), b: "\U00010000", want: 1},
		{name: "astral against a lone low surrogate", a: "\U00010000", b: loneSurrogate(0xDFFF), want: -1},
		{name: "lone high surrogate against astral", a: loneSurrogate(0xD800), b: "\U00010000", want: -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CompareStrings(tt.a, tt.b); got != tt.want {
				t.Errorf("CompareStrings(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}
