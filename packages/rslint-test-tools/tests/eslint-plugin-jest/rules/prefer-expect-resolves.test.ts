// Upstream reference: eslint-plugin-jest@29.16.0; includes existing Rslint regressions.
// Additional cases from 29.16.6 retain their input and parser configuration.
// Complete assertions record current native output; documented differences remain separate.
import {
  RuleTester,
  type TestCase,
  type ValidTestCase,
} from '@rslint/rule-tester';
import { fixtureConfig } from '../config';

new RuleTester(fixtureConfig()).run('prefer-expect-resolves', {
  valid: [
    ...(
      [
        'expect.hasAssertions()',
        'await expect().resolves.toBe(true)',
        `
      it('passes', async () => {
        await expect(someValue()).resolves.toBe(true);
      });
   \u0020`,
        'expect().nothing()',
        `
      it('is true', async () => {
        const myPromise = Promise.resolve(true);

        await expect(myPromise).resolves.toBe(true);
      });
   \u0020`,
        `
      it('errors', async () => {
        await expect(Promise.reject(new Error('oh noes!'))).rejects.toThrowError(
          'oh noes!',
        );
      });
   \u0020`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `run 0 valid ${index}`,
      code,
    })),
    ...(
      [
        'expect.hasAssertions()',
        `it('passes', async () => {
  await expect().resolves.toBe(true);
});`,
        `it('passes', async () => {
  await expect(someValue()).resolves.toBe(true);
});`,
        `it('is true', async () => {
  const myPromise = Promise.resolve(true);

  await expect(myPromise).resolves.toBe(true);
});`,
        `it('errors', async () => {
  await expect(Promise.reject(new Error('oh noes!'))).rejects.toThrowError(
    'oh noes!',
  );
});`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream prefer-expect-resolves / valid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2017,
      },
    })),
  ],
  invalid: [
    ...(
      [
        [
          `
        it('passes', async () => {
          expect(await someValue(),).toBe(true);
        });
     \u0020`,
          3,
          18,
          35,
          `
        it('passes', async () => {
          await expect(someValue(),).resolves.toBe(true);
        });
     \u0020`,
        ],
        [
          `
        it('is true', async () => {
          const myPromise = Promise.resolve(true);

          expect(await myPromise).toBe(true);
        });
     \u0020`,
          5,
          18,
          33,
          `
        it('is true', async () => {
          const myPromise = Promise.resolve(true);

          await expect(myPromise).resolves.toBe(true);
        });
     \u0020`,
        ],
        [
          `
        import { expect as pleaseExpect } from '@jest/globals';

        it('is true', async () => {
          const myPromise = Promise.resolve(true);

          pleaseExpect(await myPromise).toBe(true);
        });
     \u0020`,
          7,
          24,
          39,
          `
        import { expect as pleaseExpect } from '@jest/globals';

        it('is true', async () => {
          const myPromise = Promise.resolve(true);

          await pleaseExpect(myPromise).resolves.toBe(true);
        });
     \u0020`,
        ],
      ] as const
    ).map(([code, line, column, endColumn, output], index): TestCase => ({
      name: `run 0 invalid ${index}`,
      code,
      errors: [
        {
          messageId: 'expectResolves',
          message: 'Use `await expect(...).resolves instead',
          line,
          column,
          endLine: line,
          endColumn,
        },
      ],
      output,
    })),
    ...(
      [
        [
          `it('passes', async () => {
  expect(await someValue(),).toBe(true);
});`,
          2,
          27,
          `it('passes', async () => {
  await expect(someValue(),).resolves.toBe(true);
});`,
        ],
        [
          `it('is true', async () => {
  const myPromise = Promise.resolve(true);

  expect(await myPromise).toBe(true);
});`,
          4,
          25,
          `it('is true', async () => {
  const myPromise = Promise.resolve(true);

  await expect(myPromise).resolves.toBe(true);
});`,
        ],
      ] as const
    ).map(([code, line, endColumn, output], index): TestCase => ({
      name: `upstream prefer-expect-resolves / invalid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'expectResolves',
          message: 'Use `await expect(...).resolves instead',
          line,
          column: 10,
          endLine: line,
          endColumn,
        },
      ],
      output,
    })),
    {
      name: 'upstream prefer-expect-resolves / invalid 2',
      code: `import { expect as pleaseExpect } from '@jest/globals';

it('is true', async () => {
  const myPromise = Promise.resolve(true);

  pleaseExpect(await myPromise).toBe(true);
});`,
      languageOptions: {
        sourceType: 'module',
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'expectResolves',
          message: 'Use `await expect(...).resolves instead',
          line: 6,
          column: 16,
          endLine: 6,
          endColumn: 31,
        },
      ],
      output: `import { expect as pleaseExpect } from '@jest/globals';

it('is true', async () => {
  const myPromise = Promise.resolve(true);

  await pleaseExpect(myPromise).resolves.toBe(true);
});`,
    },
  ],
});
