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
// When enforcePlugins is true (JS/TS config), rules with a plugin prefix (e.g. "@typescript-eslint/")
// are only included if the corresponding plugin is declared in the merged config's Plugins set.
// Core rules (no "/" prefix) are always included regardless of enforcePlugins.
// cwd is the config directory used to resolve files/ignores patterns.
func (r *RuleRegistry) GetEnabledRules(config RslintConfig, filePath string, cwd string, enforcePlugins bool) ([]linter.ConfiguredRule, *MergedConfig) {
	mergedConfig := config.GetConfigForFile(filePath, cwd)
	if mergedConfig == nil {
		return nil, nil // file is globally ignored
	}

	var enabledRules []linter.ConfiguredRule
	for ruleName, ruleConfig := range mergedConfig.Rules {
		if ruleConfig.IsEnabled() {
			// Plugin gate: when enforcePlugins is true, skip plugin rules
			// whose plugin is not declared in the merged plugins set.
			if enforcePlugins {
				prefix := RulePluginPrefix(ruleName)
				if prefix != "" {
					if _, declared := mergedConfig.Plugins[prefix]; !declared {
						continue
					}
				}
			}

			if ruleImpl, exists := r.rules[ruleName]; exists {
				ruleConfigCopy := ruleConfig
				enabledRules = append(enabledRules, linter.ConfiguredRule{
					Name:             ruleName,
					Settings:         CloneSettings(mergedConfig.Settings),
					Severity:         ruleConfig.GetSeverity(),
					RequiresTypeInfo: ruleImpl.RequiresTypeInfo,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return ruleImpl.Run(ctx, ruleConfigCopy.Options)
					},
				})
			}
		}
	}

	return enabledRules, mergedConfig
}

// GetActiveRulesForFile returns the lint rules that should run on a file.
// It resolves the config, gets enabled rules, and filters out type-aware rules
// for files not covered by parserOptions.project tsconfigs. This encapsulates
// the rule selection logic shared by both CLI and LSP.
func (r *RuleRegistry) GetActiveRulesForFile(
	rslintConfig RslintConfig,
	filePath string,
	cwd string,
	enforcePlugins bool,
	typeInfoFiles map[string]struct{},
) []linter.ConfiguredRule {
	activeRules, _ := r.GetEnabledRules(rslintConfig, filePath, cwd, enforcePlugins)
	if typeInfoFiles != nil {
		if _, hasTypeInfo := typeInfoFiles[filePath]; !hasTypeInfo {
			activeRules = linter.FilterNonTypeAwareRules(activeRules)
		}
	}
	return activeRules
}

func CloneSettings(settings map[string]interface{}) map[string]interface{} {
	if len(settings) == 0 {
		return nil
	}

	cloned := make(map[string]interface{}, len(settings))
	for k, v := range settings {
		cloned[k] = v
	}
	return cloned
}

// Global rule registry instance
var GlobalRuleRegistry = NewRuleRegistry()
