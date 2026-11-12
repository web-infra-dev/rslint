import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

const ruleTester = new RuleTester(config);

ruleTester.run('expect-expect', {
  valid: [
    {
      code: `
        import { expect, test } from '@rstest/core';
        test('has assertion', () => {
          expect(1).toBe(1);
        });
      `,
    },
    {
      code: `
        test.todo('later');
      `,
    },
  ],
  invalid: [
    {
      code: `
        import { test } from '@rstest/core';
        test('no assertion', () => {
          doSomething();
        });
      `,
      errors: [
        {
          messageId: 'noAssertions',
          message: 'Test has no assertions',
          line: 3,
          column: 9,
          endLine: 3,
          endColumn: 13,
        },
      ],
    },
  ],
});
