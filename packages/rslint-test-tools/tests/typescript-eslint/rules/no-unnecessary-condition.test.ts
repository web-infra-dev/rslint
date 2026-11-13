import { RuleTester, type TestCase } from '@rslint/rule-tester';
import { config } from '../config';

const necessaryConditionTest = (condition: string): string => `
declare const b1: ${condition};
declare const b2: boolean;
const t1 = b1 && b2;
`;

const unnecessaryConditionTest = (
  condition: string,
  messageId: 'alwaysTruthy' | 'alwaysFalsy' | 'never',
): TestCase => ({
  code: necessaryConditionTest(condition),
  errors: [
    {
      messageId,
      message: {
        alwaysTruthy: 'Unnecessary conditional, value is always truthy.',
        alwaysFalsy: 'Unnecessary conditional, value is always falsy.',
        never: 'Unnecessary conditional, value is `never`.',
      }[messageId],
      line: 4,
      column: 12,
      endLine: 4,
      endColumn: 14,
    },
  ],
});

new RuleTester(config).run('no-unnecessary-condition', {
  valid: [
    `
declare const b1: boolean;
declare const b2: boolean;
const t1 = b1 && b2;
const t2 = b1 || b2;
if (b1 && b2) {
}
while (b1 && b2) {}
for (let i = 0; b1 && b2; i++) {
  break;
}
const t1 = b1 && b2 ? 'yes' : 'no';
if (b1 && b2) {
}
while (b1 && b2) {}
for (let i = 0; b1 && b2; i++) {
  break;
}
const t1 = b1 && b2 ? 'yes' : 'no';
for (;;) {}
switch (b1) {
  case true:
  default:
}
    `,
    `
declare function foo(): number | void;
const result1 = foo() === undefined;
const result2 = foo() == null;
    `,
    `
declare function foo(): number | void;
foo() ?? 1;
    `,
    `
declare const bigInt: 0n | 1n;
if (bigInt) {
}
    `,
    // Truthy literal and falsy literal
    necessaryConditionTest('false | 5'),
    // boolean and truthy literal
    necessaryConditionTest('boolean | "foo"'),
    // boolean and falsy literal
    necessaryConditionTest('0 | boolean'),
    // boolean and always-truthy type
    necessaryConditionTest('boolean | object'),
    // always truthy type and falsy literal
    necessaryConditionTest('false | object'),
    // always falsy type and always truthy type
    necessaryConditionTest('null | object'),
    necessaryConditionTest('undefined | true'),
    necessaryConditionTest('void | true'),
    // "branded" type
    necessaryConditionTest('string & {}'),
    necessaryConditionTest('string & { __brand: string }'),
    necessaryConditionTest('number & { __brand: string }'),
    necessaryConditionTest('boolean & { __brand: string }'),
    necessaryConditionTest('bigint & { __brand: string }'),
    necessaryConditionTest('string & {} & { __brand: string }'),
    necessaryConditionTest(
      'string & { __brandA: string } & { __brandB: string }',
    ),
    necessaryConditionTest('string & { __brand: string } | number'),
    necessaryConditionTest('(string | number) & { __brand: string }'),
    necessaryConditionTest('string & ({ __brand: string } | number)'),
    necessaryConditionTest('("" | "foo") & { __brand: string }'),
    necessaryConditionTest(
      '(string & { __brandA: string }) | (number & { __brandB: string })',
    ),
    necessaryConditionTest(
      '((string & { __brandA: string }) | (number & { __brandB: string }) & ("" | "foo"))',
    ),
    necessaryConditionTest(
      '{ __brandA: string} & (({ __brandB: string } & string) | ({ __brandC: string } & number))',
    ),
    necessaryConditionTest(
      '(string | number) & ("foo" | 123 | { __brandA: string })',
    ),
    necessaryConditionTest('string & string'),
    // any
    necessaryConditionTest('any'),
    // unknown
    necessaryConditionTest('unknown'),
    // Generic type params
    `
function test<T extends string>(t: T) {
  return t ? 'yes' : 'no';
}
    `,
    `
// Naked type param
function test<T>(t: T) {
  return t ? 'yes' : 'no';
}
    `,
    `
// Naked type param in union
function test<T>(t: T | []) {
  return t ? 'yes' : 'no';
}
    `,
    `
function test<T>(arg: T, key: keyof T) {
  if (arg[key]?.toString()) {
  }
}
    `,
    `
function test<T>(arg: T, key: keyof T) {
  if (arg?.toString()) {
  }
}
    `,
    `
function test<T>(arg: T | { value: string }) {
  if (arg?.value) {
  }
}
    `,
    // Boolean expressions
    `
function test(a: string) {
  const t1 = a === 'a';
  const t2 = 'a' === a;
}
    `,
    `
function test(a?: string) {
  const t1 = a === undefined;
  const t2 = undefined === a;
  const t1 = a !== undefined;
  const t2 = undefined !== a;
}
    `,
    `
function test(a: null | string) {
  const t1 = a === null;
  const t2 = null === a;
  const t1 = a !== null;
  const t2 = null !== a;
}
    `,
    `
function test(a?: null | string) {
  const t1 = a == null;
  const t2 = null == a;
  const t3 = a != null;
  const t4 = null != a;
  const t5 = a == undefined;
  const t6 = undefined == a;
  const t7 = a != undefined;
  const t8 = undefined != a;
}
    `,
    `
function test(a?: string) {
  const t1 = a == null;
  const t2 = null == a;
  const t3 = a != null;
  const t4 = null != a;
  const t5 = a == undefined;
  const t6 = undefined == a;
  const t7 = a != undefined;
  const t8 = undefined != a;
}
    `,
    `
function test(a: null | string) {
  const t1 = a == null;
  const t2 = null == a;
  const t3 = a != null;
  const t4 = null != a;
  const t5 = a == undefined;
  const t6 = undefined == a;
  const t7 = a != undefined;
  const t8 = undefined != a;
}
    `,
    `
function test(a: any) {
  const t1 = a == null;
  const t2 = null == a;
  const t3 = a != null;
  const t4 = null != a;
  const t5 = a == undefined;
  const t6 = undefined == a;
  const t7 = a != undefined;
  const t8 = undefined != a;
  const t9 = a === null;
  const t10 = null === a;
  const t11 = a !== null;
  const t12 = null !== a;
  const t13 = a === undefined;
  const t14 = undefined === a;
  const t15 = a !== undefined;
  const t16 = undefined !== a;
}
    `,
    `
function test(a: unknown) {
  const t1 = a == null;
  const t2 = null == a;
  const t3 = a != null;
  const t4 = null != a;
  const t5 = a == undefined;
  const t6 = undefined == a;
  const t7 = a != undefined;
  const t8 = undefined != a;
  const t9 = a === null;
  const t10 = null === a;
  const t11 = a !== null;
  const t12 = null !== a;
  const t13 = a === undefined;
  const t14 = undefined === a;
  const t15 = a !== undefined;
  const t16 = undefined !== a;
}
    `,
    `
function test<T>(a: T) {
  const t1 = a == null;
  const t2 = null == a;
  const t3 = a != null;
  const t4 = null != a;
  const t5 = a == undefined;
  const t6 = undefined == a;
  const t7 = a != undefined;
  const t8 = undefined != a;
  const t9 = a === null;
  const t10 = null === a;
  const t11 = a !== null;
  const t12 = null !== a;
  const t13 = a === undefined;
  const t14 = undefined === a;
  const t15 = a !== undefined;
  const t16 = undefined !== a;
}
    `,
    `
function foo<T extends object>(arg: T, key: keyof T): void {
  arg[key] == null;
}
    `,
    // Predicate functions
    `
// with literal arrow function
[0, 1, 2].filter(x => x);

// filter with named function
function length(x: string) {
  return x.length;
}
['a', 'b', ''].filter(length);

// with non-literal array
function nonEmptyStrings(x: string[]) {
  return x.filter(length);
}

// filter-like predicate
function count(
  list: string[],
  predicate: (value: string, index: number, array: string[]) => unknown,
) {
  return list.filter(predicate).length;
}
    `,
    `
declare const test: <T>() => T;

[1, null].filter(test);
    `,
    `
declare const test: <T extends boolean>() => T;

[1, null].filter(test);
    `,
    `
[1, null].filter(1 as any);
    `,
    `
[1, null].filter(1 as never);
    `,
    // Ignores non-array methods of the same name
    `
const notArray = {
  filter: (func: () => boolean) => func(),
  find: (func: () => boolean) => func(),
};
notArray.filter(() => true);
notArray.find(() => true);
    `,
    // Nullish coalescing operator
    `
function test(a: string | null) {
  return a ?? 'default';
}
    `,
    `
function test(a: string | undefined) {
  return a ?? 'default';
}
    `,
    `
function test(a: string | null | undefined) {
  return a ?? 'default';
}
    `,
    `
function test(a: unknown) {
  return a ?? 'default';
}
    `,
    `
function test<T>(a: T) {
  return a ?? 'default';
}
    `,
    `
function test<T extends string | null>(a: T) {
  return a ?? 'default';
}
    `,
    `
function foo<T extends object>(arg: T, key: keyof T): void {
  arg[key] ?? 'default';
}
    `,
    // Indexing cases
    `
declare const arr: object[];
if (arr[42]) {
} // looks unnecessary from the types, but isn't

const tuple = [{}] as [object];
declare const n: number;
if (tuple[n]) {
}
    `,
    // Optional-chaining indexing
    `
declare const arr: Array<{ value: string } & (() => void)>;
if (arr[42]?.value) {
}
arr[41]?.();

// An array access can "infect" deeper into the chain
declare const arr2: Array<{ x: { y: { z: object } } }>;
arr2[42]?.x?.y?.z;

const tuple = ['foo'] as const;
declare const n: number;
tuple[n]?.toUpperCase();
    `,
    {
      code: `
declare const arr: Array<{ value: string } & (() => void)>;
if (arr[42]?.value) {
}
arr[41]?.();
      `,
      languageOptions: {
        parserOptions: {
          project: './tsconfig.noUncheckedIndexedAccess.json',
          projectService: false,
        },
      },
    },
    `
if (arr?.[42]) {
}
    `,
    `
type ItemA = { bar: string; baz: string };
type ItemB = { bar: string; qux: string };
declare const foo: ItemA[] | ItemB[];
foo[0]?.bar;
    `,
    `
type TupleA = [string, number];
type TupleB = [string, number];

declare const foo: TupleA | TupleB;
declare const index: number;
foo[index]?.toString();
    `,
    {
      code: `
type TupleA = [string, number];
type TupleB = [string, number];

declare const foo: TupleA | TupleB;
declare const index: number;
foo[index]?.toString();
      `,
      languageOptions: {
        parserOptions: {
          project: './tsconfig.noUncheckedIndexedAccess.json',
          projectService: false,
        },
      },
    },
    `
declare const returnsArr: undefined | (() => string[]);
if (returnsArr?.()[42]) {
}
returnsArr?.()[42]?.toUpperCase();
    `,
    // nullish + array index
    `
declare const arr: string[][];
arr[x] ?? [];
    `,
    // nullish + optional array index
    `
declare const arr: { foo: number }[];
const bar = arr[42]?.foo ?? 0;
    `,
    // Doesn't check the right-hand side of a logical expression
    //  in a non-conditional context
    `
declare const b1: boolean;
declare const b2: true;
const x = b1 && b2;
      `,
    {
      code: `
while (true) {}
      `,
      options: [
        {
          allowConstantLoopConditions: true,
        },
      ],
    },
    {
      code: `
for (; true; ) {}
      `,
      options: [
        {
          allowConstantLoopConditions: true,
        },
      ],
    },
    {
      code: `
do {} while (true);
      `,
      options: [
        {
          allowConstantLoopConditions: true,
        },
      ],
    },
    {
      code: `
while (true) {}
      `,
      options: [
        {
          allowConstantLoopConditions: 'always',
        },
      ],
    },
    {
      code: `
for (; true; ) {}
      `,
      options: [
        {
          allowConstantLoopConditions: 'always',
        },
      ],
    },
    {
      code: `
for (; true; ) {}
      `,
      options: [
        {
          allowConstantLoopConditions: 'only-allowed-literals',
        },
      ],
    },
    {
      code: `
for (; 0; ) {}
      `,
      options: [
        {
          allowConstantLoopConditions: 'only-allowed-literals',
        },
      ],
    },
    {
      code: `
do {} while (0);
      `,
      options: [
        {
          allowConstantLoopConditions: 'only-allowed-literals',
        },
      ],
    },
    {
      code: `
do {} while (true);
      `,
      options: [
        {
          allowConstantLoopConditions: 'always',
        },
      ],
    },
    {
      code: `
while (true) {}
      `,
      options: [
        {
          allowConstantLoopConditions: 'only-allowed-literals',
        },
      ],
    },
    {
      code: `
while (1) {}
      `,
      options: [
        {
          allowConstantLoopConditions: 'only-allowed-literals',
        },
      ],
    },
    {
      code: `
while (false) {}
      `,
      options: [
        {
          allowConstantLoopConditions: 'only-allowed-literals',
        },
      ],
    },
    {
      code: `
while (0) {}
      `,
      options: [
        {
          allowConstantLoopConditions: 'only-allowed-literals',
        },
      ],
    },
    `
let variable = 'abc' as string | void;
variable?.[0];
    `,
    `
let foo: undefined | { bar: true };
foo?.bar;
    `,
    `
let foo: null | { bar: true };
foo?.bar;
    `,
    `
let foo: undefined;
foo?.bar;
    `,
    `
let foo: undefined;
foo?.bar.baz;
    `,
    `
let foo: null;
foo?.bar;
    `,
    `
let anyValue: any;
anyValue?.foo;
    `,
    `
let unknownValue: unknown;
unknownValue?.foo;
    `,
    `
let foo: undefined | (() => {});
foo?.();
    `,
    `
let foo: null | (() => {});
foo?.();
    `,
    `
let foo: undefined;
foo?.();
    `,
    `
let foo: undefined;
foo?.().bar;
    `,
    `
let foo: null;
foo?.();
    `,
    `
let anyValue: any;
anyValue?.();
    `,
    `
let unknownValue: unknown;
unknownValue?.();
    `,
    'const foo = [1, 2, 3][0];',
    `
declare const foo: { bar?: { baz: { c: string } } } | null;
foo?.bar?.baz;
    `,
    `
foo?.bar?.baz?.qux;
    `,
    `
declare const foo: { bar: { baz: string } };
foo.bar.qux?.();
    `,
    `
type Foo = { baz: number } | null;
type Bar = { baz: null | string | { qux: string } };
declare const foo: { fooOrBar: Foo | Bar } | null;
foo?.fooOrBar?.baz?.qux;
    `,
    `
type Foo = { [key: string]: string } | null;
declare const foo: Foo;

const key = '1';
foo?.[key]?.trim();
    `,
    `
type Foo = { [key: string]: string; foo: 'foo'; bar: 'bar' } | null;
type Key = 'bar' | 'foo';
declare const foo: Foo;
declare const key: Key;

foo?.[key].trim();
    `,
    `
interface Outer {
  inner?: {
    [key: string]: string | undefined;
  };
}

function Foo(outer: Outer, key: string): number | undefined {
  return outer.inner?.[key]?.charCodeAt(0);
}
    `,
    `
interface Outer {
  inner?: {
    [key: string]: string | undefined;
    bar: 'bar';
  };
}
type Foo = 'foo';

function Foo(outer: Outer, key: Foo): number | undefined {
  return outer.inner?.[key]?.charCodeAt(0);
}
    `,
    `
type Foo = { [key: string]: string; foo: 'foo'; bar: 'bar' } | null;
type Key = 'bar' | 'foo' | 'baz';
declare const foo: Foo;
declare const key: Key;

foo?.[key]?.trim();
    `,
    // https://github.com/typescript-eslint/typescript-eslint/issues/7700
    `
type BrandedKey = string & { __brand: string };
type Foo = { [key: BrandedKey]: string } | null;
declare const foo: Foo;
const key = '1' as BrandedKey;
foo?.[key]?.trim();
    `,
    `
type BrandedKey<S extends string> = S & { __brand: string };
type Foo = { [key: string]: string; foo: 'foo'; bar: 'bar' } | null;
type Key = BrandedKey<'bar'> | BrandedKey<'foo'>;
declare const foo: Foo;
declare const key: Key;
foo?.[key].trim();
    `,
    `
type BrandedKey = string & { __brand: string };
interface Outer {
  inner?: {
    [key: BrandedKey]: string | undefined;
  };
}
function Foo(outer: Outer, key: BrandedKey): number | undefined {
  return outer.inner?.[key]?.charCodeAt(0);
}
    `,
    `
interface Outer {
  inner?: {
    [key: string & { __brand: string }]: string | undefined;
    bar: 'bar';
  };
}
type Foo = 'foo' & { __brand: string };
function Foo(outer: Outer, key: Foo): number | undefined {
  return outer.inner?.[key]?.charCodeAt(0);
}
    `,
    `
type BrandedKey<S extends string> = S & { __brand: string };
type Foo = { [key: string]: string; foo: 'foo'; bar: 'bar' } | null;
type Key = BrandedKey<'bar'> | BrandedKey<'foo'> | BrandedKey<'baz'>;
declare const foo: Foo;
declare const key: Key;
foo?.[key]?.trim();
    `,
    {
      code: `
type BrandedKey = string & { __brand: string };
type Foo = { [key: BrandedKey]: string } | null;
declare const foo: Foo;
const key = '1' as BrandedKey;
foo?.[key]?.trim();
      `,
      languageOptions: {
        parserOptions: {
          project: './tsconfig.noUncheckedIndexedAccess.json',
          projectService: false,
        },
      },
    },
    {
      code: `
type BrandedKey<S extends string> = S & { __brand: string };
type Foo = { [key: string]: string; foo: 'foo'; bar: 'bar' } | null;
type Key = BrandedKey<'bar'> | BrandedKey<'foo'>;
declare const foo: Foo;
declare const key: Key;
foo?.[key].trim();
      `,
      languageOptions: {
        parserOptions: {
          project: './tsconfig.noUncheckedIndexedAccess.json',
          projectService: false,
        },
      },
    },
    {
      code: `
type BrandedKey = string & { __brand: string };
interface Outer {
  inner?: {
    [key: BrandedKey]: string | undefined;
  };
}
function Foo(outer: Outer, key: BrandedKey): number | undefined {
  return outer.inner?.[key]?.charCodeAt(0);
}
      `,
      languageOptions: {
        parserOptions: {
          project: './tsconfig.noUncheckedIndexedAccess.json',
          projectService: false,
        },
      },
    },
    {
      code: `
interface Outer {
  inner?: {
    [key: string & { __brand: string }]: string | undefined;
    bar: 'bar';
  };
}
type Foo = 'foo' & { __brand: string };
function Foo(outer: Outer, key: Foo): number | undefined {
  return outer.inner?.[key]?.charCodeAt(0);
}
      `,
      languageOptions: {
        parserOptions: {
          project: './tsconfig.noUncheckedIndexedAccess.json',
          projectService: false,
        },
      },
    },
    {
      code: `
type BrandedKey<S extends string> = S & { __brand: string };
type Foo = { [key: string]: string; foo: 'foo'; bar: 'bar' } | null;
type Key = BrandedKey<'bar'> | BrandedKey<'foo'> | BrandedKey<'baz'>;
declare const foo: Foo;
declare const key: Key;
foo?.[key]?.trim();
      `,
      languageOptions: {
        parserOptions: {
          project: './tsconfig.noUncheckedIndexedAccess.json',
          projectService: false,
        },
      },
    },
    {
      code: `
type Foo = {
  key?: Record<string, { key: string }>;
};
declare const foo: Foo;
foo.key?.someKey?.key;
      `,
      languageOptions: {
        parserOptions: {
          project: './tsconfig.noUncheckedIndexedAccess.json',
          projectService: false,
        },
      },
    },
    {
      code: `
type Foo = {
  key?: {
    [key: string]: () => void;
  };
};
declare const foo: Foo;
foo.key?.value?.();
      `,
      languageOptions: {
        parserOptions: {
          project: './tsconfig.noUncheckedIndexedAccess.json',
          projectService: false,
        },
      },
    },
    {
      code: `
type A = {
  [name in Lowercase<string>]?: {
    [name in Lowercase<string>]: {
      a: 1;
    };
  };
};

declare const a: A;

a.a?.a?.a;
      `,
      languageOptions: {
        parserOptions: {
          project: './tsconfig.noUncheckedIndexedAccess.json',
          projectService: false,
        },
      },
    },
    `
let latencies: number[][] = [];

function recordData(): void {
  if (!latencies[0]) latencies[0] = [];
  latencies[0].push(4);
}

recordData();
    `,
    `
let latencies: number[][] = [];

function recordData(): void {
  if (latencies[0]) latencies[0] = [];
  latencies[0].push(4);
}

recordData();
    `,
    `
function test(testVal?: boolean) {
  if (testVal ?? true) {
    console.log('test');
  }
}
    `,
    `
declare const x: string[];
if (!x[0]) {
}
    `,
    // https://github.com/typescript-eslint/typescript-eslint/issues/2421
    `
const isEven = (val: number) => val % 2 === 0;
if (!isEven(1)) {
}
    `,
    `
declare const booleanTyped: boolean;
declare const unknownTyped: unknown;

if (!(booleanTyped || unknownTyped)) {
}
    `,
    {
      code: `
declare const x: string[] | null;
// eslint-disable-next-line
if (x) {
}
      `,
      options: [
        {
          allowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing: true,
        },
      ],
      languageOptions: {
        parserOptions: {
          project: './tsconfig.unstrict.json',
        },
      },
    },
    `
interface Foo {
  [key: string]: [string] | undefined;
}

type OptionalFoo = Foo | undefined;
declare const foo: OptionalFoo;
foo?.test?.length;
    `,
    `
interface Foo {
  [key: number]: [string] | undefined;
}

type OptionalFoo = Foo | undefined;
declare const foo: OptionalFoo;
foo?.[1]?.length;
    `,
    `
declare let foo: number | null;
foo ??= 1;
    `,
    `
declare let foo: number;
foo ||= 1;
    `,
    `
declare const foo: { bar: { baz?: number; qux: number } };
type Key = 'baz' | 'qux';
declare const key: Key;
foo.bar[key] ??= 1;
    `,
    `
enum Keys {
  A = 'A',
  B = 'B',
}
type Foo = {
  [Keys.A]: number | null;
  [Keys.B]: number;
};
declare const foo: Foo;
declare const key: Keys;
foo[key] ??= 1;
    `,
    {
      code: `
declare const foo: { bar?: number };
foo.bar ??= 1;
      `,
      languageOptions: {
        parserOptions: {
          project: './tsconfig.exactOptionalPropertyTypes.json',
        },
      },
    },
    {
      code: `
declare const foo: { bar: { baz?: number } };
foo['bar'].baz ??= 1;
      `,
      languageOptions: {
        parserOptions: {
          project: './tsconfig.exactOptionalPropertyTypes.json',
        },
      },
    },
    {
      code: `
declare const foo: { bar: { baz?: number; qux: number } };
type Key = 'baz' | 'qux';
declare const key: Key;
foo.bar[key] ??= 1;
      `,
      languageOptions: {
        parserOptions: {
          project: './tsconfig.exactOptionalPropertyTypes.json',
        },
      },
    },
    `
declare let foo: number;
foo &&= 1;
    `,
    `
function foo<T extends object>(arg: T, key: keyof T): void {
  arg[key] ??= 'default';
}
    `,
    // https://github.com/typescript-eslint/typescript-eslint/issues/6264
    `
function get<Obj, Key extends keyof Obj>(obj: Obj, key: Key) {
  const value = obj[key];
  if (value) {
    return value;
  }
  throw new Error('BOOM!');
}

get({ foo: null }, 'foo');
    `,
    {
      code: `
function getElem(dict: Record<string, { foo: string }>, key: string) {
  if (dict[key]) {
    return dict[key].foo;
  } else {
    return '';
  }
}
      `,
      languageOptions: {
        parserOptions: {
          project: './tsconfig.noUncheckedIndexedAccess.json',
          projectService: false,
        },
      },
    },
    `
type Foo = { bar: () => number | undefined } | null;
declare const foo: Foo;
foo?.bar()?.toExponential();
    `,
    `
type Foo = (() => number | undefined) | null;
declare const foo: Foo;
foo?.()?.toExponential();
    `,
    `
type FooUndef = () => undefined;
type FooNum = () => number;
type Foo = FooUndef | FooNum | null;
declare const foo: Foo;
foo?.()?.toExponential();
    `,
    `
type Foo = { [key: string]: () => number | undefined } | null;
declare const foo: Foo;
foo?.['bar']()?.toExponential();
    `,
    `
declare function foo(): void | { key: string };
const bar = foo()?.key;
    `,
    `
type fn = () => void;
declare function foo(): void | fn;
const bar = foo()?.();
    `,
    `
declare function maybe<T>(v: T): T | void;
const a = maybe({ key1: { key2: maybe({ i: 1 }) } })?.key1.key2?.i;
    `,
    `
type Foo = { bar: { baz: number } | void } | null;
declare const foo: Foo;
foo?.bar?.baz;
    `,
    `
declare function maybeFn(): { fn: (() => number) | void } | undefined;
maybeFn()?.fn?.();
    `,
    {
      code: `
class ConsistentRand {
  #rand?: number;

  getCachedRand() {
    this.#rand ??= Math.random();
    return this.#rand;
  }
}
      `,
      languageOptions: {
        parserOptions: {
          project: './tsconfig.exactOptionalPropertyTypes.json',
        },
      },
    },
    {
      code: `
declare function assert(x: unknown): asserts x;

assert(Math.random() > 0.5);
      `,
      options: [
        {
          checkTypePredicates: true,
        },
      ],
    },
    {
      code: `
declare function assert(x: unknown, y: unknown): asserts x;

assert(Math.random() > 0.5, true);
      `,
      options: [
        {
          checkTypePredicates: true,
        },
      ],
    },
    // should not report because option is disabled.
    {
      code: `
declare function assert(x: unknown): asserts x;
assert(true);
      `,
      options: [
        {
          checkTypePredicates: false,
        },
      ],
    },
    // could be argued that this should report since `thisAsserter` is truthy.
    {
      code: `
class ThisAsserter {
  assertThis(this: unknown, arg2: unknown): asserts this {}
}

const thisAsserter: ThisAsserter = new ThisAsserter();
thisAsserter.assertThis(true);
      `,
      options: [
        {
          checkTypePredicates: true,
        },
      ],
    },
    // could be argued that this should report since `thisAsserter` is truthy.
    {
      code: `
class ThisAsserter {
  assertThis(this: unknown, arg2: unknown): asserts this {}
}

const thisAsserter: ThisAsserter = new ThisAsserter();
thisAsserter.assertThis(Math.random());
      `,
      options: [
        {
          checkTypePredicates: true,
        },
      ],
    },
    {
      code: `
declare function assert(x: unknown): asserts x;
assert(...[]);
      `,
      options: [
        {
          checkTypePredicates: true,
        },
      ],
    },
    // ok to report if we start unpacking spread params one day.
    {
      code: `
declare function assert(x: unknown): asserts x;
assert(...[], {});
      `,
      options: [
        {
          checkTypePredicates: true,
        },
      ],
    },
    {
      code: `
declare function assertString(x: unknown): asserts x is string;
declare const a: string;
assertString(a);
      `,
      options: [
        {
          checkTypePredicates: false,
        },
      ],
    },
    // typeGuardAlreadyIsType: lock down the {{typeGuardOrAssertionFunction}}
    // placeholder fill for both type-predicate and assertion-function forms.
    {
      code: `
declare function isString(x: unknown): x is string;
declare const a: string;
isString(a);
      `,
      options: [
        {
          checkTypePredicates: false,
        },
      ],
    },
    // Technically, this has type 'falafel' and not string.
    {
      code: `
declare function assertString(x: unknown): asserts x is string;
assertString('falafel');
      `,
      options: [
        {
          checkTypePredicates: true,
        },
      ],
    },
    // Technically, this has type 'falafel' and not string.
    {
      code: `
declare function isString(x: unknown): x is string;
isString('falafel');
      `,
      options: [
        {
          checkTypePredicates: true,
        },
      ],
    },
    // string | number | bigint is not a subtype of string | number
    {
      code: `
declare function isStringOrNumber(x: unknown): x is string | number;
declare const s: string | number | bigint;
if (isStringOrNumber(s)) {
}
      `,
      options: [
        {
          checkTypePredicates: true,
        },
      ],
    },
    // Narrower is not a subtype of Wider
    {
      code: `
interface Wider {
  a: string;
}
interface Narrower {
  a: string;
  b: number;
}
declare function isNarrower(x: unknown): x is Narrower;
declare const w: Wider;
if (isNarrower(w)) {
}
      `,
      options: [
        {
          checkTypePredicates: true,
        },
      ],
    },
    `
type A = { [name in Lowercase<string>]?: A };
declare const a: A;
a.a?.a?.a;
    `,
    `
interface T {
  [name: Lowercase<string>]: {
    [name: Lowercase<string>]: {
      [name: Lowercase<string>]: {
        value: 'value';
      };
    };
  };
  [name: Uppercase<string>]: null | {
    [name: Uppercase<string>]: null | {
      [name: Uppercase<string>]: null | {
        VALUE: 'VALUE';
      };
    };
  };
}

declare const t: T;

t.a.a.a.value;
t.A?.A?.A?.VALUE;
    `,
  ],
  invalid: [
    // Ensure that it's checking in all the right places
    {
      code: `
const b1 = true;
declare const b2: boolean;
const t1 = b1 && b2;
const t2 = b1 || b2;
if (b1 && b2) {
}
if (b2 && b1) {
}
while (b1 && b2) {}
while (b2 && b1) {}
for (let i = 0; b1 && b2; i++) {
  break;
}
const t1 = b1 && b2 ? 'yes' : 'no';
const t1 = b2 && b1 ? 'yes' : 'no';
switch (b1) {
  case true:
  default:
}
      `,
      errors: [
        {
          messageId: 'alwaysTruthy',
          message: 'Unnecessary conditional, value is always truthy.',
          line: 4,
          column: 12,
          endLine: 4,
          endColumn: 14,
        },
        {
          messageId: 'alwaysTruthy',
          message: 'Unnecessary conditional, value is always truthy.',
          line: 5,
          column: 12,
          endLine: 5,
          endColumn: 14,
        },
        {
          messageId: 'alwaysTruthy',
          message: 'Unnecessary conditional, value is always truthy.',
          line: 6,
          column: 5,
          endLine: 6,
          endColumn: 7,
        },
        {
          messageId: 'alwaysTruthy',
          message: 'Unnecessary conditional, value is always truthy.',
          line: 8,
          column: 11,
          endLine: 8,
          endColumn: 13,
        },
        {
          messageId: 'alwaysTruthy',
          message: 'Unnecessary conditional, value is always truthy.',
          line: 10,
          column: 8,
          endLine: 10,
          endColumn: 10,
        },
        {
          messageId: 'alwaysTruthy',
          message: 'Unnecessary conditional, value is always truthy.',
          line: 11,
          column: 14,
          endLine: 11,
          endColumn: 16,
        },
        {
          messageId: 'alwaysTruthy',
          message: 'Unnecessary conditional, value is always truthy.',
          line: 12,
          column: 17,
          endLine: 12,
          endColumn: 19,
        },
        {
          messageId: 'alwaysTruthy',
          message: 'Unnecessary conditional, value is always truthy.',
          line: 15,
          column: 12,
          endLine: 15,
          endColumn: 14,
        },
        {
          messageId: 'alwaysTruthy',
          message: 'Unnecessary conditional, value is always truthy.',
          line: 16,
          column: 18,
          endLine: 16,
          endColumn: 20,
        },
        {
          messageId: 'comparisonBetweenLiteralTypes',
          message:
            'Unnecessary conditional, comparison is always true, since `true === true` is true.',
          line: 18,
          column: 8,
          endLine: 18,
          endColumn: 12,
        },
      ],
    },
    // Ensure that it's complaining about the right things
    unnecessaryConditionTest('object', 'alwaysTruthy'),
    unnecessaryConditionTest('object | true', 'alwaysTruthy'),
    // Two falsy literals
    unnecessaryConditionTest('"" | false', 'alwaysFalsy'),
    unnecessaryConditionTest('"always truthy"', 'alwaysTruthy'),
    unnecessaryConditionTest('undefined', 'alwaysFalsy'),
    unnecessaryConditionTest('null', 'alwaysFalsy'),
    unnecessaryConditionTest('void', 'alwaysFalsy'),
    unnecessaryConditionTest('never', 'never'),
    unnecessaryConditionTest('string & number', 'never'),
    // More complex logical expressions
    {
      code: `
declare const falseyBigInt: 0n;
if (falseyBigInt) {
}
      `,
      errors: [
        {
          messageId: 'alwaysFalsy',
          message: 'Unnecessary conditional, value is always falsy.',
          line: 3,
          column: 5,
          endLine: 3,
          endColumn: 17,
        },
      ],
    },
    {
      code: `
declare const posbigInt: 1n;
if (posbigInt) {
}
      `,
      errors: [
        {
          messageId: 'alwaysTruthy',
          message: 'Unnecessary conditional, value is always truthy.',
          line: 3,
          column: 5,
          endLine: 3,
          endColumn: 14,
        },
      ],
    },
    {
      code: `
declare const negBigInt: -2n;
if (negBigInt) {
}
      `,
      errors: [
        {
          messageId: 'alwaysTruthy',
          message: 'Unnecessary conditional, value is always truthy.',
          line: 3,
          column: 5,
          endLine: 3,
          endColumn: 14,
        },
      ],
    },
    {
      code: `
declare const b1: boolean;
declare const b2: boolean;
if (true && b1 && b2) {
}
if (b1 && false && b2) {
}
if (b1 || b2 || true) {
}
      `,
      errors: [
        {
          messageId: 'alwaysTruthy',
          message: 'Unnecessary conditional, value is always truthy.',
          line: 4,
          column: 5,
          endLine: 4,
          endColumn: 9,
        },
        {
          messageId: 'alwaysFalsy',
          message: 'Unnecessary conditional, value is always falsy.',
          line: 6,
          column: 11,
          endLine: 6,
          endColumn: 16,
        },
        {
          messageId: 'alwaysTruthy',
          message: 'Unnecessary conditional, value is always truthy.',
          line: 8,
          column: 17,
          endLine: 8,
          endColumn: 21,
        },
      ],
    },
    // Generic type params
    {
      code: `
function test<T extends object>(t: T) {
  return t ? 'yes' : 'no';
}
      `,
      errors: [
        {
          messageId: 'alwaysTruthy',
          message: 'Unnecessary conditional, value is always truthy.',
          line: 3,
          column: 10,
          endLine: 3,
          endColumn: 11,
        },
      ],
    },
    {
      code: `
function test<T extends false>(t: T) {
  return t ? 'yes' : 'no';
}
      `,
      errors: [
        {
          messageId: 'alwaysFalsy',
          message: 'Unnecessary conditional, value is always falsy.',
          line: 3,
          column: 10,
          endLine: 3,
          endColumn: 11,
        },
      ],
    },
    {
      code: `
function test<T extends 'a' | 'b'>(t: T) {
  return t ? 'yes' : 'no';
}
      `,
      errors: [
        {
          messageId: 'alwaysTruthy',
          message: 'Unnecessary conditional, value is always truthy.',
          line: 3,
          column: 10,
          endLine: 3,
          endColumn: 11,
        },
      ],
    },
    // Boolean expressions
    {
      code: `
function test(a: 'a') {
  return a === 'a';
}
      `,
      errors: [
        {
          messageId: 'comparisonBetweenLiteralTypes',
          message:
            'Unnecessary conditional, comparison is always true, since `"a" === "a"` is true.',
          line: 3,
          column: 10,
          endLine: 3,
          endColumn: 19,
        },
      ],
    },
    {
      code: `
declare const a: '34';
declare const b: '56';
a > b;
      `,
      errors: [
        {
          messageId: 'comparisonBetweenLiteralTypes',
          message:
            'Unnecessary conditional, comparison is always false, since `"34" > "56"` is false.',
          line: 4,
          column: 1,
          endLine: 4,
          endColumn: 6,
        },
      ],
    },
    {
      code: `
const y = 1;
if (y === 0) {
}
      `,
      errors: [
        {
          messageId: 'comparisonBetweenLiteralTypes',
          message:
            'Unnecessary conditional, comparison is always false, since `1 === 0` is false.',
          line: 3,
          column: 5,
          endLine: 3,
          endColumn: 12,
        },
      ],
    },
    {
      code: `
// @ts-expect-error
if (1 == '1') {
}
      `,
      errors: [
        {
          messageId: 'comparisonBetweenLiteralTypes',
          message:
            'Unnecessary conditional, comparison is always true, since `1 == "1"` is true.',
          line: 3,
          column: 5,
          endLine: 3,
          endColumn: 13,
        },
      ],
    },
    {
      code: `
2.3 > 2.3;
      `,
      errors: [
        {
          messageId: 'comparisonBetweenLiteralTypes',
          message:
            'Unnecessary conditional, comparison is always false, since `2.3 > 2.3` is false.',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 10,
        },
      ],
    },
    {
      code: `
2.3 >= 2.3;
      `,
      errors: [
        {
          messageId: 'comparisonBetweenLiteralTypes',
          message:
            'Unnecessary conditional, comparison is always true, since `2.3 >= 2.3` is true.',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 11,
        },
      ],
    },
    {
      code: `
2n < 2n;
      `,
      errors: [
        {
          messageId: 'comparisonBetweenLiteralTypes',
          message:
            'Unnecessary conditional, comparison is always false, since `2n < 2n` is false.',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 8,
        },
      ],
    },
    {
      code: `
2n <= 2n;
      `,
      errors: [
        {
          messageId: 'comparisonBetweenLiteralTypes',
          message:
            'Unnecessary conditional, comparison is always true, since `2n <= 2n` is true.',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 9,
        },
      ],
    },
    {
      code: `
-2n !== 2n;
      `,
      errors: [
        {
          messageId: 'comparisonBetweenLiteralTypes',
          message:
            'Unnecessary conditional, comparison is always true, since `-2n !== 2n` is true.',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 11,
        },
      ],
    },
    {
      code: `
// @ts-expect-error
if (1 == '2') {
}
      `,
      errors: [
        {
          messageId: 'comparisonBetweenLiteralTypes',
          message:
            'Unnecessary conditional, comparison is always false, since `1 == "2"` is false.',
          line: 3,
          column: 5,
          endLine: 3,
          endColumn: 13,
        },
      ],
    },
    {
      code: `
// @ts-expect-error
if (1 != '2') {
}
      `,
      errors: [
        {
          messageId: 'comparisonBetweenLiteralTypes',
          message:
            'Unnecessary conditional, comparison is always true, since `1 != "2"` is true.',
          line: 3,
          column: 5,
          endLine: 3,
          endColumn: 13,
        },
      ],
    },
    {
      code: `
enum Foo {
  a = 1,
  b = 2,
}

const x = Foo.a;
if (x === Foo.a) {
}
      `,
      errors: [
        {
          messageId: 'comparisonBetweenLiteralTypes',
          message:
            'Unnecessary conditional, comparison is always true, since `Foo.a === Foo.a` is true.',
          line: 8,
          column: 5,
          endLine: 8,
          endColumn: 16,
        },
      ],
    },
    {
      code: `
enum Foo {
  a = 1,
  b = 2,
}

const x = Foo.a;
if (x === 1) {
}
      `,
      errors: [
        {
          messageId: 'comparisonBetweenLiteralTypes',
          message:
            'Unnecessary conditional, comparison is always true, since `Foo.a === 1` is true.',
          line: 8,
          column: 5,
          endLine: 8,
          endColumn: 12,
        },
      ],
    },
    // narrowed to null. always-true because of loose nullish equality
    {
      code: `
function takesMaybeValue(a: null | object) {
  if (a) {
  } else if (a == undefined) {
  }
}
      `,
      errors: [
        {
          messageId: 'comparisonBetweenLiteralTypes',
          message:
            'Unnecessary conditional, comparison is always true, since `null == undefined` is true.',
          line: 4,
          column: 14,
          endLine: 4,
          endColumn: 28,
        },
      ],
    },
    // narrowed to null. always-false because of strict undefined equality
    {
      code: `
function takesMaybeValue(a: null | object) {
  if (a) {
  } else if (a === undefined) {
  }
}
      `,
      errors: [
        {
          messageId: 'comparisonBetweenLiteralTypes',
          message:
            'Unnecessary conditional, comparison is always false, since `null === undefined` is false.',
          line: 4,
          column: 14,
          endLine: 4,
          endColumn: 29,
        },
      ],
    },
    // narrowed to null. always-false because of loose nullish equality
    {
      code: `
function takesMaybeValue(a: null | object) {
  if (a) {
  } else if (a != undefined) {
  }
}
      `,
      errors: [
        {
          messageId: 'comparisonBetweenLiteralTypes',
          message:
            'Unnecessary conditional, comparison is always false, since `null != undefined` is false.',
          line: 4,
          column: 14,
          endLine: 4,
          endColumn: 28,
        },
      ],
    },
    // narrowed to null. always-true because of strict undefined equality
    {
      code: `
function takesMaybeValue(a: null | object) {
  if (a) {
  } else if (a !== undefined) {
  }
}
      `,
      errors: [
        {
          messageId: 'comparisonBetweenLiteralTypes',
          message:
            'Unnecessary conditional, comparison is always true, since `null !== undefined` is true.',
          line: 4,
          column: 14,
          endLine: 4,
          endColumn: 29,
        },
      ],
    },
    {
      code: `
true === false;
      `,
      errors: [
        {
          messageId: 'comparisonBetweenLiteralTypes',
          message:
            'Unnecessary conditional, comparison is always false, since `true === false` is false.',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 15,
        },
      ],
    },
    {
      code: `
true === true;
      `,
      errors: [
        {
          messageId: 'comparisonBetweenLiteralTypes',
          message:
            'Unnecessary conditional, comparison is always true, since `true === true` is true.',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 14,
        },
      ],
    },
    {
      code: `
true === undefined;
      `,
      errors: [
        {
          messageId: 'comparisonBetweenLiteralTypes',
          message:
            'Unnecessary conditional, comparison is always false, since `true === undefined` is false.',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 19,
        },
      ],
    },
    // Workaround https://github.com/microsoft/TypeScript/issues/37160
    {
      code: `
function test(a: string) {
  const t1 = a === undefined;
  const t2 = undefined === a;
  const t3 = a !== undefined;
  const t4 = undefined !== a;
  const t5 = a === null;
  const t6 = null === a;
  const t7 = a !== null;
  const t8 = null !== a;
}
      `,
      errors: [
        {
          messageId: 'noOverlapBooleanExpression',
          message: 'Unnecessary conditional, the types have no overlap.',
          line: 3,
          column: 14,
          endLine: 3,
          endColumn: 29,
        },
        {
          messageId: 'noOverlapBooleanExpression',
          message: 'Unnecessary conditional, the types have no overlap.',
          line: 4,
          column: 14,
          endLine: 4,
          endColumn: 29,
        },
        {
          messageId: 'noOverlapBooleanExpression',
          message: 'Unnecessary conditional, the types have no overlap.',
          line: 5,
          column: 14,
          endLine: 5,
          endColumn: 29,
        },
        {
          messageId: 'noOverlapBooleanExpression',
          message: 'Unnecessary conditional, the types have no overlap.',
          line: 6,
          column: 14,
          endLine: 6,
          endColumn: 29,
        },
        {
          messageId: 'noOverlapBooleanExpression',
          message: 'Unnecessary conditional, the types have no overlap.',
          line: 7,
          column: 14,
          endLine: 7,
          endColumn: 24,
        },
        {
          messageId: 'noOverlapBooleanExpression',
          message: 'Unnecessary conditional, the types have no overlap.',
          line: 8,
          column: 14,
          endLine: 8,
          endColumn: 24,
        },
        {
          messageId: 'noOverlapBooleanExpression',
          message: 'Unnecessary conditional, the types have no overlap.',
          line: 9,
          column: 14,
          endLine: 9,
          endColumn: 24,
        },
        {
          messageId: 'noOverlapBooleanExpression',
          message: 'Unnecessary conditional, the types have no overlap.',
          line: 10,
          column: 14,
          endLine: 10,
          endColumn: 24,
        },
      ],
    },
    {
      code: `
function test(a?: string) {
  const t1 = a === undefined;
  const t2 = undefined === a;
  const t3 = a !== undefined;
  const t4 = undefined !== a;
  const t5 = a === null;
  const t6 = null === a;
  const t7 = a !== null;
  const t8 = null !== a;
}
      `,
      errors: [
        {
          messageId: 'noOverlapBooleanExpression',
          message: 'Unnecessary conditional, the types have no overlap.',
          line: 7,
          column: 14,
          endLine: 7,
          endColumn: 24,
        },
        {
          messageId: 'noOverlapBooleanExpression',
          message: 'Unnecessary conditional, the types have no overlap.',
          line: 8,
          column: 14,
          endLine: 8,
          endColumn: 24,
        },
        {
          messageId: 'noOverlapBooleanExpression',
          message: 'Unnecessary conditional, the types have no overlap.',
          line: 9,
          column: 14,
          endLine: 9,
          endColumn: 24,
        },
        {
          messageId: 'noOverlapBooleanExpression',
          message: 'Unnecessary conditional, the types have no overlap.',
          line: 10,
          column: 14,
          endLine: 10,
          endColumn: 24,
        },
      ],
    },
    {
      code: `
function test(a: null | string) {
  const t1 = a === undefined;
  const t2 = undefined === a;
  const t3 = a !== undefined;
  const t4 = undefined !== a;
  const t5 = a === null;
  const t6 = null === a;
  const t7 = a !== null;
  const t8 = null !== a;
}
      `,
      errors: [
        {
          messageId: 'noOverlapBooleanExpression',
          message: 'Unnecessary conditional, the types have no overlap.',
          line: 3,
          column: 14,
          endLine: 3,
          endColumn: 29,
        },
        {
          messageId: 'noOverlapBooleanExpression',
          message: 'Unnecessary conditional, the types have no overlap.',
          line: 4,
          column: 14,
          endLine: 4,
          endColumn: 29,
        },
        {
          messageId: 'noOverlapBooleanExpression',
          message: 'Unnecessary conditional, the types have no overlap.',
          line: 5,
          column: 14,
          endLine: 5,
          endColumn: 29,
        },
        {
          messageId: 'noOverlapBooleanExpression',
          message: 'Unnecessary conditional, the types have no overlap.',
          line: 6,
          column: 14,
          endLine: 6,
          endColumn: 29,
        },
      ],
    },
    {
      code: `
function test<T extends object>(a: T) {
  const t1 = a == null;
  const t2 = null == a;
  const t3 = a != null;
  const t4 = null != a;
  const t5 = a == undefined;
  const t6 = undefined == a;
  const t7 = a != undefined;
  const t8 = undefined != a;
  const t9 = a === null;
  const t10 = null === a;
  const t11 = a !== null;
  const t12 = null !== a;
  const t13 = a === undefined;
  const t14 = undefined === a;
  const t15 = a !== undefined;
  const t16 = undefined !== a;
}
      `,
      errors: [
        {
          messageId: 'noOverlapBooleanExpression',
          message: 'Unnecessary conditional, the types have no overlap.',
          line: 3,
          column: 14,
          endLine: 3,
          endColumn: 23,
        },
        {
          messageId: 'noOverlapBooleanExpression',
          message: 'Unnecessary conditional, the types have no overlap.',
          line: 4,
          column: 14,
          endLine: 4,
          endColumn: 23,
        },
        {
          messageId: 'noOverlapBooleanExpression',
          message: 'Unnecessary conditional, the types have no overlap.',
          line: 5,
          column: 14,
          endLine: 5,
          endColumn: 23,
        },
        {
          messageId: 'noOverlapBooleanExpression',
          message: 'Unnecessary conditional, the types have no overlap.',
          line: 6,
          column: 14,
          endLine: 6,
          endColumn: 23,
        },
        {
          messageId: 'noOverlapBooleanExpression',
          message: 'Unnecessary conditional, the types have no overlap.',
          line: 7,
          column: 14,
          endLine: 7,
          endColumn: 28,
        },
        {
          messageId: 'noOverlapBooleanExpression',
          message: 'Unnecessary conditional, the types have no overlap.',
          line: 8,
          column: 14,
          endLine: 8,
          endColumn: 28,
        },
        {
          messageId: 'noOverlapBooleanExpression',
          message: 'Unnecessary conditional, the types have no overlap.',
          line: 9,
          column: 14,
          endLine: 9,
          endColumn: 28,
        },
        {
          messageId: 'noOverlapBooleanExpression',
          message: 'Unnecessary conditional, the types have no overlap.',
          line: 10,
          column: 14,
          endLine: 10,
          endColumn: 28,
        },
        {
          messageId: 'noOverlapBooleanExpression',
          message: 'Unnecessary conditional, the types have no overlap.',
          line: 11,
          column: 14,
          endLine: 11,
          endColumn: 24,
        },
        {
          messageId: 'noOverlapBooleanExpression',
          message: 'Unnecessary conditional, the types have no overlap.',
          line: 12,
          column: 15,
          endLine: 12,
          endColumn: 25,
        },
        {
          messageId: 'noOverlapBooleanExpression',
          message: 'Unnecessary conditional, the types have no overlap.',
          line: 13,
          column: 15,
          endLine: 13,
          endColumn: 25,
        },
        {
          messageId: 'noOverlapBooleanExpression',
          message: 'Unnecessary conditional, the types have no overlap.',
          line: 14,
          column: 15,
          endLine: 14,
          endColumn: 25,
        },
        {
          messageId: 'noOverlapBooleanExpression',
          message: 'Unnecessary conditional, the types have no overlap.',
          line: 15,
          column: 15,
          endLine: 15,
          endColumn: 30,
        },
        {
          messageId: 'noOverlapBooleanExpression',
          message: 'Unnecessary conditional, the types have no overlap.',
          line: 16,
          column: 15,
          endLine: 16,
          endColumn: 30,
        },
        {
          messageId: 'noOverlapBooleanExpression',
          message: 'Unnecessary conditional, the types have no overlap.',
          line: 17,
          column: 15,
          endLine: 17,
          endColumn: 30,
        },
        {
          messageId: 'noOverlapBooleanExpression',
          message: 'Unnecessary conditional, the types have no overlap.',
          line: 18,
          column: 15,
          endLine: 18,
          endColumn: 30,
        },
      ],
    },
    // Nullish coalescing operator
    {
      code: `
function test(a: string) {
  return a ?? 'default';
}
      `,
      errors: [
        {
          messageId: 'neverNullish',
          message:
            'Unnecessary conditional, expected left-hand side of `??` operator to be possibly null or undefined.',
          line: 3,
          column: 10,
          endLine: 3,
          endColumn: 11,
        },
      ],
    },
    {
      code: `
function test(a: string | false) {
  return a ?? 'default';
}
      `,
      errors: [
        {
          messageId: 'neverNullish',
          message:
            'Unnecessary conditional, expected left-hand side of `??` operator to be possibly null or undefined.',
          line: 3,
          column: 10,
          endLine: 3,
          endColumn: 11,
        },
      ],
    },
    {
      code: `
function test<T extends string>(a: T) {
  return a ?? 'default';
}
      `,
      errors: [
        {
          messageId: 'neverNullish',
          message:
            'Unnecessary conditional, expected left-hand side of `??` operator to be possibly null or undefined.',
          line: 3,
          column: 10,
          endLine: 3,
          endColumn: 11,
        },
      ],
    },
    // nullish + array index without optional chaining
    {
      code: `
function test(a: { foo: string }[]) {
  return a[0].foo ?? 'default';
}
      `,
      errors: [
        {
          messageId: 'neverNullish',
          message:
            'Unnecessary conditional, expected left-hand side of `??` operator to be possibly null or undefined.',
          line: 3,
          column: 10,
          endLine: 3,
          endColumn: 18,
        },
      ],
    },
    {
      code: `
function test(a: null) {
  return a ?? 'default';
}
      `,
      errors: [
        {
          messageId: 'alwaysNullish',
          message:
            'Unnecessary conditional, left-hand side of `??` operator is always `null` or `undefined`.',
          line: 3,
          column: 10,
          endLine: 3,
          endColumn: 11,
        },
      ],
    },
    {
      code: `
function test(a: null[]) {
  return a[0] ?? 'default';
}
      `,
      errors: [
        {
          messageId: 'alwaysNullish',
          message:
            'Unnecessary conditional, left-hand side of `??` operator is always `null` or `undefined`.',
          line: 3,
          column: 10,
          endLine: 3,
          endColumn: 14,
        },
      ],
    },
    {
      code: `
function test<T extends null>(a: T) {
  return a ?? 'default';
}
      `,
      errors: [
        {
          messageId: 'alwaysNullish',
          message:
            'Unnecessary conditional, left-hand side of `??` operator is always `null` or `undefined`.',
          line: 3,
          column: 10,
          endLine: 3,
          endColumn: 11,
        },
      ],
    },
    {
      code: `
function test(a: never) {
  return a ?? 'default';
}
      `,
      errors: [
        {
          messageId: 'never',
          message: 'Unnecessary conditional, value is `never`.',
          line: 3,
          column: 10,
          endLine: 3,
          endColumn: 11,
        },
      ],
    },
    {
      code: `
function test<T extends { foo: number }, K extends 'foo'>(num: T[K]) {
  num ?? 'default';
}
      `,
      errors: [
        {
          messageId: 'neverNullish',
          message:
            'Unnecessary conditional, expected left-hand side of `??` operator to be possibly null or undefined.',
          line: 3,
          column: 3,
          endLine: 3,
          endColumn: 6,
        },
      ],
    },
    // Predicate functions
    {
      code: `
[1, 3, 5].filter(() => true);
[1, 2, 3].find(() => {
  return false;
});

// with non-literal array
function nothing(x: string[]) {
  return x.filter(() => false);
}
// with readonly array
function nothing2(x: readonly string[]) {
  return x.filter(() => false);
}
// with tuple
function nothing3(x: [string, string]) {
  return x.filter(() => false);
}
      `,
      errors: [
        {
          messageId: 'alwaysTruthy',
          message: 'Unnecessary conditional, value is always truthy.',
          line: 2,
          column: 24,
          endLine: 2,
          endColumn: 28,
        },
        {
          messageId: 'alwaysFalsy',
          message: 'Unnecessary conditional, value is always falsy.',
          line: 4,
          column: 10,
          endLine: 4,
          endColumn: 15,
        },
        {
          messageId: 'alwaysFalsy',
          message: 'Unnecessary conditional, value is always falsy.',
          line: 9,
          column: 25,
          endLine: 9,
          endColumn: 30,
        },
        {
          messageId: 'alwaysFalsy',
          message: 'Unnecessary conditional, value is always falsy.',
          line: 13,
          column: 25,
          endLine: 13,
          endColumn: 30,
        },
        {
          messageId: 'alwaysFalsy',
          message: 'Unnecessary conditional, value is always falsy.',
          line: 17,
          column: 25,
          endLine: 17,
          endColumn: 30,
        },
      ],
    },
    {
      code: `
declare const test: <T extends true>() => T;

[1, null].filter(test);
      `,
      errors: [
        {
          messageId: 'alwaysTruthyFunc',
          message:
            'This callback should return a conditional, but return is always truthy.',
          line: 4,
          column: 18,
          endLine: 4,
          endColumn: 22,
        },
      ],
    },
    // Indexing cases
    // This is an error because 'dict' doesn't represent
    //  the potential for undefined in its types
    {
      code: `
declare const dict: Record<string, object>;
if (dict['mightNotExist']) {
}
      `,
      errors: [
        {
          messageId: 'alwaysTruthy',
          message: 'Unnecessary conditional, value is always truthy.',
          line: 3,
          column: 5,
          endLine: 3,
          endColumn: 26,
        },
      ],
    },
    // Should still check tuples when accessed with literal numbers, since they don't have
    //   unsound index signatures
    {
      code: `
const x = [{}] as [{ foo: string }];
if (x[0]) {
}
if (x[0]?.foo) {
}
      `,
      errors: [
        {
          messageId: 'alwaysTruthy',
          message: 'Unnecessary conditional, value is always truthy.',
          line: 3,
          column: 5,
          endLine: 3,
          endColumn: 9,
        },
        {
          messageId: 'neverOptionalChain',
          message: 'Unnecessary optional chain on a non-nullish value.',
          line: 5,
          column: 9,
          endLine: 5,
          endColumn: 11,
          suggestions: [
            {
              messageId: 'suggestRemoveOptionalChain',
              desc: 'Remove unnecessary optional chain',
              output: `
const x = [{}] as [{ foo: string }];
if (x[0]) {
}
if (x[0].foo) {
}
      `,
            },
          ],
        },
      ],
    },
    // Shouldn't mistake this for an array indexing case
    {
      code: `
declare const arr: object[];
if (arr.filter) {
}
      `,
      errors: [
        {
          messageId: 'alwaysTruthy',
          message: 'Unnecessary conditional, value is always truthy.',
          line: 3,
          column: 5,
          endLine: 3,
          endColumn: 15,
        },
      ],
    },
    {
      code: `
function truthy() {
  return [];
}
function falsy() {}
[1, 3, 5].filter(truthy);
[1, 2, 3].find(falsy);
[1, 2, 3].findLastIndex(falsy);
      `,
      errors: [
        {
          messageId: 'alwaysTruthyFunc',
          message:
            'This callback should return a conditional, but return is always truthy.',
          line: 6,
          column: 18,
          endLine: 6,
          endColumn: 24,
        },
        {
          messageId: 'alwaysFalsyFunc',
          message:
            'This callback should return a conditional, but return is always falsy.',
          line: 7,
          column: 16,
          endLine: 7,
          endColumn: 21,
        },
        {
          messageId: 'alwaysFalsyFunc',
          message:
            'This callback should return a conditional, but return is always falsy.',
          line: 8,
          column: 25,
          endLine: 8,
          endColumn: 30,
        },
      ],
    },
    // Supports generics
    // const isTruthy = <T>(t: T) => T;
    // // Valid: numbers can be truthy or falsy (0).
    // [0,1,2,3].filter(isTruthy);
    // // Invalid: arrays are always falsy.
    // [[1,2], [3,4]].filter(isTruthy);
    {
      code: `
declare const test: true;

while (test) {}
      `,
      options: [
        {
          allowConstantLoopConditions: false,
        },
      ],
      errors: [
        {
          messageId: 'alwaysTruthy',
          message: 'Unnecessary conditional, value is always truthy.',
          line: 4,
          column: 8,
          endLine: 4,
          endColumn: 12,
        },
      ],
    },
    {
      code: `
declare const test: true;

for (; test; ) {}
      `,
      options: [
        {
          allowConstantLoopConditions: false,
        },
      ],
      errors: [
        {
          messageId: 'alwaysTruthy',
          message: 'Unnecessary conditional, value is always truthy.',
          line: 4,
          column: 8,
          endLine: 4,
          endColumn: 12,
        },
      ],
    },
    {
      code: `
declare const test: true;

do {} while (test);
      `,
      options: [
        {
          allowConstantLoopConditions: false,
        },
      ],
      errors: [
        {
          messageId: 'alwaysTruthy',
          message: 'Unnecessary conditional, value is always truthy.',
          line: 4,
          column: 14,
          endLine: 4,
          endColumn: 18,
        },
      ],
    },
    {
      code: `
declare const test: true;

while (test) {}
      `,
      options: [
        {
          allowConstantLoopConditions: 'never',
        },
      ],
      errors: [
        {
          messageId: 'alwaysTruthy',
          message: 'Unnecessary conditional, value is always truthy.',
          line: 4,
          column: 8,
          endLine: 4,
          endColumn: 12,
        },
      ],
    },
    {
      code: `
declare const test: true;

for (; test; ) {}
      `,
      options: [
        {
          allowConstantLoopConditions: 'never',
        },
      ],
      errors: [
        {
          messageId: 'alwaysTruthy',
          message: 'Unnecessary conditional, value is always truthy.',
          line: 4,
          column: 8,
          endLine: 4,
          endColumn: 12,
        },
      ],
    },
    {
      code: `
declare const test: true;

do {} while (test);
      `,
      options: [
        {
          allowConstantLoopConditions: 'never',
        },
      ],
      errors: [
        {
          messageId: 'alwaysTruthy',
          message: 'Unnecessary conditional, value is always truthy.',
          line: 4,
          column: 14,
          endLine: 4,
          endColumn: 18,
        },
      ],
    },
    {
      code: `
declare const test: true;

while (test) {}
      `,
      options: [
        {
          allowConstantLoopConditions: 'only-allowed-literals',
        },
      ],
      errors: [
        {
          messageId: 'alwaysTruthy',
          message: 'Unnecessary conditional, value is always truthy.',
          line: 4,
          column: 8,
          endLine: 4,
          endColumn: 12,
        },
      ],
    },
    {
      code: `
declare const test: 1;

while (test) {}
      `,
      options: [
        {
          allowConstantLoopConditions: 'only-allowed-literals',
        },
      ],
      errors: [
        {
          messageId: 'alwaysTruthy',
          message: 'Unnecessary conditional, value is always truthy.',
          line: 4,
          column: 8,
          endLine: 4,
          endColumn: 12,
        },
      ],
    },
    {
      code: `
declare const test: true;

for (; test; ) {}
      `,
      options: [
        {
          allowConstantLoopConditions: 'only-allowed-literals',
        },
      ],
      errors: [
        {
          messageId: 'alwaysTruthy',
          message: 'Unnecessary conditional, value is always truthy.',
          line: 4,
          column: 8,
          endLine: 4,
          endColumn: 12,
        },
      ],
    },
    {
      code: `
declare const test: true;

do {} while (test);
      `,
      options: [
        {
          allowConstantLoopConditions: 'only-allowed-literals',
        },
      ],
      errors: [
        {
          messageId: 'alwaysTruthy',
          message: 'Unnecessary conditional, value is always truthy.',
          line: 4,
          column: 14,
          endLine: 4,
          endColumn: 18,
        },
      ],
    },
    {
      code: `
let shouldRun = true;

while ((shouldRun = true)) {}
      `,
      options: [
        {
          allowConstantLoopConditions: 'only-allowed-literals',
        },
      ],
      errors: [
        {
          messageId: 'alwaysTruthy',
          message: 'Unnecessary conditional, value is always truthy.',
          line: 4,
          column: 9,
          endLine: 4,
          endColumn: 25,
        },
      ],
    },
    {
      code: `
while (2) {}
      `,
      options: [
        {
          allowConstantLoopConditions: 'only-allowed-literals',
        },
      ],
      errors: [
        {
          messageId: 'alwaysTruthy',
          message: 'Unnecessary conditional, value is always truthy.',
          line: 2,
          column: 8,
          endLine: 2,
          endColumn: 9,
        },
      ],
    },
    {
      code: `
while ('truthy') {}
      `,
      options: [
        {
          allowConstantLoopConditions: 'only-allowed-literals',
        },
      ],
      errors: [
        {
          messageId: 'alwaysTruthy',
          message: 'Unnecessary conditional, value is always truthy.',
          line: 2,
          column: 8,
          endLine: 2,
          endColumn: 16,
        },
      ],
    },
    {
      code: `
let foo = { bar: true };
foo?.bar;
foo ?. bar;
foo ?.
  bar;
foo
  ?. bar;
      `,
      errors: [
        {
          messageId: 'neverOptionalChain',
          message: 'Unnecessary optional chain on a non-nullish value.',
          line: 3,
          column: 4,
          endLine: 3,
          endColumn: 6,
          suggestions: [
            {
              messageId: 'suggestRemoveOptionalChain',
              desc: 'Remove unnecessary optional chain',
              output: `
let foo = { bar: true };
foo.bar;
foo ?. bar;
foo ?.
  bar;
foo
  ?. bar;
      `,
            },
          ],
        },
        {
          messageId: 'neverOptionalChain',
          message: 'Unnecessary optional chain on a non-nullish value.',
          line: 4,
          column: 5,
          endLine: 4,
          endColumn: 7,
          suggestions: [
            {
              messageId: 'suggestRemoveOptionalChain',
              desc: 'Remove unnecessary optional chain',
              output: `
let foo = { bar: true };
foo?.bar;
foo . bar;
foo ?.
  bar;
foo
  ?. bar;
      `,
            },
          ],
        },
        {
          messageId: 'neverOptionalChain',
          message: 'Unnecessary optional chain on a non-nullish value.',
          line: 5,
          column: 5,
          endLine: 5,
          endColumn: 7,
          suggestions: [
            {
              messageId: 'suggestRemoveOptionalChain',
              desc: 'Remove unnecessary optional chain',
              output: `
let foo = { bar: true };
foo?.bar;
foo ?. bar;
foo .
  bar;
foo
  ?. bar;
      `,
            },
          ],
        },
        {
          messageId: 'neverOptionalChain',
          message: 'Unnecessary optional chain on a non-nullish value.',
          line: 8,
          column: 3,
          endLine: 8,
          endColumn: 5,
          suggestions: [
            {
              messageId: 'suggestRemoveOptionalChain',
              desc: 'Remove unnecessary optional chain',
              output: `
let foo = { bar: true };
foo?.bar;
foo ?. bar;
foo ?.
  bar;
foo
  . bar;
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
let foo = () => {};
foo?.();
foo ?. ();
foo ?.
  ();
foo
  ?. ();
      `,
      errors: [
        {
          messageId: 'neverOptionalChain',
          message: 'Unnecessary optional chain on a non-nullish value.',
          line: 3,
          column: 4,
          endLine: 3,
          endColumn: 6,
          suggestions: [
            {
              messageId: 'suggestRemoveOptionalChain',
              desc: 'Remove unnecessary optional chain',
              output: `
let foo = () => {};
foo();
foo ?. ();
foo ?.
  ();
foo
  ?. ();
      `,
            },
          ],
        },
        {
          messageId: 'neverOptionalChain',
          message: 'Unnecessary optional chain on a non-nullish value.',
          line: 4,
          column: 5,
          endLine: 4,
          endColumn: 7,
          suggestions: [
            {
              messageId: 'suggestRemoveOptionalChain',
              desc: 'Remove unnecessary optional chain',
              output: `
let foo = () => {};
foo?.();
foo  ();
foo ?.
  ();
foo
  ?. ();
      `,
            },
          ],
        },
        {
          messageId: 'neverOptionalChain',
          message: 'Unnecessary optional chain on a non-nullish value.',
          line: 5,
          column: 5,
          endLine: 5,
          endColumn: 7,
          suggestions: [
            {
              messageId: 'suggestRemoveOptionalChain',
              desc: 'Remove unnecessary optional chain',
              output: `
let foo = () => {};
foo?.();
foo ?. ();
foo\x20
  ();
foo
  ?. ();
      `,
            },
          ],
        },
        {
          messageId: 'neverOptionalChain',
          message: 'Unnecessary optional chain on a non-nullish value.',
          line: 8,
          column: 3,
          endLine: 8,
          endColumn: 5,
          suggestions: [
            {
              messageId: 'suggestRemoveOptionalChain',
              desc: 'Remove unnecessary optional chain',
              output: `
let foo = () => {};
foo?.();
foo ?. ();
foo ?.
  ();
foo
   ();
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
let foo = () => {};
foo?.(bar);
foo ?. (bar);
foo ?.
  (bar);
foo
  ?. (bar);
      `,
      errors: [
        {
          messageId: 'neverOptionalChain',
          message: 'Unnecessary optional chain on a non-nullish value.',
          line: 3,
          column: 4,
          endLine: 3,
          endColumn: 6,
          suggestions: [
            {
              messageId: 'suggestRemoveOptionalChain',
              desc: 'Remove unnecessary optional chain',
              output: `
let foo = () => {};
foo(bar);
foo ?. (bar);
foo ?.
  (bar);
foo
  ?. (bar);
      `,
            },
          ],
        },
        {
          messageId: 'neverOptionalChain',
          message: 'Unnecessary optional chain on a non-nullish value.',
          line: 4,
          column: 5,
          endLine: 4,
          endColumn: 7,
          suggestions: [
            {
              messageId: 'suggestRemoveOptionalChain',
              desc: 'Remove unnecessary optional chain',
              output: `
let foo = () => {};
foo?.(bar);
foo  (bar);
foo ?.
  (bar);
foo
  ?. (bar);
      `,
            },
          ],
        },
        {
          messageId: 'neverOptionalChain',
          message: 'Unnecessary optional chain on a non-nullish value.',
          line: 5,
          column: 5,
          endLine: 5,
          endColumn: 7,
          suggestions: [
            {
              messageId: 'suggestRemoveOptionalChain',
              desc: 'Remove unnecessary optional chain',
              output: `
let foo = () => {};
foo?.(bar);
foo ?. (bar);
foo\x20
  (bar);
foo
  ?. (bar);
      `,
            },
          ],
        },
        {
          messageId: 'neverOptionalChain',
          message: 'Unnecessary optional chain on a non-nullish value.',
          line: 8,
          column: 3,
          endLine: 8,
          endColumn: 5,
          suggestions: [
            {
              messageId: 'suggestRemoveOptionalChain',
              desc: 'Remove unnecessary optional chain',
              output: `
let foo = () => {};
foo?.(bar);
foo ?. (bar);
foo ?.
  (bar);
foo
   (bar);
      `,
            },
          ],
        },
      ],
    },
    {
      code: 'const foo = [1, 2, 3]?.[0];',
      errors: [
        {
          messageId: 'neverOptionalChain',
          message: 'Unnecessary optional chain on a non-nullish value.',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 24,
          suggestions: [
            {
              messageId: 'suggestRemoveOptionalChain',
              desc: 'Remove unnecessary optional chain',
              output: 'const foo = [1, 2, 3][0];',
            },
          ],
        },
      ],
    },
    {
      code: `
declare const x: { a?: { b: string } };
x?.a?.b;
      `,
      errors: [
        {
          messageId: 'neverOptionalChain',
          message: 'Unnecessary optional chain on a non-nullish value.',
          line: 3,
          column: 2,
          endLine: 3,
          endColumn: 4,
          suggestions: [
            {
              messageId: 'suggestRemoveOptionalChain',
              desc: 'Remove unnecessary optional chain',
              output: `
declare const x: { a?: { b: string } };
x.a?.b;
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
declare const x: { a: { b?: { c: string } } };
x.a?.b?.c;
      `,
      errors: [
        {
          messageId: 'neverOptionalChain',
          message: 'Unnecessary optional chain on a non-nullish value.',
          line: 3,
          column: 4,
          endLine: 3,
          endColumn: 6,
          suggestions: [
            {
              messageId: 'suggestRemoveOptionalChain',
              desc: 'Remove unnecessary optional chain',
              output: `
declare const x: { a: { b?: { c: string } } };
x.a.b?.c;
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
let x: { a?: string };
x?.a;
      `,
      errors: [
        {
          messageId: 'neverOptionalChain',
          message: 'Unnecessary optional chain on a non-nullish value.',
          line: 3,
          column: 2,
          endLine: 3,
          endColumn: 4,
          suggestions: [
            {
              messageId: 'suggestRemoveOptionalChain',
              desc: 'Remove unnecessary optional chain',
              output: `
let x: { a?: string };
x.a;
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
declare const foo: { bar: { baz: { c: string } } } | null;
foo?.bar?.baz;
      `,
      errors: [
        {
          messageId: 'neverOptionalChain',
          message: 'Unnecessary optional chain on a non-nullish value.',
          line: 3,
          column: 9,
          endLine: 3,
          endColumn: 11,
          suggestions: [
            {
              messageId: 'suggestRemoveOptionalChain',
              desc: 'Remove unnecessary optional chain',
              output: `
declare const foo: { bar: { baz: { c: string } } } | null;
foo?.bar.baz;
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
declare const foo: { bar?: { baz: { qux: string } } } | null;
foo?.bar?.baz?.qux;
      `,
      errors: [
        {
          messageId: 'neverOptionalChain',
          message: 'Unnecessary optional chain on a non-nullish value.',
          line: 3,
          column: 14,
          endLine: 3,
          endColumn: 16,
          suggestions: [
            {
              messageId: 'suggestRemoveOptionalChain',
              desc: 'Remove unnecessary optional chain',
              output: `
declare const foo: { bar?: { baz: { qux: string } } } | null;
foo?.bar?.baz.qux;
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
declare const foo: { bar: { baz: { qux?: () => {} } } } | null;
foo?.bar?.baz?.qux?.();
      `,
      errors: [
        {
          messageId: 'neverOptionalChain',
          message: 'Unnecessary optional chain on a non-nullish value.',
          line: 3,
          column: 9,
          endLine: 3,
          endColumn: 11,
          suggestions: [
            {
              messageId: 'suggestRemoveOptionalChain',
              desc: 'Remove unnecessary optional chain',
              output: `
declare const foo: { bar: { baz: { qux?: () => {} } } } | null;
foo?.bar.baz?.qux?.();
      `,
            },
          ],
        },
        {
          messageId: 'neverOptionalChain',
          message: 'Unnecessary optional chain on a non-nullish value.',
          line: 3,
          column: 14,
          endLine: 3,
          endColumn: 16,
          suggestions: [
            {
              messageId: 'suggestRemoveOptionalChain',
              desc: 'Remove unnecessary optional chain',
              output: `
declare const foo: { bar: { baz: { qux?: () => {} } } } | null;
foo?.bar?.baz.qux?.();
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
declare const foo: { bar: { baz: { qux: () => {} } } } | null;
foo?.bar?.baz?.qux?.();
      `,
      errors: [
        {
          messageId: 'neverOptionalChain',
          message: 'Unnecessary optional chain on a non-nullish value.',
          line: 3,
          column: 9,
          endLine: 3,
          endColumn: 11,
          suggestions: [
            {
              messageId: 'suggestRemoveOptionalChain',
              desc: 'Remove unnecessary optional chain',
              output: `
declare const foo: { bar: { baz: { qux: () => {} } } } | null;
foo?.bar.baz?.qux?.();
      `,
            },
          ],
        },
        {
          messageId: 'neverOptionalChain',
          message: 'Unnecessary optional chain on a non-nullish value.',
          line: 3,
          column: 14,
          endLine: 3,
          endColumn: 16,
          suggestions: [
            {
              messageId: 'suggestRemoveOptionalChain',
              desc: 'Remove unnecessary optional chain',
              output: `
declare const foo: { bar: { baz: { qux: () => {} } } } | null;
foo?.bar?.baz.qux?.();
      `,
            },
          ],
        },
        {
          messageId: 'neverOptionalChain',
          message: 'Unnecessary optional chain on a non-nullish value.',
          line: 3,
          column: 19,
          endLine: 3,
          endColumn: 21,
          suggestions: [
            {
              messageId: 'suggestRemoveOptionalChain',
              desc: 'Remove unnecessary optional chain',
              output: `
declare const foo: { bar: { baz: { qux: () => {} } } } | null;
foo?.bar?.baz?.qux();
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
type baz = () => { qux: () => {} };
declare const foo: { bar: { baz: baz } } | null;
foo?.bar?.baz?.().qux?.();
      `,
      errors: [
        {
          messageId: 'neverOptionalChain',
          message: 'Unnecessary optional chain on a non-nullish value.',
          line: 4,
          column: 9,
          endLine: 4,
          endColumn: 11,
          suggestions: [
            {
              messageId: 'suggestRemoveOptionalChain',
              desc: 'Remove unnecessary optional chain',
              output: `
type baz = () => { qux: () => {} };
declare const foo: { bar: { baz: baz } } | null;
foo?.bar.baz?.().qux?.();
      `,
            },
          ],
        },
        {
          messageId: 'neverOptionalChain',
          message: 'Unnecessary optional chain on a non-nullish value.',
          line: 4,
          column: 14,
          endLine: 4,
          endColumn: 16,
          suggestions: [
            {
              messageId: 'suggestRemoveOptionalChain',
              desc: 'Remove unnecessary optional chain',
              output: `
type baz = () => { qux: () => {} };
declare const foo: { bar: { baz: baz } } | null;
foo?.bar?.baz().qux?.();
      `,
            },
          ],
        },
        {
          messageId: 'neverOptionalChain',
          message: 'Unnecessary optional chain on a non-nullish value.',
          line: 4,
          column: 22,
          endLine: 4,
          endColumn: 24,
          suggestions: [
            {
              messageId: 'suggestRemoveOptionalChain',
              desc: 'Remove unnecessary optional chain',
              output: `
type baz = () => { qux: () => {} };
declare const foo: { bar: { baz: baz } } | null;
foo?.bar?.baz?.().qux();
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
type baz = null | (() => { qux: () => {} });
declare const foo: { bar: { baz: baz } } | null;
foo?.bar?.baz?.().qux?.();
      `,
      errors: [
        {
          messageId: 'neverOptionalChain',
          message: 'Unnecessary optional chain on a non-nullish value.',
          line: 4,
          column: 9,
          endLine: 4,
          endColumn: 11,
          suggestions: [
            {
              messageId: 'suggestRemoveOptionalChain',
              desc: 'Remove unnecessary optional chain',
              output: `
type baz = null | (() => { qux: () => {} });
declare const foo: { bar: { baz: baz } } | null;
foo?.bar.baz?.().qux?.();
      `,
            },
          ],
        },
        {
          messageId: 'neverOptionalChain',
          message: 'Unnecessary optional chain on a non-nullish value.',
          line: 4,
          column: 22,
          endLine: 4,
          endColumn: 24,
          suggestions: [
            {
              messageId: 'suggestRemoveOptionalChain',
              desc: 'Remove unnecessary optional chain',
              output: `
type baz = null | (() => { qux: () => {} });
declare const foo: { bar: { baz: baz } } | null;
foo?.bar?.baz?.().qux();
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
type baz = null | (() => { qux: () => {} } | null);
declare const foo: { bar: { baz: baz } } | null;
foo?.bar?.baz?.()?.qux?.();
      `,
      errors: [
        {
          messageId: 'neverOptionalChain',
          message: 'Unnecessary optional chain on a non-nullish value.',
          line: 4,
          column: 9,
          endLine: 4,
          endColumn: 11,
          suggestions: [
            {
              messageId: 'suggestRemoveOptionalChain',
              desc: 'Remove unnecessary optional chain',
              output: `
type baz = null | (() => { qux: () => {} } | null);
declare const foo: { bar: { baz: baz } } | null;
foo?.bar.baz?.()?.qux?.();
      `,
            },
          ],
        },
        {
          messageId: 'neverOptionalChain',
          message: 'Unnecessary optional chain on a non-nullish value.',
          line: 4,
          column: 23,
          endLine: 4,
          endColumn: 25,
          suggestions: [
            {
              messageId: 'suggestRemoveOptionalChain',
              desc: 'Remove unnecessary optional chain',
              output: `
type baz = null | (() => { qux: () => {} } | null);
declare const foo: { bar: { baz: baz } } | null;
foo?.bar?.baz?.()?.qux();
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
type Foo = { baz: number };
type Bar = { baz: null | string | { qux: string } };
declare const foo: { fooOrBar: Foo | Bar } | null;
foo?.fooOrBar?.baz?.qux;
      `,
      errors: [
        {
          messageId: 'neverOptionalChain',
          message: 'Unnecessary optional chain on a non-nullish value.',
          line: 5,
          column: 14,
          endLine: 5,
          endColumn: 16,
          suggestions: [
            {
              messageId: 'suggestRemoveOptionalChain',
              desc: 'Remove unnecessary optional chain',
              output: `
type Foo = { baz: number };
type Bar = { baz: null | string | { qux: string } };
declare const foo: { fooOrBar: Foo | Bar } | null;
foo?.fooOrBar.baz?.qux;
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
declare const x: { a: { b: number } }[];
x[0].a?.b;
      `,
      errors: [
        {
          messageId: 'neverOptionalChain',
          message: 'Unnecessary optional chain on a non-nullish value.',
          line: 3,
          column: 7,
          endLine: 3,
          endColumn: 9,
          suggestions: [
            {
              messageId: 'suggestRemoveOptionalChain',
              desc: 'Remove unnecessary optional chain',
              output: `
declare const x: { a: { b: number } }[];
x[0].a.b;
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
type Foo = { [key: string]: string; foo: 'foo'; bar: 'bar' } | null;
type Key = 'bar' | 'foo';
declare const foo: Foo;
declare const key: Key;

foo?.[key]?.trim();
      `,
      errors: [
        {
          messageId: 'neverOptionalChain',
          message: 'Unnecessary optional chain on a non-nullish value.',
          line: 7,
          column: 11,
          endLine: 7,
          endColumn: 13,
          suggestions: [
            {
              messageId: 'suggestRemoveOptionalChain',
              desc: 'Remove unnecessary optional chain',
              output: `
type Foo = { [key: string]: string; foo: 'foo'; bar: 'bar' } | null;
type Key = 'bar' | 'foo';
declare const foo: Foo;
declare const key: Key;

foo?.[key].trim();
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
type Foo = { [key: string]: string; foo: 'foo'; bar: 'bar' } | null;
declare const foo: Foo;
const key = 'bar';
foo?.[key]?.trim();
      `,
      errors: [
        {
          messageId: 'neverOptionalChain',
          message: 'Unnecessary optional chain on a non-nullish value.',
          line: 5,
          column: 11,
          endLine: 5,
          endColumn: 13,
          suggestions: [
            {
              messageId: 'suggestRemoveOptionalChain',
              desc: 'Remove unnecessary optional chain',
              output: `
type Foo = { [key: string]: string; foo: 'foo'; bar: 'bar' } | null;
declare const foo: Foo;
const key = 'bar';
foo?.[key].trim();
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
interface Outer {
  inner?: {
    [key: string]: string | undefined;
    bar: 'bar';
  };
}

export function test(outer: Outer): number | undefined {
  const key = 'bar';
  return outer.inner?.[key]?.charCodeAt(0);
}
      `,
      errors: [
        {
          messageId: 'neverOptionalChain',
          message: 'Unnecessary optional chain on a non-nullish value.',
          line: 11,
          column: 28,
          endLine: 11,
          endColumn: 30,
          suggestions: [
            {
              messageId: 'suggestRemoveOptionalChain',
              desc: 'Remove unnecessary optional chain',
              output: `
interface Outer {
  inner?: {
    [key: string]: string | undefined;
    bar: 'bar';
  };
}

export function test(outer: Outer): number | undefined {
  const key = 'bar';
  return outer.inner?.[key].charCodeAt(0);
}
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
interface Outer {
  inner?: {
    [key: string]: string | undefined;
    bar: 'bar';
  };
}
type Bar = 'bar';

function Foo(outer: Outer, key: Bar): number | undefined {
  return outer.inner?.[key]?.charCodeAt(0);
}
      `,
      errors: [
        {
          messageId: 'neverOptionalChain',
          message: 'Unnecessary optional chain on a non-nullish value.',
          line: 11,
          column: 28,
          endLine: 11,
          endColumn: 30,
          suggestions: [
            {
              messageId: 'suggestRemoveOptionalChain',
              desc: 'Remove unnecessary optional chain',
              output: `
interface Outer {
  inner?: {
    [key: string]: string | undefined;
    bar: 'bar';
  };
}
type Bar = 'bar';

function Foo(outer: Outer, key: Bar): number | undefined {
  return outer.inner?.[key].charCodeAt(0);
}
      `,
            },
          ],
        },
      ],
    },
    // https://github.com/typescript-eslint/typescript-eslint/issues/2384
    {
      code: `
function test(testVal?: true) {
  if (testVal ?? true) {
    console.log('test');
  }
}
      `,
      errors: [
        {
          messageId: 'alwaysTruthy',
          message: 'Unnecessary conditional, value is always truthy.',
          line: 3,
          column: 7,
          endLine: 3,
          endColumn: 22,
        },
      ],
    },
    // https://github.com/typescript-eslint/typescript-eslint/issues/2255
    {
      code: `
const a = null;
if (!a) {
}
      `,
      errors: [
        {
          messageId: 'alwaysTruthy',
          message: 'Unnecessary conditional, value is always truthy.',
          line: 3,
          column: 5,
          endLine: 3,
          endColumn: 7,
        },
      ],
    },
    {
      code: `
const a = true;
if (!a) {
}
      `,
      errors: [
        {
          messageId: 'alwaysFalsy',
          message: 'Unnecessary conditional, value is always falsy.',
          line: 3,
          column: 5,
          endLine: 3,
          endColumn: 7,
        },
      ],
    },
    {
      code: `
function sayHi(): void {
  console.log('Hi!');
}

let speech: never = sayHi();
if (!speech) {
}
      `,
      errors: [
        {
          messageId: 'never',
          message: 'Unnecessary conditional, value is `never`.',
          line: 7,
          column: 5,
          endLine: 7,
          endColumn: 12,
        },
      ],
    },
    {
      code: `
declare const x: string[] | null;
if (x) {
}
      `,
      languageOptions: {
        parserOptions: {
          project: './tsconfig.unstrict.json',
        },
      },
      errors: [
        {
          messageId: 'noStrictNullCheck',
          message:
            'This rule requires the `strictNullChecks` compiler option to be turned on to function correctly.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 1,
        },
        {
          messageId: 'alwaysTruthy',
          message: 'Unnecessary conditional, value is always truthy.',
          line: 3,
          column: 5,
          endLine: 3,
          endColumn: 6,
        },
      ],
    },
    {
      code: `
interface Foo {
  test: string;
  [key: string]: [string] | undefined;
}

type OptionalFoo = Foo | undefined;
declare const foo: OptionalFoo;
foo?.test?.length;
      `,
      errors: [
        {
          messageId: 'neverOptionalChain',
          message: 'Unnecessary optional chain on a non-nullish value.',
          line: 9,
          column: 10,
          endLine: 9,
          endColumn: 12,
          suggestions: [
            {
              messageId: 'suggestRemoveOptionalChain',
              desc: 'Remove unnecessary optional chain',
              output: `
interface Foo {
  test: string;
  [key: string]: [string] | undefined;
}

type OptionalFoo = Foo | undefined;
declare const foo: OptionalFoo;
foo?.test.length;
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
function pick<Obj extends Record<string, 1 | 2 | 3>, Key extends keyof Obj>(
  obj: Obj,
  key: Key,
): Obj[Key] {
  const k = obj[key];
  if (obj[key]) {
    return obj[key];
  }
  throw new Error('Boom!');
}

pick({ foo: 1, bar: 2 }, 'bar');
      `,
      errors: [
        {
          messageId: 'alwaysTruthy',
          message: 'Unnecessary conditional, value is always truthy.',
          line: 7,
          column: 7,
          endLine: 7,
          endColumn: 15,
        },
      ],
    },
    {
      code: `
function getElem(dict: Record<string, { foo: string }>, key: string) {
  if (dict[key]) {
    return dict[key].foo;
  } else {
    return '';
  }
}
      `,
      errors: [
        {
          messageId: 'alwaysTruthy',
          message: 'Unnecessary conditional, value is always truthy.',
          line: 3,
          column: 7,
          endLine: 3,
          endColumn: 16,
        },
      ],
    },
    {
      code: `
declare let foo: {};
foo ??= 1;
      `,
      errors: [
        {
          messageId: 'neverNullish',
          message:
            'Unnecessary conditional, expected left-hand side of `??` operator to be possibly null or undefined.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 4,
        },
      ],
    },
    {
      code: `
declare let foo: number;
foo ??= 1;
      `,
      errors: [
        {
          messageId: 'neverNullish',
          message:
            'Unnecessary conditional, expected left-hand side of `??` operator to be possibly null or undefined.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 4,
        },
      ],
    },
    {
      code: `
declare let foo: null;
foo ??= null;
      `,
      errors: [
        {
          messageId: 'alwaysNullish',
          message:
            'Unnecessary conditional, left-hand side of `??` operator is always `null` or `undefined`.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 4,
        },
      ],
    },
    {
      code: `
declare let foo: {};
foo ||= 1;
      `,
      errors: [
        {
          messageId: 'alwaysTruthy',
          message: 'Unnecessary conditional, value is always truthy.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 4,
        },
      ],
    },
    {
      code: `
declare let foo: null;
foo ||= null;
      `,
      errors: [
        {
          messageId: 'alwaysFalsy',
          message: 'Unnecessary conditional, value is always falsy.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 4,
        },
      ],
    },
    {
      code: `
declare let foo: {};
foo &&= 1;
      `,
      errors: [
        {
          messageId: 'alwaysTruthy',
          message: 'Unnecessary conditional, value is always truthy.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 4,
        },
      ],
    },
    {
      code: `
declare let foo: null;
foo &&= null;
      `,
      errors: [
        {
          messageId: 'alwaysFalsy',
          message: 'Unnecessary conditional, value is always falsy.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 4,
        },
      ],
    },
    {
      code: `
declare const foo: { bar: number };
foo.bar ??= 1;
      `,
      languageOptions: {
        parserOptions: {
          project: './tsconfig.exactOptionalPropertyTypes.json',
        },
      },
      errors: [
        {
          messageId: 'neverNullish',
          message:
            'Unnecessary conditional, expected left-hand side of `??` operator to be possibly null or undefined.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 8,
        },
      ],
    },
    {
      code: `
type Foo = { bar: () => number } | null;
declare const foo: Foo;
foo?.bar()?.toExponential();
      `,
      errors: [
        {
          messageId: 'neverOptionalChain',
          message: 'Unnecessary optional chain on a non-nullish value.',
          line: 4,
          column: 11,
          endLine: 4,
          endColumn: 13,
          suggestions: [
            {
              messageId: 'suggestRemoveOptionalChain',
              desc: 'Remove unnecessary optional chain',
              output: `
type Foo = { bar: () => number } | null;
declare const foo: Foo;
foo?.bar().toExponential();
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
type Foo = { bar: null | { baz: () => { qux: number } } } | null;
declare const foo: Foo;
foo?.bar?.baz()?.qux?.toExponential();
      `,
      errors: [
        {
          messageId: 'neverOptionalChain',
          message: 'Unnecessary optional chain on a non-nullish value.',
          line: 4,
          column: 16,
          endLine: 4,
          endColumn: 18,
          suggestions: [
            {
              messageId: 'suggestRemoveOptionalChain',
              desc: 'Remove unnecessary optional chain',
              output: `
type Foo = { bar: null | { baz: () => { qux: number } } } | null;
declare const foo: Foo;
foo?.bar?.baz().qux?.toExponential();
      `,
            },
          ],
        },
        {
          messageId: 'neverOptionalChain',
          message: 'Unnecessary optional chain on a non-nullish value.',
          line: 4,
          column: 21,
          endLine: 4,
          endColumn: 23,
          suggestions: [
            {
              messageId: 'suggestRemoveOptionalChain',
              desc: 'Remove unnecessary optional chain',
              output: `
type Foo = { bar: null | { baz: () => { qux: number } } } | null;
declare const foo: Foo;
foo?.bar?.baz()?.qux.toExponential();
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
type Foo = (() => number) | null;
declare const foo: Foo;
foo?.()?.toExponential();
      `,
      errors: [
        {
          messageId: 'neverOptionalChain',
          message: 'Unnecessary optional chain on a non-nullish value.',
          line: 4,
          column: 8,
          endLine: 4,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestRemoveOptionalChain',
              desc: 'Remove unnecessary optional chain',
              output: `
type Foo = (() => number) | null;
declare const foo: Foo;
foo?.().toExponential();
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
type Foo = { [key: string]: () => number } | null;
declare const foo: Foo;
foo?.['bar']()?.toExponential();
      `,
      errors: [
        {
          messageId: 'neverOptionalChain',
          message: 'Unnecessary optional chain on a non-nullish value.',
          line: 4,
          column: 15,
          endLine: 4,
          endColumn: 17,
          suggestions: [
            {
              messageId: 'suggestRemoveOptionalChain',
              desc: 'Remove unnecessary optional chain',
              output: `
type Foo = { [key: string]: () => number } | null;
declare const foo: Foo;
foo?.['bar']().toExponential();
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
type Foo = { [key: string]: () => number } | null;
declare const foo: Foo;
foo?.['bar']?.()?.toExponential();
      `,
      errors: [
        {
          messageId: 'neverOptionalChain',
          message: 'Unnecessary optional chain on a non-nullish value.',
          line: 4,
          column: 17,
          endLine: 4,
          endColumn: 19,
          suggestions: [
            {
              messageId: 'suggestRemoveOptionalChain',
              desc: 'Remove unnecessary optional chain',
              output: `
type Foo = { [key: string]: () => number } | null;
declare const foo: Foo;
foo?.['bar']?.().toExponential();
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
        const a = true;
        if (!!a) {
        }
      `,
      errors: [
        {
          messageId: 'alwaysTruthy',
          message: 'Unnecessary conditional, value is always truthy.',
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 16,
        },
      ],
    },
    {
      code: `
declare function assert(x: unknown): asserts x;
assert(true);
      `,
      options: [
        {
          checkTypePredicates: true,
        },
      ],
      errors: [
        {
          messageId: 'alwaysTruthy',
          message: 'Unnecessary conditional, value is always truthy.',
          line: 3,
          column: 8,
          endLine: 3,
          endColumn: 12,
        },
      ],
    },
    {
      code: `
declare function assert(x: unknown): asserts x;
assert(false);
      `,
      options: [
        {
          checkTypePredicates: true,
        },
      ],
      errors: [
        {
          messageId: 'alwaysFalsy',
          message: 'Unnecessary conditional, value is always falsy.',
          line: 3,
          column: 8,
          endLine: 3,
          endColumn: 13,
        },
      ],
    },
    {
      code: `
declare function assert(x: unknown, y: unknown): asserts x;

assert(true, Math.random() > 0.5);
      `,
      options: [
        {
          checkTypePredicates: true,
        },
      ],
      errors: [
        {
          messageId: 'alwaysTruthy',
          message: 'Unnecessary conditional, value is always truthy.',
          line: 4,
          column: 8,
          endLine: 4,
          endColumn: 12,
        },
      ],
    },
    {
      code: `
declare function assert(x: unknown): asserts x;
assert({});
      `,
      options: [
        {
          checkTypePredicates: true,
        },
      ],
      errors: [
        {
          messageId: 'alwaysTruthy',
          message: 'Unnecessary conditional, value is always truthy.',
          line: 3,
          column: 8,
          endLine: 3,
          endColumn: 10,
        },
      ],
    },
    {
      code: `
declare function assertsString(x: unknown): asserts x is string;
declare const a: string;
assertsString(a);
      `,
      options: [
        {
          checkTypePredicates: true,
        },
      ],
      errors: [
        {
          messageId: 'typeGuardAlreadyIsType',
          message:
            'Unnecessary conditional, expression already has the type being checked by the assertion function.',
          line: 4,
          column: 15,
          endLine: 4,
          endColumn: 16,
        },
      ],
    },
    {
      code: `
declare function isString(x: unknown): x is string;
declare const a: string;
isString(a);
      `,
      options: [
        {
          checkTypePredicates: true,
        },
      ],
      errors: [
        {
          messageId: 'typeGuardAlreadyIsType',
          message:
            'Unnecessary conditional, expression already has the type being checked by the type guard.',
          line: 4,
          column: 10,
          endLine: 4,
          endColumn: 11,
        },
      ],
    },
    {
      code: `
declare function isString(x: unknown): x is string;
declare const a: string;
isString('fa' + 'lafel');
      `,
      options: [
        {
          checkTypePredicates: true,
        },
      ],
      errors: [
        {
          messageId: 'typeGuardAlreadyIsType',
          message:
            'Unnecessary conditional, expression already has the type being checked by the type guard.',
          line: 4,
          column: 10,
          endLine: 4,
          endColumn: 24,
        },
      ],
    },
    // string is a strict subtype of string | number.
    {
      code: `
declare function isStringOrNumber(x: unknown): x is string | number;
declare const s: string;
if (isStringOrNumber(s)) {
}
      `,
      options: [
        {
          checkTypePredicates: true,
        },
      ],
      errors: [
        {
          messageId: 'typeGuardAlreadyIsType',
          message:
            'Unnecessary conditional, expression already has the type being checked by the type guard.',
          line: 4,
          column: 22,
          endLine: 4,
          endColumn: 23,
        },
      ],
    },
    {
      code: `
interface Wider {
  a: string;
}
interface Narrower {
  a: string;
  b?: number;
}
declare function isWider(x: unknown): x is Wider;
declare const n: Narrower;
if (isWider(n)) {
}
      `,
      options: [
        {
          checkTypePredicates: true,
        },
      ],
      errors: [
        {
          messageId: 'typeGuardAlreadyIsType',
          message:
            'Unnecessary conditional, expression already has the type being checked by the type guard.',
          line: 11,
          column: 13,
          endLine: 11,
          endColumn: 14,
        },
      ],
    },
    // Mutually assignable types: Wider is assignable to Narrower (b is
    // optional), so the type guard condition is always true.
    {
      code: `
interface Wider {
  a: string;
}
interface Narrower {
  a: string;
  b?: number;
}
declare function isNarrower(x: unknown): x is Narrower;
declare const w: Wider;
if (isNarrower(w)) {
}
      `,
      options: [
        {
          checkTypePredicates: true,
        },
      ],
      errors: [
        {
          messageId: 'typeGuardAlreadyIsType',
          message:
            'Unnecessary conditional, expression already has the type being checked by the type guard.',
          line: 11,
          column: 16,
          endLine: 11,
          endColumn: 17,
        },
      ],
    },
    // "branded" types
    unnecessaryConditionTest('"" & {}', 'alwaysFalsy'),
    unnecessaryConditionTest('"" & { __brand: string }', 'alwaysFalsy'),
    unnecessaryConditionTest(
      '("" | false) & { __brand: string }',
      'alwaysFalsy',
    ),
    unnecessaryConditionTest(
      '((string & { __brandA: string }) | (number & { __brandB: string })) & ""',
      'alwaysFalsy',
    ),
    unnecessaryConditionTest(
      '("foo" | "bar") & { __brand: string }',
      'alwaysTruthy',
    ),
    unnecessaryConditionTest(
      '(123 | true) & { __brand: string }',
      'alwaysTruthy',
    ),
    unnecessaryConditionTest(
      '(string | number) & ("foo" | 123) & { __brand: string }',
      'alwaysTruthy',
    ),
    unnecessaryConditionTest(
      '((string & { __brandA: string }) | (number & { __brandB: string })) & "foo"',
      'alwaysTruthy',
    ),
    unnecessaryConditionTest(
      '((string & { __brandA: string }) | (number & { __brandB: string })) & ("foo" | 123)',
      'alwaysTruthy',
    ),
    {
      code: `
type A = {
  [name in Lowercase<string>]?: {
    [name in Lowercase<string>]: {
      a: 1;
    };
  };
};

declare const a: A;

a.a?.a?.a;
      `,
      errors: [
        {
          messageId: 'neverOptionalChain',
          message: 'Unnecessary optional chain on a non-nullish value.',
          line: 12,
          column: 7,
          endLine: 12,
          endColumn: 9,
          suggestions: [
            {
              messageId: 'suggestRemoveOptionalChain',
              desc: 'Remove unnecessary optional chain',
              output: `
type A = {
  [name in Lowercase<string>]?: {
    [name in Lowercase<string>]: {
      a: 1;
    };
  };
};

declare const a: A;

a.a?.a.a;
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
interface T {
  [name: Lowercase<string>]: {
    [name: Lowercase<string>]: {
      [name: Lowercase<string>]: {
        value: 'value';
      };
    };
  };
  [name: Uppercase<string>]: null | {
    [name: Uppercase<string>]: null | {
      [name: Uppercase<string>]: null | {
        VALUE: 'VALUE';
      };
    };
  };
}

declare const t: T;

t.a?.a?.a?.value;
      `,
      errors: [
        {
          messageId: 'neverOptionalChain',
          message: 'Unnecessary optional chain on a non-nullish value.',
          line: 21,
          column: 4,
          endLine: 21,
          endColumn: 6,
          suggestions: [
            {
              messageId: 'suggestRemoveOptionalChain',
              desc: 'Remove unnecessary optional chain',
              output: `
interface T {
  [name: Lowercase<string>]: {
    [name: Lowercase<string>]: {
      [name: Lowercase<string>]: {
        value: 'value';
      };
    };
  };
  [name: Uppercase<string>]: null | {
    [name: Uppercase<string>]: null | {
      [name: Uppercase<string>]: null | {
        VALUE: 'VALUE';
      };
    };
  };
}

declare const t: T;

t.a.a?.a?.value;
      `,
            },
          ],
        },
        {
          messageId: 'neverOptionalChain',
          message: 'Unnecessary optional chain on a non-nullish value.',
          line: 21,
          column: 7,
          endLine: 21,
          endColumn: 9,
          suggestions: [
            {
              messageId: 'suggestRemoveOptionalChain',
              desc: 'Remove unnecessary optional chain',
              output: `
interface T {
  [name: Lowercase<string>]: {
    [name: Lowercase<string>]: {
      [name: Lowercase<string>]: {
        value: 'value';
      };
    };
  };
  [name: Uppercase<string>]: null | {
    [name: Uppercase<string>]: null | {
      [name: Uppercase<string>]: null | {
        VALUE: 'VALUE';
      };
    };
  };
}

declare const t: T;

t.a?.a.a?.value;
      `,
            },
          ],
        },
        {
          messageId: 'neverOptionalChain',
          message: 'Unnecessary optional chain on a non-nullish value.',
          line: 21,
          column: 10,
          endLine: 21,
          endColumn: 12,
          suggestions: [
            {
              messageId: 'suggestRemoveOptionalChain',
              desc: 'Remove unnecessary optional chain',
              output: `
interface T {
  [name: Lowercase<string>]: {
    [name: Lowercase<string>]: {
      [name: Lowercase<string>]: {
        value: 'value';
      };
    };
  };
  [name: Uppercase<string>]: null | {
    [name: Uppercase<string>]: null | {
      [name: Uppercase<string>]: null | {
        VALUE: 'VALUE';
      };
    };
  };
}

declare const t: T;

t.a?.a?.a.value;
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
declare const test: Array<{ a?: string }>;

if (test[0]?.a) {
  test[0]?.a;
}
      `,
      languageOptions: {
        parserOptions: {
          project: './tsconfig.noUncheckedIndexedAccess.json',
          projectService: false,
        },
      },
      errors: [
        {
          messageId: 'neverOptionalChain',
          message: 'Unnecessary optional chain on a non-nullish value.',
          line: 5,
          column: 10,
          endLine: 5,
          endColumn: 12,
          suggestions: [
            {
              messageId: 'suggestRemoveOptionalChain',
              desc: 'Remove unnecessary optional chain',
              output: `
declare const test: Array<{ a?: string }>;

if (test[0]?.a) {
  test[0].a;
}
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
declare const arr2: Array<{ x: { y: { z: object } } }>;
arr2[42]?.x?.y?.z;
      `,
      languageOptions: {
        parserOptions: {
          project: './tsconfig.noUncheckedIndexedAccess.json',
          projectService: false,
        },
      },
      errors: [
        {
          messageId: 'neverOptionalChain',
          message: 'Unnecessary optional chain on a non-nullish value.',
          line: 3,
          column: 12,
          endLine: 3,
          endColumn: 14,
          suggestions: [
            {
              messageId: 'suggestRemoveOptionalChain',
              desc: 'Remove unnecessary optional chain',
              output: `
declare const arr2: Array<{ x: { y: { z: object } } }>;
arr2[42]?.x.y?.z;
      `,
            },
          ],
        },
        {
          messageId: 'neverOptionalChain',
          message: 'Unnecessary optional chain on a non-nullish value.',
          line: 3,
          column: 15,
          endLine: 3,
          endColumn: 17,
          suggestions: [
            {
              messageId: 'suggestRemoveOptionalChain',
              desc: 'Remove unnecessary optional chain',
              output: `
declare const arr2: Array<{ x: { y: { z: object } } }>;
arr2[42]?.x?.y.z;
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
declare const arr: string[];

if (arr[0]) {
  arr[0] ?? 'foo';
}
      `,
      languageOptions: {
        parserOptions: {
          project: './tsconfig.noUncheckedIndexedAccess.json',
          projectService: false,
        },
      },
      errors: [
        {
          messageId: 'neverNullish',
          message:
            'Unnecessary conditional, expected left-hand side of `??` operator to be possibly null or undefined.',
          line: 5,
          column: 3,
          endLine: 5,
          endColumn: 9,
        },
      ],
    },
    {
      code: `
declare const arr: object[];

if (arr[42] && arr[42]) {
}
      `,
      languageOptions: {
        parserOptions: {
          project: './tsconfig.noUncheckedIndexedAccess.json',
          projectService: false,
        },
      },
      errors: [
        {
          messageId: 'alwaysTruthy',
          message: 'Unnecessary conditional, value is always truthy.',
          line: 4,
          column: 16,
          endLine: 4,
          endColumn: 23,
        },
      ],
    },
    {
      code: `
declare function isString(x: unknown): x is string;
declare const a: string;
isString(a);
      `,
      options: [
        {
          checkTypePredicates: true,
        },
      ],
      errors: [
        {
          messageId: 'typeGuardAlreadyIsType',
          message:
            'Unnecessary conditional, expression already has the type being checked by the type guard.',
          line: 4,
          column: 10,
          endLine: 4,
          endColumn: 11,
        },
      ],
    },
    {
      code: `
declare function assertsString(x: unknown): asserts x is string;
declare const a: string;
assertsString(a);
      `,
      options: [
        {
          checkTypePredicates: true,
        },
      ],
      errors: [
        {
          messageId: 'typeGuardAlreadyIsType',
          message:
            'Unnecessary conditional, expression already has the type being checked by the assertion function.',
          line: 4,
          column: 15,
          endLine: 4,
          endColumn: 16,
        },
      ],
    },
    // alwaysTruthy / alwaysFalsy in array-predicate callback body: when the
    // body is an expression or a single-return BlockStatement, the rule
    // descends into it and emits alwaysTruthy/alwaysFalsy at the body
    // expression. Verified against upstream v8 CLI behavior. Locks down all
    // standard array-predicate methods recognized by isArrayPredicateFunction.
    {
      code: `
[1, 2, 3].filter(() => 'truthy-string');
      `,
      errors: [
        {
          messageId: 'alwaysTruthy',
          message: 'Unnecessary conditional, value is always truthy.',
          line: 2,
          column: 24,
          endLine: 2,
          endColumn: 39,
        },
      ],
    },
    {
      code: `
[1, 2, 3].find(function () {
  return 0;
});
      `,
      errors: [
        {
          messageId: 'alwaysFalsy',
          message: 'Unnecessary conditional, value is always falsy.',
          line: 3,
          column: 10,
          endLine: 3,
          endColumn: 11,
        },
      ],
    },
    {
      code: `
[1, 2, 3].some(() => 1);
      `,
      errors: [
        {
          messageId: 'alwaysTruthy',
          message: 'Unnecessary conditional, value is always truthy.',
          line: 2,
          column: 22,
          endLine: 2,
          endColumn: 23,
        },
      ],
    },
    {
      code: `
[1, 2, 3].every(() => null);
      `,
      errors: [
        {
          messageId: 'alwaysFalsy',
          message: 'Unnecessary conditional, value is always falsy.',
          line: 2,
          column: 23,
          endLine: 2,
          endColumn: 27,
        },
      ],
    },
    {
      code: `
[1, 2, 3].findIndex(function namedFalsy() {
  return undefined;
});
      `,
      errors: [
        {
          messageId: 'alwaysFalsy',
          message: 'Unnecessary conditional, value is always falsy.',
          line: 3,
          column: 10,
          endLine: 3,
          endColumn: 19,
        },
      ],
    },
    {
      code: `
[1, 2, 3].findLast(() => 'always');
      `,
      errors: [
        {
          messageId: 'alwaysTruthy',
          message: 'Unnecessary conditional, value is always truthy.',
          line: 2,
          column: 26,
          endLine: 2,
          endColumn: 34,
        },
      ],
    },
  ],
});
