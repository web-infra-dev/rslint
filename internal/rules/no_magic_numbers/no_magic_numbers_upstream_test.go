// TestNoMagicNumbersUpstream migrates the full valid/invalid suite from
// ESLint v10.8.1 tests/lib/rules/no-magic-numbers.js 1:1, including the
// trailing `ruleTesterTypeScript` block (upstream runs the exact same rule
// through @typescript-eslint/parser there; rslint always parses TypeScript
// syntax, so both blocks apply equally here). languageOptions.ecmaVersion is
// dropped from ported cases: rslint's parser does not gate syntax on it the
// way espree does. languageOptions.sourceType is dropped too, except where a
// case specifically exercises legacy (non-strict) octal literal syntax
// (`0123`, `071`), which rslint's parser rejects regardless of module-ness;
// those cases are omitted — see no_magic_numbers_extras_test.go for the
// rslint-specific coverage that replaces them.
package no_magic_numbers

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoMagicNumbersUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoMagicNumbersRule, []rule_tester.ValidTestCase{
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

		{Code: `var data = ['foo', 'bar', 'baz']; var third = data[3];`, Options: map[string]interface{}{"ignoreArrayIndexes": true}},
		{Code: `foo[0]`, Options: map[string]interface{}{"ignoreArrayIndexes": true}},
		{Code: `foo[-0]`, Options: map[string]interface{}{"ignoreArrayIndexes": true}}, // -0 is coerced to "0", so foo[-0] refers to the element at index 0.
		{Code: `foo[1]`, Options: map[string]interface{}{"ignoreArrayIndexes": true}},
		{Code: `foo[100]`, Options: map[string]interface{}{"ignoreArrayIndexes": true}},
		{Code: `foo[200.00]`, Options: map[string]interface{}{"ignoreArrayIndexes": true}},
		{Code: `foo[3e4]`, Options: map[string]interface{}{"ignoreArrayIndexes": true}},
		{Code: `foo[1.23e2]`, Options: map[string]interface{}{"ignoreArrayIndexes": true}},
		{Code: `foo[230e-1]`, Options: map[string]interface{}{"ignoreArrayIndexes": true}},
		{Code: `foo[0b110]`, Options: map[string]interface{}{"ignoreArrayIndexes": true}},
		{Code: `foo[0o71]`, Options: map[string]interface{}{"ignoreArrayIndexes": true}},
		{Code: `foo[0xABC]`, Options: map[string]interface{}{"ignoreArrayIndexes": true}},
		{Code: `foo[5.0000000000000001]`, Options: map[string]interface{}{"ignoreArrayIndexes": true}}, // loses precision and evaluates to 5
		{Code: `foo[4294967294]`, Options: map[string]interface{}{"ignoreArrayIndexes": true}},         // max array index
		{Code: `foo[0n]`, Options: map[string]interface{}{"ignoreArrayIndexes": true}},                 // 0n is coerced to "0", so foo[0n] refers to the element at index 0.
		{Code: `foo[-0n]`, Options: map[string]interface{}{"ignoreArrayIndexes": true}},                // -0n evaluates to 0n which is coerced to "0".
		{Code: `foo[1n]`, Options: map[string]interface{}{"ignoreArrayIndexes": true}},
		{Code: `foo[100n]`, Options: map[string]interface{}{"ignoreArrayIndexes": true}},
		{Code: `foo[0xABn]`, Options: map[string]interface{}{"ignoreArrayIndexes": true}}, // evaluates to 171n, coerced to "171".
		{Code: `foo[4294967294n]`, Options: map[string]interface{}{"ignoreArrayIndexes": true}},

		{Code: `var a = <input maxLength={10} />;`, FileName: "src/virtual.tsx"},
		{Code: `var a = <div objectProp={{ test: 1}}></div>;`, FileName: "src/virtual.tsx"},

		{Code: `f(100n)`, Options: map[string]interface{}{"ignore": []interface{}{"100n"}}},
		{Code: `f(-100n)`, Options: map[string]interface{}{"ignore": []interface{}{"-100n"}}},

		{Code: `const { param = 123 } = sourceObject;`, Options: map[string]interface{}{"ignoreDefaultValues": true}},
		{Code: `const func = (param = 123) => {}`, Options: map[string]interface{}{"ignoreDefaultValues": true}},
		{Code: `const func = ({ param = 123 }) => {}`, Options: map[string]interface{}{"ignoreDefaultValues": true}},
		{Code: `const [one = 1, two = 2] = []`, Options: map[string]interface{}{"ignoreDefaultValues": true}},
		{Code: `var one, two; [one = 1, two = 2] = []`, Options: map[string]interface{}{"ignoreDefaultValues": true}},

		// Optional chaining
		{Code: `var x = parseInt?.(y, 10);`},
		{Code: `var x = Number?.parseInt(y, 10);`},
		{Code: `var x = (Number?.parseInt)(y, 10);`},
		{Code: `foo?.[777]`, Options: map[string]interface{}{"ignoreArrayIndexes": true}},

		// ignoreClassFieldInitialValues
		{Code: `class C { foo = 2; }`, Options: map[string]interface{}{"ignoreClassFieldInitialValues": true}},
		{Code: `class C { foo = -2; }`, Options: map[string]interface{}{"ignoreClassFieldInitialValues": true}},
		{Code: `class C { static foo = 2; }`, Options: map[string]interface{}{"ignoreClassFieldInitialValues": true}},
		{Code: `class C { #foo = 2; }`, Options: map[string]interface{}{"ignoreClassFieldInitialValues": true}},
		{Code: `class C { static #foo = 2; }`, Options: map[string]interface{}{"ignoreClassFieldInitialValues": true}},

		{Code: `foo[+0]`, Options: map[string]interface{}{"ignoreArrayIndexes": true}}, // consistent with the default behavior, which allows: var foo = +0
		{Code: `foo[+1]`, Options: map[string]interface{}{"ignoreArrayIndexes": true}},
		{Code: `foo[+0n]`, Options: map[string]interface{}{"ignoreArrayIndexes": true}},
		{Code: `foo[+1n]`, Options: map[string]interface{}{"ignoreArrayIndexes": true}},

		// ---- TypeScript parser suite (upstream ruleTesterTypeScript) ----
		{Code: `const FOO = 10;`, Options: map[string]interface{}{"ignoreNumericLiteralTypes": true}},
		{Code: `type Foo = 'bar';`},
		{Code: `type Foo = true;`},
		{Code: `type Foo = 1;`, Options: map[string]interface{}{"ignoreNumericLiteralTypes": true}},
		{Code: `type Foo = -1;`, Options: map[string]interface{}{"ignoreNumericLiteralTypes": true}},
		{Code: `type Nested = ('' | ('' | (1)));`, Options: map[string]interface{}{"ignoreNumericLiteralTypes": true}},
		{Code: `type Foo = 1 | 2 | 3;`, Options: map[string]interface{}{"ignoreNumericLiteralTypes": true}},
		{Code: `type Foo = 1 | -1;`, Options: map[string]interface{}{"ignoreNumericLiteralTypes": true}},
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
		{
			Code: `
class C {
	readonly foo = +42;
	bar = +42;
}

const MY_NUMBER = +42;
`,
			Options: map[string]interface{}{
				"ignoreClassFieldInitialValues": true,
				"ignoreReadonlyClassProperties": true,
			},
		},
	}, []rule_tester.InvalidTestCase{
		{
			Code:    `var foo = 42`,
			Options: map[string]interface{}{"enforceConst": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "useConst"}},
		},
		{
			Code: `var foo = 0 + 1;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noMagic", Message: "No magic number: 0."},
				{MessageId: "noMagic", Message: "No magic number: 1."},
			},
		},
		{
			Code:    `var foo = 42n`,
			Options: map[string]interface{}{"enforceConst": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "useConst"}},
		},
		{
			Code: `var foo = 0n + 1n;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noMagic", Message: "No magic number: 0n."},
				{MessageId: "noMagic", Message: "No magic number: 1n."},
			},
		},
		{
			Code:   `a = a + 5;`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 5."}},
		},
		{
			Code:   `a += 5;`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 5."}},
		},
		{
			Code: `var foo = 0 + 1 + -2 + 2;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noMagic", Message: "No magic number: 0."},
				{MessageId: "noMagic", Message: "No magic number: 1."},
				{MessageId: "noMagic", Message: "No magic number: -2."},
				{MessageId: "noMagic", Message: "No magic number: 2."},
			},
		},
		{
			Code:    `var foo = 0 + 1 + 2;`,
			Options: map[string]interface{}{"ignore": []interface{}{float64(0), float64(1)}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 2."}},
		},
		{
			Code:    `var foo = { bar:10 }`,
			Options: map[string]interface{}{"detectObjects": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 10."}},
		},
		{
			Code: `var stats = {avg: 42};`,
			Options: map[string]interface{}{
				"detectObjects": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 42."}},
		},
		{
			Code: `var colors = {}; colors.RED = 2; colors.YELLOW = 3; colors.BLUE = 4 + 5;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noMagic", Message: "No magic number: 4."},
				{MessageId: "noMagic", Message: "No magic number: 5."},
			},
		},
		{
			Code:   `function getSecondsInMinute() {return 60;}`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 60."}},
		},
		{
			Code:   `function getNegativeSecondsInMinute() {return -60;}`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: -60."}},
		},
		{
			Code: `var Promise = require('bluebird');
var MINUTE = 60;
var HOUR = 3600;
const DAY = 86400;
var configObject = {
key: 90,
another: 10 * 10,
10: 'an "integer" key'
};
function getSecondsInDay() {
 return 24 * HOUR;
}
function getMillisecondsInDay() {
return (getSecondsInDay() *
(1000)
);
}
function callSetTimeoutZero(func) {
setTimeout(func, 0);
}
function invokeInTen(func) {
setTimeout(func, 10);
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noMagic", Message: "No magic number: 10.", Line: 7},
				{MessageId: "noMagic", Message: "No magic number: 10.", Line: 7},
				{MessageId: "noMagic", Message: "No magic number: 24.", Line: 11},
				{MessageId: "noMagic", Message: "No magic number: 1000.", Line: 15},
				{MessageId: "noMagic", Message: "No magic number: 0.", Line: 19},
				{MessageId: "noMagic", Message: "No magic number: 10.", Line: 22},
			},
		},
		{
			Code:   `var data = ['foo', 'bar', 'baz']; var third = data[3];`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 3.", Line: 1}},
		},
		{
			Code:    `var data = ['foo', 'bar', 'baz']; var third = data[3];`,
			Options: map[string]interface{}{},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 3.", Line: 1}},
		},
		{
			Code:    `var data = ['foo', 'bar', 'baz']; var third = data[3];`,
			Options: map[string]interface{}{"ignoreArrayIndexes": false},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 3.", Line: 1}},
		},
		{
			Code:    `foo[-100]`,
			Options: map[string]interface{}{"ignoreArrayIndexes": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: -100.", Line: 1}},
		},
		{
			Code:    `foo[-1.5]`,
			Options: map[string]interface{}{"ignoreArrayIndexes": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: -1.5.", Line: 1}},
		},
		{
			Code:    `foo[-1]`,
			Options: map[string]interface{}{"ignoreArrayIndexes": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: -1.", Line: 1}},
		},
		{
			Code:    `foo[-0.1]`,
			Options: map[string]interface{}{"ignoreArrayIndexes": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: -0.1.", Line: 1}},
		},
		{
			Code:    `foo[-0b110]`,
			Options: map[string]interface{}{"ignoreArrayIndexes": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: -0b110.", Line: 1}},
		},
		{
			Code:    `foo[-0o71]`,
			Options: map[string]interface{}{"ignoreArrayIndexes": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: -0o71.", Line: 1}},
		},
		{
			Code:    `foo[-0x12]`,
			Options: map[string]interface{}{"ignoreArrayIndexes": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: -0x12.", Line: 1}},
		},
		{
			Code:    `foo[0.1]`,
			Options: map[string]interface{}{"ignoreArrayIndexes": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 0.1.", Line: 1}},
		},
		{
			Code:    `foo[0.12e1]`,
			Options: map[string]interface{}{"ignoreArrayIndexes": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 0.12e1.", Line: 1}},
		},
		{
			Code:    `foo[1.5]`,
			Options: map[string]interface{}{"ignoreArrayIndexes": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 1.5.", Line: 1}},
		},
		{
			Code:    `foo[1.678e2]`, // 167.8
			Options: map[string]interface{}{"ignoreArrayIndexes": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 1.678e2.", Line: 1}},
		},
		{
			Code:    `foo[56e-1]`, // 5.6
			Options: map[string]interface{}{"ignoreArrayIndexes": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 56e-1.", Line: 1}},
		},
		{
			Code:    `foo[5.000000000000001]`, // doesn't lose precision
			Options: map[string]interface{}{"ignoreArrayIndexes": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 5.000000000000001.", Line: 1}},
		},
		{
			Code:    `foo[100.9]`,
			Options: map[string]interface{}{"ignoreArrayIndexes": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 100.9.", Line: 1}},
		},
		{
			Code:    `foo[4294967295]`, // first above the max index
			Options: map[string]interface{}{"ignoreArrayIndexes": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 4294967295.", Line: 1}},
		},
		{
			Code:    `foo[1e300]`, // above the max, and also coerces to "1e+300"
			Options: map[string]interface{}{"ignoreArrayIndexes": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 1e300.", Line: 1}},
		},
		{
			Code:    `foo[1e310]`, // refers to property "Infinity"
			Options: map[string]interface{}{"ignoreArrayIndexes": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 1e310.", Line: 1}},
		},
		{
			Code:    `foo[-1e310]`, // refers to property "-Infinity"
			Options: map[string]interface{}{"ignoreArrayIndexes": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: -1e310.", Line: 1}},
		},
		{
			Code:    `foo[-1n]`,
			Options: map[string]interface{}{"ignoreArrayIndexes": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: -1n.", Line: 1}},
		},
		{
			Code:    `foo[-100n]`,
			Options: map[string]interface{}{"ignoreArrayIndexes": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: -100n.", Line: 1}},
		},
		{
			Code:    `foo[-0x12n]`,
			Options: map[string]interface{}{"ignoreArrayIndexes": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: -0x12n.", Line: 1}},
		},
		{
			Code:    `foo[4294967295n]`, // first above the max index
			Options: map[string]interface{}{"ignoreArrayIndexes": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 4294967295n.", Line: 1}},
		},
		{
			Code:    `foo[-(-1)]`, // consistent with the default behavior, which doesn't allow: var foo = -(-1)
			Options: map[string]interface{}{"ignoreArrayIndexes": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: -1.", Line: 1}},
		},
		{
			Code:    `foo[- -1n]`, // consistent with the default behavior, which doesn't allow: var foo = - -1n
			Options: map[string]interface{}{"ignoreArrayIndexes": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: -1n.", Line: 1}},
		},
		{
			Code:    `100 .toString()`,
			Options: map[string]interface{}{"ignoreArrayIndexes": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 100.", Line: 1}},
		},
		{
			Code:    `200[100]`,
			Options: map[string]interface{}{"ignoreArrayIndexes": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 200.", Line: 1}},
		},
		{
			Code:     `var a = <div arrayProp={[1,2,3]}></div>;`,
			FileName: "src/virtual.tsx",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noMagic", Message: "No magic number: 1.", Line: 1},
				{MessageId: "noMagic", Message: "No magic number: 2.", Line: 1},
				{MessageId: "noMagic", Message: "No magic number: 3.", Line: 1},
			},
		},
		{
			Code: `var min, max, mean; min = 1; max = 10; mean = 4;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noMagic", Message: "No magic number: 1.", Line: 1},
				{MessageId: "noMagic", Message: "No magic number: 10.", Line: 1},
				{MessageId: "noMagic", Message: "No magic number: 4.", Line: 1},
			},
		},
		{
			Code:    `f(100n)`,
			Options: map[string]interface{}{"ignore": []interface{}{float64(100)}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 100n.", Line: 1}},
		},
		{
			Code:    `f(-100n)`,
			Options: map[string]interface{}{"ignore": []interface{}{"100n"}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: -100n.", Line: 1}},
		},
		{
			Code:    `f(100n)`,
			Options: map[string]interface{}{"ignore": []interface{}{"-100n"}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 100n.", Line: 1}},
		},
		{
			Code:    `f(100)`,
			Options: map[string]interface{}{"ignore": []interface{}{"100n"}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 100.", Line: 1}},
		},
		{
			Code:    `const func = (param = 123) => {}`,
			Options: map[string]interface{}{"ignoreDefaultValues": false},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 123.", Line: 1}},
		},
		{
			Code:    `const { param = 123 } = sourceObject;`,
			Options: map[string]interface{}{},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 123.", Line: 1}},
		},
		{
			Code:   `const { param = 123 } = sourceObject;`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 123.", Line: 1}},
		},
		{
			Code:    `const { param = 123 } = sourceObject;`,
			Options: map[string]interface{}{"ignoreDefaultValues": false},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 123.", Line: 1}},
		},
		{
			Code:    `const [one = 1, two = 2] = []`,
			Options: map[string]interface{}{"ignoreDefaultValues": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noMagic", Message: "No magic number: 1.", Line: 1},
				{MessageId: "noMagic", Message: "No magic number: 2.", Line: 1},
			},
		},
		{
			Code:    `var one, two; [one = 1, two = 2] = []`,
			Options: map[string]interface{}{"ignoreDefaultValues": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noMagic", Message: "No magic number: 1.", Line: 1},
				{MessageId: "noMagic", Message: "No magic number: 2.", Line: 1},
			},
		},

		// ignoreClassFieldInitialValues
		{
			Code:   `class C { foo = 2; }`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 2.", Line: 1, Column: 17}},
		},
		{
			Code:    `class C { foo = 2; }`,
			Options: map[string]interface{}{},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 2.", Line: 1, Column: 17}},
		},
		{
			Code:    `class C { foo = 2; }`,
			Options: map[string]interface{}{"ignoreClassFieldInitialValues": false},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 2.", Line: 1, Column: 17}},
		},
		{
			Code:    `class C { foo = -2; }`,
			Options: map[string]interface{}{"ignoreClassFieldInitialValues": false},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: -2.", Line: 1, Column: 17}},
		},
		{
			Code:    `class C { static foo = 2; }`,
			Options: map[string]interface{}{"ignoreClassFieldInitialValues": false},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 2.", Line: 1, Column: 24}},
		},
		{
			Code:    `class C { #foo = 2; }`,
			Options: map[string]interface{}{"ignoreClassFieldInitialValues": false},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 2.", Line: 1, Column: 18}},
		},
		{
			Code:    `class C { static #foo = 2; }`,
			Options: map[string]interface{}{"ignoreClassFieldInitialValues": false},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 2.", Line: 1, Column: 25}},
		},
		{
			Code:    `class C { foo = 2 + 3; }`,
			Options: map[string]interface{}{"ignoreClassFieldInitialValues": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noMagic", Message: "No magic number: 2.", Line: 1, Column: 17},
				{MessageId: "noMagic", Message: "No magic number: 3.", Line: 1, Column: 21},
			},
		},
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

		// ---- TypeScript parser suite (upstream ruleTesterTypeScript) ----
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
		{
			Code: `
interface Foo {
  bar: 1;
}
`,
			Options: map[string]interface{}{"ignoreNumericLiteralTypes": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 1.", Line: 3, Column: 8}},
		},
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
				{MessageId: "noMagic", Message: "No magic number: +1.", Line: 6, Column: 9},
			},
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
}
`,
			Options: map[string]interface{}{"ignoreReadonlyClassProperties": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noMagic", Message: "No magic number: 1.", Line: 3, Column: 16},
				{MessageId: "noMagic", Message: "No magic number: 2.", Line: 4, Column: 16},
				{MessageId: "noMagic", Message: "No magic number: 3.", Line: 5, Column: 30},
				{MessageId: "noMagic", Message: "No magic number: 4.", Line: 6, Column: 23},
				{MessageId: "noMagic", Message: "No magic number: -5.", Line: 7, Column: 16},
				{MessageId: "noMagic", Message: "No magic number: +6.", Line: 8, Column: 16},
				{MessageId: "noMagic", Message: "No magic number: 100n.", Line: 9, Column: 24},
			},
		},
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
	})
}
