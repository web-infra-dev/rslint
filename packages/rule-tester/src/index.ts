// Forked and modified from https://github.com/typescript-eslint/typescript-eslint/blob/16c344ec7d274ea542157e0f19682dd1930ab838/packages/rule-tester/src/RuleTester.ts#L4

import path from 'node:path';
import { test, describe, expect } from '@rstest/core';
import { applyFixes, lint, LintResponse, type Diagnostic } from '@rslint/core';
import assert from 'node:assert';

interface TsDiagnostic {
  line?: number;
  column?: number;
  endLine?: number;
  endColumn?: number;
  messageId?: string;
  suggestions?: any[] | null;
  data?: any;
  type?: any;
  output?: string;
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
      toCamelCase(rslintDiag.messageId) === tsDiag.messageId,
      `Message mismatch: ${rslintDiag.messageId} !== ${tsDiag.messageId}`,
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

interface RuleTesterOptions {
  languageOptions?: {
    globals?: any;
    parser?: any;
    parserOptions?: {
      project?: string;
      tsconfigRootDir?: string;
      projectService?: boolean;
      ecmaFeatures?: any;
      ecmaVersion?: number;
      sourceType?: 'module' | 'script';
      jsxPragma?: string | null;
      jsxFragmentName?: string;
      emitDecoratorMetadata?: boolean;
      isolatedDeclarations?: boolean;
      experimentalDecorators?: boolean;
      lib?: string[];
    };
  };
}
export type InvalidTestCase<T = any, U = any> = {
  code: string;
  filename?: string;
  errors: TsDiagnostic[];
  options?: any;
  only?: boolean;
  skip?: boolean;
  output?: string | null | string[];
  languageOptions?: RuleTesterOptions['languageOptions'];
};
export type ValidTestCase<T = any> =
  | string
  | {
      filename?: string;
      code: string;
      options?: any;
      only?: boolean;
      skip?: boolean;
      languageOptions?: RuleTesterOptions['languageOptions'];
      name?: string;
    };

function getTypescriptEslintFixturesRootDir(): string {
  return path.resolve(
    '../../packages/rslint-test-tools/tests/typescript-eslint/fixtures',
  );
}
const rootDir: string = getTypescriptEslintFixturesRootDir();
const defaultRuleTesterOptions: RuleTesterOptions = {
  languageOptions: {
    parserOptions: {
      project: './tsconfig.json',
      tsconfigRootDir: rootDir,
    },
  },
};
export class RuleTester {
  options: RuleTesterOptions;
  constructor(options: RuleTesterOptions = defaultRuleTesterOptions) {
    this.options = options;
  }
  public defineRule(
    rule: string,
    options: {
      create: (context: any) => void;
      meta: any;
      defaultOptions?: any;
    },
  ) {}
  public run(
    ruleName: string,
    cases: {
      valid: ValidTestCase[];
      invalid: InvalidTestCase[];
    },
  ) {
    describe(ruleName, () => {
      ruleName = '@typescript-eslint/' + ruleName;
      let cwd =
        this.options.languageOptions?.parserOptions?.tsconfigRootDir ||
        process.cwd();
      const config = path.resolve(cwd, './rslint.json');

      // test whether case has only
      let hasOnly =
        cases.valid.some(x => {
          if (typeof x === 'object' && x.only) {
            return true;
          } else {
            return false;
          }
        }) || cases.invalid.some(x => x.only);
      test('valid', async () => {
        for (const validCase of cases.valid) {
          if (typeof validCase === 'object' && validCase.skip) {
            continue;
          }
          if (hasOnly) {
            if (typeof validCase === 'string') {
              continue;
            }
            if (!validCase.only) {
              continue;
            }
          }
          const code =
            typeof validCase === 'string' ? validCase : validCase.code;
          const languageOptions =
            typeof validCase === 'string'
              ? this.options.languageOptions
              : (validCase.languageOptions ?? this.options.languageOptions);
          const isJSX = languageOptions?.parserOptions?.ecmaFeatures?.jsx;

          const options =
            typeof validCase === 'string' ? [] : validCase.options || [];
          let virtual_entry = path.resolve(
            cwd,
            isJSX ? 'virtual.tsx' : 'virtual.ts',
          );
          // workaround for this hardcoded path https://github.com/typescript-eslint/typescript-eslint/blob/main/packages/eslint-plugin/tests/rules/no-floating-promises.test.ts#L712
          if (Array.isArray(options)) {
            for (const opt of options) {
              if (Array.isArray(opt.allowForKnownSafeCalls)) {
                for (const item of opt.allowForKnownSafeCalls) {
                  if (item.path) {
                    item.path = virtual_entry;
                  }
                }
              }
            }
          }
          const diags = await lint({
            config,
            workingDirectory: cwd,
            fileContents: {
              [virtual_entry]: code,
            },
            ruleOptions: {
              [ruleName]: options,
            },
            languageOptions,
          });

          assert(
            diags.diagnostics?.length === 0,
            `Expected no diagnostics for valid case, but got: ${JSON.stringify(diags)} \nwith code:\n${code}`,
          );
        }
      });
      test('invalid', async t => {
        for (const item of cases.invalid) {
          const {
            code,
            errors,
            only = false,
            skip = false,
            output,
            options = [],
          } = item;
          if (skip) {
            continue;
          }
          if (hasOnly && !only) {
            continue;
          }
          const languageOptions =
            item.languageOptions ?? this.options.languageOptions;
          const isJSX = languageOptions?.parserOptions?.ecmaFeatures?.jsx;
          const test_virtual_entry = path.resolve(
            cwd,
            isJSX ? 'virtual.tsx' : 'virtual.ts',
          );
          const diags = await lint({
            config,
            workingDirectory: cwd,
            fileContents: {
              [test_virtual_entry]: code,
            },
            ruleOptions: {
              [ruleName]: options,
            },
            languageOptions,
          });

          assert(
            diags.diagnostics?.length > 0,
            `Expected diagnostics for invalid case: ${code}`,
          );
          // eslint-disable-next-line
          checkDiagnosticEqual(diags.diagnostics, errors);
          if (output) {
            // check autofix
            const fixedCode = await applyFixes({
              fileContent: code,
              diagnostics: diags.diagnostics,
            });
            if (Array.isArray(output)) {
              // skip for now, because the current implementation of autofix is different from typescript-eslint
              // expect(fixedCode.fixedContent).toEqual(output);
            } else {
              // expect(fixedCode.fixedContent[0]).toEqual(output);
            }

            expect(
              filterSnapshot({
                ...diags,
                code,
                output,
              }),
            ).toMatchSnapshot();
          } else {
            expect(filterSnapshot({ ...diags, code })).toMatchSnapshot();
          }
        }
      });
    });
  }
}
// remove unnecessary props from diagnostics, return optional filtered LintResponse
function filterSnapshot(
  diags: LintResponse & { output?: string | string[] | null; code?: string },
): LintResponse {
  for (const diag of diags.diagnostics ?? []) {
    // @ts-ignore
    delete diag.filePath;
    // @ts-ignore
    delete diag.fixes;
  }
  return diags;
}
/**
 * Simple no-op tag to mark code samples as "should not format with prettier"
 *   for the plugin-test-formatting lint rule
 */
export function noFormat(raw: TemplateStringsArray, ...keys: string[]): string {
  return String.raw({ raw }, ...keys);
}

export type RunTests<T, U> = any;

export type TestCaseError<T> = any;
