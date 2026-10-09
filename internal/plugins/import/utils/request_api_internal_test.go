package utils_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

func TestImportRulesUseModuleRequestAPIs(t *testing.T) {
	t.Parallel()

	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate request API architecture test")
	}
	rulesDir := filepath.Join(filepath.Dir(thisFile), "..", "rules")
	forbiddenProgramCalls := map[string]bool{
		"GetResolvedModule":                    true,
		"GetResolvedModuleFromModuleSpecifier": true,
		"ResolveModuleName":                    true,
	}

	err := filepath.WalkDir(rulesDir, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		fileSet := token.NewFileSet()
		file, err := parser.ParseFile(fileSet, path, nil, 0)
		if err != nil {
			return err
		}
		moduleAliases := make(map[string]bool)
		for _, imported := range file.Imports {
			importPath, err := strconv.Unquote(imported.Path.Value)
			if err != nil || importPath != "github.com/web-infra-dev/rslint/internal/utils/modules" {
				continue
			}
			name := "modules"
			if imported.Name != nil {
				name = imported.Name.Name
			}
			moduleAliases[name] = true
		}

		ast.Inspect(file, func(node ast.Node) bool {
			switch node := node.(type) {
			case *ast.CallExpr:
				selector, ok := node.Fun.(*ast.SelectorExpr)
				if ok && forbiddenProgramCalls[selector.Sel.Name] {
					position := fileSet.Position(selector.Sel.Pos())
					t.Errorf("%s: import rules must resolve modules through modules.Source-aware APIs, not %s", position, selector.Sel.Name)
				}
			case *ast.CompositeLit:
				selector, ok := node.Type.(*ast.SelectorExpr)
				if !ok {
					break
				}
				qualifier, qualified := selector.X.(*ast.Ident)
				if qualified && selector.Sel.Name == "Source" && moduleAliases[qualifier.Name] {
					position := fileSet.Position(selector.Sel.Pos())
					t.Errorf("%s: construct module requests through modules collectors or factories", position)
				}
			}
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
