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
  filePayload: Record<string, unknown> = {},
  resultPayload: Record<string, unknown> = {},
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
    summary: { fixedIssues: 0, threads: 1 },
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
        ...filePayload,
      },
    ],
    ...resultPayload,
  };
  binding.lintTsrsSync = () => JSON.stringify(response);
  try {
    check();
  } finally {
    binding.lintTsrsSync = original;
  }
}

describe('experimental direct Rust lint engine', () => {
  test('requires measured fix counts and lint worker metadata', () => {
    const input = request({ fix: true });
    for (const summary of [
      undefined,
      null,
      [],
      {},
      { threads: 1 },
      { fixedIssues: 0 },
      ...[-1, 0.5, Number.MAX_SAFE_INTEGER + 1, '0', null].map(
        (fixedIssues) => ({ fixedIssues, threads: 1 }),
      ),
      ...[-1, 0, 1.5, Number.MAX_SAFE_INTEGER + 1, '1', null].map(
        (threads) => ({ fixedIssues: 0, threads }),
      ),
    ]) {
      withNativeResponse(
        input,
        {},
        () => {
          expect(() => lintRustFilesSync(input)).toThrow(
            'native addon returned an invalid or incomplete lint result',
          );
        },
        {},
        { summary },
      );
    }
    withNativeResponse(
      input,
      {},
      () => {
        expect(lintRustFilesSync(input).summary).toEqual({
          fixedIssues: 2,
          threads: 1,
        });
      },
      {},
      { summary: { fixedIssues: 2, threads: 1 } },
    );
    const noFix = request({ fix: false });
    withNativeResponse(
      noFix,
      {},
      () => {
        expect(() => lintRustFilesSync(noFix)).toThrow(
          'native addon returned an invalid or incomplete lint result',
        );
      },
      {},
      { summary: { fixedIssues: 1, threads: 1 } },
    );
  });

  test('validates compiler diagnostic admission without accepting forged rule results', () => {
    const input = request();
    const syntaxMessage = {
      ruleId: 'TypeScript(TS8010)',
      messageId: undefined,
      message: 'Type annotations can only be used in TypeScript files.',
    };
    const syntaxFile = { hasSyntaxErrors: true, appliedRules: [] };
    const reject = (
      payload: Record<string, unknown>,
      file: Record<string, unknown>,
    ) => {
      withNativeResponse(
        input,
        payload,
        () => {
          expect(() => lintRustFilesSync(input)).toThrow(
            'native addon returned an invalid or incomplete lint result',
          );
        },
        file,
      );
    };
    for (const hasSyntaxErrors of [false, null, 0, 1, 'true', [], {}]) {
      reject(syntaxMessage, { ...syntaxFile, hasSyntaxErrors });
    }
    reject(syntaxMessage, {});
    reject(syntaxMessage, { ...syntaxFile, messages: [] });
    reject(syntaxMessage, { ...syntaxFile, appliedRules: ['no-debugger'] });
    for (const payload of [
      { ruleId: 'unknown-rule' },
      { ruleId: 'no-debugger' },
      { ruleId: 'TypeScript(TSinvalid)' },
      { ruleId: 'prefixTypeScript(TS8010)' },
      { severity: 1 },
      { messageId: '' },
      { messageId: null },
      { fix: { range: [0, 1], text: '' } },
      { suggestions: [] },
    ]) {
      reject({ ...syntaxMessage, ...payload }, syntaxFile);
    }
    reject({ ruleId: 'unknown-rule' }, {});
    reject({ messageId: undefined }, {});
    withNativeResponse(
      input,
      syntaxMessage,
      () => {
        const file = lintRustFilesSync(input).files[0];
        expect(file.hasSyntaxErrors).toBe(true);
        expect(file.appliedRules).toEqual([]);
        expect(file.messages[0]).not.toHaveProperty('messageId');
      },
      syntaxFile,
    );
    // A fix can introduce a syntax error. Its final source still belongs in
    // output, while automatic edits remain forbidden for requests without fix.
    for (const fix of [undefined, false, true]) {
      const input = request({ fix });
      withNativeResponse(
        input,
        syntaxMessage,
        () => {
          if (fix === true) {
            expect(lintRustFilesSync(input).files[0].output).toBe('const = ;');
          } else {
            expect(() => lintRustFilesSync(input)).toThrow(
              'native addon returned an invalid or incomplete lint result',
            );
          }
        },
        { ...syntaxFile, output: 'const = ;' },
      );
    }
  });

  test('validates native output against the requested fix mode and preserves empty output', () => {
    for (const fix of [undefined, false, true]) {
      const input = request({ fix });
      for (const output of [null, false, 0, [], {}]) {
        withNativeResponse(
          input,
          {},
          () => {
            expect(() => lintRustFilesSync(input)).toThrow(
              'native addon returned an invalid or incomplete lint result',
            );
          },
          { output },
        );
      }
      for (const output of ['', input.files[0].text!, 'fixed;']) {
        withNativeResponse(
          input,
          {},
          () => {
            if (fix === true) {
              const file = lintRustFilesSync(input).files[0];
              expect(file).toHaveProperty('output');
              expect(file.output).toBe(output);
            } else {
              expect(() => lintRustFilesSync(input)).toThrow(
                'native addon returned an invalid or incomplete lint result',
              );
            }
          },
          { output },
        );
      }
    }
  });

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
        fix: false,
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
    expect(result.summary).toEqual({ fixedIssues: 0, threads: 1 });
    expect(source).toBe("\uFEFF'😀'; value!!; object?.prop!;");
  });

  test('applies nested automatic fixes in memory and reports diagnostics from the final source', () => {
    const cwd = fixture();
    const source = "\uFEFF'😀'; value!!!!; object?.prop!; debugger;";
    const unchangedDiskSource = 'export const original = true;';
    fs.writeFileSync(path.join(cwd, 'assertions.ts'), unchangedDiskSource);
    const result = lintRustFilesSync(
      request({
        cwd,
        fix: true,
        files: [
          { path: 'assertions.ts', text: source },
          { path: 'silent.ts', text: 'const value = 1;' },
          { path: 'suggestion.ts', text: 'object?.prop!;' },
        ],
        config: [
          {
            rules: {
              '@typescript-eslint/no-extra-non-null-assertion': 'error',
              '@typescript-eslint/no-non-null-asserted-optional-chain': 'warn',
              'no-debugger': 'error',
            },
          },
        ],
      }),
    );
    const [fixed, silent, suggestion] = result.files;
    expect(result.summary).toEqual({ fixedIssues: 3, threads: 1 });
    const output = "\uFEFF'😀'; value!; object?.prop!; debugger;";
    expect(fixed.output).toBe(output);
    expect(fixed.messages.map(({ ruleId }) => ruleId)).toEqual([
      '@typescript-eslint/no-non-null-asserted-optional-chain',
      'no-debugger',
    ]);
    const text = output.slice(1);
    expect(fixed.messages[0].suggestions![0].fix).toEqual({
      range: [text.lastIndexOf('!'), text.lastIndexOf('!') + 1],
      text: '',
    });
    expect(fixed.messages[1].range).toEqual([
      text.indexOf('debugger'),
      text.indexOf('debugger') + 9,
    ]);
    expect(fixed.messages[1].column).toBe(text.indexOf('debugger') + 1);
    expect(silent.messages).toEqual([]);
    expect(silent).not.toHaveProperty('output');
    expect(suggestion.messages).toHaveLength(1);
    expect(suggestion.messages[0].suggestions).toHaveLength(1);
    expect(suggestion).not.toHaveProperty('output');
    expect(result.files.every((file) => file.appliedRules.length === 3)).toBe(
      true,
    );
    expect(source).toBe("\uFEFF'😀'; value!!!!; object?.prop!; debugger;");
    expect(fs.readFileSync(path.join(cwd, 'assertions.ts'), 'utf8')).toBe(
      unchangedDiskSource,
    );
    expect(fs.existsSync(path.join(cwd, 'silent.ts'))).toBe(false);
    expect(fs.existsSync(path.join(cwd, 'suggestion.ts'))).toBe(false);
  });

  test('fixes disk inputs through the config-file entry without writing files', async () => {
    const cwd = fixture();
    fs.writeFileSync(
      path.join(cwd, 'rslint.config.mjs'),
      `export default [{ rules: { '@typescript-eslint/no-extra-non-null-assertion': 'error' } }];`,
    );
    const source = 'declare const value: string | undefined; value!!!;';
    fs.writeFileSync(path.join(cwd, 'input.ts'), source);
    const result = await lintRustFiles({
      cwd,
      files: [{ path: 'input.ts' }],
      configFile: 'rslint.config.mjs',
      fix: true,
    });
    expect(result.files[0].messages).toEqual([]);
    expect(result.files[0].output).toBe(
      'declare const value: string | undefined; value!;',
    );
    expect(fs.readFileSync(path.join(cwd, 'input.ts'), 'utf8')).toBe(source);
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

  test('honors source suppressions and restores reports after enable directives', () => {
    for (const text of [
      '// eslint-disable-next-line no-debugger\ndebugger;',
      '/* rslint-disable no-debugger */ debugger;',
      'debugger; // rslint-disable-line no-debugger',
    ]) {
      const result = lintRustFilesSync(
        request({ files: [{ path: 'input.ts', text }] }),
      );
      expect(result.files).toHaveLength(1);
      expect(result.files[0].appliedRules).toEqual(['no-debugger']);
      expect(result.files[0].messages).toEqual([]);
    }
    const source =
      'debugger;\n/* eslint-disable no-debugger */\ndebugger;\n/* rslint-enable no-debugger */\ndebugger;';
    const result = lintRustFilesSync(
      request({ files: [{ path: 'input.ts', text: source }] }),
    );
    expect(result.files[0].messages).toEqual(
      [1, 5].map((line) => {
        const start =
          source
            .split('\n')
            .slice(0, line - 1)
            .join('\n').length + (line > 1 ? 1 : 0);
        return {
          ruleId: 'no-debugger',
          messageId: 'no-debugger',
          message: "Unexpected 'debugger' statement.",
          severity: 2,
          line,
          column: 1,
          endLine: line,
          endColumn: 10,
          range: [start, start + 9],
        };
      }),
    );
    const fakeComment = lintRustFilesSync(
      request({
        files: [
          {
            path: 'input.ts',
            text: 'const text = "// eslint-disable-next-line no-debugger"; debugger;',
          },
        ],
      }),
    );
    expect(fakeComment.files[0].messages).toHaveLength(1);
    expect(fakeComment.files[0].messages[0].ruleId).toBe('no-debugger');
  });

  test('rejects unsupported auxiliary directives with their source positions', () => {
    for (const [text, kind, content] of [
      [
        '/* ordinary */\n/* eslint no-debugger: off */ debugger;',
        'InlineConfig',
        'eslint no-debugger: off',
      ],
    ]) {
      const input = request({ files: [{ path: 'input.ts', text }] });
      const filePath = path.resolve(input.cwd!, 'input.ts');
      expect(() => lintRustFilesSync(input)).toThrow(
        Object.assign(
          new Error(
            `unsupported ${kind} lint directive at byte 15: ${content} in ${filePath}`,
          ),
          { code: 'GenericFailure' },
        ),
      );
    }
  });

  test('global directives preserve unrelated diagnostics and disable builtin tracking', () => {
    const source = '/* global external:readonly */ external; debugger;';
    const result = lintRustFilesSync(
      request({ files: [{ path: 'input.ts', text: source }] }),
    );
    expect(result.files).toHaveLength(1);
    expect(result.files[0].messages).toHaveLength(1);
    expect(result.files[0].messages[0].ruleId).toBe('no-debugger');
    const disabled = lintRustFilesSync(
      request({
        files: [
          {
            path: 'input.ts',
            text: "/* global Object:off */ Object.defineProperty(foo, 'bar', { get() {} });",
          },
        ],
        config: [{ rules: { 'getter-return': 'error' } }],
      }),
    );
    expect(disabled.files[0].appliedRules).toContain('getter-return');
    expect(disabled.files[0].messages).toEqual([]);
  });

  test('exported directives do not hide default module diagnostics', () => {
    const result = lintRustFilesSync(
      request({
        files: [
          {
            path: 'module.ts',
            text: '/* exported value */ var value = 1; debugger;',
          },
        ],
      }),
    );
    expect(result.files[0].messages).toHaveLength(1);
    expect(result.files[0].messages[0].ruleId).toBe('no-debugger');
  });

  test('does not expose or apply fixes to suppressed source occurrences', () => {
    const text =
      'let value: string | null = null;\n// rslint-disable-next-line @typescript-eslint/no-extra-non-null-assertion\nvalue!!;\nvalue!!;';
    const input = request({
      files: [{ path: 'input.ts', text }],
      config: [
        {
          rules: { '@typescript-eslint/no-extra-non-null-assertion': 'error' },
        },
      ],
    });
    const observed = lintRustFilesSync(input);
    expect(observed.files[0].messages).toHaveLength(1);
    expect(observed.files[0].messages[0].line).toBe(4);
    expect(observed.files[0].messages[0].fix).toBeDefined();
    const fixed = lintRustFilesSync({ ...input, fix: true });
    expect(fixed.files[0].messages).toEqual([]);
    expect(fixed.files[0].output).toBe(
      text.replace('value!!;\nvalue!!;', 'value!!;\nvalue!;'),
    );
    const fullySuppressed = lintRustFilesSync({
      ...input,
      fix: true,
      files: [{ path: 'input.ts', text: '/* eslint-disable */ value!!;' }],
    });
    expect(fullySuppressed.files[0].messages).toEqual([]);
    expect(fullySuppressed.files[0]).not.toHaveProperty('output');
  });

  test('reports compiler syntax errors and still fixes valid admitted files', () => {
    const input = request({
      fix: true,
      files: [
        { path: 'invalid.js', text: 'const value: number = 1; debugger;' },
        { path: 'valid.ts', text: 'value!!; debugger;' },
        { path: 'silent.ts', text: 'const safe = 1;' },
      ],
      config: [
        {
          rules: {
            'no-debugger': 'error',
            '@typescript-eslint/no-extra-non-null-assertion': 'error',
          },
        },
      ],
    });
    const result = lintRustFilesSync(input);
    expect(result.files.map((file) => path.basename(file.filePath))).toEqual([
      'invalid.js',
      'valid.ts',
      'silent.ts',
    ]);
    const [invalid, valid, silent] = result.files;
    expect(invalid.hasSyntaxErrors).toBe(true);
    expect(invalid.appliedRules).toEqual([]);
    expect(invalid.messages).toEqual([
      {
        ruleId: 'TypeScript(TS8010)',
        message: 'Type annotations can only be used in TypeScript files.',
        severity: 2,
        line: 1,
        column: 14,
        endLine: 1,
        endColumn: 20,
        range: [13, 19],
      },
    ]);
    expect(invalid).not.toHaveProperty('output');
    expect(valid.output).toBe('value!; debugger;');
    expect(valid.messages.map(({ ruleId }) => ruleId)).toEqual(['no-debugger']);
    expect(silent.messages).toEqual([]);
    expect(silent).not.toHaveProperty('output');
    for (const file of [valid, silent]) {
      expect(file).not.toHaveProperty('hasSyntaxErrors');
      expect(file.appliedRules).toEqual([
        'no-debugger',
        '@typescript-eslint/no-extra-non-null-assertion',
      ]);
    }
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
    for (const fix of [null, 0, 1, 'true', {}, []]) {
      expect(() =>
        lintRustFilesSync({
          ...request(),
          fix,
        } as unknown as RustLintOptions),
      ).toThrow('fix must be a boolean');
    }
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
