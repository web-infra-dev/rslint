package utils

import (
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"gotest.tools/v3/assert"
)

func TestIsSymbolFromDefaultLibrary(t *testing.T) {
	rootDir := fixtures.GetRootDir()

	getTypes := func(code string) (*compiler.Program, *ast.Symbol) {
		filePath := tspath.ResolvePath(rootDir, "file.ts")
		fs := NewOverlayVFSForFile(filePath, code)

		program, err := CreateProgram(true, fs, rootDir, "tsconfig.json", CreateCompilerHost(rootDir, fs))
		assert.NilError(t, err, "couldn't create program")
		sourceFile := program.GetSourceFile(filePath)
		c, done := program.GetTypeChecker(t.Context())
		defer done()
		t := c.GetTypeAtLocation(sourceFile.Statements.Nodes[0].AsTypeAliasDeclaration().Name())
		return program, checker.Type_symbol(t)
	}

	runTestForAliasDeclaration := func(code string, expected bool) {
		program, symbol := getTypes(code)
		result := IsSymbolFromDefaultLibrary(program, symbol)
		if result != expected {
			t.Errorf("Expected %v. Actual %v", expected, result)
		}
	}

	t.Run("is symbol from default library", func(t *testing.T) {
		cases := []string{
			"type Test = Array<number>;",
			"type Test = Map<string,number>;",
			"type Test = Promise<void>",
			"type Test = Error",
			"type Test = Object",
		}
		for _, code := range cases {
			t.Run(code, func(t *testing.T) {
				runTestForAliasDeclaration(code, true)
			})
		}
	})

	t.Run("is not symbol from default library", func(t *testing.T) {
		cases := []string{
			// "const test: Array<number> = [1,2,3];",
			"type Test = number;",
			"type Test = { bar: string; };",
		}
		for _, code := range cases {
			t.Run(code, func(t *testing.T) {
				runTestForAliasDeclaration(code, false)
			})
		}
	})
}

func TestAddDefaultLibraryGlobals(t *testing.T) {
	rootDir := fixtures.GetRootDir()
	filePath := tspath.ResolvePath(rootDir, "file.ts")
	fs := NewOverlayVFSForFile(filePath, "export {}; const top = 1; const localOnly = 1;")
	program, err := CreateProgram(true, fs, rootDir, "tsconfig.json", CreateCompilerHost(rootDir, fs))
	assert.NilError(t, err, "couldn't create program")

	typeChecker, done := program.GetTypeChecker(t.Context())
	defer done()
	globals := map[string]bool{}
	AddDefaultLibraryGlobals(globals, program, typeChecker)

	for _, name := range []string{"Object", "Promise", "top"} {
		if !globals[name] {
			t.Errorf("expected %q to be collected from the active default libraries", name)
		}
	}
	if globals["localOnly"] {
		t.Error("module-local declaration was collected as a default-library global")
	}
}

func TestAddDefaultLibraryTypeGlobalNames(t *testing.T) {
	rootDir := fixtures.GetRootDir()
	filePath := tspath.ResolvePath(rootDir, "file.ts")
	fs := NewOverlayVFSForFile(filePath, "export {}; type NodeListOf = 1; const localOnly = 1;")
	program, err := CreateProgram(true, fs, rootDir, "tsconfig.json", CreateCompilerHost(rootDir, fs))
	assert.NilError(t, err, "couldn't create program")

	typeChecker, done := program.GetTypeChecker(t.Context())
	defer done()
	typeGlobals := map[string]bool{}
	AddDefaultLibraryTypeGlobalNames(typeGlobals, program, typeChecker)

	for _, name := range []string{"Object", "NodeListOf", "ImportMeta"} {
		if !typeGlobals[name] {
			t.Errorf("expected %q to be collected from the default-library type space", name)
		}
	}
	for _, name := range []string{"top", "document", "localOnly"} {
		if typeGlobals[name] {
			t.Errorf("did not expect value-only/local name %q in the default-library type space", name)
		}
	}
}
