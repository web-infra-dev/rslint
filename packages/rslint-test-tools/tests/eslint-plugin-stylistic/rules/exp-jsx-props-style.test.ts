/**
 * @fileoverview Tests for exp-jsx-props-style rule (experimental).
 *
 * Ported verbatim from @stylistic/eslint-plugin v5.10.0:
 *   packages/eslint-plugin/rules/jsx-props-style/jsx-props-style.test.ts
 *
 * The upstream rule lives under the directory `jsx-props-style` but is registered
 * by the plugin as the EXPERIMENTAL rule id `exp-jsx-props-style` (verified against
 * the installed plugin's `Object.keys(plugin.rules)`), so the RuleTester runs it as
 * `'exp-jsx-props-style'`.
 *
 * Transformations applied per the porting spec:
 *  - `run({ name, rule, parserOptions, ... })` ->
 *    `ruleTester.run('exp-jsx-props-style', null as never, { valid, invalid })`.
 *  - `parserOptions` (`ecmaFeatures.jsx: true`) dropped — rslint resolves via
 *    tsconfig (`run`'s default `lang: 'ts'`) and the RuleTester routes JSX
 *    fixtures to `.tsx`.
 *  - The `$` (unindent) template tag is evaluated to its real multi-line string
 *    using the canonical `@antfu/utils` algorithm (drop leading/trailing all-
 *    whitespace lines, strip the common indent). Rendered here as `\n`-escaped
 *    string literals so the `line`/`column`/`endLine`/`endColumn` pins line up
 *    byte-for-byte with the source the rule actually sees.
 *  - `description` fields (upstream test labels) dropped.
 *
 * Expected messages resolve from the plugin's own `meta.messages`:
 *   shouldWrap:    "Prop `{{prop}}` must be placed on a new line"
 *   shouldNotWrap: "Prop `{{prop}}` should not be placed on a new line"
 *
 * The upstream file is a single `run()` block (no `valids()`/`invalids()` parser
 * multiplexer, no `if (!skipBabel)` block, no Babel/Flow cases, no external-fixture
 * `readFileSync` cases, no `suggestions`). Every invalid case pins an explicit
 * `errors` array — there are NO output-only cases. The `._css_` / `._json_` /
 * `._markdown_` test files don't exist for this rule.
 *
 * No case surfaces a rslint<->upstream gap, so nothing is moved to KNOWN GAPS.
 */

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('exp-jsx-props-style', null as never, {
  valid: [
    '<App />',
    '<App foo />',
    '<App foo bar />',
    '<App foo bar baz />',
    '<App {...props} />',
    '<App foo {...props} bar />',
    '<App\n  foo\n  bar\n/>',
    '<App\n  foo\n  bar\n  baz\n/>',
    '<App\n  foo\n  {...props}\n  bar\n/>',
    {
      code: '<App foo />',
      options: [{ singleLine: { maxItems: 1 } }],
    },
    {
      code: '<App\n  foo\n  bar\n/>',
      options: [{ singleLine: { maxItems: 1 } }],
    },
    {
      code: '<App foo bar />',
      options: [{ multiLine: { minItems: 3 } }],
    },
    {
      code: '<App\n  foo\n  bar\n  baz\n/>',
      options: [{ multiLine: { minItems: 3 } }],
    },
    {
      code: '<App foo bar />',
      options: [{ singleLine: { maxItems: 2 } }],
    },
    '<App\n  foo={{\n    a: 1,\n  }}\n  bar\n/>',
    {
      code: '<App\n  foo bar\n  baz qux\n/>',
      options: [{ multiLine: { maxItemsPerLine: 2 } }],
    },
    {
      code: '<App\n  foo bar baz\n/>',
      options: [{ multiLine: { maxItemsPerLine: 3 } }],
    },
  ],

  invalid: [
    {
      code: '<App foo bar />',
      output: '<App\nfoo\nbar />',
      options: [{ singleLine: { maxItems: 1 } }],
      errors: [
        { messageId: 'shouldWrap', data: { prop: 'foo' }, line: 1, column: 6, endLine: 1, endColumn: 9 },
        { messageId: 'shouldWrap', data: { prop: 'bar' }, line: 1, column: 10, endLine: 1, endColumn: 13 },
      ],
    },
    {
      code: '<App foo {...props} bar />',
      output: '<App\nfoo\n{...props}\nbar />',
      options: [{ singleLine: { maxItems: 1 } }],
      errors: [
        { messageId: 'shouldWrap', data: { prop: 'foo' }, line: 1, column: 6, endLine: 1, endColumn: 9 },
        { messageId: 'shouldWrap', data: { prop: 'props' }, line: 1, column: 10, endLine: 1, endColumn: 20 },
        { messageId: 'shouldWrap', data: { prop: 'bar' }, line: 1, column: 21, endLine: 1, endColumn: 24 },
      ],
    },
    {
      code: '<App\n  foo bar\n  baz\n/>',
      output: '<App\n  foo\nbar\n  baz\n/>',
      errors: [
        { messageId: 'shouldWrap', data: { prop: 'bar' }, line: 2, column: 7, endLine: 2, endColumn: 10 },
      ],
    },
    {
      code: '<App\n  foo\n  bar\n/>',
      output: '<App foo bar\n/>',
      options: [{ multiLine: { minItems: 3 } }],
      errors: [
        { messageId: 'shouldNotWrap', data: { prop: 'foo' }, line: 2, column: 3, endLine: 2, endColumn: 6 },
        { messageId: 'shouldNotWrap', data: { prop: 'bar' }, line: 3, column: 3, endLine: 3, endColumn: 6 },
      ],
    },
    {
      code: '<App foo />',
      output: '<App\nfoo />',
      options: [{ singleLine: { maxItems: 0 } }],
      errors: [
        { messageId: 'shouldWrap', data: { prop: 'foo' }, line: 1, column: 6, endLine: 1, endColumn: 9 },
      ],
    },
    {
      code: '<App foo\n  bar\n/>',
      output: '<App foo bar\n/>',
      errors: [
        { messageId: 'shouldNotWrap', data: { prop: 'bar' }, line: 2, column: 3, endLine: 2, endColumn: 6 },
      ],
    },
    {
      code: '<App\n  foo bar\n/>',
      output: '<App\n  foo\nbar\n/>',
      errors: [
        { messageId: 'shouldWrap', data: { prop: 'bar' }, line: 2, column: 7, endLine: 2, endColumn: 10 },
      ],
    },
    {
      code: '<App foo /* comment */ bar />',
      output: '<App\nfoo /* comment */ bar />',
      options: [{ singleLine: { maxItems: 1 } }],
      errors: [
        { messageId: 'shouldWrap', data: { prop: 'foo' } },
        { messageId: 'shouldWrap', data: { prop: 'bar' } },
      ],
    },
    {
      code: '<App\n  foo\n  /* comment */\n  bar\n/>',
      output: '<App foo\n  /* comment */\n  bar\n/>',
      options: [{ multiLine: { minItems: 3 } }],
      errors: [
        { messageId: 'shouldNotWrap', data: { prop: 'foo' } },
        { messageId: 'shouldNotWrap', data: { prop: 'bar' } },
      ],
    },
    {
      code: '<App\n  foo bar baz\n/>',
      output: '<App\n  foo bar\nbaz\n/>',
      options: [{ multiLine: { maxItemsPerLine: 2 } }],
      errors: [
        { messageId: 'shouldWrap', data: { prop: 'baz' } },
      ],
    },
    {
      code: '<App\n  a b c d\n/>',
      output: '<App\n  a b\nc d\n/>',
      options: [{ multiLine: { maxItemsPerLine: 2 } }],
      errors: [
        { messageId: 'shouldWrap', data: { prop: 'c' } },
      ],
    },
  ],
});
