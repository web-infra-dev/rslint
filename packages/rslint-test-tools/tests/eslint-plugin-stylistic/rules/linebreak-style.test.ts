/**
 * @fileoverview No mixed linebreaks
 * @author Erik Mueller
 *
 * Ported verbatim from @stylistic/eslint-plugin v5.10.0:
 *   packages/eslint-plugin/rules/linebreak-style/linebreak-style.test.ts
 *
 * Transformations applied per the porting spec:
 *  - `run({ name, rule, ... })` -> `ruleTester.run('linebreak-style', null as never, { valid, invalid })`
 *  - `parserOptions` / `type` fields: none present upstream, nothing dropped.
 *
 * The `\r` / `\n` / `\u2028` / `\u2029` byte sequences in every `code` and
 * `output` are preserved EXACTLY as upstream wrote them (single-quoted string
 * literals with escape sequences), since this rule is entirely about line-ending
 * bytes — a stray normalization would change the test's meaning.
 *
 * The upstream `linebreak-style._css_.test.ts` / `linebreak-style._unknown_.test.ts`
 * files are excluded per the porting spec (non-TS dialects).
 *
 * Cases that surface a real rslint<->upstream gap are NOT deleted or altered:
 * they are moved to the `linebreak-style — KNOWN GAPS` block comment at the
 * bottom, each annotated with what upstream expects vs. what rslint produces.
 */

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('linebreak-style', null as never, {
  valid: [
    'var a = \'a\',\n b = \'b\';\n\n function foo(params) {\n /* do stuff */ \n }\n',
    {
      code: 'var a = \'a\',\n b = \'b\';\n\n function foo(params) {\n /* do stuff */ \n }\n',
      options: ['unix'],
    },
    {
      code: 'var a = \'a\',\r\n b = \'b\';\r\n\r\n function foo(params) {\r\n /* do stuff */ \r\n }\r\n',
      options: ['windows'],
    },
    {
      code: 'var b = \'b\';',
      options: ['unix'],
    },
    {
      code: 'var b = \'b\';',
      options: ['windows'],
    },
  ],

  invalid: [
    {
      code: 'var a = \'a\';\r\n',
      output: 'var a = \'a\';\n',
      errors: [{
        line: 1,
        column: 13,
        endLine: 2,
        endColumn: 1,
        messageId: 'expectedLF',
      }],
    },
    {
      code: 'var a = \'a\';\r\n',
      output: 'var a = \'a\';\n',
      options: ['unix'],
      errors: [{
        line: 1,
        column: 13,
        endLine: 2,
        endColumn: 1,
        messageId: 'expectedLF',
      }],
    },
    {
      code: 'var a = \'a\';\n',
      output: 'var a = \'a\';\r\n',
      options: ['windows'],
      errors: [{
        line: 1,
        column: 13,
        endLine: 2,
        endColumn: 1,
        messageId: 'expectedCRLF',
      }],
    },
    {
      code: 'var a = \'a\';\u2028',
      output: 'var a = \'a\';\n',
      options: ['unix'],
      errors: [{
        line: 1,
        column: 13,
        endLine: 2,
        endColumn: 1,
        messageId: 'expectedLF',
      }],
    },
    {
      code: 'var a = \'a\';\u2029',
      output: 'var a = \'a\';\n',
      options: ['unix'],
      errors: [{
        line: 1,
        column: 13,
        endLine: 2,
        endColumn: 1,
        messageId: 'expectedLF',
      }],
    },
    {
      code: 'var a = \'a\',\n b = \'b\';\n\n function foo(params) {\r\n /* do stuff */ \n }\r\n',
      output: 'var a = \'a\',\n b = \'b\';\n\n function foo(params) {\n /* do stuff */ \n }\n',
      errors: [{
        line: 4,
        column: 24,
        endLine: 5,
        endColumn: 1,
        messageId: 'expectedLF',
      }, {
        line: 6,
        column: 3,
        endLine: 7,
        endColumn: 1,
        messageId: 'expectedLF',
      }],
    },
    {
      code: 'var a = \'a\',\r\n b = \'b\';\r\n\n function foo(params) {\r\n\n /* do stuff */ \n }\r\n',
      output: 'var a = \'a\',\r\n b = \'b\';\r\n\r\n function foo(params) {\r\n\r\n /* do stuff */ \r\n }\r\n',
      options: ['windows'],
      errors: [{
        line: 3,
        column: 1,
        endLine: 4,
        endColumn: 1,
        messageId: 'expectedCRLF',
      }, {
        line: 5,
        column: 1,
        endLine: 6,
        endColumn: 1,
        messageId: 'expectedCRLF',
      }, {
        line: 6,
        column: 17,
        endLine: 7,
        endColumn: 1,
        messageId: 'expectedCRLF',
      }],
    },
    {
      code: '\r\n',
      output: '\n',
      options: ['unix'],
      errors: [{
        line: 1,
        column: 1,
        endLine: 2,
        endColumn: 1,
        messageId: 'expectedLF',
      }],
    },
    {
      code: '\n',
      output: '\r\n',
      options: ['windows'],
      errors: [{
        line: 1,
        column: 1,
        endLine: 2,
        endColumn: 1,
        messageId: 'expectedCRLF',
      }],
    },
  ],
});
