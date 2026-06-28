import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

// Mirrors the upstream curly valid/invalid semantic set (Layer 1). It verifies
// rule registration + wire protocol + ESLint-compatible diagnostic shape; the
// exhaustive edge-shape / branch lock-in coverage lives in the Go suite.
ruleTester.run('curly', {
  valid: [
    // ── "all" (default) ──
    'if (foo) { bar() }',
    'if (foo) { bar() } else if (foo2) { baz() }',
    'while (foo) { bar() }',
    'do { bar(); } while (foo)',
    'for (;foo;) { bar() }',
    'for (var foo in bar) { console.log(foo) }',
    'for (var foo of bar) { console.log(foo) }',

    // ── "multi" ──
    { code: 'if (foo) bar()', options: ['multi'] as any },
    { code: 'if (a) { b; c; }', options: ['multi'] as any },
    {
      code: 'for (var foo in bar) console.log(foo)',
      options: ['multi'] as any,
    },
    {
      code: 'for (var foo of bar) console.log(foo)',
      options: ['multi'] as any,
    },
    { code: "if (foo) { const bar = 'baz'; }", options: ['multi'] as any },
    { code: "while (foo) { let bar = 'baz'; }", options: ['multi'] as any },
    { code: 'for(;;) { function foo() {} }', options: ['multi'] as any },
    { code: 'for (foo in bar) { class Baz {} }', options: ['multi'] as any },

    // ── "multi-line" ──
    { code: 'if (foo) bar()', options: ['multi-line'] as any },
    { code: 'if (foo) bar(); else baz()', options: ['multi-line'] as any },
    { code: 'do baz(); while (foo)', options: ['multi-line'] as any },
    { code: 'if (foo) { bar() }', options: ['multi-line'] as any },

    // ── "multi-or-nest" ──
    { code: 'if (foo) \n quz = true;', options: ['multi-or-nest'] as any },
    {
      code: 'if (foo) { \n // line of comment \n quz = true; \n }',
      options: ['multi-or-nest'] as any,
    },
    {
      code: 'while (true) \n doSomething();',
      options: ['multi-or-nest'] as any,
    },
    {
      code: "if (foo) { \n const bar = 'baz'; \n }",
      options: ['multi-or-nest'] as any,
    },

    // ── "consistent" ──
    {
      code: 'if (foo) { let bar; } else { baz(); }',
      options: ['multi', 'consistent'] as any,
    },
    {
      code: 'if (true) { foo(); } else if (true) { faa(); } else { bar(); baz(); }',
      options: ['multi', 'consistent'] as any,
    },
  ],
  invalid: [
    // ── missingCurlyAfterCondition ──
    {
      code: 'if (foo) bar()',
      errors: [{ messageId: 'missingCurlyAfterCondition' }],
    },
    {
      code: 'while (foo) bar()',
      errors: [{ messageId: 'missingCurlyAfterCondition' }],
    },
    {
      code: 'for (;foo;) bar()',
      errors: [{ messageId: 'missingCurlyAfterCondition' }],
    },
    // ── missingCurlyAfter ──
    {
      code: 'do bar(); while (foo)',
      errors: [{ messageId: 'missingCurlyAfter' }],
    },
    {
      code: 'for (var foo in bar) console.log(foo)',
      errors: [{ messageId: 'missingCurlyAfter' }],
    },
    {
      code: 'for (var foo of bar) console.log(foo)',
      errors: [{ messageId: 'missingCurlyAfter' }],
    },
    {
      code: 'if (foo) { bar() } else baz()',
      errors: [{ messageId: 'missingCurlyAfter' }],
    },
    // ── unexpectedCurlyAfterCondition (multi) ──
    {
      code: 'if (foo) { bar() }',
      options: ['multi'] as any,
      errors: [{ messageId: 'unexpectedCurlyAfterCondition' }],
    },
    {
      code: 'while (foo) { bar() }',
      options: ['multi'] as any,
      errors: [{ messageId: 'unexpectedCurlyAfterCondition' }],
    },
    {
      code: 'for (;foo;) { bar() }',
      options: ['multi'] as any,
      errors: [{ messageId: 'unexpectedCurlyAfterCondition' }],
    },
    // ── unexpectedCurlyAfter (multi) ──
    {
      code: 'do{foo();} while(bar);',
      options: ['multi'] as any,
      errors: [{ messageId: 'unexpectedCurlyAfter' }],
    },
    {
      code: 'for (var foo in bar) {console.log(foo)}',
      options: ['multi'] as any,
      errors: [{ messageId: 'unexpectedCurlyAfter' }],
    },
    {
      code: 'if (foo) baz(); else { bar() }',
      options: ['multi'] as any,
      errors: [{ messageId: 'unexpectedCurlyAfter' }],
    },
    // ── multi-line ──
    {
      code: 'if (foo) \n baz()',
      options: ['multi-line'] as any,
      errors: [{ messageId: 'missingCurlyAfterCondition' }],
    },
    // ── multi-or-nest ──
    {
      code: 'if (foo) { \n quz = true; \n }',
      options: ['multi-or-nest'] as any,
      errors: [{ messageId: 'unexpectedCurlyAfterCondition' }],
    },
    {
      code: 'if (foo) \n quz = { \n bar: baz, \n qux: foo \n };',
      options: ['multi-or-nest'] as any,
      errors: [{ messageId: 'missingCurlyAfterCondition' }],
    },
    // ── consistent ──
    {
      code: 'if (foo) { let bar; } else baz();',
      options: ['multi', 'consistent'] as any,
      errors: [{ messageId: 'missingCurlyAfter' }],
    },
    // ── ASI hazard: reported, but not auto-fixed ──
    {
      code: 'if (foo) {bar()} baz()',
      options: ['multi'] as any,
      errors: [{ messageId: 'unexpectedCurlyAfterCondition' }],
    },
  ],
});
