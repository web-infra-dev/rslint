package discovery

import (
	"slices"
	"sort"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
)

// discoveryTargetTrie bounds a directory walk to branches that can govern an
// invocation target. A nil trie means the current subtree is unbounded. A
// leaf still visits that directory so its sources can govern files within it.
type discoveryTargetTrie struct {
	children map[tspath.Path]*discoveryTargetTrie
	files    []string
}

func addDiscoveryTarget(
	root *discoveryTargetTrie,
	configDir string,
	filePath string,
	useCaseSensitive bool,
) bool {
	relative, within := rslintconfig.RelativePathWithinConfigRoot(
		tspath.GetDirectoryPath(filePath),
		configDir,
		useCaseSensitive,
	)
	if !within {
		return false
	}

	node := root
	for _, component := range splitDiscoveryPath(relative) {
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
	node.files = append(node.files, filePath)
	return true
}

func (coordinator *discoveryCoordinator) normalizedDirectoryRoots() []rslintconfig.DirectoryIdentity {
	raw := coordinator.request.Directories
	if len(raw) == 0 && len(coordinator.request.Files) == 0 && coordinator.request.ImplicitCWD {
		raw = rslintconfig.ResolveDirectoryIdentities([]string{coordinator.request.CWD}, coordinator.fs)
	}
	useCaseSensitive := coordinator.fs.UseCaseSensitiveFileNames()
	byIdentity := make(map[tspath.Path]rslintconfig.DirectoryIdentity, len(raw))
	for _, directory := range raw {
		directory.LexicalPath = normalizeDiscoveryPath(directory.LexicalPath, coordinator.request.CWD)
		if directory.CanonicalPath == "" {
			directory.CanonicalPath = directory.LexicalPath
		} else {
			directory.CanonicalPath = normalizeDiscoveryPath(directory.CanonicalPath, coordinator.request.CWD)
		}
		identity := tspath.ToPath(directory.LexicalPath, "", useCaseSensitive)
		if current, exists := byIdentity[identity]; !exists || directory.LexicalPath < current.LexicalPath {
			byIdentity[identity] = directory
		}
	}
	roots := make([]rslintconfig.DirectoryIdentity, 0, len(byIdentity))
	for _, directory := range byIdentity {
		roots = append(roots, directory)
	}
	sort.Slice(roots, func(i, j int) bool { return roots[i].LexicalPath < roots[j].LexicalPath })
	return roots
}

func (coordinator *discoveryCoordinator) normalizedFiles() []DiscoveryFile {
	if coordinator.request.Files == nil {
		return nil
	}
	files := make([]DiscoveryFile, 0, len(coordinator.request.Files))
	indexByPath := make(map[string]int, len(coordinator.request.Files))
	for _, file := range coordinator.request.Files {
		file.Path = normalizeDiscoveryPath(file.Path, coordinator.request.CWD)
		if file.CanonicalPath != "" {
			file.CanonicalPath = normalizeDiscoveryPath(file.CanonicalPath, coordinator.request.CWD)
		}
		if file.CanonicalParentPath != "" {
			file.CanonicalParentPath = normalizeDiscoveryPath(file.CanonicalParentPath, coordinator.request.CWD)
		}
		file.ScanDirectories = slices.Clone(file.ScanDirectories)
		for index, directory := range file.ScanDirectories {
			file.ScanDirectories[index] = normalizeDiscoveryPath(directory, coordinator.request.CWD)
		}
		if index, exists := indexByPath[file.Path]; exists {
			files[index].Explicit = files[index].Explicit || file.Explicit
			if files[index].ScanDirectories != nil {
				if file.ScanDirectories == nil {
					files[index].ScanDirectories = nil
				} else {
					files[index].ScanDirectories = append(files[index].ScanDirectories, file.ScanDirectories...)
				}
			}
			if files[index].CanonicalPath == "" {
				files[index].CanonicalPath = file.CanonicalPath
			}
			if files[index].CanonicalParentPath == "" {
				files[index].CanonicalParentPath = file.CanonicalParentPath
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
	roots []rslintconfig.DirectoryIdentity,
	files []DiscoveryFile,
) map[string]*discoveryTargetTrie {
	tries := make(map[string]*discoveryTargetTrie, len(roots))
	useCaseSensitive := coordinator.fs.UseCaseSensitiveFileNames()
	for _, file := range files {
		for _, directory := range roots {
			root := directory.LexicalPath
			if !file.Explicit && file.ScanDirectories != nil && !slices.ContainsFunc(file.ScanDirectories, func(directory string) bool {
				return discoveryPathsEqual(directory, root, useCaseSensitive)
			}) {
				continue
			}
			trie := tries[root]
			if trie == nil {
				trie = &discoveryTargetTrie{}
			}
			if addDiscoveryTarget(trie, root, file.Path, useCaseSensitive) {
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
