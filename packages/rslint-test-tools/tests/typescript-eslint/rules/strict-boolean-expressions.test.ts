// Ported from typescript-eslint v8.67.0 (101 valid and 117 invalid cases).
// https://github.com/typescript-eslint/typescript-eslint/blob/v8.67.0/packages/eslint-plugin/tests/rules/strict-boolean-expressions.test.ts
// The upstream noStrictNullCheck location (line 0) is mapped to the file start
// (1:1-1:1), because Rslint exposes 1-based source locations.
// Original source, options and suggestion outputs are preserved. Missing message
// text and full locations were verified with that release's rule and parser.
import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

new RuleTester(config).run('strict-boolean-expressions', {
  valid: [
    "true ? 'a' : 'b';",
    `
if (false) {
}
    `,
    'while (true) {}',
    'for (; false;) {}',
    '!true;',
    'false || 123;',
    "true && 'foo';",
    '!(false || true);',
    'true && false ? true : false;',
    '(false && true) || false;',
    '(false && true) || [];',
    '(false && 1) || (true && 2);',
    `
declare const x: boolean;
if (x) {
}
    `,
    '(x: boolean) => !x;',
    '<T extends boolean>(x: T) => (x ? 1 : 0);',
    `
declare const x: never;
if (x) {
}
    `,
    `
if ('') {
}
    `,
    "while ('x') {}",
    "for (; '';) {}",
    "('' && '1') || x;",
    `
declare const x: string;
if (x) {
}
    `,
    '(x: string) => !x;',
    '<T extends string>(x: T) => (x ? 1 : 0);',
    `
if (0) {
}
    `,
    'while (1n) {}',
    'for (; Infinity;) {}',
    '(0 / 0 && 1 + 2) || x;',
    `
declare const x: number;
if (x) {
}
    `,
    '(x: bigint) => !x;',
    '<T extends number>(x: T) => (x ? 1 : 0);',
    `
declare const x: null | object;
if (x) {
}
    `,
    '(x?: { a: any }) => !x;',
    '<T extends {} | null | undefined>(x: T) => (x ? 1 : 0);',
    {
      code: `
declare const x: boolean | null;
if (x) {
}
      `,
      options: [
        {
          allowNullableBoolean: true,
        },
      ],
    },
    {
      code: `
(x?: boolean) => !x;
      `,
      options: [
        {
          allowNullableBoolean: true,
        },
      ],
    },
    {
      code: `
<T extends boolean | null | undefined>(x: T) => (x ? 1 : 0);
      `,
      options: [
        {
          allowNullableBoolean: true,
        },
      ],
    },
    {
      code: `
const a: (undefined | boolean | null)[] = [true, undefined, null];
a.some(x => x);
      `,
      options: [
        {
          allowNullableBoolean: true,
        },
      ],
    },
    {
      code: `
declare const x: string | null;
if (x) {
}
      `,
      options: [
        {
          allowNullableString: true,
        },
      ],
    },
    {
      code: `
(x?: string) => !x;
      `,
      options: [
        {
          allowNullableString: true,
        },
      ],
    },
    {
      code: `
<T extends string | null | undefined>(x: T) => (x ? 1 : 0);
      `,
      options: [
        {
          allowNullableString: true,
        },
      ],
    },
    {
      code: `
declare const x: number | null;
if (x) {
}
      `,
      options: [
        {
          allowNullableNumber: true,
        },
      ],
    },
    {
      code: `
(x?: number) => !x;
      `,
      options: [
        {
          allowNullableNumber: true,
        },
      ],
    },
    {
      code: `
<T extends number | null | undefined>(x: T) => (x ? 1 : 0);
      `,
      options: [
        {
          allowNullableNumber: true,
        },
      ],
    },
    {
      code: `
declare const arrayOfArrays: (null | unknown[])[];
const isAnyNonEmptyArray1 = arrayOfArrays.some(array => array?.length);
      `,
      options: [
        {
          allowNullableNumber: true,
        },
      ],
    },
    {
      code: `
declare const x: any;
if (x) {
}
      `,
      options: [
        {
          allowAny: true,
        },
      ],
    },
    {
      code: `
x => !x;
      `,
      options: [
        {
          allowAny: true,
        },
      ],
    },
    {
      code: `
<T extends any>(x: T) => (x ? 1 : 0);
      `,
      options: [
        {
          allowAny: true,
        },
      ],
    },
    {
      code: `
declare const arrayOfArrays: any[];
const isAnyNonEmptyArray1 = arrayOfArrays.some(array => array);
      `,
      options: [
        {
          allowAny: true,
        },
      ],
    },
    {
      code: `
1 && true && 'x' && {};
      `,
      options: [
        {
          allowNumber: true,
          allowString: true,
        },
      ],
    },
    {
      code: `
let x = 0 || false || '' || null;
      `,
      options: [
        {
          allowNumber: true,
          allowString: true,
        },
      ],
    },
    {
      code: `
if (1 && true && 'x') void 0;
      `,
      options: [
        {
          allowNumber: true,
          allowString: true,
        },
      ],
    },
    {
      code: `
if (0 || false || '') void 0;
      `,
      options: [
        {
          allowNumber: true,
          allowString: true,
        },
      ],
    },
    {
      code: `
1 && true && 'x' ? {} : null;
      `,
      options: [
        {
          allowNumber: true,
          allowString: true,
        },
      ],
    },
    {
      code: `
0 || false || '' ? null : {};
      `,
      options: [
        {
          allowNumber: true,
          allowString: true,
        },
      ],
    },
    {
      code: `
declare const arrayOfArrays: string[];
const isAnyNonEmptyArray1 = arrayOfArrays.some(array => array);
      `,
      options: [
        {
          allowString: true,
        },
      ],
    },
    {
      code: `
declare const arrayOfArrays: number[];
const isAnyNonEmptyArray1 = arrayOfArrays.some(array => array);
      `,
      options: [
        {
          allowNumber: true,
        },
      ],
    },
    {
      code: `
declare const arrayOfArrays: (null | object)[];
const isAnyNonEmptyArray1 = arrayOfArrays.some(array => array);
      `,
      options: [
        {
          allowNullableObject: true,
        },
      ],
    },
    {
      code: `
enum ExampleEnum {
  This = 0,
  That = 1,
}
const rand = Math.random();
let theEnum: ExampleEnum | null = null;
if (rand < 0.3) {
  theEnum = ExampleEnum.This;
}
if (theEnum) {
}
      `,
      options: [
        {
          allowNullableEnum: true,
        },
      ],
    },
    {
      code: `
enum ExampleEnum {
  This = 0,
  That = 1,
}
const rand = Math.random();
let theEnum: ExampleEnum | null = null;
if (rand < 0.3) {
  theEnum = ExampleEnum.This;
}
if (!theEnum) {
}
      `,
      options: [
        {
          allowNullableEnum: true,
        },
      ],
    },
    {
      code: `
enum ExampleEnum {
  This = 1,
  That = 2,
}
const rand = Math.random();
let theEnum: ExampleEnum | null = null;
if (rand < 0.3) {
  theEnum = ExampleEnum.This;
}
if (!theEnum) {
}
      `,
      options: [
        {
          allowNullableEnum: true,
        },
      ],
    },
    {
      code: `
enum ExampleEnum {
  This = 'one',
  That = 'two',
}
const rand = Math.random();
let theEnum: ExampleEnum | null = null;
if (rand < 0.3) {
  theEnum = ExampleEnum.This;
}
if (!theEnum) {
}
      `,
      options: [
        {
          allowNullableEnum: true,
        },
      ],
    },
    {
      code: `
enum ExampleEnum {
  This = 0,
  That = 'one',
}
(value?: ExampleEnum) => (value ? 1 : 0);
      `,
      options: [
        {
          allowNullableEnum: true,
        },
      ],
    },
    {
      code: `
enum ExampleEnum {
  This = '',
  That = 1,
}
(value?: ExampleEnum) => (!value ? 1 : 0);
      `,
      options: [
        {
          allowNullableEnum: true,
        },
      ],
    },
    {
      code: `
enum ExampleEnum {
  This = 'this',
  That = 1,
}
(value?: ExampleEnum) => (!value ? 1 : 0);
      `,
      options: [
        {
          allowNullableEnum: true,
        },
      ],
    },
    {
      code: `
enum ExampleEnum {
  This = '',
  That = 0,
}
(value?: ExampleEnum) => (!value ? 1 : 0);
      `,
      options: [
        {
          allowNullableEnum: true,
        },
      ],
    },
    {
      code: `
enum ExampleEnum {
  This = '',
  That = 0,
}
declare const arrayOfArrays: (ExampleEnum | null)[];
const isAnyNonEmptyArray1 = arrayOfArrays.some(array => array);
      `,
      options: [
        {
          allowNullableEnum: true,
        },
      ],
    },
    {
      code: `
declare const x: string[] | null;
// eslint-disable-next-line
if (x) {
}
      `,
      languageOptions: {
        parserOptions: {
          project: ['unstrict/tsconfig.json'],
        },
      },
      options: [
        {
          allowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing: true,
        },
      ],
      filename: 'unstrict/file.ts',
    },
    `
function f(arg: 'a' | null) {
  if (arg) console.log(arg);
}
    `,
    `
function f(arg: 'a' | 'b' | null) {
  if (arg) console.log(arg);
}
    `,
    {
      code: `
declare const x: 1 | null;
declare const y: 1;
if (x) {
}
if (y) {
}
      `,
      options: [
        {
          allowNumber: true,
        },
      ],
    },
    `
function f(arg: 1 | null) {
  if (arg) console.log(arg);
}
    `,
    `
function f(arg: 1 | 2 | null) {
  if (arg) console.log(arg);
}
    `,
    `
interface Options {
  readonly enableSomething?: true;
}

function f(opts: Options): void {
  if (opts.enableSomething) console.log('Do something');
}
    `,
    `
declare const x: true | null;
if (x) {
}
    `,
    {
      code: `
declare const x: 'a' | null;
declare const y: 'a';
if (x) {
}
if (y) {
}
      `,
      options: [
        {
          allowString: true,
        },
      ],
    },
    `
declare const foo: boolean & { __BRAND: 'Foo' };
if (foo) {
}
    `,
    `
declare const foo: true & { __BRAND: 'Foo' };
if (foo) {
}
    `,
    `
declare const foo: false & { __BRAND: 'Foo' };
if (foo) {
}
    `,
    `
declare function assert(a: number, b: unknown): asserts a;
declare const nullableString: string | null;
declare const boo: boolean;
assert(boo, nullableString);
    `,
    `
declare function assert(a: boolean, b: unknown): asserts b is string;
declare const nullableString: string | null;
declare const boo: boolean;
assert(boo, nullableString);
    `,
    `
declare function assert(a: number, b: unknown): asserts b;
declare const nullableString: string | null;
declare const boo: boolean;
assert(nullableString, boo);
    `,
    `
declare function assert(a: number, b: unknown): asserts b;
declare const nullableString: string | null;
declare const boo: boolean;
assert(...nullableString, nullableString);
    `,
    `
declare function assert(
  this: object,
  a: number,
  b?: unknown,
  c?: unknown,
): asserts c;
declare const nullableString: string | null;
declare const foo: number;
const o: { assert: typeof assert } = {
  assert,
};
o.assert(foo, nullableString);
    `,
    {
      code: `
declare function assert(x: unknown): x is string;
declare const nullableString: string | null;
assert(nullableString);
      `,
    },
    {
      code: `
class ThisAsserter {
  assertThis(this: unknown, arg2: unknown): asserts this {}
}

declare const lol: string | number | unknown | null;

const thisAsserter: ThisAsserter = new ThisAsserter();
thisAsserter.assertThis(lol);
      `,
    },
    {
      code: `
function assert(this: object, a: number, b: unknown): asserts b;
function assert(a: bigint, b: unknown): asserts b;
function assert(this: object, a: string, two: string): asserts two;
function assert(
  this: object,
  a: string,
  assertee: string,
  c: bigint,
  d: object,
): asserts assertee;
function assert(...args: any[]): void;

function assert(...args: any[]) {
  throw new Error('lol');
}

declare const nullableString: string | null;
assert(3 as any, nullableString);
      `,
    },
    `
declare const assert: any;
declare const nullableString: string | null;
assert(nullableString);
    `,
    `
for (let x = 0; ; x++) {
  break;
}
    `,
    `
[true, false].some(function (x) {
  return x;
});
    `,
    `
[true, false].some(function check(x) {
  return x;
});
    `,
    `
[true, false].some(x => {
  return x;
});
    `,
    `
[1, null].filter(function (x) {
  return x != null;
});
    `,
    `
['one', 'two', ''].filter(function (x) {
  return !!x;
});
    `,
    `
['one', 'two', ''].filter(function (x): boolean {
  return !!x;
});
    `,
    `
['one', 'two', ''].filter(function (x): boolean {
  if (x) {
    return true;
  }
});
    `,
    `
['one', 'two', ''].filter(function (x): boolean {
  if (x) {
    return true;
  }

  throw new Error('oops');
});
    `,
    `
declare const predicate: (string) => boolean;
['one', 'two', ''].filter(predicate);
    `,
    `
declare function notNullish<T>(x: T): x is NonNullable<T>;
['one', null].filter(notNullish);
    `,
    `
declare function predicate(x: string | null): x is string;
['one', null].filter(predicate);
    `,
    `
declare function predicate<T extends boolean>(x: string | null): T;
['one', null].filter(predicate);
    `,
    `
declare function f(x: number): boolean;
declare function f(x: string | null): boolean;

[35].filter(f);
    `,
  ],
  invalid: [
    {
      code: `
if (true && 1 + 1) {
}
      `,
      errors: [
        {
          column: 13,
          line: 2,
          messageId: 'conditionErrorNumber',
          suggestions: [
            {
              messageId: 'conditionFixCompareZero',
              output: `
if (true && ((1 + 1) !== 0)) {
}
      `,
              desc: 'Change condition to check for 0 (`value !== 0`)',
            },
            {
              messageId: 'conditionFixCompareNaN',
              output: `
if (true && (!Number.isNaN((1 + 1)))) {
}
      `,
              desc: 'Change condition to check for NaN (`!Number.isNaN(value)`)',
            },
            {
              messageId: 'conditionFixCastBoolean',
              output: `
if (true && (Boolean((1 + 1)))) {
}
      `,
              desc: 'Explicitly convert value to a boolean (`Boolean(value)`)',
            },
          ],
          message:
            'Unexpected number value in conditional. An explicit zero/NaN check is required.',
          endLine: 2,
          endColumn: 18,
        },
      ],
      options: [
        {
          allowNullableObject: false,
          allowNumber: false,
          allowString: false,
        },
      ],
    },
    {
      code: "while (false || 'a' + 'b') {}",
      errors: [
        {
          column: 17,
          line: 1,
          messageId: 'conditionErrorString',
          suggestions: [
            {
              messageId: 'conditionFixCompareStringLength',
              output: "while (false || (('a' + 'b').length > 0)) {}",
              desc: "Change condition to check string's length (`value.length !== 0`)",
            },
            {
              messageId: 'conditionFixCompareEmptyString',
              output: "while (false || (('a' + 'b') !== \"\")) {}",
              desc: 'Change condition to check for empty string (`value !== ""`)',
            },
            {
              messageId: 'conditionFixCastBoolean',
              output: "while (false || (Boolean(('a' + 'b')))) {}",
              desc: 'Explicitly convert value to a boolean (`Boolean(value)`)',
            },
          ],
          message:
            'Unexpected string value in conditional. An explicit empty string check is required.',
          endLine: 1,
          endColumn: 26,
        },
      ],
      options: [
        {
          allowNullableObject: false,
          allowNumber: false,
          allowString: false,
        },
      ],
    },
    {
      code: '(x: object) => (true || false || x ? true : false);',
      errors: [
        {
          column: 34,
          line: 1,
          messageId: 'conditionErrorObject',
          message:
            'Unexpected object value in conditional. The condition is always true.',
          endLine: 1,
          endColumn: 35,
        },
      ],
      options: [
        {
          allowNullableObject: false,
          allowNumber: false,
          allowString: false,
        },
      ],
    },
    {
      code: "if (('' && {}) || (0 && void 0)) { }",
      errors: [
        {
          column: 6,
          line: 1,
          messageId: 'conditionErrorString',
          suggestions: [
            {
              messageId: 'conditionFixCompareStringLength',
              output: "if (((''.length > 0) && {}) || (0 && void 0)) { }",
              desc: "Change condition to check string's length (`value.length !== 0`)",
            },
            {
              messageId: 'conditionFixCompareEmptyString',
              output: 'if (((\'\' !== "") && {}) || (0 && void 0)) { }',
              desc: 'Change condition to check for empty string (`value !== ""`)',
            },
            {
              messageId: 'conditionFixCastBoolean',
              output: "if (((Boolean('')) && {}) || (0 && void 0)) { }",
              desc: 'Explicitly convert value to a boolean (`Boolean(value)`)',
            },
          ],
          message:
            'Unexpected string value in conditional. An explicit empty string check is required.',
          endLine: 1,
          endColumn: 8,
        },
        {
          column: 12,
          line: 1,
          messageId: 'conditionErrorObject',
          message:
            'Unexpected object value in conditional. The condition is always true.',
          endLine: 1,
          endColumn: 14,
        },
        {
          column: 20,
          line: 1,
          messageId: 'conditionErrorNumber',
          suggestions: [
            {
              messageId: 'conditionFixCompareZero',
              output: "if (('' && {}) || ((0 !== 0) && void 0)) { }",
              desc: 'Change condition to check for 0 (`value !== 0`)',
            },
            {
              messageId: 'conditionFixCompareNaN',
              output: "if (('' && {}) || ((!Number.isNaN(0)) && void 0)) { }",
              desc: 'Change condition to check for NaN (`!Number.isNaN(value)`)',
            },
            {
              messageId: 'conditionFixCastBoolean',
              output: "if (('' && {}) || ((Boolean(0)) && void 0)) { }",
              desc: 'Explicitly convert value to a boolean (`Boolean(value)`)',
            },
          ],
          message:
            'Unexpected number value in conditional. An explicit zero/NaN check is required.',
          endLine: 1,
          endColumn: 21,
        },
        {
          column: 25,
          line: 1,
          messageId: 'conditionErrorNullish',
          message:
            'Unexpected nullish value in conditional. The condition is always false.',
          endLine: 1,
          endColumn: 31,
        },
      ],
      options: [
        {
          allowNullableObject: false,
          allowNumber: false,
          allowString: false,
        },
      ],
    },
    {
      code: `
declare const array: string[];
array.some(x => x);
      `,
      errors: [
        {
          messageId: 'conditionErrorString',
          suggestions: [
            {
              messageId: 'conditionFixCompareStringLength',
              output: `
declare const array: string[];
array.some(x => x.length > 0);
      `,
              desc: "Change condition to check string's length (`value.length !== 0`)",
            },
            {
              messageId: 'conditionFixCompareEmptyString',
              output: `
declare const array: string[];
array.some(x => x !== "");
      `,
              desc: 'Change condition to check for empty string (`value !== ""`)',
            },
            {
              messageId: 'conditionFixCastBoolean',
              output: `
declare const array: string[];
array.some(x => Boolean(x));
      `,
              desc: 'Explicitly convert value to a boolean (`Boolean(value)`)',
            },
            {
              messageId: 'explicitBooleanReturnType',
              output: `
declare const array: string[];
array.some((x): boolean => x);
      `,
              desc: 'Add an explicit `boolean` return type annotation.',
            },
          ],
          message:
            'Unexpected string value in array predicate return type. An explicit empty string check is required.',
          line: 3,
          column: 12,
          endLine: 3,
          endColumn: 18,
        },
      ],
      options: [
        {
          allowNullableBoolean: true,
          allowString: false,
        },
      ],
    },
    {
      code: `
declare const foo: true & { __BRAND: 'Foo' };
if (('' && foo) || (0 && void 0)) { }
      `,
      errors: [
        {
          column: 6,
          line: 3,
          messageId: 'conditionErrorString',
          suggestions: [
            {
              messageId: 'conditionFixCompareStringLength',
              output: `
declare const foo: true & { __BRAND: 'Foo' };
if (((''.length > 0) && foo) || (0 && void 0)) { }
      `,
              desc: "Change condition to check string's length (`value.length !== 0`)",
            },
            {
              messageId: 'conditionFixCompareEmptyString',
              output: `
declare const foo: true & { __BRAND: 'Foo' };
if ((('' !== "") && foo) || (0 && void 0)) { }
      `,
              desc: 'Change condition to check for empty string (`value !== ""`)',
            },
            {
              messageId: 'conditionFixCastBoolean',
              output: `
declare const foo: true & { __BRAND: 'Foo' };
if (((Boolean('')) && foo) || (0 && void 0)) { }
      `,
              desc: 'Explicitly convert value to a boolean (`Boolean(value)`)',
            },
          ],
          message:
            'Unexpected string value in conditional. An explicit empty string check is required.',
          endLine: 3,
          endColumn: 8,
        },
        {
          column: 21,
          line: 3,
          messageId: 'conditionErrorNumber',
          suggestions: [
            {
              messageId: 'conditionFixCompareZero',
              output: `
declare const foo: true & { __BRAND: 'Foo' };
if (('' && foo) || ((0 !== 0) && void 0)) { }
      `,
              desc: 'Change condition to check for 0 (`value !== 0`)',
            },
            {
              messageId: 'conditionFixCompareNaN',
              output: `
declare const foo: true & { __BRAND: 'Foo' };
if (('' && foo) || ((!Number.isNaN(0)) && void 0)) { }
      `,
              desc: 'Change condition to check for NaN (`!Number.isNaN(value)`)',
            },
            {
              messageId: 'conditionFixCastBoolean',
              output: `
declare const foo: true & { __BRAND: 'Foo' };
if (('' && foo) || ((Boolean(0)) && void 0)) { }
      `,
              desc: 'Explicitly convert value to a boolean (`Boolean(value)`)',
            },
          ],
          message:
            'Unexpected number value in conditional. An explicit zero/NaN check is required.',
          endLine: 3,
          endColumn: 22,
        },
        {
          column: 26,
          line: 3,
          messageId: 'conditionErrorNullish',
          message:
            'Unexpected nullish value in conditional. The condition is always false.',
          endLine: 3,
          endColumn: 32,
        },
      ],
      options: [
        {
          allowNullableObject: false,
          allowNumber: false,
          allowString: false,
        },
      ],
    },
    {
      code: `
declare const foo: false & { __BRAND: 'Foo' };
if (('' && {}) || (foo && void 0)) { }
      `,
      errors: [
        {
          column: 6,
          line: 3,
          messageId: 'conditionErrorString',
          suggestions: [
            {
              messageId: 'conditionFixCompareStringLength',
              output: `
declare const foo: false & { __BRAND: 'Foo' };
if (((''.length > 0) && {}) || (foo && void 0)) { }
      `,
              desc: "Change condition to check string's length (`value.length !== 0`)",
            },
            {
              messageId: 'conditionFixCompareEmptyString',
              output: `
declare const foo: false & { __BRAND: 'Foo' };
if ((('' !== "") && {}) || (foo && void 0)) { }
      `,
              desc: 'Change condition to check for empty string (`value !== ""`)',
            },
            {
              messageId: 'conditionFixCastBoolean',
              output: `
declare const foo: false & { __BRAND: 'Foo' };
if (((Boolean('')) && {}) || (foo && void 0)) { }
      `,
              desc: 'Explicitly convert value to a boolean (`Boolean(value)`)',
            },
          ],
          message:
            'Unexpected string value in conditional. An explicit empty string check is required.',
          endLine: 3,
          endColumn: 8,
        },
        {
          column: 12,
          line: 3,
          messageId: 'conditionErrorObject',
          message:
            'Unexpected object value in conditional. The condition is always true.',
          endLine: 3,
          endColumn: 14,
        },
        {
          column: 27,
          line: 3,
          messageId: 'conditionErrorNullish',
          message:
            'Unexpected nullish value in conditional. The condition is always false.',
          endLine: 3,
          endColumn: 33,
        },
      ],
      options: [
        {
          allowNullableObject: false,
          allowNumber: false,
          allowString: false,
        },
      ],
    },
    {
      code: "'asd' && 123 && [] && null;",
      errors: [
        {
          column: 1,
          line: 1,
          messageId: 'conditionErrorString',
          suggestions: [
            {
              messageId: 'conditionFixCompareStringLength',
              output: "('asd'.length > 0) && 123 && [] && null;",
              desc: "Change condition to check string's length (`value.length !== 0`)",
            },
            {
              messageId: 'conditionFixCompareEmptyString',
              output: '(\'asd\' !== "") && 123 && [] && null;',
              desc: 'Change condition to check for empty string (`value !== ""`)',
            },
            {
              messageId: 'conditionFixCastBoolean',
              output: "(Boolean('asd')) && 123 && [] && null;",
              desc: 'Explicitly convert value to a boolean (`Boolean(value)`)',
            },
          ],
          message:
            'Unexpected string value in conditional. An explicit empty string check is required.',
          endLine: 1,
          endColumn: 6,
        },
        {
          column: 10,
          line: 1,
          messageId: 'conditionErrorNumber',
          suggestions: [
            {
              messageId: 'conditionFixCompareZero',
              output: "'asd' && (123 !== 0) && [] && null;",
              desc: 'Change condition to check for 0 (`value !== 0`)',
            },
            {
              messageId: 'conditionFixCompareNaN',
              output: "'asd' && (!Number.isNaN(123)) && [] && null;",
              desc: 'Change condition to check for NaN (`!Number.isNaN(value)`)',
            },
            {
              messageId: 'conditionFixCastBoolean',
              output: "'asd' && (Boolean(123)) && [] && null;",
              desc: 'Explicitly convert value to a boolean (`Boolean(value)`)',
            },
          ],
          message:
            'Unexpected number value in conditional. An explicit zero/NaN check is required.',
          endLine: 1,
          endColumn: 13,
        },
        {
          column: 17,
          line: 1,
          messageId: 'conditionErrorObject',
          message:
            'Unexpected object value in conditional. The condition is always true.',
          endLine: 1,
          endColumn: 19,
        },
      ],
      options: [
        {
          allowNumber: false,
          allowString: false,
        },
      ],
    },
    {
      code: "'asd' || 123 || [] || null;",
      errors: [
        {
          column: 1,
          line: 1,
          messageId: 'conditionErrorString',
          suggestions: [
            {
              messageId: 'conditionFixCompareStringLength',
              output: "('asd'.length > 0) || 123 || [] || null;",
              desc: "Change condition to check string's length (`value.length !== 0`)",
            },
            {
              messageId: 'conditionFixCompareEmptyString',
              output: '(\'asd\' !== "") || 123 || [] || null;',
              desc: 'Change condition to check for empty string (`value !== ""`)',
            },
            {
              messageId: 'conditionFixCastBoolean',
              output: "(Boolean('asd')) || 123 || [] || null;",
              desc: 'Explicitly convert value to a boolean (`Boolean(value)`)',
            },
          ],
          message:
            'Unexpected string value in conditional. An explicit empty string check is required.',
          endLine: 1,
          endColumn: 6,
        },
        {
          column: 10,
          line: 1,
          messageId: 'conditionErrorNumber',
          suggestions: [
            {
              messageId: 'conditionFixCompareZero',
              output: "'asd' || (123 !== 0) || [] || null;",
              desc: 'Change condition to check for 0 (`value !== 0`)',
            },
            {
              messageId: 'conditionFixCompareNaN',
              output: "'asd' || (!Number.isNaN(123)) || [] || null;",
              desc: 'Change condition to check for NaN (`!Number.isNaN(value)`)',
            },
            {
              messageId: 'conditionFixCastBoolean',
              output: "'asd' || (Boolean(123)) || [] || null;",
              desc: 'Explicitly convert value to a boolean (`Boolean(value)`)',
            },
          ],
          message:
            'Unexpected number value in conditional. An explicit zero/NaN check is required.',
          endLine: 1,
          endColumn: 13,
        },
        {
          column: 17,
          line: 1,
          messageId: 'conditionErrorObject',
          message:
            'Unexpected object value in conditional. The condition is always true.',
          endLine: 1,
          endColumn: 19,
        },
      ],
      options: [
        {
          allowNumber: false,
          allowString: false,
        },
      ],
    },
    {
      code: "let x = (1 && 'a' && null) || 0 || '' || {};",
      errors: [
        {
          column: 10,
          line: 1,
          messageId: 'conditionErrorNumber',
          suggestions: [
            {
              messageId: 'conditionFixCompareZero',
              output: "let x = ((1 !== 0) && 'a' && null) || 0 || '' || {};",
              desc: 'Change condition to check for 0 (`value !== 0`)',
            },
            {
              messageId: 'conditionFixCompareNaN',
              output:
                "let x = ((!Number.isNaN(1)) && 'a' && null) || 0 || '' || {};",
              desc: 'Change condition to check for NaN (`!Number.isNaN(value)`)',
            },
            {
              messageId: 'conditionFixCastBoolean',
              output: "let x = ((Boolean(1)) && 'a' && null) || 0 || '' || {};",
              desc: 'Explicitly convert value to a boolean (`Boolean(value)`)',
            },
          ],
          message:
            'Unexpected number value in conditional. An explicit zero/NaN check is required.',
          endLine: 1,
          endColumn: 11,
        },
        {
          column: 15,
          line: 1,
          messageId: 'conditionErrorString',
          suggestions: [
            {
              messageId: 'conditionFixCompareStringLength',
              output:
                "let x = (1 && ('a'.length > 0) && null) || 0 || '' || {};",
              desc: "Change condition to check string's length (`value.length !== 0`)",
            },
            {
              messageId: 'conditionFixCompareEmptyString',
              output: "let x = (1 && ('a' !== \"\") && null) || 0 || '' || {};",
              desc: 'Change condition to check for empty string (`value !== ""`)',
            },
            {
              messageId: 'conditionFixCastBoolean',
              output: "let x = (1 && (Boolean('a')) && null) || 0 || '' || {};",
              desc: 'Explicitly convert value to a boolean (`Boolean(value)`)',
            },
          ],
          message:
            'Unexpected string value in conditional. An explicit empty string check is required.',
          endLine: 1,
          endColumn: 18,
        },
        {
          column: 22,
          line: 1,
          messageId: 'conditionErrorNullish',
          message:
            'Unexpected nullish value in conditional. The condition is always false.',
          endLine: 1,
          endColumn: 26,
        },
        {
          column: 31,
          line: 1,
          messageId: 'conditionErrorNumber',
          suggestions: [
            {
              messageId: 'conditionFixCompareZero',
              output: "let x = (1 && 'a' && null) || (0 !== 0) || '' || {};",
              desc: 'Change condition to check for 0 (`value !== 0`)',
            },
            {
              messageId: 'conditionFixCompareNaN',
              output:
                "let x = (1 && 'a' && null) || (!Number.isNaN(0)) || '' || {};",
              desc: 'Change condition to check for NaN (`!Number.isNaN(value)`)',
            },
            {
              messageId: 'conditionFixCastBoolean',
              output: "let x = (1 && 'a' && null) || (Boolean(0)) || '' || {};",
              desc: 'Explicitly convert value to a boolean (`Boolean(value)`)',
            },
          ],
          message:
            'Unexpected number value in conditional. An explicit zero/NaN check is required.',
          endLine: 1,
          endColumn: 32,
        },
        {
          column: 36,
          line: 1,
          messageId: 'conditionErrorString',
          suggestions: [
            {
              messageId: 'conditionFixCompareStringLength',
              output:
                "let x = (1 && 'a' && null) || 0 || (''.length > 0) || {};",
              desc: "Change condition to check string's length (`value.length !== 0`)",
            },
            {
              messageId: 'conditionFixCompareEmptyString',
              output: "let x = (1 && 'a' && null) || 0 || ('' !== \"\") || {};",
              desc: 'Change condition to check for empty string (`value !== ""`)',
            },
            {
              messageId: 'conditionFixCastBoolean',
              output: "let x = (1 && 'a' && null) || 0 || (Boolean('')) || {};",
              desc: 'Explicitly convert value to a boolean (`Boolean(value)`)',
            },
          ],
          message:
            'Unexpected string value in conditional. An explicit empty string check is required.',
          endLine: 1,
          endColumn: 38,
        },
      ],
      options: [
        {
          allowNumber: false,
          allowString: false,
        },
      ],
    },
    {
      code: "return (1 || 'a' || null) && 0 && '' && {};",
      errors: [
        {
          column: 9,
          line: 1,
          messageId: 'conditionErrorNumber',
          suggestions: [
            {
              messageId: 'conditionFixCompareZero',
              output: "return ((1 !== 0) || 'a' || null) && 0 && '' && {};",
              desc: 'Change condition to check for 0 (`value !== 0`)',
            },
            {
              messageId: 'conditionFixCompareNaN',
              output:
                "return ((!Number.isNaN(1)) || 'a' || null) && 0 && '' && {};",
              desc: 'Change condition to check for NaN (`!Number.isNaN(value)`)',
            },
            {
              messageId: 'conditionFixCastBoolean',
              output: "return ((Boolean(1)) || 'a' || null) && 0 && '' && {};",
              desc: 'Explicitly convert value to a boolean (`Boolean(value)`)',
            },
          ],
          message:
            'Unexpected number value in conditional. An explicit zero/NaN check is required.',
          endLine: 1,
          endColumn: 10,
        },
        {
          column: 14,
          line: 1,
          messageId: 'conditionErrorString',
          suggestions: [
            {
              messageId: 'conditionFixCompareStringLength',
              output:
                "return (1 || ('a'.length > 0) || null) && 0 && '' && {};",
              desc: "Change condition to check string's length (`value.length !== 0`)",
            },
            {
              messageId: 'conditionFixCompareEmptyString',
              output: "return (1 || ('a' !== \"\") || null) && 0 && '' && {};",
              desc: 'Change condition to check for empty string (`value !== ""`)',
            },
            {
              messageId: 'conditionFixCastBoolean',
              output: "return (1 || (Boolean('a')) || null) && 0 && '' && {};",
              desc: 'Explicitly convert value to a boolean (`Boolean(value)`)',
            },
          ],
          message:
            'Unexpected string value in conditional. An explicit empty string check is required.',
          endLine: 1,
          endColumn: 17,
        },
        {
          column: 21,
          line: 1,
          messageId: 'conditionErrorNullish',
          message:
            'Unexpected nullish value in conditional. The condition is always false.',
          endLine: 1,
          endColumn: 25,
        },
        {
          column: 30,
          line: 1,
          messageId: 'conditionErrorNumber',
          suggestions: [
            {
              messageId: 'conditionFixCompareZero',
              output: "return (1 || 'a' || null) && (0 !== 0) && '' && {};",
              desc: 'Change condition to check for 0 (`value !== 0`)',
            },
            {
              messageId: 'conditionFixCompareNaN',
              output:
                "return (1 || 'a' || null) && (!Number.isNaN(0)) && '' && {};",
              desc: 'Change condition to check for NaN (`!Number.isNaN(value)`)',
            },
            {
              messageId: 'conditionFixCastBoolean',
              output: "return (1 || 'a' || null) && (Boolean(0)) && '' && {};",
              desc: 'Explicitly convert value to a boolean (`Boolean(value)`)',
            },
          ],
          message:
            'Unexpected number value in conditional. An explicit zero/NaN check is required.',
          endLine: 1,
          endColumn: 31,
        },
        {
          column: 35,
          line: 1,
          messageId: 'conditionErrorString',
          suggestions: [
            {
              messageId: 'conditionFixCompareStringLength',
              output:
                "return (1 || 'a' || null) && 0 && (''.length > 0) && {};",
              desc: "Change condition to check string's length (`value.length !== 0`)",
            },
            {
              messageId: 'conditionFixCompareEmptyString',
              output: "return (1 || 'a' || null) && 0 && ('' !== \"\") && {};",
              desc: 'Change condition to check for empty string (`value !== ""`)',
            },
            {
              messageId: 'conditionFixCastBoolean',
              output: "return (1 || 'a' || null) && 0 && (Boolean('')) && {};",
              desc: 'Explicitly convert value to a boolean (`Boolean(value)`)',
            },
          ],
          message:
            'Unexpected string value in conditional. An explicit empty string check is required.',
          endLine: 1,
          endColumn: 37,
        },
      ],
      options: [
        {
          allowNumber: false,
          allowString: false,
        },
      ],
    },
    {
      code: "console.log((1 && []) || ('a' && {}));",
      errors: [
        {
          column: 14,
          line: 1,
          messageId: 'conditionErrorNumber',
          suggestions: [
            {
              messageId: 'conditionFixCompareZero',
              output: "console.log(((1 !== 0) && []) || ('a' && {}));",
              desc: 'Change condition to check for 0 (`value !== 0`)',
            },
            {
              messageId: 'conditionFixCompareNaN',
              output: "console.log(((!Number.isNaN(1)) && []) || ('a' && {}));",
              desc: 'Change condition to check for NaN (`!Number.isNaN(value)`)',
            },
            {
              messageId: 'conditionFixCastBoolean',
              output: "console.log(((Boolean(1)) && []) || ('a' && {}));",
              desc: 'Explicitly convert value to a boolean (`Boolean(value)`)',
            },
          ],
          message:
            'Unexpected number value in conditional. An explicit zero/NaN check is required.',
          endLine: 1,
          endColumn: 15,
        },
        {
          column: 19,
          line: 1,
          messageId: 'conditionErrorObject',
          message:
            'Unexpected object value in conditional. The condition is always true.',
          endLine: 1,
          endColumn: 21,
        },
        {
          column: 27,
          line: 1,
          messageId: 'conditionErrorString',
          suggestions: [
            {
              messageId: 'conditionFixCompareStringLength',
              output: "console.log((1 && []) || (('a'.length > 0) && {}));",
              desc: "Change condition to check string's length (`value.length !== 0`)",
            },
            {
              messageId: 'conditionFixCompareEmptyString',
              output: 'console.log((1 && []) || ((\'a\' !== "") && {}));',
              desc: 'Change condition to check for empty string (`value !== ""`)',
            },
            {
              messageId: 'conditionFixCastBoolean',
              output: "console.log((1 && []) || ((Boolean('a')) && {}));",
              desc: 'Explicitly convert value to a boolean (`Boolean(value)`)',
            },
          ],
          message:
            'Unexpected string value in conditional. An explicit empty string check is required.',
          endLine: 1,
          endColumn: 30,
        },
      ],
      options: [
        {
          allowNumber: false,
          allowString: false,
        },
      ],
    },
    {
      code: "if ((1 && []) || ('a' && {})) void 0;",
      errors: [
        {
          column: 6,
          line: 1,
          messageId: 'conditionErrorNumber',
          suggestions: [
            {
              messageId: 'conditionFixCompareZero',
              output: "if (((1 !== 0) && []) || ('a' && {})) void 0;",
              desc: 'Change condition to check for 0 (`value !== 0`)',
            },
            {
              messageId: 'conditionFixCompareNaN',
              output: "if (((!Number.isNaN(1)) && []) || ('a' && {})) void 0;",
              desc: 'Change condition to check for NaN (`!Number.isNaN(value)`)',
            },
            {
              messageId: 'conditionFixCastBoolean',
              output: "if (((Boolean(1)) && []) || ('a' && {})) void 0;",
              desc: 'Explicitly convert value to a boolean (`Boolean(value)`)',
            },
          ],
          message:
            'Unexpected number value in conditional. An explicit zero/NaN check is required.',
          endLine: 1,
          endColumn: 7,
        },
        {
          column: 11,
          line: 1,
          messageId: 'conditionErrorObject',
          message:
            'Unexpected object value in conditional. The condition is always true.',
          endLine: 1,
          endColumn: 13,
        },
        {
          column: 19,
          line: 1,
          messageId: 'conditionErrorString',
          suggestions: [
            {
              messageId: 'conditionFixCompareStringLength',
              output: "if ((1 && []) || (('a'.length > 0) && {})) void 0;",
              desc: "Change condition to check string's length (`value.length !== 0`)",
            },
            {
              messageId: 'conditionFixCompareEmptyString',
              output: 'if ((1 && []) || ((\'a\' !== "") && {})) void 0;',
              desc: 'Change condition to check for empty string (`value !== ""`)',
            },
            {
              messageId: 'conditionFixCastBoolean',
              output: "if ((1 && []) || ((Boolean('a')) && {})) void 0;",
              desc: 'Explicitly convert value to a boolean (`Boolean(value)`)',
            },
          ],
          message:
            'Unexpected string value in conditional. An explicit empty string check is required.',
          endLine: 1,
          endColumn: 22,
        },
        {
          column: 26,
          line: 1,
          messageId: 'conditionErrorObject',
          message:
            'Unexpected object value in conditional. The condition is always true.',
          endLine: 1,
          endColumn: 28,
        },
      ],
      options: [
        {
          allowNumber: false,
          allowString: false,
        },
      ],
    },
    {
      code: "let x = null || 0 || 'a' || [] ? {} : undefined;",
      errors: [
        {
          column: 9,
          line: 1,
          messageId: 'conditionErrorNullish',
          message:
            'Unexpected nullish value in conditional. The condition is always false.',
          endLine: 1,
          endColumn: 13,
        },
        {
          column: 17,
          line: 1,
          messageId: 'conditionErrorNumber',
          suggestions: [
            {
              messageId: 'conditionFixCompareZero',
              output:
                "let x = null || (0 !== 0) || 'a' || [] ? {} : undefined;",
              desc: 'Change condition to check for 0 (`value !== 0`)',
            },
            {
              messageId: 'conditionFixCompareNaN',
              output:
                "let x = null || (!Number.isNaN(0)) || 'a' || [] ? {} : undefined;",
              desc: 'Change condition to check for NaN (`!Number.isNaN(value)`)',
            },
            {
              messageId: 'conditionFixCastBoolean',
              output:
                "let x = null || (Boolean(0)) || 'a' || [] ? {} : undefined;",
              desc: 'Explicitly convert value to a boolean (`Boolean(value)`)',
            },
          ],
          message:
            'Unexpected number value in conditional. An explicit zero/NaN check is required.',
          endLine: 1,
          endColumn: 18,
        },
        {
          column: 22,
          line: 1,
          messageId: 'conditionErrorString',
          suggestions: [
            {
              messageId: 'conditionFixCompareStringLength',
              output:
                "let x = null || 0 || ('a'.length > 0) || [] ? {} : undefined;",
              desc: "Change condition to check string's length (`value.length !== 0`)",
            },
            {
              messageId: 'conditionFixCompareEmptyString',
              output:
                'let x = null || 0 || (\'a\' !== "") || [] ? {} : undefined;',
              desc: 'Change condition to check for empty string (`value !== ""`)',
            },
            {
              messageId: 'conditionFixCastBoolean',
              output:
                "let x = null || 0 || (Boolean('a')) || [] ? {} : undefined;",
              desc: 'Explicitly convert value to a boolean (`Boolean(value)`)',
            },
          ],
          message:
            'Unexpected string value in conditional. An explicit empty string check is required.',
          endLine: 1,
          endColumn: 25,
        },
        {
          column: 29,
          line: 1,
          messageId: 'conditionErrorObject',
          message:
            'Unexpected object value in conditional. The condition is always true.',
          endLine: 1,
          endColumn: 31,
        },
      ],
      options: [
        {
          allowNumber: false,
          allowString: false,
        },
      ],
    },
    {
      code: "return !(null || 0 || 'a' || []);",
      errors: [
        {
          column: 10,
          line: 1,
          messageId: 'conditionErrorNullish',
          message:
            'Unexpected nullish value in conditional. The condition is always false.',
          endLine: 1,
          endColumn: 14,
        },
        {
          column: 18,
          line: 1,
          messageId: 'conditionErrorNumber',
          suggestions: [
            {
              messageId: 'conditionFixCompareZero',
              output: "return !(null || (0 !== 0) || 'a' || []);",
              desc: 'Change condition to check for 0 (`value !== 0`)',
            },
            {
              messageId: 'conditionFixCompareNaN',
              output: "return !(null || (!Number.isNaN(0)) || 'a' || []);",
              desc: 'Change condition to check for NaN (`!Number.isNaN(value)`)',
            },
            {
              messageId: 'conditionFixCastBoolean',
              output: "return !(null || (Boolean(0)) || 'a' || []);",
              desc: 'Explicitly convert value to a boolean (`Boolean(value)`)',
            },
          ],
          message:
            'Unexpected number value in conditional. An explicit zero/NaN check is required.',
          endLine: 1,
          endColumn: 19,
        },
        {
          column: 23,
          line: 1,
          messageId: 'conditionErrorString',
          suggestions: [
            {
              messageId: 'conditionFixCompareStringLength',
              output: "return !(null || 0 || ('a'.length > 0) || []);",
              desc: "Change condition to check string's length (`value.length !== 0`)",
            },
            {
              messageId: 'conditionFixCompareEmptyString',
              output: 'return !(null || 0 || (\'a\' !== "") || []);',
              desc: 'Change condition to check for empty string (`value !== ""`)',
            },
            {
              messageId: 'conditionFixCastBoolean',
              output: "return !(null || 0 || (Boolean('a')) || []);",
              desc: 'Explicitly convert value to a boolean (`Boolean(value)`)',
            },
          ],
          message:
            'Unexpected string value in conditional. An explicit empty string check is required.',
          endLine: 1,
          endColumn: 26,
        },
        {
          column: 30,
          line: 1,
          messageId: 'conditionErrorObject',
          message:
            'Unexpected object value in conditional. The condition is always true.',
          endLine: 1,
          endColumn: 32,
        },
      ],
      options: [
        {
          allowNumber: false,
          allowString: false,
        },
      ],
    },
    {
      code: 'null || {};',
      errors: [
        {
          column: 1,
          line: 1,
          messageId: 'conditionErrorNullish',
          message:
            'Unexpected nullish value in conditional. The condition is always false.',
          endLine: 1,
          endColumn: 5,
        },
      ],
    },
    {
      code: 'undefined && [];',
      errors: [
        {
          column: 1,
          line: 1,
          messageId: 'conditionErrorNullish',
          message:
            'Unexpected nullish value in conditional. The condition is always false.',
          endLine: 1,
          endColumn: 10,
        },
      ],
    },
    {
      code: `
declare const x: null;
if (x) {
}
      `,
      errors: [
        {
          column: 5,
          line: 3,
          messageId: 'conditionErrorNullish',
          message:
            'Unexpected nullish value in conditional. The condition is always false.',
          endLine: 3,
          endColumn: 6,
        },
      ],
    },
    {
      code: '(x: undefined) => !x;',
      errors: [
        {
          column: 20,
          line: 1,
          messageId: 'conditionErrorNullish',
          message:
            'Unexpected nullish value in conditional. The condition is always false.',
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: '<T extends null | undefined>(x: T) => (x ? 1 : 0);',
      errors: [
        {
          column: 40,
          line: 1,
          messageId: 'conditionErrorNullish',
          message:
            'Unexpected nullish value in conditional. The condition is always false.',
          endLine: 1,
          endColumn: 41,
        },
      ],
    },
    {
      code: '<T extends null>(x: T) => (x ? 1 : 0);',
      errors: [
        {
          column: 28,
          line: 1,
          messageId: 'conditionErrorNullish',
          message:
            'Unexpected nullish value in conditional. The condition is always false.',
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      code: '<T extends undefined>(x: T) => (x ? 1 : 0);',
      errors: [
        {
          column: 33,
          line: 1,
          messageId: 'conditionErrorNullish',
          message:
            'Unexpected nullish value in conditional. The condition is always false.',
          endLine: 1,
          endColumn: 34,
        },
      ],
    },
    {
      code: '[] || 1;',
      errors: [
        {
          column: 1,
          line: 1,
          messageId: 'conditionErrorObject',
          message:
            'Unexpected object value in conditional. The condition is always true.',
          endLine: 1,
          endColumn: 3,
        },
      ],
    },
    {
      code: "({}) && 'a';",
      errors: [
        {
          column: 2,
          line: 1,
          messageId: 'conditionErrorObject',
          message:
            'Unexpected object value in conditional. The condition is always true.',
          endLine: 1,
          endColumn: 4,
        },
      ],
    },
    {
      code: `
declare const x: symbol;
if (x) {
}
      `,
      errors: [
        {
          column: 5,
          line: 3,
          messageId: 'conditionErrorObject',
          message:
            'Unexpected object value in conditional. The condition is always true.',
          endLine: 3,
          endColumn: 6,
        },
      ],
    },
    {
      code: '(x: () => void) => !x;',
      errors: [
        {
          column: 21,
          line: 1,
          messageId: 'conditionErrorObject',
          message:
            'Unexpected object value in conditional. The condition is always true.',
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: '<T extends object>(x: T) => (x ? 1 : 0);',
      errors: [
        {
          column: 30,
          line: 1,
          messageId: 'conditionErrorObject',
          message:
            'Unexpected object value in conditional. The condition is always true.',
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: '<T extends Object | Function>(x: T) => (x ? 1 : 0);',
      errors: [
        {
          column: 41,
          line: 1,
          messageId: 'conditionErrorObject',
          message:
            'Unexpected object value in conditional. The condition is always true.',
          endLine: 1,
          endColumn: 42,
        },
      ],
    },
    {
      code: '<T extends { a: number }>(x: T) => (x ? 1 : 0);',
      errors: [
        {
          column: 37,
          line: 1,
          messageId: 'conditionErrorObject',
          message:
            'Unexpected object value in conditional. The condition is always true.',
          endLine: 1,
          endColumn: 38,
        },
      ],
    },
    {
      code: '<T extends () => void>(x: T) => (x ? 1 : 0);',
      errors: [
        {
          column: 34,
          line: 1,
          messageId: 'conditionErrorObject',
          message:
            'Unexpected object value in conditional. The condition is always true.',
          endLine: 1,
          endColumn: 35,
        },
      ],
    },
    {
      code: "while ('') {}",
      errors: [
        {
          column: 8,
          line: 1,
          messageId: 'conditionErrorString',
          suggestions: [
            {
              messageId: 'conditionFixCompareStringLength',
              output: "while (''.length > 0) {}",
              desc: "Change condition to check string's length (`value.length !== 0`)",
            },
            {
              messageId: 'conditionFixCompareEmptyString',
              output: 'while (\'\' !== "") {}',
              desc: 'Change condition to check for empty string (`value !== ""`)',
            },
            {
              messageId: 'conditionFixCastBoolean',
              output: "while (Boolean('')) {}",
              desc: 'Explicitly convert value to a boolean (`Boolean(value)`)',
            },
          ],
          message:
            'Unexpected string value in conditional. An explicit empty string check is required.',
          endLine: 1,
          endColumn: 10,
        },
      ],
      options: [
        {
          allowString: false,
        },
      ],
    },
    {
      code: "for (; 'foo';) {}",
      errors: [
        {
          column: 8,
          line: 1,
          messageId: 'conditionErrorString',
          suggestions: [
            {
              messageId: 'conditionFixCompareStringLength',
              output: "for (; 'foo'.length > 0;) {}",
              desc: "Change condition to check string's length (`value.length !== 0`)",
            },
            {
              messageId: 'conditionFixCompareEmptyString',
              output: 'for (; \'foo\' !== "";) {}',
              desc: 'Change condition to check for empty string (`value !== ""`)',
            },
            {
              messageId: 'conditionFixCastBoolean',
              output: "for (; Boolean('foo');) {}",
              desc: 'Explicitly convert value to a boolean (`Boolean(value)`)',
            },
          ],
          message:
            'Unexpected string value in conditional. An explicit empty string check is required.',
          endLine: 1,
          endColumn: 13,
        },
      ],
      options: [
        {
          allowString: false,
        },
      ],
    },
    {
      code: `
declare const x: string;
if (x) {
}
      `,
      errors: [
        {
          column: 5,
          line: 3,
          messageId: 'conditionErrorString',
          suggestions: [
            {
              messageId: 'conditionFixCompareStringLength',
              output: `
declare const x: string;
if (x.length > 0) {
}
      `,
              desc: "Change condition to check string's length (`value.length !== 0`)",
            },
            {
              messageId: 'conditionFixCompareEmptyString',
              output: `
declare const x: string;
if (x !== "") {
}
      `,
              desc: 'Change condition to check for empty string (`value !== ""`)',
            },
            {
              messageId: 'conditionFixCastBoolean',
              output: `
declare const x: string;
if (Boolean(x)) {
}
      `,
              desc: 'Explicitly convert value to a boolean (`Boolean(value)`)',
            },
          ],
          message:
            'Unexpected string value in conditional. An explicit empty string check is required.',
          endLine: 3,
          endColumn: 6,
        },
      ],
      options: [
        {
          allowString: false,
        },
      ],
    },
    {
      code: '(x: string) => !x;',
      errors: [
        {
          column: 17,
          line: 1,
          messageId: 'conditionErrorString',
          suggestions: [
            {
              messageId: 'conditionFixCompareStringLength',
              output: '(x: string) => x.length === 0;',
              desc: "Change condition to check string's length (`value.length !== 0`)",
            },
            {
              messageId: 'conditionFixCompareEmptyString',
              output: '(x: string) => x === "";',
              desc: 'Change condition to check for empty string (`value !== ""`)',
            },
            {
              messageId: 'conditionFixCastBoolean',
              output: '(x: string) => !Boolean(x);',
              desc: 'Explicitly convert value to a boolean (`Boolean(value)`)',
            },
          ],
          message:
            'Unexpected string value in conditional. An explicit empty string check is required.',
          endLine: 1,
          endColumn: 18,
        },
      ],
      options: [
        {
          allowString: false,
        },
      ],
    },
    {
      code: '<T extends string>(x: T) => (x ? 1 : 0);',
      errors: [
        {
          column: 30,
          line: 1,
          messageId: 'conditionErrorString',
          suggestions: [
            {
              messageId: 'conditionFixCompareStringLength',
              output: '<T extends string>(x: T) => ((x.length > 0) ? 1 : 0);',
              desc: "Change condition to check string's length (`value.length !== 0`)",
            },
            {
              messageId: 'conditionFixCompareEmptyString',
              output: '<T extends string>(x: T) => ((x !== "") ? 1 : 0);',
              desc: 'Change condition to check for empty string (`value !== ""`)',
            },
            {
              messageId: 'conditionFixCastBoolean',
              output: '<T extends string>(x: T) => ((Boolean(x)) ? 1 : 0);',
              desc: 'Explicitly convert value to a boolean (`Boolean(value)`)',
            },
          ],
          message:
            'Unexpected string value in conditional. An explicit empty string check is required.',
          endLine: 1,
          endColumn: 31,
        },
      ],
      options: [
        {
          allowString: false,
        },
      ],
    },
    {
      code: 'while (0n) {}',
      errors: [
        {
          column: 8,
          line: 1,
          messageId: 'conditionErrorNumber',
          suggestions: [
            {
              messageId: 'conditionFixCompareZero',
              output: 'while (0n !== 0) {}',
              desc: 'Change condition to check for 0 (`value !== 0`)',
            },
            {
              messageId: 'conditionFixCompareNaN',
              output: 'while (!Number.isNaN(0n)) {}',
              desc: 'Change condition to check for NaN (`!Number.isNaN(value)`)',
            },
            {
              messageId: 'conditionFixCastBoolean',
              output: 'while (Boolean(0n)) {}',
              desc: 'Explicitly convert value to a boolean (`Boolean(value)`)',
            },
          ],
          message:
            'Unexpected number value in conditional. An explicit zero/NaN check is required.',
          endLine: 1,
          endColumn: 10,
        },
      ],
      options: [
        {
          allowNumber: false,
        },
      ],
    },
    {
      code: 'for (; 123;) {}',
      errors: [
        {
          column: 8,
          line: 1,
          messageId: 'conditionErrorNumber',
          suggestions: [
            {
              messageId: 'conditionFixCompareZero',
              output: 'for (; 123 !== 0;) {}',
              desc: 'Change condition to check for 0 (`value !== 0`)',
            },
            {
              messageId: 'conditionFixCompareNaN',
              output: 'for (; !Number.isNaN(123);) {}',
              desc: 'Change condition to check for NaN (`!Number.isNaN(value)`)',
            },
            {
              messageId: 'conditionFixCastBoolean',
              output: 'for (; Boolean(123);) {}',
              desc: 'Explicitly convert value to a boolean (`Boolean(value)`)',
            },
          ],
          message:
            'Unexpected number value in conditional. An explicit zero/NaN check is required.',
          endLine: 1,
          endColumn: 11,
        },
      ],
      options: [
        {
          allowNumber: false,
        },
      ],
    },
    {
      code: `
declare const x: number;
if (x) {
}
      `,
      errors: [
        {
          column: 5,
          line: 3,
          messageId: 'conditionErrorNumber',
          suggestions: [
            {
              messageId: 'conditionFixCompareZero',
              output: `
declare const x: number;
if (x !== 0) {
}
      `,
              desc: 'Change condition to check for 0 (`value !== 0`)',
            },
            {
              messageId: 'conditionFixCompareNaN',
              output: `
declare const x: number;
if (!Number.isNaN(x)) {
}
      `,
              desc: 'Change condition to check for NaN (`!Number.isNaN(value)`)',
            },
            {
              messageId: 'conditionFixCastBoolean',
              output: `
declare const x: number;
if (Boolean(x)) {
}
      `,
              desc: 'Explicitly convert value to a boolean (`Boolean(value)`)',
            },
          ],
          message:
            'Unexpected number value in conditional. An explicit zero/NaN check is required.',
          endLine: 3,
          endColumn: 6,
        },
      ],
      options: [
        {
          allowNumber: false,
        },
      ],
    },
    {
      code: '(x: bigint) => !x;',
      errors: [
        {
          column: 17,
          line: 1,
          messageId: 'conditionErrorNumber',
          suggestions: [
            {
              messageId: 'conditionFixCompareZero',
              output: '(x: bigint) => x === 0;',
              desc: 'Change condition to check for 0 (`value !== 0`)',
            },
            {
              messageId: 'conditionFixCompareNaN',
              output: '(x: bigint) => Number.isNaN(x);',
              desc: 'Change condition to check for NaN (`!Number.isNaN(value)`)',
            },
            {
              messageId: 'conditionFixCastBoolean',
              output: '(x: bigint) => !Boolean(x);',
              desc: 'Explicitly convert value to a boolean (`Boolean(value)`)',
            },
          ],
          message:
            'Unexpected number value in conditional. An explicit zero/NaN check is required.',
          endLine: 1,
          endColumn: 18,
        },
      ],
      options: [
        {
          allowNumber: false,
        },
      ],
    },
    {
      code: '<T extends number>(x: T) => (x ? 1 : 0);',
      errors: [
        {
          column: 30,
          line: 1,
          messageId: 'conditionErrorNumber',
          suggestions: [
            {
              messageId: 'conditionFixCompareZero',
              output: '<T extends number>(x: T) => ((x !== 0) ? 1 : 0);',
              desc: 'Change condition to check for 0 (`value !== 0`)',
            },
            {
              messageId: 'conditionFixCompareNaN',
              output:
                '<T extends number>(x: T) => ((!Number.isNaN(x)) ? 1 : 0);',
              desc: 'Change condition to check for NaN (`!Number.isNaN(value)`)',
            },
            {
              messageId: 'conditionFixCastBoolean',
              output: '<T extends number>(x: T) => ((Boolean(x)) ? 1 : 0);',
              desc: 'Explicitly convert value to a boolean (`Boolean(value)`)',
            },
          ],
          message:
            'Unexpected number value in conditional. An explicit zero/NaN check is required.',
          endLine: 1,
          endColumn: 31,
        },
      ],
      options: [
        {
          allowNumber: false,
        },
      ],
    },
    {
      code: "![]['length']; // doesn't count as array.length when computed",
      errors: [
        {
          column: 2,
          line: 1,
          messageId: 'conditionErrorNumber',
          suggestions: [
            {
              messageId: 'conditionFixCompareZero',
              output:
                "[]['length'] === 0; // doesn't count as array.length when computed",
              desc: 'Change condition to check for 0 (`value !== 0`)',
            },
            {
              messageId: 'conditionFixCompareNaN',
              output:
                "Number.isNaN([]['length']); // doesn't count as array.length when computed",
              desc: 'Change condition to check for NaN (`!Number.isNaN(value)`)',
            },
            {
              messageId: 'conditionFixCastBoolean',
              output:
                "!Boolean([]['length']); // doesn't count as array.length when computed",
              desc: 'Explicitly convert value to a boolean (`Boolean(value)`)',
            },
          ],
          message:
            'Unexpected number value in conditional. An explicit zero/NaN check is required.',
          endLine: 1,
          endColumn: 14,
        },
      ],
      options: [
        {
          allowNumber: false,
        },
      ],
    },
    {
      code: `
declare const a: any[] & { notLength: number };
if (a.notLength) {
}
      `,
      errors: [
        {
          column: 5,
          line: 3,
          messageId: 'conditionErrorNumber',
          suggestions: [
            {
              messageId: 'conditionFixCompareZero',
              output: `
declare const a: any[] & { notLength: number };
if (a.notLength !== 0) {
}
      `,
              desc: 'Change condition to check for 0 (`value !== 0`)',
            },
            {
              messageId: 'conditionFixCompareNaN',
              output: `
declare const a: any[] & { notLength: number };
if (!Number.isNaN(a.notLength)) {
}
      `,
              desc: 'Change condition to check for NaN (`!Number.isNaN(value)`)',
            },
            {
              messageId: 'conditionFixCastBoolean',
              output: `
declare const a: any[] & { notLength: number };
if (Boolean(a.notLength)) {
}
      `,
              desc: 'Explicitly convert value to a boolean (`Boolean(value)`)',
            },
          ],
          message:
            'Unexpected number value in conditional. An explicit zero/NaN check is required.',
          endLine: 3,
          endColumn: 16,
        },
      ],
      options: [
        {
          allowNumber: false,
        },
      ],
    },
    {
      code: `
if (![].length) {
}
      `,
      errors: [
        {
          column: 6,
          line: 2,
          messageId: 'conditionErrorNumber',
          suggestions: [
            {
              messageId: 'conditionFixCompareArrayLengthZero',
              output: `
if ([].length === 0) {
}
      `,
              desc: "Change condition to check array's length (`value.length === 0`)",
            },
          ],
          message:
            'Unexpected number value in conditional. An explicit zero/NaN check is required.',
          endLine: 2,
          endColumn: 15,
        },
      ],
      options: [
        {
          allowNumber: false,
        },
      ],
    },
    {
      code: `
(a: number[]) => a.length && '...';
      `,
      errors: [
        {
          column: 18,
          line: 2,
          messageId: 'conditionErrorNumber',
          suggestions: [
            {
              messageId: 'conditionFixCompareArrayLengthNonzero',
              output: `
(a: number[]) => (a.length > 0) && '...';
      `,
              desc: "Change condition to check array's length (`value.length > 0`)",
            },
          ],
          message:
            'Unexpected number value in conditional. An explicit zero/NaN check is required.',
          endLine: 2,
          endColumn: 26,
        },
      ],
      options: [
        {
          allowNumber: false,
        },
      ],
    },
    {
      code: `
<T extends unknown[]>(...a: T) => a.length || 'empty';
      `,
      errors: [
        {
          column: 35,
          line: 2,
          messageId: 'conditionErrorNumber',
          suggestions: [
            {
              messageId: 'conditionFixCompareArrayLengthNonzero',
              output: `
<T extends unknown[]>(...a: T) => (a.length > 0) || 'empty';
      `,
              desc: "Change condition to check array's length (`value.length > 0`)",
            },
          ],
          message:
            'Unexpected number value in conditional. An explicit zero/NaN check is required.',
          endLine: 2,
          endColumn: 43,
        },
      ],
      options: [
        {
          allowNumber: false,
        },
      ],
    },
    {
      code: `
declare const x: string | number;
if (x) {
}
      `,
      errors: [
        {
          column: 5,
          line: 3,
          messageId: 'conditionErrorOther',
          message:
            'Unexpected value in conditional. A boolean expression is required.',
          endLine: 3,
          endColumn: 6,
        },
      ],
      options: [
        {
          allowNumber: true,
          allowString: true,
        },
      ],
    },
    {
      code: '(x: bigint | string) => !x;',
      errors: [
        {
          column: 26,
          line: 1,
          messageId: 'conditionErrorOther',
          message:
            'Unexpected value in conditional. A boolean expression is required.',
          endLine: 1,
          endColumn: 27,
        },
      ],
      options: [
        {
          allowNumber: true,
          allowString: true,
        },
      ],
    },
    {
      code: '<T extends number | bigint | string>(x: T) => (x ? 1 : 0);',
      errors: [
        {
          column: 48,
          line: 1,
          messageId: 'conditionErrorOther',
          message:
            'Unexpected value in conditional. A boolean expression is required.',
          endLine: 1,
          endColumn: 49,
        },
      ],
      options: [
        {
          allowNumber: true,
          allowString: true,
        },
      ],
    },
    {
      code: `
declare const x: boolean | null;
if (x) {
}
      `,
      errors: [
        {
          column: 5,
          line: 3,
          messageId: 'conditionErrorNullableBoolean',
          suggestions: [
            {
              messageId: 'conditionFixDefaultFalse',
              output: `
declare const x: boolean | null;
if (x ?? false) {
}
      `,
              desc: 'Explicitly treat nullish value the same as false (`value ?? false`)',
            },
            {
              messageId: 'conditionFixCompareTrue',
              output: `
declare const x: boolean | null;
if (x === true) {
}
      `,
              desc: 'Change condition to check if true (`value === true`)',
            },
          ],
          message:
            'Unexpected nullable boolean value in conditional. Please handle the nullish case explicitly.',
          endLine: 3,
          endColumn: 6,
        },
      ],
      options: [
        {
          allowNullableBoolean: false,
        },
      ],
    },
    {
      code: '(x?: boolean) => !x;',
      errors: [
        {
          column: 19,
          line: 1,
          messageId: 'conditionErrorNullableBoolean',
          suggestions: [
            {
              messageId: 'conditionFixDefaultFalse',
              output: '(x?: boolean) => !(x ?? false);',
              desc: 'Explicitly treat nullish value the same as false (`value ?? false`)',
            },
            {
              messageId: 'conditionFixCompareFalse',
              output: '(x?: boolean) => x === false;',
              desc: 'Change condition to check if false (`value === false`)',
            },
          ],
          message:
            'Unexpected nullable boolean value in conditional. Please handle the nullish case explicitly.',
          endLine: 1,
          endColumn: 20,
        },
      ],
      options: [
        {
          allowNullableBoolean: false,
        },
      ],
    },
    {
      code: '<T extends boolean | null | undefined>(x: T) => (x ? 1 : 0);',
      errors: [
        {
          column: 50,
          line: 1,
          messageId: 'conditionErrorNullableBoolean',
          suggestions: [
            {
              messageId: 'conditionFixDefaultFalse',
              output:
                '<T extends boolean | null | undefined>(x: T) => ((x ?? false) ? 1 : 0);',
              desc: 'Explicitly treat nullish value the same as false (`value ?? false`)',
            },
            {
              messageId: 'conditionFixCompareTrue',
              output:
                '<T extends boolean | null | undefined>(x: T) => ((x === true) ? 1 : 0);',
              desc: 'Change condition to check if true (`value === true`)',
            },
          ],
          message:
            'Unexpected nullable boolean value in conditional. Please handle the nullish case explicitly.',
          endLine: 1,
          endColumn: 51,
        },
      ],
      options: [
        {
          allowNullableBoolean: false,
        },
      ],
    },
    {
      code: `
declare const x: object | null;
if (x) {
}
      `,
      errors: [
        {
          column: 5,
          line: 3,
          messageId: 'conditionErrorNullableObject',
          suggestions: [
            {
              messageId: 'conditionFixCompareNullish',
              output: `
declare const x: object | null;
if (x != null) {
}
      `,
              desc: 'Change condition to check for null/undefined (`value != null`)',
            },
          ],
          message:
            'Unexpected nullable object value in conditional. An explicit null check is required.',
          endLine: 3,
          endColumn: 6,
        },
      ],
      options: [
        {
          allowNullableObject: false,
        },
      ],
    },
    {
      code: '(x?: { a: number }) => !x;',
      errors: [
        {
          column: 25,
          line: 1,
          messageId: 'conditionErrorNullableObject',
          suggestions: [
            {
              messageId: 'conditionFixCompareNullish',
              output: '(x?: { a: number }) => x == null;',
              desc: 'Change condition to check for null/undefined (`value != null`)',
            },
          ],
          message:
            'Unexpected nullable object value in conditional. An explicit null check is required.',
          endLine: 1,
          endColumn: 26,
        },
      ],
      options: [
        {
          allowNullableObject: false,
        },
      ],
    },
    {
      code: '<T extends {} | null | undefined>(x: T) => (x ? 1 : 0);',
      errors: [
        {
          column: 45,
          line: 1,
          messageId: 'conditionErrorNullableObject',
          suggestions: [
            {
              messageId: 'conditionFixCompareNullish',
              output:
                '<T extends {} | null | undefined>(x: T) => ((x != null) ? 1 : 0);',
              desc: 'Change condition to check for null/undefined (`value != null`)',
            },
          ],
          message:
            'Unexpected nullable object value in conditional. An explicit null check is required.',
          endLine: 1,
          endColumn: 46,
        },
      ],
      options: [
        {
          allowNullableObject: false,
        },
      ],
    },
    {
      code: `
declare const x: string | null;
if (x) {
}
      `,
      errors: [
        {
          column: 5,
          line: 3,
          messageId: 'conditionErrorNullableString',
          suggestions: [
            {
              messageId: 'conditionFixCompareNullish',
              output: `
declare const x: string | null;
if (x != null) {
}
      `,
              desc: 'Change condition to check for null/undefined (`value != null`)',
            },
            {
              messageId: 'conditionFixDefaultEmptyString',
              output: `
declare const x: string | null;
if (x ?? "") {
}
      `,
              desc: 'Explicitly treat nullish value the same as an empty string (`value ?? ""`)',
            },
            {
              messageId: 'conditionFixCastBoolean',
              output: `
declare const x: string | null;
if (Boolean(x)) {
}
      `,
              desc: 'Explicitly convert value to a boolean (`Boolean(value)`)',
            },
          ],
          message:
            'Unexpected nullable string value in conditional. Please handle the nullish/empty cases explicitly.',
          endLine: 3,
          endColumn: 6,
        },
      ],
    },
    {
      code: '(x?: string) => !x;',
      errors: [
        {
          column: 18,
          line: 1,
          messageId: 'conditionErrorNullableString',
          suggestions: [
            {
              messageId: 'conditionFixCompareNullish',
              output: '(x?: string) => x == null;',
              desc: 'Change condition to check for null/undefined (`value != null`)',
            },
            {
              messageId: 'conditionFixDefaultEmptyString',
              output: '(x?: string) => !(x ?? "");',
              desc: 'Explicitly treat nullish value the same as an empty string (`value ?? ""`)',
            },
            {
              messageId: 'conditionFixCastBoolean',
              output: '(x?: string) => !Boolean(x);',
              desc: 'Explicitly convert value to a boolean (`Boolean(value)`)',
            },
          ],
          message:
            'Unexpected nullable string value in conditional. Please handle the nullish/empty cases explicitly.',
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: '<T extends string | null | undefined>(x: T) => (x ? 1 : 0);',
      errors: [
        {
          column: 49,
          line: 1,
          messageId: 'conditionErrorNullableString',
          suggestions: [
            {
              messageId: 'conditionFixCompareNullish',
              output:
                '<T extends string | null | undefined>(x: T) => ((x != null) ? 1 : 0);',
              desc: 'Change condition to check for null/undefined (`value != null`)',
            },
            {
              messageId: 'conditionFixDefaultEmptyString',
              output:
                '<T extends string | null | undefined>(x: T) => ((x ?? "") ? 1 : 0);',
              desc: 'Explicitly treat nullish value the same as an empty string (`value ?? ""`)',
            },
            {
              messageId: 'conditionFixCastBoolean',
              output:
                '<T extends string | null | undefined>(x: T) => ((Boolean(x)) ? 1 : 0);',
              desc: 'Explicitly convert value to a boolean (`Boolean(value)`)',
            },
          ],
          message:
            'Unexpected nullable string value in conditional. Please handle the nullish/empty cases explicitly.',
          endLine: 1,
          endColumn: 50,
        },
      ],
    },
    {
      code: `
function foo(x: '' | 'bar' | null) {
  if (!x) {
  }
}
      `,
      errors: [
        {
          column: 8,
          line: 3,
          messageId: 'conditionErrorNullableString',
          suggestions: [
            {
              messageId: 'conditionFixCompareNullish',
              output: `
function foo(x: '' | 'bar' | null) {
  if (x == null) {
  }
}
      `,
              desc: 'Change condition to check for null/undefined (`value != null`)',
            },
            {
              messageId: 'conditionFixDefaultEmptyString',
              output: `
function foo(x: '' | 'bar' | null) {
  if (!(x ?? "")) {
  }
}
      `,
              desc: 'Explicitly treat nullish value the same as an empty string (`value ?? ""`)',
            },
            {
              messageId: 'conditionFixCastBoolean',
              output: `
function foo(x: '' | 'bar' | null) {
  if (!Boolean(x)) {
  }
}
      `,
              desc: 'Explicitly convert value to a boolean (`Boolean(value)`)',
            },
          ],
          message:
            'Unexpected nullable string value in conditional. Please handle the nullish/empty cases explicitly.',
          endLine: 3,
          endColumn: 9,
        },
      ],
    },
    {
      code: `
declare const x: number | null;
if (x) {
}
      `,
      errors: [
        {
          column: 5,
          line: 3,
          messageId: 'conditionErrorNullableNumber',
          suggestions: [
            {
              messageId: 'conditionFixCompareNullish',
              output: `
declare const x: number | null;
if (x != null) {
}
      `,
              desc: 'Change condition to check for null/undefined (`value != null`)',
            },
            {
              messageId: 'conditionFixDefaultZero',
              output: `
declare const x: number | null;
if (x ?? 0) {
}
      `,
              desc: 'Explicitly treat nullish value the same as 0 (`value ?? 0`)',
            },
            {
              messageId: 'conditionFixCastBoolean',
              output: `
declare const x: number | null;
if (Boolean(x)) {
}
      `,
              desc: 'Explicitly convert value to a boolean (`Boolean(value)`)',
            },
          ],
          message:
            'Unexpected nullable number value in conditional. Please handle the nullish/zero/NaN cases explicitly.',
          endLine: 3,
          endColumn: 6,
        },
      ],
    },
    {
      code: '(x?: number) => !x;',
      errors: [
        {
          column: 18,
          line: 1,
          messageId: 'conditionErrorNullableNumber',
          suggestions: [
            {
              messageId: 'conditionFixCompareNullish',
              output: '(x?: number) => x == null;',
              desc: 'Change condition to check for null/undefined (`value != null`)',
            },
            {
              messageId: 'conditionFixDefaultZero',
              output: '(x?: number) => !(x ?? 0);',
              desc: 'Explicitly treat nullish value the same as 0 (`value ?? 0`)',
            },
            {
              messageId: 'conditionFixCastBoolean',
              output: '(x?: number) => !Boolean(x);',
              desc: 'Explicitly convert value to a boolean (`Boolean(value)`)',
            },
          ],
          message:
            'Unexpected nullable number value in conditional. Please handle the nullish/zero/NaN cases explicitly.',
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: '<T extends number | null | undefined>(x: T) => (x ? 1 : 0);',
      errors: [
        {
          column: 49,
          line: 1,
          messageId: 'conditionErrorNullableNumber',
          suggestions: [
            {
              messageId: 'conditionFixCompareNullish',
              output:
                '<T extends number | null | undefined>(x: T) => ((x != null) ? 1 : 0);',
              desc: 'Change condition to check for null/undefined (`value != null`)',
            },
            {
              messageId: 'conditionFixDefaultZero',
              output:
                '<T extends number | null | undefined>(x: T) => ((x ?? 0) ? 1 : 0);',
              desc: 'Explicitly treat nullish value the same as 0 (`value ?? 0`)',
            },
            {
              messageId: 'conditionFixCastBoolean',
              output:
                '<T extends number | null | undefined>(x: T) => ((Boolean(x)) ? 1 : 0);',
              desc: 'Explicitly convert value to a boolean (`Boolean(value)`)',
            },
          ],
          message:
            'Unexpected nullable number value in conditional. Please handle the nullish/zero/NaN cases explicitly.',
          endLine: 1,
          endColumn: 50,
        },
      ],
    },
    {
      code: `
function foo(x: 0 | 1 | null) {
  if (!x) {
  }
}
      `,
      errors: [
        {
          column: 8,
          line: 3,
          messageId: 'conditionErrorNullableNumber',
          suggestions: [
            {
              messageId: 'conditionFixCompareNullish',
              output: `
function foo(x: 0 | 1 | null) {
  if (x == null) {
  }
}
      `,
              desc: 'Change condition to check for null/undefined (`value != null`)',
            },
            {
              messageId: 'conditionFixDefaultZero',
              output: `
function foo(x: 0 | 1 | null) {
  if (!(x ?? 0)) {
  }
}
      `,
              desc: 'Explicitly treat nullish value the same as 0 (`value ?? 0`)',
            },
            {
              messageId: 'conditionFixCastBoolean',
              output: `
function foo(x: 0 | 1 | null) {
  if (!Boolean(x)) {
  }
}
      `,
              desc: 'Explicitly convert value to a boolean (`Boolean(value)`)',
            },
          ],
          message:
            'Unexpected nullable number value in conditional. Please handle the nullish/zero/NaN cases explicitly.',
          endLine: 3,
          endColumn: 9,
        },
      ],
    },
    {
      code: `
enum ExampleEnum {
  This = 0,
  That = 1,
}
const theEnum = Math.random() < 0.3 ? ExampleEnum.This : null;
if (theEnum) {
}
      `,
      errors: [
        {
          column: 5,
          endColumn: 12,
          endLine: 7,
          line: 7,
          messageId: 'conditionErrorNullableEnum',
          suggestions: [
            {
              messageId: 'conditionFixCompareNullish',
              output: `
enum ExampleEnum {
  This = 0,
  That = 1,
}
const theEnum = Math.random() < 0.3 ? ExampleEnum.This : null;
if (theEnum != null) {
}
      `,
              desc: 'Change condition to check for null/undefined (`value != null`)',
            },
          ],
          message:
            'Unexpected nullable enum value in conditional. Please handle the nullish/zero/NaN cases explicitly.',
        },
      ],
      options: [
        {
          allowNullableEnum: false,
        },
      ],
    },
    {
      code: `
enum ExampleEnum {
  This = 0,
  That = 1,
}
const theEnum = Math.random() < 0.3 ? ExampleEnum.This : null;
if (!theEnum) {
}
      `,
      errors: [
        {
          column: 6,
          endColumn: 13,
          endLine: 7,
          line: 7,
          messageId: 'conditionErrorNullableEnum',
          suggestions: [
            {
              messageId: 'conditionFixCompareNullish',
              output: `
enum ExampleEnum {
  This = 0,
  That = 1,
}
const theEnum = Math.random() < 0.3 ? ExampleEnum.This : null;
if (theEnum == null) {
}
      `,
              desc: 'Change condition to check for null/undefined (`value != null`)',
            },
          ],
          message:
            'Unexpected nullable enum value in conditional. Please handle the nullish/zero/NaN cases explicitly.',
        },
      ],
      options: [
        {
          allowNullableEnum: false,
        },
      ],
    },
    {
      code: `
enum ExampleEnum {
  This,
  That,
}
const theEnum = Math.random() < 0.3 ? ExampleEnum.This : null;
if (!theEnum) {
}
      `,
      errors: [
        {
          column: 6,
          endColumn: 13,
          endLine: 7,
          line: 7,
          messageId: 'conditionErrorNullableEnum',
          suggestions: [
            {
              messageId: 'conditionFixCompareNullish',
              output: `
enum ExampleEnum {
  This,
  That,
}
const theEnum = Math.random() < 0.3 ? ExampleEnum.This : null;
if (theEnum == null) {
}
      `,
              desc: 'Change condition to check for null/undefined (`value != null`)',
            },
          ],
          message:
            'Unexpected nullable enum value in conditional. Please handle the nullish/zero/NaN cases explicitly.',
        },
      ],
      options: [
        {
          allowNullableEnum: false,
        },
      ],
    },
    {
      code: `
enum ExampleEnum {
  This = '',
  That = 'a',
}
const theEnum = Math.random() < 0.3 ? ExampleEnum.This : null;
if (!theEnum) {
}
      `,
      errors: [
        {
          column: 6,
          endColumn: 13,
          endLine: 7,
          line: 7,
          messageId: 'conditionErrorNullableEnum',
          suggestions: [
            {
              messageId: 'conditionFixCompareNullish',
              output: `
enum ExampleEnum {
  This = '',
  That = 'a',
}
const theEnum = Math.random() < 0.3 ? ExampleEnum.This : null;
if (theEnum == null) {
}
      `,
              desc: 'Change condition to check for null/undefined (`value != null`)',
            },
          ],
          message:
            'Unexpected nullable enum value in conditional. Please handle the nullish/zero/NaN cases explicitly.',
        },
      ],
      options: [
        {
          allowNullableEnum: false,
        },
      ],
    },
    {
      code: `
enum ExampleEnum {
  This = '',
  That = 0,
}
const theEnum = Math.random() < 0.3 ? ExampleEnum.This : null;
if (!theEnum) {
}
      `,
      errors: [
        {
          column: 6,
          endColumn: 13,
          endLine: 7,
          line: 7,
          messageId: 'conditionErrorNullableEnum',
          suggestions: [
            {
              messageId: 'conditionFixCompareNullish',
              output: `
enum ExampleEnum {
  This = '',
  That = 0,
}
const theEnum = Math.random() < 0.3 ? ExampleEnum.This : null;
if (theEnum == null) {
}
      `,
              desc: 'Change condition to check for null/undefined (`value != null`)',
            },
          ],
          message:
            'Unexpected nullable enum value in conditional. Please handle the nullish/zero/NaN cases explicitly.',
        },
      ],
      options: [
        {
          allowNullableEnum: false,
        },
      ],
    },
    {
      code: `
enum ExampleEnum {
  This = 'one',
  That = 'two',
}
const theEnum = Math.random() < 0.3 ? ExampleEnum.This : null;
if (!theEnum) {
}
      `,
      errors: [
        {
          column: 6,
          endColumn: 13,
          endLine: 7,
          line: 7,
          messageId: 'conditionErrorNullableEnum',
          suggestions: [
            {
              messageId: 'conditionFixCompareNullish',
              output: `
enum ExampleEnum {
  This = 'one',
  That = 'two',
}
const theEnum = Math.random() < 0.3 ? ExampleEnum.This : null;
if (theEnum == null) {
}
      `,
              desc: 'Change condition to check for null/undefined (`value != null`)',
            },
          ],
          message:
            'Unexpected nullable enum value in conditional. Please handle the nullish/zero/NaN cases explicitly.',
        },
      ],
      options: [
        {
          allowNullableEnum: false,
        },
      ],
    },
    {
      code: `
enum ExampleEnum {
  This = 1,
  That = 2,
}
const theEnum = Math.random() < 0.3 ? ExampleEnum.This : null;
if (!theEnum) {
}
      `,
      errors: [
        {
          column: 6,
          endColumn: 13,
          endLine: 7,
          line: 7,
          messageId: 'conditionErrorNullableEnum',
          suggestions: [
            {
              messageId: 'conditionFixCompareNullish',
              output: `
enum ExampleEnum {
  This = 1,
  That = 2,
}
const theEnum = Math.random() < 0.3 ? ExampleEnum.This : null;
if (theEnum == null) {
}
      `,
              desc: 'Change condition to check for null/undefined (`value != null`)',
            },
          ],
          message:
            'Unexpected nullable enum value in conditional. Please handle the nullish/zero/NaN cases explicitly.',
        },
      ],
      options: [
        {
          allowNullableEnum: false,
        },
      ],
    },
    {
      code: `
enum ExampleEnum {
  This = 0,
  That = 'one',
}
(value?: ExampleEnum) => (value ? 1 : 0);
      `,
      errors: [
        {
          column: 27,
          endColumn: 32,
          endLine: 6,
          line: 6,
          messageId: 'conditionErrorNullableEnum',
          suggestions: [
            {
              messageId: 'conditionFixCompareNullish',
              output: `
enum ExampleEnum {
  This = 0,
  That = 'one',
}
(value?: ExampleEnum) => ((value != null) ? 1 : 0);
      `,
              desc: 'Change condition to check for null/undefined (`value != null`)',
            },
          ],
          message:
            'Unexpected nullable enum value in conditional. Please handle the nullish/zero/NaN cases explicitly.',
        },
      ],
      options: [
        {
          allowNullableEnum: false,
        },
      ],
    },
    {
      code: `
enum ExampleEnum {
  This = '',
  That = 1,
}
(value?: ExampleEnum) => (!value ? 1 : 0);
      `,
      errors: [
        {
          column: 28,
          endColumn: 33,
          endLine: 6,
          line: 6,
          messageId: 'conditionErrorNullableEnum',
          suggestions: [
            {
              messageId: 'conditionFixCompareNullish',
              output: `
enum ExampleEnum {
  This = '',
  That = 1,
}
(value?: ExampleEnum) => ((value == null) ? 1 : 0);
      `,
              desc: 'Change condition to check for null/undefined (`value != null`)',
            },
          ],
          message:
            'Unexpected nullable enum value in conditional. Please handle the nullish/zero/NaN cases explicitly.',
        },
      ],
      options: [
        {
          allowNullableEnum: false,
        },
      ],
    },
    {
      code: `
enum ExampleEnum {
  This = 'this',
  That = 1,
}
(value?: ExampleEnum) => (!value ? 1 : 0);
      `,
      errors: [
        {
          column: 28,
          endColumn: 33,
          endLine: 6,
          line: 6,
          messageId: 'conditionErrorNullableEnum',
          suggestions: [
            {
              messageId: 'conditionFixCompareNullish',
              output: `
enum ExampleEnum {
  This = 'this',
  That = 1,
}
(value?: ExampleEnum) => ((value == null) ? 1 : 0);
      `,
              desc: 'Change condition to check for null/undefined (`value != null`)',
            },
          ],
          message:
            'Unexpected nullable enum value in conditional. Please handle the nullish/zero/NaN cases explicitly.',
        },
      ],
      options: [
        {
          allowNullableEnum: false,
        },
      ],
    },
    {
      code: `
enum ExampleEnum {
  This = '',
  That = 0,
}
(value?: ExampleEnum) => (!value ? 1 : 0);
      `,
      errors: [
        {
          column: 28,
          endColumn: 33,
          endLine: 6,
          line: 6,
          messageId: 'conditionErrorNullableEnum',
          suggestions: [
            {
              messageId: 'conditionFixCompareNullish',
              output: `
enum ExampleEnum {
  This = '',
  That = 0,
}
(value?: ExampleEnum) => ((value == null) ? 1 : 0);
      `,
              desc: 'Change condition to check for null/undefined (`value != null`)',
            },
          ],
          message:
            'Unexpected nullable enum value in conditional. Please handle the nullish/zero/NaN cases explicitly.',
        },
      ],
      options: [
        {
          allowNullableEnum: false,
        },
      ],
    },
    {
      code: `
if (x) {
}
      `,
      errors: [
        {
          column: 5,
          line: 2,
          messageId: 'conditionErrorAny',
          suggestions: [
            {
              messageId: 'conditionFixCastBoolean',
              output: `
if (Boolean(x)) {
}
      `,
              desc: 'Explicitly convert value to a boolean (`Boolean(value)`)',
            },
          ],
          message:
            'Unexpected any value in conditional. An explicit comparison or type conversion is required.',
          endLine: 2,
          endColumn: 6,
        },
      ],
    },
    {
      code: 'x => !x;',
      errors: [
        {
          column: 7,
          line: 1,
          messageId: 'conditionErrorAny',
          suggestions: [
            {
              messageId: 'conditionFixCastBoolean',
              output: 'x => !(Boolean(x));',
              desc: 'Explicitly convert value to a boolean (`Boolean(value)`)',
            },
          ],
          message:
            'Unexpected any value in conditional. An explicit comparison or type conversion is required.',
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: '<T extends any>(x: T) => (x ? 1 : 0);',
      errors: [
        {
          column: 27,
          line: 1,
          messageId: 'conditionErrorAny',
          suggestions: [
            {
              messageId: 'conditionFixCastBoolean',
              output: '<T extends any>(x: T) => ((Boolean(x)) ? 1 : 0);',
              desc: 'Explicitly convert value to a boolean (`Boolean(value)`)',
            },
          ],
          message:
            'Unexpected any value in conditional. An explicit comparison or type conversion is required.',
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      code: '<T,>(x: T) => (x ? 1 : 0);',
      errors: [
        {
          column: 16,
          line: 1,
          messageId: 'conditionErrorAny',
          suggestions: [
            {
              messageId: 'conditionFixCastBoolean',
              output: '<T,>(x: T) => ((Boolean(x)) ? 1 : 0);',
              desc: 'Explicitly convert value to a boolean (`Boolean(value)`)',
            },
          ],
          message:
            'Unexpected any value in conditional. An explicit comparison or type conversion is required.',
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      code: `
declare const x: string[] | null;
if (x) {
}
      `,
      errors: [
        {
          column: 1,
          line: 1,
          messageId: 'noStrictNullCheck',
          message:
            'This rule requires the `strictNullChecks` compiler option to be turned on to function correctly.',
          endLine: 1,
          endColumn: 1,
        },
        {
          column: 5,
          line: 3,
          messageId: 'conditionErrorObject',
          message:
            'Unexpected object value in conditional. The condition is always true.',
          endLine: 3,
          endColumn: 6,
        },
      ],
      languageOptions: {
        parserOptions: {
          project: ['unstrict/tsconfig.json'],
        },
      },
      filename: 'unstrict/file.ts',
    },
    {
      code: `
        declare const obj: { x: number } | null;
        !obj ? 1 : 0
        !obj
        obj || 0
        obj && 1 || 0
      `,
      errors: [
        {
          column: 10,
          line: 3,
          messageId: 'conditionErrorNullableObject',
          suggestions: [
            {
              messageId: 'conditionFixCompareNullish',
              output: `
        declare const obj: { x: number } | null;
        (obj == null) ? 1 : 0
        !obj
        obj || 0
        obj && 1 || 0
      `,
              desc: 'Change condition to check for null/undefined (`value != null`)',
            },
          ],
          message:
            'Unexpected nullable object value in conditional. An explicit null check is required.',
          endLine: 3,
          endColumn: 13,
        },
        {
          column: 10,
          line: 4,
          messageId: 'conditionErrorNullableObject',
          suggestions: [
            {
              messageId: 'conditionFixCompareNullish',
              output: `
        declare const obj: { x: number } | null;
        !obj ? 1 : 0
        obj == null
        obj || 0
        obj && 1 || 0
      `,
              desc: 'Change condition to check for null/undefined (`value != null`)',
            },
          ],
          message:
            'Unexpected nullable object value in conditional. An explicit null check is required.',
          endLine: 4,
          endColumn: 13,
        },
        {
          column: 9,
          line: 5,
          messageId: 'conditionErrorNullableObject',
          suggestions: [
            {
              messageId: 'conditionFixCompareNullish',
              output: `
        declare const obj: { x: number } | null;
        !obj ? 1 : 0
        !obj
        ;(obj != null) || 0
        obj && 1 || 0
      `,
              desc: 'Change condition to check for null/undefined (`value != null`)',
            },
          ],
          message:
            'Unexpected nullable object value in conditional. An explicit null check is required.',
          endLine: 5,
          endColumn: 12,
        },
        {
          column: 9,
          line: 6,
          messageId: 'conditionErrorNullableObject',
          suggestions: [
            {
              messageId: 'conditionFixCompareNullish',
              output: `
        declare const obj: { x: number } | null;
        !obj ? 1 : 0
        !obj
        obj || 0
        ;(obj != null) && 1 || 0
      `,
              desc: 'Change condition to check for null/undefined (`value != null`)',
            },
          ],
          message:
            'Unexpected nullable object value in conditional. An explicit null check is required.',
          endLine: 6,
          endColumn: 12,
        },
      ],
      options: [
        {
          allowNullableObject: false,
        },
      ],
    },
    {
      code: `
declare function assert(x: unknown): asserts x;
declare const nullableString: string | null;
assert(nullableString);
      `,
      errors: [
        {
          column: 8,
          line: 4,
          messageId: 'conditionErrorNullableString',
          suggestions: [
            {
              messageId: 'conditionFixCompareNullish',
              output: `
declare function assert(x: unknown): asserts x;
declare const nullableString: string | null;
assert(nullableString != null);
      `,
              desc: 'Change condition to check for null/undefined (`value != null`)',
            },
            {
              messageId: 'conditionFixDefaultEmptyString',
              output: `
declare function assert(x: unknown): asserts x;
declare const nullableString: string | null;
assert(nullableString ?? "");
      `,
              desc: 'Explicitly treat nullish value the same as an empty string (`value ?? ""`)',
            },
            {
              messageId: 'conditionFixCastBoolean',
              output: `
declare function assert(x: unknown): asserts x;
declare const nullableString: string | null;
assert(Boolean(nullableString));
      `,
              desc: 'Explicitly convert value to a boolean (`Boolean(value)`)',
            },
          ],
          message:
            'Unexpected nullable string value in conditional. Please handle the nullish/empty cases explicitly.',
          endLine: 4,
          endColumn: 22,
        },
      ],
    },
    {
      code: `
declare function assert(a: number, b: unknown): asserts b;
declare const nullableString: string | null;
assert(foo, nullableString);
      `,
      errors: [
        {
          column: 13,
          line: 4,
          messageId: 'conditionErrorNullableString',
          suggestions: [
            {
              messageId: 'conditionFixCompareNullish',
              output: `
declare function assert(a: number, b: unknown): asserts b;
declare const nullableString: string | null;
assert(foo, nullableString != null);
      `,
              desc: 'Change condition to check for null/undefined (`value != null`)',
            },
            {
              messageId: 'conditionFixDefaultEmptyString',
              output: `
declare function assert(a: number, b: unknown): asserts b;
declare const nullableString: string | null;
assert(foo, nullableString ?? "");
      `,
              desc: 'Explicitly treat nullish value the same as an empty string (`value ?? ""`)',
            },
            {
              messageId: 'conditionFixCastBoolean',
              output: `
declare function assert(a: number, b: unknown): asserts b;
declare const nullableString: string | null;
assert(foo, Boolean(nullableString));
      `,
              desc: 'Explicitly convert value to a boolean (`Boolean(value)`)',
            },
          ],
          message:
            'Unexpected nullable string value in conditional. Please handle the nullish/empty cases explicitly.',
          endLine: 4,
          endColumn: 27,
        },
      ],
    },
    {
      code: `
declare function assert(a: number, b: unknown): asserts b;
declare function assert(one: number, two: unknown): asserts two;
declare const nullableString: string | null;
assert(foo, nullableString);
      `,
      errors: [
        {
          column: 13,
          line: 5,
          messageId: 'conditionErrorNullableString',
          suggestions: [
            {
              messageId: 'conditionFixCompareNullish',
              output: `
declare function assert(a: number, b: unknown): asserts b;
declare function assert(one: number, two: unknown): asserts two;
declare const nullableString: string | null;
assert(foo, nullableString != null);
      `,
              desc: 'Change condition to check for null/undefined (`value != null`)',
            },
            {
              messageId: 'conditionFixDefaultEmptyString',
              output: `
declare function assert(a: number, b: unknown): asserts b;
declare function assert(one: number, two: unknown): asserts two;
declare const nullableString: string | null;
assert(foo, nullableString ?? "");
      `,
              desc: 'Explicitly treat nullish value the same as an empty string (`value ?? ""`)',
            },
            {
              messageId: 'conditionFixCastBoolean',
              output: `
declare function assert(a: number, b: unknown): asserts b;
declare function assert(one: number, two: unknown): asserts two;
declare const nullableString: string | null;
assert(foo, Boolean(nullableString));
      `,
              desc: 'Explicitly convert value to a boolean (`Boolean(value)`)',
            },
          ],
          message:
            'Unexpected nullable string value in conditional. Please handle the nullish/empty cases explicitly.',
          endLine: 5,
          endColumn: 27,
        },
      ],
    },
    {
      code: `
declare function assert(this: object, a: number, b: unknown): asserts b;
declare const nullableString: string | null;
assert(foo, nullableString);
      `,
      errors: [
        {
          column: 13,
          line: 4,
          messageId: 'conditionErrorNullableString',
          suggestions: [
            {
              messageId: 'conditionFixCompareNullish',
              output: `
declare function assert(this: object, a: number, b: unknown): asserts b;
declare const nullableString: string | null;
assert(foo, nullableString != null);
      `,
              desc: 'Change condition to check for null/undefined (`value != null`)',
            },
            {
              messageId: 'conditionFixDefaultEmptyString',
              output: `
declare function assert(this: object, a: number, b: unknown): asserts b;
declare const nullableString: string | null;
assert(foo, nullableString ?? "");
      `,
              desc: 'Explicitly treat nullish value the same as an empty string (`value ?? ""`)',
            },
            {
              messageId: 'conditionFixCastBoolean',
              output: `
declare function assert(this: object, a: number, b: unknown): asserts b;
declare const nullableString: string | null;
assert(foo, Boolean(nullableString));
      `,
              desc: 'Explicitly convert value to a boolean (`Boolean(value)`)',
            },
          ],
          message:
            'Unexpected nullable string value in conditional. Please handle the nullish/empty cases explicitly.',
          endLine: 4,
          endColumn: 27,
        },
      ],
    },
    {
      skip: 'TypeScript does not expose the truthiness predicate of a union of assertion signatures (microsoft/TypeScript#59707); also skipped by typescript-eslint v8.67.0.',
      code: `
function asserts1(x: string | number | undefined): asserts x {}
function asserts2(x: string | number | undefined): asserts x {}

const maybeString = Math.random() ? 'string'.slice() : undefined;

const someAssert: typeof asserts1 | typeof asserts2 =
  Math.random() > 0.5 ? asserts1 : asserts2;

someAssert(maybeString);
      `,
      errors: [
        {
          messageId: 'conditionErrorNullableString',
          suggestions: [
            {
              messageId: 'conditionFixCompareNullish',
              output: `
function asserts1(x: string | number | undefined): asserts x {}
function asserts2(x: string | number | undefined): asserts x {}

const maybeString = Math.random() ? 'string'.slice() : undefined;

const someAssert: typeof asserts1 | typeof asserts2 =
  Math.random() > 0.5 ? asserts1 : asserts2;

someAssert(maybeString != null);
      `,
              desc: 'Change condition to check for null/undefined (`value != null`)',
            },
            {
              messageId: 'conditionFixDefaultEmptyString',
              output: `
function asserts1(x: string | number | undefined): asserts x {}
function asserts2(x: string | number | undefined): asserts x {}

const maybeString = Math.random() ? 'string'.slice() : undefined;

const someAssert: typeof asserts1 | typeof asserts2 =
  Math.random() > 0.5 ? asserts1 : asserts2;

someAssert(maybeString ?? "");
      `,
              desc: 'Explicitly treat nullish value the same as an empty string (`value ?? ""`)',
            },
            {
              messageId: 'conditionFixCastBoolean',
              output: `
function asserts1(x: string | number | undefined): asserts x {}
function asserts2(x: string | number | undefined): asserts x {}

const maybeString = Math.random() ? 'string'.slice() : undefined;

const someAssert: typeof asserts1 | typeof asserts2 =
  Math.random() > 0.5 ? asserts1 : asserts2;

someAssert(Boolean(maybeString));
      `,
              desc: 'Explicitly convert value to a boolean (`Boolean(value)`)',
            },
          ],
          message:
            'Unexpected nullable string value in conditional. Please handle the nullish/empty cases explicitly.',
          line: 10,
          column: 12,
          endLine: 10,
          endColumn: 23,
        },
      ],
    },
    {
      code: `
function assert(this: object, a: number, b: unknown): asserts b;
function assert(a: bigint, b: unknown): asserts b;
function assert(this: object, a: string, two: string): asserts two;
function assert(
  this: object,
  a: string,
  assertee: string,
  c: bigint,
  d: object,
): asserts assertee;

function assert(...args: any[]) {
  throw new Error('lol');
}

declare const nullableString: string | null;
assert(3 as any, nullableString);
      `,
      errors: [
        {
          column: 18,
          line: 18,
          messageId: 'conditionErrorNullableString',
          suggestions: [
            {
              messageId: 'conditionFixCompareNullish',
              output: `
function assert(this: object, a: number, b: unknown): asserts b;
function assert(a: bigint, b: unknown): asserts b;
function assert(this: object, a: string, two: string): asserts two;
function assert(
  this: object,
  a: string,
  assertee: string,
  c: bigint,
  d: object,
): asserts assertee;

function assert(...args: any[]) {
  throw new Error('lol');
}

declare const nullableString: string | null;
assert(3 as any, nullableString != null);
      `,
              desc: 'Change condition to check for null/undefined (`value != null`)',
            },
            {
              messageId: 'conditionFixDefaultEmptyString',
              output: `
function assert(this: object, a: number, b: unknown): asserts b;
function assert(a: bigint, b: unknown): asserts b;
function assert(this: object, a: string, two: string): asserts two;
function assert(
  this: object,
  a: string,
  assertee: string,
  c: bigint,
  d: object,
): asserts assertee;

function assert(...args: any[]) {
  throw new Error('lol');
}

declare const nullableString: string | null;
assert(3 as any, nullableString ?? "");
      `,
              desc: 'Explicitly treat nullish value the same as an empty string (`value ?? ""`)',
            },
            {
              messageId: 'conditionFixCastBoolean',
              output: `
function assert(this: object, a: number, b: unknown): asserts b;
function assert(a: bigint, b: unknown): asserts b;
function assert(this: object, a: string, two: string): asserts two;
function assert(
  this: object,
  a: string,
  assertee: string,
  c: bigint,
  d: object,
): asserts assertee;

function assert(...args: any[]) {
  throw new Error('lol');
}

declare const nullableString: string | null;
assert(3 as any, Boolean(nullableString));
      `,
              desc: 'Explicitly convert value to a boolean (`Boolean(value)`)',
            },
          ],
          message:
            'Unexpected nullable string value in conditional. Please handle the nullish/empty cases explicitly.',
          endLine: 18,
          endColumn: 32,
        },
      ],
    },
    {
      code: `
function assert(this: object, a: number, b: unknown): asserts b;
function assert(a: bigint, b: unknown): asserts b;
function assert(this: object, a: string, two: string): asserts two;
function assert(
  this: object,
  a: string,
  assertee: string,
  c: bigint,
  d: object,
): asserts assertee;
function assert(a: any, two: unknown, ...rest: any[]): asserts two;

function assert(...args: any[]) {
  throw new Error('lol');
}

declare const nullableString: string | null;
assert(3 as any, nullableString, 'more', 'args', 'afterwards');
      `,
      errors: [
        {
          column: 18,
          line: 19,
          messageId: 'conditionErrorNullableString',
          suggestions: [
            {
              messageId: 'conditionFixCompareNullish',
              output: `
function assert(this: object, a: number, b: unknown): asserts b;
function assert(a: bigint, b: unknown): asserts b;
function assert(this: object, a: string, two: string): asserts two;
function assert(
  this: object,
  a: string,
  assertee: string,
  c: bigint,
  d: object,
): asserts assertee;
function assert(a: any, two: unknown, ...rest: any[]): asserts two;

function assert(...args: any[]) {
  throw new Error('lol');
}

declare const nullableString: string | null;
assert(3 as any, nullableString != null, 'more', 'args', 'afterwards');
      `,
              desc: 'Change condition to check for null/undefined (`value != null`)',
            },
            {
              messageId: 'conditionFixDefaultEmptyString',
              output: `
function assert(this: object, a: number, b: unknown): asserts b;
function assert(a: bigint, b: unknown): asserts b;
function assert(this: object, a: string, two: string): asserts two;
function assert(
  this: object,
  a: string,
  assertee: string,
  c: bigint,
  d: object,
): asserts assertee;
function assert(a: any, two: unknown, ...rest: any[]): asserts two;

function assert(...args: any[]) {
  throw new Error('lol');
}

declare const nullableString: string | null;
assert(3 as any, nullableString ?? "", 'more', 'args', 'afterwards');
      `,
              desc: 'Explicitly treat nullish value the same as an empty string (`value ?? ""`)',
            },
            {
              messageId: 'conditionFixCastBoolean',
              output: `
function assert(this: object, a: number, b: unknown): asserts b;
function assert(a: bigint, b: unknown): asserts b;
function assert(this: object, a: string, two: string): asserts two;
function assert(
  this: object,
  a: string,
  assertee: string,
  c: bigint,
  d: object,
): asserts assertee;
function assert(a: any, two: unknown, ...rest: any[]): asserts two;

function assert(...args: any[]) {
  throw new Error('lol');
}

declare const nullableString: string | null;
assert(3 as any, Boolean(nullableString), 'more', 'args', 'afterwards');
      `,
              desc: 'Explicitly convert value to a boolean (`Boolean(value)`)',
            },
          ],
          message:
            'Unexpected nullable string value in conditional. Please handle the nullish/empty cases explicitly.',
          endLine: 19,
          endColumn: 32,
        },
      ],
    },
    {
      code: `
declare function assert(a: boolean, b: unknown): asserts b;
declare function assert({ a }: { a: boolean }, b: unknown): asserts b;
declare const nullableString: string | null;
declare const boo: boolean;
assert(boo, nullableString);
      `,
      errors: [
        {
          line: 6,
          messageId: 'conditionErrorNullableString',
          suggestions: [
            {
              messageId: 'conditionFixCompareNullish',
              output: `
declare function assert(a: boolean, b: unknown): asserts b;
declare function assert({ a }: { a: boolean }, b: unknown): asserts b;
declare const nullableString: string | null;
declare const boo: boolean;
assert(boo, nullableString != null);
      `,
              desc: 'Change condition to check for null/undefined (`value != null`)',
            },
            {
              messageId: 'conditionFixDefaultEmptyString',
              output: `
declare function assert(a: boolean, b: unknown): asserts b;
declare function assert({ a }: { a: boolean }, b: unknown): asserts b;
declare const nullableString: string | null;
declare const boo: boolean;
assert(boo, nullableString ?? "");
      `,
              desc: 'Explicitly treat nullish value the same as an empty string (`value ?? ""`)',
            },
            {
              messageId: 'conditionFixCastBoolean',
              output: `
declare function assert(a: boolean, b: unknown): asserts b;
declare function assert({ a }: { a: boolean }, b: unknown): asserts b;
declare const nullableString: string | null;
declare const boo: boolean;
assert(boo, Boolean(nullableString));
      `,
              desc: 'Explicitly convert value to a boolean (`Boolean(value)`)',
            },
          ],
          message:
            'Unexpected nullable string value in conditional. Please handle the nullish/empty cases explicitly.',
          column: 13,
          endLine: 6,
          endColumn: 27,
        },
      ],
    },
    {
      code: `
function assert(one: unknown): asserts one;
function assert(one: unknown, two: unknown): asserts two;
function assert(...args: unknown[]) {
  throw new Error('not implemented');
}
declare const nullableString: string | null;
assert(nullableString);
      `,
      errors: [
        {
          line: 8,
          messageId: 'conditionErrorNullableString',
          suggestions: [
            {
              messageId: 'conditionFixCompareNullish',
              output: `
function assert(one: unknown): asserts one;
function assert(one: unknown, two: unknown): asserts two;
function assert(...args: unknown[]) {
  throw new Error('not implemented');
}
declare const nullableString: string | null;
assert(nullableString != null);
      `,
              desc: 'Change condition to check for null/undefined (`value != null`)',
            },
            {
              messageId: 'conditionFixDefaultEmptyString',
              output: `
function assert(one: unknown): asserts one;
function assert(one: unknown, two: unknown): asserts two;
function assert(...args: unknown[]) {
  throw new Error('not implemented');
}
declare const nullableString: string | null;
assert(nullableString ?? "");
      `,
              desc: 'Explicitly treat nullish value the same as an empty string (`value ?? ""`)',
            },
            {
              messageId: 'conditionFixCastBoolean',
              output: `
function assert(one: unknown): asserts one;
function assert(one: unknown, two: unknown): asserts two;
function assert(...args: unknown[]) {
  throw new Error('not implemented');
}
declare const nullableString: string | null;
assert(Boolean(nullableString));
      `,
              desc: 'Explicitly convert value to a boolean (`Boolean(value)`)',
            },
          ],
          message:
            'Unexpected nullable string value in conditional. Please handle the nullish/empty cases explicitly.',
          column: 8,
          endLine: 8,
          endColumn: 22,
        },
      ],
    },
    {
      code: `
['one', 'two', ''].find(x => {
  return x;
});
      `,
      errors: [
        {
          column: 25,
          endColumn: 2,
          endLine: 4,
          line: 2,
          messageId: 'conditionErrorString',
          suggestions: [
            {
              messageId: 'explicitBooleanReturnType',
              output: `
['one', 'two', ''].find((x): boolean => {
  return x;
});
      `,
              desc: 'Add an explicit `boolean` return type annotation.',
            },
          ],
          message:
            'Unexpected string value in array predicate return type. An explicit empty string check is required.',
        },
      ],
      options: [
        {
          allowString: false,
        },
      ],
    },
    {
      code: `
['one', 'two', ''].find(x => {
  return;
});
      `,
      errors: [
        {
          column: 25,
          endColumn: 2,
          endLine: 4,
          line: 2,
          messageId: 'conditionErrorNullish',
          suggestions: [
            {
              messageId: 'explicitBooleanReturnType',
              output: `
['one', 'two', ''].find((x): boolean => {
  return;
});
      `,
              desc: 'Add an explicit `boolean` return type annotation.',
            },
          ],
          message:
            'Unexpected nullish value in conditional. The condition is always false.',
        },
      ],
    },
    {
      code: `
['one', 'two', ''].findLast(x => {
  return undefined;
});
      `,
      errors: [
        {
          column: 29,
          endColumn: 2,
          endLine: 4,
          line: 2,
          messageId: 'conditionErrorNullish',
          suggestions: [
            {
              messageId: 'explicitBooleanReturnType',
              output: `
['one', 'two', ''].findLast((x): boolean => {
  return undefined;
});
      `,
              desc: 'Add an explicit `boolean` return type annotation.',
            },
          ],
          message:
            'Unexpected nullish value in conditional. The condition is always false.',
        },
      ],
    },
    {
      code: `
['one', 'two', ''].find(x => {
  if (x) {
    return Math.random() > 0.5;
  }
});
      `,
      errors: [
        {
          column: 25,
          endColumn: 2,
          endLine: 6,
          line: 2,
          messageId: 'conditionErrorNullableBoolean',
          suggestions: [
            {
              messageId: 'explicitBooleanReturnType',
              output: `
['one', 'two', ''].find((x): boolean => {
  if (x) {
    return Math.random() > 0.5;
  }
});
      `,
              desc: 'Add an explicit `boolean` return type annotation.',
            },
          ],
          message:
            'Unexpected nullable boolean value in array predicate return type. Please handle the nullish case explicitly.',
        },
      ],
    },
    {
      code: `
const predicate = (x: string) => {
  if (x) {
    return Math.random() > 0.5;
  }
};

['one', 'two', ''].find(predicate);
      `,
      errors: [
        {
          column: 25,
          endColumn: 34,
          endLine: 8,
          line: 8,
          messageId: 'conditionErrorNullableBoolean',
          message:
            'Unexpected nullable boolean value in array predicate return type. Please handle the nullish case explicitly.',
        },
      ],
      options: [
        {
          allowNullableBoolean: false,
        },
      ],
    },
    {
      code: `
[1, null].every(async x => {
  return x != null;
});
      `,
      errors: [
        {
          column: 17,
          endColumn: 2,
          endLine: 4,
          line: 2,
          messageId: 'predicateCannotBeAsync',
          message:
            "Predicate function should not be 'async'; expected a boolean return type.",
        },
      ],
    },
    {
      code: `
const predicate = async x => {
  return x != null;
};

[1, null].every(predicate);
      `,
      errors: [
        {
          column: 17,
          endColumn: 26,
          endLine: 6,
          line: 6,
          messageId: 'conditionErrorObject',
          message:
            'Unexpected object value in array predicate return type. The condition is always true.',
        },
      ],
    },
    {
      code: `
[1, null].every((x): boolean | number => {
  return x != null;
});
      `,
      errors: [
        {
          column: 17,
          endColumn: 2,
          endLine: 4,
          line: 2,
          messageId: 'conditionErrorOther',
          message:
            'Unexpected value in conditional. A boolean expression is required.',
        },
      ],
    },
    {
      code: `
[1, null].every((x): boolean | undefined => {
  return x != null;
});
      `,
      errors: [
        {
          column: 17,
          endColumn: 2,
          endLine: 4,
          line: 2,
          messageId: 'conditionErrorNullableBoolean',
          message:
            'Unexpected nullable boolean value in array predicate return type. Please handle the nullish case explicitly.',
        },
      ],
    },
    {
      code: `
[1, null].every((x, i) => {});
      `,
      errors: [
        {
          column: 17,
          endColumn: 29,
          endLine: 2,
          line: 2,
          messageId: 'conditionErrorNullish',
          suggestions: [
            {
              messageId: 'explicitBooleanReturnType',
              output: `
[1, null].every((x, i): boolean => {});
      `,
              desc: 'Add an explicit `boolean` return type annotation.',
            },
          ],
          message:
            'Unexpected nullish value in conditional. The condition is always false.',
        },
      ],
    },
    {
      code: `
[() => {}, null].every((x: () => void) => {});
      `,
      errors: [
        {
          column: 24,
          endColumn: 45,
          endLine: 2,
          line: 2,
          messageId: 'conditionErrorNullish',
          suggestions: [
            {
              messageId: 'explicitBooleanReturnType',
              output: `
[() => {}, null].every((x: () => void): boolean => {});
      `,
              desc: 'Add an explicit `boolean` return type annotation.',
            },
          ],
          message:
            'Unexpected nullish value in conditional. The condition is always false.',
        },
      ],
    },
    {
      code: `
[() => {}, null].every(function (x: () => void) {});
      `,
      errors: [
        {
          column: 24,
          endColumn: 51,
          endLine: 2,
          line: 2,
          messageId: 'conditionErrorNullish',
          suggestions: [
            {
              messageId: 'explicitBooleanReturnType',
              output: `
[() => {}, null].every(function (x: () => void): boolean {});
      `,
              desc: 'Add an explicit `boolean` return type annotation.',
            },
          ],
          message:
            'Unexpected nullish value in conditional. The condition is always false.',
        },
      ],
    },
    {
      code: `
[() => {}, null].every(() => {});
      `,
      errors: [
        {
          column: 24,
          endColumn: 32,
          endLine: 2,
          line: 2,
          messageId: 'conditionErrorNullish',
          suggestions: [
            {
              messageId: 'explicitBooleanReturnType',
              output: `
[() => {}, null].every((): boolean => {});
      `,
              desc: 'Add an explicit `boolean` return type annotation.',
            },
          ],
          message:
            'Unexpected nullish value in conditional. The condition is always false.',
        },
      ],
    },
    {
      code: `
declare function f(x: number): string;
declare function f(x: string | null): boolean;

[35].filter(f);
      `,
      errors: [
        {
          column: 13,
          endColumn: 14,
          endLine: 5,
          line: 5,
          messageId: 'conditionErrorOther',
          message:
            'Unexpected value in conditional. A boolean expression is required.',
        },
      ],
    },
    {
      code: `
declare function f(x: number): string;
declare function f(x: number | boolean): boolean;
declare function f(x: string | null): boolean;

[35].filter(f);
      `,
      errors: [
        {
          column: 13,
          endColumn: 14,
          endLine: 6,
          line: 6,
          messageId: 'conditionErrorOther',
          message:
            'Unexpected value in conditional. A boolean expression is required.',
        },
      ],
    },
    {
      code: `
declare function foo<T>(x: number): T;
[1, null].every(foo);
      `,
      errors: [
        {
          column: 17,
          endColumn: 20,
          endLine: 3,
          line: 3,
          messageId: 'conditionErrorAny',
          message:
            'Unexpected any value in array predicate return type. An explicit comparison or type conversion is required.',
        },
      ],
    },
    {
      code: `
function foo<T extends number>(x: number): T {}
[1, null].every(foo);
      `,
      errors: [
        {
          column: 17,
          endColumn: 20,
          endLine: 3,
          line: 3,
          messageId: 'conditionErrorNumber',
          message:
            'Unexpected number value in array predicate return type. An explicit zero/NaN check is required.',
        },
      ],
      options: [
        {
          allowNumber: false,
        },
      ],
    },
    {
      code: `
declare const nullOrString: string | null;
['one', null].filter(x => nullOrString);
      `,
      errors: [
        {
          column: 22,
          endColumn: 39,
          endLine: 3,
          line: 3,
          messageId: 'conditionErrorNullableString',
          suggestions: [
            {
              messageId: 'conditionFixCompareNullish',
              output: `
declare const nullOrString: string | null;
['one', null].filter(x => nullOrString != null);
      `,
              desc: 'Change condition to check for null/undefined (`value != null`)',
            },
            {
              messageId: 'conditionFixDefaultEmptyString',
              output: `
declare const nullOrString: string | null;
['one', null].filter(x => nullOrString ?? "");
      `,
              desc: 'Explicitly treat nullish value the same as an empty string (`value ?? ""`)',
            },
            {
              messageId: 'conditionFixCastBoolean',
              output: `
declare const nullOrString: string | null;
['one', null].filter(x => Boolean(nullOrString));
      `,
              desc: 'Explicitly convert value to a boolean (`Boolean(value)`)',
            },
            {
              messageId: 'explicitBooleanReturnType',
              output: `
declare const nullOrString: string | null;
['one', null].filter((x): boolean => nullOrString);
      `,
              desc: 'Add an explicit `boolean` return type annotation.',
            },
          ],
          message:
            'Unexpected nullable string value in array predicate return type. Please handle the nullish/empty cases explicitly.',
        },
      ],
    },
    {
      code: `
declare const nullOrString: string | null;
['one', null].filter(x => !nullOrString);
      `,
      errors: [
        {
          column: 28,
          endColumn: 40,
          endLine: 3,
          line: 3,
          messageId: 'conditionErrorNullableString',
          suggestions: [
            {
              messageId: 'conditionFixCompareNullish',
              output: `
declare const nullOrString: string | null;
['one', null].filter(x => nullOrString == null);
      `,
              desc: 'Change condition to check for null/undefined (`value != null`)',
            },
            {
              messageId: 'conditionFixDefaultEmptyString',
              output: `
declare const nullOrString: string | null;
['one', null].filter(x => !(nullOrString ?? ""));
      `,
              desc: 'Explicitly treat nullish value the same as an empty string (`value ?? ""`)',
            },
            {
              messageId: 'conditionFixCastBoolean',
              output: `
declare const nullOrString: string | null;
['one', null].filter(x => !Boolean(nullOrString));
      `,
              desc: 'Explicitly convert value to a boolean (`Boolean(value)`)',
            },
          ],
          message:
            'Unexpected nullable string value in conditional. Please handle the nullish/empty cases explicitly.',
        },
      ],
    },
    {
      code: `
declare const anyValue: any;
['one', null].filter(x => anyValue);
      `,
      errors: [
        {
          column: 22,
          endColumn: 35,
          endLine: 3,
          line: 3,
          messageId: 'conditionErrorAny',
          suggestions: [
            {
              messageId: 'conditionFixCastBoolean',
              output: `
declare const anyValue: any;
['one', null].filter(x => Boolean(anyValue));
      `,
              desc: 'Explicitly convert value to a boolean (`Boolean(value)`)',
            },
            {
              messageId: 'explicitBooleanReturnType',
              output: `
declare const anyValue: any;
['one', null].filter((x): boolean => anyValue);
      `,
              desc: 'Add an explicit `boolean` return type annotation.',
            },
          ],
          message:
            'Unexpected any value in array predicate return type. An explicit comparison or type conversion is required.',
        },
      ],
    },
    {
      code: `
declare const nullOrBoolean: boolean | null;
[true, null].filter(x => nullOrBoolean);
      `,
      errors: [
        {
          column: 21,
          endColumn: 39,
          endLine: 3,
          line: 3,
          messageId: 'conditionErrorNullableBoolean',
          suggestions: [
            {
              messageId: 'conditionFixDefaultFalse',
              output: `
declare const nullOrBoolean: boolean | null;
[true, null].filter(x => nullOrBoolean ?? false);
      `,
              desc: 'Explicitly treat nullish value the same as false (`value ?? false`)',
            },
            {
              messageId: 'conditionFixCompareTrue',
              output: `
declare const nullOrBoolean: boolean | null;
[true, null].filter(x => nullOrBoolean === true);
      `,
              desc: 'Change condition to check if true (`value === true`)',
            },
            {
              messageId: 'explicitBooleanReturnType',
              output: `
declare const nullOrBoolean: boolean | null;
[true, null].filter((x): boolean => nullOrBoolean);
      `,
              desc: 'Add an explicit `boolean` return type annotation.',
            },
          ],
          message:
            'Unexpected nullable boolean value in array predicate return type. Please handle the nullish case explicitly.',
        },
      ],
    },
    {
      code: `
enum ExampleEnum {
  This = 0,
  That = 1,
}
const theEnum = Math.random() < 0.3 ? ExampleEnum.This : null;
[0, 1].filter(x => theEnum);
      `,
      errors: [
        {
          column: 15,
          endColumn: 27,
          endLine: 7,
          line: 7,
          messageId: 'conditionErrorNullableEnum',
          suggestions: [
            {
              messageId: 'conditionFixCompareNullish',
              output: `
enum ExampleEnum {
  This = 0,
  That = 1,
}
const theEnum = Math.random() < 0.3 ? ExampleEnum.This : null;
[0, 1].filter(x => theEnum != null);
      `,
              desc: 'Change condition to check for null/undefined (`value != null`)',
            },
            {
              messageId: 'explicitBooleanReturnType',
              output: `
enum ExampleEnum {
  This = 0,
  That = 1,
}
const theEnum = Math.random() < 0.3 ? ExampleEnum.This : null;
[0, 1].filter((x): boolean => theEnum);
      `,
              desc: 'Add an explicit `boolean` return type annotation.',
            },
          ],
          message:
            'Unexpected nullable enum value in array predicate return type. Please handle the nullish/zero/NaN cases explicitly.',
        },
      ],
    },
    {
      code: `
declare const nullOrNumber: number | null;
[0, null].filter(x => nullOrNumber);
      `,
      errors: [
        {
          column: 18,
          endColumn: 35,
          endLine: 3,
          line: 3,
          messageId: 'conditionErrorNullableNumber',
          suggestions: [
            {
              messageId: 'conditionFixCompareNullish',
              output: `
declare const nullOrNumber: number | null;
[0, null].filter(x => nullOrNumber != null);
      `,
              desc: 'Change condition to check for null/undefined (`value != null`)',
            },
            {
              messageId: 'conditionFixDefaultZero',
              output: `
declare const nullOrNumber: number | null;
[0, null].filter(x => nullOrNumber ?? 0);
      `,
              desc: 'Explicitly treat nullish value the same as 0 (`value ?? 0`)',
            },
            {
              messageId: 'conditionFixCastBoolean',
              output: `
declare const nullOrNumber: number | null;
[0, null].filter(x => Boolean(nullOrNumber));
      `,
              desc: 'Explicitly convert value to a boolean (`Boolean(value)`)',
            },
            {
              messageId: 'explicitBooleanReturnType',
              output: `
declare const nullOrNumber: number | null;
[0, null].filter((x): boolean => nullOrNumber);
      `,
              desc: 'Add an explicit `boolean` return type annotation.',
            },
          ],
          message:
            'Unexpected nullable number value in array predicate return type. Please handle the nullish/zero/NaN cases explicitly.',
        },
      ],
    },
    {
      code: `
const objectValue: object = {};
[{ a: 0 }, {}].filter(x => objectValue);
      `,
      errors: [
        {
          column: 23,
          endColumn: 39,
          endLine: 3,
          line: 3,
          messageId: 'conditionErrorObject',
          suggestions: [
            {
              messageId: 'explicitBooleanReturnType',
              output: `
const objectValue: object = {};
[{ a: 0 }, {}].filter((x): boolean => objectValue);
      `,
              desc: 'Add an explicit `boolean` return type annotation.',
            },
          ],
          message:
            'Unexpected object value in array predicate return type. The condition is always true.',
        },
      ],
    },
    {
      code: `
const objectValue: object = {};
[{ a: 0 }, {}].filter(x => {
  return objectValue;
});
      `,
      errors: [
        {
          column: 23,
          endColumn: 2,
          endLine: 5,
          line: 3,
          messageId: 'conditionErrorObject',
          suggestions: [
            {
              messageId: 'explicitBooleanReturnType',
              output: `
const objectValue: object = {};
[{ a: 0 }, {}].filter((x): boolean => {
  return objectValue;
});
      `,
              desc: 'Add an explicit `boolean` return type annotation.',
            },
          ],
          message:
            'Unexpected object value in array predicate return type. The condition is always true.',
        },
      ],
    },
    {
      code: `
declare const nullOrObject: object | null;
[{ a: 0 }, null].filter(x => nullOrObject);
      `,
      errors: [
        {
          column: 25,
          endColumn: 42,
          endLine: 3,
          line: 3,
          messageId: 'conditionErrorNullableObject',
          suggestions: [
            {
              messageId: 'conditionFixCompareNullish',
              output: `
declare const nullOrObject: object | null;
[{ a: 0 }, null].filter(x => nullOrObject != null);
      `,
              desc: 'Change condition to check for null/undefined (`value != null`)',
            },
            {
              messageId: 'explicitBooleanReturnType',
              output: `
declare const nullOrObject: object | null;
[{ a: 0 }, null].filter((x): boolean => nullOrObject);
      `,
              desc: 'Add an explicit `boolean` return type annotation.',
            },
          ],
          message:
            'Unexpected nullable object value in array predicate return type. An explicit null check is required.',
        },
      ],
      options: [
        {
          allowNullableObject: false,
        },
      ],
    },
    {
      code: `
const numbers: number[] = [1];
[1, 2].filter(x => numbers.length);
      `,
      errors: [
        {
          column: 15,
          endColumn: 34,
          endLine: 3,
          line: 3,
          messageId: 'conditionErrorNumber',
          suggestions: [
            {
              messageId: 'conditionFixCompareArrayLengthNonzero',
              output: `
const numbers: number[] = [1];
[1, 2].filter(x => numbers.length > 0);
      `,
              desc: "Change condition to check array's length (`value.length > 0`)",
            },
            {
              messageId: 'explicitBooleanReturnType',
              output: `
const numbers: number[] = [1];
[1, 2].filter((x): boolean => numbers.length);
      `,
              desc: 'Add an explicit `boolean` return type annotation.',
            },
          ],
          message:
            'Unexpected number value in array predicate return type. An explicit zero/NaN check is required.',
        },
      ],
      options: [
        {
          allowNumber: false,
        },
      ],
    },
    {
      code: `
const numberValue: number = 1;
[1, 2].filter(x => numberValue);
      `,
      errors: [
        {
          column: 15,
          endColumn: 31,
          endLine: 3,
          line: 3,
          messageId: 'conditionErrorNumber',
          suggestions: [
            {
              messageId: 'conditionFixCompareZero',
              output: `
const numberValue: number = 1;
[1, 2].filter(x => numberValue !== 0);
      `,
              desc: 'Change condition to check for 0 (`value !== 0`)',
            },
            {
              messageId: 'conditionFixCompareNaN',
              output: `
const numberValue: number = 1;
[1, 2].filter(x => !Number.isNaN(numberValue));
      `,
              desc: 'Change condition to check for NaN (`!Number.isNaN(value)`)',
            },
            {
              messageId: 'conditionFixCastBoolean',
              output: `
const numberValue: number = 1;
[1, 2].filter(x => Boolean(numberValue));
      `,
              desc: 'Explicitly convert value to a boolean (`Boolean(value)`)',
            },
            {
              messageId: 'explicitBooleanReturnType',
              output: `
const numberValue: number = 1;
[1, 2].filter((x): boolean => numberValue);
      `,
              desc: 'Add an explicit `boolean` return type annotation.',
            },
          ],
          message:
            'Unexpected number value in array predicate return type. An explicit zero/NaN check is required.',
        },
      ],
      options: [
        {
          allowNumber: false,
        },
      ],
    },
    {
      code: `
const stringValue: string = 'hoge';
['hoge', 'foo'].filter(x => stringValue);
      `,
      errors: [
        {
          column: 24,
          endColumn: 40,
          endLine: 3,
          line: 3,
          messageId: 'conditionErrorString',
          suggestions: [
            {
              messageId: 'conditionFixCompareStringLength',
              output: `
const stringValue: string = 'hoge';
['hoge', 'foo'].filter(x => stringValue.length > 0);
      `,
              desc: "Change condition to check string's length (`value.length !== 0`)",
            },
            {
              messageId: 'conditionFixCompareEmptyString',
              output: `
const stringValue: string = 'hoge';
['hoge', 'foo'].filter(x => stringValue !== "");
      `,
              desc: 'Change condition to check for empty string (`value !== ""`)',
            },
            {
              messageId: 'conditionFixCastBoolean',
              output: `
const stringValue: string = 'hoge';
['hoge', 'foo'].filter(x => Boolean(stringValue));
      `,
              desc: 'Explicitly convert value to a boolean (`Boolean(value)`)',
            },
            {
              messageId: 'explicitBooleanReturnType',
              output: `
const stringValue: string = 'hoge';
['hoge', 'foo'].filter((x): boolean => stringValue);
      `,
              desc: 'Add an explicit `boolean` return type annotation.',
            },
          ],
          message:
            'Unexpected string value in array predicate return type. An explicit empty string check is required.',
        },
      ],
      options: [
        {
          allowString: false,
        },
      ],
    },
  ],
});
