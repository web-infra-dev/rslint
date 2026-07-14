package lsp

import (
	"context"
	stdjson "encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/jsonrpc"
	"github.com/microsoft/typescript-go/shim/lsp/lsproto"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"

	"github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/config/discovery"
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

func (fs *configBoundaryIdentityFS) UseCaseSensitiveFileNames() bool {
	return fs.caseSensitive
}

func (fs *configBoundaryIdentityFS) Realpath(filePath string) string {
	filePath = tspath.NormalizePath(filePath)
	if realPath := fs.realPaths[filePath]; realPath != "" {
		return realPath
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
	result := make(chan configRefreshTestResult, 1)
	go func() {
		response, err := s.handleConfigRefresh(context.Background(), map[string]any{
			"protocolVersion": discovery.ConfigDiscoveryProtocolVersion,
			"reason":          reason,
		})
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
			ID:                loadRequest.Candidates[0].ID,
			Status:            "loaded",
			Entries:           entries,
			SourceFingerprint: "test-fingerprint",
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
		Generation:      activation.TransactionID,
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
		Generation:    control.TransactionID,
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
		Generation:    control.TransactionID,
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
	entries := config.RslintConfig{{Rules: config.Rules{"no-console": "error"}}}
	s.jsConfigs = map[string]config.RslintConfig{root: entries}
	s.jsConfigOwnerResolver = config.NewConfigOwnerResolver(s.jsConfigs, s.fs)
	s.jsConfigKeyByPath = map[string]string{root: root}
	s.jsUnavailableConfigs = make(map[string]struct{})
	s.tsConfigPathsByConfig = map[string][]string{root: nil}
	s.eslintPluginConfigGeneration = "last-good"
	s.configDiscoveryV2HasLastGood = true
}

func TestHandleConfigRefreshCommitsFilesystemPathCatalog(t *testing.T) {
	config.RegisterAllRules()
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
	respondToConfigReverseRequest(t, s, activationMessage, activationResponse, nil)

	commitMessage := nextConfigReverseRequest(t, outgoing, methodCommitConfigs)
	commitResponse := commitResponseForRequest(t, commitMessage, true)
	respondToConfigReverseRequest(t, s, commitMessage, commitResponse, nil)

	completed := awaitConfigRefreshResult(t, result)
	if completed.err != nil {
		t.Fatalf("configRefresh failed: %v", completed.err)
	}
	if completed.response.Generation != loadRequest.TransactionID || completed.response.ConfigCount != 1 {
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
	if got := s.pluginConfigKeyForURI(fileURI); got != root {
		t.Fatalf("v2 plugin configKey = %q, want exact catalog path %q", got, root)
	}
}

func TestPrepareDiscoveredConfigSnapshotUsesChildGitignoreSourceBoundaries(t *testing.T) {
	root := tspath.NormalizePath(t.TempDir())
	child := tspath.NormalizePath(filepath.Join(root, "packages", "app"))
	rootTarget := tspath.NormalizePath(filepath.Join(root, "root.ts"))
	childTarget := tspath.NormalizePath(filepath.Join(child, "source.ts"))
	if err := os.MkdirAll(child, 0o755); err != nil {
		t.Fatal(err)
	}
	for filePath := range map[string]struct{}{rootTarget: {}, childTarget: {}} {
		if err := os.WriteFile(filePath, []byte("debugger;\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, ".gitignore"), []byte("root.ts\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(child, ".gitignore"), []byte("source.ts\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	fsys := bundled.WrapFS(osvfs.FS())
	catalog := &discovery.ConfigCatalog{
		TransactionID: "snapshot-boundaries",
		Configs: map[string]config.RslintConfig{
			root:  {{Rules: config.Rules{"no-console": "error"}}},
			child: {{Rules: config.Rules{"no-debugger": "error"}}},
		},
	}
	s := newTestServer()
	s.cwd = root
	snapshot, err := s.prepareDiscoveredConfigSnapshot(fsys, catalog)
	if err != nil {
		t.Fatal(err)
	}
	if !snapshot.configs[root].IsFileIgnored(rootTarget, root) {
		t.Fatal("root config did not collect its own .gitignore")
	}
	if snapshot.configs[root].IsFileIgnored(childTarget, root) {
		t.Fatal("root config crossed the child config's .gitignore source boundary")
	}
	if !snapshot.configs[child].IsFileIgnored(childTarget, child) {
		t.Fatal("child config did not collect its own .gitignore")
	}
	if snapshot.jsonConfig.IsFileIgnored(childTarget, root) {
		t.Fatal("JSON fallback crossed the child JS config's .gitignore source boundary")
	}
}

func TestHandleConfigRefreshInitialPluginHostFailureCommitsNativeCatalog(t *testing.T) {
	config.RegisterAllRules()
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
	if !s.configDiscoveryV2HasLastGood || s.eslintPluginConfigGeneration != activationRequest.TransactionID {
		t.Fatalf("degraded generation was not committed: response=%+v generation=%q hasLastGood=%t", completed.response, s.eslintPluginConfigGeneration, s.configDiscoveryV2HasLastGood)
	}
	entries := s.jsConfigs[root]
	if value, found := configRuleValue(entries, "no-debugger"); !found || value != "error" {
		t.Fatalf("native catalog was lost with plugin host: %+v", s.jsConfigs)
	}
}

func TestHandleConfigRefreshActivationFailureAbortsAndKeepsLastGood(t *testing.T) {
	config.RegisterAllRules()
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
	if !s.configDiscoveryV2Active {
		t.Fatal("valid v2 configRefresh did not enable ancestor watcher retries")
	}
	if s.eslintPluginConfigGeneration != "last-good" || s.jsConfigs[root][0].Rules["no-console"] != "error" {
		t.Fatalf("failed activation replaced last-good state: generation=%q configs=%+v", s.eslintPluginConfigGeneration, s.jsConfigs)
	}
}

func TestHandleConfigRefreshCommitFailureAbortsAndKeepsLastGood(t *testing.T) {
	config.RegisterAllRules()
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
}

func TestHandleConfigRefreshInitialAllFailedCommitsUnavailableBoundaries(t *testing.T) {
	config.RegisterAllRules()
	s, outgoing, root := newConfigRefreshTestServer(t)
	writeConfigCandidate(t, root)
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
	if completed.response.TransactionID != loadRequest.TransactionID || completed.response.ConfigCount != 0 {
		t.Fatalf("unexpected recovery response: %+v", completed.response)
	}
	if s.configDiscoveryV2HasLastGood {
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
	outside := documentURIFromPath(filepath.Join(filepath.Dir(root), "outside.ts"))
	outsideConfig, _, outsideIsJS := s.getConfigForURI(outside)
	if value, found := configRuleValue(outsideConfig, "no-console"); outsideIsJS || !found || value != "error" {
		t.Fatalf("JSON fallback outside broken boundary = %+v, isJS=%t", outsideConfig, outsideIsJS)
	}
}

func TestHandleConfigRefreshPartialCatalogCommitsUnavailableParentAndUsableChild(t *testing.T) {
	config.RegisterAllRules()
	s, outgoing, root := newConfigRefreshTestServer(t)
	nested := tspath.NormalizePath(filepath.Join(root, "packages", "app"))
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	writeConfigCandidate(t, root)
	writeConfigCandidate(t, nested)
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
	if completed.response.ConfigCount != 1 || !s.configDiscoveryV2HasLastGood {
		t.Fatalf("partial recovery state = response %+v, hasLastGood=%t", completed.response, s.configDiscoveryV2HasLastGood)
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
		t.Fatalf("broken root fell through to JSON: entries=%+v isJS=%t", entries, isJS)
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
	resolver := config.NewConfigOwnerResolver(catalog.Configs, fsy)
	if owner, _ := resolver.Resolve(filepath.Join(aliasRoot, "src", "index.ts")); owner != realRoot {
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
	config.RegisterAllRules()
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
	if s.eslintPluginConfigGeneration != "last-good" || !s.configDiscoveryV2HasLastGood ||
		s.jsConfigs[root][0].Rules["no-console"] != "error" {
		t.Fatalf("all-failed refresh replaced last-good state: generation=%q configs=%+v", s.eslintPluginConfigGeneration, s.jsConfigs)
	}
}

func TestHandleConfigRefreshPartialFailureAtCommittedBoundaryAborts(t *testing.T) {
	config.RegisterAllRules()
	s, outgoing, root := newConfigRefreshTestServer(t)
	nested := tspath.NormalizePath(filepath.Join(root, "packages", "app"))
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	writeConfigCandidate(t, root)
	writeConfigCandidate(t, nested)
	installLastGoodConfig(s, root)
	s.jsConfigs[nested] = config.RslintConfig{{Rules: config.Rules{"old-nested": "error"}}}
	s.jsConfigOwnerResolver = config.NewConfigOwnerResolver(s.jsConfigs, s.fs)
	s.jsConfigKeyByPath[nested] = nested
	s.tsConfigPathsByConfig[nested] = nil

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
	config.RegisterAllRules()
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
	config.RegisterAllRules()
	s, outgoing, root := newConfigRefreshTestServer(t)
	writeConfigCandidate(t, root)

	first := startConfigRefreshForTest(s, "initial")
	firstLoad := nextConfigReverseRequest(t, outgoing, methodLoadConfigs)
	firstLoadRequest, firstLoadResponse := loadedConfigResponse(t, firstLoad, config.RslintConfig{{}})
	respondToConfigReverseRequest(t, s, firstLoad, firstLoadResponse, nil)
	firstActivation := nextConfigReverseRequest(t, outgoing, methodActivateConfigs)
	_, firstActivationResponse := activationResponseForRequest(t, firstActivation, true)
	respondToConfigReverseRequest(t, s, firstActivation, firstActivationResponse, nil)
	firstCommit := nextConfigReverseRequest(t, outgoing, methodCommitConfigs)
	respondToConfigReverseRequest(t, s, firstCommit, commitResponseForRequest(t, firstCommit, true), nil)
	if completed := awaitConfigRefreshResult(t, first); completed.err != nil {
		t.Fatalf("initial configRefresh failed: %v", completed.err)
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
	if len(s.jsConfigs) != 0 || completed.response.ConfigCount != 0 {
		t.Fatalf("deleted config remained committed: response=%+v configs=%+v", completed.response, s.jsConfigs)
	}
	if s.eslintPluginConfigGeneration != secondActivationRequest.TransactionID {
		t.Fatalf("empty catalog generation = %q, want %q", s.eslintPluginConfigGeneration, secondActivationRequest.TransactionID)
	}
	if s.configDiscoveryV2HasLastGood {
		t.Fatal("empty JavaScript catalog was marked as a usable JavaScript last-good generation")
	}

	// A newly created broken config now establishes an unavailable JS boundary
	// instead of preserving the empty catalog's JSON fallback.
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
	if completed.response.TransactionID != thirdLoadRequest.TransactionID || completed.response.ConfigCount != 0 {
		t.Fatalf("unavailable recovery response = %+v", completed.response)
	}
	if _, unavailable := s.jsUnavailableConfigs[root]; !unavailable || s.configDiscoveryV2HasLastGood {
		t.Fatalf("new broken config state: unavailable=%t hasLastGood=%t", unavailable, s.configDiscoveryV2HasLastGood)
	}
	entries, _, isJS := s.getConfigForURI(documentURIFromPath(filepath.Join(root, "src", "index.ts")))
	if !isJS || hasPublicConfigContent(entries) {
		t.Fatalf("new broken config fell through its empty JS boundary: entries=%+v isJS=%t", entries, isJS)
	}
}

func TestHandleConfigRefreshFailureKeepsCommittedIgnoreAfterDeletion(t *testing.T) {
	config.RegisterAllRules()
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

	targetURI := documentURIFromPath(target)
	isIgnored := func() bool {
		effective, cwd, _ := s.getLintConfigForURI(targetURI)
		return effective.IsFileIgnored(target, cwd)
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
	config.RegisterAllRules()
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

	targetURI := documentURIFromPath(target)
	isIgnored := func() bool {
		effective, cwd, _ := s.getLintConfigForURI(targetURI)
		return effective.IsFileIgnored(target, cwd)
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

func TestGitignoreWatcherRetriesV2AndPreservesLastGoodOnFailure(t *testing.T) {
	config.RegisterAllRules()
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
	s.configDiscoveryV2Active = true
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
	if handlers()[lsproto.Method("rslint/configUpdate")] == nil {
		t.Fatal("legacy rslint/configUpdate fallback was removed")
	}

	s, outgoing := newTestServerWithQueue()
	id := jsonrpc.NewIDString("config-capabilities-v2")
	req := &lsproto.RequestMessage{ID: id, Method: lsproto.Method("rslint/configCapabilities")}
	if err := handlers()[req.Method](s, context.Background(), req); err != nil {
		t.Fatal(err)
	}
	select {
	case message := <-outgoing:
		response := message.AsResponse()
		data, err := stdjson.Marshal(response.Result)
		if err != nil {
			t.Fatal(err)
		}
		var capabilities struct {
			TransactionVersion int `json:"transactionVersion"`
		}
		if err := stdjson.Unmarshal(data, &capabilities); err != nil {
			t.Fatal(err)
		}
		if capabilities.TransactionVersion != 2 {
			t.Fatalf("transactionVersion = %d, want 2", capabilities.TransactionVersion)
		}
	case <-time.After(time.Second):
		t.Fatal("configCapabilities was not acknowledged")
	}
}

func TestHandleConfigRefreshRejectsInvalidProtocolWithoutMutation(t *testing.T) {
	s := newTestServer()
	s.eslintPluginConfigGeneration = "last-good"
	_, err := s.handleConfigRefresh(context.Background(), map[string]any{
		"protocolVersion": discovery.ConfigDiscoveryProtocolVersion + 1,
		"reason":          "initial",
	})
	if err == nil || !strings.Contains(err.Error(), "unsupported config refresh protocol") {
		t.Fatalf("configRefresh error = %v", err)
	}
	if s.eslintPluginConfigGeneration != "last-good" {
		t.Fatal("invalid protocol mutated committed generation")
	}
	if s.configDiscoveryV2Active {
		t.Fatal("invalid protocol enabled v2 reverse watcher transactions")
	}
}
