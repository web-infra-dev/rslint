// TestFuncNameMatchingUpstream migrates the full valid/invalid suite from
// upstream eslint/tests/lib/rules/func-name-matching.js 1:1. Position
// assertions cover line/column for every invalid case. One case is migrated
// with an adjusted (not upstream-identical) expected outcome, marked inline —
// see the rule doc's "Differences from ESLint" section. rslint-specific
// lock-in cases live in func_name_matching_extras_test.go.
package func_name_matching

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestFuncNameMatchingUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&FuncNameMatchingRule,
		[]rule_tester.ValidTestCase{
			{Code: `var foo;`},
			{Code: `var foo = function foo() {};`},
			{Code: `var foo = function foo() {};`, Options: []any{"always"}},
			{Code: `var foo = function bar() {};`, Options: []any{"never"}},
			{Code: `var foo = function() {}`},
			{Code: `var foo = () => {}`, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2015}},
			{Code: `foo = function foo() {};`},
			{Code: `foo = function foo() {};`, Options: []any{"always"}},
			{Code: `foo = function bar() {};`, Options: []any{"never"}},
			{Code: `foo &&= function foo() {};`, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2021}},
			{Code: `obj.foo ||= function foo() {};`, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2021}},
			{Code: `obj['foo'] ??= function foo() {};`, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2021}},
			{Code: `obj.foo = function foo() {};`},
			{Code: `obj.foo = function foo() {};`, Options: []any{"always"}},
			{Code: `obj.foo = function bar() {};`, Options: []any{"never"}},
			{Code: `obj.foo = function() {};`},
			{Code: `obj.foo = function() {};`, Options: []any{"always"}},
			{Code: `obj.foo = function() {};`, Options: []any{"never"}},
			{Code: `obj.bar.foo = function foo() {};`},
			{Code: `obj.bar.foo = function foo() {};`, Options: []any{"always"}},
			{Code: `obj.bar.foo = function baz() {};`, Options: []any{"never"}},
			{Code: `obj['foo'] = function foo() {};`},
			{Code: `obj['foo'] = function foo() {};`, Options: []any{"always"}},
			{Code: `obj['foo'] = function bar() {};`, Options: []any{"never"}},
			{Code: `obj['foo//bar'] = function foo() {};`},
			{Code: `obj['foo//bar'] = function foo() {};`, Options: []any{"always"}},
			{Code: `obj['foo//bar'] = function foo() {};`, Options: []any{"never"}},
			{Code: `obj[foo] = function bar() {};`},
			{Code: `obj[foo] = function bar() {};`, Options: []any{"always"}},
			{Code: `obj[foo] = function bar() {};`, Options: []any{"never"}},
			{Code: `var obj = {foo: function foo() {}};`},
			{Code: `var obj = {foo: function foo() {}};`, Options: []any{"always"}},
			{Code: `var obj = {foo: function bar() {}};`, Options: []any{"never"}},
			{Code: `var obj = {'foo': function foo() {}};`},
			{Code: `var obj = {'foo': function foo() {}};`, Options: []any{"always"}},
			{Code: `var obj = {'foo': function bar() {}};`, Options: []any{"never"}},
			{Code: `var obj = {'foo//bar': function foo() {}};`},
			{Code: `var obj = {'foo//bar': function foo() {}};`, Options: []any{"always"}},
			{Code: `var obj = {'foo//bar': function foo() {}};`, Options: []any{"never"}},
			{Code: `var obj = {foo: function() {}};`},
			{Code: `var obj = {foo: function() {}};`, Options: []any{"always"}},
			{Code: `var obj = {foo: function() {}};`, Options: []any{"never"}},
			{Code: `var obj = {[foo]: function bar() {}} `, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2015}},
			{Code: `var obj = {['x' + 2]: function bar(){}};`, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2015}},
			{Code: `obj['x' + 2] = function bar(){};`},
			{Code: `var [ bar ] = [ function bar(){} ];`, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2015}},
			{Code: `function a(foo = function bar() {}) {}`, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2015}},
			{Code: `module.exports = function foo(name) {};`},
			{Code: `module['exports'] = function foo(name) {};`},
			{Code: `module.exports = function foo(name) {};`, Options: []any{map[string]any{"includeCommonJSModuleExports": false}}, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2015}},
			{Code: `module.exports = function foo(name) {};`, Options: []any{"always", map[string]any{"includeCommonJSModuleExports": false}}, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2015}},
			{Code: `module.exports = function foo(name) {};`, Options: []any{"never", map[string]any{"includeCommonJSModuleExports": false}}, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2015}},
			{Code: `module['exports'] = function foo(name) {};`, Options: []any{map[string]any{"includeCommonJSModuleExports": false}}, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2015}},
			{Code: `module['exports'] = function foo(name) {};`, Options: []any{"always", map[string]any{"includeCommonJSModuleExports": false}}, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2015}},
			{Code: `module['exports'] = function foo(name) {};`, Options: []any{"never", map[string]any{"includeCommonJSModuleExports": false}}, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2015}},
			{Code: `({['foo']: function foo() {}})`, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2015}},
			{Code: `({['foo']: function foo() {}})`, Options: []any{"always"}, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2015}},
			{Code: `({['foo']: function bar() {}})`, Options: []any{"never"}, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2015}},
			{Code: `({['❤']: function foo() {}})`, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2015}},
			{Code: `({[foo]: function bar() {}})`, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2015}},
			{Code: `({[null]: function foo() {}})`, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2015}},
			{Code: `({[1]: function foo() {}})`, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2015}},
			{Code: `({[true]: function foo() {}})`, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2015}},
			{Code: "({[`x`]: function foo() {}})", LanguageOptions: rule.LanguageOptions{ECMAVersion: 2015}},
			{Code: `({[/abc/]: function foo() {}})`, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2015}},
			{Code: `({[[1, 2, 3]]: function foo() {}})`, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2015}},
			{Code: `({[{x: 1}]: function foo() {}})`, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2015}},
			{Code: `[] = function foo() {}`, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2015}},
			{Code: `({} = function foo() {})`, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2015}},
			{Code: `[a] = function foo() {}`, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2015}},
			{Code: `({a} = function foo() {})`, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2015}},
			{Code: `var [] = function foo() {}`, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2015}},
			{Code: `var {} = function foo() {}`, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2015}},
			{Code: `var [a] = function foo() {}`, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2015}},
			{Code: `var {a} = function foo() {}`, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2015}},
			{Code: `({ value: function value() {} })`, Options: []any{map[string]any{"considerPropertyDescriptor": true}}},
			{Code: `obj.foo = function foo() {};`, Options: []any{"always", map[string]any{"considerPropertyDescriptor": true}}},
			{Code: `obj.bar.foo = function foo() {};`, Options: []any{"always", map[string]any{"considerPropertyDescriptor": true}}},
			{Code: `var obj = {foo: function foo() {}};`, Options: []any{"always", map[string]any{"considerPropertyDescriptor": true}}},
			{Code: `var obj = {foo: function() {}};`, Options: []any{"always", map[string]any{"considerPropertyDescriptor": true}}},
			{Code: `var obj = { value: function value() {} }`, Options: []any{"always", map[string]any{"considerPropertyDescriptor": true}}},
			{Code: `Object.defineProperty(foo, 'bar', { value: function bar() {} })`, Options: []any{"always", map[string]any{"considerPropertyDescriptor": true}}},
			{Code: `Object.defineProperties(foo, { bar: { value: function bar() {} } })`, Options: []any{"always", map[string]any{"considerPropertyDescriptor": true}}},
			{Code: `Object.create(proto, { bar: { value: function bar() {} } })`, Options: []any{"always", map[string]any{"considerPropertyDescriptor": true}}},
			{Code: `Object.defineProperty(foo, 'b' + 'ar', { value: function bar() {} })`, Options: []any{"always", map[string]any{"considerPropertyDescriptor": true}}},
			{Code: `Object.defineProperties(foo, { ['bar']: { value: function bar() {} } })`, Options: []any{"always", map[string]any{"considerPropertyDescriptor": true}}, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2015}},
			{Code: `Object.create(proto, { ['bar']: { value: function bar() {} } })`, Options: []any{"always", map[string]any{"considerPropertyDescriptor": true}}, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2015}},
			{Code: `Object.defineProperty(foo, 'bar', { value() {} })`, Options: []any{"never", map[string]any{"considerPropertyDescriptor": true}}, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2015}},
			{Code: `Object.defineProperties(foo, { bar: { value() {} } })`, Options: []any{"never", map[string]any{"considerPropertyDescriptor": true}}, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2015}},
			{Code: `Object.create(proto, { bar: { value() {} } })`, Options: []any{"never", map[string]any{"considerPropertyDescriptor": true}}, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2015}},
			{Code: `Reflect.defineProperty(foo, 'bar', { value: function bar() {} })`, Options: []any{"always", map[string]any{"considerPropertyDescriptor": true}}},
			{Code: `Reflect.defineProperty(foo, 'b' + 'ar', { value: function baz() {} })`, Options: []any{"always", map[string]any{"considerPropertyDescriptor": true}}},
			{Code: `Reflect.defineProperty(foo, 'bar', { value() {} })`, Options: []any{"never", map[string]any{"considerPropertyDescriptor": true}}, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2015}},
			{Code: `foo({ value: function value() {} })`, Options: []any{"always", map[string]any{"considerPropertyDescriptor": true}}},

			// ---- class fields, private names are ignored ----
			{Code: `class C { x = function () {}; }`, Options: []any{"always"}, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022}},
			{Code: `class C { x = function () {}; }`, Options: []any{"never"}, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022}},
			{Code: `class C { 'x' = function () {}; }`, Options: []any{"always"}, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022}},
			{Code: `class C { 'x' = function () {}; }`, Options: []any{"never"}, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022}},
			{Code: `class C { #x = function () {}; }`, Options: []any{"always"}, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022}},
			{Code: `class C { #x = function () {}; }`, Options: []any{"never"}, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022}},
			{Code: `class C { [x] = function () {}; }`, Options: []any{"always"}, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022}},
			{Code: `class C { [x] = function () {}; }`, Options: []any{"never"}, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022}},
			{Code: `class C { ['x'] = function () {}; }`, Options: []any{"always"}, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022}},
			{Code: `class C { ['x'] = function () {}; }`, Options: []any{"never"}, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022}},
			{Code: `class C { x = function x() {}; }`, Options: []any{"always"}, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022}},
			{Code: `class C { x = function y() {}; }`, Options: []any{"never"}, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022}},
			{Code: `class C { 'x' = function x() {}; }`, Options: []any{"always"}, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022}},
			{Code: `class C { 'x' = function y() {}; }`, Options: []any{"never"}, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022}},
			{Code: `class C { #x = function x() {}; }`, Options: []any{"always"}, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022}},
			{Code: `class C { #x = function x() {}; }`, Options: []any{"never"}, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022}},
			{Code: `class C { #x = function y() {}; }`, Options: []any{"always"}, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022}},
			{Code: `class C { #x = function y() {}; }`, Options: []any{"never"}, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022}},
			{Code: `class C { [x] = function x() {}; }`, Options: []any{"always"}, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022}},
			{Code: `class C { [x] = function x() {}; }`, Options: []any{"never"}, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022}},
			{Code: `class C { [x] = function y() {}; }`, Options: []any{"always"}, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022}},
			{Code: `class C { [x] = function y() {}; }`, Options: []any{"never"}, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022}},
			{Code: `class C { ['x'] = function x() {}; }`, Options: []any{"always"}, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022}},
			{Code: `class C { ['x'] = function y() {}; }`, Options: []any{"never"}, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022}},
			{Code: `class C { 'xy ' = function foo() {}; }`, Options: []any{"always"}, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022}},
			{Code: `class C { 'xy ' = function xy() {}; }`, Options: []any{"never"}, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022}},
			{Code: `class C { ['xy '] = function foo() {}; }`, Options: []any{"always"}, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022}},
			{Code: `class C { ['xy '] = function xy() {}; }`, Options: []any{"never"}, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022}},
			{Code: `class C { 1 = function x0() {}; }`, Options: []any{"always"}, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022}},
			{Code: `class C { 1 = function x1() {}; }`, Options: []any{"never"}, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022}},
			{Code: `class C { [1] = function x0() {}; }`, Options: []any{"always"}, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022}},
			{Code: `class C { [1] = function x1() {}; }`, Options: []any{"never"}, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022}},
			{Code: `class C { [f()] = function g() {}; }`, Options: []any{"always"}, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022}},
			{Code: `class C { [f()] = function f() {}; }`, Options: []any{"never"}, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022}},
			{Code: `class C { static x = function x() {}; }`, Options: []any{"always"}, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022}},
			{Code: `class C { static x = function y() {}; }`, Options: []any{"never"}, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022}},
			{Code: `class C { x = (function y() {})(); }`, Options: []any{"always"}, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022}},
			{Code: `class C { x = (function x() {})(); }`, Options: []any{"never"}, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022}},
			{Code: `(class { x = function x() {}; })`, Options: []any{"always"}, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022}},
			{Code: `(class { x = function y() {}; })`, Options: []any{"never"}, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022}},
			{Code: `class C { #x; foo() { this.#x = function x() {}; } }`, Options: []any{"always"}, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022}},
			{Code: `class C { #x; foo() { this.#x = function x() {}; } }`, Options: []any{"never"}, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022}},
			{Code: `class C { #x; foo() { this.#x = function y() {}; } }`, Options: []any{"always"}, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022}},
			{Code: `class C { #x; foo() { this.#x = function y() {}; } }`, Options: []any{"never"}, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022}},
			{Code: `class C { #x; foo() { a.b.#x = function x() {}; } }`, Options: []any{"always"}, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022}},
			{Code: `class C { #x; foo() { a.b.#x = function x() {}; } }`, Options: []any{"never"}, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022}},
			{Code: `class C { #x; foo() { a.b.#x = function y() {}; } }`, Options: []any{"always"}, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022}},
			{Code: `class C { #x; foo() { a.b.#x = function y() {}; } }`, Options: []any{"never"}, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022}},

			// NOTE: upstream keeps `var obj = { 'ᢅ': function foo() {} };`
			// valid at ecmaVersion 5 (esutils' frozen ES5.1/Unicode-v9 table
			// excludes U+1885 from identifier characters). rslint's identifier
			// check uses one Unicode table at every ecmaVersion and accepts
			// U+1885, so this case is migrated to the invalid list below
			// instead — see the rule doc's "Differences from ESLint" section.
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:            `let foo = function bar() {};`,
				Options:         []any{"always"},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2015},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matchVariable", Line: 1, Column: 5},
				},
			},
			{
				Code:            `let foo = function bar() {};`,
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2015},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matchVariable", Line: 1, Column: 5},
				},
			},
			{
				Code:            `foo = function bar() {};`,
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2015},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matchVariable", Line: 1, Column: 1},
				},
			},
			{
				Code:            `foo &&= function bar() {};`,
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2021},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matchVariable", Line: 1, Column: 1},
				},
			},
			{
				Code:            `obj.foo ||= function bar() {};`,
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2021},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matchProperty", Line: 1, Column: 1},
				},
			},
			{
				Code:            `obj['foo'] ??= function bar() {};`,
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2021},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matchProperty", Line: 1, Column: 1},
				},
			},
			{
				Code:            `obj.foo = function bar() {};`,
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2015},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matchProperty", Line: 1, Column: 1},
				},
			},
			{
				Code:            `obj.bar.foo = function bar() {};`,
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2015},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matchProperty", Line: 1, Column: 1},
				},
			},
			{
				Code:            `obj['foo'] = function bar() {};`,
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2015},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matchProperty", Line: 1, Column: 1},
				},
			},
			{
				Code:            `let obj = {foo: function bar() {}};`,
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2015},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matchProperty", Line: 1, Column: 12},
				},
			},
			{
				Code:            `let obj = {'foo': function bar() {}};`,
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2015},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matchProperty", Line: 1, Column: 12},
				},
			},
			{
				Code:            `({['foo']: function bar() {}})`,
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2015},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matchProperty", Line: 1, Column: 3},
				},
			},
			{
				Code:            `module.exports = function foo(name) {};`,
				Options:         []any{map[string]any{"includeCommonJSModuleExports": true}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2015},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matchProperty", Line: 1, Column: 1},
				},
			},
			{
				Code:            `module.exports = function foo(name) {};`,
				Options:         []any{"always", map[string]any{"includeCommonJSModuleExports": true}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2015},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matchProperty", Line: 1, Column: 1},
				},
			},
			{
				Code:            `module.exports = function exports(name) {};`,
				Options:         []any{"never", map[string]any{"includeCommonJSModuleExports": true}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2015},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notMatchProperty", Line: 1, Column: 1},
				},
			},
			{
				Code:            `module['exports'] = function foo(name) {};`,
				Options:         []any{map[string]any{"includeCommonJSModuleExports": true}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2015},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matchProperty", Line: 1, Column: 1},
				},
			},
			{
				Code:            `module['exports'] = function foo(name) {};`,
				Options:         []any{"always", map[string]any{"includeCommonJSModuleExports": true}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2015},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matchProperty", Line: 1, Column: 1},
				},
			},
			{
				Code:            `module['exports'] = function exports(name) {};`,
				Options:         []any{"never", map[string]any{"includeCommonJSModuleExports": true}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2015},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notMatchProperty", Line: 1, Column: 1},
				},
			},
			{
				Code:    `var foo = function foo(name) {};`,
				Options: []any{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notMatchVariable", Line: 1, Column: 5},
				},
			},
			{
				Code:    `obj.foo = function foo(name) {};`,
				Options: []any{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notMatchProperty", Line: 1, Column: 1},
				},
			},
			{
				Code:    `Object.defineProperty(foo, 'bar', { value: function baz() {} })`,
				Options: []any{"always", map[string]any{"considerPropertyDescriptor": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matchProperty", Line: 1, Column: 37},
				},
			},
			{
				Code:    `Object.defineProperties(foo, { bar: { value: function baz() {} } })`,
				Options: []any{"always", map[string]any{"considerPropertyDescriptor": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matchProperty", Line: 1, Column: 39},
				},
			},
			{
				Code:    `Object.create(proto, { bar: { value: function baz() {} } })`,
				Options: []any{"always", map[string]any{"considerPropertyDescriptor": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matchProperty", Line: 1, Column: 31},
				},
			},
			{
				Code:    `var obj = { value: function foo(name) {} }`,
				Options: []any{"always", map[string]any{"considerPropertyDescriptor": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matchProperty", Line: 1, Column: 13},
				},
			},
			{
				Code:    `Object.defineProperty(foo, 'bar', { value: function bar() {} })`,
				Options: []any{"never", map[string]any{"considerPropertyDescriptor": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notMatchProperty", Line: 1, Column: 37},
				},
			},
			{
				Code:    `Object.defineProperties(foo, { bar: { value: function bar() {} } })`,
				Options: []any{"never", map[string]any{"considerPropertyDescriptor": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notMatchProperty", Line: 1, Column: 39},
				},
			},
			{
				Code:    `Object.create(proto, { bar: { value: function bar() {} } })`,
				Options: []any{"never", map[string]any{"considerPropertyDescriptor": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notMatchProperty", Line: 1, Column: 31},
				},
			},
			{
				Code:    `Reflect.defineProperty(foo, 'bar', { value: function baz() {} })`,
				Options: []any{"always", map[string]any{"considerPropertyDescriptor": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matchProperty", Line: 1, Column: 38},
				},
			},
			{
				Code:    `Reflect.defineProperty(foo, 'bar', { value: function bar() {} })`,
				Options: []any{"never", map[string]any{"considerPropertyDescriptor": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notMatchProperty", Line: 1, Column: 38},
				},
			},
			{
				Code:    `foo({ value: function bar() {} })`,
				Options: []any{"always", map[string]any{"considerPropertyDescriptor": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matchProperty", Line: 1, Column: 7},
				},
			},

			// ---- Optional chaining ----
			{
				Code:            `(obj?.aaa).foo = function bar() {};`,
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2020},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matchProperty", Line: 1, Column: 1},
				},
			},
			{
				Code:            `Object?.defineProperty(foo, 'bar', { value: function baz() {} })`,
				Options:         []any{"always", map[string]any{"considerPropertyDescriptor": true}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2020},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matchProperty", Line: 1, Column: 38},
				},
			},
			{
				Code:            `(Object?.defineProperty)(foo, 'bar', { value: function baz() {} })`,
				Options:         []any{"always", map[string]any{"considerPropertyDescriptor": true}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2020},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matchProperty", Line: 1, Column: 40},
				},
			},
			{
				Code:            `Object?.defineProperty(foo, 'bar', { value: function bar() {} })`,
				Options:         []any{"never", map[string]any{"considerPropertyDescriptor": true}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2020},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notMatchProperty", Line: 1, Column: 38},
				},
			},
			{
				Code:            `(Object?.defineProperty)(foo, 'bar', { value: function bar() {} })`,
				Options:         []any{"never", map[string]any{"considerPropertyDescriptor": true}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2020},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notMatchProperty", Line: 1, Column: 40},
				},
			},
			{
				Code:            `Object?.defineProperties(foo, { bar: { value: function baz() {} } })`,
				Options:         []any{"always", map[string]any{"considerPropertyDescriptor": true}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2020},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matchProperty", Line: 1, Column: 40},
				},
			},
			{
				Code:            `(Object?.defineProperties)(foo, { bar: { value: function baz() {} } })`,
				Options:         []any{"always", map[string]any{"considerPropertyDescriptor": true}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2020},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matchProperty", Line: 1, Column: 42},
				},
			},
			{
				Code:            `Object?.defineProperties(foo, { bar: { value: function bar() {} } })`,
				Options:         []any{"never", map[string]any{"considerPropertyDescriptor": true}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2020},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notMatchProperty", Line: 1, Column: 40},
				},
			},
			{
				Code:            `(Object?.defineProperties)(foo, { bar: { value: function bar() {} } })`,
				Options:         []any{"never", map[string]any{"considerPropertyDescriptor": true}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2020},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notMatchProperty", Line: 1, Column: 42},
				},
			},

			// ---- class fields ----
			{
				Code:            `class C { x = function y() {}; }`,
				Options:         []any{"always"},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matchProperty", Line: 1, Column: 11},
				},
			},
			{
				Code:            `class C { x = function x() {}; }`,
				Options:         []any{"never"},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notMatchProperty", Line: 1, Column: 11},
				},
			},
			{
				Code:            `class C { 'x' = function y() {}; }`,
				Options:         []any{"always"},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matchProperty", Line: 1, Column: 11},
				},
			},
			{
				Code:            `class C { 'x' = function x() {}; }`,
				Options:         []any{"never"},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notMatchProperty", Line: 1, Column: 11},
				},
			},
			{
				Code:            `class C { ['x'] = function y() {}; }`,
				Options:         []any{"always"},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matchProperty", Line: 1, Column: 11},
				},
			},
			{
				Code:            `class C { ['x'] = function x() {}; }`,
				Options:         []any{"never"},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notMatchProperty", Line: 1, Column: 11},
				},
			},
			{
				Code:            `class C { static x = function y() {}; }`,
				Options:         []any{"always"},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matchProperty", Line: 1, Column: 11},
				},
			},
			{
				Code:            `class C { static x = function x() {}; }`,
				Options:         []any{"never"},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notMatchProperty", Line: 1, Column: 11},
				},
			},
			{
				Code:            `(class { x = function y() {}; })`,
				Options:         []any{"always"},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matchProperty", Line: 1, Column: 10},
				},
			},
			{
				Code:            `(class { x = function x() {}; })`,
				Options:         []any{"never"},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notMatchProperty", Line: 1, Column: 10},
				},
			},
			{
				Code:            `var obj = { 'ᢅ': function foo() {} };`,
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2015},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matchProperty", Line: 1, Column: 13},
				},
			},

			// DIVERGENCE: upstream keeps this valid at ecmaVersion 5 (esutils'
			// frozen ES5.1/Unicode-v9 identifier table excludes U+1885).
			// rslint's identifier-shape check uses a single, current Unicode
			// table at every ecmaVersion and accepts U+1885, so this reports
			// here where ESLint stays silent — see func_name_matching.md's
			// "Differences from ESLint" section.
			{
				Code:            `var obj = { 'ᢅ': function foo() {} };`,
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 5},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matchProperty", Line: 1, Column: 13},
				},
			},
		},
	)
}
