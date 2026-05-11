import { rmSync, writeFileSync } from 'node:fs';
import path from 'node:path';
import util from 'node:util';

import { lint } from '@rslint/core';
import { loadConfigFile, normalizeConfig } from '@rslint/core/config-loader';

// Per-test rslint config builder. The user-facing base config is
// `rslint.config.mjs` (ESM, flat-config); we load it via `loadConfigFile`,
// merge in any per-test `settings`, and serialize the result to a small
// temporary file that `lint()` consumes.
//
// Why we still emit a temporary file: rslint's `lint()` JS API takes a
// config path and the underlying Go binary parses it as hujson (a JSON
// superset) — it does NOT execute JS/TS configs the way the CLI does
// (the CLI pre-loads JS configs and pipes them to the binary via stdin,
// see packages/rslint/dist/cli.js). Until `lint()` exposes the same
// pre-load path or a `settings` override, a serialized intermediate is
// the cleanest way to thread per-test `settings` through to the linter.
//
// Cached at module load so the .mjs is parsed exactly once per test run.
let cachedBase: Record<string, unknown>[] | null = null;
async function getBaseConfig(
  baseConfigPath: string,
): Promise<Record<string, unknown>[]> {
  if (cachedBase) return cachedBase;
  const raw = await loadConfigFile(baseConfigPath);
  cachedBase = normalizeConfig(raw);
  return cachedBase;
}

async function buildConfigForSettings(
  baseConfigPath: string,
  settings: Record<string, unknown> | undefined,
): Promise<{ configPath: string; cleanup: () => void }> {
  const base = await getBaseConfig(baseConfigPath);
  const merged = base.map((entry) => ({
    ...entry,
    settings: {
      ...((entry.settings as object | undefined) ?? {}),
      ...(settings ?? {}),
    },
  }));
  // Write the serialized config next to the base — rslint resolves the
  // relative `tsconfig.*.json` paths inside each entry against the config
  // file's directory, so the temp file must live in the same dir.
  const baseDir = path.dirname(baseConfigPath);
  const cfg = path.join(
    baseDir,
    `.rslint.test-${process.pid}-${Date.now()}-${Math.random().toString(36).slice(2)}.tmp`,
  );
  writeFileSync(cfg, JSON.stringify(merged), 'utf8');
  return {
    configPath: cfg,
    cleanup: () => {
      try {
        rmSync(cfg, { force: true });
      } catch {
        /* best-effort cleanup; never fail a test on rmdir */
      }
    },
  };
}

export interface ValidTestCase {
  name?: string;
  code: string;
  options?: any;
  filename?: string | undefined;
  only?: boolean;
  settings?: Record<string, any> | undefined;
}

interface SuggestionOutput {
  messageId?: string;
  desc?: string;
  data?: Record<string, unknown> | undefined;
  output: string;
}

export interface InvalidTestCase extends ValidTestCase {
  errors: number | (TestCaseError | string)[];
  output?: string | null | undefined;
}

interface TestCaseError {
  message?: string | RegExp;
  messageId?: string;
  type?: string | undefined;
  data?: any;
  line?: number | undefined;
  column?: number | undefined;
  endLine?: number | undefined;
  endColumn?: number | undefined;
  suggestions?: SuggestionOutput[] | undefined;
}

export class RuleTester {
  run(
    ruleName: string,
    rule: never,
    cases: {
      valid: ValidTestCase[];
      invalid: InvalidTestCase[];
    },
  ) {
    ruleName = 'jsx-a11y/' + ruleName;
    describe(ruleName, () => {
      const cwd = process.cwd();
      const config = path.resolve(import.meta.dirname, './rslint.config.mjs');

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
          const options =
            typeof validCase === 'string' ? [] : validCase.options || [];
          const settings =
            typeof validCase === 'string' ? undefined : validCase.settings;
          const defaultFilename = 'src/virtual.tsx';
          const filename =
            typeof validCase === 'string'
              ? defaultFilename
              : (validCase.filename ?? defaultFilename);
          const absoluteFilename = path.resolve(import.meta.dirname, filename);

          const { configPath, cleanup } = await buildConfigForSettings(
            config,
            settings,
          );
          let diags;
          try {
            diags = await lint({
              config: configPath,
              workingDirectory: cwd,
              fileContents: {
                [absoluteFilename]: code,
              },
              ruleOptions: {
                [ruleName]: options,
              },
            });
          } finally {
            cleanup();
          }

          assert(
            diags.diagnostics?.length === 0,
            `Expected no diagnostics for valid case, but got: ${JSON.stringify(diags)}\nCode: ${code}`,
          );
        }
      });
      test('invalid', async (t) => {
        for (const item of cases.invalid) {
          assert.ok(
            item.errors || item.errors === 0,
            `Did not specify errors for an invalid test of ${ruleName}`,
          );
          if (Array.isArray(item.errors) && item.errors.length === 0) {
            assert.fail('Invalid cases must have at least one error');
          }

          const { code, only = false, options = [], settings } = item;
          if (hasOnly && !only) {
            continue;
          }
          const defaultFilename = 'src/virtual.tsx';
          const filename =
            typeof item === 'string'
              ? defaultFilename
              : (item.filename ?? defaultFilename);
          const absoluteFilename = path.resolve(import.meta.dirname, filename);
          const { configPath, cleanup } = await buildConfigForSettings(
            config,
            settings,
          );
          let diags;
          try {
            diags = await lint({
              config: configPath,
              workingDirectory: cwd,
              fileContents: {
                [absoluteFilename]: code,
              },
              ruleOptions: {
                [ruleName]: options,
              },
            });
          } finally {
            cleanup();
          }

          if (typeof item.errors === 'number') {
            if (item.errors === 0) {
              assert.fail(
                "Invalid cases must have 'error' value greater than 0",
              );
            }

            assert.strictEqual(
              diags.diagnostics.length,
              item.errors,
              util.format(
                'Should have %d error%s but had %d: %s',
                item.errors,
                item.errors === 1 ? '' : 's',
                diags.diagnostics.length,
                util.inspect(diags.diagnostics),
              ),
            );
          } else {
            assert.strictEqual(
              diags.diagnostics.length,
              item.errors.length,
              util.format(
                'Should have %d error%s but had %d: %s',
                item.errors.length,
                item.errors.length === 1 ? '' : 's',
                diags.diagnostics.length,
                util.inspect(diags.diagnostics),
              ),
            );

            const hasMessageOfThisRule = diags.diagnostics.some(
              (d) => d.ruleName === ruleName,
            );

            for (let i = 0, l = item.errors.length; i < l; i++) {
              const error = item.errors[i];
              const message = diags.diagnostics[i];

              assert(
                hasMessageOfThisRule,
                'Error rule name should be the same as the name of the rule being tested',
              );
              if (typeof error === 'string' || error instanceof RegExp) {
                assertMessageMatches(message.message, error);
                assert.ok(
                  message.suggestions === void 0,
                  `Error at index ${i} has suggestions. Please convert the test error into an object and specify 'suggestions' property on it to test suggestions.`,
                );
              } else if (typeof error === 'object' && error !== null) {
                if (typeof error.message === 'string') {
                  assertMessageMatches(message.message, error.message);
                }
              }
            }
          }
        }
      });
    });
  }
}

function assertMessageMatches(actual: string, expected: string | RegExp): void {
  if (expected instanceof RegExp) {
    assert.ok(
      expected.test(actual),
      `Expected '${actual}' to match ${expected}`,
    );
  } else {
    assert.strictEqual(actual, expected);
  }
}
