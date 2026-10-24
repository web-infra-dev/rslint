package gitignore

import (
	"sort"
	"strings"
	"sync"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/pathutil"
)

// Source is one observed .gitignore in physical repository coordinates.
type Source struct {
	Directory string
	Patterns  []Pattern
}

// Snapshot owns the Git observations of one request. Empty roots and missing
// sources remain meaningful: they establish boundaries and source coverage.
type Snapshot struct {
	Roots                   []string
	Sources                 []Source
	CoveredDirectories      []string
	BlockedDirectoryEntries []string
	UseCaseSensitive        bool
}

// Reader observes Git metadata once per request. It never resolves paths:
// callers pass the physical directory identities frozen at their boundary.
// Same-directory observations share one initializer; unrelated directories may
// perform IO concurrently. The mutex protects publication, never filesystem IO.
type Reader struct {
	fs               vfs.FS
	useCaseSensitive bool
	mu               sync.Mutex
	markers          map[tspath.PathKey]func() bool
	sources          map[tspath.PathKey]func() Source
	roots            map[tspath.PathKey]string
	covered          map[tspath.PathKey]string
	blockedEntries   map[tspath.PathKey]string
	ancestors        map[tspath.PathKey]Cursor
}

func NewReader(fsys vfs.FS) *Reader {
	return &Reader{
		fs: fsys, useCaseSensitive: fsys == nil || fsys.CaseSensitivity().IsCaseSensitive(),
		markers: make(map[tspath.PathKey]func() bool), sources: make(map[tspath.PathKey]func() Source),
		roots: make(map[tspath.PathKey]string), covered: make(map[tspath.PathKey]string),
		blockedEntries: make(map[tspath.PathKey]string),
		ancestors:      make(map[tspath.PathKey]Cursor),
	}
}

func (reader *Reader) pathID(path string) tspath.PathKey {
	return DirectoryPathID(path, reader.useCaseSensitive)
}

func (reader *Reader) hasMarker(directory string) bool {
	id := reader.pathID(directory)
	reader.mu.Lock()
	load, ok := reader.markers[id]
	if !ok {
		marker := tspath.CombinePaths(directory, ".git")
		load = sync.OnceValue(func() bool {
			return reader.fs != nil && (reader.fs.DirectoryExists(pathutil.Directory(marker)) || reader.fs.FileExists(pathutil.File(marker)))
		})
		reader.markers[id] = load
	}
	reader.mu.Unlock()
	return load()
}

func (reader *Reader) newRoot(directory string) Cursor {
	reader.mu.Lock()
	reader.roots[reader.pathID(directory)] = directory
	reader.mu.Unlock()
	return NewCursor(directory, reader.useCaseSensitive)
}

// Ancestors finds the nearest .git directory or file, falling back to the
// filesystem volume root. It observes only the chain above physicalDir; the
// caller decides whether to Observe physicalDir itself.
func (reader *Reader) Ancestors(physicalDir string) Cursor {
	physicalDir = normalizeGlobPath(physicalDir)
	id := reader.pathID(physicalDir)
	reader.mu.Lock()
	if cursor, exists := reader.ancestors[id]; exists {
		reader.mu.Unlock()
		return cursor
	}
	reader.mu.Unlock()
	root := reader.RootDirectory(physicalDir)
	chain := []string{physicalDir}
	for current := physicalDir; reader.pathID(current) != reader.pathID(root); {
		parent := parentDir(current)
		if parent == "" || parent == current {
			break
		}
		chain = append(chain, parent)
		current = parent
	}
	cursor := reader.newRoot(chain[len(chain)-1])
	for index := len(chain) - 1; index >= 0; index-- {
		if index < len(chain)-1 {
			cursor, _ = reader.Enter(cursor, chain[index])
		}
		if index > 0 {
			cursor = reader.Observe(cursor, chain[index])
		}
	}
	reader.mu.Lock()
	reader.ancestors[id] = cursor
	reader.mu.Unlock()
	return cursor
}

// RootDirectory locates a boundary without observing any .gitignore source.
// Config discovery uses this while independently searching for authored
// configuration above or outside the Git chain.
func (reader *Reader) RootDirectory(physicalDir string) string {
	current := normalizeGlobPath(physicalDir)
	for !reader.hasMarker(current) {
		parent := parentDir(current)
		if parent == "" || parent == current {
			break
		}
		current = parent
	}
	reader.newRoot(current)
	return current
}

// Enter checks the directory entry against the current repository's rules.
// The returned cursor describes the interior repository; blocked separately
// preserves the outer admission decision, which callers must honor before IO.
func (reader *Reader) Enter(parent Cursor, physicalDir string) (Cursor, bool) {
	physicalDir = normalizeGlobPath(physicalDir)
	next, blocked := parent.Enter(physicalDir)
	if blocked {
		reader.mu.Lock()
		reader.blockedEntries[reader.pathID(physicalDir)] = physicalDir
		reader.mu.Unlock()
	}
	if reader.hasMarker(physicalDir) {
		return reader.newRoot(physicalDir), blocked
	}
	return next, blocked
}

// Observe freezes local source availability, including a proven raw traversal
// block. Authored config negations never make a hidden Git source readable.
// Cursors come from this reader's frozen boundary walk, so one physical source
// always has the same repository-root coordinate, even on overlapping routes.
func (reader *Reader) Observe(cursor Cursor, physicalDir string) Cursor {
	physicalDir = normalizeGlobPath(physicalDir)
	id := reader.pathID(physicalDir)
	reader.mu.Lock()
	reader.covered[id] = physicalDir
	if !cursor.SourceReachable() {
		reader.mu.Unlock()
		return cursor
	}
	load, ok := reader.sources[id]
	if !ok {
		relative, within := relativeDir(cursor.rootDir, physicalDir, cursor.useCaseSensitive)
		if !within {
			reader.mu.Unlock()
			cursor.sourceReachable = false
			return cursor
		}
		load = sync.OnceValue(func() Source {
			source := Source{Directory: physicalDir}
			if reader.fs != nil {
				if content, exists := reader.fs.ReadFile(pathutil.File(tspath.CombinePaths(physicalDir, ".gitignore"))); exists {
					source.Patterns = convertGitignoreToPatterns(content, relative)
				}
			}
			return source
		})
		reader.sources[id] = load
	}
	reader.mu.Unlock()
	return cursor.appendPatterns(load().Patterns)
}

// Snapshot returns detached, deterministically ordered data. Later reader
// observations cannot change an already compiled policy. It waits for source
// observations published at the snapshot boundary without holding the mutex.
func (reader *Reader) Snapshot() Snapshot {
	reader.mu.Lock()
	snapshot := Snapshot{UseCaseSensitive: reader.useCaseSensitive}
	for _, root := range reader.roots {
		snapshot.Roots = append(snapshot.Roots, root)
	}
	loads := make([]func() Source, 0, len(reader.sources))
	for _, load := range reader.sources {
		loads = append(loads, load)
	}
	for _, directory := range reader.covered {
		snapshot.CoveredDirectories = append(snapshot.CoveredDirectories, directory)
	}
	for _, directory := range reader.blockedEntries {
		snapshot.BlockedDirectoryEntries = append(snapshot.BlockedDirectoryEntries, directory)
	}
	reader.mu.Unlock()
	for _, load := range loads {
		source := load()
		source.Patterns = append([]Pattern(nil), source.Patterns...)
		snapshot.Sources = append(snapshot.Sources, source)
	}
	sort.Strings(snapshot.Roots)
	sort.Strings(snapshot.CoveredDirectories)
	sort.Strings(snapshot.BlockedDirectoryEntries)
	sort.Slice(snapshot.Sources, func(i, j int) bool { return snapshot.Sources[i].Directory < snapshot.Sources[j].Directory })
	return snapshot
}

// ParentDirectory stops at filesystem volume boundaries, including UNC shares.
func ParentDirectory(directory string) string { return parentDir(directory) }

// DirectoryPathID makes root spellings with or without a trailing slash share
// one identity while preserving filesystem and drive roots.
func DirectoryPathID(directory string, useCaseSensitive bool) tspath.PathKey {
	directory = normalizeGlobPath(directory)
	if directory != "/" && (len(directory) != 3 || directory[1] != ':') {
		directory = strings.TrimSuffix(directory, "/")
	}
	return pathutil.ToPath(utils.NormalizeAbsoluteDrive(directory), "", useCaseSensitive)
}
