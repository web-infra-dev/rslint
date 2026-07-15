package lsp

import (
	"context"
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
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"

	"github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/config/discovery"
)

type configRefreshTestResult struct {
	response configRefreshResponse
	err      error
}

// snapshotConfigModuleLoader lets snapshot tests obtain the same evaluator-
// bearing catalog that production receives from discovery.Build. Constructing
// ConfigCatalog literals would bypass the live ConfigArray instance whose
// identity the LSP is required to retain.
type snapshotConfigModuleLoader struct {
	configs map[string]config.RslintConfig
}

func (loader *snapshotConfigModuleLoader) LoadConfigs(
	_ context.Context,
	request discovery.ConfigLoadBatchRequest,
) (discovery.ConfigLoadBatchResponse, error) {
	response := discovery.ConfigLoadBatchResponse{TransactionID: request.TransactionID}
	for _, candidate := range request.Candidates {
		response.Results = append(response.Results, discovery.ConfigLoadResult{
			ID:      candidate.ID,
			Status:  "loaded",
			Entries: loader.configs[config.NormalizeHostPath(candidate.ConfigDirectory)],
		})
	}
	return response, nil
}

func (*snapshotConfigModuleLoader) EvaluateConfigPredicates(
	_ context.Context,
	request discovery.ConfigPredicateBatchRequest,
) (discovery.ConfigPredicateBatchResponse, error) {
	response := discovery.ConfigPredicateBatchResponse{TransactionID: request.TransactionID}
	for _, call := range request.Calls {
		response.Results = append(response.Results, discovery.ConfigPredicateResult{
			CallID: call.CallID,
			Status: "evaluated",
		})
	}
	return response, nil
}

func (*snapshotConfigModuleLoader) ActivateConfigs(
	_ context.Context,
	request discovery.ConfigActivationRequest,
) (discovery.ConfigActivationResponse, error) {
	return discovery.ConfigActivationResponse{TransactionID: request.TransactionID}, nil
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
	return startConfigRefreshForTestWithContext(context.Background(), s, reason)
}

func startConfigRefreshForTestWithContext(
	ctx context.Context,
	s *Server,
	reason string,
) <-chan configRefreshTestResult {
	result := make(chan configRefreshTestResult, 1)
	go func() {
		response, err := s.handleConfigRefresh(ctx, map[string]any{
			"reason": reason,
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
	s.jsConfigEvaluators = map[string]*config.ConfigEvaluator{
		root: config.NewConfigEvaluatorWithGitignore(
			discoveredJSConfigForTest(entries),
			root,
			s.fs,
			nil,
		),
	}
	s.jsConfigOwnerResolver = config.NewConfigOwnerResolver(s.jsConfigs, s.fs)
	s.jsUnavailableConfigs = make(map[string]struct{})
	s.tsConfigPathsByConfig = map[string][]string{root: nil}
	s.eslintPluginConfigGeneration = "last-good"
	s.configDiscoveryHasLastGood = true
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
	if got := s.pluginConfigKeyForURI(fileURI); got != root {
		t.Fatalf("plugin configKey = %q, want exact catalog path %q", got, root)
	}
	if _, active := s.eslintPluginRules[pluginRuleName]; !active {
		t.Fatalf("committed plugin rule set does not contain %q", pluginRuleName)
	}
	registered, ok := config.GlobalRuleRegistry.GetRule(pluginRuleName)
	if !ok || !registered.IsEslintPluginRule {
		t.Fatalf("plugin placeholder %q was not registered: %+v", pluginRuleName, registered)
	}
}

func TestCommittedConfigRefreshKeepsLivePredicateSessionForLaterFiles(t *testing.T) {
	config.RegisterAllRules()
	s, outgoing, root := newConfigRefreshTestServer(t)
	writeConfigCandidate(t, root)

	refreshCtx, cancelRefresh := context.WithCancel(context.Background())
	refresh := startConfigRefreshForTestWithContext(refreshCtx, s, "initial")
	loadMessage := nextConfigReverseRequest(t, outgoing, methodLoadConfigs)
	loadRequest, ok := loadMessage.Params.(discovery.ConfigLoadBatchRequest)
	if !ok || len(loadRequest.Candidates) != 1 {
		t.Fatalf("loadConfigs params = %#v", loadMessage.Params)
	}
	// Respond in the actual JSON wire shape so ConfigLoadResult's trusted
	// decoder, rather than an in-process ConfigEntry copy, creates the opaque
	// live-predicate descriptor.
	loadResponse := map[string]any{
		"transactionId": loadRequest.TransactionID,
		"results": []any{map[string]any{
			"id":     loadRequest.Candidates[0].ID,
			"status": "loaded",
			"entries": []any{map[string]any{
				"files": []any{map[string]any{"$rslintPredicate": "selector-1"}},
				"rules": map[string]any{"no-debugger": "error"},
			}},
		}},
	}
	respondToConfigReverseRequest(t, s, loadMessage, loadResponse, nil)
	activationMessage := nextConfigReverseRequest(t, outgoing, methodActivateConfigs)
	_, activationResponse := activationResponseForRequest(t, activationMessage, true)
	respondToConfigReverseRequest(t, s, activationMessage, activationResponse, nil)
	commitMessage := nextConfigReverseRequest(t, outgoing, methodCommitConfigs)
	respondToConfigReverseRequest(t, s, commitMessage, commitResponseForRequest(t, commitMessage, true), nil)
	if completed := awaitConfigRefreshResult(t, refresh); completed.err != nil {
		t.Fatalf("configRefresh failed: %v", completed.err)
	}
	// Production dispatch cancels the request context immediately after sending
	// the configRefresh response. The committed evaluator/session must not inherit
	// that short-lived cancellation boundary.
	cancelRefresh()
	t.Cleanup(func() {
		if s.activeConfigCatalog != nil {
			s.activeConfigCatalog.ClosePredicateEvaluation()
		}
	})

	type selectionResult struct {
		selection lspLintConfigSelection
		err       error
	}
	resolve := func(filePath string, predicateValue bool) lspLintConfigSelection {
		t.Helper()
		result := make(chan selectionResult, 1)
		go func() {
			selection, resolveErr := s.resolveLintConfigForURI(
				context.Background(),
				documentURIFromPath(filePath),
			)
			result <- selectionResult{selection: selection, err: resolveErr}
		}()

		var predicateMessage *lsproto.RequestMessage
		select {
		case message := <-outgoing:
			predicateMessage = message.AsRequest()
			if predicateMessage.Method != methodEvaluateConfigPredicates {
				t.Fatalf("reverse request method = %q, want %q", predicateMessage.Method, methodEvaluateConfigPredicates)
			}
		case resolved := <-result:
			t.Fatalf("selection completed without predicate request: selection=%+v err=%v", resolved.selection, resolved.err)
		case <-time.After(5 * time.Second):
			t.Fatalf("timed out waiting for predicate request for %q", filePath)
		}
		predicateRequest, ok := predicateMessage.Params.(discovery.ConfigPredicateBatchRequest)
		if !ok {
			t.Fatalf("evaluateConfigPredicates params type = %T", predicateMessage.Params)
		}
		if predicateRequest.TransactionID != loadRequest.TransactionID || len(predicateRequest.Calls) != 1 {
			t.Fatalf("predicate request = %+v", predicateRequest)
		}
		call := predicateRequest.Calls[0]
		if call.PredicateID != "selector-1" || call.AbsolutePath != tspath.NormalizePath(filePath) || call.Directory {
			t.Fatalf("predicate call = %+v", call)
		}
		respondToConfigReverseRequest(t, s, predicateMessage, discovery.ConfigPredicateBatchResponse{
			TransactionID: predicateRequest.TransactionID,
			Results: []discovery.ConfigPredicateResult{{
				CallID: call.CallID,
				Status: "evaluated",
				Value:  predicateValue,
			}},
		}, nil)
		resolved := <-result
		if resolved.err != nil {
			t.Fatalf("resolveLintConfigForURI: %v", resolved.err)
		}
		return resolved.selection
	}

	keepPath := tspath.NormalizePath(filepath.Join(root, "keep.ts"))
	keep := resolve(keepPath, true)
	if keep.merged == nil {
		t.Fatal("keep selection is unconfigured")
	}
	keepRule := keep.merged.Rules["no-debugger"]
	if keepRule == nil || keepRule.GetLevel() != "error" {
		t.Fatalf("keep selection = %+v", keep.merged)
	}
	// ConfigArray's exact-path cache must serve every later native/plugin/fix
	// consumer without another reverse predicate call.
	keepAgain, err := s.resolveLintConfigForURI(context.Background(), documentURIFromPath(keepPath))
	if err != nil || keepAgain.merged != keep.merged {
		t.Fatalf("cached keep selection = %+v, err=%v", keepAgain.merged, err)
	}
	select {
	case message := <-outgoing:
		t.Fatalf("cached selection triggered reverse request %q", message.AsRequest().Method)
	case <-time.After(100 * time.Millisecond):
	}

	dropPath := tspath.NormalizePath(filepath.Join(root, "drop.ts"))
	drop := resolve(dropPath, false)
	if drop.merged == nil {
		t.Fatal("drop selection lost the default script configuration")
	}
	if rule := drop.merged.Rules["no-debugger"]; rule != nil {
		t.Fatalf("drop selection retained predicate rule: %+v", rule)
	}
}

func TestHandleDidOpenRefreshesExactHiddenConfigOncePerDirectory(t *testing.T) {
	config.RegisterAllRules()
	s, outgoing, root := newConfigRefreshTestServer(t)
	writeConfigCandidate(t, root)
	hidden := tspath.NormalizePath(filepath.Join(root, "hidden", "app"))
	if err := os.MkdirAll(hidden, 0o755); err != nil {
		t.Fatal(err)
	}
	writeConfigCandidate(t, hidden)

	installLastGoodConfig(s, root)
	s.jsConfigs[root] = config.RslintConfig{
		{Ignores: []string{"hidden/**"}},
		{Rules: config.Rules{"no-console": "error"}},
	}
	s.jsConfigOwnerResolver = config.NewConfigOwnerResolver(s.jsConfigs, s.fs)
	s.configDiscoveryActive = true
	s.configDiscoveryLookupDirs = make(map[string]struct{})

	documentPath := tspath.NormalizePath(filepath.Join(hidden, "virtual.ts"))
	didOpenDone := make(chan error, 1)
	go func() {
		didOpenDone <- s.handleDidOpen(context.Background(), &lsproto.DidOpenTextDocumentParams{
			TextDocument: &lsproto.TextDocumentItem{
				Uri:        documentURIFromPath(documentPath),
				LanguageId: "typescript",
				Version:    1,
				Text:       "debugger;\n",
			},
		})
	}()

	committed := false
	for !committed {
		select {
		case message := <-outgoing:
			request := message.AsRequest()
			switch request.Method {
			case methodLoadConfigs:
				loadRequest, ok := request.Params.(discovery.ConfigLoadBatchRequest)
				if !ok {
					t.Fatalf("loadConfigs params type = %T", request.Params)
				}
				results := make([]discovery.ConfigLoadResult, len(loadRequest.Candidates))
				for index, candidate := range loadRequest.Candidates {
					entries := config.RslintConfig{{Rules: config.Rules{"no-console": "error"}}}
					if candidate.ConfigDirectory == hidden {
						entries = config.RslintConfig{{Rules: config.Rules{"no-debugger": "error"}}}
					}
					results[index] = discovery.ConfigLoadResult{
						ID:      candidate.ID,
						Status:  "loaded",
						Entries: entries,
					}
				}
				respondToConfigReverseRequest(t, s, request, discovery.ConfigLoadBatchResponse{
					TransactionID: loadRequest.TransactionID,
					Results:       results,
				}, nil)
			case methodActivateConfigs:
				_, response := activationResponseForRequest(t, request, true)
				respondToConfigReverseRequest(t, s, request, response, nil)
			case methodCommitConfigs:
				respondToConfigReverseRequest(t, s, request, commitResponseForRequest(t, request, true), nil)
				committed = true
			default:
				t.Fatalf("unexpected reverse request %q", request.Method)
			}
		case <-time.After(5 * time.Second):
			t.Fatal("timed out waiting for didOpen config refresh")
		}
	}
	if err := <-didOpenDone; err != nil {
		t.Fatalf("handleDidOpen: %v", err)
	}

	owner, entries := s.jsConfigOwnerResolver.Resolve(documentPath)
	if owner != hidden {
		t.Fatalf("owner = %q, want hidden exact config %q", owner, hidden)
	}
	if value, found := configRuleValue(entries, "no-debugger"); !found || value != "error" {
		t.Fatalf("hidden config was not committed: %+v", entries)
	}
	directoryID := lspLexicalPathID(hidden, s.fs.UseCaseSensitiveFileNames())
	if _, lookedUp := s.configDiscoveryLookupDirs[directoryID]; !lookedUp {
		t.Fatalf("open-document directory %q was not coalesced", hidden)
	}

	secondPath := tspath.NormalizePath(filepath.Join(hidden, "second.ts"))
	if err := s.handleDidOpen(context.Background(), &lsproto.DidOpenTextDocumentParams{
		TextDocument: &lsproto.TextDocumentItem{
			Uri:        documentURIFromPath(secondPath),
			LanguageId: "typescript",
			Version:    1,
			Text:       "console.log(1);\n",
		},
	}); err != nil {
		t.Fatalf("second handleDidOpen: %v", err)
	}
	select {
	case message := <-outgoing:
		t.Fatalf("second open in the same directory triggered %q", message.AsRequest().Method)
	case <-time.After(100 * time.Millisecond):
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
	writeConfigCandidate := func(directory string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(directory, "rslint.config.mjs"), []byte("export default [];\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	writeConfigCandidate(root)
	writeConfigCandidate(child)

	fsys := bundled.WrapFS(osvfs.FS())
	catalog, err := discovery.Build(
		context.Background(),
		fsys,
		&snapshotConfigModuleLoader{configs: map[string]config.RslintConfig{
			root:  {{Rules: config.Rules{"no-console": "error"}}},
			child: {{Rules: config.Rules{"no-debugger": "error"}}},
		}},
		discovery.ConfigDiscoveryRequest{
			CWD:                     root,
			Mode:                    discovery.ConfigDiscoveryAuto,
			LookupPaths:             []string{rootTarget, childTarget},
			RetainUnconfiguredAreas: true,
			AllowMissingConfig:      true,
		},
	)
	if err != nil {
		t.Fatalf("build evaluator-bearing catalog: %v", err)
	}
	defer catalog.ClosePredicateEvaluation()
	s := newTestServer()
	s.cwd = root
	snapshot, err := s.prepareDiscoveredConfigSnapshot(fsys, catalog, nil)
	if err != nil {
		t.Fatal(err)
	}
	assertIgnoredByOwner := func(filePath string, wantOwner string) {
		t.Helper()
		resolvedDir, _ := snapshot.ownerResolver.Resolve(filePath)
		if resolvedDir != wantOwner {
			t.Fatalf("owner for %q = %q, want %q", filePath, resolvedDir, wantOwner)
		}
		evaluator := snapshot.evaluators[resolvedDir]
		if evaluator == nil {
			t.Fatalf("owner %q has no evaluator", resolvedDir)
		}
		resolution, err := evaluator.GetConfigForFile(context.Background(), filePath)
		if err != nil {
			t.Fatal(err)
		}
		if resolution.Status != config.ConfigFileIgnored {
			t.Fatalf("resolution for %q = %q, want ignored", filePath, resolution.Status)
		}
	}
	assertIgnoredByOwner(rootTarget, root)
	assertIgnoredByOwner(childTarget, child)
	if snapshot.jsonConfig.IsFileIgnored(childTarget, root) {
		t.Fatal("JSON fallback crossed the child JS config's .gitignore source boundary")
	}
}

func TestPrepareDiscoveredConfigSnapshotRejectsCatalogWithoutEvaluator(t *testing.T) {
	root := tspath.NormalizePath(t.TempDir())
	s := newTestServer()
	s.cwd = root
	_, err := s.prepareDiscoveredConfigSnapshot(
		bundled.WrapFS(osvfs.FS()),
		&discovery.ConfigCatalog{
			Configs: map[string]config.RslintConfig{
				root: {{Rules: config.Rules{"no-console": "error"}}},
			},
		},
		nil,
	)
	if err == nil || !strings.Contains(err.Error(), "has no committed evaluator") {
		t.Fatalf("error = %v, want missing evaluator invariant failure", err)
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
	if !s.configDiscoveryHasLastGood || s.eslintPluginConfigGeneration != activationRequest.TransactionID {
		t.Fatalf("degraded generation was not committed: response=%+v generation=%q hasLastGood=%t", completed.response, s.eslintPluginConfigGeneration, s.configDiscoveryHasLastGood)
	}
	entries := s.jsConfigs[root]
	if value, found := configRuleValue(entries, "no-debugger"); !found || value != "error" {
		t.Fatalf("native catalog was lost with plugin host: %+v", s.jsConfigs)
	}
	if len(s.eslintPluginRules) != 0 {
		t.Fatalf("degraded plugin host committed unroutable plugin rules: %+v", s.eslintPluginRules)
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
	if !s.configDiscoveryActive {
		t.Fatal("valid configRefresh did not enable ancestor watcher retries")
	}
	if s.eslintPluginConfigGeneration != "last-good" || s.jsConfigs[root][0].Rules["no-console"] != "error" {
		t.Fatalf("failed activation replaced last-good state: generation=%q configs=%+v", s.eslintPluginConfigGeneration, s.jsConfigs)
	}
}

func TestHandleConfigRefreshInvalidRuleOptionsAbortsAndKeepsLastGood(t *testing.T) {
	config.RegisterAllRules()
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
	outside := documentURIFromPath(filepath.Join(filepath.Dir(root), "outside.ts"))
	outsideConfig, _, outsideIsJS := s.getConfigForURI(outside)
	if value, found := configRuleValue(outsideConfig, "no-console"); outsideIsJS || !found || value != "error" {
		t.Fatalf("JSON fallback outside broken boundary = %+v, isJS=%t", outsideConfig, outsideIsJS)
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
	configs := map[string]config.RslintConfig{realRoot: {{}}}
	resolver := config.NewConfigOwnerResolver(configs, fsy)
	if owner, _ := resolver.Resolve(filepath.Join(aliasRoot, "src", "index.ts")); owner != realRoot {
		t.Fatalf("test precondition: canonical resolver owner = %q, want %q", owner, realRoot)
	}
	boundaries := unavailableConfigBoundaryDirectories(
		fsy,
		[]discovery.ConfigFailure{{Directory: aliasRoot}},
	)
	if len(boundaries) != 1 || boundaries[0] != aliasRoot {
		t.Fatalf("symlink failure boundaries = %v, want lexical alias %q", boundaries, aliasRoot)
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
	if s.eslintPluginConfigGeneration != "last-good" || !s.configDiscoveryHasLastGood ||
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

	abortMessage := nextConfigReverseRequest(t, outgoing, methodAbortConfigs)
	respondToConfigReverseRequest(t, s, abortMessage, abortResponseForRequest(t, abortMessage), nil)

	completed := awaitConfigRefreshResult(t, result)
	if !errors.Is(completed.err, discovery.ErrAllConfigsFailed) {
		t.Fatalf("partial refresh error = %v, want strict nearest-config failure", completed.err)
	}
	rootValue, rootOK := configRuleValue(s.jsConfigs[root], "no-console")
	nestedValue, nestedOK := configRuleValue(s.jsConfigs[nested], "old-nested")
	if s.eslintPluginConfigGeneration != "last-good" || !rootOK || rootValue != "error" || !nestedOK || nestedValue != "error" {
		t.Fatalf("partial boundary failure replaced last-good state: %+v", s.jsConfigs)
	}
}

func TestHandleConfigRefreshKnownUnavailableSiblingAllowsReplacement(t *testing.T) {
	config.RegisterAllRules()
	s, outgoing, root := newConfigRefreshTestServer(t)
	broken := tspath.NormalizePath(filepath.Join(root, "packages", "broken"))
	if err := os.MkdirAll(broken, 0o755); err != nil {
		t.Fatal(err)
	}
	writeConfigCandidate(t, root)
	writeConfigCandidate(t, broken)
	installLastGoodConfig(s, root)
	s.jsConfigs[broken] = config.RslintConfig{}
	s.jsUnavailableConfigs[broken] = struct{}{}
	s.jsConfigOwnerResolver = config.NewConfigOwnerResolver(s.jsConfigs, s.fs)
	s.tsConfigPathsByConfig[broken] = nil

	result := startConfigRefreshForTest(s, "config-change")
	rootLoad := nextConfigReverseRequest(t, outgoing, methodLoadConfigs)
	_, rootResponse := loadedConfigResponse(t, rootLoad, config.RslintConfig{{
		Rules: config.Rules{"no-debugger": "error"},
	}})
	respondToConfigReverseRequest(t, s, rootLoad, rootResponse, nil)

	brokenLoad := nextConfigReverseRequest(t, outgoing, methodLoadConfigs)
	_, brokenResponse := failedConfigResponse(t, brokenLoad, "known broken sibling")
	respondToConfigReverseRequest(t, s, brokenLoad, brokenResponse, nil)

	activation := nextConfigReverseRequest(t, outgoing, methodActivateConfigs)
	_, activationResponse := activationResponseForRequest(t, activation, true)
	respondToConfigReverseRequest(t, s, activation, activationResponse, nil)
	commit := nextConfigReverseRequest(t, outgoing, methodCommitConfigs)
	respondToConfigReverseRequest(t, s, commit, commitResponseForRequest(t, commit, true), nil)

	completed := awaitConfigRefreshResult(t, result)
	if completed.err != nil {
		t.Fatalf("known unavailable sibling blocked replacement: %v", completed.err)
	}
	if value, ok := configRuleValue(s.jsConfigs[root], "no-debugger"); !ok || value != "error" {
		t.Fatalf("updated root was not committed: %+v", s.jsConfigs[root])
	}
	if _, unavailable := s.jsUnavailableConfigs[broken]; !unavailable {
		t.Fatalf("known broken sibling lost unavailable boundary: %+v", s.jsUnavailableConfigs)
	}
	if !s.configDiscoveryHasLastGood {
		t.Fatal("partial replacement lost its usable last-good state")
	}
}

func TestConfigFailuresCoveredByCommittedUnavailableUsesPhysicalAlias(t *testing.T) {
	root := t.TempDir()
	realRoot := filepath.Join(root, "real-workspace")
	aliasRoot := filepath.Join(root, "alias-workspace")
	realBroken := filepath.Join(realRoot, "packages", "broken")
	if err := os.MkdirAll(realBroken, 0o755); err != nil {
		t.Fatal(err)
	}
	baseFS := bundled.WrapFS(osvfs.FS())
	canonicalRealRoot := baseFS.Realpath(tspath.NormalizePath(realRoot))
	fsys := &realpathAliasLSPTestFS{
		FS:        baseFS,
		aliasRoot: aliasRoot,
		realRoot:  canonicalRealRoot,
	}
	canonicalBroken := tspath.NormalizePath(filepath.Join(canonicalRealRoot, "packages", "broken"))
	aliasBroken := tspath.NormalizePath(filepath.Join(aliasRoot, "packages", "broken"))
	failures := []discovery.ConfigFailure{{Directory: aliasBroken}}

	if !configFailuresCoveredByCommittedUnavailable(
		fsys,
		failures,
		map[string]config.RslintConfig{canonicalBroken: {}},
		map[string]struct{}{canonicalBroken: {}},
	) {
		t.Fatal("physical alias of a committed unavailable boundary was rejected")
	}
	if configFailuresCoveredByCommittedUnavailable(
		fsys,
		failures,
		map[string]config.RslintConfig{
			canonicalRealRoot: {},
			canonicalBroken:   {{Rules: config.Rules{"no-debugger": "error"}}},
		},
		map[string]struct{}{canonicalRealRoot: {}},
	) {
		t.Fatal("physical alias of an active child inside an unavailable parent was accepted")
	}
}

func TestHandleConfigRefreshNewFailedBoundaryIsFatal(t *testing.T) {
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
	abortMessage := nextConfigReverseRequest(t, outgoing, methodAbortConfigs)
	respondToConfigReverseRequest(t, s, abortMessage, abortResponseForRequest(t, abortMessage), nil)

	completed := awaitConfigRefreshResult(t, result)
	if !errors.Is(completed.err, discovery.ErrAllConfigsFailed) {
		t.Fatalf("new failed child error = %v, want strict nearest-config failure", completed.err)
	}
	if value, ok := configRuleValue(s.jsConfigs[root], "no-console"); len(s.jsConfigs) != 1 || !ok || value != "error" {
		t.Fatalf("failed child replaced last-good catalog = %+v", s.jsConfigs)
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
	if len(s.jsConfigs) != 0 {
		t.Fatalf("deleted config remained committed: response=%+v configs=%+v", completed.response, s.jsConfigs)
	}
	if s.eslintPluginConfigGeneration != secondActivationRequest.TransactionID {
		t.Fatalf("empty catalog generation = %q, want %q", s.eslintPluginConfigGeneration, secondActivationRequest.TransactionID)
	}
	if s.configDiscoveryHasLastGood {
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
		selection, err := s.resolveLintConfigForURI(context.Background(), targetURI)
		if err != nil {
			t.Fatal(err)
		}
		return selection.merged == nil
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

func TestGitignoreWatcherRetriesRefreshAndPreservesLastGoodOnFailure(t *testing.T) {
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
