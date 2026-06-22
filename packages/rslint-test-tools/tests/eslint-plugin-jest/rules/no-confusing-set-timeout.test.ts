import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-confusing-set-timeout', {} as never, {
  valid: [
    {
      code: `
      jest.setTimeout(1000);
      describe('A', () => {
        beforeEach(async () => { await new Promise(resolve => { setTimeout(resolve, 10000).unref(); });});
        it('A.1', async () => { await new Promise(resolve => { setTimeout(resolve, 10000).unref(); });});
        it('A.2', async () => { await new Promise(resolve => { setTimeout(resolve, 10000).unref(); });});
      });
    `,
    },
    {
      code: `
      jest.setTimeout(1000);
      window.setTimeout(6000)
      describe('A', () => {
        beforeEach(async () => { await new Promise(resolve => { setTimeout(resolve, 10000).unref(); });});
        it('test foo', async () => { await new Promise(resolve => { setTimeout(resolve, 10000).unref(); });});
      });
    `,
    },
    {
      code: `
        import { handler } from 'dep/mod';
        jest.setTimeout(800);
        describe('A', () => {
          beforeEach(async () => { await new Promise(resolve => { setTimeout(resolve, 10000).unref(); });});
          it('A.1', async () => { await new Promise(resolve => { setTimeout(resolve, 10000).unref(); });});
          it('A.2', async () => { await new Promise(resolve => { setTimeout(resolve, 10000).unref(); });});
        });
      `,
    },
    {
      code: `
      function handler() {}
      jest.setTimeout(800);
      describe('A', () => {
        beforeEach(async () => { await new Promise(resolve => { setTimeout(resolve, 10000).unref(); });});
        it('A.1', async () => { await new Promise(resolve => { setTimeout(resolve, 10000).unref(); });});
        it('A.2', async () => { await new Promise(resolve => { setTimeout(resolve, 10000).unref(); });});
      });
    `,
    },
    {
      code: `
      const { handler } = require('dep/mod');
      jest.setTimeout(800);
      describe('A', () => {
        beforeEach(async () => { await new Promise(resolve => { setTimeout(resolve, 10000).unref(); });});
        it('A.1', async () => { await new Promise(resolve => { setTimeout(resolve, 10000).unref(); });});
        it('A.2', async () => { await new Promise(resolve => { setTimeout(resolve, 10000).unref(); });});
      });
    `,
    },
    {
      code: `
      jest.setTimeout(1000);
      window.setTimeout(60000);
    `,
    },
    { code: 'window.setTimeout(60000);' },
    { code: 'setTimeout(1000);' },
    {
      code: `
      jest.setTimeout(1000);
      test('test case', () => {
        setTimeout(() => {
          Promise.resolv();
        }, 5000);
      });
    `,
    },
    {
      code: `
      test('test case', () => {
        setTimeout(() => {
          Promise.resolv();
        }, 5000);
      });
    `,
    },
  ],
  invalid: [
    {
      code: `
        jest.setTimeout(1000);
        setTimeout(1000);
        window.setTimeout(1000);
        describe('A', () => {
          beforeEach(async () => { await new Promise(resolve => { setTimeout(resolve, 10000).unref(); });});
          it('A.1', async () => { await new Promise(resolve => { setTimeout(resolve, 10000).unref(); });});
          it('A.2', async () => { await new Promise(resolve => { setTimeout(resolve, 10000).unref(); });});
        });
        jest.setTimeout(800);
      `,
      errors: [
        {
          messageId: 'orderSetTimeout',
          line: 9,
          column: 1,
        },
        {
          messageId: 'multipleSetTimeouts',
          line: 9,
          column: 1,
        },
      ],
    },
    {
      code: `
        describe('A', () => {
          jest.setTimeout(800);
          beforeEach(async () => { await new Promise(resolve => { setTimeout(resolve, 10000).unref(); });});
          it('A.1', async () => { await new Promise(resolve => { setTimeout(resolve, 10000).unref(); });});
          it('A.2', async () => { await new Promise(resolve => { setTimeout(resolve, 10000).unref(); });});
        });
      `,
      errors: [
        {
          messageId: 'globalSetTimeout',
          line: 2,
          column: 3,
        },
        {
          messageId: 'orderSetTimeout',
          line: 2,
          column: 3,
        },
      ],
    },
    {
      code: `
        describe('B', () => {
          it('B.1', async () => {
            await new Promise((resolve) => {
              jest.setTimeout(1000);
              setTimeout(resolve, 10000).unref();
            });
          });
          it('B.2', async () => {
            await new Promise((resolve) => { setTimeout(resolve, 10000).unref(); });
          });
        });
      `,
      errors: [
        {
          messageId: 'globalSetTimeout',
          line: 4,
          column: 7,
        },
        {
          messageId: 'orderSetTimeout',
          line: 4,
          column: 7,
        },
      ],
    },
    {
      code: `
        test('test-suite', () => {
          jest.setTimeout(1000);
        });
      `,
      errors: [
        {
          messageId: 'globalSetTimeout',
          line: 2,
          column: 3,
        },
        {
          messageId: 'orderSetTimeout',
          line: 2,
          column: 3,
        },
      ],
    },
    {
      code: `
        describe('A', () => {
          beforeEach(async () => { await new Promise(resolve => { setTimeout(resolve, 10000).unref(); });});
          it('A.1', async () => { await new Promise(resolve => { setTimeout(resolve, 10000).unref(); });});
          it('A.2', async () => { await new Promise(resolve => { setTimeout(resolve, 10000).unref(); });});
        });
        jest.setTimeout(1000);
      `,
      errors: [
        {
          messageId: 'orderSetTimeout',
          line: 6,
          column: 1,
        },
      ],
    },
    {
      code: `
        import { jest } from '@jest/globals';
        {
          jest.setTimeout(800);
        }
        describe('A', () => {
          beforeEach(async () => { await new Promise(resolve => { setTimeout(resolve, 10000).unref(); });});
          it('A.1', async () => { await new Promise(resolve => { setTimeout(resolve, 10000).unref(); });});
          it('A.2', async () => { await new Promise(resolve => { setTimeout(resolve, 10000).unref(); });});
        });
      `,
      errors: [
        {
          messageId: 'globalSetTimeout',
          line: 3,
          column: 3,
        },
      ],
    },
    {
      code: `
        jest.setTimeout(800);
        jest.setTimeout(900);
      `,
      errors: [
        {
          messageId: 'multipleSetTimeouts',
          line: 2,
          column: 1,
        },
      ],
    },
    {
      code: `
        expect(1 + 2).toEqual(3);
        jest.setTimeout(800);
      `,
      errors: [
        {
          messageId: 'orderSetTimeout',
          line: 2,
          column: 1,
        },
      ],
    },
    {
      code: `
        import { jest as Jest } from '@jest/globals';
        {
          Jest.setTimeout(800);
        }
      `,
      errors: [
        {
          messageId: 'globalSetTimeout',
          line: 3,
          column: 3,
        },
      ],
    },
  ],
});
