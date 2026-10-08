package dynamic_import_chunkname_test

import (
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/dynamic_import_chunkname"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Cases beyond the pinned upstream suite. Expected diagnostics, ranges and
// suggestion outputs were recorded from the upstream v2.32.0 rule running under
// ESLint 8.57.1 with typescript-eslint, after widening only its magic comment
// keys to accept the rspack prefix and webpackFetchPriority (the intended
// differences, see the rule documentation). Everything not involving those two
// differences matches the unmodified upstream rule.
func TestDynamicImportChunknameExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &dynamic_import_chunkname.DynamicImportChunknameRule,
		[]rule_tester.ValidTestCase{
			// rspack prefix: every magic comment key accepts the rspack prefix
			{Code: `import(
  /* rspackChunkName: "a" */
  /* rspackChunkName: "a" */
  'm',
)`},
			{Code: `import(
  /* rspackChunkName: "a" */
  /* rspackMode: "lazy" */
  'm',
)`},
			{Code: `import(
  /* rspackChunkName: "a" */
  /* rspackPrefetch: true */
  'm',
)`},
			{Code: `import(
  /* rspackChunkName: "a" */
  /* rspackPrefetch: 10 */
  'm',
)`},
			{Code: `import(
  /* rspackChunkName: "a" */
  /* rspackPreload: false */
  'm',
)`},
			{Code: `import(
  /* rspackChunkName: "a" */
  /* rspackIgnore: true */
  'm',
)`},
			{Code: `import(
  /* rspackChunkName: "a" */
  /* rspackInclude: /\.json$/ */
  'm',
)`},
			{Code: `import(
  /* rspackChunkName: "a" */
  /* rspackExclude: /\.json$/ */
  'm',
)`},
			{Code: `import(
  /* rspackChunkName: "a" */
  /* rspackExports: ["default", "named"] */
  'm',
)`},
			{Code: `import(
  /* rspackChunkName: "a" */
  /* rspackFetchPriority: "high" */
  'm',
)`},
			{Code: `import(
  /* rspackChunkName: 'a', rspackPrefetch: true */
  'm',
)`},
			{Code: `import(
  /* rspackChunkName: "a", rspackExports: ['default', 'named'] */
  'm',
)`},
			{Code: `import(
  /* rspackPrefetch: true */
  /* rspackChunkName: "a" */
  'm',
)`},
			{Code: `dynamicImport(
  /* rspackChunkName: "someModule" */
  'm',
)`, Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}}},
			// mixed prefixes
			{Code: `import(
  /* webpackChunkName: "a" */
  /* rspackPrefetch: true */
  'm',
)`},
			{Code: `import(
  /* rspackChunkName: "a" */
  /* webpackPreload: true */
  'm',
)`},
			{Code: `import(
  /* webpackChunkName: "a", rspackPrefetch: true */
  'm',
)`},
			{Code: `import(
  /* rspackChunkName: "a", webpackMode: "lazy" */
  'm',
)`},
			// webpackFetchPriority and the rspack equivalents
			{Code: `import(
  /* webpackChunkName: "a" */
  /* webpackFetchPriority: "low" */
  'm',
)`},
			{Code: `import(
  /* webpackChunkName: "a" */
  /* webpackFetchPriority: "high" */
  'm',
)`},
			{Code: `import(
  /* webpackChunkName: "a" */
  /* webpackFetchPriority: "auto" */
  'm',
)`},
			{Code: `import(
  /* rspackChunkName: "a", rspackFetchPriority: 'low' */
  'm',
)`},
			// rspack prefix: invalid comments
			{Code: `import(
  /* rspackChunkName: "some-module" */
  'm',
)`, Options: []any{map[string]any{"webpackChunknameFormat": "[a-zA-Z-_/.]+"}}},
			{Code: `import(
  /* rspackChunkName: "abc" */
  'm',
)`, Options: []any{map[string]any{"webpackChunknameFormat": "[a-zA-Z-_/.]+"}}},
			{Code: `import(
  /* rspackPrefetch: true */
  'm',
)`, Options: []any{map[string]any{"allowEmpty": true}}},
			// rspack prefix: eager mode with a chunk name
			{Code: `import(
  /* rspackMode: "eager" */
  'm',
)`},
			{Code: `import(
  /* rspackMode: "eager" */
  'm',
)`, Options: []any{map[string]any{"allowEmpty": true}}},
			// magic comment body: syntax checks
			{Code: `import(
  /* webpackChunkName: "a" */
  'm',
)`},
			{Code: `import(
  /* webpackChunkName: "a", */
  'm',
)`},
			{Code: `import(
  /* webpackChunkName: "a" */
  'm',
)`},
			{Code: `import(
  /* webpackChunkName: 'a' */
  'm',
)`},
			// leading comments: placement and token shapes
			{Code: `import(/* webpackChunkName: "a" */ 'm')`},
			{Code: `import( /* webpackChunkName: "a" */ 'm' )`},
			{Code: `import((/* webpackChunkName: "a" */ 'm'))`},
			{Code: "import(/* webpackChunkName: \"a\" */ `m`)"},
			{Code: "import(/* webpackChunkName: \"a\" */ `${name}`)"},
			{Code: `import(/* webpackChunkName: "a" */ name)`},
			{Code: `import(/* webpackChunkName: "a" */ './' + name)`},
			{Code: `import(/* webpackChunkName: "a" */ 'm', { with: { type: 'json' } })`},
			{Code: `import(/* webpackChunkName: "a" */ 'm' as string)`},
			{Code: `import(/* webpackChunkName: "a" */ <string>'m')`},
			{Code: `const m = await import(
  /* webpackChunkName: "a" */
  'm'
)`},
			{Code: `import(
  /* webpackChunkName: "a" */

  /* webpackPrefetch: true */
  'm'
)`},
			{Code: "import(\r\n  /* webpackChunkName: \"a\" */\r\n  'm'\r\n)"},
			{Code: `const あ = 1; import(/* webpackChunkName: "a" */ 'm')`},
			// call shapes and importFunctions
			{Code: `dynamicImport(/* webpackChunkName: "a" */ 'm')`, Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}}},
			{Code: `dynamicImport(/* rspackChunkName: "a" */ 'm')`, Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}}},
			{Code: `dynamicImport('m')`},
			{Code: `dynamicImport('m')`, Options: []any{map[string]any{}}},
			{Code: `dynamicImport('m')`, Options: []any{map[string]any{"importFunctions": []any{}}}},
			{Code: `dynamicImport('m')`, Options: []any{map[string]any{"importFunctions": []any{"other"}}}},
			{Code: `dynamicImport?.(/* webpackChunkName: "a" */ 'm')`, Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}}},
			{Code: `(dynamicImport as any)('m')`, Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}}},
			{Code: `dynamicImport!('m')`, Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}}},
			{Code: `new dynamicImport('m')`, Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}}},
			{Code: `obj.dynamicImport('m')`, Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}}},
			{Code: `obj?.dynamicImport('m')`, Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}}},
			{Code: `dynamicImport.call(null, 'm')`, Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}}},
			{Code: "dynamicImport`m`", Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}}},
			{Code: `dynamicImport()`, Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}}},
			{Code: `dynamicImport(/* webpackChunkName: "a" */)`, Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}}},
			{Code: `dynamicImport(/* webpackChunkName: "a" */ ...args)`, Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}}},
			{Code: `import.meta.url`, Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}}},
			{Code: `type T = typeof import('m')`},
			{Code: `type T = import('m').Foo`},
			// options
			{Code: `import(
  /* webpackChunkName: "a" */
  'm',
)`, Options: []any{map[string]any{"webpackChunknameFormat": "[a-z]+"}}},
			{Code: `import(
  /* webpackChunkName: "ab" */
  'm',
)`, Options: []any{map[string]any{"webpackChunknameFormat": "a|b"}}},
			{Code: `import(
  /* webpackChunkName: "ab" */
  'm',
)`, Options: []any{map[string]any{"webpackChunknameFormat": "\\w+"}}},
			{Code: `import(
  /* webpackChunkName: "[index]" */
  'm',
)`},
			{Code: `import(
  /* webpackChunkName: "a.b/c_d-e" */
  'm',
)`},
			{Code: `import(
  /* webpackChunkName: "" */
  'm',
)`, Options: []any{map[string]any{"webpackChunknameFormat": ""}}},
			{Code: `import(
  /* webpackChunkName: "a" */
  'm',
)`, Options: []any{map[string]any{"allowEmpty": false}}},
			{Code: `import('m')`, Options: []any{map[string]any{"allowEmpty": true}}},
			{Code: `import(
  /* webpackChunkName: "a" */
  /* webpackChunkName: "b" */
  'm',
)`},
			{Code: `import(
  /* webpackPrefetch: true */
  'm',
)`, Options: []any{map[string]any{"allowEmpty": true}}},
			{Code: `import(
  /* webpackChunkName: "a" */
  'm',
)`, Options: []any{map[string]any{"importFunctions": []any{"x"}, "allowEmpty": true, "webpackChunknameFormat": "[a-z]+"}}},
			// suggestions
			{Code: `import(/* rspackMode: "eager", rspackChunkName: "someModule6" */ 'm')`, Options: []any{map[string]any{"webpackChunknameFormat": "[a-zA-Z-_/.]+"}}},

			// Upstream throws a TypeError while reading the missing first argument;
			// there is no argument to inspect, so nothing is reported.
			{Code: `dynamicImport()`, Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}}},
			{Code: `dynamicImport(/* webpackChunkName: "a" */)`, Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}}},
			// Upstream compiles the format while creating the rule, so an invalid
			// pattern aborts the lint run; rslint reports nothing for it.
			{Code: "import('m')", Options: []any{map[string]any{"webpackChunknameFormat": "("}}},
		},
		[]rule_tester.InvalidTestCase{
			// webpackFetchPriority and the rspack equivalents
			{
				Code: `import(
  /* webpackChunkName: "a" */
  /* webpackFetchPriority: "fast" */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 5, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* rspackChunkName: "a" */
  /* rspackFetchPriority: fast */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 5, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* rspackChunkName: "a" */
  /* rspackFetchPriority: 'HIGH' */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 5, EndColumn: 2},
				},
			},
			// rspack prefix: invalid comments
			{
				Code: `import(
  /* rspackPrefetch: true */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a leading comment in the form /*webpackChunkName: ["']([0-9a-zA-Z-_/.]|\[(request|index)\])+["'],? */`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* rspackMode: "lazy" */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a leading comment in the form /*webpackChunkName: ["']([0-9a-zA-Z-_/.]|\[(request|index)\])+["'],? */`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /*rspackChunkName: "a"*/
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a block comment padded with spaces - /* foo */`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* rspackChunkName : "a" */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* rspackChunkName: a */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* rspackChunkName: "a" */
  // rspackPrefetch: true
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a /* foo */ style comment, not a // foo comment`, Line: 1, Column: 1, EndLine: 5, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* rspackFoo: "a" */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* rspackignore: true */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* Rspack: 1 */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* rspackIgnore: 1 */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* rspackMode: "other" */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* rspackChunkName: "someModule6" */
  'm',
)`,
				Options: []any{map[string]any{"webpackChunknameFormat": "[a-zA-Z-_/.]+"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a leading comment in the form /*webpackChunkName: ["'][a-zA-Z-_/.]+["'],? */`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /*rspackPrefetch: true*/
  'm',
)`,
				Options: []any{map[string]any{"allowEmpty": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a block comment padded with spaces - /* foo */`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			// rspack prefix: eager mode with a chunk name
			{
				Code: `import(
  /* rspackMode: "eager" */
  /* rspackChunkName: "a" */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports using eager mode do not need a webpackChunkName`, Line: 1, Column: 1, EndLine: 5, EndColumn: 2,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{Output: `import(
  /* rspackMode: "eager" */
  
  'm',
)`}, // Remove webpackChunkName
							{Output: `import(
  
  /* rspackChunkName: "a" */
  'm',
)`}, // Remove webpackMode
						}},
				},
			},
			{
				Code: `import(
  /* rspackChunkName: "a" */
  /* rspackMode: "eager" */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports using eager mode do not need a webpackChunkName`, Line: 1, Column: 1, EndLine: 5, EndColumn: 2,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{Output: `import(
  
  /* rspackMode: "eager" */
  'm',
)`}, // Remove webpackChunkName
							{Output: `import(
  /* rspackChunkName: "a" */
  
  'm',
)`}, // Remove webpackMode
						}},
				},
			},
			{
				Code: `import(
  /* rspackMode: "eager", rspackChunkName: "a" */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports using eager mode do not need a webpackChunkName`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{Output: `import(
  /* rspackMode: "eager" */
  'm',
)`}, // Remove webpackChunkName
							{Output: `import(
  /* rspackChunkName: "a" */
  'm',
)`}, // Remove webpackMode
						}},
				},
			},
			{
				Code: `import(
  /* rspackChunkName: "a", rspackMode: 'eager' */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports using eager mode do not need a webpackChunkName`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{Output: `import(
  /* rspackMode: 'eager' */
  'm',
)`}, // Remove webpackChunkName
							{Output: `import(
  /* rspackChunkName: "a" */
  'm',
)`}, // Remove webpackMode
						}},
				},
			},
			{
				Code: `import(
  /* rspackChunkName: "a", rspackPrefetch: true, rspackMode: "eager" */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports using eager mode do not need a webpackChunkName`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{Output: `import(
  /* rspackPrefetch: true, rspackMode: "eager" */
  'm',
)`}, // Remove webpackChunkName
							{Output: `import(
  /* rspackChunkName: "a", rspackPrefetch: true */
  'm',
)`}, // Remove webpackMode
						}},
				},
			},
			{
				Code: `import(
  /* webpackMode: "eager" */
  /* rspackChunkName: "a" */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports using eager mode do not need a webpackChunkName`, Line: 1, Column: 1, EndLine: 5, EndColumn: 2,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{Output: `import(
  /* webpackMode: "eager" */
  
  'm',
)`}, // Remove webpackChunkName
							{Output: `import(
  
  /* rspackChunkName: "a" */
  'm',
)`}, // Remove webpackMode
						}},
				},
			},
			{
				Code: `import(
  /* rspackMode: "eager" */
  /* webpackChunkName: "a" */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports using eager mode do not need a webpackChunkName`, Line: 1, Column: 1, EndLine: 5, EndColumn: 2,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{Output: `import(
  /* rspackMode: "eager" */
  
  'm',
)`}, // Remove webpackChunkName
							{Output: `import(
  
  /* webpackChunkName: "a" */
  'm',
)`}, // Remove webpackMode
						}},
				},
			},
			{
				Code: `import(
  /* rspackMode: "eager", rspackChunkName: "a" */
  'm',
)`,
				Options: []any{map[string]any{"allowEmpty": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports using eager mode do not need a webpackChunkName`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{Output: `import(
  /* rspackMode: "eager" */
  'm',
)`}, // Remove webpackChunkName
							{Output: `import(
  /* rspackChunkName: "a" */
  'm',
)`}, // Remove webpackMode
						}},
				},
			},
			{
				Code: `import(/* rspackMode: "eager", rspackChunkName: "a" */ 'm')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports using eager mode do not need a webpackChunkName`, Line: 1, Column: 1, EndLine: 1, EndColumn: 60,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{Output: `import(/* rspackMode: "eager" */ 'm')`},  // Remove webpackChunkName
							{Output: `import(/* rspackChunkName: "a" */ 'm')`}, // Remove webpackMode
						}},
				},
			},
			// magic comment body: syntax checks
			{
				Code: `import(
  /* webpackChunkName: "a",, */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* , webpackChunkName: "a" */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName: "a"; */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName "a" */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName: "a" webpackPrefetch: true */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName: "a" // x */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName: "a" /* x */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName: "a" as string */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName: "a"! */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName: <string>"a" */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName: "a" satisfies string */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName: foo */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName: undefined */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a leading comment in the form /*webpackChunkName: ["']([0-9a-zA-Z-_/.]|\[(request|index)\])+["'],? */`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName: NaN */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a leading comment in the form /*webpackChunkName: ["']([0-9a-zA-Z-_/.]|\[(request|index)\])+["'],? */`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName: Infinity */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a leading comment in the form /*webpackChunkName: ["']([0-9a-zA-Z-_/.]|\[(request|index)\])+["'],? */`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName: typeof foo */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a leading comment in the form /*webpackChunkName: ["']([0-9a-zA-Z-_/.]|\[(request|index)\])+["'],? */`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName: foo.bar */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName: Math.max(1) */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a leading comment in the form /*webpackChunkName: ["']([0-9a-zA-Z-_/.]|\[(request|index)\])+["'],? */`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName: console */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a leading comment in the form /*webpackChunkName: ["']([0-9a-zA-Z-_/.]|\[(request|index)\])+["'],? */`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName: arguments */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a leading comment in the form /*webpackChunkName: ["']([0-9a-zA-Z-_/.]|\[(request|index)\])+["'],? */`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName: this */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a leading comment in the form /*webpackChunkName: ["']([0-9a-zA-Z-_/.]|\[(request|index)\])+["'],? */`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName: [foo] */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName: {a: foo} */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: "import(\n  /* webpackChunkName: `a${foo}` */\n  'm',\n)",
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName: "a" + foo */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName: function () { return foo } */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a leading comment in the form /*webpackChunkName: ["']([0-9a-zA-Z-_/.]|\[(request|index)\])+["'],? */`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName: () => foo */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a leading comment in the form /*webpackChunkName: ["']([0-9a-zA-Z-_/.]|\[(request|index)\])+["'],? */`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName: "a", webpackPrefetch: foo */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName, */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* ...foo, webpackChunkName: "a" */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName() { return 1 } */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* get webpackChunkName() { return "a" } */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* [foo]: "a", webpackChunkName: "a" */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* "webpackChunkName": "a" */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName: "a"
   */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a block comment padded with spaces - /* foo */`, Line: 1, Column: 1, EndLine: 5, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName: "a",
 webpackPrefetch: true */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 5, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /*  webpackChunkName: "a" */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a block comment padded with spaces - /* foo */`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName: "a"  */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a block comment padded with spaces - /* foo */`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /**/
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a block comment padded with spaces - /* foo */`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a block comment padded with spaces - /* foo */`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /*  */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a block comment padded with spaces - /* foo */`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* a */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a block comment padded with spaces - /* foo */`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /** webpackChunkName: "a" */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a block comment padded with spaces - /* foo */`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName: "a" **/
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a block comment padded with spaces - /* foo */`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /*	webpackChunkName: "a" */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a block comment padded with spaces - /* foo */`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			// leading comments: placement and token shapes
			{
				Code: `import('m' /* webpackChunkName: "a" */)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a leading comment with the webpack chunkname`, Line: 1, Column: 1, EndLine: 1, EndColumn: 40},
				},
			},
			{
				Code: `import('m') /* webpackChunkName: "a" */`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a leading comment with the webpack chunkname`, Line: 1, Column: 1, EndLine: 1, EndColumn: 12},
				},
			},
			{
				Code: `/* webpackChunkName: "a" */ import('m')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a leading comment with the webpack chunkname`, Line: 1, Column: 29, EndLine: 1, EndColumn: 40},
				},
			},
			{
				Code: `import(/* webpackChunkName: "a" */ ('m'))`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a leading comment with the webpack chunkname`, Line: 1, Column: 1, EndLine: 1, EndColumn: 42},
				},
			},
			{
				Code: `import(/* webpackChunkName: "a" */ (/* webpackPrefetch: true */ 'm'))`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a leading comment in the form /*webpackChunkName: ["']([0-9a-zA-Z-_/.]|\[(request|index)\])+["'],? */`, Line: 1, Column: 1, EndLine: 1, EndColumn: 70},
				},
			},
			{
				Code: `import('m', { /* webpackChunkName: "a" */ with: { type: 'json' } })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a leading comment with the webpack chunkname`, Line: 1, Column: 1, EndLine: 1, EndColumn: 68},
				},
			},
			{
				Code: `import('m', /* webpackChunkName: "a" */ { with: { type: 'json' } })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a leading comment with the webpack chunkname`, Line: 1, Column: 1, EndLine: 1, EndColumn: 68},
				},
			},
			{
				Code: `import(/* webpackChunkName: "a" */ 'm')
  .then(() => import('n'))`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a leading comment with the webpack chunkname`, Line: 2, Column: 15, EndLine: 2, EndColumn: 26},
				},
			},
			{
				Code: `foo(import('m'))`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a leading comment with the webpack chunkname`, Line: 1, Column: 5, EndLine: 1, EndColumn: 16},
				},
			},
			{
				Code: `import(
  // a
  /* webpackChunkName: "a" */
  'm'
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a /* foo */ style comment, not a // foo comment`, Line: 1, Column: 1, EndLine: 5, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName: "a" */
  // a
  'm'
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a /* foo */ style comment, not a // foo comment`, Line: 1, Column: 1, EndLine: 5, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName: "a" */ // a
  'm'
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a /* foo */ style comment, not a // foo comment`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: "import(\r\n  'm'\r\n)",
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a leading comment with the webpack chunkname`, Line: 1, Column: 1, EndLine: 3, EndColumn: 2},
				},
			},
			{
				Code: `import(/* 日本語 */ 'm')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 1, EndColumn: 22},
				},
			},
			{
				Code: `const あ = 1; import('m')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a leading comment with the webpack chunkname`, Line: 1, Column: 14, EndLine: 1, EndColumn: 25},
				},
			},
			// call shapes and importFunctions
			{
				Code:    `dynamicImport('m')`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a leading comment with the webpack chunkname`, Line: 1, Column: 1, EndLine: 1, EndColumn: 19},
				},
			},
			{
				Code:    `other('m')`,
				Options: []any{map[string]any{"importFunctions": []any{"other", "dynamicImport"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a leading comment with the webpack chunkname`, Line: 1, Column: 1, EndLine: 1, EndColumn: 11},
				},
			},
			{
				Code:    `dynamicImport('m'); other('n')`,
				Options: []any{map[string]any{"importFunctions": []any{"other", "dynamicImport"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a leading comment with the webpack chunkname`, Line: 1, Column: 1, EndLine: 1, EndColumn: 19},
					{Message: `dynamic imports require a leading comment with the webpack chunkname`, Line: 1, Column: 21, EndLine: 1, EndColumn: 31},
				},
			},
			{
				Code:    `dynamicImport?.('m')`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a leading comment with the webpack chunkname`, Line: 1, Column: 1, EndLine: 1, EndColumn: 21},
				},
			},
			{
				Code:    `(dynamicImport)('m')`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a leading comment with the webpack chunkname`, Line: 1, Column: 1, EndLine: 1, EndColumn: 21},
				},
			},
			{
				Code:    `dynamicImport<string>('m')`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a leading comment with the webpack chunkname`, Line: 1, Column: 1, EndLine: 1, EndColumn: 27},
				},
			},
			{
				Code:    `dynamicImport(...args)`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a leading comment with the webpack chunkname`, Line: 1, Column: 1, EndLine: 1, EndColumn: 23},
				},
			},
			{
				Code:    `require('m')`,
				Options: []any{map[string]any{"importFunctions": []any{"require"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a leading comment with the webpack chunkname`, Line: 1, Column: 1, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code:    `import('m')`,
				Options: []any{map[string]any{"importFunctions": []any{"import"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a leading comment with the webpack chunkname`, Line: 1, Column: 1, EndLine: 1, EndColumn: 12},
				},
			},
			{
				Code:    `function f(dynamicImport) { return dynamicImport('m') }`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a leading comment with the webpack chunkname`, Line: 1, Column: 36, EndLine: 1, EndColumn: 54},
				},
			},
			{
				Code: `dynamicImport('m')
import('n')`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a leading comment with the webpack chunkname`, Line: 1, Column: 1, EndLine: 1, EndColumn: 19},
					{Message: `dynamic imports require a leading comment with the webpack chunkname`, Line: 2, Column: 1, EndLine: 2, EndColumn: 12},
				},
			},
			{
				Code: `dynamicImport(
  /* webpackMode: "eager", webpackChunkName: "a" */
  'm'
)`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports using eager mode do not need a webpackChunkName`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{Output: `dynamicImport(
  /* webpackMode: "eager" */
  'm'
)`}, // Remove webpackChunkName
							{Output: `dynamicImport(
  /* webpackChunkName: "a" */
  'm'
)`}, // Remove webpackMode
						}},
				},
			},
			// options
			{
				Code: `import(
  /* webpackChunkName: "a1" */
  'm',
)`,
				Options: []any{map[string]any{"webpackChunknameFormat": "[a-z]+"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a leading comment in the form /*webpackChunkName: ["'][a-z]+["'],? */`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName: "A" */
  'm',
)`,
				Options: []any{map[string]any{"webpackChunknameFormat": "[a-z]+"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a leading comment in the form /*webpackChunkName: ["'][a-z]+["'],? */`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName: "ab" */
  'm',
)`,
				Options: []any{map[string]any{"webpackChunknameFormat": "[a-z]"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a leading comment in the form /*webpackChunkName: ["'][a-z]["'],? */`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName: "ab" */
  'm',
)`,
				Options: []any{map[string]any{"webpackChunknameFormat": "^[a-z]+$"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a leading comment in the form /*webpackChunkName: ["']^[a-z]+$["'],? */`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName: "[request]" */
  'm',
)`,
				Options: []any{map[string]any{"webpackChunknameFormat": "\\w+"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a leading comment in the form /*webpackChunkName: ["']\w+["'],? */`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName: "[name]" */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a leading comment in the form /*webpackChunkName: ["']([0-9a-zA-Z-_/.]|\[(request|index)\])+["'],? */`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName: "" */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a leading comment in the form /*webpackChunkName: ["']([0-9a-zA-Z-_/.]|\[(request|index)\])+["'],? */`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName: "a" */
  'm',
)`,
				Options: []any{map[string]any{"webpackChunknameFormat": ""}},
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a leading comment in the form /*webpackChunkName: ["']["'],? */`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  // a
  'm'
)`,
				Options: []any{map[string]any{"allowEmpty": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a /* foo */ style comment, not a // foo comment`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code:    `import(/*a*/ 'm')`,
				Options: []any{map[string]any{"allowEmpty": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a block comment padded with spaces - /* foo */`, Line: 1, Column: 1, EndLine: 1, EndColumn: 18},
				},
			},
			{
				Code:    `import(/* a */ 'm')`,
				Options: []any{map[string]any{"allowEmpty": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a block comment padded with spaces - /* foo */`, Line: 1, Column: 1, EndLine: 1, EndColumn: 20},
				},
			},
			{
				Code:    `import(/* webpackChunkName: someModule */ 'm')`,
				Options: []any{map[string]any{"allowEmpty": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 1, EndColumn: 47},
				},
			},
			{
				Code: `import(
  /* webpackChunkName: "a" */
  // x
  'm',
)`,
				Options: []any{map[string]any{"allowEmpty": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a /* foo */ style comment, not a // foo comment`, Line: 1, Column: 1, EndLine: 5, EndColumn: 2},
				},
			},
			// suggestions
			{
				Code: `import(/* webpackMode: "eager", webpackChunkName: "a" */ 'm')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports using eager mode do not need a webpackChunkName`, Line: 1, Column: 1, EndLine: 1, EndColumn: 62,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{Output: `import(/* webpackMode: "eager" */ 'm')`},  // Remove webpackChunkName
							{Output: `import(/* webpackChunkName: "a" */ 'm')`}, // Remove webpackMode
						}},
				},
			},
			{
				Code: `import(/* webpackChunkName: "a", webpackMode: "eager" */ 'm')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports using eager mode do not need a webpackChunkName`, Line: 1, Column: 1, EndLine: 1, EndColumn: 62,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{Output: `import(/* webpackMode: "eager" */ 'm')`},  // Remove webpackChunkName
							{Output: `import(/* webpackChunkName: "a" */ 'm')`}, // Remove webpackMode
						}},
				},
			},
			{
				Code: `import(/* webpackChunkName: "a", webpackMode: "eager", webpackPrefetch: true */ 'm')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports using eager mode do not need a webpackChunkName`, Line: 1, Column: 1, EndLine: 1, EndColumn: 85,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{Output: `import(/* webpackMode: "eager", webpackPrefetch: true */ 'm')`},  // Remove webpackChunkName
							{Output: `import(/* webpackChunkName: "a", webpackPrefetch: true */ 'm')`}, // Remove webpackMode
						}},
				},
			},
			{
				Code: `import(/* webpackMode: "eager" */ /* webpackChunkName: "a" */ 'm')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports using eager mode do not need a webpackChunkName`, Line: 1, Column: 1, EndLine: 1, EndColumn: 67,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{Output: `import(/* webpackMode: "eager" */  'm')`},  // Remove webpackChunkName
							{Output: `import( /* webpackChunkName: "a" */ 'm')`}, // Remove webpackMode
						}},
				},
			},
			{
				Code: `import(/* webpackChunkName: "a" */ /* webpackMode: "eager" */ 'm')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports using eager mode do not need a webpackChunkName`, Line: 1, Column: 1, EndLine: 1, EndColumn: 67,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{Output: `import( /* webpackMode: "eager" */ 'm')`},  // Remove webpackChunkName
							{Output: `import(/* webpackChunkName: "a" */  'm')`}, // Remove webpackMode
						}},
				},
			},
			{
				Code: `import(/* webpackChunkName: "a" */ /* webpackChunkName: "b" */ /* webpackMode: "eager" */ 'm')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports using eager mode do not need a webpackChunkName`, Line: 1, Column: 1, EndLine: 1, EndColumn: 95,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{Output: `import( /* webpackChunkName: "b" */ /* webpackMode: "eager" */ 'm')`},  // Remove webpackChunkName
							{Output: `import(/* webpackChunkName: "a" */ /* webpackChunkName: "b" */  'm')`}, // Remove webpackMode
						}},
				},
			},
			{
				Code: `import(/* webpackMode: "eager" */ /* webpackMode: 'eager' */ /* webpackChunkName: "a" */ 'm')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports using eager mode do not need a webpackChunkName`, Line: 1, Column: 1, EndLine: 1, EndColumn: 94,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{Output: `import(/* webpackMode: "eager" */ /* webpackMode: 'eager' */  'm')`},  // Remove webpackChunkName
							{Output: `import( /* webpackMode: 'eager' */ /* webpackChunkName: "a" */ 'm')`}, // Remove webpackMode
						}},
				},
			},
			{
				Code: `import(/* webpackChunkName: "a", webpackChunkName: "b", webpackMode: "eager" */ 'm')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports using eager mode do not need a webpackChunkName`, Line: 1, Column: 1, EndLine: 1, EndColumn: 85,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{Output: `import(/* webpackChunkName: "b", webpackMode: "eager" */ 'm')`},  // Remove webpackChunkName
							{Output: `import(/* webpackChunkName: "a", webpackChunkName: "b" */ 'm')`}, // Remove webpackMode
						}},
				},
			},
			{
				Code: `import(/* webpackChunkName: "a" */ /* webpackMode: "eager", webpackPrefetch: true */ 'm')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports using eager mode do not need a webpackChunkName`, Line: 1, Column: 1, EndLine: 1, EndColumn: 90,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{Output: `import( /* webpackMode: "eager", webpackPrefetch: true */ 'm')`},      // Remove webpackChunkName
							{Output: `import(/* webpackChunkName: "a" */ /* webpackPrefetch: true */ 'm')`}, // Remove webpackMode
						}},
				},
			},
			{
				Code: `import(/* webpackPrefetch: true, webpackChunkName: "a" */ /* webpackMode: "eager" */ 'm')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports using eager mode do not need a webpackChunkName`, Line: 1, Column: 1, EndLine: 1, EndColumn: 90,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{Output: `import(/* webpackPrefetch: true */ /* webpackMode: "eager" */ 'm')`}, // Remove webpackChunkName
							{Output: `import(/* webpackPrefetch: true, webpackChunkName: "a" */  'm')`},    // Remove webpackMode
						}},
				},
			},
			{
				Code: `import(/* webpackMode: 'eager',webpackChunkName: "a" */ 'm')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 1, EndColumn: 61},
				},
			},
			{
				Code: `import(
  /* webpackMode: "eager" */
  /* webpackChunkName: "a" */
  'm'
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports using eager mode do not need a webpackChunkName`, Line: 1, Column: 1, EndLine: 5, EndColumn: 2,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{Output: `import(
  /* webpackMode: "eager" */
  
  'm'
)`}, // Remove webpackChunkName
							{Output: `import(
  
  /* webpackChunkName: "a" */
  'm'
)`}, // Remove webpackMode
						}},
				},
			},
			{
				Code: `foo(/*x*/ import(/* webpackMode: "eager", webpackChunkName: "a" */ 'm'))`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports using eager mode do not need a webpackChunkName`, Line: 1, Column: 11, EndLine: 1, EndColumn: 72,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{Output: `foo(/*x*/ import(/* webpackMode: "eager" */ 'm'))`},  // Remove webpackChunkName
							{Output: `foo(/*x*/ import(/* webpackChunkName: "a" */ 'm'))`}, // Remove webpackMode
						}},
				},
			},
			{
				Code:    `dynamicImport(/* webpackMode: "eager", webpackChunkName: "a" */ 'm')`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports using eager mode do not need a webpackChunkName`, Line: 1, Column: 1, EndLine: 1, EndColumn: 69,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{Output: `dynamicImport(/* webpackMode: "eager" */ 'm')`},  // Remove webpackChunkName
							{Output: `dynamicImport(/* webpackChunkName: "a" */ 'm')`}, // Remove webpackMode
						}},
				},
			},
			{
				Code:    `dynamicImport(/* rspackMode: "eager", rspackChunkName: "a" */ 'm')`,
				Options: []any{map[string]any{"importFunctions": []any{"dynamicImport"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports using eager mode do not need a webpackChunkName`, Line: 1, Column: 1, EndLine: 1, EndColumn: 67,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{Output: `dynamicImport(/* rspackMode: "eager" */ 'm')`},  // Remove webpackChunkName
							{Output: `dynamicImport(/* rspackChunkName: "a" */ 'm')`}, // Remove webpackMode
						}},
				},
			},
			{
				Code: `import(/* rspackChunkName: "a", webpackMode: "eager" */ 'm')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports using eager mode do not need a webpackChunkName`, Line: 1, Column: 1, EndLine: 1, EndColumn: 61,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{Output: `import(/* webpackMode: "eager" */ 'm')`}, // Remove webpackChunkName
							{Output: `import(/* rspackChunkName: "a" */ 'm')`}, // Remove webpackMode
						}},
				},
			},
			{
				Code: `import(/* rspackMode: "eager", rspackChunkName: "a", rspackPrefetch: true */ 'm')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports using eager mode do not need a webpackChunkName`, Line: 1, Column: 1, EndLine: 1, EndColumn: 82,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{Output: `import(/* rspackMode: "eager", rspackPrefetch: true */ 'm')`},  // Remove webpackChunkName
							{Output: `import(/* rspackChunkName: "a", rspackPrefetch: true */ 'm')`}, // Remove webpackMode
						}},
				},
			},
			{
				Code:    `import(/* rspackMode: "eager", rspackChunkName: "a" */ 'm')`,
				Options: []any{map[string]any{"allowEmpty": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports using eager mode do not need a webpackChunkName`, Line: 1, Column: 1, EndLine: 1, EndColumn: 60,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{Output: `import(/* rspackMode: "eager" */ 'm')`},  // Remove webpackChunkName
							{Output: `import(/* rspackChunkName: "a" */ 'm')`}, // Remove webpackMode
						}},
				},
			},
			{
				Code:    `import(/* webpackMode: "eager", webpackChunkName: "a" */ 'm')`,
				Options: []any{map[string]any{"allowEmpty": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports using eager mode do not need a webpackChunkName`, Line: 1, Column: 1, EndLine: 1, EndColumn: 62,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{Output: `import(/* webpackMode: "eager" */ 'm')`},  // Remove webpackChunkName
							{Output: `import(/* webpackChunkName: "a" */ 'm')`}, // Remove webpackMode
						}},
				},
			},
		},
	)
}

func TestDynamicImportChunknameSchema(t *testing.T) {
	schema := dynamic_import_chunkname.DynamicImportChunknameRule.Schema
	for _, options := range [][]any{
		{},
		{map[string]any{}},
		{map[string]any{"importFunctions": []any{}, "allowEmpty": true, "webpackChunknameFormat": "[a-z]+"}},
		{map[string]any{"importFunctions": []any{"a", "b"}}},
	} {
		if err := schema.Validate(options); err != nil {
			t.Errorf("rejected valid options %#v: %v", options, err)
		}
	}
	for _, options := range [][]any{
		{true},
		{map[string]any{"importFunctions": "a"}},
		{map[string]any{"importFunctions": []any{"a", "a"}}},
		{map[string]any{"importFunctions": []any{1}}},
		{map[string]any{"allowEmpty": "yes"}},
		{map[string]any{"webpackChunknameFormat": 1}},
		{map[string]any{}, map[string]any{}},
	} {
		if err := schema.Validate(options); err == nil {
			t.Errorf("accepted invalid options %#v", options)
		}
	}
}

func TestDynamicImportChunknameEditDemand(t *testing.T) {
	const message = "dynamic imports using eager mode do not need a webpackChunkName"
	for _, tc := range []struct {
		code        string
		suggestions []string
	}{
		{
			code:        `import(/* webpackMode: "eager", webpackChunkName: "a" */ 'm')`,
			suggestions: []string{`import(/* webpackMode: "eager" */ 'm')`, `import(/* webpackChunkName: "a" */ 'm')`},
		},
		{
			code:        `import(/* rspackChunkName: "a" */ /* rspackMode: "eager" */ 'm')`,
			suggestions: []string{`import( /* rspackMode: "eager" */ 'm')`, `import(/* rspackChunkName: "a" */  'm')`},
		},
		{
			code:        `import(/* webpackMode: "eager" */ /* webpackChunkName: "a" */ 'm')`,
			suggestions: []string{`import(/* webpackMode: "eager" */  'm')`, `import( /* webpackChunkName: "a" */ 'm')`},
		},
	} {
		t.Run(tc.code, func(t *testing.T) {
			file := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/test.ts", Path: "/test.ts"}, tc.code, core.ScriptKindTS)
			call := file.Statements.Nodes[0].AsExpressionStatement().Expression
			var all rule.RuleDiagnostic
			for _, demand := range []rule.EditDemand{rule.EditDemandAll, rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion} {
				var diagnostics []rule.RuleDiagnostic
				ctx := (rule.RuleContext{SourceFile: file}).WithDiagnosticConsumer(dynamic_import_chunkname.DynamicImportChunknameRule.Name, rule.SeverityError, rule.DiagnosticConsumer{
					Demand: demand,
					Report: func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) },
				})
				dynamic_import_chunkname.DynamicImportChunknameRule.Run(ctx, nil)[ast.KindCallExpression](call)
				if len(diagnostics) != 1 {
					t.Fatalf("demand %d: got %d diagnostics", demand, len(diagnostics))
				}
				got := diagnostics[0]
				if demand == rule.EditDemandAll {
					all = got
				}
				if got.FixesPtr != nil {
					t.Fatalf("demand %d: unexpected autofix", demand)
				}
				wantSuggestions := demand == rule.EditDemandAll || demand == rule.EditDemandSuggestion
				if (got.Suggestions != nil) != wantSuggestions {
					t.Fatalf("demand %d: incorrect suggestion availability", demand)
				}
				if wantSuggestions {
					if !reflect.DeepEqual(got.Suggestions, all.Suggestions) {
						t.Fatalf("demand %d: suggestions differ from all edits", demand)
					}
					if len(*got.Suggestions) != len(tc.suggestions) {
						t.Fatalf("got %d suggestions", len(*got.Suggestions))
					}
					for index, wantOutput := range tc.suggestions {
						description := []string{"Remove webpackChunkName", "Remove webpackMode"}[index]
						suggestion := (*got.Suggestions)[index]
						if suggestion.Message.Id != "" || suggestion.Message.Description != description {
							t.Fatalf("incorrect suggestion message: %#v", suggestion.Message)
						}
						output, _, _ := linter.ApplyRuleFixes(tc.code, []rule.RuleSuggestion{suggestion})
						if output != wantOutput {
							t.Fatalf("suggestion output %q, want %q", output, wantOutput)
						}
					}
				}
				want := all
				got.Suggestions, want.Suggestions = nil, nil
				if !reflect.DeepEqual(got, want) || got.Message.Description != message || got.Message.Id != "" ||
					got.Range != core.NewTextRange(0, len(tc.code)) {
					t.Fatalf("demand %d: incorrect diagnostic identity", demand)
				}
			}
		})
	}
}

// Documented difference: upstream evaluates the magic comment with
// vm.runInNewContext and so also reports errors thrown while running it, such as
// a ReferenceError inside an immediately invoked arrow function. rslint parses
// the same wrapper and detects syntax errors, TypeScript-only syntax and
// references to undefined globals, but does not run the code.
func TestDynamicImportChunknameEvaluationDifference(t *testing.T) {
	const code = "import(\n  /* webpackChunkName: (() => missing)() */\n  'm'\n)"
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &dynamic_import_chunkname.DynamicImportChunknameRule,
		[]rule_tester.ValidTestCase{
			// Upstream: `dynamic imports require a "webpack" comment with valid syntax`.
			{Code: code, Options: []any{map[string]any{"allowEmpty": true}}},
		},
		[]rule_tester.InvalidTestCase{
			// Upstream: the same "valid syntax" message instead of the chunk name form.
			{
				Code: code,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: `dynamic imports require a leading comment in the form /*webpackChunkName: ["']([0-9a-zA-Z-_/.]|\[(request|index)\])+["'],? */`,
						Line:    1, Column: 1, EndLine: 4, EndColumn: 2,
					},
				},
			},
		},
	)
}
