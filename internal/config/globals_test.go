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
