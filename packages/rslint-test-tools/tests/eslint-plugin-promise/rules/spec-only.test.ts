// Upstream: eslint-plugin-promise v7.3.0 (__tests__/spec-only.js and docs/rules/spec-only.md).
import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

const ruleTester = new RuleTester(config);

ruleTester.run('spec-only', {
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
      errors: [
        {
          messageId: 'avoidNonStandard',
          message: "Avoid using non-standard 'Promise.done'",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 13,
        },
      ],
    },
    {
      code: 'Promise.something()',
      errors: [
        {
          messageId: 'avoidNonStandard',
          message: "Avoid using non-standard 'Promise.something'",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: 'new Promise.done()',
      errors: [
        {
          messageId: 'avoidNonStandard',
          message: "Avoid using non-standard 'Promise.done'",
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      code: `
        function foo() {
          var a = getA()
          return Promise.done(a)
        }
      `,
      errors: [
        {
          messageId: 'avoidNonStandard',
          message: "Avoid using non-standard 'Promise.done'",
          line: 4,
          column: 18,
          endLine: 4,
          endColumn: 30,
        },
      ],
    },
    {
      code: `
        function foo() {
          getA(Promise.done)
        }
      `,
      errors: [
        {
          messageId: 'avoidNonStandard',
          message: "Avoid using non-standard 'Promise.done'",
          line: 3,
          column: 16,
          endLine: 3,
          endColumn: 28,
        },
      ],
    },
    {
      code: 'var done = Promise.prototype.done',
      errors: [
        {
          messageId: 'avoidNonStandard',
          message: "Avoid using non-standard 'Promise.prototype'",
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      code: 'Promise["done"];',
      errors: [
        {
          messageId: 'avoidNonStandard',
          message: "Avoid using non-standard 'Promise.done'",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      // Pinned documentation: invalid example.
      code: "const x = Promise.done('bad')",
      errors: [
        {
          messageId: 'avoidNonStandard',
          message: "Avoid using non-standard 'Promise.done'",
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
  ],
});
