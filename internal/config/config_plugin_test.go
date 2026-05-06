package config

import (
	"strings"
	"testing"
)

func TestNormalizePluginName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Known plugins: declaration name → rule prefix
		{"@typescript-eslint", "@typescript-eslint"},
		{"eslint-plugin-import", "import"},
		{"import", "import"},
		{"react", "react"},
		// Unknown plugins: returned as-is
		{"eslint-plugin-react", "eslint-plugin-react"},
		{"custom-plugin", "custom-plugin"},
	}

	for _, tt := range tests {
		result := NormalizePluginName(tt.input)
		if result != tt.expected {
			t.Errorf("NormalizePluginName(%q) = %q, expected %q", tt.input, result, tt.expected)
		}
	}
}

func TestRulePluginPrefix(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"@typescript-eslint/no-explicit-any", "@typescript-eslint"},
		{"import/no-unresolved", "import"},
		{"react/jsx-uses-react", "react"},
		{"no-debugger", ""},
		{"for-direction", ""},
	}

	for _, tt := range tests {
		result := RulePluginPrefix(tt.input)
		if result != tt.expected {
			t.Errorf("RulePluginPrefix(%q) = %q, expected %q", tt.input, result, tt.expected)
		}
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

	// Sum across core + every known plugin, so this test does not need to be
	// edited each time a new plugin is added.
	total := len(GetCoreRules())
	for _, plugin := range KnownPlugins {
		total += len(GetPluginRules(plugin.RulePrefix))
	}

	allRules := GlobalRuleRegistry.GetAllRules()
	if total != len(allRules) {
		t.Errorf("Expected total %d to equal all rules %d", total, len(allRules))
	}
}

func TestPluginByDeclName_AllDeclNamesRegistered(t *testing.T) {
	// Every DeclName in KnownPlugins must resolve to the correct RulePrefix
	for _, plugin := range KnownPlugins {
		for _, declName := range plugin.DeclNames {
			info, ok := pluginByDeclName[declName]
			if !ok {
				t.Errorf("DeclName %q not found in pluginByDeclName lookup table", declName)
				continue
			}
			if info.RulePrefix != plugin.RulePrefix {
				t.Errorf("DeclName %q resolved to RulePrefix %q, expected %q", declName, info.RulePrefix, plugin.RulePrefix)
			}
		}
	}
}

func TestKnownPlugins_GetAllRulesMatchRulePrefix(t *testing.T) {
	RegisterAllRules()

	for _, plugin := range KnownPlugins {
		rules := plugin.getAllRules()
		if len(rules) == 0 {
			t.Errorf("Plugin %q getAllRules() returned 0 rules", plugin.RulePrefix)
			continue
		}
		prefix := plugin.RulePrefix + "/"
		for _, r := range rules {
			if !strings.HasPrefix(r.Name, prefix) {
				t.Errorf("Plugin %q getAllRules() returned rule %q which does not have prefix %q", plugin.RulePrefix, r.Name, prefix)
			}
		}
	}
}

func TestKnownPlugins_GetAllRulesConsistentWithGetPluginRules(t *testing.T) {
	RegisterAllRules()

	for _, plugin := range KnownPlugins {
		fromGetAll := plugin.getAllRules()
		fromRegistry := GetPluginRules(plugin.RulePrefix)
		if len(fromGetAll) != len(fromRegistry) {
			t.Errorf("Plugin %q: getAllRules() returned %d rules but GetPluginRules returned %d",
				plugin.RulePrefix, len(fromGetAll), len(fromRegistry))
		}
	}
}

func TestGetConfigForFile_MergesPlugins(t *testing.T) {
	config := RslintConfig{
		{
			Plugins: []string{"@typescript-eslint"},
			Rules:   Rules{"@typescript-eslint/no-explicit-any": "error"},
		},
		{
			Plugins: []string{"react"},
			Rules:   Rules{"react/jsx-uses-react": "error"},
		},
	}

	merged := config.GetConfigForFile("src/app.tsx", "")
	if merged == nil {
		t.Fatal("Expected non-nil config")
		return
	}

	// Both plugins should be merged
	if _, ok := merged.Plugins["@typescript-eslint"]; !ok {
		t.Error("Expected @typescript-eslint in merged plugins")
	}
	if _, ok := merged.Plugins["react"]; !ok {
		t.Error("Expected react in merged plugins")
	}
}

func TestGetConfigForFile_NormalizesEslintPluginPrefix(t *testing.T) {
	config := RslintConfig{
		{
			Plugins: []string{"eslint-plugin-import"},
			Rules:   Rules{"import/no-unresolved": "error"},
		},
	}

	merged := config.GetConfigForFile("src/app.ts", "")
	if merged == nil {
		t.Fatal("Expected non-nil config")
		return
	}

	// "eslint-plugin-import" should be normalized to "import"
	if _, ok := merged.Plugins["import"]; !ok {
		t.Error("Expected 'import' in merged plugins (normalized from 'eslint-plugin-import')")
	}
}

func TestGetConfigForFile_PluginsOnlyFromMatchingEntries(t *testing.T) {
	config := RslintConfig{
		{
			Files:   []string{"**/*.ts"},
			Plugins: []string{"@typescript-eslint"},
			Rules:   Rules{"@typescript-eslint/no-explicit-any": "error"},
		},
		{
			Files:   []string{"**/*.jsx"},
			Plugins: []string{"react"},
			Rules:   Rules{"react/jsx-uses-react": "error"},
		},
	}

	// .ts file should only have @typescript-eslint plugin
	merged := config.GetConfigForFile("src/app.ts", "")
	if merged == nil {
		t.Fatal("Expected non-nil config for .ts file")
		return
	}
	if _, ok := merged.Plugins["@typescript-eslint"]; !ok {
		t.Error("Expected @typescript-eslint plugin for .ts file")
	}
	if _, ok := merged.Plugins["react"]; ok {
		t.Error("Expected no react plugin for .ts file")
	}

	// .jsx file should only have react plugin
	merged = config.GetConfigForFile("src/app.jsx", "")
	if merged == nil {
		t.Fatal("Expected non-nil config for .jsx file")
		return
	}
	if _, ok := merged.Plugins["react"]; !ok {
		t.Error("Expected react plugin for .jsx file")
	}
	if _, ok := merged.Plugins["@typescript-eslint"]; ok {
		t.Error("Expected no @typescript-eslint plugin for .jsx file")
	}
}

func TestGetConfigForFile_MultiplePluginsInSameEntry(t *testing.T) {
	RegisterAllRules()

	config := RslintConfig{
		{
			Plugins: []string{"@typescript-eslint", "react"},
			Rules: Rules{
				"@typescript-eslint/no-explicit-any": "error",
				"react/jsx-uses-react":               "error",
			},
		},
	}

	merged := config.GetConfigForFile("src/app.tsx", "")
	if merged == nil {
		t.Fatal("Expected non-nil config")
		return
	}

	if _, ok := merged.Plugins["@typescript-eslint"]; !ok {
		t.Error("Expected @typescript-eslint in merged plugins")
	}
	if _, ok := merged.Plugins["react"]; !ok {
		t.Error("Expected react in merged plugins")
	}
	if len(merged.Plugins) != 2 {
		t.Errorf("Expected exactly 2 plugins, got %d", len(merged.Plugins))
	}
}

func TestGetConfigForFile_DuplicatePluginInSameEntry(t *testing.T) {
	config := RslintConfig{
		{
			Plugins: []string{"@typescript-eslint", "@typescript-eslint"},
			Rules:   Rules{"@typescript-eslint/no-explicit-any": "error"},
		},
	}

	merged := config.GetConfigForFile("src/app.ts", "")
	if merged == nil {
		t.Fatal("Expected non-nil config")
		return
	}

	if _, ok := merged.Plugins["@typescript-eslint"]; !ok {
		t.Error("Expected @typescript-eslint in merged plugins")
	}
	// Duplicates should be deduplicated
	if len(merged.Plugins) != 1 {
		t.Errorf("Expected exactly 1 plugin after deduplication, got %d", len(merged.Plugins))
	}
}

func TestGetConfigForFile_SamePluginDifferentNamesAcrossEntries(t *testing.T) {
	config := RslintConfig{
		{
			Plugins: []string{"eslint-plugin-import"},
			Rules:   Rules{"import/no-self-import": "error"},
		},
		{
			// Same plugin but written without the eslint-plugin- prefix
			Plugins: []string{"import"},
		},
	}

	merged := config.GetConfigForFile("src/app.ts", "")
	if merged == nil {
		t.Fatal("Expected non-nil config")
		return
	}

	// Both normalize to "import", so should be deduplicated to 1
	if _, ok := merged.Plugins["import"]; !ok {
		t.Error("Expected 'import' in merged plugins")
	}
	if len(merged.Plugins) != 1 {
		t.Errorf("Expected 1 plugin after normalization, got %d", len(merged.Plugins))
	}
}

func TestGetConfigForFile_PluginsEntry_WithAndWithoutPlugins(t *testing.T) {
	config := RslintConfig{
		{
			Plugins: []string{"@typescript-eslint"},
			Rules:   Rules{"@typescript-eslint/no-explicit-any": "error"},
		},
		{
			// No plugins field at all
			Rules: Rules{"no-debugger": "error"},
		},
	}

	merged := config.GetConfigForFile("src/app.ts", "")
	if merged == nil {
		t.Fatal("Expected non-nil config")
		return
	}

	// Plugins from entry1 should be present
	if _, ok := merged.Plugins["@typescript-eslint"]; !ok {
		t.Error("Expected @typescript-eslint from entry1")
	}
	if len(merged.Plugins) != 1 {
		t.Errorf("Expected exactly 1 plugin, got %d", len(merged.Plugins))
	}
}

func TestGetConfigForFile_SamePluginAcrossEntries(t *testing.T) {
	config := RslintConfig{
		{
			Plugins: []string{"@typescript-eslint"},
			Rules:   Rules{"@typescript-eslint/no-explicit-any": "error"},
		},
		{
			Plugins: []string{"@typescript-eslint"},
			Rules:   Rules{"@typescript-eslint/ban-ts-comment": "error"},
		},
	}

	merged := config.GetConfigForFile("src/app.ts", "")
	if merged == nil {
		t.Fatal("Expected non-nil config")
		return
	}

	if _, ok := merged.Plugins["@typescript-eslint"]; !ok {
		t.Error("Expected @typescript-eslint in merged plugins")
	}
	if len(merged.Plugins) != 1 {
		t.Errorf("Expected 1 plugin after cross-entry deduplication, got %d", len(merged.Plugins))
	}
	// Both rules should be present
	if _, ok := merged.Rules["@typescript-eslint/no-explicit-any"]; !ok {
		t.Error("Expected no-explicit-any from entry1")
	}
	if _, ok := merged.Rules["@typescript-eslint/ban-ts-comment"]; !ok {
		t.Error("Expected ban-ts-comment from entry2")
	}
}

func TestGetConfigForFile_OverlappingPluginsAcrossEntries(t *testing.T) {
	config := RslintConfig{
		{
			Plugins: []string{"@typescript-eslint", "react"},
			Rules:   Rules{"@typescript-eslint/no-explicit-any": "error"},
		},
		{
			Plugins: []string{"react", "import"},
			Rules:   Rules{"import/no-self-import": "error"},
		},
	}

	merged := config.GetConfigForFile("src/app.tsx", "")
	if merged == nil {
		t.Fatal("Expected non-nil config")
		return
	}

	// Union: @typescript-eslint + react + import
	if _, ok := merged.Plugins["@typescript-eslint"]; !ok {
		t.Error("Expected @typescript-eslint")
	}
	if _, ok := merged.Plugins["react"]; !ok {
		t.Error("Expected react")
	}
	if _, ok := merged.Plugins["import"]; !ok {
		t.Error("Expected import")
	}
	if len(merged.Plugins) != 3 {
		t.Errorf("Expected 3 plugins (union with overlap), got %d", len(merged.Plugins))
	}
}

func TestGetConfigForFile_AllEntriesNoPlugins(t *testing.T) {
	config := RslintConfig{
		{
			Rules: Rules{"no-debugger": "error"},
		},
		{
			Rules: Rules{"no-console": "warn"},
		},
	}

	merged := config.GetConfigForFile("src/app.ts", "")
	if merged == nil {
		t.Fatal("Expected non-nil config")
		return
	}

	if len(merged.Plugins) != 0 {
		t.Errorf("Expected 0 plugins when none declared, got %d", len(merged.Plugins))
	}
}

func TestGetConfigForFile_EmptyPluginsArray(t *testing.T) {
	config := RslintConfig{
		{
			Plugins: []string{"@typescript-eslint"},
			Rules:   Rules{"@typescript-eslint/no-explicit-any": "error"},
		},
		{
			Plugins: []string{}, // explicitly empty
			Rules:   Rules{"no-debugger": "error"},
		},
	}

	merged := config.GetConfigForFile("src/app.ts", "")
	if merged == nil {
		t.Fatal("Expected non-nil config")
		return
	}

	// Only entry1's plugin should be present; empty array contributes nothing
	if _, ok := merged.Plugins["@typescript-eslint"]; !ok {
		t.Error("Expected @typescript-eslint from entry1")
	}
	if len(merged.Plugins) != 1 {
		t.Errorf("Expected 1 plugin, got %d", len(merged.Plugins))
	}
}

func TestGetConfigForFile_ThreeEntries_MixedPlugins(t *testing.T) {
	config := RslintConfig{
		{
			Plugins: []string{"@typescript-eslint"},
			Rules:   Rules{"@typescript-eslint/no-explicit-any": "error"},
		},
		{
			// No plugins
			Rules: Rules{"no-debugger": "error"},
		},
		{
			Plugins: []string{"react"},
			Rules:   Rules{"react/jsx-uses-react": "error"},
		},
	}

	merged := config.GetConfigForFile("src/app.tsx", "")
	if merged == nil {
		t.Fatal("Expected non-nil config")
		return
	}

	// Union of entry1 + entry3 plugins; entry2 contributes none
	if _, ok := merged.Plugins["@typescript-eslint"]; !ok {
		t.Error("Expected @typescript-eslint from entry1")
	}
	if _, ok := merged.Plugins["react"]; !ok {
		t.Error("Expected react from entry3")
	}
	if len(merged.Plugins) != 2 {
		t.Errorf("Expected 2 plugins, got %d", len(merged.Plugins))
	}

	// All 3 rules should be merged
	if len(merged.Rules) != 3 {
		t.Errorf("Expected 3 rules, got %d", len(merged.Rules))
	}
}
