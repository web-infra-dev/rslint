/**
 * @author Toru Nagashima <https://github.com/mysticatea>
 *
 * Ported verbatim from @eslint-community/eslint-plugin-eslint-comments v4.7.2:
 *   tests/lib/rules/no-use.js
 *
 * Transformations applied per the porting spec:
 *  - `tester.run("no-use", rule, { valid, invalid })`
 *    -> `ruleTester.run('no-use', null as never, { valid, invalid })`
 *  - `require`/`new RuleTester()` setup dropped.
 *  - Upstream guards a handful of cases behind `semver.satisfies(Linter.version, ...)`:
 *      • `// eslint-env` / `/* eslint-env *​/` — only on ESLint <=9 (rule removed
 *        in v10). The plugin is pinned to eslint@10.5.0 here, so these are NOT
 *        ported (they don't exist on the runtime ESLint).
 *      • The `language: "css/css"` (`@eslint/css`) case — needs an ESLint v9.6+
 *        language plugin; rslint's eslint-plugin runner has no `language` plumbing,
 *        so it's recorded under KNOWN GAPS, not ported.
 *  - errors are bare strings (`"Unexpected ESLint directive comment."`), copied
 *    verbatim. No line/column is pinned upstream, so none is asserted here.
 *
 * The `disallow` report uses `utils.toForceLocation` (the directive comment's
 * START, `column: -1`), so upstream pins no column — there is nothing to diverge
 * on for the ported cases, and the diagnostic count + message match exactly.
 */

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-use', null as never, {
  valid: [
    '// eslint foo',
    '// eslint-disable',
    '// eslint-enable',
    '// exported',
    '// global',
    '// globals',
    '/* just eslint in a normal comment */',
    {
      code: '/* eslint */',
      options: [{ allow: ['eslint'] }],
    },
    {
      code: '/* eslint-enable */',
      options: [{ allow: ['eslint-enable'] }],
    },
    {
      code: '/* eslint-disable */',
      options: [{ allow: ['eslint-disable'] }],
    },
    {
      code: '// eslint-disable-line',
      options: [{ allow: ['eslint-disable-line'] }],
    },
    {
      code: '// eslint-disable-next-line',
      options: [{ allow: ['eslint-disable-next-line'] }],
    },
    {
      code: '/* eslint-disable-line */',
      options: [{ allow: ['eslint-disable-line'] }],
    },
    {
      code: '/* eslint-disable-next-line */',
      options: [{ allow: ['eslint-disable-next-line'] }],
    },
    {
      code: '/* exported */',
      options: [{ allow: ['exported'] }],
    },
    {
      code: '/* global */',
      options: [{ allow: ['global'] }],
    },
    {
      code: '/* globals */',
      options: [{ allow: ['globals'] }],
    },
  ],
  invalid: [
    {
      code: '/* eslint */',
      errors: ['Unexpected ESLint directive comment.'],
    },
    {
      code: '/* eslint-enable */',
      errors: ['Unexpected ESLint directive comment.'],
    },
    {
      code: '/* eslint-disable */',
      errors: ['Unexpected ESLint directive comment.'],
    },
    {
      code: '// eslint-disable-line',
      errors: ['Unexpected ESLint directive comment.'],
    },
    {
      code: '// eslint-disable-next-line',
      errors: ['Unexpected ESLint directive comment.'],
    },
    {
      code: '/* eslint-disable-line */',
      errors: ['Unexpected ESLint directive comment.'],
    },
    {
      code: '/* eslint-disable-next-line */',
      errors: ['Unexpected ESLint directive comment.'],
    },
    {
      code: '/* exported */',
      errors: ['Unexpected ESLint directive comment.'],
    },
    {
      code: '/* global */',
      errors: ['Unexpected ESLint directive comment.'],
    },
    {
      code: '/* globals */',
      errors: ['Unexpected ESLint directive comment.'],
    },
  ],
});

// ---------------------------------------------------------------------------
// KNOWN GAPS (upstream cases NOT ported above — recorded, never silently
// dropped):
//
//  - `// eslint-env` (valid) and `/* eslint-env *​/` (invalid): upstream only
//    runs these on ESLint <=9.0.0; the `eslint-env` directive was removed in
//    ESLint v10 and this plugin runs against eslint@10.5.0, so the directive
//    no longer exists and the cases are inapplicable.
//
//  - `/* eslint-disable *​/ a {}` with `plugins: { css }, language: "css/css"`
//    (one valid + one invalid): requires an ESLint v9.6+ LANGUAGE plugin
//    (`@eslint/css`). rslint's eslint-plugin runner exposes no `language`
//    configuration, so this cross-language directive scenario cannot be
//    reproduced. Upstream behaviour: identical "Unexpected ESLint directive
//    comment." report on the CSS source's directive.
// ---------------------------------------------------------------------------
