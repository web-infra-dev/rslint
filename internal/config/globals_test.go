package config

import "testing"

// ExtractGlobals preserves the declared/disabled state needed by native rules.
// Access aliases, including null-as-readonly, remain declared; only "off"
// explicitly disables a name.
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
			},
		},
	}

	globals := ExtractGlobals(langOpts)

	cases := map[string]bool{
		"boolTrue":       true,
		"stringTrue":     true,
		"writable":       true,
		"writeable":      true,
		"boolFalse":      true,
		"stringFalse":    true,
		"readonly":       true,
		"readable":       true,
		"nullReadonly":   true,
		"stringDisabled": false,
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
