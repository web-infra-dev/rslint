package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
)

// TestParseConfigPayload_MultiConfig_OriginalConfigDirPreservesRaw pins the
// Option-3 invariant at the parse layer: configMap is keyed by the NORMALIZED
// dir (Go matches normalized file paths against it) while OriginalConfigDir
// recovers the RAW string the JS host sent, so the eslint-plugin wire configKey
// round-trips raw (byte-matching the worker's plugin map key) instead of Go's
// normalized form.
func TestParseConfigPayload_MultiConfig_OriginalConfigDirPreservesRaw(t *testing.T) {
	// "C:\\proj" in JSON is the raw string `C:\proj`; "/posix/proj" is POSIX.
	data := []byte(`{"configs":[{"configDirectory":"C:\\proj","entries":[]},{"configDirectory":"/posix/proj","entries":[]}]}`)
	result, err := parseConfigPayload(data)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if !result.IsMultiConfig {
		t.Fatal("expected multi-config")
	}

	const winRaw = `C:\proj`
	winNorm := tspath.NormalizePath(winRaw) // "C:/proj"
	if winNorm == winRaw {
		t.Fatalf("test precondition: NormalizePath should change %q (got identical)", winRaw)
	}
	if _, ok := result.ConfigMap[winNorm]; !ok {
		t.Errorf("configMap missing normalized key %q", winNorm)
	}
	if got := result.OriginalConfigDir[winNorm]; got != winRaw {
		t.Errorf("OriginalConfigDir[%q] = %q, want raw %q", winNorm, got, winRaw)
	}
	// POSIX: raw == normalized, round-trips trivially.
	if got := result.OriginalConfigDir["/posix/proj"]; got != "/posix/proj" {
		t.Errorf("posix OriginalConfigDir = %q, want /posix/proj", got)
	}
}

func TestParseConfigPayload_MultiConfig(t *testing.T) {
	data := []byte(`{
		"configs": [
			{"configDirectory": "/project/packages/foo", "entries": [{"rules": {"no-console": "error"}}]},
			{"configDirectory": "/project/packages/bar", "entries": [{"rules": {"no-debugger": "error"}}]}
		]
	}`)

	result, err := parseConfigPayload(data)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !result.IsMultiConfig {
		t.Error("Expected IsMultiConfig=true")
	}
	if len(result.ConfigMap) != 2 {
		t.Errorf("Expected 2 configs, got %d", len(result.ConfigMap))
	}
	if _, ok := result.ConfigMap["/project/packages/foo"]; !ok {
		t.Error("Expected foo config in map")
	}
	if _, ok := result.ConfigMap["/project/packages/bar"]; !ok {
		t.Error("Expected bar config in map")
	}
}

func TestParseConfigPayload_MultiConfigTargetFiles(t *testing.T) {
	data := []byte(`{
		"configs": [
			{
				"configDirectory": "/project/packages/foo",
				"entries": [{"rules": {"no-console": "error"}}],
				"targetFiles": ["/project/packages/foo/src/a.ts", "src/b.ts"]
			}
		]
	}`)

	result, err := parseConfigPayload(data)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	targets := result.ConfigTargetFiles["/project/packages/foo"]
	if len(targets) != 2 {
		t.Fatalf("Expected 2 target files, got %v", targets)
	}
	if targets[0] != "/project/packages/foo/src/a.ts" {
		t.Errorf("Expected absolute target to be normalized, got %q", targets[0])
	}
	if targets[1] != "/project/packages/foo/src/b.ts" {
		t.Errorf("Expected relative target to resolve from config dir, got %q", targets[1])
	}
}

func TestParseConfigPayload_SingleConfig(t *testing.T) {
	data := []byte(`{
		"configDirectory": "/project/packages/foo",
		"entries": [{"rules": {"no-console": "error"}}]
	}`)

	result, err := parseConfigPayload(data)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result.IsMultiConfig {
		t.Error("Expected IsMultiConfig=false for legacy format")
	}
	if result.ConfigMap != nil {
		t.Error("Expected nil configMap for legacy format")
	}
	if result.SingleConfigDir != "/project/packages/foo" {
		t.Errorf("Expected configDir /project/packages/foo, got %s", result.SingleConfigDir)
	}
	if result.SingleConfig == nil {
		t.Fatal("Expected non-nil singleConfig")
	}
}

func TestParseConfigPayload_RejectsEmptyFilesArray(t *testing.T) {
	cases := []struct {
		name string
		data []byte
	}{
		{
			name: "multi config",
			data: []byte(`{
				"configs": [
					{"configDirectory": "/project", "entries": [{"files": [], "rules": {"no-console": "error"}}]}
				]
			}`),
		},
		{
			name: "legacy single config",
			data: []byte(`{
				"configDirectory": "/project",
				"entries": [{"files": [], "rules": {"no-console": "error"}}]
			}`),
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseConfigPayload(tt.data)
			if err == nil {
				t.Fatal("expected parseConfigPayload to reject empty files array")
			}
			if !strings.Contains(err.Error(), `key "files": expected value to be a non-empty array`) {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestParseConfigPayload_EmptyConfigs(t *testing.T) {
	// Empty configs array → falls back to legacy parsing
	data := []byte(`{"configs": []}`)

	result, err := parseConfigPayload(data)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result.IsMultiConfig {
		t.Error("Expected IsMultiConfig=false for empty configs array")
	}
}

func TestParseConfigPayload_SingleConfigInMultiFormat(t *testing.T) {
	// Single config wrapped in the multi-config format
	data := []byte(`{
		"configs": [
			{"configDirectory": "/project", "entries": [{"rules": {"no-console": "error"}}]}
		]
	}`)

	result, err := parseConfigPayload(data)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !result.IsMultiConfig {
		t.Error("Expected IsMultiConfig=true even for single config in array")
	}
	if len(result.ConfigMap) != 1 {
		t.Errorf("Expected 1 config, got %d", len(result.ConfigMap))
	}
}

func TestParseConfigPayload_InvalidJSON(t *testing.T) {
	data := []byte(`{invalid json}`)

	_, err := parseConfigPayload(data)
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestParseConfigPayload_ConfigDirNormalized(t *testing.T) {
	// Verify paths are normalized
	data := []byte(`{
		"configs": [
			{"configDirectory": "/project/./src/../packages/foo", "entries": []}
		]
	}`)

	result, err := parseConfigPayload(data)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	// tspath.NormalizePath should clean up the path
	if _, ok := result.ConfigMap["/project/packages/foo"]; !ok {
		t.Errorf("Expected normalized path /project/packages/foo in map, got keys: %v", keysOf(result.ConfigMap))
	}
}

func TestParseConfigPayload_DuplicateConfigDirs(t *testing.T) {
	data := []byte(`{
		"configs": [
			{"configDirectory": "/project", "entries": [{"rules": {"rule-a": "error"}}]},
			{"configDirectory": "/project/./", "entries": [{"rules": {"rule-b": "error"}}]}
		]
	}`)

	_, err := parseConfigPayload(data)
	if err == nil || !strings.Contains(err.Error(), "duplicate config directories") {
		t.Fatalf("expected duplicate normalized directories to be rejected, got %v", err)
	}
}

type caseInsensitivePayloadFS struct {
	vfs.FS
}

func (f *caseInsensitivePayloadFS) UseCaseSensitiveFileNames() bool { return false }
func (f *caseInsensitivePayloadFS) Realpath(filePath string) string {
	return tspath.NormalizePath(filePath)
}

func TestParseConfigPayload_RejectsCaseEquivalentConfigDirs(t *testing.T) {
	data := []byte(`{
		"configs": [
			{"configDirectory": "C:/Repo", "entries": [{"rules": {"rule-a": "error"}}]},
			{"configDirectory": "c:/repo", "entries": [{"rules": {"rule-b": "error"}}]}
		]
	}`)

	_, err := parseConfigPayload(data, &caseInsensitivePayloadFS{FS: osvfs.FS()})
	if err == nil || !strings.Contains(err.Error(), "equivalent on this filesystem") {
		t.Fatalf("expected case-equivalent config roots to be rejected, got %v", err)
	}
}

func TestParseConfigPayload_RejectsSymlinkEquivalentConfigDirs(t *testing.T) {
	parent := t.TempDir()
	realDir := filepath.Join(parent, "real")
	aliasDir := filepath.Join(parent, "alias")
	if err := os.Mkdir(realDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(realDir, aliasDir); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	payload, err := json.Marshal(map[string]any{
		"configs": []any{
			map[string]any{"configDirectory": realDir, "entries": []any{}},
			map[string]any{"configDirectory": aliasDir, "entries": []any{}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	_, err = parseConfigPayload(payload, fsys)
	if err == nil || !strings.Contains(err.Error(), "same filesystem location") {
		t.Fatalf("expected symlink-equivalent config roots to be rejected, got %v", err)
	}
}

func TestParseConfigPayload_EmptyObject(t *testing.T) {
	data := []byte(`{}`)

	result, err := parseConfigPayload(data)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	// No configs and no configDirectory → legacy mode with empty values
	if result.IsMultiConfig {
		t.Error("Expected IsMultiConfig=false for empty object")
	}
}

func TestParseConfigPayload_MissingConfigDirectory(t *testing.T) {
	// Config entry without configDirectory → stored with empty string key
	data := []byte(`{
		"configs": [
			{"entries": [{"rules": {"no-console": "error"}}]}
		]
	}`)

	result, err := parseConfigPayload(data)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !result.IsMultiConfig {
		t.Error("Expected IsMultiConfig=true")
	}
	// Empty configDirectory normalizes to empty string
	if len(result.ConfigMap) != 1 {
		t.Errorf("Expected 1 config, got %d", len(result.ConfigMap))
	}
}

func TestParseConfigPayload_NullEntries(t *testing.T) {
	data := []byte(`{
		"configs": [
			{"configDirectory": "/project", "entries": null}
		]
	}`)

	result, err := parseConfigPayload(data)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !result.IsMultiConfig {
		t.Error("Expected IsMultiConfig=true")
	}
	cfg, ok := result.ConfigMap["/project"]
	if !ok {
		t.Fatal("Expected /project in config map")
	}
	if cfg != nil {
		t.Errorf("Expected nil entries for null, got %v", cfg)
	}
}

func TestParseConfigPayload_LegacyEmptyConfigDirectory(t *testing.T) {
	data := []byte(`{
		"configDirectory": "",
		"entries": [{"rules": {"no-console": "error"}}]
	}`)

	result, err := parseConfigPayload(data)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result.IsMultiConfig {
		t.Error("Expected legacy mode")
	}
	if result.SingleConfigDir != "" {
		t.Errorf("Expected empty configDir, got %s", result.SingleConfigDir)
	}
	if result.SingleConfig == nil {
		t.Error("Expected non-nil singleConfig")
	}
}

func keysOf[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
