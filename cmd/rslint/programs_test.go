package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestFilterNonTypeAwareRules(t *testing.T) {
	rules := []linter.ConfiguredRule{
		{Name: "syntax-rule", RequiresTypeInfo: false},
		{Name: "type-rule", RequiresTypeInfo: true},
		{Name: "another-syntax", RequiresTypeInfo: false},
	}

	filtered := linter.FilterNonTypeAwareRules(rules)

	if len(filtered) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(filtered))
	}
	if filtered[0].Name != "syntax-rule" {
		t.Errorf("expected syntax-rule, got %s", filtered[0].Name)
	}
	if filtered[1].Name != "another-syntax" {
		t.Errorf("expected another-syntax, got %s", filtered[1].Name)
	}
}

func TestFilterNonTypeAwareRules_AllTypeAware(t *testing.T) {
	rules := []linter.ConfiguredRule{
		{Name: "type-rule-1", RequiresTypeInfo: true},
		{Name: "type-rule-2", RequiresTypeInfo: true},
	}

	filtered := linter.FilterNonTypeAwareRules(rules)

	if len(filtered) != 0 {
		t.Fatalf("expected 0 rules, got %d", len(filtered))
	}
}

func TestFilterNonTypeAwareRules_NoneTypeAware(t *testing.T) {
	rules := []linter.ConfiguredRule{
		{Name: "rule-a", RequiresTypeInfo: false},
		{Name: "rule-b", RequiresTypeInfo: false},
	}

	filtered := linter.FilterNonTypeAwareRules(rules)

	if len(filtered) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(filtered))
	}
}

func TestFilterNonTypeAwareRules_Empty(t *testing.T) {
	filtered := linter.FilterNonTypeAwareRules(nil)
	if len(filtered) != 0 {
		t.Fatalf("expected 0 rules, got %d", len(filtered))
	}
}

// Verify RequiresTypeInfo propagates through CreateRule.
func TestRequiresTypeInfo_Propagation(t *testing.T) {
	r := rule.CreateRule(rule.Rule{
		Name:             "test-rule",
		RequiresTypeInfo: true,
		Run:              func(ctx rule.RuleContext, options any) rule.RuleListeners { return nil },
	})

	if r.Name != "@typescript-eslint/test-rule" {
		t.Errorf("unexpected name: %s", r.Name)
	}
	if !r.RequiresTypeInfo {
		t.Error("RequiresTypeInfo should be true after CreateRule")
	}
}

// createTestProgram creates a Program from temp files for testing utils.CollectProgramFiles.
func createTestProgram(t *testing.T, files map[string]string) *compiler.Program {
	t.Helper()
	tmpDir := t.TempDir()

	includes := make([]string, 0, len(files))
	for name, content := range files {
		fp := filepath.Join(tmpDir, name)
		os.MkdirAll(filepath.Dir(fp), 0755)
		os.WriteFile(fp, []byte(content), 0644)
		includes = append(includes, "./"+name)
	}

	tsconfig := `{"include":["` + includes[0] + `"]}`
	if len(includes) > 1 {
		tsconfig = `{"include":["**/*.ts"]}`
	}
	os.WriteFile(filepath.Join(tmpDir, "tsconfig.json"), []byte(tsconfig), 0644)

	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	host := utils.CreateCompilerHost(tmpDir, fs)
	program, err := utils.CreateProgram(true, fs, tmpDir, "tsconfig.json", host)
	if err != nil {
		t.Fatalf("Failed to create program: %v", err)
	}
	return program
}

func TestBuildProgramFileSet_ContainsSourceFiles(t *testing.T) {
	program := createTestProgram(t, map[string]string{
		"a.ts": "const a = 1;",
		"b.ts": "const b = 2;",
	})

	fileSet := utils.CollectProgramFiles([]*compiler.Program{program}, bundled.WrapFS(cachedvfs.From(osvfs.FS())))

	// Should contain user source files
	found := 0
	for k := range fileSet {
		if filepath.Ext(k) == ".ts" && !strings.Contains(k, "bundled:") {
			found++
		}
	}
	if found < 2 {
		t.Errorf("Expected at least 2 .ts files in fileSet, got %d", found)
	}
}

func TestBuildProgramFileSet_RealpathKeyForSymlinks(t *testing.T) {
	// On macOS, /tmp is a symlink to /private/tmp.
	// Verify that utils.CollectProgramFiles adds the resolved path as an alternate key.
	tmpDir := t.TempDir()
	realTmpDir, err := filepath.EvalSymlinks(tmpDir)
	if err != nil {
		t.Skip("EvalSymlinks not available")
	}
	if tspath.NormalizePath(realTmpDir) == tspath.NormalizePath(tmpDir) {
		t.Skip("No symlink divergence on this system")
	}

	// Create a file and program using the UNRESOLVED path
	os.WriteFile(filepath.Join(tmpDir, "test.ts"), []byte("const x = 1;"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "tsconfig.json"), []byte(`{"include":["**/*.ts"]}`), 0644)

	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	host := utils.CreateCompilerHost(tmpDir, fs)
	program, err := utils.CreateProgram(true, fs, tmpDir, "tsconfig.json", host)
	if err != nil {
		t.Fatalf("Failed to create program: %v", err)
	}

	fileSet := utils.CollectProgramFiles([]*compiler.Program{program}, bundled.WrapFS(cachedvfs.From(osvfs.FS())))

	// The file should be findable via BOTH the original and resolved paths
	resolvedFilePath := tspath.NormalizePath(filepath.Join(realTmpDir, "test.ts"))
	if _, exists := fileSet[resolvedFilePath]; !exists {
		// Dump keys for debugging
		for k := range fileSet {
			if !strings.Contains(k, "bundled:") {
				t.Logf("fileSet key: %s", k)
			}
		}
		t.Errorf("Expected fileSet to contain resolved path %s", resolvedFilePath)
	}
}
