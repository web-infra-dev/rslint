package discovery

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/hostfs"
	"github.com/web-infra-dev/rslint/internal/hostpath"
)

type fakeConfigModuleLoader struct {
	mu          sync.Mutex
	configs     map[string]rslintconfig.RslintConfig
	failures    map[string]ConfigModuleError
	batches     []ConfigLoadBatchRequest
	activations []ConfigActivationRequest
}

func (loader *fakeConfigModuleLoader) LoadConfigs(_ context.Context, request ConfigLoadBatchRequest) (ConfigLoadBatchResponse, error) {
	loader.mu.Lock()
	loader.batches = append(loader.batches, request)
	loader.mu.Unlock()
	response := ConfigLoadBatchResponse{TransactionID: request.TransactionID}
	for _, candidate := range request.Candidates {
		path := candidate.ConfigPath
		failure, failed := loader.failures[path]
		if !failed {
			failure, failed = loader.failures[tspath.NormalizePath(path)]
		}
		if failed {
			failure := failure
			response.Results = append(response.Results, ConfigLoadResult{
				ID: candidate.ID, Status: "failed", Error: &failure,
			})
			continue
		}
		entries, found := loader.configs[path]
		if !found {
			entries = loader.configs[tspath.NormalizePath(path)]
		}
		response.Results = append(response.Results, ConfigLoadResult{
			ID: candidate.ID, Status: "loaded",
			Entries: append(rslintconfig.RslintConfig(nil), entries...),
		})
	}
	return response, nil
}

func (loader *fakeConfigModuleLoader) ActivateConfigs(_ context.Context, request ConfigActivationRequest) (ConfigActivationResponse, error) {
	loader.mu.Lock()
	loader.activations = append(loader.activations, request)
	loader.mu.Unlock()
	return ConfigActivationResponse{TransactionID: request.TransactionID}, nil
}

func (loader *fakeConfigModuleLoader) EvaluateConfigPredicates(
	_ context.Context,
	request ConfigPredicateBatchRequest,
) (ConfigPredicateBatchResponse, error) {
	response := ConfigPredicateBatchResponse{TransactionID: request.TransactionID}
	for _, call := range request.Calls {
		response.Results = append(response.Results, ConfigPredicateResult{
			CallID: call.CallID,
			Status: "evaluated",
		})
	}
	return response, nil
}

func newFixtureConfigLoader() *fakeConfigModuleLoader {
	return &fakeConfigModuleLoader{
		configs:  make(map[string]rslintconfig.RslintConfig),
		failures: make(map[string]ConfigModuleError),
	}
}

func discoveryTestFS() vfs.FS {
	return bundled.WrapFS(cachedvfs.From(hostfs.NativeOS(osvfs.FS())))
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
	return hostpath.Normalize(path)
}

func requestedConfigPaths(loader *fakeConfigModuleLoader) []string {
	loader.mu.Lock()
	defer loader.mu.Unlock()
	var paths []string
	for _, batch := range loader.batches {
		for _, candidate := range batch.Candidates {
			paths = append(paths, candidate.ConfigPath)
		}
	}
	return paths
}

func TestSearchDiscoveryConfigFilenamePriority(t *testing.T) {
	root := t.TempDir()
	jsPath := writeConfigCandidate(t, root, "rslint.config.js")
	mjsPath := writeConfigCandidate(t, root, "rslint.config.mjs")
	writeDiscoveryFixture(t, root, "index.ts", "export {}\n")
	loader := newFixtureConfigLoader()
	loader.configs[jsPath] = namedConfig("js")
	loader.configs[mjsPath] = namedConfig("mjs")

	catalog, err := Build(context.Background(), discoveryTestFS(), loader, ConfigDiscoveryRequest{
		CWD: root, Mode: ConfigDiscoveryAuto, Inputs: []string{"."},
		CollectTargets: true, GlobInputPaths: true, ErrorOnUnmatchedPattern: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := requestedConfigPaths(loader); len(got) != 1 || got[0] != jsPath {
		t.Fatalf("requested configs = %v, want priority candidate %q", got, jsPath)
	}
	if len(catalog.Configs) != 1 {
		t.Fatalf("configs = %v", catalog.ConfigDirectories())
	}
}

func TestConfigDiscoveryTransactionIDsAreUniqueAcrossBuilds(t *testing.T) {
	root := t.TempDir()
	const builds = 64
	results := make(chan string, builds)
	errs := make(chan error, builds)
	var waitGroup sync.WaitGroup
	waitGroup.Add(builds)
	for range builds {
		go func() {
			defer waitGroup.Done()
			catalog, err := Build(context.Background(), discoveryTestFS(), nil, ConfigDiscoveryRequest{
				CWD: root, Inputs: []string{}, GlobInputPaths: true,
			})
			if err != nil {
				errs <- err
				return
			}
			results <- catalog.TransactionID
		}()
	}
	waitGroup.Wait()
	close(results)
	close(errs)
	for err := range errs {
		t.Fatalf("Build: %v", err)
	}
	seen := make(map[string]struct{}, builds)
	for id := range results {
		if id == "" {
			t.Fatal("transaction ID is empty")
		}
		if _, duplicate := seen[id]; duplicate {
			t.Fatalf("duplicate transaction ID %q", id)
		}
		seen[id] = struct{}{}
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
		TransactionID: "txn",
		LoadMode:      ConfigModuleLoadCached,
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

func TestValidateConfigPredicateBatchRejectsProtocolViolations(t *testing.T) {
	request := ConfigPredicateBatchRequest{
		TransactionID: "tx",
		Calls: []ConfigPredicateCall{
			{CallID: "call-a", PredicateID: "a", AbsolutePath: "/repo/a.ts"},
			{CallID: "call-b", PredicateID: "b", AbsolutePath: "/repo/b.ts", Directory: true},
		},
	}
	valid := func() ConfigPredicateBatchResponse {
		return ConfigPredicateBatchResponse{
			TransactionID: "tx",
			Results: []ConfigPredicateResult{
				{CallID: "call-a", Status: "evaluated", Value: true},
				{CallID: "call-b", Status: "failed", Error: &ConfigModuleError{Message: "boom"}},
			},
		}
	}
	if results, err := validateConfigPredicateBatch(request, valid()); err != nil || len(results) != 2 {
		t.Fatalf("valid response = %v, %v", results, err)
	}

	tests := []struct {
		name   string
		mutate func(*ConfigPredicateBatchResponse)
	}{
		{name: "transaction", mutate: func(response *ConfigPredicateBatchResponse) { response.TransactionID = "other" }},
		{name: "count", mutate: func(response *ConfigPredicateBatchResponse) { response.Results = response.Results[:1] }},
		{name: "order", mutate: func(response *ConfigPredicateBatchResponse) {
			response.Results[0], response.Results[1] = response.Results[1], response.Results[0]
		}},
		{name: "unknown call", mutate: func(response *ConfigPredicateBatchResponse) { response.Results[0].CallID = "unknown" }},
		{name: "duplicate call", mutate: func(response *ConfigPredicateBatchResponse) { response.Results[1].CallID = "call-a" }},
		{name: "invalid status", mutate: func(response *ConfigPredicateBatchResponse) { response.Results[0].Status = "pending" }},
		{name: "evaluated with error", mutate: func(response *ConfigPredicateBatchResponse) {
			response.Results[0].Error = &ConfigModuleError{Message: "impossible"}
		}},
		{name: "failed without error", mutate: func(response *ConfigPredicateBatchResponse) { response.Results[1].Error = nil }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response := valid()
			test.mutate(&response)
			if _, err := validateConfigPredicateBatch(request, response); err == nil {
				t.Fatal("expected protocol error")
			} else {
				var protocolError *ConfigDiscoveryProtocolError
				if !errors.As(err, &protocolError) {
					t.Fatalf("error = %T %v, want ConfigDiscoveryProtocolError", err, err)
				}
			}
		})
	}
}

func TestConfigDirectoriesAreSorted(t *testing.T) {
	catalog := &ConfigCatalog{Configs: map[string]rslintconfig.RslintConfig{"/z": nil, "/a": nil}}
	got := catalog.ConfigDirectories()
	want := []string{"/a", "/z"}
	sort.Strings(got)
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("directories = %v, want %v", got, want)
		}
	}
}
