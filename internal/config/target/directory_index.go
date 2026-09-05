package target

import (
	"sort"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
)

type configDirectoryIndex struct {
	configKeyByPath          map[tspath.Path]string
	caseFoldedConfigKeys     map[tspath.Path][]string
	canonicalConfigKeyByPath map[tspath.Path]string
	ambiguousCanonicalPaths  map[tspath.Path]struct{}
	normalizedByKey          map[string]string
	canonicalByKey           map[string]string
	childrenByKey            map[string][]string
}

func newConfigDirectoryIndexWithPathSpaces(
	configMap map[string]rslintconfig.RslintConfig,
	fsys vfs.FS,
	pathSpaces *rslintconfig.PathSpaceSnapshot,
) *configDirectoryIndex {
	index := &configDirectoryIndex{
		configKeyByPath:          make(map[tspath.Path]string, len(configMap)),
		caseFoldedConfigKeys:     make(map[tspath.Path][]string, len(configMap)),
		canonicalConfigKeyByPath: make(map[tspath.Path]string, len(configMap)),
		ambiguousCanonicalPaths:  make(map[tspath.Path]struct{}),
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
		normalized := tspath.NormalizePath(configKey)
		if len(normalized) > tspath.GetRootLength(normalized) {
			normalized = tspath.RemoveTrailingDirectorySeparators(normalized)
		}
		index.normalizedByKey[configKey] = normalized
		pathID := tspath.ToPath(normalized, "", true)
		if _, exists := index.configKeyByPath[pathID]; !exists {
			index.configKeyByPath[pathID] = configKey
		}
		foldedPathID := tspath.ToPath(normalized, "", false)
		index.caseFoldedConfigKeys[foldedPathID] = append(index.caseFoldedConfigKeys[foldedPathID], configKey)

		canonical, frozen := pathSpaces.PhysicalDirectory(normalized)
		if !frozen {
			panic("config owner is missing from path-space snapshot: " + normalized)
		}
		index.canonicalByKey[configKey] = canonical
		canonicalID := tspath.ToPath(canonical, "", true)
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
		if parentKey, ok := index.nearestLexicalConfigAncestor(normalized, fsys); ok {
			index.addChildBoundary(parentKey, normalized)
		}
	}
	for configKey := range index.childrenByKey {
		sort.Strings(index.childrenByKey[configKey])
	}
	return index
}

func (index *configDirectoryIndex) nearestLexicalConfigAncestor(
	configDir string,
	fsys vfs.FS,
) (string, bool) {
	current := tspath.GetDirectoryPath(configDir)
	for current != "" && current != configDir {
		if configKey, ok := index.configKeyForLexicalDirectory(current, fsys); ok {
			return configKey, true
		}
		next := tspath.GetDirectoryPath(current)
		if next == current {
			break
		}
		current = next
	}
	return "", false
}

func (index *configDirectoryIndex) addChildBoundary(configKey string, boundary string) {
	boundary = tspath.NormalizePath(boundary)
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

// nearestConfigForIdentity resolves a frozen target without consulting the
// filesystem. Construction is the only phase that receives one. Exact lexical
// ancestry wins first. A native-case spelling is accepted only when the
// target's frozen physical parent reaches the candidate's frozen physical
// directory at the same ancestor depth; otherwise physical file ancestry is
// the final fallback.
func (index *configDirectoryIndex) nearestConfigForIdentity(
	identity rslintconfig.PathIdentity,
) (string, bool) {
	if index == nil {
		return "", false
	}
	filePath := tspath.NormalizePath(identity.Path)
	canonicalParent := tspath.NormalizePath(identity.CanonicalParentPath)
	if canonicalParent == "" && identity.CanonicalPath != "" {
		canonicalParent = tspath.GetDirectoryPath(tspath.NormalizePath(identity.CanonicalPath))
	}
	for lexicalDirectory := tspath.GetDirectoryPath(filePath); lexicalDirectory != ""; {
		if configKey, ok := index.configKeyByPath[tspath.ToPath(lexicalDirectory, "", true)]; ok {
			return configKey, true
		}
		if canonicalParent != "" {
			candidates := index.caseFoldedConfigKeys[tspath.ToPath(lexicalDirectory, "", false)]
			for _, configKey := range candidates {
				if rslintconfig.ExactPathID(canonicalParent) ==
					rslintconfig.ExactPathID(index.canonicalByKey[configKey]) {
					return configKey, true
				}
			}
		}

		nextLexical := tspath.GetDirectoryPath(lexicalDirectory)
		if nextLexical == lexicalDirectory {
			break
		}
		lexicalDirectory = nextLexical
		if canonicalParent != "" {
			nextCanonical := tspath.GetDirectoryPath(canonicalParent)
			if nextCanonical == canonicalParent {
				canonicalParent = ""
			} else {
				canonicalParent = nextCanonical
			}
		}
	}
	if identity.CanonicalPath == "" {
		return "", false
	}
	canonicalPath := tspath.NormalizePath(identity.CanonicalPath)
	return index.nearestConfigInPathSpace(
		canonicalPath,
		index.canonicalConfigKeyByPath,
	)
}

func (index *configDirectoryIndex) configKeyForLexicalDirectory(
	directory string,
	fsys vfs.FS,
) (string, bool) {
	if index == nil {
		return "", false
	}
	if configKey, ok := index.configKeyByPath[tspath.ToPath(directory, "", true)]; ok {
		return configKey, true
	}
	if fsys == nil {
		return "", false
	}
	candidates := index.caseFoldedConfigKeys[tspath.ToPath(directory, "", false)]
	if len(candidates) == 0 {
		return "", false
	}
	canonicalDirectory := fsys.Realpath(directory)
	if canonicalDirectory == "" {
		return "", false
	}
	canonicalDirectory = tspath.NormalizePath(canonicalDirectory)
	for _, configKey := range candidates {
		if pathsEqual(canonicalDirectory, index.canonicalByKey[configKey], true) {
			return configKey, true
		}
	}
	return "", false
}

func (index *configDirectoryIndex) nearestConfigInPathSpace(
	filePath string,
	configKeyByPath map[tspath.Path]string,
) (string, bool) {
	current := tspath.GetDirectoryPath(filePath)
	for current != "" {
		if configKey, ok := configKeyByPath[tspath.ToPath(current, "", true)]; ok {
			return configKey, true
		}
		next := tspath.GetDirectoryPath(current)
		if next == current {
			break
		}
		current = next
	}
	return "", false
}
