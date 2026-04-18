package no_extend_native

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoExtendNativeRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoExtendNativeRule,
		// Valid cases — mirrors ESLint's `tests/lib/rules/no-extend-native.js`.
		[]rule_tester.ValidTestCase{
			{Code: `x.prototype.p = 0`},
			{Code: `x.prototype['p'] = 0`},
			{Code: `Object.p = 0`},
			{Code: `Object.toString.bind = 0`},
			{Code: `Object['toString'].bind = 0`},
			{Code: `Object.defineProperty(x, 'p', {value: 0})`},
			{Code: `Object.defineProperties(x, {p: {value: 0}})`},
			{Code: `global.Object.prototype.toString = 0`},
			{Code: `this.Object.prototype.toString = 0`},
			{Code: `with(Object) { prototype.p = 0; }`},
			{Code: `o = Object; o.prototype.toString = 0`},
			{Code: `eval('Object.prototype.toString = 0')`},

			// `parseFloat` is a lowercase-first ECMAScript global and so is not
			// considered a "native" builtin by this rule.
			{Code: `parseFloat.prototype.x = 1`},

			// Exception option allows extending Object's prototype.
			{
				Code:    `Object.prototype.g = 0`,
				Options: map[string]interface{}{"exceptions": []interface{}{"Object"}},
			},

			// `Object.prototype` appears as the *index* of a member access, not
			// as the assignment target.
			{Code: `obj[Object.prototype] = 0`},

			// https://github.com/eslint/eslint/issues/4438
			{Code: `Object.defineProperty()`},
			{Code: `Object.defineProperties()`},

			// https://github.com/eslint/eslint/issues/8461 — locally shadowed
			// references are not reported.
			{Code: `function foo() { var Object = function() {}; Object.prototype.p = 0 }`},
			{Code: `{ let Object = function() {}; Object.prototype.p = 0 }`},

			// ---- Extra valid coverage ----
			// Pure read of prototype, not a write.
			{Code: `var x = Object.prototype`},
			{Code: `var x = Object.prototype.p`},
			{Code: `console.log(Object.prototype.p)`},

			// `Reflect.defineProperty` is not flagged — the rule only targets
			// `Object.defineProperty` and `Object.defineProperties`.
			{Code: `Reflect.defineProperty(Array.prototype, 'p', {value: 0})`},

			// Removing a property is not "adding" one.
			{Code: `delete Object.prototype.p`},

			// Increment/decrement are UpdateExpressions, not AssignmentExpressions.
			{Code: `Object.prototype.p++`},
			{Code: `--Object.prototype.p`},

			// Assigning to `prototype` itself (not to a property of it) is a
			// different operation; ESLint deliberately does not report it here.
			{Code: `(Object.prototype) = 0`},

			// TypeScript type assertion wrapping the builtin reference.
			{Code: `(Object as any).prototype.p = 0`},
		},
		// Invalid cases — mirrors ESLint's `tests/lib/rules/no-extend-native.js`.
		[]rule_tester.InvalidTestCase{
			{
				Code: `Object.prototype.p = 0`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code: `BigInt.prototype.p = 0`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code: `WeakRef.prototype.p = 0`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code: `FinalizationRegistry.prototype.p = 0`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code: `AggregateError.prototype.p = 0`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code: `Function.prototype['p'] = 0`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code: `String['prototype'].p = 0`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code: `Number['prototype']['p'] = 0`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code: `Object.defineProperty(Array.prototype, 'p', {value: 0})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code: `Object.defineProperties(Array.prototype, {p: {value: 0}})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code: `Object.defineProperties(Array.prototype, {p: {value: 0}, q: {value: 0}})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code:    `Number['prototype']['p'] = 0`,
				Options: map[string]interface{}{"exceptions": []interface{}{"Object"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code: `Object.prototype.p = 0; Object.prototype.q = 0`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
					{MessageId: "unexpected", Line: 1, Column: 25},
				},
			},
			{
				Code: `function foo() { Object.prototype.p = 0 }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 18},
				},
			},

			// ---- Optional chaining ----
			{
				Code: `(Object?.prototype).p = 0`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code: `Object.defineProperty(Object?.prototype, 'p', { value: 0 })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code: `Object?.defineProperty(Object.prototype, 'p', { value: 0 })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code: `(Object?.defineProperty)(Object.prototype, 'p', { value: 0 })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},

			// ---- Logical assignments ----
			{
				Code: `Array.prototype.p &&= 0`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code: `Array.prototype.p ||= 0`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code: `Array.prototype.p ??= 0`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},

			// ---- Extra coverage ----
			// Multi-line assignment — message position covers the whole assignment.
			{
				Code: "Object\n  .prototype\n  .p = 0",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1, EndLine: 3, EndColumn: 9},
				},
			},
			// Compound (non-logical) assignment.
			{
				Code: `String.prototype.p += 'x'`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			// Exception option does NOT apply to other builtins.
			{
				Code:    `Array.prototype.p = 0`,
				Options: map[string]interface{}{"exceptions": []interface{}{"Object"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			// Message text assertion for the `unexpected` messageId.
			{
				Code: `Object.prototype.p = 0`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected",
						Message:   "Object prototype is read only, properties should not be added.",
						Line:      1,
						Column:    1,
					},
				},
			},

			// ---- Parenthesized identifier (tsgo-only quirk) ----
			// ESLint's ESTree treats parens as transparent; tsgo wraps them
			// in a ParenthesizedExpression node. Both forms must still fire.
			{
				Code: `(Object).prototype.p = 0`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code: `((Object)).prototype.p = 0`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code: `Object.defineProperty((Array).prototype, 'p', {value: 0})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},

			// ---- Builtin coverage ----
			{
				Code: `Symbol.prototype.p = 0`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code: `Map.prototype.p = 0`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code: `Promise.prototype.p = 0`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code: `WeakMap.prototype.p = 0`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},

			// ---- Chained assignment — both sides report ----
			{
				Code: `Object.prototype.p = Array.prototype.q = 0`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
					{MessageId: "unexpected", Line: 1, Column: 22},
				},
			},

			// ---- Template literal as static key ----
			{
				Code: "Object[`prototype`].p = 0",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code: "Object.prototype[`p`] = 0",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},

			// ---- Options schema ----
			// Empty options object behaves like the default (no exceptions).
			{
				Code:    `Object.prototype.p = 0`,
				Options: map[string]interface{}{},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			// Explicit empty exceptions list also matches the default.
			{
				Code:    `Object.prototype.p = 0`,
				Options: map[string]interface{}{"exceptions": []interface{}{}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
		},
	)
}
