// Package lint resolves already-selected lint targets against their governing
// configuration. It resolves frozen targets before Program construction and
// joins the same configuration to source paths after Program binding.
package lint

import (
	"fmt"
	"sort"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	"github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/config/target"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// Resolver maps Program source paths back to their selected lint targets and
// resolves the effective configuration owned by those targets.
type Resolver struct {
	configsByOwner       map[string]config.RslintConfig
	config               config.RslintConfig
	configDirectory      string
	defaultRootDirectory string
	targetsBySourcePath  map[string]target.File
	fsys                 vfs.FS
	singleResolver       *config.FileConfigResolver
	resolversByOwnerPath map[string]*config.FileConfigResolver
}

// ResolverOptions contains one request's frozen configuration and source-path
// binding. Catalog and PathSpaces must belong to the same config generation.
type ResolverOptions struct {
	ConfigsByOwner  map[string]config.RslintConfig
	Config          config.RslintConfig
	ConfigDirectory string
	// DefaultRootDirectory is the loaded config module directory, or the
	// invocation cwd for an inline-only Config. It must not be inferred from
	// ConfigDirectory, which may instead be a synthetic matching/path base.
	// ConfigsByOwner supplies each module's own directory and ignores this field.
	DefaultRootDirectory string
	Catalog              *rule.Catalog
	TargetsBySourcePath  map[string]target.File
	// SourceMappingsIncludeCanonicalPaths indicates that Program binding
	// already supplied both lexical and canonical source keys, so normalization
	// needs no filesystem IO.
	SourceMappingsIncludeCanonicalPaths bool
	PathSpaces                          *config.PathSpaceSnapshot
	FS                                  vfs.FS
}

// NewResolver creates a request-scoped source/config resolver.
func NewResolver(options ResolverOptions) *Resolver {
	if options.Catalog == nil {
		panic("rule catalog is required")
	}
	if options.PathSpaces == nil {
		panic("path-space snapshot is required")
	}
	resolver := &Resolver{
		configsByOwner:       options.ConfigsByOwner,
		config:               options.Config,
		configDirectory:      options.ConfigDirectory,
		defaultRootDirectory: options.DefaultRootDirectory,
		targetsBySourcePath: normalizeSourceTargetMappings(
			options.TargetsBySourcePath,
			options.FS,
			options.SourceMappingsIncludeCanonicalPaths,
		),
		fsys: options.FS,
	}
	newFileResolver := func(entries config.RslintConfig, configDirectory string) *config.FileConfigResolver {
		fileResolver, err := config.NewFileConfigResolverWithPathSpaces(
			entries,
			configDirectory,
			options.FS,
			options.PathSpaces,
			options.Catalog,
		)
		if err != nil {
			panic(err)
		}
		return fileResolver
	}
	if options.ConfigsByOwner == nil {
		resolver.singleResolver = newFileResolver(options.Config, options.ConfigDirectory)
		return resolver
	}
	resolver.resolversByOwnerPath = make(map[string]*config.FileConfigResolver, len(options.ConfigsByOwner)*2)
	ownerDirectories := make([]string, 0, len(options.ConfigsByOwner))
	for ownerDirectory := range options.ConfigsByOwner {
		ownerDirectories = append(ownerDirectories, ownerDirectory)
	}
	sort.Strings(ownerDirectories)
	for _, ownerDirectory := range ownerDirectories {
		fileResolver := newFileResolver(options.ConfigsByOwner[ownerDirectory], ownerDirectory)
		resolver.resolversByOwnerPath[ownerDirectory] = fileResolver
		physicalDirectory, _ := options.PathSpaces.PhysicalDirectory(ownerDirectory)
		canonicalOwner := config.ExactPathID(physicalDirectory)
		if _, isLiteralOwner := options.ConfigsByOwner[canonicalOwner]; !isLiteralOwner {
			resolver.resolversByOwnerPath[canonicalOwner] = fileResolver
		}
	}
	return resolver
}

// WithSourceMappings adds one Program generation's source binding without
// rebuilding configuration resolvers or changing their frozen path spaces.
func (resolver *Resolver) WithSourceMappings(
	mapping map[string]target.File,
	fsys vfs.FS,
	canonicalKeysPresent bool,
) *Resolver {
	bound := *resolver
	bound.targetsBySourcePath = normalizeSourceTargetMappings(mapping, fsys, canonicalKeysPresent)
	bound.fsys = fsys
	return &bound
}

// ResolveTarget evaluates an already-selected target under its frozen owner.
// The boolean reports owner availability; a nil merged config is a valid miss.
func (resolver *Resolver) ResolveTarget(file target.File) (config.ResolvedFileConfig, bool) {
	if resolver.singleResolver != nil {
		return resolver.singleResolver.ResolveTarget(file.Identity()), true
	}
	fileResolver := resolver.resolversByOwnerPath[file.ConfigDirectory]
	if fileResolver == nil {
		return config.ResolvedFileConfig{}, false
	}
	return fileResolver.ResolveTarget(file.Identity()), true
}

// ProjectPolicies resolves only owners with service/root/reset options. The
// returned values carry no project lists, and zero policies need no override.
func (resolver *Resolver) ProjectPolicies(files []target.File) (map[target.File]config.ProjectPolicy, error) {
	type ownerProjectContext struct {
		hasOptions    bool
		rootDirectory string
	}
	ownerContexts := make(map[*config.FileConfigResolver]ownerProjectContext, len(resolver.configsByOwner))
	for owner, entries := range resolver.configsByOwner {
		ownerContexts[resolver.resolversByOwnerPath[owner]] = ownerProjectContext{
			hasOptions: config.HasProjectOptions(entries), rootDirectory: owner,
		}
	}
	singleHasOptions := config.HasProjectOptions(resolver.config)
	var policies map[target.File]config.ProjectPolicy
	for _, file := range files {
		hasOptions := singleHasOptions
		defaultRootDirectory := resolver.defaultRootDirectory
		if resolver.configsByOwner != nil {
			context := ownerContexts[resolver.resolversByOwnerPath[file.ConfigDirectory]]
			hasOptions = context.hasOptions
			defaultRootDirectory = context.rootDirectory
		}
		if !hasOptions {
			continue
		}
		resolved, ok := resolver.ResolveTarget(file)
		if !ok {
			continue
		}
		policy, err := config.ResolveProjectPolicy(resolved, defaultRootDirectory)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", file.Path, err)
		}
		if policy == (config.ProjectPolicy{}) {
			continue
		}
		if policies == nil {
			policies = make(map[target.File]config.ProjectPolicy)
		}
		policies[file] = policy
	}
	return policies, nil
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
	for sourcePath, lintTarget := range mapping {
		lintTarget.Path = tspath.NormalizePath(lintTarget.Path)
		lintTarget.CanonicalPath = tspath.NormalizePath(lintTarget.CanonicalPath)
		lintTarget.CanonicalParentPath = tspath.NormalizePath(lintTarget.CanonicalParentPath)
		lintTarget.ConfigDirectory = tspath.NormalizePath(lintTarget.ConfigDirectory)
		normalizedPath := config.ExactPathID(sourcePath)
		normalized[normalizedPath] = lintTarget
		if !canonicalKeysPresent {
			normalized[canonicalPathID(normalizedPath, fsys)] = lintTarget
		}
	}
	return normalized
}

func authoritativePath(filePath string, fsys vfs.FS) string {
	filePath = tspath.NormalizePath(filePath)
	if fsys != nil {
		if realPath := fsys.Realpath(filePath); realPath != "" {
			return tspath.NormalizePath(realPath)
		}
	}
	return filePath
}

func canonicalPathID(filePath string, fsys vfs.FS) string {
	return config.ExactPathID(authoritativePath(filePath, fsys))
}

// TargetForSourcePath returns the selected lint target represented by a
// Program source path. It never infers a new owner from the source path.
func (resolver *Resolver) TargetForSourcePath(sourcePath string) (target.File, bool) {
	if resolver == nil {
		return target.File{}, false
	}
	return target.LookupSourceTarget(resolver.targetsBySourcePath, sourcePath, resolver.fsys)
}

// ResolveSourcePath returns the target's governing owner and effective config.
// An unbound source is rejected in multi-config mode and uses the invocation-
// wide config in single-config mode.
func (resolver *Resolver) ResolveSourcePath(
	sourcePath string,
) (string, config.ResolvedFileConfig, bool) {
	lintTarget, bound := resolver.TargetForSourcePath(sourcePath)
	if !bound {
		lintTarget.Path = tspath.NormalizePath(sourcePath)
	}
	if resolver.configsByOwner != nil {
		if bound {
			resolved, ok := resolver.ResolveTarget(lintTarget)
			return lintTarget.ConfigDirectory, resolved, ok
		}
		return "", config.ResolvedFileConfig{}, false
	}
	return resolver.configDirectory, resolver.singleResolver.ResolveTarget(lintTarget.Identity()), true
}

// EnabledRulesForSourcePath returns the complete configured rule set. Program
// capability filtering remains the lint planner's responsibility.
func (resolver *Resolver) EnabledRulesForSourcePath(sourcePath string) []rule.ConfiguredRule {
	_, resolved, ok := resolver.ResolveSourcePath(sourcePath)
	if !ok {
		return nil
	}
	return resolved.EnabledRules
}
