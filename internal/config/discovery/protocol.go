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
	ID                string                           `json:"id"`
	Status            string                           `json:"status"`
	Entries           rslintconfig.RslintConfig        `json:"entries,omitempty"`
	SourceFingerprint string                           `json:"sourceFingerprint,omitempty"`
	EslintPlugins     []rslintconfig.EslintPluginEntry `json:"eslintPlugins,omitempty"`
	Error             *ConfigModuleError               `json:"error,omitempty"`
}

type ConfigLoadBatchResponse struct {
	TransactionID string             `json:"transactionId"`
	Results       []ConfigLoadResult `json:"results"`
}

// ConfigModuleLoader is the only JavaScript-aware boundary in config
// discovery. Implementations evaluate and normalize modules; Go validates the
// returned batch envelope and config entries before using them.
type ConfigModuleLoader interface {
	LoadConfigs(ctx context.Context, request ConfigLoadBatchRequest) (ConfigLoadBatchResponse, error)
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

// ConfigModuleActivator is optional. Native adapters implement it by asking
// the Node host to summarize exactly the effective IDs, which also closes the
// source-fingerprint race before a plugin host is prepared.
type ConfigModuleActivator interface {
	ActivateConfigs(ctx context.Context, request ConfigActivationRequest) (ConfigActivationResponse, error)
}
