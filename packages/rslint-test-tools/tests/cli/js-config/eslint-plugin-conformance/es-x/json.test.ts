/**
 * Conformance: eslint-plugin-es-x (json) mounted in rslint via `plugins` must
 * report identically to ESLint v10. es-x rules are AST pattern matches over ES
 * version features / builtin APIs (no type info), so rslint reproduces ESLint
 * byte-for-byte. Representative triggers from the upstream test suite (v9.6.0).
 */
import { runConformanceSuite } from '../conformance.js';
import type { DiffCase } from '../harness.js';

const CASES: DiffCase[] = [
  { pkg: 'eslint-plugin-es-x', rule: 'no-json', code: 'JSON' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-json', code: 'JSON.parse' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-json', code: 'JSON.stringify' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-json-israwjson',
    code: 'JSON.isRawJSON',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-json-israwjson',
    code: 'if (JSON.isRawJSON) JSON.isRawJSON(foo)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-json-israwjson',
    code: 'JSON.isRawJSON(foo)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-json-modules',
    code: "import foo from 'foo' with { type: 'json' }",
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-json-modules',
    code: "export {foo} from 'foo' with { type: 'json' }",
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-json-modules',
    code: "export * from 'foo' with { type: 'json' }",
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-json-modules',
    code: "import('foo', { with: { type: 'json'} })",
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-json-parse-reviver-context-parameter',
    code: 'JSON.parse("{}", (key, value, context) => { return value; })',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-json-parse-reviver-context-parameter',
    code: 'JSON.parse("{}", function(key, value, context) { return value; })',
  },
  { pkg: 'eslint-plugin-es-x', rule: 'no-json-rawjson', code: 'JSON.rawJSON' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-json-rawjson',
    code: 'if (JSON.rawJSON) JSON.rawJSON(foo)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-json-rawjson',
    code: 'JSON.rawJSON(foo)',
  },
];

const CLEAN_CASES: DiffCase[] = [
  { pkg: 'eslint-plugin-es-x', rule: 'no-json', code: 'let JSON = 0; JSON' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-json-israwjson', code: 'JSON' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-json-israwjson', code: 'JSON.parse' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-json-modules',
    code: "import foo from 'foo'",
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-json-modules',
    code: "export {foo} from 'foo'",
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-json-parse-reviver-context-parameter',
    code: 'JSON.parse("{}")',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-json-parse-reviver-context-parameter',
    code: 'JSON.parse("{}", (key, value) => value)',
  },
  { pkg: 'eslint-plugin-es-x', rule: 'no-json-rawjson', code: 'JSON' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-json-rawjson', code: 'JSON.parse' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-json-superset', code: 'let a = null' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-json-superset',
    code: "let a = '\\u2028'",
  },
];

runConformanceSuite('eslint-plugin-es-x', CASES, CLEAN_CASES);
