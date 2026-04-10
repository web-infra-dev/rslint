package no_unsafe_optional_chaining

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUnsafeOptionalChainingRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUnsafeOptionalChainingRule,
		// Valid cases
		[]rule_tester.ValidTestCase{
			// Standalone optional chaining
			{Code: `obj?.foo;`},
			{Code: `obj?.foo();`},
			{Code: `obj?.foo.bar;`},
			{Code: `obj?.foo?.bar;`},
			{Code: `obj?.foo?.bar();`},
			{Code: `obj?.foo?.bar?.();`},

			// Fallback operators neutralize the chain
			{Code: `(obj?.foo ?? bar)();`},
			{Code: `(obj?.foo ?? bar).baz;`},
			{Code: `(obj?.foo ?? bar)[0];`},
			{Code: `new (obj?.foo ?? bar)();`},
			{Code: `(obj?.foo || bar)();`},
			{Code: `(obj?.foo || bar).baz;`},

			// Arithmetic without option is valid
			{Code: `obj?.foo + bar;`},
			{Code: `+obj?.foo;`},

			// Spread in object literal is safe
			{Code: `({...obj?.foo});`},
			{Code: `({...obj?.foo, x: 1});`},

			// Sequence: chain is NOT last — only last element matters
			{Code: `(obj?.foo, bar)();`},
			{Code: `(obj?.foo, bar).baz;`},

			// for-in tolerates undefined (not for-of)
			{Code: `for (const x in obj?.foo) {}`},

			// Non-destructuring binding defaults are safe
			{Code: `const {x = obj?.foo} = obj;`},
			{Code: `function f({x = obj?.foo}: any) {}`},
			{Code: `const [a = obj?.foo] = arr;`},

			// Non-null assertion — developer explicitly asserts
			{Code: `(obj?.foo)!.bar;`},
			{Code: `(obj?.foo)!();`},
			{Code: `new (obj?.foo)!();`},

			// && left is chain but consumed by ||/?? fallback
			{Code: `((obj?.foo && bar) ?? fallback)();`},
			{Code: `((obj?.foo && bar) || fallback).baz;`},

			// Nested fallbacks
			{Code: `(obj?.foo ?? (obj?.bar ?? baz))();`},
			{Code: `((obj?.foo || a) ?? b).bar;`},

			// Non-optional access on non-chain
			{Code: `obj.foo.bar;`},
			{Code: `new obj.Foo();`},

			// Arithmetic with fallback (with option)
			{
				Code:    `(obj?.foo ?? 0) + bar;`,
				Options: map[string]interface{}{"disallowArithmeticOperators": true},
			},
			{
				Code:    `+(obj?.foo ?? 0);`,
				Options: map[string]interface{}{"disallowArithmeticOperators": true},
			},
			{
				Code:    `bar -= (obj?.foo ?? 0);`,
				Options: map[string]interface{}{"disallowArithmeticOperators": true},
			},

			// Class extends with non-optional
			{Code: `class Foo extends Bar {}`},

			// in/instanceof with non-optional
			{Code: `"foo" in obj;`},
			{Code: `obj instanceof Foo;`},

			// Spread with non-optional
			{Code: `[...arr];`},

			// Destructuring with non-optional
			{Code: `const {foo} = obj;`},
			{Code: `const [a] = arr;`},
		},
		// Invalid cases
		[]rule_tester.InvalidTestCase{
			// === BASIC CONTEXTS ===
			{
				Code: `(obj?.foo)();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeOptionalChain", Line: 1, Column: 2},
				},
			},
			{
				Code: `(obj?.foo).bar;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeOptionalChain", Line: 1, Column: 2},
				},
			},
			{
				Code: `(obj?.foo)[0];`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeOptionalChain", Line: 1, Column: 2},
				},
			},
			{
				Code: `new (obj?.foo)();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeOptionalChain", Line: 1, Column: 6},
				},
			},
			{
				Code: "(obj?.foo)`text`;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeOptionalChain", Line: 1, Column: 2},
				},
			},
			{
				Code: `[...obj?.foo];`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeOptionalChain", Line: 1, Column: 5},
				},
			},
			// Spread in function call args
			{
				Code: `foo(...obj?.foo);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeOptionalChain", Line: 1, Column: 8},
				},
			},
			{
				Code: `for (const x of obj?.foo) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeOptionalChain", Line: 1, Column: 17},
				},
			},
			{
				Code: `"foo" in obj?.bar;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeOptionalChain", Line: 1, Column: 10},
				},
			},
			{
				Code: `foo instanceof obj?.bar;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeOptionalChain", Line: 1, Column: 16},
				},
			},

			// === DESTRUCTURING ===
			{
				Code: `const {foo} = obj?.bar;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeOptionalChain", Line: 1, Column: 15},
				},
			},
			{
				Code: `const [a] = obj?.bar;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeOptionalChain", Line: 1, Column: 13},
				},
			},
			// Destructuring assignment
			{
				Code: `({foo} = obj?.bar);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeOptionalChain", Line: 1, Column: 10},
				},
			},
			{
				Code: `([a] = obj?.bar);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeOptionalChain", Line: 1, Column: 8},
				},
			},
			// Nested binding element: object pattern default
			{
				Code: `const {x: {y} = obj?.foo} = obj;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeOptionalChain", Line: 1, Column: 17},
				},
			},
			// Nested binding element: array pattern default
			{
				Code: `const [, [a] = obj?.foo] = arr;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeOptionalChain", Line: 1, Column: 16},
				},
			},
			// Function parameter binding element
			{
				Code: `function f({x: {y} = obj?.foo}: any) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeOptionalChain", Line: 1, Column: 22},
				},
			},
			// Arrow parameter binding element
			{
				Code: `const g = ({x: [y] = obj?.foo}: any) => {};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeOptionalChain", Line: 1, Column: 22},
				},
			},

			// === CLASS EXTENDS ===
			{
				Code: `class Foo extends obj?.bar {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeOptionalChain", Line: 1, Column: 19},
				},
			},
			// Class expression
			{
				Code: `const C = class extends obj?.bar {};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeOptionalChain", Line: 1, Column: 25},
				},
			},

			// === checkUndefinedShortCircuit TRAVERSAL ===
			// Deep parentheses
			{
				Code: `((((obj?.foo))))();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeOptionalChain", Line: 1, Column: 5},
				},
			},
			// Ternary: both branches — 2 errors
			{
				Code: `(cond ? obj?.foo : obj?.bar)();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeOptionalChain", Line: 1, Column: 9},
					{MessageId: "unsafeOptionalChain", Line: 1, Column: 20},
				},
			},
			// Ternary: only consequent
			{
				Code: `(cond ? obj?.foo : bar)();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeOptionalChain", Line: 1, Column: 9},
				},
			},
			// Ternary: only alternate
			{
				Code: `(cond ? bar : obj?.foo)();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeOptionalChain", Line: 1, Column: 15},
				},
			},
			// Nested ternary — 3 errors
			{
				Code: `(cond ? (cond ? obj?.a : obj?.b) : obj?.c)();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeOptionalChain", Line: 1, Column: 17},
					{MessageId: "unsafeOptionalChain", Line: 1, Column: 26},
					{MessageId: "unsafeOptionalChain", Line: 1, Column: 36},
				},
			},
			// && propagates both sides — 2 errors
			{
				Code: `(obj?.foo && obj?.bar)();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeOptionalChain", Line: 1, Column: 2},
					{MessageId: "unsafeOptionalChain", Line: 1, Column: 14},
				},
			},
			// && left only
			{
				Code: `(obj?.foo && bar)();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeOptionalChain", Line: 1, Column: 2},
				},
			},
			// && right only
			{
				Code: `(a && obj?.foo)();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeOptionalChain", Line: 1, Column: 7},
				},
			},
			// Sequence: chain IS last
			{
				Code: `(a, b, obj?.foo)();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeOptionalChain", Line: 1, Column: 8},
				},
			},
			// Await
			{
				Code: `async function h() { (await obj?.foo)(); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeOptionalChain", Line: 1, Column: 29},
				},
			},
			// Await in ternary — 2 errors
			{
				Code: `async function h2() { (cond ? await obj?.foo : obj?.bar)(); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeOptionalChain", Line: 1, Column: 37},
					{MessageId: "unsafeOptionalChain", Line: 1, Column: 48},
				},
			},
			// Mixed: && inside ||, chain on && left — safe via ||, but && right has chain
			{
				Code: `(((obj?.foo && a) || b) && obj?.bar)();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeOptionalChain", Line: 1, Column: 28},
				},
			},
			// Deeply nested parens + ternary + await — 2 errors
			{
				Code: `async function h3() { (cond ? ((await obj?.foo)) : ((obj?.bar)))(); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeOptionalChain", Line: 1, Column: 39},
					{MessageId: "unsafeOptionalChain", Line: 1, Column: 54},
				},
			},

			// === ARITHMETIC (with disallowArithmeticOperators) ===
			// Both sides — 2 errors
			{
				Code:    `obj?.foo + obj?.bar;`,
				Options: map[string]interface{}{"disallowArithmeticOperators": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeArithmetic", Line: 1, Column: 1},
					{MessageId: "unsafeArithmetic", Line: 1, Column: 12},
				},
			},
			// All arithmetic operators
			{
				Code:    `obj?.foo - bar;`,
				Options: map[string]interface{}{"disallowArithmeticOperators": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeArithmetic", Line: 1, Column: 1},
				},
			},
			{
				Code:    `obj?.foo * bar;`,
				Options: map[string]interface{}{"disallowArithmeticOperators": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeArithmetic", Line: 1, Column: 1},
				},
			},
			{
				Code:    `obj?.foo / bar;`,
				Options: map[string]interface{}{"disallowArithmeticOperators": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeArithmetic", Line: 1, Column: 1},
				},
			},
			{
				Code:    `obj?.foo % bar;`,
				Options: map[string]interface{}{"disallowArithmeticOperators": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeArithmetic", Line: 1, Column: 1},
				},
			},
			{
				Code:    `obj?.foo ** bar;`,
				Options: map[string]interface{}{"disallowArithmeticOperators": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeArithmetic", Line: 1, Column: 1},
				},
			},
			// Unary +/-
			{
				Code:    `+obj?.foo;`,
				Options: map[string]interface{}{"disallowArithmeticOperators": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeArithmetic", Line: 1, Column: 2},
				},
			},
			{
				Code:    `-obj?.foo;`,
				Options: map[string]interface{}{"disallowArithmeticOperators": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeArithmetic", Line: 1, Column: 2},
				},
			},
			// All compound assignments
			{
				Code:    `x += obj?.foo;`,
				Options: map[string]interface{}{"disallowArithmeticOperators": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeArithmetic", Line: 1, Column: 6},
				},
			},
			{
				Code:    `x -= obj?.foo;`,
				Options: map[string]interface{}{"disallowArithmeticOperators": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeArithmetic", Line: 1, Column: 6},
				},
			},
			{
				Code:    `x *= obj?.foo;`,
				Options: map[string]interface{}{"disallowArithmeticOperators": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeArithmetic", Line: 1, Column: 6},
				},
			},
			{
				Code:    `x /= obj?.foo;`,
				Options: map[string]interface{}{"disallowArithmeticOperators": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeArithmetic", Line: 1, Column: 6},
				},
			},
			{
				Code:    `x %= obj?.foo;`,
				Options: map[string]interface{}{"disallowArithmeticOperators": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeArithmetic", Line: 1, Column: 6},
				},
			},
			{
				Code:    `x **= obj?.foo;`,
				Options: map[string]interface{}{"disallowArithmeticOperators": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeArithmetic", Line: 1, Column: 7},
				},
			},
			// Arithmetic with ternary chain — 2 errors
			{
				Code:    `(cond ? obj?.a : obj?.b) + 1;`,
				Options: map[string]interface{}{"disallowArithmeticOperators": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeArithmetic", Line: 1, Column: 9},
					{MessageId: "unsafeArithmetic", Line: 1, Column: 18},
				},
			},

			// === COMPLEX COMBINATIONS ===
			// in with ternary — 2 errors
			{
				Code: `"key" in (cond ? obj?.a : obj?.b);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeOptionalChain", Line: 1, Column: 18},
					{MessageId: "unsafeOptionalChain", Line: 1, Column: 27},
				},
			},
			// for-of with conditional — 2 errors
			{
				Code: `for (const v of (cond ? obj?.a : obj?.b)) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeOptionalChain", Line: 1, Column: 25},
					{MessageId: "unsafeOptionalChain", Line: 1, Column: 34},
				},
			},
			// Spread with conditional — 2 errors
			{
				Code: `[...(cond ? obj?.a : obj?.b)];`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeOptionalChain", Line: 1, Column: 13},
					{MessageId: "unsafeOptionalChain", Line: 1, Column: 22},
				},
			},
			// Destructuring with conditional — 2 errors
			{
				Code: `const {w} = (cond ? obj?.a : obj?.b);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeOptionalChain", Line: 1, Column: 21},
					{MessageId: "unsafeOptionalChain", Line: 1, Column: 30},
				},
			},
			// Class extends with conditional — 2 errors
			{
				Code: `class Bar extends (cond ? obj?.a : obj?.b) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeOptionalChain", Line: 1, Column: 27},
					{MessageId: "unsafeOptionalChain", Line: 1, Column: 36},
				},
			},
		},
	)
}
