package program_test

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"weak"

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

func TestRootProgramRecreatesCollectedASTsFromFrozenSources(t *testing.T) {
	for _, root := range []string{"/root-snapshot", "C:/root-snapshot", "//server/share/root-snapshot"} {
		t.Run(root, func(t *testing.T) {
			first := tspath.ResolvePath(root, "first.ts")
			second := tspath.ResolvePath(root, "second.ts")
			const original = "import './second'; export const first = 1;"
			contents := map[string]string{first: original, second: "export const second = 2;"}
			// Virtual drive and UNC paths must never reach the host filesystem.
			fs := utils.NewOverlayVFS(rootSnapshotEmptyTestFS{}, contents)
			host := &rootSnapshotTestHost{CompilerHost: utils.CreateCompilerHost(root, fs)}
			roots := []string{first, second, first}
			p, err := lintprogram.NewFromRoots(lintprogram.RootOptions{
				RootFileNames: roots, Host: host, CompilerOptions: lintprogram.SourceOnlyCompilerOptions(),
			})
			if err != nil {
				t.Fatal(err)
			}
			roots[0] = "changed.ts"
			source, found := p.LookupSource(first)
			_, absent := p.LookupSource("absent.ts")
			if !found || source.FileName() != first || absent || len(p.RootFileNames()) != 2 {
				t.Fatal("source membership lost its frozen normalized identities")
			}
			dead := rootSnapshotWeakFile(p, first)
			runtime.GC()
			runtime.GC()
			if dead.Value() != nil {
				t.Fatal("Program retains an unreferenced root AST")
			}
			contents[first] = "changed on disk"
			delete(contents, second)
			file := p.GetSourceFile(first)
			if file.Text() != original || !file.IsBound() || !p.OwnsSourceFile(file) || p.CanProvideTypeChecker(file) {
				t.Fatal("materialization changed the frozen source generation or its capabilities")
			}
			if host.parses.Load() != 2 {
				t.Fatal("materialization reread the host instead of the frozen source")
			}
			foreign := parser.ParseSourceFile(file.ParseOptions(), file.Text(), file.ScriptKind)
			if p.OwnsSourceFile(foreign) {
				t.Fatal("Program accepted a foreign AST of the same file")
			}
			files := p.SourceFiles()
			if len(files) != 2 || files[0] != file || files[1].FileName() != second || !p.OwnsSourceFile(files[1]) {
				t.Fatal("materialization reduced the complete source universe")
			}
			resolved := p.GetResolvedModuleFromModuleSpecifier(file, file.Imports()[0])
			if resolved == nil || p.GetSourceFileForResolvedModule(resolved.ResolvedFileName) != files[1] {
				t.Fatal("collection changed the generation's import resolution")
			}
			runtime.GC()
			if p.GetSourceFile(first) != file || !p.OwnsSourceFile(file) {
				t.Fatal("a borrowed AST lost its identity across collection")
			}
		})
	}
}

func TestRootProgramConcurrentMaterializationSharesLiveIdentity(t *testing.T) {
	const root = "/root-concurrent"
	name := tspath.ResolvePath(root, "file.tsx")
	p, err := lintprogram.NewFromRoots(lintprogram.RootOptions{
		RootFileNames: []string{name}, CompilerOptions: lintprogram.SourceOnlyCompilerOptions(),
		Host: utils.CreateCompilerHost(root, utils.NewOverlayVFS(rootSnapshotEmptyTestFS{}, map[string]string{
			name: "export const view = <div />;",
		})),
	})
	if err != nil {
		t.Fatal(err)
	}
	dead := rootSnapshotWeakFile(p, name)
	runtime.GC()
	runtime.GC()
	if dead.Value() != nil {
		t.Fatal("fixture still retains its original AST")
	}
	files := make([]*ast.SourceFile, 32)
	var workers sync.WaitGroup
	for i := range files {
		workers.Go(func() { files[i] = p.GetSourceFile(name) })
	}
	workers.Wait()
	for _, file := range files {
		if file != files[0] || !p.OwnsSourceFile(file) || !file.IsBound() || file.ScriptKind != core.ScriptKindTSX {
			t.Fatal("concurrent materialization published inconsistent ASTs")
		}
	}
}

func rootSnapshotWeakFile(p *lintprogram.Program, name string) weak.Pointer[ast.SourceFile] {
	return weak.Make(p.GetSourceFile(name))
}

type rootSnapshotTestHost struct {
	compiler.CompilerHost
	parses atomic.Int32
}

func (h *rootSnapshotTestHost) GetSourceFile(options ast.SourceFileParseOptions) *ast.SourceFile {
	h.parses.Add(1)
	return h.CompilerHost.GetSourceFile(options)
}

type rootSnapshotEmptyTestFS struct{ vfs.FS }

func (rootSnapshotEmptyTestFS) UseCaseSensitiveFileNames() bool { return true }
func (rootSnapshotEmptyTestFS) FileExists(string) bool          { return false }
func (rootSnapshotEmptyTestFS) ReadFile(string) (string, bool)  { return "", false }
func (rootSnapshotEmptyTestFS) DirectoryExists(string) bool     { return false }
func (rootSnapshotEmptyTestFS) Realpath(path string) string     { return path }
