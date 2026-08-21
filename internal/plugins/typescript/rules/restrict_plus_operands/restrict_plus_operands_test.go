package restrict_plus_operands

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestParseOptions(t *testing.T) {
	tests := []struct {
		name    string
		options []any
		want    RestrictPlusOperandsOptions
	}{
		{
			name: "defaults",
			want: RestrictPlusOperandsOptions{
				AllowAny:             true,
				AllowBoolean:         true,
				AllowNullish:         true,
				AllowNumberAndString: true,
				AllowRegExp:          true,
			},
		},
		{
			name: "serialized partial overrides",
			options: []any{map[string]any{
				"allowAny":                false,
				"allowNullish":            false,
				"skipCompoundAssignments": true,
			}},
			want: RestrictPlusOperandsOptions{
				AllowBoolean:            true,
				AllowNumberAndString:    true,
				AllowRegExp:             true,
				SkipCompoundAssignments: true,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := parseOptions(test.options); got != test.want {
				t.Fatalf("resolved options = %+v, want %+v", got, test.want)
			}
		})
	}
}

func TestRestrictPlusOperandsRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &RestrictPlusOperandsRule, []rule_tester.ValidTestCase{
		{Code: "let x = 5;"},
		{Code: "let y = '10';"},
		{Code: "let z = 8.2;"},
		{Code: "let w = '6.5';"},
		{Code: "let foo = 5 + 10;"},
		{Code: "let foo = '5.5' + '10';"},
		{Code: "let foo = parseInt('5.5', 10) + 10;"},
		{Code: "let foo = parseFloat('5.5', 10) + 10;"},
		{Code: "let foo = 1n + 1n;"},
		{Code: "let foo = BigInt(1) + 1n;"},
		{Code: `
      let foo = 1n;
      foo + 2n;
    `},
		{Code: `
function test(s: string, n: number): number {
  return 2;
}
let foo = test('5.5', 10) + 10;
    `},
		{Code: `
let x = 5;
let z = 8.2;
let foo = x + z;
    `},
		{Code: `
let w = '6.5';
let y = '10';
let foo = y + w;
    `},
		{Code: "let foo = 1 + 1;"},
		{Code: "let foo = '1' + '1';"},
		{Code: `
let pair: { first: number; second: string } = { first: 5, second: '10' };
let foo = pair.first + 10;
    `},
		{Code: `
let pair: { first: number; second: string } = { first: 5, second: '10' };
let foo = pair.first + (10 as number);
    `},
		{Code: `
let pair: { first: number; second: string } = { first: 5, second: '10' };
let foo = '5.5' + pair.second;
    `},
		{Code: `
let pair: { first: number; second: string } = { first: 5, second: '10' };
let foo = ('5.5' as string) + pair.second;
    `},
		{Code: `
      const foo =
        'hello' +
        (someBoolean ? 'a' : 'b') +
        (() => (someBoolean ? 'c' : 'd'))() +
        'e';
    `},
		{Code: "const balls = true;"},
		{Code: "balls === true;"},
		{Code: `
function foo<T extends string>(a: T) {
  return a + '';
}
    `},
		{Code: `
function foo<T extends 'a' | 'b'>(a: T) {
  return a + '';
}
    `},
		{Code: `
function foo<T extends number>(a: T) {
  return a + 1;
}
    `},
		{Code: `
function foo<T extends 1>(a: T) {
  return a + 1;
}
    `},
		{Code: `
declare const a: {} & string;
declare const b: string;
const x = a + b;
    `},
		{Code: `
declare const a: unknown & string;
declare const b: string;
const x = a + b;
    `},
		{Code: `
declare const a: string & string;
declare const b: string;
const x = a + b;
    `},
		{Code: `
declare const a: 'string literal' & string;
declare const b: string;
const x = a + b;
    `},
		{Code: `
declare const a: {} & number;
declare const b: number;
const x = a + b;
    `},
		{Code: `
declare const a: unknown & number;
declare const b: number;
const x = a + b;
    `},
		{Code: `
declare const a: number & number;
declare const b: number;
const x = a + b;
    `},
		{Code: `
declare const a: 42 & number;
declare const b: number;
const x = a + b;
    `},
		{Code: `
declare const a: {} & bigint;
declare const b: bigint;
const x = a + b;
    `},
		{Code: `
declare const a: unknown & bigint;
declare const b: bigint;
const x = a + b;
    `},
		{Code: `
declare const a: bigint & bigint;
declare const b: bigint;
const x = a + b;
    `},
		{Code: `
declare const a: 42n & bigint;
declare const b: bigint;
const x = a + b;
    `},
		{Code: `
function A(s: string) {
  return ` + "`" + `a${s}b` + "`" + ` as const;
}
const b = A('') + '!';
    `},
		{Code: `
declare const a: ` + "`" + `template${string}` + "`" + `;
declare const b: '';
const x = a + b;
    `},
		{Code: `
const a: ` + "`" + `template${0}` + "`" + `;
declare const b: '';
const x = a + b;
    `},
		{
			Code: `
        declare const a: RegExp;
        declare const b: string;
        const x = a + b;
      `,
			Options: map[string]interface{}{
				"allowAny":             false,
				"allowBoolean":         false,
				"allowNullish":         false,
				"allowNumberAndString": false,
				"allowRegExp":          true,
			},
		},
		{
			Code: `
        const a = /regexp/;
        declare const b: string;
        const x = a + b;
      `,
			Options: map[string]interface{}{
				"allowAny":             false,
				"allowBoolean":         false,
				"allowNullish":         false,
				"allowNumberAndString": false,
				"allowRegExp":          true,
			},
		},
		{
			Code: `
const f = (a: RegExp, b: RegExp) => a + b;
      `,
			Options: map[string]interface{}{"allowRegExp": true},
		},
		{
			Code: `
let foo: string | undefined;
foo = foo + 'some data';
      `,
			Options: map[string]interface{}{"allowNullish": true},
		},
		{
			Code: `
let foo: string | null;
foo = foo + 'some data';
      `,
			Options: map[string]interface{}{"allowNullish": true},
		},
		{
			Code: `
let foo: string | null | undefined;
foo = foo + 'some data';
      `,
			Options: map[string]interface{}{"allowNullish": true},
		},
		{
			Code: `
let foo = '';
foo += 0;
      `,
			Options: map[string]interface{}{
				"allowAny":                false,
				"allowBoolean":            false,
				"allowNullish":            false,
				"allowNumberAndString":    false,
				"allowRegExp":             false,
				"skipCompoundAssignments": true,
			},
		},
		{
			Code: `
let foo = 0;
foo += '';
      `,
			Options: map[string]interface{}{
				"allowAny":                false,
				"allowBoolean":            false,
				"allowNullish":            false,
				"allowNumberAndString":    false,
				"allowRegExp":             false,
				"skipCompoundAssignments": true,
			},
		},
		{
			Code: `
const f = (a: any, b: any) => a + b;
      `,
			Options: map[string]interface{}{"allowAny": true},
		},
		{
			Code: `
const f = (a: any, b: string) => a + b;
      `,
			Options: map[string]interface{}{"allowAny": true},
		},
		{
			Code: `
const f = (a: any, b: bigint) => a + b;
      `,
			Options: map[string]interface{}{"allowAny": true},
		},
		{
			Code: `
const f = (a: any, b: number) => a + b;
      `,
			Options: map[string]interface{}{"allowAny": true},
		},
		{
			Code: `
const f = (a: any, b: boolean) => a + b;
      `,
			Options: map[string]interface{}{"allowAny": true, "allowBoolean": true},
		},
		{
			Code: `
const f = (a: string, b: string | number) => a + b;
      `,
			Options: map[string]interface{}{
				"allowAny":             true,
				"allowBoolean":         true,
				"allowNullish":         true,
				"allowNumberAndString": true,
				"allowRegExp":          true,
			},
		},
		{
			Code: `
const f = (a: string | number, b: number) => a + b;
      `,
			Options: map[string]interface{}{
				"allowAny":             true,
				"allowBoolean":         true,
				"allowNullish":         true,
				"allowNumberAndString": true,
				"allowRegExp":          true,
			},
		},
		{
			Code: `
const f = (a: string | number, b: string | number) => a + b;
      `,
			Options: map[string]interface{}{
				"allowAny":             true,
				"allowBoolean":         true,
				"allowNullish":         true,
				"allowNumberAndString": true,
				"allowRegExp":          true,
			},
		},
		{
			Code:    "let foo = '1' + 1n;",
			Options: map[string]interface{}{"allowNumberAndString": true},
		},
	}, []rule_tester.InvalidTestCase{
		{
			Code:    "let foo = '1' + 1;",
			Options: map[string]interface{}{"allowNumberAndString": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mismatched",
					Line:      1,
					Column:    11,
				},
			},
		},
		{
			Code: "let foo = '1' + 1;",
			Options: map[string]interface{}{
				"allowAny":             false,
				"allowBoolean":         false,
				"allowNullish":         false,
				"allowNumberAndString": false,
				"allowRegExp":          false,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mismatched",
					Line:      1,
					Column:    11,
				},
			},
		},
		{
			Code: "let foo = [] + {};",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalid",
					Line:      1,
					Column:    11,
					EndColumn: 13,
				},
				{
					MessageId: "invalid",
					Line:      1,
					Column:    16,
					EndColumn: 18,
				},
			},
		},
		{
			Code: "let foo = 5 + '10';",
			Options: map[string]interface{}{
				"allowAny":             false,
				"allowBoolean":         false,
				"allowNullish":         false,
				"allowNumberAndString": false,
				"allowRegExp":          false,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mismatched",
					Line:      1,
					Column:    11,
				},
			},
		},
		{
			Code: "let foo = [] + 5;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalid",
					Line:      1,
					Column:    11,
					EndColumn: 13,
				},
			},
		},
		{
			Code: "let foo = [] + [];",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalid",
					Line:      1,
					Column:    11,
					EndColumn: 13,
				},
				{
					MessageId: "invalid",
					Line:      1,
					Column:    16,
					EndColumn: 18,
				},
			},
		},
		{
			Code: "let foo = 5 + [3];",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalid",
					Line:      1,
					Column:    15,
					EndColumn: 18,
				},
			},
		},
		{
			Code: "let foo = '5' + {};",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalid",
					Line:      1,
					Column:    17,
					EndColumn: 19,
				},
			},
		},
		{
			Code:    "let foo = 5.5 + '5';",
			Options: map[string]interface{}{"allowNumberAndString": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mismatched",
					Line:      1,
					Column:    11,
				},
			},
		},
		{
			Code:    "let foo = '5.5' + 5;",
			Options: map[string]interface{}{"allowNumberAndString": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mismatched",
					Line:      1,
					Column:    11,
				},
			},
		},
		{
			Code: `
let x = 5;
let y = '10';
let foo = x + y;
      `,
			Options: map[string]interface{}{"allowNumberAndString": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mismatched",
					Line:      4,
					Column:    11,
				},
			},
		},
		{
			Code: `
let x = 5;
let y = '10';
let foo = y + x;
      `,
			Options: map[string]interface{}{"allowNumberAndString": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mismatched",
					Line:      4,
					Column:    11,
				},
			},
		},
		{
			Code: `
let x = 5;
let foo = x + {};
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalid",
					Line:      3,
					Column:    15,
				},
			},
		},
		{
			Code: `
let y = '10';
let foo = [] + y;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalid",
					Line:      3,
					Column:    11,
				},
			},
		},
		{
			Code: `
let pair = { first: 5, second: '10' };
let foo = pair + pair;
      `,
			Options: map[string]interface{}{
				"allowAny":             false,
				"allowBoolean":         false,
				"allowNullish":         false,
				"allowNumberAndString": false,
				"allowRegExp":          false,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalid",
					Line:      3,
					Column:    11,
					EndColumn: 15,
				},
				{
					MessageId: "invalid",
					Line:      3,
					Column:    18,
					EndColumn: 22,
				},
			},
		},
		{
			Code: `
type Valued = { value: number };
let value: Valued = { value: 0 };
let combined = value + 0;
      `,
			Options: map[string]interface{}{
				"allowAny":             false,
				"allowBoolean":         false,
				"allowNullish":         false,
				"allowNumberAndString": false,
				"allowRegExp":          false,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalid",
					Line:      4,
					Column:    16,
					EndColumn: 21,
				},
			},
		},
		{
			Code: "let foo = 1n + 1;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bigintAndNumber",
					Line:      1,
					Column:    11,
				},
			},
		},
		{
			Code: "let foo = 1 + 1n;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bigintAndNumber",
					Line:      1,
					Column:    11,
				},
			},
		},
		{
			Code: `
        let foo = 1n;
        foo + 1;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bigintAndNumber",
					Line:      3,
					Column:    9,
				},
			},
		},
		{
			Code: `
        let foo = 1;
        foo + 1n;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bigintAndNumber",
					Line:      3,
					Column:    9,
				},
			},
		},
		{
			Code: `
function foo<T extends string>(a: T) {
  return a + 1;
}
      `,
			Options: map[string]interface{}{
				"allowAny":             false,
				"allowBoolean":         false,
				"allowNullish":         false,
				"allowNumberAndString": false,
				"allowRegExp":          false,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mismatched",
					Line:      3,
					Column:    10,
				},
			},
		},
		{
			Code: `
function foo<T extends 'a' | 'b'>(a: T) {
  return a + 1;
}
      `,
			Options: map[string]interface{}{
				"allowAny":             false,
				"allowBoolean":         false,
				"allowNullish":         false,
				"allowNumberAndString": false,
				"allowRegExp":          false,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mismatched",
					Line:      3,
					Column:    10,
				},
			},
		},
		{
			Code: `
function foo<T extends number>(a: T) {
  return a + '';
}
      `,
			Options: map[string]interface{}{
				"allowAny":             false,
				"allowBoolean":         false,
				"allowNullish":         false,
				"allowNumberAndString": false,
				"allowRegExp":          false,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mismatched",
					Line:      3,
					Column:    10,
				},
			},
		},
		{
			Code: `
function foo<T extends 1>(a: T) {
  return a + '';
}
      `,
			Options: map[string]interface{}{
				"allowAny":             false,
				"allowBoolean":         false,
				"allowNullish":         false,
				"allowNumberAndString": false,
				"allowRegExp":          false,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mismatched",
					Line:      3,
					Column:    10,
				},
			},
		},
		{
			Code: `
        declare const a: ` + "`" + `template${number}` + "`" + `;
        declare const b: number;
        const x = a + b;
      `,
			Options: map[string]interface{}{
				"allowAny":             false,
				"allowBoolean":         false,
				"allowNullish":         false,
				"allowNumberAndString": false,
				"allowRegExp":          false,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mismatched",
					Line:      4,
					Column:    19,
				},
			},
		},
		{
			Code: `
        declare const a: never;
        declare const b: string;
        const x = a + b;
      `,
			Options: map[string]interface{}{
				"allowAny":             false,
				"allowBoolean":         false,
				"allowNullish":         false,
				"allowNumberAndString": false,
				"allowRegExp":          false,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalid",
					Line:      4,
					Column:    19,
				},
			},
		},
		{
			Code: `
        declare const a: never & string;
        declare const b: string;
        const x = a + b;
      `,
			Options: map[string]interface{}{
				"allowAny":             false,
				"allowBoolean":         false,
				"allowNullish":         false,
				"allowNumberAndString": false,
				"allowRegExp":          false,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalid",
					Line:      4,
					Column:    19,
				},
			},
		},
		{
			Code: `
        declare const a: boolean & string;
        declare const b: string;
        const x = a + b;
      `,
			Options: map[string]interface{}{
				"allowAny":             false,
				"allowBoolean":         false,
				"allowNullish":         false,
				"allowNumberAndString": false,
				"allowRegExp":          false,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalid",
					Line:      4,
					Column:    19,
				},
			},
		},
		{
			Code: `
        declare const a: any & string;
        declare const b: string;
        const x = a + b;
      `,
			Options: map[string]interface{}{
				"allowAny":             false,
				"allowBoolean":         false,
				"allowNullish":         false,
				"allowNumberAndString": false,
				"allowRegExp":          false,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalid",
					Line:      4,
					Column:    19,
				},
			},
		},
		{
			Code: `
        declare const a: { a: 1 } & { b: 2 };
        declare const b: string;
        const x = a + b;
      `,
			Options: map[string]interface{}{
				"allowAny":             false,
				"allowBoolean":         false,
				"allowNullish":         false,
				"allowNumberAndString": false,
				"allowRegExp":          false,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalid",
					Line:      4,
					Column:    19,
				},
			},
		},
		{
			Code: `
        interface A {
          a: 1;
        }
        declare const a: A;
        declare const b: string;
        const x = a + b;
      `,
			Options: map[string]interface{}{
				"allowAny":             false,
				"allowBoolean":         false,
				"allowNullish":         false,
				"allowNumberAndString": false,
				"allowRegExp":          false,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalid",
					Line:      7,
					Column:    19,
				},
			},
		},
		{
			Code: `
        interface A {
          a: 1;
        }
        interface A2 extends A {
          b: 2;
        }
        declare const a: A2;
        declare const b: string;
        const x = a + b;
      `,
			Options: map[string]interface{}{
				"allowAny":             false,
				"allowBoolean":         false,
				"allowNullish":         false,
				"allowNumberAndString": false,
				"allowRegExp":          false,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalid",
					Line:      10,
					Column:    19,
				},
			},
		},
		{
			Code: `
        type A = { a: 1 } & { b: 2 };
        declare const a: A;
        declare const b: string;
        const x = a + b;
      `,
			Options: map[string]interface{}{
				"allowAny":             false,
				"allowBoolean":         false,
				"allowNullish":         false,
				"allowNumberAndString": false,
				"allowRegExp":          false,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalid",
					Line:      5,
					Column:    19,
				},
			},
		},
		{
			Code: `
        declare const a: { a: 1 } & { b: 2 };
        declare const b: number;
        const x = a + b;
      `,
			Options: map[string]interface{}{
				"allowAny":             false,
				"allowBoolean":         false,
				"allowNullish":         false,
				"allowNumberAndString": false,
				"allowRegExp":          false,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalid",
					Line:      4,
					Column:    19,
				},
			},
		},
		{
			Code: `
        declare const a: never;
        declare const b: bigint;
        const x = a + b;
      `,
			Options: map[string]interface{}{
				"allowAny":             false,
				"allowBoolean":         false,
				"allowNullish":         false,
				"allowNumberAndString": false,
				"allowRegExp":          false,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalid",
					Line:      4,
					Column:    19,
				},
			},
		},
		{
			Code: `
        declare const a: any;
        declare const b: bigint;
        const x = a + b;
      `,
			Options: map[string]interface{}{
				"allowAny":             false,
				"allowBoolean":         false,
				"allowNullish":         false,
				"allowNumberAndString": false,
				"allowRegExp":          false,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalid",
					Line:      4,
					Column:    19,
				},
			},
		},
		{
			Code: `
        declare const a: { a: 1 } & { b: 2 };
        declare const b: bigint;
        const x = a + b;
      `,
			Options: map[string]interface{}{
				"allowAny":             false,
				"allowBoolean":         false,
				"allowNullish":         false,
				"allowNumberAndString": false,
				"allowRegExp":          false,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalid",
					Line:      4,
					Column:    19,
				},
			},
		},
		{
			Code: `
        declare const a: RegExp;
        declare const b: string;
        const x = a + b;
      `,
			Options: map[string]interface{}{
				"allowAny":             false,
				"allowBoolean":         false,
				"allowNullish":         false,
				"allowNumberAndString": false,
				"allowRegExp":          false,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalid",
					Line:      4,
					Column:    19,
				},
			},
		},
		{
			Code: `
        const a = /regexp/;
        declare const b: string;
        const x = a + b;
      `,
			Options: map[string]interface{}{
				"allowAny":             false,
				"allowBoolean":         false,
				"allowNullish":         false,
				"allowNumberAndString": false,
				"allowRegExp":          false,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalid",
					Line:      4,
					Column:    19,
				},
			},
		},
		{
			Code: `
        declare const a: Symbol;
        declare const b: string;
        const x = a + b;
      `,
			Options: map[string]interface{}{
				"allowAny":             false,
				"allowBoolean":         false,
				"allowNullish":         false,
				"allowNumberAndString": false,
				"allowRegExp":          false,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalid",
					Line:      4,
					Column:    19,
				},
			},
		},
		{
			Code: `
        declare const a: symbol;
        declare const b: string;
        const x = a + b;
      `,
			Options: map[string]interface{}{
				"allowAny":             false,
				"allowBoolean":         false,
				"allowNullish":         false,
				"allowNumberAndString": false,
				"allowRegExp":          false,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalid",
					Line:      4,
					Column:    19,
				},
			},
		},
		{
			Code: `
        declare const a: unique symbol;
        declare const b: string;
        const x = a + b;
      `,
			Options: map[string]interface{}{
				"allowAny":             false,
				"allowBoolean":         false,
				"allowNullish":         false,
				"allowNumberAndString": false,
				"allowRegExp":          false,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalid",
					Line:      4,
					Column:    19,
				},
			},
		},
		{
			Code: `
        const a = Symbol('');
        declare const b: string;
        const x = a + b;
      `,
			Options: map[string]interface{}{
				"allowAny":             false,
				"allowBoolean":         false,
				"allowNullish":         false,
				"allowNumberAndString": false,
				"allowRegExp":          false,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalid",
					Line:      4,
					Column:    19,
				},
			},
		},
		{
			Code: `
let foo: string | undefined;
foo += 'some data';
      `,
			Options: map[string]interface{}{
				"allowAny":                false,
				"allowBoolean":            false,
				"allowNullish":            false,
				"allowNumberAndString":    false,
				"allowRegExp":             false,
				"skipCompoundAssignments": false,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalid",
					Line:      3,
					Column:    1,
				},
			},
		},
		{
			Code: `
let foo: string | null;
foo += 'some data';
      `,
			Options: map[string]interface{}{
				"allowAny":             false,
				"allowBoolean":         false,
				"allowNullish":         false,
				"allowNumberAndString": false,
				"allowRegExp":          false,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalid",
					Line:      3,
					Column:    1,
				},
			},
		},
		{
			Code: `
let foo: string = '';
foo += 1;
      `,
			Options: map[string]interface{}{
				"allowAny":             false,
				"allowBoolean":         false,
				"allowNullish":         false,
				"allowNumberAndString": false,
				"allowRegExp":          false,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mismatched",
					Line:      3,
					Column:    1,
				},
			},
		},
		{
			Code: `
let foo = 0;
foo += '';
      `,
			Options: map[string]interface{}{
				"allowAny":             false,
				"allowBoolean":         false,
				"allowNullish":         false,
				"allowNumberAndString": false,
				"allowRegExp":          false,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mismatched",
					Line:      3,
					Column:    1,
				},
			},
		},
		{
			Code: `
const f = (a: any, b: boolean) => a + b;
      `,
			Options: map[string]interface{}{"allowAny": true, "allowBoolean": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalid",
					Line:      2,
					Column:    39,
				},
			},
		},
		{
			Code: `
const f = (a: any, b: []) => a + b;
      `,
			Options: map[string]interface{}{"allowAny": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalid",
					Line:      2,
					Column:    34,
				},
			},
		},
		{
			Code: `
const f = (a: any, b: boolean) => a + b;
      `,
			Options: map[string]interface{}{"allowAny": false, "allowBoolean": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalid",
					Line:      2,
					Column:    35,
				},
			},
		},
		{
			Code: `
const f = (a: any, b: any) => a + b;
      `,
			Options: map[string]interface{}{"allowAny": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalid",
					Line:      2,
					Column:    31,
				},
				{
					MessageId: "invalid",
					Line:      2,
					Column:    35,
				},
			},
		},
		{
			Code: `
const f = (a: any, b: string) => a + b;
      `,
			Options: map[string]interface{}{"allowAny": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalid",
					Line:      2,
					Column:    34,
				},
			},
		},
		{
			Code: `
const f = (a: any, b: bigint) => a + b;
      `,
			Options: map[string]interface{}{"allowAny": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalid",
					Line:      2,
					Column:    34,
				},
			},
		},
		{
			Code: `
const f = (a: any, b: number) => a + b;
      `,
			Options: map[string]interface{}{"allowAny": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalid",
					Line:      2,
					Column:    34,
				},
			},
		},
		{
			Code: `
const f = (a: any, b: boolean) => a + b;
      `,
			Options: map[string]interface{}{"allowAny": false, "allowBoolean": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalid",
					Line:      2,
					Column:    35,
				},
				{
					MessageId: "invalid",
					Line:      2,
					Column:    39,
				},
			},
		},
		{
			Code: `
const f = (a: number, b: RegExp) => a + b;
      `,
			Options: map[string]interface{}{"allowRegExp": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalid",
					Line:      2,
					Column:    41,
				},
			},
		},
		{
			Code: `
let foo: string | boolean;
foo = foo + 'some data';
      `,
			Options: map[string]interface{}{"allowBoolean": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalid",
					Line:      3,
					Column:    7,
				},
			},
		},
		{
			Code: `
let foo: boolean;
foo = foo + 'some data';
      `,
			Options: map[string]interface{}{"allowBoolean": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalid",
					Line:      3,
					Column:    7,
				},
			},
		},
		{
			Code: `
const f = (a: any, b: unknown) => a + b;
      `,
			Options: map[string]interface{}{
				"allowAny":     true,
				"allowBoolean": true,
				"allowNullish": true,
				"allowRegExp":  true,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalid",
					Line:      2,
					Column:    39,
				},
			},
		},
		{
			Code:    "let foo = '1' + 1n;",
			Options: map[string]interface{}{"allowNumberAndString": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mismatched",
					Line:      1,
					Column:    11,
				},
			},
		},
	})
}

// TestRestrictPlusOperandsSerializedOptions exercises the strict preset's
// map-shaped options together, including the compound-assignment escape hatch.
func TestRestrictPlusOperandsSerializedOptions(t *testing.T) {
	strictOptions := map[string]interface{}{
		"allowAny":             false,
		"allowBoolean":         false,
		"allowNullish":         false,
		"allowNumberAndString": false,
		"allowRegExp":          false,
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &RestrictPlusOperandsRule,
		[]rule_tester.ValidTestCase{
			{
				Code: "let value = ''; value += 1;",
				Options: map[string]interface{}{
					"allowNumberAndString":    false,
					"skipCompoundAssignments": true,
				},
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:    "let value = 'x' + 1;",
				Options: strictOptions,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "mismatched",
					Line:      1,
					Column:    13,
				}},
			},
			{
				Code:    "declare const value: boolean;\nvalue + '';",
				Options: strictOptions,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "invalid",
					Line:      2,
					Column:    1,
				}},
			},
			{
				Code:    "declare const value: null | undefined;\nvalue + '';",
				Options: strictOptions,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "invalid",
					Line:      2,
					Column:    1,
				}},
			},
			{
				Code:    "declare const value: any;\nvalue + '';",
				Options: strictOptions,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "invalid",
					Line:      2,
					Column:    1,
				}},
			},
			{
				Code:    "declare const value: RegExp;\nvalue + '';",
				Options: strictOptions,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "invalid",
					Line:      2,
					Column:    1,
				}},
			},
		},
	)
}

// TestRestrictPlusOperandsExtras locks in object-operand diagnostics that are
// not message-checked by the migrated suite in restrict_plus_operands_test.go.
func TestRestrictPlusOperandsExtras(t *testing.T) {
	const invalidMessagePrefix = "Invalid operand for a '+' operation. Operands must each be a number or string, allowing a string + any of: `any`, `boolean`, `null`, `RegExp`, `undefined`. Got `"
	strictOptions := map[string]any{
		"allowAny":             false,
		"allowBoolean":         false,
		"allowNullish":         false,
		"allowNumberAndString": false,
		"allowRegExp":          false,
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &RestrictPlusOperandsRule, []rule_tester.ValidTestCase{
		{
			// Locks in upstream getTypeName() behavior for a shadowed RegExp.
			Code: `function convert() {
  class RegExp {}
  const value = new RegExp();
  return '' + value;
}`,
		},
		{
			// Locks in upstream getTypeName() behavior for a qualified RegExp.
			Code: `namespace Custom {
  export class RegExp {}
}
declare const value: Custom.RegExp;
const result = '' + value;`,
		},
		{
			// Type-parameter constraints participate in allowRegExp by type name.
			Code: "function convert<T extends RegExp>(value: T) { return '' + value; }",
		},
		{
			// Reporting through an unwrapped operand must still honor directives.
			Code: "// eslint-disable-next-line test\nconst result = '' + ({});",
		},
	}, []rule_tester.InvalidTestCase{
		{
			// Real-user regression: rspack's ModuleError.ts narrows err to Error.
			Code: `function createMessage(err: Error): string {
  let message = '';

  if (err && typeof err === 'object' && err.message) {
    message += err.message;
  } else if (err) {
    message += err;
  }

  return message;
}`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "invalid",
				Message:   invalidMessagePrefix + "Error`.",
				Line:      7,
				Column:    16,
				EndLine:   7,
				EndColumn: 19,
			}},
		},
		{
			// Locks in upstream's per-union-constituent object reports.
			Code: "declare const value: Date | Error;\nconst result = '' + value;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalid",
					Message:   invalidMessagePrefix + "Date`.",
					Line:      2,
					Column:    21,
					EndLine:   2,
					EndColumn: 26,
				},
				{
					MessageId: "invalid",
					Message:   invalidMessagePrefix + "Error`.",
					Line:      2,
					Column:    21,
					EndLine:   2,
					EndColumn: 26,
				},
			},
		},
		{
			// Parentheses and leading comments are not part of the ESTree operand.
			Code: "declare const value: Error;\nconst result = '' + ((/* keep */ value));",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "invalid",
				Message:   invalidMessagePrefix + "Error`.",
				Line:      2,
				Column:    34,
				EndLine:   2,
				EndColumn: 39,
			}},
		},
		{
			// Parentheses around a TS assertion are omitted, but the assertion stays.
			Code: "declare const value: Error;\nconst result = '' + (value as Error);",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "invalid",
				Message:   invalidMessagePrefix + "Error`.",
				Line:      2,
				Column:    22,
				EndLine:   2,
				EndColumn: 36,
			}},
		},
		{
			// Parentheses around a satisfies expression are omitted from its range.
			Code: "declare const value: Error;\nconst result = '' + (value satisfies Error);",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "invalid",
				Message:   invalidMessagePrefix + "Error`.",
				Line:      2,
				Column:    22,
				EndLine:   2,
				EndColumn: 43,
			}},
		},
		{
			// The same unwrapping applies when the invalid operand is on the left.
			Code: "declare const value: Error;\nconst result = ((value)) + '';",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "invalid",
				Message:   invalidMessagePrefix + "Error`.",
				Line:      2,
				Column:    18,
				EndLine:   2,
				EndColumn: 23,
			}},
		},
		{
			// A RegExp-named local type is rejected when paired with a number.
			Code: `function convert() {
  class RegExp {}
  const value = new RegExp();
  return 1 + value;
}`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "invalid",
				Message:   invalidMessagePrefix + "RegExp`.",
				Line:      4,
				Column:    14,
				EndLine:   4,
				EndColumn: 19,
			}},
		},
		{
			// Disabling allowRegExp rejects shadowed types named RegExp too.
			Code: `function convert() {
  class RegExp {}
  const value = new RegExp();
  return '' + value;
}`,
			Options: strictOptions,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "invalid",
				Message:   "Invalid operand for a '+' operation. Operands must each be a number or string. Got `RegExp`.",
				Line:      4,
				Column:    15,
				EndLine:   4,
				EndColumn: 20,
			}},
		},
		{
			// Disabling allowRegExp also rejects a RegExp-constrained parameter.
			Code:    "function convert<T extends RegExp>(value: T) { return '' + value; }",
			Options: strictOptions,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "invalid",
				Message:   "Invalid operand for a '+' operation. Operands must each be a number or string. Got `RegExp`.",
				Line:      1,
				Column:    60,
				EndLine:   1,
				EndColumn: 65,
			}},
		},
		{
			// The zero-stringLikes branch retains the exact strict message.
			Code:    "declare const value: { value: number };\nconst result = '' + value;",
			Options: strictOptions,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "invalid",
				Message:   "Invalid operand for a '+' operation. Operands must each be a number or string. Got `{ value: number; }`.",
				Line:      2,
				Column:    21,
				EndLine:   2,
				EndColumn: 26,
			}},
		},
		{
			// The one-stringLike branch retains the exact message.
			Code: "declare const value: Error;\nconst result = '' + value;",
			Options: map[string]any{
				"allowAny":     false,
				"allowBoolean": false,
				"allowNullish": false,
				"allowRegExp":  true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "invalid",
				Message:   "Invalid operand for a '+' operation. Operands must each be a number or string, allowing a string + `RegExp`. Got `Error`.",
				Line:      2,
				Column:    21,
				EndLine:   2,
				EndColumn: 26,
			}},
		},
		{
			// Locks in the exact mismatched diagnostic after message construction changes.
			Code:    "const result = 1 + '';",
			Options: strictOptions,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "mismatched",
				Message:   "Operands of '+' operations must be a number or string. Got `number` + `string`.",
				Line:      1,
				Column:    16,
				EndLine:   1,
				EndColumn: 22,
			}},
		},
		{
			// Locks in the exact bigint/number diagnostic after message construction changes.
			Code: "const result = 1 + 1n;",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "bigintAndNumber",
				Message:   "Numeric '+' operations must either be both bigints or both numbers. Got `number` + `bigint`.",
				Line:      1,
				Column:    16,
				EndLine:   1,
				EndColumn: 22,
			}},
		},
		{
			// An invalid primitive union is one base-type complaint, not object reports.
			Code: "declare const value: symbol | Error;\nconst result = '' + value;",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "invalid",
				Line:      2,
				Column:    21,
				EndLine:   2,
				EndColumn: 26,
			}},
		},
		{
			// A valid string constituent does not mask an invalid object constituent.
			Code: "declare const value: Error | string;\nconst result = '' + value;",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "invalid",
				Message:   invalidMessagePrefix + "Error`.",
				Line:      2,
				Column:    21,
				EndLine:   2,
				EndColumn: 26,
			}},
		},
		{
			// Re-enabling the rule after a scoped directive reports normally.
			Code: `/* eslint-disable test */
const suppressed = '' + ({});
/* eslint-enable test */
const reported = '' + ({});`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "invalid",
				Line:      4,
				Column:    24,
				EndLine:   4,
				EndColumn: 26,
			}},
		},
	})
}
