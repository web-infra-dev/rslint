import { RuleTester } from '@typescript-eslint/rule-tester';

import { getFixturesRootDir } from '../RuleTester';

const rootDir = getFixturesRootDir();
const ruleTester = new RuleTester({
  languageOptions: {
    parserOptions: {
      project: './tsconfig.json',
      projectService: false,
      tsconfigRootDir: rootDir,
    },
  },
});

ruleTester.run('no-useless-default-assignment', {
  valid: [
    `
function Bar({ foo = '' }: { foo?: string }) {
  return foo;
}
    `,
    `
const { foo } = { foo: 'bar' };
    `,
    `
[1, 2, 3, undefined].map((a = 42) => a + 1);
    `,
    `
function test(a?: number) {
  return a;
}
    `,
    `
const obj: { a?: string } = {};
const { a = 'default' } = obj;
    `,
    `
function test(a: string | undefined = 'default') {
  return a;
}
    `,
    `
(a: string = 'default') => a;
    `,
    `
function test(a: string = 'default') {
  return a;
}
    `,
    `
class C {
  public test(a: string = 'default') {
    return a;
  }
}
    `,
    `
const obj: { a: string | undefined } = { a: undefined };
const { a = 'default' } = obj;
    `,
    `
function test(arr: number[] | undefined = []) {
  return arr;
}
    `,
    `
function Bar({ nested: { foo = '' } = {} }: { nested?: { foo?: string } }) {
  return foo;
}
    `,
    `
function test(a: any = 'default') {
  return a;
}
    `,
    `
function test(a: unknown = 'default') {
  return a;
}
    `,
    `
function test(a = 5) {
  return a;
}
    `,
    `
function createValidator(): () => void {
  return (param = 5) => {};
}
    `,
    `
function getValue(): undefined;
function getValue(box: { value: string }): string;
function getValue({ value = '' }: { value?: string } = {}): string | undefined {
  return value;
}
    `,
    `
function getValueObject({ value = '' }: Partial<{ value: string }>) {
  return value;
}
    `,
    `
const { value = 'default' } = someUnknownFunction();
    `,
    `
const [value = 'default'] = someUnknownFunction();
    `,
    `
for (const { value = 'default' } of []) {
}
    `,
    `
for (const [value = 'default'] of []) {
}
    `,
    `
declare const x: [[number | undefined]];
const [[a = 1]] = x;
    `,
    `
function foo(x: string = '') {}
    `,
    `
const foo = (x: string = '') => {};
    `,
    `
declare const g: Array<string>;
const [foo = ''] = g;
    `,
    `
declare const g: Record<string, string>;
const { foo = '' } = g;
    `,
    `
declare const h: { [key: string]: string };
const { bar = '' } = h;
    `,
    // https://github.com/typescript-eslint/typescript-eslint/issues/11849
    `
type Merge = boolean | ((incoming: string[]) => void);

const policy: { merge: Merge } = {
  merge: (incoming: string[] = []) => {
    incoming;
  },
};
    `,
    // https://github.com/typescript-eslint/typescript-eslint/issues/11846
    `
const [a, b = ''] = 'somestr'.split('.');
    `,
    `
declare const params: string[];
const [c = '123'] = params;
    `,
    `
declare function useCallback<T>(callback: T);
useCallback((value: number[] = []) => {});
    `,
    `
declare const tuple: [string];
const [a, b = 'default'] = tuple;
    `,
    // https://github.com/typescript-eslint/typescript-eslint/issues/11850
    `
const { a = 'default' } = Math.random() > 0.5 ? { a: 'Hello' } : {};
    `,
    // https://github.com/typescript-eslint/typescript-eslint/issues/11980
    `
const { a = 'baz' } = cond ? {} : { a: 'bar' };
    `,
    `
const { a = 'baz' } = foo && { a: 'bar' };
    `,
    `
const { a = 'baz' } = cond ? { a: 'foo', ...extra } : { a: 'bar' };
    `,
    `
const { a = 'baz' } = cond ? { ...foo } : { a: 'bar' };
    `,
    `
function f(this: void, { bar = 42 }: { bar?: number }) {
  return bar;
}
    `,
  ],
  invalid: [
    {
      code: `
function Bar({ foo = '' }: { foo: string }) {
  return foo;
}
      `,
      errors: [
        {
          column: 22,
          endColumn: 24,
          line: 2,
          messageId: 'uselessDefaultAssignment',
        },
      ],
      output: `
function Bar({ foo }: { foo: string }) {
  return foo;
}
      `,
    },
    {
      code: `
class C {
  public method({ foo = '' }: { foo: string }) {
    return foo;
  }
}
      `,
      errors: [
        {
          column: 25,
          endColumn: 27,
          line: 3,
          messageId: 'uselessDefaultAssignment',
        },
      ],
      output: `
class C {
  public method({ foo }: { foo: string }) {
    return foo;
  }
}
      `,
    },
    {
      code: `
const { 'literal-key': literalKey = 'default' } = { 'literal-key': 'value' };
      `,
      errors: [
        {
          column: 37,
          endColumn: 46,
          line: 2,
          messageId: 'uselessDefaultAssignment',
        },
      ],
      output: `
const { 'literal-key': literalKey } = { 'literal-key': 'value' };
      `,
    },
    {
      code: `
[1, 2, 3].map((a = 42) => a + 1);
      `,
      errors: [
        {
          column: 20,
          endColumn: 22,
          line: 2,
          messageId: 'uselessDefaultAssignment',
        },
      ],
      output: `
[1, 2, 3].map((a) => a + 1);
      `,
    },
    {
      code: `
function getValue([value = '']: [string]) {
  return value;
}
      `,
      errors: [
        {
          column: 28,
          endColumn: 30,
          line: 2,
          messageId: 'uselessDefaultAssignment',
        },
      ],
      output: `
function getValue([value]: [string]) {
  return value;
}
      `,
    },
    {
      code: `
interface B {
  foo: (b: boolean | string) => void;
}

const h: B = {
  foo: (b = false) => {},
};
      `,
      errors: [
        {
          column: 13,
          endColumn: 18,
          line: 7,
          messageId: 'uselessDefaultAssignment',
        },
      ],
      output: `
interface B {
  foo: (b: boolean | string) => void;
}

const h: B = {
  foo: (b) => {},
};
      `,
    },
    {
      code: `
function foo(a = undefined) {}
      `,
      errors: [
        {
          column: 18,
          endColumn: 27,
          line: 2,
          messageId: 'uselessUndefined',
        },
      ],
      output: `
function foo(a) {}
      `,
    },
    {
      code: `
const { a = undefined } = {};
      `,
      errors: [
        {
          column: 13,
          endColumn: 22,
          line: 2,
          messageId: 'uselessUndefined',
        },
      ],
      output: `
const { a } = {};
      `,
    },
    {
      code: `
const [a = undefined] = [];
      `,
      errors: [
        {
          column: 12,
          endColumn: 21,
          line: 2,
          messageId: 'uselessUndefined',
        },
      ],
      output: `
const [a] = [];
      `,
    },
    {
      code: `
function foo({ a = undefined }) {}
      `,
      errors: [
        {
          column: 20,
          endColumn: 29,
          line: 2,
          messageId: 'uselessUndefined',
        },
      ],
      output: `
function foo({ a }) {}
      `,
    },
    // https://github.com/typescript-eslint/typescript-eslint/issues/11847
    {
      code: `
function myFunction(p1: string, p2: number | undefined = undefined) {
  console.log(p1, p2);
}
      `,
      errors: [
        {
          column: 58,
          endColumn: 67,
          line: 2,
          messageId: 'preferOptionalSyntax',
        },
      ],
      output: `
function myFunction(p1: string, p2?: number | undefined) {
  console.log(p1, p2);
}
      `,
    },
  ],
});
