package config

import (
	"testing"
)

func TestGetConfigForFile_ExplicitRulesOnly(t *testing.T) {
	config := RslintConfig{
		{
			Rules: Rules{
				"no-debugger": "error",
			},
		},
	}

	merged := config.GetConfigForFile("src/app.ts", "")
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

func TestGetConfigForFile_WithoutNormalize_PluginDoesNotAutoEnable(t *testing.T) {
	RegisterAllRules()

	// Without normalizeJSONConfig, plugins should not auto-enable rules
	config := RslintConfig{
		{
			Plugins: []string{"@typescript-eslint"},
		},
	}

	merged := config.GetConfigForFile("src/app.ts", "")
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
	merged := config.GetConfigForFile("dist/bundle.js", "")
	if merged != nil {
		t.Error("Expected nil for globally ignored file")
	}

	// File not in dist should not be ignored
	merged = config.GetConfigForFile("src/app.ts", "")
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
	merged := config.GetConfigForFile("src/app.test.ts", "")
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
	merged := config.GetConfigForFile("src/app.test.ts", "")
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
	merged := config.GetConfigForFile("src/app.ts", "")
	if merged == nil {
		t.Fatal("Expected non-nil config")
	}
	if _, ok := merged.Rules["no-debugger"]; !ok {
		t.Error("Expected no-debugger for matching .ts file")
	}

	// JS file should not match — no entry matches, return nil
	merged = config.GetConfigForFile("src/app.js", "")
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

	merged := config.GetConfigForFile("src/app.ts", "")
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

	merged := config.GetConfigForFile("src/app.ts", "")
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

func TestGetConfigForFile_ArrayRuleConfig(t *testing.T) {
	config := RslintConfig{
		{
			Rules: Rules{
				"array-type": []interface{}{"warn", map[string]interface{}{"default": "array-simple"}},
			},
		},
	}

	merged := config.GetConfigForFile("src/app.ts", "")
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

	merged := config.GetConfigForFile("src/app.ts", "")
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

	merged := config.GetConfigForFile("src/app.ts", "")
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

func TestGetConfigForFile_ArrayRuleOff(t *testing.T) {
	config := RslintConfig{
		{
			Rules: Rules{
				"no-debugger": []interface{}{"off"},
			},
		},
	}

	merged := config.GetConfigForFile("src/app.ts", "")
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
	merged := config.GetConfigForFile("src/app.ts", "")
	if merged == nil {
		t.Fatal("Expected non-nil for non-ignored file")
	}
	if _, ok := merged.Rules["no-debugger"]; !ok {
		t.Error("Expected no-debugger for non-ignored file")
	}

	// Ignored file — no entry matches, return nil
	merged = config.GetConfigForFile("src/app.test.ts", "")
	if merged != nil {
		t.Error("Expected nil for ignored file with no other matching entry")
	}
}

func TestGetConfigForFile_EmptyConfig(t *testing.T) {
	config := RslintConfig{}

	merged := config.GetConfigForFile("src/app.ts", "")
	if merged != nil {
		t.Error("Expected nil for empty config (no entries)")
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
	merged := config.GetConfigForFile("src/app.ts", "")
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
	merged = config.GetConfigForFile("src/app.js", "")
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
	merged = config.GetConfigForFile("src/app.vue", "")
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
	merged := config.GetConfigForFile("src/app.ts", "")
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
	merged = config.GetConfigForFile("src/app.vue", "")
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

	merged := config.GetConfigForFile("src/app.ts", "")
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

	merged := config.GetConfigForFile("src/app.ts", "")
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
	merged := config.GetConfigForFile("dist/bundle.js", "")
	if merged != nil {
		t.Error("Expected nil for dist file (global ignore)")
	}

	// Test file: entry-level ignore, no other entry matches → nil
	merged = config.GetConfigForFile("src/app.test.ts", "")
	if merged != nil {
		t.Error("Expected nil for test file (entry-level ignore, no other match)")
	}

	// Normal file: not ignored anywhere, entry2 matches
	merged = config.GetConfigForFile("src/app.ts", "")
	if merged == nil {
		t.Fatal("Expected non-nil for normal file")
	}
	if _, ok := merged.Rules["no-debugger"]; !ok {
		t.Error("Expected no-debugger from entry2")
	}
}

// TestGetConfigForFile_CwdAffectsMatching verifies that the cwd parameter
// controls how files/ignores patterns are matched against absolute file paths.
// This is critical for monorepo sub-package configs where the config directory
// differs from the process cwd.
func TestGetConfigForFile_CwdAffectsMatching(t *testing.T) {
	config := RslintConfig{
		{
			Files: []string{"src/**/*.ts"},
			Rules: Rules{"no-console": "error"},
		},
	}

	// Absolute path: /monorepo/packages/foo/src/index.ts
	absPath := "/monorepo/packages/foo/src/index.ts"

	// With cwd = config's own directory (/monorepo/packages/foo),
	// relative path = src/index.ts → matches src/**/*.ts ✓
	merged := config.GetConfigForFile(absPath, "/monorepo/packages/foo")
	if merged == nil {
		t.Fatal("Expected match when cwd is the config directory")
	}
	if merged.Rules["no-console"] == nil {
		t.Error("Expected no-console rule to be enabled")
	}

	// With cwd = monorepo root (/monorepo),
	// relative path = packages/foo/src/index.ts → does NOT match src/**/*.ts ✗
	merged = config.GetConfigForFile(absPath, "/monorepo")
	if merged != nil {
		t.Error("Expected no match when cwd is the monorepo root (wrong base for pattern)")
	}
}

// TestGetConfigForFile_CwdIgnoresMatching verifies cwd affects ignores resolution.
func TestGetConfigForFile_CwdIgnoresMatching(t *testing.T) {
	config := RslintConfig{
		{
			Ignores: []string{"dist/**"},
		},
		{
			Files: []string{"**/*.ts"},
			Rules: Rules{"no-console": "error"},
		},
	}

	absPath := "/project/dist/bundle.ts"

	// With cwd = /project, relative path = dist/bundle.ts → matches dist/** → globally ignored
	merged := config.GetConfigForFile(absPath, "/project")
	if merged != nil {
		t.Error("Expected file to be ignored when cwd matches config directory")
	}

	// With wrong cwd = /other, relative path won't start with dist/ → NOT ignored
	merged = config.GetConfigForFile(absPath, "/other")
	if merged == nil {
		t.Fatal("Expected file to NOT be ignored with wrong cwd")
	}
}

// TestGetConfigForFile_WindowsPaths verifies cwd matching works with Windows-style paths.
// uriToPath produces forward-slash paths (C:/Users/...) and os.Getwd may produce
// backslash paths (C:\Users\...). Both must compute correct relative paths.
func TestGetConfigForFile_WindowsPaths(t *testing.T) {
	cfg := RslintConfig{
		{
			Files: []string{"src/**/*.ts"},
			Rules: Rules{"no-console": "error"},
		},
	}

	tests := []struct {
		name     string
		filePath string
		cwd      string
		wantHit  bool
	}{
		{
			name:     "forward-slash cwd (from uriToPath)",
			filePath: "C:/Users/project/src/index.ts",
			cwd:      "C:/Users/project",
			wantHit:  true,
		},
		{
			name:     "backslash cwd (from os.Getwd on Windows)",
			filePath: "C:/Users/project/src/index.ts",
			cwd:      "C:\\Users\\project",
			wantHit:  true,
		},
		{
			name:     "monorepo sub-package Windows cwd",
			filePath: "C:/repo/packages/foo/src/index.ts",
			cwd:      "C:/repo/packages/foo",
			wantHit:  true,
		},
		{
			name:     "wrong cwd on Windows — should not match",
			filePath: "C:/repo/packages/foo/src/index.ts",
			cwd:      "C:/repo",
			wantHit:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			merged := cfg.GetConfigForFile(tt.filePath, tt.cwd)
			if tt.wantHit && merged == nil {
				t.Errorf("expected match for filePath=%q cwd=%q", tt.filePath, tt.cwd)
			}
			if !tt.wantHit && merged != nil {
				t.Errorf("expected no match for filePath=%q cwd=%q", tt.filePath, tt.cwd)
			}
		})
	}
}
