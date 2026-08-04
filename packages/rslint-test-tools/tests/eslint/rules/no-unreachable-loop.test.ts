import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-unreachable-loop', {
  valid: [
    // one body shape per loop kind
    'while (a) { bar(); }',
    'while (a) { continue; }',
    'while (a) { if (foo) break; }',
    'do { bar(); } while (a)',
    'do { continue; } while (a)',
    'for (a; b; c) { bar(); }',
    'for (var i = 0; i < a.length; i++) { if (foo) break; }',
    'for (;;) { continue; }',
    'for (a in b) { bar(); }',
    'for (a of b) { if (foo) break; }',
    'async function foo() { for await (const a of b) { bar(); } }',
    'function someFunc() { while (a) { if (foo) { return; } bar(); } }',
    'while (a) { switch (foo) { case 1: continue; default: return; } throw err; }',
    'while (a) { try { return bar(); } catch (e) {} }',

    // out of scope for the code path analysis and consequently out of scope for this rule
    'while (false) { foo(); }',
    'while (bar) { foo(); if (true) { break; } }',
    'do foo(); while (false)',
    'for (x = 1; x < 10; i++) { if (x > 0) { foo(); throw err; } }',
    'for (x of []);',
    'for (x of [1]);',

    // doesn't report unreachable loop statements, regardless of whether they would be valid or not in a reachable position
    'function foo() { return; while (a); }',
    'function foo() { return; while (a) break; }',
    'while(true); while(true);',
    'while(true); while(true) break;',

    // "ignore"
    {
      code: 'while (a) break;',
      options: [{ ignore: ['WhileStatement'] }] as any,
    },
    {
      code: 'do break; while (a)',
      options: [{ ignore: ['DoWhileStatement'] }] as any,
    },
    {
      code: 'for (a; b; c) break;',
      options: [{ ignore: ['ForStatement'] }] as any,
    },
    {
      code: 'for (a in b) break;',
      options: [{ ignore: ['ForInStatement'] }] as any,
    },
    {
      code: 'for (a of b) break;',
      options: [{ ignore: ['ForOfStatement'] }] as any,
    },
    {
      code: 'for (var key in obj) { hasEnumerableProperties = true; break; } for (const a of b) break;',
      options: [{ ignore: ['ForInStatement', 'ForOfStatement'] }] as any,
    },
  ],
  invalid: [
    // one body shape per loop kind
    {
      code: 'while (a) break;',
      errors: [{ messageId: 'invalid', line: 1, column: 1, endColumn: 17 }],
    },
    {
      code: 'do break; while (a)',
      errors: [{ messageId: 'invalid', line: 1, column: 1, endColumn: 20 }],
    },
    {
      code: 'for (a; b; c) { foo(); break; }',
      errors: [{ messageId: 'invalid', line: 1, column: 1, endColumn: 32 }],
    },
    {
      code: 'for (;;) throw err;',
      errors: [{ messageId: 'invalid' }],
    },
    {
      code: 'for (a in b) break;',
      errors: [{ messageId: 'invalid' }],
    },
    {
      code: 'for (a of b) break;',
      errors: [{ messageId: 'invalid' }],
    },
    {
      code: 'function someFunc() { while (a) { if (foo) { return; } else { break; } bar(); } }',
      errors: [{ messageId: 'invalid' }],
    },
    {
      code: 'while (a) { switch (foo) { case 1: throw err; default: return; } }',
      errors: [{ messageId: 'invalid' }],
    },
    {
      code: 'while (a) { try { return bar(); } catch (e) { break; } }',
      errors: [{ messageId: 'invalid' }],
    },

    // an infinite loop inside means the outer loop cannot have more than one iteration
    {
      code: 'while (a) for (;;);',
      errors: [{ messageId: 'invalid' }],
    },

    // invalid loop nested in a valid loop (valid in valid, and valid in invalid are covered by basic tests)
    {
      code: 'while (foo) { for (a of b) { if (baz) { break; } else { throw err; } } }',
      errors: [{ messageId: 'invalid' }],
    },
    {
      code: 'lbl: for (var i = 0; i < 10; i++) { while (foo) break lbl; } /* outer is valid because inner can have 0 iterations */',
      errors: [{ messageId: 'invalid' }],
    },

    // invalid loop nested in another invalid loop
    {
      code: 'for (a in b) { while (foo) { if(baz) { break; } else { break; } } break; }',
      errors: [{ messageId: 'invalid' }, { messageId: 'invalid' }],
    },

    // loop and nested loop both invalid because of the same exit statement
    {
      code: 'function foo() { for (var i = 0; i < 10; i++) { do { return; } while(i) } }',
      errors: [{ messageId: 'invalid' }, { messageId: 'invalid' }],
    },
    {
      code: 'lbl: while(foo) { do { break lbl; } while(baz) }',
      errors: [{ messageId: 'invalid' }, { messageId: 'invalid' }],
    },

    // inner loop has continue, but to an outer loop
    {
      code: 'lbl: for (a in b) { while(foo) { continue lbl; } }',
      errors: [{ messageId: 'invalid' }],
    },

    // edge cases - inner loop has only one exit path, but at the same time it exits the outer loop in the first iteration
    {
      code: 'for (a of b) { for(;;) { if (foo) { throw err; } } }',
      errors: [{ messageId: 'invalid' }],
    },
    {
      code: 'function foo () { for (a in b) { while (true) { if (bar) { return; } } } }',
      errors: [{ messageId: 'invalid' }],
    },

    // edge cases where parts of the loops belong to the same code path segment, tests for false negatives
    {
      code: 'do for (var i = 1; i < 10; i++) break; while(foo)',
      errors: [{ messageId: 'invalid' }],
    },
    {
      code: 'do { for (var i = 1; i < 10; i++) continue; break; } while(foo)',
      errors: [{ messageId: 'invalid' }],
    },
    {
      code: 'for (;;) { for (var i = 1; i < 10; i ++) break; if (foo) break; continue; }',
      errors: [{ messageId: 'invalid', column: 12 }],
    },

    // "ignore"
    {
      code: 'while (a) break; do break; while (b); for (;;) break; for (c in d) break; for (e of f) break;',
      options: [{ ignore: [] }] as any,
      errors: [
        { messageId: 'invalid' },
        { messageId: 'invalid' },
        { messageId: 'invalid' },
        { messageId: 'invalid' },
        { messageId: 'invalid' },
      ],
    },
    {
      code: 'while (a) break;',
      options: [{ ignore: ['DoWhileStatement'] }] as any,
      errors: [{ messageId: 'invalid' }],
    },
    {
      code: 'for (a in b) break; for (;;) break; for (c of d) break;',
      options: [{ ignore: ['ForInStatement', 'ForOfStatement'] }] as any,
      errors: [{ messageId: 'invalid' }],
    },
  ],
});
