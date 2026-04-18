import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-extra-label', {
  valid: [
    // Upstream ESLint valid suite
    'A: break A;',
    'A: { if (a) break A; }',
    'A: { while (b) { break A; } }',
    'A: { switch (b) { case 0: break A; } }',
    'A: while (a) { while (b) { break; } break; }',
    'A: while (a) { while (b) { break A; } }',
    'A: while (a) { while (b) { continue A; } }',
    'A: while (a) { switch (b) { case 0: break A; } }',
    'A: while (a) { switch (b) { case 0: continue A; } }',
    'A: switch (a) { case 0: while (b) { break A; } }',
    'A: switch (a) { case 0: switch (b) { case 0: break A; } }',
    'A: for (;;) { while (b) { break A; } }',
    'A: do { switch (b) { case 0: break A; break; } } while (a);',
    'A: for (a in obj) { while (b) { break A; } }',
    'A: for (a of ary) { switch (b) { case 0: break A; } }',

    // Naked break / continue
    'while (a) { break; }',
    'do { break; } while (a);',
    'for (;;) { break; continue; }',
    'switch (a) { case 0: break; default: break; }',

    // Labels on non-breakable bodies
    'A: if (x) { if (y) break A; }',
    'A: { for (;;) { break A; } }',
    'A: { do { break A; } while (x); }',
    'A: { switch (y) { case 0: break A; } }',

    // Chained labels — only the innermost is redundant; outer labels are valid
    'A: B: while (true) { break A; }',
    'A: B: while (true) { continue A; }',
    'A: B: C: while (true) { break A; continue B; }',

    // Deep nesting
    'A: while (a) { while (b) { while (c) { break A; } } }',
    'A: for (;;) { for (;;) { for (;;) { continue A; } } }',
    'A: while (a) { while (b) { while (c) { while (d) { break A; } } } }',

    // Real-world
    'outer: for (let i = 0; i < 10; i++) { for (let j = 0; j < 10; j++) { if (i === j) { break outer; } } }',
    'scan: for (const t of ary) { switch (t) { case 1: break; case 2: break scan; } }',

    // Mixed nesting
    'A: while (a) { B: { C: for (;;) { break A; } } }',
    'A: while (a) { B: { while (x) { if (y) break B; else continue A; } } }',
    'A: while (a) { B: while (b) { break A; } }',
    'A: for (;;) { B: for (;;) { C: for (;;) { break A; continue B; } } }',

    // for-in / for-of with continue to outer loop
    'A: for (const k in obj) { for (const v of ary) { if (k === v) continue A; } }',
    'A: for (const x of ary) { for (;;) { break A; } }',

    // Labels inside function / arrow / method — scope doesn't falsely match
    'A: while (a) { function f() { while (b) { break; } } }',
    'A: while (a) { const f = () => { while (b) { break; } }; }',
    'A: while (a) { class C { m() { while (b) { break; } } } }',

    // Async / generator
    'A: while (a) { async function f() { while (b) { break; } } }',
    'A: while (a) { function* g() { while (b) { yield 1; break; } } }',

    // TypeScript-specific
    'A: for (const x of (ary as number[])) { for (const y of ary) { if (x > y) break A; } }',
    'function f<T>(xs: T[]) { A: for (const x of xs) { for (const y of xs) { if (x === y) break A; } } }',

    // Empty bodies / no break at all
    'A: while (a) {}',
    'A: {}',
    'A: ;',

    // Multi-byte characters — CJK identifier label with legitimate nested break,
    // and multi-byte content in surrounding strings that don't affect semantics
    '中文: while (a) { while (b) { break 中文; } }',
    'const s = "日本語"; A: while (a) { while (b) { break A; } }',
    'const s = "🚀"; A: while (a) { while (b) { break A; } }',
  ],
  invalid: [
    // Upstream ESLint invalid suite
    { code: 'A: while (a) break A;', errors: [{ messageId: 'unexpected' }] },
    {
      code: 'A: while (a) { B: { continue A; } }',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'X: while (x) { A: while (a) { B: { break A; break B; continue X; } } }',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'A: do { break A; } while (a);',
      errors: [{ messageId: 'unexpected' }],
    },
    { code: 'A: for (;;) { break A; }', errors: [{ messageId: 'unexpected' }] },
    {
      code: 'A: for (a in obj) { break A; }',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'A: for (a of ary) { break A; }',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'A: switch (a) { case 0: break A; }',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'X: while (x) { A: switch (a) { case 0: break A; } }',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'X: switch (a) { case 0: A: while (b) break A; }',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: `
                A: while (true) {
                    break A;
                    while (true) {
                        break A;
                    }
                }
            `,
      errors: [{ messageId: 'unexpected' }],
    },

    // Comments between break/continue and label suppress autofix
    {
      code: 'A: while(true) { /*comment*/break A; }',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'A: while(true) { break/**/ A; }',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'A: while(true) { continue /**/ A; }',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'A: while(true) { break /**/A; }',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'A: while(true) { continue/**/A; }',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'A: while(true) { continue A/*comment*/; }',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'A: while(true) { break A//comment\n }',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'A: while(true) { break A/*comment*/\nfoo() }',
      errors: [{ messageId: 'unexpected' }],
    },

    // Unnecessary continue on every loop variant
    {
      code: 'A: while (a) { continue A; }',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'A: do { continue A; } while (x);',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'A: for (;;) { continue A; }',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'A: for (const k in obj) { continue A; }',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'A: for (const x of ary) { continue A; }',
      errors: [{ messageId: 'unexpected' }],
    },

    // Same-name shadowing
    {
      code: 'A: while (a) { A: while (b) { break A; } }',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'A: { A: while (b) { break A; } }',
      errors: [{ messageId: 'unexpected' }],
    },

    // Chained labels — innermost redundant
    {
      code: 'A: B: while (true) { break B; }',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'A: B: C: while (true) { continue C; }',
      errors: [{ messageId: 'unexpected' }],
    },

    // Sequential labels
    {
      code: 'A: while (a) { break A; } B: while (b) { break B; }',
      errors: [{ messageId: 'unexpected' }, { messageId: 'unexpected' }],
    },

    // Inner unnecessary, outer necessary
    {
      code: 'A: while (a) { B: while (b) { break B; break A; } }',
      errors: [{ messageId: 'unexpected' }],
    },

    // Multiple unnecessary labels in one scope
    {
      code: 'A: while (a) { break A; continue A; }',
      errors: [{ messageId: 'unexpected' }, { messageId: 'unexpected' }],
    },

    // Real-world matrix search with redundant inner label
    {
      code: 'outer: for (let i = 0; i < 10; i++) { inner: for (let j = 0; j < 10; j++) { if (i === j) { break inner; } } }',
      errors: [{ messageId: 'unexpected' }],
    },

    // Labels inside function / arrow / method / async / generator
    {
      code: 'function f() { A: while (a) { break A; } }',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'const f = () => { A: while (a) { break A; } };',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'class C { m() { A: while (a) { break A; } } }',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'async function f() { A: while (a) { break A; } }',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'function* g() { A: while (a) { break A; } }',
      errors: [{ messageId: 'unexpected' }],
    },

    // TypeScript-specific
    {
      code: 'function f<T>(arr: T[]) { A: while (arr.length) { break A; } }',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'A: for (const x of (ary as number[])) { break A; }',
      errors: [{ messageId: 'unexpected' }],
    },

    // Multi-line with comment on keyword-label boundary
    {
      code: 'A: while (a) {\n  break /* note */ A;\n}',
      errors: [{ messageId: 'unexpected' }],
    },

    // Nested labeled switch — inner redundant
    {
      code: 'A: switch (a) { case 0: B: switch (b) { case 0: break B; } }',
      errors: [{ messageId: 'unexpected' }],
    },

    // Labeled do-while
    {
      code: 'A: do { if (x) continue A; } while (y);',
      errors: [{ messageId: 'unexpected' }],
    },

    // Switch case bare break vs break A/B
    {
      code: 'A: while (a) { B: switch (b) { case 0: break B; case 1: break A; } }',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'A: switch (a) { case 0: B: { if (x) break B; break A; } }',
      errors: [{ messageId: 'unexpected' }],
    },

    // Cross-labeled loops — inner direct label on continue redundant
    {
      code: 'A: while (a) { B: while (b) { if (x) break A; else continue B; } }',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'A: for (;;) { B: for (;;) { C: for (;;) { break A; continue B; break C; } } }',
      errors: [{ messageId: 'unexpected' }],
    },

    // Multi-byte characters — verifies UTF-16 column math and autofix byte
    // ranges across CJK identifier labels and multi-byte trivia
    {
      code: '中文: while (a) break 中文;',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'A: while(true) { break/*日本語*/ A; }',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'A: while(true) { /*日本語*/break A; }',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'const s = "🚀"; A: while (a) { break A; }',
      errors: [{ messageId: 'unexpected' }],
    },
  ],
});
