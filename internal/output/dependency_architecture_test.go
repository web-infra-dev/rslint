package output

import (
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// TestOutputDependencyBoundary protects output as a presentation-only module.
// Lint-domain and compiler-owned values must be projected before they reach it.
func TestOutputDependencyBoundary(t *testing.T) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("locate output architecture test")
	}
	outputDirectory := filepath.Dir(currentFile)
	repositoryRoot := filepath.Clean(filepath.Join(outputDirectory, "..", ".."))

	var violations []string
	err := filepath.WalkDir(outputDirectory, func(filePath string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			return nil
		}

		fileSet := token.NewFileSet()
		file, err := parser.ParseFile(fileSet, filePath, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		relativePath, err := filepath.Rel(repositoryRoot, filePath)
		if err != nil {
			return err
		}
		relativePath = filepath.ToSlash(relativePath)

		for _, imported := range file.Imports {
			importPath, err := strconv.Unquote(imported.Path.Value)
			if err != nil {
				return err
			}
			if prohibitedOutputDependency(importPath) {
				violations = append(violations, relativePath+": imports lint-domain dependency "+importPath)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("scan internal/output: %v", err)
	}

	if len(violations) > 0 {
		sort.Strings(violations)
		t.Fatalf("output architecture boundary violations:\n%s", strings.Join(violations, "\n"))
	}
}

func prohibitedOutputDependency(importPath string) bool {
	return importPath == "github.com/web-infra-dev/rslint/internal/rule" ||
		importPath == "github.com/web-infra-dev/rslint/internal/linter" ||
		strings.HasPrefix(importPath, "github.com/web-infra-dev/rslint/internal/config") ||
		strings.HasPrefix(importPath, "github.com/web-infra-dev/rslint/internal/program") ||
		strings.HasPrefix(importPath, "github.com/microsoft/TypeScript/tsc/shim/ast") ||
		strings.HasPrefix(importPath, "github.com/microsoft/TypeScript/tsc/shim/core") ||
		strings.HasPrefix(importPath, "github.com/microsoft/TypeScript/tsc/shim/scanner") ||
		strings.HasPrefix(importPath, "github.com/microsoft/TypeScript/tsc/shim/vfs")
}
