/**
 * @fileoverview Tests for jsx-child-element-spacing
 *
 * Ported verbatim from @stylistic/eslint-plugin v5.10.0:
 *   packages/eslint-plugin/rules/jsx-child-element-spacing/jsx-child-element-spacing.test.ts
 *
 * Transformations applied per the porting spec:
 *  - `run({ name, rule, parserOptions, ... })` ->
 *    `ruleTester.run('jsx-child-element-spacing', null as never, { valid, invalid })`
 *  - `parserOptions` (`ecmaFeatures.jsx`) dropped — rslint resolves via tsconfig
 *    and the RuleTester routes every JSX fixture (these carry `</Tag` / `<>`) to
 *    a `.tsx` file, where ts-go parses JSX correctly.
 *  - The two messageIds carry a `{{element}}` placeholder; the RuleTester renders
 *    them from the plugin's own `meta.messages` with the case's `data`
 *    (`spacingBeforeNext`/`spacingAfterPrev` -> "Ambiguous spacing before/after …
 *    element <name>").
 *
 * The upstream file wraps every case in the `valids()` / `invalids()` helpers from
 * `shared/test-utils/parsers-jsx.ts`. Those helpers multiplex each case across
 * several PARSERS (ESLint-default, @babel/eslint-parser, @typescript-eslint/parser)
 * and append a `// features: [...], parser: ...` comment to `code`/`output`. With
 * the resolved toolchain (ESLint 10.x) the babel variant is skipped
 * (`skipBabel = gte(ESLint.version, '10.0.0')` === true). The two `features:
 * ['fragment']` cases additionally skip the default-parser variant
 * (`skipBase` includes `fragment`), so they run only under @typescript-eslint —
 * exactly ts-go territory. That parser-multiplexing is a pure upstream-harness
 * artifact with no rslint analog, so each case is ported ONCE as its literal
 * source (the `features` field and the appended parser-comment are dropped); the
 * code itself is verbatim, including the leading newline + indentation of every
 * plain backtick template (load-bearing: the `line`/`column` pins are computed
 * against that exact indented source).
 *
 * There are NO `$` unindent template tags, NO `readFileSync` external-fixture
 * cases, NO output-only invalid cases (every invalid pins `errors`), and NO
 * `suggestions`. The `._css_` / `._json_` / `._markdown_` test files don't exist
 * for this rule.
 *
 * No case surfaces a rslint<->upstream gap, so nothing is moved to KNOWN GAPS.
 */

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('jsx-child-element-spacing', null as never, {
  valid: [
    {
      code: `
        <App>
          foo
        </App>
      `,
    },
    {
      code: `
        <>
          foo
        </>
      `,
    },
    {
      code: `
        <App>
          <a>bar</a>
        </App>
      `,
    },
    {
      code: `
        <App>
          <a>
            <b>nested</b>
          </a>
        </App>
      `,
    },
    {
      code: `
        <App>
          foo
          bar
        </App>
      `,
    },
    {
      code: `
        <App>
          foo<a>bar</a>baz
        </App>
      `,
    },
    {
      code: `
        <App>
          foo
          {' '}
          <a>bar</a>
          {' '}
          baz
        </App>
      `,
    },
    {
      code: `
        <App>
          foo
          {' '}<a>bar</a>{' '}
          baz
        </App>
      `,
    },
    {
      code: `
        <App>
          foo{' '}
          <a>bar</a>
          {' '}baz
        </App>
      `,
    },
    {
      code: `
        <App>
          foo{/*
          */}<a>bar</a>{/*
          */}baz
        </App>
      `,
    },
    {
      code: `
        <App>
          Please take a look at <a href="https://js.org">this link</a>.
        </App>
      `,
    },
    {
      code: `
        <App>
          Please take a look at
          {' '}
          <a href="https://js.org">this link</a>.
        </App>
      `,
    },
    {
      code: `
        <App>
          <p>A</p>
          <p>B</p>
        </App>
      `,
    },
    {
      code: `
        <App>
          <p>A</p><p>B</p>
        </App>
      `,
    },
    {
      code: `
        <App>
          <a>foo</a>
          <a>bar</a>
        </App>
      `,
    },
    {
      code: `
        <App>
          <a>
            <b>nested1</b>
            <b>nested2</b>
          </a>
        </App>
      `,
    },
    {
      code: `
        <App>
          A
          B
        </App>
      `,
    },
    {
      code: `
        <App>
          A
          <br/>
          B
        </App>
      `,
    },
    {
      code: `
        <App>
          A<br/>
          B
        </App>
      `,
    },
    {
      code: `
        <App>
          A<br/>B
        </App>
      `,
    },
    {
      code: `
        <App>A<br/>B</App>
      `,
    },
  ],

  invalid: [
    {
      code: `
        <App>
          foo
          <a>bar</a>
        </App>
      `,
      errors: [
        {
          messageId: 'spacingBeforeNext',
          data: { element: 'a' },
          line: 4,
          column: 11,
        },
      ],
    },
    {
      code: `
        <>
          foo
          <a>bar</a>
        </>
      `,
      errors: [
        {
          messageId: 'spacingBeforeNext',
          data: { element: 'a' },
          line: 4,
          column: 11,
        },
      ],
    },
    {
      code: `
        <App>
          <a>bar</a>
          baz
        </App>
      `,
      errors: [
        {
          messageId: 'spacingAfterPrev',
          data: { element: 'a' },
          line: 3,
          column: 21,
        },
      ],
    },
    {
      code: `
        <App>
          {' '}<a>bar</a>
          baz
        </App>
      `,
      errors: [
        {
          messageId: 'spacingAfterPrev',
          data: { element: 'a' },
          line: 3,
          column: 26,
        },
      ],
    },
    {
      code: `
        <App>
          Please take a look at
          <a href="https://js.org">this link</a>.
        </App>
      `,
      errors: [
        {
          messageId: 'spacingBeforeNext',
          data: { element: 'a' },
          line: 4,
          column: 11,
        },
      ],
    },
    {
      code: `
        <App>
          Some <code>loops</code> and some
          <code>if</code> statements.
        </App>
      `,
      errors: [
        {
          messageId: 'spacingBeforeNext',
          data: { element: 'code' },
          line: 4,
          column: 11,
        },
      ],
    },
    {
      code: `
        <App>
          Here is
          <a href="https://js.org">a link</a> and here is
          <a href="https://js.org">another</a>
        </App>
      `,
      errors: [
        {
          messageId: 'spacingBeforeNext',
          data: { element: 'a' },
          line: 4,
          column: 11,
        },
        {
          messageId: 'spacingBeforeNext',
          data: { element: 'a' },
          line: 5,
          column: 11,
        },
      ],
    },
  ],
});
