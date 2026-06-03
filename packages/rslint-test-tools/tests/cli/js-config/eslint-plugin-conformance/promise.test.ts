/**
 * Conformance: eslint-plugin-promise rules mounted in rslint via `eslintPlugins` must
 * report identically to ESLint v10. Shared assertion + excluded-category notes
 * live in ./conformance.ts.
 */
import { runConformanceSuite } from './conformance.js';
import type { DiffCase } from './harness.js';

/** 15 rules that report IDENTICALLY on a minimal trigger. */
const CASES: DiffCase[] = [
  {
    pkg: 'eslint-plugin-promise',
    rule: 'avoid-new',
    code: 'const p = new Promise(function (resolve, reject) {\n  resolve(1);\n});\n',
  },
  {
    pkg: 'eslint-plugin-promise',
    rule: 'catch-or-return',
    code: 'myPromise.then(function () {\n  return 1;\n});\n',
  },
  {
    pkg: 'eslint-plugin-promise',
    rule: 'no-callback-in-promise',
    code: 'myPromise.then(function () {\n  callback(null, 1);\n});\n',
  },
  {
    pkg: 'eslint-plugin-promise',
    rule: 'no-native',
    code: 'const p = Promise.resolve(1);\n',
  },
  {
    pkg: 'eslint-plugin-promise',
    rule: 'no-nesting',
    code: 'myPromise.then(function () {\n  return inner.then(function () {\n    return 1;\n  });\n});\n',
  },
  {
    pkg: 'eslint-plugin-promise',
    rule: 'no-new-statics',
    code: 'const p = new Promise.resolve(1);\n',
  },
  {
    pkg: 'eslint-plugin-promise',
    rule: 'no-promise-in-callback',
    code: 'doSomething(function (err) {\n  if (err) return;\n  myPromise.then(function () {});\n});\n',
  },
  {
    pkg: 'eslint-plugin-promise',
    rule: 'no-return-in-finally',
    code: 'myPromise.finally(function () {\n  return 1;\n});\n',
  },
  {
    pkg: 'eslint-plugin-promise',
    rule: 'no-return-wrap',
    code: 'myPromise.then(function () {\n  return Promise.resolve(1);\n});\n',
  },
  {
    pkg: 'eslint-plugin-promise',
    rule: 'param-names',
    code: 'const p = new Promise(function (reject, resolve) {\n  resolve(1);\n});\n',
  },
  {
    pkg: 'eslint-plugin-promise',
    rule: 'prefer-await-to-callbacks',
    code: 'function f(callback) {\n  callback();\n}\n',
  },
  {
    pkg: 'eslint-plugin-promise',
    rule: 'prefer-await-to-then',
    code: 'function f() {\n  return myPromise.then(function () {\n    return 1;\n  });\n}\n',
  },
  {
    pkg: 'eslint-plugin-promise',
    rule: 'prefer-catch',
    code: 'myPromise.then(handleResolve, handleReject);\n',
  },
  {
    pkg: 'eslint-plugin-promise',
    rule: 'spec-only',
    code: 'const p = Promise.foo(1);\n',
  },
  {
    pkg: 'eslint-plugin-promise',
    rule: 'valid-params',
    code: 'const p = Promise.resolve(1, 2);\n',
  },
];

const CLEAN_CASES: DiffCase[] = [
  {
    pkg: 'eslint-plugin-promise',
    rule: 'prefer-await-to-then',
    code: 'async function f() { const x = await g(); return x; }\n',
  },
];

runConformanceSuite('eslint-plugin-promise', CASES, CLEAN_CASES);
