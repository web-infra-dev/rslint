import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-unused-labels', {
  valid: [
    'A: break A;',
    'A: { foo(); break A; bar(); }',
    'A: if (a) { foo(); if (b) break A; bar(); }',
    'A: for (var i = 0; i < 10; ++i) { foo(); if (a) break A; bar(); }',
    'A: for (var i = 0; i < 10; ++i) { foo(); if (a) continue A; bar(); }',
    'A: { B: break B; C: for (var i = 0; i < 10; ++i) { foo(); if (a) break A; if (c) continue C; bar(); } }',
    'A: { var A = 0; console.log(A); break A; console.log(A); }',
  ],
  invalid: [
    { code: 'A: var foo = 0;', errors: [{ messageId: 'unused' }] },
    { code: 'A: { foo(); bar(); }', errors: [{ messageId: 'unused' }] },
    {
      code: 'A: if (a) { foo(); bar(); }',
      errors: [{ messageId: 'unused' }],
    },
    {
      code: 'A: for (var i = 0; i < 10; ++i) { foo(); if (a) break; bar(); }',
      errors: [{ messageId: 'unused' }],
    },
    {
      code: 'A: for (var i = 0; i < 10; ++i) { foo(); if (a) continue; bar(); }',
      errors: [{ messageId: 'unused' }],
    },
    {
      code: 'A: for (var i = 0; i < 10; ++i) { B: break A; }',
      errors: [{ messageId: 'unused' }],
    },
    {
      code: 'A: { var A = 0; console.log(A); }',
      errors: [{ messageId: 'unused' }],
    },
    { code: 'A: /* comment */ foo', errors: [{ messageId: 'unused' }] },
    { code: 'A /* comment */: foo', errors: [{ messageId: 'unused' }] },
    { code: 'A: "use strict"', errors: [{ messageId: 'unused' }] },
    { code: '"use strict"; foo: "bar"', errors: [{ messageId: 'unused' }] },
    { code: 'A: ("use strict")', errors: [{ messageId: 'unused' }] },
    { code: 'A: `use strict`', errors: [{ messageId: 'unused' }] },
    { code: "if (foo) { bar: 'baz' }", errors: [{ messageId: 'unused' }] },
    {
      code: "A: B: 'foo'",
      errors: [{ messageId: 'unused' }, { messageId: 'unused' }],
    },
    {
      code: "A: B: C: 'foo'",
      errors: [
        { messageId: 'unused' },
        { messageId: 'unused' },
        { messageId: 'unused' },
      ],
    },
    {
      code: "A: B: C: D: 'foo'",
      errors: [
        { messageId: 'unused' },
        { messageId: 'unused' },
        { messageId: 'unused' },
        { messageId: 'unused' },
      ],
    },
    {
      code: "A: B: C: D: E: 'foo'",
      errors: [
        { messageId: 'unused' },
        { messageId: 'unused' },
        { messageId: 'unused' },
        { messageId: 'unused' },
        { messageId: 'unused' },
      ],
    },
    { code: 'A: 42', errors: [{ messageId: 'unused' }] },
  ],
});
