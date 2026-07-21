package discovery

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"sort"
	"strconv"
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

type configDiscoveryNoSymlinkMetadataFS struct {
	vfs.FS
}

type configDiscoveryReadSpyFS struct {
	vfs.FS
	mu             sync.Mutex
	gitignoreReads int
	gitignorePaths []string
}

type configDiscoveryCandidateFS struct {
	vfs.FS
	files map[string]struct{}
}

func (fs *configDiscoveryCandidateFS) FileExists(path string) bool {
	_, exists := fs.files[tspath.NormalizePath(path)]
	return exists
}

func (fs *configDiscoveryReadSpyFS) ReadFile(path string) (string, bool) {
	if strings.EqualFold(tspath.GetBaseFileName(path), ".gitignore") {
		fs.mu.Lock()
		fs.gitignoreReads++
		fs.gitignorePaths = append(fs.gitignorePaths, tspath.NormalizePath(path))
		fs.mu.Unlock()
	}
	return fs.FS.ReadFile(path)
}

func (fs *configDiscoveryCaseSensitivityFS) UseCaseSensitiveFileNames() bool {
	return fs.caseSensitive
}

func (fs *configDiscoveryNoSymlinkMetadataFS) GetAccessibleEntries(path string) vfs.Entries {
	entries := fs.FS.GetAccessibleEntries(path)
	entries.Symlinks = nil
	return entries
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
			ID:      candidate.ID,
			Status:  "loaded",
			Entries: entries,
		})
	}
	if loader.mutate != nil {
		response = loader.mutate(request, response)
	}
	return response, nil
}

func (loader *fakeConfigModuleLoader) ActivateConfigs(_ context.Context, request ConfigActivationRequest) (ConfigActivationResponse, error) {
	loader.activations = append(loader.activations, request)
	return ConfigActivationResponse{
		TransactionID:       request.TransactionID,
		EslintPluginEntries: fixtureActivationPlugins(loader.pluginsByID, request.EffectiveConfigIDs),
	}, nil
}

func fixtureActivationPlugins(
	pluginsByID map[string][]rslintconfig.EslintPluginEntry,
	effectiveIDs []string,
) []rslintconfig.EslintPluginEntry {
	byPrefix := make(map[string]map[string]struct{})
	for _, id := range effectiveIDs {
		for _, plugin := range pluginsByID[id] {
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
		if got := namedConfigEntry(catalog.Configs[tspath.NormalizePath(root)]).Name; got != "selected" {
			t.Fatalf("selected config %q, want highest-priority .js config", got)
		}
		if got := requestedConfigPaths(loader); !reflect.DeepEqual(got, []string{jsPath}) {
			t.Fatalf("config discovery consulted .gitignore: %v", got)
		}
	})
}

func TestConfigDiscoveryReadsGitignoreAndPrunesHiddenConfig(t *testing.T) {
	root := t.TempDir()
	writeDiscoveryFixture(t, root, ".gitignore", "ignored/\nrslint.config.js\n")
	configPath := writeConfigCandidate(t, root, "rslint.config.js")
	writeConfigCandidate(t, root, "ignored/rslint.config.js")
	loader := newFixtureConfigLoader()
	loader.configs[configPath] = namedConfig("root")
	loader.configs[tspath.CombinePaths(root, "ignored/rslint.config.js")] = namedConfig("nested")
	fsys := &configDiscoveryReadSpyFS{FS: discoveryTestFS()}

	catalog, err := DiscoverAutomatic(context.Background(), fsys, loader, ConfigDiscoveryRequest{
		CWD:         root,
		ImplicitCWD: true,
	})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if got := catalog.ConfigDirectories(); !reflect.DeepEqual(got, []string{tspath.NormalizePath(root)}) {
		t.Fatalf("Git-hidden config directories = %v", got)
	}
	if got := requestedConfigPaths(loader); !reflect.DeepEqual(got, []string{configPath}) {
		t.Fatalf("Git-hidden config was evaluated: %v", got)
	}
	fsys.mu.Lock()
	reads := fsys.gitignoreReads
	fsys.mu.Unlock()
	if reads == 0 {
		t.Fatal("config discovery did not read the owner .gitignore")
	}
}

func TestConfigDiscoveryGitignoreDirectoryNegation(t *testing.T) {
	for _, test := range []struct {
		name       string
		gitignore  string
		wantNested bool
	}{
		{
			name:       "exact directory negation reopens",
			gitignore:  "ignored/\n!ignored/\n",
			wantNested: true,
		},
		{
			name:       "file negation does not reopen",
			gitignore:  "ignored/\n!ignored/keep.ts\n",
			wantNested: false,
		},
		{
			name:       "descendant negation does not reopen",
			gitignore:  "ignored/\n!ignored/**/*\n",
			wantNested: false,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			rootConfig := writeConfigCandidate(t, root, "rslint.config.js")
			nestedConfig := writeConfigCandidate(t, root, "ignored/rslint.config.js")
			writeDiscoveryFixture(t, root, ".gitignore", test.gitignore)
			loader := newFixtureConfigLoader()
			loader.configs[rootConfig] = namedConfig("root")
			loader.configs[nestedConfig] = namedConfig("nested")

			catalog := buildFixtureCatalog(t, root, loader, ConfigDiscoveryRequest{
				CWD:         root,
				ImplicitCWD: true,
			})
			wantDirectories := []string{tspath.NormalizePath(root)}
			wantRequests := []string{rootConfig}
			if test.wantNested {
				wantDirectories = append(wantDirectories, tspath.CombinePaths(root, "ignored"))
				wantRequests = append(wantRequests, nestedConfig)
			}
			if got := catalog.ConfigDirectories(); !reflect.DeepEqual(got, wantDirectories) {
				t.Fatalf("config directories = %v, want %v", got, wantDirectories)
			}
			if got := requestedConfigPaths(loader); !reflect.DeepEqual(got, wantRequests) {
				t.Fatalf("requested configs = %v, want %v", got, wantRequests)
			}
		})
	}
}

func TestConfigDiscoveryAuthoredNegationReopensMatchingGitDirectory(t *testing.T) {
	for _, test := range []struct {
		name       string
		negation   string
		wantNested bool
	}{
		{name: "bare directory node", negation: "!ignored", wantNested: true},
		{name: "directory node", negation: "!ignored/", wantNested: true},
		{name: "directory and contents", negation: "!ignored/**", wantNested: true},
		{name: "contents only", negation: "!ignored/**/*", wantNested: false},
		{name: "one file", negation: "!ignored/keep.ts", wantNested: false},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			rootConfig := writeConfigCandidate(t, root, "rslint.config.js")
			nestedConfig := writeConfigCandidate(t, root, "ignored/rslint.config.js")
			writeDiscoveryFixture(t, root, ".gitignore", "ignored/\n")
			loader := newFixtureConfigLoader()
			loader.configs[rootConfig] = rslintconfig.RslintConfig{
				{Ignores: []string{test.negation}},
				{Name: "root", Rules: rslintconfig.Rules{}},
			}
			loader.configs[nestedConfig] = namedConfig("nested")

			catalog := buildFixtureCatalog(t, root, loader, ConfigDiscoveryRequest{
				CWD:         root,
				ImplicitCWD: true,
			})
			gotNested := slices.Contains(
				catalog.ConfigDirectories(),
				tspath.CombinePaths(root, "ignored"),
			)
			if gotNested != test.wantNested {
				t.Fatalf("nested config loaded = %v, want %v", gotNested, test.wantNested)
			}
		})
	}
}

func TestConfigDiscoveryGitignoreDoesNotFilterCandidateFilename(t *testing.T) {
	root := t.TempDir()
	rootConfig := writeConfigCandidate(t, root, "rslint.config.js")
	jsConfig := writeConfigCandidate(t, root, "packages/app/rslint.config.js")
	mjsConfig := writeConfigCandidate(t, root, "packages/app/rslint.config.mjs")
	writeDiscoveryFixture(t, root, ".gitignore", "packages/app/rslint.config.js\n")
	loader := newFixtureConfigLoader()
	loader.configs[rootConfig] = namedConfig("root")
	loader.configs[jsConfig] = namedConfig("selected-js")
	loader.configs[mjsConfig] = namedConfig("must-not-fall-through")

	catalog := buildFixtureCatalog(t, root, loader, ConfigDiscoveryRequest{
		CWD:         root,
		ImplicitCWD: true,
	})
	appDirectory := tspath.CombinePaths(root, "packages/app")
	if got := namedConfigEntry(catalog.Configs[appDirectory]).Name; got != "selected-js" {
		t.Fatalf("selected config = %q, want highest-priority JS candidate", got)
	}
	if got := requestedConfigPaths(loader); !reflect.DeepEqual(got, []string{rootConfig, jsConfig}) {
		t.Fatalf("Git filtered a config filename or changed priority: %v", got)
	}
}

func TestConfigDiscoveryGitignoreOwnerBoundary(t *testing.T) {
	t.Run("successful child resets inherited Git and applies its own source", func(t *testing.T) {
		root := t.TempDir()
		rootConfig := writeConfigCandidate(t, root, "rslint.config.js")
		childConfig := writeConfigCandidate(t, root, "packages/app/rslint.config.js")
		reachableConfig := writeConfigCandidate(t, root, "packages/app/parent-only/rslint.config.js")
		hiddenConfig := writeConfigCandidate(t, root, "packages/app/child-only/rslint.config.js")
		writeDiscoveryFixture(t, root, ".gitignore", "packages/app/parent-only/\n")
		writeDiscoveryFixture(t, root, "packages/app/.gitignore", "child-only/\n")
		loader := newFixtureConfigLoader()
		loader.configs[rootConfig] = namedConfig("root")
		loader.configs[childConfig] = namedConfig("child")
		loader.configs[reachableConfig] = namedConfig("reachable-after-reset")
		loader.configs[hiddenConfig] = namedConfig("must-not-load")

		catalog := buildFixtureCatalog(t, root, loader, ConfigDiscoveryRequest{
			CWD:         root,
			ImplicitCWD: true,
		})
		wantDirectories := []string{
			tspath.NormalizePath(root),
			tspath.CombinePaths(root, "packages/app"),
			tspath.CombinePaths(root, "packages/app/parent-only"),
		}
		if got := catalog.ConfigDirectories(); !reflect.DeepEqual(got, wantDirectories) {
			t.Fatalf("owner-boundary directories = %v, want %v", got, wantDirectories)
		}
		if got := requestedConfigPaths(loader); !reflect.DeepEqual(got, []string{rootConfig, childConfig, reachableConfig}) {
			t.Fatalf("owner-boundary requests = %v", got)
		}
	})

	t.Run("failed child inherits parent Git", func(t *testing.T) {
		root := t.TempDir()
		rootConfig := writeConfigCandidate(t, root, "rslint.config.js")
		failedConfig := writeConfigCandidate(t, root, "packages/app/rslint.config.js")
		hiddenConfig := writeConfigCandidate(t, root, "packages/app/deep/rslint.config.js")
		writeDiscoveryFixture(t, root, "packages/app/.gitignore", "deep/\n")
		loader := newFixtureConfigLoader()
		loader.configs[rootConfig] = namedConfig("root")
		loader.failures[failedConfig] = ConfigModuleError{Code: "ERR", Message: "broken"}
		loader.configs[hiddenConfig] = namedConfig("must-not-load")

		catalog := buildFixtureCatalog(t, root, loader, ConfigDiscoveryRequest{
			CWD:         root,
			ImplicitCWD: true,
		})
		if got := catalog.ConfigDirectories(); !reflect.DeepEqual(got, []string{tspath.NormalizePath(root)}) {
			t.Fatalf("failed-boundary directories = %v", got)
		}
		if got := requestedConfigPaths(loader); !reflect.DeepEqual(got, []string{rootConfig, failedConfig}) {
			t.Fatalf("failed child did not inherit parent Git: %v", got)
		}
	})
}

func TestConfigDiscoveryDoesNotReadGitignoreBelowBlockedDirectory(t *testing.T) {
	root := t.TempDir()
	rootConfig := writeConfigCandidate(t, root, "rslint.config.js")
	hiddenConfig := writeConfigCandidate(t, root, "ignored/deep/rslint.config.js")
	rootGitignore := writeDiscoveryFixture(t, root, ".gitignore", "ignored/\n")
	hiddenGitignore := writeDiscoveryFixture(t, root, "ignored/.gitignore", "!deep/\n")
	loader := newFixtureConfigLoader()
	loader.configs[rootConfig] = namedConfig("root")
	loader.configs[hiddenConfig] = namedConfig("must-not-load")
	fsys := &configDiscoveryReadSpyFS{FS: discoveryTestFS()}

	_, err := DiscoverAutomatic(context.Background(), fsys, loader, ConfigDiscoveryRequest{
		CWD:         root,
		ImplicitCWD: true,
	})
	if err != nil {
		t.Fatalf("DiscoverAutomatic: %v", err)
	}
	fsys.mu.Lock()
	paths := append([]string(nil), fsys.gitignorePaths...)
	fsys.mu.Unlock()
	if !slices.Contains(paths, rootGitignore) {
		t.Fatalf("root .gitignore was not read: %v", paths)
	}
	if slices.Contains(paths, hiddenGitignore) {
		t.Fatalf("blocked nested .gitignore was read: %v", paths)
	}
	if got := requestedConfigPaths(loader); !reflect.DeepEqual(got, []string{rootConfig}) {
		t.Fatalf("nested source revived a hidden config: %v", got)
	}
}

func TestConfigDiscoveryOverlappingRootsPreserveExplicitOrigin(t *testing.T) {
	root := t.TempDir()
	ignoredDirectory := tspath.CombinePaths(root, "ignored")
	rootConfig := writeConfigCandidate(t, root, "rslint.config.js")
	childConfig := writeConfigCandidate(t, root, "ignored/rslint.config.js")
	writeDiscoveryFixture(t, root, ".gitignore", "ignored/\n")
	loader := newFixtureConfigLoader()
	loader.configs[rootConfig] = namedConfig("root")
	loader.configs[childConfig] = namedConfig("child")

	catalog := buildFixtureCatalog(t, root, loader, ConfigDiscoveryRequest{
		CWD:         root,
		Directories: []string{root, ignoredDirectory},
	})
	if got := catalog.ConfigDirectories(); !reflect.DeepEqual(got, []string{tspath.NormalizePath(root), ignoredDirectory}) {
		t.Fatalf("overlapping-root config directories = %v", got)
	}
	if got := requestedConfigPaths(loader); !reflect.DeepEqual(got, []string{rootConfig, childConfig}) {
		t.Fatalf("overlapping roots executed duplicate or hidden configs: %v", got)
	}
}

func TestConfigDiscoveryCatalogFreezesCollectedGitignore(t *testing.T) {
	root := t.TempDir()
	rootConfig := writeConfigCandidate(t, root, "rslint.config.js")
	target := writeDiscoveryFixture(t, root, "dist/index.ts", "export {}\n")
	reincludedTarget := writeDiscoveryFixture(t, root, "dist/keep.ts", "export {}\n")
	writeDiscoveryFixture(t, root, ".gitignore", "dist/\n")
	loader := newFixtureConfigLoader()
	loader.configs[rootConfig] = rslintconfig.RslintConfig{
		{Ignores: []string{"!dist/keep.ts"}},
		{Name: "root", Rules: rslintconfig.Rules{}},
	}

	catalog := buildFixtureCatalog(t, root, loader, ConfigDiscoveryRequest{
		CWD:         root,
		ImplicitCWD: true,
	})
	root = tspath.NormalizePath(root)
	if !catalog.Configs[root].IsFileIgnored(target, root) {
		t.Fatal("catalog did not materialize the discovered Git source")
	}
	if catalog.Configs[root].IsFileIgnored(reincludedTarget, root) {
		t.Fatal("authored negation did not follow and override the Git projection")
	}
	writeDiscoveryFixture(t, root, ".gitignore", "")
	if !catalog.Configs[root].IsFileIgnored(target, root) {
		t.Fatal("published catalog changed after .gitignore mutation")
	}
	if catalog.Configs[root].IsFileIgnored(reincludedTarget, root) {
		t.Fatal("published Git-to-authored ordering changed after source mutation")
	}
}

func TestConfigDiscoveryMergesLiteralGitignoreChainIntoAutomaticOwner(t *testing.T) {
	root := t.TempDir()
	rootConfig := writeConfigCandidate(t, root, "rslint.config.js")
	automaticTarget := writeDiscoveryFixture(t, root, "automatic/index.ts", "export {}\n")
	reincludedLiteral := writeDiscoveryFixture(t, root, "outside/keep.tmp", "export {}\n")
	ignoredLiteral := writeDiscoveryFixture(t, root, "outside/drop.tmp", "export {}\n")
	writeDiscoveryFixture(t, root, ".gitignore", "*.tmp\n")
	writeDiscoveryFixture(t, root, "outside/.gitignore", "!keep.tmp\n")
	loader := newFixtureConfigLoader()
	loader.configs[rootConfig] = namedConfig("root")

	catalog := buildFixtureCatalog(t, root, loader, ConfigDiscoveryRequest{
		CWD:         root,
		Directories: []string{tspath.GetDirectoryPath(automaticTarget)},
		Files: []DiscoveryFile{
			{Path: automaticTarget, Explicit: false},
			{Path: reincludedLiteral, Explicit: true},
			{Path: ignoredLiteral, Explicit: true},
		},
	})
	root = tspath.NormalizePath(root)
	effective := catalog.Configs[root]
	if effective.IsFileIgnored(reincludedLiteral, root) {
		t.Fatal("nested literal-chain negation was lost or parent Git source was duplicated")
	}
	if !effective.IsFileIgnored(ignoredLiteral, root) {
		t.Fatal("automatic owner did not include the literal target's Git chain")
	}
}

func TestConfigDiscoveryCachesLiteralGitignoreSourceChains(t *testing.T) {
	root := t.TempDir()
	rootConfig := writeConfigCandidate(t, root, "rslint.config.js")
	writeDiscoveryFixture(t, root, ".gitignore", "*.tmp\n")
	writeDiscoveryFixture(t, root, "outside/.gitignore", "!keep.tmp\n")
	loader := newFixtureConfigLoader()
	loader.configs[rootConfig] = namedConfig("root")
	var files []DiscoveryFile
	for _, name := range []string{"keep.tmp", "drop.tmp", "nested/one.tmp", "nested/two.tmp"} {
		files = append(files, DiscoveryFile{
			Path:     writeDiscoveryFixture(t, root, "outside/"+name, "export {}\n"),
			Explicit: true,
		})
	}
	fsys := &configDiscoveryReadSpyFS{FS: discoveryTestFS()}

	_, err := DiscoverAutomatic(context.Background(), fsys, loader, ConfigDiscoveryRequest{
		CWD:   root,
		Files: files,
	})
	if err != nil {
		t.Fatalf("DiscoverAutomatic: %v", err)
	}
	fsys.mu.Lock()
	reads := append([]string(nil), fsys.gitignorePaths...)
	fsys.mu.Unlock()
	counts := make(map[string]int)
	for _, path := range reads {
		counts[path]++
	}
	for _, path := range []string{
		tspath.CombinePaths(root, ".gitignore"),
		tspath.CombinePaths(root, "outside/.gitignore"),
		tspath.CombinePaths(root, "outside/nested/.gitignore"),
	} {
		if got := counts[path]; got != 1 {
			t.Fatalf(".gitignore reads for %q = %d, want 1; all reads: %v", path, got, reads)
		}
	}
}

func TestConfigDiscoveryFreezesGitignoreBytesAcrossOverlappingRoutes(t *testing.T) {
	root := t.TempDir()
	directDirectory := tspath.CombinePaths(root, "direct")
	rootConfig := writeConfigCandidate(t, root, "rslint.config.js")
	directConfig := writeConfigCandidate(t, root, "direct/rslint.config.js")
	hiddenConfig := writeConfigCandidate(t, root, "hidden/rslint.config.js")
	hiddenTarget := writeDiscoveryFixture(t, root, "hidden/index.ts", "export {}\n")
	gitignorePath := writeDiscoveryFixture(t, root, ".gitignore", "hidden/\n")
	loader := newFixtureConfigLoader()
	loader.configs[rootConfig] = namedConfig("root")
	loader.configs[directConfig] = namedConfig("direct")
	loader.configs[hiddenConfig] = namedConfig("must-not-load")
	mutated := false
	loader.mutate = func(request ConfigLoadBatchRequest, response ConfigLoadBatchResponse) ConfigLoadBatchResponse {
		if mutated {
			return response
		}
		for _, candidate := range request.Candidates {
			if candidate.ConfigPath == directConfig {
				if err := os.WriteFile(filepath.FromSlash(gitignorePath), nil, 0o644); err != nil {
					t.Fatalf("mutate .gitignore: %v", err)
				}
				mutated = true
				break
			}
		}
		return response
	}
	fsys := &configDiscoveryReadSpyFS{FS: discoveryTestFS()}

	catalog, err := DiscoverAutomatic(context.Background(), fsys, loader, ConfigDiscoveryRequest{
		CWD:         root,
		Directories: []string{root, directDirectory},
	})
	if err != nil {
		t.Fatalf("DiscoverAutomatic: %v", err)
	}
	if !mutated {
		t.Fatal("fixture did not mutate .gitignore during config evaluation")
	}
	if got := catalog.ConfigDirectories(); !reflect.DeepEqual(got, []string{
		tspath.NormalizePath(root),
		directDirectory,
	}) {
		t.Fatalf("catalog used post-mutation Git bytes: %v", got)
	}
	if got := requestedConfigPaths(loader); !reflect.DeepEqual(got, []string{rootConfig, directConfig}) {
		t.Fatalf("post-mutation route loaded a hidden config: %v", got)
	}
	if !catalog.Configs[tspath.NormalizePath(root)].IsFileIgnored(hiddenTarget, root) {
		t.Fatal("published Git projection differs from discovery reachability")
	}
	fsys.mu.Lock()
	reads := slices.Clone(fsys.gitignorePaths)
	fsys.mu.Unlock()
	count := 0
	for _, path := range reads {
		if path == gitignorePath {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("root .gitignore reads = %d, want one frozen read; all reads: %v", count, reads)
	}
}

func TestConfigDiscoverySkipsDescendantSymlinkWithoutMetadata(t *testing.T) {
	root := t.TempDir()
	external := t.TempDir()
	rootConfig := writeConfigCandidate(t, root, "rslint.config.js")
	writeDiscoveryFixture(t, external, ".gitignore", "deep/\n")
	writeConfigCandidate(t, external, "rslint.config.js")
	alias := filepath.Join(root, "linked")
	if err := os.Symlink(external, alias); err != nil {
		t.Skipf("symlink fixture unavailable: %v", err)
	}
	aliasConfig := tspath.CombinePaths(alias, "rslint.config.js")
	aliasGitignore := tspath.CombinePaths(alias, ".gitignore")
	loader := newFixtureConfigLoader()
	loader.configs[rootConfig] = namedConfig("root")
	loader.configs[aliasConfig] = namedConfig("must-not-load")
	fsys := &configDiscoveryReadSpyFS{
		FS: &configDiscoveryNoSymlinkMetadataFS{FS: discoveryTestFS()},
	}

	catalog, err := DiscoverAutomatic(context.Background(), fsys, loader, ConfigDiscoveryRequest{
		CWD:         root,
		ImplicitCWD: true,
	})
	if err != nil {
		t.Fatalf("DiscoverAutomatic: %v", err)
	}
	if got := catalog.ConfigDirectories(); !reflect.DeepEqual(got, []string{tspath.NormalizePath(root)}) {
		t.Fatalf("descendant symlink config was loaded: %v", got)
	}
	if got := requestedConfigPaths(loader); !reflect.DeepEqual(got, []string{rootConfig}) {
		t.Fatalf("descendant symlink candidate was evaluated: %v", got)
	}
	fsys.mu.Lock()
	reads := slices.Clone(fsys.gitignorePaths)
	fsys.mu.Unlock()
	if slices.Contains(reads, aliasGitignore) {
		t.Fatalf("descendant symlink .gitignore was read: %v", reads)
	}
}

func TestConfigDiscoveryDirectorySymlinkRootPreservesGitSourceBoundary(t *testing.T) {
	for _, withoutMetadata := range []bool{false, true} {
		name := "symlink metadata"
		if withoutMetadata {
			name = "realpath fallback"
		}
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			external := t.TempDir()
			rootConfig := writeConfigCandidate(t, root, "rslint.config.js")
			rootGitignore := writeDiscoveryFixture(t, root, ".gitignore", "linked/scope/parent-hidden/\n")
			writeDiscoveryFixture(t, external, ".gitignore", "scope/target-hidden/\n")
			writeDiscoveryFixture(t, external, "scope/parent-hidden/index.ts", "export {}\n")
			writeDiscoveryFixture(t, external, "scope/target-hidden/index.ts", "export {}\n")
			alias := filepath.Join(root, "linked")
			if err := os.Symlink(external, alias); err != nil {
				t.Skipf("directory symlink unavailable: %v", err)
			}
			aliasGitignore := tspath.CombinePaths(alias, ".gitignore")
			requestDirectory := tspath.CombinePaths(alias, "scope")
			loader := newFixtureConfigLoader()
			loader.configs[rootConfig] = namedConfig("root")
			fsys := discoveryTestFS()
			if withoutMetadata {
				fsys = &configDiscoveryNoSymlinkMetadataFS{FS: fsys}
			}
			spy := &configDiscoveryReadSpyFS{FS: fsys}

			catalog, err := DiscoverAutomatic(context.Background(), spy, loader, ConfigDiscoveryRequest{
				CWD:         root,
				Directories: []string{requestDirectory},
			})
			if err != nil {
				t.Fatalf("DiscoverAutomatic: %v", err)
			}
			root = tspath.NormalizePath(root)
			alias = tspath.NormalizePath(alias)
			if got := catalog.ConfigDirectories(); !reflect.DeepEqual(got, []string{root}) {
				t.Fatalf("symlink-root owner = %v, want ancestor %q", got, root)
			}
			effective := catalog.Configs[root]
			if !effective.IsFileIgnored(tspath.CombinePaths(alias, "scope/parent-hidden/index.ts"), root) {
				t.Fatal("ancestor .gitignore rule was lost at the symlink boundary")
			}
			if effective.IsFileIgnored(tspath.CombinePaths(alias, "scope/target-hidden/index.ts"), root) {
				t.Fatal("target-side .gitignore crossed the symlink boundary")
			}
			spy.mu.Lock()
			reads := slices.Clone(spy.gitignorePaths)
			spy.mu.Unlock()
			if !slices.Contains(reads, rootGitignore) {
				t.Fatalf("ancestor .gitignore was not read: %v", reads)
			}
			if slices.Contains(reads, aliasGitignore) {
				t.Fatalf("target-side .gitignore was read through the symlink: %v", reads)
			}
		})
	}

	t.Run("config at the symlink root starts a new source boundary", func(t *testing.T) {
		root := t.TempDir()
		external := t.TempDir()
		rootConfig := writeConfigCandidate(t, root, "rslint.config.js")
		writeConfigCandidate(t, external, "rslint.config.js")
		writeDiscoveryFixture(t, external, ".gitignore", "hidden/\n")
		writeDiscoveryFixture(t, external, "hidden/index.ts", "export {}\n")
		alias := filepath.Join(root, "linked")
		if err := os.Symlink(external, alias); err != nil {
			t.Skipf("directory symlink unavailable: %v", err)
		}
		alias = tspath.NormalizePath(alias)
		aliasConfig := tspath.CombinePaths(alias, "rslint.config.js")
		aliasGitignore := tspath.CombinePaths(alias, ".gitignore")
		loader := newFixtureConfigLoader()
		loader.configs[rootConfig] = namedConfig("root")
		loader.configs[aliasConfig] = namedConfig("alias")
		spy := &configDiscoveryReadSpyFS{FS: discoveryTestFS()}

		catalog, err := DiscoverAutomatic(context.Background(), spy, loader, ConfigDiscoveryRequest{
			CWD:         root,
			Directories: []string{alias},
		})
		if err != nil {
			t.Fatalf("DiscoverAutomatic: %v", err)
		}
		if got := catalog.ConfigDirectories(); !reflect.DeepEqual(got, []string{alias}) {
			t.Fatalf("symlink-root config did not become the owner: %v", got)
		}
		if got := requestedConfigPaths(loader); !reflect.DeepEqual(got, []string{
			rootConfig,
			aliasConfig,
		}) {
			t.Fatalf("config requests = %v", got)
		}
		spy.mu.Lock()
		reads := slices.Clone(spy.gitignorePaths)
		spy.mu.Unlock()
		if !slices.Contains(reads, aliasGitignore) {
			t.Fatalf("symlink-root config did not start a Git source boundary: %v", reads)
		}
	})

	t.Run("successful child config starts a new source boundary", func(t *testing.T) {
		root := t.TempDir()
		external := t.TempDir()
		rootConfig := writeConfigCandidate(t, root, "rslint.config.js")
		writeDiscoveryFixture(t, external, ".gitignore", "child/\n")
		writeConfigCandidate(t, external, "child/rslint.config.js")
		writeDiscoveryFixture(t, external, "child/.gitignore", "hidden/\n")
		writeDiscoveryFixture(t, external, "child/hidden/index.ts", "export {}\n")
		alias := filepath.Join(root, "linked")
		if err := os.Symlink(external, alias); err != nil {
			t.Skipf("directory symlink unavailable: %v", err)
		}
		aliasGitignore := tspath.CombinePaths(alias, ".gitignore")
		aliasChild := tspath.CombinePaths(alias, "child")
		aliasChildConfig := tspath.CombinePaths(aliasChild, "rslint.config.js")
		aliasChildGitignore := tspath.CombinePaths(aliasChild, ".gitignore")
		loader := newFixtureConfigLoader()
		loader.configs[rootConfig] = namedConfig("root")
		loader.configs[aliasChildConfig] = namedConfig("child")
		spy := &configDiscoveryReadSpyFS{FS: discoveryTestFS()}

		catalog, err := DiscoverAutomatic(context.Background(), spy, loader, ConfigDiscoveryRequest{
			CWD:         root,
			Directories: []string{alias},
		})
		if err != nil {
			t.Fatalf("DiscoverAutomatic: %v", err)
		}
		if got := catalog.ConfigDirectories(); !reflect.DeepEqual(got, []string{
			tspath.NormalizePath(root),
			aliasChild,
		}) {
			t.Fatalf("child config did not reopen its Git source boundary: %v", got)
		}
		if got := requestedConfigPaths(loader); !reflect.DeepEqual(got, []string{
			rootConfig,
			aliasChildConfig,
		}) {
			t.Fatalf("config requests = %v", got)
		}
		spy.mu.Lock()
		reads := slices.Clone(spy.gitignorePaths)
		spy.mu.Unlock()
		if slices.Contains(reads, aliasGitignore) {
			t.Fatalf("pre-config target .gitignore crossed the symlink boundary: %v", reads)
		}
		if !slices.Contains(reads, aliasChildGitignore) {
			t.Fatalf("successful child config did not start a new Git source boundary: %v", reads)
		}
	})

	t.Run("failed child config keeps the source boundary closed", func(t *testing.T) {
		root := t.TempDir()
		external := t.TempDir()
		rootConfig := writeConfigCandidate(t, root, "rslint.config.js")
		failedConfig := writeConfigCandidate(t, external, "child/rslint.config.js")
		writeDiscoveryFixture(t, external, "child/.gitignore", "hidden/\n")
		writeConfigCandidate(t, external, "child/hidden/rslint.config.js")
		alias := filepath.Join(root, "linked")
		if err := os.Symlink(external, alias); err != nil {
			t.Skipf("directory symlink unavailable: %v", err)
		}
		aliasChildConfig := tspath.CombinePaths(alias, "child/rslint.config.js")
		aliasChildGitignore := tspath.CombinePaths(alias, "child/.gitignore")
		aliasHiddenConfig := tspath.CombinePaths(alias, "child/hidden/rslint.config.js")
		loader := newFixtureConfigLoader()
		loader.configs[rootConfig] = namedConfig("root")
		loader.failures[aliasChildConfig] = ConfigModuleError{Code: "ERR", Message: "broken"}
		loader.configs[aliasHiddenConfig] = namedConfig("visible-through-failed-boundary")
		spy := &configDiscoveryReadSpyFS{FS: discoveryTestFS()}

		catalog, err := DiscoverAutomatic(context.Background(), spy, loader, ConfigDiscoveryRequest{
			CWD:         root,
			Directories: []string{alias},
		})
		if err != nil {
			t.Fatalf("DiscoverAutomatic: %v", err)
		}
		if got := catalog.ConfigDirectories(); !reflect.DeepEqual(got, []string{
			tspath.NormalizePath(root),
			tspath.GetDirectoryPath(aliasHiddenConfig),
		}) {
			t.Fatalf("failed child unexpectedly changed Git source reachability: %v", got)
		}
		if got := requestedConfigPaths(loader); !reflect.DeepEqual(got, []string{
			rootConfig,
			aliasChildConfig,
			aliasHiddenConfig,
		}) {
			t.Fatalf("config requests = %v, physical failed config %q", got, failedConfig)
		}
		spy.mu.Lock()
		reads := slices.Clone(spy.gitignorePaths)
		spy.mu.Unlock()
		if slices.Contains(reads, aliasChildGitignore) {
			t.Fatalf("failed child config reopened Git source traversal: %v", reads)
		}
	})
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
		{
			name: "bounded static glob root",
			request: func(root string, target string) ConfigDiscoveryRequest {
				return ConfigDiscoveryRequest{
					CWD:                       root,
					Directories:               []string{target},
					LimitDirectoryWalkToFiles: true,
					Files: []DiscoveryFile{{
						Path:     tspath.CombinePaths(target, "index.ts"),
						Explicit: false,
					}},
				}
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
			wantRequested := []string{rootConfig, ignoredNestedConfig}
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

	t.Run("gitignore prunes a target branch before its config", func(t *testing.T) {
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
		wantDirs := []string{tspath.NormalizePath(root)}
		if got := catalog.ConfigDirectories(); !reflect.DeepEqual(got, wantDirs) {
			t.Fatalf("gitignore catalog = %v, want %v", got, wantDirs)
		}
		if got := requestedConfigPaths(loader); !reflect.DeepEqual(got, []string{rootConfig}) {
			t.Fatalf("gitignore-hidden config was evaluated: %v", got)
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

func TestConfigDiscoveryExplicitConfigLoadsBeforeProjectingGitignore(t *testing.T) {
	root := t.TempDir()
	ignoredTarget := writeDiscoveryFixture(t, root, "ignored/index.ts", "export {};\n")
	visibleTarget := writeDiscoveryFixture(t, root, "src/index.ts", "export {};\n")
	writeDiscoveryFixture(t, root, ".gitignore", "ignored/\n")
	loader := newFixtureConfigLoader()
	configPath := writeConfigCandidate(t, root, "ignored/custom.config.cjs")
	loader.configs[configPath] = namedConfig("explicit")

	catalog, err := LoadExplicitConfig(
		context.Background(),
		discoveryTestFS(),
		loader,
		ExplicitConfigRequest{CWD: root, ConfigPath: configPath},
	)
	if err != nil {
		t.Fatalf("LoadExplicitConfig: %v", err)
	}
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
	config := catalog.Configs[root]
	if got := config.GetConfigForFile(ignoredTarget, root); got != nil {
		t.Fatalf("Git-ignored target received explicit config: %+v", got)
	}
	if got := config.GetConfigForFile(visibleTarget, root); got == nil {
		t.Fatal("visible target did not receive explicit config")
	}
}

func TestConfigDiscoveryExplicitConfigDoesNotDiscoverNestedCandidates(t *testing.T) {
	root := t.TempDir()
	loader := newFixtureConfigLoader()
	configPath := writeConfigCandidate(t, root, "custom.config.cjs")
	nestedConfigPath := writeConfigCandidate(t, root, "nested/rslint.config.js")
	loader.configs[configPath] = namedConfig("explicit")
	loader.configs[nestedConfigPath] = namedConfig("must-not-load")
	writeDiscoveryFixture(t, root, "nested/.gitignore", "generated.ts\n")
	ignoredTarget := writeDiscoveryFixture(t, root, "nested/generated.ts", "export {};\n")

	catalog, err := LoadExplicitConfig(
		context.Background(),
		discoveryTestFS(),
		loader,
		ExplicitConfigRequest{CWD: root, ConfigPath: configPath},
	)
	if err != nil {
		t.Fatalf("LoadExplicitConfig: %v", err)
	}
	if got := requestedConfigPaths(loader); !reflect.DeepEqual(got, []string{configPath}) {
		t.Fatalf("explicit projection requested nested configs: %v", got)
	}
	root = tspath.NormalizePath(root)
	if got := catalog.Configs[root].GetConfigForFile(ignoredTarget, root); got != nil {
		t.Fatalf("nested Git-ignored target received explicit config: %+v", got)
	}
}

func TestConfigDiscoveryExplicitConfigTargetFileProjection(t *testing.T) {
	root := t.TempDir()
	loader := newFixtureConfigLoader()
	configPath := writeConfigCandidate(t, root, "custom.config.cjs")
	loader.configs[configPath] = namedConfig("explicit")
	target := writeDiscoveryFixture(t, root, "src/index.ts", "export {};\n")
	writeDiscoveryFixture(t, root, "src/.gitignore", "index.ts\n")
	writeDiscoveryFixture(t, root, "unrelated/.gitignore", "other.ts\n")

	spyFS := &configDiscoveryReadSpyFS{FS: discoveryTestFS()}
	catalog, err := LoadExplicitConfig(
		context.Background(),
		spyFS,
		loader,
		ExplicitConfigRequest{
			CWD:         root,
			ConfigPath:  configPath,
			TargetFiles: []DiscoveryFile{{Path: target}},
		},
	)
	if err != nil {
		t.Fatalf("LoadExplicitConfig: %v", err)
	}
	root = tspath.NormalizePath(root)
	if got := catalog.Configs[root].GetConfigForFile(target, root); got != nil {
		t.Fatalf("target-chain Git ignore was not projected: %+v", got)
	}
	wantReads := []string{
		tspath.CombinePaths(root, ".gitignore"),
		tspath.CombinePaths(root, "src/.gitignore"),
	}
	if got := spyFS.gitignorePaths; !reflect.DeepEqual(got, wantReads) {
		t.Fatalf("exact projection Git reads = %v, want %v", got, wantReads)
	}
}

func TestConfigDiscoveryExplicitConfigTargetProjectionUsesConfigPathSpace(t *testing.T) {
	physicalRoot := t.TempDir()
	writeConfigCandidate(t, physicalRoot, "custom.config.cjs")
	writeDiscoveryFixture(t, physicalRoot, ".gitignore", "src/index.ts\n")
	physicalTarget := writeDiscoveryFixture(t, physicalRoot, "src/index.ts", "export {};\n")
	aliasRoot := filepath.Join(t.TempDir(), "workspace-alias")
	if err := os.Symlink(physicalRoot, aliasRoot); err != nil {
		t.Skipf("directory symlink unavailable: %v", err)
	}
	aliasRoot = tspath.NormalizePath(aliasRoot)
	aliasTarget := tspath.CombinePaths(aliasRoot, "src/index.ts")
	externalSpelling := writeDiscoveryFixture(t, t.TempDir(), "index.ts", "export {};\n")
	canonicalTarget := discoveryTestFS().Realpath(physicalTarget)

	for _, test := range []struct {
		name      string
		cwd       string
		target    string
		canonical string
	}{
		{name: "aliased CWD and physical target", cwd: aliasRoot, target: physicalTarget},
		{name: "physical CWD and aliased target", cwd: physicalRoot, target: aliasTarget},
		{
			name:      "canonical target hint maps an external spelling",
			cwd:       physicalRoot,
			target:    externalSpelling,
			canonical: canonicalTarget,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			configPath := tspath.CombinePaths(test.cwd, "custom.config.cjs")
			loader := newFixtureConfigLoader()
			loader.configs[configPath] = namedConfig("explicit")
			spyFS := &configDiscoveryReadSpyFS{FS: discoveryTestFS()}

			catalog, err := LoadExplicitConfig(
				context.Background(),
				spyFS,
				loader,
				ExplicitConfigRequest{
					CWD:        test.cwd,
					ConfigPath: configPath,
					TargetFiles: []DiscoveryFile{{
						Path:          test.target,
						CanonicalPath: test.canonical,
					}},
				},
			)
			if err != nil {
				t.Fatalf("LoadExplicitConfig: %v", err)
			}
			cwd := tspath.NormalizePath(test.cwd)
			matchFile, matchDir := rslintconfig.ResolveConfigPathSpaceWithCanonical(
				test.target,
				test.canonical,
				cwd,
				spyFS,
			)
			if got := catalog.Configs[cwd].GetConfigForFile(matchFile, matchDir); got != nil {
				t.Fatalf(
					"path-space Git ignore was not projected: config=%+v match=(%q,%q) reads=%v",
					catalog.Configs[cwd],
					matchFile,
					matchDir,
					spyFS.gitignorePaths,
				)
			}
			wantReads := []string{
				tspath.CombinePaths(cwd, ".gitignore"),
				tspath.CombinePaths(cwd, "src/.gitignore"),
			}
			if got := spyFS.gitignorePaths; !reflect.DeepEqual(got, wantReads) {
				t.Fatalf("path-space Git reads = %v, want %v", got, wantReads)
			}
		})
	}
}

func TestConfigDiscoveryExplicitConfigEmptyTargetSetSkipsGitProjection(t *testing.T) {
	root := t.TempDir()
	loader := newFixtureConfigLoader()
	configPath := writeConfigCandidate(t, root, "custom.config.cjs")
	loader.configs[configPath] = namedConfig("explicit")
	writeDiscoveryFixture(t, root, ".gitignore", "*.ts\n")

	spyFS := &configDiscoveryReadSpyFS{FS: discoveryTestFS()}
	catalog, err := LoadExplicitConfig(
		context.Background(),
		spyFS,
		loader,
		ExplicitConfigRequest{
			CWD:         root,
			ConfigPath:  configPath,
			TargetFiles: []DiscoveryFile{},
		},
	)
	if err != nil {
		t.Fatalf("LoadExplicitConfig: %v", err)
	}
	if spyFS.gitignoreReads != 0 {
		t.Fatalf("empty target projection read %d .gitignore files: %v", spyFS.gitignoreReads, spyFS.gitignorePaths)
	}
	root = tspath.NormalizePath(root)
	if got := namedConfigEntry(catalog.Configs[root]).Name; got != "explicit" {
		t.Fatalf("explicit config missing after empty projection: %q", got)
	}
}

func TestConfigDiscoveryExplicitGitProjectionMatchesStandalonePolicy(t *testing.T) {
	tests := []struct {
		name          string
		files         map[string]string
		configIgnores []string
		exactTargets  []string
		exact         bool
		check         []string
	}{
		{
			name: "full nested sources and Git negation",
			files: map[string]string{
				".gitignore":              "*.generated.ts\n",
				"src/.gitignore":          "ignored.ts\n!keep.generated.ts\n",
				"src/index.ts":            "export {};\n",
				"src/ignored.ts":          "export {};\n",
				"src/drop.generated.ts":   "export {};\n",
				"src/keep.generated.ts":   "export {};\n",
				"other/keep.generated.ts": "export {};\n",
			},
			check: []string{
				"src/index.ts",
				"src/ignored.ts",
				"src/drop.generated.ts",
				"src/keep.generated.ts",
				"other/keep.generated.ts",
			},
		},
		{
			name: "authored global negation after Git projection",
			files: map[string]string{
				".gitignore":    "dist/\n",
				"dist/index.ts": "export {};\n",
				"src/index.ts":  "export {};\n",
			},
			configIgnores: []string{"!dist/**"},
			check:         []string{"dist/index.ts", "src/index.ts"},
		},
		{
			name: "exact chains exclude unrelated sources",
			files: map[string]string{
				".gitignore":              "*.generated.ts\n",
				"src/.gitignore":          "index.ts\n",
				"src/index.ts":            "export {};\n",
				"unrelated/.gitignore":    "index.ts\n",
				"unrelated/index.ts":      "export {};\n",
				"unrelated/generated.ts":  "export {};\n",
				"outside-target/index.ts": "export {};\n",
			},
			exactTargets: []string{"src/index.ts"},
			exact:        true,
			check: []string{
				"src/index.ts",
				"unrelated/index.ts",
				"unrelated/generated.ts",
				"outside-target/index.ts",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			for relativePath, content := range test.files {
				writeDiscoveryFixture(t, root, relativePath, content)
			}
			configPath := writeConfigCandidate(t, root, "custom.config.cjs")
			entries := rslintconfig.RslintConfig{}
			if test.configIgnores != nil {
				entries = append(entries, rslintconfig.ConfigEntry{
					Ignores: append([]string(nil), test.configIgnores...),
				})
			}
			entries = append(entries, namedConfig("explicit")...)
			loader := newFixtureConfigLoader()
			loader.configs[configPath] = entries

			var requestTargets []DiscoveryFile
			var standaloneTargets []string
			if test.exact {
				requestTargets = make([]DiscoveryFile, 0, len(test.exactTargets))
				standaloneTargets = make([]string, 0, len(test.exactTargets))
				for _, relativePath := range test.exactTargets {
					path := tspath.NormalizePath(filepath.Join(root, filepath.FromSlash(relativePath)))
					requestTargets = append(requestTargets, DiscoveryFile{Path: path})
					standaloneTargets = append(standaloneTargets, path)
				}
			}
			fsys := discoveryTestFS()
			catalog, err := LoadExplicitConfig(
				context.Background(),
				fsys,
				loader,
				ExplicitConfigRequest{
					CWD:         root,
					ConfigPath:  configPath,
					TargetFiles: requestTargets,
				},
			)
			if err != nil {
				t.Fatalf("LoadExplicitConfig: %v", err)
			}
			root = tspath.NormalizePath(root)
			projected := catalog.Configs[root]
			standalone := rslintconfig.ConfigWithGitignore(entries, root, fsys, standaloneTargets)
			for _, relativePath := range test.check {
				path := tspath.NormalizePath(filepath.Join(root, filepath.FromSlash(relativePath)))
				gotSelected := projected.GetConfigForFile(path, root) != nil
				wantSelected := standalone.GetConfigForFile(path, root) != nil
				if gotSelected != wantSelected {
					t.Errorf(
						"%s selected by staged projection = %t, standalone policy = %t",
						relativePath,
						gotSelected,
						wantSelected,
					)
				}
			}
		})
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

		catalog, err := DiscoverAutomatic(context.Background(), fsys, loader, ConfigDiscoveryRequest{
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

	t.Run("directory reads lexical Git source in canonical matching space", func(t *testing.T) {
		lexicalRoot := t.TempDir()
		physicalRoot := t.TempDir()
		loader := newFixtureConfigLoader()
		configPath := writeConfigCandidate(t, physicalRoot, "rslint.config.js")
		lexicalIgnored := writeDiscoveryFixture(t, lexicalRoot, "lexical-hidden/index.ts", "export {}\n")
		physicalOnly := writeDiscoveryFixture(t, lexicalRoot, "physical-hidden/index.ts", "export {}\n")
		lexicalGitignore := writeDiscoveryFixture(t, lexicalRoot, ".gitignore", "lexical-hidden/\n")
		physicalGitignore := writeDiscoveryFixture(t, physicalRoot, ".gitignore", "physical-hidden/\n")
		loader.configs[configPath] = namedConfig("physical-directory")
		fsys := &configDiscoveryReadSpyFS{FS: &configDiscoveryRealpathFS{
			FS: discoveryTestFS(),
			realPaths: map[string]string{
				tspath.NormalizePath(lexicalRoot): tspath.NormalizePath(physicalRoot),
			},
		}}

		catalog, err := DiscoverAutomatic(context.Background(), fsys, loader, ConfigDiscoveryRequest{
			CWD:                       lexicalRoot,
			Directories:               []string{lexicalRoot},
			LimitDirectoryWalkToFiles: true,
			Files: []DiscoveryFile{
				{Path: lexicalIgnored, Explicit: false},
				{Path: physicalOnly, Explicit: false},
			},
		})
		if err != nil {
			t.Fatalf("DiscoverAutomatic: %v", err)
		}
		physicalRoot = tspath.NormalizePath(physicalRoot)
		effective := catalog.Configs[physicalRoot]
		if !effective.IsFileIgnored(
			tspath.CombinePaths(physicalRoot, "lexical-hidden/index.ts"),
			physicalRoot,
		) {
			t.Fatal("lexical .gitignore was not projected into canonical owner space")
		}
		if effective.IsFileIgnored(
			tspath.CombinePaths(physicalRoot, "physical-hidden/index.ts"),
			physicalRoot,
		) {
			t.Fatal("canonical-root .gitignore replaced the lexical walk source")
		}
		fsys.mu.Lock()
		reads := slices.Clone(fsys.gitignorePaths)
		fsys.mu.Unlock()
		if !slices.Contains(reads, lexicalGitignore) || slices.Contains(reads, physicalGitignore) {
			t.Fatalf("Git source IO paths = %v, want lexical root only", reads)
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

func TestConfigDiscoveryCaseInsensitiveRootsChooseStableRepresentative(t *testing.T) {
	fsys := &configDiscoveryCaseSensitivityFS{
		FS:            discoveryTestFS(),
		caseSensitive: false,
	}
	build := func(directories []string) []string {
		builder := configCatalogBuilder{
			fs: fsys,
			request: ConfigDiscoveryRequest{
				CWD:         "C:/repo",
				Directories: directories,
			},
		}
		return builder.normalizedDirectoryRoots()
	}
	first := build([]string{"c:/repo", "C:/Repo"})
	second := build([]string{"C:/Repo", "c:/repo"})
	want := []string{"C:/Repo"}
	if !reflect.DeepEqual(first, want) || !reflect.DeepEqual(second, want) {
		t.Fatalf("case-insensitive roots depend on input order: first=%v second=%v", first, second)
	}
}

func TestConfigDiscoveryCandidateSearchStopsAtUNCShareRoot(t *testing.T) {
	shareConfig := "//server/share/rslint.config.js"
	serverConfig := "//server/rslint.config.js"

	t.Run("parent traversal", func(t *testing.T) {
		for _, test := range []struct {
			directory string
			parent    string
		}{
			{directory: "/repo/pkg", parent: "/repo"},
			{directory: "/", parent: ""},
			{directory: "C:/repo", parent: "C:/"},
			{directory: "C:/", parent: ""},
			{directory: "//server/share/repo", parent: "//server/share"},
			{directory: "//server/share", parent: ""},
			{directory: "//server/share/", parent: ""},
			{directory: "//server", parent: ""},
		} {
			if got := configDiscoveryParent(test.directory); got != test.parent {
				t.Fatalf("configDiscoveryParent(%q) = %q, want %q", test.directory, got, test.parent)
			}
		}
	})

	t.Run("ancestor chain includes the share root but not the server", func(t *testing.T) {
		fsys := &configDiscoveryCandidateFS{
			FS: discoveryTestFS(),
			files: map[string]struct{}{
				shareConfig:  {},
				serverConfig: {},
			},
		}
		builder := configCatalogBuilder{
			fs:      fsys,
			request: ConfigDiscoveryRequest{CWD: "//server/share/repo"},
		}

		candidates := builder.findCandidateChain("//server/share/repo")
		if len(candidates) != 1 || candidates[0].path != shareConfig {
			t.Fatalf("UNC ancestor candidates = %+v, want share root only", candidates)
		}
	})

	t.Run("nearest search cannot escape to the server", func(t *testing.T) {
		fsys := &configDiscoveryCandidateFS{
			FS: discoveryTestFS(),
			files: map[string]struct{}{
				serverConfig: {},
			},
		}
		builder := configCatalogBuilder{
			fs:      fsys,
			request: ConfigDiscoveryRequest{CWD: "//server/share/repo"},
		}

		if candidate, found := builder.findCandidateUp("//server/share/repo"); found {
			t.Fatalf("UNC nearest search escaped its share: %+v", candidate)
		}
	})

	t.Run("failed share candidate cannot fall back to the server", func(t *testing.T) {
		fsys := &configDiscoveryCandidateFS{
			FS: discoveryTestFS(),
			files: map[string]struct{}{
				shareConfig:  {},
				serverConfig: {},
			},
		}
		loader := newFixtureConfigLoader()
		loader.failures[shareConfig] = ConfigModuleError{Code: "ERR", Message: "broken"}
		loader.configs[serverConfig] = namedConfig("outside-share")

		_, err := DiscoverAutomatic(context.Background(), fsys, loader, ConfigDiscoveryRequest{
			CWD: "//server/share/repo",
			Files: []DiscoveryFile{{
				Path:     "//server/share/repo/src/index.ts",
				Explicit: true,
			}},
		})
		if !errors.Is(err, ErrAllConfigsFailed) {
			t.Fatalf("failed share config error = %v, want ErrAllConfigsFailed", err)
		}
		if got := requestedConfigPaths(loader); !reflect.DeepEqual(got, []string{shareConfig}) {
			t.Fatalf("requested configs = %v, want failed share config only", got)
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

		catalog, err := DiscoverAutomatic(context.Background(), discoveryTestFS(), loader, ConfigDiscoveryRequest{
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

		catalog, err := DiscoverAutomatic(context.Background(), fsys, loader, ConfigDiscoveryRequest{
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
		loader.plugins[upperConfig] = []rslintconfig.EslintPluginEntry{{Prefix: "upper", RuleNames: []string{"rule"}}}
		loader.plugins[lowerConfig] = []rslintconfig.EslintPluginEntry{{Prefix: "lower", RuleNames: []string{"rule"}}}
		realpathFS := &configDiscoveryRealpathFS{
			FS: discoveryTestFS(),
			realPaths: map[string]string{
				tspath.NormalizePath(upperRoot):   tspath.NormalizePath(upperRoot),
				tspath.NormalizePath(lowerRoot):   tspath.NormalizePath(upperRoot),
				tspath.NormalizePath(upperConfig): tspath.NormalizePath(upperConfig),
				tspath.NormalizePath(lowerConfig): tspath.NormalizePath(upperConfig),
			},
		}
		fsys := &configDiscoveryCaseSensitivityFS{FS: realpathFS, caseSensitive: false}

		catalog, err := DiscoverAutomatic(context.Background(), fsys, loader, ConfigDiscoveryRequest{
			CWD:         root,
			Directories: []string{upperRoot, lowerRoot},
		})
		if err != nil {
			t.Fatalf("Build: %v", err)
		}
		if len(catalog.Configs) != 1 || len(catalog.Scopes) != 1 {
			t.Fatalf("native case alias catalog: configs=%v scopes=%v, want one representative", catalog.Configs, catalog.Scopes)
		}
		if got := requestedConfigPaths(loader); !reflect.DeepEqual(got, []string{upperConfig}) {
			t.Fatalf("requested config paths = %v, want one deterministic representative", got)
		}
		if catalog.Stats.ConfigsRequested != 1 || catalog.Stats.ConfigsLoaded != 1 {
			t.Fatalf("native case alias stats = %+v, want one requested and loaded config", catalog.Stats)
		}
		if got := pluginPrefixes(catalog.EslintPlugins); !reflect.DeepEqual(got, []string{"upper"}) {
			t.Fatalf("effective plugin prefixes = %v, want representative plugin metadata", got)
		}
		if len(loader.activations) != 1 {
			t.Fatalf("activation count = %d, want 1", len(loader.activations))
		}
	})
}

func TestConfigDiscoveryReusesNativeCaseAliasAcrossLoadFrontiers(t *testing.T) {
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
			tspath.NormalizePath(upperRoot):   tspath.NormalizePath(upperRoot),
			tspath.NormalizePath(lowerRoot):   tspath.NormalizePath(upperRoot),
			tspath.NormalizePath(upperConfig): tspath.NormalizePath(upperConfig),
			tspath.NormalizePath(lowerConfig): tspath.NormalizePath(upperConfig),
		},
	}
	fsys := &configDiscoveryCaseSensitivityFS{FS: realpathFS, caseSensitive: false}
	builder := configCatalogBuilder{
		ctx:                 context.Background(),
		fs:                  fsys,
		loader:              loader,
		transactionID:       "case-alias-frontiers",
		loadStates:          make(map[string]*configLoadState),
		loadStateByIdentity: make(map[tspath.Path]*configLoadState),
		failureByPath:       make(map[string]ConfigFailure),
	}
	if err := builder.ensureCandidates([]configCandidate{{path: upperConfig, directory: upperRoot}}); err != nil {
		t.Fatalf("first frontier: %v", err)
	}
	if err := builder.ensureCandidates([]configCandidate{{path: lowerConfig, directory: lowerRoot}}); err != nil {
		t.Fatalf("second frontier: %v", err)
	}
	if len(loader.batches) != 1 || len(loader.batches[0].Candidates) != 1 {
		t.Fatalf("load batches = %+v, want one request across frontiers", loader.batches)
	}
	if builder.loadStates[upperConfig] == nil || builder.loadStates[upperConfig] != builder.loadStates[lowerConfig] {
		t.Fatal("case aliases did not resolve to the same representative load state")
	}
	if got := builder.loadStates[lowerConfig].candidate.path; got != upperConfig {
		t.Fatalf("representative config path = %q, want %q", got, upperConfig)
	}
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
	parallelCatalog, err := DiscoverAutomatic(context.Background(), probeFS, parallelLoader, ConfigDiscoveryRequest{
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
			catalog, err := DiscoverAutomatic(context.Background(), discoveryTestFS(), nil, ConfigDiscoveryRequest{CWD: root})
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

func TestConfigDiscoveryTransactionIDIncludesProcessNonce(t *testing.T) {
	id := nextConfigDiscoveryTransactionID()
	prefix := "config-discovery-" + configDiscoveryProcessNonce + "-"
	if !strings.HasPrefix(id, prefix) {
		t.Fatalf("transaction ID %q does not contain process nonce %q", id, configDiscoveryProcessNonce)
	}
	sequence, err := strconv.ParseUint(strings.TrimPrefix(id, prefix), 10, 64)
	if err != nil || sequence == 0 {
		t.Fatalf("transaction ID %q has invalid sequence: value=%d err=%v", id, sequence, err)
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
		return ConfigLoadResult{ID: id, Status: "loaded", Entries: rslintconfig.RslintConfig{}}
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
	nestedConfigPath := writeConfigCandidate(t, root, "packages/app/rslint.config.js")
	loader.failures[configPath] = ConfigModuleError{Code: "ERR", Message: "nope"}
	loader.failures[nestedConfigPath] = ConfigModuleError{Code: "ERR_NESTED", Message: "nested nope"}
	_, err := DiscoverAutomatic(context.Background(), discoveryTestFS(), loader, ConfigDiscoveryRequest{
		CWD:         root,
		ImplicitCWD: true,
	})
	if !errors.Is(err, ErrAllConfigsFailed) {
		t.Fatalf("error = %v, want ErrAllConfigsFailed", err)
	}
	if !strings.Contains(err.Error(), "failed to load config") {
		t.Fatalf("error = %v, want load-stage context", err)
	}
	var allFailed *AllConfigsFailedError
	wantFailurePaths := []string{configPath, nestedConfigPath}
	sort.Strings(wantFailurePaths)
	if !errors.As(err, &allFailed) || len(allFailed.Failures) != 2 ||
		!reflect.DeepEqual(
			[]string{allFailed.Failures[0].Path, allFailed.Failures[1].Path},
			wantFailurePaths,
		) {
		t.Fatalf("typed failure metadata = %#v, want sorted paths %v", allFailed, wantFailurePaths)
	}

	invalidLoader := newFixtureConfigLoader()
	invalidLoader.failures[configPath] = ConfigModuleError{Code: "invalid", Message: "not an array"}
	invalidLoader.failures[nestedConfigPath] = ConfigModuleError{Code: "invalid", Message: "nested not an array"}
	_, err = DiscoverAutomatic(context.Background(), discoveryTestFS(), invalidLoader, ConfigDiscoveryRequest{
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
	catalog, err := DiscoverAutomatic(context.Background(), discoveryTestFS(), loader, request)
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

func namedConfigEntry(entries rslintconfig.RslintConfig) rslintconfig.ConfigEntry {
	for _, entry := range entries {
		if entry.Name != "" {
			return entry
		}
	}
	return rslintconfig.ConfigEntry{}
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
