/**
 * @fileoverview Tests for eol-last rule.
 * @author Nodeca Team <https://github.com/nodeca>
 *
 * Ported verbatim from @stylistic/eslint-plugin v5.10.0:
 *   packages/eslint-plugin/rules/eol-last/eol-last.test.ts
 *
 * Transformations applied per the porting spec:
 *  - `run({ name, rule, ... })` -> `ruleTester.run('eol-last', null as never, { valid, invalid })`
 *  - All fixtures are plain single-quoted strings with explicit `\n` / `\r\n`
 *    escapes, kept byte-for-byte (no `$` unindent tag is used here).
 *  - `parserOptions` / `type` fields dropped. No `data` is attached to any error:
 *    both messageIds (`missing` / `unexpected`) render static text with no `{{}}`.
 *  - The `missing` cases pin only line/column (upstream gives no endLine/endColumn);
 *    the `unexpected` cases pin line/column/endLine/endColumn — preserved exactly.
 *
 * No Babel/Flow cases, no `suggestions`, and no external-fixture cases exist in the
 * upstream eol-last text tests. The `eol-last._css_.test.ts` (CSS source) and
 * `eol-last._unknown_.test.ts` (a synthetic *binary* `@eslint/plugin-kit` fake-
 * language plugin) files are excluded: neither is JS/TS, and the `_unknown_` one
 * depends on a fake plugin rslint cannot mount.
 *
 * No rslint<->upstream gap surfaced for this rule — every valid/invalid case
 * (including the bare-`Position`-loc end-of-file `unexpected` reports) runs green,
 * so the KNOWN GAPS block below is intentionally empty.
 */

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('eol-last', null as never, {
  valid: [
    '',
    '\n',
    'var a = 123;\n',
    'var a = 123;\n\n',
    'var a = 123;\n   \n',

    '\r\n',
    'var a = 123;\r\n',
    'var a = 123;\r\n\r\n',
    'var a = 123;\r\n   \r\n',

    { code: 'var a = 123;', options: ['never'] },
    { code: 'var a = 123;\nvar b = 456;', options: ['never'] },
    { code: 'var a = 123;\r\nvar b = 456;', options: ['never'] },

    // Deprecated: `"unix"` parameter
    { code: '', options: ['unix'] },
    { code: '\n', options: ['unix'] },
    { code: 'var a = 123;\n', options: ['unix'] },
    { code: 'var a = 123;\n\n', options: ['unix'] },
    { code: 'var a = 123;\n   \n', options: ['unix'] },

    // Deprecated: `"windows"` parameter
    { code: '', options: ['windows'] },
    { code: '\n', options: ['windows'] },
    { code: '\r\n', options: ['windows'] },
    { code: 'var a = 123;\r\n', options: ['windows'] },
    { code: 'var a = 123;\r\n\r\n', options: ['windows'] },
    { code: 'var a = 123;\r\n   \r\n', options: ['windows'] },
  ],

  invalid: [
    {
      code: 'var a = 123;',
      output: 'var a = 123;\n',
      errors: [{
        messageId: 'missing',
        line: 1,
        column: 13,
      }],
    },
    {
      code: 'var a = 123;\n   ',
      output: 'var a = 123;\n   \n',
      errors: [{
        messageId: 'missing',
        line: 2,
        column: 4,
      }],
    },
    {
      code: 'var a = 123;\n',
      output: 'var a = 123;',
      options: ['never'],
      errors: [{
        messageId: 'unexpected',
        line: 1,
        column: 13,
        endLine: 2,
        endColumn: 1,
      }],
    },
    {
      code: 'var a = 123;\r\n',
      output: 'var a = 123;',
      options: ['never'],
      errors: [{
        messageId: 'unexpected',
        line: 1,
        column: 13,
        endLine: 2,
        endColumn: 1,
      }],
    },
    {
      code: 'var a = 123;\r\n\r\n',
      output: 'var a = 123;',
      options: ['never'],
      errors: [{
        messageId: 'unexpected',
        line: 2,
        column: 1,
        endLine: 3,
        endColumn: 1,
      }],
    },
    {
      code: 'var a = 123;\nvar b = 456;\n',
      output: 'var a = 123;\nvar b = 456;',
      options: ['never'],
      errors: [{
        messageId: 'unexpected',
        line: 2,
        column: 13,
        endLine: 3,
        endColumn: 1,
      }],
    },
    {
      code: 'var a = 123;\r\nvar b = 456;\r\n',
      output: 'var a = 123;\r\nvar b = 456;',
      options: ['never'],
      errors: [{
        messageId: 'unexpected',
        line: 2,
        column: 13,
        endLine: 3,
        endColumn: 1,
      }],
    },
    {
      code: 'var a = 123;\n\n',
      output: 'var a = 123;',
      options: ['never'],
      errors: [{
        messageId: 'unexpected',
        line: 2,
        column: 1,
        endLine: 3,
        endColumn: 1,
      }],
    },

    // Deprecated: `"unix"` parameter
    {
      code: 'var a = 123;',
      output: 'var a = 123;\n',
      options: ['unix'],
      errors: [{
        messageId: 'missing',
        line: 1,
        column: 13,
      }],
    },
    {
      code: 'var a = 123;\n   ',
      output: 'var a = 123;\n   \n',
      options: ['unix'],
      errors: [{
        messageId: 'missing',
        line: 2,
        column: 4,
      }],
    },

    // Deprecated: `"windows"` parameter
    {
      code: 'var a = 123;',
      output: 'var a = 123;\r\n',
      options: ['windows'],
      errors: [{
        messageId: 'missing',
        line: 1,
        column: 13,
      }],
    },
    {
      code: 'var a = 123;\r\n   ',
      output: 'var a = 123;\r\n   \r\n',
      options: ['windows'],
      errors: [{
        messageId: 'missing',
        line: 2,
        column: 4,
      }],
    },
  ],
});
