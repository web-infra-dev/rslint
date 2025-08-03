package config

import (
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// EnabledRuleWithConfig combines a rule with its configuration
type EnabledRuleWithConfig struct {
	Rule   rule.Rule
	Config *RuleConfig
}

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
				// Capture the ruleConfig in the closure to pass options correctly
				ruleConfigCopy := ruleConfig
				enabledRules = append(enabledRules, linter.ConfiguredRule{
					Name:     ruleName, // Use the registered rule name, not the implementation name
					Severity: ruleConfig.GetSeverity(),
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return ruleImpl.Run(ctx, ruleConfigCopy.Options)
					},
				})
			}
		}
	}

	return enabledRules
}

// GetEnabledRulesWithConfig returns rules with their configurations for a given file
func (r *RuleRegistry) GetEnabledRulesWithConfig(config RslintConfig, filePath string) []EnabledRuleWithConfig {
	enabledRuleConfigs := config.GetRulesForFile(filePath)
	var enabledRules []EnabledRuleWithConfig

	for ruleName, ruleConfig := range enabledRuleConfigs {
		if ruleConfig.IsEnabled() {
			if ruleImpl, exists := r.rules[ruleName]; exists {
				enabledRules = append(enabledRules, EnabledRuleWithConfig{
					Rule:   ruleImpl,
					Config: ruleConfig,
				})
			}
		}
	}

	return enabledRules
}

// Global rule registry instance
var GlobalRuleRegistry = NewRuleRegistry()
