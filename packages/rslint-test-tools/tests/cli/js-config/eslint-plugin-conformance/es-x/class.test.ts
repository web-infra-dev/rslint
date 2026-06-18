/**
 * Conformance: eslint-plugin-es-x (class) mounted in rslint via `plugins` must
 * report identically to ESLint v10. es-x rules are AST pattern matches over ES
 * version features / builtin APIs (no type info), so rslint reproduces ESLint
 * byte-for-byte. Representative triggers from the upstream test suite (v9.6.0).
 */
import { runConformanceSuite } from '../conformance.js';
import type { DiffCase } from '../harness.js';

const CASES: DiffCase[] = [
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-class-fields',
    code: 'class A { #foo() {} }',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-class-fields',
    code: 'class A { get #foo() {} }',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-class-fields',
    code: 'class A { set #foo(v) {} }',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-class-fields',
    code: 'class A { *#foo() {} }',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-class-instance-fields',
    code: 'class A { foo }',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-class-instance-fields',
    code: 'class A { foo = 42 }',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-class-instance-fields',
    code: 'class A { [foo] }',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-class-instance-fields',
    code: 'class A { [foo] = 42 }',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-class-private-fields',
    code: 'class A { #foo }',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-class-private-fields',
    code: 'class A { #foo = 42 }',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-class-private-fields',
    code: 'class A { static #foo }',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-class-private-fields',
    code: 'class A { static #foo = 42 }',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-class-private-methods',
    code: 'class A { #foo() {} }',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-class-private-methods',
    code: 'class A { get #foo() {} }',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-class-private-methods',
    code: 'class A { set #foo(v) {} }',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-class-private-methods',
    code: 'class A { *#foo() {} }',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-class-static-block',
    code: 'class A { static {}; }',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-class-static-block',
    code: '(class { static {} })',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-class-static-fields',
    code: 'class A { static foo }',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-class-static-fields',
    code: 'class A { static foo = 42 }',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-class-static-fields',
    code: 'class A { static [foo] }',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-class-static-fields',
    code: 'class A { static [foo] = 42 }',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-subclassing-builtins',
    code: 'class MyArray extends Array {}',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-subclassing-builtins',
    code: 'class MyBoolean extends Boolean {}',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-subclassing-builtins',
    code: 'class MyError extends Error {}',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-subclassing-builtins',
    code: 'class MyRegExp extends RegExp {}',
  },
];

const CLEAN_CASES: DiffCase[] = [
  { pkg: 'eslint-plugin-es-x', rule: 'no-class-fields', code: 'class A {}' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-class-fields',
    code: 'class A { foo() {} }',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-class-instance-fields',
    code: 'class A {}',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-class-instance-fields',
    code: 'class A { foo() {} }',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-class-private-fields',
    code: 'class A {}',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-class-private-fields',
    code: 'class A { foo() {} }',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-class-private-methods',
    code: 'class A {}',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-class-private-methods',
    code: 'class A { foo() {} }',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-class-static-block',
    code: 'class A { static f() {} }',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-class-static-block',
    code: 'class A { static get f() {} }',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-class-static-fields',
    code: 'class A {}',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-class-static-fields',
    code: 'class A { foo() {} }',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-subclassing-builtins',
    code: 'class MyObject extends Object {}',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-subclassing-builtins',
    code: 'let Array = 0; class MyArray extends Array {}',
  },
];

runConformanceSuite('eslint-plugin-es-x', CASES, CLEAN_CASES);
