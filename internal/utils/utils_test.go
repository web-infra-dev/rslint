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

func TestDefaultIgnoreDirGlobs(t *testing.T) {
	globs := DefaultIgnoreDirGlobs()

	if len(globs) != len(DefaultExcludeDirNames) {
		t.Fatalf("Expected %d globs, got %d", len(DefaultExcludeDirNames), len(globs))
	}

	for i, name := range DefaultExcludeDirNames {
		expected := name + "/**"
		if globs[i] != expected {
			t.Errorf("Expected glob %q for dir %q, got %q", expected, name, globs[i])
		}
	}
}

func TestDefaultExcludeDirNames_ContainsExpected(t *testing.T) {
	expected := map[string]bool{"node_modules": false, ".git": false}

	for _, name := range DefaultExcludeDirNames {
		if _, ok := expected[name]; ok {
			expected[name] = true
		}
	}

	for name, found := range expected {
		if !found {
			t.Errorf("Expected %q in DefaultExcludeDirNames", name)
		}
	}
}
