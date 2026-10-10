package no_deprecated_test

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/no_deprecated"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/testutil/txtarfs"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func deprecatedRoot(t *testing.T, archiveName string) rule_tester.Root {
	t.Helper()
	root := fixtures.GetRootDir()
	root.Dir = tspath.ResolvePath(root.Dir, "no-deprecated-rule")
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

// All v2.32.0 tests/src/rules/no-deprecated.js groups, expanded SYNTAX_CASES,
// and docs/rules/no-deprecated.md examples. Ranges and messages are recorded
// from the pinned upstream rule. Upstream has literal messages without IDs.
func TestNoDeprecatedUpstream(t *testing.T) {
	t.Run("no-deprecated", func(t *testing.T) {
		rule_tester.RunRuleTester(deprecatedRoot(t, "testdata/upstream.txtar"), "tsconfig.json", t, &no_deprecated.NoDeprecatedRule, []rule_tester.ValidTestCase{
			{Code: "import { x } from './fake' "},
			{Code: "import bar from './bar'"},
			{Code: "import { fine } from './deprecated'"},
			{Code: "import { _undocumented } from './deprecated'"},
			{Code: "import { fn } from './deprecated'", Settings: map[string]any{"import/docstyle": []any{"tomdoc"}}},
			{Code: "import { fine } from './tomdoc-deprecated'", Settings: map[string]any{"import/docstyle": []any{"tomdoc"}}},
			{Code: "import { _undocumented } from './tomdoc-deprecated'", Settings: map[string]any{"import/docstyle": []any{"tomdoc"}}},
			{Code: "import * as depd from './deprecated'"},
			{Code: "import * as depd from './deprecated'; console.log(depd.fine())"},
			{Code: "import { deepDep } from './deep-deprecated'"},
			{Code: "import { deepDep } from './deep-deprecated'; console.log(deepDep.fine())"},
			{Code: "import { deepDep } from './deep-deprecated'; function x(deepDep) { console.log(deepDep.MY_TERRIBLE_ACTION) }"},
			{Code: "for (let { foo, bar } of baz) {}"},
			{Code: "for (let [ foo, bar ] of baz) {}"},
			{Code: "const { x, y } = bar"},
			{Code: "const { x, y, ...z } = bar"},
			{Code: "let x; export { x }"},
			{Code: "let x; export { x as y }"},
			{Code: "export const x = null"},
			{Code: "export var x = null"},
			{Code: "export let x = null"},
			{Code: "export default x"},
			{Code: "export default class x {}"},
			{Code: "import json from \"./data.json\"", Settings: map[string]any{"import/extensions": []any{".js"}}},
			{Code: "import foo from \"./foobar.json\";", Settings: map[string]any{"import/extensions": []any{".js"}}},
			{Code: "import foo from \"./foobar\";", Settings: map[string]any{"import/extensions": []any{".js"}}},
			{Code: "import { foo } from \"./issue-370-commonjs-namespace/bar\"", Settings: map[string]any{"import/ignore": []any{"foo"}}},
			{Code: "export * from \"./issue-370-commonjs-namespace/bar\"", Settings: map[string]any{"import/ignore": []any{"foo"}}},
			{Code: "import * as a from \"./commonjs-namespace/a\"; a.b"},
			{Code: "import { foo } from \"./ignore.invalid.extension\""},
		}, []rule_tester.InvalidTestCase{
			// The shared import map does not translate dependency parser errors.
			{Code: "import './malformed.js'", Skip: true},
			{Code: "import { fn } from './deprecated'", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "Deprecated: please use 'x' instead.", Line: 1, Column: 10, EndLine: 1, EndColumn: 12}}},
			{Code: "import TerribleClass from './deprecated'", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "Deprecated: this is awful, use NotAsBadClass.", Line: 1, Column: 8, EndLine: 1, EndColumn: 21}}},
			{Code: "import { MY_TERRIBLE_ACTION } from './deprecated'", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "Deprecated: please stop sending/handling this action type.", Line: 1, Column: 10, EndLine: 1, EndColumn: 28}}},
			{Code: "import { fn } from './deprecated'", Settings: map[string]any{"import/docstyle": []any{"jsdoc", "tomdoc"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "Deprecated: please use 'x' instead.", Line: 1, Column: 10, EndLine: 1, EndColumn: 12}}},
			{Code: "import { fn } from './tomdoc-deprecated'", Settings: map[string]any{"import/docstyle": []any{"tomdoc"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "Deprecated: This function is terrible.", Line: 1, Column: 10, EndLine: 1, EndColumn: 12}}},
			{Code: "import TerribleClass from './tomdoc-deprecated'", Settings: map[string]any{"import/docstyle": []any{"tomdoc"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "Deprecated: this is awful, use NotAsBadClass.", Line: 1, Column: 8, EndLine: 1, EndColumn: 21}}},
			{Code: "import { MY_TERRIBLE_ACTION } from './tomdoc-deprecated'", Settings: map[string]any{"import/docstyle": []any{"tomdoc"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "Deprecated: Please stop sending/handling this action type.", Line: 1, Column: 10, EndLine: 1, EndColumn: 28}}},
			{Code: "import { MY_TERRIBLE_ACTION } from './deprecated'; function shadow(MY_TERRIBLE_ACTION) { console.log(MY_TERRIBLE_ACTION); }", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "Deprecated: please stop sending/handling this action type.", Line: 1, Column: 10, EndLine: 1, EndColumn: 28}}},
			{Code: "import { MY_TERRIBLE_ACTION, fine } from './deprecated'; console.log(fine)", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "Deprecated: please stop sending/handling this action type.", Line: 1, Column: 10, EndLine: 1, EndColumn: 28}}},
			{Code: "import { MY_TERRIBLE_ACTION } from './deprecated'; console.log(MY_TERRIBLE_ACTION)", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "Deprecated: please stop sending/handling this action type.", Line: 1, Column: 10, EndLine: 1, EndColumn: 28}, {MessageId: "deprecated", Message: "Deprecated: please stop sending/handling this action type.", Line: 1, Column: 64, EndLine: 1, EndColumn: 82}}},
			{Code: "import { MY_TERRIBLE_ACTION } from './deprecated'; console.log(someOther.MY_TERRIBLE_ACTION)", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "Deprecated: please stop sending/handling this action type.", Line: 1, Column: 10, EndLine: 1, EndColumn: 28}}},
			{Code: "import { MY_TERRIBLE_ACTION } from './deprecated'; console.log(MY_TERRIBLE_ACTION.whatever())", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "Deprecated: please stop sending/handling this action type.", Line: 1, Column: 10, EndLine: 1, EndColumn: 28}, {MessageId: "deprecated", Message: "Deprecated: please stop sending/handling this action type.", Line: 1, Column: 64, EndLine: 1, EndColumn: 82}}},
			{Code: "import { MY_TERRIBLE_ACTION } from './deprecated'; console.log(MY_TERRIBLE_ACTION(this, is, the, worst))", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "Deprecated: please stop sending/handling this action type.", Line: 1, Column: 10, EndLine: 1, EndColumn: 28}, {MessageId: "deprecated", Message: "Deprecated: please stop sending/handling this action type.", Line: 1, Column: 64, EndLine: 1, EndColumn: 82}}},
			{Code: "import Thing from './deprecated-file'", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "Deprecated: this module is the worst.", Line: 1, Column: 1, EndLine: 1, EndColumn: 38}}},
			{Code: "import Thing from './deprecated-file'; console.log(other.Thing)", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "Deprecated: this module is the worst.", Line: 1, Column: 1, EndLine: 1, EndColumn: 39}}},
			{Code: "import * as depd from './deprecated'; console.log(depd.MY_TERRIBLE_ACTION)", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "Deprecated: please stop sending/handling this action type.", Line: 1, Column: 56, EndLine: 1, EndColumn: 74}}},
			{Code: "import * as deep from './deep-deprecated'; console.log(deep.deepDep.MY_TERRIBLE_ACTION)", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "Deprecated: please stop sending/handling this action type.", Line: 1, Column: 69, EndLine: 1, EndColumn: 87}}},
			{Code: "import { deepDep } from './deep-deprecated'; console.log(deepDep.MY_TERRIBLE_ACTION)", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "Deprecated: please stop sending/handling this action type.", Line: 1, Column: 66, EndLine: 1, EndColumn: 84}}},
			{Code: "import { deepDep } from './deep-deprecated'; function x(deepNDep) { console.log(deepDep.MY_TERRIBLE_ACTION) }", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "Deprecated: please stop sending/handling this action type.", Line: 1, Column: 89, EndLine: 1, EndColumn: 107}}},
		})
	})
	t.Run("no-deprecated: hoisting", func(t *testing.T) {
		rule_tester.RunRuleTester(deprecatedRoot(t, "testdata/upstream.txtar"), "tsconfig.json", t, &no_deprecated.NoDeprecatedRule, []rule_tester.ValidTestCase{
			{Code: "function x(deepDep) { console.log(deepDep.MY_TERRIBLE_ACTION) } import { deepDep } from './deep-deprecated'"},
		}, []rule_tester.InvalidTestCase{
			{Code: "console.log(MY_TERRIBLE_ACTION); import { MY_TERRIBLE_ACTION } from './deprecated'", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "Deprecated: please stop sending/handling this action type.", Line: 1, Column: 13, EndLine: 1, EndColumn: 31}, {MessageId: "deprecated", Message: "Deprecated: please stop sending/handling this action type.", Line: 1, Column: 43, EndLine: 1, EndColumn: 61}}},
		})
	})
	t.Run("typescript", func(t *testing.T) {
		rule_tester.RunRuleTester(deprecatedRoot(t, "testdata/upstream.txtar"), "tsconfig.json", t, &no_deprecated.NoDeprecatedRule, []rule_tester.ValidTestCase{
			{Code: "import * as hasDeprecated from './ts-deprecated.ts'"},
		}, []rule_tester.InvalidTestCase{
			{Code: "import { foo } from './ts-deprecated.ts'; console.log(foo())", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "Deprecated: don't use this!", Line: 1, Column: 10, EndLine: 1, EndColumn: 13}, {MessageId: "deprecated", Message: "Deprecated: don't use this!", Line: 1, Column: 55, EndLine: 1, EndColumn: 58}}},
		})
	})
	t.Run("documentation", func(t *testing.T) {
		rule_tester.RunRuleTester(deprecatedRoot(t, "testdata/upstream.txtar"), "tsconfig.json", t, &no_deprecated.NoDeprecatedRule, []rule_tester.ValidTestCase{
			{Code: "\n// @file: ./answer.js\n\n/**\n * this is what you get when you trust a mouse talk show\n * @deprecated need to restart the experiment\n * @returns {Number} nonsense\n */\nexport function multiply(six, nine) {\n  return 42\n}\n"},
			{Code: "// Deprecated: This is what you get when you trust a mouse talk show, need to\n// restart the experiment.\n//\n// Returns a Number nonsense\nexport function multiply(six, nine) { return 42 }"},
		}, []rule_tester.InvalidTestCase{
			{Code: "import { multiply } from './answer'\nfunction whatever(y, z) { return multiply(y, z) }", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "Deprecated: need to restart the experiment", Line: 1, Column: 10, EndLine: 1, EndColumn: 18}, {MessageId: "deprecated", Message: "Deprecated: need to restart the experiment", Line: 2, Column: 34, EndLine: 2, EndColumn: 42}}},
			{Code: "import { multiply } from './tomdoc-answer'\nfunction whatever(y, z) { return multiply(y, z) }", Settings: map[string]any{"import/docstyle": []any{"jsdoc", "tomdoc"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "Deprecated: This is what you get when you trust a mouse talk show, need to restart the experiment.", Line: 1, Column: 10, EndLine: 1, EndColumn: 18}, {MessageId: "deprecated", Message: "Deprecated: This is what you get when you trust a mouse talk show, need to restart the experiment.", Line: 2, Column: 34, EndLine: 2, EndColumn: 42}}},
		})
	})
}
