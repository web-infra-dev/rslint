import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-empty', {
  valid: [
    'if (foo) { bar() }',
    'while (foo) { bar() }',
    'for (;;) { bar() }',
    'try { foo() } catch (e) { bar() }',
    'switch (foo) { case 1: break; }',
    'function foo() {}',
    'var foo = function() {}',
    'var foo = () => {}',
    'if (foo) { /* comment */ }',
    'while (foo) { /* comment */ }',
    'try { foo() } catch (e) { /* comment */ }',
    // allowEmptyCatch option
    {
      code: 'try { foo() } catch (e) {}',
      options: { allowEmptyCatch: true },
    },
  ],
  invalid: [
    {
      code: 'if (foo) {}',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'while (foo) {}',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'for (;;) {}',
      errors: [{ messageId: 'unexpected' }],
    },
    // Both try and catch empty
    {
      code: 'try {} catch (e) {}',
      errors: [{ messageId: 'unexpected' }, { messageId: 'unexpected' }],
    },
    {
      code: 'try { foo() } catch (e) {}',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'switch (foo) {}',
      errors: [{ messageId: 'unexpected' }],
    },
  ],
});
