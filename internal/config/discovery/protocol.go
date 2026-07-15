package discovery

import (
	"context"
	"encoding/json"
	"fmt"

	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
)

type ConfigModuleLoadMode string

const (
	ConfigModuleLoadCached ConfigModuleLoadMode = "cached"
	ConfigModuleLoadFresh  ConfigModuleLoadMode = "fresh"
)

type ConfigLoadCandidate struct {
	ID         string `json:"id"`
	ConfigPath string `json:"configPath"`
	// ConfigDirectory is a Go-owned opaque routing identity. Node must return
	// and reuse its exact spelling rather than native-normalizing separators.
	ConfigDirectory string `json:"configDirectory"`
}

type ConfigLoadBatchRequest struct {
	TransactionID  string                `json:"transactionId"`
	LoadMode       ConfigModuleLoadMode  `json:"loadMode"`
	SingleThreaded bool                  `json:"singleThreaded,omitempty"`
	Candidates     []ConfigLoadCandidate `json:"candidates"`
}

type ConfigModuleError struct {
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

type ConfigLoadResult struct {
	ID      string                    `json:"id"`
	Status  string                    `json:"status"`
	Entries rslintconfig.RslintConfig `json:"entries,omitempty"`
	Error   *ConfigModuleError        `json:"error,omitempty"`
}

func (result *ConfigLoadResult) UnmarshalJSON(data []byte) error {
	var wire struct {
		ID      string             `json:"id"`
		Status  string             `json:"status"`
		Entries json.RawMessage    `json:"entries"`
		Error   *ConfigModuleError `json:"error"`
	}
	if err := json.Unmarshal(data, &wire); err != nil {
		return err
	}
	result.ID = wire.ID
	result.Status = wire.Status
	result.Error = wire.Error
	result.Entries = nil
	if len(wire.Entries) > 0 && string(wire.Entries) != "null" {
		entries, err := rslintconfig.DecodeModuleConfig(wire.Entries)
		if err != nil {
			return fmt.Errorf("decode trusted module config entries: %w", err)
		}
		result.Entries = entries
	}
	return nil
}

type ConfigLoadBatchResponse struct {
	TransactionID string             `json:"transactionId"`
	Results       []ConfigLoadResult `json:"results"`
}

// ConfigModuleLoader is the only JavaScript-aware boundary in config
// discovery. Implementations evaluate and normalize modules, then activate the
// final effective set; Go validates both protocol responses before using them.
type ConfigModuleLoader interface {
	LoadConfigs(ctx context.Context, request ConfigLoadBatchRequest) (ConfigLoadBatchResponse, error)
	EvaluateConfigPredicates(ctx context.Context, request ConfigPredicateBatchRequest) (ConfigPredicateBatchResponse, error)
	ActivateConfigs(ctx context.Context, request ConfigActivationRequest) (ConfigActivationResponse, error)
}

type ConfigPredicateCall struct {
	CallID       string `json:"callId"`
	PredicateID  string `json:"predicateId"`
	AbsolutePath string `json:"absolutePath"`
	Directory    bool   `json:"directory,omitempty"`
}

type ConfigPredicateBatchRequest struct {
	TransactionID string                `json:"transactionId"`
	Calls         []ConfigPredicateCall `json:"calls"`
}

type ConfigPredicateResult struct {
	CallID string             `json:"callId"`
	Status string             `json:"status"`
	Value  bool               `json:"value,omitempty"`
	Error  *ConfigModuleError `json:"error,omitempty"`
}

type ConfigPredicateBatchResponse struct {
	TransactionID string                  `json:"transactionId"`
	Results       []ConfigPredicateResult `json:"results"`
}

func validateConfigPredicateBatch(
	request ConfigPredicateBatchRequest,
	response ConfigPredicateBatchResponse,
) (map[string]ConfigPredicateResult, error) {
	if request.TransactionID == "" {
		return nil, configDiscoveryProtocolError("predicate request transactionId is empty")
	}
	if response.TransactionID != request.TransactionID {
		return nil, configDiscoveryProtocolError("predicate transaction mismatch: got %q, want %q", response.TransactionID, request.TransactionID)
	}
	if len(response.Results) != len(request.Calls) {
		return nil, configDiscoveryProtocolError("predicate result count mismatch: got %d, want %d", len(response.Results), len(request.Calls))
	}
	requested := make(map[string]struct{}, len(request.Calls))
	for _, call := range request.Calls {
		if call.CallID == "" || call.PredicateID == "" || call.AbsolutePath == "" {
			return nil, configDiscoveryProtocolError("predicate request contains an empty call field")
		}
		if _, duplicate := requested[call.CallID]; duplicate {
			return nil, configDiscoveryProtocolError("predicate request contains duplicate call id %q", call.CallID)
		}
		requested[call.CallID] = struct{}{}
	}
	results := make(map[string]ConfigPredicateResult, len(response.Results))
	for index, result := range response.Results {
		if _, exists := requested[result.CallID]; !exists {
			return nil, configDiscoveryProtocolError("predicate response contains unknown call id %q", result.CallID)
		}
		if _, duplicate := results[result.CallID]; duplicate {
			return nil, configDiscoveryProtocolError("predicate response contains duplicate call id %q", result.CallID)
		}
		if request.Calls[index].CallID != result.CallID {
			return nil, configDiscoveryProtocolError("predicate result order mismatch at index %d: got %q, want %q", index, result.CallID, request.Calls[index].CallID)
		}
		if result.Status != "evaluated" && result.Status != "failed" {
			return nil, configDiscoveryProtocolError("predicate call %q has invalid status %q", result.CallID, result.Status)
		}
		if result.Status == "evaluated" && result.Error != nil {
			return nil, configDiscoveryProtocolError("evaluated predicate call %q contains an error", result.CallID)
		}
		if result.Status == "failed" && (result.Error == nil || result.Error.Message == "") {
			return nil, configDiscoveryProtocolError("failed predicate call %q has no error message", result.CallID)
		}
		results[result.CallID] = result
	}
	return results, nil
}

type ConfigActivationRequest struct {
	TransactionID      string   `json:"transactionId"`
	EffectiveConfigIDs []string `json:"effectiveConfigIds"`
}

type ConfigActivationResponse struct {
	TransactionID       string                           `json:"transactionId"`
	EslintPluginEntries []rslintconfig.EslintPluginEntry `json:"eslintPluginEntries"`
}

func validateConfigLoadBatch(request ConfigLoadBatchRequest, response ConfigLoadBatchResponse) (map[string]ConfigLoadResult, error) {
	if request.TransactionID == "" {
		return nil, configDiscoveryProtocolError("request transactionId is empty")
	}
	if response.TransactionID != request.TransactionID {
		return nil, configDiscoveryProtocolError("transaction mismatch: got %q, want %q", response.TransactionID, request.TransactionID)
	}
	if len(response.Results) != len(request.Candidates) {
		return nil, configDiscoveryProtocolError("result count mismatch: got %d, want %d", len(response.Results), len(request.Candidates))
	}
	requestByID := make(map[string]ConfigLoadCandidate, len(request.Candidates))
	for _, candidate := range request.Candidates {
		if candidate.ID == "" {
			return nil, configDiscoveryProtocolError("request contains an empty candidate id")
		}
		if _, duplicate := requestByID[candidate.ID]; duplicate {
			return nil, configDiscoveryProtocolError("request contains duplicate candidate id %q", candidate.ID)
		}
		requestByID[candidate.ID] = candidate
	}
	results := make(map[string]ConfigLoadResult, len(response.Results))
	for index, result := range response.Results {
		if _, exists := requestByID[result.ID]; !exists {
			return nil, configDiscoveryProtocolError("response contains unknown candidate id %q", result.ID)
		}
		if _, duplicate := results[result.ID]; duplicate {
			return nil, configDiscoveryProtocolError("response contains duplicate candidate id %q", result.ID)
		}
		if request.Candidates[index].ID != result.ID {
			return nil, configDiscoveryProtocolError("result order mismatch at index %d: got id %q, want %q", index, result.ID, request.Candidates[index].ID)
		}
		if result.Status != "loaded" && result.Status != "failed" {
			return nil, configDiscoveryProtocolError("candidate %q has invalid status %q", result.ID, result.Status)
		}
		if result.Status == "loaded" && result.Error != nil {
			return nil, configDiscoveryProtocolError("loaded candidate %q contains an error", result.ID)
		}
		if result.Status == "failed" && (result.Error == nil || result.Error.Message == "") {
			return nil, configDiscoveryProtocolError("failed candidate %q has no error message", result.ID)
		}
		results[result.ID] = result
	}
	return results, nil
}

func cloneEslintPluginEntries(entries []rslintconfig.EslintPluginEntry) []rslintconfig.EslintPluginEntry {
	if len(entries) == 0 {
		return nil
	}
	cloned := make([]rslintconfig.EslintPluginEntry, len(entries))
	for index, entry := range entries {
		cloned[index] = rslintconfig.EslintPluginEntry{
			Prefix:    entry.Prefix,
			RuleNames: append([]string(nil), entry.RuleNames...),
		}
	}
	return cloned
}
