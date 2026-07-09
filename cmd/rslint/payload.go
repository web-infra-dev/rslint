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
	TargetFiles     []string                  `json:"targetFiles,omitempty"`
}

// parsedPayload is the result of parsing the serialized JS-config payload.
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

	// ConfigTargetFiles restricts a config to explicit target files when the JS
	// host had to keep a nearest config that directory discovery would have
	// skipped because of a parent global ignore.
	ConfigTargetFiles map[string][]string

	// singleConfig and singleConfigDir are set in legacy single-config mode.
	SingleConfig    rslintconfig.RslintConfig
	SingleConfigDir string

	// IsMultiConfig indicates whether the payload used the multi-config format.
	IsMultiConfig bool
}

// parseConfigPayload parses the serialized JS-config payload.
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
		configTargetFiles := make(map[string][]string)
		for _, cfg := range multiPayload.Configs {
			// configMap is keyed by the normalized dir (Go matches normalized
			// file paths against it); originalConfigDir recovers the raw string
			// for the eslint-plugin wire configKey. Both are populated in this
			// one loop from the same NormalizePath call, so they cannot disagree.
			configDir := tspath.NormalizePath(cfg.ConfigDirectory)
			if err := rslintconfig.ValidateConfig(cfg.Entries); err != nil {
				return nil, fmt.Errorf("invalid config for %q: %w", cfg.ConfigDirectory, err)
			}
			configMap[configDir] = cfg.Entries
			originalConfigDir[configDir] = cfg.ConfigDirectory
			if len(cfg.TargetFiles) > 0 {
				files := make([]string, 0, len(cfg.TargetFiles))
				for _, file := range cfg.TargetFiles {
					if tspath.PathIsAbsolute(file) {
						files = append(files, tspath.NormalizePath(file))
					} else {
						files = append(files, tspath.ResolvePath(configDir, file))
					}
				}
				configTargetFiles[configDir] = files
			}
		}
		return &parsedPayload{
			ConfigMap:         configMap,
			OriginalConfigDir: originalConfigDir,
			ConfigTargetFiles: configTargetFiles,
			IsMultiConfig:     true,
		}, nil
	}

	// Fall back to legacy single-config format: { configDirectory, entries }
	var singlePayload configPayloadEntry
	if err := json.Unmarshal(data, &singlePayload); err != nil {
		return nil, fmt.Errorf("error parsing config from stdin: %w", err)
	}
	if err := rslintconfig.ValidateConfig(singlePayload.Entries); err != nil {
		return nil, fmt.Errorf("invalid config for %q: %w", singlePayload.ConfigDirectory, err)
	}

	return &parsedPayload{
		SingleConfig:    singlePayload.Entries,
		SingleConfigDir: tspath.NormalizePath(singlePayload.ConfigDirectory),
		IsMultiConfig:   false,
	}, nil
}
