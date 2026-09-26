package program_test

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/binder"
	"github.com/microsoft/TypeScript/tsc/shim/bundled"
	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestSourceOnlyProgramOwnsBoundUniverseWithoutCheckerCapability(t *testing.T) {
	const root = "/program-source-only-test"
	files := map[string]string{
		tspath.ResolvePath(root, "a.ts"): `import "./b"; export const a = 1;`,
		tspath.ResolvePath(root, "b.ts"): `import "./a"; export const b = 1;`,
	}
	fs := utils.NewOverlayVFS(bundled.WrapFS(osvfs.FS()), files)
	host := utils.CreateCompilerHost(root, fs)
	raw, err := utils.CreateProgramFromOptions(true, &core.CompilerOptions{
		Module: core.ModuleKindESNext,
	}, []string{tspath.ResolvePath(root, "a.ts"), tspath.ResolvePath(root, "b.ts")}, host)
	if err != nil {
		t.Fatalf("CreateProgramFromOptions: %v", err)
	}

	a := raw.GetSourceFile(tspath.ResolvePath(root, "a.ts"))
	b := raw.GetSourceFile(tspath.ResolvePath(root, "b.ts"))
	if a == nil || b == nil {
		t.Fatal("fixture Program did not contain both roots")
	}
	sourceProgram, err := lintprogram.NewFromBoundSources(raw, []*ast.SourceFile{nil, a, a, b})
	if err != nil {
		t.Fatalf("NewFromBoundSources: %v", err)
	}
	if !sourceProgram.IsValid() || sourceProgram.CanProvideTypeChecker(a) {
		t.Fatal("source-only Program exposed a checker capability")
	}
	for _, file := range sourceProgram.SourceFiles() {
		if !file.IsBound() || !sourceProgram.OwnsSourceFile(file) || sourceProgram.GetSourceFile(file.FileName()) != file {
			t.Fatalf("source Program lost exact source identity for %q", file.FileName())
		}
	}
	if diagnostics := sourceProgram.SyntacticDiagnostics(context.Background(), sourceProgram.SourceFiles()[0]); len(diagnostics) != 0 {
		t.Fatalf("unexpected syntactic diagnostics: %+v", diagnostics)
	}
	foreign := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: a.FileName(),
		Path:     a.Path(),
	}, a.Text(), core.ScriptKindTS)
	importSpecifier := a.Imports()[0]
	mode := sourceProgram.GetModeForUsageLocation(a, importSpecifier)
	if sourceProgram.GetResolvedModule(a, importSpecifier.Text(), mode) == nil {
		t.Fatal("owned source lost its cached module resolution")
	}
	if sourceProgram.OwnsSourceFile(foreign) ||
		sourceProgram.SourceFileMetadata(foreign) != (ast.SourceFileMetaData{}) ||
		sourceProgram.SyntacticDiagnostics(context.Background(), foreign) != nil ||
		sourceProgram.GetResolvedModule(foreign, importSpecifier.Text(), mode) != nil {
		t.Fatal("file-scoped facade methods accepted a foreign AST generation")
	}
}

func TestSourceOnlyProgramRejectsInvalidSourceGeneration(t *testing.T) {
	const root = "/program-source-only-invalid-test"
	fileName := tspath.ResolvePath(root, "a.ts")
	fs := utils.NewOverlayVFS(bundled.WrapFS(osvfs.FS()), map[string]string{
		fileName: "export const a = 1;",
	})
	host := utils.CreateCompilerHost(root, fs)
	raw, err := utils.CreateProgramFromOptions(true, &core.CompilerOptions{}, []string{fileName}, host)
	if err != nil {
		t.Fatalf("CreateProgramFromOptions: %v", err)
	}
	owned := raw.GetSourceFile(fileName)
	if owned == nil {
		t.Fatal("fixture Program did not contain a.ts")
	}
	reparsed := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: owned.FileName(),
		Path:     owned.Path(),
	}, owned.Text(), core.ScriptKindTS)
	if _, err := lintprogram.NewFromBoundSources(raw, []*ast.SourceFile{reparsed}); err == nil ||
		err.Error() != "program: source \"/program-source-only-invalid-test/a.ts\" is not bound" {
		t.Fatalf("unbound source error = %v", err)
	}
	binder.BindSourceFile(reparsed)
	if _, err := lintprogram.NewFromBoundSources(raw, []*ast.SourceFile{reparsed}); err == nil ||
		err.Error() != "program: source services do not own source \"/program-source-only-invalid-test/a.ts\"" {
		t.Fatalf("foreign source generation error = %v", err)
	}
	if _, err := lintprogram.NewFromBoundSources(raw, []*ast.SourceFile{reparsed, owned}); err == nil ||
		err.Error() != fmt.Sprintf("program: source universe contains different ASTs for path %q", owned.Path()) {
		t.Fatalf("same-path source conflict error = %v", err)
	}
}

func TestRootProgramRejectsMissingHost(t *testing.T) {
	var host *typedNilCompilerHost
	_, err := lintprogram.NewFromRoots(lintprogram.RootOptions{
		Host:            host,
		CompilerOptions: &core.CompilerOptions{},
	})
	if err == nil || err.Error() != "program: root construction requires a compiler host" {
		t.Fatalf("typed-nil host error = %v", err)
	}
	_, err = lintprogram.NewFromRoots(lintprogram.RootOptions{
		Host:            &typedNilFSCompilerHost{},
		CompilerOptions: &core.CompilerOptions{},
	})
	if err == nil || err.Error() != "program: root construction requires a filesystem" {
		t.Fatalf("typed-nil filesystem error = %v", err)
	}
}

func TestRootProgramRejectsCaseFoldedRootCollision(t *testing.T) {
	const root = "/program-case-fold-test"
	fs := caseInsensitiveFS{FS: bundled.WrapFS(osvfs.FS())}
	host := utils.CreateCompilerHost(root, fs)
	_, err := lintprogram.NewFromRoots(lintprogram.RootOptions{
		RootFileNames: []string{
			tspath.ResolvePath(root, "Pkg.ts"),
			tspath.ResolvePath(root, "pkg.ts"),
		},
		Host:            host,
		CompilerOptions: &core.CompilerOptions{},
	})
	if err == nil || !strings.Contains(err.Error(), "have the same path identity") {
		t.Fatalf("case-folded root collision error = %v", err)
	}
}

func TestRootProgramCachesSyntacticDiagnosticsDuringConstruction(t *testing.T) {
	const root = "/program-root-syntax-test"
	fileName := tspath.ResolvePath(root, "invalid.js")
	fs := utils.NewOverlayVFS(bundled.WrapFS(osvfs.FS()), map[string]string{
		fileName: "const value = ;",
	})
	sourceProgram, err := lintprogram.NewFromRoots(lintprogram.RootOptions{
		RootFileNames:   []string{fileName},
		Host:            utils.CreateCompilerHost(root, fs),
		CompilerOptions: &core.CompilerOptions{AllowJs: core.TSTrue},
	})
	if err != nil {
		t.Fatalf("NewFromRoots: %v", err)
	}
	file := sourceProgram.SourceFiles()[0]
	cached := sourceProgram.SyntacticDiagnostics(context.Background(), file)
	if len(cached) == 0 {
		t.Fatal("root construction did not cache parser diagnostics")
	}
	got := sourceProgram.SyntacticDiagnostics(context.Background(), file)
	if len(got) != len(cached) || &got[0] != &cached[0] {
		t.Fatal("SyntacticDiagnostics did not reuse immutable construction output")
	}
}

func TestSourceSetDefersParsingAndPreservesPlatformPaths(t *testing.T) {
	for _, root := range []string{"/source-set", "C:/source-set", "//server/share/source-set"} {
		t.Run(root, func(t *testing.T) {
			first := tspath.ResolvePath(root, "first.ts")
			second := tspath.ResolvePath(root, "second.ts")
			// Virtual drive and UNC paths must never reach the host filesystem.
			fs := utils.NewOverlayVFS(sourceSetEmptyTestFS{}, map[string]string{
				first: "import './second'; export const first = 1;", second: "export const second = 2;",
			})
			host := &sourceSetTestHost{CompilerHost: utils.CreateCompilerHost(root, fs)}
			roots := []string{first, second, first}
			sources, err := lintprogram.NewSourceSet(lintprogram.RootOptions{
				RootFileNames: roots, Host: host, CompilerOptions: lintprogram.SourceOnlyCompilerOptions(),
			})
			if err != nil {
				t.Fatal(err)
			}
			roots[0] = "changed.ts"
			if host.parses.Load() != 0 || len(sources.FileNames()) != 2 || sources.FileNames()[0] != first {
				t.Fatal("source set parsed eagerly or retained the caller's root slice")
			}
			one, err := sources.BuildFile(0)
			if err != nil || host.parses.Load() != 1 || len(one.SourceFiles()) != 1 || one.GetSourceFile(second) != nil {
				t.Fatalf("one-file construction: parses=%d error=%v", host.parses.Load(), err)
			}
			file := one.SourceFiles()[0]
			if !file.IsBound() || file.FileName() != first || one.CanProvideTypeChecker(file) {
				t.Fatal("one-file construction changed source identity or capability")
			}
			all, err := sources.Build()
			if err != nil || len(all.SourceFiles()) != 2 || all.GetSourceFile(second) == nil {
				t.Fatalf("whole-set construction: %v", err)
			}
			if all.GetSourceFile(first) == file {
				t.Fatal("independent builds shared a mutable bound AST")
			}
			if _, err := sources.BuildFile(2); err == nil {
				t.Fatal("out-of-range source index was accepted")
			}
		})
	}
}

func TestSourceSetRejectsInvalidInputsBeforeParsing(t *testing.T) {
	var nilHost *typedNilCompilerHost
	for _, options := range []lintprogram.RootOptions{
		{Host: nilHost, CompilerOptions: lintprogram.SourceOnlyCompilerOptions()},
		{Host: &typedNilFSCompilerHost{}, CompilerOptions: lintprogram.SourceOnlyCompilerOptions()},
		{Host: utils.CreateCompilerHost("/source-set", osvfs.FS())},
		{
			RootFileNames:   []string{"C:/source-set/Case.ts", "c:/source-set/case.ts"},
			Host:            utils.CreateCompilerHost("C:/source-set", caseInsensitiveFS{FS: osvfs.FS()}),
			CompilerOptions: lintprogram.SourceOnlyCompilerOptions(),
		},
	} {
		if _, err := lintprogram.NewSourceSet(options); err == nil {
			t.Fatalf("invalid source set accepted: %+v", options)
		}
	}
	var empty lintprogram.SourceSet
	if empty.IsValid() || empty.FileNames() != nil {
		t.Fatal("zero source set is valid")
	}
	if _, err := empty.Build(); err == nil {
		t.Fatal("zero source set was built")
	}
}

type sourceSetTestHost struct {
	compiler.CompilerHost
	parses atomic.Int32
}

// Overlay misses are absent, never reads against a real drive or UNC share.
// Unused filesystem operations deliberately have no implementation.
type sourceSetEmptyTestFS struct{ vfs.FS }

func (sourceSetEmptyTestFS) UseCaseSensitiveFileNames() bool { return true }
func (sourceSetEmptyTestFS) FileExists(string) bool          { return false }
func (sourceSetEmptyTestFS) ReadFile(string) (string, bool)  { return "", false }
func (sourceSetEmptyTestFS) DirectoryExists(string) bool     { return false }
func (sourceSetEmptyTestFS) Realpath(path string) string     { return path }

func (h *sourceSetTestHost) GetSourceFile(options ast.SourceFileParseOptions) *ast.SourceFile {
	h.parses.Add(1)
	return h.CompilerHost.GetSourceFile(options)
}

type typedNilCompilerHost struct {
	compiler.CompilerHost
}

type caseInsensitiveFS struct {
	vfs.FS
}

func (caseInsensitiveFS) UseCaseSensitiveFileNames() bool { return false }

type typedNilFS struct {
	vfs.FS
}

type typedNilFSCompilerHost struct {
	compiler.CompilerHost
}

func (*typedNilFSCompilerHost) FS() vfs.FS {
	var fs *typedNilFS
	return fs
}
