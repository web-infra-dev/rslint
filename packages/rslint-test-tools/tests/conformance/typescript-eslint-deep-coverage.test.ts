import { describe, test, expect } from '@rstest/core';
import { pathToFileURL } from 'node:url';
import path from 'node:path';

import tsPlugin from '@typescript-eslint/eslint-plugin';
import {
  runConformance,
  formatReport,
  type ConformanceFixture,
} from '../../src/eslint-conformance.js';

/**
 * Deep @typescript-eslint coverage — ten more in-file syntax rules,
 * different categories from `typescript-eslint-coverage`:
 *
 *   - this binding (no-this-alias)
 *   - constructor patterns (no-misused-new, no-useless-constructor)
 *   - enum / interface shape (no-empty-interface, no-duplicate-enum-values)
 *   - type-alias style (consistent-indexed-object-style)
 *   - parameter ordering (default-param-last)
 *   - prefer (prefer-namespace-keyword, prefer-for-of)
 *   - module syntax (no-require-imports)
 *
 * Coverage advances 10/134 → 20/134 (~15%).
 */

describe('@typescript-eslint deep-coverage conformance', () => {
  test('eslint and rslint agree on a further TS rule set', async () => {
    const baseUrl = pathToFileURL(
      path.resolve(__dirname, '..', '..', 'package.json'),
    ).href;

    const fixtures: ConformanceFixture[] = [
      {
        filePath: 'no-this-alias.ts',
        text: `class C {\n  do() {\n    const self = this;\n    void self;\n  }\n}\n`,
        rules: { '@typescript-eslint/no-this-alias': 'error' },
        isTypeScript: true,
      },
      {
        filePath: 'no-misused-new.ts',
        text: `interface I {\n  new(): I;\n}\n`,
        rules: { '@typescript-eslint/no-misused-new': 'error' },
        isTypeScript: true,
      },
      {
        filePath: 'no-useless-constructor.ts',
        text: `class A {\n  constructor() {}\n}\n`,
        rules: { '@typescript-eslint/no-useless-constructor': 'error' },
        isTypeScript: true,
      },
      {
        filePath: 'no-empty-interface.ts',
        text: `interface I {}\n`,
        rules: { '@typescript-eslint/no-empty-interface': 'error' },
        isTypeScript: true,
      },
      {
        filePath: 'no-duplicate-enum-values.ts',
        text: `enum E {\n  A = 1,\n  B = 1,\n}\n`,
        rules: { '@typescript-eslint/no-duplicate-enum-values': 'error' },
        isTypeScript: true,
      },
      {
        filePath: 'consistent-indexed-object-style.ts',
        text: `type Foo = { [key: string]: number };\n`,
        rules: {
          '@typescript-eslint/consistent-indexed-object-style': 'error',
        },
        isTypeScript: true,
      },
      {
        filePath: 'default-param-last.ts',
        text: `function f(a = 1, b: number) {\n  return a + b;\n}\n`,
        rules: { '@typescript-eslint/default-param-last': 'error' },
        isTypeScript: true,
      },
      {
        filePath: 'prefer-namespace-keyword.ts',
        text: `module M {\n  export const x = 1;\n}\n`,
        rules: { '@typescript-eslint/prefer-namespace-keyword': 'error' },
        isTypeScript: true,
      },
      {
        filePath: 'prefer-for-of.ts',
        text: `const a = [1, 2, 3];\nfor (let i = 0; i < a.length; i++) {\n  void a[i];\n}\n`,
        rules: { '@typescript-eslint/prefer-for-of': 'error' },
        isTypeScript: true,
      },
      {
        filePath: 'no-require-imports.ts',
        text: `const fs = require("fs");\nvoid fs;\n`,
        rules: { '@typescript-eslint/no-require-imports': 'error' },
        isTypeScript: true,
      },
    ];

    const report = await runConformance({
      plugin: {
        prefix: '@typescript-eslint',
        plugin: tsPlugin as never,
        specifier: '@typescript-eslint/eslint-plugin',
        ruleNames: [
          'no-this-alias',
          'no-misused-new',
          'no-useless-constructor',
          'no-empty-interface',
          'no-duplicate-enum-values',
          'consistent-indexed-object-style',
          'default-param-last',
          'prefer-namespace-keyword',
          'prefer-for-of',
          'no-require-imports',
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
