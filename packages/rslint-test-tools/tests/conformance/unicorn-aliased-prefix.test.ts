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
 * Conformance under a non-default plugin prefix. Re-exposes
 * `eslint-plugin-unicorn` as `'uni-alias/...'` so this file proves:
 *
 *   1. The runner's prefix → plugin instance routing honors whatever
 *      string the config supplied — it doesn't hard-code `'unicorn'`.
 *   2. Diagnostics on the rslint side come from the JS plugin's `create()`,
 *      NOT from any rslint-native rule that happens to share a fully-
 *      qualified name. (rslint ships native ports for a subset of
 *      unicorn rules — see `tests/eslint-plugin-unicorn/rules/`. Those
 *      ports are dispatched by Go-side name match, so the only way to
 *      guarantee we're exercising the JS plugin is to give it a prefix
 *      no native rule uses.)
 *   3. ESLint v10 produces identical diagnostics when wired the same
 *      way (`plugins: { 'uni-alias': unicornPlugin }`).
 *
 * If this passes but `unicorn-combined-rules.test.ts` regresses, it
 * means a native port has shadowed a JS plugin rule and the alias
 * indirection is masking the divergence.
 */

const ALIAS = 'uni-alias';

const ALL_RULES = {
  [`${ALIAS}/no-null`]: 'error',
  [`${ALIAS}/no-typeof-undefined`]: 'error',
  [`${ALIAS}/prefer-includes`]: 'error',
  [`${ALIAS}/prefer-number-properties`]: 'error',
  [`${ALIAS}/no-array-for-each`]: 'error',
} as const satisfies ConformanceFixture['rules'];

describe('eslint-plugin-unicorn under aliased prefix conformance', () => {
  test('eslint and rslint route `uni-alias/*` rules through the JS plugin', async () => {
    const baseUrl = pathToFileURL(
      path.resolve(__dirname, '..', '..', 'package.json'),
    ).href;

    const fixtures: ConformanceFixture[] = [
      // Mixes rules that route via different internal paths:
      //   - no-null:                listener
      //   - no-typeof-undefined:    isGlobalReference + scope walk
      //   - prefer-includes:        binary-expression analysis
      //   - prefer-number-properties: GlobalReferenceTracker + meta.defaultOptions
      //   - no-array-for-each:      method-chain selector + fixer
      {
        filePath: 'mixed.js',
        text:
          `const a = null;\n` +
          `function f(p) { if (typeof p === "undefined") return null; }\n` +
          `if (arr.indexOf(2) !== -1) {}\n` +
          `const x = NaN;\n` +
          `[1,2,3].forEach(n => null);\n`,
        rules: ALL_RULES,
      },
      // Clean — locked to verify no unexpected fires.
      {
        filePath: 'clean.js',
        text:
          `function g(p) { if (p === undefined) return; }\n` +
          `if (arr.includes(2)) {}\n` +
          `const x = Number.NaN;\n` +
          `for (const n of [1,2,3]) console.log(n);\n`,
        rules: ALL_RULES,
      },
    ];

    const report = await runConformance({
      plugin: {
        prefix: ALIAS,
        plugin: unicornPlugin as never,
        specifier: 'eslint-plugin-unicorn',
        ruleNames: [
          'no-null',
          'no-typeof-undefined',
          'prefer-includes',
          'prefer-number-properties',
          'no-array-for-each',
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
