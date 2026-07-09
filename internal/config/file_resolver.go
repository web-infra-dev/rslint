package config

import (
	"sync"

	"github.com/web-infra-dev/rslint/internal/linter"
)

type cachedMergedConfig struct {
	value *MergedConfig
}

type cachedEnabledRules struct {
	value []linter.ConfiguredRule
}

// FileConfigResolver caches per-file config and rule resolution for one lint
// run. It is safe for concurrent use by RunLinter workers.
type FileConfigResolver struct {
	config         RslintConfig
	cwd            string
	enforcePlugins bool

	mu          sync.RWMutex
	configCache map[string]cachedMergedConfig
	rulesCache  map[string]cachedEnabledRules
}

// NewFileConfigResolver creates a per-run resolver for one config root.
func NewFileConfigResolver(config RslintConfig, cwd string, enforcePlugins bool) *FileConfigResolver {
	return &FileConfigResolver{
		config:         config,
		cwd:            cwd,
		enforcePlugins: enforcePlugins,
		configCache:    make(map[string]cachedMergedConfig),
		rulesCache:     make(map[string]cachedEnabledRules),
	}
}

// ConfigForFile returns the merged config for filePath, caching nil misses.
func (r *FileConfigResolver) ConfigForFile(filePath string) *MergedConfig {
	r.mu.RLock()
	if cached, ok := r.configCache[filePath]; ok {
		r.mu.RUnlock()
		return cached.value
	}
	r.mu.RUnlock()

	merged := r.config.GetConfigForFile(filePath, r.cwd)

	r.mu.Lock()
	if cached, ok := r.configCache[filePath]; ok {
		r.mu.Unlock()
		return cached.value
	}
	r.configCache[filePath] = cachedMergedConfig{value: merged}
	r.mu.Unlock()
	return merged
}

// EnabledRulesForFile returns cached enabled rules and their merged config.
// The returned rule slice is shared cache state and must be treated read-only.
func (r *FileConfigResolver) EnabledRulesForFile(filePath string) ([]linter.ConfiguredRule, *MergedConfig) {
	r.mu.RLock()
	if cached, ok := r.rulesCache[filePath]; ok {
		merged := r.configCache[filePath].value
		r.mu.RUnlock()
		return cached.value, merged
	}
	r.mu.RUnlock()

	merged := r.ConfigForFile(filePath)
	enabledRules := GlobalRuleRegistry.GetEnabledRulesForMergedConfig(merged, r.enforcePlugins)

	r.mu.Lock()
	if cached, ok := r.rulesCache[filePath]; ok {
		merged = r.configCache[filePath].value
		r.mu.Unlock()
		return cached.value, merged
	}
	r.rulesCache[filePath] = cachedEnabledRules{value: enabledRules}
	r.mu.Unlock()
	return enabledRules, merged
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
