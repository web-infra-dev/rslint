import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

const ruleTester = new RuleTester(config);

ruleTester.run('prefer-expect-resolves', {
  valid: [
    { code: 'expect.hasAssertions()' },
    { code: 'await expect().resolves.toBe(true)' },
    {
      code: `
      it('passes', async () => {
        await expect(someValue()).resolves.toBe(true);
      });
    `,
    },
    { code: 'expect().nothing()' },
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
      errors: [
        {
          messageId: 'expectResolves',
          message: 'Use `expect().resolves` instead',
          line: 3,
          column: 18,
          endLine: 3,
          endColumn: 35,
          suggestions: [
            {
              messageId: 'suggestExpectResolves',
              desc: 'Replace with an awaited resolves assertion',
              output:
                "\n        it('passes', async () => {\n          await expect(someValue(),).resolves.toBe(true);\n        });\n      ",
            },
          ],
        },
      ],
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
      errors: [
        {
          messageId: 'expectResolves',
          message: 'Use `expect().resolves` instead',
          line: 5,
          column: 18,
          endLine: 5,
          endColumn: 33,
        },
      ],
    },
    {
      code: `
        import { expect as pleaseExpect } from '@rstest/core';

        it('is true', async () => {
          const myPromise = Promise.resolve(true);

          pleaseExpect(await myPromise).toBe(true);
        });
      `,
      output: `
        import { expect as pleaseExpect } from '@rstest/core';

        it('is true', async () => {
          const myPromise = Promise.resolve(true);

          await pleaseExpect(myPromise).resolves.toBe(true);
        });
      `,
      errors: [
        {
          messageId: 'expectResolves',
          message: 'Use `expect().resolves` instead',
          line: 7,
          column: 24,
          endLine: 7,
          endColumn: 39,
        },
      ],
    },
  ],
});
