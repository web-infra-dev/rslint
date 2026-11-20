package discovery

import (
	"slices"
	"sort"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
)

// discoveryTargetTrie bounds a directory walk to branches that can govern an
// invocation target. A nil trie means the current subtree is unbounded. A
// leaf still visits that directory so its sources can govern files within it.
type discoveryTargetTrie struct {
	children map[string]*discoveryTargetTrie
	files    []*rslintconfig.PathIdentity
}

func addDiscoveryTarget(
	root *discoveryTargetTrie,
	configDir string,
	file *rslintconfig.PathIdentity,
	fsys vfs.FS,
) bool {
	parent := tspath.GetDirectoryPath(file.Path)
	relative, within := rslintconfig.RelativePathWithinConfigRoot(
		parent, configDir, true,
	)
	if !within {
		relative, within = rslintconfig.RelativePathWithinConfigRoot(parent, configDir, false)
		if !within {
			return false
		}
		aliasRoot := parent
		for range splitDiscoveryPath(relative) {
			aliasRoot = tspath.GetDirectoryPath(aliasRoot)
		}
		if !fsys.DirectoryExists(aliasRoot) || !sameDirectoryEntries(fsys, aliasRoot, configDir) {
			return false
		}
	}

	node := root
	for _, component := range splitDiscoveryPath(relative) {
		if node.children == nil {
			node.children = make(map[string]*discoveryTargetTrie)
		}
		child := node.children[component]
		if child == nil {
			child = &discoveryTargetTrie{}
			node.children[component] = child
		}
		node = child
	}
	node.files = append(node.files, file)
	return true
}

func (coordinator *discoveryCoordinator) resolveInputNames(files []DiscoveryFile, directories []rslintconfig.DirectoryIdentity) {
	for i := range directories {
		directories[i].CanonicalPath = coordinator.inputNames.Directory(directories[i].CanonicalPath)
	}
	for i := range files {
		file := &files[i]
		file.PathIdentity = coordinator.inputNames.File(file.PathIdentity, !file.Explicit && len(file.ScanDirectories) != 0)
	}
}

func (coordinator *discoveryCoordinator) normalizedDirectoryRoots() []rslintconfig.DirectoryIdentity {
	raw := coordinator.request.Directories
	if len(raw) == 0 && len(coordinator.request.Files) == 0 && coordinator.request.ImplicitCWD {
		raw = rslintconfig.ResolveDirectoryIdentities([]string{coordinator.request.CWD}, coordinator.fs)
	}
	byIdentity := make(map[string]rslintconfig.DirectoryIdentity, len(raw))
	for _, directory := range raw {
		directory.LexicalPath = normalizeDiscoveryPath(directory.LexicalPath, coordinator.request.CWD)
		if directory.CanonicalPath == "" {
			directory.CanonicalPath = directory.LexicalPath
		} else {
			directory.CanonicalPath = normalizeDiscoveryPath(directory.CanonicalPath, coordinator.request.CWD)
		}
		identity := rslintconfig.ExactPathID(directory.LexicalPath)
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

func (coordinator *discoveryCoordinator) recursiveDirectory(path string) bool {
	root := rslintconfig.ExactPathID(path)
	return slices.ContainsFunc(coordinator.request.RecursiveDirectories, func(directory rslintconfig.DirectoryIdentity) bool {
		return rslintconfig.ExactPathID(normalizeDiscoveryPath(directory.LexicalPath, coordinator.request.CWD)) == root
	})
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
	for _, directory := range roots {
		root := directory.LexicalPath
		rootID := rslintconfig.ExactPathID(root)
		trie := &discoveryTargetTrie{}
		for index := range files {
			file := &files[index]
			if !file.Explicit && file.ScanDirectories != nil && !slices.ContainsFunc(file.ScanDirectories, func(directory string) bool {
				return rslintconfig.ExactPathID(directory) == rootID
			}) {
				continue
			}
			if addDiscoveryTarget(trie, root, &file.PathIdentity, coordinator.fs) {
				tries[root] = trie
			}
		}
	}
	return tries
}

// Each trie belongs to one walk root. A caller may supply a native alias for
// an existing directory, while glob expansion already supplies stored names.
func (trie *discoveryTargetTrie) resolveChildNames(parent string, directories []string, fsys vfs.FS) {
	for requested, child := range trie.children {
		if _, exists := slices.BinarySearch(directories, requested); exists {
			continue
		}
		name := rslintconfig.StoredEntryName(directories, requested)
		if name == "" || !fsys.DirectoryExists(tspath.CombinePaths(parent, requested)) {
			continue
		}
		if existing := trie.children[name]; existing != nil {
			existing.merge(child)
		} else {
			trie.children[name] = child
		}
		delete(trie.children, requested)
	}
}

func (trie *discoveryTargetTrie) merge(other *discoveryTargetTrie) {
	trie.files = append(trie.files, other.files...)
	if trie.children == nil {
		trie.children = other.children
		return
	}
	for name, child := range other.children {
		if existing := trie.children[name]; existing != nil {
			existing.merge(child)
		} else {
			trie.children[name] = child
		}
	}
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
