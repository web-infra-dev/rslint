import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-unmodified-loop-condition', {
  valid: ['var foo = 0; while (foo) { ++foo; }', 'while (ok(foo)) { }'],
  invalid: [
    {
      code: 'var foo = 0; while (foo) { }',
      errors: [{ messageId: 'loopConditionNotModified' }],
    },
  ],
});
