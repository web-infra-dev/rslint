/**
 * @fileoverview Validate closing tag location in JSX
 * @author Ross Solomon
 *
 * Ported verbatim from @stylistic/eslint-plugin v5.10.0:
 *   packages/eslint-plugin/rules/jsx-closing-tag-location/jsx-closing-tag-location.test.ts
 *
 * Transformations applied per the porting spec:
 *  - `run({ name, rule, parserOptions, valid, invalid })` ->
 *    `ruleTester.run('jsx-closing-tag-location', null as never, { valid, invalid })`
 *    (the `name`, `rule`, `parserOptions` keys are dropped).
 *  - The upstream `valids(...)` / `invalids(...)` wrappers are NOT identity
 *    helpers — `applyAllParsers` fans each case out across the ESLint parser
 *    matrix (default / `@babel/eslint-parser` / `@typescript-eslint/parser`)
 *    and appends a `// features: […], parser: …, parserOptions: {…}` bookkeeping
 *    comment to `code`/`output`. With ESLint 10.5.0 (`skipBabel = gte(version,
 *    '10.0.0')` is true) the babel slot is always dropped, so every case reduces
 *    to the SAME logical fixture under the default + ts parsers, differing only
 *    by that cosmetic trailing comment. rslint has a single parser (ts-go), so
 *    each logical case is ported once with the harness scaffolding stripped —
 *    identical to how the reference `quotes` port drops `parserOptions`/`lang`.
 *  - `features: ['fragment']` / `features: ['fragment', 'no-ts-old']` dropped:
 *    `fragment` only gates JSX-fragment parser support (ts-go supports `<>…</>`),
 *    and `no-ts-old` excluded only the OLD typescript-eslint parser — ts-go is
 *    unaffected. Both fragment invalid cases were verified to report the correct
 *    messageId and produce the exact upstream `output` under rslint, so they are
 *    GREEN (no gap).
 *  - Error helpers are already plain `{ messageId }` objects; no inlining needed.
 *
 * JSX code (`</Tag>`, `/>`, `<>`) is auto-routed by the RuleTester to a `.tsx`
 * fixture, which ts-go parses correctly. The `._css_` / `._json_` / `._markdown_`
 * test files don't exist for this rule, and there are no Babel/Flow-only,
 * suggestion, or external-fixture cases. No case surfaces a rslint<->upstream
 * gap, so there is no KNOWN GAPS block.
 */

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('jsx-closing-tag-location', null as never, {
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
        <App>foo</App>
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
        <>foo</>
      `,
    },
    {
      code: `
        const foo = () => {
          return <App>
       bar</App>
        }
      `,
      options: ['line-aligned'],
    },
    {
      code: `
        const foo = () => {
          return <App>
              bar</App>
        }
      `,
    },
    {
      code: `
        const foo = () => {
          return <App>
              bar
          </App>
        }
      `,
      options: ['line-aligned'],
    },
    {
      code: `
        const foo = <App>
              bar
        </App>
      `,
      options: ['line-aligned'],
    },
    {
      code: `
        const x = <App>
              foo
                  </App>
      `,
    },
    {
      code: `
        const foo =
          <App>
              bar
          </App>
      `,
      options: ['line-aligned'],
    },
  ],

  invalid: [
    {
      code: `
        <App>
          foo
          </App>
      `,
      output: `
        <App>
          foo
        </App>
      `,
      errors: [{ messageId: 'matchIndent' }],
    },
    {
      code: `
        <App>
          foo</App>
      `,
      output: `
        <App>
          foo
        </App>
      `,
      errors: [{ messageId: 'onOwnLine' }],
    },
    {
      code: `
        <>
          foo
          </>
      `,
      output: `
        <>
          foo
        </>
      `,
      errors: [{ messageId: 'matchIndent' }],
    },
    {
      code: `
        <>
          foo</>
      `,
      output: `
        <>
          foo
        </>
      `,
      errors: [{ messageId: 'onOwnLine' }],
    },
    {
      code: `
        const x = () => {
          return <App>
              foo</App>
        }
      `,
      output: `
        const x = () => {
          return <App>
              foo
          </App>
        }
      `,
      errors: [{ messageId: 'onOwnLine' }],
      options: ['line-aligned'],
    },
    {
      code: `
        const x = <App>
              foo
                  </App>
      `,
      output: `
        const x = <App>
              foo
        </App>
      `,
      errors: [{ messageId: 'alignWithOpening' }],
      options: ['line-aligned'],
    },
  ],
});
