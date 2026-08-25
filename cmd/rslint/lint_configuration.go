package main

import (
	"fmt"
	"os"
	"slices"

	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rules"
)

func cloneConfigMap(configMap map[string]rslintconfig.RslintConfig) map[string]rslintconfig.RslintConfig {
	if configMap == nil {
		return nil
	}
	cloned := make(map[string]rslintconfig.RslintConfig, len(configMap))
	for dir, cfg := range configMap {
		cloned[dir] = slices.Clone(cfg)
	}
	return cloned
}

func configsForOwners(
	configMap map[string]rslintconfig.RslintConfig,
	owners []string,
) map[string]rslintconfig.RslintConfig {
	if len(configMap) == 0 || len(owners) == 0 {
		return nil
	}
	active := make(map[string]rslintconfig.RslintConfig, len(owners))
	for _, owner := range owners {
		if entries, ok := configMap[owner]; ok {
			active[owner] = entries
		}
	}
	return active
}

// validateResolvedRuleOptions runs rule-options schema validation over the
// resolved configuration: every configMap config in multi-config mode (each
// message suffixed with the owning config directory, since the same rule can
// be misconfigured differently per config), or the single rslintConfig
// otherwise. It returns normalized copies plus one formatted message per
// failure, leaving the input configs untouched.
func validateResolvedRuleOptions(
	configMap map[string]rslintconfig.RslintConfig,
	rslintConfig rslintconfig.RslintConfig,
	catalog *rule.Catalog,
) (map[string]rslintconfig.RslintConfig, rslintconfig.RslintConfig, []string) {
	var messages []string
	if configMap == nil {
		normalized, optionsErrors := rslintconfig.ValidateRuleOptions(rslintConfig, catalog)
		for _, optionsError := range optionsErrors {
			messages = append(messages, optionsError.Error())
		}
		return nil, normalized, messages
	}
	configDirs := make([]string, 0, len(configMap))
	for dir := range configMap {
		configDirs = append(configDirs, dir)
	}
	slices.Sort(configDirs)
	normalizedMap := make(map[string]rslintconfig.RslintConfig, len(configMap))
	for _, dir := range configDirs {
		normalized, optionsErrors := rslintconfig.ValidateRuleOptions(configMap[dir], catalog)
		normalizedMap[dir] = normalized
		for _, optionsError := range optionsErrors {
			messages = append(messages, fmt.Sprintf("%s (config at %s)", optionsError.Error(), dir))
		}
	}
	return normalizedMap, rslintConfig, messages
}

func deriveRuleCatalog(plugins []rslintconfig.EslintPluginEntry) (*rule.Catalog, []string) {
	return rules.All().ForESLintPlugins(plugins)
}

func reportShadowedPluginRules(shadowed []string) {
	for _, ruleName := range shadowed {
		fmt.Fprintf(
			os.Stderr,
			"rslint: plugin rule %q is shadowed by a built-in rule of the same name; using the built-in.\n",
			ruleName,
		)
	}
}
