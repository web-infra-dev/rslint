/**
 * Conformance: eslint-plugin-es-x (regexp) mounted in rslint via `plugins` must
 * report identically to ESLint v10. es-x rules are AST pattern matches over ES
 * version features / builtin APIs (no type info), so rslint reproduces ESLint
 * byte-for-byte. Representative triggers from the upstream test suite (v9.6.0).
 */
import { runConformanceSuite } from '../conformance.js';
import type { DiffCase } from '../harness.js';

const CASES: DiffCase[] = [
  { pkg: 'eslint-plugin-es-x', rule: 'no-regexp-d-flag', code: '/foo/d' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-regexp-d-flag', code: '/foo/gimsuyd' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-regexp-d-flag',
    code: "new RegExp('foo', 'd')",
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-regexp-d-flag',
    code: "new RegExp('foo', 'gimsuyd')",
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-regexp-escape',
    code: 'RegExp.escape',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-regexp-prototype-flags',
    code: '\n            const re = /a/\n            if (typeof re.flags === "string") {\n                console.log(re.flags)\n            }',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-regexp-prototype-flags',
    code: '\n            const re = /a/\n            if (re.flags) {\n                console.log(re.flags)\n            }',
  },
  { pkg: 'eslint-plugin-es-x', rule: 'no-regexp-s-flag', code: '/foo/s' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-regexp-s-flag', code: '/foo/gimsuy' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-regexp-s-flag',
    code: "new RegExp('foo', 's')",
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-regexp-s-flag',
    code: "new RegExp('foo', 'gimsuy')",
  },
  { pkg: 'eslint-plugin-es-x', rule: 'no-regexp-u-flag', code: '/foo/u' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-regexp-u-flag', code: '/foo/gimsuy' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-regexp-u-flag',
    code: "new RegExp('foo', 'u')",
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-regexp-u-flag',
    code: "new RegExp('foo', 'gimsuy')",
  },
  { pkg: 'eslint-plugin-es-x', rule: 'no-regexp-v-flag', code: '/foo/v' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-regexp-v-flag', code: '/foo/gimsyv' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-regexp-v-flag',
    code: "new RegExp('foo', 'v')",
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-regexp-v-flag',
    code: "new RegExp('foo', 'gimsyv')",
  },
  { pkg: 'eslint-plugin-es-x', rule: 'no-regexp-y-flag', code: '/foo/y' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-regexp-y-flag', code: '/foo/gimsuy' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-regexp-y-flag',
    code: "new RegExp('foo', 'y')",
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-regexp-y-flag',
    code: "new RegExp('foo', 'gimsuy')",
  },
];

const CLEAN_CASES: DiffCase[] = [
  { pkg: 'eslint-plugin-es-x', rule: 'no-regexp-d-flag', code: '/foo/gimuys' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-regexp-d-flag', code: 'a\n/b/d' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-regexp-escape', code: 'RegExp.$1' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-regexp-prototype-compile',
    code: 'foo.global',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-regexp-prototype-flags',
    code: 'foo.flags',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-regexp-prototype-flags',
    code: 'foo.global',
  },
  { pkg: 'eslint-plugin-es-x', rule: 'no-regexp-s-flag', code: '/foo/gimuy' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-regexp-s-flag', code: 'a\n/b/s' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-regexp-u-flag', code: '/foo/' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-regexp-u-flag', code: '/foo/gimsy' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-regexp-v-flag', code: '/foo/gimsu' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-regexp-v-flag', code: 'a\n/b/y' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-regexp-y-flag', code: '/foo/gimsu' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-regexp-y-flag', code: 'a\n/b/y' },
];

runConformanceSuite('eslint-plugin-es-x', CASES, CLEAN_CASES);
