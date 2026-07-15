package config

import "testing"

func TestParseIgnorePatternPreservesOrderedNegation(t *testing.T) {
	tests := []struct {
		raw     string
		glob    string
		negated bool
	}{
		{raw: "src/**", glob: "src/**"},
		{raw: "!./src/**", glob: "src/**", negated: true},
		{raw: "!!!src/**", glob: "src/**", negated: true},
		{raw: "", glob: ""},
	}
	for _, test := range tests {
		parsed := ParseIgnorePattern(test.raw)
		if parsed.Glob != test.glob || parsed.Negated != test.negated {
			t.Fatalf("ParseIgnorePattern(%q) = {%q, %v}, want {%q, %v}", test.raw, parsed.Glob, parsed.Negated, test.glob, test.negated)
		}
	}
}

func TestIsFileIgnoredUsesOrderedLastMatch(t *testing.T) {
	patterns := ParseIgnorePatterns([]string{"**/*.js", "!src/keep.js", "src/keep.js"})
	if !isFileIgnored("src/keep.js", patterns, "") {
		t.Fatal("the final matching positive pattern must win")
	}
	patterns = append(patterns, ParseIgnorePattern("!src/keep.js"))
	if isFileIgnored("src/keep.js", patterns, "") {
		t.Fatal("the final matching negation must re-include the file")
	}
}
