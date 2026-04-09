import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-unsafe-negation', {
  valid: [
    'a in b',
    'a instanceof b',
    'a in b === false',
    'a instanceof b === false',
    '!(a in b)',
    '!(a instanceof b)',
    '(!a) in b',
    '(!a) instanceof b',

    // Ordering relations NOT checked by default
    '!a < b',
    '!a > b',
    '!a <= b',
    '!a >= b',
    'if (! a < b) {}',
    'while (! a > b) {}',
    'foo = ! a <= b',
    'foo = ! a >= b',

    // Empty options still defaults to not enforcing ordering
    { code: '! a <= b', options: {} },
    // Explicitly disabled
    { code: 'foo = ! a >= b', options: { enforceForOrderingRelations: false } },

    // Parenthesized negation with ordering operators (option enabled)
    {
      code: '(!a) < b',
      options: { enforceForOrderingRelations: true },
    },
    {
      code: '(!a) > b',
      options: { enforceForOrderingRelations: true },
    },
    {
      code: '(!a) <= b',
      options: { enforceForOrderingRelations: true },
    },
    {
      code: '(!a) >= b',
      options: { enforceForOrderingRelations: true },
    },

    // No negation at all with option enabled
    { code: 'a <= b', options: { enforceForOrderingRelations: true } },
    { code: 'foo = a > b', options: { enforceForOrderingRelations: true } },
    // Properly negated whole expression
    { code: '!(a < b)', options: { enforceForOrderingRelations: true } },
  ],
  invalid: [
    {
      code: '!a in b',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: '(!a in b)',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: '!(a) in b',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: '!a instanceof b',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: '(!a instanceof b)',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: '!(a) instanceof b',
      errors: [{ messageId: 'unexpected' }],
    },

    // Ordering relations with enforceForOrderingRelations: true
    {
      code: 'if (! a < b) {}',
      options: { enforceForOrderingRelations: true },
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'while (! a > b) {}',
      options: { enforceForOrderingRelations: true },
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'foo = ! a <= b',
      options: { enforceForOrderingRelations: true },
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'foo = ! a >= b',
      options: { enforceForOrderingRelations: true },
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: '! a <= b',
      options: { enforceForOrderingRelations: true },
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: '!a < b',
      options: { enforceForOrderingRelations: true },
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: '!a > b',
      options: { enforceForOrderingRelations: true },
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: '!a <= b',
      options: { enforceForOrderingRelations: true },
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: '!a >= b',
      options: { enforceForOrderingRelations: true },
      errors: [{ messageId: 'unexpected' }],
    },
  ],
});
