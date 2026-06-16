/**
 * @fileoverview Ensure proper position of the first property in JSX.
 * @author Joachim Seminck
 *
 * Ported verbatim from @stylistic/eslint-plugin v5.10.0:
 *   packages/eslint-plugin/rules/jsx-first-prop-new-line/jsx-first-prop-new-line.test.ts
 *
 * Transformations applied per the porting spec:
 *  - `run({ name, rule, parserOptions, ... })` ->
 *    `ruleTester.run('jsx-first-prop-new-line', null as never, { valid, invalid })`
 *  - `parserOptions` (`ecmaFeatures.jsx: true`) dropped — rslint resolves via
 *    tsconfig and the RuleTester routes JSX fixtures to `.tsx`.
 *  - Per-case `features` arrays dropped (see note below).
 *
 * The upstream file wraps its cases in the `valids()` / `invalids()` helpers from
 * `shared/test-utils/parsers-jsx.ts`. Those helpers multiplex each case across
 * several PARSERS (ESLint-default, @babel/eslint-parser, @typescript-eslint/parser)
 * and append a `// features: [...], parser: ...` comment to `code`/`output`. With
 * the resolved toolchain (ESLint 10.5.0) the babel variant is always skipped
 * (`skipBabel = gte(ESLint.version, '10.0.0')` === true). The lone `features:
 * ['ts', 'no-babel-old']` case (`<DataTable<Items> .../>`) additionally skips the
 * default-parser variant (`ts` ∈ skipBase) and runs ONLY on @typescript-eslint —
 * which is exactly the parser rslint's ts-go emulates for a `.tsx` fixture. Every
 * other case runs on default + @typescript-eslint, identical under ts-go. So each
 * case is ported ONCE as its literal source (the appended parser-comment is
 * dropped); the JSX code, including the multi-line plain-backtick templates with
 * their leading newline + indentation, is preserved verbatim and runs as `.tsx`.
 *
 * The expected messages resolve from the plugin's own `meta.messages`:
 *   propOnNewLine:  "Property should be placed on a new line"
 *   propOnSameLine: "Property should be placed on the same line as the component declaration"
 *
 * The upstream file contains NO `$` unindent template tags (the multi-line cases
 * use plain backtick literals — indentation preserved verbatim), NO `readFileSync`
 * external-fixture cases, NO `suggestions`, and only the single `run()` block
 * above. The `._css_` / `._json_` / `._markdown_` test files don't exist for this
 * rule.
 *
 * No case surfaces a rslint<->upstream gap, so nothing is moved to KNOWN GAPS.
 */

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('jsx-first-prop-new-line', null as never, {
  valid: [
    {
      code: '<Foo />',
      options: ['never'],
    },
    {
      code: '<Foo prop="bar" />',
      options: ['never'],
    },
    {
      code: '<Foo {...this.props} />',
      options: ['never'],
    },
    {
      code: '<Foo a a a />',
      options: ['never'],
    },
    {
      code: `
        <Foo a
          b
        />
      `,
      options: ['never'],
    },
    {
      code: '<Foo />',
      options: ['multiline'],
    },
    {
      code: '<Foo prop="one" />',
      options: ['multiline'],
    },
    {
      code: '<Foo {...this.props} />',
      options: ['multiline'],
    },
    {
      code: '<Foo a a a />',
      options: ['multiline'],
    },
    {
      code: `
        <Foo
          propOne="one"
          propTwo="two"
        />
      `,
      options: ['multiline'],
    },
    {
      code: `
        <Foo
          {...this.props}
          propTwo="two"
        />
      `,
      options: ['multiline'],
    },
    {
      code: `
        <Foo bar />
      `,
      options: ['multiline-multiprop'],
    },
    {
      code: `
        <Foo bar baz />
      `,
      options: ['multiline-multiprop'],
    },
    {
      code: `
        <Foo prop={{
        }} />
      `,
      options: ['multiline-multiprop'],
    },
    {
      code: `
        <Foo
          foo={{
          }}
          bar
        />
      `,
      options: ['multiline-multiprop'],
    },
    {
      code: '<Foo />',
      options: ['always'],
    },
    {
      code: `
        <Foo
          propOne="one"
          propTwo="two"
        />
      `,
      options: ['always'],
    },
    {
      code: `
        <Foo
          {...this.props}
          propTwo="two"
        />
      `,
      options: ['always'],
    },
    {
      code: `
        <Foo />
      `,
      options: ['multiprop'],
    },
    {
      code: `
        <Foo bar />
      `,
      options: ['multiprop'],
    },
    {
      code: `
        <Foo {...this.props} />
      `,
      options: ['multiprop'],
    },
  ],

  invalid: [
    {
      code: `
        <Foo propOne="one" propTwo="two" />
      `,
      output: `
        <Foo
propOne="one" propTwo="two" />
      `,
      options: ['always'],
      errors: [{ messageId: 'propOnNewLine' }],
    },
    {
      code: `
        <Foo propOne="one"
          propTwo="two"
        />
      `,
      output: `
        <Foo
propOne="one"
          propTwo="two"
        />
      `,
      options: ['always'],
      errors: [{ messageId: 'propOnNewLine' }],
    },
    {
      code: `
        <Foo
          propOne="one"
          propTwo="two"
        />
      `,
      output: `
        <Foo propOne="one"
          propTwo="two"
        />
      `,
      options: ['never'],
      errors: [{ messageId: 'propOnSameLine' }],
    },
    {
      code: `
        <Foo prop={{
        }} />
      `,
      output: `
        <Foo
prop={{
        }} />
      `,
      options: ['multiline'],
      errors: [{ messageId: 'propOnNewLine' }],
    },
    {
      code: `
        <Foo bar={{
        }} baz />
      `,
      output: `
        <Foo
bar={{
        }} baz />
      `,
      options: ['multiline-multiprop'],
      errors: [{ messageId: 'propOnNewLine' }],
    },
    {
      code: `
      <Foo propOne="one" propTwo="two" />
      `,
      output: `
      <Foo
propOne="one" propTwo="two" />
      `,
      options: ['multiprop'],
      errors: [{ messageId: 'propOnNewLine' }],
    },
    {
      code: `
      <Foo
bar />
      `,
      output: `
      <Foo bar />
      `,
      options: ['multiprop'],
      errors: [{ messageId: 'propOnSameLine' }],
    },
    {
      code: `
      <Foo
{...this.props} />
      `,
      output: `
      <Foo {...this.props} />
      `,
      options: ['multiprop'],
      errors: [{ messageId: 'propOnSameLine' }],
    },
    {
      code: `
        <DataTable<Items> fullscreen keyField="id" items={items}
          activeSortableColumn={sorting}
          onSortClick={handleSortedClick}
          rowActions={[
          ]}
        />
      `,
      output: `
        <DataTable<Items>
fullscreen keyField="id" items={items}
          activeSortableColumn={sorting}
          onSortClick={handleSortedClick}
          rowActions={[
          ]}
        />
      `,
      options: ['multiline'],
      errors: [{ messageId: 'propOnNewLine' }],
    },
  ],
});
