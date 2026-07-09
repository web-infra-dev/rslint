package main

import (
	"sync"

	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/linter"
)

type lintConfigResolver struct {
	configMap              map[string]rslintconfig.RslintConfig
	rslintConfig           rslintconfig.RslintConfig
	currentDirectory       string
	enforcePlugins         bool
	typeInfoFiles          map[string]struct{}
	configPathBySourcePath map[string]string

	mu              sync.Mutex
	singleResolver  *rslintconfig.FileConfigResolver
	configResolvers map[string]*rslintconfig.FileConfigResolver
}

func newLintConfigResolver(
	configMap map[string]rslintconfig.RslintConfig,
	rslintConfig rslintconfig.RslintConfig,
	currentDirectory string,
	enforcePlugins bool,
	typeInfoFiles map[string]struct{},
	configPathBySourcePath map[string]string,
) *lintConfigResolver {
	return &lintConfigResolver{
		configMap:              configMap,
		rslintConfig:           rslintConfig,
		currentDirectory:       currentDirectory,
		enforcePlugins:         enforcePlugins,
		typeInfoFiles:          typeInfoFiles,
		configPathBySourcePath: configPathBySourcePath,
		configResolvers:        make(map[string]*rslintconfig.FileConfigResolver),
	}
}

func (r *lintConfigResolver) resolverForConfig(configDir string, cfg rslintconfig.RslintConfig) *rslintconfig.FileConfigResolver {
	r.mu.Lock()
	defer r.mu.Unlock()
	if resolver := r.configResolvers[configDir]; resolver != nil {
		return resolver
	}
	resolver := rslintconfig.NewFileConfigResolver(cfg, configDir, r.enforcePlugins)
	r.configResolvers[configDir] = resolver
	return resolver
}

func (r *lintConfigResolver) resolverForFile(filePath string) (string, *rslintconfig.FileConfigResolver, bool) {
	if r.configMap != nil {
		cfgDir, cfg := rslintconfig.FindNearestConfig(filePath, r.configMap)
		if cfg == nil {
			return "", nil, false
		}
		return cfgDir, r.resolverForConfig(cfgDir, cfg), true
	}

	r.mu.Lock()
	if r.singleResolver == nil {
		r.singleResolver = rslintconfig.NewFileConfigResolver(r.rslintConfig, r.currentDirectory, r.enforcePlugins)
	}
	resolver := r.singleResolver
	r.mu.Unlock()
	return r.currentDirectory, resolver, true
}

func (r *lintConfigResolver) configPathFor(filePath string) string {
	if r == nil || len(r.configPathBySourcePath) == 0 {
		return filePath
	}
	if configPath := r.configPathBySourcePath[filePath]; configPath != "" {
		return configPath
	}
	return filePath
}

func (r *lintConfigResolver) ConfigForFile(filePath string) *rslintconfig.MergedConfig {
	configPath := r.configPathFor(filePath)
	_, resolver, ok := r.resolverForFile(configPath)
	if !ok {
		return nil
	}
	return resolver.ConfigForFile(configPath)
}

func (r *lintConfigResolver) ActiveRulesForFile(filePath string) []linter.ConfiguredRule {
	configPath := r.configPathFor(filePath)
	_, resolver, ok := r.resolverForFile(configPath)
	if !ok {
		return nil
	}
	activeRules, _ := resolver.EnabledRulesForFile(configPath)
	if r.typeInfoFiles != nil {
		if _, hasTypeInfo := r.typeInfoFiles[filePath]; !hasTypeInfo {
			if _, hasTypeInfo = r.typeInfoFiles[configPath]; !hasTypeInfo {
				activeRules = linter.FilterNonTypeAwareRules(activeRules)
			}
		}
	}
	return activeRules
}
