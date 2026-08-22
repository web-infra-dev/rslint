package config

import (
	"slices"
	"strings"

	"github.com/web-infra-dev/rslint/internal/rule"
)

// ResolveEnabledRules evaluates config for filePath and resolves its enabled
// rule names against catalog. It returns nil rules and a nil merged config when
// no config entry selects the file.
func ResolveEnabledRules(
	catalog *rule.Catalog,
	config RslintConfig,
	filePath string,
	cwd string,
	enforcePlugins bool,
) ([]rule.ConfiguredRule, *MergedConfig) {
	if catalog == nil {
		panic("rule catalog is required")
	}
	mergedConfig := config.GetConfigForFile(filePath, cwd)
	if mergedConfig == nil {
		return nil, nil
	}
	return ConfiguredRules(catalog, mergedConfig, enforcePlugins), mergedConfig
}

// ConfiguredRules converts an already-resolved config into enabled rule
// handlers without re-running files or ignores matching.
func ConfiguredRules(
	catalog *rule.Catalog,
	mergedConfig *MergedConfig,
	enforcePlugins bool,
) []rule.ConfiguredRule {
	if catalog == nil {
		panic("rule catalog is required")
	}
	if mergedConfig == nil {
		return nil
	}

	var environment *rule.RuleEnvironment
	var enabledRules []rule.ConfiguredRule
	for ruleName, ruleConfig := range mergedConfig.Rules {
		if !ruleConfig.IsEnabled() {
			continue
		}

		if enforcePlugins {
			prefix := RulePluginPrefix(ruleName)
			if prefix != "" {
				if _, declared := mergedConfig.Plugins[prefix]; !declared {
					continue
				}
			}
		}

		ruleImpl, exists := catalog.Lookup(ruleName)
		if !exists {
			continue
		}
		if environment == nil {
			environment = &rule.RuleEnvironment{
				Settings:        CloneSettings(mergedConfig.Settings),
				LanguageOptions: ExtractLanguageOptions(mergedConfig.LanguageOptions),
				Globals:         ExtractGlobals(mergedConfig.LanguageOptions),
			}
		}
		ruleConfigCopy := ruleConfig
		options := rule.NormalizeOptions(ruleConfigCopy.Options)
		enabledRules = append(enabledRules, rule.ConfiguredRule{
			Name:               ruleName,
			Environment:        environment,
			Severity:           ruleConfig.GetSeverity(),
			RequiresTypeInfo:   ruleImpl.RequiresTypeInfo,
			IsEslintPluginRule: ruleImpl.IsEslintPluginRule,
			Options:            options,
			Run: func(ctx rule.RuleContext) rule.RuleListeners {
				return ruleImpl.Run(ctx, options)
			},
		})
	}

	// mergedConfig.Rules is a map, so collection order is random per process.
	// Sorting preserves deterministic listener registration and diagnostics.
	slices.SortFunc(enabledRules, func(a, b rule.ConfiguredRule) int {
		return strings.Compare(a.Name, b.Name)
	})
	return enabledRules
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
