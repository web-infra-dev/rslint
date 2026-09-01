import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('require-awaited-expect-poll', {} as never, {
  valid: [
    {
      code: `
        test('should pass', async () => {
          await expect.poll(() => element).toBeInTheDocument();
        });
      `,
    },
    {
      code: `
        test('should pass', async () => {
          await expect.element(element).toBeInTheDocument();
        });
      `,
    },
    {
      code: `
        test('should pass', () => {
          expect.syncElement(element).toBeInTheDocument();
        });
      `,
    },
    {
      code: `
        test('should pass', () => {
          return expect.poll(() => element).toBeInTheDocument();
        });
      `,
    },
    {
      code: `
        test('should pass', () => {
          return expect.element(element).toBeInTheDocument();
        });
      `,
    },
    {
      code: `
        test('should pass', () => {
          return expect(true).toBe(true);
        });
      `,
    },
    {
      code: `
        test('should pass', async () => {
          (sideEffect(), await expect.poll(() => element).toBeInTheDocument());
        });
      `,
    },
    {
      code: `
        test('should pass', async () => {
          await (sideEffect(), expect.poll(() => element).toBeInTheDocument());
        });
      `,
    },
    {
      code: `
        test('should pass', async () => {
          await (sideEffect(), (sideEffect(), (sideEffect(), expect.poll(() => element).toBeInTheDocument())));
        });
      `,
    },
    {
      code: `
        test('should pass', () => {
          return (sideEffect(), expect.poll(() => element).toBeInTheDocument());
        });
      `,
    },
  ],
  invalid: [
    {
      code: `
        test('should fail', () => {
          expect.poll(() => element).toBeInTheDocument();
        });
      `,
      errors: [
        {
          messageId: 'notAwaited',
          message: '`expect.poll` calls should be awaited',
          line: 3,
          column: 11,
          endLine: 3,
          endColumn: 22,
        },
      ],
    },
    {
      code: `
        test('should fail', () => {
          expect.element(element).toBeInTheDocument();
        });
      `,
      errors: [
        {
          messageId: 'notAwaited',
          message: '`expect.element` calls should be awaited',
          line: 3,
          column: 11,
          endLine: 3,
          endColumn: 25,
        },
      ],
    },
    {
      code: `
        test('should fail', () => {
          expect['poll'](() => element).toBeInTheDocument();
        });
      `,
      errors: [
        {
          messageId: 'notAwaited',
          message: '`expect.poll` calls should be awaited',
          line: 3,
          column: 11,
          endLine: 3,
          endColumn: 25,
        },
      ],
    },
    {
      code: `
        test('should fail', () => {
          expect['element'](element).toBeInTheDocument();
        });
      `,
      errors: [
        {
          messageId: 'notAwaited',
          message: '`expect.element` calls should be awaited',
          line: 3,
          column: 11,
          endLine: 3,
          endColumn: 28,
        },
      ],
    },
    {
      code: `
        test('should fail', () => {
          (expect.poll(() => element).toBeInTheDocument(), expect(true).toBe(true));
        });
      `,
      errors: [
        {
          messageId: 'notAwaited',
          message: '`expect.poll` calls should be awaited',
          line: 3,
          column: 12,
          endLine: 3,
          endColumn: 23,
        },
      ],
    },
    {
      code: `
        test('should fail', () => {
          (expect.element(() => element).toBeInTheDocument(), expect(true).toBe(true));
        });
      `,
      errors: [
        {
          messageId: 'notAwaited',
          message: '`expect.element` calls should be awaited',
          line: 3,
          column: 12,
          endLine: 3,
          endColumn: 26,
        },
      ],
    },
    {
      code: `
        test('should fail', () => {
          return (expect.poll(() => element).toBeInTheDocument(), expect(true).toBe(true));
        });
      `,
      errors: [
        {
          messageId: 'notAwaited',
          message: '`expect.poll` calls should be awaited',
          line: 3,
          column: 19,
          endLine: 3,
          endColumn: 30,
        },
      ],
    },
  ],
});
