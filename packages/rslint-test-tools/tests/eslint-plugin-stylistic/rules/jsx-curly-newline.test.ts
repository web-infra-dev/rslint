/**
 * @fileoverview Tests for jsx-curly-newline rule.
 *
 * Ported verbatim from @stylistic/eslint-plugin v5.10.0:
 *   packages/eslint-plugin/rules/jsx-curly-newline/jsx-curly-newline.test.ts
 *
 * Transformations applied per the porting spec:
 *  - `run({ name, rule, parserOptions, valid: valids(...), invalid: invalids(...) })`
 *    -> `ruleTester.run('jsx-curly-newline', null as never, { valid: [...], invalid: [...] })`.
 *  - The upstream `valids()` / `invalids()` helpers (from `#test/parsers-jsx`) fan
 *    each source case out across the default, `@babel/eslint-parser`, and
 *    `@typescript-eslint/parser` parsers, appending a `// parser: ...` comment to the
 *    code. That is test-harness plumbing — the per-case objects below are the verbatim
 *    source cases. rslint runs on ts-go (the TS-parser semantics upstream verifies),
 *    and the appended comment is inert for a JSX rule. (Upstream's own CI also skips the
 *    babel variant: `skipBabel = ESLint.version >= 10`, and the installed ESLint is 10.5.0.)
 *  - The local option consts (`CONSISTENT`/`NEVER`/`MULTILINE_REQUIRE`) and error consts
 *    (`LEFT_MISSING_ERROR` = `{ messageId: 'expectedAfter' }`, `LEFT_UNEXPECTED_ERROR` =
 *    `{ messageId: 'unexpectedAfter' }`, `RIGHT_MISSING_ERROR` = `{ messageId: 'expectedBefore' }`,
 *    `RIGHT_UNEXPECTED_ERROR` = `{ messageId: 'unexpectedBefore' }`) are inlined to their
 *    final values.
 *  - `parserOptions` (ecmaFeatures.jsx) dropped — rslint resolves via tsconfig and the
 *    RuleTester routes JSX fixtures to `.tsx`.
 *  - `name` / `rule` / imports dropped.
 *
 * The upstream cases use no `$`/unindent tag (all plain backtick or single-quote
 * literals, preserved byte-for-byte), no `suggestions`, no `features`, and no
 * external-fixture (`readFileSync`) cases. The `._css_` / `._json_` / `._markdown_`
 * test files don't exist for this rule. The two `output: null` invalid cases pin both
 * `errors` and `output: null` (a comment blocks the fix), so they are full cases, not
 * output-only. No case surfaced an rslint<->upstream gap, so there is no KNOWN GAPS
 * section.
 */

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('jsx-curly-newline', null as never, {
  valid: [
    // consistent option (default)
    {
      code: '<div>{foo}</div>',
      options: ['consistent'],
    },

    {
      code: `
        <div>
          {
            foo
          }
        </div>
      `,
      options: ['consistent'],
    },

    {
      code: `
        <div>
          { foo &&
            foo.bar }
        </div>
      `,
      options: ['consistent'],
    },

    {
      code: `
        <div>
          {
            foo &&
            foo.bar
          }
        </div>
      `,
      options: ['consistent'],
    },

    {
      code: `
        <div foo={
          bar
        } />
      `,
      options: ['consistent'],
    },

    // {singleline: 'consistent', multiline: 'require'} option
    {
      code: '<div>{foo}</div>',
      options: [{ singleline: 'consistent', multiline: 'require' }],
    },
    {
      code: '<div foo={bar} />',
      options: [{ singleline: 'consistent', multiline: 'require' }],
    },
    {
      code: `
        <div>
          {
            foo &&
            foo.bar
          }
        </div>
      `,
      options: [{ singleline: 'consistent', multiline: 'require' }],
    },
    {
      code: `
        <div>
          {
            foo
          }
        </div>
      `,
      options: [{ singleline: 'consistent', multiline: 'require' }],
    },

    // never option

    {
      code: '<div>{foo}</div>',
      options: ['never'],
    },

    {
      code: '<div foo={bar} />',
      options: ['never'],
    },

    {
      code: `
        <div>
          { foo &&
            foo.bar }
        </div>
      `,
      options: ['never'],
    },
  ],

  invalid: [
    // consistent option (default)
    {
      code: `
        <div>
          { foo \n}
        </div>
      `,
      output: `
        <div>
          { foo}
        </div>
      `,
      options: ['consistent'],
      errors: [{ messageId: 'unexpectedBefore' }],
    },

    {
      code: `
        <div>
          { foo &&
            foo.bar \n}
        </div>
      `,
      output: `
        <div>
          { foo &&
            foo.bar}
        </div>
      `,
      options: ['consistent'],
      errors: [{ messageId: 'unexpectedBefore' }],
    },
    {
      code: `
        <div>
          { foo &&
            bar
          }
        </div>
      `,
      output: `
        <div>
          { foo &&
            bar}
        </div>
      `,
      options: ['consistent'],
      errors: [{ messageId: 'unexpectedBefore' }],
    },

    // {singleline: 'consistent', multiline: 'require'} option
    {
      code: '<div>{foo\n}</div>',
      output: '<div>{foo}</div>',
      errors: [{ messageId: 'unexpectedBefore' }],
      options: [{ singleline: 'consistent', multiline: 'require' }],
    },
    {
      code: '<div>{\nfoo}</div>',
      output: '<div>{\nfoo\n}</div>',
      errors: [{ messageId: 'expectedBefore' }],
      options: [{ singleline: 'consistent', multiline: 'require' }],
    },
    {
      code: `
        <div>
          { foo &&
            bar }
        </div>
      `,
      output: `
        <div>
          {\n foo &&
            bar \n}
        </div>
      `,
      errors: [{ messageId: 'expectedAfter' }, { messageId: 'expectedBefore' }],
      options: [{ singleline: 'consistent', multiline: 'require' }],
    },
    {
      code: `
        <div style={foo &&
          foo.bar
        } />
      `,
      output: `
        <div style={\nfoo &&
          foo.bar
        } />
      `,
      errors: [{ messageId: 'expectedAfter' }],
      options: [{ singleline: 'consistent', multiline: 'require' }],
    },

    // never options
    {
      code: `
        <div>
          {\nfoo\n}
        </div>
      `,
      output: `
        <div>
          {foo}
        </div>
      `,
      options: ['never'],
      errors: [{ messageId: 'unexpectedAfter' }, { messageId: 'unexpectedBefore' }],
    },

    {
      code: `
        <div>
          {
            foo &&
            foo.bar
          }
        </div>
      `,
      output: `
        <div>
          {foo &&
            foo.bar}
        </div>
      `,
      options: ['never'],
      errors: [{ messageId: 'unexpectedAfter' }, { messageId: 'unexpectedBefore' }],
    },

    {
      code: `
        <div>
          { foo &&
            foo.bar
          }
        </div>
      `,
      output: `
        <div>
          { foo &&
            foo.bar}
        </div>
      `,
      options: ['never'],
      errors: [{ messageId: 'unexpectedBefore' }],
    },

    {
      code: `
        <div>
          { /* not fixed due to comment */
            foo }
        </div>
      `,
      output: null,
      options: ['never'],
      errors: [{ messageId: 'unexpectedAfter' }],
    },

    {
      code: `
        <div>
          { foo
            /* not fixed due to comment */}
        </div>
      `,
      output: null,
      options: ['never'],
      errors: [{ messageId: 'unexpectedBefore' }],
    },
  ],
});
