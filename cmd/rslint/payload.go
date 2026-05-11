package main

import (
	"encoding/json"
	"fmt"

	"github.com/microsoft/typescript-go/shim/tspath"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
)

// configPayloadEntry represents a single config with its directory.
type configPayloadEntry struct {
	ConfigDirectory string                    `json:"configDirectory"`
	Entries         rslintconfig.RslintConfig `json:"entries"`
}

// parsedPayload is the result of parsing a stdin config payload.
type parsedPayload struct {
	// configMap maps normalized configDirectory to config entries (multi-config mode).
	// nil when using legacy single-config mode.
	ConfigMap map[string]rslintconfig.RslintConfig

	// singleConfig and singleConfigDir are set in legacy single-config mode.
	SingleConfig    rslintconfig.RslintConfig
	SingleConfigDir string

	// IsMultiConfig indicates whether the payload used the multi-config format.
	IsMultiConfig bool
}

// parseConfigPayload parses the JSON config payload that
// `executeLintPipeline` reads from stdin. The pipeline accepts two
// shapes:
//   - multi-config:   `{ configs: [{configDirectory, entries}, ...] }`
//   - single-config:  `{ configDirectory, entries }` (kept for
//     backward compatibility with older callers).
//
// In the IPC entry path (runCLI) the payload is synthesized from the
// `init` message's Configs field; the pipeline doesn't care which
// transport delivered the bytes.
func parseConfigPayload(data []byte) (*parsedPayload, error) {
	// Step 1: detect whether the wire shape is the multi-config one by
	// asking the JSON layer directly whether `configs` was present.
	// `json.Unmarshal` with `Configs []T` decodes:
	//   - missing / `null`     → `Configs == nil`
	//   - `"configs": []`      → `Configs != nil, len == 0`
	//   - `"configs": [...]`   → `Configs != nil, len > 0`
	// The previous logic conflated empty-array with missing and silently
	// fell through to the legacy single-config decoder, which then
	// produced an empty single-config (no rules, no files). That is a
	// silent no-lint, not an obvious failure. Distinguish them by the
	// nil-vs-non-nil signal so an explicit `{configs:[]}` lands in the
	// multi-config branch (with zero configs — equally valid).
	var multiPayload struct {
		Configs []configPayloadEntry `json:"configs"`
	}
	if err := json.Unmarshal(data, &multiPayload); err != nil {
		return nil, fmt.Errorf("error parsing config from stdin: %w", err)
	}

	if multiPayload.Configs != nil {
		configMap := make(map[string]rslintconfig.RslintConfig, len(multiPayload.Configs))
		for _, cfg := range multiPayload.Configs {
			configDir := tspath.NormalizePath(cfg.ConfigDirectory)
			configMap[configDir] = cfg.Entries
		}
		return &parsedPayload{
			ConfigMap:     configMap,
			IsMultiConfig: true,
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
