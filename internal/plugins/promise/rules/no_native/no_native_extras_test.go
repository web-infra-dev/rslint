// Additional AST, scope and regression coverage for no-native.
// The pinned upstream migration lives in no_native_upstream_test.go.
package no_native_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/promise/rules/no_native"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// N/A: no fixes, suggestions or options; literal values are never compared.
func TestNoNativeExtras(t *testing.T) {
	rule_tester.RunRuleTester(noNativeTestRoot(), "tsconfig.allowJs.json", t, &no_native.NoNativeRule,
		[]rule_tester.ValidTestCase{
			// Locks in upstream isDeclared() arm 1: unrelated names do not declare Promise
			{
				Code: `const OtherPromise = library; OtherPromise.resolve();`,
			},
			// Locks in upstream isDeclared() arm 3: a script definition allows references
			{
				Code:            `Promise.resolve(); var Promise = polyfill;`,
				LanguageOptions: rule.LanguageOptions{SourceType: "script"},
			},
			// Locks in upstream Program:exit: non-Promise globals and unresolved names are ignored
			{
				Code: `Map; Unknown; window.Promise; globalThis.Promise;`,
			},
			// Dimension 4: member names, keys and private names are not references
			{
				Code: `const object = { Promise: 1, "Promise": 2, 0: 3, Promise() {} }; object.Promise; object["Promise"]; class C { #Promise; Promise() {} }`,
			},
			// Dimension 4: JSX attributes are not references
			{
				Code:     `const x = <div Promise="yes" />;`,
				FileName: "file.tsx",
			},
			// Dimension 4: local class declaration
			{
				Code: `class Promise {} new Promise();`,
			},
			// Dimension 4: function forms and async/generator variants
			{
				Code: `function Promise() {} Promise(); (function f(Promise) { return Promise; }); ((Promise) => Promise); (async function* f(Promise) { yield Promise; });`,
			},
			// Dimension 4: parameter defaults can reference a parameter binding
			{
				Code: `function f(Promise, value = Promise.resolve()) { return value; }`,
			},
			// Dimension 4: destructuring and rest bindings
			{
				Code: `const { Promise } = library; function f(...Promise) { return Promise; } Promise.resolve();`,
			},
			// Script global definitions suppress references across namespaces
			{
				Code:            `interface Promise<T> {} Promise.resolve();`,
				LanguageOptions: rule.LanguageOptions{SourceType: "script"},
			},
			// Script value definition also permits type references
			{
				Code:            `const Promise = polyfill; let p: Promise<void>;`,
				LanguageOptions: rule.LanguageOptions{SourceType: "script"},
			},
			// CommonJS TypeScript retains the parser global scope
			{
				Code:            `type Promise = {}; Promise.resolve();`,
				LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
			},
			// CommonJS local declarations
			{
				Code:            `const Promise = require("bluebird"); module.exports = Promise;`,
				FileName:        "file.cjs",
				LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
			},
			// Type-only imports follow the upstream scope manager
			{
				Code: `import type Promise from "bluebird"; Promise.resolve();`,
			},
			// Named and namespace imports bind locally
			{
				Code: `import { Promise } from "library"; Promise.resolve();`,
			},
			// Namespace import
			{
				Code: `import * as Promise from "library"; Promise.resolve();`,
			},
			// Re-export labels are not references
			{
				Code: `export { Promise } from "library"; export * as Promise from "library";`,
			},
			// Ambient declarations authored in this file
			{
				Code: `declare const Promise: any; Promise.resolve();`,
			},
			// Real-user: issue #2 explicit Bluebird import
			{
				Code: `import Promise from "bluebird"; export default Promise.resolve(null);`,
			},
			// Disabled globals still permit local declarations
			{
				Code:    `const Promise = polyfill; Promise.resolve();`,
				Globals: map[string]any{"Promise": "off"},
			},
			// ES5 script var binding
			{
				Code:            `var Promise = polyfill; Promise.resolve();`,
				FileName:        "file.js",
				LanguageOptions: rule.LanguageOptions{SourceType: "script", ECMAVersion: 5},
			},
			// Declaration hoisting
			{
				Code: `Promise.resolve(); const Promise = polyfill;`,
			},
			// Labels are not references
			{
				Code: `Promise: for (;;) { break Promise; }`,
			},
			// Default parameters preserve their outer binding
			{
				Code: `const Promise = polyfill; function f(value = Promise) { const Promise = local; }`,
			},
		},
		[]rule_tester.InvalidTestCase{
			// Locks in upstream isDeclared() arm 2: an implicit global has no definition
			{
				Code:    `Promise;`,
				Globals: map[string]any{"Promise": "readonly"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "name", Message: noNativeMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 8}},
			},
			// Locks in upstream validatePromiseReference(): unresolved reads and writes report
			{
				Code: `Promise; typeof Promise; Promise = polyfill;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "name", Message: noNativeMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 8},
					{MessageId: "name", Message: noNativeMessage, Line: 1, Column: 17, EndLine: 1, EndColumn: 24},
					{MessageId: "name", Message: noNativeMessage, Line: 1, Column: 26, EndLine: 1, EndColumn: 33}},
			},
			// Dimension 4: parenthesized receivers and constructor
			{
				Code: `(Promise).resolve(); ((Promise)).reject(); new ((Promise))();`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "name", Message: noNativeMessage, Line: 1, Column: 2, EndLine: 1, EndColumn: 9},
					{MessageId: "name", Message: noNativeMessage, Line: 1, Column: 24, EndLine: 1, EndColumn: 31},
					{MessageId: "name", Message: noNativeMessage, Line: 1, Column: 50, EndLine: 1, EndColumn: 57}},
			},
			// Dimension 4: TS wrappers still reference the identifier
			{
				Code: `Promise!.resolve(); (Promise as any).resolve(); (Promise satisfies object).resolve();`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "name", Message: noNativeMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 8},
					{MessageId: "name", Message: noNativeMessage, Line: 1, Column: 22, EndLine: 1, EndColumn: 29},
					{MessageId: "name", Message: noNativeMessage, Line: 1, Column: 50, EndLine: 1, EndColumn: 57}},
			},
			// Dimension 4: optional property access and calls
			{
				Code: `Promise?.resolve(); Promise?.(); Promise.resolve?.();`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "name", Message: noNativeMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 8},
					{MessageId: "name", Message: noNativeMessage, Line: 1, Column: 21, EndLine: 1, EndColumn: 28},
					{MessageId: "name", Message: noNativeMessage, Line: 1, Column: 34, EndLine: 1, EndColumn: 41}},
			},
			// Dimension 4: computed member access
			{
				Code: "Promise[\"resolve\"](); Promise[`resolve`](); Promise[0]; Promise[Symbol.iterator];",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "name", Message: noNativeMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 8},
					{MessageId: "name", Message: noNativeMessage, Line: 1, Column: 23, EndLine: 1, EndColumn: 30},
					{MessageId: "name", Message: noNativeMessage, Line: 1, Column: 45, EndLine: 1, EndColumn: 52},
					{MessageId: "name", Message: noNativeMessage, Line: 1, Column: 57, EndLine: 1, EndColumn: 64}},
			},
			// Dimension 4: shorthand and computed keys reference Promise
			{
				Code: `const object = { Promise, [Promise]: 1 };`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "name", Message: noNativeMessage, Line: 1, Column: 18, EndLine: 1, EndColumn: 25},
					{MessageId: "name", Message: noNativeMessage, Line: 1, Column: 28, EndLine: 1, EndColumn: 35}},
			},
			// Dimension 4: JSX references its opening tag
			{
				Code:     `const x = <Promise></Promise>;`,
				FileName: "file.jsx",
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "name", Message: noNativeMessage, Line: 1, Column: 12, EndLine: 1, EndColumn: 19}},
			},
			// Dimension 4: TypeScript JSX references both tags
			{
				Code:     `const x = <Promise></Promise>;`,
				FileName: "file.tsx",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "name", Message: noNativeMessage, Line: 1, Column: 12, EndLine: 1, EndColumn: 19},
					{MessageId: "name", Message: noNativeMessage, Line: 1, Column: 22, EndLine: 1, EndColumn: 29}},
			},
			// Dimension 4: dotted JSX tags
			{
				Code:     `const x = <Promise.Component value={Promise} />;`,
				FileName: "file.tsx",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "name", Message: noNativeMessage, Line: 1, Column: 12, EndLine: 1, EndColumn: 19},
					{MessageId: "name", Message: noNativeMessage, Line: 1, Column: 37, EndLine: 1, EndColumn: 44}},
			},
			// Dimension 4: class expressions do not leak their name
			{
				Code:   `const C = class Promise { method() { return Promise; } }; Promise;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "name", Message: noNativeMessage, Line: 1, Column: 59, EndLine: 1, EndColumn: 66}},
			},
			// Dimension 4: function expressions do not leak their name
			{
				Code:   `const f = function Promise() { return Promise; }; Promise;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "name", Message: noNativeMessage, Line: 1, Column: 51, EndLine: 1, EndColumn: 58}},
			},
			// Dimension 4: method, field arrow and static block boundaries
			{
				Code: `class C { method(Promise) { return Promise; } field = () => Promise; static { const Promise = polyfill; Promise.resolve(); } } Promise;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "name", Message: noNativeMessage, Line: 1, Column: 61, EndLine: 1, EndColumn: 68},
					{MessageId: "name", Message: noNativeMessage, Line: 1, Column: 128, EndLine: 1, EndColumn: 135}},
			},
			// Dimension 4: nested parameter and block scopes
			{
				Code:   `function outer(Promise) { return () => Promise.resolve(); } { let Promise = polyfill; Promise.resolve(); } Promise.resolve();`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "name", Message: noNativeMessage, Line: 1, Column: 108, EndLine: 1, EndColumn: 115}},
			},
			// Dimension 4: body bindings cannot reach a parameter initializer
			{
				Code:   `function f(value = Promise.resolve()) { var Promise = polyfill; return Promise; }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "name", Message: noNativeMessage, Line: 1, Column: 20, EndLine: 1, EndColumn: 27}},
			},
			// Dimension 4: catch scope
			{
				Code:   `try {} catch (Promise) { Promise.resolve(); } Promise;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "name", Message: noNativeMessage, Line: 1, Column: 47, EndLine: 1, EndColumn: 54}},
			},
			// Dimension 4: object spread and destructuring defaults
			{
				Code: `const object = { ...Promise, value: Promise }; const { value = Promise } = object;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "name", Message: noNativeMessage, Line: 1, Column: 21, EndLine: 1, EndColumn: 28},
					{MessageId: "name", Message: noNativeMessage, Line: 1, Column: 37, EndLine: 1, EndColumn: 44},
					{MessageId: "name", Message: noNativeMessage, Line: 1, Column: 64, EndLine: 1, EndColumn: 71}},
			},
			// Dimension 4: destructuring writes do not declare Promise
			{
				Code: `({ Promise } = library); [Promise] = values;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "name", Message: noNativeMessage, Line: 1, Column: 4, EndLine: 1, EndColumn: 11},
					{MessageId: "name", Message: noNativeMessage, Line: 1, Column: 27, EndLine: 1, EndColumn: 34}},
			},
			// Dimension 4: empty containers and no arguments
			{
				Code:   `class C {} function f() {} const {} = {}; new Promise();`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "name", Message: noNativeMessage, Line: 1, Column: 47, EndLine: 1, EndColumn: 54}},
			},
			// Dimension 4: heritage references
			{
				Code: `class Child extends Promise {} class Other implements Promise<void> {}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "name", Message: noNativeMessage, Line: 1, Column: 21, EndLine: 1, EndColumn: 28},
					{MessageId: "name", Message: noNativeMessage, Line: 1, Column: 55, EndLine: 1, EndColumn: 62}},
			},
			// Dimension 4: type references also require an import or declaration
			{
				Code: `let value: Promise<void>; type Constructor = typeof Promise;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "name", Message: noNativeMessage, Line: 1, Column: 12, EndLine: 1, EndColumn: 19},
					{MessageId: "name", Message: noNativeMessage, Line: 1, Column: 53, EndLine: 1, EndColumn: 60}},
			},
			// Dimension 4: type parameters resolve only in type space
			{
				Code:   `function f<Promise>(value: Promise) { return Promise.resolve(); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "name", Message: noNativeMessage, Line: 1, Column: 46, EndLine: 1, EndColumn: 53}},
			},
			// Dimension 4: type declarations do not define a module value
			{
				Code:   `interface Promise<T> {} let p: Promise<void>; Promise.resolve();`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "name", Message: noNativeMessage, Line: 1, Column: 47, EndLine: 1, EndColumn: 54}},
			},
			// Dimension 4: type alias does not define a module value
			{
				Code:   `type Promise = {}; Promise.resolve();`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "name", Message: noNativeMessage, Line: 1, Column: 20, EndLine: 1, EndColumn: 27}},
			},
			// Dimension 4: value declaration does not define a module type
			{
				Code:   `const Promise = polyfill; let p: Promise<void>;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "name", Message: noNativeMessage, Line: 1, Column: 34, EndLine: 1, EndColumn: 41}},
			},
			// CommonJS native references
			{
				Code:            `module.exports = Promise.resolve();`,
				FileName:        "file.cjs",
				LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
				Errors:          []rule_tester.InvalidTestCaseError{{MessageId: "name", Message: noNativeMessage, Line: 1, Column: 18, EndLine: 1, EndColumn: 25}},
			},
			// Aliased import does not bind Promise
			{
				Code:   `import { Promise as Other } from "library"; Promise.resolve();`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "name", Message: noNativeMessage, Line: 1, Column: 45, EndLine: 1, EndColumn: 52}},
			},
			// Local export and type export reference the binding
			{
				Code: `export { Promise }; export type { Promise as P };`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "name", Message: noNativeMessage, Line: 1, Column: 10, EndLine: 1, EndColumn: 17},
					{MessageId: "name", Message: noNativeMessage, Line: 1, Column: 35, EndLine: 1, EndColumn: 42}},
			},
			// Global augmentation does not supply a local definition
			{
				Code:   `export {}; declare global { var Promise: any; } Promise.resolve();`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "name", Message: noNativeMessage, Line: 1, Column: 49, EndLine: 1, EndColumn: 56}},
			},
			// JSDoc type declarations do not bind runtime references
			{
				Code: `/** @typedef {object} Promise */
Promise.resolve();`,
				FileName: "file.js",
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "name", Message: noNativeMessage, Line: 2, Column: 1, EndLine: 2, EndColumn: 8}},
			},
			// JSDoc cast wrappers preserve runtime references
			{
				Code:     `(/** @type {any} */ (Promise)).resolve();`,
				FileName: "file.js",
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "name", Message: noNativeMessage, Line: 1, Column: 22, EndLine: 1, EndColumn: 29}},
			},
			// Dimension 4: multiline and UTF-16 diagnostic ranges
			{
				Code: `const label = "😀"; Promise.resolve();
function f() {
  return (
    Promise
  );
}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "name", Message: noNativeMessage, Line: 1, Column: 21, EndLine: 1, EndColumn: 28},
					{MessageId: "name", Message: noNativeMessage, Line: 4, Column: 5, EndLine: 4, EndColumn: 12}},
			},
			// Escaped identifier spelling uses the original range
			{
				Code:   `\u0050\u0072\u006f\u006d\u0069\u0073\u0065.resolve();`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "name", Message: noNativeMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 43}},
			},
			// Real-user: issue #2 missing a Promise import
			{
				Code:   `export default Promise.resolve(null);`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "name", Message: noNativeMessage, Line: 1, Column: 16, EndLine: 1, EndColumn: 23}},
			},
			// Real-user: issue #611 native globals must still report
			{
				Code:    `const ready = Promise.resolve(); ready.then(() => Promise.resolve());`,
				Globals: map[string]any{"Promise": "readonly"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "name", Message: noNativeMessage, Line: 1, Column: 15, EndLine: 1, EndColumn: 22},
					{MessageId: "name", Message: noNativeMessage, Line: 1, Column: 51, EndLine: 1, EndColumn: 58}},
			},
			// Inline globals cannot replace an authored binding
			{
				Code: `/* global Promise */
Promise.resolve();`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "name", Message: noNativeMessage, Line: 2, Column: 1, EndLine: 2, EndColumn: 8}},
			},
			// ES5 unresolved references
			{
				Code:            `new Promise(function () {});`,
				FileName:        "file.js",
				LanguageOptions: rule.LanguageOptions{SourceType: "script", ECMAVersion: 5},
				Errors:          []rule_tester.InvalidTestCaseError{{MessageId: "name", Message: noNativeMessage, Line: 1, Column: 5, EndLine: 1, EndColumn: 12}},
			},
			// Loop binding is limited to its loop
			{
				Code:   `for (const Promise of implementations) { Promise.resolve(); } Promise;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "name", Message: noNativeMessage, Line: 1, Column: 63, EndLine: 1, EndColumn: 70}},
			},
		},
	)
}
