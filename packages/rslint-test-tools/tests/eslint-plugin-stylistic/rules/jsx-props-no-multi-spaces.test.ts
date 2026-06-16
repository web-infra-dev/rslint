/**
 * @fileoverview Tests for jsx-props-no-multi-spaces rule.
 *
 * Ported verbatim from @stylistic/eslint-plugin v5.10.0:
 *   packages/eslint-plugin/rules/jsx-props-no-multi-spaces/jsx-props-no-multi-spaces.test.ts
 *
 * Transformations applied per the porting spec:
 *  - `run({ name, rule, parserOptions, valid, invalid })`
 *      -> `ruleTester.run('jsx-props-no-multi-spaces', null as never, { valid, invalid })`
 *  - The upstream `valids(...)` / `invalids(...)` wrappers (`#test/parsers-jsx`)
 *    are a multi-parser fan-out harness (default / `@babel/eslint-parser` /
 *    `@typescript-eslint/parser`) that also appends a `// features: [...],
 *    parser: ...` trailing comment to each fixture. That augmentation is a pure
 *    test-harness artifact, so every case is ported to its bare `code` /
 *    `output` / `errors`, evaluated to its final form.
 *  - `parserOptions` (ecmaFeatures.jsx) dropped — rslint resolves via tsconfig and
 *    the RuleTester routes JSX fixtures to `.tsx`.
 *  - Per-case `features` dropped (see below for the one that carried any).
 *  - The plain backtick multi-line templates are kept byte-for-byte (their literal
 *    leading indentation is significant and is preserved verbatim).
 *
 * Autofix: `@stylistic/jsx-props-no-multi-spaces` is fixable (meta.fixable ===
 * 'code'). The `onlyOneSpace` invalid cases pin an upstream `output` (collapsing
 * the extra spaces); they are ported with that `output`. The `noLineGap` invalid
 * cases pin NO `output` upstream (a blank line between props is reported but not
 * autofixed) — they are ported `output`-less, so the RuleTester checks the
 * diagnostics only and never asserts an invented fix.
 *
 * Feature-tagged case — verified to parse cleanly under ts-go and to match the
 * upstream expectation, so it stays in the green set:
 *  - `<App<T> foo bar />` carried `features: ['ts', 'no-babel']` (upstream runs it
 *    only on the TS parser). ts-go parses the generic JSX element without a syntax
 *    error and the rule emits 0 diagnostics, matching upstream's `valid`
 *    expectation. (Independently confirmed: the 2-space variant `<App<T>  foo bar />`
 *    reports `onlyOneSpace` with prop1 `App`, so the syntax is genuinely parsed.)
 *
 * No `._css_` / `._json_` / `._markdown_` test files exist for this rule, and the
 * single `.test.ts` has exactly one `run()` block (no skipBabel block). No
 * suggestions are pinned anywhere in the upstream file. No KNOWN GAPS surfaced:
 * every fixture parses under ts-go and aligns with upstream.
 */

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('jsx-props-no-multi-spaces', null as never, {
  valid: [
    {
      code: `
        <App />
      `,
    },
    {
      code: `
        <App foo />
      `,
    },
    {
      code: `
        <App foo bar />
      `,
    },
    {
      code: `
        <App foo="with  spaces   " bar />
      `,
    },
    {
      code: `
        <App
          foo bar />
      `,
    },
    {
      code: `
        <App
          foo
          bar />
      `,
    },
    {
      code: `
        <App
          foo {...test}
          bar />
      `,
    },
    {
      code: '<App<T> foo bar />',
    },
    {
      code: '<Foo.Bar baz="quux" />',
    },
    {
      code: '<Foobar.Foo.Bar.Baz.Qux.Quux.Quuz.Corge.Grault.Garply.Waldo.Fred.Plugh xyzzy="thud" />',
    },
    {
      code: `
        <button
          title="Some button"
          type="button"
        />
      `,
    },
    {
      code: `
        <button
          title="Some button"
          onClick={(value) => {
            console.log(value);
          }}
          type="button"
        />
      `,
    },
    {
      code: `
        <button
          title="Some button"
          // this is a comment
          onClick={(value) => {
            console.log(value);
          }}
          type="button"
        />
      `,
    },
    {
      code: `
        <button
          title="Some button"
          // this is a comment
          // this is a second comment
          onClick={(value) => {
            console.log(value);
          }}
          type="button"
        />
      `,
    },
    {
      code: `
        <App
          foo="Some button" // comment
          // comment
          bar=""
        />
      `,
    },
    {
      code: `
        <button
          title="Some button"
          /* this is a multiline comment
              ...
              ... */
          onClick={(value) => {
            console.log(value);
          }}
          type="button"
        />
      `,
    },
  ],

  invalid: [
    {
      code: `
        <App  foo />
      `,
      output: `
        <App foo />
      `,
      errors: [
        {
          messageId: 'onlyOneSpace',
          data: { prop1: 'App', prop2: 'foo' },
        },
      ],
    },
    {
      code: `
        <App foo="with  spaces   "   bar />
      `,
      output: `
        <App foo="with  spaces   " bar />
      `,
      errors: [
        {
          messageId: 'onlyOneSpace',
          data: { prop1: 'foo', prop2: 'bar' },
        },
      ],
    },
    {
      code: `
        <App foo  bar />
      `,
      output: `
        <App foo bar />
      `,
      errors: [
        {
          messageId: 'onlyOneSpace',
          data: { prop1: 'foo', prop2: 'bar' },
        },
      ],
    },
    {
      code: `
        <App  foo   bar />
      `,
      output: `
        <App foo bar />
      `,
      errors: [
        {
          messageId: 'onlyOneSpace',
          data: { prop1: 'App', prop2: 'foo' },
        },
        {
          messageId: 'onlyOneSpace',
          data: { prop1: 'foo', prop2: 'bar' },
        },
      ],
    },
    {
      code: `
        <App foo  {...test}  bar />
      `,
      output: `
        <App foo {...test} bar />
      `,
      errors: [
        {
          messageId: 'onlyOneSpace',
          data: { prop1: 'foo', prop2: 'test' },
        },
        {
          messageId: 'onlyOneSpace',
          data: { prop1: 'test', prop2: 'bar' },
        },
      ],
    },
    {
      code: '<Foo.Bar  baz="quux" />',
      output: '<Foo.Bar baz="quux" />',
      errors: [
        {
          messageId: 'onlyOneSpace',
          data: { prop1: 'Foo.Bar', prop2: 'baz' },
        },
      ],
    },
    {
      code: `
        <Foobar.Foo.Bar.Baz.Qux.Quux.Quuz.Corge.Grault.Garply.Waldo.Fred.Plugh  xyzzy="thud" />
      `,
      output: `
        <Foobar.Foo.Bar.Baz.Qux.Quux.Quuz.Corge.Grault.Garply.Waldo.Fred.Plugh xyzzy="thud" />
      `,
      errors: [
        {
          messageId: 'onlyOneSpace',
          data: { prop1: 'Foobar.Foo.Bar.Baz.Qux.Quux.Quuz.Corge.Grault.Garply.Waldo.Fred.Plugh', prop2: 'xyzzy' },
        },
      ],
    },
    {
      code: `
        <button
          title='Some button'

          type="button"
        />
      `,
      errors: [
        {
          messageId: 'noLineGap',
          data: { prop1: 'title', prop2: 'type' },
        },
      ],
    },
    {
      code: `
        <button
          title="Some button"

          onClick={(value) => {
            console.log(value);
          }}

          type="button"
        />
      `,
      errors: [
        {
          messageId: 'noLineGap',
          data: { prop1: 'title', prop2: 'onClick' },
        },
        {
          messageId: 'noLineGap',
          data: { prop1: 'onClick', prop2: 'type' },
        },
      ],
    },
    {
      code: `
        <button
          title="Some button"
          // this is a comment
          onClick={(value) => {
            console.log(value);
          }}

          type="button"
        />
      `,
      errors: [
        {
          messageId: 'noLineGap',
          data: { prop1: 'onClick', prop2: 'type' },
        },
      ],
    },
    {
      code: `
        <button
          title="Some button"
          // this is a comment
          // second comment

          onClick={(value) => {
            console.log(value);
          }}

          type="button"
        />
      `,
      errors: [
        {
          messageId: 'noLineGap',
          data: { prop1: 'title', prop2: 'onClick' },
        },
        {
          messageId: 'noLineGap',
          data: { prop1: 'onClick', prop2: 'type' },
        },
      ],
    },
    {
      code: `
          <button
            title="Some button"
            /*this is a
              multiline
              comment
            */

            onClick={(value) => {
              console.log(value);
            }}

            type="button"
          />
        `,
      errors: [
        {
          messageId: 'noLineGap',
          data: { prop1: 'title', prop2: 'onClick' },
        },
        {
          messageId: 'noLineGap',
          data: { prop1: 'onClick', prop2: 'type' },
        },
      ],
    },
  ],
});
