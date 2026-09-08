import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-async-mock-factory', {} as never, {
  valid: [
    { code: `rs.mock('./sum')` },
    { code: `rs.mock('./sum', { spy: true })` },
    { code: `rs.mock('./sum', () => ({ sum: () => 0 }))` },
    { code: `rs.mock('./sum', function* () { yield 1; })` },
    { code: `rs.mock('./sum', async function* () { yield 1; })` },
    { code: `rs.mock('./sum', () => ({ then() {} }))` },
    { code: `rs['mock']('./sum', async () => ({ sum: 0 }))` },
    { code: `rs?.mock('./sum', async () => ({ sum: 0 }))` },
    { code: `rs.mock?.('./sum', async () => ({ sum: 0 }))` },
    { code: `const mocked = rs.mock('./sum', async () => ({ sum: 0 }))` },
    { code: `rs.mock('./sum', async () => ({ sum: 0 }), extra)` },
    { code: `rs.mock('./sum', ...factories)` },
    {
      code: `import { rs as mocker } from '@rstest/core';\nmocker.mock('./sum', async () => ({ sum: 0 }))`,
    },
  ],
  invalid: [
    {
      code: `rs.mock('./sum', async () => ({ sum: () => 0 }))`,
      errors: [
        {
          messageId: 'asyncMockFactory',
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 48,
        },
      ],
    },
    {
      code: `rstest.doMock('./sum', () => Promise.resolve({ sum: 0 }))`,
      errors: [
        {
          messageId: 'asyncMockFactory',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 57,
        },
      ],
    },
    {
      code: `rs.mockRequire('./sum', () => import('./sum'))`,
      errors: [
        {
          messageId: 'asyncMockFactory',
          line: 1,
          column: 25,
          endLine: 1,
          endColumn: 46,
        },
      ],
    },
    {
      code: `rs.doMockRequire('./sum', async function () { return { sum: 0 }; })`,
      errors: [
        {
          messageId: 'asyncMockFactory',
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 67,
        },
      ],
    },
    {
      code: `rs.mock<{ sum: number }>('./sum', async () => ({ sum: 0 }))`,
      errors: [
        {
          messageId: 'asyncMockFactory',
          line: 1,
          column: 35,
          endLine: 1,
          endColumn: 59,
        },
      ],
    },
  ],
});
