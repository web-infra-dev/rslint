package ecmascript

import "testing"

func TestJSXTextTokenValuesEqual(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		left, right string
		want        bool
	}{
		{name: "identical raw", left: "plain text", right: "plain text", want: true},
		{name: "named entity", left: "&amp;", right: "&", want: true},
		{name: "decimal entity", left: "&#65;", right: "A", want: true},
		{name: "hex entity", left: "&#x41;", right: "A", want: true},
		{name: "uppercase hex marker is invalid", left: "&#X41;", right: "A", want: false},
		{name: "semicolon is required", left: "&amp", right: "&", want: false},
		{name: "tenth unit may be semicolon", left: "&#00000000;", right: "\x00", want: true},
		{name: "largest decimal at tenth unit", left: "&#99999999;", right: "\uE0FF", want: true},
		{name: "largest hex at tenth unit", left: "&#xFFFFFFF;", right: "\uFFFF", want: true},
		{name: "semicolon after tenth unit is too late", left: "&#000000000;", right: "\x00", want: false},
		{name: "invalid decimal digits stay raw", left: "&#12a;", right: "\f", want: false},
		{name: "invalid hex digits stay raw", left: "&#x1g;", right: "\x01", want: false},
		{name: "invalid entity resumes after ampersand", left: "&bogus;&amp;", right: "&bogus;&", want: true},
		{name: "adjacent ampersands", left: "&&amp;", right: "&&", want: true},
		{name: "source CRLF folds", left: "left\r\nright", right: "left\nright", want: true},
		{name: "lone source CR remains", left: "left\rright", right: "left\nright", want: false},
		{name: "entity CR does not fold", left: "&#13;\n", right: "\r\n", want: false},
		{name: "fromCharCode truncates astral numeric entity", left: "&#x1F4A9;", right: "\uF4A9", want: true},
		{name: "single astral numeric entity is not the astral character", left: "&#x1F4A9;", right: "💩", want: false},
		{name: "surrogate entity pair equals astral character", left: "&#xD83D;&#xDCA9;", right: "💩", want: true},
		{name: "lone surrogate preserves its code unit", left: "&#xD83D;", right: "&#55357;", want: true},
		{name: "different lone surrogates differ", left: "&#xD83D;", right: "&#xD83E;", want: false},
		{name: "numeric NUL", left: "a&#0;b", right: "a\x00b", want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := JSXTextTokenValuesEqual(tt.left, tt.right); got != tt.want {
				t.Errorf("JSXTextTokenValuesEqual(%q, %q) = %v, want %v", tt.left, tt.right, got, tt.want)
			}
		})
	}
}

func TestDecodeJSXEntities(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name, raw, want string
	}{
		{name: "named entity", raw: "no&amp;pener", want: "no&pener"},
		{name: "hex entity", raw: "no&#x6f;pener", want: "noopener"},
		{name: "invalid entity", raw: "no&#xZZ;pener", want: "no&#xZZ;pener"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := DecodeJSXEntities(tt.raw); got != tt.want {
				t.Errorf("DecodeJSXEntities(%q) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}
