package config

import (
	"github.com/typescript-eslint/rslint/internal/rule"
)

// ConfiguredRule represents a rule with its configuration level
type ConfiguredRule struct {
	Name  string
	Level rule.DiagnosticLevel
	Run   func(ctx rule.RuleContext, options any) rule.RuleListeners
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
func (r *RuleRegistry) GetEnabledRules(config RslintConfig, filePath string) []ConfiguredRule {
	enabledRuleConfigs := config.GetRulesForFile(filePath)
	var enabledRules []ConfiguredRule

	for ruleName, ruleConfig := range enabledRuleConfigs {
		// Parse the diagnostic level from the rule configuration
		level := rule.ParseDiagnosticLevel(ruleConfig.GetLevel())

		// Only include rules that are not "off"
		if level != rule.DiagnosticLevelOff {
			if ruleImpl, exists := r.rules[ruleName]; exists {
				enabledRules = append(enabledRules, ConfiguredRule{
					Name:  ruleName,
					Level: level,
					Run:   ruleImpl.Run,
				})
			}
		}
	}

	return enabledRules
}

// Global rule registry instance
var GlobalRuleRegistry = NewRuleRegistry()
