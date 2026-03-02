import path from 'node:path';
import { test, describe, expect } from '@rstest/core';
import { lint, type LintResponse } from '@rslint/core';
import assert from 'node:assert';

interface TsDiagnostic {
  line?: number;
  column?: number;
  endLine?: number;
  endColumn?: number;
  messageId?: string;
}

function toCamelCase(name: string): string {
  return name.replace(/-([a-z])/g, g => g[1].toUpperCase());
}

function checkDiagnosticEqual(
  rslintDiagnostic: {
    messageId: string;
    range: {
      start: { line: number; column: number };
      end: { line: number; column: number };
    };
  }[],
  tsDiagnostic: TsDiagnostic[],
) {
  assert(
    rslintDiagnostic.length === tsDiagnostic.length,
    `Length mismatch: ${rslintDiagnostic.length} !== ${tsDiagnostic.length}`,
  );
  for (let i = 0; i < rslintDiagnostic.length; i++) {
    const rslintDiag = rslintDiagnostic[i];
    const tsDiag = tsDiagnostic[i];
    assert(
      toCamelCase(rslintDiag.messageId) === tsDiag.messageId,
      `Message mismatch: ${rslintDiag.messageId} !== ${tsDiag.messageId}`,
    );
    if (tsDiag.line) {
      assert(
        rslintDiag.range.start.line === tsDiag.line,
        `Start line mismatch: ${rslintDiag.range.start.line} !== ${tsDiag.line}`,
      );
    }
    if (tsDiag.column) {
      assert(
        rslintDiag.range.start.column === tsDiag.column,
        `Start column mismatch: ${rslintDiag.range.start.column} !== ${tsDiag.column}`,
      );
    }
  }
}

export type ValidTestCase =
  | string
  | {
      code: string;
      options?: Record<string, unknown>;
      only?: boolean;
      skip?: boolean;
    };

export interface InvalidTestCase {
  code: string;
  errors: TsDiagnostic[];
  options?: Record<string, unknown>;
  only?: boolean;
  skip?: boolean;
}

function filterSnapshot(diags: LintResponse & { code?: string }): LintResponse {
  for (const diag of diags.diagnostics ?? []) {
    const d = diag as unknown as Record<string, unknown>;
    delete d.filePath;
    delete d.fixes;
  }
  return diags;
}

export class RuleTester {
  public run(
    ruleName: string,
    cases: {
      valid: ValidTestCase[];
      invalid: InvalidTestCase[];
    },
  ) {
    describe(ruleName, () => {
      const cwd = path.resolve(import.meta.dirname);
      const config = path.resolve(cwd, './rslint.json');

      let hasOnly =
        cases.valid.some(x => typeof x === 'object' && x.only) ||
        cases.invalid.some(x => x.only);

      test('valid', async () => {
        for (const validCase of cases.valid) {
          if (typeof validCase === 'object' && validCase.skip) continue;
          if (hasOnly && (typeof validCase === 'string' || !validCase.only))
            continue;

          const code =
            typeof validCase === 'string' ? validCase : validCase.code;
          const options =
            typeof validCase === 'object' ? validCase.options : undefined;
          const virtual_entry = path.resolve(cwd, 'src/virtual.ts');

          const diags = await lint({
            config,
            workingDirectory: cwd,
            fileContents: { [virtual_entry]: code },
            ruleOptions: {
              [ruleName]: options ? [options] : [],
            } as any,
          });

          assert(
            diags.diagnostics?.length === 0,
            `Expected no diagnostics for valid case, but got: ${JSON.stringify(diags)} \nwith code:\n${code}`,
          );
        }
      });

      test('invalid', async () => {
        for (const item of cases.invalid) {
          if (item.skip) continue;
          if (hasOnly && !item.only) continue;

          const { code, errors, options } = item;
          const virtual_entry = path.resolve(cwd, 'src/virtual.ts');

          const diags = await lint({
            config,
            workingDirectory: cwd,
            fileContents: { [virtual_entry]: code },
            ruleOptions: {
              [ruleName]: options ? [options] : [],
            } as any,
          });

          assert(
            diags.diagnostics?.length > 0,
            `Expected diagnostics for invalid case: ${code}`,
          );
          checkDiagnosticEqual(diags.diagnostics, errors);
          expect(filterSnapshot({ ...diags, code })).toMatchSnapshot();
        }
      });
    });
  }
}
