import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-new-native-nonconstructor', {
  valid: [
    // Symbol
    "var foo = Symbol('foo');",
    "function bar(Symbol) { var baz = new Symbol('baz');}",
    'function Symbol() {} new Symbol();',
    'new foo(Symbol);',
    'new foo(bar, Symbol);',

    // BigInt
    'var foo = BigInt(9007199254740991);',
    'function bar(BigInt) { var baz = new BigInt(9007199254740991);}',
    'function BigInt() {} new BigInt();',
    'new foo(BigInt);',
    'new foo(bar, BigInt);',
  ],
  invalid: [
    // Symbol
    {
      code: "var foo = new Symbol('foo');",
      errors: [{ messageId: 'noNewNonconstructor' }],
    },
    {
      code: "function bar() { return function Symbol() {}; } var baz = new Symbol('baz');",
      errors: [{ messageId: 'noNewNonconstructor' }],
    },

    // BigInt
    {
      code: 'var foo = new BigInt(9007199254740991);',
      errors: [{ messageId: 'noNewNonconstructor' }],
    },
    {
      code: 'function bar() { return function BigInt() {}; } var baz = new BigInt(9007199254740991);',
      errors: [{ messageId: 'noNewNonconstructor' }],
    },
  ],
});
