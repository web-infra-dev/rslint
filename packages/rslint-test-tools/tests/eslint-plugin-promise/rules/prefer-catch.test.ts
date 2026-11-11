import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

const ruleTester = new RuleTester(config);

ruleTester.run('prefer-catch', {
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
      errors: [
        {
          messageId: 'preferCatchToThen',
          message: 'Prefer `catch` to `then(a, b)`/`then(null, b)`.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 9,
        },
      ],
      output: 'hey.catch(fn2).then(fn1)',
    },
    {
      code: 'hey.then(fn1, (fn2))',
      errors: [
        {
          messageId: 'preferCatchToThen',
          message: 'Prefer `catch` to `then(a, b)`/`then(null, b)`.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 9,
        },
      ],
      output: 'hey.catch(fn2).then(fn1)',
    },
    {
      code: 'hey.then(null, fn2)',
      errors: [
        {
          messageId: 'preferCatchToThen',
          message: 'Prefer `catch` to `then(a, b)`/`then(null, b)`.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 9,
        },
      ],
      output: 'hey.catch(fn2)',
    },
    {
      code: 'hey.then(undefined, fn2)',
      errors: [
        {
          messageId: 'preferCatchToThen',
          message: 'Prefer `catch` to `then(a, b)`/`then(null, b)`.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 9,
        },
      ],
      output: 'hey.catch(fn2)',
    },
    {
      code: 'function foo() { hey.then(x => {}, () => {}) }',
      errors: [
        {
          messageId: 'preferCatchToThen',
          message: 'Prefer `catch` to `then(a, b)`/`then(null, b)`.',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 26,
        },
      ],
      output: 'function foo() { hey.catch(() => {}).then(x => {}) }',
    },
    {
      code: `
function foo() {
  hey.then(function a() { }, function b() {}).then(fn1, fn2)
}
`,
      errors: [
        {
          messageId: 'preferCatchToThen',
          message: 'Prefer `catch` to `then(a, b)`/`then(null, b)`.',
          line: 3,
          column: 7,
          endLine: 3,
          endColumn: 11,
        },
        {
          messageId: 'preferCatchToThen',
          message: 'Prefer `catch` to `then(a, b)`/`then(null, b)`.',
          line: 3,
          column: 47,
          endLine: 3,
          endColumn: 51,
        },
      ],
      output:
        '\nfunction foo() {\n  hey.catch(function b() {}).then(function a() { }).catch(fn2).then(fn1)\n}\n',
    },
  ],
});
