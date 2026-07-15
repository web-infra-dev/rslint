package config

import (
	"os"
	"testing"
)

func TestIsFileIgnored_Negation(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Unable to get working directory: %v", err)
	}

	tests := []struct {
		name         string
		filePath     string
		patterns     []string
		shouldIgnore bool
	}{
		{
			name:         "basic negation re-includes file",
			filePath:     "build/test.js",
			patterns:     []string{"build/**/*", "!build/test.js"},
			shouldIgnore: false,
		},
		{
			name:         "negation does not re-include non-matching files",
			filePath:     "build/other.js",
			patterns:     []string{"build/**/*", "!build/test.js"},
			shouldIgnore: true,
		},
		{
			name:         "negation with directory glob",
			filePath:     "vendor/keep/src/b.ts",
			patterns:     []string{"vendor/**/*", "!vendor/keep/**/*"},
			shouldIgnore: false,
		},
		{
			name:         "negation does not affect other directories",
			filePath:     "vendor/lib/src/a.ts",
			patterns:     []string{"vendor/**/*", "!vendor/keep/**/*"},
			shouldIgnore: true,
		},
		{
			name:         "only negation pattern does not ignore",
			filePath:     "src/index.ts",
			patterns:     []string{"!src/**"},
			shouldIgnore: false,
		},
		{
			name:         "later ignore overrides earlier negation",
			filePath:     "dist/a.js",
			patterns:     []string{"dist/**", "!dist/**", "dist/**"},
			shouldIgnore: true,
		},
		{
			name:         "negation at end re-includes",
			filePath:     "dist/a.js",
			patterns:     []string{"dist/**", "!dist/**"},
			shouldIgnore: false,
		},
		{
			name:         "exact file negation in nested directory",
			filePath:     "vendor/keep/config.ts",
			patterns:     []string{"vendor/**/*", "!vendor/keep/config.ts"},
			shouldIgnore: false,
		},
		{
			name:         "negation with **/ prefix pattern",
			filePath:     "packages/app/dist/index.js",
			patterns:     []string{"**/dist/**/*", "!**/dist/index.js"},
			shouldIgnore: false,
		},
		{
			name:         "negation with **/ prefix does not affect other files",
			filePath:     "packages/app/dist/other.js",
			patterns:     []string{"**/dist/**/*", "!**/dist/index.js"},
			shouldIgnore: true,
		},
		{
			name:         "multiple negation patterns first file",
			filePath:     "build/keep-a.js",
			patterns:     []string{"build/**/*", "!build/keep-a.js", "!build/keep-b.js"},
			shouldIgnore: false,
		},
		{
			name:         "multiple negation patterns second file",
			filePath:     "build/keep-b.js",
			patterns:     []string{"build/**/*", "!build/keep-a.js", "!build/keep-b.js"},
			shouldIgnore: false,
		},
		{
			name:         "multiple negation patterns non-matching file still ignored",
			filePath:     "build/other.js",
			patterns:     []string{"build/**/*", "!build/keep-a.js", "!build/keep-b.js"},
			shouldIgnore: true,
		},
		{
			name:         "non-matching negation pattern is harmless",
			filePath:     "dist/a.js",
			patterns:     []string{"dist/**/*", "!nonexistent/**"},
			shouldIgnore: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isFileIgnored(tt.filePath, ParseIgnorePatterns(tt.patterns), cwd)
			if result != tt.shouldIgnore {
				t.Errorf("isFileIgnored(%q, %v) = %v, expected %v",
					tt.filePath, tt.patterns, result, tt.shouldIgnore)
			}
		})
	}
}

func TestConfigIgnorePatternsUseMinimatchUTF16Units(t *testing.T) {
	if isFileIgnored("😀", ParseIgnorePatterns([]string{"?"}), "") {
		t.Fatal("one ? must not ignore an astral filename composed of two UTF-16 units")
	}
	if !isFileIgnored("😀", ParseIgnorePatterns([]string{"??"}), "") {
		t.Fatal("two ? wildcards must ignore one astral filename like Minimatch")
	}
}

func TestGetConfigForFile_NegationInGlobalIgnore(t *testing.T) {
	config := RslintConfig{
		{
			Ignores: []string{"build/**/*", "!build/test.js"},
		},
		{
			Files: []string{"**/*.js"},
			Rules: Rules{"no-debugger": "error"},
		},
	}

	merged := config.GetConfigForFile("build/test.js", "")
	if merged == nil {
		t.Fatal("Expected build/test.js to NOT be ignored (re-included by !)")
		return
	}
	if _, ok := merged.Rules["no-debugger"]; !ok {
		t.Error("Expected no-debugger rule for re-included file")
	}

	merged2 := config.GetConfigForFile("build/other.js", "")
	if merged2 != nil {
		t.Error("Expected build/other.js to be ignored")
	}

	merged3 := config.GetConfigForFile("src/index.js", "")
	if merged3 == nil {
		t.Fatal("Expected src/index.js to be linted")
		return
	}
}

func TestGetConfigForFile_NegationInEntryIgnore(t *testing.T) {
	config := RslintConfig{
		{
			Files:   []string{"**/*.ts"},
			Ignores: []string{"vendor/**/*", "!vendor/keep/**/*"},
			Rules:   Rules{"no-debugger": "error"},
		},
	}

	merged := config.GetConfigForFile("vendor/keep/src/b.ts", "")
	if merged == nil {
		t.Fatal("Expected vendor/keep/src/b.ts to NOT be ignored (re-included by !)")
		return
	}
	if _, ok := merged.Rules["no-debugger"]; !ok {
		t.Error("Expected no-debugger rule for re-included file")
	}

	merged2 := config.GetConfigForFile("vendor/lib/src/a.ts", "")
	if merged2 == nil || merged2.Rules["no-debugger"] != nil {
		t.Error("entry-local ignores must leave only the default empty config")
	}
}

func TestIsFileIgnored_NegationBeforePositive(t *testing.T) {
	cwd, _ := os.Getwd()

	// Negation before positive pattern: ! has nothing to negate yet,
	// then positive pattern ignores. Result: ignored.
	result := isFileIgnored("build/test.js", ParseIgnorePatterns([]string{"!build/test.js", "build/**"}), cwd)
	if !result {
		t.Error("Expected ignored: negation before positive has no effect, positive wins")
	}
}

func TestIsFileIgnored_FileExtensionNegation(t *testing.T) {
	cwd, _ := os.Getwd()

	tests := []struct {
		name         string
		filePath     string
		patterns     []string
		shouldIgnore bool
	}{
		{
			name:         "ignore all test files except integration tests",
			filePath:     "src/utils.integration.test.ts",
			patterns:     []string{"**/*.test.ts", "!**/*.integration.test.ts"},
			shouldIgnore: false,
		},
		{
			name:         "non-integration test file stays ignored",
			filePath:     "src/utils.unit.test.ts",
			patterns:     []string{"**/*.test.ts", "!**/*.integration.test.ts"},
			shouldIgnore: true,
		},
		{
			name:         "non-test file not affected",
			filePath:     "src/utils.ts",
			patterns:     []string{"**/*.test.ts", "!**/*.integration.test.ts"},
			shouldIgnore: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isFileIgnored(tt.filePath, ParseIgnorePatterns(tt.patterns), cwd)
			if result != tt.shouldIgnore {
				t.Errorf("isFileIgnored(%q, %v) = %v, expected %v",
					tt.filePath, tt.patterns, result, tt.shouldIgnore)
			}
		})
	}
}

func TestIsFileIgnored_MultiLevelNegateAndReIgnore(t *testing.T) {
	cwd, _ := os.Getwd()

	// Complex: ignore everything → re-include src → re-ignore src/test → re-include keep.ts
	patterns := []string{"**/*", "!src/**/*", "src/test/**/*", "!src/test/keep.ts"}

	tests := []struct {
		name         string
		filePath     string
		shouldIgnore bool
	}{
		{"root file ignored", "README.md", true},
		{"src file re-included", "src/index.ts", false},
		{"src/test file re-ignored", "src/test/utils.ts", true},
		{"src/test/keep.ts re-included", "src/test/keep.ts", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isFileIgnored(tt.filePath, ParseIgnorePatterns(patterns), cwd)
			if result != tt.shouldIgnore {
				t.Errorf("isFileIgnored(%q) = %v, expected %v", tt.filePath, result, tt.shouldIgnore)
			}
		})
	}
}

func TestIsFileIgnoredSimple_Negation(t *testing.T) {
	// isFileIgnoredSimple is the fallback when cwd is empty (JSON config path)
	tests := []struct {
		name         string
		filePath     string
		patterns     []string
		shouldIgnore bool
	}{
		{
			name:         "basic negation re-includes",
			filePath:     "build/test.js",
			patterns:     []string{"build/**/*", "!build/test.js"},
			shouldIgnore: false,
		},
		{
			name:         "non-negated file stays ignored",
			filePath:     "build/other.js",
			patterns:     []string{"build/**/*", "!build/test.js"},
			shouldIgnore: true,
		},
		{
			name:         "only negation does not ignore",
			filePath:     "src/a.ts",
			patterns:     []string{"!src/**"},
			shouldIgnore: false,
		},
		{
			name:         "empty negation pattern is harmless",
			filePath:     "src/a.ts",
			patterns:     []string{"src/**", "!"},
			shouldIgnore: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isFileIgnoredSimple(tt.filePath, ParseIgnorePatterns(tt.patterns))
			if result != tt.shouldIgnore {
				t.Errorf("isFileIgnoredSimple(%q, %v) = %v, expected %v",
					tt.filePath, tt.patterns, result, tt.shouldIgnore)
			}
		})
	}
}

func TestGetConfigForFile_NegationWithEmptyCwd(t *testing.T) {
	// JSON config path uses empty cwd → isFileIgnoredSimple
	config := RslintConfig{
		{
			Ignores: []string{"vendor/**/*", "!vendor/keep/**/*"},
		},
		{
			Files: []string{"**/*.ts"},
			Rules: Rules{"no-debugger": "error"},
		},
	}

	// cwd="" → uses isFileIgnoredSimple path
	merged := config.GetConfigForFile("vendor/keep/src/b.ts", "")
	if merged != nil {
		t.Fatal("a descendant negation cannot reopen the ignored vendor/keep parent")
	}

	merged2 := config.GetConfigForFile("vendor/lib/src/a.ts", "")
	if merged2 != nil {
		t.Error("Expected vendor/lib/src/a.ts to be ignored with empty cwd")
	}
}

func TestGetConfigForFile_NegationGlobalAndEntryInteraction(t *testing.T) {
	config := RslintConfig{
		// Global ignore re-includes build/test.js
		{
			Ignores: []string{"build/**/*", "!build/test.js"},
		},
		// Entry-level ignore then excludes build/test.js again
		{
			Files:   []string{"**/*.js"},
			Ignores: []string{"build/**"},
			Rules:   Rules{"no-debugger": "error"},
		},
	}

	// build/test.js: global re-includes it, but entry-level ignores it again
	merged := config.GetConfigForFile("build/test.js", "")
	if merged == nil || merged.Rules["no-debugger"] != nil {
		t.Error("entry-local ignores must suppress its rule while retaining the default baseline")
	}

	// src/index.js: not in any ignore → linted
	merged2 := config.GetConfigForFile("src/index.js", "")
	if merged2 == nil {
		t.Fatal("Expected src/index.js to be linted")
		return
	}
}

func TestGetConfigForFile_NegationAcrossGlobalIgnoreEntries(t *testing.T) {
	// build/** is directory-level → blocks entirely, ! cannot undo.
	// To allow negation, use build/**/* (file-level) instead.
	config := RslintConfig{
		{Ignores: []string{"build/**"}},
		{Ignores: []string{"!build/test.js"}},
		{Files: []string{"**/*.js"}, Rules: Rules{"no-debugger": "error"}},
	}

	// build/test.js: build/** blocks directory → ! has no effect → ignored
	merged := config.GetConfigForFile("build/test.js", "")
	if merged != nil {
		t.Error("Expected build/test.js to be ignored (dir/** blocks, ! cannot undo)")
	}

	// build/other.js: also ignored
	merged2 := config.GetConfigForFile("build/other.js", "")
	if merged2 != nil {
		t.Error("Expected build/other.js to remain ignored")
	}

	// With file-level pattern build/**/* → negation DOES work
	config2 := RslintConfig{
		{Ignores: []string{"build/**/*"}},
		{Ignores: []string{"!build/test.js"}},
		{Files: []string{"**/*.js"}, Rules: Rules{"no-debugger": "error"}},
	}
	merged3 := config2.GetConfigForFile("build/test.js", "")
	if merged3 == nil {
		t.Fatal("Expected build/test.js to be re-included with build/**/* (file-level)")
		return
	}

	// Same-entry with dir/** + ! → dir blocks, ! has no effect
	config3 := RslintConfig{
		{Ignores: []string{"build/**", "!build/test.js"}},
		{Files: []string{"**/*.js"}, Rules: Rules{"no-debugger": "error"}},
	}
	merged4 := config3.GetConfigForFile("build/test.js", "")
	if merged4 != nil {
		t.Error("Expected build/test.js to be ignored (dir/** blocks even in same entry)")
	}
}

func TestGetConfigForFile_NegationSequentialOverride(t *testing.T) {
	// A later negation of the same directory pattern reopens it; a subsequent
	// narrower positive pattern can then ignore a child again.
	config := RslintConfig{
		{
			Ignores: []string{"dist/**", "!dist/**", "dist/generated/**"},
		},
		{
			Files: []string{"**/*.js"},
			Rules: Rules{"no-debugger": "error"},
		},
	}

	// !dist/** reopens the directory.
	merged := config.GetConfigForFile("dist/index.js", "")
	if merged == nil {
		t.Error("Expected dist/index.js to be re-included by !dist/**")
	}

	// With file-level pattern: sequential override works
	config2 := RslintConfig{
		{
			Ignores: []string{"dist/**/*", "!dist/**/*", "dist/generated/**/*"},
		},
		{
			Files: []string{"**/*.js"},
			Rules: Rules{"no-debugger": "error"},
		},
	}

	// dist/**/* → ignored, !dist/**/* → re-included, dist/generated/**/* → re-ignored
	merged2 := config2.GetConfigForFile("dist/index.js", "")
	if merged2 == nil {
		t.Fatal("Expected dist/index.js to be linted (file-level sequential override)")
		return
	}

	merged3 := config2.GetConfigForFile("dist/generated/a.js", "")
	if merged3 != nil {
		t.Error("Expected dist/generated/a.js to be ignored")
	}
}

// IsFileIgnored reports only on global ignores, distinct from GetConfigForFile
// which ALSO returns nil when no entry matched the file. This distinction
// matters for --type-check: ignored files must be silenced, but files outside
// rslint's `files` scope should still receive type diagnostics.
func TestIsFileIgnored_OnlyReflectsIgnores(t *testing.T) {
	config := RslintConfig{
		{Ignores: []string{"**/fixtures/**", "**/*.gen.ts"}},
		{Files: []string{"**/*.ts"}, Rules: Rules{"r": "error"}},
	}

	tests := []struct {
		name    string
		path    string
		ignored bool
	}{
		{"directory pattern hits", "packages/x/fixtures/a.ts", true},
		{"file pattern hits", "src/schema.gen.ts", true},
		{"in-scope file not ignored", "src/index.ts", false},
		// Critical: a file that no entry matches (e.g. .js when only **/*.ts
		// has rules) is NOT ignored — it's just out of rslint's scope.
		// GetConfigForFile returns nil for both cases, but IsFileIgnored
		// distinguishes them so --type-check can still report on out-of-scope files.
		{"out-of-scope file not ignored", "src/index.js", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := config.IsFileIgnored(tt.path, ""); got != tt.ignored {
				t.Errorf("IsFileIgnored(%q) = %v, want %v", tt.path, got, tt.ignored)
			}
		})
	}
}

// Empty patterns short-circuit — no config with ignores should never flag a file.
func TestIsFileIgnored_NoPatterns(t *testing.T) {
	config := RslintConfig{
		{Files: []string{"**/*.ts"}, Rules: Rules{"r": "error"}},
	}
	if config.IsFileIgnored("any/file.ts", "") {
		t.Error("IsFileIgnored should return false when config has no ignore patterns")
	}
}

// Directory-level blocking cannot be undone by `!` negation, matching
// ESLint v10's isDirectoryIgnored. IsFileIgnored must reflect this so
// --type-check honors the same blocking rule.
func TestIsFileIgnored_DirectoryBlockingBeatsNegation(t *testing.T) {
	config := RslintConfig{
		{Ignores: []string{"blocked/**", "!blocked/keep.ts"}},
	}
	if !config.IsFileIgnored("blocked/keep.ts", "") {
		t.Error("directory-level block should ignore blocked/keep.ts even with negation")
	}
}

func TestCompileConfigIgnoreLayersOnlyIncludesGlobalEntries(t *testing.T) {
	config := RslintConfig{
		{Ignores: []string{"**/tests/**"}},
		{Files: []string{"**/*.ts"}, Rules: Rules{"r": "error"}},
		{Ignores: []string{"scripts/**"}},
		{Files: []string{}, Ignores: []string{"not-global-empty"}},
		{Ignores: []string{"vendor/**"}, Rules: Rules{"r": "error"}},
		{Files: []string{"**/*.js"}, Ignores: []string{"not-global"}},
	}

	layers := compileConfigIgnoreLayers(config, "/repo", nil)
	if len(layers) != 2 || layers[0].patterns[0].Glob != "**/tests/**" || layers[1].patterns[0].Glob != "scripts/**" {
		t.Fatalf("compileConfigIgnoreLayers() = %#v", layers)
	}
}

func TestCompileConfigIgnoreLayersPreservesEntryOrder(t *testing.T) {
	config := RslintConfig{
		{Ignores: []string{"**/tests/**", "packages/example/compiled/**"}},
		{Ignores: []string{"crates/**"}},
	}

	layers := compileConfigIgnoreLayers(config, "/repo", nil)
	if len(layers) != 2 || len(layers[0].patterns) != 2 || len(layers[1].patterns) != 1 ||
		layers[0].patterns[0].Glob != "**/tests/**" ||
		layers[0].patterns[1].Glob != "packages/example/compiled/**" ||
		layers[1].patterns[0].Glob != "crates/**" {
		t.Fatalf("compileConfigIgnoreLayers() = %#v", layers)
	}
}

func TestIgnoreNegationAndDotNormalizationMatchesESLint(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		ignored bool
	}{
		{name: "one negation strips dot prefix", pattern: "!./src/**", ignored: false},
		{name: "two negations reinclude", pattern: "!!src/**", ignored: false},
		{name: "two negations preserve dot prefix", pattern: "!!./src/**", ignored: true},
		{name: "three negations also reinclude matching positive body", pattern: "!!!src/**", ignored: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			patterns := ParseIgnorePatterns([]string{"**/*.js", test.pattern})
			if got := isFileIgnored("src/a.js", patterns, ""); got != test.ignored {
				t.Fatalf("patterns %q: ignored=%v, want %v", test.pattern, got, test.ignored)
			}
		})
	}
}
