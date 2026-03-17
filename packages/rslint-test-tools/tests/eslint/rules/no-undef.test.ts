import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-undef', {
  valid: [
    'var a = 1; a;',
    'function f() { }; f();',
    'var a: number; a = 1;',
    'typeof undeclaredVar',
    'console.log("hello");',
  ],
  invalid: [
    {
      code: 'unknownVariable123 = 1;',
      errors: [{ messageId: 'undef' }],
    },
  ],
});
