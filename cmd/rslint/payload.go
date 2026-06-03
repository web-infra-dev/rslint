package main

import (
	"encoding/json"
	"fmt"

	"github.com/microsoft/typescript-go/shim/tspath"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
)

// configPayloadEntry represents a single config with its directory.
type configPayloadEntry struct {
	ConfigDirectory string                   `json:"configDirectory"`
	Entries         rslintconfig.RslintConfig `json:"entries"`
}

// parsedPayload is the result of parsing a config-stdin payload.
type parsedPayload struct {
	// configMap maps normalized configDirectory to config entries (multi-config mode).
	// nil when using legacy single-config mode.
	ConfigMap map[string]rslintconfig.RslintConfig

	// OriginalConfigDir maps each normalized ConfigMap key back to the raw
	// configDirectory string the JS host sent. The eslint-plugin wire configKey
	// uses this RAW form so it is byte-identical to the worker's plugin map key
	// (the host keys that map on the same raw string); Go's normalized key is
	// only for its own file matching. This makes the CLI round-trip the routing
	// token faithfully — the same property the LSP path has — instead of two
	// sides independently normalizing the path and having to agree. nil in
	// legacy single-config mode.
	OriginalConfigDir map[string]string

	// singleConfig and singleConfigDir are set in legacy single-config mode.
	SingleConfig    rslintconfig.RslintConfig
	SingleConfigDir string

	// IsMultiConfig indicates whether the payload used the multi-config format.
	IsMultiConfig bool
}

// parseConfigPayload parses the JSON payload from --config-stdin.
// It supports both the new multi-config format ({ configs: [...] })
// and the legacy single-config format ({ configDirectory, entries }).
func parseConfigPayload(data []byte) (*parsedPayload, error) {
	// Try multi-config format: { configs: [...] }
	var multiPayload struct {
		Configs []configPayloadEntry `json:"configs"`
	}
	if err := json.Unmarshal(data, &multiPayload); err != nil {
		return nil, fmt.Errorf("error parsing config from stdin: %w", err)
	}

	if len(multiPayload.Configs) > 0 {
		configMap := make(map[string]rslintconfig.RslintConfig, len(multiPayload.Configs))
		originalConfigDir := make(map[string]string, len(multiPayload.Configs))
		for _, cfg := range multiPayload.Configs {
			// configMap is keyed by the normalized dir (Go matches normalized
			// file paths against it); originalConfigDir recovers the raw string
			// for the eslint-plugin wire configKey. Both are populated in this
			// one loop from the same NormalizePath call, so they cannot disagree.
			configDir := tspath.NormalizePath(cfg.ConfigDirectory)
			configMap[configDir] = cfg.Entries
			originalConfigDir[configDir] = cfg.ConfigDirectory
		}
		return &parsedPayload{
			ConfigMap:         configMap,
			OriginalConfigDir: originalConfigDir,
			IsMultiConfig:     true,
		}, nil
	}

	// Fall back to legacy single-config format: { configDirectory, entries }
	var singlePayload configPayloadEntry
	if err := json.Unmarshal(data, &singlePayload); err != nil {
		return nil, fmt.Errorf("error parsing config from stdin: %w", err)
	}

	return &parsedPayload{
		SingleConfig:    singlePayload.Entries,
		SingleConfigDir: tspath.NormalizePath(singlePayload.ConfigDirectory),
		IsMultiConfig:   false,
	}, nil
}
