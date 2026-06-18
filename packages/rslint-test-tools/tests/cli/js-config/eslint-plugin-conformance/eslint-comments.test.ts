/**
 * Conformance: @eslint-community/eslint-plugin-eslint-comments mounted in rslint
 * via `plugins` must report identically to ESLint v10. Only the subset rslint
 * reproduces byte-for-byte is included; each case was verified individually.
 *
 * Intentionally EXCLUDED — runner limitations surfaced by this comparison, not
 * test gaps (kept honest, never faked green):
 *   - Diagnostics pointing at a whole-comment disable/enable directive with NO
 *     rule list (no-unlimited-disable, no-use, require-description,
 *     no-aggregating-enable on a bare enable, no-restricted-disable on a bare
 *     disable): rslint reports the start as (1,1) rather than the comment's real
 *     start line/column, so they diverge from ESLint. Cases that name specific
 *     rules in the directive locate correctly and ARE included.
 *   - no-unused-disable: the rule patches Linter#verify /
 *     reportUnusedDisableDirectives, which the rslint plugin runner does not
 *     expose, so it never fires.
 */
import { runConformanceSuite } from './conformance.js';
import type { DiffCase } from './harness.js';

const CASES: DiffCase[] = [
  {
    pkg: '@eslint-community/eslint-plugin-eslint-comments',
    rule: 'disable-enable-pair',
    code: '\n/*eslint-disable no-undef*/\n',
  },
  {
    pkg: '@eslint-community/eslint-plugin-eslint-comments',
    rule: 'disable-enable-pair',
    code: '\n/*eslint-disable no-undef,no-unused-vars*/\n/*eslint-enable no-undef*/\n',
  },
  {
    pkg: '@eslint-community/eslint-plugin-eslint-comments',
    rule: 'disable-enable-pair',
    code: '\n/*eslint-disable no-undef*/\nconsole.log();\n/*eslint-disable no-unused-vars*/\n',
    options: [{ allowWholeFile: true }],
  },
  {
    pkg: '@eslint-community/eslint-plugin-eslint-comments',
    rule: 'disable-enable-pair',
    code: '\n{\n/*eslint-disable no-unused-vars -- description */\n}\n',
    options: [{ allowWholeFile: true }],
  },
  {
    pkg: '@eslint-community/eslint-plugin-eslint-comments',
    rule: 'no-duplicate-disable',
    code: '\n/*eslint-disable no-undef*/\n//eslint-disable-line no-undef\n',
  },
  {
    pkg: '@eslint-community/eslint-plugin-eslint-comments',
    rule: 'no-duplicate-disable',
    code: '\n/*eslint-disable no-undef*/\n//eslint-disable-next-line no-undef\n',
  },
  {
    pkg: '@eslint-community/eslint-plugin-eslint-comments',
    rule: 'no-duplicate-disable',
    code: '\n//eslint-disable-next-line no-undef\n//eslint-disable-line no-undef\n',
  },
  {
    pkg: '@eslint-community/eslint-plugin-eslint-comments',
    rule: 'no-duplicate-disable',
    code: '\n// eslint-disable-next-line no-undef -- description\n// eslint-disable-line no-undef -- description\n',
  },
  {
    pkg: '@eslint-community/eslint-plugin-eslint-comments',
    rule: 'no-restricted-disable',
    code: '/*eslint-disable eqeqeq*/',
    options: ['eqeqeq'],
  },
  {
    pkg: '@eslint-community/eslint-plugin-eslint-comments',
    rule: 'no-restricted-disable',
    code: '/*eslint-disable eqeqeq, no-undef, no-redeclare*/',
    options: ['*', '!no-undef', '!no-redeclare'],
  },
  {
    pkg: '@eslint-community/eslint-plugin-eslint-comments',
    rule: 'no-restricted-disable',
    code: '/*eslint-disable semi, no-extra-semi, semi-style, comma-style*/',
    options: ['*semi*'],
  },
  {
    pkg: '@eslint-community/eslint-plugin-eslint-comments',
    rule: 'no-unused-enable',
    code: '/*eslint-enable no-undef*/',
  },
  {
    pkg: '@eslint-community/eslint-plugin-eslint-comments',
    rule: 'no-unused-enable',
    code: '\n/*eslint-disable no-unused-vars*/\n/*eslint-enable no-undef*/\n',
  },
];

const CLEAN_CASES: DiffCase[] = [
  {
    pkg: '@eslint-community/eslint-plugin-eslint-comments',
    rule: 'disable-enable-pair',
    code: '\n/*eslint-disable no-undef,no-unused-vars*/\n/*eslint-enable no-undef,no-unused-vars*/\n',
  },
  {
    pkg: '@eslint-community/eslint-plugin-eslint-comments',
    rule: 'disable-enable-pair',
    code: '\n/*eslint-disable no-undef -- description*/\n/*eslint-enable no-undef*/\n',
  },
  {
    pkg: '@eslint-community/eslint-plugin-eslint-comments',
    rule: 'no-aggregating-enable',
    code: '\n            /*eslint-disable no-redeclare*/\n            /*eslint-enable no-redeclare*/\n        ',
  },
  {
    pkg: '@eslint-community/eslint-plugin-eslint-comments',
    rule: 'no-aggregating-enable',
    code: '\n            /*eslint-disable no-redeclare, no-shadow*/\n            /*eslint-enable no-redeclare*/\n            /*eslint-enable no-shadow*/\n        ',
  },
  {
    pkg: '@eslint-community/eslint-plugin-eslint-comments',
    rule: 'no-duplicate-disable',
    code: '\n/*eslint-disable no-undef*/\n//eslint-disable-line no-unused-vars\n//eslint-disable-next-line semi\n/*eslint-disable eqeqeq*/\n',
  },
  {
    pkg: '@eslint-community/eslint-plugin-eslint-comments',
    rule: 'no-duplicate-disable',
    code: '\n//eslint-disable-line\n',
  },
  {
    pkg: '@eslint-community/eslint-plugin-eslint-comments',
    rule: 'no-restricted-disable',
    code: '/*eslint-disable eqeqeq*/',
    options: ['no-unused-vars'],
  },
  {
    pkg: '@eslint-community/eslint-plugin-eslint-comments',
    rule: 'no-restricted-disable',
    code: '/*eslint-disable eqeqeq*/',
    options: ['*', '!eqeqeq'],
  },
  {
    pkg: '@eslint-community/eslint-plugin-eslint-comments',
    rule: 'no-unlimited-disable',
    code: '/*eslint-disable eqeqeq*/',
  },
  {
    pkg: '@eslint-community/eslint-plugin-eslint-comments',
    rule: 'no-unlimited-disable',
    code: 'var foo;\n//eslint-disable-line eqeqeq',
  },
  {
    pkg: '@eslint-community/eslint-plugin-eslint-comments',
    rule: 'no-unused-enable',
    code: '\n/*eslint no-undef:error*/\n/*eslint-disable*/\nvar a = b\n/*eslint-enable*/\n',
  },
  {
    pkg: '@eslint-community/eslint-plugin-eslint-comments',
    rule: 'no-unused-enable',
    code: '\n/*eslint no-undef:error, no-unused-vars:error*/\n/*eslint-disable no-undef,no-unused-vars*/\nvar a = b\n/*eslint-enable no-undef*/\n',
  },
  {
    pkg: '@eslint-community/eslint-plugin-eslint-comments',
    rule: 'no-use',
    code: '/* just eslint in a normal comment */',
  },
  {
    pkg: '@eslint-community/eslint-plugin-eslint-comments',
    rule: 'no-use',
    code: '/* eslint */',
    options: [{ allow: ['eslint'] }],
  },
  {
    pkg: '@eslint-community/eslint-plugin-eslint-comments',
    rule: 'require-description',
    code: '/* eslint eqeqeq: "off", curly: "error" -- Here\'s a description about why this configuration is necessary. */',
  },
  {
    pkg: '@eslint-community/eslint-plugin-eslint-comments',
    rule: 'require-description',
    code: '// eslint-disable-line eqeqeq -- description',
    options: [{ ignore: ['eslint-disable-line'] }],
  },
];

runConformanceSuite(
  '@eslint-community/eslint-plugin-eslint-comments',
  CASES,
  CLEAN_CASES,
);
