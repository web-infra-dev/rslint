import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-dupe-keys', {
  valid: [
    'var x = { a: 1, b: 2 };',
    'var x = { a: 1, b: 2, c: 3 };',
    'var x = { get a() {}, set a(v: any) {} };',
    'var x = { [Symbol()]: 1, [Symbol()]: 2 };',
    'var x = { "": 1, " ": 2 };',
  ],
  invalid: [
    {
      code: 'var x = { a: 1, a: 2 };',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'var x = { a: 1, b: 2, a: 3 };',
      errors: [{ messageId: 'unexpected' }],
    },
    // Multiple duplicates
    {
      code: 'var x = { a: 1, b: 2, a: 3, a: 4 };',
      errors: [{ messageId: 'unexpected' }, { messageId: 'unexpected' }],
    },
    {
      code: 'var x = { "a": 1, "a": 2 };',
      errors: [{ messageId: 'unexpected' }],
    },
    // Numeric literal equivalence: 0x1 and 1 are the same key
    {
      code: 'var x = { 0x1: "a", 1: "b" };',
      errors: [{ messageId: 'unexpected' }],
    },
  ],
});
