package program_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	utils "github.com/web-infra-dev/rslint/internal/utils"
)

func TestResolveSourceFileFromSourceFile(t *testing.T) {
	t.Parallel()

	program, sourceFile, specifier := programForImport(t, map[string]string{
		"/import-fixture/file.ts": "import value from \"./bar\";\nvoid value;\n",
		"/import-fixture/bar.ts":  "export default 1;\n",
	}, "/import-fixture/file.ts")

	resolvedPath, target, ok := program.ResolveModule(sourceFile, specifier)
	if !ok {
		t.Fatal("ResolveSourceFileFromSourceFile() did not resolve ./bar")
	}
	if target == nil {
		t.Fatal("ResolveSourceFileFromSourceFile() returned nil target")
		return
	}
	if got := tspath.NormalizeSlashes(resolvedPath); !strings.HasSuffix(got, "/bar.ts") {
		t.Fatalf("resolvedPath = %q, want suffix /bar.ts", resolvedPath)
	}
	if target.FileName() != resolvedPath {
		t.Fatalf("target.FileName() = %q, want %q", target.FileName(), resolvedPath)
	}
}

func TestResolveSourceFileFromSourceFileInvalidInput(t *testing.T) {
	t.Parallel()

	program, sourceFile, _ := programForImport(t, map[string]string{
		"/import-fixture/file.ts": "import value from \"./bar\";\nvoid value;\n",
		"/import-fixture/bar.ts":  "export default 1;\n",
	}, "/import-fixture/file.ts")

	if resolvedPath, target, ok := program.ResolveModule(sourceFile, nil); ok || resolvedPath != "" || target != nil {
		t.Fatalf("ResolveSourceFileFromSourceFile(nil) = (%q, %#v, %v), want empty result", resolvedPath, target, ok)
	}
}

// TestResolveFromSourceFileRequire covers the three `require()` shapes that
// differ in how their specifier reaches a file: TypeScript's own resolution
// records the call in a JavaScript file, while the same call in a TypeScript
// file — relative or package — is absent from that cache and reaches its target
// through the resolver.
func TestResolveFromSourceFileRequire(t *testing.T) {
	t.Parallel()

	const packageFile = "/require-fixture/node_modules/some-package/index.d.ts"
	files := map[string]string{
		packageFile:                    "declare const value: unknown;\nexport = value;\n",
		"/require-fixture/dep.ts":      "export const dep = 1;\n",
		"/require-fixture/consumer.js": `const pkg = require("some-package");`,
		"/require-fixture/package.ts":  `const pkg = require("some-package");`,
		"/require-fixture/relative.ts": `const dep = require("./dep");`,
	}

	tests := []struct {
		name     string
		fileName string
		wantPath string
	}{
		{
			name:     "package specifier in a JavaScript file",
			fileName: "/require-fixture/consumer.js",
			wantPath: packageFile,
		},
		{
			name:     "relative specifier in a TypeScript file",
			fileName: "/require-fixture/relative.ts",
			wantPath: "/require-fixture/dep.ts",
		},
		{
			name:     "package specifier in a TypeScript file",
			fileName: "/require-fixture/package.ts",
			wantPath: packageFile,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			program, sourceFile, specifier := programForRequire(t, files, test.fileName, &core.CompilerOptions{
				Module:  core.ModuleKindCommonJS,
				AllowJs: core.TSTrue,
			})

			resolvedPath, _, ok := program.ResolveModule(sourceFile, specifier)
			if ok != (test.wantPath != "") {
				t.Fatalf("ResolveFromSourceFile() ok = %v, want %v (path %q)", ok, test.wantPath != "", resolvedPath)
			}
			if got := tspath.NormalizeSlashes(resolvedPath); got != test.wantPath {
				t.Fatalf("resolvedPath = %q, want %q", got, test.wantPath)
			}
		})
	}
}

// TestResolveFromSourceFileModuleSuffixes covers a specifier the Program never
// collected resolving under the compiler options it was built with. Probing the
// specifier against the loaded files would answer `./dep` with dep.ts, since a
// probe knows nothing of `moduleSuffixes`; TypeScript's resolver answers with
// dep.ios.ts, and that is the file the import actually names.
func TestResolveFromSourceFileModuleSuffixes(t *testing.T) {
	t.Parallel()

	files := map[string]string{
		"/suffix-fixture/dep.ts":      "export const dep = 1;\n",
		"/suffix-fixture/dep.ios.ts":  "export const dep = 2;\n",
		"/suffix-fixture/consumer.ts": `const dep = require("./dep");`,
	}

	program, sourceFile, specifier := programForRequire(t, files, "/suffix-fixture/consumer.ts", &core.CompilerOptions{
		Module:         core.ModuleKindCommonJS,
		ModuleSuffixes: []string{".ios", ""},
	})

	resolvedPath, _, ok := program.ResolveModule(sourceFile, specifier)
	if !ok {
		t.Fatal("ResolveFromSourceFile() did not resolve ./dep")
	}
	if got, want := tspath.NormalizeSlashes(resolvedPath), "/suffix-fixture/dep.ios.ts"; got != want {
		t.Fatalf("resolvedPath = %q, want %q", got, want)
	}
}

// TestResolveFromSourceFileUnloadedTarget covers a specifier TypeScript
// resolves to a file the Program never loaded. Probing the loaded files would
// answer `./dep` with dep.ts, since a probe knows nothing of `moduleSuffixes`;
// dep.ios.ts is the file the import names, and the probe is reserved for a
// specifier TypeScript resolves nowhere at all.
func TestResolveFromSourceFileUnloadedTarget(t *testing.T) {
	t.Parallel()

	const consumer = "/unloaded-fixture/consumer.ts"
	const loadedDep = "/unloaded-fixture/dep.ts"
	files := map[string]string{
		loadedDep:                      "export const dep = 1;\n",
		"/unloaded-fixture/dep.ios.ts": "export const dep = 2;\n",
		consumer:                       `const dep = require("./dep");`,
	}

	// dep.ios.ts stays out of the Program, the way `exclude` or a narrow `files`
	// list would keep it out, so the resolved file has no source file to report.
	program, sourceFile, specifier := programForRequireRoots(t, files, []string{consumer, loadedDep}, consumer, &core.CompilerOptions{
		Module:         core.ModuleKindCommonJS,
		ModuleSuffixes: []string{".ios", ""},
	})

	resolvedPath, target, ok := program.ResolveModule(sourceFile, specifier)
	if !ok {
		t.Fatal("ResolveFromSourceFile() did not resolve ./dep")
	}
	if got, want := tspath.NormalizeSlashes(resolvedPath), "/unloaded-fixture/dep.ios.ts"; got != want {
		t.Fatalf("resolvedPath = %q, want %q", got, want)
	}
	if target != nil {
		t.Fatalf("ResolveModule() target = %#v, want nil for a resolved file outside the Program", target)
	}
}

// TestResolveFromSourceFileParenthesizedRequireCondition covers a parenthesized
// `require` in an ES module. TypeScript reads a call as a `require` only through
// a bare `require` identifier, so it would answer this call with the file's own
// ESM format and select the package's `import` condition; the call is a
// `require` either way it is spelled, so the `require` condition is the one that
// holds.
func TestResolveFromSourceFileParenthesizedRequireCondition(t *testing.T) {
	t.Parallel()

	const consumer = "/condition-fixture/consumer.ts"
	const requireFile = "/condition-fixture/node_modules/some-package/cjs.d.cts"
	files := map[string]string{
		"/condition-fixture/package.json":                           `{"name": "root", "type": "module"}`,
		"/condition-fixture/node_modules/some-package/package.json": `{"name": "some-package", "exports": {".": {"import": "./esm.d.mts", "require": "./cjs.d.cts"}}}`,
		"/condition-fixture/node_modules/some-package/esm.d.mts":    "export const value: unknown;\n",
		requireFile: "declare const value: unknown;\nexport = value;\n",
		consumer:    `const pkg = (require)("some-package");`,
	}

	program, sourceFile, specifier := programForRequireRoots(t, files, []string{consumer}, consumer, &core.CompilerOptions{
		Module:           core.ModuleKindNodeNext,
		ModuleResolution: core.ModuleResolutionKindNodeNext,
	})

	resolvedPath, _, ok := program.ResolveModule(sourceFile, specifier)
	if !ok {
		t.Fatal("ResolveFromSourceFile() did not resolve some-package")
	}
	if got := tspath.NormalizeSlashes(resolvedPath); got != requireFile {
		t.Fatalf("resolvedPath = %q, want %q", got, requireFile)
	}
}

// contextForRequire parses fileName out of an in-memory tree and returns the
// argument of the first `require()` call in it.
func programForRequire(t *testing.T, files map[string]string, fileName string, options *core.CompilerOptions) (*lintprogram.Program, *ast.SourceFile, *ast.Node) {
	t.Helper()

	// Every file outside node_modules is a root, the way a tsconfig covering the
	// project would load them, so whichever file a specifier resolves to is one
	// the Program holds.
	rootFiles := make([]string, 0, len(files))
	for path := range files {
		if !strings.Contains(path, "/node_modules/") {
			rootFiles = append(rootFiles, path)
		}
	}
	slices.Sort(rootFiles)

	return programForRequireRoots(t, files, rootFiles, fileName, options)
}

// contextForRequireRoots parses fileName out of an in-memory tree the Program
// loads rootFiles from, and returns the argument of the first `require()` call
// in it.
func programForRequireRoots(t *testing.T, files map[string]string, rootFiles []string, fileName string, options *core.CompilerOptions) (*lintprogram.Program, *ast.SourceFile, *ast.Node) {
	t.Helper()

	fs := utils.NewOverlayVFS(bundled.WrapFS(osvfs.FS()), files)
	host := utils.CreateCompilerHost("/", fs)
	raw, err := utils.CreateProgramFromOptions(true, options, rootFiles, host)
	if err != nil {
		t.Fatalf("CreateProgramFromOptions: %v", err)
	}

	program := lintprogram.NewFromCompiler(raw)
	sourceFile := program.GetSourceFile(fileName)
	if sourceFile == nil {
		t.Fatalf("%s was not parsed", fileName)
		return nil, nil, nil
	}

	var specifier *ast.Node
	var visit func(node *ast.Node) bool
	visit = func(node *ast.Node) bool {
		if ast.IsCallExpression(node) {
			call := node.AsCallExpression()
			callee := ast.SkipParentheses(call.Expression)
			if ast.IsIdentifier(callee) && callee.Text() == "require" && len(call.Arguments.Nodes) == 1 {
				specifier = call.Arguments.Nodes[0]
				return true
			}
		}
		return node.ForEachChild(visit)
	}
	sourceFile.AsNode().ForEachChild(visit)

	if specifier == nil {
		t.Fatalf("%s has no require() call", fileName)
	}
	return program, sourceFile, specifier
}

// programForImport parses fileName out of an in-memory tree and returns the
// module specifier of the first import declaration in it.
func programForImport(t *testing.T, files map[string]string, fileName string) (*lintprogram.Program, *ast.SourceFile, *ast.Node) {
	t.Helper()

	rootFiles := make([]string, 0, len(files))
	for path := range files {
		rootFiles = append(rootFiles, path)
	}
	slices.Sort(rootFiles)

	fs := utils.NewOverlayVFS(bundled.WrapFS(osvfs.FS()), files)
	host := utils.CreateCompilerHost("/", fs)
	raw, err := utils.CreateProgramFromOptions(true, &core.CompilerOptions{}, rootFiles, host)
	if err != nil {
		t.Fatalf("CreateProgramFromOptions: %v", err)
	}

	program := lintprogram.NewFromCompiler(raw)
	sourceFile := program.GetSourceFile(fileName)
	if sourceFile == nil {
		t.Fatalf("%s was not parsed", fileName)
		return nil, nil, nil
	}
	for _, stmt := range sourceFile.Statements.Nodes {
		if stmt != nil && stmt.Kind == ast.KindImportDeclaration {
			return program, sourceFile, stmt.AsImportDeclaration().ModuleSpecifier
		}
	}
	t.Fatalf("%s has no import declaration", fileName)
	return nil, nil, nil
}
