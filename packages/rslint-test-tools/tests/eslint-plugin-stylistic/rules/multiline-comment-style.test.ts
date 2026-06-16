/**
 * @fileoverview enforce a particular style for multiline comments
 * @author Teddy Katz
 *
 * Ported verbatim from @stylistic/eslint-plugin v5.10.0:
 *   packages/eslint-plugin/rules/multiline-comment-style/multiline-comment-style.test.ts
 *
 * Transformations applied per the porting spec:
 *  - `run({ name, rule, ... })` -> `ruleTester.run('multiline-comment-style', null as never, { valid, invalid })`
 *  - The `$` unindent template tag (3 valid cases) is evaluated to its real
 *    multi-line string; plain-backtick templates keep their literal indentation
 *    verbatim, with `${' '}` / `${' '.repeat(3)}` / `${'<spaces>'}`
 *    interpolations evaluated to the exact whitespace they produce.
 *  - The messageIds (`expectedBlock` / `expectedBareBlock` / `startNewline` /
 *    `endNewline` / `missingStar` / `alignment` / `expectedLines`) take no
 *    `data`, so they map 1:1 to a fixed message rendered from the plugin's meta.
 *  - No `parserOptions`, no `type` fields, no Babel/Flow cases, no output-only
 *    invalid cases, and no suggestions exist for this rule.
 *
 * KNOWN GAPS: none. Every upstream case parses under rslint's ts-go parser and
 * produces byte-identical diagnostics (count + messageId + line) and autofix
 * output. The comment-only fixtures contain no TS-incompatible syntax. The
 * `._css_` / `._json_` / `._markdown_` test files don't exist for this rule.
 */

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('multiline-comment-style', null as never, {
  valid: [
    "\n            /*\n             * this is\n             * a comment\n             */\n        ",
    "\n            /**\n             * this is\n             * a JSDoc comment\n             */\n        ",
    "\n            /*!\n             * this is\n             * an exclamation comment\n             */\n        ",
    "\n            /*! this is a single line exclamation comment */\n        ",
    "\n            /* eslint semi: [\n              \"error\"\n            ] */\n        ",
    "\n            // this is a single-line comment\n        ",
    "\n            /* foo */\n        ",
    "\n            // this is a comment\n            foo();\n            // this is another comment\n        ",
    "\n            /*\n             * Function overview\n             * ...\n             */\n\n            // Step 1: Do the first thing\n            foo();\n        ",
    "\n            /*\n             * Function overview\n             * ...\n             */\n\n            /*\n             * Step 1: Do the first thing.\n             * The first thing is foo().\n             */\n            foo();\n        ",
    "\t\t/**\n\t\t * this comment\n\t\t * is tab-aligned\n\t\t */",
    "/**\r\n * this comment\r\n * uses windows linebreaks\r\n */",
    "/**  * this comment  * uses paragraph separators  */",
    "\n            foo(/* this is an\n                inline comment */);\n        ",
    "\n            // The following line comment\n            // contains '*/'.\n        ",
    {
      code: "\n                // The following line comment\n                // contains '*/'.\n            ",
      options: ["bare-block"],
    },
    {
      code: "\n                /*\n                 * this is\n                 * a comment\n                 */\n            ",
      options: ["starred-block"],
    },
    {
      code: "\n                /**\n                 * this is\n                 * a JSDoc comment\n                 */\n            ",
      options: ["starred-block"],
    },
    {
      code: "\n                /* eslint semi: [\n                  \"error\"\n                ] */\n            ",
      options: ["starred-block"],
    },
    {
      code: "\n                // this is a single-line comment\n            ",
      options: ["starred-block"],
    },
    {
      code: "\n                /* foo */\n            ",
      options: ["starred-block"],
    },
    {
      code: "\n                /*\n                 * foo\n                 */\n            ",
      options: ["starred-block"],
    },
    {
      code: "\n                /* foo */\n            ",
      options: ["bare-block"],
    },
    {
      code: "\n                /*\n                   foo */\n            ",
      options: ["bare-block"],
    },
    {
      code: "\n                /*\n                   foo\n                */\n            ",
      options: ["bare-block"],
    },
    {
      code: "\n                /*\n              foo */\n            ",
      options: ["bare-block"],
    },
    {
      code: "\n                /*\n            foo\n        */\n            ",
      options: ["bare-block"],
    },
    {
      code: "\n                // this is\n                // a comment\n            ",
      options: ["separate-lines"],
    },
    {
      code: "\n                /* this is\n                   a comment */ foo;\n            ",
      options: ["separate-lines"],
    },
    {
      code: "\n                // a comment\n\n                // another comment\n            ",
      options: ["separate-lines"],
    },
    {
      code: "\n                // a comment\n\n                // another comment\n            ",
      options: ["bare-block"],
    },
    {
      code: "\n                // a comment\n\n                // another comment\n            ",
      options: ["starred-block"],
    },
    {
      code: "\n                /* eslint semi: \"error\" */\n            ",
      options: ["separate-lines"],
    },
    {
      code: "\n                /**\n                 * This is\n                 * a JSDoc comment\n                 */\n            ",
      options: ["separate-lines"],
    },
    {
      code: "\n                /**\n                 * This is\n                 * a JSDoc comment\n                 */\n            ",
      options: ["starred-block"],
    },
    {
      code: "\n                /**\n                 * This is\n                 * a JSDoc comment\n                 */\n            ",
      options: ["bare-block"],
    },
    {
      code: "\n                /*!\n                 * This is\n                 * an exclamation comment\n                 */\n            ",
      options: ["separate-lines"],
    },
    {
      code: "\n                /*!\n                 * This is\n                 * an exclamation comment\n                 */\n            ",
      options: ["starred-block"],
    },
    {
      code: "\n                /*!\n                 * This is\n                 * an exclamation comment\n                 */\n            ",
      options: ["bare-block"],
    },
    {
      code: "\n                /* This is\n                   a comment */\n            ",
      options: ["bare-block"],
    },
    {
      code: "\n                /* This is\n                         a comment */\n            ",
      options: ["bare-block"],
    },
    {
      code: "\n                /* eslint semi: [\n                    \"error\"\n                ] */\n            ",
      options: ["separate-lines"],
    },
    {
      code: "\n                /* The value of 5\n                 + 4 is 9, and the value of 5\n                 * 4 is 20. */\n            ",
      options: ["bare-block"],
    },
    {
      code: "\n                /*\n                 *    foo\n                 *  bar\n                 *   baz\n                 * qux\n                 */\n            ",
      options: ["starred-block"],
    },
    {
      code: "\n                /*    foo\n                 *  bar\n                 *   baz\n                 * qux\n                 */\n            ",
      options: ["bare-block"],
    },
    {
      code: "\n                /**\n                 *    JSDoc blocks\n                 *  are\n                 *   ignored\n                 * !\n                 */\n            ",
      options: ["bare-block"],
    },
    {
      code: "\n                /**\n                 *    JSDoc blocks\n                 *  are\n                 *   ignored\n                 * !\n                 */\n            ",
      options: ["starred-block"],
    },
    {
      code: "\n                /**\n                 *    JSDoc blocks\n                 *  are\n                 *   ignored\n                 * !\n                 */\n            ",
      options: ["separate-lines"],
    },
    {
      code: "\n                /*!\n                 *    Exclamation blocks\n                 *  are\n                 *   ignored\n                 * !\n                 */\n            ",
      options: ["bare-block"],
    },
    {
      code: "\n                /*!\n                 *    Exclamation blocks\n                 *  are\n                 *   ignored\n                 * !\n                 */\n            ",
      options: ["starred-block"],
    },
    {
      code: "\n                /*!\n                 *    Exclamation blocks\n                 *  are\n                 *   ignored\n                 * !\n                 */\n            ",
      options: ["separate-lines"],
    },
    {
      code: "\n                /*\n                 * // a line comment\n                 *some.code();\n                 */\n            ",
      options: ["starred-block"],
    },
    {
      code: "// djb2 algorithm\nlet hash = 5381;\nfor (let i = 0; i < view.length; i++) {\n\n  // eslint-disable\n  // hash * 33 + current byte -> truncate\n  hash = (((hash << 5) + hash) + view[i]) | 0;\n}",
      options: ["starred-block"],
    },
    {
      code: "// eslint-disable\n// @ts-nocheck\n\nimport type { ESLint } from 'eslint';",
      options: ["starred-block"],
    },
    {
      code: "let x = 5; // first number\n// second number\nlet y = 10;\n\nconsole.log(x + y);",
      options: ["starred-block"],
    },
  ],

  invalid: [
    {
      code: "\n                // these are\n                // line comments\n            ",
      output: "\n                /*\n                 * these are\n                 * line comments\n                 */\n            ",
      errors: [
        { messageId: "expectedBlock", line: 2 }
      ],
    },
    {
      code: "\n                //foo\n                ///bar\n            ",
      output: null,
      errors: [
        { messageId: "expectedBlock", line: 2 }
      ],
    },
    {
      code: "\n                // foo\n                // bar\n\n                // baz\n                // qux\n            ",
      output: "\n                /*\n                 * foo\n                 * bar\n                 */\n\n                /*\n                 * baz\n                 * qux\n                 */\n            ",
      errors: [
        { messageId: "expectedBlock", line: 2 },
        { messageId: "expectedBlock", line: 5 }
      ],
    },
    {
      code: "\n                //  foo\n                // bar\n                //    baz\n                // qux\n            ",
      output: "\n                /*\n                 *  foo\n                 * bar\n                 *    baz\n                 * qux\n                 */\n            ",
      errors: [
        { messageId: "expectedBlock", line: 2 }
      ],
    },
    {
      code: "\n                //  foo\n                //\n                //    baz\n                // qux\n            ",
      output: "\n                /*\n                 *  foo\n                 * \n                 *    baz\n                 * qux\n                 */\n            ",
      errors: [
        { messageId: "expectedBlock", line: 2 }
      ],
    },
    {
      code: "\n                //    foo\n                     // bar\n           //  baz\n                // qux\n            ",
      output: "\n                /*\n                 *    foo\n                 * bar\n                 *  baz\n                 * qux\n                 */\n            ",
      errors: [
        { messageId: "expectedBlock", line: 2 }
      ],
    },
    {
      code: "\n                /* this block\n                 * is missing a newline at the start\n                 */\n            ",
      output: "\n                /*\n                 * this block\n                 * is missing a newline at the start\n                 */\n            ",
      errors: [
        { messageId: "startNewline", line: 2 }
      ],
    },
    {
      code: "\n                /** this JSDoc comment\n                 * is missing a newline at the start\n                 */\n            ",
      output: "\n                /**\n                 * this JSDoc comment\n                 * is missing a newline at the start\n                 */\n            ",
      errors: [
        { messageId: "startNewline", line: 2 }
      ],
    },
    {
      code: "\n                /*! this Exclamation comment\n                 * is missing a newline\n                 * at the start\n                 */\n            ",
      output: "\n                /*!\n                 * this Exclamation comment\n                 * is missing a newline\n                 * at the start\n                 */\n            ",
      errors: [
        { messageId: "startNewline", line: 2 }
      ],
    },
    {
      code: "\n                /*! this Exclamation comment is missing a newline at the start\n                 */\n            ",
      output: "\n                /*!\n                 * this Exclamation comment is missing a newline at the start\n                 */\n            ",
      errors: [
        { messageId: "startNewline", line: 2 }
      ],
    },
    {
      code: "\n                /*\n                 * this block\n                 * is missing a newline at the end*/\n            ",
      output: "\n                /*\n                 * this block\n                 * is missing a newline at the end\n                 */\n            ",
      errors: [
        { messageId: "endNewline", line: 4 }
      ],
    },
    {
      code: "\n                /*\n                 * the following line\n                   is missing a '*' at the start\n                 */\n            ",
      output: "\n                /*\n                 * the following line\n                 * is missing a '*' at the start\n                 */\n            ",
      errors: [
        { messageId: "missingStar", line: 4 }
      ],
    },
    {
      code: "\n                /*\n                 * the following line\n                 is missing a '*' at the start\n                 */\n            ",
      output: "\n                /*\n                 * the following line\n                 * is missing a '*' at the start\n                 */\n            ",
      errors: [
        { messageId: "missingStar", line: 4 }
      ],
    },
    {
      code: "\n                /*\n                 * the following line\n                is missing a '*' at the start\n                 */\n            ",
      output: "\n                /*\n                 * the following line\n                 * is missing a '*' at the start\n                 */\n            ",
      errors: [
        { messageId: "missingStar", line: 4 }
      ],
    },
    {
      code: "\n                /*\n                 * the following line\n                      * has a '*' with the wrong offset at the start\n                 */\n            ",
      output: "\n                /*\n                 * the following line\n                 * has a '*' with the wrong offset at the start\n                 */\n            ",
      errors: [
        { messageId: "alignment", line: 4 }
      ],
    },
    {
      code: "\n                  /*\n                   * the following line\n                 * has a '*' with the wrong offset at the start\n                   */\n            ",
      output: "\n                  /*\n                   * the following line\n                   * has a '*' with the wrong offset at the start\n                   */\n            ",
      errors: [
        { messageId: "alignment", line: 4 }
      ],
    },
    {
      code: "\n                /*\n                 * the last line of this comment\n                 * is misaligned\n                   */\n            ",
      output: "\n                /*\n                 * the last line of this comment\n                 * is misaligned\n                 */\n            ",
      errors: [
        { messageId: "alignment", line: 5 }
      ],
    },
    {
      code: "\n                /*\n                 * the following line\n                *\n                 * is blank\n                 */\n            ",
      output: "\n                /*\n                 * the following line\n                 *\n                 * is blank\n                 */\n            ",
      errors: [
        { messageId: "alignment", line: 4 }
      ],
    },
    {
      code: "\n                /*\n                 * the following line\n                  *\n                 * is blank\n                 */\n            ",
      output: "\n                /*\n                 * the following line\n                 *\n                 * is blank\n                 */\n            ",
      errors: [
        { messageId: "alignment", line: 4 }
      ],
    },
    {
      code: "\n                /*\n                 * the last line of this comment\n                 * is misaligned\n                   */ foo\n            ",
      output: "\n                /*\n                 * the last line of this comment\n                 * is misaligned\n                 */ foo\n            ",
      errors: [
        { messageId: "alignment", line: 5 }
      ],
    },
    {
      code: "\n                /*\n                 * foo\n                 * bar\n                 */\n            ",
      options: ["separate-lines"],
      output: "\n                // foo\n                // bar\n            ",
      errors: [
        { messageId: "expectedLines", line: 2 }
      ],
    },
    {
      code: "\n                /**\n                 * JSDoc\n                 * Comment\n                 */\n            ",
      options: ["separate-lines",{"checkJSDoc":true}],
      output: "\n                // JSDoc\n                // Comment\n            ",
      errors: [
        { messageId: "expectedLines", line: 2 }
      ],
    },
    {
      code: "\n                /*!\n                 * Exclamation\n                 * Comment\n                 */\n            ",
      options: ["separate-lines",{"checkExclamation":true}],
      output: "\n                // Exclamation\n                // Comment\n            ",
      errors: [
        { messageId: "expectedLines", line: 2 }
      ],
    },
    {
      code: "\n                /* foo\n                 *bar\n                 baz\n                 qux*/\n            ",
      options: ["separate-lines"],
      output: "\n                // foo\n                // bar\n                // baz\n                // qux\n            ",
      errors: [
        { messageId: "expectedLines", line: 2 }
      ],
    },
    {
      code: "\n                // foo\n                // bar\n            ",
      options: ["bare-block"],
      output: "\n                /* foo\n                   bar */\n            ",
      errors: [
        { messageId: "expectedBlock", line: 2 }
      ],
    },
    {
      code: "\n                // foo\n                //\n                // bar\n            ",
      options: ["bare-block"],
      output: "\n                /* foo\n                   \n                   bar */\n            ",
      errors: [
        { messageId: "expectedBlock", line: 2 }
      ],
    },
    {
      code: "\n                //foo\n                //bar\n            ",
      options: ["bare-block"],
      output: "\n                /* foo\n                   bar */\n            ",
      errors: [
        { messageId: "expectedBlock", line: 2 }
      ],
    },
    {
      code: "\n                //   foo\n                //   bar\n            ",
      options: ["bare-block"],
      output: "\n                /*   foo\n                     bar */\n            ",
      errors: [
        { messageId: "expectedBlock", line: 2 }
      ],
    },
    {
      code: "\n                // foo\n              // bar\n            ",
      options: ["bare-block"],
      output: "\n                /* foo\n                   bar */\n            ",
      errors: [
        { messageId: "expectedBlock", line: 2 }
      ],
    },
    {
      code: "\n                //    foo\n                     // bar\n           //  baz\n                // qux\n            ",
      options: ["bare-block"],
      output: "\n                /*    foo\n                   bar\n                    baz\n                   qux */\n            ",
      errors: [
        { messageId: "expectedBlock", line: 2 }
      ],
    },
    {
      code: "\n                /*\n                * foo\n                * bar\n                */\n            ",
      options: ["bare-block"],
      output: "\n                /* foo\n                   bar */\n            ",
      errors: [
        { messageId: "expectedBareBlock", line: 2 }
      ],
    },
    {
      code: "\n                /*\n                *foo\n                *bar\n                */\n            ",
      options: ["bare-block"],
      output: "\n                /* foo\n                   bar */\n            ",
      errors: [
        { messageId: "expectedBareBlock", line: 2 }
      ],
    },
    {
      code: "\n                /*\n                *   foo\n                *   bar\n                */\n            ",
      options: ["bare-block"],
      output: "\n                /*   foo\n                     bar */\n            ",
      errors: [
        { messageId: "expectedBareBlock", line: 2 }
      ],
    },
    {
      code: "\n                /*\n                * foo\n             * bar\n                */\n            ",
      options: ["bare-block"],
      output: "\n                /* foo\n                   bar */\n            ",
      errors: [
        { messageId: "expectedBareBlock", line: 2 }
      ],
    },
    {
      code: "\n                /*\n                 *    foo\n                 *  bar\n                 *   baz\n                 * qux\n                 */\n            ",
      options: ["bare-block"],
      output: "\n                /*    foo\n                    bar\n                     baz\n                   qux */\n            ",
      errors: [
        { messageId: "expectedBareBlock", line: 2 }
      ],
    },
    {
      code: "\n                /*\n                {\n                    \"foo\": 1,\n                    \"bar\": 2\n                }\n                */\n            ",
      output: "\n                /*\n                 *{\n                 *    \"foo\": 1,\n                 *    \"bar\": 2\n                 *}\n                 */\n            ",
      errors: [
        { messageId: "missingStar", line: 3 },
        { messageId: "missingStar", line: 4 },
        { messageId: "missingStar", line: 5 },
        { messageId: "missingStar", line: 6 },
        { messageId: "alignment", line: 7 }
      ],
    },
    {
      code: "\n                /*\n                {\n                \t\"foo\": 1,\n                \t\"bar\": 2\n                }\n                */\n            ",
      output: "\n                /*\n                 *{\n                 *\t\"foo\": 1,\n                 *\t\"bar\": 2\n                 *}\n                 */\n            ",
      errors: [
        { messageId: "missingStar", line: 3 },
        { messageId: "missingStar", line: 4 },
        { messageId: "missingStar", line: 5 },
        { messageId: "missingStar", line: 6 },
        { messageId: "alignment", line: 7 }
      ],
    },
    {
      code: "\n                /*\n                {\n                \t  \"foo\": 1,\n                \t  \"bar\": 2\n                }\n                */\n            ",
      output: "\n                /*\n                 *{\n                 *\t  \"foo\": 1,\n                 *\t  \"bar\": 2\n                 *}\n                 */\n            ",
      errors: [
        { messageId: "missingStar", line: 3 },
        { messageId: "missingStar", line: 4 },
        { messageId: "missingStar", line: 5 },
        { messageId: "missingStar", line: 6 },
        { messageId: "alignment", line: 7 }
      ],
    },
    {
      code: "\n                /*\n                {\n               \t\"foo\": 1,\n               \t\"bar\": 2\n                }\n                */\n            ",
      output: "\n                /*\n                 *{\n                 *\"foo\": 1,\n                 *\"bar\": 2\n                 *}\n                 */\n            ",
      errors: [
        { messageId: "missingStar", line: 3 },
        { messageId: "missingStar", line: 4 },
        { messageId: "missingStar", line: 5 },
        { messageId: "missingStar", line: 6 },
        { messageId: "alignment", line: 7 }
      ],
    },
    {
      code: "\n                \t /*\n                      \t    {\n                  \t    \"foo\": 1,\n                \t   \"bar\": 2\n                }\n                */\n            ",
      output: "\n                \t /*\n                \t  *{\n                \t  *\"foo\": 1,\n                \t  *\"bar\": 2\n                \t  *}\n                \t  */\n            ",
      errors: [
        { messageId: "missingStar", line: 3 },
        { messageId: "missingStar", line: 4 },
        { messageId: "missingStar", line: 5 },
        { messageId: "missingStar", line: 6 },
        { messageId: "alignment", line: 7 }
      ],
    },
    {
      code: "\n                //{\n                //    \"foo\": 1,\n                //    \"bar\": 2\n                //}\n            ",
      output: "\n                /*\n                 * {\n                 *     \"foo\": 1,\n                 *     \"bar\": 2\n                 * }\n                 */\n            ",
      errors: [
        { messageId: "expectedBlock", line: 2 }
      ],
    },
    {
      code: "\n                /*\n                 * {\n                 *     \"foo\": 1,\n                 *     \"bar\": 2\n                 * }\n                 */\n            ",
      options: ["separate-lines"],
      output: "\n                // {\n                //     \"foo\": 1,\n                //     \"bar\": 2\n                // }\n            ",
      errors: [
        { messageId: "expectedLines", line: 2 }
      ],
    },
    {
      code: "\n                /*\n                 *\n                 * {\n                 *     \"foo\": 1,\n                 *     \"bar\": 2\n                 * }\n                 *\n                 */\n            ",
      options: ["separate-lines"],
      output: "\n                // \n                // {\n                //     \"foo\": 1,\n                //     \"bar\": 2\n                // }\n                // \n            ",
      errors: [
        { messageId: "expectedLines", line: 2 }
      ],
    },
    {
      code: "\n                /*\n                 *\n                 * {\n                 *     \"foo\": 1,\n                 *     \"bar\": 2\n                 * }\n                 *\n                 */\n            ",
      options: ["bare-block"],
      output: "\n                /* \n                   {\n                       \"foo\": 1,\n                       \"bar\": 2\n                   }\n                    */\n            ",
      errors: [
        { messageId: "expectedBareBlock", line: 2 }
      ],
    },
    {
      code: "\n                /*\n                 *\n                 *{\n                 *    \"foo\": 1,\n                 *    \"bar\": 2\n                 *}\n                 *\n                 */\n            ",
      options: ["bare-block"],
      output: "\n                /* \n                   {\n                       \"foo\": 1,\n                       \"bar\": 2\n                   }\n                    */\n            ",
      errors: [
        { messageId: "expectedBareBlock", line: 2 }
      ],
    },
    {
      code: "\n                /*\n                 *{\n                 *    \"foo\": 1,\n                 *    \"bar\": 2\n                 *}\n                 */\n            ",
      options: ["separate-lines"],
      output: "\n                // {\n                //     \"foo\": 1,\n                //     \"bar\": 2\n                // }\n            ",
      errors: [
        { messageId: "expectedLines", line: 2 }
      ],
    },
    {
      code: "\n                /*\n                 *   {\n                 *       \"foo\": 1,\n                 *       \"bar\": 2\n                 *   }\n                 */\n            ",
      options: ["separate-lines"],
      output: "\n                //   {\n                //       \"foo\": 1,\n                //       \"bar\": 2\n                //   }\n            ",
      errors: [
        { messageId: "expectedLines", line: 2 }
      ],
    },
    {
      code: "\n                /*\n            *{\n                 *    \"foo\": 1,\n                    *    \"bar\": 2\n                 *}\n                  */\n            ",
      options: ["separate-lines"],
      output: "\n                // {\n                //     \"foo\": 1,\n                //     \"bar\": 2\n                // }\n            ",
      errors: [
        { messageId: "expectedLines", line: 2 }
      ],
    },
    {
      code: "\n                /*\n                 *   {\n                 *       \"foo\": 1,\n                 *       \"bar\": 2\n           *}\n                 */\n            ",
      options: ["separate-lines"],
      output: "\n                //    {\n                //        \"foo\": 1,\n                //        \"bar\": 2\n                // }\n            ",
      errors: [
        { messageId: "expectedLines", line: 2 }
      ],
    },
    {
      code: "\n                /*\n                {\n                    \"foo\": 1,\n                    \"bar\": 2\n                }\n                */\n            ",
      options: ["separate-lines"],
      output: "\n                // \n                // {\n                //     \"foo\": 1,\n                //     \"bar\": 2\n                // }\n                // \n            ",
      errors: [
        { messageId: "expectedLines", line: 2 }
      ],
    },
    {
      code: "\n                /* {\n                       \"foo\": 1,\n                       \"bar\": 2\n                   } */\n            ",
      options: ["separate-lines"],
      output: "\n                // {\n                //     \"foo\": 1,\n                //     \"bar\": 2\n                // } \n            ",
      errors: [
        { messageId: "expectedLines", line: 2 }
      ],
    },
    {
      code: "\n                /*\n                 * foo\n                 *\n                 * bar\n                 */\n            ",
      options: ["separate-lines"],
      output: "\n                // foo\n                // \n                // bar\n            ",
      errors: [
        { messageId: "expectedLines", line: 2 }
      ],
    },
    {
      code: "\n                /*\n                 * foo\n                 * \n                 * bar\n                 */\n            ",
      options: ["separate-lines"],
      output: "\n                // foo\n                // \n                // bar\n            ",
      errors: [
        { messageId: "expectedLines", line: 2 }
      ],
    },
    {
      code: "\n                /*\n                 * foo\n                 *\n                 * bar\n                 */\n            ",
      options: ["bare-block"],
      output: "\n                /* foo\n                   \n                   bar */\n            ",
      errors: [
        { messageId: "expectedBareBlock", line: 2 }
      ],
    },
    {
      code: "\n                /*\n                 * foo\n                 * \n                 * bar\n                 */\n            ",
      options: ["bare-block"],
      output: "\n                /* foo\n                   \n                   bar */\n            ",
      errors: [
        { messageId: "expectedBareBlock", line: 2 }
      ],
    },
    {
      code: "\n                // foo\n                //\n                // bar\n            ",
      options: ["starred-block"],
      output: "\n                /*\n                 * foo\n                 * \n                 * bar\n                 */\n            ",
      errors: [
        { messageId: "expectedBlock", line: 2 }
      ],
    },
    {
      code: "\n                // foo\n                // \n                // bar\n            ",
      options: ["starred-block"],
      output: "\n                /*\n                 * foo\n                 * \n                 * bar\n                 */\n            ",
      errors: [
        { messageId: "expectedBlock", line: 2 }
      ],
    },
    {
      code: "\n                // foo\n                // \n                // bar\n            ",
      options: ["bare-block"],
      output: "\n                /* foo\n                   \n                   bar */\n            ",
      errors: [
        { messageId: "expectedBlock", line: 2 }
      ],
    },
    {
      code: "\n                /* foo\n\n                   bar */\n            ",
      options: ["separate-lines"],
      output: "\n                // foo\n                // \n                // bar \n            ",
      errors: [
        { messageId: "expectedLines", line: 2 }
      ],
    },
    {
      code: "\n                /* foo\n                   \n                   bar */\n            ",
      options: ["separate-lines"],
      output: "\n                // foo\n                // \n                // bar \n            ",
      errors: [
        { messageId: "expectedLines", line: 2 }
      ],
    },
    {
      code: "\n                /* foo\n\n                   bar */\n            ",
      options: ["starred-block"],
      output: "\n                /*\n                 * foo\n                 * \n                 * bar \n                 */\n            ",
      errors: [
        { messageId: "startNewline", line: 2 },
        { messageId: "missingStar", line: 3 },
        { messageId: "missingStar", line: 4 },
        { messageId: "endNewline", line: 4 }
      ],
    },
    {
      code: "\n                /* foo\n                   \n                   bar */\n            ",
      options: ["starred-block"],
      output: "\n                /*\n                 * foo\n                 * \n                 * bar \n                 */\n            ",
      errors: [
        { messageId: "startNewline", line: 2 },
        { messageId: "missingStar", line: 3 },
        { messageId: "missingStar", line: 4 },
        { messageId: "endNewline", line: 4 }
      ],
    },
    {
      code: "\n                /*foo\n\n                  bar */\n            ",
      options: ["starred-block"],
      output: "\n                /*\n                 *foo\n                 *\n                 *bar \n                 */\n            ",
      errors: [
        { messageId: "startNewline", line: 2 },
        { messageId: "missingStar", line: 3 },
        { messageId: "missingStar", line: 4 },
        { messageId: "endNewline", line: 4 }
      ],
    },
    {
      code: "\n                /*foo\n                   \n                  bar */\n            ",
      options: ["starred-block"],
      output: "\n                /*\n                 *foo\n                 * \n                 *bar \n                 */\n            ",
      errors: [
        { messageId: "startNewline", line: 2 },
        { messageId: "missingStar", line: 3 },
        { messageId: "missingStar", line: 4 },
        { messageId: "endNewline", line: 4 }
      ],
    },
    {
      code: "\n                /*\n                 // a line comment\n                 some.code();\n                 */\n            ",
      options: ["starred-block"],
      output: "\n                /*\n                 * // a line comment\n                 *some.code();\n                 */\n            ",
      errors: [
        { messageId: "missingStar", line: 3 },
        { messageId: "missingStar", line: 4 }
      ],
    },
    {
      code: "\n                /*\n                 // a line comment\n                 * some.code();\n                 */\n            ",
      options: ["starred-block"],
      output: "\n                /*\n                 * // a line comment\n                 * some.code();\n                 */\n            ",
      errors: [
        { messageId: "missingStar", line: 3 }
      ],
    },
    {
      code: "\n                ////This comment is in\n                //`separate-lines` format.\n            ",
      options: ["starred-block"],
      output: null,
      errors: [
        { messageId: "expectedBlock", line: 2 }
      ],
    },
    {
      code: "\n                // // This comment is in\n                // `separate-lines` format.\n            ",
      options: ["starred-block"],
      output: null,
      errors: [
        { messageId: "expectedBlock", line: 2 }
      ],
    },
    {
      code: "\n                /*\n                {\n                \t\"foo\": 1,\n                \t//\"bar\": 2\n                }\n                */\n            ",
      options: ["starred-block"],
      output: "\n                /*\n                 *{\n                 *\t\"foo\": 1,\n                 *\t//\"bar\": 2\n                 *}\n                 */\n            ",
      errors: [
        { messageId: "missingStar", line: 3 },
        { messageId: "missingStar", line: 4 },
        { messageId: "missingStar", line: 5 },
        { messageId: "missingStar", line: 6 },
        { messageId: "alignment", line: 7 }
      ],
    },
  ],
});
