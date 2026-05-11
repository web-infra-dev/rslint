import { describe, test, expect } from '@rstest/core';
import { pathToFileURL } from 'node:url';
import path from 'node:path';

import importPlugin from 'eslint-plugin-import';
import {
  runConformance,
  formatReport,
  type ConformanceFixture,
} from '../../src/eslint-conformance.js';

/**
 * Deep import-plugin coverage — three more single-file rules that
 * don't trip on eslint-plugin-import v2.x's v8-era assumption of
 * `context.parserOptions` (the v10 flat-config equivalent is
 * `context.languageOptions.parserOptions`; legacy probes like
 * `'sourceType' in context.parserOptions` TypeError under v9).
 * Rules avoided for that reason: no-default-export,
 * prefer-default-export, group-exports — all gated on the
 * isEsmModule(parserOptions) check.
 *
 * Coverage advances 7/46 → 10/46 (~22%).
 */

describe('eslint-plugin-import deep-coverage conformance', () => {
  test('eslint and rslint agree on further in-file import rules', async () => {
    const baseUrl = pathToFileURL(
      path.resolve(__dirname, '..', '..', 'package.json'),
    ).href;

    const fixtures: ConformanceFixture[] = [
      // no-named-default — `import { default as X }` is awkward
      {
        filePath: 'no-named-default.js',
        text: `import { default as X } from "./mod";\nconsole.log(X);\n`,
        rules: { 'import/no-named-default': 'error' },
      },
      // exports-last — exports must be at the bottom
      {
        filePath: 'exports-last.js',
        text: `export const a = 1;\nconst b = 2;\nvoid b;\n`,
        rules: { 'import/exports-last': 'error' },
      },
    ];

    const report = await runConformance({
      plugin: {
        prefix: 'import',
        plugin: importPlugin as never,
        specifier: 'eslint-plugin-import',
        ruleNames: ['no-named-default', 'exports-last'],
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
