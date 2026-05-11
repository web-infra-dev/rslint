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
 * Second bulk @typescript-eslint coverage round — further syntax-only
 * rules across method/parameter/enum/import shapes. Type-info rules
 * (no-floating-promises, no-misused-promises, restrict-*, etc.) are
 * still skipped — they need `parserOptions.project` wiring that the
 * harness doesn't have.
 */

describe('@typescript-eslint bulk2-coverage conformance', () => {
  test('eslint and rslint agree on a second batch of TS rules', async () => {
    const baseUrl = pathToFileURL(
      path.resolve(__dirname, '..', '..', 'package.json'),
    ).href;

    const fixtures: ConformanceFixture[] = [
      {
        filePath: 'class-methods-use-this.ts',
        text: `class C { foo() { return 1; } }\nvoid C;`,
        rules: { '@typescript-eslint/class-methods-use-this': 'error' },
        isTypeScript: true,
      },
      {
        filePath: 'consistent-generic-constructors.ts',
        text: `const m: Map<string, number> = new Map<string, number>();\nvoid m;`,
        rules: {
          '@typescript-eslint/consistent-generic-constructors': 'error',
        },
        isTypeScript: true,
      },
      {
        filePath: 'consistent-type-assertions.ts',
        text: `const a = <number>1;\nvoid a;`,
        rules: { '@typescript-eslint/consistent-type-assertions': 'error' },
        isTypeScript: true,
      },
      {
        filePath: 'explicit-function-return-type.ts',
        text: `function f(x: number) { return x + 1; }\nvoid f;`,
        rules: { '@typescript-eslint/explicit-function-return-type': 'error' },
        isTypeScript: true,
      },
      {
        filePath: 'explicit-member-accessibility.ts',
        text: `class C { x = 1; }\nvoid C;`,
        rules: { '@typescript-eslint/explicit-member-accessibility': 'error' },
        isTypeScript: true,
      },
      {
        filePath: 'explicit-module-boundary-types.ts',
        text: `export function f(x) { return x; }\nvoid f;`,
        rules: { '@typescript-eslint/explicit-module-boundary-types': 'error' },
        isTypeScript: true,
      },
      {
        filePath: 'init-declarations.ts',
        text: `let x: number;\nvoid x;`,
        rules: { '@typescript-eslint/init-declarations': 'error' },
        isTypeScript: true,
      },
      {
        filePath: 'max-params.ts',
        text: `function f(a: number, b: number, c: number, d: number, e: number) {\n  return a + b + c + d + e;\n}\nvoid f;`,
        rules: { '@typescript-eslint/max-params': 'error' },
        isTypeScript: true,
      },
      {
        filePath: 'member-ordering.ts',
        text: `class C {\n  foo() {}\n  x = 1;\n}\nvoid C;`,
        rules: { '@typescript-eslint/member-ordering': 'error' },
        isTypeScript: true,
      },
      {
        filePath: 'method-signature-style.ts',
        text: `interface I {\n  foo(): void;\n}\nconst x: I = { foo() {} };\nvoid x;`,
        rules: { '@typescript-eslint/method-signature-style': 'error' },
        isTypeScript: true,
      },
      {
        filePath: 'no-dynamic-delete.ts',
        text: `const o: Record<string, number> = { x: 1 };\nconst k: string = 'x';\ndelete o[k];`,
        rules: { '@typescript-eslint/no-dynamic-delete': 'error' },
        isTypeScript: true,
      },
      {
        filePath: 'no-invalid-void-type.ts',
        text: `type T = void | number;\nconst v: T = 1;\nvoid v;`,
        rules: { '@typescript-eslint/no-invalid-void-type': 'error' },
        isTypeScript: true,
      },
      {
        filePath: 'no-loss-of-precision.ts',
        text: `const a = 9007199254740993;\nvoid a;`,
        rules: { '@typescript-eslint/no-loss-of-precision': 'error' },
        isTypeScript: true,
      },
      {
        filePath: 'no-magic-numbers.ts',
        text: `function f() { return 42 * 7; }\nvoid f;`,
        rules: { '@typescript-eslint/no-magic-numbers': 'error' },
        isTypeScript: true,
      },
      {
        filePath: 'no-redeclare.ts',
        text: `let a = 1;\nlet a = 2;\nvoid a;`,
        rules: { '@typescript-eslint/no-redeclare': 'error' },
        isTypeScript: true,
      },
      {
        filePath: 'no-restricted-types.ts',
        text: `function f(x: String) { return x; }\nvoid f;`,
        rules: {
          '@typescript-eslint/no-restricted-types': [
            'error',
            { types: { String: 'use string' } },
          ],
        },
        isTypeScript: true,
      },
      {
        filePath: 'no-shadow.ts',
        text: `let x = 1;\nfunction f() {\n  let x = 2;\n  return x;\n}\nvoid x; void f;`,
        rules: { '@typescript-eslint/no-shadow': 'error' },
        isTypeScript: true,
      },
      {
        filePath: 'no-unnecessary-type-constraint.ts',
        text: `function f<T extends unknown>(x: T) { return x; }\nvoid f;`,
        rules: { '@typescript-eslint/no-unnecessary-type-constraint': 'error' },
        isTypeScript: true,
      },
      {
        filePath: 'no-use-before-define.ts',
        text: `const v = f();\nfunction f() { return 1; }\nvoid v;`,
        rules: { '@typescript-eslint/no-use-before-define': 'error' },
        isTypeScript: true,
      },
      {
        filePath: 'no-wrapper-object-types.ts',
        text: `function f(x: Number) { return x; }\nvoid f;`,
        rules: { '@typescript-eslint/no-wrapper-object-types': 'error' },
        isTypeScript: true,
      },
      {
        filePath: 'parameter-properties.ts',
        text: `class C { constructor(public x: number) {} }\nvoid C;`,
        rules: { '@typescript-eslint/parameter-properties': 'error' },
        isTypeScript: true,
      },
      {
        filePath: 'prefer-enum-initializers.ts',
        text: `enum E { A, B }\nvoid E;`,
        rules: { '@typescript-eslint/prefer-enum-initializers': 'error' },
        isTypeScript: true,
      },
      {
        filePath: 'unified-signatures.ts',
        text: `interface I {\n  foo(x: string): void;\n  foo(x: number): void;\n}\nconst v: I = { foo(_x) {} };\nvoid v;`,
        rules: { '@typescript-eslint/unified-signatures': 'error' },
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
