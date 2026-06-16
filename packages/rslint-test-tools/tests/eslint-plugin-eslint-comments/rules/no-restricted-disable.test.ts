/**
 * @author Toru Nagashima <https://github.com/mysticatea>
 *
 * Ported verbatim from @eslint-community/eslint-plugin-eslint-comments v4.7.2:
 *   tests/lib/rules/no-restricted-disable.js
 *
 * Transformations applied per the porting spec:
 *  - `tester.run('no-restricted-disable', rule, { valid, invalid })` ->
 *    `ruleTester.run('no-restricted-disable', null as never, { valid, invalid })`.
 *  - Dropped the CJS `require`/`RuleTester`/`rule`/`semver`/`Linter` setup and
 *    the `foo/no-undef` / `foo/no-redeclare` plugin-rule registration: this rule
 *    only inspects the textual rule list inside the directive comment, so the
 *    referenced rules need not actually be mounted for it to report.
 *  - The `semver.satisfies(Linter.version, ">=7.0.0")` description case is
 *    inlined unconditionally (installed eslint is v9).
 *  - Upstream errors are bare strings (the rendered message); kept verbatim.
 *  - `options` preserved verbatim.
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

ruleTester.run('no-restricted-disable', null as never, {
  valid: [
    '/*eslint-disable*/',
    '//eslint-disable-line',
    '//eslint-disable-next-line',
    '/*eslint-disable-line*/',
    '/*eslint-disable-next-line*/',
    {
      code: '/*eslint-disable eqeqeq*/',
      options: ['no-unused-vars'],
    },
    {
      code: '/*eslint-enable eqeqeq*/',
      options: ['eqeqeq'],
    },
    {
      code: '/*eslint-disable eqeqeq*/',
      options: ['*', '!eqeqeq'],
    },
  ],
  invalid: [
    {
      code: '/*eslint-disable eqeqeq*/',
      options: ['eqeqeq'],
      errors: ["Disabling 'eqeqeq' is not allowed."],
    },
    {
      code: '/*eslint-disable*/',
      options: ['eqeqeq'],
      errors: ["Disabling 'eqeqeq' is not allowed."],
    },
    {
      code: '//eslint-disable-line eqeqeq',
      options: ['eqeqeq'],
      errors: ["Disabling 'eqeqeq' is not allowed."],
    },
    {
      code: '/*eslint-disable-line eqeqeq*/',
      options: ['eqeqeq'],
      errors: ["Disabling 'eqeqeq' is not allowed."],
    },
    {
      code: '//eslint-disable-line',
      options: ['eqeqeq'],
      errors: ["Disabling 'eqeqeq' is not allowed."],
    },
    {
      code: '/*eslint-disable-line*/',
      options: ['eqeqeq'],
      errors: ["Disabling 'eqeqeq' is not allowed."],
    },
    {
      code: '//eslint-disable-next-line eqeqeq',
      options: ['eqeqeq'],
      errors: ["Disabling 'eqeqeq' is not allowed."],
    },
    {
      code: '/*eslint-disable-next-line eqeqeq*/',
      options: ['eqeqeq'],
      errors: ["Disabling 'eqeqeq' is not allowed."],
    },
    {
      code: '//eslint-disable-next-line',
      options: ['eqeqeq'],
      errors: ["Disabling 'eqeqeq' is not allowed."],
    },
    {
      code: '/*eslint-disable-next-line*/',
      options: ['eqeqeq'],
      errors: ["Disabling 'eqeqeq' is not allowed."],
    },

    {
      code: '/*eslint-disable eqeqeq, no-undef, no-redeclare*/',
      options: ['*', '!no-undef', '!no-redeclare'],
      errors: ["Disabling 'eqeqeq' is not allowed."],
    },
    {
      code: '/*eslint-disable*/',
      options: ['*', '!no-undef', '!no-redeclare'],
      errors: ["Disabling '*,!no-undef,!no-redeclare' is not allowed."],
    },
    {
      code: '//eslint-disable-line eqeqeq, no-undef, no-redeclare',
      options: ['*', '!no-undef', '!no-redeclare'],
      errors: ["Disabling 'eqeqeq' is not allowed."],
    },
    {
      code: '/*eslint-disable-line eqeqeq, no-undef, no-redeclare*/',
      options: ['*', '!no-undef', '!no-redeclare'],
      errors: ["Disabling 'eqeqeq' is not allowed."],
    },
    {
      code: '//eslint-disable-line',
      options: ['*', '!no-undef', '!no-redeclare'],
      errors: ["Disabling '*,!no-undef,!no-redeclare' is not allowed."],
    },
    {
      code: '/*eslint-disable-line*/',
      options: ['*', '!no-undef', '!no-redeclare'],
      errors: ["Disabling '*,!no-undef,!no-redeclare' is not allowed."],
    },
    {
      code: '//eslint-disable-next-line eqeqeq, no-undef, no-redeclare',
      options: ['*', '!no-undef', '!no-redeclare'],
      errors: ["Disabling 'eqeqeq' is not allowed."],
    },
    {
      code: '/*eslint-disable-next-line eqeqeq, no-undef, no-redeclare*/',
      options: ['*', '!no-undef', '!no-redeclare'],
      errors: ["Disabling 'eqeqeq' is not allowed."],
    },
    {
      code: '//eslint-disable-next-line',
      options: ['*', '!no-undef', '!no-redeclare'],
      errors: ["Disabling '*,!no-undef,!no-redeclare' is not allowed."],
    },
    {
      code: '/*eslint-disable-next-line*/',
      options: ['*', '!no-undef', '!no-redeclare'],
      errors: ["Disabling '*,!no-undef,!no-redeclare' is not allowed."],
    },

    {
      code: '/*eslint-disable semi, no-extra-semi, semi-style, comma-style*/',
      options: ['*semi*'],
      errors: [
        "Disabling 'semi' is not allowed.",
        "Disabling 'no-extra-semi' is not allowed.",
        "Disabling 'semi-style' is not allowed.",
      ],
    },
    {
      code: '/*eslint-disable no-undef, no-redeclare, foo/no-undef, foo/no-redeclare*/',
      options: ['foo/*'],
      errors: [
        "Disabling 'foo/no-undef' is not allowed.",
        "Disabling 'foo/no-redeclare' is not allowed.",
      ],
    },
    // -- description
    {
      code: '/*eslint-disable -- description*/',
      options: ['eqeqeq'],
      errors: ["Disabling 'eqeqeq' is not allowed."],
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
    code: '/-*eslint-disable eqeqeq*-/ a {}',
    options: ['*', '!eqeqeq'],
    plugins: { css: require('@eslint/css').default },
    language: 'css/css',
  }

  // invalid
  {
    code: '/-*eslint-disable eqeqeq*-/ a {}',
    options: ['eqeqeq'],
    plugins: { css: require('@eslint/css').default },
    language: 'css/css',
    errors: ["Disabling 'eqeqeq' is not allowed."],
  }
*/
