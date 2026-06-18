/**
 * Conformance: eslint-plugin-es-x (iterator) mounted in rslint via `plugins` must
 * report identically to ESLint v10. es-x rules are AST pattern matches over ES
 * version features / builtin APIs (no type info), so rslint reproduces ESLint
 * byte-for-byte. Representative triggers from the upstream test suite (v9.6.0).
 */
import { runConformanceSuite } from '../conformance.js';
import type { DiffCase } from '../harness.js';

const CASES: DiffCase[] = [
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-generators',
    code: 'function* f() {}',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-generators',
    code: '(function*() {})',
  },
  { pkg: 'eslint-plugin-es-x', rule: 'no-generators', code: '({ *f() {} })' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-generators',
    code: 'class A { *f() {} }',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-iterator',
    code: 'Iterator.from(object)',
  },
  { pkg: 'eslint-plugin-es-x', rule: 'no-iterator', code: 'Iterator' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-iterator-concat',
    code: 'Iterator.concat',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-iterator-prototype-drop',
    code: 'foo().drop(3); function * foo() {  }',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-iterator-prototype-drop',
    code: 'Iterator.from(foo).drop(3);',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-iterator-prototype-every',
    code: 'foo().every(n => n % 2); function * foo() {  }',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-iterator-prototype-every',
    code: 'Iterator.from(foo).every(n => n % 2);',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-iterator-prototype-filter',
    code: 'foo().filter(n => n % 2); function * foo() {  }',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-iterator-prototype-filter',
    code: 'Iterator.from(foo).filter(n => n % 2);',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-iterator-prototype-find',
    code: 'foo().find(n => n % 2); function * foo() {  }',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-iterator-prototype-find',
    code: 'Iterator.from(foo).find(n => n % 2);',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-iterator-prototype-flatmap',
    code: 'foo().flatMap(n => [n, -n]); function * foo() {  }',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-iterator-prototype-flatmap',
    code: 'Iterator.from(foo).flatMap(n => [n, -n]);',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-iterator-prototype-foreach',
    code: 'foo().forEach(n => {/* ... */}); function * foo() {  }',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-iterator-prototype-foreach',
    code: 'Iterator.from(foo).forEach(n => {/* ... */});',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-iterator-prototype-map',
    code: 'foo().map(n => n % 2); function * foo() {  }',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-iterator-prototype-map',
    code: 'Iterator.from(foo).map(n => n % 2);',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-iterator-prototype-reduce',
    code: 'foo().reduce((sum, value) => sum + value, 3); function * foo() {  }',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-iterator-prototype-reduce',
    code: 'Iterator.from(foo).reduce((sum, value) => sum + value, 3);',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-iterator-prototype-some',
    code: 'foo().some(n => n % 2); function * foo() {  }',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-iterator-prototype-some',
    code: 'Iterator.from(foo).some(n => n % 2);',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-iterator-prototype-take',
    code: 'foo().take(3); function * foo() {  }',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-iterator-prototype-take',
    code: 'Iterator.from(foo).take(3);',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-iterator-prototype-toarray',
    code: 'foo().toArray(); function * foo() {  }',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-iterator-prototype-toarray',
    code: 'Iterator.from(foo).toArray();',
  },
];

const CLEAN_CASES: DiffCase[] = [
  { pkg: 'eslint-plugin-es-x', rule: 'no-generators', code: 'function f() {}' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-generators', code: 'yield = 0' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-iterator',
    code: 'Array.from(object)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-iterator',
    code: 'const Iterator = Array; Iterator.from(object)',
  },
  { pkg: 'eslint-plugin-es-x', rule: 'no-iterator-concat', code: 'Iterator' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-iterator-concat',
    code: 'Iterator.length',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-iterator-prototype-drop',
    code: 'foo.drop(3)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-iterator-prototype-drop',
    code: 'drop(3)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-iterator-prototype-drop',
    code: 'foo.unknown(0)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-iterator-prototype-every',
    code: 'foo.every(n => n % 2)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-iterator-prototype-every',
    code: 'every(n => n % 2)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-iterator-prototype-every',
    code: 'foo.unknown(0)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-iterator-prototype-filter',
    code: 'foo.filter(n => n % 2)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-iterator-prototype-filter',
    code: 'filter(n => n % 2)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-iterator-prototype-filter',
    code: 'foo.unknown(0)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-iterator-prototype-find',
    code: 'foo.find(n => n % 2)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-iterator-prototype-find',
    code: 'find(n => n % 2)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-iterator-prototype-find',
    code: 'foo.unknown(0)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-iterator-prototype-flatmap',
    code: 'foo.flatMap(n => [n, -n])',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-iterator-prototype-flatmap',
    code: 'flatMap(n => [n, -n])',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-iterator-prototype-flatmap',
    code: 'foo.unknown(0)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-iterator-prototype-foreach',
    code: 'foo.forEach(n => {/* ... */})',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-iterator-prototype-foreach',
    code: 'forEach(n => {/* ... */})',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-iterator-prototype-foreach',
    code: 'foo.unknown(0)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-iterator-prototype-map',
    code: 'foo.map(n => n % 2)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-iterator-prototype-map',
    code: 'map(n => n % 2)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-iterator-prototype-map',
    code: 'foo.unknown(0)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-iterator-prototype-reduce',
    code: 'foo.reduce((sum, value) => sum + value, 3)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-iterator-prototype-reduce',
    code: 'reduce((sum, value) => sum + value, 3)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-iterator-prototype-reduce',
    code: 'foo.unknown(0)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-iterator-prototype-some',
    code: 'foo.some(n => n % 2)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-iterator-prototype-some',
    code: 'some(n => n % 2)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-iterator-prototype-some',
    code: 'foo.unknown(0)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-iterator-prototype-take',
    code: 'foo.take(3)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-iterator-prototype-take',
    code: 'take(3)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-iterator-prototype-take',
    code: 'foo.unknown(0)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-iterator-prototype-toarray',
    code: 'foo.toArray()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-iterator-prototype-toarray',
    code: 'toArray()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-iterator-prototype-toarray',
    code: 'foo.unknown(0)',
  },
];

runConformanceSuite('eslint-plugin-es-x', CASES, CLEAN_CASES);
