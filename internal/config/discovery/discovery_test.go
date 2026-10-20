// cspell:ignore gitdir

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

	"github.com/microsoft/TypeScript/tsc/shim/bundled"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/cachedvfs"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/config/target"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type fakeConfigModuleLoader struct {
	configs     map[string]rslintconfig.RslintConfig
	failures    map[string]ConfigModuleError
	plugins     map[string][]rslintconfig.EslintPluginEntry
	batches     []ConfigLoadBatchRequest
	activations []ConfigActivationRequest
	pluginsByID map[string][]rslintconfig.EslintPluginEntry
	mutate      func(ConfigLoadBatchRequest, ConfigLoadBatchResponse) ConfigLoadBatchResponse
	activate    func(ConfigActivationRequest) (ConfigActivationResponse, error)
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
	accessedDirs   []string
	markerPaths    []string
}

type configDiscoveryCandidateFS struct {
	vfs.FS
	files map[string]struct{}
}

func (fs *configDiscoveryCandidateFS) FileExists(path string) bool {
	_, exists := fs.files[tspath.NormalizePath(path)]
	return exists
}

func (fs *configDiscoveryReadSpyFS) recordMarker(path string) {
	if tspath.GetBaseFileName(path) != ".git" {
		return
	}
	fs.mu.Lock()
	fs.markerPaths = append(fs.markerPaths, tspath.NormalizePath(path))
	fs.mu.Unlock()
}

func (fs *configDiscoveryReadSpyFS) FileExists(path string) bool {
	fs.recordMarker(path)
	return fs.FS.FileExists(path)
}

func (fs *configDiscoveryReadSpyFS) DirectoryExists(path string) bool {
	fs.recordMarker(path)
	return fs.FS.DirectoryExists(path)
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

func (fs *configDiscoveryReadSpyFS) GetAccessibleEntries(path string) vfs.Entries {
	fs.mu.Lock()
	fs.accessedDirs = append(fs.accessedDirs, tspath.NormalizePath(path))
	fs.mu.Unlock()
	return fs.FS.GetAccessibleEntries(path)
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
		lexicalPath := tspath.CombinePaths(candidate.ConfigDirectory, tspath.GetBaseFileName(path))
		if failure, failed := configFixtureAtPath(loader.failures, path, lexicalPath); failed {
			failure := failure
			response.Results = append(response.Results, ConfigLoadResult{
				ID:     candidate.ID,
				Status: "failed",
				Error:  &failure,
			})
			continue
		}
		config, _ := configFixtureAtPath(loader.configs, path, lexicalPath)
		entries := append(rslintconfig.RslintConfig(nil), config...)
		pluginEntries, _ := configFixtureAtPath(loader.plugins, path, lexicalPath)
		plugins := cloneEslintPluginEntries(pluginEntries)
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

func configFixtureAtPath[T any](values map[string]T, physical, lexical string) (T, bool) {
	for _, path := range []string{physical, lexical} {
		if value, exists := values[path]; exists {
			return value, true
		}
		for key, value := range values {
			if rslintconfig.ExactPathID(key) == rslintconfig.ExactPathID(path) {
				return value, true
			}
		}
	}
	var empty T
	return empty, false
}

func (loader *fakeConfigModuleLoader) ActivateConfigs(_ context.Context, request ConfigActivationRequest) (ConfigActivationResponse, error) {
	loader.activations = append(loader.activations, request)
	if loader.activate != nil {
		return loader.activate(request)
	}
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
		root := discoveryTempDir(t)
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
		root := discoveryTempDir(t)
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
	root := discoveryTempDir(t)
	writeDiscoveryFixture(t, root, ".gitignore", "ignored/\nrslint.config.js\n")
	configPath := writeConfigCandidate(t, root, "rslint.config.js")
	writeConfigCandidate(t, root, "ignored/rslint.config.js")
	loader := newFixtureConfigLoader()
	loader.configs[configPath] = namedConfig("root")
	loader.configs[tspath.CombinePaths(root, "ignored/rslint.config.js")] = namedConfig("nested")
	fsys := &configDiscoveryReadSpyFS{FS: discoveryTestFS()}

	catalog, err := discoverTestAutomatic(context.Background(), fsys, loader, ConfigDiscoveryRequest{
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
			root := discoveryTempDir(t)
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
			root := discoveryTempDir(t)
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
	root := discoveryTempDir(t)
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
	t.Run("successful child preserves inherited Git and applies its own source", func(t *testing.T) {
		root := discoveryTempDir(t)
		rootConfig := writeConfigCandidate(t, root, "rslint.config.js")
		childConfig := writeConfigCandidate(t, root, "packages/app/rslint.config.js")
		reachableConfig := writeConfigCandidate(t, root, "packages/app/parent-only/rslint.config.js")
		hiddenConfig := writeConfigCandidate(t, root, "packages/app/child-only/rslint.config.js")
		writeDiscoveryFixture(t, root, ".gitignore", "packages/app/parent-only/\n")
		writeDiscoveryFixture(t, root, "packages/app/.gitignore", "child-only/\n")
		loader := newFixtureConfigLoader()
		loader.configs[rootConfig] = namedConfig("root")
		loader.configs[childConfig] = namedConfig("child")
		loader.configs[reachableConfig] = namedConfig("must-remain-ignored")
		loader.configs[hiddenConfig] = namedConfig("must-not-load")

		catalog := buildFixtureCatalog(t, root, loader, ConfigDiscoveryRequest{
			CWD:         root,
			ImplicitCWD: true,
		})
		wantDirectories := []string{
			tspath.NormalizePath(root),
			tspath.CombinePaths(root, "packages/app"),
		}
		if got := catalog.ConfigDirectories(); !reflect.DeepEqual(got, wantDirectories) {
			t.Fatalf("owner-boundary directories = %v, want %v", got, wantDirectories)
		}
		if got := requestedConfigPaths(loader); !reflect.DeepEqual(got, []string{rootConfig, childConfig}) {
			t.Fatalf("owner-boundary requests = %v", got)
		}
	})

	t.Run("failed child inherits parent Git", func(t *testing.T) {
		root := discoveryTempDir(t)
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
	root := discoveryTempDir(t)
	rootConfig := writeConfigCandidate(t, root, "rslint.config.js")
	hiddenConfig := writeConfigCandidate(t, root, "ignored/deep/rslint.config.js")
	rootGitignore := writeDiscoveryFixture(t, root, ".gitignore", "ignored/\n")
	hiddenGitignore := writeDiscoveryFixture(t, root, "ignored/.gitignore", "!deep/\n")
	loader := newFixtureConfigLoader()
	loader.configs[rootConfig] = namedConfig("root")
	loader.configs[hiddenConfig] = namedConfig("must-not-load")
	fsys := &configDiscoveryReadSpyFS{FS: discoveryTestFS()}

	_, err := discoverTestAutomatic(context.Background(), fsys, loader, ConfigDiscoveryRequest{
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
	root := discoveryTempDir(t)
	ignoredDirectory := tspath.CombinePaths(root, "ignored")
	rootConfig := writeConfigCandidate(t, root, "rslint.config.js")
	childConfig := writeConfigCandidate(t, root, "ignored/rslint.config.js")
	writeDiscoveryFixture(t, root, ".gitignore", "ignored/\n")
	loader := newFixtureConfigLoader()
	loader.configs[rootConfig] = namedConfig("root")
	loader.configs[childConfig] = namedConfig("child")

	catalog := buildFixtureCatalog(t, root, loader, ConfigDiscoveryRequest{
		CWD:         root,
		Directories: discoveryDirectoryInputs([]string{root, ignoredDirectory}),
	})
	if got := catalog.ConfigDirectories(); !reflect.DeepEqual(got, []string{tspath.NormalizePath(root), ignoredDirectory}) {
		t.Fatalf("overlapping-root config directories = %v", got)
	}
	if got := requestedConfigPaths(loader); !reflect.DeepEqual(got, []string{rootConfig, childConfig}) {
		t.Fatalf("overlapping roots executed duplicate or hidden configs: %v", got)
	}
}

func TestConfigDiscoveryCatalogFreezesCollectedGitignore(t *testing.T) {
	root := discoveryTempDir(t)
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
	if !catalogTestConfig(t, catalog, root).IsFileIgnored(target, root) {
		t.Fatal("catalog did not materialize the discovered Git source")
	}
	if catalogTestConfig(t, catalog, root).IsFileIgnored(reincludedTarget, root) {
		t.Fatal("authored negation did not follow and override the Git projection")
	}
	writeDiscoveryFixture(t, root, ".gitignore", "")
	if !catalogTestConfig(t, catalog, root).IsFileIgnored(target, root) {
		t.Fatal("published catalog changed after .gitignore mutation")
	}
	if catalogTestConfig(t, catalog, root).IsFileIgnored(reincludedTarget, root) {
		t.Fatal("published Git-to-authored ordering changed after source mutation")
	}
}

func TestConfigDiscoveryMergesLiteralGitignoreChainIntoAutomaticOwner(t *testing.T) {
	root := discoveryTempDir(t)
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
		Directories: discoveryDirectoryInputs([]string{tspath.GetDirectoryPath(automaticTarget)}),
		Files: []DiscoveryFile{
			{PathIdentity: rslintconfig.PathIdentity{Path: automaticTarget}, Explicit: false},
			{PathIdentity: rslintconfig.PathIdentity{Path: reincludedLiteral}, Explicit: true},
			{PathIdentity: rslintconfig.PathIdentity{Path: ignoredLiteral}, Explicit: true},
		},
	})
	root = tspath.NormalizePath(root)
	effective := catalogTestConfig(t, catalog, root)
	if effective.IsFileIgnored(reincludedLiteral, root) {
		t.Fatal("nested literal-chain negation was lost or parent Git source was duplicated")
	}
	if !effective.IsFileIgnored(ignoredLiteral, root) {
		t.Fatal("automatic owner did not include the literal target's Git chain")
	}
}

func TestConfigDiscoveryCachesLiteralGitignoreSourceChains(t *testing.T) {
	root := discoveryTempDir(t)
	rootConfig := writeConfigCandidate(t, root, "rslint.config.js")
	writeDiscoveryFixture(t, root, ".gitignore", "*.tmp\n")
	writeDiscoveryFixture(t, root, "outside/.gitignore", "!keep.tmp\n")
	loader := newFixtureConfigLoader()
	loader.configs[rootConfig] = namedConfig("root")
	var files []DiscoveryFile
	for _, name := range []string{"keep.tmp", "drop.tmp", "nested/one.tmp", "nested/two.tmp"} {
		files = append(files, DiscoveryFile{PathIdentity: rslintconfig.PathIdentity{Path: writeDiscoveryFixture(t, root, "outside/"+name, "export {}\n")}, Explicit: true})
	}
	fsys := &configDiscoveryReadSpyFS{FS: discoveryTestFS()}

	_, err := discoverTestAutomatic(context.Background(), fsys, loader, ConfigDiscoveryRequest{
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
	root := discoveryTempDir(t)
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

	catalog, err := discoverTestAutomatic(context.Background(), fsys, loader, ConfigDiscoveryRequest{
		CWD:         root,
		Directories: discoveryDirectoryInputs([]string{root, directDirectory}),
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
	if !catalogTestConfig(t, catalog, tspath.NormalizePath(root)).IsFileIgnored(hiddenTarget, root) {
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
	root := discoveryTempDir(t)
	external := discoveryTempDir(t)
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

	catalog, err := discoverTestAutomatic(context.Background(), fsys, loader, ConfigDiscoveryRequest{
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

func TestConfigDiscoveryDirectorySymlinkRootUsesPhysicalGitTree(t *testing.T) {
	for _, withoutMetadata := range []bool{false, true} {
		name := "symlink metadata"
		if withoutMetadata {
			name = "realpath fallback"
		}
		t.Run(name, func(t *testing.T) {
			root, external := discoveryTempDir(t), discoveryTempDir(t)
			rootConfig := writeConfigCandidate(t, root, "rslint.config.js")
			rootGitignore := writeDiscoveryFixture(t, root, ".gitignore", "linked/scope/parent-hidden/\n")
			physicalGitignore := writeDiscoveryFixture(t, external, ".gitignore", "scope/target-hidden/\n")
			if err := os.Mkdir(filepath.Join(external, ".git"), 0o755); err != nil {
				t.Fatal(err)
			}
			writeDiscoveryFixture(t, external, "scope/parent-hidden/index.ts", "export {}\n")
			writeDiscoveryFixture(t, external, "scope/target-hidden/index.ts", "export {}\n")
			alias := filepath.Join(root, "linked")
			if err := os.Symlink(external, alias); err != nil {
				t.Skipf("directory symlink unavailable: %v", err)
			}
			loader := newFixtureConfigLoader()
			loader.configs[rootConfig] = namedConfig("lexical-owner")
			fsys := discoveryTestFS()
			if withoutMetadata {
				fsys = &configDiscoveryNoSymlinkMetadataFS{FS: fsys}
			}
			spy := &configDiscoveryReadSpyFS{FS: fsys}
			catalog, err := discoverTestAutomatic(context.Background(), spy, loader, ConfigDiscoveryRequest{
				CWD: root, Directories: discoveryDirectoryInputs([]string{tspath.CombinePaths(alias, "scope")}),
			})
			if err != nil {
				t.Fatal(err)
			}
			if got := catalog.ConfigDirectories(); !reflect.DeepEqual(got, []string{root}) {
				t.Fatalf("config owner = %v", got)
			}
			matcher := catalogTestConfig(t, catalog, root)
			if matcher.IsFileIgnored(tspath.CombinePaths(alias, "scope/parent-hidden/index.ts"), root) {
				t.Fatal("lexical owner leaked its unrelated Git source")
			}
			if !matcher.IsFileIgnored(tspath.CombinePaths(alias, "scope/target-hidden/index.ts"), root) {
				t.Fatal("physical ancestor Git rule was lost")
			}
			if slices.Contains(spy.gitignorePaths, rootGitignore) || slices.Contains(spy.gitignorePaths, tspath.CombinePaths(alias, ".gitignore")) || !slices.Contains(spy.gitignorePaths, physicalGitignore) {
				t.Fatalf("Git source paths = %v, want physical chain only", spy.gitignorePaths)
			}
		})
	}

	t.Run("failed config does not close physical Git sources", func(t *testing.T) {
		root, external := discoveryTempDir(t), discoveryTempDir(t)
		rootConfig := writeConfigCandidate(t, root, "rslint.config.js")
		writeConfigCandidate(t, external, "child/rslint.config.js")
		writeConfigCandidate(t, external, "child/hidden/rslint.config.js")
		physicalSource := writeDiscoveryFixture(t, external, "child/.gitignore", "hidden/\n")
		alias := filepath.Join(root, "linked")
		if err := os.Symlink(external, alias); err != nil {
			t.Skipf("directory symlink unavailable: %v", err)
		}
		childConfig := tspath.CombinePaths(alias, "child/rslint.config.js")
		loader := newFixtureConfigLoader()
		loader.configs[rootConfig] = namedConfig("root")
		loader.failures[childConfig] = ConfigModuleError{Code: "ERR", Message: "broken"}
		spy := &configDiscoveryReadSpyFS{FS: discoveryTestFS()}
		catalog, err := discoverTestAutomatic(context.Background(), spy, loader, ConfigDiscoveryRequest{CWD: root, Directories: discoveryDirectoryInputs([]string{alias})})
		if err != nil {
			t.Fatal(err)
		}
		if got := catalog.ConfigDirectories(); !reflect.DeepEqual(got, []string{root}) {
			t.Fatalf("owners = %v", got)
		}
		if got := requestedConfigPaths(loader); !reflect.DeepEqual(got, []string{rootConfig, tspath.CombinePaths(external, "child/rslint.config.js")}) {
			t.Fatalf("config requests = %v", got)
		}
		if !slices.Contains(spy.gitignorePaths, physicalSource) {
			t.Fatalf("failed config prevented physical source read: %v", spy.gitignorePaths)
		}
	})
}

func TestConfigDiscoveryParentGlobalIgnorePrunesNestedConfig(t *testing.T) {
	root := discoveryTempDir(t)
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
		Directories:               discoveryDirectoryInputs([]string{root}),
		LimitDirectoryWalkToFiles: true,
		Files: []DiscoveryFile{
			{PathIdentity: rslintconfig.PathIdentity{Path: filepath.Join(root, "fixtures/deep/index.ts")}, Explicit: false},
			{PathIdentity: rslintconfig.PathIdentity{Path: visibleFile}, Explicit: false},
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
		root := discoveryTempDir(t)
		loader := newFixtureConfigLoader()
		rootConfig := writeConfigCandidate(t, root, "rslint.config.js")
		nestedConfig := writeConfigCandidate(t, root, "packages/app/rslint.config.js")
		writeDiscoveryFixture(t, root, "packages/app/index.ts", "export {}\n")
		loader.configs[rootConfig] = rslintconfig.RslintConfig{
			{Ignores: []string{"packages/**/*"}},
			{Name: "root", Rules: rslintconfig.Rules{}},
		}
		loader.configs[nestedConfig] = namedConfig("must-not-load")

		catalog := buildFixtureCatalog(t, root, loader, ConfigDiscoveryRequest{CWD: root, Directories: discoveryDirectoryInputs([]string{root})})
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
		root := discoveryTempDir(t)
		loader := newFixtureConfigLoader()
		rootConfig := writeConfigCandidate(t, root, "rslint.config.js")
		nestedConfig := writeConfigCandidate(t, root, "packages/app/rslint.config.js")
		loader.configs[rootConfig] = rslintconfig.RslintConfig{
			{Ignores: []string{"packages/**/*", "!packages/app/rslint.config.js"}},
			{Name: "root", Rules: rslintconfig.Rules{}},
		}
		loader.configs[nestedConfig] = namedConfig("nested")

		catalog := buildFixtureCatalog(t, root, loader, ConfigDiscoveryRequest{CWD: root, Directories: discoveryDirectoryInputs([]string{root})})
		wantDirs := []string{tspath.NormalizePath(root), tspath.CombinePaths(root, "packages/app")}
		if got := catalog.ConfigDirectories(); !reflect.DeepEqual(got, wantDirs) {
			t.Fatalf("negated candidate config directories = %v, want %v", got, wantDirs)
		}
	})

	t.Run("ignored highest priority falls through", func(t *testing.T) {
		root := discoveryTempDir(t)
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

		catalog := buildFixtureCatalog(t, root, loader, ConfigDiscoveryRequest{CWD: root, Directories: discoveryDirectoryInputs([]string{root})})
		if got := catalog.Configs[tspath.CombinePaths(root, "packages/app")][0].Name; got != "selected" {
			t.Fatalf("selected config = %q, want selected .mjs config", got)
		}
		if slices.Contains(requestedConfigPaths(loader), ignoredJS) {
			t.Fatalf("ignored highest-priority candidate was requested: %v", requestedConfigPaths(loader))
		}
	})
}

func TestConfigDiscoveryAutomaticCandidateWinsSameDirectoryLiteralConflict(t *testing.T) {
	root := tspath.NormalizePath(discoveryTempDir(t))
	writeConfigCandidate(t, root, "rslint.config.js")
	writeConfigCandidate(t, root, "packages/app/rslint.config.js")
	writeConfigCandidate(t, root, "packages/app/rslint.config.mjs")
	literal := writeDiscoveryFixture(t, root, "packages/app/index.ts", "export {}\n")
	appDir := tspath.CombinePaths(root, "packages/app")
	writeDiscoveryFixture(t, root, "packages/app/.gitignore", "ignored.ts\n")

	virtualFiles := make(map[string]string)
	realPaths := make(map[string]string)
	for _, drive := range []string{"C:", "c:"} {
		for _, name := range []string{"", "packages", "packages/app"} {
			realPaths[tspath.ResolvePath(drive+"/Repo", name)] = tspath.ResolvePath("c:/Repo", name)
		}
		for _, name := range []string{"rslint.config.js", "packages/app/rslint.config.js", "packages/app/rslint.config.mjs", "packages/app/index.ts", "packages/app/.gitignore"} {
			path := tspath.ResolvePath(drive+"/Repo", name)
			virtualFiles[path] = ""
			realPaths[path] = tspath.ResolvePath("c:/Repo", name)
			if name == "packages/app/.gitignore" {
				virtualFiles[path] = "ignored.ts\n"
			}
		}
	}
	windowsFS := &configDiscoveryRealpathFS{
		FS: utils.NewOverlayVFS(&configDiscoveryCaseSensitivityFS{
			FS: discoveryTestFS(), caseSensitive: false,
		}, virtualFiles),
		realPaths: realPaths,
	}

	for _, test := range []struct {
		name, root, literal, owner string
		directories                []string
		includesRoot               bool
		fs                         vfs.FS
	}{
		{name: "literal activates first", root: root, directories: []string{root}, literal: literal, owner: appDir, includesRoot: true, fs: discoveryTestFS()},
		{name: "automatic activates first", root: root, directories: []string{appDir}, literal: literal, owner: appDir, fs: discoveryTestFS()},
		{name: "same drive", root: "C:/Repo", directories: []string{"C:/Repo/packages/app"}, literal: "C:/Repo/packages/app/index.ts", owner: "C:/Repo/packages/app", fs: windowsFS},
		{name: "mixed drive automatic first", root: "C:/Repo", directories: []string{"C:/Repo/packages/app"}, literal: "c:/Repo/packages/app/index.ts", owner: "C:/Repo/packages/app", fs: windowsFS},
		{name: "mixed drive literal first", root: "C:/Repo", directories: []string{"C:/Repo"}, literal: "c:/Repo/packages/app/index.ts", owner: "c:/Repo/packages/app", includesRoot: true, fs: windowsFS},
		{name: "mixed drive directory roots", root: "C:/Repo", directories: []string{"C:/Repo", "c:/Repo/packages/app"}, literal: "C:/Repo/packages/app/index.ts", owner: "c:/Repo/packages/app", includesRoot: true, fs: windowsFS},
	} {
		t.Run(test.name, func(t *testing.T) {
			for iteration := range 5 {
				loader := newFixtureConfigLoader()
				rootConfig := tspath.CombinePaths(test.root, "rslint.config.js")
				loader.configs[rootConfig] = rslintconfig.RslintConfig{
					{Ignores: []string{"packages/app/rslint.config.js"}},
					{Name: "root", Rules: rslintconfig.Rules{}},
				}
				loader.plugins[rootConfig] = []rslintconfig.EslintPluginEntry{{Prefix: "root-plugin"}}
				for _, owner := range []string{test.root + "/packages/app", tspath.GetDirectoryPath(test.literal), test.owner} {
					ignoredJS := tspath.CombinePaths(owner, "rslint.config.js")
					automaticMJS := tspath.CombinePaths(owner, "rslint.config.mjs")
					loader.configs[ignoredJS] = namedConfig("literal-js")
					loader.configs[automaticMJS] = namedConfig("automatic-mjs")
					loader.plugins[ignoredJS] = []rslintconfig.EslintPluginEntry{{Prefix: "literal-js-plugin"}}
					loader.plugins[automaticMJS] = []rslintconfig.EslintPluginEntry{{Prefix: "automatic-mjs-plugin"}}
				}

				catalog, err := discoverTestAutomatic(context.Background(), test.fs, loader, ConfigDiscoveryRequest{
					CWD:         test.root,
					Directories: discoveryDirectoryInputs(test.directories),
					Files:       []DiscoveryFile{{PathIdentity: rslintconfig.PathIdentity{Path: test.literal}, Explicit: true}},
				})
				if err != nil {
					t.Fatal(err)
				}
				if got := namedConfigEntry(catalog.Configs[test.owner]).Name; got != "automatic-mjs" {
					t.Fatalf("iteration %d selected config = %q, want automatic .mjs; owners=%v", iteration, got, catalog.ConfigDirectories())
				}
				scope := catalog.Scopes[test.owner]
				if scope.ExplicitOnly || !reflect.DeepEqual(scope.ExplicitFiles, []string{test.literal}) {
					t.Fatalf("iteration %d mixed scope = %+v", iteration, scope)
				}
				if !catalogTestConfig(t, catalog, test.owner).IsFileIgnored(test.owner+"/ignored.ts", test.owner) {
					t.Fatal("merged owner lost its collected Git ignore")
				}
				wantPlugins := []string{"automatic-mjs-plugin"}
				if test.includesRoot {
					wantPlugins = append(wantPlugins, "root-plugin")
				}
				if got := pluginPrefixes(catalog.EslintPlugins); !reflect.DeepEqual(got, wantPlugins) {
					t.Fatalf("iteration %d effective plugins = %v, want %v", iteration, got, wantPlugins)
				}
				if len(catalog.Configs) != len(wantPlugins) || len(catalog.Scopes) != len(wantPlugins) {
					t.Fatalf("one directory produced multiple owners: configs=%v scopes=%v", catalog.ConfigDirectories(), catalog.Scopes)
				}

				var wantIDs []string
				for _, batch := range loader.batches {
					for _, candidate := range batch.Candidates {
						if rslintconfig.ExactPathID(candidate.ConfigPath) == rslintconfig.ExactPathID(rootConfig) {
							if test.includesRoot {
								wantIDs = append(wantIDs, candidate.ID)
							}
							continue
						}
						if candidate.ConfigDirectory != test.owner {
							t.Fatalf("candidate routing owner=%q, catalog owner=%q", candidate.ConfigDirectory, test.owner)
						}
						if tspath.GetBaseFileName(candidate.ConfigPath) == "rslint.config.mjs" {
							wantIDs = append(wantIDs, candidate.ID)
						}
					}
				}
				sort.Strings(wantIDs)
				if !reflect.DeepEqual(catalog.EffectiveConfigIDs, wantIDs) {
					t.Fatalf("iteration %d effective IDs = %v, want %v", iteration, catalog.EffectiveConfigIDs, wantIDs)
				}
			}
		})
	}
}

func TestConfigDiscoveryDirectoryTargetCannotBypassAncestorGlobalIgnore(t *testing.T) {
	root := discoveryTempDir(t)
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
		Directories: discoveryDirectoryInputs([]string{filepath.Join(root, "ignored/deep")}),
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
				return ConfigDiscoveryRequest{CWD: root, Directories: discoveryDirectoryInputs([]string{target})}
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
					Directories:               discoveryDirectoryInputs([]string{target}),
					LimitDirectoryWalkToFiles: true,
					Files:                     []DiscoveryFile{{PathIdentity: rslintconfig.PathIdentity{Path: tspath.CombinePaths(target, "index.ts")}, Explicit: false}},
				}
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := discoveryTempDir(t)
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
			if catalog.Stats.DirectoriesPruned != 1 || catalog.Stats.DirectoriesVisited != 1 {
				t.Fatalf("Git pruning changed the supplied-root ownership exception: %+v", catalog.Stats)
			}
			if catalog.Scopes[nearestDir].ExplicitOnly {
				t.Fatal("directory ancestry owner must remain available to automatic targets")
			}
		})
	}
}

func TestConfigDiscoveryDefaultExcludedDirectoryRootOnlyLoadsReachableAncestor(t *testing.T) {
	root := discoveryTempDir(t)
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
		Directories: discoveryDirectoryInputs([]string{target}),
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
		root := discoveryTempDir(t)
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
			Directories:               discoveryDirectoryInputs([]string{rootA, rootB, rootWithoutTarget}),
			LimitDirectoryWalkToFiles: true,
			Files: []DiscoveryFile{
				{PathIdentity: rslintconfig.PathIdentity{Path: aTarget}, Explicit: false},
				{PathIdentity: rslintconfig.PathIdentity{Path: bTarget}, Explicit: false},
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
		root := discoveryTempDir(t)
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
			Directories:               discoveryDirectoryInputs([]string{root}),
			LimitDirectoryWalkToFiles: true,
			Files:                     []DiscoveryFile{{PathIdentity: rslintconfig.PathIdentity{Path: target}, Explicit: false}},
		})
		if got := catalog.ConfigDirectories(); !reflect.DeepEqual(got, []string{tspath.NormalizePath(root)}) {
			t.Fatalf("parent-ignore catalog = %v", got)
		}
		if got := requestedConfigPaths(loader); !reflect.DeepEqual(got, []string{rootConfig}) {
			t.Fatalf("parent-ignore branch evaluated nested config: %v", got)
		}
	})

	t.Run("gitignore prunes a target branch before its config", func(t *testing.T) {
		root := discoveryTempDir(t)
		writeDiscoveryFixture(t, root, ".gitignore", "target/deep/\n")
		loader := newFixtureConfigLoader()
		rootConfig := writeConfigCandidate(t, root, "rslint.config.js")
		nestedConfig := writeConfigCandidate(t, root, "target/deep/rslint.config.js")
		target := writeDiscoveryFixture(t, root, "target/deep/index.ts", "export {}\n")
		loader.configs[rootConfig] = namedConfig("root")
		loader.configs[nestedConfig] = namedConfig("nested")

		catalog := buildFixtureCatalog(t, root, loader, ConfigDiscoveryRequest{
			CWD:                       root,
			Directories:               discoveryDirectoryInputs([]string{root}),
			LimitDirectoryWalkToFiles: true,
			Files:                     []DiscoveryFile{{PathIdentity: rslintconfig.PathIdentity{Path: target}, Explicit: false}},
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
		root := discoveryTempDir(t)
		loader := newFixtureConfigLoader()
		rootConfig := writeConfigCandidate(t, root, "rslint.config.js")
		nestedConfig := writeConfigCandidate(t, root, "unrelated/deep/rslint.config.js")
		loader.configs[rootConfig] = namedConfig("root")
		loader.configs[nestedConfig] = namedConfig("nested")

		catalog := buildFixtureCatalog(t, root, loader, ConfigDiscoveryRequest{
			CWD:                       root,
			Directories:               discoveryDirectoryInputs([]string{root}),
			LimitDirectoryWalkToFiles: true,
		})
		wantDirs := []string{tspath.NormalizePath(root), tspath.CombinePaths(root, "unrelated/deep")}
		if got := catalog.ConfigDirectories(); !reflect.DeepEqual(got, wantDirs) {
			t.Fatalf("directory-only catalog = %v, want %v", got, wantDirs)
		}
	})

	t.Run("mixed CLI-style files and directories remain unbounded without the API flag", func(t *testing.T) {
		root := discoveryTempDir(t)
		loader := newFixtureConfigLoader()
		rootConfig := writeConfigCandidate(t, root, "rslint.config.js")
		nestedConfig := writeConfigCandidate(t, root, "unrelated/deep/rslint.config.js")
		literal := writeDiscoveryFixture(t, root, "index.ts", "export {}\n")
		loader.configs[rootConfig] = namedConfig("root")
		loader.configs[nestedConfig] = namedConfig("nested")

		catalog := buildFixtureCatalog(t, root, loader, ConfigDiscoveryRequest{
			CWD:         root,
			Directories: discoveryDirectoryInputs([]string{root}),
			Files:       []DiscoveryFile{{PathIdentity: rslintconfig.PathIdentity{Path: literal}, Explicit: true}},
		})
		wantDirs := []string{tspath.NormalizePath(root), tspath.CombinePaths(root, "unrelated/deep")}
		if got := catalog.ConfigDirectories(); !reflect.DeepEqual(got, wantDirs) {
			t.Fatalf("mixed CLI-style catalog = %v, want %v", got, wantDirs)
		}
	})
}

func TestConfigDiscoveryAutomaticDirectoryTargetKeepsAncestorOwnerGitChain(t *testing.T) {
	repositoryRoot := discoveryTempDir(t)
	cwd := filepath.Join(repositoryRoot, "packages/app")
	selectedDir := filepath.Join(cwd, "selected")
	for relativePath, content := range map[string]string{
		".gitignore":                                 "root.generated.ts\n",
		"packages/.gitignore":                        "package.generated.ts\n",
		"packages/app/.gitignore":                    "app.generated.ts\n",
		"packages/app/selected/.gitignore":           "local.ts\n",
		"packages/app/selected/deep/.gitignore":      "deep.ts\n",
		"packages/app/selected/root.generated.ts":    "export {};\n",
		"packages/app/selected/package.generated.ts": "export {};\n",
		"packages/app/selected/app.generated.ts":     "export {};\n",
		"packages/app/selected/local.ts":             "export {};\n",
		"packages/app/selected/deep/deep.ts":         "export {};\n",
		"packages/app/unrelated/.gitignore":          "index.ts\n",
		"packages/app/unrelated/index.ts":            "export {};\n",
	} {
		writeDiscoveryFixture(t, repositoryRoot, relativePath, content)
	}
	configPath := writeConfigCandidate(t, repositoryRoot, "rslint.config.js")
	entries := namedConfig("ancestor")
	loader := newFixtureConfigLoader()
	loader.configs[configPath] = entries
	spy := &configDiscoveryReadSpyFS{FS: discoveryTestFS()}

	catalog, err := discoverTestAutomatic(
		context.Background(),
		spy,
		loader,
		ConfigDiscoveryRequest{
			CWD:            cwd,
			Directories:    discoveryDirectoryInputs([]string{selectedDir}),
			SingleThreaded: true,
		},
	)
	if err != nil {
		t.Fatalf("DiscoverAutomatic: %v", err)
	}
	repositoryRoot = tspath.NormalizePath(repositoryRoot)
	cwd = tspath.NormalizePath(cwd)
	selectedDir = tspath.NormalizePath(selectedDir)
	if got := catalog.ConfigDirectories(); !reflect.DeepEqual(got, []string{repositoryRoot}) {
		t.Fatalf("automatic owner = %v, want ancestor %q", got, repositoryRoot)
	}
	if got := requestedConfigPaths(loader); !reflect.DeepEqual(got, []string{configPath}) {
		t.Fatalf("automatic config requests = %v, want %q", got, configPath)
	}

	projected := catalogTestConfig(t, catalog, repositoryRoot)
	full := collectedTestConfig(t, entries, repositoryRoot, discoveryTestFS(), nil, nil)
	for _, test := range []struct {
		relativePath string
		ignored      bool
	}{
		{relativePath: "packages/app/selected/root.generated.ts", ignored: true},
		{relativePath: "packages/app/selected/package.generated.ts", ignored: true},
		{relativePath: "packages/app/selected/app.generated.ts", ignored: true},
		{relativePath: "packages/app/selected/local.ts", ignored: true},
		{relativePath: "packages/app/selected/deep/deep.ts", ignored: true},
	} {
		path := tspath.ResolvePath(repositoryRoot, test.relativePath)
		if got := full.IsFileIgnored(path, repositoryRoot); got != test.ignored {
			t.Fatalf("invalid full-projection fixture for %s: ignored = %t, want %t", test.relativePath, got, test.ignored)
		}
		if got := projected.IsFileIgnored(path, repositoryRoot); got != test.ignored {
			t.Errorf("%s ignored by targeted projection = %t, want %t", test.relativePath, got, test.ignored)
		}
	}

	wantGitignoreReads := []string{
		tspath.ResolvePath(repositoryRoot, ".gitignore"),
		tspath.ResolvePath(repositoryRoot, "packages/.gitignore"),
		tspath.ResolvePath(cwd, ".gitignore"),
		tspath.ResolvePath(selectedDir, ".gitignore"),
		tspath.ResolvePath(selectedDir, "deep/.gitignore"),
	}
	if got := spy.gitignorePaths; !reflect.DeepEqual(got, wantGitignoreReads) {
		t.Fatalf("automatic ancestor-owner Git reads = %v, want %v", got, wantGitignoreReads)
	}
	unrelatedDir := tspath.ResolvePath(cwd, "unrelated")
	for _, accessed := range spy.accessedDirs {
		if accessed == unrelatedDir || tspath.StartsWithDirectory(accessed, unrelatedDir, true) {
			t.Fatalf("automatic target walk entered unrelated directory %q", accessed)
		}
	}
}

func TestConfigDiscoveryExplicitConfigLoadsBeforeProjectingGitignore(t *testing.T) {
	root := discoveryTempDir(t)
	ignoredTarget := writeDiscoveryFixture(t, root, "config/index.ts", "export {};\n")
	visibleTarget := writeDiscoveryFixture(t, root, "config/visible.ts", "export {};\n")
	writeDiscoveryFixture(t, root, ".gitignore", "config/custom.config.cjs\nconfig/index.ts\n")
	loader := newFixtureConfigLoader()
	configPath := writeConfigCandidate(t, root, "config/custom.config.cjs")
	loader.configs[configPath] = rslintconfig.RslintConfig{{
		Name:  "explicit",
		Files: []string{"*.ts"},
		Rules: rslintconfig.Rules{},
	}}

	catalog, err := loadTestExplicit(
		context.Background(),
		discoveryTestFS(),
		loader,
		ExplicitConfigRequest{CWD: root, ConfigPath: configPath},
	)
	if err != nil {
		t.Fatalf("LoadExplicitConfig: %v", err)
	}
	configDir := tspath.GetDirectoryPath(tspath.NormalizePath(configPath))
	if got := catalog.ConfigDirectories(); !reflect.DeepEqual(got, []string{configDir}) {
		t.Fatalf("explicit config should be anchored at its directory: %v", got)
	}
	if !catalog.Explicit {
		t.Fatal("explicit discovery catalog was not marked invocation-wide")
	}
	if got := requestedConfigPaths(loader); !reflect.DeepEqual(got, []string{configPath}) {
		t.Fatalf("explicit source = %v, want %q", got, configPath)
	}
	if got := loader.batches[0].Candidates[0].ConfigDirectory; got != configDir {
		t.Fatalf("wire configDirectory = %q, want config directory %q", got, configDir)
	}
	config := catalogTestConfig(t, catalog, configDir)
	if got := config.GetConfigForFile(ignoredTarget, configDir); got != nil {
		t.Fatalf("Git-ignored target received explicit config: %+v", got)
	}
	if got := config.GetConfigForFile(visibleTarget, configDir); got == nil {
		t.Fatal("visible target did not receive explicit config")
	}
}

func TestConfigDiscoveryExplicitConfigRejectsNonModulePathsBeforeLoading(t *testing.T) {
	root := discoveryTempDir(t)
	for _, name := range []string{"rslint.json", "rslint.jsonc", "rslint.yaml"} {
		t.Run(name, func(t *testing.T) {
			loader := newFixtureConfigLoader()
			configPath := writeDiscoveryFixture(t, root, name, `[{"rules":{"no-debugger":"error"}}]`)
			_, err := loadTestExplicit(
				context.Background(),
				discoveryTestFS(),
				loader,
				ExplicitConfigRequest{CWD: root, ConfigPath: configPath},
			)
			if err == nil || !strings.Contains(err.Error(), "JS/TS config module") {
				t.Fatalf("LoadExplicitConfig(%q) error = %v", configPath, err)
			}
			if len(loader.batches) != 0 {
				t.Fatalf("unsupported config reached module loader: %+v", loader.batches)
			}
		})
	}
}

func TestConfigDiscoveryExplicitConfigDoesNotDiscoverNestedCandidates(t *testing.T) {
	root := discoveryTempDir(t)
	loader := newFixtureConfigLoader()
	configPath := writeConfigCandidate(t, root, "custom.config.cjs")
	nestedConfigPath := writeConfigCandidate(t, root, "nested/rslint.config.js")
	loader.configs[configPath] = namedConfig("explicit")
	loader.configs[nestedConfigPath] = namedConfig("must-not-load")
	writeDiscoveryFixture(t, root, "nested/.gitignore", "generated.ts\n")
	ignoredTarget := writeDiscoveryFixture(t, root, "nested/generated.ts", "export {};\n")

	catalog, err := loadTestExplicit(
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
	if got := catalogTestConfig(t, catalog, root).GetConfigForFile(ignoredTarget, root); got != nil {
		t.Fatalf("nested Git-ignored target received explicit config: %+v", got)
	}
}

func TestConfigDiscoveryExplicitConfigTargetFileProjection(t *testing.T) {
	root := discoveryTempDir(t)
	loader := newFixtureConfigLoader()
	configPath := writeConfigCandidate(t, root, "custom.config.cjs")
	loader.configs[configPath] = namedConfig("explicit")
	target := writeDiscoveryFixture(t, root, "src/index.ts", "export {};\n")
	writeDiscoveryFixture(t, root, "src/.gitignore", "index.ts\n")
	writeDiscoveryFixture(t, root, "unrelated/.gitignore", "other.ts\n")

	spyFS := &configDiscoveryReadSpyFS{FS: discoveryTestFS()}
	catalog, err := loadTestExplicit(
		context.Background(),
		spyFS,
		loader,
		ExplicitConfigRequest{
			CWD:         root,
			ConfigPath:  configPath,
			TargetFiles: []DiscoveryFile{{PathIdentity: rslintconfig.PathIdentity{Path: target}}},
		},
	)
	if err != nil {
		t.Fatalf("LoadExplicitConfig: %v", err)
	}
	root = tspath.NormalizePath(root)
	if got := catalogTestConfig(t, catalog, root).GetConfigForFile(target, root); got != nil {
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
	physicalRoot := discoveryTempDir(t)
	writeConfigCandidate(t, physicalRoot, "custom.config.cjs")
	writeDiscoveryFixture(t, physicalRoot, ".gitignore", "src/index.ts\n")
	physicalTarget := writeDiscoveryFixture(t, physicalRoot, "src/index.ts", "export {};\n")
	aliasRoot := filepath.Join(discoveryTempDir(t), "workspace-alias")
	if err := os.Symlink(physicalRoot, aliasRoot); err != nil {
		t.Skipf("directory symlink unavailable: %v", err)
	}
	aliasRoot = tspath.NormalizePath(aliasRoot)
	aliasTarget := tspath.CombinePaths(aliasRoot, "src/index.ts")
	externalRoot := discoveryTempDir(t)
	if err := os.Mkdir(filepath.Join(externalRoot, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	externalSpelling := writeDiscoveryFixture(t, externalRoot, "index.ts", "export {};\n")
	canonicalTarget := discoveryTestFS().Realpath(physicalTarget)

	for _, test := range []struct {
		name        string
		cwd         string
		target      string
		canonical   string
		wantIgnored bool
	}{
		{name: "aliased CWD and physical target", cwd: aliasRoot, target: physicalTarget, wantIgnored: true},
		{name: "physical CWD and aliased target", cwd: physicalRoot, target: aliasTarget, wantIgnored: true},
		{
			name:      "canonical target hint does not move an external lexical spelling into the Git root",
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

			catalog, err := loadTestExplicit(
				context.Background(),
				spyFS,
				loader,
				ExplicitConfigRequest{
					CWD:        test.cwd,
					ConfigPath: configPath,
					TargetFiles: []DiscoveryFile{{PathIdentity: rslintconfig.PathIdentity{Path: test.target,
						CanonicalPath: test.canonical},
					}},
				},
			)
			if err != nil {
				t.Fatalf("LoadExplicitConfig: %v", err)
			}
			cwd := tspath.NormalizePath(test.cwd)
			matchFile, matchDir := rslintconfig.ResolveConfigFilePathSpaceWithCanonical(
				test.target,
				test.canonical,
				cwd,
				spyFS,
			)
			ignored := catalogTestConfig(t, catalog, cwd).GetConfigForFile(matchFile, matchDir) == nil
			if ignored != test.wantIgnored {
				t.Fatalf(
					"path-space Git ignore = %t, want %t: config=%+v match=(%q,%q) reads=%v",
					ignored,
					test.wantIgnored,
					catalog.Configs[cwd],
					matchFile,
					matchDir,
					spyFS.gitignorePaths,
				)
			}
			wantReads := []string{tspath.CombinePaths(externalRoot, ".gitignore")}
			if test.wantIgnored {
				wantReads = []string{
					tspath.CombinePaths(physicalRoot, ".gitignore"),
					tspath.CombinePaths(physicalRoot, "src/.gitignore"),
				}
			}
			if got := spyFS.gitignorePaths; !reflect.DeepEqual(got, wantReads) {
				t.Fatalf("path-space Git reads = %v, want %v", got, wantReads)
			}
		})
	}
}

func TestConfigDiscoveryExplicitConfigEmptyTargetSetSkipsGitProjection(t *testing.T) {
	root := discoveryTempDir(t)
	loader := newFixtureConfigLoader()
	configPath := writeConfigCandidate(t, root, "custom.config.cjs")
	loader.configs[configPath] = namedConfig("explicit")
	writeDiscoveryFixture(t, root, ".gitignore", "*.ts\n")

	spyFS := &configDiscoveryReadSpyFS{FS: discoveryTestFS()}
	catalog, err := loadTestExplicit(
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
			root := discoveryTempDir(t)
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
					requestTargets = append(requestTargets, DiscoveryFile{PathIdentity: rslintconfig.PathIdentity{Path: path}})
					standaloneTargets = append(standaloneTargets, path)
				}
			}
			fsys := discoveryTestFS()
			catalog, err := loadTestExplicit(
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
			projected := catalogTestConfig(t, catalog, root)
			standalone := collectedTestConfig(t, entries, root, fsys, standaloneTargets, nil)
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

func TestConfigDiscoveryExplicitTargetsSkipUnrelatedGitignoreSubtrees(t *testing.T) {
	root := discoveryTempDir(t)
	for relativePath, content := range map[string]string{
		".gitignore":                                "*.generated.ts\n",
		"packages/selected/.gitignore":              "local.ts\n!keep.generated.ts\n",
		"packages/selected/index.ts":                "export {};\n",
		"packages/selected/local.ts":                "export {};\n",
		"packages/selected/keep.generated.ts":       "export {};\n",
		"packages/selected/deep/.gitignore":         "drop.ts\n",
		"packages/selected/deep/drop.ts":            "export {};\n",
		"packages/selected/deep/other.generated.ts": "export {};\n",
		"packages/unrelated/.gitignore":             "index.ts\n",
		"packages/unrelated/index.ts":               "export {};\n",
		"tools/.gitignore":                          "exact.ts\n",
		"tools/exact.ts":                            "export {};\n",
	} {
		writeDiscoveryFixture(t, root, relativePath, content)
	}
	configPath := writeConfigCandidate(t, root, "custom.config.cjs")
	selectedConfig := writeConfigCandidate(t, root, "packages/selected/rslint.config.js")
	unrelatedConfig := writeConfigCandidate(t, root, "packages/unrelated/rslint.config.js")
	entries := namedConfig("explicit")
	loader := newFixtureConfigLoader()
	loader.configs[configPath] = entries
	loader.configs[selectedConfig] = namedConfig("must-not-load-selected")
	loader.configs[unrelatedConfig] = namedConfig("must-not-load-unrelated")
	selectedDir := tspath.NormalizePath(filepath.Join(root, "packages/selected"))
	exactFile := tspath.NormalizePath(filepath.Join(root, "tools/exact.ts"))
	spy := &configDiscoveryReadSpyFS{FS: discoveryTestFS()}

	catalog, err := loadTestExplicit(
		context.Background(),
		spy,
		loader,
		ExplicitConfigRequest{
			CWD:               root,
			ConfigPath:        configPath,
			TargetFiles:       []DiscoveryFile{{PathIdentity: rslintconfig.PathIdentity{Path: exactFile}}},
			TargetDirectories: discoveryDirectoryInputs([]string{selectedDir}),
		},
	)
	if err != nil {
		t.Fatalf("LoadExplicitConfig: %v", err)
	}
	root = tspath.NormalizePath(root)
	projected := catalogTestConfig(t, catalog, root)
	full := collectedTestConfig(t, entries, root, discoveryTestFS(), nil, nil)
	for _, relativePath := range []string{
		"packages/selected/index.ts",
		"packages/selected/local.ts",
		"packages/selected/keep.generated.ts",
		"packages/selected/deep/drop.ts",
		"packages/selected/deep/other.generated.ts",
		"tools/exact.ts",
	} {
		path := tspath.ResolvePath(root, relativePath)
		if got, want := projected.IsFileIgnored(path, root), full.IsFileIgnored(path, root); got != want {
			t.Errorf("%s ignored by targeted projection = %t, full projection = %t", relativePath, got, want)
		}
	}
	if got := requestedConfigPaths(loader); !reflect.DeepEqual(got, []string{configPath}) {
		t.Fatalf("explicit projection evaluated nested configs: %v", got)
	}

	wantGitignoreReads := []string{
		tspath.ResolvePath(root, ".gitignore"),
		tspath.ResolvePath(root, "packages/.gitignore"),
		tspath.ResolvePath(root, "packages/selected/.gitignore"),
		tspath.ResolvePath(root, "packages/selected/deep/.gitignore"),
		tspath.ResolvePath(root, "tools/.gitignore"),
	}
	gotGitignoreReads := append([]string(nil), spy.gitignorePaths...)
	sort.Strings(gotGitignoreReads)
	sort.Strings(wantGitignoreReads)
	if !reflect.DeepEqual(gotGitignoreReads, wantGitignoreReads) {
		t.Fatalf("targeted Git reads = %v, want %v", gotGitignoreReads, wantGitignoreReads)
	}
	unrelatedDir := tspath.ResolvePath(root, "packages/unrelated")
	for _, accessed := range spy.accessedDirs {
		if accessed == unrelatedDir || tspath.StartsWithDirectory(accessed, unrelatedDir, true) {
			t.Fatalf("explicit projection entered unrelated directory %q", accessed)
		}
	}
}

func TestConfigDiscoveryExplicitDirectoryUsesConfigBaseAndTargetGitRoot(t *testing.T) {
	repositoryRoot := discoveryTempDir(t)
	cwd := filepath.Join(repositoryRoot, "packages/app")
	selectedDir := filepath.Join(cwd, "selected")
	for relativePath, content := range map[string]string{
		".gitignore":                              "packages/app/selected/parent-only.ts\n",
		"packages/.gitignore":                     "app/selected/package-only.ts\n",
		"packages/app/.gitignore":                 "*.generated.ts\n",
		"packages/app/selected/.gitignore":        "local.ts\n",
		"packages/app/selected/deep/.gitignore":   "deep.ts\n",
		"packages/app/selected/index.ts":          "export {};\n",
		"packages/app/selected/local.ts":          "export {};\n",
		"packages/app/selected/drop.generated.ts": "export {};\n",
		"packages/app/selected/parent-only.ts":    "export {};\n",
		"packages/app/selected/package-only.ts":   "export {};\n",
		"packages/app/selected/deep/deep.ts":      "export {};\n",
		"packages/app/unrelated/.gitignore":       "index.ts\n",
		"packages/app/unrelated/index.ts":         "export {};\n",
	} {
		writeDiscoveryFixture(t, repositoryRoot, relativePath, content)
	}
	configPath := writeConfigCandidate(t, repositoryRoot, "custom.config.cjs")
	entries := namedConfig("explicit")
	loader := newFixtureConfigLoader()
	loader.configs[configPath] = entries
	spy := &configDiscoveryReadSpyFS{FS: discoveryTestFS()}

	catalog, err := loadTestExplicit(
		context.Background(),
		spy,
		loader,
		ExplicitConfigRequest{
			CWD:               cwd,
			ConfigPath:        configPath,
			TargetDirectories: discoveryDirectoryInputs([]string{selectedDir}),
			SingleThreaded:    true,
		},
	)
	if err != nil {
		t.Fatalf("LoadExplicitConfig: %v", err)
	}
	repositoryRoot = tspath.NormalizePath(repositoryRoot)
	cwd = tspath.NormalizePath(cwd)
	selectedDir = tspath.NormalizePath(selectedDir)
	if got := catalog.ConfigDirectories(); !reflect.DeepEqual(got, []string{repositoryRoot}) {
		t.Fatalf("explicit owner = %v, want config directory %q", got, repositoryRoot)
	}
	if got := loader.batches[0].Candidates[0].ConfigDirectory; got != repositoryRoot {
		t.Fatalf("wire configDirectory = %q, want config directory %q", got, repositoryRoot)
	}

	projected := catalogTestConfig(t, catalog, repositoryRoot)
	full := collectedTestConfig(t, entries, repositoryRoot, discoveryTestFS(), nil, []string{selectedDir})
	for _, test := range []struct {
		relativePath string
		ignored      bool
	}{
		{relativePath: "selected/index.ts", ignored: false},
		{relativePath: "selected/local.ts", ignored: true},
		{relativePath: "selected/drop.generated.ts", ignored: true},
		{relativePath: "selected/parent-only.ts", ignored: true},
		{relativePath: "selected/package-only.ts", ignored: true},
		{relativePath: "selected/deep/deep.ts", ignored: true},
	} {
		path := tspath.ResolvePath(cwd, test.relativePath)
		if got := full.IsFileIgnored(path, repositoryRoot); got != test.ignored {
			t.Fatalf("invalid invocation-root fixture for %s: ignored = %t, want %t", test.relativePath, got, test.ignored)
		}
		if got := projected.IsFileIgnored(path, repositoryRoot); got != test.ignored {
			t.Errorf("%s ignored by targeted projection = %t, want %t", test.relativePath, got, test.ignored)
		}
	}

	wantGitignoreReads := []string{
		tspath.ResolvePath(repositoryRoot, ".gitignore"),
		tspath.ResolvePath(repositoryRoot, "packages/.gitignore"),
		tspath.ResolvePath(cwd, ".gitignore"),
		tspath.ResolvePath(selectedDir, ".gitignore"),
		tspath.ResolvePath(selectedDir, "deep/.gitignore"),
	}
	if got := spy.gitignorePaths; !reflect.DeepEqual(got, wantGitignoreReads) {
		t.Fatalf("explicit invocation-rooted Git reads = %v, want %v", got, wantGitignoreReads)
	}
	unrelatedDir := tspath.ResolvePath(cwd, "unrelated")
	for _, accessed := range spy.accessedDirs {
		if accessed == unrelatedDir || tspath.StartsWithDirectory(accessed, unrelatedDir, true) {
			t.Fatalf("explicit target walk entered unrelated directory %q", accessed)
		}
	}
}

func TestConfigDiscoveryExplicitConfigGitignoreMatchesPhysicalInvocationTarget(t *testing.T) {
	configRoot := discoveryTempDir(t)
	physicalWorkspace := discoveryTempDir(t)
	aliasWorkspace := filepath.Join(discoveryTempDir(t), "workspace-alias")
	if err := os.Symlink(physicalWorkspace, aliasWorkspace); err != nil {
		t.Skipf("directory symlink unavailable: %v", err)
	}
	physicalIgnored := filepath.Join(physicalWorkspace, "ignored.js")
	aliasIgnored := filepath.Join(aliasWorkspace, "ignored.js")
	configVisible := filepath.Join(configRoot, "ignored.js")
	for _, filePath := range []string{physicalIgnored, configVisible} {
		if err := os.WriteFile(filePath, []byte("debugger;\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(aliasWorkspace, ".gitignore"), []byte("ignored.js\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	physicalIgnored = tspath.NormalizePath(discoveryTestFS().Realpath(physicalIgnored))
	configPath := writeConfigCandidate(t, configRoot, "custom.config.cjs")
	loader := newFixtureConfigLoader()
	loader.configs[configPath] = namedConfig("explicit")

	catalog, err := loadTestExplicit(
		context.Background(),
		discoveryTestFS(),
		loader,
		ExplicitConfigRequest{
			CWD:         aliasWorkspace,
			ConfigPath:  configPath,
			TargetFiles: []DiscoveryFile{{PathIdentity: rslintconfig.PathIdentity{Path: physicalIgnored}, Explicit: true}},
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	configRoot = tspath.NormalizePath(configRoot)
	projected := catalogTestConfig(t, catalog, configRoot)
	for _, filePath := range []string{physicalIgnored, aliasIgnored} {
		if !projected.IsFileIgnored(tspath.NormalizePath(filePath), configRoot) {
			t.Fatalf("projected Git ignore missed invocation-root identity %q", filePath)
		}
	}
	if projected.IsFileIgnored(tspath.NormalizePath(configVisible), configRoot) {
		t.Fatalf("projected Git ignore escaped invocation root onto %q", configVisible)
	}
}

func TestConfigDiscoveryExplicitDirectoriesKeepGitScopesIndependent(t *testing.T) {
	configRoot := discoveryTempDir(t)
	invocationRoot := discoveryTempDir(t)
	physicalRoot := discoveryTempDir(t)
	aliasDirectory := filepath.Join(invocationRoot, "link")
	if err := os.Symlink(physicalRoot, aliasDirectory); err != nil {
		t.Skipf("directory symlink unavailable: %v", err)
	}
	physicalIgnored := writeDiscoveryFixture(t, physicalRoot, "ignored.ts", "debugger;\n")
	writeDiscoveryFixture(t, physicalRoot, ".gitignore", "ignored.ts\n")

	configPath := writeConfigCandidate(t, configRoot, "custom.config.cjs")
	loader := newFixtureConfigLoader()
	loader.configs[configPath] = namedConfig("explicit")
	catalog, err := loadTestExplicit(
		context.Background(),
		discoveryTestFS(),
		loader,
		ExplicitConfigRequest{
			CWD:               invocationRoot,
			ConfigPath:        configPath,
			TargetDirectories: discoveryDirectoryInputs([]string{invocationRoot, physicalRoot}),
			SingleThreaded:    true,
		},
	)
	if err != nil {
		t.Fatalf("LoadExplicitConfig: %v", err)
	}

	configRoot = tspath.NormalizePath(configRoot)
	physicalIgnored = tspath.NormalizePath(physicalIgnored)
	aliasIgnored := tspath.NormalizePath(filepath.Join(aliasDirectory, "ignored.ts"))
	matcher := catalogTestConfig(t, catalog, configRoot)
	if !matcher.IsFileIgnored(physicalIgnored, configRoot) {
		t.Fatal("physical directory did not apply its Git policy")
	}
	if !matcher.IsFileIgnored(aliasIgnored, configRoot) {
		t.Fatal("the same physical directory entry changed through an alias")
	}
}

func TestConfigDiscoveryExplicitCWDDirectoryProjectsFullGitTree(t *testing.T) {
	root := discoveryTempDir(t)
	configPath := writeConfigCandidate(t, root, "custom.config.cjs")
	target := writeDiscoveryFixture(t, root, "nested/deep/index.ts", "export {};\n")
	writeDiscoveryFixture(t, root, "nested/deep/.gitignore", "index.ts\n")
	loader := newFixtureConfigLoader()
	loader.configs[configPath] = namedConfig("explicit")

	catalog, err := loadTestExplicit(
		context.Background(),
		discoveryTestFS(),
		loader,
		ExplicitConfigRequest{
			CWD:               root,
			ConfigPath:        configPath,
			TargetDirectories: discoveryDirectoryInputs([]string{root}),
		},
	)
	if err != nil {
		t.Fatalf("LoadExplicitConfig: %v", err)
	}
	root = tspath.NormalizePath(root)
	if got := catalogTestConfig(t, catalog, root).GetConfigForFile(target, root); got != nil {
		t.Fatalf("CWD target did not project nested Git ignore: %+v", got)
	}
}

func TestConfigDiscoveryExplicitDirectoryContainingAliasedCWDProjectsFullGitTree(t *testing.T) {
	physicalParent := discoveryTempDir(t)
	physicalRoot := filepath.Join(physicalParent, "workspace")
	writeConfigCandidate(t, physicalRoot, "custom.config.cjs")
	physicalTarget := writeDiscoveryFixture(t, physicalRoot, "nested/index.ts", "export {};\n")
	writeDiscoveryFixture(t, physicalRoot, "nested/.gitignore", "index.ts\n")
	aliasRoot := filepath.Join(discoveryTempDir(t), "workspace-alias")
	if err := os.Symlink(physicalRoot, aliasRoot); err != nil {
		t.Skipf("directory symlink unavailable: %v", err)
	}
	aliasRoot = tspath.NormalizePath(aliasRoot)
	configPath := tspath.CombinePaths(aliasRoot, "custom.config.cjs")
	loader := newFixtureConfigLoader()
	loader.configs[configPath] = namedConfig("explicit")

	catalog, err := loadTestExplicit(
		context.Background(),
		discoveryTestFS(),
		loader,
		ExplicitConfigRequest{
			CWD:               aliasRoot,
			ConfigPath:        configPath,
			TargetDirectories: discoveryDirectoryInputs([]string{physicalParent}),
		},
	)
	if err != nil {
		t.Fatalf("LoadExplicitConfig: %v", err)
	}
	matchFile, matchDir := rslintconfig.ResolveConfigFilePathSpace(
		physicalTarget,
		aliasRoot,
		discoveryTestFS(),
	)
	if got := catalogTestConfig(t, catalog, aliasRoot).GetConfigForFile(matchFile, matchDir); got != nil {
		t.Fatalf("containing directory did not project aliased CWD Git ignore: %+v", got)
	}
}

func TestConfigDiscoveryExplicitPhysicalAncestorCollectsSiblingGitignore(t *testing.T) {
	physicalParent := discoveryTempDir(t)
	physicalWorkspace := filepath.Join(physicalParent, "workspace")
	configPath := writeConfigCandidate(t, physicalWorkspace, "custom.config.cjs")
	sibling := writeDiscoveryFixture(t, physicalParent, "sibling.ts", "debugger;\n")
	writeDiscoveryFixture(t, physicalParent, ".gitignore", "sibling.ts\n")
	aliasRoot := filepath.Join(discoveryTempDir(t), "workspace-alias")
	if err := os.Symlink(physicalWorkspace, aliasRoot); err != nil {
		t.Skipf("directory symlink unavailable: %v", err)
	}
	aliasRoot = tspath.NormalizePath(aliasRoot)
	aliasConfigPath := tspath.CombinePaths(aliasRoot, filepath.Base(configPath))
	loader := newFixtureConfigLoader()
	loader.configs[aliasConfigPath] = namedConfig("explicit")

	catalog, err := loadTestExplicit(
		context.Background(),
		discoveryTestFS(),
		loader,
		ExplicitConfigRequest{
			CWD:               aliasRoot,
			ConfigPath:        aliasConfigPath,
			TargetDirectories: discoveryDirectoryInputs([]string{tspath.NormalizePath(physicalParent)}),
			SingleThreaded:    true,
		},
	)
	if err != nil {
		t.Fatalf("LoadExplicitConfig: %v", err)
	}
	matcher := catalogTestConfig(t, catalog, aliasRoot)
	if !matcher.IsFileIgnored(tspath.NormalizePath(sibling), aliasRoot) {
		t.Fatal("requested physical ancestor lost its sibling Git source")
	}
}

func TestConfigDiscoveryExplicitFileFindsConfigInsideGitignoredDirectory(t *testing.T) {
	root := discoveryTempDir(t)
	writeDiscoveryFixture(t, root, ".gitignore", "ignored/\n")
	loader := newFixtureConfigLoader()
	rootConfig := writeConfigCandidate(t, root, "rslint.config.js")
	ignoredConfig := writeConfigCandidate(t, root, "ignored/rslint.config.js")
	loader.configs[rootConfig] = namedConfig("root")
	loader.configs[ignoredConfig] = namedConfig("nearest")
	filePath := writeDiscoveryFixture(t, root, "ignored/index.ts", "export {}\n")

	catalog := buildFixtureCatalog(t, root, loader, ConfigDiscoveryRequest{
		CWD:   root,
		Files: []DiscoveryFile{{PathIdentity: rslintconfig.PathIdentity{Path: filePath}, Explicit: true}},
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
	if !reflect.DeepEqual(scope.ExplicitFiles, []string{filePath}) {
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
			root := discoveryTempDir(t)
			external := discoveryTempDir(t)
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
				Files: []DiscoveryFile{{PathIdentity: rslintconfig.PathIdentity{Path: lexicalFile,
					CanonicalPath: canonicalFile}, Explicit: true,
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
	root := discoveryTempDir(t)
	external := discoveryTempDir(t)
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
		CWD:   root,
		Files: []DiscoveryFile{{PathIdentity: rslintconfig.PathIdentity{Path: filePath}, Explicit: true}},
	})
	root = tspath.NormalizePath(root)
	if got := catalog.ConfigDirectories(); !reflect.DeepEqual(got, []string{root}) {
		t.Fatalf("lexical symlink target owner = %v, want root", got)
	}
	if got := catalog.Scopes[root].ExplicitFiles; !reflect.DeepEqual(got, []string{filePath}) {
		t.Fatalf("explicit target scope = %v, want %q", got, filePath)
	}
}

func TestConfigDiscoveryDefaultExcludedExplicitFileOnlyFindsReachableAncestor(t *testing.T) {
	root := discoveryTempDir(t)
	loader := newFixtureConfigLoader()
	rootConfig := writeConfigCandidate(t, root, "rslint.config.js")
	excludedConfig := writeConfigCandidate(t, root, "node_modules/pkg/rslint.config.js")
	loader.configs[rootConfig] = namedConfig("reachable-root")
	loader.configs[excludedConfig] = namedConfig("must-not-load")
	filePath := writeDiscoveryFixture(t, root, "node_modules/pkg/index.ts", "export {}\n")

	catalog := buildFixtureCatalog(t, root, loader, ConfigDiscoveryRequest{
		CWD:   root,
		Files: []DiscoveryFile{{PathIdentity: rslintconfig.PathIdentity{Path: filePath}, Explicit: true}},
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
	root := discoveryTempDir(t)
	loader := newFixtureConfigLoader()
	rootConfig := writeConfigCandidate(t, root, "rslint.config.js")
	brokenConfig := writeConfigCandidate(t, root, "packages/broken/rslint.config.js")
	filePath := writeDiscoveryFixture(t, root, "packages/broken/index.ts", "debugger;\n")
	loader.configs[rootConfig] = namedConfig("root")
	loader.failures[brokenConfig] = ConfigModuleError{Code: "ERR_MODULE", Message: "broken fixture"}

	catalog := buildFixtureCatalog(t, root, loader, ConfigDiscoveryRequest{
		CWD:   root,
		Files: []DiscoveryFile{{PathIdentity: rslintconfig.PathIdentity{Path: filePath}, Explicit: true}},
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
	if scope := catalog.Scopes[root]; !reflect.DeepEqual(scope.ExplicitFiles, []string{filePath}) {
		t.Fatalf("explicit file was not reassigned to ancestor: %+v", scope)
	}
	if !catalog.Scopes[root].ExplicitOnly {
		t.Fatal("explicit-only source was not propagated to target scope")
	}
}

func TestConfigDiscoveryFailedDirectoryAliasKeepsAncestorFallback(t *testing.T) {
	root := tspath.NormalizePath(discoveryTempDir(t))
	loader := newFixtureConfigLoader()
	rootConfig := writeConfigCandidate(t, root, "rslint.config.js")
	brokenConfig := writeConfigCandidate(t, root, "rslint.config.mjs")
	writeDiscoveryFixture(t, root, "index.ts", "export {};\n")
	alias := tspath.NormalizePath(filepath.Join(root, "alias"))
	if err := os.Symlink(root, alias); err != nil {
		t.Skipf("directory symlink unavailable: %v", err)
	}
	loader.configs[rootConfig] = rslintconfig.RslintConfig{
		{Ignores: []string{"**/rslint.config.js"}},
		{Name: "root", Rules: rslintconfig.Rules{}},
	}
	loader.failures[brokenConfig] = ConfigModuleError{Code: "ERR_MODULE", Message: "broken nearest module"}

	catalog, err := discoverTestAutomatic(context.Background(), discoveryTestFS(), loader, ConfigDiscoveryRequest{
		CWD:         root,
		Directories: discoveryDirectoryInputs([]string{alias}),
	})
	if err != nil {
		t.Fatalf("failed alias blocked its valid ancestor: %v", err)
	}
	if got := catalog.ConfigDirectories(); !reflect.DeepEqual(got, []string{root}) {
		t.Fatalf("fallback owners = %v, want %q", got, root)
	}
	failedAlias := tspath.CombinePaths(alias, "rslint.config.mjs")
	if len(catalog.Failures) != 1 || catalog.Failures[0].Path != failedAlias {
		t.Fatalf("failed alias metadata = %+v, want %q", catalog.Failures, failedAlias)
	}
	if got := requestedConfigPathsByBatch(loader); !reflect.DeepEqual(got, [][]string{{rootConfig}, {brokenConfig}}) {
		t.Fatalf("fallback module requests = %v", got)
	}
	if len(catalog.EffectiveConfigIDs) != 1 {
		t.Fatalf("effective candidates = %v, want only the successful ancestor", catalog.EffectiveConfigIDs)
	}
}

func TestConfigDiscoveryUsesCanonicalOwnerOnlyAfterLexicalAncestryIsEmpty(t *testing.T) {
	t.Run("canonical file fallback", func(t *testing.T) {
		lexicalRoot := discoveryTempDir(t)
		physicalRoot := discoveryTempDir(t)
		loader := newFixtureConfigLoader()
		configPath := writeConfigCandidate(t, physicalRoot, "rslint.config.js")
		loader.configs[configPath] = namedConfig("physical")
		lexicalFile := writeDiscoveryFixture(t, lexicalRoot, "pkg/index.ts", "export {}\n")
		canonicalFile := writeDiscoveryFixture(t, physicalRoot, "pkg/index.ts", "export {}\n")

		catalog := buildFixtureCatalog(t, lexicalRoot, loader, ConfigDiscoveryRequest{
			CWD: lexicalRoot,
			Files: []DiscoveryFile{{PathIdentity: rslintconfig.PathIdentity{Path: lexicalFile,
				CanonicalPath: canonicalFile}, Explicit: true,
			}},
		})
		physicalRoot = tspath.NormalizePath(physicalRoot)
		if got := catalog.ConfigDirectories(); !reflect.DeepEqual(got, []string{physicalRoot}) {
			t.Fatalf("canonical fallback owner = %v, want %q", got, physicalRoot)
		}
		if got := requestedConfigPaths(loader); !reflect.DeepEqual(got, []string{configPath}) {
			t.Fatalf("canonical fallback candidates = %v", got)
		}
		if scope := catalog.Scopes[physicalRoot]; !reflect.DeepEqual(scope.ExplicitFiles, []string{lexicalFile}) {
			t.Fatalf("canonical fallback discarded lexical target: %+v", scope)
		}
	})

	t.Run("lexical candidate wins", func(t *testing.T) {
		lexicalRoot := discoveryTempDir(t)
		physicalRoot := discoveryTempDir(t)
		loader := newFixtureConfigLoader()
		lexicalConfig := writeConfigCandidate(t, lexicalRoot, "rslint.config.js")
		physicalConfig := writeConfigCandidate(t, physicalRoot, "pkg/rslint.config.js")
		loader.configs[lexicalConfig] = namedConfig("lexical")
		loader.configs[physicalConfig] = namedConfig("physical")
		lexicalFile := writeDiscoveryFixture(t, lexicalRoot, "pkg/index.ts", "export {}\n")
		canonicalFile := writeDiscoveryFixture(t, physicalRoot, "pkg/index.ts", "export {}\n")

		catalog := buildFixtureCatalog(t, lexicalRoot, loader, ConfigDiscoveryRequest{
			CWD: lexicalRoot,
			Files: []DiscoveryFile{{PathIdentity: rslintconfig.PathIdentity{Path: lexicalFile,
				CanonicalPath: canonicalFile}, Explicit: true,
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
		lexicalRoot := discoveryTempDir(t)
		physicalRoot := discoveryTempDir(t)
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

		catalog, err := discoverTestAutomatic(context.Background(), fsys, loader, ConfigDiscoveryRequest{
			CWD:                       lexicalRoot,
			Directories:               discoveryDirectoryInputs([]string{lexicalRoot}),
			LimitDirectoryWalkToFiles: true,
			Files:                     []DiscoveryFile{{PathIdentity: rslintconfig.PathIdentity{Path: targetPath}, Explicit: false}},
		})
		if err != nil {
			t.Fatalf("Build: %v", err)
		}
		if got := catalog.ConfigDirectories(); !reflect.DeepEqual(got, []string{lexicalRoot}) {
			t.Fatalf("directory physical IO changed lexical owner = %v, want %q", got, lexicalRoot)
		}
		if got := requestedConfigPaths(loader); !reflect.DeepEqual(got, []string{configPath}) {
			t.Fatalf("directory canonical candidates = %v", got)
		}
		if got := fsys.realpathCallCount(lexicalRoot); got != 1 {
			t.Fatalf("lexical directory realpath calls = %d, want 1", got)
		}
	})

	t.Run("directory reads physical Git source independently of canonical owner", func(t *testing.T) {
		lexicalRoot := discoveryTempDir(t)
		physicalRoot := discoveryTempDir(t)
		loader := newFixtureConfigLoader()
		configPath := writeConfigCandidate(t, physicalRoot, "rslint.config.js")
		lexicalIgnored := writeDiscoveryFixture(t, lexicalRoot, "lexical-hidden/index.ts", "export {}\n")
		physicalOnly := writeDiscoveryFixture(t, lexicalRoot, "physical-hidden/index.ts", "export {}\n")
		lexicalGitignore := writeDiscoveryFixture(t, lexicalRoot, ".gitignore", "lexical-hidden/\n")
		physicalGitignore := writeDiscoveryFixture(t, physicalRoot, ".gitignore", "physical-hidden/\n")
		writeDiscoveryFixture(t, physicalRoot, "lexical-hidden/index.ts", "export {}\n")
		writeDiscoveryFixture(t, physicalRoot, "physical-hidden/index.ts", "export {}\n")
		loader.configs[configPath] = namedConfig("physical-directory")
		fsys := &configDiscoveryReadSpyFS{FS: &configDiscoveryRealpathFS{
			FS: discoveryTestFS(),
			realPaths: map[string]string{
				tspath.NormalizePath(lexicalRoot):                   tspath.NormalizePath(physicalRoot),
				tspath.CombinePaths(lexicalRoot, "lexical-hidden"):  tspath.CombinePaths(physicalRoot, "lexical-hidden"),
				tspath.CombinePaths(lexicalRoot, "physical-hidden"): tspath.CombinePaths(physicalRoot, "physical-hidden"),
			},
		}}

		catalog, err := discoverTestAutomatic(context.Background(), fsys, loader, ConfigDiscoveryRequest{
			CWD:                       lexicalRoot,
			Directories:               discoveryDirectoryInputs([]string{lexicalRoot}),
			LimitDirectoryWalkToFiles: true,
			Files: []DiscoveryFile{
				{PathIdentity: rslintconfig.PathIdentity{Path: lexicalIgnored}, Explicit: false},
				{PathIdentity: rslintconfig.PathIdentity{Path: physicalOnly}, Explicit: false},
			},
		})
		if err != nil {
			t.Fatalf("DiscoverAutomatic: %v", err)
		}
		physicalRoot = tspath.NormalizePath(physicalRoot)
		effective := catalogTestConfig(t, catalog, lexicalRoot)
		if effective.IsFileIgnored(
			tspath.CombinePaths(physicalRoot, "lexical-hidden/index.ts"),
			physicalRoot,
		) {
			t.Fatal("lexical .gitignore leaked into the physical Git tree")
		}
		if !effective.IsFileIgnored(
			tspath.CombinePaths(physicalRoot, "physical-hidden/index.ts"),
			physicalRoot,
		) {
			t.Fatal("physical .gitignore was not applied")
		}
		fsys.mu.Lock()
		reads := slices.Clone(fsys.gitignorePaths)
		fsys.mu.Unlock()
		if slices.Contains(reads, lexicalGitignore) || !slices.Contains(reads, physicalGitignore) {
			t.Fatalf("Git source IO paths = %v, want physical root only", reads)
		}
	})

	t.Run("directory symlink preserves lexical config", func(t *testing.T) {
		physicalRoot := discoveryTempDir(t)
		aliasParent := discoveryTempDir(t)
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
			Directories:               discoveryDirectoryInputs([]string{aliasRoot}),
			LimitDirectoryWalkToFiles: true,
			Files:                     []DiscoveryFile{{PathIdentity: rslintconfig.PathIdentity{Path: lexicalTarget}, Explicit: false}},
		})
		aliasRoot = tspath.NormalizePath(aliasRoot)
		if got := catalog.ConfigDirectories(); !reflect.DeepEqual(got, []string{aliasRoot}) {
			t.Fatalf("symlink directory lost lexical owner: %v", got)
		}
		if got := requestedConfigPaths(loader); !reflect.DeepEqual(got, []string{tspath.CombinePaths(physicalRoot, "rslint.config.js")}) {
			t.Fatalf("symlink directory did not load its physical module: %v", got)
		}
	})
}

func TestConfigDiscoveryCaseInsensitiveRootsChooseStableRepresentative(t *testing.T) {
	fsys := &configDiscoveryCaseSensitivityFS{
		FS:            discoveryTestFS(),
		caseSensitive: false,
	}
	build := func(directories []string) []string {
		coordinator := discoveryCoordinator{
			fs: fsys,
			request: ConfigDiscoveryRequest{
				CWD:         "C:/repo",
				Directories: discoveryDirectoryInputs(directories),
			},
		}
		var paths []string
		for _, directory := range coordinator.normalizedDirectoryRoots() {
			paths = append(paths, directory.LexicalPath)
		}
		return paths
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
		coordinator := discoveryCoordinator{
			fs:      fsys,
			request: ConfigDiscoveryRequest{CWD: "//server/share/repo"},
		}

		candidates := coordinator.findCandidateChain("//server/share/repo")
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
		coordinator := discoveryCoordinator{
			fs:      fsys,
			request: ConfigDiscoveryRequest{CWD: "//server/share/repo"},
		}

		if candidate, found := coordinator.findCandidateUp("//server/share/repo"); found {
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

		catalog, err := discoverTestAutomatic(context.Background(), fsys, loader, ConfigDiscoveryRequest{
			CWD:   "//server/share/repo",
			Files: []DiscoveryFile{{PathIdentity: rslintconfig.PathIdentity{Path: "//server/share/repo/src/index.ts"}, Explicit: true}},
		})
		if err != nil {
			t.Fatal(err)
		}
		err = catalog.AllConfigsFailedError()
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
		root := discoveryTempDir(t)
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

		catalog, err := discoverTestAutomatic(context.Background(), discoveryTestFS(), loader, ConfigDiscoveryRequest{
			CWD:         root,
			Directories: discoveryDirectoryInputs([]string{aliasA, aliasB}),
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
		root := discoveryTempDir(t)
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

		catalog, err := discoverTestAutomatic(context.Background(), fsys, loader, ConfigDiscoveryRequest{
			CWD:         root,
			Directories: discoveryDirectoryInputs([]string{upperAlias, lowerAlias}),
		})
		if catalog != nil || err == nil || !strings.Contains(err.Error(), "resolve to the same filesystem location") {
			t.Fatalf("case-only physical aliases: catalog=%+v err=%v", catalog, err)
		}
		if len(loader.activations) != 0 {
			t.Fatalf("ambiguous case aliases activated Node state: %+v", loader.activations)
		}
	})

	t.Run("native case aliases are allowed", func(t *testing.T) {
		root := discoveryTempDir(t)
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

		catalog, err := discoverTestAutomatic(context.Background(), fsys, loader, ConfigDiscoveryRequest{
			CWD:         root,
			Directories: discoveryDirectoryInputs([]string{upperRoot, lowerRoot}),
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

func TestConfigDiscoveryCoalescesNativeCaseAliasesWithoutRealpathCaseNormalization(t *testing.T) {
	root := discoveryTempDir(t)
	upperRoot := tspath.NormalizePath(filepath.Join(root, "Project"))
	lowerRoot := tspath.NormalizePath(filepath.Join(root, "project"))
	upperConfig := writeConfigCandidate(t, upperRoot, "rslint.config.js")
	lowerConfig := tspath.CombinePaths(lowerRoot, "rslint.config.js")
	if _, err := os.Stat(lowerConfig); err != nil {
		if os.IsNotExist(err) {
			t.Skip("fixture requires native case-insensitive path lookup")
		}
		t.Fatal(err)
	}
	upperFile := writeDiscoveryFixture(t, upperRoot, "a.ts", "debugger;\n")
	lowerFile := writeDiscoveryFixture(t, lowerRoot, "b.ts", "debugger;\n")
	loader := newFixtureConfigLoader()
	loader.configs[upperConfig] = namedConfig("owner")
	// EvalSymlinks, now used by upstream on macOS, preserves the caller's
	// spelling. Make that behavior explicit on every case-insensitive host.
	fsys := &configDiscoveryCaseSensitivityFS{
		FS: &configDiscoveryRealpathFS{
			FS: discoveryTestFS(),
			realPaths: map[string]string{
				upperRoot:   upperRoot,
				lowerRoot:   lowerRoot,
				upperConfig: upperConfig,
				lowerConfig: lowerConfig,
			},
		},
		caseSensitive: false,
	}
	catalog, err := discoverTestAutomatic(context.Background(), fsys, loader, ConfigDiscoveryRequest{
		CWD: root,
		Files: []DiscoveryFile{
			{PathIdentity: rslintconfig.PathIdentity{Path: upperFile}, Explicit: true},
			{PathIdentity: rslintconfig.PathIdentity{Path: lowerFile}, Explicit: true},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(catalog.Configs) != 1 || len(catalog.Scopes) != 1 {
		t.Fatalf("native case aliases: configs=%v scopes=%v, want one owner", catalog.Configs, catalog.Scopes)
	}
	if got := requestedConfigPaths(loader); !reflect.DeepEqual(got, []string{upperConfig}) {
		t.Fatalf("requested configs = %v, want one representative", got)
	}
}

func TestConfigDiscoveryReusesNativeCaseAliasAcrossLoadFrontiers(t *testing.T) {
	root := discoveryTempDir(t)
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
	coordinator := newDiscoveryCoordinator(
		context.Background(),
		fsys,
		loader,
		ConfigDiscoveryRequest{CWD: upperRoot},
		"",
		"case-alias-frontiers",
	)
	if err := coordinator.modules.loadCandidates([]configCandidate{{path: upperConfig, directory: upperRoot}}); err != nil {
		t.Fatalf("first frontier: %v", err)
	}
	if err := coordinator.modules.loadCandidates([]configCandidate{{path: lowerConfig, directory: lowerRoot}}); err != nil {
		t.Fatalf("second frontier: %v", err)
	}
	if len(loader.batches) != 1 || len(loader.batches[0].Candidates) != 1 {
		t.Fatalf("load batches = %+v, want one request across frontiers", loader.batches)
	}
	upperState := coordinator.modules.state(upperConfig)
	lowerState := coordinator.modules.state(lowerConfig)
	if upperState == nil || upperState != lowerState {
		t.Fatal("case aliases did not resolve to the same representative load state")
	}
	if got := lowerState.candidate.path; got != upperConfig {
		t.Fatalf("representative config path = %q, want %q", got, upperConfig)
	}
	alternateConfig := writeConfigCandidate(t, upperRoot, "rslint.config.mjs")
	lowerAlternate := tspath.CombinePaths(lowerRoot, "rslint.config.mjs")
	realpathFS.realPaths[tspath.NormalizePath(lowerAlternate)] = tspath.NormalizePath(alternateConfig)
	loader.configs[alternateConfig] = namedConfig("alternate")
	if err := coordinator.modules.loadCandidates([]configCandidate{{path: lowerAlternate, directory: lowerRoot}}); err != nil {
		t.Fatalf("alternate filename frontier: %v", err)
	}
	if state := coordinator.modules.state(lowerAlternate); state == nil || state.candidate.directory != tspath.NormalizePath(upperRoot) {
		t.Fatalf("alternate filename lost native-case owner spelling: %+v", state)
	}
}

func TestConfigDiscoverySiblingFrontierIsBatchedDeterministically(t *testing.T) {
	root := discoveryTempDir(t)
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
	root := discoveryTempDir(t)
	rootConfig := writeConfigCandidate(t, root, "rslint.config.js")
	aDir := filepath.Join(root, "a")
	bDir := filepath.Join(root, "b")
	writeDiscoveryFixture(t, root, "a/index.ts", "export {}\n")
	writeDiscoveryFixture(t, root, "b/index.ts", "export {}\n")

	parallelLoader := newFixtureConfigLoader()
	parallelLoader.configs[rootConfig] = namedConfig("root")
	probeFS := newConfigDiscoveryConcurrencyFS(discoveryTestFS(), aDir, bDir)
	parallelCatalog, err := discoverTestAutomatic(context.Background(), probeFS, parallelLoader, ConfigDiscoveryRequest{
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
	root := discoveryTempDir(t)
	loader := newFixtureConfigLoader()
	catalog := buildFixtureCatalog(t, root, loader, ConfigDiscoveryRequest{CWD: root, ImplicitCWD: true})
	if len(catalog.Configs) != 0 || len(loader.batches) != 0 {
		t.Fatalf("empty discovery produced config work: configs=%v batches=%v", catalog.ConfigDirectories(), loader.batches)
	}
	if len(loader.activations) != 0 {
		t.Fatalf("empty discovery activated an unknown host transaction: %+v", loader.activations)
	}
}

func TestConfigDiscoveryForwardsFreshModuleLoadMode(t *testing.T) {
	root := discoveryTempDir(t)
	loader := newFixtureConfigLoader()
	configPath := writeConfigCandidate(t, root, "rslint.config.js")
	loader.configs[configPath] = namedConfig("root")

	buildFixtureCatalog(t, root, loader, ConfigDiscoveryRequest{
		CWD:         root,
		ImplicitCWD: true,
		Fresh:       true,
	})
	if len(loader.batches) != 1 {
		t.Fatalf("load batch count = %d, want 1", len(loader.batches))
	}
	if got := loader.batches[0].LoadMode; got != ConfigModuleLoadFresh {
		t.Fatalf("load mode = %q, want %q", got, ConfigModuleLoadFresh)
	}
}

func TestConfigDiscoveryPropagatesActivationFailures(t *testing.T) {
	t.Run("transport", func(t *testing.T) {
		root := discoveryTempDir(t)
		loader := newFixtureConfigLoader()
		configPath := writeConfigCandidate(t, root, "rslint.config.js")
		loader.configs[configPath] = namedConfig("root")
		activationErr := errors.New("activation transport failed")
		loader.activate = func(ConfigActivationRequest) (ConfigActivationResponse, error) {
			return ConfigActivationResponse{}, activationErr
		}

		_, err := discoverTestAutomatic(context.Background(), discoveryTestFS(), loader, ConfigDiscoveryRequest{
			CWD:         root,
			ImplicitCWD: true,
		})
		if !errors.Is(err, activationErr) {
			t.Fatalf("error = %v, want %v", err, activationErr)
		}
	})

	t.Run("transaction mismatch", func(t *testing.T) {
		root := discoveryTempDir(t)
		loader := newFixtureConfigLoader()
		configPath := writeConfigCandidate(t, root, "rslint.config.js")
		loader.configs[configPath] = namedConfig("root")
		loader.activate = func(ConfigActivationRequest) (ConfigActivationResponse, error) {
			return ConfigActivationResponse{TransactionID: "other"}, nil
		}

		_, err := discoverTestAutomatic(context.Background(), discoveryTestFS(), loader, ConfigDiscoveryRequest{
			CWD:         root,
			ImplicitCWD: true,
		})
		var protocolErr *ConfigDiscoveryProtocolError
		if !errors.As(err, &protocolErr) {
			t.Fatalf("error = %v, want ConfigDiscoveryProtocolError", err)
		}
	})
}

func TestConfigDiscoveryTransactionIDsAreUniqueAcrossBuilds(t *testing.T) {
	root := discoveryTempDir(t)
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
			catalog, err := discoverTestAutomatic(context.Background(), discoveryTestFS(), nil, ConfigDiscoveryRequest{CWD: root})
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
	root := discoveryTempDir(t)
	loader := newFixtureConfigLoader()
	configPath := writeConfigCandidate(t, root, "rslint.config.js")
	nestedConfigPath := writeConfigCandidate(t, root, "packages/app/rslint.config.js")
	loader.failures[configPath] = ConfigModuleError{Code: "ERR", Message: "nope"}
	loader.failures[nestedConfigPath] = ConfigModuleError{Code: "ERR_NESTED", Message: "nested nope"}
	catalog, err := discoverTestAutomatic(context.Background(), discoveryTestFS(), loader, ConfigDiscoveryRequest{
		CWD:         root,
		ImplicitCWD: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	err = catalog.AllConfigsFailedError()
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
	catalog, err = discoverTestAutomatic(context.Background(), discoveryTestFS(), invalidLoader, ConfigDiscoveryRequest{
		CWD:         root,
		ImplicitCWD: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	err = catalog.AllConfigsFailedError()
	if !errors.Is(err, ErrAllConfigsFailed) || !strings.Contains(err.Error(), "invalid config in") {
		t.Fatalf("invalid error = %v, want ErrAllConfigsFailed with invalid-stage context", err)
	}
	if catalog.GitPolicy == nil || !catalog.GitPolicy.CoversDirectory(root) {
		t.Fatal("all-failed catalog lost its Git policy")
	}
	if len(loader.activations) != 0 || len(invalidLoader.activations) != 0 {
		t.Fatal("all-failed discovery activated an empty module generation")
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
	catalog, err := discoverTestAutomatic(context.Background(), discoveryTestFS(), loader, request)
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
	// Fixtures are repositories unless a test constructs its module directly.
	if err := os.MkdirAll(filepath.Join(root, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
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

// These adapters stand in for entrypoints: discovery receives identities that
// have already been frozen through the same composed VFS.
func discoveryDirectoryInputs(paths []string) []rslintconfig.DirectoryIdentity {
	if paths == nil {
		return nil
	}
	directories := make([]rslintconfig.DirectoryIdentity, 0, len(paths))
	for _, path := range paths {
		directories = append(directories, rslintconfig.DirectoryIdentity{LexicalPath: path})
	}
	return directories
}

func freezeDiscoveryTestInputs(fsys vfs.FS, cwd string, files []DiscoveryFile, directories []rslintconfig.DirectoryIdentity) {
	for index := range files {
		file := &files[index]
		identity := rslintconfig.FreezeFileIdentity(normalizeDiscoveryPath(file.Path, cwd), fsys)
		if file.CanonicalPath != "" {
			identity.CanonicalPath = file.CanonicalPath
		}
		if file.CanonicalParentPath != "" {
			identity.CanonicalParentPath = file.CanonicalParentPath
		}
		file.PathIdentity = identity
	}
	for index := range directories {
		if directories[index].CanonicalPath == "" {
			directories[index] = rslintconfig.ResolveDirectoryIdentities([]string{normalizeDiscoveryPath(directories[index].LexicalPath, cwd)}, fsys)[0]
		}
	}
}

func discoverTestAutomatic(ctx context.Context, fsys vfs.FS, loader ConfigModuleLoader, request ConfigDiscoveryRequest) (*ConfigCatalog, error) {
	freezeDiscoveryTestInputs(fsys, request.CWD, request.Files, request.Directories)
	return DiscoverAutomatic(ctx, fsys, loader, request)
}

func loadTestExplicit(ctx context.Context, fsys vfs.FS, loader ConfigModuleLoader, request ExplicitConfigRequest) (*ConfigCatalog, error) {
	freezeDiscoveryTestInputs(fsys, request.CWD, request.TargetFiles, request.TargetDirectories)
	return LoadExplicitConfig(ctx, fsys, loader, request)
}

type discoveredTestConfig struct {
	entries rslintconfig.RslintConfig
	matcher rslintconfig.TargetMatcher
	fs      vfs.FS
}

func catalogTestConfig(t *testing.T, catalog *ConfigCatalog, directory string) discoveredTestConfig {
	t.Helper()
	entries, exists := catalog.Configs[directory]
	if !exists {
		t.Fatalf("config owner %q not found in %v", directory, catalog.ConfigDirectories())
	}
	return boundTestConfig(t, entries, directory, discoveryTestFS(), catalog.GitPolicy)
}

func boundTestConfig(t *testing.T, entries rslintconfig.RslintConfig, directory string, fsys vfs.FS, policy *rslintconfig.GitPolicy) discoveredTestConfig {
	t.Helper()
	spaces := rslintconfig.NewPathSpaceSnapshot(map[string]rslintconfig.RslintConfig{directory: entries}, fsys)
	matcher, err := rslintconfig.NewTargetMatcherWithPathSpaces(entries, directory, fsys, spaces, policy)
	if err != nil {
		t.Fatal(err)
	}
	return discoveredTestConfig{entries: entries, matcher: matcher, fs: fsys}
}

func collectedTestConfig(t *testing.T, entries rslintconfig.RslintConfig, directory string, fsys vfs.FS, files []string, directories []string) discoveredTestConfig {
	t.Helper()
	var identities []rslintconfig.PathIdentity
	if files != nil {
		identities = make([]rslintconfig.PathIdentity, 0, len(files))
	}
	for _, file := range files {
		identities = append(identities, rslintconfig.FreezeFileIdentity(file, fsys))
	}
	if files == nil && directories == nil {
		directories = []string{directory}
	}
	policy := rslintconfig.CollectGitPolicy(entries, directory, fsys, identities, rslintconfig.ResolveDirectoryIdentities(directories, fsys))
	return boundTestConfig(t, entries, directory, fsys, policy)
}

func (config discoveredTestConfig) IsFileIgnored(filePath string, _ string) bool {
	return config.matcher.MatchFile(rslintconfig.FreezeFileIdentity(filePath, config.fs)).GloballyIgnored()
}

func (config discoveredTestConfig) GetConfigForFile(filePath string, cwd string) *rslintconfig.MergedConfig {
	match := config.matcher.MatchFile(rslintconfig.FreezeFileIdentity(filePath, config.fs))
	if match.GloballyIgnored() || !match.Selected() {
		return nil
	}
	return config.entries.GetConfigForFile(filePath, cwd)
}

func discoveryTempDir(t *testing.T) string {
	t.Helper()
	directory := t.TempDir()
	physical, err := filepath.EvalSymlinks(directory)
	if err != nil {
		t.Fatal(err)
	}
	return tspath.NormalizePath(physical)
}

func TestConfigDiscoveryExplicitConfigGitBoundaryIndependentOfCWD(t *testing.T) {
	root := discoveryTempDir(t)
	configPath := writeConfigCandidate(t, root, "rslint.config.js")
	packageDirectory := tspath.CombinePaths(root, "packages/app")
	target := writeDiscoveryFixture(t, root, "packages/app/dist/index.js", "debugger;\n")
	writeDiscoveryFixture(t, root, ".gitignore", "dist/\n")
	writeDiscoveryFixture(t, root, "packages/app/.gitignore", "binary\n")
	hiddenSource := writeDiscoveryFixture(t, root, "packages/app/dist/.gitignore", "!index.js\n")
	for _, explicit := range []bool{false, true} {
		for _, directory := range []bool{false, true} {
			for _, cwd := range []string{root, packageDirectory} {
				name := strconv.FormatBool(explicit) + "/directory=" + strconv.FormatBool(directory) + "/cwd=" + tspath.GetBaseFileName(cwd)
				t.Run(name, func(t *testing.T) {
					loader := newFixtureConfigLoader()
					loader.configs[configPath] = namedConfig("root")
					spy := &configDiscoveryReadSpyFS{FS: discoveryTestFS()}
					var files []DiscoveryFile
					var directories []rslintconfig.DirectoryIdentity
					if directory {
						directories = discoveryDirectoryInputs([]string{tspath.GetDirectoryPath(target)})
					} else {
						files = []DiscoveryFile{{PathIdentity: rslintconfig.PathIdentity{Path: target}, Explicit: true}}
					}
					var catalog *ConfigCatalog
					var err error
					if explicit {
						catalog, err = loadTestExplicit(context.Background(), spy, loader, ExplicitConfigRequest{CWD: cwd, ConfigPath: configPath, TargetFiles: files, TargetDirectories: directories})
					} else {
						catalog, err = discoverTestAutomatic(context.Background(), spy, loader, ConfigDiscoveryRequest{CWD: cwd, Files: files, Directories: directories})
					}
					if err != nil {
						t.Fatal(err)
					}
					if !catalogTestConfig(t, catalog, root).IsFileIgnored(target, root) {
						t.Fatal("changing cwd or target spelling lost the repository's ancestor rule")
					}
					if slices.Contains(spy.gitignorePaths, hiddenSource) {
						t.Fatalf("raw Git negation source below an ignored directory was read: %v", spy.gitignorePaths)
					}
				})
			}
		}
	}
}

func TestConfigDiscoveryNestedGitBoundaryAndOverlappingRoots(t *testing.T) {
	for _, marker := range []string{"directory", "file"} {
		for _, nestedLocation := range []string{"nested", "hidden/nested"} {
			for _, localSource := range []bool{false, true} {
				t.Run(marker+"/"+nestedLocation+"/source="+strconv.FormatBool(localSource), func(t *testing.T) {
					root := discoveryTempDir(t)
					configPath := writeConfigCandidate(t, root, "rslint.config.js")
					writeDiscoveryFixture(t, root, ".gitignore", "index.ts\nnested/\nhidden/\n")
					outerFile := writeDiscoveryFixture(t, root, "index.ts", "export {};\n")
					nestedFile := writeDiscoveryFixture(t, root, nestedLocation+"/index.ts", "export {};\n")
					nestedDirectory := tspath.GetDirectoryPath(nestedFile)
					if marker == "directory" {
						if err := os.Mkdir(filepath.Join(nestedDirectory, ".git"), 0o755); err != nil {
							t.Fatal(err)
						}
					} else {
						writeDiscoveryFixture(t, nestedDirectory, ".git", "gitdir: /unrelated/git-metadata\n")
					}
					localFile := writeDiscoveryFixture(t, nestedDirectory, "local.ts", "export {};\n")
					if localSource {
						writeDiscoveryFixture(t, nestedDirectory, ".gitignore", "local.ts\n")
					}
					loader := newFixtureConfigLoader()
					loader.configs[configPath] = namedConfig("owner-above-git-root")
					catalog := buildFixtureCatalog(t, root, loader, ConfigDiscoveryRequest{CWD: root, Directories: discoveryDirectoryInputs([]string{root, nestedDirectory})})
					if len(catalog.TargetDirectories) != 2 {
						t.Fatalf("explicit nested directory input was coalesced: %+v", catalog.TargetDirectories)
					}
					matcher := catalogTestConfig(t, catalog, root)
					if !matcher.IsFileIgnored(outerFile, root) || matcher.IsFileIgnored(nestedFile, root) {
						t.Fatal("nearest Git boundary did not block inherited rules, including the empty-root case")
					}
					if got := matcher.IsFileIgnored(localFile, root); got != localSource {
						t.Fatalf("nested source ignored = %t, want %t", got, localSource)
					}
					if !catalog.GitPolicy.CoversDirectory(nestedDirectory) {
						t.Fatal("overlapping outer traversal hid the separately requested nested repository")
					}
					if got := requestedConfigPaths(loader); !reflect.DeepEqual(got, []string{configPath}) {
						t.Fatalf("Git boundary changed authored ancestor lookup: %v", got)
					}
				})
			}
		}
	}
}

func TestConfigDiscoveryLeafNegationWalksWithoutExecutingHiddenConfig(t *testing.T) {
	root := discoveryTempDir(t)
	rootConfig := writeConfigCandidate(t, root, "rslint.config.js")
	hiddenConfig := writeConfigCandidate(t, root, "ignored/rslint.config.js")
	nestedConfig := writeConfigCandidate(t, root, "ignored/nested/rslint.config.js")
	writeDiscoveryFixture(t, root, ".gitignore", "ignored/\n")
	hiddenSource := writeDiscoveryFixture(t, root, "ignored/.gitignore", "keep.ts\n")
	keep := writeDiscoveryFixture(t, root, "ignored/keep.ts", "export {};\n")
	nestedFile := writeDiscoveryFixture(t, root, "ignored/nested/index.ts", "export {};\n")
	nestedDirectory := tspath.GetDirectoryPath(nestedFile)
	if err := os.Mkdir(filepath.Join(nestedDirectory, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	loader := newFixtureConfigLoader()
	loader.configs[rootConfig] = append(rslintconfig.RslintConfig{{Ignores: []string{"!ignored/keep.ts"}}}, namedConfig("root")...)
	loader.configs[hiddenConfig] = namedConfig("must-not-execute")
	loader.configs[nestedConfig] = namedConfig("nested")
	spy := &configDiscoveryReadSpyFS{FS: discoveryTestFS()}
	catalog, err := discoverTestAutomatic(context.Background(), spy, loader, ConfigDiscoveryRequest{CWD: root, ImplicitCWD: true})
	if err != nil {
		t.Fatal(err)
	}
	if got := requestedConfigPaths(loader); !reflect.DeepEqual(got, []string{rootConfig, nestedConfig}) {
		t.Fatalf("leaf negation executed a hidden enclosing config or hid a visited Git boundary: %v", got)
	}
	if slices.Contains(spy.gitignorePaths, hiddenSource) {
		t.Fatalf("authored negation reopened raw Git source reachability: %v", spy.gitignorePaths)
	}
	if catalogTestConfig(t, catalog, root).IsFileIgnored(keep, root) {
		t.Fatal("authored leaf negation did not reopen its lint target")
	}
	if catalogTestConfig(t, catalog, nestedDirectory).IsFileIgnored(nestedFile, nestedDirectory) {
		t.Fatal("visited nested Git root retained the outer raw ignore")
	}
}

func TestConfigDiscoveryAdditionalGitFilesDoNotDiscoverModules(t *testing.T) {
	root, external := discoveryTempDir(t), discoveryTempDir(t)
	rootConfig := writeConfigCandidate(t, root, "rslint.config.js")
	externalConfig := writeConfigCandidate(t, external, "rslint.config.js")
	document := writeDiscoveryFixture(t, external, "doc.generated.ts", "export {};\n")
	writeDiscoveryFixture(t, external, ".gitignore", "*.generated.ts\n")
	loader := newFixtureConfigLoader()
	loader.configs[rootConfig] = namedConfig("workspace")
	loader.configs[externalConfig] = namedConfig("must-not-execute")
	identity := rslintconfig.FreezeFileIdentity(document, discoveryTestFS())
	catalog := buildFixtureCatalog(t, root, loader, ConfigDiscoveryRequest{CWD: root, ImplicitCWD: true, AdditionalGitFiles: []rslintconfig.PathIdentity{identity}})
	if got := requestedConfigPaths(loader); !reflect.DeepEqual(got, []string{rootConfig}) {
		t.Fatalf("Git-only input participated in module discovery: %v", got)
	}
	if len(catalog.TargetFiles) != 0 || !catalog.GitPolicy.CoversDirectory(external) {
		t.Fatalf("Git-only input changed lint targets or lost coverage: %+v", catalog.TargetFiles)
	}
	if !catalogTestConfig(t, catalog, root).IsFileIgnored(document, root) {
		t.Fatal("outside-workspace document did not receive its own Git policy")
	}
}

func TestConfigDiscoveryWithoutSuccessfulModuleStillFreezesGitPolicy(t *testing.T) {
	for _, mode := range []string{"no candidate", "automatic failure", "explicit failure"} {
		t.Run(mode, func(t *testing.T) {
			root := discoveryTempDir(t)
			target := writeDiscoveryFixture(t, root, "src/generated.ts", "export {};\n")
			writeDiscoveryFixture(t, root, ".gitignore", "generated.ts\n")
			loader := newFixtureConfigLoader()
			configPath := ""
			if mode != "no candidate" {
				configPath = writeConfigCandidate(t, root, "rslint.config.js")
				loader.failures[configPath] = ConfigModuleError{Code: "ERR", Message: "broken"}
			}
			files := []DiscoveryFile{{PathIdentity: rslintconfig.PathIdentity{Path: target}, Explicit: true}}
			var catalog *ConfigCatalog
			var err error
			if mode == "explicit failure" {
				catalog, err = loadTestExplicit(context.Background(), discoveryTestFS(), loader, ExplicitConfigRequest{CWD: root, ConfigPath: configPath, TargetFiles: files})
			} else {
				catalog, err = discoverTestAutomatic(context.Background(), discoveryTestFS(), loader, ConfigDiscoveryRequest{CWD: root, Files: files})
			}
			if err != nil {
				t.Fatal(err)
			}
			if len(catalog.Configs) != 0 || catalog.GitPolicy == nil || !catalog.GitPolicy.CoversDirectory(tspath.GetDirectoryPath(target)) {
				t.Fatalf("incomplete catalog without a successful module: %+v", catalog)
			}
			if got := catalog.AllConfigsFailedError(); (got != nil) != (mode != "no candidate") {
				t.Fatalf("business failure = %v", got)
			}
			if len(loader.activations) != 0 {
				t.Fatal("discovery activated an empty successful module set")
			}
			matcher := boundTestConfig(t, nil, root, discoveryTestFS(), catalog.GitPolicy)
			if !matcher.IsFileIgnored(target, root) {
				t.Fatal("fallback lost the request's Git rule")
			}
		})
	}
}

func TestConfigDiscoveryConsumesFrozenInputParentAndOriginalLeaf(t *testing.T) {
	configRoot, physicalRoot, leafRoot := discoveryTempDir(t), discoveryTempDir(t), discoveryTempDir(t)
	configPath := writeConfigCandidate(t, configRoot, "custom.config.js")
	writeConfigCandidate(t, physicalRoot, "unused.config.js")
	writeConfigCandidate(t, leafRoot, "unused.config.js")
	writeDiscoveryFixture(t, physicalRoot, ".gitignore", "entry.ts\n")
	leaf := writeDiscoveryFixture(t, leafRoot, "source.ts", "export {};\n")
	if err := os.Symlink(leaf, filepath.Join(physicalRoot, "entry.ts")); err != nil {
		t.Skipf("file symlink unavailable: %v", err)
	}
	alias := filepath.Join(discoveryTempDir(t), "alias")
	if err := os.Symlink(physicalRoot, alias); err != nil {
		t.Skipf("directory symlink unavailable: %v", err)
	}
	file := tspath.CombinePaths(alias, "entry.ts")
	baseFS := discoveryTestFS()
	identity := rslintconfig.FreezeFileIdentity(file, baseFS)
	directories := rslintconfig.ResolveDirectoryIdentities([]string{alias}, baseFS)
	fsys := &configDiscoveryRealpathFS{FS: baseFS}
	loader := newFixtureConfigLoader()
	loader.configs[configPath] = namedConfig("external-config")
	catalog, err := LoadExplicitConfig(context.Background(), fsys, loader, ExplicitConfigRequest{
		CWD: configRoot, ConfigPath: configPath,
		TargetFiles: []DiscoveryFile{{PathIdentity: identity, Explicit: true}}, TargetDirectories: directories,
	})
	if err != nil {
		t.Fatal(err)
	}
	if fsys.realpathCallCount(alias) != 0 || fsys.realpathCallCount(file) != 0 {
		t.Fatalf("discovery re-resolved frozen input identities: dir=%d file=%d", fsys.realpathCallCount(alias), fsys.realpathCallCount(file))
	}
	if !reflect.DeepEqual(catalog.TargetFiles, []rslintconfig.PathIdentity{identity}) || !reflect.DeepEqual(catalog.TargetDirectories, directories) {
		t.Fatal("catalog did not preserve the request's frozen input identities")
	}
	spaces := rslintconfig.NewPathSpaceSnapshot(catalog.Configs, baseFS)
	matcher, err := rslintconfig.NewTargetMatcherWithPathSpaces(catalog.Configs[configRoot], configRoot, baseFS, spaces, catalog.GitPolicy)
	if err != nil {
		t.Fatal(err)
	}
	if !matcher.MatchFile(identity).GloballyIgnored() {
		t.Fatalf("Git followed the leaf destination instead of parent plus original basename: %+v", identity)
	}
}

func TestConfigDiscoveryPreservesNilAndEmptyTargetHandoff(t *testing.T) {
	root := discoveryTempDir(t)
	configPath := writeConfigCandidate(t, root, "rslint.config.js")
	file := writeDiscoveryFixture(t, root, "src/index.ts", "export {};\n")
	for _, mode := range []string{"implicit cwd", "directory", "empty exact set", "empty file input with directory"} {
		t.Run(mode, func(t *testing.T) {
			loader := newFixtureConfigLoader()
			loader.configs[configPath] = namedConfig("root")
			loader.configs[configPath][0].Files = []string{"**/*.ts"}
			request := ConfigDiscoveryRequest{CWD: root}
			emptyExactSet := strings.HasPrefix(mode, "empty")
			if emptyExactSet {
				request.Files = []DiscoveryFile{}
			}
			if mode == "implicit cwd" {
				request.ImplicitCWD = true
			}
			if mode == "directory" || mode == "empty file input with directory" {
				request.Directories = discoveryDirectoryInputs([]string{root})
			}
			catalog := buildFixtureCatalog(t, root, loader, request)
			if (catalog.TargetFiles != nil) != emptyExactSet {
				t.Fatalf("nil/empty handoff changed: %+v", catalog.TargetFiles)
			}
			plan, err := target.Resolve(target.Request{ConfigMap: catalog.Configs, ConfigDirectory: root, ScanRoot: root, OwnerScopes: catalog.Scopes, FS: discoveryTestFS(), Files: catalog.TargetFiles, Directories: catalog.TargetDirectories, GitPolicy: catalog.GitPolicy})
			if err != nil {
				t.Fatal(err)
			}
			if mode == "empty exact set" {
				if len(plan.Files) != 0 {
					t.Fatalf("empty exact set gained targets: %+v", plan.Files)
				}
			} else if !slices.ContainsFunc(plan.Files, func(targetFile target.File) bool { return targetFile.Path == file }) {
				t.Fatalf("recursive input was discarded during discovery handoff: %+v", plan.Files)
			}
		})
	}
}

func TestConfigDiscoveryIgnoredDirectoryRootDoesNotDependOnCWD(t *testing.T) {
	root := discoveryTempDir(t)
	rootConfig := writeConfigCandidate(t, root, "rslint.config.js")
	ignoredConfig := writeConfigCandidate(t, root, "ignored/rslint.config.js")
	nestedConfig := writeConfigCandidate(t, root, "ignored/nested/rslint.config.js")
	writeDiscoveryFixture(t, root, ".gitignore", "ignored/\n")
	writeDiscoveryFixture(t, root, "ignored/keep.ts", "export {};\n")
	nestedFile := writeDiscoveryFixture(t, root, "ignored/nested/index.ts", "export {};\n")
	nestedMarker := writeDiscoveryFixture(t, root, "ignored/nested/.git", "gitdir: /unused\n")
	ignored := tspath.CombinePaths(root, "ignored")
	nested := tspath.GetDirectoryPath(nestedFile)
	for _, explicit := range []bool{false, true} {
		for _, separateNestedInput := range []bool{false, true} {
			for _, authoredNegation := range []bool{false, true} {
				name := "explicit=" + strconv.FormatBool(explicit) + "/nested-input=" + strconv.FormatBool(separateNestedInput) + "/negation=" + strconv.FormatBool(authoredNegation)
				t.Run(name, func(t *testing.T) {
					var previousPaths []string
					for _, cwd := range []string{root, ignored} {
						loader := newFixtureConfigLoader()
						loader.configs[rootConfig] = namedConfig("root")
						loader.configs[ignoredConfig] = namedConfig("supplied-root")
						loader.configs[nestedConfig] = namedConfig("nested")
						if authoredNegation {
							loader.configs[rootConfig] = append(rslintconfig.RslintConfig{{Ignores: []string{"!ignored/keep.ts"}}}, loader.configs[rootConfig]...)
							loader.configs[ignoredConfig] = append(rslintconfig.RslintConfig{{Ignores: []string{"!keep.ts"}}}, loader.configs[ignoredConfig]...)
						}
						directories := []string{ignored}
						if separateNestedInput {
							directories = append(directories, nested)
						}
						spy := &configDiscoveryReadSpyFS{FS: discoveryTestFS()}
						var catalog *ConfigCatalog
						var err error
						if explicit {
							catalog, err = loadTestExplicit(context.Background(), spy, loader, ExplicitConfigRequest{CWD: cwd, ConfigPath: rootConfig, TargetDirectories: discoveryDirectoryInputs(directories)})
						} else {
							catalog, err = discoverTestAutomatic(context.Background(), spy, loader, ConfigDiscoveryRequest{CWD: cwd, Directories: discoveryDirectoryInputs(directories)})
						}
						if err != nil {
							t.Fatal(err)
						}
						if !explicit && !slices.Contains(requestedConfigPaths(loader), ignoredConfig) {
							t.Fatal("raw traversal pruning removed the supplied-root config ownership exception")
						}
						mayVisitNested := separateNestedInput || authoredNegation
						if !mayVisitNested {
							if slices.Contains(spy.markerPaths, nestedMarker) {
								t.Fatalf("hidden nested Git marker was probed: %v", spy.markerPaths)
							}
							if slices.Contains(spy.accessedDirs, ignored) {
								t.Fatalf("ignored supplied root was expanded: %v", spy.accessedDirs)
							}
							if slices.Contains(requestedConfigPaths(loader), nestedConfig) || catalog.GitPolicy.CoversDirectory(nested) {
								t.Fatal("hidden nested repository entered the catalog")
							}
						} else if !catalog.GitPolicy.CoversDirectory(nested) {
							t.Fatal("explicit nested input or authored negation lost nested Git coverage")
						}
						request := target.Request{ConfigDirectory: root, ScanRoot: cwd, FS: discoveryTestFS(), Files: catalog.TargetFiles, Directories: catalog.TargetDirectories, GitPolicy: catalog.GitPolicy}
						if explicit {
							request.Config = catalog.Configs[root]
						} else {
							request.ConfigMap = catalog.Configs
							request.OwnerScopes = catalog.Scopes
						}
						plan, err := target.Resolve(request)
						if err != nil {
							t.Fatal(err)
						}
						paths := make([]string, 0, len(plan.Files))
						for _, file := range plan.Files {
							paths = append(paths, file.Path)
						}
						if !mayVisitNested && len(paths) != 0 {
							t.Fatalf("ignored root gained lint targets at cwd %s: %v", cwd, paths)
						}
						if mayVisitNested && !slices.Contains(paths, nestedFile) {
							t.Fatalf("reachable nested input was lost at cwd %s: %v", cwd, paths)
						}
						if previousPaths != nil && !reflect.DeepEqual(paths, previousPaths) {
							t.Fatalf("same target/config changed with cwd: %v != %v", paths, previousPaths)
						}
						previousPaths = paths
					}
				})
			}
		}
	}
}

type movingDiscoveryDirectoryFS struct {
	vfs.FS
	t                    *testing.T
	alias, first, second string
	once                 sync.Once
}

func (fsys *movingDiscoveryDirectoryFS) GetAccessibleEntries(directory string) vfs.Entries {
	if directory == fsys.first || directory == fsys.alias {
		fsys.once.Do(func() {
			if err := os.Remove(fsys.alias); err != nil {
				fsys.t.Fatal(err)
			}
			if err := os.Symlink(fsys.second, fsys.alias); err != nil {
				fsys.t.Fatal(err)
			}
		})
	}
	return fsys.FS.GetAccessibleEntries(directory)
}

func TestConfigDiscoveryMovedDirectoryRetainsPhysicalTreeAndLexicalOwner(t *testing.T) {
	for _, hasOwnConfig := range []bool{false, true} {
		t.Run(strconv.FormatBool(hasOwnConfig), func(t *testing.T) {
			root := discoveryTempDir(t)
			first, second := tspath.CombinePaths(root, "first"), tspath.CombinePaths(root, "second")
			rootConfig := writeConfigCandidate(t, root, "rslint.config.js")
			source := writeDiscoveryFixture(t, first, "scope/from-first.js", "debugger;\n")
			writeDiscoveryFixture(t, second, "scope/from-second.js", "debugger;\n")
			secondConfig := writeConfigCandidate(t, second, "scope/rslint.config.js")
			alias := tspath.CombinePaths(root, "alias")
			if err := os.Symlink(first, alias); err != nil {
				t.Skipf("directory symlink unavailable: %v", err)
			}
			loader := newFixtureConfigLoader()
			loader.configs[rootConfig] = namedConfig("outer")
			loader.configs[secondConfig] = namedConfig("wrong-second")
			firstConfig := tspath.CombinePaths(first, "scope/rslint.config.js")
			if hasOwnConfig {
				writeConfigCandidate(t, first, "scope/rslint.config.js")
				loader.configs[firstConfig] = namedConfig("first")
			}
			fsys := &movingDiscoveryDirectoryFS{FS: discoveryTestFS(), t: t, alias: alias, first: first, second: second}
			directories := rslintconfig.ResolveDirectoryIdentities([]string{alias}, fsys)
			catalog, err := DiscoverAutomatic(context.Background(), fsys, loader, ConfigDiscoveryRequest{CWD: root, Directories: directories, SingleThreaded: true})
			if err != nil {
				t.Fatal(err)
			}
			owner := tspath.CombinePaths(alias, "scope")
			if entries, exists := catalog.Configs[owner]; exists != hasOwnConfig || exists && namedConfigEntry(entries).Name != "first" {
				t.Fatalf("moved alias changed owner contents: %v", catalog.Configs)
			}
			for _, batch := range loader.batches {
				for _, candidate := range batch.Candidates {
					if candidate.ConfigPath == secondConfig || candidate.ConfigPath == tspath.CombinePaths(owner, "rslint.config.js") {
						t.Fatalf("module loader received moving alias or other tree: %+v", candidate)
					}
					if candidate.ConfigPath == firstConfig && candidate.ConfigDirectory != owner {
						t.Fatalf("physical module lost lexical routing: %+v", candidate)
					}
				}
			}
			plan, err := target.Resolve(target.Request{ConfigMap: catalog.Configs, OwnerScopes: catalog.Scopes, FS: fsys, Directories: catalog.TargetDirectories, Files: catalog.TargetFiles, GitPolicy: catalog.GitPolicy, PathSpaces: catalog.PathSpaces, SingleThreaded: true})
			if err != nil {
				t.Fatal(err)
			}
			found := false
			for _, file := range plan.Files {
				if !strings.HasPrefix(file.CanonicalPath, first+"/") || !strings.HasPrefix(file.Path, alias+"/") {
					t.Fatalf("mixed requested and physical tree: %+v", file)
				}
				found = found || file.CanonicalPath == source
			}
			if !found {
				t.Fatalf("original input disappeared: %+v", plan.Files)
			}
			if moved := tspath.NormalizePath(osvfs.FS().Realpath(alias)); moved != second {
				t.Fatalf("attack did not move the directory alias: %s", moved)
			}
			if physical, _ := plan.PathSpaces().PhysicalDirectory(alias); physical != first {
				t.Fatalf("target stage re-resolved the owner: %s", physical)
			}
		})
	}
}
