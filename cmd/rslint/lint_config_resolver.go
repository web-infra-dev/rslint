package main

import (
	"sort"

	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/config/target"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type lintConfigResolver struct {
	configMap              map[string]rslintconfig.RslintConfig
	currentDirectory       string
	lintTargetBySourcePath map[string]target.File
	fsys                   vfs.FS
	singleResolver         *rslintconfig.FileConfigResolver
	configResolvers        map[string]*rslintconfig.FileConfigResolver
}

type lintConfigResolverOptions struct {
	ConfigMap              map[string]rslintconfig.RslintConfig
	Config                 rslintconfig.RslintConfig
	CurrentDirectory       string
	RuleCatalog            *rule.Catalog
	EnforcePlugins         bool
	LintTargetBySourcePath map[string]target.File
	// SourceMappingsCanonical indicates that binding already supplied both
	// lexical and canonical source keys, so normalization needs no filesystem IO.
	SourceMappingsCanonical bool
	PathSpaces              *rslintconfig.PathSpaceSnapshot
	FS                      vfs.FS
}

func newLintConfigResolver(opts lintConfigResolverOptions) *lintConfigResolver {
	if opts.RuleCatalog == nil {
		panic("rule catalog is required")
	}
	if opts.PathSpaces == nil {
		panic("path-space snapshot is required")
	}
	resolver := &lintConfigResolver{
		configMap:              opts.ConfigMap,
		currentDirectory:       opts.CurrentDirectory,
		lintTargetBySourcePath: normalizeSourceTargetMappings(opts.LintTargetBySourcePath, opts.FS, opts.SourceMappingsCanonical),
		fsys:                   opts.FS,
	}
	newFileResolver := func(
		entries rslintconfig.RslintConfig,
		configDirectory string,
	) *rslintconfig.FileConfigResolver {
		fileResolver, err := rslintconfig.NewFileConfigResolverWithPathSpaces(
			entries,
			configDirectory,
			opts.FS,
			opts.PathSpaces,
			opts.RuleCatalog,
			opts.EnforcePlugins,
		)
		if err != nil {
			panic(err)
		}
		return fileResolver
	}
	if opts.ConfigMap == nil {
		resolver.singleResolver = newFileResolver(
			opts.Config,
			opts.CurrentDirectory,
		)
		return resolver
	}
	resolver.configResolvers = make(map[string]*rslintconfig.FileConfigResolver, len(opts.ConfigMap)*2)
	configDirs := make([]string, 0, len(opts.ConfigMap))
	for configDir := range opts.ConfigMap {
		configDirs = append(configDirs, configDir)
	}
	sort.Strings(configDirs)
	for _, configDir := range configDirs {
		cfg := opts.ConfigMap[configDir]
		fileResolver := newFileResolver(
			cfg,
			configDir,
		)
		resolver.configResolvers[configDir] = fileResolver
		resolver.configResolvers[canonicalFilesystemPathID(configDir, opts.FS)] = fileResolver
	}
	return resolver
}

func normalizeSourceTargetMappings(
	mapping map[string]target.File,
	fsys vfs.FS,
	canonicalKeysPresent bool,
) map[string]target.File {
	if len(mapping) == 0 {
		return mapping
	}
	normalized := make(map[string]target.File, len(mapping)*2)
	for sourcePath, target := range mapping {
		target.Path = tspath.NormalizePath(target.Path)
		target.CanonicalPath = tspath.NormalizePath(target.CanonicalPath)
		target.CanonicalParentPath = tspath.NormalizePath(target.CanonicalParentPath)
		target.ConfigDirectory = tspath.NormalizePath(target.ConfigDirectory)
		normalizedPath := exactFilesystemPathID(sourcePath)
		normalized[normalizedPath] = target
		if !canonicalKeysPresent {
			normalized[canonicalFilesystemPathID(normalizedPath, fsys)] = target
		}
	}
	return normalized
}

func lookupLintTarget(
	mapping map[string]target.File,
	filePath string,
	fsys vfs.FS,
) (target.File, bool) {
	if len(mapping) == 0 {
		return target.File{}, false
	}
	if target, ok := mapping[exactFilesystemPathID(filePath)]; ok {
		return target, true
	}
	target, ok := mapping[canonicalFilesystemPathID(filePath, fsys)]
	return target, ok
}

func (r *lintConfigResolver) configResolver(configDir string) *rslintconfig.FileConfigResolver {
	if resolver := r.configResolvers[configDir]; resolver != nil {
		return resolver
	}
	return r.configResolvers[canonicalFilesystemPathID(configDir, r.fsys)]
}

func (r *lintConfigResolver) targetForFile(filePath string) (target.File, bool) {
	if r != nil {
		if target, ok := lookupLintTarget(r.lintTargetBySourcePath, filePath, r.fsys); ok {
			return target, true
		}
	}
	return target.File{PathIdentity: rslintconfig.PathIdentity{Path: tspath.NormalizePath(filePath)}}, false
}

func (r *lintConfigResolver) resolveFile(
	sourcePath string,
) (string, rslintconfig.ResolvedFileConfig, bool) {
	target, bound := r.targetForFile(sourcePath)
	if r.configMap != nil {
		if bound {
			resolver := r.configResolver(target.ConfigDirectory)
			if resolver == nil {
				return "", rslintconfig.ResolvedFileConfig{}, false
			}
			return target.ConfigDirectory, resolver.ResolveTarget(target.Identity()), true
		}
		return "", rslintconfig.ResolvedFileConfig{}, false
	}
	return r.currentDirectory, r.singleResolver.ResolveTarget(target.Identity()), true
}

func (r *lintConfigResolver) ConfigForFile(filePath string) *rslintconfig.MergedConfig {
	_, resolved, ok := r.resolveFile(filePath)
	if !ok {
		return nil
	}
	return resolved.MergedConfig
}

// EnabledRulesForFile returns the complete configured rule set. Program
// capabilities are applied once by the lint planner, not while resolving
// configuration.
func (r *lintConfigResolver) EnabledRulesForFile(filePath string) []rule.ConfiguredRule {
	_, resolved, ok := r.resolveFile(filePath)
	if !ok {
		return nil
	}
	return resolved.EnabledRules
}
