package config

import (
	"reflect"
	"strings"
	"testing"
)

// TestEslintPluginDeclNameAliases_PinnedForJSGuard pins the set of
// `eslint-plugin-*` declaration names so the JS collision guard
// (define-config.ts NATIVE_PLUGIN_DECL_ALIASES), a hand-maintained mirror of
// this set, cannot silently drift. If a newly-ported plugin adds such an alias,
// this fails — a prompt to mirror it on the JS side, else the gate would
// silently drop community plugins mounted under that key.
func TestEslintPluginDeclNameAliases_PinnedForJSGuard(t *testing.T) {
	want := map[string]struct{}{
		"eslint-plugin-import":      {},
		"eslint-plugin-jest":        {},
		"eslint-plugin-jsx-a11y":    {},
		"eslint-plugin-promise":     {},
		"eslint-plugin-react-hooks": {},
		"eslint-plugin-unicorn":     {},
	}
	got := map[string]struct{}{}
	for _, plugin := range nativePluginDeclarations {
		for _, declarationName := range plugin.declarationNames {
			if strings.HasPrefix(declarationName, "eslint-plugin-") {
				got[declarationName] = struct{}{}
			}
		}
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("eslint-plugin-* declaration names drifted:\n got  = %v\n want = %v\nMirror them in packages/rslint/src/config/define-config.ts NATIVE_PLUGIN_DECL_ALIASES.", got, want)
	}
}

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

func TestNativePluginDeclarationsCoverCatalogNamespaces(t *testing.T) {
	namespaces := make(map[string]bool, len(nativePluginDeclarations))
	for _, plugin := range nativePluginDeclarations {
		namespaces[plugin.rulePrefix] = false
	}
	hasCoreRule := false
	for ruleName := range nativeRuleCatalog().AllRules() {
		namespace := RulePluginPrefix(ruleName)
		if namespace == "" {
			hasCoreRule = true
			continue
		}
		if _, known := namespaces[namespace]; !known {
			t.Errorf("native rule %q uses undeclared plugin namespace %q", ruleName, namespace)
			continue
		}
		namespaces[namespace] = true
	}
	if !hasCoreRule {
		t.Error("native catalog contains no core rules")
	}
	for namespace, hasRule := range namespaces {
		if !hasRule {
			t.Errorf("native plugin namespace %q contains no rules", namespace)
		}
	}
}

func TestNativePluginDeclarationNamesAreIndexed(t *testing.T) {
	for _, plugin := range nativePluginDeclarations {
		for _, declarationName := range plugin.declarationNames {
			indexed, ok := nativePluginByDeclarationName[declarationName]
			if !ok {
				t.Errorf("declaration name %q is not indexed", declarationName)
				continue
			}
			if indexed.rulePrefix != plugin.rulePrefix {
				t.Errorf("declaration name %q resolved to namespace %q, want %q", declarationName, indexed.rulePrefix, plugin.rulePrefix)
			}
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
