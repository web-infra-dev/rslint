import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-case-declarations', {
  valid: [
    'switch (a) { case 1: { let x = 1; break; } }',
    'switch (a) { case 1: { const x = 1; break; } }',
    'switch (a) { case 1: { function f() {} break; } }',
    'switch (a) { case 1: { class C {} break; } }',
    'switch (a) { case 1: var x = 1; break; }',
    'switch (a) { default: var x = 1; break; }',
    'switch (a) { case 1: break; }',
  ],
  invalid: [
    {
      code: 'switch (a) { case 1: let x = 1; break; }',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'switch (a) { case 1: const x = 1; break; }',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'switch (a) { case 1: function f() {} break; }',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'switch (a) { case 1: class C {} break; }',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'switch (a) { default: let x = 1; break; }',
      errors: [{ messageId: 'unexpected' }],
    },
  ],
});
