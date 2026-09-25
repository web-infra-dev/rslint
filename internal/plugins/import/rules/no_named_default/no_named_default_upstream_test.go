package no_named_default_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/no_named_default"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/tests/src/rules/no-named-default.js
func TestNoNamedDefaultUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_named_default.NoNamedDefaultRule,
		[]rule_tester.ValidTestCase{
			{Code: `import bar from "./bar";`},
			{Code: `import bar, { foo } from "./bar";`},
			// Upstream's Flow type case also parses as TypeScript.
			{Code: `import { type default as Foo } from "./bar";`},
			// Flow's typeof import specifiers are not supported by tsgo.
			{Code: `import { typeof default as Foo } from "./bar";`, Skip: true},
			// SYNTAX_CASES from tests/src/utils.js at the same tag.
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
			// https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/docs/rules/no-named-default.md
			{Code: "// foo.js\nexport default 'foo';\nexport const bar = 'baz';"},
			{Code: `import foo from './foo.js';`},
			{Code: `import foo, { bar } from './foo.js';`},
		},
		[]rule_tester.InvalidTestCase{
			// Upstream uses literal messages without message IDs, fixes or suggestions.
			{
				Code: `import { default as bar } from "./bar";`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: "Use default import syntax to import 'bar'.", Line: 1, Column: 21, EndLine: 1, EndColumn: 24},
				},
			},
			{
				Code: `import { foo, default as bar } from "./bar";`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: "Use default import syntax to import 'bar'.", Line: 1, Column: 26, EndLine: 1, EndColumn: 29},
				},
			},
			{
				Code: `import { "default" as bar } from "./bar";`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: "Use default import syntax to import 'bar'.", Line: 1, Column: 23, EndLine: 1, EndColumn: 26},
				},
			},
			// Documentation examples; use the source's message, not the stale doc comment.
			{
				Code: `import { default as foo } from './foo.js';`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: "Use default import syntax to import 'foo'.", Line: 1, Column: 21, EndLine: 1, EndColumn: 24},
				},
			},
			{
				Code: `import { default as foo, bar } from './foo.js';`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: "Use default import syntax to import 'foo'.", Line: 1, Column: 21, EndLine: 1, EndColumn: 24},
				},
			},
		},
	)
}
