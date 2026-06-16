/**
 * @author Toru Nagashima <https://github.com/mysticatea>
 *
 * Ported verbatim from @eslint-community/eslint-plugin-eslint-comments v4.7.2:
 *   tests/lib/rules/no-duplicate-disable.js
 *
 * Transformations applied per the porting spec:
 *  - `tester.run('no-duplicate-disable', rule, { valid, invalid })` ->
 *    `ruleTester.run('no-duplicate-disable', null as never, { valid, invalid })`.
 *  - Dropped the CJS `require`/`RuleTester`/`rule`/`semver`/`Linter` setup.
 *  - The `semver.satisfies(Linter.version, ">=7.0.0")` description case is
 *    inlined unconditionally (installed eslint is v9).
 *  - Multi-line backtick fixtures are kept byte-for-byte.
 *
 * This rule emits no autofix, so there are no `output` cases.
 *
 * KNOWN GAPS (moved out of the run() block, upstream expectation preserved):
 *  - The `semver.satisfies(Linter.version, ">=9.6.0")` CSS language-plugin cases
 *    (one valid, one invalid) need `plugins`/`language` the alignment RuleTester
 *    cannot mount; the bare `a {}` CSS body is a ts-go syntax error. Isolated
 *    below, not deleted.
 */

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-duplicate-disable', null as never, {
  valid: [
    `
//eslint-disable-line
`,
    `
/*eslint-disable-line*/
`,
    `
/*eslint-disable no-undef*/
//eslint-disable-line no-unused-vars
//eslint-disable-next-line semi
/*eslint-disable eqeqeq*/
`,
    `
/*eslint-disable no-undef*/
/*eslint-disable-line no-unused-vars*/
/*eslint-disable-next-line semi*/
/*eslint-disable eqeqeq*/
`,
  ],
  invalid: [
    {
      code: `
/*eslint-disable no-undef*/
//eslint-disable-line no-undef
`,
      errors: [
        {
          message: "'no-undef' rule has been disabled already.",
          line: 3,
          column: 23,
          endLine: 3,
          endColumn: 31,
        },
      ],
    },
    {
      code: `
/*eslint-disable no-undef*/
/*eslint-disable-line no-undef*/
`,
      errors: [
        {
          message: "'no-undef' rule has been disabled already.",
          line: 3,
          column: 23,
          endLine: 3,
          endColumn: 31,
        },
      ],
    },
    {
      code: `
/*eslint-disable no-undef*/
//eslint-disable-next-line no-undef
`,
      errors: [
        {
          message: "'no-undef' rule has been disabled already.",
          line: 3,
          column: 28,
          endLine: 3,
          endColumn: 36,
        },
      ],
    },
    {
      code: `
/*eslint-disable no-undef*/
/*eslint-disable-next-line no-undef*/
`,
      errors: [
        {
          message: "'no-undef' rule has been disabled already.",
          line: 3,
          column: 28,
          endLine: 3,
          endColumn: 36,
        },
      ],
    },
    {
      code: `
//eslint-disable-next-line no-undef
//eslint-disable-line no-undef
`,
      errors: [
        {
          message: "'no-undef' rule has been disabled already.",
          line: 3,
          column: 23,
          endLine: 3,
          endColumn: 31,
        },
      ],
    },
    {
      code: `
/*eslint-disable-next-line no-undef*/
/*eslint-disable-line no-undef*/
`,
      errors: [
        {
          message: "'no-undef' rule has been disabled already.",
          line: 3,
          column: 23,
          endLine: 3,
          endColumn: 31,
        },
      ],
    },
    // -- description
    {
      code: `
// eslint-disable-next-line no-undef -- description
// eslint-disable-line no-undef -- description
`,
      errors: [
        {
          message: "'no-undef' rule has been disabled already.",
          line: 3,
          column: 24,
          endLine: 3,
          endColumn: 32,
        },
      ],
    },
  ],
});

/*
KNOWN GAPS — CSS language-plugin cases (upstream gated on Linter.version >=9.6.0).
These require `plugins: { css: require('@eslint/css').default }` + `language: 'css/css'`,
which the alignment RuleTester does not support (the bare `a {}` CSS body is a
ts-go syntax error). Upstream expectations preserved verbatim:

  // valid
  {
    code: `
/-*eslint-disable no-undef*-/
/-*eslint-disable-line no-unused-vars*-/
/-*eslint-disable-next-line semi*-/
/-*eslint-disable eqeqeq*-/
a {}`,
    plugins: { css: require('@eslint/css').default },
    language: 'css/css',
  }

  // invalid
  {
    code: `
/-* eslint-disable-next-line no-undef *-/
/-* eslint-disable-line no-undef *-/
a {}`,
    plugins: { css: require('@eslint/css').default },
    language: 'css/css',
    errors: [
      {
        message: "'no-undef' rule has been disabled already.",
        line: 3,
        column: 24,
        endLine: 3,
        endColumn: 32,
      },
    ],
  }
*/
