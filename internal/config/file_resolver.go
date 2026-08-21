package config

import "github.com/web-infra-dev/rslint/internal/rule"

type effectiveConfigPlan struct {
	mergedConfig *MergedConfig
	enabledRules []rule.ConfiguredRule
}

// FileConfigResolver resolves config and rules for one immutable config root.
// It interns files with the same exact matched-entry shape for the resolver's
// lifetime and is safe for concurrent use by native and plugin lint workers.
type FileConfigResolver struct {
	config         RslintConfig
	cwd            string
	enforcePlugins bool
	// globalIgnorePatterns is parsed once per run instead of per file: the
	// config's global `ignores` set is fixed for the whole run, so
	// re-deriving it (string parsing + allocation) on every ConfigForFile
	// call is pure waste at thousands-of-files scale.
	globalIgnorePatterns []IgnorePattern
	entryIgnorePatterns  [][]IgnorePattern
	directoryBlocks      *directoryBlockMatcher

	filePlans  publishOnceCache[string, *effectiveConfigPlan]
	shapePlans publishOnceCache[configMatchKey, *effectiveConfigPlan]
}

// NewFileConfigResolver creates a per-run resolver for one config root.
func NewFileConfigResolver(config RslintConfig, cwd string, enforcePlugins bool) *FileConfigResolver {
	globalIgnorePatterns := extractConfigIgnores(config)
	return &FileConfigResolver{
		config:               config,
		cwd:                  cwd,
		enforcePlugins:       enforcePlugins,
		globalIgnorePatterns: globalIgnorePatterns,
		entryIgnorePatterns:  parseEntryIgnorePatterns(config),
		directoryBlocks:      newDirectoryBlockMatcher(globalIgnorePatterns, cwd),
	}
}

// ConfigForFile returns the merged config for filePath, caching nil misses.
// The returned value is shared immutable resolver state and must be read-only.
func (r *FileConfigResolver) ConfigForFile(filePath string) *MergedConfig {
	plan := r.planForFile(filePath)
	if plan == nil {
		return nil
	}
	return plan.mergedConfig
}

// EnabledRulesForFile returns cached enabled rules and their merged config.
// Both returned values are shared immutable resolver state and must be read-only.
func (r *FileConfigResolver) EnabledRulesForFile(filePath string) ([]rule.ConfiguredRule, *MergedConfig) {
	plan := r.planForFile(filePath)
	if plan == nil {
		return nil, nil
	}
	return plan.enabledRules, plan.mergedConfig
}

func (r *FileConfigResolver) planForFile(filePath string) *effectiveConfigPlan {
	return r.filePlans.getOrInit(filePath, func() *effectiveConfigPlan {
		key, matched := r.config.matchConfigEntries(
			filePath,
			r.cwd,
			r.globalIgnorePatterns,
			r.entryIgnorePatterns,
			r.directoryBlocks,
		)
		if !matched {
			return nil
		}

		return r.shapePlans.getOrInit(key, func() *effectiveConfigPlan {
			mergedConfig := r.config.mergeConfigEntries(key)
			return &effectiveConfigPlan{
				mergedConfig: mergedConfig,
				enabledRules: GlobalRuleRegistry.GetEnabledRulesForMergedConfig(mergedConfig, r.enforcePlugins),
			}
		})
	})
}
