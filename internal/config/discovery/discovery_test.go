package discovery

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
)

type fakeConfigModuleLoader struct {
	configs     map[string]rslintconfig.RslintConfig
	failures    map[string]ConfigModuleError
	plugins     map[string][]rslintconfig.EslintPluginEntry
	batches     []ConfigLoadBatchRequest
	activations []ConfigActivationRequest
	pluginsByID map[string][]rslintconfig.EslintPluginEntry
	mutate      func(ConfigLoadBatchRequest, ConfigLoadBatchResponse) ConfigLoadBatchResponse
}

type configDiscoveryRealpathFS struct {
	vfs.FS
	realPaths map[string]string

	mu            sync.Mutex
	realpathCalls map[string]int
}

type configDiscoveryCaseSensitivityFS struct {
	vfs.FS
	caseSensitive bool
}

type configDiscoveryReadSpyFS struct {
	vfs.FS
	mu             sync.Mutex
	gitignoreReads int
}

func (fs *configDiscoveryReadSpyFS) ReadFile(path string) (string, bool) {
	if strings.EqualFold(tspath.GetBaseFileName(path), ".gitignore") {
		fs.mu.Lock()
		fs.gitignoreReads++
		fs.mu.Unlock()
	}
	return fs.FS.ReadFile(path)
}

func (fs *configDiscoveryCaseSensitivityFS) UseCaseSensitiveFileNames() bool {
	return fs.caseSensitive
}

func (fs *configDiscoveryRealpathFS) Realpath(path string) string {
	path = tspath.NormalizePath(path)
	fs.mu.Lock()
	if fs.realpathCalls == nil {
		fs.realpathCalls = make(map[string]int)
	}
	fs.realpathCalls[path]++
	fs.mu.Unlock()
	if realPath := fs.realPaths[path]; realPath != "" {
		return tspath.NormalizePath(realPath)
	}
	return fs.FS.Realpath(path)
}

func (fs *configDiscoveryRealpathFS) realpathCallCount(path string) int {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	return fs.realpathCalls[tspath.NormalizePath(path)]
}

type configDiscoveryConcurrencyFS struct {
	vfs.FS
	watched map[string]struct{}
	release chan struct{}
	once    sync.Once

	mu        sync.Mutex
	active    int
	maxActive int
}

func newConfigDiscoveryConcurrencyFS(fsys vfs.FS, watched ...string) *configDiscoveryConcurrencyFS {
	paths := make(map[string]struct{}, len(watched))
	for _, path := range watched {
		paths[tspath.NormalizePath(path)] = struct{}{}
	}
	return &configDiscoveryConcurrencyFS{
		FS:      fsys,
		watched: paths,
		release: make(chan struct{}),
	}
}

func (fs *configDiscoveryConcurrencyFS) GetAccessibleEntries(path string) vfs.Entries {
	if _, watched := fs.watched[tspath.NormalizePath(path)]; !watched {
		return fs.FS.GetAccessibleEntries(path)
	}

	fs.mu.Lock()
	fs.active++
	if fs.active > fs.maxActive {
		fs.maxActive = fs.active
	}
	if fs.active == len(fs.watched) {
		fs.once.Do(func() { close(fs.release) })
	}
	fs.mu.Unlock()

	select {
	case <-fs.release:
	case <-time.After(2 * time.Second):
	}
	entries := fs.FS.GetAccessibleEntries(path)

	fs.mu.Lock()
	fs.active--
	fs.mu.Unlock()
	return entries
}

func (fs *configDiscoveryConcurrencyFS) peak() int {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	return fs.maxActive
}

func (loader *fakeConfigModuleLoader) LoadConfigs(_ context.Context, request ConfigLoadBatchRequest) (ConfigLoadBatchResponse, error) {
	loader.batches = append(loader.batches, request)
	if loader.pluginsByID == nil {
		loader.pluginsByID = make(map[string][]rslintconfig.EslintPluginEntry)
	}
	response := ConfigLoadBatchResponse{TransactionID: request.TransactionID}
	for _, candidate := range request.Candidates {
		path := tspath.NormalizePath(candidate.ConfigPath)
		if failure, failed := loader.failures[path]; failed {
			failure := failure
			response.Results = append(response.Results, ConfigLoadResult{
				ID:     candidate.ID,
				Status: "failed",
				Error:  &failure,
			})
			continue
		}
		entries := append(rslintconfig.RslintConfig(nil), loader.configs[path]...)
		plugins := cloneEslintPluginEntries(loader.plugins[path])
		loader.pluginsByID[candidate.ID] = plugins
		response.Results = append(response.Results, ConfigLoadResult{
			ID:                candidate.ID,
			Status:            "loaded",
			Entries:           entries,
			SourceFingerprint: "fixture:" + filepath.Base(path),
			EslintPlugins:     plugins,
		})
	}
	if loader.mutate != nil {
		response = loader.mutate(request, response)
	}
	return response, nil
}

func (loader *fakeConfigModuleLoader) ActivateConfigs(_ context.Context, request ConfigActivationRequest) (ConfigActivationResponse, error) {
	loader.activations = append(loader.activations, request)
	sources := make(map[string]configSource, len(request.EffectiveConfigIDs))
	for _, id := range request.EffectiveConfigIDs {
		sources[id] = configSource{EslintPlugins: loader.pluginsByID[id]}
	}
	return ConfigActivationResponse{
		TransactionID:       request.TransactionID,
		EslintPluginEntries: aggregateEffectiveEslintPlugins(sources),
	}, nil
}

func TestConfigDiscoveryPriorityDoesNotConsultGitignore(t *testing.T) {
	t.Run("priority", func(t *testing.T) {
		root := t.TempDir()
		loader := newFixtureConfigLoader()
		for _, name := range AutoJSConfigFileNames {
			path := writeConfigCandidate(t, root, name)
			loader.configs[path] = namedConfig(name)
		}

		catalog := buildFixtureCatalog(t, root, loader, ConfigDiscoveryRequest{CWD: root, ImplicitCWD: true})
		if got := catalog.Configs[tspath.NormalizePath(root)][0].Name; got != "rslint.config.js" {
			t.Fatalf("selected config %q, want highest-priority .js config", got)
		}
		if got := requestedConfigPaths(loader); !reflect.DeepEqual(got, []string{tspath.CombinePaths(root, "rslint.config.js")}) {
			t.Fatalf("unexpected load candidates: %v", got)
		}
	})

	t.Run("gitignored higher priority still wins", func(t *testing.T) {
		root := t.TempDir()
		writeDiscoveryFixture(t, root, ".gitignore", "rslint.config.js\n")
		loader := newFixtureConfigLoader()
		jsPath := writeConfigCandidate(t, root, "rslint.config.js")
		mjsPath := writeConfigCandidate(t, root, "rslint.config.mjs")
		loader.configs[jsPath] = namedConfig("selected")
		loader.configs[mjsPath] = namedConfig("lower-priority")

		catalog := buildFixtureCatalog(t, root, loader, ConfigDiscoveryRequest{CWD: root, ImplicitCWD: true})
		if got := catalog.Configs[tspath.NormalizePath(root)][0].Name; got != "selected" {
			t.Fatalf("selected config %q, want highest-priority .js config", got)
		}
		if got := requestedConfigPaths(loader); !reflect.DeepEqual(got, []string{jsPath}) {
			t.Fatalf("config discovery consulted .gitignore: %v", got)
		}
	})
}

func TestConfigDiscoveryDoesNotReadGitignore(t *testing.T) {
	root := t.TempDir()
	writeDiscoveryFixture(t, root, ".gitignore", "ignored/\nrslint.config.js\n")
	configPath := writeConfigCandidate(t, root, "rslint.config.js")
	writeConfigCandidate(t, root, "ignored/rslint.config.js")
	loader := newFixtureConfigLoader()
	loader.configs[configPath] = namedConfig("root")
	loader.configs[tspath.CombinePaths(root, "ignored/rslint.config.js")] = namedConfig("nested")
	fsys := &configDiscoveryReadSpyFS{FS: discoveryTestFS()}

	_, err := Build(context.Background(), fsys, loader, ConfigDiscoveryRequest{
		CWD:         root,
		ImplicitCWD: true,
	})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	fsys.mu.Lock()
	defer fsys.mu.Unlock()
	if fsys.gitignoreReads != 0 {
		t.Fatalf("config discovery read .gitignore %d times", fsys.gitignoreReads)
	}
}

func TestConfigDiscoveryParentGlobalIgnorePrunesNestedConfig(t *testing.T) {
	root := t.TempDir()
	loader := newFixtureConfigLoader()
	rootConfig := writeConfigCandidate(t, root, "rslint.config.js")
	ignoredConfig := writeConfigCandidate(t, root, "fixtures/deep/rslint.config.js")
	visibleConfig := writeConfigCandidate(t, root, "packages/app/rslint.config.js")
	visibleFile := writeDiscoveryFixture(t, root, "packages/app/index.ts", "export {}\n")
	loader.configs[rootConfig] = rslintconfig.RslintConfig{
		{Ignores: []string{"fixtures/**"}},
		{Name: "root", Rules: rslintconfig.Rules{}},
	}
	loader.configs[ignoredConfig] = namedConfig("ignored")
	loader.configs[visibleConfig] = namedConfig("visible")
	loader.plugins[rootConfig] = []rslintconfig.EslintPluginEntry{{Prefix: "root-plugin", RuleNames: []string{"a"}}}
	loader.plugins[ignoredConfig] = []rslintconfig.EslintPluginEntry{{Prefix: "leak", RuleNames: []string{"bad"}}}
	loader.plugins[visibleConfig] = []rslintconfig.EslintPluginEntry{{Prefix: "visible-plugin", RuleNames: []string{"b"}}}

	catalog := buildFixtureCatalog(t, root, loader, ConfigDiscoveryRequest{
		CWD:                       root,
		Directories:               []string{root},
		LimitDirectoryWalkToFiles: true,
		Files: []DiscoveryFile{
			{Path: filepath.Join(root, "fixtures/deep/index.ts"), Explicit: false},
			{Path: visibleFile, Explicit: false},
		},
	})
	wantDirs := []string{tspath.NormalizePath(root), tspath.CombinePaths(root, "packages/app")}
	if got := catalog.ConfigDirectories(); !reflect.DeepEqual(got, wantDirs) {
		t.Fatalf("effective config directories = %v, want %v", got, wantDirs)
	}
	for _, path := range requestedConfigPaths(loader) {
		if path == ignoredConfig {
			t.Fatalf("parent-global-ignored config was evaluated: %v", requestedConfigPaths(loader))
		}
	}
	if got := pluginPrefixes(catalog.EslintPlugins); !reflect.DeepEqual(got, []string{"root-plugin", "visible-plugin"}) {
		t.Fatalf("effective plugin aggregation leaked a pruned config: %v", got)
	}
	if len(loader.activations) != 1 || !reflect.DeepEqual(loader.activations[0].EffectiveConfigIDs, catalog.EffectiveConfigIDs) {
		t.Fatalf("activation did not receive the final effective IDs: %+v", loader.activations)
	}
}

func TestConfigDiscoveryFileCoverIgnoreKeepsTraversalButFiltersCandidates(t *testing.T) {
	t.Run("candidate remains ignored", func(t *testing.T) {
		root := t.TempDir()
		loader := newFixtureConfigLoader()
		rootConfig := writeConfigCandidate(t, root, "rslint.config.js")
		nestedConfig := writeConfigCandidate(t, root, "packages/app/rslint.config.js")
		writeDiscoveryFixture(t, root, "packages/app/index.ts", "export {}\n")
		loader.configs[rootConfig] = rslintconfig.RslintConfig{
			{Ignores: []string{"packages/**/*"}},
			{Name: "root", Rules: rslintconfig.Rules{}},
		}
		loader.configs[nestedConfig] = namedConfig("must-not-load")

		catalog := buildFixtureCatalog(t, root, loader, ConfigDiscoveryRequest{CWD: root, Directories: []string{root}})
		if got := catalog.ConfigDirectories(); !reflect.DeepEqual(got, []string{tspath.NormalizePath(root)}) {
			t.Fatalf("file-cover ignore config directories = %v", got)
		}
		if got := requestedConfigPaths(loader); !reflect.DeepEqual(got, []string{rootConfig}) {
			t.Fatalf("file-cover ignored candidate was evaluated: %v", got)
		}
		if catalog.Stats.DirectoriesVisited < 3 {
			t.Fatalf("file-cover ignore unexpectedly pruned traversal: %+v", catalog.Stats)
		}
	})

	t.Run("candidate negation makes nested config reachable", func(t *testing.T) {
		root := t.TempDir()
		loader := newFixtureConfigLoader()
		rootConfig := writeConfigCandidate(t, root, "rslint.config.js")
		nestedConfig := writeConfigCandidate(t, root, "packages/app/rslint.config.js")
		loader.configs[rootConfig] = rslintconfig.RslintConfig{
			{Ignores: []string{"packages/**/*", "!packages/app/rslint.config.js"}},
			{Name: "root", Rules: rslintconfig.Rules{}},
		}
		loader.configs[nestedConfig] = namedConfig("nested")

		catalog := buildFixtureCatalog(t, root, loader, ConfigDiscoveryRequest{CWD: root, Directories: []string{root}})
		wantDirs := []string{tspath.NormalizePath(root), tspath.CombinePaths(root, "packages/app")}
		if got := catalog.ConfigDirectories(); !reflect.DeepEqual(got, wantDirs) {
			t.Fatalf("negated candidate config directories = %v, want %v", got, wantDirs)
		}
	})

	t.Run("ignored highest priority falls through", func(t *testing.T) {
		root := t.TempDir()
		loader := newFixtureConfigLoader()
		rootConfig := writeConfigCandidate(t, root, "rslint.config.js")
		ignoredJS := writeConfigCandidate(t, root, "packages/app/rslint.config.js")
		selectedMJS := writeConfigCandidate(t, root, "packages/app/rslint.config.mjs")
		loader.configs[rootConfig] = rslintconfig.RslintConfig{
			{Ignores: []string{"packages/app/rslint.config.js"}},
			{Name: "root", Rules: rslintconfig.Rules{}},
		}
		loader.configs[ignoredJS] = namedConfig("must-not-load")
		loader.configs[selectedMJS] = namedConfig("selected")

		catalog := buildFixtureCatalog(t, root, loader, ConfigDiscoveryRequest{CWD: root, Directories: []string{root}})
		if got := catalog.Configs[tspath.CombinePaths(root, "packages/app")][0].Name; got != "selected" {
			t.Fatalf("selected config = %q, want selected .mjs config", got)
		}
		if slices.Contains(requestedConfigPaths(loader), ignoredJS) {
			t.Fatalf("ignored highest-priority candidate was requested: %v", requestedConfigPaths(loader))
		}
	})
}

func TestConfigDiscoveryAutomaticCandidateWinsSameDirectoryLiteralConflict(t *testing.T) {
	root := t.TempDir()
	rootConfig := writeConfigCandidate(t, root, "rslint.config.js")
	ignoredJS := writeConfigCandidate(t, root, "packages/app/rslint.config.js")
	automaticMJS := writeConfigCandidate(t, root, "packages/app/rslint.config.mjs")
	literal := writeDiscoveryFixture(t, root, "packages/app/index.ts", "export {}\n")
	appDir := tspath.CombinePaths(root, "packages/app")

	for _, test := range []struct {
		name         string
		directory    string
		includesRoot bool
	}{
		{name: "literal activates first", directory: root, includesRoot: true},
		{name: "automatic activates first", directory: appDir},
	} {
		t.Run(test.name, func(t *testing.T) {
			for iteration := range 5 {
				loader := newFixtureConfigLoader()
				loader.configs[rootConfig] = rslintconfig.RslintConfig{
					{Ignores: []string{"packages/app/rslint.config.js"}},
					{Name: "root", Rules: rslintconfig.Rules{}},
				}
				loader.configs[ignoredJS] = namedConfig("literal-js")
				loader.configs[automaticMJS] = namedConfig("automatic-mjs")
				loader.plugins[rootConfig] = []rslintconfig.EslintPluginEntry{{Prefix: "root-plugin"}}
				loader.plugins[ignoredJS] = []rslintconfig.EslintPluginEntry{{Prefix: "literal-js-plugin"}}
				loader.plugins[automaticMJS] = []rslintconfig.EslintPluginEntry{{Prefix: "automatic-mjs-plugin"}}

				catalog := buildFixtureCatalog(t, root, loader, ConfigDiscoveryRequest{
					CWD:         root,
					Directories: []string{test.directory},
					Files:       []DiscoveryFile{{Path: literal, Explicit: true}},
				})
				if got := catalog.Configs[appDir][0].Name; got != "automatic-mjs" {
					t.Fatalf("iteration %d selected config = %q, want automatic .mjs", iteration, got)
				}
				scope := catalog.Scopes[appDir]
				if scope.ExplicitOnly || !reflect.DeepEqual(scope.Files, []string{literal}) {
					t.Fatalf("iteration %d mixed scope = %+v", iteration, scope)
				}
				wantPlugins := []string{"automatic-mjs-plugin"}
				if test.includesRoot {
					wantPlugins = append(wantPlugins, "root-plugin")
				}
				if got := pluginPrefixes(catalog.EslintPlugins); !reflect.DeepEqual(got, wantPlugins) {
					t.Fatalf("iteration %d effective plugins = %v", iteration, got)
				}

				idByPath := make(map[string]string)
				for _, batch := range loader.batches {
					for _, candidate := range batch.Candidates {
						idByPath[candidate.ConfigPath] = candidate.ID
					}
				}
				wantIDs := []string{idByPath[automaticMJS]}
				if test.includesRoot {
					wantIDs = append(wantIDs, idByPath[rootConfig])
				}
				sort.Strings(wantIDs)
				if !reflect.DeepEqual(catalog.EffectiveConfigIDs, wantIDs) {
					t.Fatalf("iteration %d effective IDs = %v, want %v", iteration, catalog.EffectiveConfigIDs, wantIDs)
				}
				if slices.Contains(catalog.EffectiveConfigIDs, idByPath[ignoredJS]) {
					t.Fatalf("iteration %d activated literal-only candidate: %v", iteration, catalog.EffectiveConfigIDs)
				}
			}
		})
	}
}

func TestConfigDiscoveryDirectoryTargetCannotBypassAncestorGlobalIgnore(t *testing.T) {
	root := t.TempDir()
	loader := newFixtureConfigLoader()
	rootConfig := writeConfigCandidate(t, root, "rslint.config.js")
	nestedConfig := writeConfigCandidate(t, root, "ignored/deep/rslint.config.js")
	writeDiscoveryFixture(t, root, "ignored/deep/index.ts", "export {}\n")
	loader.configs[rootConfig] = rslintconfig.RslintConfig{
		{Ignores: []string{"ignored/**"}},
		{Name: "root", Rules: rslintconfig.Rules{}},
	}
	loader.configs[nestedConfig] = namedConfig("must-not-load")

	catalog := buildFixtureCatalog(t, root, loader, ConfigDiscoveryRequest{
		CWD:         root,
		Directories: []string{filepath.Join(root, "ignored/deep")},
	})
	root = tspath.NormalizePath(root)
	if got := catalog.ConfigDirectories(); !reflect.DeepEqual(got, []string{root}) {
		t.Fatalf("ignored directory target owners = %v, want only ancestor root", got)
	}
	if got := requestedConfigPathsByBatch(loader); !reflect.DeepEqual(got, [][]string{{rootConfig}}) {
		t.Fatalf("ignored directory target executed nested config: %v", got)
	}
}

func TestConfigDiscoveryGitignoredDirectoryStillLoadsNearestConfig(t *testing.T) {
	for _, test := range []struct {
		name    string
		request func(root string, target string) ConfigDiscoveryRequest
	}{
		{
			name: "explicit directory root",
			request: func(root string, target string) ConfigDiscoveryRequest {
				return ConfigDiscoveryRequest{CWD: root, Directories: []string{target}}
			},
		},
		{
			name: "implicit cwd root",
			request: func(_ string, target string) ConfigDiscoveryRequest {
				return ConfigDiscoveryRequest{CWD: target, ImplicitCWD: true}
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			loader := newFixtureConfigLoader()
			rootConfig := writeConfigCandidate(t, root, "rslint.config.js")
			ignoredConfig := writeConfigCandidate(t, root, "ignored/rslint.config.js")
			ignoredNestedConfig := writeConfigCandidate(t, root, "ignored/deep/rslint.config.js")
			target := tspath.GetDirectoryPath(writeDiscoveryFixture(t, root, "ignored/deep/index.ts", "export {}\n"))
			writeDiscoveryFixture(t, root, ".gitignore", "ignored/\n")
			loader.configs[rootConfig] = namedConfig("root")
			loader.configs[ignoredConfig] = namedConfig("middle")
			loader.configs[ignoredNestedConfig] = namedConfig("nearest")

			catalog := buildFixtureCatalog(t, root, loader, test.request(root, target))
			nearestDir := tspath.CombinePaths(root, "ignored/deep")
			if got := catalog.ConfigDirectories(); !reflect.DeepEqual(got, []string{nearestDir}) {
				t.Fatalf("gitignored directory owner = %v, want nearest config", got)
			}
			wantRequested := []string{rootConfig, ignoredConfig, ignoredNestedConfig}
			if got := requestedConfigPaths(loader); !reflect.DeepEqual(got, wantRequested) {
				t.Fatalf("gitignored directory requested configs = %v, want %v", got, wantRequested)
			}
			if catalog.Stats.DirectoriesPruned != 0 || catalog.Stats.DirectoriesVisited != 1 {
				t.Fatalf("gitignore unexpectedly pruned config discovery: %+v", catalog.Stats)
			}
			if catalog.Scopes[nearestDir].ExplicitOnly {
				t.Fatal("directory ancestry owner must remain available to automatic targets")
			}
		})
	}
}

func TestConfigDiscoveryDefaultExcludedDirectoryRootOnlyLoadsReachableAncestor(t *testing.T) {
	root := t.TempDir()
	loader := newFixtureConfigLoader()
	rootConfig := writeConfigCandidate(t, root, "rslint.config.js")
	excludedConfig := writeConfigCandidate(t, root, "node_modules/rslint.config.js")
	excludedNestedConfig := writeConfigCandidate(t, root, "node_modules/pkg/rslint.config.js")
	target := tspath.GetDirectoryPath(writeDiscoveryFixture(t, root, "node_modules/pkg/index.ts", "export {}\n"))
	loader.configs[rootConfig] = namedConfig("reachable-root")
	loader.configs[excludedConfig] = namedConfig("must-not-load")
	loader.configs[excludedNestedConfig] = namedConfig("must-not-load")

	catalog := buildFixtureCatalog(t, root, loader, ConfigDiscoveryRequest{
		CWD:         root,
		Directories: []string{target},
	})
	root = tspath.NormalizePath(root)
	if got := catalog.ConfigDirectories(); !reflect.DeepEqual(got, []string{root}) {
		t.Fatalf("default-excluded directory ancestry = %v, want reachable root", got)
	}
	if got := requestedConfigPaths(loader); !reflect.DeepEqual(got, []string{rootConfig}) {
		t.Fatalf("default-excluded directory requested configs = %v, want only %q", got, rootConfig)
	}
	if catalog.Stats.DirectoriesPruned != 1 || catalog.Stats.DirectoriesVisited != 0 {
		t.Fatalf("default-excluded directory should not enter a walk frontier: %+v", catalog.Stats)
	}
}

func TestConfigDiscoveryBoundsExpandedTargetWalkToAncestorTrie(t *testing.T) {
	t.Run("loads only branches that can govern supplied targets", func(t *testing.T) {
		root := t.TempDir()
		rootA := filepath.Join(root, "packages/a")
		rootB := filepath.Join(root, "packages/b")
		rootWithoutTarget := filepath.Join(root, "packages/c")
		loader := newFixtureConfigLoader()
		aConfig := writeConfigCandidate(t, rootA, "rslint.config.js")
		aTargetConfig := writeConfigCandidate(t, rootA, "src/deep/rslint.config.js")
		aUnrelatedConfig := writeConfigCandidate(t, rootA, "src/unused/rslint.config.js")
		bConfig := writeConfigCandidate(t, rootB, "rslint.config.js")
		bUnrelatedConfig := writeConfigCandidate(t, rootB, "unused/rslint.config.js")
		rootWithoutTargetConfig := writeConfigCandidate(t, rootWithoutTarget, "rslint.config.js")
		aTarget := writeDiscoveryFixture(t, rootA, "src/deep/index.ts", "export {}\n")
		bTarget := writeDiscoveryFixture(t, rootB, "index.ts", "export {}\n")
		loader.configs[aConfig] = namedConfig("a")
		loader.configs[aTargetConfig] = namedConfig("a-target")
		loader.configs[bConfig] = namedConfig("b")
		loader.failures[aUnrelatedConfig] = ConfigModuleError{Code: "ERR", Message: "must not load"}
		loader.failures[bUnrelatedConfig] = ConfigModuleError{Code: "ERR", Message: "must not load"}
		loader.failures[rootWithoutTargetConfig] = ConfigModuleError{Code: "ERR", Message: "must not load"}

		catalog := buildFixtureCatalog(t, root, loader, ConfigDiscoveryRequest{
			CWD:                       root,
			Directories:               []string{rootA, rootB, rootWithoutTarget},
			LimitDirectoryWalkToFiles: true,
			Files: []DiscoveryFile{
				{Path: aTarget, Explicit: false},
				{Path: bTarget, Explicit: false},
			},
		})
		wantDirs := []string{
			tspath.NormalizePath(rootA),
			tspath.CombinePaths(rootA, "src/deep"),
			tspath.NormalizePath(rootB),
		}
		sort.Strings(wantDirs)
		if got := catalog.ConfigDirectories(); !reflect.DeepEqual(got, wantDirs) {
			t.Fatalf("target-bound config directories = %v, want %v", got, wantDirs)
		}
		requested := requestedConfigPaths(loader)
		for _, unrelated := range []string{aUnrelatedConfig, bUnrelatedConfig, rootWithoutTargetConfig} {
			if slices.Contains(requested, unrelated) {
				t.Fatalf("unrelated config %q was evaluated: %v", unrelated, requested)
			}
		}
	})

	t.Run("parent global ignore prunes a target branch before its config", func(t *testing.T) {
		root := t.TempDir()
		loader := newFixtureConfigLoader()
		rootConfig := writeConfigCandidate(t, root, "rslint.config.js")
		nestedConfig := writeConfigCandidate(t, root, "target/deep/rslint.config.js")
		target := writeDiscoveryFixture(t, root, "target/deep/index.ts", "export {}\n")
		loader.configs[rootConfig] = rslintconfig.RslintConfig{
			{Ignores: []string{"target/deep/**"}},
			{Name: "root", Rules: rslintconfig.Rules{}},
		}
		loader.failures[nestedConfig] = ConfigModuleError{Code: "ERR", Message: "must not load"}

		catalog := buildFixtureCatalog(t, root, loader, ConfigDiscoveryRequest{
			CWD:                       root,
			Directories:               []string{root},
			LimitDirectoryWalkToFiles: true,
			Files:                     []DiscoveryFile{{Path: target, Explicit: false}},
		})
		if got := catalog.ConfigDirectories(); !reflect.DeepEqual(got, []string{tspath.NormalizePath(root)}) {
			t.Fatalf("parent-ignore catalog = %v", got)
		}
		if got := requestedConfigPaths(loader); !reflect.DeepEqual(got, []string{rootConfig}) {
			t.Fatalf("parent-ignore branch evaluated nested config: %v", got)
		}
	})

	t.Run("gitignore does not prune a target branch before its config", func(t *testing.T) {
		root := t.TempDir()
		writeDiscoveryFixture(t, root, ".gitignore", "target/deep/\n")
		loader := newFixtureConfigLoader()
		rootConfig := writeConfigCandidate(t, root, "rslint.config.js")
		nestedConfig := writeConfigCandidate(t, root, "target/deep/rslint.config.js")
		target := writeDiscoveryFixture(t, root, "target/deep/index.ts", "export {}\n")
		loader.configs[rootConfig] = namedConfig("root")
		loader.configs[nestedConfig] = namedConfig("nested")

		catalog := buildFixtureCatalog(t, root, loader, ConfigDiscoveryRequest{
			CWD:                       root,
			Directories:               []string{root},
			LimitDirectoryWalkToFiles: true,
			Files:                     []DiscoveryFile{{Path: target, Explicit: false}},
		})
		wantDirs := []string{tspath.NormalizePath(root), tspath.CombinePaths(root, "target/deep")}
		if got := catalog.ConfigDirectories(); !reflect.DeepEqual(got, wantDirs) {
			t.Fatalf("gitignore catalog = %v, want %v", got, wantDirs)
		}
		if got := requestedConfigPaths(loader); !reflect.DeepEqual(got, []string{rootConfig, nestedConfig}) {
			t.Fatalf("gitignore must not hide nested config: %v", got)
		}
	})

	t.Run("directory-only requests keep the unbounded walk", func(t *testing.T) {
		root := t.TempDir()
		loader := newFixtureConfigLoader()
		rootConfig := writeConfigCandidate(t, root, "rslint.config.js")
		nestedConfig := writeConfigCandidate(t, root, "unrelated/deep/rslint.config.js")
		loader.configs[rootConfig] = namedConfig("root")
		loader.configs[nestedConfig] = namedConfig("nested")

		catalog := buildFixtureCatalog(t, root, loader, ConfigDiscoveryRequest{
			CWD:                       root,
			Directories:               []string{root},
			LimitDirectoryWalkToFiles: true,
		})
		wantDirs := []string{tspath.NormalizePath(root), tspath.CombinePaths(root, "unrelated/deep")}
		if got := catalog.ConfigDirectories(); !reflect.DeepEqual(got, wantDirs) {
			t.Fatalf("directory-only catalog = %v, want %v", got, wantDirs)
		}
	})

	t.Run("mixed CLI-style files and directories remain unbounded without the API flag", func(t *testing.T) {
		root := t.TempDir()
		loader := newFixtureConfigLoader()
		rootConfig := writeConfigCandidate(t, root, "rslint.config.js")
		nestedConfig := writeConfigCandidate(t, root, "unrelated/deep/rslint.config.js")
		literal := writeDiscoveryFixture(t, root, "index.ts", "export {}\n")
		loader.configs[rootConfig] = namedConfig("root")
		loader.configs[nestedConfig] = namedConfig("nested")

		catalog := buildFixtureCatalog(t, root, loader, ConfigDiscoveryRequest{
			CWD:         root,
			Directories: []string{root},
			Files:       []DiscoveryFile{{Path: literal, Explicit: true}},
		})
		wantDirs := []string{tspath.NormalizePath(root), tspath.CombinePaths(root, "unrelated/deep")}
		if got := catalog.ConfigDirectories(); !reflect.DeepEqual(got, wantDirs) {
			t.Fatalf("mixed CLI-style catalog = %v, want %v", got, wantDirs)
		}
	})
}

func TestConfigDiscoveryExplicitConfigDoesNotConsultGitignore(t *testing.T) {
	root := t.TempDir()
	writeDiscoveryFixture(t, root, ".gitignore", "ignored/\n")
	loader := newFixtureConfigLoader()
	configPath := writeConfigCandidate(t, root, "ignored/custom.config.cjs")
	loader.configs[configPath] = namedConfig("explicit")

	catalog := buildFixtureCatalog(t, root, loader, ConfigDiscoveryRequest{
		CWD:                root,
		Mode:               ConfigDiscoveryExplicit,
		ExplicitConfigPath: configPath,
	})
	root = tspath.NormalizePath(root)
	if got := catalog.ConfigDirectories(); !reflect.DeepEqual(got, []string{root}) {
		t.Fatalf("explicit config should be anchored at cwd: %v", got)
	}
	if !catalog.Explicit {
		t.Fatal("explicit discovery catalog was not marked invocation-wide")
	}
	if got := requestedConfigPaths(loader); !reflect.DeepEqual(got, []string{configPath}) {
		t.Fatalf("explicit source = %v, want %q", got, configPath)
	}
	if got := loader.batches[0].Candidates[0].ConfigDirectory; got != root {
		t.Fatalf("wire configDirectory = %q, want cwd %q", got, root)
	}
}

func TestConfigDiscoveryExplicitFileFindsConfigInsideGitignoredDirectory(t *testing.T) {
	root := t.TempDir()
	writeDiscoveryFixture(t, root, ".gitignore", "ignored/\n")
	loader := newFixtureConfigLoader()
	rootConfig := writeConfigCandidate(t, root, "rslint.config.js")
	ignoredConfig := writeConfigCandidate(t, root, "ignored/rslint.config.js")
	loader.configs[rootConfig] = namedConfig("root")
	loader.configs[ignoredConfig] = namedConfig("nearest")
	filePath := writeDiscoveryFixture(t, root, "ignored/index.ts", "export {}\n")

	catalog := buildFixtureCatalog(t, root, loader, ConfigDiscoveryRequest{
		CWD: root,
		Files: []DiscoveryFile{{
			Path:     filePath,
			Explicit: true,
		}},
	})
	ignoredDir := tspath.CombinePaths(root, "ignored")
	if got := catalog.ConfigDirectories(); !reflect.DeepEqual(got, []string{ignoredDir}) {
		t.Fatalf("gitignored explicit file owner = %v, want nested config", got)
	}
	if got := requestedConfigPaths(loader); !reflect.DeepEqual(got, []string{ignoredConfig}) {
		t.Fatalf("gitignored explicit file requested configs: %v", got)
	}
	scope, exists := catalog.Scopes[ignoredDir]
	if !exists || !scope.ExplicitOnly {
		t.Fatalf("explicit-only source was not propagated to target scope: %+v", scope)
	}
	if !reflect.DeepEqual(scope.Files, []string{filePath}) {
		t.Fatalf("explicit file scope = %+v, want %q", scope, filePath)
	}
}

func TestConfigDiscoveryCanonicalFallbackIgnoresGitignoreButNotDefaultExclusions(t *testing.T) {
	for _, test := range []struct {
		name        string
		lexicalPath string
		gitignore   string
	}{
		{name: "gitignored", lexicalPath: "ignored/index.ts", gitignore: "ignored/\n"},
		{name: "default excluded", lexicalPath: "node_modules/pkg/index.ts"},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			external := t.TempDir()
			if test.gitignore != "" {
				writeDiscoveryFixture(t, root, ".gitignore", test.gitignore)
			}
			lexicalFile := writeDiscoveryFixture(t, root, test.lexicalPath, "export {};\n")
			canonicalFile := writeDiscoveryFixture(t, external, "pkg/index.ts", "export {};\n")
			loader := newFixtureConfigLoader()
			externalConfig := writeConfigCandidate(t, external, "pkg/rslint.config.js")
			loader.configs[externalConfig] = namedConfig("must-not-load")

			catalog := buildFixtureCatalog(t, root, loader, ConfigDiscoveryRequest{
				CWD: root,
				Files: []DiscoveryFile{{
					Path:          lexicalFile,
					CanonicalPath: canonicalFile,
					Explicit:      true,
				}},
			})
			if test.gitignore != "" {
				externalDir := tspath.NormalizePath(filepath.Join(external, "pkg"))
				if got := catalog.ConfigDirectories(); !reflect.DeepEqual(got, []string{externalDir}) {
					t.Fatalf("gitignore blocked canonical fallback: %v", got)
				}
				if got := requestedConfigPaths(loader); !reflect.DeepEqual(got, []string{externalConfig}) {
					t.Fatalf("canonical config requests = %v", got)
				}
			} else {
				if got := catalog.ConfigDirectories(); len(got) != 0 {
					t.Fatalf("default-excluded target escaped into canonical config: %v", got)
				}
				if got := requestedConfigPaths(loader); len(got) != 0 {
					t.Fatalf("default-excluded target evaluated canonical config: %v", got)
				}
			}
		})
	}
}

func TestConfigDiscoveryExplicitFileDoesNotReadGitignoreThroughDescendantSymlink(t *testing.T) {
	root := t.TempDir()
	external := t.TempDir()
	loader := newFixtureConfigLoader()
	rootConfig := writeConfigCandidate(t, root, "rslint.config.js")
	loader.configs[rootConfig] = namedConfig("root")
	writeDiscoveryFixture(t, external, ".gitignore", "index.ts\n")
	writeDiscoveryFixture(t, external, "index.ts", "export {}\n")
	link := filepath.Join(root, "link")
	if err := os.Symlink(external, link); err != nil {
		t.Skipf("directory symlink unavailable: %v", err)
	}
	filePath := tspath.NormalizePath(filepath.Join(link, "index.ts"))

	catalog := buildFixtureCatalog(t, root, loader, ConfigDiscoveryRequest{
		CWD: root,
		Files: []DiscoveryFile{{
			Path:     filePath,
			Explicit: true,
		}},
	})
	root = tspath.NormalizePath(root)
	if got := catalog.ConfigDirectories(); !reflect.DeepEqual(got, []string{root}) {
		t.Fatalf("lexical symlink target owner = %v, want root", got)
	}
	if got := catalog.Scopes[root].Files; !reflect.DeepEqual(got, []string{filePath}) {
		t.Fatalf("explicit target scope = %v, want %q", got, filePath)
	}
}

func TestConfigDiscoveryDefaultExcludedExplicitFileOnlyFindsReachableAncestor(t *testing.T) {
	root := t.TempDir()
	loader := newFixtureConfigLoader()
	rootConfig := writeConfigCandidate(t, root, "rslint.config.js")
	excludedConfig := writeConfigCandidate(t, root, "node_modules/pkg/rslint.config.js")
	loader.configs[rootConfig] = namedConfig("reachable-root")
	loader.configs[excludedConfig] = namedConfig("must-not-load")
	filePath := writeDiscoveryFixture(t, root, "node_modules/pkg/index.ts", "export {}\n")

	catalog := buildFixtureCatalog(t, root, loader, ConfigDiscoveryRequest{
		CWD: root,
		Files: []DiscoveryFile{{
			Path:     filePath,
			Explicit: true,
		}},
	})
	root = tspath.NormalizePath(root)
	if got := catalog.ConfigDirectories(); !reflect.DeepEqual(got, []string{root}) {
		t.Fatalf("default-excluded target owner = %v, want root", got)
	}
	if got := requestedConfigPaths(loader); !reflect.DeepEqual(got, []string{rootConfig}) {
		t.Fatalf("default-excluded config was evaluated: %v", got)
	}
}

func TestConfigDiscoveryBrokenNearestFallsBackToAncestor(t *testing.T) {
	root := t.TempDir()
	loader := newFixtureConfigLoader()
	rootConfig := writeConfigCandidate(t, root, "rslint.config.js")
	brokenConfig := writeConfigCandidate(t, root, "packages/broken/rslint.config.js")
	filePath := writeDiscoveryFixture(t, root, "packages/broken/index.ts", "debugger;\n")
	loader.configs[rootConfig] = namedConfig("root")
	loader.failures[brokenConfig] = ConfigModuleError{Code: "ERR_MODULE", Message: "broken fixture"}

	catalog := buildFixtureCatalog(t, root, loader, ConfigDiscoveryRequest{
		CWD: root,
		Files: []DiscoveryFile{{
			Path:     filePath,
			Explicit: true,
		}},
	})
	root = tspath.NormalizePath(root)
	if got := catalog.ConfigDirectories(); !reflect.DeepEqual(got, []string{root}) {
		t.Fatalf("fallback catalog = %v, want root", got)
	}
	if len(catalog.Failures) != 1 || catalog.Failures[0].Path != brokenConfig {
		t.Fatalf("broken boundary not retained as a failure: %+v", catalog.Failures)
	}
	if got := requestedConfigPathsByBatch(loader); !reflect.DeepEqual(got, [][]string{{brokenConfig}, {rootConfig}}) {
		t.Fatalf("fallback load order = %v", got)
	}
	if scope := catalog.Scopes[root]; !reflect.DeepEqual(scope.Files, []string{filePath}) {
		t.Fatalf("explicit file was not reassigned to ancestor: %+v", scope)
	}
	if !catalog.Scopes[root].ExplicitOnly {
		t.Fatal("explicit-only source was not propagated to target scope")
	}
}

func TestConfigDiscoveryUsesCanonicalOwnerOnlyAfterLexicalAncestryIsEmpty(t *testing.T) {
	t.Run("canonical file fallback", func(t *testing.T) {
		lexicalRoot := t.TempDir()
		physicalRoot := t.TempDir()
		loader := newFixtureConfigLoader()
		configPath := writeConfigCandidate(t, physicalRoot, "rslint.config.js")
		loader.configs[configPath] = namedConfig("physical")
		lexicalFile := writeDiscoveryFixture(t, lexicalRoot, "pkg/index.ts", "export {}\n")
		canonicalFile := writeDiscoveryFixture(t, physicalRoot, "pkg/index.ts", "export {}\n")

		catalog := buildFixtureCatalog(t, lexicalRoot, loader, ConfigDiscoveryRequest{
			CWD: lexicalRoot,
			Files: []DiscoveryFile{{
				Path:          lexicalFile,
				CanonicalPath: canonicalFile,
				Explicit:      true,
			}},
		})
		physicalRoot = tspath.NormalizePath(physicalRoot)
		if got := catalog.ConfigDirectories(); !reflect.DeepEqual(got, []string{physicalRoot}) {
			t.Fatalf("canonical fallback owner = %v, want %q", got, physicalRoot)
		}
		if got := requestedConfigPaths(loader); !reflect.DeepEqual(got, []string{configPath}) {
			t.Fatalf("canonical fallback candidates = %v", got)
		}
		if scope := catalog.Scopes[physicalRoot]; !reflect.DeepEqual(scope.Files, []string{lexicalFile}) {
			t.Fatalf("canonical fallback discarded lexical target: %+v", scope)
		}
	})

	t.Run("lexical candidate wins", func(t *testing.T) {
		lexicalRoot := t.TempDir()
		physicalRoot := t.TempDir()
		loader := newFixtureConfigLoader()
		lexicalConfig := writeConfigCandidate(t, lexicalRoot, "rslint.config.js")
		physicalConfig := writeConfigCandidate(t, physicalRoot, "pkg/rslint.config.js")
		loader.configs[lexicalConfig] = namedConfig("lexical")
		loader.configs[physicalConfig] = namedConfig("physical")
		lexicalFile := writeDiscoveryFixture(t, lexicalRoot, "pkg/index.ts", "export {}\n")
		canonicalFile := writeDiscoveryFixture(t, physicalRoot, "pkg/index.ts", "export {}\n")

		catalog := buildFixtureCatalog(t, lexicalRoot, loader, ConfigDiscoveryRequest{
			CWD: lexicalRoot,
			Files: []DiscoveryFile{{
				Path:          lexicalFile,
				CanonicalPath: canonicalFile,
				Explicit:      true,
			}},
		})
		lexicalRoot = tspath.NormalizePath(lexicalRoot)
		if got := catalog.ConfigDirectories(); !reflect.DeepEqual(got, []string{lexicalRoot}) {
			t.Fatalf("lexical owner = %v, want %q", got, lexicalRoot)
		}
		if got := requestedConfigPaths(loader); !reflect.DeepEqual(got, []string{lexicalConfig}) {
			t.Fatalf("physical fallback ran despite lexical candidate: %v", got)
		}
	})

	t.Run("directory resolves physical root once", func(t *testing.T) {
		lexicalRoot := t.TempDir()
		physicalRoot := t.TempDir()
		loader := newFixtureConfigLoader()
		configPath := writeConfigCandidate(t, physicalRoot, "rslint.config.js")
		targetPath := writeDiscoveryFixture(t, lexicalRoot, "src/index.ts", "export {}\n")
		loader.configs[configPath] = namedConfig("physical-directory")
		fsys := &configDiscoveryRealpathFS{
			FS: discoveryTestFS(),
			realPaths: map[string]string{
				tspath.NormalizePath(lexicalRoot): tspath.NormalizePath(physicalRoot),
			},
		}

		catalog, err := Build(context.Background(), fsys, loader, ConfigDiscoveryRequest{
			CWD:                       lexicalRoot,
			Directories:               []string{lexicalRoot},
			LimitDirectoryWalkToFiles: true,
			Files:                     []DiscoveryFile{{Path: targetPath, Explicit: false}},
		})
		if err != nil {
			t.Fatalf("Build: %v", err)
		}
		physicalRoot = tspath.NormalizePath(physicalRoot)
		if got := catalog.ConfigDirectories(); !reflect.DeepEqual(got, []string{physicalRoot}) {
			t.Fatalf("directory canonical fallback = %v, want %q", got, physicalRoot)
		}
		if got := requestedConfigPaths(loader); !reflect.DeepEqual(got, []string{configPath}) {
			t.Fatalf("directory canonical candidates = %v", got)
		}
		if got := fsys.realpathCallCount(lexicalRoot); got != 1 {
			t.Fatalf("lexical directory realpath calls = %d, want 1", got)
		}
	})

	t.Run("directory symlink preserves lexical config", func(t *testing.T) {
		physicalRoot := t.TempDir()
		aliasParent := t.TempDir()
		aliasRoot := filepath.Join(aliasParent, "workspace-alias")
		if err := os.Symlink(physicalRoot, aliasRoot); err != nil {
			t.Skipf("symlink fixture unavailable: %v", err)
		}
		writeConfigCandidate(t, physicalRoot, "rslint.config.js")
		writeDiscoveryFixture(t, physicalRoot, "src/index.ts", "export {}\n")
		lexicalTarget := tspath.CombinePaths(aliasRoot, "src/index.ts")
		lexicalConfig := tspath.CombinePaths(aliasRoot, "rslint.config.js")
		loader := newFixtureConfigLoader()
		loader.configs[lexicalConfig] = namedConfig("lexical-alias")

		catalog := buildFixtureCatalog(t, aliasRoot, loader, ConfigDiscoveryRequest{
			CWD:                       aliasRoot,
			Directories:               []string{aliasRoot},
			LimitDirectoryWalkToFiles: true,
			Files:                     []DiscoveryFile{{Path: lexicalTarget, Explicit: false}},
		})
		aliasRoot = tspath.NormalizePath(aliasRoot)
		if got := catalog.ConfigDirectories(); !reflect.DeepEqual(got, []string{aliasRoot}) {
			t.Fatalf("symlink directory lost lexical owner: %v", got)
		}
		if got := requestedConfigPaths(loader); !reflect.DeepEqual(got, []string{lexicalConfig}) {
			t.Fatalf("symlink directory did not stay lexical-first: %v", got)
		}
	})
}

func TestConfigDiscoveryValidatesPhysicalConfigDirectoryIdentityBeforeActivation(t *testing.T) {
	t.Run("distinct symlink aliases are rejected", func(t *testing.T) {
		root := t.TempDir()
		physicalRoot := filepath.Join(root, "shared")
		configPath := writeConfigCandidate(t, physicalRoot, "rslint.config.js")
		aliasA := filepath.Join(root, "owner-a")
		aliasB := filepath.Join(root, "owner-b")
		if err := os.Symlink(physicalRoot, aliasA); err != nil {
			t.Skipf("symlink fixture unavailable: %v", err)
		}
		if err := os.Symlink(physicalRoot, aliasB); err != nil {
			t.Skipf("symlink fixture unavailable: %v", err)
		}

		loader := newFixtureConfigLoader()
		aliasAConfig := tspath.CombinePaths(aliasA, filepath.Base(configPath))
		aliasBConfig := tspath.CombinePaths(aliasB, filepath.Base(configPath))
		loader.configs[aliasAConfig] = namedConfig("owner-a")
		loader.configs[aliasBConfig] = namedConfig("owner-b")

		catalog, err := Build(context.Background(), discoveryTestFS(), loader, ConfigDiscoveryRequest{
			CWD:         root,
			Directories: []string{aliasA, aliasB},
		})
		if catalog != nil {
			t.Fatalf("ambiguous catalog = %+v, want nil", catalog)
		}
		if err == nil || !strings.Contains(err.Error(), "Config directories") ||
			!strings.Contains(err.Error(), "resolve to the same filesystem location") {
			t.Fatalf("error = %v, want physical config-directory collision", err)
		}
		if len(loader.activations) != 0 {
			t.Fatalf("ambiguous catalog activated Node state: %+v", loader.activations)
		}
	})

	t.Run("case-only symlink aliases on case-sensitive filesystem are rejected", func(t *testing.T) {
		fsys := discoveryTestFS()
		if !fsys.UseCaseSensitiveFileNames() {
			t.Skip("fixture requires a case-sensitive filesystem")
		}
		root := t.TempDir()
		physicalRoot := filepath.Join(root, "shared")
		configPath := writeConfigCandidate(t, physicalRoot, "rslint.config.js")
		upperAlias := filepath.Join(root, "Owner")
		lowerAlias := filepath.Join(root, "owner")
		if err := os.Symlink(physicalRoot, upperAlias); err != nil {
			t.Skipf("symlink fixture unavailable: %v", err)
		}
		if err := os.Symlink(physicalRoot, lowerAlias); err != nil {
			t.Skipf("case-distinct symlink fixture unavailable: %v", err)
		}
		loader := newFixtureConfigLoader()
		loader.configs[tspath.CombinePaths(upperAlias, filepath.Base(configPath))] = namedConfig("upper")
		loader.configs[tspath.CombinePaths(lowerAlias, filepath.Base(configPath))] = namedConfig("lower")

		catalog, err := Build(context.Background(), fsys, loader, ConfigDiscoveryRequest{
			CWD:         root,
			Directories: []string{upperAlias, lowerAlias},
		})
		if catalog != nil || err == nil || !strings.Contains(err.Error(), "resolve to the same filesystem location") {
			t.Fatalf("case-only physical aliases: catalog=%+v err=%v", catalog, err)
		}
		if len(loader.activations) != 0 {
			t.Fatalf("ambiguous case aliases activated Node state: %+v", loader.activations)
		}
	})

	t.Run("native case aliases are allowed", func(t *testing.T) {
		root := t.TempDir()
		upperRoot := filepath.Join(root, "Project")
		lowerRoot := filepath.Join(root, "project")
		upperConfig := writeConfigCandidate(t, upperRoot, "rslint.config.js")
		lowerConfig := writeConfigCandidate(t, lowerRoot, "rslint.config.js")
		loader := newFixtureConfigLoader()
		loader.configs[upperConfig] = namedConfig("upper")
		loader.configs[lowerConfig] = namedConfig("lower")
		realpathFS := &configDiscoveryRealpathFS{
			FS: discoveryTestFS(),
			realPaths: map[string]string{
				tspath.NormalizePath(upperRoot): tspath.NormalizePath(upperRoot),
				tspath.NormalizePath(lowerRoot): tspath.NormalizePath(upperRoot),
			},
		}
		fsys := &configDiscoveryCaseSensitivityFS{FS: realpathFS, caseSensitive: false}

		catalog, err := Build(context.Background(), fsys, loader, ConfigDiscoveryRequest{
			CWD:         root,
			Directories: []string{upperRoot, lowerRoot},
		})
		if err != nil {
			t.Fatalf("Build: %v", err)
		}
		if len(catalog.Configs) == 0 {
			t.Fatal("native case aliases produced an empty catalog")
		}
		if len(loader.activations) != 1 {
			t.Fatalf("activation count = %d, want 1", len(loader.activations))
		}
	})
}

func TestConfigDiscoverySiblingFrontierIsBatchedDeterministically(t *testing.T) {
	root := t.TempDir()
	loader := newFixtureConfigLoader()
	rootConfig := writeConfigCandidate(t, root, "rslint.config.js")
	aConfig := writeConfigCandidate(t, root, "packages/a/rslint.config.js")
	bConfig := writeConfigCandidate(t, root, "packages/b/rslint.config.js")
	loader.configs[rootConfig] = namedConfig("root")
	loader.configs[aConfig] = namedConfig("a")
	loader.configs[bConfig] = namedConfig("b")

	catalog := buildFixtureCatalog(t, root, loader, ConfigDiscoveryRequest{CWD: root, ImplicitCWD: true})
	if got := requestedConfigPathsByBatch(loader); !reflect.DeepEqual(got, [][]string{{rootConfig}, {aConfig, bConfig}}) {
		t.Fatalf("staged batches = %v", got)
	}
	if !sort.StringsAreSorted(catalog.EffectiveConfigIDs) {
		t.Fatalf("effective IDs are not deterministic: %v", catalog.EffectiveConfigIDs)
	}
	if len(catalog.EffectiveConfigIDs) != 3 {
		t.Fatalf("effective ID count = %d", len(catalog.EffectiveConfigIDs))
	}
}

func TestConfigDiscoveryWalksSiblingFrontierConcurrentlyWithStableOutput(t *testing.T) {
	root := t.TempDir()
	rootConfig := writeConfigCandidate(t, root, "rslint.config.js")
	aDir := filepath.Join(root, "a")
	bDir := filepath.Join(root, "b")
	writeDiscoveryFixture(t, root, "a/index.ts", "export {}\n")
	writeDiscoveryFixture(t, root, "b/index.ts", "export {}\n")

	parallelLoader := newFixtureConfigLoader()
	parallelLoader.configs[rootConfig] = namedConfig("root")
	probeFS := newConfigDiscoveryConcurrencyFS(discoveryTestFS(), aDir, bDir)
	parallelCatalog, err := Build(context.Background(), probeFS, parallelLoader, ConfigDiscoveryRequest{
		CWD:         root,
		ImplicitCWD: true,
	})
	if err != nil {
		t.Fatalf("parallel Build: %v", err)
	}
	if got := probeFS.peak(); got < 2 {
		t.Fatalf("sibling frontier peak concurrency = %d, want at least 2", got)
	}

	serialLoader := newFixtureConfigLoader()
	serialLoader.configs[rootConfig] = namedConfig("root")
	serialCatalog := buildFixtureCatalog(t, root, serialLoader, ConfigDiscoveryRequest{
		CWD:            root,
		ImplicitCWD:    true,
		SingleThreaded: true,
	})
	for _, batch := range parallelLoader.batches {
		if batch.SingleThreaded {
			t.Fatal("parallel config load batch unexpectedly requested serialization")
		}
	}
	for _, batch := range serialLoader.batches {
		if !batch.SingleThreaded {
			t.Fatal("single-threaded config load batch omitted serialization hint")
		}
	}
	if !reflect.DeepEqual(parallelCatalog.ConfigDirectories(), serialCatalog.ConfigDirectories()) ||
		!reflect.DeepEqual(parallelCatalog.Stats, serialCatalog.Stats) ||
		!reflect.DeepEqual(requestedConfigPathsByBatch(parallelLoader), requestedConfigPathsByBatch(serialLoader)) {
		t.Fatalf("parallel and serial discovery diverged: parallel=%+v serial=%+v", parallelCatalog, serialCatalog)
	}
}

func TestConfigDiscoveryEmptyCatalogDoesNotActivateUnknownTransaction(t *testing.T) {
	root := t.TempDir()
	loader := newFixtureConfigLoader()
	catalog := buildFixtureCatalog(t, root, loader, ConfigDiscoveryRequest{CWD: root, ImplicitCWD: true})
	if len(catalog.Configs) != 0 || len(loader.batches) != 0 {
		t.Fatalf("empty discovery produced config work: configs=%v batches=%v", catalog.ConfigDirectories(), loader.batches)
	}
	if len(loader.activations) != 0 {
		t.Fatalf("empty discovery activated an unknown host transaction: %+v", loader.activations)
	}
}

func TestConfigDiscoveryTransactionIDsAreUniqueAcrossBuilds(t *testing.T) {
	root := t.TempDir()
	const builds = 64
	type result struct {
		id  string
		err error
	}
	results := make(chan result, builds)
	var waitGroup sync.WaitGroup
	waitGroup.Add(builds)
	for range builds {
		go func() {
			defer waitGroup.Done()
			catalog, err := Build(context.Background(), discoveryTestFS(), nil, ConfigDiscoveryRequest{CWD: root})
			if err != nil {
				results <- result{err: err}
				return
			}
			results <- result{id: catalog.TransactionID}
		}()
	}
	waitGroup.Wait()
	close(results)

	seen := make(map[string]struct{}, builds)
	for result := range results {
		if result.err != nil {
			t.Fatalf("Build: %v", result.err)
		}
		if result.id == "" {
			t.Fatal("transaction ID is empty")
		}
		if _, duplicate := seen[result.id]; duplicate {
			t.Fatalf("duplicate transaction ID %q", result.id)
		}
		seen[result.id] = struct{}{}
	}
	if len(seen) != builds {
		t.Fatalf("unique transaction IDs = %d, want %d", len(seen), builds)
	}
}

func TestValidateConfigLoadBatchRejectsProtocolViolations(t *testing.T) {
	request := ConfigLoadBatchRequest{
		ProtocolVersion: ConfigDiscoveryProtocolVersion,
		TransactionID:   "txn",
		LoadMode:        ConfigModuleLoadCached,
		Candidates: []ConfigLoadCandidate{
			{ID: "a", ConfigPath: "/a.js", ConfigDirectory: "/"},
			{ID: "b", ConfigPath: "/b.js", ConfigDirectory: "/"},
		},
	}
	loaded := func(id string) ConfigLoadResult {
		return ConfigLoadResult{ID: id, Status: "loaded", Entries: rslintconfig.RslintConfig{}, SourceFingerprint: "fixture:" + id}
	}
	tests := []struct {
		name     string
		response ConfigLoadBatchResponse
	}{
		{name: "transaction", response: ConfigLoadBatchResponse{TransactionID: "other", Results: []ConfigLoadResult{loaded("a"), loaded("b")}}},
		{name: "count", response: ConfigLoadBatchResponse{TransactionID: "txn", Results: []ConfigLoadResult{loaded("a")}}},
		{name: "unknown id", response: ConfigLoadBatchResponse{TransactionID: "txn", Results: []ConfigLoadResult{loaded("a"), loaded("c")}}},
		{name: "duplicate id", response: ConfigLoadBatchResponse{TransactionID: "txn", Results: []ConfigLoadResult{loaded("a"), loaded("a")}}},
		{name: "order", response: ConfigLoadBatchResponse{TransactionID: "txn", Results: []ConfigLoadResult{loaded("b"), loaded("a")}}},
		{name: "status", response: ConfigLoadBatchResponse{TransactionID: "txn", Results: []ConfigLoadResult{{ID: "a", Status: "maybe"}, loaded("b")}}},
		{name: "loaded with error", response: ConfigLoadBatchResponse{TransactionID: "txn", Results: []ConfigLoadResult{{ID: "a", Status: "loaded", Error: &ConfigModuleError{Message: "bad"}}, loaded("b")}}},
		{name: "loaded without fingerprint", response: ConfigLoadBatchResponse{TransactionID: "txn", Results: []ConfigLoadResult{{ID: "a", Status: "loaded"}, loaded("b")}}},
		{name: "failed without error", response: ConfigLoadBatchResponse{TransactionID: "txn", Results: []ConfigLoadResult{{ID: "a", Status: "failed"}, loaded("b")}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := validateConfigLoadBatch(request, test.response)
			var protocolErr *ConfigDiscoveryProtocolError
			if !errors.As(err, &protocolErr) {
				t.Fatalf("error = %v, want ConfigDiscoveryProtocolError", err)
			}
		})
	}
}

func TestConfigDiscoveryAllBrokenDoesNotSilentlyProduceEmptyCatalog(t *testing.T) {
	root := t.TempDir()
	loader := newFixtureConfigLoader()
	configPath := writeConfigCandidate(t, root, "rslint.config.js")
	loader.failures[configPath] = ConfigModuleError{Code: "ERR", Message: "nope"}
	_, err := Build(context.Background(), discoveryTestFS(), loader, ConfigDiscoveryRequest{
		CWD:         root,
		ImplicitCWD: true,
	})
	if !errors.Is(err, ErrAllConfigsFailed) {
		t.Fatalf("error = %v, want ErrAllConfigsFailed", err)
	}
	if !strings.Contains(err.Error(), "failed to load config") {
		t.Fatalf("error = %v, want load-stage context", err)
	}

	invalidLoader := newFixtureConfigLoader()
	invalidLoader.failures[configPath] = ConfigModuleError{Code: "invalid", Message: "not an array"}
	_, err = Build(context.Background(), discoveryTestFS(), invalidLoader, ConfigDiscoveryRequest{
		CWD:         root,
		ImplicitCWD: true,
	})
	if !errors.Is(err, ErrAllConfigsFailed) || !strings.Contains(err.Error(), "invalid config in") {
		t.Fatalf("invalid error = %v, want ErrAllConfigsFailed with invalid-stage context", err)
	}
}

func newFixtureConfigLoader() *fakeConfigModuleLoader {
	return &fakeConfigModuleLoader{
		configs:  make(map[string]rslintconfig.RslintConfig),
		failures: make(map[string]ConfigModuleError),
		plugins:  make(map[string][]rslintconfig.EslintPluginEntry),
	}
}

func buildFixtureCatalog(t *testing.T, root string, loader ConfigModuleLoader, request ConfigDiscoveryRequest) *ConfigCatalog {
	t.Helper()
	if request.CWD == "" {
		request.CWD = root
	}
	catalog, err := Build(context.Background(), discoveryTestFS(), loader, request)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	return catalog
}

func discoveryTestFS() vfs.FS {
	return bundled.WrapFS(cachedvfs.From(osvfs.FS()))
}

func namedConfig(name string) rslintconfig.RslintConfig {
	return rslintconfig.RslintConfig{{Name: name, Rules: rslintconfig.Rules{}}}
}

func writeConfigCandidate(t *testing.T, root string, relativePath string) string {
	t.Helper()
	return writeDiscoveryFixture(t, root, relativePath, "export default [];\n")
}

func writeDiscoveryFixture(t *testing.T, root string, relativePath string, content string) string {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(relativePath))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir fixture: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	return tspath.NormalizePath(path)
}

func requestedConfigPaths(loader *fakeConfigModuleLoader) []string {
	var paths []string
	for _, batch := range loader.batches {
		for _, candidate := range batch.Candidates {
			paths = append(paths, candidate.ConfigPath)
		}
	}
	return paths
}

func requestedConfigPathsByBatch(loader *fakeConfigModuleLoader) [][]string {
	batches := make([][]string, len(loader.batches))
	for index, batch := range loader.batches {
		for _, candidate := range batch.Candidates {
			batches[index] = append(batches[index], candidate.ConfigPath)
		}
	}
	return batches
}

func pluginPrefixes(entries []rslintconfig.EslintPluginEntry) []string {
	prefixes := make([]string, 0, len(entries))
	for _, entry := range entries {
		prefixes = append(prefixes, entry.Prefix)
	}
	sort.Strings(prefixes)
	return prefixes
}
