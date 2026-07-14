package discovery

import (
	"context"

	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
)

const ConfigDiscoveryProtocolVersion = 1

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
	ProtocolVersion int                   `json:"protocolVersion"`
	TransactionID   string                `json:"transactionId"`
	LoadMode        ConfigModuleLoadMode  `json:"loadMode"`
	SingleThreaded  bool                  `json:"singleThreaded,omitempty"`
	Candidates      []ConfigLoadCandidate `json:"candidates"`
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

type ConfigLoadBatchResponse struct {
	TransactionID string             `json:"transactionId"`
	Results       []ConfigLoadResult `json:"results"`
}

// ConfigModuleLoader is the only JavaScript-aware boundary in config
// discovery. Implementations evaluate and normalize modules, then activate the
// final effective set; Go validates both protocol responses before using them.
type ConfigModuleLoader interface {
	LoadConfigs(ctx context.Context, request ConfigLoadBatchRequest) (ConfigLoadBatchResponse, error)
	ActivateConfigs(ctx context.Context, request ConfigActivationRequest) (ConfigActivationResponse, error)
}

type ConfigActivationRequest struct {
	ProtocolVersion    int      `json:"protocolVersion"`
	TransactionID      string   `json:"transactionId"`
	EffectiveConfigIDs []string `json:"effectiveConfigIds"`
}

type ConfigActivationResponse struct {
	TransactionID       string                           `json:"transactionId"`
	EslintPluginEntries []rslintconfig.EslintPluginEntry `json:"eslintPluginEntries"`
}
