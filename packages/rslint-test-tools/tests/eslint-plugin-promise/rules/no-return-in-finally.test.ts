import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

const noReturnMsg = 'No return in finally';

ruleTester.run('no-return-in-finally', {} as never, {
  valid: [
    // ESLint upstream
    { code: 'Promise.resolve(1).finally(() => { console.log(2) })' },
    { code: 'Promise.reject(4).finally(() => { console.log(2) })' },
    { code: 'Promise.reject(4).finally(() => {})' },
    { code: 'myPromise.finally(() => {});' },
    { code: 'Promise.resolve(1).finally(function () { })' },
  ],

  invalid: [
    // ESLint upstream
    {
      code: 'Promise.resolve(1).finally(() => { return 2 })',
      errors: [{ message: noReturnMsg }],
    },
    {
      code: 'Promise.reject(0).finally(() => { return 2 })',
      errors: [{ message: noReturnMsg }],
    },
    {
      code: 'myPromise.finally(() => { return 2 });',
      errors: [{ message: noReturnMsg }],
    },
    {
      code: 'Promise.resolve(1).finally(function () { return 2 })',
      errors: [{ message: noReturnMsg }],
    },
  ],
});
