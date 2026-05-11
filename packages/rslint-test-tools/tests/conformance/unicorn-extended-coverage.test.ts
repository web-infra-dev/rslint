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
 * Extended unicorn conformance — picks ten rules across distinct
 * categories beyond what `unicorn-multi-rule.test.ts` covers, to
 * widen the regression net over the worker JS plugin path. Each
 * category exercises a different rule shape:
 *
 *   - regex             → better-regex (regex literal walking)
 *   - string content    → no-hex-escape, escape-case, prefer-string-slice
 *   - error handling    → catch-error-name, throw-new-error
 *   - control flow      → prefer-ternary
 *   - numeric           → no-zero-fractions
 *   - scope analysis    → consistent-function-scoping
 *   - function calls    → prefer-spread
 *
 * Coverage moves from 10/149 → 20/149 (~13%). Still spot-check, but
 * spans more rule implementation patterns (visitor / scope / parent /
 * tokens) than the prior set.
 */

describe('unicorn extended-coverage conformance', () => {
  test('eslint and rslint agree across rule categories', async () => {
    const baseUrl = pathToFileURL(
      path.resolve(__dirname, '..', '..', 'package.json'),
    ).href;

    const fixtures: ConformanceFixture[] = [
      // regex — better-regex rewrites simple character classes.
      {
        filePath: 'better-regex.js',
        text: 'const r = /[0-9]+/g;',
        rules: { 'unicorn/better-regex': 'error' },
      },
      // string content — \\x hex escape should be \\u escape.
      {
        filePath: 'no-hex-escape.js',
        text: 'const s = "\\x42";',
        rules: { 'unicorn/no-hex-escape': 'error' },
      },
      // string content — escape sequences should use a consistent case.
      {
        filePath: 'escape-case.js',
        text: 'const s = "\\u00A9";',
        rules: { 'unicorn/escape-case': 'error' },
      },
      // string content — .substring/.substr → .slice
      {
        filePath: 'prefer-string-slice.js',
        text: 'const t = "hello".substring(0, 3);',
        rules: { 'unicorn/prefer-string-slice': 'error' },
      },
      // error handling — catch (e) → catch (error)
      {
        filePath: 'catch-error-name.js',
        text: 'try { throw 1 } catch (e) { console.log(e); }',
        rules: { 'unicorn/catch-error-name': 'error' },
      },
      // error handling — throw Error() → throw new Error()
      {
        filePath: 'throw-new-error.js',
        text: 'function f() { throw Error("boom"); }',
        rules: { 'unicorn/throw-new-error': 'error' },
      },
      // control flow — multi-branch if/else returning constants → ternary
      {
        filePath: 'prefer-ternary.js',
        text: 'function f(x) { if (x) { return 1; } else { return 2; } }',
        rules: { 'unicorn/prefer-ternary': 'error' },
      },
      // numeric — 1.0 → 1
      {
        filePath: 'no-zero-fractions.js',
        text: 'const a = 1.0; const b = 0.5;',
        rules: { 'unicorn/no-zero-fractions': 'error' },
      },
      // scope analysis — functions that don't reference enclosing scope
      // should be hoisted out.
      {
        filePath: 'consistent-function-scoping.js',
        text: 'function outer() {\n  function inner() { return 1; }\n  return inner();\n}',
        rules: { 'unicorn/consistent-function-scoping': 'error' },
      },
      // function calls — Array.from(args) / apply patterns → spread
      {
        filePath: 'prefer-spread.js',
        text: 'function f() { const a = Array.from(arguments); return a; }',
        rules: { 'unicorn/prefer-spread': 'error' },
      },
    ];

    const report = await runConformance({
      plugin: {
        prefix: 'unicorn',
        plugin: unicornPlugin as never,
        specifier: 'eslint-plugin-unicorn',
        ruleNames: [
          'better-regex',
          'no-hex-escape',
          'escape-case',
          'prefer-string-slice',
          'catch-error-name',
          'throw-new-error',
          'prefer-ternary',
          'no-zero-fractions',
          'consistent-function-scoping',
          'prefer-spread',
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
    // Vacuous-pass guard: if every fixture happens to produce 0
    // diagnostics on BOTH sides (e.g. dispatch silently no-ops),
    // mismatched==0 still passes and the coverage claim becomes
    // unfalsifiable. Require ≥1 ESLint diagnostic across the suite.
    expect(
      report.fixtureResults.reduce((n, r) => n + r.eslint.length, 0),
    ).toBeGreaterThan(0);
  });
});
