import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-unmodified-loop-condition', {
  valid: [
    'var foo = 0; while (foo) { ++foo; }',
    'while (ok(foo)) { }',
    // Group semantics: a < b is one group, a is modified
    'var a = 0, b = 10; while (a < b) { a++; }',
  ],
  invalid: [
    {
      code: 'var foo = 0; while (foo) { }',
      errors: [{ messageId: 'loopConditionNotModified' }],
    },
    // Both unmodified in group
    {
      code: 'var a = 0, b = 0; while (a < b) { }',
      errors: [
        { messageId: 'loopConditionNotModified' },
        { messageId: 'loopConditionNotModified' },
      ],
    },
  ],
});
