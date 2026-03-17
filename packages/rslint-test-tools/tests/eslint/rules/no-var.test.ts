import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-var', {
  valid: [
    "const JOE = 'schmoe';",
    "let moo = 'car';",
    'const a = 1; let b = 2;',
  ],
  invalid: [
    {
      code: 'var foo = bar;',
      errors: [{ messageId: 'unexpectedVar' }],
    },
    {
      code: 'var foo = bar, toast = most;',
      errors: [{ messageId: 'unexpectedVar' }],
    },
  ],
});
