package config

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ParseCLIRuleFlag parses a single --rule flag value in ESLint-compatible format.
// Supported formats:
//   - "ruleName: error"
//   - "ruleName: warn"
//   - "ruleName: off"
//   - "ruleName: [error, {\"allow\": [\"warn\"]}]"
//
// Returns the rule name and its parsed configuration.
func ParseCLIRuleFlag(input string) (string, interface{}, error) {
	// Split on first ":" to separate rule name from value, then trim spaces.
	// Rule names may contain "/" (e.g. "@typescript-eslint/no-explicit-any")
	// but never ":", so the first ":" is always the separator.
	idx := strings.Index(input, ":")
	if idx < 0 {
		return "", nil, fmt.Errorf("invalid --rule format %q: expected \"ruleName: severity\"", input)
	}

	name := strings.TrimSpace(input[:idx])
	value := strings.TrimSpace(input[idx+1:])

	if name == "" {
		return "", nil, fmt.Errorf("invalid --rule format %q: empty rule name", input)
	}
	if value == "" {
		return "", nil, fmt.Errorf("invalid --rule format %q: empty value", input)
	}

	// If value starts with "[", parse as JSON array → array-style rule config.
	if strings.HasPrefix(value, "[") {
		var arr []interface{}
		if err := json.Unmarshal([]byte(value), &arr); err != nil {
			return "", nil, fmt.Errorf("invalid --rule format %q: %w", input, err)
		}
		return name, arr, nil
	}

	// Otherwise treat as a plain severity string ("error", "warn", "off").
	return name, value, nil
}

// BuildCLIRuleEntry converts a list of --rule flag values into a synthetic
// ConfigEntry with no Files field (matches all files). Returns nil if flags is empty.
func BuildCLIRuleEntry(flags []string) (*ConfigEntry, error) {
	if len(flags) == 0 {
		return nil, nil //nolint:nilnil // no flags means no entry, not an error
	}

	rules := make(Rules, len(flags))
	for _, f := range flags {
		name, value, err := ParseCLIRuleFlag(f)
		if err != nil {
			return nil, err
		}
		rules[name] = value
	}

	return &ConfigEntry{Rules: rules}, nil
}
