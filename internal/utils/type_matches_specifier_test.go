package utils

import (
	"strings"
	"testing"
	"testing/fstest"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/iovfs"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
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

// typeOfLastAlias returns the type of the last type alias declared in fileName.
func typeOfLastAlias(t *testing.T, program *compiler.Program, c *checker.Checker, fileName string) *checker.Type {
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

	return c.GetTypeAtLocation(alias.AsTypeAliasDeclaration().Name())
}

// typeOfTestAlias returns the type `type Test = ...` resolves to in fileName.
func typeOfTestAlias(t *testing.T, program *compiler.Program, c *checker.Checker, fileName string) *checker.Type {
	t.Helper()
	declared := typeOfLastAlias(t, program, c, fileName)
	assert.Assert(t, checker.Type_symbol(declared) != nil, "expected the alias to resolve to a symbol")
	return declared
}

func firstVariableInitializer(t *testing.T, program *compiler.Program, fileName string) *ast.Node {
	t.Helper()
	sourceFile := program.GetSourceFile(fileName)
	assert.Assert(t, sourceFile != nil, "expected %s in the program", fileName)
	for _, statement := range sourceFile.Statements.Nodes {
		if statement.Kind != ast.KindVariableStatement {
			continue
		}
		declarations := statement.AsVariableStatement().DeclarationList.AsVariableDeclarationList().Declarations.Nodes
		if len(declarations) > 0 {
			return declarations[0].AsVariableDeclaration().Initializer
		}
	}
	t.Fatal("expected a variable initializer")
	return nil
}

func TestPromiseInstanceAndConstructorTypes(t *testing.T) {
	rootDir, resolve, baseFS := fixtureRoot()
	for _, test := range []struct {
		name, subject         string
		instance, constructor bool
	}{
		{"native instance", "Promise<number>", true, false},
		{"native constructor", "typeof Promise", false, true},
		{"derived instance", "Derived<number>", true, false},
		{"derived constructor", "typeof Derived", false, true},
		{"indirect constructor", "typeof Child", false, true},
		{"instance union", "Promise<number> | Child", true, false},
		{"constructor union", "typeof Promise | typeof Child", false, true},
		{"mixed union", "Promise<number> | typeof Child", false, false},
		{"tagged instance", "Derived<number> & {tag: string}", true, false},
		{"tagged constructor", "typeof Derived & {tag: string}", false, true},
		{"constructable instance", "Promise<number> & (new () => object)", true, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			file := resolve("file.ts")
			fs := NewOverlayVFS(baseFS, map[string]string{file: "class Derived<T> extends Promise<T> {} class Child extends Derived<number> {} type Test = " + test.subject + ";"})
			program, err := CreateProgram(true, fs, rootDir, "tsconfig.json", CreateCompilerHost(rootDir, fs))
			assert.NilError(t, err)
			c, done := program.GetTypeChecker(t.Context())
			defer done()
			typ := typeOfLastAlias(t, program, c, file)
			facade := lintprogram.NewFromCompiler(program)
			assert.Equal(t, IsPromiseLike(facade, c, typ), test.instance)
			assert.Equal(t, IsPromiseConstructorLike(facade, c, typ), test.constructor)
		})
	}
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
		}}, nil, lintprogram.NewFromCompiler(program))
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

func TestTypeMatchesDeclarationSpecifier(t *testing.T) {
	rootDir, resolve, baseFS := fixtureRoot()
	filePath := resolve("file.ts")
	fs := NewOverlayVFS(baseFS, map[string]string{
		filePath: `import type { External } from "demo-pkg";
import type { Named, Nested } from "ambient";
class Local {}
type LocalType = Local;
type ExternalType = External;
type LibraryType = string;
type AmbientType = Named;
type NestedType = Nested.Inner;
type CustomType = Custom;`,
		resolve("tsconfig.json"):                      `{"compilerOptions":{"strict":true,"typeRoots":["./types"]}}`,
		resolve("node_modules/demo-pkg/package.json"): `{"name":"demo-pkg","version":"1.0.0","types":"index.d.ts"}`,
		resolve("node_modules/demo-pkg/index.d.ts"):   `export declare class External {}`,
		resolve("ambient.d.ts"):                       `declare module "ambient" { export class Named {} export namespace Nested { class Inner {} } }`,
		resolve("types/custom.d.ts"):                  `interface Custom { value: string }`,
	})
	program, err := CreateProgram(true, fs, rootDir, "tsconfig.json", CreateCompilerHost(rootDir, fs))
	assert.NilError(t, err)
	c, done := program.GetTypeChecker(t.Context())
	defer done()
	types := map[string]*checker.Type{}
	for _, node := range program.GetSourceFile(filePath).Statements.Nodes {
		if ast.IsTypeAliasDeclaration(node) {
			types[node.Name().Text()] = c.GetTypeAtLocation(node.Name())
		}
	}
	for _, test := range []struct {
		name   string
		option map[string]any
		want   bool
	}{
		{"LocalType", map[string]any{"from": "file"}, true},
		{"LocalType", map[string]any{"from": "file", "path": "./file.ts"}, true},
		{"LocalType", map[string]any{"from": "file", "path": "**/file.ts"}, true},
		{"LocalType", map[string]any{"from": "file", "path": "file.ts"}, false},
		{"LocalType", map[string]any{"from": "file", "path": ""}, false},
		{"LocalType", map[string]any{"from": "package"}, false},
		{"LibraryType", map[string]any{"from": "lib"}, true},
		{"LibraryType", map[string]any{"from": "file"}, false},
		{"ExternalType", map[string]any{"from": "package"}, true},
		{"ExternalType", map[string]any{"from": "package", "package": "demo", "name": []any{"unrelated"}}, true},
		{"ExternalType", map[string]any{"from": "package", "package": "other"}, false},
		{"ExternalType", map[string]any{"from": "file"}, false},
		{"AmbientType", map[string]any{"from": "package"}, true},
		{"AmbientType", map[string]any{"from": "package", "package": "ambient"}, true},
		{"AmbientType", map[string]any{"from": "package", "package": ""}, false},
		// Declaration-location matching stops at the nearest namespace, unlike
		// the existing TypeScript-ESLint specifier's transparent namespace walk.
		{"NestedType", map[string]any{"from": "package"}, false},
		{"NestedType", map[string]any{"from": "file"}, true},
		{"CustomType", map[string]any{"from": "file"}, false},
	} {
		specifier, ok := ParseTypeOrValueSpecifier(test.option)
		assert.Assert(t, ok)
		assert.Assert(t, types[test.name] != nil)
		assert.Equal(t, TypeMatchesDeclarationSpecifier(types[test.name], specifier, lintprogram.NewFromCompiler(program)), test.want, "%s: %v", test.name, test.option)
	}
	// Directly constructed specifiers honor non-empty paths and package names,
	// just like specifiers decoded from rule options.
	assert.Equal(t, TypeMatchesDeclarationSpecifier(types["LocalType"], TypeOrValueSpecifier{
		From: TypeOrValueSpecifierFromFile, Path: "./missing.ts",
	}, lintprogram.NewFromCompiler(program)), false)
	assert.Equal(t, TypeMatchesDeclarationSpecifier(types["AmbientType"], TypeOrValueSpecifier{
		From: TypeOrValueSpecifierFromPackage, Package: "other",
	}, lintprogram.NewFromCompiler(program)), false)
}

func TestTypeMatchesDeclarationSpecifierDefaultTypeRoots(t *testing.T) {
	rootDir, resolve, baseFS := fixtureRoot()
	filePath := resolve("file.ts")
	fs := NewOverlayVFS(baseFS, map[string]string{
		filePath:                 `type Test = DefaultRoot;`,
		resolve("tsconfig.json"): `{"compilerOptions":{"types":[]},"files":["file.ts","node_modules/@types/global/index.d.ts"]}`,
		resolve("node_modules/@types/global/index.d.ts"): `interface DefaultRoot { value: string }`,
	})
	program, err := CreateProgram(true, fs, rootDir, "tsconfig.json", CreateCompilerHost(rootDir, fs))
	assert.NilError(t, err)
	c, done := program.GetTypeChecker(t.Context())
	defer done()
	specifier, ok := ParseTypeOrValueSpecifier(map[string]any{"from": "file", "path": "**/*.ts"})
	assert.Assert(t, ok)
	assert.Equal(t, TypeMatchesDeclarationSpecifier(typeOfTestAlias(t, program, c, filePath), specifier, lintprogram.NewFromCompiler(program)), false)
}

// The overlay supplies every file. Missing package.json probes must not pass
// synthetic UNC roots to io/fs, whose paths cannot start with a slash.
type declarationPathFS struct{ vfs.FS }

func (declarationPathFS) FileExists(string) bool { return false }

func TestTypeMatchesDeclarationSpecifierPaths(t *testing.T) {
	for _, test := range []struct {
		name, directory, file, pattern string
		caseSensitive                  bool
		typeRoots                      []string
		want                           bool
	}{
		// File-glob differences are documented with concrete examples.
		{"glob ./src/*.ts", "/repo", "/repo/src/foo.ts", "./src/*.ts", true, nil, true},
		{"glob **/[!a-z]*.ts", "/repo", "/repo/foo.ts", "**/[!a-z]*.ts", true, nil, false},
		{"glob **/[^a-z]*.ts", "/repo", "/repo/.foo.ts", "**/[^a-z]*.ts", true, nil, false},
		{"glob **/[[:digit:]]*.ts", "/repo", "/repo/1.ts", "**/[[:digit:]]*.ts", true, nil, false},
		{"glob **/!(foo|bar).ts", "/repo", "/repo/foobar.ts", "**/!(foo|bar).ts", true, nil, true},
		{"glob !(**/foo.ts)", "/repo", "/repo/foo.ts", "!(**/foo.ts)", true, nil, true},
		{"glob **/.*/foo.ts", "/repo", "/repo/foo.ts", "**/.*/foo.ts", true, nil, true},
		{"glob ./*", "/repo", "/repo/foo.ts", "./*", true, nil, true},
		{"glob **/foo.(ts)", "/repo", "/repo/foo.ts", "**/foo.(ts)", true, nil, false},
		{"glob **/[foo].ts", "/repo", "/repo/[foo].ts", "**/[foo].ts", true, nil, false},
		{"glob **/file{01..03}.ts", "/repo", "/repo/file01.ts", "**/file{01..03}.ts", true, nil, true},
		{"glob **/file{1..3..2}.ts", "/repo", "/repo/file2.ts", "**/file{1..3..2}.ts", true, nil, false},
		{"glob ./#*.ts", "/repo", "/repo/#foo.ts", "./#*.ts", true, nil, true},
		{"glob **//foo.ts", "/repo", "/repo/foo.ts", "**//foo.ts", true, nil, true},
		{"glob ./!(bar).ts", "/repo", "/repo/foo.ts", "./!(bar).ts", true, nil, true},
		{"glob **/foo.ts/**", "/repo", "/repo/foo.ts", "**/foo.ts/**", true, nil, false},
		{"relative POSIX", "/repo/project", "/repo/project/src/file.ts", "./src/*.ts", true, nil, true},
		{"absolute POSIX", "/repo/project", "/repo/project/src/file.ts", "/repo/project/src/*.ts", true, nil, true},
		{"case-sensitive match", "/repo/project", "/repo/project/Src/File.ts", "./Src/File.ts", true, nil, true},
		{"case-sensitive miss", "/repo/project", "/repo/project/Src/File.ts", "./src/file.ts", true, nil, false},
		{"canonical insensitive path", "/Repo/Project", "/Repo/Project/Src/File.ts", "./src/file.ts", false, nil, true},
		// Upstream folds the path on insensitive filesystems, not the pattern.
		{"pattern remains case-sensitive", "/Repo/Project", "/Repo/Project/Src/File.ts", "./Src/File.ts", false, nil, false},
		{"Windows relative", `C:\Repo\Project`, `C:\Repo\Project\src\file.ts`, "./src/*.ts", false, nil, true},
		{"Windows absolute", "C:/Repo/Project", "C:/Repo/Project/src/file.ts", "c:/repo/project/src/*.ts", false, nil, true},
		{"Windows drive casing", "C:/Repo/Project", "c:/Repo/Project/src/file.ts", "./src/*.ts", false, nil, true},
		{"Windows other drive", "C:/repo/project", "D:/repo/project/src/file.ts", "**/file.ts", false, nil, false},
		{"UNC relative", "//Server/Share/Project", "//Server/Share/Project/src/file.ts", "./src/*.ts", false, nil, true},
		{"UNC absolute", "//Server/Share/Project", "//Server/Share/Project/src/file.ts", "//server/share/project/src/*.ts", false, nil, true},
		{"UNC other share", "//server/share/project", "//server/other/project/src/file.ts", "**/file.ts", false, nil, false},
		{"Unicode directory", "/repo/Kit/代码", "/repo/Kit/代码/file.ts", "./file.ts", false, nil, true},
		{"spaces in directory", "C:/My Project", "C:/My Project/src/file.ts", "./src/*.ts", false, nil, true},
		{"parent segments", "/repo/project/src/..", "/repo/project/src/../file.ts", "./file.ts", true, nil, true},
		{"trailing directory separator", "C:/repo/project/", "C:/repo/project/file.ts", "./file.ts", false, nil, true},
		{"backslashes escape glob characters", "C:/repo/project", "C:/repo/project/src/file.ts", `**\file.ts`, false, nil, false},
		{"escaped dot in pattern", "C:/repo/project", "C:/repo/project/file.ts", `./file\.ts`, false, nil, true},
		{"leading pattern space", "/repo/project", "/repo/project/file.ts", " **/file.ts", true, nil, false},
		{"trailing pattern space", "/repo/project", "/repo/project/file.ts", "**/file.ts ", true, nil, false},
		{"Unicode pattern space", "C:/repo/project", "C:/repo/project/file.ts", "\u00a0**/file.ts", false, nil, false},
		{"outside working directory", "/repo/project", "/repo/shared/file.ts", "**/file.ts", true, nil, false},
		// Directory containment must not confuse siblings sharing a prefix.
		{"sibling directory sharing cwd prefix", "/repo/project", "/repo/project-extra/file.ts", "**/file.ts", true, nil, false},
		{"configured type root", "/repo/project", "/repo/project/types/file.ts", "**/file.ts", true, []string{"/repo/project/types"}, false},
		{"type root prefix", "/repo/project", "/repo/project/types-extra/file.ts", "**/file.ts", true, []string{"/repo/project/types"}, true},
		{"Windows cwd prefix", "C:/Repo/Project", "c:/repo/project-extra/file.ts", "**/file.ts", false, nil, false},
		{"UNC cwd prefix", "//server/share/project", "//server/share/project-extra/file.ts", "**/file.ts", false, nil, false},
		{"Windows type root prefix", "C:/Repo/Project", "c:/repo/project/Types-extra/file.ts", "**/file.ts", false, []string{"C:/repo/project/types"}, true},
		{"UNC type root prefix", "//server/share/project", "//server/share/project/types-extra/file.ts", "**/file.ts", false, []string{"//server/share/project/types"}, true},
		{"type root trailing separator", "/repo/project", "/repo/project/types/file.ts", "**/file.ts", true, []string{"/repo/project/types/"}, false},
		{"Windows type root casing", "C:/Repo/Project", "C:/Repo/Project/Types/file.ts", "**/file.ts", false, []string{"c:/repo/project/types"}, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			directory := tspath.NormalizePath(test.directory)
			file := tspath.NormalizePath(test.file)
			// All reads use exact fixture names; the flag controls the compiler's
			// path identity without relying on the machine running this test.
			fs := NewOverlayVFS(declarationPathFS{iovfs.From(fstest.MapFS{}, test.caseSensitive)}, map[string]string{
				file: "class Demo {}\ntype Test = Demo;",
			})
			program, err := CreateProgramFromOptions(true, &core.CompilerOptions{
				NoLib: core.TSTrue, Types: []string{}, TypeRoots: test.typeRoots,
			}, []string{file}, CreateCompilerHost(directory, fs))
			assert.NilError(t, err)
			c, done := program.GetTypeChecker(t.Context())
			defer done()
			specifier, ok := ParseTypeOrValueSpecifier(map[string]any{"from": "file", "path": test.pattern})
			assert.Assert(t, ok)
			assert.Equal(t, TypeMatchesDeclarationSpecifier(typeOfTestAlias(t, program, c, file), specifier, lintprogram.NewFromCompiler(program)), test.want)
			if test.typeRoots != nil {
				// Omitting the glob must retain the same type-root exclusion.
				assert.Equal(t, TypeMatchesDeclarationSpecifier(typeOfTestAlias(t, program, c, file), TypeOrValueSpecifier{
					From: TypeOrValueSpecifierFromFile,
				}, lintprogram.NewFromCompiler(program)), test.want)
			}
		})
	}
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
		}}, nil, lintprogram.NewFromCompiler(program))
	}

	assert.Equal(t, matches("demo-pkg"), true)
	assert.Equal(t, matches("other"), false)
	// no-sync must retain package identity after resolving a workspace link.
	assert.Equal(t, TypeMatchesDeclarationSpecifier(demo, TypeOrValueSpecifier{
		From: TypeOrValueSpecifierFromPackage, Package: "demo-pkg",
	}, lintprogram.NewFromCompiler(program)), true)
	assert.Equal(t, TypeMatchesDeclarationSpecifier(demo, TypeOrValueSpecifier{
		From: TypeOrValueSpecifierFromFile, Path: "**/index.d.ts",
	}, lintprogram.NewFromCompiler(program)), false)
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
		}}, nil, lintprogram.NewFromCompiler(program))
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
		}}, nil, lintprogram.NewFromCompiler(program))
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
	}}, nil, lintprogram.NewFromCompiler(program)), false)
}

func TestTypeMatchesSomeSpecifierExpandsIntersections(t *testing.T) {
	rootDir, resolve, baseFS := fixtureRoot()

	filePath := resolve("file.ts")
	fs := NewOverlayVFS(baseFS, map[string]string{
		filePath: "interface Extra { extra: true }\ntype Test = Error & Extra;\n",
	})

	program, err := CreateProgram(true, fs, rootDir, "tsconfig.json", CreateCompilerHost(rootDir, fs))
	assert.NilError(t, err, "couldn't create program")
	c, done := program.GetTypeChecker(t.Context())
	defer done()

	intersection := typeOfLastAlias(t, program, c, filePath)
	assert.Equal(t, TypeMatchesSomeSpecifier(intersection, []TypeOrValueSpecifier{{
		From: TypeOrValueSpecifierFromLib,
		Name: NameList{"Error"},
	}}, nil, lintprogram.NewFromCompiler(program)), true)
}

func TestTypeMatchesSomeSpecifierUsesResolutionPackageID(t *testing.T) {
	rootDir, resolve, baseFS := fixtureRoot()

	tests := []struct {
		name            string
		rootPackageJSON string
		nestedPackage   string
		wantMatch       bool
	}{
		{
			name:            "named nested package does not replace entry package",
			rootPackageJSON: `{"name":"demo-pkg","version":"1.0.0","types":"dist/index.d.ts"}`,
			nestedPackage:   `{"name":"nested-pkg","version":"2.0.0"}`,
			wantMatch:       true,
		},
		{
			name:            "package without version has no package ID",
			rootPackageJSON: `{"name":"demo-pkg","types":"dist/index.d.ts"}`,
			nestedPackage:   `{"sideEffects":false}`,
			wantMatch:       false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			filePath := resolve("file.ts")
			fs := NewOverlayVFS(baseFS, map[string]string{
				filePath: "import { Demo, demoValue } from 'demo-pkg';\nconst value = demoValue;\ntype Test = Demo;\n",
				resolve("node_modules/demo-pkg/package.json"):      test.rootPackageJSON,
				resolve("node_modules/demo-pkg/dist/package.json"): test.nestedPackage,
				resolve("node_modules/demo-pkg/dist/index.d.ts"):   "export declare class Demo {}\nexport declare function demoValue(): void;\n",
			})

			program, err := CreateProgram(true, fs, rootDir, "tsconfig.json", CreateCompilerHost(rootDir, fs))
			assert.NilError(t, err, "couldn't create program")
			c, done := program.GetTypeChecker(t.Context())
			defer done()

			demo := typeOfTestAlias(t, program, c, filePath)
			assert.Equal(t, TypeMatchesSomeSpecifier(demo, []TypeOrValueSpecifier{{
				From:    TypeOrValueSpecifierFromPackage,
				Name:    NameList{"Demo"},
				Package: "demo-pkg",
			}}, nil, lintprogram.NewFromCompiler(program)), test.wantMatch)

			valueNode := firstVariableInitializer(t, program, filePath)
			assert.Equal(t, ValueMatchesSomeSpecifier(valueNode, []TypeOrValueSpecifier{{
				From:    TypeOrValueSpecifierFromPackage,
				Name:    NameList{"demoValue"},
				Package: "demo-pkg",
			}}, lintprogram.NewFromCompiler(program), c.GetTypeAtLocation(valueNode)), test.wantMatch)
		})
	}
}

func TestTypeMatchesSomeSpecifierChecksWholeIntersectionBeforeConstituents(t *testing.T) {
	rootDir, resolve, baseFS := fixtureRoot()

	tests := []struct {
		name string
		code string
		want bool
	}{
		{
			name: "whole alias matches",
			code: "type SafePromise = Promise<number> & { __safeBrand: string };\ntype Test = SafePromise;\n",
			want: true,
		},
		{
			name: "inlined alias is not a constituent",
			code: "type SafePromise = Promise<number> & { __safeBrand: string };\ntype Test = SafePromise & {};\n",
			want: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			filePath := resolve("file.ts")
			fs := NewOverlayVFS(baseFS, map[string]string{filePath: test.code})
			program, err := CreateProgram(true, fs, rootDir, "tsconfig.json", CreateCompilerHost(rootDir, fs))
			assert.NilError(t, err, "couldn't create program")
			c, done := program.GetTypeChecker(t.Context())
			defer done()

			intersection := typeOfLastAlias(t, program, c, filePath)
			assert.Equal(t, TypeMatchesSomeSpecifier(intersection, []TypeOrValueSpecifier{{
				From: TypeOrValueSpecifierFromFile,
				Name: NameList{"SafePromise"},
			}}, nil, lintprogram.NewFromCompiler(program)), test.want)
		})
	}
}

func TestTypeMatchesSomeSpecifierDistinguishesOmittedAndEmptyFilePath(t *testing.T) {
	rootDir, resolve, baseFS := fixtureRoot()

	filePath := resolve("file.ts")
	fs := NewOverlayVFS(baseFS, map[string]string{
		filePath: "interface Demo { value: string }\ntype Test = Demo;\n",
	})

	program, err := CreateProgram(true, fs, rootDir, "tsconfig.json", CreateCompilerHost(rootDir, fs))
	assert.NilError(t, err, "couldn't create program")
	c, done := program.GetTypeChecker(t.Context())
	defer done()

	demo := typeOfTestAlias(t, program, c, filePath)
	decode := func(raw map[string]any) TypeOrValueSpecifier {
		t.Helper()
		specifier, ok := ParseTypeOrValueSpecifier(raw)
		assert.Assert(t, ok)
		return specifier
	}

	assert.Equal(t, TypeMatchesSomeSpecifier(demo, []TypeOrValueSpecifier{
		decode(map[string]any{"from": "file", "name": "Demo"}),
	}, nil, lintprogram.NewFromCompiler(program)), true)
	assert.Equal(t, TypeMatchesSomeSpecifier(demo, []TypeOrValueSpecifier{
		decode(map[string]any{"from": "file", "name": "Demo", "path": ""}),
	}, nil, lintprogram.NewFromCompiler(program)), false)
}
