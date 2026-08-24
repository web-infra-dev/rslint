package config

import "github.com/web-infra-dev/rslint/internal/rule"

// EslintPluginEntry is the metadata Go receives for one ESLint plugin
// mounted via a config's object-form `plugins`. The live plugin object stays in
// Node (the worker re-imports the config file to obtain it); Go only needs the
// prefix and rule names to build the run-scoped rule catalog.
type EslintPluginEntry = rule.ESLintPluginMetadata

// PluginMergedMaps extracts the per-file languageOptions (the raw map) and
// settings a plugin-lint dispatch needs from a resolved MergedConfig, so the
// linter-side assembly (linter.BuildEslintPluginFileInput) stays free of the
// config type. languageOptions is nil when merged or its LanguageOptions is
// nil; settings is nil only when merged is nil (otherwise it is merged.Settings,
// which is itself nil when the config declares none).
func PluginMergedMaps(merged *MergedConfig) (languageOptions, settings map[string]any) {
	if merged == nil {
		return nil, nil
	}
	settings = merged.Settings
	if merged.LanguageOptions != nil {
		languageOptions = merged.LanguageOptions.Raw
	}
	return languageOptions, settings
}
