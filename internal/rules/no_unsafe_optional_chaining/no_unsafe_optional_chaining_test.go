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
			// Standalone optional chaining is fine
			{Code: `obj?.foo;`},
			{Code: `obj?.foo();`},
			{Code: `obj?.foo.bar;`},
			{Code: `obj?.foo?.bar;`},
			{Code: `obj?.foo?.bar();`},
			{Code: `obj?.foo?.bar?.();`},

			// Nullish coalescing provides a fallback
			{Code: `(obj?.foo ?? bar)();`},
			{Code: `(obj?.foo ?? bar).baz;`},
			{Code: `(obj?.foo ?? bar)[0];`},
			{Code: `new (obj?.foo ?? bar)();`},

			// Logical OR provides a fallback
			{Code: `(obj?.foo || bar)();`},
			{Code: `(obj?.foo || bar).baz;`},

			// Optional call is safe
			{Code: `obj?.foo?.();`},

			// Arithmetic without option is valid
			{Code: `obj?.foo + bar;`},
			{Code: `+obj?.foo;`},

			// Standalone in for-of is fine with non-optional chain
			{Code: `for (const x of arr) {}`},

			// in/instanceof with non-optional chain
			{Code: `"foo" in obj;`},
			{Code: `obj instanceof Foo;`},

			// Spread with non-optional
			{Code: `[...arr];`},

			// Destructuring with non-optional
			{Code: `const {foo} = obj;`},
			{Code: `const [a] = arr;`},

			// Sequence expression: only last matters
			{Code: `(obj?.foo, bar)();`},

			// Spread in object literal is safe
			{Code: `({...obj?.foo});`},

			// Class extends with non-optional
			{Code: `class Foo extends Bar {}`},

			// Arithmetic: safe even with option when using ?? fallback
			{
				Code:    `(obj?.foo ?? 0) + bar;`,
				Options: map[string]interface{}{"disallowArithmeticOperators": true},
			},
			{
				Code:    `+(obj?.foo ?? 0);`,
				Options: map[string]interface{}{"disallowArithmeticOperators": true},
			},
		},
		// Invalid cases
		[]rule_tester.InvalidTestCase{
			// Call expression with optional chain callee
			{
				Code: `(obj?.foo)();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeOptionalChain", Line: 1, Column: 1},
				},
			},
			// Property access on optional chain
			{
				Code: `(obj?.foo).bar;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeOptionalChain", Line: 1, Column: 1},
				},
			},
			// Element access on optional chain
			{
				Code: `(obj?.foo)[0];`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeOptionalChain", Line: 1, Column: 1},
				},
			},
			// New expression with optional chain
			{
				Code: `new (obj?.foo)();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeOptionalChain", Line: 1, Column: 1},
				},
			},
			// Destructuring variable declaration with optional chain
			{
				Code: `const {foo} = obj?.bar;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeOptionalChain", Line: 1, Column: 7},
				},
			},
			// Array destructuring with optional chain
			{
				Code: `const [a] = obj?.bar;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeOptionalChain", Line: 1, Column: 7},
				},
			},
			// Spread element with optional chain
			{
				Code: `[...obj?.foo];`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeOptionalChain", Line: 1, Column: 2},
				},
			},
			// For-of with optional chain
			{
				Code: `for (const x of obj?.foo) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeOptionalChain", Line: 1, Column: 1},
				},
			},
			// in operator with optional chain on right
			{
				Code: `"foo" in obj?.bar;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeOptionalChain", Line: 1, Column: 1},
				},
			},
			// instanceof with optional chain on right
			{
				Code: `foo instanceof obj?.bar;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeOptionalChain", Line: 1, Column: 1},
				},
			},
			// Class extends with optional chain
			{
				Code: `class Foo extends obj?.bar {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeOptionalChain", Line: 1, Column: 19},
				},
			},
			// Ternary: both branches have optional chain
			{
				Code: `(cond ? obj?.foo : obj?.bar)();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeOptionalChain", Line: 1, Column: 1},
				},
			},
			// Ternary: one branch has optional chain (still unsafe since it can be taken)
			{
				Code: `(cond ? obj?.foo : bar)();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeOptionalChain", Line: 1, Column: 1},
				},
			},
			{
				Code: `(cond ? bar : obj?.foo)();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeOptionalChain", Line: 1, Column: 1},
				},
			},
			// && propagates optional chain
			{
				Code: `(a && obj?.foo)();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeOptionalChain", Line: 1, Column: 1},
				},
			},

			// Arithmetic with disallowArithmeticOperators option
			{
				Code:    `obj?.foo + bar;`,
				Options: map[string]interface{}{"disallowArithmeticOperators": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeArithmetic", Line: 1, Column: 1},
				},
			},
			{
				Code:    `bar + obj?.foo;`,
				Options: map[string]interface{}{"disallowArithmeticOperators": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeArithmetic", Line: 1, Column: 1},
				},
			},
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
					{MessageId: "unsafeArithmetic", Line: 1, Column: 1},
				},
			},
			{
				Code:    `-obj?.foo;`,
				Options: map[string]interface{}{"disallowArithmeticOperators": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeArithmetic", Line: 1, Column: 1},
				},
			},

			// Destructuring assignment (not declaration)
			{
				Code: `({foo} = obj?.bar);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeOptionalChain", Line: 1, Column: 2},
				},
			},

			// Sequence expression: last element has optional chain
			{
				Code: `(bar, obj?.foo)();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeOptionalChain", Line: 1, Column: 1},
				},
			},

			// NonNullExpression wrapping optional chain
			{
				Code: `(obj?.foo)!.bar;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeOptionalChain", Line: 1, Column: 1},
				},
			},
			{
				Code: `(obj?.foo)!();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeOptionalChain", Line: 1, Column: 1},
				},
			},
			{
				Code: `new (obj?.foo)!();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeOptionalChain", Line: 1, Column: 1},
				},
			},

			// Await expression with optional chain
			{
				Code: `async function f() { (await obj?.foo)(); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeOptionalChain", Line: 1, Column: 22},
				},
			},

			// Compound assignment with optional chain
			{
				Code:    `x += obj?.foo;`,
				Options: map[string]interface{}{"disallowArithmeticOperators": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeArithmetic", Line: 1, Column: 1},
				},
			},
		},
	)
}
