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
    // Use messageId if available, otherwise fall back to camelCased ruleName
    const rslintMessageId = rslintDiag.messageId || toCamelCase(rslintDiag.ruleName);
    assert(
      rslintMessageId === tsDiag.messageId,
      `Message mismatch: ${rslintMessageId} !== ${tsDiag.messageId}`,
    );

    // check range match
    // tsDiag sometimes doesn't have line and column, so we need to check that
    // RSLint returns 1-based line/column numbers, same as TypeScript-ESLint tests expect
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
      valid: (string | { code: string; options?: any[]; languageOptions?: any })[];
      invalid: {
        code: string;
        errors: any[];
        options?: any[];
        output?: string | null;
      }[];
    },
  ) {
    test(ruleName, async () => {
      let cwd = path.resolve(import.meta.dirname, './fixtures');
      const config = path.resolve(
        import.meta.dirname,
        './fixtures/rslint.json',
      );
      let virtual_entry = path.resolve(cwd, 'src/virtual.ts');
      await test('valid', async () => {
        for (const testCase of cases.valid) {
          const code = typeof testCase === 'string' ? testCase : testCase.code;
          const options = typeof testCase === 'string' ? undefined : testCase.options;
          
          const ruleConfig = options ? ['error', ...options] : 'error';
          
          const diags = await Promise.race([
            lint({
              config,
              workingDirectory: cwd,
              fileContents: {
                [virtual_entry]: code,
              },
              ruleOptions: {
                [ruleName]: ruleConfig,
              },
            }),
            new Promise((_, reject) => 
              setTimeout(() => reject(new Error(`Timeout after 30s for rule ${ruleName} valid case`)), 30000)
            )
          ]);
          if (diags.diagnostics?.length > 0) {
            console.error('Failed valid test case:', code);
            console.error('Options:', JSON.stringify(options));
            console.error('Rule config:', JSON.stringify(ruleConfig));
          }
          assert(
            diags.diagnostics?.length === 0,
            `Expected no diagnostics for valid case, but got: ${JSON.stringify(diags)}`,
          );
        }
      });
      await test('invalid', async t => {
        for (let i = 0; i < cases.invalid.length; i++) {
          const testCase = cases.invalid[i];
          const { errors, code, options } = testCase;
          
          const ruleConfig = options ? ['error', ...options] : 'error';
          
          const diags = await Promise.race([
            lint({
              config,
              workingDirectory: cwd,
              fileContents: {
                [virtual_entry]: code,
              },
              ruleOptions: {
                [ruleName]: ruleConfig,
              },
            }),
            new Promise((_, reject) => 
              setTimeout(() => reject(new Error(`Timeout after 30s for rule ${ruleName} invalid case`)), 30000)
            )
          ]);
          t.assert.snapshot(diags);
          assert(
            diags.diagnostics?.length > 0,
            `Expected diagnostics for invalid case: ${JSON.stringify({code, options, diags})}`,
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