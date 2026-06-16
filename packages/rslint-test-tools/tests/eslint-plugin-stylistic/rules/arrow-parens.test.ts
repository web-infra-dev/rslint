/**
 * @fileoverview Tests for arrow-parens
 * @author Jxck
 *
 * Ported verbatim from @stylistic/eslint-plugin v5.10.0:
 *   packages/eslint-plugin/rules/arrow-parens/arrow-parens.test.ts
 *
 * Transformations applied per the porting spec:
 *  - `run({ name, rule, ... })` -> `ruleTester.run('arrow-parens', null as never, { valid, invalid })`
 *  - The `$` unindent template tag is evaluated to its real multi-line string.
 *  - `parserOptions` (ecmaVersion) dropped — rslint always parses at esnext.
 *  - The rule's messageIds (`expectedParens` / `unexpectedParens` /
 *    `expectedParensBlock` / `unexpectedParensInline`) take no `data`, so they map
 *    1:1 to a fixed message; the RuleTester renders them from the plugin's meta.
 *
 * There is no rslint<->upstream gap for this rule: every upstream valid/invalid
 * case is run verbatim through the green `ruleTester.run` above and matches.
 *   - The bare-`<T>` generic-arrow valid cases (`<T>(a) => b`, `<T>() => b`,
 *     `async <T>(a) => b`) route to a `.ts` fixture (the `needsJsx` classifier
 *     matches only real JSX markers — `</Tag`, `/>`, `<>`), where ts-go reads
 *     `<T>` as an unambiguous generic-arrow type parameter (not JSX) and yields
 *     0 diagnostics, same as upstream.
 *   - The upstream Babel/Flow `if (!skipBabel)` valid cases (`(a: T) => a`,
 *     `(a): T => a`, `(a?) => a`) are plain TypeScript — `(a: T)`/`(a): T` are
 *     type annotations and `(a?)` is an optional binding, none Flow-specific — so
 *     ts-go parses them with 0 diagnostics. They are run on a `.ts` fixture
 *     (rslint has no Babel/Flow parser, but none is needed here).
 *
 * No external-fixture (`readFileSync`) cases exist. There are no output-only
 * invalid cases — every upstream invalid pins `errors`.
 */

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('arrow-parens', null as never, {
  valid: [
    // "always" (by default)
    '() => {}',
    '(a) => {}',
    '(a) => a',
    '(a) => {\n}',
    'a.then((foo) => {});',
    'a.then((foo) => { if (true) {}; });',
    'const f = (/* */a) => a + a;',
    'const f = (a/** */) => a + a;',
    'const f = (a//\n) => a + a;',
    'const f = (//\na) => a + a;',
    'const f = (/*\n */a//\n) => a + a;',
    'const f = (/** @type {number} */a/**hello*/) => a + a;',
    { code: 'a.then(async (foo) => { if (true) {}; });' },

    // "always" (explicit)
    { code: '() => {}', options: ['always'] },
    { code: '(a) => {}', options: ['always'] },
    { code: '(a) => a', options: ['always'] },
    { code: '(a) => {\n}', options: ['always'] },
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

    // "as-needed", { "requireForBlockBody": true }
    { code: '() => {}', options: ['as-needed', { requireForBlockBody: true }] },
    { code: 'a => a', options: ['as-needed', { requireForBlockBody: true }] },
    { code: 'a => (a)', options: ['as-needed', { requireForBlockBody: true }] },
    { code: '(a => a)', options: ['as-needed', { requireForBlockBody: true }] },
    { code: '((a => a))', options: ['as-needed', { requireForBlockBody: true }] },
    { code: '([a, b]) => {}', options: ['as-needed', { requireForBlockBody: true }] },
    { code: '([a, b]) => a', options: ['as-needed', { requireForBlockBody: true }] },
    { code: '({ a, b }) => {}', options: ['as-needed', { requireForBlockBody: true }] },
    { code: '({ a, b }) => a + b', options: ['as-needed', { requireForBlockBody: true }] },
    { code: '(a = 10) => {}', options: ['as-needed', { requireForBlockBody: true }] },
    { code: '(...a) => a[0]', options: ['as-needed', { requireForBlockBody: true }] },
    { code: '(a, b) => {}', options: ['as-needed', { requireForBlockBody: true }] },
    { code: 'a => ({})', options: ['as-needed', { requireForBlockBody: true }] },
    { code: 'async a => ({})', options: ['as-needed', { requireForBlockBody: true }] },
    { code: 'async a => a', options: ['as-needed', { requireForBlockBody: true }] },
    {
      code: 'const f = (/** @type {number} */a/**hello*/) => a + a;',
      options: ['as-needed'],
    },
    {
      code: 'const f = (/* */a) => a + a;',
      options: ['as-needed'],
    },
    {
      code: 'const f = (a/** */) => a + a;',
      options: ['as-needed'],
    },
    {
      code: 'const f = (a//\n) => a + a;',
      options: ['as-needed'],
    },
    {
      code: 'const f = (//\na) => a + a;',
      options: ['as-needed'],
    },
    {
      code: 'const f = (/*\n */a//\n) => a + a;',
      options: ['as-needed'],
    },
    {
      code: 'var foo = (a,/**/) => b;',
      options: ['as-needed'],
    },
    {
      code: 'var foo = (a , /**/) => b;',
      options: ['as-needed'],
    },
    {
      code: 'var foo = (a\n,\n/**/) => b;',
      options: ['as-needed'],
    },
    {
      code: 'var foo = (a,//\n) => b;',
      options: ['as-needed'],
    },
    {
      code: 'const i = (a/**/,) => a + a;',
      options: ['as-needed'],
    },
    {
      code: 'const i = (a \n /**/,) => a + a;',
      options: ['as-needed'],
    },
    {
      code: 'var bar = ({/*comment here*/a}) => a',
      options: ['as-needed'],
    },
    {
      code: 'var bar = (/*comment here*/{a}) => a',
      options: ['as-needed'],
    },

    // generics
    // The bare-`<T>` / `<T>()` / `async <T>` forms route to a `.ts` fixture (the
    // `needsJsx` classifier matches only real JSX markers — `</Tag`, `/>`, `<>`),
    // where ts-go reads `<T>` as an unambiguous generic-arrow type parameter, not
    // JSX. So they parse cleanly with 0 diagnostics, same as upstream.
    {
      code: '<T>(a) => b',
      options: ['always'],
    },
    {
      code: '<T>(a) => b',
      options: ['as-needed'],
    },
    {
      code: '<T>(a) => b',
      options: ['as-needed', { requireForBlockBody: true }],
    },
    {
      code: 'async <T>(a) => b',
      options: ['always'],
    },
    {
      code: 'async <T>(a) => b',
      options: ['as-needed'],
    },
    {
      code: 'async <T>(a) => b',
      options: ['as-needed', { requireForBlockBody: true }],
    },
    {
      code: '<T>() => b',
      options: ['always'],
    },
    {
      code: '<T>() => b',
      options: ['as-needed'],
    },
    {
      code: '<T>() => b',
      options: ['as-needed', { requireForBlockBody: true }],
    },
    {
      code: '<T extends A>(a) => b',
      options: ['always'],
    },
    {
      code: '<T extends A>(a) => b',
      options: ['as-needed'],
    },
    {
      code: '<T extends A>(a) => b',
      options: ['as-needed', { requireForBlockBody: true }],
    },
    {
      code: '<T extends (A | B) & C>(a) => b',
      options: ['always'],
    },
    {
      code: '<T extends (A | B) & C>(a) => b',
      options: ['as-needed'],
    },
    {
      code: '<T extends (A | B) & C>(a) => b',
      options: ['as-needed', { requireForBlockBody: true }],
    },

    // type-annotated arrows
    // Ported from upstream's Babel/Flow `if (!skipBabel)` block. rslint has no
    // Babel/Flow parser, but ts-go parses each of these as plain TypeScript with
    // 0 diagnostics: `(a: T)` / `(a): T` are TS type annotations and `(a?)` is a
    // TS optional binding — none is Flow-specific syntax. So they route to a
    // `.ts` fixture and produce the same 0-diagnostic result upstream expects.
    {
      code: '(a: T) => a',
      options: ['always'],
    },
    {
      code: '(a): T => a',
      options: ['always'],
    },
    {
      code: '(a: T) => a',
      options: ['as-needed'],
    },
    {
      code: '(a?) => a',
      options: ['as-needed'],
    },
    {
      code: '(a): T => a',
      options: ['as-needed'],
    },
    {
      code: '(a: T) => a',
      options: ['as-needed', { requireForBlockBody: true }],
    },
    {
      code: '(a): T => a',
      options: ['as-needed', { requireForBlockBody: true }],
    },
  ],
  invalid: [
    // "always" (by default)
    {
      code: 'a => {}',
      output: '(a) => {}',
      errors: [{
        line: 1,
        column: 1,
        endColumn: 2,
        messageId: 'expectedParens',
      }],
    },
    {
      code: 'a => a',
      output: '(a) => a',
      errors: [{
        line: 1,
        column: 1,
        endColumn: 2,
        messageId: 'expectedParens',
      }],
    },
    {
      code: 'a => {\n}',
      output: '(a) => {\n}',
      errors: [{
        line: 1,
        column: 1,
        endColumn: 2,
        messageId: 'expectedParens',
      }],
    },
    {
      code: 'a.then(foo => {});',
      output: 'a.then((foo) => {});',
      errors: [{
        line: 1,
        column: 8,
        endColumn: 11,
        messageId: 'expectedParens',
      }],
    },
    {
      code: 'a.then(foo => a);',
      output: 'a.then((foo) => a);',
      errors: [{
        line: 1,
        column: 8,
        endColumn: 11,
        messageId: 'expectedParens',
      }],
    },
    {
      code: 'a(foo => { if (true) {}; });',
      output: 'a((foo) => { if (true) {}; });',
      errors: [{
        line: 1,
        column: 3,
        endColumn: 6,
        messageId: 'expectedParens',
      }],
    },
    {
      code: 'a(async foo => { if (true) {}; });',
      output: 'a(async (foo) => { if (true) {}; });',
      errors: [{
        line: 1,
        column: 9,
        endColumn: 12,
        messageId: 'expectedParens',
      }],
    },

    // "as-needed"
    {
      code: '(a) => a',
      output: 'a => a',
      options: ['as-needed'],
      errors: [{
        line: 1,
        column: 2,
        endColumn: 3,
        messageId: 'unexpectedParens',
      }],
    },
    {
      code: '(  a  ) => b',
      output: 'a => b',
      options: ['as-needed'],
      errors: [{
        line: 1,
        column: 4,
        endColumn: 5,
        messageId: 'unexpectedParens',
      }],
    },
    {
      code: '(\na\n) => b',
      output: 'a => b',
      options: ['as-needed'],
      errors: [{
        line: 2,
        column: 1,
        endColumn: 2,
        messageId: 'unexpectedParens',
      }],
    },
    {
      code: '(a,) => a',
      output: 'a => a',
      options: ['as-needed'],
      errors: [{
        line: 1,
        column: 2,
        endColumn: 3,
        messageId: 'unexpectedParens',
      }],
    },
    {
      code: 'async (a) => a',
      output: 'async a => a',
      options: ['as-needed'],
      errors: [{
        line: 1,
        column: 8,
        endColumn: 9,
        messageId: 'unexpectedParens',
      }],
    },
    {
      code: 'async(a) => a',
      output: 'async a => a',
      options: ['as-needed'],
      errors: [{
        line: 1,
        column: 7,
        endColumn: 8,
        messageId: 'unexpectedParens',
      }],
    },
    {
      code: 'typeof((a) => {})',
      output: 'typeof(a => {})',
      options: ['as-needed'],
      errors: [{
        line: 1,
        column: 9,
        endColumn: 10,
        messageId: 'unexpectedParens',
      }],
    },
    {
      code: 'function *f() { yield(a) => a; }',
      output: 'function *f() { yield a => a; }',
      options: ['as-needed'],
      errors: [{
        line: 1,
        column: 23,
        endColumn: 24,
        messageId: 'unexpectedParens',
      }],
    },

    // "as-needed", { "requireForBlockBody": true }
    {
      code: 'a => {}',
      output: '(a) => {}',
      options: ['as-needed', { requireForBlockBody: true }],
      errors: [{
        line: 1,
        column: 1,
        endColumn: 2,
        messageId: 'expectedParensBlock',
      }],
    },
    {
      code: '(a) => a',
      output: 'a => a',
      options: ['as-needed', { requireForBlockBody: true }],
      errors: [{
        line: 1,
        column: 2,
        endColumn: 3,
        messageId: 'unexpectedParensInline',
      }],
    },
    {
      code: 'async a => {}',
      output: 'async (a) => {}',
      options: ['as-needed', { requireForBlockBody: true }],
      errors: [{
        line: 1,
        column: 7,
        endColumn: 8,
        messageId: 'expectedParensBlock',
      }],
    },
    {
      code: 'async (a) => a',
      output: 'async a => a',
      options: ['as-needed', { requireForBlockBody: true }],
      errors: [{
        line: 1,
        column: 8,
        endColumn: 9,
        messageId: 'unexpectedParensInline',
      }],
    },
    {
      code: 'async(a) => a',
      output: 'async a => a',
      options: ['as-needed', { requireForBlockBody: true }],
      errors: [{
        line: 1,
        column: 7,
        endColumn: 8,
        messageId: 'unexpectedParensInline',
      }],
    },
    {
      code: 'const f = /** @type {number} */(a)/**hello*/ => a + a;',
      options: ['as-needed'],
      output: 'const f = /** @type {number} */a/**hello*/ => a + a;',
      errors: [{
        line: 1,
        column: 33,
        messageId: 'unexpectedParens',
        endLine: 1,
        endColumn: 34,
      }],
    },
    {
      code: 'const f = //\n(a) => a + a;',
      output: 'const f = //\na => a + a;',
      options: ['as-needed'],
      errors: [{
        line: 2,
        column: 2,
        messageId: 'unexpectedParens',
        endLine: 2,
        endColumn: 3,
      }],
    },
    {
      code: 'var foo = /**/ a => b;',
      output: 'var foo = /**/ (a) => b;',
      errors: [{
        line: 1,
        column: 16,
        messageId: 'expectedParens',
        endLine: 1,
        endColumn: 17,
      }],
    },
    {
      code: 'var bar = a /**/ =>  b;',
      output: 'var bar = (a) /**/ =>  b;',
      errors: [{
        line: 1,
        column: 11,
        messageId: 'expectedParens',
        endLine: 1,
        endColumn: 12,
      }],
    },
    {
      code: 'const foo = a => {};\n\n// comment between \'a\' and an unrelated closing paren\n\nbar();',
      output: 'const foo = (a) => {};\n\n// comment between \'a\' and an unrelated closing paren\n\nbar();',
      errors: [{
        line: 1,
        column: 13,
        messageId: 'expectedParens',
        endLine: 1,
        endColumn: 14,
      }],
    },
  ],
});
