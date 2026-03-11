import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-dupe-keys', {
  valid: [
    'var x = { a: 1, b: 2 };',
    'var x = { a: 1, b: 2, c: 3 };',
    'var x = { get a() {}, set a(v: any) {} };',
    'var x = { [Symbol()]: 1, [Symbol()]: 2 };',
    'var x = { "": 1, " ": 2 };',
    // __proto__ as proto setter is allowed to appear multiple times
    'var x = { __proto__: foo, __proto__: bar };',
    'var x = { "__proto__": foo, "__proto__": bar };',
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
    // Computed __proto__ is a regular property, not a proto setter
    {
      code: 'var x = { ["__proto__"]: 1, ["__proto__"]: 2 };',
      errors: [{ messageId: 'unexpected' }],
    },
    // Numeric literal equivalence: 0x1 and 1 are the same key
    {
      code: 'var x = { 0x1: "a", 1: "b" };',
      errors: [{ messageId: 'unexpected' }],
    },
    // BigInt literal equivalence: 0x1n and 1n normalize to the same key
    {
      code: 'var x = { [0x1n]: "a", [1n]: "b" };',
      errors: [{ messageId: 'unexpected' }],
    },
    // Template literal computed property
    {
      code: 'var x = { [`key`]: 1, [`key`]: 2 };',
      errors: [{ messageId: 'unexpected' }],
    },
    // Numeric overflow to Infinity: 1e309 and 1e999 both normalize to "Infinity"
    {
      code: 'var x = { [1e309]: "a", [1e999]: "b" };',
      errors: [{ messageId: 'unexpected' }],
    },
  ],
});
