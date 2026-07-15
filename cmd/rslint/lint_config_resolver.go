package main

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/vfs"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type lintConfigResolver struct {
	configMap                  map[string]rslintconfig.RslintConfig
	currentDirectory           string
	typeInfoFiles              map[string]struct{}
	configPathBySourcePath     map[string]string
	ownerConfigDirBySourcePath map[string]string
	mergedConfigBySourcePath   map[string]*rslintconfig.MergedConfig
	fsys                       vfs.FS
	enforcePlugins             bool
	singleResolver             *rslintconfig.FileConfigResolver
	configDirectoryIDs         map[string]struct{}
}

type lintConfigResolverOptions struct {
	ConfigMap                  map[string]rslintconfig.RslintConfig
	Config                     rslintconfig.RslintConfig
	CurrentDirectory           string
	EnforcePlugins             bool
	TypeInfoFiles              map[string]struct{}
	ConfigPathBySourcePath     map[string]string
	OwnerConfigDirBySourcePath map[string]string
	MergedConfigBySourcePath   map[string]*rslintconfig.MergedConfig
	RequiredSourcePaths        []string
	// SourceMappingsCanonical indicates that binding already supplied both
	// lexical and canonical source keys, so normalization needs no filesystem IO.
	SourceMappingsCanonical bool
	FS                      vfs.FS
}

func newLintConfigResolver(opts lintConfigResolverOptions) (*lintConfigResolver, error) {
	resolver := &lintConfigResolver{
		configMap:                  opts.ConfigMap,
		currentDirectory:           opts.CurrentDirectory,
		typeInfoFiles:              opts.TypeInfoFiles,
		configPathBySourcePath:     normalizeSourcePathMappings(opts.ConfigPathBySourcePath, opts.FS, opts.SourceMappingsCanonical),
		ownerConfigDirBySourcePath: normalizeSourcePathMappings(opts.OwnerConfigDirBySourcePath, opts.FS, opts.SourceMappingsCanonical),
		mergedConfigBySourcePath:   normalizeMergedConfigMappings(opts.MergedConfigBySourcePath, opts.FS, opts.SourceMappingsCanonical),
		fsys:                       opts.FS,
		enforcePlugins:             opts.EnforcePlugins,
	}
	if opts.ConfigMap == nil {
		resolver.singleResolver = rslintconfig.NewFileConfigResolver(
			opts.Config,
			opts.CurrentDirectory,
			opts.EnforcePlugins,
			opts.FS,
		)
		return resolver, nil
	}
	resolver.configDirectoryIDs = make(map[string]struct{}, len(opts.ConfigMap)*2)
	for configDir := range opts.ConfigMap {
		resolver.configDirectoryIDs[exactHostPathID(configDir)] = struct{}{}
		resolver.configDirectoryIDs[canonicalHostPathID(configDir, opts.FS)] = struct{}{}
	}
	// A modern JS/TS discovery binding must carry the exact ConfigArray
	// selection made during discovery for every source identity. Replaying the
	// serializable entries here is not equivalent when files/ignores contain
	// live predicates, so a missing pair is an invariant violation, never a
	// fallback opportunity.
	for sourcePathID, ownerDir := range resolver.ownerConfigDirBySourcePath {
		if merged, ok := resolver.mergedConfigBySourcePath[sourcePathID]; !ok || merged == nil {
			return nil, fmt.Errorf("config discovery invariant: source %q has owner %q but no exact merged config", sourcePathID, ownerDir)
		}
		if _, ok := resolver.configDirectoryIDs[exactHostPathID(ownerDir)]; !ok {
			if _, ok = resolver.configDirectoryIDs[canonicalHostPathID(ownerDir, opts.FS)]; !ok {
				return nil, fmt.Errorf("config discovery invariant: source %q references unknown owner %q", sourcePathID, ownerDir)
			}
		}
	}
	for _, sourcePath := range opts.RequiredSourcePaths {
		ownerDir, ownerOK := resolver.ownerConfigDirForFile(sourcePath)
		if !ownerOK {
			return nil, fmt.Errorf("config discovery invariant: target source %q has no exact owner", sourcePath)
		}
		if merged, mergedOK := resolver.mergedConfigForFile(sourcePath); !mergedOK || merged == nil {
			return nil, fmt.Errorf("config discovery invariant: target source %q owned by %q has no exact merged config", sourcePath, ownerDir)
		}
	}
	return resolver, nil
}

func normalizeMergedConfigMappings(
	mapping map[string]*rslintconfig.MergedConfig,
	fsys vfs.FS,
	canonicalKeysPresent bool,
) map[string]*rslintconfig.MergedConfig {
	if len(mapping) == 0 {
		return mapping
	}
	normalized := make(map[string]*rslintconfig.MergedConfig, len(mapping)*2)
	for sourcePath, merged := range mapping {
		normalizedPath := compilerPathID(sourcePath)
		normalized[normalizedPath] = merged
		if !canonicalKeysPresent {
			normalized[compilerPathID(authoritativeHostPath(sourcePath, fsys))] = merged
		}
	}
	return normalized
}

func normalizeSourcePathMappings(mapping map[string]string, fsys vfs.FS, canonicalKeysPresent bool) map[string]string {
	if len(mapping) == 0 {
		return mapping
	}
	normalized := make(map[string]string, len(mapping)*2)
	for sourcePath, value := range mapping {
		normalizedPath := compilerPathID(sourcePath)
		normalized[normalizedPath] = value
		if !canonicalKeysPresent {
			normalized[compilerPathID(authoritativeHostPath(sourcePath, fsys))] = value
		}
	}
	return normalized
}

func (r *lintConfigResolver) pathMappingValue(mapping map[string]string, filePath string) string {
	if len(mapping) == 0 {
		return ""
	}
	if value := mapping[compilerPathID(filePath)]; value != "" {
		return value
	}
	return mapping[compilerPathID(authoritativeHostPath(filePath, r.fsys))]
}

func (r *lintConfigResolver) ownerConfigDirForFile(sourcePath string) (string, bool) {
	if r.configMap != nil {
		ownerConfigDir := r.pathMappingValue(r.ownerConfigDirBySourcePath, sourcePath)
		if ownerConfigDir == "" {
			return "", false
		}
		_, direct := r.configDirectoryIDs[exactHostPathID(ownerConfigDir)]
		_, canonical := r.configDirectoryIDs[canonicalHostPathID(ownerConfigDir, r.fsys)]
		return ownerConfigDir, direct || canonical
	}
	return r.currentDirectory, r.singleResolver != nil
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
	if r.configMap != nil {
		merged, _ := r.mergedConfigForFile(filePath)
		return merged
	}
	configPath := r.configPathFor(filePath)
	if r.singleResolver == nil {
		return nil
	}
	return r.singleResolver.ConfigForFile(configPath)
}

func (r *lintConfigResolver) mergedConfigForFile(filePath string) (*rslintconfig.MergedConfig, bool) {
	if r == nil || len(r.mergedConfigBySourcePath) == 0 {
		return nil, false
	}
	if merged, ok := r.mergedConfigBySourcePath[compilerPathID(filePath)]; ok && merged != nil {
		return merged, true
	}
	merged, ok := r.mergedConfigBySourcePath[compilerPathID(authoritativeHostPath(filePath, r.fsys))]
	return merged, ok && merged != nil
}

func (r *lintConfigResolver) ActiveRulesForFile(filePath string) []linter.ConfiguredRule {
	if r.configMap != nil {
		merged, ok := r.mergedConfigForFile(filePath)
		if !ok {
			return nil
		}
		activeRules := rslintconfig.GlobalRuleRegistry.GetEnabledRulesForMergedConfig(merged, r.enforcePlugins)
		return r.filterTypeAwareRules(filePath, r.configPathFor(filePath), activeRules)
	}
	configPath := r.configPathFor(filePath)
	if r.singleResolver == nil {
		return nil
	}
	activeRules, _ := r.singleResolver.EnabledRulesForFile(configPath)
	return r.filterTypeAwareRules(filePath, configPath, activeRules)
}

func (r *lintConfigResolver) filterTypeAwareRules(filePath string, configPath string, activeRules []linter.ConfiguredRule) []linter.ConfiguredRule {
	if r.typeInfoFiles != nil {
		if _, hasTypeInfo := r.typeInfoFiles[filePath]; !hasTypeInfo {
			if _, hasTypeInfo = r.typeInfoFiles[configPath]; !hasTypeInfo {
				activeRules = linter.FilterNonTypeAwareRules(activeRules)
			}
		}
	}
	return activeRules
}

// buildBindingLintProgramViews wires the target binding's lexical identities
// into the linter's phase-1 view model. Config resolution and diagnostic path
// remapping are deliberately view-local because one Program SourceFile may be
// selected through multiple aliases governed by different configs.
func buildBindingLintProgramViews(
	binding lintTargetBinding,
	resolverOptions lintConfigResolverOptions,
	onDiagnostic linter.DiagnosticHandler,
) ([]linter.LintProgramView, []*lintConfigResolver, error) {
	views := make([]linter.LintProgramView, 0, len(binding.Views))
	resolvers := make([]*lintConfigResolver, 0, len(binding.Views))
	for _, bindingView := range binding.Views {
		if bindingView.ProgramIndex < 0 || bindingView.ProgramIndex >= len(binding.Programs) {
			continue
		}
		viewOptions := resolverOptions
		viewOptions.TypeInfoFiles = bindingView.TypeInfoFiles
		viewOptions.ConfigPathBySourcePath = bindingView.ConfigPathBySourcePath
		viewOptions.OwnerConfigDirBySourcePath = bindingView.OwnerConfigDirBySourcePath
		viewOptions.MergedConfigBySourcePath = bindingView.MergedConfigBySourcePath
		viewOptions.RequiredSourcePaths = bindingView.TargetFiles
		viewOptions.SourceMappingsCanonical = true
		resolver, err := newLintConfigResolver(viewOptions)
		if err != nil {
			return nil, nil, err
		}
		targetPaths := bindingView.TargetPathBySourcePath
		targetPathForFile := func(sourceFile *ast.SourceFile) string {
			if sourceFile == nil {
				return ""
			}
			sourcePath := sourceFile.FileName()
			if targetPath := targetPaths[compilerPathID(sourcePath)]; targetPath != "" {
				return targetPath
			}
			return sourcePath
		}
		viewDiagnostic := onDiagnostic
		if onDiagnostic != nil {
			viewDiagnostic = func(d rule.RuleDiagnostic) {
				if sourceFile, ok := d.SourceFile.(*ast.SourceFile); ok {
					d.FilePath = targetPathForFile(sourceFile)
				}
				onDiagnostic(d)
			}
		}
		views = append(views, linter.LintProgramView{
			Program:     binding.Programs[bindingView.ProgramIndex],
			TargetFiles: bindingView.TargetFiles,
			GetRulesForFile: func(sourceFile *ast.SourceFile) []linter.ConfiguredRule {
				return resolver.ActiveRulesForFile(sourceFile.FileName())
			},
			TypeInfoFiles:     bindingView.TypeInfoFiles,
			OnDiagnostic:      viewDiagnostic,
			TargetPathForFile: targetPathForFile,
		})
		resolvers = append(resolvers, resolver)
	}
	return views, resolvers, nil
}
