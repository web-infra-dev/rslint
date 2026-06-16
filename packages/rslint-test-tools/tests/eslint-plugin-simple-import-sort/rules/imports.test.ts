/**
 * @fileoverview Alignment tests for eslint-plugin-simple-import-sort 'imports' rule.
 *
 * Ported verbatim from eslint-plugin-simple-import-sort v13.0.0:
 *   test/imports.test.js
 *
 * Upstream structure: three suites driven by ESLint's RuleTester —
 *   - `baseTests`        run 3x there (JS / Flow / TS parsers); the cases are
 *                       identical, so they are ported here ONCE.
 *   - `typescriptTests` TypeScript-specific cases — merged into run().
 *   - `flowTests`        Flow-specific cases — those that parse under ts-go are
 *                       merged into run(); the rest are isolated (see KNOWN GAPS).
 *
 * Transformations applied per the porting spec:
 *  - The upstream `input\`…\`` template tag is evaluated to its real multi-line
 *    string (strip 10-space + `|` prefix; drop the leading/trailing newline).
 *  - Each invalid case's `output: (actual) => expect(actual).toMatchInlineSnapshot(\`…\`)`
 *    is evaluated to the real expected fixed source (pipes stripped; the helper's
 *    visible-char hacks decoded: `→` → TAB, `<CR>` → CR).
 *  - Strings are emitted as escaped literals for byte-exact fidelity (trailing
 *    spaces / tabs / CR / unicode are load-bearing).
 *  - `parserOptions` dropped (rslint resolves via tsconfig); `options` kept verbatim.
 *  - The single message id is `sort` ("Run autofix to sort these imports!"),
 *    no data interpolation.
 *
 * Every upstream invalid case pins BOTH `output` (the fixed source) AND `errors`
 * (a count, or — for one positional case — `{ messageId, line, column, endLine,
 * endColumn }`). None are output-only.
 *
 * Cases that surface a real rslint<->upstream gap are NOT deleted or altered: they
 * are listed in the KNOWN GAPS block at the bottom, each annotated with the upstream
 * expectation vs. what rslint does.
 */

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('imports', null as never, {
  valid: [
    // Simple cases.
    "import \"a\"",
    "import a from \"a\"",
    "import {a} from \"a\"",
    "import a, {b} from \"a\"",
    "import {a,b} from \"a\"",
    "import {} from \"a\"",
    "import {    } from \"a\"",
    "import * as a from \"a\"",
    // Side-effect only imports are kept in the original order.
    "import \"b\";\nimport \"a\"",
    // Side-effect only imports use a stable sort (issue #34).
    "import \"codemirror/addon/fold/brace-fold\"\nimport \"codemirror/addon/edit/closebrackets\"\nimport \"codemirror/addon/fold/foldgutter\"\nimport \"codemirror/addon/fold/foldgutter.css\"\nimport \"codemirror/addon/lint/json-lint\"\nimport \"codemirror/addon/lint/lint\"\nimport \"codemirror/addon/lint/lint.css\"\nimport \"codemirror/addon/scroll/simplescrollbars\"\nimport \"codemirror/addon/scroll/simplescrollbars.css\"\nimport \"codemirror/lib/codemirror.css\"\nimport \"codemirror/mode/javascript/javascript\"",
    // Sorted alphabetically.
    "import x1 from \"a\";\nimport x2 from \"b\"",
    // Opt-out.
    "// eslint-disable-next-line\nimport x2 from \"b\"\nimport x1 from \"a\";",
    // Whitespace before comment at last specifier should stay.
    "import {\n  a, // a\n  b // b\n} from \"specifiers-comment-space\"\nimport {\n  c, // c\n  d, // d\n} from \"specifiers-comment-space-2\"",
    // Accidental trailing spaces doesn’t produce a sorting error.
    "import a from \"a\"    \nimport b from \"b\";    \nimport c from \"c\";  /* comment */  ",
    // Simple cases.
    "import type a from \"a\"",
    "import type {a} from \"a\"",
    "import type {} from \"a\"",
    "import type {    } from \"a\"",
    // type specifiers.
    "import { a, type b, c, type d } from \"a\"",
    // Sorted alphabetically.
    "import type x1 from \"a\";\nimport type x2 from \"b\"",
    // Simple cases.
    "import type a from \"a\"",
    "import type {a} from \"a\"",
    "import type a, {b} from \"a\"",
    "import type {} from \"a\"",
    "import type {    } from \"a\"",
    "import json from \"./foo.json\" with { type: \"json\" };",
    // type specifiers.
    "import { a, type b, c, type d } from \"a\"",
  ],

  invalid: [
    // Sorting alphabetically.
    {
      code: "import x2 from \"b\"\nimport x1 from \"a\";",
      output: "import x1 from \"a\";\nimport x2 from \"b\"",
      errors: 1,
    },
    // Semicolon-free code style, with start-of-line guarding semicolon.
    {
      code: "import x2 from \"b\"\nimport x1 from \"a\"\n\n;[].forEach()",
      output: "import x1 from \"a\"\nimport x2 from \"b\"\n\n;[].forEach()",
      errors: 1,
    },
    // Semicolon-free code style 2.
    {
      code: "import { foo } from \"bar\"\nimport a from \"a\"\n\n;(async function() {\n  await foo()\n})()",
      output: "import a from \"a\"\nimport { foo } from \"bar\"\n\n;(async function() {\n  await foo()\n})()",
      errors: 1,
    },
    // Semicolons edge cases.
    {
      code: "import x2 from \"b\"\nimport x7 from \"g\";\nimport x6 from \"f\"\n;import x5 from \"e\"\nimport x4 from \"d\" ; import x3 from \"c\"\nimport x1 from \"a\" ; [].forEach()",
      output: "import x1 from \"a\" ; \nimport x2 from \"b\"\nimport x3 from \"c\"\nimport x4 from \"d\" ; \nimport x5 from \"e\"\nimport x6 from \"f\"\n;\nimport x7 from \"g\";[].forEach()",
      errors: 1,
    },
    // Comments around start-of-line guarding semicolon.
    {
      code: "import x2 from \"b\"\nimport x1 from \"a\" // a\n\n;/* comment */[].forEach()",
      output: "import x1 from \"a\" // a\nimport x2 from \"b\"\n\n;/* comment */[].forEach()",
      errors: 1,
    },
    // No more code after last semicolon.
    {
      code: "import x2 from \"b\"\nimport x1 from \"a\"\n\n;",
      output: "import x1 from \"a\"\n;\nimport x2 from \"b\"",
      errors: 1,
    },
    // Sorting specifiers.
    {
      code: "import { e, b, a as c } from \"specifiers\"",
      output: "import { a as c,b, e } from \"specifiers\"",
      errors: 1,
    },
    // Sorting specifiers with default import.
    {
      code: "import d, { e, b, a as c } from \"specifiers-default\"",
      output: "import d, { a as c,b, e } from \"specifiers-default\"",
      errors: 1,
    },
    // Sorting specifiers with trailing comma.
    {
      code: "import d, { e, b, a as c, } from \"specifiers-trailing-comma\"",
      output: "import d, { a as c,b, e,  } from \"specifiers-trailing-comma\"",
      errors: 1,
    },
    // Sorting specifiers with renames.
    {
      code: "import { a as c, a as b2, b, a } from \"specifiers-renames\"",
      output: "import { a,a as b2, a as c, b } from \"specifiers-renames\"",
      errors: 1,
    },
    // Sorting specifiers like humans do.
    {
      code: "import {\n  B,\n  a,\n  A,\n  b,\n  B2,\n  bb,\n  BB,\n  bB,\n  Bb,\n  ab,\n  ba,\n  Ba,\n  BA,\n  bA,\n  x as d,\n  x as C,\n  img10,\n  img2,\n  img1,\n  img10_black,\n} from \"specifiers-human-sort\"",
      output: "import {\n  A,\n  a,\n  ab,\n  B,\n  b,\n  B2,\n  BA,\n  Ba,\n  bA,\n  ba,\n  BB,\n  Bb,\n  bB,\n  bb,\n  img1,\n  img2,\n  img10,\n  img10_black,\n  x as C,\n  x as d,\n} from \"specifiers-human-sort\"",
      errors: 1,
    },
    // Keyword-like specifiers.
    {
      code: "import { aaNotKeyword, zzNotKeyword, abstract, as, asserts, any, async, /*await,*/ boolean, constructor, declare, get, infer, is, keyof, module, namespace, never, readonly, require, number, object, set, string, symbol, type, undefined, unique, unknown, from, global, bigint, of } from 'keyword-identifiers';",
      output: "import { aaNotKeyword, abstract, any, as, asserts, async, /*await,*/ bigint, boolean, constructor, declare, from, get, global, infer, is, keyof, module, namespace, never, number, object, of,readonly, require, set, string, symbol, type, undefined, unique, unknown, zzNotKeyword } from 'keyword-identifiers';",
      errors: 1,
    },
    // No spaces in specifiers.
    {
      code: "import {e,b,a as c} from \"specifiers-no-spaces\"",
      output: "import {a as c,b,e} from \"specifiers-no-spaces\"",
      errors: 1,
    },
    // Space before specifiers.
    {
      code: "import { b,a} from \"specifiers-no-space-before\"",
      output: "import { a,b} from \"specifiers-no-space-before\"",
      errors: 1,
    },
    // Space after specifiers.
    {
      code: "import {b,a } from \"specifiers-no-space-after\"",
      output: "import {a,b } from \"specifiers-no-space-after\"",
      errors: 1,
    },
    // Space after specifiers.
    {
      code: "import {b,a, } from \"specifiers-no-space-after-trailing\"",
      output: "import {a,b, } from \"specifiers-no-space-after-trailing\"",
      errors: 1,
    },
    // Sorting specifiers with comments.
    {
      code: "import {\n  // c\n  c,\n  b, // b\n  a\n  // last\n} from \"specifiers-comments\"",
      output: "import {\n  a,\n  b, // b\n  // c\n  c\n  // last\n} from \"specifiers-comments\"",
      errors: 1,
    },
    // Comment after last specifier should stay last.
    {
      code: "import {\n  // c\n  c,b, // b\n  a\n  // last\n} from \"specifiers-comments-last\"",
      output: "import {\n  a,\nb, // b\n  // c\n  c\n  // last\n} from \"specifiers-comments-last\"",
      errors: 1,
    },
    // Sorting specifiers with comment between.
    {
      code: "import { b /* b */, a } from \"specifiers-comment-between\"",
      output: "import { a,b /* b */ } from \"specifiers-comment-between\"",
      errors: 1,
    },
    // Sorting specifiers with trailing comma and trailing comments.
    {
      code: "import {\n  c,\n  a,\n  // x\n  // y\n} from \"specifiers-trailing\"",
      output: "import {\n  a,\n  c,\n  // x\n  // y\n} from \"specifiers-trailing\"",
      errors: 1,
    },
    // Sorting specifiers with multiline comments.
    {
      code: "import {\n  /*c1*/ c, /*c2*/ /*a1\n  */a, /*a2*/ /*\n  after */\n  // x\n  // y\n} from \"specifiers-multiline-comments\"",
      output: "import {\n/*a1\n  */a, /*a2*/ \n  /*c1*/ c, /*c2*/ /*\n  after */\n  // x\n  // y\n} from \"specifiers-multiline-comments\"",
      errors: 1,
    },
    // Sorting specifiers with multiline end comment.
    {
      code: "import {\n  b,\n  a /*\n  after */\n} from \"specifiers-multiline-end-comment\"",
      output: "import {\n  a,   b/*\n  after */\n} from \"specifiers-multiline-end-comment\"",
      errors: 1,
    },
    // Sorting specifiers with multiline end comment after newline.
    {
      code: "import {\n  b,\n  a /*a*/\n  /*\n  after */\n} from \"specifiers-multiline-end-comment-after-newline\"",
      output: "import {\n  a, /*a*/\n  b  /*\n  after */\n} from \"specifiers-multiline-end-comment-after-newline\"",
      errors: 1,
    },
    // Sorting specifiers with multiline end comment and no newline.
    {
      code: "import {\n  b,\n  a /*\n  after */ } from \"specifiers-multiline-end-comment-no-newline\"",
      output: "import {\n  a,   b/*\n  after */ } from \"specifiers-multiline-end-comment-no-newline\"",
      errors: 1,
    },
    // Sorting specifiers with lots of comments.
    {
      code: "/*1*//*2*/import/*3*/def,/*4*/{/*{*/e/*e1*/,/*e2*//*e3*/b/*b1*/,/*b2*/a/*a1*/as/*a2*/c/*a3*/,/*a4*/}/*5*/from/*6*/\"specifiers-lots-of-comments\"/*7*//*8*/",
      output: "/*1*//*2*/import/*3*/def,/*4*/{/*{*/a/*a1*/as/*a2*/c/*a3*/,/*a4*/b/*b1*/,/*b2*/e/*e1*/,/*e2*//*e3*/}/*5*/from/*6*/\"specifiers-lots-of-comments\"/*7*//*8*/",
      errors: 1,
    },
    // Sorting specifiers with lots of comments, multiline.
    {
      code: "import { // start\n  /* c1 */ c /* c2 */, // c3\n  // b1\n\n  b as /* b2 */ renamed\n  , /* b3 */ /* a1\n  */ a /* not-a\n  */ // comment at end\n} from \"specifiers-lots-of-comments-multiline\";\nimport {\n  e,\n  d, /* d */ /* not-d\n  */ // comment at end after trailing comma\n} from \"specifiers-lots-of-comments-multiline-2\";",
      output: "import { // start\n/* a1\n  */ a, \n  // b1\n  b as /* b2 */ renamed\n  , /* b3 */ \n  /* c1 */ c /* c2 */// c3\n/* not-a\n  */ // comment at end\n} from \"specifiers-lots-of-comments-multiline\";\nimport {\n  d, /* d */   e,\n/* not-d\n  */ // comment at end after trailing comma\n} from \"specifiers-lots-of-comments-multiline-2\";",
      errors: 1,
    },
    // No empty line after last specifier due to newline before comma.
    {
      code: "import {\n  b/*b*/\n  ,\n  a\n} from \"specifiers-blank\";",
      output: "import {\n  a,\n  b/*b*/\n  } from \"specifiers-blank\";",
      errors: 1,
    },
    // Sorting both inline and multiline specifiers.
    {
      code: "import {z, y,\n  x,\n\nw,\n    v\n  as /*v*/\n\n    u , t /*t*/, // t\n    s\n} from \"specifiers-inline-multiline\"",
      output: "import {    s,\nt /*t*/, // t\n    v\n  as /*v*/\n    u , w,\n  x,\ny,\nz} from \"specifiers-inline-multiline\"",
      errors: 1,
    },
    // Indent: 0.
    {
      code: "import {\nb,\na,\n} from \"specifiers-indent-0\"",
      output: "import {\na,\nb,\n} from \"specifiers-indent-0\"",
      errors: 1,
    },
    // Indent: 4.
    {
      code: "import {\n    b,\n    a,\n} from \"specifiers-indent-4\"",
      output: "import {\n    a,\n    b,\n} from \"specifiers-indent-4\"",
      errors: 1,
    },
    // Indent: tab.
    {
      code: "import {\n\tb,\n\ta,\n} from \"specifiers-indent-tab\"",
      output: "import {\n\ta,\n\tb,\n} from \"specifiers-indent-tab\"",
      errors: 1,
    },
    // Indent: mixed.
    {
      code: "import {\n //\n\tb,\n  a,\n\n    c,\n} from \"specifiers-indent-mixed\"",
      output: "import {\n  a,\n //\n\tb,\n    c,\n} from \"specifiers-indent-mixed\"",
      errors: 1,
    },
    // Several chunks.
    {
      code: "require(\"c\");\n\nimport x1 from \"b\"\nimport x2 from \"a\"\nrequire(\"c\");\n\nimport x3 from \"b\"\nimport x4 from \"a\" // x4\n\n// c1\nrequire(\"c\");\nimport x5 from \"b\"\n// x6-1\nimport x6 from \"a\" /* after\n*/\n\nrequire(\"c\"); import x7 from \"b\"; import x8 from \"a\"; require(\"c\")",
      output: "require(\"c\");\n\nimport x2 from \"a\"\nimport x1 from \"b\"\nrequire(\"c\");\n\nimport x4 from \"a\" // x4\nimport x3 from \"b\"\n\n// c1\nrequire(\"c\");\n// x6-1\nimport x6 from \"a\" \nimport x5 from \"b\"/* after\n*/\n\nrequire(\"c\"); import x8 from \"a\"; \nimport x7 from \"b\"; require(\"c\")",
      errors: 4,
    },
    // Original order is preserved for duplicate imports with the same style.
    {
      code: "import b from \"b\"\nimport a1 from \"a\"\nimport a2 from \"a\"",
      output: "import a1 from \"a\"\nimport a2 from \"a\"\nimport b from \"b\"",
      errors: 1,
    },
    // Original order is preserved for duplicate imports with the same style (reversed).
    {
      code: "import b from \"b\"\nimport a2 from \"a\"\nimport a1 from \"a\"",
      output: "import a2 from \"a\"\nimport a1 from \"a\"\nimport b from \"b\"",
      errors: 1,
    },
    // Deterministic order for imports of the same module with different styles.
    // Order: namespace < default < named
    {
      code: "import b from \"b\"\nimport {a2} from \"a\"\nimport a1 from \"a\"",
      output: "import a1 from \"a\"\nimport {a2} from \"a\"\nimport b from \"b\"",
      errors: 1,
    },
    // Deterministic order for all import styles of the same module.
    // Order: namespace < default < named
    // Note: For simplicity, we don’t try to detect default+named.
    // There is no use case for importing the default import multiple times with different names.
    // So the three default imports below stay in their internal original order.
    {
      code: "import {x} from \"a\"\nimport {} from \"a\"\nimport w1, {} from \"a\"\nimport w2 from \"a\"\nimport w3, {z} from \"a\"\nimport * as v from \"a\"",
      output: "import * as v from \"a\"\nimport w1, {} from \"a\"\nimport w2 from \"a\"\nimport w3, {z} from \"a\"\nimport {x} from \"a\"\nimport {} from \"a\"",
      errors: 1,
    },
    // Special characters sorting order.
    {
      code: "import {} from \"\";\nimport {} from \".\";\nimport {} from \".//\";\nimport {} from \"./\";\nimport {} from \"./B\"; // B1\nimport {} from \"./b\";\nimport {} from \"./B\"; // B2\nimport {} from \"./A\";\nimport {} from \"./a\";\nimport {} from \"./_a\";\nimport {} from \"./-a\";\nimport {} from \"./[id]\";\nimport {} from \"./,\";\nimport {} from \"./ä\";\nimport {} from \"./ä\"; // “a” followed by “̈̈” (COMBINING DIAERESIS).\nimport {} from \"..\";\nimport {} from \"../\";\nimport {} from \"../a\";\nimport {} from \"../_a\";\nimport {} from \"../-a\";\nimport {} from \"../[id]\";\nimport {} from \"../,\";\nimport {} from \"../a/..\";\nimport {} from \"../a/../\";\nimport {} from \"../a/...\";\nimport {} from \"../a/../b\";\nimport {} from \"../../\";\nimport {} from \"../..\";\nimport {} from \"../../a\";\nimport {} from \"../../_a\";\nimport {} from \"../../-a\";\nimport {} from \"../../[id]\";\nimport {} from \"../../,\";\nimport {} from \"../../utils\";\nimport {} from \"../../..\";\nimport {} from \"../../../\";\nimport {} from \"../../../a\";\nimport {} from \"../../../_a\";\nimport {} from \"../../../[id]\";\nimport {} from \"../../../,\";\nimport {} from \"../../../utils\";\nimport {} from \"...\";\nimport {} from \".../\";\nimport {} from \".a\";\nimport {} from \"/\";\nimport {} from \"/a\";\nimport {} from \"/a/b\";\nimport {} from \"https://example.com/script.js\";\nimport {} from \"http://example.com/script.js\";\nimport {} from \"react\";\nimport {} from \"async\";\nimport {} from \"./a/-\";\nimport {} from \"./a/.\";\nimport {} from \"./a/0\";\nimport {} from \"@/components/error.vue\"\nimport {} from \"@/components/Alert\"\nimport {} from \"~/test\"\nimport {} from \"#/test\"\nimport {} from \"fs\";\nimport {} from \"fs/something\";\nimport {} from \"node:fs/something\";\nimport {} from \"node:fs\";\nimport {} from \"Fs\";\nimport {} from \"lodash/fp\";\nimport {} from \"@storybook/react\";\nimport {} from \"@storybook/react/something\";\nimport {} from \"1\";\nimport {} from \"1*\";\nimport {} from \"a*\";\nimport img2 from \"./img2\";\nimport img10 from \"./img10\";\nimport img1 from \"./img1\";",
      output: "import {} from \"node:fs\";\nimport {} from \"node:fs/something\";\n\nimport {} from \"@storybook/react\";\nimport {} from \"@storybook/react/something\";\nimport {} from \"1\";\nimport {} from \"1*\";\nimport {} from \"a*\";\nimport {} from \"async\";\nimport {} from \"Fs\";\nimport {} from \"fs\";\nimport {} from \"fs/something\";\nimport {} from \"http://example.com/script.js\";\nimport {} from \"https://example.com/script.js\";\nimport {} from \"lodash/fp\";\nimport {} from \"react\";\n\nimport {} from \"\";\nimport {} from \"/\";\nimport {} from \"/a\";\nimport {} from \"/a/b\";\nimport {} from \"@/components/Alert\"\nimport {} from \"@/components/error.vue\"\nimport {} from \"#/test\"\nimport {} from \"~/test\"\n\nimport {} from \"...\";\nimport {} from \".../\";\nimport {} from \"../../..\";\nimport {} from \"../../../\";\nimport {} from \"../../../,\";\nimport {} from \"../../../_a\";\nimport {} from \"../../../[id]\";\nimport {} from \"../../../a\";\nimport {} from \"../../../utils\";\nimport {} from \"../..\";\nimport {} from \"../../\";\nimport {} from \"../../,\";\nimport {} from \"../../_a\";\nimport {} from \"../../[id]\";\nimport {} from \"../../-a\";\nimport {} from \"../../a\";\nimport {} from \"../../utils\";\nimport {} from \"..\";\nimport {} from \"../\";\nimport {} from \"../,\";\nimport {} from \"../_a\";\nimport {} from \"../[id]\";\nimport {} from \"../-a\";\nimport {} from \"../a\";\nimport {} from \"../a/..\";\nimport {} from \"../a/...\";\nimport {} from \"../a/../\";\nimport {} from \"../a/../b\";\nimport {} from \".//\";\nimport {} from \".\";\nimport {} from \"./\";\nimport {} from \"./,\";\nimport {} from \"./_a\";\nimport {} from \"./[id]\";\nimport {} from \"./-a\";\nimport {} from \"./A\";\nimport {} from \"./a\";\nimport {} from \"./ä\"; // “a” followed by “̈̈” (COMBINING DIAERESIS).\nimport {} from \"./ä\";\nimport {} from \"./a/.\";\nimport {} from \"./a/-\";\nimport {} from \"./a/0\";\nimport {} from \"./B\"; // B1\nimport {} from \"./B\"; // B2\nimport {} from \"./b\";\nimport img1 from \"./img1\";\nimport img2 from \"./img2\";\nimport img10 from \"./img10\";\nimport {} from \".a\";",
      errors: 1,
    },
    // Comments.
    {
      code: "// before\n\n/* also\nbefore */ /* b */ import b from \"b\" // b\n// above d\n  import d /*d1*/ from   \"d\" ; /* d2 */ /* before\n  c0 */ // before c1\n  /* c0\n*/ /*c1*/ /*c2*/import c from 'c' ; /*c3*/ import a from \"a\" /*a*/ /*\n   x1 */ /* x2 */",
      output: "// before\n\n/* also\nbefore */ import a from \"a\" /*a*/ \n/* b */ import b from \"b\" // b\n/* before\n  c0 */ // before c1\n  /* c0\n*/ /*c1*/ /*c2*/import c from 'c' ; /*c3*/ \n// above d\n  import d /*d1*/ from   \"d\" ; /* d2 */ /*\n   x1 */ /* x2 */",
      errors: 1,
    },
    // Line comment and code after.
    {
      code: "import b from \"b\"; // b\nimport a from \"a\"; code();",
      output: "import a from \"a\"; \nimport b from \"b\"; // b\ncode();",
      errors: 1,
    },
    // Line comment and multiline block comment after.
    {
      code: "import b from \"b\"; // b\nimport a from \"a\"; /*\nafter */",
      output: "import a from \"a\"; \nimport b from \"b\"; // b\n/*\nafter */",
      errors: 1,
    },
    // Line comment but _singleline_ block comment after.
    {
      code: "import b from \"b\"; // b\nimport a from \"a\"; /* a */",
      output: "import a from \"a\"; /* a */\nimport b from \"b\"; // b",
      errors: 1,
    },
    // Test messageId, lines and columns.
    {
      code: "// before\n/* also\nbefore */ import b from \"b\";\nimport a from \"a\"; /*a*/ /* comment\nafter */ // after",
      output: "// before\n/* also\nbefore */ import a from \"a\"; /*a*/ \nimport b from \"b\";/* comment\nafter */ // after",
      errors: [
        { messageId: "sort", line: 3, column: 11, endLine: 4, endColumn: 26 },
      ],
    },
    // Collapse blank lines between comments.
    {
      code: "import c from \"c\"\n// b1\n\n// b2\nimport b from \"b\"\n// a\n\nimport a from \"a\"",
      output: "// a\nimport a from \"a\"\n// b1\n// b2\nimport b from \"b\"\nimport c from \"c\"",
      errors: 1,
    },
    // Collapse blank lines between comments – CR.
    {
      code: "import c from \"c\"\r\n// b1\r\n\r\n// b2\r\nimport b from \"b\"\r\n// a\r\n\r\nimport a from \"a\"\r\nafter();\r",
      output: "// a\r\nimport a from \"a\"\r\n// b1\r\n// b2\r\nimport b from \"b\"\r\nimport c from \"c\"\r\nafter();\r",
      errors: 1,
    },
    // Collapse blank lines inside import statements.
    {
      code: "import\n\n// import\n\ndef /* default */\n\n,\n\n// default\n\n {\n\n  // c\n\n  c /*c*/,\n\n  /* b\n   */\n\n  b // b\n  ,\n\n  // a1\n\n  // a2\n\n  a\n\n  // a3\n\n  as\n\n  // a4\n\n  d\n\n  // a5\n\n  , // a6\n\n  // last\n\n}\n\n// from1\n\nfrom\n\n// from2\n\n\"c\"\n\n// final\n\n;",
      output: "import\n// import\ndef /* default */\n,\n// default\n {\n  // a1\n  // a2\n  a\n  // a3\n  as\n  // a4\n  d\n  // a5\n  , // a6\n  /* b\n   */\n  b // b\n  ,\n  // c\n  c /*c*/,\n  // last\n}\n// from1\nfrom\n// from2\n\"c\"\n// final\n;",
      errors: 1,
    },
    // Collapse blank lines inside empty specifier list.
    {
      code: "import {\n\n    } from \"specifiers-empty\"",
      output: "import {\n    } from \"specifiers-empty\"",
      errors: 1,
    },
    // Single-line comment at the end of the last specifier should not comment
    // out the `from` part.
    {
      code: "import {\n\n  b // b\n  ,a} from \"specifiers-line-comment\"",
      output: "import {\na,  b // b\n  } from \"specifiers-line-comment\"",
      errors: 1,
    },
    // Preserve indentation (for `<script>` tags).
    {
      code: "  import e from \"e\"\n  // b\n  import {\n    b4, b3,\n    b2\n  } from \"b\";\n  /* a */ import a from \"a\"; import c from \"c\"\n  \n    // d\n    import d from \"d\"",
      output: "  /* a */ import a from \"a\"; \n  // b\n  import {\n    b2,\nb3,\n    b4  } from \"b\";\nimport c from \"c\"\n    // d\n    import d from \"d\"\n  import e from \"e\"",
      errors: 1,
    },
    // Preserve indentation (for `<script>` tags) – CR.
    {
      code: "      \r\n  import e from \"e\"\r\n  // b\r\n  import {\r\n    b4, b3,\r\n    b2\r\n  } from \"b\";\r\n  /* a */ import a from \"a\"; import c from \"c\"\r\n \r\n    // d\r\n    import d from \"d\"\r\n",
      output: "      \r\n  /* a */ import a from \"a\"; \r\n  // b\r\n  import {\r\n    b2,\r\nb3,\r\n    b4  } from \"b\";\r\nimport c from \"c\"\r\n    // d\r\n    import d from \"d\"\r\n  import e from \"e\"\r\n",
      errors: 1,
    },
    // Trailing spaces.
    {
      code: "import c from \"c\";  /* comment */  \nimport b from \"b\";    \nimport d from \"d\";  /* multiline\ncomment */  \nimport a from \"a\"    \nimport e from \"e\"; /* multiline\ncomment 2 */ import f from \"f\";",
      output: "/* multiline\ncomment */  \nimport a from \"a\"    \nimport b from \"b\";    \nimport c from \"c\";  /* comment */  \nimport d from \"d\";  \nimport e from \"e\"; \n/* multiline\ncomment 2 */ import f from \"f\";",
      errors: 1,
    },
    // Sort like IntelliJ/WebStorm (case insensitive on `from`).
    // https://github.com/lydell/eslint-plugin-simple-import-sort/issues/7#issuecomment-500593886
    {
      code: "import FloatingActionButton from 'src/components/FloatingActionButton'\nimport { Select, linkButton, buttonPrimary, spinnerOverlay } from 'src/components/common'\nimport { icon, spinner } from 'src/components/icons'\nimport { notify } from 'src/components/Notifications'\nimport { IGVBrowser } from 'src/components/IGVBrowser'\nimport { IGVFileSelector } from 'src/components/IGVFileSelector'\nimport { DelayedSearchInput, TextInput } from 'src/components/input'\nimport DataTable from 'src/components/DataTable'\nimport { FlexTable, SimpleTable, HeaderCell, TextCell } from 'src/components/table'\nimport Modal from 'src/components/Modal'\nimport ExportDataModal from 'src/components/ExportDataModal'",
      output: "import { buttonPrimary, linkButton, Select, spinnerOverlay } from 'src/components/common'\nimport DataTable from 'src/components/DataTable'\nimport ExportDataModal from 'src/components/ExportDataModal'\nimport FloatingActionButton from 'src/components/FloatingActionButton'\nimport { icon, spinner } from 'src/components/icons'\nimport { IGVBrowser } from 'src/components/IGVBrowser'\nimport { IGVFileSelector } from 'src/components/IGVFileSelector'\nimport { DelayedSearchInput, TextInput } from 'src/components/input'\nimport Modal from 'src/components/Modal'\nimport { notify } from 'src/components/Notifications'\nimport { FlexTable, HeaderCell, SimpleTable, TextCell } from 'src/components/table'",
      errors: 1,
    },
    // https://github.com/gothinkster/react-redux-realworld-example-app/blob/b5557d1fd40afebe023e3102ad6ef50475146506/src/components/App.js#L1-L16
    {
      code: "import agent from '../agent';\nimport Header from './Header';\nimport React from 'react';\nimport { connect } from 'react-redux';\nimport { APP_LOAD, REDIRECT } from '../constants/actionTypes';\nimport { Route, Switch } from 'react-router-dom';\nimport Article from '../components/Article';\nimport Editor from '../components/Editor';\nimport Home from '../components/Home';\nimport Login from '../components/Login';\nimport Profile from '../components/Profile';\nimport ProfileFavorites from '../components/ProfileFavorites';\nimport Register from '../components/Register';\nimport Settings from '../components/Settings';\nimport { store } from '../store';\nimport { push } from 'react-router-redux';",
      output: "import React from 'react';\nimport { connect } from 'react-redux';\nimport { Route, Switch } from 'react-router-dom';\nimport { push } from 'react-router-redux';\n\nimport agent from '../agent';\nimport Article from '../components/Article';\nimport Editor from '../components/Editor';\nimport Home from '../components/Home';\nimport Login from '../components/Login';\nimport Profile from '../components/Profile';\nimport ProfileFavorites from '../components/ProfileFavorites';\nimport Register from '../components/Register';\nimport Settings from '../components/Settings';\nimport { APP_LOAD, REDIRECT } from '../constants/actionTypes';\nimport { store } from '../store';\nimport Header from './Header';",
      errors: 1,
    },
    // https://github.com/gothinkster/react-redux-realworld-example-app/blob/b5557d1fd40afebe023e3102ad6ef50475146506/src/components/Editor.js#L1-L12
    {
      code: "import ListErrors from './ListErrors';\nimport React from 'react';\nimport agent from '../agent';\nimport { connect } from 'react-redux';\nimport {\n  ADD_TAG,\n  EDITOR_PAGE_LOADED,\n  REMOVE_TAG,\n  ARTICLE_SUBMITTED,\n  EDITOR_PAGE_UNLOADED,\n  UPDATE_FIELD_EDITOR\n} from '../constants/actionTypes';",
      output: "import React from 'react';\nimport { connect } from 'react-redux';\n\nimport agent from '../agent';\nimport {\n  ADD_TAG,\n  ARTICLE_SUBMITTED,\n  EDITOR_PAGE_LOADED,\n  EDITOR_PAGE_UNLOADED,\n  REMOVE_TAG,\n  UPDATE_FIELD_EDITOR\n} from '../constants/actionTypes';\nimport ListErrors from './ListErrors';",
      errors: 1,
    },
    // `groups` – `u` flag.
    {
      code: "import b from '.';\nimport a from 'ä';",
      options: [{ groups: [["^\\p{L}"], ["^\\."]] }],
      output: "import a from 'ä';\n\nimport b from '.';",
      errors: 1,
    },
    // `groups` – non-matching imports end up last.
    {
      code: "import c from '';\nimport b from '.';\nimport a from 'a';\nimport d from '@/a';",
      options: [{ groups: [["^\\w"], ["^\\."]] }],
      output: "import a from 'a';\n\nimport b from '.';\n\nimport c from '';\nimport d from '@/a';",
      errors: 1,
    },
    // `groups` – first longest match wins.
    {
      code: "import c from './';\nimport b from 'bx';\nimport a from 'a';",
      options: [{ groups: [["^\\w"], ["^\\w{2}"], ["^.{2}"]] }],
      output: "import a from 'a';\n\nimport b from 'bx';\n\nimport c from './';",
      errors: 1,
    },
    // `groups` – side effect imports.
    {
      code: "import '@/';\nimport c from '@/';\nimport b from './';\nimport './';\nimport a from 'a';\nimport 'a';\nimport {} from 'a';",
      options: [{ groups: [["^\\w"], ["^\\."], ["^\\u0000"]] }],
      output: "import a from 'a';\nimport {} from 'a';\n\nimport b from './';\n\nimport '@/';\nimport './';\nimport 'a';\n\nimport c from '@/';",
      errors: 1,
    },
    // `groups` – side effect imports keep internal order but are sorted otherwise.
    {
      code: "import b from 'b';\nimport 'c';\nimport d from 'd';\nimport 'a';\nimport '.';\nimport x from './x';",
      options: [{ groups: [] }],
      output: "import 'c';\nimport 'a';\nimport '.';\nimport x from './x';\nimport b from 'b';\nimport d from 'd';",
      errors: 1,
    },
    // `groups` – no line breaks between inner array items.
    {
      code: "import react from 'react';\nimport a from 'a';\nimport webpack from \"webpack\"\nimport Select from 'react-select';\nimport App from './App';",
      options: [{ groups: [["^\\w", "^react"], ["^\\."]] }],
      output: "import a from 'a';\nimport webpack from \"webpack\"\nimport react from 'react';\nimport Select from 'react-select';\n\nimport App from './App';",
      errors: 1,
    },
    // Type imports.
    {
      code: "import React from \"react\";\nimport Button from \"../Button\";\nimport type {target, type as tipe, Button} from \"../Button\";\nimport {a, type type as type, z} from \"../type\";\n\nimport styles from \"./styles.css\";\nimport { getUser } from \"../../api\";\n\nimport PropTypes from \"prop-types\";\nimport { /* X */ } from \"prop-types\";\nimport classnames from \"classnames\";\nimport { truncate, formatNumber } from \"../../utils\";\nimport type X from \"../Button\";\n\nfunction pluck<T, K extends keyof T>(o: T, names: K[]): T[K][] {\n  return names.map(n => o[n]);\n}",
      output: "import classnames from \"classnames\";\nimport PropTypes from \"prop-types\";\nimport { /* X */ } from \"prop-types\";\nimport React from \"react\";\n\nimport { getUser } from \"../../api\";\nimport { formatNumber,truncate } from \"../../utils\";\nimport type X from \"../Button\";\nimport type {Button,target, type as tipe} from \"../Button\";\nimport Button from \"../Button\";\nimport {a, type type as type, z} from \"../type\";\nimport styles from \"./styles.css\";\n\nfunction pluck<T, K extends keyof T>(o: T, names: K[]): T[K][] {\n  return names.map(n => o[n]);\n}",
      errors: 1,
    },
    // `groups` – type imports.
    {
      code: "import '@/';\nimport c from '@/';\nimport type C from '@/';\nimport b from './';\nimport type B from './';\nimport './';\nimport a from 'a';\nimport type A from 'a';\nimport 'a';\nimport {} from 'a';",
      options: [{ groups: [["^\\w"], ["^\\."], ["^.*\\u0000$"]] }],
      output: "import a from 'a';\nimport {} from 'a';\n\nimport b from './';\n\nimport type B from './';\nimport type C from '@/';\nimport type A from 'a';\n\nimport '@/';\nimport './';\nimport 'a';\nimport c from '@/';",
      errors: 1,
    },
    // `groups` – real-world example from issue #61 based on:
    // https://github.com/polkadot-js/apps/blob/074b245a725873f7c3c16fc83b80fb9c02351a65/packages/apps/src/Endpoints/Group.tsx
    // https://github.com/polkadot-js/apps/blob/074b245a725873f7c3c16fc83b80fb9c02351a65/.eslintrc.js#L31-L39
    {
      code: "import React, { useCallback } from 'react';\nimport styled from 'styled-components';\n\nimport { Icon } from '@polkadot/react-components';\nimport type { ThemeProps } from '@polkadot/react-components/types';\n\nimport Network from './Network';\nimport type { Group } from './types';",
      options: [
        {
          groups: [
            ["^[^@.].*\\u0000$", "^[^/.]"],
            ["^@polkadot.*\\u0000$", "^@polkadot"],
            ["^\\..*\\u0000$", "^\\."],
          ],
        },
      ],
      output: "import React, { useCallback } from 'react';\nimport styled from 'styled-components';\n\nimport type { ThemeProps } from '@polkadot/react-components/types';\nimport { Icon } from '@polkadot/react-components';\n\nimport type { Group } from './types';\nimport Network from './Network';",
      errors: 1,
    },
    // Import attributes.
    {
      code: "import json from \"./foo.json\" with { type: \"json\" };\nimport {b, a} from \"./bar.json\" with {\n  // json\n  type: \"json\",\n  a: \"b\",\n} /* bar */ /* end\n comment */\n;[].forEach()",
      output: "import {a,b} from \"./bar.json\" with {\n  // json\n  type: \"json\",\n  a: \"b\",\n} /* bar */ \nimport json from \"./foo.json\" with { type: \"json\" };/* end\n comment */\n;[].forEach()",
      errors: 1,
    },
    // Imports inside module declarations.
    {
      code: "import type { ParsedPath } from 'path';\nimport type { CopyOptions } from 'fs';\n\ndeclare module 'my-module' {\n  import type { PlatformPath, ParsedPath } from 'path';\n  import { type CopyOptions } from 'fs';\n  export function normalize(p: string): string;\n  // comment\n    import \"d\"\nimport c from \"c\"; /*\n  */\timport * as b from \"b\"; // b\n}",
      output: "import type { CopyOptions } from 'fs';\nimport type { ParsedPath } from 'path';\n\ndeclare module 'my-module' {\n  import { type CopyOptions } from 'fs';\n  import type { ParsedPath,PlatformPath } from 'path';\n  export function normalize(p: string): string;\n  // comment\n    import \"d\"\n\n/*\n  */\timport * as b from \"b\"; // b\nimport c from \"c\"; \n}",
      errors: 3,
    },
    // Deterministic order for type imports from the same source.
    // Order: type namespace < type default < type named
    {
      code: "import type {X} from \"a\"\nimport type Y from \"a\"\nimport type * as Z from \"a\"",
      output: "import type * as Z from \"a\"\nimport type Y from \"a\"\nimport type {X} from \"a\"",
      errors: 1,
    },
    // Deterministic order: type imports before value imports, then by style.
    // Note: For simplicity, we don’t try to detect default+named.
    // There is no use case for importing the default import multiple times with different names.
    // So the default imports below stay in their internal original order.
    {
      code: "import {x} from \"a\"\nimport type {X} from \"a\"\nimport type {} from \"a\"\nimport y1, {} from \"a\"\nimport y2 from \"a\"\nimport y3, {l} from \"a\"\nimport type Y1, {} from \"a\"\nimport type Y2 from \"a\"\nimport type Y3, {L} from \"a\"\nimport * as z from \"a\"\nimport type * as Z from \"a\"",
      output: "import type * as Z from \"a\"\nimport type Y1, {} from \"a\"\nimport type Y2 from \"a\"\nimport type Y3, {L} from \"a\"\nimport type {X} from \"a\"\nimport type {} from \"a\"\nimport * as z from \"a\"\nimport y1, {} from \"a\"\nimport y2 from \"a\"\nimport y3, {l} from \"a\"\nimport {x} from \"a\"",
      errors: 1,
    },
    // Import attributes.
    {
      code: "import json from \"./foo.json\" with { type: \"json\" };\nimport {default as b} from \"./bar.json\" with {\n  // json\n  type: \"json\",\n  a: \"b\",\n} /* bar */ /* end\n comment */\n;[].forEach()",
      output: "import {default as b} from \"./bar.json\" with {\n  // json\n  type: \"json\",\n  a: \"b\",\n} /* bar */ \nimport json from \"./foo.json\" with { type: \"json\" };/* end\n comment */\n;[].forEach()",
      errors: 1,
    },
    // https://github.com/graphql/graphql-js/blob/64b194c6c9b9aaa1c139f1b7c3692a6ef851928e/src/execution/execute.js#L10-L69
    {
      code: "import { forEach, isCollection } from 'iterall';\nimport { GraphQLError } from '../error/GraphQLError';\nimport { locatedError } from '../error/locatedError';\nimport inspect from '../jsutils/inspect';\nimport invariant from '../jsutils/invariant';\nimport isInvalid from '../jsutils/isInvalid';\nimport isNullish from '../jsutils/isNullish';\nimport isPromise from '../jsutils/isPromise';\nimport memoize3 from '../jsutils/memoize3';\nimport promiseForObject from '../jsutils/promiseForObject';\nimport promiseReduce from '../jsutils/promiseReduce';\nimport type { ObjMap } from '../jsutils/ObjMap';\nimport type { MaybePromise } from '../jsutils/MaybePromise';\n\nimport { getOperationRootType } from '../utilities/getOperationRootType';\nimport { typeFromAST } from '../utilities/typeFromAST';\nimport { Kind } from '../language/kinds';\nimport {\n  getVariableValues,\n  getArgumentValues,\n  getDirectiveValues,\n} from './values';\nimport {\n  isObjectType,\n  isAbstractType,\n  isLeafType,\n  isListType,\n  isNonNullType,\n} from '../type/definition';\nimport type {\n  GraphQLObjectType,\n  GraphQLOutputType,\n  GraphQLLeafType,\n  GraphQLAbstractType,\n  GraphQLField,\n  GraphQLFieldResolver,\n  GraphQLResolveInfo,\n  ResponsePath,\n  GraphQLList,\n} from '../type/definition';\nimport type { GraphQLSchema } from '../type/schema';\nimport {\n  SchemaMetaFieldDef,\n  TypeMetaFieldDef,\n  TypeNameMetaFieldDef,\n} from '../type/introspection';\nimport {\n  GraphQLIncludeDirective,\n  GraphQLSkipDirective,\n} from '../type/directives';\nimport { assertValidSchema } from '../type/validate';\nimport type {\n  DocumentNode,\n  OperationDefinitionNode,\n  SelectionSetNode,\n  FieldNode,\n  FragmentSpreadNode,\n  InlineFragmentNode,\n  FragmentDefinitionNode,\n} from '../language/ast';",
      output: "import { forEach, isCollection } from 'iterall';\n\nimport { GraphQLError } from '../error/GraphQLError';\nimport { locatedError } from '../error/locatedError';\nimport inspect from '../jsutils/inspect';\nimport invariant from '../jsutils/invariant';\nimport isInvalid from '../jsutils/isInvalid';\nimport isNullish from '../jsutils/isNullish';\nimport isPromise from '../jsutils/isPromise';\nimport type { MaybePromise } from '../jsutils/MaybePromise';\nimport memoize3 from '../jsutils/memoize3';\nimport type { ObjMap } from '../jsutils/ObjMap';\nimport promiseForObject from '../jsutils/promiseForObject';\nimport promiseReduce from '../jsutils/promiseReduce';\nimport type {\n  DocumentNode,\n  FieldNode,\n  FragmentDefinitionNode,\n  FragmentSpreadNode,\n  InlineFragmentNode,\n  OperationDefinitionNode,\n  SelectionSetNode,\n} from '../language/ast';\nimport { Kind } from '../language/kinds';\nimport type {\n  GraphQLAbstractType,\n  GraphQLField,\n  GraphQLFieldResolver,\n  GraphQLLeafType,\n  GraphQLList,\n  GraphQLObjectType,\n  GraphQLOutputType,\n  GraphQLResolveInfo,\n  ResponsePath,\n} from '../type/definition';\nimport {\n  isAbstractType,\n  isLeafType,\n  isListType,\n  isNonNullType,\n  isObjectType,\n} from '../type/definition';\nimport {\n  GraphQLIncludeDirective,\n  GraphQLSkipDirective,\n} from '../type/directives';\nimport {\n  SchemaMetaFieldDef,\n  TypeMetaFieldDef,\n  TypeNameMetaFieldDef,\n} from '../type/introspection';\nimport type { GraphQLSchema } from '../type/schema';\nimport { assertValidSchema } from '../type/validate';\nimport { getOperationRootType } from '../utilities/getOperationRootType';\nimport { typeFromAST } from '../utilities/typeFromAST';\nimport {\n  getArgumentValues,\n  getDirectiveValues,\n  getVariableValues,\n} from './values';",
      errors: 1,
    },
  ],
});

/*
 * ===========================================================================
 * KNOWN GAPS (11) — isolated, NOT run. Verbatim upstream below.
 * ===========================================================================
 *
 * rslint parses with the TypeScript-Go (ts-go) parser. The cases below use
 * Flow-only / deprecated syntax that ts-go rejects as a SYNTAX error (TS1003 /
 * TS1005 / TS2880). ts-go aborts the whole batch on any syntax error, so these
 * are isolated rather than run; they are NOT altered to pass. This is a genuine
 * parser-surface gap (Flow and the deprecated `assert` attribute keyword are
 * out of scope for rslint / ts-go).
 *
 * Upstream expectation for each (preserved verbatim):
 *
 *   typescript.valid[4]  [TS2880 deprecated `assert` import attribute keyword (ts-go requires `with`)]
 *     code:   "import json from \"./foo.json\" assert { type: \"json\" };"
 *
 *   flow.valid[6]  [TS1005 Flow `import typeof`]
 *     code:   "import typeof a from \"a\""
 *
 *   flow.valid[7]  [TS1005 Flow `import typeof {a}`]
 *     code:   "import typeof {a} from \"a\""
 *
 *   flow.valid[8]  [TS1005 Flow `import typeof a, {b}`]
 *     code:   "import typeof a, {b} from \"a\""
 *
 *   flow.valid[9]  [TS1005 Flow `import typeof {}`]
 *     code:   "import typeof {} from \"a\""
 *
 *   flow.valid[10]  [TS1005 Flow `import typeof {    }`]
 *     code:   "import typeof {    } from \"a\""
 *
 *   flow.valid[12]  [TS1003 Flow `import { ..., typeof b, ... }` specifier]
 *     code:   "import { a, typeof b, c, typeof d } from \"a\""
 *
 *   flow.valid[13]  [TS1003 Flow `import { type a, typeof b, c }` specifier]
 *     code:   "import { type a, typeof b, c } from \"a\""
 *
 *   flow.valid[14]  [TS1005 Flow `import typeof` block]
 *     code:   "import type x1 from \"a\";\nimport typeof x2 from \"b\"\nimport typeof x3 from \"c\";\nimport type x4 from \"d\""
 *
 *   flow.invalid[0]  [TS1003 Flow `import typeof` / `typeof T` specifier mix]
 *     code:   "import react from \"react\"\nimport type {Z} from \"Z\";\nimport './global.css';\nimport type {X} from \"X\";\nimport {truncate, typeof T, type Y, pluralize} from \"./utils\"\nimport type B from \"./B\";\nimport type C from \"/B\";\nimport type E from \"@/B\";\nimport typeof A from \"A\";\nimport typeof D from \"./D\";"
 *     output: "import './global.css';\n\nimport typeof A from \"A\";\nimport react from \"react\"\nimport type {X} from \"X\";\nimport type {Z} from \"Z\";\n\nimport type C from \"/B\";\nimport type E from \"@/B\";\n\nimport type B from \"./B\";\nimport typeof D from \"./D\";\nimport {pluralize,typeof T, truncate, type Y} from \"./utils\""
 *     errors: 1
 *
 *   flow.invalid[1]  [TS1005 Flow `import typeof`]
 *     code:   "import '@/';\nimport c from '@/';\nimport type C from '@/';\nimport b from './';\nimport type B from './';\nimport './';\nimport a from 'a';\nimport typeof A from 'a';\nimport 'a';\nimport {} from 'a';"
 *     output: "import a from 'a';\nimport {} from 'a';\n\nimport b from './';\n\nimport type B from './';\nimport type C from '@/';\nimport typeof A from 'a';\n\nimport '@/';\nimport './';\nimport 'a';\nimport c from '@/';"
 *     errors: 1
 */
