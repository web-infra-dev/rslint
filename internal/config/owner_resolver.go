package config

import (
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
)

// ConfigOwnerResolver snapshots an already-loaded config catalog and resolves
// which config object governs a runtime file path. It never discovers, reads,
// or parses config files. Construction is linear in config count. Each lookup
// tries lexical ancestors, including verified native case aliases, and consults
// realpath ancestry only when no lexical owner exists.
type ConfigOwnerResolver struct {
	configMap map[string]RslintConfig
	index     *configDirectoryIndex
}

func NewConfigOwnerResolver(configMap map[string]RslintConfig, fsys vfs.FS) *ConfigOwnerResolver {
	configSnapshot := make(map[string]RslintConfig, len(configMap))
	for configDir, entries := range configMap {
		configSnapshot[configDir] = entries
	}
	return &ConfigOwnerResolver{
		configMap: configSnapshot,
		index:     newConfigDirectoryIndex(configSnapshot, fsys),
	}
}

func (resolver *ConfigOwnerResolver) Resolve(filePath string) (string, RslintConfig) {
	if resolver == nil || resolver.index == nil {
		return "", nil
	}
	configDir, ok := resolver.index.nearestConfig(filePath)
	if !ok {
		return "", nil
	}
	return configDir, resolver.configMap[configDir]
}

// ChildConfigDirs returns the direct lexical child config directories that
// form ownership handoff boundaries for configDir. The returned slice is a
// copy and may be used concurrently with resolver lookups.
func (resolver *ConfigOwnerResolver) ChildConfigDirs(configDir string) []string {
	if resolver == nil || resolver.index == nil {
		return nil
	}
	return append([]string(nil), resolver.index.childConfigDirs(configDir)...)
}

// ResolveConfigPathSpace returns the physical path pair used for files and
// ignores matching. It preserves the file's path relative to the authored
// config directory, then anchors both paths on the config directory's realpath.
// Lexical and symlink aliases therefore share one matching space without
// case-folding distinct path identities.
func ResolveConfigPathSpace(filePath string, configDir string, fsys vfs.FS) (string, string) {
	return ResolveConfigPathSpaceWithCanonical(filePath, "", configDir, fsys)
}

// ResolveConfigPathSpaceWithCanonical is ResolveConfigPathSpace with an
// optional physical file identity already established by target discovery.
func ResolveConfigPathSpaceWithCanonical(filePath string, canonicalPath string, configDir string, fsys vfs.FS) (string, string) {
	filePath = tspath.NormalizePath(filePath)
	configDir = tspath.NormalizePath(configDir)
	physicalConfigDir := configDir
	if fsys != nil {
		if realPath := fsys.Realpath(configDir); realPath != "" {
			physicalConfigDir = tspath.NormalizePath(realPath)
		}
	}

	return resolveConfigPathSpace(filePath, canonicalPath, configDir, physicalConfigDir, fsys), physicalConfigDir
}

func resolveConfigPathSpace(
	filePath string,
	canonicalPath string,
	configDir string,
	physicalConfigDir string,
	fsys vfs.FS,
) string {
	if relative, ok := RelativePathWithinConfigRoot(filePath, configDir, true); ok {
		return tspath.ResolvePath(physicalConfigDir, relative)
	}
	if relative, ok := RelativePathWithinConfigRoot(filePath, configDir, false); ok && fsys != nil {
		aliasRoot := filePath
		for remaining := relative; remaining != ""; remaining = tspath.GetDirectoryPath(remaining) {
			aliasRoot = tspath.GetDirectoryPath(aliasRoot)
		}
		if realRoot := fsys.Realpath(aliasRoot); realRoot != "" &&
			tspath.ComparePaths(
				tspath.NormalizePath(realRoot),
				physicalConfigDir,
				tspath.ComparePathsOptions{UseCaseSensitiveFileNames: true},
			) == 0 {
			return tspath.ResolvePath(physicalConfigDir, relative)
		}
	}

	physicalFilePath := ""
	if canonicalPath != "" {
		physicalFilePath = tspath.NormalizePath(canonicalPath)
	}
	if physicalFilePath == "" {
		physicalFilePath = filePath
		if fsys != nil {
			if realPath := fsys.Realpath(filePath); realPath != "" {
				physicalFilePath = tspath.NormalizePath(realPath)
			}
		}
	}
	if relative, ok := RelativePathWithinConfigRoot(physicalFilePath, physicalConfigDir, true); ok {
		return tspath.ResolvePath(physicalConfigDir, relative)
	}
	return physicalFilePath
}

// RelativePathWithinConfigRoot returns filePath relative to configDir when it
// is inside the config's lexical path space.
func RelativePathWithinConfigRoot(filePath string, configDir string, useCaseSensitive bool) (string, bool) {
	options := tspath.ComparePathsOptions{
		CurrentDirectory:          configDir,
		UseCaseSensitiveFileNames: useCaseSensitive,
	}
	if tspath.ComparePaths(filePath, configDir, options) == 0 {
		return "", true
	}
	if !tspath.StartsWithDirectory(filePath, configDir, useCaseSensitive) {
		return "", false
	}
	return tspath.GetRelativePathFromDirectory(configDir, filePath, options), true
}
