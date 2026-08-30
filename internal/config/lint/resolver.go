// Package lint resolves already-selected lint targets against their governing
// configuration. Target discovery and Program binding happen before this
// package; it only joins their immutable results to config.FileConfigResolver.
package lint

import (
	"sort"

	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	"github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/config/target"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// Resolver maps Program source paths back to their selected lint targets and
// resolves the effective configuration owned by those targets.
type Resolver struct {
	configsByOwner       map[string]config.RslintConfig
	configDirectory      string
	targetsBySourcePath  map[string]target.File
	fsys                 vfs.FS
	singleResolver       *config.FileConfigResolver
	resolversByOwnerPath map[string]*config.FileConfigResolver
}

// ResolverOptions contains one request's frozen configuration and source-path
// binding. Catalog and PathSpaces must belong to the same config generation.
type ResolverOptions struct {
	ConfigsByOwner      map[string]config.RslintConfig
	Config              config.RslintConfig
	ConfigDirectory     string
	Catalog             *rule.Catalog
	TargetsBySourcePath map[string]target.File
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
		configsByOwner:  options.ConfigsByOwner,
		configDirectory: options.ConfigDirectory,
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
		resolver.resolversByOwnerPath[canonicalPathID(ownerDirectory, options.FS)] = fileResolver
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

func (resolver *Resolver) resolverForOwner(ownerDirectory string) *config.FileConfigResolver {
	if fileResolver := resolver.resolversByOwnerPath[ownerDirectory]; fileResolver != nil {
		return fileResolver
	}
	return resolver.resolversByOwnerPath[canonicalPathID(ownerDirectory, resolver.fsys)]
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
			fileResolver := resolver.resolverForOwner(lintTarget.ConfigDirectory)
			if fileResolver == nil {
				return "", config.ResolvedFileConfig{}, false
			}
			return lintTarget.ConfigDirectory, fileResolver.ResolveTarget(lintTarget.Identity()), true
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
