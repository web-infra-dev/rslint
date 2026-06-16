/**
 * @fileoverview Tests for jsx-function-call-newline rule.
 *
 * Ported verbatim from @stylistic/eslint-plugin v5.10.0:
 *   packages/eslint-plugin/rules/jsx-function-call-newline/jsx-function-call-newline.test.ts
 *
 * Transformations applied per the porting spec:
 *  - `run({ name, rule, parserOptions, valid, invalid })`
 *    -> `ruleTester.run('jsx-function-call-newline', null as never, { valid, invalid })`
 *    (the `name`, `rule`, `parserOptions` keys are dropped).
 *  - Upstream wraps its arrays in `valids(...)` / `invalids(...)` (from
 *    `#test/parsers-jsx`). Those helpers MULTIPLY every case across three parsers
 *    (default ESLint, `@babel/eslint-parser`, `@typescript-eslint/parser`) and
 *    append a `// features: [...], parser: ...` comment to `code`/`output`. That
 *    is upstream test-harness machinery, NOT rule semantics — rslint runs the
 *    single ts-go path (equivalent to the `@typescript-eslint/parser` variant), so
 *    each upstream entry is ported as ONE case with NO appended parser comment.
 *  - `parserOptions.ecmaFeatures.jsx` dropped — rslint resolves JSX via tsconfig;
 *    the RuleTester routes JSX code to a `.tsx` fixture.
 *  - The single messageId `missingLineBreak` carries no `data`, so its message is
 *    asserted from the rule's own `meta.messages` (no interpolation needed).
 *  - All `code`/`output` are plain backtick template literals whose `\n` escapes
 *    are REAL newlines in the string — preserved byte-for-byte.
 *
 * No `features`-flagged, Babel/Flow, `$`-unindent, suggestion, or external-fixture
 * (`readFileSync`) cases exist in the upstream jsx-function-call-newline test, so
 * nothing was skipped or isolated on those grounds. The `._css_` / `._json_` /
 * `._markdown_` test files don't exist for this rule.
 *
 * No case surfaced a real rslint<->upstream gap, so there is no KNOWN GAPS block.
 */

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('jsx-function-call-newline', null as never, {
  valid: [
    {
      code: `fn(<div />)`,
    },
    {
      code: `fn(<div />, <div />)`,
    },
    {
      code: `fn(<div />,\n<div />)`,
    },
    {
      code: `fn(\n<div />, <div />)`,
    },
    {
      code: `fn(\n<div />, <div />\n)`,
    },
    {
      code: `fn(\n<div />\n)`,
      options: ['always'],
    },
    {
      code: `fn(<div />, \n<div \n style={{ color: 'red' }}\n />\n)`,
    },
    {
      code: `fn(<div />, <div />, <div />)`,
    },
    {
      code: `fn(<div />, <div />\n, <div />)`,
    },
    {
      code: `fn(\n<div />\n,\n<div />\n,\n<div />\n)`,
    },
    {
      code: `fn(\n<div />\n,\n<div />\n,\n<div />\n)`,
      options: ['always'],
    },
    {
      code: `fn(\n<div />\n,\n<div ></div>)`,
    },
    {
      code: `fn((<div style={{}} />), <div />, <div />)`,
    },
    {
      code: `new OBJ((<div style={{}} />), <div />, <div />)`,
    },
    {
      code: `new OBJ(<div />, <div />, <div />)`,
    },
    {
      code: `new OBJ(<div />, <div />\n, <div />)`,
    },
    {
      code: `new OBJ(\n<div />\n,\n<div />\n,\n<div />\n)`,
    },
    {
      code: `new OBJ(\n<div />\n,\n<div />\n,\n<div />\n)`,
      options: ['always'],
    },
    {
      code: `new OBJ(\n<div />\n,\n<div ></div>)`,
    },
  ],
  invalid: [
    {
      code: `fn(<div
        />)`,
      output: `fn(\n<div
        />\n)`,
      errors: [
        { messageId: 'missingLineBreak' },
      ],
    },
    {
      code: `new OBJ(<div
        />)`,
      output: `new OBJ(\n<div
        />\n)`,
      errors: [
        { messageId: 'missingLineBreak' },
      ],
    },
    {
      code: `fn(<div />)`,
      output: `fn(\n<div />\n)`,
      errors: [
        { messageId: 'missingLineBreak' },
      ],
      options: ['always'],
    },
    {
      code: `fn(\n<div />,<div />,\n<div />)`,
      output: `fn(\n<div />,\n<div />,\n<div />\n)`,
      errors: [
        { messageId: 'missingLineBreak' },
        { messageId: 'missingLineBreak' },
      ],
      options: ['always'],
    },
    {
      code: `new OBJ(\n<div />,<div />,\n<div />)`,
      output: `new OBJ(\n<div />,\n<div />,\n<div />\n)`,
      errors: [
        { messageId: 'missingLineBreak' },
        { messageId: 'missingLineBreak' },
      ],
      options: ['always'],
    },
    {
      code: `fn((\n<div />),<div />,\n<div />)`,
      output: `fn((\n<div />\n),\n<div />,\n<div />\n)`,
      errors: [
        { messageId: 'missingLineBreak' },
        { messageId: 'missingLineBreak' },
        { messageId: 'missingLineBreak' },
      ],
      options: ['always'],
    },
    {
      code: `fn(<div />, <span>\n</span>)`,
      output: `fn(<div />, \n<span>\n</span>\n)`,
      errors: [
        { messageId: 'missingLineBreak' },
      ],
    },
    {
      code: `fn(<div \n />, <span>\n</span>)`,
      output: `fn(\n<div \n />, \n<span>\n</span>\n)`,
      errors: [
        { messageId: 'missingLineBreak' },
        { messageId: 'missingLineBreak' },
      ],
    },
  ],
});
