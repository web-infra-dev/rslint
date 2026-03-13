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

func TestInitDefaultConfig_JSProject_NoTypeModule(t *testing.T) {
	dir := t.TempDir()

	// No tsconfig.json, no package.json → should generate .mjs (ESM always works)
	if err := InitDefaultConfig(dir); err != nil {
		t.Fatalf("InitDefaultConfig failed: %v", err)
	}

	configPath := filepath.Join(dir, "rslint.config.mjs")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Expected rslint.config.mjs to be created")
	}
	// .js should NOT exist
	if _, err := os.Stat(filepath.Join(dir, "rslint.config.js")); err == nil {
		t.Error("rslint.config.js should not be created for non-ESM project")
	}
}

func TestInitDefaultConfig_JSProject_TypeModule(t *testing.T) {
	dir := t.TempDir()

	// package.json with "type": "module" → should generate .js
	pkgJSON := []byte(`{"type": "module"}`)
	if err := os.WriteFile(filepath.Join(dir, "package.json"), pkgJSON, 0644); err != nil {
		t.Fatal(err)
	}

	if err := InitDefaultConfig(dir); err != nil {
		t.Fatalf("InitDefaultConfig failed: %v", err)
	}

	configPath := filepath.Join(dir, "rslint.config.js")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Expected rslint.config.js to be created for ESM project")
	}
	// .mjs should NOT exist
	if _, err := os.Stat(filepath.Join(dir, "rslint.config.mjs")); err == nil {
		t.Error("rslint.config.mjs should not be created for ESM project")
	}
}

func TestInitDefaultConfig_JSProject_CJSPackage(t *testing.T) {
	dir := t.TempDir()

	// package.json without "type": "module" → should generate .mjs
	pkgJSON := []byte(`{"name": "my-project"}`)
	if err := os.WriteFile(filepath.Join(dir, "package.json"), pkgJSON, 0644); err != nil {
		t.Fatal(err)
	}

	if err := InitDefaultConfig(dir); err != nil {
		t.Fatalf("InitDefaultConfig failed: %v", err)
	}

	configPath := filepath.Join(dir, "rslint.config.mjs")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Expected rslint.config.mjs to be created for CJS project")
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
