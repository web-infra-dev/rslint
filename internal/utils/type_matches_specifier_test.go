package utils

import (
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"gotest.tools/v3/assert"
)

func TestTypeMatchesSomeSpecifierFromPackage(t *testing.T) {
	rootDir := fixtures.GetRootDir()
	resolve := func(name string) string { return tspath.ResolvePath(rootDir.Dir, name) }

	filePath := resolve("file.ts")
	fs := NewOverlayVFS(rootDir.FS, map[string]string{
		filePath: "import { Demo } from 'demo-pkg';\ntype Test = Demo;\n",
		resolve("node_modules/demo-pkg/package.json"): `{"name":"demo-pkg","version":"1.0.0","types":"index.d.ts"}`,
		resolve("node_modules/demo-pkg/index.d.ts"):   "export declare class Demo {}\n",
	})

	program, err := CreateProgram(true, fs, rootDir.Dir, "tsconfig.json", CreateCompilerHost(rootDir.Dir, fs))
	assert.NilError(t, err, "couldn't create program")
	sourceFile := program.GetSourceFile(filePath)
	c, done := program.GetTypeChecker(t.Context())
	defer done()

	var alias *ast.Node
	for _, statement := range sourceFile.Statements.Nodes {
		if ast.IsTypeAliasDeclaration(statement) {
			alias = statement
		}
	}
	assert.Assert(t, alias != nil, "expected a type alias declaration")
	demo := c.GetTypeAtLocation(alias.AsTypeAliasDeclaration().Name())
	assert.Assert(t, checker.Type_symbol(demo) != nil, "expected Demo to resolve to a symbol")

	matches := func(packageName string) bool {
		return TypeMatchesSomeSpecifier(demo, []TypeOrValueSpecifier{{
			From:    TypeOrValueSpecifierFromPackage,
			Name:    NameList{"Demo"},
			Package: packageName,
		}}, nil, program)
	}

	// The package name is tested against the resolved package name without
	// anchors, so any substring of "demo-pkg" matches it.
	assert.Equal(t, matches("demo-pkg"), true)
	assert.Equal(t, matches("demo"), true)
	assert.Equal(t, matches("emo-pk"), true)
	assert.Equal(t, matches("other"), false)
}

func TestResolvedPackageName(t *testing.T) {
	assert.Equal(t, resolvedPackageName("/repo/node_modules/demo-pkg/index.d.ts"), "demo-pkg")
	assert.Equal(t, resolvedPackageName("/repo/node_modules/@scope/pkg/lib/index.d.ts"), "@scope/pkg")
	assert.Equal(t, resolvedPackageName("/repo/node_modules/a/node_modules/b/index.d.ts"), "b")
	assert.Equal(t, resolvedPackageName("/repo/src/index.ts"), "")
}
