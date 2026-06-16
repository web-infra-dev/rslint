/**
 * @fileoverview Disallows multiple blank lines.
 * @author Greg Cochard
 *
 * Ported verbatim from @stylistic/eslint-plugin v5.10.0:
 *   packages/eslint-plugin/rules/no-multiple-empty-lines/no-multiple-empty-lines.test.ts
 *
 * Transformations applied per the porting spec:
 *  - `run({ name, rule, ... })` -> `ruleTester.run('no-multiple-empty-lines', null as never, { valid, invalid })`
 *  - `parserOptions` (`ecmaVersion: 6`) dropped — rslint resolves via tsconfig;
 *    template literals already parse under the default ESNext target.
 *  - The upstream `getExpectedError` / `getExpectedErrorEOF` / `getExpectedErrorBOF`
 *    helpers are evaluated inline to their final error objects:
 *      getExpectedError(n)    -> { messageId: 'consecutiveBlank',
 *                                  data: { max: n, pluralizedLines: n === 1 ? 'line' : 'lines' },
 *                                  column: 1 }
 *      getExpectedErrorEOF(n) -> { messageId: 'blankEndOfFile', data: { max: n }, column: 1 }
 *      getExpectedErrorBOF(n) -> { messageId: 'blankBeginningOfFile', data: { max: n }, column: 1 }
 *    The two cases that already used a literal error object (with `line` + `column`)
 *    are ported verbatim.
 *  - The one `errors: 2` bare-count case is preserved as a bare count.
 *  - The `${'a'.repeat(1e5)}` template (eslint/eslint#7893) is evaluated to its
 *    real string via concatenation: `'a\n\n\n\n' + 'a'.repeat(1e5)` (code) and
 *    `'a\n\n\n' + 'a'.repeat(1e5)` (output).
 *
 * Every other `code`/`output` string is a JS string literal whose `\n` / `\r\n` /
 * space escapes are preserved byte-for-byte from upstream — there is no `$`
 * unindent tag and no plain-backtick multi-line template in this file.
 *
 * This rule IS fixable (`meta.fixable: 'whitespace'`); every invalid case pins
 * `output`, so the autofix pass runs and the fixed source is asserted.
 *
 * The upstream file contains NO `readFileSync` external-fixture cases, NO
 * `suggestions`, NO Babel/Flow cases, and NO second (skipBabel) `run()` block.
 * The `._css_` / `._json_` / `._markdown_` test files don't exist for this rule.
 *
 * No case surfaces a rslint<->upstream gap, so the KNOWN GAPS block below is
 * intentionally empty.
 */

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-multiple-empty-lines', null as never, {
  valid: [
    {
      code: '// valid 1\nvar a = 5;\nvar b = 3;\n\n',
      options: [{ max: 1 }],
    },
    {
      code: '// valid 2\n\nvar a = 5;\n\nvar b = 3;',
      options: [{ max: 1 }],
    },
    {
      code: '// valid 3\nvar a = 5;\n\nvar b = 3;\n\n\n',
      options: [{ max: 2 }],
    },
    {
      code: '// valid 4\nvar a = 5,\n    b = 3;',
      options: [{ max: 2 }],
    },
    {
      code: '// valid 5\nvar a = 5;\n\n\n\n\nvar b = 3;\n\n\n\n\n',
      options: [{ max: 4 }],
    },
    {
      code: '// valid 6\nvar a = 5;\n/* comment */\nvar b = 5;',
      options: [{ max: 0 }],
    },
    {
      code: '// valid 7\nvar a = 5;\n',
      options: [{ max: 0 }],
    },
    {
      code: '// valid 8\nvar a = 5;\n',
      options: [{ max: 0, maxEOF: 0 }],
    },
    {
      code: '// valid 9\nvar a = 1;\n\n',
      options: [{ max: 2, maxEOF: 1 }],
    },
    {
      code: '// valid 10\nvar a = 5;\n',
      options: [{ max: 0, maxBOF: 0 }],
    },
    {
      code: '\n// valid 11\nvar a = 1;\n',
      options: [{ max: 2, maxBOF: 1 }],
    },
    {
      code: '// valid 12\r\n// windows line endings\r\nvar a = 5;\r\nvar b = 3;\r\n\r\n',
      options: [{ max: 1 }],
    },

    // template strings
    {
      code: '// valid 12\nx = `\n\n\n\nhi\n\n\n\n`',
      options: [{ max: 2 }],
    },
    {
      code: '// valid 13\n`\n\n`',
      options: [{ max: 0 }],
    },
    {
      code: '// valid 14\nvar a = 5;`\n\n\n\n\n`',
      options: [{ max: 0, maxEOF: 0 }],
    },
    {
      code: '`\n\n\n\n\n`\n// valid 15\nvar a = 5;',
      options: [{ max: 0, maxBOF: 0 }],
    },
    {
      code: '\n\n\n\n// valid 16\nvar a = 5;\n',
      options: [{ max: 0, maxBOF: 4 }],
    },
    {
      code: '// valid 17\nvar a = 5;\n\n',
      options: [{ max: 0, maxEOF: 1 }],
    },
    {
      code: 'var a = 5;',
      options: [{ max: 1 }],
    },
  ],

  invalid: [
    {
      code: '// invalid 1\nvar a = 5;\n\n\nvar b = 3;',
      output: '// invalid 1\nvar a = 5;\n\nvar b = 3;',
      options: [{ max: 1 }],
      errors: [{ messageId: 'consecutiveBlank', data: { max: 1, pluralizedLines: 'line' }, column: 1 }],
    },
    {
      code: '// invalid 2\n\n\n\n\nvar a = 5;',
      output: '// invalid 2\n\n\nvar a = 5;',
      options: [{ max: 2 }],
      errors: [{ messageId: 'consecutiveBlank', data: { max: 2, pluralizedLines: 'lines' }, column: 1 }],
    },
    {
      code: '// invalid 3\nvar a = 5;\n\n\n\n',
      output: '// invalid 3\nvar a = 5;\n\n\n',
      options: [{ max: 2 }],
      errors: [{ messageId: 'blankEndOfFile', data: { max: 2 }, column: 1 }],
    },
    {
      code: '// invalid 4\nvar a = 5;\n \n \n \n',
      output: '// invalid 4\nvar a = 5;\n \n \n',
      options: [{ max: 2 }],
      errors: [{ messageId: 'blankEndOfFile', data: { max: 2 }, column: 1 }],
    },
    {
      code: '// invalid 5\nvar a=5;\n\n\n\nvar b = 3;',
      output: '// invalid 5\nvar a=5;\n\n\nvar b = 3;',
      options: [{ max: 2 }],
      errors: [{ messageId: 'consecutiveBlank', data: { max: 2, pluralizedLines: 'lines' }, column: 1 }],
    },
    {
      code: '// invalid 6\nvar a=5;\n\n\n\nvar b = 3;\n',
      output: '// invalid 6\nvar a=5;\n\n\nvar b = 3;\n',
      options: [{ max: 2 }],
      errors: [{ messageId: 'consecutiveBlank', data: { max: 2, pluralizedLines: 'lines' }, column: 1 }],
    },
    {
      code: '// invalid 7\nvar a = 5;\n\n\n\nb = 3;\nvar c = 5;\n\n\n\nvar d = 3;',
      output: '// invalid 7\nvar a = 5;\n\n\nb = 3;\nvar c = 5;\n\n\nvar d = 3;',
      options: [{ max: 2 }],
      errors: 2,
    },
    {
      code: '// invalid 8\nvar a = 5;\n\n\n\n\n\n\n\n\n\n\n\n\n\nb = 3;',
      output: '// invalid 8\nvar a = 5;\n\n\nb = 3;',
      options: [{ max: 2 }],
      errors: [{ messageId: 'consecutiveBlank', data: { max: 2, pluralizedLines: 'lines' }, column: 1 }],
    },
    {
      code: '// invalid 9\nvar a=5;\n\n\n\n\n',
      output: '// invalid 9\nvar a=5;\n\n\n',
      options: [{ max: 2 }],
      errors: [{ messageId: 'blankEndOfFile', data: { max: 2 }, column: 1 }],
    },
    {
      code: '// invalid 10\nvar a = 5;\n\nvar b = 3;',
      output: '// invalid 10\nvar a = 5;\nvar b = 3;',
      options: [{ max: 0 }],
      errors: [{ messageId: 'consecutiveBlank', data: { max: 0, pluralizedLines: 'lines' }, column: 1 }],
    },
    {
      code: '// invalid 11\nvar a = 5;\n\n\n',
      output: '// invalid 11\nvar a = 5;\n\n',
      options: [{ max: 5, maxEOF: 1 }],
      errors: [{ messageId: 'blankEndOfFile', data: { max: 1 }, column: 1 }],
    },
    {
      code: '// invalid 12\nvar a = 5;\n\n\n\n\n\n',
      output: '// invalid 12\nvar a = 5;\n\n\n\n\n',
      options: [{ max: 0, maxEOF: 4 }],
      errors: [{ messageId: 'blankEndOfFile', data: { max: 4 }, column: 1 }],
    },
    {
      code: '// invalid 13\n\n\n\n\n\n\n\n\nvar a = 5;\n\n\n',
      output: '// invalid 13\n\n\n\n\n\n\n\n\nvar a = 5;\n\n',
      options: [{ max: 10, maxEOF: 1 }],
      errors: [{ messageId: 'blankEndOfFile', data: { max: 1 }, column: 1 }],
    },
    {
      code: '// invalid 14\nvar a = 5;\n\n',
      output: '// invalid 14\nvar a = 5;\n',
      options: [{ max: 2, maxEOF: 0 }],
      errors: [{ messageId: 'blankEndOfFile', data: { max: 0 }, column: 1 }],
    },
    {
      code: '\n\n// invalid 15\nvar a = 5;\n',
      output: '\n// invalid 15\nvar a = 5;\n',
      options: [{ max: 5, maxBOF: 1 }],
      errors: [{ messageId: 'blankBeginningOfFile', data: { max: 1 }, column: 1 }],
    },
    {
      code: '\n\n\n\n\n// invalid 16\nvar a = 5;\n',
      output: '\n\n\n\n// invalid 16\nvar a = 5;\n',
      options: [{ max: 0, maxBOF: 4 }],
      errors: [{ messageId: 'blankBeginningOfFile', data: { max: 4 }, column: 1 }],
    },
    {
      code: '\n\n// invalid 17\n\n\n\n\n\n\n\n\nvar a = 5;\n',
      output: '\n// invalid 17\n\n\n\n\n\n\n\n\nvar a = 5;\n',
      options: [{ max: 10, maxBOF: 1 }],
      errors: [{ messageId: 'blankBeginningOfFile', data: { max: 1 }, column: 1 }],
    },
    {
      code: '\n// invalid 18\nvar a = 5;\n',
      output: '// invalid 18\nvar a = 5;\n',
      options: [{ max: 2, maxBOF: 0 }],
      errors: [{ messageId: 'blankBeginningOfFile', data: { max: 0 }, column: 1 }],
    },
    {
      code: '\n\n\n// invalid 19\nvar a = 5;\n\n',
      output: '// invalid 19\nvar a = 5;\n',
      options: [{ max: 2, maxBOF: 0, maxEOF: 0 }],
      errors: [{ messageId: 'blankBeginningOfFile', data: { max: 0 }, column: 1 }, { messageId: 'blankEndOfFile', data: { max: 0 }, column: 1 }],
    },
    {
      code: '// invalid 20\r\n// windows line endings\r\nvar a = 5;\r\nvar b = 3;\r\n\r\n\r\n',
      output: '// invalid 20\r\n// windows line endings\r\nvar a = 5;\r\nvar b = 3;\r\n\r\n',
      options: [{ max: 1 }],
      errors: [{ messageId: 'blankEndOfFile', data: { max: 1 }, column: 1 }],
    },
    {
      code: '// invalid 21\n// unix line endings\nvar a = 5;\nvar b = 3;\n\n\n',
      output: '// invalid 21\n// unix line endings\nvar a = 5;\nvar b = 3;\n\n',
      options: [{ max: 1 }],
      errors: [{ messageId: 'blankEndOfFile', data: { max: 1 }, column: 1 }],
    },
    {
      code:
            '\'foo\';\n'
            + '\n'
            + '\n'
            + '`bar`;\n'
            + '`baz`;',
      output:
            '\'foo\';\n'
            + '\n'
            + '`bar`;\n'
            + '`baz`;',
      options: [{ max: 1 }],
      errors: [{ messageId: 'consecutiveBlank', data: { max: 1, pluralizedLines: 'line' }, column: 1 }],
    },
    {
      code: '`template ${foo\n\n\n} literal`;',
      output: '`template ${foo\n\n} literal`;',
      options: [{ max: 1 }],
      errors: [{ messageId: 'consecutiveBlank', data: { max: 1, pluralizedLines: 'line' }, column: 1 }],
    },
    {

      // https://github.com/eslint/eslint/issues/7893
      code: 'a\n\n\n\n' + 'a'.repeat(1e5),
      output: 'a\n\n\n' + 'a'.repeat(1e5),
      errors: [{ messageId: 'consecutiveBlank', data: { max: 2, pluralizedLines: 'lines' }, column: 1 }],
    },
    {

      // https://github.com/eslint/eslint/issues/8401
      code: 'foo\n ',
      output: 'foo\n',
      options: [{ max: 1, maxEOF: 0 }],
      errors: [{ messageId: 'blankEndOfFile', data: { max: 0 }, column: 1 }],
    },
    {

      // https://github.com/eslint/eslint/pull/12594
      code: 'var a;\n\n\n\n\nvar b;',
      output: 'var a;\n\nvar b;',
      options: [{ max: 1 }],
      errors: [{
        messageId: 'consecutiveBlank',
        data: {
          max: 1,
          pluralizedLines: 'line',
        },
        line: 3,
        column: 1,
      }],
    },
    {

      // https://github.com/eslint/eslint/pull/12594
      code: 'var a;\n\n\n\n\nvar b;',
      output: 'var a;\n\n\nvar b;',
      options: [{ max: 2 }],
      errors: [{
        messageId: 'consecutiveBlank',
        data: {
          max: 2,
          pluralizedLines: 'lines',
        },
        line: 4,
        column: 1,
      }],
    },
  ],
});
