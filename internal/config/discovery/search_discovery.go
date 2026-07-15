package discovery

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"

	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
)

var (
	ErrConfigFileMissing = errors.New("could not find config file")
	ErrAllFilesIgnored   = errors.New("all files matched by a pattern are ignored")
)

type ConfigFileMissingError struct {
	Path string
}

func (err *ConfigFileMissingError) Error() string {
	if err == nil || err.Path == "" {
		return "Could not find config file."
	}
	return fmt.Sprintf("Could not find config file for %q.", err.Path)
}

func (err *ConfigFileMissingError) Unwrap() error { return ErrConfigFileMissing }

type AllFilesIgnoredError struct {
	Pattern string
}

func (err *AllFilesIgnoredError) Error() string {
	if err == nil {
		return ErrAllFilesIgnored.Error()
	}
	return fmt.Sprintf("All files matched by %q are ignored.", err.Pattern)
}

func (err *AllFilesIgnoredError) Unwrap() error { return ErrAllFilesIgnored }

type configCandidate struct {
	path      string
	directory string
}

type searchConfigLoadState struct {
	ready     chan struct{}
	id        string
	candidate configCandidate
	entries   rslintconfig.RslintConfig
	failure   *ConfigModuleError
	err       error
}

type searchOwner struct {
	directory   string
	config      rslintconfig.RslintConfig
	evaluator   *rslintconfig.ConfigEvaluator
	missing     bool
	unavailable bool
	source      *searchConfigLoadState
}

type searchOwnerState struct {
	ready chan struct{}
	owner *searchOwner
	err   error
}

type searchCatalogBuilder struct {
	ctx           context.Context
	fs            vfs.FS
	loader        ConfigModuleLoader
	request       ConfigDiscoveryRequest
	transactionID string

	mu                  sync.Mutex
	loadStates          map[tspath.Path]*searchConfigLoadState
	owners              map[tspath.Path]*searchOwnerState
	configs             map[string]rslintconfig.RslintConfig
	evaluators          map[string]*rslintconfig.ConfigEvaluator
	sourceByDirectory   map[string]*searchConfigLoadState
	nextRequestID       int
	stats               ConfigDiscoveryStats
	fixedOwner          *searchOwner
	fixedCandidate      *configCandidate
	missingConfigPaths  []string
	missingConfigPathID map[string]struct{}
	failures            []ConfigFailure
	failurePathIDs      map[string]struct{}
	cleanupDeferred     bool

	predicateOnce sync.Once
	predicate     *predicateCoordinator
}

type explicitSearchResult struct {
	file    ExplicitFileSearch
	owner   *searchOwner
	target  *DiscoveredTarget
	input   ExplicitInputResult
	missing string
	err     error
}

type globSearchResult struct {
	targets []DiscoveredTarget
	missing []string
	err     error
}

type searchDirectoryEntry struct {
	name      string
	directory bool
}

func buildSearchCatalog(
	ctx context.Context,
	fsys vfs.FS,
	loader ConfigModuleLoader,
	request ConfigDiscoveryRequest,
	transactionID string,
) (*ConfigCatalog, error) {
	builder := &searchCatalogBuilder{
		ctx:                 ctx,
		fs:                  fsys,
		loader:              loader,
		request:             request,
		transactionID:       transactionID,
		loadStates:          make(map[tspath.Path]*searchConfigLoadState),
		owners:              make(map[tspath.Path]*searchOwnerState),
		configs:             make(map[string]rslintconfig.RslintConfig),
		evaluators:          make(map[string]*rslintconfig.ConfigEvaluator),
		sourceByDirectory:   make(map[string]*searchConfigLoadState),
		missingConfigPathID: make(map[string]struct{}),
		failurePathIDs:      make(map[string]struct{}),
	}
	catalog, err := builder.build()
	// A parallel fail-fast return transfers cleanup to the deferred goroutine.
	// Test that handoff first so we never race a surviving member that is still
	// publishing the lazily-created predicate coordinator.
	if err != nil && !builder.cleanupDeferred && builder.predicate != nil {
		builder.predicate.Close()
	}
	return catalog, err
}

func (builder *searchCatalogBuilder) build() (*ConfigCatalog, error) {
	if err := builder.ctx.Err(); err != nil {
		return nil, err
	}
	if builder.request.CWD == "" {
		return nil, errors.New("config discovery requires a working directory")
	}
	builder.request.CWD = rslintconfig.NormalizeHostPath(builder.request.CWD)

	plan, err := BuildSearchPlan(builder.fs, builder.request.CWD, builder.request.Inputs, SearchPlanOptions{
		GlobInputPaths:          builder.request.GlobInputPaths,
		ErrorOnUnmatchedPattern: builder.request.ErrorOnUnmatchedPattern,
		SingleThreaded:          builder.request.SingleThreaded,
	})
	if err != nil {
		return nil, err
	}

	if err := builder.prepareInvocationConfig(); err != nil {
		return nil, err
	}
	if builder.request.ProbeRootConfig && builder.request.Mode == ConfigDiscoveryAuto {
		if _, err := builder.resolveOwner(builder.request.CWD); err != nil {
			return nil, err
		}
	}

	explicitResults := make([]explicitSearchResult, len(plan.ExplicitFiles))
	globResults := make([]globSearchResult, len(plan.GlobSearches))
	lookupErrors := make([]error, len(builder.request.LookupPaths))
	lookupOwner := func(index int) {
		path := normalizeDiscoveryPath(builder.request.LookupPaths[index], builder.request.CWD)
		owner, err := builder.resolveOwner(rslintconfig.HostDirectoryPath(path))
		if err == nil && owner != nil && !owner.missing && !owner.unavailable {
			// Lookup paths are exact documents, not workspace search inputs. LSP
			// evaluates these eagerly so the committed evaluator owns both their
			// immutable .gitignore ancestry and their one ConfigArray selection.
			// Ordinary workspace files remain untouched when CollectTargets=false.
			_, _, err = builder.fileStatus(path, owner)
		}
		lookupErrors[index] = err
	}
	if builder.request.SingleThreaded {
		for index := range builder.request.LookupPaths {
			lookupOwner(index)
			if lookupErrors[index] != nil {
				return nil, lookupErrors[index]
			}
		}
		for index, file := range plan.ExplicitFiles {
			explicitResults[index] = builder.preloadExplicitFile(file)
			if explicitResults[index].err != nil {
				return nil, explicitResults[index].err
			}
		}
		for index, search := range plan.GlobSearches {
			globResults[index] = builder.searchGlob(search)
			if globResults[index].err != nil &&
				(!builder.request.AllowMissingConfig ||
					(!errors.Is(globResults[index].err, ErrNoFilesFound) && !errors.Is(globResults[index].err, ErrAllFilesIgnored))) {
				return nil, globResults[index].err
			}
		}
	} else {
		parallelContext, cancel := context.WithCancel(builder.ctx)
		builder.ctx = parallelContext
		defer cancel()

		type memberSettlement struct {
			err error
		}
		memberCount := len(builder.request.LookupPaths) + len(plan.ExplicitFiles)
		if len(plan.GlobSearches) > 0 {
			memberCount++
		}
		settled := make(chan memberSettlement, memberCount)
		var memberWaitGroup sync.WaitGroup
		memberWaitGroup.Add(memberCount)
		for index := range builder.request.LookupPaths {
			go func() {
				defer memberWaitGroup.Done()
				lookupOwner(index)
				settled <- memberSettlement{err: lookupErrors[index]}
			}()
		}
		for index, file := range plan.ExplicitFiles {
			go func() {
				defer memberWaitGroup.Done()
				explicitResults[index] = builder.preloadExplicitFile(file)
				settled <- memberSettlement{err: explicitResults[index].err}
			}()
		}
		if len(plan.GlobSearches) > 0 {
			go func() {
				defer memberWaitGroup.Done()
				var globWaitGroup sync.WaitGroup
				globWaitGroup.Add(len(plan.GlobSearches))
				for index, search := range plan.GlobSearches {
					go func() {
						defer globWaitGroup.Done()
						globResults[index] = builder.searchGlob(search)
					}()
				}
				globWaitGroup.Wait()
				// globMultiSearch is one outer Promise.all member. Its internal
				// allSettled chooses the first failure in stable search-map order.
				var globErr error
				for _, result := range globResults {
					if result.err != nil {
						globErr = result.err
						break
					}
				}
				settled <- memberSettlement{err: globErr}
			}()
		}

		for range memberCount {
			select {
			case member := <-settled:
				if member.err == nil {
					continue
				}
				// JSON/JSONC fallback is the sole product exception: an unmatched
				// glob can be delegated only if the completed invocation found no
				// JavaScript config. Defer that decision, but fail fast for every
				// module, predicate, and transport failure.
				if builder.request.AllowMissingConfig &&
					(errors.Is(member.err, ErrNoFilesFound) || errors.Is(member.err, ErrAllFilesIgnored)) {
					continue
				}
				cancel()
				builder.deferCleanupUntil(&memberWaitGroup)
				return nil, member.err
			case <-builder.ctx.Done():
				builder.deferCleanupUntil(&memberWaitGroup)
				return nil, builder.ctx.Err()
			}
		}
	}

	fatalMissingConfig := !builder.request.AllowMissingConfig ||
		(builder.configCount() > 0 && !builder.request.RetainUnconfiguredAreas)
	for _, result := range explicitResults {
		if result.err != nil {
			return nil, result.err
		} else if result.missing != "" && fatalMissingConfig {
			return nil, &ConfigFileMissingError{Path: result.missing}
		}
	}
	var globError error
	for _, result := range globResults {
		if len(result.missing) > 0 && fatalMissingConfig &&
			(result.err == nil || errors.Is(result.err, ErrNoFilesFound) || errors.Is(result.err, ErrAllFilesIgnored)) {
			// With JSON/JSONC fallback enabled we initially retain unconfigured
			// paths so the whole invocation can determine whether any JS config
			// participated. Once one did, ESLint's config lookup error belongs to
			// this glob search and takes precedence over its derived unmatched
			// pattern error.
			globError = &ConfigFileMissingError{Path: result.missing[0]}
			break
		}
		if result.err != nil {
			// findFiles queues every explicit-file config load before one
			// globMultiSearch promise; inside globMultiSearch failures are inspected
			// in search-map insertion order (not first-pattern order).
			globError = result.err
			break
		}
	}
	if globError != nil {
		if !builder.request.AllowMissingConfig || builder.configCount() != 0 ||
			(!errors.Is(globError, ErrNoFilesFound) && !errors.Is(globError, ErrAllFilesIgnored)) {
			return nil, globError
		}
	}
	for _, lookupErr := range lookupErrors {
		if lookupErr != nil {
			return nil, lookupErr
		}
	}

	// ESLint's findFiles() has completed now. Only this later selection phase
	// executes explicit-file ConfigArray predicates and produces merged config.
	if builder.request.SingleThreaded {
		for index := range explicitResults {
			builder.evaluateExplicitFile(&explicitResults[index])
			if explicitResults[index].err != nil {
				return nil, explicitResults[index].err
			}
		}
	} else {
		var waitGroup sync.WaitGroup
		waitGroup.Add(len(explicitResults))
		for index := range explicitResults {
			go func() {
				defer waitGroup.Done()
				builder.evaluateExplicitFile(&explicitResults[index])
			}()
		}
		waitGroup.Wait()
		for index := range explicitResults {
			if explicitResults[index].err != nil {
				return nil, explicitResults[index].err
			}
		}
	}

	var explicitInputs []ExplicitInputResult
	var targets []DiscoveredTarget
	for _, result := range explicitResults {
		explicitInputs = append(explicitInputs, result.input)
		if result.missing != "" {
			builder.recordMissingConfig(result.missing)
		}
		if result.target != nil {
			targets = append(targets, *result.target)
		}
	}
	for _, result := range globResults {
		for _, path := range result.missing {
			builder.recordMissingConfig(path)
		}
		targets = append(targets, result.targets...)
	}

	// JSON/JSONC fallback is an intentional product exception. It is allowed
	// only when the entire search reached no JavaScript config; once one JS
	// config participates, an unconfigured file is the same fatal error ESLint
	// reports.
	if len(builder.missingConfigPaths) > 0 &&
		(!builder.request.AllowMissingConfig ||
			(builder.configCount() > 0 && !builder.request.RetainUnconfiguredAreas)) {
		return nil, &ConfigFileMissingError{Path: builder.missingConfigPaths[0]}
	}

	targets = deduplicateDiscoveredTargets(targets)
	if !builder.request.CollectTargets {
		targets = nil
	}
	return builder.catalog(targets, explicitInputs)
}

func (builder *searchCatalogBuilder) deferCleanupUntil(waitGroup *sync.WaitGroup) {
	if builder == nil || waitGroup == nil || builder.cleanupDeferred {
		return
	}
	builder.cleanupDeferred = true
	go func() {
		// A JavaScript module promise cannot be forcibly interrupted. Never make
		// the rejecting outer member wait for a sibling that may remain pending
		// forever; once every cancellable member has unwound, close any predicate
		// worker created by those searches. If a host ignores cancellation, this
		// goroutine intentionally owns the remaining transaction state.
		waitGroup.Wait()
		if builder.predicate != nil {
			builder.predicate.Close()
		}
	}()
}

func (builder *searchCatalogBuilder) prepareInvocationConfig() error {
	switch builder.request.Mode {
	case ConfigDiscoveryAuto:
		return nil
	case ConfigDiscoveryExplicit:
		if builder.request.ExplicitConfigPath == "" {
			return errors.New("explicit config discovery requires a config path")
		}
		candidate := configCandidate{
			path:      normalizeDiscoveryPath(builder.request.ExplicitConfigPath, builder.request.CWD),
			directory: builder.request.CWD,
		}
		// ESLint does not evaluate --config until findFiles reaches a viable
		// filesystem entry. In particular, a nonexistent-base glob reports the
		// unmatched-pattern error even when the explicit config is broken. Keep
		// the fixed owner lexical identity now and let resolveOwner coalesce its
		// first real use across concurrent explicit/glob search members.
		builder.fixedCandidate = &candidate
		return nil
	case ConfigDiscoveryInline:
		owner := builder.newConfiguredOwner(builder.request.CWD, nil, nil)
		builder.fixedOwner = owner
		builder.installOwner(owner)
		return nil
	default:
		return fmt.Errorf("unsupported config discovery mode %d", builder.request.Mode)
	}
}

func (builder *searchCatalogBuilder) preloadExplicitFile(file ExplicitFileSearch) explicitSearchResult {
	result := explicitSearchResult{
		file: file,
		input: ExplicitInputResult{
			Path:   file.Path,
			Order:  file.Order,
			Status: ExplicitInputUnconfigured,
		},
	}
	owner, err := builder.resolveOwner(rslintconfig.HostDirectoryPath(file.Path))
	if err != nil {
		result.err = err
		return result
	}
	if owner.missing {
		result.missing = file.Path
		if !builder.request.AllowMissingConfig {
			result.err = &ConfigFileMissingError{Path: file.Path}
		}
		return result
	}
	result.owner = owner
	result.input.ConfigDirectory = owner.directory
	return result
}

func (builder *searchCatalogBuilder) evaluateExplicitFile(result *explicitSearchResult) {
	if result == nil || result.err != nil || result.missing != "" || result.owner == nil {
		return
	}
	status, merged, err := builder.fileStatus(result.file.Path, result.owner)
	if err != nil {
		result.err = err
		return
	}
	result.input.Status = status
	if status == ExplicitInputConfigured && builder.request.CollectTargets {
		result.target = &DiscoveredTarget{
			Path:            result.file.Path,
			ConfigDirectory: result.owner.directory,
			Explicit:        true,
			MergedConfig:    merged,
		}
	}
}

func (builder *searchCatalogBuilder) searchGlob(search GlobSearch) globSearchResult {
	compiled := make([]SearchPattern, len(search.Patterns))
	for index, pattern := range search.Patterns {
		matcher, err := CompileSearchPattern(pattern, search.BasePath)
		if err != nil {
			return globSearchResult{err: err}
		}
		compiled[index] = matcher
	}
	matched := make([]bool, len(compiled))
	result := globSearchResult{}
	if builder.fs.DirectoryExists(search.BasePath) {
		// The final component of a direct directory input is followed exactly
		// once. Keep the lexical and scan paths paired through recursion so
		// overlay/custom filesystems need not re-resolve the alias at every level.
		walkBase := search.BasePath
		if realPath := builder.fs.Realpath(search.BasePath); realPath != "" {
			walkBase = realPath
		}
		result.err = builder.walkGlobSearch(search, compiled, matched, search.BasePath, walkBase, "", &result)
		if result.err != nil {
			return result
		}
		if builder.request.ErrorOnUnmatchedPattern {
			for index, wasMatched := range matched {
				if wasMatched {
					continue
				}
				hasFilesystemMatch, err := builder.searchPatternHasFilesystemMatch(search.BasePath, walkBase, compiled[index])
				if err != nil {
					result.err = err
					break
				}
				if hasFilesystemMatch {
					result.err = &AllFilesIgnoredError{Pattern: search.RawPatterns[index]}
				} else {
					result.err = &NoFilesFoundError{Pattern: search.RawPatterns[index]}
				}
				break
			}
		}
	} else if builder.request.ErrorOnUnmatchedPattern && len(compiled) > 0 {
		result.err = &NoFilesFoundError{Pattern: search.RawPatterns[0]}
	}
	return result
}

func (builder *searchCatalogBuilder) walkGlobSearch(
	search GlobSearch,
	patterns []SearchPattern,
	matched []bool,
	directory string,
	walkDirectory string,
	relativeDirectory string,
	result *globSearchResult,
) error {
	if err := builder.ctx.Err(); err != nil {
		return err
	}
	builder.addDirectoriesVisited(1)
	items, err := builder.readDirectory(directory, walkDirectory)
	if err != nil {
		return err
	}

	var owner *searchOwner
	ownerLoaded := false
	loadOwner := func() (*searchOwner, error) {
		if ownerLoaded {
			return owner, nil
		}
		ownerLoaded = true
		var err error
		owner, err = builder.resolveOwner(directory)
		return owner, err
	}
	if !builder.request.CollectTargets {
		// Catalog-only walks still need to probe the directory itself. Deferring
		// owner lookup until a child is encountered would miss a config in a leaf
		// package whose only entries are files (the common LSP workspace case).
		if _, err := loadOwner(); err != nil {
			return err
		}
	}

	for _, item := range items {
		child := joinDiscoveryPath(directory, item.name)
		relative := item.name
		if relativeDirectory != "" {
			relative = relativeDirectory + "/" + item.name
		}
		if item.directory {
			viable := false
			for _, pattern := range patterns {
				if pattern.PartialMatch(relative) {
					viable = true
					break
				}
			}
			if !viable {
				continue
			}
			parentOwner, err := loadOwner()
			if err != nil {
				return err
			}
			if parentOwner.unavailable {
				builder.addDirectoriesPruned(1)
				continue
			}
			ignored, err := builder.directoryIgnored(child, parentOwner)
			if err != nil {
				return err
			}
			if ignored {
				builder.addDirectoriesPruned(1)
				continue
			}
			walkChild := joinDiscoveryPath(walkDirectory, item.name)
			if err := builder.walkGlobSearch(search, patterns, matched, child, walkChild, relative, result); err != nil {
				return err
			}
			continue
		}

		// Catalog-only LSP refreshes need config ownership boundaries, not a
		// workspace-wide execution of files/local-ignore predicates. Directory
		// global predicates above still participate in traversal pruning.
		if !builder.request.CollectTargets && !builder.request.ErrorOnUnmatchedPattern {
			continue
		}
		fileOwner, err := loadOwner()
		if err != nil {
			return err
		}
		if fileOwner.unavailable {
			continue
		}
		if fileOwner.missing {
			result.missing = append(result.missing, child)
			if !builder.request.AllowMissingConfig {
				return &ConfigFileMissingError{Path: child}
			}
			continue
		}
		status, merged, err := builder.fileStatus(child, fileOwner)
		if err != nil {
			return err
		}
		configured := status == ExplicitInputConfigured
		matchesAny := false
		for index, pattern := range patterns {
			if pattern.Match(relative) {
				matchesAny = true
				if configured {
					// ESLint tracks unmatched patterns by their original relative
					// spelling but deletes Minimatch.pattern after a match. Minimatch
					// strips leading ! markers, so a leading-negated search remains
					// unmatched even when it selected other files. Preserve that v10
					// behavior instead of treating matched[index] as a simple boolean.
					removalKey := patterns[index].MatchRemovalKey()
					for unmatchedIndex := range patterns {
						if patterns[unmatchedIndex].UnmatchedKey() == removalKey {
							matched[unmatchedIndex] = true
						}
					}
				}
			}
		}
		if configured && matchesAny && builder.request.CollectTargets {
			result.targets = append(result.targets, DiscoveredTarget{
				Path:            child,
				ConfigDirectory: fileOwner.directory,
				MergedConfig:    merged,
			})
		}
	}
	return nil
}

func (builder *searchCatalogBuilder) resolveOwner(directory string) (*searchOwner, error) {
	if builder.fixedOwner != nil {
		return builder.fixedOwner, nil
	}
	if builder.fixedCandidate != nil {
		// Every target governed by --config shares cwd as its ConfigArray
		// basePath, regardless of the target directory that triggered the load.
		directory = builder.request.CWD
	}
	directory = rslintconfig.NormalizeHostPath(directory)
	identity := discoveryPathIdentity(directory, builder.fs.UseCaseSensitiveFileNames())

	builder.mu.Lock()
	if existing := builder.owners[identity]; existing != nil {
		builder.mu.Unlock()
		select {
		case <-existing.ready:
			return existing.owner, existing.err
		case <-builder.ctx.Done():
			return nil, builder.ctx.Err()
		}
	}
	state := &searchOwnerState{ready: make(chan struct{})}
	builder.owners[identity] = state
	builder.mu.Unlock()

	var owner *searchOwner
	var err error
	if builder.fixedCandidate != nil {
		var load *searchConfigLoadState
		load, err = builder.ensureConfigCandidate(*builder.fixedCandidate)
		if err == nil {
			owner = builder.newConfiguredOwner(builder.request.CWD, load.entries, load)
			builder.installOwner(owner)
		}
	} else {
		owner, err = builder.calculateOwner(directory)
	}
	state.owner = owner
	state.err = err
	close(state.ready)
	return owner, err
}

func (builder *searchCatalogBuilder) calculateOwner(directory string) (*searchOwner, error) {
	if candidate, found := builder.findCandidateInDirectory(directory); found {
		state, err := builder.ensureConfigCandidate(candidate)
		if err != nil {
			if builder.request.CollectConfigFailures && state != nil {
				builder.recordConfigFailure(state)
				return builder.newUnconfiguredOwner(candidate.directory, state, false, true), nil
			}
			return nil, err
		}
		owner := builder.newConfiguredOwner(candidate.directory, state.entries, state)
		builder.installOwner(owner)
		return owner, nil
	}

	parent := rslintconfig.HostDirectoryPath(directory)
	if parent != "" && parent != directory {
		// Share the parent's in-flight owner state. This makes a config and its
		// compiled ignore matcher one immutable ownership object for the whole
		// inherited subtree instead of rebuilding it for every visited directory.
		return builder.resolveOwner(parent)
	}

	{
		return builder.newUnconfiguredOwner(builder.request.CWD, nil, true, false), nil
	}
}

func (builder *searchCatalogBuilder) findCandidateInDirectory(directory string) (configCandidate, bool) {
	directory = rslintconfig.NormalizeHostPath(directory)
	for _, name := range autoJSConfigFileNames {
		path := joinDiscoveryPath(directory, name)
		if builder.fs.FileExists(path) {
			return configCandidate{path: rslintconfig.NormalizeHostPath(path), directory: directory}, true
		}
	}
	return configCandidate{}, false
}

func (builder *searchCatalogBuilder) ensureConfigCandidate(candidate configCandidate) (*searchConfigLoadState, error) {
	candidate.path = rslintconfig.NormalizeHostPath(candidate.path)
	candidate.directory = rslintconfig.NormalizeHostPath(candidate.directory)
	identity := discoveryPathIdentity(candidate.path, builder.fs.UseCaseSensitiveFileNames())

	builder.mu.Lock()
	if existing := builder.loadStates[identity]; existing != nil {
		builder.mu.Unlock()
		select {
		case <-existing.ready:
			return existing, builder.configLoadStateError(existing)
		case <-builder.ctx.Done():
			return nil, builder.ctx.Err()
		}
	}
	builder.nextRequestID++
	state := &searchConfigLoadState{
		ready:     make(chan struct{}),
		id:        fmt.Sprintf("config-%06d", builder.nextRequestID),
		candidate: candidate,
	}
	builder.loadStates[identity] = state
	builder.mu.Unlock()

	if builder.loader == nil {
		state.err = errors.New("javascript config candidates require a module loader")
		close(state.ready)
		return nil, state.err
	}
	builder.performConfigLoad(state)
	return state, builder.configLoadStateError(state)
}

func (builder *searchCatalogBuilder) performConfigLoad(state *searchConfigLoadState) {
	loadMode := ConfigModuleLoadCached
	if builder.request.Fresh {
		loadMode = ConfigModuleLoadFresh
	}
	request := ConfigLoadBatchRequest{
		TransactionID:  builder.transactionID,
		LoadMode:       loadMode,
		SingleThreaded: builder.request.SingleThreaded,
		Candidates: []ConfigLoadCandidate{{
			ID: state.id,
			// Preserve the lexical config path as a routing identity. Node owns
			// physical module identity and shares the raw module source by realpath,
			// while normalizing the export once per lexical candidate.
			ConfigPath:      state.candidate.path,
			ConfigDirectory: state.candidate.directory,
		}},
	}
	builder.mu.Lock()
	builder.stats.ConfigsRequested++
	builder.mu.Unlock()

	response, err := builder.loader.LoadConfigs(builder.ctx, request)
	var results map[string]ConfigLoadResult
	if err == nil {
		results, err = validateConfigLoadBatch(request, response)
	}
	if err != nil {
		state.err = err
		close(state.ready)
		return
	}
	result := results[state.id]
	if result.Status == "failed" {
		failure := *result.Error
		state.failure = &failure
	} else if validateErr := rslintconfig.ValidateConfig(result.Entries); validateErr != nil {
		state.failure = &ConfigModuleError{Code: "invalid", Message: validateErr.Error()}
	} else {
		state.entries = append(rslintconfig.RslintConfig(nil), result.Entries...)
		builder.mu.Lock()
		builder.stats.ConfigsLoaded++
		builder.mu.Unlock()
	}
	close(state.ready)
}

func (builder *searchCatalogBuilder) configLoadStateError(state *searchConfigLoadState) error {
	if state == nil {
		return errors.New("config load completed without state")
	}
	if state.err != nil {
		return state.err
	}
	if state.failure == nil {
		return nil
	}
	failure := ConfigFailure{
		Path:      state.candidate.path,
		Directory: state.candidate.directory,
		Kind:      state.failure.Code,
		Message:   state.failure.Message,
	}
	return &AllConfigsFailedError{
		Failures: []ConfigFailure{failure},
		message:  fmt.Sprintf("failed to load nearest JavaScript config %q: %s", failure.Path, failure.Message),
	}
}

func (builder *searchCatalogBuilder) newConfiguredOwner(
	directory string,
	moduleConfig rslintconfig.RslintConfig,
	source *searchConfigLoadState,
) *searchOwner {
	directory = rslintconfig.NormalizeHostPath(directory)
	baseline := defaultSearchConfig()
	effective := make(rslintconfig.RslintConfig, 0, len(baseline)+len(moduleConfig)+len(builder.request.OverrideConfig))
	effective = append(effective, baseline...)
	effective = append(effective, moduleConfig...)
	effective = append(effective, builder.request.OverrideConfig...)
	owner := &searchOwner{
		directory: directory,
		config:    effective,
		source:    source,
	}
	predicateResolver := builder.predicateResolverFor(effective)
	owner.evaluator = rslintconfig.NewConfigEvaluatorWithGitignore(effective, directory, builder.fs, predicateResolver)
	return owner
}

// Missing and unavailable owners still need ESLint's ordered default and
// override global-ignore gate while searching for nested config boundaries.
// They deliberately omit the rslint .gitignore product layer: .gitignore is
// rooted at an already-selected governing config and cannot select that owner.
func (builder *searchCatalogBuilder) newUnconfiguredOwner(
	directory string,
	source *searchConfigLoadState,
	missing bool,
	unavailable bool,
) *searchOwner {
	directory = rslintconfig.NormalizeHostPath(directory)
	effective := append(defaultSearchIgnores(), builder.request.OverrideConfig...)
	return &searchOwner{
		directory:   directory,
		config:      effective,
		evaluator:   rslintconfig.NewConfigEvaluator(effective, directory, builder.fs, builder.predicateResolverFor(effective)),
		missing:     missing,
		unavailable: unavailable,
		source:      source,
	}
}

func (builder *searchCatalogBuilder) predicateResolverFor(
	config rslintconfig.RslintConfig,
) rslintconfig.ConfigPredicateResolver {
	if !rslintconfig.HasConfigPredicates(config) {
		return nil
	}
	builder.predicateOnce.Do(func() {
		builder.predicate = newPredicateCoordinator(builder.ctx, builder.loader, builder.transactionID)
	})
	return builder.predicate
}

func (builder *searchCatalogBuilder) installOwner(owner *searchOwner) {
	if owner == nil || owner.missing || owner.unavailable {
		return
	}
	builder.mu.Lock()
	defer builder.mu.Unlock()
	if existing, ok := builder.configs[owner.directory]; ok {
		if source := builder.sourceByDirectory[owner.directory]; source != owner.source {
			// Lexically identical config directories cannot select two different
			// nearest modules in one immutable build.
			_ = existing
			return
		}
		return
	}
	builder.configs[owner.directory] = append(rslintconfig.RslintConfig(nil), owner.config...)
	builder.evaluators[owner.directory] = owner.evaluator
	builder.sourceByDirectory[owner.directory] = owner.source
}

func (builder *searchCatalogBuilder) recordConfigFailure(state *searchConfigLoadState) {
	if state == nil {
		return
	}
	failure := ConfigFailure{Path: state.candidate.path, Directory: state.candidate.directory}
	if state.failure != nil {
		failure.Kind = state.failure.Code
		failure.Message = state.failure.Message
	} else if state.err != nil {
		failure.Kind = "load"
		failure.Message = state.err.Error()
	}
	builder.mu.Lock()
	defer builder.mu.Unlock()
	if _, exists := builder.failurePathIDs[failure.Path]; exists {
		return
	}
	builder.failurePathIDs[failure.Path] = struct{}{}
	builder.failures = append(builder.failures, failure)
}

func (builder *searchCatalogBuilder) fileStatus(filePath string, owner *searchOwner) (ExplicitInputStatus, *rslintconfig.MergedConfig, error) {
	if owner == nil || owner.missing || owner.unavailable {
		return ExplicitInputUnconfigured, nil, nil
	}
	if builder.request.Mode != ConfigDiscoveryAuto {
		caseSensitive := builder.fs == nil || builder.fs.UseCaseSensitiveFileNames()
		if _, within := rslintconfig.RelativePathWithinConfigRoot(filePath, owner.directory, caseSensitive); !within {
			return ExplicitInputExternal, nil, nil
		}
	}
	resolution, err := owner.evaluator.GetConfigForFile(builder.ctx, filePath)
	if err != nil {
		return ExplicitInputUnconfigured, nil, err
	}
	switch resolution.Status {
	case rslintconfig.ConfigFileConfigured:
		return ExplicitInputConfigured, resolution.Config, nil
	case rslintconfig.ConfigFileIgnored:
		return ExplicitInputIgnored, nil, nil
	case rslintconfig.ConfigFileExternal:
		return ExplicitInputExternal, nil, nil
	default:
		return ExplicitInputUnconfigured, nil, nil
	}
}

func (builder *searchCatalogBuilder) directoryIgnored(directory string, owner *searchOwner) (bool, error) {
	if owner == nil {
		return false, nil
	}
	return owner.evaluator.IsDirectoryIgnored(builder.ctx, directory)
}

func defaultSearchConfig() rslintconfig.RslintConfig {
	patterns := make([]string, 0, len(rslintconfig.DefaultLintFileExtensions))
	for _, extension := range rslintconfig.DefaultLintFileExtensions {
		patterns = append(patterns, "**/*"+extension)
	}
	return rslintconfig.WithDefaultGlobalIgnores(rslintconfig.RslintConfig{{
		Name:  "rslint/default-file-patterns",
		Files: patterns,
	}})
}

func defaultSearchIgnores() rslintconfig.RslintConfig {
	return rslintconfig.WithDefaultGlobalIgnores(nil)
}

func (builder *searchCatalogBuilder) searchPatternHasFilesystemMatch(basePath string, walkBase string, pattern SearchPattern) (bool, error) {
	if !builder.fs.DirectoryExists(basePath) {
		return false, nil
	}
	var walk func(string, string, string) (bool, error)
	walk = func(directory string, walkDirectory string, relativeDirectory string) (bool, error) {
		if err := builder.ctx.Err(); err != nil {
			return false, err
		}
		entries, err := builder.readDirectory(directory, walkDirectory)
		if err != nil {
			return false, err
		}
		for _, entry := range entries {
			relative := entry.name
			if relativeDirectory != "" {
				relative = relativeDirectory + "/" + entry.name
			}
			if !entry.directory && pattern.Match(relative) {
				return true, nil
			}
			if !entry.directory {
				continue
			}
			if !pattern.PartialMatch(relative) {
				continue
			}
			matched, err := walk(
				joinDiscoveryPath(directory, entry.name),
				joinDiscoveryPath(walkDirectory, entry.name),
				relative,
			)
			if err != nil || matched {
				return matched, err
			}
		}
		return false, nil
	}
	return walk(basePath, walkBase, "")
}

// readDirectory projects the one-level lstat-style directory-entry view used by
// ESLint's humanfs walker. Symlinks are leaf entries regardless of their
// target, broken symlinks remain visible, ENOENT is an empty walk, and other
// readdir errors stay fatal. WalkDir keeps the behavior available to overlay
// and test filesystems without a second OS-only discovery model.
func (builder *searchCatalogBuilder) readDirectory(directory string, walkDirectory string) ([]searchDirectoryEntry, error) {
	entriesByName := make(map[string]searchDirectoryEntry)
	// A direct directory input is stat'ed with symlink-following semantics by
	// findFiles. Enumerate its physical root so a root alias is entered, while
	// still returning lexical directory/name paths and treating nested symlink
	// directory entries as leaves.
	root := rslintconfig.NormalizeHostPath(walkDirectory)
	err := builder.fs.WalkDir(walkDirectory, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			if errors.Is(walkErr, fs.ErrNotExist) {
				return nil
			}
			return walkErr
		}
		if rslintconfig.NormalizeHostPath(path) == root {
			return nil
		}
		if entry == nil {
			return nil
		}
		name := entry.Name()
		if entry.Type()&fs.ModeSymlink != 0 {
			entriesByName[name] = searchDirectoryEntry{name: name}
			return nil
		}
		if entry.IsDir() {
			entriesByName[name] = searchDirectoryEntry{name: name, directory: true}
			return fs.SkipDir
		}
		entriesByName[name] = searchDirectoryEntry{name: name}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("read directory %q: %w", directory, err)
	}

	entries := make([]searchDirectoryEntry, 0, len(entriesByName))
	for _, entry := range entriesByName {
		entries = append(entries, entry)
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].name != entries[j].name {
			return entries[i].name < entries[j].name
		}
		return entries[i].directory && !entries[j].directory
	})
	return entries, nil
}

func joinDiscoveryPath(directory string, name string) string {
	if runtime.GOOS == "windows" {
		return tspath.CombinePaths(directory, name)
	}
	return filepath.Join(directory, name)
}

func discoveryPathIdentity(path string, caseSensitive bool) tspath.Path {
	path = rslintconfig.NormalizeHostPath(path)
	if !caseSensitive {
		path = strings.ToLower(path)
	}
	return tspath.Path(path)
}

func deduplicateDiscoveredTargets(targets []DiscoveredTarget) []DiscoveredTarget {
	seen := make(map[string]int, len(targets))
	result := make([]DiscoveredTarget, 0, len(targets))
	for _, target := range targets {
		target.Path = rslintconfig.NormalizeHostPath(target.Path)
		if index, exists := seen[target.Path]; exists {
			result[index].Explicit = result[index].Explicit || target.Explicit
			continue
		}
		seen[target.Path] = len(result)
		result = append(result, target)
	}
	return result
}

func (builder *searchCatalogBuilder) recordMissingConfig(path string) {
	path = rslintconfig.NormalizeHostPath(path)
	if _, exists := builder.missingConfigPathID[path]; exists {
		return
	}
	builder.missingConfigPathID[path] = struct{}{}
	builder.missingConfigPaths = append(builder.missingConfigPaths, path)
}

func (builder *searchCatalogBuilder) configCount() int {
	builder.mu.Lock()
	defer builder.mu.Unlock()
	return len(builder.configs)
}

func (builder *searchCatalogBuilder) addDirectoriesVisited(count int) {
	builder.mu.Lock()
	builder.stats.DirectoriesVisited += count
	builder.mu.Unlock()
}

func (builder *searchCatalogBuilder) addDirectoriesPruned(count int) {
	builder.mu.Lock()
	builder.stats.DirectoriesPruned += count
	builder.mu.Unlock()
}

func (builder *searchCatalogBuilder) catalog(
	targets []DiscoveredTarget,
	explicitInputs []ExplicitInputResult,
) (*ConfigCatalog, error) {
	builder.mu.Lock()
	configs := make(map[string]rslintconfig.RslintConfig, len(builder.configs))
	for directory, config := range builder.configs {
		configs[directory] = append(rslintconfig.RslintConfig(nil), config...)
	}
	evaluators := make(map[string]*rslintconfig.ConfigEvaluator, len(builder.evaluators))
	for directory, evaluator := range builder.evaluators {
		evaluators[directory] = evaluator
	}
	sources := make([]*searchConfigLoadState, 0, len(builder.sourceByDirectory))
	seenSource := make(map[string]struct{}, len(builder.sourceByDirectory))
	for _, source := range builder.sourceByDirectory {
		if source == nil || source.id == "" {
			continue
		}
		if _, exists := seenSource[source.id]; exists {
			continue
		}
		seenSource[source.id] = struct{}{}
		sources = append(sources, source)
	}
	stats := builder.stats
	failures := append([]ConfigFailure(nil), builder.failures...)
	builder.mu.Unlock()

	// Candidate IDs are allocated by concurrent search arrival and are therefore
	// intentionally opaque. Activation order is instead derived from the stable
	// lexical config path so plugin aggregation cannot depend on goroutine
	// scheduling.
	sort.Slice(sources, func(i, j int) bool {
		if sources[i].candidate.path != sources[j].candidate.path {
			return sources[i].candidate.path < sources[j].candidate.path
		}
		return sources[i].id < sources[j].id
	})
	sort.Slice(failures, func(i, j int) bool {
		if failures[i].Directory != failures[j].Directory {
			return failures[i].Directory < failures[j].Directory
		}
		return failures[i].Path < failures[j].Path
	})
	effectiveIDs := make([]string, 0, len(sources))
	for _, source := range sources {
		effectiveIDs = append(effectiveIDs, source.id)
	}
	var eslintPlugins []rslintconfig.EslintPluginEntry
	if len(effectiveIDs) > 0 {
		if builder.loader == nil {
			return nil, errors.New("javascript config activation requires a module loader")
		}
		response, err := builder.loader.ActivateConfigs(builder.ctx, ConfigActivationRequest{
			TransactionID:      builder.transactionID,
			EffectiveConfigIDs: effectiveIDs,
		})
		if err != nil {
			return nil, err
		}
		if response.TransactionID != builder.transactionID {
			return nil, configDiscoveryProtocolError(
				"activation transaction mismatch: got %q, want %q",
				response.TransactionID,
				builder.transactionID,
			)
		}
		eslintPlugins = cloneEslintPluginEntries(response.EslintPluginEntries)
	}

	return &ConfigCatalog{
		TransactionID:      builder.transactionID,
		Configs:            configs,
		EffectiveConfigIDs: effectiveIDs,
		EslintPlugins:      eslintPlugins,
		Targets:            targets,
		ExplicitInputs:     explicitInputs,
		Stats:              stats,
		Failures:           failures,
		Explicit: builder.request.Mode == ConfigDiscoveryExplicit ||
			builder.request.Mode == ConfigDiscoveryInline,
		predicateCoordinator: builder.predicate,
		configEvaluators:     evaluators,
	}, nil
}
