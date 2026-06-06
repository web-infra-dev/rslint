import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('prefer-expect-resolves', {} as never, {
  valid: [
    { code: 'expect.hasAssertions()' },
    {
      code: `
      it('passes', async () => {
        await expect(someValue()).resolves.toBe(true);
      });
    `,
    },
    {
      code: `
      it('is true', async () => {
        const myPromise = Promise.resolve(true);

        await expect(myPromise).resolves.toBe(true);
      });
    `,
    },
    {
      code: `
      it('errors', async () => {
        await expect(Promise.reject(new Error('oh noes!'))).rejects.toThrowError(
          'oh noes!',
        );
      });
    `,
    },
  ],
  invalid: [
    {
      code: `
        it('passes', async () => {
          expect(await someValue(),).toBe(true);
        });
      `,
      output: `
        it('passes', async () => {
          await expect(someValue(),).resolves.toBe(true);
        });
      `,
      errors: [{ endColumn: 27, column: 10, messageId: 'expectResolves' }],
    },
    {
      code: `
        it('is true', async () => {
          const myPromise = Promise.resolve(true);

          expect(await myPromise).toBe(true);
        });
      `,
      output: `
        it('is true', async () => {
          const myPromise = Promise.resolve(true);

          await expect(myPromise).resolves.toBe(true);
        });
      `,
      errors: [{ endColumn: 25, column: 10, messageId: 'expectResolves' }],
    },
    {
      code: `
        import { expect as pleaseExpect } from '@jest/globals';

        it('is true', async () => {
          const myPromise = Promise.resolve(true);

          pleaseExpect(await myPromise).toBe(true);
        });
      `,
      output: `
        import { expect as pleaseExpect } from '@jest/globals';

        it('is true', async () => {
          const myPromise = Promise.resolve(true);

          await pleaseExpect(myPromise).resolves.toBe(true);
        });
      `,
      errors: [{ endColumn: 31, column: 16, messageId: 'expectResolves' }],
    },
  ],
});
