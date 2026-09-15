import { readFileSync, mkdtempSync, rmSync } from 'node:fs';
import { tmpdir } from 'node:os';
import path from 'node:path';
import { Linter } from 'eslint';
import * as parser from '@typescript-eslint/parser';
import { lint } from '@rslint/core/internal';
import { afterAll, expect, test } from 'rstack/test';

interface ExpectedDiagnostic {
  message: string;
  line: number;
  column: number;
  endLine: number;
  endColumn: number;
}

interface MemberCase {
  rule: string;
  name: string;
  code: string;
  options: unknown[];
  globals: Record<string, 'readonly' | 'writable' | 'off'>;
  errors: ExpectedDiagnostic[];
}

const cases: MemberCase[] = JSON.parse(
  readFileSync(
    new URL(
      '../../../../internal/rules/testdata/heritage_members.json',
      import.meta.url,
    ),
    'utf8',
  ),
);
if (cases.length === 0)
  throw new Error('heritage member corpus must not be empty');

const cwd = mkdtempSync(path.join(tmpdir(), 'rslint-heritage-members-'));
const file = path.join(cwd, 'member.ts');
const eslint = new Linter();
afterAll(() => rmSync(cwd, { recursive: true, force: true }));

for (const entry of cases) {
  test(`${entry.rule}: ${entry.name}`, async () => {
    const rules = {
      [entry.rule]: ['error', ...entry.options] as Linter.RuleEntry,
    };
    const languageOptions = { globals: entry.globals };
    const expected = entry.errors.map(
      ({ message, line, column, endLine, endColumn }) => ({
        message,
        line,
        column,
        endLine,
        endColumn,
      }),
    );
    const reference = eslint.verify(
      entry.code,
      [
        {
          files: ['**/*.ts'],
          languageOptions: { ...languageOptions, parser },
          rules,
        },
      ],
      { filename: 'member.ts' },
    );
    expect(
      reference.map(({ message, line, column, endLine, endColumn }) => ({
        message,
        line,
        column,
        endLine,
        endColumn,
      })),
    ).toEqual(expected);

    const actual = await lint({
      config: [{ files: ['**/*.ts'], languageOptions, rules }],
      configDirectory: cwd,
      workingDirectory: cwd,
      files: [file],
      fileContents: { [file]: entry.code },
    });
    expect(actual.fileCount).toBe(1);
    expect(
      actual.diagnostics.map(({ message, range }) => ({
        message,
        line: range.start.line,
        column: range.start.column,
        endLine: range.end.line,
        endColumn: range.end.column,
      })),
    ).toEqual(expected);
  });
}
