package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bmatcuk/doublestar/v4"
)

func TestCwdHandling(t *testing.T) {
	// Save the original working directory
	originalCwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Unable to get current working directory: %v", err)
	}
	defer t.Chdir(originalCwd) // Restore after test completes

	tests := []struct {
		name         string
		filePath     string
		patterns     []string
		shouldIgnore bool
		description  string
	}{
		{
			name:         "Relative path matching",
			filePath:     "src/component.ts",
			patterns:     []string{"src/**"},
			shouldIgnore: true,
			description:  "Relative paths should match relative patterns",
		},
		{
			name:         "Absolute path to relative path matching",
			filePath:     filepath.Join(originalCwd, "src/component.ts"),
			patterns:     []string{"src/**"},
			shouldIgnore: true,
			description:  "Absolute paths should be converted to relative paths before matching",
		},
		{
			name:         "node_modules recursive matching",
			filePath:     "node_modules/package/deep/file.js",
			patterns:     []string{"node_modules/**"},
			shouldIgnore: true,
			description:  "Recursive patterns should match deeply nested files",
		},
		{
			name:         "Test file pattern matching",
			filePath:     "src/utils/helper.test.ts",
			patterns:     []string{"**/*.test.ts"},
			shouldIgnore: true,
			description:  "Global recursive patterns should match test files in any location",
		},
		{
			name:         "Non-matching file",
			filePath:     "src/component.ts",
			patterns:     []string{"dist/**", "*.log"},
			shouldIgnore: false,
			description:  "Files not matching any pattern should not be ignored",
		},
		{
			name:         "Cross-platform path handling",
			filePath:     "src\\windows\\style\\path.ts",
			patterns:     []string{"src/**"},
			shouldIgnore: true,
			description:  "Windows style paths should be handled correctly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isFileIgnored(tt.filePath, tt.patterns)
			if result != tt.shouldIgnore {
				t.Errorf("%s: isFileIgnored(%q, %v) = %v, expected %v",
					tt.description, tt.filePath, tt.patterns, result, tt.shouldIgnore)
			}
		})
	}
}

func TestNormalizePath(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Unable to get working directory: %v", err)
	}

	tests := []struct {
		name     string
		filePath string
		expected string
	}{
		{
			name:     "Relative path remains unchanged",
			filePath: "src/component.ts",
			expected: "src/component.ts",
		},
		{
			name:     "Absolute path converts to relative path",
			filePath: filepath.Join(cwd, "src/component.ts"),
			expected: "src/component.ts",
		},
		{
			name:     "Current directory marker",
			filePath: "./src/component.ts",
			expected: "src/component.ts",
		},
		{
			name:     "Complex relative path",
			filePath: "src/../src/component.ts",
			expected: "src/component.ts",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizePath(tt.filePath, cwd)
			if result != tt.expected {
				t.Errorf("normalizePath(%q, %q) = %q, expected %q",
					tt.filePath, cwd, result, tt.expected)
			}
		})
	}
}

func TestDoublestarBehavior(t *testing.T) {
	// Test specific behavior of the doublestar library
	tests := []struct {
		pattern  string
		path     string
		expected bool
		name     string
	}{
		{"**/*.ts", "src/file.ts", true, "Recursive matching of TypeScript files"},
		{"**/*.ts", "src/deep/nested/file.ts", true, "Deep recursive matching"},
		{"src/**", "src/file.ts", true, "Directory recursive matching"},
		{"src/**", "src/deep/nested/file.ts", true, "Deep directory recursive matching"},
		{"*.ts", "file.ts", true, "Single-level wildcard"},
		{"*.ts", "src/file.ts", false, "Single-level wildcard doesn't match nested files"},
		{"node_modules/**", "node_modules/package/file.js", true, "node_modules recursive matching"},
		{"**/test/**", "src/test/file.js", true, "Middle recursive matching"},
		{"**/test/**", "test/file.js", true, "Beginning recursive matching"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched, err := doublestar.PathMatch(tt.pattern, tt.path)
			if err != nil {
				t.Errorf("doublestar.PathMatch error: %v", err)
				return
			}
			if matched != tt.expected {
				t.Errorf("doublestar.PathMatch(%q, %q) = %v, expected %v",
					tt.pattern, tt.path, matched, tt.expected)
			}
		})
	}
}
