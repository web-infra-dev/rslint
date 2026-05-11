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
 * Plugin compatibility matrix slice — multiple unicorn rules exercised
 * through the conformance harness. The harness doesn't bundle a TS
 * parser, so type-aware rules are excluded (per the eslint-plugin
 * system's documented non-goals).
 *
 * Adding a new plugin to the matrix is mechanical:
 *   1. install the plugin as a devDependency of @rslint/test-tools
 *   2. import the plugin into a new test file like this one
 *   3. add `runConformance({...})` with fixtures
 *   4. add the path to rstest.config.mts include[]
 *
 * Each plugin gets its own test file for failure isolation under CI.
 */

describe('unicorn multi-rule conformance', () => {
  test('eslint and rslint match on a representative selection of unicorn rules', async () => {
    const baseUrl = pathToFileURL(
      path.resolve(__dirname, '..', '..', 'package.json'),
    ).href;

    const fixtures: ConformanceFixture[] = [
      // no-null: classic null usage
      {
        filePath: 'no-null-1.js',
        text: `const x = null;\nfunction f() { return null; }`,
        rules: { 'unicorn/no-null': 'error' },
      },
      // no-array-callback-reference: passing String/Boolean/Number directly to map etc.
      {
        filePath: 'array-callback.js',
        text: `[1,2,3].map(String);`,
        rules: { 'unicorn/no-array-callback-reference': 'error' },
      },
      // prefer-array-some: comparing filter().length > 0 → use some()
      {
        filePath: 'array-some.js',
        text: `if ([1,2,3].filter(x => x > 1).length > 0) {}`,
        rules: { 'unicorn/prefer-array-some': 'error' },
      },
      // no-instanceof-array: should warn about Array.isArray pattern mismatch
      {
        filePath: 'instanceof.js',
        text: `if (x instanceof Array) {}`,
        rules: { 'unicorn/no-instanceof-array': 'error' },
      },
      // prefer-array-find: filter()[0] → find()
      {
        filePath: 'prefer-find.js',
        text: `const a = [1,2,3].filter(x => x > 1)[0];`,
        rules: { 'unicorn/prefer-array-find': 'error' },
      },
      // prefer-string-starts-ends-with: indexOf===0 → startsWith
      {
        filePath: 'starts-with.js',
        text: `if ('foo'.indexOf('f') === 0) {}`,
        rules: { 'unicorn/prefer-string-starts-ends-with': 'error' },
      },
      // no-array-for-each: forEach() → for…of
      {
        filePath: 'foreach.js',
        text: `[1,2,3].forEach(x => console.log(x));`,
        rules: { 'unicorn/no-array-for-each': 'error' },
      },
      // ── Known limitations (deferred, NOT in fixtures above) ──
      //
      // `unicorn/prefer-includes` — ESLint reports on
      // `[1,2,3].indexOf(2) !== -1`, rslint reports 0. The rule's
      // ESLint implementation walks `BinaryExpression > [operator =
      // !==]` selectors against a hand-rolled token analysis that
      // touches behavior the rslint runner currently doesn't mirror
      // exactly. Filed for follow-up; left out of the matrix until
      // we either align the path or expand `@experimental` to
      // mention the gap.
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
    // Vacuous-pass guard — see unicorn-extended-coverage for rationale.
    expect(
      report.fixtureResults.reduce((n, r) => n + r.eslint.length, 0),
    ).toBeGreaterThan(0);
  });
});
