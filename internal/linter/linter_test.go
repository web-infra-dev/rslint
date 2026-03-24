package linter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// noopRule returns a rule that reports on every identifier (for testing file filtering).
func noopRule() []ConfiguredRule {
	return []ConfiguredRule{
		{
			Name:     "test-rule",
			Severity: rule.SeverityWarning,
			Run: func(ctx rule.RuleContext) rule.RuleListeners {
				return rule.RuleListeners{
					ast.KindIdentifier: func(node *ast.Node) {
						ctx.ReportNode(node, rule.RuleMessage{Id: "test", Description: "test"})
					},
				}
			},
		},
	}
}

// createTestProgramWithFiles creates a TS program in a temp directory with the given files.
// Returns the program and a map of short filename -> normalized absolute path.
func createTestProgramWithFiles(t *testing.T, sourceFiles map[string]string) (*compiler.Program, map[string]string) {
	t.Helper()

	tmpDir := t.TempDir()

	includes := make([]string, 0, len(sourceFiles))
	normalizedPaths := make(map[string]string, len(sourceFiles))
	for name, content := range sourceFiles {
		filePath := filepath.Join(tmpDir, name)
		dir := filepath.Dir(filePath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write %s: %v", name, err)
		}
		includes = append(includes, "./"+name)
		normalizedPaths[name] = tspath.NormalizePath(filePath)
	}

	includeJSON := `"` + strings.Join(includes, `","`) + `"`
	tsconfig := `{"include":[` + includeJSON + `]}`
	if err := os.WriteFile(filepath.Join(tmpDir, "tsconfig.json"), []byte(tsconfig), 0644); err != nil {
		t.Fatalf("Failed to write tsconfig: %v", err)
	}

	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	host := utils.CreateCompilerHost(tmpDir, fs)
	program, err := utils.CreateProgram(true, fs, tmpDir, "tsconfig.json", host)
	if err != nil {
		t.Fatalf("Failed to create program: %v", err)
	}

	return program, normalizedPaths
}

// tmpDirPath returns the normalized directory path for a file in normalizedPaths.
func tmpDirPath(t *testing.T, normalizedPaths map[string]string, fileName string) string {
	t.Helper()
	return tspath.GetDirectoryPath(normalizedPaths[fileName])
}

