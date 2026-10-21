import { afterEach, describe, expect, test } from 'rstack/test';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import {
  lintRustFiles,
  lintRustFilesSync,
  type RustLintOptions,
} from '../src/rust/index.js';

const roots: string[] = [];

function fixture(): string {
  const root = fs.mkdtempSync(path.join(os.tmpdir(), 'rslint-direct-rust-'));
  roots.push(root);
  return root;
}

afterEach(() => {
  for (const root of roots.splice(0)) {
    fs.rmSync(root, { recursive: true, force: true });
  }
});

function request(overrides: Partial<RustLintOptions> = {}): RustLintOptions {
  return {
    cwd: fixture(),
    files: [{ path: 'input.ts', text: 'debugger;' }],
    config: [{ rules: { 'no-debugger': 'error' } }],
    ...overrides,
  };
}

describe('experimental direct Rust lint engine', () => {
  test('executes typed cross-file rules and retains silent admitted files', () => {
    const input = request({
      files: [
        {
          path: 'types.ts',
          text: 'export declare const unsafe: string; export declare const safe: number;',
        },
        {
          path: 'main.ts',
          text: "import { unsafe, safe } from './types';\n-unsafe; -safe; debugger;",
        },
        { path: 'clean.ts', text: 'export const value = -1;' },
      ],
      config: [
        {
          name: 'direct Rust contract',
          rules: {
            'no-debugger': 'warn',
            '@typescript-eslint/no-unsafe-unary-minus': 'error',
          },
        },
      ],
    });
    const result = lintRustFilesSync(input);
    expect(result.schemaVersion).toBe(1);
    expect(result.backend).toBe('tsrs-rust');
    expect(result.compiler.revision).toMatch(/^[a-f\d]{40}$/);
    expect(result.files.map((file) => path.basename(file.filePath))).toEqual([
      'types.ts',
      'main.ts',
      'clean.ts',
    ]);
    for (const file of result.files) {
      expect(file.appliedRules).toEqual([
        'no-debugger',
        '@typescript-eslint/no-unsafe-unary-minus',
      ]);
    }
    expect(result.files[0].messages).toEqual([]);
    expect(result.files[2].messages).toEqual([]);
    const source = input.files[1].text!;
    const start = source.indexOf('-unsafe');
    expect(result.files[1].messages).toEqual([
      {
        ruleId: '@typescript-eslint/no-unsafe-unary-minus',
        messageId: 'unaryMinus',
        message:
          'Argument of unary negation should be assignable to number | bigint but is string instead.',
        severity: 2,
        line: 2,
        column: 1,
        endLine: 2,
        endColumn: 8,
        range: [start, start + 7],
      },
      {
        ruleId: 'no-debugger',
        messageId: 'no-debugger',
        message: "Unexpected 'debugger' statement.",
        severity: 1,
        line: 2,
        column: 17,
        endLine: 2,
        endColumn: 26,
        range: [source.indexOf('debugger'), source.indexOf('debugger') + 9],
      },
    ]);
  });

  test('loads only an explicit JS config and lints disk sources in one tsconfig', async () => {
    const cwd = fixture();
    fs.writeFileSync(
      path.join(cwd, 'rslint.config.mjs'),
      `export default [{ rules: {
        'no-debugger': 'error',
        '@typescript-eslint/no-unsafe-unary-minus': 'warn'
      } }];`,
    );
    fs.writeFileSync(
      path.join(cwd, 'tsconfig.json'),
      JSON.stringify({
        compilerOptions: { strict: true, target: 'ES2022', module: 'ESNext' },
        files: ['types.ts', 'main.ts', 'clean.ts', 'unselected.ts'],
      }),
    );
    fs.writeFileSync(
      path.join(cwd, 'types.ts'),
      'export const value = "text";',
    );
    fs.writeFileSync(
      path.join(cwd, 'main.ts'),
      "import { value } from './types'; -value; debugger;",
    );
    fs.writeFileSync(path.join(cwd, 'clean.ts'), 'export const value = -1;');
    fs.writeFileSync(path.join(cwd, 'unselected.ts'), 'debugger;');
    const result = await lintRustFiles({
      cwd,
      files: [{ path: 'main.ts' }, { path: 'clean.ts' }],
      project: 'tsconfig.json',
      configFile: 'rslint.config.mjs',
    });
    expect(result.files.map((file) => path.basename(file.filePath))).toEqual([
      'main.ts',
      'clean.ts',
    ]);
    expect(
      result.files[0].messages.map(({ ruleId, severity }) => [
        ruleId,
        severity,
      ]),
    ).toEqual([
      ['@typescript-eslint/no-unsafe-unary-minus', 1],
      ['no-debugger', 2],
    ]);
    expect(result.files[1].messages).toEqual([]);
    expect(fs.readFileSync(path.join(cwd, 'main.ts'), 'utf8')).toBe(
      "import { value } from './types'; -value; debugger;",
    );
  });

  test('preserves UTF-16 locations, config order, and explicit disabled rules', () => {
    const result = lintRustFilesSync(
      request({
        files: [{ path: 'input.ts', text: "'😀'; debugger;\r\ndebugger;" }],
        config: [
          { rules: { 'no-debugger': 'error', 'not-migrated': 'error' } },
          [{ rules: { 'no-debugger': 'warn', 'not-migrated': 'off' } }],
        ],
      }),
    );
    expect(result.files[0].appliedRules).toEqual(['no-debugger']);
    expect(
      result.files[0].messages.map(({ severity, line, column, range }) => ({
        severity,
        line,
        column,
        range,
      })),
    ).toEqual([
      { severity: 1, line: 1, column: 7, range: [6, 15] },
      { severity: 1, line: 2, column: 1, range: [17, 26] },
    ]);
  });

  test('rejects enabled rules pending migration, unsupported options, and duplicate admissions', () => {
    expect(() =>
      lintRustFilesSync(
        request({ config: [{ rules: { 'not-migrated': 'error' } }] }),
      ),
    ).toThrow(/not been migrated to Rust: not-migrated/);
    expect(() =>
      lintRustFilesSync(
        request({
          config: [
            { rules: { 'no-debugger': ['error', { unsupported: true }] } },
            { rules: { 'no-debugger': 'warn' } },
          ],
        }),
      ),
    ).toThrow(/no-debugger does not accept options/);
    expect(() =>
      lintRustFilesSync(
        request({
          files: [
            { path: 'same.ts', text: '' },
            { path: './same.ts', text: '' },
          ],
        }),
      ),
    ).toThrow(/duplicate admitted file/);
    expect(() =>
      lintRustFilesSync(request({ files: [{ path: 'missing.ts' }] })),
    ).toThrow(/missing from compiler project/);
  });

  test('rejects source directives and syntax failures rather than reporting partial success', () => {
    expect(() =>
      lintRustFilesSync(
        request({
          files: [
            {
              path: 'input.ts',
              text: '// eslint-disable-next-line no-debugger\ndebugger;',
            },
          ],
        }),
      ),
    ).toThrow(/directive/i);
    expect(() =>
      lintRustFilesSync(
        request({
          files: [
            {
              path: 'input.ts',
              text: '/* rslint-disable no-debugger */ debugger;',
            },
          ],
        }),
      ),
    ).toThrow(/directive/i);
    expect(() =>
      lintRustFilesSync(
        request({ files: [{ path: 'input.ts', text: 'const = ;' }] }),
      ),
    ).toThrow(/syntax|parse/i);
    const result = lintRustFilesSync(
      request({
        files: [
          {
            path: 'input.ts',
            text: 'const text = "// eslint-disable-next-line no-debugger"; debugger;',
          },
        ],
      }),
    );
    expect(result.files[0].messages).toHaveLength(1);
    expect(result.files[0].messages[0].ruleId).toBe('no-debugger');
  });

  test('rejects unsupported config semantics before normalization can discard fields', () => {
    for (const [key, value] of Object.entries({
      files: ['**/*.ts'],
      ignores: ['ignored.ts'],
      basePath: '.',
      plugins: [],
      languageOptions: { parserOptions: { projectService: true } },
      settings: {},
      linterOptions: { noInlineConfig: true },
      arbitraryUnsupportedField: undefined,
    })) {
      expect(() =>
        lintRustFilesSync(
          request({
            config: [{ rules: { 'no-debugger': 'error' }, [key]: value }],
          }),
        ),
      ).toThrow(`config[0] field ${key} is not supported yet`);
    }
    expect(() =>
      lintRustFilesSync({
        ...request(),
        fix: true,
      } as unknown as RustLintOptions),
    ).toThrow('fix is not supported yet');
    expect(() => lintRustFilesSync({ ...request(), files: [] })).toThrow(
      /every admitted input/,
    );
    expect(() =>
      lintRustFilesSync(
        request({ project: 'tsconfig.json', compilerOptions: {} }),
      ),
    ).toThrow(/compilerOptions cannot override an explicit project/);
  });

  test('does not silently JSON-coerce unsupported options or accept ambiguous config sources', async () => {
    expect(() =>
      lintRustFilesSync(
        request({
          config: [{ rules: { 'no-debugger': ['error', undefined] } }],
        }),
      ),
    ).toThrow(/must contain only JSON values/);
    expect(() =>
      lintRustFilesSync(
        request({
          files: [{ path: 'input.ts', text: "const text = '\uD800';" }],
        }),
      ),
    ).toThrow(/unpaired surrogate/);
    expect(() =>
      lintRustFilesSync(request({ compilerOptions: { target: Number.NaN } })),
    ).toThrow(/must contain only JSON values/);
    await expect(
      lintRustFiles({
        ...request(),
        configFile: 'rslint.config.mjs',
      } as RustLintOptions),
    ).rejects.toThrow(/options field configFile is not supported yet/);
  });
});
