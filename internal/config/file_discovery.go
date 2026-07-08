package config

import (
	"io/fs"
	"path"
	"runtime"
	"sort"
	"strings"
	"sync"

	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// DefaultLintFileExtensions are the file extensions rslint discovers when a
// config entry omits `files`. This intentionally extends ESLint's default
// .js/.mjs/.cjs set with JSX and TypeScript-family files.
var DefaultLintFileExtensions = []string{".js", ".mjs", ".cjs", ".jsx", ".ts", ".tsx", ".mts", ".cts"}

var defaultLintFileExtensionSet = func() map[string]struct{} {
	m := make(map[string]struct{}, len(DefaultLintFileExtensions))
	for _, ext := range DefaultLintFileExtensions {
		m[ext] = struct{}{}
	}
	return m
}()

func defaultLintFilePatterns() []string {
	patterns := make([]string, 0, len(DefaultLintFileExtensions))
	for _, ext := range DefaultLintFileExtensions {
		patterns = append(patterns, "**/*"+ext)
	}
	return patterns
}

// IsSupportedLintFile reports whether rslint can parse and lint this path.
func IsSupportedLintFile(filePath string) bool {
	_, ok := defaultLintFileExtensionSet[strings.ToLower(path.Ext(filePath))]
	return ok
}

// defaultExcludeDirs is a set of directory names always excluded from walking.
var defaultExcludeDirs = func() map[string]struct{} {
	m := make(map[string]struct{}, len(utils.DefaultExcludeDirNames))
	for _, name := range utils.DefaultExcludeDirNames {
		m[name] = struct{}{}
	}
	return m
}()

// DiscoverLintFiles resolves the lint target set for one config directory.
// Target selection is independent from TypeScript Program membership:
//
//   - CLI/API files and directories first constrain the search space.
//   - Non-global config entries then contribute `files` patterns; an omitted
//     files field contributes rslint's default lintable extensions.
//   - Global ignores, including injected .gitignore entries, remove files.
//   - Entry-level ignores are honored through GetConfigForFile for discovered
//     files. Explicit file targets bypass `files` and entry-level ignores for
//     target selection, but still use GetConfigForFile later for rule selection
//     and can therefore be counted as 0-rule lint results.
//   - Explicit file targets are retained even when they do not match any
//     config entry's files patterns, matching ESLint's empty-result behavior.
//
// Returned paths are absolute, normalized, deduplicated, and sorted.
func DiscoverLintFiles(
	config RslintConfig,
	configDir string,
	fsys vfs.FS,
	allowFiles []string,
	allowDirs []string,
	singleThreaded bool,
) []string {
	globalIgnores := ExtractConfigIgnores(config)
	globalIgnores = append(ParseIgnorePatterns(utils.DefaultIgnoreDirGlobs()), globalIgnores...)

	filesPatterns := collectLintFilePatterns(config)

	var allowFileSet map[string]struct{}
	if allowFiles != nil {
		allowFileSet = make(map[string]struct{}, len(allowFiles))
		for _, f := range allowFiles {
			allowFileSet[tspath.NormalizePath(f)] = struct{}{}
		}
	}

	targetFiles := []string{}
	seenTargets := make(map[string]struct{})
	addTarget := func(filePath string) {
		if _, seen := seenTargets[filePath]; seen {
			return
		}
		seenTargets[filePath] = struct{}{}
		targetFiles = append(targetFiles, filePath)
	}
	isGloballyIgnored := func(filePath string) bool {
		return isDirBlockedByIgnores(filePath, globalIgnores, configDir) ||
			isFileIgnored(filePath, globalIgnores, configDir)
	}

	includeExplicitFile := func(filePath string) bool {
		if !IsSupportedLintFile(filePath) {
			return false
		}
		if fsys != nil && !fsys.FileExists(filePath) {
			return false
		}
		return !isGloballyIgnored(filePath)
	}

	includeDiscoveredFile := func(filePath string) bool {
		if !IsSupportedLintFile(filePath) {
			return false
		}
		if len(filesPatterns) == 0 || !isFileMatched(filePath, filesPatterns, configDir) {
			return false
		}
		if isGloballyIgnored(filePath) {
			return false
		}
		return config.GetConfigForFile(filePath, configDir) != nil
	}

	addExplicitTargets := func() {
		for f := range allowFileSet {
			if includeExplicitFile(f) {
				addTarget(f)
			}
		}
	}

	// Fast path for explicit file-only invocations, e.g. lint-staged.
	if allowFileSet != nil && allowDirs == nil {
		addExplicitTargets()
		sort.Strings(targetFiles)
		return targetFiles
	}

	normalizedConfigDir := normalizeGlobPath(configDir)
	fsAdapter := &vfsAdapter{vfs: fsys, root: normalizedConfigDir}

	var (
		targetMu  sync.Mutex
		dirIgnore sync.Map // map[string]bool — pattern check cache
	)

	neg := buildNegReach(globalIgnores)

	workers := runtime.GOMAXPROCS(0)
	if workers < 2 {
		workers = 2
	}
	if singleThreaded {
		workers = 1
	}

	processFile := func(walkPath string) {
		fullPath := tspath.NormalizePath(path.Join(normalizedConfigDir, walkPath))

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

		if !includeDiscoveredFile(fullPath) {
			return
		}

		targetMu.Lock()
		addTarget(fullPath)
		targetMu.Unlock()
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
					blocked, _ := cached.(bool)
					if blocked {
						continue
					}
				} else {
					blocked := canPruneDir(childPath, globalIgnores, neg)
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

	if allowFileSet != nil {
		addExplicitTargets()
	}

	sort.Strings(targetFiles)
	return targetFiles
}

func collectLintFilePatterns(config RslintConfig) []string {
	var patterns []string
	for _, entry := range config {
		if isGlobalIgnoreEntry(entry) {
			continue
		}
		if len(entry.Files) > 0 {
			patterns = append(patterns, entry.Files...)
		} else {
			patterns = append(patterns, defaultLintFilePatterns()...)
		}
	}
	return deduplicate(patterns)
}

// DiscoverLintFilesMultiConfig resolves lint targets across a config map.
func DiscoverLintFilesMultiConfig(
	configMap map[string]RslintConfig,
	fsys vfs.FS,
	allowFiles []string,
	allowDirs []string,
	singleThreaded bool,
) []string {
	if len(configMap) == 0 {
		return nil
	}

	seen := make(map[string]struct{})
	var allTargets []string
	for configDir, cfg := range configMap {
		targets := DiscoverLintFiles(cfg, configDir, fsys, allowFiles, allowDirs, singleThreaded)
		for _, f := range targets {
			ownerDir, _ := FindNearestConfig(f, configMap)
			if ownerDir != configDir {
				continue
			}
			if _, exists := seen[f]; !exists {
				seen[f] = struct{}{}
				allTargets = append(allTargets, f)
			}
		}
	}
	sort.Strings(allTargets)
	return allTargets
}

// DiscoverGapFiles returns resolved lint targets that are absent from existing
// Programs. The filesystem walk and config/default-files matching are owned by
// DiscoverLintFiles; this helper only subtracts programFiles for callers that
// need an AST-only fallback Program.
//
// Files pass through these filters in DiscoverLintFiles:
//  1. Inside CLI/API file or directory scope
//  2. Selected by config `files` patterns or default lintable extensions;
//     explicit file targets may pass this step to produce an empty result
//  3. Not in global ignores or .gitignore-injected ignores
//  4. Not already in programFiles
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
//   - []: no gaps found
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
	gapFiles := []string{}
	targetFiles := DiscoverLintFiles(config, configDir, fsys, allowFiles, allowDirs, singleThreaded)
	for _, fullPath := range targetFiles {
		if _, exists := programFiles[fullPath]; exists {
			continue
		}
		gapFiles = append(gapFiles, fullPath)
	}
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
			ownerDir, _ := FindNearestConfig(f, configMap)
			if ownerDir != configDir {
				continue
			}
			if _, exists := seen[f]; !exists {
				seen[f] = struct{}{}
				allGapFiles = append(allGapFiles, f)
			}
		}
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
