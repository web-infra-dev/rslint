import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-restricted-rstest-methods', {} as never, {
  valid: [
    { code: `rs.fn()` },
    { code: `rs.fn()`, options: [{}] },
    { code: `rs.spyOn(target, 'method')`, options: [{ fn: null }] },
    { code: `helpers.fn()`, options: [{ fn: null }] },
    { code: `const rs = { fn: () => {} }; rs.fn();`, options: [{ fn: null }] },
    {
      code: `import { rs as mocker } from '@rstest/core'; mocker.mock('./m');`,
      options: [{ mock: null }],
    },
    { code: `rs['mock']('./m')`, options: [{ mock: null }] },
    { code: `rs.mock?.('./m')`, options: [{ mock: null }] },
    {
      code: `const rs = { importActual: async (p) => ({}) }; rs.importActual('./m');`,
      options: [{ importActual: null }],
    },
  ],
  invalid: [
    {
      code: `rs.fn()`,
      options: [{ fn: null }],
      errors: [
        {
          messageId: 'restrictedRstestMethod',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 6,
        },
      ],
    },
    {
      code: `rs.fn()`,
      options: [{ fn: 'Use the shared factory instead.' }],
      errors: [
        {
          messageId: 'restrictedRstestMethodWithMessage',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 6,
        },
      ],
    },
    {
      code: `rs['fn']()`,
      options: [{ fn: null }],
      errors: [
        {
          messageId: 'restrictedRstestMethod',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: `rs.mock('./m')`,
      options: [{ mock: null }],
      errors: [
        {
          messageId: 'restrictedRstestMethod',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: `const rs = { mock() {} }; rs.mock('./m');`,
      options: [{ mock: null }],
      errors: [
        {
          messageId: 'restrictedRstestMethod',
          line: 1,
          column: 30,
          endLine: 1,
          endColumn: 34,
        },
      ],
    },
  ],
});
