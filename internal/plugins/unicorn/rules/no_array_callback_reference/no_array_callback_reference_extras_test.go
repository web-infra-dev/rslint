// cspell:ignore evals
package no_array_callback_reference_test

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_array_callback_reference"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// AST and scope regressions checked against eslint-plugin-unicorn v76.0.0.
func TestNoArrayCallbackReferenceExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_array_callback_reference.NoArrayCallbackReferenceRule,
		[]rule_tester.ValidTestCase{
			{Code: "myLib?.().map(callback)", FileName: "file.js", Options: []any{map[string]any{"ignore": []any{"myLib"}}}},
			{Code: "jQuery?.().map(callback)", FileName: "file.js"},
			{Code: "((myLib))?.().map(callback)", FileName: "file.js", Options: []any{map[string]any{"ignore": []any{"myLib"}}}},
			{Code: "(myLib()).map(callback)", FileName: "file.js", Options: []any{map[string]any{"ignore": []any{"myLib"}}}},
			{Code: "myLib.tools?.().map(callback)", FileName: "file.js", Options: []any{map[string]any{"ignore": []any{"myLib.tools"}}}},
			// Checked against v76 with TypeScript parser services enabled.
			{Code: "export {}; class Service { map(callback: Function) {} } function Base(): typeof Service { return Service; } class Child extends Base() { run() { super.map(callback); } }", FileName: "file.ts"},
			{Code: "export {}; class Service { map(callback: Function) {} } const Parent = () => Service; class Child extends Parent() { run() { super.map(callback); } }", FileName: "file.ts"},
			{Code: "export {}; class Rows extends Array<number> { static map(callback: Function) {} } function rows(): typeof Rows { return Rows; } rows().map(callback);", FileName: "file.ts"},
			{Code: "function guard(x: unknown): x is string { return true; } interface guard {} array.filter(guard);", FileName: "file.ts"},
			{Code: "const Query = class {}; const Alias = Query; new Alias().map(callback);", FileName: "file.js"},
			{Code: "const cb = function guard(x: unknown): x is string { array.filter(guard); return true; };", FileName: "file.ts"},
			{Code: "array.map?.(callback)", FileName: "file.js"},
			{Code: "(array?.map)(callback)", FileName: "file.js"},
			{Code: "array.map(callback, ...args)", FileName: "file.js"},
			{Code: "array[\"map\"](callback)", FileName: "file.js"},
			{Code: "class C { #map() {} use() { this.#map(callback); } }", FileName: "file.js"},
			{Code: "async function run() { await (array.map(callback)); }", FileName: "file.js"},
			{Code: "async function run() { array.map(await callback); }", FileName: "file.js"},
			{Code: "const query = {}; { const alias = query; model.find(alias); }", FileName: "file.js"},
			{Code: "array.map(({} as const)!)", FileName: "file.ts"},
			{Code: "class Service {} const Alias = Service; new Alias().map(callback);", FileName: "file.js"},
			{Code: "class Service {} class Child extends Service { use() { super.map(callback); } }", FileName: "file.js"},
			{Code: "class Rows extends Array { static use() { super.map(callback); } }", FileName: "file.js"},
			{Code: "(flag ? [] : new Set()).map(callback);", FileName: "file.js"},
			{Code: "(flag ? [] : null).map(callback);", FileName: "file.js"},
			{Code: "declare const array: unknown | Set<string>; array.forEach(callback);", FileName: "file.ts"},
			{Code: "function rows(): Set<string> { return new Set(); } rows().forEach(callback);", FileName: "file.ts"},
			{Code: "const guard = ((x: unknown): x is string => true); array.every(guard);", FileName: "file.ts"},
			{Code: "function f(guard: ((x: unknown) => x is string)) { array.filter(guard); }", FileName: "file.ts"},
			{Code: "Promise.map(callback)", FileName: "file.js", Options: []any{map[string]any{}}},
			{Code: "lib.tools().map(callback)", FileName: "file.js", Options: []any{map[string]any{"ignore": []any{" lib.tools "}}}},
		}, []rule_tester.InvalidTestCase{
			{Code: "(myLib?.()).map(callback)", FileName: "file.js", Options: []any{map[string]any{"ignore": []any{"myLib"}}}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `callback` directly to `.map(…)`.", Line: 1, Column: 17, EndLine: 1, EndColumn: 25, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "(myLib?.()).map((element) => callback(element))"},
					{MessageId: "replace-with-name", Output: "(myLib?.()).map((element, index) => callback(element, index))"},
					{MessageId: "replace-with-name", Output: "(myLib?.()).map((element, index, array) => callback(element, index, array))"},
				}},
			}},
			{Code: "(jQuery?.()).map(callback)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `callback` directly to `.map(…)`.", Line: 1, Column: 18, EndLine: 1, EndColumn: 26, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "(jQuery?.()).map((element) => callback(element))"},
					{MessageId: "replace-with-name", Output: "(jQuery?.()).map((element, index) => callback(element, index))"},
					{MessageId: "replace-with-name", Output: "(jQuery?.()).map((element, index, array) => callback(element, index, array))"},
				}},
			}},
			{Code: "(myLib.tools?.()).map(callback)", FileName: "file.js", Options: []any{map[string]any{"ignore": []any{"myLib.tools"}}}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `callback` directly to `.map(…)`.", Line: 1, Column: 23, EndLine: 1, EndColumn: 31, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "(myLib.tools?.()).map((element) => callback(element))"},
					{MessageId: "replace-with-name", Output: "(myLib.tools?.()).map((element, index) => callback(element, index))"},
					{MessageId: "replace-with-name", Output: "(myLib.tools?.()).map((element, index, array) => callback(element, index, array))"},
				}},
			}},
			{Code: "myLib?.().map(callback)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `callback` directly to `.map(…)`.", Line: 1, Column: 15, EndLine: 1, EndColumn: 23, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "myLib?.().map((element) => callback(element))"},
					{MessageId: "replace-with-name", Output: "myLib?.().map((element, index) => callback(element, index))"},
					{MessageId: "replace-with-name", Output: "myLib?.().map((element, index, array) => callback(element, index, array))"},
				}},
			}},
			// Checked against v76 with TypeScript parser services enabled.
			{Code: "export {}; class Rows extends Array<number> {} const Parent = () => Rows; class Child extends Parent() { run() { super.map(callback); } }", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `callback` directly to `.map(…)`.", Line: 1, Column: 124, EndLine: 1, EndColumn: 132, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "export {}; class Rows extends Array<number> {} const Parent = () => Rows; class Child extends Parent() { run() { super.map((element) => callback(element)); } }"},
					{MessageId: "replace-with-name", Output: "export {}; class Rows extends Array<number> {} const Parent = () => Rows; class Child extends Parent() { run() { super.map((element, index) => callback(element, index)); } }"},
					{MessageId: "replace-with-name", Output: "export {}; class Rows extends Array<number> {} const Parent = () => Rows; class Child extends Parent() { run() { super.map((element, index, array) => callback(element, index, array)); } }"},
				}},
			}},
			{Code: "array.map(object?.callback!)", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-without-name", Message: "Do not pass function directly to `.map(…)`.", Line: 1, Column: 11, EndLine: 1, EndColumn: 28, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-without-name", Output: "array.map((element) => object?.callback!(element))"},
					{MessageId: "replace-without-name", Output: "array.map((element, index) => object?.callback!(element, index))"},
					{MessageId: "replace-without-name", Output: "array.map((element, index, array) => object?.callback!(element, index, array))"},
				}},
			}},
			{Code: "array.map(callback?.()!)", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-without-name", Message: "Do not pass function directly to `.map(…)`.", Line: 1, Column: 11, EndLine: 1, EndColumn: 24, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-without-name", Output: "array.map((element) => callback?.()!(element))"},
					{MessageId: "replace-without-name", Output: "array.map((element, index) => callback?.()!(element, index))"},
					{MessageId: "replace-without-name", Output: "array.map((element, index, array) => callback?.()!(element, index, array))"},
				}},
			}},
			{Code: "array.map(callback<string>)", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-without-name", Message: "Do not pass function directly to `.map(…)`.", Line: 1, Column: 11, EndLine: 1, EndColumn: 27, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-without-name", Output: "array.map((element) => callback<string>(element))"},
					{MessageId: "replace-without-name", Output: "array.map((element, index) => callback<string>(element, index))"},
					{MessageId: "replace-without-name", Output: "array.map((element, index, array) => callback<string>(element, index, array))"},
				}},
			}},
			{Code: "import guard = require(\"guards\"); array.filter(guard); array.map(guard);", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `guard` directly to `.map(…)`.", Line: 1, Column: 66, EndLine: 1, EndColumn: 71, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "import guard = require(\"guards\"); array.filter(guard); array.map((element) => guard(element));"},
					{MessageId: "replace-with-name", Output: "import guard = require(\"guards\"); array.filter(guard); array.map((element, index) => guard(element, index));"},
					{MessageId: "replace-with-name", Output: "import guard = require(\"guards\"); array.filter(guard); array.map((element, index, array) => guard(element, index, array));"},
				}},
			}},
			{Code: "interface guard {} function guard(x: unknown): x is string { return true; } array.filter(guard);", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `guard` directly to `.filter(…)`.", Line: 1, Column: 90, EndLine: 1, EndColumn: 95, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "interface guard {} function guard(x: unknown): x is string { return true; } array.filter((element) => guard(element));"},
					{MessageId: "replace-with-name", Output: "interface guard {} function guard(x: unknown): x is string { return true; } array.filter((element, index) => guard(element, index));"},
					{MessageId: "replace-with-name", Output: "interface guard {} function guard(x: unknown): x is string { return true; } array.filter((element, index, array) => guard(element, index, array));"},
				}},
			}},
			{Code: "namespace guard {} function guard(x: unknown): x is string { return true; } array.filter(guard);", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `guard` directly to `.filter(…)`.", Line: 1, Column: 90, EndLine: 1, EndColumn: 95, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "namespace guard {} function guard(x: unknown): x is string { return true; } array.filter((element) => guard(element));"},
					{MessageId: "replace-with-name", Output: "namespace guard {} function guard(x: unknown): x is string { return true; } array.filter((element, index) => guard(element, index));"},
					{MessageId: "replace-with-name", Output: "namespace guard {} function guard(x: unknown): x is string { return true; } array.filter((element, index, array) => guard(element, index, array));"},
				}},
			}},
			{Code: "const query = {}; function f(query = callback) { array.filter(query); }", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `query` directly to `.filter(…)`.", Line: 1, Column: 63, EndLine: 1, EndColumn: 68, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "const query = {}; function f(query = callback) { array.filter((element) => query(element)); }"},
					{MessageId: "replace-with-name", Output: "const query = {}; function f(query = callback) { array.filter((element, index) => query(element, index)); }"},
					{MessageId: "replace-with-name", Output: "const query = {}; function f(query = callback) { array.filter((element, index, array) => query(element, index, array)); }"},
				}},
			}},
			{Code: "class Service { map() {} } function Base() { return Service; } class Child extends Base() { run() { super.map(callback); } }", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `callback` directly to `.map(…)`.", Line: 1, Column: 111, EndLine: 1, EndColumn: 119, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "class Service { map() {} } function Base() { return Service; } class Child extends Base() { run() { super.map((element) => callback(element)); } }"},
					{MessageId: "replace-with-name", Output: "class Service { map() {} } function Base() { return Service; } class Child extends Base() { run() { super.map((element, index) => callback(element, index)); } }"},
					{MessageId: "replace-with-name", Output: "class Service { map() {} } function Base() { return Service; } class Child extends Base() { run() { super.map((element, index, array) => callback(element, index, array)); } }"},
				}},
			}},
			{Code: "class Rows extends Array { static { super.map(callback); } field = () => super.map(callback); }", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `callback` directly to `.map(…)`.", Line: 1, Column: 84, EndLine: 1, EndColumn: 92, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "class Rows extends Array { static { super.map(callback); } field = () => super.map((element) => callback(element)); }"},
					{MessageId: "replace-with-name", Output: "class Rows extends Array { static { super.map(callback); } field = () => super.map((element, index) => callback(element, index)); }"},
					{MessageId: "replace-with-name", Output: "class Rows extends Array { static { super.map(callback); } field = () => super.map((element, index, array) => callback(element, index, array)); }"},
				}},
			}},
			{Code: "using query = {}; array.map(query);", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `query` directly to `.map(…)`.", Line: 1, Column: 29, EndLine: 1, EndColumn: 34, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "using query = {}; array.map((element) => query(element));"},
					{MessageId: "replace-with-name", Output: "using query = {}; array.map((element, index) => query(element, index));"},
					{MessageId: "replace-with-name", Output: "using query = {}; array.map((element, index, array) => query(element, index, array));"},
				}},
			}},
			// The documented inflection difference affects only suggestion names.
			{Code: "People.map(callback)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `callback` directly to `.map(…)`.", Line: 1, Column: 12, EndLine: 1, EndColumn: 20, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "People.map((element) => callback(element))"},
					{MessageId: "replace-with-name", Output: "People.map((element, index) => callback(element, index))"},
					{MessageId: "replace-with-name", Output: "People.map((element, index, array) => callback(element, index, array))"},
				}},
			}},
			{Code: "((array)).map(((callback)))", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `callback` directly to `.map(…)`.", Line: 1, Column: 17, EndLine: 1, EndColumn: 25, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "((array)).map((element) => ((callback))(element))"},
					{MessageId: "replace-with-name", Output: "((array)).map((element, index) => ((callback))(element, index))"},
					{MessageId: "replace-with-name", Output: "((array)).map((element, index, array) => ((callback))(element, index, array))"},
				}},
			}},
			{Code: "array.m\\u0061p(callback)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `callback` directly to `.map(…)`.", Line: 1, Column: 16, EndLine: 1, EndColumn: 24, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "array.m\\u0061p((element) => callback(element))"},
					{MessageId: "replace-with-name", Output: "array.m\\u0061p((element, index) => callback(element, index))"},
					{MessageId: "replace-with-name", Output: "array.m\\u0061p((element, index, array) => callback(element, index, array))"},
				}},
			}},
			{Code: "array.map(obj[\"callback\"])", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-without-name", Message: "Do not pass function directly to `.map(…)`.", Line: 1, Column: 11, EndLine: 1, EndColumn: 26, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-without-name", Output: "array.map((element) => obj[\"callback\"](element))"},
					{MessageId: "replace-without-name", Output: "array.map((element, index) => obj[\"callback\"](element, index))"},
					{MessageId: "replace-without-name", Output: "array.map((element, index, array) => obj[\"callback\"](element, index, array))"},
				}},
			}},
			{Code: "class C { #callback() {} use() { array.map(this.#callback); } }", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-without-name", Message: "Do not pass function directly to `.map(…)`.", Line: 1, Column: 44, EndLine: 1, EndColumn: 58, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-without-name", Output: "class C { #callback() {} use() { array.map((element) => this.#callback(element)); } }"},
					{MessageId: "replace-without-name", Output: "class C { #callback() {} use() { array.map((element, index) => this.#callback(element, index)); } }"},
					{MessageId: "replace-without-name", Output: "class C { #callback() {} use() { array.map((element, index, array) => this.#callback(element, index, array)); } }"},
				}},
			}},
			{Code: "array.map(obj?.callback)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-without-name", Message: "Do not pass function directly to `.map(…)`.", Line: 1, Column: 11, EndLine: 1, EndColumn: 24, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-without-name", Output: "array.map((element) => obj?.callback(element))"},
					{MessageId: "replace-without-name", Output: "array.map((element, index) => obj?.callback(element, index))"},
					{MessageId: "replace-without-name", Output: "array.map((element, index, array) => obj?.callback(element, index, array))"},
				}},
			}},
			{Code: "array.map(callback?.())", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-without-name", Message: "Do not pass function directly to `.map(…)`.", Line: 1, Column: 11, EndLine: 1, EndColumn: 23, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-without-name", Output: "array.map((element) => callback?.()(element))"},
					{MessageId: "replace-without-name", Output: "array.map((element, index) => callback?.()(element, index))"},
					{MessageId: "replace-without-name", Output: "array.map((element, index, array) => callback?.()(element, index, array))"},
				}},
			}},
			{Code: "array.map(import(\"callback\"))", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-without-name", Message: "Do not pass function directly to `.map(…)`.", Line: 1, Column: 11, EndLine: 1, EndColumn: 29, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-without-name", Output: "array.map((element) => import(\"callback\")(element))"},
					{MessageId: "replace-without-name", Output: "array.map((element, index) => import(\"callback\")(element, index))"},
					{MessageId: "replace-without-name", Output: "array.map((element, index, array) => import(\"callback\")(element, index, array))"},
				}},
			}},
			{Code: "async function run() { await array?.map(callback); }", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `callback` directly to `.map(…)`.", Line: 1, Column: 41, EndLine: 1, EndColumn: 49, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "async function run() { await array?.map((element) => callback(element)); }"},
					{MessageId: "replace-with-name", Output: "async function run() { await array?.map((element, index) => callback(element, index)); }"},
					{MessageId: "replace-with-name", Output: "async function run() { await array?.map((element, index, array) => callback(element, index, array)); }"},
				}},
			}},
			{Code: "function* run() { array.reduce(yield callback); }", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-without-name", Message: "Do not pass function directly to `.reduce(…)`.", Line: 1, Column: 32, EndLine: 1, EndColumn: 46, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "array.map(/* before */ ((callback /* inside */)) /* after */)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `callback` directly to `.map(…)`.", Line: 1, Column: 26, EndLine: 1, EndColumn: 34, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "array.map(/* before */ (element) => ((callback /* inside */))(element) /* after */)"},
					{MessageId: "replace-with-name", Output: "array.map(/* before */ (element, index) => ((callback /* inside */))(element, index) /* after */)"},
					{MessageId: "replace-with-name", Output: "array.map(/* before */ (element, index, array) => ((callback /* inside */))(element, index, array) /* after */)"},
				}},
			}},
			{Code: "array.map(/** @type {Function} */ (callback))", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `callback` directly to `.map(…)`.", Line: 1, Column: 36, EndLine: 1, EndColumn: 44, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "array.map(/** @type {Function} */ (element) => (callback)(element))"},
					{MessageId: "replace-with-name", Output: "array.map(/** @type {Function} */ (element, index) => (callback)(element, index))"},
					{MessageId: "replace-with-name", Output: "array.map(/** @type {Function} */ (element, index, array) => (callback)(element, index, array))"},
				}},
			}},
			{Code: "array.map(callback as Function)", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-without-name", Message: "Do not pass function directly to `.map(…)`.", Line: 1, Column: 11, EndLine: 1, EndColumn: 31, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-without-name", Output: "array.map((element) => (callback as Function)(element))"},
					{MessageId: "replace-without-name", Output: "array.map((element, index) => (callback as Function)(element, index))"},
					{MessageId: "replace-without-name", Output: "array.map((element, index, array) => (callback as Function)(element, index, array))"},
				}},
			}},
			{Code: "array.map(callback!)", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-without-name", Message: "Do not pass function directly to `.map(…)`.", Line: 1, Column: 11, EndLine: 1, EndColumn: 20, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-without-name", Output: "array.map((element) => (callback!)(element))"},
					{MessageId: "replace-without-name", Output: "array.map((element, index) => (callback!)(element, index))"},
					{MessageId: "replace-without-name", Output: "array.map((element, index, array) => (callback!)(element, index, array))"},
				}},
			}},
			{Code: "array.map(<Function>callback)", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-without-name", Message: "Do not pass function directly to `.map(…)`.", Line: 1, Column: 11, EndLine: 1, EndColumn: 29, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-without-name", Output: "array.map((element) => (<Function>callback)(element))"},
					{MessageId: "replace-without-name", Output: "array.map((element, index) => (<Function>callback)(element, index))"},
					{MessageId: "replace-without-name", Output: "array.map((element, index, array) => (<Function>callback)(element, index, array))"},
				}},
			}},
			{Code: "array.map(callback satisfies Function)", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-without-name", Message: "Do not pass function directly to `.map(…)`.", Line: 1, Column: 11, EndLine: 1, EndColumn: 38, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-without-name", Output: "array.map((element) => (callback satisfies Function)(element))"},
					{MessageId: "replace-without-name", Output: "array.map((element, index) => (callback satisfies Function)(element, index))"},
					{MessageId: "replace-without-name", Output: "array.map((element, index, array) => (callback satisfies Function)(element, index, array))"},
				}},
			}},
			{Code: "const view = <View values={array.map(callback)} />;", FileName: "file.tsx", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `callback` directly to `.map(…)`.", Line: 1, Column: 38, EndLine: 1, EndColumn: 46, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "const view = <View values={array.map((element) => callback(element))} />;"},
					{MessageId: "replace-with-name", Output: "const view = <View values={array.map((element, index) => callback(element, index))} />;"},
					{MessageId: "replace-with-name", Output: "const view = <View values={array.map((element, index, array) => callback(element, index, array))} />;"},
				}},
			}},
			{Code: "const 文字 = \"😀\";\narray.map(\n  工具.转换\n);", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-without-name", Message: "Do not pass function directly to `.map(…)`.", Line: 3, Column: 3, EndLine: 3, EndColumn: 8, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-without-name", Output: "const 文字 = \"😀\";\narray.map(\n  (element) => 工具.转换(element)\n);"},
					{MessageId: "replace-without-name", Output: "const 文字 = \"😀\";\narray.map(\n  (element, index) => 工具.转换(element, index)\n);"},
					{MessageId: "replace-without-name", Output: "const 文字 = \"😀\";\narray.map(\n  (element, index, array) => 工具.转换(element, index, array)\n);"},
				}},
			}},
			{Code: "const a = b; const b = a; array.map(a);", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `a` directly to `.map(…)`.", Line: 1, Column: 37, EndLine: 1, EndColumn: 38, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "const a = b; const b = a; array.map((element) => a(element));"},
					{MessageId: "replace-with-name", Output: "const a = b; const b = a; array.map((element, index) => a(element, index));"},
					{MessageId: "replace-with-name", Output: "const a = b; const b = a; array.map((element, index, array) => a(element, index, array));"},
				}},
			}},
			{Code: "const result = factory(); array.map(result);", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `result` directly to `.map(…)`.", Line: 1, Column: 37, EndLine: 1, EndColumn: 43, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "const result = factory(); array.map((element) => result(element));"},
					{MessageId: "replace-with-name", Output: "const result = factory(); array.map((element, index) => result(element, index));"},
					{MessageId: "replace-with-name", Output: "const result = factory(); array.map((element, index, array) => result(element, index, array));"},
				}},
			}},
			{Code: "const fn = source.callback; array.map(fn);", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `fn` directly to `.map(…)`.", Line: 1, Column: 39, EndLine: 1, EndColumn: 41, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "const fn = source.callback; array.map((element) => fn(element));"},
					{MessageId: "replace-with-name", Output: "const fn = source.callback; array.map((element, index) => fn(element, index));"},
					{MessageId: "replace-with-name", Output: "const fn = source.callback; array.map((element, index, array) => fn(element, index, array));"},
				}},
			}},
			{Code: "const {query} = source; model.find(query);", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `query` directly to `.find(…)`.", Line: 1, Column: 36, EndLine: 1, EndColumn: 41, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "const {query} = source; model.find((element) => query(element));"},
					{MessageId: "replace-with-name", Output: "const {query} = source; model.find((element, index) => query(element, index));"},
					{MessageId: "replace-with-name", Output: "const {query} = source; model.find((element, index, array) => query(element, index, array));"},
				}},
			}},
			{Code: "const query = {}; function run(query) { model.find(query); }", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `query` directly to `.find(…)`.", Line: 1, Column: 52, EndLine: 1, EndColumn: 57, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "const query = {}; function run(query) { model.find((element) => query(element)); }"},
					{MessageId: "replace-with-name", Output: "const query = {}; function run(query) { model.find((element, index) => query(element, index)); }"},
					{MessageId: "replace-with-name", Output: "const query = {}; function run(query) { model.find((element, index, array) => query(element, index, array)); }"},
				}},
			}},
			{Code: "var query = {}; model.find(query);", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `query` directly to `.find(…)`.", Line: 1, Column: 28, EndLine: 1, EndColumn: 33, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "var query = {}; model.find((element) => query(element));"},
					{MessageId: "replace-with-name", Output: "var query = {}; model.find((element, index) => query(element, index));"},
					{MessageId: "replace-with-name", Output: "var query = {}; model.find((element, index, array) => query(element, index, array));"},
				}},
			}},
			{Code: "array.map(condition ? (inner ? callback : Boolean) : (callback2))", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `callback` directly to `.map(…)`.", Line: 1, Column: 32, EndLine: 1, EndColumn: 40, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "array.map(condition ? (inner ? (element) => callback(element) : Boolean) : (callback2))"},
					{MessageId: "replace-with-name", Output: "array.map(condition ? (inner ? (element, index) => callback(element, index) : Boolean) : (callback2))"},
					{MessageId: "replace-with-name", Output: "array.map(condition ? (inner ? (element, index, array) => callback(element, index, array) : Boolean) : (callback2))"},
				}},
				{MessageId: "error-with-name", Message: "Do not pass function `callback2` directly to `.map(…)`.", Line: 1, Column: 55, EndLine: 1, EndColumn: 64, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "array.map(condition ? (inner ? callback : Boolean) : (element) => (callback2)(element))"},
					{MessageId: "replace-with-name", Output: "array.map(condition ? (inner ? callback : Boolean) : (element, index) => (callback2)(element, index))"},
					{MessageId: "replace-with-name", Output: "array.map(condition ? (inner ? callback : Boolean) : (element, index, array) => (callback2)(element, index, array))"},
				}},
			}},
			{Code: "class Rows extends Array {} new Rows().map(callback);", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `callback` directly to `.map(…)`.", Line: 1, Column: 44, EndLine: 1, EndColumn: 52, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "class Rows extends Array {} new Rows().map((element) => callback(element));"},
					{MessageId: "replace-with-name", Output: "class Rows extends Array {} new Rows().map((element, index) => callback(element, index));"},
					{MessageId: "replace-with-name", Output: "class Rows extends Array {} new Rows().map((element, index, array) => callback(element, index, array));"},
				}},
			}},
			{Code: "const Rows = class extends Array {}; const Alias = Rows; new Alias().map(callback);", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `callback` directly to `.map(…)`.", Line: 1, Column: 74, EndLine: 1, EndColumn: 82, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "const Rows = class extends Array {}; const Alias = Rows; new Alias().map((element) => callback(element));"},
					{MessageId: "replace-with-name", Output: "const Rows = class extends Array {}; const Alias = Rows; new Alias().map((element, index) => callback(element, index));"},
					{MessageId: "replace-with-name", Output: "const Rows = class extends Array {}; const Alias = Rows; new Alias().map((element, index, array) => callback(element, index, array));"},
				}},
			}},
			{Code: "class Rows extends Array { use() { super.map(callback); } }", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `callback` directly to `.map(…)`.", Line: 1, Column: 46, EndLine: 1, EndColumn: 54, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "class Rows extends Array { use() { super.map((element) => callback(element)); } }"},
					{MessageId: "replace-with-name", Output: "class Rows extends Array { use() { super.map((element, index) => callback(element, index)); } }"},
					{MessageId: "replace-with-name", Output: "class Rows extends Array { use() { super.map((element, index, array) => callback(element, index, array)); } }"},
				}},
			}},
			{Code: "type Maybe = string[] | null; declare const array: Maybe; array.map(callback);", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `callback` directly to `.map(…)`.", Line: 1, Column: 69, EndLine: 1, EndColumn: 77, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "type Maybe = string[] | null; declare const array: Maybe; array.map((element) => callback(element));"},
					{MessageId: "replace-with-name", Output: "type Maybe = string[] | null; declare const array: Maybe; array.map((element, index) => callback(element, index));"},
					{MessageId: "replace-with-name", Output: "type Maybe = string[] | null; declare const array: Maybe; array.map((element, index, array) => callback(element, index, array));"},
				}},
			}},
			{Code: "type Null = null; declare const array: string[] | Null; array.map(callback);", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `callback` directly to `.map(…)`.", Line: 1, Column: 67, EndLine: 1, EndColumn: 75, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "type Null = null; declare const array: string[] | Null; array.map((element) => callback(element));"},
					{MessageId: "replace-with-name", Output: "type Null = null; declare const array: string[] | Null; array.map((element, index) => callback(element, index));"},
					{MessageId: "replace-with-name", Output: "type Null = null; declare const array: string[] | Null; array.map((element, index, array) => callback(element, index, array));"},
				}},
			}},
			{Code: "class Rows extends Array<string> {} function f(rows: Rows) { rows.map(callback); }", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `callback` directly to `.map(…)`.", Line: 1, Column: 71, EndLine: 1, EndColumn: 79, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "class Rows extends Array<string> {} function f(rows: Rows) { rows.map((row) => callback(row)); }"},
					{MessageId: "replace-with-name", Output: "class Rows extends Array<string> {} function f(rows: Rows) { rows.map((row, index) => callback(row, index)); }"},
					{MessageId: "replace-with-name", Output: "class Rows extends Array<string> {} function f(rows: Rows) { rows.map((row, index, rows) => callback(row, index, rows)); }"},
				}},
			}},
			{Code: "import guard from \"./guards\"; array.map(guard); array.filter(guard);", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `guard` directly to `.map(…)`.", Line: 1, Column: 41, EndLine: 1, EndColumn: 46, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "import guard from \"./guards\"; array.map((element) => guard(element)); array.filter(guard);"},
					{MessageId: "replace-with-name", Output: "import guard from \"./guards\"; array.map((element, index) => guard(element, index)); array.filter(guard);"},
					{MessageId: "replace-with-name", Output: "import guard from \"./guards\"; array.map((element, index, array) => guard(element, index, array)); array.filter(guard);"},
				}},
			}},
			{Code: "function guard(x: unknown): x is string { return true } array.some(guard); array.findIndex(guard); array.findLastIndex(guard);", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `guard` directly to `.some(…)`.", Line: 1, Column: 68, EndLine: 1, EndColumn: 73, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "function guard(x: unknown): x is string { return true } array.some((element) => guard(element)); array.findIndex(guard); array.findLastIndex(guard);"},
					{MessageId: "replace-with-name", Output: "function guard(x: unknown): x is string { return true } array.some((element, index) => guard(element, index)); array.findIndex(guard); array.findLastIndex(guard);"},
					{MessageId: "replace-with-name", Output: "function guard(x: unknown): x is string { return true } array.some((element, index, array) => guard(element, index, array)); array.findIndex(guard); array.findLastIndex(guard);"},
				}},
				{MessageId: "error-with-name", Message: "Do not pass function `guard` directly to `.findIndex(…)`.", Line: 1, Column: 92, EndLine: 1, EndColumn: 97, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "function guard(x: unknown): x is string { return true } array.some(guard); array.findIndex((element) => guard(element)); array.findLastIndex(guard);"},
					{MessageId: "replace-with-name", Output: "function guard(x: unknown): x is string { return true } array.some(guard); array.findIndex((element, index) => guard(element, index)); array.findLastIndex(guard);"},
					{MessageId: "replace-with-name", Output: "function guard(x: unknown): x is string { return true } array.some(guard); array.findIndex((element, index, array) => guard(element, index, array)); array.findLastIndex(guard);"},
				}},
				{MessageId: "error-with-name", Message: "Do not pass function `guard` directly to `.findLastIndex(…)`.", Line: 1, Column: 120, EndLine: 1, EndColumn: 125, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "function guard(x: unknown): x is string { return true } array.some(guard); array.findIndex(guard); array.findLastIndex((element) => guard(element));"},
					{MessageId: "replace-with-name", Output: "function guard(x: unknown): x is string { return true } array.some(guard); array.findIndex(guard); array.findLastIndex((element, index) => guard(element, index));"},
					{MessageId: "replace-with-name", Output: "function guard(x: unknown): x is string { return true } array.some(guard); array.findIndex(guard); array.findLastIndex((element, index, array) => guard(element, index, array));"},
				}},
			}},
			{Code: "type Guard = (x: unknown) => x is string; declare const guard: Guard; array.filter(guard);", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `guard` directly to `.filter(…)`.", Line: 1, Column: 84, EndLine: 1, EndColumn: 89, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "type Guard = (x: unknown) => x is string; declare const guard: Guard; array.filter((element) => guard(element));"},
					{MessageId: "replace-with-name", Output: "type Guard = (x: unknown) => x is string; declare const guard: Guard; array.filter((element, index) => guard(element, index));"},
					{MessageId: "replace-with-name", Output: "type Guard = (x: unknown) => x is string; declare const guard: Guard; array.filter((element, index, array) => guard(element, index, array));"},
				}},
			}},
			{Code: "array.map(callback)", FileName: "file.js", Options: []any{map[string]any{"ignore": []any{}}}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `callback` directly to `.map(…)`.", Line: 1, Column: 11, EndLine: 1, EndColumn: 19, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "array.map((element) => callback(element))"},
					{MessageId: "replace-with-name", Output: "array.map((element, index) => callback(element, index))"},
					{MessageId: "replace-with-name", Output: "array.map((element, index, array) => callback(element, index, array))"},
				}},
			}},
			{Code: "lib[\"tools\"].map(callback)", FileName: "file.js", Options: []any{map[string]any{"ignore": []any{"lib.tools"}}}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `callback` directly to `.map(…)`.", Line: 1, Column: 18, EndLine: 1, EndColumn: 26, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "lib[\"tools\"].map((element) => callback(element))"},
					{MessageId: "replace-with-name", Output: "lib[\"tools\"].map((element, index) => callback(element, index))"},
					{MessageId: "replace-with-name", Output: "lib[\"tools\"].map((element, index, array) => callback(element, index, array))"},
				}},
			}},
			{Code: "lib?.tools.map(callback)", FileName: "file.js", Options: []any{map[string]any{"ignore": []any{"lib.tools"}}}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `callback` directly to `.map(…)`.", Line: 1, Column: 16, EndLine: 1, EndColumn: 24, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "lib?.tools.map((element) => callback(element))"},
					{MessageId: "replace-with-name", Output: "lib?.tools.map((element, index) => callback(element, index))"},
					{MessageId: "replace-with-name", Output: "lib?.tools.map((element, index, array) => callback(element, index, array))"},
				}},
			}},
			{Code: "lib.other.map(callback)", FileName: "file.js", Options: []any{map[string]any{"ignore": []any{"lib.tools"}}}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `callback` directly to `.map(…)`.", Line: 1, Column: 15, EndLine: 1, EndColumn: 23, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "lib.other.map((element) => callback(element))"},
					{MessageId: "replace-with-name", Output: "lib.other.map((element, index) => callback(element, index))"},
					{MessageId: "replace-with-name", Output: "lib.other.map((element, index, array) => callback(element, index, array))"},
				}},
			}},
			{Code: "people.map(person)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `person` directly to `.map(…)`.", Line: 1, Column: 12, EndLine: 1, EndColumn: 18, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "people.map((element) => person(element))"},
					{MessageId: "replace-with-name", Output: "people.map((element, index) => person(element, index))"},
					{MessageId: "replace-with-name", Output: "people.map((element, index, people) => person(element, index, people))"},
				}},
			}},
			{Code: "arguments.map(callback)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `callback` directly to `.map(…)`.", Line: 1, Column: 15, EndLine: 1, EndColumn: 23, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "arguments.map((argument) => callback(argument))"},
					{MessageId: "replace-with-name", Output: "arguments.map((argument, index) => callback(argument, index))"},
					{MessageId: "replace-with-name", Output: "arguments.map((argument, index, array) => callback(argument, index, array))"},
				}},
			}},
			{Code: "evals.map(callback)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `callback` directly to `.map(…)`.", Line: 1, Column: 11, EndLine: 1, EndColumn: 19, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "evals.map((element) => callback(element))"},
					{MessageId: "replace-with-name", Output: "evals.map((element, index) => callback(element, index))"},
					{MessageId: "replace-with-name", Output: "evals.map((element, index, evals) => callback(element, index, evals))"},
				}},
			}},
		})
}

func TestNoArrayCallbackReferenceEditDemand(t *testing.T) {
	for _, test := range []struct {
		code, callback, method string
		parameters             []string
	}{
		{"bar.map(fn)", "fn", "map", []string{"element", "element, index", "element, index, array"}},
		{"bar.reduce(lib.fn)", "lib.fn", "reduce", []string{"accumulator, element", "accumulator, element, index", "accumulator, element, index, array"}},
		{"bar.forEach(fn)", "fn", "forEach", []string{"element", "element, index", "element, index, array"}},
		{"function* run() { bar.map(yield fn); }", "yield fn", "map", nil},
	} {
		t.Run(test.code, func(t *testing.T) {
			helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
			program, file, err := helper.CreateTestProgram(test.code, "edit-demand.js", "tsconfig.json")
			if err != nil {
				t.Fatal(err)
			}
			var all rule.RuleDiagnostic
			for _, demand := range []rule.EditDemand{rule.EditDemandAll, rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion} {
				var diagnostics []rule.RuleDiagnostic
				linter.LintSingleFile(linter.LintSingleFileOptions{
					Program: lintprogram.NewFromCompiler(program), File: file.FileName(),
					GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
						return []rule.ConfiguredRule{{Name: no_array_callback_reference.NoArrayCallbackReferenceRule.Name, Severity: rule.SeverityError,
							Run: func(ctx rule.RuleContext) rule.RuleListeners {
								return no_array_callback_reference.NoArrayCallbackReferenceRule.Run(ctx, nil)
							},
						}}
					},
					Consumer: rule.DiagnosticConsumer{Demand: demand, Report: func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) }},
				})
				if len(diagnostics) != 1 {
					t.Fatalf("demand %d: got %d diagnostics", demand, len(diagnostics))
				}
				diagnostic := diagnostics[0]
				if demand == rule.EditDemandAll {
					all = diagnostic
				}
				if diagnostic.Range != all.Range || !reflect.DeepEqual(diagnostic.Message, all.Message) {
					t.Fatalf("demand %d changed the diagnostic", demand)
				}
				if diagnostic.FixesPtr != nil {
					t.Fatalf("demand %d produced an autofix", demand)
				}
				if len(test.parameters) == 0 || demand == rule.EditDemandNone || demand == rule.EditDemandAutofix {
					if diagnostic.Suggestions != nil {
						t.Fatalf("demand %d produced suggestions without demand", demand)
					}
					continue
				}
				if diagnostic.Suggestions == nil || len(*diagnostic.Suggestions) != 3 || !reflect.DeepEqual(diagnostic.Suggestions, all.Suggestions) {
					t.Fatalf("demand %d changed suggestions", demand)
				}
				for i, parameters := range test.parameters {
					suggestion := (*diagnostic.Suggestions)[i]
					messageID, description := "replace-without-name", fmt.Sprintf("Replace function with `… => …(%s)`.", parameters)
					if test.callback == "fn" {
						messageID, description = "replace-with-name", fmt.Sprintf("Replace function `fn` with `… => fn(%s)`.", parameters)
					}
					if suggestion.Message.Id != messageID || suggestion.Message.Description != description {
						t.Fatalf("wrong suggestion message: %#v", suggestion.Message)
					}
					replacement := "(" + parameters + ") => " + test.callback + "(" + parameters + ")"
					if test.method == "forEach" {
						replacement = "(" + parameters + ") => { " + test.callback + "(" + parameters + "); }"
					}
					want := strings.Replace(test.code, "("+test.callback+")", "("+replacement+")", 1)
					got, _, fixed := linter.ApplyRuleFixes(test.code, []rule.RuleSuggestion{suggestion})
					if !fixed || got != want {
						t.Fatalf("suggestion output %q, want %q", got, want)
					}
				}
			}
		})
	}
}

func TestNoArrayCallbackReferenceWithoutTypeChecker(t *testing.T) {
	for _, test := range []struct {
		code string
		want int
	}{
		{"declare const array: string[]; array.map(callback);", 1},
		{"declare const array: string[] | null; array.map(callback);", 1},
		{"declare const array: string[] | Set<string>; array.forEach(callback);", 0},
		{"class Rows extends Array<string> {} const array = new Rows(); array.map(callback);", 1},
		{"class Uint8Array { map(fn: Function) {} } new Uint8Array().map(callback);", 0},
		{"const value = {} as const; array.map(value);", 0},
		{"function guard(x: unknown): x is string { return true } array.filter(guard);", 0},
	} {
		t.Run(test.code, func(t *testing.T) {
			helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
			program, file, err := helper.CreateTestProgram(test.code, "syntax.ts", "tsconfig.json")
			if err != nil {
				t.Fatal(err)
			}
			count := 0
			linter.LintSingleFile(linter.LintSingleFileOptions{
				Program: lintprogram.NewFromCompiler(program), File: file.FileName(),
				GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
					return []rule.ConfiguredRule{{Name: no_array_callback_reference.NoArrayCallbackReferenceRule.Name, Severity: rule.SeverityError,
						Run: func(ctx rule.RuleContext) rule.RuleListeners {
							ctx.TypeChecker = nil
							return no_array_callback_reference.NoArrayCallbackReferenceRule.Run(ctx, nil)
						},
					}}
				},
				Consumer: rule.DiagnosticConsumer{Demand: rule.EditDemandNone, Report: func(d rule.RuleDiagnostic) { count++ }},
			})
			if count != test.want {
				t.Fatalf("got %d diagnostics, want %d", count, test.want)
			}
		})
	}
}
