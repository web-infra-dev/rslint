import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('eqeqeq', {
  valid: [
    // ── Default "always" mode ──
    'a === b',
    'a !== b',
    { code: 'a === b', options: ['always'] as any },

    // ── "smart" mode ──
    { code: 'typeof a == "number"', options: ['smart'] as any },
    { code: '"string" != typeof a', options: ['smart'] as any },
    { code: '"hello" != "world"', options: ['smart'] as any },
    { code: '2 == 3', options: ['smart'] as any },
    { code: 'true == true', options: ['smart'] as any },
    { code: 'null == a', options: ['smart'] as any },
    { code: 'a == null', options: ['smart'] as any },

    // ── "allow-null" mode ──
    { code: 'null == a', options: ['allow-null'] as any },
    { code: 'a == null', options: ['allow-null'] as any },

    // ── "always" with null:"ignore" ──
    { code: 'a == null', options: ['always', { null: 'ignore' }] as any },
    { code: 'a != null', options: ['always', { null: 'ignore' }] as any },
    { code: 'a !== null', options: ['always', { null: 'ignore' }] as any },

    // ── "always" with null:"always" ──
    { code: 'a === null', options: ['always', { null: 'always' }] as any },
    { code: 'a !== null', options: ['always', { null: 'always' }] as any },
    { code: 'null === null', options: ['always', { null: 'always' }] as any },
    { code: 'null !== null', options: ['always', { null: 'always' }] as any },

    // ── "always" with null:"never" ──
    { code: 'a == null', options: ['always', { null: 'never' }] as any },
    { code: 'a != null', options: ['always', { null: 'never' }] as any },
    { code: 'null == null', options: ['always', { null: 'never' }] as any },
    { code: 'null != null', options: ['always', { null: 'never' }] as any },
  ],
  invalid: [
    // ── Default "always" mode ──
    {
      code: 'a == b',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'a != b',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'typeof a == "number"',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: '"string" != typeof a',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'true == true',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: '2 == 3',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: "'hello' != 'world'",
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'a == null',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'null != a',
      errors: [{ messageId: 'unexpected' }],
    },

    // ── "smart" mode ──
    {
      code: 'a == b',
      options: ['smart'] as any,
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'a != b',
      options: ['smart'] as any,
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'true == 1',
      options: ['smart'] as any,
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: '0 != "1"',
      options: ['smart'] as any,
      errors: [{ messageId: 'unexpected' }],
    },

    // ── "allow-null" mode ──
    {
      code: 'typeof a == "number"',
      options: ['allow-null'] as any,
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: '"hello" != "world"',
      options: ['allow-null'] as any,
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: '2 == 3',
      options: ['allow-null'] as any,
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'true == true',
      options: ['allow-null'] as any,
      errors: [{ messageId: 'unexpected' }],
    },

    // ── "always" with null:"always" ──
    {
      code: 'true == null',
      options: ['always', { null: 'always' }] as any,
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'true != null',
      options: ['always', { null: 'always' }] as any,
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'null == null',
      options: ['always', { null: 'always' }] as any,
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'null != null',
      options: ['always', { null: 'always' }] as any,
      errors: [{ messageId: 'unexpected' }],
    },

    // ── "always" with null:"never" ──
    {
      code: 'a === null',
      options: ['always', { null: 'never' }] as any,
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'null !== a',
      options: ['always', { null: 'never' }] as any,
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'null === null',
      options: ['always', { null: 'never' }] as any,
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'null !== null',
      options: ['always', { null: 'never' }] as any,
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'a == b',
      options: ['always', { null: 'never' }] as any,
      errors: [{ messageId: 'unexpected' }],
    },

    // ── "always" with null:"ignore" ──
    {
      code: 'a == b',
      options: ['always', { null: 'ignore' }] as any,
      errors: [{ messageId: 'unexpected' }],
    },

    // ── Parenthesized expressions ──
    {
      code: '(a) == b',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'a == (b)',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: '(a) == (b)',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: '(a == b) == (c)',
      errors: [{ messageId: 'unexpected' }, { messageId: 'unexpected' }],
    },
  ],
});
