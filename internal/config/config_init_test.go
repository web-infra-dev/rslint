package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// --- Existing init tests (updated) ---

func TestInitDefaultConfig_TSProject(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "tsconfig.json"), []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := InitDefaultConfig(dir); err != nil {
		t.Fatalf("InitDefaultConfig failed: %v", err)
	}

	content := readFile(t, filepath.Join(dir, "rslint.config.ts"))
	if len(content) == 0 {
		t.Error("Expected non-empty config file")
	}
}

func TestInitDefaultConfig_JSProject_NoTypeModule(t *testing.T) {
	dir := t.TempDir()
	if err := InitDefaultConfig(dir); err != nil {
		t.Fatalf("InitDefaultConfig failed: %v", err)
	}

	assertFileExists(t, filepath.Join(dir, "rslint.config.mjs"))
	assertFileNotExists(t, filepath.Join(dir, "rslint.config.js"))
}

func TestInitDefaultConfig_JSProject_TypeModule(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "package.json"), `{"type": "module"}`)

	if err := InitDefaultConfig(dir); err != nil {
		t.Fatalf("InitDefaultConfig failed: %v", err)
	}

	assertFileExists(t, filepath.Join(dir, "rslint.config.js"))
	assertFileNotExists(t, filepath.Join(dir, "rslint.config.mjs"))
}

func TestInitDefaultConfig_JSProject_CJSPackage(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "package.json"), `{"name": "my-project"}`)

	if err := InitDefaultConfig(dir); err != nil {
		t.Fatalf("InitDefaultConfig failed: %v", err)
	}

	assertFileExists(t, filepath.Join(dir, "rslint.config.mjs"))
}

func TestInitDefaultConfig_AlreadyExists(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "rslint.config.ts"), "")

	err := InitDefaultConfig(dir)
	if err == nil {
		t.Error("Expected error when JS/TS config already exists")
	}
	assertContains(t, err.Error(), "config file already exists")
}

// --- Migration branch tests ---

func TestMigrate_JSONExists_TriggersInit(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "tsconfig.json"), "{}")
	writeFile(t, filepath.Join(dir, "rslint.json"), `[{
		"plugins": ["@typescript-eslint"],
		"rules": { "@typescript-eslint/no-floating-promises": "warn" }
	}]`)

	if err := InitDefaultConfig(dir); err != nil {
		t.Fatalf("InitDefaultConfig should migrate, got error: %v", err)
	}

	// JSON should be deleted
	assertFileNotExists(t, filepath.Join(dir, "rslint.json"))
	// JS/TS config should be created
	assertFileExists(t, filepath.Join(dir, "rslint.config.ts"))
}

func TestMigrate_JSONCExists(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "tsconfig.json"), "{}")
	writeFile(t, filepath.Join(dir, "rslint.jsonc"), `[{
		// comment is allowed
		"plugins": ["@typescript-eslint"],
		"rules": {}
	}]`)

	if err := InitDefaultConfig(dir); err != nil {
		t.Fatalf("InitDefaultConfig should migrate jsonc, got error: %v", err)
	}

	assertFileNotExists(t, filepath.Join(dir, "rslint.jsonc"))
	assertFileExists(t, filepath.Join(dir, "rslint.config.ts"))
}

func TestMigrate_BothConfigAndJSON_ErrorsOnExisting(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "rslint.config.ts"), "")
	writeFile(t, filepath.Join(dir, "rslint.json"), `[{"rules":{}}]`)

	err := InitDefaultConfig(dir)
	if err == nil {
		t.Fatal("Expected error when JS/TS config already exists")
		return
	}
	assertContains(t, err.Error(), "config file already exists")
	// JSON should NOT be deleted
	assertFileExists(t, filepath.Join(dir, "rslint.json"))
}

func TestMigrate_EmptyJSON_Errors(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "rslint.json"), `[]`)

	err := InitDefaultConfig(dir)
	if err == nil {
		t.Fatal("Expected error for empty JSON config")
		return
	}
	assertContains(t, err.Error(), "empty")
}

// --- migrateJSONConfig unit tests ---

func TestMigrate_SingleEntry_TSPlugin_RuleDeduplicate(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "tsconfig.json"), "{}")
	writeFile(t, filepath.Join(dir, "rslint.json"), `[{
		"plugins": ["@typescript-eslint"],
		"rules": {
			"@typescript-eslint/no-namespace": "error",
			"@typescript-eslint/no-require-imports": "error",
			"@typescript-eslint/no-explicit-any": "warn",
			"@typescript-eslint/no-floating-promises": "warn",
			"no-console": "warn"
		}
	}]`)

	if err := InitDefaultConfig(dir); err != nil {
		t.Fatal(err)
	}

	content := readFile(t, filepath.Join(dir, "rslint.config.ts"))
	// These rules are in ts.configs.recommended with same severity → should be stripped
	assertNotContains(t, content, "no-namespace")
	assertNotContains(t, content, "no-require-imports")
	// Different severity from preset → should be kept
	assertContains(t, content, "no-explicit-any")
	// Not in preset → should be kept
	assertContains(t, content, "no-floating-promises")
	assertContains(t, content, "no-console")
	// Should reference ts preset
	assertContains(t, content, "ts.configs.recommended")
}

func TestMigrate_SingleEntry_NoPlugin_JSPreset(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "rslint.json"), `[{
		"rules": {
			"for-direction": "error",
			"no-console": "warn"
		}
	}]`)

	if err := InitDefaultConfig(dir); err != nil {
		t.Fatal(err)
	}

	content := readFile(t, filepath.Join(dir, "rslint.config.mjs"))
	// for-direction is in js.configs.recommended at error → strip
	assertNotContains(t, content, "for-direction")
	// no-console not in preset → keep
	assertContains(t, content, "no-console")
	assertContains(t, content, "js.configs.recommended")
}

func TestMigrate_SingleEntry_ReactPlugin(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "rslint.json"), `[{
		"plugins": ["react"],
		"rules": {
			"react/jsx-uses-react": "error",
			"react/no-unsafe": "off",
			"react/self-closing-comp": "error"
		}
	}]`)

	if err := InitDefaultConfig(dir); err != nil {
		t.Fatal(err)
	}

	content := readFile(t, filepath.Join(dir, "rslint.config.mjs"))
	// jsx-uses-react and no-unsafe match preset → strip
	assertNotContains(t, content, "jsx-uses-react")
	assertNotContains(t, content, "no-unsafe")
	// self-closing-comp not in preset → keep
	assertContains(t, content, "self-closing-comp")
	assertContains(t, content, "reactPlugin.configs.recommended")
	assertContains(t, content, "js.configs.recommended")
}

func TestMigrate_SingleEntry_TSAndReactPlugins(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "tsconfig.json"), "{}")
	writeFile(t, filepath.Join(dir, "rslint.json"), `[{
		"plugins": ["@typescript-eslint", "react"],
		"rules": {
			"@typescript-eslint/no-namespace": "error",
			"react/jsx-uses-react": "error",
			"react/self-closing-comp": "error",
			"no-console": "warn"
		}
	}]`)

	if err := InitDefaultConfig(dir); err != nil {
		t.Fatal(err)
	}

	content := readFile(t, filepath.Join(dir, "rslint.config.ts"))
	// Deduped against both presets
	assertNotContains(t, content, "no-namespace")
	assertNotContains(t, content, "jsx-uses-react")
	// Kept
	assertContains(t, content, "self-closing-comp")
	assertContains(t, content, "no-console")
	// Both presets referenced
	assertContains(t, content, "ts.configs.recommended")
	assertContains(t, content, "reactPlugin.configs.recommended")
}

func TestMigrate_SingleEntry_ImportPlugin(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "rslint.json"), `[{
		"plugins": ["eslint-plugin-import"],
		"rules": {}
	}]`)

	if err := InitDefaultConfig(dir); err != nil {
		t.Fatal(err)
	}

	content := readFile(t, filepath.Join(dir, "rslint.config.mjs"))
	assertContains(t, content, "importPlugin.configs.recommended")
	assertContains(t, content, "js.configs.recommended")
}

func TestMigrate_SingleEntry_AllPlugins(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "tsconfig.json"), "{}")
	writeFile(t, filepath.Join(dir, "rslint.json"), `[{
		"plugins": ["@typescript-eslint", "react", "eslint-plugin-import"],
		"rules": {
			"@typescript-eslint/no-explicit-any": "error",
			"react/react-in-jsx-scope": "error",
			"no-console": "warn"
		}
	}]`)

	if err := InitDefaultConfig(dir); err != nil {
		t.Fatal(err)
	}

	content := readFile(t, filepath.Join(dir, "rslint.config.ts"))
	// Both match preset → stripped
	assertNotContains(t, content, "no-explicit-any")
	assertNotContains(t, content, "react-in-jsx-scope")
	// Kept
	assertContains(t, content, "no-console")
	// All presets
	assertContains(t, content, "ts.configs.recommended")
	assertContains(t, content, "reactPlugin.configs.recommended")
	assertContains(t, content, "importPlugin.configs.recommended")
}

func TestMigrate_GlobalIgnoreEntry(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "rslint.json"), `[
		{ "ignores": ["dist/**", "node_modules/**"] },
		{
			"rules": { "no-console": "warn" }
		}
	]`)

	if err := InitDefaultConfig(dir); err != nil {
		t.Fatal(err)
	}

	content := readFile(t, filepath.Join(dir, "rslint.config.mjs"))
	assertContains(t, content, "ignores:")
	assertContains(t, content, "dist/**")
	assertContains(t, content, "node_modules/**")
	assertContains(t, content, "no-console")
}

func TestMigrate_MultiEntry_DifferentPlugins(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "tsconfig.json"), "{}")
	writeFile(t, filepath.Join(dir, "rslint.json"), `[
		{ "ignores": ["dist/**"] },
		{
			"plugins": ["@typescript-eslint"],
			"rules": { "@typescript-eslint/no-unused-vars": "warn" }
		},
		{
			"files": ["**/*.jsx", "**/*.tsx"],
			"plugins": ["react"],
			"rules": { "react/self-closing-comp": "error" }
		}
	]`)

	if err := InitDefaultConfig(dir); err != nil {
		t.Fatal(err)
	}

	content := readFile(t, filepath.Join(dir, "rslint.config.ts"))
	// Global ignore preserved
	assertContains(t, content, "dist/**")
	// TS entry: no-unused-vars is warn (preset is error) → kept
	assertContains(t, content, "no-unused-vars")
	assertContains(t, content, "ts.configs.recommended")
	// React entry
	assertContains(t, content, "reactPlugin.configs.recommended")
	assertContains(t, content, "self-closing-comp")
	// JS preset also referenced (react entry has no TS)
	assertContains(t, content, "js.configs.recommended")
}

func TestMigrate_RuleArrayValue_AlwaysKept(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "tsconfig.json"), "{}")
	writeFile(t, filepath.Join(dir, "rslint.json"), `[{
		"plugins": ["@typescript-eslint"],
		"rules": {
			"@typescript-eslint/no-explicit-any": ["error", { "fixToUnknown": true }]
		}
	}]`)

	if err := InitDefaultConfig(dir); err != nil {
		t.Fatal(err)
	}

	content := readFile(t, filepath.Join(dir, "rslint.config.ts"))
	// Array form should always be kept even though severity matches preset
	assertContains(t, content, "no-explicit-any")
}

func TestMigrate_NumericSeverity(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "rslint.json"), `[{
		"rules": {
			"for-direction": 2,
			"no-console": 1
		}
	}]`)

	if err := InitDefaultConfig(dir); err != nil {
		t.Fatal(err)
	}

	content := readFile(t, filepath.Join(dir, "rslint.config.mjs"))
	// for-direction: 2 == "error", matches js preset → strip
	assertNotContains(t, content, "for-direction")
	// no-console: 1 == "warn", not in preset → keep
	assertContains(t, content, "no-console")
}

func TestMigrate_LanguageOptions_ProjectServiceMatchesDefault(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "tsconfig.json"), "{}")
	writeFile(t, filepath.Join(dir, "rslint.json"), `[{
		"plugins": ["@typescript-eslint"],
		"languageOptions": {
			"parserOptions": {
				"projectService": true
			}
		},
		"rules": { "no-console": "warn" }
	}]`)

	if err := InitDefaultConfig(dir); err != nil {
		t.Fatal(err)
	}

	content := readFile(t, filepath.Join(dir, "rslint.config.ts"))
	// projectService: true is the TS preset default → should be omitted
	assertNotContains(t, content, "projectService")
	assertContains(t, content, "no-console")
}

func TestMigrate_LanguageOptions_ProjectServiceFalse(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "tsconfig.json"), "{}")
	writeFile(t, filepath.Join(dir, "rslint.json"), `[{
		"plugins": ["@typescript-eslint"],
		"languageOptions": {
			"parserOptions": {
				"projectService": false,
				"project": ["./tsconfig.json"]
			}
		},
		"rules": {}
	}]`)

	if err := InitDefaultConfig(dir); err != nil {
		t.Fatal(err)
	}

	content := readFile(t, filepath.Join(dir, "rslint.config.ts"))
	// projectService: false differs from TS preset default → kept
	assertContains(t, content, "projectService: false")
	assertContains(t, content, "tsconfig.json")
}

func TestMigrate_LanguageOptions_ProjectPaths(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "tsconfig.json"), "{}")
	writeFile(t, filepath.Join(dir, "rslint.json"), `[{
		"plugins": ["@typescript-eslint"],
		"languageOptions": {
			"parserOptions": {
				"project": ["./packages/*/tsconfig.json", "./tsconfig.base.json"]
			}
		},
		"rules": {}
	}]`)

	if err := InitDefaultConfig(dir); err != nil {
		t.Fatal(err)
	}

	content := readFile(t, filepath.Join(dir, "rslint.config.ts"))
	assertContains(t, content, "packages/*/tsconfig.json")
	assertContains(t, content, "tsconfig.base.json")
}

func TestMigrate_IgnoresInSingleEntry_ExtractedAsGlobal(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "rslint.json"), `[{
		"ignores": ["test/**", "*.spec.ts"],
		"plugins": ["@typescript-eslint"],
		"rules": { "no-console": "warn" }
	}]`)

	if err := InitDefaultConfig(dir); err != nil {
		t.Fatal(err)
	}

	content := readFile(t, filepath.Join(dir, "rslint.config.mjs"))
	assertContains(t, content, "test/**")
	assertContains(t, content, "*.spec.ts")
	assertContains(t, content, "no-console")

	// Ignores should be extracted as a separate global ignore entry,
	// appearing before the preset so that ignored files are skipped globally.
	ignoresIdx := strings.Index(content, "{ ignores:")
	presetIdx := strings.Index(content, "ts.configs.recommended")
	if ignoresIdx == -1 || presetIdx == -1 {
		t.Fatal("Expected both ignores entry and preset reference")
	}
	if ignoresIdx > presetIdx {
		t.Error("Global ignores entry should appear before preset reference")
	}
}

func TestMigrate_FilesInEntry(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "rslint.json"), `[{
		"files": ["**/*.ts"],
		"plugins": ["@typescript-eslint"],
		"rules": { "no-console": "warn" }
	}]`)

	if err := InitDefaultConfig(dir); err != nil {
		t.Fatal(err)
	}

	content := readFile(t, filepath.Join(dir, "rslint.config.mjs"))
	assertContains(t, content, "files:")
	assertContains(t, content, "**/*.ts")
}

func TestMigrate_EmptyRules_NoOverrideBlock(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "tsconfig.json"), "{}")
	writeFile(t, filepath.Join(dir, "rslint.json"), `[{
		"plugins": ["@typescript-eslint"],
		"rules": {}
	}]`)

	if err := InitDefaultConfig(dir); err != nil {
		t.Fatal(err)
	}

	content := readFile(t, filepath.Join(dir, "rslint.config.ts"))
	assertContains(t, content, "ts.configs.recommended")
	// No rules block needed
	assertNotContains(t, content, "rules:")
}

func TestMigrate_AllRulesMatchPreset_NoOverrideBlock(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "tsconfig.json"), "{}")
	writeFile(t, filepath.Join(dir, "rslint.json"), `[{
		"plugins": ["@typescript-eslint"],
		"rules": {
			"@typescript-eslint/no-namespace": "error",
			"@typescript-eslint/no-require-imports": "error",
			"@typescript-eslint/no-explicit-any": "error",
			"for-direction": "error"
		}
	}]`)

	if err := InitDefaultConfig(dir); err != nil {
		t.Fatal(err)
	}

	content := readFile(t, filepath.Join(dir, "rslint.config.ts"))
	// All rules match preset → no override block
	assertNotContains(t, content, "rules:")
}

func TestMigrate_CoreRuleOffInTSPreset(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "tsconfig.json"), "{}")
	writeFile(t, filepath.Join(dir, "rslint.json"), `[{
		"plugins": ["@typescript-eslint"],
		"rules": {
			"constructor-super": "off",
			"getter-return": "off",
			"no-unused-vars": "off"
		}
	}]`)

	if err := InitDefaultConfig(dir); err != nil {
		t.Fatal(err)
	}

	content := readFile(t, filepath.Join(dir, "rslint.config.ts"))
	// These are "off" in TS preset too → should be stripped
	assertNotContains(t, content, "constructor-super")
	assertNotContains(t, content, "getter-return")
	assertNotContains(t, content, "no-unused-vars")
}

func TestMigrate_CoreRuleSeverityDiffersFromPreset(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "rslint.json"), `[{
		"rules": {
			"no-debugger": "warn",
			"no-empty": "off"
		}
	}]`)

	if err := InitDefaultConfig(dir); err != nil {
		t.Fatal(err)
	}

	content := readFile(t, filepath.Join(dir, "rslint.config.mjs"))
	// JS preset has both as "error" → user overrides should be kept
	assertContains(t, content, "no-debugger")
	assertContains(t, content, "'warn'")
	assertContains(t, content, "no-empty")
	assertContains(t, content, "'off'")
}

func TestMigrate_OutputFormat_ESM(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "package.json"), `{"type": "module"}`)
	writeFile(t, filepath.Join(dir, "rslint.json"), `[{
		"rules": { "no-console": "warn" }
	}]`)

	if err := InitDefaultConfig(dir); err != nil {
		t.Fatal(err)
	}

	// ESM project → .js
	assertFileExists(t, filepath.Join(dir, "rslint.config.js"))
	assertFileNotExists(t, filepath.Join(dir, "rslint.config.mjs"))
}

func TestMigrate_OutputFormat_CJS(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "rslint.json"), `[{
		"rules": { "no-console": "warn" }
	}]`)

	if err := InitDefaultConfig(dir); err != nil {
		t.Fatal(err)
	}

	// CJS project → .mjs
	assertFileExists(t, filepath.Join(dir, "rslint.config.mjs"))
}

func TestMigrate_OutputFormat_TSProject(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "tsconfig.json"), "{}")
	writeFile(t, filepath.Join(dir, "rslint.json"), `[{
		"rules": { "no-console": "warn" }
	}]`)

	if err := InitDefaultConfig(dir); err != nil {
		t.Fatal(err)
	}

	assertFileExists(t, filepath.Join(dir, "rslint.config.ts"))
}

func TestMigrate_JSONDeleted(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "rslint.json"), `[{ "rules": {} }]`)

	if err := InitDefaultConfig(dir); err != nil {
		t.Fatal(err)
	}

	assertFileNotExists(t, filepath.Join(dir, "rslint.json"))
}

func TestMigrate_Settings_Preserved(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "rslint.json"), `[{
		"settings": { "react": { "version": "18" } },
		"plugins": ["react"],
		"rules": {}
	}]`)

	if err := InitDefaultConfig(dir); err != nil {
		t.Fatal(err)
	}

	content := readFile(t, filepath.Join(dir, "rslint.config.mjs"))
	assertContains(t, content, "settings:")
}

func TestMigrate_LanguageField_Dropped(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "rslint.json"), `[{
		"language": "javascript",
		"rules": { "no-console": "warn" }
	}]`)

	if err := InitDefaultConfig(dir); err != nil {
		t.Fatal(err)
	}

	content := readFile(t, filepath.Join(dir, "rslint.config.mjs"))
	// "language" JSON field should not appear as a standalone key in JS config
	assertNotContains(t, content, "'language'")
}

func TestMigrate_EmptyFilesArray_Dropped(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "rslint.json"), `[{
		"files": [],
		"rules": { "no-console": "warn" }
	}]`)

	if err := InitDefaultConfig(dir); err != nil {
		t.Fatal(err)
	}

	content := readFile(t, filepath.Join(dir, "rslint.config.mjs"))
	// Empty files array should be dropped
	assertNotContains(t, content, "files:")
}

func TestMigrate_RealWorldConfig(t *testing.T) {
	// Simulates the repo's own rslint.json
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "tsconfig.json"), "{}")
	writeFile(t, filepath.Join(dir, "rslint.json"), `[{
		"language": "javascript",
		"files": [],
		"ignores": [
			"node_modules/**",
			"**/dist/**",
			"packages/rslint/fixtures/**"
		],
		"languageOptions": {
			"parserOptions": {
				"projectService": false,
				"project": [
					"./packages/*/tsconfig.build.json",
					"./packages/*/tsconfig.spec.json"
				]
			}
		},
		"rules": {
			"@typescript-eslint/no-namespace": "error",
			"@typescript-eslint/no-require-imports": "error",
			"@typescript-eslint/no-explicit-any": "warn",
			"@typescript-eslint/no-floating-promises": "warn",
			"@typescript-eslint/no-empty-function": "error",
			"@typescript-eslint/no-empty-interface": "error",
			"@typescript-eslint/no-unused-vars": "warn",
			"@typescript-eslint/no-dynamic-delete": "off",
			"@typescript-eslint/prefer-includes": "off",
			"no-console": "warn"
		},
		"plugins": ["@typescript-eslint"]
	}]`)

	if err := InitDefaultConfig(dir); err != nil {
		t.Fatal(err)
	}

	content := readFile(t, filepath.Join(dir, "rslint.config.ts"))

	// Preset reference
	assertContains(t, content, "ts.configs.recommended")
	// Import line
	assertContains(t, content, "import { defineConfig, ts } from '@rslint/core'")

	// Deduplicated (same severity as TS preset)
	assertNotContains(t, content, "no-namespace")
	assertNotContains(t, content, "no-require-imports")

	// Kept (different severity or not in preset)
	assertContains(t, content, "no-explicit-any")
	assertContains(t, content, "no-floating-promises")
	assertContains(t, content, "no-empty-function")
	assertContains(t, content, "no-empty-interface")
	assertContains(t, content, "no-unused-vars")
	assertContains(t, content, "no-dynamic-delete")
	assertContains(t, content, "prefer-includes")
	assertContains(t, content, "no-console")

	// Language options preserved (projectService: false differs from preset default)
	assertContains(t, content, "projectService: false")
	assertContains(t, content, "tsconfig.build.json")

	// Ignores extracted as global ignore entry (single non-ignore entry)
	assertContains(t, content, "node_modules/**")
	assertContains(t, content, "**/dist/**")
	ignoresIdx := strings.Index(content, "{ ignores:")
	presetIdx := strings.Index(content, "ts.configs.recommended")
	if ignoresIdx == -1 || presetIdx == -1 {
		t.Fatal("Expected both ignores entry and preset reference")
	}
	if ignoresIdx > presetIdx {
		t.Error("Global ignores entry should appear before preset reference")
	}

	// "language" JSON field should not appear as a standalone key in JS config
	// (but "languageOptions" is expected)
	assertNotContains(t, content, "'language'")

	// JSON deleted
	assertFileNotExists(t, filepath.Join(dir, "rslint.json"))
}

// --- Additional edge case tests ---

func TestMigrate_MultipleGlobalIgnoreEntries(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "rslint.json"), `[
		{ "ignores": ["dist/**"] },
		{ "ignores": ["build/**"] },
		{ "rules": { "no-console": "warn" } }
	]`)

	if err := InitDefaultConfig(dir); err != nil {
		t.Fatal(err)
	}

	content := readFile(t, filepath.Join(dir, "rslint.config.mjs"))
	assertContains(t, content, "dist/**")
	assertContains(t, content, "build/**")
	assertContains(t, content, "no-console")
}

func TestMigrate_OnlyGlobalIgnoreEntries(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "rslint.json"), `[
		{ "ignores": ["dist/**", "node_modules/**"] }
	]`)

	if err := InitDefaultConfig(dir); err != nil {
		t.Fatal(err)
	}

	content := readFile(t, filepath.Join(dir, "rslint.config.mjs"))
	assertContains(t, content, "defineConfig")
	assertContains(t, content, "dist/**")
	// No preset should be referenced when there are only ignore entries
	assertNotContains(t, content, "recommended")
}

func TestMigrate_EmptyPluginsArray_TreatedAsNoPlugin(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "rslint.json"), `[{
		"plugins": [],
		"rules": { "no-console": "warn" }
	}]`)

	if err := InitDefaultConfig(dir); err != nil {
		t.Fatal(err)
	}

	content := readFile(t, filepath.Join(dir, "rslint.config.mjs"))
	// Empty plugins → should use JS preset
	assertContains(t, content, "js.configs.recommended")
	assertNotContains(t, content, "ts.configs.recommended")
}

func TestMigrate_UnknownPlugin_StillGeneratesPreset(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "rslint.json"), `[{
		"plugins": ["some-unknown-plugin"],
		"rules": { "no-console": "warn" }
	}]`)

	if err := InitDefaultConfig(dir); err != nil {
		t.Fatal(err)
	}

	content := readFile(t, filepath.Join(dir, "rslint.config.mjs"))
	// Unknown plugin → falls through to JS preset (no TS detected)
	assertContains(t, content, "js.configs.recommended")
	assertContains(t, content, "no-console")
}

func TestMigrate_ArrayRuleSingleElement(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "rslint.json"), `[{
		"rules": {
			"for-direction": ["error"]
		}
	}]`)

	if err := InitDefaultConfig(dir); err != nil {
		t.Fatal(err)
	}

	content := readFile(t, filepath.Join(dir, "rslint.config.mjs"))
	// Array form should always be kept even with single element matching preset
	assertContains(t, content, "for-direction")
}

func TestMigrate_NumericSeverity_Zero(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "tsconfig.json"), "{}")
	writeFile(t, filepath.Join(dir, "rslint.json"), `[{
		"plugins": ["@typescript-eslint"],
		"rules": {
			"constructor-super": 0,
			"no-debugger": 0
		}
	}]`)

	if err := InitDefaultConfig(dir); err != nil {
		t.Fatal(err)
	}

	content := readFile(t, filepath.Join(dir, "rslint.config.ts"))
	// constructor-super: 0 == "off", matches TS preset → strip
	assertNotContains(t, content, "constructor-super")
	// no-debugger: 0 == "off", TS preset has "error" → keep
	assertContains(t, content, "no-debugger")
	assertContains(t, content, "'off'")
}

func TestMigrate_MixedNumericAndStringSeverity(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "rslint.json"), `[{
		"rules": {
			"for-direction": 2,
			"no-debugger": "error",
			"no-empty": 1,
			"no-console": "warn"
		}
	}]`)

	if err := InitDefaultConfig(dir); err != nil {
		t.Fatal(err)
	}

	content := readFile(t, filepath.Join(dir, "rslint.config.mjs"))
	// for-direction: 2 matches preset → strip
	assertNotContains(t, content, "for-direction")
	// no-debugger: "error" matches preset → strip
	assertNotContains(t, content, "no-debugger")
	// no-empty: 1 ("warn") differs from preset "error" → keep
	assertContains(t, content, "no-empty")
	// no-console: not in preset → keep
	assertContains(t, content, "no-console")
}

func TestMigrate_LongIgnoresArray_MultiLine(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "rslint.json"), `[{
		"ignores": ["dist/**", "build/**", "node_modules/**", ".cache/**"],
		"rules": { "no-console": "warn" }
	}]`)

	if err := InitDefaultConfig(dir); err != nil {
		t.Fatal(err)
	}

	content := readFile(t, filepath.Join(dir, "rslint.config.mjs"))
	// >3 items should be formatted multi-line
	assertContains(t, content, "dist/**")
	assertContains(t, content, "build/**")
	assertContains(t, content, "node_modules/**")
	assertContains(t, content, ".cache/**")
}

func TestMigrate_GlobPatternWithSpecialChars(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "rslint.json"), `[{
		"ignores": ["src/**/*.test.ts", "packages/foo-bar/**"],
		"rules": {}
	}]`)

	if err := InitDefaultConfig(dir); err != nil {
		t.Fatal(err)
	}

	content := readFile(t, filepath.Join(dir, "rslint.config.mjs"))
	assertContains(t, content, "src/**/*.test.ts")
	assertContains(t, content, "packages/foo-bar/**")
}

func TestMigrate_StringWithSingleQuote(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "rslint.json"), `[{
		"ignores": ["vendor's/**"],
		"rules": {}
	}]`)

	if err := InitDefaultConfig(dir); err != nil {
		t.Fatal(err)
	}

	content := readFile(t, filepath.Join(dir, "rslint.config.mjs"))
	// Single quotes should be escaped
	assertContains(t, content, "vendor\\'s/**")
}

func TestMigrate_DuplicatePresetImport_OnlyOnce(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "tsconfig.json"), "{}")
	writeFile(t, filepath.Join(dir, "rslint.json"), `[
		{
			"files": ["src/**/*.ts"],
			"plugins": ["@typescript-eslint"],
			"rules": { "no-console": "warn" }
		},
		{
			"files": ["lib/**/*.ts"],
			"plugins": ["@typescript-eslint"],
			"rules": { "@typescript-eslint/no-dynamic-delete": "off" }
		}
	]`)

	if err := InitDefaultConfig(dir); err != nil {
		t.Fatal(err)
	}

	content := readFile(t, filepath.Join(dir, "rslint.config.ts"))
	// "ts" should only appear once in the import line
	importLine := ""
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(line, "import") {
			importLine = line
			break
		}
	}
	if strings.Count(importLine, "ts") != 1 {
		t.Errorf("Expected 'ts' to appear exactly once in import line, got: %s", importLine)
	}
}

func TestMigrate_TSAndJSEntries_BothImported(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "tsconfig.json"), "{}")
	writeFile(t, filepath.Join(dir, "rslint.json"), `[
		{
			"files": ["**/*.ts"],
			"plugins": ["@typescript-eslint"],
			"rules": { "no-console": "warn" }
		},
		{
			"files": ["**/*.js"],
			"rules": { "no-console": "warn" }
		}
	]`)

	if err := InitDefaultConfig(dir); err != nil {
		t.Fatal(err)
	}

	content := readFile(t, filepath.Join(dir, "rslint.config.ts"))
	// Both ts and js should be imported
	assertContains(t, content, "ts.configs.recommended")
	assertContains(t, content, "js.configs.recommended")
	importLine := ""
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(line, "import") {
			importLine = line
			break
		}
	}
	assertContains(t, importLine, " ts")
	assertContains(t, importLine, " js")
}

func TestMigrate_EntryWithOnlyIgnoresAndFiles_NotGlobalIgnore(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "rslint.json"), `[{
		"files": ["**/*.ts"],
		"ignores": ["**/*.test.ts"],
		"rules": {}
	}]`)

	if err := InitDefaultConfig(dir); err != nil {
		t.Fatal(err)
	}

	content := readFile(t, filepath.Join(dir, "rslint.config.mjs"))
	// Has files → not a global ignore entry → should get a preset
	assertContains(t, content, "js.configs.recommended")
	assertContains(t, content, "**/*.ts")
	assertContains(t, content, "**/*.test.ts")
}

func TestMigrate_EntryWithOnlyLanguageOptions(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "tsconfig.json"), "{}")
	writeFile(t, filepath.Join(dir, "rslint.json"), `[{
		"plugins": ["@typescript-eslint"],
		"languageOptions": {
			"parserOptions": {
				"projectService": false,
				"project": ["./tsconfig.app.json"]
			}
		},
		"rules": {}
	}]`)

	if err := InitDefaultConfig(dir); err != nil {
		t.Fatal(err)
	}

	content := readFile(t, filepath.Join(dir, "rslint.config.ts"))
	// Override block should exist for languageOptions even with no rules
	assertContains(t, content, "languageOptions")
	assertContains(t, content, "projectService: false")
	assertContains(t, content, "tsconfig.app.json")
}

func TestMigrate_LanguageOptions_EmptyParserOptions(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "tsconfig.json"), "{}")
	writeFile(t, filepath.Join(dir, "rslint.json"), `[{
		"plugins": ["@typescript-eslint"],
		"languageOptions": {
			"parserOptions": {}
		},
		"rules": {}
	}]`)

	if err := InitDefaultConfig(dir); err != nil {
		t.Fatal(err)
	}

	content := readFile(t, filepath.Join(dir, "rslint.config.ts"))
	// Empty parserOptions → no languageOptions output
	assertNotContains(t, content, "languageOptions")
}

func TestMigrate_LanguageOptions_ProjectServiceTrueForJSEntry(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "rslint.json"), `[{
		"languageOptions": {
			"parserOptions": {
				"projectService": true
			}
		},
		"rules": {}
	}]`)

	if err := InitDefaultConfig(dir); err != nil {
		t.Fatal(err)
	}

	content := readFile(t, filepath.Join(dir, "rslint.config.mjs"))
	// JS preset has no projectService default → true is a non-default value → should be kept
	assertContains(t, content, "projectService: true")
}

func TestMigrate_ProjectPathSingleString(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "tsconfig.json"), "{}")
	writeFile(t, filepath.Join(dir, "rslint.json"), `[{
		"plugins": ["@typescript-eslint"],
		"languageOptions": {
			"parserOptions": {
				"project": "./tsconfig.json"
			}
		},
		"rules": {}
	}]`)

	if err := InitDefaultConfig(dir); err != nil {
		t.Fatal(err)
	}

	content := readFile(t, filepath.Join(dir, "rslint.config.ts"))
	assertContains(t, content, "tsconfig.json")
}

func TestMigrate_InvalidJSON_ReturnsError(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "rslint.json"), `{not valid json`)

	err := InitDefaultConfig(dir)
	if err == nil {
		t.Fatal("Expected error for invalid JSON")
		return
	}
	assertContains(t, err.Error(), "failed to parse")
}

func TestMigrate_JSONNotArray_ReturnsError(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "rslint.json"), `{"rules": {}}`)

	err := InitDefaultConfig(dir)
	if err == nil {
		t.Fatal("Expected error for non-array JSON")
		return
	}
}

func TestMigrate_JSONCWithTrailingCommas(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "rslint.json"), `[
		{
			"plugins": ["@typescript-eslint",],
			"rules": {
				"no-console": "warn",
			},
		},
	]`)

	if err := InitDefaultConfig(dir); err != nil {
		t.Fatal(err)
	}

	content := readFile(t, filepath.Join(dir, "rslint.config.mjs"))
	assertContains(t, content, "no-console")
}

func TestMigrate_AllRulesDeduped_ButHasIgnores_ExtractedAsGlobal(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "tsconfig.json"), "{}")
	writeFile(t, filepath.Join(dir, "rslint.json"), `[{
		"plugins": ["@typescript-eslint"],
		"ignores": ["test/**"],
		"rules": {
			"@typescript-eslint/no-namespace": "error",
			"@typescript-eslint/no-explicit-any": "error"
		}
	}]`)

	if err := InitDefaultConfig(dir); err != nil {
		t.Fatal(err)
	}

	content := readFile(t, filepath.Join(dir, "rslint.config.ts"))
	// All rules deduped, ignores extracted as global entry
	assertNotContains(t, content, "rules:")
	assertContains(t, content, "{ ignores:")
	assertContains(t, content, "test/**")
}

func TestMigrate_ReactRuleDifferentSeverity(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "rslint.json"), `[{
		"plugins": ["react"],
		"rules": {
			"react/jsx-uses-react": "warn",
			"react/no-unsafe": "error"
		}
	}]`)

	if err := InitDefaultConfig(dir); err != nil {
		t.Fatal(err)
	}

	content := readFile(t, filepath.Join(dir, "rslint.config.mjs"))
	// Both differ from preset defaults → should be kept
	assertContains(t, content, "jsx-uses-react")
	assertContains(t, content, "no-unsafe")
}

func TestMigrate_MultiEntry_PresetOrderCorrect(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "tsconfig.json"), "{}")
	writeFile(t, filepath.Join(dir, "rslint.json"), `[
		{
			"plugins": ["@typescript-eslint", "react", "eslint-plugin-import"],
			"rules": { "no-console": "warn" }
		}
	]`)

	if err := InitDefaultConfig(dir); err != nil {
		t.Fatal(err)
	}

	content := readFile(t, filepath.Join(dir, "rslint.config.ts"))
	// Preset references should appear in order: ts, react, import
	tsIdx := strings.Index(content, "ts.configs.recommended")
	reactIdx := strings.Index(content, "reactPlugin.configs.recommended")
	importIdx := strings.Index(content, "importPlugin.configs.recommended")
	if tsIdx == -1 || reactIdx == -1 || importIdx == -1 {
		t.Fatal("Expected all three preset references")
	}
	if tsIdx > reactIdx || reactIdx > importIdx {
		t.Error("Expected preset order: ts, react, import")
	}
}

func TestMigrate_RulesOrderDeterministic(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "rslint.json"), `[{
		"rules": {
			"no-console": "warn",
			"array-callback-return": "error",
			"@typescript-eslint/no-floating-promises": "warn"
		},
		"plugins": ["@typescript-eslint"]
	}]`)

	if err := InitDefaultConfig(dir); err != nil {
		t.Fatal(err)
	}

	content := readFile(t, filepath.Join(dir, "rslint.config.mjs"))
	// Rules should be sorted alphabetically
	tsIdx := strings.Index(content, "@typescript-eslint/no-floating-promises")
	arrIdx := strings.Index(content, "array-callback-return")
	consIdx := strings.Index(content, "no-console")
	if tsIdx == -1 || arrIdx == -1 || consIdx == -1 {
		t.Fatal("Expected all three rules in output")
	}
	if tsIdx > arrIdx || arrIdx > consIdx {
		t.Errorf("Expected rules sorted: @typescript-eslint < array-callback-return < no-console")
	}
}

func TestMigrate_EmptyIgnoresArray_Dropped(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "rslint.json"), `[{
		"ignores": [],
		"rules": { "no-console": "warn" }
	}]`)

	if err := InitDefaultConfig(dir); err != nil {
		t.Fatal(err)
	}

	content := readFile(t, filepath.Join(dir, "rslint.config.mjs"))
	assertNotContains(t, content, "ignores:")
}

func TestMigrate_NilRulesField(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "rslint.json"), `[{
		"plugins": ["@typescript-eslint"]
	}]`)

	if err := InitDefaultConfig(dir); err != nil {
		t.Fatal(err)
	}

	content := readFile(t, filepath.Join(dir, "rslint.config.mjs"))
	assertContains(t, content, "ts.configs.recommended")
	assertNotContains(t, content, "rules:")
}

func TestMigrate_LargeMultiEntryConfig(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "tsconfig.json"), "{}")
	writeFile(t, filepath.Join(dir, "rslint.json"), `[
		{
			"ignores": ["dist/**", "node_modules/**", "coverage/**", ".next/**"]
		},
		{
			"files": ["**/*.ts", "**/*.tsx"],
			"plugins": ["@typescript-eslint"],
			"languageOptions": {
				"parserOptions": {
					"projectService": false,
					"project": ["./tsconfig.json"]
				}
			},
			"rules": {
				"@typescript-eslint/no-explicit-any": "warn",
				"@typescript-eslint/no-namespace": "error",
				"@typescript-eslint/no-unused-vars": ["error", { "argsIgnorePattern": "^_" }],
				"no-console": "warn"
			}
		},
		{
			"files": ["**/*.jsx", "**/*.tsx"],
			"plugins": ["react"],
			"rules": {
				"react/jsx-uses-react": "error",
				"react/self-closing-comp": "error"
			},
			"settings": { "react": { "version": "detect" } }
		},
		{
			"files": ["**/*.test.ts"],
			"rules": {
				"no-console": "off"
			}
		}
	]`)

	if err := InitDefaultConfig(dir); err != nil {
		t.Fatal(err)
	}

	content := readFile(t, filepath.Join(dir, "rslint.config.ts"))

	// Global ignores
	assertContains(t, content, "dist/**")
	assertContains(t, content, ".next/**")

	// TS entry: no-namespace deduped, no-explicit-any kept (warn), no-unused-vars kept (array)
	assertNotContains(t, content, "'@typescript-eslint/no-namespace'")
	assertContains(t, content, "no-explicit-any")
	assertContains(t, content, "no-unused-vars")
	assertContains(t, content, "argsIgnorePattern")

	// React entry: jsx-uses-react deduped, self-closing-comp kept
	assertContains(t, content, "self-closing-comp")
	assertContains(t, content, "settings:")

	// Test entry: separate JS preset for test overrides
	assertContains(t, content, "**/*.test.ts")

	// Import line should have ts, js, reactPlugin
	assertContains(t, content, "defineConfig")
	assertContains(t, content, "ts.configs.recommended")
	assertContains(t, content, "reactPlugin.configs.recommended")
}

func TestMigrate_ConfigEntry_OutputIsValidJSStructure(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "rslint.json"), `[{
		"rules": { "no-console": "warn" }
	}]`)

	if err := InitDefaultConfig(dir); err != nil {
		t.Fatal(err)
	}

	content := readFile(t, filepath.Join(dir, "rslint.config.mjs"))
	// Verify basic structural elements
	assertContains(t, content, "import {")
	assertContains(t, content, "} from '@rslint/core'")
	assertContains(t, content, "export default defineConfig([")
	assertContains(t, content, "]);")
	// Should not have double commas or syntax errors
	assertNotContains(t, content, ",,")
}

func TestMigrate_ExistingGlobalIgnorePlusSingleEntryWithIgnores(t *testing.T) {
	// Already have a global ignore entry + one entry with its own ignores.
	// The entry's ignores should be extracted (only 1 non-global-ignore entry).
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "rslint.json"), `[
		{ "ignores": ["dist/**"] },
		{
			"ignores": ["test/**"],
			"plugins": ["@typescript-eslint"],
			"rules": { "no-console": "warn" }
		}
	]`)

	if err := InitDefaultConfig(dir); err != nil {
		t.Fatal(err)
	}

	content := readFile(t, filepath.Join(dir, "rslint.config.mjs"))
	assertContains(t, content, "dist/**")
	assertContains(t, content, "test/**")
	assertContains(t, content, "no-console")
	// The entry's ignores should be extracted as another global ignore entry
	// Both should appear before the preset
	presetIdx := strings.Index(content, "ts.configs.recommended")
	distIdx := strings.Index(content, "dist/**")
	testIdx := strings.Index(content, "test/**")
	if distIdx > presetIdx || testIdx > presetIdx {
		t.Error("Both ignore entries should appear before preset reference")
	}
}

func TestMigrate_GlobalIgnoreEntry_LongArray(t *testing.T) {
	// Test that global ignore entries with >3 items format correctly
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "rslint.json"), `[
		{ "ignores": ["dist/**", "build/**", "node_modules/**", ".cache/**", "tmp/**"] },
		{ "rules": { "no-console": "warn" } }
	]`)

	if err := InitDefaultConfig(dir); err != nil {
		t.Fatal(err)
	}

	content := readFile(t, filepath.Join(dir, "rslint.config.mjs"))
	assertContains(t, content, "dist/**")
	assertContains(t, content, "tmp/**")
	assertContains(t, content, "no-console")
}

func TestMigrate_MultiEntry_IgnoresNotExtracted(t *testing.T) {
	// When there are multiple non-global-ignore entries, ignores should stay
	// in their respective entries, not be extracted as global ignores.
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "tsconfig.json"), "{}")
	writeFile(t, filepath.Join(dir, "rslint.json"), `[
		{
			"files": ["**/*.ts"],
			"ignores": ["test/**"],
			"plugins": ["@typescript-eslint"],
			"rules": { "no-console": "warn" }
		},
		{
			"files": ["**/*.jsx"],
			"ignores": ["stories/**"],
			"plugins": ["react"],
			"rules": { "react/self-closing-comp": "error" }
		}
	]`)

	if err := InitDefaultConfig(dir); err != nil {
		t.Fatal(err)
	}

	content := readFile(t, filepath.Join(dir, "rslint.config.ts"))
	// Both ignores should be in their own override blocks, not extracted
	assertContains(t, content, "test/**")
	assertContains(t, content, "stories/**")
	// No standalone global ignore entry (ignores are inside override blocks)
	assertNotContains(t, content, "{ ignores:")
}

func TestMigrate_SingleEntryNoIgnores_NoExtraction(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "rslint.json"), `[{
		"plugins": ["@typescript-eslint"],
		"rules": { "no-console": "warn" }
	}]`)

	if err := InitDefaultConfig(dir); err != nil {
		t.Fatal(err)
	}

	content := readFile(t, filepath.Join(dir, "rslint.config.mjs"))
	// No ignores → no global ignore entry
	assertNotContains(t, content, "{ ignores:")
}

// --- deduplicateRules unit tests ---

func TestDeduplicateRules_AllMatch(t *testing.T) {
	user := Rules{
		"@typescript-eslint/no-namespace":       "error",
		"@typescript-eslint/no-require-imports": "error",
	}
	result := deduplicateRules(user, tsRecommendedRules)
	if result != nil {
		t.Errorf("Expected nil, got %v", result)
	}
}

func TestDeduplicateRules_NoneMatch(t *testing.T) {
	user := Rules{
		"no-console":                            "warn",
		"@typescript-eslint/no-floating-promises": "warn",
	}
	result := deduplicateRules(user, tsRecommendedRules)
	if len(result) != 2 {
		t.Errorf("Expected 2 rules, got %d", len(result))
	}
}

func TestDeduplicateRules_Mixed(t *testing.T) {
	user := Rules{
		"@typescript-eslint/no-namespace":     "error", // match → strip
		"@typescript-eslint/no-explicit-any":  "warn",  // differs → keep
		"no-console":                          "warn",  // not in preset → keep
	}
	result := deduplicateRules(user, tsRecommendedRules)
	if len(result) != 2 {
		t.Errorf("Expected 2 rules, got %d: %v", len(result), result)
	}
	if _, ok := result["@typescript-eslint/no-namespace"]; ok {
		t.Error("no-namespace should have been stripped")
	}
}

func TestDeduplicateRules_ArrayValueAlwaysKept(t *testing.T) {
	user := Rules{
		"@typescript-eslint/no-explicit-any": []interface{}{"error", map[string]interface{}{"fixToUnknown": true}},
	}
	result := deduplicateRules(user, tsRecommendedRules)
	if len(result) != 1 {
		t.Errorf("Expected array rule to be kept, got %d rules", len(result))
	}
}

func TestDeduplicateRules_EmptyInput(t *testing.T) {
	result := deduplicateRules(nil, tsRecommendedRules)
	if result != nil {
		t.Errorf("Expected nil for nil input, got %v", result)
	}
}

func TestNormalizeSeverity(t *testing.T) {
	tests := []struct {
		input, expected string
	}{
		{"error", "error"},
		{"warn", "warn"},
		{"off", "off"},
		{"2", "error"},
		{"1", "warn"},
		{"0", "off"},
	}
	for _, tc := range tests {
		if got := normalizeSeverity(tc.input); got != tc.expected {
			t.Errorf("normalizeSeverity(%q) = %q, want %q", tc.input, got, tc.expected)
		}
	}
}

// --- Test helpers ---

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read %s: %v", path, err)
	}
	return string(data)
}

func assertFileExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("Expected file to exist: %s", filepath.Base(path))
	}
}

func assertFileNotExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err == nil {
		t.Errorf("Expected file NOT to exist: %s", filepath.Base(path))
	}
}

func assertContains(t *testing.T, content, substr string) {
	t.Helper()
	if !strings.Contains(content, substr) {
		t.Errorf("Expected output to contain %q, got:\n%s", substr, content)
	}
}

func assertNotContains(t *testing.T, content, substr string) {
	t.Helper()
	if strings.Contains(content, substr) {
		t.Errorf("Expected output NOT to contain %q, got:\n%s", substr, content)
	}
}
