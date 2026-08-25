package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	configLint "github.com/web-infra-dev/rslint/internal/config/lint"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// eslintPluginConfigResolver projects the config selected for a Program source into
// the plugin worker's wire identity and serializable config maps. The low-level
// API may replace an owner directory with its distinct opaque routing key;
// CLI discovery uses the Go-owned owner directory directly.
type eslintPluginConfigResolver struct {
	lintResolver                 *configLint.Resolver
	pluginConfigDirectoryByOwner map[string]string
}

func (resolver eslintPluginConfigResolver) resolve(filePath string) linter.EslintPluginFileConfig {
	if resolver.lintResolver == nil {
		return linter.EslintPluginFileConfig{}
	}
	ownerDirectory, resolved, ok := resolver.lintResolver.ResolveSourcePath(filePath)
	if !ok {
		return linter.EslintPluginFileConfig{}
	}
	configKey := ownerDirectory
	if pluginConfigDirectory, ok := resolver.pluginConfigDirectoryByOwner[ownerDirectory]; ok {
		configKey = pluginConfigDirectory
	}
	languageOptions, settings := rslintconfig.PluginMergedMaps(resolved.MergedConfig)
	return linter.EslintPluginFileConfig{
		ConfigKey:       configKey,
		LanguageOptions: languageOptions,
		Settings:        settings,
	}
}

// dispatchEslintPluginRulesAsync runs plugin dispatch concurrently with native
// linting. Dispatch failures are written before diagnostics are published,
// preserving the command surfaces' existing stderr timing even when native
// linting returns early without receiving from the channel.
func dispatchEslintPluginRulesAsync(
	ctx context.Context,
	dispatch linter.EslintPluginDispatcher,
	inputs []linter.EslintPluginFileInput,
	fix bool,
	suggestionsMode string,
	timing *linter.TimingCollector,
) <-chan []rule.RuleDiagnostic {
	ch := make(chan []rule.RuleDiagnostic, 1)
	go func() {
		outcome := linter.DispatchEslintPluginRulesWithOutcome(
			ctx,
			dispatch,
			inputs,
			fix,
			suggestionsMode,
			timing,
		)
		if outcome.DispatchError != nil && !errors.Is(outcome.DispatchError, context.Canceled) {
			fmt.Fprintf(os.Stderr, "rslint: eslint-plugin lint error: %v\n", outcome.DispatchError)
		}
		ch <- outcome.Diagnostics
	}()
	return ch
}
