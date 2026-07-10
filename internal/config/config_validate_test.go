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

func TestConfigFilesJSONSupportsMixedStringsAndAndGroups(t *testing.T) {
	input := []byte(`[{
		"files": ["special.ts", ["**/*.js", "!**/*.test.js"]],
		"rules": {"no-console": "error"}
	}]`)
	var cfg RslintConfig
	if err := json.Unmarshal(input, &cfg); err != nil {
		t.Fatalf("unmarshal mixed files selectors: %v", err)
	}
	if len(cfg) != 1 || len(cfg[0].Files) != 1 || cfg[0].Files[0] != "special.ts" {
		t.Fatalf("unexpected top-level files: %+v", cfg)
	}
	if len(cfg[0].FilePatternGroups) != 1 || len(cfg[0].FilePatternGroups[0]) != 2 {
		t.Fatalf("unexpected AND groups: %+v", cfg[0].FilePatternGroups)
	}

	roundTrip, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("marshal mixed files selectors: %v", err)
	}
	var decoded RslintConfig
	if err := json.Unmarshal(roundTrip, &decoded); err != nil {
		t.Fatalf("round-trip mixed files selectors: %v", err)
	}
	if len(decoded[0].Files) != 1 || len(decoded[0].FilePatternGroups) != 1 {
		t.Fatalf("mixed files selectors were not preserved: %s", roundTrip)
	}

	var entry ConfigEntry
	if err := json.Unmarshal(input[1:len(input)-1], &entry); err != nil {
		t.Fatalf("direct ConfigEntry unmarshal must support mixed files selectors: %v", err)
	}
	if len(entry.Files) != 1 || len(entry.FilePatternGroups) != 1 {
		t.Fatalf("direct ConfigEntry unmarshal lost mixed selectors: %+v", entry)
	}
}

func TestConfigFilesJSONRejectsEmptyAndGroup(t *testing.T) {
	var cfg RslintConfig
	err := json.Unmarshal([]byte(`[{"files": [[]], "rules": {}}]`), &cfg)
	if err == nil {
		t.Fatal("expected an empty files AND group to be rejected")
	}
}
