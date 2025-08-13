package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bmatcuk/doublestar/v4"
)

func TestMatchesPattern(t *testing.T) {
	tests := []struct {
		filePath string
		pattern  string
		expected bool
		name     string
	}{
		{
			filePath: "src/file.ts",
			pattern:  "**/*.ts",
			expected: true,
			name:     "recursive TypeScript files",
		},
		{
			filePath: "src/file.js",
			pattern:  "**/*.ts",
			expected: false,
			name:     "non-matching extension",
		},
		{
			filePath: "test/file.spec.ts",
			pattern:  "**/test/**",
			expected: true,
			name:     "directory pattern",
		},
		{
			filePath: "src/test/file.ts",
			pattern:  "**/test/**",
			expected: true,
			name:     "nested directory pattern",
		},
		{
			filePath: "test.ts",
			pattern:  "*.ts",
			expected: true,
			name:     "single level wildcard",
		},
		{
			filePath: "src/test.ts",
			pattern:  "*.ts",
			expected: false,
			name:     "single level wildcard should not match nested",
		},
		{
			filePath: "node_modules/package/file.ts",
			pattern:  "node_modules/**",
			expected: true,
			name:     "node_modules pattern",
		},
		{
			filePath: "dist/output.js",
			pattern:  "dist/**",
			expected: true,
			name:     "directory pattern with recursive wildcard",
		},
		{
			filePath: "exact/file.ts",
			pattern:  "exact/file.ts",
			expected: true,
			name:     "exact match",
		},
		{
			filePath: "src/components/Button.tsx",
			pattern:  "src/**/*.tsx",
			expected: true,
			name:     "nested recursive pattern with extension",
		},
		{
			filePath: "lib/utils/helper.js",
			pattern:  "lib/**/helper.js",
			expected: true,
			name:     "specific file in nested directory",
		},
		{
			filePath: "docs/readme.md",
			pattern:  "docs/*.md",
			expected: true,
			name:     "single level directory with extension",
		},
		{
			filePath: "src/docs/readme.md",
			pattern:  "docs/*.md",
			expected: false,
			name:     "single level should not match nested path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := doublestar.Match(tt.pattern, tt.filePath)
			if err != nil {
				t.Errorf("doublestar.Match(%q, %q) error: %v", tt.pattern, tt.filePath, err)
				return
			}
			if result != tt.expected {
				t.Errorf("doublestar.Match(%q, %q) = %v, expected %v", tt.pattern, tt.filePath, result, tt.expected)
			}
		})
	}
}

func TestIsFileIgnored(t *testing.T) {
	ignorePatterns := []string{
		"node_modules/**",
		"dist/**",
		"**/*.test.ts",
		"temp/**",
	}

	tests := []struct {
		filePath string
		expected bool
		name     string
	}{
		{
			filePath: "src/file.ts",
			expected: false,
			name:     "normal source file should not be ignored",
		},
		{
			filePath: "node_modules/package/file.ts",
			expected: true,
			name:     "node_modules file should be ignored",
		},
		{
			filePath: "dist/bundle.js",
			expected: true,
			name:     "dist file should be ignored",
		},
		{
			filePath: "src/component.test.ts",
			expected: true,
			name:     "test file should be ignored",
		},
		{
			filePath: "temp/cache.ts",
			expected: true,
			name:     "temp directory should be ignored",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isFileIgnored(tt.filePath, ignorePatterns)
			if result != tt.expected {
				t.Errorf("isFileIgnored(%q, %v) = %v, expected %v", tt.filePath, ignorePatterns, result, tt.expected)
			}
		})
	}
}

func TestGetRulesForFileWithIgnores(t *testing.T) {
	config := RslintConfig{
		{
			Language: "typescript",
			Files:    []string{"**/*.ts", "**/*.tsx"},
			Ignores:  []string{"**/*.test.ts", "node_modules/**"},
			Rules: Rules{
				"@typescript-eslint/no-unused-vars": "error",
			},
		},
	}

	tests := []struct {
		filePath        string
		shouldHaveRules bool
		name            string
	}{
		{
			filePath:        "src/component.ts",
			shouldHaveRules: true,
			name:            "normal TypeScript file should have rules",
		},
		{
			filePath:        "src/component.test.ts",
			shouldHaveRules: false,
			name:            "test file should be ignored",
		},
		{
			filePath:        "node_modules/package/file.ts",
			shouldHaveRules: false,
			name:            "node_modules file should be ignored",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rules := config.GetRulesForFile(tt.filePath)
			hasRules := len(rules) > 0

			if hasRules != tt.shouldHaveRules {
				t.Errorf("GetRulesForFile(%q) hasRules = %v, expected %v (rules: %v)",
					tt.filePath, hasRules, tt.shouldHaveRules, rules)
			}
		})
	}
}

func TestGetImportPluginRules(t *testing.T) {
	baseConfig := ConfigEntry{
		Language: "typescript",
		Files:    []string{"**/*.ts", "**/*.tsx"},
		Ignores:  []string{"**/*.test.ts", "node_modules/**"},
		Rules:    Rules{},
	}

	tests := []struct {
		rulesCount int
		plugin     string
	}{
		{
			rulesCount: 1,
			plugin:     "eslint-plugin-import",
		},
		{
			rulesCount: 0,
			plugin:     "eslint-plugin-import/recommended",
		},
	}

	for _, tt := range tests {
		t.Run(tt.plugin, func(t *testing.T) {
			config := RslintConfig{
				baseConfig,
				{
					Plugins: []string{tt.plugin},
				},
			}
			rules := config.GetRulesForFile("foo.ts")

			if len(rules) != tt.rulesCount {
				t.Errorf("GetRulesForFile(foo.ts) with plugin %v ruleCount = %v, expected %v (rules: %v)",
					tt.plugin, len(rules), tt.rulesCount, rules)
			}
		})
	}
}

func TestParseArrayRuleConfig(t *testing.T) {
	tests := []struct {
		name            string
		input           []interface{}
		expectedLevel   string
		expectedOptions map[string]interface{}
		expectedNil     bool
	}{
		{
			name:        "empty array",
			input:       []interface{}{},
			expectedNil: true,
		},
		{
			name:            "error level only",
			input:           []interface{}{"error"},
			expectedLevel:   "error",
			expectedOptions: map[string]interface{}{},
		},
		{
			name:            "warn level only",
			input:           []interface{}{"warn"},
			expectedLevel:   "warn",
			expectedOptions: map[string]interface{}{},
		},
		{
			name:            "off level only",
			input:           []interface{}{"off"},
			expectedLevel:   "off",
			expectedOptions: map[string]interface{}{},
		},
		{
			name:            "error with options",
			input:           []interface{}{"error", map[string]interface{}{"option1": "value1", "option2": 42}},
			expectedLevel:   "error",
			expectedOptions: map[string]interface{}{"option1": "value1", "option2": 42},
		},
		{
			name:            "warn with options",
			input:           []interface{}{"warn", map[string]interface{}{"allowExpressions": true}},
			expectedLevel:   "warn",
			expectedOptions: map[string]interface{}{"allowExpressions": true},
		},
		{
			name:            "error with null options",
			input:           []interface{}{"error", nil},
			expectedLevel:   "error",
			expectedOptions: map[string]interface{}{},
		},
		{
			name:            "invalid options type",
			input:           []interface{}{"warn", "invalid"},
			expectedLevel:   "warn",
			expectedOptions: map[string]interface{}{},
		},
		{
			name:        "invalid level type",
			input:       []interface{}{123},
			expectedNil: true,
		},
		{
			name:            "extra elements ignored",
			input:           []interface{}{"error", map[string]interface{}{"test": true}, "extra", 123},
			expectedLevel:   "error",
			expectedOptions: map[string]interface{}{"test": true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseArrayRuleConfig(tt.input)

			if tt.expectedNil {
				if result != nil {
					t.Errorf("expected nil result, got %+v", result)
				}
				return
			}

			if result == nil {
				t.Error("expected non-nil result")
				return
			}

			if result.Level != tt.expectedLevel {
				t.Errorf("expected level %q, got %q", tt.expectedLevel, result.Level)
			}

			if len(result.Options) != len(tt.expectedOptions) {
				t.Errorf("expected %d options, got %d", len(tt.expectedOptions), len(result.Options))
				return
			}

			for key, expectedValue := range tt.expectedOptions {
				actualValue, exists := result.Options[key]
				if !exists {
					t.Errorf("expected option %q not found", key)
					continue
				}
				if actualValue != expectedValue {
					t.Errorf("expected option %q = %v, got %v", key, expectedValue, actualValue)
				}
			}
		})
	}
}

func TestRuleConfigMethods(t *testing.T) {
	t.Run("GetOptions with nil config", func(t *testing.T) {
		var rc *RuleConfig
		options := rc.GetOptions()
		if options == nil {
			t.Error("expected non-nil options map")
		}
		if len(options) != 0 {
			t.Errorf("expected empty options map, got %d items", len(options))
		}
	})

	t.Run("GetOptions with nil Options field", func(t *testing.T) {
		rc := &RuleConfig{Level: "error"}
		options := rc.GetOptions()
		if options == nil {
			t.Error("expected non-nil options map")
		}
		if len(options) != 0 {
			t.Errorf("expected empty options map, got %d items", len(options))
		}
	})

	t.Run("GetOptions with existing options", func(t *testing.T) {
		expectedOptions := map[string]interface{}{"test": true}
		rc := &RuleConfig{
			Level:   "warn",
			Options: expectedOptions,
		}
		options := rc.GetOptions()
		if len(options) != 1 {
			t.Errorf("expected 1 option, got %d", len(options))
		}
		if options["test"] != true {
			t.Error("expected test option to be true")
		}
	})

	t.Run("SetOptions", func(t *testing.T) {
		rc := &RuleConfig{Level: "error"}
		newOptions := map[string]interface{}{"newOption": "value"}
		rc.SetOptions(newOptions)

		if len(rc.Options) != 1 {
			t.Errorf("expected 1 option after SetOptions, got %d", len(rc.Options))
		}
		if rc.Options["newOption"] != "value" {
			t.Error("expected newOption to be 'value'")
		}
	})

	t.Run("SetOptions with nil config", func(t *testing.T) {
		var rc *RuleConfig
		// Should not panic
		rc.SetOptions(map[string]interface{}{"test": true})
		// Config should remain nil (this test verifies the method handles nil gracefully)
	})
}

func TestGetRulesForFileWithArrayConfig(t *testing.T) {
	config := RslintConfig{
		{
			Language: "javascript",
			Files:    []string{},
			Rules: map[string]interface{}{
				"rule1": "error",
				"rule2": []interface{}{"warn"},
				"rule3": []interface{}{"error", map[string]interface{}{"option1": "value1"}},
				"rule4": []interface{}{"off"},
				"rule5": map[string]interface{}{
					"level":   "warn",
					"options": map[string]interface{}{"option2": "value2"},
				},
			},
		},
	}

	rules := config.GetRulesForFile("test.ts")

	// Test rule1 - simple string config
	if rule1, exists := rules["rule1"]; exists {
		if rule1.Level != "error" {
			t.Errorf("expected rule1 level 'error', got %q", rule1.Level)
		}
		if len(rule1.GetOptions()) != 0 {
			t.Errorf("expected rule1 to have no options, got %d", len(rule1.GetOptions()))
		}
	} else {
		t.Error("expected rule1 to exist")
	}

	// Test rule2 - array with level only
	if rule2, exists := rules["rule2"]; exists {
		if rule2.Level != "warn" {
			t.Errorf("expected rule2 level 'warn', got %q", rule2.Level)
		}
		if len(rule2.GetOptions()) != 0 {
			t.Errorf("expected rule2 to have no options, got %d", len(rule2.GetOptions()))
		}
	} else {
		t.Error("expected rule2 to exist")
	}

	// Test rule3 - array with level and options
	if rule3, exists := rules["rule3"]; exists {
		if rule3.Level != "error" {
			t.Errorf("expected rule3 level 'error', got %q", rule3.Level)
		}
		options := rule3.GetOptions()
		if len(options) != 1 {
			t.Errorf("expected rule3 to have 1 option, got %d", len(options))
		}
		if options["option1"] != "value1" {
			t.Errorf("expected rule3 option1 to be 'value1', got %v", options["option1"])
		}
	} else {
		t.Error("expected rule3 to exist")
	}

	// Test rule4 - array with "off" (should not exist in enabled rules)
	if _, exists := rules["rule4"]; exists {
		t.Error("expected rule4 to be disabled (not exist in enabled rules)")
	}

	// Test rule5 - object config
	if rule5, exists := rules["rule5"]; exists {
		if rule5.Level != "warn" {
			t.Errorf("expected rule5 level 'warn', got %q", rule5.Level)
		}
		options := rule5.GetOptions()
		if len(options) != 1 {
			t.Errorf("expected rule5 to have 1 option, got %d", len(options))
		}
		if options["option2"] != "value2" {
			t.Errorf("expected rule5 option2 to be 'value2', got %v", options["option2"])
		}
	} else {
		t.Error("expected rule5 to exist")
	}
}

func TestInitDefaultConfig(t *testing.T) {
	t.Run("create config in empty directory", func(t *testing.T) {
		tempDir := t.TempDir()

		err := InitDefaultConfig(tempDir)
		if err != nil {
			t.Fatalf("InitDefaultConfig failed: %v", err)
		}

		configPath := filepath.Join(tempDir, "rslint.jsonc")
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Error("rslint.jsonc file was not created")
		}

		content, err := os.ReadFile(configPath)
		if err != nil {
			t.Fatalf("failed to read created config file: %v", err)
		}

		if string(content) != defaultJsonc {
			t.Error("created config file content does not match expected default content")
		}
	})

	t.Run("fail when config already exists", func(t *testing.T) {
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "rslint.jsonc")

		err := os.WriteFile(configPath, []byte("existing content"), 0644)
		if err != nil {
			t.Fatalf("failed to create existing config file: %v", err)
		}

		err = InitDefaultConfig(tempDir)
		if err == nil {
			t.Error("expected InitDefaultConfig to fail when config already exists")
		}

		expectedErrorMsg := "rslint.json already exists in " + tempDir
		if err.Error() != expectedErrorMsg {
			t.Errorf("expected error message %q, got %q", expectedErrorMsg, err.Error())
		}

		content, err := os.ReadFile(configPath)
		if err != nil {
			t.Fatalf("failed to read config file: %v", err)
		}

		if string(content) != "existing content" {
			t.Error("existing config file was modified")
		}
	})

	t.Run("fail with invalid directory", func(t *testing.T) {
		invalidDir := "/nonexistent/directory/path"

		err := InitDefaultConfig(invalidDir)
		if err == nil {
			t.Error("expected InitDefaultConfig to fail with invalid directory")
		}

		expectedPrefix := "failed to create rslint.json:"
		if !strings.HasPrefix(err.Error(), expectedPrefix) {
			t.Errorf("expected error to start with %q, got %q", expectedPrefix, err.Error())
		}
	})

	t.Run("verify file permissions", func(t *testing.T) {
		tempDir := t.TempDir()

		err := InitDefaultConfig(tempDir)
		if err != nil {
			t.Fatalf("InitDefaultConfig failed: %v", err)
		}

		configPath := filepath.Join(tempDir, "rslint.jsonc")
		fileInfo, err := os.Stat(configPath)
		if err != nil {
			t.Fatalf("failed to stat config file: %v", err)
		}

		expectedMode := os.FileMode(0644)
		if fileInfo.Mode().Perm() != expectedMode {
			t.Errorf("expected file mode %v, got %v", expectedMode, fileInfo.Mode().Perm())
		}
	})

	t.Run("create config with relative path", func(t *testing.T) {
		tempDir := t.TempDir()

		originalWD, err := os.Getwd()
		if err != nil {
			t.Fatalf("failed to get current working directory: %v", err)
		}
		defer func() {
			if err := os.Chdir(originalWD); err != nil {
				t.Errorf("failed to restore working directory: %v", err)
			}
		}()

		err = os.Chdir(tempDir)
		if err != nil {
			t.Fatalf("failed to change working directory: %v", err)
		}

		err = InitDefaultConfig(".")
		if err != nil {
			t.Fatalf("InitDefaultConfig with relative path failed: %v", err)
		}

		configPath := "rslint.jsonc"
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Error("rslint.jsonc file was not created with relative path")
		}
	})

	t.Run("create config in nested directory", func(t *testing.T) {
		tempDir := t.TempDir()

		nestedDir := filepath.Join(tempDir, "project", "config")
		err := os.MkdirAll(nestedDir, 0755)
		if err != nil {
			t.Fatalf("failed to create nested directory: %v", err)
		}

		err = InitDefaultConfig(nestedDir)
		if err != nil {
			t.Fatalf("InitDefaultConfig in nested directory failed: %v", err)
		}

		configPath := filepath.Join(nestedDir, "rslint.jsonc")
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Error("rslint.jsonc file was not created in nested directory")
		}
	})

	t.Run("verify config content is valid JSON", func(t *testing.T) {
		tempDir := t.TempDir()

		err := InitDefaultConfig(tempDir)
		if err != nil {
			t.Fatalf("InitDefaultConfig failed: %v", err)
		}

		configPath := filepath.Join(tempDir, "rslint.jsonc")
		content, err := os.ReadFile(configPath)
		if err != nil {
			t.Fatalf("failed to read config file: %v", err)
		}

		if len(content) == 0 {
			t.Error("created config file is empty")
		}

		contentStr := string(content)
		expectedElements := []string{
			"ignores",
			"languageOptions",
			"parserOptions",
			"project",
			"rules",
			"plugins",
			"@typescript-eslint",
		}

		for _, element := range expectedElements {
			if !strings.Contains(contentStr, element) {
				t.Errorf("config content missing expected element: %s", element)
			}
		}
	})
}
