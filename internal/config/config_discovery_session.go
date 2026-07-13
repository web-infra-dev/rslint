package config

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"sort"
	"strings"
	"sync"

	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
)

type configCandidate struct {
	path      string
	directory string
}

type configLoadState struct {
	id              string
	candidate       configCandidate
	entries         RslintConfig
	fingerprint     string
	eslintPlugins   []EslintPluginEntry
	hasPluginConfig bool
	failure         *ConfigModuleError
}

type discoverySeed struct {
	path               string
	searchDir          string
	canonicalSearchDir string
	canonicalWalkDir   string
	usingCanonical     bool
	lexicalCandidate   bool
	explicitFile       bool
	ownerDir           string
	done               bool
}

type discoveryWalkNode struct {
	directory          string
	canonicalDirectory string
	ownerDir           string
	targets            *discoveryTargetTrie
}

// discoveryTargetTrie bounds a directory catalog walk to the lexical ancestor
// paths of an already-expanded target set. A nil trie means an unbounded
// directory-only walk (CLI/LSP); a non-nil leaf still visits that directory so
// its config can govern a file located directly within it.
type discoveryTargetTrie struct {
	children map[tspath.Path]*discoveryTargetTrie
}

type suspendedDiscoveryNode struct {
	node      discoveryWalkNode
	candidate configCandidate
}

type discoveryWalkResult struct {
	children            []discoveryWalkNode
	pending             *suspendedDiscoveryNode
	activation          *configLoadState
	directoriesVisited  int
	directoriesPruned   int
	discoveredCandidate bool
	err                 error
}

type directorySeedResolution struct {
	seed       *discoverySeed
	candidates []configCandidate
	next       int
}

type configCatalogBuilder struct {
	ctx           context.Context
	fs            vfs.FS
	loader        ConfigModuleLoader
	request       ConfigDiscoveryRequest
	generation    uint64
	transactionID string

	loadStates    map[string]*configLoadState
	configs       map[string]RslintConfig
	sources       map[string]ConfigSource
	scopes        map[string]LintDiscoveryScope
	failureByPath map[string]ConfigFailure
	hadCandidates bool
	nextRequestID int
	stats         ConfigDiscoveryStats
}

func (builder *configCatalogBuilder) build() (*ConfigCatalog, error) {
	if err := builder.ctx.Err(); err != nil {
		return nil, err
	}
	cwd := builder.request.CWD
	if cwd == "" {
		return nil, errors.New("config discovery requires a working directory")
	}
	cwd = tspath.NormalizePath(cwd)
	builder.request.CWD = cwd

	if builder.request.Mode == ConfigDiscoveryExplicit {
		if builder.request.ExplicitConfigPath == "" {
			return nil, errors.New("explicit config discovery requires a config path")
		}
		candidate := configCandidate{
			path: normalizeDiscoveryPath(builder.request.ExplicitConfigPath, cwd),
			// Explicit flat configs resolve files/ignores/projects from the
			// invocation cwd, not the module's physical directory.
			directory: cwd,
		}
		builder.hadCandidates = true
		if err := builder.ensureCandidates([]configCandidate{candidate}); err != nil {
			return nil, err
		}
		state := builder.loadStates[candidate.path]
		if state == nil || state.failure != nil {
			return nil, builder.allConfigsFailedError()
		}
		builder.activate(state, false, true)
		return builder.catalog()
	}

	directoryRoots := builder.normalizedDirectoryRoots()
	files := builder.normalizedFiles()
	var targetsByRoot map[string]*discoveryTargetTrie
	boundedDirectoryWalk := builder.request.LimitDirectoryWalkToFiles && len(directoryRoots) > 0 && len(files) > 0
	if boundedDirectoryWalk {
		targetsByRoot = builder.targetAncestorTries(directoryRoots, files)
	}
	seeds := make([]*discoverySeed, 0, len(directoryRoots)+len(builder.request.Files))
	directorySeeds := make([]*discoverySeed, 0, len(directoryRoots))
	directorySeedByPath := make(map[string]*discoverySeed, len(directoryRoots))
	for _, directory := range directoryRoots {
		if boundedDirectoryWalk && targetsByRoot[directory] == nil {
			continue
		}
		useCaseSensitive := builder.fs.UseCaseSensitiveFileNames()
		defaultExcluded := isDefaultDiscoveryExcluded(directory, cwd, useCaseSensitive)
		ancestryOnly := defaultExcluded
		searchDirectory := directory
		if ancestryOnly {
			builder.stats.DirectoriesPruned++
			// A default-excluded directory is a downward traversal boundary, not a
			// reason to discard still-reachable configuration outside it. Skip the
			// root and every default-excluded ancestor, then resolve normally.
			searchDirectory = tspath.GetDirectoryPath(directory)
			for searchDirectory != "" && isDefaultDiscoveryExcluded(searchDirectory, cwd, useCaseSensitive) {
				parent := tspath.GetDirectoryPath(searchDirectory)
				if parent == searchDirectory {
					searchDirectory = ""
					break
				}
				searchDirectory = parent
			}
			if searchDirectory == "" {
				continue
			}
		}
		seed := &discoverySeed{
			path:      directory,
			searchDir: searchDirectory,
		}
		if !ancestryOnly {
			builder.addCanonicalSeedFallback(seed, builder.fs.Realpath(directory), true)
			directorySeedByPath[directory] = seed
		}
		seeds = append(seeds, seed)
		directorySeeds = append(directorySeeds, seed)
	}

	explicitSeeds := make([]*discoverySeed, 0, len(builder.request.Files))
	fileSeeds := make([]*discoverySeed, 0, len(builder.request.Files))
	for _, file := range files {
		if !file.Explicit {
			// Files produced by a glob/directory walk inherit the staged target
			// trie. Resolving each file independently would reopen parent-global-
			// ignored subtrees and turn a bounded walk back into O(files * ancestry).
			continue
		}
		fileDirectory := tspath.GetDirectoryPath(file.Path)
		defaultExcluded := isDefaultDiscoveryExcluded(file.Path, cwd, builder.fs.UseCaseSensitiveFileNames())
		seed := &discoverySeed{
			path:         file.Path,
			searchDir:    fileDirectory,
			explicitFile: true,
		}
		canonicalPath := file.CanonicalPath
		if canonicalPath == "" {
			canonicalPath = builder.fs.Realpath(file.Path)
		}
		if !defaultExcluded {
			// A default-excluded literal may search only its reachable lexical
			// ancestry to explain the ignored result. A realpath fallback could
			// escape the ignored subtree and execute an unrelated physical config
			// even though the target can never enter the lint scope.
			builder.addCanonicalSeedFallback(seed, canonicalPath, false)
		}
		seeds = append(seeds, seed)
		fileSeeds = append(fileSeeds, seed)
		if !defaultExcluded {
			explicitSeeds = append(explicitSeeds, seed)
		}
	}

	if err := builder.resolveDirectorySeedOwners(directorySeeds); err != nil {
		return nil, err
	}
	if err := builder.resolveSeedOwners(fileSeeds); err != nil {
		return nil, err
	}
	for _, seed := range seeds {
		if seed.ownerDir == "" {
			continue
		}
		state := builder.loadStateForDirectory(seed.ownerDir)
		if state == nil {
			continue
		}
		builder.activate(state, seed.explicitFile, false)
	}
	for _, seed := range explicitSeeds {
		if seed.ownerDir == "" {
			continue
		}
		scope := builder.scopes[seed.ownerDir]
		scope.Files = appendUniqueSortedPath(scope.Files, seed.path)
		builder.scopes[seed.ownerDir] = scope
	}

	walkRoots := make([]discoveryWalkNode, 0, len(directorySeedByPath))
	for _, directory := range directoryRoots {
		seed := directorySeedByPath[directory]
		if seed == nil {
			continue
		}
		walkRoots = append(walkRoots, discoveryWalkNode{
			directory:          directory,
			canonicalDirectory: seed.canonicalWalkDir,
			ownerDir:           seed.ownerDir,
			targets:            targetsByRoot[directory],
		})
	}
	if err := builder.walkDirectories(walkRoots); err != nil {
		return nil, err
	}

	if builder.hadCandidates && len(builder.configs) == 0 {
		return nil, builder.allConfigsFailedError()
	}
	return builder.catalog()
}

func (builder *configCatalogBuilder) normalizedDirectoryRoots() []string {
	raw := builder.request.Directories
	if len(raw) == 0 && len(builder.request.Files) == 0 && builder.request.ImplicitCWD {
		raw = []string{builder.request.CWD}
	}
	seen := make(map[string]struct{}, len(raw))
	roots := make([]string, 0, len(raw))
	for _, directory := range raw {
		directory = normalizeDiscoveryPath(directory, builder.request.CWD)
		if _, exists := seen[directory]; exists {
			continue
		}
		seen[directory] = struct{}{}
		roots = append(roots, directory)
	}
	sort.Slice(roots, func(i, j int) bool {
		if len(roots[i]) != len(roots[j]) {
			return len(roots[i]) < len(roots[j])
		}
		return roots[i] < roots[j]
	})
	compact := roots[:0]
	for _, root := range roots {
		covered := false
		for _, parent := range compact {
			if pathsEqual(root, parent, builder.fs.UseCaseSensitiveFileNames()) ||
				tspath.StartsWithDirectory(root, parent, builder.fs.UseCaseSensitiveFileNames()) {
				covered = true
				break
			}
		}
		if !covered {
			compact = append(compact, root)
		}
	}
	return compact
}

func (builder *configCatalogBuilder) normalizedFiles() []DiscoveryFile {
	files := make([]DiscoveryFile, 0, len(builder.request.Files))
	indexByPath := make(map[string]int, len(builder.request.Files))
	for _, file := range builder.request.Files {
		file.Path = normalizeDiscoveryPath(file.Path, builder.request.CWD)
		if file.CanonicalPath != "" {
			file.CanonicalPath = normalizeDiscoveryPath(file.CanonicalPath, builder.request.CWD)
		}
		if index, exists := indexByPath[file.Path]; exists {
			files[index].Explicit = files[index].Explicit || file.Explicit
			if files[index].CanonicalPath == "" {
				files[index].CanonicalPath = file.CanonicalPath
			}
			continue
		}
		indexByPath[file.Path] = len(files)
		files = append(files, file)
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	return files
}

// targetAncestorTries maps each compact directory root to the lexical paths
// that can govern a supplied target. Native API callers have already expanded
// their globs, so configs in every other subtree cannot affect this request and
// must not be evaluated. Directory-only CLI/LSP requests have no files and keep
// the existing unbounded walk by leaving the returned map empty.
func (builder *configCatalogBuilder) targetAncestorTries(
	roots []string,
	files []DiscoveryFile,
) map[string]*discoveryTargetTrie {
	tries := make(map[string]*discoveryTargetTrie, len(roots))
	useCaseSensitive := builder.fs.UseCaseSensitiveFileNames()
	for _, file := range files {
		for _, root := range roots {
			relative, within := relativeConfigPath(file.Path, root, useCaseSensitive)
			if !within {
				continue
			}
			trie := tries[root]
			if trie == nil {
				trie = &discoveryTargetTrie{}
				tries[root] = trie
			}
			relativeDirectory := tspath.GetDirectoryPath(relative)
			relativeDirectory = strings.ReplaceAll(relativeDirectory, "\\", "/")
			for _, segment := range strings.Split(relativeDirectory, "/") {
				if segment == "" || segment == "." {
					continue
				}
				if trie.children == nil {
					trie.children = make(map[tspath.Path]*discoveryTargetTrie)
				}
				key := tspath.ToPath(segment, "", useCaseSensitive)
				child := trie.children[key]
				if child == nil {
					child = &discoveryTargetTrie{}
					trie.children[key] = child
				}
				trie = child
			}
			// normalizedDirectoryRoots removes overlapping roots, so a file can
			// belong to at most one retained lexical root.
			break
		}
	}
	return tries
}

func (trie *discoveryTargetTrie) child(name string, useCaseSensitive bool) *discoveryTargetTrie {
	if trie == nil || len(trie.children) == 0 {
		return nil
	}
	return trie.children[tspath.ToPath(name, "", useCaseSensitive)]
}

// addCanonicalSeedFallback records one physical ancestry without replacing the
// authored path. The fallback is intentionally dormant until the complete
// lexical ancestry has produced no candidate at all.
func (builder *configCatalogBuilder) addCanonicalSeedFallback(seed *discoverySeed, canonicalPath string, isDirectory bool) {
	if seed == nil || canonicalPath == "" {
		return
	}
	canonicalPath = tspath.NormalizePath(canonicalPath)
	canonicalDirectory := canonicalPath
	if !isDirectory {
		canonicalDirectory = tspath.GetDirectoryPath(canonicalPath)
	}
	if canonicalDirectory == "" || canonicalDirectory == seed.searchDir {
		return
	}
	seed.canonicalWalkDir = canonicalDirectory
	seed.canonicalSearchDir = canonicalDirectory
}

// resolveDirectorySeedOwners evaluates each directory's complete config
// ancestry from outermost to innermost. This ordering is what makes an
// ancestor config's global ignore a discovery boundary for a nested config.
// Only the final reachable successful owner is activated later; ancestors
// loaded solely to decide reachability never leak into the effective catalog.
func (builder *configCatalogBuilder) resolveDirectorySeedOwners(seeds []*discoverySeed) error {
	resolutions := make([]directorySeedResolution, 0, len(seeds))
	for _, seed := range seeds {
		candidates := builder.findCandidateChain(seed.searchDir)
		if len(candidates) == 0 && seed.canonicalSearchDir != "" {
			seed.usingCanonical = true
			candidates = builder.findCandidateChain(seed.canonicalSearchDir)
		}
		resolutions = append(resolutions, directorySeedResolution{seed: seed, candidates: candidates})
	}

	for {
		if err := builder.ctx.Err(); err != nil {
			return err
		}
		var candidates []configCandidate
		pending := make([]*directorySeedResolution, 0, len(resolutions))
		for index := range resolutions {
			resolution := &resolutions[index]
			for resolution.next < len(resolution.candidates) {
				candidateDirectory := resolution.candidates[resolution.next].directory
				if resolution.seed.ownerDir != "" && builder.isGloballyIgnoredDirectory(
					resolution.seed.ownerDir,
					candidateDirectory,
					candidateDirectory,
				) {
					// An authored absolute directory ignore is a permanent traversal
					// boundary, so every deeper candidate in this ancestry is unreachable.
					resolution.next = len(resolution.candidates)
					break
				}
				candidate, found := builder.findCandidateForOwner(
					candidateDirectory,
					resolution.seed.ownerDir,
					candidateDirectory,
				)
				if !found {
					// File-cover ignores keep the directory traversable for later
					// negations, but a config candidate that remains ignored is not
					// evaluated. Continue at the next candidate-bearing directory.
					resolution.next++
					continue
				}
				builder.hadCandidates = true
				// The highest-priority on-disk candidate may be hidden by the
				// current owner's authored ignore while a lower-priority filename
				// remains reachable. Persist the actual request candidate so the
				// post-batch ownership update reads the matching load state.
				resolution.candidates[resolution.next] = candidate
				candidates = append(candidates, candidate)
				pending = append(pending, resolution)
				break
			}
		}
		if len(candidates) == 0 {
			return nil
		}
		if err := builder.ensureCandidates(candidates); err != nil {
			return err
		}
		for _, resolution := range pending {
			candidate := resolution.candidates[resolution.next]
			state := builder.loadStates[candidate.path]
			if state != nil && state.failure == nil {
				resolution.seed.ownerDir = candidate.directory
			}
			resolution.next++
		}
	}
}

func (builder *configCatalogBuilder) findCandidateChain(startDirectory string) []configCandidate {
	var reverse []configCandidate
	for directory := tspath.NormalizePath(startDirectory); directory != ""; {
		if candidate, ok := builder.findCandidate(directory); ok {
			reverse = append(reverse, candidate)
		}
		parent := tspath.GetDirectoryPath(directory)
		if parent == directory {
			break
		}
		directory = parent
	}
	candidates := make([]configCandidate, len(reverse))
	for index := range reverse {
		candidates[len(reverse)-1-index] = reverse[index]
	}
	return candidates
}

// resolveSeedOwners deliberately remains nearest-first for literal files.
// That is the only config-global-ignore ownership exception. Default
// exclusions are still enforced by findCandidate.
func (builder *configCatalogBuilder) resolveSeedOwners(seeds []*discoverySeed) error {
	for {
		if err := builder.ctx.Err(); err != nil {
			return err
		}
		var candidates []configCandidate
		candidateBySeed := make(map[*discoverySeed]configCandidate)
		for _, seed := range seeds {
			if seed.done {
				continue
			}
			candidate, found := builder.findCandidateUp(seed.searchDir)
			if !found {
				if !seed.usingCanonical && !seed.lexicalCandidate && seed.canonicalSearchDir != "" {
					seed.usingCanonical = true
					seed.searchDir = seed.canonicalSearchDir
					candidate, found = builder.findCandidateUp(seed.searchDir)
				}
			}
			if !found {
				seed.done = true
				continue
			}
			if !seed.usingCanonical {
				seed.lexicalCandidate = true
			}
			builder.hadCandidates = true
			candidateBySeed[seed] = candidate
			candidates = append(candidates, candidate)
		}
		if len(candidates) == 0 {
			return nil
		}
		if err := builder.ensureCandidates(candidates); err != nil {
			return err
		}
		for _, seed := range seeds {
			candidate, ok := candidateBySeed[seed]
			if !ok {
				continue
			}
			state := builder.loadStates[candidate.path]
			if state != nil && state.failure == nil {
				seed.ownerDir = candidate.directory
				seed.done = true
				continue
			}
			seed.searchDir = tspath.GetDirectoryPath(candidate.directory)
			if seed.searchDir == candidate.directory || seed.searchDir == "" {
				seed.done = true
			}
		}
	}
}

func (builder *configCatalogBuilder) findCandidateUp(startDirectory string) (configCandidate, bool) {
	for directory := tspath.NormalizePath(startDirectory); directory != ""; {
		if candidate, ok := builder.findCandidate(directory); ok {
			return candidate, true
		}
		parent := tspath.GetDirectoryPath(directory)
		if parent == directory {
			break
		}
		directory = parent
	}
	return configCandidate{}, false
}

func (builder *configCatalogBuilder) findCandidate(directory string) (configCandidate, bool) {
	return builder.findCandidateForOwner(directory, "", "")
}

func (builder *configCatalogBuilder) findCandidateForOwner(
	directory string,
	ownerDir string,
	canonicalDirectory string,
) (configCandidate, bool) {
	for _, name := range AutoJSConfigFileNames {
		candidatePath := tspath.CombinePaths(directory, name)
		if isDefaultDiscoveryExcluded(candidatePath, builder.request.CWD, builder.fs.UseCaseSensitiveFileNames()) ||
			!builder.fs.FileExists(candidatePath) {
			continue
		}
		canonicalCandidate := ""
		if canonicalDirectory != "" {
			canonicalCandidate = tspath.CombinePaths(canonicalDirectory, name)
		}
		if builder.isGloballyIgnoredCandidate(ownerDir, candidatePath, canonicalCandidate) {
			continue
		}
		return configCandidate{path: tspath.NormalizePath(candidatePath), directory: directory}, true
	}
	return configCandidate{}, false
}

func (builder *configCatalogBuilder) ensureCandidates(rawCandidates []configCandidate) error {
	unique := make(map[string]configCandidate, len(rawCandidates))
	for _, candidate := range rawCandidates {
		candidate.path = tspath.NormalizePath(candidate.path)
		candidate.directory = tspath.NormalizePath(candidate.directory)
		if _, loaded := builder.loadStates[candidate.path]; loaded {
			continue
		}
		unique[candidate.path] = candidate
	}
	if len(unique) == 0 {
		return nil
	}
	if builder.loader == nil {
		return errors.New("JavaScript config candidates require a module loader")
	}

	paths := make([]string, 0, len(unique))
	for path := range unique {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	loadMode := ConfigModuleLoadCached
	if builder.request.Fresh {
		loadMode = ConfigModuleLoadFresh
	}
	request := ConfigLoadBatchRequest{
		ProtocolVersion: ConfigDiscoveryProtocolVersion,
		TransactionID:   builder.transactionID,
		LoadMode:        loadMode,
		SingleThreaded:  builder.request.SingleThreaded,
	}
	request.Candidates = make([]ConfigLoadCandidate, 0, len(paths))
	for _, path := range paths {
		candidate := unique[path]
		builder.nextRequestID++
		id := fmt.Sprintf("%d:%06d", builder.generation, builder.nextRequestID)
		wireCandidate := ConfigLoadCandidate{
			ID:              id,
			ConfigPath:      candidate.path,
			ConfigDirectory: candidate.directory,
		}
		request.Candidates = append(request.Candidates, wireCandidate)
	}
	builder.stats.CandidatesFound += len(request.Candidates)
	builder.stats.ConfigsRequested += len(request.Candidates)
	response, err := builder.loader.LoadConfigs(builder.ctx, request)
	if err != nil {
		return err
	}
	results, err := validateConfigLoadBatch(request, response)
	if err != nil {
		return err
	}

	for _, wireCandidate := range request.Candidates {
		candidate := unique[wireCandidate.ConfigPath]
		result := results[wireCandidate.ID]
		state := &configLoadState{
			id:              wireCandidate.ID,
			candidate:       candidate,
			fingerprint:     result.SourceFingerprint,
			eslintPlugins:   cloneEslintPluginEntries(result.EslintPlugins),
			hasPluginConfig: result.HasPluginConfig,
		}
		if result.Status == "failed" {
			failure := *result.Error
			state.failure = &failure
		} else if err := ValidateConfig(result.Entries); err != nil {
			state.failure = &ConfigModuleError{Code: "invalid", Message: err.Error()}
		} else {
			state.entries = append(RslintConfig(nil), result.Entries...)
			builder.stats.ConfigsLoaded++
		}
		builder.loadStates[candidate.path] = state
		if state.failure != nil {
			builder.failureByPath[candidate.path] = ConfigFailure{
				Path:      candidate.path,
				Directory: candidate.directory,
				Kind:      state.failure.Code,
				Message:   state.failure.Message,
			}
		}
	}
	return nil
}

func validateConfigLoadBatch(request ConfigLoadBatchRequest, response ConfigLoadBatchResponse) (map[string]ConfigLoadResult, error) {
	if request.ProtocolVersion != ConfigDiscoveryProtocolVersion {
		return nil, configDiscoveryProtocolError("unsupported request protocol version %d", request.ProtocolVersion)
	}
	if request.TransactionID == "" {
		return nil, configDiscoveryProtocolError("request transactionId is empty")
	}
	if response.TransactionID != request.TransactionID {
		return nil, configDiscoveryProtocolError("transaction mismatch: got %q, want %q", response.TransactionID, request.TransactionID)
	}
	if len(response.Results) != len(request.Candidates) {
		return nil, configDiscoveryProtocolError("result count mismatch: got %d, want %d", len(response.Results), len(request.Candidates))
	}
	requestByID := make(map[string]ConfigLoadCandidate, len(request.Candidates))
	for _, candidate := range request.Candidates {
		if candidate.ID == "" {
			return nil, configDiscoveryProtocolError("request contains an empty candidate id")
		}
		if _, duplicate := requestByID[candidate.ID]; duplicate {
			return nil, configDiscoveryProtocolError("request contains duplicate candidate id %q", candidate.ID)
		}
		requestByID[candidate.ID] = candidate
	}
	results := make(map[string]ConfigLoadResult, len(response.Results))
	for index, result := range response.Results {
		_, exists := requestByID[result.ID]
		if !exists {
			return nil, configDiscoveryProtocolError("response contains unknown candidate id %q", result.ID)
		}
		if _, duplicate := results[result.ID]; duplicate {
			return nil, configDiscoveryProtocolError("response contains duplicate candidate id %q", result.ID)
		}
		if request.Candidates[index].ID != result.ID {
			return nil, configDiscoveryProtocolError("result order mismatch at index %d: got id %q, want %q", index, result.ID, request.Candidates[index].ID)
		}
		if result.Status != "loaded" && result.Status != "failed" {
			return nil, configDiscoveryProtocolError("candidate %q has invalid status %q", result.ID, result.Status)
		}
		if result.Status == "loaded" && result.Error != nil {
			return nil, configDiscoveryProtocolError("loaded candidate %q contains an error", result.ID)
		}
		if result.Status == "loaded" && result.SourceFingerprint == "" {
			return nil, configDiscoveryProtocolError("loaded candidate %q has no source fingerprint", result.ID)
		}
		if result.Status == "failed" && (result.Error == nil || result.Error.Message == "") {
			return nil, configDiscoveryProtocolError("failed candidate %q has no error message", result.ID)
		}
		results[result.ID] = result
	}
	return results, nil
}

func (builder *configCatalogBuilder) walkDirectories(roots []discoveryWalkNode) error {
	queue := append([]discoveryWalkNode(nil), roots...)
	for len(queue) > 0 {
		if err := builder.ctx.Err(); err != nil {
			return err
		}
		sort.Slice(queue, func(i, j int) bool { return queue[i].directory < queue[j].directory })
		var next []discoveryWalkNode
		var suspended []suspendedDiscoveryNode
		var candidates []configCandidate
		results := builder.processWalkFrontier(queue)
		// Results are indexed by the already-sorted frontier. Merging on this
		// goroutine keeps catalog state, stats, and loader batches deterministic
		// regardless of worker completion order.
		for _, result := range results {
			if result.err != nil {
				return result.err
			}
			builder.stats.DirectoriesVisited += result.directoriesVisited
			builder.stats.DirectoriesPruned += result.directoriesPruned
			builder.hadCandidates = builder.hadCandidates || result.discoveredCandidate
			if result.activation != nil {
				builder.activate(result.activation, false, false)
			}
			next = append(next, result.children...)
			if result.pending != nil {
				suspended = append(suspended, *result.pending)
				candidates = append(candidates, result.pending.candidate)
			}
		}
		if len(candidates) > 0 {
			if err := builder.ensureCandidates(candidates); err != nil {
				return err
			}
			for _, item := range suspended {
				state := builder.loadStates[item.candidate.path]
				if state != nil && state.failure == nil {
					builder.activate(state, false, false)
					item.node.ownerDir = state.candidate.directory
				}
				next = append(next, item.node)
			}
		}
		queue = deduplicateWalkNodes(next)
	}
	return nil
}

func (builder *configCatalogBuilder) processWalkFrontier(nodes []discoveryWalkNode) []discoveryWalkResult {
	results := make([]discoveryWalkResult, len(nodes))
	workers := runtime.GOMAXPROCS(0)
	if workers < 2 {
		workers = 2
	}
	if builder.request.SingleThreaded {
		workers = 1
	}
	if workers > len(nodes) {
		workers = len(nodes)
	}
	if workers <= 1 {
		for index, node := range nodes {
			results[index] = builder.processWalkNode(node)
		}
		return results
	}

	jobs := make(chan int)
	var waitGroup sync.WaitGroup
	waitGroup.Add(workers)
	for range workers {
		go func() {
			defer waitGroup.Done()
			for index := range jobs {
				results[index] = builder.processWalkNode(nodes[index])
			}
		}()
	}
	for index := range nodes {
		jobs <- index
	}
	close(jobs)
	waitGroup.Wait()
	return results
}

func (builder *configCatalogBuilder) processWalkNode(node discoveryWalkNode) discoveryWalkResult {
	if err := builder.ctx.Err(); err != nil {
		return discoveryWalkResult{err: err}
	}
	if builder.isGloballyIgnoredDirectory(node.ownerDir, node.directory, node.canonicalDirectory) {
		return discoveryWalkResult{directoriesPruned: 1}
	}
	result := discoveryWalkResult{}

	if candidate, found := builder.findCandidateForOwner(
		node.directory,
		node.ownerDir,
		node.canonicalDirectory,
	); found && candidate.directory != node.ownerDir {
		result.discoveredCandidate = true
		state, resolved := builder.loadStates[candidate.path]
		if !resolved {
			result.pending = &suspendedDiscoveryNode{node: node, candidate: candidate}
			return result
		}
		if state.failure == nil {
			result.activation = state
			node.ownerDir = candidate.directory
		}
	}
	result.directoriesVisited = 1
	if node.targets != nil && len(node.targets.children) == 0 {
		return result
	}

	entries := builder.fs.GetAccessibleEntries(node.directory)
	directories := append([]string(nil), entries.Directories...)
	sort.Strings(directories)
	children := make([]discoveryWalkNode, 0, len(directories))
	for _, name := range directories {
		var childTargets *discoveryTargetTrie
		if node.targets != nil {
			childTargets = node.targets.child(name, builder.fs.UseCaseSensitiveFileNames())
			if childTargets == nil {
				continue
			}
		}
		if isDefaultExcludedDirName(name, builder.fs.UseCaseSensitiveFileNames()) {
			continue
		}
		if entries.Symlinks != nil {
			if _, symlink := entries.Symlinks[name]; symlink {
				continue
			}
		}
		child := tspath.CombinePaths(node.directory, name)
		canonicalChild := ""
		if node.canonicalDirectory != "" {
			canonicalChild = tspath.CombinePaths(node.canonicalDirectory, name)
		}
		if builder.isGloballyIgnoredDirectory(node.ownerDir, child, canonicalChild) {
			result.directoriesPruned++
			continue
		}
		children = append(children, discoveryWalkNode{
			directory:          child,
			canonicalDirectory: canonicalChild,
			ownerDir:           node.ownerDir,
			targets:            childTargets,
		})
	}
	result.children = children
	return result
}

func (builder *configCatalogBuilder) isGloballyIgnoredDirectory(ownerDir string, directory string, canonicalDirectory string) bool {
	relative, patterns, ok := builder.authoredIgnoreRelativePath(ownerDir, directory, canonicalDirectory)
	return ok && len(patterns) > 0 && isDirAbsolutelyBlocked(relative, patterns)
}

func (builder *configCatalogBuilder) isGloballyIgnoredCandidate(ownerDir string, candidatePath string, canonicalPath string) bool {
	relative, patterns, ok := builder.authoredIgnoreRelativePath(ownerDir, candidatePath, canonicalPath)
	if !ok || len(patterns) == 0 {
		return false
	}
	return isDirBlockedByIgnores(relative, patterns, "") || isFileIgnored(relative, patterns, "")
}

func (builder *configCatalogBuilder) authoredIgnoreRelativePath(ownerDir string, targetPath string, canonicalPath string) (string, []IgnorePattern, bool) {
	if ownerDir == "" {
		return "", nil, false
	}
	config, ok := builder.configs[ownerDir]
	if !ok {
		state := builder.loadStateForDirectory(ownerDir)
		if state == nil || state.failure != nil {
			return "", nil, false
		}
		config = state.entries
	}
	relative, ok := relativeConfigPath(targetPath, ownerDir, builder.fs.UseCaseSensitiveFileNames())
	if !ok && canonicalPath != "" {
		matchPath, matchOwnerDir := ResolveConfigPathSpaceWithCanonical(targetPath, canonicalPath, ownerDir, builder.fs)
		relative, ok = relativeConfigPath(matchPath, matchOwnerDir, true)
	}
	if !ok || relative == "" {
		return "", nil, false
	}
	relative = strings.ReplaceAll(tspath.NormalizePath(relative), "\\", "/")
	patterns := extractConfigIgnores(config)
	return relative, patterns, true
}

func (builder *configCatalogBuilder) activate(state *configLoadState, explicitOnly bool, explicitConfig bool) {
	if state == nil || state.failure != nil {
		return
	}
	directory := state.candidate.directory
	builder.configs[directory] = append(RslintConfig(nil), state.entries...)
	source, exists := builder.sources[directory]
	if !exists {
		source = ConfigSource{
			CandidateID:     state.id,
			Path:            state.candidate.path,
			Directory:       directory,
			Fingerprint:     state.fingerprint,
			EslintPlugins:   cloneEslintPluginEntries(state.eslintPlugins),
			HasPluginConfig: state.hasPluginConfig,
			ExplicitOnly:    explicitOnly,
			ExplicitConfig:  explicitConfig,
		}
	} else {
		source.ExplicitOnly = source.ExplicitOnly && explicitOnly
		source.ExplicitConfig = source.ExplicitConfig || explicitConfig
	}
	builder.sources[directory] = source
}

func (builder *configCatalogBuilder) loadStateForDirectory(directory string) *configLoadState {
	for _, state := range builder.loadStates {
		if state.candidate.directory == directory && state.failure == nil {
			return state
		}
	}
	return nil
}

func (builder *configCatalogBuilder) catalog() (*ConfigCatalog, error) {
	if err := builder.validateConfigDirectoryIdentities(); err != nil {
		return nil, err
	}
	// ExplicitOnly is derived from every activation path, not just from the
	// explicit-file scope itself. Publish the final source value with the scope
	// so target discovery can keep this config out of automatic ownership and
	// handoff decisions while still assigning its literal files to it.
	for directory, source := range builder.sources {
		scope := builder.scopes[directory]
		scope.ExplicitOnly = source.ExplicitOnly
		builder.scopes[directory] = scope
	}
	failures := make([]ConfigFailure, 0, len(builder.failureByPath))
	for _, failure := range builder.failureByPath {
		failures = append(failures, failure)
	}
	sort.Slice(failures, func(i, j int) bool { return failures[i].Path < failures[j].Path })
	effectiveIDs := make([]string, 0, len(builder.sources))
	for _, source := range builder.sources {
		effectiveIDs = append(effectiveIDs, source.CandidateID)
	}
	sort.Strings(effectiveIDs)
	eslintPlugins := aggregateEffectiveEslintPlugins(builder.sources)
	if activator, ok := builder.loader.(ConfigModuleActivator); ok && builder.hadCandidates && len(builder.loadStates) > 0 {
		response, err := activator.ActivateConfigs(builder.ctx, ConfigActivationRequest{
			ProtocolVersion:    ConfigDiscoveryProtocolVersion,
			TransactionID:      builder.transactionID,
			EffectiveConfigIDs: effectiveIDs,
		})
		if err != nil {
			return nil, err
		}
		if response.TransactionID != builder.transactionID {
			return nil, configDiscoveryProtocolError("activation transaction mismatch: got %q, want %q", response.TransactionID, builder.transactionID)
		}
		eslintPlugins = cloneEslintPluginEntries(response.EslintPluginEntries)
	}
	return &ConfigCatalog{
		Generation:         builder.generation,
		TransactionID:      builder.transactionID,
		Configs:            builder.configs,
		Sources:            builder.sources,
		EffectiveConfigIDs: effectiveIDs,
		EslintPlugins:      eslintPlugins,
		Scopes:             builder.scopes,
		Failures:           failures,
		Stats:              builder.stats,
		Resolver:           NewConfigOwnerResolver(builder.configs, builder.fs),
	}, nil
}

// validateConfigDirectoryIdentities rejects ambiguous ownership before Node
// activates the effective candidate set. Alternate native casing is safe: its
// lexical spelling differs only by case and Realpath verified one exact
// physical directory. Other aliases (notably distinct symlink roots) must stay
// distinct and therefore cannot both govern that physical location.
func (builder *configCatalogBuilder) validateConfigDirectoryIdentities() error {
	directories := make([]string, 0, len(builder.configs))
	for directory := range builder.configs {
		directories = append(directories, tspath.NormalizePath(directory))
	}
	sort.Strings(directories)

	lexicalByPhysical := make(map[tspath.Path]string, len(directories))
	for _, directory := range directories {
		physicalDirectory := directory
		if realPath := builder.fs.Realpath(directory); realPath != "" {
			physicalDirectory = tspath.NormalizePath(realPath)
		}
		physicalID := tspath.ToPath(physicalDirectory, "", true)
		existing, collision := lexicalByPhysical[physicalID]
		if !collision {
			lexicalByPhysical[physicalID] = directory
			continue
		}
		if !builder.fs.UseCaseSensitiveFileNames() && strings.EqualFold(existing, directory) {
			continue
		}
		return fmt.Errorf(
			"config directories %q and %q resolve to the same filesystem location %q",
			existing,
			directory,
			physicalDirectory,
		)
	}
	return nil
}

func cloneEslintPluginEntries(entries []EslintPluginEntry) []EslintPluginEntry {
	if len(entries) == 0 {
		return nil
	}
	cloned := make([]EslintPluginEntry, len(entries))
	for index, entry := range entries {
		cloned[index] = EslintPluginEntry{
			Prefix:    entry.Prefix,
			RuleNames: append([]string(nil), entry.RuleNames...),
		}
	}
	return cloned
}

func aggregateEffectiveEslintPlugins(sources map[string]ConfigSource) []EslintPluginEntry {
	byPrefix := make(map[string]map[string]struct{})
	for _, source := range sources {
		for _, plugin := range source.EslintPlugins {
			rules := byPrefix[plugin.Prefix]
			if rules == nil {
				rules = make(map[string]struct{})
				byPrefix[plugin.Prefix] = rules
			}
			for _, ruleName := range plugin.RuleNames {
				rules[ruleName] = struct{}{}
			}
		}
	}
	prefixes := make([]string, 0, len(byPrefix))
	for prefix := range byPrefix {
		prefixes = append(prefixes, prefix)
	}
	sort.Strings(prefixes)
	entries := make([]EslintPluginEntry, 0, len(prefixes))
	for _, prefix := range prefixes {
		ruleNames := make([]string, 0, len(byPrefix[prefix]))
		for ruleName := range byPrefix[prefix] {
			ruleNames = append(ruleNames, ruleName)
		}
		sort.Strings(ruleNames)
		entries = append(entries, EslintPluginEntry{Prefix: prefix, RuleNames: ruleNames})
	}
	return entries
}

func (builder *configCatalogBuilder) allConfigsFailedError() error {
	failures := make([]ConfigFailure, 0, len(builder.failureByPath))
	for _, failure := range builder.failureByPath {
		failures = append(failures, failure)
	}
	sort.Slice(failures, func(i, j int) bool { return failures[i].Path < failures[j].Path })
	if len(failures) == 0 {
		return ErrAllConfigsFailed
	}
	failure := failures[0]
	switch failure.Kind {
	case "invalid":
		return fmt.Errorf("%w: invalid config in %s: %s", ErrAllConfigsFailed, failure.Path, failure.Message)
	default:
		return fmt.Errorf("%w: failed to load config %s: %s", ErrAllConfigsFailed, failure.Path, failure.Message)
	}
}

func appendUniqueSortedPath(paths []string, path string) []string {
	for _, existing := range paths {
		if existing == path {
			return paths
		}
	}
	paths = append(paths, path)
	sort.Strings(paths)
	return paths
}

func deduplicateWalkNodes(nodes []discoveryWalkNode) []discoveryWalkNode {
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].directory < nodes[j].directory })
	result := nodes[:0]
	for _, node := range nodes {
		if len(result) > 0 && result[len(result)-1].directory == node.directory {
			continue
		}
		result = append(result, node)
	}
	return result
}

func isDefaultDiscoveryExcluded(path string, cwd string, useCaseSensitive bool) bool {
	return IsDefaultExcludedPath(path, cwd, useCaseSensitive)
}
