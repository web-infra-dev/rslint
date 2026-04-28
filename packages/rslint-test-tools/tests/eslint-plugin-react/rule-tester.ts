import { mkdtempSync, readFileSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import path from 'node:path';
import util from 'node:util';

import { lint } from '@rslint/core';

// Per-test rslint.json builder used to thread `settings` through to the
// linter. The base rslint.json registered for the suite has `settings: {}`;
// when a test case carries its own settings, we emit a temporary config
// that merges the base with the test's settings and point `lint()` at it.
//
// rslint's LintOptions does NOT expose a runtime `settings` override (see
// packages/rslint/dist/types.d.ts), so a config file is the only path. The
// temp dir is cleaned on the same tick the lint() promise resolves.
function buildConfigForSettings(
  baseConfigPath: string,
  settings: Record<string, unknown> | undefined,
): { configPath: string; cleanup: () => void } {
  if (!settings || Object.keys(settings).length === 0) {
    return { configPath: baseConfigPath, cleanup: () => {} };
  }
  const base = JSON.parse(readFileSync(baseConfigPath, 'utf8'));
  const merged = base.map((entry: any) => ({
    ...entry,
    settings: { ...(entry.settings ?? {}), ...settings },
  }));
  // Write the temp config into the SAME directory as the base config:
  // the base entries reference `tsconfig.*.json` via relative paths
  // that rslint resolves from the config file's location, so the temp
  // config must sit next to the originals or those paths break.
  const baseDir = path.dirname(baseConfigPath);
  const cfg = path.join(
    baseDir,
    `rslint.test-${process.pid}-${Date.now()}-${Math.random().toString(36).slice(2)}.json`,
  );
  writeFileSync(cfg, JSON.stringify(merged), 'utf8');
  // Suppress unused-import warnings for the temp-dir helpers — they are
  // kept on the imports list to keep the surface stable for future edits
  // that may need an out-of-tree config (e.g. one that doesn't share the
  // base config's relative tsconfig references).
  void mkdtempSync;
  void tmpdir;
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

// Port from 'eslint'
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
  /**
   * @deprecated `type` is deprecated and will be removed in the next major version.
   */
  type?: string | undefined;
  data?: any;
  line?: number | undefined;
  column?: number | undefined;
  endLine?: number | undefined;
  endColumn?: number | undefined;
  suggestions?: SuggestionOutput[] | undefined;
}

// Port from https://github.com/eslint/eslint/blob/34f0723e2d0faf8ac8dc95ec56e6d181bd6b67f2/lib/rule-tester/rule-tester.js#L1145

export class RuleTester {
  run(
    ruleName: string,
    rule: never,
    cases: {
      valid: ValidTestCase[];
      invalid: InvalidTestCase[];
    },
  ) {
    ruleName = 'react/' + ruleName;
    describe(ruleName, () => {
      const cwd = process.cwd();
      const config = path.resolve(import.meta.dirname, './rslint.json');

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

          const { configPath, cleanup } = buildConfigForSettings(
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
          const { configPath, cleanup } = buildConfigForSettings(
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
                // Just an error message.
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

/**
 * Asserts that the message matches its expected value. If the expected
 * value is a regular expression, it is checked against the actual
 * value.
 */
function assertMessageMatches(actual: string, expected: string | RegExp): void {
  if (expected instanceof RegExp) {
    // assert.js doesn't have a built-in RegExp match function
    assert.ok(
      expected.test(actual),
      `Expected '${actual}' to match ${expected}`,
    );
  } else {
    assert.strictEqual(actual, expected);
  }
}
