/**
 * Conformance: eslint-plugin-es-x (object) mounted in rslint via `plugins` must
 * report identically to ESLint v10. es-x rules are AST pattern matches over ES
 * version features / builtin APIs (no type info), so rslint reproduces ESLint
 * byte-for-byte. Representative triggers from the upstream test suite (v9.6.0).
 */
import { runConformanceSuite } from '../conformance.js';
import type { DiffCase } from '../harness.js';

const CASES: DiffCase[] = [
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-object-assign',
    code: 'Object.assign',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-object-create',
    code: 'Object.create',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-object-defineproperties',
    code: 'Object.defineProperties',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-object-defineproperty',
    code: 'Object.defineProperty',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-object-entries',
    code: 'Object.entries',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-object-freeze',
    code: 'Object.freeze',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-object-fromentries',
    code: 'Object.fromEntries',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-object-getownpropertydescriptor',
    code: 'Object.getOwnPropertyDescriptor',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-object-getownpropertydescriptors',
    code: 'Object.getOwnPropertyDescriptors',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-object-getownpropertynames',
    code: 'Object.getOwnPropertyNames',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-object-getownpropertysymbols',
    code: 'Object.getOwnPropertySymbols',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-object-getprototypeof',
    code: 'Object.getPrototypeOf',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-object-groupby',
    code: 'Object.groupBy',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-object-hasown',
    code: 'Object.hasOwn',
  },
  { pkg: 'eslint-plugin-es-x', rule: 'no-object-is', code: 'Object.is' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-object-isextensible',
    code: 'Object.isExtensible',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-object-isfrozen',
    code: 'Object.isFrozen',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-object-issealed',
    code: 'Object.isSealed',
  },
  { pkg: 'eslint-plugin-es-x', rule: 'no-object-keys', code: 'Object.keys' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-object-map-groupby',
    code: 'Object.groupBy',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-object-map-groupby',
    code: 'Map.groupBy',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-object-preventextensions',
    code: 'Object.preventExtensions',
  },
  { pkg: 'eslint-plugin-es-x', rule: 'no-object-seal', code: 'Object.seal' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-object-setprototypeof',
    code: 'Object.setPrototypeOf',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-object-super-properties',
    code: '({ foo() { super.a } })',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-object-super-properties',
    code: '({ foo() { super.foo() } })',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-object-super-properties',
    code: '({ foo() { return () => super.a } })',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-object-super-properties',
    code: '({ foo() { ({ foo() { return () => super.a } }); class A { foo() { super.a } } return () => super.a } })',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-object-values',
    code: 'Object.values',
  },
];

const CLEAN_CASES: DiffCase[] = [
  { pkg: 'eslint-plugin-es-x', rule: 'no-object-assign', code: 'Object' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-object-assign', code: 'Object.is' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-object-create', code: 'Object' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-object-create', code: 'Object.foo' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-object-defineproperties',
    code: 'Object',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-object-defineproperties',
    code: 'Object.foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-object-defineproperty',
    code: 'Object',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-object-defineproperty',
    code: 'Object.foo',
  },
  { pkg: 'eslint-plugin-es-x', rule: 'no-object-entries', code: 'Object' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-object-entries',
    code: 'Object.assign',
  },
  { pkg: 'eslint-plugin-es-x', rule: 'no-object-freeze', code: 'Object' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-object-freeze', code: 'Object.foo' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-object-fromentries', code: 'Object' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-object-fromentries',
    code: 'Object.assign',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-object-getownpropertydescriptor',
    code: 'Object',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-object-getownpropertydescriptor',
    code: 'Object.foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-object-getownpropertydescriptors',
    code: 'Object',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-object-getownpropertydescriptors',
    code: 'Object.assign',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-object-getownpropertynames',
    code: 'Object',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-object-getownpropertynames',
    code: 'Object.foo',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-object-getownpropertysymbols',
    code: 'Object',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-object-getownpropertysymbols',
    code: 'Object.assign',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-object-getprototypeof',
    code: 'Object',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-object-getprototypeof',
    code: 'Object.foo',
  },
  { pkg: 'eslint-plugin-es-x', rule: 'no-object-groupby', code: 'Object' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-object-groupby', code: 'Map' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-object-hasown', code: 'Object' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-object-hasown',
    code: 'Object.assign',
  },
  { pkg: 'eslint-plugin-es-x', rule: 'no-object-is', code: 'Object' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-object-is', code: 'Object.assign' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-object-isextensible', code: 'Object' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-object-isextensible',
    code: 'Object.foo',
  },
  { pkg: 'eslint-plugin-es-x', rule: 'no-object-isfrozen', code: 'Object' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-object-isfrozen', code: 'Object.foo' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-object-issealed', code: 'Object' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-object-issealed', code: 'Object.foo' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-object-keys', code: 'Object' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-object-keys', code: 'Object.assign' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-object-map-groupby', code: 'Object' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-object-map-groupby', code: 'Map' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-object-preventextensions',
    code: 'Object',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-object-preventextensions',
    code: 'Object.foo',
  },
  { pkg: 'eslint-plugin-es-x', rule: 'no-object-seal', code: 'Object' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-object-seal', code: 'Object.foo' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-object-setprototypeof',
    code: 'Object',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-object-setprototypeof',
    code: 'Object.assign',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-object-super-properties',
    code: 'class A { foo() { super.a } }',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-object-super-properties',
    code: 'class A { foo() { super.foo() } }',
  },
  { pkg: 'eslint-plugin-es-x', rule: 'no-object-values', code: 'Object' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-object-values',
    code: 'Object.assign',
  },
];

runConformanceSuite('eslint-plugin-es-x', CASES, CLEAN_CASES);
