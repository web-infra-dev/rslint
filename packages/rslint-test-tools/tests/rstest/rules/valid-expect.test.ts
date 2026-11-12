import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

const ruleTester = new RuleTester(config);

ruleTester.run('valid-expect', {
  valid: [
    {
      code: `
        import { expect } from '@rstest/core';
        expect(value).toBe(1);
      `,
    },
    {
      code: `
        import { expect } from '@rstest/core';
        expect(value, 'message').toBe(1);
      `,
    },
    {
      code: `
        import { expect, test } from '@rstest/core';
        test('t', async () => {
          await expect(promise).resolves.toBe(1);
        });
      `,
    },
    {
      code: `
        import { expect, test } from '@rstest/core';
        test('t', async () => {
          await expect(promise).resolves.to.be.true;
        });
      `,
    },
    {
      code: `
        import { expect, test } from '@rstest/core';
        test('t', () => {
          return expect(promise).resolves.to.be.true;
        });
      `,
    },
    {
      code: `
        import { expect, test } from '@rstest/core';
        test('t', () => expect(promise).resolves.to.be.true);
      `,
    },
    {
      code: `
        import { expect, test } from '@rstest/core';
        test('t', async () => {
          await expect(promise).resolves.to.be.a('string').that.contains('x');
        });
      `,
    },
    {
      code: `
        import { expect, test } from '@rstest/core';
        test('t', async () => {
          await Promise.all(([expect(p).resolves.toBe(1), expect(q).resolves.toBe(2)]));
        });
      `,
    },
  ],
  invalid: [
    {
      code: `
        import { expect } from '@rstest/core';
        expect(value);
      `,
      errors: [
        {
          messageId: 'matcherNotFound',
          message: 'Expect must have a corresponding matcher call',
          line: 3,
          column: 9,
          endLine: 3,
          endColumn: 22,
        },
      ],
    },
    {
      code: `
        import { expect, test } from '@rstest/core';
        test('t', async () => {
          expect(promise).resolves.toBe(1);
        });
      `,
      errors: [
        {
          messageId: 'asyncMustBeAwaited',
          message: 'Async assertions must be awaited or returned',
          line: 4,
          column: 11,
          endLine: 4,
          endColumn: 43,
        },
      ],
      output:
        "\n        import { expect, test } from '@rstest/core';\n        test('t', async () => {\n          await expect(promise).resolves.toBe(1);\n        });\n      ",
    },
    {
      code: `
        import { expect, test } from '@rstest/core';
        test('t', async () => {
          expect(promise).resolves.to.be.true;
        });
      `,
      errors: [
        {
          messageId: 'asyncMustBeAwaited',
          message: 'Async assertions must be awaited or returned',
          line: 4,
          column: 11,
          endLine: 4,
          endColumn: 46,
        },
      ],
      output: `
        import { expect, test } from '@rstest/core';
        test('t', async () => {
          await expect(promise).resolves.to.be.true;
        });
      `,
    },
    {
      code: `
        import { expect, test } from '@rstest/core';
        test('t', () => {
          Promise.all(([expect(p).resolves.toBe(1), expect(q).resolves.toBe(2)]));
        });
      `,
      errors: [
        {
          messageId: 'promisesWithAsyncAssertionsMustBeAwaited',
          message:
            'Promises which return async assertions must be awaited or returned',
          line: 4,
          column: 11,
          endLine: 4,
          endColumn: 82,
        },
      ],
      output: `
        import { expect, test } from '@rstest/core';
        test('t', async () => {
          await Promise.all(([expect(p).resolves.toBe(1), expect(q).resolves.toBe(2)]));
        });
      `,
    },
  ],
});
