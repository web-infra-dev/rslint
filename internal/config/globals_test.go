package config

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/utils"
)

// ExtractGlobals resolves every alias ESLint accepts to one of its three access
// levels. Booleans follow the `globals` npm package: true is writable, false is
// readonly.
func TestExtractGlobals(t *testing.T) {
	langOpts := &LanguageOptions{
		Raw: map[string]any{
			"globals": map[string]any{
				"boolTrue":       true,
				"stringTrue":     "true",
				"writable":       "writable",
				"writeable":      "writeable",
				"boolFalse":      false,
				"stringFalse":    "false",
				"readonly":       "readonly",
				"readable":       "readable",
				"nullReadonly":   nil,
				"stringDisabled": "off",
				"invalid":        "nonsense",
			},
		},
	}

	globals := ExtractGlobals(langOpts)

	cases := map[string]utils.GlobalAccess{
		"boolTrue":       utils.GlobalAccessWritable,
		"stringTrue":     utils.GlobalAccessWritable,
		"writable":       utils.GlobalAccessWritable,
		"writeable":      utils.GlobalAccessWritable,
		"boolFalse":      utils.GlobalAccessReadonly,
		"stringFalse":    utils.GlobalAccessReadonly,
		"readonly":       utils.GlobalAccessReadonly,
		"readable":       utils.GlobalAccessReadonly,
		"nullReadonly":   utils.GlobalAccessReadonly,
		"stringDisabled": utils.GlobalAccessOff,
		"invalid":        utils.GlobalAccessUnset,
	}
	for name, want := range cases {
		if got := globals[name]; got != want {
			t.Errorf("ExtractGlobals()[%q] = %v, want %v", name, got, want)
		}
	}
}

func TestExtractGlobals_NoLanguageOptions(t *testing.T) {
	if got := ExtractGlobals(nil); got != nil {
		t.Errorf("ExtractGlobals(nil) = %v, want nil", got)
	}
	if got := ExtractGlobals(&LanguageOptions{}); got != nil {
		t.Errorf("ExtractGlobals(empty) = %v, want nil", got)
	}
}

func TestExtractSourceType(t *testing.T) {
	if got := ExtractSourceType(nil); got != "" {
		t.Errorf("ExtractSourceType(nil) = %q, want \"\"", got)
	}
	if got := ExtractSourceType(&LanguageOptions{}); got != "" {
		t.Errorf("ExtractSourceType(empty) = %q, want \"\"", got)
	}
	if got := ExtractSourceType(&LanguageOptions{Raw: map[string]any{"sourceType": "module"}}); got != "module" {
		t.Errorf("ExtractSourceType(sourceType) = %q, want \"module\"", got)
	}
	if got := ExtractSourceType(&LanguageOptions{Raw: map[string]any{
		"parserOptions": map[string]any{"sourceType": "script"},
	}}); got != "script" {
		t.Errorf("ExtractSourceType(parserOptions.sourceType) = %q, want \"script\"", got)
	}
	if got := ExtractSourceType(&LanguageOptions{Raw: map[string]any{
		"sourceType":    "module",
		"parserOptions": map[string]any{"sourceType": "script"},
	}}); got != "module" {
		t.Errorf("ExtractSourceType prefers languageOptions.sourceType, got %q", got)
	}
}

func TestMergeLanguageOptions_MergesGlobalsByName(t *testing.T) {
	base := &LanguageOptions{Raw: map[string]any{
		"globals": map[string]any{
			"baseOnly": "readonly",
			"shared":   "writable",
		},
	}}
	override := &LanguageOptions{Raw: map[string]any{
		"globals": map[string]any{
			"overrideOnly": "readonly",
			"shared":       "off",
		},
	}}

	merged := mergeLanguageOptions(base, override)
	merged = mergeLanguageOptions(merged, &LanguageOptions{Raw: map[string]any{
		"globals": map[string]any{},
	}})

	rawGlobals, ok := merged.Raw["globals"].(map[string]any)
	if !ok {
		t.Fatalf("merged globals has type %T, want map[string]any", merged.Raw["globals"])
	}
	if got := rawGlobals["baseOnly"]; got != "readonly" {
		t.Errorf("baseOnly = %v, want readonly", got)
	}
	if got := rawGlobals["overrideOnly"]; got != "readonly" {
		t.Errorf("overrideOnly = %v, want readonly", got)
	}
	if got := rawGlobals["shared"]; got != "off" {
		t.Errorf("shared = %v, want off", got)
	}
	if got := ExtractGlobals(merged)["shared"]; got != utils.GlobalAccessOff {
		t.Errorf("later off value should undeclare shared, got %v", got)
	}
}
