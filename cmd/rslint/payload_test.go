package main

import (
	"testing"
)

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
	// Same configDirectory twice → last one wins (map semantics)
	data := []byte(`{
		"configs": [
			{"configDirectory": "/project", "entries": [{"rules": {"rule-a": "error"}}]},
			{"configDirectory": "/project", "entries": [{"rules": {"rule-b": "error"}}]}
		]
	}`)

	result, err := parseConfigPayload(data)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(result.ConfigMap) != 1 {
		t.Errorf("Expected 1 config (deduped), got %d", len(result.ConfigMap))
	}
	// Last entry should win
	cfg := result.ConfigMap["/project"]
	if cfg == nil {
		t.Fatal("Expected config for /project")
	}
	if _, ok := cfg[0].Rules["rule-b"]; !ok {
		t.Error("Expected last entry (rule-b) to win")
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
