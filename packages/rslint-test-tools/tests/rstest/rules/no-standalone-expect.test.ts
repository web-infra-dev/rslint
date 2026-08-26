import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-standalone-expect', {} as never, {
  valid: [
    {
      code: "expect.any(String); expect.extend({}); expect.not.stringContaining('value');",
    },
    {
      code: `
        describe('suite', () => {
          test('case', () => expect(1).toBe(1));
        });
      `,
    },
    {
      code: `
        test.for([1])('case', (value, context) => {
          context.expect(value).toBe(1);
        });
      `,
    },
    {
      code: `
        test('case', ({ expect }) => expect(1).toBe(1));
        test('case', context => context.expect(1).toBe(1));
      `,
    },
    {
      code: `
        test('case', () => expect(1).toBe(1), 1000);
        test('case', { timeout: 1 }, () => expect(1).toBe(1));
      `,
    },
    {
      code: `
        import.meta.rstest.test('case', () => {
          import.meta.rstest.expect(1).to.be.ok;
        });
      `,
    },
    {
      code: `
        describe('suite', () => {
          const helper = () => expect(1).toBe(1);
          test('case', helper);
        });
      `,
    },
    {
      code: `
        describe('suite', () => {
          t('case', () => expect(1).toBe(1));
        });
      `,
      options: [{ additionalTestBlockFunctions: ['t'] }],
    },
    {
      code: `
        import { expect } from 'vitest';
        expect(1).toBe(1);
      `,
    },
  ],
  invalid: [
    {
      code: 'expect(1).toBe(1);',
      errors: [
        {
          messageId: 'unexpectedExpect',
          message: 'Expect must be inside of a test block',
          line: 1,
          column: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: `
        describe('suite', () => {
          expect(1).toBe(1);
        });
      `,
      errors: [{ messageId: 'unexpectedExpect', line: 3, column: 11 }],
    },
    {
      code: `
        describe('suite', () => {
          test('case', () => expect(1).toBe(1));
          expect(2).toBe(2);
        });
      `,
      errors: [{ messageId: 'unexpectedExpect', line: 4, column: 11 }],
    },
    {
      code: 'expect.hasAssertions(); expect.assertions(1);',
      errors: [
        { messageId: 'unexpectedExpect', line: 1, column: 1 },
        { messageId: 'unexpectedExpect', line: 1, column: 25 },
      ],
    },
    {
      code: `
        import { describe as d, expect as check } from '@rstest/core';
        d('suite', () => check(1).toBe(1));
      `,
      errors: [{ messageId: 'unexpectedExpect', line: 3, column: 26 }],
    },
    {
      code: 'import.meta.rstest.expect(1).toBe(1);',
      errors: [{ messageId: 'unexpectedExpect', line: 1, column: 1 }],
    },
    {
      code: `
        const { expect: check } = require('@rstest/core');
        check(1).toBe(1);
      `,
      errors: [{ messageId: 'unexpectedExpect', line: 3, column: 9 }],
    },
    {
      code: `
        if (import.meta.rstest) {
          const api = import.meta.rstest;
          api.expect(1).toBe(1);
        }
      `,
      errors: [{ messageId: 'unexpectedExpect', line: 4, column: 11 }],
    },
    {
      code: `
        describe('suite', () => {
          expect(1).to.be.ok;
        });
      `,
      errors: [{ messageId: 'unexpectedExpect', line: 3, column: 11 }],
    },
    {
      code: `
        test.todo(expect(1).toBe(1));
        test('case', { timeout: expect(1).toBe(1) }, () => {});
      `,
      errors: [
        { messageId: 'unexpectedExpect', line: 2, column: 19 },
        { messageId: 'unexpectedExpect', line: 3, column: 33 },
      ],
    },
  ],
});
