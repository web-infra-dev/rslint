import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

const ruleTester = new RuleTester(config);

ruleTester.run('no-return-in-finally', {
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
      errors: [
        {
          messageId: 'noReturnInFinally',
          message: 'No return in finally',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    {
      code: 'Promise.reject(0).finally(() => { return 2 })',
      errors: [
        {
          messageId: 'noReturnInFinally',
          message: 'No return in finally',
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: 'myPromise.finally(() => { return 2 });',
      errors: [
        {
          messageId: 'noReturnInFinally',
          message: 'No return in finally',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: 'Promise.resolve(1).finally(function () { return 2 })',
      errors: [
        {
          messageId: 'noReturnInFinally',
          message: 'No return in finally',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
  ],
});
