package discovery

import (
	"sort"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
)

// discoveryTargetTrie bounds a directory walk to branches that can govern an
// invocation target. A nil trie means the current subtree is unbounded. A
// recursive node makes its subtree unbounded, while a non-recursive leaf still
// visits that directory so its sources can govern a file directly within it.
type discoveryTargetTrie struct {
	recursive bool
	children  map[tspath.Path]*discoveryTargetTrie
}

func addDiscoveryTarget(
	root *discoveryTargetTrie,
	configDir string,
	targetDirectory string,
	recursive bool,
	useCaseSensitive bool,
) bool {
	relative, within := rslintconfig.RelativePathWithinConfigRoot(
		targetDirectory,
		configDir,
		useCaseSensitive,
	)
	if !within {
		if recursive {
			if _, containsRoot := rslintconfig.RelativePathWithinConfigRoot(
				configDir,
				targetDirectory,
				useCaseSensitive,
			); containsRoot {
				root.recursive = true
				root.children = nil
				return true
			}
		}
		return false
	}

	node := root
	for _, component := range splitDiscoveryPath(relative) {
		if node.recursive {
			return true
		}
		if node.children == nil {
			node.children = make(map[tspath.Path]*discoveryTargetTrie)
		}
		identity := tspath.ToPath(component, "", useCaseSensitive)
		child := node.children[identity]
		if child == nil {
			child = &discoveryTargetTrie{}
			node.children[identity] = child
		}
		node = child
	}
	if recursive {
		node.recursive = true
		node.children = nil
	}
	return true
}

func (coordinator *discoveryCoordinator) normalizedDirectoryRoots() []string {
	raw := coordinator.request.Directories
	if len(raw) == 0 && len(coordinator.request.Files) == 0 && coordinator.request.ImplicitCWD {
		raw = []string{coordinator.request.CWD}
	}
	useCaseSensitive := coordinator.fs.UseCaseSensitiveFileNames()
	byIdentity := make(map[tspath.Path]string, len(raw))
	for _, directory := range raw {
		directory = normalizeDiscoveryPath(directory, coordinator.request.CWD)
		identity := tspath.ToPath(directory, "", useCaseSensitive)
		if current, exists := byIdentity[identity]; !exists || directory < current {
			byIdentity[identity] = directory
		}
	}
	roots := make([]string, 0, len(byIdentity))
	for _, directory := range byIdentity {
		roots = append(roots, directory)
	}
	sort.Strings(roots)
	return roots
}

func (coordinator *discoveryCoordinator) normalizedFiles() []DiscoveryFile {
	files := make([]DiscoveryFile, 0, len(coordinator.request.Files))
	indexByPath := make(map[string]int, len(coordinator.request.Files))
	for _, file := range coordinator.request.Files {
		file.Path = normalizeDiscoveryPath(file.Path, coordinator.request.CWD)
		if file.CanonicalPath != "" {
			file.CanonicalPath = normalizeDiscoveryPath(file.CanonicalPath, coordinator.request.CWD)
		}
		if index, exists := indexByPath[file.Path]; exists {
			files[index].Explicit = files[index].Explicit || file.Explicit
			if files[index].CanonicalPath == "" {
				files[index].CanonicalPath = file.CanonicalPath
			}
			continue
		}
		indexByPath[file.Path] = len(files)
		files = append(files, file)
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	return files
}

// targetAncestorTries maps each directory root to the lexical paths
// that can govern a supplied target. Native API callers have already expanded
// their globs, so configs in every other subtree cannot affect this request and
// must not be evaluated. Directory-only CLI/LSP requests have no files and keep
// the existing unbounded walk by leaving the returned map empty.
func (coordinator *discoveryCoordinator) targetAncestorTries(
	roots []string,
	files []DiscoveryFile,
) map[string]*discoveryTargetTrie {
	tries := make(map[string]*discoveryTargetTrie, len(roots))
	useCaseSensitive := coordinator.fs.UseCaseSensitiveFileNames()
	for _, file := range files {
		for _, root := range roots {
			trie := tries[root]
			if trie == nil {
				trie = &discoveryTargetTrie{}
			}
			if addDiscoveryTarget(trie, root, tspath.GetDirectoryPath(file.Path), false, useCaseSensitive) {
				tries[root] = trie
			}
		}
	}
	return tries
}

func (trie *discoveryTargetTrie) child(name string, useCaseSensitive bool) *discoveryTargetTrie {
	if trie == nil || len(trie.children) == 0 {
		return nil
	}
	return trie.children[tspath.ToPath(name, "", useCaseSensitive)]
}
func splitDiscoveryPath(path string) []string {
	path = strings.ReplaceAll(tspath.NormalizePath(path), "\\", "/")
	var components []string
	for _, component := range strings.Split(path, "/") {
		if component == "" || component == "." {
			continue
		}
		components = append(components, component)
	}
	return components
}
