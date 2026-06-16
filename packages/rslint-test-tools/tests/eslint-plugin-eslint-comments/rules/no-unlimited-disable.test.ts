/**
 * @author Toru Nagashima <https://github.com/mysticatea>
 *
 * Ported verbatim from @eslint-community/eslint-plugin-eslint-comments v4.7.2:
 *   tests/lib/rules/no-unlimited-disable.js
 *
 * Transformations applied per the porting spec:
 *  - `tester.run('no-unlimited-disable', rule, { valid, invalid })` ->
 *    `ruleTester.run('no-unlimited-disable', null as never, { valid, invalid })`.
 *  - Dropped the CJS `require`/`RuleTester`/`rule`/`semver`/`Linter` setup.
 *  - The `semver.satisfies(Linter.version, ">=7.0.0")` description case is
 *    inlined unconditionally (installed eslint is v9).
 *  - Upstream errors are bare strings (the rendered message); kept verbatim.
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

ruleTester.run('no-unlimited-disable', null as never, {
  valid: [
    '/*eslint-enable*/',
    '/*eslint-disable eqeqeq*/',
    '//eslint-disable-line eqeqeq',
    '//eslint-disable-next-line eqeqeq',
    '/*eslint-disable-line eqeqeq*/',
    '/*eslint-disable-next-line eqeqeq*/',
    'var foo;\n//eslint-disable-line eqeqeq',
    'var foo;\n/*eslint-disable-line eqeqeq*/',
  ],
  invalid: [
    {
      code: '/*eslint-disable */',
      errors: [
        "Unexpected unlimited 'eslint-disable' comment. Specify some rule names to disable.",
      ],
    },
    {
      code: '/* eslint-disable */',
      errors: [
        "Unexpected unlimited 'eslint-disable' comment. Specify some rule names to disable.",
      ],
    },
    {
      code: '//eslint-disable-line',
      errors: [
        "Unexpected unlimited 'eslint-disable-line' comment. Specify some rule names to disable.",
      ],
    },
    {
      code: '/*eslint-disable-line*/',
      errors: [
        "Unexpected unlimited 'eslint-disable-line' comment. Specify some rule names to disable.",
      ],
    },
    {
      code: '// eslint-disable-line ',
      errors: [
        "Unexpected unlimited 'eslint-disable-line' comment. Specify some rule names to disable.",
      ],
    },
    {
      code: '/* eslint-disable-line */',
      errors: [
        "Unexpected unlimited 'eslint-disable-line' comment. Specify some rule names to disable.",
      ],
    },
    {
      code: '//eslint-disable-next-line',
      errors: [
        "Unexpected unlimited 'eslint-disable-next-line' comment. Specify some rule names to disable.",
      ],
    },
    {
      code: '/*eslint-disable-next-line*/',
      errors: [
        "Unexpected unlimited 'eslint-disable-next-line' comment. Specify some rule names to disable.",
      ],
    },
    {
      code: '// eslint-disable-next-line ',
      errors: [
        "Unexpected unlimited 'eslint-disable-next-line' comment. Specify some rule names to disable.",
      ],
    },
    {
      code: '/* eslint-disable-next-line */',
      errors: [
        "Unexpected unlimited 'eslint-disable-next-line' comment. Specify some rule names to disable.",
      ],
    },
    {
      code: 'var foo;\n//eslint-disable-line',
      errors: [
        "Unexpected unlimited 'eslint-disable-line' comment. Specify some rule names to disable.",
      ],
    },
    {
      code: 'var foo;\n/*eslint-disable-line*/',
      errors: [
        "Unexpected unlimited 'eslint-disable-line' comment. Specify some rule names to disable.",
      ],
    },
    // -- description
    {
      code: '/*eslint-disable -- description */',
      errors: [
        "Unexpected unlimited 'eslint-disable' comment. Specify some rule names to disable.",
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
    code: '/-*eslint-disable-line eqeqeq*-/ a {}',
    plugins: { css: require('@eslint/css').default },
    language: 'css/css',
  }

  // invalid
  {
    code: '/-* eslint-disable *-/ a {}',
    plugins: { css: require('@eslint/css').default },
    language: 'css/css',
    errors: [
      "Unexpected unlimited 'eslint-disable' comment. Specify some rule names to disable.",
    ],
  }
*/
