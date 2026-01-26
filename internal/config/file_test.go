package config

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/bmatcuk/doublestar/v4"
)

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

func TestNormalizePath(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		cwd      string
		expected string
	}{
		{
			name:     "Relative path converts to absolute path",
			filePath: "src/main.ts",
			cwd:      "/Users/labmda47/Code/test",
			expected: "/Users/labmda47/Code/test/src/main.ts",
		},
		{
			name:     "Absolute path remains unchanged",
			filePath: "/Users/labmda47/Code/test/src/main.ts",
			cwd:      "/Users/labmda47/Code/test",
			expected: "/Users/labmda47/Code/test/src/main.ts",
		},
		{
			name:     "Complex relative path",
			filePath: "./src/../src/main.ts",
			cwd:      "/Users/labmda47/Code/test",
			expected: "/Users/labmda47/Code/test/src/main.ts",
		},
		{
			name:     "Windows path normalized",
			filePath: "src\\main.ts",
			cwd:      "D:\\Code",
			expected: "D:/Code/src/main.ts",
		},
		{
			name:     "Windows relative path converts to absolute path",
			filePath: ".\\src\\..\\src\\main.ts",
			cwd:      "D:\\Code",
			expected: "D:/Code/src/main.ts",
		},
		{
			name:     "Empty cwd returns normalized input path",
			filePath: "src/main.ts",
			cwd:      "",
			expected: "src/main.ts",
		},
		{
			name:     "Empty cwd returns normalized input path",
			filePath: "src\\main.ts",
			cwd:      "",
			expected: "src\\main.ts",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeAbsolutePath(tt.filePath, tt.cwd)
			if result != tt.expected {
				t.Errorf("normalizePath(%q, %q) = %q, expected %q",
					tt.filePath, tt.cwd, result, tt.expected)
			}
		})
	}
}

func TestParseNegationPattern(t *testing.T) {
	tests := []struct {
		name          string
		pattern       string
		isNeg         bool
		actualPattern string
	}{
		{
			name:          "Negation pattern",
			pattern:       "!src/components/**/*.ts",
			isNeg:         true,
			actualPattern: "src/components/**/*.ts"},
		{
			name:          "Positive pattern",
			pattern:       "src/components/**/*.ts",
			isNeg:         false,
			actualPattern: "src/components/**/*.ts"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isNeg, actualPattern := parseNegationPattern(tt.pattern)
			if isNeg != tt.isNeg || actualPattern != tt.actualPattern {
				t.Errorf("parseNegationPattern(%q) = (%v, %q), expected (%v, %q)",
					tt.pattern, isNeg, actualPattern, tt.isNeg, tt.actualPattern)
			}
		})
	}
}

func TestClassifyPatterns(t *testing.T) {
	tests := []struct {
		name             string
		patterns         []string
		positivePatterns []string
		negativePatterns []string
	}{
		{
			name: "Positive and negative patterns",
			patterns: []string{
				"src/components/**/*.ts",
				"!src/components/**/*.test.ts",
			},
			positivePatterns: []string{
				"src/components/**/*.ts",
			},
			negativePatterns: []string{
				"src/components/**/*.test.ts",
			},
		},
		{
			name: "Only positive patterns",
			patterns: []string{
				"src/components/**/*.ts",
				"src/components/**/*.test.ts",
			},
			positivePatterns: []string{
				"src/components/**/*.ts",
				"src/components/**/*.test.ts",
			},
			negativePatterns: []string{},
		},
		{
			name: "Only negative patterns",
			patterns: []string{
				"!src/components/**/*.ts",
				"!src/components/**/*.test.ts",
			},
			positivePatterns: []string{},
			negativePatterns: []string{
				"src/components/**/*.ts",
				"src/components/**/*.test.ts",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			positivePatterns, negativePatterns := classifyPatterns(tt.patterns)
			if !slices.Equal(positivePatterns, tt.positivePatterns) || !slices.Equal(negativePatterns, tt.negativePatterns) {
				t.Errorf("classifyPatterns(%#v) = (%#v, %#v), expected (%#v, %#v)",
					tt.patterns, positivePatterns, negativePatterns, tt.positivePatterns, tt.negativePatterns)
			}
		})
	}
}

func TestNormalizePatterns(t *testing.T) {
	tests := []struct {
		name     string
		patterns []string
		cwd      string
		expected []string
	}{
		{
			name: "Normalize patterns with relative path",
			patterns: []string{
				"src/components/**/*.ts",
				"src/components/**/*.test.ts",
			},
			cwd: "/Users/labmda47/Code/test",
			expected: []string{
				"/Users/labmda47/Code/test/src/components/**/*.ts",
				"/Users/labmda47/Code/test/src/components/**/*.test.ts",
			},
		},
		{
			name: "Normalize patterns with relative path and absolute path",
			patterns: []string{
				"src/components/**/*.ts",
				"/Users/labmda47/Code/test/src/components/**/*.test.ts",
			},
			cwd: "/Users/labmda47/Code/test",
			expected: []string{
				"/Users/labmda47/Code/test/src/components/**/*.ts",
				"/Users/labmda47/Code/test/src/components/**/*.test.ts",
			},
		},
		{
			name: "Normalize patterns with windows cwd path",
			patterns: []string{
				"src/components/**/*.ts",
				"src/components/**/*.test.ts",
			},
			cwd: "D:\\Code\\test",
			expected: []string{
				"D:/Code/test/src/components/**/*.ts",
				"D:/Code/test/src/components/**/*.test.ts",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			normalizedPatterns := normalizePatterns(tt.patterns, tt.cwd)
			if !slices.Equal(normalizedPatterns, tt.expected) {
				t.Errorf("normalizePatterns(%#v, %q) = %#v, expected %#v",
					tt.patterns, tt.cwd, normalizedPatterns, tt.expected)
			}
		})
	}
}

func TestNormalizedFilePatterns_isFileMatched(t *testing.T) {
	tests := []struct {
		name     string
		patterns []string
		cwd      string
		filePath string
		expected bool
	}{
		{
			name: "File matched by positive pattern",
			patterns: []string{
				"src/components/**/*.ts",
			},
			cwd:      "/Users/labmda47/Code/test",
			filePath: "/Users/labmda47/Code/test/src/components/button.ts",
			expected: true,
		},
		{
			name: "File excluded by negative pattern",
			patterns: []string{
				"src/components/**/*.ts",
				"!src/components/**/*.test.ts",
			},
			cwd:      "/Users/labmda47/Code/test",
			filePath: "/Users/labmda47/Code/test/src/components/button.test.ts",
			expected: false,
		},
		{
			name: "File matched by positive pattern, then excluded by negative pattern",
			patterns: []string{
				"!src/components/**/*.test.ts",
				"src/components/**/*.ts",
			},
			cwd:      "/Users/labmda47/Code/test",
			filePath: "/Users/labmda47/Code/test/src/components/button.test.ts",
			expected: false,
		},
		{
			name: "File not matched by any pattern",
			patterns: []string{
				"src/components/**/*.ts",
			},
			cwd:      "/Users/labmda47/Code/test",
			filePath: "/Users/labmda47/Code/test/src/utils/helper.ts",
			expected: false,
		},
		{
			name:     "No patterns means all files match",
			patterns: []string{},
			cwd:      "/Users/labmda47/Code/test",
			filePath: "/Users/labmda47/Code/test/src/components/button.ts",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			normalizedFilePatterns := newNormalizedFilePatterns(tt.patterns, tt.cwd)
			result := normalizedFilePatterns.isFileMatched(tt.filePath)
			if result != tt.expected {
				t.Errorf("patterns: %#v, isFileMatched(%q) = %v, expected %v",
					tt.patterns, tt.filePath, result, tt.expected)
			}
		})
	}
}

func TestNormalizedIgnorePatterns_isFileIgnored(t *testing.T) {
	tests := []struct {
		name     string
		patterns []string
		cwd      string
		filePath string
		expected bool
	}{
		{
			name: "File ignored by positive pattern",
			patterns: []string{
				"src/components/**/*.ts",
			},
			cwd:      "/Users/labmda47/Code/test",
			filePath: "/Users/labmda47/Code/test/src/components/button.ts",
			expected: true,
		},
		{
			name: "File not ignored by positive pattern",
			patterns: []string{
				"src/components/**/*.ts",
			},
			cwd:      "/Users/labmda47/Code/test",
			filePath: "/Users/labmda47/Code/test/src/utils/helper.ts",
			expected: false,
		},
		{
			name: "File not ignored by negation pattern",
			patterns: []string{
				"!src/components/**/*.ts",
			},
			cwd:      "/Users/labmda47/Code/test",
			filePath: "/Users/labmda47/Code/test/src/components/button.ts",
			expected: false,
		},
		{
			name: "File not matching negation pattern remains not ignored",
			patterns: []string{
				"!src/components/**/*.ts",
			},
			cwd:      "/Users/labmda47/Code/test",
			filePath: "/Users/labmda47/Code/test/src/utils/helper.ts",
			expected: false,
		},
		{
			name: "File ignored by positive pattern, then not ignored by negative pattern",
			patterns: []string{
				"src/components/**/*.ts",
				"!src/components/**/*.test.ts",
			},
			cwd:      "/Users/labmda47/Code/test",
			filePath: "/Users/labmda47/Code/test/src/components/button.test.ts",
			expected: false,
		},
		{
			name: "File not ignored by positive pattern, then ignored by negative pattern",
			patterns: []string{
				"!src/components/**/*.test.ts",
				"src/components/**/button.test.ts",
			},
			cwd:      "/Users/labmda47/Code/test",
			filePath: "/Users/labmda47/Code/test/src/components/button.test.ts",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			normalizedIgnorePatterns := newNormalizedIgnorePatterns(tt.patterns, tt.cwd)
			result := normalizedIgnorePatterns.isFileIgnored(tt.filePath)
			if result != tt.expected {
				t.Errorf("patterns: %#v, isFileIgnored(%q) = %v, expected %v",
					tt.patterns, tt.filePath, result, tt.expected)
			}
		})
	}
}

func TestFileMatcher_isFileMatched(t *testing.T) {
	// Save the original working directory
	originalCwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Unable to get current working directory: %v", err)
	}
	defer t.Chdir(originalCwd) // Restore after test completes

	tests := []struct {
		name        string
		description string
		config      ConfigEntry
		filePath    string
		expected    bool
	}{
		{
			name:        "No patterns means all files match",
			description: "No patterns means all files match",
			config: ConfigEntry{
				Files:   []string{},
				Ignores: []string{},
			},
			filePath: "src/component.ts",
			expected: true,
		},
		{
			name:        "Relative path matching",
			description: "Relative paths should match relative patterns",
			config: ConfigEntry{
				Files: []string{
					"src/**",
				},
			},
			filePath: "src/component.ts",
			expected: true,
		},
		{
			name:        "Absolute path matches relative pattern",
			description: "File with absolute path should be normalized and match relative glob pattern",
			filePath:    filepath.Join(originalCwd, "src/component.ts"),
			config: ConfigEntry{
				Files: []string{
					"src/**",
				},
			},
			expected: true,
		},
		{
			name:        "Deep nested file matches recursive pattern",
			description: "File in deeply nested directory should match recursive glob pattern",
			filePath:    "node_modules/package/deep/file.js",
			config: ConfigEntry{
				Files: []string{
					"node_modules/**",
				},
			},
			expected: true,
		},
		{
			name:        "Test file matches global recursive pattern",
			description: "Test file in any location should match global recursive pattern",
			filePath:    "src/utils/helper.test.ts",
			config: ConfigEntry{
				Files: []string{
					"**/*.test.ts",
				},
			},
			expected: true,
		},
		{
			name:        "File not matching any pattern",
			description: "File that doesn't match any pattern should be ignored",
			filePath:    "src/component.ts",
			config: ConfigEntry{
				Files: []string{
					"dist/**",
					"*.log",
				},
			},
			expected: false,
		},
		{
			name:        "Windows path separator handled correctly",
			description: "File path with Windows backslash separators should be normalized and matched",
			filePath:    "src\\windows\\style\\path.ts",
			config: ConfigEntry{
				Files: []string{
					"src/**",
				},
			},
			expected: true,
		},
		{
			name:        "File excluded by single negation pattern",
			description: "File matching a negation pattern should be excluded even if it matches the base pattern",
			filePath:    "src/component.ts",
			config: ConfigEntry{
				Files: []string{
					"!src/component.ts",
				},
			},
			expected: false,
		},
		{
			name:        "No patterns means all files match",
			description: "No patterns means all files match",
			filePath:    "src/component.ts",
			config: ConfigEntry{
				Files: []string{},
			},
			expected: true,
		},
		{
			name:        "File ignored by ignore pattern",
			description: "File matching ignore pattern should be ignored",
			filePath:    "src/component.ts",
			config: ConfigEntry{
				Files: []string{
					"src/**",
				},
				Ignores: []string{
					"src/component.ts",
				},
			},
			expected: false,
		},
		{
			name:        "File ignored by ignore pattern, then not ignored by negation pattern",
			description: "File matching ignore pattern should be ignored, then not ignored by negation pattern",
			filePath:    "src/utils/helper.test.ts",
			config: ConfigEntry{
				Files: []string{
					"src/**",
				},
				Ignores: []string{
					"src/**/*.test.ts",
					"!src/**/helper.test.ts",
				},
			},
			expected: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fileMatcher := newFileMatcher(&tt.config, originalCwd)
			result := fileMatcher.isFileMatched(normalizeAbsolutePath(tt.filePath, originalCwd))
			if result != tt.expected {
				t.Errorf("%s, config: %#v, isFileMatched(%q) = %v, expected %v",
					tt.description, tt.config, tt.filePath, result, tt.expected)
			}
		})
	}
}
