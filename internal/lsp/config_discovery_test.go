package lsp

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/microsoft/TypeScript/tsc/shim/bundled"
	"github.com/microsoft/TypeScript/tsc/shim/jsonrpc"
	"github.com/microsoft/TypeScript/tsc/shim/lsp/lsproto"
	"github.com/microsoft/TypeScript/tsc/shim/project"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"

	"github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/config/discovery"
	"github.com/web-infra-dev/rslint/internal/config/target"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils/pathutil"
)

type configRefreshTestResult struct {
	response configRefreshResponse
	err      error
}

func configRuleValue(entries config.RslintConfig, name string) (any, bool) {
	for index := len(entries) - 1; index >= 0; index-- {
		if value, ok := entries[index].Rules[name]; ok {
			return value, true
		}
	}
	return nil, false
}

func hasPublicConfigContent(entries config.RslintConfig) bool {
	for _, entry := range entries {
		if entry.Name != "" || len(entry.Files) != 0 || len(entry.FilePatternGroups) != 0 ||
			len(entry.Ignores) != 0 || entry.LanguageOptions != nil || entry.Rules != nil ||
			len(entry.Plugins) != 0 || entry.Settings != nil {
			return true
		}
	}
	return false
}

type configBoundaryIdentityFS struct {
	vfs.FS
	caseSensitive bool
	realPaths     map[string]string
}

func (fs *configBoundaryIdentityFS) CaseSensitivity() tspath.CaseSensitivity {
	return pathutil.CaseSensitivity(fs.caseSensitive)
}

func (fs *configBoundaryIdentityFS) Realpath(filePath tspath.RootedPath) tspath.RootedPath {
	filePath = pathutil.Path(tspath.NormalizePath(filePath.AsString()))
	if realPath := fs.realPaths[filePath.AsString()]; realPath != "" {
		return pathutil.Path(realPath)
	}
	return filePath
}

func newConfigRefreshTestServer(t *testing.T) (*Server, <-chan *lsproto.Message, string) {
	t.Helper()
	root := tspath.NormalizePath(t.TempDir())
	s, outgoing := newTestServerWithQueue()
	s.cwd = root
	s.fs = bundled.WrapFS(osvfs.FS())
	s.pendingServerRequests = make(map[jsonrpc.ID]chan *lsproto.ResponseMessage)
	return s, outgoing, root
}

func startConfigRefreshForTest(s *Server, reason string) <-chan configRefreshTestResult {
	return startConfigRefreshRequestForTest(s, configRefreshRequest{
		ProtocolVersion: discovery.ConfigDiscoveryProtocolVersion,
		Reason:          reason,
	})
}

func startExplicitConfigRefreshForTest(s *Server, reason string, configPath string) <-chan configRefreshTestResult {
	return startConfigRefreshRequestForTest(s, configRefreshRequest{
		ProtocolVersion: discovery.ConfigDiscoveryProtocolVersion,
		Reason:          reason,
		ConfigPath: optionalConfigRefreshPath{
			Value:   configPath,
			Present: true,
		},
	})
}

func startConfigRefreshRequestForTest(s *Server, request configRefreshRequest) <-chan configRefreshTestResult {
	result := make(chan configRefreshTestResult, 1)
	go func() {
		response, err := s.handleConfigRefresh(context.Background(), request)
		result <- configRefreshTestResult{response: response, err: err}
	}()
	return result
}

func nextConfigReverseRequest(
	t *testing.T,
	outgoing <-chan *lsproto.Message,
	wantMethod lsproto.Method,
) *lsproto.RequestMessage {
	t.Helper()
	select {
	case message := <-outgoing:
		request := message.AsRequest()
		if request.Method != wantMethod {
			t.Fatalf("reverse request method = %q, want %q", request.Method, wantMethod)
		}
		if request.ID == nil {
			t.Fatalf("reverse request %q has no ID", request.Method)
		}
		return request
	case <-time.After(5 * time.Second):
		t.Fatalf("timed out waiting for reverse request %q", wantMethod)
		return nil
	}
}

func respondToConfigReverseRequest(
	t *testing.T,
	s *Server,
	request *lsproto.RequestMessage,
	result any,
	responseErr error,
) {
	t.Helper()
	response := &lsproto.ResponseMessage{ID: request.ID, Result: result}
	if responseErr != nil {
		response.Error = &jsonrpc.ResponseError{
			Code:    int32(lsproto.ErrorCodeInternalError),
			Message: responseErr.Error(),
		}
	}
	s.pendingServerRequestsMu.Lock()
	responseCh, ok := s.pendingServerRequests[*request.ID]
	if ok {
		responseCh <- response
		close(responseCh)
		delete(s.pendingServerRequests, *request.ID)
	}
	s.pendingServerRequestsMu.Unlock()
	if !ok {
		t.Fatalf("reverse request %q was not pending", request.Method)
	}
}

func loadedConfigResponse(
	t *testing.T,
	request *lsproto.RequestMessage,
	entries config.RslintConfig,
) (discovery.ConfigLoadBatchRequest, discovery.ConfigLoadBatchResponse) {
	t.Helper()
	loadRequest, ok := request.Params.(discovery.ConfigLoadBatchRequest)
	if !ok {
		t.Fatalf("loadConfigs params type = %T", request.Params)
	}
	if len(loadRequest.Candidates) != 1 {
		t.Fatalf("loadConfigs candidate count = %d, want 1", len(loadRequest.Candidates))
	}
	return loadRequest, discovery.ConfigLoadBatchResponse{
		TransactionID: loadRequest.TransactionID,
		Results: []discovery.ConfigLoadResult{{
			ID:      loadRequest.Candidates[0].ID,
			Status:  "loaded",
			Entries: entries,
		}},
	}
}

func failedConfigResponse(
	t *testing.T,
	request *lsproto.RequestMessage,
	message string,
) (discovery.ConfigLoadBatchRequest, discovery.ConfigLoadBatchResponse) {
	t.Helper()
	loadRequest, ok := request.Params.(discovery.ConfigLoadBatchRequest)
	if !ok {
		t.Fatalf("loadConfigs params type = %T", request.Params)
	}
	results := make([]discovery.ConfigLoadResult, len(loadRequest.Candidates))
	for index, candidate := range loadRequest.Candidates {
		results[index] = discovery.ConfigLoadResult{
			ID:     candidate.ID,
			Status: "failed",
			Error: &discovery.ConfigModuleError{
				Code:    "load",
				Message: message,
			},
		}
	}
	return loadRequest, discovery.ConfigLoadBatchResponse{
		TransactionID: loadRequest.TransactionID,
		Results:       results,
	}
}

func activationResponseForRequest(
	t *testing.T,
	request *lsproto.RequestMessage,
	pluginHostReady bool,
) (discovery.ConfigActivationRequest, configActivationWireResponse) {
	t.Helper()
	activation, ok := request.Params.(discovery.ConfigActivationRequest)
	if !ok {
		t.Fatalf("activateConfigs params type = %T", request.Params)
	}
	return activation, configActivationWireResponse{
		TransactionID:   activation.TransactionID,
		PluginHostReady: pluginHostReady,
	}
}

func commitResponseForRequest(
	t *testing.T,
	request *lsproto.RequestMessage,
	committed bool,
) configCommitWireResponse {
	t.Helper()
	control, ok := request.Params.(configTransactionControlRequest)
	if !ok {
		t.Fatalf("commitConfigs params type = %T", request.Params)
	}
	return configCommitWireResponse{
		TransactionID: control.TransactionID,
		Committed:     committed,
	}
}

func abortResponseForRequest(t *testing.T, request *lsproto.RequestMessage) configAbortWireResponse {
	t.Helper()
	control, ok := request.Params.(configTransactionControlRequest)
	if !ok {
		t.Fatalf("abortConfigs params type = %T", request.Params)
	}
	return configAbortWireResponse{
		TransactionID: control.TransactionID,
		Aborted:       true,
	}
}

func awaitConfigRefreshResult(t *testing.T, result <-chan configRefreshTestResult) configRefreshTestResult {
	t.Helper()
	select {
	case value := <-result:
		return value
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for configRefresh result")
		return configRefreshTestResult{}
	}
}

func completeSuccessfulConfigRefreshForTest(
	t *testing.T,
	s *Server,
	outgoing <-chan *lsproto.Message,
	entries config.RslintConfig,
) discovery.ConfigLoadBatchRequest {
	t.Helper()
	loadMessage := nextConfigReverseRequest(t, outgoing, methodLoadConfigs)
	loadRequest, loadResponse := loadedConfigResponse(t, loadMessage, entries)
	respondToConfigReverseRequest(t, s, loadMessage, loadResponse, nil)
	activationMessage := nextConfigReverseRequest(t, outgoing, methodActivateConfigs)
	_, activationResponse := activationResponseForRequest(t, activationMessage, true)
	respondToConfigReverseRequest(t, s, activationMessage, activationResponse, nil)
	commitMessage := nextConfigReverseRequest(t, outgoing, methodCommitConfigs)
	respondToConfigReverseRequest(t, s, commitMessage, commitResponseForRequest(t, commitMessage, true), nil)
	return loadRequest
}

func writeConfigCandidate(t *testing.T, root string) {
	t.Helper()
	if err := os.WriteFile(
		filepath.Join(root, "rslint.config.mjs"),
		[]byte("export default [];\n"),
		0o644,
	); err != nil {
		t.Fatal(err)
	}
}

func installLastGoodConfig(s *Server, root string) {
	entries := config.RslintConfig{{
		Plugins: []string{"last-good"},
		Rules: config.Rules{
			"no-console":      "error",
			"last-good/check": "error",
		},
	}}
	s.gitPolicy = config.CollectGitPolicy(entries, root, s.fs, nil, config.ResolveDirectoryIdentities([]string{root}, s.fs))
	installJSConfigsForTest(s, map[string]config.RslintConfig{root: entries})
	s.eslintPluginConfigGeneration = "last-good"
	s.ruleCatalog, _ = deriveLSPRuleCatalog(s.currentRuleCatalog(), []config.EslintPluginEntry{{
		Prefix:    "last-good",
		RuleNames: []string{"check"},
	}})
	s.configDiscoveryHasLastGood = true
	// Tests using this helper model a catalog committed by an earlier automatic
	// initial refresh, then exercise a later config/dependency/watch refresh.
	s.configRefreshInitialized = true
	s.configRefreshConfigPath = ""
}

func assertLastGoodRuleCatalog(t *testing.T, s *Server) {
	t.Helper()
	if ruleImpl, ok := s.ruleCatalog.Lookup("last-good/check"); !ok || !ruleImpl.IsEslintPluginRule {
		t.Fatalf("last-good plugin catalog was replaced: %+v", ruleImpl)
	}
}

func TestHandleConfigRefreshCommitsFilesystemPathCatalog(t *testing.T) {
	s, outgoing, root := newConfigRefreshTestServer(t)
	writeConfigCandidate(t, root)

	result := startConfigRefreshForTest(s, "initial")
	loadMessage := nextConfigReverseRequest(t, outgoing, methodLoadConfigs)
	loadRequest, loadResponse := loadedConfigResponse(t, loadMessage, config.RslintConfig{{
		Rules: config.Rules{"no-debugger": "error"},
	}})
	if loadRequest.LoadMode != discovery.ConfigModuleLoadFresh {
		t.Fatalf("load mode = %q, want fresh", loadRequest.LoadMode)
	}
	if got := loadRequest.Candidates[0].ConfigDirectory; got != root {
		t.Fatalf("configDirectory = %q, want filesystem path %q", got, root)
	}
	respondToConfigReverseRequest(t, s, loadMessage, loadResponse, nil)

	activationMessage := nextConfigReverseRequest(t, outgoing, methodActivateConfigs)
	activationRequest, activationResponse := activationResponseForRequest(t, activationMessage, true)
	if activationRequest.TransactionID != loadRequest.TransactionID || len(activationRequest.EffectiveConfigIDs) != 1 {
		t.Fatalf("unexpected activation request: %+v", activationRequest)
	}
	const pluginRuleName = "config-refresh-test/no-foo"
	activationResponse.EslintPluginEntries = []config.EslintPluginEntry{{
		Prefix:    "config-refresh-test",
		RuleNames: []string{"no-foo"},
	}}
	respondToConfigReverseRequest(t, s, activationMessage, activationResponse, nil)

	commitMessage := nextConfigReverseRequest(t, outgoing, methodCommitConfigs)
	commitResponse := commitResponseForRequest(t, commitMessage, true)
	respondToConfigReverseRequest(t, s, commitMessage, commitResponse, nil)

	completed := awaitConfigRefreshResult(t, result)
	if completed.err != nil {
		t.Fatalf("configRefresh failed: %v", completed.err)
	}
	if completed.response.TransactionID != loadRequest.TransactionID || len(s.jsConfigs) != 1 {
		t.Fatalf("unexpected configRefresh response: %+v", completed.response)
	}
	if s.eslintPluginConfigGeneration != loadRequest.TransactionID {
		t.Fatalf("committed generation = %q, want %q", s.eslintPluginConfigGeneration, loadRequest.TransactionID)
	}
	entries, ok := s.jsConfigs[root]
	if value, found := configRuleValue(entries, "no-debugger"); !ok || !found || value != "error" {
		t.Fatalf("filesystem-path catalog was not committed: %+v", s.jsConfigs)
	}
	fileURI := documentURIFromPath(filepath.Join(root, "src", "index.ts"))
	if got, ok := s.nearestJSConfigKey(fileURI); !ok || got != root {
		t.Fatalf("plugin configKey = %q, want exact catalog path %q", got, root)
	}
	registered, ok := s.ruleCatalog.Lookup(pluginRuleName)
	if !ok || !registered.IsEslintPluginRule {
		t.Fatalf("committed rule catalog does not contain plugin rule %q: %+v", pluginRuleName, registered)
	}
}

func TestConfigRefreshRequestPreservesConfigPathPresence(t *testing.T) {
	tests := []struct {
		name        string
		params      string
		wantPresent bool
		wantPath    string
		wantError   bool
	}{
		{name: "absent", params: `{"protocolVersion":2,"reason":"initial"}`},
		{name: "string", params: `{"protocolVersion":2,"reason":"initial","configPath":"/repo/custom.config.mjs"}`, wantPresent: true, wantPath: "/repo/custom.config.mjs"},
		{name: "null", params: `{"protocolVersion":2,"reason":"initial","configPath":null}`, wantError: true},
		{name: "number", params: `{"protocolVersion":2,"reason":"initial","configPath":42}`, wantError: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			data := []byte(`{"jsonrpc":"2.0","id":1,"method":"rslint/configRefresh","params":` + test.params + `}`)
			var message lsproto.Message
			if err := json.Unmarshal(data, &message); err != nil {
				t.Fatalf("unmarshal request: %v", err)
			}
			request, err := decodeParams[configRefreshRequest](message.AsRequest())
			if test.wantError {
				if err == nil {
					t.Fatal("invalid configPath decoded successfully")
				}
				return
			}
			if err != nil {
				t.Fatalf("decode configRefresh: %v", err)
			}
			if request.ConfigPath.Present != test.wantPresent || request.ConfigPath.Value != test.wantPath {
				t.Fatalf("decoded configPath = %+v", request.ConfigPath)
			}
		})
	}
}

func TestNormalizeLSPExplicitConfigPath(t *testing.T) {
	root := t.TempDir()
	for _, extension := range []string{".js", ".mjs", ".cjs", ".ts", ".mts", ".cts"} {
		configPath := filepath.Join(root, "custom.config"+extension)
		if got, err := normalizeLSPExplicitConfigPath(configPath); err != nil || got != tspath.NormalizePath(configPath) {
			t.Fatalf("normalize %q = %q, %v", configPath, got, err)
		}
	}
	for _, configPath := range []string{
		"",
		"relative.config.mjs",
		"file:///repo/custom.config.mjs",
		filepath.Join(root, "rslint.jsonc"),
		filepath.Join(root, "custom.config.MJS"),
		filepath.Join(root, "custom.config.mjs") + "\x00suffix",
	} {
		if _, err := normalizeLSPExplicitConfigPath(configPath); err == nil {
			t.Fatalf("invalid configPath %q was accepted", configPath)
		}
	}
}

func TestHandleConfigRefreshRequiresInitialBeforeMutation(t *testing.T) {
	s, _, _ := newConfigRefreshTestServer(t)
	_, err := s.handleConfigRefresh(context.Background(), configRefreshRequest{
		ProtocolVersion: discovery.ConfigDiscoveryProtocolVersion,
		Reason:          "config-change",
	})
	if err == nil || !errors.Is(err, lsproto.ErrorCodeInvalidParams) {
		t.Fatalf("configRefresh error = %v, want InvalidParams", err)
	}
	if s.configRefreshInitialized || s.configDiscoveryActive {
		t.Fatal("non-initial refresh mutated config source state")
	}
}

func TestHandleConfigRefreshUsesFixedExplicitConfig(t *testing.T) {
	s, outgoing, root := newConfigRefreshTestServer(t)
	writeConfigCandidate(t, root)
	if err := os.WriteFile(filepath.Join(root, "rslint.jsonc"), []byte("{"), 0o644); err != nil {
		t.Fatal(err)
	}
	externalRoot := t.TempDir()
	configPath := filepath.Join(externalRoot, "generated-rstack-config.cjs")
	if err := os.WriteFile(configPath, []byte("module.exports = [];\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	targetPath := filepath.Join(root, "src", "index.ts")
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(targetPath, []byte("debugger;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	relativeTarget, err := filepath.Rel(externalRoot, targetPath)
	if err != nil {
		t.Fatal(err)
	}

	initial := startExplicitConfigRefreshForTest(s, "initial", configPath)
	loadRequest := completeSuccessfulConfigRefreshForTest(t, s, outgoing, config.RslintConfig{{
		Files: []string{filepath.ToSlash(relativeTarget)},
		Rules: config.Rules{"no-debugger": "error"},
	}})
	if completed := awaitConfigRefreshResult(t, initial); completed.err != nil {
		t.Fatalf("explicit initial refresh failed: %v", completed.err)
	}
	wantModulePath := tspath.NormalizePath(osvfs.FS().Realpath(pathutil.Path(configPath)).AsString())
	if len(loadRequest.Candidates) != 1 || loadRequest.Candidates[0].ConfigPath != wantModulePath {
		t.Fatalf("explicit candidates = %+v, want only %q", loadRequest.Candidates, wantModulePath)
	}
	configDir := tspath.NormalizePath(externalRoot)
	if loadRequest.Candidates[0].ConfigDirectory != configDir {
		t.Fatalf("explicit configDirectory = %q, want config directory %q", loadRequest.Candidates[0].ConfigDirectory, configDir)
	}
	if !s.configRefreshInitialized || s.configRefreshConfigPath != tspath.NormalizePath(configPath) {
		t.Fatalf("fixed config path = %q, initialized=%t", s.configRefreshConfigPath, s.configRefreshInitialized)
	}
	if s.fallbackConfigOwnerIndex != nil || len(s.fallbackConfig) != 0 {
		t.Fatalf("explicit config retained a workspace fallback: %+v", s.fallbackConfig)
	}
	if value, found := configRuleValue(s.jsConfigs[configDir], "no-debugger"); !found || value != "error" {
		t.Fatalf("explicit config was not committed at its directory: %+v", s.jsConfigs)
	}
	entries, resolvedDirectory, isJSConfig := s.getConfigForURI(documentURIFromPath(targetPath))
	if !isJSConfig || resolvedDirectory != configDir ||
		config.NewFileConfigResolverWithFS(entries, resolvedDirectory, s.fs, s.currentRuleCatalog()).ConfigForFile(targetPath) == nil {
		t.Fatalf("explicit external config did not govern workspace target from its authored base")
	}

	reload := startExplicitConfigRefreshForTest(s, "config-change", configPath)
	reloadRequest := completeSuccessfulConfigRefreshForTest(t, s, outgoing, config.RslintConfig{{
		Rules: config.Rules{"no-console": "error"},
	}})
	if completed := awaitConfigRefreshResult(t, reload); completed.err != nil {
		t.Fatalf("same-path refresh failed: %v", completed.err)
	}
	if len(reloadRequest.Candidates) != 1 || reloadRequest.Candidates[0].ConfigPath != wantModulePath {
		t.Fatalf("same-path candidates = %+v", reloadRequest.Candidates)
	}

	gitignorePath := filepath.Join(root, ".gitignore")
	if err := os.WriteFile(gitignorePath, []byte("generated/\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	watchResult := make(chan error, 1)
	go func() {
		watchResult <- s.handleDidChangeWatchedFiles(context.Background(), &lsproto.DidChangeWatchedFilesParams{
			Changes: []*lsproto.FileEvent{{
				Uri:  documentURIFromPath(gitignorePath),
				Type: lsproto.FileChangeTypeChanged,
			}},
		})
	}()
	watchRequest := completeSuccessfulConfigRefreshForTest(t, s, outgoing, config.RslintConfig{{
		Rules: config.Rules{"no-console": "error"},
	}})
	if err := <-watchResult; err != nil {
		t.Fatalf("explicit .gitignore refresh failed: %v", err)
	}
	if len(watchRequest.Candidates) != 1 || watchRequest.Candidates[0].ConfigPath != wantModulePath {
		t.Fatalf("explicit watcher candidates = %+v", watchRequest.Candidates)
	}

	for _, request := range []configRefreshRequest{
		{ProtocolVersion: discovery.ConfigDiscoveryProtocolVersion, Reason: "config-change"},
		{
			ProtocolVersion: discovery.ConfigDiscoveryProtocolVersion,
			Reason:          "config-change",
			ConfigPath: optionalConfigRefreshPath{
				Value:   filepath.Join(externalRoot, "other.config.cjs"),
				Present: true,
			},
		},
	} {
		if _, err := s.handleConfigRefresh(context.Background(), request); err == nil || !errors.Is(err, lsproto.ErrorCodeInvalidParams) {
			t.Fatalf("changed selection error = %v, want InvalidParams", err)
		}
	}
	select {
	case message := <-outgoing:
		t.Fatalf("rejected selection started reverse request: %+v", message)
	default:
	}
}

func TestHandleConfigRefreshKeepsExplicitPathAfterInitialLoadFailure(t *testing.T) {
	s, outgoing, _ := newConfigRefreshTestServer(t)
	configPath := filepath.Join(t.TempDir(), "generated-rstack-config.mjs")
	if err := os.WriteFile(configPath, []byte("export default [];\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	initial := startExplicitConfigRefreshForTest(s, "initial", configPath)
	loadMessage := nextConfigReverseRequest(t, outgoing, methodLoadConfigs)
	loadRequest, loadResponse := failedConfigResponse(t, loadMessage, "synthetic explicit load failure")
	respondToConfigReverseRequest(t, s, loadMessage, loadResponse, nil)
	activationMessage := nextConfigReverseRequest(t, outgoing, methodActivateConfigs)
	_, activationResponse := activationResponseForRequest(t, activationMessage, true)
	respondToConfigReverseRequest(t, s, activationMessage, activationResponse, nil)
	commitMessage := nextConfigReverseRequest(t, outgoing, methodCommitConfigs)
	respondToConfigReverseRequest(t, s, commitMessage, commitResponseForRequest(t, commitMessage, true), nil)
	completed := awaitConfigRefreshResult(t, initial)
	if completed.err != nil {
		t.Fatalf("initial explicit failure stopped the LSP: %v", completed.err)
	}
	if completed.response.TransactionID != loadRequest.TransactionID {
		t.Fatalf("explicit recovery response = %+v", completed.response)
	}
	if s.configRefreshConfigPath != tspath.NormalizePath(configPath) || !s.configRefreshInitialized {
		t.Fatalf("failed initial load lost fixed path: path=%q initialized=%t", s.configRefreshConfigPath, s.configRefreshInitialized)
	}
	configDir := tspath.NormalizePath(filepath.Dir(configPath))
	if _, unavailable := s.jsUnavailableConfigs[configDir]; !unavailable || s.configDiscoveryHasLastGood {
		t.Fatalf("explicit failure state: unavailable=%t lastGood=%t", unavailable, s.configDiscoveryHasLastGood)
	}
	if _, err := s.handleConfigRefresh(context.Background(), configRefreshRequest{
		ProtocolVersion: discovery.ConfigDiscoveryProtocolVersion,
		Reason:          "config-change",
	}); err == nil || !errors.Is(err, lsproto.ErrorCodeInvalidParams) {
		t.Fatalf("explicit path omission error = %v, want InvalidParams", err)
	}
	select {
	case message := <-outgoing:
		t.Fatalf("omitted explicit path started reverse request: %+v", message)
	default:
	}
}

func TestPrepareDiscoveredConfigSnapshotSharesGitPolicyAcrossOwners(t *testing.T) {
	root := tspath.NormalizePath(t.TempDir())
	child := tspath.ResolvePath(root, "packages/app")
	if err := os.MkdirAll(child, 0o755); err != nil {
		t.Fatal(err)
	}
	for path, content := range map[string]string{
		tspath.ResolvePath(root, ".gitignore"):  "ignored.ts\n",
		tspath.ResolvePath(root, "ignored.ts"):  "debugger;\n",
		tspath.ResolvePath(child, "ignored.ts"): "debugger;\n",
	} {
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	fsys := bundled.WrapFS(osvfs.FS())
	policy := config.CollectGitPolicy(nil, root, fsys, nil, config.ResolveDirectoryIdentities([]string{root}, fsys))
	catalog := &discovery.ConfigCatalog{
		TransactionID: "snapshot-boundaries",
		GitPolicy:     policy,
		Configs: map[string]config.RslintConfig{
			root:  {{Rules: config.Rules{"no-console": "error"}}},
			child: {{Rules: config.Rules{"no-debugger": "error"}}},
		},
	}
	s := newTestServer()
	s.cwd, s.fs = root, fsys
	if catalog.GitPolicy == nil {
		catalog.GitPolicy = config.CollectGitPolicy(nil, root, fsys, nil, config.ResolveDirectoryIdentities([]string{root}, fsys))
	}
	prepared, err := s.prepareDiscoveredConfigSnapshot(fsys, catalog)
	if err != nil {
		t.Fatal(err)
	}
	snapshot, err := completeDiscoveredConfigSnapshot(prepared, nil, fsys)
	if err != nil {
		t.Fatal(err)
	}
	s.commitDiscoveredConfigSnapshot(context.Background(), snapshot)
	for _, directory := range []string{root, child} {
		file := tspath.ResolvePath(directory, "ignored.ts")
		document := s.documentLintSnapshot(documentURIFromPath(file))
		if document.gitPolicy != policy || !document.resolvedConfig.GloballyIgnored || document.unavailable {
			t.Fatalf("owner %q did not share the complete Git policy: %+v", directory, document)
		}
		if catalog.Configs[directory].IsFileIgnored(file, directory) {
			t.Fatal("Git policy was embedded into authored config")
		}
	}
	if snapshot.gitPolicy != policy {
		t.Fatal("fallback policy generation differs")
	}
}

func TestPrepareDiscoveredConfigSnapshotDefersProjectPathsUntilDocumentMatch(t *testing.T) {
	root := tspath.NormalizePath(t.TempDir())
	fsys := bundled.WrapFS(osvfs.FS())
	catalog := &discovery.ConfigCatalog{
		TransactionID: "document-projects",
		Configs: map[string]config.RslintConfig{root: {
			{Rules: config.Rules{"no-var": "error"}},
			{
				Files: []string{"**/*.ts"},
				LanguageOptions: &config.LanguageOptions{ParserOptions: &config.ParserOptions{
					Project: config.ProjectPaths{"./missing.json"},
				}},
			},
		}},
	}
	s := newTestServer()
	s.cwd, s.fs = root, fsys
	if catalog.GitPolicy == nil {
		catalog.GitPolicy = config.CollectGitPolicy(nil, root, fsys, nil, config.ResolveDirectoryIdentities([]string{root}, fsys))
	}
	prepared, err := s.prepareDiscoveredConfigSnapshot(fsys, catalog)
	if err != nil {
		t.Fatalf("unmatched project prevented config preparation: %v", err)
	}
	completed, err := completeDiscoveredConfigSnapshot(prepared, nil, fsys)
	if err != nil {
		t.Fatal(err)
	}
	s.commitDiscoveredConfigSnapshot(context.Background(), completed)
	javascript := s.documentLintSnapshot(documentURIFromPath(tspath.ResolvePath(root, "script.js")))
	if !javascript.configResolved || javascript.unavailable || javascript.projectPolicyError != nil || len(javascript.typeScriptConfigPaths) != 0 {
		t.Fatalf("unmatched project affected JavaScript snapshot: %+v", javascript)
	}
	typescript := s.documentLintSnapshot(documentURIFromPath(tspath.ResolvePath(root, "source.ts")))
	if typescript.projectPolicyError == nil || !strings.Contains(typescript.projectPolicyError.Error(), "missing.json") {
		t.Fatalf("matched missing project error = %v", typescript.projectPolicyError)
	}
}

func TestHandleConfigRefreshPublishesEffectiveProjectErrors(t *testing.T) {
	for _, test := range []struct {
		name    string
		options *config.ParserOptions
		message string
	}{
		{
			name:    "missing project",
			options: &config.ParserOptions{Project: config.ProjectPaths{"./missing.json"}},
			message: "missing.json",
		},
		{
			name:    "unmatched project glob",
			options: &config.ParserOptions{Project: config.ProjectPaths{"./missing-*.json"}},
			message: "matched no files",
		},
		{
			name: "conflicting project options",
			options: &config.ParserOptions{
				Project: config.ProjectPaths{"./tsconfig.json"}, ProjectService: config.BoolPtr(true),
			},
			message: "projectService",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			s, outgoing, root := newConfigRefreshTestServer(t)
			const source = "export function read() { var value = 1; return value; }\n"
			files := []string{"source.ts", "script.js", "ignored.ts", "node_modules/dependency/index.ts"}
			for _, name := range files {
				filePath := tspath.ResolvePath(root, name)
				if err := os.MkdirAll(tspath.GetDirectoryPath(filePath), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filePath, []byte(source), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			if err := os.WriteFile(tspath.ResolvePath(root, "tsconfig.json"), []byte(`{"compilerOptions":{"noLib":true},"files":["source.ts"]}`), 0o644); err != nil {
				t.Fatal(err)
			}
			s.defaultLibraryPath = bundled.LibPath().AsString()
			if _, err := s.handleInitialize(context.Background(), &lsproto.InitializeParams{}); err != nil {
				t.Fatal(err)
			}
			if err := s.handleInitialized(context.Background(), &lsproto.InitializedParams{}); err != nil {
				t.Fatal(err)
			}
			defer s.session.Close()
			for _, name := range files {
				uri := documentURIFromPath(tspath.ResolvePath(root, name))
				s.documents[uri] = source
				language := lsproto.LanguageKindTypeScript
				if strings.HasSuffix(name, ".js") {
					language = lsproto.LanguageKindJavaScript
				}
				s.session.DidOpenFile(context.Background(), uri, 1, source, language)
			}
			refresh := func(reason string, options *config.ParserOptions) {
				t.Helper()
				entries := config.RslintConfig{
					{Ignores: []string{"ignored.ts"}},
					{Rules: config.Rules{"no-var": "error"}},
					{
						Files:           []string{"**/*.ts"},
						LanguageOptions: &config.LanguageOptions{ParserOptions: options},
					},
				}
				data, err := json.Marshal(entries)
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(tspath.ResolvePath(root, "rslint.config.mjs"), []byte("export default "+string(data)+";\n"), 0o644); err != nil {
					t.Fatal(err)
				}
				result := startConfigRefreshForTest(s, reason)
				completeSuccessfulConfigRefreshForTest(t, s, outgoing, entries)
				if completed := awaitConfigRefreshResult(t, result); completed.err != nil {
					t.Fatalf("refresh failed instead of isolating the document error: %v", completed.err)
				}
			}
			publish := func(name string) []*lsproto.Diagnostic {
				t.Helper()
				uri := documentURIFromPath(tspath.ResolvePath(root, name))
				s.pushDiagnostics(uri)
				select {
				case message := <-outgoing:
					request := message.AsRequest()
					params, ok := request.Params.(*lsproto.PublishDiagnosticsParams)
					if request.Method != "textDocument/publishDiagnostics" || !ok || params.Uri != uri {
						t.Fatalf("unexpected publication: %+v", request)
					}
					return params.Diagnostics
				case <-time.After(5 * time.Second):
					t.Fatal("timed out waiting for diagnostics")
					return nil
				}
			}
			assertRuleDiagnostic := func(name string) {
				t.Helper()
				diagnostics := publish(name)
				if len(diagnostics) != 1 || diagnostics[0].Message.String == nil || !strings.HasPrefix(*diagnostics[0].Message.String, "[no-var]") {
					t.Fatalf("%s: expected no-var diagnostic, got %+v", name, diagnostics)
				}
			}
			valid := &config.ParserOptions{Project: config.ProjectPaths{"./tsconfig.json"}}
			refresh("initial", valid)
			assertRuleDiagnostic("source.ts")
			uri := documentURIFromPath(tspath.ResolvePath(root, "source.ts"))
			if len(s.diagnostics[uri]) != 1 {
				t.Fatal("initial rule diagnostic was not cached")
			}
			previousDiagnostics := s.diagnostics[uri]
			generation := s.eslintPluginConfigGeneration
			refresh("config-change", test.options)
			if generation == s.eslintPluginConfigGeneration {
				t.Fatal("document project error prevented the new config generation from committing")
			}
			// Exercise publication's own stale-fix cleanup, independently of the
			// config refresh transaction's diagnostic invalidation.
			s.diagnostics[uri] = previousDiagnostics
			diagnostics := publish("source.ts")
			if len(diagnostics) != 1 || diagnostics[0].Message.String == nil ||
				!strings.Contains(*diagnostics[0].Message.String, test.message) ||
				!strings.HasPrefix(*diagnostics[0].Message.String, "Project configuration error:") ||
				diagnostics[0].Source == nil || *diagnostics[0].Source != "rslint" ||
				diagnostics[0].Severity == nil || *diagnostics[0].Severity != lsproto.DiagnosticSeverityError ||
				diagnostics[0].Range != (lsproto.Range{}) {
				t.Fatalf("expected a visible project configuration error, got %+v", diagnostics)
			}
			if _, stale := s.diagnostics[uri]; stale {
				t.Fatal("project error retained stale rule diagnostics and fixes")
			}
			actions, err := s.handleCodeAction(context.Background(), &lsproto.CodeActionParams{
				TextDocument: lsproto.TextDocumentIdentifier{Uri: uri},
				Range:        lsproto.Range{End: lsproto.Position{Character: uint32(len(source))}},
			})
			if err != nil || actions.CommandOrCodeActionArray == nil || len(*actions.CommandOrCodeActionArray) != 0 {
				t.Fatalf("configuration error offered rule actions: %+v, %v", actions, err)
			}
			fixes, err := s.handleFixAllCodeAction(context.Background(), uri)
			if err != nil || fixes.CommandOrCodeActionArray == nil || len(*fixes.CommandOrCodeActionArray) != 0 {
				t.Fatalf("invalid project offered fixes: %+v, %v", fixes, err)
			}
			assertRuleDiagnostic("script.js")
			for _, name := range files[2:] {
				if diagnostics := publish(name); len(diagnostics) != 0 {
					t.Fatalf("excluded %s received project diagnostics: %+v", name, diagnostics)
				}
			}
			canceled, cancel := context.WithCancel(context.Background())
			cancel()
			s.backgroundCtx = canceled
			if diagnostics := publish("source.ts"); len(diagnostics) != 0 {
				t.Fatalf("cancellation was mislabeled as a project error: %+v", diagnostics)
			}
			s.backgroundCtx = context.Background()
			refresh("config-change", valid)
			assertRuleDiagnostic("source.ts")
			fixes, err = s.handleFixAllCodeAction(context.Background(), uri)
			if err != nil || fixes.CommandOrCodeActionArray == nil || len(*fixes.CommandOrCodeActionArray) == 0 {
				t.Fatalf("correcting the project did not restore fixes: %+v, %v", fixes, err)
			}
		})
	}
}

func TestCompleteDiscoveredConfigSnapshotBuildsOneResolverPerOwner(t *testing.T) {
	root := tspath.NormalizePath(t.TempDir())
	child := tspath.NormalizePath(filepath.Join(root, "packages", "app"))
	if err := os.MkdirAll(child, 0o755); err != nil {
		t.Fatal(err)
	}
	fsys := bundled.WrapFS(osvfs.FS())
	catalog := &discovery.ConfigCatalog{
		TransactionID: "resolver-generation",
		Configs: map[string]config.RslintConfig{
			root:  {{Rules: config.Rules{"no-console": "error"}}},
			child: {{Rules: config.Rules{"no-debugger": "error"}}},
		},
	}
	s := newTestServer()
	s.cwd = root
	if catalog.GitPolicy == nil {
		catalog.GitPolicy = config.CollectGitPolicy(nil, root, fsys, nil, config.ResolveDirectoryIdentities([]string{root}, fsys))
	}
	prepared, err := s.prepareDiscoveredConfigSnapshot(fsys, catalog)
	if err != nil {
		t.Fatal(err)
	}
	completed, err := completeDiscoveredConfigSnapshot(prepared, nil, fsys)
	if err != nil {
		t.Fatal(err)
	}
	if len(completed.fileConfigResolvers) != len(catalog.Configs) {
		t.Fatalf("file config resolvers = %d, want %d", len(completed.fileConfigResolvers), len(catalog.Configs))
	}
	for _, configDirectory := range []string{root, child} {
		if completed.fileConfigResolvers[configDirectory] == nil {
			t.Fatalf("missing resolver for %q", configDirectory)
		}
	}
	if completed.fallbackFileConfigResolver == nil {
		t.Fatal("automatic config generation is missing its workspace fallback resolver")
	}
}

func TestHandleConfigRefreshInitialPluginHostFailureCommitsNativeCatalog(t *testing.T) {
	s, outgoing, root := newConfigRefreshTestServer(t)
	writeConfigCandidate(t, root)

	result := startConfigRefreshForTest(s, "initial")
	loadMessage := nextConfigReverseRequest(t, outgoing, methodLoadConfigs)
	_, loadResponse := loadedConfigResponse(t, loadMessage, config.RslintConfig{{
		Plugins: []string{"community"},
		Rules: config.Rules{
			"no-debugger":     "error",
			"community/check": "error",
		},
	}})
	respondToConfigReverseRequest(t, s, loadMessage, loadResponse, nil)

	activationMessage := nextConfigReverseRequest(t, outgoing, methodActivateConfigs)
	activationRequest, activationResponse := activationResponseForRequest(t, activationMessage, false)
	activationResponse.EslintPluginEntries = []config.EslintPluginEntry{{
		Prefix:    "community",
		RuleNames: []string{"check"},
	}}
	respondToConfigReverseRequest(t, s, activationMessage, activationResponse, nil)
	commitMessage := nextConfigReverseRequest(t, outgoing, methodCommitConfigs)
	respondToConfigReverseRequest(t, s, commitMessage, commitResponseForRequest(t, commitMessage, true), nil)

	completed := awaitConfigRefreshResult(t, result)
	if completed.err != nil {
		t.Fatalf("initial degraded plugin configRefresh failed: %v", completed.err)
	}
	if !s.configDiscoveryHasLastGood || s.eslintPluginConfigGeneration != activationRequest.TransactionID {
		t.Fatalf("degraded generation was not committed: response=%+v generation=%q hasLastGood=%t", completed.response, s.eslintPluginConfigGeneration, s.configDiscoveryHasLastGood)
	}
	entries := s.jsConfigs[root]
	if value, found := configRuleValue(entries, "no-debugger"); !found || value != "error" {
		t.Fatalf("base rule catalog was lost with plugin host: %+v", s.jsConfigs)
	}
	if _, ok := s.ruleCatalog.Lookup("community/check"); ok {
		t.Fatal("degraded plugin host committed an unroutable plugin rule")
	}
}

func TestHandleConfigRefreshActivationFailureAbortsAndKeepsLastGood(t *testing.T) {
	s, outgoing, root := newConfigRefreshTestServer(t)
	writeConfigCandidate(t, root)
	installLastGoodConfig(s, root)

	result := startConfigRefreshForTest(s, "config-change")
	loadMessage := nextConfigReverseRequest(t, outgoing, methodLoadConfigs)
	loadRequest, loadResponse := loadedConfigResponse(t, loadMessage, config.RslintConfig{{
		Rules: config.Rules{"no-debugger": "error"},
	}})
	respondToConfigReverseRequest(t, s, loadMessage, loadResponse, nil)

	activationMessage := nextConfigReverseRequest(t, outgoing, methodActivateConfigs)
	_, activationResponse := activationResponseForRequest(t, activationMessage, false)
	respondToConfigReverseRequest(t, s, activationMessage, activationResponse, nil)

	abortMessage := nextConfigReverseRequest(t, outgoing, methodAbortConfigs)
	abortControl := abortMessage.Params.(configTransactionControlRequest)
	if abortControl.TransactionID != loadRequest.TransactionID {
		t.Fatalf("abort transaction = %q, want %q", abortControl.TransactionID, loadRequest.TransactionID)
	}
	respondToConfigReverseRequest(t, s, abortMessage, abortResponseForRequest(t, abortMessage), nil)

	completed := awaitConfigRefreshResult(t, result)
	if completed.err == nil || !strings.Contains(completed.err.Error(), "could not prepare") {
		t.Fatalf("configRefresh error = %v, want plugin-host preparation failure", completed.err)
	}
	if !s.configDiscoveryActive {
		t.Fatal("valid configRefresh did not enable ancestor watcher retries")
	}
	if s.eslintPluginConfigGeneration != "last-good" || s.jsConfigs[root][0].Rules["no-console"] != "error" {
		t.Fatalf("failed activation replaced last-good state: generation=%q configs=%+v", s.eslintPluginConfigGeneration, s.jsConfigs)
	}
	assertLastGoodRuleCatalog(t, s)
}

func TestHandleConfigRefreshInvalidRuleOptionsAbortsAndKeepsLastGood(t *testing.T) {
	s, outgoing, root := newConfigRefreshTestServer(t)
	writeConfigCandidate(t, root)
	installLastGoodConfig(s, root)

	result := startConfigRefreshForTest(s, "config-change")
	loadMessage := nextConfigReverseRequest(t, outgoing, methodLoadConfigs)
	loadRequest, loadResponse := loadedConfigResponse(t, loadMessage, config.RslintConfig{{
		Rules: config.Rules{
			"no-console": []any{"error", map[string]any{"allow": "warn"}},
		},
	}})
	respondToConfigReverseRequest(t, s, loadMessage, loadResponse, nil)

	activationMessage := nextConfigReverseRequest(t, outgoing, methodActivateConfigs)
	_, activationResponse := activationResponseForRequest(t, activationMessage, true)
	respondToConfigReverseRequest(t, s, activationMessage, activationResponse, nil)

	abortMessage := nextConfigReverseRequest(t, outgoing, methodAbortConfigs)
	abortControl := abortMessage.Params.(configTransactionControlRequest)
	if abortControl.TransactionID != loadRequest.TransactionID {
		t.Fatalf("abort transaction = %q, want %q", abortControl.TransactionID, loadRequest.TransactionID)
	}
	respondToConfigReverseRequest(t, s, abortMessage, abortResponseForRequest(t, abortMessage), nil)

	completed := awaitConfigRefreshResult(t, result)
	if completed.err == nil || !strings.Contains(completed.err.Error(), `invalid options for rule "no-console"`) {
		t.Fatalf("configRefresh error = %v, want invalid rule options", completed.err)
	}
	if s.eslintPluginConfigGeneration != "last-good" || s.jsConfigs[root][0].Rules["no-console"] != "error" {
		t.Fatalf("invalid options replaced last-good state: generation=%q configs=%+v", s.eslintPluginConfigGeneration, s.jsConfigs)
	}
	assertLastGoodRuleCatalog(t, s)
}

func TestHandleConfigRefreshCommitFailureAbortsAndKeepsLastGood(t *testing.T) {
	s, outgoing, root := newConfigRefreshTestServer(t)
	writeConfigCandidate(t, root)
	installLastGoodConfig(s, root)

	result := startConfigRefreshForTest(s, "dependency-change")
	loadMessage := nextConfigReverseRequest(t, outgoing, methodLoadConfigs)
	_, loadResponse := loadedConfigResponse(t, loadMessage, config.RslintConfig{{
		Rules: config.Rules{"no-debugger": "error"},
	}})
	respondToConfigReverseRequest(t, s, loadMessage, loadResponse, nil)

	activationMessage := nextConfigReverseRequest(t, outgoing, methodActivateConfigs)
	_, activationResponse := activationResponseForRequest(t, activationMessage, true)
	respondToConfigReverseRequest(t, s, activationMessage, activationResponse, nil)

	commitMessage := nextConfigReverseRequest(t, outgoing, methodCommitConfigs)
	respondToConfigReverseRequest(t, s, commitMessage, commitResponseForRequest(t, commitMessage, false), nil)

	abortMessage := nextConfigReverseRequest(t, outgoing, methodAbortConfigs)
	respondToConfigReverseRequest(t, s, abortMessage, abortResponseForRequest(t, abortMessage), nil)

	completed := awaitConfigRefreshResult(t, result)
	if completed.err == nil || !strings.Contains(completed.err.Error(), "invalid commitConfigs response") {
		t.Fatalf("configRefresh error = %v, want commit failure", completed.err)
	}
	if s.eslintPluginConfigGeneration != "last-good" || s.jsConfigs[root][0].Rules["no-console"] != "error" {
		t.Fatalf("failed commit replaced last-good state: generation=%q configs=%+v", s.eslintPluginConfigGeneration, s.jsConfigs)
	}
	assertLastGoodRuleCatalog(t, s)
}

func TestHandleConfigRefreshInitialAllFailedCommitsUnavailableBoundaries(t *testing.T) {
	s, outgoing, root := newConfigRefreshTestServer(t)
	writeConfigCandidate(t, root)
	external := tspath.NormalizePath(t.TempDir())
	if err := os.WriteFile(tspath.ResolvePath(external, ".gitignore"), []byte("outside.ts\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	outside := documentURIFromPath(tspath.ResolvePath(external, "outside.ts"))
	s.documents[outside] = "debugger;"
	if err := os.WriteFile(
		filepath.Join(root, "rslint.json"),
		[]byte(`[{"rules":{"no-console":"error"}}]`),
		0o644,
	); err != nil {
		t.Fatal(err)
	}

	result := startConfigRefreshForTest(s, "initial")
	loadMessage := nextConfigReverseRequest(t, outgoing, methodLoadConfigs)
	loadRequest, loadResponse := failedConfigResponse(t, loadMessage, "synthetic import failure")
	respondToConfigReverseRequest(t, s, loadMessage, loadResponse, nil)

	activationMessage := nextConfigReverseRequest(t, outgoing, methodActivateConfigs)
	activationRequest, activationResponse := activationResponseForRequest(t, activationMessage, true)
	if activationRequest.EffectiveConfigIDs == nil || len(activationRequest.EffectiveConfigIDs) != 0 {
		t.Fatalf("unavailable recovery activation IDs = %#v, want []", activationRequest.EffectiveConfigIDs)
	}
	respondToConfigReverseRequest(t, s, activationMessage, activationResponse, nil)

	commitMessage := nextConfigReverseRequest(t, outgoing, methodCommitConfigs)
	respondToConfigReverseRequest(t, s, commitMessage, commitResponseForRequest(t, commitMessage, true), nil)

	completed := awaitConfigRefreshResult(t, result)
	if completed.err != nil {
		t.Fatalf("initial all-failed configRefresh stopped LSP startup: %v", completed.err)
	}
	if completed.response.TransactionID != loadRequest.TransactionID {
		t.Fatalf("unexpected recovery response: %+v", completed.response)
	}
	if s.configDiscoveryHasLastGood {
		t.Fatal("synthetic unavailable snapshot was marked as usable last-good")
	}
	inside := documentURIFromPath(filepath.Join(root, "src", "index.ts"))
	if !s.isUnavailableConfigForURI(inside) {
		t.Fatal("broken startup config did not suppress lint beneath its boundary")
	}
	entries, _, isJS := s.getConfigForURI(inside)
	if !isJS || hasPublicConfigContent(entries) {
		t.Fatalf("broken boundary config = %+v, isJS=%t; want empty JS boundary", entries, isJS)
	}
	if snapshot := s.documentLintSnapshot(outside); snapshot.unavailable || !snapshot.resolvedConfig.GloballyIgnored {
		t.Fatalf("all-failed recovery lost its frozen fallback Git policy: %+v", snapshot)
	}
	outsideConfig, _, outsideUsesCatalog := s.getConfigForURI(outside)
	if outsideUsesCatalog || hasPublicConfigContent(outsideConfig) {
		t.Fatalf("legacy JSON leaked into workspace fallback: config=%+v, usesCatalog=%t", outsideConfig, outsideUsesCatalog)
	}
}

func TestHandleConfigRefreshPartialCatalogCommitsUnavailableParentAndUsableChild(t *testing.T) {
	s, outgoing, root := newConfigRefreshTestServer(t)
	nested := tspath.NormalizePath(filepath.Join(root, "packages", "app"))
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	writeConfigCandidate(t, root)
	writeConfigCandidate(t, nested)
	s.documents[documentURIFromPath(tspath.ResolvePath(nested, "src/index.ts"))] = "debugger;"
	if err := os.WriteFile(
		filepath.Join(root, "rslint.json"),
		[]byte(`[{"rules":{"no-console":"error"}}]`),
		0o644,
	); err != nil {
		t.Fatal(err)
	}

	result := startConfigRefreshForTest(s, "initial")
	rootLoad := nextConfigReverseRequest(t, outgoing, methodLoadConfigs)
	_, rootResponse := failedConfigResponse(t, rootLoad, "broken root module")
	respondToConfigReverseRequest(t, s, rootLoad, rootResponse, nil)

	nestedLoad := nextConfigReverseRequest(t, outgoing, methodLoadConfigs)
	nestedRequest, nestedResponse := loadedConfigResponse(t, nestedLoad, config.RslintConfig{{
		Rules: config.Rules{"no-debugger": "error"},
	}})
	respondToConfigReverseRequest(t, s, nestedLoad, nestedResponse, nil)

	activationMessage := nextConfigReverseRequest(t, outgoing, methodActivateConfigs)
	activationRequest, activationResponse := activationResponseForRequest(t, activationMessage, true)
	if len(activationRequest.EffectiveConfigIDs) != 1 ||
		activationRequest.EffectiveConfigIDs[0] != nestedRequest.Candidates[0].ID {
		t.Fatalf("partial activation IDs = %v, want nested config only", activationRequest.EffectiveConfigIDs)
	}
	respondToConfigReverseRequest(t, s, activationMessage, activationResponse, nil)
	commitMessage := nextConfigReverseRequest(t, outgoing, methodCommitConfigs)
	respondToConfigReverseRequest(t, s, commitMessage, commitResponseForRequest(t, commitMessage, true), nil)

	completed := awaitConfigRefreshResult(t, result)
	if completed.err != nil {
		t.Fatalf("partial initial configRefresh failed: %v", completed.err)
	}
	if !s.configDiscoveryHasLastGood {
		t.Fatalf("partial recovery did not establish last-good: response %+v", completed.response)
	}
	rootEntries, exists := s.jsConfigs[root]
	if !exists || hasPublicConfigContent(rootEntries) {
		t.Fatalf("broken root boundary = %+v, want empty JS config", rootEntries)
	}
	if _, unavailable := s.jsUnavailableConfigs[root]; !unavailable {
		t.Fatal("broken root was not committed as an unavailable boundary")
	}
	outsideNested := documentURIFromPath(filepath.Join(root, "src", "index.ts"))
	if !s.isUnavailableConfigForURI(outsideNested) {
		t.Fatal("broken root did not suppress lint outside the usable child")
	}
	entries, _, isJS := s.getConfigForURI(outsideNested)
	if !isJS || hasPublicConfigContent(entries) {
		t.Fatalf("broken root fell through to workspace fallback: entries=%+v usesCatalog=%t", entries, isJS)
	}
	nestedFile := documentURIFromPath(filepath.Join(nested, "src", "index.ts"))
	if s.isUnavailableConfigForURI(nestedFile) {
		t.Fatal("usable child inherited its parent's unavailable state")
	}
	entries, configDir, isJS := s.getConfigForURI(nestedFile)
	value, found := configRuleValue(entries, "no-debugger")
	if !isJS || configDir != nested || !found || value != "error" {
		t.Fatalf("usable child config = %+v, dir=%q, isJS=%t", entries, configDir, isJS)
	}

	// The synthetic parent is a retryable tombstone, not last-good data. A
	// later failure at that same path must not reject a transaction that could
	// discover or update usable descendants. The real child remains protected.
	if failure, invalidates := s.failureAtCommittedConfigBoundary(s.fs, &discovery.ConfigCatalog{
		Failures: []discovery.ConfigFailure{{Directory: root, Message: "still broken"}},
	}); invalidates {
		t.Fatalf("unavailable tombstone invalidated refresh: %+v", failure)
	}
	if failure, invalidates := s.failureAtCommittedConfigBoundary(s.fs, &discovery.ConfigCatalog{
		Failures: []discovery.ConfigFailure{{Directory: nested, Message: "child broke"}},
	}); !invalidates || failure.Directory != nested {
		t.Fatalf("usable child failure = %+v, invalidates=%t", failure, invalidates)
	}
}

func TestUnavailableConfigBoundaryKeepsLexicalSymlinkFailure(t *testing.T) {
	parent := t.TempDir()
	realRoot := tspath.NormalizePath(filepath.Join(parent, "real"))
	aliasRoot := tspath.NormalizePath(filepath.Join(parent, "alias"))
	if err := os.MkdirAll(realRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(realRoot, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(realRoot, "src", "index.ts"), []byte("export {};\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(realRoot, aliasRoot); err != nil {
		t.Fatal(err)
	}
	fsy := bundled.WrapFS(osvfs.FS())
	catalog := &discovery.ConfigCatalog{
		Configs: map[string]config.RslintConfig{
			realRoot: {{}},
		},
		Failures: []discovery.ConfigFailure{{Directory: aliasRoot}},
	}
	resolver := target.NewOwnerIndex(catalog.Configs, fsy)
	if owner, _ := resolver.Resolve(config.FreezeFileIdentity(
		filepath.Join(aliasRoot, "src", "index.ts"),
		fsy,
	)); owner != realRoot {
		t.Fatalf("test precondition: canonical resolver owner = %q, want %q", owner, realRoot)
	}
	boundaries := unavailableConfigBoundaryDirectories(fsy, catalog)
	if len(boundaries) != 1 || boundaries[0] != aliasRoot {
		t.Fatalf("symlink failure boundaries = %v, want lexical alias %q", boundaries, aliasRoot)
	}
}

func TestFailureAtCommittedConfigBoundaryUsesLexicalIdentity(t *testing.T) {
	fsy := &configBoundaryIdentityFS{
		FS:            bundled.WrapFS(osvfs.FS()),
		caseSensitive: true,
		realPaths: map[string]string{
			"/repo/a": "/shared",
			"/repo/b": "/shared",
		},
	}
	s := newTestServer()
	s.fs = fsy
	s.jsConfigs = map[string]config.RslintConfig{
		"/repo/a": {{}},
	}
	s.jsUnavailableConfigs = make(map[string]struct{})

	if failure, invalidates := s.failureAtCommittedConfigBoundary(fsy, &discovery.ConfigCatalog{
		Failures: []discovery.ConfigFailure{{Directory: "/repo/b"}},
	}); invalidates {
		t.Fatalf("physical-only alias invalidated lexical last-good: %+v", failure)
	}
	if failure, invalidates := s.failureAtCommittedConfigBoundary(fsy, &discovery.ConfigCatalog{
		Failures: []discovery.ConfigFailure{{Directory: "/repo/a"}},
	}); !invalidates || failure.Directory != "/repo/a" {
		t.Fatalf("same lexical boundary = %+v, invalidates=%t", failure, invalidates)
	}

	fsy.caseSensitive = false
	s.jsConfigs = map[string]config.RslintConfig{
		"C:/Repo/App": {{}},
	}
	if failure, invalidates := s.failureAtCommittedConfigBoundary(fsy, &discovery.ConfigCatalog{
		Failures: []discovery.ConfigFailure{{Directory: "c:/repo/app"}},
	}); !invalidates || failure.Directory != "c:/repo/app" {
		t.Fatalf("native case alias = %+v, invalidates=%t", failure, invalidates)
	}
}

func TestHandleConfigRefreshAllFailedAbortsWhenUsableLastGoodExists(t *testing.T) {
	s, outgoing, root := newConfigRefreshTestServer(t)
	writeConfigCandidate(t, root)
	installLastGoodConfig(s, root)

	result := startConfigRefreshForTest(s, "config-change")
	loadMessage := nextConfigReverseRequest(t, outgoing, methodLoadConfigs)
	loadRequest, loadResponse := failedConfigResponse(t, loadMessage, "reload failure")
	respondToConfigReverseRequest(t, s, loadMessage, loadResponse, nil)

	abortMessage := nextConfigReverseRequest(t, outgoing, methodAbortConfigs)
	abortControl := abortMessage.Params.(configTransactionControlRequest)
	if abortControl.TransactionID != loadRequest.TransactionID {
		t.Fatalf("abort transaction = %q, want %q", abortControl.TransactionID, loadRequest.TransactionID)
	}
	respondToConfigReverseRequest(t, s, abortMessage, abortResponseForRequest(t, abortMessage), nil)

	completed := awaitConfigRefreshResult(t, result)
	if !errors.Is(completed.err, discovery.ErrAllConfigsFailed) {
		t.Fatalf("configRefresh error = %v, want ErrAllConfigsFailed", completed.err)
	}
	if s.eslintPluginConfigGeneration != "last-good" || !s.configDiscoveryHasLastGood ||
		s.jsConfigs[root][0].Rules["no-console"] != "error" {
		t.Fatalf("all-failed refresh replaced last-good state: generation=%q configs=%+v", s.eslintPluginConfigGeneration, s.jsConfigs)
	}
}

func TestHandleConfigRefreshPartialFailureAtCommittedBoundaryAborts(t *testing.T) {
	s, outgoing, root := newConfigRefreshTestServer(t)
	nested := tspath.NormalizePath(filepath.Join(root, "packages", "app"))
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	writeConfigCandidate(t, root)
	writeConfigCandidate(t, nested)
	installLastGoodConfig(s, root)
	s.jsConfigs[nested] = config.RslintConfig{{Rules: config.Rules{"old-nested": "error"}}}
	s.jsConfigOwnerIndex = target.NewOwnerIndex(s.jsConfigs, s.fs)

	result := startConfigRefreshForTest(s, "config-change")
	rootLoad := nextConfigReverseRequest(t, outgoing, methodLoadConfigs)
	_, rootResponse := loadedConfigResponse(t, rootLoad, config.RslintConfig{{
		Rules: config.Rules{"new-root": "error"},
	}})
	respondToConfigReverseRequest(t, s, rootLoad, rootResponse, nil)

	nestedLoad := nextConfigReverseRequest(t, outgoing, methodLoadConfigs)
	_, nestedResponse := failedConfigResponse(t, nestedLoad, "nested reload failure")
	respondToConfigReverseRequest(t, s, nestedLoad, nestedResponse, nil)

	activationMessage := nextConfigReverseRequest(t, outgoing, methodActivateConfigs)
	activationRequest, activationResponse := activationResponseForRequest(t, activationMessage, true)
	if len(activationRequest.EffectiveConfigIDs) != 1 {
		t.Fatalf("partial activation IDs = %v, want root only", activationRequest.EffectiveConfigIDs)
	}
	respondToConfigReverseRequest(t, s, activationMessage, activationResponse, nil)

	abortMessage := nextConfigReverseRequest(t, outgoing, methodAbortConfigs)
	respondToConfigReverseRequest(t, s, abortMessage, abortResponseForRequest(t, abortMessage), nil)

	completed := awaitConfigRefreshResult(t, result)
	if completed.err == nil || !strings.Contains(completed.err.Error(), "last-good boundary") {
		t.Fatalf("partial refresh error = %v, want last-good boundary failure", completed.err)
	}
	if s.eslintPluginConfigGeneration != "last-good" ||
		s.jsConfigs[root][0].Rules["no-console"] != "error" ||
		s.jsConfigs[nested][0].Rules["old-nested"] != "error" {
		t.Fatalf("partial boundary failure replaced last-good state: %+v", s.jsConfigs)
	}
}

func TestHandleConfigRefreshNewFailedBoundaryUsesParentFallback(t *testing.T) {
	s, outgoing, root := newConfigRefreshTestServer(t)
	nested := tspath.NormalizePath(filepath.Join(root, "packages", "app"))
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	writeConfigCandidate(t, root)
	writeConfigCandidate(t, nested)
	installLastGoodConfig(s, root)

	result := startConfigRefreshForTest(s, "config-change")
	rootLoad := nextConfigReverseRequest(t, outgoing, methodLoadConfigs)
	_, rootResponse := loadedConfigResponse(t, rootLoad, config.RslintConfig{{
		Rules: config.Rules{"new-root": "error"},
	}})
	respondToConfigReverseRequest(t, s, rootLoad, rootResponse, nil)
	nestedLoad := nextConfigReverseRequest(t, outgoing, methodLoadConfigs)
	_, nestedResponse := failedConfigResponse(t, nestedLoad, "new nested config is broken")
	respondToConfigReverseRequest(t, s, nestedLoad, nestedResponse, nil)
	activationMessage := nextConfigReverseRequest(t, outgoing, methodActivateConfigs)
	_, activationResponse := activationResponseForRequest(t, activationMessage, true)
	respondToConfigReverseRequest(t, s, activationMessage, activationResponse, nil)
	commitMessage := nextConfigReverseRequest(t, outgoing, methodCommitConfigs)
	respondToConfigReverseRequest(t, s, commitMessage, commitResponseForRequest(t, commitMessage, true), nil)

	completed := awaitConfigRefreshResult(t, result)
	if completed.err != nil {
		t.Fatalf("new failed child should use core parent fallback: %v", completed.err)
	}
	if len(s.jsConfigs) != 1 || s.jsConfigs[root][0].Rules["new-root"] != "error" {
		t.Fatalf("parent fallback catalog = %+v", s.jsConfigs)
	}
	owner, ok := s.nearestJSConfigKey(documentURIFromPath(filepath.Join(nested, "src", "index.ts")))
	if !ok || owner != root {
		t.Fatalf("new failed child owner = %q, ok=%t; want root", owner, ok)
	}
}

func TestHandleConfigRefreshUsesFreshFilesystemAndCommitsEmptyCatalog(t *testing.T) {
	s, outgoing, root := newConfigRefreshTestServer(t)
	writeConfigCandidate(t, root)

	first := startConfigRefreshForTest(s, "initial")
	firstLoad := nextConfigReverseRequest(t, outgoing, methodLoadConfigs)
	firstLoadRequest, firstLoadResponse := loadedConfigResponse(t, firstLoad, config.RslintConfig{{
		Plugins: []string{"generation-plugin"},
		Rules:   config.Rules{"generation-plugin/check": "error"},
	}})
	respondToConfigReverseRequest(t, s, firstLoad, firstLoadResponse, nil)
	firstActivation := nextConfigReverseRequest(t, outgoing, methodActivateConfigs)
	_, firstActivationResponse := activationResponseForRequest(t, firstActivation, true)
	firstActivationResponse.EslintPluginEntries = []config.EslintPluginEntry{{
		Prefix:    "generation-plugin",
		RuleNames: []string{"check"},
	}}
	respondToConfigReverseRequest(t, s, firstActivation, firstActivationResponse, nil)
	firstCommit := nextConfigReverseRequest(t, outgoing, methodCommitConfigs)
	respondToConfigReverseRequest(t, s, firstCommit, commitResponseForRequest(t, firstCommit, true), nil)
	if completed := awaitConfigRefreshResult(t, first); completed.err != nil {
		t.Fatalf("initial configRefresh failed: %v", completed.err)
	}
	if _, ok := s.ruleCatalog.Lookup("generation-plugin/check"); !ok {
		t.Fatal("initial generation did not commit its plugin rule")
	}

	if err := os.Remove(filepath.Join(root, "rslint.config.mjs")); err != nil {
		t.Fatal(err)
	}
	second := startConfigRefreshForTest(s, "config-change")
	// With no candidates, Go must still establish a real ConfigModuleHost
	// transaction using an explicit empty load frontier before activation.
	// A non-empty frontier here would prove stale generation cache reuse.
	secondLoad := nextConfigReverseRequest(t, outgoing, methodLoadConfigs)
	secondLoadRequest, ok := secondLoad.Params.(discovery.ConfigLoadBatchRequest)
	if !ok {
		t.Fatalf("empty loadConfigs params type = %T", secondLoad.Params)
	}
	if secondLoadRequest.LoadMode != discovery.ConfigModuleLoadFresh || secondLoadRequest.Candidates == nil || len(secondLoadRequest.Candidates) != 0 {
		t.Fatalf("empty catalog load request = %+v", secondLoadRequest)
	}
	respondToConfigReverseRequest(t, s, secondLoad, discovery.ConfigLoadBatchResponse{
		TransactionID: secondLoadRequest.TransactionID,
		Results:       []discovery.ConfigLoadResult{},
	}, nil)

	secondActivation := nextConfigReverseRequest(t, outgoing, methodActivateConfigs)
	secondActivationRequest, secondActivationResponse := activationResponseForRequest(t, secondActivation, true)
	if len(secondActivationRequest.EffectiveConfigIDs) != 0 {
		t.Fatalf("empty catalog activation IDs = %v", secondActivationRequest.EffectiveConfigIDs)
	}
	if secondLoadRequest.TransactionID != secondActivationRequest.TransactionID {
		t.Fatalf("empty load transaction = %q, activation transaction = %q", secondLoadRequest.TransactionID, secondActivationRequest.TransactionID)
	}
	if secondActivationRequest.TransactionID == firstLoadRequest.TransactionID {
		t.Fatal("successive config refreshes reused a transaction ID")
	}
	respondToConfigReverseRequest(t, s, secondActivation, secondActivationResponse, nil)
	secondCommit := nextConfigReverseRequest(t, outgoing, methodCommitConfigs)
	respondToConfigReverseRequest(t, s, secondCommit, commitResponseForRequest(t, secondCommit, true), nil)

	completed := awaitConfigRefreshResult(t, second)
	if completed.err != nil {
		t.Fatalf("empty configRefresh failed: %v", completed.err)
	}
	if len(s.jsConfigs) != 0 {
		t.Fatalf("deleted config remained committed: response=%+v configs=%+v", completed.response, s.jsConfigs)
	}
	if s.eslintPluginConfigGeneration != secondActivationRequest.TransactionID {
		t.Fatalf("empty catalog generation = %q, want %q", s.eslintPluginConfigGeneration, secondActivationRequest.TransactionID)
	}
	if _, ok := s.ruleCatalog.Lookup("generation-plugin/check"); ok {
		t.Fatal("deleted config's plugin rule leaked into the empty generation")
	}
	if s.configDiscoveryHasLastGood {
		t.Fatal("empty JavaScript catalog was marked as a usable JavaScript last-good generation")
	}

	// A newly created broken config now establishes an unavailable module-config
	// boundary instead of exposing the empty catalog's workspace fallback.
	writeConfigCandidate(t, root)
	third := startConfigRefreshForTest(s, "config-change")
	thirdLoad := nextConfigReverseRequest(t, outgoing, methodLoadConfigs)
	thirdLoadRequest, thirdLoadResponse := failedConfigResponse(t, thirdLoad, "new config is broken")
	respondToConfigReverseRequest(t, s, thirdLoad, thirdLoadResponse, nil)
	thirdActivation := nextConfigReverseRequest(t, outgoing, methodActivateConfigs)
	thirdActivationRequest, thirdActivationResponse := activationResponseForRequest(t, thirdActivation, true)
	if len(thirdActivationRequest.EffectiveConfigIDs) != 0 {
		t.Fatalf("unavailable activation IDs = %v, want none", thirdActivationRequest.EffectiveConfigIDs)
	}
	respondToConfigReverseRequest(t, s, thirdActivation, thirdActivationResponse, nil)
	thirdCommit := nextConfigReverseRequest(t, outgoing, methodCommitConfigs)
	respondToConfigReverseRequest(t, s, thirdCommit, commitResponseForRequest(t, thirdCommit, true), nil)
	completed = awaitConfigRefreshResult(t, third)
	if completed.err != nil {
		t.Fatalf("new broken config refresh failed instead of committing an unavailable boundary: %v", completed.err)
	}
	if completed.response.TransactionID != thirdLoadRequest.TransactionID {
		t.Fatalf("unavailable recovery response = %+v", completed.response)
	}
	if _, unavailable := s.jsUnavailableConfigs[root]; !unavailable || s.configDiscoveryHasLastGood {
		t.Fatalf("new broken config state: unavailable=%t hasLastGood=%t", unavailable, s.configDiscoveryHasLastGood)
	}
	entries, _, isJS := s.getConfigForURI(documentURIFromPath(filepath.Join(root, "src", "index.ts")))
	if !isJS || hasPublicConfigContent(entries) {
		t.Fatalf("new broken config fell through its empty JS boundary: entries=%+v isJS=%t", entries, isJS)
	}
}

func TestHandleConfigRefreshFailureKeepsCommittedIgnoreAfterDeletion(t *testing.T) {
	s, outgoing, root := newConfigRefreshTestServer(t)
	nested := tspath.NormalizePath(filepath.Join(root, "ignored"))
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	writeConfigCandidate(t, root)
	target := tspath.NormalizePath(filepath.Join(nested, "source.ts"))
	if err := os.WriteFile(target, []byte("debugger;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitignorePath := filepath.Join(root, ".gitignore")
	if err := os.WriteFile(gitignorePath, []byte("ignored/\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	s.documents[documentURIFromPath(target)] = "debugger;"

	// Commit a usable root config together with the current ignore snapshot.
	initial := startConfigRefreshForTest(s, "initial")
	initialLoad := nextConfigReverseRequest(t, outgoing, methodLoadConfigs)
	_, initialLoadResponse := loadedConfigResponse(t, initialLoad, config.RslintConfig{{
		Rules: config.Rules{"no-debugger": "error"},
	}})
	respondToConfigReverseRequest(t, s, initialLoad, initialLoadResponse, nil)
	initialActivation := nextConfigReverseRequest(t, outgoing, methodActivateConfigs)
	_, initialActivationResponse := activationResponseForRequest(t, initialActivation, true)
	respondToConfigReverseRequest(t, s, initialActivation, initialActivationResponse, nil)
	initialCommit := nextConfigReverseRequest(t, outgoing, methodCommitConfigs)
	respondToConfigReverseRequest(t, s, initialCommit, commitResponseForRequest(t, initialCommit, true), nil)
	if completed := awaitConfigRefreshResult(t, initial); completed.err != nil {
		t.Fatalf("initial configRefresh failed: %v", completed.err)
	}

	isIgnored := func() bool {
		return s.documentLintSnapshot(documentURIFromPath(target)).resolvedConfig.GloballyIgnored
	}
	if !isIgnored() {
		t.Fatal("target behind the committed .gitignore would produce diagnostics")
	}
	committedGeneration := s.eslintPluginConfigGeneration

	// Deleting the ignore while the replacement config fails must retain both
	// the prior catalog and its old ignore policy.
	if err := os.Remove(gitignorePath); err != nil {
		t.Fatal(err)
	}
	refresh := startConfigRefreshForTest(s, "gitignore-change")
	load := nextConfigReverseRequest(t, outgoing, methodLoadConfigs)
	_, failed := failedConfigResponse(t, load, "broken nested config")
	respondToConfigReverseRequest(t, s, load, failed, nil)
	abort := nextConfigReverseRequest(t, outgoing, methodAbortConfigs)
	respondToConfigReverseRequest(t, s, abort, abortResponseForRequest(t, abort), nil)
	completed := awaitConfigRefreshResult(t, refresh)
	if !errors.Is(completed.err, discovery.ErrAllConfigsFailed) {
		t.Fatalf("refresh error = %v, want ErrAllConfigsFailed", completed.err)
	}
	if s.eslintPluginConfigGeneration != committedGeneration {
		t.Fatalf("failed refresh changed generation from %q to %q", committedGeneration, s.eslintPluginConfigGeneration)
	}
	if !isIgnored() {
		t.Fatal("failed refresh leaked deleted .gitignore state; target would now produce diagnostics")
	}
}

func TestHandleConfigRefreshFailureKeepsCommittedIgnoreAfterCreation(t *testing.T) {
	s, outgoing, root := newConfigRefreshTestServer(t)
	writeConfigCandidate(t, root)
	target := tspath.NormalizePath(filepath.Join(root, "source.ts"))
	if err := os.WriteFile(target, []byte("debugger;\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	initial := startConfigRefreshForTest(s, "initial")
	initialLoad := nextConfigReverseRequest(t, outgoing, methodLoadConfigs)
	_, initialLoadResponse := loadedConfigResponse(t, initialLoad, config.RslintConfig{{
		Rules: config.Rules{"no-debugger": "error"},
	}})
	respondToConfigReverseRequest(t, s, initialLoad, initialLoadResponse, nil)
	initialActivation := nextConfigReverseRequest(t, outgoing, methodActivateConfigs)
	_, initialActivationResponse := activationResponseForRequest(t, initialActivation, true)
	respondToConfigReverseRequest(t, s, initialActivation, initialActivationResponse, nil)
	initialCommit := nextConfigReverseRequest(t, outgoing, methodCommitConfigs)
	respondToConfigReverseRequest(t, s, initialCommit, commitResponseForRequest(t, initialCommit, true), nil)
	if completed := awaitConfigRefreshResult(t, initial); completed.err != nil {
		t.Fatalf("initial configRefresh failed: %v", completed.err)
	}

	isIgnored := func() bool {
		return s.documentLintSnapshot(documentURIFromPath(target)).resolvedConfig.GloballyIgnored
	}
	if isIgnored() {
		t.Fatal("target was not diagnostic-visible in the committed generation")
	}
	committedGeneration := s.eslintPluginConfigGeneration

	if err := os.WriteFile(filepath.Join(root, ".gitignore"), []byte("source.ts\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	refresh := startConfigRefreshForTest(s, "gitignore-change")
	load := nextConfigReverseRequest(t, outgoing, methodLoadConfigs)
	_, loadResponse := loadedConfigResponse(t, load, config.RslintConfig{{
		Rules: config.Rules{"no-debugger": "error"},
	}})
	respondToConfigReverseRequest(t, s, load, loadResponse, nil)
	activation := nextConfigReverseRequest(t, outgoing, methodActivateConfigs)
	_, failedActivation := activationResponseForRequest(t, activation, false)
	respondToConfigReverseRequest(t, s, activation, failedActivation, nil)
	abort := nextConfigReverseRequest(t, outgoing, methodAbortConfigs)
	respondToConfigReverseRequest(t, s, abort, abortResponseForRequest(t, abort), nil)
	completed := awaitConfigRefreshResult(t, refresh)
	if completed.err == nil {
		t.Fatal("plugin activation failure unexpectedly committed")
	}
	if s.eslintPluginConfigGeneration != committedGeneration {
		t.Fatalf("failed refresh changed generation from %q to %q", committedGeneration, s.eslintPluginConfigGeneration)
	}
	if isIgnored() {
		t.Fatal("failed refresh leaked newly created .gitignore; prior diagnostics would disappear")
	}
}

func TestGitignoreWatcherRetriesRefreshAndPreservesLastGoodOnFailure(t *testing.T) {
	workspace := tspath.NormalizePath(t.TempDir())
	writeConfigCandidate(t, workspace)
	gitignorePath := filepath.Join(workspace, ".gitignore")
	if err := os.WriteFile(gitignorePath, []byte("unrelated/\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	s, outgoing := newTestServerWithQueue()
	s.cwd = workspace
	s.fs = bundled.WrapFS(osvfs.FS())
	s.pendingServerRequests = make(map[jsonrpc.ID]chan *lsproto.ResponseMessage)
	s.configDiscoveryActive = true
	installLastGoodConfig(s, workspace)
	documentURI := documentURIFromPath(filepath.Join(workspace, "src", "index.ts"))
	s.documents[documentURI] = "debugger;\n"
	s.diagnostics[documentURI] = nil
	s.docGeneration[documentURI] = 4

	watchResult := make(chan error, 1)
	go func() {
		watchResult <- s.handleDidChangeWatchedFiles(context.Background(), &lsproto.DidChangeWatchedFilesParams{
			Changes: []*lsproto.FileEvent{{
				Uri:  documentURIFromPath(gitignorePath),
				Type: lsproto.FileChangeTypeChanged,
			}},
		})
	}()

	loadMessage := nextConfigReverseRequest(t, outgoing, methodLoadConfigs)
	_, loadResponse := loadedConfigResponse(t, loadMessage, config.RslintConfig{{
		Rules: config.Rules{"no-debugger": "error"},
	}})
	respondToConfigReverseRequest(t, s, loadMessage, loadResponse, nil)
	activationMessage := nextConfigReverseRequest(t, outgoing, methodActivateConfigs)
	_, activationResponse := activationResponseForRequest(t, activationMessage, false)
	respondToConfigReverseRequest(t, s, activationMessage, activationResponse, nil)
	abortMessage := nextConfigReverseRequest(t, outgoing, methodAbortConfigs)
	respondToConfigReverseRequest(t, s, abortMessage, abortResponseForRequest(t, abortMessage), nil)

	select {
	case err := <-watchResult:
		if err != nil {
			t.Fatalf(".gitignore watcher returned error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal(".gitignore watcher did not finish")
	}
	if s.eslintPluginConfigGeneration != "last-good" || s.jsConfigs[workspace][0].Rules["no-console"] != "error" {
		t.Fatalf("failed watcher refresh replaced last-good state: generation=%q configs=%+v", s.eslintPluginConfigGeneration, s.jsConfigs)
	}
	if _, exists := s.diagnostics[documentURI]; exists {
		t.Fatal("failed watcher refresh retained diagnostics from the old ignore policy")
	}
	if s.docGeneration[documentURI] != 5 {
		t.Fatalf("document generation = %d, want 5", s.docGeneration[documentURI])
	}
	select {
	case <-s.refreshCh:
	default:
		t.Fatal("failed watcher refresh did not request diagnostics recomputation")
	}
}

func TestConfigRefreshProtocolRegistration(t *testing.T) {
	if !isBlockingMethod(methodConfigRefresh) {
		t.Fatal("rslint/configRefresh must run on the serialized dispatch loop")
	}
	if handlers()[methodConfigRefresh] == nil {
		t.Fatal("rslint/configRefresh handler is not registered")
	}
	if handlers()[lsproto.Method("rslint/configUpdate")] != nil {
		t.Fatal("deprecated rslint/configUpdate handler is still registered")
	}
	if handlers()[lsproto.Method("rslint/configCapabilities")] != nil {
		t.Fatal("deprecated rslint/configCapabilities handler is still registered")
	}
}

func TestHandleConfigRefreshRejectsInvalidProtocolWithoutMutation(t *testing.T) {
	s := newTestServer()
	s.eslintPluginConfigGeneration = "last-good"
	_, err := s.handleConfigRefresh(context.Background(), configRefreshRequest{
		ProtocolVersion: discovery.ConfigDiscoveryProtocolVersion + 1,
		Reason:          "initial",
	})
	if err == nil || !strings.Contains(err.Error(), "unsupported config refresh protocol") {
		t.Fatalf("configRefresh error = %v", err)
	}
	if s.eslintPluginConfigGeneration != "last-good" {
		t.Fatal("invalid protocol mutated committed generation")
	}
	if s.configDiscoveryActive {
		t.Fatal("invalid protocol enabled reverse watcher transactions")
	}
}

func TestDocumentGitPolicyCoverageRejectsRetargetedAliasesAndCodeActions(t *testing.T) {
	fsys := &configBoundaryIdentityFS{
		FS:            &mockFS{files: map[string]bool{}},
		caseSensitive: true,
		realPaths: map[string]string{
			"/alias":           "/owner-a",
			"/alias/source.ts": "/owner-a/source.ts",
		},
	}
	s := newTestServer()
	s.cwd, s.fs = "/alias", fsys
	s.gitPolicy = config.CollectGitPolicy(nil, s.cwd, fsys, nil, config.ResolveDirectoryIdentities([]string{s.cwd}, fsys))
	installJSConfigsForTest(s, map[string]config.RslintConfig{s.cwd: {{Rules: config.Rules{"no-debugger": "error"}}}})
	uri := lsproto.DocumentUri("file:///alias/source.ts")
	if snapshot := s.documentLintSnapshot(uri); snapshot.unavailable || !snapshot.configResolved {
		t.Fatalf("covered document was rejected: %+v", snapshot)
	}
	fsys.realPaths["/alias"] = "/owner-b"
	fsys.realPaths["/alias/source.ts"] = "/owner-b/source.ts"
	snapshot := s.documentLintSnapshot(uri)
	if !snapshot.unavailable || snapshot.configResolved {
		t.Fatalf("retargeted alias reused a policy from a different parent: %+v", snapshot)
	}
	// Stored results must not offer edits after the URI changes identity.
	s.diagnostics[uri] = []rule.RuleDiagnostic{{}}
	actions, err := s.handleCodeAction(context.Background(), &lsproto.CodeActionParams{TextDocument: lsproto.TextDocumentIdentifier{Uri: uri}})
	if err != nil || actions.CommandOrCodeActionArray == nil || len(*actions.CommandOrCodeActionArray) != 0 {
		t.Fatalf("uncovered document offered cached edits: %+v, %v", actions, err)
	}
	provider := &documentGenerationProvider{server: s, uri: uri, snapshot: snapshot}
	if _, release, err := provider.AcquireGeneration(context.Background(), linter.SourceSnapshot{}); err != nil {
		t.Fatalf("uncovered diagnostic generation did not stop: %v", err)
	} else if release != nil {
		release()
	}
	if _, release, err := acquireSpeculativeGeneration(context.Background(), "debugger;", snapshot, speculativeLintEnvironment{processCwd: s.cwd}); err != nil {
		t.Fatalf("uncovered fix generation did not stop: %v", err)
	} else if release != nil {
		release()
	}
}

func TestHandleDidOpenCollectsGitChainWithoutRetryingEveryEdit(t *testing.T) {
	s, outgoing, root := newConfigRefreshTestServer(t)
	writeConfigCandidate(t, root)
	entries := config.RslintConfig{{Rules: config.Rules{"no-debugger": "error"}}}
	initial := startConfigRefreshForTest(s, "initial")
	load := nextConfigReverseRequest(t, outgoing, methodLoadConfigs)
	_, response := loadedConfigResponse(t, load, entries)
	respondToConfigReverseRequest(t, s, load, response, nil)
	completeConfigWatchActivation(t, s, outgoing)
	if result := awaitConfigRefreshResult(t, initial); result.err != nil {
		t.Fatal(result.err)
	}
	previousPolicy, previousGeneration := s.gitPolicy, s.eslintPluginConfigGeneration

	external := tspath.NormalizePath(t.TempDir())
	for name, content := range map[string]string{
		".git":             "gitdir: elsewhere\n",
		".gitignore":       "ignored.ts\n",
		"rslint.config.js": "throw new Error('must not execute outside workspace');\n",
		"ignored.ts":       "debugger;\n",
	} {
		if err := os.WriteFile(tspath.ResolvePath(external, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	uri := documentURIFromPath(tspath.ResolvePath(external, "ignored.ts"))
	opened := make(chan error, 1)
	go func() {
		opened <- s.handleDidOpen(context.Background(), &lsproto.DidOpenTextDocumentParams{TextDocument: &lsproto.TextDocumentItem{Uri: uri, Text: "debugger;", LanguageId: "typescript", Version: 1}})
	}()
	load = nextConfigReverseRequest(t, outgoing, methodLoadConfigs)
	request, failed := failedConfigResponse(t, load, "transient workspace config failure")
	if len(request.Candidates) != 1 || request.Candidates[0].ConfigDirectory != root {
		t.Fatalf("open document became a config discovery seed: %+v", request.Candidates)
	}
	respondToConfigReverseRequest(t, s, load, failed, nil)
	abort := nextConfigReverseRequest(t, outgoing, methodAbortConfigs)
	respondToConfigReverseRequest(t, s, abort, abortResponseForRequest(t, abort), nil)
	if err := awaitConfigWatchEvent(t, opened); err != nil {
		t.Fatal(err)
	}
	if s.gitPolicy != previousPolicy || s.eslintPluginConfigGeneration != previousGeneration {
		t.Fatal("failed open refresh mixed configuration and Git generations")
	}
	if !s.documentLintSnapshot(uri).unavailable {
		t.Fatal("uncovered external document escaped last-good policy")
	}

	for version := int32(2); version <= 4; version++ {
		changed := make(chan error, 1)
		go func() {
			changed <- s.handleDidChange(context.Background(), makeDidChangeParams(uri, version, "debugger;\n"))
		}()
		select {
		case err := <-changed:
			if err != nil {
				t.Fatal(err)
			}
		case message := <-outgoing:
			load := message.AsRequest()
			t.Errorf("content-only edit %d reloaded workspace configuration", version)
			_, failed := failedConfigResponse(t, load, "workspace config still broken")
			respondToConfigReverseRequest(t, s, load, failed, nil)
			abort := nextConfigReverseRequest(t, outgoing, methodAbortConfigs)
			respondToConfigReverseRequest(t, s, abort, abortResponseForRequest(t, abort), nil)
			if err := awaitConfigWatchEvent(t, changed); err != nil {
				t.Fatal(err)
			}
		case <-time.After(5 * time.Second):
			t.Fatal("document edit did not complete")
		}
	}

	// Config changes still retry the failed transaction and recover coverage.
	retried := startConfigRefreshForTest(s, "config-change")
	load = nextConfigReverseRequest(t, outgoing, methodLoadConfigs)
	request, response = loadedConfigResponse(t, load, entries)
	if request.Candidates[0].ConfigDirectory != root {
		t.Fatal("retry discovered external config")
	}
	respondToConfigReverseRequest(t, s, load, response, nil)
	completeConfigWatchActivation(t, s, outgoing)
	if result := awaitConfigRefreshResult(t, retried); result.err != nil {
		t.Fatal(result.err)
	}
	snapshot := s.documentLintSnapshot(uri)
	if snapshot.unavailable || !snapshot.resolvedConfig.GloballyIgnored || snapshot.configKey != "" {
		t.Fatalf("external Git chain was not committed independently of config ownership: %+v", snapshot)
	}
	select {
	case message := <-outgoing:
		t.Fatalf("covered snapshot made reverse request: %+v", message)
	default:
	}
}

type gitObservationGuardFS struct {
	vfs.FS
	reject       bool
	observations int
}

func (fsys *gitObservationGuardFS) ReadFile(path tspath.RootedFilePath) (string, bool) {
	if tspath.GetBaseFileName(path.AsString()) == ".gitignore" {
		fsys.observations++
	}
	if fsys.reject && tspath.GetBaseFileName(path.AsString()) == ".gitignore" {
		panic("Git source read after discovery")
	}
	return fsys.FS.ReadFile(path)
}
func (fsys *gitObservationGuardFS) DirectoryExists(path tspath.RootedDirectoryPath) bool {
	if tspath.GetBaseFileName(path.AsString()) == ".git" {
		fsys.observations++
	}
	if fsys.reject && tspath.GetBaseFileName(path.AsString()) == ".git" {
		panic("Git marker read after discovery")
	}
	return fsys.FS.DirectoryExists(path)
}
func (fsys *gitObservationGuardFS) FileExists(path tspath.RootedFilePath) bool {
	if tspath.GetBaseFileName(path.AsString()) == ".git" {
		fsys.observations++
	}
	if fsys.reject && tspath.GetBaseFileName(path.AsString()) == ".git" {
		panic("Git marker read after discovery")
	}
	return fsys.FS.FileExists(path)
}

func TestHandleConfigRefreshDoesNotObserveGitAfterDiscovery(t *testing.T) {
	s, outgoing, root := newConfigRefreshTestServer(t)
	writeConfigCandidate(t, root)
	for name, content := range map[string]string{".gitignore": "source.ts\n", "source.ts": "debugger;\n"} {
		if err := os.WriteFile(tspath.ResolvePath(root, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	fsys := &gitObservationGuardFS{FS: s.fs}
	s.fs = fsys
	uri := documentURIFromPath(tspath.ResolvePath(root, "source.ts"))
	entries := config.RslintConfig{{Rules: config.Rules{"no-debugger": "error"}}}
	result := startConfigRefreshForTest(s, "initial")
	load := nextConfigReverseRequest(t, outgoing, methodLoadConfigs)
	_, response := loadedConfigResponse(t, load, entries)
	respondToConfigReverseRequest(t, s, load, response, nil)
	activation := nextConfigReverseRequest(t, outgoing, methodActivateConfigs)
	// These changes belong to the next generation. Preparing resolvers and
	// committing the current generation must not consult either file again.
	if err := os.WriteFile(tspath.ResolvePath(root, ".gitignore"), []byte("# now visible\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(tspath.ResolvePath(root, ".git"), []byte("gitdir: elsewhere\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	fsys.reject = true
	_, active := activationResponseForRequest(t, activation, true)
	respondToConfigReverseRequest(t, s, activation, active, nil)
	commit := nextConfigReverseRequest(t, outgoing, methodCommitConfigs)
	respondToConfigReverseRequest(t, s, commit, commitResponseForRequest(t, commit, true), nil)
	if completed := awaitConfigRefreshResult(t, result); completed.err != nil {
		t.Fatal(completed.err)
	}
	if snapshot := s.documentLintSnapshot(uri); snapshot.unavailable || !snapshot.resolvedConfig.GloballyIgnored {
		t.Fatalf("committed policy observed later filesystem state: %+v", snapshot)
	}
	fsys.reject = false
	result = startConfigRefreshForTest(s, "gitignore-change")
	load = nextConfigReverseRequest(t, outgoing, methodLoadConfigs)
	_, response = loadedConfigResponse(t, load, entries)
	respondToConfigReverseRequest(t, s, load, response, nil)
	completeConfigWatchActivation(t, s, outgoing)
	if completed := awaitConfigRefreshResult(t, result); completed.err != nil {
		t.Fatal(completed.err)
	}
	if s.documentLintSnapshot(uri).resolvedConfig.GloballyIgnored {
		t.Fatal("next generation did not observe changed Git sources")
	}
}

func TestGitMarkerWatcherCreatesAndRemovesBoundary(t *testing.T) {
	outer := tspath.NormalizePath(t.TempDir())
	root := tspath.ResolvePath(outer, "workspace")
	if err := os.MkdirAll(tspath.ResolvePath(outer, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	writeConfigCandidate(t, root)
	if err := os.WriteFile(tspath.ResolvePath(outer, ".gitignore"), []byte("source.ts\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(tspath.ResolvePath(root, "source.ts"), []byte("debugger;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	s, outgoing := newAncestorConfigWatchTestServer(root)
	entries := config.RslintConfig{{Rules: config.Rules{"no-debugger": "error"}}}
	initial := startConfigRefreshForTest(s, "initial")
	load := nextConfigReverseRequest(t, outgoing, methodLoadConfigs)
	_, response := loadedConfigResponse(t, load, entries)
	respondToConfigReverseRequest(t, s, load, response, nil)
	completeConfigWatchActivation(t, s, outgoing)
	if completed := awaitConfigRefreshResult(t, initial); completed.err != nil {
		t.Fatal(completed.err)
	}
	uri := documentURIFromPath(tspath.ResolvePath(root, "source.ts"))
	if !s.documentLintSnapshot(uri).resolvedConfig.GloballyIgnored {
		t.Fatal("initial ancestor source did not apply")
	}
	marker := tspath.ResolvePath(root, ".git")
	for _, create := range []bool{true, false} {
		event := lsproto.FileChangeTypeDeleted
		if create {
			if err := os.WriteFile(marker, []byte("gitdir: elsewhere\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			event = lsproto.FileChangeTypeCreated
		} else if err := os.Remove(marker); err != nil {
			t.Fatal(err)
		}
		changed := startConfigWatchEvent(s, marker, event)
		load = nextConfigReverseRequest(t, outgoing, methodLoadConfigs)
		_, response = loadedConfigResponse(t, load, entries)
		respondToConfigReverseRequest(t, s, load, response, nil)
		completeConfigWatchActivation(t, s, outgoing)
		if err := awaitConfigWatchEvent(t, changed); err != nil {
			t.Fatal(err)
		}
		if ignored := s.documentLintSnapshot(uri).resolvedConfig.GloballyIgnored; ignored == create {
			t.Fatalf("marker exists=%v ignored=%v", create, ignored)
		}
	}
}

func TestConfigRefreshRegistersPhysicalGitWatchesBeforeReadingSources(t *testing.T) {
	s, outgoing, root := newConfigRefreshTestServer(t)
	writeConfigCandidate(t, root)
	fsys := &gitObservationGuardFS{FS: s.fs}
	s.fs, s.watchEnabled = fsys, true
	initial := startConfigRefreshForTest(s, "initial")
	registration := nextConfigReverseRequest(t, outgoing, lsproto.MethodClientRegisterCapability)
	if fsys.observations != 0 {
		t.Fatal("Git sources were observed before watcher registration completed")
	}
	respondToConfigReverseRequest(t, s, registration, nil, nil)
	load := nextConfigReverseRequest(t, outgoing, methodLoadConfigs)
	_, response := loadedConfigResponse(t, load, config.RslintConfig{{}})
	respondToConfigReverseRequest(t, s, load, response, nil)
	completeConfigWatchActivation(t, s, outgoing)
	if result := awaitConfigRefreshResult(t, initial); result.err != nil {
		t.Fatal(result.err)
	}
	previousID := s.gitPolicyWatcherID

	external := tspath.NormalizePath(t.TempDir())
	physicalExternal := tspath.NormalizePath(fsys.Realpath(pathutil.Path(external)).AsString())
	uri := documentURIFromPath(tspath.ResolvePath(external, "source.ts"))
	s.documents[uri] = "debugger;"
	previousObservations := fsys.observations
	refresh := startConfigRefreshForTest(s, "gitignore-change")
	registration = nextConfigReverseRequest(t, outgoing, lsproto.MethodClientRegisterCapability)
	params, ok := registration.Params.(*lsproto.RegistrationParams)
	if !ok || len(params.Registrations) != 1 {
		t.Fatalf("invalid registration: %+v", registration.Params)
	}
	watchers := params.Registrations[0].RegisterOptions.WorkspaceDidChangeWatchedFiles.Watchers
	foundMarker, foundIgnore := false, false
	for _, watcher := range watchers {
		if watcher.GlobPattern.Pattern == nil {
			t.Fatal("unexpected relative watcher")
		}
		switch *watcher.GlobPattern.Pattern {
		case tspath.ResolvePath(physicalExternal, ".git"):
			foundMarker = true
		case tspath.ResolvePath(physicalExternal, ".gitignore"):
			foundIgnore = true
		}
	}
	if !foundMarker || !foundIgnore {
		t.Fatal("external source and missing marker were not watched")
	}
	if fsys.observations != previousObservations {
		t.Fatal("new Git sources were read while watcher registration was pending")
	}
	respondToConfigReverseRequest(t, s, registration, nil, nil)
	for attempt := range 3 {
		if attempt > 0 {
			refresh = startConfigRefreshForTest(s, "gitignore-change")
		}
		unregistration := nextConfigReverseRequest(t, outgoing, lsproto.MethodClientUnregisterCapability)
		respondToConfigReverseRequest(t, s, unregistration, nil, errors.New("temporary unregistration failure"))
		completeSuccessfulConfigRefreshForTest(t, s, outgoing, config.RslintConfig{{}})
		if result := awaitConfigRefreshResult(t, refresh); result.err != nil {
			t.Fatalf("old watcher cleanup blocked refresh %d: %v", attempt, result.err)
		}
		if s.gitPolicyWatcherID == previousID || !s.watchers.Has(previousID) || !s.watchers.Has(s.gitPolicyWatcherID) {
			t.Fatal("failed cleanup lost an active or pending registration")
		}
		if !s.gitPolicy.CoversDirectory(physicalExternal) {
			t.Fatal("failed cleanup prevented new Git source coverage")
		}
	}
	refresh = startConfigRefreshForTest(s, "gitignore-change")
	unregistration := nextConfigReverseRequest(t, outgoing, lsproto.MethodClientUnregisterCapability)
	respondToConfigReverseRequest(t, s, unregistration, nil, nil)
	completeSuccessfulConfigRefreshForTest(t, s, outgoing, config.RslintConfig{{}})
	if result := awaitConfigRefreshResult(t, refresh); result.err != nil {
		t.Fatal(result.err)
	}
	if s.gitPolicyWatcherID == previousID || s.watchers.Has(previousID) {
		t.Fatal("watch registration was not replaced")
	}
}

func TestHandleDidOpenCollectsIgnoredDirectoryWithoutLoadingItsConfig(t *testing.T) {
	s, outgoing, root := newConfigRefreshTestServer(t)
	writeConfigCandidate(t, root)
	deep := tspath.ResolvePath(root, "ignored/deep")
	if err := os.MkdirAll(deep, 0o755); err != nil {
		t.Fatal(err)
	}
	for path, content := range map[string]string{
		tspath.ResolvePath(root, ".gitignore"):       "ignored/\n",
		tspath.ResolvePath(deep, "rslint.config.js"): "throw new Error('ignored config must not execute');\n",
		tspath.ResolvePath(deep, ".gitignore"):       "!source.ts\n",
		tspath.ResolvePath(deep, "source.ts"):        "debugger;\n",
	} {
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	entries := config.RslintConfig{{Rules: config.Rules{"no-debugger": "error"}}}
	initial := startConfigRefreshForTest(s, "initial")
	load := nextConfigReverseRequest(t, outgoing, methodLoadConfigs)
	_, response := loadedConfigResponse(t, load, entries)
	respondToConfigReverseRequest(t, s, load, response, nil)
	completeConfigWatchActivation(t, s, outgoing)
	if result := awaitConfigRefreshResult(t, initial); result.err != nil {
		t.Fatal(result.err)
	}
	uri := documentURIFromPath(tspath.ResolvePath(deep, "source.ts"))
	if !s.documentLintSnapshot(uri).unavailable {
		t.Fatal("pruned directory was treated as already covered")
	}
	opened := make(chan error, 1)
	go func() {
		opened <- s.handleDidOpen(context.Background(), &lsproto.DidOpenTextDocumentParams{TextDocument: &lsproto.TextDocumentItem{Uri: uri, Text: "debugger;", Version: 1, LanguageId: "typescript"}})
	}()
	load = nextConfigReverseRequest(t, outgoing, methodLoadConfigs)
	request, response := loadedConfigResponse(t, load, entries)
	if request.Candidates[0].ConfigDirectory != root {
		t.Fatal("ignored document enabled a new config seed")
	}
	respondToConfigReverseRequest(t, s, load, response, nil)
	completeConfigWatchActivation(t, s, outgoing)
	if err := awaitConfigWatchEvent(t, opened); err != nil {
		t.Fatal(err)
	}
	if snapshot := s.documentLintSnapshot(uri); snapshot.unavailable || !snapshot.resolvedConfig.GloballyIgnored {
		t.Fatalf("ignored document did not acquire its complete Git chain: %+v", snapshot)
	}
	if err := s.handleDidOpen(context.Background(), &lsproto.DidOpenTextDocumentParams{TextDocument: &lsproto.TextDocumentItem{Uri: uri, Text: "debugger;", Version: 2, LanguageId: "typescript"}}); err != nil {
		t.Fatal(err)
	}
	select {
	case message := <-outgoing:
		t.Fatalf("covered didOpen refreshed again: %+v", message)
	default:
	}
}

func TestHandleDidOpenVirtualDirectoryAliasInheritsPhysicalGitPolicy(t *testing.T) {
	s, outgoing, root := newConfigRefreshTestServer(t)
	writeConfigCandidate(t, root)
	physical := tspath.ResolvePath(root, "packages/app")
	if err := os.MkdirAll(physical, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(tspath.ResolvePath(root, ".gitignore"), []byte("generated/\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	alias := tspath.ResolvePath(tspath.NormalizePath(t.TempDir()), "alias")
	if err := os.Symlink(physical, alias); err != nil {
		t.Skipf("directory symlink unavailable: %v", err)
	}
	entries := config.RslintConfig{{Rules: config.Rules{"no-debugger": "error"}}}
	initial := startConfigRefreshForTest(s, "initial")
	load := nextConfigReverseRequest(t, outgoing, methodLoadConfigs)
	_, response := loadedConfigResponse(t, load, entries)
	respondToConfigReverseRequest(t, s, load, response, nil)
	completeConfigWatchActivation(t, s, outgoing)
	if result := awaitConfigRefreshResult(t, initial); result.err != nil {
		t.Fatal(result.err)
	}
	file := tspath.ResolvePath(alias, "generated/deep/new.ts")
	uri := documentURIFromPath(file)
	if !s.documentLintSnapshot(uri).unavailable {
		t.Fatal("unobserved virtual parent was treated as covered")
	}
	opened := make(chan error, 1)
	go func() {
		opened <- s.handleDidOpen(context.Background(), &lsproto.DidOpenTextDocumentParams{TextDocument: &lsproto.TextDocumentItem{Uri: uri, Text: "debugger;", Version: 1, LanguageId: "typescript"}})
	}()
	load = nextConfigReverseRequest(t, outgoing, methodLoadConfigs)
	request, response := loadedConfigResponse(t, load, entries)
	if len(request.Candidates) != 1 || request.Candidates[0].ConfigDirectory != root {
		t.Fatalf("virtual document became a config discovery seed: %+v", request.Candidates)
	}
	respondToConfigReverseRequest(t, s, load, response, nil)
	completeConfigWatchActivation(t, s, outgoing)
	if err := awaitConfigWatchEvent(t, opened); err != nil {
		t.Fatal(err)
	}
	snapshot := s.documentLintSnapshot(uri)
	wantParent := tspath.ResolvePath(tspath.NormalizePath(osvfs.FS().Realpath(pathutil.Path(physical)).AsString()), "generated/deep")
	if snapshot.unavailable || !snapshot.configResolved || !snapshot.resolvedConfig.GloballyIgnored || snapshot.target.Path != uriToPath(uri) || snapshot.target.CanonicalParentPath != wantParent {
		t.Fatalf("virtual alias lost its physical Git chain: %+v", snapshot)
	}
	if _, err := os.Stat(tspath.GetDirectoryPath(file)); !os.IsNotExist(err) {
		t.Fatalf("virtual parent must remain absent: %v", err)
	}
}

func TestDocumentSnapshotDoesNotBootstrapAfterHandshakeWithoutCommit(t *testing.T) {
	s := newTestServer()
	fsys := &gitObservationGuardFS{FS: &mockFS{files: map[string]bool{}}, reject: true}
	s.cwd, s.fs, s.configDiscoveryActive = "/project", fsys, true
	uri := lsproto.DocumentUri("file:///project/source.ts")
	if snapshot := s.documentLintSnapshot(uri); !snapshot.unavailable || snapshot.configResolved || snapshot.gitPolicy != nil {
		t.Fatalf("uncommitted handshake resumed bootstrap policy: %+v", snapshot)
	}
	fsys.reject = false
	s.configDiscoveryActive = false
	if snapshot := s.documentLintSnapshot(uri); snapshot.unavailable || !snapshot.configResolved || snapshot.gitPolicy == nil {
		t.Fatalf("pre-handshake bootstrap could not collect the exact document: %+v", snapshot)
	}
}

func TestDocumentGitPolicyCoveredIgnoredAliasSuppressesCachedEdits(t *testing.T) {
	s, _, root := newConfigRefreshTestServer(t)
	for name, content := range map[string]string{".gitignore": "b/source.ts\n", "a/source.ts": "debugger;", "b/source.ts": "debugger;"} {
		path := tspath.ResolvePath(root, name)
		if err := os.MkdirAll(tspath.GetDirectoryPath(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	alias := tspath.ResolvePath(root, "alias")
	if err := os.Symlink(tspath.ResolvePath(root, "a"), alias); err != nil {
		t.Skip(err)
	}
	s.gitPolicy = config.CollectGitPolicy(nil, root, s.fs, nil, config.ResolveDirectoryIdentities([]string{root}, s.fs))
	installJSConfigsForTest(s, map[string]config.RslintConfig{root: {{Rules: config.Rules{"no-debugger": "error"}}}})
	uri := documentURIFromPath(tspath.ResolvePath(alias, "source.ts"))
	s.documents[uri] = "debugger;"
	if snap := s.documentLintSnapshot(uri); snap.unavailable || snap.resolvedConfig.GloballyIgnored {
		t.Fatalf("bad original snapshot: %+v", snap)
	}
	s.diagnostics[uri] = []rule.RuleDiagnostic{pluginDiag("debugger;", "no-debugger", "old target diagnostic", 0, 8)}
	if err := os.Remove(alias); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(tspath.ResolvePath(root, "b"), alias); err != nil {
		t.Fatal(err)
	}
	snap := s.documentLintSnapshot(uri)
	if snap.unavailable || !snap.resolvedConfig.GloballyIgnored {
		t.Fatalf("test requires covered ignored destination: %+v", snap)
	}
	actions, err := s.handleCodeAction(context.Background(), &lsproto.CodeActionParams{TextDocument: lsproto.TextDocumentIdentifier{Uri: uri}, Range: lsproto.Range{Start: lsproto.Position{}, End: lsproto.Position{Character: 8}}})
	if err != nil {
		t.Fatal(err)
	}
	if actions.CommandOrCodeActionArray != nil && len(*actions.CommandOrCodeActionArray) != 0 {
		t.Fatalf("ignored retargeted alias offered %d stale code actions", len(*actions.CommandOrCodeActionArray))
	}
}

func TestGitWatchRegistrationKeyIncludesRetargetedRecursiveRoots(t *testing.T) {
	s, outgoing, _ := newConfigRefreshTestServer(t)
	s.watchEnabled = true
	rootsA := []config.DirectoryIdentity{{LexicalPath: "/alias", CanonicalPath: "/a"}}
	files := []config.PathIdentity{{Path: "/b/source.ts", CanonicalParentPath: "/b"}}
	watchedA := gitPolicyWatchDirectories(rootsA, files)
	result := make(chan error, 1)
	go func() { result <- s.updateGitPolicyWatchers(context.Background(), rootsA, watchedA) }()
	registration := nextConfigReverseRequest(t, outgoing, lsproto.MethodClientRegisterCapability)
	respondToConfigReverseRequest(t, s, registration, nil, nil)
	if err := awaitConfigWatchEvent(t, result); err != nil {
		t.Fatal(err)
	}
	oldID := s.gitPolicyWatcherID
	s.committedGitWatchDirectories = watchedA
	rootsB := []config.DirectoryIdentity{{LexicalPath: "/alias", CanonicalPath: "/b"}}
	go func() {
		result <- s.updateGitPolicyWatchers(context.Background(), rootsB, gitPolicyWatchDirectories(rootsB, files))
	}()
	registration = nextConfigReverseRequest(t, outgoing, lsproto.MethodClientRegisterCapability)
	params, ok := registration.Params.(*lsproto.RegistrationParams)
	if !ok || len(params.Registrations) != 1 {
		t.Fatalf("invalid registration: %+v", registration.Params)
	}
	var sourceWatch, markerWatch bool
	for _, watcher := range params.Registrations[0].RegisterOptions.WorkspaceDidChangeWatchedFiles.Watchers {
		if watcher.GlobPattern.Pattern == nil {
			t.Fatal("unexpected relative watcher")
		}
		switch *watcher.GlobPattern.Pattern {
		case "/b/**/.gitignore":
			sourceWatch = true
		case "/b/**/.git":
			markerWatch = true
		}
	}
	if !sourceWatch || !markerWatch {
		t.Fatal("retargeting lost recursive watches for the new physical root")
	}
	respondToConfigReverseRequest(t, s, registration, nil, nil)
	unregistration := nextConfigReverseRequest(t, outgoing, lsproto.MethodClientUnregisterCapability)
	respondToConfigReverseRequest(t, s, unregistration, nil, nil)
	if err := awaitConfigWatchEvent(t, result); err != nil {
		t.Fatal(err)
	}
	if s.gitPolicyWatcherID == oldID || s.watchers.Has(oldID) {
		t.Fatal("unchanged exact-directory union suppressed recursive watch replacement")
	}
}

func TestCanceledGitWatcherRegistrationCleansAttemptOnNextRefresh(t *testing.T) {
	s, outgoing, root := newConfigRefreshTestServer(t)
	writeConfigCandidate(t, root)
	fsys := &gitObservationGuardFS{FS: s.fs}
	s.fs, s.watchEnabled = fsys, true
	initial := startConfigRefreshForTest(s, "initial")
	registration := nextConfigReverseRequest(t, outgoing, lsproto.MethodClientRegisterCapability)
	respondToConfigReverseRequest(t, s, registration, nil, nil)
	completeSuccessfulConfigRefreshForTest(t, s, outgoing, config.RslintConfig{{}})
	if completed := awaitConfigRefreshResult(t, initial); completed.err != nil {
		t.Fatal(completed.err)
	}
	activeID, activePolicy := s.gitPolicyWatcherID, s.gitPolicy
	external := tspath.NormalizePath(t.TempDir())
	s.documents[documentURIFromPath(tspath.ResolvePath(external, "source.ts"))] = "debugger;"
	observations := fsys.observations
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	result := make(chan configRefreshTestResult, 1)
	go func() {
		response, err := s.refreshConfig(ctx)
		result <- configRefreshTestResult{response: response, err: err}
	}()
	registration = nextConfigReverseRequest(t, outgoing, lsproto.MethodClientRegisterCapability)
	params := registration.Params.(*lsproto.RegistrationParams)
	attemptedID := params.Registrations[0].Id
	// The request reached the client, which may install it before Go observes
	// cancellation. Its missing acknowledgement cannot prove it was rejected.
	cancel()
	if completed := awaitConfigRefreshResult(t, result); !errors.Is(completed.err, context.Canceled) {
		t.Fatalf("registration cancellation = %v", completed.err)
	}
	select {
	case message := <-outgoing:
		if message.AsRequest().Method != lsproto.MethodCancelRequest {
			t.Fatalf("unexpected cancellation message: %+v", message)
		}
	default:
		t.Fatal("registration cancellation was not forwarded")
	}
	uncertainID := project.WatcherID(attemptedID)
	if !s.watchers.Has(uncertainID) {
		t.Fatal("canceled registration attempt was forgotten")
	}
	if s.gitPolicyWatcherID != activeID || s.gitPolicy != activePolicy || !s.watchers.Has(activeID) {
		t.Fatal("cancellation replaced the active watcher or config generation")
	}
	if fsys.observations != observations {
		t.Fatal("canceled registration read new Git sources")
	}

	retry := startConfigRefreshForTest(s, "gitignore-change")
	registration = nextConfigReverseRequest(t, outgoing, lsproto.MethodClientRegisterCapability)
	if fsys.observations != observations {
		t.Fatal("retry read Git sources before the new registration completed")
	}
	respondToConfigReverseRequest(t, s, registration, nil, nil)
	for range 2 {
		unregistration := nextConfigReverseRequest(t, outgoing, lsproto.MethodClientUnregisterCapability)
		respondToConfigReverseRequest(t, s, unregistration, nil, nil)
	}
	completeSuccessfulConfigRefreshForTest(t, s, outgoing, config.RslintConfig{{}})
	if completed := awaitConfigRefreshResult(t, retry); completed.err != nil {
		t.Fatal(completed.err)
	}
	if s.watchers.Has(activeID) || s.watchers.Has(uncertainID) || !s.watchers.Has(s.gitPolicyWatcherID) {
		t.Fatal("retry leaked a superseded or uncertain registration")
	}
}
