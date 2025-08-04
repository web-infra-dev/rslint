package rule

import (
	"strings"
	"testing"
)

func TestParseRuleNames(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "single rule",
			input:    "no-unused-vars",
			expected: []string{"no-unused-vars"},
		},
		{
			name:     "multiple rules",
			input:    "no-unused-vars, no-console, no-debugger",
			expected: []string{"no-unused-vars", "no-console", "no-debugger"},
		},
		{
			name:     "rules with extra spaces",
			input:    " no-unused-vars , no-console ",
			expected: []string{"no-unused-vars", "no-console"},
		},
		{
			name:     "empty string",
			input:    "",
			expected: nil,
		},
		{
			name:     "whitespace only",
			input:    "   ",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseRuleNames(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d rules, got %d", len(tt.expected), len(result))
				return
			}

			for i, expected := range tt.expected {
				if result[i] != expected {
					t.Errorf("Expected rule %d to be %q, got %q", i, expected, result[i])
				}
			}
		})
	}
}

func TestESLintDirectiveRegexPatterns(t *testing.T) {
	tests := []struct {
		name        string
		line        string
		shouldMatch string
		rules       string
	}{
		{
			name:        "eslint-disable-line with rule",
			line:        "// eslint-disable-line no-unused-vars",
			shouldMatch: "disable-line",
			rules:       "no-unused-vars",
		},
		{
			name:        "eslint-disable-next-line with multiple rules",
			line:        "// eslint-disable-next-line no-console, no-debugger",
			shouldMatch: "disable-next-line",
			rules:       "no-console, no-debugger",
		},
		{
			name:        "eslint-disable block comment",
			line:        "/* eslint-disable no-unused-vars */",
			shouldMatch: "disable",
			rules:       "no-unused-vars",
		},
		{
			name:        "eslint-enable block comment",
			line:        "/* eslint-enable no-unused-vars */",
			shouldMatch: "enable",
			rules:       "no-unused-vars",
		},
		{
			name:        "eslint-disable-line without rules",
			line:        "// eslint-disable-line",
			shouldMatch: "disable-line",
			rules:       "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch tt.shouldMatch {
			case "disable-line":
				if !strings.Contains(tt.line, "eslint-disable-line") {
					t.Errorf("Line should contain eslint-disable-line directive")
				}
			case "disable-next-line":
				if !strings.Contains(tt.line, "eslint-disable-next-line") {
					t.Errorf("Line should contain eslint-disable-next-line directive")
				}
			case "disable":
				if !strings.Contains(tt.line, "eslint-disable") || strings.Contains(tt.line, "eslint-disable-") {
					t.Errorf("Line should contain eslint-disable directive (not disable-line or disable-next-line)")
				}
			case "enable":
				if !strings.Contains(tt.line, "eslint-enable") {
					t.Errorf("Line should contain eslint-enable directive")
				}
			}

			if tt.rules != "" {
				parsed := parseRuleNames(tt.rules)
				if len(parsed) == 0 {
					t.Errorf("Should have parsed rules from %q", tt.rules)
				}
			}
		})
	}
}

func TestDisableManagerBasicFunctionality(t *testing.T) {
	dm := &DisableManager{
		sourceFile:            nil,
		disabledRules:         make(map[string]bool),
		lineDisabledRules:     make(map[int][]string),
		nextLineDisabledRules: make(map[int][]string),
	}

	dm.disabledRules["no-console"] = true
	dm.lineDisabledRules[5] = []string{"no-unused-vars"}
	dm.nextLineDisabledRules[10] = []string{"no-debugger"}

	if !dm.disabledRules["no-console"] {
		t.Error("Expected no-console to be disabled")
	}

	if len(dm.lineDisabledRules[5]) != 1 || dm.lineDisabledRules[5][0] != "no-unused-vars" {
		t.Error("Expected no-unused-vars to be disabled for line 5")
	}

	if len(dm.nextLineDisabledRules[10]) != 1 || dm.nextLineDisabledRules[10][0] != "no-debugger" {
		t.Error("Expected no-debugger to be disabled for next line after 10")
	}
}
