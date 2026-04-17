import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('guard-for-in', {
  valid: [
    // ================================================================
    // ESLint upstream valid cases
    // ================================================================
    'for (var x in o);',
    'for (var x in o) {}',
    'for (var x in o) if (x) f();',
    'for (var x in o) { if (x) { f(); } }',
    'for (var x in o) { if (x) continue; f(); }',
    'for (var x in o) { if (x) { continue; } f(); }',

    // ================================================================
    // Declaration forms
    // ================================================================
    'for (let x in o) if (x) f();',
    'for (const x in o) if (x) f();',
    'for (x in o) if (x) f();',

    // ================================================================
    // Destructuring initializer (body shape is what matters)
    // ================================================================
    'for (const { a } in o) if (a) f();',
    'for (const [a, b] in o) if (a) f();',

    // ================================================================
    // Labeled continue still counts as a continue-guard
    // ================================================================
    'outer: for (var x in o) { if (x) continue outer; f(); }',
    'outer: for (var x in o) { if (x) { continue outer; } f(); }',

    // ================================================================
    // If-else: ESLint only inspects consequent. A continue-consequent
    // guards the rest of the body even when an else branch is present.
    // ================================================================
    'for (var x in o) { if (x) continue; else f(); g(); }',
    'for (var x in o) { if (x) { continue; } else { f(); } g(); }',

    // ================================================================
    // for-of must never be flagged (separate AST kind)
    // ================================================================
    'for (const x of arr) f();',
    'for (const x of arr) { f(); g(); }',

    // ================================================================
    // Nested for-in, both guarded
    // ================================================================
    'for (var x in o) { if (x) continue; for (var y in o2) if (y) g(); }',
    'for (var x in o) if (x) { for (var y in o2) if (y) g(); }',

    // ================================================================
    // Wrapped in various scope contexts — rule fires per for-in
    // regardless of enclosing scope.
    // ================================================================
    'function f() { for (var x in o) if (x) g(); }',
    'const f = () => { for (var x in o) if (x) g(); }',
    'class A { static { for (var x in o) if (x) g(); } }',
    'class A { m() { for (var x in o) if (x) g(); } }',
    '(function () { for (var x in o) if (x) g(); })();',

    // ================================================================
    // Block with only empty if (still matches "block with just if")
    // ================================================================
    'for (var x in o) { if (x); }',

    // ================================================================
    // Comments between tokens are trivia — body shape unchanged
    // ================================================================
    'for (var x in o) /* c */ if (x) f();',
    'for (var x in o) { /* c1 */ if (x) /* c2 */ continue; /* c3 */ f(); }',

    // ================================================================
    // TypeScript expressions on the iterated value don't affect body
    // ================================================================
    'for (var x in (o)) if (x) f();',
    'for (var x in o!) if (x) f();',
    'for (var x in o as any) if (x) f();',

    // ================================================================
    // 3-level nesting, every level guarded
    // ================================================================
    'for (var a in x) { if (a) continue; for (var b in y) { if (b) continue; for (var c in z) if (c) g(); } }',
  ],
  invalid: [
    // ================================================================
    // ESLint upstream invalid cases
    // ================================================================
    {
      code: 'for (var x in o) { if (x) { f(); continue; } g(); }',
      errors: [{ messageId: 'wrap' }],
    },
    {
      code: 'for (var x in o) { if (x) { continue; f(); } g(); }',
      errors: [{ messageId: 'wrap' }],
    },
    {
      code: 'for (var x in o) { if (x) { f(); } g(); }',
      errors: [{ messageId: 'wrap' }],
    },
    {
      code: 'for (var x in o) { if (x) f(); g(); }',
      errors: [{ messageId: 'wrap' }],
    },
    {
      code: 'for (var x in o) { foo() }',
      errors: [{ messageId: 'wrap' }],
    },
    {
      code: 'for (var x in o) foo();',
      errors: [{ messageId: 'wrap' }],
    },

    // ================================================================
    // Single-statement body that isn't EmptyStatement/IfStatement
    // ================================================================
    {
      code: 'for (var x in o) throw x;',
      errors: [{ messageId: 'wrap' }],
    },
    {
      code: 'for (var x in o) for (var y in o2) g();',
      errors: [
        { messageId: 'wrap', line: 1, column: 1 },
        { messageId: 'wrap', line: 1, column: 18 },
      ],
    },
    {
      code: 'for (var x in o) lbl: continue;',
      errors: [{ messageId: 'wrap' }],
    },

    // ================================================================
    // Block variants that must NOT count as guarded
    // ================================================================
    {
      code: 'for (var x in o) { f(); if (x) continue; g(); }',
      errors: [{ messageId: 'wrap' }],
    },
    {
      code: 'for (var x in o) { ; if (x) continue; f(); }',
      errors: [{ messageId: 'wrap' }],
    },
    {
      code: 'for (var x in o) { { if (x) continue; } f(); }',
      errors: [{ messageId: 'wrap' }],
    },
    {
      code: 'for (var x in o) { if (x); f(); }',
      errors: [{ messageId: 'wrap' }],
    },
    {
      code: 'for (var x in o) { if (x) f(); else continue; g(); }',
      errors: [{ messageId: 'wrap' }],
    },
    {
      code: 'for (var x in o) { if (x) { continue; continue; } f(); }',
      errors: [{ messageId: 'wrap' }],
    },

    // ================================================================
    // Nested: only the unguarded loop should be reported
    // ================================================================
    {
      code: 'for (var x in o) { if (x) continue; for (var y in o2) g(); }',
      errors: [{ messageId: 'wrap', line: 1, column: 37 }],
    },
    {
      code: 'for (var x in o) { for (var y in o2) if (y) g(); }',
      errors: [{ messageId: 'wrap', line: 1, column: 1 }],
    },
    {
      code: 'for (var x in o) { for (var y in o2) { g(); } }',
      errors: [
        { messageId: 'wrap', line: 1, column: 1 },
        { messageId: 'wrap', line: 1, column: 20 },
      ],
    },

    // ================================================================
    // Multi-line: line/column must reflect the `for` keyword position
    // ================================================================
    {
      code: 'function run() {\n  for (var x in o) {\n    bar();\n  }\n}',
      errors: [{ messageId: 'wrap', line: 2, column: 3 }],
    },

    // ================================================================
    // Block first statement is not an IfStatement
    // ================================================================
    {
      code: 'for (var x in o) { var y = 1; if (x) continue; f(); }',
      errors: [{ messageId: 'wrap' }],
    },
    {
      code: 'for (var x in o) { lbl: if (x) continue; }',
      errors: [{ messageId: 'wrap' }],
    },

    // ================================================================
    // IfStatement with an empty-block consequent — not a continue-guard
    // ================================================================
    {
      code: 'for (var x in o) { if (x) {} else { continue; } f(); }',
      errors: [{ messageId: 'wrap' }],
    },

    // ================================================================
    // Non-If/Empty single-statement bodies
    // ================================================================
    {
      code: 'for (var x in o) switch (x) { case 1: f(); }',
      errors: [{ messageId: 'wrap' }],
    },
    {
      code: 'for (var x in o) do f(); while (false);',
      errors: [{ messageId: 'wrap' }],
    },

    // ================================================================
    // 3-level nested for-in where only the deepest is unguarded
    // ================================================================
    {
      code: 'for (var a in x) { if (a) continue; for (var b in y) { if (b) continue; for (var c in z) { g(); } } }',
      errors: [{ messageId: 'wrap', line: 1, column: 73 }],
    },
  ],
});
