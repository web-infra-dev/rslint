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
 * Bulk @typescript-eslint coverage — pushes the matrix to ~37/134
 * after adding 17 further syntax-only rules. Type-info-dependent
 * rules (no-floating-promises, no-unnecessary-condition, etc.) are
 * deferred — the harness does not wire `parserOptions.project`, so
 * those would silently produce 0 on both sides (vacuously aligned but
 * not meaningfully exercised). When `project` support lands they can
 * graduate from "Deferred" to here.
 */

describe('@typescript-eslint bulk-coverage conformance', () => {
  test('eslint and rslint agree on additional syntax-only TS rules', async () => {
    const baseUrl = pathToFileURL(
      path.resolve(__dirname, '..', '..', 'package.json'),
    ).href;

    const fixtures: ConformanceFixture[] = [
      // class members style
      {
        filePath: 'adjacent-overload-signatures.ts',
        text: `interface I {\n  foo(a: number): void;\n  bar(): void;\n  foo(a: string): void;\n}\n`,
        rules: { '@typescript-eslint/adjacent-overload-signatures': 'error' },
        isTypeScript: true,
      },
      {
        filePath: 'array-type.ts',
        text: `const a: Array<number> = [];\nvoid a;\n`,
        rules: { '@typescript-eslint/array-type': 'error' },
        isTypeScript: true,
      },
      {
        filePath: 'ban-tslint-comment.ts',
        text: `// tslint:disable-next-line\nconst x: number = 1;\nvoid x;\n`,
        rules: { '@typescript-eslint/ban-tslint-comment': 'error' },
        isTypeScript: true,
      },
      {
        filePath: 'class-literal-property-style.ts',
        text: `class C { get x() { return 1; } }\n`,
        rules: { '@typescript-eslint/class-literal-property-style': 'error' },
        isTypeScript: true,
      },
      // constructor / inferrable / no-empty-function
      {
        filePath: 'no-inferrable-types.ts',
        text: `const a: number = 1;\nvoid a;\n`,
        rules: { '@typescript-eslint/no-inferrable-types': 'error' },
        isTypeScript: true,
      },
      {
        filePath: 'no-empty-function.ts',
        text: `function f() {}\nvoid f;\n`,
        rules: { '@typescript-eslint/no-empty-function': 'error' },
        isTypeScript: true,
      },
      {
        filePath: 'no-array-constructor.ts',
        text: `const a = new Array(1, 2);\nvoid a;\n`,
        rules: { '@typescript-eslint/no-array-constructor': 'error' },
        isTypeScript: true,
      },
      {
        filePath: 'no-extraneous-class.ts',
        text: `class Util { static helper() {} }\nvoid Util;\n`,
        rules: { '@typescript-eslint/no-extraneous-class': 'error' },
        isTypeScript: true,
      },
      // assertion / type style
      {
        filePath: 'no-non-null-assertion.ts',
        text: `declare const a: number | null;\nconst b = a!;\nvoid b;\n`,
        rules: { '@typescript-eslint/no-non-null-assertion': 'error' },
        isTypeScript: true,
      },
      {
        filePath: 'no-non-null-asserted-nullish-coalescing.ts',
        text: `declare const a: number | null;\nconst b = a! ?? 0;\nvoid b;\n`,
        rules: {
          '@typescript-eslint/no-non-null-asserted-nullish-coalescing': 'error',
        },
        isTypeScript: true,
      },
      // members / control flow
      {
        filePath: 'no-loop-func.ts',
        text: `for (let i = 0; i < 3; i++) {\n  setTimeout(function () {\n    return i;\n  }, 0);\n}\n`,
        rules: { '@typescript-eslint/no-loop-func': 'error' },
        isTypeScript: true,
      },
      {
        filePath: 'no-dupe-class-members.ts',
        text: `class A {\n  foo() {}\n  foo() {}\n}\n`,
        rules: { '@typescript-eslint/no-dupe-class-members': 'error' },
        isTypeScript: true,
      },
      // unused / before-define
      {
        filePath: 'no-unused-expressions.ts',
        text: `1 + 2;\n`,
        rules: { '@typescript-eslint/no-unused-expressions': 'error' },
        isTypeScript: true,
      },
      {
        filePath: 'no-useless-empty-export.ts',
        text: `export const x = 1;\nexport {};\n`,
        rules: { '@typescript-eslint/no-useless-empty-export': 'error' },
        isTypeScript: true,
      },
      {
        filePath: 'prefer-literal-enum-member.ts',
        text: `enum E {\n  A = 1 + 1,\n}\nvoid E;\n`,
        rules: { '@typescript-eslint/prefer-literal-enum-member': 'error' },
        isTypeScript: true,
      },
      // functions
      {
        filePath: 'prefer-function-type.ts',
        text: `interface I {\n  (a: number): boolean;\n}\nconst f: I = () => true;\nvoid f;\n`,
        rules: { '@typescript-eslint/prefer-function-type': 'error' },
        isTypeScript: true,
      },
      // ts directive
      {
        filePath: 'prefer-ts-expect-error.ts',
        text: `// @ts-ignore\nconst x: number = 1;\nvoid x;\n`,
        rules: { '@typescript-eslint/prefer-ts-expect-error': 'error' },
        isTypeScript: true,
      },
    ];

    const ruleNames = Array.from(
      new Set(fixtures.flatMap((f) => Object.keys(f.rules))),
    )
      .map((r) => r.replace(/^@typescript-eslint\//, ''))
      .sort();

    const report = await runConformance({
      plugin: {
        prefix: '@typescript-eslint',
        plugin: tsPlugin as never,
        specifier: '@typescript-eslint/eslint-plugin',
        ruleNames,
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
