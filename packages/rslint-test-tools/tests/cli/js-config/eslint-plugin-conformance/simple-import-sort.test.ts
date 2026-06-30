/**
 * Conformance: eslint-plugin-simple-import-sort rules mounted in rslint via
 * `plugins` must report identically to ESLint v10. Shared assertion lives in
 * ./conformance.ts.
 *
 * Both rules emit a single `sort` diagnostic spanning the unsorted block. The
 * sort is pure (no type information), so rslint and ESLint agree byte-for-byte
 * on the message and the reported range. Cases are representative triggers
 * drawn from the upstream test suite (v13.0.0), one or more per scenario.
 */
import { runConformanceSuite } from './conformance.js';
import type { DiffCase } from './harness.js';

const PKG = 'eslint-plugin-simple-import-sort';

/** Triggers that must report IDENTICALLY on both engines. */
const CASES: DiffCase[] = [
  // imports — alphabetical sort of statements.
  {
    pkg: PKG,
    rule: 'imports',
    code: 'import x2 from "b"\nimport x1 from "a";\n',
  },
  // imports — semicolon-free style with a start-of-line guarding semicolon.
  {
    pkg: PKG,
    rule: 'imports',
    code: 'import x2 from "b"\nimport x1 from "a"\n\n;[].forEach()\n',
  },
  // imports — sorting named specifiers.
  {
    pkg: PKG,
    rule: 'imports',
    code: 'import { e, b, a as c } from "specifiers"\n',
  },
  // imports — specifiers alongside a default import.
  {
    pkg: PKG,
    rule: 'imports',
    code: 'import d, { e, b, a as c } from "specifiers-default"\n',
  },
  // imports — renamed specifiers sort by their external (stable) name.
  {
    pkg: PKG,
    rule: 'imports',
    code: 'import { a as c, a as b2, b, a } from "specifiers-renames"\n',
  },
  // imports — no spaces inside the braces.
  {
    pkg: PKG,
    rule: 'imports',
    code: 'import {e,b,a as c} from "specifiers-no-spaces"\n',
  },
  // imports — specifiers carrying line comments.
  {
    pkg: PKG,
    rule: 'imports',
    code: 'import {\n  // c\n  c,\n  b, // b\n  a\n  // last\n} from "specifiers-comments"\n',
  },

  // exports — alphabetical sort of statements.
  {
    pkg: PKG,
    rule: 'exports',
    code: 'export {x2} from "b"\nexport {x1} from "a";\n',
  },
  // exports — semicolon-free style with a guarding semicolon.
  {
    pkg: PKG,
    rule: 'exports',
    code: 'export {x2} from "b"\nexport {x1} from "a"\n\n;[].forEach()\n',
  },
  // exports — sorting named specifiers (the stable name wins on `a as c`).
  {
    pkg: PKG,
    rule: 'exports',
    code: 'export { d, a as c, a as b2, b, a } from "specifiers"\n',
  },
  // exports — `as default`.
  {
    pkg: PKG,
    rule: 'exports',
    code: "export { something, something as default } from './something'\n",
  },
  // exports — a type specifier sorts ahead of the same-named value specifier.
  {
    pkg: PKG,
    rule: 'exports',
    code: 'export {MyClass, type MyClass} from "../type";\n',
  },
];

/** Snippets that must report NOTHING on both engines. */
const CLEAN_CASES: DiffCase[] = [
  // imports — already-correct forms.
  { pkg: PKG, rule: 'imports', code: 'import "a"\n' },
  { pkg: PKG, rule: 'imports', code: 'import a from "a"\n' },
  { pkg: PKG, rule: 'imports', code: 'import {a,b} from "a"\n' },
  { pkg: PKG, rule: 'imports', code: 'import * as a from "a"\n' },
  {
    pkg: PKG,
    rule: 'imports',
    code: 'import x1 from "a";\nimport x2 from "b"\n',
  },
  // Side-effect-only imports keep their original order.
  { pkg: PKG, rule: 'imports', code: 'import "b";\nimport "a"\n' },

  // exports — already-correct forms.
  { pkg: PKG, rule: 'exports', code: 'export {a} from "a"\n' },
  { pkg: PKG, rule: 'exports', code: 'export {a,b} from "a"\n' },
  { pkg: PKG, rule: 'exports', code: 'export * as a from "a"\n' },
  { pkg: PKG, rule: 'exports', code: 'export var one = 1;\n' },
  { pkg: PKG, rule: 'exports', code: 'export default whatever;\n' },
  {
    pkg: PKG,
    rule: 'exports',
    code: 'export type {x1} from "a";\nexport type {x2} from "b"\n',
  },
  {
    pkg: PKG,
    rule: 'exports',
    code: 'export { a, type b, c, type d } from "a"\n',
  },
];

runConformanceSuite(PKG, CASES, CLEAN_CASES);
