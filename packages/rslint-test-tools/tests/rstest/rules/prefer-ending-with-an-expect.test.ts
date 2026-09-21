import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('prefer-ending-with-an-expect', {} as never, {
  valid: [
    {
      code: `
        import { expect, test } from '@rstest/core';

        test('places the order', () => {
          checkout();
          expect(cart.state).toBe('ordered');
        });
      `,
    },
    {
      code: `
        test('places the order', { timeout: 100 }, context => {
          context.expect(checkout()).toBe('ok');
        });
      `,
    },
    {
      code: `
        test.todo('supports gift cards', () => {
          checkout();
        });
      `,
    },
  ],
  invalid: [
    {
      code: `
        import { expect, test } from '@rstest/core';

        test('places the order', () => {
          expect(cart.state).toBe('open');
          checkout();
        });
      `,
      errors: [{ messageId: 'mustEndWithExpect', line: 4, column: 9 }],
    },
    {
      code: `
        test('places the order', { timeout: 100 }, () => {
          checkout();
        });
      `,
      errors: [{ messageId: 'mustEndWithExpect', line: 2, column: 9 }],
    },
    {
      code: `
        test('places the order', () => {
          checkForSaga(cart).run();
        });
      `,
      options: [{ assertFunctionNames: ['expect', 'expectSaga'] }],
      errors: [{ messageId: 'mustEndWithExpect', line: 2, column: 9 }],
    },
  ],
});
