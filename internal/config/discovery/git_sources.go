package discovery

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
)

// physicalDirectory freezes non-input directory identities on first use.
// Request roots and file parents have already populated this map.
func (coordinator *discoveryCoordinator) physicalDirectory(directory string) string {
	directory = tspath.NormalizePath(directory)
	if coordinator.physicalDirectories == nil {
		coordinator.physicalDirectories = make(map[string]string)
	}
	if physical, exists := coordinator.physicalDirectories[directory]; exists {
		return physical
	}
	physical := directory
	for ancestor := configDiscoveryParent(directory); ancestor != ""; ancestor = configDiscoveryParent(ancestor) {
		if parentPhysical, exists := coordinator.physicalDirectories[ancestor]; exists {
			relative, within := rslintconfig.RelativePathWithinConfigRoot(directory, ancestor, true)
			if within {
				physical = tspath.ResolvePath(parentPhysical, relative)
				break
			}
		}
	}
	if realPath := coordinator.fs.Realpath(physical); realPath != "" {
		physical = tspath.NormalizePath(realPath)
	}
	coordinator.physicalDirectories[directory] = physical
	return physical
}

// collectExactGitFiles supplements the same reader with every supplied file,
// including glob-expanded inputs and documents that do not discover configs.
func (coordinator *discoveryCoordinator) collectExactGitFiles() {
	for _, file := range coordinator.request.Files {
		coordinator.collectExactGitFile(file.PathIdentity)
	}
	for _, file := range coordinator.request.AdditionalGitFiles {
		coordinator.collectExactGitFile(file)
	}
}

func (coordinator *discoveryCoordinator) collectExactGitFile(file rslintconfig.PathIdentity) {
	parent := file.CanonicalParentPath
	if parent == "" {
		parent = tspath.GetDirectoryPath(normalizeDiscoveryPath(file.Path, coordinator.request.CWD))
	} else {
		parent = normalizeDiscoveryPath(parent, coordinator.request.CWD)
	}
	cursor := coordinator.git.Ancestors(parent)
	coordinator.git.Observe(cursor, parent)
}

func isSymlinkDirectoryChild(
	fsys vfs.FS,
	parentDirectory string,
	parentRealPath string,
	name string,
	entries vfs.Entries,
) bool {
	if entries.Symlinks != nil {
		for symlink := range entries.Symlinks {
			if symlink == name || (!fsys.UseCaseSensitiveFileNames() && strings.EqualFold(symlink, name)) {
				return true
			}
		}
		return false
	}
	childDirectory := tspath.CombinePaths(parentDirectory, name)
	childRealPath := fsys.Realpath(childDirectory)
	if parentRealPath == "" || childRealPath == "" {
		return false
	}
	expectedRealPath := tspath.CombinePaths(parentRealPath, name)
	return tspath.ComparePaths(childRealPath, expectedRealPath, tspath.ComparePathsOptions{
		UseCaseSensitiveFileNames: fsys.UseCaseSensitiveFileNames(),
	}) != 0
}
