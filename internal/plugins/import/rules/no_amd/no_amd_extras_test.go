package no_amd_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/no_amd"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Expectations checked against eslint-plugin-import v2.32.0 with
// @typescript-eslint/parser 8.65.0. Empty message IDs, outputs and suggestions
// also assert that the upstream rule supplies no IDs or edits.
func TestNoAmdExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.allow-js.json", t, &no_amd.NoAmdRule,
		[]rule_tester.ValidTestCase{
			// Script and CommonJS files do not have ES module scope.
			{Code: `define([], cb); require([], cb);`, FileName: "script.js", LanguageOptions: rule.LanguageOptions{SourceType: "script"}},
			{Code: `define([], cb); require([], cb);`, FileName: "common.cjs", LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"}},
			// Argument and callee shapes that upstream does not match.
			{Code: `define(); define([]); define([], cb, extra);`},
			{Code: `define(dependencies, cb); define(...[], cb); define({}, cb);`},
			{Code: `loader.define([], cb); loader["require"]([], cb); loader?.define([], cb);`},
			{Code: `(0, define)([], cb); (define || require)([], cb); new define([], cb);`},
			{Code: `(loader?.define)([], cb); (define?.())([], cb);`},
			// Lexical scopes matter, even outside functions.
			{Code: `{ define([], cb); } if (ready) { require([], cb); }`},
			{Code: `try { define([], cb); } catch (error) { require([], cb); } finally { define([], cb); }`},
			{Code: `switch (value) { case define([], cb): require([], cb); }`},
			{Code: `for (let value = define([], cb); ready; require([], cb)) {}`},
			{Code: `for (const value of require([], cb)) {}`},
			{Code: `while (ready) { define([], cb); }`},
			// Parameters, functions, classes and TypeScript scopes are nested.
			{Code: `function f(value = define([], cb)) { require([], cb); }`},
			{Code: `const f = () => define([], cb); const g = function named() { require([], cb); };`},
			{Code: `const object = { method(value = define([], cb)) { require([], cb); }, get value() { return define([], cb); } };`},
			{Code: `class C extends define([], cb) { [require([], cb)]() {} field = define([], cb); static { require([], cb); } }`},
			{Code: `namespace N { define([], cb); } enum E { A = require([], cb) }`},
			// Authored TypeScript wrappers remain visible in ESTree.
			{Code: `(define as any)([], cb); (require satisfies Function)([], cb); require!([], cb);`},
			{Code: `define([] as const, cb); require([] satisfies string[], cb); define(<string[]>[], cb);`},
			{Code: `const view = <loader.define value={() => require([], cb)} />;`, FileName: "view.tsx"},
			// getScope acquires the switch scope even for its discriminant.
			{Code: `switch (define([], cb)) { default: break; }`},
			{Code: `switch (require([], cb)) {}`},
			// Arrow defaults and decorators belong to their enclosing function/class scopes.
			{Code: `const factory = (x = define([], cb)) => require([], cb);`},
			{Code: `@define([], cb) class C { @require([], cb) method(@define([], cb) parameter: unknown) {} }`},
			// RunRuleTester registers the rule as "test" for inline directives.
			{Code: `/* eslint-disable test */
require([], cb);`},
		},
		[]rule_tester.InvalidTestCase{
			// Identifier spelling is decoded, but diagnostic columns follow the source text.
			{
				Code: `req\u0075ire([], cb); def\u0069ne([], cb);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: requireMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 21},
					{MessageId: "", Message: defineMessage, Line: 1, Column: 23, EndLine: 1, EndColumn: 42},
				},
			},
			// Destructuring keys and defaults evaluate in module scope.
			{
				Code: `const { [define([], cb)]: value } = object;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: defineMessage, Line: 1, Column: 10, EndLine: 1, EndColumn: 24},
				},
			},
			{
				Code: `let [value = require([], cb)] = [];`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: requireMessage, Line: 1, Column: 14, EndLine: 1, EndColumn: 29},
				},
			},
			// Accessor keys stay outside the getter/setter function scopes.
			{
				Code: `const object = { set [define([], cb)](value) {}, get [require([], cb)]() { return 1; } };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: defineMessage, Line: 1, Column: 23, EndLine: 1, EndColumn: 37},
					{MessageId: "", Message: requireMessage, Line: 1, Column: 55, EndLine: 1, EndColumn: 70},
				},
			},
			{
				Code: `/* eslint-disable-next-line test */
define([], cb);
require([], cb);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: requireMessage, Line: 3, Column: 1, EndLine: 3, EndColumn: 16},
				},
			},
			// Report only the inner AMD call in an optional member/call chain.
			{
				Code: `define?.([], cb)?.then(handler);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: defineMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 17},
				},
			},
			// The second argument is unrestricted, including a spread.
			{
				Code: `define([], cb);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: defineMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 15},
				},
			},
			{
				Code: `require([], 0);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: requireMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 15},
				},
			},
			{
				Code: `define([, ...modules], ...callbacks);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: defineMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 37},
				},
			},
			// Parentheses, optional calls and type arguments preserve direct calls.
			{
				Code: `(define)(([]), cb);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: defineMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 19},
				},
			},
			{
				Code: `((require([], cb)));`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: requireMessage, Line: 1, Column: 3, EndLine: 1, EndColumn: 18},
				},
			},
			{
				Code: `define?.([], cb);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: defineMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 17},
				},
			},
			{
				Code: `(require)?.([], cb);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: requireMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 20},
				},
			},
			{
				Code: `require<string>([], cb);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: requireMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 24},
				},
			},
			// Statements without a nested lexical scope still report.
			{
				Code: `if (ready) define([], cb);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: defineMessage, Line: 1, Column: 12, EndLine: 1, EndColumn: 26},
				},
			},
			{
				Code: `while (ready) require([], cb);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: requireMessage, Line: 1, Column: 15, EndLine: 1, EndColumn: 30},
				},
			},
			{
				Code: `for (var value = define([], cb); ready;) break;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: defineMessage, Line: 1, Column: 18, EndLine: 1, EndColumn: 32},
				},
			},
			{
				Code: `for (var value of require([], cb)) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: requireMessage, Line: 1, Column: 19, EndLine: 1, EndColumn: 34},
				},
			},
			// Computed object keys are outside the method scope.
			{
				Code: `const object = { [define([], cb)]() {}, value: require([], cb) };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: defineMessage, Line: 1, Column: 19, EndLine: 1, EndColumn: 33},
					{MessageId: "", Message: requireMessage, Line: 1, Column: 48, EndLine: 1, EndColumn: 63},
				},
			},
			// Module-scope bindings do not exempt these names.
			{
				Code: `const define = custom; define([], cb);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: defineMessage, Line: 1, Column: 24, EndLine: 1, EndColumn: 38},
				},
			},
			{
				Code: `import require from "loader"; require([], cb);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: requireMessage, Line: 1, Column: 31, EndLine: 1, EndColumn: 46},
				},
			},
			{
				Code: `declare function define(...args: any[]): void; define([], cb);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: defineMessage, Line: 1, Column: 48, EndLine: 1, EndColumn: 62},
				},
			},
			// Nested calls report in source order; function bodies stay excluded.
			{
				Code: `define([], () => { require([], cb); }); require([], cb);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: defineMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 39},
					{MessageId: "", Message: requireMessage, Line: 1, Column: 41, EndLine: 1, EndColumn: 56},
				},
			},
			{
				Code: `define([], require([], cb));`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: defineMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 28},
					{MessageId: "", Message: requireMessage, Line: 1, Column: 12, EndLine: 1, EndColumn: 27},
				},
			},
			// Complete UTF-16 ranges, JavaScript casts and JSX expressions.
			{
				Code: `/* 😀 */ define(
  ["依赖"],
  callback
);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: defineMessage, Line: 1, Column: 10, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code:     `/** @type {any} */ (define)(/** @type {any[]} */ ([]), cb);`,
				FileName: "cast.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: defineMessage, Line: 1, Column: 20, EndLine: 1, EndColumn: 59},
				},
			},
			{
				Code:     `const view = <div>{require([], cb)}</div>;`,
				FileName: "view.tsx",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: requireMessage, Line: 1, Column: 20, EndLine: 1, EndColumn: 35},
				},
			},
		},
	)
}
