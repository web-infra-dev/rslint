package config

import (
	"testing"
)

func TestNormalizeJSONConfig_CoreRulesDefault(t *testing.T) {
	RegisterAllRules()

	config := normalizeJSONConfig(RslintConfig{
		{
			Rules: Rules{},
		},
	})

	merged := config.GetConfigForFile("src/app.ts", "")
	if merged == nil {
		t.Fatal("Expected non-nil merged config")
		return
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

	merged := config.GetConfigForFile("src/app.ts", "")
	if merged == nil {
		t.Fatal("Expected non-nil merged config")
		return
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

	merged := config.GetConfigForFile("src/app.ts", "")
	if merged == nil {
		t.Fatal("Expected non-nil merged config")
		return
	}

	// User's "off" should override the auto-enabled "error"
	rc := merged.Rules["no-debugger"]
	if rc == nil {
		t.Fatal("Expected no-debugger rule to be present")
		return
	}
	if rc.IsEnabled() {
		t.Error("Expected no-debugger to be disabled (user set 'off')")
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
	merged := config.GetConfigForFile("src/app.test.ts", "")
	if merged != nil {
		t.Error("Expected nil for ignored file — should not be linted at all")
	}

	// Non-ignored file should have the "off" rule honored
	merged = config.GetConfigForFile("src/app.ts", "")
	if merged == nil {
		t.Fatal("Expected non-nil for non-ignored file")
		return
	}
	rc := merged.Rules["no-template-curly-in-string"]
	if rc == nil {
		t.Fatal("Expected no-template-curly-in-string in merged rules")
		return
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

func TestNormalizeJSONConfig_ArrayUserRuleTakesPrecedence(t *testing.T) {
	RegisterAllRules()

	config := normalizeJSONConfig(RslintConfig{
		{
			Rules: Rules{
				"no-console": []interface{}{"warn", map[string]interface{}{"allow": []interface{}{"error"}}},
			},
		},
	})

	merged := config.GetConfigForFile("src/app.ts", "")
	if merged == nil {
		t.Fatal("Expected non-nil config")
		return
	}

	rc := merged.Rules["no-console"]
	if rc == nil {
		t.Fatal("Expected no-console rule to be present")
		return
	}
	// User's ["warn", {allow: ["error"]}] should take precedence over auto-enabled "error"
	if rc.Level != "warn" {
		t.Errorf("Expected level 'warn' from user config, got %q", rc.Level)
	}
	if rc.Options == nil {
		t.Fatal("Expected options to be set")
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
	merged := config.GetConfigForFile("src/app.ts", "")
	if merged == nil {
		t.Fatal("Expected non-nil for .ts file")
		return
	}
	if _, ok := merged.Rules["@typescript-eslint/no-explicit-any"]; !ok {
		t.Error("Expected TS rule for .ts file")
	}

	merged = config.GetConfigForFile("src/app.js", "")
	if merged == nil {
		t.Fatal("Expected non-nil for .js file")
		return
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
	merged := config.GetConfigForFile("src/app.ts", "")
	if merged == nil {
		t.Fatal("Expected non-nil for non-ignored file")
		return
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
	merged = config.GetConfigForFile("packages/rslint-test-tools/tests/some-rule.test.ts", "")
	if merged != nil {
		t.Error("Expected nil for ignored test file")
	}
}

func TestNormalizeJSONConfig_EslintPluginImport(t *testing.T) {
	RegisterAllRules()

	// JSON config using "eslint-plugin-import" declaration name.
	// normalizeJSONConfig must normalize plugin name before calling GetPluginRules,
	// so that "import/" prefixed rules are correctly injected.
	config := normalizeJSONConfig(RslintConfig{
		{
			Plugins: []string{"eslint-plugin-import"},
			Rules:   Rules{},
		},
	})

	merged := config.GetConfigForFile("src/app.ts", "")
	if merged == nil {
		t.Fatal("Expected non-nil merged config")
		return
	}

	// Import plugin rules should be injected (e.g. import/no-self-import)
	foundImportRule := false
	for name := range merged.Rules {
		if len(name) > 7 && name[:7] == "import/" {
			foundImportRule = true
			break
		}
	}
	if !foundImportRule {
		t.Error("Expected import/ rules to be injected when plugins contains 'eslint-plugin-import'")
	}
}

func TestNormalizeJSONConfig_PluginNameAlias(t *testing.T) {
	RegisterAllRules()

	// "import" is an alias for "eslint-plugin-import" in KnownPlugins.
	// Both should produce the same result.
	config1 := normalizeJSONConfig(RslintConfig{
		{Plugins: []string{"eslint-plugin-import"}, Rules: Rules{}},
	})
	config2 := normalizeJSONConfig(RslintConfig{
		{Plugins: []string{"import"}, Rules: Rules{}},
	})

	merged1 := config1.GetConfigForFile("src/app.ts", "")
	merged2 := config2.GetConfigForFile("src/app.ts", "")

	if merged1 == nil || merged2 == nil {
		t.Fatal("Expected non-nil configs")
		return
	}

	// Count import/ rules in each — should be identical
	count1, count2 := 0, 0
	for name := range merged1.Rules {
		if len(name) > 7 && name[:7] == "import/" {
			count1++
		}
	}
	for name := range merged2.Rules {
		if len(name) > 7 && name[:7] == "import/" {
			count2++
		}
	}

	if count1 == 0 {
		t.Error("Expected import/ rules for 'eslint-plugin-import'")
	}
	if count1 != count2 {
		t.Errorf("Expected same import rule count for both aliases, got %d vs %d", count1, count2)
	}
}
