import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('valid-expect-in-promise', {} as never, {
  valid: [
    {
      code: `test('case', async ({ expect }) => {
        await load().then(value => expect(value).toBe(1));
      });`,
    },
    {
      code: `test('case', ({ expect }) =>
        load().then(value => expect(value).toBe(1))
      );`,
    },
    {
      code: `test('case', () => {
        try {
          return load().then(value => expect(value).toBe(1));
        } catch (error) {}
      });`,
    },
    {
      code: `test('case', async () => {
        try {
          return load().then(value => expect(value).toBe(1));
        } catch (error) {}
      });`,
    },
    {
      code: `test('case', async () => {
        const pending = load().then(value => assert.equal(value, 1));
        await pending;
      });`,
    },
    {
      code: `test('case', async () => {
        const pending = load().then(value => assert.equal(value, 1));
        if (!ready) throw new Error('no');
        await pending;
      });`,
    },
    {
      code: `test('case', async () => {
        const pending = load().then(value => assert.equal(value, 1));
        try {
          setup();
          await pending;
        } catch (error) {
          throw error;
        }
      });`,
    },
    {
      code: `test('case', () => {
        const pending = load().then(value => expect(value).toBe(1));
        expect(pending).resolves.toBeUndefined();
      });`,
    },
    {
      code: `try {
        test('case', async () => {
          const pending = load().then(value => expect(value).toBe(1));
          await pending;
        });
      } catch (error) {}`,
    },
    {
      code: `try {
        test('case', () => {
          const pending = load().then(value => expect(value).toBe(1));
          return Promise.resolve(pending);
        });
      } catch (error) {}`,
    },
    {
      code: `try {
        test('case', () => {
          const pending = load().then(value => expect(value).toBe(1));
          return Promise.all([pending]);
        });
      } catch (error) {}`,
    },
    {
      code: `test('case', () => {
        const { pending } = {
          pending: load().then(value => assert.equal(value, 1)),
        };
        return pending;
      });`,
    },
    {
      code: `import assert from 'node:assert';
      test('case', () => {
        load().then(value => assert.equal(value, 1));
      });`,
    },
    {
      code: `import { assert } from 'chai';
      test('case', () => {
        load().then(value => assert.equal(value, 1));
      });`,
    },
  ],
  invalid: [
    {
      code: `test('case', context => {
        load().then(value => context.expect(value).toBe(1));
      });`,
      errors: [
        {
          messageId: 'expectInFloatingPromise',
          message:
            'This promise should either be returned or awaited to ensure the assertions in its chain are called',
        },
      ],
    },
    {
      code: `test.for([{ value: 1 }])('case', (row, context) => {
        load().then(value => context.expect(value).toBe(row.value));
      });`,
      errors: [{ messageId: 'expectInFloatingPromise' }],
    },
    {
      code: `test.each([{ value: 1 }])('case', row => {
        load().then(value => expect(value).toBe(row.value));
      });`,
      errors: [{ messageId: 'expectInFloatingPromise' }],
    },
    {
      code: `test('case', { timeout: 1000 }, ({ expect }) => {
        load().then(value => expect(value).toBe(1));
      });`,
      errors: [{ messageId: 'expectInFloatingPromise' }],
    },
    {
      code: `import { assert, test } from '@rstest/core';
      test('case', () => {
        load().then(value => assert.equal(value, 1));
      });`,
      errors: [{ messageId: 'expectInFloatingPromise' }],
    },
    {
      code: `const { assert, test } = require('@rstest/core');
      test('case', () => {
        load().then(value => assert(value));
      });`,
      errors: [{ messageId: 'expectInFloatingPromise' }],
    },
    {
      code: `import * as rstest from '@rstest/core';
      rstest.test('case', () => {
        load().then(value => rstest.assert.equal(value, 1));
      });`,
      errors: [{ messageId: 'expectInFloatingPromise' }],
    },
    {
      code: `import.meta.rstest.test('case', () => {
        load().then(value => import.meta.rstest.assert.equal(value, 1));
      });`,
      errors: [{ messageId: 'expectInFloatingPromise' }],
    },
    {
      code: `const api = import.meta.rstest;
      api.test('case', () => {
        load().then(value => api.assert.equal(value, 1));
      });`,
      errors: [{ messageId: 'expectInFloatingPromise' }],
    },
    {
      code: `import { expect, test } from '@rstest/playwright';
      test('case', async ({ page }) => {
        page.goto('url').then(() => expect(page).toBeDefined());
      });`,
      errors: [{ messageId: 'expectInFloatingPromise' }],
    },
    {
      code: `test('case', () => {
        load()
          .then(value => expect(value).to.be.ok)
          .then(value => assert.equal(value, 1));
      });`,
      errors: 1,
    },
    {
      code: `test('case', async () => {
        const pending = load().then(value => assert.equal(value, 1));
        try {
          throw new Error('caught');
        } catch {}
      });`,
      errors: [{ messageId: 'expectInFloatingPromise' }],
    },
    {
      code: `test('case', async () => {
        const pending = load().then(value => assert.equal(value, 1));
        try {
          throw new Error('suppressed');
        } finally {
          return;
        }
      });`,
      errors: [{ messageId: 'expectInFloatingPromise' }],
    },
  ],
});
