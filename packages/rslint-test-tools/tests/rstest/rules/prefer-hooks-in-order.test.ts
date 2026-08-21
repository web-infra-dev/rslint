import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('prefer-hooks-in-order', {} as never, {
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
          data: { currentHook: 'beforeAll', previousHook: 'afterAll' },
          line: 2,
          column: 1,
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
          data: { currentHook: 'beforeEach', previousHook: 'afterEach' },
          line: 2,
          column: 1,
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
          data: { currentHook: 'beforeAll', previousHook: 'beforeEach' },
          line: 2,
          column: 1,
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
          data: { currentHook: 'afterEach', previousHook: 'afterAll' },
          line: 2,
          column: 1,
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
          data: { currentHook: 'afterEach', previousHook: 'afterAll' },
          line: 3,
          column: 3,
        },
        {
          messageId: 'reorderHooks',
          data: { currentHook: 'beforeAll', previousHook: 'beforeEach' },
          line: 8,
          column: 3,
        },
      ],
    },
  ],
});
