package main

import (
	"fmt"
	"os"
	"slices"

	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
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

func reportShadowedPluginRules(shadowed []string) {
	for _, ruleName := range shadowed {
		fmt.Fprintf(
			os.Stderr,
			"rslint: plugin rule %q is shadowed by a built-in rule of the same name; using the built-in.\n",
			ruleName,
		)
	}
}
