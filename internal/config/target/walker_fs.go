package target

import (
	"sort"
	"strings"

	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
)

type lintDirectoryEntry struct {
	name          string
	directory     bool
	needsRealpath bool
}

// readLintDirectory reads and sorts one directory without following directory
// symlinks. That is the target walker's ESLint-compatible policy; config's
// project-glob adapter remains independent because it intentionally follows
// directory aliases with cycle detection.
func readLintDirectory(fsys vfs.FS, directory string) []lintDirectoryEntry {
	if fsys == nil {
		return nil
	}
	accessible := fsys.GetAccessibleEntries(directory)
	entries := make([]lintDirectoryEntry, 0, len(accessible.Directories)+len(accessible.Files))

	parentRealPath := ""
	for _, name := range accessible.Directories {
		directoryPath := tspath.CombinePaths(directory, name)
		isSymlink := false
		if accessible.Symlinks != nil {
			_, isSymlink = accessible.Symlinks[name]
		} else {
			if parentRealPath == "" {
				parentRealPath = fsys.Realpath(directory)
			}
			realPath := fsys.Realpath(directoryPath)
			expectedRealPath := tspath.CombinePaths(parentRealPath, name)
			isSymlink = !rslintconfig.PathsEqual(realPath, expectedRealPath, fsys.UseCaseSensitiveFileNames())
		}
		if !isSymlink {
			entries = append(entries, lintDirectoryEntry{name: name, directory: true})
		}
	}
	for _, name := range accessible.Files {
		isSymlink := false
		needsRealpath := accessible.Symlinks == nil
		if accessible.Symlinks != nil {
			_, isSymlink = accessible.Symlinks[name]
		}
		entries = append(entries, lintDirectoryEntry{
			name:          name,
			needsRealpath: needsRealpath || isSymlink,
		})
	}
	sort.Slice(entries, func(left, right int) bool {
		return entries[left].name < entries[right].name
	})
	return entries
}

func normalizeWalkPath(filePath string) string {
	return strings.ReplaceAll(tspath.NormalizePath(filePath), "\\", "/")
}
