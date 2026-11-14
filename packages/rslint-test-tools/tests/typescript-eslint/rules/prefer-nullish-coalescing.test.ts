// Upstream: @typescript-eslint/eslint-plugin v8.72.2.
// All 286 valid and 351 invalid cases preserve their input and rule options.
// Messages, complete ranges and suggestion outputs were recorded by running the
// upstream suite with that exact version; the rule has no autofixes.
// The upstream `unstrict` fixture directory is reproduced as `unstrict/`.
// https://github.com/typescript-eslint/typescript-eslint/blob/v8.72.2/packages/eslint-plugin/tests/rules/prefer-nullish-coalescing.test.ts
import { RuleTester } from '@rslint/rule-tester';

const compilerOptions = {
  jsx: 'preserve',
  target: 'es2015',
  module: 'commonjs',
  lib: ['es2015', 'es2017', 'esnext'],
  experimentalDecorators: true,
};

const ruleTester = new RuleTester({
  plugin: '@typescript-eslint',
  filename: 'file.ts',
  fixtureFiles: {
    'tsconfig.json': JSON.stringify({
      compilerOptions: { ...compilerOptions, strict: true, esModuleInterop: true },
      include: ['file.ts'],
    }),
    'unstrict/tsconfig.json': JSON.stringify({
      compilerOptions: { ...compilerOptions, strict: false },
      include: ['file.ts'],
    }),
  },
  languageOptions: { parserOptions: { project: './tsconfig.json' } },
});

ruleTester.run('prefer-nullish-coalescing', {
  valid: [
    `
declare let x: string;
x || 'foo';
    `,
    `
declare let x: number;
x || 'foo';
    `,
    `
declare let x: boolean;
x || 'foo';
    `,
    `
declare let x: object;
x || 'foo';
    `,
    `
declare let x: string;
x ||= 'foo';
    `,
    `
declare let x: number;
x ||= 'foo';
    `,
    `
declare let x: boolean;
x ||= 'foo';
    `,
    `
declare let x: object;
x ||= 'foo';
    `,
    `
declare let x: string | null | undefined;
x ?? 'foo';
    `,
    `
declare let x: string | null | undefined;
x ??= 'foo';
    `,
    `
declare let x: number | null | undefined;
x ?? 'foo';
    `,
    `
declare let x: number | null | undefined;
x ??= 'foo';
    `,
    `
declare let x: boolean | null | undefined;
x ?? 'foo';
    `,
    `
declare let x: boolean | null | undefined;
x ??= 'foo';
    `,
    `
declare let x: object | null | undefined;
x ?? 'foo';
    `,
    `
declare let x: object | null | undefined;
x ??= 'foo';
    `,

    {
      code: 'x !== undefined && x !== null ? x : y;',
      options: [{ ignoreTernaryTests: true }],
    },

    {
      code: "x !== undefined && x !== null ? 'foo' : 'bar';",
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: 'x !== null && x !== undefined && x !== 5 ? x : y;',
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: 'x === null || x === undefined || x === 5 ? x : y;',
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: 'x === undefined && x !== null ? x : y;',
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: 'x === undefined && x === null ? x : y;',
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: 'x !== undefined && x === null ? x : y;',
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: 'x === undefined || x !== null ? x : y;',
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: 'x === undefined || x === null ? x : y;',
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: 'x !== undefined || x === null ? x : y;',
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: 'x !== undefined || x === null ? y : x;',
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: 'x === null || x === null ? y : x;',
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: 'x === undefined || x === undefined ? y : x;',
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: 'x == null ? x : y;',
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: 'undefined == null ? x : y;',
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: 'undefined != z ? x : y;',
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: 'x == undefined ? x : y;',
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: 'x != null ? y : x;',
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: 'x != undefined ? y : x;',
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: 'null == x ? x : y;',
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: 'undefined == x ? x : y;',
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: 'null != x ? y : x;',
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: 'undefined != x ? y : x;',
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: number | undefined;
x !== 15 && x !== undefined ? x : y;
      `,
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: number | undefined;
x !== undefined && x !== 15 ? x : y;
      `,
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: number | undefined;
15 !== x && undefined !== x ? x : y;
      `,
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: number | undefined;
undefined !== x && 15 !== x ? x : y;
      `,
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: number | undefined;
15 !== x && x !== undefined ? x : y;
      `,
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: number | undefined;
undefined !== x && x !== 15 ? x : y;
      `,
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: string | undefined;
x !== 'foo' && x !== undefined ? x : y;
      `,
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
function test(value: number | undefined): number {
  return value !== foo() && value !== undefined ? value : 1;
}
      `,
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
const test = (value: boolean | undefined): boolean =>
  value !== undefined && value !== false ? value : false;
      `,
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: string;
x === null ? x : y;
      `,
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: string | undefined | null;
x === null ? x : y;
      `,
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: string | undefined | null;
x === undefined ? x : y;
      `,
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: string | undefined | null;
x !== null ? x : y;
      `,
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: string | undefined | null;
x !== undefined ? x : y;
      `,
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
function test<T>(val: T | undefined) {
  return val !== undefined ? val : 'foo';
}
      `,
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
function test<T>(val: T | null) {
  return val !== null ? val : 'foo';
}
      `,
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
function test<T extends string | null>(val: T | undefined) {
  return val !== undefined ? val : 'foo';
}
      `,
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
class Property<T> {
  value: T | undefined;
  defaultValue!: T;

  get(): T {
    return this.value !== undefined ? this.value : this.defaultValue;
  }
}
      `,
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
function test<T>(val: T | undefined, fallback: T) {
  if (val === undefined) {
    val = fallback;
  }
  return val;
}
      `,
      options: [{ ignoreIfStatements: false }],
    },
    {
      code: `
function test<T extends string | undefined>(val: T | null) {
  return val !== null ? val : 'foo';
}
      `,
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
function test<T>(val: NoInfer<T> | undefined, fallback: T): T {
  if (val === undefined) {
    val = fallback;
  }
  return val;
}
      `,
      options: [{ ignoreIfStatements: false }],
    },
    {
      code: `
function test<T extends Record<string, null>, K extends keyof T>(
  val: T[K] | undefined,
) {
  return val !== undefined ? val : 'fallback';
}
      `,
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
function test<T>(val: T) {
  return val !== undefined ? val : 'fallback';
}
      `,
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
function test<T extends unknown>(val: T | undefined) {
  return val !== undefined ? val : 'fallback';
}
      `,
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
function test<T, U extends T>(val: U | undefined) {
  return val !== undefined ? val : 'fallback';
}
      `,
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: any;
x === null ? x : y;
      `,
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: unknown;
x === null ? x : y;
      `,
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: string;
x ? x : y;
      `,
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: string;
!x ? y : x;
      `,
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: string | object;
x ? x : y;
      `,
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: string | object;
!x ? y : x;
      `,
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: number;
x ? x : y;
      `,
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: number;
!x ? y : x;
      `,
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: bigint;
x ? x : y;
      `,
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: bigint;
!x ? y : x;
      `,
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: boolean;
x ? x : y;
      `,
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: boolean;
!x ? y : x;
      `,
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: object;
x ? x : y;
      `,
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: object;
!x ? y : x;
      `,
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: string[];
x ? x : y;
      `,
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: string[];
!x ? y : x;
      `,
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: Function;
x ? x : y;
      `,
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: Function;
!x ? y : x;
      `,
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: () => string;
x ? x : y;
      `,
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: () => string;
!x ? y : x;
      `,
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: () => string | null | undefined;
x ? x : y;
      `,
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: () => string | null | undefined;
!x ? y : x;
      `,
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: () => string | null;
x() ? x() : y;
      `,
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: () => string | null;
!x() ? y : x();
      `,
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
const a = 'foo';
declare let x: (a: string | null) => string | null;
x(a) ? x(a) : y;
      `,
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
const a = 'foo';
declare let x: (a: string | null) => string | null;
!x(a) ? y : x(a);
      `,
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: { n: string };
x.n ? x.n : y;
      `,
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: { n: string };
!x.n ? y : x.n;
      `,
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: { n: string | object };
x.n ? x.n : y;
      `,
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: { n: string | object };
!x.n ? y : x.n;
      `,
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: { n: number };
x.n ? x.n : y;
      `,
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: { n: number };
!x.n ? y : x.n;
      `,
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: { n: bigint };
x.n ? x.n : y;
      `,
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: { n: bigint };
!x.n ? y : x.n;
      `,
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: { n: boolean };
x.n ? x.n : y;
      `,
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: { n: boolean };
!x.n ? y : x.n;
      `,
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: { n: object };
x ? x : y;
      `,
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: { n: object };
!x ? y : x;
      `,
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: { n: string[] };
x.n ? x.n : y;
      `,
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: { n: string[] };
!x.n ? y : x.n;
      `,
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: { n: Function };
x.n ? x.n : y;
      `,
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: { n: Function };
!x.n ? y : x.n;
      `,
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: { n: () => string };
x.n ? x.n : y;
      `,
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: { n: () => string };
!x.n ? y : x.n;
      `,
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: { n: () => string | null | undefined };
x.n ? x.n : y;
      `,
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: { n: () => string | null | undefined };
!x.n ? y : x.n;
      `,
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let foo: string;
declare function makeFoo(): string;

function lazyInitialize() {
  if (!foo) {
    foo = makeFoo();
  }
}
      `,
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (foo) {
    foo = makeFoo();
  }
}
      `,
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (foo != null) {
    foo = makeFoo();
  }
}
      `,
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (foo == null) {
    foo = makeFoo();
    return foo;
  }
}
      `,
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (foo == null) {
    foo = makeFoo();
  } else {
    return 'bar';
  }
}
      `,
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (foo == null) {
    foo = makeFoo();
  } else if (foo.a) {
    return 'bar';
  }
}
      `,
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };
function shadowed() {
  if (foo == null) {
    const foo = makeFoo();
  }
}
      `,
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let foo: { foo: string } | null;
declare function makeFoo(): { foo: { foo: string } };
function weirdDestructuringAssignment() {
  if (foo == null) {
    ({ foo } = makeFoo());
  }
}
      `,
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare const nullOrObject: null | { a: string };

const test = nullOrObject !== undefined && null !== null ? nullOrObject : 42;
      `,
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare const nullOrObject: null | { a: string };

const test = nullOrObject !== undefined && null != null ? nullOrObject : 42;
      `,
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare const nullOrObject: null | { a: string };

const test =
  nullOrObject !== undefined && null != undefined ? nullOrObject : 42;
      `,
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare const nullOrObject: null | { a: string };

const test = nullOrObject === undefined || null === null ? 42 : nullOrObject;
      `,
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare const nullOrObject: null | { a: string };

const test = nullOrObject === undefined || null == null ? 42 : nullOrObject;
      `,
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare const nullOrObject: null | { a: string };

const test =
  nullOrObject === undefined || null == undefined ? 42 : nullOrObject;
      `,
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
const a = 'b';
declare let x: { a: string; b: string } | null;

x?.a != null ? x[a] : 'foo';
      `,
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
const a = 'b';
declare let x: { a: string; b: string } | null;

x?.[a] != null ? x.a : 'foo';
      `,
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: { a: string } | null;
declare let y: { a: string } | null;

x?.a ? y?.a : 'foo';
      `,
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (!foo) {
    foo = makeFoo();
  }
}
      `,
      options: [{ ignoreIfStatements: true }],
    },
    {
      code: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (!foo) foo = makeFoo();
}
      `,
      options: [{ ignoreIfStatements: true }],
    },

    // ignoreConditionalTests
    `
declare let x: string | null | undefined;
x || 'foo' ? null : null;
    `,
    `
declare let x: string | null | undefined;
(x ||= 'foo') ? null : null;
    `,
    `
declare let x: number | null | undefined;
x || 'foo' ? null : null;
    `,
    `
declare let x: number | null | undefined;
(x ||= 'foo') ? null : null;
    `,
    `
declare let x: boolean | null | undefined;
x || 'foo' ? null : null;
    `,
    `
declare let x: boolean | null | undefined;
(x ||= 'foo') ? null : null;
    `,
    `
declare let x: object | null | undefined;
x || 'foo' ? null : null;
    `,
    `
declare let x: object | null | undefined;
(x ||= 'foo') ? null : null;
    `,
    `
declare let x: string | null | undefined;
if (x || 'foo') {
}
    `,
    `
declare let x: string | null | undefined;
if ((x ||= 'foo')) {
}
    `,
    `
declare let x: number | null | undefined;
if (x || 'foo') {
}
    `,
    `
declare let x: number | null | undefined;
if ((x ||= 'foo')) {
}
    `,
    `
declare let x: boolean | null | undefined;
if (x || 'foo') {
}
    `,
    `
declare let x: boolean | null | undefined;
if ((x ||= 'foo')) {
}
    `,
    `
declare let x: object | null | undefined;
if (x || 'foo') {
}
    `,
    `
declare let x: object | null | undefined;
if ((x ||= 'foo')) {
}
    `,
    `
declare let x: string | null | undefined;
do {} while (x || 'foo');
    `,
    `
declare let x: string | null | undefined;
do {} while ((x ||= 'foo'));
    `,
    `
declare let x: number | null | undefined;
do {} while (x || 'foo');
    `,
    `
declare let x: number | null | undefined;
do {} while ((x ||= 'foo'));
    `,
    `
declare let x: boolean | null | undefined;
do {} while (x || 'foo');
    `,
    `
declare let x: boolean | null | undefined;
do {} while ((x ||= 'foo'));
    `,
    `
declare let x: object | null | undefined;
do {} while (x || 'foo');
    `,
    `
declare let x: object | null | undefined;
do {} while ((x ||= 'foo'));
    `,
    `
declare let x: string | null | undefined;
for (; x || 'foo';) {}
    `,
    `
declare let x: string | null | undefined;
for (; (x ||= 'foo');) {}
    `,
    `
declare let x: number | null | undefined;
for (; x || 'foo';) {}
    `,
    `
declare let x: number | null | undefined;
for (; (x ||= 'foo');) {}
    `,
    `
declare let x: boolean | null | undefined;
for (; x || 'foo';) {}
    `,
    `
declare let x: boolean | null | undefined;
for (; (x ||= 'foo');) {}
    `,
    `
declare let x: object | null | undefined;
for (; x || 'foo';) {}
    `,
    `
declare let x: object | null | undefined;
for (; (x ||= 'foo');) {}
    `,
    `
declare let x: string | null | undefined;
while (x || 'foo') {}
    `,
    `
declare let x: string | null | undefined;
while ((x ||= 'foo')) {}
    `,
    `
declare let x: number | null | undefined;
while (x || 'foo') {}
    `,
    `
declare let x: number | null | undefined;
while ((x ||= 'foo')) {}
    `,
    `
declare let x: boolean | null | undefined;
while (x || 'foo') {}
    `,
    `
declare let x: boolean | null | undefined;
while ((x ||= 'foo')) {}
    `,
    `
declare let x: object | null | undefined;
while (x || 'foo') {}
    `,
    `
declare let x: object | null | undefined;
while ((x ||= 'foo')) {}
    `,

    // ignoreMixedLogicalExpressions
    {
      code: `
declare let a: string | null | undefined;
declare let b: string | null | undefined;
declare let c: string | null | undefined;
a || (b && c);
      `,
      options: [{ ignoreMixedLogicalExpressions: true }],
    },
    {
      code: `
declare let a: number | null | undefined;
declare let b: number | null | undefined;
declare let c: number | null | undefined;
a || (b && c);
      `,
      options: [{ ignoreMixedLogicalExpressions: true }],
    },
    {
      code: `
declare let a: boolean | null | undefined;
declare let b: boolean | null | undefined;
declare let c: boolean | null | undefined;
a || (b && c);
      `,
      options: [{ ignoreMixedLogicalExpressions: true }],
    },
    {
      code: `
declare let a: object | null | undefined;
declare let b: object | null | undefined;
declare let c: object | null | undefined;
a || (b && c);
      `,
      options: [{ ignoreMixedLogicalExpressions: true }],
    },
    {
      code: `
declare let a: string | null | undefined;
declare let b: string | null | undefined;
declare let c: string | null | undefined;
declare let d: string | null | undefined;
a || b || (c && d);
      `,
      options: [{ ignoreMixedLogicalExpressions: true }],
    },
    {
      code: `
declare let a: number | null | undefined;
declare let b: number | null | undefined;
declare let c: number | null | undefined;
declare let d: number | null | undefined;
a || b || (c && d);
      `,
      options: [{ ignoreMixedLogicalExpressions: true }],
    },
    {
      code: `
declare let a: boolean | null | undefined;
declare let b: boolean | null | undefined;
declare let c: boolean | null | undefined;
declare let d: boolean | null | undefined;
a || b || (c && d);
      `,
      options: [{ ignoreMixedLogicalExpressions: true }],
    },
    {
      code: `
declare let a: object | null | undefined;
declare let b: object | null | undefined;
declare let c: object | null | undefined;
declare let d: object | null | undefined;
a || b || (c && d);
      `,
      options: [{ ignoreMixedLogicalExpressions: true }],
    },
    {
      code: `
declare let a: string | null | undefined;
declare let b: string | null | undefined;
declare let c: string | null | undefined;
declare let d: string | null | undefined;
(a && b) || c || d;
      `,
      options: [{ ignoreMixedLogicalExpressions: true }],
    },
    {
      code: `
declare let a: number | null | undefined;
declare let b: number | null | undefined;
declare let c: number | null | undefined;
declare let d: number | null | undefined;
(a && b) || c || d;
      `,
      options: [{ ignoreMixedLogicalExpressions: true }],
    },
    {
      code: `
declare let a: boolean | null | undefined;
declare let b: boolean | null | undefined;
declare let c: boolean | null | undefined;
declare let d: boolean | null | undefined;
(a && b) || c || d;
      `,
      options: [{ ignoreMixedLogicalExpressions: true }],
    },
    {
      code: `
declare let a: object | null | undefined;
declare let b: object | null | undefined;
declare let c: object | null | undefined;
declare let d: object | null | undefined;
(a && b) || c || d;
      `,
      options: [{ ignoreMixedLogicalExpressions: true }],
    },

    // ignorePrimitives
    {
      code: `
declare let x: string | undefined;
x || y;
      `,
      options: [{ ignorePrimitives: { string: true } }],
    },
    {
      code: `
declare let x: number | undefined;
x || y;
      `,
      options: [{ ignorePrimitives: { number: true } }],
    },
    {
      code: `
declare let x: boolean | undefined;
x || y;
      `,
      options: [{ ignorePrimitives: { boolean: true } }],
    },
    {
      code: `
declare let x: bigint | undefined;
x || y;
      `,
      options: [{ ignorePrimitives: { bigint: true } }],
    },
    {
      code: `
declare let x: string | undefined;
x || y;
      `,
      options: [{ ignorePrimitives: true }],
    },
    {
      code: `
declare let x: number | undefined;
x || y;
      `,
      options: [{ ignorePrimitives: true }],
    },
    {
      code: `
declare let x: boolean | undefined;
x || y;
      `,
      options: [{ ignorePrimitives: true }],
    },
    {
      code: `
declare let x: bigint | undefined;
x || y;
      `,
      options: [{ ignorePrimitives: true }],
    },
    {
      code: `
declare let x: (string & { __brand?: any }) | undefined;
x || y;
      `,
      options: [{ ignorePrimitives: { string: true } }],
    },
    {
      code: `
declare let x: (number & { __brand?: any }) | undefined;
x || y;
      `,
      options: [{ ignorePrimitives: { number: true } }],
    },
    {
      code: `
declare let x: (boolean & { __brand?: any }) | undefined;
x || y;
      `,
      options: [{ ignorePrimitives: { boolean: true } }],
    },
    {
      code: `
declare let x: (bigint & { __brand?: any }) | undefined;
x || y;
      `,
      options: [{ ignorePrimitives: { bigint: true } }],
    },
    {
      code: `
declare let x: (string & { __brand?: any }) | undefined;
x || y;
      `,
      options: [{ ignorePrimitives: true }],
    },
    {
      code: `
declare let x: (number & { __brand?: any }) | undefined;
x || y;
      `,
      options: [{ ignorePrimitives: true }],
    },
    {
      code: `
declare let x: (boolean & { __brand?: any }) | undefined;
x || y;
      `,
      options: [{ ignorePrimitives: true }],
    },
    {
      code: `
declare let x: (bigint & { __brand?: any }) | undefined;
x || y;
      `,
      options: [{ ignorePrimitives: true }],
    },
    {
      code: `
declare let x: string | undefined;
x ? x : y;
      `,
      options: [{ ignorePrimitives: { string: true } }],
    },
    {
      code: `
declare let x: number | undefined;
x ? x : y;
      `,
      options: [{ ignorePrimitives: { number: true } }],
    },
    {
      code: `
declare let x: boolean | undefined;
x ? x : y;
      `,
      options: [{ ignorePrimitives: { boolean: true } }],
    },
    {
      code: `
declare let x: bigint | undefined;
x ? x : y;
      `,
      options: [{ ignorePrimitives: { bigint: true } }],
    },
    {
      code: `
declare let x: string | undefined;
!x ? y : x;
      `,
      options: [{ ignorePrimitives: { string: true } }],
    },
    {
      code: `
declare let x: number | undefined;
!x ? y : x;
      `,
      options: [{ ignorePrimitives: { number: true } }],
    },
    {
      code: `
declare let x: boolean | undefined;
!x ? y : x;
      `,
      options: [{ ignorePrimitives: { boolean: true } }],
    },
    {
      code: `
declare let x: bigint | undefined;
!x ? y : x;
      `,
      options: [{ ignorePrimitives: { bigint: true } }],
    },
    {
      code: `
declare let x: string | undefined;
x ? x : y;
      `,
      options: [{ ignorePrimitives: true }],
    },
    {
      code: `
declare let x: number | undefined;
x ? x : y;
      `,
      options: [{ ignorePrimitives: true }],
    },
    {
      code: `
declare let x: boolean | undefined;
x ? x : y;
      `,
      options: [{ ignorePrimitives: true }],
    },
    {
      code: `
declare let x: bigint | undefined;
x ? x : y;
      `,
      options: [{ ignorePrimitives: true }],
    },
    {
      code: `
declare let x: string | undefined;
!x ? y : x;
      `,
      options: [{ ignorePrimitives: true }],
    },
    {
      code: `
declare let x: number | undefined;
!x ? y : x;
      `,
      options: [{ ignorePrimitives: true }],
    },
    {
      code: `
declare let x: boolean | undefined;
!x ? y : x;
      `,
      options: [{ ignorePrimitives: true }],
    },
    {
      code: `
declare let x: bigint | undefined;
!x ? y : x;
      `,
      options: [{ ignorePrimitives: true }],
    },
    {
      code: `
declare let x: (string & { __brand?: any }) | undefined;
x ? x : y;
      `,
      options: [{ ignorePrimitives: { string: true } }],
    },
    {
      code: `
declare let x: (number & { __brand?: any }) | undefined;
x ? x : y;
      `,
      options: [{ ignorePrimitives: { number: true } }],
    },
    {
      code: `
declare let x: (boolean & { __brand?: any }) | undefined;
x ? x : y;
      `,
      options: [{ ignorePrimitives: { boolean: true } }],
    },
    {
      code: `
declare let x: (bigint & { __brand?: any }) | undefined;
x ? x : y;
      `,
      options: [{ ignorePrimitives: { bigint: true } }],
    },
    {
      code: `
declare let x: (string & { __brand?: any }) | undefined;
!x ? y : x;
      `,
      options: [{ ignorePrimitives: { string: true } }],
    },
    {
      code: `
declare let x: (number & { __brand?: any }) | undefined;
!x ? y : x;
      `,
      options: [{ ignorePrimitives: { number: true } }],
    },
    {
      code: `
declare let x: (boolean & { __brand?: any }) | undefined;
!x ? y : x;
      `,
      options: [{ ignorePrimitives: { boolean: true } }],
    },
    {
      code: `
declare let x: (bigint & { __brand?: any }) | undefined;
!x ? y : x;
      `,
      options: [{ ignorePrimitives: { bigint: true } }],
    },
    {
      code: `
declare let x: (string & { __brand?: any }) | undefined;
x ? x : y;
      `,
      options: [{ ignorePrimitives: true }],
    },
    {
      code: `
declare let x: (number & { __brand?: any }) | undefined;
x ? x : y;
      `,
      options: [{ ignorePrimitives: true }],
    },
    {
      code: `
declare let x: (boolean & { __brand?: any }) | undefined;
x ? x : y;
      `,
      options: [{ ignorePrimitives: true }],
    },
    {
      code: `
declare let x: (bigint & { __brand?: any }) | undefined;
x ? x : y;
      `,
      options: [{ ignorePrimitives: true }],
    },
    {
      code: `
declare let x: (string & { __brand?: any }) | undefined;
!x ? y : x;
      `,
      options: [{ ignorePrimitives: true }],
    },
    {
      code: `
declare let x: (number & { __brand?: any }) | undefined;
!x ? y : x;
      `,
      options: [{ ignorePrimitives: true }],
    },
    {
      code: `
declare let x: (boolean & { __brand?: any }) | undefined;
!x ? y : x;
      `,
      options: [{ ignorePrimitives: true }],
    },
    {
      code: `
declare let x: (bigint & { __brand?: any }) | undefined;
!x ? y : x;
      `,
      options: [{ ignorePrimitives: true }],
    },
    `
declare let x: never;
declare let y: number;
x || y;
    `,
    `
declare let x: never;
declare let y: number;
x ? x : y;
    `,
    `
declare let x: never;
declare let y: number;
!x ? y : x;
    `,
    `
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBoxOptional: { a?: { b?: Box | undefined } };

defaultBoxOptional.a?.b !== null ? defaultBoxOptional.a?.b : getFallbackBox();
    `,
    `
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBoxOptional: { a?: { b?: Box | null } };

defaultBoxOptional.a?.b !== null ? defaultBoxOptional.a?.b : getFallbackBox();
    `,
    `
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBoxOptional: { a?: { b?: Box | null } };

defaultBoxOptional.a?.b !== undefined
  ? defaultBoxOptional.a?.b
  : getFallbackBox();
    `,
    `
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBoxOptional: { a?: { b?: Box | null } };

defaultBoxOptional.a?.b !== undefined
  ? defaultBoxOptional.a.b
  : getFallbackBox();
    `,
    {
      code: `
declare let x: 0 | 1 | 0n | 1n | undefined;
x || y;
      `,
      options: [
        {
          ignorePrimitives: {
            bigint: true,
            boolean: true,
            number: false,
            string: true,
          },
        },
      ],
    },
    {
      code: `
declare let x: 0 | 1 | 0n | 1n | undefined;
x || y;
      `,
      options: [
        {
          ignorePrimitives: {
            bigint: false,
            boolean: true,
            number: true,
            string: true,
          },
        },
      ],
    },
    {
      code: `
declare let x: 0 | 'foo' | undefined;
x || y;
      `,
      options: [{ ignorePrimitives: { number: true, string: true } }],
    },
    {
      code: `
declare let x: 0 | 'foo' | undefined;
x || y;
      `,
      options: [{ ignorePrimitives: { number: true, string: false } }],
    },
    {
      code: `
enum Enum {
  A = 0,
  B = 1,
  C = 2,
}
declare let x: Enum | undefined;
x || y;
      `,
      options: [{ ignorePrimitives: { number: true } }],
    },
    {
      code: `
enum Enum {
  A = 0,
  B = 1,
  C = 2,
}
declare let x: Enum.A | Enum.B | undefined;
x || y;
      `,
      options: [{ ignorePrimitives: { number: true } }],
    },
    {
      code: `
enum Enum {
  A = 'a',
  B = 'b',
  C = 'c',
}
declare let x: Enum | undefined;
x || y;
      `,
      options: [{ ignorePrimitives: { string: true } }],
    },
    {
      code: `
enum Enum {
  A = 'a',
  B = 'b',
  C = 'c',
}
declare let x: Enum.A | Enum.B | undefined;
x || y;
      `,
      options: [{ ignorePrimitives: { string: true } }],
    },
    {
      code: `
declare let x: 0 | 1 | 0n | 1n | undefined;
x ? x : y;
      `,
      options: [
        {
          ignorePrimitives: {
            bigint: true,
            boolean: true,
            number: false,
            string: true,
          },
        },
      ],
    },
    {
      code: `
declare let x: 0 | 1 | 0n | 1n | undefined;
!x ? y : x;
      `,
      options: [
        {
          ignorePrimitives: {
            bigint: true,
            boolean: true,
            number: false,
            string: true,
          },
        },
      ],
    },
    {
      code: `
declare let x: 0 | 1 | 0n | 1n | undefined;
x ? x : y;
      `,
      options: [
        {
          ignorePrimitives: {
            bigint: false,
            boolean: true,
            number: true,
            string: true,
          },
        },
      ],
    },
    {
      code: `
declare let x: 0 | 1 | 0n | 1n | undefined;
!x ? y : x;
      `,
      options: [
        {
          ignorePrimitives: {
            bigint: false,
            boolean: true,
            number: true,
            string: true,
          },
        },
      ],
    },
    {
      code: `
declare let x: 0 | 'foo' | undefined;
x ? x : y;
      `,
      options: [
        {
          ignorePrimitives: {
            number: true,
            string: true,
          },
        },
      ],
    },
    {
      code: `
declare let x: 0 | 'foo' | undefined;
!x ? y : x;
      `,
      options: [
        {
          ignorePrimitives: {
            number: true,
            string: true,
          },
        },
      ],
    },
    {
      code: `
declare let x: 0 | 'foo' | undefined;
x ? x : y;
      `,
      options: [
        {
          ignorePrimitives: {
            number: true,
            string: false,
          },
        },
      ],
    },
    {
      code: `
declare let x: 0 | 'foo' | undefined;
!x ? y : x;
      `,
      options: [
        {
          ignorePrimitives: {
            number: true,
            string: false,
          },
        },
      ],
    },
    {
      code: `
enum Enum {
  A = 0,
  B = 1,
  C = 2,
}
declare let x: Enum | undefined;
x ? x : y;
      `,
      options: [
        {
          ignorePrimitives: {
            number: true,
          },
        },
      ],
    },
    {
      code: `
enum Enum {
  A = 0,
  B = 1,
  C = 2,
}
declare let x: Enum | undefined;
!x ? y : x;
      `,
      options: [
        {
          ignorePrimitives: {
            number: true,
          },
        },
      ],
    },
    {
      code: `
enum Enum {
  A = 0,
  B = 1,
  C = 2,
}
declare let x: Enum.A | Enum.B | undefined;
x ? x : y;
      `,
      options: [
        {
          ignorePrimitives: {
            number: true,
          },
        },
      ],
    },
    {
      code: `
enum Enum {
  A = 0,
  B = 1,
  C = 2,
}
declare let x: Enum.A | Enum.B | undefined;
!x ? y : x;
      `,
      options: [
        {
          ignorePrimitives: {
            number: true,
          },
        },
      ],
    },
    {
      code: `
enum Enum {
  A = 'a',
  B = 'b',
  C = 'c',
}
declare let x: Enum | undefined;
x ? x : y;
      `,
      options: [
        {
          ignorePrimitives: {
            string: true,
          },
        },
      ],
    },
    {
      code: `
enum Enum {
  A = 'a',
  B = 'b',
  C = 'c',
}
declare let x: Enum | undefined;
!x ? y : x;
      `,
      options: [
        {
          ignorePrimitives: {
            string: true,
          },
        },
      ],
    },
    {
      code: `
enum Enum {
  A = 'a',
  B = 'b',
  C = 'c',
}
declare let x: Enum.A | Enum.B | undefined;
x ? x : y;
      `,
      options: [
        {
          ignorePrimitives: {
            string: true,
          },
        },
      ],
    },
    {
      code: `
enum Enum {
  A = 'a',
  B = 'b',
  C = 'c',
}
declare let x: Enum.A | Enum.B | undefined;
!x ? y : x;
      `,
      options: [
        {
          ignorePrimitives: {
            string: true,
          },
        },
      ],
    },
    {
      code: `
let a: string | true | undefined;
let b: string | boolean | undefined;

const x = Boolean(a || b);
      `,
      options: [{ ignoreBooleanCoercion: true }],
    },
    {
      code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;

const test = Boolean(a || b || c);
      `,
      options: [{ ignoreBooleanCoercion: true }],
    },
    {
      code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;

const test = Boolean(a || (b && c));
      `,
      options: [{ ignoreBooleanCoercion: true }],
    },
    {
      code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;

const test = Boolean((a || b) ?? c);
      `,
      options: [{ ignoreBooleanCoercion: true }],
    },
    {
      code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;

const test = Boolean(a ?? (b || c));
      `,
      options: [{ ignoreBooleanCoercion: true }],
    },
    {
      code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;

const test = Boolean(a ? b || c : 'fail');
      `,
      options: [{ ignoreBooleanCoercion: true }],
    },
    {
      code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;

const test = Boolean(a ? 'success' : b || c);
      `,
      options: [{ ignoreBooleanCoercion: true }],
    },
    {
      code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;

const test = Boolean(((a = b), b || c));
      `,
      options: [{ ignoreBooleanCoercion: true }],
    },
    {
      code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;

const test = Boolean((a ? a : b) || c);
      `,
      options: [{ ignoreBooleanCoercion: true }],
    },
    {
      code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;

const test = Boolean(c || (!a ? b : a));
      `,
      options: [{ ignoreBooleanCoercion: true }],
    },
    {
      code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;

if (a || b || c) {
}
      `,
      options: [{ ignoreConditionalTests: true }],
    },
    {
      code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;

if (a || (b && c)) {
}
      `,
      options: [{ ignoreConditionalTests: true }],
    },
    {
      code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;

if ((a || b) ?? c) {
}
      `,
      options: [{ ignoreConditionalTests: true }],
    },
    {
      code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;

if (a ?? (b || c)) {
}
      `,
      options: [{ ignoreConditionalTests: true }],
    },
    {
      code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;

if (a ? b || c : 'fail') {
}
      `,
      options: [{ ignoreConditionalTests: true }],
    },
    {
      code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;

if (a ? 'success' : b || c) {
}
      `,
      options: [{ ignoreConditionalTests: true }],
    },
    {
      code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;

if (((a = b), b || c)) {
}
      `,
      options: [{ ignoreConditionalTests: true }],
    },
    {
      code: `
let a: string | undefined;
let b: string | undefined;

if (!(a || b)) {
}
      `,
      options: [{ ignoreConditionalTests: true }],
    },
    {
      code: `
let a: string | undefined;
let b: string | undefined;

if (!!(a || b)) {
}
      `,
      options: [{ ignoreConditionalTests: true }],
    },
    {
      code: `
let a: string | true | undefined;
let b: string | boolean | undefined;

if (a ? a : b) {
}
      `,
      options: [{ ignoreConditionalTests: true }],
    },
    {
      code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;

if (!a ? b : a) {
}
      `,
      options: [{ ignoreConditionalTests: true }],
    },
    {
      code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;

if ((a ? a : b) || c) {
}
      `,
      options: [{ ignoreConditionalTests: true }],
    },
    {
      code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;

if (c || (!a ? b : a)) {
}
      `,
      options: [{ ignoreConditionalTests: true }],
    },

    {
      code: `
declare const a: any;
declare const b: any;
a ? a : b;
      `,
      options: [{ ignorePrimitives: true }],
    },

    {
      code: `
declare const a: any;
declare const b: any;
a ? a : b;
      `,
      options: [
        {
          ignorePrimitives: {
            number: true,
          },
        },
      ],
    },

    {
      code: `
declare const a: unknown;
const b = a || 'bar';
      `,
      options: [
        {
          ignorePrimitives: {
            bigint: true,
            boolean: false,
            number: false,
            string: false,
          },
        },
      ],
    },
  ],

  invalid: [
    {
      code: `
declare let x: string | null | undefined;
x || 'foo';
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 3,
          column: 3,
          endLine: 3,
          endColumn: 5,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: string | null | undefined;
x ?? 'foo';
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
declare let x: string | null | undefined;
x ||= 'foo';
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.',
          line: 3,
          column: 3,
          endLine: 3,
          endColumn: 6,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??=`).',
              output: `
declare let x: string | null | undefined;
x ??= 'foo';
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
declare let x: number | null | undefined;
x || 'foo';
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 3,
          column: 3,
          endLine: 3,
          endColumn: 5,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: number | null | undefined;
x ?? 'foo';
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
declare let x: number | null | undefined;
x ||= 'foo';
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.',
          line: 3,
          column: 3,
          endLine: 3,
          endColumn: 6,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??=`).',
              output: `
declare let x: number | null | undefined;
x ??= 'foo';
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
declare let x: boolean | null | undefined;
x || 'foo';
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 3,
          column: 3,
          endLine: 3,
          endColumn: 5,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: boolean | null | undefined;
x ?? 'foo';
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
declare let x: boolean | null | undefined;
x ||= 'foo';
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.',
          line: 3,
          column: 3,
          endLine: 3,
          endColumn: 6,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??=`).',
              output: `
declare let x: boolean | null | undefined;
x ??= 'foo';
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
declare let x: object | null | undefined;
x || 'foo';
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 3,
          column: 3,
          endLine: 3,
          endColumn: 5,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: object | null | undefined;
x ?? 'foo';
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
declare let x: object | null | undefined;
x ||= 'foo';
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.',
          line: 3,
          column: 3,
          endLine: 3,
          endColumn: 6,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??=`).',
              output: `
declare let x: object | null | undefined;
x ??= 'foo';
      `,
            },
          ],
        },
      ],
    },

    {
      code: 'x !== undefined && x !== null ? x : y;',
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 38,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `x ?? y;`,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
x.z[1][this[this.o]]['3'][a.b.c] !== undefined &&
x.z[1][this[this.o]]['3'][a.b.c] !== null
  ? x.z[1][this[this.o]]['3'][a.b.c]
  : y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 2,
          column: 1,
          endLine: 5,
          endColumn: 6,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
x.z[1][this[this.o]]['3'][a.b.c] ?? y;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: 'x !== undefined && x !== null ? x : (z = y);',
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 44,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `x ?? (z = y);`,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: 'x !== null && x !== undefined ? x : y;',
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 38,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `x ?? y;`,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
x.z[1][this[this.o]]['3'][a.b.c] !== null &&
x.z[1][this[this.o]]['3'][a.b.c] !== undefined
  ? x.z[1][this[this.o]]['3'][a.b.c]
  : y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 2,
          column: 1,
          endLine: 5,
          endColumn: 6,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
x.z[1][this[this.o]]['3'][a.b.c] ?? y;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: 'x !== null && x !== undefined ? x : (z = y);',
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 44,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `x ?? (z = y);`,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: 'x === undefined || x === null ? y : x;',
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 38,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `x ?? y;`,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
x.z[1][this[this.o]]['3'][a.b.c] === undefined ||
x.z[1][this[this.o]]['3'][a.b.c] === null
  ? y
  : x.z[1][this[this.o]]['3'][a.b.c];
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 2,
          column: 1,
          endLine: 5,
          endColumn: 37,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
x.z[1][this[this.o]]['3'][a.b.c] ?? y;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: 'x === undefined || x === null ? (z = y) : x;',
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 44,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `x ?? (z = y);`,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: 'x === null || x === undefined ? y : x;',
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 38,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `x ?? y;`,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
x.z[1][this[this.o]]['3'][a.b.c] === null ||
x.z[1][this[this.o]]['3'][a.b.c] === undefined
  ? y
  : x.z[1][this[this.o]]['3'][a.b.c];
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 2,
          column: 1,
          endLine: 5,
          endColumn: 37,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
x.z[1][this[this.o]]['3'][a.b.c] ?? y;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: 'x === null || x === undefined ? (z = y) : x;',
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 44,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `x ?? (z = y);`,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: 'undefined !== x && x !== null ? x : y;',
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 38,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `x ?? y;`,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
undefined !== x.z[1][this[this.o]]['3'][a.b.c] &&
x.z[1][this[this.o]]['3'][a.b.c] !== null
  ? x.z[1][this[this.o]]['3'][a.b.c]
  : y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 2,
          column: 1,
          endLine: 5,
          endColumn: 6,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
x.z[1][this[this.o]]['3'][a.b.c] ?? y;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: 'undefined !== x && x !== null ? x : (z = y);',
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 44,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `x ?? (z = y);`,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: 'null !== x && x !== undefined ? x : y;',
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 38,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `x ?? y;`,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
null !== x.z[1][this[this.o]]['3'][a.b.c] &&
x.z[1][this[this.o]]['3'][a.b.c] !== undefined
  ? x.z[1][this[this.o]]['3'][a.b.c]
  : y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 2,
          column: 1,
          endLine: 5,
          endColumn: 6,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
x.z[1][this[this.o]]['3'][a.b.c] ?? y;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: 'null !== x && x !== undefined ? x : (z = y);',
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 44,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `x ?? (z = y);`,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: 'undefined === x || x === null ? y : x;',
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 38,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `x ?? y;`,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
undefined === x.z[1][this[this.o]]['3'][a.b.c] ||
x.z[1][this[this.o]]['3'][a.b.c] === null
  ? y
  : x.z[1][this[this.o]]['3'][a.b.c];
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 2,
          column: 1,
          endLine: 5,
          endColumn: 37,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
x.z[1][this[this.o]]['3'][a.b.c] ?? y;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: 'undefined === x || x === null ? (z = y) : x;',
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 44,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `x ?? (z = y);`,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: 'null === x || x === undefined ? y : x;',
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 38,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `x ?? y;`,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
null === x.z[1][this[this.o]]['3'][a.b.c] ||
x.z[1][this[this.o]]['3'][a.b.c] === undefined
  ? y
  : x.z[1][this[this.o]]['3'][a.b.c];
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 2,
          column: 1,
          endLine: 5,
          endColumn: 37,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
x.z[1][this[this.o]]['3'][a.b.c] ?? y;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: 'null === x || x === undefined ? (z = y) : x;',
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 44,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `x ?? (z = y);`,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: 'x !== undefined && null !== x ? x : y;',
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 38,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `x ?? y;`,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
x.z[1][this[this.o]]['3'][a.b.c] !== undefined &&
null !== x.z[1][this[this.o]]['3'][a.b.c]
  ? x.z[1][this[this.o]]['3'][a.b.c]
  : y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 2,
          column: 1,
          endLine: 5,
          endColumn: 6,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
x.z[1][this[this.o]]['3'][a.b.c] ?? y;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: 'x !== undefined && null !== x ? x : (z = y);',
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 44,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `x ?? (z = y);`,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: 'x !== null && undefined !== x ? x : y;',
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 38,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `x ?? y;`,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
x.z[1][this[this.o]]['3'][a.b.c] !== null &&
undefined !== x.z[1][this[this.o]]['3'][a.b.c]
  ? x.z[1][this[this.o]]['3'][a.b.c]
  : y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 2,
          column: 1,
          endLine: 5,
          endColumn: 6,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
x.z[1][this[this.o]]['3'][a.b.c] ?? y;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: 'x !== null && undefined !== x ? x : (z = y);',
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 44,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `x ?? (z = y);`,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: 'x === undefined || null === x ? y : x;',
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 38,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `x ?? y;`,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
x.z[1][this[this.o]]['3'][a.b.c] === undefined ||
null === x.z[1][this[this.o]]['3'][a.b.c]
  ? y
  : x.z[1][this[this.o]]['3'][a.b.c];
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 2,
          column: 1,
          endLine: 5,
          endColumn: 37,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
x.z[1][this[this.o]]['3'][a.b.c] ?? y;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: 'x === undefined || null === x ? (z = y) : x;',
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 44,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `x ?? (z = y);`,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: 'x === null || undefined === x ? y : x;',
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 38,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `x ?? y;`,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
x.z[1][this[this.o]]['3'][a.b.c] === null ||
undefined === x.z[1][this[this.o]]['3'][a.b.c]
  ? y
  : x.z[1][this[this.o]]['3'][a.b.c];
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 2,
          column: 1,
          endLine: 5,
          endColumn: 37,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
x.z[1][this[this.o]]['3'][a.b.c] ?? y;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: 'x === null || undefined === x ? (z = y) : x;',
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 44,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `x ?? (z = y);`,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: 'undefined !== x && null !== x ? x : y;',
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 38,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `x ?? y;`,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
undefined !== x.z[1][this[this.o]]['3'][a.b.c] &&
null !== x.z[1][this[this.o]]['3'][a.b.c]
  ? x.z[1][this[this.o]]['3'][a.b.c]
  : y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 2,
          column: 1,
          endLine: 5,
          endColumn: 6,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
x.z[1][this[this.o]]['3'][a.b.c] ?? y;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: 'undefined !== x && null !== x ? x : (z = y);',
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 44,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `x ?? (z = y);`,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: 'null !== x && undefined !== x ? x : y;',
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 38,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `x ?? y;`,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
null !== x.z[1][this[this.o]]['3'][a.b.c] &&
undefined !== x.z[1][this[this.o]]['3'][a.b.c]
  ? x.z[1][this[this.o]]['3'][a.b.c]
  : y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 2,
          column: 1,
          endLine: 5,
          endColumn: 6,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
x.z[1][this[this.o]]['3'][a.b.c] ?? y;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: 'null !== x && undefined !== x ? x : (z = y);',
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 44,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `x ?? (z = y);`,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: 'undefined === x || null === x ? y : x;',
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 38,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `x ?? y;`,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
undefined === x.z[1][this[this.o]]['3'][a.b.c] ||
null === x.z[1][this[this.o]]['3'][a.b.c]
  ? y
  : x.z[1][this[this.o]]['3'][a.b.c];
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 2,
          column: 1,
          endLine: 5,
          endColumn: 37,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
x.z[1][this[this.o]]['3'][a.b.c] ?? y;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: 'undefined === x || null === x ? (z = y) : x;',
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 44,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `x ?? (z = y);`,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: 'null === x || undefined === x ? y : x;',
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 38,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `x ?? y;`,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
null === x.z[1][this[this.o]]['3'][a.b.c] ||
undefined === x.z[1][this[this.o]]['3'][a.b.c]
  ? y
  : x.z[1][this[this.o]]['3'][a.b.c];
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 2,
          column: 1,
          endLine: 5,
          endColumn: 37,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
x.z[1][this[this.o]]['3'][a.b.c] ?? y;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: 'null === x || undefined === x ? (z = y) : x;',
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 44,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `x ?? (z = y);`,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: 'x != undefined && x != null ? x : y;',
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 36,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `x ?? y;`,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
x.z[1][this[this.o]]['3'][a.b.c] != undefined &&
x.z[1][this[this.o]]['3'][a.b.c] != null
  ? x.z[1][this[this.o]]['3'][a.b.c]
  : y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 2,
          column: 1,
          endLine: 5,
          endColumn: 6,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
x.z[1][this[this.o]]['3'][a.b.c] ?? y;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: 'x != undefined && x != null ? x : (z = y);',
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 42,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `x ?? (z = y);`,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: 'x == undefined || x == null ? y : x;',
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 36,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `x ?? y;`,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
x.z[1][this[this.o]]['3'][a.b.c] == undefined ||
x.z[1][this[this.o]]['3'][a.b.c] == null
  ? y
  : x.z[1][this[this.o]]['3'][a.b.c];
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 2,
          column: 1,
          endLine: 5,
          endColumn: 37,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
x.z[1][this[this.o]]['3'][a.b.c] ?? y;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: 'x == undefined || x == null ? (z = y) : x;',
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 42,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `x ?? (z = y);`,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: 'x != undefined && x !== null ? x : y;',
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 37,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `x ?? y;`,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
x.z[1][this[this.o]]['3'][a.b.c] != undefined &&
x.z[1][this[this.o]]['3'][a.b.c] !== null
  ? x.z[1][this[this.o]]['3'][a.b.c]
  : y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 2,
          column: 1,
          endLine: 5,
          endColumn: 6,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
x.z[1][this[this.o]]['3'][a.b.c] ?? y;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: 'x != undefined && x !== null ? x : (z = y);',
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 43,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `x ?? (z = y);`,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: 'x == undefined || x === null ? y : x;',
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 37,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `x ?? y;`,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
x.z[1][this[this.o]]['3'][a.b.c] == undefined ||
x.z[1][this[this.o]]['3'][a.b.c] === null
  ? y
  : x.z[1][this[this.o]]['3'][a.b.c];
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 2,
          column: 1,
          endLine: 5,
          endColumn: 37,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
x.z[1][this[this.o]]['3'][a.b.c] ?? y;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: 'x == undefined || x === null ? (z = y) : x;',
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 43,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `x ?? (z = y);`,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: 'x !== undefined && x != null ? x : y;',
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 37,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `x ?? y;`,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
x.z[1][this[this.o]]['3'][a.b.c] !== undefined &&
x.z[1][this[this.o]]['3'][a.b.c] != null
  ? x.z[1][this[this.o]]['3'][a.b.c]
  : y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 2,
          column: 1,
          endLine: 5,
          endColumn: 6,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
x.z[1][this[this.o]]['3'][a.b.c] ?? y;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: 'x !== undefined && x != null ? x : (z = y);',
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 43,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `x ?? (z = y);`,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: 'undefined != x ? x : y;',
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 23,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `x ?? y;`,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
undefined != x.z[1][this[this.o]]['3'][a.b.c]
  ? x.z[1][this[this.o]]['3'][a.b.c]
  : y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 2,
          column: 1,
          endLine: 4,
          endColumn: 6,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
x.z[1][this[this.o]]['3'][a.b.c] ?? y;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: 'undefined != x ? x : (z = y);',
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 29,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `x ?? (z = y);`,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: 'null != x ? x : y;',
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 18,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `x ?? y;`,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: "null != x.z[1][this[this.o]]['3'][a.b.c] ? x.z[1][this[this.o]]['3'][a.b.c] : y;",
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 80,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `x.z[1][this[this.o]]['3'][a.b.c] ?? y;`,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: 'null != x ? x : (z = y);',
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 24,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `x ?? (z = y);`,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: 'undefined == x ? y : x;',
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 23,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `x ?? y;`,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
undefined == x.z[1][this[this.o]]['3'][a.b.c]
  ? y
  : x.z[1][this[this.o]]['3'][a.b.c];
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 2,
          column: 1,
          endLine: 4,
          endColumn: 37,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
x.z[1][this[this.o]]['3'][a.b.c] ?? y;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: 'undefined == x ? (z = y) : x;',
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 29,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `x ?? (z = y);`,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: 'null == x ? y : x;',
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 18,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `x ?? y;`,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: "null == x.z[1][this[this.o]]['3'][a.b.c] ? y : x.z[1][this[this.o]]['3'][a.b.c];",
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 80,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `x.z[1][this[this.o]]['3'][a.b.c] ?? y;`,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: 'null == x ? (z = y) : x;',
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 24,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `x ?? (z = y);`,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: 'x != undefined ? x : y;',
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 23,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `x ?? y;`,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
x.z[1][this[this.o]]['3'][a.b.c] != undefined
  ? x.z[1][this[this.o]]['3'][a.b.c]
  : y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 2,
          column: 1,
          endLine: 4,
          endColumn: 6,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
x.z[1][this[this.o]]['3'][a.b.c] ?? y;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: 'x != undefined ? x : (z = y);',
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 29,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `x ?? (z = y);`,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: 'x != null ? x : y;',
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 18,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `x ?? y;`,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: "x.z[1][this[this.o]]['3'][a.b.c] != null ? x.z[1][this[this.o]]['3'][a.b.c] : y;",
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 80,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `x.z[1][this[this.o]]['3'][a.b.c] ?? y;`,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: 'x != null ? x : (z = y);',
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 24,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `x ?? (z = y);`,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: 'x == undefined ? y : x;',
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 23,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `x ?? y;`,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
x.z[1][this[this.o]]['3'][a.b.c] == undefined
  ? y
  : x.z[1][this[this.o]]['3'][a.b.c];
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 2,
          column: 1,
          endLine: 4,
          endColumn: 37,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
x.z[1][this[this.o]]['3'][a.b.c] ?? y;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: 'x == undefined ? (z = y) : x;',
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 29,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `x ?? (z = y);`,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: 'x == null ? y : x;',
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 18,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `x ?? y;`,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: "x.z[1][this[this.o]]['3'][a.b.c] == null ? y : x.z[1][this[this.o]]['3'][a.b.c];",
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 80,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `x.z[1][this[this.o]]['3'][a.b.c] ?? y;`,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: 'x == null ? (z = y) : x;',
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 24,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `x ?? (z = y);`,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },

    {
      code: 'this != undefined ? this : y;',
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 29,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `this ?? y;`,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },

    {
      code: `
declare let x: string | null | undefined;
x ? x : y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: string | null | undefined;
x ?? y;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: string | null | undefined;
!x ? y : x;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 11,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: string | null | undefined;
x ?? y;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: number | null | undefined;
x ? x : y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: number | null | undefined;
x ?? y;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: number | null | undefined;
!x ? y : x;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 11,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: number | null | undefined;
x ?? y;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: boolean | null | undefined;
x ? x : y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: boolean | null | undefined;
x ?? y;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: boolean | null | undefined;
!x ? y : x;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 11,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: boolean | null | undefined;
x ?? y;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: object | null | undefined;
x ? x : y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: object | null | undefined;
x ?? y;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: object | null | undefined;
!x ? y : x;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 11,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: object | null | undefined;
x ?? y;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },

    {
      code: `
declare let x: { n: string | null | undefined };
x.n ? x.n : y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 14,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: { n: string | null | undefined };
x.n ?? y;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: { n: string | null | undefined };
!x.n ? y : x.n;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 15,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: { n: string | null | undefined };
x.n ?? y;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: { n: number | null | undefined };
x.n ? x.n : y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 14,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: { n: number | null | undefined };
x.n ?? y;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: { n: number | null | undefined };
!x.n ? y : x.n;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 15,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: { n: number | null | undefined };
x.n ?? y;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: { n: boolean | null | undefined };
x.n ? x.n : y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 14,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: { n: boolean | null | undefined };
x.n ?? y;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: { n: boolean | null | undefined };
!x.n ? y : x.n;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 15,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: { n: boolean | null | undefined };
x.n ?? y;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: { n: object | null | undefined };
x.n ? x.n : y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 14,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: { n: object | null | undefined };
x.n ?? y;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: { n: object | null | undefined };
!x.n ? y : x.n;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 15,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: { n: object | null | undefined };
x.n ?? y;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },

    {
      code: `
declare let x: { n?: { a?: string } };
x.n?.a ? x?.n?.a : y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 21,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: { n?: { a?: string } };
x.n?.a ?? y;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: { n?: { a?: string } };
x.n?.a ? x?.n.a : y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 20,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: { n?: { a?: string } };
x.n?.a ?? y;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: { n?: { a?: string } };
x.n?.a ? x.n.a : y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 19,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: { n?: { a?: string } };
x.n?.a ?? y;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: { n?: { a?: string } };
x.n?.a !== undefined ? x?.n?.a : y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 35,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: { n?: { a?: string } };
x.n?.a ?? y;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: { n?: { a?: string } };
x.n?.a !== undefined ? x?.n.a : y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 34,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: { n?: { a?: string } };
x.n?.a ?? y;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: { n?: { a?: string } };
x.n?.a !== undefined ? x.n.a : y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 33,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: { n?: { a?: string } };
x.n?.a ?? y;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: { n?: { a?: string } };
x.n?.a != undefined ? x?.n?.a : y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 34,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: { n?: { a?: string } };
x.n?.a ?? y;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: { n?: { a?: string } };
x.n?.a != undefined ? x?.n.a : y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 33,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: { n?: { a?: string } };
x.n?.a ?? y;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: { n?: { a?: string } };
x.n?.a != undefined ? x.n.a : y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 32,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: { n?: { a?: string } };
x.n?.a ?? y;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: { n?: { a?: string } };
x.n?.a != null ? x?.n?.a : y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 29,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: { n?: { a?: string } };
x.n?.a ?? y;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: { n?: { a?: string } };
x.n?.a != null ? x?.n.a : y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 28,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: { n?: { a?: string } };
x.n?.a ?? y;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: { n?: { a?: string } };
x.n?.a != null ? x.n.a : y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 27,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: { n?: { a?: string } };
x.n?.a ?? y;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: { n?: { a?: string | null } };
x.n?.a !== undefined && x.n.a !== null ? x?.n?.a : y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 53,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: { n?: { a?: string | null } };
x.n?.a ?? y;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: { n?: { a?: string | null } };
x.n?.a !== undefined && x.n.a !== null ? x.n.a : y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 51,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: { n?: { a?: string | null } };
x.n?.a ?? y;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: { n?: { a?: string } };
x?.n?.a ? x?.n?.a : y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 22,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: { n?: { a?: string } };
x?.n?.a ?? y;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: { n?: { a?: string } };
x?.n?.a ? x.n?.a : y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 21,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: { n?: { a?: string } };
x?.n?.a ?? y;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: { n?: { a?: string } };
x?.n?.a ? x?.n.a : y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 21,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: { n?: { a?: string } };
x?.n?.a ?? y;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: { n?: { a?: string } };
x?.n?.a ? x.n.a : y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 20,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: { n?: { a?: string } };
x?.n?.a ?? y;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: { n?: { a?: string } };
x?.n?.a !== undefined ? x?.n?.a : y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 36,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: { n?: { a?: string } };
x?.n?.a ?? y;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: { n?: { a?: string } };
x?.n?.a !== undefined ? x.n?.a : y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 35,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: { n?: { a?: string } };
x?.n?.a ?? y;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: { n?: { a?: string } };
x?.n?.a !== undefined ? x?.n.a : y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 35,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: { n?: { a?: string } };
x?.n?.a ?? y;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: { n?: { a?: string } };
x?.n?.a !== undefined ? x.n.a : y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 34,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: { n?: { a?: string } };
x?.n?.a ?? y;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: { n?: { a?: string } };
x?.n?.a != undefined ? x?.n?.a : y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 35,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: { n?: { a?: string } };
x?.n?.a ?? y;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: { n?: { a?: string } };
x?.n?.a != undefined ? x.n?.a : y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 34,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: { n?: { a?: string } };
x?.n?.a ?? y;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: { n?: { a?: string } };
x?.n?.a != undefined ? x?.n.a : y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 34,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: { n?: { a?: string } };
x?.n?.a ?? y;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: { n?: { a?: string } };
x?.n?.a != undefined ? x.n.a : y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 33,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: { n?: { a?: string } };
x?.n?.a ?? y;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: { n?: { a?: string } };
x?.n?.a != null ? x?.n?.a : y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 30,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: { n?: { a?: string } };
x?.n?.a ?? y;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: { n?: { a?: string } };
x?.n?.a != null ? x.n?.a : y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 29,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: { n?: { a?: string } };
x?.n?.a ?? y;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: { n?: { a?: string } };
x?.n?.a != null ? x?.n.a : y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 29,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: { n?: { a?: string } };
x?.n?.a ?? y;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: { n?: { a?: string } };
x?.n?.a != null ? x.n.a : y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 28,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: { n?: { a?: string } };
x?.n?.a ?? y;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: { n?: { a?: string | null } };
x?.n?.a !== undefined && x.n.a !== null ? x?.n?.a : y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 54,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: { n?: { a?: string | null } };
x?.n?.a ?? y;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: { n?: { a?: string | null } };
x?.n?.a !== undefined && x.n.a !== null ? x.n?.a : y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 53,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: { n?: { a?: string | null } };
x?.n?.a ?? y;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: { n?: { a?: string | null } };
x?.n?.a !== undefined && x.n.a !== null ? x?.n.a : y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 53,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: { n?: { a?: string | null } };
x?.n?.a ?? y;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: { n?: { a?: string | null } };
x?.n?.a !== undefined && x.n.a !== null ? x.n.a : y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 52,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: { n?: { a?: string | null } };
x?.n?.a ?? y;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: { n?: { a?: string | null } };
x?.n?.a ? x?.n?.a : y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 22,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: { n?: { a?: string | null } };
x?.n?.a ?? y;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: { n?: { a?: string | null } };
x?.n?.a ? x.n?.a : y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 21,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: { n?: { a?: string | null } };
x?.n?.a ?? y;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: { n?: { a?: string | null } };
x?.n?.a ? x?.n.a : y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 21,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: { n?: { a?: string | null } };
x?.n?.a ?? y;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: { n?: { a?: string | null } };
x?.n?.a ? x.n.a : y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 20,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: { n?: { a?: string | null } };
x?.n?.a ?? y;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },

    {
      code: `
declare let x: { n?: { a?: string | null } };
x.n?.a ? x?.n?.a : y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 21,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: { n?: { a?: string | null } };
x.n?.a ?? y;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: { n?: { a?: string | null } };
x.n?.a ? x.n?.a : y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 20,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: { n?: { a?: string | null } };
x.n?.a ?? y;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: { n?: { a?: string | null } };
x.n?.a ? x?.n.a : y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 20,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: { n?: { a?: string | null } };
x.n?.a ?? y;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: { n?: { a?: string | null } };
x.n?.a ? x.n.a : y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 19,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: { n?: { a?: string | null } };
x.n?.a ?? y;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: { n?: { a?: string | null } };
x.n?.a ? (x?.n).a : y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 22,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: { n?: { a?: string | null } };
x.n?.a ?? y;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },

    // noStrictNullCheck
    {
      code: `
declare let x: string[] | null;
if (x) {
}
      `,
      errors: [
        {
          messageId: 'noStrictNullCheck',
          message: 'This rule requires the `strictNullChecks` compiler option to be turned on to function correctly.',
          // Upstream reports `loc: { line: 0, column: 0 }`, outside ESLint's
          // 1-based lines, and expects line 0. rslint reports offset 0 of the
          // file, which is 1:1.
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 1,
        },
      ],
      filename: 'unstrict/file.ts',
      languageOptions: {
        parserOptions: { project: './unstrict/tsconfig.json' },
      },
    },

    // ignoreConditionalTests
    {
      code: `
declare let x: string | null | undefined;
x || 'foo' ? null : null;
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 3,
          column: 3,
          endLine: 3,
          endColumn: 5,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: string | null | undefined;
x ?? 'foo' ? null : null;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreConditionalTests: false }],
    },
    {
      code: `
declare let x: string | null | undefined;
(x ||= 'foo') ? null : null;
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.',
          line: 3,
          column: 4,
          endLine: 3,
          endColumn: 7,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??=`).',
              output: `
declare let x: string | null | undefined;
(x ??= 'foo') ? null : null;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreConditionalTests: false }],
    },
    {
      code: `
declare let x: number | null | undefined;
x || 'foo' ? null : null;
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 3,
          column: 3,
          endLine: 3,
          endColumn: 5,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: number | null | undefined;
x ?? 'foo' ? null : null;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreConditionalTests: false }],
    },
    {
      code: `
declare let x: number | null | undefined;
(x ||= 'foo') ? null : null;
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.',
          line: 3,
          column: 4,
          endLine: 3,
          endColumn: 7,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??=`).',
              output: `
declare let x: number | null | undefined;
(x ??= 'foo') ? null : null;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreConditionalTests: false }],
    },
    {
      code: `
declare let x: boolean | null | undefined;
x || 'foo' ? null : null;
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 3,
          column: 3,
          endLine: 3,
          endColumn: 5,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: boolean | null | undefined;
x ?? 'foo' ? null : null;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreConditionalTests: false }],
    },
    {
      code: `
declare let x: boolean | null | undefined;
(x ||= 'foo') ? null : null;
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.',
          line: 3,
          column: 4,
          endLine: 3,
          endColumn: 7,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??=`).',
              output: `
declare let x: boolean | null | undefined;
(x ??= 'foo') ? null : null;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreConditionalTests: false }],
    },
    {
      code: `
declare let x: object | null | undefined;
x || 'foo' ? null : null;
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 3,
          column: 3,
          endLine: 3,
          endColumn: 5,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: object | null | undefined;
x ?? 'foo' ? null : null;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreConditionalTests: false }],
    },
    {
      code: `
declare let x: object | null | undefined;
(x ||= 'foo') ? null : null;
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.',
          line: 3,
          column: 4,
          endLine: 3,
          endColumn: 7,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??=`).',
              output: `
declare let x: object | null | undefined;
(x ??= 'foo') ? null : null;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreConditionalTests: false }],
    },
    {
      code: `
declare let x: string | null | undefined;
if (x || 'foo') {
}
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 3,
          column: 7,
          endLine: 3,
          endColumn: 9,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: string | null | undefined;
if (x ?? 'foo') {
}
      `,
            },
          ],
        },
      ],
      options: [{ ignoreConditionalTests: false }],
    },
    {
      code: `
declare let x: string | null | undefined;
if ((x ||= 'foo')) {
}
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.',
          line: 3,
          column: 8,
          endLine: 3,
          endColumn: 11,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??=`).',
              output: `
declare let x: string | null | undefined;
if ((x ??= 'foo')) {
}
      `,
            },
          ],
        },
      ],
      options: [{ ignoreConditionalTests: false }],
    },
    {
      code: `
declare let x: number | null | undefined;
if (x || 'foo') {
}
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 3,
          column: 7,
          endLine: 3,
          endColumn: 9,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: number | null | undefined;
if (x ?? 'foo') {
}
      `,
            },
          ],
        },
      ],
      options: [{ ignoreConditionalTests: false }],
    },
    {
      code: `
declare let x: number | null | undefined;
if ((x ||= 'foo')) {
}
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.',
          line: 3,
          column: 8,
          endLine: 3,
          endColumn: 11,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??=`).',
              output: `
declare let x: number | null | undefined;
if ((x ??= 'foo')) {
}
      `,
            },
          ],
        },
      ],
      options: [{ ignoreConditionalTests: false }],
    },
    {
      code: `
declare let x: boolean | null | undefined;
if (x || 'foo') {
}
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 3,
          column: 7,
          endLine: 3,
          endColumn: 9,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: boolean | null | undefined;
if (x ?? 'foo') {
}
      `,
            },
          ],
        },
      ],
      options: [{ ignoreConditionalTests: false }],
    },
    {
      code: `
declare let x: boolean | null | undefined;
if ((x ||= 'foo')) {
}
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.',
          line: 3,
          column: 8,
          endLine: 3,
          endColumn: 11,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??=`).',
              output: `
declare let x: boolean | null | undefined;
if ((x ??= 'foo')) {
}
      `,
            },
          ],
        },
      ],
      options: [{ ignoreConditionalTests: false }],
    },
    {
      code: `
declare let x: object | null | undefined;
if (x || 'foo') {
}
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 3,
          column: 7,
          endLine: 3,
          endColumn: 9,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: object | null | undefined;
if (x ?? 'foo') {
}
      `,
            },
          ],
        },
      ],
      options: [{ ignoreConditionalTests: false }],
    },
    {
      code: `
declare let x: object | null | undefined;
if ((x ||= 'foo')) {
}
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.',
          line: 3,
          column: 8,
          endLine: 3,
          endColumn: 11,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??=`).',
              output: `
declare let x: object | null | undefined;
if ((x ??= 'foo')) {
}
      `,
            },
          ],
        },
      ],
      options: [{ ignoreConditionalTests: false }],
    },
    {
      code: `
declare let x: string | null | undefined;
do {} while (x || 'foo');
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 3,
          column: 16,
          endLine: 3,
          endColumn: 18,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: string | null | undefined;
do {} while (x ?? 'foo');
      `,
            },
          ],
        },
      ],
      options: [{ ignoreConditionalTests: false }],
    },
    {
      code: `
declare let x: string | null | undefined;
do {} while ((x ||= 'foo'));
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.',
          line: 3,
          column: 17,
          endLine: 3,
          endColumn: 20,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??=`).',
              output: `
declare let x: string | null | undefined;
do {} while ((x ??= 'foo'));
      `,
            },
          ],
        },
      ],
      options: [{ ignoreConditionalTests: false }],
    },
    {
      code: `
declare let x: number | null | undefined;
do {} while (x || 'foo');
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 3,
          column: 16,
          endLine: 3,
          endColumn: 18,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: number | null | undefined;
do {} while (x ?? 'foo');
      `,
            },
          ],
        },
      ],
      options: [{ ignoreConditionalTests: false }],
    },
    {
      code: `
declare let x: number | null | undefined;
do {} while ((x ||= 'foo'));
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.',
          line: 3,
          column: 17,
          endLine: 3,
          endColumn: 20,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??=`).',
              output: `
declare let x: number | null | undefined;
do {} while ((x ??= 'foo'));
      `,
            },
          ],
        },
      ],
      options: [{ ignoreConditionalTests: false }],
    },
    {
      code: `
declare let x: boolean | null | undefined;
do {} while (x || 'foo');
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 3,
          column: 16,
          endLine: 3,
          endColumn: 18,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: boolean | null | undefined;
do {} while (x ?? 'foo');
      `,
            },
          ],
        },
      ],
      options: [{ ignoreConditionalTests: false }],
    },
    {
      code: `
declare let x: boolean | null | undefined;
do {} while ((x ||= 'foo'));
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.',
          line: 3,
          column: 17,
          endLine: 3,
          endColumn: 20,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??=`).',
              output: `
declare let x: boolean | null | undefined;
do {} while ((x ??= 'foo'));
      `,
            },
          ],
        },
      ],
      options: [{ ignoreConditionalTests: false }],
    },
    {
      code: `
declare let x: object | null | undefined;
do {} while (x || 'foo');
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 3,
          column: 16,
          endLine: 3,
          endColumn: 18,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: object | null | undefined;
do {} while (x ?? 'foo');
      `,
            },
          ],
        },
      ],
      options: [{ ignoreConditionalTests: false }],
    },
    {
      code: `
declare let x: object | null | undefined;
do {} while ((x ||= 'foo'));
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.',
          line: 3,
          column: 17,
          endLine: 3,
          endColumn: 20,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??=`).',
              output: `
declare let x: object | null | undefined;
do {} while ((x ??= 'foo'));
      `,
            },
          ],
        },
      ],
      options: [{ ignoreConditionalTests: false }],
    },
    {
      code: `
declare let x: string | null | undefined;
for (; x || 'foo';) {}
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 3,
          column: 10,
          endLine: 3,
          endColumn: 12,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: string | null | undefined;
for (; x ?? 'foo';) {}
      `,
            },
          ],
        },
      ],
      options: [{ ignoreConditionalTests: false }],
    },
    {
      code: `
declare let x: string | null | undefined;
for (; (x ||= 'foo');) {}
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.',
          line: 3,
          column: 11,
          endLine: 3,
          endColumn: 14,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??=`).',
              output: `
declare let x: string | null | undefined;
for (; (x ??= 'foo');) {}
      `,
            },
          ],
        },
      ],
      options: [{ ignoreConditionalTests: false }],
    },
    {
      code: `
declare let x: number | null | undefined;
for (; x || 'foo';) {}
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 3,
          column: 10,
          endLine: 3,
          endColumn: 12,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: number | null | undefined;
for (; x ?? 'foo';) {}
      `,
            },
          ],
        },
      ],
      options: [{ ignoreConditionalTests: false }],
    },
    {
      code: `
declare let x: number | null | undefined;
for (; (x ||= 'foo');) {}
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.',
          line: 3,
          column: 11,
          endLine: 3,
          endColumn: 14,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??=`).',
              output: `
declare let x: number | null | undefined;
for (; (x ??= 'foo');) {}
      `,
            },
          ],
        },
      ],
      options: [{ ignoreConditionalTests: false }],
    },
    {
      code: `
declare let x: boolean | null | undefined;
for (; x || 'foo';) {}
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 3,
          column: 10,
          endLine: 3,
          endColumn: 12,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: boolean | null | undefined;
for (; x ?? 'foo';) {}
      `,
            },
          ],
        },
      ],
      options: [{ ignoreConditionalTests: false }],
    },
    {
      code: `
declare let x: boolean | null | undefined;
for (; (x ||= 'foo');) {}
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.',
          line: 3,
          column: 11,
          endLine: 3,
          endColumn: 14,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??=`).',
              output: `
declare let x: boolean | null | undefined;
for (; (x ??= 'foo');) {}
      `,
            },
          ],
        },
      ],
      options: [{ ignoreConditionalTests: false }],
    },
    {
      code: `
declare let x: object | null | undefined;
for (; x || 'foo';) {}
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 3,
          column: 10,
          endLine: 3,
          endColumn: 12,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: object | null | undefined;
for (; x ?? 'foo';) {}
      `,
            },
          ],
        },
      ],
      options: [{ ignoreConditionalTests: false }],
    },
    {
      code: `
declare let x: object | null | undefined;
for (; (x ||= 'foo');) {}
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.',
          line: 3,
          column: 11,
          endLine: 3,
          endColumn: 14,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??=`).',
              output: `
declare let x: object | null | undefined;
for (; (x ??= 'foo');) {}
      `,
            },
          ],
        },
      ],
      options: [{ ignoreConditionalTests: false }],
    },
    {
      code: `
declare let x: string | null | undefined;
while (x || 'foo') {}
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 3,
          column: 10,
          endLine: 3,
          endColumn: 12,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: string | null | undefined;
while (x ?? 'foo') {}
      `,
            },
          ],
        },
      ],
      options: [{ ignoreConditionalTests: false }],
    },
    {
      code: `
declare let x: string | null | undefined;
while ((x ||= 'foo')) {}
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.',
          line: 3,
          column: 11,
          endLine: 3,
          endColumn: 14,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??=`).',
              output: `
declare let x: string | null | undefined;
while ((x ??= 'foo')) {}
      `,
            },
          ],
        },
      ],
      options: [{ ignoreConditionalTests: false }],
    },
    {
      code: `
declare let x: number | null | undefined;
while (x || 'foo') {}
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 3,
          column: 10,
          endLine: 3,
          endColumn: 12,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: number | null | undefined;
while (x ?? 'foo') {}
      `,
            },
          ],
        },
      ],
      options: [{ ignoreConditionalTests: false }],
    },
    {
      code: `
declare let x: number | null | undefined;
while ((x ||= 'foo')) {}
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.',
          line: 3,
          column: 11,
          endLine: 3,
          endColumn: 14,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??=`).',
              output: `
declare let x: number | null | undefined;
while ((x ??= 'foo')) {}
      `,
            },
          ],
        },
      ],
      options: [{ ignoreConditionalTests: false }],
    },
    {
      code: `
declare let x: boolean | null | undefined;
while (x || 'foo') {}
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 3,
          column: 10,
          endLine: 3,
          endColumn: 12,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: boolean | null | undefined;
while (x ?? 'foo') {}
      `,
            },
          ],
        },
      ],
      options: [{ ignoreConditionalTests: false }],
    },
    {
      code: `
declare let x: boolean | null | undefined;
while ((x ||= 'foo')) {}
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.',
          line: 3,
          column: 11,
          endLine: 3,
          endColumn: 14,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??=`).',
              output: `
declare let x: boolean | null | undefined;
while ((x ??= 'foo')) {}
      `,
            },
          ],
        },
      ],
      options: [{ ignoreConditionalTests: false }],
    },
    {
      code: `
declare let x: object | null | undefined;
while (x || 'foo') {}
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 3,
          column: 10,
          endLine: 3,
          endColumn: 12,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: object | null | undefined;
while (x ?? 'foo') {}
      `,
            },
          ],
        },
      ],
      options: [{ ignoreConditionalTests: false }],
    },
    {
      code: `
declare let x: object | null | undefined;
while ((x ||= 'foo')) {}
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.',
          line: 3,
          column: 11,
          endLine: 3,
          endColumn: 14,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??=`).',
              output: `
declare let x: object | null | undefined;
while ((x ??= 'foo')) {}
      `,
            },
          ],
        },
      ],
      options: [{ ignoreConditionalTests: false }],
    },

    // ignoreMixedLogicalExpressions
    {
      code: `
declare let a: string | null | undefined;
declare let b: string | null | undefined;
declare let c: string | null | undefined;
a || (b && c);
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 5,
          column: 3,
          endLine: 5,
          endColumn: 5,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let a: string | null | undefined;
declare let b: string | null | undefined;
declare let c: string | null | undefined;
a ?? (b && c);
      `,
            },
          ],
        },
      ],
      options: [{ ignoreMixedLogicalExpressions: false }],
    },
    {
      code: `
declare let a: number | null | undefined;
declare let b: number | null | undefined;
declare let c: number | null | undefined;
a || (b && c);
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 5,
          column: 3,
          endLine: 5,
          endColumn: 5,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let a: number | null | undefined;
declare let b: number | null | undefined;
declare let c: number | null | undefined;
a ?? (b && c);
      `,
            },
          ],
        },
      ],
      options: [{ ignoreMixedLogicalExpressions: false }],
    },
    {
      code: `
declare let a: boolean | null | undefined;
declare let b: boolean | null | undefined;
declare let c: boolean | null | undefined;
a || (b && c);
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 5,
          column: 3,
          endLine: 5,
          endColumn: 5,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let a: boolean | null | undefined;
declare let b: boolean | null | undefined;
declare let c: boolean | null | undefined;
a ?? (b && c);
      `,
            },
          ],
        },
      ],
      options: [{ ignoreMixedLogicalExpressions: false }],
    },
    {
      code: `
declare let a: object | null | undefined;
declare let b: object | null | undefined;
declare let c: object | null | undefined;
a || (b && c);
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 5,
          column: 3,
          endLine: 5,
          endColumn: 5,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let a: object | null | undefined;
declare let b: object | null | undefined;
declare let c: object | null | undefined;
a ?? (b && c);
      `,
            },
          ],
        },
      ],
      options: [{ ignoreMixedLogicalExpressions: false }],
    },
    {
      code: `
declare let a: string | null | undefined;
declare let b: string | null | undefined;
declare let c: string | null | undefined;
declare let d: string | null | undefined;
a || b || (c && d);
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 6,
          column: 3,
          endLine: 6,
          endColumn: 5,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let a: string | null | undefined;
declare let b: string | null | undefined;
declare let c: string | null | undefined;
declare let d: string | null | undefined;
(a ?? b) || (c && d);
      `,
            },
          ],
        },
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 6,
          column: 8,
          endLine: 6,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let a: string | null | undefined;
declare let b: string | null | undefined;
declare let c: string | null | undefined;
declare let d: string | null | undefined;
a || b ?? (c && d);
      `,
            },
          ],
        },
      ],
      options: [{ ignoreMixedLogicalExpressions: false }],
    },
    {
      code: `
declare let a: number | null | undefined;
declare let b: number | null | undefined;
declare let c: number | null | undefined;
declare let d: number | null | undefined;
a || b || (c && d);
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 6,
          column: 3,
          endLine: 6,
          endColumn: 5,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let a: number | null | undefined;
declare let b: number | null | undefined;
declare let c: number | null | undefined;
declare let d: number | null | undefined;
(a ?? b) || (c && d);
      `,
            },
          ],
        },
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 6,
          column: 8,
          endLine: 6,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let a: number | null | undefined;
declare let b: number | null | undefined;
declare let c: number | null | undefined;
declare let d: number | null | undefined;
a || b ?? (c && d);
      `,
            },
          ],
        },
      ],
      options: [{ ignoreMixedLogicalExpressions: false }],
    },
    {
      code: `
declare let a: boolean | null | undefined;
declare let b: boolean | null | undefined;
declare let c: boolean | null | undefined;
declare let d: boolean | null | undefined;
a || b || (c && d);
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 6,
          column: 3,
          endLine: 6,
          endColumn: 5,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let a: boolean | null | undefined;
declare let b: boolean | null | undefined;
declare let c: boolean | null | undefined;
declare let d: boolean | null | undefined;
(a ?? b) || (c && d);
      `,
            },
          ],
        },
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 6,
          column: 8,
          endLine: 6,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let a: boolean | null | undefined;
declare let b: boolean | null | undefined;
declare let c: boolean | null | undefined;
declare let d: boolean | null | undefined;
a || b ?? (c && d);
      `,
            },
          ],
        },
      ],
      options: [{ ignoreMixedLogicalExpressions: false }],
    },
    {
      code: `
declare let a: object | null | undefined;
declare let b: object | null | undefined;
declare let c: object | null | undefined;
declare let d: object | null | undefined;
a || b || (c && d);
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 6,
          column: 3,
          endLine: 6,
          endColumn: 5,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let a: object | null | undefined;
declare let b: object | null | undefined;
declare let c: object | null | undefined;
declare let d: object | null | undefined;
(a ?? b) || (c && d);
      `,
            },
          ],
        },
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 6,
          column: 8,
          endLine: 6,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let a: object | null | undefined;
declare let b: object | null | undefined;
declare let c: object | null | undefined;
declare let d: object | null | undefined;
a || b ?? (c && d);
      `,
            },
          ],
        },
      ],
      options: [{ ignoreMixedLogicalExpressions: false }],
    },
    {
      code: `
declare let a: string | null | undefined;
declare let b: string | null | undefined;
declare let c: string | null | undefined;
declare let d: string | null | undefined;
(a && b) || c || d;
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 6,
          column: 10,
          endLine: 6,
          endColumn: 12,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let a: string | null | undefined;
declare let b: string | null | undefined;
declare let c: string | null | undefined;
declare let d: string | null | undefined;
(a && (b) ?? c) || d;
      `,
            },
          ],
        },
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 6,
          column: 15,
          endLine: 6,
          endColumn: 17,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let a: string | null | undefined;
declare let b: string | null | undefined;
declare let c: string | null | undefined;
declare let d: string | null | undefined;
(a && b) || c ?? d;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreMixedLogicalExpressions: false }],
    },
    {
      code: `
declare let a: number | null | undefined;
declare let b: number | null | undefined;
declare let c: number | null | undefined;
declare let d: number | null | undefined;
(a && b) || c || d;
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 6,
          column: 10,
          endLine: 6,
          endColumn: 12,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let a: number | null | undefined;
declare let b: number | null | undefined;
declare let c: number | null | undefined;
declare let d: number | null | undefined;
(a && (b) ?? c) || d;
      `,
            },
          ],
        },
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 6,
          column: 15,
          endLine: 6,
          endColumn: 17,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let a: number | null | undefined;
declare let b: number | null | undefined;
declare let c: number | null | undefined;
declare let d: number | null | undefined;
(a && b) || c ?? d;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreMixedLogicalExpressions: false }],
    },
    {
      code: `
declare let a: boolean | null | undefined;
declare let b: boolean | null | undefined;
declare let c: boolean | null | undefined;
declare let d: boolean | null | undefined;
(a && b) || c || d;
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 6,
          column: 10,
          endLine: 6,
          endColumn: 12,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let a: boolean | null | undefined;
declare let b: boolean | null | undefined;
declare let c: boolean | null | undefined;
declare let d: boolean | null | undefined;
(a && (b) ?? c) || d;
      `,
            },
          ],
        },
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 6,
          column: 15,
          endLine: 6,
          endColumn: 17,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let a: boolean | null | undefined;
declare let b: boolean | null | undefined;
declare let c: boolean | null | undefined;
declare let d: boolean | null | undefined;
(a && b) || c ?? d;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreMixedLogicalExpressions: false }],
    },
    {
      code: `
declare let a: object | null | undefined;
declare let b: object | null | undefined;
declare let c: object | null | undefined;
declare let d: object | null | undefined;
(a && b) || c || d;
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 6,
          column: 10,
          endLine: 6,
          endColumn: 12,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let a: object | null | undefined;
declare let b: object | null | undefined;
declare let c: object | null | undefined;
declare let d: object | null | undefined;
(a && (b) ?? c) || d;
      `,
            },
          ],
        },
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 6,
          column: 15,
          endLine: 6,
          endColumn: 17,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let a: object | null | undefined;
declare let b: object | null | undefined;
declare let c: object | null | undefined;
declare let d: object | null | undefined;
(a && b) || c ?? d;
      `,
            },
          ],
        },
      ],
      options: [{ ignoreMixedLogicalExpressions: false }],
    },

    // should not false positive for functions inside conditional tests
    {
      code: `
declare let x: string | null | undefined;
if (() => x || 'foo') {
}
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 15,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: string | null | undefined;
if (() => x ?? 'foo') {
}
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
declare let x: string | null | undefined;
if (() => (x ||= 'foo')) {
}
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.',
          line: 3,
          column: 14,
          endLine: 3,
          endColumn: 17,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??=`).',
              output: `
declare let x: string | null | undefined;
if (() => (x ??= 'foo')) {
}
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
declare let x: number | null | undefined;
if (() => x || 'foo') {
}
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 15,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: number | null | undefined;
if (() => x ?? 'foo') {
}
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
declare let x: number | null | undefined;
if (() => (x ||= 'foo')) {
}
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.',
          line: 3,
          column: 14,
          endLine: 3,
          endColumn: 17,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??=`).',
              output: `
declare let x: number | null | undefined;
if (() => (x ??= 'foo')) {
}
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
declare let x: boolean | null | undefined;
if (() => x || 'foo') {
}
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 15,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: boolean | null | undefined;
if (() => x ?? 'foo') {
}
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
declare let x: boolean | null | undefined;
if (() => (x ||= 'foo')) {
}
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.',
          line: 3,
          column: 14,
          endLine: 3,
          endColumn: 17,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??=`).',
              output: `
declare let x: boolean | null | undefined;
if (() => (x ??= 'foo')) {
}
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
declare let x: object | null | undefined;
if (() => x || 'foo') {
}
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 15,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: object | null | undefined;
if (() => x ?? 'foo') {
}
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
declare let x: object | null | undefined;
if (() => (x ||= 'foo')) {
}
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.',
          line: 3,
          column: 14,
          endLine: 3,
          endColumn: 17,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??=`).',
              output: `
declare let x: object | null | undefined;
if (() => (x ??= 'foo')) {
}
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
declare let x: string | null | undefined;
if (
  function weird() {
    return x || 'foo';
  }
) {
}
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 5,
          column: 14,
          endLine: 5,
          endColumn: 16,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: string | null | undefined;
if (
  function weird() {
    return x ?? 'foo';
  }
) {
}
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
declare let x: string | null | undefined;
if (
  function weird() {
    return (x ||= 'foo');
  }
) {
}
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.',
          line: 5,
          column: 15,
          endLine: 5,
          endColumn: 18,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??=`).',
              output: `
declare let x: string | null | undefined;
if (
  function weird() {
    return (x ??= 'foo');
  }
) {
}
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
declare let x: number | null | undefined;
if (
  function weird() {
    return x || 'foo';
  }
) {
}
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 5,
          column: 14,
          endLine: 5,
          endColumn: 16,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: number | null | undefined;
if (
  function weird() {
    return x ?? 'foo';
  }
) {
}
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
declare let x: number | null | undefined;
if (
  function weird() {
    return (x ||= 'foo');
  }
) {
}
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.',
          line: 5,
          column: 15,
          endLine: 5,
          endColumn: 18,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??=`).',
              output: `
declare let x: number | null | undefined;
if (
  function weird() {
    return (x ??= 'foo');
  }
) {
}
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
declare let x: boolean | null | undefined;
if (
  function weird() {
    return x || 'foo';
  }
) {
}
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 5,
          column: 14,
          endLine: 5,
          endColumn: 16,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: boolean | null | undefined;
if (
  function weird() {
    return x ?? 'foo';
  }
) {
}
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
declare let x: boolean | null | undefined;
if (
  function weird() {
    return (x ||= 'foo');
  }
) {
}
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.',
          line: 5,
          column: 15,
          endLine: 5,
          endColumn: 18,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??=`).',
              output: `
declare let x: boolean | null | undefined;
if (
  function weird() {
    return (x ??= 'foo');
  }
) {
}
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
declare let x: object | null | undefined;
if (
  function weird() {
    return x || 'foo';
  }
) {
}
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 5,
          column: 14,
          endLine: 5,
          endColumn: 16,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: object | null | undefined;
if (
  function weird() {
    return x ?? 'foo';
  }
) {
}
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
declare let x: object | null | undefined;
if (
  function weird() {
    return (x ||= 'foo');
  }
) {
}
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.',
          line: 5,
          column: 15,
          endLine: 5,
          endColumn: 18,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??=`).',
              output: `
declare let x: object | null | undefined;
if (
  function weird() {
    return (x ??= 'foo');
  }
) {
}
      `,
            },
          ],
        },
      ],
    },
    // https://github.com/typescript-eslint/typescript-eslint/issues/1290
    {
      code: `
declare let a: string | null | undefined;
declare let b: string;
declare let c: string;
a || b || c;
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 5,
          column: 3,
          endLine: 5,
          endColumn: 5,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let a: string | null | undefined;
declare let b: string;
declare let c: string;
(a ?? b) || c;
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
declare let a: number | null | undefined;
declare let b: number;
declare let c: number;
a || b || c;
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 5,
          column: 3,
          endLine: 5,
          endColumn: 5,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let a: number | null | undefined;
declare let b: number;
declare let c: number;
(a ?? b) || c;
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
declare let a: boolean | null | undefined;
declare let b: boolean;
declare let c: boolean;
a || b || c;
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 5,
          column: 3,
          endLine: 5,
          endColumn: 5,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let a: boolean | null | undefined;
declare let b: boolean;
declare let c: boolean;
(a ?? b) || c;
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
declare let a: object | null | undefined;
declare let b: object;
declare let c: object;
a || b || c;
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 5,
          column: 3,
          endLine: 5,
          endColumn: 5,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let a: object | null | undefined;
declare let b: object;
declare let c: object;
(a ?? b) || c;
      `,
            },
          ],
        },
      ],
    },
    // default for missing option
    {
      code: `
declare let x: string | undefined;
x || y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 3,
          column: 3,
          endLine: 3,
          endColumn: 5,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: string | undefined;
x ?? y;
      `,
            },
          ],
        },
      ],
      options: [
        { ignorePrimitives: { bigint: true, boolean: true, number: true } },
      ],
    },
    {
      code: `
declare let x: number | undefined;
x || y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 3,
          column: 3,
          endLine: 3,
          endColumn: 5,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: number | undefined;
x ?? y;
      `,
            },
          ],
        },
      ],
      options: [
        { ignorePrimitives: { bigint: true, boolean: true, string: true } },
      ],
    },
    {
      code: `
declare let x: boolean | undefined;
x || y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 3,
          column: 3,
          endLine: 3,
          endColumn: 5,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: boolean | undefined;
x ?? y;
      `,
            },
          ],
        },
      ],
      options: [
        { ignorePrimitives: { bigint: true, number: true, string: true } },
      ],
    },
    {
      code: `
declare let x: bigint | undefined;
x || y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 3,
          column: 3,
          endLine: 3,
          endColumn: 5,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: bigint | undefined;
x ?? y;
      `,
            },
          ],
        },
      ],
      options: [
        { ignorePrimitives: { boolean: true, number: true, string: true } },
      ],
    },
    {
      code: `
declare let x: string | undefined;
x ? x : y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: string | undefined;
x ?? y;
      `,
            },
          ],
        },
      ],
      options: [
        { ignorePrimitives: { bigint: true, boolean: true, number: true } },
      ],
    },
    {
      code: `
declare let x: number | undefined;
x ? x : y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: number | undefined;
x ?? y;
      `,
            },
          ],
        },
      ],
      options: [
        { ignorePrimitives: { bigint: true, boolean: true, string: true } },
      ],
    },
    {
      code: `
declare let x: boolean | undefined;
x ? x : y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: boolean | undefined;
x ?? y;
      `,
            },
          ],
        },
      ],
      options: [
        { ignorePrimitives: { bigint: true, number: true, string: true } },
      ],
    },
    {
      code: `
declare let x: bigint | undefined;
x ? x : y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: bigint | undefined;
x ?? y;
      `,
            },
          ],
        },
      ],
      options: [
        { ignorePrimitives: { boolean: true, number: true, string: true } },
      ],
    },
    // falsy
    {
      code: `
declare let x: '' | undefined;
x || y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 3,
          column: 3,
          endLine: 3,
          endColumn: 5,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: '' | undefined;
x ?? y;
      `,
            },
          ],
        },
      ],
      options: [
        {
          ignorePrimitives: {
            bigint: true,
            boolean: true,
            number: true,
            string: false,
          },
        },
      ],
    },
    {
      code: `
declare let x: \`\` | undefined;
x || y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 3,
          column: 3,
          endLine: 3,
          endColumn: 5,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: \`\` | undefined;
x ?? y;
      `,
            },
          ],
        },
      ],
      options: [
        {
          ignorePrimitives: {
            bigint: true,
            boolean: true,
            number: true,
            string: false,
          },
        },
      ],
    },
    {
      code: `
declare let x: 0 | undefined;
x || y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 3,
          column: 3,
          endLine: 3,
          endColumn: 5,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: 0 | undefined;
x ?? y;
      `,
            },
          ],
        },
      ],
      options: [
        {
          ignorePrimitives: {
            bigint: true,
            boolean: true,
            number: false,
            string: true,
          },
        },
      ],
    },
    {
      code: `
declare let x: 0n | undefined;
x || y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 3,
          column: 3,
          endLine: 3,
          endColumn: 5,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: 0n | undefined;
x ?? y;
      `,
            },
          ],
        },
      ],
      options: [
        {
          ignorePrimitives: {
            bigint: false,
            boolean: true,
            number: true,
            string: true,
          },
        },
      ],
    },
    {
      code: `
declare let x: false | undefined;
x || y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 3,
          column: 3,
          endLine: 3,
          endColumn: 5,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: false | undefined;
x ?? y;
      `,
            },
          ],
        },
      ],
      options: [
        {
          ignorePrimitives: {
            bigint: true,
            boolean: false,
            number: true,
            string: true,
          },
        },
      ],
    },
    {
      code: `
declare let x: '' | undefined;
x ? x : y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: '' | undefined;
x ?? y;
      `,
            },
          ],
        },
      ],
      options: [
        {
          ignorePrimitives: {
            bigint: true,
            boolean: true,
            number: true,
            string: false,
          },
        },
      ],
    },
    {
      code: `
declare let x: \`\` | undefined;
x ? x : y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: \`\` | undefined;
x ?? y;
      `,
            },
          ],
        },
      ],
      options: [
        {
          ignorePrimitives: {
            bigint: true,
            boolean: true,
            number: true,
            string: false,
          },
        },
      ],
    },
    {
      code: `
declare let x: 0 | undefined;
x ? x : y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: 0 | undefined;
x ?? y;
      `,
            },
          ],
        },
      ],
      options: [
        {
          ignorePrimitives: {
            bigint: true,
            boolean: true,
            number: false,
            string: true,
          },
        },
      ],
    },
    {
      code: `
declare let x: 0n | undefined;
x ? x : y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: 0n | undefined;
x ?? y;
      `,
            },
          ],
        },
      ],
      options: [
        {
          ignorePrimitives: {
            bigint: false,
            boolean: true,
            number: true,
            string: true,
          },
        },
      ],
    },
    {
      code: `
declare let x: false | undefined;
x ? x : y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: false | undefined;
x ?? y;
      `,
            },
          ],
        },
      ],
      options: [
        {
          ignorePrimitives: {
            bigint: true,
            boolean: false,
            number: true,
            string: true,
          },
        },
      ],
    },
    // truthy
    {
      code: `
declare let x: 'a' | undefined;
x || y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 3,
          column: 3,
          endLine: 3,
          endColumn: 5,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: 'a' | undefined;
x ?? y;
      `,
            },
          ],
        },
      ],
      options: [
        {
          ignorePrimitives: {
            bigint: true,
            boolean: true,
            number: true,
            string: false,
          },
        },
      ],
    },
    {
      code: `
declare let x: \`hello\${'string'}\` | undefined;
x || y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 3,
          column: 3,
          endLine: 3,
          endColumn: 5,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: \`hello\${'string'}\` | undefined;
x ?? y;
      `,
            },
          ],
        },
      ],
      options: [
        {
          ignorePrimitives: {
            bigint: true,
            boolean: true,
            number: true,
            string: false,
          },
        },
      ],
    },
    {
      code: `
declare let x: 1 | undefined;
x || y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 3,
          column: 3,
          endLine: 3,
          endColumn: 5,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: 1 | undefined;
x ?? y;
      `,
            },
          ],
        },
      ],
      options: [
        {
          ignorePrimitives: {
            bigint: true,
            boolean: true,
            number: false,
            string: true,
          },
        },
      ],
    },
    {
      code: `
declare let x: 1n | undefined;
x || y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 3,
          column: 3,
          endLine: 3,
          endColumn: 5,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: 1n | undefined;
x ?? y;
      `,
            },
          ],
        },
      ],
      options: [
        {
          ignorePrimitives: {
            bigint: false,
            boolean: true,
            number: true,
            string: true,
          },
        },
      ],
    },
    {
      code: `
declare let x: true | undefined;
x || y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 3,
          column: 3,
          endLine: 3,
          endColumn: 5,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: true | undefined;
x ?? y;
      `,
            },
          ],
        },
      ],
      options: [
        {
          ignorePrimitives: {
            bigint: true,
            boolean: false,
            number: true,
            string: true,
          },
        },
      ],
    },
    {
      code: `
declare let x: 'a' | undefined;
x ? x : y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: 'a' | undefined;
x ?? y;
      `,
            },
          ],
        },
      ],
      options: [
        {
          ignorePrimitives: {
            bigint: true,
            boolean: true,
            number: true,
            string: false,
          },
        },
      ],
    },
    {
      code: `
declare let x: 'a' | undefined;
!x ? y : x;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 11,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: 'a' | undefined;
x ?? y;
      `,
            },
          ],
        },
      ],
      options: [
        {
          ignorePrimitives: {
            bigint: true,
            boolean: true,
            number: true,
            string: false,
          },
        },
      ],
    },
    {
      code: `
declare let x: \`hello\${'string'}\` | undefined;
x ? x : y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: \`hello\${'string'}\` | undefined;
x ?? y;
      `,
            },
          ],
        },
      ],
      options: [
        {
          ignorePrimitives: {
            bigint: true,
            boolean: true,
            number: true,
            string: false,
          },
        },
      ],
    },
    {
      code: `
declare let x: \`hello\${'string'}\` | undefined;
!x ? y : x;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 11,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: \`hello\${'string'}\` | undefined;
x ?? y;
      `,
            },
          ],
        },
      ],
      options: [
        {
          ignorePrimitives: {
            bigint: true,
            boolean: true,
            number: true,
            string: false,
          },
        },
      ],
    },
    {
      code: `
declare let x: 1 | undefined;
x ? x : y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: 1 | undefined;
x ?? y;
      `,
            },
          ],
        },
      ],
      options: [
        {
          ignorePrimitives: {
            bigint: true,
            boolean: true,
            number: false,
            string: true,
          },
        },
      ],
    },
    {
      code: `
declare let x: 1 | undefined;
!x ? y : x;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 11,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: 1 | undefined;
x ?? y;
      `,
            },
          ],
        },
      ],
      options: [
        {
          ignorePrimitives: {
            bigint: true,
            boolean: true,
            number: false,
            string: true,
          },
        },
      ],
    },
    {
      code: `
declare let x: 1n | undefined;
x ? x : y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: 1n | undefined;
x ?? y;
      `,
            },
          ],
        },
      ],
      options: [
        {
          ignorePrimitives: {
            bigint: false,
            boolean: true,
            number: true,
            string: true,
          },
        },
      ],
    },
    {
      code: `
declare let x: 1n | undefined;
!x ? y : x;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 11,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: 1n | undefined;
x ?? y;
      `,
            },
          ],
        },
      ],
      options: [
        {
          ignorePrimitives: {
            bigint: false,
            boolean: true,
            number: true,
            string: true,
          },
        },
      ],
    },
    {
      code: `
declare let x: true | undefined;
x ? x : y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: true | undefined;
x ?? y;
      `,
            },
          ],
        },
      ],
      options: [
        {
          ignorePrimitives: {
            bigint: true,
            boolean: false,
            number: true,
            string: true,
          },
        },
      ],
    },
    {
      code: `
declare let x: true | undefined;
!x ? y : x;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 11,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: true | undefined;
x ?? y;
      `,
            },
          ],
        },
      ],
      options: [
        {
          ignorePrimitives: {
            bigint: true,
            boolean: false,
            number: true,
            string: true,
          },
        },
      ],
    },
    // Unions of same primitive
    {
      code: `
declare let x: 'a' | 'b' | undefined;
x || y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 3,
          column: 3,
          endLine: 3,
          endColumn: 5,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: 'a' | 'b' | undefined;
x ?? y;
      `,
            },
          ],
        },
      ],
      options: [
        {
          ignorePrimitives: {
            bigint: true,
            boolean: true,
            number: true,
            string: false,
          },
        },
      ],
    },
    {
      code: `
declare let x: 'a' | \`b\` | undefined;
x || y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 3,
          column: 3,
          endLine: 3,
          endColumn: 5,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: 'a' | \`b\` | undefined;
x ?? y;
      `,
            },
          ],
        },
      ],
      options: [
        {
          ignorePrimitives: {
            bigint: true,
            boolean: true,
            number: true,
            string: false,
          },
        },
      ],
    },
    {
      code: `
declare let x: 0 | 1 | undefined;
x || y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 3,
          column: 3,
          endLine: 3,
          endColumn: 5,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: 0 | 1 | undefined;
x ?? y;
      `,
            },
          ],
        },
      ],
      options: [
        {
          ignorePrimitives: {
            bigint: true,
            boolean: true,
            number: false,
            string: true,
          },
        },
      ],
    },
    {
      code: `
declare let x: 1 | 2 | 3 | undefined;
x || y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 3,
          column: 3,
          endLine: 3,
          endColumn: 5,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: 1 | 2 | 3 | undefined;
x ?? y;
      `,
            },
          ],
        },
      ],
      options: [
        {
          ignorePrimitives: {
            bigint: true,
            boolean: true,
            number: false,
            string: true,
          },
        },
      ],
    },
    {
      code: `
declare let x: 0n | 1n | undefined;
x || y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 3,
          column: 3,
          endLine: 3,
          endColumn: 5,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: 0n | 1n | undefined;
x ?? y;
      `,
            },
          ],
        },
      ],
      options: [
        {
          ignorePrimitives: {
            bigint: false,
            boolean: true,
            number: true,
            string: true,
          },
        },
      ],
    },
    {
      code: `
declare let x: 1n | 2n | 3n | undefined;
x || y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 3,
          column: 3,
          endLine: 3,
          endColumn: 5,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: 1n | 2n | 3n | undefined;
x ?? y;
      `,
            },
          ],
        },
      ],
      options: [
        {
          ignorePrimitives: {
            bigint: false,
            boolean: true,
            number: true,
            string: true,
          },
        },
      ],
    },
    {
      code: `
declare let x: true | false | undefined;
x || y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 3,
          column: 3,
          endLine: 3,
          endColumn: 5,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: true | false | undefined;
x ?? y;
      `,
            },
          ],
        },
      ],
      options: [
        {
          ignorePrimitives: {
            bigint: true,
            boolean: false,
            number: true,
            string: true,
          },
        },
      ],
    },
    {
      code: `
declare let x: 'a' | 'b' | undefined;
x ? x : y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: 'a' | 'b' | undefined;
x ?? y;
      `,
            },
          ],
        },
      ],
      options: [
        {
          ignorePrimitives: {
            bigint: true,
            boolean: true,
            number: true,
            string: false,
          },
        },
      ],
    },
    {
      code: `
declare let x: 'a' | 'b' | undefined;
!x ? y : x;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 11,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: 'a' | 'b' | undefined;
x ?? y;
      `,
            },
          ],
        },
      ],
      options: [
        {
          ignorePrimitives: {
            bigint: true,
            boolean: true,
            number: true,
            string: false,
          },
        },
      ],
    },
    {
      code: `
declare let x: 'a' | \`b\` | undefined;
x ? x : y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: 'a' | \`b\` | undefined;
x ?? y;
      `,
            },
          ],
        },
      ],
      options: [
        {
          ignorePrimitives: {
            bigint: true,
            boolean: true,
            number: true,
            string: false,
          },
        },
      ],
    },
    {
      code: `
declare let x: 'a' | \`b\` | undefined;
!x ? y : x;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 11,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: 'a' | \`b\` | undefined;
x ?? y;
      `,
            },
          ],
        },
      ],
      options: [
        {
          ignorePrimitives: {
            bigint: true,
            boolean: true,
            number: true,
            string: false,
          },
        },
      ],
    },
    {
      code: `
declare let x: 0 | 1 | undefined;
x ? x : y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: 0 | 1 | undefined;
x ?? y;
      `,
            },
          ],
        },
      ],
      options: [
        {
          ignorePrimitives: {
            bigint: true,
            boolean: true,
            number: false,
            string: true,
          },
        },
      ],
    },
    {
      code: `
declare let x: 0 | 1 | undefined;
!x ? y : x;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 11,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: 0 | 1 | undefined;
x ?? y;
      `,
            },
          ],
        },
      ],
      options: [
        {
          ignorePrimitives: {
            bigint: true,
            boolean: true,
            number: false,
            string: true,
          },
        },
      ],
    },
    {
      code: `
declare let x: 1 | 2 | 3 | undefined;
x ? x : y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: 1 | 2 | 3 | undefined;
x ?? y;
      `,
            },
          ],
        },
      ],
      options: [
        {
          ignorePrimitives: {
            bigint: true,
            boolean: true,
            number: false,
            string: true,
          },
        },
      ],
    },
    {
      code: `
declare let x: 1 | 2 | 3 | undefined;
!x ? y : x;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 11,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: 1 | 2 | 3 | undefined;
x ?? y;
      `,
            },
          ],
        },
      ],
      options: [
        {
          ignorePrimitives: {
            bigint: true,
            boolean: true,
            number: false,
            string: true,
          },
        },
      ],
    },
    {
      code: `
declare let x: 0n | 1n | undefined;
x ? x : y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: 0n | 1n | undefined;
x ?? y;
      `,
            },
          ],
        },
      ],
      options: [
        {
          ignorePrimitives: {
            bigint: false,
            boolean: true,
            number: true,
            string: true,
          },
        },
      ],
    },
    {
      code: `
declare let x: 0n | 1n | undefined;
!x ? y : x;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 11,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: 0n | 1n | undefined;
x ?? y;
      `,
            },
          ],
        },
      ],
      options: [
        {
          ignorePrimitives: {
            bigint: false,
            boolean: true,
            number: true,
            string: true,
          },
        },
      ],
    },
    {
      code: `
declare let x: 1n | 2n | 3n | undefined;
x ? x : y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: 1n | 2n | 3n | undefined;
x ?? y;
      `,
            },
          ],
        },
      ],
      options: [
        {
          ignorePrimitives: {
            bigint: false,
            boolean: true,
            number: true,
            string: true,
          },
        },
      ],
    },
    {
      code: `
declare let x: 1n | 2n | 3n | undefined;
!x ? y : x;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 11,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: 1n | 2n | 3n | undefined;
x ?? y;
      `,
            },
          ],
        },
      ],
      options: [
        {
          ignorePrimitives: {
            bigint: false,
            boolean: true,
            number: true,
            string: true,
          },
        },
      ],
    },
    {
      code: `
declare let x: true | false | undefined;
x ? x : y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: true | false | undefined;
x ?? y;
      `,
            },
          ],
        },
      ],
      options: [
        {
          ignorePrimitives: {
            bigint: true,
            boolean: false,
            number: true,
            string: true,
          },
        },
      ],
    },
    {
      code: `
declare let x: true | false | undefined;
!x ? y : x;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 11,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: true | false | undefined;
x ?? y;
      `,
            },
          ],
        },
      ],
      options: [
        {
          ignorePrimitives: {
            bigint: true,
            boolean: false,
            number: true,
            string: true,
          },
        },
      ],
    },
    // Mixed unions
    {
      code: `
declare let x: 0 | 1 | 0n | 1n | undefined;
x || y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 3,
          column: 3,
          endLine: 3,
          endColumn: 5,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: 0 | 1 | 0n | 1n | undefined;
x ?? y;
      `,
            },
          ],
        },
      ],
      options: [
        {
          ignorePrimitives: {
            bigint: false,
            boolean: true,
            number: false,
            string: true,
          },
        },
      ],
    },
    {
      code: `
declare let x: true | false | null | undefined;
x || y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 3,
          column: 3,
          endLine: 3,
          endColumn: 5,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: true | false | null | undefined;
x ?? y;
      `,
            },
          ],
        },
      ],
      options: [
        {
          ignorePrimitives: {
            bigint: true,
            boolean: false,
            number: true,
            string: true,
          },
        },
      ],
    },
    {
      code: `
declare let x: 0 | 1 | 0n | 1n | undefined;
x ? x : y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: 0 | 1 | 0n | 1n | undefined;
x ?? y;
      `,
            },
          ],
        },
      ],
      options: [
        {
          ignorePrimitives: {
            bigint: false,
            boolean: true,
            number: false,
            string: true,
          },
        },
      ],
    },
    {
      code: `
declare let x: 0 | 1 | 0n | 1n | undefined;
!x ? y : x;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 11,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: 0 | 1 | 0n | 1n | undefined;
x ?? y;
      `,
            },
          ],
        },
      ],
      options: [
        {
          ignorePrimitives: {
            bigint: false,
            boolean: true,
            number: false,
            string: true,
          },
        },
      ],
    },
    {
      code: `
declare let x: true | false | null | undefined;
x ? x : y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: true | false | null | undefined;
x ?? y;
      `,
            },
          ],
        },
      ],
      options: [
        {
          ignorePrimitives: {
            bigint: true,
            boolean: false,
            number: true,
            string: true,
          },
        },
      ],
    },
    {
      code: `
declare let x: true | false | null | undefined;
!x ? y : x;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 11,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: true | false | null | undefined;
x ?? y;
      `,
            },
          ],
        },
      ],
      options: [
        {
          ignorePrimitives: {
            bigint: true,
            boolean: false,
            number: true,
            string: true,
          },
        },
      ],
    },
    {
      code: `
declare let x: null;
x || y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 3,
          column: 3,
          endLine: 3,
          endColumn: 5,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: null;
x ?? y;
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
const x = undefined;
x || y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 3,
          column: 3,
          endLine: 3,
          endColumn: 5,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
const x = undefined;
x ?? y;
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
null || y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 2,
          column: 6,
          endLine: 2,
          endColumn: 8,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
null ?? y;
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
undefined || y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 2,
          column: 11,
          endLine: 2,
          endColumn: 13,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
undefined ?? y;
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
enum Enum {
  A = 0,
  B = 1,
  C = 2,
}
declare let x: Enum | undefined;
x || y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 8,
          column: 3,
          endLine: 8,
          endColumn: 5,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
enum Enum {
  A = 0,
  B = 1,
  C = 2,
}
declare let x: Enum | undefined;
x ?? y;
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
enum Enum {
  A = 0,
  B = 1,
  C = 2,
}
declare let x: Enum.A | Enum.B | undefined;
x || y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 8,
          column: 3,
          endLine: 8,
          endColumn: 5,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
enum Enum {
  A = 0,
  B = 1,
  C = 2,
}
declare let x: Enum.A | Enum.B | undefined;
x ?? y;
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
enum Enum {
  A = 'a',
  B = 'b',
  C = 'c',
}
declare let x: Enum | undefined;
x || y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 8,
          column: 3,
          endLine: 8,
          endColumn: 5,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
enum Enum {
  A = 'a',
  B = 'b',
  C = 'c',
}
declare let x: Enum | undefined;
x ?? y;
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
enum Enum {
  A = 'a',
  B = 'b',
  C = 'c',
}
declare let x: Enum.A | Enum.B | undefined;
x || y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 8,
          column: 3,
          endLine: 8,
          endColumn: 5,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
enum Enum {
  A = 'a',
  B = 'b',
  C = 'c',
}
declare let x: Enum.A | Enum.B | undefined;
x ?? y;
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
let a: string | true | undefined;
let b: string | boolean | undefined;
let c: boolean | undefined;

const x = Boolean(a || b);
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 6,
          column: 21,
          endLine: 6,
          endColumn: 23,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
let a: string | true | undefined;
let b: string | boolean | undefined;
let c: boolean | undefined;

const x = Boolean(a ?? b);
      `,
            },
          ],
        },
      ],
      options: [{ ignoreBooleanCoercion: false }],
    },
    {
      code: `
function outer() {
  const Boolean = (x: unknown) => x;

  return (a: string | null, b: string) => Boolean(a || b);
}
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 5,
          column: 53,
          endLine: 5,
          endColumn: 55,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
function outer() {
  const Boolean = (x: unknown) => x;

  return (a: string | null, b: string) => Boolean(a ?? b);
}
      `,
            },
          ],
        },
      ],
      options: [{ ignoreBooleanCoercion: true }],
    },
    {
      code: `
let a: string | true | undefined;
let b: string | boolean | undefined;

const x = String(a || b);
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 5,
          column: 20,
          endLine: 5,
          endColumn: 22,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
let a: string | true | undefined;
let b: string | boolean | undefined;

const x = String(a ?? b);
      `,
            },
          ],
        },
      ],
      options: [{ ignoreBooleanCoercion: true }],
    },
    {
      code: `
let a: string | true | undefined;
let b: string | boolean | undefined;

const x = Boolean(() => a || b);
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 5,
          column: 27,
          endLine: 5,
          endColumn: 29,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
let a: string | true | undefined;
let b: string | boolean | undefined;

const x = Boolean(() => a ?? b);
      `,
            },
          ],
        },
      ],
      options: [{ ignoreBooleanCoercion: true }],
    },
    {
      code: `
let a: string | true | undefined;
let b: string | boolean | undefined;

const x = Boolean(function weird() {
  return a || b;
});
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 6,
          column: 12,
          endLine: 6,
          endColumn: 14,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
let a: string | true | undefined;
let b: string | boolean | undefined;

const x = Boolean(function weird() {
  return a ?? b;
});
      `,
            },
          ],
        },
      ],
      options: [{ ignoreBooleanCoercion: true }],
    },
    {
      code: `
let a: string | true | undefined;
let b: string | boolean | undefined;

declare function f(x: unknown): unknown;

const x = Boolean(f(a || b));
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 7,
          column: 23,
          endLine: 7,
          endColumn: 25,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
let a: string | true | undefined;
let b: string | boolean | undefined;

declare function f(x: unknown): unknown;

const x = Boolean(f(a ?? b));
      `,
            },
          ],
        },
      ],
      options: [{ ignoreBooleanCoercion: true }],
    },
    {
      code: `
let a: string | true | undefined;
let b: string | boolean | undefined;

const x = Boolean(1 + (a || b));
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 5,
          column: 26,
          endLine: 5,
          endColumn: 28,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
let a: string | true | undefined;
let b: string | boolean | undefined;

const x = Boolean(1 + (a ?? b));
      `,
            },
          ],
        },
      ],
      options: [{ ignoreBooleanCoercion: true }],
    },
    {
      code: `
let a: string | true | undefined;
let b: string | boolean | undefined;

const x = Boolean(a ? a : b);
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 5,
          column: 19,
          endLine: 5,
          endColumn: 28,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
let a: string | true | undefined;
let b: string | boolean | undefined;

const x = Boolean(a ?? b);
      `,
            },
          ],
        },
      ],
      options: [{ ignoreBooleanCoercion: true }],
    },
    {
      code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;

const test = Boolean(!a ? b : a);
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 5,
          column: 22,
          endLine: 5,
          endColumn: 32,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;

const test = Boolean(a ?? b);
      `,
            },
          ],
        },
      ],
      options: [{ ignoreBooleanCoercion: true }],
    },
    {
      code: `
let a: string | true | undefined;
let b: string | boolean | undefined;

declare function f(x: unknown): unknown;

if (f(a || b)) {
}
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 7,
          column: 9,
          endLine: 7,
          endColumn: 11,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
let a: string | true | undefined;
let b: string | boolean | undefined;

declare function f(x: unknown): unknown;

if (f(a ?? b)) {
}
      `,
            },
          ],
        },
      ],
      options: [{ ignoreConditionalTests: true }],
    },
    {
      code: `
declare const a: string | undefined;
declare const b: string;

if (+(a || b)) {
}
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 5,
          column: 9,
          endLine: 5,
          endColumn: 11,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare const a: string | undefined;
declare const b: string;

if (+(a ?? b)) {
}
      `,
            },
          ],
        },
      ],
      options: [{ ignoreConditionalTests: true }],
    },
    {
      code: `
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBox: Box | undefined;

defaultBox || getFallbackBox();
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 8,
          column: 12,
          endLine: 8,
          endColumn: 14,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBox: Box | undefined;

defaultBox ?? getFallbackBox();
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBox: Box | undefined;

defaultBox ? defaultBox : getFallbackBox();
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 8,
          column: 1,
          endLine: 8,
          endColumn: 43,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBox: Box | undefined;

defaultBox ?? getFallbackBox();
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBoxOptional: { a?: { b?: Box | undefined } };

defaultBoxOptional.a?.b != null ? defaultBoxOptional.a?.b : getFallbackBox();
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 8,
          column: 1,
          endLine: 8,
          endColumn: 77,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBoxOptional: { a?: { b?: Box | undefined } };

defaultBoxOptional.a?.b ?? getFallbackBox();
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare const x: any;
declare const y: any;
x || y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 4,
          column: 3,
          endLine: 4,
          endColumn: 5,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare const x: any;
declare const y: any;
x ?? y;
      `,
            },
          ],
        },
      ],
    },

    {
      code: `
declare const x: unknown;
declare const y: any;
x || y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.',
          line: 4,
          column: 3,
          endLine: 4,
          endColumn: 5,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare const x: unknown;
declare const y: any;
x ?? y;
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBoxOptional: { a?: { b?: Box | undefined } };

defaultBoxOptional.a?.b != null ? defaultBoxOptional.a.b : getFallbackBox();
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 8,
          column: 1,
          endLine: 8,
          endColumn: 76,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBoxOptional: { a?: { b?: Box | undefined } };

defaultBoxOptional.a?.b ?? getFallbackBox();
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBoxOptional: { a?: { b?: Box | undefined } };

defaultBoxOptional.a?.b ? defaultBoxOptional.a?.b : getFallbackBox();
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 8,
          column: 1,
          endLine: 8,
          endColumn: 69,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBoxOptional: { a?: { b?: Box | undefined } };

defaultBoxOptional.a?.b ?? getFallbackBox();
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBoxOptional: { a?: { b?: Box | undefined } };

defaultBoxOptional.a?.b ? defaultBoxOptional.a.b : getFallbackBox();
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 8,
          column: 1,
          endLine: 8,
          endColumn: 68,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBoxOptional: { a?: { b?: Box | undefined } };

defaultBoxOptional.a?.b ?? getFallbackBox();
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBoxOptional: { a?: { b?: Box | undefined } };

defaultBoxOptional.a?.b !== undefined
  ? defaultBoxOptional.a?.b
  : getFallbackBox();
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 8,
          column: 1,
          endLine: 10,
          endColumn: 21,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBoxOptional: { a?: { b?: Box | undefined } };

defaultBoxOptional.a?.b ?? getFallbackBox();
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBoxOptional: { a?: { b?: Box | undefined } };

defaultBoxOptional.a?.b !== undefined
  ? defaultBoxOptional.a.b
  : getFallbackBox();
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 8,
          column: 1,
          endLine: 10,
          endColumn: 21,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBoxOptional: { a?: { b?: Box | undefined } };

defaultBoxOptional.a?.b ?? getFallbackBox();
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBoxOptional: { a?: { b?: Box | undefined } };

defaultBoxOptional.a?.b !== undefined && defaultBoxOptional.a?.b !== null
  ? defaultBoxOptional.a?.b
  : getFallbackBox();
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 8,
          column: 1,
          endLine: 10,
          endColumn: 21,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBoxOptional: { a?: { b?: Box | undefined } };

defaultBoxOptional.a?.b ?? getFallbackBox();
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBoxOptional: { a?: { b?: Box | undefined } };

defaultBoxOptional.a?.b !== undefined && defaultBoxOptional.a?.b !== null
  ? defaultBoxOptional.a.b
  : getFallbackBox();
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 8,
          column: 1,
          endLine: 10,
          endColumn: 21,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBoxOptional: { a?: { b?: Box | undefined } };

defaultBoxOptional.a?.b ?? getFallbackBox();
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
declare let x: unknown;
declare let y: number;
!x ? y : x;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 4,
          column: 1,
          endLine: 4,
          endColumn: 11,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: unknown;
declare let y: number;
x ?? y;
      `,
            },
          ],
        },
      ],
    },

    {
      code: `
declare let x: unknown;
declare let y: number;
x ? x : y;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 4,
          column: 1,
          endLine: 4,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: unknown;
declare let y: number;
x ?? y;
      `,
            },
          ],
        },
      ],
    },

    {
      code: `
declare let x: { n: unknown };
!x.n ? y : x.n;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 15,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: { n: unknown };
x.n ?? y;
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
declare let x: { a: string } | null;

x?.['a'] != null ? x['a'] : 'foo';
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 4,
          column: 1,
          endLine: 4,
          endColumn: 34,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: { a: string } | null;

x?.['a'] ?? 'foo';
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
declare let x: { a: string } | null;

x?.['a'] != null ? x.a : 'foo';
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 4,
          column: 1,
          endLine: 4,
          endColumn: 31,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: { a: string } | null;

x?.['a'] ?? 'foo';
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
declare let x: { a: string } | null;

x?.a != null ? x['a'] : 'foo';
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 4,
          column: 1,
          endLine: 4,
          endColumn: 30,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare let x: { a: string } | null;

x?.a ?? 'foo';
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
const a = 'b';
declare let x: { a: string; b: string } | null;

x?.[a] != null ? x[a] : 'foo';
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 5,
          column: 1,
          endLine: 5,
          endColumn: 30,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
const a = 'b';
declare let x: { a: string; b: string } | null;

x?.[a] ?? 'foo';
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (!foo) {
    foo = makeFoo();
  }
}
      `,
      errors: [
        {
          messageId: 'preferNullishOverAssignment',
          message: 'Prefer using nullish coalescing operator (`??=`) instead of an assignment expression, as it is simpler to read.',
          line: 6,
          column: 3,
          endLine: 8,
          endColumn: 4,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??=`).',
              output: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  foo ??= makeFoo();
}
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (foo == null) {
    foo = makeFoo();
  }
}
      `,
      errors: [
        {
          messageId: 'preferNullishOverAssignment',
          message: 'Prefer using nullish coalescing operator (`??=`) instead of an assignment expression, as it is simpler to read.',
          line: 6,
          column: 3,
          endLine: 8,
          endColumn: 4,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??=`).',
              output: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  foo ??= makeFoo();
}
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (foo == null) {
    foo ??= makeFoo();
  }
}
      `,
      errors: [
        {
          messageId: 'preferNullishOverAssignment',
          message: 'Prefer using nullish coalescing operator (`??=`) instead of an assignment expression, as it is simpler to read.',
          line: 6,
          column: 3,
          endLine: 8,
          endColumn: 4,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??=`).',
              output: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  foo ??= makeFoo();
}
      `,
            },
          ],
        },
      ],
    },
    // Multiple errors
    /* eslint-disable @typescript-eslint/internal/no-multiple-lines-of-errors */
    {
      code: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (foo == null) {
    foo ||= makeFoo();
  }
}
      `,
      errors: [
        {
          messageId: 'preferNullishOverAssignment',
          message: 'Prefer using nullish coalescing operator (`??=`) instead of an assignment expression, as it is simpler to read.',
          line: 6,
          column: 3,
          endLine: 8,
          endColumn: 4,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??=`).',
              output: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  foo ??= makeFoo();
}
      `,
            },
          ],
        },
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.',
          line: 7,
          column: 9,
          endLine: 7,
          endColumn: 12,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??=`).',
              output: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (foo == null) {
    foo ??= makeFoo();
  }
}
      `,
            },
          ],
        },
      ],
    },
    /* eslint-enable @typescript-eslint/internal/no-multiple-lines-of-errors */
    {
      code: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (foo === null) {
    foo = makeFoo();
  }
}
      `,
      errors: [
        {
          messageId: 'preferNullishOverAssignment',
          message: 'Prefer using nullish coalescing operator (`??=`) instead of an assignment expression, as it is simpler to read.',
          line: 6,
          column: 3,
          endLine: 8,
          endColumn: 4,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??=`).',
              output: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  foo ??= makeFoo();
}
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (foo == null) foo = makeFoo();
  const bar = 42;
  return bar;
}
      `,
      errors: [
        {
          messageId: 'preferNullishOverAssignment',
          message: 'Prefer using nullish coalescing operator (`??=`) instead of an assignment expression, as it is simpler to read.',
          line: 6,
          column: 3,
          endLine: 6,
          endColumn: 36,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??=`).',
              output: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  foo ??= makeFoo();
  const bar = 42;
  return bar;
}
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (foo == null) foo ??= makeFoo();
  const bar = 42;
  return bar;
}
      `,
      errors: [
        {
          messageId: 'preferNullishOverAssignment',
          message: 'Prefer using nullish coalescing operator (`??=`) instead of an assignment expression, as it is simpler to read.',
          line: 6,
          column: 3,
          endLine: 6,
          endColumn: 38,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??=`).',
              output: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  foo ??= makeFoo();
  const bar = 42;
  return bar;
}
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (foo == null) foo ||= makeFoo();
  const bar = 42;
  return bar;
}
      `,
      errors: [
        {
          messageId: 'preferNullishOverAssignment',
          message: 'Prefer using nullish coalescing operator (`??=`) instead of an assignment expression, as it is simpler to read.',
          line: 6,
          column: 3,
          endLine: 6,
          endColumn: 38,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??=`).',
              output: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  foo ??= makeFoo();
  const bar = 42;
  return bar;
}
      `,
            },
          ],
        },
        {
          messageId: 'preferNullishOverOr',
          message: 'Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.',
          line: 6,
          column: 24,
          endLine: 6,
          endColumn: 27,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??=`).',
              output: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (foo == null) foo ??= makeFoo();
  const bar = 42;
  return bar;
}
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
declare let foo: { a: string } | undefined;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (foo === undefined) {
    foo = makeFoo();
  }
}
      `,
      errors: [
        {
          messageId: 'preferNullishOverAssignment',
          message: 'Prefer using nullish coalescing operator (`??=`) instead of an assignment expression, as it is simpler to read.',
          line: 6,
          column: 3,
          endLine: 8,
          endColumn: 4,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??=`).',
              output: `
declare let foo: { a: string } | undefined;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  foo ??= makeFoo();
}
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
declare let foo: { a: string } | null | undefined;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (foo === undefined || foo === null) {
    foo = makeFoo();
  }
}
      `,
      errors: [
        {
          messageId: 'preferNullishOverAssignment',
          message: 'Prefer using nullish coalescing operator (`??=`) instead of an assignment expression, as it is simpler to read.',
          line: 6,
          column: 3,
          endLine: 8,
          endColumn: 4,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??=`).',
              output: `
declare let foo: { a: string } | null | undefined;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  foo ??= makeFoo();
}
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
declare let foo: { a: string } | null;
declare function makeFoo(): string;

function lazyInitialize() {
  if (foo.a == null) {
    foo.a = makeFoo();
  }
}
      `,
      errors: [
        {
          messageId: 'preferNullishOverAssignment',
          message: 'Prefer using nullish coalescing operator (`??=`) instead of an assignment expression, as it is simpler to read.',
          line: 6,
          column: 3,
          endLine: 8,
          endColumn: 4,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??=`).',
              output: `
declare let foo: { a: string } | null;
declare function makeFoo(): string;

function lazyInitialize() {
  foo.a ??= makeFoo();
}
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
declare let foo: { a: string } | null;
declare function makeFoo(): string;

function lazyInitialize() {
  if (foo?.a == null) {
    foo.a = makeFoo();
  }
}
      `,
      errors: [
        {
          messageId: 'preferNullishOverAssignment',
          message: 'Prefer using nullish coalescing operator (`??=`) instead of an assignment expression, as it is simpler to read.',
          line: 6,
          column: 3,
          endLine: 8,
          endColumn: 4,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??=`).',
              output: `
declare let foo: { a: string } | null;
declare function makeFoo(): string;

function lazyInitialize() {
  foo.a ??= makeFoo();
}
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
declare let foo: string | null;
declare function makeFoo(): string;

function lazyInitialize() {
  if (foo == null) {
    // comment
    foo = makeFoo();
  }
}
      `,
      errors: [
        {
          messageId: 'preferNullishOverAssignment',
          message: 'Prefer using nullish coalescing operator (`??=`) instead of an assignment expression, as it is simpler to read.',
          line: 6,
          column: 3,
          endLine: 9,
          endColumn: 4,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??=`).',
              output: `
declare let foo: string | null;
declare function makeFoo(): string;

function lazyInitialize() {
  // comment
foo ??= makeFoo();
}
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
declare let foo: string | null;
declare function makeFoo(): string;

if (foo == null) {
  // comment before 1
  /* comment before 2 */
  /* comment before 3
    which is multiline
  */
  /**
   * comment before 4
   * which is also multiline
   */
  foo = makeFoo(); // comment inline
  // comment after 1
  /* comment after 2 */
  /* comment after 3
    which is multiline
  */
  /**
   * comment after 4
   * which is also multiline
   */
}
      `,
      errors: [
        {
          messageId: 'preferNullishOverAssignment',
          message: 'Prefer using nullish coalescing operator (`??=`) instead of an assignment expression, as it is simpler to read.',
          line: 5,
          column: 1,
          endLine: 25,
          endColumn: 2,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??=`).',
              output: `
declare let foo: string | null;
declare function makeFoo(): string;

// comment before 1
/* comment before 2 */
/* comment before 3
    which is multiline
  */
/**
   * comment before 4
   * which is also multiline
   */
foo ??= makeFoo(); // comment inline
// comment after 1
/* comment after 2 */
/* comment after 3
    which is multiline
  */
/**
   * comment after 4
   * which is also multiline
   */
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
declare let foo: string | null;
declare function makeFoo(): string;

if (foo == null) /* comment before 1 */ /* comment before 2 */ foo = makeFoo(); // comment inline
      `,
      errors: [
        {
          messageId: 'preferNullishOverAssignment',
          message: 'Prefer using nullish coalescing operator (`??=`) instead of an assignment expression, as it is simpler to read.',
          line: 5,
          column: 1,
          endLine: 5,
          endColumn: 80,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??=`).',
              output: `
declare let foo: string | null;
declare function makeFoo(): string;

/* comment before 1 */ /* comment before 2 */ foo ??= makeFoo(); // comment inline
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
declare let foo: { a: string | null };
declare function makeString(): string;

function weirdParens() {
  if (((((foo.a)) == null))) {
    ((((((((foo).a))))) = makeString()));
  }
}
      `,
      errors: [
        {
          messageId: 'preferNullishOverAssignment',
          message: 'Prefer using nullish coalescing operator (`??=`) instead of an assignment expression, as it is simpler to read.',
          line: 6,
          column: 3,
          endLine: 8,
          endColumn: 4,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??=`).',
              output: `
declare let foo: { a: string | null };
declare function makeString(): string;

function weirdParens() {
  ((foo).a) ??= makeString();
}
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
let a: string | undefined;
let b: { message: string } | undefined;

const foo = a ? a : b ? 1 : 2;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 5,
          column: 13,
          endLine: 5,
          endColumn: 30,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
let a: string | undefined;
let b: { message: string } | undefined;

const foo = a ?? (b ? 1 : 2);
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
let a: string | undefined;
let b: { message: string } | undefined;

const foo = a ? a : (b ? 1 : 2);
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 5,
          column: 13,
          endLine: 5,
          endColumn: 32,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
let a: string | undefined;
let b: { message: string } | undefined;

const foo = a ?? (b ? 1 : 2);
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
declare const c: string | null;
c !== null ? c : c ? 1 : 2;
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 27,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
declare const c: string | null;
c ?? (c ? 1 : 2);
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
function test<T extends string>(val: T | undefined) {
  return val !== undefined ? val : 'foo';
}
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 10,
          endLine: 3,
          endColumn: 41,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
function test<T extends string>(val: T | undefined) {
  return val ?? 'foo';
}
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
function test<T extends {}>(val: T | undefined) {
  return val !== undefined ? val : 'fallback';
}
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 10,
          endLine: 3,
          endColumn: 46,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
function test<T extends {}>(val: T | undefined) {
  return val ?? 'fallback';
}
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
function test<T>(val: T | undefined) {
  return val != undefined ? val : 'fallback';
}
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 10,
          endLine: 3,
          endColumn: 45,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
function test<T>(val: T | undefined) {
  return val ?? 'fallback';
}
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
function test<T>(val: T | undefined) {
  return val !== undefined && val !== null ? val : 'fallback';
}
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 10,
          endLine: 3,
          endColumn: 62,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
function test<T>(val: T | undefined) {
  return val ?? 'fallback';
}
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
function test<T extends string, U extends T>(val: U | undefined) {
  return val !== undefined ? val : 'fallback';
}
      `,
      errors: [
        {
          messageId: 'preferNullishOverTernary',
          message: 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.',
          line: 3,
          column: 10,
          endLine: 3,
          endColumn: 46,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??`).',
              output: `
function test<T extends string, U extends T>(val: U | undefined) {
  return val ?? 'fallback';
}
      `,
            },
          ],
        },
      ],
      options: [{ ignoreTernaryTests: false }],
    },
    {
      code: `
function test<T extends string>(val: T | undefined, fallback: T) {
  if (val === undefined) {
    val = fallback;
  }
  return val;
}
      `,
      errors: [
        {
          messageId: 'preferNullishOverAssignment',
          message: 'Prefer using nullish coalescing operator (`??=`) instead of an assignment expression, as it is simpler to read.',
          line: 3,
          column: 3,
          endLine: 5,
          endColumn: 4,
          suggestions: [
            {
              messageId: 'suggestNullish',
              desc: 'Fix to nullish coalescing operator (`??=`).',
              output: `
function test<T extends string>(val: T | undefined, fallback: T) {
  val ??= fallback;
  return val;
}
      `,
            },
          ],
        },
      ],
      options: [{ ignoreIfStatements: false }],
    },
  ],
});
