package utils_test

import (
	"strings"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	import_utils "github.com/web-infra-dev/rslint/internal/plugins/import/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	rslint_utils "github.com/web-infra-dev/rslint/internal/utils"
)

func TestHasExport(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		source     string
		exportName string
		wantFound  bool
		wantOK     bool
	}{
		{
			name:       "direct named export",
			source:     "./named-exports",
			exportName: "foo",
			wantFound:  true,
			wantOK:     true,
		},
		{
			name:       "direct TypeScript type export",
			source:     "./typescript",
			exportName: "MyType",
			wantFound:  true,
			wantOK:     true,
		},
		{
			name:       "direct TypeScript namespace export",
			source:     "./typescript",
			exportName: "MyNamespace",
			wantFound:  true,
			wantOK:     true,
		},
		{
			name:       "star export includes named exports",
			source:     "./re-export",
			exportName: "baz",
			wantFound:  true,
			wantOK:     true,
		},
		{
			name:       "multiple star exports continue after a miss",
			source:     "./multi-star-reexport",
			exportName: "baz",
			wantFound:  true,
			wantOK:     true,
		},
		{
			name:       "star export does not include default",
			source:     "./re-export",
			exportName: "default",
			wantFound:  false,
			wantOK:     true,
		},
		{
			name:       "explicit default re-export resolves deeply",
			source:     "./default-export-from",
			exportName: "default",
			wantFound:  true,
			wantOK:     true,
		},
		{
			name:       "TypeScript export assignment re-export is default-visible",
			source:     "./typescript-export-assign-default-reexport",
			exportName: "default",
			wantFound:  true,
			wantOK:     true,
		},
		{
			name:       "TypeScript export assignment property expression is default-visible",
			source:     "./typescript-export-assign-property",
			exportName: "default",
			wantFound:  true,
			wantOK:     true,
		},
		{
			name:       "explicit re-export validates remote local name",
			source:     "./reexport-missing-as-default",
			exportName: "default",
			wantFound:  false,
			wantOK:     true,
		},
		{
			name:       "explicit unresolved re-export is treated as present",
			source:     "./reexport-unresolved-as-default",
			exportName: "default",
			wantFound:  true,
			wantOK:     true,
		},
		{
			name:       "namespace export as default",
			source:     "./namespace-default",
			exportName: "default",
			wantFound:  true,
			wantOK:     true,
		},
		{
			name:       "commonjs target has no export map",
			source:     "./common",
			exportName: "default",
			wantFound:  false,
			wantOK:     false,
		},
		{
			name:       "cycle without local default terminates",
			source:     "./cycle-default-a",
			exportName: "default",
			wantFound:  false,
			wantOK:     true,
		},
		{
			name:       "cycle with local default wins before dependency walk",
			source:     "./cycle-with-local-default-a",
			exportName: "default",
			wantFound:  true,
			wantOK:     true,
		},
		{
			name:       "exported destructuring binding",
			source:     "./destructured-exports",
			exportName: "renamed",
			wantFound:  true,
			wantOK:     true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx, specifier := contextForImport(t, tc.source)
			gotFound, gotOK := import_utils.HasExport(ctx, specifier, tc.exportName)
			if gotFound != tc.wantFound || gotOK != tc.wantOK {
				t.Fatalf("HasExport(%q, %q) = (%v, %v), want (%v, %v)", tc.source, tc.exportName, gotFound, gotOK, tc.wantFound, tc.wantOK)
			}
		})
	}
}

func TestGetExportMap(t *testing.T) {
	t.Parallel()

	t.Run("direct exports and export-all namespace alias", func(t *testing.T) {
		t.Parallel()

		ctx, specifier := contextForImport(t, "./named-exports")
		exportMap, ok := import_utils.GetExportMap(ctx, specifier)
		if !ok {
			t.Fatal("GetExportMap returned no map")
		}
		for _, name := range []string{"a", "b", "d", "ExportedClass"} {
			if !exportMap.Has(name) {
				t.Fatalf("expected export %q to exist", name)
			}
		}
		deep := exportMap.Get("deep")
		if deep == nil {
			t.Fatal("expected export-all namespace alias to expose deep")
		}
		if deep.Namespace != nil {
			t.Fatal("expected export-all namespace alias not to carry namespace metadata")
		}
	})

	t.Run("star export excludes default", func(t *testing.T) {
		t.Parallel()

		ctx, specifier := contextForImport(t, "./re-export")
		exportMap, ok := import_utils.GetExportMap(ctx, specifier)
		if !ok {
			t.Fatal("GetExportMap returned no map")
		}
		if !exportMap.Has("baz") {
			t.Fatal("expected star export to include baz")
		}
		if exportMap.Has("default") {
			t.Fatal("expected star export not to include default")
		}
	})

	t.Run("nested export-all namespace aliases expose names without namespace metadata", func(t *testing.T) {
		t.Parallel()

		ctx, specifier := contextForImport(t, "./deep-namespace-chain/entry")
		exportMap, ok := import_utils.GetExportMap(ctx, specifier)
		if !ok {
			t.Fatal("GetExportMap returned no map")
		}
		b := exportMap.Get("b")
		if b == nil {
			t.Fatal("expected b namespace export alias")
		}
		if b.Namespace != nil {
			t.Fatal("expected export-all namespace alias not to carry namespace metadata")
		}
	})

	t.Run("unresolved star export keeps map open", func(t *testing.T) {
		t.Parallel()

		ctx, specifier := contextForImport(t, "./unresolved-star-export")
		exportMap, ok := import_utils.GetExportMap(ctx, specifier)
		if !ok {
			t.Fatal("GetExportMap returned no map")
		}
		if !exportMap.Has("anything") {
			t.Fatal("expected unresolved star export to make unknown names valid")
		}
		if exportMap.Get("anything") != nil {
			t.Fatal("expected unknown export to have no namespace metadata")
		}
	})

	t.Run("ambient namespace declaration exposes name without namespace metadata", func(t *testing.T) {
		t.Parallel()

		ctx, specifier := contextForImport(t, "./typescript-ambient-namespace")
		exportMap, ok := import_utils.GetExportMap(ctx, specifier)
		if !ok {
			t.Fatal("GetExportMap returned no map")
		}
		ambient := exportMap.Get("ambient")
		if ambient == nil {
			t.Fatal("expected ambient namespace export")
		}
		if ambient.Namespace != nil {
			t.Fatal("expected ambient namespace declaration not to carry namespace metadata")
		}
	})

	t.Run("string literal default export name", func(t *testing.T) {
		t.Parallel()

		ctx, specifier := contextForImport(t, "./default-export-string")
		exportMap, ok := import_utils.GetExportMap(ctx, specifier)
		if !ok {
			t.Fatal("GetExportMap returned no map")
		}
		if !exportMap.Has("default") {
			t.Fatal("expected string-literal default export to be visible")
		}
	})

	t.Run("string literal namespace export name exposes default without namespace metadata", func(t *testing.T) {
		t.Parallel()

		ctx, specifier := contextForImport(t, "./default-export-namespace-string")
		exportMap, ok := import_utils.GetExportMap(ctx, specifier)
		if !ok {
			t.Fatal("GetExportMap returned no map")
		}
		def := exportMap.Get("default")
		if def == nil {
			t.Fatal("expected string-literal namespace default to be visible")
		}
		if def.Namespace != nil {
			t.Fatal("expected string-literal namespace default not to carry namespace metadata")
		}
	})

	t.Run("default export of namespace import carries namespace metadata", func(t *testing.T) {
		t.Parallel()

		ctx, specifier := contextForImport(t, "./default-from-namespace-import")
		exportMap, ok := import_utils.GetExportMap(ctx, specifier)
		if !ok {
			t.Fatal("GetExportMap returned no map")
		}
		def := exportMap.Get("default")
		if def == nil || def.Namespace == nil || !def.Namespace.Has("a") {
			t.Fatal("expected default export of namespace import to expose a")
		}
	})

	t.Run("declaration default export of namespace import carries namespace metadata", func(t *testing.T) {
		t.Parallel()

		ctx, specifier := contextForImport(t, "./default-from-namespace-import-declaration")
		exportMap, ok := import_utils.GetExportMap(ctx, specifier)
		if !ok {
			t.Fatal("GetExportMap returned no map")
		}
		def := exportMap.Get("default")
		if def == nil || def.Namespace == nil || !def.Namespace.Has("a") {
			t.Fatal("expected declaration default export of namespace import to expose a")
		}
	})

	t.Run("default export of default import does not carry namespace metadata", func(t *testing.T) {
		t.Parallel()

		ctx, specifier := contextForImport(t, "./default-chain-from-namespace-import")
		exportMap, ok := import_utils.GetExportMap(ctx, specifier)
		if !ok {
			t.Fatal("GetExportMap returned no map")
		}
		def := exportMap.Get("default")
		if def == nil {
			t.Fatal("expected default export to be visible")
		}
		if def.Namespace != nil {
			t.Fatal("expected default-import re-export not to carry namespace metadata")
		}
	})

	t.Run("local export of namespace import carries namespace metadata", func(t *testing.T) {
		t.Parallel()

		ctx, specifier := contextForImport(t, "./namespace-import-local-reexport")
		exportMap, ok := import_utils.GetExportMap(ctx, specifier)
		if !ok {
			t.Fatal("GetExportMap returned no map")
		}
		forwarded := exportMap.Get("forwarded")
		if forwarded == nil || forwarded.Namespace == nil || !forwarded.Namespace.Has("a") {
			t.Fatal("expected local namespace import re-export to expose a")
		}
	})

	t.Run("source re-export preserves remote namespace metadata", func(t *testing.T) {
		t.Parallel()

		ctx, specifier := contextForImport(t, "./namespace-import-source-reexport")
		exportMap, ok := import_utils.GetExportMap(ctx, specifier)
		if !ok {
			t.Fatal("GetExportMap returned no map")
		}
		forwarded := exportMap.Get("forwardedAgain")
		if forwarded == nil || forwarded.Namespace == nil || !forwarded.Namespace.Has("a") {
			t.Fatal("expected source re-export to preserve namespace metadata")
		}
	})

	t.Run("local export of namespace-valued named import does not carry namespace metadata", func(t *testing.T) {
		t.Parallel()

		ctx, specifier := contextForImport(t, "./namespace-import-local-reexport-chain")
		exportMap, ok := import_utils.GetExportMap(ctx, specifier)
		if !ok {
			t.Fatal("GetExportMap returned no map")
		}
		forwarded := exportMap.Get("forwardedLocal")
		if forwarded == nil {
			t.Fatal("expected local re-export to be visible")
		}
		if forwarded.Namespace != nil {
			t.Fatal("expected named-import re-export not to carry namespace metadata")
		}
	})
}

func TestHasDefaultExportRespectsESModuleInterop(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		source            string
		esModuleInterop   core.Tristate
		wantDefaultExport bool
	}{
		{
			name:              "named TypeScript exports synthesize default when enabled",
			source:            "./typescript",
			esModuleInterop:   core.TSTrue,
			wantDefaultExport: true,
		},
		{
			name:              "named TypeScript exports do not synthesize default when disabled",
			source:            "./typescript",
			esModuleInterop:   core.TSFalse,
			wantDefaultExport: false,
		},
		{
			name:              "export equals namespace synthesizes default when enabled",
			source:            "./typescript-export-assign-default-namespace",
			esModuleInterop:   core.TSTrue,
			wantDefaultExport: true,
		},
		{
			name:              "export equals namespace does not synthesize default when disabled",
			source:            "./typescript-export-assign-default-namespace",
			esModuleInterop:   core.TSFalse,
			wantDefaultExport: false,
		},
		{
			name:              "export equals local variable is default-visible when disabled",
			source:            "./typescript-export-assign-local",
			esModuleInterop:   core.TSFalse,
			wantDefaultExport: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx, specifier := contextForImportWithCompilerOptions(t, tc.source, &core.CompilerOptions{
				ESModuleInterop: tc.esModuleInterop,
			})
			gotDefaultExport, gotOK := import_utils.HasDefaultExport(ctx, specifier)
			if gotDefaultExport != tc.wantDefaultExport || !gotOK {
				t.Fatalf("HasDefaultExport with esModuleInterop=%v = (%v, %v), want (%v, true)", tc.esModuleInterop, gotDefaultExport, gotOK, tc.wantDefaultExport)
			}
		})
	}
}

func TestHasExportRespectsImportIgnore(t *testing.T) {
	t.Parallel()

	ctx, specifier := contextForImport(t, "./ignored-missing-default")
	ctx.Settings = map[string]interface{}{
		"import/ignore": []interface{}{"ignored-missing-default"},
	}

	gotDefaultExport, gotOK := import_utils.HasDefaultExport(ctx, specifier)
	if gotDefaultExport || gotOK {
		t.Fatalf("HasDefaultExport for ignored import = (%v, %v), want (false, false)", gotDefaultExport, gotOK)
	}
}

func TestIsImportPathIgnored(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		settings map[string]interface{}
		fileName string
		want     bool
	}{
		{
			name:     "array of interface strings matches as regexp",
			settings: map[string]interface{}{"import/ignore": []interface{}{"ignored-missing-default"}},
			fileName: "/repo/ignored-missing-default.ts",
			want:     true,
		},
		{
			name:     "array of strings matches as regexp",
			settings: map[string]interface{}{"import/ignore": []string{`\.css$`}},
			fileName: "/repo/styles.css",
			want:     true,
		},
		{
			name:     "non-string entries and invalid regexps are ignored",
			settings: map[string]interface{}{"import/ignore": []interface{}{123, "["}},
			fileName: "/repo/ignored-missing-default.ts",
			want:     false,
		},
		{
			name:     "missing setting does not ignore",
			settings: map[string]interface{}{},
			fileName: "/repo/ignored-missing-default.ts",
			want:     false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := import_utils.IsImportPathIgnored(tc.settings, tc.fileName)
			if got != tc.want {
				t.Fatalf("IsImportPathIgnored() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestHasDefaultExportFollowsHostCaseSensitivity(t *testing.T) {
	t.Parallel()

	rootDir := fixtures.GetRootDir()
	consumerPath := tspath.ResolvePath(rootDir, "case-consumer.ts")
	actualTargetPath := tspath.ResolvePath(rootDir, "case-target.ts")
	virtualFiles := map[string]string{
		consumerPath:     `import value from "./Case-Target";`,
		actualTargetPath: `export const named = 1;`,
	}

	tests := []struct {
		name        string
		fs          vfs.FS
		wantFound   bool
		wantOK      bool
		wantResolve bool
	}{
		{
			name:        "case-sensitive host leaves mismatched import unresolved",
			fs:          rslint_utils.NewOverlayVFS(bundled.WrapFS(osvfs.FS()), virtualFiles),
			wantFound:   false,
			wantOK:      false,
			wantResolve: false,
		},
		{
			name:        "case-insensitive host resolves then checks exports",
			fs:          newCaseInsensitiveOverlayFS(virtualFiles),
			wantFound:   false,
			wantOK:      true,
			wantResolve: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx, specifier := contextForImportWithFS(t, tc.fs, consumerPath)
			resolved := ctx.Program.GetResolvedModuleFromModuleSpecifier(ctx.SourceFile, specifier)
			if (resolved != nil && resolved.ResolvedFileName != "") != tc.wantResolve {
				t.Fatalf("resolved = %#v, want resolved %v", resolved, tc.wantResolve)
			}

			gotFound, gotOK := import_utils.HasDefaultExport(ctx, specifier)
			if gotFound != tc.wantFound || gotOK != tc.wantOK {
				t.Fatalf("HasDefaultExport() = (%v, %v), want (%v, %v)", gotFound, gotOK, tc.wantFound, tc.wantOK)
			}
		})
	}
}

type caseInsensitiveOverlayFS struct {
	vfs.FS
	virtualFiles map[string]string
}

func newCaseInsensitiveOverlayFS(virtualFiles map[string]string) vfs.FS {
	base := rslint_utils.NewOverlayVFS(bundled.WrapFS(osvfs.FS()), virtualFiles)
	return &caseInsensitiveOverlayFS{
		FS:           base,
		virtualFiles: virtualFiles,
	}
}

func (fsys *caseInsensitiveOverlayFS) UseCaseSensitiveFileNames() bool {
	return false
}

func (fsys *caseInsensitiveOverlayFS) FileExists(path string) bool {
	if _, ok := fsys.findVirtualPath(path); ok {
		return true
	}
	return fsys.FS.FileExists(path)
}

func (fsys *caseInsensitiveOverlayFS) ReadFile(path string) (contents string, ok bool) {
	if actualPath, ok := fsys.findVirtualPath(path); ok {
		return fsys.virtualFiles[actualPath], true
	}
	return fsys.FS.ReadFile(path)
}

func (fsys *caseInsensitiveOverlayFS) Stat(path string) vfs.FileInfo {
	if actualPath, ok := fsys.findVirtualPath(path); ok {
		return fsys.FS.Stat(actualPath)
	}
	return fsys.FS.Stat(path)
}

func (fsys *caseInsensitiveOverlayFS) Realpath(path string) string {
	if actualPath, ok := fsys.findVirtualPath(path); ok {
		return actualPath
	}
	return fsys.FS.Realpath(path)
}

func (fsys *caseInsensitiveOverlayFS) findVirtualPath(path string) (string, bool) {
	for virtualPath := range fsys.virtualFiles {
		if strings.EqualFold(virtualPath, path) {
			return virtualPath, true
		}
	}
	return "", false
}

func contextForImport(t *testing.T, source string) (rule.RuleContext, *ast.Node) {
	t.Helper()

	rootDir := fixtures.GetRootDir()
	fileName := "file.ts"
	code := `import value from "` + source + `";`
	fs := rslint_utils.NewOverlayVFSForFile(tspath.ResolvePath(rootDir, fileName), code)
	host := rslint_utils.CreateCompilerHost(rootDir, fs)
	program, err := rslint_utils.CreateProgram(true, fs, rootDir, "tsconfig.json", host)
	if err != nil {
		t.Fatalf("CreateProgram: %v", err)
	}

	sourceFile := program.GetSourceFile(fileName)
	if sourceFile == nil || sourceFile.Statements == nil || len(sourceFile.Statements.Nodes) == 0 {
		t.Fatal("test source file was not parsed")
	}
	importDecl := sourceFile.Statements.Nodes[0].AsImportDeclaration()
	if importDecl == nil || importDecl.ModuleSpecifier == nil {
		t.Fatal("test import declaration was not parsed")
	}

	return rule.RuleContext{
		Program:    program,
		SourceFile: sourceFile,
	}, importDecl.ModuleSpecifier
}

func contextForImportWithFS(t *testing.T, fs vfs.FS, filePath string) (rule.RuleContext, *ast.Node) {
	t.Helper()

	host := rslint_utils.CreateCompilerHost(fixtures.GetRootDir(), fs)
	program, err := rslint_utils.CreateProgramFromOptions(true, &core.CompilerOptions{
		ESModuleInterop: core.TSFalse,
		Module:          core.ModuleKindCommonJS,
	}, []string{filePath}, host)
	if err != nil {
		t.Fatalf("CreateProgramFromOptions: %v", err)
	}

	sourceFile := program.GetSourceFile(filePath)
	if sourceFile == nil || sourceFile.Statements == nil || len(sourceFile.Statements.Nodes) == 0 {
		t.Fatal("test source file was not parsed")
	}
	importDecl := sourceFile.Statements.Nodes[0].AsImportDeclaration()
	if importDecl == nil || importDecl.ModuleSpecifier == nil {
		t.Fatal("test import declaration was not parsed")
	}

	return rule.RuleContext{
		Program:    program,
		SourceFile: sourceFile,
	}, importDecl.ModuleSpecifier
}

func contextForImportWithCompilerOptions(t *testing.T, source string, options *core.CompilerOptions) (rule.RuleContext, *ast.Node) {
	t.Helper()

	rootDir := fixtures.GetRootDir()
	fileName := "file.ts"
	filePath := tspath.ResolvePath(rootDir, fileName)
	code := `import value from "` + source + `";`
	fs := rslint_utils.NewOverlayVFSForFile(filePath, code)
	host := rslint_utils.CreateCompilerHost(rootDir, fs)
	program, err := rslint_utils.CreateProgramFromOptions(true, options, []string{filePath}, host)
	if err != nil {
		t.Fatalf("CreateProgramFromOptions: %v", err)
	}

	sourceFile := program.GetSourceFile(filePath)
	if sourceFile == nil || sourceFile.Statements == nil || len(sourceFile.Statements.Nodes) == 0 {
		t.Fatal("test source file was not parsed")
	}
	importDecl := sourceFile.Statements.Nodes[0].AsImportDeclaration()
	if importDecl == nil || importDecl.ModuleSpecifier == nil {
		t.Fatal("test import declaration was not parsed")
	}

	return rule.RuleContext{
		Program:    program,
		SourceFile: sourceFile,
	}, importDecl.ModuleSpecifier
}
