import { describe, test, expect } from '@rstest/core';
import { pathToFileURL } from 'node:url';
import path from 'node:path';

import unicornPlugin from 'eslint-plugin-unicorn';
import { runConformance, formatReport } from '../../src/eslint-conformance.js';

/**
 * Conformance harness exercised against `eslint-plugin-unicorn`.
 *
 * Runs the same fixture through both ESLint v10 and rslint, then
 * structurally diffs the diagnostics. A passing test means rslint's
 * compat layer produces output indistinguishable from ESLint for this
 * rule on these fixtures.
 *
 * If a future plugin upgrade changes wording or shifts diagnostic
 * positions, this test fails with a per-line diff identifying exactly
 * what drifted.
 */

describe('unicorn/no-null conformance', () => {
  test('eslint and rslint produce equivalent diagnostics across fixtures', async () => {
    const baseUrl = pathToFileURL(
      path.resolve(__dirname, '..', '..', 'package.json'),
    ).href;

    const fixtures = [
      {
        filePath: 'a.js',
        text: `const x = null;`,
        rules: { 'unicorn/no-null': 'error' as const },
      },
      {
        filePath: 'b.js',
        text: `const v = 42;`,
        rules: { 'unicorn/no-null': 'error' as const },
      },
      {
        filePath: 'c.js',
        text: `function f() { return null; }`,
        rules: { 'unicorn/no-null': 'error' as const },
      },
      {
        filePath: 'd.js',
        text: `if (x === null) {} const y = null;`,
        rules: { 'unicorn/no-null': 'error' as const },
      },
    ];

    const report = await runConformance({
      plugin: {
        prefix: 'unicorn',
        plugin: unicornPlugin as never,
        specifier: 'eslint-plugin-unicorn',
        ruleNames: ['no-null'],
      },
      fixtures,
      resolverBaseUrl: baseUrl,
      workerCount: 1, // deterministic ordering
    });

    if (report.mismatched > 0) {
      // Surface the diff in the test failure
      throw new Error(`conformance mismatch:\n${formatReport(report)}`);
    }
    expect(report.mismatched).toBe(0);
    expect(report.matched).toBe(fixtures.length);
  });
});
