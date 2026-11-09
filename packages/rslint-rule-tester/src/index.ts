import path from 'node:path';
import fs from 'node:fs/promises';
import { tmpdir } from 'node:os';
import { afterAll, beforeAll, describe, expect, test } from 'rstack/test';
import type { LanguageOptions } from '@rslint/core';
import { lint } from '@rslint/core/internal';

export interface ExpectedError {
  messageId: string;
  message: string | ((root: string) => string);
  line: number;
  column: number;
  endLine: number;
  endColumn: number;
  fix?: { range: [number, number]; text: string };
}

export interface TestCase {
  name?: string;
  /** Reason for skipping this case. */
  skip?: string;
  code: string;
  /** Path relative to the fixture root. */
  filename?: string;
  options?: unknown[] | ((root: string) => unknown[]);
  settings?: Record<string, unknown>;
  languageOptions?: LanguageOptions;
  /** A string compares only the message text. */
  errors?: (string | ExpectedError)[];
  output?: string;
}

export interface RuleTesterConfig {
  /** Plugin name, used as the rule-name prefix. */
  plugin: string;
  /** Files written to the fixture root before the cases run. */
  fixtureFiles?: Record<string, string>;
  languageOptions?: LanguageOptions;
}

export class RuleTester {
  constructor(private readonly config: RuleTesterConfig) {}

  run(
    name: string,
    _rule: unknown,
    cases: { valid: TestCase[]; invalid: TestCase[] },
  ): void {
    const { plugin, fixtureFiles = {}, languageOptions } = this.config;
    const ruleName = `${plugin}/${name}`;
    describe(ruleName, () => {
      let root: string;
      beforeAll(async () => {
        root = await fs.mkdtemp(path.join(tmpdir(), 'rslint-rule-tester-'));
        // Resolved imports use real paths. Keep the fixture root in the same
        // path space on macOS, where the temporary directory is a symlink.
        root = await fs.realpath(root);
        for (const [file, content] of Object.entries(fixtureFiles)) {
          const fullPath = path.join(root, file);
          await fs.mkdir(path.dirname(fullPath), { recursive: true });
          await fs.writeFile(fullPath, content, 'utf8');
        }
      });
      afterAll(async () => {
        // Windows can retain a child process's working directory during
        // teardown, so retry the removal.
        if (root) {
          await fs.rm(root, {
            recursive: true,
            force: true,
            maxRetries: 20,
            retryDelay: 1_000,
          });
        }
      });
      for (const [kind, entries] of Object.entries(cases)) {
        entries.forEach((item, index) => {
          const run = item.skip ? test.skip : test;
          const label = item.skip ?? item.name;
          run(`${kind} ${index}${label ? `: ${label}` : ''}`, async () => {
            const filename = path.join(root, item.filename ?? 'input.js');
            const request = {
              configDirectory: root,
              workingDirectory: root,
              config: [
                {
                  plugins: [plugin],
                  languageOptions: {
                    ...languageOptions,
                    ...item.languageOptions,
                    globals: {
                      ...languageOptions?.globals,
                      ...item.languageOptions?.globals,
                    },
                    parserOptions: {
                      projectService: false,
                      ...languageOptions?.parserOptions,
                      ...item.languageOptions?.parserOptions,
                    },
                  },
                  settings: item.settings,
                  rules: {
                    [ruleName]: [
                      'error',
                      ...(typeof item.options === 'function'
                        ? item.options(root)
                        : (item.options ?? [])),
                    ],
                  },
                },
              ],
              fileContents: { [filename]: item.code },
            };
            const result = await lint(request);
            expect(result.fileCount).toBe(1);
            expect(result.ruleCount).toBe(1);
            expect(result.diagnostics).toHaveLength(item.errors?.length ?? 0);
            for (const [i, expected] of (item.errors ?? []).entries()) {
              const diagnostic = result.diagnostics[i];
              expect(diagnostic.ruleName).toBe(ruleName);
              expect(diagnostic.suggestions).toBeUndefined();
              if (typeof expected === 'string') {
                expect(diagnostic.message).toBe(expected);
                continue;
              }
              expect(diagnostic.message).toBe(
                typeof expected.message === 'function'
                  ? expected.message(root.replaceAll('\\', '/'))
                  : expected.message,
              );
              expect(diagnostic.messageId).toBe(expected.messageId);
              expect(diagnostic.range).toEqual({
                start: { line: expected.line, column: expected.column },
                end: { line: expected.endLine, column: expected.endColumn },
              });
              if (expected.fix) {
                expect(diagnostic.fixes).toEqual([
                  {
                    startPos: expected.fix.range[0],
                    endPos: expected.fix.range[1],
                    text: expected.fix.text,
                  },
                ]);
              } else if (item.output === undefined) {
                expect(diagnostic.fixes).toBeUndefined();
              } else {
                expect(diagnostic.fixes?.length).toBeGreaterThan(0);
              }
            }
            if (item.output !== undefined) {
              const fixed = await lint({ ...request, fix: true });
              expect(Object.values(fixed.output ?? {})).toEqual([item.output]);
              expect(fixed.diagnostics).toEqual([]);
            }
          });
        });
      }
    });
  }
}
