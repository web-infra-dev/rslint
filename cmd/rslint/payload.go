package main

import (
	"encoding/json"
	"fmt"

	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
)

// configPayloadEntry represents a single config with its directory.
type configPayloadEntry struct {
	ConfigDirectory string                    `json:"configDirectory"`
	Entries         rslintconfig.RslintConfig `json:"entries"`
	TargetFiles     []string                  `json:"targetFiles,omitempty"`
	ExplicitOnly    bool                      `json:"explicitOnly,omitempty"`
}

// parsedPayload is the result of parsing the serialized JS-config payload.
type parsedPayload struct {
	// configMap maps normalized configDirectory to config entries (multi-config mode).
	// nil when using legacy single-config mode.
	ConfigMap map[string]rslintconfig.RslintConfig

	// OriginalConfigDir maps each normalized ConfigMap key back to the raw
	// Go-owned configDirectory routing identity shared with the JS host. The eslint-plugin wire configKey
	// uses this RAW form so it is byte-identical to the worker's plugin map key
	// (the host keys that map on the same raw string); Go's normalized key is
	// only for its own file matching. This makes the CLI round-trip the routing
	// token faithfully — the same property the LSP path has — instead of two
	// sides independently normalizing the path and having to agree. nil in
	// legacy single-config mode.
	OriginalConfigDir map[string]string

	// ConfigTargetScopes carries explicit-file provenance established by Go
	// config discovery. Go uses it to preserve lexical ownership and to keep configs that
	// survived only for an explicit file out of directory discovery.
	ConfigTargetScopes map[string]rslintconfig.LintDiscoveryScope

	// singleConfig and singleConfigDir are set in legacy single-config mode.
	SingleConfig    rslintconfig.RslintConfig
	SingleConfigDir string

	// IsMultiConfig indicates whether the payload used the multi-config format.
	IsMultiConfig bool
}

// parseConfigPayload parses the serialized JS-config payload.
// It supports both the new multi-config format ({ configs: [...] })
// and the legacy single-config format ({ configDirectory, entries }).
func parseConfigPayload(data []byte, filesystems ...vfs.FS) (*parsedPayload, error) {
	var fsys vfs.FS
	if len(filesystems) > 0 {
		fsys = filesystems[0]
	}

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
		configTargetScopes := make(map[string]rslintconfig.LintDiscoveryScope)
		configDirByPathID := make(map[string]string, len(multiPayload.Configs))
		configDirByCanonicalID := make(map[string]string, len(multiPayload.Configs))
		for _, cfg := range multiPayload.Configs {
			// configMap is keyed by the normalized dir (Go matches normalized
			// file paths against it); originalConfigDir recovers the raw string
			// for the eslint-plugin wire configKey. Both are populated in this
			// one loop from the same NormalizePath call, so they cannot disagree.
			configDir := tspath.NormalizePath(cfg.ConfigDirectory)
			if len(configDir) > tspath.GetRootLength(configDir) {
				configDir = tspath.RemoveTrailingDirectorySeparators(configDir)
			}
			pathID := exactFilesystemPathID(configDir)
			if previous, exists := configDirByPathID[pathID]; exists {
				return nil, fmt.Errorf(
					"duplicate config directories %q and %q normalize to the same path %q",
					previous,
					cfg.ConfigDirectory,
					configDir,
				)
			}
			// Config roots are ownership boundaries. Two lexical roots for one
			// physical directory cannot both own the same canonical target without
			// order-dependent config loss, so reject them at ingress. This is one
			// realpath per config, not per target.
			canonicalID := canonicalFilesystemPathID(configDir, fsys)
			if previous, exists := configDirByCanonicalID[canonicalID]; exists {
				return nil, fmt.Errorf(
					"duplicate config directories %q and %q resolve to the same filesystem location",
					previous,
					cfg.ConfigDirectory,
				)
			}
			if err := rslintconfig.ValidateConfig(cfg.Entries); err != nil {
				return nil, fmt.Errorf("invalid config for %q: %w", cfg.ConfigDirectory, err)
			}
			configMap[configDir] = cfg.Entries
			originalConfigDir[configDir] = cfg.ConfigDirectory
			configDirByPathID[pathID] = cfg.ConfigDirectory
			configDirByCanonicalID[canonicalID] = cfg.ConfigDirectory
			if cfg.TargetFiles != nil || cfg.ExplicitOnly {
				files := make([]string, 0, len(cfg.TargetFiles))
				for _, file := range cfg.TargetFiles {
					if tspath.PathIsAbsolute(file) {
						files = append(files, tspath.NormalizePath(file))
					} else {
						files = append(files, tspath.ResolvePath(configDir, file))
					}
				}
				scope := rslintconfig.LintDiscoveryScope{
					Files:        files,
					ExplicitOnly: cfg.ExplicitOnly,
				}
				configTargetScopes[configDir] = scope
			}
		}
		return &parsedPayload{
			ConfigMap:          configMap,
			OriginalConfigDir:  originalConfigDir,
			ConfigTargetScopes: configTargetScopes,
			IsMultiConfig:      true,
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
