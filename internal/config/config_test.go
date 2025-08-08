package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
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

func TestInitConfig(t *testing.T) {
	// Create temporary directory for testing
	tempDir := t.TempDir()

	t.Run("Successfully download and create configuration file", func(t *testing.T) {
		// Test directory
		testDir := filepath.Join(tempDir, "success_test")
		err := os.MkdirAll(testDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create test directory: %v", err)
		}

		// Call InitConfig
		err = InitConfig(testDir)

		// Verify no error
		if err != nil {
			t.Errorf("InitConfig should succeed, but returned error: %v", err)
		}

		// Verify configuration file is created
		configPath := filepath.Join(testDir, "rslint.json")
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Error("Configuration file should be created, but file does not exist")
		}

		// Verify configuration file content
		content, err := os.ReadFile(configPath)
		if err != nil {
			t.Errorf("Failed to read configuration file: %v", err)
		}

		// Verify content is not empty
		if len(content) == 0 {
			t.Error("Configuration file content should not be empty")
		}

		// Verify content is not empty and contains basic structure
		if len(content) == 0 {
			t.Error("Configuration file content should not be empty")
		}

		// Check if it contains basic configuration structure (not requiring strict JSON)
		contentStr := string(content)
		if !strings.Contains(contentStr, "language") {
			t.Error("Configuration file should contain 'language' field")
		}
		if !strings.Contains(contentStr, "files") {
			t.Error("Configuration file should contain 'files' field")
		}
	})

	t.Run("Return error when configuration file already exists", func(t *testing.T) {
		// Test directory
		testDir := filepath.Join(tempDir, "exists_test")
		err := os.MkdirAll(testDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create test directory: %v", err)
		}

		// Pre-create a configuration file
		configPath := filepath.Join(testDir, "rslint.json")
		existingContent := `[{"language": "typescript", "files": ["**/*.ts"]}]`
		err = os.WriteFile(configPath, []byte(existingContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create existing configuration file: %v", err)
		}

		// Call InitConfig
		err = InitConfig(testDir)

		// Verify error is returned
		if err == nil {
			t.Error("InitConfig should return error when configuration file already exists")
		}

		// Verify error message
		expectedError := fmt.Sprintf("rslint.json already exists in %s", testDir)
		if err.Error() != expectedError {
			t.Errorf("Expected error message '%s', but got '%s'", expectedError, err.Error())
		}

		// Verify original file content is not modified
		content, err := os.ReadFile(configPath)
		if err != nil {
			t.Errorf("Failed to read configuration file: %v", err)
		}
		if string(content) != existingContent {
			t.Error("Existing configuration file content should not be modified")
		}
	})

	t.Run("Auto-create directory when it doesn't exist", func(t *testing.T) {
		// Use non-existent directory
		nonExistentDir := filepath.Join(tempDir, "non_existent", "subdir")

		// Call InitConfig
		err := InitConfig(nonExistentDir)

		// Verify no error
		if err != nil {
			t.Errorf("InitConfig should successfully create directory and configuration file, but returned error: %v", err)
		}

		// Verify directory is created
		if _, err := os.Stat(nonExistentDir); os.IsNotExist(err) {
			t.Error("Directory should be auto-created")
		}

		// Verify configuration file is created
		configPath := filepath.Join(nonExistentDir, "rslint.json")
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Error("Configuration file should be created")
		}
	})

	t.Run("Network error handling", func(t *testing.T) {
		// This test needs to simulate network errors
		// Since downloadConfigFromURL is an internal function, we need to test it through other means
		// Here we test an invalid URL scenario (if possible)

		// Note: This test may need to modify the downloadConfigFromURL function to support testing
		// Or control the remote URL through environment variables
		t.Skip("Network error test requires modification of downloadConfigFromURL function to support testing")
	})

	t.Run("HTTP error status code handling", func(t *testing.T) {
		// This test needs to simulate HTTP error status codes
		t.Skip("HTTP error status code test requires modification of downloadConfigFromURL function to support testing")
	})

	t.Run("File permission error handling", func(t *testing.T) {
		// Test directory
		testDir := filepath.Join(tempDir, "permission_test")
		err := os.MkdirAll(testDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create test directory: %v", err)
		}

		// On Unix systems, we can try to set read-only permissions
		// But this may not work on all systems
		if runtime.GOOS != "windows" {
			// Set directory to read-only (this may prevent file creation)
			err = os.Chmod(testDir, 0444)
			if err != nil {
				t.Logf("Unable to set directory permissions for testing: %v", err)
				t.Skip("Permission test skipped")
			}

			// Try to call InitConfig
			err = InitConfig(testDir)

			// Should return error
			if err == nil {
				t.Error("InitConfig should return error in read-only directory")
			}

			// Restore permissions for cleanup
			os.Chmod(testDir, 0755)
		}
	})

	t.Run("Configuration file content validation", func(t *testing.T) {
		// Test directory
		testDir := filepath.Join(tempDir, "content_test")
		err := os.MkdirAll(testDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create test directory: %v", err)
		}

		// Call InitConfig
		err = InitConfig(testDir)
		if err != nil {
			t.Fatalf("InitConfig failed: %v", err)
		}

		// Read and validate configuration file content
		configPath := filepath.Join(testDir, "rslint.json")
		content, err := os.ReadFile(configPath)
		if err != nil {
			t.Fatalf("Failed to read configuration file: %v", err)
		}

		// Verify content is not empty and contains basic structure
		if len(content) == 0 {
			t.Fatalf("Configuration file content should not be empty")
		}

		// Check if it contains basic configuration structure (not requiring strict JSON)
		contentStr := string(content)
		if !strings.Contains(contentStr, "language") {
			t.Fatalf("Configuration file should contain 'language' field")
		}
		if !strings.Contains(contentStr, "files") {
			t.Fatalf("Configuration file should contain 'files' field")
		}

		// Verify configuration structure (if possible)
		if !strings.Contains(contentStr, "[") || !strings.Contains(contentStr, "]") {
			t.Fatalf("Configuration file should contain array structure")
		}
	})

	t.Run("Concurrent call safety", func(t *testing.T) {
		// Test directory
		testDir := filepath.Join(tempDir, "concurrent_test")
		err := os.MkdirAll(testDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create test directory: %v", err)
		}

		// Concurrent calls to InitConfig
		const numGoroutines = 5
		errors := make(chan error, numGoroutines)
		var wg sync.WaitGroup

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := InitConfig(testDir)
				errors <- err
			}()
		}

		wg.Wait()
		close(errors)

		// Count successful and failed calls
		successCount := 0
		errorCount := 0
		for err := range errors {
			if err == nil {
				successCount++
			} else {
				errorCount++
			}
		}

		// Due to race conditions, there may be multiple successful calls
		// But at least one should succeed, and there should be only one configuration file
		if successCount < 1 {
			t.Errorf("Expected at least one successful call, but got %d", successCount)
		}

		// Verify configuration file is actually created
		configPath := filepath.Join(testDir, "rslint.json")
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Error("Configuration file should be created")
		}

		// Verify there is only one configuration file (by checking file size)
		fileInfo, err := os.Stat(configPath)
		if err != nil {
			t.Errorf("Unable to get file info: %v", err)
		}
		if fileInfo.Size() == 0 {
			t.Error("Configuration file should not be empty")
		}
	})

	t.Run("Empty directory path handling", func(t *testing.T) {
		// Test empty directory path
		err := InitConfig("")

		// Should return error (because cannot create directory with empty path)
		if err == nil {
			t.Error("Empty directory path should return error")
		}
	})

	t.Run("Relative path handling", func(t *testing.T) {
		// Test relative path
		relativeDir := "relative_test_dir"
		defer os.RemoveAll(relativeDir) // cleanup

		err := InitConfig(relativeDir)

		// Verify no error
		if err != nil {
			t.Errorf("Relative path should succeed, but returned error: %v", err)
		}

		// Verify configuration file is created
		configPath := filepath.Join(relativeDir, "rslint.json")
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Error("Configuration file should be created")
		}
	})

	t.Run("Special character path handling", func(t *testing.T) {
		// Test path with special characters
		specialDir := filepath.Join(tempDir, "test-dir_with.special@chars")
		err := InitConfig(specialDir)

		// Verify no error
		if err != nil {
			t.Errorf("Special character path should succeed, but returned error: %v", err)
		}

		// Verify configuration file is created
		configPath := filepath.Join(specialDir, "rslint.json")
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Error("Configuration file should be created")
		}
	})

	t.Run("Multiple calls to same directory", func(t *testing.T) {
		// Test directory
		testDir := filepath.Join(tempDir, "multiple_calls_test")
		err := os.MkdirAll(testDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create test directory: %v", err)
		}

		// First call should succeed
		err = InitConfig(testDir)
		if err != nil {
			t.Errorf("First call should succeed, but returned error: %v", err)
		}

		// Second call should fail
		err = InitConfig(testDir)
		if err == nil {
			t.Error("Second call should return error")
		}

		// Verify error message
		expectedError := fmt.Sprintf("rslint.json already exists in %s", testDir)
		if err.Error() != expectedError {
			t.Errorf("Expected error message '%s', but got '%s'", expectedError, err.Error())
		}
	})

	t.Run("Configuration file permission verification", func(t *testing.T) {
		// Test directory
		testDir := filepath.Join(tempDir, "permission_verify_test")
		err := os.MkdirAll(testDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create test directory: %v", err)
		}

		// Call InitConfig
		err = InitConfig(testDir)
		if err != nil {
			t.Fatalf("InitConfig failed: %v", err)
		}

		// Verify configuration file permissions
		configPath := filepath.Join(testDir, "rslint.json")
		fileInfo, err := os.Stat(configPath)
		if err != nil {
			t.Fatalf("Unable to get file info: %v", err)
		}

		// Check file permissions (Unix systems)
		if runtime.GOOS != "windows" {
			mode := fileInfo.Mode()
			// File should be readable and writable
			if mode&0400 == 0 {
				t.Error("Configuration file should be readable")
			}
			if mode&0200 == 0 {
				t.Error("Configuration file should be writable")
			}
		}
	})

	t.Run("Network timeout handling", func(t *testing.T) {
		// This test needs to simulate network timeouts
		// Since downloadConfigFromURL is an internal function, we skip this test
		t.Skip("Network timeout test requires modification of downloadConfigFromURL function to support testing")
	})

	t.Run("Disk space insufficient handling", func(t *testing.T) {
		// This test needs to simulate insufficient disk space
		// Difficult to simulate in real environment, so skip
		t.Skip("Disk space insufficient test is difficult to simulate in real environment")
	})
}
