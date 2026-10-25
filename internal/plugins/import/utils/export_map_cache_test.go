package utils_test

import (
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"
	"weak"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	import_utils "github.com/web-infra-dev/rslint/internal/plugins/import/utils"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	rslint_utils "github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/modules"
)

func TestExportQueriesAreConservativeAcrossCycles(t *testing.T) {
	files := make(map[string]string)
	entries := make([]string, 24)
	for i := range entries {
		entries[i] = fmt.Sprintf("query-consumer-%d.ts", i)
		files[entries[i]] = fmt.Sprintf("import * as value from './query-graph-%d';", i)
		files[fmt.Sprintf("query-graph-%d.ts", i)] = fmt.Sprintf(`
export const name%d = 1;
export * from './query-graph-%d';
export { nested as renamed } from './query-graph-%d';
import * as local from './query-graph-%d';
export { local as nested };
export default local;
export { missing as broken } from './query-missing';
`, i%7, (i+1)%24, (i*7+3)%24, (i+5)%24)
	}
	contexts := contextsForFiles(t, files, entries...)
	names := []string{"name0", "name1", "name2", "name3", "name4", "name5", "name6", "nested", "renamed", "default", "broken", "absent"}
	for _, ctx := range contexts {
		request := modules.SourceFromSpecifier(firstImportSpecifier(t, ctx.SourceFile))
		exports, ok := import_utils.GetExportMap(ctx, request)
		if !ok {
			t.Fatalf("missing export map for %s", ctx.SourceFile.FileName())
		}
		for _, name := range names {
			meta := exports.Get(name)
			if meta != nil && !import_utils.MayHaveExport(ctx, request, name) {
				t.Errorf("%s: filtered declared export %q", ctx.SourceFile.FileName(), name)
			}
			if meta != nil && meta.Namespace != nil && !import_utils.MayHaveNamespaceExport(ctx, request, name) {
				t.Errorf("%s: filtered namespace metadata for %q", ctx.SourceFile.FileName(), name)
			}
		}
	}
	unknown := contextsForFiles(t, map[string]string{
		"query-unknown-entry.ts": `import * as value from './query-unknown';`,
		"query-unknown.ts":       `export * from './query-missing';`,
	}, "query-unknown-entry.ts")[0]
	source := modules.SourceFromSpecifier(firstImportSpecifier(t, unknown.SourceFile))
	if !import_utils.MayHaveExport(unknown, source, "anything") || !import_utils.MayHaveNamespaceExport(unknown, source, "anything") {
		t.Fatal("unresolved star must retain the complete export-map query")
	}
}

func TestExportQueriesUseLocalFacts(t *testing.T) {
	ctx := contextsForFiles(t, map[string]string{
		"local-query-entry.ts": `import * as all from './local-query';`,
		"local-query.ts": `
import * as ns from './local-query-leaf';
export const object = {};
export { object as alias, ns as nested, ns as overwritten };
export { object as overwritten };
export default ns;
`,
		"local-query-leaf.ts": `export const member = 1;`,
	}, "local-query-entry.ts")[0]
	source := modules.SourceFromSpecifier(firstImportSpecifier(t, ctx.SourceFile))
	for _, name := range []string{"object", "alias", "nested", "overwritten", "default"} {
		if !import_utils.MayHaveExport(ctx, source, name) {
			t.Fatalf("filtered local export %q", name)
		}
	}
	for _, name := range []string{"nested", "overwritten", "default"} {
		if !import_utils.MayHaveNamespaceExport(ctx, source, name) {
			t.Fatalf("filtered possible namespace %q", name)
		}
	}
	for _, name := range []string{"object", "alias", "absent"} {
		if import_utils.MayHaveNamespaceExport(ctx, source, name) {
			t.Fatalf("local facts did not exclude namespace %q", name)
		}
	}
	if import_utils.MayHaveExport(ctx, source, "absent") {
		t.Fatal("local facts did not exclude an absent export")
	}
}

func TestExportQueriesLeaveBarrelsToExportMap(t *testing.T) {
	files := map[string]string{"wide-query-entry.ts": `import * as all from './wide-query';`}
	for i := range 16 {
		files[fmt.Sprintf("wide-query-%d.ts", i)] = fmt.Sprintf("export const object%d = {};", i)
		files["wide-query.ts"] += fmt.Sprintf("export * from './wide-query-%d';\n", i)
	}
	ctx := contextsForFiles(t, files, "wide-query-entry.ts")[0]
	source := modules.SourceFromSpecifier(firstImportSpecifier(t, ctx.SourceFile))
	// A per-name recursive negative probe would scan every leaf for every
	// ordinary object. The local filter must hand these queries to ExportMap.
	for i := range 16 {
		name := fmt.Sprintf("object%d", i)
		if !import_utils.MayHaveNamespaceExport(ctx, source, name) {
			t.Fatalf("barrel query followed dependencies to exclude %q", name)
		}
	}
	if import_utils.MayHaveExport(ctx, source, "default") {
		t.Fatal("star re-exports must not introduce a default export")
	}
	exports, ok := import_utils.GetExportMap(ctx, source)
	if !ok || exports.Size() != 16 {
		t.Fatal("barrel export map lost declared names")
	}
	for i := range 16 {
		name := fmt.Sprintf("object%d", i)
		if meta := exports.Get(name); meta == nil || meta.Namespace != nil {
			t.Fatalf("ordinary object %q has unexpected namespace metadata", name)
		}
	}
}

func TestExportSyntaxRebindsSharedSourceFiles(t *testing.T) {
	root := fixtures.GetRootDir()
	entry := tspath.ResolvePath(root.Dir, "syntax-entry.ts")
	barrel := tspath.ResolvePath(root.Dir, "syntax-barrel.ts")
	cache := rslint_utils.NewParseCache()
	build := func(name, barrelText string) (*compiler.Program, rule.RuleContext) {
		fs := rslint_utils.NewOverlayVFS(root.FS, map[string]string{
			entry:  `import * as all from "./syntax-barrel";`,
			barrel: barrelText,
			tspath.ResolvePath(root.Dir, "syntax-target.ts"): "export const " + name + " = 1;",
		})
		host := rslint_utils.WithParseCache(rslint_utils.CreateCompilerHost(root.Dir, fs), cache)
		raw, err := rslint_utils.CreateProgramFromOptions(true, &core.CompilerOptions{NoLib: core.TSTrue}, []string{entry}, host)
		if err != nil {
			t.Fatal(err)
		}
		return raw, (rule.RuleContext{SourceFile: raw.GetSourceFile(entry)}).WithProgram(lintprogram.NewFromCompiler(raw))
	}
	first, firstCtx := build("first", `export * from "./syntax-target";`)
	second, secondCtx := build("second", `export * from "./syntax-target";`)
	if first.GetSourceFile(barrel) != second.GetSourceFile(barrel) {
		t.Fatal("fixture did not reuse the exact barrel SourceFile")
	}
	var group sync.WaitGroup
	for range 4 {
		for _, test := range []struct {
			ctx             rule.RuleContext
			present, absent string
		}{{firstCtx, "first", "second"}, {secondCtx, "second", "first"}} {
			group.Go(func() {
				source := modules.SourceFromSpecifier(firstImportSpecifier(t, test.ctx.SourceFile))
				exports, ok := import_utils.GetExportMap(test.ctx, source)
				if !ok || !exports.Has(test.present) || exports.Has(test.absent) {
					t.Errorf("shared syntax leaked Program resolution for %s", test.present)
				}
			})
		}
	}
	group.Wait()
	replaced, replacedCtx := build("second", `export { second as renamed } from "./syntax-target";`)
	if replaced.GetSourceFile(barrel) == first.GetSourceFile(barrel) {
		t.Fatal("changed syntax retained the previous SourceFile")
	}
	exports, ok := import_utils.GetExportMap(replacedCtx, modules.SourceFromSpecifier(firstImportSpecifier(t, replacedCtx.SourceFile)))
	if !ok || !exports.Has("renamed") || exports.Has("first") || exports.Has("second") {
		t.Fatal("replaced SourceFile reused stale export syntax")
	}
}

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
		import_utils.MayHaveExport(contexts[0], modules.SourceFromSpecifier(firstImportSpecifier(t, contexts[0].SourceFile)), "absent")
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
