import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('valid-expect', {} as never, {
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
  ],
  invalid: [
    {
      code: `
        import { expect } from '@rstest/core';
        expect(value);
      `,
      errors: [{ messageId: 'matcherNotFound' }],
    },
    {
      code: `
        import { expect, test } from '@rstest/core';
        test('t', async () => {
          expect(promise).resolves.toBe(1);
        });
      `,
      errors: [{ messageId: 'asyncMustBeAwaited' }],
    },
    {
      code: `
        import { expect, test } from '@rstest/core';
        test('t', async () => {
          expect(promise).resolves.to.be.true;
        });
      `,
      errors: [{ messageId: 'asyncMustBeAwaited' }],
      output: `
        import { expect, test } from '@rstest/core';
        test('t', async () => {
          await expect(promise).resolves.to.be.true;
        });
      `,
    },
  ],
});
