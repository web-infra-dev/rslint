package utils

import "testing"

func TestEscapeJSSingleQuotedString(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{input: "plain 日本語", want: "plain 日本語"},
		{input: `a\b'c`, want: `a\\b\'c`},
		{input: "\b\f\n\r\t\v", want: `\b\f\n\r\t\v`},
		{input: "\x00\x01\x1f", want: `\x00\x01\x1F`},
		{input: "a\u2028b\u2029c", want: `a\u2028b\u2029c`},
		{input: string([]byte{0xED, 0xA0, 0x80}) + "x" + string([]byte{0xED, 0xB0, 0x80}), want: `\uD800x\uDC00`},
		{input: string([]byte{0xFF, 'a'}), want: `\xFFa`},
	}
	for _, test := range tests {
		if got := EscapeJSSingleQuotedString(test.input); got != test.want {
			t.Errorf("EscapeJSSingleQuotedString(%q) = %q, want %q", test.input, got, test.want)
		}
	}
}
