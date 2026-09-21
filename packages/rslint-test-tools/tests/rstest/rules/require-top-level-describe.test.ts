import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('require-top-level-describe', {} as never, {
  valid: [
    {
      code: `
        import { beforeEach, describe, test } from '@rstest/core';

        describe('checkout', () => {
          beforeEach(() => resetCart());
          test('places the order', () => {});
        });
      `,
    },
    {
      code: `
        describe.each(['member', 'guest'])('%s cart', () => {
          it('places the order', () => {});
        });
      `,
    },
    {
      code: `
        function cases() {
          it('places the order', () => {});
        }

        describe('checkout', cases);
      `,
    },
  ],
  invalid: [
    {
      code: `
        import { beforeEach } from '@rstest/core';

        beforeEach(() => resetCart());
      `,
      errors: [{ messageId: 'unexpectedHook', line: 4, column: 9 }],
    },
    {
      code: `
        test.concurrent('places the order', () => {});
      `,
      errors: [{ messageId: 'unexpectedTestCase', line: 2, column: 9 }],
    },
    {
      code: `
        describe('checkout', () => {});
        describe('returns', () => {});
      `,
      options: [{ maxNumberOfTopLevelDescribes: 1 }],
      errors: [{ messageId: 'tooManyDescribes', line: 3, column: 9 }],
    },
  ],
});
