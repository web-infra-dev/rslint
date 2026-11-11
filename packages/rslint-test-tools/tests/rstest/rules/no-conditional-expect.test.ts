import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

const ruleTester = new RuleTester(config);

ruleTester.run('no-conditional-expect', {
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
      errors: [
        {
          messageId: 'conditionalExpect',
          message: 'Avoid calling `expect` conditionally',
          line: 4,
          column: 26,
          endLine: 4,
          endColumn: 47,
        },
      ],
    },
    {
      code: `
        test('case', ({ expect }) => {
          condition && expect(value).toBe(1);
        });
      `,
      errors: [
        {
          messageId: 'conditionalExpect',
          message: 'Avoid calling `expect` conditionally',
          line: 3,
          column: 24,
          endLine: 3,
          endColumn: 45,
        },
      ],
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
      errors: [
        {
          messageId: 'conditionalExpect',
          message: 'Avoid calling `expect` conditionally',
          line: 5,
          column: 26,
          endLine: 5,
          endColumn: 47,
        },
      ],
    },
    {
      code: `
        test('browser case', async () => {
          if (condition) {
            await expect.element(locator).toBeVisible();
          }
        });
      `,
      errors: [
        {
          messageId: 'conditionalExpect',
          message: 'Avoid calling `expect` conditionally',
          line: 4,
          column: 19,
          endLine: 4,
          endColumn: 56,
        },
      ],
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
      errors: [
        {
          messageId: 'conditionalExpect',
          message: 'Avoid calling `expect` conditionally',
          line: 5,
          column: 19,
          endLine: 5,
          endColumn: 61,
        },
      ],
    },
  ],
});
