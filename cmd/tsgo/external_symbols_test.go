package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
)

func TestCollectExternalSymbolsGlobals(t *testing.T) {
	fixture := buildSemanticFixture(t, "export const value = Math.abs(-1);")

	for _, expected := range []string{
		"globalThis",
		"Math",
		"Math.abs",
		"Object.prototype.hasOwnProperty",
		"console.log",
	} {
		if !hasExternalSymbol(fixture.semantic, globalNamespace, expected) {
			t.Errorf("missing global external symbol %q", expected)
		}
	}
}

func TestCollectExternalSymbolsAmbientDependencyGlobalsWithoutDefaultLib(t *testing.T) {
	tmpDir := t.TempDir()
	dependencyDir := filepath.Join(tmpDir, "node_modules", "@types", "example")
	if err := os.MkdirAll(dependencyDir, 0o755); err != nil {
		t.Fatalf("Failed to create dependency directory: %v", err)
	}

	files := map[string]string{
		filepath.Join(tmpDir, "tsconfig.json"): `{
			"compilerOptions": {"noLib": true, "types": ["example"]},
			"include": ["./index.ts"]
		}`,
		filepath.Join(tmpDir, "index.ts"): `ambientDependencyGlobal();`,
		filepath.Join(dependencyDir, "index.d.ts"): `
			declare function ambientDependencyGlobal(): void;
		`,
	}
	for path, contents := range files {
		if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
			t.Fatalf("Failed to write %s: %v", path, err)
		}
	}

	t.Chdir(tmpDir)
	program, err := CreateProgram("tsconfig.json")
	if err != nil {
		t.Fatalf("Failed to create program: %v", err)
	}
	semantic := CollectSemantic(program)

	if !hasExternalSymbol(semantic, globalNamespace, "ambientDependencyGlobal") {
		t.Error("missing ambient dependency global external symbol")
	}
}

func TestCollectExternalSymbolsDoesNotMarkLocalShadowAsGlobal(t *testing.T) {
	fixture := buildSemanticFixture(
		t,
		`export {}; const Math = { abs(value: number) { return value; } }; Math.abs(-1);`,
	)

	found := map[string]bool{"Math": false, "abs": false}
	for symbolID, symbol := range fixture.semantic.Symtab {
		name := string(symbol.Name)
		if _, ok := found[name]; !ok || symbol.Decl == nil || symbol.Decl.SourceFileId != fixture.sourceFileID {
			continue
		}
		found[name] = true
		if external, ok := findExternalSymbolByID(fixture.semantic, symbolID); ok {
			t.Errorf(
				"local symbol %q was incorrectly marked as %q.%q",
				name,
				external.Namespace,
				external.Name,
			)
		}
	}
	for name, wasFound := range found {
		if !wasFound {
			t.Errorf("local symbol %q not found", name)
		}
	}
}

func TestCollectExternalSymbolsDependencyExports(t *testing.T) {
	tmpDir := t.TempDir()
	dependencyDir := filepath.Join(tmpDir, "node_modules", "example-dependency")
	if err := os.MkdirAll(dependencyDir, 0o755); err != nil {
		t.Fatalf("Failed to create dependency directory: %v", err)
	}

	files := map[string]string{
		filepath.Join(tmpDir, "tsconfig.json"): `{
			"compilerOptions": {"moduleResolution": "node"},
			"include": ["./index.ts"]
		}`,
		filepath.Join(tmpDir, "index.ts"): `import { api } from "example-dependency"; api.run();`,
		filepath.Join(dependencyDir, "package.json"): `{
			"name": "example-dependency",
			"version": "1.0.0",
			"types": "index.d.ts"
		}`,
		filepath.Join(dependencyDir, "index.d.ts"): `
			export declare const api: { run(): void };
			export declare function value(): void;
		`,
	}
	for path, contents := range files {
		if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
			t.Fatalf("Failed to write %s: %v", path, err)
		}
	}

	t.Chdir(tmpDir)
	program, err := CreateProgram("tsconfig.json")
	if err != nil {
		t.Fatalf("Failed to create program: %v", err)
	}
	semantic := CollectSemantic(program)

	for _, expected := range []string{"api", "api.run", "value"} {
		if !hasExternalSymbol(semantic, "example-dependency", expected) {
			t.Errorf("missing dependency external symbol %q", expected)
		}
	}
}

func TestExternalModuleName(t *testing.T) {
	tests := []struct {
		moduleName string
		namespace  string
		prefix     string
		ok         bool
	}{
		{moduleName: "react", namespace: "react", ok: true},
		{moduleName: "react/jsx-runtime", namespace: "react", prefix: "jsx-runtime", ok: true},
		{moduleName: "@scope/pkg", namespace: "@scope/pkg", ok: true},
		{moduleName: "@scope/pkg/subpath", namespace: "@scope/pkg", prefix: "subpath", ok: true},
		{moduleName: "node:assert", namespace: "node:assert", ok: true},
		{moduleName: "./local", ok: false},
	}

	for _, test := range tests {
		t.Run(test.moduleName, func(t *testing.T) {
			namespace, prefix, ok := externalModuleName(test.moduleName)
			if namespace != test.namespace || prefix != test.prefix || ok != test.ok {
				t.Fatalf(
					"externalModuleName(%q) = (%q, %q, %v), want (%q, %q, %v)",
					test.moduleName,
					namespace,
					prefix,
					ok,
					test.namespace,
					test.prefix,
					test.ok,
				)
			}
		})
	}
}

func hasExternalSymbol(semantic Semantic, namespace string, name string) bool {
	for _, external := range semantic.ExternalSymbols {
		if string(external.Namespace) == namespace && string(external.Name) == name {
			return true
		}
	}
	return false
}

func findExternalSymbolByID(semantic Semantic, symbolID ast.SymbolId) (ExternalSymbol, bool) {
	for _, external := range semantic.ExternalSymbols {
		if external.SymbolId == symbolID {
			return external, true
		}
	}
	return ExternalSymbol{}, false
}
