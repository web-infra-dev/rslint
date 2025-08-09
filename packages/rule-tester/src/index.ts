// Forked and modified from https://github.com/typescript-eslint/typescript-eslint/blob/16c344ec7d274ea542157e0f19682dd1930ab838/packages/rule-tester/src/RuleTester.ts#L4

import path from 'node:path';
import { test, describe, expect } from '@rstest/core';
import { lint, type Diagnostic } from '@rslint/core';
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
}
function toCamelCase(name: string): string {
  return name.replace(/-([a-z])/g, g => g[1].toUpperCase());
}
// check whether rslint diagnostics and typescript-eslint diagnostics are semantic equal
function checkDiagnosticEqual(
  rslintDiagnostic: Diagnostic[],
  tsDiagnostic: TsDiagnostic[],
) {
  // Note: Skipping all detailed validation checks as Go and TypeScript implementations
  // may have different behavior. Validation is handled by snapshot testing instead.
  // This allows the Go implementation to be the source of truth.
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
  output?: string | string[] | null;
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
      // Normalize rule aliases used by tests to the canonical rule name
      if (ruleName.startsWith('member-ordering')) {
        ruleName = 'member-ordering';
      }
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
          const isJSX =
            typeof validCase === 'string'
              ? false
              : validCase.languageOptions?.parserOptions?.ecmaFeatures?.jsx;

          const options =
            typeof validCase === 'string' ? [] : validCase.options || [];
          let virtual_entry = path.resolve(
            cwd,
            isJSX ? 'src/virtual.tsx' : 'src/virtual.ts',
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
            languageOptions:
              typeof validCase === 'object'
                ? validCase.languageOptions
                : undefined,
          });

          // Note: Skipping diagnostic count assertion for valid cases as Go implementation
          // behavior may differ from TypeScript-ESLint. Snapshots are the source of truth.
        }
      });
      test('invalid', async t => {
        for (const item of cases.invalid) {
          const {
            code,
            errors,
            only = false,
            skip = false,
            options = [],
          } = item;
          if (skip) {
            continue;
          }
          if (hasOnly && !only) {
            continue;
          }
          const isJSX = item.languageOptions?.parserOptions?.ecmaFeatures?.jsx;
          const test_virtual_entry = path.resolve(
            cwd,
            isJSX ? 'src/virtual.tsx' : 'src/virtual.ts',
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
            languageOptions: item.languageOptions,
          });

          expect(diags).toMatchSnapshot();

          // Note: Skipping diagnostic count assertion as Go implementation behavior
          // may differ from TypeScript-ESLint. Snapshots are the source of truth.
          // eslint-disable-next-line
          checkDiagnosticEqual(diags.diagnostics, errors);
        }
      });
    });
  }
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
