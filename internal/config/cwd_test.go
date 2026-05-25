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
		{
			name:         "Pattern with ./ prefix matches normalized path",
			filePath:     "src/component.ts",
			patterns:     []string{"./src/**"},
			shouldIgnore: true,
			description:  "Patterns with ./ prefix should match paths without ./ prefix",
		},
		{
			name:         "Pattern with ./ prefix matches exact file",
			filePath:     "src/rslint-test-cases.ts",
			patterns:     []string{"./src/rslint-test-cases.ts"},
			shouldIgnore: true,
			description:  "Exact file pattern with ./ prefix should match normalized path",
		},
		{
			name:         "Pattern with .. segment ignores correctly",
			filePath:     "lib/component.ts",
			patterns:     []string{"src/../lib/**"},
			shouldIgnore: true,
			description:  "Patterns with .. segments should resolve and match",
		},
		{
			name:         "Pattern with mid-path /./ ignores correctly",
			filePath:     "src/utils/helper.ts",
			patterns:     []string{"src/./utils/**"},
			shouldIgnore: true,
			description:  "Patterns with mid-path /./ should collapse and match",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isFileIgnored(tt.filePath, tt.patterns, originalCwd)
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

func TestNormalizePattern(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		expected string
	}{
		{name: "No-op for clean pattern", pattern: "src/**/*.ts", expected: "src/**/*.ts"},
		{name: "Strip leading ./", pattern: "./src/**/*.ts", expected: "src/**/*.ts"},
		{name: "Collapse mid-path /./", pattern: "src/./utils/*.ts", expected: "src/utils/*.ts"},
		{name: "Resolve .. segment", pattern: "src/../lib/*.ts", expected: "lib/*.ts"},
		{name: "Combined ./ prefix and .. segment", pattern: "./src/../lib/*.ts", expected: "lib/*.ts"},
		{name: "Multiple /./", pattern: "src/./utils/./deep/*.ts", expected: "src/utils/deep/*.ts"},
		{name: "Exact file with ./", pattern: "./src/file.ts", expected: "src/file.ts"},
		{name: "Plain filename unchanged", pattern: "*.ts", expected: "*.ts"},
		{name: "Double star unchanged", pattern: "**/*.ts", expected: "**/*.ts"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizePattern(tt.pattern)
			if result != tt.expected {
				t.Errorf("normalizePattern(%q) = %q, expected %q",
					tt.pattern, result, tt.expected)
			}
		})
	}
}

func TestIsFileMatchedDotSlashPrefix(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Unable to get working directory: %v", err)
	}

	tests := []struct {
		name     string
		filePath string
		patterns []string
		expected bool
	}{
		{
			name:     "Pattern with ./ prefix matches relative path",
			filePath: "src/rslint-test-cases.ts",
			patterns: []string{"./src/rslint-test-cases.ts"},
			expected: true,
		},
		{
			name:     "Pattern with ./ prefix glob matches relative path",
			filePath: "src/component.ts",
			patterns: []string{"./src/**"},
			expected: true,
		},
		{
			name:     "Pattern without ./ still works",
			filePath: "src/component.ts",
			patterns: []string{"src/**"},
			expected: true,
		},
		{
			name:     "Non-matching pattern with ./ prefix",
			filePath: "lib/component.ts",
			patterns: []string{"./src/**"},
			expected: false,
		},
		{
			name:     "Pattern with mid-path /./",
			filePath: "src/utils/helper.ts",
			patterns: []string{"src/./utils/*.ts"},
			expected: true,
		},
		{
			name:     "Pattern with .. segment",
			filePath: "lib/component.ts",
			patterns: []string{"src/../lib/*.ts"},
			expected: true,
		},
		{
			name:     "Pattern with .. segment non-matching",
			filePath: "src/component.ts",
			patterns: []string{"src/../lib/*.ts"},
			expected: false,
		},
		{
			name:     "Pattern with combined ./ prefix and .. segment",
			filePath: "lib/component.ts",
			patterns: []string{"./src/../lib/*.ts"},
			expected: true,
		},
		{
			name:     "Pattern with ./ prefix and ** glob",
			filePath: "src/deep/nested/file.ts",
			patterns: []string{"./src/**/*.ts"},
			expected: true,
		},
		{
			name:     "Multiple patterns with first matching",
			filePath: "src/component.ts",
			patterns: []string{"./lib/**", "./src/**"},
			expected: true,
		},
		{
			name:     "Multiple patterns with none matching",
			filePath: "test/component.ts",
			patterns: []string{"./lib/**", "./src/**"},
			expected: false,
		},
		{
			name:     "Empty patterns list",
			filePath: "src/component.ts",
			patterns: []string{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isFileMatched(tt.filePath, tt.patterns, cwd)
			if result != tt.expected {
				t.Errorf("isFileMatched(%q, %v, %q) = %v, expected %v",
					tt.filePath, tt.patterns, cwd, result, tt.expected)
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
			matched, err := doublestar.Match(tt.pattern, tt.path)
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
