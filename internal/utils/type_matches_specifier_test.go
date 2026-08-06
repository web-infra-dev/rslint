package utils

import (
	"strings"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"gotest.tools/v3/assert"
)

// fixtureRoot resolves paths against the shared fixtures directory, which is
// also the directory the test tsconfig files below live in.
func fixtureRoot() (string, func(name string) string, vfs.FS) {
	rootDir := fixtures.GetRootDir()
	return rootDir.Dir, func(name string) string {
		return tspath.ResolvePath(rootDir.Dir, name)
	}, rootDir.FS
}

// symlinkVFS reports one directory as a link to another, the layout a workspace
// install produces when it links node_modules/<pkg> at a package checked into
// the repository.
type symlinkVFS struct {
	vfs.FS
	link   string
	target string
}

func (f *symlinkVFS) Realpath(path string) string {
	if path == f.link {
		return f.target
	}
	if rest, linked := strings.CutPrefix(path, f.link+"/"); linked {
		return f.target + "/" + rest
	}
	return f.FS.Realpath(path)
}

// typeOfTestAlias returns the type `type Test = ...` resolves to in fileName.
func typeOfTestAlias(t *testing.T, program *compiler.Program, c *checker.Checker, fileName string) *checker.Type {
	t.Helper()
	sourceFile := program.GetSourceFile(fileName)
	assert.Assert(t, sourceFile != nil, "expected %s in the program", fileName)

	var alias *ast.Node
	for _, statement := range sourceFile.Statements.Nodes {
		if ast.IsTypeAliasDeclaration(statement) {
			alias = statement
		}
	}
	assert.Assert(t, alias != nil, "expected a type alias declaration")

	declared := c.GetTypeAtLocation(alias.AsTypeAliasDeclaration().Name())
	assert.Assert(t, checker.Type_symbol(declared) != nil, "expected the alias to resolve to a symbol")
	return declared
}

func TestTypeMatchesSomeSpecifierFromPackage(t *testing.T) {
	rootDir, resolve, baseFS := fixtureRoot()

	filePath := resolve("file.ts")
	fs := NewOverlayVFS(baseFS, map[string]string{
		filePath: "import { Demo } from 'demo-pkg';\ntype Test = Demo;\n",
		resolve("node_modules/demo-pkg/package.json"): `{"name":"demo-pkg","version":"1.0.0","types":"index.d.ts"}`,
		resolve("node_modules/demo-pkg/index.d.ts"):   "export declare class Demo {}\n",
	})

	program, err := CreateProgram(true, fs, rootDir, "tsconfig.json", CreateCompilerHost(rootDir, fs))
	assert.NilError(t, err, "couldn't create program")
	c, done := program.GetTypeChecker(t.Context())
	defer done()

	demo := typeOfTestAlias(t, program, c, filePath)

	matches := func(packageName string) bool {
		return TypeMatchesSomeSpecifier(demo, []TypeOrValueSpecifier{{
			From:    TypeOrValueSpecifierFromPackage,
			Name:    NameList{"Demo"},
			Package: packageName,
		}}, nil, program)
	}

	// The package name becomes an unanchored JavaScript pattern, so any substring
	// of "demo-pkg/index.d.ts" matches it, and regular expression syntax applies.
	assert.Equal(t, matches("demo-pkg"), true)
	assert.Equal(t, matches("demo"), true)
	assert.Equal(t, matches("emo-pk"), true)
	assert.Equal(t, matches("demo.pkg"), true)
	assert.Equal(t, matches("d[ei]mo"), true)
	assert.Equal(t, matches("other"), false)
	// An unparsable pattern matches nothing.
	assert.Equal(t, matches("("), false)
}

// A workspace package is installed as a link, so its declarations resolve to a
// real path outside node_modules while still belonging to the linked package.
func TestTypeMatchesSomeSpecifierFromLinkedWorkspacePackage(t *testing.T) {
	rootDir, resolve, baseFS := fixtureRoot()

	filePath := resolve("file.ts")
	files := map[string]string{
		filePath:                           "import { Demo } from 'demo-pkg';\ntype Test = Demo;\n",
		resolve("tsconfig.workspace.json"): `{"compilerOptions":{"target":"esnext","module":"commonjs","strict":true,"types":[]},"files":["file.ts"]}`,
	}
	// The resolver walks the link path and the program then loads the real one,
	// so the package has to be readable under both.
	for _, directory := range []string{"node_modules/demo-pkg", "packages/demo-pkg"} {
		files[resolve(directory+"/package.json")] = `{"name":"demo-pkg","version":"1.0.0","types":"index.d.ts"}`
		files[resolve(directory+"/index.d.ts")] = "export declare class Demo {}\n"
	}
	fs := &symlinkVFS{
		FS:     NewOverlayVFS(baseFS, files),
		link:   resolve("node_modules/demo-pkg"),
		target: resolve("packages/demo-pkg"),
	}

	program, err := CreateProgram(true, fs, rootDir, "tsconfig.workspace.json", CreateCompilerHost(rootDir, fs))
	assert.NilError(t, err, "couldn't create program")
	c, done := program.GetTypeChecker(t.Context())
	defer done()

	demo := typeOfTestAlias(t, program, c, filePath)

	declaration := ast.GetSourceFileOfNode(checker.Type_symbol(demo).Declarations[0])
	assert.Assert(t, !strings.Contains(declaration.FileName(), "/node_modules/"),
		"expected the declaration to resolve through the link, got %s", declaration.FileName())
	assert.Equal(t, program.IsSourceFileFromExternalLibrary(declaration), true)

	matches := func(packageName string) bool {
		return TypeMatchesSomeSpecifier(demo, []TypeOrValueSpecifier{{
			From:    TypeOrValueSpecifierFromPackage,
			Name:    NameList{"Demo"},
			Package: packageName,
		}}, nil, program)
	}

	assert.Equal(t, matches("demo-pkg"), true)
	assert.Equal(t, matches("other"), false)
}

// Dual-published packages drop unnamed package.json files into their build
// directories; the type still belongs to the named package enclosing them.
func TestTypeMatchesSomeSpecifierFromNestedPackageJson(t *testing.T) {
	rootDir, resolve, baseFS := fixtureRoot()

	filePath := resolve("file.ts")
	fs := NewOverlayVFS(baseFS, map[string]string{
		filePath: "import { Demo } from 'demo-pkg';\ntype Test = Demo;\n",
		resolve("node_modules/demo-pkg/package.json"):      `{"name":"demo-pkg","version":"1.0.0","types":"dist/index.d.ts"}`,
		resolve("node_modules/demo-pkg/dist/package.json"): `{"sideEffects":false}`,
		resolve("node_modules/demo-pkg/dist/index.d.ts"):   "export declare class Demo {}\n",
	})

	program, err := CreateProgram(true, fs, rootDir, "tsconfig.json", CreateCompilerHost(rootDir, fs))
	assert.NilError(t, err, "couldn't create program")
	c, done := program.GetTypeChecker(t.Context())
	defer done()

	demo := typeOfTestAlias(t, program, c, filePath)

	matches := func(packageName string) bool {
		return TypeMatchesSomeSpecifier(demo, []TypeOrValueSpecifier{{
			From:    TypeOrValueSpecifierFromPackage,
			Name:    NameList{"Demo"},
			Package: packageName,
		}}, nil, program)
	}

	assert.Equal(t, matches("demo-pkg"), true)
	assert.Equal(t, matches("other"), false)
}

// A namespace between the declaration and its `declare module` is transparent.
func TestTypeMatchesSomeSpecifierFromDeclareModuleNamespace(t *testing.T) {
	rootDir, resolve, baseFS := fixtureRoot()

	filePath := resolve("file.ts")
	fs := NewOverlayVFS(baseFS, map[string]string{
		filePath:                         "import inner = require('ambient-pkg');\ntype Test = inner.Demo;\n",
		resolve("ambient.d.ts"):          "declare module 'ambient-pkg' {\n  namespace inner {\n    class Demo {}\n  }\n  export = inner;\n}\n",
		resolve("tsconfig.ambient.json"): `{"compilerOptions":{"target":"esnext","module":"commonjs","strict":true,"types":[]},"files":["file.ts","ambient.d.ts"]}`,
	})

	program, err := CreateProgram(true, fs, rootDir, "tsconfig.ambient.json", CreateCompilerHost(rootDir, fs))
	assert.NilError(t, err, "couldn't create program")
	c, done := program.GetTypeChecker(t.Context())
	defer done()

	demo := typeOfTestAlias(t, program, c, filePath)

	matches := func(packageName string) bool {
		return TypeMatchesSomeSpecifier(demo, []TypeOrValueSpecifier{{
			From:    TypeOrValueSpecifierFromPackage,
			Name:    NameList{"Demo"},
			Package: packageName,
		}}, nil, program)
	}

	// `declare module` names are compared exactly, unlike declaration file paths.
	assert.Equal(t, matches("ambient-pkg"), true)
	assert.Equal(t, matches("ambient"), false)
}

// A type declared by the project itself belongs to no package, even when the
// project has a named package.json of its own.
func TestTypeMatchesSomeSpecifierFromPackageIgnoresLocalTypes(t *testing.T) {
	rootDir, resolve, baseFS := fixtureRoot()

	filePath := resolve("file.ts")
	fs := NewOverlayVFS(baseFS, map[string]string{
		filePath:                       "import { Demo } from './local';\ntype Test = Demo;\n",
		resolve("local.ts"):            "export declare class Demo {}\n",
		resolve("package.json"):        `{"name":"demo-pkg","version":"1.0.0"}`,
		resolve("tsconfig.local.json"): `{"compilerOptions":{"target":"esnext","module":"commonjs","strict":true,"types":[]},"files":["file.ts","local.ts"]}`,
	})

	program, err := CreateProgram(true, fs, rootDir, "tsconfig.local.json", CreateCompilerHost(rootDir, fs))
	assert.NilError(t, err, "couldn't create program")
	c, done := program.GetTypeChecker(t.Context())
	defer done()

	demo := typeOfTestAlias(t, program, c, filePath)

	assert.Equal(t, TypeMatchesSomeSpecifier(demo, []TypeOrValueSpecifier{{
		From:    TypeOrValueSpecifierFromPackage,
		Name:    NameList{"Demo"},
		Package: "demo-pkg",
	}}, nil, program), false)
}
