package no_underscore_dangle

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestNoUnderscoreDangleExtrasBranches locks in every reachable branch of the
// upstream rule source, including the arms upstream's own suite never reaches —
// each helper's early exits, each allow-after option in isolation, and the
// parent walk that decides which destructuring option applies. Each case names
// the upstream arm it pins. Upstream-migrated cases live in
// no_underscore_dangle_upstream_test.go.
func TestNoUnderscoreDangleExtrasBranches(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUnderscoreDangleRule,
		[]rule_tester.ValidTestCase{
			// Locks in upstream hasDanglingUnderscore() arm 1: the bare `_` name is exempt everywhere
			{Code: `var _ = 1;`},
			{Code: `foo._;`},
			{Code: `class A { _ = 1; #_ = 1; }`, Options: map[string]any{"enforceInClassFields": true}},
			{Code: `class A { _() {} #_() {} }`, Options: map[string]any{"enforceInMethodNames": true}},
			{Code: `function foo(_) {}`, Options: map[string]any{"allowFunctionParams": false}},

			// Locks in upstream hasDanglingUnderscore() arm 2: leading vs trailing underscore
			{Code: `var foo_bar = 1;`},

			// Locks in upstream isSpecialCaseIdentifierForMemberExpression(): member-access only
			{Code: `foo.__proto__;`},
			{Code: `this.__proto__;`, Options: map[string]any{"allowAfterThis": false}},

			// Locks in upstream isThisConstructorReference() arm 2: receiver key is not `constructor`
			{Code: `this[constructor]._bar;`, Options: map[string]any{"allowAfterThisConstructor": true}},

			// Locks in upstream checkForDanglingUnderscoreInMemberExpression(): each allow-after option is independent
			{Code: `this._bar;`, Options: map[string]any{"allowAfterThis": true, "allowAfterSuper": false}},
			{Code: `class A extends B { m() { super._bar; } }`, Options: map[string]any{"allowAfterSuper": true}},
			{Code: `foo._bar;`, Options: map[string]any{"allow": []any{"_bar"}}},

			// Locks in upstream checkForDanglingUnderscoreInFunction(): only a named FunctionDeclaration is checked
			{Code: `(function _foo() {})`},
			{Code: `export default function () {}`},

			// Locks in upstream checkForDanglingUnderscoreInFunctionParameters(): RestElement / AssignmentPattern / plain arms
			{Code: `function foo([_a]) {}`, Options: map[string]any{"allowFunctionParams": false}},
			{Code: `function foo({ _a }) {}`, Options: map[string]any{"allowFunctionParams": false}},
			{Code: `function foo(...[_a]) {}`, Options: map[string]any{"allowFunctionParams": false}},
			{Code: `function foo([_a] = []) {}`, Options: map[string]any{"allowFunctionParams": false}},

			// Locks in upstream checkForDanglingUnderscoreInVariableExpression(): the parent walk stops at the nearest pattern
			{Code: `const { _a: b } = obj;`, Options: map[string]any{"allowInObjectDestructuring": false}},
			{Code: `const { [_k]: v } = obj;`, Options: map[string]any{"allowInObjectDestructuring": false}},

			// Locks in upstream checkForDanglingUnderscoreInMethod(): `isMethod` excludes non-method properties and object accessors
			{Code: `const o = { _p: 1 };`, Options: map[string]any{"enforceInMethodNames": true}},
			{Code: `const o = { _p: function () {} };`, Options: map[string]any{"enforceInMethodNames": true}},
			{Code: `const o = { _p: () => {} };`, Options: map[string]any{"enforceInMethodNames": true}},
			{Code: `const o = { get _g() { return 1; }, set _s(v) {} };`, Options: map[string]any{"enforceInMethodNames": true}},
			{Code: `const { _p } = obj;`, Options: map[string]any{"enforceInMethodNames": true}},

			// Locks in upstream checkForDanglingUnderscoreInMethod()/InClassField(): the option gates are independent
			{Code: `class A { _m() {} _f = 1; }`},
			{Code: `class A { _m() {} _f = 1; }`, Options: map[string]any{"enforceInMethodNames": true, "enforceInClassFields": true, "allow": []any{"_m", "_f"}}},
		},
		[]rule_tester.InvalidTestCase{
			// Locks in upstream hasDanglingUnderscore() arm 2: leading vs trailing underscore
			{
				Code: `var _foo = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_foo'.", Line: 1, Column: 5, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code: `var foo_ = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in 'foo_'.", Line: 1, Column: 5, EndLine: 1, EndColumn: 13},
				},
			},

			// Locks in upstream isSpecialCaseIdentifierForMemberExpression(): member-access only
			{
				Code: `var __proto__ = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '__proto__'.", Line: 1, Column: 5, EndLine: 1, EndColumn: 18},
				},
			},
			{
				Code:    `class A { __proto__ = 1; }`,
				Options: map[string]any{"enforceInClassFields": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '__proto__'.", Line: 1, Column: 11, EndLine: 1, EndColumn: 25},
				},
			},
			{
				Code:    `const o = { __proto__() {} };`,
				Options: map[string]any{"enforceInMethodNames": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '__proto__'.", Line: 1, Column: 13, EndLine: 1, EndColumn: 27},
				},
			},

			// Locks in upstream isThisConstructorReference() arm 1: receiver is not a member access
			{
				Code:    `this._bar;`,
				Options: map[string]any{"allowAfterThisConstructor": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_bar'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 10},
				},
			},

			// Locks in upstream isThisConstructorReference() arm 2: receiver key is not `constructor`
			{
				Code:    `this.foo._bar;`,
				Options: map[string]any{"allowAfterThisConstructor": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_bar'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 14},
				},
			},
			{
				Code:    `this['constructor']._bar;`,
				Options: map[string]any{"allowAfterThisConstructor": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_bar'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 25},
				},
			},

			// Locks in upstream isThisConstructorReference() arm 3: receiver base is not `this`
			{
				Code:    `a.constructor._bar;`,
				Options: map[string]any{"allowAfterThisConstructor": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_bar'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 19},
				},
			},
			{
				Code:    `class A extends B { m() { super.constructor._bar; } }`,
				Options: map[string]any{"allowAfterThisConstructor": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_bar'.", Line: 1, Column: 27, EndLine: 1, EndColumn: 49},
				},
			},

			// Locks in upstream checkForDanglingUnderscoreInMemberExpression(): each allow-after option is independent
			{
				Code:    `class A extends B { m() { super._bar; } }`,
				Options: map[string]any{"allowAfterThis": true, "allowAfterSuper": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_bar'.", Line: 1, Column: 27, EndLine: 1, EndColumn: 37},
				},
			},
			{
				Code:    `this._bar;`,
				Options: map[string]any{"allowAfterSuper": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_bar'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 10},
				},
			},
			{
				Code:    `foo._bar;`,
				Options: map[string]any{"allowAfterThis": true, "allowAfterSuper": true, "allowAfterThisConstructor": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_bar'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 9},
				},
			},

			// Locks in upstream checkForDanglingUnderscoreInFunction(): only a named FunctionDeclaration is checked
			{
				Code: `const _foo = function () {};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_foo'.", Line: 1, Column: 7, EndLine: 1, EndColumn: 28},
				},
			},

			// Locks in upstream checkForDanglingUnderscoreInFunctionParameters(): RestElement / AssignmentPattern / plain arms
			{
				Code:    `function foo(..._a) {}`,
				Options: map[string]any{"allowFunctionParams": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_a'.", Line: 1, Column: 14, EndLine: 1, EndColumn: 19},
				},
			},
			{
				Code:    `function foo(_a = 1) {}`,
				Options: map[string]any{"allowFunctionParams": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_a'.", Line: 1, Column: 14, EndLine: 1, EndColumn: 20},
				},
			},
			{
				Code:    `function foo(_a) {}`,
				Options: map[string]any{"allowFunctionParams": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_a'.", Line: 1, Column: 14, EndLine: 1, EndColumn: 16},
				},
			},
			{
				Code:    `function foo(_a, _b) {}`,
				Options: map[string]any{"allowFunctionParams": false, "allow": []any{"_a"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_b'.", Line: 1, Column: 18, EndLine: 1, EndColumn: 20},
				},
			},
			{
				Code:    `const o = { set _s(_v) {} };`,
				Options: map[string]any{"allowFunctionParams": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_v'.", Line: 1, Column: 20, EndLine: 1, EndColumn: 22},
				},
			},
			{
				Code:    `class A { set _s(_v) {} }`,
				Options: map[string]any{"allowFunctionParams": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_v'.", Line: 1, Column: 18, EndLine: 1, EndColumn: 20},
				},
			},
			{
				Code:    `class A { constructor(_a) {} }`,
				Options: map[string]any{"allowFunctionParams": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_a'.", Line: 1, Column: 23, EndLine: 1, EndColumn: 25},
				},
			},
			{
				Code:    `const fn = async (_a) => {};`,
				Options: map[string]any{"allowFunctionParams": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_a'.", Line: 1, Column: 19, EndLine: 1, EndColumn: 21},
				},
			},

			// Locks in upstream checkForDanglingUnderscoreInVariableExpression(): the parent walk stops at the nearest pattern
			{
				Code:    `const { foo: [_bar] } = obj;`,
				Options: map[string]any{"allowInArrayDestructuring": false, "allowInObjectDestructuring": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_bar'.", Line: 1, Column: 7, EndLine: 1, EndColumn: 28},
				},
			},
			{
				Code:    `const [{ _bar }] = arr;`,
				Options: map[string]any{"allowInArrayDestructuring": true, "allowInObjectDestructuring": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_bar'.", Line: 1, Column: 7, EndLine: 1, EndColumn: 23},
				},
			},
			{
				Code:    `const { _a = 1 } = obj;`,
				Options: map[string]any{"allowInObjectDestructuring": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_a'.", Line: 1, Column: 7, EndLine: 1, EndColumn: 23},
				},
			},
			{
				Code:    `const [_a = 1] = arr;`,
				Options: map[string]any{"allowInArrayDestructuring": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_a'.", Line: 1, Column: 7, EndLine: 1, EndColumn: 21},
				},
			},
			{
				Code:    `const { a: _b } = obj;`,
				Options: map[string]any{"allowInObjectDestructuring": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_b'.", Line: 1, Column: 7, EndLine: 1, EndColumn: 22},
				},
			},
			// Locks in upstream checkForDanglingUnderscoreInVariableExpression(): getDeclaredVariables() yields one variable per name, so a name bound twice by one declarator reports once
			{
				Code:    `var {a: _x, b: _x} = o;`,
				Options: map[string]any{"allowInObjectDestructuring": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_x'.", Line: 1, Column: 5, EndLine: 1, EndColumn: 23},
				},
			},
			{
				Code:    `var [_x, _x] = o;`,
				Options: map[string]any{"allowInArrayDestructuring": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_x'.", Line: 1, Column: 5, EndLine: 1, EndColumn: 17},
				},
			},
			{
				Code: `let _a, _b;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_a'.", Line: 1, Column: 5, EndLine: 1, EndColumn: 7},
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_b'.", Line: 1, Column: 9, EndLine: 1, EndColumn: 11},
				},
			},
			{
				Code: `for (const _x of xs) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_x'.", Line: 1, Column: 12, EndLine: 1, EndColumn: 14},
				},
			},
			{
				Code:    `for (const [_x] of xs) {}`,
				Options: map[string]any{"allowInArrayDestructuring": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_x'.", Line: 1, Column: 12, EndLine: 1, EndColumn: 16},
				},
			},
			{
				Code: `using _x = f();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_x'.", Line: 1, Column: 7, EndLine: 1, EndColumn: 15},
				},
			},
			{
				Code: `class A { static { let _x; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_x'.", Line: 1, Column: 24, EndLine: 1, EndColumn: 26},
				},
			},

			// Locks in upstream checkForDanglingUnderscoreInMethod(): `isMethod` excludes non-method properties and object accessors
			{
				Code:    `class A { get _g() { return 1; } set _s(v) {} }`,
				Options: map[string]any{"enforceInMethodNames": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_g'.", Line: 1, Column: 11, EndLine: 1, EndColumn: 33},
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_s'.", Line: 1, Column: 34, EndLine: 1, EndColumn: 46},
				},
			},

			// Locks in upstream checkForDanglingUnderscoreInMethod()/InClassField(): the option gates are independent
			{
				Code:    `class A { _m() {} _f = 1; }`,
				Options: map[string]any{"enforceInMethodNames": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_m'.", Line: 1, Column: 11, EndLine: 1, EndColumn: 18},
				},
			},
			{
				Code:    `class A { _m() {} _f = 1; }`,
				Options: map[string]any{"enforceInClassFields": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_f'.", Line: 1, Column: 19, EndLine: 1, EndColumn: 26},
				},
			},
			{
				Code:    `class A { _m() {} _f = 1; }`,
				Options: map[string]any{"enforceInMethodNames": true, "enforceInClassFields": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_m'.", Line: 1, Column: 11, EndLine: 1, EndColumn: 18},
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_f'.", Line: 1, Column: 19, EndLine: 1, EndColumn: 26},
				},
			},

			// ---- Option defaults: no options behaves exactly like an empty option object ----
			{
				Code: `var _foo = 1; foo._bar; const { _a } = obj; const [_b] = arr; function f(_p) {} class A { _m() {} _f = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_foo'.", Line: 1, Column: 5, EndLine: 1, EndColumn: 13},
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_bar'.", Line: 1, Column: 15, EndLine: 1, EndColumn: 23},
				},
			},
			{
				Code:    `var _foo = 1; foo._bar; const { _a } = obj; const [_b] = arr; function f(_p) {} class A { _m() {} _f = 1; }`,
				Options: map[string]any{},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_foo'.", Line: 1, Column: 5, EndLine: 1, EndColumn: 13},
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_bar'.", Line: 1, Column: 15, EndLine: 1, EndColumn: 23},
				},
			},
		},
	)
}
