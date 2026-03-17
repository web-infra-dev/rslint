import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-fallthrough', {
  valid: [
    'switch(foo) { case 0: a(); break; case 1: b(); }',
    'switch(foo) { case 0: case 1: a(); break; }',
    'switch(foo) { case 0: a(); /* falls through */ case 1: b(); }',
  ],
  invalid: [
    {
      code: 'switch(foo) { case 0: a(); case 1: b(); }',
      errors: [{ messageId: 'case' }],
    },
    {
      code: 'switch(foo) { case 0: a(); default: b(); }',
      errors: [{ messageId: 'default' }],
    },
  ],
});
