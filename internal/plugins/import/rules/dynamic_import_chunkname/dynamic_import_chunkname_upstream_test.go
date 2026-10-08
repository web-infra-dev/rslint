package dynamic_import_chunkname_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/dynamic_import_chunkname"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Every case from https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/tests/src/rules/dynamic-import-chunkname.js
// in upstream order. The JavaScript section runs upstream with Babel and the
// TypeScript section with typescript-eslint; both are parsed by tsgo here, so
// the two sections can only differ in the cases they list. Upstream asserts
// messages, not positions, so every range below was recorded from the pinned
// upstream rule running under ESLint 8.57.1 with the matching parser.
//
// The `Output` of each suggestion is the source after applying that suggestion;
// the suggestions are listed in upstream order (Remove webpackChunkName, then
// Remove webpackMode).
func TestDynamicImportChunknameUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &dynamic_import_chunkname.DynamicImportChunknameRule,
		[]rule_tester.ValidTestCase{
			// ---- upstream valid: JavaScript ----
			{
				Code: `dynamicImport(
        /* webpackChunkName: "someModule" */
        'test'
      )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
			},
			{
				Code: `dynamicImport(
        /* webpackChunkName: "Some_Other_Module" */
        "test"
      )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
			},
			{
				Code: `dynamicImport(
        /* webpackChunkName: "SomeModule123" */
        "test"
      )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
			},
			{
				Code: `dynamicImport(
        /* webpackChunkName: "someModule" */
        'someModule'
      )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}, "webpackChunknameFormat": "[a-zA-Z-_/.]+"}},
			},
			{
				Code: `dynamicImport(
        /* webpackChunkName: "[request]" */
        'someModule'
      )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
			},
			{
				Code: `dynamicImport(
        /* webpackChunkName: "my-chunk-[request]-custom" */
        'someModule'
      )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
			},
			{
				Code: `dynamicImport(
        /* webpackChunkName: '[index]' */
        'someModule'
      )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
			},
			{
				Code: `dynamicImport(
        /* webpackChunkName: 'my-chunk.[index].with-index' */
        'someModule'
      )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
			},
			{
				Code:    `import('test')`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}, "allowEmpty": true}},
			},
			{
				Code: `import(
        /* webpackMode: "lazy" */
        'test'
      )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}, "allowEmpty": true}},
			},
			{
				Code: `import(
        /* webpackChunkName: "someModule" */
        'test'
      )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
			},
			{
				Code: `import(
        /* webpackChunkName: "Some_Other_Module" */
        "test"
      )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
			},
			{
				Code: `import(
        /* webpackChunkName: "SomeModule123" */
        "test"
      )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
			},
			{
				Code: `import(
        /* webpackChunkName: "someModule", webpackPrefetch: true */
        'test'
      )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
			},
			{
				Code: `import(
        /* webpackChunkName: "someModule", webpackPrefetch: true, */
        'test'
      )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
			},
			{
				Code: `import(
        /* webpackPrefetch: true, webpackChunkName: "someModule" */
        'test'
      )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
			},
			{
				Code: `import(
        /* webpackPrefetch: true, webpackChunkName: "someModule", */
        'test'
      )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
			},
			{
				Code: `import(
        /* webpackPrefetch: true */
        /* webpackChunkName: "someModule" */
        'test'
      )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
			},
			{
				Code: `import(
        /* webpackChunkName: "someModule" */
        /* webpackPrefetch: true */
        'test'
      )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
			},
			{
				Code: `import(
        /* webpackChunkName: "someModule" */
        /* webpackPrefetch: 12 */
        'test'
      )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
			},
			{
				Code: `import(
        /* webpackChunkName: "someModule" */
        /* webpackPrefetch: -30 */
        'test'
      )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
			},
			{
				Code: `import(
        /* webpackChunkName: 'someModule' */
        'someModule'
      )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
			},
			{
				Code: `import(
        /* webpackChunkName: "someModule" */
        'someModule'
      )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}, "webpackChunknameFormat": "[a-zA-Z-_/.]+"}},
			},
			{
				Code: `import(
        /* webpackChunkName: "[request]" */
        'someModule'
      )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
			},
			{
				Code: `import(
        /* webpackChunkName: "my-chunk-[request]-custom" */
        'someModule'
      )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
			},
			{
				Code: `import(
        /* webpackChunkName: '[index]' */
        'someModule'
      )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
			},
			{
				Code: `import(
        /* webpackChunkName: 'my-chunk.[index].with-index' */
        'someModule'
      )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
			},
			{
				Code: `import(
        /* webpackChunkName: "someModule" */
        /* webpackInclude: /\.json$/ */
        'someModule'
      )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
			},
			{
				Code: `import(
        /* webpackChunkName: "someModule", webpackInclude: /\.json$/ */
        'someModule'
      )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
			},
			{
				Code: `import(
        /* webpackChunkName: "someModule" */
        /* webpackExclude: /\.json$/ */
        'someModule'
      )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
			},
			{
				Code: `import(
        /* webpackChunkName: "someModule", webpackExclude: /\.json$/ */
        'someModule'
      )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
			},
			{
				Code: `import(
        /* webpackChunkName: "someModule" */
        /* webpackPreload: true */
        'someModule'
      )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
			},
			{
				Code: `import(
        /* webpackChunkName: "someModule" */
        /* webpackPreload: 0 */
        'someModule'
      )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
			},
			{
				Code: `import(
        /* webpackChunkName: "someModule" */
        /* webpackPreload: -2 */
        'someModule'
      )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
			},
			{
				Code: `import(
        /* webpackChunkName: "someModule", webpackPreload: false */
        'someModule'
      )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
			},
			{
				Code: `import(
        /* webpackChunkName: "someModule" */
        /* webpackIgnore: false */
        'someModule'
      )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
			},
			{
				Code: `import(
        /* webpackChunkName: "someModule", webpackIgnore: true */
        'someModule'
      )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
			},
			{
				Code: `import(
        /* webpackChunkName: "someModule" */
        /* webpackMode: "lazy" */
        'someModule'
      )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
			},
			{
				Code: `import(
        /* webpackChunkName: 'someModule', webpackMode: 'lazy' */
        'someModule'
      )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
			},
			{
				Code: `import(
        /* webpackChunkName: "someModule" */
        /* webpackMode: "lazy-once" */
        'someModule'
      )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
			},
			{
				Code: `import(
        /* webpackMode: "eager" */
        'someModule'
      )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
			},
			{
				Code: `import(
        /* webpackChunkName: "someModule" */
        /* webpackMode: "weak" */
        'someModule'
      )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
			},
			{
				Code: `import(
        /* webpackChunkName: "someModule" */
        /* webpackExports: "default" */
        'someModule'
      )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
			},
			{
				Code: `import(
        /* webpackChunkName: "someModule", webpackExports: "named" */
        'someModule'
      )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
			},
			{
				Code: `import(
        /* webpackChunkName: "someModule" */
        /* webpackExports: ["default", "named"] */
        'someModule'
      )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
			},
			{
				Code: `import(
        /* webpackChunkName: 'someModule', webpackExports: ['default', 'named'] */
        'someModule'
      )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
			},
			{
				Code: `import(
        /* webpackChunkName: "someModule" */
        /* webpackInclude: /\.json$/ */
        /* webpackExclude: /\.json$/ */
        /* webpackPrefetch: true */
        /* webpackPreload: true */
        /* webpackIgnore: false */
        /* webpackMode: "lazy" */
        /* webpackExports: ["default", "named"] */
        'someModule'
      )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
			},
			{
				Code: `for (let { foo, bar } of baz) {}`,
			},
			{
				Code: `for (let [ foo, bar ] of baz) {}`,
			},
			{
				Code: `const { x, y } = bar`,
			},
			{
				Code: `const { x, y, ...z } = bar`,
			},
			{
				Code: `let x; export { x }`,
			},
			{
				Code: `let x; export { x as y }`,
			},
			{
				Code: `export const x = null`,
			},
			{
				Code: `export var x = null`,
			},
			{
				Code: `export let x = null`,
			},
			{
				Code: `export default x`,
			},
			{
				Code: `export default class x {}`,
			},
			{
				Code: `import json from "./data.json"`,
			},
			{
				Code: `import foo from "./foobar.json";`,
			},
			{
				Code: `import foo from "./foobar";`,
			},
			{
				Code: `import { foo } from "./issue-370-commonjs-namespace/bar"`,
			},
			{
				Code: `export * from "./issue-370-commonjs-namespace/bar"`,
			},
			{
				Code: `import * as a from "./commonjs-namespace/a"; a.b`,
			},
			{
				Code: `import { foo } from "./ignore.invalid.extension"`,
			},
			// ---- upstream valid: TypeScript ----
			{
				Code:    `import('test')`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}, "allowEmpty": true}},
			},
			{
				Code: `import(
            /* webpackMode: "lazy" */
            'test'
          )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}, "allowEmpty": true}},
			},
			{
				Code: `import(
            /* webpackChunkName: "someModule" */
            'test'
          )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
			},
			{
				Code: `import(
            /* webpackChunkName: "Some_Other_Module" */
            "test"
          )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
			},
			{
				Code: `import(
            /* webpackChunkName: "SomeModule123" */
            "test"
          )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
			},
			{
				Code: `import(
            /* webpackChunkName: "someModule", webpackPrefetch: true */
            'test'
          )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
			},
			{
				Code: `import(
            /* webpackChunkName: "someModule", webpackPrefetch: true, */
            'test'
          )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
			},
			{
				Code: `import(
            /* webpackPrefetch: true, webpackChunkName: "someModule" */
            'test'
          )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
			},
			{
				Code: `import(
            /* webpackPrefetch: true, webpackChunkName: "someModule", */
            'test'
          )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
			},
			{
				Code: `import(
            /* webpackPrefetch: true */
            /* webpackChunkName: "someModule" */
            'test'
          )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
			},
			{
				Code: `import(
            /* webpackChunkName: "someModule" */
            /* webpackPrefetch: true */
            'test'
          )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
			},
			{
				Code: `import(
            /* webpackChunkName: "someModule" */
            /* webpackPrefetch: 11 */
            'test'
          )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
			},
			{
				Code: `import(
            /* webpackChunkName: "someModule" */
            /* webpackPrefetch: -11 */
            'test'
          )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
			},
			{
				Code: `import(
            /* webpackChunkName: "someModule" */
            'someModule'
          )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}, "webpackChunknameFormat": "[a-zA-Z-_/.]+"}},
			},
			{
				Code: `import(
            /* webpackChunkName: 'someModule' */
            'test'
          )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
			},
			{
				Code: `import(
            /* webpackChunkName: "[request]" */
            'someModule'
          )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
			},
			{
				Code: `import(
            /* webpackChunkName: "my-chunk-[request]-custom" */
            'someModule'
          )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
			},
			{
				Code: `import(
            /* webpackChunkName: '[index]' */
            'someModule'
          )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
			},
			{
				Code: `import(
            /* webpackChunkName: 'my-chunk.[index].with-index' */
            'someModule'
          )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
			},
			{
				Code: `import(
            /* webpackChunkName: "someModule" */
            /* webpackInclude: /\.json$/ */
            'someModule'
          )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
			},
			{
				Code: `import(
            /* webpackChunkName: "someModule", webpackInclude: /\.json$/ */
            'someModule'
          )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
			},
			{
				Code: `import(
            /* webpackChunkName: "someModule" */
            /* webpackExclude: /\.json$/ */
            'someModule'
          )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
			},
			{
				Code: `import(
            /* webpackChunkName: "someModule", webpackExclude: /\.json$/ */
            'someModule'
          )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
			},
			{
				Code: `import(
            /* webpackChunkName: "someModule" */
            /* webpackPreload: true */
            'someModule'
          )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
			},
			{
				Code: `import(
            /* webpackChunkName: "someModule", webpackPreload: false */
            'someModule'
          )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
			},
			{
				Code: `import(
            /* webpackChunkName: "someModule" */
            /* webpackIgnore: false */
            'someModule'
          )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
			},
			{
				Code: `import(
            /* webpackChunkName: "someModule", webpackIgnore: true */
            'someModule'
          )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
			},
			{
				Code: `import(
            /* webpackChunkName: 'someModule', webpackMode: 'lazy' */
            'someModule'
          )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
			},
			{
				Code: `import(
            /* webpackChunkName: "someModule" */
            /* webpackMode: "lazy-once" */
            'someModule'
          )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
			},
			{
				Code: `import(
            /* webpackChunkName: "someModule" */
            /* webpackMode: "lazy" */
            'someModule'
          )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
			},
			{
				Code: `import(
            /* webpackChunkName: "someModule" */
            /* webpackMode: "weak" */
            'someModule'
          )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
			},
			{
				Code: `import(
            /* webpackChunkName: "someModule" */
            /* webpackExports: "default" */
            'someModule'
          )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
			},
			{
				Code: `import(
            /* webpackChunkName: "someModule", webpackExports: "named" */
            'someModule'
          )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
			},
			{
				Code: `import(
            /* webpackChunkName: "someModule" */
            /* webpackExports: ["default", "named"] */
            'someModule'
          )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
			},
			{
				Code: `import(
            /* webpackChunkName: 'someModule', webpackExports: ['default', 'named'] */
            'someModule'
          )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
			},
			{
				Code: `import(
            /* webpackChunkName: "someModule" */
            /* webpackInclude: /\.json$/ */
            /* webpackExclude: /\.json$/ */
            /* webpackPrefetch: true */
            /* webpackPreload: true */
            /* webpackIgnore: false */
            /* webpackMode: "lazy" */
            /* webpackExports: ["default", "named"] */
            'someModule'
          )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
			},
			{
				Code: `import(
            /* webpackMode: "eager" */
            'someModule'
          )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
			},
		},
		[]rule_tester.InvalidTestCase{
			// ---- upstream invalid: JavaScript ----
			{
				Code: `import(
        // webpackChunkName: "someModule"
        'someModule'
      )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: `dynamic imports require a /* foo */ style comment, not a // foo comment`,
						Line:    1, Column: 1, EndLine: 4, EndColumn: 8,
					},
				},
			},
			{
				Code:    `import('test')`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: `dynamic imports require a leading comment with the webpack chunkname`,
						Line:    1, Column: 1, EndLine: 1, EndColumn: 15,
					},
				},
			},
			{
				Code: `import(
        /* webpackChunkName: someModule */
        'someModule'
      )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: `dynamic imports require a "webpack" comment with valid syntax`,
						Line:    1, Column: 1, EndLine: 4, EndColumn: 8,
					},
				},
			},
			{
				Code: `import(
        /* webpackChunkName: "someModule' */
        'someModule'
      )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: `dynamic imports require a "webpack" comment with valid syntax`,
						Line:    1, Column: 1, EndLine: 4, EndColumn: 8,
					},
				},
			},
			{
				Code: `import(
        /* webpackChunkName: 'someModule" */
        'someModule'
      )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: `dynamic imports require a "webpack" comment with valid syntax`,
						Line:    1, Column: 1, EndLine: 4, EndColumn: 8,
					},
				},
			},
			{
				Code: `import(
        /* webpackChunkName "someModule" */
        'someModule'
      )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: `dynamic imports require a "webpack" comment with valid syntax`,
						Line:    1, Column: 1, EndLine: 4, EndColumn: 8,
					},
				},
			},
			{
				Code: `import(
        /* webpackChunkName:"someModule" */
        'someModule'
      )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: `dynamic imports require a "webpack" comment with valid syntax`,
						Line:    1, Column: 1, EndLine: 4, EndColumn: 8,
					},
				},
			},
			{
				Code: `import(
        /* webpackChunkName: true */
        'someModule'
      )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: `dynamic imports require a leading comment in the form /*webpackChunkName: ["']([0-9a-zA-Z-_/.]|\[(request|index)\])+["'],? */`,
						Line:    1, Column: 1, EndLine: 4, EndColumn: 8,
					},
				},
			},
			{
				Code: `import(
        /* webpackChunkName: "my-module-[id]" */
        'someModule'
      )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: `dynamic imports require a leading comment in the form /*webpackChunkName: ["']([0-9a-zA-Z-_/.]|\[(request|index)\])+["'],? */`,
						Line:    1, Column: 1, EndLine: 4, EndColumn: 8,
					},
				},
			},
			{
				Code: `import(
        /* webpackChunkName: ["request"] */
        'someModule'
      )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: `dynamic imports require a leading comment in the form /*webpackChunkName: ["']([0-9a-zA-Z-_/.]|\[(request|index)\])+["'],? */`,
						Line:    1, Column: 1, EndLine: 4, EndColumn: 8,
					},
				},
			},
			{
				Code: `import(
        /*webpackChunkName: "someModule"*/
        'someModule'
      )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: `dynamic imports require a block comment padded with spaces - /* foo */`,
						Line:    1, Column: 1, EndLine: 4, EndColumn: 8,
					},
				},
			},
			{
				Code: `import(
        /* webpackChunkName  :  "someModule" */
        'someModule'
      )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: `dynamic imports require a "webpack" comment with valid syntax`,
						Line:    1, Column: 1, EndLine: 4, EndColumn: 8,
					},
				},
			},
			{
				Code: `import(
        /* webpackChunkName: "someModule" ; */
        'someModule'
      )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: `dynamic imports require a "webpack" comment with valid syntax`,
						Line:    1, Column: 1, EndLine: 4, EndColumn: 8,
					},
				},
			},
			{
				Code: `import(
        /* totally not webpackChunkName: "someModule" */
        'someModule'
      )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: `dynamic imports require a "webpack" comment with valid syntax`,
						Line:    1, Column: 1, EndLine: 4, EndColumn: 8,
					},
				},
			},
			{
				Code: `import(
        /* webpackPrefetch: true */
        /* webpackChunk: "someModule" */
        'someModule'
      )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: `dynamic imports require a "webpack" comment with valid syntax`,
						Line:    1, Column: 1, EndLine: 5, EndColumn: 8,
					},
				},
			},
			{
				Code: `import(
        /* webpackPrefetch: true, webpackChunk: "someModule" */
        'someModule'
      )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: `dynamic imports require a "webpack" comment with valid syntax`,
						Line:    1, Column: 1, EndLine: 4, EndColumn: 8,
					},
				},
			},
			{
				Code: `import(
        /* webpackChunkName: "someModule123" */
        'someModule'
      )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}, "webpackChunknameFormat": "[a-zA-Z-_/.]+"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: `dynamic imports require a leading comment in the form /*webpackChunkName: ["'][a-zA-Z-_/.]+["'],? */`,
						Line:    1, Column: 1, EndLine: 4, EndColumn: 8,
					},
				},
			},
			{
				Code: `import(
        /* webpackPrefetch: "module", webpackChunkName: "someModule" */
        'someModule'
      )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: `dynamic imports require a "webpack" comment with valid syntax`,
						Line:    1, Column: 1, EndLine: 4, EndColumn: 8,
					},
				},
			},
			{
				Code: `import(
        /* webpackPreload: "module", webpackChunkName: "someModule" */
        'someModule'
      )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: `dynamic imports require a "webpack" comment with valid syntax`,
						Line:    1, Column: 1, EndLine: 4, EndColumn: 8,
					},
				},
			},
			{
				Code: `import(
        /* webpackIgnore: "no", webpackChunkName: "someModule" */
        'someModule'
      )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: `dynamic imports require a "webpack" comment with valid syntax`,
						Line:    1, Column: 1, EndLine: 4, EndColumn: 8,
					},
				},
			},
			{
				Code: `import(
        /* webpackInclude: "someModule", webpackChunkName: "someModule" */
        'someModule'
      )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: `dynamic imports require a "webpack" comment with valid syntax`,
						Line:    1, Column: 1, EndLine: 4, EndColumn: 8,
					},
				},
			},
			{
				Code: `import(
        /* webpackInclude: true, webpackChunkName: "someModule" */
        'someModule'
      )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: `dynamic imports require a "webpack" comment with valid syntax`,
						Line:    1, Column: 1, EndLine: 4, EndColumn: 8,
					},
				},
			},
			{
				Code: `import(
        /* webpackExclude: "someModule", webpackChunkName: "someModule" */
        'someModule'
      )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: `dynamic imports require a "webpack" comment with valid syntax`,
						Line:    1, Column: 1, EndLine: 4, EndColumn: 8,
					},
				},
			},
			{
				Code: `import(
        /* webpackExclude: true, webpackChunkName: "someModule" */
        'someModule'
      )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: `dynamic imports require a "webpack" comment with valid syntax`,
						Line:    1, Column: 1, EndLine: 4, EndColumn: 8,
					},
				},
			},
			{
				Code: `import(
        /* webpackMode: "fast", webpackChunkName: "someModule" */
        'someModule'
      )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: `dynamic imports require a "webpack" comment with valid syntax`,
						Line:    1, Column: 1, EndLine: 4, EndColumn: 8,
					},
				},
			},
			{
				Code: `import(
        /* webpackMode: true, webpackChunkName: "someModule" */
        'someModule'
      )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: `dynamic imports require a "webpack" comment with valid syntax`,
						Line:    1, Column: 1, EndLine: 4, EndColumn: 8,
					},
				},
			},
			{
				Code: `import(
        /* webpackExports: true, webpackChunkName: "someModule" */
        'someModule'
      )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: `dynamic imports require a "webpack" comment with valid syntax`,
						Line:    1, Column: 1, EndLine: 4, EndColumn: 8,
					},
				},
			},
			{
				Code: `import(
        /* webpackExports: /default/, webpackChunkName: "someModule" */
        'someModule'
      )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: `dynamic imports require a "webpack" comment with valid syntax`,
						Line:    1, Column: 1, EndLine: 4, EndColumn: 8,
					},
				},
			},
			{
				Code: `dynamicImport(
        /* webpackChunkName "someModule" */
        'someModule'
      )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport", "definitelyNotStaticImport"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: `dynamic imports require a "webpack" comment with valid syntax`,
						Line:    1, Column: 1, EndLine: 4, EndColumn: 8,
					},
				},
			},
			{
				Code: `definitelyNotStaticImport(
        /* webpackChunkName "someModule" */
        'someModule'
      )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport", "definitelyNotStaticImport"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: `dynamic imports require a "webpack" comment with valid syntax`,
						Line:    1, Column: 1, EndLine: 4, EndColumn: 8,
					},
				},
			},
			{
				Code: `dynamicImport(
        // webpackChunkName: "someModule"
        'someModule'
      )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: `dynamic imports require a /* foo */ style comment, not a // foo comment`,
						Line:    1, Column: 1, EndLine: 4, EndColumn: 8,
					},
				},
			},
			{
				Code:    `dynamicImport('test')`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: `dynamic imports require a leading comment with the webpack chunkname`,
						Line:    1, Column: 1, EndLine: 1, EndColumn: 22,
					},
				},
			},
			{
				Code: `dynamicImport(
        /* webpackChunkName: someModule */
        'someModule'
      )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: `dynamic imports require a "webpack" comment with valid syntax`,
						Line:    1, Column: 1, EndLine: 4, EndColumn: 8,
					},
				},
			},
			{
				Code: `dynamicImport(
        /* webpackChunkName "someModule" */
        'someModule'
      )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: `dynamic imports require a "webpack" comment with valid syntax`,
						Line:    1, Column: 1, EndLine: 4, EndColumn: 8,
					},
				},
			},
			{
				Code: `dynamicImport(
        /* webpackChunkName:"someModule" */
        'someModule'
      )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: `dynamic imports require a "webpack" comment with valid syntax`,
						Line:    1, Column: 1, EndLine: 4, EndColumn: 8,
					},
				},
			},
			{
				Code: `dynamicImport(
        /* webpackChunkName: "someModule123" */
        'someModule'
      )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}, "webpackChunknameFormat": "[a-zA-Z-_/.]+"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: `dynamic imports require a leading comment in the form /*webpackChunkName: ["'][a-zA-Z-_/.]+["'],? */`,
						Line:    1, Column: 1, EndLine: 4, EndColumn: 8,
					},
				},
			},
			{
				Code: `import(
        /* webpackChunkName: "someModule" */
        /* webpackMode: "eager" */
        'someModule'
      )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: `dynamic imports using eager mode do not need a webpackChunkName`,
						Line:    1, Column: 1, EndLine: 5, EndColumn: 8,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							// Remove webpackChunkName
							{Output: `import(
        
        /* webpackMode: "eager" */
        'someModule'
      )`},
							// Remove webpackMode
							{Output: `import(
        /* webpackChunkName: "someModule" */
        
        'someModule'
      )`},
						},
					},
				},
			},
			// ---- upstream invalid: TypeScript ----
			{
				Code: `import(
            // webpackChunkName: "someModule"
            'someModule'
          )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: `dynamic imports require a /* foo */ style comment, not a // foo comment`,
						Line:    1, Column: 1, EndLine: 4, EndColumn: 12,
					},
				},
			},
			{
				Code:    `import('test')`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: `dynamic imports require a leading comment with the webpack chunkname`,
						Line:    1, Column: 1, EndLine: 1, EndColumn: 15,
					},
				},
			},
			{
				Code: `import(
            /* webpackChunkName: someModule */
            'someModule'
          )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: `dynamic imports require a "webpack" comment with valid syntax`,
						Line:    1, Column: 1, EndLine: 4, EndColumn: 12,
					},
				},
			},
			{
				Code: `import(
            /* webpackChunkName "someModule' */
            'someModule'
          )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: `dynamic imports require a "webpack" comment with valid syntax`,
						Line:    1, Column: 1, EndLine: 4, EndColumn: 12,
					},
				},
			},
			{
				Code: `import(
            /* webpackChunkName 'someModule" */
            'someModule'
          )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: `dynamic imports require a "webpack" comment with valid syntax`,
						Line:    1, Column: 1, EndLine: 4, EndColumn: 12,
					},
				},
			},
			{
				Code: `import(
            /* webpackChunkName "someModule" */
            'someModule'
          )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: `dynamic imports require a "webpack" comment with valid syntax`,
						Line:    1, Column: 1, EndLine: 4, EndColumn: 12,
					},
				},
			},
			{
				Code: `import(
            /* webpackChunkName:"someModule" */
            'someModule'
          )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: `dynamic imports require a "webpack" comment with valid syntax`,
						Line:    1, Column: 1, EndLine: 4, EndColumn: 12,
					},
				},
			},
			{
				Code: `import(
            /*webpackChunkName: "someModule"*/
            'someModule'
          )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: `dynamic imports require a block comment padded with spaces - /* foo */`,
						Line:    1, Column: 1, EndLine: 4, EndColumn: 12,
					},
				},
			},
			{
				Code: `import(
            /* webpackChunkName  :  "someModule" */
            'someModule'
          )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: `dynamic imports require a "webpack" comment with valid syntax`,
						Line:    1, Column: 1, EndLine: 4, EndColumn: 12,
					},
				},
			},
			{
				Code: `import(
            /* webpackChunkName: "someModule" ; */
            'someModule'
          )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: `dynamic imports require a "webpack" comment with valid syntax`,
						Line:    1, Column: 1, EndLine: 4, EndColumn: 12,
					},
				},
			},
			{
				Code: `import(
            /* totally not webpackChunkName: "someModule" */
            'someModule'
          )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: `dynamic imports require a "webpack" comment with valid syntax`,
						Line:    1, Column: 1, EndLine: 4, EndColumn: 12,
					},
				},
			},
			{
				Code: `import(
            /* webpackPrefetch: true */
            /* webpackChunk: "someModule" */
            'someModule'
          )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: `dynamic imports require a "webpack" comment with valid syntax`,
						Line:    1, Column: 1, EndLine: 5, EndColumn: 12,
					},
				},
			},
			{
				Code: `import(
            /* webpackPrefetch: true, webpackChunk: "someModule" */
            'someModule'
          )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: `dynamic imports require a "webpack" comment with valid syntax`,
						Line:    1, Column: 1, EndLine: 4, EndColumn: 12,
					},
				},
			},
			{
				Code: `import(
            /* webpackChunkName: true */
            'someModule'
          )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: `dynamic imports require a leading comment in the form /*webpackChunkName: ["']([0-9a-zA-Z-_/.]|\[(request|index)\])+["'],? */`,
						Line:    1, Column: 1, EndLine: 4, EndColumn: 12,
					},
				},
			},
			{
				Code: `import(
            /* webpackChunkName: "my-module-[id]" */
            'someModule'
          )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: `dynamic imports require a leading comment in the form /*webpackChunkName: ["']([0-9a-zA-Z-_/.]|\[(request|index)\])+["'],? */`,
						Line:    1, Column: 1, EndLine: 4, EndColumn: 12,
					},
				},
			},
			{
				Code: `import(
            /* webpackChunkName: ["request"] */
            'someModule'
          )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: `dynamic imports require a leading comment in the form /*webpackChunkName: ["']([0-9a-zA-Z-_/.]|\[(request|index)\])+["'],? */`,
						Line:    1, Column: 1, EndLine: 4, EndColumn: 12,
					},
				},
			},
			{
				Code: `import(
            /* webpackChunkName: "someModule123" */
            'someModule'
          )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}, "webpackChunknameFormat": "[a-zA-Z-_/.]+"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: `dynamic imports require a leading comment in the form /*webpackChunkName: ["'][a-zA-Z-_/.]+["'],? */`,
						Line:    1, Column: 1, EndLine: 4, EndColumn: 12,
					},
				},
			},
			{
				Code: `import(
            /* webpackPrefetch: "module", webpackChunkName: "someModule" */
            'someModule'
          )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: `dynamic imports require a "webpack" comment with valid syntax`,
						Line:    1, Column: 1, EndLine: 4, EndColumn: 12,
					},
				},
			},
			{
				Code: `import(
            /* webpackPreload: "module", webpackChunkName: "someModule" */
            'someModule'
          )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: `dynamic imports require a "webpack" comment with valid syntax`,
						Line:    1, Column: 1, EndLine: 4, EndColumn: 12,
					},
				},
			},
			{
				Code: `import(
            /* webpackIgnore: "no", webpackChunkName: "someModule" */
            'someModule'
          )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: `dynamic imports require a "webpack" comment with valid syntax`,
						Line:    1, Column: 1, EndLine: 4, EndColumn: 12,
					},
				},
			},
			{
				Code: `import(
            /* webpackInclude: "someModule", webpackChunkName: "someModule" */
            'someModule'
          )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: `dynamic imports require a "webpack" comment with valid syntax`,
						Line:    1, Column: 1, EndLine: 4, EndColumn: 12,
					},
				},
			},
			{
				Code: `import(
            /* webpackInclude: true, webpackChunkName: "someModule" */
            'someModule'
          )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: `dynamic imports require a "webpack" comment with valid syntax`,
						Line:    1, Column: 1, EndLine: 4, EndColumn: 12,
					},
				},
			},
			{
				Code: `import(
            /* webpackExclude: "someModule", webpackChunkName: "someModule" */
            'someModule'
          )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: `dynamic imports require a "webpack" comment with valid syntax`,
						Line:    1, Column: 1, EndLine: 4, EndColumn: 12,
					},
				},
			},
			{
				Code: `import(
            /* webpackExclude: true, webpackChunkName: "someModule" */
            'someModule'
          )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: `dynamic imports require a "webpack" comment with valid syntax`,
						Line:    1, Column: 1, EndLine: 4, EndColumn: 12,
					},
				},
			},
			{
				Code: `import(
            /* webpackMode: "fast", webpackChunkName: "someModule" */
            'someModule'
          )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: `dynamic imports require a "webpack" comment with valid syntax`,
						Line:    1, Column: 1, EndLine: 4, EndColumn: 12,
					},
				},
			},
			{
				Code: `import(
            /* webpackMode: true, webpackChunkName: "someModule" */
            'someModule'
          )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: `dynamic imports require a "webpack" comment with valid syntax`,
						Line:    1, Column: 1, EndLine: 4, EndColumn: 12,
					},
				},
			},
			{
				Code: `import(
            /* webpackExports: true, webpackChunkName: "someModule" */
            'someModule'
          )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: `dynamic imports require a "webpack" comment with valid syntax`,
						Line:    1, Column: 1, EndLine: 4, EndColumn: 12,
					},
				},
			},
			{
				Code: `import(
            /* webpackExports: /default/, webpackChunkName: "someModule" */
            'someModule'
          )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: `dynamic imports require a "webpack" comment with valid syntax`,
						Line:    1, Column: 1, EndLine: 4, EndColumn: 12,
					},
				},
			},
			{
				Code: `import(
            /* webpackChunkName: "someModule", webpackMode: "eager" */
            'someModule'
          )`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: `dynamic imports using eager mode do not need a webpackChunkName`,
						Line:    1, Column: 1, EndLine: 4, EndColumn: 12,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							// Remove webpackChunkName
							{Output: `import(
            /* webpackMode: "eager" */
            'someModule'
          )`},
							// Remove webpackMode
							{Output: `import(
            /* webpackChunkName: "someModule" */
            'someModule'
          )`},
						},
					},
				},
			},
			{
				Code: `
            import(
              /* webpackMode: "eager", webpackChunkName: "someModule" */
              'someModule'
            )
          `,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: `dynamic imports using eager mode do not need a webpackChunkName`,
						Line:    2, Column: 13, EndLine: 5, EndColumn: 14,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							// Remove webpackChunkName
							{Output: `
            import(
              /* webpackMode: "eager" */
              'someModule'
            )
          `},
							// Remove webpackMode
							{Output: `
            import(
              /* webpackChunkName: "someModule" */
              'someModule'
            )
          `},
						},
					},
				},
			},
			{
				Code: `
            import(
              /* webpackMode: "eager", webpackPrefetch: true, webpackChunkName: "someModule" */
              'someModule'
            )
          `,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: `dynamic imports using eager mode do not need a webpackChunkName`,
						Line:    2, Column: 13, EndLine: 5, EndColumn: 14,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							// Remove webpackChunkName
							{Output: `
            import(
              /* webpackMode: "eager", webpackPrefetch: true */
              'someModule'
            )
          `},
							// Remove webpackMode
							{Output: `
            import(
              /* webpackPrefetch: true, webpackChunkName: "someModule" */
              'someModule'
            )
          `},
						},
					},
				},
			},
			{
				Code: `
            import(
              /* webpackChunkName: "someModule" */
              /* webpackMode: "eager" */
              'someModule'
            )
          `,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: `dynamic imports using eager mode do not need a webpackChunkName`,
						Line:    2, Column: 13, EndLine: 6, EndColumn: 14,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							// Remove webpackChunkName
							{Output: `
            import(
              
              /* webpackMode: "eager" */
              'someModule'
            )
          `},
							// Remove webpackMode
							{Output: `
            import(
              /* webpackChunkName: "someModule" */
              
              'someModule'
            )
          `},
						},
					},
				},
			},
		},
	)
}

// Every example from https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/docs/rules/dynamic-import-chunkname.md,
// run with the configuration the page shows. Upstream tests only assert
// messages; the ranges and suggestion outputs were recorded from the pinned rule.
func TestDynamicImportChunknameDocs(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &dynamic_import_chunkname.DynamicImportChunknameRule,
		[]rule_tester.ValidTestCase{
			// docs valid (configuration shown in the docs)
			{Code: `import(
  /* webpackChunkName: "someModule" */
  'someModule',
);`, Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}, "webpackChunknameFormat": "[a-zA-Z0-57-9-/_]+", "allowEmpty": false}}},
			{Code: `import(
  /* webpackChunkName: "someOtherModule12345789" */
  'someModule',
);`, Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}, "webpackChunknameFormat": "[a-zA-Z0-57-9-/_]+", "allowEmpty": false}}},
			{Code: `import(
  /* webpackChunkName: "someModule" */
  /* webpackPrefetch: true */
  'someModule',
);`, Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}, "webpackChunknameFormat": "[a-zA-Z0-57-9-/_]+", "allowEmpty": false}}},
			{Code: `import(
  /* webpackChunkName: "someModule", webpackPrefetch: true */
  'someModule',
);`, Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}, "webpackChunknameFormat": "[a-zA-Z0-57-9-/_]+", "allowEmpty": false}}},
			{Code: `import(
  /* webpackChunkName: 'someModule' */
  'someModule',
);`, Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}, "webpackChunknameFormat": "[a-zA-Z0-57-9-/_]+", "allowEmpty": false}}},
			// docs allowEmpty: true, valid
			{Code: `import('someModule');`, Options: []any{map[string]any{"allowEmpty": true}}},
			{Code: `import(
  /* webpackChunkName: "someModule" */
  'someModule',
);`, Options: []any{map[string]any{"allowEmpty": true}}},
		},
		[]rule_tester.InvalidTestCase{
			// docs invalid (configuration shown in the docs)
			{
				Code:    `import('someModule');`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}, "webpackChunknameFormat": "[a-zA-Z0-57-9-/_]+", "allowEmpty": false}},
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a leading comment with the webpack chunkname`, Line: 1, Column: 1, EndLine: 1, EndColumn: 21},
				},
			},
			{
				Code: `import(
  /*webpackChunkName:"someModule"*/
  'someModule',
);`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}, "webpackChunknameFormat": "[a-zA-Z0-57-9-/_]+", "allowEmpty": false}},
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a block comment padded with spaces - /* foo */`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName : "someModule" */
  'someModule',
);`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}, "webpackChunknameFormat": "[a-zA-Z0-57-9-/_]+", "allowEmpty": false}},
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName: "someModule6" */
  'someModule',
);`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}, "webpackChunknameFormat": "[a-zA-Z0-57-9-/_]+", "allowEmpty": false}},
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a leading comment in the form /*webpackChunkName: ["'][a-zA-Z0-57-9-/_]+["'],? */`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* totally not webpackChunkName: "someModule" */
  'someModule',
);`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}, "webpackChunknameFormat": "[a-zA-Z0-57-9-/_]+", "allowEmpty": false}},
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  // webpackChunkName: "someModule"
  'someModule',
);`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}, "webpackChunknameFormat": "[a-zA-Z0-57-9-/_]+", "allowEmpty": false}},
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a /* foo */ style comment, not a // foo comment`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackMode: "eager" */
  /* webpackChunkName: "someModule" */
  'someModule',
)`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}, "webpackChunknameFormat": "[a-zA-Z0-57-9-/_]+", "allowEmpty": false}},
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports using eager mode do not need a webpackChunkName`, Line: 1, Column: 1, EndLine: 5, EndColumn: 2,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{Output: `import(
  /* webpackMode: "eager" */
  
  'someModule',
)`}, // Remove webpackChunkName
							{Output: `import(
  
  /* webpackChunkName: "someModule" */
  'someModule',
)`}, // Remove webpackMode
						}},
				},
			},
			// docs allowEmpty: true, invalid
			{
				Code: `import(
  /*webpackChunkName:"someModule"*/
  'someModule',
);`,
				Options: []any{map[string]any{"allowEmpty": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a block comment padded with spaces - /* foo */`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
		},
	)
}
