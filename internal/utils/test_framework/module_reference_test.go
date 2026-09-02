package test_framework

import (
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/parser"
)

func TestResolveModuleImportSpecifierTypeOnlyPolicy(t *testing.T) {
	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/test.ts",
		Path:     "/test.ts",
	}, `import { type Mock } from "@rstest/core";`, core.ScriptKindTS)
	specifier := findFirstNodeOfKind(sourceFile, ast.KindImportSpecifier)
	if specifier == nil {
		t.Fatal("expected an import specifier")
	}

	modules := []string{"@rstest/core"}
	if _, _, ok := resolveModuleImportSpecifier(specifier, modules, false); ok {
		t.Fatal("function reference resolution accepted an inline type-only specifier")
	}
	name, _, ok := resolveModuleImportSpecifier(specifier, modules, true)
	if !ok || name != "Mock" {
		t.Fatalf("type reference resolution = %q, %v; want Mock, true", name, ok)
	}
}
