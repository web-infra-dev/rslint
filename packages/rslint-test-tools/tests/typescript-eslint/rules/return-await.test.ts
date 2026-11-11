// Every case from typescript-eslint v8.67.0, with messages and full ranges
// additionally pinned from the same upstream version. No upstream expectations are relaxed.
// https://github.com/typescript-eslint/typescript-eslint/blob/v8.67.0/packages/eslint-plugin/tests/rules/return-await.test.ts
import { RuleTester } from '@rslint/rule-tester';

const ruleTester = new RuleTester({
  plugin: '@typescript-eslint',
  fixtureFiles: {
    'tsconfig.json':
      '{"compilerOptions":{"jsx":"preserve","target":"es2015","module":"commonjs","strict":true,"types":[],"lib":["es2015","es2017","esnext"],"experimentalDecorators":true},"include":["*.ts","*.tsx"]}',
  },
  languageOptions: {
    parserOptions: {
      project: './tsconfig.json',
    },
  },
});

ruleTester.run('return-await', {
  valid: [
    {
      code: 'return;',
      filename: 'file.ts',
    },
    {
      code: `
function test() {
  return;
}
    `,
      filename: 'file.ts',
    },
    {
      code: `
function test() {
  return 1;
}
    `,
      filename: 'file.ts',
    },
    {
      code: `
async function test() {
  return;
}
    `,
      filename: 'file.ts',
    },
    {
      code: `
async function test() {
  return 1;
}
    `,
      filename: 'file.ts',
    },
    {
      code: 'const test = () => 1;',
      filename: 'file.ts',
    },
    {
      code: 'const test = async () => 1;',
      filename: 'file.ts',
    },
    {
      code: `
async function test() {
  return Promise.resolve(1);
}
    `,
      filename: 'file.ts',
    },
    {
      code: `
async function test() {
  try {
    return await Promise.resolve(1);
  } catch (e) {
    return await Promise.resolve(2);
  } finally {
    console.log('cleanup');
  }
}
    `,
      filename: 'file.ts',
    },
    {
      code: `
const fn = (): any => null;
async function test() {
  return await fn();
}
    `,
      filename: 'file.ts',
    },
    {
      code: `
const fn = (): unknown => null;
async function test() {
  return await fn();
}
    `,
      filename: 'file.ts',
    },
    {
      code: `
async function test(unknownParam: unknown) {
  try {
    return await unknownParam;
  } finally {
    console.log('In finally block');
  }
}
    `,
      filename: 'file.ts',
    },
    {
      code: `
async function test() {
  if (Math.random() < 0.33) {
    return await Promise.resolve(1);
  } else if (Math.random() < 0.5) {
    return Promise.resolve(2);
  }

  try {
  } catch (e) {
    return await Promise.resolve(3);
  } finally {
    console.log('cleanup');
  }
}
      `,
      filename: 'file.ts',
      options: ['error-handling-correctness-only'],
    },
    {
      code: `
async function test() {
  try {
    const one = await Promise.resolve(1);
    return one;
  } catch (e) {
    const two = await Promise.resolve(2);
    return two;
  } finally {
    console.log('cleanup');
  }
}
    `,
      filename: 'file.ts',
    },
    {
      code: `
function test() {
  return 1;
}
      `,
      filename: 'file.ts',
      options: ['in-try-catch'],
    },
    {
      code: `
async function test() {
  return 1;
}
      `,
      filename: 'file.ts',
      options: ['in-try-catch'],
    },
    {
      code: 'const test = () => 1;',
      filename: 'file.ts',
      options: ['in-try-catch'],
    },
    {
      code: 'const test = async () => 1;',
      filename: 'file.ts',
      options: ['in-try-catch'],
    },
    {
      code: `
async function test() {
  return Promise.resolve(1);
}
      `,
      filename: 'file.ts',
      options: ['in-try-catch'],
    },
    {
      code: `
async function test() {
  try {
    return await Promise.resolve(1);
  } catch (e) {
    return await Promise.resolve(2);
  } finally {
    console.log('cleanup');
  }
}
      `,
      filename: 'file.ts',
      options: ['in-try-catch'],
    },
    {
      code: `
async function test() {
  try {
    throw 'foo';
  } catch (e) {
    return Promise.resolve(1);
  }
}
      `,
      filename: 'file.ts',
      options: ['in-try-catch'],
    },
    {
      code: `
async function test() {
  try {
    throw 'foo';
  } catch (e) {
    throw 'foo2';
  } finally {
    return Promise.resolve(1);
  }
}
      `,
      filename: 'file.ts',
      options: ['in-try-catch'],
    },
    {
      code: `
async function test() {
  try {
    const one = await Promise.resolve(1);
    return one;
  } catch (e) {
    const two = await Promise.resolve(2);
    return two;
  } finally {
    console.log('cleanup');
  }
}
      `,
      filename: 'file.ts',
      options: ['in-try-catch'],
    },
    {
      code: `
async function test() {
  return Promise.resolve(1);
}
      `,
      filename: 'file.ts',
      options: ['never'],
    },
    {
      code: 'const test = async () => Promise.resolve(1);',
      filename: 'file.ts',
      options: ['never'],
    },
    {
      code: `
async function test() {
  try {
    return Promise.resolve(1);
  } catch (e) {
    return Promise.resolve(2);
  } finally {
    console.log('cleanup');
  }
}
      `,
      filename: 'file.ts',
      options: ['never'],
    },
    {
      code: `
async function test() {
  return await Promise.resolve(1);
}
      `,
      filename: 'file.ts',
      options: ['always'],
    },
    {
      code: 'const test = async () => await Promise.resolve(1);',
      filename: 'file.ts',
      options: ['always'],
    },
    {
      code: `
async function test() {
  try {
    return await Promise.resolve(1);
  } catch (e) {
    return await Promise.resolve(2);
  } finally {
    console.log('cleanup');
  }
}
      `,
      filename: 'file.ts',
      options: ['always'],
    },
    {
      code: `
declare function foo(): Promise<boolean>;

function bar(baz: boolean): Promise<boolean> | boolean {
  if (baz) {
    return true;
  } else {
    return foo();
  }
}
      `,
      filename: 'file.ts',
      options: ['always'],
    },
    {
      code: `
async function test(): Promise<string> {
  const res = await Promise.resolve('{}');
  try {
    return JSON.parse(res);
  } catch (error) {
    return res;
  }
}
      `,
      filename: 'file.ts',
    },
    {
      code: `
async function test() {
  const res = await Promise.resolve('{}');
  try {
    async function nested() {
      return Promise.resolve('ok');
    }
    return await nested();
  } catch (error) {
    return Promise.resolve('error');
  }
}
      `,
      filename: 'file.ts',
    },
    {
      code: `
async function f() {
  try {
  } catch {
    try {
    } catch {
      return Promise.reject();
    }
  }
}
      `,
      filename: 'file.ts',
    },
    {
      code: `
async function f() {
  try {
  } finally {
    try {
    } catch {
      return Promise.reject();
    }
  }
}
      `,
      filename: 'file.ts',
    },
    {
      code: `
async function f() {
  try {
  } finally {
    try {
    } finally {
      try {
      } catch {
        return Promise.reject();
      }
    }
  }
}
      `,
      filename: 'file.ts',
    },
    {
      code: `
declare const bleh: any;
async function f() {
  using something = bleh;
  return await Promise.resolve(2);
}
      `,
      filename: 'file.ts',
    },
    {
      code: `
declare const bleh: any;
async function f() {
  await using something = bleh;
  return await Promise.resolve(2);
}
      `,
      filename: 'file.ts',
    },
    {
      code: `
declare const bleh: any;
async function f() {
  using something = bleh;
  {
    return await Promise.resolve(2);
  }
}
      `,
      filename: 'file.ts',
    },
    {
      code: `
declare const bleh: any;
async function f() {
  return Promise.resolve(2);
  using something = bleh;
}
      `,
      filename: 'file.ts',
    },
    {
      code: `
declare const bleh: any;
async function f() {
  return await Promise.resolve(2);
  using something = bleh;
}
      `,
      filename: 'file.ts',
      options: ['always'],
    },
    {
      code: `
declare function asyncFn(): Promise<unknown>;
async function returnAwait() {
  using _ = {
    [Symbol.dispose]: () => {
      console.log('dispose');
    },
  };

  return await asyncFn();
}
      `,
      filename: 'file.ts',
      options: ['in-try-catch'],
    },
    {
      code: `
declare function asyncFn(): Promise<unknown>;
async function outerFunction() {
  using _ = {
    [Symbol.dispose]: () => {
      console.log('dispose');
    },
  };

  async function innerFunction() {
    return asyncFn();
  }
}
      `,
      filename: 'file.ts',
      options: ['in-try-catch'],
    },
    {
      code: `
declare function asyncFn(): Promise<unknown>;
async function outerFunction() {
  using _ = {
    [Symbol.dispose]: () => {
      console.log('dispose');
    },
  };

  const innerFunction = async () => asyncFn();
}
      `,
      filename: 'file.ts',
      options: ['in-try-catch'],
    },
    {
      code: `
using foo = 1 as any;
return Promise.resolve(42);
      `,
      filename: 'file.ts',
    },
    {
      code: `
{
  using foo = 1 as any;
  return Promise.resolve(42);
}
      `,
      filename: 'file.ts',
    },
    {
      code: `
async function wrapper<T>(value: T) {
  return await value;
}
      `,
      filename: 'file.ts',
    },
    {
      code: `
async function wrapper<T extends unknown>(value: T) {
  return await value;
}
      `,
      filename: 'file.ts',
    },
    {
      code: `
async function wrapper<T extends any>(value: T) {
  return await value;
}
      `,
      filename: 'file.ts',
    },
    {
      code: `
class C<T> {
  async wrapper<T>(value: T) {
    return await value;
  }
}
      `,
      filename: 'file.ts',
    },
    {
      code: `
class C<R> {
  async wrapper<T extends R>(value: T) {
    return await value;
  }
}
      `,
      filename: 'file.ts',
    },
    {
      code: `
class C<R extends unknown> {
  async wrapper<T extends R>(value: T) {
    return await value;
  }
}
      `,
      filename: 'file.ts',
    },
  ],
  invalid: [
    {
      code: `
async function test() {
  return await 1;
}
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'nonPromiseAwait',
          message:
            'Returning an awaited value that is not a promise is not allowed.',
          line: 3,
          column: 10,
          endLine: 3,
          endColumn: 17,
        },
      ],
      output: `
async function test() {
  return 1;
}
      `,
    },
    {
      code: `
async function test() {
  const foo = 1;
  return await { foo };
}
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'nonPromiseAwait',
          message:
            'Returning an awaited value that is not a promise is not allowed.',
          line: 4,
          column: 10,
          endLine: 4,
          endColumn: 23,
        },
      ],
      output: `
async function test() {
  const foo = 1;
  return { foo };
}
      `,
    },
    {
      code: `
async function test() {
  const foo = 1;
  return await foo;
}
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'nonPromiseAwait',
          message:
            'Returning an awaited value that is not a promise is not allowed.',
          line: 4,
          column: 10,
          endLine: 4,
          endColumn: 19,
        },
      ],
      output: `
async function test() {
  const foo = 1;
  return foo;
}
      `,
    },
    {
      code: 'const test = async () => await 1;',
      filename: 'file.ts',
      errors: [
        {
          messageId: 'nonPromiseAwait',
          message:
            'Returning an awaited value that is not a promise is not allowed.',
          line: 1,
          column: 26,
          endLine: 1,
          endColumn: 33,
        },
      ],
      output: 'const test = async () => 1;',
    },
    {
      code: 'const test = async () => await /* comment */ 1;',
      filename: 'file.ts',
      errors: [
        {
          messageId: 'nonPromiseAwait',
          message:
            'Returning an awaited value that is not a promise is not allowed.',
          line: 1,
          column: 26,
          endLine: 1,
          endColumn: 47,
        },
      ],
      output: 'const test = async () => /* comment */ 1;',
    },
    {
      code: 'const test = async () => await Promise.resolve(1);',
      filename: 'file.ts',
      errors: [
        {
          messageId: 'disallowedPromiseAwait',
          message:
            'Returning an awaited promise is not allowed in this context.',
          line: 1,
          column: 26,
          endLine: 1,
          endColumn: 50,
        },
      ],
      output: 'const test = async () => Promise.resolve(1);',
    },
    {
      code: `
async function test() {
  try {
    return Promise.resolve(1);
  } catch (e) {
    return Promise.resolve(2);
  } finally {
    console.log('cleanup');
  }
}
      `,
      filename: 'file.ts',
      options: ['error-handling-correctness-only'],
      errors: [
        {
          messageId: 'requiredPromiseAwait',
          message: 'Returning an awaited promise is required in this context.',
          line: 4,
          column: 12,
          endLine: 4,
          endColumn: 30,
          suggestions: [
            {
              messageId: 'requiredPromiseAwaitSuggestion',
              desc: 'Add `await` before the expression. Use caution as this may impact control flow.',
              output: `
async function test() {
  try {
    return await Promise.resolve(1);
  } catch (e) {
    return Promise.resolve(2);
  } finally {
    console.log('cleanup');
  }
}
      `,
            },
          ],
        },
        {
          messageId: 'requiredPromiseAwait',
          message: 'Returning an awaited promise is required in this context.',
          line: 6,
          column: 12,
          endLine: 6,
          endColumn: 30,
          suggestions: [
            {
              messageId: 'requiredPromiseAwaitSuggestion',
              desc: 'Add `await` before the expression. Use caution as this may impact control flow.',
              output: `
async function test() {
  try {
    return Promise.resolve(1);
  } catch (e) {
    return await Promise.resolve(2);
  } finally {
    console.log('cleanup');
  }
}
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
async function test() {
  try {
    return Promise.resolve(1);
  } catch (e) {
    return Promise.resolve(2);
  } finally {
    console.log('cleanup');
  }
}
      `,
      filename: 'file.ts',
      options: ['always'],
      errors: [
        {
          messageId: 'requiredPromiseAwait',
          message: 'Returning an awaited promise is required in this context.',
          line: 4,
          column: 12,
          endLine: 4,
          endColumn: 30,
          suggestions: [
            {
              messageId: 'requiredPromiseAwaitSuggestion',
              desc: 'Add `await` before the expression. Use caution as this may impact control flow.',
              output: `
async function test() {
  try {
    return await Promise.resolve(1);
  } catch (e) {
    return Promise.resolve(2);
  } finally {
    console.log('cleanup');
  }
}
      `,
            },
          ],
        },
        {
          messageId: 'requiredPromiseAwait',
          message: 'Returning an awaited promise is required in this context.',
          line: 6,
          column: 12,
          endLine: 6,
          endColumn: 30,
          suggestions: [
            {
              messageId: 'requiredPromiseAwaitSuggestion',
              desc: 'Add `await` before the expression. Use caution as this may impact control flow.',
              output: `
async function test() {
  try {
    return Promise.resolve(1);
  } catch (e) {
    return await Promise.resolve(2);
  } finally {
    console.log('cleanup');
  }
}
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
async function test() {
  try {
    return Promise.resolve(1);
  } catch (e) {
    return Promise.resolve(2);
  } finally {
    console.log('cleanup');
  }
}
      `,
      filename: 'file.ts',
      options: ['in-try-catch'],
      errors: [
        {
          messageId: 'requiredPromiseAwait',
          message: 'Returning an awaited promise is required in this context.',
          line: 4,
          column: 12,
          endLine: 4,
          endColumn: 30,
          suggestions: [
            {
              messageId: 'requiredPromiseAwaitSuggestion',
              desc: 'Add `await` before the expression. Use caution as this may impact control flow.',
              output: `
async function test() {
  try {
    return await Promise.resolve(1);
  } catch (e) {
    return Promise.resolve(2);
  } finally {
    console.log('cleanup');
  }
}
      `,
            },
          ],
        },
        {
          messageId: 'requiredPromiseAwait',
          message: 'Returning an awaited promise is required in this context.',
          line: 6,
          column: 12,
          endLine: 6,
          endColumn: 30,
          suggestions: [
            {
              messageId: 'requiredPromiseAwaitSuggestion',
              desc: 'Add `await` before the expression. Use caution as this may impact control flow.',
              output: `
async function test() {
  try {
    return Promise.resolve(1);
  } catch (e) {
    return await Promise.resolve(2);
  } finally {
    console.log('cleanup');
  }
}
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
async function test() {
  return await Promise.resolve(1);
}
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'disallowedPromiseAwait',
          message:
            'Returning an awaited promise is not allowed in this context.',
          line: 3,
          column: 10,
          endLine: 3,
          endColumn: 34,
        },
      ],
      output: `
async function test() {
  return Promise.resolve(1);
}
      `,
    },
    {
      code: `
async function test() {
  return await 1;
}
      `,
      filename: 'file.ts',
      options: ['in-try-catch'],
      errors: [
        {
          messageId: 'nonPromiseAwait',
          message:
            'Returning an awaited value that is not a promise is not allowed.',
          line: 3,
          column: 10,
          endLine: 3,
          endColumn: 17,
        },
      ],
      output: `
async function test() {
  return 1;
}
      `,
    },
    {
      code: 'const test = async () => await 1;',
      filename: 'file.ts',
      options: ['in-try-catch'],
      errors: [
        {
          messageId: 'nonPromiseAwait',
          message:
            'Returning an awaited value that is not a promise is not allowed.',
          line: 1,
          column: 26,
          endLine: 1,
          endColumn: 33,
        },
      ],
      output: 'const test = async () => 1;',
    },
    {
      code: 'const test = async () => await Promise.resolve(1);',
      filename: 'file.ts',
      options: ['in-try-catch'],
      errors: [
        {
          messageId: 'disallowedPromiseAwait',
          message:
            'Returning an awaited promise is not allowed in this context.',
          line: 1,
          column: 26,
          endLine: 1,
          endColumn: 50,
        },
      ],
      output: 'const test = async () => Promise.resolve(1);',
    },
    {
      code: `
async function test() {
  return await Promise.resolve(1);
}
      `,
      filename: 'file.ts',
      options: ['in-try-catch'],
      errors: [
        {
          messageId: 'disallowedPromiseAwait',
          message:
            'Returning an awaited promise is not allowed in this context.',
          line: 3,
          column: 10,
          endLine: 3,
          endColumn: 34,
        },
      ],
      output: `
async function test() {
  return Promise.resolve(1);
}
      `,
    },
    {
      code: `
async function test() {
  return await 1;
}
      `,
      filename: 'file.ts',
      options: ['never'],
      errors: [
        {
          messageId: 'nonPromiseAwait',
          message:
            'Returning an awaited value that is not a promise is not allowed.',
          line: 3,
          column: 10,
          endLine: 3,
          endColumn: 17,
        },
      ],
      output: `
async function test() {
  return 1;
}
      `,
    },
    {
      code: `
async function test() {
  try {
    return await Promise.resolve(1);
  } catch (e) {
    return await Promise.resolve(2);
  } finally {
    console.log('cleanup');
  }
}
      `,
      filename: 'file.ts',
      options: ['never'],
      errors: [
        {
          messageId: 'disallowedPromiseAwait',
          message:
            'Returning an awaited promise is not allowed in this context.',
          line: 4,
          column: 12,
          endLine: 4,
          endColumn: 36,
          suggestions: [
            {
              messageId: 'disallowedPromiseAwaitSuggestion',
              desc: 'Remove `await` before the expression. Use caution as this may impact control flow.',
              output: `
async function test() {
  try {
    return Promise.resolve(1);
  } catch (e) {
    return await Promise.resolve(2);
  } finally {
    console.log('cleanup');
  }
}
      `,
            },
          ],
        },
        {
          messageId: 'disallowedPromiseAwait',
          message:
            'Returning an awaited promise is not allowed in this context.',
          line: 6,
          column: 12,
          endLine: 6,
          endColumn: 36,
          suggestions: [
            {
              messageId: 'disallowedPromiseAwaitSuggestion',
              desc: 'Remove `await` before the expression. Use caution as this may impact control flow.',
              output: `
async function test() {
  try {
    return await Promise.resolve(1);
  } catch (e) {
    return Promise.resolve(2);
  } finally {
    console.log('cleanup');
  }
}
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
async function test() {
  return await Promise.resolve(1);
}
      `,
      filename: 'file.ts',
      options: ['never'],
      errors: [
        {
          messageId: 'disallowedPromiseAwait',
          message:
            'Returning an awaited promise is not allowed in this context.',
          line: 3,
          column: 10,
          endLine: 3,
          endColumn: 34,
        },
      ],
      output: `
async function test() {
  return Promise.resolve(1);
}
      `,
    },
    {
      code: `
async function test() {
  return await 1;
}
      `,
      filename: 'file.ts',
      options: ['always'],
      errors: [
        {
          messageId: 'nonPromiseAwait',
          message:
            'Returning an awaited value that is not a promise is not allowed.',
          line: 3,
          column: 10,
          endLine: 3,
          endColumn: 17,
        },
      ],
      output: `
async function test() {
  return 1;
}
      `,
    },
    {
      code: `
async function test() {
  return Promise.resolve(1);
}
      `,
      filename: 'file.ts',
      options: ['always'],
      errors: [
        {
          messageId: 'requiredPromiseAwait',
          message: 'Returning an awaited promise is required in this context.',
          line: 3,
          column: 10,
          endLine: 3,
          endColumn: 28,
        },
      ],
      output: `
async function test() {
  return await Promise.resolve(1);
}
      `,
    },
    {
      code: 'const test = async () => Promise.resolve(1);',
      filename: 'file.ts',
      options: ['always'],
      errors: [
        {
          messageId: 'requiredPromiseAwait',
          message: 'Returning an awaited promise is required in this context.',
          line: 1,
          column: 26,
          endLine: 1,
          endColumn: 44,
        },
      ],
      output: 'const test = async () => await Promise.resolve(1);',
    },
    {
      code: `
async function foo() {}
async function bar() {}
async function baz() {}
async function qux() {}
async function buzz() {
  return (await foo()) ? bar() : baz();
}
      `,
      filename: 'file.ts',
      options: ['always'],
      errors: [
        {
          messageId: 'requiredPromiseAwait',
          message: 'Returning an awaited promise is required in this context.',
          line: 7,
          column: 26,
          endLine: 7,
          endColumn: 31,
        },
        {
          messageId: 'requiredPromiseAwait',
          message: 'Returning an awaited promise is required in this context.',
          line: 7,
          column: 34,
          endLine: 7,
          endColumn: 39,
        },
      ],
      output: `
async function foo() {}
async function bar() {}
async function baz() {}
async function qux() {}
async function buzz() {
  return (await foo()) ? await bar() : await baz();
}
      `,
    },
    {
      code: `
async function foo() {}
async function bar() {}
async function baz() {}
async function qux() {}
async function buzz() {
  return (await foo())
    ? (
      bar ? bar() : baz()
    ) : baz ? baz() : bar();
}
      `,
      filename: 'file.ts',
      options: ['always'],
      errors: [
        {
          messageId: 'requiredPromiseAwait',
          message: 'Returning an awaited promise is required in this context.',
          line: 9,
          column: 13,
          endLine: 9,
          endColumn: 18,
        },
        {
          messageId: 'requiredPromiseAwait',
          message: 'Returning an awaited promise is required in this context.',
          line: 9,
          column: 21,
          endLine: 9,
          endColumn: 26,
        },
        {
          messageId: 'requiredPromiseAwait',
          message: 'Returning an awaited promise is required in this context.',
          line: 10,
          column: 15,
          endLine: 10,
          endColumn: 20,
        },
        {
          messageId: 'requiredPromiseAwait',
          message: 'Returning an awaited promise is required in this context.',
          line: 10,
          column: 23,
          endLine: 10,
          endColumn: 28,
        },
      ],
      output: `
async function foo() {}
async function bar() {}
async function baz() {}
async function qux() {}
async function buzz() {
  return (await foo())
    ? (
      bar ? await bar() : await baz()
    ) : baz ? await baz() : await bar();
}
      `,
    },
    {
      code: `
async function foo() {}
async function bar() {}
async function buzz() {
  return (await foo()) ? await 1 : bar();
}
      `,
      filename: 'file.ts',
      options: ['always'],
      errors: [
        {
          messageId: 'nonPromiseAwait',
          message:
            'Returning an awaited value that is not a promise is not allowed.',
          line: 5,
          column: 26,
          endLine: 5,
          endColumn: 33,
        },
        {
          messageId: 'requiredPromiseAwait',
          message: 'Returning an awaited promise is required in this context.',
          line: 5,
          column: 36,
          endLine: 5,
          endColumn: 41,
        },
      ],
      output: `
async function foo() {}
async function bar() {}
async function buzz() {
  return (await foo()) ? 1 : await bar();
}
      `,
    },
    {
      code: `
async function foo() {}
async function bar() {}
async function baz() {}
const buzz = async () => ((await foo()) ? bar() : baz());
      `,
      filename: 'file.ts',
      options: ['always'],
      errors: [
        {
          messageId: 'requiredPromiseAwait',
          message: 'Returning an awaited promise is required in this context.',
          line: 5,
          column: 43,
          endLine: 5,
          endColumn: 48,
        },
        {
          messageId: 'requiredPromiseAwait',
          message: 'Returning an awaited promise is required in this context.',
          line: 5,
          column: 51,
          endLine: 5,
          endColumn: 56,
        },
      ],
      output: `
async function foo() {}
async function bar() {}
async function baz() {}
const buzz = async () => ((await foo()) ? await bar() : await baz());
      `,
    },
    {
      code: `
async function foo() {}
async function bar() {}
const buzz = async () => ((await foo()) ? await 1 : bar());
      `,
      filename: 'file.ts',
      options: ['always'],
      errors: [
        {
          messageId: 'nonPromiseAwait',
          message:
            'Returning an awaited value that is not a promise is not allowed.',
          line: 4,
          column: 43,
          endLine: 4,
          endColumn: 50,
        },
        {
          messageId: 'requiredPromiseAwait',
          message: 'Returning an awaited promise is required in this context.',
          line: 4,
          column: 53,
          endLine: 4,
          endColumn: 58,
        },
      ],
      output: `
async function foo() {}
async function bar() {}
const buzz = async () => ((await foo()) ? 1 : await bar());
      `,
    },
    {
      code: `
async function test<T>(): Promise<T> {
  const res = await fetch('...');
  try {
    return res.json() as Promise<T>;
  } catch (err) {
    throw Error('Request Failed.');
  }
}
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'requiredPromiseAwait',
          message: 'Returning an awaited promise is required in this context.',
          line: 5,
          column: 12,
          endLine: 5,
          endColumn: 36,
          suggestions: [
            {
              messageId: 'requiredPromiseAwaitSuggestion',
              desc: 'Add `await` before the expression. Use caution as this may impact control flow.',
              output: `
async function test<T>(): Promise<T> {
  const res = await fetch('...');
  try {
    return await (res.json() as Promise<T>);
  } catch (err) {
    throw Error('Request Failed.');
  }
}
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
async function test() {
  try {
    const callback1 = function () {};
    const callback2 = async function () {};
    function callback3() {}
    async function callback4() {}
    const callback5 = () => {};
    const callback6 = async () => {};
    return Promise.resolve('try');
  } finally {
    return Promise.resolve('finally');
  }
}
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'requiredPromiseAwait',
          message: 'Returning an awaited promise is required in this context.',
          line: 10,
          column: 12,
          endLine: 10,
          endColumn: 34,
          suggestions: [
            {
              messageId: 'requiredPromiseAwaitSuggestion',
              desc: 'Add `await` before the expression. Use caution as this may impact control flow.',
              output: `
async function test() {
  try {
    const callback1 = function () {};
    const callback2 = async function () {};
    function callback3() {}
    async function callback4() {}
    const callback5 = () => {};
    const callback6 = async () => {};
    return await Promise.resolve('try');
  } finally {
    return Promise.resolve('finally');
  }
}
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
async function bar() {}
async function foo() {
  try {
    return undefined || bar();
  } catch {}
}
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'requiredPromiseAwait',
          message: 'Returning an awaited promise is required in this context.',
          line: 5,
          column: 12,
          endLine: 5,
          endColumn: 30,
          suggestions: [
            {
              messageId: 'requiredPromiseAwaitSuggestion',
              desc: 'Add `await` before the expression. Use caution as this may impact control flow.',
              output: `
async function bar() {}
async function foo() {
  try {
    return await (undefined || bar());
  } catch {}
}
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
async function bar() {}
async function foo() {
  try {
    return bar() || undefined || bar();
  } catch {}
}
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'requiredPromiseAwait',
          message: 'Returning an awaited promise is required in this context.',
          line: 5,
          column: 12,
          endLine: 5,
          endColumn: 39,
          suggestions: [
            {
              messageId: 'requiredPromiseAwaitSuggestion',
              desc: 'Add `await` before the expression. Use caution as this may impact control flow.',
              output: `
async function bar() {}
async function foo() {
  try {
    return await (bar() || undefined || bar());
  } catch {}
}
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
async function bar() {}
async function func1() {
  try {
    return null ?? bar();
  } catch {}
}
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'requiredPromiseAwait',
          message: 'Returning an awaited promise is required in this context.',
          line: 5,
          column: 12,
          endLine: 5,
          endColumn: 25,
          suggestions: [
            {
              messageId: 'requiredPromiseAwaitSuggestion',
              desc: 'Add `await` before the expression. Use caution as this may impact control flow.',
              output: `
async function bar() {}
async function func1() {
  try {
    return await (null ?? bar());
  } catch {}
}
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
async function bar() {}
async function func2() {
  try {
    return 1 && bar();
  } catch {}
}
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'requiredPromiseAwait',
          message: 'Returning an awaited promise is required in this context.',
          line: 5,
          column: 12,
          endLine: 5,
          endColumn: 22,
          suggestions: [
            {
              messageId: 'requiredPromiseAwaitSuggestion',
              desc: 'Add `await` before the expression. Use caution as this may impact control flow.',
              output: `
async function bar() {}
async function func2() {
  try {
    return await (1 && bar());
  } catch {}
}
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
const foo = {
  bar: async function () {},
};
async function func3() {
  try {
    return foo.bar();
  } catch {}
}
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'requiredPromiseAwait',
          message: 'Returning an awaited promise is required in this context.',
          line: 7,
          column: 12,
          endLine: 7,
          endColumn: 21,
          suggestions: [
            {
              messageId: 'requiredPromiseAwaitSuggestion',
              desc: 'Add `await` before the expression. Use caution as this may impact control flow.',
              output: `
const foo = {
  bar: async function () {},
};
async function func3() {
  try {
    return await foo.bar();
  } catch {}
}
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
class X {
  async bar() {
    return;
  }
  async func2() {
    try {
      return this.bar();
    } catch {}
  }
}
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'requiredPromiseAwait',
          message: 'Returning an awaited promise is required in this context.',
          line: 8,
          column: 14,
          endLine: 8,
          endColumn: 24,
          suggestions: [
            {
              messageId: 'requiredPromiseAwaitSuggestion',
              desc: 'Add `await` before the expression. Use caution as this may impact control flow.',
              output: `
class X {
  async bar() {
    return;
  }
  async func2() {
    try {
      return await this.bar();
    } catch {}
  }
}
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
async function test() {
  const res = await Promise.resolve('{}');
  try {
    async function nested() {
      return Promise.resolve('ok');
    }
    return await nested();
  } catch (error) {
    return await Promise.resolve('error');
  }
}
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'disallowedPromiseAwait',
          message:
            'Returning an awaited promise is not allowed in this context.',
          line: 10,
          column: 12,
          endLine: 10,
          endColumn: 42,
        },
      ],
      output: `
async function test() {
  const res = await Promise.resolve('{}');
  try {
    async function nested() {
      return Promise.resolve('ok');
    }
    return await nested();
  } catch (error) {
    return Promise.resolve('error');
  }
}
      `,
    },
    {
      code: `
async function f() {
  try {
    try {
    } finally {
      // affects error handling of outer catch
      return Promise.reject();
    }
  } catch {}
}
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'requiredPromiseAwait',
          message: 'Returning an awaited promise is required in this context.',
          line: 7,
          column: 14,
          endLine: 7,
          endColumn: 30,
          suggestions: [
            {
              messageId: 'requiredPromiseAwaitSuggestion',
              desc: 'Add `await` before the expression. Use caution as this may impact control flow.',
              output: `
async function f() {
  try {
    try {
    } finally {
      // affects error handling of outer catch
      return await Promise.reject();
    }
  } catch {}
}
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
async function f() {
  try {
    try {
    } catch {
      // affects error handling of outer catch
      return Promise.reject();
    }
  } catch {}
}
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'requiredPromiseAwait',
          message: 'Returning an awaited promise is required in this context.',
          line: 7,
          column: 14,
          endLine: 7,
          endColumn: 30,
          suggestions: [
            {
              messageId: 'requiredPromiseAwaitSuggestion',
              desc: 'Add `await` before the expression. Use caution as this may impact control flow.',
              output: `
async function f() {
  try {
    try {
    } catch {
      // affects error handling of outer catch
      return await Promise.reject();
    }
  } catch {}
}
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
async function f() {
  try {
  } catch {
    try {
    } finally {
      try {
      } catch {
        return Promise.reject();
      }
    }
  } finally {
  }
}
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'requiredPromiseAwait',
          message: 'Returning an awaited promise is required in this context.',
          line: 9,
          column: 16,
          endLine: 9,
          endColumn: 32,
          suggestions: [
            {
              messageId: 'requiredPromiseAwaitSuggestion',
              desc: 'Add `await` before the expression. Use caution as this may impact control flow.',
              output: `
async function f() {
  try {
  } catch {
    try {
    } finally {
      try {
      } catch {
        return await Promise.reject();
      }
    }
  } finally {
  }
}
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
declare const bleh: any;
async function f() {
  if (cond) {
    using something = bleh;
    if (anotherCondition) {
      return Promise.resolve(2);
    }
  }
}
      `,
      filename: 'file.ts',
      options: ['always'],
      errors: [
        {
          messageId: 'requiredPromiseAwait',
          message: 'Returning an awaited promise is required in this context.',
          line: 7,
          column: 14,
          endLine: 7,
          endColumn: 32,
          suggestions: [
            {
              messageId: 'requiredPromiseAwaitSuggestion',
              desc: 'Add `await` before the expression. Use caution as this may impact control flow.',
              output: `
declare const bleh: any;
async function f() {
  if (cond) {
    using something = bleh;
    if (anotherCondition) {
      return await Promise.resolve(2);
    }
  }
}
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
declare const bleh: any;
async function f() {
  if (cond) {
    await using something = bleh;
    if (anotherCondition) {
      return Promise.resolve(2);
    }
  }
}
      `,
      filename: 'file.ts',
      options: ['always'],
      errors: [
        {
          messageId: 'requiredPromiseAwait',
          message: 'Returning an awaited promise is required in this context.',
          line: 7,
          column: 14,
          endLine: 7,
          endColumn: 32,
          suggestions: [
            {
              messageId: 'requiredPromiseAwaitSuggestion',
              desc: 'Add `await` before the expression. Use caution as this may impact control flow.',
              output: `
declare const bleh: any;
async function f() {
  if (cond) {
    await using something = bleh;
    if (anotherCondition) {
      return await Promise.resolve(2);
    }
  }
}
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
declare const bleh: any;
async function f() {
  if (cond) {
    using something = bleh;
  } else if (anotherCondition) {
    return Promise.resolve(2);
  }
}
      `,
      filename: 'file.ts',
      options: ['always'],
      errors: [
        {
          messageId: 'requiredPromiseAwait',
          message: 'Returning an awaited promise is required in this context.',
          line: 7,
          column: 12,
          endLine: 7,
          endColumn: 30,
        },
      ],
      output: `
declare const bleh: any;
async function f() {
  if (cond) {
    using something = bleh;
  } else if (anotherCondition) {
    return await Promise.resolve(2);
  }
}
      `,
    },
    {
      code: `
declare function asyncFn(): Promise<unknown>;
async function outerFunction() {
  using _ = {
    [Symbol.dispose]: () => {
      console.log('dispose');
    },
  };

  async function innerFunction() {
    return await asyncFn();
  }
}
      `,
      filename: 'file.ts',
      options: ['in-try-catch'],
      errors: [
        {
          messageId: 'disallowedPromiseAwait',
          message:
            'Returning an awaited promise is not allowed in this context.',
          line: 11,
          column: 12,
          endLine: 11,
          endColumn: 27,
        },
      ],
      output: `
declare function asyncFn(): Promise<unknown>;
async function outerFunction() {
  using _ = {
    [Symbol.dispose]: () => {
      console.log('dispose');
    },
  };

  async function innerFunction() {
    return asyncFn();
  }
}
      `,
    },
    {
      code: `
declare function asyncFn(): Promise<unknown>;
async function outerFunction() {
  using _ = {
    [Symbol.dispose]: () => {
      console.log('dispose');
    },
  };

  const innerFunction = async () => await asyncFn();
}
      `,
      filename: 'file.ts',
      options: ['in-try-catch'],
      errors: [
        {
          messageId: 'disallowedPromiseAwait',
          message:
            'Returning an awaited promise is not allowed in this context.',
          line: 10,
          column: 37,
          endLine: 10,
          endColumn: 52,
        },
      ],
      output: `
declare function asyncFn(): Promise<unknown>;
async function outerFunction() {
  using _ = {
    [Symbol.dispose]: () => {
      console.log('dispose');
    },
  };

  const innerFunction = async () => asyncFn();
}
      `,
    },
    {
      code: `
async function wrapper<T extends number>(value: T) {
  return await value;
}
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'nonPromiseAwait',
          message:
            'Returning an awaited value that is not a promise is not allowed.',
          line: 3,
          column: 10,
          endLine: 3,
          endColumn: 21,
        },
      ],
      output: `
async function wrapper<T extends number>(value: T) {
  return value;
}
      `,
    },
    {
      code: `
class C<T> {
  async wrapper<T extends string>(value: T) {
    return await value;
  }
}
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'nonPromiseAwait',
          message:
            'Returning an awaited value that is not a promise is not allowed.',
          line: 4,
          column: 12,
          endLine: 4,
          endColumn: 23,
        },
      ],
      output: `
class C<T> {
  async wrapper<T extends string>(value: T) {
    return value;
  }
}
      `,
    },
    {
      code: `
class C<R extends number> {
  async wrapper<T extends R>(value: T) {
    return await value;
  }
}
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'nonPromiseAwait',
          message:
            'Returning an awaited value that is not a promise is not allowed.',
          line: 4,
          column: 12,
          endLine: 4,
          endColumn: 23,
        },
      ],
      output: `
class C<R extends number> {
  async wrapper<T extends R>(value: T) {
    return value;
  }
}
      `,
    },
  ],
});
