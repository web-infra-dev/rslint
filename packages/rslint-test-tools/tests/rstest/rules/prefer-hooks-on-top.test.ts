import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('prefer-hooks-on-top', {} as never, {
  valid: [
    {
      code: `
        describe('checkout', () => {
          beforeEach(() => {});
          someSetupFn();
          afterEach(() => {});

          test('charges the card', () => {
            someFn();
          });
        });
      `,
    },
    {
      code: `
        describe('checkout', () => {
          beforeEach(() => {});
          test('charges the card', () => {});

          describe('with a declined card', () => {
            beforeEach(() => {});
            test('reports the decline', () => {});
          });
        });
      `,
    },
    {
      // Building an extended test API registers nothing, so the hooks that
      // follow it are still on top.
      code: `
        import { test as baseTest } from '@rstest/core';

        const test = baseTest.extend({});

        beforeEach(() => {});
        afterEach(() => {});
      `,
    },
  ],
  invalid: [
    {
      code: `
        describe('checkout', () => {
          beforeEach(() => {});
          test('charges the card', () => {});

          beforeAll(() => {});
          test('reports the decline', () => {});
        });
      `,
      errors: [
        {
          messageId: 'noHookOnTop',
          line: 6,
          column: 11,
        },
      ],
    },
    {
      code: `
        describe('checkout', () => {
          test('charges the card', () => {});

          beforeEach(() => {});
          afterAll(() => {});
        });
      `,
      errors: [
        {
          messageId: 'noHookOnTop',
          line: 5,
          column: 11,
        },
        {
          messageId: 'noHookOnTop',
          line: 6,
          column: 11,
        },
      ],
    },
  ],
});
