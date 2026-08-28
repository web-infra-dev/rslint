import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('prefer-import-in-mock', {} as never, {
  valid: [
    { code: `rs.mock(import('./sum'))` },
    { code: `rstest.doMock(import('./sum'), { spy: true })` },
    { code: `rs.mockRequire('./sum')` },
    { code: `rs.doMockRequire('./sum')` },
    { code: `rs.unmock('./sum')` },
    { code: `rs.mock<{ sum: number }>('./sum', () => ({ sum: 1 }))` },
    { code: `rs.mock(modulePath)` },
    { code: `rs?.mock('./sum')` },
    { code: `consume(rs.mock('./sum'))` },
    { code: `const mocked = rs.mock('./sum')` },
    { code: `rs.mock(...args)` },
    { code: `rs.mock('./sum', () => ({ sum: 1 }), extra)` },
    { code: 'rs.mock(`./sum`)' },
    { code: `await rs.mock('./sum')` },
    {
      code: `import { rstest as vi } from '@rstest/core';\nvi.mock('./sum')`,
    },
  ],
  invalid: [
    {
      code: `rs.mock('./sum', () => ({ sum: () => 0 }))`,
      output: `rs.mock(import('./sum'), () => ({ sum: () => 0 }))`,
      errors: [
        {
          messageId: 'preferImport',
          message: `Replace './sum' with import('./sum')`,
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      code: `rstest.doMock("./sum")`,
      output: `rstest.doMock(import("./sum"))`,
      errors: [
        {
          messageId: 'preferImport',
          message: `Replace "./sum" with import("./sum")`,
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: `(rs as any)!.mock(('./sum'))`,
      output: `(rs as any)!.mock(import('./sum'))`,
      errors: [
        {
          messageId: 'preferImport',
          message: `Replace './sum' with import('./sum')`,
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      code: `const rs = getMocker();\nrs.mock('./sum')`,
      output: `const rs = getMocker();\nrs.mock(import('./sum'))`,
      errors: [
        {
          messageId: 'preferImport',
          message: `Replace './sum' with import('./sum')`,
          line: 2,
          column: 9,
          endLine: 2,
          endColumn: 16,
        },
      ],
    },
    {
      code: `rs.mock('./sum')`,
      options: [{ fixable: false }],
      errors: [
        {
          messageId: 'preferImport',
          message: `Replace './sum' with import('./sum')`,
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
  ],
});
