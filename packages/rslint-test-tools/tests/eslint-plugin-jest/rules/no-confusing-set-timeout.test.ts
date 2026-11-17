// Upstream reference: eslint-plugin-jest@29.16.0; includes existing Rslint regressions.
// Additional cases from 29.16.6 retain their input and parser configuration.
// Complete assertions record current native output; documented differences remain separate.
import {
  RuleTester,
  type TestCase,
  type ValidTestCase,
} from '@rslint/rule-tester';
import { fixtureConfig } from '../config';

// Shared source templates; reporting ranges remain explicit in each matrix.
const testSetup = `  beforeEach(async () => { await new Promise(resolve => { setTimeout(resolve, 10000).unref(); });});
  it('A.1', async () => { await new Promise(resolve => { setTimeout(resolve, 10000).unref(); });});
  it('A.2', async () => { await new Promise(resolve => { setTimeout(resolve, 10000).unref(); });});
});`;
const testSetup2 = `          beforeEach(async () => { await new Promise(resolve => { setTimeout(resolve, 10000).unref(); });});
          it('A.1', async () => { await new Promise(resolve => { setTimeout(resolve, 10000).unref(); });});
          it('A.2', async () => { await new Promise(resolve => { setTimeout(resolve, 10000).unref(); });});
        });
     \u0020`;

new RuleTester(fixtureConfig()).run('no-confusing-set-timeout', {
  valid: [
    ...(
      [
        `
      jest.setTimeout(1000);
      describe('A', () => {
        beforeEach(async () => { await new Promise(resolve => { setTimeout(resolve, 10000).unref(); });});
        it('A.1', async () => { await new Promise(resolve => { setTimeout(resolve, 10000).unref(); });});
        it('A.2', async () => { await new Promise(resolve => { setTimeout(resolve, 10000).unref(); });});
      });
   \u0020`,
        `
      jest.setTimeout(1000);
      window.setTimeout(6000)
      describe('A', () => {
        beforeEach(async () => { await new Promise(resolve => { setTimeout(resolve, 10000).unref(); });});
        it('test foo', async () => { await new Promise(resolve => { setTimeout(resolve, 10000).unref(); });});
      });
   \u0020`,
        `
        import { handler } from 'dep/mod';
        jest.setTimeout(800);
        describe('A', () => {
` + testSetup2,
        `
      function handler() {}
      jest.setTimeout(800);
      describe('A', () => {
        beforeEach(async () => { await new Promise(resolve => { setTimeout(resolve, 10000).unref(); });});
        it('A.1', async () => { await new Promise(resolve => { setTimeout(resolve, 10000).unref(); });});
        it('A.2', async () => { await new Promise(resolve => { setTimeout(resolve, 10000).unref(); });});
      });
   \u0020`,
        `
      const { handler } = require('dep/mod');
      jest.setTimeout(800);
      describe('A', () => {
        beforeEach(async () => { await new Promise(resolve => { setTimeout(resolve, 10000).unref(); });});
        it('A.1', async () => { await new Promise(resolve => { setTimeout(resolve, 10000).unref(); });});
        it('A.2', async () => { await new Promise(resolve => { setTimeout(resolve, 10000).unref(); });});
      });
   \u0020`,
        `
      jest.setTimeout(1000);
      window.setTimeout(60000);
   \u0020`,
        'window.setTimeout(60000);',
        'setTimeout(1000);',
        `
      jest.setTimeout(1000);
      test('test case', () => {
        setTimeout(() => {
          Promise.resolv();
        }, 5000);
      });
   \u0020`,
        `
      test('test case', () => {
        setTimeout(() => {
          Promise.resolv();
        }, 5000);
      });
   \u0020`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `run 0 valid ${index}`,
      code,
    })),
    ...(
      [
        `jest.setTimeout(1000);
describe('A', () => {
` + testSetup,
        `jest.setTimeout(1000);
window.setTimeout(6000)
describe('A', () => {
  beforeEach(async () => { await new Promise(resolve => { setTimeout(resolve, 10000).unref(); });});
  it('test foo', async () => { await new Promise(resolve => { setTimeout(resolve, 10000).unref(); });});
});`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream no-confusing-set-timeout / valid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2020,
      },
    })),
    {
      name: 'upstream no-confusing-set-timeout / valid 2',
      code:
        `import { handler } from 'dep/mod';
jest.setTimeout(800);
describe('A', () => {
` + testSetup,
      languageOptions: {
        sourceType: 'module',
        ecmaVersion: 2020,
      },
    },
    ...(
      [
        `function handler() {}
jest.setTimeout(800);
describe('A', () => {
` + testSetup,
        `const { handler } = require('dep/mod');
jest.setTimeout(800);
describe('A', () => {
` + testSetup,
        `jest.setTimeout(1000);
window.setTimeout(60000);`,
        'window.setTimeout(60000);',
        'setTimeout(1000);',
        `jest.setTimeout(1000);
test('test case', () => {
  setTimeout(() => {
    Promise.resolv();
  }, 5000);
});`,
        `test('test case', () => {
  setTimeout(() => {
    Promise.resolv();
  }, 5000);
});`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream no-confusing-set-timeout / valid ${index + 3}`,
      code,
      languageOptions: {
        ecmaVersion: 2020,
      },
    })),
  ],
  invalid: [
    {
      name: 'run 0 invalid 0',
      code:
        `
        jest.setTimeout(1000);
        setTimeout(1000);
        window.setTimeout(1000);
        describe('A', () => {
` +
        testSetup2 +
        `  jest.setTimeout(800);
     \u0020`,
      errors: [
        {
          messageId: 'orderSetTimeout',
          message:
            '`jest.setTimeout` should be placed before any other jest methods',
          line: 10,
          column: 9,
          endLine: 10,
          endColumn: 29,
        },
        {
          messageId: 'multipleSetTimeouts',
          message:
            'Do not call `jest.setTimeout` multiple times, as only the last call will have an effect',
          line: 10,
          column: 9,
          endLine: 10,
          endColumn: 29,
        },
      ],
    },
    ...(
      [
        [
          `
        describe('A', () => {
          jest.setTimeout(800);
` + testSetup2,
          3,
          11,
          31,
          3,
          11,
          31,
        ],
        [
          `
        describe('B', () => {
          it('B.1', async () => {
            await new Promise((resolve) => {
              jest.setTimeout(1000);
              setTimeout(resolve, 10000).unref();
            });
          });
          it('B.2', async () => {
            await new Promise((resolve) => { setTimeout(resolve, 10000).unref(); });
          });
        });
     \u0020`,
          5,
          15,
          36,
          5,
          15,
          36,
        ],
        [
          `
        test('test-suite', () => {
          jest.setTimeout(1000);
        });
     \u0020`,
          3,
          11,
          32,
          3,
          11,
          32,
        ],
      ] as const
    ).map(
      (
        [code, line1, column1, endColumn1, line2, column2, endColumn2],
        index,
      ): TestCase => ({
        name: `run 0 invalid ${index + 1}`,
        code,
        errors: [
          {
            messageId: 'globalSetTimeout',
            message: '`jest.setTimeout` should be call in `global` scope',
            line: line1,
            column: column1,
            endLine: line1,
            endColumn: endColumn1,
          },
          {
            messageId: 'orderSetTimeout',
            message:
              '`jest.setTimeout` should be placed before any other jest methods',
            line: line2,
            column: column2,
            endLine: line2,
            endColumn: endColumn2,
          },
        ],
      }),
    ),
    {
      name: 'run 0 invalid 4',
      code:
        `
        describe('A', () => {
` +
        testSetup2 +
        `  jest.setTimeout(1000);
     \u0020`,
      errors: [
        {
          messageId: 'orderSetTimeout',
          message:
            '`jest.setTimeout` should be placed before any other jest methods',
          line: 7,
          column: 9,
          endLine: 7,
          endColumn: 30,
        },
      ],
    },
    {
      name: 'run 0 invalid 5',
      code:
        `
        import { jest } from '@jest/globals';
        {
          jest.setTimeout(800);
        }
        describe('A', () => {
` + testSetup2,
      errors: [
        {
          messageId: 'globalSetTimeout',
          message: '`jest.setTimeout` should be call in `global` scope',
          line: 4,
          column: 11,
          endLine: 4,
          endColumn: 31,
        },
      ],
    },
    {
      name: 'run 0 invalid 6',
      code: `
        jest.setTimeout(800);
        jest.setTimeout(900);
     \u0020`,
      errors: [
        {
          messageId: 'multipleSetTimeouts',
          message:
            'Do not call `jest.setTimeout` multiple times, as only the last call will have an effect',
          line: 3,
          column: 9,
          endLine: 3,
          endColumn: 29,
        },
      ],
    },
    {
      name: 'run 0 invalid 7',
      code: `
        expect(1 + 2).toEqual(3);
        jest.setTimeout(800);
     \u0020`,
      errors: [
        {
          messageId: 'orderSetTimeout',
          message:
            '`jest.setTimeout` should be placed before any other jest methods',
          line: 3,
          column: 9,
          endLine: 3,
          endColumn: 29,
        },
      ],
    },
    {
      name: 'run 0 invalid 8',
      code: `
        import { jest as Jest } from '@jest/globals';
        {
          Jest.setTimeout(800);
        }
     \u0020`,
      errors: [
        {
          messageId: 'globalSetTimeout',
          message: '`jest.setTimeout` should be call in `global` scope',
          line: 4,
          column: 11,
          endLine: 4,
          endColumn: 31,
        },
      ],
    },
    {
      name: 'upstream no-confusing-set-timeout / invalid 0',
      code:
        `jest.setTimeout(1000);
setTimeout(1000);
window.setTimeout(1000);
describe('A', () => {
` +
        testSetup +
        `
jest.setTimeout(800);`,
      languageOptions: {
        ecmaVersion: 2020,
      },
      errors: [
        {
          messageId: 'orderSetTimeout',
          message:
            '`jest.setTimeout` should be placed before any other jest methods',
          line: 9,
          column: 1,
          endLine: 9,
          endColumn: 21,
        },
        {
          messageId: 'multipleSetTimeouts',
          message:
            'Do not call `jest.setTimeout` multiple times, as only the last call will have an effect',
          line: 9,
          column: 1,
          endLine: 9,
          endColumn: 21,
        },
      ],
    },
    ...(
      [
        [
          `describe('A', () => {
  jest.setTimeout(800);
` + testSetup,
          2,
          3,
          23,
          2,
          3,
          23,
        ],
        [
          `describe('B', () => {
  it('B.1', async () => {
    await new Promise((resolve) => {
      jest.setTimeout(1000);
      setTimeout(resolve, 10000).unref();
    });
  });
  it('B.2', async () => {
    await new Promise((resolve) => { setTimeout(resolve, 10000).unref(); });
  });
});`,
          4,
          7,
          28,
          4,
          7,
          28,
        ],
        [
          `test('test-suite', () => {
  jest.setTimeout(1000);
});`,
          2,
          3,
          24,
          2,
          3,
          24,
        ],
      ] as const
    ).map(
      (
        [code, line1, column1, endColumn1, line2, column2, endColumn2],
        index,
      ): TestCase => ({
        name: `upstream no-confusing-set-timeout / invalid ${index + 1}`,
        code,
        languageOptions: {
          ecmaVersion: 2020,
        },
        errors: [
          {
            messageId: 'globalSetTimeout',
            message: '`jest.setTimeout` should be call in `global` scope',
            line: line1,
            column: column1,
            endLine: line1,
            endColumn: endColumn1,
          },
          {
            messageId: 'orderSetTimeout',
            message:
              '`jest.setTimeout` should be placed before any other jest methods',
            line: line2,
            column: column2,
            endLine: line2,
            endColumn: endColumn2,
          },
        ],
      }),
    ),
    {
      name: 'upstream no-confusing-set-timeout / invalid 4',
      code:
        `describe('A', () => {
` +
        testSetup +
        `
jest.setTimeout(1000);`,
      languageOptions: {
        ecmaVersion: 2020,
      },
      errors: [
        {
          messageId: 'orderSetTimeout',
          message:
            '`jest.setTimeout` should be placed before any other jest methods',
          line: 6,
          column: 1,
          endLine: 6,
          endColumn: 22,
        },
      ],
    },
    {
      name: 'upstream no-confusing-set-timeout / invalid 5',
      code:
        `import { jest } from '@jest/globals';
{
  jest.setTimeout(800);
}
describe('A', () => {
` + testSetup,
      languageOptions: {
        sourceType: 'module',
        ecmaVersion: 2020,
      },
      errors: [
        {
          messageId: 'globalSetTimeout',
          message: '`jest.setTimeout` should be call in `global` scope',
          line: 3,
          column: 3,
          endLine: 3,
          endColumn: 23,
        },
      ],
    },
    {
      name: 'upstream no-confusing-set-timeout / invalid 6',
      code: `jest.setTimeout(800);
jest.setTimeout(900);`,
      languageOptions: {
        ecmaVersion: 2020,
      },
      errors: [
        {
          messageId: 'multipleSetTimeouts',
          message:
            'Do not call `jest.setTimeout` multiple times, as only the last call will have an effect',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 21,
        },
      ],
    },
    {
      name: 'upstream no-confusing-set-timeout / invalid 7',
      code: `expect(1 + 2).toEqual(3);
jest.setTimeout(800);`,
      languageOptions: {
        ecmaVersion: 2020,
      },
      errors: [
        {
          messageId: 'orderSetTimeout',
          message:
            '`jest.setTimeout` should be placed before any other jest methods',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 21,
        },
      ],
    },
    {
      name: 'upstream no-confusing-set-timeout / invalid 8',
      code: `import { jest as Jest } from '@jest/globals';
{
  Jest.setTimeout(800);
}`,
      languageOptions: {
        sourceType: 'module',
        ecmaVersion: 2020,
      },
      errors: [
        {
          messageId: 'globalSetTimeout',
          message: '`jest.setTimeout` should be call in `global` scope',
          line: 3,
          column: 3,
          endLine: 3,
          endColumn: 23,
        },
      ],
    },
  ],
});
