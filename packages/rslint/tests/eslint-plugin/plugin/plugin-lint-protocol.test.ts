/**
 * Unit tests for the shared eslintPlugins lint boundary helpers used by
 * both the CLI host (engine.ts) and the LSP host (PluginLintPool.ts).
 *
 * `buildPluginLintTasks` forwards each file's `configKey` verbatim — the
 * worker picks the right `LoadedPlugins` from its per-config map via that
 * key. The helper here is responsible for:
 *
 *   - emitting `configKey` on every task (empty string when absent),
 *   - firing `onUnknownConfigKey` for hosts that want a clearer log
 *     before the worker's internal-error parseError lands,
 *   - propagating the shared `rules` / `collectFixes` / `suggestionsMode`
 *     block to every task,
 *   - forwarding `languageOptions` / `settings` opaquely to the worker.
 */

import { describe, test, expect } from 'rstack/test';

import {
  buildPluginLintTasks,
  buildPluginLintResult,
  type EslintPluginLintRequest,
  type ResolvedEslintPluginLintRequest,
} from '../../../src/eslint-plugin/plugin/plugin-lint-protocol.js';
import type { LintFileResult } from '../../../src/eslint-plugin/linter/ecma-language-plugin.js';
import { resolvePluginAttachments } from '../../../src/eslint-plugin/plugin/attachments.js';

describe('shared plugin host attachment references', () => {
  test('preserves complete text, native capabilities and file metadata', () => {
    const source = '\ufeffconst café = "😀";\r\n// \u0000';
    const capability = { lease: 7, offset: 3, length: 2 };
    const request: EslintPluginLintRequest = {
      collectFixes: true,
      rules: { 'local/check': {} },
      files: [
        { path: 'a.ts', textAttachment: 0, configKey: 'config-a' },
        { path: 'b.ts', text: 'existing inline' },
        { path: 'c.ts', textAttachment: 1 },
        { path: 'd.ts' },
      ],
    };
    expect(resolvePluginAttachments(request, [source, capability])).toEqual({
      ...request,
      files: [
        { path: 'a.ts', text: source, configKey: 'config-a' },
        { path: 'b.ts', text: 'existing inline' },
        { path: 'c.ts', sharedSource: capability },
        { path: 'd.ts' },
      ],
    });
    expect(request.files[0]).toHaveProperty('textAttachment', 0);
  });

  test('preserves an explicitly empty snapshot', () => {
    expect(
      resolvePluginAttachments(
        {
          files: [{ path: '/missing.ts', textAttachment: 0 }],
          collectFixes: false,
        },
        [''],
      ),
    ).toEqual({
      files: [{ path: '/missing.ts', text: '' }],
      collectFixes: false,
    });
  });

  test('an explicitly absent attachment remains an inline snapshot', () => {
    const resolved = resolvePluginAttachments({
      files: [
        { path: '/missing.ts', text: 'snapshot', textAttachment: undefined },
      ],
      collectFixes: false,
    });
    const [task] = buildPluginLintTasks(resolved, { configDirSet: new Set() });
    expect(task.text).toBe('snapshot');
    expect(resolved.files[0]).not.toHaveProperty('textAttachment');
  });

  test('decodes owned binary source at the application boundary', () => {
    const text = '\ufeffconst café = "😀";\r\n// \u0000';
    const resolved = resolvePluginAttachments(
      { files: [{ path: 'a.ts', textAttachment: 0 }], collectFixes: false },
      [Buffer.from(text, 'utf8')],
    );
    expect(resolved.files).toEqual([{ path: 'a.ts', text }]);
    const tasks = buildPluginLintTasks(resolved, { configDirSet: new Set() });
    expect(tasks[0].text).toBe(text);
    expect(tasks[0].sharedSource).toBeUndefined();
  });

  test.each(
    [
      [0xff],
      [0xc0, 0xaf],
      [0xed, 0xa0, 0x80],
      [0xe2, 0x82],
      [0x61, 0x80, 0x62],
    ].map((bytes) => ({ bytes })),
  )(
    'rejects malformed UTF-8 byte snapshots instead of replacing source text: %j',
    ({ bytes }) => {
      expect(() =>
        resolvePluginAttachments(
          { files: [{ path: 'a.ts', textAttachment: 0 }], collectFixes: false },
          [Uint8Array.from(bytes)],
        ),
      ).toThrow();
    },
  );

  test('preserves a BOM and decodes only the supplied byte view', () => {
    const text = '\ufeffconst café = "😀";\r\n';
    const bytes = Buffer.concat([
      Buffer.from([0xff]),
      Buffer.from(text),
      Buffer.from([0xff]),
    ]);
    const view = new Uint8Array(
      bytes.buffer,
      bytes.byteOffset + 1,
      bytes.byteLength - 2,
    );
    const resolved = resolvePluginAttachments(
      { files: [{ path: 'a.ts', textAttachment: 0 }], collectFixes: false },
      [view],
    );
    expect(resolved.files[0].text).toBe(text);
  });

  test.each([
    { files: [{ textAttachment: -1 }] },
    { files: [{ textAttachment: 0.5 }] },
    { files: [{ textAttachment: '0' }] },
    { files: [{ textAttachment: 1 }] },
    { files: [{ textAttachment: 0, text: '' }] },
    { files: [{ textAttachment: 0, sharedSource: {} }] },
    { files: [{ textAttachment: 0 }, { textAttachment: 0 }] },
    { files: [{}] },
    { files: [null] },
  ])('rejects ambiguous or incomplete references: %j', (request) => {
    expect(() =>
      // @ts-expect-error Exercise the runtime guard with malformed wire input.
      resolvePluginAttachments(request, ['complete source']),
    ).toThrow();
  });

  test.each([
    { sharedSource: { lease: 1, offset: 0, length: 0 } },
    { sourceRange: { offset: 0, length: 0 } },
    { sourceIndex: 0 },
  ])(
    'rejects native or legacy wire fields without attachments: %j',
    (source) => {
      expect(() =>
        resolvePluginAttachments({
          files: [{ path: '/missing.ts', ...source }],
          collectFixes: false,
        }),
      ).toThrow('invalid plugin source file');
    },
  );

  test('a missing attachment cannot become a filesystem read', () => {
    expect(() =>
      resolvePluginAttachments({
        files: [{ path: '/missing.ts', textAttachment: 0 }],
        collectFixes: false,
      }),
    ).toThrow('invalid plugin source attachment');
  });
});

function input(
  files: ResolvedEslintPluginLintRequest['files'],
  opts: Partial<ResolvedEslintPluginLintRequest> = {},
): ResolvedEslintPluginLintRequest {
  return {
    files,
    rules: opts.rules ?? { 'uc/no-null': { options: [] } },
    collectFixes: opts.collectFixes ?? false,
    suggestionsMode: opts.suggestionsMode,
  };
}

describe('buildPluginLintTasks', () => {
  test('rejects an unresolved wire range instead of reading disk', () => {
    const request: EslintPluginLintRequest = {
      files: [{ path: '/missing.ts', textAttachment: 0 }],
      collectFixes: false,
    };
    expect(() =>
      // @ts-expect-error A wire request must pass through the attachment adapter.
      buildPluginLintTasks(request, { configDirSet: new Set() }),
    ).toThrow('unresolved shared plugin source');
  });
  test('forwards native source capabilities without decoding or reading files', () => {
    const sharedSource = { lease: 1, offset: 0, length: 12 };
    const tasks = buildPluginLintTasks(
      input([{ path: '/missing.ts', sharedSource }]),
      { configDirSet: new Set() },
    );
    expect(tasks[0].sharedSource).toBe(sharedSource);
    expect(tasks[0].text).toBeUndefined();
  });
  test('configKey absent on file → empty string on task', () => {
    const tasks = buildPluginLintTasks(
      input([{ path: '/a.ts', text: 'const x = 1;' }]),
      { configDirSet: new Set() },
    );
    expect(tasks).toHaveLength(1);
    expect(tasks[0].configKey).toBe('');
  });

  test('configKey known in set → passed through verbatim', () => {
    const tasks = buildPluginLintTasks(
      input([{ path: '/proj/pkg-a/a.ts', text: '', configKey: '/proj/pkg-a' }]),
      { configDirSet: new Set(['/proj/pkg-a']) },
    );
    expect(tasks).toHaveLength(1);
    expect(tasks[0].configKey).toBe('/proj/pkg-a');
  });

  test('configKey unknown → onUnknownConfigKey fires, key still passed through', () => {
    const warnings: Array<{ filePath: string; configKey: string }> = [];
    const tasks = buildPluginLintTasks(
      input([{ path: '/x.ts', text: '', configKey: '/nowhere' }]),
      {
        configDirSet: new Set(['/proj/pkg-a']),
        onUnknownConfigKey: (filePath, configKey) =>
          warnings.push({ filePath, configKey }),
      },
    );
    expect(tasks).toHaveLength(1);
    // The unknown key still flows through; the worker is the source of
    // truth for whether it's actually an invariant violation. The host
    // hook is purely for surfacing a clearer log line.
    expect(tasks[0].configKey).toBe('/nowhere');
    expect(warnings).toEqual([{ filePath: '/x.ts', configKey: '/nowhere' }]);
  });

  test('configKey unknown without onUnknownConfigKey → still no throw', () => {
    const tasks = buildPluginLintTasks(
      input([{ path: '/x.ts', text: '', configKey: '/nowhere' }]),
      { configDirSet: new Set() },
    );
    expect(tasks[0].configKey).toBe('/nowhere');
  });

  test('shared rules / collectFixes / suggestionsMode propagate to every task', () => {
    const tasks = buildPluginLintTasks(
      input(
        [
          { path: '/a.ts', text: '' },
          { path: '/b.ts', text: '' },
        ],
        {
          rules: {
            'uc/no-null': { options: [{ checkStrictEquality: true }] },
            'uc/prefer-array-some': { options: [] },
          },
          collectFixes: true,
          suggestionsMode: 'eager',
        },
      ),
      { configDirSet: new Set() },
    );
    expect(tasks).toHaveLength(2);
    for (const t of tasks) {
      expect(t.collectFixes).toBe(true);
      expect(t.suggestionsMode).toBe('eager');
      expect(Object.keys(t.rules).sort()).toEqual([
        'uc/no-null',
        'uc/prefer-array-some',
      ]);
      expect(t.rules['uc/no-null'].options).toEqual([
        { checkStrictEquality: true },
      ]);
    }
  });

  test('rules / collectFixes / suggestionsMode defaults are sensible', () => {
    const tasks = buildPluginLintTasks(
      { files: [{ path: '/a.ts', text: '' }], collectFixes: false },
      { configDirSet: new Set() },
    );
    expect(tasks[0].collectFixes).toBe(false);
    expect(tasks[0].suggestionsMode).toBe('off');
    expect(tasks[0].rules).toEqual({});
  });

  test('languageOptions and settings pass through opaquely', () => {
    const langOpts = { parserOptions: { ecmaVersion: 2024 as const } };
    const settings = { react: { version: '19.0.0' } };
    const tasks = buildPluginLintTasks(
      input([
        {
          path: '/a.tsx',
          text: '',
          languageOptions: langOpts,
          settings,
        },
      ]),
      { configDirSet: new Set() },
    );
    expect(tasks[0].languageOptions).toBe(langOpts);
    expect(tasks[0].settings).toBe(settings);
  });
});

describe('buildPluginLintResult', () => {
  test('projects to the 6-field wire shape, drops aggregate convenience fields', () => {
    const results: LintFileResult[] = [
      {
        filePath: '/a.ts',
        diagnostics: [
          {
            ruleName: 'uc/no-null',
            message: 'do not use null',
            startPos: 10,
            endPos: 14,
          },
        ],
        // Aggregate convenience fields: present on LintFileResult,
        // absent on the wire shape Go decodes. The projection MUST
        // drop them so we can't silently grow the wire contract by
        // accident.
        fixes: [{ range: [10, 14], text: 'undefined' }],
        suggestionsCount: 0,
        cancelled: false,
      },
    ];
    const projected = buildPluginLintResult(results);
    expect(projected.results).toHaveLength(1);
    const r = projected.results[0];
    // Exactly the 6 Go-visible fields, no more.
    expect(Object.keys(r).sort()).toEqual(
      [
        'cancelled',
        'diagnostics',
        'filePath',
        'parseError',
        'ruleErrors',
        'ruleTimes',
      ].sort(),
    );
    expect(r.filePath).toBe('/a.ts');
    expect(r.diagnostics).toHaveLength(1);
    expect(r.cancelled).toBe(false);
  });

  test('forwards parseError and ruleErrors when present', () => {
    const results: LintFileResult[] = [
      {
        filePath: '/broken.ts',
        diagnostics: [],
        fixes: [],
        suggestionsCount: 0,
        cancelled: false,
        parseError: 'parse: unexpected token',
        ruleErrors: [
          { rule: 'uc/no-null', message: 'create threw: x undefined' },
        ],
      },
    ];
    const projected = buildPluginLintResult(results);
    expect(projected.results[0].parseError).toBe('parse: unexpected token');
    expect(projected.results[0].ruleErrors).toEqual([
      { rule: 'uc/no-null', message: 'create threw: x undefined' },
    ]);
  });

  test('empty input yields empty results array (not undefined)', () => {
    expect(buildPluginLintResult([])).toEqual({ results: [] });
  });
});
