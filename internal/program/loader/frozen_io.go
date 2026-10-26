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
	"github.com/web-infra-dev/rslint/internal/utils/pathutil"
)

// WithFrozenIO keeps Program filenames in their authored path space while
// directing filesystem operations through the identities captured by planning.
// Apply it before NewSession, and put each fix generation's memory overlay
// above the returned view. Unknown paths retain their current coordinates.
func WithFrozenIO(baseFS vfs.FS, pathSpaces *config.PathSpaceSnapshot, files []target.File) (vfs.FS, error) {
	if isNilFS(baseFS) {
		return nil, errors.New("program loader: frozen IO requires a filesystem")
	}
	view := &frozenFileSystem{
		FS:    baseFS,
		files: make(map[string]string, len(files)), parents: make(map[string]string),
	}
	for _, directory := range pathSpaces.DirectoryIdentities() {
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
		aliases := []string{path}
		if file.CanonicalParentPath != "" {
			parent := tspath.GetDirectoryPath(path)
			physicalParent := tspath.NormalizePath(file.CanonicalParentPath)
			parentID := config.ExactPathID(parent)
			if existing, ok := view.parents[parentID]; ok && config.ExactPathID(existing) != config.ExactPathID(physicalParent) {
				return nil, fmt.Errorf("program loader: conflicting frozen parent identities for %q", parent)
			}
			view.parents[parentID] = physicalParent
			aliases = append(aliases, tspath.CombinePaths(physicalParent, tspath.GetBaseFileName(path)))
		}
		for _, alias := range aliases {
			id := config.ExactPathID(alias)
			if _, exists := view.files[id]; !exists {
				view.files[id] = canonical
			}
		}
	}
	if !baseFS.CaseSensitivity().IsCaseSensitive() {
		view.keyPrefix = frozenKeyPrefix(view.files, view.parents)
		view.foldedKeyPrefix = pathutil.GetCanonicalFileName(view.keyPrefix, false)
		view.foldedFiles = foldedFrozenPaths(view.files, view.keyPrefix)
		view.foldedParents = foldedFrozenPaths(view.parents, view.keyPrefix)
	}
	return view, nil
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
	folded := make(map[string]string, len(paths))
	for path, physical := range paths {
		id := pathutil.GetCanonicalFileName(strings.TrimPrefix(path, prefix), false)
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
		return pathutil.GetCanonicalFileName(suffix, false)
	}
	// Fold before removing a differently-cased prefix: Unicode folding can
	// change byte lengths, including inside the prefix itself.
	folded := pathutil.GetCanonicalFileName(path, false)
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
			relative := pathutil.GetRelativePathFromDirectory(directory, path, pathutil.ComparePathsOptions{UseCaseSensitiveFileNames: true})
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

func (f *frozenFileSystem) FileExists(path tspath.RootedFilePath) bool {
	return f.FS.FileExists(pathutil.File(f.ioPath(path.AsString())))
}

func (f *frozenFileSystem) ReadFile(path tspath.RootedFilePath) (string, bool) {
	return f.FS.ReadFile(pathutil.File(f.ioPath(path.AsString())))
}

func (f *frozenFileSystem) SourceHasBOM(path string) bool {
	return utils.SourceHasBOM(f.FS, f.ioPath(path))
}

func (f *frozenFileSystem) DirectoryExists(path tspath.RootedDirectoryPath) bool {
	return f.FS.DirectoryExists(pathutil.Directory(f.ioPath(path.AsString())))
}

func (f *frozenFileSystem) GetAccessibleEntries(path tspath.RootedDirectoryPath) vfs.Entries {
	return f.FS.GetAccessibleEntries(pathutil.Directory(f.ioPath(path.AsString())))
}

func (f *frozenFileSystem) Stat(path tspath.RootedPath) vfs.FileInfo {
	return f.FS.Stat(pathutil.Path(f.ioPath(path.AsString())))
}

func (f *frozenFileSystem) Realpath(path tspath.RootedPath) tspath.RootedPath {
	physical, frozenFile := f.resolvePath(path.AsString())
	if frozenFile {
		return pathutil.Path(physical)
	}
	return f.FS.Realpath(pathutil.Path(physical))
}

func (f *frozenFileSystem) WriteFile(path tspath.RootedFilePath, data string) error {
	return f.FS.WriteFile(pathutil.File(f.ioPath(path.AsString())), data)
}

func (f *frozenFileSystem) AppendFile(path tspath.RootedFilePath, data string) error {
	return f.FS.AppendFile(pathutil.File(f.ioPath(path.AsString())), data)
}

func (f *frozenFileSystem) Remove(path tspath.RootedPath) error {
	return f.FS.Remove(pathutil.Path(f.ioPath(path.AsString())))
}

func (f *frozenFileSystem) Chtimes(path tspath.RootedPath, atime time.Time, mtime time.Time) error {
	return f.FS.Chtimes(pathutil.Path(f.ioPath(path.AsString())), atime, mtime)
}
