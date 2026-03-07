package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInitDefaultConfig_TSProject(t *testing.T) {
	dir := t.TempDir()

	// Create tsconfig.json to trigger TS config generation
	if err := os.WriteFile(filepath.Join(dir, "tsconfig.json"), []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := InitDefaultConfig(dir); err != nil {
		t.Fatalf("InitDefaultConfig failed: %v", err)
	}

	configPath := filepath.Join(dir, "rslint.config.ts")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Expected rslint.config.ts to be created")
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(content) == 0 {
		t.Error("Expected non-empty config file")
	}
}

func TestInitDefaultConfig_JSProject(t *testing.T) {
	dir := t.TempDir()

	// No tsconfig.json, should generate JS config
	if err := InitDefaultConfig(dir); err != nil {
		t.Fatalf("InitDefaultConfig failed: %v", err)
	}

	configPath := filepath.Join(dir, "rslint.config.js")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Expected rslint.config.js to be created")
	}
}

func TestInitDefaultConfig_AlreadyExists(t *testing.T) {
	dir := t.TempDir()

	// Create an existing config
	if err := os.WriteFile(filepath.Join(dir, "rslint.config.ts"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	err := InitDefaultConfig(dir)
	if err == nil {
		t.Error("Expected error when config already exists")
	}
}

func TestInitDefaultConfig_JSONExists(t *testing.T) {
	dir := t.TempDir()

	// Create an existing JSON config
	if err := os.WriteFile(filepath.Join(dir, "rslint.json"), []byte("[]"), 0644); err != nil {
		t.Fatal(err)
	}

	err := InitDefaultConfig(dir)
	if err == nil {
		t.Error("Expected error when rslint.json already exists")
	}
}
