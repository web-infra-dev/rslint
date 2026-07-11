package main

import (
	"sort"

	"github.com/microsoft/typescript-go/shim/vfs"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/linter"
)

type lintConfigResolver struct {
	configMap                  map[string]rslintconfig.RslintConfig
	currentDirectory           string
	typeInfoFiles              map[string]struct{}
	configPathBySourcePath     map[string]string
	ownerConfigDirBySourcePath map[string]string
	fsys                       vfs.FS
	singleResolver             *rslintconfig.FileConfigResolver
	configResolvers            map[string]*rslintconfig.FileConfigResolver
	matchConfigMap             map[string]rslintconfig.RslintConfig
	ownerByMatchConfigDir      map[string]string
	configOwnerResolver        *rslintconfig.ConfigOwnerResolver
	matchConfigOwnerResolver   *rslintconfig.ConfigOwnerResolver
}

type lintConfigResolverOptions struct {
	ConfigMap                  map[string]rslintconfig.RslintConfig
	Config                     rslintconfig.RslintConfig
	CurrentDirectory           string
	EnforcePlugins             bool
	TypeInfoFiles              map[string]struct{}
	ConfigPathBySourcePath     map[string]string
	OwnerConfigDirBySourcePath map[string]string
	// SourceMappingsCanonical indicates that binding already supplied both
	// lexical and canonical source keys, so normalization needs no filesystem IO.
	SourceMappingsCanonical bool
	FS                      vfs.FS
}

func newLintConfigResolver(opts lintConfigResolverOptions) *lintConfigResolver {
	resolver := &lintConfigResolver{
		configMap:                  opts.ConfigMap,
		currentDirectory:           opts.CurrentDirectory,
		typeInfoFiles:              opts.TypeInfoFiles,
		configPathBySourcePath:     normalizeSourcePathMappings(opts.ConfigPathBySourcePath, opts.FS, opts.SourceMappingsCanonical),
		ownerConfigDirBySourcePath: normalizeSourcePathMappings(opts.OwnerConfigDirBySourcePath, opts.FS, opts.SourceMappingsCanonical),
		fsys:                       opts.FS,
	}
	if opts.ConfigMap == nil {
		matchRoot := authoritativeFilesystemPath(opts.CurrentDirectory, opts.FS)
		resolver.singleResolver = rslintconfig.NewFileConfigResolver(opts.Config, matchRoot, opts.EnforcePlugins)
		return resolver
	}
	resolver.configResolvers = make(map[string]*rslintconfig.FileConfigResolver, len(opts.ConfigMap)*2)
	resolver.matchConfigMap = make(map[string]rslintconfig.RslintConfig, len(opts.ConfigMap))
	resolver.ownerByMatchConfigDir = make(map[string]string, len(opts.ConfigMap))
	resolver.configOwnerResolver = rslintconfig.NewConfigOwnerResolver(opts.ConfigMap, opts.FS)
	configDirs := make([]string, 0, len(opts.ConfigMap))
	for configDir := range opts.ConfigMap {
		configDirs = append(configDirs, configDir)
	}
	sort.Strings(configDirs)
	for _, configDir := range configDirs {
		cfg := opts.ConfigMap[configDir]
		matchRoot := authoritativeFilesystemPath(configDir, opts.FS)
		fileResolver := rslintconfig.NewFileConfigResolver(cfg, matchRoot, opts.EnforcePlugins)
		resolver.configResolvers[configDir] = fileResolver
		resolver.configResolvers[canonicalFilesystemPathID(configDir, opts.FS)] = fileResolver
		matchRootID := canonicalFilesystemPathID(matchRoot, opts.FS)
		if _, exists := resolver.ownerByMatchConfigDir[matchRootID]; !exists {
			resolver.matchConfigMap[matchRoot] = cfg
			resolver.ownerByMatchConfigDir[matchRootID] = configDir
		}
	}
	resolver.matchConfigOwnerResolver = rslintconfig.NewConfigOwnerResolver(resolver.matchConfigMap, opts.FS)
	return resolver
}

func normalizeSourcePathMappings(mapping map[string]string, fsys vfs.FS, canonicalKeysPresent bool) map[string]string {
	if len(mapping) == 0 {
		return mapping
	}
	normalized := make(map[string]string, len(mapping)*2)
	for sourcePath, value := range mapping {
		normalizedPath := exactFilesystemPathID(sourcePath)
		normalized[normalizedPath] = value
		if !canonicalKeysPresent {
			normalized[canonicalFilesystemPathID(normalizedPath, fsys)] = value
		}
	}
	return normalized
}

func (r *lintConfigResolver) pathMappingValue(mapping map[string]string, filePath string) string {
	if len(mapping) == 0 {
		return ""
	}
	if value := mapping[exactFilesystemPathID(filePath)]; value != "" {
		return value
	}
	return mapping[canonicalFilesystemPathID(filePath, r.fsys)]
}

func (r *lintConfigResolver) configResolver(configDir string) *rslintconfig.FileConfigResolver {
	if resolver := r.configResolvers[configDir]; resolver != nil {
		return resolver
	}
	return r.configResolvers[canonicalFilesystemPathID(configDir, r.fsys)]
}

func (r *lintConfigResolver) resolverForFile(sourcePath string, configPath string) (string, *rslintconfig.FileConfigResolver, bool) {
	if r.configMap != nil {
		ownerConfigDir := r.pathMappingValue(r.ownerConfigDirBySourcePath, sourcePath)
		if ownerConfigDir != "" {
			resolver := r.configResolver(ownerConfigDir)
			return ownerConfigDir, resolver, resolver != nil
		}

		// Compatibility fallback for callers that do not provide a target binding
		// (for example the legacy plugin protocol and focused resolver tests).
		cfgDir, cfg := r.configOwnerResolver.Resolve(sourcePath)
		if cfg != nil {
			resolver := r.configResolver(cfgDir)
			return cfgDir, resolver, resolver != nil
		}
		matchDir, cfg := r.matchConfigOwnerResolver.Resolve(configPath)
		if cfg == nil {
			return "", nil, false
		}
		ownerConfigDir = r.ownerByMatchConfigDir[canonicalFilesystemPathID(matchDir, r.fsys)]
		resolver := r.configResolver(ownerConfigDir)
		return ownerConfigDir, resolver, resolver != nil
	}
	return r.currentDirectory, r.singleResolver, true
}

func (r *lintConfigResolver) configPathFor(filePath string) string {
	if r == nil || len(r.configPathBySourcePath) == 0 {
		return filePath
	}
	if configPath := r.pathMappingValue(r.configPathBySourcePath, filePath); configPath != "" {
		return configPath
	}
	return filePath
}

func (r *lintConfigResolver) ConfigForFile(filePath string) *rslintconfig.MergedConfig {
	configPath := r.configPathFor(filePath)
	_, resolver, ok := r.resolverForFile(filePath, configPath)
	if !ok {
		return nil
	}
	return resolver.ConfigForFile(configPath)
}

func (r *lintConfigResolver) ActiveRulesForFile(filePath string) []linter.ConfiguredRule {
	configPath := r.configPathFor(filePath)
	_, resolver, ok := r.resolverForFile(filePath, configPath)
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
