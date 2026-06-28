/**
 * Conformance: @stylistic/eslint-plugin TypeScript-only spacing rules
 * (member-delimiter-style, type-annotation-spacing, type-generic-spacing,
 * type-named-tuple-spacing). These have no JS upstream test suite to draw from,
 * so the triggers are preserved from the original conformance set; each still
 * reproduces ESLint v10 byte-for-byte.
 */
import { runConformanceSuite } from '../conformance.js';
import type { DiffCase } from '../harness.js';

const CASES: DiffCase[] = [
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'member-delimiter-style',
    code: 'interface Foo {\n  a: number,\n  b: string,\n}\n',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'type-annotation-spacing',
    code: 'const a:number = 1;',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'type-generic-spacing',
    code: 'type A<T= number> = T;',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'type-named-tuple-spacing',
    code: 'type T = [name:string];',
  },
];

runConformanceSuite('@stylistic/eslint-plugin', CASES);
