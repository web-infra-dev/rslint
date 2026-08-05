package utils_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	import_utils "github.com/web-infra-dev/rslint/internal/plugins/import/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	rslint_utils "github.com/web-infra-dev/rslint/internal/utils"
)

func TestResolveSourceFileFromSourceFile(t *testing.T) {
	t.Parallel()

	ctx, specifier := contextForImport(t, "./bar")

	resolvedPath, target, ok := import_utils.ResolveSourceFileFromSourceFile(ctx, ctx.SourceFile, specifier)
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

	ctx, _ := contextForImport(t, "./bar")

	if resolvedPath, target, ok := import_utils.ResolveSourceFileFromSourceFile(ctx, ctx.SourceFile, nil); ok || resolvedPath != "" || target != nil {
		t.Fatalf("ResolveSourceFileFromSourceFile(nil) = (%q, %#v, %v), want empty result", resolvedPath, target, ok)
	}
}

// TestResolveFromSourceFileRequire covers the three `require()` shapes that
// differ in how their specifier reaches a file: TypeScript's own resolution
// records the call in a JavaScript file, while the same call in a TypeScript
// file reaches a relative target through the specifier probe and a package
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

			ctx, specifier := contextForRequire(t, files, test.fileName)

			resolvedPath, ok := import_utils.ResolveFromSourceFile(ctx, ctx.SourceFile, specifier)
			if ok != (test.wantPath != "") {
				t.Fatalf("ResolveFromSourceFile() ok = %v, want %v (path %q)", ok, test.wantPath != "", resolvedPath)
			}
			if got := tspath.NormalizeSlashes(resolvedPath); got != test.wantPath {
				t.Fatalf("resolvedPath = %q, want %q", got, test.wantPath)
			}
		})
	}
}

// contextForRequire parses fileName out of an in-memory tree and returns the
// argument of the first `require()` call in it.
func contextForRequire(t *testing.T, files map[string]string, fileName string) (rule.RuleContext, *ast.Node) {
	t.Helper()

	// Every file outside node_modules is a root, the way a tsconfig covering the
	// project would load them. The specifier probe only reaches loaded files.
	rootFiles := make([]string, 0, len(files))
	for path := range files {
		if !strings.Contains(path, "/node_modules/") {
			rootFiles = append(rootFiles, path)
		}
	}
	slices.Sort(rootFiles)

	fs := rslint_utils.NewOverlayVFS(bundled.WrapFS(osvfs.FS()), files)
	host := rslint_utils.CreateCompilerHost("/", fs)
	program, err := rslint_utils.CreateProgramFromOptions(true, &core.CompilerOptions{
		Module:  core.ModuleKindCommonJS,
		AllowJs: core.TSTrue,
	}, rootFiles, host)
	if err != nil {
		t.Fatalf("CreateProgramFromOptions: %v", err)
	}

	sourceFile := program.GetSourceFile(fileName)
	if sourceFile == nil {
		t.Fatalf("%s was not parsed", fileName)
		return rule.RuleContext{}, nil
	}

	var specifier *ast.Node
	var visit func(node *ast.Node) bool
	visit = func(node *ast.Node) bool {
		if ast.IsCallExpression(node) {
			call := node.AsCallExpression()
			if ast.IsIdentifier(call.Expression) && call.Expression.Text() == "require" && len(call.Arguments.Nodes) == 1 {
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
	return rule.RuleContext{Program: program, SourceFile: sourceFile}, specifier
}
