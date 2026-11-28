package program_test

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
	"weak"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/binder"
	"github.com/microsoft/TypeScript/tsc/shim/bundled"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestOwnedCompilerProgramsShareCacheOnlyWithinBatch(t *testing.T) {
	raw, _ := specifierCacheProgram(t, specifierCacheFiles(false))
	other, _ := specifierCacheProgram(t, specifierCacheFiles(true))
	inputs := []*compiler.Program{nil, raw, raw, other, nil}
	programs := lintprogram.NewOwnedFromCompilers(inputs)
	if len(programs) != len(inputs) {
		t.Fatalf("got %d facades, want %d", len(programs), len(inputs))
	}
	for index, input := range inputs {
		if input == nil {
			if programs[index] != nil {
				t.Fatalf("nil input %d became a valid facade", index)
			}
			continue
		}
		if programs[index].GetSourceFile(specifierCacheImporter) != specifierCacheFile(t, input) {
			t.Fatalf("facade %d changed source generation or input order", index)
		}
	}
	if programs[1] == programs[2] {
		t.Fatal("duplicate inputs must retain distinct facade identities")
	}
	if !programs[1].OwnsSourceFile(specifierCacheFile(t, raw)) || programs[1].OwnsSourceFile(specifierCacheFile(t, other)) {
		t.Fatal("source ownership must compare exact SourceFile objects, even at the same path")
	}

	type key struct{}
	var builds atomic.Int32
	build := func() *int { value := int(builds.Add(1)); return &value }
	var first *int
	for range 2 {
		var group sync.WaitGroup
		answers := make([]*int, 32)
		for index := range answers {
			group.Go(func() {
				answers[index] = lintprogram.Cached(programs[1+index%2], key{}, build)
			})
		}
		group.Wait()
		if first == nil {
			first = answers[0]
		}
		for _, answer := range answers {
			if answer != first {
				t.Fatal("duplicate raw inputs published different cache values")
			}
		}
		runtime.GC()
	}
	if builds.Load() != 1 {
		t.Fatalf("built %d times across concurrent requests and GC, want 1", builds.Load())
	}
	if answer := lintprogram.Cached(programs[3], key{}, build); answer == first {
		t.Fatal("different raw generations shared their cache")
	}
	separate := lintprogram.NewOwnedFromCompilers([]*compiler.Program{raw})
	if answer := lintprogram.Cached(separate[0], key{}, build); answer == first {
		t.Fatal("separate ownership batches shared their cache")
	}
	registered := lintprogram.Cached(lintprogram.NewFromCompiler(raw), key{}, build)
	if registered == first || lintprogram.Cached(lintprogram.NewFromCompiler(raw), key{}, build) != registered {
		t.Fatal("owned cache polluted the raw generation's shared cache")
	}
}

func TestOwnedCompilerCacheReleasesWhileRawGenerationLives(t *testing.T) {
	raw, _ := specifierCacheProgram(t, specifierCacheFiles(false))
	type key int
	type payload [128]byte
	var cached weak.Pointer[payload]
	var facade weak.Pointer[lintprogram.Program]
	var retained []string
	func() {
		program := lintprogram.NewOwnedFromCompilers([]*compiler.Program{raw})[0]
		facade = weak.Make(program)
		cached = weak.Make(lintprogram.Cached(program, key(0), func() *payload { return &payload{} }))
		retained = lintprogram.Cached(program, key(1), func() []string { return []string{"immutable answer"} })
	}()
	waitForProgramCacheCollection(t, func() bool { return facade.Value() == nil && cached.Value() == nil })
	if len(retained) != 1 || retained[0] != "immutable answer" {
		t.Fatal("releasing a cache invalidated its independently retained answer")
	}
	if raw.GetSourceFile(specifierCacheImporter) == nil {
		t.Fatal("raw generation was lost before its owned cache")
	}
	runtime.KeepAlive(raw)
}

func TestCompilerCachePreservesAnswerAcrossFacadeGap(t *testing.T) {
	files := specifierCacheFiles(false)
	raw, fsys := specifierCacheProgram(t, files)
	type key struct{}
	builds := 0
	build := func() string {
		builds++
		text, ok := fsys.ReadFile(specifierCacheImporter)
		if !ok {
			t.Fatal("fixture source is missing")
		}
		return text
	}
	var facade weak.Pointer[lintprogram.Program]
	func() {
		program := lintprogram.NewFromCompiler(raw)
		facade = weak.Make(program)
		if got := lintprogram.Cached(program, key{}, build); got != files[specifierCacheImporter] {
			t.Fatal("initial cache answer did not reflect the source view")
		}
	}()
	files[specifierCacheImporter] = "export const changed = true;"
	waitForProgramCacheCollection(t, func() bool { return facade.Value() == nil })
	if got := lintprogram.Cached(lintprogram.NewFromCompiler(raw), key{}, build); got == files[specifierCacheImporter] || builds != 1 {
		t.Fatal("a temporary facade gap changed the raw generation's cached answer")
	}
	runtime.KeepAlive(raw)
}

func waitForProgramCacheCollection(t *testing.T, collected func() bool) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		runtime.GC()
		if collected() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("unreachable facade or derived cache was retained")
}

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

type singleSlotExternalCheckerPool struct {
	program *compiler.Program
	sem     chan struct{}
	once    sync.Once
	checker *checker.Checker
}

func (p *singleSlotExternalCheckerPool) GetChecker(ctx context.Context, _ *ast.SourceFile) (*checker.Checker, func()) {
	p.once.Do(func() {
		p.checker, _ = checker.NewChecker(p.program, nil)
	})
	select {
	case p.sem <- struct{}{}:
	case <-ctx.Done():
		panic("deadlock waiting for single-slot external checker pool")
	}
	return p.checker, sync.OnceFunc(func() { <-p.sem })
}

func (p *singleSlotExternalCheckerPool) GetGlobalDiagnostics() []*ast.Diagnostic {
	return nil
}

type requestIDCheckingPool struct {
	program        *compiler.Program
	sem            chan struct{}
	once           sync.Once
	checker        *checker.Checker
	mu             sync.Mutex
	nonEmptyReqIDs []string
}

func (p *requestIDCheckingPool) GetChecker(ctx context.Context, _ *ast.SourceFile) (*checker.Checker, func()) {
	if id := core.GetRequestID(ctx); id != "" {
		p.mu.Lock()
		p.nonEmptyReqIDs = append(p.nonEmptyReqIDs, id)
		p.mu.Unlock()
	}
	p.once.Do(func() {
		p.checker, _ = checker.NewChecker(p.program, nil)
	})
	select {
	case p.sem <- struct{}{}:
	case <-ctx.Done():
		panic("deadlock waiting for external checker pool")
	}
	return p.checker, sync.OnceFunc(func() { <-p.sem })
}

func TestCompilerQueriesDeclarationDiagnosticsDoesNotDeadlockWithSingleSlotPool(t *testing.T) {
	const root = "/program-decl-pool-test"
	fileA := tspath.ResolvePath(root, "a.ts")
	fileB := tspath.ResolvePath(root, "b.ts")
	fs := utils.NewOverlayVFS(bundled.WrapFS(osvfs.FS()), map[string]string{
		fileA: "export const answer = 42;\n",
		fileB: "import { answer } from './a';\nexport const doubled = answer * 2;\n",
	})
	host := utils.CreateCompilerHost(root, fs)
	base, err := utils.CreateProgramFromOptions(true, &core.CompilerOptions{
		Declaration: core.TSTrue,
		NoLib:       core.TSTrue,
	}, []string{fileA, fileB}, host)
	if err != nil {
		t.Fatalf("CreateProgramFromOptions: %v", err)
	}

	t.Run("default compiler checker pool", func(t *testing.T) {
		p := lintprogram.NewFromCompiler(base)
		ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
		defer cancel()
		file := p.GetSourceFile(fileA)
		if file == nil {
			t.Fatal("missing source file")
		}
		// Must not self-deadlock against EmitResolver's internal checker mutex.
		_ = p.CompilerQueries().Diagnostics(ctx, lintprogram.DeclarationDiagnostics, file)
		_ = p.CompilerQueries().Diagnostics(ctx, lintprogram.DeclarationDiagnostics, nil)

		// Non-program files (JSON and TS) must return nil without panicking on
		// either DeclarationDiagnostics or SemanticDiagnostics.
		foreignJSON := parser.ParseSourceFile(ast.SourceFileParseOptions{
			FileName: tspath.ResolvePath(root, "tsconfig.json"),
			Path:     tspath.Path(tspath.ResolvePath(root, "tsconfig.json")),
		}, "{}", core.ScriptKindJSON)
		if diags := p.CompilerQueries().Diagnostics(ctx, lintprogram.DeclarationDiagnostics, foreignJSON); len(diags) != 0 {
			t.Fatalf("expected nil declaration diagnostics for non-program JSON file, got %d", len(diags))
		}
		foreignTS := parser.ParseSourceFile(ast.SourceFileParseOptions{
			FileName: tspath.ResolvePath(root, "foreign.ts"),
			Path:     tspath.Path(tspath.ResolvePath(root, "foreign.ts")),
		}, "export const x = 1;\n", core.ScriptKindTS)
		if diags := p.CompilerQueries().Diagnostics(ctx, lintprogram.SemanticDiagnostics, foreignTS); len(diags) != 0 {
			t.Fatalf("expected nil semantic diagnostics for non-program TS file, got %d", len(diags))
		}
	})

	t.Run("isolatedDeclarations concurrent checker and declaration diagnostics", func(t *testing.T) {
		// Multi-threaded program with isolatedDeclarations exercises tsgo's
		// unlocked EmitResolver.*Unsafe paths (function property assignments +
		// inference fallbacks) concurrently with exclusive checker queries under -race.
		isoRoot := "/program-isolated-decl-test"
		isoFiles := map[string]string{}
		var isoNames []string
		for i := range 6 {
			name := tspath.ResolvePath(isoRoot, fmt.Sprintf("f%d.ts", i))
			isoNames = append(isoNames, name)
			isoFiles[name] = fmt.Sprintf(
				"export function foo%d() {}\nfoo%d.bar = 1;\nexport function f%d(x = 1 + 2) { return x; }\n",
				i, i, i,
			)
		}
		isoFS := utils.NewOverlayVFS(bundled.WrapFS(osvfs.FS()), isoFiles)
		isoHost := utils.CreateCompilerHost(isoRoot, isoFS)
		isoRaw, err := utils.CreateProgramFromOptions(false, &core.CompilerOptions{
			Declaration:          core.TSTrue,
			IsolatedDeclarations: core.TSTrue,
			NoLib:                core.TSTrue,
		}, isoNames, isoHost)
		if err != nil {
			t.Fatalf("CreateProgramFromOptions(isolatedDeclarations): %v", err)
		}
		p := lintprogram.NewFromCompiler(isoRaw)
		ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
		defer cancel()
		firstFile := p.GetSourceFile(isoNames[0])
		if firstFile == nil {
			t.Fatal("missing isolatedDeclarations source file")
		}
		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			for _, sf := range p.SourceFiles() {
				c, release := p.TypeCheckerForFileExclusive(ctx, sf)
				for _, stmt := range sf.Statements.Nodes {
					_ = c.GetTypeAtLocation(stmt)
				}
				release()
			}
		}()
		go func() {
			defer wg.Done()
			// Uncached program-wide declaration diagnostics concurrently with
			// active TypeCheckerForFileExclusive queries.
			_ = p.CompilerQueries().Diagnostics(ctx, lintprogram.DeclarationDiagnostics, nil)
		}()
		wg.Wait()
	})

	t.Run("single-slot external pool", func(t *testing.T) {
		raw := compiler.NewProgram(compiler.ProgramOptions{
			Config: base.CommandLine(),
			Host:   host,
			CreateCheckerPool: func(prog *compiler.Program) compiler.CheckerPool {
				return &singleSlotExternalCheckerPool{program: prog, sem: make(chan struct{}, 1)}
			},
		})
		p := lintprogram.NewFromCompiler(raw)
		ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
		defer cancel()
		file := p.GetSourceFile(fileA)
		if file == nil {
			t.Fatal("missing source file")
		}
		// Must not self-deadlock on single-slot external CheckerPool that also
		// implements GetGlobalDiagnostics (mirroring tsgo's project.CheckerPool).
		_ = p.CompilerQueries().Diagnostics(ctx, lintprogram.DeclarationDiagnostics, file)
		_ = p.CompilerQueries().Diagnostics(ctx, lintprogram.DeclarationDiagnostics, nil)
	})

	t.Run("program-wide diagnostics clear request affinity", func(t *testing.T) {
		var pool *requestIDCheckingPool
		raw := compiler.NewProgram(compiler.ProgramOptions{
			Config: base.CommandLine(),
			Host:   host,
			CreateCheckerPool: func(prog *compiler.Program) compiler.CheckerPool {
				pool = &requestIDCheckingPool{program: prog, sem: make(chan struct{}, 1)}
				return pool
			},
		})
		p := lintprogram.NewFromCompiler(raw)
		ctx, cancel := context.WithTimeout(core.WithRequestID(t.Context(), "lsp-req-1"), 5*time.Second)
		defer cancel()
		_ = p.CompilerQueries().Diagnostics(ctx, lintprogram.SemanticDiagnostics, nil)
		_ = p.CompilerQueries().Diagnostics(ctx, lintprogram.DeclarationDiagnostics, nil)
		pool.mu.Lock()
		got := append([]string(nil), pool.nonEmptyReqIDs...)
		pool.mu.Unlock()
		if len(got) != 0 {
			t.Fatalf("program-wide diagnostics leaked RequestID into parallel pool borrows: %v", got)
		}
	})
}
