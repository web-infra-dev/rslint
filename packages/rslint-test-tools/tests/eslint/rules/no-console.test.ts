import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-console', null as never, {
  valid: [
    'Console.log("foo")',
    {
      code: 'console.log("foo")',
      options: [{ allow: ['log'] }],
    },
    {
      code: 'console.error("foo")',
      options: [{ allow: ['error'] }],
    },
    {
      code: 'console["log"]("foo")',
      options: [{ allow: ['log'] }],
    },
    {
      code: 'console[`log`]("foo")',
      options: [{ allow: ['log'] }],
    },
  ],
  invalid: [
    {
      code: 'console.log("foo")',
      errors: [{ messageId: 'unexpected', line: 1, column: 1 }],
    },
    {
      code: 'console.error("foo")',
      errors: [{ messageId: 'unexpected', line: 1, column: 1 }],
    },
    {
      code: 'console.log',
      errors: [{ messageId: 'unexpected', line: 1, column: 1 }],
    },
    {
      code: 'console.log = foo',
      errors: [{ messageId: 'unexpected', line: 1, column: 1 }],
    },
    {
      code: 'console["log"]',
      errors: [{ messageId: 'unexpected', line: 1, column: 1 }],
    },
    {
      code: 'console.warn("foo")',
      options: [{ allow: ['log'] }],
      errors: [{ messageId: 'unexpected', line: 1, column: 1 }],
    },
  ],
});
