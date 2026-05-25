package no_magic_numbers

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoMagicNumbers(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoMagicNumbersRule, []rule_tester.ValidTestCase{
		// ---- Core ESLint: Always allowed ----
		{Code: `var x = parseInt(y, 10);`},
		{Code: `var x = parseInt(y, -10);`},
		{Code: `var x = Number.parseInt(y, 10);`},
		{Code: `const MY_NUMBER = +42;`},
		{Code: `const foo = 42;`},
		{Code: `var foo = 42;`, Options: map[string]interface{}{"enforceConst": false}},
		{Code: `var foo = -42;`},
		{Code: `var foo = 0 + 1 - 2 + -2;`, Options: map[string]interface{}{"ignore": []interface{}{float64(0), float64(1), float64(2), float64(-2)}}},
		{Code: `var foo = 0 + 1 + 2 + 3 + 4;`, Options: map[string]interface{}{"ignore": []interface{}{float64(0), float64(1), float64(2), float64(3), float64(4)}}},
		{Code: `var foo = { bar:10 }`},
		{Code: `setTimeout(function() {return 1;}, 0);`, Options: map[string]interface{}{"ignore": []interface{}{float64(0), float64(1)}}},

		// ---- Core ESLint: ignoreArrayIndexes ----
		{Code: `var data = ['foo', 'bar', 'baz']; var third = data[3];`, Options: map[string]interface{}{"ignoreArrayIndexes": true}},
		{Code: `foo[0]`, Options: map[string]interface{}{"ignoreArrayIndexes": true}},
		{Code: `foo[-0]`, Options: map[string]interface{}{"ignoreArrayIndexes": true}},
		{Code: `foo[1]`, Options: map[string]interface{}{"ignoreArrayIndexes": true}},
		{Code: `foo[100]`, Options: map[string]interface{}{"ignoreArrayIndexes": true}},
		{Code: `foo[200.00]`, Options: map[string]interface{}{"ignoreArrayIndexes": true}},
		{Code: `foo[3e4]`, Options: map[string]interface{}{"ignoreArrayIndexes": true}},
		{Code: `foo[1.23e2]`, Options: map[string]interface{}{"ignoreArrayIndexes": true}},
		{Code: `foo[230e-1]`, Options: map[string]interface{}{"ignoreArrayIndexes": true}},
		{Code: `foo[0xABC]`, Options: map[string]interface{}{"ignoreArrayIndexes": true}},
		{Code: `foo[4294967294]`, Options: map[string]interface{}{"ignoreArrayIndexes": true}},
		{Code: `foo[0n]`, Options: map[string]interface{}{"ignoreArrayIndexes": true}},
		{Code: `foo[-0n]`, Options: map[string]interface{}{"ignoreArrayIndexes": true}},
		{Code: `foo[1n]`, Options: map[string]interface{}{"ignoreArrayIndexes": true}},
		{Code: `foo[100n]`, Options: map[string]interface{}{"ignoreArrayIndexes": true}},
		{Code: `foo[4294967294n]`, Options: map[string]interface{}{"ignoreArrayIndexes": true}},
		{Code: `foo[+0]`, Options: map[string]interface{}{"ignoreArrayIndexes": true}},
		{Code: `foo[+1]`, Options: map[string]interface{}{"ignoreArrayIndexes": true}},
		{Code: `foo[+0n]`, Options: map[string]interface{}{"ignoreArrayIndexes": true}},
		{Code: `foo[+1n]`, Options: map[string]interface{}{"ignoreArrayIndexes": true}},

		// ---- Core ESLint: JSX ----
		{Code: `var a = <input maxLength={10} />;`, FileName: "test.tsx"},
		{Code: `var a = <div objectProp={{ test: 1}}></div>;`, FileName: "test.tsx"},

		// ---- Core ESLint: BigInt ignore ----
		{Code: `f(100n)`, Options: map[string]interface{}{"ignore": []interface{}{"100n"}}},
		{Code: `f(-100n)`, Options: map[string]interface{}{"ignore": []interface{}{"-100n"}}},

		// ---- Core ESLint: ignoreDefaultValues ----
		{Code: `const { param = 123 } = sourceObject;`, Options: map[string]interface{}{"ignoreDefaultValues": true}},
		{Code: `const func = (param = 123) => {}`, Options: map[string]interface{}{"ignoreDefaultValues": true}},
		{Code: `const func = ({ param = 123 }) => {}`, Options: map[string]interface{}{"ignoreDefaultValues": true}},
		{Code: `const [one = 1, two = 2] = []`, Options: map[string]interface{}{"ignoreDefaultValues": true}},
		{Code: `var one, two; [one = 1, two = 2] = []`, Options: map[string]interface{}{"ignoreDefaultValues": true}},

		// ---- Core ESLint: Optional chaining ----
		{Code: `var x = parseInt?.(y, 10);`},
		{Code: `var x = Number?.parseInt(y, 10);`},
		{Code: `var x = (Number?.parseInt)(y, 10);`},
		{Code: `foo?.[777]`, Options: map[string]interface{}{"ignoreArrayIndexes": true}},

		// ---- Core ESLint: ignoreClassFieldInitialValues ----
		{Code: `class C { foo = 2; }`, Options: map[string]interface{}{"ignoreClassFieldInitialValues": true}},
		{Code: `class C { foo = -2; }`, Options: map[string]interface{}{"ignoreClassFieldInitialValues": true}},
		{Code: `class C { static foo = 2; }`, Options: map[string]interface{}{"ignoreClassFieldInitialValues": true}},
		{Code: `class C { #foo = 2; }`, Options: map[string]interface{}{"ignoreClassFieldInitialValues": true}},
		{Code: `class C { static #foo = 2; }`, Options: map[string]interface{}{"ignoreClassFieldInitialValues": true}},

		// ---- TS: ignoreNumericLiteralTypes ----
		{Code: `const FOO = 10;`, Options: map[string]interface{}{"ignoreNumericLiteralTypes": true}},
		{Code: `type Foo = 'bar';`},
		{Code: `type Foo = true;`},
		{Code: `type Foo = 1;`, Options: map[string]interface{}{"ignoreNumericLiteralTypes": true}},
		{Code: `type Foo = -1;`, Options: map[string]interface{}{"ignoreNumericLiteralTypes": true}},
		{Code: `type Foo = 1 | 2 | 3;`, Options: map[string]interface{}{"ignoreNumericLiteralTypes": true}},
		{Code: `type Foo = 1 | -1;`, Options: map[string]interface{}{"ignoreNumericLiteralTypes": true}},

		// ---- TS: ignoreEnums ----
		{
			Code: `
        enum foo {
          SECOND = 1000,
          NUM = '0123456789',
          NEG = -1,
          POS = +1,
        }
      `,
			Options: map[string]interface{}{"ignoreEnums": true},
		},

		// ---- TS: ignoreReadonlyClassProperties ----
		{
			Code: `
class Foo {
  readonly A = 1;
  readonly B = 2;
  public static readonly C = 1;
  static readonly D = 1;
  readonly E = -1;
  readonly F = +1;
  private readonly G = 100n;
}
      `,
			Options: map[string]interface{}{"ignoreReadonlyClassProperties": true},
		},

		// ---- TS: ignoreTypeIndexes ----
		{Code: `type Foo = Bar[0];`, Options: map[string]interface{}{"ignoreTypeIndexes": true}},
		{Code: `type Foo = Bar[-1];`, Options: map[string]interface{}{"ignoreTypeIndexes": true}},
		{Code: `type Foo = Bar[0xab];`, Options: map[string]interface{}{"ignoreTypeIndexes": true}},
		{Code: `type Foo = Bar[5.6e1];`, Options: map[string]interface{}{"ignoreTypeIndexes": true}},
		{Code: `type Foo = Bar[10n];`, Options: map[string]interface{}{"ignoreTypeIndexes": true}},
		{Code: `type Foo = Bar[1 | -2];`, Options: map[string]interface{}{"ignoreTypeIndexes": true}},
		{Code: `type Foo = Bar[1 & -2];`, Options: map[string]interface{}{"ignoreTypeIndexes": true}},
		{Code: `type Foo = Bar[1 & number];`, Options: map[string]interface{}{"ignoreTypeIndexes": true}},
		{Code: `type Foo = Bar[((1 & -2) | 3) | 4];`, Options: map[string]interface{}{"ignoreTypeIndexes": true}},
		{Code: `type Foo = Parameters<Bar>[2];`, Options: map[string]interface{}{"ignoreTypeIndexes": true}},
		{Code: `type Foo = Bar['baz'];`, Options: map[string]interface{}{"ignoreTypeIndexes": true}},
		{Code: `type Foo = Bar['baz'];`, Options: map[string]interface{}{"ignoreTypeIndexes": false}},
		{
			Code: `
type Others = [['a'], ['b']];

type Foo = {
  [K in keyof Others[0]]: Others[K];
};
      `,
			Options: map[string]interface{}{"ignoreTypeIndexes": true},
		},

		// ---- TS: ignore option with type aliases ----
		{Code: `type Foo = 1;`, Options: map[string]interface{}{"ignore": []interface{}{float64(1)}}},
		{Code: `type Foo = -2;`, Options: map[string]interface{}{"ignore": []interface{}{float64(-2)}}},
		{Code: `type Foo = 3n;`, Options: map[string]interface{}{"ignore": []interface{}{"3n"}}},
		{Code: `type Foo = -4n;`, Options: map[string]interface{}{"ignore": []interface{}{"-4n"}}},
		{Code: `type Foo = 5.6;`, Options: map[string]interface{}{"ignore": []interface{}{5.6}}},
		{Code: `type Foo = -7.8;`, Options: map[string]interface{}{"ignore": []interface{}{-7.8}}},
		{Code: `type Foo = 0x0a;`, Options: map[string]interface{}{"ignore": []interface{}{float64(0x0a)}}},
		{Code: `type Foo = -0xbc;`, Options: map[string]interface{}{"ignore": []interface{}{float64(-0xbc)}}},
		{Code: `type Foo = 1e2;`, Options: map[string]interface{}{"ignore": []interface{}{float64(1e2)}}},
		{Code: `type Foo = -3e4;`, Options: map[string]interface{}{"ignore": []interface{}{float64(-3e4)}}},
		{Code: `type Foo = 5e-6;`, Options: map[string]interface{}{"ignore": []interface{}{5e-6}}},
		{Code: `type Foo = -7e-8;`, Options: map[string]interface{}{"ignore": []interface{}{-7e-8}}},
		{Code: `type Foo = 1.1e2;`, Options: map[string]interface{}{"ignore": []interface{}{float64(1.1e2)}}},
		{Code: `type Foo = -3.1e4;`, Options: map[string]interface{}{"ignore": []interface{}{float64(-3.1e4)}}},
		{Code: `type Foo = 5.1e-6;`, Options: map[string]interface{}{"ignore": []interface{}{5.1e-6}}},
		{Code: `type Foo = -7.1e-8;`, Options: map[string]interface{}{"ignore": []interface{}{-7.1e-8}}},

		// ---- TS: ignore with other TS options ----
		{
			Code: `
interface Foo {
  bar: 1;
}
      `,
			Options: map[string]interface{}{"ignore": []interface{}{float64(1)}, "ignoreNumericLiteralTypes": true},
		},
		{
			Code: `
enum foo {
  SECOND = 1000,
  NUM = '0123456789',
  NEG = -1,
  POS = +2,
}
      `,
			Options: map[string]interface{}{"ignore": []interface{}{float64(1000), float64(-1), float64(2)}, "ignoreEnums": false},
		},
		{
			Code: `
class Foo {
  readonly A = 1;
  readonly B = 2;
  public static readonly C = 3;
  static readonly D = 4;
  readonly E = -5;
  readonly F = +6;
  private readonly G = 100n;
  private static readonly H = -2000n;
}
      `,
			Options: map[string]interface{}{
				"ignore":                        []interface{}{float64(1), float64(2), float64(3), float64(4), float64(-5), float64(6), "100n", "-2000n"},
				"ignoreReadonlyClassProperties": false,
			},
		},
		{Code: `type Foo = Bar[0];`, Options: map[string]interface{}{"ignore": []interface{}{float64(0)}, "ignoreTypeIndexes": false}},
		{
			Code: `
type Other = {
  [0]: 3;
};

type Foo = {
  [K in keyof Other]: ` + "`${K & number}`" + `;
};
      `,
			Options: map[string]interface{}{"ignore": []interface{}{float64(0), float64(3)}, "ignoreTypeIndexes": true},
		},

		// ---- JSON path options test (array-wrapped, matches multi-element config shape) ----
		{Code: `type Foo = Bar[0];`, Options: []interface{}{map[string]interface{}{"ignoreTypeIndexes": true}}},

		// ---- Binary / octal literal array indexes (upstream core) ----
		{Code: `foo[0b110]`, Options: map[string]interface{}{"ignoreArrayIndexes": true}},
		{Code: `foo[0o71]`, Options: map[string]interface{}{"ignoreArrayIndexes": true}},

		// ---- ParenthesizedExpression (tsgo-specific, ESTree has no paren nodes) ----
		{Code: `const X = (42);`},
		{Code: `const X = ((42));`},
		{Code: `var x = { foo: (42) };`},
		{Code: `obj.prop = (1);`},
		{Code: `parseInt(y, (10));`},
		{Code: `foo[(0)]`, Options: map[string]interface{}{"ignoreArrayIndexes": true}},
		{Code: `enum E { A = (42) }`, Options: map[string]interface{}{"ignoreEnums": true}},
		{Code: `class C { readonly x = (42); }`, Options: map[string]interface{}{"ignoreReadonlyClassProperties": true}},
		{Code: `class C { foo = (2); }`, Options: map[string]interface{}{"ignoreClassFieldInitialValues": true}},
		{Code: `const func = (param = (123)) => {}`, Options: map[string]interface{}{"ignoreDefaultValues": true}},
		{Code: `var a = <input maxLength={(10)} />;`, FileName: "test.tsx"},
		{Code: `const X = -(42);`, Options: map[string]interface{}{"ignore": []interface{}{float64(-42)}}},

		// ---- Edge cases: tsgo AST shapes ----
		{Code: `type Nested = ('' | ('' | (1)));`, Options: map[string]interface{}{"ignoreNumericLiteralTypes": true}},
		{Code: `type Foo = Bar[((((1))))];`, Options: map[string]interface{}{"ignoreTypeIndexes": true}},
		{Code: `enum E { A = 1 << 0, B = 1 << 1 }`, Options: map[string]interface{}{"ignoreEnums": true, "ignore": []interface{}{float64(0), float64(1)}}},
		{Code: `class C { protected readonly x = 42; }`, Options: map[string]interface{}{"ignoreReadonlyClassProperties": true}},
		{Code: `foo?.bar?.[0]`, Options: map[string]interface{}{"ignoreArrayIndexes": true}},
		{Code: `var x = { foo: 42 };`},
		{Code: `var colors = {}; colors.RED = 2;`},
		{Code: `const { a: { b = 1 } } = obj;`, Options: map[string]interface{}{"ignoreDefaultValues": true}},
		{Code: `class C { foo = -42; }`, Options: map[string]interface{}{"ignoreClassFieldInitialValues": true}},
		{Code: `class C { readonly x = 100n; }`, Options: map[string]interface{}{"ignoreReadonlyClassProperties": true}},
		{Code: `var x = {[42]: true}`},
		{Code: `var one; ({one = 1} = {})`, Options: map[string]interface{}{"ignoreDefaultValues": true}},
		{Code: `var a, b; ({a = 1, b = 2} = {})`, Options: map[string]interface{}{"ignoreDefaultValues": true}},
		{Code: `var x; ({a: x = 42} = {})`, Options: map[string]interface{}{"ignoreDefaultValues": true}},

		// ---- Upstream semantic lock-in ----
		{Code: `var stats = {avg: 42};`},
		{Code: `({key: 90, another: 10})`},
		{Code: `colors.RED = 2;`},
		{Code: `const DAY = 86400;`},
		{Code: `var HOUR = 3600;`},
		{Code: `foo[0xABn]`, Options: map[string]interface{}{"ignoreArrayIndexes": true}},
		{Code: `foo[5.0000000000000001]`, Options: map[string]interface{}{"ignoreArrayIndexes": true}},
	}, []rule_tester.InvalidTestCase{
		// ---- Core ESLint: enforceConst ----
		{
			Code:    `var foo = 42`,
			Options: map[string]interface{}{"enforceConst": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "useConst"}},
		},
		// ---- Core ESLint: basic violations ----
		{
			Code: `var foo = 0 + 1;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noMagic", Message: "No magic number: 0.", Line: 1, Column: 11},
				{MessageId: "noMagic", Message: "No magic number: 1.", Line: 1, Column: 15},
			},
		},
		// BigInt enforceConst
		{
			Code:    `var foo = 42n`,
			Options: map[string]interface{}{"enforceConst": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "useConst"}},
		},
		// BigInt violations
		{
			Code: `var foo = 0n + 1n;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noMagic", Message: "No magic number: 0n.", Line: 1, Column: 11},
				{MessageId: "noMagic", Message: "No magic number: 1n.", Line: 1, Column: 16},
			},
		},
		// Assignment
		{
			Code:   `a = a + 5;`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 5.", Line: 1, Column: 9}},
		},
		// Compound assignment
		{
			Code:   `a += 5;`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 5.", Line: 1, Column: 6}},
		},
		// Multiple with negatives
		{
			Code: `var foo = 0 + 1 + -2 + 2;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noMagic", Message: "No magic number: 0.", Line: 1, Column: 11},
				{MessageId: "noMagic", Message: "No magic number: 1.", Line: 1, Column: 15},
				{MessageId: "noMagic", Message: "No magic number: -2.", Line: 1, Column: 19},
				{MessageId: "noMagic", Message: "No magic number: 2.", Line: 1, Column: 24},
			},
		},
		// Partial ignore
		{
			Code:    `var foo = 0 + 1 + 2;`,
			Options: map[string]interface{}{"ignore": []interface{}{float64(0), float64(1)}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 2."}},
		},
		// detectObjects
		{
			Code:    `var foo = { bar:10 }`,
			Options: map[string]interface{}{"detectObjects": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 10."}},
		},
		// Functions
		{
			Code:   `function getSecondsInMinute() {return 60;}`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 60."}},
		},
		// Negative in function
		{
			Code:   `function getNegativeSecondsInMinute() {return -60;}`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: -60."}},
		},
		// Array index violations
		{
			Code:   `var data = ['foo', 'bar', 'baz']; var third = data[3];`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 3."}},
		},
		// Negative array index (not a valid index)
		{
			Code:    `foo[-100]`,
			Options: map[string]interface{}{"ignoreArrayIndexes": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: -100."}},
		},
		{
			Code:    `foo[-1]`,
			Options: map[string]interface{}{"ignoreArrayIndexes": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: -1."}},
		},
		// Non-integer array index
		{
			Code:    `foo[0.1]`,
			Options: map[string]interface{}{"ignoreArrayIndexes": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 0.1."}},
		},
		{
			Code:    `foo[1.5]`,
			Options: map[string]interface{}{"ignoreArrayIndexes": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 1.5."}},
		},
		// Above max array index
		{
			Code:    `foo[4294967295]`,
			Options: map[string]interface{}{"ignoreArrayIndexes": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 4294967295."}},
		},
		// Negative BigInt array index
		{
			Code:    `foo[-1n]`,
			Options: map[string]interface{}{"ignoreArrayIndexes": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: -1n."}},
		},
		// BigInt above max index
		{
			Code:    `foo[4294967295n]`,
			Options: map[string]interface{}{"ignoreArrayIndexes": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 4294967295n."}},
		},
		// Assignment to identifier
		{
			Code: `var min, max, mean; min = 1; max = 10; mean = 4;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noMagic", Message: "No magic number: 1."},
				{MessageId: "noMagic", Message: "No magic number: 10."},
				{MessageId: "noMagic", Message: "No magic number: 4."},
			},
		},
		// BigInt ignore mismatch
		{
			Code:    `f(100n)`,
			Options: map[string]interface{}{"ignore": []interface{}{float64(100)}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 100n."}},
		},
		{
			Code:    `f(-100n)`,
			Options: map[string]interface{}{"ignore": []interface{}{"100n"}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: -100n."}},
		},
		{
			Code:    `f(100n)`,
			Options: map[string]interface{}{"ignore": []interface{}{"-100n"}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 100n."}},
		},
		{
			Code:    `f(100)`,
			Options: map[string]interface{}{"ignore": []interface{}{"100n"}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 100."}},
		},
		// ignoreDefaultValues: false
		{
			Code:    `const func = (param = 123) => {}`,
			Options: map[string]interface{}{"ignoreDefaultValues": false},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 123."}},
		},
		{
			Code:    `const { param = 123 } = sourceObject;`,
			Options: map[string]interface{}{"ignoreDefaultValues": false},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 123."}},
		},

		// ---- ignoreClassFieldInitialValues ----
		{
			Code:   `class C { foo = 2; }`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 2.", Line: 1, Column: 17}},
		},
		{
			Code:    `class C { foo = -2; }`,
			Options: map[string]interface{}{"ignoreClassFieldInitialValues": false},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: -2.", Line: 1, Column: 17}},
		},
		// Expression in class field
		{
			Code:    `class C { foo = 2 + 3; }`,
			Options: map[string]interface{}{"ignoreClassFieldInitialValues": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noMagic", Message: "No magic number: 2.", Line: 1, Column: 17},
				{MessageId: "noMagic", Message: "No magic number: 3.", Line: 1, Column: 21},
			},
		},

		// ---- TS: ignoreNumericLiteralTypes: false ----
		{
			Code:    `type Foo = 1;`,
			Options: map[string]interface{}{"ignoreNumericLiteralTypes": false},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 1.", Line: 1, Column: 12}},
		},
		{
			Code:    `type Foo = -1;`,
			Options: map[string]interface{}{"ignoreNumericLiteralTypes": false},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: -1.", Line: 1, Column: 12}},
		},
		{
			Code:    `type Foo = 1 | 2 | 3;`,
			Options: map[string]interface{}{"ignoreNumericLiteralTypes": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noMagic", Message: "No magic number: 1.", Line: 1, Column: 12},
				{MessageId: "noMagic", Message: "No magic number: 2.", Line: 1, Column: 16},
				{MessageId: "noMagic", Message: "No magic number: 3.", Line: 1, Column: 20},
			},
		},
		{
			Code:    `type Foo = 1 | -1;`,
			Options: map[string]interface{}{"ignoreNumericLiteralTypes": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noMagic", Message: "No magic number: 1.", Line: 1, Column: 12},
				{MessageId: "noMagic", Message: "No magic number: -1.", Line: 1, Column: 16},
			},
		},
		// Interface property (not a type alias)
		{
			Code: `
interface Foo {
  bar: 1;
}
      `,
			Options: map[string]interface{}{"ignoreNumericLiteralTypes": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 1.", Line: 3, Column: 8}},
		},

		// ---- TS: ignoreEnums: false ----
		{
			Code: `
enum foo {
  SECOND = 1000,
  NUM = '0123456789',
  NEG = -1,
  POS = +1,
}
      `,
			Options: map[string]interface{}{"ignoreEnums": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noMagic", Message: "No magic number: 1000.", Line: 3, Column: 12},
				{MessageId: "noMagic", Message: "No magic number: -1.", Line: 5, Column: 9},
				{MessageId: "noMagic", Message: "No magic number: 1.", Line: 6, Column: 10},
			},
		},

		// ---- TS: ignoreReadonlyClassProperties: false ----
		{
			Code: `
class Foo {
  readonly A = 1;
  readonly B = 2;
  public static readonly C = 3;
  static readonly D = 4;
  readonly E = -5;
  readonly F = +6;
  private readonly G = 100n;
}
      `,
			Options: map[string]interface{}{"ignoreReadonlyClassProperties": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noMagic", Message: "No magic number: 1.", Line: 3, Column: 16},
				{MessageId: "noMagic", Message: "No magic number: 2.", Line: 4, Column: 16},
				{MessageId: "noMagic", Message: "No magic number: 3.", Line: 5, Column: 30},
				{MessageId: "noMagic", Message: "No magic number: 4.", Line: 6, Column: 23},
				{MessageId: "noMagic", Message: "No magic number: -5.", Line: 7, Column: 16},
				{MessageId: "noMagic", Message: "No magic number: 6.", Line: 8, Column: 17},
				{MessageId: "noMagic", Message: "No magic number: 100n.", Line: 9, Column: 24},
			},
		},

		// ---- TS: ignoreTypeIndexes: false ----
		{
			Code:    `type Foo = Bar[0];`,
			Options: map[string]interface{}{"ignoreTypeIndexes": false},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 0.", Line: 1, Column: 16}},
		},
		{
			Code:    `type Foo = Bar[-1];`,
			Options: map[string]interface{}{"ignoreTypeIndexes": false},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: -1.", Line: 1, Column: 16}},
		},
		{
			Code:    `type Foo = Bar[0xab];`,
			Options: map[string]interface{}{"ignoreTypeIndexes": false},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 0xab.", Line: 1, Column: 16}},
		},
		{
			Code:    `type Foo = Bar[5.6e1];`,
			Options: map[string]interface{}{"ignoreTypeIndexes": false},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 5.6e1.", Line: 1, Column: 16}},
		},
		{
			Code:    `type Foo = Bar[10n];`,
			Options: map[string]interface{}{"ignoreTypeIndexes": false},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 10n.", Line: 1, Column: 16}},
		},
		{
			Code:    `type Foo = Bar[1 | -2];`,
			Options: map[string]interface{}{"ignoreTypeIndexes": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noMagic", Message: "No magic number: 1.", Line: 1, Column: 16},
				{MessageId: "noMagic", Message: "No magic number: -2.", Line: 1, Column: 20},
			},
		},
		{
			Code:    `type Foo = Bar[1 & -2];`,
			Options: map[string]interface{}{"ignoreTypeIndexes": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noMagic", Message: "No magic number: 1.", Line: 1, Column: 16},
				{MessageId: "noMagic", Message: "No magic number: -2.", Line: 1, Column: 20},
			},
		},
		{
			Code:    `type Foo = Bar[1 & number];`,
			Options: map[string]interface{}{"ignoreTypeIndexes": false},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 1.", Line: 1, Column: 16}},
		},
		{
			Code:    `type Foo = Bar[((1 & -2) | 3) | 4];`,
			Options: map[string]interface{}{"ignoreTypeIndexes": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noMagic", Message: "No magic number: 1.", Line: 1, Column: 18},
				{MessageId: "noMagic", Message: "No magic number: -2.", Line: 1, Column: 22},
				{MessageId: "noMagic", Message: "No magic number: 3.", Line: 1, Column: 28},
				{MessageId: "noMagic", Message: "No magic number: 4.", Line: 1, Column: 33},
			},
		},
		{
			Code:    `type Foo = Parameters<Bar>[2];`,
			Options: map[string]interface{}{"ignoreTypeIndexes": false},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 2.", Line: 1, Column: 28}},
		},
		{
			Code: `
type Others = [['a'], ['b']];

type Foo = {
  [K in keyof Others[0]]: Others[K];
};
      `,
			Options: map[string]interface{}{"ignoreTypeIndexes": false},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 0.", Line: 5, Column: 22}},
		},
		// Type index in non-indexed context
		{
			Code: `
type Other = {
  [0]: 3;
};

type Foo = {
  [K in keyof Other]: ` + "`${K & number}`" + `;
};
      `,
			Options: map[string]interface{}{"ignoreTypeIndexes": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noMagic", Message: "No magic number: 0.", Line: 3, Column: 4},
				{MessageId: "noMagic", Message: "No magic number: 3.", Line: 3, Column: 8},
			},
		},
		// Mapped type
		{
			Code: `
type Foo = {
  [K in 0 | 1 | 2]: 0;
};
      `,
			Options: map[string]interface{}{"ignoreTypeIndexes": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noMagic", Message: "No magic number: 0.", Line: 3, Column: 9},
				{MessageId: "noMagic", Message: "No magic number: 1.", Line: 3, Column: 13},
				{MessageId: "noMagic", Message: "No magic number: 2.", Line: 3, Column: 17},
				{MessageId: "noMagic", Message: "No magic number: 0.", Line: 3, Column: 21},
			},
		},

		// ---- TS: ignore option with negation mismatch ----
		{
			Code:    `type Foo = 1;`,
			Options: map[string]interface{}{"ignore": []interface{}{float64(-1)}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 1.", Line: 1, Column: 12}},
		},
		{
			Code:    `type Foo = -2;`,
			Options: map[string]interface{}{"ignore": []interface{}{float64(2)}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: -2.", Line: 1, Column: 12}},
		},
		{
			Code:    `type Foo = 3n;`,
			Options: map[string]interface{}{"ignore": []interface{}{"-3n"}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 3n.", Line: 1, Column: 12}},
		},
		{
			Code:    `type Foo = -4n;`,
			Options: map[string]interface{}{"ignore": []interface{}{"4n"}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: -4n.", Line: 1, Column: 12}},
		},
		{
			Code:    `type Foo = 5.6;`,
			Options: map[string]interface{}{"ignore": []interface{}{-5.6}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 5.6.", Line: 1, Column: 12}},
		},
		{
			Code:    `type Foo = -7.8;`,
			Options: map[string]interface{}{"ignore": []interface{}{7.8}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: -7.8.", Line: 1, Column: 12}},
		},
		{
			Code:    `type Foo = 0x0a;`,
			Options: map[string]interface{}{"ignore": []interface{}{float64(-0x0a)}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 0x0a.", Line: 1, Column: 12}},
		},
		{
			Code:    `type Foo = -0xbc;`,
			Options: map[string]interface{}{"ignore": []interface{}{float64(0xbc)}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: -0xbc.", Line: 1, Column: 12}},
		},
		{
			Code:    `type Foo = 1e2;`,
			Options: map[string]interface{}{"ignore": []interface{}{float64(-1e2)}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 1e2.", Line: 1, Column: 12}},
		},
		{
			Code:    `type Foo = -3e4;`,
			Options: map[string]interface{}{"ignore": []interface{}{float64(3e4)}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: -3e4.", Line: 1, Column: 12}},
		},
		{
			Code:    `type Foo = 5e-6;`,
			Options: map[string]interface{}{"ignore": []interface{}{-5e-6}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 5e-6.", Line: 1, Column: 12}},
		},
		{
			Code:    `type Foo = -7e-8;`,
			Options: map[string]interface{}{"ignore": []interface{}{7e-8}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: -7e-8.", Line: 1, Column: 12}},
		},
		{
			Code:    `type Foo = 1.1e2;`,
			Options: map[string]interface{}{"ignore": []interface{}{float64(-1.1e2)}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 1.1e2.", Line: 1, Column: 12}},
		},
		{
			Code:    `type Foo = -3.1e4;`,
			Options: map[string]interface{}{"ignore": []interface{}{float64(3.1e4)}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: -3.1e4.", Line: 1, Column: 12}},
		},
		{
			Code:    `type Foo = 5.1e-6;`,
			Options: map[string]interface{}{"ignore": []interface{}{-5.1e-6}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 5.1e-6.", Line: 1, Column: 12}},
		},
		{
			Code:    `type Foo = -7.1e-8;`,
			Options: map[string]interface{}{"ignore": []interface{}{7.1e-8}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: -7.1e-8.", Line: 1, Column: 12}},
		},

		// ---- ParenthesizedExpression invalid ----
		{
			Code:   `function f() { return -(1); }`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noMagic"}},
		},
		{
			Code:   `a = (1);`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 1."}},
		},

		// ---- Edge case invalid: type literal property (NOT type alias) ----
		{
			Code:    `type Foo = { bar: 42 };`,
			Options: map[string]interface{}{"ignoreNumericLiteralTypes": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 42.", Line: 1, Column: 19}},
		},
		{
			Code:    `type Foo = { bar: 2 | 3 };`,
			Options: map[string]interface{}{"ignoreNumericLiteralTypes": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noMagic", Message: "No magic number: 2.", Line: 1, Column: 19},
				{MessageId: "noMagic", Message: "No magic number: 3.", Line: 1, Column: 23},
			},
		},
		// ---- Edge case invalid: class computed key vs initializer ----
		{
			Code:    `class C { 2; }`,
			Options: map[string]interface{}{"ignoreClassFieldInitialValues": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 2.", Line: 1, Column: 11}},
		},
		{
			Code:    `class C { [2]; }`,
			Options: map[string]interface{}{"ignoreClassFieldInitialValues": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 2.", Line: 1, Column: 12}},
		},
		// ---- Edge case invalid: object destructuring default ----
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

		// ---- Upstream core invalid: array index variants ----
		{Code: `foo[-0.1]`, Options: map[string]interface{}{"ignoreArrayIndexes": true}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: -0.1."}}},
		{Code: `foo[-0b110]`, Options: map[string]interface{}{"ignoreArrayIndexes": true}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: -0b110."}}},
		{Code: `foo[-0o71]`, Options: map[string]interface{}{"ignoreArrayIndexes": true}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: -0o71."}}},
		{Code: `foo[-0x12]`, Options: map[string]interface{}{"ignoreArrayIndexes": true}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: -0x12."}}},
		{Code: `foo[0.12e1]`, Options: map[string]interface{}{"ignoreArrayIndexes": true}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 0.12e1."}}},
		{Code: `foo[1.678e2]`, Options: map[string]interface{}{"ignoreArrayIndexes": true}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 1.678e2."}}},
		{Code: `foo[100.9]`, Options: map[string]interface{}{"ignoreArrayIndexes": true}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 100.9."}}},
		{Code: `foo[1e300]`, Options: map[string]interface{}{"ignoreArrayIndexes": true}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 1e300."}}},
		{Code: `foo[1e310]`, Options: map[string]interface{}{"ignoreArrayIndexes": true}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 1e310."}}},
		{Code: `foo[-1e310]`, Options: map[string]interface{}{"ignoreArrayIndexes": true}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: -1e310."}}},
		{Code: `foo[-0x12n]`, Options: map[string]interface{}{"ignoreArrayIndexes": true}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: -0x12n."}}},
		{Code: `foo[- -1n]`, Options: map[string]interface{}{"ignoreArrayIndexes": true}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: -1n."}}},

		// ---- Upstream core invalid: default values ----
		{Code: `const { param = 123 } = sourceObject;`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 123."}}},
		{Code: `const { param = 123 } = sourceObject;`, Options: map[string]interface{}{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 123."}}},
		{Code: `const [one = 1, two = 2] = []`, Options: map[string]interface{}{"ignoreDefaultValues": false}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 1."}, {MessageId: "noMagic", Message: "No magic number: 2."}}},
		{Code: `var one, two; [one = 1, two = 2] = []`, Options: map[string]interface{}{"ignoreDefaultValues": false}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 1."}, {MessageId: "noMagic", Message: "No magic number: 2."}}},

		// ---- Upstream core invalid: class field variants ----
		{Code: `class C { static foo = 2; }`, Options: map[string]interface{}{"ignoreClassFieldInitialValues": false}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 2.", Line: 1, Column: 24}}},
		{Code: `class C { #foo = 2; }`, Options: map[string]interface{}{"ignoreClassFieldInitialValues": false}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 2.", Line: 1, Column: 18}}},
		{Code: `class C { static #foo = 2; }`, Options: map[string]interface{}{"ignoreClassFieldInitialValues": false}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 2.", Line: 1, Column: 25}}},
		{Code: `class C { foo = 2; }`, Options: map[string]interface{}{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 2.", Line: 1, Column: 17}}},

		// ---- Upstream core invalid: hex in expressions ----
		{Code: `console.log(0x1A + 0x02);`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 0x1A."}, {MessageId: "noMagic", Message: "No magic number: 0x02."}}},

		// ---- Upstream core invalid: misc ----
		{Code: `var colors = {}; colors.RED = 2; colors.YELLOW = 3; colors.BLUE = 4 + 5;`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 4."}, {MessageId: "noMagic", Message: "No magic number: 5."}}},
		{Code: `var a = <div arrayProp={[1,2,3]}></div>;`, FileName: "test.tsx", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 1."}, {MessageId: "noMagic", Message: "No magic number: 2."}, {MessageId: "noMagic", Message: "No magic number: 3."}}},

		// ---- Upstream semantic lock-in invalid ----
		{Code: `var stats = {avg: 42};`, Options: map[string]interface{}{"detectObjects": true}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 42."}}},
		{Code: `min = 1;`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 1."}}},
		{Code: `function f() { return 60; }`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 60."}}},
	})
}
