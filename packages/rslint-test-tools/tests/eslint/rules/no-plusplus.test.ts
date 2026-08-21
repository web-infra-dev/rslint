import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-plusplus', {
  valid: [
    'var foo = 0; foo=+1;',

    // With "allowForLoopAfterthoughts" allowed
    {
      code: 'var foo = 0; foo=+1;',
      options: { allowForLoopAfterthoughts: true },
    },
    {
      code: 'for (i = 0; i < l; i++) { console.log(i); }',
      options: { allowForLoopAfterthoughts: true },
    },
    {
      code: 'for (var i = 0, j = i + 1; j < example.length; i++, j++) {}',
      options: { allowForLoopAfterthoughts: true },
    },
    {
      code: 'for (;; i--, foo());',
      options: { allowForLoopAfterthoughts: true },
    },
    {
      code: 'for (;; foo(), --i);',
      options: { allowForLoopAfterthoughts: true },
    },
    {
      code: 'for (;; foo(), ++i, bar);',
      options: { allowForLoopAfterthoughts: true },
    },
    {
      code: 'for (;; i++, (++j, k--));',
      options: { allowForLoopAfterthoughts: true },
    },
    {
      code: 'for (;; foo(), (bar(), i++), baz());',
      options: { allowForLoopAfterthoughts: true },
    },
    {
      code: 'for (;; (--i, j += 2), bar = j + 1);',
      options: { allowForLoopAfterthoughts: true },
    },
    {
      code: 'for (;; a, (i--, (b, ++j, c)), d);',
      options: { allowForLoopAfterthoughts: true },
    },
  ],
  invalid: [
    {
      code: 'var foo = 0; foo++;',
      errors: [
        {
          messageId: 'unexpectedUnaryOp',
          line: 1,
          column: 14,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: 'var foo = 0; foo--;',
      errors: [
        {
          messageId: 'unexpectedUnaryOp',
          line: 1,
          column: 14,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: 'for (i = 0; i < l; i++) { console.log(i); }',
      errors: [
        {
          messageId: 'unexpectedUnaryOp',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: 'for (i = 0; i < l; foo, i++) { console.log(i); }',
      errors: [
        {
          messageId: 'unexpectedUnaryOp',
          line: 1,
          column: 25,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },

    // With "allowForLoopAfterthoughts" allowed
    {
      code: 'var foo = 0; foo++;',
      options: { allowForLoopAfterthoughts: true },
      errors: [
        {
          messageId: 'unexpectedUnaryOp',
          line: 1,
          column: 14,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: 'for (i = 0; i < l; i++) { v++; }',
      options: { allowForLoopAfterthoughts: true },
      errors: [
        {
          messageId: 'unexpectedUnaryOp',
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
    {
      code: 'for (i++;;);',
      options: { allowForLoopAfterthoughts: true },
      errors: [
        {
          messageId: 'unexpectedUnaryOp',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 9,
        },
      ],
    },
    {
      code: 'for (;--i;);',
      options: { allowForLoopAfterthoughts: true },
      errors: [
        {
          messageId: 'unexpectedUnaryOp',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 10,
        },
      ],
    },
    {
      code: 'for (;;) ++i;',
      options: { allowForLoopAfterthoughts: true },
      errors: [
        {
          messageId: 'unexpectedUnaryOp',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 13,
        },
      ],
    },
    {
      code: 'for (;; i = j++);',
      options: { allowForLoopAfterthoughts: true },
      errors: [
        {
          messageId: 'unexpectedUnaryOp',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      code: 'for (;; i++, f(--j));',
      options: { allowForLoopAfterthoughts: true },
      errors: [
        {
          messageId: 'unexpectedUnaryOp',
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: 'for (;; foo + (i++, bar));',
      options: { allowForLoopAfterthoughts: true },
      errors: [
        {
          messageId: 'unexpectedUnaryOp',
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
  ],
});
