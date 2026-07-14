package discovery

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"

	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
)

type configCandidate struct {
	path      string
	directory string
}

type configLoadState struct {
	id            string
	candidate     configCandidate
	entries       rslintconfig.RslintConfig
	ignoreMatcher rslintconfig.GlobalIgnoreMatcher
	eslintPlugins []rslintconfig.EslintPluginEntry
	failure       *ConfigModuleError
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
	ownerPath          string
	done               bool
}

type discoveryWalkNode struct {
	directory          string
	canonicalDirectory string
	ownerDir           string
	ownerPath          string
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
	transactionID string

	loadStates    map[string]*configLoadState
	configs       map[string]rslintconfig.RslintConfig
	sources       map[string]configSource
	scopes        map[string]rslintconfig.LintDiscoveryScope
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
		if err := builder.activate(state, false); err != nil {
			return nil, err
		}
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
		state := builder.loadStates[seed.ownerPath]
		if state == nil {
			continue
		}
		if err := builder.activate(state, seed.explicitFile); err != nil {
			return nil, err
		}
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
			ownerPath:          seed.ownerPath,
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
			if discoveryPathsEqual(root, parent, builder.fs.UseCaseSensitiveFileNames()) ||
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
			relative, within := rslintconfig.RelativePathWithinConfigRoot(file.Path, root, useCaseSensitive)
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
				if resolution.seed.ownerPath != "" && builder.isGloballyIgnoredDirectory(
					resolution.seed.ownerPath,
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
					resolution.seed.ownerPath,
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
				resolution.seed.ownerPath = candidate.path
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
				seed.ownerPath = candidate.path
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
	ownerPath string,
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
		if builder.isGloballyIgnoredCandidate(ownerPath, candidatePath, canonicalCandidate) {
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
		return errors.New("javascript config candidates require a module loader")
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
		id := fmt.Sprintf("config-%06d", builder.nextRequestID)
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
			id:            wireCandidate.ID,
			candidate:     candidate,
			eslintPlugins: cloneEslintPluginEntries(result.EslintPlugins),
		}
		if result.Status == "failed" {
			failure := *result.Error
			state.failure = &failure
		} else if err := rslintconfig.ValidateConfig(result.Entries); err != nil {
			state.failure = &ConfigModuleError{Code: "invalid", Message: err.Error()}
		} else {
			state.entries = append(rslintconfig.RslintConfig(nil), result.Entries...)
			// Bind authored-ignore semantics to the exact loaded candidate. One
			// directory can be admitted through automatic and literal contexts
			// with different candidate filenames; directory-only matcher storage
			// would make the later load silently replace the active owner's policy.
			state.ignoreMatcher = rslintconfig.NewGlobalIgnoreMatcher(
				state.entries,
				candidate.directory,
				builder.fs,
			)
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
				if err := builder.activate(result.activation, false); err != nil {
					return err
				}
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
					if err := builder.activate(state, false); err != nil {
						return err
					}
					item.node.ownerDir = state.candidate.directory
					item.node.ownerPath = state.candidate.path
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
	if builder.isGloballyIgnoredDirectory(node.ownerPath, node.directory, node.canonicalDirectory) {
		return discoveryWalkResult{directoriesPruned: 1}
	}
	result := discoveryWalkResult{}

	if candidate, found := builder.findCandidateForOwner(
		node.directory,
		node.ownerPath,
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
			node.ownerPath = candidate.path
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
		if rslintconfig.IsDefaultExcludedPath(name, "", builder.fs.UseCaseSensitiveFileNames()) {
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
		if builder.isGloballyIgnoredDirectory(node.ownerPath, child, canonicalChild) {
			result.directoriesPruned++
			continue
		}
		children = append(children, discoveryWalkNode{
			directory:          child,
			canonicalDirectory: canonicalChild,
			ownerDir:           node.ownerDir,
			ownerPath:          node.ownerPath,
			targets:            childTargets,
		})
	}
	result.children = children
	return result
}

func (builder *configCatalogBuilder) isGloballyIgnoredDirectory(ownerPath string, directory string, canonicalDirectory string) bool {
	matcher, ok := builder.globalIgnoreMatcher(ownerPath)
	return ok && matcher.BlocksDirectory(directory, canonicalDirectory)
}

func (builder *configCatalogBuilder) isGloballyIgnoredCandidate(ownerPath string, candidatePath string, canonicalPath string) bool {
	matcher, ok := builder.globalIgnoreMatcher(ownerPath)
	return ok && matcher.IgnoresPath(candidatePath, canonicalPath)
}

func (builder *configCatalogBuilder) globalIgnoreMatcher(ownerPath string) (rslintconfig.GlobalIgnoreMatcher, bool) {
	if ownerPath == "" {
		return rslintconfig.GlobalIgnoreMatcher{}, false
	}
	state := builder.loadStates[ownerPath]
	if state == nil || state.failure != nil {
		return rslintconfig.GlobalIgnoreMatcher{}, false
	}
	return state.ignoreMatcher, true
}

func (builder *configCatalogBuilder) activate(state *configLoadState, explicitOnly bool) error {
	if state == nil || state.failure != nil {
		return nil
	}
	directory := state.candidate.directory
	source, exists := builder.sources[directory]
	if !exists {
		builder.installSource(state, explicitOnly)
		return nil
	}
	if source.CandidateID == state.id {
		source.ExplicitOnly = source.ExplicitOnly && explicitOnly
		builder.sources[directory] = source
		return nil
	}
	if source.ExplicitOnly && !explicitOnly {
		// A literal target may bypass its parent's authored ignore and discover a
		// different filename in this directory. Once an automatic route reaches
		// the directory, that route defines its single shared config boundary.
		builder.installSource(state, false)
		return nil
	}
	if !source.ExplicitOnly && explicitOnly {
		// Keep the automatic candidate, while the caller still records the
		// literal file in this directory's scope below.
		return nil
	}
	return fmt.Errorf(
		"ambiguous config candidates %q and %q for directory %q",
		source.CandidatePath,
		state.candidate.path,
		directory,
	)
}

func (builder *configCatalogBuilder) installSource(state *configLoadState, explicitOnly bool) {
	directory := state.candidate.directory
	builder.configs[directory] = append(rslintconfig.RslintConfig(nil), state.entries...)
	builder.sources[directory] = configSource{
		CandidateID:   state.id,
		CandidatePath: state.candidate.path,
		EslintPlugins: cloneEslintPluginEntries(state.eslintPlugins),
		ExplicitOnly:  explicitOnly,
	}
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
		TransactionID:      builder.transactionID,
		Configs:            builder.configs,
		EffectiveConfigIDs: effectiveIDs,
		EslintPlugins:      eslintPlugins,
		Scopes:             builder.scopes,
		Failures:           failures,
		Stats:              builder.stats,
		Explicit:           builder.request.Mode == ConfigDiscoveryExplicit,
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
		return fmt.Errorf( //nolint:staticcheck // Preserve the established user-facing JS API error contract.
			"Config directories %q and %q resolve to the same filesystem location %q",
			existing,
			directory,
			physicalDirectory,
		)
	}
	return nil
}

func cloneEslintPluginEntries(entries []rslintconfig.EslintPluginEntry) []rslintconfig.EslintPluginEntry {
	if len(entries) == 0 {
		return nil
	}
	cloned := make([]rslintconfig.EslintPluginEntry, len(entries))
	for index, entry := range entries {
		cloned[index] = rslintconfig.EslintPluginEntry{
			Prefix:    entry.Prefix,
			RuleNames: append([]string(nil), entry.RuleNames...),
		}
	}
	return cloned
}

func aggregateEffectiveEslintPlugins(sources map[string]configSource) []rslintconfig.EslintPluginEntry {
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
	entries := make([]rslintconfig.EslintPluginEntry, 0, len(prefixes))
	for _, prefix := range prefixes {
		ruleNames := make([]string, 0, len(byPrefix[prefix]))
		for ruleName := range byPrefix[prefix] {
			ruleNames = append(ruleNames, ruleName)
		}
		sort.Strings(ruleNames)
		entries = append(entries, rslintconfig.EslintPluginEntry{Prefix: prefix, RuleNames: ruleNames})
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
	displayPath := filepath.FromSlash(failure.Path)
	switch failure.Kind {
	case "invalid":
		return fmt.Errorf("%w: invalid config in %s: %s", ErrAllConfigsFailed, displayPath, failure.Message)
	default:
		return fmt.Errorf("%w: failed to load config %s: %s", ErrAllConfigsFailed, displayPath, failure.Message)
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
	return rslintconfig.IsDefaultExcludedPath(path, cwd, useCaseSensitive)
}

func discoveryPathsEqual(a string, b string, useCaseSensitive bool) bool {
	if useCaseSensitive {
		return a == b
	}
	return strings.EqualFold(a, b)
}
