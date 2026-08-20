package config

import (
	"sync"

	"github.com/web-infra-dev/rslint/internal/rule"
)

// EffectiveFileConfig is the immutable result of matching one file against
// one config generation. Rules, plugin maps, and parserOptions.project all
// come from this same matched-entry shape.
//
// Values are owned by FileConfigResolver and may be shared by every file with
// the same shape. Callers must treat returned maps and slices as read-only.
type EffectiveFileConfig struct {
	resolver     *FileConfigResolver
	mergedConfig *MergedConfig
	projectEntry int

	rulesOnce    sync.Once
	enabledRules []rule.ConfiguredRule
}

// MergedConfig returns the effective flat config for the matched file shape.
func (plan *EffectiveFileConfig) MergedConfig() *MergedConfig {
	if plan == nil {
		return nil
	}
	return plan.mergedConfig
}

// EnabledRules returns the effective configured rules. Rule materialization is
// independent from project-path resolution so LSP config generations do not
// consult the process-global registry merely to select a TypeScript project.
func (plan *EffectiveFileConfig) EnabledRules() []rule.ConfiguredRule {
	if plan == nil {
		return nil
	}
	plan.rulesOnce.Do(func() {
		plan.enabledRules = GlobalRuleRegistry.GetEnabledRulesForMergedConfig(
			plan.mergedConfig,
			plan.resolver.enforcePlugins,
		)
	})
	return plan.enabledRules
}

// FileConfigResolver resolves config and rules for one immutable config root.
// It interns files with the same exact matched-entry shape for the resolver's
// lifetime and is safe for concurrent use by native and plugin lint workers.
type FileConfigResolver struct {
	config         RslintConfig
	matchDirectory string
	enforcePlugins bool
	// globalIgnorePatterns is parsed once per run instead of per file: the
	// config's global `ignores` set is fixed for the whole run, so
	// re-deriving it (string parsing + allocation) on every ConfigForFile
	// call is pure waste at thousands-of-files scale.
	globalIgnorePatterns []IgnorePattern
	entryIgnorePatterns  [][]IgnorePattern
	directoryBlocks      *directoryBlockMatcher

	filePlans  publishOnceCache[string, *EffectiveFileConfig]
	shapePlans publishOnceCache[configMatchKey, *EffectiveFileConfig]
}

// NewFileConfigResolver creates a per-run resolver for one config root.
func NewFileConfigResolver(config RslintConfig, cwd string, enforcePlugins bool) *FileConfigResolver {
	return newFileConfigResolver(config, cwd, enforcePlugins)
}

func newFileConfigResolver(
	config RslintConfig,
	matchDirectory string,
	enforcePlugins bool,
) *FileConfigResolver {
	globalIgnorePatterns := extractConfigIgnores(config)
	return &FileConfigResolver{
		config:               config,
		matchDirectory:       matchDirectory,
		enforcePlugins:       enforcePlugins,
		globalIgnorePatterns: globalIgnorePatterns,
		entryIgnorePatterns:  parseEntryIgnorePatterns(config),
		directoryBlocks:      newDirectoryBlockMatcher(globalIgnorePatterns, matchDirectory),
	}
}

// ConfigForFile returns the merged config for filePath, caching nil misses.
// The returned value is shared immutable resolver state and must be read-only.
func (r *FileConfigResolver) ConfigForFile(filePath string) *MergedConfig {
	plan := r.PlanForFile(filePath)
	if plan == nil {
		return nil
	}
	return plan.MergedConfig()
}

// EnabledRulesForFile returns cached enabled rules and their merged config.
// Both returned values are shared immutable resolver state and must be read-only.
func (r *FileConfigResolver) EnabledRulesForFile(filePath string) ([]rule.ConfiguredRule, *MergedConfig) {
	plan := r.PlanForFile(filePath)
	if plan == nil {
		return nil, nil
	}
	return plan.EnabledRules(), plan.MergedConfig()
}

// PlanForFile returns the shared effective config for filePath. A nil result
// means no config entry contributes to that target.
func (r *FileConfigResolver) PlanForFile(filePath string) *EffectiveFileConfig {
	return r.filePlans.getOrInit(filePath, func() *EffectiveFileConfig {
		key, matched := r.config.matchConfigEntries(
			filePath,
			r.matchDirectory,
			r.globalIgnorePatterns,
			r.entryIgnorePatterns,
			r.directoryBlocks,
		)
		if !matched {
			return nil
		}

		return r.shapePlans.getOrInit(key, func() *EffectiveFileConfig {
			mergedConfig := r.config.mergeConfigEntries(key)
			return &EffectiveFileConfig{
				resolver:     r,
				mergedConfig: mergedConfig,
				projectEntry: r.config.effectiveProjectEntry(key),
			}
		})
	})
}
