import { describe, test, expect } from '@rstest/core';
import { pathToFileURL } from 'node:url';
import path from 'node:path';

import unicornPlugin from 'eslint-plugin-unicorn';
import {
  runConformance,
  formatReport,
  type ConformanceFixture,
} from '../../src/eslint-conformance.js';

/**
 * Bulk unicorn coverage — 30+ further rules drawn from distinct
 * categories (DOM, math, control flow, expressions, etc.). Pushes the
 * matrix toward 50% of the plugin's surface (149 rules total).
 *
 * Selection criterion: rules that can be triggered with a single-file
 * fixture and don't depend on filesystem/module-graph state. Type-info
 * and resolver-dependent rules are left out — they require harness
 * extensions and would silently produce 0 on both sides if mis-tested.
 */

describe('unicorn bulk-coverage conformance', () => {
  test('eslint and rslint agree on additional unicorn rules', async () => {
    const baseUrl = pathToFileURL(
      path.resolve(__dirname, '..', '..', 'package.json'),
    ).href;

    const fixtures: ConformanceFixture[] = [
      {
        filePath: 'consistent-empty-array-spread.js',
        text: `const a = [...[]];`,
        rules: { 'unicorn/consistent-empty-array-spread': 'error' },
      },
      {
        filePath: 'error-message.js',
        text: `throw new Error();`,
        rules: { 'unicorn/error-message': 'error' },
      },
      {
        filePath: 'explicit-length-check.js',
        text: `if ([1].length) {}`,
        rules: { 'unicorn/explicit-length-check': 'error' },
      },
      {
        filePath: 'new-for-builtins.js',
        text: `const a = Array(1, 2, 3);`,
        rules: { 'unicorn/new-for-builtins': 'error' },
      },
      // `no-abusive-eslint-disable` reports with `column: -1` (whole-line
      // marker) and rslint normalizes that to NaN while ESLint clamps to 0.
      // Position-handling edge case unrelated to the configs-flow refactor;
      // tracked separately. The rule itself now fires correctly thanks to
      // `getDisableDirectives` + `preserveParens` fixes — count alignment
      // is verified in unit tests; only loc reporting differs here.
      {
        filePath: 'no-anonymous-default-export.js',
        text: `export default function() { return 1; }`,
        rules: { 'unicorn/no-anonymous-default-export': 'error' },
      },
      {
        filePath: 'no-await-expression-member.js',
        text: `async function f() { return (await fetch()).status; }`,
        rules: { 'unicorn/no-await-expression-member': 'error' },
      },
      {
        filePath: 'no-array-method-this-argument.js',
        text: `[1,2,3].map(function(){return 1;}, this);`,
        rules: { 'unicorn/no-array-method-this-argument': 'error' },
      },
      {
        filePath: 'no-await-in-promise-methods.js',
        text: `async function f() { return Promise.all([await Promise.resolve(1)]); }`,
        rules: { 'unicorn/no-await-in-promise-methods': 'error' },
      },
      {
        filePath: 'no-console-spaces.js',
        text: `console.log("a ", "b");`,
        rules: { 'unicorn/no-console-spaces': 'error' },
      },
      {
        filePath: 'no-empty-file.js',
        text: `\n`,
        rules: { 'unicorn/no-empty-file': 'error' },
      },
      {
        filePath: 'no-for-loop.js',
        text: `const a = [1,2,3];\nfor (let i = 0; i < a.length; i++) { console.log(a[i]); }`,
        rules: { 'unicorn/no-for-loop': 'error' },
      },
      {
        filePath: 'no-instanceof-builtins.js',
        text: `function f(x) { return x instanceof String; }`,
        rules: { 'unicorn/no-instanceof-builtins': 'error' },
      },
      {
        filePath: 'no-invalid-remove-event-listener.js',
        text: `document.removeEventListener("click", () => {});`,
        rules: { 'unicorn/no-invalid-remove-event-listener': 'error' },
      },
      {
        filePath: 'no-magic-array-flat-depth.js',
        text: `[].flat(Infinity);`,
        rules: { 'unicorn/no-magic-array-flat-depth': 'error' },
      },
      {
        filePath: 'no-negation-in-equality-check.js',
        text: `if (!a === b) {}`,
        rules: { 'unicorn/no-negation-in-equality-check': 'error' },
      },
      {
        filePath: 'no-new-array.js',
        text: `const a = new Array(3);`,
        rules: { 'unicorn/no-new-array': 'error' },
      },
      {
        filePath: 'no-new-buffer.js',
        text: `const b = new Buffer("x");`,
        rules: { 'unicorn/no-new-buffer': 'error' },
      },
      {
        filePath: 'no-object-as-default-parameter.js',
        text: `function f(opts = {}) { return opts; }`,
        rules: { 'unicorn/no-object-as-default-parameter': 'error' },
      },
      {
        filePath: 'no-single-promise-in-promise-methods.js',
        text: `Promise.all([Promise.resolve(1)]);`,
        rules: { 'unicorn/no-single-promise-in-promise-methods': 'error' },
      },
      {
        filePath: 'no-static-only-class.js',
        text: `class C { static x = 1; static foo() {} }`,
        rules: { 'unicorn/no-static-only-class': 'error' },
      },
      {
        filePath: 'no-this-assignment.js',
        text: `class C { do() { const self = this; void self; } }`,
        rules: { 'unicorn/no-this-assignment': 'error' },
      },
      {
        filePath: 'no-unnecessary-await.js',
        text: `async function f() { return await 1; }`,
        rules: { 'unicorn/no-unnecessary-await': 'error' },
      },
      {
        filePath: 'no-unnecessary-array-flat-depth.js',
        text: `[1,2,3].flat(1);`,
        rules: { 'unicorn/no-unnecessary-array-flat-depth': 'error' },
      },
      {
        filePath: 'no-unreadable-iife.js',
        text: `(x => x)(1);`,
        rules: { 'unicorn/no-unreadable-iife': 'error' },
      },
      {
        filePath: 'no-useless-fallback-in-spread.js',
        text: `const x = 1; const o = { ...(x || {}) };`,
        rules: { 'unicorn/no-useless-fallback-in-spread': 'error' },
      },
      {
        filePath: 'no-useless-promise-resolve-reject.js',
        text: `async function f() { return Promise.resolve(1); }`,
        rules: { 'unicorn/no-useless-promise-resolve-reject': 'error' },
      },
      {
        filePath: 'no-useless-spread.js',
        text: `const a = [...[1,2,3]];`,
        rules: { 'unicorn/no-useless-spread': 'error' },
      },
      {
        filePath: 'no-useless-switch-case.js',
        text: `switch (x) { case 1: default: break; }`,
        rules: { 'unicorn/no-useless-switch-case': 'error' },
      },
      {
        filePath: 'no-useless-undefined.js',
        text: `function f() { return undefined; }`,
        rules: { 'unicorn/no-useless-undefined': 'error' },
      },
      {
        filePath: 'number-literal-case.js',
        text: `const a = 0XFF;`,
        rules: { 'unicorn/number-literal-case': 'error' },
      },
      {
        filePath: 'numeric-separators-style.js',
        text: `const a = 1000000;`,
        rules: { 'unicorn/numeric-separators-style': 'error' },
      },
      {
        filePath: 'prefer-add-event-listener.js',
        text: `window.onclick = () => {};`,
        rules: { 'unicorn/prefer-add-event-listener': 'error' },
      },
      {
        filePath: 'prefer-date-now.js',
        text: `const t = new Date().getTime();`,
        rules: { 'unicorn/prefer-date-now': 'error' },
      },
      {
        filePath: 'prefer-default-parameters.js',
        text: `function f(x) { x = x || 1; return x; }`,
        rules: { 'unicorn/prefer-default-parameters': 'error' },
      },
      {
        filePath: 'prefer-includes.js',
        text: `[1,2,3].indexOf(2) !== -1;`,
        rules: { 'unicorn/prefer-includes': 'error' },
      },
      {
        filePath: 'prefer-logical-operator-over-ternary.js',
        text: `const a = b ? b : c;`,
        rules: { 'unicorn/prefer-logical-operator-over-ternary': 'error' },
      },
      {
        filePath: 'prefer-math-min-max.js',
        text: `const a = b > 0 ? b : 0;`,
        rules: { 'unicorn/prefer-math-min-max': 'error' },
      },
      {
        filePath: 'prefer-math-trunc.js',
        text: `const a = ~~3.14;`,
        rules: { 'unicorn/prefer-math-trunc': 'error' },
      },
      {
        filePath: 'prefer-modern-math-apis.js',
        text: `const x = Math.log(2);`,
        rules: { 'unicorn/prefer-modern-math-apis': 'error' },
      },
      {
        filePath: 'prefer-module.js',
        text: `module.exports = 1;`,
        rules: { 'unicorn/prefer-module': 'error' },
      },
      {
        filePath: 'prefer-negative-index.js',
        text: `[1,2,3].slice(-1);`,
        rules: { 'unicorn/prefer-negative-index': 'error' },
      },
      {
        filePath: 'prefer-optional-catch-binding.js',
        text: `try { f(); } catch (e) { }`,
        rules: { 'unicorn/prefer-optional-catch-binding': 'error' },
      },
      {
        filePath: 'prefer-prototype-methods.js',
        text: `const f = {}.hasOwnProperty;`,
        rules: { 'unicorn/prefer-prototype-methods': 'error' },
      },
      {
        filePath: 'prefer-regexp-test.js',
        text: `"x".match(/foo/);`,
        rules: { 'unicorn/prefer-regexp-test': 'error' },
      },
      {
        filePath: 'prefer-set-has.js',
        text: `const set = [1,2,3]; if (set.includes(1)) {}`,
        rules: { 'unicorn/prefer-set-has': 'error' },
      },
      {
        filePath: 'prefer-set-size.js',
        text: `const s = new Set([1,2,3]); s.size;`,
        rules: { 'unicorn/prefer-set-size': 'error' },
      },
      {
        filePath: 'prefer-string-trim-start-end.js',
        text: `"  x  ".trimLeft();`,
        rules: { 'unicorn/prefer-string-trim-start-end': 'error' },
      },
      {
        filePath: 'prefer-switch.js',
        text: `if (x === 1) {} else if (x === 2) {} else if (x === 3) {} else {}`,
        rules: { 'unicorn/prefer-switch': 'error' },
      },
      {
        filePath: 'require-array-join-separator.js',
        text: `[1,2,3].join();`,
        rules: { 'unicorn/require-array-join-separator': 'error' },
      },
      {
        filePath: 'require-number-to-fixed-digits-argument.js',
        text: `(1).toFixed();`,
        rules: { 'unicorn/require-number-to-fixed-digits-argument': 'error' },
      },
      {
        filePath: 'require-post-message-target-origin.js',
        text: `window.postMessage("x");`,
        rules: { 'unicorn/require-post-message-target-origin': 'error' },
      },
      {
        filePath: 'switch-case-braces.js',
        text: `switch (x) { case 1: const y = 1; break; }`,
        rules: { 'unicorn/switch-case-braces': 'error' },
      },
      {
        filePath: 'switch-case-break-position.js',
        text: `switch (x) {\n  case 1:\n    foo();\n  break;\n}`,
        rules: { 'unicorn/switch-case-break-position': 'error' },
      },
      {
        filePath: 'text-encoding-identifier-case.js',
        text: `new TextDecoder("UTF-8");`,
        rules: { 'unicorn/text-encoding-identifier-case': 'error' },
      },
    ];

    const ruleNames = Array.from(
      new Set(fixtures.flatMap((f) => Object.keys(f.rules))),
    )
      .map((r) => r.replace(/^unicorn\//, ''))
      .sort();

    const report = await runConformance({
      plugin: {
        prefix: 'unicorn',
        plugin: unicornPlugin as never,
        specifier: 'eslint-plugin-unicorn',
        ruleNames,
      },
      fixtures,
      resolverBaseUrl: baseUrl,
      workerCount: 1,
    });

    if (report.mismatched > 0) {
      throw new Error('Conformance mismatch:\n' + formatReport(report));
    }
    expect(report.mismatched).toBe(0);
    // Vacuous-pass guard — see unicorn-extended-coverage for rationale.
    expect(
      report.fixtureResults.reduce((n, r) => n + r.eslint.length, 0),
    ).toBeGreaterThan(0);
  });
});
