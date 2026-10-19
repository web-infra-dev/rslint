package utils_test

import (
	"runtime"
	"sync"
	"testing"
	"time"
	"weak"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	import_utils "github.com/web-infra-dev/rslint/internal/plugins/import/utils"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/testutil/txtarfs"
	rslint_utils "github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/modules"
)

// exportMapNames is the set of names a map exposes, in a form tests can
// compare. Only names matter here; the metadata is covered elsewhere.
func exportMapNames(exportMap *import_utils.ExportMap, candidates []string) map[string]bool {
	present := make(map[string]bool, len(candidates))
	for _, name := range candidates {
		present[name] = exportMap.Has(name)
	}
	return present
}

// contextsForFiles builds one Program over the given virtual files and
// returns a context per entry file. Everything derived from a Program is
// shared by every context over it, so these see each other's work the way two
// files of one lint run do.
func contextsForFiles(t *testing.T, files map[string]string, entries ...string) []rule.RuleContext {
	contexts, _ := contextsForFilesWithCompiler(t, files, entries...)
	return contexts
}

func contextsForFilesWithCompiler(t *testing.T, files map[string]string, entries ...string) ([]rule.RuleContext, *compiler.Program) {
	t.Helper()

	rootDir := fixtures.GetRootDir()
	resolved := make(map[string]string, len(files))
	for name, code := range files {
		resolved[tspath.ResolvePath(rootDir.Dir, name)] = code
	}
	fs := rslint_utils.NewOverlayVFS(rootDir.FS, resolved)
	host := rslint_utils.CreateCompilerHost(rootDir.Dir, fs)
	program, err := rslint_utils.CreateProgram(true, fs, rootDir.Dir, "tsconfig.json", host)
	if err != nil {
		t.Fatalf("CreateProgram: %v", err)
	}

	sourceProgram := lintprogram.NewFromCompiler(program)
	contexts := make([]rule.RuleContext, 0, len(entries))
	for _, entry := range entries {
		sourceFile := program.GetSourceFile(entry)
		if sourceFile == nil {
			t.Fatalf("entry %q was not parsed", entry)
		}
		contexts = append(contexts, (rule.RuleContext{SourceFile: sourceFile}).WithProgram(sourceProgram))
	}
	return contexts, program
}

func firstImportSpecifier(t *testing.T, sourceFile *ast.SourceFile) *ast.Node {
	t.Helper()
	for _, stmt := range sourceFile.Statements.Nodes {
		if stmt == nil || stmt.Kind != ast.KindImportDeclaration {
			continue
		}
		if specifier := stmt.AsImportDeclaration().ModuleSpecifier; specifier != nil {
			return specifier
		}
	}
	t.Fatalf("no import declaration in %s", sourceFile.FileName())
	return nil
}

// TestExportMapReuseAcrossFiles locks in that sharing one map between files
// gives each of them what building it from scratch would have.
func TestExportMapReuseAcrossFiles(t *testing.T) {
	t.Parallel()

	names := []string{"leaf", "barrel", "default"}
	files := map[string]string{
		"cache-leaf.ts":     `export const leaf = 1;`,
		"cache-barrel.ts":   `export * from "./cache-leaf"; export const barrel = 2;`,
		"cache-first.ts":    `import * as all from "./cache-barrel"; export const first = all;`,
		"cache-second.ts":   `import * as all from "./cache-barrel"; export const second = all;`,
		"cache-isolated.ts": `import * as all from "./cache-barrel"; export const isolated = all;`,
	}

	shared := contextsForFiles(t, files, "cache-first.ts", "cache-second.ts")
	firstMap, ok := import_utils.GetExportMap(shared[0], modules.SourceFromSpecifier(firstImportSpecifier(t, shared[0].SourceFile)))
	if !ok {
		t.Fatal("GetExportMap returned no map for the first file")
	}
	secondMap, ok := import_utils.GetExportMap(shared[1], modules.SourceFromSpecifier(firstImportSpecifier(t, shared[1].SourceFile)))
	if !ok {
		t.Fatal("GetExportMap returned no map for the second file")
	}

	// A separate Program, so nothing is shared with the run above.
	isolated := contextsForFiles(t, files, "cache-isolated.ts")
	isolatedMap, ok := import_utils.GetExportMap(isolated[0], modules.SourceFromSpecifier(firstImportSpecifier(t, isolated[0].SourceFile)))
	if !ok {
		t.Fatal("GetExportMap returned no map for the isolated file")
	}

	want := exportMapNames(isolatedMap, names)
	for label, got := range map[string]map[string]bool{
		"first":  exportMapNames(firstMap, names),
		"second": exportMapNames(secondMap, names),
	} {
		for _, name := range names {
			if got[name] != want[name] {
				t.Fatalf("%s: export %q present=%v, want %v", label, name, got[name], want[name])
			}
		}
	}
}

// TestExportMapCyclicReexportEntryDependent locks in that a module inside a
// re-export cycle still reports what it reports today, which depends on the
// module the query entered from. Sharing such a map between entry points
// would quietly change one of these answers.
func TestExportMapCyclicReexportEntryDependent(t *testing.T) {
	t.Parallel()

	names := []string{"alpha", "beta", "outer"}
	files := map[string]string{
		"cycle-alpha.ts": `export * from "./cycle-beta"; export const alpha = 1;`,
		"cycle-beta.ts":  `export * from "./cycle-alpha"; export const beta = 2;`,
		"cycle-outer.ts": `export * from "./cycle-alpha"; export const outer = 3;`,
		"enter-alpha.ts": `import * as all from "./cycle-alpha"; export const v = all;`,
		"enter-beta.ts":  `import * as all from "./cycle-beta"; export const v = all;`,
		"enter-outer.ts": `import * as all from "./cycle-outer"; export const v = all;`,
	}

	// Each entry gets its own Program, so each answer is the one today's
	// code produces with nothing carried over.
	baseline := make(map[string]map[string]bool)
	for _, entry := range []string{"enter-alpha.ts", "enter-beta.ts", "enter-outer.ts"} {
		contexts := contextsForFiles(t, files, entry)
		exportMap, ok := import_utils.GetExportMap(contexts[0], modules.SourceFromSpecifier(firstImportSpecifier(t, contexts[0].SourceFile)))
		if !ok {
			t.Fatalf("GetExportMap returned no map for %s", entry)
		}
		baseline[entry] = exportMapNames(exportMap, names)
	}

	// Now run all three against one shared Program, in an order that would
	// let a shared map from the first entry answer for the others.
	entries := []string{"enter-alpha.ts", "enter-beta.ts", "enter-outer.ts"}
	shared := contextsForFiles(t, files, entries...)
	for i, ctx := range shared {
		exportMap, ok := import_utils.GetExportMap(ctx, modules.SourceFromSpecifier(firstImportSpecifier(t, ctx.SourceFile)))
		if !ok {
			t.Fatalf("GetExportMap returned no map for %s", entries[i])
		}
		got := exportMapNames(exportMap, names)
		for _, name := range names {
			if got[name] != baseline[entries[i]][name] {
				t.Fatalf("%s: export %q present=%v, want %v", entries[i], name, got[name], baseline[entries[i]][name])
			}
		}
	}
}

// TestExportMapCyclicReexportRepeatedQuery locks in that asking twice from
// the same file gives the same answer, which a half-built map escaping into
// the index would break.
func TestExportMapCyclicReexportRepeatedQuery(t *testing.T) {
	t.Parallel()

	names := []string{"alpha", "beta"}
	files := map[string]string{
		"repeat-alpha.ts": `export * from "./repeat-beta"; export const alpha = 1;`,
		"repeat-beta.ts":  `export * from "./repeat-alpha"; export const beta = 2;`,
		"repeat-entry.ts": `import * as all from "./repeat-alpha"; export const v = all;`,
	}

	contexts := contextsForFiles(t, files, "repeat-entry.ts")
	specifier := firstImportSpecifier(t, contexts[0].SourceFile)

	first, ok := import_utils.GetExportMap(contexts[0], modules.SourceFromSpecifier(specifier))
	if !ok {
		t.Fatal("GetExportMap returned no map on the first query")
	}
	second, ok := import_utils.GetExportMap(contexts[0], modules.SourceFromSpecifier(specifier))
	if !ok {
		t.Fatal("GetExportMap returned no map on the second query")
	}

	want := exportMapNames(first, names)
	got := exportMapNames(second, names)
	for _, name := range names {
		if got[name] != want[name] {
			t.Fatalf("repeated query: export %q present=%v, want %v", name, got[name], want[name])
		}
	}
}

// TestIndexReleasesItsProgram is the failure the Program cache is actually
// exposed to. The cache keys entries by a weak pointer and drops them from a
// cleanup on the Program, so an entry that reaches its own Program keeps it
// reachable, the cleanup never runs, and the entry never goes.
//
// The editor is where that bites: a buffer edit produces a new Program, so a
// leak of one Program per lint is a leak per keystroke, over a Program that
// holds every file of the project. A test that stores an int into the cache
// passes whatever the real entries hold; this one stores what a rule asking
// for an export map really stores.
func TestIndexReleasesItsProgram(t *testing.T) {
	var key weak.Pointer[compiler.Program]

	// The Program, its contexts, and the map built from it are confined to
	// this call, so afterwards only the cache could still be holding it.
	func() {
		contexts, raw := contextsForFilesWithCompiler(t, map[string]string{
			"release-entry.ts":  `import { value } from "./release-target"; export const used = value;`,
			"release-target.ts": `export const value = 1;`,
		}, "release-entry.ts")

		key = weak.Make(raw)

		// Reach the index the way import/namespace and import/default do, so
		// the entry under test is a populated ModuleIndex rather than an
		// empty one.
		if _, ok := import_utils.GetExportMap(contexts[0], modules.SourceFromSpecifier(firstImportSpecifier(t, contexts[0].SourceFile))); !ok {
			t.Fatal("GetExportMap returned no map for the fixture")
		}
	}()

	// Cleanups run on a collection and then on their own goroutine, so the
	// Program goes shortly after it becomes unreachable rather than at a point
	// this test can name.
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		if key.Value() == nil {
			return
		}
		runtime.GC()
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("the import index outlived its program")
}

func commonJSCacheCompiler(t *testing.T, root, config string, fs vfs.FS) *compiler.Program {
	t.Helper()
	raw, err := rslint_utils.CreateProgram(true, fs, root, config, rslint_utils.CreateCompilerHost(root, fs))
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func commonJSCacheRoot(t *testing.T) string {
	t.Helper()
	root := txtarfs.MustParseFile(t, "testdata/commonjs-cache.txtar").Materialize(t, "")
	// Resolve the temporary directory before installing overlays so module
	// resolution's realpath normalization cannot bypass an overlaid file.
	return tspath.NormalizePath(osvfs.FS().Realpath(root))
}

func commonJSCacheBackend(t *testing.T, raw *compiler.Program, backend string) *lintprogram.Program {
	t.Helper()
	if backend == "compiler" {
		return lintprogram.NewFromCompiler(raw)
	}
	if backend == "bound sources" {
		p, err := lintprogram.NewFromBoundSources(raw, raw.SourceFiles())
		if err != nil {
			t.Fatal(err)
		}
		return p
	}
	if backend != "parsed roots" {
		t.Fatalf("unknown backend %q", backend)
	}
	roots := make([]string, 0, len(raw.SourceFiles()))
	for _, file := range raw.SourceFiles() {
		roots = append(roots, file.FileName())
	}
	p, err := lintprogram.NewFromRoots(lintprogram.RootOptions{
		RootFileNames:   roots,
		Host:            raw.Host(),
		CompilerOptions: raw.Options(),
	})
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func commonJSCacheContext(t *testing.T, p *lintprogram.Program, fileName string) (rule.RuleContext, modules.Source) {
	t.Helper()
	file := p.GetSourceFile(tspath.ResolvePath(p.CurrentDirectory(), fileName))
	if file == nil {
		t.Fatalf("missing consumer %q", fileName)
	}
	return (rule.RuleContext{SourceFile: file}).WithProgram(p), modules.SourceFromSpecifier(firstImportSpecifier(t, file))
}

// Each file pass is serial; the counter observes runtime cache reuse without
// sharing a FileCache between concurrent workers.
type commonJSReadCountingFS struct {
	vfs.FS
	reads int
}

func (fs *commonJSReadCountingFS) ReadFile(path string) (string, bool) {
	fs.reads++
	return fs.FS.ReadFile(path)
}

func TestCommonJSImplementationCacheAcrossRequestsAndBackends(t *testing.T) {
	root := commonJSCacheRoot(t)
	for _, config := range []string{"tsconfig.json", "tsconfig.forced.json"} {
		for _, backend := range []string{"compiler", "bound sources", "parsed roots"} {
			t.Run(config+"/"+backend, func(t *testing.T) {
				p := commonJSCacheBackend(t, commonJSCacheCompiler(t, root, config, osvfs.FS()), backend)
				type request struct {
					ctx    rule.RuleContext
					source modules.Source
					want   bool
				}
				var requests []request
				var declaration *ast.SourceFile
				for _, tc := range []struct {
					file string
					want bool
				}{
					{"consumer.mts", false},
					{"consumer.cts", true},
					{"second-consumer.mts", true},
				} {
					ctx, source := commonJSCacheContext(t, p, tc.file)
					_, target, ok := p.ResolveModule(ctx.SourceFile, source)
					if !ok || target == nil || !target.IsDeclarationFile {
						t.Fatalf("%s did not resolve to a declaration", tc.file)
					}
					if declaration != nil && target != declaration {
						t.Fatal("the imports must share a declaration to exercise the cache boundary")
					}
					declaration = target
					requests = append(requests, request{ctx: ctx, source: source, want: tc.want})
				}
				// Race the first queries, then repeat in different orders. Both
				// modes and subpath requests share a declaration, but not an implementation.
				start := make(chan struct{})
				var workers sync.WaitGroup
				for worker := range 24 {
					workers.Go(func() {
						<-start
						for iteration := range 12 {
							query := requests[(worker+iteration)%len(requests)]
							if got := import_utils.IsCommonJSModule(query.ctx, query.source); got != query.want {
								t.Errorf("%s: IsCommonJSModule(%q) = %v, want %v", query.ctx.SourceFile.FileName(), query.source.Specifier().Text(), got, query.want)
								return
							}
						}
					})
				}
				close(start)
				workers.Wait()
			})
		}
	}
}

func TestCommonJSImplementationCacheGenerationIsolation(t *testing.T) {
	root := commonJSCacheRoot(t)
	for _, backend := range []string{"compiler", "bound sources", "parsed roots"} {
		t.Run(backend, func(t *testing.T) {
			before := commonJSCacheBackend(t, commonJSCacheCompiler(t, root, "tsconfig.json", osvfs.FS()), backend)
			changed := rslint_utils.NewOverlayVFS(osvfs.FS(), map[string]string{
				tspath.ResolvePath(root, "node_modules/shared/cjs/index.js"): `export default {}; export const foo = 1;`,
			})
			after := commonJSCacheBackend(t, commonJSCacheCompiler(t, root, "tsconfig.json", changed), backend)
			beforeContext, beforeSource := commonJSCacheContext(t, before, "second-consumer.mts")
			afterContext, afterSource := commonJSCacheContext(t, after, "second-consumer.mts")
			if beforeContext.SourceFile == afterContext.SourceFile {
				t.Fatal("expected distinct AST generations")
			}
			for range 3 {
				if !import_utils.IsCommonJSModule(beforeContext, beforeSource) {
					t.Error("the original generation lost its CommonJS implementation")
				}
				if import_utils.IsCommonJSModule(afterContext, afterSource) {
					t.Error("the new generation reused the old implementation format")
				}
				foreign := afterContext
				foreign.SourceFile = beforeContext.SourceFile
				if import_utils.IsCommonJSModule(foreign, beforeSource) {
					t.Error("a foreign AST borrowed the new generation's cache")
				}
			}
		})
	}
	t.Run("reused compiler with separate file passes", func(t *testing.T) {
		raw := commonJSCacheCompiler(t, root, "tsconfig.json", osvfs.FS())
		for _, change := range []struct {
			name string
			path string
			text string
		}{
			{"implementation syntax", "node_modules/shared/cjs/index.js", `export default {}; export const foo = 1;`},
			{"new package boundary", "node_modules/shared/cjs/package.json", `{"type":"module"}`},
			{"package entry", "node_modules/shared/package.json", `{"exports":{"./commonjs":{"types":"./index.d.ts","default":"./esm/index.js"}}}`},
		} {
			t.Run(change.name, func(t *testing.T) {
				beforeFS := &commonJSReadCountingFS{FS: osvfs.FS()}
				afterFS := &commonJSReadCountingFS{FS: rslint_utils.NewOverlayVFS(osvfs.FS(), map[string]string{
					tspath.ResolvePath(root, change.path): change.text,
				})}
				before := lintprogram.NewFromCompilerWithExternalFileSystem(raw, beforeFS)
				after := lintprogram.NewFromCompilerWithExternalFileSystem(raw, afterFS)
				beforeContext, beforeSource := commonJSCacheContext(t, before, "second-consumer.mts")
				afterContext, afterSource := commonJSCacheContext(t, after, "second-consumer.mts")
				beforeContext = beforeContext.WithFileCache(rule.NewFileCache())
				afterContext = afterContext.WithFileCache(rule.NewFileCache())
				if beforeContext.SourceFile != afterContext.SourceFile || before.FS() != after.FS() {
					t.Fatal("file passes must reuse the same compiler source generation")
				}
				if !import_utils.IsCommonJSModule(beforeContext, beforeSource) {
					t.Fatal("initial file pass lost its CommonJS implementation")
				}
				if import_utils.IsCommonJSModule(afterContext, afterSource) {
					t.Fatal("new file pass reused the previous runtime evidence")
				}
				beforeReads, afterReads := beforeFS.reads, afterFS.reads
				if beforeReads == 0 || afterReads == 0 {
					t.Fatal("file passes did not inspect their external filesystems")
				}
				for range 3 {
					if !import_utils.IsCommonJSModule(beforeContext, beforeSource) || import_utils.IsCommonJSModule(afterContext, afterSource) {
						t.Fatal("independent file-pass caches changed each other's result")
					}
				}
				if beforeFS.reads != beforeReads || afterFS.reads != afterReads {
					t.Fatalf("warm file passes repeated external reads: before %d -> %d; after %d -> %d", beforeReads, beforeFS.reads, afterReads, afterFS.reads)
				}
			})
		}
	})
}

func TestCommonJSImplementationCacheUsesReferencedOptions(t *testing.T) {
	archive := txtarfs.MustParseFile(t, "testdata/commonjs-cache.txtar")
	root := tspath.NormalizePath(osvfs.FS().Realpath(archive.Materialize(t, "reference-options")))
	for _, options := range []string{"conditions", "paths", "suffixes"} {
		for _, backend := range []string{"compiler", "filesystem"} {
			for _, referenceFirst := range []bool{false, true} {
				order := "root first"
				if referenceFirst {
					order = "reference first"
				}
				t.Run(options+"/"+backend+"/"+order, func(t *testing.T) {
					fs := osvfs.FS()
					raw, err := rslint_utils.CreateProgramUsingProjectReferenceSources(
						true, fs, root, "tsconfig."+options+".json", rslint_utils.CreateCompilerHost(root, fs))
					if err != nil {
						t.Fatal(err)
					}
					p := lintprogram.NewFromCompiler(raw)
					if backend == "filesystem" {
						p = lintprogram.NewFromCompilerWithExternalFileSystem(raw, fs)
					}
					entries := []string{"root.ts", "reference.ts"}
					if referenceFirst {
						entries[0], entries[1] = entries[1], entries[0]
					}
					// Both sources share one compiler Program. Runtime reads must
					// still use the importing source's referenced project options.
					for range 3 {
						for _, entry := range entries {
							ctx, source := commonJSCacheContext(t, p, entry)
							redirect := raw.GetRedirectForResolution(ctx.SourceFile)
							if entry == "reference.ts" {
								if redirect == nil || redirect.ConfigName() != tspath.ResolvePath(root, "leaf."+options+".json") {
									t.Fatal("fixture did not retain referenced-project options")
								}
							} else if redirect != nil {
								t.Fatal("root source unexpectedly used referenced-project options")
							}
							declarationName := "node_modules/runtime/types.d.ts"
							if options == "paths" {
								declarationName = "mapped/esm.d.ts"
								if entry == "reference.ts" {
									declarationName = "mapped/cjs.d.ts"
								}
							} else if options == "suffixes" && entry == "reference.ts" {
								declarationName = "node_modules/runtime/types.native.d.ts"
							}
							if path, declaration, resolved := p.ResolveModule(ctx.SourceFile, source); !resolved || declaration == nil ||
								!declaration.IsDeclarationFile || path != tspath.ResolvePath(root, declarationName) {
								t.Fatalf("%s declaration = %q, want %q", entry, path, declarationName)
							}
							if got, want := import_utils.IsCommonJSModule(ctx, source), entry == "reference.ts" && options != "suffixes"; got != want {
								t.Errorf("%s: IsCommonJSModule = %v, want %v", entry, got, want)
							}
						}
					}
				})
			}
		}
	}
}

func TestCommonJSImplementationCachePackageJSONBoundaries(t *testing.T) {
	root := commonJSCacheRoot(t)
	packagePath := tspath.ResolvePath(root, "node_modules/shared/package.json")
	for _, tc := range []struct {
		name        string
		packageJSON string
		declaration string
		want        bool
	}{
		{
			name:        "leading typesVersions",
			packageJSON: `{"typesVersions":{"*":{"*":["types/current"]}},"name":"shared","main":"./esm/index.js","types":"./index.d.ts"}`,
			declaration: "types/current.d.ts",
		},
		{
			name:        "trailing typesVersions",
			packageJSON: `{"name":"shared","main":"./esm/index.js","types":"./index.d.ts","typesVersions":{"*":{"*":["types/current"]}}}`,
			declaration: "types/current.d.ts",
		},
		{
			name:        "escaped typesVersions name",
			packageJSON: `{"name":"shared","main":"./esm/index.js","types":"./index.d.ts","typesVersion\u0073":{"*":{"*":["types/current"]}}}`,
			declaration: "types/current.d.ts",
		},
		{
			name:        "duplicate typesVersions names",
			packageJSON: `{"typesVersions":{"*":{"*":["types/current"]}},"main":"./esm/index.js","types":"./index.d.ts","typesVersions":{"*":{"*":["types/current"]}}}`,
			declaration: "types/current.d.ts",
		},
		{
			name:        "import precedes default",
			packageJSON: `{"exports":{"types":"./index.d.ts","import":"./esm/index.js","default":"./cjs/index.js"},"typesVersions":{"*":{"*":["types/current"]}}}`,
			declaration: "index.d.ts",
		},
		{
			name:        "default precedes import",
			packageJSON: `{"exports":{"types":"./index.d.ts","default":"./cjs/index.js","import":"./esm/index.js"},"typesVersions":{"*":{"*":["types/current"]}}}`,
			declaration: "index.d.ts",
			want:        true,
		},
		{
			name:        "non-object package",
			packageJSON: `null`,
			declaration: "index.d.ts",
		},
		{
			name:        "malformed package",
			packageJSON: `{"typesVersions":`,
			declaration: "index.d.ts",
		},
	} {
		for _, backend := range []string{"compiler", "bound sources", "parsed roots"} {
			t.Run(tc.name+"/"+backend, func(t *testing.T) {
				fs := rslint_utils.NewOverlayVFS(osvfs.FS(), map[string]string{packagePath: tc.packageJSON})
				p := commonJSCacheBackend(t, commonJSCacheCompiler(t, root, "tsconfig.json", fs), backend)
				ctx, source := commonJSCacheContext(t, p, "consumer.mts")
				wantPath := tspath.ResolvePath(root, "node_modules/shared/"+tc.declaration)
				originalNoDtsResolution := p.Options().NoDtsResolution
				for range 2 {
					if path, target, ok := p.ResolveModule(ctx.SourceFile, source); !ok || target == nil || path != wantPath {
						t.Fatalf("type resolution = (%q, %v, %v), want %q", path, target != nil, ok, wantPath)
					}
					if got := import_utils.IsCommonJSModule(ctx, source); got != tc.want {
						t.Errorf("IsCommonJSModule = %v, want %v", got, tc.want)
					}
					if text, ok := p.FS().ReadFile(packagePath); !ok || text != tc.packageJSON {
						t.Error("runtime lookup changed the Program's package metadata")
					}
					if p.Options().NoDtsResolution != originalNoDtsResolution {
						t.Error("runtime lookup changed the Program's declaration resolution")
					}
				}
			})
		}
	}
}

func TestCommonJSImplementationCacheReleasesItsProgram(t *testing.T) {
	root := commonJSCacheRoot(t)
	var key weak.Pointer[compiler.Program]
	func() {
		raw := commonJSCacheCompiler(t, root, "tsconfig.json", osvfs.FS())
		key = weak.Make(raw)
		ctx, source := commonJSCacheContext(t, lintprogram.NewFromCompiler(raw), "second-consumer.mts")
		if !import_utils.IsCommonJSModule(ctx, source) {
			t.Fatal("the fixture must populate the implementation resolver and format caches")
		}
	}()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		if key.Value() == nil {
			return
		}
		runtime.GC()
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("the implementation resolver cache retained its compiler Program")
}
