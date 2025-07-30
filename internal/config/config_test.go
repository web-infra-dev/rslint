package config

import (
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
