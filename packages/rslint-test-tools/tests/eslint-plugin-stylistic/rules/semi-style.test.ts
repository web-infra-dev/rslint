/**
 * @fileoverview Tests for semi-style rule.
 * @author Toru Nagashima
 *
 * Ported verbatim from @stylistic/eslint-plugin v5.10.0:
 *   packages/eslint-plugin/rules/semi-style/semi-style.test.ts
 *
 * Transformations applied per the porting spec:
 *  - `run({ name, rule, valid, invalid })` -> `ruleTester.run('semi-style', null as never, { valid, invalid })`
 *  - The `$` unindent template tag is evaluated to its real multi-line string
 *    (common leading indentation stripped, first/last blank lines dropped), then
 *    written as an escaped single-line string literal.
 *  - `parserOptions: { ecmaVersion: 2022 }` dropped — rslint always parses at
 *    esnext, where class fields and `static {}` blocks are accepted.
 *  - The single messageId (`expectedSemiColon`) takes a `{ pos }` data field; the
 *    RuleTester renders `Expected this semicolon to be at {{pos}}.` from the
 *    plugin's meta + the case's `data`.
 *
 * The upstream file is a single `run()` block: no `if (!skipBabel)` block, no
 * Babel/Flow cases, no external-fixture `readFileSync` cases, no `suggestions`.
 * The `._css_` / `._json_` / `._markdown_` test files don't exist for this rule.
 *
 * Two invalid cases pin `output: null` (the autofix is intentionally skipped
 * because a comment sits between the statement and the misplaced semicolon) —
 * these keep an explicit `errors` array, so they are NOT output-only cases; there
 * are NO output-only cases in this file.
 *
 * KNOWN GAPS: none. Every fixture — including the bare `;` statements, empty
 * `do ; while`, `for(a;b;c)` clauses, `switch`/`case` bodies, class fields with
 * leading-semicolon `accessor` members, and `static {}` blocks — parses under
 * rslint's ts-go parser and produces byte-identical diagnostics and autofix
 * output. No octal/`\8` syntax, no `assert` import attributes, no sloppy-mode-only
 * constructs exist for this rule, so nothing is isolated.
 */

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('semi-style', null as never, {
  valid: [
    ';',
    ';foo;bar;baz;',
    'foo;\nbar;',
    'for(a;b;c);',
    'for(a;\nb;\nc);',
    'for((a\n);\n(b\n);\n(c));',
    'if(a)foo;\nbar',
    { code: ';', options: ['last'] },
    { code: ';foo;bar;baz;', options: ['last'] },
    { code: 'foo;\nbar;', options: ['last'] },
    { code: 'for(a;b;c);', options: ['last'] },
    { code: 'for(a;\nb;\nc);', options: ['last'] },
    { code: 'for((a\n);\n(b\n);\n(c));', options: ['last'] },
    { code: 'class C { a; b; }', options: ['last'] },
    { code: 'class C {\na;\nb;\n}', options: ['last'] },
    { code: 'if(a)foo;\nbar', options: ['last'] },
    { code: ';', options: ['first'] },
    { code: ';foo;bar;baz;', options: ['first'] },
    { code: 'foo\n;bar;', options: ['first'] },
    { code: 'for(a;b;c);', options: ['first'] },
    { code: 'for(a;\nb;\nc);', options: ['first'] },
    { code: 'for((a\n);\n(b\n);\n(c));', options: ['first'] },
    { code: 'class C { a ;b }', options: ['first'] },
    { code: 'class C {\na\n;b\n}', options: ['first'] },

    // edge cases
    {
      code: '{\n    ;\n}',
      options: ['first'],
    },
    {
      code: 'while (a)\n    ;\nfoo',
      options: ['first'],
    },
    {
      code: 'do\n    ;\nwhile (a)',
      options: ['first'],
    },
    {
      code: 'do\n    foo;\nwhile (a)',
      options: ['first'],
    },
    {
      code: 'if (a)\n    foo;\nelse\n    bar',
      options: ['first'],
    },
    {
      code: 'if (a)\n    foo\n;bar',
      options: ['first'],
    },
    {
      code: '{\n    ;\n}',
      options: ['last'],
    },
    {
      code: 'switch (a) {\n    case 1:\n        ;foo\n}',
      options: ['last'],
    },
    {
      code: 'while (a)\n    ;\nfoo',
      options: ['last'],
    },
    {
      code: 'do\n    ;\nwhile (a)',
      options: ['last'],
    },

    {
      code: 'class C {\n  ;foo\n  ;accessor bar\n}',
      options: ['first'],
    },
    {
      code: 'class C {\n  foo;\n  accessor bar;\n}',
      options: ['last'],
    },

    // Class static blocks
    {
      code: 'class C {\n    static {}\n}',
      options: ['last'],
    },
    {
      code: 'class C {\n    static {\n        foo\n    }\n}',
      options: ['last'],
    },
    {
      code: 'class C {\n    static {\n        foo\n        bar\n    }\n}',
      options: ['last'],
    },
    {
      code: 'class C {\n    static {\n        ;\n    }\n}',
      options: ['last'],
    },
    {
      code: 'class C {\n    static {\n        foo;\n    }\n}',
      options: ['last'],
    },
    {
      code: 'class C {\n    static {\n        foo;\n        bar;\n    }\n}',
      options: ['last'],
    },
    {
      code: 'class C {\n    static {\n        foo;\n        bar;\n        baz;\n    }\n}',
      options: ['last'],
    },
    {
      code: 'class C {\n    static {}\n}',
      options: ['first'],
    },
    {
      code: 'class C {\n    static {\n        foo\n    }\n}',
      options: ['first'],
    },
    {
      code: 'class C {\n    static {\n        foo\n        bar\n    }\n}',
      options: ['first'],
    },
    {
      code: 'class C {\n    static {\n        ;\n    }\n}',
      options: ['first'],
    },
    {
      code: 'class C {\n    static {\n        ;foo\n    }\n}',
      options: ['first'],
    },
    {
      code: 'class C {\n    static {\n        foo;\n    }\n}',
      options: ['first'],
    },
    {
      code: 'class C {\n    static {\n        foo\n        ;bar\n    }\n}',
      options: ['first'],
    },
    {
      code: 'class C {\n    static {\n        foo\n        ;bar;\n    }\n}',
      options: ['first'],
    },
    {
      code: 'class C {\n    static {\n        foo\n        ;bar\n        ;baz\n    }\n}',
      options: ['first'],
    },
    {
      code: 'class C {\n    static {\n        foo\n        ;bar\n        ;baz;\n    }\n}',
      options: ['first'],
    },
  ],
  invalid: [
    {
      code: 'foo\n;bar',
      output: 'foo;\nbar',
      errors: [{
        messageId: 'expectedSemiColon',
        data: {
          pos: 'the end of the previous line',
        },
      }],
    },
    {
      code: 'if(a)foo\n;bar',
      output: 'if(a)foo;\nbar',
      errors: [{
        messageId: 'expectedSemiColon',
        data: {
          pos: 'the end of the previous line',
        },
      }],
    },
    {
      code: 'var foo\n;bar',
      output: 'var foo;\nbar',
      errors: [{
        messageId: 'expectedSemiColon',
        data: {
          pos: 'the end of the previous line',
        },
      }],
    },
    {
      code: 'foo\n;\nbar',
      output: 'foo;\nbar',
      errors: [{
        messageId: 'expectedSemiColon',
        data: {
          pos: 'the end of the previous line',
        },
      }],
    },
    {
      code: 'for(a\n;b;c)d',
      output: 'for(a;\nb;c)d',
      errors: [{
        messageId: 'expectedSemiColon',
        data: {
          pos: 'the end of the previous line',
        },
      }],
    },
    {
      code: 'for(a;b\n;c)d',
      output: 'for(a;b;\nc)d',
      errors: [{
        messageId: 'expectedSemiColon',
        data: {
          pos: 'the end of the previous line',
        },
      }],
    },
    {
      code: 'do;while(a)\n;b',
      output: 'do;while(a);\nb',
      errors: [{
        messageId: 'expectedSemiColon',
        data: {
          pos: 'the end of the previous line',
        },
      }],
    },

    {
      code: 'foo\n;bar',
      output: 'foo;\nbar',
      options: ['last'],
      errors: [{
        messageId: 'expectedSemiColon',
        data: {
          pos: 'the end of the previous line',
        },
      }],
    },
    {
      code: 'if(a)foo\n;bar',
      output: 'if(a)foo;\nbar',
      options: ['last'],
      errors: [{
        messageId: 'expectedSemiColon',
        data: {
          pos: 'the end of the previous line',
        },
      }],
    },
    {
      code: 'var foo\n;bar',
      output: 'var foo;\nbar',
      options: ['last'],
      errors: [{
        messageId: 'expectedSemiColon',
        data: {
          pos: 'the end of the previous line',
        },
      }],
    },
    {
      code: 'foo\n;\nbar',
      output: 'foo;\nbar',
      options: ['last'],
      errors: [{
        messageId: 'expectedSemiColon',
        data: {
          pos: 'the end of the previous line',
        },
      }],
    },
    {
      code: 'for(a\n;b;c)d',
      output: 'for(a;\nb;c)d',
      options: ['last'],
      errors: [{
        messageId: 'expectedSemiColon',
        data: {
          pos: 'the end of the previous line',
        },
      }],
    },
    {
      code: 'for(a;b\n;c)d',
      output: 'for(a;b;\nc)d',
      options: ['last'],
      errors: [{
        messageId: 'expectedSemiColon',
        data: {
          pos: 'the end of the previous line',
        },
      }],
    },
    {
      code: 'foo()\n;',
      output: 'foo();\n',
      options: ['last'],
      errors: [{
        messageId: 'expectedSemiColon',
        data: {
          pos: 'the end of the previous line',
        },
      }],
    },

    {
      code: 'foo;\nbar',
      output: 'foo\n;bar',
      options: ['first'],
      errors: [{
        messageId: 'expectedSemiColon',
        data: {
          pos: 'the beginning of the next line',
        },
      }],
    },
    {
      code: 'if(a)foo;\nbar',
      output: 'if(a)foo\n;bar',
      options: ['first'],
      errors: [{
        messageId: 'expectedSemiColon',
        data: {
          pos: 'the beginning of the next line',
        },
      }],
    },
    {
      code: 'var foo;\nbar',
      output: 'var foo\n;bar',
      options: ['first'],
      errors: [{
        messageId: 'expectedSemiColon',
        data: {
          pos: 'the beginning of the next line',
        },
      }],
    },
    {
      code: 'foo\n;\nbar',
      output: 'foo\n;bar',
      options: ['first'],
      errors: [{
        messageId: 'expectedSemiColon',
        data: {
          pos: 'the beginning of the next line',
        },
      }],
    },
    {
      code: 'for(a\n;b;c)d',
      output: 'for(a;\nb;c)d',
      options: ['first'],
      errors: [{
        messageId: 'expectedSemiColon',
        data: {
          pos: 'the end of the previous line',
        },
      }],
    },
    {
      code: 'for(a;b\n;c)d',
      output: 'for(a;b;\nc)d',
      options: ['first'],
      errors: [{
        messageId: 'expectedSemiColon',
        data: {
          pos: 'the end of the previous line',
        },
      }],
    },

    {
      code: 'foo\n;/**/bar',
      output: null,
      errors: [{
        messageId: 'expectedSemiColon',
        data: {
          pos: 'the end of the previous line',
        },
      }],
    },
    {
      code: 'foo\n/**/;bar',
      output: null,
      errors: [{
        messageId: 'expectedSemiColon',
        data: {
          pos: 'the end of the previous line',
        },
      }],
    },

    {
      code: 'foo;\n/**/bar',
      output: null,
      options: ['first'],
      errors: [{
        messageId: 'expectedSemiColon',
        data: {
          pos: 'the beginning of the next line',
        },
      }],
    },
    {
      code: 'foo/**/;\nbar',
      output: null,
      options: ['first'],
      errors: [{
        messageId: 'expectedSemiColon',
        data: {
          pos: 'the beginning of the next line',
        },
      }],
    },

    // Class fields
    {
      code: 'class C { foo\n;bar }',
      output: 'class C { foo;\nbar }',
      options: ['last'],
      errors: [{
        messageId: 'expectedSemiColon',
        data: {
          pos: 'the end of the previous line',
        },
      }],
    },
    {
      code: 'class C { foo;\nbar }',
      output: 'class C { foo\n;bar }',
      options: ['first'],
      errors: [{
        messageId: 'expectedSemiColon',
        data: {
          pos: 'the beginning of the next line',
        },
      }],
    },
    {
      code: 'class C {\n  ;accessor foo\n  ;accessor bar\n}',
      output: 'class C {\n  ;accessor foo;\naccessor bar\n}',
      options: ['last'],
      errors: [{
        messageId: 'expectedSemiColon',
        data: { pos: 'the end of the previous line' },
      }],
    },
    {
      code: 'class C {\n  accessor foo;\n  accessor bar\n}',
      output: 'class C {\n  accessor foo\n;accessor bar\n}',
      options: ['first'],
      errors: [{
        messageId: 'expectedSemiColon',
        data: { pos: 'the beginning of the next line' },
      }],
    },

    // Class static blocks
    {
      code: 'class C { static { foo\n; } }',
      output: 'class C { static { foo;\n} }',
      options: ['last'],
      errors: [{
        messageId: 'expectedSemiColon',
        data: {
          pos: 'the end of the previous line',
        },
      }],
    },
    {
      code: 'class C { static { foo\n ;bar } }',
      output: 'class C { static { foo;\nbar } }',
      options: ['last'],
      errors: [{
        messageId: 'expectedSemiColon',
        data: {
          pos: 'the end of the previous line',
        },
      }],
    },
    {
      code: 'class C { static { foo;\nbar\n ; } }',
      output: 'class C { static { foo;\nbar;\n} }',
      options: ['last'],
      errors: [{
        messageId: 'expectedSemiColon',
        data: {
          pos: 'the end of the previous line',
        },
      }],
    },
    {
      code: 'class C { static { foo;\nbar } }',
      output: 'class C { static { foo\n;bar } }',
      options: ['first'],
      errors: [{
        messageId: 'expectedSemiColon',
        data: {
          pos: 'the beginning of the next line',
        },
      }],
    },
  ],
});
