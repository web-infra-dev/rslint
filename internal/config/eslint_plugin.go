package config

import (
	"fmt"
	"os"

	"github.com/web-infra-dev/rslint/internal/rule"
)

// EslintPluginEntry is the metadata Go receives for one ESLint plugin
// mounted via a config's object-form `plugins`. The live plugin object stays in
// Node (the worker re-imports the config file to obtain it); Go only needs
// the prefix and rule names so it can (a) register placeholder rules that
// make `<prefix>/<rule>` resolvable instead of being silently dropped, and
// (b) route those rules to the Node plugin-lint host (via IsEslintPluginRule)
// instead of trying to run them natively.
type EslintPluginEntry struct {
	Prefix    string   `json:"prefix"`
	RuleNames []string `json:"ruleNames"`
}

// RegisterEslintPluginRules registers a placeholder rule.Rule for every
// "<prefix>/<ruleName>" so the rule resolver treats them as known. The
// placeholder's Run is a no-op — plugin rules never execute in Go; the
// linter splits them out (by rule.Rule.IsEslintPluginRule) and dispatches
// them to the Node worker.
//
// MUST be called after RegisterAllRules() so native rules already exist:
// a same-named native rule wins, in which case the placeholder is skipped
// with a stderr warning (native always takes precedence, so a placeholder
// would shadow nothing yet mislead). Re-registering an existing plugin
// placeholder is a harmless overwrite (idempotent across LSP config
// updates).
func RegisterEslintPluginRules(entries []EslintPluginEntry) {
	for _, entry := range entries {
		if entry.Prefix == "" {
			continue
		}
		for _, ruleName := range entry.RuleNames {
			fullName := entry.Prefix + "/" + ruleName
			if existing, ok := GlobalRuleRegistry.GetRule(fullName); ok && !existing.IsEslintPluginRule {
				fmt.Fprintf(os.Stderr,
					"rslint: plugin rule %q is shadowed by a built-in rule of the same name; using the built-in.\n",
					fullName)
				continue
			}
			GlobalRuleRegistry.Register(fullName, rule.Rule{
				Name:               fullName,
				RequiresTypeInfo:   false,
				IsEslintPluginRule: true,
				Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
					// Never executed: plugin rules run in the Node worker.
					return rule.RuleListeners{}
				},
			})
		}
	}
}

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
