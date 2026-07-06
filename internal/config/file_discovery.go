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
)

// sourceFileExtensions are the extensions walkDir treats as lintable source
// files. Mirrors the extension list historically used for the no-tsconfig
// directory scan (cmd/rslint/programs.go).
var sourceFileExtensions = []string{".ts", ".tsx", ".js", ".jsx", ".mts", ".mjs"}

func hasSourceExtension(name string) bool {
	for _, ext := range sourceFileExtensions {
		if strings.HasSuffix(name, ext) {
			return true
		}
	}
	return false
}

// walkJob is one unit of work for walkDir's worker pool: a directory to read,
// plus the gitignore-derived ignore chain inherited from ancestors (grows as
// nested .gitignore files are found) and the combined ignores/negation-reach
// ready for use (chain ++ staticIgnores, cached — see walkDir's doc comment
// on why gitignore patterns must stay ordered before staticIgnores).
type walkJob struct {
	path    string
	chain   []IgnorePattern
	ignores []IgnorePattern
	neg     negReach
}

// walkDir does a single top-down traversal of cwd, returning:
//   - every file not excluded by gitignoreSeed, any .gitignore found during
//     the walk, or staticIgnores (no extension filtering here —
//     DiscoverGapFiles applies that itself, only for the no-`files`-field
//     case; when a config entry sets `files`, the pattern itself is the only
//     filter, matching pre-existing behavior for e.g. `files: ["**/*"]`).
//     ".git" is the only hard-coded directory skip; everything else,
//     including node_modules, is only excluded if matched by a .gitignore
//     or the user's own config `ignores` — there is no built-in default.
//   - every .gitignore-derived glob pattern discovered during the walk
//     (flat, in top-down discovery order), for callers that need to fold them
//     into a config's global ignores for consumers outside this walk (e.g.
//     GetConfigForFile on a file that is already in a tsconfig Program).
//     Patterns already present in gitignoreSeed are NOT re-returned.
//
// gitignoreSeed and staticIgnores are kept as two separate inputs — rather
// than one flat pre-merged list — because they must stay in a fixed relative
// order that a single "seed, then append newly found .gitignore" list can't
// guarantee: gitignore-derived patterns (ancestor + seed + anything found
// during the walk) must ALWAYS be evaluated before staticIgnores (the user's
// own config `ignores`, e.g. `!dist/keep.ts`), so a user's `!` can override a
// gitignore rule regardless of whether that rule came from an ancestor
// .gitignore (known upfront) or one discovered mid-walk. Every job carries
// `chain` (the gitignore patterns accumulated root-to-here) separately from
// the cached `ignores = chain ++ staticIgnores`; only chain growth (a new
// .gitignore found in that directory) triggers recomputing `ignores`/`neg`.
//
// A .gitignore found in a directory applies to that directory and everything
// beneath it: its rules are appended after the inherited chain (so a nested
// `!` can override an ancestor .gitignore's rule — sequential evaluation,
// later wins) and threaded down to child jobs. Directories without their own
// .gitignore just pass their inherited chain/ignores/negReach to children
// unchanged (no recomputation).
//
// Symlinked subdirectories are skipped (vfsAdapter with followSymlinks=false),
// matching ESLint v10's flat-config walker semantics.
func walkDir(
	cwd string,
	fsys vfs.FS,
	gitignoreSeed []IgnorePattern,
	staticIgnores []IgnorePattern,
	singleThreaded bool,
) (files []string, gitignoreGlobs []string) {
	normalizedCwd := normalizeGlobPath(cwd)
	fsAdapter := &vfsAdapter{vfs: fsys, root: normalizedCwd}

	var filesMu, globsMu sync.Mutex

	workers := runtime.GOMAXPROCS(0)
	if workers < 2 {
		workers = 2
	}
	if singleThreaded {
		workers = 1
	}

	combine := func(chain []IgnorePattern) ([]IgnorePattern, negReach) {
		combined := make([]IgnorePattern, 0, len(chain)+len(staticIgnores))
		combined = append(combined, chain...)
		combined = append(combined, staticIgnores...)
		return combined, buildNegReach(combined)
	}

	work := func(job walkJob) []walkJob {
		f, err := fsAdapter.Open(job.path)
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

		chain := job.chain
		ignores := job.ignores
		neg := job.neg

		for _, e := range entries {
			if e.IsDir() || e.Name() != ".gitignore" {
				continue
			}
			dirAbsPath := normalizedCwd
			if job.path != "." {
				dirAbsPath = normalizedCwd + "/" + job.path
			}
			content, ok := fsys.ReadFile(dirAbsPath + "/.gitignore")
			if !ok {
				break
			}
			relDir := ""
			if job.path != "." {
				relDir = job.path
			}
			localGlobs := convertGitignoreToGlobs(content, relDir)
			if len(localGlobs) == 0 {
				break
			}
			globsMu.Lock()
			gitignoreGlobs = append(gitignoreGlobs, localGlobs...)
			globsMu.Unlock()

			newChain := make([]IgnorePattern, len(chain), len(chain)+len(localGlobs))
			copy(newChain, chain)
			chain = append(newChain, ParseIgnorePatterns(localGlobs)...)
			ignores, neg = combine(chain)
			break
		}

		var childJobs []walkJob
		for _, e := range entries {
			name := e.Name()
			if e.IsDir() {
				if name == ".git" {
					continue
				}
				childPath := path.Join(job.path, name)
				if canPruneDir(childPath, ignores, neg) {
					continue
				}
				childJobs = append(childJobs, walkJob{path: childPath, chain: chain, ignores: ignores, neg: neg})
				continue
			}
			walkPath := path.Join(job.path, name)
			fullPath := tspath.NormalizePath(path.Join(normalizedCwd, walkPath))
			if isFileIgnored(fullPath, ignores, cwd) {
				continue
			}
			filesMu.Lock()
			files = append(files, fullPath)
			filesMu.Unlock()
		}
		return childJobs
	}

	rootIgnores, rootNeg := combine(gitignoreSeed)
	pool := newWalkPool[walkJob](workers)
	pool.submitMany([]walkJob{{path: ".", chain: gitignoreSeed, ignores: rootIgnores, neg: rootNeg}})
	pool.run(work)

	sort.Strings(files)
	return files, gitignoreGlobs
}

// collectNestedGitignoreChain reads the .gitignore files, if any, in each
// directory strictly between configDir (exclusive — the caller reads
// configDir's own .gitignore separately) and the directory containing
// filePath (inclusive), in root-to-leaf order. This mirrors the order walkDir
// would encounter them while descending, so a deeper .gitignore's rules are
// appended after — and can override — a shallower one's.
//
// Returns the chain as parsed IgnorePatterns (ready to combine with other
// ignore sources) and as raw glob strings (for folding into the config's
// global ignores, matching what walkDir returns for the directories it
// visits).
func collectNestedGitignoreChain(filePath, configDir string, fsys vfs.FS) (chain []IgnorePattern, globs []string) {
	normalizedConfigDir := normalizeGlobPath(configDir)
	fileDir := path.Dir(normalizeGlobPath(filePath))
	if fileDir == normalizedConfigDir || !tspath.StartsWithDirectory(fileDir, normalizedConfigDir, true) {
		return nil, nil
	}

	rel := strings.TrimPrefix(fileDir, normalizedConfigDir+"/")
	var segments []string
	for rel != "" && rel != "." {
		segments = append(segments, rel)
		rel = path.Dir(rel)
	}
	for i, j := 0, len(segments)-1; i < j; i, j = i+1, j-1 {
		segments[i], segments[j] = segments[j], segments[i]
	}

	for _, seg := range segments {
		content, ok := fsys.ReadFile(normalizedConfigDir + "/" + seg + "/.gitignore")
		if !ok {
			continue
		}
		localGlobs := convertGitignoreToGlobs(content, seg)
		if len(localGlobs) == 0 {
			continue
		}
		globs = append(globs, localGlobs...)
		chain = append(chain, ParseIgnorePatterns(localGlobs)...)
	}
	return chain, globs
}

// DiscoverGapFiles discovers files that should get a fallback (AST-only, no
// type info) Program: they are not already covered by any tsconfig-backed
// Program, but the linter would still assign them rules.
//
// Two scenarios trigger this:
//   - hasTsConfig is false: there is no tsconfig at all for configDir (a pure
//     JS/TS project). Every ts/js file under configDir not already in
//     programFiles is a gap file — a whole-directory fallback Program takes
//     the place of the old ad-hoc "no tsconfig" directory scan.
//   - a config entry sets `files`: files matching those patterns but absent
//     from every tsconfig Program are gaps, even when a tsconfig exists.
//
// When hasTsConfig is true and no entry sets `files`, the tsconfig's own
// `include`/`exclude` resolution is authoritative and this function is a
// no-op (nil, nil) — legacy behavior.
//
// Files pass through these filters:
//  1. Match at least one config entry's `files` pattern (skipped when no
//     entry sets `files` — legacy semantics: applies to everything)
//  2. Not in global ignores or any .gitignore (enforced inside walkDir)
//  3. Not already in programFiles (existing tsconfig Programs)
//  4. Pass GetConfigForFile (would actually receive lint rules)
//
// Returns the gap files (sorted, deduplicated) and every .gitignore-derived
// glob pattern relevant to configDir (ancestor + discovered during the walk),
// for the caller to fold into the config's global ignores.
func DiscoverGapFiles(
	config RslintConfig,
	configDir string,
	fsys vfs.FS,
	programFiles map[string]struct{},
	allowFiles []string,
	allowDirs []string,
	singleThreaded bool,
	hasTsConfig bool,
) (gapFiles []string, gitignoreGlobs []string) {
	configIgnores := ExtractConfigIgnores(config)

	var allFilesPatterns []string
	hasFilesField := false
	for _, entry := range config {
		if len(entry.Files) > 0 {
			hasFilesField = true
			allFilesPatterns = append(allFilesPatterns, entry.Files...)
		}
	}

	// tsconfig found and no entry restricts scope via `files` → legacy
	// behavior, tsconfig `include`/`exclude` is authoritative.
	if hasTsConfig && !hasFilesField {
		return nil, nil
	}

	allFilesPatterns = deduplicate(allFilesPatterns)

	var allowFileSet map[string]struct{}
	if allowFiles != nil {
		allowFileSet = make(map[string]struct{}, len(allowFiles))
		for _, f := range allowFiles {
			allowFileSet[tspath.NormalizePath(f)] = struct{}{}
		}
	}

	// staticIgnores (the user's own config `ignores`) must always be
	// evaluated AFTER gitignore-derived patterns, so a user's `!` can
	// override a gitignore rule regardless of where that rule came from. See
	// walkDir's doc comment for why these stay as two separate lists instead
	// of one pre-merged seed.
	staticIgnores := configIgnores

	normalizedConfigDirForGitignore := normalizeGlobPath(configDir)
	ancestorGlobs := collectAncestorGitignoreGlobs(normalizedConfigDirForGitignore, fsys)
	gitignoreSeed := ParseIgnorePatterns(ancestorGlobs)

	result := []string{}

	// Fast path: when only specific files are provided (e.g., lint-staged),
	// check each file directly instead of walking the entire filesystem. This
	// path skips walkDir entirely, so it wouldn't otherwise see any
	// .gitignore at all — read configDir's own .gitignore directly (one
	// extra file read, not a tree walk) so the common case (a single
	// top-level .gitignore) still works. For a file nested under configDir,
	// the .gitignore files in directories between configDir and the file are
	// also read directly (one read per intervening directory, not a walk of
	// the whole tree) via collectNestedGitignoreChain, so an explicitly-passed
	// nested file is still correctly reported as ignored.
	if allowFileSet != nil && allowDirs == nil {
		ownGlobs, _ := fsys.ReadFile(normalizedConfigDirForGitignore + "/.gitignore")
		fastGitignore := gitignoreSeed
		rootGlobs := ancestorGlobs
		if ownGlobs != "" {
			converted := convertGitignoreToGlobs(ownGlobs, "")
			if len(converted) > 0 {
				fastGitignore = append(append([]IgnorePattern{}, gitignoreSeed...), ParseIgnorePatterns(converted)...)
				rootGlobs = append(append([]string{}, ancestorGlobs...), converted...)
			}
		}
		seenNestedGlobs := make(map[string]struct{})
		for f := range allowFileSet {
			// See the same check in the walk path below for why this only
			// applies when hasTsConfig is true.
			if hasTsConfig {
				if _, exists := programFiles[f]; exists {
					continue
				}
			}
			nestedChain, nestedGlobs := collectNestedGitignoreChain(f, configDir, fsys)
			for _, g := range nestedGlobs {
				if _, exists := seenNestedGlobs[g]; !exists {
					seenNestedGlobs[g] = struct{}{}
					rootGlobs = append(rootGlobs, g)
				}
			}
			fastIgnores := append(append([]IgnorePattern{}, fastGitignore...), nestedChain...)
			fastIgnores = append(fastIgnores, staticIgnores...)
			if isFileIgnored(f, fastIgnores, configDir) {
				continue
			}
			if config.GetConfigForFile(f, configDir) == nil {
				continue
			}
			result = append(result, f)
		}
		result = deduplicate(result)
		sort.Strings(result)
		return result, rootGlobs
	}

	walked, discoveredGlobs := walkDir(configDir, fsys, gitignoreSeed, staticIgnores, singleThreaded)

	normalizedConfigDir := normalizeGlobPath(configDir)
	relativePatterns := make([]string, len(allFilesPatterns))
	for i, pattern := range allFilesPatterns {
		normalizedPattern := normalizeGlobPath(tspath.ResolvePath(configDir, pattern))
		relativePatterns[i] = strings.TrimPrefix(normalizedPattern, normalizedConfigDir+"/")
	}

	for _, fullPath := range walked {
		// 1. pattern match (skipped when no entry sets `files`: legacy
		// semantics treat that as "applies to everything" — but only to
		// ts/js source files, matching the old no-tsconfig directory scan
		// this case replaces; a `files` pattern like "**/*" is itself the
		// only filter when one is present, so e.g. .md/.json files stay
		// discoverable there).
		if hasFilesField {
			relPath := strings.TrimPrefix(fullPath, normalizedConfigDir+"/")
			matched := false
			for _, pattern := range relativePatterns {
				if ok, _ := doublestar.Match(pattern, relPath); ok {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		} else if !hasSourceExtension(fullPath) {
			continue
		}
		// 2. already in tsconfig Programs. Only applies when THIS configDir
		// has its own tsconfig: programFiles is a cross-config set, and an
		// ancestor config's broader tsconfig `include` can already cover a
		// file that structurally belongs to a nested config with no
		// tsconfig of its own (e.g. a monorepo root tsconfig with
		// `include: ["**/*.ts"]` sweeping up every package). Excluding it
		// here would mean it never gets the nested config's own rules at
		// all. Instead, when hasTsConfig is false, always treat it as a gap
		// file for this config — the existing nearest-config ownership
		// resolution (buildFileOwnerMap/buildFileFilters) is what decides
		// which program's rules actually apply per file, exactly as it did
		// for the old no-tsconfig whole-directory scan Program.
		if hasTsConfig {
			if _, exists := programFiles[fullPath]; exists {
				continue
			}
		}
		// 3. CLI scope
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
				continue
			}
		}
		// 4. final config check
		if config.GetConfigForFile(fullPath, configDir) == nil {
			continue
		}
		result = append(result, fullPath)
	}

	result = deduplicate(result)
	sort.Strings(result)

	allGlobs := make([]string, 0, len(ancestorGlobs)+len(discoveredGlobs))
	allGlobs = append(allGlobs, ancestorGlobs...)
	allGlobs = append(allGlobs, discoveredGlobs...)

	return result, allGlobs
}

// DiscoverGapFilesMultiConfig runs DiscoverGapFiles for each config in a
// monorepo config map and returns the union of all gap files, plus the
// .gitignore-derived globs discovered per configDir (kept separate per dir —
// a child config's own .gitignore must not leak into its siblings or parent).
//
// hasTsConfigByDir carries, per configDir, whether a tsconfig was found for
// that config (see DiscoverGapFiles). A missing entry defaults to false
// (walks that configDir), which is the safe default.
func DiscoverGapFilesMultiConfig(
	configMap map[string]RslintConfig,
	fsys vfs.FS,
	programFiles map[string]struct{},
	allowFiles []string,
	allowDirs []string,
	singleThreaded bool,
	hasTsConfigByDir map[string]bool,
) (gapFiles []string, gitignoreGlobsByDir map[string][]string) {
	if len(configMap) == 0 {
		return nil, nil
	}

	seen := make(map[string]struct{})
	var allGapFiles []string
	globsByDir := make(map[string][]string, len(configMap))
	for configDir, cfg := range configMap {
		dirGapFiles, dirGlobs := DiscoverGapFiles(
			cfg, configDir, fsys, programFiles, allowFiles, allowDirs, singleThreaded, hasTsConfigByDir[configDir],
		)
		for _, f := range dirGapFiles {
			if _, exists := seen[f]; !exists {
				seen[f] = struct{}{}
				allGapFiles = append(allGapFiles, f)
			}
		}
		if len(dirGlobs) > 0 {
			globsByDir[configDir] = dirGlobs
		}
	}

	if len(allGapFiles) > 0 {
		sort.Strings(allGapFiles)
	} else {
		allGapFiles = nil
	}
	if len(globsByDir) == 0 {
		globsByDir = nil
	}
	return allGapFiles, globsByDir
}

// walkPool is a fixed-size worker pool with an unbounded internal LIFO queue,
// used by walkDir to walk directory trees with a bounded number of live
// goroutines. Properties:
//
//   - At most `workers` goroutines exist concurrently.
//   - submitMany never blocks (queue grows as needed; total memory ~ peak
//     queue size, bounded by FS branching × depth).
//   - run() returns once all submitted work and all transitively submitted
//     work have completed. Detection: when every worker is simultaneously
//     idle and the queue is empty.
//
// workers=1 degenerates to an effectively serial DFS-style traversal (LIFO
// pops the most recently submitted item); the only goroutine is the worker
// itself. This is what --singleThreaded relies on.
type walkPool[T any] struct {
	mu      sync.Mutex
	cond    *sync.Cond
	queue   []T
	idle    int
	workers int
	done    bool
}

func newWalkPool[T any](workers int) *walkPool[T] {
	if workers < 1 {
		workers = 1
	}
	p := &walkPool[T]{workers: workers}
	p.cond = sync.NewCond(&p.mu)
	return p
}

// submitMany appends items to the queue and wakes idle workers. Safe to call
// from any goroutine.
func (p *walkPool[T]) submitMany(items []T) {
	if len(items) == 0 {
		return
	}
	p.mu.Lock()
	p.queue = append(p.queue, items...)
	p.mu.Unlock()
	if len(items) == 1 {
		p.cond.Signal()
	} else {
		p.cond.Broadcast()
	}
}

// take pops a job from the queue, blocking if empty. Returns (zero, false)
// only when the queue is empty AND every worker is simultaneously idle —
// which means no more work will ever appear.
func (p *walkPool[T]) take() (T, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for {
		if len(p.queue) > 0 {
			n := len(p.queue) - 1
			item := p.queue[n]
			p.queue = p.queue[:n]
			return item, true
		}
		p.idle++
		if p.idle == p.workers {
			p.done = true
			p.cond.Broadcast()
			var zero T
			return zero, false
		}
		p.cond.Wait()
		if p.done {
			var zero T
			return zero, false
		}
		p.idle--
	}
}

// run drives the worker pool. Each worker pulls jobs and calls work(item),
// which returns the child jobs to enqueue. Returns when all reachable work
// is processed.
//
// When workers == 1, runs the loop on the calling goroutine directly — no
// goroutines are spawned at all. This is what --singleThreaded relies on:
// callers can rely on the Go side spawning no extra goroutines.
func (p *walkPool[T]) run(work func(T) []T) {
	if p.workers == 1 {
		for {
			item, ok := p.take()
			if !ok {
				return
			}
			p.submitMany(work(item))
		}
	}
	var wg sync.WaitGroup
	for range p.workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				item, ok := p.take()
				if !ok {
					return
				}
				p.submitMany(work(item))
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
