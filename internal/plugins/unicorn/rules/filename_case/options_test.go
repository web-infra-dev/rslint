package filename_case

import (
	"slices"
	"testing"
)

func TestParseOptionsUsesCanonicalCasesOrder(t *testing.T) {
	opts := parseOptions([]any{map[string]any{"cases": map[string]any{
		"snakeCase": true,
		"kebabCase": true,
	}}})
	got := make([]string, opts.caseCount)
	for index, configuredCase := range opts.selectedCases() {
		got[index] = configuredCase.key
	}
	want := []string{"kebabCase", "snakeCase"}
	if !slices.Equal(got, want) {
		t.Fatalf("selected cases = %v, want %v", got, want)
	}
}
