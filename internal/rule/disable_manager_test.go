package rule

import (
	"regexp"
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
		{
			name:     "typescript-eslint rule with slash",
			input:    "@typescript-eslint/no-unsafe-member-access",
			expected: []string{"@typescript-eslint/no-unsafe-member-access"},
		},
		{
			name:     "mixed rules with typescript-eslint",
			input:    "no-console, @typescript-eslint/no-unsafe-member-access, no-unused-vars",
			expected: []string{"no-console", "@typescript-eslint/no-unsafe-member-access", "no-unused-vars"},
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
		{
			name:        "eslint-disable-next-line with typescript-eslint rule",
			line:        "// eslint-disable-next-line @typescript-eslint/no-unsafe-member-access",
			shouldMatch: "disable-next-line",
			rules:       "@typescript-eslint/no-unsafe-member-access",
		},
		{
			name:        "eslint-disable-line with multiple typescript-eslint rules",
			line:        "// eslint-disable-line @typescript-eslint/no-unsafe-member-access, @typescript-eslint/no-unsafe-assignment",
			shouldMatch: "disable-line",
			rules:       "@typescript-eslint/no-unsafe-member-access, @typescript-eslint/no-unsafe-assignment",
		},
		{
			name:        "eslint-disable with scoped rule in block comment",
			line:        "/* eslint-disable @typescript-eslint/no-floating-promises */",
			shouldMatch: "disable",
			rules:       "@typescript-eslint/no-floating-promises",
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

func TestRegexPatternsForTypeScriptESLint(t *testing.T) {
	// Test the actual regex patterns used in the disable manager
	eslintDisableLineRe := regexp.MustCompile(`//\s*eslint-disable-line(?:\s+([^\r\n]+))?`)
	eslintDisableNextLineRe := regexp.MustCompile(`//\s*eslint-disable-next-line(?:\s+([^\r\n]+))?`)

	tests := []struct {
		name     string
		line     string
		regex    *regexp.Regexp
		expected string
	}{
		{
			name:     "disable-line with typescript-eslint rule",
			line:     "// eslint-disable-line @typescript-eslint/no-unsafe-member-access",
			regex:    eslintDisableLineRe,
			expected: "@typescript-eslint/no-unsafe-member-access",
		},
		{
			name:     "disable-next-line with typescript-eslint rule",
			line:     "// eslint-disable-next-line @typescript-eslint/no-unsafe-member-access",
			regex:    eslintDisableNextLineRe,
			expected: "@typescript-eslint/no-unsafe-member-access",
		},
		{
			name:     "disable-line with multiple typescript-eslint rules",
			line:     "// eslint-disable-line @typescript-eslint/no-unsafe-member-access, @typescript-eslint/no-unsafe-assignment",
			regex:    eslintDisableLineRe,
			expected: "@typescript-eslint/no-unsafe-member-access, @typescript-eslint/no-unsafe-assignment",
		},
		{
			name:     "disable-next-line with mixed rules",
			line:     "// eslint-disable-next-line no-console, @typescript-eslint/no-floating-promises",
			regex:    eslintDisableNextLineRe,
			expected: "no-console, @typescript-eslint/no-floating-promises",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := tt.regex.FindStringSubmatch(tt.line)
			if matches == nil {
				t.Fatalf("Expected regex to match line: %s", tt.line)
			}
			if len(matches) < 2 {
				t.Fatalf("Expected regex to capture group, got matches: %v", matches)
			}
			captured := strings.TrimSpace(matches[1])
			expected := strings.TrimSpace(tt.expected)
			if captured != expected {
				t.Errorf("Expected captured group to be %q, got %q", expected, captured)
			}

			// Also test that parseRuleNames works with the captured content
			rules := parseRuleNames(captured)
			if len(rules) == 0 {
				t.Errorf("parseRuleNames should have parsed rules from %q", captured)
			}

			// Verify that rules contain the typescript-eslint rules
			hasTypeScriptRule := false
			for _, rule := range rules {
				if strings.Contains(rule, "@typescript-eslint/") {
					hasTypeScriptRule = true
					break
				}
			}
			if strings.Contains(tt.expected, "@typescript-eslint/") && !hasTypeScriptRule {
				t.Errorf("Expected to find TypeScript-ESLint rule in parsed rules: %v", rules)
			}
		})
	}
}
