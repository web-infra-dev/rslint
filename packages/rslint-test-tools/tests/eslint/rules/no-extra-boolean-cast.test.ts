import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

const enforceLogicalOperands = [{ enforceForLogicalOperands: true }] as any;
const enforceInnerExpressions = [{ enforceForInnerExpressions: true }] as any;

ruleTester.run('no-extra-boolean-cast', {
  valid: [
    // !! not in boolean context
    'var foo = !!bar;',
    'function foo() { return !!bar; }',
    // Boolean() not in boolean context
    'var foo = Boolean(bar);',
    // No extra cast
    'if (foo) {}',
    // new Boolean is never flagged — it produces a truthy object,
    // so it is not equivalent to the plain value (matches ESLint).
    'var x = new Boolean(foo);',
    'if (new Boolean(foo)) {}',
    '!new Boolean(foo)',
    // Not flagged when enforceForLogicalOperands / enforceForInnerExpressions are off
    'if (x || !!y) {}',
    'if (x && Boolean(y)) {}',
    'if (x ? !!y : z) {}',
    'if (x ?? !!y) {}',
    // enforceForLogicalOperands does NOT cover `??`
    { code: 'if (x ?? !!y) {}', options: enforceLogicalOperands },
    // enforceForInnerExpressions: not the last expression of a sequence
    { code: 'if ((Boolean(a), b)) {}', options: enforceInnerExpressions },
    // enforceForInnerExpressions: !! is the LEFT side of ??
    { code: 'if (!!x ?? y) {}', options: enforceInnerExpressions },
  ],
  invalid: [
    // !! in if test
    {
      code: 'if (!!foo) {}',
      errors: [{ messageId: 'unexpectedNegation' }],
    },
    // Boolean() in if test
    {
      code: 'if (Boolean(foo)) {}',
      errors: [{ messageId: 'unexpectedCall' }],
    },
    // !! in while test
    {
      code: 'while (!!foo) {}',
      errors: [{ messageId: 'unexpectedNegation' }],
    },
    // !! inside Boolean() call
    {
      code: 'Boolean(!!foo)',
      errors: [{ messageId: 'unexpectedNegation' }],
    },
    // !! nested inside new Boolean() — inner !! is flagged
    {
      code: 'new Boolean(!!foo)',
      errors: [{ messageId: 'unexpectedNegation' }],
    },
    // Parenthesized !! in boolean context
    {
      code: 'if ((!!foo)) {}',
      errors: [{ messageId: 'unexpectedNegation' }],
    },

    // enforceForLogicalOperands: !! inside ||
    {
      code: 'if (x || !!y) {}',
      options: enforceLogicalOperands,
      errors: [{ messageId: 'unexpectedNegation' }],
    },
    // enforceForLogicalOperands: Boolean() inside &&
    {
      code: 'while (x && Boolean(y)) {}',
      options: enforceLogicalOperands,
      errors: [{ messageId: 'unexpectedCall' }],
    },

    // enforceForInnerExpressions: !! on the right of ??
    {
      code: 'if (x ?? !!y) {}',
      options: enforceInnerExpressions,
      errors: [{ messageId: 'unexpectedNegation' }],
    },
    // enforceForInnerExpressions: Boolean() in ternary consequent
    {
      code: 'if (cond ? Boolean(a) : b) {}',
      options: enforceInnerExpressions,
      errors: [{ messageId: 'unexpectedCall' }],
    },
    // enforceForInnerExpressions: !! in ternary alternate
    {
      code: 'if (cond ? a : !!b) {}',
      options: enforceInnerExpressions,
      errors: [{ messageId: 'unexpectedNegation' }],
    },
    // enforceForInnerExpressions: last expression of a sequence
    {
      code: 'if ((a, b, Boolean(c))) {}',
      options: enforceInnerExpressions,
      errors: [{ messageId: 'unexpectedCall' }],
    },
  ],
});
