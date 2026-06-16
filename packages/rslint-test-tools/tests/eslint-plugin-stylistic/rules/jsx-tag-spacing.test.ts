/**
 * @fileoverview Tests for jsx-tag-spacing
 *
 * Ported verbatim from @stylistic/eslint-plugin v5.10.0:
 *   packages/eslint-plugin/rules/jsx-tag-spacing/jsx-tag-spacing.test.ts
 *
 * Transformations applied per the porting spec:
 *  - `run({ name, rule, parserOptions, ... })` ->
 *    `ruleTester.run('jsx-tag-spacing', null as never, { valid, invalid })`
 *  - `parserOptions` (`ecmaFeatures.jsx`) dropped — rslint resolves via tsconfig
 *    and the RuleTester routes JSX fixtures to a `.tsx` file (where ts-go parses
 *    JSX correctly) by detecting a `</Tag` / `/>` / `<>` marker in the code. This
 *    rule deliberately mangles that very spacing, so several cases (`<p/ >`,
 *    `</ div>`, `<App/ >;`, the `/`-then-newline-`>` self-close, `< App></ App>`,
 *    `<App></ App>`, `<App2></ App2>`, `< App / >`) carry `/ >` or `</ Tag` and
 *    miss the marker. Each such case pins `filename: 'case.tsx'` to force the
 *    `.tsx` route — these are all valid TSX (verified), so this is purely the
 *    rslint equivalent of upstream's `ecmaFeatures.jsx: true`, with the code left
 *    verbatim. (The RuleTester's `needsJsx` heuristic is intentionally narrow to
 *    avoid mis-routing TS generics, and is not modified.)
 *  - The four option-builder helpers (`closingSlashOptions`,
 *    `beforeSelfClosingOptions`, `afterOpeningOptions`, `beforeClosingOptions`)
 *    are inlined to the literal option object each returns — every one disables
 *    the three checks it does not test (`'allow'`) and sets only the tested one.
 *  - All `meta.messages` for this rule are STATIC strings (no `{{placeholder}}`),
 *    so the RuleTester asserts each rendered message directly; no `data` needed.
 *  - Two invalid outputs embed `${' '}` (a deliberate trailing space upstream
 *    used so an editor wouldn't strip it). The `${' '}` template form is kept
 *    verbatim — it evaluates to that exact trailing space, which is load-bearing
 *    for the autofix `output` assertion.
 *
 * The upstream file wraps every case in the `valids()` / `invalids()` helpers
 * from `shared/test-utils/parsers-jsx.ts`. Those helpers multiplex each case
 * across several PARSERS (ESLint-default, @babel/eslint-parser,
 * @typescript-eslint/parser) and append a `// features: [...], parser: ...`
 * comment to `code`/`output`. With the resolved toolchain (ESLint 10.x) the
 * babel variant is skipped (`skipBabel = gte(ESLint.version, '10.0.0')` ===
 * true). That parser-multiplexing is a pure upstream-harness artifact with no
 * rslint analog, so each case is ported ONCE as its literal source (the
 * `features` field and the appended parser-comment are dropped); the code itself
 * is verbatim, including the leading newline + indentation of every plain
 * backtick template.
 *
 * `features: ['no-ts']` (upstream skips the @typescript-eslint parser for that
 * case) marked five cases. Two of them — `< App></ App>` (valid) and
 * `<App prop="foo"></App>` → `<App prop="foo">< /App>` (invalid, closeSlashNeed-
 * Space) — were verified to parse and behave identically under ts-go (the `</ Tag>`
 * "space AFTER the slash" form is valid TSX, and the invalid case's INPUT parses
 * even though its fixed output does not), so they stay in the green set. The
 * other three carry the `< /` "space BETWEEN `<` and `/`" form in their INPUT,
 * which ts-go rejects as a syntax error (TS1003 "Identifier expected" / TS1382
 * "Unexpected token"); they are moved to KNOWN GAPS below.
 *
 * There are NO `$` unindent template tags, NO `readFileSync` external-fixture
 * cases, NO output-only invalid cases (every invalid pins `errors`), and NO
 * `suggestions`. The `._css_` / `._json_` / `._markdown_` test files don't exist
 * for this rule.
 */

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('jsx-tag-spacing', null as never, {
  valid: [
    {
      code: '<App />',
    },
    {
      code: '<App />',
      options: [{ closingSlash: 'allow', beforeSelfClosing: 'always', afterOpening: 'allow', beforeClosing: 'allow' }],
    },
    {
      code: '<App foo />',
      options: [{ closingSlash: 'allow', beforeSelfClosing: 'always', afterOpening: 'allow', beforeClosing: 'allow' }],
    },
    {
      code: '<App foo={bar} />',
      options: [{ closingSlash: 'allow', beforeSelfClosing: 'always', afterOpening: 'allow', beforeClosing: 'allow' }],
    },
    {
      code: '<App {...props} />',
      options: [{ closingSlash: 'allow', beforeSelfClosing: 'always', afterOpening: 'allow', beforeClosing: 'allow' }],
    },
    {
      code: '<App></App>',
      options: [{ closingSlash: 'allow', beforeSelfClosing: 'always', afterOpening: 'allow', beforeClosing: 'allow' }],
    },
    {
      code: `
        <App
          foo={bar}
        />
      `,
      options: [{ closingSlash: 'allow', beforeSelfClosing: 'always', afterOpening: 'allow', beforeClosing: 'allow' }],
    },
    {
      code: '<App/>',
      options: [{ closingSlash: 'allow', beforeSelfClosing: 'never', afterOpening: 'allow', beforeClosing: 'allow' }],
    },
    {
      code: '<App />',
      options: [{ closingSlash: 'allow', beforeSelfClosing: 'proportional-always', afterOpening: 'allow', beforeClosing: 'allow' }],
    },
    {
      code: '<App foo />',
      options: [{ closingSlash: 'allow', beforeSelfClosing: 'proportional-always', afterOpening: 'allow', beforeClosing: 'allow' }],
    },
    {
      code: `
        <App
          foo={bar}
          blat
        >
          hello
        </App>
      `,
      options: [{ closingSlash: 'allow', beforeSelfClosing: 'allow', afterOpening: 'allow', beforeClosing: 'proportional-always' }],
    },
    {
      code: `
        <App foo={bar}>
          hello
        </App>
      `,
      options: [{ closingSlash: 'allow', beforeSelfClosing: 'allow', afterOpening: 'allow', beforeClosing: 'proportional-always' }],
    },
    {
      code: `
        <App
          foo={bar}
        />
      `,
      options: [{ closingSlash: 'allow', beforeSelfClosing: 'proportional-always', afterOpening: 'allow', beforeClosing: 'allow' }],
    },
    {
      code: '<App foo/>',
      options: [{ closingSlash: 'allow', beforeSelfClosing: 'never', afterOpening: 'allow', beforeClosing: 'allow' }],
    },
    {
      code: '<App foo={bar}/>',
      options: [{ closingSlash: 'allow', beforeSelfClosing: 'never', afterOpening: 'allow', beforeClosing: 'allow' }],
    },
    {
      code: '<App {...props}/>',
      options: [{ closingSlash: 'allow', beforeSelfClosing: 'never', afterOpening: 'allow', beforeClosing: 'allow' }],
    },
    {
      code: '<App></App>',
      options: [{ closingSlash: 'allow', beforeSelfClosing: 'never', afterOpening: 'allow', beforeClosing: 'allow' }],
    },
    {
      code: `
        <App
          foo={bar}
        />
      `,
      options: [{ closingSlash: 'allow', beforeSelfClosing: 'never', afterOpening: 'allow', beforeClosing: 'allow' }],
    },
    {
      code: '<App/>;',
      options: [{ closingSlash: 'never', beforeSelfClosing: 'allow', afterOpening: 'allow', beforeClosing: 'allow' }],
    },
    {
      code: '<App />;',
      options: [{ closingSlash: 'never', beforeSelfClosing: 'allow', afterOpening: 'allow', beforeClosing: 'allow' }],
    },
    {
      code: '<div className="bar"></div>;',
      options: [{ closingSlash: 'never', beforeSelfClosing: 'allow', afterOpening: 'allow', beforeClosing: 'allow' }],
    },
    {
      // `</ div>` (space after the slash) is valid TSX but lacks the `</Tag`/`/>`
      // marker the RuleTester's JSX auto-detect keys on, so pin a `.tsx` filename.
      code: '<div className="bar"></ div>;',
      filename: 'case.tsx',
      options: [{ closingSlash: 'never', beforeSelfClosing: 'allow', afterOpening: 'allow', beforeClosing: 'allow' }],
    },
    {
      // `/ >` (space before `>`) lacks the `/>` marker → pin `.tsx`.
      code: '<p/ >',
      filename: 'case.tsx',
      options: [{ closingSlash: 'always', beforeSelfClosing: 'allow', afterOpening: 'allow', beforeClosing: 'allow' }],
    },
    {
      code: '<App/>',
      options: [{ closingSlash: 'allow', beforeSelfClosing: 'allow', afterOpening: 'never', beforeClosing: 'allow' }],
    },
    {
      code: '<App></App>',
      options: [{ closingSlash: 'allow', beforeSelfClosing: 'allow', afterOpening: 'never', beforeClosing: 'allow' }],
    },
    {
      // `</ App>` (space after the slash) is valid TSX but lacks the `</Tag`/`/>`
      // marker the JSX auto-detect keys on, so pin `.tsx` (upstream: `no-ts`,
      // but verified to parse + report identically under ts-go).
      code: '< App></ App>',
      filename: 'case.tsx',
      options: [{ closingSlash: 'allow', beforeSelfClosing: 'allow', afterOpening: 'always', beforeClosing: 'allow' }],
    },
    {
      code: '< App/>',
      options: [{ closingSlash: 'allow', beforeSelfClosing: 'allow', afterOpening: 'always', beforeClosing: 'allow' }],
    },
    {
      code: `
        <
        App/>
      `,
      options: [{ closingSlash: 'allow', beforeSelfClosing: 'allow', afterOpening: 'allow-multiline', beforeClosing: 'allow' }],
    },
    {
      code: '<App />',
      options: [{ closingSlash: 'allow', beforeSelfClosing: 'allow', afterOpening: 'allow', beforeClosing: 'never' }],
    },
    {
      code: '<App></App>',
      options: [{ closingSlash: 'allow', beforeSelfClosing: 'allow', afterOpening: 'allow', beforeClosing: 'never' }],
    },
    {
      code: `
        <App
        foo="bar"
        >
        </App>
      `,
      options: [{ closingSlash: 'allow', beforeSelfClosing: 'allow', afterOpening: 'allow', beforeClosing: 'never' }],
    },
    {
      code: `
        <App
           foo="bar"
        >
        </App>
      `,
      options: [{ closingSlash: 'allow', beforeSelfClosing: 'allow', afterOpening: 'allow', beforeClosing: 'never' }],
    },
    {
      code: '<App ></App >',
      options: [{ closingSlash: 'allow', beforeSelfClosing: 'allow', afterOpening: 'allow', beforeClosing: 'always' }],
    },
    {
      code: `
        <App
        foo="bar"
        >
        </App >
      `,
      options: [{ closingSlash: 'allow', beforeSelfClosing: 'allow', afterOpening: 'allow', beforeClosing: 'always' }],
    },
    {
      code: `
        <App
            foo="bar"
        >
        </App >
      `,
      options: [{ closingSlash: 'allow', beforeSelfClosing: 'allow', afterOpening: 'allow', beforeClosing: 'always' }],
    },
    {
      code: '<App/>',
      options: [
        {
          closingSlash: 'never',
          beforeSelfClosing: 'never',
          afterOpening: 'never',
          beforeClosing: 'never',
        },
      ],
    },
    {
      // `/ >` lacks the `/>` marker → pin `.tsx`.
      code: '< App / >',
      filename: 'case.tsx',
      options: [
        {
          closingSlash: 'always',
          beforeSelfClosing: 'always',
          afterOpening: 'always',
          beforeClosing: 'always',
        },
      ],
    },
  ],

  invalid: [
    {
      code: '<App/>',
      output: '<App />',
      options: [{ closingSlash: 'allow', beforeSelfClosing: 'always', afterOpening: 'allow', beforeClosing: 'allow' }],
      errors: [{ messageId: 'beforeSelfCloseNeedSpace' }],
    },
    {
      code: '<App foo/>',
      output: '<App foo />',
      options: [{ closingSlash: 'allow', beforeSelfClosing: 'always', afterOpening: 'allow', beforeClosing: 'allow' }],
      errors: [{ messageId: 'beforeSelfCloseNeedSpace' }],
    },
    {
      code: '<App foo={bar}/>',
      output: '<App foo={bar} />',
      options: [{ closingSlash: 'allow', beforeSelfClosing: 'always', afterOpening: 'allow', beforeClosing: 'allow' }],
      errors: [{ messageId: 'beforeSelfCloseNeedSpace' }],
    },
    {
      code: '<App {...props}/>',
      output: '<App {...props} />',
      options: [{ closingSlash: 'allow', beforeSelfClosing: 'always', afterOpening: 'allow', beforeClosing: 'allow' }],
      errors: [{ messageId: 'beforeSelfCloseNeedSpace' }],
    },
    {
      code: '<App />',
      output: '<App/>',
      options: [{ closingSlash: 'allow', beforeSelfClosing: 'never', afterOpening: 'allow', beforeClosing: 'allow' }],
      errors: [{ messageId: 'beforeSelfCloseNoSpace' }],
    },
    {
      code: '<App foo />',
      output: '<App foo/>',
      options: [{ closingSlash: 'allow', beforeSelfClosing: 'never', afterOpening: 'allow', beforeClosing: 'allow' }],
      errors: [{ messageId: 'beforeSelfCloseNoSpace' }],
    },
    {
      code: '<App foo={bar} />',
      output: '<App foo={bar}/>',
      options: [{ closingSlash: 'allow', beforeSelfClosing: 'never', afterOpening: 'allow', beforeClosing: 'allow' }],
      errors: [{ messageId: 'beforeSelfCloseNoSpace' }],
    },
    {
      code: '<App/>',
      output: '<App />',
      options: [{ closingSlash: 'allow', beforeSelfClosing: 'proportional-always', afterOpening: 'allow', beforeClosing: 'allow' }],
      errors: [{ messageId: 'beforeSelfCloseNeedSpace' }],
    },
    {
      code: '<App foo/>',
      output: '<App foo />',
      options: [{ closingSlash: 'allow', beforeSelfClosing: 'proportional-always', afterOpening: 'allow', beforeClosing: 'allow' }],
      errors: [{ messageId: 'beforeSelfCloseNeedSpace' }],
    },
    {
      code: `
        <App
          foo={bar}/>`,
      output: `
        <App
          foo={bar}
/>`,
      options: [{ closingSlash: 'allow', beforeSelfClosing: 'proportional-always', afterOpening: 'allow', beforeClosing: 'allow' }],
      errors: [{ messageId: 'beforeSelfCloseNeedNewline' }],
    },
    {
      code: `
        <App
          foo={bar} />`,
      output: `
        <App
          foo={bar}${' '}
/>`,
      options: [{ closingSlash: 'allow', beforeSelfClosing: 'proportional-always', afterOpening: 'allow', beforeClosing: 'allow' }],
      errors: [{ messageId: 'beforeSelfCloseNeedNewline' }],
    },
    {
      code: `
        <App
          foo={bar}
          blat >
          hello
        </App>
      `,
      output: `
        <App
          foo={bar}
          blat${' '}
>
          hello
        </App>
      `,
      options: [{ closingSlash: 'allow', beforeSelfClosing: 'allow', afterOpening: 'allow', beforeClosing: 'proportional-always' }],
      errors: [{ messageId: 'beforeCloseNeedNewline' }],
    },
    {
      code: `
        <App
          foo={bar}>
          hello
        </App>
      `,
      output: `
        <App
          foo={bar}
>
          hello
        </App>
      `,
      options: [{ closingSlash: 'allow', beforeSelfClosing: 'allow', afterOpening: 'allow', beforeClosing: 'proportional-always' }],
      errors: [{ messageId: 'beforeCloseNeedNewline' }],
    },
    {
      code: '<App {...props} />',
      output: '<App {...props}/>',
      options: [{ closingSlash: 'allow', beforeSelfClosing: 'never', afterOpening: 'allow', beforeClosing: 'allow' }],
      errors: [{ messageId: 'beforeSelfCloseNoSpace' }],
    },
    {
      // `/ >` (space before `>`) lacks the `/>` marker → pin `.tsx`.
      code: '<App/ >;',
      filename: 'case.tsx',
      output: '<App/>;',
      errors: [{ messageId: 'selfCloseSlashNoSpace' }],
      options: [{ closingSlash: 'never', beforeSelfClosing: 'allow', afterOpening: 'allow', beforeClosing: 'allow' }],
    },
    {
      // `/` then newline then `>` (slash not glued to `>`) lacks the `/>` marker
      // → pin `.tsx`.
      code: `
        <App_selfCloseSlashNoSpace/
        >
      `,
      filename: 'case.tsx',
      output: `
        <App_selfCloseSlashNoSpace/>
      `,
      errors: [{ messageId: 'selfCloseSlashNoSpace' }],
      options: [{ closingSlash: 'never', beforeSelfClosing: 'allow', afterOpening: 'allow', beforeClosing: 'allow' }],
    },
    {
      code: '<App prop="foo"></App>',
      output: '<App prop="foo">< /App>',
      errors: [{ messageId: 'closeSlashNeedSpace' }],
      options: [{ closingSlash: 'always', beforeSelfClosing: 'allow', afterOpening: 'allow', beforeClosing: 'allow' }],
    },
    {
      code: '<p/>',
      output: '<p/ >',
      errors: [{ messageId: 'selfCloseSlashNeedSpace' }],
      options: [{ closingSlash: 'always', beforeSelfClosing: 'allow', afterOpening: 'allow', beforeClosing: 'allow' }],
    },
    {
      code: '< App/>',
      output: '<App/>',
      errors: [{ messageId: 'afterOpenNoSpace' }],
      options: [{ closingSlash: 'allow', beforeSelfClosing: 'allow', afterOpening: 'never', beforeClosing: 'allow' }],
    },
    {
      code: '< App></App>',
      output: '<App></App>',
      errors: [{ messageId: 'afterOpenNoSpace' }],
      options: [{ closingSlash: 'allow', beforeSelfClosing: 'allow', afterOpening: 'never', beforeClosing: 'allow' }],
    },
    {
      // `</ App>` (space after the slash) lacks the `</Tag` marker → pin `.tsx`.
      code: '<App></ App>',
      filename: 'case.tsx',
      output: '<App></App>',
      errors: [{ messageId: 'afterOpenNoSpace' }],
      options: [{ closingSlash: 'allow', beforeSelfClosing: 'allow', afterOpening: 'never', beforeClosing: 'allow' }],
    },
    {
      // `</ App>` (space after the slash) lacks the `</Tag` marker → pin `.tsx`.
      code: '< App></ App>',
      filename: 'case.tsx',
      output: '<App></App>',
      errors: [
        { messageId: 'afterOpenNoSpace' },
        { messageId: 'afterOpenNoSpace' },
      ],
      options: [{ closingSlash: 'allow', beforeSelfClosing: 'allow', afterOpening: 'never', beforeClosing: 'allow' }],
    },
    {
      code: `
        <
        App1/>
      `,
      output: `
        <App1/>
      `,
      errors: [{ messageId: 'afterOpenNoSpace' }],
      options: [{ closingSlash: 'allow', beforeSelfClosing: 'allow', afterOpening: 'never', beforeClosing: 'allow' }],
    },
    {
      // `</ App2>` (space after the slash) lacks the `</Tag` marker → pin `.tsx`.
      code: '<App2></ App2>',
      filename: 'case.tsx',
      output: '< App2></ App2>',
      errors: [{ messageId: 'afterOpenNeedSpace' }],
      options: [{ closingSlash: 'allow', beforeSelfClosing: 'allow', afterOpening: 'always', beforeClosing: 'allow' }],
    },
    {
      code: '< App3></App3>',
      output: '< App3></ App3>',
      errors: [{ messageId: 'afterOpenNeedSpace' }],
      options: [{ closingSlash: 'allow', beforeSelfClosing: 'allow', afterOpening: 'always', beforeClosing: 'allow' }],
    },
    {
      code: '<App4></App4>',
      output: '< App4></ App4>',
      errors: [
        { messageId: 'afterOpenNeedSpace' },
        { messageId: 'afterOpenNeedSpace' },
      ],
      options: [{ closingSlash: 'allow', beforeSelfClosing: 'allow', afterOpening: 'always', beforeClosing: 'allow' }],
    },
    {
      code: '<App5/>',
      output: '< App5/>',
      errors: [{ messageId: 'afterOpenNeedSpace' }],
      options: [{ closingSlash: 'allow', beforeSelfClosing: 'allow', afterOpening: 'always', beforeClosing: 'allow' }],
    },
    {
      code: '< App6/>',
      output: '<App6/>',
      errors: [{ messageId: 'afterOpenNoSpace' }],
      options: [{ closingSlash: 'allow', beforeSelfClosing: 'allow', afterOpening: 'allow-multiline', beforeClosing: 'allow' }],
    },
    {
      code: '<App7 ></App7>',
      output: '<App7></App7>',
      errors: [{ messageId: 'beforeCloseNoSpace' }],
      options: [{ closingSlash: 'allow', beforeSelfClosing: 'allow', afterOpening: 'allow', beforeClosing: 'never' }],
    },
    {
      code: '<App8></App8 >',
      output: '<App8></App8>',
      errors: [{ messageId: 'beforeCloseNoSpace' }],
      options: [{ closingSlash: 'allow', beforeSelfClosing: 'allow', afterOpening: 'allow', beforeClosing: 'never' }],
    },
    {
      code: `
        <App9
        foo="bar"
        >
        </App9 >
      `,
      output: `
        <App9
        foo="bar"
        >
        </App9>
      `,
      errors: [{ messageId: 'beforeCloseNoSpace' }],
      options: [{ closingSlash: 'allow', beforeSelfClosing: 'allow', afterOpening: 'allow', beforeClosing: 'never' }],
    },
    {
      code: '<App10></App10 >',
      output: '<App10 ></App10 >',
      errors: [{ messageId: 'beforeCloseNeedSpace' }],
      options: [{ closingSlash: 'allow', beforeSelfClosing: 'allow', afterOpening: 'allow', beforeClosing: 'always' }],
    },
    {
      code: '<App11 ></App11>',
      output: '<App11 ></App11 >',
      errors: [{ messageId: 'beforeCloseNeedSpace' }],
      options: [{ closingSlash: 'allow', beforeSelfClosing: 'allow', afterOpening: 'allow', beforeClosing: 'always' }],
    },
    {
      code: `
        <App12
        foo="bar"
        >
        </App12>
      `,
      output: `
        <App12
        foo="bar"
        >
        </App12 >
      `,
      errors: [{ messageId: 'beforeCloseNeedSpace' }],
      options: [{ closingSlash: 'allow', beforeSelfClosing: 'allow', afterOpening: 'allow', beforeClosing: 'always' }],
    },
  ],
});

/*
 * ─────────────────────────────────────────────────────────────────────────────
 * KNOWN GAPS — ts-go syntax errors (NOT behavioral rslint<->upstream gaps)
 * ─────────────────────────────────────────────────────────────────────────────
 *
 * These upstream cases were tagged `features: ['no-ts']`, meaning upstream itself
 * never runs them through the @typescript-eslint parser. Each carries the `< /`
 * form (a space BETWEEN `<` and `/` of a CLOSING tag) in its INPUT. ts-go (which
 * rslint uses) rejects that as a hard syntax error — `< /div>` lexes the `<` as
 * the start of a new JSX element whose tag name is then missing:
 *   TS17008 "JSX element has no corresponding closing tag"
 *   TS1003  "Identifier expected"
 *   TS1382  "Unexpected token. Did you mean `{'>'}` or `&gt;`?"
 * (verified by writing each fixture to a `.tsx` file and running rslint).
 *
 * Because a syntax error on ANY fixture aborts the whole CLI batch (no JSONL),
 * these MUST stay isolated here — they cannot be assertion-tested, and including
 * them would fail-loud the entire rule batch. They are upstream-valid only under
 * Babel/ESLint-default parsers, which tolerate the malformed closing tag; this
 * is a pure parser-tolerance gap, NOT a difference in jsx-tag-spacing behavior.
 *
 * NOTE: the sibling `</ Tag>` form (space AFTER the slash) parses fine in ts-go
 * and IS covered in the green sets above, as is the closeSlashNeedSpace invalid
 * whose INPUT `<App prop="foo"></App>` parses (only its fixed output `< /App>`
 * is unparseable, which the autofix `output` assertion above still verifies).
 *
 * 1) valid (closingSlash: 'always'):
 *      code: '<App prop="foo">< /App>'
 *
 * 2) invalid (closingSlash: 'never', closeSlashNoSpace):
 *      code:   '<div className="bar">< /div>;'
 *      output: '<div className="bar"></div>;'
 *      errors: [{ messageId: 'closeSlashNoSpace' }]
 *
 * 3) invalid (closingSlash: 'never', closeSlashNoSpace — `<` then newline then `/div`):
 *      code:   `
 *        <div className="bar"><
 *        /div>;
 *      `
 *      output: `
 *        <div className="bar"></div>;
 *      `
 *      errors: [{ messageId: 'closeSlashNoSpace' }]
 * ─────────────────────────────────────────────────────────────────────────────
 */
