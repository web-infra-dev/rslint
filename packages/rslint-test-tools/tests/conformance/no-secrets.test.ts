import { describe, test, expect } from '@rstest/core';
import { pathToFileURL } from 'node:url';
import path from 'node:path';

import noSecretsPlugin from 'eslint-plugin-no-secrets';
import {
  runConformance,
  formatReport,
  type ConformanceFixture,
} from '../../src/eslint-conformance.js';

/**
 * Conformance harness: eslint-plugin-no-secrets.
 *
 * A second real plugin in the matrix (alongside eslint-plugin-unicorn).
 * Picked because:
 *   - small surface (one rule, no-secrets)
 *   - stateless (no scope analysis needed beyond default)
 *   - matches a category not covered by unicorn (security)
 *   - actively maintained, ESLint v10 compatible
 *
 * Adding this kind of plugin to the matrix follows the same operational
 * pattern as the unicorn ones: each plugin gets its own CI failure-
 * isolation boundary, with its own allow-list YAML if wording drifts
 * upstream.
 */

describe('eslint-plugin-no-secrets conformance', () => {
  test('eslint and rslint match on no-secrets fixtures', async () => {
    const baseUrl = pathToFileURL(
      path.resolve(__dirname, '..', '..', 'package.json'),
    ).href;

    const fixtures: ConformanceFixture[] = [
      {
        // Obvious-looking secret (long base64-ish string).
        filePath: 'secret.js',
        text: `const apiKey = "sk-abcdef0123456789abcdef0123456789abcdef0123456789";`,
        rules: {
          'sec/no-secrets': ['error', { tolerance: 4.0 }],
        },
      },
      {
        // Plain code, no secrets — must produce 0 diagnostics on both sides.
        filePath: 'clean.js',
        text: `const x = 42; export function add(a, b) { return a + b; }`,
        rules: {
          'sec/no-secrets': ['error', { tolerance: 4.0 }],
        },
      },
    ];

    const report = await runConformance({
      plugin: {
        prefix: 'sec',
        plugin: noSecretsPlugin as never,
        specifier: 'eslint-plugin-no-secrets',
        ruleNames: ['no-secrets'],
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
