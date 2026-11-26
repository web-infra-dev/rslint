package test_framework

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
)

func TestResolveModuleImportSpecifierTypeOnlyPolicy(t *testing.T) {
	// The `type` modifier is recorded on the specifier when it is written
	// inline and on the enclosing clause when it heads the declaration. Both
	// name a binding that does not exist at runtime.
	for _, code := range []string{
		`import { type Mock } from "@rstest/core";`,
		`import type { Mock } from "@rstest/core";`,
	} {
		t.Run(code, func(t *testing.T) {
			sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
				FileName: "/test.ts",
				Path:     "/test.ts",
			}, code, core.ScriptKindTS)
			specifier := findFirstNodeOfKind(sourceFile, ast.KindImportSpecifier)
			if specifier == nil {
				t.Fatal("expected an import specifier")
			}

			modules := []string{"@rstest/core"}
			if _, _, ok := resolveModuleImportSpecifier(specifier, modules, false); ok {
				t.Fatal("function reference resolution accepted a type-only specifier")
			}
			name, _, ok := resolveModuleImportSpecifier(specifier, modules, true)
			if !ok || name != "Mock" {
				t.Fatalf("type reference resolution = %q, %v; want Mock, true", name, ok)
			}
		})
	}
}

func TestResolveFunctionReferenceForModules(t *testing.T) {
	modules := []string{"@jest/globals"}

	if name, node, _ := ResolveFunctionReferenceForModules(nil, "test", nil, nil, nil, modules); name != "" || node != nil {
		t.Fatalf("nil node = (%q, %v), want (\"\", nil)", name, node)
	}

	ident := parseFirstNodeOfKind(t, `test;`, ast.KindIdentifier)
	if name, node, _ := ResolveFunctionReferenceForModules(ident, "test", ident, nil, nil, modules); name != "" || node != nil {
		t.Fatalf("non-call node = (%q, %v), want (\"\", nil)", name, node)
	}

	_, regexCall := parseFirstCall(t, `/a/.test(text);`)
	if name, node, _ := ResolveFunctionReferenceForModules(regexCall, "test", nil, nil, nil, modules); name != "" || node != nil {
		t.Fatalf("non-identifier receiver call = (%q, %v), want (\"\", nil)", name, node)
	}

	_, testCall := parseFirstCall(t, `test("case", () => {});`)
	name, node, mode := ResolveFunctionReferenceForModules(testCall, "test", nil, nil, nil, modules)
	if name != "test" || node == nil || node.Kind != ast.KindIdentifier || node.AsIdentifier().Text != "test" || mode != ReferenceModeGlobal {
		t.Fatalf("identifier call without checker = (%q, %v, %v), want (\"test\", identifier, global)", name, node, mode)
	}
}

func TestIsImportEqualsOfModules(t *testing.T) {
	decl := parseFirstNodeOfKind(t, `import r = require("@rstest/core");`, ast.KindImportEqualsDeclaration)
	if !IsImportEqualsOfModules(decl, []string{"@rstest/core"}) {
		t.Fatal("expected matching import-equals declaration to return true")
	}
	if IsImportEqualsOfModules(decl, []string{"vitest"}) {
		t.Fatal("expected non-matching import-equals declaration to return false")
	}
	if IsImportEqualsOfModules(nil, []string{"@rstest/core"}) {
		t.Fatal("expected nil declaration to return false")
	}
}
