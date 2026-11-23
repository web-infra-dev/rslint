package gitignore

import (
	"os"
	"os/exec"
	"sort"
	"strings"
	"sync"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// StoredEntryNamer reports the name stored in the parent directory without
// following the entry's final symlink.
type StoredEntryNamer interface {
	StoredEntryName(path string) (string, error)
}

// Source is one observed .gitignore in physical repository coordinates.
type Source struct {
	Directory       string
	Patterns        []Pattern
	CaseInsensitive bool
}

// Snapshot owns the Git observations of one request. Empty roots and missing
// sources remain meaningful: they establish boundaries and source coverage.
type Snapshot struct {
	Roots                   []string
	Sources                 []Source
	CoveredDirectories      []string
	BlockedDirectoryEntries []string
}

// Reader observes Git metadata once per request. It never resolves paths:
// callers pass the physical directory identities frozen at their boundary.
// Same-directory observations share one initializer; unrelated directories may
// perform IO concurrently. The mutex protects publication, never filesystem IO.
type Reader struct {
	fs             vfs.FS
	mu             sync.Mutex
	markers        map[tspath.Path]func() *repositoryMarker
	sources        map[tspath.Path]func() Source
	roots          map[tspath.Path]string
	covered        map[tspath.Path]string
	blockedEntries map[tspath.Path]string
	ancestors      map[tspath.Path]Cursor

	// BeforeConfigRead subscribes to a native repository's configuration files
	// before reading its ignore-case setting. Nil paths mean Git could not
	// identify the current files. It can be called concurrently.
	BeforeConfigRead func(paths []string) error
}

type repositoryMarker struct {
	caseInsensitive func() bool // Probed only if the repository contributes a rule.
}

func NewReader(fsys vfs.FS) *Reader {
	return &Reader{
		fs:      fsys,
		markers: make(map[tspath.Path]func() *repositoryMarker), sources: make(map[tspath.Path]func() Source),
		roots: make(map[tspath.Path]string), covered: make(map[tspath.Path]string),
		blockedEntries: make(map[tspath.Path]string),
		ancestors:      make(map[tspath.Path]Cursor),
	}
}

func (reader *Reader) marker(directory string) *repositoryMarker {
	id := DirectoryPathID(directory)
	reader.mu.Lock()
	load, ok := reader.markers[id]
	if !ok {
		marker := tspath.CombinePaths(directory, ".git")
		load = sync.OnceValue(func() *repositoryMarker {
			if reader.fs == nil || (!reader.fs.DirectoryExists(marker) && !reader.fs.FileExists(marker)) {
				return nil
			}
			return &repositoryMarker{sync.OnceValue(func() bool { return reader.repositoryCaseInsensitive(directory, marker) })}
		})
		reader.markers[id] = load
	}
	reader.mu.Unlock()
	return load()
}

func (reader *Reader) repositoryCaseInsensitive(directory, marker string) bool {
	// A linked worktree can have its config on another volume. A repository
	// can also override the mode chosen by git init.
	// Let Git resolve its own config only for a marker on the host filesystem.
	info := reader.fs.Stat(marker)
	if os.SameFile(info, info) && (info.Mode().IsRegular() && info.Size() > 0 ||
		info.IsDir() && reader.fs.FileExists(tspath.CombinePaths(marker, "HEAD"))) {
		if native, err := os.Stat(marker); err == nil && os.SameFile(info, native) {
			if reader.BeforeConfigRead != nil {
				if err := reader.watchRepositoryConfig(directory, marker); err != nil {
					return false
				}
			}
			if value, ok := repositoryGitOutput(directory, "--git-dir="+marker, "config", "--bool", "--default=false", "--get", "core.ignoreCase"); ok {
				return strings.TrimSpace(value) == "true"
			}
		}
	}
	return reader.caseInsensitive(directory, ".git")
}

func repositoryGitOutput(directory string, args ...string) (string, bool) {
	cmd := exec.Command("git", args...)
	cmd.Dir = directory
	value, err := cmd.Output()
	return strings.TrimSuffix(string(value), "\n"), err == nil
}

func (reader *Reader) watchRepositoryConfig(directory, marker string) error {
	var entries []string
	for _, file := range []struct{ option, name string }{{"--git-common-dir", "config"}, {"--absolute-git-dir", "config.worktree"}} {
		dir, ok := repositoryGitOutput(directory, "--git-dir="+marker, "rev-parse", "--path-format=absolute", file.option)
		if !ok {
			return reader.BeforeConfigRead(nil)
		}
		entries = append(entries, tspath.CombinePaths(tspath.NormalizePath(dir), file.name))
	}
	if err := reader.BeforeConfigRead(entries); err != nil {
		return err
	}
	// Subscribe to changes of the link before observing its current target.
	var targets []string
	for _, entry := range entries {
		if target := reader.fs.Realpath(entry); target != "" && target != entry {
			targets = append(targets, tspath.NormalizePath(target))
		}
	}
	if len(targets) != 0 {
		return reader.BeforeConfigRead(targets)
	}
	return nil
}

func (reader *Reader) newRoot(directory string) Cursor {
	reader.mu.Lock()
	reader.roots[DirectoryPathID(directory)] = directory
	reader.mu.Unlock()
	return NewCursor(directory, true)
}

// The observed metadata entry is a read-only case probe on its own mount.
// A case-sensitive directory can contain both spellings, even as hard links.
func (reader *Reader) caseInsensitive(directory, name string) bool {
	path := tspath.CombinePaths(directory, name)
	alternate := tspath.CombinePaths(directory, strings.ToUpper(name))
	if !reader.fs.FileExists(alternate) && !reader.fs.DirectoryExists(alternate) {
		// Some virtual sources can be read without a corresponding file entry.
		if name == ".gitignore" && !reader.fs.FileExists(path) {
			return !reader.fs.UseCaseSensitiveFileNames()
		}
		return false
	}
	if native, ok := reader.fs.(StoredEntryNamer); ok {
		left, _ := native.StoredEntryName(path)
		right, _ := native.StoredEntryName(alternate)
		if left != "" && right != "" {
			return left == right
		}
	}
	left, right := reader.fs.Stat(path), reader.fs.Stat(alternate)
	// SameFile can only compare native FileInfo. Distinct native objects
	// already prove independent names without reading their directory.
	if os.SameFile(left, left) && os.SameFile(right, right) && !os.SameFile(left, right) {
		return false
	}
	entries := reader.fs.GetAccessibleEntries(directory)
	count := 0
	for _, group := range [][]string{entries.Files, entries.Directories} {
		for _, entry := range group {
			if strings.EqualFold(entry, name) {
				count++
			}
		}
	}
	if count == 0 {
		return !reader.fs.UseCaseSensitiveFileNames()
	}
	return count == 1
}

// Ancestors finds the nearest .git directory or file, falling back to the
// filesystem volume root. It observes only the chain above physicalDir; the
// caller decides whether to Observe physicalDir itself.
func (reader *Reader) Ancestors(physicalDir string) Cursor {
	physicalDir = normalizeGlobPath(physicalDir)
	id := DirectoryPathID(physicalDir)
	reader.mu.Lock()
	if cursor, exists := reader.ancestors[id]; exists {
		reader.mu.Unlock()
		return cursor
	}
	reader.mu.Unlock()
	root := reader.RootDirectory(physicalDir)
	chain := []string{physicalDir}
	for current := physicalDir; DirectoryPathID(current) != DirectoryPathID(root); {
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
	for reader.marker(current) == nil {
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
		reader.blockedEntries[DirectoryPathID(physicalDir)] = physicalDir
		reader.mu.Unlock()
	}
	if reader.marker(physicalDir) != nil {
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
	id := DirectoryPathID(physicalDir)
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
				if content, exists := reader.fs.ReadFile(tspath.CombinePaths(physicalDir, ".gitignore")); exists {
					source.Patterns = convertGitignoreToPatterns(content, relative)
					if len(source.Patterns) > 0 {
						if marker := reader.marker(cursor.rootDir); marker != nil {
							source.CaseInsensitive = marker.caseInsensitive()
						} else {
							source.CaseInsensitive = reader.caseInsensitive(physicalDir, ".gitignore")
						}
					}
				}
			}
			return source
		})
		reader.sources[id] = load
	}
	reader.mu.Unlock()
	source := load()
	return cursor.appendPatterns(source.Patterns, source.CaseInsensitive)
}

// Snapshot returns detached, deterministically ordered data. Later reader
// observations cannot change an already compiled policy. It waits for source
// observations published at the snapshot boundary without holding the mutex.
func (reader *Reader) Snapshot() Snapshot {
	reader.mu.Lock()
	var snapshot Snapshot
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

// DirectoryPathID keys a frozen physical directory, preserving its entries
// even when a different volume in this process is case-insensitive.
func DirectoryPathID(directory string) tspath.Path {
	directory = normalizeGlobPath(directory)
	if directory != "/" && (len(directory) != 3 || directory[1] != ':') {
		directory = strings.TrimSuffix(directory, "/")
	}
	return tspath.ToPath(utils.NormalizeAbsoluteDrive(directory), "", true)
}
