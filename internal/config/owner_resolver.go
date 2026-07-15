package config

import (
	"sort"
	"strings"

	"github.com/microsoft/typescript-go/shim/vfs"
	"github.com/web-infra-dev/rslint/internal/hostpath"
)

type configDirectoryIndex struct {
	fsys                     vfs.FS
	configKeyByPath          map[string]string
	caseFoldedConfigKeys     map[string][]string
	canonicalConfigKeyByPath map[string]string
	ambiguousCanonicalPaths  map[string]struct{}
	normalizedByKey          map[string]string
	canonicalByKey           map[string]string
	childrenByKey            map[string][]string
}

func newConfigDirectoryIndex(configMap map[string]RslintConfig, fsys vfs.FS) *configDirectoryIndex {
	index := &configDirectoryIndex{
		fsys:                     fsys,
		configKeyByPath:          make(map[string]string, len(configMap)),
		caseFoldedConfigKeys:     make(map[string][]string, len(configMap)),
		canonicalConfigKeyByPath: make(map[string]string, len(configMap)),
		ambiguousCanonicalPaths:  make(map[string]struct{}),
		normalizedByKey:          make(map[string]string, len(configMap)),
		canonicalByKey:           make(map[string]string, len(configMap)),
		childrenByKey:            make(map[string][]string, len(configMap)),
	}
	configKeys := make([]string, 0, len(configMap))
	for configKey := range configMap {
		configKeys = append(configKeys, configKey)
	}
	sort.Strings(configKeys)
	for _, configKey := range configKeys {
		normalized := normalizePathForRoot(configKey, configKey)
		index.normalizedByKey[configKey] = normalized
		pathID := ownerPathIdentity(normalized, true)
		if _, exists := index.configKeyByPath[pathID]; !exists {
			index.configKeyByPath[pathID] = configKey
		}
		foldedPathID := ownerPathIdentity(normalized, false)
		index.caseFoldedConfigKeys[foldedPathID] = append(index.caseFoldedConfigKeys[foldedPathID], configKey)

		canonical := normalized
		if fsys != nil {
			if realPath := fsys.Realpath(normalized); realPath != "" {
				canonical = normalizePathForRoot(normalized, realPath)
			}
		}
		index.canonicalByKey[configKey] = canonical
		canonicalID := ownerPathIdentity(canonical, true)
		if _, ambiguous := index.ambiguousCanonicalPaths[canonicalID]; ambiguous {
			continue
		}
		if existing, exists := index.canonicalConfigKeyByPath[canonicalID]; !exists {
			index.canonicalConfigKeyByPath[canonicalID] = configKey
		} else if existing != configKey {
			// Lexical aliases remain independently addressable. A physical-path
			// fallback cannot choose between them, so leave it unresolved instead
			// of silently assigning the file to the first map entry.
			delete(index.canonicalConfigKeyByPath, canonicalID)
			index.ambiguousCanonicalPaths[canonicalID] = struct{}{}
		}
	}

	for _, configKey := range configKeys {
		normalized := index.normalizedByKey[configKey]
		if parentKey, ok := index.nearestLexicalConfigAncestor(normalized); ok {
			index.addChildBoundary(parentKey, normalized)
		}
	}
	for configKey := range index.childrenByKey {
		sort.Strings(index.childrenByKey[configKey])
	}
	return index
}

func (index *configDirectoryIndex) nearestLexicalConfigAncestor(configDir string) (string, bool) {
	current := directoryPathForRoot(configDir, configDir)
	for current != "" && current != configDir {
		if configKey, ok := index.configKeyForLexicalDirectory(current); ok {
			return configKey, true
		}
		next := directoryPathForRoot(configDir, current)
		if next == current {
			break
		}
		current = next
	}
	return "", false
}

func (index *configDirectoryIndex) addChildBoundary(configKey string, boundary string) {
	boundary = normalizePathForRoot(boundary, boundary)
	for _, existing := range index.childrenByKey[configKey] {
		if existing == boundary {
			return
		}
	}
	index.childrenByKey[configKey] = append(index.childrenByKey[configKey], boundary)
}

func (index *configDirectoryIndex) childConfigDirs(configKey string) []string {
	if index == nil {
		return nil
	}
	return index.childrenByKey[configKey]
}

func (index *configDirectoryIndex) nearestConfig(filePath string) (string, bool) {
	if index == nil {
		return "", false
	}
	filePath = normalizePathForRoot(filePath, filePath)
	if configKey, ok := index.nearestLexicalConfig(filePath); ok {
		return configKey, true
	}
	if index.fsys == nil {
		return "", false
	}
	realPath := index.fsys.Realpath(filePath)
	if realPath == "" {
		return "", false
	}
	return index.nearestConfigInPathSpace(
		normalizePathForRoot(filePath, realPath),
		index.canonicalConfigKeyByPath,
	)
}

func (index *configDirectoryIndex) nearestLexicalConfig(filePath string) (string, bool) {
	if index == nil {
		return "", false
	}
	filePath = normalizePathForRoot(filePath, filePath)
	current := directoryPathForRoot(filePath, filePath)
	for current != "" {
		if configKey, ok := index.configKeyForLexicalDirectory(current); ok {
			return configKey, true
		}
		next := directoryPathForRoot(filePath, current)
		if next == current {
			break
		}
		current = next
	}
	return "", false
}

func (index *configDirectoryIndex) configKeyForLexicalDirectory(directory string) (string, bool) {
	if index == nil {
		return "", false
	}
	if configKey, ok := index.configKeyByPath[ownerPathIdentity(directory, true)]; ok {
		return configKey, true
	}
	if index.fsys == nil {
		return "", false
	}
	candidates := index.caseFoldedConfigKeys[ownerPathIdentity(directory, false)]
	if len(candidates) == 0 {
		return "", false
	}
	canonicalDirectory := index.fsys.Realpath(directory)
	if canonicalDirectory == "" {
		return "", false
	}
	canonicalDirectory = normalizePathForRoot(directory, canonicalDirectory)
	for _, configKey := range candidates {
		if pathsEqualForRoot(canonicalDirectory, canonicalDirectory, index.canonicalByKey[configKey], true) {
			return configKey, true
		}
	}
	return "", false
}

func (index *configDirectoryIndex) nearestConfigInPathSpace(
	filePath string,
	configKeyByPath map[string]string,
) (string, bool) {
	current := directoryPathForRoot(filePath, filePath)
	for current != "" {
		if configKey, ok := configKeyByPath[ownerPathIdentity(current, true)]; ok {
			return configKey, true
		}
		next := directoryPathForRoot(filePath, current)
		if next == current {
			break
		}
		current = next
	}
	return "", false
}

func ownerPathIdentity(path string, caseSensitive bool) string {
	path = normalizePathForRoot(path, path)
	if !caseSensitive {
		path = strings.ToLower(path)
	}
	return path
}

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

// ResolveConfigPathSpace returns the authored lexical path pair used for files
// and ignores matching. Canonical identity is consulted only to recover the
// target's relative location when a Program reports a physical path; the final
// pair is projected back under configDir so absolute basePath values, aliases,
// and diagnostics all observe the same path space ESLint received.
func ResolveConfigPathSpace(filePath string, configDir string, fsys vfs.FS) (string, string) {
	return ResolveConfigPathSpaceWithCanonical(filePath, "", configDir, fsys)
}

// ResolveConfigPathSpaceWithCanonical is ResolveConfigPathSpace with an
// optional physical file identity already established by target discovery.
func ResolveConfigPathSpaceWithCanonical(filePath string, canonicalPath string, configDir string, fsys vfs.FS) (string, string) {
	governingRoot := configDir
	filePath = normalizePathForRoot(governingRoot, filePath)
	configDir = normalizePathForRoot(governingRoot, configDir)
	physicalConfigDir := configDir
	if fsys != nil {
		if realPath := fsys.Realpath(configDir); realPath != "" {
			physicalConfigDir = normalizePathForRoot(configDir, realPath)
		}
	}

	physicalMatchPath := resolveConfigPathSpace(filePath, canonicalPath, configDir, physicalConfigDir, fsys)
	if relative, within := RelativePathWithinConfigRoot(
		physicalMatchPath,
		physicalConfigDir,
		selectorScopeCaseSensitive(physicalConfigDir),
	); within {
		return resolvePathForRoot(configDir, configDir, relative), configDir
	}
	return filePath, configDir
}

func resolveConfigPathSpace(
	filePath string,
	canonicalPath string,
	configDir string,
	physicalConfigDir string,
	fsys vfs.FS,
) string {
	if relative, ok := RelativePathWithinConfigRoot(filePath, configDir, selectorScopeCaseSensitive(configDir)); ok {
		return resolvePathForRoot(physicalConfigDir, physicalConfigDir, relative)
	}
	if relative, ok := RelativePathWithinConfigRoot(filePath, configDir, false); ok && fsys != nil {
		aliasRoot := filePath
		for remaining := relative; remaining != ""; remaining = directoryPathForRoot(configDir, remaining) {
			aliasRoot = directoryPathForRoot(configDir, aliasRoot)
		}
		if realRoot := fsys.Realpath(aliasRoot); realRoot != "" &&
			pathsEqualForRoot(physicalConfigDir, realRoot, physicalConfigDir, true) {
			return resolvePathForRoot(physicalConfigDir, physicalConfigDir, relative)
		}
	}

	physicalFilePath := ""
	if canonicalPath != "" {
		physicalFilePath = normalizePathForRoot(physicalConfigDir, canonicalPath)
	}
	if physicalFilePath == "" {
		physicalFilePath = filePath
		if fsys != nil {
			if realPath := fsys.Realpath(filePath); realPath != "" {
				physicalFilePath = normalizePathForRoot(physicalConfigDir, realPath)
			}
		}
	}
	if relative, ok := RelativePathWithinConfigRoot(
		physicalFilePath,
		physicalConfigDir,
		selectorScopeCaseSensitive(physicalConfigDir),
	); ok {
		return resolvePathForRoot(physicalConfigDir, physicalConfigDir, relative)
	}
	return physicalFilePath
}

// RelativePathWithinConfigRoot returns filePath relative to configDir when it
// is inside the config's lexical path space.
func RelativePathWithinConfigRoot(filePath string, configDir string, useCaseSensitive bool) (string, bool) {
	return hostpath.RelativeWithin(filePath, configDir, useCaseSensitive)
}
