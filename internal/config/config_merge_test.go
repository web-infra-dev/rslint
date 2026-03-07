package config

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func TestGetConfigForFile_ExplicitRulesOnly(t *testing.T) {
	config := RslintConfig{
		{
			Rules: Rules{
				"no-debugger": "error",
			},
		},
	}

	merged := config.GetConfigForFile("src/app.ts")
	if merged == nil {
		t.Fatal("Expected non-nil merged config")
	}

	// Only explicitly listed rules should be present
	if _, ok := merged.Rules["no-debugger"]; !ok {
		t.Error("Expected no-debugger rule to be present")
	}
	if len(merged.Rules) != 1 {
		t.Errorf("Expected exactly 1 rule, got %d", len(merged.Rules))
	}
}

func TestNormalizeJSONConfig_CoreRulesDefault(t *testing.T) {
	RegisterAllRules()

	config := normalizeJSONConfig(RslintConfig{
		{
			Rules: Rules{},
		},
	})

	merged := config.GetConfigForFile("src/app.ts")
	if merged == nil {
		t.Fatal("Expected non-nil merged config")
	}

	// After normalization, core rules should be present
	coreRules := GetCoreRules()
	if len(coreRules) == 0 {
		t.Fatal("Expected at least one core rule to be registered")
	}

	for _, r := range coreRules {
		rc, ok := merged.Rules[r.Name]
		if !ok {
			t.Errorf("Expected core rule %q to be in merged config", r.Name)
			continue
		}
		if rc.Level != "error" {
			t.Errorf("Expected core rule %q level to be 'error', got %q", r.Name, rc.Level)
		}
	}
}

func TestNormalizeJSONConfig_PluginAutoEnablesRules(t *testing.T) {
	RegisterAllRules()

	config := normalizeJSONConfig(RslintConfig{
		{
			Plugins: []string{"@typescript-eslint"},
		},
	})

	merged := config.GetConfigForFile("src/app.ts")
	if merged == nil {
		t.Fatal("Expected non-nil merged config")
	}

	// Should have TS rules enabled
	if _, ok := merged.Rules["@typescript-eslint/no-explicit-any"]; !ok {
		t.Error("Expected @typescript-eslint/no-explicit-any to be enabled via plugin")
	}

	// Should NOT have import rules (only @typescript-eslint/ prefix)
	for name := range merged.Rules {
		if len(name) > 7 && name[:7] == "import/" {
			t.Errorf("Unexpected import rule %q enabled by @typescript-eslint plugin", name)
		}
	}
}

func TestNormalizeJSONConfig_UserRulesTakePrecedence(t *testing.T) {
	RegisterAllRules()

	config := normalizeJSONConfig(RslintConfig{
		{
			Rules: Rules{
				"no-debugger": "off",
			},
		},
	})

	merged := config.GetConfigForFile("src/app.ts")
	if merged == nil {
		t.Fatal("Expected non-nil merged config")
	}

	// User's "off" should override the auto-enabled "error"
	rc := merged.Rules["no-debugger"]
	if rc == nil {
		t.Fatal("Expected no-debugger rule to be present")
	}
	if rc.IsEnabled() {
		t.Error("Expected no-debugger to be disabled (user set 'off')")
	}
}

func TestGetConfigForFile_WithoutNormalize_PluginDoesNotAutoEnable(t *testing.T) {
	RegisterAllRules()

	// Without normalizeJSONConfig, plugins should not auto-enable rules
	config := RslintConfig{
		{
			Plugins: []string{"@typescript-eslint"},
		},
	}

	merged := config.GetConfigForFile("src/app.ts")
	if merged == nil {
		t.Fatal("Expected non-nil merged config")
	}

	// No rules should be enabled (JS config behavior)
	if len(merged.Rules) != 0 {
		t.Errorf("Expected 0 rules without normalization, got %d", len(merged.Rules))
	}
}

func TestGetConfigForFile_GlobalIgnores(t *testing.T) {
	config := RslintConfig{
		{
			Ignores: []string{"dist/**"},
		},
		{
			Rules: Rules{"no-debugger": "error"},
		},
	}

	// File in dist should be ignored
	merged := config.GetConfigForFile("dist/bundle.js")
	if merged != nil {
		t.Error("Expected nil for globally ignored file")
	}

	// File not in dist should not be ignored
	merged = config.GetConfigForFile("src/app.ts")
	if merged == nil {
		t.Fatal("Expected non-nil for non-ignored file")
	}
}

func TestGetConfigForFile_EntryIgnores_NoMatch(t *testing.T) {
	config := RslintConfig{
		{
			Files:   []string{"**/*.ts"},
			Ignores: []string{"**/*.test.ts"},
			Rules:   Rules{"no-debugger": "error"},
		},
	}

	// Test file is ignored by entry-level ignores and no other entry matches
	// Should return nil (file should not be linted)
	merged := config.GetConfigForFile("src/app.test.ts")
	if merged != nil {
		t.Error("Expected nil for file ignored by all entries")
	}
}

func TestGetConfigForFile_EntryIgnores_OtherEntryMatches(t *testing.T) {
	config := RslintConfig{
		{
			Files:   []string{"**/*.ts"},
			Ignores: []string{"**/*.test.ts"},
			Rules:   Rules{"no-debugger": "error"},
		},
		{
			Files: []string{"**/*.test.ts"},
			Rules: Rules{"no-console": "warn"},
		},
	}

	// Test file is ignored by first entry but matched by second
	merged := config.GetConfigForFile("src/app.test.ts")
	if merged == nil {
		t.Fatal("Expected non-nil config (matched by second entry)")
	}
	if _, ok := merged.Rules["no-debugger"]; ok {
		t.Error("Expected no-debugger to not be present (from ignored entry)")
	}
	if _, ok := merged.Rules["no-console"]; !ok {
		t.Error("Expected no-console from second entry")
	}
}

func TestGetConfigForFile_FilesMatching(t *testing.T) {
	config := RslintConfig{
		{
			Files: []string{"**/*.ts"},
			Rules: Rules{"no-debugger": "error"},
		},
	}

	// TS file should match
	merged := config.GetConfigForFile("src/app.ts")
	if merged == nil {
		t.Fatal("Expected non-nil config")
	}
	if _, ok := merged.Rules["no-debugger"]; !ok {
		t.Error("Expected no-debugger for matching .ts file")
	}

	// JS file should not match — no entry matches, return nil
	merged = config.GetConfigForFile("src/app.js")
	if merged != nil {
		t.Error("Expected nil for non-matching file with no other entries")
	}
}

func TestGetConfigForFile_RulesShallowMerge(t *testing.T) {
	config := RslintConfig{
		{
			Rules: Rules{
				"no-debugger": "error",
				"no-console":  "error",
			},
		},
		{
			Rules: Rules{
				"no-debugger":   "warn",
				"for-direction": "error",
			},
		},
	}

	merged := config.GetConfigForFile("src/app.ts")
	if merged == nil {
		t.Fatal("Expected non-nil config")
	}

	// no-debugger should be overridden to "warn"
	if merged.Rules["no-debugger"].Level != "warn" {
		t.Errorf("Expected no-debugger to be 'warn', got %q", merged.Rules["no-debugger"].Level)
	}
	// no-console should remain
	if merged.Rules["no-console"].Level != "error" {
		t.Errorf("Expected no-console to be 'error', got %q", merged.Rules["no-console"].Level)
	}
	// for-direction should be added
	if merged.Rules["for-direction"].Level != "error" {
		t.Errorf("Expected for-direction to be 'error', got %q", merged.Rules["for-direction"].Level)
	}
}

func TestGetConfigForFile_SettingsShallowMerge(t *testing.T) {
	config := RslintConfig{
		{
			Settings: Settings{
				"importResolver": "node",
				"react":          "17",
			},
		},
		{
			Settings: Settings{
				"react": "18",
			},
		},
	}

	merged := config.GetConfigForFile("src/app.ts")
	if merged == nil {
		t.Fatal("Expected non-nil config")
	}

	if merged.Settings["importResolver"] != "node" {
		t.Errorf("Expected importResolver to be 'node', got %v", merged.Settings["importResolver"])
	}
	if merged.Settings["react"] != "18" {
		t.Errorf("Expected react to be '18' (overridden), got %v", merged.Settings["react"])
	}
}

func TestMergeLanguageOptions(t *testing.T) {
	t.Run("nil override returns base", func(t *testing.T) {
		base := &LanguageOptions{
			ParserOptions: &ParserOptions{
				ProjectService: BoolPtr(true),
			},
		}
		result := mergeLanguageOptions(base, nil)
		if result != base {
			t.Error("Expected base to be returned when override is nil")
		}
	})

	t.Run("nil base returns override", func(t *testing.T) {
		override := &LanguageOptions{
			ParserOptions: &ParserOptions{
				ProjectService: BoolPtr(true),
			},
		}
		result := mergeLanguageOptions(nil, override)
		if result != override {
			t.Error("Expected override to be returned when base is nil")
		}
	})

	t.Run("deep merge parserOptions", func(t *testing.T) {
		base := &LanguageOptions{
			ParserOptions: &ParserOptions{
				ProjectService: BoolPtr(true),
			},
		}
		override := &LanguageOptions{
			ParserOptions: &ParserOptions{
				ProjectService: BoolPtr(false),
				Project:        ProjectPaths{"./tsconfig.json"},
			},
		}
		result := mergeLanguageOptions(base, override)

		if result.ParserOptions.ProjectService == nil || *result.ParserOptions.ProjectService != false {
			t.Error("Expected ProjectService to be overridden to false")
		}
		if len(result.ParserOptions.Project) != 1 || result.ParserOptions.Project[0] != "./tsconfig.json" {
			t.Error("Expected Project to be set from override")
		}
	})

	t.Run("nil ProjectService in override preserves base", func(t *testing.T) {
		base := &LanguageOptions{
			ParserOptions: &ParserOptions{
				ProjectService: BoolPtr(true),
			},
		}
		override := &LanguageOptions{
			ParserOptions: &ParserOptions{
				Project: ProjectPaths{"./tsconfig.json"},
			},
		}
		result := mergeLanguageOptions(base, override)

		if result.ParserOptions.ProjectService == nil || *result.ParserOptions.ProjectService != true {
			t.Error("Expected ProjectService to be preserved from base")
		}
	})
}

func TestIsGlobalIgnoreEntry(t *testing.T) {
	tests := []struct {
		name     string
		entry    ConfigEntry
		expected bool
	}{
		{
			name:     "only ignores",
			entry:    ConfigEntry{Ignores: []string{"dist/**"}},
			expected: true,
		},
		{
			name:     "ignores with rules",
			entry:    ConfigEntry{Ignores: []string{"dist/**"}, Rules: Rules{"no-debugger": "error"}},
			expected: false,
		},
		{
			name:     "ignores with files",
			entry:    ConfigEntry{Ignores: []string{"dist/**"}, Files: []string{"**/*.ts"}},
			expected: false,
		},
		{
			name:     "ignores with plugins",
			entry:    ConfigEntry{Ignores: []string{"dist/**"}, Plugins: []string{"@typescript-eslint"}},
			expected: false,
		},
		{
			name:     "ignores with languageOptions",
			entry:    ConfigEntry{Ignores: []string{"dist/**"}, LanguageOptions: &LanguageOptions{}},
			expected: false,
		},
		{
			name:     "ignores with settings",
			entry:    ConfigEntry{Ignores: []string{"dist/**"}, Settings: Settings{"key": "val"}},
			expected: false,
		},
		{
			name:     "empty entry",
			entry:    ConfigEntry{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isGlobalIgnoreEntry(tt.entry)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestGetPluginRules(t *testing.T) {
	RegisterAllRules()

	tsRules := GetPluginRules("@typescript-eslint")
	if len(tsRules) == 0 {
		t.Error("Expected at least one TS rule")
	}

	// Verify we only get TS-ESLint rules, not core or import rules.
	// GetPluginRules filters by registry key prefix, so the returned count
	// should be less than all rules.
	allRules := GlobalRuleRegistry.GetAllRules()
	if len(tsRules) >= len(allRules) {
		t.Errorf("Expected fewer plugin rules than total rules, got %d vs %d", len(tsRules), len(allRules))
	}
}

func TestGetCoreRules(t *testing.T) {
	RegisterAllRules()

	coreRules := GetCoreRules()
	if len(coreRules) == 0 {
		t.Error("Expected at least one core rule")
	}

	// Core rules are registered with keys that don't contain "/"
	// The total should be much less than all rules
	allRules := GlobalRuleRegistry.GetAllRules()
	if len(coreRules) >= len(allRules) {
		t.Errorf("Expected fewer core rules than total, got %d vs %d", len(coreRules), len(allRules))
	}
}

func TestGetPluginRules_Disjoint(t *testing.T) {
	RegisterAllRules()

	tsRules := GetPluginRules("@typescript-eslint")
	coreRules := GetCoreRules()
	importRules := GetPluginRules("import")

	// Together they should cover all registered rules
	total := len(tsRules) + len(coreRules) + len(importRules)
	allRules := GlobalRuleRegistry.GetAllRules()
	if total != len(allRules) {
		t.Errorf("Expected total %d to equal all rules %d", total, len(allRules))
	}
}

func TestGetConfigForFile_ArrayRuleConfig(t *testing.T) {
	config := RslintConfig{
		{
			Rules: Rules{
				"array-type": []interface{}{"warn", map[string]interface{}{"default": "array-simple"}},
			},
		},
	}

	merged := config.GetConfigForFile("src/app.ts")
	if merged == nil {
		t.Fatal("Expected non-nil config")
	}

	rc := merged.Rules["array-type"]
	if rc == nil {
		t.Fatal("Expected array-type rule to be present")
	}
	if rc.Level != "warn" {
		t.Errorf("Expected level 'warn', got %q", rc.Level)
	}
	if rc.Options == nil || rc.Options["default"] != "array-simple" {
		t.Error("Expected options to contain default: array-simple")
	}
}

func TestGetConfigForFile_RuleOff(t *testing.T) {
	config := RslintConfig{
		{
			Rules: Rules{
				"no-debugger": "error",
			},
		},
		{
			Rules: Rules{
				"no-debugger": "off",
			},
		},
	}

	merged := config.GetConfigForFile("src/app.ts")
	if merged == nil {
		t.Fatal("Expected non-nil config")
	}

	rc := merged.Rules["no-debugger"]
	if rc == nil {
		t.Fatal("Expected no-debugger rule config to be present")
	}
	if rc.IsEnabled() {
		t.Error("Expected no-debugger to be disabled after being turned off")
	}
}

func TestGetConfigForFile_MultipleEntries_LanguageOptionsMerge(t *testing.T) {
	config := RslintConfig{
		{
			Files: []string{"**/*.ts"},
			LanguageOptions: &LanguageOptions{
				ParserOptions: &ParserOptions{
					ProjectService: BoolPtr(true),
				},
			},
			Rules: Rules{"no-debugger": "error"},
		},
		{
			Files: []string{"**/*.ts"},
			LanguageOptions: &LanguageOptions{
				ParserOptions: &ParserOptions{
					ProjectService: BoolPtr(false),
					Project:        ProjectPaths{"./tsconfig.json"},
				},
			},
		},
	}

	merged := config.GetConfigForFile("src/app.ts")
	if merged == nil {
		t.Fatal("Expected non-nil config")
	}

	if merged.LanguageOptions == nil || merged.LanguageOptions.ParserOptions == nil {
		t.Fatal("Expected languageOptions with parserOptions")
	}
	if merged.LanguageOptions.ParserOptions.ProjectService == nil || *merged.LanguageOptions.ParserOptions.ProjectService != false {
		t.Error("Expected projectService to be overridden to false")
	}
	if len(merged.LanguageOptions.ParserOptions.Project) != 1 {
		t.Error("Expected project to be set")
	}
}

func TestNormalizeJSONConfig_IgnoredFilesNotLinted(t *testing.T) {
	RegisterAllRules()

	config := normalizeJSONConfig(RslintConfig{
		{
			Ignores: []string{"**/*.test.ts"},
			Rules: Rules{
				"no-template-curly-in-string": "off",
			},
			Plugins: []string{"@typescript-eslint"},
		},
	})

	// Ignored file should return nil (not linted)
	merged := config.GetConfigForFile("src/app.test.ts")
	if merged != nil {
		t.Error("Expected nil for ignored file — should not be linted at all")
	}

	// Non-ignored file should have the "off" rule honored
	merged = config.GetConfigForFile("src/app.ts")
	if merged == nil {
		t.Fatal("Expected non-nil for non-ignored file")
	}
	rc := merged.Rules["no-template-curly-in-string"]
	if rc == nil {
		t.Fatal("Expected no-template-curly-in-string in merged rules")
	}
	if rc.IsEnabled() {
		t.Error("Expected no-template-curly-in-string to be disabled (user set 'off')")
	}
}

func TestNormalizeJSONConfig_SkipsGlobalIgnoreEntries(t *testing.T) {
	RegisterAllRules()

	config := normalizeJSONConfig(RslintConfig{
		{
			Ignores: []string{"dist/**"},
		},
		{
			Rules: Rules{},
		},
	})

	// Global ignore entry should not get rules injected
	if len(config[0].Rules) != 0 {
		t.Errorf("Expected global ignore entry to have 0 rules, got %d", len(config[0].Rules))
	}

	// Regular entry should get core rules
	if len(config[1].Rules) == 0 {
		t.Error("Expected regular entry to have injected core rules")
	}
}

func TestGetConfigForFile_ArrayRuleOff(t *testing.T) {
	config := RslintConfig{
		{
			Rules: Rules{
				"no-debugger": []interface{}{"off"},
			},
		},
	}

	merged := config.GetConfigForFile("src/app.ts")
	if merged == nil {
		t.Fatal("Expected non-nil config")
	}

	rc := merged.Rules["no-debugger"]
	if rc == nil {
		t.Fatal("Expected no-debugger rule config to be present")
	}
	if rc.IsEnabled() {
		t.Error("Expected no-debugger to be disabled via [\"off\"] array syntax")
	}
}

func TestGetConfigForFile_EntryIgnores_NoFiles(t *testing.T) {
	// Entry with ignores but no files — applies to all files except ignored ones
	config := RslintConfig{
		{
			Ignores: []string{"**/*.test.ts"},
			Rules:   Rules{"no-debugger": "error"},
		},
	}

	// Non-ignored file should match
	merged := config.GetConfigForFile("src/app.ts")
	if merged == nil {
		t.Fatal("Expected non-nil for non-ignored file")
	}
	if _, ok := merged.Rules["no-debugger"]; !ok {
		t.Error("Expected no-debugger for non-ignored file")
	}

	// Ignored file — no entry matches, return nil
	merged = config.GetConfigForFile("src/app.test.ts")
	if merged != nil {
		t.Error("Expected nil for ignored file with no other matching entry")
	}
}

func TestGetConfigForFile_EmptyConfig(t *testing.T) {
	config := RslintConfig{}

	merged := config.GetConfigForFile("src/app.ts")
	if merged != nil {
		t.Error("Expected nil for empty config (no entries)")
	}
}

func TestGetEnabledRules_FiltersByEnabledState(t *testing.T) {
	RegisterAllRules()

	config := RslintConfig{
		{
			Rules: Rules{
				"no-debugger":   "error",
				"no-console":    "warn",
				"for-direction": "off",
			},
		},
	}

	rules, mergedConfig := GlobalRuleRegistry.GetEnabledRules(config, "src/app.ts")
	if mergedConfig == nil {
		t.Fatal("Expected non-nil merged config")
	}

	// Build lookup for returned enabled rules
	ruleMap := make(map[string]linter.ConfiguredRule)
	for _, r := range rules {
		ruleMap[r.Name] = r
	}

	// "error" rule should be included with error severity
	if r, ok := ruleMap["no-debugger"]; !ok {
		t.Error("Expected no-debugger to be enabled")
	} else if r.Severity != rule.SeverityError {
		t.Errorf("Expected no-debugger severity to be error, got %v", r.Severity)
	}

	// "warn" rule should be included with warning severity
	if r, ok := ruleMap["no-console"]; !ok {
		t.Error("Expected no-console to be enabled")
	} else if r.Severity != rule.SeverityWarning {
		t.Errorf("Expected no-console severity to be warning, got %v", r.Severity)
	}

	// "off" rule should NOT be included
	if _, ok := ruleMap["for-direction"]; ok {
		t.Error("Expected for-direction to be excluded (set to off)")
	}
}

func TestGetEnabledRules_IgnoredFileReturnsNil(t *testing.T) {
	config := RslintConfig{
		{
			Ignores: []string{"dist/**"},
		},
		{
			Rules: Rules{"no-debugger": "error"},
		},
	}

	rules, mergedConfig := GlobalRuleRegistry.GetEnabledRules(config, "dist/bundle.js")
	if rules != nil {
		t.Error("Expected nil rules for ignored file")
	}
	if mergedConfig != nil {
		t.Error("Expected nil merged config for ignored file")
	}
}

func TestNormalizeJSONConfig_ArrayUserRuleTakesPrecedence(t *testing.T) {
	RegisterAllRules()

	config := normalizeJSONConfig(RslintConfig{
		{
			Rules: Rules{
				"no-console": []interface{}{"warn", map[string]interface{}{"allow": []interface{}{"error"}}},
			},
		},
	})

	merged := config.GetConfigForFile("src/app.ts")
	if merged == nil {
		t.Fatal("Expected non-nil config")
	}

	rc := merged.Rules["no-console"]
	if rc == nil {
		t.Fatal("Expected no-console rule to be present")
	}
	// User's ["warn", {allow: ["error"]}] should take precedence over auto-enabled "error"
	if rc.Level != "warn" {
		t.Errorf("Expected level 'warn' from user config, got %q", rc.Level)
	}
	if rc.Options == nil {
		t.Fatal("Expected options to be set")
	}
}

func TestGetConfigForFile_MultipleEntries_DifferentFilesPatterns(t *testing.T) {
	config := RslintConfig{
		{
			Files: []string{"**/*.ts"},
			Rules: Rules{"no-debugger": "error"},
		},
		{
			Files: []string{"**/*.js"},
			Rules: Rules{"no-console": "warn"},
		},
	}

	// .ts file: only entry1 matches
	merged := config.GetConfigForFile("src/app.ts")
	if merged == nil {
		t.Fatal("Expected non-nil for .ts file")
	}
	if _, ok := merged.Rules["no-debugger"]; !ok {
		t.Error("Expected no-debugger from entry1")
	}
	if _, ok := merged.Rules["no-console"]; ok {
		t.Error("Expected no-console to not be present (entry2 doesn't match .ts)")
	}

	// .js file: only entry2 matches
	merged = config.GetConfigForFile("src/app.js")
	if merged == nil {
		t.Fatal("Expected non-nil for .js file")
	}
	if _, ok := merged.Rules["no-console"]; !ok {
		t.Error("Expected no-console from entry2")
	}
	if _, ok := merged.Rules["no-debugger"]; ok {
		t.Error("Expected no-debugger to not be present (entry1 doesn't match .js)")
	}

	// .vue file: no entry matches → nil
	merged = config.GetConfigForFile("src/app.vue")
	if merged != nil {
		t.Error("Expected nil for .vue file (no entry matches)")
	}
}

func TestGetConfigForFile_MultipleEntries_PartialMatch(t *testing.T) {
	// entry1: only TS files; entry2: only Vue files; entry3: all files (no files pattern)
	config := RslintConfig{
		{
			Files: []string{"**/*.ts"},
			Rules: Rules{"no-debugger": "error"},
		},
		{
			Files: []string{"**/*.vue"},
			Rules: Rules{"no-console": "warn"},
		},
		{
			// No files → applies to all
			Rules: Rules{"for-direction": "error"},
		},
	}

	// .ts file: matches entry1 + entry3
	merged := config.GetConfigForFile("src/app.ts")
	if merged == nil {
		t.Fatal("Expected non-nil for .ts file")
	}
	if _, ok := merged.Rules["no-debugger"]; !ok {
		t.Error("Expected no-debugger from entry1")
	}
	if _, ok := merged.Rules["for-direction"]; !ok {
		t.Error("Expected for-direction from entry3")
	}
	if _, ok := merged.Rules["no-console"]; ok {
		t.Error("Expected no-console to not be present (entry2 doesn't match .ts)")
	}

	// .vue file: matches entry2 + entry3
	merged = config.GetConfigForFile("src/app.vue")
	if merged == nil {
		t.Fatal("Expected non-nil for .vue file")
	}
	if _, ok := merged.Rules["no-console"]; !ok {
		t.Error("Expected no-console from entry2")
	}
	if _, ok := merged.Rules["for-direction"]; !ok {
		t.Error("Expected for-direction from entry3")
	}
	if _, ok := merged.Rules["no-debugger"]; ok {
		t.Error("Expected no-debugger to not be present (entry1 doesn't match .vue)")
	}
}

func TestGetConfigForFile_ThreeEntries_CascadingOverride(t *testing.T) {
	config := RslintConfig{
		{
			Rules: Rules{
				"no-debugger": "error",
				"no-console":  "error",
			},
		},
		{
			// Override no-debugger to warn, add for-direction
			Rules: Rules{
				"no-debugger":   "warn",
				"for-direction": "error",
			},
		},
		{
			// Turn off for-direction
			Rules: Rules{
				"for-direction": "off",
			},
		},
	}

	merged := config.GetConfigForFile("src/app.ts")
	if merged == nil {
		t.Fatal("Expected non-nil config")
	}

	// no-debugger: entry1 "error" → entry2 "warn" → final "warn"
	if merged.Rules["no-debugger"].Level != "warn" {
		t.Errorf("Expected no-debugger 'warn', got %q", merged.Rules["no-debugger"].Level)
	}
	// no-console: entry1 "error", never overridden → final "error"
	if merged.Rules["no-console"].Level != "error" {
		t.Errorf("Expected no-console 'error', got %q", merged.Rules["no-console"].Level)
	}
	// for-direction: entry2 "error" → entry3 "off" → final "off"
	if merged.Rules["for-direction"].IsEnabled() {
		t.Error("Expected for-direction to be disabled (turned off in entry3)")
	}
}

func TestGetConfigForFile_MultipleEntries_ArrayRuleOverridesString(t *testing.T) {
	config := RslintConfig{
		{
			Rules: Rules{
				"no-console": "error",
			},
		},
		{
			// Later entry overrides string config with array config
			Rules: Rules{
				"no-console": []interface{}{"warn", map[string]interface{}{"allow": []interface{}{"error", "warn"}}},
			},
		},
	}

	merged := config.GetConfigForFile("src/app.ts")
	if merged == nil {
		t.Fatal("Expected non-nil config")
	}

	rc := merged.Rules["no-console"]
	if rc == nil {
		t.Fatal("Expected no-console in merged rules")
	}
	if rc.Level != "warn" {
		t.Errorf("Expected level 'warn' from array override, got %q", rc.Level)
	}
	if rc.Options == nil {
		t.Fatal("Expected options from array config")
	}
	allow, ok := rc.Options["allow"].([]interface{})
	if !ok || len(allow) != 2 {
		t.Error("Expected allow option with 2 items")
	}
}

func TestGetConfigForFile_GlobalIgnore_PlusEntryIgnores(t *testing.T) {
	config := RslintConfig{
		{
			// Global ignore for dist
			Ignores: []string{"dist/**"},
		},
		{
			// Entry with its own ignores for test files
			Ignores: []string{"**/*.test.ts"},
			Rules:   Rules{"no-debugger": "error"},
		},
	}

	// File in dist: global ignore → nil
	merged := config.GetConfigForFile("dist/bundle.js")
	if merged != nil {
		t.Error("Expected nil for dist file (global ignore)")
	}

	// Test file: entry-level ignore, no other entry matches → nil
	merged = config.GetConfigForFile("src/app.test.ts")
	if merged != nil {
		t.Error("Expected nil for test file (entry-level ignore, no other match)")
	}

	// Normal file: not ignored anywhere, entry2 matches
	merged = config.GetConfigForFile("src/app.ts")
	if merged == nil {
		t.Fatal("Expected non-nil for normal file")
	}
	if _, ok := merged.Rules["no-debugger"]; !ok {
		t.Error("Expected no-debugger from entry2")
	}
}

func TestNormalizeJSONConfig_MultipleEntries_DifferentPlugins(t *testing.T) {
	RegisterAllRules()

	config := normalizeJSONConfig(RslintConfig{
		{
			Files:   []string{"**/*.ts"},
			Plugins: []string{"@typescript-eslint"},
			Rules:   Rules{},
		},
		{
			Files: []string{"**/*.js"},
			Rules: Rules{},
		},
	})

	// Entry1 should have TS plugin rules + core rules
	if _, exists := config[0].Rules["@typescript-eslint/no-explicit-any"]; !exists {
		t.Error("Expected TS plugin rule in entry1")
	}
	if _, exists := config[0].Rules["no-debugger"]; !exists {
		t.Error("Expected core rule in entry1")
	}

	// Entry2 should have core rules only (no plugins)
	if _, exists := config[1].Rules["@typescript-eslint/no-explicit-any"]; exists {
		t.Error("Expected no TS plugin rule in entry2 (no plugin declared)")
	}
	if _, exists := config[1].Rules["no-debugger"]; !exists {
		t.Error("Expected core rule in entry2")
	}

	// Verify via GetConfigForFile: .ts file gets TS rules, .js file doesn't
	merged := config.GetConfigForFile("src/app.ts")
	if merged == nil {
		t.Fatal("Expected non-nil for .ts file")
	}
	if _, ok := merged.Rules["@typescript-eslint/no-explicit-any"]; !ok {
		t.Error("Expected TS rule for .ts file")
	}

	merged = config.GetConfigForFile("src/app.js")
	if merged == nil {
		t.Fatal("Expected non-nil for .js file")
	}
	if _, ok := merged.Rules["@typescript-eslint/no-explicit-any"]; ok {
		t.Error("Expected no TS rule for .js file")
	}
}

func TestNormalizeJSONConfig_IgnoresWithRulesOff(t *testing.T) {
	RegisterAllRules()

	// Simulates the real rslint.json: single entry with ignores + rules (including "off")
	config := normalizeJSONConfig(RslintConfig{
		{
			Ignores: []string{
				"packages/rslint-test-tools/tests/**/*.test.ts",
			},
			Rules: Rules{
				"no-template-curly-in-string": "off",
				"no-console":                  "warn",
			},
			Plugins: []string{"@typescript-eslint"},
		},
	})

	// Non-ignored file: all rules apply, "off" honored, plugin rules enabled
	merged := config.GetConfigForFile("src/app.ts")
	if merged == nil {
		t.Fatal("Expected non-nil for non-ignored file")
	}
	if merged.Rules["no-template-curly-in-string"].IsEnabled() {
		t.Error("Expected no-template-curly-in-string to be disabled")
	}
	if !merged.Rules["no-console"].IsEnabled() {
		t.Error("Expected no-console to be enabled")
	}
	if merged.Rules["no-console"].Level != "warn" {
		t.Errorf("Expected no-console level 'warn', got %q", merged.Rules["no-console"].Level)
	}
	if _, ok := merged.Rules["@typescript-eslint/no-explicit-any"]; !ok {
		t.Error("Expected TS plugin rules to be present")
	}

	// Ignored test file: should return nil (not linted at all)
	merged = config.GetConfigForFile("packages/rslint-test-tools/tests/some-rule.test.ts")
	if merged != nil {
		t.Error("Expected nil for ignored test file")
	}
}
