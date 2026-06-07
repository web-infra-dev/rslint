import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

const message = '"Promise" is not defined.';

ruleTester.run('no-native', {} as never, {
  valid: [
    {
      code: 'var Promise = null; function x() { return Promise.resolve("hi"); }',
    },
    {
      code: 'var Promise = window.Promise || require("bluebird"); var x = Promise.reject();',
    },
    { code: 'import Promise from "bluebird"; var x = Promise.reject();' },
    { code: 'function f(Promise) { return Promise.resolve(1); }' },
  ],

  invalid: [
    {
      code: 'new Promise(function(reject, resolve) { })',
      errors: [{ message }],
    },
    {
      code: 'Promise.resolve()',
      errors: [{ message }],
    },
    {
      code: 'Promise.all([]); Promise.resolve(1);',
      errors: [{ message }, { message }],
    },
  ],
});
