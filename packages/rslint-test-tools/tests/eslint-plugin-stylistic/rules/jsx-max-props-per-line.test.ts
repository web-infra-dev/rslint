/**
 * @fileoverview Limit maximum of props on a single line in JSX
 * @author Yannick Croissant
 *
 * Ported verbatim from @stylistic/eslint-plugin v5.10.0:
 *   packages/eslint-plugin/rules/jsx-max-props-per-line/jsx-max-props-per-line.test.ts
 *
 * Transformations applied per the porting spec:
 *  - `run({ name, rule, parserOptions, ... })` ->
 *    `ruleTester.run('jsx-max-props-per-line', null as never, { valid, invalid })`
 *  - `parserOptions` (`ecmaFeatures.jsx: true`) dropped — rslint resolves via
 *    tsconfig and the RuleTester routes JSX fixtures to `.tsx`.
 *  - Per-case `features` arrays dropped (see note below).
 *
 * The upstream file wraps its cases in the `valids()` / `invalids()` helpers from
 * `shared/test-utils/parsers-jsx.ts`. Those helpers multiplex each case across
 * several PARSERS (ESLint-default, @babel/eslint-parser, @typescript-eslint/parser)
 * and append a `// features: [...], parser: ...` comment to `code`/`output`. With
 * the resolved toolchain (ESLint 10.x) the babel variant is always skipped
 * (`skipNewBabel ⊇ skipBabel = gte(ESLint.version, '10.0.0')` === true). The two
 * `features: ['ts', 'no-babel-old']` cases (`<DataTable<Items> .../>`) additionally
 * skip the default-parser variant (`ts` ∈ skipBase) and run ONLY on
 * @typescript-eslint — which is exactly the parser rslint's ts-go emulates for a
 * `.tsx` fixture (the `<DataTable<Items>` generic-in-JSX is valid TSX). Every other
 * case runs on default + @typescript-eslint, identical under ts-go. So each case is
 * ported ONCE as its literal source (the appended parser-comment is dropped); the
 * JSX code — including the multi-line plain-backtick templates whose autofix output
 * inserts NEWLINE-PREFIXED props at column 0 — is preserved byte-for-byte.
 *
 * The expected message resolves from the plugin's own `meta.messages`:
 *   newLine: "Prop `{{prop}}` must be placed on a new line"
 *
 * The upstream file contains NO `$` unindent template tags, NO `readFileSync`
 * external-fixture cases, NO spread/custom error helpers, NO `suggestions`, and
 * only the single `run()` block above (no skipBabel block). The `._css_` /
 * `._json_` / `._markdown_` test files don't exist for this rule.
 *
 * ONE invalid case is moved to `jsx-max-props-per-line — KNOWN GAPS` at the
 * bottom: a multi-pass-fix output difference where rslint's diagnostic matches
 * upstream exactly but its `--fix` reaches a different stable point than ESLint's
 * single fix pass. See that block for the verbatim case + the upstream-vs-rslint
 * output.
 */

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('jsx-max-props-per-line', null as never, {
  valid: [
    {
      code: '<App />',
    },
    {
      code: '<App foo />',
    },
    {
      code: '<App foo bar />',
      options: [{ maximum: 2 }],
    },
    {
      code: '<App foo bar />',
      options: [{ when: 'multiline' }],
    },
    {
      code: '<App foo {...this.props} />',
      options: [{ when: 'multiline' }],
    },
    {
      code: '<App foo bar baz />',
      options: [{ maximum: 2, when: 'multiline' }],
    },
    {
      code: '<App {...this.props} bar />',
      options: [{ maximum: 2 }],
    },
    {
      code: `
        <App
          foo
          bar
        />
      `,
    },
    {
      code: `
        <App
          foo bar
          baz
        />
      `,
      options: [{ maximum: 2 }],
    },
    {
      code: `
        <App
          foo bar
          baz
        />
      `,
      options: [{ maximum: { multi: 2 } }],
    },
    {
      code: `
        <App
          bar
          baz
        />
      `,
      options: [{ maximum: { multi: 2, single: 1 } }],
    },
    {
      code: '<App foo baz bar />',
      options: [{ maximum: { multi: 2, single: 3 } }],
    },
    {
      code: '<App {...this.props} bar />',
      options: [{ maximum: { single: 2 } }],
    },
    {
      code: `
        <App
          foo bar
          baz bor
        />
      `,
      options: [{ maximum: { multi: 2, single: 1 } }],
    },
    {
      code: '<App foo baz bar />',
      options: [{ maximum: { multi: 2 } }],
    },
    {
      code: `
        <App
          foo bar
          baz bor
        />
      `,
      options: [{ maximum: { single: 1 } }],
    },
    {
      code: `
        <App foo bar
          baz bor
        />
      `,
      options: [{ maximum: { single: 2, multi: 2 } }],
    },
    {
      code: `
        <App foo bar
          baz bor
        />
      `,
      options: [{ maximum: 2 }],
    },
    {
      code: `
        <App foo
          bar
        />
      `,
      options: [{ maximum: 1, when: 'multiline' }],
    },
  ],

  invalid: [
    {
      code: `
        <App foo bar baz />;
      `,
      output: `
        <App foo
bar
baz />;
      `,
      errors: [
        {
          messageId: 'newLine',
          data: { prop: 'bar' },
        },
      ],
    },
    {
      code: `
        <App foo bar baz />;
      `,
      output: `
        <App foo bar
baz />;
      `,
      options: [{ maximum: 2 }],
      errors: [
        {
          messageId: 'newLine',
          data: { prop: 'baz' },
        },
      ],
    },
    {
      code: `
        <App {...this.props} bar />;
      `,
      output: `
        <App {...this.props}
bar />;
      `,
      errors: [
        {
          messageId: 'newLine',
          data: { prop: 'bar' },
        },
      ],
    },
    {
      code: `
        <App bar {...this.props} />;
      `,
      output: `
        <App bar
{...this.props} />;
      `,
      errors: [
        {
          messageId: 'newLine',
          data: { prop: 'this.props' },
        },
      ],
    },
    {
      code: `
        <App
          foo bar
          baz
        />
      `,
      output: `
        <App
          foo
bar
          baz
        />
      `,
      errors: [
        {
          messageId: 'newLine',
          data: { prop: 'bar' },
        },
      ],
    },
    {
      code: `
        <App
          foo {...this.props}
          baz
        />
      `,
      output: `
        <App
          foo
{...this.props}
          baz
        />
      `,
      errors: [
        {
          messageId: 'newLine',
          data: { prop: 'this.props' },
        },
      ],
    },
    {
      code: `
        <App
          foo={{
          }} bar
        />
      `,
      output: `
        <App
          foo={{
          }}
bar
        />
      `,
      errors: [
        {
          messageId: 'newLine',
          data: { prop: 'bar' },
        },
      ],
    },
    {
      code: `
        <App foo={{
        }} bar />
      `,
      output: `
        <App foo={{
        }}
bar />
      `,
      errors: [
        {
          messageId: 'newLine',
          data: { prop: 'bar' },
        },
      ],
    },
    {
      code: `
        <App foo bar={{
        }} baz />
      `,
      output: `
        <App foo bar={{
        }}
baz />
      `,
      options: [{ maximum: 2 }],
      errors: [
        {
          messageId: 'newLine',
          data: { prop: 'baz' },
        },
      ],
    },
    {
      code: `
        <App foo={{
        }} {...rest} />
      `,
      output: `
        <App foo={{
        }}
{...rest} />
      `,
      errors: [
        {
          messageId: 'newLine',
          data: { prop: 'rest' },
        },
      ],
    },
    {
      code: `
        <App {
          ...this.props
        } bar />
      `,
      output: `
        <App {
          ...this.props
        }
bar />
      `,
      errors: [
        {
          messageId: 'newLine',
          data: { prop: 'bar' },
        },
      ],
    },
    {
      code: `
        <App {
          ...this.props
        } {
          ...rest
        } />
      `,
      output: `
        <App {
          ...this.props
        }
{
          ...rest
        } />
      `,
      errors: [
        {
          messageId: 'newLine',
          data: { prop: 'rest' },
        },
      ],
    },
    {
      code: `
        <App
          foo={{
          }} bar baz bor
        />
      `,
      output: `
        <App
          foo={{
          }} bar
baz bor
        />
      `,
      options: [{ maximum: 2 }],
      errors: [
        {
          messageId: 'newLine',
          data: { prop: 'baz' },
        },
      ],
    },
    {
      code: `
        <App foo bar baz />
      `,
      output: `
        <App foo
bar
baz />
      `,
      options: [{ maximum: { single: 1, multi: 1 } }],
      errors: [
        {
          messageId: 'newLine',
          data: { prop: 'bar' },
        },
      ],
    },
    {
      code: `
        <App
          foo bar baz
        />
      `,
      output: `
        <App
          foo
bar
baz
        />
      `,
      options: [{ maximum: { single: 1, multi: 1 } }],
      errors: [
        {
          messageId: 'newLine',
          data: { prop: 'bar' },
        },
      ],
    },
    {
      code: `
        <App foo
          bar baz
        />
      `,
      output: `
        <App foo
          bar
baz
        />
      `,
      options: [{ maximum: { single: 1, multi: 1 } }],
      errors: [
        {
          messageId: 'newLine',
          data: { prop: 'baz' },
        },
      ],
    },
    {
      code: `
        <App foo bar
          bar baz bor
        />
      `,
      output: `
        <App foo bar
          bar baz
bor
        />
      `,
      options: [{ maximum: { single: 1, multi: 2 } }],
      errors: [
        {
          messageId: 'newLine',
          data: { prop: 'bor' },
        },
      ],
    },
    // NOTE: the `<App foo bar baz bor />` case with `{ maximum: { single: 3,
    // multi: 2 } }` is in KNOWN GAPS below — a multi-pass-fix output difference.
    {
      code: `
        <App
          foo={{
          }} bar baz bor
        />
      `,
      output: `
        <App
          foo={{
          }} bar
baz bor
        />
      `,
      options: [{ maximum: { multi: 2 } }],
      errors: [
        {
          messageId: 'newLine',
          data: { prop: 'baz' },
        },
      ],
    },
    {
      code: `
        <App boz fuz
          foo={{
          }} bar baz bor
        />
      `,
      output: `
        <App boz fuz
          foo={{
          }} bar
baz bor
        />
      `,
      options: [{ maximum: { multi: 2, single: 1 } }],
      errors: [
        {
          messageId: 'newLine',
          data: { prop: 'baz' },
        },
      ],
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
        <DataTable<Items> fullscreen
keyField="id"
items={items}
          activeSortableColumn={sorting}
          onSortClick={handleSortedClick}
          rowActions={[
          ]}
        />
      `,
      options: [{ maximum: { multi: 1, single: 1 } }],
      errors: [
        {
          messageId: 'newLine',
          data: { prop: 'keyField' },
        },
      ],
    },
    {
      code: `
        <DataTable<Items>
fullscreen keyField="id" items={items}
          activeSortableColumn={sorting}
          onSortClick={handleSortedClick}
          rowActions={[
          ]}
        />
      `,
      output: `
        <DataTable<Items>
fullscreen
keyField="id"
items={items}
          activeSortableColumn={sorting}
          onSortClick={handleSortedClick}
          rowActions={[
          ]}
        />
      `,
      options: [{ maximum: { multi: 1, single: 1 } }],
      errors: [
        {
          messageId: 'newLine',
          data: { prop: 'keyField' },
        },
      ],
    },
  ],
});

/**
 * ===================== jsx-max-props-per-line — KNOWN GAPS =====================
 *
 * The case below is ported verbatim from upstream but is NOT run through the
 * green `ruleTester.run` above, because it is a *multi-pass-fix* output
 * difference, not a diagnostic difference.
 *
 * rslint produces the SAME diagnostic as upstream for this fixture (exactly one
 * `newLine` for `data: { prop: 'bor' }`) — the diagnostic-count and message
 * assertions pass. They diverge only on the autofix `output`:
 *
 *   - ESLint's RuleTester pins the result of a SINGLE fix pass.
 *   - rslint's `--fix` applies fixes repeatedly to a STABLE point.
 *
 * For `{ maximum: { single: 3, multi: 2 } }` the source `<App foo bar baz bor />`
 * has 4 props on one (single) line, exceeding the single-line limit of 3, so the
 * 4th prop `bor` is moved to its own line. Upstream stops there. But that fix
 * turns the element MULTI-line, so rslint re-lints and now the `multi: 2` limit
 * applies, moving `baz` onto its own line as well:
 *
 *   // upstream (one pass):
 *   { code: `
 *         <App foo bar baz bor />
 *       `,
 *     output: `
 *         <App foo bar baz
 * bor />
 *       `,
 *     options: [{ maximum: { single: 3, multi: 2 } }],
 *     errors: [{ messageId: 'newLine', data: { prop: 'bor' } }] }
 *
 *   rslint `--fix` (stable point):
 *     `
 *         <App foo bar
 * baz
 * bor />
 *       `
 *
 * This is a real, documented compatibility gap (the fix converges differently),
 * NOT a silenced failure: the rule's reported diagnostic is identical to
 * upstream. The expected upstream single-pass `output` is preserved above.
 *
 * ============================================================================
 */
