import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

const message = 'Prefer `catch` to `then(a, b)`/`then(null, b)`.';

ruleTester.run('prefer-catch', {} as never, {
  valid: [
    // ESLint upstream
    { code: 'prom.then()' },
    { code: 'prom.then(fn)' },
    { code: 'prom.then(fn1).then(fn2)' },
    { code: 'prom.then(() => {})' },
    { code: 'prom.then(function () {})' },
    { code: 'prom.catch()' },
    { code: 'prom.catch(handleErr).then(handle)' },
    { code: 'prom.catch(handleErr)' },
  ],

  invalid: [
    // ESLint upstream
    {
      code: 'hey.then(fn1, fn2)',
      errors: [{ message }],
      output: 'hey.catch(fn2).then(fn1)',
    },
    {
      code: 'hey.then(fn1, (fn2))',
      errors: [{ message }],
      output: 'hey.catch(fn2).then(fn1)',
    },
    {
      code: 'hey.then(null, fn2)',
      errors: [{ message }],
      output: 'hey.catch(fn2)',
    },
    {
      code: 'hey.then(undefined, fn2)',
      errors: [{ message }],
      output: 'hey.catch(fn2)',
    },
    {
      code: 'function foo() { hey.then(x => {}, () => {}) }',
      errors: [{ message }],
      output: 'function foo() { hey.catch(() => {}).then(x => {}) }',
    },
    {
      code: `
function foo() {
  hey.then(function a() { }, function b() {}).then(fn1, fn2)
}
`,
      errors: [{ message }, { message }],
      output: `
function foo() {
  hey.catch(function b() {}).then(function a() { }).catch(fn2).then(fn1)
}
`,
    },
  ],
});
