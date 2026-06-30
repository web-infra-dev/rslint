/**
 * Conformance: eslint-plugin-es-x (promise) mounted in rslint via `plugins` must
 * report identically to ESLint v10. es-x rules are AST pattern matches over ES
 * version features / builtin APIs (no type info), so rslint reproduces ESLint
 * byte-for-byte. Representative triggers from the upstream test suite (v9.6.0).
 */
import { runConformanceSuite } from '../conformance.js';
import type { DiffCase } from '../harness.js';

const CASES: DiffCase[] = [
  { pkg: 'eslint-plugin-es-x', rule: 'no-promise', code: 'Promise' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-promise',
    code: 'function f() { Promise }',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-promise-all-settled',
    code: 'Promise.allSettled',
  },
  { pkg: 'eslint-plugin-es-x', rule: 'no-promise-any', code: 'Promise.any' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-promise-any', code: 'AggregateError' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-promise-any',
    code: 'console.log(e instanceof AggregateError)',
  },
  { pkg: 'eslint-plugin-es-x', rule: 'no-promise-try', code: 'Promise.try' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-promise-withresolvers',
    code: 'Promise.withResolvers',
  },
];

const CLEAN_CASES: DiffCase[] = [
  { pkg: 'eslint-plugin-es-x', rule: 'no-promise', code: 'Array' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-promise', code: 'Object' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-promise-all-settled',
    code: 'Promise.all',
  },
  { pkg: 'eslint-plugin-es-x', rule: 'no-promise-any', code: 'Promise.all' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-promise-any', code: 'Error' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-promise-prototype-finally',
    code: 'foo.finally(() => {})',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-promise-prototype-finally',
    code: 'foo.then(() => {})',
  },
  { pkg: 'eslint-plugin-es-x', rule: 'no-promise-try', code: 'Promise.all' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-promise-withresolvers',
    code: 'Promise.all',
  },
];

runConformanceSuite('eslint-plugin-es-x', CASES, CLEAN_CASES);
