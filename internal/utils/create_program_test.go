package utils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
)

func TestCreateProgramWithOverrides(t *testing.T) {
	// Create a temporary directory for our test
	tempDir, err := os.MkdirTemp("", "rslint-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a basic tsconfig.json
	tsconfigContent := `{
  "compilerOptions": {
    "target": "ES2015",
    "strict": false,
    "noImplicitAny": false
  }
}`
	tsconfigPath := filepath.Join(tempDir, "tsconfig.json")
	if err := os.WriteFile(tsconfigPath, []byte(tsconfigContent), 0644); err != nil {
		t.Fatalf("Failed to write tsconfig.json: %v", err)
	}

	// Create a simple TypeScript file
	tsFileContent := `const x: any = 42;`
	tsFilePath := filepath.Join(tempDir, "test.ts")
	if err := os.WriteFile(tsFilePath, []byte(tsFileContent), 0644); err != nil {
		t.Fatalf("Failed to write test.ts: %v", err)
	}

	// Create compiler host with proper bundled FS for TypeScript libs
	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	host := CreateCompilerHost(tempDir, fs)

	// Test 1: Create program without overrides
	program1, err := CreateProgram(false, fs, tempDir, tsconfigPath, host)
	if err != nil {
		t.Fatalf("Failed to create program without overrides: %v", err)
	}

	// Verify original config settings
	options1 := program1.Options()
	if options1.Target != core.ScriptTargetES2015 {
		t.Errorf("Expected target ES2015, got %v", options1.Target)
	}
	if options1.Strict != core.TSFalse {
		t.Errorf("Expected strict false, got %v", options1.Strict)
	}

	// Test 2: Create program with overrides
	overrides := map[string]interface{}{
		"target":       float64(core.ScriptTargetES2020), // TypeScript expects numeric enum values
		"strict":       true,
		"noImplicitAny": true,
	}

	program2, err := CreateProgramWithOverrides(false, fs, tempDir, tsconfigPath, host, overrides)
	if err != nil {
		t.Fatalf("Failed to create program with overrides: %v", err)
	}

	// Verify overrides applied
	options2 := program2.Options()
	if options2.Target != core.ScriptTargetES2020 {
		t.Errorf("Expected target ES2020 after override, got %v", options2.Target)
	}
	if options2.Strict != core.TSTrue {
		t.Errorf("Expected strict true after override, got %v", options2.Strict)
	}
	if options2.NoImplicitAny != core.TSTrue {
		t.Errorf("Expected noImplicitAny true after override, got %v", options2.NoImplicitAny)
	}
}
