import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('valid-describe-callback', {} as never, {
  valid: [
    { code: "describe.each([1, 2, 3])('%s', (a, b) => {});" },
    { code: "describe('foo', function() {})" },
    { code: "describe('foo', () => {})" },
    { code: 'describe(`foo`, () => {})' },
    { code: "xdescribe('foo', () => {})" },
    { code: "fdescribe('foo', () => {})" },
    { code: "describe.only('foo', () => {})" },
    { code: "describe.skip('foo', () => {})" },
    {
      code: `describe('foo', () => {
        it('bar', () => {
          return Promise.resolve(42).then(value => {
            expect(value).toBe(42)
          })
        })
      })`,
    },
    {
      code: `describe('foo', () => {
        it('bar', async () => {
          expect(await Promise.resolve(42)).toBe(42)
        })
      })`,
    },
    { code: 'if (hasOwnProperty(obj, key)) {}' },
    {
      code: "describe.each`\n          foo  | foe\n          ${'1'} | ${'2'}\n        `('$something', ({ foo, foe }) => {});",
    },
  ],
  invalid: [
    {
      code: 'describe.each()()',
      errors: [{ message: 'Describe requires name and callback arguments' }],
    },
    {
      code: "describe['each']()()",
      errors: [{ message: 'Describe requires name and callback arguments' }],
    },
    {
      code: 'describe.each(() => {})()',
      errors: [{ message: 'Describe requires name and callback arguments' }],
    },
    {
      code: "describe.each(() => {})('foo')",
      errors: [{ message: 'Describe requires name and callback arguments' }],
    },
    {
      code: 'describe.each()(() => {})',
      errors: [{ message: 'Describe requires name and callback arguments' }],
    },
    {
      code: "describe['each']()(() => {})",
      errors: [{ message: 'Describe requires name and callback arguments' }],
    },
    {
      code: "describe.each('foo')(() => {})",
      errors: [{ message: 'Describe requires name and callback arguments' }],
    },
    {
      code: "describe.only.each('foo')(() => {})",
      errors: [{ message: 'Describe requires name and callback arguments' }],
    },
    {
      code: 'describe(() => {})',
      errors: [{ message: 'Describe requires name and callback arguments' }],
    },
    {
      code: "describe('foo')",
      errors: [{ message: 'Describe requires name and callback arguments' }],
    },
    {
      code: "describe('foo', 'foo2')",
      errors: [{ message: 'Second argument must be function' }],
    },
    {
      code: 'describe()',
      errors: [{ message: 'Describe requires name and callback arguments' }],
    },
    {
      code: "describe('foo', async () => {})",
      errors: [{ message: 'No async describe callback' }],
    },
    {
      code: "describe('foo', async function () {})",
      errors: [{ message: 'No async describe callback' }],
    },
    {
      code: "xdescribe('foo', async function () {})",
      errors: [{ message: 'No async describe callback' }],
    },
    {
      code: "fdescribe('foo', async function () {})",
      errors: [{ message: 'No async describe callback' }],
    },
    {
      code: `
        import { fdescribe } from '@jest/globals';
        fdescribe('foo', async function () {})
      `,
      errors: [{ message: 'No async describe callback' }],
    },
    {
      code: "describe.only('foo', async function () {})",
      errors: [{ message: 'No async describe callback' }],
    },
    {
      code: "describe.skip('foo', async function () {})",
      errors: [{ message: 'No async describe callback' }],
    },
    {
      code: `describe('sample case', () => {
        it('works', () => {
          expect(true).toEqual(true);
        });
        describe('async', async () => {
          await new Promise(setImmediate);
          it('breaks', () => {
            throw new Error('Fail');
          });
        });
      });`,
      errors: [{ message: 'No async describe callback' }],
    },
    {
      code: `describe('foo', function () {
        return Promise.resolve().then(() => {
          it('breaks', () => {
            throw new Error('Fail')
          })
        })
      })`,
      errors: [{ message: 'Unexpected return statement in describe callback' }],
    },
    {
      code: `describe('foo', () => {
        return Promise.resolve().then(() => {
          it('breaks', () => {
            throw new Error('Fail')
          })
        })
        describe('nested', () => {
          return Promise.resolve().then(() => {
            it('breaks', () => {
              throw new Error('Fail')
            })
          })
        })
      })`,
      errors: [
        { message: 'Unexpected return statement in describe callback' },
        { message: 'Unexpected return statement in describe callback' },
      ],
    },
    {
      code: `describe('foo', async () => {
        await something()
        it('does something')
        describe('nested', () => {
          return Promise.resolve().then(() => {
            it('breaks', () => {
              throw new Error('Fail')
            })
          })
        })
      })`,
      errors: [
        { message: 'No async describe callback' },
        { message: 'Unexpected return statement in describe callback' },
      ],
    },
    {
      code: "describe('foo', () => test('bar', () => {}))",
      errors: [{ message: 'Unexpected return statement in describe callback' }],
    },
    {
      code: "describe('foo', done => {})",
      errors: [{ message: 'Unexpected argument(s) in describe callback' }],
    },
    {
      code: "describe('foo', function (done) {})",
      errors: [{ message: 'Unexpected argument(s) in describe callback' }],
    },
    {
      code: "describe('foo', function (one, two, three) {})",
      errors: [{ message: 'Unexpected argument(s) in describe callback' }],
    },
    {
      code: "describe('foo', async function (done) {})",
      errors: [
        { message: 'No async describe callback' },
        { message: 'Unexpected argument(s) in describe callback' },
      ],
    },
  ],
});
