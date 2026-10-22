import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { afterEach, describe, expect, test } from 'rstack/test';
import { getNativeBinding } from '../src/native/binding.js';
import { isRecord } from '../src/rust/engine.js';
import {
  lintRustProjectsSync,
  lintRustProjectsWithSourceSnapshotsSync,
  toPublicLintResults,
  type RustProjectsOptions,
} from '../src/rust/projects.js';

const roots: string[] = [];
function fixture(files: Record<string, string>): string {
  const root = fs.realpathSync(
    fs.mkdtempSync(path.join(os.tmpdir(), 'rslint-rust-projects-')),
  );
  roots.push(root);
  for (const [name, text] of Object.entries(files)) {
    fs.mkdirSync(path.dirname(path.join(root, name)), { recursive: true });
    fs.writeFileSync(path.join(root, name), text);
  }
  return root;
}
afterEach(() => {
  for (const root of roots.splice(0))
    fs.rmSync(root, { recursive: true, force: true });
});

function withBinding(
  lint: (request: string) => string,
  capabilities: (request: string) => string,
  run: () => void,
): void {
  const binding: unknown = getNativeBinding();
  if (!isRecord(binding)) throw new Error('Missing native binding');
  const originalLint = binding.lintProjectsTsrsSync;
  const originalCapabilities = binding.ruleCapabilitiesTsrsSync;
  binding.lintProjectsTsrsSync = lint;
  binding.ruleCapabilitiesTsrsSync = capabilities;
  try {
    run();
  } finally {
    binding.lintProjectsTsrsSync = originalLint;
    binding.ruleCapabilitiesTsrsSync = originalCapabilities;
  }
}

function collectRequest(options: RustProjectsOptions): unknown {
  let captured: unknown;
  withBinding(
    (raw) => {
      captured = JSON.parse(raw);
      throw new Error('captured configured request');
    },
    (raw) => {
      const names: unknown = JSON.parse(raw);
      if (
        !Array.isArray(names) ||
        !names.every((name) => typeof name === 'string')
      )
        throw new Error('Invalid rule metadata request');
      return JSON.stringify(
        Object.fromEntries(names.map((name: string) => [name, false])),
      );
    },
    () => {
      expect(() => lintRustProjectsSync(options)).toThrow(
        'captured configured request',
      );
    },
  );
  return captured;
}

describe('configured direct Rust lint', () => {
  test('the CLI snapshot boundary cannot bypass admission or public option validation', () => {
    const cwd = fixture({
      'input.ts': 'const previous = 1;',
      'ignored.ts': 'debugger;',
    });
    const options = {
      cwd,
      patterns: ['input.ts'],
      config: [{ rules: { 'no-debugger': 'error' } }],
    };
    const snapshots: string[] = [];
    const readSource = (filename: string) => {
      snapshots.push(filename);
      return 'debugger;';
    };
    expect(() =>
      lintRustProjectsWithSourceSnapshotsSync(
        { ...options, admitted: [] } as typeof options,
        readSource,
      ),
    ).toThrow('options field admitted is not supported yet');
    expect(() =>
      lintRustProjectsWithSourceSnapshotsSync(
        { ...options, sources: [] } as typeof options,
        readSource,
      ),
    ).toThrow('cannot be combined');
    expect(snapshots).toEqual([]);
    let captured: unknown;
    withBinding(
      (raw) => {
        captured = JSON.parse(raw);
        throw new Error('captured snapshot request');
      },
      () => JSON.stringify({ 'no-debugger': false }),
      () => {
        expect(() =>
          lintRustProjectsWithSourceSnapshotsSync(options, readSource),
        ).toThrow('captured snapshot request');
      },
    );
    expect(snapshots).toEqual([path.join(cwd, 'input.ts')]);
    expect(captured).toMatchObject({
      files: [
        {
          path: path.join(cwd, 'input.ts'),
          text: 'debugger;',
          policy: { kind: 'sourceOnly' },
          rules: [{ name: 'no-debugger', severity: 2, options: [] }],
        },
      ],
      sources: [],
    });
    if (!isRecord(captured) || !Array.isArray(captured.files))
      throw new Error('missing configured request');
    expect(captured.files).toHaveLength(1);
  });

  test('keeps API configuration cwd separate from CLI source-only cwd', () => {
    const cwd = fixture({
      'entry.ts': 'debugger;',
      'config/settings.json': '{}',
    });
    const configDirectory = path.join(cwd, 'config');
    const options = {
      cwd,
      configDirectory,
      config: [{ basePath: cwd, rules: { 'no-debugger': 'error' as const } }],
    };
    expect(collectRequest(options)).toMatchObject({
      cwd,
      sourceOnlyCurrentDirectory: configDirectory,
    });
    expect(
      collectRequest({ ...options, sourceOnlyCurrentDirectory: cwd }),
    ).not.toHaveProperty('sourceOnlyCurrentDirectory');
    expect(() =>
      lintRustProjectsSync({
        ...options,
        sourceOnlyCurrentDirectory: 'relative',
      }),
    ).toThrow('sourceOnlyCurrentDirectory must be absolute');
  });
  test('retains every admitted file, per-file severity, inherited options and source overlays', () => {
    const cwd = fixture({
      'a.ts': 'debugger;',
      'b.ts': '',
      'clean.ts': '',
      'ignored.ts': '',
    });
    expect(
      collectRequest({
        cwd,
        fix: true,
        singleThreaded: true,
        sources: [
          { path: 'a.ts', text: 'value!!;' },
          { path: 'dependency.ts', text: 'export const n = 1;' },
        ],
        config: [
          { ignores: ['ignored.ts'] },
          { rules: { 'no-debugger': ['error', { sentinel: ['keep', 3] }] } },
          {
            files: ['a.ts'],
            languageOptions: {
              parserOptions: { projectService: true, tsconfigRootDir: cwd },
            },
            rules: { 'no-debugger': 'warn' },
          },
          { files: ['clean.ts'], rules: { 'no-debugger': 'off' } },
        ],
      }),
    ).toEqual({
      schemaVersion: 1,
      cwd,
      fix: true,
      singleThreaded: true,
      files: [
        {
          path: path.join(cwd, 'a.ts'),
          text: 'value!!;',
          policy: { kind: 'service', rootDirectory: cwd },
          rules: [
            {
              name: 'no-debugger',
              severity: 1,
              options: [{ sentinel: ['keep', 3] }],
            },
          ],
        },
        {
          path: path.join(cwd, 'b.ts'),
          policy: { kind: 'sourceOnly' },
          rules: [
            {
              name: 'no-debugger',
              severity: 2,
              options: [{ sentinel: ['keep', 3] }],
            },
          ],
        },
        {
          path: path.join(cwd, 'clean.ts'),
          policy: { kind: 'sourceOnly' },
          rules: [],
        },
      ],
      sources: [
        { path: path.join(cwd, 'dependency.ts'), text: 'export const n = 1;' },
      ],
    });
  });

  test('expands explicit projects in declaration order, sorting each glob and retaining lexical aliases', () => {
    const cwd = fixture({
      'input.ts': '',
      'literal[base]/z.json': '{}',
      'literal[base]/a.json': '{}',
      'literal[base]/sub/tsconfig.json': '{}',
    });
    const base = path.join(cwd, 'literal[base]');
    fs.symlinkSync('sub', path.join(base, 'alias'));
    fs.symlinkSync('..', path.join(base, 'sub/cycle'));
    const actual = collectRequest({
      cwd,
      patterns: ['input.ts'],
      config: [
        {
          languageOptions: {
            parserOptions: {
              tsconfigRootDir: base,
              project: ['z.json', '*.json', '*/tsconfig.json', 'z.json'],
            },
          },
          rules: {},
        },
      ],
    });
    expect(actual).toEqual({
      schemaVersion: 1,
      cwd,
      sources: [],
      files: [
        {
          path: path.join(cwd, 'input.ts'),
          rules: [],
          policy: {
            kind: 'explicit',
            configFiles: [
              path.join(base, 'z.json'),
              path.join(base, 'a.json'),
              path.join(base, 'alias/tsconfig.json'),
              path.join(base, 'sub/tsconfig.json'),
            ],
          },
        },
      ],
    });
  });

  test('accepts an explicit overlay config, preserves empty project declarations, and rejects unmatched paths', () => {
    const cwd = fixture({ 'input.ts': '' });
    const config = (project: unknown) => [
      { languageOptions: { parserOptions: { project } }, rules: {} },
    ];
    expect(
      collectRequest({
        cwd,
        config: config('virtual.json'),
        sources: [{ path: 'virtual.json', text: '{"files":["input.ts"]}' }],
      }),
    ).toEqual({
      schemaVersion: 1,
      cwd,
      files: [
        {
          path: path.join(cwd, 'input.ts'),
          rules: [],
          policy: {
            kind: 'explicit',
            configFiles: [path.join(cwd, 'virtual.json')],
          },
        },
      ],
      sources: [
        {
          path: path.join(cwd, 'virtual.json'),
          text: '{"files":["input.ts"]}',
        },
      ],
    });
    expect(collectRequest({ cwd, config: config([]) })).toEqual({
      schemaVersion: 1,
      cwd,
      sources: [],
      files: [
        {
          path: path.join(cwd, 'input.ts'),
          rules: [],
          policy: { kind: 'explicit', configFiles: [] },
        },
      ],
    });
    expect(() =>
      lintRustProjectsSync({ cwd, config: config('missing.json') }),
    ).toThrow(/tsconfig file .*missing.json.*doesn't exist/);
    expect(() =>
      lintRustProjectsSync({ cwd, config: config('missing/*.json') }),
    ).toThrow('glob pattern "missing/*.json" matched no files');
  });

  test('runs independent overrides, retains source-only files, and reports UTF-16 positions', () => {
    const cwd = fixture({
      'a.ts': "'😀'; debugger;",
      'b.ts': 'debugger;',
      'clean.ts': 'const n = 1;',
    });
    const result = lintRustProjectsSync({
      cwd,
      config: [
        { rules: { 'no-debugger': 'error' } },
        { files: ['a.ts'], rules: { 'no-debugger': 'warn' } },
        { files: ['b.ts'], rules: { 'no-debugger': 'off' } },
      ],
    });
    expect(
      result.files.map((file) => [
        path.basename(file.filePath),
        file.appliedRules,
        file.hasTypeInformation,
      ]),
    ).toEqual([
      ['a.ts', ['no-debugger'], false],
      ['b.ts', [], false],
      ['clean.ts', ['no-debugger'], false],
    ]);
    expect(toPublicLintResults(result)).toEqual([
      {
        filePath: path.join(cwd, 'a.ts'),
        messages: [
          {
            ruleId: 'no-debugger',
            messageId: 'no-debugger',
            message: "Unexpected 'debugger' statement.",
            severity: 1,
            line: 1,
            column: 7,
            endLine: 1,
            endColumn: 16,
          },
        ],
        errorCount: 0,
        warningCount: 1,
        fixableErrorCount: 0,
        fixableWarningCount: 0,
      },
      {
        filePath: path.join(cwd, 'b.ts'),
        messages: [],
        errorCount: 0,
        warningCount: 0,
        fixableErrorCount: 0,
        fixableWarningCount: 0,
      },
      {
        filePath: path.join(cwd, 'clean.ts'),
        messages: [],
        errorCount: 0,
        warningCount: 0,
        fixableErrorCount: 0,
        fixableWarningCount: 0,
      },
    ]);
  });

  test('uses selected cross-file types, leaves imported non-roots source-only, and fixes without writes', () => {
    const main =
      "import { unsafe } from './dependency'; -unsafe; value!!; debugger;";
    const dependency = 'export const unsafe = "text"; debugger;';
    const cwd = fixture({
      'tsconfig.json': JSON.stringify({
        compilerOptions: { strict: true },
        files: ['main.ts', 'clean.ts'],
      }),
      'main.ts': main,
      'dependency.ts': dependency,
      'clean.ts': 'export const n = 1;',
    });
    const config = [
      {
        plugins: ['@typescript-eslint'],
        languageOptions: {
          parserOptions: { projectService: true, tsconfigRootDir: cwd },
        },
        rules: {
          '@typescript-eslint/no-unsafe-unary-minus': 'error',
          '@typescript-eslint/no-extra-non-null-assertion': 'error',
          'no-debugger': 'warn',
        },
      },
    ];
    const result = lintRustProjectsSync({ cwd, config, fix: true });
    const byName = new Map(
      result.files.map((file) => [path.basename(file.filePath), file]),
    );
    expect(byName.get('main.ts')?.hasTypeInformation).toBe(true);
    expect(byName.get('clean.ts')?.hasTypeInformation).toBe(true);
    expect(byName.get('dependency.ts')?.hasTypeInformation).toBe(false);
    expect(byName.get('dependency.ts')?.appliedRules).toEqual([
      '@typescript-eslint/no-extra-non-null-assertion',
      'no-debugger',
    ]);
    expect(byName.get('clean.ts')?.messages).toEqual([]);
    const fixed = byName.get('main.ts')!;
    expect(fixed.output).toBe(main.replace('value!!', 'value!'));
    expect(
      fixed.messages.map((message) => [message.ruleId, message.severity]),
    ).toEqual([
      ['@typescript-eslint/no-unsafe-unary-minus', 2],
      ['no-debugger', 1],
    ]);
    expect(fixed.messages[0]).toEqual({
      ruleId: '@typescript-eslint/no-unsafe-unary-minus',
      messageId: 'unaryMinus',
      message:
        'Argument of unary negation should be assignable to number | bigint but is "text" instead.',
      severity: 2,
      line: 1,
      column: main.indexOf('-unsafe') + 1,
      endLine: 1,
      endColumn: main.indexOf('-unsafe') + 8,
      range: [main.indexOf('-unsafe'), main.indexOf('-unsafe') + 7],
    });
    expect(fs.readFileSync(path.join(cwd, 'main.ts'), 'utf8')).toBe(main);
    expect(fs.readFileSync(path.join(cwd, 'dependency.ts'), 'utf8')).toBe(
      dependency,
    );
    const observed = lintRustProjectsSync({ cwd, config });
    const publicMain = toPublicLintResults(observed).find((file) =>
      file.filePath.endsWith('/main.ts'),
    )!;
    expect(publicMain.errorCount).toBe(2);
    expect(publicMain.warningCount).toBe(1);
    expect(publicMain.fixableErrorCount).toBe(1);
  });

  test('preserves compiler diagnostics and suggestion payloads in the public result', () => {
    const cwd = fixture({
      'invalid.js': 'const n: number = 1;',
      'suggestion.ts': 'object?.prop!;',
    });
    const result = toPublicLintResults(
      lintRustProjectsSync({
        cwd,
        fix: true,
        config: [
          {
            plugins: ['@typescript-eslint'],
            rules: {
              '@typescript-eslint/no-non-null-asserted-optional-chain': 'warn',
            },
          },
        ],
      }),
    );
    expect(result[0]).toEqual({
      filePath: path.join(cwd, 'invalid.js'),
      messages: [
        {
          ruleId: 'TypeScript(TS8010)',
          message: 'Type annotations can only be used in TypeScript files.',
          severity: 2,
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 16,
        },
      ],
      errorCount: 1,
      warningCount: 0,
      fixableErrorCount: 0,
      fixableWarningCount: 0,
    });
    expect(result[1]).toEqual({
      filePath: path.join(cwd, 'suggestion.ts'),
      messages: [
        {
          ruleId: '@typescript-eslint/no-non-null-asserted-optional-chain',
          messageId: 'noNonNullOptionalChain',
          message:
            'Optional chain expressions can return undefined by design - using a non-null assertion is unsafe and wrong.',
          severity: 1,
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 14,
          suggestions: [
            {
              messageId: 'suggestRemovingNonNull',
              desc: 'You should remove the non-null assertion.',
              fix: { range: [12, 13], text: '' },
            },
          ],
        },
      ],
      errorCount: 0,
      warningCount: 1,
      fixableErrorCount: 0,
      fixableWarningCount: 0,
    });
  });

  test('rejects unknown rules and unsupported options even when no project grants types', () => {
    const cwd = fixture({ 'input.ts': '' });
    expect(() =>
      lintRustProjectsSync({
        cwd,
        config: [{ rules: { 'pending-rule': 'error' } }],
      }),
    ).toThrow(/not been migrated to Rust: pending-rule/);
    expect(() =>
      lintRustProjectsSync({
        cwd,
        config: [
          { rules: { 'no-debugger': ['error', { unsupported: true }] } },
          { rules: { 'no-debugger': 'warn' } },
        ],
      }),
    ).toThrow(/no-debugger does not accept options/);
    expect(() =>
      lintRustProjectsSync({
        cwd,
        config: [
          {
            plugins: ['@typescript-eslint'],
            rules: {
              '@typescript-eslint/no-unsafe-unary-minus': [
                'error',
                { unsupported: true },
              ],
            },
          },
        ],
      }),
    ).toThrow(/does not accept options/);
    expect(() =>
      lintRustProjectsSync({
        cwd,
        config: [{ settings: { unsupported: true } }],
      }),
    ).toThrow('settings are not supported yet');
    expect(() =>
      lintRustProjectsSync({
        cwd,
        config: [{ languageOptions: { globals: { custom: 'readonly' } } }],
      }),
    ).toThrow('languageOptions.globals is not supported yet');
  });

  test('returns an empty selection without inventing a linted file', () => {
    const cwd = fixture({ 'ignored.ts': 'debugger;' });
    const result = lintRustProjectsSync({
      cwd,
      config: [{ ignores: ['**/*.ts'] }, { rules: { 'no-debugger': 'error' } }],
    });
    expect(result.files).toEqual([]);
    expect(result.compiler.revision).toMatch(/^[a-f\d]{40}$/);
    expect(toPublicLintResults(result)).toEqual([]);
  });

  test('rejects omitted rules, forged capabilities, missing files and invalid fix payloads', () => {
    const cwd = fixture({ 'input.ts': 'debugger;' });
    const options = { cwd, config: [{ rules: { 'no-debugger': 'error' } }] };
    const response = {
      schemaVersion: 1,
      backend: 'tsrs-rust',
      summary: { fixedIssues: 0, threads: 1 },
      compiler: { revision: '280b426b89cd4ecd6153cf48228826bd60a99c61' },
      files: [
        {
          filePath: path.join(cwd, 'input.ts'),
          hasTypeInformation: false,
          appliedRules: ['no-debugger'],
          messages: [],
        },
      ],
    };
    for (const changed of [
      { ...response, files: [] },
      {
        ...response,
        files: [{ ...response.files[0], hasTypeInformation: undefined }],
      },
      {
        ...response,
        files: [{ ...response.files[0], hasTypeInformation: true }],
      },
      { ...response, files: [{ ...response.files[0], appliedRules: [] }] },
      {
        ...response,
        files: [
          {
            ...response.files[0],
            messages: [
              {
                ruleId: 'no-debugger',
                messageId: 'no-debugger',
                message: 'Unexpected debugger.',
                severity: 2,
                line: 1,
                column: 1,
                endLine: 1,
                endColumn: 10,
                range: [0, 9],
                fix: { range: [0, 1], text: 7 },
              },
            ],
          },
        ],
      },
    ]) {
      withBinding(
        () => JSON.stringify(changed),
        () => '{"no-debugger":false}',
        () => {
          expect(() => lintRustProjectsSync(options)).toThrow();
        },
      );
    }
    for (const metadata of [
      '{}',
      '{"no-debugger":1}',
      '{"no-debugger":false,"extra":false}',
    ]) {
      withBinding(
        () => JSON.stringify(response),
        () => metadata,
        () => {
          expect(() => lintRustProjectsSync(options)).toThrow(
            'native addon returned invalid rule capabilities',
          );
        },
      );
    }
  });
});
