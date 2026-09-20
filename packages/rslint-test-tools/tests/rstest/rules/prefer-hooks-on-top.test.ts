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
    {
      // A function body is a scope of its own: a hook in a function that is
      // never called registers nothing.
      code: `
        describe('checkout', () => {
          test('charges the card', () => {});

          function unused() {
            beforeEach(() => {});
          }
        });
      `,
    },
    {
      // A function passed to describe by name is that suite's body, so the
      // cases registered above it in the file do not count against it.
      code: `
        test('refunds the card', () => {});

        function checkout() {
          beforeEach(() => {});
          test('charges the card', () => {});
        }

        describe('checkout', checkout);
      `,
    },
  ],
  invalid: [
    {
      // A named suite body is still judged on its own order.
      code: `
        function checkout() {
          test('charges the card', () => {});

          beforeEach(() => {});
        }

        describe('checkout', checkout);
      `,
      errors: [
        {
          messageId: 'noHookOnTop',
          line: 5,
          column: 11,
        },
      ],
    },
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
