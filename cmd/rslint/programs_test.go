package main

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
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
		Run:              func(ctx rule.RuleContext, options []any) rule.RuleListeners { return nil },
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

	fileSet := utils.CollectProgramFiles([]*compiler.Program{program}, bundled.WrapFS(cachedvfs.From(osvfs.FS())), false)

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

	fileSet := utils.CollectProgramFiles([]*compiler.Program{program}, bundled.WrapFS(cachedvfs.From(osvfs.FS())), false)

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

type realpathCountingFS struct {
	vfs.FS

	mu      sync.Mutex
	aliases map[string]string
	calls   map[string]int
}

func (fsys *realpathCountingFS) Realpath(path string) string {
	fsys.mu.Lock()
	fsys.calls[path]++
	alias, ok := fsys.aliases[path]
	fsys.mu.Unlock()
	if ok {
		return alias
	}
	return fsys.FS.Realpath(path)
}

func (fsys *realpathCountingFS) callCount(path string) int {
	fsys.mu.Lock()
	defer fsys.mu.Unlock()
	return fsys.calls[path]
}

func TestBuildProgramFileIndex_DedupesRealpathAcrossPrograms(t *testing.T) {
	program := createTestProgram(t, map[string]string{
		"a.ts": "const a = 1;",
	})

	var target string
	for _, sf := range program.GetSourceFiles() {
		if strings.HasSuffix(sf.FileName(), "/a.ts") {
			target = sf.FileName()
			break
		}
	}
	if target == "" {
		t.Fatal("expected program to include a.ts")
	}

	realTarget := target + ".real"
	fsys := &realpathCountingFS{
		FS:      bundled.WrapFS(cachedvfs.From(osvfs.FS())),
		aliases: map[string]string{target: realTarget},
		calls:   make(map[string]int),
	}

	index := buildProgramFileIndex([]*compiler.Program{program, program}, fsys, true)

	if got := fsys.callCount(target); got != 1 {
		t.Fatalf("expected one realpath lookup for duplicate source path, got %d", got)
	}
	if _, ok := index.files[target]; !ok {
		t.Fatalf("expected original path %q in program index", target)
	}
	if _, ok := index.files[realTarget]; !ok {
		t.Fatalf("expected realpath alias %q in program index", realTarget)
	}
	candidates := index.byPath[realTarget]
	if len(candidates) != 2 {
		t.Fatalf("expected both program candidates for realpath alias, got %d", len(candidates))
	}
	if candidates[0].programIndex != 0 || candidates[1].programIndex != 1 {
		t.Fatalf("expected candidates to preserve program order, got %#v", candidates)
	}
}
