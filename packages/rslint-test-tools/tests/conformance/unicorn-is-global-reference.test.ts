import { describe, test, expect } from '@rstest/core';
import { pathToFileURL } from 'node:url';
import path from 'node:path';

import unicornPlugin from 'eslint-plugin-unicorn';
import { runConformance, formatReport } from '../../src/eslint-conformance.js';

/**
 * Safety net for `SourceCode.isGlobalReference` — `eslint-plugin-unicorn`'s
 * `no-typeof-undefined` calls it internally to decide whether `typeof X`
 * should be flagged. Each fixture pins a specific branch of that decision
 * so any regression in rslint's scope-manager glue surfaces here.
 *
 * Behavior pinned against ESLint v10:
 *   - undeclared identifier  → global, no diag
 *   - `/* global foo *\/`     → declared global, no diag
 *   - `const foo = 1`        → local, diag
 *   - `function f(foo) {}`   → param, diag
 *   - `globalThis`           → built-in global, no diag
 */
describe('unicorn/no-typeof-undefined conformance (isGlobalReference)', () => {
  test('eslint and rslint agree across global/local/param identifiers', async () => {
    const baseUrl = pathToFileURL(
      path.resolve(__dirname, '..', '..', 'package.json'),
    ).href;

    const fixtures = [
      {
        filePath: 'undeclared.js',
        text: `if (typeof foo === "undefined") {}`,
        rules: { 'unicorn/no-typeof-undefined': 'error' as const },
      },
      {
        filePath: 'declared-global.js',
        text: `/* global foo */\nif (typeof foo === "undefined") {}`,
        rules: { 'unicorn/no-typeof-undefined': 'error' as const },
      },
      {
        filePath: 'local.js',
        text: `const foo = 1; if (typeof foo === "undefined") {}`,
        rules: { 'unicorn/no-typeof-undefined': 'error' as const },
      },
      {
        filePath: 'param.js',
        text: `function f(foo) { if (typeof foo === "undefined") {} }`,
        rules: { 'unicorn/no-typeof-undefined': 'error' as const },
      },
      {
        filePath: 'global-this.js',
        text: `if (typeof globalThis === "undefined") {}`,
        rules: { 'unicorn/no-typeof-undefined': 'error' as const },
      },
    ];

    const report = await runConformance({
      plugin: {
        prefix: 'unicorn',
        plugin: unicornPlugin as never,
        specifier: 'eslint-plugin-unicorn',
        ruleNames: ['no-typeof-undefined'],
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
