package discovery

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
)

type blockingFixtureConfigLoader struct {
	delegate *fakeConfigModuleLoader
	started  chan struct{}
	release  chan struct{}
	once     sync.Once
	loads    atomic.Int32
}

type fakePredicateConfigModuleLoader struct {
	*fakeConfigModuleLoader
	predicateMu      sync.Mutex
	predicateBatches []ConfigPredicateBatchRequest
	predicateValue   func(ConfigPredicateCall) bool
	predicateError   func(ConfigPredicateCall) error
}

func (loader *fakePredicateConfigModuleLoader) EvaluateConfigPredicates(
	_ context.Context,
	request ConfigPredicateBatchRequest,
) (ConfigPredicateBatchResponse, error) {
	loader.predicateMu.Lock()
	loader.predicateBatches = append(loader.predicateBatches, request)
	loader.predicateMu.Unlock()
	response := ConfigPredicateBatchResponse{TransactionID: request.TransactionID}
	for _, call := range request.Calls {
		if loader.predicateError != nil {
			if err := loader.predicateError(call); err != nil {
				response.Results = append(response.Results, ConfigPredicateResult{
					CallID: call.CallID,
					Status: "failed",
					Error:  &ConfigModuleError{Message: err.Error()},
				})
				continue
			}
		}
		value := tspath.GetBaseFileName(call.AbsolutePath) == "keep.ts"
		if loader.predicateValue != nil {
			value = loader.predicateValue(call)
		}
		response.Results = append(response.Results, ConfigPredicateResult{
			CallID: call.CallID,
			Status: "evaluated",
			Value:  value,
		})
	}
	return response, nil
}

func (loader *blockingFixtureConfigLoader) LoadConfigs(ctx context.Context, request ConfigLoadBatchRequest) (ConfigLoadBatchResponse, error) {
	loader.loads.Add(1)
	loader.once.Do(func() { close(loader.started) })
	select {
	case <-loader.release:
		return loader.delegate.LoadConfigs(ctx, request)
	case <-ctx.Done():
		return ConfigLoadBatchResponse{}, ctx.Err()
	}
}

func (loader *blockingFixtureConfigLoader) ActivateConfigs(ctx context.Context, request ConfigActivationRequest) (ConfigActivationResponse, error) {
	return loader.delegate.ActivateConfigs(ctx, request)
}

func (loader *blockingFixtureConfigLoader) EvaluateConfigPredicates(
	ctx context.Context,
	request ConfigPredicateBatchRequest,
) (ConfigPredicateBatchResponse, error) {
	return loader.delegate.EvaluateConfigPredicates(ctx, request)
}

type signalingWalkFS struct {
	vfs.FS
	started chan struct{}
	once    sync.Once
}

type realpathCountingFS struct {
	vfs.FS
	mu    sync.Mutex
	calls map[string]int
}

func (fsys *realpathCountingFS) Realpath(path string) string {
	path = rslintconfig.NormalizeHostPath(path)
	fsys.mu.Lock()
	fsys.calls[path]++
	fsys.mu.Unlock()
	return fsys.FS.Realpath(path)
}

func (fsys *realpathCountingFS) snapshot() map[string]int {
	fsys.mu.Lock()
	defer fsys.mu.Unlock()
	result := make(map[string]int, len(fsys.calls))
	for path, count := range fsys.calls {
		result[path] = count
	}
	return result
}

type failFastConfigLoader struct {
	delegate     *fakeConfigModuleLoader
	hangingPath  string
	failingPath  string
	hangStarted  chan struct{}
	hangFinished chan struct{}
	hangOnce     sync.Once
	finishOnce   sync.Once
}

func (loader *failFastConfigLoader) LoadConfigs(ctx context.Context, request ConfigLoadBatchRequest) (ConfigLoadBatchResponse, error) {
	if len(request.Candidates) != 1 {
		return ConfigLoadBatchResponse{}, fmt.Errorf("expected unary config load, got %d candidates", len(request.Candidates))
	}
	candidate := request.Candidates[0]
	switch candidate.ConfigPath {
	case loader.hangingPath:
		loader.hangOnce.Do(func() { close(loader.hangStarted) })
		<-ctx.Done()
		loader.finishOnce.Do(func() { close(loader.hangFinished) })
		return ConfigLoadBatchResponse{}, ctx.Err()
	case loader.failingPath:
		select {
		case <-loader.hangStarted:
		case <-ctx.Done():
			return ConfigLoadBatchResponse{}, ctx.Err()
		}
		return ConfigLoadBatchResponse{
			TransactionID: request.TransactionID,
			Results: []ConfigLoadResult{{
				ID:     candidate.ID,
				Status: "failed",
				Error:  &ConfigModuleError{Code: "load", Message: "controlled failure"},
			}},
		}, nil
	default:
		return loader.delegate.LoadConfigs(ctx, request)
	}
}

func (loader *failFastConfigLoader) ActivateConfigs(ctx context.Context, request ConfigActivationRequest) (ConfigActivationResponse, error) {
	return loader.delegate.ActivateConfigs(ctx, request)
}

func (loader *failFastConfigLoader) EvaluateConfigPredicates(ctx context.Context, request ConfigPredicateBatchRequest) (ConfigPredicateBatchResponse, error) {
	return loader.delegate.EvaluateConfigPredicates(ctx, request)
}

func (fsys *signalingWalkFS) GetAccessibleEntries(path string) vfs.Entries {
	fsys.once.Do(func() { close(fsys.started) })
	return fsys.FS.GetAccessibleEntries(path)
}

func buildSearchFixtureCatalog(
	t *testing.T,
	root string,
	loader ConfigModuleLoader,
	inputs []string,
	mutate func(*ConfigDiscoveryRequest),
) (*ConfigCatalog, error) {
	t.Helper()
	request := ConfigDiscoveryRequest{
		CWD:                     rslintconfig.NormalizeHostPath(root),
		Mode:                    ConfigDiscoveryAuto,
		Inputs:                  inputs,
		CollectTargets:          true,
		GlobInputPaths:          true,
		ErrorOnUnmatchedPattern: true,
		SingleThreaded:          true,
	}
	if mutate != nil {
		mutate(&request)
	}
	return Build(context.Background(), discoveryTestFS(), loader, request)
}

func TestSearchDiscoveryCarriesOneLiveMatcherSelectionIntoTargets(t *testing.T) {
	root := t.TempDir()
	configPath := writeConfigCandidate(t, root, "rslint.config.js")
	keep := writeDiscoveryFixture(t, root, "keep.ts", "debugger;\n")
	drop := writeDiscoveryFixture(t, root, "drop.ts", "debugger;\n")
	moduleConfig, err := rslintconfig.DecodeModuleConfig([]byte(`[
		{"files":[{"$rslintPredicate":"files-1"}],"rules":{"no-debugger":"error"}}
	]`))
	if err != nil {
		t.Fatal(err)
	}
	baseLoader := newFixtureConfigLoader()
	baseLoader.configs[configPath] = moduleConfig
	loader := &fakePredicateConfigModuleLoader{fakeConfigModuleLoader: baseLoader}
	catalog, err := Build(context.Background(), discoveryTestFS(), loader, ConfigDiscoveryRequest{
		CWD: root, Mode: ConfigDiscoveryAuto, Inputs: []string{"*.ts"},
		CollectTargets: true, GlobInputPaths: true, ErrorOnUnmatchedPattern: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer catalog.ClosePredicateEvaluation()
	if len(catalog.Targets) != 2 {
		t.Fatalf("targets = %+v, want keep and drop", catalog.Targets)
	}
	byPath := make(map[string]DiscoveredTarget, len(catalog.Targets))
	for _, target := range catalog.Targets {
		byPath[target.Path] = target
		if target.MergedConfig == nil {
			t.Fatalf("target %q has no merged config", target.Path)
		}
	}
	if _, enabled := byPath[keep].MergedConfig.Rules["no-debugger"]; !enabled {
		t.Fatalf("keep target did not retain predicate-selected entry: %+v", byPath[keep].MergedConfig)
	}
	if _, enabled := byPath[drop].MergedConfig.Rules["no-debugger"]; enabled {
		t.Fatalf("drop target retained predicate-selected entry: %+v", byPath[drop].MergedConfig)
	}

	loader.predicateMu.Lock()
	callCount := 0
	for _, batch := range loader.predicateBatches {
		for _, call := range batch.Calls {
			callCount++
			if !tspath.IsRootedDiskPath(call.AbsolutePath) || call.Directory {
				t.Fatalf("predicate call = %+v, want absolute file path", call)
			}
		}
	}
	loader.predicateMu.Unlock()
	// ESLint asks ConfigArray for every filesystem entry before applying the
	// input glob, so the config module itself is the third evaluated file.
	if callCount != 3 {
		t.Fatalf("predicate calls = %d, want one per visited file", callCount)
	}

	evaluator := catalog.ConfigEvaluatorForDirectory(tspath.NormalizePath(root))
	if evaluator == nil {
		t.Fatal("catalog has no evaluator")
	}
	for _, path := range []string{keep, drop} {
		if _, err := evaluator.GetConfigForFile(context.Background(), path); err != nil {
			t.Fatal(err)
		}
	}
	loader.predicateMu.Lock()
	defer loader.predicateMu.Unlock()
	secondCount := 0
	for _, batch := range loader.predicateBatches {
		secondCount += len(batch.Calls)
	}
	if secondCount != callCount {
		t.Fatalf("cached target lookup added predicate calls: before=%d after=%d", callCount, secondCount)
	}
}

func TestSearchDiscoveryGlobUsesMinimatchUTF16Units(t *testing.T) {
	root := t.TempDir()
	configPath := writeConfigCandidate(t, root, "rslint.config.js")
	targetPath := writeDiscoveryFixture(t, root, "😀.js", "export {}\n")
	loader := newFixtureConfigLoader()
	loader.configs[configPath] = namedConfig("root")

	// is-glob deliberately classifies bare ? as literal in strict mode, so the
	// terminal * makes this a search glob while the two ? still carry the
	// UTF-16 boundary under test.
	catalog, err := buildSearchFixtureCatalog(t, root, loader, []string{"??.js*"}, nil)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	defer catalog.ClosePredicateEvaluation()
	if got := discoveredTargetPaths(catalog.Targets); !reflect.DeepEqual(got, []string{targetPath}) {
		t.Fatalf("targets = %v, want astral filename selected by two UTF-16 wildcards", got)
	}
}

func TestSearchDiscoveryDefersExplicitPredicatesUntilGlobSearchCompletes(t *testing.T) {
	root := t.TempDir()
	configPath := writeConfigCandidate(t, root, "rslint.config.js")
	explicit := writeDiscoveryFixture(t, root, "explicit.ts", "export {}\n")
	writeDiscoveryFixture(t, root, "glob/g.ts", "export {}\n")
	moduleConfig, err := rslintconfig.DecodeModuleConfig([]byte(`[
		{"files":[{"$rslintPredicate":"files-1"}],"rules":{"no-debugger":"error"}}
	]`))
	if err != nil {
		t.Fatal(err)
	}

	t.Run("successful glob runs before explicit selection", func(t *testing.T) {
		baseLoader := newFixtureConfigLoader()
		baseLoader.configs[configPath] = moduleConfig
		loader := &fakePredicateConfigModuleLoader{
			fakeConfigModuleLoader: baseLoader,
			predicateValue: func(ConfigPredicateCall) bool {
				return true
			},
		}
		catalog, err := buildSearchFixtureCatalog(t, root, loader, []string{explicit, "glob"}, nil)
		if err != nil {
			t.Fatalf("Build: %v", err)
		}
		defer catalog.ClosePredicateEvaluation()
		loader.predicateMu.Lock()
		var names []string
		for _, batch := range loader.predicateBatches {
			for _, call := range batch.Calls {
				names = append(names, tspath.GetBaseFileName(call.AbsolutePath))
			}
		}
		loader.predicateMu.Unlock()
		if want := []string{"g.ts", "explicit.ts"}; !reflect.DeepEqual(names, want) {
			t.Fatalf("predicate order = %v, want glob discovery before explicit selection %v", names, want)
		}
	})

	t.Run("glob failure skips explicit selection", func(t *testing.T) {
		baseLoader := newFixtureConfigLoader()
		baseLoader.configs[configPath] = moduleConfig
		loader := &fakePredicateConfigModuleLoader{
			fakeConfigModuleLoader: baseLoader,
			predicateError: func(call ConfigPredicateCall) error {
				if tspath.GetBaseFileName(call.AbsolutePath) == "g.ts" {
					return errors.New("glob predicate boom")
				}
				return nil
			},
		}
		_, err := buildSearchFixtureCatalog(t, root, loader, []string{explicit, "glob"}, nil)
		if err == nil || !strings.Contains(err.Error(), "glob predicate boom") {
			t.Fatalf("error = %v, want glob predicate failure", err)
		}
		loader.predicateMu.Lock()
		defer loader.predicateMu.Unlock()
		for _, batch := range loader.predicateBatches {
			for _, call := range batch.Calls {
				if call.AbsolutePath == explicit {
					t.Fatal("explicit predicate ran after glob discovery had already failed")
				}
			}
		}
	})
}

func TestSearchDiscoveryDirectRootCanReachNestedConfig(t *testing.T) {
	root := t.TempDir()
	rootConfig := writeConfigCandidate(t, root, "rslint.config.js")
	appConfig := writeConfigCandidate(t, root, "packages/app/rslint.config.js")
	appFile := writeDiscoveryFixture(t, root, "packages/app/index.ts", "export {}\n")

	loader := newFixtureConfigLoader()
	loader.configs[rootConfig] = rslintconfig.RslintConfig{{
		Ignores: []string{"packages/app/**"},
	}}
	loader.configs[appConfig] = namedConfig("app")

	catalog, err := buildSearchFixtureCatalog(t, root, loader, []string{"packages/app"}, nil)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if _, loaded := catalog.Configs[tspath.GetDirectoryPath(appConfig)]; !loaded {
		t.Fatalf("direct search root did not reach its own config: %v", catalog.ConfigDirectories())
	}
	want := []string{appConfig, appFile}
	sort.Strings(want)
	if got := discoveredTargetPaths(catalog.Targets); !reflect.DeepEqual(got, want) {
		t.Fatalf("targets = %v, want app config and source", got)
	}
	if got := requestedConfigPaths(loader); !reflect.DeepEqual(got, []string{appConfig}) {
		t.Fatalf("loaded configs = %v, want only nearest app config", got)
	}
}

func TestSearchDiscoveryDynamicGlobUsesParentGate(t *testing.T) {
	root := t.TempDir()
	rootConfig := writeConfigCandidate(t, root, "rslint.config.js")
	appConfig := writeConfigCandidate(t, root, "packages/app/rslint.config.js")
	writeDiscoveryFixture(t, root, "packages/app/index.ts", "export {}\n")

	loader := newFixtureConfigLoader()
	loader.configs[rootConfig] = rslintconfig.RslintConfig{{
		Ignores: []string{"packages/app/**"},
	}}
	loader.configs[appConfig] = namedConfig("app")

	_, err := buildSearchFixtureCatalog(t, root, loader, []string{"packages/*/*.ts"}, nil)
	var ignored *AllFilesIgnoredError
	if !errors.As(err, &ignored) {
		t.Fatalf("error = %v, want AllFilesIgnoredError", err)
	}
	if got := requestedConfigPaths(loader); !reflect.DeepEqual(got, []string{rootConfig}) {
		t.Fatalf("loaded configs = %v, nested config must remain behind the parent gate", got)
	}
}

func TestSearchDiscoveryNearestBrokenConfigIsFatal(t *testing.T) {
	root := t.TempDir()
	rootConfig := writeConfigCandidate(t, root, "rslint.config.js")
	appConfig := writeConfigCandidate(t, root, "app/rslint.config.js")
	writeDiscoveryFixture(t, root, "app/index.ts", "export {}\n")

	loader := newFixtureConfigLoader()
	loader.configs[rootConfig] = namedConfig("root")
	loader.failures[appConfig] = ConfigModuleError{Code: "load", Message: "broken nearest"}

	_, err := buildSearchFixtureCatalog(t, root, loader, []string{"app/index.ts"}, nil)
	if !errors.Is(err, ErrAllConfigsFailed) {
		t.Fatalf("error = %v, want strict nearest config failure", err)
	}
	if got := requestedConfigPaths(loader); !reflect.DeepEqual(got, []string{appConfig}) {
		t.Fatalf("loaded configs = %v, ancestor fallback must not occur", got)
	}
}

func TestSearchDiscoveryEmptyRootDoesNotLoadAncestorConfig(t *testing.T) {
	root := t.TempDir()
	rootConfig := writeConfigCandidate(t, root, "rslint.config.js")
	emptyDir := filepath.Join(root, "empty")
	if err := os.MkdirAll(emptyDir, 0o755); err != nil {
		t.Fatal(err)
	}
	loader := newFixtureConfigLoader()
	loader.configs[rootConfig] = namedConfig("root")

	_, err := buildSearchFixtureCatalog(t, root, loader, []string{"empty"}, nil)
	var noFiles *NoFilesFoundError
	if !errors.As(err, &noFiles) {
		t.Fatalf("error = %v, want unmatched direct-directory search", err)
	}
	if len(loader.batches) != 0 {
		t.Fatalf("empty search root executed config modules: %v", requestedConfigPaths(loader))
	}
}

func TestSearchDiscoveryExplicitConfigLoadsOnlyAfterFindFilesReachesAnEntry(t *testing.T) {
	root := t.TempDir()
	configPath := writeConfigCandidate(t, root, "broken.config.js")
	loader := newFixtureConfigLoader()
	loader.failures[configPath] = ConfigModuleError{Code: "load", Message: "broken explicit"}
	explicit := func(request *ConfigDiscoveryRequest) {
		request.Mode = ConfigDiscoveryExplicit
		request.ExplicitConfigPath = configPath
	}

	_, err := buildSearchFixtureCatalog(t, root, loader, []string{"absent/**/*.ts"}, explicit)
	var noFiles *NoFilesFoundError
	if !errors.As(err, &noFiles) {
		t.Fatalf("unmatched error = %v, want NoFilesFoundError before explicit config evaluation", err)
	}
	if len(loader.batches) != 0 {
		t.Fatalf("unmatched search loaded explicit config: %v", requestedConfigPaths(loader))
	}

	writeDiscoveryFixture(t, root, "src/index.ts", "export {};\n")
	_, err = buildSearchFixtureCatalog(t, root, loader, []string{"src/*.ts"}, explicit)
	if !errors.Is(err, ErrAllConfigsFailed) || !strings.Contains(err.Error(), "broken explicit") {
		t.Fatalf("viable search error = %v, want explicit config failure", err)
	}
	if got := requestedConfigPaths(loader); !reflect.DeepEqual(got, []string{configPath}) {
		t.Fatalf("loaded configs = %v, want one lazy explicit load", got)
	}
}

func TestSearchDiscoveryUnmatchedGlobDoesNotWaitForExplicitConfigLoad(t *testing.T) {
	root := t.TempDir()
	configPath := writeConfigCandidate(t, root, "hanging.config.js")
	delegate := newFixtureConfigLoader()
	delegate.configs[configPath] = namedConfig("explicit")
	loader := &blockingFixtureConfigLoader{
		delegate: delegate,
		started:  make(chan struct{}),
		release:  make(chan struct{}),
	}
	done := make(chan error, 1)
	go func() {
		_, err := buildSearchFixtureCatalog(t, root, loader, []string{"absent/**/*.ts"}, func(request *ConfigDiscoveryRequest) {
			request.Mode = ConfigDiscoveryExplicit
			request.ExplicitConfigPath = configPath
		})
		done <- err
	}()
	defer close(loader.release)

	select {
	case err := <-done:
		var noFiles *NoFilesFoundError
		if !errors.As(err, &noFiles) {
			t.Fatalf("error = %v, want unmatched glob", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("unmatched glob waited for explicit config module")
	}
	if got := loader.loads.Load(); got != 0 {
		t.Fatalf("explicit config loads = %d, want zero", got)
	}
}

func TestSearchDiscoveryExplicitConfigUsesCWDForBasePath(t *testing.T) {
	root := t.TempDir()
	configPath := writeConfigCandidate(t, root, "configs/rslint.config.js")
	target := writeDiscoveryFixture(t, root, "packages/app/index.ts", "export {};\n")
	loader := newFixtureConfigLoader()
	loader.configs[configPath] = rslintconfig.RslintConfig{{
		Files: []string{"packages/**/*.ts"},
		Rules: rslintconfig.Rules{"base-path": "error"},
	}}
	catalog, err := buildSearchFixtureCatalog(t, root, loader, []string{"packages/**/*.ts"}, func(request *ConfigDiscoveryRequest) {
		request.Mode = ConfigDiscoveryExplicit
		request.ExplicitConfigPath = configPath
	})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if len(catalog.Targets) != 1 || catalog.Targets[0].Path != target ||
		catalog.Targets[0].ConfigDirectory != rslintconfig.NormalizeHostPath(root) ||
		catalog.Targets[0].MergedConfig == nil {
		t.Fatalf("targets = %+v, want explicit owner rooted at cwd", catalog.Targets)
	}
}

func TestSearchDiscoveryCatalogProbeLoadsAncestorForEmptyRoot(t *testing.T) {
	root := t.TempDir()
	rootConfig := writeConfigCandidate(t, root, "rslint.config.js")
	emptyDir := filepath.Join(root, "empty")
	if err := os.MkdirAll(emptyDir, 0o755); err != nil {
		t.Fatal(err)
	}
	loader := newFixtureConfigLoader()
	loader.configs[rootConfig] = namedConfig("root")

	catalog, err := buildSearchFixtureCatalog(t, emptyDir, loader, []string{"."}, func(request *ConfigDiscoveryRequest) {
		request.CollectTargets = false
		request.ErrorOnUnmatchedPattern = false
		request.AllowMissingConfig = true
		request.ProbeRootConfig = true
	})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if got := requestedConfigPaths(loader); !reflect.DeepEqual(got, []string{rootConfig}) {
		t.Fatalf("loaded configs = %v, want ancestor probe %q", got, rootConfig)
	}
	if _, ok := catalog.Configs[tspath.GetDirectoryPath(rootConfig)]; !ok {
		t.Fatalf("catalog configs = %v, want ancestor owner", catalog.ConfigDirectories())
	}
}

func TestSearchDiscoveryOverrideCanReopenDefaultNodeModules(t *testing.T) {
	root := t.TempDir()
	configPath := writeConfigCandidate(t, root, "rslint.config.js")
	targetPath := writeDiscoveryFixture(t, root, "node_modules/local/index.ts", "export {}\n")
	loader := newFixtureConfigLoader()
	loader.configs[configPath] = namedConfig("root")

	catalog, err := buildSearchFixtureCatalog(t, root, loader, []string{"."}, func(request *ConfigDiscoveryRequest) {
		request.OverrideConfig = rslintconfig.RslintConfig{{Ignores: []string{"!node_modules/"}}}
	})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if !containsPath(discoveredTargetPaths(catalog.Targets), targetPath) {
		t.Fatalf("targets = %v, negated default node_modules ignore did not reopen target", discoveredTargetPaths(catalog.Targets))
	}
}

func TestSearchDiscoveryMissingOwnerStillAppliesTraversalIgnores(t *testing.T) {
	t.Run("defaults prune config candidates in node_modules and git", func(t *testing.T) {
		root := t.TempDir()
		visibleConfig := writeConfigCandidate(t, root, "packages/app/rslint.config.js")
		nodeModulesConfig := writeConfigCandidate(t, root, "node_modules/local/rslint.config.js")
		gitConfig := writeConfigCandidate(t, root, ".git/generated/rslint.config.js")
		writeDiscoveryFixture(t, root, "packages/app/index.ts", "export {}\n")
		writeDiscoveryFixture(t, root, "node_modules/local/index.ts", "export {}\n")
		writeDiscoveryFixture(t, root, ".git/generated/index.ts", "export {}\n")

		loader := newFixtureConfigLoader()
		loader.configs[visibleConfig] = namedConfig("visible")
		loader.configs[nodeModulesConfig] = namedConfig("node_modules")
		loader.configs[gitConfig] = namedConfig("git")
		_, err := buildSearchFixtureCatalog(t, root, loader, []string{"."}, func(request *ConfigDiscoveryRequest) {
			request.AllowMissingConfig = true
			request.RetainUnconfiguredAreas = true
		})
		if err != nil {
			t.Fatalf("Build: %v", err)
		}
		if got := requestedConfigPaths(loader); !reflect.DeepEqual(got, []string{visibleConfig}) {
			t.Fatalf("loaded configs = %v, missing-owner defaults must prune %q and %q", got, nodeModulesConfig, gitConfig)
		}
	})

	t.Run("authored negation reopens default node_modules", func(t *testing.T) {
		root := t.TempDir()
		configPath := writeConfigCandidate(t, root, "node_modules/local/rslint.config.js")
		writeDiscoveryFixture(t, root, "node_modules/local/index.ts", "export {}\n")
		loader := newFixtureConfigLoader()
		loader.configs[configPath] = namedConfig("node_modules")

		_, err := buildSearchFixtureCatalog(t, root, loader, []string{"."}, func(request *ConfigDiscoveryRequest) {
			request.AllowMissingConfig = true
			request.RetainUnconfiguredAreas = true
			request.OverrideConfig = rslintconfig.RslintConfig{{Ignores: []string{"!node_modules/"}}}
		})
		if err != nil {
			t.Fatalf("Build: %v", err)
		}
		if got := requestedConfigPaths(loader); !reflect.DeepEqual(got, []string{configPath}) {
			t.Fatalf("loaded configs = %v, want reopened nested config %q", got, configPath)
		}
	})

	t.Run("live override ignore gates nested config discovery", func(t *testing.T) {
		root := t.TempDir()
		visibleConfig := writeConfigCandidate(t, root, "visible/rslint.config.js")
		blockedConfig := writeConfigCandidate(t, root, "blocked/rslint.config.js")
		writeDiscoveryFixture(t, root, "visible/index.ts", "export {}\n")
		writeDiscoveryFixture(t, root, "blocked/index.ts", "export {}\n")
		override, err := rslintconfig.DecodeModuleConfig([]byte(`[
			{"ignores":[{"$rslintPredicate":"ignore-blocked"}]}
		]`))
		if err != nil {
			t.Fatal(err)
		}
		baseLoader := newFixtureConfigLoader()
		baseLoader.configs[visibleConfig] = namedConfig("visible")
		baseLoader.configs[blockedConfig] = namedConfig("blocked")
		loader := &fakePredicateConfigModuleLoader{
			fakeConfigModuleLoader: baseLoader,
			predicateValue: func(call ConfigPredicateCall) bool {
				return call.Directory && tspath.GetBaseFileName(call.AbsolutePath) == "blocked"
			},
		}

		catalog, err := buildSearchFixtureCatalog(t, root, loader, []string{"."}, func(request *ConfigDiscoveryRequest) {
			request.AllowMissingConfig = true
			request.RetainUnconfiguredAreas = true
			request.OverrideConfig = override
		})
		if err != nil {
			t.Fatalf("Build: %v", err)
		}
		defer catalog.ClosePredicateEvaluation()
		if got := requestedConfigPaths(baseLoader); !reflect.DeepEqual(got, []string{visibleConfig}) {
			t.Fatalf("loaded configs = %v, live ignore must prune %q", got, blockedConfig)
		}
		loader.predicateMu.Lock()
		defer loader.predicateMu.Unlock()
		calls := 0
		for _, batch := range loader.predicateBatches {
			calls += len(batch.Calls)
		}
		if calls == 0 {
			t.Fatal("missing-owner traversal never evaluated the live override ignore")
		}
	})
}

func TestSearchDiscoveryHonorsScopedBasePathSelectorsAndIgnores(t *testing.T) {
	root := t.TempDir()
	configPath := writeConfigCandidate(t, root, "rslint.config.js")
	keep := writeDiscoveryFixture(t, root, "packages/app/src/keep.JS", "export {}\n")
	drop := writeDiscoveryFixture(t, root, "packages/app/src/drop.JS", "export {}\n")
	blocked := writeDiscoveryFixture(t, root, "packages/app/blocked/hidden.JS", "export {}\n")
	other := writeDiscoveryFixture(t, root, "packages/other/src/keep.JS", "export {}\n")

	loader := newFixtureConfigLoader()
	loader.configs[configPath] = rslintconfig.RslintConfig{
		{
			BasePath: "packages/app",
			Files:    []string{"src/**/*.JS"},
			Ignores:  []string{"src/drop.JS"},
			Rules:    rslintconfig.Rules{"scoped": "error"},
		},
		{BasePath: "packages/app", Ignores: []string{"blocked/**"}},
	}

	catalog, err := buildSearchFixtureCatalog(t, root, loader, []string{"."}, nil)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	targets := discoveredTargetPaths(catalog.Targets)
	if !containsPath(targets, keep) {
		t.Fatalf("targets = %v, scoped selected file is missing", targets)
	}
	for _, unexpected := range []string{drop, blocked, other} {
		if containsPath(targets, unexpected) {
			t.Fatalf("targets = %v, scoped basePath leaked or ignored file was admitted: %s", targets, unexpected)
		}
	}
}

func TestSearchDiscoveryMergedBasePatternsDoNotNarrowEachOther(t *testing.T) {
	root := t.TempDir()
	configPath := writeConfigCandidate(t, root, "rslint.config.js")
	jsPath := writeDiscoveryFixture(t, root, "subdir/a.js", "export {}\n")
	cjsPath := writeDiscoveryFixture(t, root, "other/b.cjs", "module.exports = {}\n")
	loader := newFixtureConfigLoader()
	loader.configs[configPath] = namedConfig("root")

	catalog, err := buildSearchFixtureCatalog(t, root, loader, []string{"subdir/*.js", "**/*.cjs"}, nil)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	got := discoveredTargetPaths(catalog.Targets)
	want := []string{cjsPath, jsPath}
	sort.Strings(want)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("targets = %v, want %v", got, want)
	}
}

func TestSearchDiscoveryLeadingNegatedPatternStaysUnmatchedLikeESLint(t *testing.T) {
	root := t.TempDir()
	configPath := writeConfigCandidate(t, root, "rslint.config.js")
	writeDiscoveryFixture(t, root, "target.ts", "export {}\n")
	writeDiscoveryFixture(t, root, "other.ts", "export {}\n")
	loader := newFixtureConfigLoader()
	loader.configs[configPath] = namedConfig("root")

	_, err := buildSearchFixtureCatalog(t, root, loader, []string{"!target.ts"}, nil)
	var ignored *AllFilesIgnoredError
	if !errors.As(err, &ignored) {
		t.Fatalf("error = %v, want ESLint's leading-negation unmatched error", err)
	}
}

func TestSearchDiscoveryLookupPathsBypassParentAndDefaultIgnoreGates(t *testing.T) {
	root := t.TempDir()
	rootConfig := writeConfigCandidate(t, root, "rslint.config.js")
	hiddenConfig := writeConfigCandidate(t, root, "hidden/app/rslint.config.js")
	modulesConfig := writeConfigCandidate(t, root, "node_modules/local/rslint.config.js")
	writeDiscoveryFixture(t, root, "visible.ts", "export {};\n")

	loader := newFixtureConfigLoader()
	loader.configs[rootConfig] = rslintconfig.RslintConfig{
		{Ignores: []string{"hidden/**"}},
		{Files: []string{"**/*.ts"}},
	}
	loader.configs[hiddenConfig] = namedConfig("hidden")
	loader.configs[modulesConfig] = namedConfig("modules")

	catalog, err := buildSearchFixtureCatalog(t, root, loader, []string{"."}, func(request *ConfigDiscoveryRequest) {
		request.LookupPaths = []string{
			filepath.Join(root, "hidden/app/virtual.ts"),
			filepath.Join(root, "node_modules/local/virtual.ts"),
		}
	})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	for _, configPath := range []string{rootConfig, hiddenConfig, modulesConfig} {
		configDir := tspath.GetDirectoryPath(configPath)
		if _, ok := catalog.Configs[configDir]; !ok {
			t.Fatalf("exact lookup did not retain %q; configs=%v", configDir, catalog.ConfigDirectories())
		}
	}
	wantTargets := []string{rootConfig, tspath.NormalizePath(filepath.Join(root, "visible.ts"))}
	sort.Strings(wantTargets)
	if got := discoveredTargetPaths(catalog.Targets); !reflect.DeepEqual(got, wantTargets) {
		t.Fatalf("LookupPaths must not become lint targets: got=%v want ordinary search results=%v", got, wantTargets)
	}
}

func TestSearchDiscoveryCollectConfigFailuresKeepsSiblingSuccess(t *testing.T) {
	root := t.TempDir()
	goodConfig := writeConfigCandidate(t, root, "good/rslint.config.js")
	badConfig := writeConfigCandidate(t, root, "bad/rslint.config.js")
	goodFile := writeDiscoveryFixture(t, root, "good/index.ts", "export {};\n")
	writeDiscoveryFixture(t, root, "bad/index.ts", "export {};\n")

	loader := newFixtureConfigLoader()
	loader.configs[goodConfig] = namedConfig("good")
	loader.failures[badConfig] = ConfigModuleError{Code: "load", Message: "broken sibling"}
	catalog, err := buildSearchFixtureCatalog(t, root, loader, []string{"good/index.ts", "bad/index.ts"}, func(request *ConfigDiscoveryRequest) {
		request.CollectConfigFailures = true
		request.ErrorOnUnmatchedPattern = false
	})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if _, ok := catalog.Configs[tspath.GetDirectoryPath(goodConfig)]; !ok {
		t.Fatalf("successful sibling missing from catalog: %v", catalog.ConfigDirectories())
	}
	if len(catalog.Failures) != 1 || catalog.Failures[0].Path != badConfig || catalog.Failures[0].Directory != tspath.GetDirectoryPath(badConfig) {
		t.Fatalf("failures = %+v, want failed nearest boundary", catalog.Failures)
	}
	if got := discoveredTargetPaths(catalog.Targets); !reflect.DeepEqual(got, []string{goodFile}) {
		t.Fatalf("targets = %v, want only successful sibling", got)
	}
}

func TestSearchDiscoveryMixedConfiguredGlobReportsMissingConfigFirst(t *testing.T) {
	root := t.TempDir()
	missingFile := writeDiscoveryFixture(t, root, "unconfigured/a.ts", "export {};\n")
	configPath := writeConfigCandidate(t, root, "configured/rslint.config.js")
	writeDiscoveryFixture(t, root, "configured/b.ts", "export {};\n")
	loader := newFixtureConfigLoader()
	loader.configs[configPath] = namedConfig("configured")

	_, err := buildSearchFixtureCatalog(t, root, loader, []string{"unconfigured/*.ts", "configured/*.ts"}, func(request *ConfigDiscoveryRequest) {
		request.AllowMissingConfig = true
	})
	var missing *ConfigFileMissingError
	if !errors.As(err, &missing) || missing.Path != missingFile {
		t.Fatalf("error = %v, want missing config for %q before unmatched-pattern errors", err, missingFile)
	}
}

func TestSearchDiscoveryOverridePatternsUseNestedOwnerAsBase(t *testing.T) {
	root := t.TempDir()
	configPath := writeConfigCandidate(t, root, "packages/app/rslint.config.js")
	target := writeDiscoveryFixture(t, root, "packages/app/index.ts", "export {};\n")
	loader := newFixtureConfigLoader()
	loader.configs[configPath] = namedConfig("app")

	catalog, err := buildSearchFixtureCatalog(t, root, loader, []string{"packages/app/index.ts"}, func(request *ConfigDiscoveryRequest) {
		request.OverrideConfig = rslintconfig.RslintConfig{{Ignores: []string{"packages/app/**"}}}
	})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if got := discoveredTargetPaths(catalog.Targets); !reflect.DeepEqual(got, []string{target}) {
		t.Fatalf("overrideConfig must be relative to nested owner, targets=%v", got)
	}
}

func TestSearchDiscoveryNormalizesPOSIXBackslashLikeESLintGlobSearch(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows uses backslash as a path separator")
	}
	root := t.TempDir()
	configPath := writeConfigCandidate(t, root, "rslint.config.js")
	writeDiscoveryFixture(t, root, "literal*/nested/index.js", "export {};\n")
	loader := newFixtureConfigLoader()
	loader.configs[configPath] = namedConfig("root")

	_, err := buildSearchFixtureCatalog(t, root, loader, []string{`literal\*/**/*.js`}, nil)
	var noFiles *NoFilesFoundError
	if !errors.As(err, &noFiles) || noFiles.Pattern != "literal/*/**/*.js" {
		t.Fatalf("error = %#v, want normalized ESLint NoFilesFound pattern", err)
	}
}

func TestSearchDiscoveryPreservesNativePOSIXBackslashTargetAndPredicatePath(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows uses backslash as a path separator")
	}
	root := t.TempDir()
	configPath := writeConfigCandidate(t, root, "rslint.config.js")
	target := filepath.Join(root, `a\b.js`)
	if err := os.WriteFile(target, []byte("debugger;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// A slash sibling proves that no fallback through TypeScript's normalized
	// VFS can silently select the wrong filesystem object.
	writeDiscoveryFixture(t, root, "a/b.js", "export {};\n")
	moduleConfig, err := rslintconfig.DecodeModuleConfig([]byte(`[
		{"files":[{"$rslintPredicate":"files-1"}],"rules":{"no-debugger":"error"}}
	]`))
	if err != nil {
		t.Fatal(err)
	}
	baseLoader := newFixtureConfigLoader()
	baseLoader.configs[configPath] = moduleConfig
	loader := &fakePredicateConfigModuleLoader{
		fakeConfigModuleLoader: baseLoader,
		predicateValue:         func(ConfigPredicateCall) bool { return true },
	}

	catalog, err := Build(context.Background(), discoveryTestFS(), loader, ConfigDiscoveryRequest{
		CWD: root, Mode: ConfigDiscoveryAuto, Inputs: []string{`a\b.js`},
		CollectTargets: true, GlobInputPaths: true, ErrorOnUnmatchedPattern: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer catalog.ClosePredicateEvaluation()
	if got := discoveredTargetPaths(catalog.Targets); !reflect.DeepEqual(got, []string{target}) {
		t.Fatalf("targets = %q, want exact native target %q", got, target)
	}
	loader.predicateMu.Lock()
	defer loader.predicateMu.Unlock()
	if len(loader.predicateBatches) != 1 || len(loader.predicateBatches[0].Calls) != 1 ||
		loader.predicateBatches[0].Calls[0].AbsolutePath != target {
		t.Fatalf("predicate batches = %+v, want exact path %q", loader.predicateBatches, target)
	}
}

func TestSearchDiscoveryPreservesPOSIXBackslashInConfigRootAndGitignoreReads(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows uses backslash as a path separator")
	}
	root := filepath.Join(t.TempDir(), `repo\root`)
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	configPath := writeConfigCandidate(t, root, "rslint.config.js")
	target := writeDiscoveryFixture(t, root, "src/index.js", "export {};\n")
	if err := os.WriteFile(filepath.Join(root, ".gitignore"), []byte("other.js\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	loader := newFixtureConfigLoader()
	loader.configs[configPath] = namedConfig("root")
	catalog, err := Build(context.Background(), discoveryTestFS(), loader, ConfigDiscoveryRequest{
		CWD: root, Mode: ConfigDiscoveryAuto, Inputs: []string{"src/index.js"},
		CollectTargets: true, GlobInputPaths: true, ErrorOnUnmatchedPattern: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := discoveredTargetPaths(catalog.Targets); !reflect.DeepEqual(got, []string{target}) {
		t.Fatalf("targets = %q, want exact source %q", got, target)
	}
	if got := requestedConfigPaths(loader); !reflect.DeepEqual(got, []string{configPath}) {
		t.Fatalf("config paths = %q, want exact %q", got, configPath)
	}
}

func TestSearchDiscoveryExistingPOSIXBackslashDirectoryUsesESLintNormalizedUnmatchedPattern(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows uses backslash as a path separator")
	}
	root := t.TempDir()
	configPath := writeConfigCandidate(t, root, "rslint.config.js")
	writeDiscoveryFixture(t, root, `literal\*/index.js`, "export {};\n")
	loader := newFixtureConfigLoader()
	loader.configs[configPath] = namedConfig("root")

	_, err := buildSearchFixtureCatalog(t, root, loader, []string{`literal\*`}, nil)
	var noFiles *NoFilesFoundError
	if !errors.As(err, &noFiles) {
		t.Fatalf("error = %v, want ESLint-style NoFilesFoundError", err)
	}
	if noFiles.Pattern != "literal/*" {
		t.Fatalf("unmatched pattern = %q, want normalized diagnostic spelling", noFiles.Pattern)
	}
}

func TestSearchDiscoveryTraversesDirectDirectorySymlinkButNotNestedSymlink(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink setup requires platform-specific privileges on Windows")
	}
	parent := t.TempDir()
	realRoot := filepath.Join(parent, "real")
	aliasRoot := filepath.Join(parent, "alias")
	outside := filepath.Join(parent, "outside")
	if err := os.MkdirAll(outside, 0o755); err != nil {
		t.Fatal(err)
	}
	configPath := writeConfigCandidate(t, realRoot, "rslint.config.js")
	writeDiscoveryFixture(t, realRoot, "index.js", "export {};\n")
	writeDiscoveryFixture(t, outside, "hidden.js", "export {};\n")
	if err := os.Symlink(outside, filepath.Join(realRoot, "nested")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(realRoot, aliasRoot); err != nil {
		t.Fatal(err)
	}
	loader := newFixtureConfigLoader()
	loader.configs[filepath.Join(aliasRoot, "rslint.config.js")] = namedConfig("root")
	loader.configs[configPath] = namedConfig("root")
	catalog, err := Build(context.Background(), discoveryTestFS(), loader, ConfigDiscoveryRequest{
		CWD: parent, Mode: ConfigDiscoveryAuto, Inputs: []string{"alias"},
		CollectTargets: true, GlobInputPaths: true, ErrorOnUnmatchedPattern: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	paths := discoveredTargetPaths(catalog.Targets)
	want := []string{
		filepath.Join(aliasRoot, "index.js"),
		filepath.Join(aliasRoot, "rslint.config.js"),
	}
	if !reflect.DeepEqual(paths, want) {
		t.Fatalf("targets = %q; direct alias must be traversed without entering nested symlink", paths)
	}
}

func TestSearchDiscoveryDoesNotRealpathRecursiveDirectories(t *testing.T) {
	root := t.TempDir()
	configPath := writeConfigCandidate(t, root, "rslint.config.js")
	writeDiscoveryFixture(t, root, "src/a/index.js", "export {};\n")
	writeDiscoveryFixture(t, root, "src/b/deep/index.js", "export {};\n")
	loader := newFixtureConfigLoader()
	loader.configs[configPath] = namedConfig("root")
	fsys := &realpathCountingFS{FS: discoveryTestFS(), calls: make(map[string]int)}

	catalog, err := Build(context.Background(), fsys, loader, ConfigDiscoveryRequest{
		CWD: root, Mode: ConfigDiscoveryAuto, Inputs: []string{"."},
		CollectTargets: true, GlobInputPaths: true, ErrorOnUnmatchedPattern: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer catalog.ClosePredicateEvaluation()
	calls := fsys.snapshot()
	normalizedRoot := rslintconfig.NormalizeHostPath(root)
	if calls[normalizedRoot] == 0 {
		t.Fatalf("glob search root was not resolved through realpath; all calls=%v", calls)
	}
	if calls[normalizedRoot] > 2 {
		t.Fatalf("root realpath calls grew with recursive targets: %v", calls)
	}
	if len(calls) != 1 {
		t.Fatalf("recursive discovery resolved paths below its one physical root through realpath: %v", calls)
	}
}

func TestSearchDiscoveryFindFilesOrdersExplicitBeforeGlobsAndDeduplicates(t *testing.T) {
	root := t.TempDir()
	configPath := writeConfigCandidate(t, root, "rslint.config.js")
	explicit := writeDiscoveryFixture(t, root, "explicit.js", "export {};\n")
	a := writeDiscoveryFixture(t, root, "src/a.js", "export {};\n")
	b := writeDiscoveryFixture(t, root, "src/b.js", "export {};\n")
	loader := newFixtureConfigLoader()
	loader.configs[configPath] = namedConfig("root")

	catalog, err := buildSearchFixtureCatalog(
		t,
		root,
		loader,
		[]string{"src/*.js", "src/a.js", "explicit.js", "src/**"},
		nil,
	)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	want := []DiscoveredTarget{
		{Path: a, ConfigDirectory: tspath.NormalizePath(root), Explicit: true},
		{Path: explicit, ConfigDirectory: tspath.NormalizePath(root), Explicit: true},
		{Path: b, ConfigDirectory: tspath.NormalizePath(root)},
	}
	got := append([]DiscoveredTarget(nil), catalog.Targets...)
	for index := range got {
		if got[index].MergedConfig == nil {
			t.Fatalf("target %q has no merged config", got[index].Path)
		}
		got[index].MergedConfig = nil
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("targets = %+v, want explicit phase then deduplicated glob phase %+v", catalog.Targets, want)
	}
}

func TestSearchDiscoverySerialAndParallelHaveStableObservableResults(t *testing.T) {
	root := t.TempDir()
	rootConfig := writeConfigCandidate(t, root, "rslint.config.js")
	aConfig := writeConfigCandidate(t, root, "a/rslint.config.js")
	bConfig := writeConfigCandidate(t, root, "b/rslint.config.js")
	rootFile := writeDiscoveryFixture(t, root, "root.ts", "export {};\n")
	aFile := writeDiscoveryFixture(t, root, "a/a.ts", "export {};\n")
	bFile := writeDiscoveryFixture(t, root, "b/b.ts", "export {};\n")

	type observableCatalog struct {
		configs        map[string]rslintconfig.RslintConfig
		targets        []DiscoveredTarget
		explicitInputs []ExplicitInputResult
		stats          ConfigDiscoveryStats
		failures       []ConfigFailure
		activatedPaths []string
		requestedPaths []string
	}
	run := func(singleThreaded bool, failB bool) (observableCatalog, error) {
		loader := newFixtureConfigLoader()
		loader.configs[rootConfig] = namedConfig("root")
		loader.configs[aConfig] = namedConfig("a")
		loader.configs[bConfig] = namedConfig("b")
		if failB {
			loader.failures[bConfig] = ConfigModuleError{Code: "load", Message: "broken b"}
		}
		catalog, err := buildSearchFixtureCatalog(
			t,
			root,
			loader,
			[]string{"b/*.ts", "root.ts", "a/*.ts"},
			func(request *ConfigDiscoveryRequest) { request.SingleThreaded = singleThreaded },
		)
		requested := requestedConfigPaths(loader)
		sort.Strings(requested)
		if err != nil {
			return observableCatalog{requestedPaths: requested}, err
		}
		loader.mu.Lock()
		idToPath := make(map[string]string)
		for _, batch := range loader.batches {
			for _, candidate := range batch.Candidates {
				idToPath[candidate.ID] = candidate.ConfigPath
			}
		}
		var activated []string
		for _, activation := range loader.activations {
			for _, id := range activation.EffectiveConfigIDs {
				activated = append(activated, idToPath[id])
			}
		}
		loader.mu.Unlock()
		targets := append([]DiscoveredTarget(nil), catalog.Targets...)
		for index := range targets {
			if targets[index].MergedConfig == nil {
				t.Fatalf("target %q has no merged config", targets[index].Path)
			}
			targets[index].MergedConfig = nil
		}
		return observableCatalog{
			configs:        catalog.Configs,
			targets:        targets,
			explicitInputs: catalog.ExplicitInputs,
			stats:          catalog.Stats,
			failures:       catalog.Failures,
			activatedPaths: activated,
			requestedPaths: requested,
		}, nil
	}

	serial, err := run(true, false)
	if err != nil {
		t.Fatalf("serial Build: %v", err)
	}
	parallel, err := run(false, false)
	if err != nil {
		t.Fatalf("parallel Build: %v", err)
	}
	if !reflect.DeepEqual(serial, parallel) {
		t.Fatalf("serial and parallel catalogs differ:\nserial=%+v\nparallel=%+v", serial, parallel)
	}
	if got, want := serial.targets, []DiscoveredTarget{
		{Path: rootFile, ConfigDirectory: tspath.NormalizePath(root), Explicit: true},
		{Path: bFile, ConfigDirectory: tspath.GetDirectoryPath(bConfig)},
		{Path: aFile, ConfigDirectory: tspath.GetDirectoryPath(aConfig)},
	}; !reflect.DeepEqual(got, want) {
		t.Fatalf("stable target order = %+v, want %+v", got, want)
	}

	_, serialErr := run(true, true)
	_, parallelErr := run(false, true)
	if serialErr == nil || parallelErr == nil || serialErr.Error() != parallelErr.Error() {
		t.Fatalf("serial error = %v, parallel error = %v", serialErr, parallelErr)
	}
	if !errors.Is(serialErr, ErrAllConfigsFailed) || !errors.Is(parallelErr, ErrAllConfigsFailed) {
		t.Fatalf("errors = (%v, %v), want nearest config failure", serialErr, parallelErr)
	}
}

func TestSearchDiscoveryCoalescesConcurrentOwnerAndCandidateLoads(t *testing.T) {
	root := t.TempDir()
	configPath := writeConfigCandidate(t, root, "rslint.config.js")
	inputs := make([]string, 0, 32)
	for index := range 32 {
		relative := filepath.ToSlash(filepath.Join("src", fmt.Sprintf("file-%02d.ts", index)))
		writeDiscoveryFixture(t, root, relative, "export {};\n")
		inputs = append(inputs, relative)
	}
	delegate := newFixtureConfigLoader()
	delegate.configs[configPath] = namedConfig("root")
	loader := &blockingFixtureConfigLoader{
		delegate: delegate,
		started:  make(chan struct{}),
		release:  make(chan struct{}),
	}
	done := make(chan error, 1)
	go func() {
		_, err := buildSearchFixtureCatalog(t, root, loader, inputs, func(request *ConfigDiscoveryRequest) {
			request.SingleThreaded = false
		})
		done <- err
	}()
	select {
	case <-loader.started:
	case <-time.After(2 * time.Second):
		t.Fatal("config load did not start")
	}
	close(loader.release)
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Build: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("concurrent discovery did not complete")
	}
	if got := loader.loads.Load(); got != 1 {
		t.Fatalf("LoadConfigs calls = %d, want one coalesced load", got)
	}
	if got := requestedConfigPaths(delegate); !reflect.DeepEqual(got, []string{configPath}) {
		t.Fatalf("candidate loads = %v, want config evaluated once", got)
	}
}

func TestSearchDiscoveryCancellationDoesNotHangLoadOrWalk(t *testing.T) {
	t.Run("module load", func(t *testing.T) {
		root := t.TempDir()
		configPath := writeConfigCandidate(t, root, "rslint.config.js")
		writeDiscoveryFixture(t, root, "index.ts", "export {};\n")
		delegate := newFixtureConfigLoader()
		delegate.configs[configPath] = namedConfig("root")
		loader := &blockingFixtureConfigLoader{
			delegate: delegate,
			started:  make(chan struct{}),
			release:  make(chan struct{}),
		}
		ctx, cancel := context.WithCancel(context.Background())
		done := make(chan error, 1)
		go func() {
			_, err := Build(ctx, discoveryTestFS(), loader, ConfigDiscoveryRequest{
				CWD: root, Mode: ConfigDiscoveryAuto, Inputs: []string{"index.ts"},
				CollectTargets: true, GlobInputPaths: true, ErrorOnUnmatchedPattern: true,
			})
			done <- err
		}()
		select {
		case <-loader.started:
		case <-time.After(2 * time.Second):
			t.Fatal("config load did not start")
		}
		cancel()
		select {
		case err := <-done:
			if !errors.Is(err, context.Canceled) {
				t.Fatalf("Build error = %v, want context.Canceled", err)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("canceled config load hung")
		}
	})

	t.Run("directory walk", func(t *testing.T) {
		root := t.TempDir()
		for index := range 64 {
			writeDiscoveryFixture(t, root, filepath.ToSlash(filepath.Join("src", fmt.Sprintf("d-%02d", index), "index.ts")), "export {};\n")
		}
		started := make(chan struct{})
		fsys := &signalingWalkFS{FS: discoveryTestFS(), started: started}
		ctx, cancel := context.WithCancel(context.Background())
		done := make(chan error, 1)
		go func() {
			_, err := Build(ctx, fsys, nil, ConfigDiscoveryRequest{
				CWD: root, Mode: ConfigDiscoveryInline, Inputs: []string{"."},
				CollectTargets: true, GlobInputPaths: true, ErrorOnUnmatchedPattern: true,
				SingleThreaded: true,
			})
			done <- err
		}()
		select {
		case <-started:
		case <-time.After(2 * time.Second):
			t.Fatal("directory walk did not start")
		}
		cancel()
		select {
		case err := <-done:
			if !errors.Is(err, context.Canceled) {
				t.Fatalf("Build error = %v, want context.Canceled", err)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("canceled directory walk hung")
		}
	})
}

func TestSearchDiscoveryParallelOuterMemberRejectsWithoutWaitingForHangingSibling(t *testing.T) {
	root := t.TempDir()
	hangingConfig := writeConfigCandidate(t, root, "hanging/rslint.config.js")
	failingConfig := writeConfigCandidate(t, root, "failing/rslint.config.js")
	writeDiscoveryFixture(t, root, "hanging/file.ts", "export {};\n")
	writeDiscoveryFixture(t, root, "failing/file.ts", "export {};\n")
	for _, test := range []struct {
		name   string
		inputs []string
	}{
		{name: "explicit hangs while glob rejects", inputs: []string{"hanging/file.ts", "failing/*.ts"}},
		{name: "glob hangs while explicit rejects", inputs: []string{"failing/file.ts", "hanging/*.ts"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			loader := &failFastConfigLoader{
				delegate:     newFixtureConfigLoader(),
				hangingPath:  hangingConfig,
				failingPath:  failingConfig,
				hangStarted:  make(chan struct{}),
				hangFinished: make(chan struct{}),
			}
			done := make(chan error, 1)
			go func() {
				_, err := Build(context.Background(), discoveryTestFS(), loader, ConfigDiscoveryRequest{
					CWD: root, Mode: ConfigDiscoveryAuto,
					// The explicit file and globMultiSearch are separate outer
					// Promise.all members. A settled rejection must not wait for the
					// other member's config module forever.
					Inputs:                  test.inputs,
					CollectTargets:          true,
					GlobInputPaths:          true,
					ErrorOnUnmatchedPattern: true,
				})
				done <- err
			}()

			select {
			case err := <-done:
				if !errors.Is(err, ErrAllConfigsFailed) || !strings.Contains(err.Error(), "controlled failure") {
					t.Fatalf("Build error = %v, want failing outer member", err)
				}
			case <-time.After(2 * time.Second):
				t.Fatal("settled config failure waited for a hanging outer member")
			}
			select {
			case <-loader.hangFinished:
			case <-time.After(2 * time.Second):
				t.Fatal("canceled sibling did not unwind after fail-fast return")
			}
		})
	}
}

func discoveredTargetPaths(targets []DiscoveredTarget) []string {
	paths := make([]string, 0, len(targets))
	for _, target := range targets {
		paths = append(paths, target.Path)
	}
	sort.Strings(paths)
	return paths
}

func containsPath(paths []string, target string) bool {
	for _, path := range paths {
		if path == target {
			return true
		}
	}
	return false
}
