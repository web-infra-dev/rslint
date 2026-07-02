/**
 * Conformance: eslint-plugin-es-x (collections) mounted in rslint via `plugins` must
 * report identically to ESLint v10. es-x rules are AST pattern matches over ES
 * version features / builtin APIs (no type info), so rslint reproduces ESLint
 * byte-for-byte. Representative triggers from the upstream test suite (v9.6.0).
 */
import { runConformanceSuite } from '../conformance.js';
import type { DiffCase } from '../harness.js';

const CASES: DiffCase[] = [
  { pkg: 'eslint-plugin-es-x', rule: 'no-map', code: 'Map' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-map', code: 'function f() { Map }' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-map-groupby', code: 'Map.groupBy' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-set', code: 'Set' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-set', code: 'function f() { Set }' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-weak-map', code: 'WeakMap' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-weak-map',
    code: 'function f() { WeakMap }',
  },
  { pkg: 'eslint-plugin-es-x', rule: 'no-weak-set', code: 'WeakSet' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-weak-set',
    code: 'function f() { WeakSet }',
  },
  { pkg: 'eslint-plugin-es-x', rule: 'no-weakrefs', code: 'WeakRef' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-weakrefs',
    code: 'function f() { WeakRef }',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-weakrefs',
    code: 'FinalizationRegistry',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-weakrefs',
    code: 'function f() { FinalizationRegistry }',
  },
];

const CLEAN_CASES: DiffCase[] = [
  { pkg: 'eslint-plugin-es-x', rule: 'no-map', code: 'Array' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-map', code: 'Object' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-map-groupby', code: 'Object' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-map-groupby', code: 'Map' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-map-prototype-getorinsert',
    code: 'foo.(key, value)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-map-prototype-getorinsert',
    code: '(key, value)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-map-prototype-getorinsertcomputed',
    code: 'foo.(key, callbackFn)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-map-prototype-getorinsertcomputed',
    code: '(key, callbackFn)',
  },
  { pkg: 'eslint-plugin-es-x', rule: 'no-set', code: 'Array' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-set', code: 'Object' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-set-prototype-difference',
    code: 'foo.(other)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-set-prototype-difference',
    code: '(other)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-set-prototype-intersection',
    code: 'foo.(other)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-set-prototype-intersection',
    code: '(other)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-set-prototype-isdisjointfrom',
    code: 'foo.(other)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-set-prototype-isdisjointfrom',
    code: '(other)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-set-prototype-issubsetof',
    code: 'foo.(other)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-set-prototype-issubsetof',
    code: '(other)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-set-prototype-issupersetof',
    code: 'foo.(other)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-set-prototype-issupersetof',
    code: '(other)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-set-prototype-symmetricdifference',
    code: 'foo.(other)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-set-prototype-symmetricdifference',
    code: '(other)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-set-prototype-union',
    code: 'foo.(other)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-set-prototype-union',
    code: '(other)',
  },
  { pkg: 'eslint-plugin-es-x', rule: 'no-weak-map', code: 'Array' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-weak-map', code: 'Object' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-weak-set', code: 'Array' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-weak-set', code: 'Object' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-weakrefs', code: 'Array' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-weakrefs', code: 'Object' },
];

runConformanceSuite('eslint-plugin-es-x', CASES, CLEAN_CASES);
