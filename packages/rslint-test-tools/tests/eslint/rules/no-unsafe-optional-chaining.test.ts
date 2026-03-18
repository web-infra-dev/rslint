import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-unsafe-optional-chaining', {
  valid: [
    'obj?.foo;',
    'obj?.foo();',
    '(obj?.foo ?? bar)();',
    // Spread in object literal is safe
    '({...obj?.foo});',
  ],
  invalid: [
    {
      code: '(obj?.foo)();',
      errors: [{ messageId: 'unsafeOptionalChain' }],
    },
    {
      code: '(obj?.foo).bar;',
      errors: [{ messageId: 'unsafeOptionalChain' }],
    },
    {
      code: 'new (obj?.foo)();',
      errors: [{ messageId: 'unsafeOptionalChain' }],
    },
    // NonNullExpression
    {
      code: '(obj?.foo)!.bar;',
      errors: [{ messageId: 'unsafeOptionalChain' }],
    },
    {
      code: '(obj?.foo)!();',
      errors: [{ messageId: 'unsafeOptionalChain' }],
    },
    // Spread in array
    {
      code: '[...obj?.foo];',
      errors: [{ messageId: 'unsafeOptionalChain' }],
    },
  ],
});
