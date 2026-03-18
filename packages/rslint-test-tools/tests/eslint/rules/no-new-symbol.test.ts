import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-new-symbol', {
  valid: ["var foo = Symbol('foo');", 'new foo(Symbol);'],
  invalid: [
    {
      code: "var foo = new Symbol('foo');",
      errors: [{ messageId: 'noNewSymbol' }],
    },
  ],
});
