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
    code: 'foo.getOrInsert(key, value)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-map-prototype-getorinsert',
    code: 'getOrInsert(key, value)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-map-prototype-getorinsertcomputed',
    code: 'foo.getOrInsertComputed(key, callbackFn)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-map-prototype-getorinsertcomputed',
    code: 'getOrInsertComputed(key, callbackFn)',
  },
  { pkg: 'eslint-plugin-es-x', rule: 'no-set', code: 'Array' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-set', code: 'Object' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-set-prototype-difference',
    code: 'foo.difference(other)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-set-prototype-difference',
    code: 'difference(other)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-set-prototype-intersection',
    code: 'foo.intersection(other)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-set-prototype-intersection',
    code: 'intersection(other)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-set-prototype-isdisjointfrom',
    code: 'foo.isDisjointFrom(other)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-set-prototype-isdisjointfrom',
    code: 'isDisjointFrom(other)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-set-prototype-issubsetof',
    code: 'foo.isSubsetOf(other)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-set-prototype-issubsetof',
    code: 'isSubsetOf(other)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-set-prototype-issupersetof',
    code: 'foo.isSupersetOf(other)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-set-prototype-issupersetof',
    code: 'isSupersetOf(other)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-set-prototype-symmetricdifference',
    code: 'foo.symmetricDifference(other)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-set-prototype-symmetricdifference',
    code: 'symmetricDifference(other)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-set-prototype-union',
    code: 'foo.union(other)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-set-prototype-union',
    code: 'union(other)',
  },
  { pkg: 'eslint-plugin-es-x', rule: 'no-weak-map', code: 'Array' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-weak-map', code: 'Object' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-weak-set', code: 'Array' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-weak-set', code: 'Object' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-weakrefs', code: 'Array' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-weakrefs', code: 'Object' },
];

runConformanceSuite('eslint-plugin-es-x', CASES, CLEAN_CASES);
