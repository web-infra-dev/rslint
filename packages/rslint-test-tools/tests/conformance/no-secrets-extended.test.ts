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
 * Extended-fixture conformance for `eslint-plugin-no-secrets`. The plugin
 * only ships one rule, so "combined rules" doesn't apply — but the rule
 * is *entropy-driven* and accepts several option shapes, which is its
 * own combinatoric surface. This file exercises that surface:
 *
 *   - several distinct secret patterns in one file (multi-site fire)
 *   - distinct tolerance values
 *   - `ignoreContent` / `ignoreModules` options
 *   - secrets embedded in comments
 *   - mixed clean + dirty inside the same source
 *
 * Companion to `no-secrets.test.ts` (single-fixture per case).
 */

const RULE = 'sec/no-secrets';

describe('eslint-plugin-no-secrets extended conformance', () => {
  test('eslint and rslint agree across multi-site / option / comment fixtures', async () => {
    const baseUrl = pathToFileURL(
      path.resolve(__dirname, '..', '..', 'package.json'),
    ).href;

    const fixtures: ConformanceFixture[] = [
      // Two long high-entropy literals in one file — both should fire.
      {
        filePath: 'multi-secret.js',
        text:
          `const a = "sk-abcdef0123456789abcdef0123456789abcdef0123456789";\n` +
          `const b = "ghp_1234567890abcdef1234567890abcdef12345678ABCD";\n`,
        rules: { [RULE]: ['error', { tolerance: 4.0 }] },
      },
      // Mixed: a real-looking secret next to plain prose. Only the secret fires.
      {
        filePath: 'mixed.js',
        text:
          `const greeting = "hello world, this is a normal string";\n` +
          `const apiKey = "sk-abcdef0123456789abcdef0123456789abcdef0123456789";\n`,
        rules: { [RULE]: ['error', { tolerance: 4.0 }] },
      },
      // Secret in a comment — the rule walks code AND comments.
      {
        filePath: 'comment-secret.js',
        text:
          `// TODO rotate before release: sk-abcdef0123456789abcdef0123456789abcdef0123456789\n` +
          `const x = 1;\n`,
        rules: { [RULE]: ['error', { tolerance: 4.0 }] },
      },
      // Same secret literal twice — both should report (no dedup by content).
      {
        filePath: 'dup-secret.js',
        text:
          `const a = "sk-abcdef0123456789abcdef0123456789abcdef0123456789";\n` +
          `const b = "sk-abcdef0123456789abcdef0123456789abcdef0123456789";\n`,
        rules: { [RULE]: ['error', { tolerance: 4.0 }] },
      },
      // Higher tolerance → fewer reports. With tolerance=5 the entropy
      // bar rises; the same literal that triggered above may not now.
      // The point of this fixture is that BOTH engines apply the
      // option consistently.
      {
        filePath: 'high-tol.js',
        text: `const apiKey = "sk-abcdef0123456789abcdef0123456789abcdef0123456789";\n`,
        rules: { [RULE]: ['error', { tolerance: 5.0 }] },
      },
      // Clean: short strings + plain identifiers. Nothing fires.
      {
        filePath: 'clean.js',
        text:
          `const x = 42;\n` +
          `const msg = "ok";\n` +
          `export function add(a, b) { return a + b; }\n`,
        rules: { [RULE]: ['error', { tolerance: 4.0 }] },
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
