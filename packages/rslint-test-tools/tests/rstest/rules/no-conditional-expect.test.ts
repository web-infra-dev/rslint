import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-conditional-expect', {} as never, {
  valid: [
    {
      code: `
        test('case', () => {
          if (condition) setup();
          expect(value).toBe(1);
        });
      `,
    },
    {
      code: `
        describe('suite', () => {
          if (condition) expect(value).toBe(1);
        });
      `,
    },
  ],
  invalid: [
    {
      code: `
        import { expect, test } from '@rstest/core';
        test('case', () => {
          if (condition) expect(value).toBe(1);
        });
      `,
      errors: [{ messageId: 'conditionalExpect', line: 4, column: 26 }],
    },
    {
      code: `
        test('case', ({ expect }) => {
          condition && expect(value).toBe(1);
        });
      `,
      errors: [{ messageId: 'conditionalExpect', line: 3, column: 24 }],
    },
    {
      code: `
        if (import.meta.rstest) {
          const { test, expect } = import.meta.rstest;
          test('case', () => {
            condition && expect(value).toBe(1);
          });
        }
      `,
      errors: [{ messageId: 'conditionalExpect', line: 5, column: 26 }],
    },
    {
      code: `
        test('browser case', async () => {
          if (condition) {
            await expect.element(locator).toBeVisible();
          }
        });
      `,
      errors: [{ messageId: 'conditionalExpect', line: 4, column: 19 }],
    },
    {
      code: `
        import { expect, test } from '@rstest/playwright';
        test.fail('playwright case', async ({ page }) => {
          if (condition) {
            await expect.soft(page).toHaveTitle('Dashboard');
          }
        });
      `,
      errors: [{ messageId: 'conditionalExpect', line: 5, column: 19 }],
    },
  ],
});
