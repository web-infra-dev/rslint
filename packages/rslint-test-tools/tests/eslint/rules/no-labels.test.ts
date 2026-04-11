import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

// NOTE: JS diagnostics are sorted by position (line, column ascending),
// which differs from Go tests where errors follow AST traversal order.

ruleTester.run('no-labels', {
  valid: [
    // Non-label contexts
    'var f = { label: foo() }',
    'while (true) {}',
    'while (true) { break; }',
    'while (true) { continue; }',
    'for (;;) { break; continue; }',
    'do { break; } while (true)',
    'switch (a) { case 0: break; }',

    // allowLoop: all iteration types
    {
      code: 'A: while (a) { break A; }',
      options: { allowLoop: true },
    },
    {
      code: 'A: do { if (b) { break A; } } while (a);',
      options: { allowLoop: true },
    },
    {
      code: 'A: for (;;) { break A; }',
      options: { allowLoop: true },
    },
    {
      code: 'A: for (var x in obj) { break A; }',
      options: { allowLoop: true },
    },
    {
      code: 'A: for (var x of arr) { break A; }',
      options: { allowLoop: true },
    },
    {
      code: 'A: while (a) { continue A; }',
      options: { allowLoop: true },
    },
    {
      code: 'A: for (var a in obj) { for (;;) { switch (a) { case 0: continue A; } } }',
      options: { allowLoop: true },
    },

    // allowSwitch
    {
      code: 'A: switch (a) { case 0: break A; }',
      options: { allowSwitch: true },
    },

    // Both options true
    {
      code: 'A: while (a) { break A; }',
      options: { allowLoop: true, allowSwitch: true },
    },
    {
      code: 'A: switch (a) { case 0: break A; }',
      options: { allowLoop: true, allowSwitch: true },
    },
    {
      code: 'A: while (a) { switch (x) { case 0: break A; continue A; } }',
      options: { allowLoop: true, allowSwitch: true },
    },
    {
      code: 'A: for (;;) { B: while (a) { break A; continue A; break B; continue B; } }',
      options: { allowLoop: true, allowSwitch: true },
    },
    {
      code: 'A: while (true) { B: while (true) { break A; break B; continue A; continue B; } }',
      options: { allowLoop: true },
    },
  ],
  invalid: [
    // ================================================================
    // Dimension 1: All body types — default options
    // ================================================================
    {
      code: 'label: while(true) {}',
      errors: [{ messageId: 'unexpectedLabel' }],
    },
    {
      code: 'A: do {} while (true);',
      errors: [{ messageId: 'unexpectedLabel' }],
    },
    {
      code: 'A: for (;;) {}',
      errors: [{ messageId: 'unexpectedLabel' }],
    },
    {
      code: 'A: for (var x in obj) {}',
      errors: [{ messageId: 'unexpectedLabel' }],
    },
    {
      code: 'A: for (var x of arr) {}',
      errors: [{ messageId: 'unexpectedLabel' }],
    },
    {
      code: 'A: switch (a) { case 0: break; }',
      errors: [{ messageId: 'unexpectedLabel' }],
    },
    {
      code: 'A: {}',
      errors: [{ messageId: 'unexpectedLabel' }],
    },
    {
      code: 'A: if (true) {}',
      errors: [{ messageId: 'unexpectedLabel' }],
    },
    {
      code: 'A: var foo = 0;',
      errors: [{ messageId: 'unexpectedLabel' }],
    },
    {
      code: 'A: foo();',
      errors: [{ messageId: 'unexpectedLabel' }],
    },

    // ================================================================
    // Dimension 2: Label + break/continue — position-sorted order
    // ================================================================
    {
      code: 'label: while (true) { break label; }',
      errors: [
        { messageId: 'unexpectedLabel' },
        { messageId: 'unexpectedLabelInBreak' },
      ],
    },
    {
      code: 'label: while (true) { continue label; }',
      errors: [
        { messageId: 'unexpectedLabel' },
        { messageId: 'unexpectedLabelInContinue' },
      ],
    },
    {
      code: 'A: while (true) { break A; continue A; }',
      errors: [
        { messageId: 'unexpectedLabel' },
        { messageId: 'unexpectedLabelInBreak' },
        { messageId: 'unexpectedLabelInContinue' },
      ],
    },
    {
      code: 'A: break A;',
      errors: [
        { messageId: 'unexpectedLabel' },
        { messageId: 'unexpectedLabelInBreak' },
      ],
    },

    // ================================================================
    // Dimension 3: Nested labels
    // ================================================================
    {
      code: 'A: { if (foo()) { break A; } bar(); };',
      errors: [
        { messageId: 'unexpectedLabel' },
        { messageId: 'unexpectedLabelInBreak' },
      ],
    },
    {
      code: 'A: if (a) { if (foo()) { break A; } bar(); };',
      errors: [
        { messageId: 'unexpectedLabel' },
        { messageId: 'unexpectedLabelInBreak' },
      ],
    },
    {
      code: 'A: switch (a) { case 0: break A; default: break; };',
      errors: [
        { messageId: 'unexpectedLabel' },
        { messageId: 'unexpectedLabelInBreak' },
      ],
    },
    {
      code: 'A: switch (a) { case 0: B: { break A; } default: break; };',
      errors: [
        { messageId: 'unexpectedLabel' },
        { messageId: 'unexpectedLabel' },
        { messageId: 'unexpectedLabelInBreak' },
      ],
    },
    {
      code: 'A: while (true) { B: for (;;) { break A; } }',
      errors: [
        { messageId: 'unexpectedLabel' },
        { messageId: 'unexpectedLabel' },
        { messageId: 'unexpectedLabelInBreak' },
      ],
    },
    {
      code: 'A: while (true) { B: { continue A; } }',
      errors: [
        { messageId: 'unexpectedLabel' },
        { messageId: 'unexpectedLabel' },
        { messageId: 'unexpectedLabelInContinue' },
      ],
    },
    {
      code: 'A: { B: { C: while (true) { break A; } } }',
      errors: [
        { messageId: 'unexpectedLabel' },
        { messageId: 'unexpectedLabel' },
        { messageId: 'unexpectedLabel' },
        { messageId: 'unexpectedLabelInBreak' },
      ],
    },

    // ================================================================
    // Dimension 4: Chained labels
    // ================================================================
    {
      code: 'A: B: while (true) {}',
      errors: [
        { messageId: 'unexpectedLabel' },
        { messageId: 'unexpectedLabel' },
      ],
    },
    {
      code: 'A: B: while (true) { break A; break B; }',
      errors: [
        { messageId: 'unexpectedLabel' },
        { messageId: 'unexpectedLabel' },
        { messageId: 'unexpectedLabelInBreak' },
        { messageId: 'unexpectedLabelInBreak' },
      ],
    },
    {
      code: 'A: B: while (true) { break B; }',
      options: { allowLoop: true },
      errors: [{ messageId: 'unexpectedLabel' }],
    },
    {
      code: 'A: B: while (true) { break A; }',
      options: { allowLoop: true },
      errors: [
        { messageId: 'unexpectedLabel' },
        { messageId: 'unexpectedLabelInBreak' },
      ],
    },
    {
      code: 'A: B: C: while (true) {}',
      errors: [
        { messageId: 'unexpectedLabel' },
        { messageId: 'unexpectedLabel' },
        { messageId: 'unexpectedLabel' },
      ],
    },
    {
      code: 'A: B: switch (a) { case 0: break B; }',
      options: { allowSwitch: true },
      errors: [{ messageId: 'unexpectedLabel' }],
    },

    // ================================================================
    // Dimension 5: Same-name label shadowing
    // ================================================================
    {
      code: 'A: { A: while (true) { break A; } }',
      errors: [
        { messageId: 'unexpectedLabel' },
        { messageId: 'unexpectedLabel' },
        { messageId: 'unexpectedLabelInBreak' },
      ],
    },
    {
      code: 'A: { A: while (true) { break A; } }',
      options: { allowLoop: true },
      errors: [{ messageId: 'unexpectedLabel' }],
    },

    // ================================================================
    // Dimension 6: Option combinations
    // ================================================================
    {
      code: 'A: var foo = 0;',
      options: { allowLoop: true },
      errors: [{ messageId: 'unexpectedLabel' }],
    },
    {
      code: 'A: break A;',
      options: { allowLoop: true },
      errors: [
        { messageId: 'unexpectedLabel' },
        { messageId: 'unexpectedLabelInBreak' },
      ],
    },
    {
      code: 'A: { if (foo()) { break A; } bar(); };',
      options: { allowLoop: true },
      errors: [
        { messageId: 'unexpectedLabel' },
        { messageId: 'unexpectedLabelInBreak' },
      ],
    },
    {
      code: 'A: if (a) { if (foo()) { break A; } bar(); };',
      options: { allowLoop: true },
      errors: [
        { messageId: 'unexpectedLabel' },
        { messageId: 'unexpectedLabelInBreak' },
      ],
    },
    {
      code: 'A: switch (a) { case 0: break A; default: break; };',
      options: { allowLoop: true },
      errors: [
        { messageId: 'unexpectedLabel' },
        { messageId: 'unexpectedLabelInBreak' },
      ],
    },
    {
      code: 'A: var foo = 0;',
      options: { allowSwitch: true },
      errors: [{ messageId: 'unexpectedLabel' }],
    },
    {
      code: 'A: break A;',
      options: { allowSwitch: true },
      errors: [
        { messageId: 'unexpectedLabel' },
        { messageId: 'unexpectedLabelInBreak' },
      ],
    },
    {
      code: 'A: { if (foo()) { break A; } bar(); };',
      options: { allowSwitch: true },
      errors: [
        { messageId: 'unexpectedLabel' },
        { messageId: 'unexpectedLabelInBreak' },
      ],
    },
    {
      code: 'A: if (a) { if (foo()) { break A; } bar(); };',
      options: { allowSwitch: true },
      errors: [
        { messageId: 'unexpectedLabel' },
        { messageId: 'unexpectedLabelInBreak' },
      ],
    },
    {
      code: 'A: while (a) { break A; }',
      options: { allowSwitch: true },
      errors: [
        { messageId: 'unexpectedLabel' },
        { messageId: 'unexpectedLabelInBreak' },
      ],
    },
    {
      code: 'A: do { if (b) { break A; } } while (a);',
      options: { allowSwitch: true },
      errors: [
        { messageId: 'unexpectedLabel' },
        { messageId: 'unexpectedLabelInBreak' },
      ],
    },
    {
      code: 'A: for (var a in obj) { for (;;) { switch (a) { case 0: break A; } } }',
      options: { allowSwitch: true },
      errors: [
        { messageId: 'unexpectedLabel' },
        { messageId: 'unexpectedLabelInBreak' },
      ],
    },
    // Both options true: block/if/var are "other" — never allowed
    {
      code: 'A: { break A; }',
      options: { allowLoop: true, allowSwitch: true },
      errors: [
        { messageId: 'unexpectedLabel' },
        { messageId: 'unexpectedLabelInBreak' },
      ],
    },
    {
      code: 'A: if (true) { break A; }',
      options: { allowLoop: true, allowSwitch: true },
      errors: [
        { messageId: 'unexpectedLabel' },
        { messageId: 'unexpectedLabelInBreak' },
      ],
    },
    {
      code: 'A: var foo = 0;',
      options: { allowLoop: true, allowSwitch: true },
      errors: [{ messageId: 'unexpectedLabel' }],
    },

    // ================================================================
    // Dimension 7: Cross-label break/continue with options
    // ================================================================
    {
      code: 'A: while (true) { B: { break A; } }',
      options: { allowLoop: true },
      errors: [{ messageId: 'unexpectedLabel' }],
    },
    {
      code: 'A: while (true) { B: { continue A; } }',
      options: { allowLoop: true },
      errors: [{ messageId: 'unexpectedLabel' }],
    },
    {
      code: 'A: switch (a) { case 0: B: { break A; } }',
      options: { allowSwitch: true },
      errors: [{ messageId: 'unexpectedLabel' }],
    },

    // ================================================================
    // Dimension 8: Multi-line
    // ================================================================
    {
      code: 'A:\n  while (true) {\n    break A;\n  }',
      errors: [
        { messageId: 'unexpectedLabel' },
        { messageId: 'unexpectedLabelInBreak' },
      ],
    },

    // ================================================================
    // Dimension 9: Sequential labels — scope cleanup
    // ================================================================
    {
      code: 'A: while (true) {} B: for (;;) {}',
      errors: [
        { messageId: 'unexpectedLabel' },
        { messageId: 'unexpectedLabel' },
      ],
    },
    {
      code: 'A: { break A; } B: while (true) {}',
      errors: [
        { messageId: 'unexpectedLabel' },
        { messageId: 'unexpectedLabelInBreak' },
        { messageId: 'unexpectedLabel' },
      ],
    },
    {
      code: 'A: { break A; } B: while (true) { break B; }',
      options: { allowLoop: true },
      errors: [
        { messageId: 'unexpectedLabel' },
        { messageId: 'unexpectedLabelInBreak' },
      ],
    },

    // ================================================================
    // Dimension 10: Multiple break/continue targeting different labels
    // ================================================================
    {
      code: 'A: while (true) { B: while (true) { break A; break B; continue A; continue B; } }',
      errors: [
        { messageId: 'unexpectedLabel' },
        { messageId: 'unexpectedLabel' },
        { messageId: 'unexpectedLabelInBreak' },
        { messageId: 'unexpectedLabelInBreak' },
        { messageId: 'unexpectedLabelInContinue' },
        { messageId: 'unexpectedLabelInContinue' },
      ],
    },

    // ================================================================
    // Dimension 11: Labels inside function/class bodies
    // ================================================================
    {
      code: 'function f() { A: while (true) { break A; } }',
      errors: [
        { messageId: 'unexpectedLabel' },
        { messageId: 'unexpectedLabelInBreak' },
      ],
    },
    {
      code: 'var f = () => { A: while (true) { break A; } }',
      errors: [
        { messageId: 'unexpectedLabel' },
        { messageId: 'unexpectedLabelInBreak' },
      ],
    },
    {
      code: 'class C { method() { A: while (true) { break A; } } }',
      errors: [
        { messageId: 'unexpectedLabel' },
        { messageId: 'unexpectedLabelInBreak' },
      ],
    },

    // ================================================================
    // Dimension 12: Rare body types
    // ================================================================
    {
      code: 'A: function f() {}',
      errors: [{ messageId: 'unexpectedLabel' }],
    },
    {
      code: 'A: class C {}',
      errors: [{ messageId: 'unexpectedLabel' }],
    },
    {
      code: 'A: ;',
      errors: [{ messageId: 'unexpectedLabel' }],
    },
    {
      code: 'A: try {} catch (e) {}',
      errors: [{ messageId: 'unexpectedLabel' }],
    },
  ],
});
