import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-caller', {
  valid: [
    'var x = arguments.length',
    'var x = arguments',
    'var x = arguments[0]',
    'var x = arguments[caller]',
  ],
  invalid: [
    {
      code: 'var x = arguments.callee',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'var x = arguments.caller',
      errors: [{ messageId: 'unexpected' }],
    },
  ],
});
