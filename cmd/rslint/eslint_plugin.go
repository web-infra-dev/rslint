package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	configLint "github.com/web-infra-dev/rslint/internal/config/lint"
	"github.com/web-infra-dev/rslint/internal/linter"
)

// eslintPluginConfigResolver projects the CLI-selected config into the plugin
// worker's owner-directory routing identity and serializable config maps.
type eslintPluginConfigResolver struct {
	lintResolver *configLint.Resolver
}

func (resolver eslintPluginConfigResolver) resolve(filePath string) linter.EslintPluginFileConfig {
	if resolver.lintResolver == nil {
		return linter.EslintPluginFileConfig{}
	}
	ownerDirectory, resolved, ok := resolver.lintResolver.ResolveSourcePath(filePath)
	if !ok {
		return linter.EslintPluginFileConfig{}
	}
	languageOptions, settings := rslintconfig.PluginMergedMaps(resolved.MergedConfig)
	return linter.EslintPluginFileConfig{
		ConfigKey:       ownerDirectory,
		LanguageOptions: languageOptions,
		Settings:        settings,
	}
}

func reportEslintPluginDispatchOutcome(outcome linter.EslintPluginDispatchOutcome) {
	writeEslintPluginDispatchOutcome(os.Stderr, outcome)
}

func writeEslintPluginDispatchOutcome(w io.Writer, outcome linter.EslintPluginDispatchOutcome) {
	for _, notice := range outcome.Notices {
		switch notice.Kind {
		case linter.EslintPluginMissingFileResult:
			fmt.Fprintf(w, "rslint: plugin-lint returned no result for %q\n", notice.FilePath)
		case linter.EslintPluginUnconfiguredDiagnostic:
			fmt.Fprintf(w, "rslint: plugin diagnostic for unconfigured rule %q in %q\n", notice.RuleName, notice.FilePath)
		}
	}
	if outcome.DispatchError != nil && !errors.Is(outcome.DispatchError, context.Canceled) {
		fmt.Fprintf(w, "rslint: eslint-plugin lint error: %v\n", outcome.DispatchError)
	}
}
