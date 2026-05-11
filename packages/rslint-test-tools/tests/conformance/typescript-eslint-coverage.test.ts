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
 * Conformance: @typescript-eslint/eslint-plugin. Exercises ten in-file
 * syntax-level rules (no type-info needed). Both sides parse TS:
 *
 *   - ESLint v10 side: harness injects `@typescript-eslint/parser` via
 *     `languageOptions.parser` when `isTypeScript: true`.
 *   - rslint side: worker uses oxc-parser, which infers TS lang from
 *     the `.ts` file extension automatically.
 *
 * Coverage: 10/134 (~7.5%). Rule categories spanned:
 *   - type annotations (no-explicit-any, no-empty-object-type)
 *   - module syntax (no-namespace, triple-slash-reference)
 *   - assertion ergonomics (prefer-as-const, no-confusing-non-null-assertion,
 *     no-extra-non-null-assertion, no-non-null-asserted-optional-chain)
 *   - directive comments (ban-ts-comment)
 *   - interface/type alias style (consistent-type-definitions)
 */

describe('@typescript-eslint conformance', () => {
  test('eslint and rslint agree on a representative TS rule set', async () => {
    const baseUrl = pathToFileURL(
      path.resolve(__dirname, '..', '..', 'package.json'),
    ).href;

    const fixtures: ConformanceFixture[] = [
      {
        filePath: 'no-explicit-any.ts',
        text: 'function f(x: any): any { return x; }\n',
        rules: { '@typescript-eslint/no-explicit-any': 'error' },
        isTypeScript: true,
      },
      {
        filePath: 'no-empty-object-type.ts',
        text: 'type X = {};\ninterface Y {}\n',
        rules: { '@typescript-eslint/no-empty-object-type': 'error' },
        isTypeScript: true,
      },
      {
        filePath: 'no-namespace.ts',
        text: 'namespace N { export const x = 1; }\n',
        rules: { '@typescript-eslint/no-namespace': 'error' },
        isTypeScript: true,
      },
      {
        filePath: 'triple-slash.ts',
        text: '/// <reference path="other.d.ts" />\nexport {};\n',
        rules: { '@typescript-eslint/triple-slash-reference': 'error' },
        isTypeScript: true,
      },
      {
        filePath: 'prefer-as-const.ts',
        text: `const a = 1 as 1;\n`,
        rules: { '@typescript-eslint/prefer-as-const': 'error' },
        isTypeScript: true,
      },
      {
        filePath: 'confusing-nn.ts',
        text: `declare const a: number | null;\nconst b = a! == 0;\n`,
        rules: {
          '@typescript-eslint/no-confusing-non-null-assertion': 'error',
        },
        isTypeScript: true,
      },
      {
        filePath: 'extra-nn.ts',
        text: `declare const a: number;\nconst b = a!!;\n`,
        rules: {
          '@typescript-eslint/no-extra-non-null-assertion': 'error',
        },
        isTypeScript: true,
      },
      {
        filePath: 'optional-chain-nn.ts',
        text: `declare const obj: { a?: number };\nconst v = obj?.a!;\n`,
        rules: {
          '@typescript-eslint/no-non-null-asserted-optional-chain': 'error',
        },
        isTypeScript: true,
      },
      {
        filePath: 'ban-ts-comment.ts',
        text: `// @ts-ignore\nconst x: number = 1;\n`,
        rules: { '@typescript-eslint/ban-ts-comment': 'error' },
        isTypeScript: true,
      },
      {
        filePath: 'consistent-type-def.ts',
        text: `type Foo = { a: number };\n`,
        rules: {
          '@typescript-eslint/consistent-type-definitions': 'error',
        },
        isTypeScript: true,
      },
    ];

    const report = await runConformance({
      plugin: {
        prefix: '@typescript-eslint',
        plugin: tsPlugin as never,
        specifier: '@typescript-eslint/eslint-plugin',
        ruleNames: [
          'no-explicit-any',
          'no-empty-object-type',
          'no-namespace',
          'triple-slash-reference',
          'prefer-as-const',
          'no-confusing-non-null-assertion',
          'no-extra-non-null-assertion',
          'no-non-null-asserted-optional-chain',
          'ban-ts-comment',
          'consistent-type-definitions',
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
