package no_magic_numbers

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestEdgeCases tests edge cases that go beyond the upstream test suite,
// particularly around tsgo AST shapes and nested type structures.
func TestEdgeCases(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoMagicNumbersRule, []rule_tester.ValidTestCase{
		// ---- Nested parenthesized type in type alias ----
		// Core ESLint test: type Nested = ('' | ('' | (1)));
		{Code: `type Nested = ('' | ('' | (1)));`, Options: map[string]interface{}{"ignoreNumericLiteralTypes": true}},

		// ---- Type index with deeply nested parenthesized types ----
		{Code: `type Foo = Bar[((((1))))];`, Options: map[string]interface{}{"ignoreTypeIndexes": true}},

		// ---- Enum with computed expressions (sub-expressions are NOT enum values) ----
		// ignoreEnums only applies to direct enum member values, not nested sub-expressions
		{Code: `enum E { A = 1 << 0, B = 1 << 1 }`, Options: map[string]interface{}{"ignoreEnums": true, "ignore": []interface{}{float64(0), float64(1)}}},

		// ---- Readonly with various access modifiers ----
		{Code: `class C { protected readonly x = 42; }`, Options: map[string]interface{}{"ignoreReadonlyClassProperties": true}},

		// ---- parseInt with optional chaining ----
		{Code: `var x = parseInt?.(y, 10);`},

		// ---- Array index with optional chaining element access ----
		{Code: `foo?.bar?.[0]`, Options: map[string]interface{}{"ignoreArrayIndexes": true}},

		// ---- Object property with shorthand (should not report) ----
		// { foo: 42 } is allowed when detectObjects is false (default)
		{Code: `var x = { foo: 42 };`},

		// ---- Color assignment to property (allowed by default) ----
		{Code: `var colors = {}; colors.RED = 2;`},

		// ---- Default value in nested destructuring ----
		{Code: `const { a: { b = 1 } } = obj;`, Options: map[string]interface{}{"ignoreDefaultValues": true}},

		// ---- Class field initializer with negative ----
		{Code: `class C { foo = -42; }`, Options: map[string]interface{}{"ignoreClassFieldInitialValues": true}},

		// ---- BigInt in readonly class property ----
		{Code: `class C { readonly x = 100n; }`, Options: map[string]interface{}{"ignoreReadonlyClassProperties": true}},

		// ---- Computed property key in object literal (number as key is ok, matching ESLint) ----
		{Code: `var x = {[42]: true}`},
		{Code: `var x = {[1 + 2]: true}`, Options: map[string]interface{}{"ignore": []interface{}{float64(1), float64(2)}}},

		// ---- Object destructuring assignment default ----
		{Code: `var one; ({one = 1} = {})`, Options: map[string]interface{}{"ignoreDefaultValues": true}},
		{Code: `var a, b; ({a = 1, b = 2} = {})`, Options: map[string]interface{}{"ignoreDefaultValues": true}},
		{Code: `var x; ({a: x = 42} = {})`, Options: map[string]interface{}{"ignoreDefaultValues": true}},
	}, []rule_tester.InvalidTestCase{
		// ---- Numbers in type literal property (NOT type alias, should report even with ignoreNumericLiteralTypes) ----
		{
			Code:    `type Foo = { bar: 42 };`,
			Options: map[string]interface{}{"ignoreNumericLiteralTypes": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 42.", Line: 1, Column: 19}},
		},
		// ---- Union in type literal property (NOT type alias, should report) ----
		{
			Code:    `type Foo = { bar: 2 | 3 };`,
			Options: map[string]interface{}{"ignoreNumericLiteralTypes": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noMagic", Message: "No magic number: 2.", Line: 1, Column: 19},
				{MessageId: "noMagic", Message: "No magic number: 3.", Line: 1, Column: 23},
			},
		},
		// ---- Type index inside type literal property (NOT an indexed access on the type alias itself) ----
		{
			Code:    `type Foo = { bar: Bar[((1 & -2) | 3) | 4] };`,
			Options: map[string]interface{}{"ignoreNumericLiteralTypes": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noMagic", Message: "No magic number: 1.", Line: 1, Column: 25},
				{MessageId: "noMagic", Message: "No magic number: -2.", Line: 1, Column: 29},
				{MessageId: "noMagic", Message: "No magic number: 3.", Line: 1, Column: 35},
				{MessageId: "noMagic", Message: "No magic number: 4.", Line: 1, Column: 40},
			},
		},
		// ---- Number used as computed class property key (NOT a class field value) ----
		{
			Code:    `class C { 2; }`,
			Options: map[string]interface{}{"ignoreClassFieldInitialValues": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 2.", Line: 1, Column: 11}},
		},
		// ---- Number in computed key brackets (NOT a class field value) ----
		{
			Code:    `class C { [2]; }`,
			Options: map[string]interface{}{"ignoreClassFieldInitialValues": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 2.", Line: 1, Column: 12}},
		},
		// ---- Number as operand in class field expression (NOT a direct initializer) ----
		{
			Code:    `class C { foo = 2 + 3; }`,
			Options: map[string]interface{}{"ignoreClassFieldInitialValues": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noMagic", Message: "No magic number: 2.", Line: 1, Column: 17},
				{MessageId: "noMagic", Message: "No magic number: 3.", Line: 1, Column: 21},
			},
		},
		// ---- Number as method call on numeric literal (not an array index even with ignoreArrayIndexes) ----
		{
			Code:    `100 .toString()`,
			Options: map[string]interface{}{"ignoreArrayIndexes": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 100."}},
		},
		// ---- Double negation in array index (not a valid index expression) ----
		{
			Code:    `foo[-(-1)]`,
			Options: map[string]interface{}{"ignoreArrayIndexes": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: -1."}},
		},
		// ---- Negative float array index (not a valid index) ----
		{
			Code:    `foo[-1.5]`,
			Options: map[string]interface{}{"ignoreArrayIndexes": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: -1.5."}},
		},
		// ---- Non-integer float close to int (doesn't lose precision = still non-integer) ----
		{
			Code:    `foo[5.000000000000001]`,
			Options: map[string]interface{}{"ignoreArrayIndexes": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 5.000000000000001."}},
		},
		// ---- Non-integer from exponent ----
		{
			Code:    `foo[56e-1]`,
			Options: map[string]interface{}{"ignoreArrayIndexes": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 56e-1."}},
		},
		// ---- Negative BigInt index ----
		{
			Code:    `foo[-100n]`,
			Options: map[string]interface{}{"ignoreArrayIndexes": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: -100n."}},
		},
		// ---- Colors assignment pattern (property access = ok, binary expr = reported) ----
		{
			Code: `var colors = {}; colors.RED = 2; colors.YELLOW = 3; colors.BLUE = 4 + 5;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noMagic", Message: "No magic number: 4."},
				{MessageId: "noMagic", Message: "No magic number: 5."},
			},
		},
		// ---- Numbers in array inside JSX (still reported) ----
		{
			Code:     `var a = <div arrayProp={[1,2,3]}></div>;`,
			FileName: "test.tsx",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noMagic", Message: "No magic number: 1."},
				{MessageId: "noMagic", Message: "No magic number: 2."},
				{MessageId: "noMagic", Message: "No magic number: 3."},
			},
		},
		// ---- BigInt vs number ignore mismatch ----
		{
			Code:    `f(100)`,
			Options: map[string]interface{}{"ignore": []interface{}{"100n"}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 100."}},
		},
		// ---- Object destructuring assignment default (NOT ignored) ----
		{
			Code:    `var one; ({one = 1} = {})`,
			Options: map[string]interface{}{"ignoreDefaultValues": false},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 1."}},
		},
		{
			Code:    `var x; ({a: x = 42} = {})`,
			Options: map[string]interface{}{"ignoreDefaultValues": false},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 42."}},
		},
	})
}

// TestUpstreamSemanticLockIn tests that lock in specific semantic branches
// from the upstream ESLint source that are not otherwise tested.
func TestUpstreamSemanticLockIn(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoMagicNumbersRule, []rule_tester.ValidTestCase{
		// Locks in: parseInt radix detection, second argument check
		{Code: `var x = parseInt(y, 10);`},
		// Locks in: Number.parseInt member access with optional chain
		{Code: `var x = Number?.parseInt(y, 10);`},
		// Locks in: ObjectExpression parent is ok when detectObjects=false
		{Code: `var stats = {avg: 42};`},
		// Locks in: PropertyAssignment in object literal is ok
		{Code: `({key: 90, another: 10})`},
		// Locks in: Assignment to property access is ok (not to identifier)
		{Code: `colors.RED = 2;`},
		// Locks in: const declaration is ok by default
		{Code: `const DAY = 86400;`},
		// Locks in: var declaration is ok by default (enforceConst defaults to false)
		{Code: `var HOUR = 3600;`},
		// Locks in: BigInt index >= 0 and < maxArrayLength passes
		{Code: `foo[0xABn]`, Options: map[string]interface{}{"ignoreArrayIndexes": true}},
		// Locks in: Float that loses precision and evaluates to integer passes
		{Code: `foo[5.0000000000000001]`, Options: map[string]interface{}{"ignoreArrayIndexes": true}},
	}, []rule_tester.InvalidTestCase{
		// Locks in: enforceConst reports useConst on non-const declarations
		{
			Code:    `var foo = 42`,
			Options: map[string]interface{}{"enforceConst": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "useConst"}},
		},
		// Locks in: enforceConst with BigInt
		{
			Code:    `var foo = 42n`,
			Options: map[string]interface{}{"enforceConst": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "useConst"}},
		},
		// Locks in: detectObjects reports numbers in object properties
		{
			Code:    `var stats = {avg: 42};`,
			Options: map[string]interface{}{"detectObjects": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 42."}},
		},
		// Locks in: assignment to identifier is NOT ok (even though assignment is normally ok)
		{
			Code:   `min = 1;`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 1."}},
		},
		// Locks in: compound assignment reports
		{
			Code:   `a += 5;`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 5."}},
		},
		// Locks in: numbers inside function bodies are reported
		{
			Code:   `function f() { return 60; }`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 60."}},
		},
	})
}
