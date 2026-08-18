package no_promise_executor_return

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestNoPromiseExecutorReturnUpstream migrates the full valid/invalid suite
// from upstream tests/lib/rules/no-promise-executor-return.js 1:1. Position
// assertions cover line/column for every invalid case. rslint-specific lock-in
// cases live in the no_promise_executor_return_extras_test.go file.
func TestNoPromiseExecutorReturnUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoPromiseExecutorReturnRule,
		[]rule_tester.ValidTestCase{
			// ---- General ----

			// not a promise executor
			{Code: `function foo(resolve, reject) { return 1; }`},
			{Code: `function Promise(resolve, reject) { return 1; }`},
			{Code: `(function (resolve, reject) { return 1; })`},
			{Code: `(function foo(resolve, reject) { return 1; })`},
			{Code: `(function Promise(resolve, reject) { return 1; })`},
			{Code: `var foo = function (resolve, reject) { return 1; }`},
			{Code: `var foo = function Promise(resolve, reject) { return 1; }`},
			{Code: `var Promise = function (resolve, reject) { return 1; }`},
			{Code: `(resolve, reject) => { return 1; }`},
			{Code: `(resolve, reject) => 1`},
			{Code: `var foo = (resolve, reject) => { return 1; }`},
			{Code: `var Promise = (resolve, reject) => { return 1; }`},
			{Code: `var foo = (resolve, reject) => 1`},
			{Code: `var Promise = (resolve, reject) => 1`},
			{Code: `var foo = { bar(resolve, reject) { return 1; } }`},
			{Code: `var foo = { Promise(resolve, reject) { return 1; } }`},
			{Code: `new foo(function (resolve, reject) { return 1; });`},
			{Code: `new foo(function bar(resolve, reject) { return 1; });`},
			{Code: `new foo(function Promise(resolve, reject) { return 1; });`},
			{Code: `new foo((resolve, reject) => { return 1; });`},
			{Code: `new foo((resolve, reject) => 1);`},
			{Code: `new promise(function foo(resolve, reject) { return 1; });`},
			{Code: `new Promise.foo(function foo(resolve, reject) { return 1; });`},
			{Code: `new foo.Promise(function foo(resolve, reject) { return 1; });`},
			{Code: `new Promise.Promise(function foo(resolve, reject) { return 1; });`},
			{Code: `new Promise()(function foo(resolve, reject) { return 1; });`},

			// not a promise executor - Promise() without new
			{Code: `Promise(function (resolve, reject) { return 1; });`},
			{Code: `Promise((resolve, reject) => { return 1; });`},
			{Code: `Promise((resolve, reject) => 1);`},

			// not a promise executor - not the first argument
			{Code: `new Promise(foo, function (resolve, reject) { return 1; });`},
			{Code: `new Promise(foo, (resolve, reject) => { return 1; });`},
			{Code: `new Promise(foo, (resolve, reject) => 1);`},

			// global Promise doesn't exist
			{Code: `/* globals Promise:off */ new Promise(function (resolve, reject) { return 1; });`},
			{
				Code:    `new Promise((resolve, reject) => { return 1; });`,
				Globals: map[string]any{"Promise": "off"},
			},

			// global Promise is shadowed
			{Code: `let Promise; new Promise(function (resolve, reject) { return 1; });`},
			{Code: `function f() { new Promise((resolve, reject) => { return 1; }); var Promise; }`},
			{Code: `function f(Promise) { new Promise((resolve, reject) => 1); }`},
			{Code: `if (x) { const Promise = foo(); new Promise(function (resolve, reject) { return 1; }); }`},
			{Code: `x = function Promise() { new Promise((resolve, reject) => { return 1; }); }`},

			// return without a value is allowed
			{Code: `new Promise(function (resolve, reject) { return; });`},
			{Code: `new Promise(function (resolve, reject) { reject(new Error()); return; });`},
			{Code: `new Promise(function (resolve, reject) { if (foo) { return; } });`},
			{Code: `new Promise((resolve, reject) => { return; });`},
			{Code: `new Promise((resolve, reject) => { if (foo) { resolve(1); return; } reject(new Error()); });`},

			// throw is allowed
			{Code: `new Promise(function (resolve, reject) { throw new Error(); });`},
			{Code: `new Promise((resolve, reject) => { throw new Error(); });`},

			// not returning from the promise executor
			{Code: `new Promise(function (resolve, reject) { function foo() { return 1; } });`},
			{Code: `new Promise((resolve, reject) => { (function foo() { return 1; })(); });`},
			{Code: `new Promise(function (resolve, reject) { () => { return 1; } });`},
			{Code: `new Promise((resolve, reject) => { () => 1 });`},
			{Code: `function foo() { return new Promise(function (resolve, reject) { resolve(bar); }) };`},
			{Code: `foo => new Promise((resolve, reject) => { bar(foo, (err, data) => { if (err) { reject(err); return; } resolve(data); })});`},

			// promise executors do not have effect on other functions (tests function info tracking)
			{Code: `new Promise(function (resolve, reject) {}); function foo() { return 1; }`},
			{Code: `new Promise((resolve, reject) => {}); (function () { return 1; });`},
			{Code: `new Promise(function (resolve, reject) {}); () => { return 1; };`},
			{Code: `new Promise((resolve, reject) => {}); () => 1;`},

			// does not report global return
			//
			// Upstream selects `sourceType: "commonjs"` / `ecmaFeatures.globalReturn`
			// to make these parse; rslint has no such switch and simply parses a
			// top-level return, so the behavior under test is still covered.
			{Code: `return 1;`},
			{Code: `return 1; function foo(){ return 1; } return 1;`},
			{Code: `function foo(){} return 1; var bar = function*(){ return 1; }; return 1; var baz = () => {}; return 1;`},
			{Code: `new Promise(function (resolve, reject) {}); return 1;`},

			// allowVoid: true — `=> void` and `return void` are allowed
			{
				Code:    `new Promise((r) => void cbf(r));`,
				Options: map[string]any{"allowVoid": true},
			},
			{
				Code:    `new Promise(r => void 0)`,
				Options: map[string]any{"allowVoid": true},
			},
			{
				Code:    `new Promise(r => { return void 0 })`,
				Options: map[string]any{"allowVoid": true},
			},
			{
				Code:    `new Promise(r => { if (foo) { return void 0 } return void 0 })`,
				Options: map[string]any{"allowVoid": true},
			},
			{Code: `new Promise(r => {0})`},
		},
		[]rule_tester.InvalidTestCase{

			// ---- full error tests ----
			{
				Code: `new Promise(function (resolve, reject) { return 1; })`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "returnsValue",
					Message:   "Return values from promise executor functions cannot be read.",
					Line:      1, Column: 42, EndLine: 1, EndColumn: 51,
				}},
			},
			{
				Code:    `new Promise((resolve, reject) => resolve(1))`,
				Options: map[string]any{"allowVoid": true},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "returnsValue",
					Message:   "Return values from promise executor functions cannot be read.",
					Line:      1, Column: 34, EndLine: 1, EndColumn: 44,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "prependVoid", Output: `new Promise((resolve, reject) => void resolve(1))`},
						{MessageId: "wrapBraces", Output: `new Promise((resolve, reject) => {resolve(1)})`},
					},
				}},
			},
			{
				Code:    `new Promise((resolve, reject) => { return 1 })`,
				Options: map[string]any{"allowVoid": true},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "returnsValue",
					Line:      1, Column: 36, EndLine: 1, EndColumn: 44,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "prependVoid", Output: `new Promise((resolve, reject) => { return void 1 })`},
					},
				}},
			},

			// ---- suggestions arrow function expression ----
			{
				Code:    `new Promise(r => 1)`,
				Options: map[string]any{"allowVoid": true},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "returnsValue",
					Line:      1, Column: 18, EndLine: 1, EndColumn: 19,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "prependVoid", Output: `new Promise(r => void 1)`},
						{MessageId: "wrapBraces", Output: `new Promise(r => {1})`},
					},
				}},
			},
			{
				Code:    `new Promise(r => 1 ? 2 : 3)`,
				Options: map[string]any{"allowVoid": true},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "returnsValue",
					Line:      1, Column: 18, EndLine: 1, EndColumn: 27,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "prependVoid", Output: `new Promise(r => void (1 ? 2 : 3))`},
						{MessageId: "wrapBraces", Output: `new Promise(r => {1 ? 2 : 3})`},
					},
				}},
			},
			{
				Code:    `new Promise(r => (1 ? 2 : 3))`,
				Options: map[string]any{"allowVoid": true},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "returnsValue",
					Line:      1, Column: 19, EndLine: 1, EndColumn: 28,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "prependVoid", Output: `new Promise(r => void (1 ? 2 : 3))`},
						{MessageId: "wrapBraces", Output: `new Promise(r => {(1 ? 2 : 3)})`},
					},
				}},
			},
			{
				Code:    `new Promise(r => (1))`,
				Options: map[string]any{"allowVoid": true},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "returnsValue",
					Line:      1, Column: 19, EndLine: 1, EndColumn: 20,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "prependVoid", Output: `new Promise(r => void (1))`},
						{MessageId: "wrapBraces", Output: `new Promise(r => {(1)})`},
					},
				}},
			},
			{
				Code:    `new Promise(r => () => {})`,
				Options: map[string]any{"allowVoid": true},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "returnsValue",
					Line:      1, Column: 18, EndLine: 1, EndColumn: 26,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "prependVoid", Output: `new Promise(r => void (() => {}))`},
						{MessageId: "wrapBraces", Output: `new Promise(r => {() => {}})`},
					},
				}},
			},

			// ---- primitives ----
			{
				Code:    `new Promise(r => null)`,
				Options: map[string]any{"allowVoid": true},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "returnsValue",
					Line:      1, Column: 18, EndLine: 1, EndColumn: 22,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "prependVoid", Output: `new Promise(r => void null)`},
						{MessageId: "wrapBraces", Output: `new Promise(r => {null})`},
					},
				}},
			},
			{
				Code:    `new Promise(r => null)`,
				Options: map[string]any{"allowVoid": false},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "returnsValue",
					Line:      1, Column: 18, EndLine: 1, EndColumn: 22,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "wrapBraces", Output: `new Promise(r => {null})`},
					},
				}},
			},

			// ---- inline comments ----
			{
				Code:    `new Promise(r => /*hi*/ ~0)`,
				Options: map[string]any{"allowVoid": true},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "returnsValue",
					Line:      1, Column: 25, EndLine: 1, EndColumn: 27,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "prependVoid", Output: `new Promise(r => /*hi*/ void ~0)`},
						{MessageId: "wrapBraces", Output: `new Promise(r => /*hi*/ {~0})`},
					},
				}},
			},
			{
				Code:    `new Promise(r => /*hi*/ ~0)`,
				Options: map[string]any{"allowVoid": false},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "returnsValue",
					Line:      1, Column: 25, EndLine: 1, EndColumn: 27,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "wrapBraces", Output: `new Promise(r => /*hi*/ {~0})`},
					},
				}},
			},

			// ---- suggestions function ----
			{
				Code:    `new Promise(r => { return 0 })`,
				Options: map[string]any{"allowVoid": true},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "returnsValue",
					Line:      1, Column: 20, EndLine: 1, EndColumn: 28,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "prependVoid", Output: `new Promise(r => { return void 0 })`},
					},
				}},
			},
			{
				Code:    `new Promise(r => { return 0 })`,
				Options: map[string]any{"allowVoid": false},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "returnsValue",
					Line:      1, Column: 20, EndLine: 1, EndColumn: 28,
				}},
			},

			// ---- multiple returns ----
			{
				Code:    `new Promise(r => { if (foo) { return void 0 } return 0 })`,
				Options: map[string]any{"allowVoid": true},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "returnsValue",
					Line:      1, Column: 47, EndLine: 1, EndColumn: 55,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "prependVoid", Output: `new Promise(r => { if (foo) { return void 0 } return void 0 })`},
					},
				}},
			},

			// ---- return assignment ----
			{
				Code:    `new Promise(resolve => { return (foo = resolve(1)); })`,
				Options: map[string]any{"allowVoid": true},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "returnsValue",
					Line:      1, Column: 26, EndLine: 1, EndColumn: 52,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "prependVoid", Output: `new Promise(resolve => { return void (foo = resolve(1)); })`},
					},
				}},
			},
			{
				Code:    `new Promise(resolve => r = resolve)`,
				Options: map[string]any{"allowVoid": true},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "returnsValue",
					Line:      1, Column: 24, EndLine: 1, EndColumn: 35,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "prependVoid", Output: `new Promise(resolve => void (r = resolve))`},
						{MessageId: "wrapBraces", Output: `new Promise(resolve => {r = resolve})`},
					},
				}},
			},

			// ---- return<immediate token> (range check) ----
			{
				Code:    `new Promise(r => { return(1) })`,
				Options: map[string]any{"allowVoid": true},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "returnsValue",
					Line:      1, Column: 20, EndLine: 1, EndColumn: 29,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "prependVoid", Output: `new Promise(r => { return void (1) })`},
					},
				}},
			},
			{
				Code:    `new Promise(r =>1)`,
				Options: map[string]any{"allowVoid": true},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "returnsValue",
					Line:      1, Column: 17, EndLine: 1, EndColumn: 18,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "prependVoid", Output: `new Promise(r =>void 1)`},
						{MessageId: "wrapBraces", Output: `new Promise(r =>{1})`},
					},
				}},
			},

			// ---- snapshot ----
			{
				Code:    `new Promise(r => ((1)))`,
				Options: map[string]any{"allowVoid": true},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "returnsValue",
					Line:      1, Column: 20, EndLine: 1, EndColumn: 21,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "prependVoid", Output: `new Promise(r => void ((1)))`},
						{MessageId: "wrapBraces", Output: `new Promise(r => {((1))})`},
					},
				}},
			},

			// ---- other basic tests ----
			{
				Code: `new Promise(function foo(resolve, reject) { return 1; })`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "returnsValue",
					Line:      1, Column: 45, EndLine: 1, EndColumn: 54,
				}},
			},
			{
				Code: `new Promise((resolve, reject) => { return 1; })`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "returnsValue",
					Line:      1, Column: 36, EndLine: 1, EndColumn: 45,
				}},
			},

			// ---- any returned value ----
			{
				Code: `new Promise(function (resolve, reject) { return undefined; })`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "returnsValue",
					Line:      1, Column: 42, EndLine: 1, EndColumn: 59,
				}},
			},
			{
				Code: `new Promise((resolve, reject) => { return null; })`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "returnsValue",
					Line:      1, Column: 36, EndLine: 1, EndColumn: 48,
				}},
			},
			{
				Code: `new Promise(function (resolve, reject) { return false; })`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "returnsValue",
					Line:      1, Column: 42, EndLine: 1, EndColumn: 55,
				}},
			},
			{
				Code: `new Promise((resolve, reject) => resolve)`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "returnsValue",
					Line:      1, Column: 34, EndLine: 1, EndColumn: 41,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "wrapBraces", Output: `new Promise((resolve, reject) => {resolve})`},
					},
				}},
			},
			{
				Code: `new Promise((resolve, reject) => null)`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "returnsValue",
					Line:      1, Column: 34, EndLine: 1, EndColumn: 38,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "wrapBraces", Output: `new Promise((resolve, reject) => {null})`},
					},
				}},
			},
			{
				Code: `new Promise(function (resolve, reject) { return resolve(foo); })`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "returnsValue",
					Line:      1, Column: 42, EndLine: 1, EndColumn: 62,
				}},
			},
			{
				Code: `new Promise((resolve, reject) => { return reject(foo); })`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "returnsValue",
					Line:      1, Column: 36, EndLine: 1, EndColumn: 55,
				}},
			},
			{
				Code: `new Promise((resolve, reject) => x + y)`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "returnsValue",
					Line:      1, Column: 34, EndLine: 1, EndColumn: 39,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "wrapBraces", Output: `new Promise((resolve, reject) => {x + y})`},
					},
				}},
			},
			{
				Code: `new Promise((resolve, reject) => { return Promise.resolve(42); })`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "returnsValue",
					Line:      1, Column: 36, EndLine: 1, EndColumn: 63,
				}},
			},

			// ---- any return statement location ----
			{
				Code: `new Promise(function (resolve, reject) { if (foo) { return 1; } })`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "returnsValue",
					Line:      1, Column: 53, EndLine: 1, EndColumn: 62,
				}},
			},
			{
				Code: `new Promise((resolve, reject) => { try { return 1; } catch(e) {} })`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "returnsValue",
					Line:      1, Column: 42, EndLine: 1, EndColumn: 51,
				}},
			},
			{
				Code: `new Promise(function (resolve, reject) { while (foo){ if (bar) break; else return 1; } })`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "returnsValue",
					Line:      1, Column: 76, EndLine: 1, EndColumn: 85,
				}},
			},

			// ---- `return void` is not allowed without `allowVoid: true` ----
			{
				Code: `new Promise(() => { return void 1; })`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "returnsValue",
					Line:      1, Column: 21, EndLine: 1, EndColumn: 35,
				}},
			},
			{
				Code: `new Promise(() => (1))`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "returnsValue",
					Line:      1, Column: 20, EndLine: 1, EndColumn: 21,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "wrapBraces", Output: `new Promise(() => {(1)})`},
					},
				}},
			},
			{
				Code: `() => new Promise(() => ({}));`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "returnsValue",
					Line:      1, Column: 26, EndLine: 1, EndColumn: 28,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "wrapBraces", Output: `() => new Promise(() => {({})});`},
					},
				}},
			},

			// ---- absence of arguments has no effect ----
			{
				Code: `new Promise(function () { return 1; })`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "returnsValue",
					Line:      1, Column: 27, EndLine: 1, EndColumn: 36,
				}},
			},
			{
				Code: `new Promise(() => { return 1; })`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "returnsValue",
					Line:      1, Column: 21, EndLine: 1, EndColumn: 30,
				}},
			},
			{
				Code: `new Promise(() => 1)`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "returnsValue",
					Line:      1, Column: 19, EndLine: 1, EndColumn: 20,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "wrapBraces", Output: `new Promise(() => {1})`},
					},
				}},
			},

			// ---- various scope tracking tests ----
			{
				Code: `function foo() {} new Promise(function () { return 1; });`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "returnsValue",
					Line:      1, Column: 45, EndLine: 1, EndColumn: 54,
				}},
			},
			{
				Code: `function foo() { return; } new Promise(() => { return 1; });`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "returnsValue",
					Line:      1, Column: 48, EndLine: 1, EndColumn: 57,
				}},
			},
			{
				Code: `function foo() { return 1; } new Promise(() => { return 2; });`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "returnsValue",
					Line:      1, Column: 50, EndLine: 1, EndColumn: 59,
				}},
			},
			{
				Code: `function foo () { return new Promise(function () { return 1; }); }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "returnsValue",
					Line:      1, Column: 52, EndLine: 1, EndColumn: 61,
				}},
			},
			{
				Code: `function foo() { return new Promise(() => { bar(() => { return 1; }); return false; }); }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "returnsValue",
					Line:      1, Column: 71, EndLine: 1, EndColumn: 84,
				}},
			},
			{
				Code: `() => new Promise(() => { if (foo) { return 0; } else bar(() => { return 1; }); })`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "returnsValue",
					Line:      1, Column: 38, EndLine: 1, EndColumn: 47,
				}},
			},
			{
				Code: `function foo () { return 1; return new Promise(function () { return 2; }); return 3;}`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "returnsValue",
					Line:      1, Column: 62, EndLine: 1, EndColumn: 71,
				}},
			},
			{
				Code: `() => 1; new Promise(() => { return 1; })`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "returnsValue",
					Line:      1, Column: 30, EndLine: 1, EndColumn: 39,
				}},
			},
			{
				Code: `new Promise(function () { return 1; }); function foo() { return 1; } `,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "returnsValue",
					Line:      1, Column: 27, EndLine: 1, EndColumn: 36,
				}},
			},
			{
				Code: `() => new Promise(() => { return 1; });`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "returnsValue",
					Line:      1, Column: 27, EndLine: 1, EndColumn: 36,
				}},
			},
			{
				Code: `() => new Promise(() => 1);`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "returnsValue",
					Line:      1, Column: 25, EndLine: 1, EndColumn: 26,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "wrapBraces", Output: `() => new Promise(() => {1});`},
					},
				}},
			},
			{
				Code: `() => new Promise(() => () => 1);`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "returnsValue",
					Line:      1, Column: 25, EndLine: 1, EndColumn: 32,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "wrapBraces", Output: `() => new Promise(() => {() => 1});`},
					},
				}},
			},
			{
				Code: `() => new Promise(() => async () => 1);`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "returnsValue",
					Line:      1, Column: 25, EndLine: 1, EndColumn: 38,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "wrapBraces", Output: `() => new Promise(() => {async () => 1});`},
					},
				}},
			},
			{
				Code: `() => new Promise(() => function () {});`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "returnsValue",
					Line:      1, Column: 25, EndLine: 1, EndColumn: 39,
				}},
			},
			{
				// No suggestion since an unnamed FunctionExpression inside braces is invalid syntax.
				Code: `() => new Promise(() => class {});`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "returnsValue",
					Line:      1, Column: 25, EndLine: 1, EndColumn: 33,
				}},
			},
			{
				// No suggestion since an unnamed ClassExpression inside braces is invalid syntax.
				Code: `() => new Promise(() => function foo() {});`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "returnsValue",
					Line:      1, Column: 25, EndLine: 1, EndColumn: 42,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "wrapBraces", Output: `() => new Promise(() => {function foo() {}});`},
					},
				}},
			},
			{
				Code: `() => new Promise(() => class Foo {});`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "returnsValue",
					Line:      1, Column: 25, EndLine: 1, EndColumn: 37,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "wrapBraces", Output: `() => new Promise(() => {class Foo {}});`},
					},
				}},
			},
			{
				Code: `() => new Promise(() => []);`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "returnsValue",
					Line:      1, Column: 25, EndLine: 1, EndColumn: 27,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "wrapBraces", Output: `() => new Promise(() => {[]});`},
					},
				}},
			},

			// ---- edge cases for global Promise reference ----
			{
				Code: `new Promise((Promise) => { return 1; })`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "returnsValue",
					Line:      1, Column: 28, EndLine: 1, EndColumn: 37,
				}},
			},
			{
				Code: `new Promise(function Promise(resolve, reject) { return 1; })`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "returnsValue",
					Line:      1, Column: 49, EndLine: 1, EndColumn: 58,
				}},
			},
		},
	)
}
