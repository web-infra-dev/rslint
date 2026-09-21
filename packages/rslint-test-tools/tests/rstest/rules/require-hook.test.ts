import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('require-hook', {} as never, {
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
      errors: [{ messageId: 'useHook', line: 2, column: 1 }],
    },
    {
      code: `
        describe('cities', () => {
          let consoleWarnSpy = rs.spyOn(console, 'warn');
        });
      `,
      errors: [{ messageId: 'useHook', line: 3, column: 3 }],
    },
    {
      code: `
        describe.skip('cities', () => {
          seedProducts();
        });
      `,
      errors: [{ messageId: 'useHook', line: 3, column: 3 }],
    },
    {
      code: `
        describe('cities', (() => {
          seedProducts();
        }));
      `,
      errors: [{ messageId: 'useHook', line: 3, column: 3 }],
    },
    {
      code: `
        import { onTestFinished } from '@rstest/core';

        onTestFinished(() => closeDatabase());
      `,
      errors: [
        {
          messageId: 'useTest',
          data: { name: 'onTestFinished' },
          line: 4,
          column: 1,
        },
      ],
    },
  ],
});
