/**
 * @fileoverview Disallow or enforce spaces around equal signs in JSX attributes.
 * @author ryym
 *
 * Ported verbatim from @stylistic/eslint-plugin v5.10.0:
 *   packages/eslint-plugin/rules/jsx-equals-spacing/jsx-equals-spacing.test.ts
 *
 * Transformations applied per the porting spec:
 *  - `run({ name, rule, parserOptions, ... })` ->
 *    `ruleTester.run('jsx-equals-spacing', null as never, { valid, invalid })`
 *  - `parserOptions` (`ecmaFeatures.jsx: true`) dropped — rslint resolves via
 *    tsconfig and the RuleTester routes JSX fixtures to `.tsx`.
 *
 * The upstream file wraps its cases in the `valids()` / `invalids()` helpers from
 * `shared/test-utils/parsers-jsx.ts`. Those helpers multiplex each case across
 * several PARSERS (ESLint-default, @babel/eslint-parser, @typescript-eslint/parser)
 * and append a `// features: [...], parser: ...` comment to `code`/`output`. With
 * the resolved toolchain (ESLint 10.5.0) the babel variant is skipped
 * (`skipBabel = gte(ESLint.version, '10.0.0')` === true), leaving only the
 * default + @typescript-eslint variants — which are identical under rslint's
 * single ts-go parser. That parser-multiplexing is a pure upstream-harness
 * artifact with no rslint analog, so each case is ported ONCE as its literal
 * source (the appended parser-comment is dropped); the JSX code is verbatim and
 * runs as a `.tsx` fixture.
 *
 * The expected messages resolve from the plugin's own `meta.messages`:
 *   noSpaceBefore:   "There should be no space before '='"
 *   noSpaceAfter:    "There should be no space after '='"
 *   needSpaceBefore: "A space is required before '='"
 *   needSpaceAfter:  "A space is required after '='"
 *
 * The upstream file contains NO `$` unindent template tags, NO multi-line
 * template literals, NO `readFileSync` external-fixture cases, NO `suggestions`,
 * and only the single `run()` block above. The `._css_` / `._json_` /
 * `._markdown_` test files don't exist for this rule.
 *
 * No case surfaces a rslint<->upstream gap, so nothing is moved to KNOWN GAPS.
 */

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('jsx-equals-spacing', null as never, {
  valid: [
    {
      code: '<App />',
    },
    {
      code: '<App foo />',
    },
    {
      code: '<App foo="bar" />',
    },
    {
      code: '<App foo={e => bar(e)} />',
    },
    {
      code: '<App {...props} />',
    },
    {
      code: '<App />',
      options: ['never'],
    },
    {
      code: '<App foo />',
      options: ['never'],
    },
    {
      code: '<App foo="bar" />',
      options: ['never'],
    },
    {
      code: '<App foo={e => bar(e)} />',
      options: ['never'],
    },
    {
      code: '<App {...props} />',
      options: ['never'],
    },
    {
      code: '<App />',
      options: ['always'],
    },
    {
      code: '<App foo />',
      options: ['always'],
    },
    {
      code: '<App foo = "bar" />',
      options: ['always'],
    },
    {
      code: '<App foo = {e => bar(e)} />',
      options: ['always'],
    },
    {
      code: '<App {...props} />',
      options: ['always'],
    },
  ],

  invalid: [
    {
      code: '<App foo = {bar} />',
      output: '<App foo={bar} />',
      errors: [
        { messageId: 'noSpaceBefore' },
        { messageId: 'noSpaceAfter' },
      ],
    },
    {
      code: '<App foo = {bar} />',
      output: '<App foo={bar} />',
      options: ['never'],
      errors: [
        { messageId: 'noSpaceBefore' },
        { messageId: 'noSpaceAfter' },
      ],
    },
    {
      code: '<App foo ={bar} />',
      output: '<App foo={bar} />',
      options: ['never'],
      errors: [{ messageId: 'noSpaceBefore' }],
    },
    {
      code: '<App foo= {bar} />',
      output: '<App foo={bar} />',
      options: ['never'],
      errors: [{ messageId: 'noSpaceAfter' }],
    },
    {
      code: '<App foo= {bar} bar = {baz} />',
      output: '<App foo={bar} bar={baz} />',
      options: ['never'],
      errors: [
        { messageId: 'noSpaceAfter' },
        { messageId: 'noSpaceBefore' },
        { messageId: 'noSpaceAfter' },
      ],
    },
    {
      code: '<App foo={bar} />',
      output: '<App foo = {bar} />',
      options: ['always'],
      errors: [
        { messageId: 'needSpaceBefore' },
        { messageId: 'needSpaceAfter' },
      ],
    },
    {
      code: '<App foo ={bar} />',
      output: '<App foo = {bar} />',
      options: ['always'],
      errors: [{ messageId: 'needSpaceAfter' }],
    },
    {
      code: '<App foo= {bar} />',
      output: '<App foo = {bar} />',
      options: ['always'],
      errors: [{ messageId: 'needSpaceBefore' }],
    },
    {
      code: '<App foo={bar} bar ={baz} />',
      output: '<App foo = {bar} bar = {baz} />',
      options: ['always'],
      errors: [
        { messageId: 'needSpaceBefore' },
        { messageId: 'needSpaceAfter' },
        { messageId: 'needSpaceAfter' },
      ],
    },
  ],
});
