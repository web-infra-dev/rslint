package no_underscore_dangle

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestNoUnderscoreDangleUpstream migrates the full valid/invalid suite from
// upstream tests/lib/rules/no-underscore-dangle.js (eslint v10.8.1) 1:1.
// Upstream's per-case `languageOptions.ecmaVersion` is dropped: tsgo always
// parses the newest syntax, and this rule reads no globals, so the setting has
// no effect here. Position assertions cover line/column for every invalid
// case. rslint-specific lock-in cases live in
// no_underscore_dangle_extras_test.go.
func TestNoUnderscoreDangleUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUnderscoreDangleRule,
		[]rule_tester.ValidTestCase{
			{Code: `var foo_bar = 1;`},
			{Code: `function foo_bar() {}`},
			{Code: `foo.bar.__proto__;`},
			{Code: `console.log(__filename); console.log(__dirname);`},
			{Code: `var _ = require('underscore');`},
			{Code: `var a = b._;`},
			{Code: `function foo(_bar) {}`},
			{Code: `function foo(bar_) {}`},
			{Code: `(function _foo() {})`},
			{Code: `function foo(_bar) {}`, Options: map[string]any{}},
			{Code: `function foo( _bar = 0) {}`},
			{Code: `const foo = { onClick(_bar) { } }`},
			{Code: `const foo = { onClick(_bar = 0) { } }`},
			{Code: `const foo = (_bar) => {}`},
			{Code: `const foo = (_bar = 0) => {}`},
			{Code: `function foo( ..._bar) {}`},
			{Code: `const foo = (..._bar) => {}`},
			{Code: `const foo = { onClick(..._bar) { } }`},
			{Code: `export default function() {}`},
			{Code: `var _foo = 1`, Options: map[string]any{"allow": []any{"_foo"}}},
			{Code: `var __proto__ = 1;`, Options: map[string]any{"allow": []any{"__proto__"}}},
			{Code: `foo._bar;`, Options: map[string]any{"allow": []any{"_bar"}}},
			{Code: `function _foo() {}`, Options: map[string]any{"allow": []any{"_foo"}}},
			{Code: `this._bar;`, Options: map[string]any{"allowAfterThis": true}},
			{Code: `class foo { constructor() { super._bar; } }`, Options: map[string]any{"allowAfterSuper": true}},
			{Code: `class foo { _onClick() { } }`},
			{Code: `class foo { onClick_() { } }`},
			{Code: `const o = { _onClick() { } }`},
			{Code: `const o = { onClick_() { } }`},
			{Code: `const o = { _onClick() { } }`, Options: map[string]any{"allow": []any{"_onClick"}, "enforceInMethodNames": true}},
			{Code: `const o = { _foo: 'bar' }`},
			{Code: `const o = { foo_: 'bar' }`},
			{Code: `this.constructor._bar`, Options: map[string]any{"allowAfterThisConstructor": true}},
			{Code: `const foo = { onClick(bar) { } }`},
			{Code: `const foo = (bar) => {}`},
			{Code: `function foo(_bar) {}`, Options: map[string]any{"allowFunctionParams": true}},
			{Code: `function foo( _bar = 0) {}`, Options: map[string]any{"allowFunctionParams": true}},
			{Code: `const foo = { onClick(_bar) { } }`, Options: map[string]any{"allowFunctionParams": true}},
			{Code: `const foo = (_bar) => {}`, Options: map[string]any{"allowFunctionParams": true}},
			{Code: `function foo(bar) {}`, Options: map[string]any{"allowFunctionParams": false}},
			{Code: `const foo = { onClick(bar) { } }`, Options: map[string]any{"allowFunctionParams": false}},
			{Code: `const foo = (bar) => {}`, Options: map[string]any{"allowFunctionParams": false}},
			{Code: `function foo(_bar) {}`, Options: map[string]any{"allowFunctionParams": false, "allow": []any{"_bar"}}},
			{Code: `const foo = { onClick(_bar) { } }`, Options: map[string]any{"allowFunctionParams": false, "allow": []any{"_bar"}}},
			{Code: `const foo = (_bar) => {}`, Options: map[string]any{"allowFunctionParams": false, "allow": []any{"_bar"}}},
			{Code: `function foo([_bar]) {}`, Options: map[string]any{"allowFunctionParams": false}},
			{Code: `function foo([_bar] = []) {}`, Options: map[string]any{"allowFunctionParams": false}},
			{Code: `function foo( { _bar }) {}`, Options: map[string]any{"allowFunctionParams": false}},
			{Code: `function foo( { _bar = 0 } = {}) {}`, Options: map[string]any{"allowFunctionParams": false}},
			{Code: `function foo(...[_bar]) {}`, Options: map[string]any{"allowFunctionParams": false}},
			{Code: `const [_foo] = arr`},
			{Code: `const [_foo] = arr`, Options: map[string]any{}},
			{Code: `const [_foo] = arr`, Options: map[string]any{"allowInArrayDestructuring": true}},
			{Code: `const [foo, ...rest] = [1, 2, 3]`, Options: map[string]any{"allowInArrayDestructuring": false}},
			{Code: `const [foo, _bar] = [1, 2, 3]`, Options: map[string]any{"allowInArrayDestructuring": false, "allow": []any{"_bar"}}},
			{Code: `const { _foo } = obj`},
			{Code: `const { _foo } = obj`, Options: map[string]any{}},
			{Code: `const { _foo } = obj`, Options: map[string]any{"allowInObjectDestructuring": true}},
			{Code: `const { foo, bar: _bar } = { foo: 1, bar: 2 }`, Options: map[string]any{"allowInObjectDestructuring": false, "allow": []any{"_bar"}}},
			{Code: `const { foo, _bar } = { foo: 1, _bar: 2 }`, Options: map[string]any{"allowInObjectDestructuring": false, "allow": []any{"_bar"}}},
			{Code: `const { foo, _bar: bar } = { foo: 1, _bar: 2 }`, Options: map[string]any{"allowInObjectDestructuring": false}},
			{Code: `class foo { _field; }`},
			{Code: `class foo { _field; }`, Options: map[string]any{"enforceInClassFields": false}},
			{Code: `class foo { #_field; }`},
			{Code: `class foo { #_field; }`, Options: map[string]any{"enforceInClassFields": false}},
			{Code: `class foo { _field; }`, Options: map[string]any{}},

			// ---- Import attribute keys ----
			{Code: `import foo from 'foo.json' with { _type: 'json' }`},
			{Code: `export * from 'foo.json' with { _type: 'json' }`},
			{Code: `export { default } from 'foo.json' with { _type: 'json' }`},
			{Code: `import('foo.json', { _with: { _type: 'json' } })`},
			{Code: `import('foo.json', { 'with': { _type: 'json' } })`},
			{Code: `import('foo.json', { _with: { _type } })`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `var _foo = 1`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_foo'.", Line: 1, Column: 5, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code: `var foo_ = 1`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in 'foo_'.", Line: 1, Column: 5, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code: `function _foo() {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_foo'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 19},
				},
			},
			{
				Code: `function foo_() {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in 'foo_'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 19},
				},
			},
			{
				Code: `var __proto__ = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '__proto__'.", Line: 1, Column: 5, EndLine: 1, EndColumn: 18},
				},
			},
			{
				Code: `foo._bar;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_bar'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 9},
				},
			},
			{
				Code: `this._prop;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_prop'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 11},
				},
			},
			{
				Code: `class foo { constructor() { super._prop; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_prop'.", Line: 1, Column: 29, EndLine: 1, EndColumn: 40},
				},
			},
			{
				Code:    `class foo { constructor() { this._prop; } }`,
				Options: map[string]any{"allowAfterSuper": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_prop'.", Line: 1, Column: 29, EndLine: 1, EndColumn: 39},
				},
			},
			{
				Code:    `class foo { _onClick() { } }`,
				Options: map[string]any{"enforceInMethodNames": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_onClick'.", Line: 1, Column: 13, EndLine: 1, EndColumn: 27},
				},
			},
			{
				Code:    `class foo { onClick_() { } }`,
				Options: map[string]any{"enforceInMethodNames": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in 'onClick_'.", Line: 1, Column: 13, EndLine: 1, EndColumn: 27},
				},
			},
			{
				Code:    `const o = { _onClick() { } }`,
				Options: map[string]any{"enforceInMethodNames": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_onClick'.", Line: 1, Column: 13, EndLine: 1, EndColumn: 27},
				},
			},
			{
				Code:    `const o = { onClick_() { } }`,
				Options: map[string]any{"enforceInMethodNames": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in 'onClick_'.", Line: 1, Column: 13, EndLine: 1, EndColumn: 27},
				},
			},
			{
				Code: `this.constructor._bar`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_bar'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 22},
				},
			},
			{
				Code:    `function foo(_bar) {}`,
				Options: map[string]any{"allowFunctionParams": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_bar'.", Line: 1, Column: 14, EndLine: 1, EndColumn: 18},
				},
			},
			{
				Code:    `(function foo(_bar) {})`,
				Options: map[string]any{"allowFunctionParams": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_bar'.", Line: 1, Column: 15, EndLine: 1, EndColumn: 19},
				},
			},
			{
				Code:    `function foo(bar, _foo) {}`,
				Options: map[string]any{"allowFunctionParams": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_foo'.", Line: 1, Column: 19, EndLine: 1, EndColumn: 23},
				},
			},
			{
				Code:    `const foo = { onClick(_bar) { } }`,
				Options: map[string]any{"allowFunctionParams": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_bar'.", Line: 1, Column: 23, EndLine: 1, EndColumn: 27},
				},
			},
			{
				Code:    `const foo = (_bar) => {}`,
				Options: map[string]any{"allowFunctionParams": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_bar'.", Line: 1, Column: 14, EndLine: 1, EndColumn: 18},
				},
			},
			{
				Code:    `function foo(_bar = 0) {}`,
				Options: map[string]any{"allowFunctionParams": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_bar'.", Line: 1, Column: 14, EndLine: 1, EndColumn: 22},
				},
			},
			{
				Code:    `const foo = { onClick(_bar = 0) { } }`,
				Options: map[string]any{"allowFunctionParams": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_bar'.", Line: 1, Column: 23, EndLine: 1, EndColumn: 31},
				},
			},
			{
				Code:    `const foo = (_bar = 0) => {}`,
				Options: map[string]any{"allowFunctionParams": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_bar'.", Line: 1, Column: 14, EndLine: 1, EndColumn: 22},
				},
			},
			{
				Code:    `function foo(..._bar) {}`,
				Options: map[string]any{"allowFunctionParams": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_bar'.", Line: 1, Column: 14, EndLine: 1, EndColumn: 21},
				},
			},
			{
				Code:    `const foo = { onClick(..._bar) { } }`,
				Options: map[string]any{"allowFunctionParams": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_bar'.", Line: 1, Column: 23, EndLine: 1, EndColumn: 30},
				},
			},
			{
				Code:    `const foo = (..._bar) => {}`,
				Options: map[string]any{"allowFunctionParams": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_bar'.", Line: 1, Column: 14, EndLine: 1, EndColumn: 21},
				},
			},
			{
				Code:    `const [foo, _bar] = [1, 2]`,
				Options: map[string]any{"allowInArrayDestructuring": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_bar'.", Line: 1, Column: 7, EndLine: 1, EndColumn: 27},
				},
			},
			{
				Code:    `const [_foo = 1] = arr`,
				Options: map[string]any{"allowInArrayDestructuring": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_foo'.", Line: 1, Column: 7, EndLine: 1, EndColumn: 23},
				},
			},
			{
				Code:    `const [foo, ..._rest] = [1, 2, 3]`,
				Options: map[string]any{"allowInArrayDestructuring": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_rest'.", Line: 1, Column: 7, EndLine: 1, EndColumn: 34},
				},
			},
			{
				Code:    `const [foo, [bar_, baz]] = [1, [2, 3]]`,
				Options: map[string]any{"allowInArrayDestructuring": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in 'bar_'.", Line: 1, Column: 7, EndLine: 1, EndColumn: 39},
				},
			},
			{
				Code:    `const { _foo, bar } = { _foo: 1, bar: 2 }`,
				Options: map[string]any{"allowInObjectDestructuring": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_foo'.", Line: 1, Column: 7, EndLine: 1, EndColumn: 42},
				},
			},
			{
				Code:    `const { _foo = 1 } = obj`,
				Options: map[string]any{"allowInObjectDestructuring": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_foo'.", Line: 1, Column: 7, EndLine: 1, EndColumn: 25},
				},
			},
			{
				Code:    `const { bar: _foo = 1 } = obj`,
				Options: map[string]any{"allowInObjectDestructuring": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_foo'.", Line: 1, Column: 7, EndLine: 1, EndColumn: 30},
				},
			},
			{
				Code:    `const { foo: _foo, bar } = { foo: 1, bar: 2 }`,
				Options: map[string]any{"allowInObjectDestructuring": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_foo'.", Line: 1, Column: 7, EndLine: 1, EndColumn: 46},
				},
			},
			{
				Code:    `const { foo, ..._rest} = { foo: 1, bar: 2, baz: 3 }`,
				Options: map[string]any{"allowInObjectDestructuring": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_rest'.", Line: 1, Column: 7, EndLine: 1, EndColumn: 52},
				},
			},
			{
				Code:    `const { foo: [_bar, { a: _a, b } ] } = { foo: [1, { a: 'a', b: 'b' }] }`,
				Options: map[string]any{"allowInArrayDestructuring": false, "allowInObjectDestructuring": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_bar'.", Line: 1, Column: 7, EndLine: 1, EndColumn: 72},
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_a'.", Line: 1, Column: 7, EndLine: 1, EndColumn: 72},
				},
			},
			{
				Code:    `const { foo: [_bar, { a: _a, b } ] } = { foo: [1, { a: 'a', b: 'b' }] }`,
				Options: map[string]any{"allowInArrayDestructuring": true, "allowInObjectDestructuring": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_a'.", Line: 1, Column: 7, EndLine: 1, EndColumn: 72},
				},
			},
			{
				Code:    `const [{ foo: [_bar, _, { bar: _baz }] }] = [{ foo: [1, 2, { bar: 'a' }] }]`,
				Options: map[string]any{"allowInArrayDestructuring": false, "allowInObjectDestructuring": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_bar'.", Line: 1, Column: 7, EndLine: 1, EndColumn: 76},
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_baz'.", Line: 1, Column: 7, EndLine: 1, EndColumn: 76},
				},
			},
			{
				Code:    `const { foo, bar: { baz, _qux } } = { foo: 1, bar: { baz: 3, _qux: 4 } }`,
				Options: map[string]any{"allowInObjectDestructuring": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_qux'.", Line: 1, Column: 7, EndLine: 1, EndColumn: 73},
				},
			},
			{
				Code:    `class foo { #_bar() {} }`,
				Options: map[string]any{"enforceInMethodNames": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '#_bar'.", Line: 1, Column: 13, EndLine: 1, EndColumn: 23},
				},
			},
			{
				Code:    `class foo { #bar_() {} }`,
				Options: map[string]any{"enforceInMethodNames": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '#bar_'.", Line: 1, Column: 13, EndLine: 1, EndColumn: 23},
				},
			},
			{
				Code:    `class foo { _field; }`,
				Options: map[string]any{"enforceInClassFields": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_field'.", Line: 1, Column: 13, EndLine: 1, EndColumn: 20},
				},
			},
			{
				Code:    `class foo { #_field; }`,
				Options: map[string]any{"enforceInClassFields": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '#_field'.", Line: 1, Column: 13, EndLine: 1, EndColumn: 21},
				},
			},
			{
				Code:    `class foo { field_; }`,
				Options: map[string]any{"enforceInClassFields": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in 'field_'.", Line: 1, Column: 13, EndLine: 1, EndColumn: 20},
				},
			},
			{
				Code:    `class foo { #field_; }`,
				Options: map[string]any{"enforceInClassFields": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '#field_'.", Line: 1, Column: 13, EndLine: 1, EndColumn: 21},
				},
			},
		},
	)
}
