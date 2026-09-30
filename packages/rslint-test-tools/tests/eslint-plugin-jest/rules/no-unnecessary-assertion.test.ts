import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-unnecessary-assertion', {} as never, {
  valid: [
    'expect(a).toBe(b)',
    'expect.toBeNull()',
    'declare function mx(): string | null; expect(mx()).toBeNull();',
    'declare function mx(): string | undefined; expect(mx()).not.toBeUndefined();',
    'declare function mx(): unknown; expect(mx()).toBeDefined();',
    'declare function mx(): string | number; expect(mx()).toBeNaN();',
    'declare function mx(): Promise<string>; async function run() { await expect(mx()).resolves.toBeNull(); }',
    "import { expect } from 'vitest'; expect('hello').toBeNull();",
  ],
  invalid: [
    {
      code: 'expect(0).toBeNull()',
      errors: [
        {
          message: 'Unnecessary assertion, subject cannot be null',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: 'expect([]).not.toBeUndefined()',
      errors: [
        {
          messageId: 'unnecessaryAssertion',
          message: 'Unnecessary assertion, subject cannot be undefined',
        },
      ],
    },
    {
      code: 'expect("hello world").toBeNaN()',
      errors: [
        {
          messageId: 'unnecessaryAssertion',
          message: 'Unnecessary assertion, subject cannot be a number',
        },
      ],
    },
    {
      code: "import { expect as check } from '@jest/globals'; check('hello')['toBeDefined']();",
      errors: [{ messageId: 'unnecessaryAssertion' }],
    },
    {
      code: 'function check<T>(value: T) { expect(value).toBeNull(); }',
      errors: [{ messageId: 'unnecessaryAssertion' }],
    },
  ],
});
