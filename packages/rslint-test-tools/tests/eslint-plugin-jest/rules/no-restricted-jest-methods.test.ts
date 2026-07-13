import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-restricted-jest-methods', {} as never, {
  valid: [
    { code: 'jest' },
    { code: 'jest()' },
    { code: 'jest.mock()' },
    { code: 'expect(a).rejects;' },
    { code: 'expect(a);' },
    {
      code: `
        import { jest } from '@jest/globals';

        jest;
      `,
    },
  ],
  invalid: [
    {
      code: 'jest.fn()',
      options: [{ fn: null }],
      errors: [
        {
          messageId: 'restrictedJestMethod',
          data: {
            message: null,
            restriction: 'fn',
          },
          column: 6,
          line: 1,
        },
      ],
    },
    {
      code: 'jest["fn"]()',
      options: [{ fn: null }],
      errors: [
        {
          messageId: 'restrictedJestMethod',
          data: {
            message: null,
            restriction: 'fn',
          },
          column: 6,
          line: 1,
        },
      ],
    },
    {
      code: 'jest.mock()',
      options: [{ mock: 'Do not use mocks' }],
      errors: [
        {
          messageId: 'restrictedJestMethodWithMessage',
          data: {
            message: 'Do not use mocks',
            restriction: 'mock',
          },
          column: 6,
          line: 1,
        },
      ],
    },
    {
      code: 'jest["mock"]()',
      options: [{ mock: 'Do not use mocks' }],
      errors: [
        {
          messageId: 'restrictedJestMethodWithMessage',
          data: {
            message: 'Do not use mocks',
            restriction: 'mock',
          },
          column: 6,
          line: 1,
        },
      ],
    },
    {
      code: `
        import { jest } from '@jest/globals';

        jest.advanceTimersByTime();
      `,
      options: [{ advanceTimersByTime: null }],
      errors: [
        {
          messageId: 'restrictedJestMethod',
          data: {
            message: null,
            restriction: 'advanceTimersByTime',
          },
          column: 6,
          line: 3,
        },
      ],
    },
  ],
});
