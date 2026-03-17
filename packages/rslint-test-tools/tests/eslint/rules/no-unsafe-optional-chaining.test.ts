import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-unsafe-optional-chaining', {
  valid: ['obj?.foo;', 'obj?.foo();', '(obj?.foo ?? bar)();'],
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
  ],
});
