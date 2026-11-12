import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

const ruleTester = new RuleTester(config);

ruleTester.run('prefer-hooks-in-order', {
  valid: [
    {
      code: `
        beforeAll(() => {});
        beforeEach(() => {});
        afterEach(() => {});
        afterAll(() => {});
      `,
    },
    {
      code: `
        afterAll(() => {});
        doSomething();
        beforeAll(() => {});
      `,
    },
    {
      // Hooks inside a hook callback form their own run and are not compared
      // against the hook that encloses them.
      code: `
        afterAll(() => {
          beforeEach(() => {});
          afterEach(() => {});
        });
      `,
    },
  ],
  invalid: [
    {
      code: `
        afterAll(() => {});
        beforeAll(() => {});
      `,
      errors: [
        {
          messageId: 'reorderHooks',
          message: '`beforeAll` hooks should be before any `afterAll` hooks',
          line: 3,
          column: 9,
          endLine: 3,
          endColumn: 28,
        },
      ],
    },
    {
      code: `
        afterEach(() => {});
        beforeEach(() => {});
      `,
      errors: [
        {
          messageId: 'reorderHooks',
          message: '`beforeEach` hooks should be before any `afterEach` hooks',
          line: 3,
          column: 9,
          endLine: 3,
          endColumn: 29,
        },
      ],
    },
    {
      code: `
        beforeEach(() => {});
        beforeAll(() => {});
      `,
      errors: [
        {
          messageId: 'reorderHooks',
          message: '`beforeAll` hooks should be before any `beforeEach` hooks',
          line: 3,
          column: 9,
          endLine: 3,
          endColumn: 28,
        },
      ],
    },
    {
      code: `
        afterAll(() => {});
        afterEach(() => {});
      `,
      errors: [
        {
          messageId: 'reorderHooks',
          message: '`afterEach` hooks should be before any `afterAll` hooks',
          line: 3,
          column: 9,
          endLine: 3,
          endColumn: 28,
        },
      ],
    },
    {
      code: `
        describe('my test', () => {
          afterAll(() => {});
          afterEach(() => {});

          doSomething();

          beforeEach(() => {});
          beforeAll(() => {});
        });
      `,
      errors: [
        {
          messageId: 'reorderHooks',
          message: '`afterEach` hooks should be before any `afterAll` hooks',
          line: 4,
          column: 11,
          endLine: 4,
          endColumn: 30,
        },
        {
          messageId: 'reorderHooks',
          message: '`beforeAll` hooks should be before any `beforeEach` hooks',
          line: 9,
          column: 11,
          endLine: 9,
          endColumn: 30,
        },
      ],
    },
    {
      // The surrounding run survives a hook callback, and keeps its own
      // previous hook for the report.
      code: `
        afterEach(() => {
          beforeEach(() => {});
          afterAll(() => {});
        });
        beforeAll(() => {});
      `,
      errors: [
        {
          messageId: 'reorderHooks',
          message: '`beforeAll` hooks should be before any `afterEach` hooks',
          line: 6,
          column: 9,
          endLine: 6,
          endColumn: 28,
        },
      ],
    },
  ],
});
