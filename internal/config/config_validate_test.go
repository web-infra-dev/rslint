package config

import (
	"encoding/json"
	"testing"
)

func TestValidateConfig_AllowsMissingFiles(t *testing.T) {
	cfg := RslintConfig{
		{Rules: Rules{"no-console": "error"}},
	}
	if err := ValidateConfig(cfg); err != nil {
		t.Fatalf("ValidateConfig returned error: %v", err)
	}
}

func TestValidateConfig_AllowsNonEmptyFiles(t *testing.T) {
	cfg := RslintConfig{
		{Files: []string{"**/*.ts"}, Rules: Rules{"no-console": "error"}},
	}
	if err := ValidateConfig(cfg); err != nil {
		t.Fatalf("ValidateConfig returned error: %v", err)
	}
}

func TestValidateConfig_RejectsEmptyFilesArray(t *testing.T) {
	cfg := RslintConfig{
		{Files: []string{}, Rules: Rules{"no-console": "error"}},
	}
	err := ValidateConfig(cfg)
	if err == nil {
		t.Fatal("expected empty files array to be rejected")
	}
	if got := err.Error(); got != `config entry at index 0: key "files": expected value to be a non-empty array` {
		t.Fatalf("unexpected error: %q", got)
	}
}

func TestValidateConfig_RejectsEmptyFilesOnIgnoreOnlyEntry(t *testing.T) {
	cfg := RslintConfig{
		{Files: []string{}, Ignores: []string{"dist/**"}},
	}
	if err := ValidateConfig(cfg); err == nil {
		t.Fatal("expected empty files array on ignore-only entry to be rejected")
	}
}

func TestValidateConfig_RejectsNullFilesFromJSON(t *testing.T) {
	var cfg RslintConfig
	err := json.Unmarshal([]byte(`[{"files": null, "rules": {}}]`), &cfg)
	if err == nil {
		t.Fatal("expected null files field to be rejected while unmarshaling")
	}
	if got := err.Error(); got != `config entry at index 0: key "files": expected value to be a non-empty array` {
		t.Fatalf("unexpected error: %q", got)
	}
}

func TestValidateConfig_RejectsEmptyFilesArrayFromJSON(t *testing.T) {
	var cfg RslintConfig
	err := json.Unmarshal([]byte(`[{"files": [], "rules": {}}]`), &cfg)
	if err == nil {
		t.Fatal("expected empty files array to be rejected while unmarshaling")
	}
	if got := err.Error(); got != `config entry at index 0: key "files": expected value to be a non-empty array` {
		t.Fatalf("unexpected error: %q", got)
	}
}
