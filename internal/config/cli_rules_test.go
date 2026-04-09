package config

import (
	"testing"

	"gotest.tools/v3/assert"
)

// ---------------------------------------------------------------------------
// ParseCLIRuleFlag — parsing
// ---------------------------------------------------------------------------

func TestParseCLIRuleFlag_SeverityOnly(t *testing.T) {
	tests := []struct {
		input    string
		wantName string
		wantVal  interface{}
	}{
		{"no-console: error", "no-console", "error"},
		{"no-console: warn", "no-console", "warn"},
		{"no-console: off", "no-console", "off"},
		{"no-debugger: error", "no-debugger", "error"},
	}
	for _, tt := range tests {
		name, val, err := ParseCLIRuleFlag(tt.input)
		assert.NilError(t, err, tt.input)
		assert.Equal(t, name, tt.wantName)
		assert.Equal(t, val, tt.wantVal)
	}
}

func TestParseCLIRuleFlag_PluginRules(t *testing.T) {
	tests := []struct {
		input    string
		wantName string
		wantVal  interface{}
	}{
		{"@typescript-eslint/no-explicit-any: off", "@typescript-eslint/no-explicit-any", "off"},
		{"import/no-unresolved: error", "import/no-unresolved", "error"},
		{"react/jsx-no-target-blank: warn", "react/jsx-no-target-blank", "warn"},
		{"jest/no-disabled-tests: error", "jest/no-disabled-tests", "error"},
	}
	for _, tt := range tests {
		name, val, err := ParseCLIRuleFlag(tt.input)
		assert.NilError(t, err, tt.input)
		assert.Equal(t, name, tt.wantName)
		assert.Equal(t, val, tt.wantVal)
	}
}

func TestParseCLIRuleFlag_WhitespaceVariations(t *testing.T) {
	tests := []struct {
		input    string
		wantName string
		wantVal  interface{}
	}{
		{"no-console:error", "no-console", "error"},           // no space after colon
		{"no-console:  error", "no-console", "error"},         // multiple spaces after colon
		{"  no-console  :  error  ", "no-console", "error"},   // leading/trailing spaces
		{"no-console:\terror", "no-console", "error"},         // tab after colon
		{"@typescript-eslint/no-explicit-any:  warn", "@typescript-eslint/no-explicit-any", "warn"}, // plugin rule with extra space
	}
	for _, tt := range tests {
		name, val, err := ParseCLIRuleFlag(tt.input)
		assert.NilError(t, err, tt.input)
		assert.Equal(t, name, tt.wantName)
		assert.Equal(t, val, tt.wantVal)
	}
}

func TestParseCLIRuleFlag_ArraySeverityOnly(t *testing.T) {
	name, val, err := ParseCLIRuleFlag(`no-console: ["error"]`)
	assert.NilError(t, err)
	assert.Equal(t, name, "no-console")

	arr, ok := val.([]interface{})
	assert.Assert(t, ok)
	assert.Equal(t, len(arr), 1)
	assert.Equal(t, arr[0], "error")
}

func TestParseCLIRuleFlag_ArrayWithObjectOptions(t *testing.T) {
	name, val, err := ParseCLIRuleFlag(`no-console: ["warn", {"allow": ["warn", "error"]}]`)
	assert.NilError(t, err)
	assert.Equal(t, name, "no-console")

	arr, ok := val.([]interface{})
	assert.Assert(t, ok, "expected []interface{}")
	assert.Equal(t, len(arr), 2)
	assert.Equal(t, arr[0], "warn")

	opts, ok := arr[1].(map[string]interface{})
	assert.Assert(t, ok, "expected map for options")
	allow, ok := opts["allow"].([]interface{})
	assert.Assert(t, ok)
	assert.Equal(t, len(allow), 2)
	assert.Equal(t, allow[0], "warn")
	assert.Equal(t, allow[1], "error")
}

func TestParseCLIRuleFlag_ArrayWithInternalWhitespace(t *testing.T) {
	// Extra spaces inside JSON value — json.Unmarshal handles this fine
	input := `no-console:  ["error",  { "allow" :  ["warn",  "error"] }]`
	name, val, err := ParseCLIRuleFlag(input)
	assert.NilError(t, err)
	assert.Equal(t, name, "no-console")

	arr, ok := val.([]interface{})
	assert.Assert(t, ok)
	assert.Equal(t, len(arr), 2)
	assert.Equal(t, arr[0], "error")

	opts, ok := arr[1].(map[string]interface{})
	assert.Assert(t, ok)
	allow, ok := opts["allow"].([]interface{})
	assert.Assert(t, ok)
	assert.Equal(t, len(allow), 2)
}

func TestParseCLIRuleFlag_ArrayWithStringOption(t *testing.T) {
	// e.g. no-inner-declarations: ["error", "both"]
	name, val, err := ParseCLIRuleFlag(`no-inner-declarations: ["error", "both"]`)
	assert.NilError(t, err)
	assert.Equal(t, name, "no-inner-declarations")

	arr, ok := val.([]interface{})
	assert.Assert(t, ok)
	assert.Equal(t, len(arr), 2)
	assert.Equal(t, arr[0], "error")
	assert.Equal(t, arr[1], "both")
}

func TestParseCLIRuleFlag_ArrayWithMultipleOptions(t *testing.T) {
	// e.g. ["error", "both", {"blockScopedFunctions": "disallow"}]
	name, val, err := ParseCLIRuleFlag(`no-inner-declarations: ["error", "both", {"blockScopedFunctions": "disallow"}]`)
	assert.NilError(t, err)
	assert.Equal(t, name, "no-inner-declarations")

	arr, ok := val.([]interface{})
	assert.Assert(t, ok)
	assert.Equal(t, len(arr), 3)
	assert.Equal(t, arr[0], "error")
	assert.Equal(t, arr[1], "both")

	opts, ok := arr[2].(map[string]interface{})
	assert.Assert(t, ok)
	assert.Equal(t, opts["blockScopedFunctions"], "disallow")
}

func TestParseCLIRuleFlag_ArrayWithDeeplyNestedOptions(t *testing.T) {
	input := `my-rule: ["error", {"paths": [{"name": "lodash", "importNames": ["default"], "message": "Use named imports"}]}]`
	name, val, err := ParseCLIRuleFlag(input)
	assert.NilError(t, err)
	assert.Equal(t, name, "my-rule")

	arr, ok := val.([]interface{})
	assert.Assert(t, ok)
	assert.Equal(t, len(arr), 2)
	assert.Equal(t, arr[0], "error")

	opts, ok := arr[1].(map[string]interface{})
	assert.Assert(t, ok)
	paths, ok := opts["paths"].([]interface{})
	assert.Assert(t, ok)
	assert.Equal(t, len(paths), 1)

	entry, ok := paths[0].(map[string]interface{})
	assert.Assert(t, ok)
	assert.Equal(t, entry["name"], "lodash")
	assert.Equal(t, entry["message"], "Use named imports")
	importNames, ok := entry["importNames"].([]interface{})
	assert.Assert(t, ok)
	assert.Equal(t, len(importNames), 1)
	assert.Equal(t, importNames[0], "default")
}

func TestParseCLIRuleFlag_ArrayWithBooleanAndNumericOptions(t *testing.T) {
	input := `indent: ["error", 2, {"SwitchCase": 1, "flatTernaryExpressions": false}]`
	name, val, err := ParseCLIRuleFlag(input)
	assert.NilError(t, err)
	assert.Equal(t, name, "indent")

	arr, ok := val.([]interface{})
	assert.Assert(t, ok)
	assert.Equal(t, len(arr), 3)
	assert.Equal(t, arr[0], "error")
	assert.Equal(t, arr[1], float64(2)) // JSON numbers are float64

	opts, ok := arr[2].(map[string]interface{})
	assert.Assert(t, ok)
	assert.Equal(t, opts["SwitchCase"], float64(1))
	assert.Equal(t, opts["flatTernaryExpressions"], false)
}

func TestParseCLIRuleFlag_ArrayPluginRuleWithOptions(t *testing.T) {
	input := `@typescript-eslint/no-unused-vars: ["warn", {"argsIgnorePattern": "^_"}]`
	name, val, err := ParseCLIRuleFlag(input)
	assert.NilError(t, err)
	assert.Equal(t, name, "@typescript-eslint/no-unused-vars")

	arr, ok := val.([]interface{})
	assert.Assert(t, ok)
	assert.Equal(t, len(arr), 2)
	assert.Equal(t, arr[0], "warn")

	opts, ok := arr[1].(map[string]interface{})
	assert.Assert(t, ok)
	assert.Equal(t, opts["argsIgnorePattern"], "^_")
}

// ---------------------------------------------------------------------------
// ParseCLIRuleFlag — error cases
// ---------------------------------------------------------------------------

func TestParseCLIRuleFlag_InvalidFormat(t *testing.T) {
	tests := []struct {
		input string
		desc  string
	}{
		{"no-console", "no colon separator"},
		{": error", "empty rule name"},
		{"  : error", "whitespace-only rule name"},
		{"no-console:", "empty value after colon"},
		{"no-console:  ", "whitespace-only value"},
		{`no-console: [broken`, "unclosed JSON array"},
		{`no-console: ["error",]`, "trailing comma in JSON"},
	}
	for _, tt := range tests {
		_, _, err := ParseCLIRuleFlag(tt.input)
		assert.Assert(t, err != nil, "expected error for %q (%s)", tt.input, tt.desc)
	}
}

// ---------------------------------------------------------------------------
// BuildCLIRuleEntry — construction
// ---------------------------------------------------------------------------

func TestBuildCLIRuleEntry_Empty(t *testing.T) {
	entry, err := BuildCLIRuleEntry(nil)
	assert.NilError(t, err)
	assert.Assert(t, entry == nil)

	entry, err = BuildCLIRuleEntry([]string{})
	assert.NilError(t, err)
	assert.Assert(t, entry == nil)
}

func TestBuildCLIRuleEntry_SingleRule(t *testing.T) {
	entry, err := BuildCLIRuleEntry([]string{"no-console: error"})
	assert.NilError(t, err)
	assert.Assert(t, entry != nil)
	assert.Assert(t, len(entry.Files) == 0, "should have no files (matches all)")
	assert.Assert(t, len(entry.Ignores) == 0, "should have no ignores")
	assert.Assert(t, len(entry.Plugins) == 0, "should have no plugins")
	assert.Assert(t, entry.LanguageOptions == nil, "should have no language options")
	assert.Equal(t, entry.Rules["no-console"], "error")
}

func TestBuildCLIRuleEntry_MultipleRules(t *testing.T) {
	entry, err := BuildCLIRuleEntry([]string{
		"no-console: off",
		"no-debugger: error",
		"@typescript-eslint/no-explicit-any: warn",
	})
	assert.NilError(t, err)
	assert.Assert(t, entry != nil)
	assert.Equal(t, entry.Rules["no-console"], "off")
	assert.Equal(t, entry.Rules["no-debugger"], "error")
	assert.Equal(t, entry.Rules["@typescript-eslint/no-explicit-any"], "warn")
}

func TestBuildCLIRuleEntry_LaterOverridesEarlier(t *testing.T) {
	entry, err := BuildCLIRuleEntry([]string{
		"no-console: error",
		"no-console: off",
	})
	assert.NilError(t, err)
	assert.Equal(t, entry.Rules["no-console"], "off")
}

func TestBuildCLIRuleEntry_ArrayValuePreserved(t *testing.T) {
	entry, err := BuildCLIRuleEntry([]string{
		`no-console: ["warn", {"allow": ["warn"]}]`,
	})
	assert.NilError(t, err)

	arr, ok := entry.Rules["no-console"].([]interface{})
	assert.Assert(t, ok, "expected array value preserved in Rules map")
	assert.Equal(t, len(arr), 2)
}

func TestBuildCLIRuleEntry_MixedStringAndArrayValues(t *testing.T) {
	entry, err := BuildCLIRuleEntry([]string{
		"no-debugger: error",
		`no-console: ["warn", {"allow": ["warn"]}]`,
		"@typescript-eslint/no-explicit-any: off",
	})
	assert.NilError(t, err)

	// String value
	assert.Equal(t, entry.Rules["no-debugger"], "error")
	// Array value
	_, ok := entry.Rules["no-console"].([]interface{})
	assert.Assert(t, ok)
	// Plugin rule string value
	assert.Equal(t, entry.Rules["@typescript-eslint/no-explicit-any"], "off")
}

func TestBuildCLIRuleEntry_PropagatesParseError(t *testing.T) {
	_, err := BuildCLIRuleEntry([]string{"bad-format"})
	assert.Assert(t, err != nil)
}

func TestBuildCLIRuleEntry_StopsAtFirstError(t *testing.T) {
	_, err := BuildCLIRuleEntry([]string{
		"no-console: error",
		"bad-format",
		"no-debugger: error",
	})
	assert.Assert(t, err != nil)
}

// ---------------------------------------------------------------------------
// Integration — CLI entry merged with config via GetConfigForFile
// ---------------------------------------------------------------------------

func TestIntegration_CLIOverridesPerFileConfig(t *testing.T) {
	config := RslintConfig{
		{
			Files: []string{"**/*.ts"},
			Rules: Rules{"no-console": "error", "no-debugger": "error"},
		},
		{
			Files: []string{"**/*.test.ts"},
			Rules: Rules{"no-console": "off"},
		},
	}

	cliEntry, err := BuildCLIRuleEntry([]string{"no-console: warn"})
	assert.NilError(t, err)
	config = append(config, *cliEntry)

	// .ts file: CLI overrides "error" to "warn"
	merged := config.GetConfigForFile("src/app.ts", "")
	assert.Assert(t, merged != nil)
	assert.Equal(t, merged.Rules["no-console"].Level, "warn")
	assert.Equal(t, merged.Rules["no-debugger"].Level, "error") // untouched

	// .test.ts file: CLI overrides "off" to "warn"
	merged = config.GetConfigForFile("src/app.test.ts", "")
	assert.Assert(t, merged != nil)
	assert.Equal(t, merged.Rules["no-console"].Level, "warn")
}

func TestIntegration_CLIAddsNewRule(t *testing.T) {
	// Config only has no-console; CLI adds no-debugger which wasn't in config
	config := RslintConfig{
		{
			Files: []string{"**/*.ts"},
			Rules: Rules{"no-console": "error"},
		},
	}

	cliEntry, err := BuildCLIRuleEntry([]string{"no-debugger: warn"})
	assert.NilError(t, err)
	config = append(config, *cliEntry)

	merged := config.GetConfigForFile("src/app.ts", "")
	assert.Assert(t, merged != nil)
	assert.Equal(t, merged.Rules["no-console"].Level, "error")
	assert.Equal(t, merged.Rules["no-debugger"].Level, "warn")
}

func TestIntegration_CLITurnsOffRule(t *testing.T) {
	config := RslintConfig{
		{
			Files: []string{"**/*.ts"},
			Rules: Rules{"no-console": "error", "no-debugger": "error"},
		},
	}

	cliEntry, err := BuildCLIRuleEntry([]string{"no-console: off"})
	assert.NilError(t, err)
	config = append(config, *cliEntry)

	merged := config.GetConfigForFile("src/app.ts", "")
	assert.Assert(t, merged != nil)
	assert.Equal(t, merged.Rules["no-console"].Level, "off")
	assert.Assert(t, !merged.Rules["no-console"].IsEnabled())
	assert.Equal(t, merged.Rules["no-debugger"].Level, "error") // untouched
}

func TestIntegration_CLIArrayOverridesConfigString(t *testing.T) {
	// Config has string "error", CLI overrides with array ["warn", options]
	config := RslintConfig{
		{
			Rules: Rules{"no-console": "error"},
		},
	}

	cliEntry, err := BuildCLIRuleEntry([]string{`no-console: ["warn", {"allow": ["warn"]}]`})
	assert.NilError(t, err)
	config = append(config, *cliEntry)

	merged := config.GetConfigForFile("src/app.ts", "")
	assert.Assert(t, merged != nil)
	assert.Equal(t, merged.Rules["no-console"].Level, "warn")
	opts, ok := merged.Rules["no-console"].Options.(map[string]interface{})
	assert.Assert(t, ok)
	allow, ok := opts["allow"].([]interface{})
	assert.Assert(t, ok)
	assert.Equal(t, len(allow), 1)
	assert.Equal(t, allow[0], "warn")
}

func TestIntegration_CLIStringOverridesConfigArray(t *testing.T) {
	// Config has array with options, CLI overrides with plain "off"
	config := RslintConfig{
		{
			Rules: Rules{"no-console": []interface{}{"error", map[string]interface{}{"allow": []interface{}{"warn"}}}},
		},
	}

	cliEntry, err := BuildCLIRuleEntry([]string{"no-console: off"})
	assert.NilError(t, err)
	config = append(config, *cliEntry)

	merged := config.GetConfigForFile("src/app.ts", "")
	assert.Assert(t, merged != nil)
	assert.Equal(t, merged.Rules["no-console"].Level, "off")
	assert.Assert(t, merged.Rules["no-console"].Options == nil)
}

func TestIntegration_CLIDoesNotAffectGloballyIgnoredFiles(t *testing.T) {
	config := RslintConfig{
		{
			// Global ignore entry (ignores only, no files/rules)
			Ignores: []string{"dist/**"},
		},
		{
			Rules: Rules{"no-console": "error"},
		},
	}

	cliEntry, err := BuildCLIRuleEntry([]string{"no-console: warn"})
	assert.NilError(t, err)
	config = append(config, *cliEntry)

	// dist/ files are globally ignored — CLI can't override that
	merged := config.GetConfigForFile("dist/bundle.js", "")
	assert.Assert(t, merged == nil, "globally ignored file should remain ignored")

	// Non-ignored file gets the CLI override
	merged = config.GetConfigForFile("src/app.ts", "")
	assert.Assert(t, merged != nil)
	assert.Equal(t, merged.Rules["no-console"].Level, "warn")
}

func TestIntegration_CLIWithMultipleConfigEntries(t *testing.T) {
	// Three config entries: base → ts-specific → test-specific
	// CLI should override all of them
	config := RslintConfig{
		{
			// Base: applies to all
			Rules: Rules{
				"no-console":  "error",
				"no-debugger": "error",
				"no-var":      "warn",
			},
		},
		{
			Files: []string{"**/*.ts"},
			Rules: Rules{"no-console": "warn"},
		},
		{
			Files: []string{"**/*.test.ts"},
			Rules: Rules{"no-console": "off", "no-debugger": "off"},
		},
	}

	cliEntry, err := BuildCLIRuleEntry([]string{
		"no-console: error",
		"no-debugger: warn",
	})
	assert.NilError(t, err)
	config = append(config, *cliEntry)

	// .js file: base + CLI
	merged := config.GetConfigForFile("src/app.js", "")
	assert.Assert(t, merged != nil)
	assert.Equal(t, merged.Rules["no-console"].Level, "error")
	assert.Equal(t, merged.Rules["no-debugger"].Level, "warn")
	assert.Equal(t, merged.Rules["no-var"].Level, "warn") // untouched

	// .ts file: base + ts-specific + CLI
	merged = config.GetConfigForFile("src/app.ts", "")
	assert.Assert(t, merged != nil)
	assert.Equal(t, merged.Rules["no-console"].Level, "error")  // CLI overrides ts-specific "warn"
	assert.Equal(t, merged.Rules["no-debugger"].Level, "warn")  // CLI overrides base "error"

	// .test.ts file: base + ts-specific + test-specific + CLI
	merged = config.GetConfigForFile("src/app.test.ts", "")
	assert.Assert(t, merged != nil)
	assert.Equal(t, merged.Rules["no-console"].Level, "error")  // CLI overrides test-specific "off"
	assert.Equal(t, merged.Rules["no-debugger"].Level, "warn")  // CLI overrides test-specific "off"
}

func TestIntegration_CLIOnlyMatchesFilesMatchedByConfig(t *testing.T) {
	// Config only targets .ts files; .js files have no matching entry
	config := RslintConfig{
		{
			Files: []string{"**/*.ts"},
			Rules: Rules{"no-console": "error"},
		},
	}

	cliEntry, err := BuildCLIRuleEntry([]string{"no-console: warn"})
	assert.NilError(t, err)
	config = append(config, *cliEntry)

	// .ts file: matched by both config and CLI entry
	merged := config.GetConfigForFile("src/app.ts", "")
	assert.Assert(t, merged != nil)
	assert.Equal(t, merged.Rules["no-console"].Level, "warn")

	// .js file: CLI entry (no files) matches, so entryMatched = true
	merged = config.GetConfigForFile("src/app.js", "")
	assert.Assert(t, merged != nil)
	assert.Equal(t, merged.Rules["no-console"].Level, "warn")
}

func TestIntegration_CLIWithEntryIgnores(t *testing.T) {
	config := RslintConfig{
		{
			Files:   []string{"**/*.ts"},
			Ignores: []string{"**/*.generated.ts"},
			Rules:   Rules{"no-console": "error"},
		},
	}

	cliEntry, err := BuildCLIRuleEntry([]string{"no-console: warn"})
	assert.NilError(t, err)
	config = append(config, *cliEntry)

	// Normal .ts file: CLI overrides
	merged := config.GetConfigForFile("src/app.ts", "")
	assert.Assert(t, merged != nil)
	assert.Equal(t, merged.Rules["no-console"].Level, "warn")

	// .generated.ts: ignored by entry1, but CLI entry (no files/ignores) still matches
	merged = config.GetConfigForFile("src/app.generated.ts", "")
	assert.Assert(t, merged != nil)
	assert.Equal(t, merged.Rules["no-console"].Level, "warn")
}
