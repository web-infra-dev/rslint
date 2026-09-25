// Package lint resolves already-selected lint targets against their governing
// configuration. It resolves frozen targets before Program construction and
// joins the same configuration to source paths after Program binding.
package lint

import (
	"errors"
	"fmt"
	"runtime"
	"sort"

	"github.com/microsoft/TypeScript/tsc/shim/core"
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
	literalOwners := make(map[string]struct{}, len(options.ConfigsByOwner))
	for ownerDirectory := range options.ConfigsByOwner {
		ownerDirectories = append(ownerDirectories, ownerDirectory)
		literalOwners[config.ExactPathID(ownerDirectory)] = struct{}{}
	}
	sort.Strings(ownerDirectories)
	for _, ownerDirectory := range ownerDirectories {
		fileResolver := newFileResolver(options.ConfigsByOwner[ownerDirectory], ownerDirectory)
		resolver.resolversByOwnerPath[config.ExactPathID(ownerDirectory)] = fileResolver
		physicalDirectory, _ := options.PathSpaces.PhysicalDirectory(ownerDirectory)
		canonicalOwner := config.ExactPathID(physicalDirectory)
		if _, isLiteralOwner := literalOwners[canonicalOwner]; !isLiteralOwner {
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
		return resolver.singleResolver.ResolveTargetWithMatch(file.Identity(), file.Match()), true
	}
	fileResolver := resolver.resolversByOwnerPath[config.ExactPathID(file.ConfigDirectory)]
	if fileResolver == nil {
		return config.ResolvedFileConfig{}, false
	}
	return fileResolver.ResolveTargetWithMatch(file.Identity(), file.Match()), true
}

// ProjectPolicies projects the same final config used by lint rules. Every
// target has an entry: a zero policy requests no project and must not inherit
// another target's declarations during Program binding. File configuration is
// resolved concurrently unless singleThreaded is set; policy projection and
// errors retain the input order.
func (resolver *Resolver) ProjectPolicies(files []target.File, singleThreaded bool) (map[target.File]config.ProjectPolicy, error) {
	ownerRoots := make(map[*config.FileConfigResolver]string, len(resolver.configsByOwner))
	for owner := range resolver.configsByOwner {
		ownerRoots[resolver.resolversByOwnerPath[config.ExactPathID(owner)]] = owner
	}
	policies := make(map[target.File]config.ProjectPolicy, len(files))
	resolvedPolicies := make(map[*config.MergedConfig]config.ProjectPolicy)
	var configs []projectTargetConfig
	if workers := min(runtime.GOMAXPROCS(0), len(files)); !singleThreaded && workers > 1 {
		configs = resolver.resolveProjectTargetConfigs(files, workers)
	}
	for index, file := range files {
		defaultRootDirectory := resolver.defaultRootDirectory
		if resolver.configsByOwner != nil {
			defaultRootDirectory = ownerRoots[resolver.resolversByOwnerPath[config.ExactPathID(file.ConfigDirectory)]]
		}
		var resolved config.ResolvedFileConfig
		var ok bool
		if configs == nil {
			resolved, ok = resolver.ResolveTarget(file)
		} else {
			if configs[index].panicValue != nil {
				panic(configs[index].panicValue)
			}
			resolved, ok = configs[index].resolved, configs[index].available
		}
		if !ok {
			return nil, fmt.Errorf("%s: missing governing configuration", file.Path)
		}
		policy, cached := resolvedPolicies[resolved.MergedConfig]
		if !cached {
			var err error
			policy, err = config.ResolveProjectPolicy(resolved, defaultRootDirectory)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", file.Path, err)
			}
			resolvedPolicies[resolved.MergedConfig] = policy
		}
		policies[file] = policy
	}
	return policies, nil
}

type projectTargetConfig struct {
	resolved   config.ResolvedFileConfig
	available  bool
	panicValue any
}

func (resolver *Resolver) resolveProjectTargetConfigs(files []target.File, workers int) []projectTargetConfig {
	configs := make([]projectTargetConfig, len(files))
	chunkSize := (len(files) + workers - 1) / workers
	work := core.NewWorkGroup(false)
	for start := 0; start < len(files); start += chunkSize {
		end := min(start+chunkSize, len(files))
		work.Queue(func() {
			index := start
			completed := false
			defer func() {
				if completed {
					return
				}
				// Replay failures on the caller only when it reaches this file,
				// so a later panic cannot hide an earlier configuration error.
				value := recover()
				if value == nil {
					value = errors.New("file configuration worker exited without a panic value")
				}
				configs[index].panicValue = value
			}()
			for ; index < end; index++ {
				configs[index].resolved, configs[index].available = resolver.ResolveTarget(files[index])
			}
			completed = true
		})
	}
	work.RunAndWait()
	return configs
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
	resolved, ok := resolver.ResolveTarget(lintTarget)
	return resolver.configDirectory, resolved, ok
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
