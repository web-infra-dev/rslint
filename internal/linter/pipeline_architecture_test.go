package linter

import (
	"go/ast"
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

const linterImportPath = "github.com/web-infra-dev/rslint/internal/linter"

// TestProductLintBoundary protects module responsibilities rather than a
// particular call shape. Product integrations may construct requests, generation
// adapters, transports, and result projections; only internal/linter may
// coordinate planning, execution, plugin scheduling, and fix application.
func TestProductLintBoundary(t *testing.T) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("locate architecture test")
	}
	repositoryRoot := filepath.Clean(filepath.Join(filepath.Dir(currentFile), "..", ".."))

	var violations []string
	for _, scanRoot := range []string{"cmd", "internal"} {
		absoluteDirectory := filepath.Join(repositoryRoot, scanRoot)
		err := filepath.WalkDir(absoluteDirectory, func(filePath string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
				return nil
			}
			relativePath, err := filepath.Rel(repositoryRoot, filePath)
			if err != nil {
				return err
			}
			relativePath = filepath.ToSlash(relativePath)
			packagePath := filepath.ToSlash(filepath.Dir(relativePath))

			fileSet := token.NewFileSet()
			file, err := parser.ParseFile(fileSet, filePath, nil, 0)
			if err != nil {
				return err
			}
			linterAliases := make(map[string]struct{})
			for _, imported := range file.Imports {
				importPath, err := strconv.Unquote(imported.Path.Value)
				if err != nil {
					return err
				}
				if importPath == "github.com/web-infra-dev/rslint/internal/lintpipeline" {
					violations = append(violations, relativePath+": imports retired internal/lintpipeline")
				}
				if packagePath == "internal/linter" && prohibitedPipelineDependency(importPath) {
					violations = append(violations, relativePath+": linter imports integration or persistence dependency "+importPath)
				}
				if importPath != linterImportPath {
					continue
				}
				alias := "linter"
				if imported.Name != nil {
					alias = imported.Name.Name
				}
				if alias == "." {
					violations = append(violations, relativePath+": dot-imports internal/linter")
				} else if alias != "_" {
					linterAliases[alias] = struct{}{}
				}
			}

			if !isProductLintIntegration(packagePath) {
				return nil
			}
			ast.Inspect(file, func(node ast.Node) bool {
				selector, ok := node.(*ast.SelectorExpr)
				if !ok {
					return true
				}
				identifier, ok := selector.X.(*ast.Ident)
				if !ok {
					return true
				}
				if _, imported := linterAliases[identifier.Name]; !imported || !rawPipelineStage(selector.Sel.Name) {
					return true
				}
				position := fileSet.Position(selector.Pos())
				violations = append(violations, relativePath+":"+strconv.Itoa(position.Line)+
					": product integration references raw linter stage "+selector.Sel.Name)
				return true
			})
			return nil
		})
		if err != nil {
			t.Fatalf("scan %s: %v", scanRoot, err)
		}
	}

	if len(violations) > 0 {
		sort.Strings(violations)
		t.Fatalf("lint architecture boundary violations:\n%s", strings.Join(violations, "\n"))
	}
}

func prohibitedPipelineDependency(importPath string) bool {
	return importPath == "os" ||
		importPath == "io/fs" ||
		importPath == "log" ||
		importPath == "github.com/web-infra-dev/rslint/internal/output" ||
		importPath == "github.com/web-infra-dev/rslint/internal/program/loader" ||
		strings.HasPrefix(importPath, "github.com/web-infra-dev/rslint/internal/api") ||
		strings.HasPrefix(importPath, "github.com/web-infra-dev/rslint/internal/config") ||
		strings.HasPrefix(importPath, "github.com/web-infra-dev/rslint/internal/lsp") ||
		strings.HasPrefix(importPath, "github.com/microsoft/typescript-go/shim/vfs")
}

func isProductLintIntegration(packagePath string) bool {
	return packagePath == "cmd/rslint" ||
		packagePath == "internal/api/server" ||
		packagePath == "internal/lsp"
}

func rawPipelineStage(symbol string) bool {
	return strings.HasPrefix(symbol, "RunLinter") ||
		strings.HasPrefix(symbol, "PrepareLintPlan") ||
		symbol == "ApplyRuleFixes" ||
		strings.HasPrefix(symbol, "BuildEslintPluginFileInput") ||
		strings.HasPrefix(symbol, "DispatchEslintPluginRules") ||
		symbol == "LintSingleFile" ||
		strings.HasPrefix(symbol, "CollectFileSyntacticDiagnostics") ||
		strings.HasPrefix(symbol, "CollectTargetSyntacticDiagnostics")
}
