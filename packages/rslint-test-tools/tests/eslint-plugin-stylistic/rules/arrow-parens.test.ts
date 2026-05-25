import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('arrow-parens', null as never, {
  valid: [
    // "always" (by default)
    { code: '() => {}' },
    { code: '(a) => {}' },
    { code: '(a) => a' },
    { code: '(a) => {\n}' },
    { code: 'a.then((foo) => {});' },
    { code: 'a.then((foo) => { if (true) {}; });' },
    { code: 'const f = (/* */a) => a + a;' },
    { code: 'const f = (a/** */) => a + a;' },
    { code: 'const f = (a//\n) => a + a;' },
    { code: 'const f = (//\na) => a + a;' },
    { code: 'const f = (/*\n */a//\n) => a + a;' },
    { code: 'const f = (/** @type {number} */a/**hello*/) => a + a;' },
    { code: 'a.then(async (foo) => { if (true) {}; });' },

    // "always" (explicit)
    { code: '() => {}', options: ['always'] },
    { code: '(a) => {}', options: ['always'] },
    { code: '(a) => a', options: ['always'] },
    { code: 'a.then((foo) => {});', options: ['always'] },
    { code: 'a.then((foo) => { if (true) {}; });', options: ['always'] },
    { code: 'a.then(async (foo) => { if (true) {}; });', options: ['always'] },

    // "as-needed"
    { code: '() => {}', options: ['as-needed'] },
    { code: 'a => {}', options: ['as-needed'] },
    { code: 'a => a', options: ['as-needed'] },
    { code: 'a => (a)', options: ['as-needed'] },
    { code: '(a => a)', options: ['as-needed'] },
    { code: '((a => a))', options: ['as-needed'] },
    { code: '([a, b]) => {}', options: ['as-needed'] },
    { code: '({ a, b }) => {}', options: ['as-needed'] },
    { code: '(a = 10) => {}', options: ['as-needed'] },
    { code: '(...a) => a[0]', options: ['as-needed'] },
    { code: '(a, b) => {}', options: ['as-needed'] },
    { code: 'async a => a', options: ['as-needed'] },
    { code: 'async ([a, b]) => {}', options: ['as-needed'] },
    { code: 'async (a, b) => {}', options: ['as-needed'] },

    // "as-needed", { requireForBlockBody: true }
    { code: '() => {}', options: ['as-needed', { requireForBlockBody: true }] },
    { code: 'a => a', options: ['as-needed', { requireForBlockBody: true }] },
    { code: 'a => (a)', options: ['as-needed', { requireForBlockBody: true }] },
    { code: '(a => a)', options: ['as-needed', { requireForBlockBody: true }] },
    {
      code: '([a, b]) => {}',
      options: ['as-needed', { requireForBlockBody: true }],
    },
    {
      code: '([a, b]) => a',
      options: ['as-needed', { requireForBlockBody: true }],
    },
    {
      code: '({ a, b }) => {}',
      options: ['as-needed', { requireForBlockBody: true }],
    },
    {
      code: '(a = 10) => {}',
      options: ['as-needed', { requireForBlockBody: true }],
    },
    {
      code: '(...a) => a[0]',
      options: ['as-needed', { requireForBlockBody: true }],
    },
    {
      code: '(a, b) => {}',
      options: ['as-needed', { requireForBlockBody: true }],
    },
    {
      code: 'a => ({})',
      options: ['as-needed', { requireForBlockBody: true }],
    },
    {
      code: 'async a => ({})',
      options: ['as-needed', { requireForBlockBody: true }],
    },
    {
      code: 'async a => a',
      options: ['as-needed', { requireForBlockBody: true }],
    },

    // comments inside parens keep them under as-needed
    {
      code: 'const f = (/** @type {number} */a/**hello*/) => a + a;',
      options: ['as-needed'],
    },
    { code: 'const f = (/* */a) => a + a;', options: ['as-needed'] },
    { code: 'const f = (a/** */) => a + a;', options: ['as-needed'] },
    { code: 'const f = (a//\n) => a + a;', options: ['as-needed'] },
    { code: 'const f = (//\na) => a + a;', options: ['as-needed'] },
    { code: 'const f = (/*\n */a//\n) => a + a;', options: ['as-needed'] },
    { code: 'var foo = (a,/**/) => b;', options: ['as-needed'] },
    { code: 'var foo = (a , /**/) => b;', options: ['as-needed'] },
    { code: 'var foo = (a\n,\n/**/) => b;', options: ['as-needed'] },
    { code: 'var foo = (a,//\n) => b;', options: ['as-needed'] },
    { code: 'const i = (a/**/,) => a + a;', options: ['as-needed'] },
    { code: 'const i = (a \n /**/,) => a + a;', options: ['as-needed'] },
    { code: 'var bar = ({/*comment here*/a}) => a', options: ['as-needed'] },
    { code: 'var bar = (/*comment here*/{a}) => a', options: ['as-needed'] },

    // generics — use .ts (not the default .tsx) so `<T>` is parsed as type
    // parameters rather than JSX
    { code: '<T>(a) => b', options: ['always'], filename: 'src/virtual.ts' },
    { code: '<T>(a) => b', options: ['as-needed'], filename: 'src/virtual.ts' },
    {
      code: '<T>(a) => b',
      options: ['as-needed', { requireForBlockBody: true }],
      filename: 'src/virtual.ts',
    },
    {
      code: 'async <T>(a) => b',
      options: ['always'],
      filename: 'src/virtual.ts',
    },
    {
      code: 'async <T>(a) => b',
      options: ['as-needed'],
      filename: 'src/virtual.ts',
    },
    {
      code: 'async <T>(a) => b',
      options: ['as-needed', { requireForBlockBody: true }],
      filename: 'src/virtual.ts',
    },
    { code: '<T>() => b', options: ['always'], filename: 'src/virtual.ts' },
    { code: '<T>() => b', options: ['as-needed'], filename: 'src/virtual.ts' },
    {
      code: '<T>() => b',
      options: ['as-needed', { requireForBlockBody: true }],
      filename: 'src/virtual.ts',
    },
    {
      code: '<T extends A>(a) => b',
      options: ['always'],
      filename: 'src/virtual.ts',
    },
    {
      code: '<T extends A>(a) => b',
      options: ['as-needed'],
      filename: 'src/virtual.ts',
    },
    {
      code: '<T extends A>(a) => b',
      options: ['as-needed', { requireForBlockBody: true }],
      filename: 'src/virtual.ts',
    },
    {
      code: '<T extends (A | B) & C>(a) => b',
      options: ['always'],
      filename: 'src/virtual.ts',
    },
    {
      code: '<T extends (A | B) & C>(a) => b',
      options: ['as-needed'],
      filename: 'src/virtual.ts',
    },
    {
      code: '<T extends (A | B) & C>(a) => b',
      options: ['as-needed', { requireForBlockBody: true }],
      filename: 'src/virtual.ts',
    },
  ],
  invalid: [
    // "always" (by default)
    {
      code: 'a => {}',
      errors: [
        { messageId: 'expectedParens', line: 1, column: 1, endColumn: 2 },
      ],
    },
    {
      code: 'a => a',
      errors: [
        { messageId: 'expectedParens', line: 1, column: 1, endColumn: 2 },
      ],
    },
    {
      code: 'a => {\n}',
      errors: [
        { messageId: 'expectedParens', line: 1, column: 1, endColumn: 2 },
      ],
    },
    {
      code: 'a.then(foo => {});',
      errors: [
        { messageId: 'expectedParens', line: 1, column: 8, endColumn: 11 },
      ],
    },
    {
      code: 'a.then(foo => a);',
      errors: [
        { messageId: 'expectedParens', line: 1, column: 8, endColumn: 11 },
      ],
    },
    {
      code: 'a(foo => { if (true) {}; });',
      errors: [
        { messageId: 'expectedParens', line: 1, column: 3, endColumn: 6 },
      ],
    },
    {
      code: 'a(async foo => { if (true) {}; });',
      errors: [
        { messageId: 'expectedParens', line: 1, column: 9, endColumn: 12 },
      ],
    },

    // "as-needed"
    {
      code: '(a) => a',
      options: ['as-needed'],
      errors: [
        { messageId: 'unexpectedParens', line: 1, column: 2, endColumn: 3 },
      ],
    },
    {
      code: '(  a  ) => b',
      options: ['as-needed'],
      errors: [
        { messageId: 'unexpectedParens', line: 1, column: 4, endColumn: 5 },
      ],
    },
    {
      code: '(\na\n) => b',
      options: ['as-needed'],
      errors: [
        { messageId: 'unexpectedParens', line: 2, column: 1, endColumn: 2 },
      ],
    },
    {
      code: '(a,) => a',
      options: ['as-needed'],
      errors: [
        { messageId: 'unexpectedParens', line: 1, column: 2, endColumn: 3 },
      ],
    },
    {
      code: 'async (a) => a',
      options: ['as-needed'],
      errors: [
        { messageId: 'unexpectedParens', line: 1, column: 8, endColumn: 9 },
      ],
    },
    {
      code: 'async(a) => a',
      options: ['as-needed'],
      errors: [
        { messageId: 'unexpectedParens', line: 1, column: 7, endColumn: 8 },
      ],
    },
    {
      code: 'typeof((a) => {})',
      options: ['as-needed'],
      errors: [
        { messageId: 'unexpectedParens', line: 1, column: 9, endColumn: 10 },
      ],
    },
    {
      code: 'function *f() { yield(a) => a; }',
      options: ['as-needed'],
      errors: [
        { messageId: 'unexpectedParens', line: 1, column: 23, endColumn: 24 },
      ],
    },

    // "as-needed", { requireForBlockBody: true }
    {
      code: 'a => {}',
      options: ['as-needed', { requireForBlockBody: true }],
      errors: [
        { messageId: 'expectedParensBlock', line: 1, column: 1, endColumn: 2 },
      ],
    },
    {
      code: '(a) => a',
      options: ['as-needed', { requireForBlockBody: true }],
      errors: [
        {
          messageId: 'unexpectedParensInline',
          line: 1,
          column: 2,
          endColumn: 3,
        },
      ],
    },
    {
      code: 'async a => {}',
      options: ['as-needed', { requireForBlockBody: true }],
      errors: [
        { messageId: 'expectedParensBlock', line: 1, column: 7, endColumn: 8 },
      ],
    },
    {
      code: 'async (a) => a',
      options: ['as-needed', { requireForBlockBody: true }],
      errors: [
        {
          messageId: 'unexpectedParensInline',
          line: 1,
          column: 8,
          endColumn: 9,
        },
      ],
    },
    {
      code: 'async(a) => a',
      options: ['as-needed', { requireForBlockBody: true }],
      errors: [
        {
          messageId: 'unexpectedParensInline',
          line: 1,
          column: 7,
          endColumn: 8,
        },
      ],
    },
    {
      code: 'const f = /** @type {number} */(a)/**hello*/ => a + a;',
      options: ['as-needed'],
      errors: [
        {
          messageId: 'unexpectedParens',
          line: 1,
          column: 33,
          endLine: 1,
          endColumn: 34,
        },
      ],
    },
    {
      code: 'const f = //\n(a) => a + a;',
      options: ['as-needed'],
      errors: [
        {
          messageId: 'unexpectedParens',
          line: 2,
          column: 2,
          endLine: 2,
          endColumn: 3,
        },
      ],
    },
    {
      code: 'var foo = /**/ a => b;',
      errors: [
        {
          messageId: 'expectedParens',
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      code: 'var bar = a /**/ =>  b;',
      errors: [
        {
          messageId: 'expectedParens',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 12,
        },
      ],
    },
  ],
});
