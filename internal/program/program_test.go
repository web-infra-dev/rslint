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

type caseSensitiveFS struct {
	vfs.FS
}

func (caseSensitiveFS) UseCaseSensitiveFileNames() bool { return true }

type caseInsensitiveFS struct {
	vfs.FS
}

func (caseInsensitiveFS) UseCaseSensitiveFileNames() bool { return false }

func TestSourceFileLookupFastPathCaseSensitivityAndConcurrency(t *testing.T) {
	t.Parallel()

	const root = "/program-lookup-fast-path"
	upperPath := tspath.ResolvePath(root, "SubDir/Target.ts")
	lowerPath := tspath.ResolvePath(root, "subdir/target.ts")
	importerPath := tspath.ResolvePath(root, "SubDir/Importer.ts")
	missingPath := tspath.ResolvePath(root, "SubDir/Missing.ts")

	for _, caseSensitive := range []bool{true, false} {
		t.Run(fmt.Sprintf("caseSensitive=%t", caseSensitive), func(t *testing.T) {
			t.Parallel()
			files := map[string]string{
				importerPath: `import { value } from "./Target"; export const out = value;`,
				upperPath:    `export const value = 1;`,
			}
			baseFS := bundled.WrapFS(osvfs.FS())
			if caseSensitive {
				baseFS = caseSensitiveFS{FS: baseFS}
			} else {
				baseFS = caseInsensitiveFS{FS: baseFS}
			}
			fs := utils.NewOverlayVFS(baseFS, files)
			host := utils.CreateCompilerHost(root, fs)

			raw, err := utils.CreateProgramFromOptions(true, &core.CompilerOptions{
				Module: core.ModuleKindESNext,
				NoLib:  core.TSTrue,
			}, []string{importerPath, upperPath}, host)
			if err != nil {
				t.Fatalf("CreateProgramFromOptions: %v", err)
			}
			rootProg, err := lintprogram.NewFromRoots(lintprogram.RootOptions{
				RootFileNames:   []string{importerPath, upperPath},
				Host:            host,
				CompilerOptions: &core.CompilerOptions{Module: core.ModuleKindESNext, NoLib: core.TSTrue},
				SingleThreaded:  true,
			})
			if err != nil {
				t.Fatalf("NewFromRoots: %v", err)
			}

			for name, prog := range map[string]*lintprogram.Program{
				"compiler": lintprogram.NewFromCompiler(raw),
				"roots":    rootProg,
			} {
				t.Run(name, func(t *testing.T) {
					wantTarget := prog.GetSourceFile(upperPath)
					if wantTarget == nil || wantTarget.FileName() != upperPath {
						t.Fatalf("exact lookup(%q) = %v, want Target.ts", upperPath, wantTarget)
					}

					var wg sync.WaitGroup
					for range 16 {
						wg.Go(func() {
							for range 4 {
								if got := prog.GetSourceFile(upperPath); got != wantTarget {
									t.Errorf("GetSourceFile(exact) = %v, want %v", got, wantTarget)
								}
								if got := prog.GetSourceFileForResolvedModule(upperPath); got != wantTarget {
									t.Errorf("GetSourceFileForResolvedModule(exact) = %v, want %v", got, wantTarget)
								}
								gotLower := prog.GetSourceFile(lowerPath)
								gotResolvedLower := prog.GetSourceFileForResolvedModule(lowerPath)
								if caseSensitive {
									if gotLower != nil || gotResolvedLower != nil {
										t.Errorf("case-sensitive lookup(%q) = (%v, %v), want (nil, nil)", lowerPath, gotLower, gotResolvedLower)
									}
								} else {
									if gotLower != wantTarget || gotResolvedLower != wantTarget {
										t.Errorf("case-insensitive lookup(%q) = (%v, %v), want %v", lowerPath, gotLower, gotResolvedLower, wantTarget)
									}
								}
								// Unnormalized relative path falls back cleanly.
								if got := prog.GetSourceFile("SubDir/Target.ts"); got != wantTarget {
									t.Errorf("GetSourceFile(relative) = %v, want %v", got, wantTarget)
								}
								// Missing path returns nil consistently across cached misses.
								if got := prog.GetSourceFileForResolvedModule(missingPath); got != nil {
									t.Errorf("GetSourceFileForResolvedModule(missing) = %v, want nil", got)
								}
								importer := prog.GetSourceFile(importerPath)
								refs := prog.ModuleGraph().References(importer, lintprogram.ESModuleReferences)
								if len(refs) != 1 || refs[0].Target != wantTarget {
									t.Errorf("ModuleGraph.References = %+v, want Target=%v", refs, wantTarget)
								}
							}
						})
					}
					wg.Wait()
				})
			}
		})
	}
}

func TestResolvedModuleFileCacheIsolatesProgramGenerations(t *testing.T) {
	files := specifierCacheFiles(false)
	fs := utils.NewOverlayVFS(caseInsensitiveFS{FS: bundled.WrapFS(osvfs.FS())}, files)
	host := utils.CreateCompilerHost("/", fs)
	raw, err := utils.CreateProgramFromOptions(
		true,
		&core.CompilerOptions{AllowJs: core.TSTrue},
		[]string{specifierCacheImporter},
		host,
	)
	if err != nil {
		t.Fatalf("CreateProgramFromOptions: %v", err)
	}
	raw.BindSourceFiles()

	caseFoldedTarget := strings.ToUpper(specifierCacheTarget)
	beforeProg := lintprogram.NewFromCompiler(raw)
	beforeTarget := beforeProg.GetSourceFileForResolvedModule(specifierCacheTarget)
	if beforeTarget == nil {
		t.Fatal("initial program missing target")
	}
	// Exercise resolvedModuleFiles via a case-folded spelling that misses both
	// sourceFileByPath and sourceFilesByFileName.
	if got := beforeProg.GetSourceFileForResolvedModule(caseFoldedTarget); got != beforeTarget {
		t.Fatalf("case-folded lookup before update = %v, want %v", got, beforeTarget)
	}
	if got := beforeProg.GetSourceFileForResolvedModule(caseFoldedTarget); got != beforeTarget {
		t.Fatalf("cached case-folded lookup before update = %v, want %v", got, beforeTarget)
	}
	// Warm the negative cache on a non-existent path as well.
	const missingPath = "/specifier-cache-fixture/nonexistent.ts"
	if beforeProg.GetSourceFileForResolvedModule(missingPath) != nil {
		t.Fatal("missing path resolved unexpectedly")
	}

	files[specifierCacheTarget] = "export const value = 2;\n"
	changed := tspath.ToPath(specifierCacheTarget, raw.GetCurrentDirectory(), fs.UseCaseSensitiveFileNames())
	updated, _, reused := raw.UpdateProgram(changed, raw.Host(), nil)
	if updated == nil || !reused {
		t.Fatalf("UpdateProgram reused=%v, want true", reused)
	}
	updated.BindSourceFiles()

	afterProg := lintprogram.NewFromCompiler(updated)
	afterTarget := afterProg.GetSourceFileForResolvedModule(specifierCacheTarget)
	if afterTarget == nil || afterTarget == beforeTarget {
		t.Fatal("updated program reused stale target from previous generation")
	}
	if got := afterProg.GetSourceFileForResolvedModule(caseFoldedTarget); got != afterTarget {
		t.Fatalf("case-folded lookup after update = %v, want %v (reused stale resolvedModuleFiles entry?)", got, afterTarget)
	}
	if got := beforeProg.GetSourceFileForResolvedModule(caseFoldedTarget); got != beforeTarget {
		t.Fatalf("previous generation case-folded lookup = %v, want %v", got, beforeTarget)
	}
	if beforeProg.GetSourceFileForResolvedModule(specifierCacheTarget) != beforeTarget {
		t.Fatal("previous generation lost its original target")
	}
}

type countingCaseFS struct {
	vfs.FS
	caseSensitive bool
	calls         atomic.Int64
}

func (f *countingCaseFS) UseCaseSensitiveFileNames() bool {
	f.calls.Add(1)
	return f.caseSensitive
}

func TestCaseInsensitiveResolvedModuleLookupOrder(t *testing.T) {
	t.Parallel()

	const root = "/program-case-insensitive-lookup-order"
	mixedPath := tspath.ResolvePath(root, "SubDir/Target.ts")
	lowerPath := tspath.ResolvePath(root, "subdir/target.ts")
	upperPath := strings.ToUpper(mixedPath)
	missingPath := tspath.ResolvePath(root, "SubDir/Missing.ts")

	t.Run("case-insensitive", func(t *testing.T) {
		t.Parallel()
		counter := &countingCaseFS{FS: bundled.WrapFS(osvfs.FS()), caseSensitive: false}
		fs := utils.NewOverlayVFS(counter, map[string]string{
			mixedPath: `export const value = 1;`,
		})
		host := utils.CreateCompilerHost(root, fs)
		rootProg, err := lintprogram.NewFromRoots(lintprogram.RootOptions{
			RootFileNames:   []string{mixedPath},
			Host:            host,
			CompilerOptions: &core.CompilerOptions{Module: core.ModuleKindESNext, NoLib: core.TSTrue},
			SingleThreaded:  true,
		})
		if err != nil {
			t.Fatalf("NewFromRoots: %v", err)
		}
		raw, err := utils.CreateProgramFromOptions(true, &core.CompilerOptions{
			Module: core.ModuleKindESNext,
			NoLib:  core.TSTrue,
		}, []string{mixedPath}, host)
		if err != nil {
			t.Fatalf("CreateProgramFromOptions: %v", err)
		}
		compilerProg := lintprogram.NewFromCompiler(raw)

		wantTarget := rootProg.SourceFiles()[0]
		if wantTarget.FileName() != mixedPath || string(wantTarget.Path()) != lowerPath {
			t.Fatalf("unexpected fixture fileName=%q path=%q", wantTarget.FileName(), wantTarget.Path())
		}

		// Mixed-case exact FileName hits sourceFilesByFileName without fallback.
		counter.calls.Store(0)
		if got := rootProg.GetSourceFileForResolvedModule(mixedPath); got != wantTarget {
			t.Fatalf("GetSourceFileForResolvedModule(%q) = %v, want %v", mixedPath, got, wantTarget)
		}
		if calls := counter.calls.Load(); calls != 1 {
			t.Fatalf("mixed-case lookup made %d UseCaseSensitiveFileNames calls, want 1 (no fallback)", calls)
		}

		// Canonical lowercase path misses sourceFilesByFileName and hits sourceFileByPath without fallback.
		counter.calls.Store(0)
		if got := rootProg.GetSourceFileForResolvedModule(lowerPath); got != wantTarget {
			t.Fatalf("GetSourceFileForResolvedModule(%q) = %v, want %v", lowerPath, got, wantTarget)
		}
		if calls := counter.calls.Load(); calls != 1 {
			t.Fatalf("canonical-lowercase lookup made %d UseCaseSensitiveFileNames calls, want 1 (no fallback)", calls)
		}

		// Alternate casing misses both fast paths and falls back once, then hits resolvedModuleFiles.
		counter.calls.Store(0)
		if got := rootProg.GetSourceFileForResolvedModule(upperPath); got != wantTarget {
			t.Fatalf("GetSourceFileForResolvedModule(%q) = %v, want %v", upperPath, got, wantTarget)
		}
		if calls := counter.calls.Load(); calls != 2 {
			t.Fatalf("first case-folded lookup made %d UseCaseSensitiveFileNames calls, want 2 (fallback)", calls)
		}
		counter.calls.Store(0)
		if got := rootProg.GetSourceFileForResolvedModule(upperPath); got != wantTarget {
			t.Fatalf("cached GetSourceFileForResolvedModule(%q) = %v, want %v", upperPath, got, wantTarget)
		}
		if calls := counter.calls.Load(); calls != 1 {
			t.Fatalf("cached case-folded lookup made %d UseCaseSensitiveFileNames calls, want 1 (cache hit)", calls)
		}

		// Missing path falls back once and caches nil.
		counter.calls.Store(0)
		if got := rootProg.GetSourceFileForResolvedModule(missingPath); got != nil {
			t.Fatalf("GetSourceFileForResolvedModule(%q) = %v, want nil", missingPath, got)
		}
		if calls := counter.calls.Load(); calls != 2 {
			t.Fatalf("first missing lookup made %d UseCaseSensitiveFileNames calls, want 2", calls)
		}
		counter.calls.Store(0)
		if got := rootProg.GetSourceFileForResolvedModule(missingPath); got != nil {
			t.Fatalf("cached GetSourceFileForResolvedModule(%q) = %v, want nil", missingPath, got)
		}
		if calls := counter.calls.Load(); calls != 1 {
			t.Fatalf("cached missing lookup made %d UseCaseSensitiveFileNames calls, want 1", calls)
		}

		// GetSourceFile skips sourceFileByPath on case-insensitive FS and delegates to sourceFile.
		counter.calls.Store(0)
		if got := rootProg.GetSourceFile(mixedPath); got != wantTarget {
			t.Fatalf("GetSourceFile(%q) = %v, want %v", mixedPath, got, wantTarget)
		}
		if calls := counter.calls.Load(); calls != 2 {
			t.Fatalf("case-insensitive GetSourceFile made %d UseCaseSensitiveFileNames calls, want 2", calls)
		}

		wantCompilerTarget := compilerProg.GetSourceFile(mixedPath)
		if wantCompilerTarget == nil {
			t.Fatal("compiler program missing target")
		}
		for _, probe := range []string{mixedPath, lowerPath, upperPath} {
			if got := compilerProg.GetSourceFileForResolvedModule(probe); got != wantCompilerTarget {
				t.Fatalf("compiler GetSourceFileForResolvedModule(%q) = %v, want %v", probe, got, wantCompilerTarget)
			}
		}
		if got := compilerProg.GetSourceFileForResolvedModule(missingPath); got != nil {
			t.Fatalf("compiler GetSourceFileForResolvedModule(%q) = %v, want nil", missingPath, got)
		}
	})

	t.Run("case-sensitive", func(t *testing.T) {
		t.Parallel()
		counter := &countingCaseFS{FS: bundled.WrapFS(osvfs.FS()), caseSensitive: true}
		fs := utils.NewOverlayVFS(counter, map[string]string{
			mixedPath: `export const value = 1;`,
		})
		host := utils.CreateCompilerHost(root, fs)
		rootProg, err := lintprogram.NewFromRoots(lintprogram.RootOptions{
			RootFileNames:   []string{mixedPath},
			Host:            host,
			CompilerOptions: &core.CompilerOptions{Module: core.ModuleKindESNext, NoLib: core.TSTrue},
			SingleThreaded:  true,
		})
		if err != nil {
			t.Fatalf("NewFromRoots: %v", err)
		}
		wantTarget := rootProg.SourceFiles()[0]

		counter.calls.Store(0)
		if got := rootProg.GetSourceFileForResolvedModule(mixedPath); got != wantTarget {
			t.Fatalf("GetSourceFileForResolvedModule(%q) = %v, want %v", mixedPath, got, wantTarget)
		}
		if calls := counter.calls.Load(); calls != 1 {
			t.Fatalf("case-sensitive resolved lookup made %d calls, want 1", calls)
		}

		counter.calls.Store(0)
		if got := rootProg.GetSourceFile(mixedPath); got != wantTarget {
			t.Fatalf("GetSourceFile(%q) = %v, want %v", mixedPath, got, wantTarget)
		}
		if calls := counter.calls.Load(); calls != 1 {
			t.Fatalf("case-sensitive GetSourceFile made %d calls, want 1", calls)
		}

		counter.calls.Store(0)
		if got := rootProg.GetSourceFileForResolvedModule(lowerPath); got != nil {
			t.Fatalf("case-sensitive GetSourceFileForResolvedModule(%q) = %v, want nil", lowerPath, got)
		}
		if calls := counter.calls.Load(); calls != 2 {
			t.Fatalf("case-sensitive miss made %d calls, want 2", calls)
		}
	})
}

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
