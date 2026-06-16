/**
 * @fileoverview Enforce the consistent use of either double or single quotes in JSX attributes.
 * @author Mathias Schreck <https://github.com/lo1tuma>
 *
 * Ported verbatim from @stylistic/eslint-plugin v5.10.0:
 *   packages/eslint-plugin/rules/jsx-quotes/jsx-quotes.test.ts
 *
 * Transformations applied per the porting spec:
 *  - `run({ name, rule, parserOptions, ... })` ->
 *    `ruleTester.run('jsx-quotes', null as never, { valid, invalid })`
 *  - `parserOptions` (`ecmaFeatures.jsx: true`) dropped — rslint resolves via
 *    tsconfig and the RuleTester routes JSX fixtures to `.tsx`.
 *
 * The expected message resolves from the plugin's own `meta.messages`:
 *   unexpected: "Unexpected usage of {{description}}."
 * (`description` is the data the rule passes: `'singlequote'` / `'doublequote'`.)
 *
 * The upstream file is a plain `run({ ... })` block (NOT wrapped in the
 * `valids()` / `invalids()` parser-multiplexing helpers). It contains NO `$`
 * unindent template tags, NO multi-line template literals, NO `readFileSync`
 * external-fixture cases, NO `suggestions`, and only the single `run()` block
 * above. The `._css_` / `._json_` / `._markdown_` test files don't exist for
 * this rule.
 *
 * Every case is plain JSX that ts-go parses identically to the upstream parser,
 * so no case surfaces a rslint<->upstream gap and nothing is moved to KNOWN GAPS.
 */

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('jsx-quotes', null as never, {
  valid: [
    '<foo bar="baz" />',
    '<foo bar=\'"\' />',
    {
      code: '<foo bar="\'" />',
      options: ['prefer-single'],
    },
    {
      code: '<foo bar=\'baz\' />',
      options: ['prefer-single'],
    },
    '<foo bar="baz">"</foo>',
    {
      code: '<foo bar=\'baz\'>\'</foo>',
      options: ['prefer-single'],
    },
    '<foo bar={\'baz\'} />',
    {
      code: '<foo bar={"baz"} />',
      options: ['prefer-single'],
    },
    '<foo bar={baz} />',
    '<foo bar />',
    {
      code: '<foo bar=\'&quot;\' />',
      options: ['prefer-single'],
    },
    '<foo bar="&quot;" />',
    {
      code: '<foo bar=\'&#39;\' />',
      options: ['prefer-single'],
    },
    '<foo bar="&#39;" />',
  ],
  invalid: [
    {
      code: '<foo bar=\'baz\' />',
      output: '<foo bar="baz" />',
      errors: [
        { messageId: 'unexpected', data: { description: 'singlequote' }, line: 1, column: 10 },
      ],
    },
    {
      code: '<foo bar="baz" />',
      output: '<foo bar=\'baz\' />',
      options: ['prefer-single'],
      errors: [
        { messageId: 'unexpected', data: { description: 'doublequote' }, line: 1, column: 10 },
      ],
    },
    {
      code: '<foo bar="&quot;" />',
      output: '<foo bar=\'&quot;\' />',
      options: ['prefer-single'],
      errors: [
        { messageId: 'unexpected', data: { description: 'doublequote' }, line: 1, column: 10 },
      ],
    },
    {
      code: '<foo bar=\'&#39;\' />',
      output: '<foo bar="&#39;" />',
      errors: [
        { messageId: 'unexpected', data: { description: 'singlequote' }, line: 1, column: 10 },
      ],
    },
  ],
});
