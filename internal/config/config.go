package config

// RslintConfig represents the top-level configuration array
type RslintConfig []ConfigEntry

// ConfigEntry represents a single configuration entry in the rslint.json array
type ConfigEntry struct {
	Language        string           `json:"language"`
	Files           []string         `json:"files"`
	LanguageOptions *LanguageOptions `json:"languageOptions,omitempty"`
	Rules           Rules            `json:"rules"`
}

// LanguageOptions contains language-specific configuration options
type LanguageOptions struct {
	ParserOptions *ParserOptions `json:"parserOptions,omitempty"`
}

// ParserOptions contains parser-specific configuration
type ParserOptions struct {
	ProjectService bool     `json:"projectService"`
	Project        []string `json:"project,omitempty"`
}

// Rules represents the rules configuration
// This can be extended to include specific rule configurations
type Rules map[string]interface{}

// Alternative: If you want type-safe rule configurations
type TypedRules struct {
	// Example rule configurations - extend as needed
	NoArrayDelete                      *RuleConfig `json:"no-array-delete,omitempty"`
	NoBaseToString                     *RuleConfig `json:"no-base-to-string,omitempty"`
	NoForInArray                       *RuleConfig `json:"no-for-in-array,omitempty"`
	NoImpliedEval                      *RuleConfig `json:"no-implied-eval,omitempty"`
	OnlyThrowError                     *RuleConfig `json:"only-throw-error,omitempty"`
	AwaitThenable                      *RuleConfig `json:"await-thenable,omitempty"`
	NoConfusingVoidExpression          *RuleConfig `json:"no-confusing-void-expression,omitempty"`
	NoDuplicateTypeConstituents        *RuleConfig `json:"no-duplicate-type-constituents,omitempty"`
	NoFloatingPromises                 *RuleConfig `json:"no-floating-promises,omitempty"`
	NoMeaninglessVoidOperator          *RuleConfig `json:"no-meaningless-void-operator,omitempty"`
	NoMisusedPromises                  *RuleConfig `json:"no-misused-promises,omitempty"`
	NoMisusedSpread                    *RuleConfig `json:"no-misused-spread,omitempty"`
	NoMixedEnums                       *RuleConfig `json:"no-mixed-enums,omitempty"`
	NoRedundantTypeConstituents        *RuleConfig `json:"no-redundant-type-constituents,omitempty"`
	NoUnnecessaryBooleanLiteralCompare *RuleConfig `json:"no-unnecessary-boolean-literal-compare,omitempty"`
	NoUnnecessaryTemplateExpression    *RuleConfig `json:"no-unnecessary-template-expression,omitempty"`
	NoUnnecessaryTypeArguments         *RuleConfig `json:"no-unnecessary-type-arguments,omitempty"`
	NoUnnecessaryTypeAssertion         *RuleConfig `json:"no-unnecessary-type-assertion,omitempty"`
	NoUnsafeArgument                   *RuleConfig `json:"no-unsafe-argument,omitempty"`
	NoUnsafeAssignment                 *RuleConfig `json:"no-unsafe-assignment,omitempty"`
	NoUnsafeCall                       *RuleConfig `json:"no-unsafe-call,omitempty"`
	NoUnsafeEnumComparison             *RuleConfig `json:"no-unsafe-enum-comparison,omitempty"`
	NoUnsafeMemberAccess               *RuleConfig `json:"no-unsafe-member-access,omitempty"`
	NoUnsafeReturn                     *RuleConfig `json:"no-unsafe-return,omitempty"`
	NoUnsafeTypeAssertion              *RuleConfig `json:"no-unsafe-type-assertion,omitempty"`
	NoUnsafeUnaryMinus                 *RuleConfig `json:"no-unsafe-unary-minus,omitempty"`
}

// RuleConfig represents individual rule configuration
type RuleConfig struct {
	Level   string                 `json:"level,omitempty"`   // "error", "warn", "off"
	Options map[string]interface{} `json:"options,omitempty"` // Rule-specific options
}

// IsEnabled returns true if the rule is enabled (not "off")
func (rc *RuleConfig) IsEnabled() bool {
	if rc == nil {
		return false
	}
	return rc.Level != "off" && rc.Level != ""
}

// GetLevel returns the rule level, defaulting to "error" if not specified
func (rc *RuleConfig) GetLevel() string {
	if rc == nil || rc.Level == "" {
		return "error"
	}
	return rc.Level
}

// GetRulesForFile returns enabled rules for a given file based on the configuration
func (config RslintConfig) GetRulesForFile(filePath string) map[string]*RuleConfig {
	enabledRules := make(map[string]*RuleConfig)

	for _, entry := range config {
		// Check if the file matches the files pattern
		matches := false
		if len(entry.Files) == 0 {
			// If no files pattern specified, match all files
			matches = true
		} else {
			for _, pattern := range entry.Files {
				// Simple pattern matching - for now just match all TypeScript files
				if pattern == "**/*.ts" || pattern == "**/*.tsx" {
					matches = true
					break
				}
				if pattern == "*" || pattern == "**/*" {
					matches = true
					break
				}
				// Add more sophisticated pattern matching here if needed
			}
		}

		if matches {
			// Merge rules from this entry
			for ruleName, ruleValue := range entry.Rules {
				switch v := ruleValue.(type) {
				case string:
					// Handle simple string values like "error", "warn", "off"
					if v != "off" {
						enabledRules[ruleName] = &RuleConfig{Level: v}
					}
				case map[string]interface{}:
					// Handle object configuration
					ruleConfig := &RuleConfig{}
					if level, ok := v["level"].(string); ok {
						ruleConfig.Level = level
					}
					if options, ok := v["options"].(map[string]interface{}); ok {
						ruleConfig.Options = options
					}
					if ruleConfig.IsEnabled() {
						enabledRules[ruleName] = ruleConfig
					}
				case []interface{}:
					// Handle array format like ["error", {...options}]
					if len(v) > 0 {
						if level, ok := v[0].(string); ok && level != "off" {
							ruleConfig := &RuleConfig{Level: level}
							if len(v) > 1 {
								if options, ok := v[1].(map[string]interface{}); ok {
									ruleConfig.Options = options
								}
							}
							enabledRules[ruleName] = ruleConfig
						}
					}
				}
			}
		}
	}

	return enabledRules
}
