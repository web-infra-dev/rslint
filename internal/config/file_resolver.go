package config

import "github.com/web-infra-dev/rslint/internal/linter"

type effectiveConfigPlan struct {
	mergedConfig *MergedConfig
	enabledRules []linter.ConfiguredRule
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

	filePlans  publishOnceCache[string, *effectiveConfigPlan]
	shapePlans publishOnceCache[configMatchKey, *effectiveConfigPlan]
}

// NewFileConfigResolver creates a per-run resolver for one config root.
func NewFileConfigResolver(config RslintConfig, cwd string, enforcePlugins bool) *FileConfigResolver {
	return &FileConfigResolver{
		config:               config,
		cwd:                  cwd,
		enforcePlugins:       enforcePlugins,
		globalIgnorePatterns: extractConfigIgnores(config),
		entryIgnorePatterns:  parseEntryIgnorePatterns(config),
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
func (r *FileConfigResolver) EnabledRulesForFile(filePath string) ([]linter.ConfiguredRule, *MergedConfig) {
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

// ActiveRulesForFile filters cached enabled rules by the optional type-info set.
func (r *FileConfigResolver) ActiveRulesForFile(filePath string, typeInfoFiles map[string]struct{}) []linter.ConfiguredRule {
	activeRules, _ := r.EnabledRulesForFile(filePath)
	if typeInfoFiles != nil {
		if _, hasTypeInfo := typeInfoFiles[filePath]; !hasTypeInfo {
			activeRules = linter.FilterNonTypeAwareRules(activeRules)
		}
	}
	return activeRules
}

// ActiveRulesForFileHasTypeInfo filters cached enabled rules by a known type-info flag.
func (r *FileConfigResolver) ActiveRulesForFileHasTypeInfo(filePath string, hasTypeInfo bool) []linter.ConfiguredRule {
	activeRules, _ := r.EnabledRulesForFile(filePath)
	if !hasTypeInfo {
		activeRules = linter.FilterNonTypeAwareRules(activeRules)
	}
	return activeRules
}
