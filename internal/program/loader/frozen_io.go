package loader

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	"github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/config/target"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// WithFrozenIO keeps Program filenames in their authored path space while
// directing filesystem operations through the identities captured by planning.
// Apply it before NewSession, and put each fix generation's memory overlay
// above the returned view. Unknown paths retain their current coordinates.
func WithFrozenIO(baseFS vfs.FS, pathSpaces *config.PathSpaceSnapshot, files []target.File) (vfs.FS, error) {
	if isNilFS(baseFS) {
		return nil, errors.New("program loader: frozen IO requires a filesystem")
	}
	directories := pathSpaces.DirectoryIdentities()
	caseSensitive := baseFS.UseCaseSensitiveFileNames()

	if canUseIdentityFastPath(caseSensitive, directories, files) {
		view := &frozenIdentityFS{
			FS: baseFS,
		}
		if len(files) > 0 {
			frozenFiles := make(map[string]struct{}, len(files))
			for _, file := range files {
				if file.CanonicalPath != "" {
					frozenFiles[tspath.NormalizePath(file.CanonicalPath)] = struct{}{}
				}
			}
			if len(frozenFiles) > 0 {
				view.files = frozenFiles
			}
		}
		if len(directories) > 0 || len(files) > 0 {
			identityDirs := make(map[string]struct{}, len(directories))
			for _, directory := range directories {
				identityDirs[directory.LexicalPath] = struct{}{}
			}
			for _, file := range files {
				if file.Path != "" && file.CanonicalPath != "" && file.CanonicalParentPath != "" {
					parent := tspath.GetDirectoryPath(tspath.NormalizePath(file.Path))
					identityDirs[parent] = struct{}{}
				}
			}
			if len(identityDirs) > 0 {
				view.identityDirs = identityDirs
			}
		}
		return view, nil
	}

	view := &frozenFileSystem{
		FS:      baseFS,
		parents: make(map[string]string, len(directories)),
	}
	if len(files) > 0 {
		view.files = make(map[string]string, len(files))
	}
	for _, directory := range directories {
		view.parents[config.ExactPathID(directory.LexicalPath)] = directory.CanonicalPath
	}
	// A canonical target path is already an identity, even if another input
	// happens to use that spelling as its lexical alias.
	for _, file := range files {
		if file.CanonicalPath != "" {
			canonical := tspath.NormalizePath(file.CanonicalPath)
			view.files[config.ExactPathID(canonical)] = canonical
		}
	}
	for _, file := range files {
		if file.Path == "" || file.CanonicalPath == "" {
			continue
		}
		path := tspath.NormalizePath(file.Path)
		canonical := tspath.NormalizePath(file.CanonicalPath)
		if path != canonical {
			id := config.ExactPathID(path)
			if _, exists := view.files[id]; !exists {
				view.files[id] = canonical
			}
		}
		if file.CanonicalParentPath != "" {
			parent := tspath.GetDirectoryPath(path)
			physicalParent := tspath.NormalizePath(file.CanonicalParentPath)
			parentID := config.ExactPathID(parent)
			if existing, ok := view.parents[parentID]; ok && config.ExactPathID(existing) != config.ExactPathID(physicalParent) {
				return nil, fmt.Errorf("program loader: conflicting frozen parent identities for %q", parent)
			}
			view.parents[parentID] = physicalParent
			if parent != physicalParent || (len(path) > tspath.GetRootLength(path) && tspath.HasTrailingDirectorySeparator(path)) {
				alias := tspath.CombinePaths(physicalParent, tspath.GetBaseFileName(path))
				id := config.ExactPathID(alias)
				if _, exists := view.files[id]; !exists {
					view.files[id] = canonical
				}
			}
		}
	}
	if !caseSensitive {
		view.keyPrefix = frozenKeyPrefix(view.files, view.parents)
		view.foldedKeyPrefix = tspath.GetCanonicalFileName(view.keyPrefix, false)
		view.foldedFiles = foldedFrozenPaths(view.files, view.keyPrefix)
		view.foldedParents = foldedFrozenPaths(view.parents, view.keyPrefix)
	}
	return view, nil
}

func isExactPOSIXIdentity(path string) bool {
	return tspath.GetRootLength(path) == 1 &&
		(len(path) <= 1 || !tspath.HasTrailingDirectorySeparator(path)) &&
		tspath.NormalizePath(path) == path &&
		config.ExactPathID(path) == path
}

// canUseIdentityFastPath reports whether every planned directory and file
// identity maps exclusively to itself with exact byte equality, so ioPath and
// Realpath can bypass parent-map walking and the projection cache lock without
// changing any path spelling.
//
// An empty snapshot is always identity-only on any filesystem. For non-empty
// snapshots, case-insensitive filesystems still canonicalize differently-cased
// caller spellings against frozen directories and files (e.g. mixed-volume
// case-sensitive mounts), and ExactPathID folds Windows drive letters and
// non-disk roots even on case-sensitive filesystems; restrict non-empty fast
// paths to case-sensitive exact POSIX identities where no rewrite is possible.
func canUseIdentityFastPath(caseSensitive bool, directories []config.DirectoryIdentity, files []target.File) bool {
	if len(directories) == 0 && len(files) == 0 {
		return true
	}
	if !caseSensitive {
		return false
	}
	for _, directory := range directories {
		if directory.LexicalPath != directory.CanonicalPath || !isExactPOSIXIdentity(directory.LexicalPath) {
			return false
		}
	}
	for _, file := range files {
		if file.CanonicalPath == "" {
			continue
		}
		canonical := tspath.NormalizePath(file.CanonicalPath)
		if !isExactPOSIXIdentity(canonical) {
			return false
		}
		if file.Path == "" {
			continue
		}
		path := tspath.NormalizePath(file.Path)
		if path != canonical {
			return false
		}
		if file.CanonicalParentPath != "" {
			parent := tspath.GetDirectoryPath(path)
			physicalParent := tspath.NormalizePath(file.CanonicalParentPath)
			if parent != physicalParent || !isExactPOSIXIdentity(parent) {
				return false
			}
		}
	}
	return true
}

// frozenIdentityFS is a lock-free frozen IO view for snapshots whose planned
// directory and file identities map exclusively to themselves (exact POSIX
// paths on case-sensitive filesystems, or empty snapshots). Keeping it separate
// from frozenFileSystem avoids adding branches or fields to the non-identity
// hot path.
type frozenIdentityFS struct {
	vfs.FS
	files        map[string]struct{}
	identityDirs map[string]struct{}
}

func (f *frozenIdentityFS) hasFrozenIdentityDirAncestor(dir string) bool {
	if len(f.identityDirs) == 0 || tspath.GetRootLength(dir) != 1 {
		return false
	}
	for current := dir; current != ""; {
		if _, ok := f.identityDirs[current]; ok {
			return true
		}
		parent := tspath.GetDirectoryPath(current)
		if parent == current {
			break
		}
		current = parent
	}
	return false
}

func (f *frozenIdentityFS) ioPath(path string) string {
	normalized := tspath.NormalizePath(path)
	if tspath.HasTrailingDirectorySeparator(normalized) && len(normalized) > tspath.GetRootLength(normalized) {
		trimmed := tspath.RemoveTrailingDirectorySeparator(normalized)
		if f.hasFrozenIdentityDirAncestor(trimmed) {
			return trimmed
		}
	}
	return normalized
}

func (f *frozenIdentityFS) FileExists(path string) bool { return f.FS.FileExists(f.ioPath(path)) }

func (f *frozenIdentityFS) ReadFile(path string) (string, bool) { return f.FS.ReadFile(f.ioPath(path)) }

func (f *frozenIdentityFS) ReadSourceBOM(path string) (bool, error) {
	return utils.ReadSourceBOM(f.FS, f.ioPath(path))
}

func (f *frozenIdentityFS) DirectoryExists(path string) bool {
	return f.FS.DirectoryExists(f.ioPath(path))
}

func (f *frozenIdentityFS) GetAccessibleEntries(path string) vfs.Entries {
	return f.FS.GetAccessibleEntries(f.ioPath(path))
}

func (f *frozenIdentityFS) Stat(path string) vfs.FileInfo { return f.FS.Stat(f.ioPath(path)) }

func (f *frozenIdentityFS) Realpath(path string) string {
	normalized := tspath.NormalizePath(path)
	if len(f.files) > 0 {
		if _, ok := f.files[normalized]; ok {
			return normalized
		}
	}
	if tspath.HasTrailingDirectorySeparator(normalized) && len(normalized) > tspath.GetRootLength(normalized) {
		trimmed := tspath.RemoveTrailingDirectorySeparator(normalized)
		if f.hasFrozenIdentityDirAncestor(trimmed) {
			normalized = trimmed
		}
	}
	return f.FS.Realpath(normalized)
}

func (f *frozenIdentityFS) WriteFile(path string, data string) error {
	return f.FS.WriteFile(f.ioPath(path), data)
}

func (f *frozenIdentityFS) AppendFile(path string, data string) error {
	return f.FS.AppendFile(f.ioPath(path), data)
}

func (f *frozenIdentityFS) Remove(path string) error { return f.FS.Remove(f.ioPath(path)) }

func (f *frozenIdentityFS) Chtimes(path string, atime time.Time, mtime time.Time) error {
	return f.FS.Chtimes(f.ioPath(path), atime, mtime)
}

func (f *frozenIdentityFS) WalkDir(root string, walkFn vfs.WalkDirFunc) error {
	root = tspath.NormalizePath(root)
	physicalRoot := f.ioPath(root)
	if physicalRoot == root {
		return f.FS.WalkDir(root, walkFn)
	}
	return f.FS.WalkDir(physicalRoot, func(path string, entry vfs.DirEntry, err error) error {
		relative := tspath.GetRelativePathFromDirectory(physicalRoot, tspath.NormalizePath(path), tspath.ComparePathsOptions{UseCaseSensitiveFileNames: true})
		return walkFn(tspath.ResolvePath(root, relative), entry, err)
	})
}

// frozenFileSystem projects immutable planning identities. Only path projection
// is memoized; existence, contents and non-frozen Realpath results use the live FS.
type frozenFileSystem struct {
	vfs.FS
	files           map[string]string
	parents         map[string]string
	foldedFiles     map[string]string
	foldedParents   map[string]string
	keyPrefix       string
	foldedKeyPrefix string
	pathMu          sync.RWMutex
	paths           map[string]frozenPath
}

const frozenPathCacheLimit = 4096

type frozenPath struct {
	path string
	file bool
}

// Fold only the varying suffix of paths in this view. In particular, an
// uppercase request directory must not allocate a complete lowercase string
// for every otherwise-lowercase dependency lookup. Roots and non-disk paths
// retain the full-key algorithm.
func frozenKeyPrefix(groups ...map[string]string) string {
	var prefix string
	first := true
	for _, paths := range groups {
		for path := range paths {
			if !tspath.IsRootedDiskPath(path) {
				return ""
			}
			if first {
				prefix = path
				first = false
			}
			length := min(len(prefix), len(path)-1)
			index := 0
			for index < length && prefix[index] == path[index] {
				index++
			}
			prefix = prefix[:index]
		}
	}
	prefix = prefix[:strings.LastIndexByte(prefix, '/')+1]
	if len(prefix) <= tspath.GetRootLength(prefix) {
		return ""
	}
	return prefix
}

// A folded spelling is usable only when all known exact spellings agree on
// one physical identity. An empty value permanently disables an ambiguous key.
func foldedFrozenPaths(paths map[string]string, prefix string) map[string]string {
	if len(paths) == 0 {
		return nil
	}
	folded := make(map[string]string, len(paths))
	for path, physical := range paths {
		id := tspath.GetCanonicalFileName(strings.TrimPrefix(path, prefix), false)
		if existing, ok := folded[id]; !ok {
			folded[id] = physical
		} else if config.ExactPathID(existing) != config.ExactPathID(physical) {
			folded[id] = ""
		}
	}
	return folded
}

func (f *frozenFileSystem) foldedKey(path string) string {
	if suffix, ok := strings.CutPrefix(path, f.keyPrefix); ok {
		return tspath.GetCanonicalFileName(suffix, false)
	}
	// Fold before removing a differently-cased prefix: Unicode folding can
	// change byte lengths, including inside the prefix itself.
	folded := tspath.GetCanonicalFileName(path, false)
	if suffix, ok := strings.CutPrefix(folded, f.foldedKeyPrefix); ok {
		return suffix
	}
	return ""
}

func (f *frozenFileSystem) ioPath(path string) string {
	physical, _ := f.resolvePath(path)
	return physical
}

// resolvePath also reports frozen file identities: Realpath must return those
// directly instead of observing a file alias again in the underlying FS.
func (f *frozenFileSystem) resolvePath(path string) (string, bool) {
	f.pathMu.RLock()
	cached, ok := f.paths[path]
	full := len(f.paths) >= frozenPathCacheLimit
	f.pathMu.RUnlock()
	if ok {
		return cached.path, cached.file
	}
	physical, file := f.computePath(path)
	if !full {
		f.pathMu.Lock()
		if len(f.paths) < frozenPathCacheLimit {
			if f.paths == nil {
				f.paths = make(map[string]frozenPath)
			}
			// Path strings may borrow source buffers; retain only their own bytes.
			f.paths[strings.Clone(path)] = frozenPath{path: strings.Clone(physical), file: file}
		}
		f.pathMu.Unlock()
	}
	return physical, file
}

func (f *frozenFileSystem) computePath(path string) (string, bool) {
	path = tspath.NormalizePath(path)
	diskPath := tspath.IsRootedDiskPath(path)
	id := utils.NormalizeAbsoluteDrive(path)
	if !diskPath {
		id = config.ExactPathID(path)
	}
	if physical, ok := f.files[id]; ok {
		return physical, true
	}
	// Fold a disk path once, then walk its ancestors independently. A
	// Unicode case fold can change byte lengths, so never slice one spelling
	// using offsets from the other.
	folded := ""
	if len(f.foldedFiles) != 0 || len(f.foldedParents) != 0 {
		folded = f.foldedKey(id)
		if physical := f.foldedFiles[folded]; physical != "" {
			return physical, true
		}
	}
	// Normalize before choosing an anchor. In particular, ../outside in an
	// aliased tsconfig keeps its lexical sibling rather than the physical one.
	if len(f.parents) == 0 {
		return path, false
	}
	directory := path
	if len(directory) > tspath.GetRootLength(directory) && tspath.HasTrailingDirectorySeparator(directory) {
		// A trailing separator still addresses this directory, not its parent.
		directory = tspath.RemoveTrailingDirectorySeparator(directory)
		id = tspath.RemoveTrailingDirectorySeparator(id)
		folded = tspath.RemoveTrailingDirectorySeparator(folded)
	}
	for directory != "" {
		physical, ok := f.parents[id]
		if !ok && folded != "" {
			physical = f.foldedParents[folded]
			ok = physical != ""
		}
		if ok {
			if physical == directory {
				// Relative-path projection drops a trailing separator, except at
				// a filesystem root. Preserve that spelling without rebuilding it.
				if len(path) > tspath.GetRootLength(path) {
					path = tspath.RemoveTrailingDirectorySeparator(path)
				}
				return path, false
			}
			relative := tspath.GetRelativePathFromDirectory(directory, path, tspath.ComparePathsOptions{UseCaseSensitiveFileNames: true})
			return tspath.ResolvePath(physical, relative), false
		}
		parent := tspath.GetDirectoryPath(directory)
		if parent == directory {
			break
		}
		directory = parent
		if diskPath {
			id = tspath.GetDirectoryPath(id)
			if folded != "" {
				if f.keyPrefix == "" {
					folded = tspath.GetDirectoryPath(folded)
				} else if index := strings.LastIndexByte(folded, '/'); index >= 0 {
					// A suffix such as c:/child is an ordinary directory here,
					// not a new drive root.
					folded = folded[:index]
				} else {
					folded = ""
				}
			}
		} else {
			// Relative paths and URLs have distinct root/key rules. Keep their
			// identity semantics without adding that work to disk-path IO.
			id = config.ExactPathID(directory)
			if len(f.foldedParents) != 0 {
				folded = f.foldedKey(id)
			}
		}
	}
	return path, false
}

func (f *frozenFileSystem) FileExists(path string) bool { return f.FS.FileExists(f.ioPath(path)) }

func (f *frozenFileSystem) ReadFile(path string) (string, bool) { return f.FS.ReadFile(f.ioPath(path)) }

func (f *frozenFileSystem) ReadSourceBOM(path string) (bool, error) {
	return utils.ReadSourceBOM(f.FS, f.ioPath(path))
}

func (f *frozenFileSystem) DirectoryExists(path string) bool {
	return f.FS.DirectoryExists(f.ioPath(path))
}

func (f *frozenFileSystem) GetAccessibleEntries(path string) vfs.Entries {
	return f.FS.GetAccessibleEntries(f.ioPath(path))
}

func (f *frozenFileSystem) Stat(path string) vfs.FileInfo { return f.FS.Stat(f.ioPath(path)) }

func (f *frozenFileSystem) Realpath(path string) string {
	physical, frozenFile := f.resolvePath(path)
	if frozenFile {
		return physical
	}
	return f.FS.Realpath(physical)
}

func (f *frozenFileSystem) WriteFile(path string, data string) error {
	return f.FS.WriteFile(f.ioPath(path), data)
}

func (f *frozenFileSystem) AppendFile(path string, data string) error {
	return f.FS.AppendFile(f.ioPath(path), data)
}

func (f *frozenFileSystem) Remove(path string) error { return f.FS.Remove(f.ioPath(path)) }

func (f *frozenFileSystem) Chtimes(path string, atime time.Time, mtime time.Time) error {
	return f.FS.Chtimes(f.ioPath(path), atime, mtime)
}

func (f *frozenFileSystem) WalkDir(root string, walkFn vfs.WalkDirFunc) error {
	root = tspath.NormalizePath(root)
	physicalRoot := f.ioPath(root)
	if physicalRoot == root {
		return f.FS.WalkDir(root, walkFn)
	}
	return f.FS.WalkDir(physicalRoot, func(path string, entry vfs.DirEntry, err error) error {
		relative := tspath.GetRelativePathFromDirectory(physicalRoot, tspath.NormalizePath(path), tspath.ComparePathsOptions{UseCaseSensitiveFileNames: true})
		return walkFn(tspath.ResolvePath(root, relative), entry, err)
	})
}
