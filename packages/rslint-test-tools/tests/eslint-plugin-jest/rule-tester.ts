import { mkdtempSync, readFileSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import path from 'node:path';
import util from 'node:util';

import { applyFixes, lint, type Diagnostic } from '@rslint/core';

interface SuggestionOutput {
  messageId?: string;
  desc?: string;
  data?: Record<string, unknown> | undefined;
  output: string;
}

type DiagnosticWithSuggestions = Diagnostic & {
  fixes?: Array<{
    text: string;
    startPos: number;
    endPos: number;
  }>;
  suggestions?: Array<{
    messageId?: string;
    desc?: string;
    data?: Record<string, string>;
    fixes?: Array<{
      text: string;
      startPos: number;
      endPos: number;
    }>;
  }>;
};

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

// Port from 'eslint'
export interface ValidTestCase {
  name?: string;
  code: string;
  options?: any;
  filename?: string | undefined;
  only?: boolean;
  settings?: Record<string, any> | undefined;
}

export interface InvalidTestCase extends ValidTestCase {
  errors: number | (TestCaseError | string)[];
  output?: string | null | undefined;
}

// Per-test rslint.json builder used to thread `settings` through to the
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

export class RuleTester {
  run(
    ruleName: string,
    rule: never,
    cases: {
      valid: ValidTestCase[];
      invalid: InvalidTestCase[];
    },
  ) {
    ruleName = 'jest/' + ruleName;
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
              const message = diags.diagnostics[i] as DiagnosticWithSuggestions;

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
                await assertFixedOutput(message, code, item.output);
                await assertSuggestionOutputs(message, error, code);
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

async function assertFixedOutput(
  diagnostic: DiagnosticWithSuggestions,
  source: string,
  expectedOutput: string | null | undefined,
): Promise<void> {
  if (expectedOutput === undefined) {
    return;
  }

  if (expectedOutput === null) {
    assert.ok(
      !diagnostic.fixes || diagnostic.fixes.length === 0,
      'Expected diagnostic to have no autofix output',
    );
    return;
  }

  assert.ok(
    diagnostic.fixes && diagnostic.fixes.length > 0,
    'Expected diagnostic to include autofix data',
  );

  const fixed = await applyFixes({
    fileContent: source,
    diagnostics: [diagnostic],
  });
  const finalOutput =
    fixed.fixedContent[fixed.fixedContent.length - 1] ?? source;

  assert.strictEqual(finalOutput, expectedOutput);
}

async function assertSuggestionOutputs(
  diagnostic: DiagnosticWithSuggestions,
  error: TestCaseError,
  source: string,
): Promise<void> {
  const expectedSuggestions = error.suggestions;
  const actualSuggestions = diagnostic.suggestions;

  if (!expectedSuggestions) {
    assert.ok(
      actualSuggestions === void 0 || actualSuggestions.length === 0,
      'Error has suggestions. Please specify the expected suggestions in the test case.',
    );
    return;
  }

  assert.ok(actualSuggestions, 'Expected diagnostic to include suggestions');
  assert.strictEqual(actualSuggestions.length, expectedSuggestions.length);

  for (let i = 0; i < expectedSuggestions.length; i++) {
    const expectedSuggestion = expectedSuggestions[i];
    const actualSuggestion = actualSuggestions[i];

    if (expectedSuggestion.messageId !== undefined) {
      assert.strictEqual(
        actualSuggestion.messageId,
        expectedSuggestion.messageId,
      );
    }
    if (expectedSuggestion.desc !== undefined) {
      assert.strictEqual(actualSuggestion.desc, expectedSuggestion.desc);
    }
    if (expectedSuggestion.data !== undefined) {
      assert.deepStrictEqual(actualSuggestion.data, expectedSuggestion.data);
    }

    assert.ok(
      actualSuggestion.fixes && actualSuggestion.fixes.length > 0,
      'Expected suggestion to include fix data',
    );

    const fixed = await applyFixes({
      fileContent: source,
      diagnostics: [
        {
          ...diagnostic,
          fixes: actualSuggestion.fixes,
          suggestions: [],
        },
      ],
    });
    const finalOutput =
      fixed.fixedContent[fixed.fixedContent.length - 1] ?? source;

    assert.strictEqual(finalOutput, expectedSuggestion.output);
  }
}
