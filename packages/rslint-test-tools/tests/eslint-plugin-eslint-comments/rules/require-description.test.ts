/**
 * @author Yosuke Ota <https://github.com/ota-meshi>
 *
 * Ported verbatim from @eslint-community/eslint-plugin-eslint-comments v4.7.2:
 *   tests/lib/rules/require-description.js
 *
 * Transformations applied per the porting spec:
 *  - `tester.run("require-description", rule, { valid, invalid })`
 *    -> `ruleTester.run('require-description', null as never, { valid, invalid })`
 *  - `require`/`new RuleTester()` setup dropped. The whole upstream file is
 *    guarded by `semver.satisfies(Linter.version, ">=7.0.0")` (the rule needs
 *    ESLint v7+); that is always true here (eslint@10.5.0), so the guard is a
 *    no-op and every (non-version-specific) case is ported.
 *  - `options` (`{ ignore: [...] }`) preserved verbatim.
 *  - errors are bare strings, copied verbatim. Upstream pins no line/column.
 *
 * Excluded (recorded under KNOWN GAPS, not silently dropped):
 *  - `eslint-env` cases (gated `<=9.0.0` upstream — directive removed in v10).
 *  - the `language: "css/css"` case (needs an ESLint v9.6+ language plugin).
 *
 * The `missingDescription` report uses `utils.toForceLocation` (directive
 * START, `column: -1`); upstream pins no column, so nothing diverges — count +
 * message match exactly.
 */

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('require-description', null as never, {
  valid: [
    '/* eslint eqeqeq: "off", curly: "error" -- Here\'s a description about why this configuration is necessary. */',
    '/* eslint-disable -- description */',
    '/* eslint-enable -- description */',
    '/* exported -- description */',
    '/* global -- description */',
    '/* globals -- description */',
    '/* just eslint in a normal comment */',
    '// eslint-disable-line -- description',
    '// eslint-disable-next-line -- description',
    '/* eslint-disable-line -- description */',
    '/* eslint-disable-next-line -- description */',
    '// eslint-disable-line eqeqeq -- description',
    '// eslint-disable-next-line eqeqeq -- description',
    {
      code: '/* eslint */',
      options: [{ ignore: ['eslint'] }],
    },
    {
      code: '/* eslint-enable */',
      options: [{ ignore: ['eslint-enable'] }],
    },
    {
      code: '/* eslint-disable */',
      options: [{ ignore: ['eslint-disable'] }],
    },
    {
      code: '// eslint-disable-line',
      options: [{ ignore: ['eslint-disable-line'] }],
    },
    {
      code: '// eslint-disable-next-line',
      options: [{ ignore: ['eslint-disable-next-line'] }],
    },
    {
      code: '/* eslint-disable-line */',
      options: [{ ignore: ['eslint-disable-line'] }],
    },
    {
      code: '/* eslint-disable-next-line */',
      options: [{ ignore: ['eslint-disable-next-line'] }],
    },
    {
      code: '/* exported */',
      options: [{ ignore: ['exported'] }],
    },
    {
      code: '/* global */',
      options: [{ ignore: ['global'] }],
    },
    {
      code: '/* globals */',
      options: [{ ignore: ['globals'] }],
    },
  ],
  invalid: [
    {
      code: '/* eslint */',
      errors: [
        'Unexpected undescribed directive comment. Include descriptions to explain why the comment is necessary.',
      ],
    },
    {
      code: '/* eslint eqeqeq: "off", curly: "error" */',
      errors: [
        'Unexpected undescribed directive comment. Include descriptions to explain why the comment is necessary.',
      ],
    },
    {
      code: '/* eslint-enable */',
      errors: [
        'Unexpected undescribed directive comment. Include descriptions to explain why the comment is necessary.',
      ],
    },
    {
      code: '/* eslint-enable eqeqeq */',
      errors: [
        'Unexpected undescribed directive comment. Include descriptions to explain why the comment is necessary.',
      ],
    },
    {
      code: '/* eslint-disable */',
      errors: [
        'Unexpected undescribed directive comment. Include descriptions to explain why the comment is necessary.',
      ],
    },
    {
      code: '/* eslint-disable eqeqeq */',
      errors: [
        'Unexpected undescribed directive comment. Include descriptions to explain why the comment is necessary.',
      ],
    },
    {
      code: '// eslint-disable-line',
      errors: [
        'Unexpected undescribed directive comment. Include descriptions to explain why the comment is necessary.',
      ],
    },
    {
      code: '// eslint-disable-line eqeqeq',
      errors: [
        'Unexpected undescribed directive comment. Include descriptions to explain why the comment is necessary.',
      ],
    },
    {
      code: '// eslint-disable-next-line',
      errors: [
        'Unexpected undescribed directive comment. Include descriptions to explain why the comment is necessary.',
      ],
    },
    {
      code: '// eslint-disable-next-line eqeqeq',
      errors: [
        'Unexpected undescribed directive comment. Include descriptions to explain why the comment is necessary.',
      ],
    },
    {
      code: '/* eslint-disable-line */',
      errors: [
        'Unexpected undescribed directive comment. Include descriptions to explain why the comment is necessary.',
      ],
    },
    {
      code: '/* eslint-disable-line eqeqeq */',
      errors: [
        'Unexpected undescribed directive comment. Include descriptions to explain why the comment is necessary.',
      ],
    },
    {
      code: '/* eslint-disable-next-line */',
      errors: [
        'Unexpected undescribed directive comment. Include descriptions to explain why the comment is necessary.',
      ],
    },
    {
      code: '/* eslint-disable-next-line eqeqeq */',
      errors: [
        'Unexpected undescribed directive comment. Include descriptions to explain why the comment is necessary.',
      ],
    },
    {
      code: '/* exported */',
      errors: [
        'Unexpected undescribed directive comment. Include descriptions to explain why the comment is necessary.',
      ],
    },
    {
      code: '/* global */',
      errors: [
        'Unexpected undescribed directive comment. Include descriptions to explain why the comment is necessary.',
      ],
    },
    {
      code: '/* global _ */',
      errors: [
        'Unexpected undescribed directive comment. Include descriptions to explain why the comment is necessary.',
      ],
    },
    {
      code: '/* globals */',
      errors: [
        'Unexpected undescribed directive comment. Include descriptions to explain why the comment is necessary.',
      ],
    },
    {
      code: '/* globals _ */',
      errors: [
        'Unexpected undescribed directive comment. Include descriptions to explain why the comment is necessary.',
      ],
    },
    // empty description
    {
      code: '/* eslint-disable-next-line eqeqeq -- */',
      errors: [
        'Unexpected undescribed directive comment. Include descriptions to explain why the comment is necessary.',
      ],
    },
  ],
});

// ---------------------------------------------------------------------------
// KNOWN GAPS (upstream cases NOT ported above — recorded, never silently
// dropped):
//
//  - `eslint-env` cases: one valid (`/* eslint-env -- description *​/`) and two
//    invalid (`/* eslint-env *​/`, `/* eslint-env node *​/`), all gated
//    `<=9.0.0` upstream. The `eslint-env` directive was removed in ESLint v10
//    and this plugin runs against eslint@10.5.0, so the directive no longer
//    exists. Inapplicable to the runtime.
//
//  - `/* eslint-disable *​/ a {}` with `plugins: { css }, language: "css/css"`
//    (one valid + one invalid): requires an ESLint v9.6+ LANGUAGE plugin
//    (`@eslint/css`). rslint's eslint-plugin runner exposes no `language`
//    configuration, so this cross-language scenario cannot be reproduced.
//    Upstream behaviour: identical "Unexpected undescribed directive comment…"
//    report on the CSS source's directive.
// ---------------------------------------------------------------------------
