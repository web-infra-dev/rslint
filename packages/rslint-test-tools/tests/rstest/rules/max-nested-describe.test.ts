import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('max-nested-describe', {} as never, {
  valid: [
    {
      code: `
        describe('one', () => {
          describe('two', () => {
            describe('three', () => {});
          });
        });
      `,
      options: [{ max: 3 }],
    },
    {
      code: `
        describe('one', () => {
          describe('first child', () => {});
          describe('second child', () => {});
        });
      `,
      options: [{ max: 2 }],
    },
  ],
  invalid: [
    {
      code: `
        describe('one', () => {
          describe.only('two', () => {
            describe.each(['three'])('%s', () => {});
          });
        });
      `,
      options: [{ max: 2 }],
      errors: [
        {
          messageId: 'exceededMaxDepth',
          message: 'Too many nested describe calls (3) - maximum allowed is 2',
          line: 4,
          column: 13,
        },
      ],
    },
    {
      code: `
        import { test } from '@rstest/playwright';
        test.describe('one', () => {
          test.describe.for([[1]])('%s', () => {});
        });
      `,
      options: [{ max: 1 }],
      errors: [
        {
          messageId: 'exceededMaxDepth',
          message: 'Too many nested describe calls (2) - maximum allowed is 1',
          line: 4,
          column: 11,
        },
      ],
    },
  ],
});
