package config

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/microsoft/typescript-go/shim/vfs"
	"github.com/web-infra-dev/rslint/internal/config/gitignore"
)

// ConfigPredicateResolver is the narrow boundary between ConfigArray semantics
// and the JavaScript host that owns live matcher closures. Implementations may
// batch calls from independent paths, but this evaluator only asks for the next
// predicate after all preceding short-circuit conditions have been resolved.
type ConfigPredicateResolver interface {
	ResolveConfigPredicate(ctx context.Context, predicateID string, filePath string, directory bool) (bool, error)
}

type ConfigFileStatus string

const (
	ConfigFileConfigured   ConfigFileStatus = "configured"
	ConfigFileIgnored      ConfigFileStatus = "ignored"
	ConfigFileUnconfigured ConfigFileStatus = "unconfigured"
	ConfigFileExternal     ConfigFileStatus = "external"
)

// ConfigFileResolution is one cached ConfigArray getConfigWithStatus result.
// Config is immutable and non-nil only for ConfigFileConfigured.
type ConfigFileResolution struct {
	Status ConfigFileStatus
	Config *MergedConfig
}

type compiledConfigMatcher struct {
	stringMatcher compiledFileSelector
	predicateID   string
}

func compileConfigMatcher(matcher configMatcher) compiledConfigMatcher {
	if matcher.isPredicate() {
		return compiledConfigMatcher{predicateID: matcher.predicateID}
	}
	return compiledConfigMatcher{stringMatcher: compileFileSelector(matcher.pattern)}
}

func (matcher compiledConfigMatcher) matches(
	ctx context.Context,
	paths []string,
	absolutePath string,
	directory bool,
	resolver ConfigPredicateResolver,
) (bool, error) {
	if matcher.predicateID == "" {
		return matcher.stringMatcher.matches(paths), nil
	}
	if resolver == nil {
		return false, fmt.Errorf("config predicate %q requires a JavaScript matcher host", matcher.predicateID)
	}
	return resolver.ResolveConfigPredicate(ctx, matcher.predicateID, absolutePath, directory)
}

type compiledConfigFileSelector struct {
	matchers  []compiledConfigMatcher
	universal bool
}

type compiledConfigEvaluatorEntry struct {
	basePath      string
	caseSensitive bool
	hasFiles      bool
	nonUniversal  []compiledConfigFileSelector
	universal     []compiledConfigFileSelector
	ignores       []compiledIgnoreMatcher
}

type configEvaluationState struct {
	ready      chan struct{}
	resolution ConfigFileResolution
	ignored    bool
	err        error
}

type mergedConfigState struct {
	ready  chan struct{}
	config *MergedConfig
}

// ConfigEvaluator is a concurrency-safe, exact ConfigArray matcher for one
// immutable effective config. It owns complete file/directory decision caches;
// individual predicate results are intentionally never memoized because the
// same closure may be reached twice by ESLint's universal fallback.
type ConfigEvaluator struct {
	config        RslintConfig
	basePath      string
	caseSensitive bool
	fs            vfs.FS
	resolver      ConfigPredicateResolver
	gitignore     *configGitignoreProvider
	globalIgnore  []configIgnoreLayer
	entries       []compiledConfigEvaluatorEntry

	mu             sync.Mutex
	fileCache      map[string]*configEvaluationState
	directoryCache map[string]*configEvaluationState
	mergedCache    map[string]*mergedConfigState
}

// configGitignoreProvider materializes only the .gitignore sources on the
// ancestry of an exact ConfigArray query. This keeps discovery bounded to its
// actual search paths and lets the .gitignore layer participate in the same
// ordered ignore reduction as defaults and authored config without a second
// matcher pass.
type configGitignoreProvider struct {
	basePath        string
	fs              vfs.FS
	caseInsensitive bool
	ancestry        *gitignore.AncestryCollector

	mu       sync.Mutex
	bySource map[string]*configGitignoreState
}

type configGitignoreState struct {
	ready  chan struct{}
	layer  configIgnoreLayer
	layers []configIgnoreLayer
}

func NewConfigEvaluator(
	config RslintConfig,
	basePath string,
	fsys vfs.FS,
	resolver ConfigPredicateResolver,
) *ConfigEvaluator {
	return newConfigEvaluator(config, basePath, fsys, resolver, false)
}

// NewConfigEvaluatorWithGitignore creates the discovery evaluator used by
// JavaScript/TypeScript flat config. Product-level .gitignore patterns are
// evaluated before the default and authored global ignore entries, while later
// authored negations can still reopen them. Sources are read lazily along exact
// target ancestry rather than by sweeping the config-owned subtree.
func NewConfigEvaluatorWithGitignore(
	config RslintConfig,
	basePath string,
	fsys vfs.FS,
	resolver ConfigPredicateResolver,
) *ConfigEvaluator {
	return newConfigEvaluator(config, basePath, fsys, resolver, true)
}

func newConfigEvaluator(
	config RslintConfig,
	basePath string,
	fsys vfs.FS,
	resolver ConfigPredicateResolver,
	includeGitignore bool,
) *ConfigEvaluator {
	basePath = normalizePathForRoot(basePath, basePath)
	evaluator := &ConfigEvaluator{
		config:         config,
		basePath:       basePath,
		caseSensitive:  selectorScopeCaseSensitive(basePath),
		fs:             fsys,
		resolver:       resolver,
		globalIgnore:   compileConfigIgnoreLayers(config, basePath, fsys),
		entries:        make([]compiledConfigEvaluatorEntry, len(config)),
		fileCache:      make(map[string]*configEvaluationState),
		directoryCache: make(map[string]*configEvaluationState),
		mergedCache:    make(map[string]*mergedConfigState),
	}
	if includeGitignore && fsys != nil {
		evaluator.gitignore = &configGitignoreProvider{
			basePath:        basePath,
			fs:              fsys,
			caseInsensitive: !fsys.UseCaseSensitiveFileNames(),
			ancestry:        gitignore.NewAncestryCollector(basePath, fsys),
			bySource:        make(map[string]*configGitignoreState),
		}
	}
	for entryIndex, entry := range config {
		entryBasePath := resolveConfigEntryBasePath(basePath, entry.BasePath)
		compiled := compiledConfigEvaluatorEntry{
			basePath:      entryBasePath,
			caseSensitive: selectorScopeCaseSensitive(entryBasePath),
			hasFiles:      hasFileSelectors(entry),
		}
		for _, selector := range fileSelectors(entry) {
			compiledSelector := compiledConfigFileSelector{
				matchers:  make([]compiledConfigMatcher, 0, len(selector.matchers)),
				universal: true,
			}
			for _, matcher := range selector.matchers {
				compiledSelector.matchers = append(compiledSelector.matchers, compileConfigMatcher(matcher))
				if matcher.isPredicate() || !isUniversalConfigPattern(matcher.pattern) {
					compiledSelector.universal = false
				}
			}
			if compiledSelector.universal {
				compiled.universal = append(compiled.universal, compiledSelector)
			} else {
				compiled.nonUniversal = append(compiled.nonUniversal, compiledSelector)
			}
		}
		for _, matcher := range ignoreMatchers(entry) {
			if matcher.isPredicate() {
				compiled.ignores = append(compiled.ignores, compiledIgnoreMatcher{predicateID: matcher.predicateID})
				continue
			}
			pattern := ParseIgnorePattern(matcher.pattern)
			compiled.ignores = append(compiled.ignores, compiledIgnoreMatcher{pattern: &pattern})
		}
		evaluator.entries[entryIndex] = compiled
	}
	return evaluator
}

func (provider *configGitignoreProvider) layersForTarget(targetPath string, directory bool) []configIgnoreLayer {
	if provider == nil || provider.fs == nil {
		return nil
	}
	targetPath = normalizePathForRoot(provider.basePath, targetPath)
	collectionTarget := targetPath
	if directory {
		// Exact-file collection reads every .gitignore from the config root
		// through the file's parent. A synthetic child gives a directory query
		// that same ancestry without requiring the probe to exist.
		collectionTarget = resolvePathForRoot(provider.basePath, targetPath, ".rslint-gitignore-probe")
	}
	collectionTarget = gitignoreCollectionFilePath(collectionTarget, provider.basePath, provider.fs)
	sourceDirectory := directoryPathForRoot(provider.basePath, collectionTarget)
	sourceDirectory = normalizePathForRoot(provider.basePath, sourceDirectory)
	state := provider.stateForDirectory(sourceDirectory)
	if state == nil {
		return nil
	}
	return state.layers
}

func (provider *configGitignoreProvider) stateForDirectory(sourceDirectory string) *configGitignoreState {
	if provider == nil || provider.ancestry == nil || sourceDirectory == "" {
		return nil
	}
	sourceID := sourceDirectory
	if provider.caseInsensitive {
		sourceID = strings.ToLower(sourceID)
	}
	provider.mu.Lock()
	if cached := provider.bySource[sourceID]; cached != nil {
		provider.mu.Unlock()
		<-cached.ready
		return cached
	}
	state := &configGitignoreState{ready: make(chan struct{})}
	provider.bySource[sourceID] = state
	provider.mu.Unlock()

	source := provider.ancestry.Source(sourceDirectory)
	if source.Parent != "" {
		parent := provider.stateForDirectory(source.Parent)
		if parent != nil {
			state.layers = parent.layers
		}
	}
	patterns := parseCollectedGitignorePatterns(source.Globs, provider.caseInsensitive)
	matchers := make([]compiledIgnoreMatcher, len(patterns))
	for index := range patterns {
		matchers[index] = compiledIgnoreMatcher{pattern: &patterns[index]}
	}
	state.layer = configIgnoreLayer{
		basePath: provider.basePath,
		patterns: patterns,
		matchers: matchers,
	}
	if len(state.layer.matchers) > 0 {
		state.layers = append(append([]configIgnoreLayer(nil), state.layers...), state.layer)
	}
	close(state.ready)
	return state
}

// GetConfigForFile mirrors ConfigArray.getConfigWithStatus, including exact
// path caching, global directory/file ignores, universal selector gating, and
// local-ignore fallback behavior.
func (evaluator *ConfigEvaluator) GetConfigForFile(ctx context.Context, filePath string) (ConfigFileResolution, error) {
	if evaluator == nil {
		return ConfigFileResolution{Status: ConfigFileUnconfigured}, nil
	}
	filePath = normalizePathForRoot(evaluator.basePath, filePath)
	return evaluator.cachedFileResolution(ctx, filePath)
}

func (evaluator *ConfigEvaluator) cachedFileResolution(ctx context.Context, filePath string) (ConfigFileResolution, error) {
	var state *configEvaluationState
	for {
		evaluator.mu.Lock()
		if existing := evaluator.fileCache[filePath]; existing != nil {
			evaluator.mu.Unlock()
			select {
			case <-existing.ready:
				if existing.err != nil {
					// The leader removed this failed state before closing ready.
					// ConfigArray retries every later synchronous query that follows
					// a throwing matcher instead of sharing that transient failure.
					continue
				}
				return existing.resolution, nil
			case <-ctx.Done():
				return ConfigFileResolution{}, ctx.Err()
			}
		}
		state = &configEvaluationState{ready: make(chan struct{})}
		evaluator.fileCache[filePath] = state
		evaluator.mu.Unlock()
		break
	}

	state.resolution, state.err = evaluator.calculateFileResolution(ctx, filePath)
	evaluator.mu.Lock()
	if state.err != nil {
		// ConfigArray does not cache a query that threw.
		delete(evaluator.fileCache, filePath)
	}
	close(state.ready)
	evaluator.mu.Unlock()
	return state.resolution, state.err
}

func (evaluator *ConfigEvaluator) calculateFileResolution(ctx context.Context, filePath string) (ConfigFileResolution, error) {
	if _, within := selectorMatchPaths(filePath, evaluator.basePath, false, evaluator.caseSensitive); !within {
		return ConfigFileResolution{Status: ConfigFileExternal}, nil
	}
	ignored, err := evaluator.IsDirectoryIgnored(ctx, directoryPathForRoot(evaluator.basePath, filePath))
	if err != nil {
		return ConfigFileResolution{}, err
	}
	if ignored {
		return ConfigFileResolution{Status: ConfigFileIgnored}, nil
	}
	ignored, err = evaluator.matchesGlobalIgnores(ctx, filePath, false)
	if err != nil {
		return ConfigFileResolution{}, err
	}
	if ignored {
		return ConfigFileResolution{Status: ConfigFileIgnored}, nil
	}

	matchingIndices := make([]int, 0, len(evaluator.config))
	matchFound := false
	for entryIndex, entry := range evaluator.config {
		if isGlobalIgnoreEntry(entry) {
			continue
		}
		compiled := evaluator.entries[entryIndex]
		paths, within := selectorMatchPaths(filePath, compiled.basePath, true, compiled.caseSensitive)
		if !within {
			continue
		}

		if !compiled.hasFiles {
			if len(compiled.ignores) > 0 {
				ignored, err := evaluator.reduceIgnores(ctx, false, compiled.ignores, filePath, paths, false)
				if err != nil {
					return ConfigFileResolution{}, err
				}
				if ignored {
					continue
				}
			}
			matchingIndices = append(matchingIndices, entryIndex)
			continue
		}

		if len(compiled.universal) > 0 {
			matched, err := evaluator.pathMatchesSelectors(ctx, compiled.nonUniversal, compiled.ignores, filePath, paths)
			if err != nil {
				return ConfigFileResolution{}, err
			}
			if matched {
				matchingIndices = append(matchingIndices, entryIndex)
				matchFound = true
				continue
			}
			matched, err = evaluator.pathMatchesSelectors(ctx, compiled.universal, compiled.ignores, filePath, paths)
			if err != nil {
				return ConfigFileResolution{}, err
			}
			if matched {
				matchingIndices = append(matchingIndices, entryIndex)
			}
			continue
		}

		matched, err := evaluator.pathMatchesSelectors(ctx, compiled.nonUniversal, compiled.ignores, filePath, paths)
		if err != nil {
			return ConfigFileResolution{}, err
		}
		if matched {
			matchingIndices = append(matchingIndices, entryIndex)
			matchFound = true
		}
	}

	if !matchFound {
		return ConfigFileResolution{Status: ConfigFileUnconfigured}, nil
	}
	return ConfigFileResolution{
		Status: ConfigFileConfigured,
		Config: evaluator.mergedConfigForIndices(matchingIndices),
	}, nil
}

func (evaluator *ConfigEvaluator) mergedConfigForIndices(indices []int) *MergedConfig {
	var key strings.Builder
	for _, index := range indices {
		key.WriteString(strconv.Itoa(index))
		key.WriteByte(',')
	}
	cacheKey := key.String()
	evaluator.mu.Lock()
	if cached := evaluator.mergedCache[cacheKey]; cached != nil {
		evaluator.mu.Unlock()
		<-cached.ready
		return cached.config
	}
	state := &mergedConfigState{ready: make(chan struct{})}
	evaluator.mergedCache[cacheKey] = state
	evaluator.mu.Unlock()
	state.config = mergeConfigEntryIndices(evaluator.config, indices)
	close(state.ready)
	return state.config
}

func (evaluator *ConfigEvaluator) pathMatchesSelectors(
	ctx context.Context,
	selectors []compiledConfigFileSelector,
	ignores []compiledIgnoreMatcher,
	filePath string,
	paths []string,
) (bool, error) {
	matched := false
	for _, selector := range selectors {
		selectorMatched := true
		for _, matcher := range selector.matchers {
			value, err := matcher.matches(ctx, paths, filePath, false, evaluator.resolver)
			if err != nil {
				return false, err
			}
			if !value {
				selectorMatched = false
				break
			}
		}
		if selectorMatched {
			matched = true
			break
		}
	}
	if !matched || len(ignores) == 0 {
		return matched, nil
	}
	ignored, err := evaluator.reduceIgnores(ctx, false, ignores, filePath, paths, false)
	return !ignored, err
}

// IsDirectoryIgnored mirrors ConfigArray.isDirectoryIgnored. Only an exact
// requested directory result is reused. A new descendant deliberately
// re-evaluates its ancestors even though those ancestor results were populated
// as a side effect of an earlier request.
func (evaluator *ConfigEvaluator) IsDirectoryIgnored(ctx context.Context, directory string) (bool, error) {
	if evaluator == nil {
		return false, nil
	}
	directory = normalizePathForRoot(evaluator.basePath, directory)
	if pathsEqualForRoot(evaluator.basePath, directory, evaluator.basePath, evaluator.caseSensitive) {
		return false, nil
	}
	if _, within := selectorMatchPaths(directory, evaluator.basePath, true, evaluator.caseSensitive); !within {
		return true, nil
	}

	var state *configEvaluationState
	for {
		evaluator.mu.Lock()
		if existing := evaluator.directoryCache[directory]; existing != nil {
			evaluator.mu.Unlock()
			select {
			case <-existing.ready:
				if existing.err != nil {
					continue
				}
				return existing.ignored, nil
			case <-ctx.Done():
				return false, ctx.Err()
			}
		}
		state = &configEvaluationState{ready: make(chan struct{})}
		evaluator.directoryCache[directory] = state
		evaluator.mu.Unlock()
		break
	}

	ancestors := pathAncestors(directory)
	start := 0
	for start < len(ancestors) && !pathsEqualForRoot(evaluator.basePath, ancestors[start], evaluator.basePath, evaluator.caseSensitive) {
		start++
	}
	if start < len(ancestors) {
		start++ // ConfigArray basePath itself is never checked.
	} else {
		state.err = errors.New("directory is outside config base path")
	}
	for index := start; state.err == nil && index < len(ancestors); index++ {
		ignored, err := evaluator.matchesGlobalIgnores(ctx, ancestors[index], true)
		if err != nil {
			state.err = err
			break
		}
		state.ignored = ignored
		evaluator.storeDirectorySideEffect(ancestors[index], ignored)
		if ignored {
			break
		}
	}

	evaluator.mu.Lock()
	if state.err != nil {
		delete(evaluator.directoryCache, directory)
	}
	close(state.ready)
	evaluator.mu.Unlock()
	return state.ignored, state.err
}

func (evaluator *ConfigEvaluator) storeDirectorySideEffect(directory string, ignored bool) {
	evaluator.mu.Lock()
	defer evaluator.mu.Unlock()
	if existing := evaluator.directoryCache[directory]; existing != nil {
		return
	}
	state := &configEvaluationState{ready: make(chan struct{}), ignored: ignored}
	close(state.ready)
	evaluator.directoryCache[directory] = state
}

func (evaluator *ConfigEvaluator) matchesGlobalIgnores(ctx context.Context, targetPath string, directory bool) (bool, error) {
	ignored := false
	if evaluator.gitignore != nil {
		for _, layer := range evaluator.gitignore.layersForTarget(targetPath, directory) {
			relative, within := layer.relativePath(targetPath, "", evaluator.fs)
			if !within {
				continue
			}
			paths := []string{relative}
			if directory {
				paths = []string{relative + "/", relative}
			}
			var err error
			ignored, err = evaluator.reduceIgnores(ctx, ignored, layer.matchers, targetPath, paths, directory)
			if err != nil {
				return false, err
			}
		}
	}
	for _, layer := range evaluator.globalIgnore {
		relative, within := layer.relativePath(targetPath, "", evaluator.fs)
		if !within {
			continue
		}
		paths := []string{relative}
		if directory {
			paths = []string{relative + "/", relative}
		}
		var err error
		ignored, err = evaluator.reduceIgnores(ctx, ignored, layer.matchers, targetPath, paths, directory)
		if err != nil {
			return false, err
		}
	}
	return ignored, nil
}

func (evaluator *ConfigEvaluator) reduceIgnores(
	ctx context.Context,
	ignored bool,
	matchers []compiledIgnoreMatcher,
	absolutePath string,
	paths []string,
	directory bool,
) (bool, error) {
	for _, matcher := range matchers {
		if !ignored {
			if matcher.predicateID != "" {
				if evaluator.resolver == nil {
					return false, fmt.Errorf("config predicate %q requires a JavaScript matcher host", matcher.predicateID)
				}
				value, err := evaluator.resolver.ResolveConfigPredicate(ctx, matcher.predicateID, absolutePath, directory)
				if err != nil {
					return false, err
				}
				ignored = value
				continue
			}
			if matcher.pattern != nil && !matcher.pattern.Negated && ignorePatternMatchesAny(*matcher.pattern, paths) {
				ignored = true
			}
			continue
		}

		// Once ignored, ConfigArray skips functions and positive strings. Only a
		// negated string may reopen the path.
		if matcher.pattern != nil && matcher.pattern.Negated && ignorePatternMatchesAny(*matcher.pattern, paths) {
			ignored = false
		}
	}
	return ignored, nil
}

func ignorePatternMatchesAny(pattern IgnorePattern, paths []string) bool {
	for _, path := range paths {
		if ignorePatternMatches(pattern, path) {
			return true
		}
	}
	return false
}
