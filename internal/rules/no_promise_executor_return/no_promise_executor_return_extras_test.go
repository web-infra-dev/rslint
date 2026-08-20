package no_promise_executor_return

import (
	"reflect"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// TestNoPromiseExecutorReturnExtras locks in branches and edge shapes that the
// upstream test suite doesn't exercise. Each case carries an inline comment
// pointing at the specific branch / Dimension 4 row / tsgo AST quirk it covers,
// so future refactors can't silently regress them without breaking a named
// lock-in.
//
// N/A Dimension 4 rows:
//   - Access / key forms (identifier vs string / numeric / private / computed
//     keys): the rule never inspects object-literal or class members — the only
//     name it reads is a bare `Promise` callee identifier.
//   - Autofix boundaries (Dimension 3): the rule emits suggestions only, never
//     an autofix. Suggestion text is asserted on every case below, and
//     TestNoPromiseExecutorReturnEditDemand covers the demand boundary.
//   - Overload signatures / `abstract` / `declare` members: a body-less member
//     holds neither a concise arrow body nor a `return` statement.
func TestNoPromiseExecutorReturnExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoPromiseExecutorReturnRule,
		[]rule_tester.ValidTestCase{
			// Locks in upstream ReturnStatement() arm 1 under an explicit
			// allowVoid=false: a value-less `return` is a control-flow statement
			// whatever the option says.
			{
				Code:    `new Promise(r => { return; })`,
				Options: map[string]any{"allowVoid": false},
			},
			// ---- Dimension 4: TS non-null assertion on the receiver ----
			{
				Code:    `new (Promise!)(r => 1)`,
				Options: map[string]any{"allowVoid": true},
			},
			// ---- Dimension 4: TS type-expression wrapper on the receiver ----
			{
				Code:    `new (Promise as any)(r => 1)`,
				Options: map[string]any{"allowVoid": true},
			},
			// ---- Dimension 4: element-access receiver is never the global Promise ----
			{
				Code:    `new x['Promise'](r => 1)`,
				Options: map[string]any{"allowVoid": true},
			},
			// ---- Dimension 4: element access off Promise is not the constructor ----
			{
				Code:    `new Promise['bind'](r => 1)`,
				Options: map[string]any{"allowVoid": true},
			},
			// ---- Dimension 4: object method boundary owns its own return ----
			{
				Code: `new Promise(function () { ({ m() { return 1; } }); });`,
			},
			// ---- Dimension 4: getter boundary owns its own return ----
			{
				Code: `new Promise(r => { const o = { get x() { return 1; } }; });`,
			},
			// ---- Dimension 4: class method boundary owns its own return ----
			{
				Code: `new Promise(r => { class A { static m() { return 1; } } });`,
			},
			// ---- Dimension 4: parameter-default function owns its own return ----
			{
				Code: `new Promise(function (a = function () { return 1; }) {});`,
			},
			// ---- Dimension 4: spread argument shifts the executor out of position 0 ----
			{
				Code: `new Promise(...args, function () { return 1; });`,
			},
			// ---- Dimension 4: arrow nested inside a spread array literal is not an argument ----
			{
				Code: `new Promise(...[() => 1]);`,
			},
			// ---- Dimension 4: empty argument list ----
			{
				Code: `new Promise();`,
			},
			// ---- Dimension 4: empty function body ----
			{
				Code: `new Promise(function () {});`,
			},
			// ---- Dimension 4: empty arrow block body ----
			{
				Code: `new Promise(() => {});`,
			},
			// Locks in upstream isPromiseExecutor() arm 1: parent is not a NewExpression.
			{
				Code: `[function () { return 1; }];`,
			},
			// Locks in upstream isPromiseExecutor() arm 3: callee is not an Identifier.
			{
				Code: `new (foo())(r => 1)`,
			},
			// Locks in upstream isPromiseExecutor() arm 5: an enum declaration shadows the global.
			{
				Code: `enum Promise { a }
new Promise(r => 1)`,
			},
			// Locks in upstream isPromiseExecutor() arm 5: a namespace declaration shadows the global.
			{
				Code: `namespace Promise { }
new Promise(r => 1)`,
			},
			// Locks in upstream isPromiseExecutor() arm 5: a class declaration shadows the global.
			{
				Code: `class Promise {}
new Promise(r => 1)`,
			},
			// Locks in upstream isPromiseExecutor() arm 5: a named import shadows the global.
			{
				Code: `import { Promise } from "x";
new Promise(r => 1)`,
			},
			// Locks in upstream isPromiseExecutor() arm 5: `declare const` counts as a definition.
			{
				Code: `declare const Promise: any;
new Promise(r => 1)`,
			},
			// Locks in upstream isPromiseExecutor() arm 5: at file scope ESLint holds
			// the source declarations and the configured global in one variable, so an
			// `interface` definition clears the global reference too.
			{
				Code: `interface Promise {}
new Promise(r => 1)`,
			},
			// Locks in upstream isPromiseExecutor() arm 5: same for a type alias.
			{
				Code: `type Promise = any;
new Promise(r => 1)`,
			},
			// Locks in upstream isPromiseExecutor() arm 5: a type-only import binds the
			// name, so the callee is no longer the global.
			{
				Code: `import type { Promise } from "x";
new Promise(r => 1)`,
			},
			// Locks in upstream isPromiseExecutor() arm 5: a `var` in an enclosing
			// namespace body shadows the global.
			{
				Code: `namespace N { var Promise: any; new Promise(r => 1); }`,
			},
			// Locks in upstream isPromiseExecutor() arm 5: so does a class declared in
			// that namespace body.
			{
				Code: `namespace N { class Promise {} new Promise(r => 1); }`,
			},
			// Locks in upstream isPromiseExecutor() arm 5: a namespace body binding is
			// visible from a function nested inside it.
			{
				Code: `namespace N { var Promise: any; function g() { new Promise(r => 1); } }`,
			},
			// Locks in upstream isPromiseExecutor() arm 5: a JSDoc `@typedef` is only a
			// comment to ESLint, but the real `const` beside it still defines the name.
			{
				Code: `/** @typedef {number} Promise */
const Promise = globalThis.Promise;
new Promise(r => 1)`,
				FileName: "promise-jsdoc-and-const.js",
				TSConfig: "tsconfig.allow-js.json",
			},
			// Locks in upstream isPromiseExecutor() arm 5: a real import declares the
			// name whether or not a JSDoc `@import` tag spells it first.
			{
				Code: `/** @import { Promise } from "x" */
import { Promise } from "y";
new Promise(r => 1)`,
				FileName: "promise-jsdoc-and-import.js",
				TSConfig: "tsconfig.allow-js.json",
			},
			{
				Code: `import { Promise } from "y";
/** @import { Promise } from "x" */
new Promise(r => 1)`,
				FileName: "promise-import-and-jsdoc.js",
				TSConfig: "tsconfig.allow-js.json",
			},
			// Locks in upstream isPromiseExecutor() arm 5: a parameter initializer sees
			// the parameters declared beside it, so the callee is not the global.
			{
				Code:     `function f(Promise, x = new Promise(r => 1)) {}`,
				FileName: "promise-sibling-parameter.js",
				TSConfig: "tsconfig.allow-js.json",
			},
			{
				Code: `function f(Promise: any, x = new Promise(r => 1)) {}`,
			},
			// Locks in upstream isPromiseExecutor() arm 5: skipping the body scope a
			// parameter initializer cannot see must continue outward, not jump to the
			// global — the enclosing function's parameter still binds the name.
			{
				Code:     `function outer(Promise) { function f(x = new Promise(r => 1)) { function Promise() {} } }`,
				FileName: "promise-outer-parameter.js",
				TSConfig: "tsconfig.allow-js.json",
			},
			{
				Code: `function outer(Promise: any) { function f(x = new Promise(r => 1)) { function Promise() {} } }`,
			},
			// Locks in upstream isPromiseExecutor() arm 5: a named function expression
			// binds its own name outside the body, so the parameter list still sees it.
			{
				Code:     `const f = function Promise(x = new Promise(r => 1)) { function Promise() {} };`,
				FileName: "promise-function-expression-name.js",
				TSConfig: "tsconfig.allow-js.json",
			},
			{
				Code: `const f = function Promise(x = new Promise(r => 1)) { function Promise() {} };`,
			},
			// Locks in the ctx.Globals gate on a `.js` file, where the JSDoc and
			// parameter-scope paths above live.
			{
				Code: `/* globals Promise:off */
new Promise(r => 1)`,
				FileName: "promise-globals-off.js",
				TSConfig: "tsconfig.allow-js.json",
			},
			// Locks in upstream onCodePathStart() arm 3: parentheses around `void` do not
			// defeat the allowVoid exemption.
			{
				Code:    `new Promise(r => (void 0))`,
				Options: map[string]any{"allowVoid": true},
			},
			// Locks in upstream ReturnStatement() arm 3: parentheses around `void` do not
			// defeat the allowVoid exemption.
			{
				Code:    `new Promise(r => { return (void 0); })`,
				Options: map[string]any{"allowVoid": true},
			},
			// Locks in upstream onCodePathStart() arm 2: a block body is not an implicit return.
			{
				Code: `new Promise(r => {})`,
			},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: parenthesized receiver — tsgo keeps `(Promise)` as a node, ESTree flattens it ----
			{
				Code:    `new (Promise)(r => 1)`,
				Options: map[string]any{"allowVoid": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "returnsValue",
						Line:      1, Column: 20, EndLine: 1, EndColumn: 21,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "prependVoid", Output: `new (Promise)(r => void 1)`},
							{MessageId: "wrapBraces", Output: `new (Promise)(r => {1})`},
						},
					},
				},
			},
			// ---- Dimension 4: multi-level parenthesized receiver ----
			{
				Code:    `new ((Promise))(r => 1)`,
				Options: map[string]any{"allowVoid": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "returnsValue",
						Line:      1, Column: 22, EndLine: 1, EndColumn: 23,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "prependVoid", Output: `new ((Promise))(r => void 1)`},
							{MessageId: "wrapBraces", Output: `new ((Promise))(r => {1})`},
						},
					},
				},
			},
			// ---- Dimension 4: parenthesized executor argument ----
			{
				Code:    `new Promise((r => 1))`,
				Options: map[string]any{"allowVoid": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "returnsValue",
						Line:      1, Column: 19, EndLine: 1, EndColumn: 20,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "prependVoid", Output: `new Promise((r => void 1))`},
							{MessageId: "wrapBraces", Output: `new Promise((r => {1}))`},
						},
					},
				},
			},
			// ---- Dimension 4: multi-level parenthesized executor argument ----
			{
				Code: `new Promise(((function () { return 1; })))`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "returnsValue",
						Line:      1, Column: 29, EndLine: 1, EndColumn: 38,
					},
				},
			},
			// ---- Dimension 4: TS type-expression wrapper on the reported body ----
			{
				Code:    `new Promise(r => x as any)`,
				Options: map[string]any{"allowVoid": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "returnsValue",
						Line:      1, Column: 18, EndLine: 1, EndColumn: 26,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "prependVoid", Output: `new Promise(r => void (x as any))`},
							{MessageId: "wrapBraces", Output: `new Promise(r => {x as any})`},
						},
					},
				},
			},
			// ---- Dimension 4: TS non-null assertion as the reported body ----
			{
				Code:    `new Promise(r => x!)`,
				Options: map[string]any{"allowVoid": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "returnsValue",
						Line:      1, Column: 18, EndLine: 1, EndColumn: 20,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "prependVoid", Output: `new Promise(r => void (x!))`},
							{MessageId: "wrapBraces", Output: `new Promise(r => {x!})`},
						},
					},
				},
			},
			// ---- Dimension 4: TS `satisfies` as the reported body ----
			{
				Code:    `new Promise(r => x satisfies number)`,
				Options: map[string]any{"allowVoid": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "returnsValue",
						Line:      1, Column: 18, EndLine: 1, EndColumn: 36,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "prependVoid", Output: `new Promise(r => void (x satisfies number))`},
							{MessageId: "wrapBraces", Output: `new Promise(r => {x satisfies number})`},
						},
					},
				},
			},
			// ---- Dimension 4: optional chain body — flag-based in tsgo, no ChainExpression wrapper ----
			{
				Code:    `new Promise(r => a?.b)`,
				Options: map[string]any{"allowVoid": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "returnsValue",
						Line:      1, Column: 18, EndLine: 1, EndColumn: 22,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "prependVoid", Output: `new Promise(r => void a?.b)`},
							{MessageId: "wrapBraces", Output: `new Promise(r => {a?.b})`},
						},
					},
				},
			},
			// ---- Dimension 4: optional call body ----
			{
				Code:    `new Promise(r => a?.())`,
				Options: map[string]any{"allowVoid": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "returnsValue",
						Line:      1, Column: 18, EndLine: 1, EndColumn: 23,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "prependVoid", Output: `new Promise(r => void a?.())`},
							{MessageId: "wrapBraces", Output: `new Promise(r => {a?.()})`},
						},
					},
				},
			},
			// ---- Dimension 4: generator function expression executor ----
			{
				Code: `new Promise(function* () { return 1; })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "returnsValue",
						Line:      1, Column: 28, EndLine: 1, EndColumn: 37,
					},
				},
			},
			// ---- Dimension 4: async function expression executor ----
			{
				Code: `new Promise(async function () { return 1; })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "returnsValue",
						Line:      1, Column: 33, EndLine: 1, EndColumn: 42,
					},
				},
			},
			// ---- Dimension 4: async arrow executor with a concise body ----
			{
				Code: `new Promise(async () => 1)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "returnsValue",
						Line:      1, Column: 25, EndLine: 1, EndColumn: 26,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "wrapBraces", Output: `new Promise(async () => {1})`},
						},
					},
				},
			},
			// ---- Dimension 4: async generator function expression executor ----
			{
				Code: `new Promise(async function* () { return 1; })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "returnsValue",
						Line:      1, Column: 34, EndLine: 1, EndColumn: 43,
					},
				},
			},
			// ---- Dimension 4: class-field initializer container ----
			{
				Code: `class A { f = new Promise(r => 1); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "returnsValue",
						Line:      1, Column: 32, EndLine: 1, EndColumn: 33,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "wrapBraces", Output: `class A { f = new Promise(r => {1}); }`},
						},
					},
				},
			},
			// ---- Dimension 4: parenthesized unnamed function body still suppresses wrapBraces ----
			{
				Code:    `new Promise(r => (function () {}))`,
				Options: map[string]any{"allowVoid": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "returnsValue",
						Line:      1, Column: 19, EndLine: 1, EndColumn: 33,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "prependVoid", Output: `new Promise(r => void (function () {}))`},
						},
					},
				},
			},
			// ---- Dimension 4: parenthesized unnamed class body still suppresses wrapBraces ----
			{
				Code:    `new Promise(r => (class {}))`,
				Options: map[string]any{"allowVoid": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "returnsValue",
						Line:      1, Column: 19, EndLine: 1, EndColumn: 27,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "prependVoid", Output: `new Promise(r => void (class {}))`},
						},
					},
				},
			},
			// ---- Dimension 4: same-kind nesting — both executors report ----
			{
				Code: `new Promise(() => new Promise(() => 1));`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "returnsValue",
						Line:      1, Column: 19, EndLine: 1, EndColumn: 39,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "wrapBraces", Output: `new Promise(() => {new Promise(() => 1)});`},
						},
					},
					{
						MessageId: "returnsValue",
						Line:      1, Column: 37, EndLine: 1, EndColumn: 38,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "wrapBraces", Output: `new Promise(() => new Promise(() => {1}));`},
						},
					},
				},
			},
			// ---- Dimension 4: class static block container ----
			{
				Code: `class A { static { new Promise(r => 1); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "returnsValue",
						Line:      1, Column: 37, EndLine: 1, EndColumn: 38,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "wrapBraces", Output: `class A { static { new Promise(r => {1}); } }`},
						},
					},
				},
			},
			// ---- Dimension 4: an empty static block does not break out of the executor ----
			{
				Code: `new Promise(r => { class A { static { } } return 1; })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "returnsValue",
						Line:      1, Column: 43, EndLine: 1, EndColumn: 52,
					},
				},
			},
			// ---- Real-user: eslint#13668 one-liner delay executor ----
			{
				Code:    `new Promise(resolve => setTimeout(resolve, 1000));`,
				Options: map[string]any{"allowVoid": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "returnsValue",
						Line:      1, Column: 24, EndLine: 1, EndColumn: 49,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "prependVoid", Output: `new Promise(resolve => void setTimeout(resolve, 1000));`},
							{MessageId: "wrapBraces", Output: `new Promise(resolve => {setTimeout(resolve, 1000)});`},
						},
					},
				},
			},
			// ---- Real-user: eslint#16123 `return resolve(...)` early exit ----
			{
				Code:    `const p = new Promise(function (resolve) { if (a) return resolve(1); resolve(2); });`,
				Options: map[string]any{"allowVoid": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "returnsValue",
						Line:      1, Column: 51, EndLine: 1, EndColumn: 69,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "prependVoid", Output: `const p = new Promise(function (resolve) { if (a) return void resolve(1); resolve(2); });`},
						},
					},
				},
			},
			// Locks in upstream isPromiseExecutor() arm 5: inside a namespace a
			// type-only declaration leaves the reference global, unlike at file scope.
			{
				Code: `namespace N { interface Promise {} new Promise(r => 1); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "returnsValue",
						Line:      1, Column: 53, EndLine: 1, EndColumn: 54,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "wrapBraces", Output: `namespace N { interface Promise {} new Promise(r => {1}); }`},
						},
					},
				},
			},
			// Locks in upstream isPromiseExecutor() arm 5: a namespace that declares
			// no `Promise` of its own still reaches the global.
			{
				Code: `namespace N { new Promise(r => 1); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "returnsValue",
						Line:      1, Column: 32, EndLine: 1, EndColumn: 33,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "wrapBraces", Output: `namespace N { new Promise(r => {1}); }`},
						},
					},
				},
			},
			// Locks in upstream isPromiseExecutor() arm 5: a JSDoc `@typedef` is a
			// comment, so it declares nothing ESLint can scope the callee to.
			{
				Code: `/** @typedef {number} Promise */
new Promise(r => 1)`,
				FileName: "promise-jsdoc-typedef.js",
				TSConfig: "tsconfig.allow-js.json",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "returnsValue",
						Line:      2, Column: 18, EndLine: 2, EndColumn: 19,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "wrapBraces", Output: `/** @typedef {number} Promise */
new Promise(r => {1})`},
						},
					},
				},
			},
			// Locks in upstream isPromiseExecutor() arm 5: a JSDoc `@import` tag binds
			// nothing ESLint can see either, in a module file as much as a script.
			{
				Code: `/** @import { Promise } from "x" */
export {};
new Promise(r => 1)`,
				FileName: "promise-jsdoc-import.js",
				TSConfig: "tsconfig.allow-js.json",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "returnsValue",
						Line:      3, Column: 18, EndLine: 3, EndColumn: 19,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "wrapBraces", Output: `/** @import { Promise } from "x" */
export {};
new Promise(r => {1})`},
						},
					},
				},
			},
			// Locks in upstream isPromiseExecutor() arm 5: the tag stays invisible when
			// it is written inside a function, where the binder hoists it to file scope.
			{
				Code: `export {};
function g() {
  /** @import { Promise } from "x" */
  return new Promise(r => 1);
}`,
				FileName: "promise-jsdoc-import-nested.js",
				TSConfig: "tsconfig.allow-js.json",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "returnsValue",
						Line:      4, Column: 27, EndLine: 4, EndColumn: 28,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "wrapBraces", Output: `export {};
function g() {
  /** @import { Promise } from "x" */
  return new Promise(r => {1});
}`},
						},
					},
				},
			},
			// Locks in upstream isPromiseExecutor() arm 5: a parameter initializer is
			// evaluated in a scope outside the body, so a function declared there is
			// invisible and the callee is still the global.
			{
				Code:     `function f(x = new Promise(r => 1)) { function Promise() {} }`,
				FileName: "promise-body-function.js",
				TSConfig: "tsconfig.allow-js.json",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "returnsValue",
						Line:      1, Column: 33, EndLine: 1, EndColumn: 34,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "wrapBraces", Output: `function f(x = new Promise(r => {1})) { function Promise() {} }`},
						},
					},
				},
			},
			{
				Code: `function f(x = new Promise(r => 1)) { function Promise() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "returnsValue",
						Line:      1, Column: 33, EndLine: 1, EndColumn: 34,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "wrapBraces", Output: `function f(x = new Promise(r => {1})) { function Promise() {} }`},
						},
					},
				},
			},
			// Locks in upstream isPromiseExecutor() arm 5: the same boundary holds for a
			// `var` in the body, which shares the function's declaration table.
			{
				Code:     `function f(x = new Promise(r => 1)) { var Promise; }`,
				FileName: "promise-body-var.js",
				TSConfig: "tsconfig.allow-js.json",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "returnsValue",
						Line:      1, Column: 33, EndLine: 1, EndColumn: 34,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "wrapBraces", Output: `function f(x = new Promise(r => {1})) { var Promise; }`},
						},
					},
				},
			},
			{
				Code: `function f(x = new Promise(r => 1)) { var Promise: any; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "returnsValue",
						Line:      1, Column: 33, EndLine: 1, EndColumn: 34,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "wrapBraces", Output: `function f(x = new Promise(r => {1})) { var Promise: any; }`},
						},
					},
				},
			},
			// Locks in upstream expressionIsVoid() arm 2: a non-`void` unary operator.
			{
				Code:    `new Promise(r => -1)`,
				Options: map[string]any{"allowVoid": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "returnsValue",
						Line:      1, Column: 18, EndLine: 1, EndColumn: 20,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "prependVoid", Output: `new Promise(r => void -1)`},
							{MessageId: "wrapBraces", Output: `new Promise(r => {-1})`},
						},
					},
				},
			},
			// Locks in upstream expressionIsVoid() arm 2: `typeof` is not `void`.
			{
				Code:    `new Promise(r => typeof x)`,
				Options: map[string]any{"allowVoid": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "returnsValue",
						Line:      1, Column: 18, EndLine: 1, EndColumn: 26,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "prependVoid", Output: `new Promise(r => void typeof x)`},
							{MessageId: "wrapBraces", Output: `new Promise(r => {typeof x})`},
						},
					},
				},
			},
			// Locks in upstream onCodePathStart() arm 3: without allowVoid, `=> void x`
			// is still reported and only wrapBraces is offered.
			{
				Code:    `new Promise(r => void 0)`,
				Options: map[string]any{"allowVoid": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "returnsValue",
						Line:      1, Column: 18, EndLine: 1, EndColumn: 24,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "wrapBraces", Output: `new Promise(r => {void 0})`},
						},
					},
				},
			},
			// Locks in upstream voidPrependFixer() requiresParens arm: `**` binds looser
			// than `void`, so the suggestion parenthesizes.
			{
				Code:    `new Promise(r => 2 ** 3)`,
				Options: map[string]any{"allowVoid": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "returnsValue",
						Line:      1, Column: 18, EndLine: 1, EndColumn: 24,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "prependVoid", Output: `new Promise(r => void (2 ** 3))`},
							{MessageId: "wrapBraces", Output: `new Promise(r => {2 ** 3})`},
						},
					},
				},
			},
			// Locks in upstream voidPrependFixer() isParenthesised arm: existing parentheses
			// suppress the added ones.
			{
				Code:    `new Promise(r => (1, 2))`,
				Options: map[string]any{"allowVoid": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "returnsValue",
						Line:      1, Column: 19, EndLine: 1, EndColumn: 23,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "prependVoid", Output: `new Promise(r => void (1, 2))`},
							{MessageId: "wrapBraces", Output: `new Promise(r => {(1, 2)})`},
						},
					},
				},
			},
			// Locks in upstream isPromiseExecutor() arm 2: extra arguments after the executor
			// keep it in position 0.
			{
				Code:    `new Promise(r => 1, 2)`,
				Options: map[string]any{"allowVoid": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "returnsValue",
						Line:      1, Column: 18, EndLine: 1, EndColumn: 19,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "prependVoid", Output: `new Promise(r => void 1, 2)`},
							{MessageId: "wrapBraces", Output: `new Promise(r => {1}, 2)`},
						},
					},
				},
			},
			// Locks in upstream voidPrependFixer() prependSpace arm: a comment between
			// `return` and the value already separates the tokens.
			{
				Code:    `new Promise(r => { return /*c*/ 1; })`,
				Options: map[string]any{"allowVoid": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "returnsValue",
						Line:      1, Column: 20, EndLine: 1, EndColumn: 35,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "prependVoid", Output: `new Promise(r => { return /*c*/ void 1; })`},
						},
					},
				},
			},
			// ---- Dimension 4: multi-line concise body ----
			{
				Code: `new Promise(r =>
  1
)`,
				Options: map[string]any{"allowVoid": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "returnsValue",
						Line:      2, Column: 3, EndLine: 2, EndColumn: 4,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "prependVoid", Output: `new Promise(r =>
  void 1
)`},
							{MessageId: "wrapBraces", Output: `new Promise(r =>
  {1}
)`},
						},
					},
				},
			},
			// Locks in the schema default: an empty options object behaves like no options.
			{
				Code:    `new Promise(r => 1)`,
				Options: map[string]any{},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "returnsValue",
						Line:      1, Column: 18, EndLine: 1, EndColumn: 19,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "wrapBraces", Output: `new Promise(r => {1})`},
						},
					},
				},
			},
		},
	)
}

// TestNoPromiseExecutorReturnEditDemand verifies that the suggestion builders do
// not change what the rule reports: diagnostic count, message, and range stay
// identical across every edit demand, and the suggestions are materialized only
// when they were requested.
func TestNoPromiseExecutorReturnEditDemand(t *testing.T) {
	t.Parallel()

	const source = "new Promise(r => resolve(1));\n"

	program, sourceFile := createNoPromiseExecutorReturnProgram(t, "edit-demand.ts", source)
	options := rule_tester.ResolveTestCaseOptions(t, &NoPromiseExecutorReturnRule, map[string]any{"allowVoid": true})

	diagnostics := make(map[rule.EditDemand]rule.RuleDiagnostic, 4)
	for _, demand := range []rule.EditDemand{
		rule.EditDemandNone,
		rule.EditDemandAutofix,
		rule.EditDemandSuggestion,
		rule.EditDemandAll,
	} {
		got := lintNoPromiseExecutorReturnWithDemand(program, sourceFile, options, demand)
		if len(got) != 1 {
			t.Fatalf("demand %d: diagnostics = %d, want 1", demand, len(got))
		}
		if got[0].Message.Id != "returnsValue" {
			t.Errorf("demand %d: unexpected message id %q", demand, got[0].Message.Id)
		}
		if got[0].Message.Description != "Return values from promise executor functions cannot be read." {
			t.Errorf("demand %d: unexpected message %q", demand, got[0].Message.Description)
		}
		diagnostics[demand] = got[0]
	}

	diagnosticsOnly := diagnostics[rule.EditDemandNone]
	for demand, diagnostic := range diagnostics {
		want, got := diagnosticsOnly, diagnostic
		want.FixesPtr, want.Suggestions = nil, nil
		got.FixesPtr, got.Suggestions = nil, nil
		if !reflect.DeepEqual(got, want) {
			t.Errorf("demand %d changed diagnostic metadata:\ngot:  %#v\nwant: %#v", demand, got, want)
		}
	}

	// The rule offers suggestions only, so neither the diagnostics-only nor the
	// autofix demand may materialize anything.
	for _, demand := range []rule.EditDemand{rule.EditDemandNone, rule.EditDemandAutofix} {
		diagnostic := diagnostics[demand]
		if diagnostic.FixesPtr != nil || diagnostic.Suggestions != nil {
			t.Errorf(
				"demand %d unexpectedly materialized edits: fixes=%#v suggestions=%#v",
				demand,
				diagnostic.FixesPtr,
				diagnostic.Suggestions,
			)
		}
	}

	wantSuggestionIDs := []string{"prependVoid", "wrapBraces"}
	for _, demand := range []rule.EditDemand{rule.EditDemandSuggestion, rule.EditDemandAll} {
		diagnostic := diagnostics[demand]
		if diagnostic.FixesPtr != nil {
			t.Errorf("demand %d unexpectedly materialized autofixes: %#v", demand, diagnostic.FixesPtr)
		}
		if diagnostic.Suggestions == nil || len(*diagnostic.Suggestions) != len(wantSuggestionIDs) {
			t.Fatalf("demand %d: suggestions = %#v, want %d", demand, diagnostic.Suggestions, len(wantSuggestionIDs))
		}
		for i, wantID := range wantSuggestionIDs {
			suggestion := (*diagnostic.Suggestions)[i]
			if suggestion.Message.Id != wantID {
				t.Errorf("demand %d: suggestion %d id = %q, want %q", demand, i, suggestion.Message.Id, wantID)
			}
			if len(suggestion.Fixes()) == 0 {
				t.Errorf("demand %d: suggestion %d has no fixes", demand, i)
			}
		}
	}
}

func lintNoPromiseExecutorReturnWithDemand(
	program *compiler.Program,
	sourceFile *ast.SourceFile,
	options []any,
	demand rule.EditDemand,
) []rule.RuleDiagnostic {
	var diagnostics []rule.RuleDiagnostic
	linter.LintSingleFile(linter.LintSingleFileOptions{
		Program:         lintprogram.NewFromCompiler(program),
		File:            sourceFile.FileName(),
		HasTypeInfo:     true,
		GetRulesForFile: noPromiseExecutorReturnConfiguredRules(options),
		ExcludePaths:    []string{},
		Consumer: rule.DiagnosticConsumer{
			Demand: demand,
			Report: func(diagnostic rule.RuleDiagnostic) {
				diagnostics = append(diagnostics, diagnostic)
			},
		},
	})
	return diagnostics
}

func noPromiseExecutorReturnConfiguredRules(options []any) func(*ast.SourceFile) []linter.ConfiguredRule {
	return func(*ast.SourceFile) []linter.ConfiguredRule {
		return []linter.ConfiguredRule{{
			Name:     NoPromiseExecutorReturnRule.Name,
			Severity: rule.SeverityError,
			Run: func(ctx rule.RuleContext) rule.RuleListeners {
				return NoPromiseExecutorReturnRule.Run(ctx, options)
			},
		}}
	}
}

func createNoPromiseExecutorReturnProgram(t testing.TB, fileName string, code string) (*compiler.Program, *ast.SourceFile) {
	t.Helper()

	root := fixtures.GetRootDir()
	fs := utils.NewOverlayVFS(root.FS, map[string]string{tspath.ResolvePath(root.Dir, fileName): code})
	host := utils.CreateCompilerHost(root.Dir, fs)
	program, err := utils.CreateProgram(true, fs, root.Dir, "tsconfig.json", host)
	if err != nil {
		t.Fatalf("failed to create program: %v", err)
	}
	sourceFile := program.GetSourceFile(fileName)
	if sourceFile == nil {
		t.Fatalf("source file %q not found", fileName)
	}
	return program, sourceFile
}
