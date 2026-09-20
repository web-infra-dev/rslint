package require_dynamic_import_entry

import (
	"fmt"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/testutil/txtarfs"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestRequireDynamicImportEntry(t *testing.T) {
	base := fixtures.GetRootDir()
	directory := tspath.ResolvePath(base.Dir, "rslim-fixtures")
	archive := txtarfs.MustParseFile(t, "testdata/modules.txtar")
	names, err := archive.FileNames("")
	if err != nil {
		t.Fatal(err)
	}
	files := map[string]string{}
	for _, name := range names {
		data, err := archive.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		files[tspath.ResolvePath(directory, name)] = string(data)
	}
	root := rule_tester.Root{Dir: directory, FS: utils.NewOverlayVFS(base.FS, files)}
	valid := []rule_tester.ValidTestCase{
		{Code: `export const { start } = await import('./partial');`},
		{Code: `consume(await import('./more-annotated'));`},
		{Code: `consume(await import('./default-alias'));`},
		{Code: `consume(await import('./namespace'));`},
		{Code: `consume(await import('./ambient'));`},
		{Code: "const { start } = await import(`./annotated`); start();"},
		{Code: `const { start } = await import('./annotated'); start();`},
		{Code: `consume(await import('./annotated'));`},
		{Code: `const { start } = await import('./partial'); start();`},
		{Code: `const mod = await import('./partial'); mod.start();`},
		{Code: `import('./partial').then(({ start }) => start());`},
		{Code: `import('./partial').then(mod => mod.start());`},
		{Code: `const pending = import('./partial'); const mod = await pending; mod.start();`},
		{Code: `const mod = await import('./alias'); mod.run();`},
		{Code: `const mod = await import('./reexport'); mod.run();`},
		{Code: `consume(await import('./default-annotated'));`},
		{Code: `consume(await import('./types'));`},
		{Code: `consume(await import('./declarations'));`},
		{Code: `consume(await import('dependency'));`},
		{Code: `consume(await import('./inline'));`},
		{Code: `consume(await import('./overload'));`},
		// Rslim also accepts trailing trivia at the declaration's full start.
		{Code: `consume(await import('./trailing'));`},
		{Code: `consume(await import('./empty'));`},
		{Code: `import('./feature');`},
		{Code: `await import('./feature');`},
		{Code: `void import('./feature');`},
		{Code: `import('./feature').then(() => {});`},
		{Code: `import { start } from './feature'; start();`},
		{Code: `type Module = typeof import('./feature');`},
		// The rule tester registers the rule under the name "test".
		{Code: "// rslint-disable-next-line test -- Manually checked @entry.\nconsume(await import(path));"},
		{Code: "// rslint-disable-next-line test -- Manually checked @entry.\nconsume(await import('./missing'));"},
		{Code: "consume(await import(\n// rslint-disable-next-line test -- Manually checked @entry.\npath\n));"},
	}
	invalid := []rule_tester.InvalidTestCase{}
	for _, specifier := range []string{"'./missing'", "`./missing`", "path", "`./${path}`", "'./' + path", "getPath()"} {
		invalid = append(invalid, rule_tester.InvalidTestCase{
			Code: "consume(await import(" + specifier + "));",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "unresolvedImport",
				Message:   "Cannot resolve this dynamic import. Manually check that the imported declarations used at runtime have @entry, then add // rslint-disable-next-line rslim/require-dynamic-import-entry before this argument's line to ignore this diagnostic.",
				Line:      1, Column: 22, EndLine: 1, EndColumn: 22 + len(specifier),
			}},
		})
	}
	invalid = append(invalid,
		rule_tester.InvalidTestCase{Code: "import(path);", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unresolvedImport"}}},
		rule_tester.InvalidTestCase{Code: "consume(await import(\n  path\n));", Errors: []rule_tester.InvalidTestCaseError{{
			MessageId: "unresolvedImport", Line: 2, Column: 3, EndLine: 2, EndColumn: 7,
		}}},
	)
	add := func(code, module, missing string) {
		t.Helper()
		start := strings.Index(code, "'"+module+"'")
		if start < 0 {
			t.Fatalf("module %q missing from %q", module, code)
		}
		line := strings.Count(code[:start], "\n") + 1
		column := start - strings.LastIndex(code[:start], "\n")
		invalid = append(invalid, rule_tester.InvalidTestCase{Code: code, Errors: []rule_tester.InvalidTestCaseError{{
			MessageId: "missingEntry", Message: fmt.Sprintf("Add @entry to the declarations of these dynamically imported exports from %q: %s.", module, missing),
			Line: line, Column: column, EndLine: line, EndColumn: column + len(module) + 2,
		}}})
	}
	add(`const { start } = await import('./feature'); start();`, "./feature", "start")
	add(`const { start: run = fallback } = await import('./feature'); run();`, "./feature", "start")
	add(`(await import('./feature')).start();`, "./feature", "start")
	add(`(await import('./feature'))['start']();`, "./feature", "start")
	add(`const mod = await import('./feature'); mod.start();`, "./feature", "start")
	add(`const pending = import('./feature'); const mod = await pending; mod.start();`, "./feature", "start")
	add(`import('./feature').then(({ start }) => start());`, "./feature", "start")
	add(`import('./feature').then(function (mod) { mod.start(); });`, "./feature", "start")
	add(`const mod = (await import('./feature'))!; const alias = mod; alias.start();`, "./feature", "start")
	add(`consume(await import('./feature'));`, "./feature", "start, unused")
	add(`export const mod = await import('./feature');`, "./feature", "start, unused")
	add(`consume(await import('./missing-namespace'));`, "./missing-namespace", "feature")
	add(`consume(await import('./cycle-a'));`, "./cycle-a", "next")
	add(`const mod = await import('./feature'); mod[key]();`, "./feature", "start, unused")
	add(`const { ...rest } = await import('./feature'); consume(rest);`, "./feature", "start, unused")
	add(`import('./feature').then(consume);`, "./feature", "start, unused")
	add(`const { stop } = await import('./partial'); stop();`, "./partial", "stop")
	add(`const { run } = await import('./missing-reexport'); run();`, "./missing-reexport", "run")
	add(`const { start } = await import('./star'); start();`, "./star", "start")
	add(`const { default: run } = await import('./default'); run();`, "./default", "default")
	add(`consume(await import('./same-line'));`, "./same-line", "start")
	add(`consume(await import('./misleading'));`, "./misleading", "start, value")
	add(`consume(await import('./misplaced'));`, "./misplaced", "start")
	add("// @entry\nconst { start } = await import('./feature'); start();", "./feature", "start")
	add("const { start } = await import(\n  './feature'\n); start();", "./feature", "start")
	add(`const { start } = await import('./feature', { with: { type: 'javascript' } }); start();`, "./feature", "start")
	rule_tester.RunRuleTester(root, "tsconfig.json", t, &RequireDynamicImportEntryRule, valid, invalid)
}
