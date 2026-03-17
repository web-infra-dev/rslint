import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-new-wrappers', {
  valid: [
    'var a = new Object();',
    "var a = String('test'), b = String.fromCharCode(32);",
  ],
  invalid: [
    {
      code: "var a = new String('hello');",
      errors: [{ messageId: 'noConstructor' }],
    },
    {
      code: 'var a = new Number(10);',
      errors: [{ messageId: 'noConstructor' }],
    },
    {
      code: 'var a = new Boolean(false);',
      errors: [{ messageId: 'noConstructor' }],
    },
  ],
});
