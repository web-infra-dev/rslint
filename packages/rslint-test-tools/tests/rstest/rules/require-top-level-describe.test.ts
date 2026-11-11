import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

const ruleTester = new RuleTester(config);

ruleTester.run('require-top-level-describe', {
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
      errors: [
        {
          messageId: 'unexpectedHook',
          message: 'All hooks must be wrapped in a describe block',
          line: 4,
          column: 9,
          endLine: 4,
          endColumn: 38,
        },
      ],
    },
    {
      code: `
        test.concurrent('places the order', () => {});
      `,
      errors: [
        {
          messageId: 'unexpectedTestCase',
          message: 'All test cases must be wrapped in a describe block',
          line: 2,
          column: 9,
          endLine: 2,
          endColumn: 54,
        },
      ],
    },
    {
      code: `
        describe('checkout', () => {});
        describe('returns', () => {});
      `,
      options: [{ maxNumberOfTopLevelDescribes: 1 }],
      errors: [
        {
          messageId: 'tooManyDescribes',
          message: 'There should not be more than 1 describe at the top level',
          line: 3,
          column: 9,
          endLine: 3,
          endColumn: 38,
        },
      ],
    },
  ],
});
