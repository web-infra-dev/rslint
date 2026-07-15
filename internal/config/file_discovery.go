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

// IsSupportedLintFile reports whether rslint can parse and lint this path.
func IsSupportedLintFile(filePath string) bool {
	_, ok := defaultLintFileExtensionSet[strings.ToLower(path.Ext(filePath))]
	return ok
}

func isDefaultLintFile(filePath string) bool {
	_, ok := defaultLintFileExtensionSet[path.Ext(filePath)]
	return ok
}

// isFileSelectedByConfig reports whether the config itself selects filePath.
// The implicit default baseline is always present. An explicit `files` entry
// extends it only for paths not excluded by that same entry's `ignores`.
// Another matching entry or the default baseline may still select the path.
func isFileSelectedByConfig(config RslintConfig, filePath string, configDir string) bool {
	return NewFileSelectorMatcher(config, configDir).Selects(filePath)
}

// DiscoverLintFiles resolves the lint target set for one config directory.
// Target selection is independent from TypeScript Program membership:
//
//   - CLI/API files and directories first constrain the search space.
//   - Rslint's default lintable extensions are always selected. Non-global
//     config entries contribute additional `files` patterns.
//   - Global ignores, including injected .gitignore entries, remove files.
//   - Entry-level ignores prevent that entry from selecting or configuring a
//     file. They do not remove a file selected by the default baseline or a
//     different entry.
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
	targets := discoverLintTargetsWithStopDirs(config, configDir, fsys, allowFiles, allowDirs, nil, singleThreaded)
	files := make([]string, 0, len(targets))
	for _, target := range targets {
		files = append(files, target.Path)
	}
	return files
}

// DiscoverLintTargets is the identity-preserving form of DiscoverLintFiles.
// CanonicalPath is derived without a per-file realpath call for regular files
// reached by directory traversal. Explicit directory aliases are resolved once
// and projected onto their descendants; explicit files and file symlinks are
// resolved individually because their physical target cannot be inferred from
// the containing directory.
func DiscoverLintTargets(
	config RslintConfig,
	configDir string,
	fsys vfs.FS,
	allowFiles []string,
	allowDirs []string,
	singleThreaded bool,
) []DiscoveredLintTarget {
	return discoverLintTargetsWithStopDirs(config, configDir, fsys, allowFiles, allowDirs, nil, singleThreaded)
}

func discoverLintTargetsWithStopDirs(
	config RslintConfig,
	configDir string,
	fsys vfs.FS,
	allowFiles []string,
	allowDirs []string,
	stopDirs []string,
	singleThreaded bool,
) []DiscoveredLintTarget {
	effectiveConfig := WithDefaultGlobalIgnores(config)
	useCaseSensitive := true
	if fsys != nil {
		useCaseSensitive = fsys.UseCaseSensitiveFileNames()
	}
	comparisonKey := func(filePath string) string {
		return string(tspath.ToPath(tspath.NormalizePath(filePath), "", true))
	}
	configDir = tspath.NormalizePath(configDir)
	configMatchDir := configDir
	if fsys != nil {
		if realPath := fsys.Realpath(configDir); realPath != "" {
			configMatchDir = tspath.NormalizePath(realPath)
		}
	}
	configPathForMatching := func(filePath string) string {
		return resolveConfigPathSpace(
			tspath.NormalizePath(filePath),
			"",
			configDir,
			configMatchDir,
			fsys,
		)
	}
	resolvedAllowDirs := resolveAllowedDirectories(allowDirs, fsys)

	globalIgnoreMatcher := NewGlobalIgnoreMatcher(effectiveConfig, configDir, fsys)
	selectorMatcher := NewFileSelectorMatcher(config, configMatchDir)

	var allowFileSet map[string]string
	if allowFiles != nil {
		allowFileSet = make(map[string]string, len(allowFiles))
		for _, f := range allowFiles {
			normalized := tspath.NormalizePath(f)
			key := comparisonKey(normalized)
			if _, exists := allowFileSet[key]; !exists {
				allowFileSet[key] = normalized
			}
		}
	}

	targetFiles := []DiscoveredLintTarget{}
	seenTargets := make(map[string]struct{})
	addTarget := func(filePath string, canonicalPath string) {
		filePath = tspath.NormalizePath(filePath)
		if canonicalPath == "" {
			canonicalPath = filePath
		} else {
			canonicalPath = tspath.NormalizePath(canonicalPath)
		}
		key := comparisonKey(filePath)
		if _, seen := seenTargets[key]; seen {
			return
		}
		seenTargets[key] = struct{}{}
		targetFiles = append(targetFiles, DiscoveredLintTarget{
			Path:            filePath,
			CanonicalPath:   canonicalPath,
			ConfigDirectory: configDir,
		})
	}
	isGloballyIgnored := func(filePath string, matchPath string) bool {
		return globalIgnoreMatcher.IgnoresPath(filePath, matchPath)
	}

	includeExplicitFile := func(filePath string) (bool, string) {
		if !IsSupportedLintFile(filePath) {
			return false, ""
		}
		if fsys != nil && !fsys.FileExists(filePath) {
			return false, ""
		}
		matchPath := configPathForMatching(filePath)
		if isGloballyIgnored(filePath, matchPath) {
			return false, ""
		}
		canonicalPath := filePath
		if fsys != nil {
			if realPath := fsys.Realpath(filePath); realPath != "" {
				canonicalPath = realPath
			}
		}
		return true, canonicalPath
	}

	includeDiscoveredFile := func(filePath string) (bool, string) {
		if !IsSupportedLintFile(filePath) {
			return false, ""
		}
		matchPath := configPathForMatching(filePath)
		if !selectorMatcher.Selects(matchPath) {
			return false, ""
		}
		if isGloballyIgnored(filePath, matchPath) {
			return false, ""
		}
		return true, matchPath
	}

	addExplicitTargets := func() {
		for _, f := range allowFileSet {
			if include, canonicalPath := includeExplicitFile(f); include {
				addTarget(f, canonicalPath)
			}
		}
	}

	// Fast path for explicit file-only invocations, e.g. lint-staged.
	if allowFileSet != nil && allowDirs == nil {
		addExplicitTargets()
		sort.Slice(targetFiles, func(i, j int) bool { return targetFiles[i].Path < targetFiles[j].Path })
		return targetFiles
	}

	normalizedConfigDir := normalizeGlobPath(configDir)
	fsAdapter := &vfsAdapter{vfs: fsys, root: normalizedConfigDir}

	var (
		targetMu  sync.Mutex
		dirIgnore sync.Map // map[string]bool — pattern check cache
	)

	stopWalkDirs := normalizeStopWalkDirs(normalizedConfigDir, stopDirs, useCaseSensitive)

	workers := runtime.GOMAXPROCS(0)
	if workers < 2 {
		workers = 2
	}
	if singleThreaded {
		workers = 1
	}

	processFile := func(walkPath string, isSymlink bool) {
		fullPath := tspath.NormalizePath(tspath.CombinePaths(normalizedConfigDir, walkPath))
		matchPath := configPathForMatching(fullPath)
		targetPath := fullPath
		scopedCanonicalPath := ""

		if allowFileSet != nil || allowDirs != nil {
			inScope := false
			if allowFileSet != nil {
				if _, ok := allowFileSet[comparisonKey(fullPath)]; ok {
					inScope = true
				}
			}
			if !inScope && allowDirs != nil {
				inScope, targetPath, scopedCanonicalPath = resolvePathThroughAllowedDirectories(
					fullPath,
					matchPath,
					resolvedAllowDirs,
					useCaseSensitive,
				)
			}
			if !inScope {
				return
			}
		}

		include, canonicalPath := includeDiscoveredFile(fullPath)
		if !include {
			return
		}
		if isSymlink && fsys != nil {
			if realPath := fsys.Realpath(fullPath); realPath != "" {
				canonicalPath = realPath
			}
		} else if scopedCanonicalPath != "" {
			canonicalPath = scopedCanonicalPath
		}

		targetMu.Lock()
		addTarget(targetPath, canonicalPath)
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
				childPath := path.Join(walkPath, name)
				if isStoppedWalkPath(childPath, stopWalkDirs, useCaseSensitive) {
					continue
				}
				if cached, ok := dirIgnore.Load(childPath); ok {
					blocked, _ := cached.(bool)
					if blocked {
						continue
					}
				} else {
					childFullPath := tspath.NormalizePath(tspath.CombinePaths(normalizedConfigDir, childPath))
					blocked := globalIgnoreMatcher.BlocksDirectory(
						childFullPath,
						configPathForMatching(childFullPath),
					)
					dirIgnore.Store(childPath, blocked)
					if blocked {
						continue
					}
				}
				childDirs = append(childDirs, childPath)
			} else {
				processFile(path.Join(walkPath, name), e.Type()&fs.ModeSymlink != 0)
			}
		}
		return childDirs
	}

	pool := newWalkPool(workers)
	walkRoots := discoverWalkRoots(normalizedConfigDir, allowDirs, fsys)
	walkRoots = filterInitialWalkRoots(
		walkRoots,
		normalizedConfigDir,
		globalIgnoreMatcher,
		configPathForMatching,
		stopWalkDirs,
		useCaseSensitive,
	)
	pool.submitMany(walkRoots)
	pool.run(work)

	if allowFileSet != nil {
		addExplicitTargets()
	}

	sort.Slice(targetFiles, func(i, j int) bool { return targetFiles[i].Path < targetFiles[j].Path })
	return targetFiles
}

func normalizeStopWalkDirs(configDir string, stopDirs []string, useCaseSensitive bool) []string {
	if len(stopDirs) == 0 {
		return nil
	}

	compareOptions := tspath.ComparePathsOptions{
		CurrentDirectory:          configDir,
		UseCaseSensitiveFileNames: useCaseSensitive,
	}
	seen := make(map[string]struct{}, len(stopDirs))
	normalized := make([]string, 0, len(stopDirs))
	for _, rawDir := range stopDirs {
		dir := tspath.NormalizePath(rawDir)
		if pathsEqual(dir, configDir, useCaseSensitive) ||
			!tspath.StartsWithDirectory(dir, configDir, useCaseSensitive) {
			continue
		}
		rel := tspath.NormalizePath(tspath.GetRelativePathFromDirectory(configDir, dir, compareOptions))
		if rel == "" || rel == "." {
			continue
		}
		if _, ok := seen[rel]; ok {
			continue
		}
		seen[rel] = struct{}{}
		normalized = append(normalized, rel)
	}
	sort.Strings(normalized)
	return normalized
}

func filterInitialWalkRoots(
	roots []string,
	configDir string,
	globalIgnores GlobalIgnoreMatcher,
	configPathForMatching func(string) string,
	stopWalkDirs []string,
	useCaseSensitive bool,
) []string {
	if len(roots) == 0 {
		return roots
	}

	filtered := make([]string, 0, len(roots))
	for _, root := range roots {
		root = tspath.NormalizePath(root)
		if root == "" {
			root = "."
		}
		if root != "." {
			fullPath := tspath.NormalizePath(tspath.CombinePaths(configDir, root))
			if isStoppedWalkPath(root, stopWalkDirs, useCaseSensitive) ||
				globalIgnores.BlocksDirectory(fullPath, configPathForMatching(fullPath)) {
				continue
			}
		}
		filtered = append(filtered, root)
	}
	return filtered
}

func isStoppedWalkPath(walkPath string, stopWalkDirs []string, useCaseSensitive bool) bool {
	if len(stopWalkDirs) == 0 {
		return false
	}
	walkPath = tspath.NormalizePath(walkPath)
	if walkPath == "" || walkPath == "." {
		return false
	}
	for _, stopDir := range stopWalkDirs {
		if pathsEqual(walkPath, stopDir, useCaseSensitive) ||
			tspath.StartsWithDirectory(walkPath, stopDir, useCaseSensitive) {
			return true
		}
	}
	return false
}

func isFileInAllowedDirs(filePath string, allowDirs []string, useCaseSensitive bool) bool {
	for _, dir := range allowDirs {
		dir = tspath.NormalizePath(dir)
		if pathsEqual(filePath, dir, useCaseSensitive) ||
			tspath.StartsWithDirectory(filePath, dir, useCaseSensitive) {
			return true
		}
	}
	return false
}

func isFileInAllowedDirsWithFS(filePath string, allowDirs []string, fsys vfs.FS) bool {
	if isFileInAllowedDirs(filePath, allowDirs, true) {
		return true
	}
	if fsys == nil {
		return false
	}
	realFilePath := fsys.Realpath(filePath)
	if realFilePath == "" {
		return false
	}
	canonicalFile := tspath.NormalizePath(realFilePath)
	if canonicalFile == tspath.NormalizePath(filePath) {
		return false
	}
	for _, dir := range allowDirs {
		realDir := fsys.Realpath(tspath.NormalizePath(dir))
		if realDir == "" {
			continue
		}
		canonicalDir := tspath.NormalizePath(realDir)
		if pathsEqual(canonicalFile, canonicalDir, true) ||
			tspath.StartsWithDirectory(canonicalFile, canonicalDir, true) {
			return true
		}
	}
	return false
}

type resolvedAllowedDirectory struct {
	lexicalPath   string
	canonicalPath string
}

func resolveAllowedDirectories(allowDirs []string, fsys vfs.FS) []resolvedAllowedDirectory {
	if allowDirs == nil {
		return nil
	}
	resolved := make([]resolvedAllowedDirectory, 0, len(allowDirs))
	for _, dir := range allowDirs {
		lexicalPath := tspath.NormalizePath(dir)
		canonicalPath := lexicalPath
		if fsys != nil {
			if realPath := fsys.Realpath(lexicalPath); realPath != "" {
				canonicalPath = tspath.NormalizePath(realPath)
			}
		}
		resolved = append(resolved, resolvedAllowedDirectory{
			lexicalPath:   lexicalPath,
			canonicalPath: canonicalPath,
		})
	}
	return resolved
}

func resolvePathThroughAllowedDirectories(
	filePath string,
	matchPath string,
	allowDirs []resolvedAllowedDirectory,
	useCaseSensitive bool,
) (bool, string, string) {
	inScope := false
	bestLexicalLength := -1
	bestCanonicalLength := -1
	lexicalPath := filePath
	canonicalPath := ""
	for _, dir := range allowDirs {
		if relative, within := RelativePathWithinConfigRoot(filePath, dir.lexicalPath, useCaseSensitive); within {
			inScope = true
			if dir.lexicalPath != dir.canonicalPath && len(dir.lexicalPath) > bestLexicalLength {
				bestLexicalLength = len(dir.lexicalPath)
				canonicalPath = tspath.ResolvePath(dir.canonicalPath, relative)
			}
		}
		if dir.lexicalPath == dir.canonicalPath || len(dir.canonicalPath) <= bestCanonicalLength {
			continue
		}
		relative, within := RelativePathWithinConfigRoot(filePath, dir.canonicalPath, true)
		if !within && matchPath != filePath {
			relative, within = RelativePathWithinConfigRoot(matchPath, dir.canonicalPath, true)
		}
		if within {
			inScope = true
			bestCanonicalLength = len(dir.canonicalPath)
			lexicalPath = tspath.ResolvePath(dir.lexicalPath, relative)
		}
	}
	return inScope, lexicalPath, canonicalPath
}

func discoverWalkRoots(configDir string, allowDirs []string, fsys vfs.FS) []string {
	if allowDirs == nil {
		return []string{"."}
	}
	if len(allowDirs) == 0 {
		return nil
	}

	configDir = tspath.NormalizePath(configDir)
	canonicalConfigDir := configDir
	if fsys != nil {
		if realPath := fsys.Realpath(configDir); realPath != "" {
			canonicalConfigDir = tspath.NormalizePath(realPath)
		}
	}
	seen := make(map[string]struct{}, len(allowDirs))
	roots := make([]string, 0, len(allowDirs))
	addRoot := func(root string) {
		if root == "" {
			root = "."
		}
		root = tspath.NormalizePath(root)
		if root == "." {
			roots = []string{"."}
			seen = map[string]struct{}{".": {}}
			return
		}
		for _, existing := range roots {
			if existing == "." {
				return
			}
			if pathsEqual(existing, root, true) ||
				tspath.StartsWithDirectory(root, existing, true) {
				return
			}
		}
		filtered := roots[:0]
		seen = make(map[string]struct{}, len(allowDirs))
		for _, existing := range roots {
			if pathsEqual(existing, root, true) ||
				tspath.StartsWithDirectory(existing, root, true) {
				continue
			}
			seen[existing] = struct{}{}
			filtered = append(filtered, existing)
		}
		roots = filtered
		if _, ok := seen[root]; ok {
			return
		}
		seen[root] = struct{}{}
		roots = append(roots, root)
	}

	for _, rawDir := range allowDirs {
		dir := tspath.NormalizePath(rawDir)
		if pathsEqual(dir, configDir, true) ||
			tspath.StartsWithDirectory(configDir, dir, true) {
			return []string{"."}
		}
		if relative, within := RelativePathWithinConfigRoot(dir, configDir, true); within {
			addRoot(relative)
			continue
		}
		if fsys == nil {
			continue
		}
		realDir := fsys.Realpath(dir)
		if realDir == "" {
			continue
		}
		canonicalDir := tspath.NormalizePath(realDir)
		if pathsEqual(canonicalDir, canonicalConfigDir, true) ||
			tspath.StartsWithDirectory(canonicalConfigDir, canonicalDir, true) {
			return []string{"."}
		}
		if relative, within := RelativePathWithinConfigRoot(canonicalDir, canonicalConfigDir, true); within {
			addRoot(relative)
		}
	}

	sort.Strings(roots)
	return roots
}

func pathsEqual(a, b string, useCaseSensitive bool) bool {
	if useCaseSensitive {
		return a == b
	}
	return strings.EqualFold(a, b)
}

// DiscoveredLintTarget preserves the config owner established during the
// directory walk so later stages do not need to infer ownership from paths.
type DiscoveredLintTarget struct {
	Path            string
	CanonicalPath   string
	ConfigDirectory string
}

// DiscoverGapFiles returns resolved lint targets that are absent from existing
// Programs. The filesystem walk and config/default-files matching are owned by
// DiscoverLintFiles; this helper only subtracts programFiles for callers that
// need a non-project-backed fallback Program.
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
