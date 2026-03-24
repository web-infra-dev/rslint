package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/web-infra-dev/rslint/internal/utils"
)

const defaultTSConfig = `import { defineConfig, ts } from '@rslint/core';

export default defineConfig([
  ts.configs.recommended,
  {
    rules: {
      // customize rules here
    },
  },
]);
`

const defaultJSConfig = `import { defineConfig, js } from '@rslint/core';

export default defineConfig([
  js.configs.recommended,
  {
    rules: {
      // customize rules here
    },
  },
]);
`

// The preset rule maps below are intentionally hardcoded and do NOT need to be updated
// when new rules are added to recommended presets. They are only used during JSON→JS/TS
// migration to strip rules that are redundant with the preset. New rules added to presets
// won't exist in users' old JSON configs, so there is nothing to strip.
// The only case requiring an update is if an existing rule's severity changes in a preset,
// which is a breaking change and should not happen in practice.

// jsRecommendedRules contains all rules in js.configs.recommended with their severities.
var jsRecommendedRules = map[string]string{
	"constructor-super":            "error",
	"for-direction":                "error",
	"getter-return":                "error",
	"no-async-promise-executor":    "error",
	"no-case-declarations":         "error",
	"no-class-assign":              "error",
	"no-compare-neg-zero":          "error",
	"no-cond-assign":               "error",
	"no-const-assign":              "error",
	"no-constant-binary-expression": "error",
	"no-constant-condition":        "error",
	"no-debugger":                  "error",
	"no-dupe-args":                 "error",
	"no-dupe-keys":                 "error",
	"no-duplicate-case":            "error",
	"no-empty":                     "error",
	"no-empty-pattern":             "error",
	"no-loss-of-precision":         "error",
	"no-sparse-arrays":             "error",
}

// tsRecommendedRules contains all rules in ts.configs.recommended with their severities.
var tsRecommendedRules = map[string]string{
	// Core rules turned off (handled by TypeScript)
	"constructor-super": "off",
	"getter-return":     "off",
	"no-class-assign":   "off",
	"no-const-assign":   "off",
	"no-dupe-args":      "off",
	"no-dupe-keys":      "off",
	// Core rules replaced by TS versions
	"no-array-constructor": "off",
	"no-unused-vars":       "off",
	// Core rules kept
	"for-direction":                "error",
	"no-async-promise-executor":    "error",
	"no-case-declarations":         "error",
	"no-compare-neg-zero":          "error",
	"no-cond-assign":               "error",
	"no-constant-binary-expression": "error",
	"no-constant-condition":        "error",
	"no-debugger":                  "error",
	"no-duplicate-case":            "error",
	"no-empty":                     "error",
	"no-empty-pattern":             "error",
	"no-loss-of-precision":         "error",
	"no-sparse-arrays":             "error",
	// TypeScript plugin rules
	"@typescript-eslint/ban-ts-comment":                  "error",
	"@typescript-eslint/no-array-constructor":             "error",
	"@typescript-eslint/no-duplicate-enum-values":         "error",
	"@typescript-eslint/no-explicit-any":                  "error",
	"@typescript-eslint/no-extra-non-null-assertion":      "error",
	"@typescript-eslint/no-misused-new":                   "error",
	"@typescript-eslint/no-namespace":                     "error",
	"@typescript-eslint/no-non-null-asserted-optional-chain": "error",
	"@typescript-eslint/no-require-imports":               "error",
	"@typescript-eslint/no-this-alias":                    "error",
	"@typescript-eslint/no-unused-vars":                   "error",
	"@typescript-eslint/prefer-as-const":                  "error",
	"@typescript-eslint/prefer-namespace-keyword":         "error",
	"@typescript-eslint/triple-slash-reference":           "error",
}

// reactRecommendedRules contains all rules in reactPlugin.configs.recommended with their severities.
var reactRecommendedRules = map[string]string{
	"react/jsx-uses-react":    "error",
	"react/jsx-uses-vars":     "error",
	"react/react-in-jsx-scope": "error",
	"react/no-unsafe":         "off",
}

// importRecommendedRules is empty — all import plugin recommended rules are not yet implemented.
var importRecommendedRules = map[string]string{}

// isESMPackage checks if the package.json in the given directory has "type": "module".
func isESMPackage(directory string) bool {
	data, err := os.ReadFile(filepath.Join(directory, "package.json"))
	if err != nil {
		return false
	}
	var pkg struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return false
	}
	return pkg.Type == "module"
}

// InitDefaultConfig initializes a default config file in the directory.
//   - JS/TS config already exists → error
//   - Only JSON/JSONC config exists → migrate to JS/TS config and delete JSON
//   - Nothing exists → create default JS/TS config
func InitDefaultConfig(directory string) error {
	// Check if JS/TS config already exists
	existingConfigs := []string{
		"rslint.config.ts", "rslint.config.mts",
		"rslint.config.js", "rslint.config.mjs",
	}
	for _, name := range existingConfigs {
		p := filepath.Join(directory, name)
		if _, err := os.Stat(p); err == nil {
			return fmt.Errorf("config file already exists: %s", name)
		}
	}

	// Check if JSON config exists → migrate
	for _, name := range []string{"rslint.json", "rslint.jsonc"} {
		p := filepath.Join(directory, name)
		if _, err := os.Stat(p); err == nil {
			return migrateJSONConfig(directory, name)
		}
	}

	// No config exists → create default
	return createDefaultConfig(directory)
}

// createDefaultConfig creates a fresh default config file.
func createDefaultConfig(directory string) error {
	tsconfigPath := filepath.Join(directory, "tsconfig.json")
	if _, err := os.Stat(tsconfigPath); err == nil {
		configPath := filepath.Join(directory, "rslint.config.ts")
		if err := os.WriteFile(configPath, []byte(defaultTSConfig), 0644); err != nil {
			return fmt.Errorf("failed to create rslint.config.ts: %w", err)
		}
		fmt.Println("Created rslint.config.ts with TypeScript recommended config.")
		return nil
	}

	var configName, content string
	if isESMPackage(directory) {
		configName = "rslint.config.js"
		content = defaultJSConfig
	} else {
		configName = "rslint.config.mjs"
		content = defaultJSConfig
	}
	configPath := filepath.Join(directory, configName)
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to create %s: %w", configName, err)
	}
	fmt.Printf("Created %s with JavaScript recommended config.\n", configName)
	return nil
}

// migrateJSONConfig reads a JSON config, converts it to a JS/TS config file,
// deduplicates rules against recommended presets, and deletes the old JSON file.
func migrateJSONConfig(directory, jsonFileName string) error {
	jsonPath := filepath.Join(directory, jsonFileName)
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", jsonFileName, err)
	}

	var entries RslintConfig
	if err := utils.ParseJSONC(data, &entries); err != nil {
		return fmt.Errorf("failed to parse %s: %w", jsonFileName, err)
	}

	if len(entries) == 0 {
		return fmt.Errorf("%s is empty", jsonFileName)
	}

	// Determine output file format
	hasTSConfig := false
	if _, err := os.Stat(filepath.Join(directory, "tsconfig.json")); err == nil {
		hasTSConfig = true
	}

	// When there is exactly one non-global-ignore entry with ignores,
	// extract ignores into a separate global ignore entry. Otherwise the
	// ignores would only apply to the override block, while the recommended
	// preset (inserted before it) would still lint those ignored files.
	entries = extractGlobalIgnores(entries)

	// Collect which preset imports are needed
	imports := newImportCollector()

	// Generate config entries
	var configEntries []string
	for _, entry := range entries {
		code := generateEntryCode(entry, imports, hasTSConfig)
		if code != "" {
			configEntries = append(configEntries, code)
		}
	}

	// Build the output file
	var buf strings.Builder
	buf.WriteString(imports.buildImportLine())
	buf.WriteString("\nexport default defineConfig([\n")
	for i, entry := range configEntries {
		buf.WriteString(entry)
		if i < len(configEntries)-1 {
			buf.WriteByte(',')
		}
		buf.WriteByte('\n')
	}
	buf.WriteString("]);\n")
	output := buf.String()

	// Determine output file name
	var configName string
	if hasTSConfig {
		configName = "rslint.config.ts"
	} else if isESMPackage(directory) {
		configName = "rslint.config.js"
	} else {
		configName = "rslint.config.mjs"
	}

	configPath := filepath.Join(directory, configName)
	if err := os.WriteFile(configPath, []byte(output), 0644); err != nil {
		return fmt.Errorf("failed to create %s: %w", configName, err)
	}

	// Delete old JSON config
	if err := os.Remove(jsonPath); err != nil {
		return fmt.Errorf("failed to delete %s: %w", jsonFileName, err)
	}

	fmt.Printf("Migrated %s → %s\n", jsonFileName, configName)
	return nil
}

// extractGlobalIgnores checks if there is exactly one non-global-ignore entry
// in the config. If that entry has ignores, they are extracted into a separate
// global ignore entry prepended to the slice, and removed from the original entry.
// This is necessary because in flat config, ignores inside an entry only apply to
// that entry, not to preset entries that come before it.
func extractGlobalIgnores(entries RslintConfig) RslintConfig {
	// Count non-global-ignore entries
	var nonIgnoreEntries int
	var targetIdx int
	for i, e := range entries {
		if !isGlobalIgnoreEntry(e) {
			nonIgnoreEntries++
			targetIdx = i
		}
	}

	// Only extract when there is exactly one non-global-ignore entry with ignores
	if nonIgnoreEntries != 1 || len(entries[targetIdx].Ignores) == 0 {
		return entries
	}

	// Create a global ignore entry and remove ignores from the original
	globalIgnore := ConfigEntry{Ignores: entries[targetIdx].Ignores}
	entries[targetIdx].Ignores = nil

	// Prepend global ignore entry
	result := make(RslintConfig, 0, len(entries)+1)
	result = append(result, globalIgnore)
	result = append(result, entries...)
	return result
}

// importCollector tracks which preset imports are needed.
type importCollector struct {
	needTS     bool
	needJS     bool
	needReact  bool
	needImport bool
}

func newImportCollector() *importCollector {
	return &importCollector{}
}

func (ic *importCollector) buildImportLine() string {
	symbols := []string{"defineConfig"}
	if ic.needTS {
		symbols = append(symbols, "ts")
	}
	if ic.needJS {
		symbols = append(symbols, "js")
	}
	if ic.needReact {
		symbols = append(symbols, "reactPlugin")
	}
	if ic.needImport {
		symbols = append(symbols, "importPlugin")
	}
	return fmt.Sprintf("import { %s } from '@rslint/core';\n", strings.Join(symbols, ", "))
}

// generateEntryCode generates the JS/TS config code for a single JSON config entry.
// It inserts recommended preset references and deduplicates rules.
func generateEntryCode(entry ConfigEntry, imports *importCollector, hasTSConfig bool) string {
	// Global ignore entry — output as-is
	if isGlobalIgnoreEntry(entry) {
		return formatIgnoreEntry(entry.Ignores)
	}

	// Determine which presets apply to this entry
	hasTS := containsPlugin(entry.Plugins, "@typescript-eslint")
	hasReact := containsPlugin(entry.Plugins, "react")
	hasImport := containsPlugin(entry.Plugins, "eslint-plugin-import")

	// Build the merged preset rules map for deduplication
	presetRules := buildPresetRules(hasTS, hasReact, hasImport)

	// Mark imports
	if hasTS {
		imports.needTS = true
	} else {
		imports.needJS = true
	}
	if hasReact {
		imports.needReact = true
	}
	if hasImport {
		imports.needImport = true
	}

	// Generate preset references
	var parts []string
	if hasTS {
		parts = append(parts, "  ts.configs.recommended")
	} else {
		parts = append(parts, "  js.configs.recommended")
	}
	if hasReact {
		parts = append(parts, "  reactPlugin.configs.recommended")
	}
	if hasImport {
		parts = append(parts, "  importPlugin.configs.recommended")
	}

	// Deduplicate rules
	remainingRules := deduplicateRules(entry.Rules, presetRules)

	// Build the user override entry (ignores, languageOptions, settings, remaining rules)
	overrideFields := buildOverrideFields(entry, remainingRules, hasTS)

	// If override entry has content, add it after presets
	if overrideFields != "" {
		parts = append(parts, overrideFields)
	}

	return strings.Join(parts, ",\n")
}

// containsPlugin checks if the plugin list contains the given plugin name.
func containsPlugin(plugins []string, name string) bool {
	for _, p := range plugins {
		if NormalizePluginName(p) == NormalizePluginName(name) {
			return true
		}
	}
	return false
}

// buildPresetRules merges all applicable preset rules into one map for deduplication.
func buildPresetRules(hasTS, hasReact, hasImport bool) map[string]string {
	merged := make(map[string]string)
	var base map[string]string
	if hasTS {
		base = tsRecommendedRules
	} else {
		base = jsRecommendedRules
	}
	for k, v := range base {
		merged[k] = v
	}
	if hasReact {
		for k, v := range reactRecommendedRules {
			merged[k] = v
		}
	}
	if hasImport {
		for k, v := range importRecommendedRules {
			merged[k] = v
		}
	}
	return merged
}

// deduplicateRules removes rules that match the preset severity exactly.
// Rules with array values (e.g., ["warn", options]) are always kept.
func deduplicateRules(userRules Rules, presetRules map[string]string) Rules {
	if len(userRules) == 0 {
		return nil
	}
	remaining := make(Rules)
	for name, value := range userRules {
		presetSeverity, inPreset := presetRules[name]
		if !inPreset {
			remaining[name] = value
			continue
		}
			userSeverity := extractSeverity(value)
		if userSeverity == "" {
			// Array form (e.g., ["warn", { ... }]) — always keep
			remaining[name] = value
			continue
		}
		if normalizeSeverity(userSeverity) != normalizeSeverity(presetSeverity) {
			remaining[name] = value
		}
		// Same severity as preset → skip (deduplicate)
	}
	if len(remaining) == 0 {
		return nil
	}
	return remaining
}

// extractSeverity extracts a severity string from a rule value.
// Returns "" for array values (which should always be kept).
func extractSeverity(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case float64:
		// JSON numbers are parsed as float64
		return strconv.Itoa(int(v))
	default:
		return ""
	}
}

// normalizeSeverity normalizes severity values: "error"/2 → "error", "warn"/1 → "warn", "off"/0 → "off".
func normalizeSeverity(s string) string {
	switch s {
	case "2":
		return "error"
	case "1":
		return "warn"
	case "0":
		return "off"
	default:
		return s
	}
}

// buildOverrideFields generates the user override object with remaining rules,
// ignores, languageOptions, and settings.
func buildOverrideFields(entry ConfigEntry, remainingRules Rules, hasTS bool) string {
	var fields []string

	// files (skip empty arrays)
	if len(entry.Files) > 0 {
		fields = append(fields, "    files: "+formatStringArray(entry.Files))
	}

	// ignores
	if len(entry.Ignores) > 0 {
		fields = append(fields, "    ignores: "+formatStringArray(entry.Ignores))
	}

	// languageOptions (skip if it matches preset defaults)
	if lo := formatLanguageOptions(entry.LanguageOptions, hasTS); lo != "" {
		fields = append(fields, lo)
	}

	// settings
	if len(entry.Settings) > 0 {
		fields = append(fields, "    settings: "+formatJSON(entry.Settings))
	}

	// rules
	if len(remainingRules) > 0 {
		fields = append(fields, formatRulesBlock(remainingRules))
	}

	if len(fields) == 0 {
		return ""
	}

	return "  {\n" + strings.Join(fields, ",\n") + ",\n  }"
}

// formatIgnoreEntry formats a global ignore entry.
func formatIgnoreEntry(ignores []string) string {
	if len(ignores) <= 3 {
		return "  { ignores: " + formatStringArray(ignores) + " }"
	}
	return "  {\n    ignores: " + formatStringArray(ignores) + ",\n  }"
}

// formatStringArray formats a string slice as a JS array literal.
func formatStringArray(arr []string) string {
	if len(arr) == 0 {
		return "[]"
	}
	parts := make([]string, len(arr))
	for i, s := range arr {
		parts[i] = "'" + escapeJSString(s) + "'"
	}
	if len(arr) <= 3 {
		return "[" + strings.Join(parts, ", ") + "]"
	}
	// Multi-line for longer arrays
	var buf strings.Builder
	buf.WriteString("[\n")
	for _, p := range parts {
		buf.WriteString("      ")
		buf.WriteString(p)
		buf.WriteString(",\n")
	}
	buf.WriteString("    ]")
	return buf.String()
}

// formatRulesBlock formats the rules as a JS object block.
func formatRulesBlock(rules Rules) string {
	if len(rules) == 0 {
		return ""
	}

	// Sort rule names for deterministic output
	names := make([]string, 0, len(rules))
	for name := range rules {
		names = append(names, name)
	}
	sort.Strings(names)

	var lines []string
	for _, name := range names {
		value := rules[name]
		lines = append(lines, fmt.Sprintf("      '%s': %s", escapeJSString(name), formatRuleValue(value)))
	}

	return "    rules: {\n" + strings.Join(lines, ",\n") + ",\n    }"
}

// formatRuleValue formats a rule value (string, number, or array) as JS.
func formatRuleValue(value interface{}) string {
	switch v := value.(type) {
	case string:
		return fmt.Sprintf("'%s'", escapeJSString(v))
	case float64:
		// Convert numeric severity to string form
		return "'" + normalizeSeverity(strconv.Itoa(int(v))) + "'"
	default:
		// For arrays or complex values, marshal to JSON.
		// JSON output uses double quotes, which is valid JS syntax.
		data, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprintf("%v", v)
		}
		return string(data)
	}
}

// formatLanguageOptions formats languageOptions, skipping fields that match preset defaults.
// For TS preset: projectService defaults to true, so only output if explicitly false or if project is set.
// For JS preset: no defaults, output everything.
func formatLanguageOptions(lo *LanguageOptions, hasTS bool) string {
	if lo == nil || lo.ParserOptions == nil {
		return ""
	}
	po := lo.ParserOptions

	var poFields []string

	// projectService
	if po.ProjectService != nil {
		defaultPS := hasTS // TS preset defaults to true
		if *po.ProjectService != defaultPS {
			poFields = append(poFields, "        projectService: "+strconv.FormatBool(*po.ProjectService))
		}
	}

	// project paths
	if len(po.Project) > 0 {
		poFields = append(poFields, "        project: "+formatStringArray([]string(po.Project)))
	}

	if len(poFields) == 0 {
		return ""
	}

	return "    languageOptions: {\n      parserOptions: {\n" +
		strings.Join(poFields, ",\n") + ",\n" +
		"      },\n    }"
}

// formatJSON marshals a value to a compact JSON string.
func formatJSON(v interface{}) string {
	data, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(data)
}

// escapeJSString escapes single quotes and backslashes in a string for JS output.
func escapeJSString(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "'", "\\'")
	return s
}
