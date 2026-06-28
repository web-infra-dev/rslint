import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

const errorMessage = 'Avoid creating new promises.';

ruleTester.run('avoid-new', {} as never, {
  valid: [
    { code: 'Promise.resolve()' },
    { code: 'Promise.reject()' },
    { code: 'Promise.all()' },
    { code: 'new Horse()' },
    { code: 'new PromiseLikeThing()' },
    { code: 'new Promise.resolve()' },
  ],

  invalid: [
    {
      code: 'var x = new Promise(function (x, y) {})',
      errors: [{ message: errorMessage }],
    },
    {
      code: 'new Promise()',
      errors: [{ message: errorMessage }],
    },
    {
      code: 'Thing(new Promise(() => {}))',
      errors: [{ message: errorMessage }],
    },
  ],
});
