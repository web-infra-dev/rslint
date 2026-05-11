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
 * Combined-rules conformance. Same eight unicorn rules are enabled on
 * EVERY fixture so that:
 *
 *   1. fixtures triggering several rules at once verify that running
 *      multiple rules in the same pass produces diagnostics identical
 *      to ESLint (no ordering/dedup drift between engines).
 *   2. clean fixtures verify negative parity — when no rule should
 *      fire, neither engine reports anything.
 *
 * This complements `unicorn-multi-rule.test.ts` (one rule per fixture)
 * by exercising rule **interaction** rather than individual rules.
 *
 * Each rule has its single-rule baseline pinned elsewhere; the role
 * of this file is purely cross-rule consistency.
 */

const ALL_RULES = {
  'unicorn/no-null': 'error',
  'unicorn/no-array-callback-reference': 'error',
  'unicorn/prefer-array-some': 'error',
  'unicorn/no-instanceof-array': 'error',
  'unicorn/prefer-array-find': 'error',
  'unicorn/prefer-string-starts-ends-with': 'error',
  'unicorn/no-array-for-each': 'error',
  'unicorn/no-typeof-undefined': 'error',
  // Pins ReferenceTracker / scope-seed / meta.defaultOptions integration —
  // each of these surfaces a different bug class:
  //   * ReferenceTracker.iterateGlobalReferences relies on globalScope
  //     having ECMA built-ins seeded (parseInt / Array / NaN).
  //   * `rule.meta.defaultOptions` must be deep-merged into user
  //     options before invoking create() — this rule ships
  //     { checkInfinity: false, checkNaN: true }.
  //   * Walk-the-parent-chain `getScope(node)` needed for the rule to
  //     resolve at the module-level CallExpression.
  'unicorn/prefer-number-properties': 'error',
  'unicorn/prefer-includes': 'error',
} as const satisfies ConformanceFixture['rules'];

describe('unicorn combined-rules conformance', () => {
  test('eslint and rslint agree when many rules fire on the same source', async () => {
    const baseUrl = pathToFileURL(
      path.resolve(__dirname, '..', '..', 'package.json'),
    ).href;

    const fixtures: ConformanceFixture[] = [
      // Three rules fire on three distinct sites:
      //   - L1 forEach   → no-array-for-each
      //   - L1 callback  → no-null (`null` literal inside the arrow)
      //   - L2           → prefer-array-find
      {
        filePath: 'array-mix.js',
        text:
          `[1,2,3].forEach(x => null);\n` +
          `const head = [1,2,3].filter(x => x > 1)[0];\n`,
        rules: ALL_RULES,
      },
      // Two unrelated rules in one function:
      //   - instanceof Array → no-instanceof-array
      //   - typeof local var === "undefined" → no-typeof-undefined
      // Plus three implicit `null` literals → no-null three times.
      {
        filePath: 'check.js',
        text:
          `function check(x) {\n` +
          `  const local = 1;\n` +
          `  if (x instanceof Array) return null;\n` +
          `  if (typeof local === "undefined") return null;\n` +
          `  return null;\n` +
          `}\n`,
        rules: ALL_RULES,
      },
      // String method patterns chained together:
      //   - .map(String)        → no-array-callback-reference
      //   - .indexOf('f') === 0 → prefer-string-starts-ends-with
      {
        filePath: 'strings.js',
        text:
          `const items = ['foo','bar'].map(String);\n` +
          `if ('foo'.indexOf('f') === 0) {}\n`,
        rules: ALL_RULES,
      },
      // prefer-array-some bypass via filter(...).length > 0:
      //   - L1 → prefer-array-some
      // Same fixture exercises no-null (no fire) and no-array-for-each
      // (no fire) to pin negative parity in the multi-rule pass.
      {
        filePath: 'some.js',
        text: `if ([1,2,3].filter(x => x > 1).length > 0) {}\n`,
        rules: ALL_RULES,
      },
      // Negative parity: clean fixture — none of the 8 rules should fire.
      // If either engine reports any diagnostic here we want to know.
      {
        filePath: 'clean.js',
        text:
          `const arr = [1,2,3];\n` +
          `const found = arr.find(x => x > 1);\n` +
          `const has = arr.some(x => x > 1);\n` +
          `if (arr.includes(2)) {}\n` +
          `if (Array.isArray(arr)) {}\n` +
          `for (const v of arr) console.log(v);\n` +
          `if ('foo'.startsWith('f')) {}\n` +
          `const s = ['x','y'].map(item => item.toUpperCase());\n`,
        rules: ALL_RULES,
      },
      // Dense file — every rule that this fixture can trigger fires
      // somewhere, multiple sites apiece. The harness sort/diff has
      // to handle out-of-order matching cleanly.
      {
        filePath: 'dense.js',
        text:
          `const arr = [1,2,3];\n` +
          `arr.forEach(x => null);\n` +
          `const first = arr.filter(x => x > 0)[0];\n` +
          `if (arr.filter(x => x > 0).length > 0) {}\n` +
          `if (arr instanceof Array) {}\n` +
          `if ('xy'.indexOf('x') === 0) {}\n` +
          `const m = ['a','b'].map(String);\n` +
          `function f(p) { if (typeof p === "undefined") return null; }\n`,
        rules: ALL_RULES,
      },
      // ECMA-globals fixture: exercises `prefer-number-properties` and
      // `prefer-includes` together with `no-null` / `no-typeof-undefined`.
      // Locks in the three-bugs-in-one fix from this round:
      //   - parent-walk getScope so the indexOf/!==-1 BinaryExpression
      //     resolves at module scope
      //   - ECMA global seeding so ReferenceTracker finds parseInt / NaN
      //     in globalScope.variables
      //   - meta.defaultOptions overlay so checkNaN: true takes effect
      {
        filePath: 'ecma-globals.js',
        text:
          `const a = NaN;\n` +
          `const b = parseInt("10", 10);\n` +
          `if (arr.indexOf(2) !== -1) {}\n` +
          `function f(p) { if (typeof p === "undefined") return null; }\n`,
        rules: ALL_RULES,
      },
    ];

    const report = await runConformance({
      plugin: {
        prefix: 'unicorn',
        plugin: unicornPlugin as never,
        specifier: 'eslint-plugin-unicorn',
        ruleNames: [
          'no-null',
          'no-array-callback-reference',
          'prefer-array-some',
          'no-instanceof-array',
          'prefer-array-find',
          'prefer-string-starts-ends-with',
          'no-array-for-each',
          'no-typeof-undefined',
          'prefer-number-properties',
          'prefer-includes',
        ],
      },
      fixtures,
      resolverBaseUrl: baseUrl,
      workerCount: 1,
    });

    if (report.mismatched > 0) {
      throw new Error(`conformance mismatch:\n${formatReport(report)}`);
    }
    expect(report.mismatched).toBe(0);
    expect(report.matched).toBe(fixtures.length);
  });
});
