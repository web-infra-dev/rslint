package config

import (
	"encoding/json"
	"fmt"
	"strings"
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

func TestValidateConfig_RejectsOverlongMinimatchPatterns(t *testing.T) {
	overlongASCII := strings.Repeat("a", 65_537)
	// One astral rune occupies two UTF-16 code units, which is the limit used
	// by Minimatch 10 rather than Go's byte or rune length.
	overlongUTF16 := strings.Repeat("😀", 32_769)
	tests := []struct {
		name string
		key  string
		cfg  RslintConfig
	}{
		{name: "files", key: "files", cfg: RslintConfig{{Files: []string{overlongASCII}}}},
		{name: "nested files", key: "files", cfg: RslintConfig{{FilePatternGroups: [][]string{{"**/*.js", overlongUTF16}}}}},
		{
			name: "module mixed files",
			key:  "files",
			cfg: RslintConfig{{moduleFileSelectors: []configFileSelector{{matchers: []configMatcher{
				{predicateID: "predicate-1"},
				{pattern: overlongASCII},
			}}}}},
		},
		{name: "ignores", key: "ignores", cfg: RslintConfig{{Ignores: []string{overlongASCII}}}},
		{
			name: "module mixed ignores",
			key:  "ignores",
			cfg: RslintConfig{{moduleIgnoreMatchers: []configMatcher{
				{predicateID: "predicate-1"},
				{pattern: overlongUTF16},
			}}},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateConfig(test.cfg)
			if err == nil {
				t.Fatal("expected overlong pattern to be rejected")
			}
			if message := err.Error(); !strings.Contains(message, `key "`+test.key+`"`) || !strings.Contains(message, "pattern is too long") {
				t.Fatalf("error lacks pattern context: %q", message)
			}
		})
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

func TestConfigFilesJSONSupportsEmptyAndGroup(t *testing.T) {
	var cfg RslintConfig
	if err := json.Unmarshal([]byte(`[{"files": [[]], "rules": {}}]`), &cfg); err != nil {
		t.Fatalf("empty files AND group should be accepted: %v", err)
	}
	if len(cfg) != 1 || len(cfg[0].FilePatternGroups) != 1 || cfg[0].FilePatternGroups[0] == nil {
		t.Fatalf("empty files AND group was not preserved: %+v", cfg)
	}
	if err := ValidateConfig(cfg); err != nil {
		t.Fatalf("empty files AND group should validate: %v", err)
	}
	if cfg.GetConfigForFile("src/app.ts", "/repo") == nil {
		t.Fatal("empty files AND group should match as a vacuously true selector")
	}
}

func TestConfigFilesJSONRejectsInvalidSelectorValues(t *testing.T) {
	for _, input := range []string{
		`[{"files":[null]}]`,
		`[{"files":[1]}]`,
		`[{"files":[{}]}]`,
		`[{"files":[[null]]}]`,
		`[{"files":[["**/*.ts", 1]]}]`,
	} {
		t.Run(input, func(t *testing.T) {
			var cfg RslintConfig
			if err := json.Unmarshal([]byte(input), &cfg); err == nil {
				t.Fatal("expected invalid files selector to be rejected")
			}
		})
	}
}

func TestValidateConfig_Globals(t *testing.T) {
	validJSON := `[{
		"languageOptions": {"globals": {
			"writableBoolean": true,
			"writableString": "true",
			"writable": "writable",
			"writeable": "writeable",
			"readonlyBoolean": false,
			"readonlyString": "false",
			"readonly": "readonly",
			"readable": "readable",
			"nullReadonly": null,
			"disabled": "off"
		}}
	}]`
	var valid RslintConfig
	if err := json.Unmarshal([]byte(validJSON), &valid); err != nil {
		t.Fatalf("unmarshal valid globals: %v", err)
	}
	if err := ValidateConfig(valid); err != nil {
		t.Fatalf("valid globals were rejected: %v", err)
	}

	for _, input := range []string{
		`[{"languageOptions":{"globals":null}}]`,
		`[{"languageOptions":{"globals":[]}}]`,
		`[{"languageOptions":{"globals":{"value":"invalid"}}}]`,
		`[{"languageOptions":{"globals":{" padded ":"readonly"}}}]`,
	} {
		t.Run(input, func(t *testing.T) {
			var cfg RslintConfig
			if err := json.Unmarshal([]byte(input), &cfg); err != nil {
				t.Fatalf("unmarshal invalid globals fixture: %v", err)
			}
			if err := ValidateConfig(cfg); err == nil {
				t.Fatal("expected invalid globals to be rejected")
			}
		})
	}
}

func TestValidateConfig_RuleSeverities(t *testing.T) {
	t.Run("JSON ingress", func(t *testing.T) {
		var cfg RslintConfig
		if err := json.Unmarshal([]byte(`[{
			"rules": {
				"string-off": "off",
				"string-warn": "warn",
				"string-error": "error",
				"numeric-off": 0,
				"numeric-warn": 1,
				"numeric-error": 2,
				"array-numeric": [2, "first", {"second": true}, null]
			}
		}]`), &cfg); err != nil {
			t.Fatalf("unmarshal numeric rule severities: %v", err)
		}
		if err := ValidateConfig(cfg); err != nil {
			t.Fatalf("valid JSON rule severities were rejected: %v", err)
		}
		merged := cfg.GetConfigForFile("src/app.ts", "")
		if merged == nil {
			t.Fatal("expected JSON config to merge")
		}
		for name, want := range map[string]string{
			"numeric-off": "off", "numeric-warn": "warn", "numeric-error": "error",
		} {
			if got := merged.Rules[name]; got == nil || got.Level != want {
				t.Errorf("JSON rule %q = %#v, want level %q", name, got, want)
			}
		}
		arrayRule := merged.Rules["array-numeric"]
		if arrayRule == nil || arrayRule.Level != "error" || len(arrayRule.Options) != 3 {
			t.Fatalf("JSON numeric array was not preserved: %#v", arrayRule)
		}
	})

	t.Run("Go construction", func(t *testing.T) {
		cfg := RslintConfig{{Rules: Rules{
			"numeric-off":   int8(0),
			"numeric-warn":  uint16(1),
			"numeric-error": float32(2),
			"array-numeric": []interface{}{uint8(1), "first", true},
		}}}
		if err := ValidateConfig(cfg); err != nil {
			t.Fatalf("valid Go-constructed rule severities were rejected: %v", err)
		}
	})
}

func TestValidateConfig_RejectsInvalidRuleValues(t *testing.T) {
	invalidJSON := []string{
		`"fatal"`,
		`3`,
		`1.5`,
		`true`,
		`null`,
		`{}`,
		`[]`,
		`["warning"]`,
		`[3]`,
		`[null]`,
	}
	for _, value := range invalidJSON {
		t.Run("JSON "+value, func(t *testing.T) {
			var cfg RslintConfig
			input := `[{"rules":{"example":` + value + `}}]`
			err := json.Unmarshal([]byte(input), &cfg)
			if err == nil {
				t.Fatal("expected invalid rule value to be rejected during JSON ingress")
			}
			if message := err.Error(); !strings.Contains(message, `key "rules": rule "example"`) {
				t.Fatalf("error does not identify the invalid rule: %q", message)
			}
		})
	}

	goConstructed := []any{"fatal", 3, 1.5, true, nil, map[string]any{}, []interface{}{}}
	for index, value := range goConstructed {
		t.Run(fmt.Sprintf("Go %d %T", index, value), func(t *testing.T) {
			err := ValidateConfig(RslintConfig{{Rules: Rules{"example": value}}})
			if err == nil {
				t.Fatal("expected invalid Go-constructed rule value to be rejected")
			}
			if message := err.Error(); !strings.Contains(message, `key "rules": rule "example"`) {
				t.Fatalf("error does not identify the invalid rule: %q", message)
			}
		})
	}
}
