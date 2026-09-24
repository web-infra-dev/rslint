package no_named_as_default_test

import (
	"fmt"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/no_named_as_default"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/testutil/txtarfs"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func namedDefaultRoot(t *testing.T, archiveName string) rule_tester.Root {
	t.Helper()
	root := fixtures.GetRootDir()
	root.Dir = tspath.ResolvePath(root.Dir, "no-named-as-default-rule")
	archive := txtarfs.MustParseFile(t, archiveName)
	names, err := archive.FileNames("")
	if err != nil {
		t.Fatal(err)
	}
	if len(names) == 0 {
		t.Fatal("fixture archive is empty")
	}
	files := make(map[string]string, len(names))
	for _, name := range names {
		data, err := archive.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		files[tspath.ResolvePath(root.Dir, name)] = string(data)
	}
	root.FS = utils.NewOverlayVFS(root.FS, files)
	return root
}

func namedDefaultError(name string, line, column, endColumn int) []rule_tester.InvalidTestCaseError {
	return []rule_tester.InvalidTestCaseError{{
		MessageId: "noNamedAsDefault",
		Message:   fmt.Sprintf("Using exported name '%s' as identifier for default import.", name),
		Line:      line, Column: column, EndLine: line, EndColumn: endColumn,
	}}
}

// Every case from eslint-plugin-import v2.32.0 tests/src/rules/no-named-as-default.js
// and docs/rules/no-named-as-default.md. SYNTAX_CASES are expanded below.
// Upstream reports literal messages, without a message ID.
func TestNoNamedAsDefaultUpstream(t *testing.T) {
	rule_tester.RunRuleTester(namedDefaultRoot(t, "testdata/upstream.txtar"), "tsconfig.json", t, &no_named_as_default.NoNamedAsDefaultRule,
		[]rule_tester.ValidTestCase{
			// The test Program rejects the malformed dependency before rules run.
			{Code: `import "./malformed.js"`, Skip: true},
			{Code: `import bar, { foo } from "./bar";`},
			{Code: `import bar, { foo } from "./empty-folder";`},
			// Babel's export-default-from proposal is not supported by tsgo.
			{Code: `export bar, { foo } from "./bar";`, Skip: true},
			{Code: `export bar from "./bar";`, Skip: true},
			{Code: `export default from "./bar";`, Skip: true}, // #566
			{Code: `import bar, { foo } from "./export-default-string-and-named"`},
			// #1594: the default and named exports are direct re-exports.
			{Code: `import something from "./no-named-as-default/re-exports.js";`},
			{Code: `import { something } from "./no-named-as-default/re-exports.js";`},
			{Code: `import myOwnNameForVariable from "./no-named-as-default/exports.js";`},
			{Code: `import { variable } from "./no-named-as-default/exports.js";`},
			{Code: `import variable from "./no-named-as-default/misleading-re-exports.js";`},
			{Code: `import foobar from "./no-named-as-default/no-default-export.js";`},
			// Same upstream cases for exports; only proposal syntax is skipped.
			{Code: `export something from "./no-named-as-default/re-exports.js";`, Skip: true},
			{Code: `export { something } from "./no-named-as-default/re-exports.js";`},
			{Code: `export myOwnNameForVariable from "./no-named-as-default/exports.js";`, Skip: true},
			{Code: `export { variable } from "./no-named-as-default/exports.js";`},
			{Code: `export variable from "./no-named-as-default/misleading-re-exports.js";`, Skip: true},
			{Code: `export foobar from "./no-named-as-default/no-default-export.js";`, Skip: true},
			// SYNTAX_CASES
			{Code: `for (let { foo, bar } of baz) {}`},
			{Code: `for (let [ foo, bar ] of baz) {}`},
			{Code: `const { x, y } = bar`},
			{Code: `const { x, y, ...z } = bar`},
			{Code: `let x; export { x }`},
			{Code: `let x; export { x as y }`},
			{Code: `export const x = null`},
			{Code: `export var x = null`},
			{Code: `export let x = null`},
			{Code: `export default x`},
			{Code: `export default class x {}`},
			{Code: `import json from "./data.json"`, Settings: map[string]interface{}{"import/extensions": []interface{}{".js"}}},
			{Code: `import foo from "./foobar.json";`, Settings: map[string]interface{}{"import/extensions": []interface{}{".js"}}},
			{Code: `import foo from "./foobar";`, Settings: map[string]interface{}{"import/extensions": []interface{}{".js"}}},
			{Code: `import { foo } from "./issue-370-commonjs-namespace/bar"`, Settings: map[string]interface{}{"import/ignore": []interface{}{"foo"}}},
			{Code: `export * from "./issue-370-commonjs-namespace/bar"`, Settings: map[string]interface{}{"import/ignore": []interface{}{"foo"}}},
			{Code: `import * as a from "./commonjs-namespace/a"; a.b`},
			{Code: `import { foo } from "./ignore.invalid.extension"`},
			// Documentation examples (the module, then its consumers).
			{Code: `export default 'foo'; export const bar = 'baz';`},
			{Code: `import foo from './foo.js';`},
			{Code: `export foo from './foo.js';`, Skip: true}, // Babel proposal
		},
		[]rule_tester.InvalidTestCase{
			{Code: `import foo from "./bar";`, Errors: namedDefaultError("foo", 1, 8, 11)},
			{Code: `import foo, { foo as bar } from "./bar";`, Errors: namedDefaultError("foo", 1, 8, 11)},
			// Babel's export-default-from proposal is not supported by tsgo.
			{Code: `export foo from "./bar";`, Skip: true, Errors: []rule_tester.InvalidTestCaseError{{Message: "Using exported name 'foo' as identifier for default export.", Line: 1, Column: 8, EndLine: 1, EndColumn: 11}}},
			{Code: `export foo, { foo as bar } from "./bar";`, Skip: true, Errors: []rule_tester.InvalidTestCaseError{{Message: "Using exported name 'foo' as identifier for default export.", Line: 1, Column: 8, EndLine: 1, EndColumn: 11}}},
			// Dependency parse errors are not converted into rule diagnostics.
			{Code: `import foo from "./malformed.js"`, Skip: true, Errors: []rule_tester.InvalidTestCaseError{{Message: "Parse errors in imported module './malformed.js': 'return' outside of function (1:1)", Line: 1, Column: 17, EndLine: 1, EndColumn: 33}}},
			{Code: `import foo from "./export-default-string-and-named"`, Errors: namedDefaultError("foo", 1, 8, 11)},
			{Code: `import foo, { foo as bar } from "./export-default-string-and-named"`, Errors: namedDefaultError("foo", 1, 8, 11)},
			{Code: `import something from "./no-named-as-default/misleading-re-exports.js";`, Errors: namedDefaultError("something", 1, 8, 17)},
			// Upstream still reports when the same value is exported locally.
			{Code: `import variable from "./no-named-as-default/exports.js";`, Errors: namedDefaultError("variable", 1, 8, 16)},
			{Code: `export variable from "./no-named-as-default/exports.js";`, Skip: true, Errors: []rule_tester.InvalidTestCaseError{{Message: "Using exported name 'variable' as identifier for default export.", Line: 1, Column: 8, EndLine: 1, EndColumn: 16}}},
			// The docs' import example labels this an export; the rule says import.
			{Code: `import bar from './foo.js';`, Errors: namedDefaultError("bar", 1, 8, 11)},
			{Code: `export bar from './foo.js';`, Skip: true, Errors: []rule_tester.InvalidTestCaseError{{Message: "Using exported name 'bar' as identifier for default export.", Line: 1, Column: 8, EndLine: 1, EndColumn: 11}}},
		},
	)
}
