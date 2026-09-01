// TestFuncStyleUpstream and TestFuncStyleUpstreamTypeScript migrate the full
// valid/invalid suite from ESLint v10.8.1 tests/lib/rules/func-style.js 1:1 —
// the plain-parser suite and the @typescript-eslint/parser suite upstream runs
// as two separate RuleTester instances. rslint parses every file through the
// same tsgo frontend regardless of a parser choice, so both suites run
// through one RuleTester here; upstream's languageOptions.ecmaVersion /
// sourceType are dropped since rslint's parser does not gate syntax on them.
// Upstream itself asserts only messageId on invalid cases; Line/Column/
// EndLine/EndColumn below were captured by running the real rule (ESLint
// v10.8.1) over each case. rslint-specific lock-ins live in
// func_style_extras_test.go.
package func_style

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestFuncStyleUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&FuncStyleRule,
		[]rule_tester.ValidTestCase{
			{Code: "function foo(){}\n function bar(){}", Options: []any{"declaration"}},
			{Code: "foo.bar = function(){};", Options: []any{"declaration"}},
			{Code: "(function() { /* code */ }());", Options: []any{"declaration"}},
			{Code: "var module = (function() { return {}; }());", Options: []any{"declaration"}},
			{Code: "var object = { foo: function(){} };", Options: []any{"declaration"}},
			{Code: "Array.prototype.foo = function(){};", Options: []any{"declaration"}},
			{Code: "foo.bar = function(){};", Options: []any{"expression"}},
			{Code: "var foo = function(){};\n var bar = function(){};", Options: []any{"expression"}},
			{Code: "var foo = () => {};\n var bar = () => {}", Options: []any{"expression"}},

			// https://github.com/eslint/eslint/issues/3819
			{Code: "var foo = function() { this; }.bind(this);", Options: []any{"declaration"}},
			{Code: "var foo = () => { this; };", Options: []any{"declaration"}},
			{Code: "class C extends D { foo() { var bar = () => { super.baz(); }; } }", Options: []any{"declaration"}},
			{Code: "var obj = { foo() { var bar = () => super.baz; } }", Options: []any{"declaration"}},
			{Code: "export default function () {};"},
			{Code: "var foo = () => {};", Options: []any{"declaration", map[string]any{"allowArrowFunctions": true}}},
			{Code: "var foo = () => { function foo() { this; } };", Options: []any{"declaration", map[string]any{"allowArrowFunctions": true}}},
			{Code: "var foo = () => ({ bar() { super.baz(); } });", Options: []any{"declaration", map[string]any{"allowArrowFunctions": true}}},
			{Code: "export function foo() {};", Options: []any{"declaration"}},
			{Code: "export function foo() {};", Options: []any{"expression", map[string]any{"overrides": map[string]any{"namedExports": "declaration"}}}},
			{Code: "export function foo() {};", Options: []any{"declaration", map[string]any{"overrides": map[string]any{"namedExports": "declaration"}}}},
			{Code: "export function foo() {};", Options: []any{"expression", map[string]any{"overrides": map[string]any{"namedExports": "ignore"}}}},
			{Code: "export function foo() {};", Options: []any{"declaration", map[string]any{"overrides": map[string]any{"namedExports": "ignore"}}}},
			{Code: "export var foo = function(){};", Options: []any{"expression"}},
			{Code: "export var foo = function(){};", Options: []any{"declaration", map[string]any{"overrides": map[string]any{"namedExports": "expression"}}}},
			{Code: "export var foo = function(){};", Options: []any{"expression", map[string]any{"overrides": map[string]any{"namedExports": "expression"}}}},
			{Code: "export var foo = function(){};", Options: []any{"declaration", map[string]any{"overrides": map[string]any{"namedExports": "ignore"}}}},
			{Code: "export var foo = function(){};", Options: []any{"expression", map[string]any{"overrides": map[string]any{"namedExports": "ignore"}}}},
			{Code: "export var foo = () => {};", Options: []any{"expression", map[string]any{"overrides": map[string]any{"namedExports": "expression"}}}},
			{Code: "export var foo = () => {};", Options: []any{"declaration", map[string]any{"overrides": map[string]any{"namedExports": "expression"}}}},
			{Code: "export var foo = () => {};", Options: []any{"declaration", map[string]any{"overrides": map[string]any{"namedExports": "ignore"}}}},
			{Code: "export var foo = () => {};", Options: []any{"expression", map[string]any{"overrides": map[string]any{"namedExports": "ignore"}}}},
			{Code: "export var foo = () => {};", Options: []any{"declaration", map[string]any{"allowArrowFunctions": true, "overrides": map[string]any{"namedExports": "expression"}}}},
			{Code: "export var foo = () => {};", Options: []any{"expression", map[string]any{"allowArrowFunctions": true, "overrides": map[string]any{"namedExports": "expression"}}}},
			{Code: "export var foo = () => {};", Options: []any{"declaration", map[string]any{"allowArrowFunctions": true, "overrides": map[string]any{"namedExports": "ignore"}}}},
			{Code: "$1: function $2() { }", Options: []any{"declaration"}},
			{Code: "switch ($0) { case $1: function $2() { } }", Options: []any{"declaration"}},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:    "var foo = function(){};",
				Options: []any{"declaration"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "declaration", Line: 1, Column: 5, EndLine: 1, EndColumn: 23},
				},
			},
			{
				Code:    "var foo = () => {};",
				Options: []any{"declaration"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "declaration", Line: 1, Column: 5, EndLine: 1, EndColumn: 19},
				},
			},
			{
				Code:    "var foo = () => { function foo() { this; } };",
				Options: []any{"declaration"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "declaration", Line: 1, Column: 5, EndLine: 1, EndColumn: 45},
				},
			},
			{
				Code:    "var foo = () => ({ bar() { super.baz(); } });",
				Options: []any{"declaration"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "declaration", Line: 1, Column: 5, EndLine: 1, EndColumn: 45},
				},
			},
			{
				Code:    "function foo(){}",
				Options: []any{"expression"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expression", Line: 1, Column: 1, EndLine: 1, EndColumn: 17},
				},
			},
			{
				Code:    "export function foo(){}",
				Options: []any{"expression"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expression", Line: 1, Column: 8, EndLine: 1, EndColumn: 24},
				},
			},
			{
				Code:    "export function foo() {};",
				Options: []any{"declaration", map[string]any{"overrides": map[string]any{"namedExports": "expression"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expression", Line: 1, Column: 8, EndLine: 1, EndColumn: 25},
				},
			},
			{
				Code:    "export function foo() {};",
				Options: []any{"expression", map[string]any{"overrides": map[string]any{"namedExports": "expression"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expression", Line: 1, Column: 8, EndLine: 1, EndColumn: 25},
				},
			},
			{
				Code:    "export var foo = function(){};",
				Options: []any{"declaration"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "declaration", Line: 1, Column: 12, EndLine: 1, EndColumn: 30},
				},
			},
			{
				Code:    "export var foo = function(){};",
				Options: []any{"expression", map[string]any{"overrides": map[string]any{"namedExports": "declaration"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "declaration", Line: 1, Column: 12, EndLine: 1, EndColumn: 30},
				},
			},
			{
				Code:    "export var foo = function(){};",
				Options: []any{"declaration", map[string]any{"overrides": map[string]any{"namedExports": "declaration"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "declaration", Line: 1, Column: 12, EndLine: 1, EndColumn: 30},
				},
			},
			{
				Code:    "export var foo = () => {};",
				Options: []any{"declaration"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "declaration", Line: 1, Column: 12, EndLine: 1, EndColumn: 26},
				},
			},
			{
				Code:    "export var b = () => {};",
				Options: []any{"expression", map[string]any{"overrides": map[string]any{"namedExports": "declaration"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "declaration", Line: 1, Column: 12, EndLine: 1, EndColumn: 24},
				},
			},
			{
				Code:    "export var c = () => {};",
				Options: []any{"declaration", map[string]any{"overrides": map[string]any{"namedExports": "declaration"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "declaration", Line: 1, Column: 12, EndLine: 1, EndColumn: 24},
				},
			},
			{
				Code:    "function foo() {};",
				Options: []any{"expression", map[string]any{"overrides": map[string]any{"namedExports": "declaration"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expression", Line: 1, Column: 1, EndLine: 1, EndColumn: 18},
				},
			},
			{
				Code:    "var foo = function() {};",
				Options: []any{"declaration", map[string]any{"overrides": map[string]any{"namedExports": "expression"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "declaration", Line: 1, Column: 5, EndLine: 1, EndColumn: 24},
				},
			},
			{
				Code:    "var foo = () => {};",
				Options: []any{"declaration", map[string]any{"overrides": map[string]any{"namedExports": "expression"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "declaration", Line: 1, Column: 5, EndLine: 1, EndColumn: 19},
				},
			},
			{
				Code:    "const foo = function() {};",
				Options: []any{"declaration", map[string]any{"allowTypeAnnotation": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "declaration", Line: 1, Column: 7, EndLine: 1, EndColumn: 26},
				},
			},
			{
				Code: "$1: function $2() { }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expression", Line: 1, Column: 5, EndLine: 1, EndColumn: 22},
				},
			},
			{
				Code:    "const foo = () => {};",
				Options: []any{"declaration", map[string]any{"allowTypeAnnotation": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "declaration", Line: 1, Column: 7, EndLine: 1, EndColumn: 21},
				},
			},
			{
				Code:    "export const foo = function() {};",
				Options: []any{"expression", map[string]any{"allowTypeAnnotation": true, "overrides": map[string]any{"namedExports": "declaration"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "declaration", Line: 1, Column: 14, EndLine: 1, EndColumn: 33},
				},
			},
			{
				Code:    "export const foo = () => {};",
				Options: []any{"expression", map[string]any{"allowTypeAnnotation": true, "overrides": map[string]any{"namedExports": "declaration"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "declaration", Line: 1, Column: 14, EndLine: 1, EndColumn: 28},
				},
			},
			{
				Code: "if (foo) function bar() {}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expression", Line: 1, Column: 10, EndLine: 1, EndColumn: 27},
				},
			},
		},
	)
}

func TestFuncStyleUpstreamTypeScript(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&FuncStyleRule,
		[]rule_tester.ValidTestCase{
			{Code: "function foo(): void {}\n function bar(): void {}", Options: []any{"declaration"}},
			{Code: "(function(): void { /* code */ }());", Options: []any{"declaration"}},
			{Code: "const module = (function(): { [key: string]: any } { return {}; }());", Options: []any{"declaration"}},
			{Code: "const object: { foo: () => void } = { foo: function(): void {} };", Options: []any{"declaration"}},
			{Code: "Array.prototype.foo = function(): void {};", Options: []any{"declaration"}},
			{Code: "const foo: () => void = function(): void {};\n const bar: () => void = function(): void {};", Options: []any{"expression"}},
			{Code: "const foo: () => void = (): void => {};\n const bar: () => void = (): void => {}", Options: []any{"expression"}},
			{Code: "const foo: () => void = function(): void { this; }.bind(this);", Options: []any{"declaration"}},
			{Code: "const foo: () => void = (): void => { this; };", Options: []any{"declaration"}},
			{Code: "class C extends D { foo(): void { const bar: () => void = (): void => { super.baz(); }; } }", Options: []any{"declaration"}},
			{Code: "const obj: { foo(): void } = { foo(): void { const bar: () => void = (): void => super.baz; } }", Options: []any{"declaration"}},
			{Code: "const foo: () => void = (): void => {};", Options: []any{"declaration", map[string]any{"allowArrowFunctions": true}}},
			{Code: "const foo: () => void = (): void => { function foo(): void { this; } };", Options: []any{"declaration", map[string]any{"allowArrowFunctions": true}}},
			{Code: "const foo: () => { bar(): void } = (): { bar(): void } => ({ bar(): void { super.baz(); } });", Options: []any{"declaration", map[string]any{"allowArrowFunctions": true}}},
			{Code: "export function foo(): void {};", Options: []any{"declaration"}},
			{Code: "export function foo(): void {};", Options: []any{"expression", map[string]any{"overrides": map[string]any{"namedExports": "declaration"}}}},
			{Code: "export function foo(): void {};", Options: []any{"declaration", map[string]any{"overrides": map[string]any{"namedExports": "declaration"}}}},
			{Code: "export function foo(): void {};", Options: []any{"expression", map[string]any{"overrides": map[string]any{"namedExports": "ignore"}}}},
			{Code: "export function foo(): void {};", Options: []any{"declaration", map[string]any{"overrides": map[string]any{"namedExports": "ignore"}}}},
			{Code: "export const foo: () => void = function(): void {};", Options: []any{"expression"}},
			{Code: "export const foo: () => void = function(): void {};", Options: []any{"declaration", map[string]any{"overrides": map[string]any{"namedExports": "expression"}}}},
			{Code: "export const foo: () => void = function(): void {};", Options: []any{"expression", map[string]any{"overrides": map[string]any{"namedExports": "expression"}}}},
			{Code: "export const foo: () => void = function(): void {};", Options: []any{"declaration", map[string]any{"overrides": map[string]any{"namedExports": "ignore"}}}},
			{Code: "export const foo: () => void = function(): void {};", Options: []any{"expression", map[string]any{"overrides": map[string]any{"namedExports": "ignore"}}}},
			{Code: "const expression: Fn = function () {}", Options: []any{"declaration", map[string]any{"allowTypeAnnotation": true}}},
			{Code: "const arrow: Fn = () => {}", Options: []any{"declaration", map[string]any{"allowTypeAnnotation": true}}},
			{Code: "export const expression: Fn = function () {}", Options: []any{"declaration", map[string]any{"allowTypeAnnotation": true}}},
			{Code: "export const arrow: Fn = () => {}", Options: []any{"declaration", map[string]any{"allowTypeAnnotation": true}}},
			{Code: "export const expression: Fn = function () {}", Options: []any{"expression", map[string]any{"allowTypeAnnotation": true, "overrides": map[string]any{"namedExports": "declaration"}}}},
			{Code: "export const arrow: Fn = () => {}", Options: []any{"expression", map[string]any{"allowTypeAnnotation": true, "overrides": map[string]any{"namedExports": "declaration"}}}},
			{Code: "$1: function $2(): void { }", Options: []any{"declaration"}},
			{Code: "switch ($0) { case $1: function $2(): void { } }", Options: []any{"declaration"}},
			{Code: "\n\t\tfunction test(a: string): string;\n\t\tfunction test(a: number): number;\n\t\tfunction test(a: unknown) {\n\t\t  return a;\n\t\t}\n\t\t"},
			{Code: "\n\t\texport function test(a: string): string;\n\t\texport function test(a: number): number;\n\t\texport function test(a: unknown) {\n\t\t  return a;\n\t\t}\n\t\t"},
			{
				Code:    "\n\t\t\texport function test(a: string): string;\n\t\t    export function test(a: number): number;\n\t\t    export function test(a: unknown) {\n\t\t      return a;\n\t\t    }\n\t\t\t",
				Options: []any{"expression", map[string]any{"overrides": map[string]any{"namedExports": "expression"}}},
			},
			{Code: "\n\t\tswitch ($0) {\n\t\t\tcase $1:\n\t\t\tfunction test(a: string): string;\n\t\t\tfunction test(a: number): number;\n\t\t\tfunction test(a: unknown) {\n\t\t\treturn a;\n\t\t\t}\n\t\t}\n\t\t"},
			{Code: "\n\t\tswitch ($0) {\n\t\t\tcase $1:\n\t\t\tfunction test(a: string): string;\n\t\t\tbreak;\n\t\t\tcase $2:\n\t\t\tfunction test(a: unknown) {\n\t\t\treturn a;\n\t\t\t}\n\t\t}\n\t\t"},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:    "const foo: () => void = function(): void {};",
				Options: []any{"declaration"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "declaration", Line: 1, Column: 7, EndLine: 1, EndColumn: 44},
				},
			},
			{
				Code:    "const foo: () => void = (): void => {};",
				Options: []any{"declaration"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "declaration", Line: 1, Column: 7, EndLine: 1, EndColumn: 39},
				},
			},
			{
				Code:    "const foo: () => void = (): void => { function foo(): void { this; } };",
				Options: []any{"declaration"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "declaration", Line: 1, Column: 7, EndLine: 1, EndColumn: 71},
				},
			},
			{
				Code:    "const foo: () => { bar(): void } = (): { bar(): void } => ({ bar(): void { super.baz(); } });",
				Options: []any{"declaration"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "declaration", Line: 1, Column: 7, EndLine: 1, EndColumn: 93},
				},
			},
			{
				Code:    "function foo(): void {}",
				Options: []any{"expression"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expression", Line: 1, Column: 1, EndLine: 1, EndColumn: 24},
				},
			},
			{
				Code:    "export function foo(): void {}",
				Options: []any{"expression"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expression", Line: 1, Column: 8, EndLine: 1, EndColumn: 31},
				},
			},
			{
				Code:    "export function foo(): void {};",
				Options: []any{"declaration", map[string]any{"overrides": map[string]any{"namedExports": "expression"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expression", Line: 1, Column: 8, EndLine: 1, EndColumn: 31},
				},
			},
			{
				Code:    "export function foo(): void {};",
				Options: []any{"expression", map[string]any{"overrides": map[string]any{"namedExports": "expression"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expression", Line: 1, Column: 8, EndLine: 1, EndColumn: 31},
				},
			},
			{
				Code:    "export const foo: () => void = function(): void {};",
				Options: []any{"declaration"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "declaration", Line: 1, Column: 14, EndLine: 1, EndColumn: 51},
				},
			},
			{
				Code:    "export const foo: () => void = function(): void {};",
				Options: []any{"expression", map[string]any{"overrides": map[string]any{"namedExports": "declaration"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "declaration", Line: 1, Column: 14, EndLine: 1, EndColumn: 51},
				},
			},
			{
				Code:    "export const foo: () => void = function(): void {};",
				Options: []any{"declaration", map[string]any{"overrides": map[string]any{"namedExports": "declaration"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "declaration", Line: 1, Column: 14, EndLine: 1, EndColumn: 51},
				},
			},
			{
				Code:    "export const foo: () => void = (): void => {};",
				Options: []any{"declaration"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "declaration", Line: 1, Column: 14, EndLine: 1, EndColumn: 46},
				},
			},
			{
				Code:    "export const b: () => void = (): void => {};",
				Options: []any{"expression", map[string]any{"overrides": map[string]any{"namedExports": "declaration"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "declaration", Line: 1, Column: 14, EndLine: 1, EndColumn: 44},
				},
			},
			{
				Code:    "export const c: () => void = (): void => {};",
				Options: []any{"declaration", map[string]any{"overrides": map[string]any{"namedExports": "declaration"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "declaration", Line: 1, Column: 14, EndLine: 1, EndColumn: 44},
				},
			},
			{
				Code:    "function foo(): void {};",
				Options: []any{"expression", map[string]any{"overrides": map[string]any{"namedExports": "declaration"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expression", Line: 1, Column: 1, EndLine: 1, EndColumn: 24},
				},
			},
			{
				Code:    "const foo: () => void = function(): void {};",
				Options: []any{"declaration", map[string]any{"overrides": map[string]any{"namedExports": "expression"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "declaration", Line: 1, Column: 7, EndLine: 1, EndColumn: 44},
				},
			},
			{
				Code:    "const foo: () => void = (): void => {};",
				Options: []any{"declaration", map[string]any{"overrides": map[string]any{"namedExports": "expression"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "declaration", Line: 1, Column: 7, EndLine: 1, EndColumn: 39},
				},
			},
			{
				Code: "$1: function $2(): void { }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expression", Line: 1, Column: 5, EndLine: 1, EndColumn: 28},
				},
			},
			{
				Code: "if (foo) function bar(): string {}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expression", Line: 1, Column: 10, EndLine: 1, EndColumn: 35},
				},
			},
			{
				Code: "\n\t\t\tfunction test1(a: string): string;\n\t\t\tfunction test2(a: number): number;\n\t\t\tfunction test3(a: unknown) {\n\t\t\t  return a;\n\t\t\t}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expression", Line: 4, Column: 4, EndLine: 6, EndColumn: 5},
				},
			},
			{
				Code: "\n\t\t\texport function test1(a: string): string;\n\t\t\texport function test2(a: number): number;\n\t\t\texport function test3(a: unknown) {\n\t\t\t  return a;\n\t\t\t}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expression", Line: 4, Column: 11, EndLine: 6, EndColumn: 5},
				},
			},
			{
				Code:    "\n\t\t\texport function test1(a: string): string;\n\t\t    export function test2(a: number): number;\n\t\t    export function test3(a: unknown) {\n\t\t      return a;\n\t\t    }\n\t\t\t",
				Options: []any{"expression", map[string]any{"overrides": map[string]any{"namedExports": "expression"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expression", Line: 4, Column: 14, EndLine: 6, EndColumn: 8},
				},
			},
			{
				Code: "\n\t\t\tswitch ($0) {\n\t\t\t\tcase $1:\n\t\t\t\tfunction test1(a: string): string;\n\t\t\t\tfunction test2(a: number): number;\n\t\t\t\tfunction test3(a: unknown) {\n\t\t\t\t\treturn a;\n\t\t\t\t}\n\t\t\t}\n\t\t\t",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expression", Line: 6, Column: 5, EndLine: 8, EndColumn: 6},
				},
			},
			{
				Code: "\n\t\t\tswitch ($0) {\n\t\t\t\tcase $1:\n\t\t\t\tfunction test1(a: string): string;\n\t\t\t\tbreak;\n\t\t\t\tcase $2:\n\t\t\t\tfunction test2(a: unknown) {\n\t\t\t\treturn a;\n\t\t\t\t}\n\t\t\t}\n\t\t\t",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expression", Line: 7, Column: 5, EndLine: 9, EndColumn: 6},
				},
			},
		},
	)
}
