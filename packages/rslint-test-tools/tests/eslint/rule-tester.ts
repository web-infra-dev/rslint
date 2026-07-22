import path from 'node:path';
import { test, describe, expect } from '@rstest/core';
import type { RslintConfigEntry } from '@rslint/core';
import { lint, type LintResponse } from '@rslint/core/internal';
import assert from 'node:assert';

import { buildConfigForSettings } from '../src/util/load-test-config';

type TestLanguageOptions = NonNullable<RslintConfigEntry['languageOptions']>;

interface TsDiagnostic {
  line?: number;
  column?: number;
  endLine?: number;
  endColumn?: number;
  messageId?: string;
}

function toCamelCase(name: string): string {
  return name.replace(/-([a-z])/g, (g) => g[1].toUpperCase());
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
      languageOptions?: TestLanguageOptions;
      filename?: string;
      only?: boolean;
      skip?: boolean;
    };

export interface InvalidTestCase {
  code: string;
  errors: TsDiagnostic[];
  options?: Record<string, unknown>;
  languageOptions?: TestLanguageOptions;
  filename?: string;
  only?: boolean;
  skip?: boolean;
}

function filterSnapshot(diags: LintResponse & { code?: string }): LintResponse {
  // Drop fields that are noise or range-unstable for a rule's diagnostic
  // snapshot: warningCount/fixable*Count are constant or fix-related; fixes &
  // suggestions carry byte-offset ranges that ⑥ rewrites to UTF-16.
  const top = diags as unknown as Record<string, unknown>;
  delete top.warningCount;
  delete top.fixableErrorCount;
  delete top.fixableWarningCount;
  delete top.lintedFiles;
  for (const diag of diags.diagnostics ?? []) {
    const d = diag as unknown as Record<string, unknown>;
    delete d.filePath;
    delete d.fixes;
    delete d.severity;
    delete d.suggestions;
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
    testName = ruleName,
  ) {
    describe(testName, () => {
      const cwd = path.resolve(import.meta.dirname);
      const config = path.resolve(cwd, './rslint.config.mjs');

      let hasOnly =
        cases.valid.some((x) => typeof x === 'object' && x.only) ||
        cases.invalid.some((x) => x.only);

      test('valid', async () => {
        for (const validCase of cases.valid) {
          if (typeof validCase === 'object' && validCase.skip) continue;
          if (hasOnly && (typeof validCase === 'string' || !validCase.only))
            continue;

          const code =
            typeof validCase === 'string' ? validCase : validCase.code;
          const options =
            typeof validCase === 'object' ? validCase.options : undefined;
          const languageOptions =
            typeof validCase === 'object'
              ? validCase.languageOptions
              : undefined;
          const filename =
            typeof validCase === 'object'
              ? (validCase.filename ?? 'src/virtual.ts')
              : 'src/virtual.ts';
          const virtual_entry = path.resolve(cwd, filename);

          const { config: resolvedConfig, configDirectory } =
            await buildConfigForSettings(config, undefined);
          const ruleArgs = options
            ? Array.isArray(options)
              ? options
              : [options]
            : [];
          const diags = await lint({
            config: [
              ...resolvedConfig,
              {
                ...(languageOptions ? { languageOptions } : {}),
                rules: {
                  [ruleName]:
                    ruleArgs.length > 0 ? ['error', ...ruleArgs] : 'error',
                },
              },
            ] as any,
            configDirectory,
            workingDirectory: cwd,
            fileContents: { [virtual_entry]: code },
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

          const { code, errors, options, languageOptions, filename } = item;
          const virtual_entry = path.resolve(cwd, filename ?? 'src/virtual.ts');

          const { config: resolvedConfig, configDirectory } =
            await buildConfigForSettings(config, undefined);
          const ruleArgs = options
            ? Array.isArray(options)
              ? options
              : [options]
            : [];
          const diags = await lint({
            config: [
              ...resolvedConfig,
              {
                ...(languageOptions ? { languageOptions } : {}),
                rules: {
                  [ruleName]:
                    ruleArgs.length > 0 ? ['error', ...ruleArgs] : 'error',
                },
              },
            ] as any,
            configDirectory,
            workingDirectory: cwd,
            fileContents: { [virtual_entry]: code },
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
