package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGlobalSymbols(t *testing.T) {
	fixture := buildSemanticFixture(t, "export const value = Math.abs(-1);")
	expectSymbol(t, fixture.semantic.ExternalSymbols, globalNamespace, "globalThis")
	expectSymbol(t, fixture.semantic.ExternalSymbols, globalNamespace, "Math")
	expectSymbol(t, fixture.semantic.ExternalSymbols, globalNamespace, "Math.abs")
	expectSymbol(t, fixture.semantic.ExternalSymbols, globalNamespace, "Object.prototype.hasOwnProperty")
	expectSymbol(t, fixture.semantic.ExternalSymbols, globalNamespace, "console.log")
}

func TestJSXIntrinsicElementSymbols(t *testing.T) {
	tmpDir := t.TempDir()
	dependencyDir := filepath.Join(tmpDir, "node_modules", "react")
	_ = os.MkdirAll(dependencyDir, 0o755)

	files := map[string]string{
		filepath.Join(tmpDir, "tsconfig.json"):       "{\n\t\"compilerOptions\": {\"moduleResolution\": \"node\"},\n\t\"include\": [\"./index.ts\"]\n}",
		filepath.Join(tmpDir, "index.ts"):            `import "react";`,
		filepath.Join(dependencyDir, "package.json"): "{\n\t\"name\": \"react\",\n\t\"version\": \"1.0.0\",\n\t\"types\": \"index.d.ts\"\n}",
		filepath.Join(dependencyDir, "index.d.ts"):   "export namespace JSX {\ninterface IntrinsicElements {\ndiv: unknown;\n}\n}",
	}
	for path, contents := range files {
		_ = os.WriteFile(path, []byte(contents), 0o644)
	}

	t.Chdir(tmpDir)
	program, _ := CreateProgram("tsconfig.json")
	semantic := CollectSemantic(program)

	expectSymbol(t, semantic.ExternalSymbols, "react", "JSX.IntrinsicElements.div")
}

func TestDependencySymbols(t *testing.T) {
	tmpDir := t.TempDir()
	dependencyDir := filepath.Join(tmpDir, "node_modules", "example-dependency")
	_ = os.MkdirAll(dependencyDir, 0o755)

	files := map[string]string{
		filepath.Join(tmpDir, "tsconfig.json"):       "{\n\t\"compilerOptions\": {\"moduleResolution\": \"node\"},\n\t\"include\": [\"./index.ts\"]\n}",
		filepath.Join(tmpDir, "index.ts"):            `import { api } from "example-dependency"; api.run();`,
		filepath.Join(dependencyDir, "package.json"): "{\n\t\"name\": \"example-dependency\",\n\t\"version\": \"1.0.0\",\n\t\"types\": \"index.d.ts\"\n}",
		filepath.Join(dependencyDir, "index.d.ts"):   "export declare const api: { run(): void };\nexport declare function value(): void;",
	}
	for path, contents := range files {
		_ = os.WriteFile(path, []byte(contents), 0o644)
	}

	t.Chdir(tmpDir)
	program, _ := CreateProgram("tsconfig.json")
	semantic := CollectSemantic(program)

	expectSymbol(t, semantic.ExternalSymbols, "example-dependency", "api")
	expectSymbol(t, semantic.ExternalSymbols, "example-dependency", "api.run")
	expectSymbol(t, semantic.ExternalSymbols, "example-dependency", "value")
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
	}
}

func expectSymbol(t *testing.T, symbols []ExternalSymbol, namespace string, symbolName string) {
	t.Helper()
	for _, symbol := range symbols {
		if string(symbol.Namespace) == namespace && string(symbol.Name) == symbolName {
			return
		}
	}
	t.Errorf("missing external symbol %q in namespace %q", symbolName, namespace)
}
