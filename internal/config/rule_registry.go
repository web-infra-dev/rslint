package config

import (
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/rule"
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

// GetEnabledRules returns rules that are enabled in the configuration for a given file.
// Returns nil if no config entry matches the file (file should not be linted).
// cwd is the config directory used to resolve files/ignores patterns.
func (r *RuleRegistry) GetEnabledRules(config RslintConfig, filePath string, cwd string) ([]linter.ConfiguredRule, *MergedConfig) {
	mergedConfig := config.GetConfigForFile(filePath, cwd)
	if mergedConfig == nil {
		return nil, nil // file is globally ignored
	}

	var enabledRules []linter.ConfiguredRule
	for ruleName, ruleConfig := range mergedConfig.Rules {
		if ruleConfig.IsEnabled() {
			if ruleImpl, exists := r.rules[ruleName]; exists {
				ruleConfigCopy := ruleConfig
				enabledRules = append(enabledRules, linter.ConfiguredRule{
					Name:     ruleName,
					Severity: ruleConfig.GetSeverity(),
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return ruleImpl.Run(ctx, ruleConfigCopy.Options)
					},
				})
			}
		}
	}

	return enabledRules, mergedConfig
}

// Global rule registry instance
var GlobalRuleRegistry = NewRuleRegistry()
