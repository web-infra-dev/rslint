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
	"github.com/web-infra-dev/rslint/internal/config/gitignore"
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
	gitDirectory       string
	gitCursor          gitignore.Cursor
	gitActive          bool
	done               bool
}

type discoveryWalkNode struct {
	directory          string
	canonicalDirectory string
	ownerDir           string
	ownerPath          string
	gitDirectory       string
	gitCursor          gitignore.Cursor
	gitActive          bool
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
	children             []discoveryWalkNode
	pending              *suspendedDiscoveryNode
	activation           *configLoadState
	gitignoreObservation *gitignoreObservation
	directoriesVisited   int
	directoriesPruned    int
	discoveredCandidate  bool
	err                  error
}

type directorySeedResolution struct {
	seed            *discoverySeed
	candidates      []configCandidate
	next            int
	gitDirectory    string
	gitCursor       gitignore.Cursor
	gitActive       bool
	configReachable bool
}

type gitignoreObservation struct {
	ownerDirectory  string
	sourceDirectory string
	globs           []string
}

type gitignoreReadResult struct {
	content string
	exists  bool
}

type configCatalogBuilder struct {
	ctx                context.Context
	fs                 vfs.FS
	loader             ConfigModuleLoader
	request            ConfigDiscoveryRequest
	explicitConfigPath string
	transactionID      string

	loadStates           map[string]*configLoadState
	loadStateByIdentity  map[tspath.Path]*configLoadState
	configs              map[string]rslintconfig.RslintConfig
	sources              map[string]configSource
	scopes               map[string]rslintconfig.LintDiscoveryScope
	failureByPath        map[string]ConfigFailure
	gitignoreSources     map[string]map[tspath.Path]gitignoreObservation
	gitignoreReadMu      sync.Mutex
	gitignoreReadCache   map[tspath.Path]gitignoreReadResult
	gitignoreReadPending map[tspath.Path]chan struct{}
	hadCandidates        bool
	nextRequestID        int
	stats                ConfigDiscoveryStats
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

	if builder.explicitConfigPath != "" {
		candidate := configCandidate{
			path: normalizeDiscoveryPath(builder.explicitConfigPath, cwd),
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
	builder.collectExplicitSeedGitignore(explicitSeeds)

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
			gitDirectory:       seed.gitDirectory,
			gitCursor:          seed.gitCursor,
			gitActive:          seed.gitActive,
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
	useCaseSensitive := builder.fs.UseCaseSensitiveFileNames()
	byIdentity := make(map[tspath.Path]string, len(raw))
	for _, directory := range raw {
		directory = normalizeDiscoveryPath(directory, builder.request.CWD)
		identity := tspath.ToPath(directory, "", useCaseSensitive)
		if current, exists := byIdentity[identity]; !exists || directory < current {
			byIdentity[identity] = directory
		}
	}
	roots := make([]string, 0, len(byIdentity))
	for _, directory := range byIdentity {
		roots = append(roots, directory)
	}
	sort.Strings(roots)
	return roots
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

// targetAncestorTries maps each directory root to the lexical paths
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
		resolutions = append(resolutions, directorySeedResolution{
			seed:            seed,
			candidates:      candidates,
			configReachable: true,
		})
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
				if resolution.gitActive {
					rootDirectory := resolution.seed.searchDir
					if resolution.seed.usingCanonical {
						rootDirectory = resolution.seed.canonicalSearchDir
					}
					reachable, authoredBlocked := builder.advanceDirectorySeedGit(
						resolution,
						candidateDirectory,
						discoveryPathsEqual(
							candidateDirectory,
							rootDirectory,
							builder.fs.UseCaseSensitiveFileNames(),
						),
					)
					if authoredBlocked {
						resolution.gitActive = false
						resolution.next = len(resolution.candidates)
						break
					}
					if !reachable {
						// Only the supplied root itself may reopen an inherited
						// Git-inaccessible ancestry. Hidden intermediate configs are
						// never evaluated.
						resolution.next++
						continue
					}
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
			break
		}
		if err := builder.ensureCandidates(candidates); err != nil {
			return err
		}
		for _, resolution := range pending {
			candidate := resolution.candidates[resolution.next]
			state := builder.loadStates[candidate.path]
			if state != nil && state.failure == nil {
				resolution.seed.ownerDir = state.candidate.directory
				resolution.seed.ownerPath = state.candidate.path
				resolution.gitCursor = gitignore.NewCursor(
					state.candidate.directory,
					builder.fs.UseCaseSensitiveFileNames(),
				)
				resolution.gitDirectory = state.candidate.directory
				resolution.gitActive = true
				resolution.configReachable = true
			}
			resolution.next++
		}
	}

	for index := range resolutions {
		resolution := &resolutions[index]
		if !resolution.gitActive {
			continue
		}
		rootDirectory := resolution.seed.searchDir
		if resolution.seed.usingCanonical {
			rootDirectory = resolution.seed.canonicalSearchDir
		}
		_, authoredBlocked := builder.advanceDirectorySeedGit(resolution, rootDirectory, true)
		if authoredBlocked {
			continue
		}
		resolution.seed.gitDirectory = resolution.gitDirectory
		resolution.seed.gitCursor = resolution.gitCursor
		resolution.seed.gitActive = true
	}
	return nil
}

// advanceDirectorySeedGit advances one requested directory root through the
// current owner's Git path space without reading the destination's local
// .gitignore. A candidate in the destination is therefore evaluated first; a
// successful candidate can reset ownership before that source is observed.
func (builder *configCatalogBuilder) advanceDirectorySeedGit(
	resolution *directorySeedResolution,
	destination string,
	reopenDestination bool,
) (reachable bool, authoredBlocked bool) {
	if resolution == nil || !resolution.gitActive {
		return true, false
	}
	useCaseSensitive := builder.fs.UseCaseSensitiveFileNames()
	destination = tspath.NormalizePath(destination)
	if discoveryPathsEqual(resolution.gitDirectory, destination, useCaseSensitive) {
		if reopenDestination {
			resolution.configReachable = true
		}
		return resolution.configReachable, false
	}
	relative, within := rslintconfig.RelativePathWithinConfigRoot(
		destination,
		resolution.gitDirectory,
		useCaseSensitive,
	)
	if !within {
		resolution.gitCursor, _ = resolution.gitCursor.Enter(destination)
		resolution.gitDirectory = destination
		resolution.configReachable = false
		return false, false
	}

	current := resolution.gitDirectory
	for _, component := range splitDiscoveryPath(relative) {
		resolution.gitCursor = builder.observeGitignoreSource(
			resolution.seed.ownerDir,
			current,
			current,
			resolution.gitCursor,
		)

		nextDirectory := tspath.CombinePaths(current, component)
		var gitBlocked bool
		resolution.gitCursor, gitBlocked = resolution.gitCursor.Enter(nextDirectory)
		resolution.gitDirectory = nextDirectory

		if builder.isGloballyIgnoredDirectory(
			resolution.seed.ownerPath,
			nextDirectory,
			nextDirectory,
		) {
			resolution.configReachable = false
			return false, true
		}
		if gitBlocked && !builder.reopensGitignoredDirectory(
			resolution.seed.ownerPath,
			nextDirectory,
			nextDirectory,
		) {
			resolution.configReachable = false
		}
		if reopenDestination &&
			discoveryPathsEqual(nextDirectory, destination, useCaseSensitive) {
			resolution.configReachable = true
		}
		current = nextDirectory
	}
	return resolution.configReachable, false
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
				seed.ownerDir = state.candidate.directory
				seed.ownerPath = state.candidate.path
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
	type candidateGroup struct {
		candidate configCandidate
		aliases   map[string]configCandidate
	}
	groups := make(map[tspath.Path]*candidateGroup, len(rawCandidates))
	for _, candidate := range rawCandidates {
		candidate.path = tspath.NormalizePath(candidate.path)
		candidate.directory = tspath.NormalizePath(candidate.directory)
		identity := tspath.ToPath(candidate.path, "", builder.fs.UseCaseSensitiveFileNames())
		if state := builder.loadStateByIdentity[identity]; state != nil {
			if err := builder.validateNativeCaseAlias(state.candidate, candidate); err != nil {
				return err
			}
			builder.loadStates[candidate.path] = state
			continue
		}
		group := groups[identity]
		if group == nil {
			group = &candidateGroup{
				candidate: candidate,
				aliases:   make(map[string]configCandidate),
			}
			groups[identity] = group
		} else {
			if err := builder.validateNativeCaseAlias(group.candidate, candidate); err != nil {
				return err
			}
			if candidate.path < group.candidate.path {
				group.candidate = candidate
			}
		}
		group.aliases[candidate.path] = candidate
	}
	if len(groups) == 0 {
		return nil
	}
	if builder.loader == nil {
		return errors.New("javascript config candidates require a module loader")
	}

	paths := make([]string, 0, len(groups))
	groupByPath := make(map[string]*candidateGroup, len(groups))
	for _, group := range groups {
		paths = append(paths, group.candidate.path)
		groupByPath[group.candidate.path] = group
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
		candidate := groupByPath[path].candidate
		builder.nextRequestID++
		id := fmt.Sprintf("config-%06d", builder.nextRequestID)
		wireCandidate := ConfigLoadCandidate{
			ID:              id,
			ConfigPath:      candidate.path,
			ConfigDirectory: candidate.directory,
		}
		request.Candidates = append(request.Candidates, wireCandidate)
	}
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
		group := groupByPath[wireCandidate.ConfigPath]
		candidate := group.candidate
		result := results[wireCandidate.ID]
		state := &configLoadState{
			id:        wireCandidate.ID,
			candidate: candidate,
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
		identity := tspath.ToPath(candidate.path, "", builder.fs.UseCaseSensitiveFileNames())
		builder.loadStateByIdentity[identity] = state
		for aliasPath := range group.aliases {
			builder.loadStates[aliasPath] = state
		}
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

// validateNativeCaseAlias protects the pre-load case-folding optimization
// with the same physical-directory invariant used by the final catalog. On a
// case-insensitive filesystem, alternate casing should execute one module;
// unrelated lexical or symlink aliases must not be silently merged.
func (builder *configCatalogBuilder) validateNativeCaseAlias(left configCandidate, right configCandidate) error {
	if left.path == right.path {
		return nil
	}
	if builder.fs.UseCaseSensitiveFileNames() || !strings.EqualFold(left.path, right.path) {
		return fmt.Errorf("config candidate identity collision between %q and %q", left.path, right.path)
	}
	physicalPath := func(path string) string {
		if realPath := builder.fs.Realpath(path); realPath != "" {
			return tspath.NormalizePath(realPath)
		}
		return tspath.NormalizePath(path)
	}
	leftPhysicalDirectory := physicalPath(left.directory)
	rightPhysicalDirectory := physicalPath(right.directory)
	leftPhysicalPath := physicalPath(left.path)
	rightPhysicalPath := physicalPath(right.path)
	if tspath.ToPath(leftPhysicalDirectory, "", true) != tspath.ToPath(rightPhysicalDirectory, "", true) ||
		tspath.ToPath(leftPhysicalPath, "", true) != tspath.ToPath(rightPhysicalPath, "", true) {
		return fmt.Errorf(
			"config candidates %q and %q differ only by case but resolve to distinct filesystem paths",
			left.path,
			right.path,
		)
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
			if result.gitignoreObservation != nil {
				builder.recordGitignoreObservation(*result.gitignoreObservation)
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
					item.node.gitCursor = gitignore.NewCursor(
						state.candidate.directory,
						builder.fs.UseCaseSensitiveFileNames(),
					)
					item.node.gitDirectory = state.candidate.directory
					item.node.gitActive = true
				}
				next = append(next, item.node)
			}
		}
		// Overlapping requested roots are intentionally independent routes. The
		// same lexical directory can carry inherited Git state on one route and
		// explicit-root reachability on another, so directory-only deduplication
		// would be incorrect.
		queue = next
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
			node.ownerDir = state.candidate.directory
			node.ownerPath = state.candidate.path
			node.gitCursor = gitignore.NewCursor(
				state.candidate.directory,
				builder.fs.UseCaseSensitiveFileNames(),
			)
			node.gitDirectory = state.candidate.directory
			node.gitActive = true
		}
	}
	result.directoriesVisited = 1
	if node.gitActive {
		nextCursor, observation := builder.readGitignoreSource(
			node.ownerDir,
			node.directory,
			node.gitDirectory,
			node.gitCursor,
		)
		node.gitCursor = nextCursor
		result.gitignoreObservation = observation
	}
	if node.targets != nil && len(node.targets.children) == 0 {
		return result
	}

	entries := builder.fs.GetAccessibleEntries(node.directory)
	directories := append([]string(nil), entries.Directories...)
	sort.Strings(directories)
	parentRealPath := ""
	if entries.Symlinks == nil && len(directories) > 0 {
		parentRealPath = builder.fs.Realpath(node.directory)
	}
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
		if builder.isSymlinkDirectoryChild(node.directory, parentRealPath, name, entries) {
			continue
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
		childGitDirectory := ""
		childGitCursor := gitignore.Cursor{}
		childGitActive := false
		if node.gitActive {
			childGitDirectory = tspath.CombinePaths(node.gitDirectory, name)
			nextCursor, gitBlocked := node.gitCursor.Enter(childGitDirectory)
			if gitBlocked && !builder.reopensGitignoredDirectory(
				node.ownerPath,
				child,
				canonicalChild,
			) {
				result.directoriesPruned++
				continue
			}
			childGitCursor = nextCursor
			childGitActive = true
		}
		children = append(children, discoveryWalkNode{
			directory:          child,
			canonicalDirectory: canonicalChild,
			ownerDir:           node.ownerDir,
			ownerPath:          node.ownerPath,
			gitDirectory:       childGitDirectory,
			gitCursor:          childGitCursor,
			gitActive:          childGitActive,
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

func (builder *configCatalogBuilder) reopensGitignoredDirectory(ownerPath string, directory string, canonicalDirectory string) bool {
	matcher, ok := builder.globalIgnoreMatcher(ownerPath)
	return ok && matcher.ReopensDirectoryNode(directory, canonicalDirectory)
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

func (builder *configCatalogBuilder) readGitignoreSource(
	ownerDirectory string,
	sourceDirectory string,
	matchDirectory string,
	cursor gitignore.Cursor,
) (gitignore.Cursor, *gitignoreObservation) {
	if ownerDirectory == "" || sourceDirectory == "" || matchDirectory == "" || !cursor.SourceReachable() {
		return cursor, nil
	}
	content := builder.readGitignoreFile(sourceDirectory)
	if !content.exists {
		return cursor, nil
	}
	next, globs := cursor.AppendSource(matchDirectory, content.content)
	if len(globs) == 0 {
		return next, nil
	}
	return next, &gitignoreObservation{
		ownerDirectory:  ownerDirectory,
		sourceDirectory: matchDirectory,
		globs:           globs,
	}
}

// readGitignoreFile freezes each lexical source for one catalog generation.
// Config modules can have side effects, and overlapping discovery routes may
// reach the same source on different frontiers; rereading would let one
// transaction make ownership decisions from different bytes. Different source
// paths still read concurrently; only duplicate in-flight reads wait for the
// first route, including a cached miss.
func (builder *configCatalogBuilder) readGitignoreFile(sourceDirectory string) gitignoreReadResult {
	path := tspath.CombinePaths(sourceDirectory, ".gitignore")
	identity := tspath.ToPath(
		tspath.NormalizePath(path),
		"",
		builder.fs.UseCaseSensitiveFileNames(),
	)
	for {
		builder.gitignoreReadMu.Lock()
		if builder.gitignoreReadCache == nil {
			builder.gitignoreReadCache = make(map[tspath.Path]gitignoreReadResult)
		}
		if builder.gitignoreReadPending == nil {
			builder.gitignoreReadPending = make(map[tspath.Path]chan struct{})
		}
		if result, exists := builder.gitignoreReadCache[identity]; exists {
			builder.gitignoreReadMu.Unlock()
			return result
		}
		if wait, pending := builder.gitignoreReadPending[identity]; pending {
			if wait == nil {
				wait = make(chan struct{})
				builder.gitignoreReadPending[identity] = wait
			}
			builder.gitignoreReadMu.Unlock()
			<-wait
			continue
		}
		builder.gitignoreReadPending[identity] = nil
		builder.gitignoreReadMu.Unlock()
		break
	}

	content, exists := builder.fs.ReadFile(path)
	result := gitignoreReadResult{content: content, exists: exists}
	builder.gitignoreReadMu.Lock()
	builder.gitignoreReadCache[identity] = result
	wait := builder.gitignoreReadPending[identity]
	delete(builder.gitignoreReadPending, identity)
	if wait != nil {
		close(wait)
	}
	builder.gitignoreReadMu.Unlock()
	return result
}

func (builder *configCatalogBuilder) observeGitignoreSource(
	ownerDirectory string,
	sourceDirectory string,
	matchDirectory string,
	cursor gitignore.Cursor,
) gitignore.Cursor {
	next, observation := builder.readGitignoreSource(
		ownerDirectory,
		sourceDirectory,
		matchDirectory,
		cursor,
	)
	if observation != nil {
		builder.recordGitignoreObservation(*observation)
	}
	return next
}

func (builder *configCatalogBuilder) recordGitignoreObservation(observation gitignoreObservation) {
	if observation.ownerDirectory == "" || observation.sourceDirectory == "" || len(observation.globs) == 0 {
		return
	}
	if builder.gitignoreSources == nil {
		builder.gitignoreSources = make(map[string]map[tspath.Path]gitignoreObservation)
	}
	sources := builder.gitignoreSources[observation.ownerDirectory]
	if sources == nil {
		sources = make(map[tspath.Path]gitignoreObservation)
		builder.gitignoreSources[observation.ownerDirectory] = sources
	}
	identity := tspath.ToPath(
		tspath.NormalizePath(observation.sourceDirectory),
		"",
		builder.fs.UseCaseSensitiveFileNames(),
	)
	if _, exists := sources[identity]; exists {
		return
	}
	observation.globs = append([]string(nil), observation.globs...)
	sources[identity] = observation
}

func splitDiscoveryPath(path string) []string {
	path = strings.ReplaceAll(tspath.NormalizePath(path), "\\", "/")
	var components []string
	for _, component := range strings.Split(path, "/") {
		if component == "" || component == "." {
			continue
		}
		components = append(components, component)
	}
	return components
}

// collectExplicitSeedGitignore records the exact directory chain for a literal
// target. Literal files bypass Git only while choosing their nearest config;
// the resulting target still uses that config owner's ordinary Git ignores.
// Keeping these observations source-scoped lets mixed directory+literal input
// merge them with the automatic walk without duplicating parent patterns.
func (builder *configCatalogBuilder) collectExplicitSeedGitignore(seeds []*discoverySeed) {
	useCaseSensitive := builder.fs.UseCaseSensitiveFileNames()
	cursorByOwnerAndDirectory := make(map[string]map[tspath.Path]gitignore.Cursor)
	barrierByOwnerAndDirectory := make(map[string]map[tspath.Path]struct{})
	for _, seed := range seeds {
		if seed == nil || seed.ownerDir == "" {
			continue
		}
		ownerCursors := cursorByOwnerAndDirectory[seed.ownerDir]
		if ownerCursors == nil {
			ownerCursors = make(map[tspath.Path]gitignore.Cursor)
			cursorByOwnerAndDirectory[seed.ownerDir] = ownerCursors
		}
		ownerBarriers := barrierByOwnerAndDirectory[seed.ownerDir]
		if ownerBarriers == nil {
			ownerBarriers = make(map[tspath.Path]struct{})
			barrierByOwnerAndDirectory[seed.ownerDir] = ownerBarriers
		}
		builder.collectExplicitSeedGitignoreChain(
			seed,
			ownerCursors,
			ownerBarriers,
			useCaseSensitive,
		)
	}
}

func (builder *configCatalogBuilder) collectExplicitSeedGitignoreChain(
	seed *discoverySeed,
	cursorByDirectory map[tspath.Path]gitignore.Cursor,
	barriers map[tspath.Path]struct{},
	useCaseSensitive bool,
) {
	targetDirectory := tspath.GetDirectoryPath(seed.path)
	matchDirectory := targetDirectory
	if _, within := rslintconfig.RelativePathWithinConfigRoot(
		matchDirectory,
		seed.ownerDir,
		useCaseSensitive,
	); !within {
		matchDirectory = seed.canonicalSearchDir
		if matchDirectory == "" {
			return
		}
		if _, within = rslintconfig.RelativePathWithinConfigRoot(
			matchDirectory,
			seed.ownerDir,
			useCaseSensitive,
		); !within {
			return
		}
	}

	relative, within := rslintconfig.RelativePathWithinConfigRoot(
		matchDirectory,
		seed.ownerDir,
		useCaseSensitive,
	)
	if !within {
		return
	}
	cursor := gitignore.NewCursor(seed.ownerDir, useCaseSensitive)
	currentMatch := seed.ownerDir
	currentSource := seed.ownerDir
	components := splitDiscoveryPath(relative)
	for index := 0; ; index++ {
		identity := tspath.ToPath(currentMatch, "", useCaseSensitive)
		if cached, exists := cursorByDirectory[identity]; exists {
			cursor = cached
		} else {
			cursor = builder.observeGitignoreSource(
				seed.ownerDir,
				currentSource,
				currentMatch,
				cursor,
			)
			cursorByDirectory[identity] = cursor
		}
		if index == len(components) || !cursor.SourceReachable() {
			return
		}
		component := components[index]
		nextSource := tspath.CombinePaths(currentSource, component)
		nextMatch := tspath.CombinePaths(currentMatch, component)
		nextIdentity := tspath.ToPath(nextMatch, "", useCaseSensitive)
		if _, blocked := barriers[nextIdentity]; blocked {
			return
		}
		if cached, exists := cursorByDirectory[nextIdentity]; exists {
			currentSource = nextSource
			currentMatch = nextMatch
			cursor = cached
			continue
		}
		entries := builder.fs.GetAccessibleEntries(currentSource)
		parentRealPath := ""
		if entries.Symlinks == nil {
			parentRealPath = builder.fs.Realpath(currentSource)
		}
		if builder.isSymlinkDirectoryChild(currentSource, parentRealPath, component, entries) {
			barriers[nextIdentity] = struct{}{}
			return
		}
		currentSource = nextSource
		currentMatch = nextMatch
		next, blocked := cursor.Enter(currentMatch)
		cursor = next
		if blocked {
			barriers[nextIdentity] = struct{}{}
			return
		}
	}
}

func (builder *configCatalogBuilder) isSymlinkDirectoryChild(
	parentDirectory string,
	parentRealPath string,
	name string,
	entries vfs.Entries,
) bool {
	if entries.Symlinks != nil {
		for symlink := range entries.Symlinks {
			if symlink == name ||
				(!builder.fs.UseCaseSensitiveFileNames() && strings.EqualFold(symlink, name)) {
				return true
			}
		}
		return false
	}
	childDirectory := tspath.CombinePaths(parentDirectory, name)
	childRealPath := builder.fs.Realpath(childDirectory)
	if parentRealPath == "" || childRealPath == "" {
		return false
	}
	expectedRealPath := tspath.CombinePaths(parentRealPath, name)
	return tspath.ComparePaths(childRealPath, expectedRealPath, tspath.ComparePathsOptions{
		UseCaseSensitiveFileNames: builder.fs.UseCaseSensitiveFileNames(),
	}) != 0
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
		ExplicitOnly:  explicitOnly,
	}
}

func (builder *configCatalogBuilder) catalog() (*ConfigCatalog, error) {
	if err := builder.validateConfigDirectoryIdentities(); err != nil {
		return nil, err
	}
	if builder.explicitConfigPath == "" {
		builder.materializeGitignoreConfigs()
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
	var eslintPlugins []rslintconfig.EslintPluginEntry
	if builder.hadCandidates && len(builder.loadStates) > 0 {
		response, err := builder.loader.ActivateConfigs(builder.ctx, ConfigActivationRequest{
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
		Explicit:           builder.explicitConfigPath != "",
	}, nil
}

func (builder *configCatalogBuilder) materializeGitignoreConfigs() {
	caseInsensitive := !builder.fs.UseCaseSensitiveFileNames()
	for ownerDirectory, entries := range builder.configs {
		sources := builder.gitignoreSources[ownerDirectory]
		if len(sources) == 0 {
			continue
		}
		observations := make([]gitignoreObservation, 0, len(sources))
		for _, observation := range sources {
			observations = append(observations, observation)
		}
		sort.Slice(observations, func(i, j int) bool {
			leftDepth := builder.gitignoreSourceDepth(ownerDirectory, observations[i].sourceDirectory)
			rightDepth := builder.gitignoreSourceDepth(ownerDirectory, observations[j].sourceDirectory)
			if leftDepth != rightDepth {
				return leftDepth < rightDepth
			}
			left := observations[i].sourceDirectory
			right := observations[j].sourceDirectory
			if caseInsensitive {
				leftFolded := strings.ToLower(left)
				rightFolded := strings.ToLower(right)
				if leftFolded != rightFolded {
					return leftFolded < rightFolded
				}
			}
			return left < right
		})
		var globs []string
		for _, observation := range observations {
			globs = append(globs, observation.globs...)
		}
		builder.configs[ownerDirectory] = rslintconfig.ConfigWithCollectedGitignore(
			entries,
			globs,
			caseInsensitive,
		)
	}
}

func (builder *configCatalogBuilder) gitignoreSourceDepth(ownerDirectory string, sourceDirectory string) int {
	relative, within := rslintconfig.RelativePathWithinConfigRoot(
		sourceDirectory,
		ownerDirectory,
		builder.fs.UseCaseSensitiveFileNames(),
	)
	if !within || relative == "" {
		return 0
	}
	return len(splitDiscoveryPath(relative))
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
	message := ""
	switch failure.Kind {
	case "invalid":
		message = fmt.Sprintf("%s: invalid config in %s: %s", ErrAllConfigsFailed, displayPath, failure.Message)
	default:
		message = fmt.Sprintf("%s: failed to load config %s: %s", ErrAllConfigsFailed, displayPath, failure.Message)
	}
	return &AllConfigsFailedError{Failures: failures, message: message}
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

func isDefaultDiscoveryExcluded(path string, cwd string, useCaseSensitive bool) bool {
	return rslintconfig.IsDefaultExcludedPath(path, cwd, useCaseSensitive)
}

func discoveryPathsEqual(a string, b string, useCaseSensitive bool) bool {
	if useCaseSensitive {
		return a == b
	}
	return strings.EqualFold(a, b)
}
