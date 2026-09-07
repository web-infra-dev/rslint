// Upstream: eslint-plugin-promise v7.3.0 (__tests__/spec-only.js and docs/rules/spec-only.md).
// Exact diagnostic IDs/ranges and absence of edits are asserted in the Go suite.
import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('spec-only', {} as never, {
  valid: [
    { code: 'Promise.resolve()' },
    { code: 'Promise.reject()' },
    { code: 'Promise.all()' },
    { code: 'Promise["all"]' },
    { code: 'Promise[method];' },
    { code: 'Promise.prototype;' },
    { code: 'Promise.prototype[method];' },
    { code: 'Promise.prototype["then"];' },
    { code: 'Promise.race()' },
    { code: 'var ctch = Promise.prototype.catch' },
    { code: 'Promise.withResolvers()' },
    { code: 'new Promise(function (resolve, reject) {})' },
    { code: 'SomeClass.resolve()' },
    { code: 'doSomething(Promise.all)' },
    {
      code: 'Promise.permittedMethod()',
      options: [{ allowedMethods: ['permittedMethod'] }],
    },
    {
      code: 'Promise.prototype.permittedInstanceMethod',
      options: [{ allowedMethods: ['permittedInstanceMethod'] }],
    },
    // Pinned documentation: valid example.
    { code: "const x = Promise.resolve('good')" },
  ],
  invalid: [
    {
      code: 'Promise.done()',
      errors: [{ message: "Avoid using non-standard 'Promise.done'" }],
    },
    {
      code: 'Promise.something()',
      errors: [{ message: "Avoid using non-standard 'Promise.something'" }],
    },
    {
      code: 'new Promise.done()',
      errors: [{ message: "Avoid using non-standard 'Promise.done'" }],
    },
    {
      code: `
        function foo() {
          var a = getA()
          return Promise.done(a)
        }
      `,
      errors: [{ message: "Avoid using non-standard 'Promise.done'" }],
    },
    {
      code: `
        function foo() {
          getA(Promise.done)
        }
      `,
      errors: [{ message: "Avoid using non-standard 'Promise.done'" }],
    },
    {
      code: 'var done = Promise.prototype.done',
      errors: [{ message: "Avoid using non-standard 'Promise.prototype'" }],
    },
    {
      code: 'Promise["done"];',
      errors: [{ message: "Avoid using non-standard 'Promise.done'" }],
    },
    {
      // Pinned documentation: invalid example.
      code: "const x = Promise.done('bad')",
      errors: [{ message: "Avoid using non-standard 'Promise.done'" }],
    },
  ],
});
