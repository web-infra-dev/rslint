/**
 * @author Toru Nagashima <https://github.com/mysticatea>
 *
 * Ported verbatim from @eslint-community/eslint-plugin-eslint-comments v4.7.2:
 *   tests/lib/rules/no-aggregating-enable.js
 *
 * Transformations applied per the porting spec:
 *  - `tester.run('no-aggregating-enable', rule, { valid, invalid })` ->
 *    `ruleTester.run('no-aggregating-enable', null as never, { valid, invalid })`.
 *  - Dropped the CJS `require`/`RuleTester`/`rule`/`semver`/`Linter` setup.
 *  - The `semver.satisfies(Linter.version, ">=7.0.0")` description case is
 *    inlined unconditionally (installed eslint is v9).
 *  - Upstream errors are bare strings (the rendered message); kept verbatim.
 *  - Multi-line backtick fixtures (indented with leading whitespace) are kept
 *    byte-for-byte.
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

ruleTester.run('no-aggregating-enable', null as never, {
  valid: [
    `
            /*eslint-disable no-redeclare*/
            /*eslint-enable no-redeclare*/
        `,
    `
            /*eslint-disable no-redeclare*/
            /*eslint-enable no-shadow*/
        `,
    `
            /*eslint-disable no-redeclare, no-shadow*/
            /*eslint-enable*/
        `,
    `
            /*eslint-disable no-redeclare, no-shadow*/
            /*eslint-enable no-redeclare, no-shadow*/
        `,
    `
            /*eslint-disable no-redeclare, no-shadow*/
            /*eslint-enable no-redeclare*/
            /*eslint-enable no-shadow*/
        `,
  ],
  invalid: [
    {
      code: `
                /*eslint-disable no-redeclare*/
                /*eslint-disable no-shadow*/
                /*eslint-enable*/
            `,
      errors: [
        'This `eslint-enable` comment affects 2 `eslint-disable` comments. An `eslint-enable` comment should be for an `eslint-disable` comment.',
      ],
    },
    {
      code: `
                /*eslint-disable no-redeclare*/
                /*eslint-disable no-shadow*/
                /*eslint-disable no-undef*/
                /*eslint-enable*/
            `,
      errors: [
        'This `eslint-enable` comment affects 3 `eslint-disable` comments. An `eslint-enable` comment should be for an `eslint-disable` comment.',
      ],
    },
    {
      code: `
                /*eslint-disable no-redeclare*/
                /*eslint-disable no-shadow*/
                /*eslint-enable no-redeclare, no-shadow*/
            `,
      errors: [
        'This `eslint-enable` comment affects 2 `eslint-disable` comments. An `eslint-enable` comment should be for an `eslint-disable` comment.',
      ],
    },
    // -- description
    {
      code: `
                /*eslint-disable no-redeclare*/
                /*eslint-disable no-shadow*/
                /*eslint-enable -- description*/
            `,
      errors: [
        'This `eslint-enable` comment affects 2 `eslint-disable` comments. An `eslint-enable` comment should be for an `eslint-disable` comment.',
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
            /-*eslint-disable no-redeclare, no-shadow*-/
            /-*eslint-enable no-redeclare*-/
            /-*eslint-enable no-shadow*-/
            a {}`,
    plugins: { css: require('@eslint/css').default },
    language: 'css/css',
  }

  // invalid
  {
    code: `
                /-*eslint-disable no-redeclare*-/
                /-*eslint-disable no-shadow*-/
                /-*eslint-enable*-/
            a {}`,
    plugins: { css: require('@eslint/css').default },
    language: 'css/css',
    errors: [
      'This `eslint-enable` comment affects 2 `eslint-disable` comments. An `eslint-enable` comment should be for an `eslint-disable` comment.',
    ],
  }
*/
