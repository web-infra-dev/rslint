package max_dependencies_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/max_dependencies"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestMaxDependenciesExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.allow-js.json", t, &max_dependencies.MaxDependenciesRule,
		[]rule_tester.ValidTestCase{
			// Defaults, option boundaries, and module identity.
			{
				Code: "import \"dep0\";\nimport \"dep1\";\nimport \"dep2\";\nimport \"dep3\";\nimport \"dep4\";\nimport \"dep5\";\nimport \"dep6\";\nimport \"dep7\";\nimport \"dep8\";\nimport \"dep9\";",
			},
			{
				Code:    "import \"dep0\";\nimport \"dep1\";\nimport \"dep2\";\nimport \"dep3\";\nimport \"dep4\";\nimport \"dep5\";\nimport \"dep6\";\nimport \"dep7\";\nimport \"dep8\";\nimport \"dep9\";",
				Options: []any{map[string]any{"max": 10, "ignoreTypeImports": false}},
			},
			// Keep the documented default when max is omitted from an options object.
			{
				Code:    "import \"dep0\";\nimport \"dep1\";\nimport \"dep2\";\nimport \"dep3\";\nimport \"dep4\";\nimport \"dep5\";\nimport \"dep6\";\nimport \"dep7\";\nimport \"dep8\";\nimport \"dep9\";",
				Options: []any{map[string]any{}},
			},
			{
				Code:    "import \"dep0\";\nimport \"dep1\";\nimport \"dep2\";\nimport \"dep3\";\nimport \"dep4\";\nimport \"dep5\";\nimport \"dep6\";\nimport \"dep7\";\nimport \"dep8\";\nimport \"dep9\";\nimport type T from \"types\";",
				Options: []any{map[string]any{"ignoreTypeImports": true}},
			},
			{
				Code:    "",
				Options: []any{map[string]any{"max": 0}},
			},
			// Upstream throws with a negative max and no source; rslint safely emits nothing.
			{
				Code:    "export const value = 1;",
				Options: []any{map[string]any{"max": -1}},
			},
			{
				Code:    "import \"a\"; require(\"\\u0061\"); import(\"a\"); export * from \"a\";",
				Options: []any{map[string]any{"max": 1}},
			},
			{
				Code:    "import a from \"a\" with { type: \"json\" }; import(\"a\", { with: { type: \"json\" } });",
				Options: []any{map[string]any{"max": 1}},
			},
			{
				Code:    "import type Default from \"a\"; import type * as NS from \"b\"; import type { Named } from \"c\";",
				Options: []any{map[string]any{"max": 0, "ignoreTypeImports": true}},
			},
			// Module syntax and ESTree representation.
			{
				Code:    "const value = 1; export { value }; export default value;\nrequire(); require(\"a\", \"b\"); require(1); require(null); require(true); require(1n); require(/a/); require(`a`); require(`a${value}`); require(\"a\" + value); require(...[\"a\"]);\nimport(`b`); import(\"b\" + value); import(value);\nmodule.require(\"c\"); require.resolve(\"d\"); require[\"resolve\"](\"e\"); obj?.require(\"f\"); (obj?.require)(\"g\"); new require(\"h\"); define([\"i\"], () => {}); require([\"j\"], () => {});",
				Options: []any{map[string]any{"max": 0}},
			},
			{
				Code:    "import alias = require(\"a\"); type T = import(\"b\").T; const load = require as any; (require as any)(\"c\"); require(\"d\" as string); require!(\"e\"); import(\"f\" as string);",
				Options: []any{map[string]any{"max": 0}},
			},
			{
				Code:     "/** @import { T } from \"a\" */\n/** @type {import(\"b\").T} */\nlet value;",
				Options:  []any{map[string]any{"max": 0}},
				FileName: "jsdoc.js",
			},
			// Unicode escapes and empty paths use decoded string identity.
			{
				Code:    "import \"😀\"; require(\"\\uD83D\\uDE00\"); import \"\"; require(\"\");",
				Options: []any{map[string]any{"max": 2}},
			},
			// Upstream checks the callee spelling rather than tracking aliases.
			{
				Code:    "const load = require; load(\"a\"); const {require: other}=globalThis; other(\"b\");",
				Options: []any{map[string]any{"max": 0}},
			},
			// Ignoring a type import must not hide a later value import.
			{
				Code:    "import type A from \"a\"; import \"a\"; import type B from \"b\";",
				Options: []any{map[string]any{"max": 1, "ignoreTypeImports": true}},
			},
			// Do not move a disabled final report to an earlier import.
			// The Go RuleTester registers the rule under the name test.
			{
				Code:    "import \"a\";\nimport \"b\";\n// eslint-disable-next-line test\nimport \"c\";",
				Options: []any{map[string]any{"max": 1}},
			},
			// Line continuations decode to the same module path.
			{
				Code:    "import \"ab\"; require(\"a\\\nb\");",
				Options: []any{map[string]any{"max": 1}},
			},
		},
		[]rule_tester.InvalidTestCase{
			// Defaults, option boundaries, and module identity.
			{
				Code:   "import \"dep0\";\nimport \"dep1\";\nimport \"dep2\";\nimport \"dep3\";\nimport \"dep4\";\nimport \"dep5\";\nimport \"dep6\";\nimport \"dep7\";\nimport \"dep8\";\nimport \"dep9\";\nimport \"dep10\";",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Maximum number of dependencies (10) exceeded.", Line: 11, Column: 8, EndLine: 11, EndColumn: 15}},
			},
			{
				Code:    "import \"dep0\";\nimport \"dep1\";\nimport \"dep2\";\nimport \"dep3\";\nimport \"dep4\";\nimport \"dep5\";\nimport \"dep6\";\nimport \"dep7\";\nimport \"dep8\";\nimport \"dep9\";\nimport \"dep10\";",
				Options: []any{map[string]any{}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Maximum number of dependencies (10) exceeded.", Line: 11, Column: 8, EndLine: 11, EndColumn: 15}},
			},
			{
				Code:    "import \"dep0\";\nimport \"dep1\";\nimport \"dep2\";\nimport \"dep3\";\nimport \"dep4\";\nimport \"dep5\";\nimport \"dep6\";\nimport \"dep7\";\nimport \"dep8\";\nimport \"dep9\";\nimport \"dep10\";",
				Options: []any{map[string]any{"ignoreTypeImports": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Maximum number of dependencies (10) exceeded.", Line: 11, Column: 8, EndLine: 11, EndColumn: 15}},
			},
			{
				Code:    "import \"dep0\";\nimport \"dep1\";\nimport \"dep2\";\nimport \"dep3\";\nimport \"dep4\";\nimport \"dep5\";\nimport \"dep6\";\nimport \"dep7\";\nimport \"dep8\";\nimport \"dep9\";\nimport type T from \"types\";",
				Options: []any{map[string]any{"ignoreTypeImports": false}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Maximum number of dependencies (10) exceeded.", Line: 11, Column: 20, EndLine: 11, EndColumn: 27}},
			},
			{
				Code:    "require(\"a\");",
				Options: []any{map[string]any{"max": 0}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Maximum number of dependencies (0) exceeded.", Line: 1, Column: 9, EndLine: 1, EndColumn: 12}},
			},
			{
				Code:    "import \"a\"; import \"b\";",
				Options: []any{map[string]any{"max": 1.5}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Maximum number of dependencies (1.5) exceeded.", Line: 1, Column: 20, EndLine: 1, EndColumn: 23}},
			},
			{
				Code:    "import \"a\";",
				Options: []any{map[string]any{"max": -1}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Maximum number of dependencies (-1) exceeded.", Line: 1, Column: 8, EndLine: 1, EndColumn: 11}},
			},
			{
				Code:    "import \"a\";",
				Options: []any{map[string]any{"max": 1e-7}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Maximum number of dependencies (1e-7) exceeded.", Line: 1, Column: 8, EndLine: 1, EndColumn: 11}},
			},
			// Duplicates still update the final diagnostic location.
			{
				Code:    "import \"a\"; import \"b\"; import \"a\";",
				Options: []any{map[string]any{"max": 1}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Maximum number of dependencies (1) exceeded.", Line: 1, Column: 32, EndLine: 1, EndColumn: 35}},
			},
			// Ignored type imports still update the final diagnostic location.
			{
				Code:    "import \"a\"; import \"b\"; import type T from \"types\";",
				Options: []any{map[string]any{"max": 1, "ignoreTypeImports": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Maximum number of dependencies (1) exceeded.", Line: 1, Column: 44, EndLine: 1, EndColumn: 51}},
			},
			// Inline type specifiers and type re-exports are counted upstream.
			{
				Code:    "import { type T } from \"a\"; export type { U } from \"b\";",
				Options: []any{map[string]any{"max": 1, "ignoreTypeImports": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Maximum number of dependencies (1) exceeded.", Line: 1, Column: 52, EndLine: 1, EndColumn: 55}},
			},
			// Count source values without resolving module paths.
			{
				Code:    "import \"./a\"; import \"./a.js\";",
				Options: []any{map[string]any{"max": 1}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Maximum number of dependencies (1) exceeded.", Line: 1, Column: 22, EndLine: 1, EndColumn: 30}},
			},
			// Module syntax and ESTree representation.
			{
				Code:    "import \"a\"; export { value } from \"b\";",
				Options: []any{map[string]any{"max": 1}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Maximum number of dependencies (1) exceeded.", Line: 1, Column: 35, EndLine: 1, EndColumn: 38}},
			},
			{
				Code:    "import \"a\"; export * as ns from \"b\";",
				Options: []any{map[string]any{"max": 1}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Maximum number of dependencies (1) exceeded.", Line: 1, Column: 33, EndLine: 1, EndColumn: 36}},
			},
			{
				Code:    "import \"a\"; export * from \"b\";",
				Options: []any{map[string]any{"max": 1}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Maximum number of dependencies (1) exceeded.", Line: 1, Column: 27, EndLine: 1, EndColumn: 30}},
			},
			{
				Code:    "import \"a\"; export type * from \"b\";",
				Options: []any{map[string]any{"max": 1, "ignoreTypeImports": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Maximum number of dependencies (1) exceeded.", Line: 1, Column: 32, EndLine: 1, EndColumn: 35}},
			},
			{
				Code:    "import \"a\"; import(\"b\", { with: { type: \"json\" } });",
				Options: []any{map[string]any{"max": 1}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Maximum number of dependencies (1) exceeded.", Line: 1, Column: 20, EndLine: 1, EndColumn: 23}},
			},
			{
				Code:    "import \"a\"; (require)((\"b\"));",
				Options: []any{map[string]any{"max": 1}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Maximum number of dependencies (1) exceeded.", Line: 1, Column: 24, EndLine: 1, EndColumn: 27}},
			},
			{
				Code:    "import \"a\"; import((\"b\"));",
				Options: []any{map[string]any{"max": 1}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Maximum number of dependencies (1) exceeded.", Line: 1, Column: 21, EndLine: 1, EndColumn: 24}},
			},
			{
				Code:    "import \"a\"; require?.(\"b\");",
				Options: []any{map[string]any{"max": 1}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Maximum number of dependencies (1) exceeded.", Line: 1, Column: 23, EndLine: 1, EndColumn: 26}},
			},
			// Calls count across scopes, including a shadowed require.
			{
				Code:    "function load(require) { return require(\"a\"); }\nclass Loader { static { require(\"b\"); } }",
				Options: []any{map[string]any{"max": 1}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Maximum number of dependencies (1) exceeded.", Line: 2, Column: 33, EndLine: 2, EndColumn: 36}},
			},
			{
				Code:    "import(\"a\").then(() => import(\"b\"));",
				Options: []any{map[string]any{"max": 1}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Maximum number of dependencies (1) exceeded.", Line: 1, Column: 31, EndLine: 1, EndColumn: 34}},
			},
			{
				Code:     "const node = <Widget value={require(\"a\")} />;\nclass Loader { #require() {} load() { return this.#require(\"ignored\"); } }\nrequire(\"b\");",
				Options:  []any{map[string]any{"max": 1}},
				FileName: "example.tsx",
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Maximum number of dependencies (1) exceeded.", Line: 3, Column: 9, EndLine: 3, EndColumn: 12}},
			},
			{
				Code:    "import \"a\";\n/* 😀 */ require(\n  \"包📦\"\n);",
				Options: []any{map[string]any{"max": 1}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Maximum number of dependencies (1) exceeded.", Line: 3, Column: 3, EndLine: 3, EndColumn: 8}},
			},
			{
				Code:    "import \"a\"; /* 😀 */ require(\"包📦\");",
				Options: []any{map[string]any{"max": 1}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Maximum number of dependencies (1) exceeded.", Line: 1, Column: 30, EndLine: 1, EndColumn: 35}},
			},
			{
				Code:     "import \"a\";\n/** @type {any} */ (require)(/** @type {string} */ (\"b\"));",
				Options:  []any{map[string]any{"max": 1}},
				FileName: "casts.js",
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Maximum number of dependencies (1) exceeded.", Line: 2, Column: 53, EndLine: 2, EndColumn: 56}},
			},
			// Empty module paths count even though resolution cannot use them.
			{
				Code:    "import \"a\"; import \"\";",
				Options: []any{map[string]any{"max": 1}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Maximum number of dependencies (1) exceeded.", Line: 1, Column: 20, EndLine: 1, EndColumn: 22}},
			},
			// Escaped identifier spelling still names require.
			{
				Code:    "import \"a\"; \\u0072\\u0065\\u0071\\u0075\\u0069\\u0072\\u0065(\"b\");",
				Options: []any{map[string]any{"max": 1}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Maximum number of dependencies (1) exceeded.", Line: 1, Column: 56, EndLine: 1, EndColumn: 59}},
			},
			// Call type arguments count; a parenthesized instantiation does not.
			{
				Code:    "require<unknown>(\"a\"); (require<unknown>)(\"ignored\");",
				Options: []any{map[string]any{"max": 0}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Maximum number of dependencies (0) exceeded.", Line: 1, Column: 18, EndLine: 1, EndColumn: 21}},
			},
			// The inner optional require counts; the outer call does not.
			{
				Code:    "(require?.(\"a\"))(\"ignored\"); import \"b\";",
				Options: []any{map[string]any{"max": 1}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Maximum number of dependencies (1) exceeded.", Line: 1, Column: 37, EndLine: 1, EndColumn: 40}},
			},
			// Nested calls are visited after the enclosing import source.
			{
				Code:    "import(\"a\", {with: require(\"b\")});",
				Options: []any{map[string]any{"max": 1}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Maximum number of dependencies (1) exceeded.", Line: 1, Column: 28, EndLine: 1, EndColumn: 31}},
			},
			// Imports inside ambient modules count; the module name does not.
			{
				Code:    "declare module \"outer\" { import a from \"a\"; export {b} from \"b\"; }",
				Options: []any{map[string]any{"max": 1}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Maximum number of dependencies (1) exceeded.", Line: 1, Column: 61, EndLine: 1, EndColumn: 64}},
			},
			// An ignored type import is still the last source with a negative limit.
			{
				Code:    "import type A from \"a\";",
				Options: []any{map[string]any{"max": -1, "ignoreTypeImports": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Maximum number of dependencies (-1) exceeded.", Line: 1, Column: 20, EndLine: 1, EndColumn: 23}},
			},
			// A disabled line still contributes to the file dependency count.
			{
				Code:    "// eslint-disable-next-line test\nimport \"a\";\nimport \"b\";",
				Options: []any{map[string]any{"max": 1}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Maximum number of dependencies (1) exceeded.", Line: 3, Column: 8, EndLine: 3, EndColumn: 11}},
			},
			// Module paths remain case-sensitive.
			{
				Code:    "import \"a\"; require(\"A\");",
				Options: []any{map[string]any{"max": 1}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Maximum number of dependencies (1) exceeded.", Line: 1, Column: 21, EndLine: 1, EndColumn: 24}},
			},
			// Loader suffixes remain part of the module path.
			{
				Code:    "import \"a\"; require(\"a!loader\");",
				Options: []any{map[string]any{"max": 1}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Maximum number of dependencies (1) exceeded.", Line: 1, Column: 21, EndLine: 1, EndColumn: 31}},
			},
			// BOM, CRLF, and astral characters preserve diagnostic columns.
			{
				Code:    "\ufeff" + "import \"a\";\r\n/* 😀 */ require(\"b\");",
				Options: []any{map[string]any{"max": 1}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Maximum number of dependencies (1) exceeded.", Line: 2, Column: 18, EndLine: 2, EndColumn: 21}},
			},
		},
	)
}
