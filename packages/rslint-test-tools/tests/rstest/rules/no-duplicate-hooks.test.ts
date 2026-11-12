import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

const ruleTester = new RuleTester(config);

ruleTester.run('no-duplicate-hooks', {
  valid: [
    {
      code: `
        describe('cart', () => {
          beforeEach(() => resetCart());
          afterEach(() => clearCart());

          describe('discounts', () => {
            beforeEach(() => seedDiscounts());
          });
        });
      `,
    },
    {
      code: `
        describe.each(['member', 'guest'])('%s cart', () => {
          beforeEach(() => resetCart());
        });
      `,
    },
  ],
  invalid: [
    {
      code: `
        describe('cart', () => {
          beforeEach(() => resetCart());
          beforeEach(() => seedProducts());
        });
      `,
      errors: [
        {
          messageId: 'noDuplicateHook',
          message: 'Duplicate beforeEach in describe block',
          line: 4,
          column: 11,
          endLine: 4,
          endColumn: 43,
        },
      ],
    },
    {
      code: `
        import { afterEach, afterEach as cleanup } from '@rstest/core';

        afterEach(() => clearCart());
        cleanup(() => closeDatabase());
      `,
      errors: [
        {
          messageId: 'noDuplicateHook',
          message: 'Duplicate afterEach in describe block',
          line: 5,
          column: 9,
          endLine: 5,
          endColumn: 39,
        },
      ],
    },
    {
      code: `
        describe.for([{ currency: 'USD' }])('$currency cart', () => {
          beforeAll(() => connectDatabase());
          beforeAll(() => seedProducts());
        });
      `,
      errors: [
        {
          messageId: 'noDuplicateHook',
          message: 'Duplicate beforeAll in describe block',
          line: 4,
          column: 11,
          endLine: 4,
          endColumn: 42,
        },
      ],
    },
  ],
});
