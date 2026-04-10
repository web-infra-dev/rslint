package config

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/rule"
)

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

	rules, mergedConfig := GlobalRuleRegistry.GetEnabledRules(config, "src/app.ts", "", false)
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

	rules, mergedConfig := GlobalRuleRegistry.GetEnabledRules(config, "dist/bundle.js", "", false)
	if rules != nil {
		t.Error("Expected nil rules for ignored file")
	}
	if mergedConfig != nil {
		t.Error("Expected nil merged config for ignored file")
	}
}

// ======== Plugin enforcement tests (enforcePlugins=true, JS/TS config) ========

func TestGetEnabledRules_EnforcePlugins_BlocksUndeclaredPlugin(t *testing.T) {
	RegisterAllRules()

	// JS config: rules declared but plugin NOT declared
	config := RslintConfig{
		{
			Rules: Rules{
				"@typescript-eslint/no-explicit-any": "error",
				"no-debugger":                        "error",
			},
		},
	}

	rules, mergedConfig := GlobalRuleRegistry.GetEnabledRules(config, "src/app.ts", "", true)
	if mergedConfig == nil {
		t.Fatal("Expected non-nil merged config")
	}

	ruleMap := make(map[string]linter.ConfiguredRule)
	for _, r := range rules {
		ruleMap[r.Name] = r
	}

	// Core rule should always be active
	if _, ok := ruleMap["no-debugger"]; !ok {
		t.Error("Expected core rule no-debugger to be enabled regardless of plugins")
	}

	// Plugin rule should be gated: @typescript-eslint not declared
	if _, ok := ruleMap["@typescript-eslint/no-explicit-any"]; ok {
		t.Error("Expected @typescript-eslint/no-explicit-any to be blocked (plugin not declared)")
	}
}

func TestGetEnabledRules_EnforcePlugins_AllowsDeclaredPlugin(t *testing.T) {
	RegisterAllRules()

	// JS config: rules declared AND plugin declared
	config := RslintConfig{
		{
			Plugins: []string{"@typescript-eslint"},
			Rules: Rules{
				"@typescript-eslint/no-explicit-any": "error",
				"no-debugger":                        "error",
			},
		},
	}

	rules, mergedConfig := GlobalRuleRegistry.GetEnabledRules(config, "src/app.ts", "", true)
	if mergedConfig == nil {
		t.Fatal("Expected non-nil merged config")
	}

	ruleMap := make(map[string]linter.ConfiguredRule)
	for _, r := range rules {
		ruleMap[r.Name] = r
	}

	// Both should be active
	if _, ok := ruleMap["no-debugger"]; !ok {
		t.Error("Expected no-debugger to be enabled")
	}
	if _, ok := ruleMap["@typescript-eslint/no-explicit-any"]; !ok {
		t.Error("Expected @typescript-eslint/no-explicit-any to be enabled (plugin declared)")
	}
}

func TestGetEnabledRules_EnforcePlugins_EslintPluginPrefix(t *testing.T) {
	RegisterAllRules()

	// Plugin declared with eslint-plugin- prefix
	config := RslintConfig{
		{
			Plugins: []string{"eslint-plugin-import"},
			Rules: Rules{
				"import/no-self-import": "error",
			},
		},
	}

	rules, _ := GlobalRuleRegistry.GetEnabledRules(config, "src/app.ts", "", true)

	ruleMap := make(map[string]linter.ConfiguredRule)
	for _, r := range rules {
		ruleMap[r.Name] = r
	}

	// "eslint-plugin-import" normalizes to "import" → matches "import/" prefix
	if _, ok := ruleMap["import/no-self-import"]; !ok {
		t.Error("Expected import/no-self-import to be enabled (eslint-plugin-import declared)")
	}
}

func TestGetEnabledRules_EnforcePlugins_MultiplePlugins(t *testing.T) {
	RegisterAllRules()

	config := RslintConfig{
		{
			Plugins: []string{"@typescript-eslint"},
			Rules: Rules{
				"@typescript-eslint/no-explicit-any": "error",
				"react/jsx-uses-react":               "error", // react plugin NOT declared
			},
		},
	}

	rules, _ := GlobalRuleRegistry.GetEnabledRules(config, "src/app.tsx", "", true)

	ruleMap := make(map[string]linter.ConfiguredRule)
	for _, r := range rules {
		ruleMap[r.Name] = r
	}

	// @typescript-eslint declared → rule enabled
	if _, ok := ruleMap["@typescript-eslint/no-explicit-any"]; !ok {
		t.Error("Expected @typescript-eslint/no-explicit-any to be enabled")
	}

	// react NOT declared → rule blocked
	if _, ok := ruleMap["react/jsx-uses-react"]; ok {
		t.Error("Expected react/jsx-uses-react to be blocked (react plugin not declared)")
	}
}

func TestGetEnabledRules_NoEnforcePlugins_AllowsAll(t *testing.T) {
	RegisterAllRules()

	// JSON config behavior: enforcePlugins=false, no plugin gating
	config := RslintConfig{
		{
			Rules: Rules{
				"@typescript-eslint/no-explicit-any": "error",
				"no-debugger":                        "error",
			},
		},
	}

	rules, _ := GlobalRuleRegistry.GetEnabledRules(config, "src/app.ts", "", false)

	ruleMap := make(map[string]linter.ConfiguredRule)
	for _, r := range rules {
		ruleMap[r.Name] = r
	}

	// Without enforcement, plugin rules are allowed even without declaration
	if _, ok := ruleMap["@typescript-eslint/no-explicit-any"]; !ok {
		t.Error("Expected @typescript-eslint/no-explicit-any to be enabled (no enforcement)")
	}
	if _, ok := ruleMap["no-debugger"]; !ok {
		t.Error("Expected no-debugger to be enabled")
	}
}

func TestGetEnabledRules_EnforcePlugins_PluginFromDifferentEntry(t *testing.T) {
	RegisterAllRules()

	// Plugin declared in one entry, rule in another — both match the same file
	config := RslintConfig{
		{
			Plugins: []string{"@typescript-eslint"},
		},
		{
			Rules: Rules{
				"@typescript-eslint/no-explicit-any": "error",
			},
		},
	}

	rules, _ := GlobalRuleRegistry.GetEnabledRules(config, "src/app.ts", "", true)

	ruleMap := make(map[string]linter.ConfiguredRule)
	for _, r := range rules {
		ruleMap[r.Name] = r
	}

	// Plugin from entry1 + rule from entry2 = merged, should work
	if _, ok := ruleMap["@typescript-eslint/no-explicit-any"]; !ok {
		t.Error("Expected rule to be enabled when plugin is declared in a different matching entry")
	}
}

func TestGetEnabledRules_EnforcePlugins_PluginEntryDoesNotMatchFile(t *testing.T) {
	RegisterAllRules()

	// Plugin declared in entry that doesn't match the file
	config := RslintConfig{
		{
			Files:   []string{"**/*.jsx"},
			Plugins: []string{"@typescript-eslint"},
		},
		{
			Files: []string{"**/*.ts"},
			Rules: Rules{
				"@typescript-eslint/no-explicit-any": "error",
			},
		},
	}

	rules, _ := GlobalRuleRegistry.GetEnabledRules(config, "src/app.ts", "", true)

	ruleMap := make(map[string]linter.ConfiguredRule)
	for _, r := range rules {
		ruleMap[r.Name] = r
	}

	// Plugin entry doesn't match .ts files — only entry2 matches
	// Plugin not in merged set → rule blocked
	if _, ok := ruleMap["@typescript-eslint/no-explicit-any"]; ok {
		t.Error("Expected rule to be blocked when plugin is declared in non-matching entry")
	}
}

// Case 2: Preset-like spread + local override
func TestGetEnabledRules_EnforcePlugins_PresetPlusOverride(t *testing.T) {
	RegisterAllRules()

	// Simulates: [...ts.configs.recommended, { rules: { override } }]
	// Entry1 = preset (has plugins + rules), Entry2 = user override (no plugins)
	config := RslintConfig{
		{
			Files:   []string{"**/*.ts"},
			Plugins: []string{"@typescript-eslint"},
			Rules: Rules{
				"@typescript-eslint/no-explicit-any": "error",
				"@typescript-eslint/ban-ts-comment":  "error",
			},
		},
		{
			// User override: turn off one rule, keep the other
			Rules: Rules{
				"@typescript-eslint/no-explicit-any": "off",
			},
		},
	}

	rules, _ := GlobalRuleRegistry.GetEnabledRules(config, "src/app.ts", "", true)

	ruleMap := make(map[string]linter.ConfiguredRule)
	for _, r := range rules {
		ruleMap[r.Name] = r
	}

	// no-explicit-any overridden to "off" → should not be in enabled rules
	if _, ok := ruleMap["@typescript-eslint/no-explicit-any"]; ok {
		t.Error("Expected no-explicit-any to be disabled (overridden to off)")
	}

	// ban-ts-comment from preset → still enabled (plugin from entry1 applies)
	if _, ok := ruleMap["@typescript-eslint/ban-ts-comment"]; !ok {
		t.Error("Expected ban-ts-comment to remain enabled from preset")
	}
}

// Multiple plugins declared in the same entry array
func TestGetEnabledRules_EnforcePlugins_MultiplePluginsInSameEntry(t *testing.T) {
	RegisterAllRules()

	config := RslintConfig{
		{
			Plugins: []string{"@typescript-eslint", "react"},
			Rules: Rules{
				"@typescript-eslint/no-explicit-any": "error",
				"react/jsx-uses-react":               "error",
				"import/no-self-import":               "error", // import plugin NOT in plugins array
			},
		},
	}

	rules, _ := GlobalRuleRegistry.GetEnabledRules(config, "src/app.tsx", "", true)

	ruleMap := make(map[string]linter.ConfiguredRule)
	for _, r := range rules {
		ruleMap[r.Name] = r
	}

	if _, ok := ruleMap["@typescript-eslint/no-explicit-any"]; !ok {
		t.Error("Expected @typescript-eslint rule to be enabled")
	}
	if _, ok := ruleMap["react/jsx-uses-react"]; !ok {
		t.Error("Expected react rule to be enabled")
	}
	// import plugin not declared → blocked
	if _, ok := ruleMap["import/no-self-import"]; ok {
		t.Error("Expected import rule to be blocked (import plugin not declared)")
	}
}

// Case 4: Plugins declared in a LATER entry (reversed order from Case 3)
func TestGetEnabledRules_EnforcePlugins_PluginInLaterEntry(t *testing.T) {
	RegisterAllRules()

	config := RslintConfig{
		{
			// Rules first, no plugins
			Rules: Rules{
				"@typescript-eslint/no-explicit-any": "error",
			},
		},
		{
			// Plugins declared later
			Plugins: []string{"@typescript-eslint"},
		},
	}

	rules, _ := GlobalRuleRegistry.GetEnabledRules(config, "src/app.ts", "", true)

	ruleMap := make(map[string]linter.ConfiguredRule)
	for _, r := range rules {
		ruleMap[r.Name] = r
	}

	// Order shouldn't matter — plugins from any matching entry are merged
	if _, ok := ruleMap["@typescript-eslint/no-explicit-any"]; !ok {
		t.Error("Expected rule to be enabled even when plugin is declared in a later entry")
	}
}

// Case 7 complete: Multiple plugins from different entries, both declared
func TestGetEnabledRules_EnforcePlugins_MultiplePluginsBothDeclared(t *testing.T) {
	RegisterAllRules()

	config := RslintConfig{
		{
			Files:   []string{"**/*.tsx"},
			Plugins: []string{"@typescript-eslint"},
			Rules: Rules{
				"@typescript-eslint/no-explicit-any": "error",
			},
		},
		{
			Files:   []string{"**/*.tsx"},
			Plugins: []string{"react"},
			Rules: Rules{
				"react/jsx-uses-react": "error",
			},
		},
	}

	rules, _ := GlobalRuleRegistry.GetEnabledRules(config, "src/app.tsx", "", true)

	ruleMap := make(map[string]linter.ConfiguredRule)
	for _, r := range rules {
		ruleMap[r.Name] = r
	}

	// Both plugins declared in their respective matching entries → both rules enabled
	if _, ok := ruleMap["@typescript-eslint/no-explicit-any"]; !ok {
		t.Error("Expected @typescript-eslint rule to be enabled")
	}
	if _, ok := ruleMap["react/jsx-uses-react"]; !ok {
		t.Error("Expected react rule to be enabled")
	}
}

// Case 9: Preset + additional plugin in separate entry
func TestGetEnabledRules_EnforcePlugins_PresetPlusAdditionalPlugin(t *testing.T) {
	RegisterAllRules()

	// Simulates: [...ts.configs.recommended, { plugins: ['react'], rules: { react/... } }]
	config := RslintConfig{
		{
			Files:   []string{"**/*.tsx"},
			Plugins: []string{"@typescript-eslint"},
			Rules: Rules{
				"@typescript-eslint/no-explicit-any": "error",
			},
		},
		{
			Files:   []string{"**/*.tsx"},
			Plugins: []string{"react"},
			Rules: Rules{
				"react/jsx-uses-react":             "error",
				"@typescript-eslint/ban-ts-comment": "error", // TS rule in react entry
			},
		},
	}

	rules, _ := GlobalRuleRegistry.GetEnabledRules(config, "src/app.tsx", "", true)

	ruleMap := make(map[string]linter.ConfiguredRule)
	for _, r := range rules {
		ruleMap[r.Name] = r
	}

	// TS plugin from entry1, react plugin from entry2 → both in merged set
	if _, ok := ruleMap["@typescript-eslint/no-explicit-any"]; !ok {
		t.Error("Expected TS rule from entry1 to be enabled")
	}
	if _, ok := ruleMap["react/jsx-uses-react"]; !ok {
		t.Error("Expected react rule from entry2 to be enabled")
	}
	// TS rule declared in entry2 also works because TS plugin was merged from entry1
	if _, ok := ruleMap["@typescript-eslint/ban-ts-comment"]; !ok {
		t.Error("Expected TS rule in entry2 to be enabled (plugin from entry1 merged)")
	}
}

// Case 10: Entry-level ignores prevent plugins from being merged
func TestGetEnabledRules_EnforcePlugins_IgnoresPreventsPluginMerge(t *testing.T) {
	RegisterAllRules()

	config := RslintConfig{
		{
			Ignores: []string{"**/*.test.ts"},
			Plugins: []string{"@typescript-eslint"},
			Rules: Rules{
				"@typescript-eslint/no-explicit-any": "error",
			},
		},
	}

	// Non-ignored file: entry matches, plugin and rule both apply
	rules, _ := GlobalRuleRegistry.GetEnabledRules(config, "src/app.ts", "", true)
	ruleMap := make(map[string]linter.ConfiguredRule)
	for _, r := range rules {
		ruleMap[r.Name] = r
	}
	if _, ok := ruleMap["@typescript-eslint/no-explicit-any"]; !ok {
		t.Error("Expected rule to be enabled for non-ignored file")
	}

	// Ignored file: entry is skipped entirely, no config returned (nil)
	rules2, merged := GlobalRuleRegistry.GetEnabledRules(config, "src/app.test.ts", "", true)
	if merged != nil {
		t.Error("Expected nil merged config for ignored file")
	}
	if rules2 != nil {
		t.Error("Expected nil rules for ignored file")
	}
}

// Case 10b: Ignores in one entry, plugin+rule in another → test file still gets plugin from second entry
func TestGetEnabledRules_EnforcePlugins_IgnoresOnlyAffectsOwnEntry(t *testing.T) {
	RegisterAllRules()

	config := RslintConfig{
		{
			Files:   []string{"**/*.ts"},
			Ignores: []string{"**/*.test.ts"},
			Plugins: []string{"@typescript-eslint"},
		},
		{
			Files: []string{"**/*.test.ts"},
			Rules: Rules{
				"@typescript-eslint/no-explicit-any": "error",
				"no-debugger":                        "error",
			},
		},
	}

	// test.ts: entry1 ignores it (plugin not merged from entry1), entry2 matches (no plugin)
	rules, _ := GlobalRuleRegistry.GetEnabledRules(config, "src/app.test.ts", "", true)
	ruleMap := make(map[string]linter.ConfiguredRule)
	for _, r := range rules {
		ruleMap[r.Name] = r
	}

	// Plugin rule should be blocked (plugin from entry1 not merged due to ignores)
	if _, ok := ruleMap["@typescript-eslint/no-explicit-any"]; ok {
		t.Error("Expected TS rule to be blocked for test file (plugin entry ignores it)")
	}
	// Core rule should still work
	if _, ok := ruleMap["no-debugger"]; !ok {
		t.Error("Expected core rule to work for test file")
	}
}

func TestGetEnabledRules_EnforcePlugins_OffRuleNotBlocked(t *testing.T) {
	RegisterAllRules()

	// A rule set to "off" with no plugin declared should not appear in enabled rules
	// (it shouldn't appear regardless — this tests there's no false positive)
	config := RslintConfig{
		{
			Rules: Rules{
				"@typescript-eslint/no-explicit-any": "off",
			},
		},
	}

	rules, _ := GlobalRuleRegistry.GetEnabledRules(config, "src/app.ts", "", true)

	for _, r := range rules {
		if r.Name == "@typescript-eslint/no-explicit-any" {
			t.Error("Expected off rule to not appear in enabled rules")
		}
	}
}

func TestGetActiveRulesForFile_FiltersTypeAwareWhenNotInTypeInfoFiles(t *testing.T) {
	RegisterAllRules()

	cfg := RslintConfig{
		{
			Rules: Rules{
				"@typescript-eslint/require-await": "error", // type-aware
				"no-console":                       "error", // not type-aware
			},
		},
	}

	typeInfoFiles := map[string]struct{}{
		"src/covered.ts": {},
	}

	// File IN typeInfoFiles — both rules returned
	covered := GlobalRuleRegistry.GetActiveRulesForFile(cfg, "src/covered.ts", "", false, typeInfoFiles)
	if len(covered) != 2 {
		t.Fatalf("Expected 2 rules for covered file, got %d: %v", len(covered), ruleNames(covered))
	}
	coveredNames := ruleNameSet(covered)
	if !coveredNames["@typescript-eslint/require-await"] {
		t.Error("Expected require-await for covered file")
	}
	if !coveredNames["no-console"] {
		t.Error("Expected no-console for covered file")
	}

	// File NOT in typeInfoFiles — type-aware filtered, only non-type-aware remains
	uncovered := GlobalRuleRegistry.GetActiveRulesForFile(cfg, "src/uncovered.ts", "", false, typeInfoFiles)
	if len(uncovered) != 1 {
		t.Fatalf("Expected 1 rule for uncovered file (only non-type-aware), got %d: %v", len(uncovered), ruleNames(uncovered))
	}
	if uncovered[0].Name != "no-console" {
		t.Errorf("Expected only no-console for uncovered file, got %q", uncovered[0].Name)
	}
}

func TestGetActiveRulesForFile_NilTypeInfoFilesNoFiltering(t *testing.T) {
	RegisterAllRules()

	cfg := RslintConfig{
		{
			Rules: Rules{
				"@typescript-eslint/require-await": "error",
			},
		},
	}

	// nil typeInfoFiles → no filtering, all rules enabled
	rules := GlobalRuleRegistry.GetActiveRulesForFile(cfg, "src/any.ts", "", false, nil)
	found := false
	for _, r := range rules {
		if r.Name == "@typescript-eslint/require-await" {
			found = true
		}
	}
	if !found {
		t.Error("Expected require-await when typeInfoFiles is nil (no filtering)")
	}
}

func ruleNames(rules []linter.ConfiguredRule) []string {
	names := make([]string, len(rules))
	for i, r := range rules {
		names[i] = r.Name
	}
	return names
}

func ruleNameSet(rules []linter.ConfiguredRule) map[string]bool {
	set := make(map[string]bool, len(rules))
	for _, r := range rules {
		set[r.Name] = true
	}
	return set
}
