package utils

import "testing"

func TestExtractRegexPatternAndFlags(t *testing.T) {
	tests := []struct {
		input   string
		pattern string
		flags   string
	}{
		{`/abc/`, `abc`, ``},
		{`/abc/gi`, `abc`, `gi`},
		{`/abc/v`, `abc`, `v`},
		{`/a\/b/`, `a\/b`, ``},
		{`//`, ``, ``},
		{"/abc/gi" + "ms", "abc", "gi" + "ms"},
		{``, ``, ``},
		{`x`, ``, ``},
	}
	for _, tt := range tests {
		p, f := ExtractRegexPatternAndFlags(tt.input)
		if p != tt.pattern || f != tt.flags {
			t.Errorf("ExtractRegexPatternAndFlags(%q) = (%q, %q), want (%q, %q)", tt.input, p, f, tt.pattern, tt.flags)
		}
	}
}
