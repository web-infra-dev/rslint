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
	semantic := semanticFromFiles(t, map[string]string{
		"tsconfig.json":                   "{\n\t\"compilerOptions\": {\"moduleResolution\": \"node\"},\n\t\"include\": [\"./index.ts\"]\n}",
		"index.ts":                        `import "react";`,
		"node_modules/react/package.json": "{\n\t\"name\": \"react\",\n\t\"version\": \"1.0.0\",\n\t\"types\": \"index.d.ts\"\n}",
		"node_modules/react/index.d.ts":   "export namespace JSX {\ninterface IntrinsicElements {\ndiv: unknown;\n}\n}",
	})

	expectSymbol(t, semantic.ExternalSymbols, "react", "JSX.IntrinsicElements.div")
}

func TestGlobalJSXIntrinsicElementSymbols(t *testing.T) {
	semantic := semanticFromFiles(t, map[string]string{
		"tsconfig.json":                          "{\n\t\"compilerOptions\": {\"jsx\": \"preserve\", \"moduleResolution\": \"node\", \"types\": [\"react\"]},\n\t\"include\": [\"./index.tsx\"]\n}",
		"index.tsx":                              `const element = <div />;`,
		"node_modules/@types/react/package.json": "{\n\t\"name\": \"@types/react\",\n\t\"version\": \"1.0.0\",\n\t\"types\": \"index.d.ts\"\n}",
		"node_modules/@types/react/index.d.ts":   "export = React;\nexport as namespace React;\ndeclare namespace React {\nfunction createElement(): unknown;\nnamespace JSX {\ninterface IntrinsicElements {\ndiv: unknown;\n}\n}\n}\n",
	})

	expectSymbol(t, semantic.ExternalSymbols, globalNamespace, "React.JSX.IntrinsicElements.div")
}

func TestDependencySymbols(t *testing.T) {
	semantic := semanticFromFiles(t, map[string]string{
		"tsconfig.json": "{\n\t\"compilerOptions\": {\"moduleResolution\": \"node\"},\n\t\"include\": [\"./index.ts\"]\n}",
		"index.ts":      `import { api } from "example-dependency"; api.run();`,
		"node_modules/example-dependency/package.json": "{\n\t\"name\": \"example-dependency\",\n\t\"version\": \"1.0.0\",\n\t\"types\": \"index.d.ts\"\n}",
		"node_modules/example-dependency/index.d.ts":   "export declare const api: { run(): void };\nexport declare function value(): void;",
	})

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

func semanticFromFiles(t *testing.T, files map[string]string) Semantic {
	t.Helper()
	tmpDir := t.TempDir()
	for name, contents := range files {
		path := filepath.Join(tmpDir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("failed to create fixture directory: %v", err)
		}
		if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
			t.Fatalf("failed to write fixture file %q: %v", name, err)
		}
	}

	t.Chdir(tmpDir)
	program, err := CreateProgram("tsconfig.json")
	if err != nil {
		t.Fatalf("failed to create program: %v", err)
	}
	return CollectSemantic(program)
}
