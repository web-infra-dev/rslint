import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

const ruleTester = new RuleTester(config);

ruleTester.run('require-hook', {
  valid: [
    {
      code: `
        import { beforeEach, describe, expect, it } from '@rstest/core';

        const cities = ['Vienna'];

        describe('cities', () => {
          beforeEach(() => {
            initializeCityDatabase();
          });

          it('has Vienna', () => {
            expect(isCity('Vienna')).toBe(true);
          });
        });
      `,
    },
    {
      code: `
        import { rs as helpers } from '@rstest/core';

        helpers.mock('../api');
        helpers.spyOn(console, 'warn').mockReturnValue(undefined);
      `,
    },
    {
      code: `
        enableAutoDestroy(afterEach);
      `,
      options: [{ allowedFunctionCalls: ['enableAutoDestroy'] }],
    },
  ],
  invalid: [
    {
      code: `
        initializeCityDatabase();
      `,
      errors: [
        {
          messageId: 'useHook',
          message: 'This should be done within a hook',
          line: 2,
          column: 9,
          endLine: 2,
          endColumn: 34,
        },
      ],
    },
    {
      code: `
        describe('cities', () => {
          let consoleWarnSpy = rs.spyOn(console, 'warn');
        });
      `,
      errors: [
        {
          messageId: 'useHook',
          message: 'This should be done within a hook',
          line: 3,
          column: 11,
          endLine: 3,
          endColumn: 58,
        },
      ],
    },
    {
      code: `
        describe.skip('cities', () => {
          seedProducts();
        });
      `,
      errors: [
        {
          messageId: 'useHook',
          message: 'This should be done within a hook',
          line: 3,
          column: 11,
          endLine: 3,
          endColumn: 26,
        },
      ],
    },
    {
      code: `
        describe('cities', (() => {
          seedProducts();
        }));
      `,
      errors: [
        {
          messageId: 'useHook',
          message: 'This should be done within a hook',
          line: 3,
          column: 11,
          endLine: 3,
          endColumn: 26,
        },
      ],
    },
    {
      code: `
        import { onTestFinished } from '@rstest/core';

        onTestFinished(() => closeDatabase());
      `,
      errors: [
        {
          messageId: 'useTest',
          message: 'onTestFinished() can only be called inside a test',
          line: 4,
          column: 9,
          endLine: 4,
          endColumn: 47,
        },
      ],
    },
  ],
});
