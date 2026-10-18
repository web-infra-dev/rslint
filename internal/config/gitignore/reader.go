package gitignore

import (
	"sort"
	"strings"
	"sync"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
)

// Source is one observed .gitignore in physical repository coordinates.
type Source struct {
	Directory string
	Patterns  []Pattern
}

// Snapshot owns the Git observations of one request. Empty roots and missing
// sources remain meaningful: they establish boundaries and source coverage.
type Snapshot struct {
	Roots              []string
	Sources            []Source
	CoveredDirectories []string
	UseCaseSensitive   bool
}

// Reader observes Git metadata once per request. It never resolves paths:
// callers pass the physical directory identities frozen at their boundary.
// Holding the mutex through each small VFS read also coalesces concurrent misses.
type Reader struct {
	fs               vfs.FS
	useCaseSensitive bool
	mu               sync.Mutex
	markers          map[tspath.Path]bool
	sources          map[tspath.Path]Source
	roots            map[tspath.Path]string
	covered          map[tspath.Path]string
	ancestors        map[tspath.Path]Cursor
}

func NewReader(fsys vfs.FS) *Reader {
	return &Reader{
		fs: fsys, useCaseSensitive: fsys == nil || fsys.UseCaseSensitiveFileNames(),
		markers: make(map[tspath.Path]bool), sources: make(map[tspath.Path]Source),
		roots: make(map[tspath.Path]string), covered: make(map[tspath.Path]string),
		ancestors: make(map[tspath.Path]Cursor),
	}
}

func (reader *Reader) pathID(path string) tspath.Path {
	return DirectoryPathID(path, reader.useCaseSensitive)
}

func (reader *Reader) hasMarker(directory string) bool {
	reader.mu.Lock()
	defer reader.mu.Unlock()
	id := reader.pathID(directory)
	if exists, ok := reader.markers[id]; ok {
		return exists
	}
	marker := tspath.CombinePaths(directory, ".git")
	exists := reader.fs != nil && (reader.fs.DirectoryExists(marker) || reader.fs.FileExists(marker))
	reader.markers[id] = exists
	return exists
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

// Enter establishes a new repository boundary before consulting ancestor
// rules. A nested repository can therefore reset raw source reachability.
func (reader *Reader) Enter(parent Cursor, physicalDir string) (Cursor, bool) {
	physicalDir = normalizeGlobPath(physicalDir)
	if reader.hasMarker(physicalDir) {
		return reader.newRoot(physicalDir), false
	}
	return parent.Enter(physicalDir)
}

// Observe freezes local source availability, including a proven raw traversal
// block. Authored config negations never make a hidden Git source readable.
func (reader *Reader) Observe(cursor Cursor, physicalDir string) Cursor {
	physicalDir = normalizeGlobPath(physicalDir)
	id := reader.pathID(physicalDir)
	reader.mu.Lock()
	reader.covered[id] = physicalDir
	if !cursor.SourceReachable() {
		reader.mu.Unlock()
		return cursor
	}
	next := cursor
	var patterns []Pattern
	if source, exists := reader.sources[id]; exists {
		patterns = source.Patterns
		next = cursor.appendPatterns(patterns)
	} else if reader.fs != nil {
		if content, exists := reader.fs.ReadFile(tspath.CombinePaths(physicalDir, ".gitignore")); exists {
			next, patterns = cursor.AppendSourcePatterns(physicalDir, content)
		}
	}
	reader.sources[id] = Source{Directory: physicalDir, Patterns: patterns}
	reader.mu.Unlock()
	return next
}

// Snapshot returns detached, deterministically ordered data. Later reader
// observations cannot change an already compiled policy.
func (reader *Reader) Snapshot() Snapshot {
	reader.mu.Lock()
	defer reader.mu.Unlock()
	snapshot := Snapshot{UseCaseSensitive: reader.useCaseSensitive}
	for _, root := range reader.roots {
		snapshot.Roots = append(snapshot.Roots, root)
	}
	for _, source := range reader.sources {
		source.Patterns = append([]Pattern(nil), source.Patterns...)
		snapshot.Sources = append(snapshot.Sources, source)
	}
	for _, directory := range reader.covered {
		snapshot.CoveredDirectories = append(snapshot.CoveredDirectories, directory)
	}
	sort.Strings(snapshot.Roots)
	sort.Strings(snapshot.CoveredDirectories)
	sort.Slice(snapshot.Sources, func(i, j int) bool { return snapshot.Sources[i].Directory < snapshot.Sources[j].Directory })
	return snapshot
}

// ParentDirectory stops at filesystem volume boundaries, including UNC shares.
func ParentDirectory(directory string) string { return parentDir(directory) }

// DirectoryPathID makes root spellings with or without a trailing slash share
// one identity while preserving filesystem and drive roots.
func DirectoryPathID(directory string, useCaseSensitive bool) tspath.Path {
	directory = normalizeGlobPath(directory)
	if directory != "/" && (len(directory) != 3 || directory[1] != ':') {
		directory = strings.TrimSuffix(directory, "/")
	}
	return tspath.ToPath(directory, "", useCaseSensitive)
}
