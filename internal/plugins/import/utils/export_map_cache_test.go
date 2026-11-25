package utils_test

import (
	"fmt"
	"runtime"
	"strings"
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

func requireSourceFile(t *testing.T, raw *compiler.Program, fileName string) *ast.SourceFile {
	t.Helper()
	file := raw.GetSourceFile(fileName)
	if file == nil {
		t.Fatalf("fixture file %q was not loaded into the program", fileName)
	}
	return file
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

// TestExportMapSkipsUnrelatedImportsAndPreservesCaching locks in that
// namespaceImportMeta does not crawl non-namespace or non-matching namespace
// imports (even cyclic ones) and does not poison ExportMap caching.
func TestExportMapSkipsUnrelatedImportsAndPreservesCaching(t *testing.T) {
	t.Parallel()

	t.Run("non-namespace and non-matching namespace imports are not crawled or cycle-poisoning", func(t *testing.T) {
		t.Parallel()

		files := map[string]string{
			"ns-consumer-a.ts": `import * as target from "./ns-target"; export const a = target;`,
			"ns-consumer-b.ts": `import * as target from "./ns-target"; export const b = target;`,
			"ns-target.ts": `
import { fromCycle } from "./ns-cyclic-non-ns";
import unrelatedDefault from "./ns-unrelated";
import "./ns-side-effect";
import * as missingNs from "./ns-does-not-exist";
import * as unusedCycle from "./ns-unused-cycle";
import * as unusedLeaf from "./ns-unused-leaf";
import defFromUsed, * as usedLeaf from "./ns-used-leaf";
export const localValue = fromCycle;
export { localValue as renamedLocal, defFromUsed, missingNs, usedLeaf, usedLeaf as aliasedLeaf };
export default (usedLeaf);
`,
			"ns-cyclic-non-ns.ts": `
import { localValue } from "./ns-target";
export const fromCycle = localValue;
export { fromCycle as again };
export default fromCycle;
`,
			"ns-unrelated.ts":   `export const unrelated = 1; export default unrelated;`,
			"ns-side-effect.ts": `export const sideEffect = 1;`,
			"ns-unused-cycle.ts": `
export * from "./ns-target";
export const fromUnusedCycle = 1;
`,
			"ns-unused-leaf.ts": `export const unusedMember = 1;`,
			"ns-used-leaf.ts":   `export const usedMember = 1; export default usedMember;`,
		}

		contexts, raw := contextsForFilesWithCompiler(t, files, "ns-consumer-a.ts", "ns-consumer-b.ts")
		reqA := modules.SourceFromSpecifier(firstImportSpecifier(t, contexts[0].SourceFile))
		reqB := modules.SourceFromSpecifier(firstImportSpecifier(t, contexts[1].SourceFile))

		firstMap, ok := import_utils.GetExportMap(contexts[0], reqA)
		if !ok {
			t.Fatal("GetExportMap returned no map for first consumer")
		}
		secondMap, ok := import_utils.GetExportMap(contexts[1], reqB)
		if !ok {
			t.Fatal("GetExportMap returned no map for second consumer")
		}
		if firstMap != secondMap {
			t.Fatal("unrelated cyclic imports poisoned ExportMap caching for ns-target.ts")
		}

		index := import_utils.IndexFor(contexts[0])
		targetFile := requireSourceFile(t, raw, "ns-target.ts")
		usedLeafFile := requireSourceFile(t, raw, "ns-used-leaf.ts")
		if !import_utils.HasCachedExportMap(index, targetFile) {
			t.Fatal("expected ns-target.ts ExportMap to be cached")
		}
		if !import_utils.HasCachedExportMap(index, usedLeafFile) {
			t.Fatal("expected matched namespace dependency ns-used-leaf.ts to be cached")
		}
		for _, skipped := range []string{
			"ns-cyclic-non-ns.ts",
			"ns-unrelated.ts",
			"ns-side-effect.ts",
			"ns-unused-cycle.ts",
			"ns-unused-leaf.ts",
		} {
			if import_utils.HasCachedExportMap(index, requireSourceFile(t, raw, skipped)) {
				t.Fatalf("unrelated import %s was eagerly crawled", skipped)
			}
		}

		for _, plain := range []string{"localValue", "renamedLocal", "defFromUsed", "missingNs"} {
			meta := firstMap.Get(plain)
			if meta == nil {
				t.Fatalf("expected export %q to be present", plain)
			}
			if meta.Namespace != nil {
				t.Fatalf("export %q unexpectedly carried namespace metadata", plain)
			}
		}
		for _, reexported := range []string{"usedLeaf", "aliasedLeaf", "default"} {
			meta := firstMap.Get(reexported)
			if meta == nil || meta.Namespace == nil || !meta.Namespace.Has("usedMember") {
				t.Fatalf("expected re-exported namespace %q to expose usedMember", reexported)
			}
		}
	})

	t.Run("cyclic non-namespace imports with no namespace imports stay cached", func(t *testing.T) {
		t.Parallel()

		files := map[string]string{
			"plain-consumer-a.ts": `import target from "./plain-a"; export const v = target;`,
			"plain-consumer-b.ts": `import target from "./plain-a"; export const v = target;`,
			"plain-a.ts":          `import { b } from "./plain-b"; export const a = b; export { a as alias }; export default a;`,
			"plain-b.ts":          `import { a } from "./plain-a"; export const b = a; export default b;`,
		}

		contexts, raw := contextsForFilesWithCompiler(t, files, "plain-consumer-a.ts", "plain-consumer-b.ts")
		firstMap, ok := import_utils.GetExportMap(contexts[0], modules.SourceFromSpecifier(firstImportSpecifier(t, contexts[0].SourceFile)))
		if !ok {
			t.Fatal("GetExportMap returned no map")
		}
		secondMap, ok := import_utils.GetExportMap(contexts[1], modules.SourceFromSpecifier(firstImportSpecifier(t, contexts[1].SourceFile)))
		if !ok {
			t.Fatal("GetExportMap returned no map on second query")
		}
		if firstMap != secondMap {
			t.Fatal("cyclic non-namespace import poisoned ExportMap cache")
		}
		index := import_utils.IndexFor(contexts[0])
		if import_utils.HasCachedExportMap(index, requireSourceFile(t, raw, "plain-b.ts")) {
			t.Fatal("non-namespace cyclic dependency plain-b.ts was eagerly crawled")
		}
	})
}

// TestExportMapNamespaceReexportsAndCycles verifies that cyclic local
// namespace re-exports retain full metadata and entry-dependent cycle semantics.
func TestExportMapNamespaceReexportsAndCycles(t *testing.T) {
	t.Parallel()

	t.Run("mutual and self cyclic namespace re-exports resolve without caching unstable maps", func(t *testing.T) {
		t.Parallel()

		files := map[string]string{
			"enter-a.ts":    `import * as mod from "./ns-cycle-a";`,
			"enter-self.ts": `import * as mod from "./ns-self";`,
			"ns-cycle-a.ts": `
import * as b from "./ns-cycle-b";
export const alpha = 1;
export { b };
export default b;
`,
			"ns-cycle-b.ts": `
import * as a from "./ns-cycle-a";
export const beta = 2;
export { a };
export default a;
`,
			"ns-self.ts": `
import * as self from "./ns-self";
export const value = 1;
export { self };
export default self;
`,
		}

		contexts, raw := contextsForFilesWithCompiler(t, files, "enter-a.ts", "enter-self.ts")
		index := import_utils.IndexFor(contexts[0])

		mapA, ok := import_utils.GetExportMap(contexts[0], modules.SourceFromSpecifier(firstImportSpecifier(t, contexts[0].SourceFile)))
		if !ok {
			t.Fatal("GetExportMap returned no map for ns-cycle-a")
		}
		if !mapA.Has("alpha") {
			t.Fatal("expected ns-cycle-a to export alpha")
		}
		for _, name := range []string{"b", "default"} {
			meta := mapA.Get(name)
			if meta == nil || meta.Namespace == nil || !meta.Namespace.Has("beta") {
				t.Fatalf("expected ns-cycle-a %q namespace to expose beta", name)
			}
			back := meta.Namespace.Get("a")
			if back == nil || back.Namespace == nil || !back.Namespace.Has("alpha") {
				t.Fatalf("expected ns-cycle-a.%s.a namespace to expose alpha", name)
			}
		}
		cycleAFile := requireSourceFile(t, raw, "ns-cycle-a.ts")
		cycleBFile := requireSourceFile(t, raw, "ns-cycle-b.ts")
		if import_utils.HasCachedExportMap(index, cycleAFile) ||
			import_utils.HasCachedExportMap(index, cycleBFile) {
			t.Fatal("cyclic namespace re-export maps must remain uncached")
		}

		mapSelf, ok := import_utils.GetExportMap(contexts[1], modules.SourceFromSpecifier(firstImportSpecifier(t, contexts[1].SourceFile)))
		if !ok {
			t.Fatal("GetExportMap returned no map for ns-self")
		}
		selfMeta := mapSelf.Get("self")
		defaultSelfMeta := mapSelf.Get("default")
		if !mapSelf.Has("value") || selfMeta == nil || selfMeta.Namespace != mapSelf || defaultSelfMeta == nil || defaultSelfMeta.Namespace != mapSelf {
			t.Fatal("expected self-cyclic namespace re-export to point back to itself")
		}
		if import_utils.HasCachedExportMap(index, requireSourceFile(t, raw, "ns-self.ts")) {
			t.Fatal("self-cyclic namespace re-export map must remain uncached")
		}
	})

	t.Run("entry-dependent namespace and star cycle matches isolated queries", func(t *testing.T) {
		t.Parallel()

		files := map[string]string{
			"enter-ns-star-a.ts": `import * as mod from "./ns-star-a";`,
			"enter-ns-star-b.ts": `import * as mod from "./ns-star-b";`,
			"ns-star-a.ts":       `import * as b from "./ns-star-b"; export { b }; export const alpha = 1;`,
			"ns-star-b.ts":       `export * from "./ns-star-a"; export const beta = 2;`,
		}

		entries := []string{"enter-ns-star-a.ts", "enter-ns-star-b.ts"}
		names := []string{"alpha", "beta", "b"}
		baseline := make(map[string]map[string]bool)
		baselineNested := make(map[string]map[string]bool)
		for _, entry := range entries {
			ctx := contextsForFiles(t, files, entry)[0]
			m, ok := import_utils.GetExportMap(ctx, modules.SourceFromSpecifier(firstImportSpecifier(t, ctx.SourceFile)))
			if !ok {
				t.Fatalf("GetExportMap returned no map for %s", entry)
			}
			meta := m.Get("b")
			if meta == nil || meta.Namespace == nil {
				t.Fatalf("%s: expected b to carry namespace metadata", entry)
			}
			baseline[entry] = exportMapNames(m, names)
			baselineNested[entry] = exportMapNames(meta.Namespace, names)
		}

		shared := contextsForFiles(t, files, entries...)
		for i, ctx := range shared {
			m, ok := import_utils.GetExportMap(ctx, modules.SourceFromSpecifier(firstImportSpecifier(t, ctx.SourceFile)))
			if !ok {
				t.Fatalf("GetExportMap returned no map for %s", entries[i])
			}
			meta := m.Get("b")
			if meta == nil || meta.Namespace == nil {
				t.Fatalf("%s: expected b to carry namespace metadata", entries[i])
			}
			got := exportMapNames(m, names)
			gotNested := exportMapNames(meta.Namespace, names)
			for _, name := range names {
				if got[name] != baseline[entries[i]][name] {
					t.Fatalf("%s: export %q present=%v, want %v", entries[i], name, got[name], baseline[entries[i]][name])
				}
				if gotNested[name] != baselineNested[entries[i]][name] {
					t.Fatalf("%s: nested b.%q present=%v, want %v", entries[i], name, gotNested[name], baselineNested[entries[i]][name])
				}
			}
		}
	})

	t.Run("2-node cycle re-exported via sibling namespaces converges members in-place and stays uncached", func(t *testing.T) {
		t.Parallel()

		for _, tc := range []struct {
			name        string
			files       map[string]string
			namespaces  []string
			wantExports []string
			uncached    []string
		}{
			{
				name: "b imported before a, value only in a",
				files: map[string]string{
					"consumer.ts": `import { a, b } from "./index";`,
					"index.ts": `
import * as b from "./b";
import * as a from "./a";
export { a, b };
`,
					"a.ts": `
export * from "./b";
export const value = 42;
`,
					"b.ts": `export * from "./a";`,
				},
				namespaces:  []string{"a", "b"},
				wantExports: []string{"value"},
				uncached:    []string{"index.ts", "a.ts", "b.ts"},
			},
			{
				name: "a imported before b, both declare exports",
				files: map[string]string{
					"consumer.ts": `import { a, b } from "./index";`,
					"index.ts": `
import * as a from "./a";
import * as b from "./b";
export { a, b };
`,
					"a.ts": `
export * from "./b";
export const alpha = 1;
`,
					"b.ts": `
export * from "./a";
export const beta = 2;
`,
				},
				namespaces:  []string{"a", "b"},
				wantExports: []string{"alpha", "beta"},
				uncached:    []string{"index.ts", "a.ts", "b.ts"},
			},
		} {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				contexts, raw := contextsForFilesWithCompiler(t, tc.files, "consumer.ts")
				index := import_utils.IndexFor(contexts[0])
				exports, ok := import_utils.GetExportMap(contexts[0], modules.SourceFromSpecifier(firstImportSpecifier(t, contexts[0].SourceFile)))
				if !ok {
					t.Fatal("GetExportMap returned no map for index.ts")
				}
				for _, ns := range tc.namespaces {
					meta := exports.Get(ns)
					if meta == nil || meta.Namespace == nil {
						t.Fatalf("expected %q to carry namespace metadata", ns)
					}
					for _, exp := range tc.wantExports {
						if !meta.Namespace.Has(exp) {
							t.Fatalf("expected %s.%s to be present", ns, exp)
						}
					}
				}
				for _, fileName := range tc.uncached {
					if import_utils.HasCachedExportMap(index, requireSourceFile(t, raw, fileName)) {
						t.Fatalf("cyclic module %s must remain uncached", fileName)
					}
				}
			})
		}
	})

	t.Run("3-node cycle and cross-edge nested component re-exported as sibling namespaces expose all exports", func(t *testing.T) {
		t.Parallel()

		for _, tc := range []struct {
			name        string
			files       map[string]string
			namespaces  []string
			wantExports []string
			uncached    []string
		}{
			{
				name: "3-node cycle a -> b -> c -> a",
				files: map[string]string{
					"consumer.ts": `import { a, b, c } from "./index";`,
					"index.ts": `
import * as c from "./c";
import * as b from "./b";
import * as a from "./a";
export { a, b, c };
`,
					"a.ts": `export * from "./b"; export const alpha = 1;`,
					"b.ts": `export * from "./c"; export const beta = 2;`,
					"c.ts": `export * from "./a"; export const gamma = 3;`,
				},
				namespaces:  []string{"a", "b", "c"},
				wantExports: []string{"alpha", "beta", "gamma"},
				uncached:    []string{"index.ts", "a.ts", "b.ts", "c.ts"},
			},
			{
				name: "3-node cycle with export only in root",
				files: map[string]string{
					"consumer.ts": `import { a, b, c } from "./index";`,
					"index.ts": `
import * as a from "./a";
import * as b from "./b";
import * as c from "./c";
export { a, b, c };
`,
					"a.ts": `export * from "./b"; export const value = 42;`,
					"b.ts": `export * from "./c";`,
					"c.ts": `export * from "./a";`,
				},
				namespaces:  []string{"a", "b", "c"},
				wantExports: []string{"value"},
				uncached:    []string{"index.ts", "a.ts", "b.ts", "c.ts"},
			},
			{
				name: "diamond cycle inside component re-exported via sibling namespaces",
				files: map[string]string{
					"consumer.ts": `import { a, b1, b2, c } from "./index";`,
					"index.ts": `
import * as a from "./a";
import * as b1 from "./b1";
import * as b2 from "./b2";
import * as c from "./c";
export { a, b1, b2, c };
`,
					"a.ts":  `export * from "./b1"; export * from "./b2"; export const value = 42;`,
					"b1.ts": `export * from "./c";`,
					"b2.ts": `export * from "./c";`,
					"c.ts":  `export * from "./a";`,
				},
				namespaces:  []string{"a", "b1", "b2", "c"},
				wantExports: []string{"value"},
				uncached:    []string{"index.ts", "a.ts", "b1.ts", "b2.ts", "c.ts"},
			},
			{
				name: "cross-edge to nested component member converges transitively incomplete sibling",
				files: map[string]string{
					"consumer.ts": `import { p, q2 } from "./index";`,
					"index.ts": `
import * as p from "./p";
import * as q2 from "./q2";
export { p, q2 };
`,
					"p.ts":  `export * from "./q1"; export * from "./q2"; export const rootValue = 1;`,
					"q1.ts": `export * from "./v"; export * from "./w";`,
					"v.ts":  `export * from "./q1";`,
					"w.ts":  `export * from "./p";`,
					"q2.ts": `export * from "./v";`,
				},
				namespaces:  []string{"p", "q2"},
				wantExports: []string{"rootValue"},
				uncached:    []string{"index.ts", "p.ts", "q1.ts", "v.ts", "w.ts", "q2.ts"},
			},
			{
				name: "cyclic namespace re-export preserves exports declared after namespace export",
				files: map[string]string{
					"consumer.ts": `import { c } from "./index";`,
					"index.ts": `
export { value } from "./a";
export { c } from "./b";
`,
					"a.ts": `export { value } from "./c";`,
					"b.ts": `
import * as c from "./c";
export { c };
export const value = 42;
export * from "./a";
`,
					"c.ts": `export * from "./b";`,
				},
				namespaces:  []string{"c"},
				wantExports: []string{"value", "c"},
				uncached:    []string{"index.ts", "a.ts", "b.ts", "c.ts"},
			},
			{
				name: "cyclic namespace re-export starting from namespace module first",
				files: map[string]string{
					"consumer.ts": `import { c } from "./index";`,
					"index.ts": `
export { c } from "./b";
export { value } from "./a";
`,
					"a.ts": `export { value } from "./c";`,
					"b.ts": `
import * as c from "./c";
export { c };
export const value = 42;
export * from "./a";
`,
					"c.ts": `export * from "./b";`,
				},
				namespaces:  []string{"c"},
				wantExports: []string{"value", "c"},
				uncached:    []string{"index.ts", "a.ts", "b.ts", "c.ts"},
			},
		} {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				contexts, raw := contextsForFilesWithCompiler(t, tc.files, "consumer.ts")
				index := import_utils.IndexFor(contexts[0])
				exports, ok := import_utils.GetExportMap(contexts[0], modules.SourceFromSpecifier(firstImportSpecifier(t, contexts[0].SourceFile)))
				if !ok {
					t.Fatal("GetExportMap returned no map for index.ts")
				}
				for _, ns := range tc.namespaces {
					meta := exports.Get(ns)
					if meta == nil || meta.Namespace == nil {
						t.Fatalf("expected %q to carry namespace metadata", ns)
					}
					for _, exp := range tc.wantExports {
						if !meta.Namespace.Has(exp) {
							t.Fatalf("expected %s.%s to be present", ns, exp)
						}
					}
				}
				for _, fileName := range tc.uncached {
					if import_utils.HasCachedExportMap(index, requireSourceFile(t, raw, fileName)) {
						t.Fatalf("cyclic module %s must remain uncached", fileName)
					}
				}
			})
		}

		t.Run("cyclic star re-export preserves resolved local metadata, namespace, deprecation, and HasDefault", func(t *testing.T) {
			t.Parallel()

			for _, tc := range []struct {
				name  string
				index string
			}{
				{
					name: "value re-export before namespace re-export",
					index: `
export { value, default } from "./a";
export { c } from "./b";
`,
				},
				{
					name: "namespace re-export before value re-export",
					index: `
export { c } from "./b";
export { value, default } from "./a";
`,
				},
			} {
				t.Run(tc.name, func(t *testing.T) {
					t.Parallel()

					files := map[string]string{
						"consumer.ts": `import { c, value } from "./index";`,
						"leaf.ts":     `export const leafProp = 1;`,
						"a.ts": `
export { value } from "./c";
export { value as default } from "./c";
`,
						"b.ts": `
import * as c from "./c";
import * as leaf from "./leaf";
export { c };
/** @deprecated use replacement */
export const deprecatedVal = 1;
export { leaf as value };
export * from "./a";
`,
						"c.ts":     `export * from "./b";`,
						"index.ts": tc.index,
					}
					contexts, raw := contextsForFilesWithCompiler(t, files, "consumer.ts")
					index := import_utils.IndexFor(contexts[0])
					exports, ok := import_utils.GetExportMap(contexts[0], modules.SourceFromSpecifier(firstImportSpecifier(t, contexts[0].SourceFile)))
					if !ok {
						t.Fatal("GetExportMap returned no map for index.ts")
					}
					if !exports.HasDefault() {
						t.Fatal("expected index.ts HasDefault() == true through cyclic default re-export")
					}
					valMeta := exports.Get("value")
					if valMeta == nil || valMeta.Namespace == nil || !valMeta.Namespace.Has("leafProp") {
						t.Fatalf("expected index.ts value to retain leaf Namespace metadata, got %#v", valMeta)
					}
					cMeta := exports.Get("c")
					if cMeta == nil || cMeta.Namespace == nil {
						t.Fatalf("expected index.ts c to retain Namespace metadata, got %#v", cMeta)
					}
					cVal := cMeta.Namespace.Get("value")
					if cVal == nil || cVal.Namespace == nil || !cVal.Namespace.Has("leafProp") {
						t.Fatalf("expected c.value to retain leaf Namespace metadata, got %#v", cVal)
					}
					cDep := cMeta.Namespace.Get("deprecatedVal")
					if cDep == nil || cDep.Deprecation(contexts[0]) == nil {
						t.Fatalf("expected c.deprecatedVal to retain @deprecated metadata, got %#v", cDep)
					}
					for _, fileName := range []string{"index.ts", "a.ts", "b.ts", "c.ts"} {
						if import_utils.HasCachedExportMap(index, requireSourceFile(t, raw, fileName)) {
							t.Fatalf("cyclic module %s must remain uncached", fileName)
						}
					}
				})
			}
		})

		t.Run("explicit unresolved re-export is not clobbered by earlier star or duplicate specifier", func(t *testing.T) {
			t.Parallel()

			files := map[string]string{
				"consumer.ts": `import * as b from "./cycB";`,
				"leaf.ts":     `export const x = 1;`,
				"cycA.ts": `
import * as leaf from "./leaf";
export * from "./cycB";
export { leaf as Button, leaf as dup, leaf as sameStep };
export const ok = 1;
`,
				"cycB.ts": `
export * from "./cycA";
export { Button } from "./missing-module";
export { dup } from "./cycA";
export { missing as dup } from "./cycA";
export { sameStep, missing as sameStep } from "./cycA";
export { ok as default } from "./cycA";
export { missing as default } from "./cycA";
`,
			}
			contexts, _ := contextsForFilesWithCompiler(t, files, "consumer.ts")
			exports, ok := import_utils.GetExportMap(contexts[0], modules.SourceFromSpecifier(firstImportSpecifier(t, contexts[0].SourceFile)))
			if !ok {
				t.Fatal("GetExportMap returned no map for cycB.ts")
			}
			// Explicit `export { Button } from "./missing-module"` shadows earlier `export * from "./cycA"`.
			if btn := exports.Get("Button"); btn == nil || btn.Namespace != nil {
				t.Fatalf("expected explicit unresolved Button to have nil Namespace, got %#v", btn)
			}
			// Earlier `export { dup }` must not overwrite later `export { missing as dup }`.
			if dup := exports.Get("dup"); dup == nil || dup.Namespace != nil {
				t.Fatalf("expected later unresolved dup to stay unresolved, got %#v", dup)
			}
			// Same-statement duplicate specifier: earlier `sameStep` must not overwrite later `missing as sameStep`.
			if same := exports.Get("sameStep"); same == nil || same.Namespace != nil {
				t.Fatalf("expected later same-step unresolved sameStep to stay unresolved, got %#v", same)
			}
			// Later `export { missing as default }` shadows earlier `export { ok as default }`.
			if exports.HasDefault() {
				t.Fatal("expected cycB.ts HasDefault() == false when shadowed by later unresolved default re-export")
			}
		})

		t.Run("cyclic default-only import attribute and NodeNext NodeDefault terminate", func(t *testing.T) {
			t.Parallel()

			root := fixtures.GetRootDir()
			baseDir := tspath.ResolvePath(root.Dir, "attr-cycle")
			files := map[string]string{
				"tsconfig.json": `{"compilerOptions":{"module":"nodenext","moduleResolution":"nodenext","allowJs":true,"noLib":true},"files":["consumer.mts","a.mts","b.mts","c.mts","d.cts","data.cjs"]}`,
				"package.json":  `{"type":"module"}`,
				"consumer.mts":  `import a from "./a.mts"; import c from "./c.mts";`,
				"a.mts":         `export { default } from "./b.mts"; export { default as txt } from "./data.cjs" with { type: "text" };`,
				"b.mts":         `export { default } from "./a.mts";`,
				"c.mts":         `export { default } from "./d.cts";`,
				"d.cts":         `export { default } from "./c.mts";`,
				"data.cjs":      `module.exports = { raw: 1 };`,
			}
			overlays := make(map[string]string, len(files))
			for rel, content := range files {
				overlays[tspath.ResolvePath(baseDir, rel)] = content
			}
			fs := rslint_utils.NewOverlayVFS(root.FS, overlays)
			host := rslint_utils.CreateCompilerHost(baseDir, fs)
			raw, err := rslint_utils.CreateProgram(true, fs, baseDir, "tsconfig.json", host)
			if err != nil {
				t.Fatalf("CreateProgram: %v", err)
			}
			prog := lintprogram.NewFromCompiler(raw)
			consumerSF := raw.GetSourceFile(tspath.ResolvePath(baseDir, "consumer.mts"))
			if consumerSF == nil {
				t.Fatal("consumer.mts was not parsed")
			}
			ctx := (rule.RuleContext{SourceFile: consumerSF}).WithProgram(prog)
			var specs []*ast.Node
			for _, stmt := range consumerSF.Statements.Nodes {
				if ast.IsImportDeclaration(stmt) {
					specs = append(specs, stmt.AsImportDeclaration().ModuleSpecifier)
				}
			}

			if len(specs) != 2 {
				t.Fatalf("expected 2 import specifiers, got %d", len(specs))
			}

			aExports, ok := import_utils.GetExportMap(ctx, modules.SourceFromSpecifier(specs[0]))
			if !ok || aExports == nil || !aExports.Has("txt") {
				t.Fatalf("expected a.mts to expose txt default-only re-export, got ok=%v exports=%#v", ok, aExports)
			}
			cExports, ok := import_utils.GetExportMap(ctx, modules.SourceFromSpecifier(specs[1]))
			if !ok || cExports == nil || !cExports.HasDefault() {
				t.Fatalf("expected c.mts <-> d.cts NodeDefault cycle to resolve default, got ok=%v exports=%#v", ok, cExports)
			}
		})

		t.Run("cyclic implicitDefault re-export terminates and upgrades synthetic to explicit", func(t *testing.T) {
			t.Parallel()

			// Repro 1: 4-file cycle where two modules carry implicitDefault (`export =`) and re-export default.
			cycleFiles := map[string]string{
				"consumer.ts": `import a from "./a";`,
				"a.ts":        `export { default } from "./b";`,
				"b.ts":        `export { default } from "./c";`,
				"c.ts":        `declare const cVal: number; export = cVal; export { default } from "./d";`,
				"d.ts":        `declare const dVal: number; export = dVal; export { default } from "./a";`,
			}
			contexts, _ := contextsForFilesWithCompiler(t, cycleFiles, "consumer.ts")
			exports, ok := import_utils.GetExportMap(contexts[0], modules.SourceFromSpecifier(firstImportSpecifier(t, contexts[0].SourceFile)))
			if !ok || exports == nil || !exports.HasDefault() {
				t.Fatalf("expected 4-file implicitDefault cycle to resolve default, got ok=%v exports=%#v", ok, exports)
			}

			// Repro 2: synthetic default placeholder must upgrade to explicit namespace default across both named and star edges.
			upgradeFiles := map[string]string{
				"consumer.ts": `import * as a from "./a"; import * as c from "./c";`,
				"leaf.ts":     `export const leafProp = 1;`,
				"a.ts": `
export { default as aliased } from "./b";
export * from "./c";
`,
				"b.ts": `
declare const cjsVal: number;
export = cjsVal;
export { default } from "./c";
`,
				"c.ts": `
import * as leaf from "./leaf";
export * from "./a";
export { leaf as default };
`,
			}
			upContexts, _ := contextsForFilesWithCompiler(t, upgradeFiles, "consumer.ts")
			var specifiers []*ast.Node
			for _, stmt := range upContexts[0].SourceFile.Statements.Nodes {
				if ast.IsImportDeclaration(stmt) {
					specifiers = append(specifiers, stmt.AsImportDeclaration().ModuleSpecifier)
				}
			}
			aExports, ok := import_utils.GetExportMap(upContexts[0], modules.SourceFromSpecifier(specifiers[0]))
			if !ok {
				t.Fatal("GetExportMap returned no map for a.ts")
			}
			if meta := aExports.Get("aliased"); meta == nil || meta.Namespace == nil || !meta.Namespace.Has("leafProp") {
				t.Fatalf("expected a.aliased to upgrade from synthetic to explicit leaf Namespace, got %#v", meta)
			}
			cExports, ok := import_utils.GetExportMap(upContexts[0], modules.SourceFromSpecifier(specifiers[1]))
			if !ok {
				t.Fatal("GetExportMap returned no map for c.ts")
			}
			if meta := cExports.Get("aliased"); meta == nil || meta.Namespace == nil || !meta.Namespace.Has("leafProp") {
				t.Fatalf("expected c.aliased (via export *) to upgrade to explicit leaf Namespace, got %#v", meta)
			}
		})

		t.Run("self-loop explicit re-export and cyclic hasUnknown converge in-place", func(t *testing.T) {
			t.Parallel()

			files := map[string]string{
				"consumer.ts": `import * as s from "./self"; import * as a from "./a";`,
				"leaf.ts":     `export const leafProp = 1;`,
				"self.ts": `
import * as leaf from "./leaf";
export { value as alias } from "./self";
export { value as default } from "./self";
export { leaf as value };
`,
				// Enter from a.ts so b.ts builds BEFORE a.ts processes `./does-not-exist`:
				// b.hasUnknown starts false and can only become true inside convergeSCC.
				"a.ts": `
import * as b from "./b";
export { b };
export * from "./b";
export * from "./does-not-exist";
`,
				"b.ts": `export * from "./a";`,
			}
			contexts, raw := contextsForFilesWithCompiler(t, files, "consumer.ts")
			index := import_utils.IndexFor(contexts[0])
			var specifiers []*ast.Node
			for _, stmt := range contexts[0].SourceFile.Statements.Nodes {
				if ast.IsImportDeclaration(stmt) {
					specifiers = append(specifiers, stmt.AsImportDeclaration().ModuleSpecifier)
				}
			}
			selfExports, ok := import_utils.GetExportMap(contexts[0], modules.SourceFromSpecifier(specifiers[0]))
			if !ok {
				t.Fatal("GetExportMap returned no map for self.ts")
			}
			aliasMeta := selfExports.Get("alias")
			if aliasMeta == nil || aliasMeta.Namespace == nil || !aliasMeta.Namespace.Has("leafProp") || !selfExports.HasDefault() {
				t.Fatalf("expected self.ts alias and default to resolve leaf Namespace, alias=%#v HasDefault=%v", aliasMeta, selfExports.HasDefault())
			}
			if import_utils.HasCachedExportMap(index, requireSourceFile(t, raw, "self.ts")) {
				t.Fatal("self-cyclic module self.ts must remain uncached")
			}
			aExports, ok := import_utils.GetExportMap(contexts[0], modules.SourceFromSpecifier(specifiers[1]))
			if !ok {
				t.Fatal("GetExportMap returned no map for a.ts")
			}
			bMeta := aExports.Get("b")
			if bMeta == nil || bMeta.Namespace == nil || !bMeta.Namespace.Has("anythingUnknown") {
				t.Fatalf("expected hasUnknown from a.ts to propagate in-place to b.ts namespace, got %#v", bMeta)
			}
		})

		t.Run("mutual export star chain and directed cycle build each module once", func(t *testing.T) {
			t.Parallel()

			const chainLen = 64
			for _, tc := range []struct {
				name     string
				directed bool
			}{
				{name: "bidirectional 64-module chain", directed: false},
				{name: "directed 64-module cycle", directed: true},
			} {
				t.Run(tc.name, func(t *testing.T) {
					t.Parallel()

					files := make(map[string]string, chainLen+3)
					files["consumer.ts"] = `import * as all from "./index";`
					files["leaf.ts"] = `export const leafProp = 42;`
					var indexSrc strings.Builder
					for i := range chainLen {
						var modSrc strings.Builder
						if i == 0 {
							modSrc.WriteString("import * as leaf from './leaf';\n")
						}
						if tc.directed {
							fmt.Fprintf(&modSrc, "export * from './m%d';\n", (i+1)%chainLen)
						} else {
							if i > 0 {
								fmt.Fprintf(&modSrc, "export * from './m%d';\n", i-1)
							}
							if i+1 < chainLen {
								fmt.Fprintf(&modSrc, "export * from './m%d';\n", i+1)
							}
						}
						// Declare `value` AFTER `export *` so DFS back-edges see an empty map
						// and `convergeSCC` must propagate its leaf Namespace across all 64 modules.
						if i == 0 {
							modSrc.WriteString("export { leaf as value };\n")
						}
						files[fmt.Sprintf("m%d.ts", i)] = modSrc.String()
						fmt.Fprintf(&indexSrc, "export { value as v%d } from './m%d';\n", i, i)
					}
					files["index.ts"] = indexSrc.String()

					contexts, raw := contextsForFilesWithCompiler(t, files, "consumer.ts")
					index := import_utils.IndexFor(contexts[0])
					exports, buildCount, ok := import_utils.GetExportMapWithBuildCount(contexts[0], modules.SourceFromSpecifier(firstImportSpecifier(t, contexts[0].SourceFile)))
					if !ok {
						t.Fatal("GetExportMapWithBuildCount returned no map for index.ts")
					}
					// +2 = index.ts + leaf.ts
					if want := chainLen + 2; buildCount != want {
						t.Fatalf("buildCount = %d, want %d (must not rebuild SCC members quadratically)", buildCount, want)
					}
					if import_utils.HasCachedExportMap(index, requireSourceFile(t, raw, "index.ts")) {
						t.Fatal("index.ts depending on cyclic SCC must remain uncached")
					}
					for i := range chainLen {
						alias := fmt.Sprintf("v%d", i)
						meta := exports.Get(alias)
						if meta == nil || meta.Namespace == nil || !meta.Namespace.Has("leafProp") {
							t.Fatalf("expected alias %q to propagate leaf Namespace through 64-module SCC, got %#v", alias, meta)
						}
						if import_utils.HasCachedExportMap(index, requireSourceFile(t, raw, fmt.Sprintf("m%d.ts", i))) {
							t.Fatalf("cyclic module m%d.ts must remain uncached", i)
						}
					}
				})
			}
		})

		t.Run("star re-export does not override explicit binding regardless of statement order", func(t *testing.T) {
			t.Parallel()

			bOrders := []struct {
				name string
				code string
			}{
				{
					name: "explicit before star",
					code: "export { alias as api } from './a.js';\nexport * from './a.js';\n",
				},
				{
					name: "star before explicit",
					code: "export * from './a.js';\nexport { alias as api } from './a.js';\n",
				},
			}

			for _, order := range bOrders {
				t.Run("plain value source/"+order.name, func(t *testing.T) {
					t.Parallel()

					files := map[string]string{
						"consumer-a.ts": `import { b } from './a.js';`,
						"consumer-b.ts": `import * as b from './b.js';`,
						"a.ts": `
import * as b from './b.js';
export { b, b as api };
export { source as alias } from './b.js';
export const source = { good: 42 };
`,
						"b.ts": order.code,
					}

					for _, entry := range []string{"consumer-a.ts", "consumer-b.ts"} {
						contexts := contextsForFiles(t, files, entry)
						exports, ok := import_utils.GetExportMap(contexts[0], modules.SourceFromSpecifier(firstImportSpecifier(t, contexts[0].SourceFile)))
						if !ok {
							t.Fatalf("%s: GetExportMap returned no map", entry)
						}
						bMap := exports
						if entry == "consumer-a.ts" {
							bMeta := exports.Get("b")
							if bMeta == nil || bMeta.Namespace == nil {
								t.Fatalf("%s: expected a.b to carry namespace metadata, got %#v", entry, bMeta)
							}
							bMap = bMeta.Namespace
						}
						apiMeta := bMap.Get("api")
						if apiMeta == nil {
							t.Fatalf("%s: expected b.api to be exported", entry)
						}
						if apiMeta.Namespace != nil {
							t.Fatalf("%s: expected b.api to resolve to non-namespace a.source, got Namespace=%p", entry, apiMeta.Namespace)
						}
						if !bMap.Has("source") || bMap.Has("missing") {
							t.Fatalf("%s: unexpected b exports: Has(source)=%v Has(missing)=%v", entry, bMap.Has("source"), bMap.Has("missing"))
						}
					}
				})

				t.Run("namespace source/"+order.name, func(t *testing.T) {
					t.Parallel()

					files := map[string]string{
						"consumer-a.ts": `import { b } from './a.js';`,
						"consumer-b.ts": `import * as b from './b.js';`,
						"source.ts":     `export const good = 42;`,
						"a.ts": `
import * as b from './b.js';
import * as source from './source.js';
export { b, b as api, source };
export { source as alias } from './b.js';
`,
						"b.ts": order.code,
					}

					for _, entry := range []string{"consumer-a.ts", "consumer-b.ts"} {
						contexts := contextsForFiles(t, files, entry)
						exports, ok := import_utils.GetExportMap(contexts[0], modules.SourceFromSpecifier(firstImportSpecifier(t, contexts[0].SourceFile)))
						if !ok {
							t.Fatalf("%s: GetExportMap returned no map", entry)
						}
						bMap := exports
						if entry == "consumer-a.ts" {
							bMeta := exports.Get("b")
							if bMeta == nil || bMeta.Namespace == nil {
								t.Fatalf("%s: expected a.b to carry namespace metadata, got %#v", entry, bMeta)
							}
							bMap = bMeta.Namespace
						}
						apiMeta := bMap.Get("api")
						if apiMeta == nil || apiMeta.Namespace == nil {
							t.Fatalf("%s: expected b.api to resolve to source namespace, got %#v", entry, apiMeta)
						}
						if apiMeta.Namespace == bMap {
							t.Fatalf("%s: star re-export overrode explicit b.api with a.api (b's own namespace)", entry)
						}
						if !apiMeta.Namespace.Has("good") || apiMeta.Namespace.Has("missing") {
							t.Fatalf("%s: expected b.api namespace to expose good only, got Has(good)=%v Has(missing)=%v",
								entry, apiMeta.Namespace.Has("good"), apiMeta.Namespace.Has("missing"))
						}
					}
				})
			}

			t.Run("unresolved explicit re-export before star stays unresolved in cyclic and acyclic modules", func(t *testing.T) {
				t.Parallel()

				files := map[string]string{
					"consumer-cyc.ts":     `import * as b from "./cycB";`,
					"consumer-acyclic.ts": `import * as b from "./acyclicB";`,
					"leaf.ts":             `export const x = 1;`,
					"cycA.ts": `
import * as leaf from "./leaf";
export * from "./cycB";
export { leaf as Button, leaf as other };
export const ok = 1;
`,
					"cycB.ts": `
export { Button } from "./missing-module";
export { missing as other } from "./cycA";
export { missing as default } from "./cycA";
export * from "./cycA";
`,
					"acyclicA.ts": `
import * as leaf from "./leaf";
export { leaf as Button, leaf as other };
export const ok = 1;
`,
					"acyclicB.ts": `
export { Button } from "./missing-module";
export { missing as other } from "./acyclicA";
export { missing as default } from "./acyclicA";
export * from "./acyclicA";
`,
				}

				for _, entry := range []string{"consumer-cyc.ts", "consumer-acyclic.ts"} {
					contexts := contextsForFiles(t, files, entry)
					exports, ok := import_utils.GetExportMap(contexts[0], modules.SourceFromSpecifier(firstImportSpecifier(t, contexts[0].SourceFile)))
					if !ok {
						t.Fatalf("%s: GetExportMap returned no map", entry)
					}
					for _, name := range []string{"Button", "other"} {
						meta := exports.Get(name)
						if meta == nil || meta.Namespace != nil {
							t.Fatalf("%s: expected unresolved explicit %q to stay unresolved with nil Namespace, got %#v", entry, name, meta)
						}
					}
					if exports.ReexportSource("Button") != "" {
						t.Fatalf("%s: expected unresolved module re-export Button to have empty ReexportSource, got %q", entry, exports.ReexportSource("Button"))
					}
					if exports.ReexportSource("other") == "" {
						t.Fatalf("%s: expected resolved module re-export other to keep ReexportSource", entry)
					}
					if !exports.Has("default") || exports.HasDefault() {
						t.Fatalf("%s: expected unresolved default re-export to have Has(default)=true, HasDefault()=false, got (%v, %v)",
							entry, exports.Has("default"), exports.HasDefault())
					}
					if !exports.Has("ok") {
						t.Fatalf("%s: expected star-forwarded export ok to remain visible", entry)
					}
				}
			})
		})

		t.Run("256-alias 2-module cycle resolves all aliases with linear specifier checks", func(t *testing.T) {
			t.Parallel()

			buildAliasCycle := func(aliasCount int, v0First, namespaceValue bool) map[string]string {
				var aSpecs, bSpecs strings.Builder
				for i := range aliasCount {
					if i%2 == 0 {
						if bSpecs.Len() > 0 {
							bSpecs.WriteString(", ")
						}
						fmt.Fprintf(&bSpecs, "v%d as v%d", i, i+1)
					} else {
						if aSpecs.Len() > 0 {
							aSpecs.WriteString(", ")
						}
						fmt.Fprintf(&aSpecs, "v%d as v%d", i, i+1)
					}
				}

				var aDecl string
				if namespaceValue {
					aDecl = "import * as leaf from './leaf';\nexport { leaf as v0 };\n"
				} else {
					aDecl = "export const v0 = 42;\n"
				}
				aReexport := fmt.Sprintf("export { %s } from './b';\n", aSpecs.String())
				aCode := aDecl + aReexport
				if !v0First {
					if namespaceValue {
						aCode = "import * as leaf from './leaf';\n" + aReexport + "export { leaf as v0 };\n"
					} else {
						aCode = aReexport + aDecl
					}
				}
				files := map[string]string{
					"consumer-a.ts": `import * as a from './a';`,
					"consumer-b.ts": `import * as b from './b';`,
					"a.ts":          aCode,
					"b.ts":          fmt.Sprintf("export { %s } from './a';\n", bSpecs.String()),
				}
				if namespaceValue {
					files["leaf.ts"] = `export const leafProp = 42;`
				}
				return files
			}

			for _, tc := range []struct {
				name    string
				v0First bool
			}{
				{name: "v0 before re-exports", v0First: true},
				{name: "v0 after re-exports", v0First: false},
			} {
				t.Run(tc.name, func(t *testing.T) {
					t.Parallel()

					const aliasCount = 256
					files := buildAliasCycle(aliasCount, tc.v0First, true)

					// Verify cold query from both a.ts and b.ts resolves all 256 aliases
					// to leaf's Namespace in O(N) specifier checks (<= 2*N vs 33,280 O(N^2)).
					for _, entry := range []string{"consumer-a.ts", "consumer-b.ts"} {
						contexts := contextsForFiles(t, files, entry)
						exports, stats, ok := import_utils.GetExportMapWithStats(contexts[0], modules.SourceFromSpecifier(firstImportSpecifier(t, contexts[0].SourceFile)))
						if !ok {
							t.Fatalf("%s: GetExportMapWithStats returned no map", entry)
						}
						if stats.BuildCount != 3 {
							t.Fatalf("%s: BuildCount = %d, want 3 (a.ts, b.ts, leaf.ts)", entry, stats.BuildCount)
						}
						if maxChecks := 2 * aliasCount; stats.NamedSpecifierChecks > maxChecks {
							t.Fatalf("%s: NamedSpecifierChecks = %d, want <= %d (must scale linearly, not quadratically)",
								entry, stats.NamedSpecifierChecks, maxChecks)
						}
						start := 0
						if entry == "consumer-b.ts" {
							start = 1
						}
						for i := start; i <= aliasCount; i += 2 {
							name := fmt.Sprintf("v%d", i)
							meta := exports.Get(name)
							if meta == nil || meta.Namespace == nil || !meta.Namespace.Has("leafProp") || meta.Namespace.Has("missing") {
								t.Fatalf("%s: alias %q failed to resolve leaf Namespace, got %#v", entry, name, meta)
							}
						}
					}
				})
			}

			t.Run("plain value 2-module chain scales linearly across 16, 64, and 256 aliases", func(t *testing.T) {
				t.Parallel()

				for _, aliasCount := range []int{16, 64, 256} {
					files := buildAliasCycle(aliasCount, true, false)
					contexts := contextsForFiles(t, files, "consumer-a.ts")
					exports, stats, ok := import_utils.GetExportMapWithStats(contexts[0], modules.SourceFromSpecifier(firstImportSpecifier(t, contexts[0].SourceFile)))
					if !ok {
						t.Fatalf("N=%d: GetExportMapWithStats returned no map", aliasCount)
					}
					if stats.BuildCount != 2 {
						t.Fatalf("N=%d: BuildCount = %d, want 2", aliasCount, stats.BuildCount)
					}
					if maxChecks := 2 * aliasCount; stats.NamedSpecifierChecks > maxChecks {
						t.Fatalf("N=%d: NamedSpecifierChecks = %d, want <= %d", aliasCount, stats.NamedSpecifierChecks, maxChecks)
					}
					if !exports.Has("v0") || !exports.Has(fmt.Sprintf("v%d", aliasCount)) {
						t.Fatalf("N=%d: missing expected boundary exports", aliasCount)
					}
				}
			})

			t.Run("cyclic mutual star re-exports with shared base skip no-op queue events", func(t *testing.T) {
				t.Parallel()

				const sharedCount = 256
				var baseSrc strings.Builder
				for i := range sharedCount {
					fmt.Fprintf(&baseSrc, "export const e%d = %d;\n", i, i)
				}
				files := map[string]string{
					"consumer.ts": `import * as a from './a';`,
					"base.ts":     baseSrc.String(),
					"a.ts":        "export * from './base';\nexport * from './b';\n",
					"b.ts":        "export * from './base';\nexport * from './a';\n",
				}
				contexts := contextsForFiles(t, files, "consumer.ts")
				spec := modules.SourceFromSpecifier(firstImportSpecifier(t, contexts[0].SourceFile))
				// a.ts populates base.ts in the index before b.ts reuses it; both
				// modules already hold every shared export at equal rank when
				// convergeSCC starts, so no initial star events should be queued.
				exports, stats, ok := import_utils.GetExportMapWithStats(contexts[0], spec)
				if !ok {
					t.Fatal("GetExportMapWithStats returned no map")
				}
				if exports.Size() != sharedCount || !exports.Has("e0") || !exports.Has("e255") {
					t.Fatalf("unexpected export map size=%d", exports.Size())
				}
				if stats.NamedSpecifierChecks != 0 {
					t.Fatalf("NamedSpecifierChecks = %d, want 0", stats.NamedSpecifierChecks)
				}
				if stats.SCCQueueEvents != 0 {
					t.Fatalf("SCCQueueEvents = %d, want 0 (initial no-op star events should be skipped)", stats.SCCQueueEvents)
				}
			})

			t.Run("already-resolved cyclic named re-exports skip initial queue events", func(t *testing.T) {
				t.Parallel()

				const sharedCount = 512
				var baseSrc, aReexport, bReexport strings.Builder
				for i := range sharedCount {
					fmt.Fprintf(&baseSrc, "export const v%d = %d;\n", i, i)
					if i > 0 {
						aReexport.WriteString(", ")
						bReexport.WriteString(", ")
					}
					fmt.Fprintf(&aReexport, "v%d as other%d", i, i)
					fmt.Fprintf(&bReexport, "v%d as other%d", i, i)
				}
				files := map[string]string{
					"consumer.ts": `import * as a from './a';`,
					"base.ts":     baseSrc.String(),
					"a.ts":        fmt.Sprintf("export * from './base';\nexport { %s } from './b';\n", aReexport.String()),
					"b.ts":        fmt.Sprintf("export * from './base';\nexport { %s } from './a';\n", bReexport.String()),
				}
				contexts := contextsForFiles(t, files, "consumer.ts")
				spec := modules.SourceFromSpecifier(firstImportSpecifier(t, contexts[0].SourceFile))
				exports, stats, ok := import_utils.GetExportMapWithStats(contexts[0], spec)
				if !ok {
					t.Fatal("GetExportMapWithStats returned no map")
				}
				if exports.Size() != sharedCount*2 || !exports.Has("v0") || !exports.Has("other511") {
					t.Fatalf("unexpected export map size=%d", exports.Size())
				}
				if stats.SCCQueueEvents != 0 {
					t.Fatalf("SCCQueueEvents = %d, want 0 (already-resolved named re-exports should not be enqueued)", stats.SCCQueueEvents)
				}
			})

			t.Run("conflicting star exports clear namespace in both star orders", func(t *testing.T) {
				t.Parallel()

				for _, tc := range []struct {
					name   string
					barrel string
				}{
					{name: "namespace_then_values", barrel: "export * from './namespace';\nexport * from './values';\n"},
					{name: "values_then_namespace", barrel: "export * from './values';\nexport * from './namespace';\n"},
				} {
					t.Run(tc.name, func(t *testing.T) {
						t.Parallel()

						files := map[string]string{
							"consumer.ts":  `import * as barrel from './barrel';`,
							"values.ts":    `export const api = 7;`,
							"namespace.ts": "import * as api from './values';\nexport { api };\n",
							"barrel.ts":    tc.barrel,
						}
						contexts := contextsForFiles(t, files, "consumer.ts")
						exports, ok := import_utils.GetExportMap(contexts[0], modules.SourceFromSpecifier(firstImportSpecifier(t, contexts[0].SourceFile)))
						if !ok {
							t.Fatal("GetExportMap returned no map")
						}
						if !exports.Has("api") {
							t.Fatal("expected barrel to retain api in Has()")
						}
						if meta := exports.Get("api"); meta == nil || meta.Namespace != nil {
							t.Fatalf("expected conflicting star export api to have nil Namespace, got %#v", meta)
						}
					})
				}
			})

			t.Run("binding origin distinguishes renamed local bindings, shadowing re-exports, synthetic defaults, and type-only imports", func(t *testing.T) {
				t.Parallel()

				// 1. Renamed local namespace export (`import * as leaf; export { leaf as api }`)
				//    re-exported via `import { api } from './direct'; export { api }` is the SAME binding.
				sameAliased := map[string]string{
					"consumer.ts": `import * as hub from './hub';`,
					"leaf.ts":     `export const good = 42;`,
					"direct.ts":   "import * as leaf from './leaf';\nexport { leaf as api };\n",
					"indirect.ts": "import { api } from './direct';\nexport { api };\n",
					"hub.ts":      "export * from './indirect';\nexport * from './direct';\n",
				}
				ctx := contextsForFiles(t, sameAliased, "consumer.ts")[0]
				exports, ok := import_utils.GetExportMap(ctx, modules.SourceFromSpecifier(firstImportSpecifier(t, ctx.SourceFile)))
				if !ok {
					t.Fatal("sameAliased: GetExportMap failed")
				}
				if meta := exports.Get("api"); meta == nil || meta.Namespace == nil || !meta.Namespace.Has("good") {
					t.Fatalf("sameAliased: expected hub.api to retain leaf Namespace, got %#v", meta)
				}

				// 2. Local `leaf` renamed to `api` alongside `export { leaf } from './other'`
				//    must NOT falsely match `export { leaf as api } from './other'`.
				shadowedExport := map[string]string{
					"consumer.ts": `import * as barrel from './barrel';`,
					"leaf.ts":     `export const good = 42;`,
					"other.ts":    `export const leaf = 7;`,
					"direct.ts":   "import * as leaf from './leaf';\nexport { leaf as api };\nexport { leaf } from './other';\n",
					"values.ts":   "export { leaf as api } from './other';\n",
					"barrel.ts":   "export * from './direct';\nexport * from './values';\n",
				}
				ctx = contextsForFiles(t, shadowedExport, "consumer.ts")[0]
				exports, ok = import_utils.GetExportMap(ctx, modules.SourceFromSpecifier(firstImportSpecifier(t, ctx.SourceFile)))
				if !ok {
					t.Fatal("shadowedExport: GetExportMap failed")
				}
				if meta := exports.Get("api"); meta == nil || meta.Namespace != nil {
					t.Fatalf("shadowedExport: expected conflicting api to clear Namespace, got %#v", meta)
				}

				// 3. Synthetic default star export (`export =`) colliding with unrelated namespace export.
				for _, barrel := range []string{
					"export * from './values';\nexport * from './namespace';\n",
					"export * from './namespace';\nexport * from './values';\n",
				} {
					syntheticConflict := map[string]string{
						"consumer.ts":  `import * as barrel from './barrel';`,
						"cjs.ts":       "const obj = { good: 42 };\nexport = obj;\n",
						"values.ts":    `export { default as api } from './cjs';`,
						"leaf.ts":      `export const other = 1;`,
						"namespace.ts": "import * as api from './leaf';\nexport { api };\n",
						"barrel.ts":    barrel,
					}
					ctx = contextsForFiles(t, syntheticConflict, "consumer.ts")[0]
					exports, ok = import_utils.GetExportMap(ctx, modules.SourceFromSpecifier(firstImportSpecifier(t, ctx.SourceFile)))
					if !ok {
						t.Fatal("syntheticConflict: GetExportMap failed")
					}
					if meta := exports.Get("api"); meta == nil || meta.Namespace != nil {
						t.Fatalf("syntheticConflict: expected synthetic vs namespace star collision to clear Namespace, got %#v", meta)
					}
				}

				// 4. Type-only import shadowed by local value must not match the imported namespace binding.
				typeOnlyShadow := map[string]string{
					"consumer.ts":  `import * as barrel from './barrel';`,
					"leaf.ts":      `export const good = 42;`,
					"namespace.ts": "import * as api from './leaf';\nexport { api };\n",
					"shadowed.ts":  "import type { api } from './namespace';\nconst api = 7;\nexport { api };\n",
					"barrel.ts":    "export * from './namespace';\nexport * from './shadowed';\n",
				}
				ctx = contextsForFiles(t, typeOnlyShadow, "consumer.ts")[0]
				exports, ok = import_utils.GetExportMap(ctx, modules.SourceFromSpecifier(firstImportSpecifier(t, ctx.SourceFile)))
				if !ok {
					t.Fatal("typeOnlyShadow: GetExportMap failed")
				}
				if meta := exports.Get("api"); meta == nil || meta.Namespace != nil {
					t.Fatalf("typeOnlyShadow: expected type-only-shadowed local value to conflict with namespace export, got %#v", meta)
				}

				// 5. Cyclic named re-export upgrades across same-file overwrite from plain to namespace.
				cyclicOverwrite := map[string]string{
					"consumer-a.ts": `import * as a from './a';`,
					"consumer-b.ts": `import * as b from './b';`,
					"leaf.ts":       `export const good = 42;`,
					"plain.ts":      `export const api = 1;`,
					"ns.ts":         "import * as api from './leaf';\nexport { api };\n",
					"b.ts":          "export { api } from './plain';\nexport { x } from './a';\nexport { api } from './ns';\n",
					"a.ts":          "export const x = 1;\nexport { api } from './b';\n",
				}
				for _, entry := range []string{"consumer-a.ts", "consumer-b.ts"} {
					ctx = contextsForFiles(t, cyclicOverwrite, entry)[0]
					exports, ok = import_utils.GetExportMap(ctx, modules.SourceFromSpecifier(firstImportSpecifier(t, ctx.SourceFile)))
					if !ok {
						t.Fatalf("cyclicOverwrite(%s): GetExportMap failed", entry)
					}
					if meta := exports.Get("api"); meta == nil || meta.Namespace == nil || !meta.Namespace.Has("good") {
						t.Fatalf("cyclicOverwrite(%s): expected api to converge to leaf Namespace, got %#v", entry, meta)
					}
				}

				// 6. `export * as local from './other'` must NOT collide with local
				//    `import * as local from './leaf'; export { local as api }` as the same binding.
				for _, tc := range []struct {
					name   string
					alias  string
					barrel string
				}{
					{
						name:   "direct_reexport_source_first",
						alias:  "export { local as api } from './source';\n",
						barrel: "export * from './source';\nexport * from './alias';\n",
					},
					{
						name:   "direct_reexport_alias_first",
						alias:  "export { local as api } from './source';\n",
						barrel: "export * from './alias';\nexport * from './source';\n",
					},
					{
						name:   "local_import_reexport",
						alias:  "import { local } from './source';\nexport { local as api };\n",
						barrel: "export * from './source';\nexport * from './alias';\n",
					},
				} {
					nsReexportConflict := map[string]string{
						"consumer.ts": `import * as barrel from './barrel';`,
						"leaf.ts":     `export const left = 1;`,
						"other.ts":    `export const right = 2;`,
						"source.ts":   "import * as local from './leaf';\nexport { local as api };\nexport * as local from './other';\n",
						"alias.ts":    tc.alias,
						"barrel.ts":   tc.barrel,
					}
					ctx = contextsForFiles(t, nsReexportConflict, "consumer.ts")[0]
					exports, ok = import_utils.GetExportMap(ctx, modules.SourceFromSpecifier(firstImportSpecifier(t, ctx.SourceFile)))
					if !ok {
						t.Fatalf("nsReexportConflict(%s): GetExportMap failed", tc.name)
					}
					if meta := exports.Get("api"); meta == nil || meta.Namespace != nil {
						t.Fatalf("nsReexportConflict(%s): expected distinct local import vs namespace re-export to conflict and clear Namespace, got %#v", tc.name, meta)
					}
				}

				// 7. Named `export default function run()` shares its binding with `export { run }`,
				//    so star-merging `{ default as run }` and `export { run }` preserves `@deprecated`
				//    in both star orders.
				for _, tc := range []struct {
					name   string
					barrel string
				}{
					{name: "first_api_second", barrel: "export * from './first';\nexport * from './api';\nexport * from './second';\n"},
					{name: "api_first_second", barrel: "export * from './api';\nexport * from './first';\nexport * from './second';\n"},
				} {
					namedDefaultDoc := map[string]string{
						"consumer.ts": `import { run } from './barrel';`,
						"api.ts":      "/** @deprecated Use fresh. */\nexport default function run() { return 42; }\nexport { run };\n",
						"first.ts":    `export { default as run } from './api';`,
						"second.ts":   `export { default as run } from './api';`,
						"barrel.ts":   tc.barrel,
					}
					ctx = contextsForFiles(t, namedDefaultDoc, "consumer.ts")[0]
					exports, ok = import_utils.GetExportMap(ctx, modules.SourceFromSpecifier(firstImportSpecifier(t, ctx.SourceFile)))
					if !ok {
						t.Fatalf("namedDefaultDoc(%s): GetExportMap failed", tc.name)
					}
					if meta := exports.Get("run"); meta == nil {
						t.Fatalf("namedDefaultDoc(%s): expected barrel.run export", tc.name)
					} else if dep := meta.Deprecation(ctx); dep == nil || !strings.Contains(dep.Description, "Use fresh.") {
						t.Fatalf("namedDefaultDoc(%s): expected barrel.run to keep @deprecated doc, got %#v", tc.name, dep)
					}
				}

				// 8. Default-only / unresolved `export { default as local } from './data.json'`
				//    must NOT collide with local `import * as local from './leaf'; export { local as api }`.
				externalReexportConflict := map[string]string{
					"consumer.ts": `import * as barrel from './barrel';`,
					"leaf.ts":     `export const left = 1;`,
					"data.json":   `{"x": 1}`,
					"source.ts":   "import * as local from './leaf';\nexport { local as api };\nexport { default as local } from './data.json';\n",
					"alias.ts":    "import { local } from './source';\nexport { local as api };\n",
					"barrel.ts":   "export * from './source';\nexport * from './alias';\n",
				}
				ctx = contextsForFiles(t, externalReexportConflict, "consumer.ts")[0]
				exports, ok = import_utils.GetExportMap(ctx, modules.SourceFromSpecifier(firstImportSpecifier(t, ctx.SourceFile)))
				if !ok {
					t.Fatal("externalReexportConflict: GetExportMap failed")
				}
				if meta := exports.Get("api"); meta == nil || meta.Namespace != nil {
					t.Fatalf("externalReexportConflict: expected default-only re-export vs local import to conflict and clear Namespace, got %#v", meta)
				}
			})

			t.Run("wide barrel origin resolution indexes local steps in linear time", func(t *testing.T) {
				t.Parallel()

				const count = 2048
				var src, impNames, expNames strings.Builder
				for i := range count {
					fmt.Fprintf(&src, "export const v%d = %d;\n", i, i)
					if i > 0 {
						impNames.WriteString(", ")
						expNames.WriteString(", ")
					}
					fmt.Fprintf(&impNames, "v%d", i)
					fmt.Fprintf(&expNames, "v%d", i)
				}
				files := map[string]string{
					"consumer.ts": `import * as barrel from './barrel';`,
					"source.ts":   src.String(),
					"reexport.ts": fmt.Sprintf("import { %s } from './source';\nexport { %s };\n", impNames.String(), expNames.String()),
					"barrel.ts":   "export * from './source';\nexport * from './reexport';\n",
				}
				ctx := contextsForFiles(t, files, "consumer.ts")[0]
				exports, stats, ok := import_utils.GetExportMapWithStats(ctx, modules.SourceFromSpecifier(firstImportSpecifier(t, ctx.SourceFile)))
				if !ok {
					t.Fatal("GetExportMapWithStats failed")
				}
				if exports.Size() != count || !exports.Has("v0") || !exports.Has("v2047") {
					t.Fatalf("unexpected wide barrel size=%d", exports.Size())
				}
				// Each step of source.ts should be indexed at most once (2,048 scans instead of ~2,098,176).
				if stats.OriginStepScans > count*2 {
					t.Fatalf("OriginStepScans = %d, want <= %d (O(N) indexing)", stats.OriginStepScans, count*2)
				}
			})
		})
	})
}
