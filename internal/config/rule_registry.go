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

	// Pre-compute the set of prefixes contributed by EslintPlugins so the
	// per-rule gate below doesn't repeat the O(n) scan for every rule.
	// EslintPlugins is a slice (multi-version monorepo allows the same
	// prefix at different resolvedPaths), but for the gate we only care
	// that the prefix has been DECLARED somewhere — version routing is
	// done per-file via ConfigKey downstream.
	var eslintPluginPrefixes map[string]struct{}
	if enforcePlugins && len(mergedConfig.EslintPlugins) > 0 {
		eslintPluginPrefixes = make(map[string]struct{}, len(mergedConfig.EslintPlugins))
		for _, e := range mergedConfig.EslintPlugins {
			eslintPluginPrefixes[e.Prefix] = struct{}{}
		}
	}

	var enabledRules []linter.ConfiguredRule
	for ruleName, ruleConfig := range mergedConfig.Rules {
		if ruleConfig.IsEnabled() {
			// Plugin gate: when enforcePlugins is true, skip plugin rules
			// whose plugin is not declared in either `plugins` (native /
			// string-form declarations) or `eslintPlugins` (object-form
			// JS plugin instances). The two declaration channels are
			// equally valid for the purpose of "the user told us this
			// prefix is in play" — gating on only `plugins` makes the
			// documented minimal `eslintPlugins: { uc: pluginObj }` +
			// `rules: { 'uc/no-null': 'error' }` example silently
			// produce zero diagnostics.
			if enforcePlugins {
				prefix := RulePluginPrefix(ruleName)
				if prefix != "" {
					_, inPlugins := mergedConfig.Plugins[prefix]
					_, inEslintPlugins := eslintPluginPrefixes[prefix]
					if !inPlugins && !inEslintPlugins {
						continue
					}
				}
			}

			if ruleImpl, exists := r.rules[ruleName]; exists {
				ruleConfigCopy := ruleConfig
				enabledRules = append(enabledRules, linter.ConfiguredRule{
					Name:               ruleName,
					Settings:           CloneSettings(mergedConfig.Settings),
					Severity:           ruleConfig.GetSeverity(),
					RequiresTypeInfo:   ruleImpl.RequiresTypeInfo,
					IsEslintPluginRule: ruleImpl.IsEslintPluginRule,
					ConfigKey:          cwd, // owning config directory for this file
					Options:            ruleConfigCopy.Options,
					LanguageOptions:    mergedToCompatLangOpts(mergedConfig),
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return ruleImpl.Run(ctx, nativeRuleOptions(ruleConfigCopy.Options))
					},
				})
			}
		}
	}

	return enabledRules, mergedConfig
}

// nativeRuleOptions unwraps a single-element positional options array back
// to the bare element for native (Go) rules, which expect the legacy
// single-value form. config.go now stores the full ESLint-aligned options
// array (so the compat layer's context.options is correct); this restores
// the historic shape for the native rule Run path only.
func nativeRuleOptions(opts interface{}) interface{} {
	if arr, ok := opts.([]interface{}); ok && len(arr) == 1 {
		return arr[0]
	}
	return opts
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

// mergedToCompatLangOpts returns the per-file `languageOptions` blob
// the worker consumes — a plain `map[string]any` flattened from the
// MergedConfig minus the native-only fields (`parserOptions.project`,
// `projectService`). All field-by-field copying / typed unwrapping has
// moved to LanguageOptions.ToCompatWire(); this wrapper exists for the
// call site in GetEnabledRules to keep the signature stable. Returns
// nil when nothing compat-relevant is set so JSON omitempty drops the
// field on the wire.
func mergedToCompatLangOpts(mc *MergedConfig) map[string]any {
	if mc == nil {
		return nil
	}
	return mc.LanguageOptions.ToCompatWire()
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
