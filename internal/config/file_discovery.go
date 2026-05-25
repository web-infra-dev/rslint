package config

import (
	"io/fs"
	"path"
	"runtime"
	"sort"
	"strings"
	"sync"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// defaultExcludeDirs is a set of directory names always excluded from walking.
var defaultExcludeDirs = func() map[string]struct{} {
	m := make(map[string]struct{}, len(utils.DefaultExcludeDirNames))
	for _, name := range utils.DefaultExcludeDirNames {
		m[name] = struct{}{}
	}
	return m
}()

// DiscoverGapFiles scans the filesystem for "gap files" — files that match a config
// entry's `files` pattern but are not in any tsconfig Program. These files get a
// fallback Program (AST-only, no type info) and only run non-type-aware rules.
//
// Files pass through these filters:
//  1. Match at least one config entry's `files` pattern
//  2. Not in global ignores (directory-level and file-level)
//  3. Not already in programFiles (existing tsconfig Programs)
//  4. Pass GetConfigForFile (would actually receive lint rules)
//
// Walking model:
//   - A bounded worker pool (see walkPool) traverses the directory tree.
//     Total live goroutines at any moment is at most `workers`.
//   - Default workers = max(2, GOMAXPROCS); singleThreaded forces workers=1
//     for fully serial traversal (a knob users rely on for debugging,
//     reproducibility, and constrained environments).
//   - The vfsAdapter is constructed with followSymlinks=false, so symlinked
//     subdirectories are skipped entirely. This matches ESLint v10's
//     flat-config file walker, which uses @humanfs/node and recurses only
//     when Dirent.isDirectory() is true (Node returns false for symlinks).
//     It also avoids the otherwise scheduling-dependent "first writer wins"
//     non-determinism a parallel walker would introduce.
//
// When allowFiles/allowDirs are provided (CLI args), only files within scope.
//
// Returns:
//   - nil: no config entry has a `files` field → caller uses legacy tsconfig-only behavior
//   - []: `files` present but no gaps found
//   - [...]: gap files to create a fallback Program for (sorted lexically)
func DiscoverGapFiles(
	config RslintConfig,
	configDir string,
	fsys vfs.FS,
	programFiles map[string]struct{},
	allowFiles []string,
	allowDirs []string,
	singleThreaded bool,
) []string {
	// 1. Collect global ignore patterns and files patterns from config entries.
	globalIgnores := ExtractConfigIgnores(config)

	var allFilesPatterns []string
	hasFilesField := false
	for _, entry := range config {
		if len(entry.Files) > 0 {
			hasFilesField = true
			allFilesPatterns = append(allFilesPatterns, entry.Files...)
		}
	}

	// No entry has files → backward compat, signal caller to skip new logic.
	if !hasFilesField {
		return nil
	}

	// Deduplicate patterns.
	allFilesPatterns = deduplicate(allFilesPatterns)

	// Build allowFiles set for fast lookup.
	var allowFileSet map[string]struct{}
	if allowFiles != nil {
		allowFileSet = make(map[string]struct{}, len(allowFiles))
		for _, f := range allowFiles {
			allowFileSet[tspath.NormalizePath(f)] = struct{}{}
		}
	}

	// 2. Prepend default directory ignores so they are always active
	// regardless of user config.
	globalIgnores = append(utils.DefaultIgnoreDirGlobs(), globalIgnores...)

	// Use non-nil empty slice to distinguish "files field present, no gaps"
	// from "no files field" (nil).
	gapFiles := []string{}

	// Fast path: when only specific files are provided (e.g., lint-staged),
	// check each file directly instead of walking the entire filesystem.
	if allowFileSet != nil && allowDirs == nil {
		for f := range allowFileSet {
			if _, exists := programFiles[f]; exists {
				continue
			}
			if isFileIgnored(f, globalIgnores, configDir) {
				continue
			}
			if config.GetConfigForFile(f, configDir) == nil {
				continue
			}
			gapFiles = append(gapFiles, f)
		}
		gapFiles = deduplicate(gapFiles)
		// Map iteration order is non-deterministic; sort to match the walker
		// path and keep output stable across runs.
		sort.Strings(gapFiles)
		return gapFiles
	}

	// 3. Walk the tree with a bounded worker pool. Goroutine count is capped
	// at `workers`; symlinks are skipped (see vfsAdapter doc) so the result
	// set is deterministic regardless of scheduling.
	normalizedConfigDir := normalizeGlobPath(configDir)
	fsAdapter := &vfsAdapter{vfs: fsys, root: normalizedConfigDir}

	// Normalize patterns relative to configDir for matching.
	relativePatterns := make([]string, len(allFilesPatterns))
	for i, pattern := range allFilesPatterns {
		normalizedPattern := normalizeGlobPath(tspath.ResolvePath(configDir, pattern))
		relativePatterns[i] = strings.TrimPrefix(normalizedPattern, normalizedConfigDir+"/")
	}

	var (
		gapMu     sync.Mutex
		seen      sync.Map // map[string]struct{} — file dedupe
		dirIgnore sync.Map // map[string]bool — pattern check cache, write-once per path
	)

	// Defer the parallelism limit to GOMAXPROCS (Go's standard knob; aligned
	// with container CGroup CPU limits). Lower bound of 2 keeps the walker
	// useful on single-core CI runners. singleThreaded overrides to 1 for
	// fully serial traversal — a knob users depend on for reproducibility,
	// debugging, and constrained environments.
	workers := runtime.GOMAXPROCS(0)
	if workers < 2 {
		workers = 2
	}
	if singleThreaded {
		workers = 1
	}

	processFile := func(walkPath string) {
		// 1. pattern match
		matched := false
		for _, pattern := range relativePatterns {
			if ok, _ := doublestar.Match(pattern, walkPath); ok {
				matched = true
				break
			}
		}
		if !matched {
			return
		}

		fullPath := tspath.NormalizePath(path.Join(normalizedConfigDir, walkPath))

		// 2. already in tsconfig Programs
		if _, exists := programFiles[fullPath]; exists {
			return
		}
		// 3. file-level global ignores
		if isFileIgnored(fullPath, globalIgnores, configDir) {
			return
		}
		// 4. CLI scope
		if allowFileSet != nil || allowDirs != nil {
			inScope := false
			if allowFileSet != nil {
				if _, ok := allowFileSet[fullPath]; ok {
					inScope = true
				}
			}
			if !inScope && allowDirs != nil {
				for _, dir := range allowDirs {
					if tspath.StartsWithDirectory(fullPath, dir, true) {
						inScope = true
						break
					}
				}
			}
			if !inScope {
				return
			}
		}
		// 5. final config check
		if config.GetConfigForFile(fullPath, configDir) == nil {
			return
		}
		// 6. dedupe + append
		if _, loaded := seen.LoadOrStore(fullPath, struct{}{}); loaded {
			return
		}
		gapMu.Lock()
		gapFiles = append(gapFiles, fullPath)
		gapMu.Unlock()
	}

	work := func(walkPath string) []string {
		f, err := fsAdapter.Open(walkPath)
		if err != nil {
			return nil
		}
		rdf, ok := f.(fs.ReadDirFile)
		if !ok {
			f.Close()
			return nil
		}
		entries, _ := rdf.ReadDir(-1)
		f.Close()

		var childDirs []string
		for _, e := range entries {
			name := e.Name()
			if e.IsDir() {
				if _, excluded := defaultExcludeDirs[name]; excluded {
					continue
				}
				childPath := path.Join(walkPath, name)
				if cached, ok := dirIgnore.Load(childPath); ok {
					blocked, _ := cached.(bool) // dirIgnore only stores bool (set by Store below)
					if blocked {
						continue
					}
				} else {
					blocked := isDirPathBlocked(childPath, globalIgnores)
					dirIgnore.Store(childPath, blocked)
					if blocked {
						continue
					}
				}
				childDirs = append(childDirs, childPath)
			} else {
				processFile(path.Join(walkPath, name))
			}
		}
		return childDirs
	}

	pool := newWalkPool(workers)
	pool.submitMany([]string{"."})
	pool.run(work)

	sort.Strings(gapFiles)
	return gapFiles
}

// DiscoverGapFilesMultiConfig runs DiscoverGapFiles for each config in a
// monorepo config map and returns the union of all gap files.
//
// Configs are processed serially. Each DiscoverGapFiles invocation already
// uses an internal worker pool, so the total live goroutine count is bounded
// by the worker pool size, not by `len(configMap) × workers`. Cross-config
// parallelism can be added later if benchmarks justify it.
func DiscoverGapFilesMultiConfig(
	configMap map[string]RslintConfig,
	fsys vfs.FS,
	programFiles map[string]struct{},
	allowFiles []string,
	allowDirs []string,
	singleThreaded bool,
) []string {
	if len(configMap) == 0 {
		return nil
	}

	seen := make(map[string]struct{})
	var allGapFiles []string
	for configDir, cfg := range configMap {
		gapFiles := DiscoverGapFiles(cfg, configDir, fsys, programFiles, allowFiles, allowDirs, singleThreaded)
		for _, f := range gapFiles {
			if _, exists := seen[f]; !exists {
				seen[f] = struct{}{}
				allGapFiles = append(allGapFiles, f)
			}
		}
	}

	if len(allGapFiles) == 0 {
		return nil
	}
	sort.Strings(allGapFiles)
	return allGapFiles
}

// walkPool is a fixed-size worker pool with an unbounded internal LIFO queue,
// used by DiscoverGapFiles to walk directory trees with a bounded number of
// live goroutines. Properties:
//
//   - At most `workers` goroutines exist concurrently.
//   - submitMany never blocks (queue grows as needed; total memory ~ peak
//     queue size, bounded by FS branching × depth).
//   - run() returns once all submitted work and all transitively submitted
//     work have completed. Detection: when every worker is simultaneously
//     idle and the queue is empty.
//
// workers=1 degenerates to an effectively serial DFS-style traversal (LIFO
// pops the most recently submitted dir); the only goroutine is the worker
// itself. This is what --singleThreaded relies on.
type walkPool struct {
	mu      sync.Mutex
	cond    *sync.Cond
	queue   []string
	idle    int
	workers int
	done    bool
}

func newWalkPool(workers int) *walkPool {
	if workers < 1 {
		workers = 1
	}
	p := &walkPool{workers: workers}
	p.cond = sync.NewCond(&p.mu)
	return p
}

// submitMany appends dirs to the queue and wakes idle workers. Safe to call
// from any goroutine.
func (p *walkPool) submitMany(dirs []string) {
	if len(dirs) == 0 {
		return
	}
	p.mu.Lock()
	p.queue = append(p.queue, dirs...)
	p.mu.Unlock()
	if len(dirs) == 1 {
		p.cond.Signal()
	} else {
		p.cond.Broadcast()
	}
}

// take pops a job from the queue, blocking if empty. Returns ("", false)
// only when the queue is empty AND every worker is simultaneously idle —
// which means no more work will ever appear.
func (p *walkPool) take() (string, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for {
		if len(p.queue) > 0 {
			n := len(p.queue) - 1
			dir := p.queue[n]
			p.queue = p.queue[:n]
			return dir, true
		}
		p.idle++
		if p.idle == p.workers {
			p.done = true
			p.cond.Broadcast()
			return "", false
		}
		p.cond.Wait()
		if p.done {
			return "", false
		}
		p.idle--
	}
}

// run drives the worker pool. Each worker pulls jobs and calls work(dir),
// which returns the child directories to enqueue. Returns when all reachable
// work is processed.
//
// When workers == 1, runs the loop on the calling goroutine directly — no
// goroutines are spawned at all. This is what --singleThreaded relies on:
// callers can rely on the Go side spawning no extra goroutines.
func (p *walkPool) run(work func(string) []string) {
	if p.workers == 1 {
		for {
			dir, ok := p.take()
			if !ok {
				return
			}
			p.submitMany(work(dir))
		}
	}
	var wg sync.WaitGroup
	for range p.workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				dir, ok := p.take()
				if !ok {
					return
				}
				p.submitMany(work(dir))
			}
		}()
	}
	wg.Wait()
}

// deduplicate returns a copy of the input slice with duplicates removed, preserving order.
func deduplicate(items []string) []string {
	if len(items) == 0 {
		return items
	}
	seen := make(map[string]struct{}, len(items))
	result := make([]string, 0, len(items))
	for _, item := range items {
		if _, exists := seen[item]; !exists {
			seen[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}
