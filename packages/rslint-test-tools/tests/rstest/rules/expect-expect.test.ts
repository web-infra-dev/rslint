import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('expect-expect', {} as never, {
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
      errors: [{ messageId: 'noAssertions', line: 3, column: 9 }],
    },
  ],
});
