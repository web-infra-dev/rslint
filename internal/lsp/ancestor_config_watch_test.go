package lsp

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/jsonrpc"
	"github.com/microsoft/typescript-go/shim/lsp/lsproto"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"

	"github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/config/discovery"
)

func TestAncestorJSConfigFileWatchersExcludeWorkspace(t *testing.T) {
	watchers := ancestorJSConfigFileWatchers("/workspace/packages/app", true)
	wantBases := []string{
		"file:///workspace/packages",
		"file:///workspace",
		"file:///",
	}
	configFileNames := discovery.AutoJSConfigFileNames()
	wantCount := len(wantBases) * len(configFileNames)
	if len(watchers) != wantCount {
		t.Fatalf("watcher count=%d, want %d", len(watchers), wantCount)
	}
	for directoryIndex, wantBase := range wantBases {
		for configIndex, wantName := range configFileNames {
			index := directoryIndex*len(configFileNames) + configIndex
			relative := watchers[index].GlobPattern.RelativePattern
			if relative == nil || relative.BaseUri.URI == nil ||
				string(*relative.BaseUri.URI) != wantBase || relative.Pattern != wantName {
				t.Fatalf(
					"watcher[%d]=%+v, want base %q pattern %q",
					index,
					watchers[index],
					wantBase,
					wantName,
				)
			}
		}
	}

	withoutRelativePatterns := ancestorJSConfigFileWatchers("/workspace/packages/app", false)
	if len(withoutRelativePatterns) != wantCount {
		t.Fatalf("absolute watcher count=%d, want %d", len(withoutRelativePatterns), wantCount)
	}
	first := withoutRelativePatterns[0].GlobPattern.Pattern
	last := withoutRelativePatterns[len(withoutRelativePatterns)-1].GlobPattern.Pattern
	if first == nil || *first != "/workspace/packages/rslint.config.js" {
		t.Fatalf("first absolute watcher=%+v", withoutRelativePatterns[0])
	}
	if last == nil || *last != "/rslint.config.mts" {
		t.Fatalf("last absolute watcher=%+v", withoutRelativePatterns[len(withoutRelativePatterns)-1])
	}
	if rootWatchers := ancestorJSConfigFileWatchers("/", true); len(rootWatchers) != 0 {
		t.Fatalf("filesystem root has ancestor watchers: %+v", rootWatchers)
	}
}

func TestIsStrictAncestorAutoJSConfigPath(t *testing.T) {
	fsys := &mockFS{files: map[string]bool{}}
	workspace := "/repo/packages/app"
	tests := []struct {
		path string
		want bool
	}{
		{path: "/repo/packages/rslint.config.js", want: true},
		{path: "/repo/rslint.config.mjs", want: true},
		{path: "/rslint.config.ts", want: true},
		{path: "/repo/packages/app/rslint.config.mts", want: false},
		{path: "/repo/packages/app/src/rslint.config.js", want: false},
		{path: "/repo/packages/sibling/rslint.config.js", want: false},
		{path: "/repo/packages/rslint.json", want: false},
		{path: "/repo/packages/rslint.config.cjs", want: false},
	}
	for _, test := range tests {
		if got := isStrictAncestorAutoJSConfigPath(test.path, workspace, fsys); got != test.want {
			t.Errorf("isStrictAncestorAutoJSConfigPath(%q)=%t, want %t", test.path, got, test.want)
		}
	}
}

func TestWorkspaceConfigEventsDoNotDuplicateTransactionalRefresh(t *testing.T) {
	workspace := tspath.NormalizePath(t.TempDir())
	tests := []string{
		filepath.Join(workspace, "rslint.config.js"),
		filepath.Join(workspace, "packages", "app", "rslint.config.mjs"),
		filepath.Join(workspace, "rslint.json"),
	}
	for _, configPath := range tests {
		t.Run(filepath.Base(configPath), func(t *testing.T) {
			s, outgoing := newAncestorConfigWatchTestServer(workspace)
			s.configDiscoveryActive = true
			s.rslintConfigPath = "/committed/rslint.json"
			result := startConfigWatchEvent(s, configPath, lsproto.FileChangeTypeChanged)
			select {
			case message := <-outgoing:
				respondToConfigReverseRequest(t, s, message.AsRequest(), nil, errors.New("unexpected duplicate refresh"))
				t.Fatalf("workspace event started a duplicate Go refresh: %+v", message)
			case err := <-result:
				if err != nil {
					t.Fatalf("workspace config event: %v", err)
				}
			case <-time.After(time.Second):
				t.Fatal("workspace config event did not complete")
			}
			if s.rslintConfigPath != "/committed/rslint.json" {
				t.Fatalf("workspace event directly reloaded JSON config: %q", s.rslintConfigPath)
			}
		})
	}
}

func TestAncestorJSConfigWatcherRefreshesChangedAndDeletedActiveConfig(t *testing.T) {
	config.RegisterAllRules()
	root := tspath.NormalizePath(t.TempDir())
	workspace := tspath.NormalizePath(filepath.Join(root, "repo", "packages", "app"))
	ancestor := tspath.NormalizePath(filepath.Join(root, "repo"))
	if err := os.MkdirAll(workspace, 0o755); err != nil {
		t.Fatal(err)
	}
	writeConfigCandidate(t, ancestor)
	configPath := filepath.Join(ancestor, "rslint.config.mjs")

	s, outgoing := newAncestorConfigWatchTestServer(workspace)
	installLastGoodConfig(s, ancestor)
	s.configDiscoveryActive = true
	if err := os.WriteFile(configPath, []byte("export default [{ rules: { 'no-debugger': 'error' } }];\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	changed := startConfigWatchEvent(s, configPath, lsproto.FileChangeTypeChanged)
	loadMessage := nextConfigReverseRequest(t, outgoing, methodLoadConfigs)
	loadRequest, loadResponse := loadedConfigResponse(t, loadMessage, config.RslintConfig{{
		Rules: config.Rules{"no-debugger": "error"},
	}})
	if got := loadRequest.Candidates[0].ConfigDirectory; got != ancestor {
		t.Fatalf("changed candidate directory=%q, want ancestor %q", got, ancestor)
	}
	respondToConfigReverseRequest(t, s, loadMessage, loadResponse, nil)
	completeConfigWatchActivation(t, s, outgoing)
	if err := awaitConfigWatchEvent(t, changed); err != nil {
		t.Fatalf("changed ancestor config refresh failed: %v", err)
	}
	if got, ok := configRuleValue(s.jsConfigs[ancestor], "no-debugger"); !ok || got != "error" {
		t.Fatalf("changed ancestor config was not committed: %+v", s.jsConfigs)
	}

	if err := os.Remove(configPath); err != nil {
		t.Fatal(err)
	}
	deleted := startConfigWatchEvent(s, configPath, lsproto.FileChangeTypeDeleted)
	emptyLoad := nextConfigReverseRequest(t, outgoing, methodLoadConfigs)
	emptyRequest, ok := emptyLoad.Params.(discovery.ConfigLoadBatchRequest)
	if !ok {
		t.Fatalf("empty load params type=%T", emptyLoad.Params)
	}
	if emptyRequest.Candidates == nil || len(emptyRequest.Candidates) != 0 {
		t.Fatalf("deleted ancestor load candidates=%#v, want []", emptyRequest.Candidates)
	}
	respondToConfigReverseRequest(t, s, emptyLoad, discovery.ConfigLoadBatchResponse{
		TransactionID: emptyRequest.TransactionID,
		Results:       []discovery.ConfigLoadResult{},
	}, nil)
	completeConfigWatchActivation(t, s, outgoing)
	if err := awaitConfigWatchEvent(t, deleted); err != nil {
		t.Fatalf("deleted ancestor config refresh failed: %v", err)
	}
	if len(s.jsConfigs) != 0 {
		t.Fatalf("deleted ancestor config remained committed: %+v", s.jsConfigs)
	}
}

func TestAncestorJSConfigWatcherDiscoversNewNearerConfig(t *testing.T) {
	config.RegisterAllRules()
	root := tspath.NormalizePath(t.TempDir())
	outer := tspath.NormalizePath(filepath.Join(root, "repo"))
	nearer := tspath.NormalizePath(filepath.Join(outer, "packages"))
	workspace := tspath.NormalizePath(filepath.Join(nearer, "app"))
	if err := os.MkdirAll(workspace, 0o755); err != nil {
		t.Fatal(err)
	}
	writeConfigCandidate(t, outer)
	nearerConfigPath := filepath.Join(nearer, "rslint.config.ts")
	if err := os.WriteFile(nearerConfigPath, []byte("export default [];\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	s, outgoing := newAncestorConfigWatchTestServer(workspace)
	installLastGoodConfig(s, outer)
	s.configDiscoveryActive = true
	created := startConfigWatchEvent(s, nearerConfigPath, lsproto.FileChangeTypeCreated)

	nearerLoad := nextConfigReverseRequest(t, outgoing, methodLoadConfigs)
	nearerRequest, nearerResponse := loadedConfigResponse(t, nearerLoad, config.RslintConfig{{
		Rules: config.Rules{"no-debugger": "error"},
	}})
	if got := nearerRequest.Candidates[0].ConfigDirectory; got != nearer {
		t.Fatalf("nearer candidate directory=%q, want %q", got, nearer)
	}
	respondToConfigReverseRequest(t, s, nearerLoad, nearerResponse, nil)
	completeConfigWatchActivation(t, s, outgoing)
	if err := awaitConfigWatchEvent(t, created); err != nil {
		t.Fatalf("created nearer ancestor config refresh failed: %v", err)
	}

	fileURI := documentURIFromPath(filepath.Join(workspace, "src", "index.ts"))
	owner, ok := s.nearestJSConfigKey(fileURI)
	if !ok || owner != nearer {
		t.Fatalf("workspace config owner=%q, ok=%t; want nearer ancestor %q", owner, ok, nearer)
	}
	if len(s.jsConfigs) != 1 {
		t.Fatalf("ancestor lookup loaded non-nearest configs: %+v", s.jsConfigs)
	}
	if got, ok := configRuleValue(s.jsConfigs[nearer], "no-debugger"); !ok || got != "error" {
		t.Fatalf("nearer ancestor config was not committed: %+v", s.jsConfigs)
	}
}

func TestAncestorGitignoreWatcherRefreshesAncestorOwnedConfig(t *testing.T) {
	config.RegisterAllRules()
	root := tspath.NormalizePath(t.TempDir())
	ancestor := tspath.NormalizePath(filepath.Join(root, "repo"))
	intermediate := tspath.NormalizePath(filepath.Join(ancestor, "packages"))
	workspace := tspath.NormalizePath(filepath.Join(intermediate, "app"))
	target := tspath.NormalizePath(filepath.Join(workspace, "src", "index.ts"))
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, []byte("debugger;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeConfigCandidate(t, ancestor)

	s, outgoing := newAncestorConfigWatchTestServer(workspace)
	installLastGoodConfig(s, ancestor)
	s.configDiscoveryActive = true
	uri := documentURIFromPath(target)
	selection, err := s.resolveLintConfigForURI(context.Background(), uri)
	if err != nil {
		t.Fatal(err)
	}
	if selection.merged == nil {
		t.Fatal("target was ignored before the ancestor-chain .gitignore existed")
	}

	gitignorePath := filepath.Join(intermediate, ".gitignore")
	if err := os.WriteFile(gitignorePath, []byte("app/src/index.ts\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	changed := startConfigWatchEvent(s, gitignorePath, lsproto.FileChangeTypeCreated)
	loadMessage := nextConfigReverseRequest(t, outgoing, methodLoadConfigs)
	loadRequest, loadResponse := loadedConfigResponse(t, loadMessage, config.RslintConfig{{
		Rules: config.Rules{"no-debugger": "error"},
	}})
	if got := loadRequest.Candidates[0].ConfigDirectory; got != ancestor {
		t.Fatalf("ancestor candidate directory=%q, want %q", got, ancestor)
	}
	respondToConfigReverseRequest(t, s, loadMessage, loadResponse, nil)
	completeConfigWatchActivation(t, s, outgoing)
	if err := awaitConfigWatchEvent(t, changed); err != nil {
		t.Fatalf("ancestor-chain .gitignore refresh failed: %v", err)
	}

	selection, err = s.resolveLintConfigForURI(context.Background(), uri)
	if err != nil {
		t.Fatal(err)
	}
	if selection.cwd != ancestor {
		t.Fatalf("effective config directory=%q, want %q", selection.cwd, ancestor)
	}
	if selection.merged != nil {
		t.Fatal("ancestor-chain .gitignore was not committed into the active snapshot")
	}
}

func TestConfigRefreshPreservesAtomicJSONLastGoodOnJSFailure(t *testing.T) {
	config.RegisterAllRules()
	workspace := tspath.NormalizePath(t.TempDir())
	writeConfigCandidate(t, workspace)
	jsonPath := filepath.Join(workspace, "rslint.json")
	if err := os.WriteFile(jsonPath, []byte(`[{"rules":{"no-console":"error"}}]`), 0o644); err != nil {
		t.Fatal(err)
	}

	s, outgoing := newAncestorConfigWatchTestServer(workspace)
	installLastGoodConfig(s, workspace)
	s.configDiscoveryActive = true
	s.jsonConfig = config.RslintConfig{{Rules: config.Rules{"no-console": "error"}}}
	s.rslintConfigPath = jsonPath
	if err := os.WriteFile(jsonPath, []byte(`[{"rules":{"no-debugger":"error"}}]`), 0o644); err != nil {
		t.Fatal(err)
	}

	changed := startConfigRefreshForTest(s, "config-change")
	loadMessage := nextConfigReverseRequest(t, outgoing, methodLoadConfigs)
	_, loadResponse := failedConfigResponse(t, loadMessage, "committed JS boundary is now broken")
	respondToConfigReverseRequest(t, s, loadMessage, loadResponse, nil)
	abortMessage := nextConfigReverseRequest(t, outgoing, methodAbortConfigs)
	respondToConfigReverseRequest(t, s, abortMessage, abortResponseForRequest(t, abortMessage), nil)
	completed := awaitConfigRefreshResult(t, changed)
	if !errors.Is(completed.err, discovery.ErrAllConfigsFailed) {
		t.Fatalf("config refresh error = %v, want ErrAllConfigsFailed after preserving last-good", completed.err)
	}

	if len(s.jsonConfig) != 1 || s.jsonConfig[0].Rules["no-console"] != "error" ||
		s.jsonConfig[0].Rules["no-debugger"] != nil {
		t.Fatalf("rejected refresh mutated live JSON fallback: %+v", s.jsonConfig)
	}
	if s.rslintConfigPath != jsonPath {
		t.Fatalf("JSON fallback path=%q, want last-good %q", s.rslintConfigPath, jsonPath)
	}
	if got, ok := configRuleValue(s.jsConfigs[workspace], "no-console"); !ok || got != "error" {
		t.Fatalf("rejected refresh mutated live JS config: %+v", s.jsConfigs)
	}
}

func newAncestorConfigWatchTestServer(workspace string) (*Server, <-chan *lsproto.Message) {
	s, outgoing := newTestServerWithQueue()
	s.cwd = workspace
	s.fs = bundled.WrapFS(osvfs.FS())
	s.pendingServerRequests = make(map[jsonrpc.ID]chan *lsproto.ResponseMessage)
	return s, outgoing
}

func startConfigWatchEvent(
	s *Server,
	configPath string,
	changeType lsproto.FileChangeType,
) <-chan error {
	result := make(chan error, 1)
	go func() {
		result <- s.handleDidChangeWatchedFiles(context.Background(), &lsproto.DidChangeWatchedFilesParams{
			Changes: []*lsproto.FileEvent{{
				Uri:  documentURIFromPath(configPath),
				Type: changeType,
			}},
		})
	}()
	return result
}

func completeConfigWatchActivation(
	t *testing.T,
	s *Server,
	outgoing <-chan *lsproto.Message,
) {
	t.Helper()
	activationMessage := nextConfigReverseRequest(t, outgoing, methodActivateConfigs)
	_, activationResponse := activationResponseForRequest(t, activationMessage, true)
	respondToConfigReverseRequest(t, s, activationMessage, activationResponse, nil)
	commitMessage := nextConfigReverseRequest(t, outgoing, methodCommitConfigs)
	respondToConfigReverseRequest(t, s, commitMessage, commitResponseForRequest(t, commitMessage, true), nil)
}

func awaitConfigWatchEvent(t *testing.T, result <-chan error) error {
	t.Helper()
	select {
	case err := <-result:
		return err
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for config watcher refresh")
		return nil
	}
}
