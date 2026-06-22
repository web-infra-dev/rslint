/**
 * Conformance: eslint-plugin-es-x (array) mounted in rslint via `plugins` must
 * report identically to ESLint v10. es-x rules are AST pattern matches over ES
 * version features / builtin APIs (no type info), so rslint reproduces ESLint
 * byte-for-byte. Representative triggers from the upstream test suite (v9.6.0).
 */
import { runConformanceSuite } from '../conformance.js';
import type { DiffCase } from '../harness.js';

const CASES: DiffCase[] = [
  { pkg: 'eslint-plugin-es-x', rule: 'no-array-from', code: 'Array.from' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-from',
    code: 'const {from} = Array',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-from',
    code: 'const {a:{from} = Array} = {}',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-from',
    code: 'if (Array.from) { Array.from }',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-fromasync',
    code: 'Array.fromAsync',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-fromasync',
    code: 'const {fromAsync} = Array',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-fromasync',
    code: 'const {a:{fromAsync} = Array} = {}',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-fromasync',
    code: 'if (Array.fromAsync) { Array.fromAsync }',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-isarray',
    code: 'Array.isArray',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-isarray',
    code: 'if (Array.isArray) { Array.isArray }',
  },
  { pkg: 'eslint-plugin-es-x', rule: 'no-array-of', code: 'Array.of' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-prototype-at',
    code: '[1,2,3].at(-1)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-prototype-at',
    code: 'const { at } = [];',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-prototype-findlast-findlastindex',
    code: "['foo'].findLast(predicate)",
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-prototype-findlast-findlastindex',
    code: "['foo'].findLastIndex(predicate)",
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-prototype-toreversed',
    code: "['foo'].toReversed()",
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-prototype-tosorted',
    code: "['foo'].toSorted()",
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-prototype-tospliced',
    code: "['foo'].toSpliced(1, 2)",
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-prototype-with',
    code: "['foo'].with(index, value)",
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-string-prototype-at',
    code: '[1,2,3].at(-1)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-string-prototype-at',
    code: "'123'.at(-1)",
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-arraybuffer-prototype-transfer',
    code: 'const foo = new ArrayBuffer(8); foo.transfer()',
  },
  { pkg: 'eslint-plugin-es-x', rule: 'no-typed-arrays', code: 'Int8Array' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-typed-arrays', code: 'Uint8Array' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-typed-arrays',
    code: 'Uint8ClampedArray',
  },
  { pkg: 'eslint-plugin-es-x', rule: 'no-typed-arrays', code: 'Int16Array' },
];

const CLEAN_CASES: DiffCase[] = [
  { pkg: 'eslint-plugin-es-x', rule: 'no-array-from', code: 'Array' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-array-from', code: 'Array.of' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-array-fromasync', code: 'Array' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-array-fromasync', code: 'Array.from' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-array-isarray', code: 'Array' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-array-isarray', code: 'Array.from' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-array-of', code: 'Array' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-array-of', code: 'Array.from' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-prototype-at',
    code: 'foo.at(-1)',
  },
  { pkg: 'eslint-plugin-es-x', rule: 'no-array-prototype-at', code: 'at(-1)' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-prototype-at',
    code: 'foo.reverse()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-prototype-copywithin',
    code: 'foo.copyWithin(0, 1, 2)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-prototype-copywithin',
    code: 'copyWithin(0, 1, 2)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-prototype-copywithin',
    code: 'foo.reverse()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-prototype-entries',
    code: 'foo.entries()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-prototype-entries',
    code: 'entries()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-prototype-entries',
    code: 'foo.reverse()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-prototype-every',
    code: 'foo.every(() => {})',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-prototype-every',
    code: 'every(() => {})',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-prototype-every',
    code: 'foo.reverse()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-prototype-fill',
    code: 'foo.fill(0)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-prototype-fill',
    code: 'fill(0)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-prototype-fill',
    code: 'foo.reverse()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-prototype-filter',
    code: 'foo.filter(() => {})',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-prototype-filter',
    code: 'filter(() => {})',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-prototype-filter',
    code: 'foo.reverse()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-prototype-find',
    code: 'foo.find(() => {})',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-prototype-find',
    code: 'find(() => {})',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-prototype-find',
    code: 'foo.reverse()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-prototype-findindex',
    code: 'foo.findIndex(() => {})',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-prototype-findindex',
    code: 'findIndex(() => {})',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-prototype-findindex',
    code: 'foo.reverse()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-prototype-findlast-findlastindex',
    code: 'foo.findLast(predicate)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-prototype-findlast-findlastindex',
    code: 'foo.findLastIndex(predicate)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-prototype-findlast-findlastindex',
    code: 'findLast(predicate)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-prototype-findlast-findlastindex',
    code: 'findLastIndex(predicate)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-prototype-flat',
    code: 'foo.flat(1)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-prototype-flat',
    code: 'foo.flatMap(() => {})',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-prototype-flat',
    code: 'flat(1)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-prototype-flat',
    code: 'flatMap(() => {})',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-prototype-foreach',
    code: 'foo.forEach(() => {})',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-prototype-foreach',
    code: 'forEach(() => {})',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-prototype-foreach',
    code: 'foo.reverse()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-prototype-includes',
    code: 'foo.includes(0)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-prototype-includes',
    code: 'includes(0)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-prototype-includes',
    code: 'foo.reverse()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-prototype-indexof',
    code: 'foo.indexOf(0)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-prototype-indexof',
    code: 'indexOf(0)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-prototype-indexof',
    code: 'foo.reverse()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-prototype-keys',
    code: 'foo.keys()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-prototype-keys',
    code: 'keys()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-prototype-keys',
    code: 'foo.reverse()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-prototype-lastindexof',
    code: 'foo.lastIndexOf(0)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-prototype-lastindexof',
    code: 'lastIndexOf(0)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-prototype-lastindexof',
    code: 'foo.reverse()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-prototype-map',
    code: 'foo.map(() => {})',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-prototype-map',
    code: 'map(() => {})',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-prototype-map',
    code: 'foo.reverse()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-prototype-reduce',
    code: 'foo.reduce(() => {})',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-prototype-reduce',
    code: 'reduce(() => {})',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-prototype-reduce',
    code: 'foo.reverse()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-prototype-reduceright',
    code: 'foo.reduceRight(() => {})',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-prototype-reduceright',
    code: 'reduceRight(() => {})',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-prototype-reduceright',
    code: 'foo.reverse()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-prototype-some',
    code: 'foo.some(() => {})',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-prototype-some',
    code: 'some(() => {})',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-prototype-some',
    code: 'foo.reverse()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-prototype-toreversed',
    code: 'foo.toReversed()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-prototype-toreversed',
    code: 'toReversed()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-prototype-toreversed',
    code: 'foo.find(predicate)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-prototype-tosorted',
    code: 'foo.toSorted()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-prototype-tosorted',
    code: 'toSorted()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-prototype-tosorted',
    code: 'foo.find(predicate)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-prototype-tospliced',
    code: 'foo.toSpliced(1, 2)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-prototype-tospliced',
    code: 'toSpliced(1, 2)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-prototype-tospliced',
    code: 'foo.find(predicate)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-prototype-values',
    code: 'foo.values()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-prototype-values',
    code: 'values()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-prototype-values',
    code: 'foo.reverse()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-prototype-with',
    code: 'foo.with(index, value)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-prototype-with',
    code: 'foo.find(predicate)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-string-prototype-at',
    code: 'foo.at(-1)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-string-prototype-at',
    code: 'at(-1)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-array-string-prototype-at',
    code: 'foo.reverse()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-arraybuffer-prototype-transfer',
    code: 'foo.transfer()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-arraybuffer-prototype-transfer',
    code: 'foo.transferToFixedLength()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-arraybuffer-prototype-transfer',
    code: 'foo.detached',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-arraybuffer-prototype-transfer',
    code: 'transfer()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-arraybuffer-prototype-transfer',
    code: 'transferToFixedLength()',
  },
  { pkg: 'eslint-plugin-es-x', rule: 'no-typed-arrays', code: 'Array' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-typed-arrays', code: 'Set' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-uint8array-frombase64',
    code: 'Uint8Array.',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-uint8array-frombase64',
    code: 'if (Uint8Array.) { Uint8Array. }',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-uint8array-frombase64',
    code: 'Uint8Array',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-uint8array-frombase64',
    code: 'Uint8Array.raw',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-uint8array-fromhex',
    code: 'Uint8Array.',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-uint8array-fromhex',
    code: 'if (Uint8Array.) { Uint8Array. }',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-uint8array-fromhex',
    code: 'Uint8Array',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-uint8array-fromhex',
    code: 'Uint8Array.raw',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-uint8array-prototype-setfrombase64',
    code: 'foo.(other)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-uint8array-prototype-setfrombase64',
    code: "\n            const t = new Uint8Array([])\n            if (Uint8Array.prototype.) {\n                console.log(t.(other))\n            }\n            if (typeof Uint8Array.prototype. === 'undefined') {\n                console.log(t.(other))\n            } else {\n                console.log(t.(other))\n            }\n            const a = Uint8Array.prototype.\n              ? t.(other)\n              : t.(other);",
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-uint8array-prototype-setfrombase64',
    code: '(other)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-uint8array-prototype-setfromhex',
    code: 'foo.(other)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-uint8array-prototype-setfromhex',
    code: "\n            const t = new Uint8Array([])\n            if (Uint8Array.prototype.) {\n                console.log(t.(other))\n            }\n            if (typeof Uint8Array.prototype. === 'undefined') {\n                console.log(t.(other))\n            } else {\n                console.log(t.(other))\n            }\n            const a = Uint8Array.prototype.\n              ? t.(other)\n              : t.(other);",
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-uint8array-prototype-setfromhex',
    code: '(other)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-uint8array-prototype-tobase64',
    code: 'foo.(other)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-uint8array-prototype-tobase64',
    code: "\n            const t = new Uint8Array([])\n            if (Uint8Array.prototype.) {\n                console.log(t.(other))\n            }\n            if (typeof Uint8Array.prototype. === 'undefined') {\n                console.log(t.(other))\n            } else {\n                console.log(t.(other))\n            }\n            const a = Uint8Array.prototype.\n              ? t.(other)\n              : t.(other);",
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-uint8array-prototype-tobase64',
    code: '(other)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-uint8array-prototype-tohex',
    code: 'foo.(other)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-uint8array-prototype-tohex',
    code: "\n            const t = new Uint8Array([])\n            if (Uint8Array.prototype.) {\n                console.log(t.(other))\n            }\n            if (typeof Uint8Array.prototype. === 'undefined') {\n                console.log(t.(other))\n            } else {\n                console.log(t.(other))\n            }\n            const a = Uint8Array.prototype.\n              ? t.(other)\n              : t.(other);",
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-uint8array-prototype-tohex',
    code: '(other)',
  },
];

runConformanceSuite('eslint-plugin-es-x', CASES, CLEAN_CASES);
