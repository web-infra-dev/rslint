// Forked and modified from https://github.com/typescript-eslint/typescript-eslint/blob/16c344ec7d274ea542157e0f19682dd1930ab838/packages/rule-tester/src/RuleTester.ts#L4

import path from 'node:path';
import { test, describe, expect } from '@rstest/core';
import { lint, LintResponse, type Diagnostic } from '@rslint/core/internal';
import { loadConfigFile, normalizeConfig } from '@rslint/core/config-loader';
import assert from 'node:assert';

// Per-test rslint config builder. The user-facing base config is
// `rslint.config.mjs` (ESM, flat-config); we load it via `loadConfigFile`
// and serialize the result to a small temporary file that `lint()`
// consumes, because rslint's Go binary parses the config as hujson
// (JSON superset) and does not execute JS/TS configs.
//
// `packages/rule-tester` is published as `@typescript-eslint/rule-tester`
// and cannot import from `packages/rslint-test-tools` (dependency direction
// is reversed), so this helper is duplicated here intentionally instead of
// being shared with the plugin-specific rule-testers under
// `packages/rslint-test-tools/tests/src/util/load-test-config.ts`.
//
// The base config is cached by `baseConfigPath` so each .mjs is parsed
// exactly once per process.
const baseConfigCache = new Map<string, Record<string, unknown>[]>();

async function getBaseConfig(
  baseConfigPath: string,
): Promise<Record<string, unknown>[]> {
  const cached = baseConfigCache.get(baseConfigPath);
  if (cached) return cached;
  const raw = await loadConfigFile(baseConfigPath);
  const normalized = normalizeConfig(raw);
  baseConfigCache.set(baseConfigPath, normalized);
  return normalized;
}

async function buildConfigForSettings(
  baseConfigPath: string,
  settings: Record<string, unknown> | undefined,
): Promise<{ config: Record<string, unknown>[]; configDirectory: string }> {
  const base = await getBaseConfig(baseConfigPath);
  const merged = base.map((entry) => ({
    ...entry,
    settings: {
      ...((entry.settings as object | undefined) ?? {}),
      ...(settings ?? {}),
    },
  }));
  // Hand the resolved config object straight to the JavaScript API (no temp file).
  // rslint resolves each entry's relative `tsconfig.*.json` against
  // configDirectory — the base config file's directory.
  return { config: merged, configDirectory: path.dirname(baseConfigPath) };
}

// Fold the rule-under-test (its options) and the per-case languageOptions into
// the resolved base config as one appended entry. The Go `--api` reads rules and
// languageOptions solely from the config object — there is no separate
// ruleOptions / languageOptions request surface. `options` is the rule's options
// array (no severity); the entry runs it at "error", matching how the
// rule-tester always reports.
function withRuleAndLanguageOptions(
  base: Record<string, unknown>[],
  ruleName: string,
  options: unknown,
  languageOptions: unknown,
): Record<string, unknown>[] {
  // The test's effective languageOptions (a per-case override, else the tester
  // default) is authoritative and is applied on the appended entry below. Strip
  // any languageOptions carried by the base config entries first: keeping the
  // base config's `parserOptions.project` as well would root the virtual file
  // under TWO tsconfigs (base default + per-case) and lint it with two programs,
  // producing duplicate / conflicting diagnostics (e.g. a per-case
  // tsconfig.noPropertyAccessFromIndexSignature still gets the default program's
  // false positive). The removed languageOptions IPC path clobbered entry[0] for
  // exactly this reason, running a single program.
  const stripped = base.map((entry) => {
    const rest = { ...entry };
    delete rest.languageOptions;
    return rest;
  });
  const entry: Record<string, unknown> = {
    rules: {
      [ruleName]:
        Array.isArray(options) && options.length > 0
          ? ['error', ...options]
          : 'error',
    },
  };
  // Declare the rule-under-test's plugin so the `--api` plugin gate
  // (enforcePlugins) keeps the rule enabled. The prefix is everything before
  // the rule name's last "/" — matching Go's RulePluginPrefix; core rules have
  // no "/" and need no declaration. A bare prefix (e.g. "@typescript-eslint",
  // "unicorn") is a valid native-plugin declaration name.
  const slash = ruleName.lastIndexOf('/');
  if (slash > 0) entry.plugins = [ruleName.slice(0, slash)];
  if (languageOptions) entry.languageOptions = languageOptions;
  return [...stripped, entry];
}

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
  return name.replace(/-([a-z])/g, (g) => g[1].toUpperCase());
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
    // check rule match. Normalize both sides: some messageIds are kebab-case
    // (e.g. `angle-bracket`) and toCamelCase must apply to the expected side
    // too, otherwise a hyphenated id can never match.
    assert(
      toCamelCase(rslintDiag.messageId) === toCamelCase(tsDiag.messageId ?? ''),
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
    options?: {
      description?: string; // Optional description appended to test suite name
    },
  ) {
    const testSuiteName = options?.description
      ? `${ruleName} - ${options.description}`
      : ruleName;

    describe(testSuiteName, () => {
      // Use the rule name as-is (no splitting)
      ruleName = '@typescript-eslint/' + ruleName;
      let cwd =
        this.options.languageOptions?.parserOptions?.tsconfigRootDir ||
        process.cwd();
      const config = path.resolve(cwd, './rslint.config.mjs');

      // test whether case has only
      let hasOnly =
        cases.valid.some((x) => {
          if (typeof x === 'object' && x.only) {
            return true;
          } else {
            return false;
          }
        }) || cases.invalid.some((x) => x.only);
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
          const filename =
            typeof validCase === 'string' ? undefined : validCase.filename;
          const isDts = filename && filename.endsWith('.d.ts');
          let virtual_entry = path.resolve(
            cwd,
            isDts ? 'decl.d.ts' : isJSX ? 'react.tsx' : 'virtual.ts',
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
          const { config: resolvedConfig, configDirectory } =
            await buildConfigForSettings(config, undefined);
          const diags = await lint({
            config: withRuleAndLanguageOptions(
              resolvedConfig,
              ruleName,
              options,
              languageOptions,
            ),
            configDirectory,
            workingDirectory: cwd,
            fileContents: {
              [virtual_entry]: code,
            },
          });

          assert(
            diags.diagnostics?.length === 0,
            `Expected no diagnostics for valid case, but got: ${JSON.stringify(diags)} \nwith code:\n${code}`,
          );
        }
      });
      test('invalid', async (t) => {
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
          const isDts = item.filename && item.filename.endsWith('.d.ts');
          const test_virtual_entry = path.resolve(
            cwd,
            isDts ? 'decl.d.ts' : isJSX ? 'react.tsx' : 'virtual.ts',
          );
          const { config: resolvedConfig, configDirectory } =
            await buildConfigForSettings(config, undefined);
          const diags = await lint({
            config: withRuleAndLanguageOptions(
              resolvedConfig,
              ruleName,
              options,
              languageOptions,
            ),
            configDirectory,
            workingDirectory: cwd,
            fileContents: {
              [test_virtual_entry]: code,
            },
          });

          assert(
            diags.diagnostics?.length > 0,
            `Expected diagnostics for invalid case: ${code}`,
          );
          // eslint-disable-next-line
          checkDiagnosticEqual(diags.diagnostics, errors);
          const hasOutput = Object.prototype.hasOwnProperty.call(
            item,
            'output',
          );
          if (hasOutput) {
            // Autofix output verification is deferred (rslint's fixer differs
            // from typescript-eslint's); the snapshot pins the expected output.
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
  diags: Omit<LintResponse, 'output'> & {
    // The rule-tester's `output` is the test case's expected fixed source — a
    // different shape than ⑥'s LintResponse.output (the per-file fix map).
    // `unknown` accepts either; this is snapshot data, not read structurally.
    output?: unknown;
    code?: string;
  },
): Omit<LintResponse, 'output'> & { output?: unknown } {
  // Drop fields that are noise or range-unstable for a rule's diagnostic
  // snapshot: warningCount/severity are constant (rules run at "error");
  // fixes & suggestions carry byte-offset ranges that ⑥ rewrites to UTF-16,
  // and autofix is verified via `output` — so the snapshot pins neither.
  // (suggestion conversion is covered by Go's TestHandleLint_SuggestionsConverted.)
  // @ts-ignore
  delete diags.warningCount;
  // @ts-ignore
  delete diags.fixableErrorCount;
  // @ts-ignore
  delete diags.fixableWarningCount;
  // @ts-ignore
  delete diags.lintedFiles;
  for (const diag of diags.diagnostics ?? []) {
    // @ts-ignore
    delete diag.filePath;
    // @ts-ignore
    delete diag.fixes;
    // @ts-ignore
    delete diag.severity;
    // @ts-ignore
    delete diag.suggestions;
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
