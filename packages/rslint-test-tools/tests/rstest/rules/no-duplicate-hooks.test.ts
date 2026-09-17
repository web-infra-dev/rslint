import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-duplicate-hooks', {} as never, {
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
          data: { hook: 'beforeEach' },
          line: 3,
          column: 3,
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
          data: { hook: 'afterEach' },
          line: 4,
          column: 1,
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
          data: { hook: 'beforeAll' },
          line: 3,
          column: 3,
        },
      ],
    },
  ],
});
