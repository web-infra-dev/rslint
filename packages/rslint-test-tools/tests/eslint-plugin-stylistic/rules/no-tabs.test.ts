/**
 * @fileoverview Tests for no-tabs rule
 * @author Gyandeep Singh
 *
 * Ported verbatim from @stylistic/eslint-plugin v5.10.0:
 *   packages/eslint-plugin/rules/no-tabs/no-tabs.test.ts
 *
 * Transformations applied per the porting spec:
 *  - `run({ name, rule, ... })` -> `ruleTester.run('no-tabs', null as never, { valid, invalid })`
 *    (the `name` / `rule` / `#test` import are dropped).
 *  - `parserOptions` / `type` fields: none present upstream, nothing dropped.
 *  - No `$`/unindent tag, no spread/helper errors, no suggestions upstream.
 *
 * The `\t` (tab) and `\n` (newline) byte sequences in every `code` are preserved
 * EXACTLY as upstream wrote them — single-quoted string literals with escape
 * sequences, and the upstream multi-line `'...\n' + '...'` concatenations kept
 * verbatim (they evaluate to the identical source). This rule is entirely about
 * tab bytes, so a stray normalization would change the test's meaning.
 *
 * `no-tabs` has no autofix (no `fixable` in meta) and no upstream invalid case
 * pins `output`; every invalid case pins an explicit `errors` array, so there
 * are no output-only cases. The `unexpectedTab` messageId renders to the literal
 * "Unexpected tab character." (no `data` interpolation).
 *
 * The upstream `._css_` / `._json_` / `._markdown_` test files don't exist for
 * this rule. No Babel/Flow or external-fixture cases exist upstream, so nothing
 * was skipped on those grounds. All cases are valid TS and run clean under
 * rslint — there are no KNOWN GAPS.
 */

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-tabs', null as never, {
  valid: [
    'function test(){\n}',
    'function test(){\n'
    + '  //   sdfdsf \n'
    + '}',

    {
      code: '\tdoSomething();',
      options: [{ allowIndentationTabs: true }],
    },
    {
      code: '\t// comment',
      options: [{ allowIndentationTabs: true }],
    },
  ],
  invalid: [
    {
      code: 'function test(){\t}',
      errors: [{
        messageId: 'unexpectedTab',
        line: 1,
        column: 17,
        endLine: 1,
        endColumn: 18,
      }],
    },
    {
      code: '/** \t comment test */',
      errors: [{
        messageId: 'unexpectedTab',
        line: 1,
        column: 5,
        endLine: 1,
        endColumn: 6,
      }],
    },
    {
      code:
            'function test(){\n'
            + '  //\tsdfdsf \n'
            + '}',
      errors: [{
        messageId: 'unexpectedTab',
        line: 2,
        column: 5,
        endLine: 2,
        endColumn: 6,
      }],
    },
    {
      code:
            'function\ttest(){\n'
            + '  //sdfdsf \n'
            + '}',
      errors: [{
        messageId: 'unexpectedTab',
        line: 1,
        column: 9,
        endLine: 1,
        endColumn: 10,
      }],
    },
    {
      code:
            'function test(){\n'
            + '  //\tsdfdsf \n'
            + '\t}',
      errors: [
        {
          messageId: 'unexpectedTab',
          line: 2,
          column: 5,
          endLine: 2,
          endColumn: 6,
        },
        {
          messageId: 'unexpectedTab',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 2,
        },
      ],
    },
    {
      code: '\t// Comment with leading tab \t and inline tab',
      options: [{ allowIndentationTabs: true }],
      errors: [{
        messageId: 'unexpectedTab',
        line: 1,
        column: 30,
        endLine: 1,
        endColumn: 31,
      }],
    },
    {
      code: '\t\ta =\t\t\tb +\tc\t\t;\t\t',
      errors: [
        {
          messageId: 'unexpectedTab',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 3,
        },
        {
          messageId: 'unexpectedTab',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 9,
        },
        {
          messageId: 'unexpectedTab',
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 13,
        },
        {
          messageId: 'unexpectedTab',
          line: 1,
          column: 14,
          endLine: 1,
          endColumn: 16,
        },
        {
          messageId: 'unexpectedTab',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
  ],
});
