// Forked and modified from https://github.com/typescript-eslint/typescript-eslint/blob/16c344ec7d274ea542157e0f19682dd1930ab838/packages/rule-tester/src/RuleTester.ts#L4

import path from 'node:path';
import test from 'node:test';
import util from 'node:util';
import { lint, type Diagnostic } from '@rslint/core';

import assert from 'node:assert';

interface TsDiagnostic {
  line: number;
  column: number;
  endLine: number;
  endColumn: number;
  messageId: string;
  suggestions: any[];
}
function toCamelCase(name: string): string {
  return name.replace(/-([a-z])/g, g => g[1].toUpperCase());
}
// check whether rslint diagnostics and typescript-eslint diagnostics are semantic equal
function checkDiagnosticEqual(
  rslintDiagnostic: Diagnostic[],
  tsDiagnostic: TsDiagnostic[],
) {
  assert(
    rslintDiagnostic.length === tsDiagnostic.length,
    `Length mismatch: ${rslintDiagnostic.length} !== ${tsDiagnostic.length}`,
  );
  for (let i = 0; i < rslintDiagnostic.length; i++) {
    const rslintDiag = rslintDiagnostic[i];
    const tsDiag = tsDiagnostic[i];
    // check rule match
    assert(
      toCamelCase(rslintDiag.ruleName) === tsDiag.messageId,
      `Message mismatch: ${rslintDiag.ruleName} !== ${tsDiag.messageId}`,
    );

    // check range match
    // tsDiag sometimes doesn't have line and column, so we need to check that
    if (tsDiag.line) {
      assert(
        rslintDiag.range.start.line === tsDiag.line,
        `Start line mismatch: ${rslintDiag.range.start.line} !== ${tsDiag.line}`,
      );
    }
    if (tsDiag.endLine) {
      assert(
        rslintDiag.range.end.line === tsDiag.endLine,
        `End line mismatch: ${rslintDiag.range.end.line} !== ${tsDiag.endLine}`,
      );
    }
    if (tsDiag.column) {
      assert(
        rslintDiag.range.start.column === tsDiag.column,
        `Start column mismatch: ${rslintDiag.range.start.column} !== ${tsDiag.column}`,
      );
    }
    if (tsDiag.endColumn) {
      assert(
        rslintDiag.range.end.column === tsDiag.endColumn,
        `End column mismatch: ${rslintDiag.range.end.column} !== ${tsDiag.endColumn}`,
      );
    }
  }
}

export class RuleTester {
  constructor(options: any) {}
  public run(
    ruleName: string,
    cases: {
      valid: string[];
      invalid: {
        code: string;
        errors: any[];
      }[];
    },
  ) {
    test(ruleName, async () => {
      let cwd = path.resolve(import.meta.dirname, './fixtures');
      const tsconfig = path.resolve(
        import.meta.dirname,
        './fixtures/tsconfig.virtual.json',
      );
      let virtual_entry = path.resolve(cwd, 'src/virtual.ts');
      await test('valid', async () => {
        for (const code of cases.valid) {
          const diags = await lint({
            tsconfig,
            workingDirectory: cwd,
            fileContents: {
              [virtual_entry]: code,
            },
            ruleOptions: {
              [ruleName]: 'error',
            },
          });
          assert(
            diags.diagnostics?.length === 0,
            `Expected no diagnostics for valid case, but got: ${JSON.stringify(diags)}`,
          );
        }
      });
      await test('invalid', async t => {
        for (const { errors, code } of cases.invalid) {
          const diags = await lint({
            tsconfig,
            workingDirectory: cwd,
            fileContents: {
              [virtual_entry]: code,
            },
            ruleOptions: {
              [ruleName]: 'error',
            },
          });
          t.assert.snapshot(diags);
          assert(
            diags.diagnostics?.length > 0,
            `Expected diagnostics for invalid case`,
          );
          checkDiagnosticEqual(diags.diagnostics, errors);
        }
      });
    });
  }
}

export function getFixturesRootDir(): string {
  return path.join(import.meta.dirname, 'fixtures');
}

/**
 * Simple no-op tag to mark code samples as "should not format with prettier"
 *   for the plugin-test-formatting lint rule
 */
export function noFormat(raw: TemplateStringsArray, ...keys: string[]): string {
  return String.raw({ raw }, ...keys);
}
