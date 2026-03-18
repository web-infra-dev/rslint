import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-empty-character-class', {
  valid: ['var foo = /^abc[a-zA-Z]/;', 'var foo = /[^]/;'],
  invalid: [
    {
      code: 'var foo = /^abc[]/;',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'var foo = /foo[]bar/;',
      errors: [{ messageId: 'unexpected' }],
    },
  ],
});
