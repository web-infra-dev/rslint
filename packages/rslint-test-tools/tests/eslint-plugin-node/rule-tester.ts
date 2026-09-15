import path from 'node:path';
import { afterAll, beforeAll, describe, expect, test } from 'rstack/test';
import { lint } from '@rslint/core/internal';
import { createTempDir, cleanupTempDir } from '../cli/js-config/helpers';

interface ExpectedError {
  messageId: string;
  message: string;
  line: number;
  column: number;
  endLine: number;
  endColumn: number;
}

interface TestCase {
  name?: string;
  code: string;
  filename?: string;
  options?: unknown[];
  settings?: Record<string, unknown>;
  errors?: (string | ExpectedError)[];
  output?: string;
}

// The package metadata mirrors tests/fixtures/shebang at eslint-plugin-n v18.3.0.
const packages = {
  'string-bin': { name: 'test', version: '0.0.0', bin: './bin/test.js' },
  'object-bin': {
    name: 'test',
    version: '0.0.0',
    bin: { a: './bin/a.js', b: './bin/b.js', c: './bin', t: './bin/t.ts' },
  },
  'no-bin-field': { name: 'test', version: '0.0.0' },
  unpublished: { name: 'test', version: '0.0.0', files: ['./published.js'] },
};

export class RuleTester {
  constructor(
    private readonly config: {
      fixtureFiles?: Record<string, string>;
      languageOptions?: {
        globals?: Record<string, 'readonly' | 'writable' | 'off'>;
        sourceType?: 'script' | 'module' | 'commonjs';
      };
    } = {},
  ) {}

  run(
    name: string,
    _rule: unknown,
    cases: { valid: TestCase[]; invalid: TestCase[] },
  ) {
    describe(`node/${name}`, () => {
      let root: string;
      beforeAll(async () => {
        root = await createTempDir(
          this.config.fixtureFiles ??
            Object.fromEntries(
              Object.entries(packages).map(([name, pkg]) => [
                `${name}/package.json`,
                JSON.stringify(pkg),
              ]),
            ),
        );
      });
      afterAll(async () => {
        if (root) await cleanupTempDir(root);
      });
      for (const [kind, entries] of Object.entries(cases)) {
        entries.forEach((item, index) => {
          test(`${kind} ${index}: ${item.name ?? 'BOM and line endings'}`, async () => {
            const filename = path.join(
              root,
              (item.filename ?? 'input.js').replace(
                /^tests\/fixtures\/shebang\//,
                '',
              ),
            );
            const request = {
              configDirectory: root,
              workingDirectory: root,
              config: [
                {
                  plugins: ['node'],
                  languageOptions: {
                    ...this.config.languageOptions,
                    parserOptions: { projectService: false },
                  },
                  settings: item.settings,
                  rules: {
                    [`node/${name}`]: ['error', ...(item.options ?? [])],
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
              expect(diagnostic.ruleName).toBe(`node/${name}`);
              expect(diagnostic.suggestions).toBeUndefined();
              if (typeof expected !== 'string') {
                expect(diagnostic.message).toBe(expected.message);
                expect(diagnostic.messageId).toBe(expected.messageId);
                expect(diagnostic.range).toEqual({
                  start: { line: expected.line, column: expected.column },
                  end: { line: expected.endLine, column: expected.endColumn },
                });
                expect(diagnostic.fixes).toBeUndefined();
                continue;
              }
              // Hashbang's upstream string expectations describe a fix on
              // the complete first line.
              const message = expected;
              const messageId = message.includes('needs shebang')
                ? 'expectedHashbangNode'
                : message.includes('needs no shebang')
                  ? 'expectedHashbang'
                  : message.includes('Unicode BOM')
                    ? 'unexpectedBOM'
                    : 'expectedLF';
              expect(diagnostic.message).toBe(message);
              expect(diagnostic.messageId).toBe(messageId);
              expect(diagnostic.range).toEqual({
                start: { line: 1, column: 1 },
                end: {
                  line: 1,
                  column:
                    item.code
                      .replace(/^\uFEFF/, '')
                      .split(/\r\n|[\r\n\u2028\u2029]/)[0].length + 1,
                },
              });
              expect(diagnostic.fixes?.length).toBe(1);
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
