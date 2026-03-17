import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('default-case-last', {
  valid: [
    'switch (foo) {}',
    'switch (foo) { case 1: bar(); break; }',
    'switch (foo) { case 1: break; case 2: break; }',
    'switch (foo) { default: bar(); break; }',
    'switch (foo) { default: }',
    'switch (foo) { case 1: break; default: break; }',
    'switch (foo) { case 1: default: break; }',
    'switch (foo) { case 1: break; case 2: break; default: break; }',
    'switch (foo) { case 1: case 2: default: }',
  ],
  invalid: [
    {
      code: 'switch (foo) { default: bar(); break; case 1: baz(); break; }',
      errors: [{ messageId: 'notLast' }],
    },
    {
      code: 'switch (foo) { default: break; case 1: break; }',
      errors: [{ messageId: 'notLast' }],
    },
    {
      code: 'switch (foo) { default: case 1: break; }',
      errors: [{ messageId: 'notLast' }],
    },
    {
      code: 'switch (foo) { case 1: break; default: break; case 2: break; }',
      errors: [{ messageId: 'notLast' }],
    },
    {
      code: 'switch (foo) { case 1: default: break; case 2: break; }',
      errors: [{ messageId: 'notLast' }],
    },
  ],
});
