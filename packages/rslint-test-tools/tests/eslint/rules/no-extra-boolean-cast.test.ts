import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-extra-boolean-cast', {
  valid: [
    'var foo = !!bar;',
    'function foo() { return !!bar; }',
    'if (foo) {}',
    // new Boolean outside boolean context
    'var x = new Boolean(foo);',
  ],
  invalid: [
    {
      code: 'if (!!foo) {}',
      errors: [{ messageId: 'unexpectedNegation' }],
    },
    {
      code: 'if (Boolean(foo)) {}',
      errors: [{ messageId: 'unexpectedCall' }],
    },
    {
      code: 'while (!!foo) {}',
      errors: [{ messageId: 'unexpectedNegation' }],
    },
    // new Boolean() in boolean context
    {
      code: 'if (new Boolean(foo)) {}',
      errors: [{ messageId: 'unexpectedCall' }],
    },
    // !! inside Boolean() call
    {
      code: 'Boolean(!!foo)',
      errors: [{ messageId: 'unexpectedNegation' }],
    },
  ],
});
