package config

import (
	"github.com/typescript-eslint/rslint/internal/linter"
	"github.com/typescript-eslint/rslint/internal/rule"
)

// RuleRegistry manages all available rules
type RuleRegistry struct {
	rules map[string]rule.Rule
}

// NewRuleRegistry creates a new rule registry
func NewRuleRegistry() *RuleRegistry {
	return &RuleRegistry{
		rules: make(map[string]rule.Rule),
	}
}

// Register adds a rule to the registry
func (r *RuleRegistry) Register(ruleName string, ruleImpl rule.Rule) {
	r.rules[ruleName] = ruleImpl
}

// GetRule returns a rule by name
func (r *RuleRegistry) GetRule(name string) (rule.Rule, bool) {
	rule, exists := r.rules[name]
	return rule, exists
}

// GetAllRules returns all registered rules
func (r *RuleRegistry) GetAllRules() map[string]rule.Rule {
	return r.rules
}

// GetEnabledRules returns rules that are enabled in the configuration for a given file
func (r *RuleRegistry) GetEnabledRules(config RslintConfig, filePath string) []linter.ConfiguredRule {
	enabledRuleConfigs := config.GetRulesForFile(filePath)
	var enabledRules []linter.ConfiguredRule

	for ruleName, ruleConfig := range enabledRuleConfigs {

		if ruleConfig.IsEnabled() {
			if ruleImpl, exists := r.rules[ruleName]; exists {
				enabledRules = append(enabledRules, linter.ConfiguredRule{
					Name:     ruleImpl.Name,
					Severity: ruleConfig.GetSeverity(),
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return ruleImpl.Run(ctx, nil)
					},
				})
			}
		}
	}

	return enabledRules
}

// Global rule registry instance
var GlobalRuleRegistry = NewRuleRegistry()
