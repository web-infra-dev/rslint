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
 * Deep unicorn coverage — ten further rules across distinct shapes,
 * complementing `unicorn-multi-rule` and `unicorn-extended-coverage`.
 *
 * Coverage advances 20/149 → 30/149 (~20%). Areas widened:
 *   - DOM / fetch (no-document-cookie, no-invalid-fetch-options)
 *   - method calls (prefer-array-flat, prefer-array-index-of)
 *   - readability (no-nested-ternary, no-negated-condition, no-lonely-if)
 *   - immutability (no-array-push-push, no-array-reduce)
 *   - process (no-process-exit)
 */

describe('unicorn deep-coverage conformance', () => {
  test('eslint and rslint agree on a further set of unicorn rules', async () => {
    const baseUrl = pathToFileURL(
      path.resolve(__dirname, '..', '..', 'package.json'),
    ).href;

    const fixtures: ConformanceFixture[] = [
      // DOM — document.cookie write
      {
        filePath: 'no-document-cookie.js',
        text: `document.cookie = "x=1";`,
        rules: { 'unicorn/no-document-cookie': 'error' },
      },
      // fetch — invalid body for GET
      {
        filePath: 'no-invalid-fetch-options.js',
        text: `fetch("https://example.com", { method: "GET", body: "x" });`,
        rules: { 'unicorn/no-invalid-fetch-options': 'error' },
      },
      // method — Array.from(arr.flat())
      {
        filePath: 'prefer-array-flat.js',
        text: `const x = [].concat(...[[1],[2]]);`,
        rules: { 'unicorn/prefer-array-flat': 'error' },
      },
      // method — indexOf vs the legacy .lastIndexOf
      {
        filePath: 'prefer-array-index-of.js',
        text: `const i = [1,2,3].findIndex(x => x === 2);`,
        rules: { 'unicorn/prefer-array-index-of': 'error' },
      },
      // readability — deeply nested ternary
      {
        filePath: 'no-nested-ternary.js',
        text: `const x = a ? b ? c : d : e ? f : g;`,
        rules: { 'unicorn/no-nested-ternary': 'error' },
      },
      // readability — `if (!x) ... else ...` instead of `if (x) ... else ...`
      {
        filePath: 'no-negated-condition.js',
        text: `function f(x) { if (!x) { return 1; } else { return 2; } }`,
        rules: { 'unicorn/no-negated-condition': 'error' },
      },
      // readability — lonely if at the end of an else branch
      {
        filePath: 'no-lonely-if.js',
        text: `if (a) { x = 1; } else { if (b) { x = 2; } }`,
        rules: { 'unicorn/no-lonely-if': 'error' },
      },
      // method — consecutive push() calls
      {
        filePath: 'no-array-push-push.js',
        text: `const a = []; a.push(1); a.push(2);`,
        rules: { 'unicorn/no-array-push-push': 'error' },
      },
      // method — reduce should use loop
      {
        filePath: 'no-array-reduce.js',
        text: `const s = [1,2,3].reduce((acc, x) => acc + x, 0);`,
        rules: { 'unicorn/no-array-reduce': 'error' },
      },
      // process — process.exit() should not be called from libraries
      {
        filePath: 'no-process-exit.js',
        text: `function fail() { process.exit(1); }`,
        rules: { 'unicorn/no-process-exit': 'error' },
      },
    ];

    const report = await runConformance({
      plugin: {
        prefix: 'unicorn',
        plugin: unicornPlugin as never,
        specifier: 'eslint-plugin-unicorn',
        ruleNames: [
          'no-document-cookie',
          'no-invalid-fetch-options',
          'prefer-array-flat',
          'prefer-array-index-of',
          'no-nested-ternary',
          'no-negated-condition',
          'no-lonely-if',
          'no-array-push-push',
          'no-array-reduce',
          'no-process-exit',
        ],
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
