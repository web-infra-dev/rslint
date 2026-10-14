package name_replacements

import (
	"fmt"
	"strings"
	"testing"
)

func TestNameReplacementCountSaturatesWithoutOverflow(t *testing.T) {
	replacements := make(map[string]bool, 10)
	for index := range 10 {
		replacements[fmt.Sprintf("name%d", index)] = true
	}
	state := &nameReplacements{opts: options{
		replacements: map[string]map[string]bool{"x": replacements},
		allowList:    map[string]bool{},
	}}
	name := strings.Repeat("x_", 18) + "x"
	result := state.nameReplacements(name, 3)
	if result.total != maximumRelevantReplacementCount {
		t.Fatalf("total = %d, want saturated %d", result.total, maximumRelevantReplacementCount)
	}
	if len(result.samples) != 3 {
		t.Fatalf("samples = %d, want 3", len(result.samples))
	}
	message := replacementMessage(name, result, "variable")
	if !strings.Contains(message.Description, "99+ more omitted") {
		t.Fatalf("message does not cap omitted replacements: %q", message.Description)
	}
}

func TestIsUpperFirstUsesFirstJavaScriptCodeUnit(t *testing.T) {
	for _, test := range []struct {
		name  string
		value string
		want  bool
	}{
		{name: "uppercase", value: "Ǆoo", want: true},
		{name: "title case", value: "ǅoo", want: false},
		{name: "lowercase", value: "ǆoo", want: false},
		{name: "astral lowercase letter", value: "𐐨oo", want: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := isUpperFirst(test.value); got != test.want {
				t.Fatalf("isUpperFirst(%q) = %t, want %t", test.value, got, test.want)
			}
		})
	}
}
