/**
 * Conformance: eslint-plugin-es-x (builtins-misc) mounted in rslint via `plugins` must
 * report identically to ESLint v10. es-x rules are AST pattern matches over ES
 * version features / builtin APIs (no type info), so rslint reproduces ESLint
 * byte-for-byte. Representative triggers from the upstream test suite (v9.6.0).
 */
import { runConformanceSuite } from '../conformance.js';
import type { DiffCase } from '../harness.js';

const CASES: DiffCase[] = [
  { pkg: 'eslint-plugin-es-x', rule: 'no-date-now', code: 'Date.now' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-date-prototype-totemporalinstant',
    code: 'const foo = new Date(); foo.toTemporalInstant',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-error-iserror',
    code: 'Error.isError',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-error-iserror',
    code: 'if (Error.isError) Error.isError(foo)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-error-iserror',
    code: 'Error.isError(foo)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-resizable-and-growable-arraybuffers',
    code: 'const buffer = new ArrayBuffer(8, { maxByteLength: 16 });',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-resizable-and-growable-arraybuffers',
    code: 'const buffer = new SharedArrayBuffer(8, { maxByteLength: 16 });',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-resizable-and-growable-arraybuffers',
    code: 'const buffer = new ArrayBuffer(8, { maxByteLength: 16 }); buffer.resize(8)',
  },
  { pkg: 'eslint-plugin-es-x', rule: 'no-symbol', code: 'Symbol' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-symbol',
    code: 'function f() { Symbol }',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-symbol-asyncdispose',
    code: 'Symbol.asyncDispose',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-symbol-dispose',
    code: 'Symbol.dispose',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-symbol-matchall',
    code: 'Symbol.matchAll',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-symbol-prototype-description',
    code: 'Symbol().description',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-symbol-prototype-description',
    code: 'Symbol.iterator.description',
  },
];

const CLEAN_CASES: DiffCase[] = [
  { pkg: 'eslint-plugin-es-x', rule: 'no-date-now', code: 'Date' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-date-now', code: 'Date.parse' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-date-prototype-getyear-setyear',
    code: 'foo.getFullYear()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-date-prototype-togmtstring',
    code: 'foo.toUTCString()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-date-prototype-totemporalinstant',
    code: 'foo.toTemporalInstant',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-date-prototype-totemporalinstant',
    code: 'toTemporalInstant',
  },
  { pkg: 'eslint-plugin-es-x', rule: 'no-error-iserror', code: 'Error' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-error-iserror',
    code: 'Error.captureStackTrace',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-resizable-and-growable-arraybuffers',
    code: 'const buffer = new window.ArrayBuffer(8, { maxByteLength: 16 });',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-resizable-and-growable-arraybuffers',
    code: 'new ArrayBuffer(8);',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-resizable-and-growable-arraybuffers',
    code: 'new SharedArrayBuffer(1024);',
  },
  { pkg: 'eslint-plugin-es-x', rule: 'no-symbol', code: 'Array' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-symbol', code: 'Object' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-symbol-asyncdispose', code: 'Symbol' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-symbol-asyncdispose',
    code: 'Symbol.length',
  },
  { pkg: 'eslint-plugin-es-x', rule: 'no-symbol-dispose', code: 'Symbol' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-symbol-dispose',
    code: 'Symbol.length',
  },
  { pkg: 'eslint-plugin-es-x', rule: 'no-symbol-matchall', code: 'Symbol' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-symbol-matchall',
    code: 'Symbol.length',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-symbol-prototype-description',
    code: 'foo.description',
  },
];

runConformanceSuite('eslint-plugin-es-x', CASES, CLEAN_CASES);
