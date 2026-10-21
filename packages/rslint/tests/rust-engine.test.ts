import { afterEach, describe, expect, test } from 'rstack/test';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { getNativeBinding } from '../src/native/binding.js';
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

function withNativeResponse(
  input: RustLintOptions,
  payload: Record<string, unknown>,
  check: () => void,
): void {
  // Match the existing IPC tests' scoped native-export replacement. Keep the
  // real addon for integration cases, restoring it even if an assertion fails.
  const binding = getNativeBinding() as ReturnType<typeof getNativeBinding> & {
    lintTsrsSync(request: string): string;
  };
  const original = binding.lintTsrsSync;
  expect(typeof original).toBe('function');
  const response = {
    schemaVersion: 1,
    backend: 'tsrs-rust',
    compiler: { revision: '280b426b89cd4ecd6153cf48228826bd60a99c61' },
    files: [
      {
        filePath: path.resolve(input.cwd!, input.files[0].path),
        appliedRules: ['no-debugger'],
        messages: [
          {
            ruleId: 'no-debugger',
            messageId: 'no-debugger',
            message: "Unexpected 'debugger' statement.",
            severity: 2,
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 10,
            range: [0, 9],
            ...payload,
          },
        ],
      },
    ],
  };
  binding.lintTsrsSync = () => JSON.stringify(response);
  try {
    check();
  } finally {
    binding.lintTsrsSync = original;
  }
}

describe('experimental direct Rust lint engine', () => {
  test('rejects malformed native fix and suggestion payloads without coercion', () => {
    const input = request();
    const fix = { range: [0, 1], text: '' };
    const suggestion = { messageId: 'remove', desc: 'Remove it.', fix };
    const malformedFixes: unknown[] = [
      null,
      [],
      {},
      { range: [0, 1], text: 1 },
      { range: ['0', 1], text: '' },
      { range: [0, '1'], text: '' },
      { range: [-2, 0], text: '' },
      { range: [-1, -1], text: '' },
      { range: [2, 1], text: '' },
      { range: [0, 1, 2], text: '' },
      { range: [0, 0.5], text: '' },
      { range: [0, Number.MAX_SAFE_INTEGER + 1], text: '' },
      { range: [null, 0], text: '' },
      { range: { 0: 0, 1: 1, length: 2 }, text: '' },
    ];
    const payloads = [
      ...malformedFixes.map((fix) => ({ fix })),
      ...malformedFixes.map((fix) => ({
        suggestions: [{ ...suggestion, fix }],
      })),
      ...[
        null,
        {},
        [null],
        [{}],
        [{ ...suggestion, messageId: 1 }],
        [{ ...suggestion, desc: 1 }],
        [{ ...suggestion, messageId: undefined }],
        [{ ...suggestion, desc: undefined }],
        [{ ...suggestion, fix: undefined }],
        [{ ...suggestion, data: null }],
        [{ ...suggestion, data: [] }],
        [{ ...suggestion, data: { name: 42 } }],
        [{ ...suggestion, data: { name: false } }],
      ].map((suggestions) => ({ suggestions })),
    ];
    for (const payload of payloads) {
      withNativeResponse(input, payload, () => {
        expect(() => lintRustFilesSync(input)).toThrow(
          'native addon returned an invalid or incomplete lint result',
        );
      });
    }
  });

  test('preserves valid native edits including merged BOM removal and suggestion data', () => {
    const input = request({
      files: [{ path: 'input.ts', text: '\uFEFFdebugger;' }],
    });
    for (const range of [
      [0, 0],
      [1, 2],
      [-1, 0],
      [-1, 3],
    ]) {
      const fix = { range, text: '😀' };
      const suggestions = [
        { messageId: 'replace', desc: 'Replace it.', fix },
        {
          messageId: 'replace',
          desc: 'Replace the name.',
          data: { name: 'value' },
          fix,
        },
      ];
      withNativeResponse(input, { fix, suggestions }, () => {
        const message = lintRustFilesSync(input).files[0].messages[0];
        expect(message.fix).toEqual(fix);
        expect(message.suggestions).toEqual(suggestions);
        expect(message.suggestions![0]).not.toHaveProperty('data');
      });
    }
  });

  test('returns UTF-16 fix and suggestion payloads without changing source', () => {
    const source = "\uFEFF'😀'; value!!; object?.prop!;";
    const result = lintRustFilesSync(
      request({
        files: [{ path: 'assertions.ts', text: source }],
        config: [
          {
            rules: {
              '@typescript-eslint/no-extra-non-null-assertion': 'error',
              '@typescript-eslint/no-non-null-asserted-optional-chain': 'warn',
            },
          },
        ],
      }),
    );
    const messages = result.files[0].messages;
    expect(messages).toHaveLength(2);
    const text = source.slice(1);
    const fix = messages[0].fix!;
    expect(fix).toEqual({
      range: [text.indexOf('!!'), text.indexOf('!!') + 1],
      text: '',
    });
    expect(messages[0].suggestions).toBeUndefined();
    expect(messages[1].fix).toBeUndefined();
    expect(messages[1].suggestions).toEqual([
      {
        messageId: 'suggestRemovingNonNull',
        desc: 'You should remove the non-null assertion.',
        fix: {
          range: [text.lastIndexOf('!'), text.lastIndexOf('!') + 1],
          text: '',
        },
      },
    ]);
    expect(result.files[0]).not.toHaveProperty('output');
    expect(source).toBe("\uFEFF'😀'; value!!; object?.prop!;");
  });
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
