package config

import "testing"

// Locks in normalizeConfigGlobal parity with ESLint (source-code.js): only
// the string "off" un-declares a global. Boolean false and null both map to
// "readonly" and remain declared — a common mix-up since other ESLint config
// knobs treat false/"off" as equivalent, but globals don't.
func TestExtractGlobals(t *testing.T) {
	langOpts := &LanguageOptions{
		Raw: map[string]any{
			"globals": map[string]any{
				"stringOff":      "off",
				"stringReadonly": "readonly",
				"stringWritable": "writable",
				"boolTrue":       true,
				"boolFalse":      false,
				"nullValue":      nil,
			},
		},
	}

	globals := ExtractGlobals(langOpts)

	cases := map[string]bool{
		"stringOff":      false,
		"stringReadonly": true,
		"stringWritable": true,
		"boolTrue":       true,
		"boolFalse":      true,
		"nullValue":      true,
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
	if got := ExtractGlobals(merged)["shared"]; got {
		t.Error("later off value should undeclare shared")
	}
}
