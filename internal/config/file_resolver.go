package config

import (
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type effectiveConfigPlan struct {
	mergedConfig *MergedConfig
	enabledRules []rule.ConfiguredRule
}

// ResolvedFileConfig is the immutable result of evaluating one frozen lint
// target against a flat config. A nil MergedConfig means no config entry
// selected the target; GloballyIgnored distinguishes the case where linting
// must be skipped entirely from the case where syntax diagnostics still apply.
type ResolvedFileConfig struct {
	MergedConfig    *MergedConfig
	EnabledRules    []rule.ConfiguredRule
	GloballyIgnored bool
}

type configTargetResolution struct {
	plan            *effectiveConfigPlan
	globallyIgnored bool
}

type configTargetCacheKey struct {
	path                string
	canonicalPath       string
	canonicalParentPath string
}

// FileConfigResolver resolves config and rules for one immutable config root.
// It interns files with the same exact matched-entry shape for the resolver's
// lifetime and is safe for concurrent use by native and plugin lint workers.
type FileConfigResolver struct {
	config         RslintConfig
	catalog        *rule.Catalog
	enforcePlugins bool
	targetResolver *configTargetResolver

	filePlans  publishOnceCache[configTargetCacheKey, *configTargetResolution]
	shapePlans publishOnceCache[configMatchKey, *effectiveConfigPlan]
}

// NewFileConfigResolver creates a per-run resolver for one config root.
func NewFileConfigResolver(
	config RslintConfig,
	cwd string,
	catalog *rule.Catalog,
	enforcePlugins bool,
) *FileConfigResolver {
	return NewFileConfigResolverWithFS(config, cwd, nil, catalog, enforcePlugins)
}

// NewFileConfigResolverWithFS creates a resolver that keeps lexical and
// canonical target identity together while matching authored path bases.
func NewFileConfigResolverWithFS(
	config RslintConfig,
	cwd string,
	fsys vfs.FS,
	catalog *rule.Catalog,
	enforcePlugins bool,
) *FileConfigResolver {
	return newFileConfigResolver(
		config,
		catalog,
		enforcePlugins,
		newConfigTargetResolver(config, cwd, fsys),
	)
}

func newFileConfigResolver(
	config RslintConfig,
	catalog *rule.Catalog,
	enforcePlugins bool,
	targetResolver *configTargetResolver,
) *FileConfigResolver {
	if catalog == nil {
		panic("rule catalog is required")
	}
	return &FileConfigResolver{
		config:         config,
		catalog:        catalog,
		enforcePlugins: enforcePlugins,
		targetResolver: targetResolver,
	}
}

// ConfigForFile returns the merged config for filePath, caching nil misses.
// The returned value is shared immutable resolver state and must be read-only.
func (r *FileConfigResolver) ConfigForFile(filePath string) *MergedConfig {
	return r.ConfigForTarget(filePath, "")
}

// ConfigForTarget returns the merged config for one stable lint-target
// identity. The resolver projects the lexical and canonical paths into every
// entry's authored path space; callers must not pre-project them to one config
// directory.
func (r *FileConfigResolver) ConfigForTarget(filePath string, canonicalPath string) *MergedConfig {
	plan := r.planForTarget(filePath, canonicalPath)
	if plan == nil {
		return nil
	}
	return plan.mergedConfig
}

// EnabledRulesForFile returns cached enabled rules and their merged config.
// Both returned values are shared immutable resolver state and must be read-only.
func (r *FileConfigResolver) EnabledRulesForFile(filePath string) ([]rule.ConfiguredRule, *MergedConfig) {
	return r.EnabledRulesForTarget(filePath, "")
}

// EnabledRulesForTarget is EnabledRulesForFile with the canonical identity
// frozen during target discovery.
func (r *FileConfigResolver) EnabledRulesForTarget(filePath string, canonicalPath string) ([]rule.ConfiguredRule, *MergedConfig) {
	plan := r.planForTarget(filePath, canonicalPath)
	if plan == nil {
		return nil, nil
	}
	return plan.enabledRules, plan.mergedConfig
}

// ResolveTarget evaluates a complete target identity once. Callers that pass
// a CanonicalParentPath avoid every later filesystem lookup needed to
// distinguish a leaf symlink from an aliased directory tree.
func (r *FileConfigResolver) ResolveTarget(
	target DiscoveredLintTarget,
) ResolvedFileConfig {
	resolution := r.resolutionForTarget(target)
	if resolution == nil {
		return ResolvedFileConfig{}
	}
	result := ResolvedFileConfig{GloballyIgnored: resolution.globallyIgnored}
	if resolution.plan != nil {
		result.MergedConfig = resolution.plan.mergedConfig
		result.EnabledRules = resolution.plan.enabledRules
	}
	return result
}

func (r *FileConfigResolver) planForFile(filePath string) *effectiveConfigPlan {
	return r.planForTarget(filePath, "")
}

func (r *FileConfigResolver) planForTarget(filePath string, canonicalPath string) *effectiveConfigPlan {
	resolution := r.resolutionForTarget(DiscoveredLintTarget{
		Path:          filePath,
		CanonicalPath: canonicalPath,
	})
	if resolution == nil {
		return nil
	}
	return resolution.plan
}

func (r *FileConfigResolver) resolutionForTarget(
	target DiscoveredLintTarget,
) *configTargetResolution {
	key := configTargetCacheKey{
		path:                tspath.NormalizePath(target.Path),
		canonicalPath:       tspath.NormalizePath(target.CanonicalPath),
		canonicalParentPath: tspath.NormalizePath(target.CanonicalParentPath),
	}
	return r.filePlans.getOrInit(key, func() *configTargetResolution {
		decision := r.targetResolver.resolveTarget(DiscoveredLintTarget{
			Path:                key.path,
			CanonicalPath:       key.canonicalPath,
			CanonicalParentPath: key.canonicalParentPath,
		})
		resolution := &configTargetResolution{
			globallyIgnored: decision.globallyIgnored,
		}
		if !decision.matched || !decision.selected || decision.globallyIgnored {
			return resolution
		}

		resolution.plan = r.shapePlans.getOrInit(decision.key, func() *effectiveConfigPlan {
			mergedConfig := r.config.mergeConfigEntries(decision.key)
			return &effectiveConfigPlan{
				mergedConfig: mergedConfig,
				enabledRules: ConfiguredRules(r.catalog, mergedConfig, r.enforcePlugins),
			}
		})
		return resolution
	})
}
