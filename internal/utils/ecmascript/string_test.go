package ecmascript

import (
	"testing"

	"github.com/microsoft/typescript-go/shim/stringutil"
)

func TestStringCompare(t *testing.T) {
	t.Parallel()
	loneHighSurrogate := stringutil.EncodeJSStringRune(0xD800)
	tests := []struct {
		name string
		a    string
		b    string
		want int
	}{
		{name: "equal", a: "module", b: "module", want: 0},
		{name: "ASCII", a: "a", b: "b", want: -1},
		{name: "prefix", a: "pkg", b: "pkg/subpath", want: -1},
		{name: "UTF-16 order", a: "\U00010000", b: "\uE000", want: -1},
		{name: "UTF-16 reverse", a: "\uE000", b: "\U00010000", want: 1},
		{name: "non-ASCII prefix", a: "é", b: "é/subpath", want: -1},
		{name: "non-ASCII prefix reverse", a: "é/subpath", b: "é", want: 1},
		{name: "lone surrogate", a: loneHighSurrogate, b: "\uE000", want: -1},
		{name: "lone surrogate prefix", a: loneHighSurrogate, b: "\U00010000", want: -1},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if got := StringCompare(test.a, test.b); got != test.want {
				t.Fatalf("StringCompare(%q, %q) = %d, want %d", test.a, test.b, got, test.want)
			}
		})
	}
}

func TestStringToLowerCase(t *testing.T) {
	t.Parallel()
	loneHighSurrogate := stringutil.EncodeJSStringRune(0xD800)
	tests := []struct {
		in   string
		want string
	}{
		{in: "ABC", want: "abc"},
		{in: "İ", want: "i\u0307"},
		{in: "ΟΣ", want: "ος"},
		{in: "ΟΣΑ", want: "οσα"},
		{in: loneHighSurrogate, want: loneHighSurrogate},
	}
	for _, test := range tests {
		if got := StringToLowerCase(test.in); got != test.want {
			t.Errorf("StringToLowerCase(%q) = %q, want %q", test.in, got, test.want)
		}
	}
}
