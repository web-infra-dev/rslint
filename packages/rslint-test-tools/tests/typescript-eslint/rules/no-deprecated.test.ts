// All 100 valid and 194 invalid upstream cases: @typescript-eslint/eslint-plugin v8.72.2.
// https://github.com/typescript-eslint/typescript-eslint/blob/v8.72.2/packages/eslint-plugin/tests/rules/no-deprecated.test.ts
// Inputs, options and all diagnostic ranges are preserved. Messages are rendered
// from upstream data. This rule has no autofixes or suggestions; their absence
// is checked by the shared tester.
import { RuleTester } from '@rslint/rule-tester';
import { noFormat } from '@typescript-eslint/rule-tester';
import fs from 'node:fs';
import path from 'node:path';
import { createRequire } from 'node:module';

const require = createRequire(import.meta.url);
const fixtureFiles: Record<string, string> = {};
function collect(directory: string, prefix: string): void {
  for (const entry of fs.readdirSync(directory, { withFileTypes: true })) {
    const filename = path.join(directory, entry.name);
    const target = prefix + entry.name;
    if (entry.isDirectory()) collect(filename, target + '/');
    else if (/\.(?:json|ts|tsx)$/.test(entry.name)) {
      fixtureFiles[target] = fs.readFileSync(filename, 'utf8');
    }
  }
}
// These aliases pin the declaration environment from upstream's lockfile.
for (const [alias, name] of [
  ['@types/node24', '@types/node'],
  ['@types/react18', '@types/react'],
]) {
  const packageFile = require.resolve(alias + '/package.json');
  collect(path.dirname(packageFile), 'node_modules/' + name + '/');
  const packageRequire = createRequire(packageFile);
  const dependencies =
    JSON.parse(fs.readFileSync(packageFile, 'utf8')).dependencies ?? {};
  for (const dependency of Object.keys(dependencies)) {
    collect(
      path.dirname(packageRequire.resolve(dependency + '/package.json')),
      'node_modules/' + dependency + '/',
    );
  }
}
collect(path.join(import.meta.dirname, '../fixtures/no-deprecated'), '');
fixtureFiles['tsconfig.json'] = JSON.stringify({
  compilerOptions: {
    jsx: 'preserve',
    target: 'es2015',
    module: 'commonjs',
    strict: true,
    types: ['node', 'react'],
    lib: ['es2015', 'es2017', 'esnext'],
    experimentalDecorators: true,
    esModuleInterop: true,
  },
  include: ['file.tsx', 'deprecated.ts', 'react.tsx'],
});
fixtureFiles['tsconfig.moduleResolution-node16.json'] = JSON.stringify({
  extends: './tsconfig.json',
  compilerOptions: { module: 'node16', moduleResolution: 'node16' },
});
const ruleTester = new RuleTester({
  plugin: '@typescript-eslint',
  filename: 'file.tsx',
  fixtureFiles,
  languageOptions: { parserOptions: { project: './tsconfig.json' } },
});

ruleTester.run('no-deprecated', {
  valid: [
    '/** @deprecated */ var a;',
    '/** @deprecated */ var a = 1;',
    '/** @deprecated */ let a;',
    '/** @deprecated */ let a = 1;',
    '/** @deprecated */ const a = 1;',
    '/** @deprecated */ declare var a: number;',
    '/** @deprecated */ declare let a: number;',
    '/** @deprecated */ declare const a: number;',
    '/** @deprecated */ export var a = 1;',
    '/** @deprecated */ export let a = 1;',
    '/** @deprecated */ export const a = 1;',
    'const [/** @deprecated */ a] = [b];',
    'const [/** @deprecated */ a] = b;',
    `
const a = {
  b: 1,
  /** @deprecated */ c: 2,
};

a.b;
    `,
    `
const a = {
  b: 1,
  /** @deprecated */ c: 2,
};

a['b'];
    `,
    `
const a = {
  b: 1,
  /** @deprecated */ c: 2,
};

a['b' + 'c'];
    `,
    `
const a = {
  b: 1,
  /** @deprecated */ c: 2,
};

const key = 'b';

a[key];
    `,
    `
const a = {
  b: 1,
  /** @deprecated */ c: 2,
};

a?.b;
    `,
    `
const a = {
  b: 1,
  /** @deprecated */ c: 2,
};

a?.['b'];
    `,
    `
declare const a: {
  b: 1;
  /** @deprecated */ c: 2;
};

a.b;
    `,
    `
declare const a: {
  b: 1;
  /** @deprecated */ c: 2;
};

a['b'];
    `,
    `
declare const a: {
  b: 1;
  /** @deprecated */ c: 2;
};

a[\`\${'b'}\`];
    `,
    `
declare const a: {
  b: 1;
  /** @deprecated */ c: 2;
};

const key = 'b';

a[\`\${key}\`];
    `,
    `
declare const a: {
  /** @deprecated */ c: 1;
  cc: 2;
};

const key = 'c';

a[\`\${key + key}\`];
    `,
    `
declare const a: {
  /** @deprecated */ c: 1;
  cc: 2;
};

const key = 'c';

a[\`\${key}\${key}\`];
    `,
    `
class A {
  b: 1;
  /** @deprecated */ c: 2;
}

new A().b;
    `,
    `
class A {
  b: 1;
  /** @deprecated */ c: 2;
}

new A()['b'];
    `,
    `
class A {
  b: 1;
  /** @deprecated */ c: 2;
}
const key = 'b';

new A()[b];
    `,
    `
class A {
  c: 1;
}
class B {
  /** @deprecated */ c: 2;
}

new A()['c'];
    `,
    `
class A {
  b: () => {};
  /** @deprecated */ c: () => {};
}

new A()['b']();
    `,
    {
      code: `
class A {
  accessor b: 1;
  /** @deprecated */ accessor c: 2;
}

new A().b;
      `,
      // `accessor` keyword requires TS 4.9
      dependencyConstraints: { typescript: '4.9' },
    },
    {
      code: `
class A {
  accessor b: 1;
  /** @deprecated */ accessor c: 2;
}

new A()['b'];
      `,
      // `accessor` keyword requires TS 4.9
      dependencyConstraints: { typescript: '4.9' },
    },
    `
declare class A {
  /** @deprecated */
  static b: string;
  static c: string;
}

A.c;
    `,
    `
declare class A {
  /** @deprecated */
  static b: string;
  static c: string;
}

A['c'];
    `,
    {
      code: `
declare class A {
  /** @deprecated */
  static accessor b: string;
  static accessor c: string;
}

A.c;
      `,
      // `accessor` keyword requires TS 4.9
      dependencyConstraints: { typescript: '4.9' },
    },
    {
      code: `
declare class A {
  /** @deprecated */
  static accessor b: string;
  static accessor c: string;
}

A['c'];
      `,
      // `accessor` keyword requires TS 4.9
      dependencyConstraints: { typescript: '4.9' },
    },
    `
namespace A {
  /** @deprecated */
  export const b = '';
  export const c = '';
}

A.c;
    `,
    `
namespace A {
  /** @deprecated */
  export const b = '';
  export const c = '';
}

A['c'];
    `,
    `
enum A {
  /** @deprecated */
  b = 'b',
  c = 'c',
}

A.c;
    `,
    `
enum A {
  /** @deprecated */
  b = 'b',
  c = 'c',
}

A['c'];
    `,
    `
function a(value: 'b' | undefined): void;
/** @deprecated */
function a(value: 'c' | undefined): void;
function a(value: string | undefined): void {
  // ...
}

a('b');
    `,
    `
function a(value: 'b' | undefined): void;
/** @deprecated */
function a(value: 'c' | undefined): void;
function a(value: string | undefined): void {
  // ...
}

export default a('b');
    `,
    `
function notDeprecated(): object {
  return {};
}

export default notDeprecated();
    `,
    `
import { deprecatedFunctionWithOverloads } from './deprecated';

const foo = deprecatedFunctionWithOverloads();
    `,
    `
import * as imported from './deprecated';

const foo = imported.deprecatedFunctionWithOverloads();
    `,
    `
import { ClassWithDeprecatedConstructor } from './deprecated';

const foo = new ClassWithDeprecatedConstructor();
    `,
    `
import * as imported from './deprecated';

const foo = new imported.ClassWithDeprecatedConstructor();
    `,
    `
class A {
  a(value: 'b'): void;
  /** @deprecated */
  a(value: 'c'): void;
}
declare const foo: A;
foo.a('b');
    `,
    `
const A = class {
  /** @deprecated */
  constructor();
  constructor(arg: string);
  constructor(arg?: string) {}
};

new A('a');
    `,
    `
type A = {
  (value: 'b'): void;
  /** @deprecated */
  (value: 'c'): void;
};
declare const foo: A;
foo('b');
    `,
    `
declare const a: {
  new (value: 'b'): void;
  /** @deprecated */
  new (value: 'c'): void;
};
new a('b');
    `,
    `
namespace assert {
  export function fail(message?: string | Error): never;
  /** @deprecated since v10.0.0 - use fail([message]) or other assert functions instead. */
  export function fail(actual: unknown, expected: unknown): never;
}

assert.fail('');
    `,
    `
import assert from 'node:assert';

assert.fail('');
    `,
    `
declare module 'deprecations' {
  /** @deprecated */
  export const value = true;
}

import { value } from 'deprecations';
    `,
    `
/** @deprecated Use ts directly. */
export * as ts from 'typescript';
    `,
    `
export {
  /** @deprecated Use ts directly. */
  default as ts,
} from 'typescript';
    `,
    {
      code: `
export { deprecatedFunction as 'bur' } from './deprecated';
      `,
      // Arbitrary module namespace identifier names require TS 5.6
      dependencyConstraints: { typescript: '5.6' },
    },
    {
      code: `
export { 'deprecatedFunction' } from './deprecated';
      `,
      // Arbitrary module namespace identifier names require TS 5.6
      dependencyConstraints: { typescript: '5.6' },
    },
    // https://github.com/typescript-eslint/typescript-eslint/issues/13116
    `
export { circular } from './react';
    `,
    `
namespace A {
  /** @deprecated */
  export type B = string;
  export type C = string;
  export type D = string;
}

export type D = A.C | A.D;
    `,
    `
interface Props {
  anchor: 'foo';
}
declare const x: Props;
const { anchor = '' } = x;
    `,
    `
namespace Foo {}

/**
 * @deprecated
 */
export import Bar = Foo;
    `,
    `
/**
 * @deprecated
 */
export import Bar = require('./deprecated');
    `,

    `
interface Props {
  anchor: 'foo';
}
declare const x: { bar: Props };
const {
  bar: { anchor = '' },
} = x;
    `,
    `
interface Props {
  anchor: 'foo';
}
declare const x: [item: Props];
const [{ anchor = 'bar' }] = x;
    `,
    'function fn(/** @deprecated */ foo = 4) {}',
    {
      code: `
async function fn() {
  const d = await import('./deprecated.js');
  d.default;
}
      `,
      languageOptions: {
        parserOptions: {
          project: './tsconfig.moduleResolution-node16.json',
          projectService: false,
        },
      },
    },
    'call();',

    // this test is to ensure the rule doesn't crash when class implements itself
    // https://github.com/typescript-eslint/typescript-eslint/issues/10031
    `
class Foo implements Foo {
  get bar(): number {
    return 42;
  }

  baz(): number {
    return this.bar;
  }
}
    `,
    `
declare namespace JSX {}

<foo bar={1} />;
    `,
    `
declare namespace JSX {
  interface IntrinsicElements {
    foo: any;
  }
}

<foo bar={1} />;
    `,
    `
declare namespace JSX {
  interface IntrinsicElements {
    foo: unknown;
  }
}

<foo bar={1} />;
    `,
    `
declare namespace JSX {
  interface IntrinsicElements {
    foo: {
      bar: any;
    };
  }
}
<foo bar={1} />;
    `,
    `
declare namespace JSX {
  interface IntrinsicElements {
    foo: {
      bar: unknown;
    };
  }
}
<foo bar={1} />;
    `,
    `
export {
  /** @deprecated */
  foo,
};
    `,
    {
      code: `
/** @deprecated */
function A() {
  return <div />;
}

const a = <A></A>;
      `,
      options: [
        {
          allow: [{ from: 'file', name: 'A' }],
        },
      ],
    },
    {
      code: `
/** @deprecated */
declare class A {}

new A();
      `,
      options: [
        {
          allow: [{ from: 'file', name: 'A' }],
        },
      ],
    },
    {
      code: `
/** @deprecated */
const deprecatedValue = 45;
const bar = deprecatedValue;
      `,
      options: [
        {
          allow: [{ from: 'file', name: 'deprecatedValue' }],
        },
      ],
    },
    {
      code: `
class MyClass {
  /** @deprecated */
  #privateProp = 42;
  value = this.#privateProp;
}
      `,
      options: [
        {
          allow: [{ from: 'file', name: 'privateProp' }],
        },
      ],
    },
    {
      code: `
/** @deprecated */
const deprecatedValue = 45;
const bar = deprecatedValue;
      `,
      options: [
        {
          allow: ['deprecatedValue'],
        },
      ],
    },
    {
      code: `
import { exists } from 'fs';
exists('/foo');
      `,
      options: [
        {
          allow: [
            {
              from: 'package',
              name: 'exists',
              package: 'fs',
            },
          ],
        },
      ],
    },
    {
      code: `
const { exists } = import('fs');
exists('/foo');
      `,
      options: [
        {
          allow: [
            {
              from: 'package',
              name: 'exists',
              package: 'fs',
            },
          ],
        },
      ],
    },
    `
declare const test: string;
const bar = { test };
    `,
    `
const a = {
  /** @deprecated */
  b: 'string',
};

const complex = Symbol() as any;
const c = a[complex];
    `,
    `
const a = {
  b: 'string',
};

const c = a['b'];
    `,
    `
const a = {
  /** @deprecated */
  b: 'string',
};

const key = {};
const c = a[key as any];
    `,
    `
const a = {
  /** @deprecated */
  b: 'string',
};

const key = Symbol();
const c = a[key as any];
    `,
    `
const a = {
  /** @deprecated */
  b: 'string',
};

const key = undefined;
const c = a[key as any];
    `,
    `
const a = {
  /** @deprecated */
  b: 'string',
};

const c = a['nonExistentProperty'];
    `,
    `
const a = {
  /** @deprecated */
  b: 'string',
};

function getKey() {
  return 'c';
}

const c = a[getKey()];
    `,
    `
const a = {
  /** @deprecated */
  b: 'string',
};

const key = {};
const c = a[key];
    `,
    `
const stringObj = new String('b');
const a = {
  /** @deprecated */
  b: 'string',
};
const c = a[stringObj];
    `,
    `
const a = {
  /** @deprecated */
  b: 'string',
};

const key = Symbol('key');
const c = a[key];
    `,
    `
const a = {
  /** @deprecated */
  b: 'string',
};

const key = null;
const c = a[key as any];
    `,
    `
interface Foo {
  /** @deprecated */
  deprecatedProperty: string;
  notDeprecatedProperty: string;
}

declare const a: Foo;
const { notDeprecatedProperty: deprecatedProperty } = a;
    `,
    `
interface A {
  /** @deprecated */
  deprecatedField: string;
}
declare let key: string;
const x: A = { [key]: 'string' };
    `,
    `
interface A {
  /** @deprecated */
  deprecatedField: string;
  notDeprecatedField: string;
}
const x: A = { notDeprecatedField: 'string' };
    `,
    `
interface A {
  /** @deprecated */
  deprecatedField: string;
  notDeprecatedField: string;
}
const x: A = { ['notDeprecatedField']: 'string' };
    `,
    {
      code: `
interface A {
  /** @deprecated */
  deprecatedField: string;
}
const x: A = { deprecatedField: 'string' };
      `,
      options: [
        {
          allow: [{ from: 'file', name: 'A' }],
        },
      ],
    },
    `
interface Foo {
  /** @deprecated */
  deprecatedProperty: string;
  notDeprecatedProperty: string;
}

declare const a: { foo: Foo };
const {
  foo: { notDeprecatedProperty: deprecatedProperty },
} = a;
    `,
  ],
  invalid: [
    {
      code: `
interface AProps {
  /** @deprecated */
  b: number | string;
}

function A(props: AProps) {
  return <div />;
}

const a = <A b="" />;
      `,
      errors: [
        {
          column: 14,
          message: "`b` is deprecated.",
          endColumn: 15,
          endLine: 11,
          line: 11,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
/** @deprecated */ var a = undefined;
a;
      `,
      errors: [
        {
          column: 1,
          message: "`a` is deprecated.",
          endColumn: 2,
          endLine: 3,
          line: 3,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
/** @deprecated */ export var a = undefined;
a;
      `,
      errors: [
        {
          column: 1,
          message: "`a` is deprecated.",
          endColumn: 2,
          endLine: 3,
          line: 3,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
/** @deprecated */ let a = undefined;
a;
      `,
      errors: [
        {
          column: 1,
          message: "`a` is deprecated.",
          endColumn: 2,
          endLine: 3,
          line: 3,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
/** @deprecated */ export let a = undefined;
a;
      `,
      errors: [
        {
          column: 1,
          message: "`a` is deprecated.",
          endColumn: 2,
          endLine: 3,
          line: 3,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
/** @deprecated */ let aLongName = undefined;
aLongName;
      `,
      errors: [
        {
          column: 1,
          message: "`aLongName` is deprecated.",
          endColumn: 10,
          endLine: 3,
          line: 3,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
/** @deprecated */ const a = { b: 1 };
const c = a;
      `,
      errors: [
        {
          column: 11,
          message: "`a` is deprecated.",
          endColumn: 12,
          endLine: 3,
          line: 3,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
/** @deprecated Reason. */ const a = { b: 1 };
const c = a;
      `,
      errors: [
        {
          column: 11,
          message: "`a` is deprecated. Reason.",
          endColumn: 12,
          endLine: 3,
          line: 3,
          messageId: 'deprecatedWithReason',
        },
      ],
    },
    {
      code: `
/** @deprecated */ const a = { b: 1 };
const { c = a } = {};
      `,
      errors: [
        {
          column: 13,
          message: "`a` is deprecated.",
          endColumn: 14,
          endLine: 3,
          line: 3,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
/** @deprecated */ const a = { b: 1 };
const [c = a] = [];
      `,
      errors: [
        {
          column: 12,
          message: "`a` is deprecated.",
          endColumn: 13,
          endLine: 3,
          line: 3,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
/** @deprecated */ const a = { b: 1 };
a;
      `,
      errors: [
        {
          column: 1,
          message: "`a` is deprecated.",
          endColumn: 2,
          endLine: 3,
          line: 3,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
/** @deprecated */ const a = 'foo';
import(\`./path/\${a}.js\`);
      `,
      errors: [
        {
          column: 18,
          message: "`a` is deprecated.",
          endColumn: 19,
          endLine: 3,
          line: 3,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
declare function log(...args: unknown): void;

/** @deprecated */ const a = { b: 1 };

log(a);
      `,
      errors: [
        {
          column: 5,
          message: "`a` is deprecated.",
          endColumn: 6,
          endLine: 6,
          line: 6,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
/** @deprecated */ const a = { b: 1 };
a.b;
      `,
      errors: [
        {
          column: 1,
          message: "`a` is deprecated.",
          endColumn: 2,
          endLine: 3,
          line: 3,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
/** @deprecated */ const a = { b: 1 };
a['b'];
      `,
      errors: [
        {
          column: 1,
          message: "`a` is deprecated.",
          endColumn: 2,
          endLine: 3,
          line: 3,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
/** @deprecated */ const a = { b: 1 };
a?.b;
      `,
      errors: [
        {
          column: 1,
          message: "`a` is deprecated.",
          endColumn: 2,
          endLine: 3,
          line: 3,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
/** @deprecated */ const a = { b: 1 };
a?.['b'];
      `,
      errors: [
        {
          column: 1,
          message: "`a` is deprecated.",
          endColumn: 2,
          endLine: 3,
          line: 3,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
/** @deprecated */ const a = { b: { c: 1 } };
a.b.c;
      `,
      errors: [
        {
          column: 1,
          message: "`a` is deprecated.",
          endColumn: 2,
          endLine: 3,
          line: 3,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
/** @deprecated */ const a = { b: { c: 1 } };
a.b?.c;
      `,
      errors: [
        {
          column: 1,
          message: "`a` is deprecated.",
          endColumn: 2,
          endLine: 3,
          line: 3,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
/** @deprecated */ const a = { b: { c: 1 } };
a?.b?.c;
      `,
      errors: [
        {
          column: 1,
          message: "`a` is deprecated.",
          endColumn: 2,
          endLine: 3,
          line: 3,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
/** @deprecated */ const a = { b: { c: 1 } };
a?.['b']?.['c'];
      `,
      errors: [
        {
          column: 1,
          message: "`a` is deprecated.",
          endColumn: 2,
          endLine: 3,
          line: 3,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
const a = {
  /** @deprecated */ b: { c: 1 },
};
a.b.c;
      `,
      errors: [
        {
          column: 3,
          message: "`b` is deprecated.",
          endColumn: 4,
          endLine: 5,
          line: 5,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
const a = {
  /** @deprecated */ b: 1,
};
const key = 'b';
a[key];
      `,
      errors: [
        {
          column: 3,
          message: "`b` is deprecated.",
          endColumn: 6,
          endLine: 6,
          line: 6,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
const a = {
  /** @deprecated */ b: { c: 1 },
};
a['b']['c'];
      `,
      errors: [
        {
          column: 3,
          message: "`b` is deprecated.",
          endColumn: 6,
          endLine: 5,
          line: 5,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
declare const a: {
  /** @deprecated */ b: { c: 1 };
};
a.b.c;
      `,
      errors: [
        {
          column: 3,
          message: "`b` is deprecated.",
          endColumn: 4,
          endLine: 5,
          line: 5,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
/** @deprecated */ const a = { b: 1 };
const c = a.b;
      `,
      errors: [
        {
          column: 11,
          message: "`a` is deprecated.",
          endColumn: 12,
          endLine: 3,
          line: 3,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
/** @deprecated */ const a = { b: 1 };
const { c } = a.b;
      `,
      errors: [
        {
          column: 15,
          message: "`a` is deprecated.",
          endColumn: 16,
          endLine: 3,
          line: 3,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
/** @deprecated */
declare const test: string;
const myObj = {
  prop: test,
};
      `,
      errors: [
        {
          column: 9,
          message: "`test` is deprecated.",
          endColumn: 13,
          endLine: 5,
          line: 5,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
/** @deprecated */
declare const test: string;
const myObj = {
  deep: {
    prop: test,
  },
};
      `,
      errors: [
        {
          column: 11,
          message: "`test` is deprecated.",
          endColumn: 15,
          endLine: 6,
          line: 6,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
/** @deprecated */
declare const test: string;
const bar = {
  test,
};
      `,
      errors: [
        {
          column: 3,
          message: "`test` is deprecated.",
          endColumn: 7,
          endLine: 5,
          line: 5,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
/** @deprecated */ const a = { b: 1 };
const { c = 'd' } = a.b;
      `,
      errors: [
        {
          column: 21,
          message: "`a` is deprecated.",
          endColumn: 22,
          endLine: 3,
          line: 3,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
/** @deprecated */ const a = { b: 1 };
const { c: d } = a.b;
      `,
      errors: [
        {
          column: 18,
          message: "`a` is deprecated.",
          endColumn: 19,
          endLine: 3,
          line: 3,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
/** @deprecated */
declare const a: string[];
const [b] = [a];
      `,
      errors: [
        {
          column: 14,
          message: "`a` is deprecated.",
          endColumn: 15,
          endLine: 4,
          line: 4,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
/** @deprecated */
class A {}

new A();
      `,
      errors: [
        {
          column: 5,
          message: "`A` is deprecated.",
          endColumn: 6,
          endLine: 5,
          line: 5,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
/** @deprecated */
export class A {}

new A();
      `,
      errors: [
        {
          column: 5,
          message: "`A` is deprecated.",
          endColumn: 6,
          endLine: 5,
          line: 5,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
/** @deprecated */
const A = class {};

new A();
      `,
      errors: [
        {
          column: 5,
          message: "`A` is deprecated.",
          endColumn: 6,
          endLine: 5,
          line: 5,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
/** @deprecated */
declare class A {}

new A();
      `,
      errors: [
        {
          column: 5,
          message: "`A` is deprecated.",
          endColumn: 6,
          endLine: 5,
          line: 5,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
const A = class {
  /** @deprecated */
  constructor() {}
};

new A();
      `,
      errors: [
        {
          column: 5,
          message: "`A` is deprecated.",
          endColumn: 6,
          endLine: 7,
          line: 7,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
const A = class {
  /** @deprecated */
  constructor();
  constructor(arg: string);
  constructor(arg?: string) {}
};

new A();
      `,
      errors: [
        {
          column: 5,
          message: "`A` is deprecated.",
          endColumn: 6,
          endLine: 9,
          line: 9,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
declare const A: {
  /** @deprecated */
  new (): string;
};

new A();
      `,
      errors: [
        {
          column: 5,
          message: "`A` is deprecated.",
          endColumn: 6,
          endLine: 7,
          line: 7,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
/** @deprecated */
declare class A {
  constructor();
}

new A();
      `,
      errors: [
        {
          column: 5,
          message: "`A` is deprecated.",
          endColumn: 6,
          endLine: 7,
          line: 7,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
class A {
  /** @deprecated */
  b: string;
}

declare const a: A;

const { b } = a;
      `,
      errors: [
        {
          column: 9,
          message: "`b` is deprecated.",
          endColumn: 10,
          endLine: 9,
          line: 9,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
declare class A {
  /** @deprecated */
  b(): string;
}

declare const a: A;

a.b;
      `,
      errors: [
        {
          column: 3,
          message: "`b` is deprecated.",
          endColumn: 4,
          endLine: 9,
          line: 9,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
declare class A {
  /** @deprecated */
  b(): string;
}

declare const a: A;
const key = 'b';

a[key];
      `,
      errors: [
        {
          column: 3,
          message: "`b` is deprecated.",
          endColumn: 6,
          endLine: 10,
          line: 10,
          messageId: 'deprecated',
        },
      ],
    },
    {
      // only: true,
      code: `
declare class A {
  /** @deprecated */
  b(): string;
}

declare const a: A;
const key = 'b';

a[key]();
      `,
      errors: [
        {
          column: 3,
          message: "`b` is deprecated.",
          endColumn: 6,
          endLine: 10,
          line: 10,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
declare class A {
  /** @deprecated */
  b(): string;
}

declare const a: A;

a.b();
      `,
      errors: [
        {
          column: 3,
          message: "`b` is deprecated.",
          endColumn: 4,
          endLine: 9,
          line: 9,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
declare class A {
  /** @deprecated */
  b: () => string;
}

declare const a: A;

a.b;
      `,
      errors: [
        {
          column: 3,
          message: "`b` is deprecated.",
          endColumn: 4,
          endLine: 9,
          line: 9,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
declare class A {
  /** @deprecated */
  b: () => string;
}

declare const a: A;

a['b'];
      `,
      errors: [
        {
          column: 3,
          message: "`b` is deprecated.",
          endColumn: 6,
          endLine: 9,
          line: 9,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
declare class A {
  /** @deprecated */
  b(): string;
}

declare const a: A;
const key = 'b';

a[\`\${key}\`];
      `,
      errors: [
        {
          column: 3,
          message: "`b` is deprecated.",
          endColumn: 11,
          endLine: 10,
          line: 10,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
declare class A {
  /** @deprecated */
  computed(): string;
}

declare const a: A;
const k1 = 'comp';
const k2 = 'uted';

a[\`\${k1}\${k2}\`];
      `,
      errors: [
        {
          column: 3,
          message: "`computed` is deprecated.",
          endColumn: 15,
          endLine: 11,
          line: 11,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
declare class A {
  /** @deprecated */
  b(): string;
}

declare const a: A;
const c = \`\${a.b}\`;
      `,
      errors: [
        {
          column: 16,
          message: "`b` is deprecated.",
          endColumn: 17,
          endLine: 8,
          line: 8,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
declare class A {
  /** @deprecated */
  b: () => string;
}

declare const a: A;

a.b();
      `,
      errors: [
        {
          column: 3,
          message: "`b` is deprecated.",
          endColumn: 4,
          endLine: 9,
          line: 9,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
declare class A {
  /** @deprecated */
  b: () => string;
}

declare const a: A;

a['b']();
      `,
      errors: [
        {
          column: 3,
          message: "`b` is deprecated.",
          endColumn: 6,
          endLine: 9,
          line: 9,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
interface A {
  /** @deprecated */
  b: () => string;
}

declare const a: A;

a.b();
      `,
      errors: [
        {
          column: 3,
          message: "`b` is deprecated.",
          endColumn: 4,
          endLine: 9,
          line: 9,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
interface A {
  /** @deprecated */
  b: () => string;
}

declare const a: A;
const key = 'b';

a[key];
      `,
      errors: [
        {
          column: 3,
          message: "`b` is deprecated.",
          endColumn: 6,
          endLine: 10,
          line: 10,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
interface A {
  /** @deprecated */
  b: () => string;
}

declare const a: A;
const key = 'b';

a[key]();
      `,
      errors: [
        {
          column: 3,
          message: "`b` is deprecated.",
          endColumn: 6,
          endLine: 10,
          line: 10,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
interface A {
  /** @deprecated */
  b: () => string;
}

declare const a: A;

a['b']();
      `,
      errors: [
        {
          column: 3,
          message: "`b` is deprecated.",
          endColumn: 6,
          endLine: 9,
          line: 9,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
class A {
  /** @deprecated */
  b(): string {
    return '';
  }
}

declare const a: A;

a.b();
      `,
      errors: [
        {
          column: 3,
          message: "`b` is deprecated.",
          endColumn: 4,
          endLine: 11,
          line: 11,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
declare class A {
  /** @deprecated Use b(value). */
  b(): string;
  b(value: string): string;
}

declare const a: A;

a.b();
      `,
      errors: [
        {
          column: 3,
          message: "`b` is deprecated. Use b(value).",
          endColumn: 4,
          endLine: 10,
          line: 10,
          messageId: 'deprecatedWithReason',
        },
      ],
    },
    {
      code: `
declare class A {
  /** @deprecated */
  static b: string;
}

A.b;
      `,
      errors: [
        {
          column: 3,
          message: "`b` is deprecated.",
          endColumn: 4,
          endLine: 7,
          line: 7,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
declare class A {
  /** @deprecated */
  static b: string;
}

A['b'];
      `,
      errors: [
        {
          column: 3,
          message: "`b` is deprecated.",
          endColumn: 6,
          endLine: 7,
          line: 7,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
declare const a: {
  /** @deprecated */
  b: string;
};

a.b;
      `,
      errors: [
        {
          column: 3,
          message: "`b` is deprecated.",
          endColumn: 4,
          endLine: 7,
          line: 7,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
declare const a: {
  /** @deprecated */
  b: string;
};

a['b'];
      `,
      errors: [
        {
          column: 3,
          message: "`b` is deprecated.",
          endColumn: 6,
          endLine: 7,
          line: 7,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
interface A {
  /** @deprecated */
  b: string;
}

declare const a: A;

a.b;
      `,
      errors: [
        {
          column: 3,
          message: "`b` is deprecated.",
          endColumn: 4,
          endLine: 9,
          line: 9,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
export interface A {
  /** @deprecated */
  b: string;
}

declare const a: A;

a.b;
      `,
      errors: [
        {
          column: 3,
          message: "`b` is deprecated.",
          endColumn: 4,
          endLine: 9,
          line: 9,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
interface A {
  /** @deprecated */
  b: string;
}

declare const a: A;

const { b } = a;
      `,
      errors: [
        {
          column: 9,
          message: "`b` is deprecated.",
          endColumn: 10,
          endLine: 9,
          line: 9,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
type A = {
  /** @deprecated */
  b: string;
};

declare const a: A;

const { b } = a;
      `,
      errors: [
        {
          column: 9,
          message: "`b` is deprecated.",
          endColumn: 10,
          endLine: 9,
          line: 9,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
export type A = {
  /** @deprecated */
  b: string;
};

declare const a: A;

const { b } = a;
      `,
      errors: [
        {
          column: 9,
          message: "`b` is deprecated.",
          endColumn: 10,
          endLine: 9,
          line: 9,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
type A = () => {
  /** @deprecated */
  b: string;
};

declare const a: A;

const { b } = a();
      `,
      errors: [
        {
          column: 9,
          message: "`b` is deprecated.",
          endColumn: 10,
          endLine: 9,
          line: 9,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
/** @deprecated */
type A = string[];

declare const a: A;

const [b] = a;
      `,
      errors: [
        {
          column: 18,
          message: "`A` is deprecated.",
          endColumn: 19,
          endLine: 5,
          line: 5,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
namespace A {
  /** @deprecated */
  export const b = '';
}

A.b;
      `,
      errors: [
        {
          column: 3,
          message: "`b` is deprecated.",
          endColumn: 4,
          endLine: 7,
          line: 7,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
namespace A {
  /** @deprecated */
  export const b = '';
}

A['b'];
      `,
      errors: [
        {
          column: 3,
          message: "`b` is deprecated.",
          endColumn: 6,
          endLine: 7,
          line: 7,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
export namespace A {
  /** @deprecated */
  export const b = '';
}

A.b;
      `,
      errors: [
        {
          column: 3,
          message: "`b` is deprecated.",
          endColumn: 4,
          endLine: 7,
          line: 7,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
namespace A {
  /** @deprecated */
  export function b() {}
}

A.b();
      `,
      errors: [
        {
          column: 3,
          message: "`b` is deprecated.",
          endColumn: 4,
          endLine: 7,
          line: 7,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
namespace A {
  /** @deprecated */
  export function b() {}
}

A['b']();
      `,
      errors: [
        {
          column: 3,
          message: "`b` is deprecated.",
          endColumn: 6,
          endLine: 7,
          line: 7,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
namespace assert {
  export function fail(message?: string | Error): never;
  /** @deprecated since v10.0.0 - use fail([message]) or other assert functions instead. */
  export function fail(actual: unknown, expected: unknown): never;
}

assert.fail({}, {});
      `,
      errors: [
        {
          column: 8,
          message: "`fail` is deprecated. since v10.0.0 - use fail([message]) or other assert functions instead.",
          endColumn: 12,
          endLine: 8,
          line: 8,
          messageId: 'deprecatedWithReason',
        },
      ],
    },
    {
      code: `
import assert from 'node:assert';

assert.fail({}, {});
      `,
      errors: [
        {
          column: 8,
          message: "`fail` is deprecated. since v10.0.0 - use fail([message]) or other assert functions instead.",
          endColumn: 12,
          endLine: 4,
          line: 4,
          messageId: 'deprecatedWithReason',
        },
      ],
    },
    {
      code: `
/** @deprecated */
enum A {
  a,
}

A.a;
      `,
      errors: [
        {
          column: 1,
          message: "`A` is deprecated.",
          endColumn: 2,
          endLine: 7,
          line: 7,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
enum A {
  /** @deprecated */
  a,
}

A.a;
      `,
      errors: [
        {
          column: 3,
          message: "`a` is deprecated.",
          endColumn: 4,
          endLine: 7,
          line: 7,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
enum A {
  /** @deprecated */
  a,
}

const key = 'a';

A[key];
      `,
      errors: [
        {
          column: 3,
          message: "`a` is deprecated.",
          endColumn: 6,
          endLine: 9,
          line: 9,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
enum A {
  /** @deprecated */
  a,
}

A['a'];
      `,
      errors: [
        {
          column: 3,
          message: "`a` is deprecated.",
          endColumn: 6,
          endLine: 7,
          line: 7,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
/** @deprecated */
function a() {}

a();
      `,
      errors: [
        {
          column: 1,
          message: "`a` is deprecated.",
          endColumn: 2,
          endLine: 5,
          line: 5,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
/** @deprecated */
function a(): void;
function a() {}

a();
      `,
      errors: [
        {
          column: 1,
          message: "`a` is deprecated.",
          endColumn: 2,
          endLine: 6,
          line: 6,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
function a(): void;
/** @deprecated */
function a(value: string): void;
function a(value?: string) {}

a('');
      `,
      errors: [
        {
          column: 1,
          message: "`a` is deprecated.",
          endColumn: 2,
          endLine: 7,
          line: 7,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
type A = {
  (value: 'b'): void;
  /** @deprecated */
  (value: 'c'): void;
};
declare const foo: A;
foo('c');
      `,
      errors: [
        {
          column: 1,
          message: "`foo` is deprecated.",
          endColumn: 4,
          endLine: 8,
          line: 8,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
function a(
  /** @deprecated */
  b?: boolean,
) {
  return b;
}
      `,
      errors: [
        {
          column: 10,
          message: "`b` is deprecated.",
          endColumn: 11,
          endLine: 6,
          line: 6,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
export function isTypeFlagSet(
  type: ts.Type,
  flagsToCheck: ts.TypeFlags,
  /** @deprecated This param is not used and will be removed in the future. */
  isReceiver?: boolean,
): boolean {
  const flags = getTypeFlags(type);

  if (isReceiver && flags & ANY_OR_UNKNOWN) {
    return true;
  }

  return (flags & flagsToCheck) !== 0;
}
      `,
      errors: [
        {
          column: 7,
          message: "`isReceiver` is deprecated. This param is not used and will be removed in the future.",
          endColumn: 17,
          endLine: 10,
          line: 10,
          messageId: 'deprecatedWithReason',
        },
      ],
    },
    {
      code: `
/** @deprecated */
declare function a(...args: unknown[]): string;

a\`\`;
      `,
      errors: [
        {
          column: 1,
          message: "`a` is deprecated.",
          endColumn: 2,
          endLine: 5,
          line: 5,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
/** @deprecated */
const A = () => <div />;

const a = <A />;
      `,
      errors: [
        {
          column: 12,
          message: "`A` is deprecated.",
          endColumn: 13,
          endLine: 5,
          line: 5,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
/** @deprecated */
const A = () => <div />;

const a = <A></A>;
      `,
      errors: [
        {
          column: 12,
          message: "`A` is deprecated.",
          endColumn: 13,
          endLine: 5,
          line: 5,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
/** @deprecated */
function A() {
  return <div />;
}

const a = <A />;
      `,
      errors: [
        {
          column: 12,
          message: "`A` is deprecated.",
          endColumn: 13,
          endLine: 7,
          line: 7,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
/** @deprecated */
function A() {
  return <div />;
}

const a = <A></A>;
      `,
      errors: [
        {
          column: 12,
          message: "`A` is deprecated.",
          endColumn: 13,
          endLine: 7,
          line: 7,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
/** @deprecated */
export type A = string;
export type B = string;
export type C = string;

export type D = A | B | C;
      `,
      errors: [
        {
          column: 17,
          message: "`A` is deprecated.",
          endColumn: 18,
          endLine: 7,
          line: 7,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
namespace A {
  /** @deprecated */
  export type B = string;
  export type C = string;
  export type D = string;
}

export type D = A.B | A.C | A.D;
      `,
      errors: [
        {
          column: 19,
          message: "`B` is deprecated.",
          endColumn: 20,
          endLine: 9,
          line: 9,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
interface Props {
  /** @deprecated */
  anchor: 'foo';
}
declare const x: Props;
const { anchor = '' } = x;
      `,
      errors: [
        {
          column: 9,
          message: "`anchor` is deprecated.",
          endColumn: 15,
          endLine: 7,
          line: 7,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
interface Props {
  /** @deprecated */
  anchor: 'foo';
}
declare const x: { bar: Props };
const {
  bar: { anchor = '' },
} = x;
      `,
      errors: [
        {
          column: 10,
          message: "`anchor` is deprecated.",
          endColumn: 16,
          endLine: 8,
          line: 8,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
interface Props {
  /** @deprecated */
  anchor: 'foo';
}
declare const x: [item: Props];
const [{ anchor = 'bar' }] = x;
      `,
      errors: [
        {
          column: 10,
          message: "`anchor` is deprecated.",
          endColumn: 16,
          endLine: 7,
          line: 7,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
interface Props {
  /** @deprecated */
  foo: Props;
}
declare const x: Props;
const { foo = x } = x;
      `,
      errors: [
        {
          column: 9,
          message: "`foo` is deprecated.",
          endColumn: 12,
          endLine: 7,
          line: 7,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
import { DeprecatedClass } from './deprecated';

const foo = new DeprecatedClass();
      `,
      errors: [
        {
          column: 17,
          message: "`DeprecatedClass` is deprecated.",
          endColumn: 32,
          endLine: 4,
          line: 4,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
import { DeprecatedClass } from './deprecated';

declare function inject(something: new () => unknown): void;

inject(DeprecatedClass);
      `,
      errors: [
        {
          column: 8,
          message: "`DeprecatedClass` is deprecated.",
          endColumn: 23,
          endLine: 6,
          line: 6,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
import { deprecatedVariable } from './deprecated';

const foo = deprecatedVariable;
      `,
      errors: [
        {
          column: 13,
          message: "`deprecatedVariable` is deprecated.",
          endColumn: 31,
          endLine: 4,
          line: 4,
          messageId: 'deprecated',
        },
      ],
    },
    // Verify deprecation propagates
    /* eslint-disable @typescript-eslint/internal/no-multiple-lines-of-errors */
    {
      code: `
import { DeprecatedClass } from './deprecated';

declare const x: DeprecatedClass;

const { foo } = x;
      `,
      errors: [
        {
          column: 18,
          message: "`DeprecatedClass` is deprecated.",
          endColumn: 33,
          endLine: 4,
          line: 4,
          messageId: 'deprecated',
        },
        {
          column: 9,
          message: "`foo` is deprecated.",
          endColumn: 12,
          endLine: 6,
          line: 6,
          messageId: 'deprecated',
        },
      ],
    },
    /* eslint-enable @typescript-eslint/internal/no-multiple-lines-of-errors */
    {
      code: `
import { deprecatedFunction } from './deprecated';

deprecatedFunction();
      `,
      errors: [
        {
          column: 1,
          message: "`deprecatedFunction` is deprecated.",
          endColumn: 19,
          endLine: 4,
          line: 4,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
import * as imported from './deprecated';

const foo = new imported.NormalClass();
      `,
      errors: [
        {
          column: 26,
          message: "`NormalClass` is deprecated.",
          endColumn: 37,
          endLine: 4,
          line: 4,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
import { NormalClass } from './deprecated';

const foo = new NormalClass();
      `,
      errors: [
        {
          column: 17,
          message: "`NormalClass` is deprecated.",
          endColumn: 28,
          endLine: 4,
          line: 4,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
import * as imported from './deprecated';

const foo = imported.NormalClass;
      `,
      errors: [
        {
          column: 22,
          message: "`NormalClass` is deprecated.",
          endColumn: 33,
          endLine: 4,
          line: 4,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
import { NormalClass } from './deprecated';

const foo = NormalClass;
      `,
      errors: [
        {
          column: 13,
          message: "`NormalClass` is deprecated.",
          endColumn: 24,
          endLine: 4,
          line: 4,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
import { normalVariable } from './deprecated';

const foo = normalVariable;
      `,
      errors: [
        {
          column: 13,
          message: "`normalVariable` is deprecated.",
          endColumn: 27,
          endLine: 4,
          line: 4,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
import * as imported from './deprecated';

const foo = imported.normalVariable;
      `,
      errors: [
        {
          column: 22,
          message: "`normalVariable` is deprecated.",
          endColumn: 36,
          endLine: 4,
          line: 4,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
import * as imported from './deprecated';

const { normalVariable } = imported;
      `,
      errors: [
        {
          column: 9,
          message: "`normalVariable` is deprecated.",
          endColumn: 23,
          endLine: 4,
          line: 4,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
import { deprecatedVariable } from './deprecated';

const test = {
  someField: deprecatedVariable,
};
      `,
      errors: [
        {
          column: 14,
          message: "`deprecatedVariable` is deprecated.",
          endColumn: 32,
          endLine: 5,
          line: 5,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
import { normalFunction } from './deprecated';

const foo = normalFunction;
      `,
      errors: [
        {
          column: 13,
          message: "`normalFunction` is deprecated.",
          endColumn: 27,
          endLine: 4,
          line: 4,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
import * as imported from './deprecated';

const foo = imported.normalFunction;
      `,
      errors: [
        {
          column: 22,
          message: "`normalFunction` is deprecated.",
          endColumn: 36,
          endLine: 4,
          line: 4,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
import * as imported from './deprecated';

const { normalFunction } = imported;
      `,
      errors: [
        {
          column: 9,
          message: "`normalFunction` is deprecated.",
          endColumn: 23,
          endLine: 4,
          line: 4,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
import { normalFunction } from './deprecated';

const foo = normalFunction();
      `,
      errors: [
        {
          column: 13,
          message: "`normalFunction` is deprecated.",
          endColumn: 27,
          endLine: 4,
          line: 4,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
import * as imported from './deprecated';

const foo = imported.normalFunction();
      `,
      errors: [
        {
          column: 22,
          message: "`normalFunction` is deprecated.",
          endColumn: 36,
          endLine: 4,
          line: 4,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
import { deprecatedFunctionWithOverloads } from './deprecated';

const foo = deprecatedFunctionWithOverloads('a');
      `,
      errors: [
        {
          column: 13,
          message: "`deprecatedFunctionWithOverloads` is deprecated.",
          endColumn: 44,
          endLine: 4,
          line: 4,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
import * as imported from './deprecated';

const foo = imported.deprecatedFunctionWithOverloads('a');
      `,
      errors: [
        {
          column: 22,
          message: "`deprecatedFunctionWithOverloads` is deprecated.",
          endColumn: 53,
          endLine: 4,
          line: 4,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
import { reexportedDeprecatedFunctionWithOverloads } from './deprecated';

const foo = reexportedDeprecatedFunctionWithOverloads;
      `,
      errors: [
        {
          column: 13,
          message: "`reexportedDeprecatedFunctionWithOverloads` is deprecated. Reason",
          endColumn: 54,
          endLine: 4,
          line: 4,
          messageId: 'deprecatedWithReason',
        },
      ],
    },
    {
      code: `
import * as imported from './deprecated';

const foo = imported.reexportedDeprecatedFunctionWithOverloads;
      `,
      errors: [
        {
          column: 22,
          message: "`reexportedDeprecatedFunctionWithOverloads` is deprecated. Reason",
          endColumn: 63,
          endLine: 4,
          line: 4,
          messageId: 'deprecatedWithReason',
        },
      ],
    },
    {
      code: `
import * as imported from './deprecated';

const { reexportedDeprecatedFunctionWithOverloads } = imported;
      `,
      errors: [
        {
          column: 9,
          message: "`reexportedDeprecatedFunctionWithOverloads` is deprecated. Reason",
          endColumn: 50,
          endLine: 4,
          line: 4,
          messageId: 'deprecatedWithReason',
        },
      ],
    },
    {
      code: `
import { reexportedDeprecatedFunctionWithOverloads } from './deprecated';

const foo = reexportedDeprecatedFunctionWithOverloads();
      `,
      errors: [
        {
          column: 13,
          message: "`reexportedDeprecatedFunctionWithOverloads` is deprecated. Reason",
          endColumn: 54,
          endLine: 4,
          line: 4,
          messageId: 'deprecatedWithReason',
        },
      ],
    },
    {
      code: `
import * as imported from './deprecated';

const foo = imported.reexportedDeprecatedFunctionWithOverloads();
      `,
      errors: [
        {
          column: 22,
          message: "`reexportedDeprecatedFunctionWithOverloads` is deprecated. Reason",
          endColumn: 63,
          endLine: 4,
          line: 4,
          messageId: 'deprecatedWithReason',
        },
      ],
    },
    {
      code: `
import { reexportedDeprecatedFunctionWithOverloads } from './deprecated';

const foo = reexportedDeprecatedFunctionWithOverloads('a');
      `,
      errors: [
        {
          column: 13,
          message: "`reexportedDeprecatedFunctionWithOverloads` is deprecated. Reason",
          endColumn: 54,
          endLine: 4,
          line: 4,
          messageId: 'deprecatedWithReason',
        },
      ],
    },
    {
      code: `
import * as imported from './deprecated';

const foo = imported.reexportedDeprecatedFunctionWithOverloads('a');
      `,
      errors: [
        {
          column: 22,
          message: "`reexportedDeprecatedFunctionWithOverloads` is deprecated. Reason",
          endColumn: 63,
          endLine: 4,
          line: 4,
          messageId: 'deprecatedWithReason',
        },
      ],
    },
    {
      code: `
import { ClassWithDeprecatedConstructor } from './deprecated';

const foo = new ClassWithDeprecatedConstructor('a');
      `,
      errors: [
        {
          column: 17,
          message: "`ClassWithDeprecatedConstructor` is deprecated.",
          endColumn: 47,
          endLine: 4,
          line: 4,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
import * as imported from './deprecated';

const foo = new imported.ClassWithDeprecatedConstructor('a');
      `,
      errors: [
        {
          column: 26,
          message: "`ClassWithDeprecatedConstructor` is deprecated.",
          endColumn: 56,
          endLine: 4,
          line: 4,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
import { ReexportedClassWithDeprecatedConstructor } from './deprecated';

const foo = ReexportedClassWithDeprecatedConstructor;
      `,
      errors: [
        {
          column: 13,
          message: "`ReexportedClassWithDeprecatedConstructor` is deprecated. Reason",
          endColumn: 53,
          endLine: 4,
          line: 4,
          messageId: 'deprecatedWithReason',
        },
      ],
    },
    {
      code: `
import * as imported from './deprecated';

const foo = imported.ReexportedClassWithDeprecatedConstructor;
      `,
      errors: [
        {
          column: 22,
          message: "`ReexportedClassWithDeprecatedConstructor` is deprecated. Reason",
          endColumn: 62,
          endLine: 4,
          line: 4,
          messageId: 'deprecatedWithReason',
        },
      ],
    },
    {
      code: `
import * as imported from './deprecated';

const { ReexportedClassWithDeprecatedConstructor } = imported;
      `,
      errors: [
        {
          column: 9,
          message: "`ReexportedClassWithDeprecatedConstructor` is deprecated. Reason",
          endColumn: 49,
          endLine: 4,
          line: 4,
          messageId: 'deprecatedWithReason',
        },
      ],
    },
    {
      code: `
import { ReexportedClassWithDeprecatedConstructor } from './deprecated';

const foo = ReexportedClassWithDeprecatedConstructor();
      `,
      errors: [
        {
          column: 13,
          message: "`ReexportedClassWithDeprecatedConstructor` is deprecated. Reason",
          endColumn: 53,
          endLine: 4,
          line: 4,
          messageId: 'deprecatedWithReason',
        },
      ],
    },
    {
      code: `
import * as imported from './deprecated';

const foo = imported.ReexportedClassWithDeprecatedConstructor();
      `,
      errors: [
        {
          column: 22,
          message: "`ReexportedClassWithDeprecatedConstructor` is deprecated. Reason",
          endColumn: 62,
          endLine: 4,
          line: 4,
          messageId: 'deprecatedWithReason',
        },
      ],
    },
    {
      code: `
import { ReexportedClassWithDeprecatedConstructor } from './deprecated';

const foo = ReexportedClassWithDeprecatedConstructor('a');
      `,
      errors: [
        {
          column: 13,
          message: "`ReexportedClassWithDeprecatedConstructor` is deprecated. Reason",
          endColumn: 53,
          endLine: 4,
          line: 4,
          messageId: 'deprecatedWithReason',
        },
      ],
    },
    {
      code: `
import * as imported from './deprecated';

const foo = imported.ReexportedClassWithDeprecatedConstructor('a');
      `,
      errors: [
        {
          column: 22,
          message: "`ReexportedClassWithDeprecatedConstructor` is deprecated. Reason",
          endColumn: 62,
          endLine: 4,
          line: 4,
          messageId: 'deprecatedWithReason',
        },
      ],
    },
    {
      code: `
import imported from './deprecated';

imported;
      `,
      errors: [
        {
          column: 1,
          message: "`imported` is deprecated.",
          endColumn: 9,
          endLine: 4,
          line: 4,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
async function fn() {
  const d = await import('./deprecated.js');
  d.default.default;
}
      `,
      errors: [
        {
          column: 13,
          message: "`default` is deprecated.",
          endColumn: 20,
          endLine: 4,
          line: 4,
          messageId: 'deprecated',
        },
      ],
      languageOptions: {
        parserOptions: {
          project: './tsconfig.moduleResolution-node16.json',
          projectService: false,
        },
      },
    },
    {
      code: `
/** @deprecated */
interface Foo {}

class Bar implements Foo {}
      `,
      errors: [
        {
          column: 22,
          message: "`Foo` is deprecated.",
          endColumn: 25,
          endLine: 5,
          line: 5,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
/** @deprecated */
interface Foo {}

export class Bar implements Foo {}
      `,
      errors: [
        {
          column: 29,
          message: "`Foo` is deprecated.",
          endColumn: 32,
          endLine: 5,
          line: 5,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
/** @deprecated */
interface Foo {}

interface Baz {}

export class Bar implements Baz, Foo {}
      `,
      errors: [
        {
          column: 34,
          message: "`Foo` is deprecated.",
          endColumn: 37,
          endLine: 7,
          line: 7,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
/** @deprecated */
class Foo {}

export class Bar extends Foo {}
      `,
      errors: [
        {
          column: 26,
          message: "`Foo` is deprecated.",
          endColumn: 29,
          endLine: 5,
          line: 5,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
/** @deprecated */
declare function decorator(constructor: Function);

@decorator
export class Foo {}
      `,
      errors: [
        {
          column: 2,
          message: "`decorator` is deprecated.",
          endColumn: 11,
          endLine: 5,
          line: 5,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
/** @deprecated */
function a(): object {
  return {};
}

export default a();
      `,
      errors: [
        {
          column: 16,
          message: "`a` is deprecated.",
          endColumn: 17,
          endLine: 7,
          line: 7,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
class A {
  /** @deprecated */
  constructor() {}
}

class B extends A {
  constructor() {
    /** should report but does not */
    super();
  }
}
      `,
      errors: [
        {
          column: 5,
          message: "`super` is deprecated.",
          endColumn: 10,
          endLine: 10,
          line: 10,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
class A {
  /** @deprecated test reason*/
  constructor() {}
}

class B extends A {
  constructor() {
    /** should report but does not */
    super();
  }
}
      `,
      errors: [
        {
          column: 5,
          message: "`super` is deprecated. test reason",
          endColumn: 10,
          endLine: 10,
          line: 10,
          messageId: 'deprecatedWithReason',
        },
      ],
    },
    {
      code: 'const a = <div aria-grabbed></div>;',
      errors: [
        {
          column: 16,
          message: "`aria-grabbed` is deprecated. in ARIA 1.1",
          endColumn: 28,
          endLine: 1,
          line: 1,
          messageId: 'deprecatedWithReason',
        },
      ],
    },
    {
      code: `
declare namespace JSX {
  interface IntrinsicElements {
    'foo-bar:baz-bam': {
      name: string;
      /**
       * @deprecated
       */
      deprecatedProp: string;
    };
  }
}

const componentDashed = <foo-bar:baz-bam name="e" deprecatedProp="oh no" />;
      `,
      errors: [
        {
          column: 51,
          message: "`deprecatedProp` is deprecated.",
          endColumn: 65,
          endLine: 14,
          line: 14,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
import * as React from 'react';

interface Props {
  /**
   * @deprecated
   */
  deprecatedProp: string;
}

interface Tab {
  List: React.FC<Props>;
}

const Tab: Tab = {
  List: () => <div>Hi</div>,
};

const anotherExample = <Tab.List deprecatedProp="oh no" />;
      `,
      errors: [
        {
          column: 34,
          message: "`deprecatedProp` is deprecated.",
          endColumn: 48,
          endLine: 19,
          line: 19,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
import { exists } from 'fs';
exists('/foo');
      `,
      errors: [
        {
          column: 1,
          message: "`exists` is deprecated. Since v1.0.0 - Use {@link stat} or {@link access} instead.",
          endColumn: 7,
          endLine: 3,
          line: 3,
          messageId: 'deprecatedWithReason',
        },
      ],
      options: [
        {
          allow: [
            {
              from: 'package',
              name: 'exists',
              package: 'hoge',
            },
          ],
        },
      ],
    },
    {
      code: `
declare class A {
  /** @deprecated */
  accessor b: () => string;
}

declare const a: A;

a.b;
      `,
      // `accessor` keyword requires TS 4.9
      dependencyConstraints: { typescript: '4.9' },
      errors: [
        {
          column: 3,
          message: "`b` is deprecated.",
          endColumn: 4,
          endLine: 9,
          line: 9,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
declare class A {
  /** @deprecated */
  accessor b: () => string;
}

declare const a: A;

a['b'];
      `,
      // `accessor` keyword requires TS 4.9
      dependencyConstraints: { typescript: '4.9' },
      errors: [
        {
          column: 3,
          message: "`b` is deprecated.",
          endColumn: 6,
          endLine: 9,
          line: 9,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
declare class A {
  /** @deprecated */
  accessor b: () => string;
}

declare const a: A;

a.b();
      `,
      // `accessor` keyword requires TS 4.9
      dependencyConstraints: { typescript: '4.9' },
      errors: [
        {
          column: 3,
          message: "`b` is deprecated.",
          endColumn: 4,
          endLine: 9,
          line: 9,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
declare class A {
  /** @deprecated */
  accessor b: () => string;
}

declare const a: A;

a['b']();
      `,
      // `accessor` keyword requires TS 4.9
      dependencyConstraints: { typescript: '4.9' },
      errors: [
        {
          column: 3,
          message: "`b` is deprecated.",
          endColumn: 6,
          endLine: 9,
          line: 9,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
class A {
  /** @deprecated */
  accessor b = (): string => {
    return '';
  };
}

declare const a: A;

a.b();
      `,
      // `accessor` keyword requires TS 4.9
      dependencyConstraints: { typescript: '4.9' },
      errors: [
        {
          column: 3,
          message: "`b` is deprecated.",
          endColumn: 4,
          endLine: 11,
          line: 11,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
declare class A {
  /** @deprecated */
  static accessor b: () => string;
}

A.b();
      `,
      // `accessor` keyword requires TS 4.9
      dependencyConstraints: { typescript: '4.9' },
      errors: [
        {
          column: 3,
          message: "`b` is deprecated.",
          endColumn: 4,
          endLine: 7,
          line: 7,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
class A {
  /** @deprecated */
  #b = () => {};

  c() {
    this.#b();
  }
}
      `,
      errors: [
        {
          column: 10,
          message: "`#b` is deprecated.",
          endColumn: 12,
          endLine: 7,
          line: 7,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
interface A {
  /** @deprecated */
  deprecatedField: string;
}
const x: A = { deprecatedField: 'string' };
      `,
      errors: [
        {
          column: 16,
          message: "`deprecatedField` is deprecated.",
          endColumn: 31,
          endLine: 6,
          line: 6,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
interface A {
  /** @deprecated use newField instead */
  deprecatedField: string;
}
const x: A = { deprecatedField: 'string' };
      `,
      errors: [
        {
          column: 16,
          message: "`deprecatedField` is deprecated. use newField instead",
          endColumn: 31,
          endLine: 6,
          line: 6,
          messageId: 'deprecatedWithReason',
        },
      ],
    },
    {
      code: `
interface A {
  /** @deprecated */
  deprecatedField: string;
}
let x: A;
x = { deprecatedField: 'string' };
      `,
      errors: [
        {
          column: 7,
          message: "`deprecatedField` is deprecated.",
          endColumn: 22,
          endLine: 7,
          line: 7,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
interface A {
  /** @deprecated */
  deprecatedField: string;
}
interface B {
  nested: A;
}
const y: B = { nested: { deprecatedField: 'string' } };
      `,
      errors: [
        {
          column: 26,
          message: "`deprecatedField` is deprecated.",
          endColumn: 41,
          endLine: 9,
          line: 9,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
interface A {
  /** @deprecated */
  deprecatedField: string;
}
declare function func(x: A): void;
func({ deprecatedField: 'string' });
      `,
      errors: [
        {
          column: 8,
          message: "`deprecatedField` is deprecated.",
          endColumn: 23,
          endLine: 7,
          line: 7,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
interface A {
  /** @deprecated */
  deprecatedField: string;
}
function func(): A {
  return { deprecatedField: 'string' };
}
      `,
      errors: [
        {
          column: 12,
          message: "`deprecatedField` is deprecated.",
          endColumn: 27,
          endLine: 7,
          line: 7,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
interface A {
  /** @deprecated */
  deprecatedField: string;
}
const x: A[] = [{ deprecatedField: 'string' }];
      `,
      errors: [
        {
          column: 19,
          message: "`deprecatedField` is deprecated.",
          endColumn: 34,
          endLine: 6,
          line: 6,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
interface A {
  /** @deprecated */
  deprecatedField: string;
}
const x = { deprecatedField: 'string' } as A;
      `,
      errors: [
        {
          column: 13,
          message: "`deprecatedField` is deprecated.",
          endColumn: 28,
          endLine: 6,
          line: 6,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
interface A {
  /** @deprecated */
  deprecatedField: string;
}
const x = { deprecatedField: 'string' } satisfies A;
      `,
      // `satisfies` requires TS 4.9
      dependencyConstraints: { typescript: '4.9' },
      errors: [
        {
          column: 13,
          message: "`deprecatedField` is deprecated.",
          endColumn: 28,
          endLine: 6,
          line: 6,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
class C {
  /** @deprecated */
  deprecatedField = 'x';
}
const x: C = { deprecatedField: 'string' };
      `,
      errors: [
        {
          column: 16,
          message: "`deprecatedField` is deprecated.",
          endColumn: 31,
          endLine: 6,
          line: 6,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
interface A {
  /** @deprecated */
  deprecatedField: string;
}
interface B extends A {}
const b: B = { deprecatedField: 'x' };
      `,
      errors: [
        {
          column: 16,
          message: "`deprecatedField` is deprecated.",
          endColumn: 31,
          endLine: 7,
          line: 7,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
interface A {
  /** @deprecated */
  deprecatedField: string;
}
const deprecatedField = 'string';
const x: A = { deprecatedField };
      `,
      errors: [
        {
          column: 16,
          message: "`deprecatedField` is deprecated.",
          endColumn: 31,
          endLine: 7,
          line: 7,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
interface A {
  /** @deprecated */
  deprecatedMethod(): void;
}
const x: A = { deprecatedMethod() {} };
      `,
      errors: [
        {
          column: 16,
          message: "`deprecatedMethod` is deprecated.",
          endColumn: 32,
          endLine: 6,
          line: 6,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: noFormat`
interface A {
  /** @deprecated */
  deprecatedField: string;
  notDeprecatedField: string;
}
const x: A = { 'deprecatedField': 'string' };
      `,
      errors: [
        {
          column: 16,
          message: "`deprecatedField` is deprecated.",
          endColumn: 33,
          endLine: 7,
          line: 7,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
interface A {
  /** @deprecated */
  deprecatedField: string;
  notDeprecatedField: string;
}
const x: A = { ['deprecatedField']: 'string' };
      `,
      errors: [
        {
          column: 17,
          message: "`deprecatedField` is deprecated.",
          endColumn: 34,
          endLine: 7,
          line: 7,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
interface A {
  /** @deprecated */
  deprecatedField: string;
  notDeprecatedField: string;
}
declare const key: 'deprecatedField';
const x: A = { [key]: 'string' };
      `,
      errors: [
        {
          column: 17,
          message: "`deprecatedField` is deprecated.",
          endColumn: 20,
          endLine: 8,
          line: 8,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
interface A {
  /** @deprecated */
  deprecatedField: string;
  notDeprecatedField: string;
}
const x: A = { ['deprecatedField' as const]: 'string' };
      `,
      errors: [
        {
          column: 17,
          message: "`deprecatedField` is deprecated.",
          endColumn: 43,
          endLine: 7,
          line: 7,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
interface A {
  /** @deprecated */
  deprecatedField: string;
  notDeprecatedField: string;
}
const x: A = { ['deprecatedField' satisfies string]: 'string' };
      `,
      // `satisfies` requires TS 4.9
      dependencyConstraints: { typescript: '4.9' },
      errors: [
        {
          column: 17,
          message: "`deprecatedField` is deprecated.",
          endColumn: 51,
          endLine: 7,
          line: 7,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
interface A {
  /** @deprecated */
  deprecatedField: string;
  notDeprecatedField: string;
}
const x: A = { ['notDeprecatedField' as 'deprecatedField']: 'string' };
      `,
      errors: [
        {
          column: 17,
          message: "`deprecatedField` is deprecated.",
          endColumn: 58,
          endLine: 7,
          line: 7,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
interface A {
  /** @deprecated */
  deprecatedField: string;
  notDeprecatedField: string;
}
const x: A = { [\`\${'deprecated'}Field\`]: 'string' };
      `,
      // `satisfies` requires TS 4.9
      dependencyConstraints: { typescript: '5.2' },
      errors: [
        {
          column: 17,
          message: "`deprecatedField` is deprecated.",
          endColumn: 39,
          endLine: 7,
          line: 7,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
interface Foo {
  /** @deprecated */
  1: string;
}
const a: Foo = { 1: '' };
      `,
      errors: [
        {
          column: 18,
          message: "`1` is deprecated.",
          endColumn: 19,
          endLine: 6,
          line: 6,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
interface Foo {
  /** @deprecated */
  1: string;
}
const a: Foo = { [1]: '' };
      `,
      errors: [
        {
          column: 19,
          message: "`1` is deprecated.",
          endColumn: 20,
          endLine: 6,
          line: 6,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
const a = {
  /** @deprecated */
  b: 'string',
};

const c = a['b'];
      `,
      errors: [
        {
          column: 13,
          message: "`b` is deprecated.",
          endColumn: 16,
          endLine: 7,
          line: 7,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
const a = {
  /** @deprecated */
  b: 'string',
};
const x = 'b';
const c = a[x];
      `,
      errors: [
        {
          column: 13,
          message: "`b` is deprecated.",
          endColumn: 14,
          endLine: 7,
          line: 7,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
const a = {
  /** @deprecated */
  [2]: 'string',
};
const x = 'b';
const c = a[2];
      `,
      errors: [
        {
          column: 13,
          message: "`2` is deprecated.",
          endColumn: 14,
          endLine: 7,
          line: 7,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
const a = {
  /** @deprecated reason for deprecation */
  b: 'string',
};

const key = 'b';
const stringKey = key as const;
const c = a[stringKey];
      `,
      errors: [
        {
          column: 13,
          message: "`b` is deprecated. reason for deprecation",
          endColumn: 22,
          endLine: 9,
          line: 9,
          messageId: 'deprecatedWithReason',
        },
      ],
    },
    {
      code: `
enum Keys {
  B = 'b',
}

const a = {
  /** @deprecated reason for deprecation */
  b: 'string',
};

const key = Keys.B;
const c = a[key];
      `,
      errors: [
        {
          column: 13,
          message: "`b` is deprecated. reason for deprecation",
          endColumn: 16,
          endLine: 12,
          line: 12,
          messageId: 'deprecatedWithReason',
        },
      ],
    },
    {
      code: `
declare const Keys: {
  a: 1;
};

const a = {
  /** @deprecated reason for deprecation */
  [1]: 'string',
};

const key = Keys.a;
const c = a[key];
      `,
      errors: [
        {
          column: 13,
          message: "`1` is deprecated. reason for deprecation",
          endColumn: 16,
          endLine: 12,
          line: 12,
          messageId: 'deprecatedWithReason',
        },
      ],
    },
    {
      code: `
const a = {
  /** @deprecated */
  b: 'string',
};

const key = \`b\`;
const c = a[key];
      `,
      errors: [
        {
          column: 13,
          message: "`b` is deprecated.",
          endColumn: 16,
          endLine: 8,
          line: 8,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
const stringObj = 'b';
const a = {
  /** @deprecated */
  b: 'string',
};
const c = a[stringObj];
      `,
      errors: [
        {
          column: 13,
          message: "`b` is deprecated.",
          endColumn: 22,
          endLine: 7,
          line: 7,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
declare function x(): 'b';

const a = {
  /** @deprecated */
  b: 'string',
};

const c = a[x()];
      `,
      errors: [
        {
          column: 13,
          message: "`b` is deprecated.",
          endColumn: 16,
          endLine: 9,
          line: 9,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
declare const x: { y: 'b' };

const a = {
  /** @deprecated */
  b: 'string',
};

const c = a[x.y];
      `,
      errors: [
        {
          column: 13,
          message: "`b` is deprecated.",
          endColumn: 16,
          endLine: 9,
          line: 9,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
import { deprecatedFunction } from './deprecated';

export { deprecatedFunction };
      `,
      errors: [
        {
          column: 10,
          message: "`deprecatedFunction` is deprecated.",
          endColumn: 28,
          endLine: 4,
          line: 4,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
export { deprecatedFunction } from './deprecated';
      `,
      errors: [
        {
          column: 10,
          message: "`deprecatedFunction` is deprecated.",
          endColumn: 28,
          endLine: 2,
          line: 2,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
export type { T, U } from './deprecated';
      `,
      errors: [
        {
          column: 15,
          message: "`T` is deprecated. Reason",
          endColumn: 16,
          endLine: 2,
          line: 2,
          messageId: 'deprecatedWithReason',
        },
      ],
    },
    {
      code: `
export { default as foo } from './deprecated';
      `,
      errors: [
        {
          column: 21,
          message: "`foo` is deprecated.",
          endColumn: 24,
          endLine: 2,
          line: 2,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
export { deprecatedFunction as bar } from './deprecated';
      `,
      errors: [
        {
          column: 32,
          message: "`bar` is deprecated.",
          endColumn: 35,
          endLine: 2,
          line: 2,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
import { deprecatedVariable } from './deprecated';

void { deprecatedVariable };
      `,
      errors: [
        {
          column: 8,
          message: "`deprecatedVariable` is deprecated.",
          endColumn: 26,
          endLine: 4,
          line: 4,
          messageId: 'deprecated',
        },
      ],
    },
    {
      code: `
import { normalVariable } from './deprecated';

void { normalVariable };
      `,
      errors: [
        {
          column: 8,
          message: "`normalVariable` is deprecated.",
          endColumn: 22,
          endLine: 4,
          line: 4,
          messageId: 'deprecated',
        },
      ],
    },
  ],
});
