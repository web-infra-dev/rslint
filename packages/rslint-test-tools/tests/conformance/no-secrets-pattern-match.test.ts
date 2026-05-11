import { describe, test, expect } from '@rstest/core';
import { pathToFileURL } from 'node:url';
import path from 'node:path';

import noSecretsPlugin from 'eslint-plugin-no-secrets';
import { runConformance, formatReport } from '../../src/eslint-conformance.js';

/**
 * Conformance: eslint-plugin-no-secrets / no-pattern-match. Together
 * with `no-secrets.test.ts` (which covers `sec/no-secrets`) this brings
 * the plugin to 100% rule coverage in the worker JS plugin path —
 * trivially small, but a useful regression net since no-secrets is the
 * only 2-rule plugin we exercise end-to-end.
 */

describe('no-secrets/no-pattern-match conformance', () => {
  test('eslint and rslint agree on no-pattern-match diagnostics', async () => {
    const baseUrl = pathToFileURL(
      path.resolve(__dirname, '..', '..', 'package.json'),
    ).href;

    const fixtures = [
      // The default config of no-pattern-match flags any string literal
      // that matches a known credentials regex. Below: a JWT-shaped
      // token (3 base64url parts separated by `.`). The rule's default
      // pattern set includes a JWT detector.
      {
        filePath: 'jwt.js',
        text: 'const t = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4ifQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c";',
        rules: { 'sec/no-pattern-match': 'error' as const },
      },
      // Plain string — should not fire.
      {
        filePath: 'plain.js',
        text: 'const s = "hello world";',
        rules: { 'sec/no-pattern-match': 'error' as const },
      },
    ];

    const report = await runConformance({
      plugin: {
        prefix: 'sec',
        plugin: noSecretsPlugin as never,
        specifier: 'eslint-plugin-no-secrets',
        ruleNames: ['no-pattern-match'],
      },
      fixtures,
      resolverBaseUrl: baseUrl,
      workerCount: 1,
    });

    if (report.mismatched > 0) {
      throw new Error('Conformance mismatch:\n' + formatReport(report));
    }
    expect(report.mismatched).toBe(0);
  });
});
