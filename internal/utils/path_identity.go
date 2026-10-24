package utils

import (
	"os"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	"github.com/web-infra-dev/rslint/internal/utils/pathutil"
)

// DirectoryIdentity keeps one directory's caller-visible and physical paths.
// CanonicalPath can be omitted when the caller has not frozen a physical path.
type DirectoryIdentity struct {
	LexicalPath   string
	CanonicalPath string
}

// DirectorySearchBoundary projects root onto directory's lexical parent chain.
// Exact lexical ancestry wins for this root, including links below it. Otherwise
// lexical ancestors establish physical directory ancestry and locate the
// equivalent boundary without assuming equal directory depths.
// A false result means the directory is outside this root. Neither global case
// policy nor leaf file identity participates in this directory-only operation.
func DirectorySearchBoundary(directory, root DirectoryIdentity, fsys vfs.FS) (string, bool) {
	normalizeDirectory := func(path string) string {
		path = ExactPathID(path)
		if len(path) > tspath.GetRootLength(path) {
			path = pathutil.RemoveTrailingDirectorySeparators(path)
		}
		return path
	}
	directory.LexicalPath = normalizeDirectory(directory.LexicalPath)
	root.LexicalPath = normalizeDirectory(root.LexicalPath)
	if directory.LexicalPath == "" || root.LexicalPath == "" {
		return "", false
	}
	containsExact := func(parent, child string) bool {
		return pathutil.ContainsPath(parent, child, pathutil.ComparePathsOptions{UseCaseSensitiveFileNames: true})
	}
	if containsExact(root.LexicalPath, directory.LexicalPath) {
		return root.LexicalPath, true
	}
	physicalDirectory := directory.CanonicalPath
	if physicalDirectory == "" {
		physicalDirectory = directory.LexicalPath
		if fsys != nil {
			if physical := fsys.Realpath(pathutil.Path(physicalDirectory)); physical != "" {
				physicalDirectory = physical.AsString()
			}
		}
	}
	physicalRoot := root.CanonicalPath
	if physicalRoot == "" {
		physicalRoot = root.LexicalPath
		if fsys != nil {
			if physical := fsys.Realpath(pathutil.Path(physicalRoot)); physical != "" {
				physicalRoot = physical.AsString()
			}
		}
	}
	physicalDirectory, physicalRoot = normalizeDirectory(physicalDirectory), normalizeDirectory(physicalRoot)
	var rootInfo vfs.FileInfo
	if fsys != nil {
		rootInfo = fsys.Stat(pathutil.Path(physicalRoot))
	}
	sameRoot := func(path string) bool {
		if normalizeDirectory(path) == physicalRoot {
			return true
		}
		if rootInfo == nil || !rootInfo.IsDir() {
			return false
		}
		info := fsys.Stat(pathutil.Path(path))
		return info != nil && info.IsDir() && os.SameFile(rootInfo, info)
	}
	containsPhysical := func(path string) bool {
		path = normalizeDirectory(path)
		if containsExact(physicalRoot, path) {
			return true
		}
		if rootInfo == nil || !rootInfo.IsDir() {
			return false
		}
		for ancestor := path; ; ancestor = tspath.GetDirectoryPath(ancestor) {
			if sameRoot(ancestor) {
				return true
			}
			if tspath.GetDirectoryPath(ancestor) == ancestor {
				return false
			}
		}
	}
	boundary := ""
	for ancestor := directory.LexicalPath; ; ancestor = tspath.GetDirectoryPath(ancestor) {
		physical := physicalDirectory
		if ancestor != directory.LexicalPath {
			if fsys == nil {
				break
			}
			physical = fsys.Realpath(pathutil.Path(ancestor)).AsString()
		}
		if physical != "" && containsPhysical(physical) {
			boundary = ancestor
		}
		if tspath.GetDirectoryPath(ancestor) == ancestor {
			break
		}
	}
	return boundary, boundary != ""
}

// ExactPathID normalizes a path without consulting the filesystem or folding
// directory and file names. Only absolute Windows drive spelling is folded.
func ExactPathID(filePath string) string {
	// Identity keys also include relative paths and the empty sentinel. Those
	// are not compiler paths and must not be coerced into a rooted PathKey.
	// A bare drive is drive-relative, even though the compiler normalizer now
	// supplies its root separator. Preserve its case-sensitive identity.
	if len(filePath) == 2 && filePath[1] == ':' {
		return filePath
	}
	if tspath.IsRootedDiskPath(filePath) {
		filePath = tspath.NormalizePath(filePath)
	} else {
		filePath = tspath.GetNormalizedAbsolutePath(filePath, "")
	}
	return NormalizeAbsoluteDrive(filePath)
}

// NormalizeAbsoluteDrive gives absolute Windows paths a consistent drive
// spelling without changing directory names, file names, or path separators.
// Callers remain responsible for resolving or normalizing the rest of the path.
func NormalizeAbsoluteDrive(filePath string) string {
	// SplitVolumePath also accepts drive-relative paths. Keep those spellings
	// intact, and avoid rebuilding paths whose volume is already canonical.
	volume, path, ok := tspath.SplitVolumePath(filePath)
	if ok && tspath.GetRootLength(filePath) > len(volume) && filePath[:len(volume)] != volume {
		return volume + path
	}
	return filePath
}
