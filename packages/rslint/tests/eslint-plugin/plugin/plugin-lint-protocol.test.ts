import { describe, test, expect } from 'rstack/test';
import {
  buildPluginLintTasks,
  buildPluginLintResult,
  type ResolvedEslintPluginLintRequest,
} from '../../../src/eslint-plugin/plugin/plugin-lint-protocol.js';
import type { LintFileResult } from '../../../src/eslint-plugin/linter/ecma-language-plugin.js';
import { resolvePluginAttachments } from '../../../src/eslint-plugin/plugin/attachments.js';

describe('binary plugin frame boundary', () => {
  test('resolves owned bytes and capabilities without decoding either', () => {
    const owned = new Uint8Array([0, 255, 1]),
      capability = { lease: 7, offset: 3, length: 2 };
    const req = {
      collectFixes: true,
      files: [
        { path: 'a.ts', frameAttachment: 0, configKey: 'config' },
        { path: 'b.ts', frame: 'AP8B' },
        { path: 'c.ts', frameAttachment: 1 },
      ],
    };
    const resolved = resolvePluginAttachments(req, [owned, capability]);
    expect(resolved.files[0].frame).toBe(owned);
    expect(resolved.files[1].frame).toEqual(Buffer.from(owned));
    expect(resolved.files[2].frame).toBe(capability);
    expect(req.files[0].frameAttachment).toBe(0);
  });
  test.each([
    {},
    { text: '' },
    { textAttachment: 0 },
    { frameAttachment: -1 },
    { frameAttachment: 0.5 },
    { frameAttachment: 1 },
    { frameAttachment: 0, frame: 'AP8B' },
    { frame: '***=' },
    { frame: 'AP8' },
    { frame: 'AB==' },
    { sharedSource: {} },
    { frameAttachment: 0, sourceRange: {} },
  ])('rejects invalid or obsolete input: %j', (input) => {
    expect(() =>
      resolvePluginAttachments(
        { collectFixes: false, files: [{ path: 'a.ts', ...input }] } as never,
        [new Uint8Array([1])],
      ),
    ).toThrow();
  });
  test('rejects text attachments and unused attachments', () => {
    expect(() =>
      resolvePluginAttachments(
        { collectFixes: false, files: [{ path: 'a.ts', frameAttachment: 0 }] },
        ['source'],
      ),
    ).toThrow();
    expect(() =>
      resolvePluginAttachments(
        { collectFixes: false, files: [{ path: 'a.ts', frame: 'AP8B' }] },
        [new Uint8Array([1])],
      ),
    ).toThrow();
  });
  test('rejects repeated attachment indices', () => {
    expect(() =>
      resolvePluginAttachments(
        {
          collectFixes: false,
          files: [
            { path: 'a.ts', frameAttachment: 0 },
            { path: 'b.ts', frameAttachment: 0 },
          ],
        },
        [new Uint8Array([1]), new Uint8Array([1])],
      ),
    ).toThrow();
  });
});

function input(
  files: ResolvedEslintPluginLintRequest['files'],
  opts: Partial<ResolvedEslintPluginLintRequest> = {},
): ResolvedEslintPluginLintRequest {
  return { files, collectFixes: false, ...opts };
}

describe('buildPluginLintTasks', () => {
  test('rejects an unresolved wire range instead of reading disk', () => {
    const request: EslintPluginLintRequest = {
      files: [{ path: '/missing.ts', frameAttachment: 0 }],
      collectFixes: false,
    };
    expect(() =>
      // @ts-expect-error A wire request must pass through the attachment adapter.
      buildPluginLintTasks(request, { configDirSet: new Set() }),
    ).toThrow('unresolved plugin frame');
  });
  test('forwards native source capabilities without decoding or reading files', () => {
    const frame = { lease: 1, offset: 0, length: 12 };
    const tasks = buildPluginLintTasks(
      input([{ path: '/missing.ts', frame }]),
      { configDirSet: new Set() },
    );
    expect(tasks[0].frame).toBe(frame);
    expect(tasks[0].text).toBeUndefined();
  });
  test('configKey absent on file → empty string on task', () => {
    const tasks = buildPluginLintTasks(
      input([{ path: '/a.ts', frame: new Uint8Array([1]) }]),
      { configDirSet: new Set() },
    );
    expect(tasks).toHaveLength(1);
    expect(tasks[0].configKey).toBe('');
  });

  test('configKey known in set → passed through verbatim', () => {
    const tasks = buildPluginLintTasks(
      input([
        {
          path: '/proj/pkg-a/a.ts',
          frame: new Uint8Array([1]),
          configKey: '/proj/pkg-a',
        },
      ]),
      { configDirSet: new Set(['/proj/pkg-a']) },
    );
    expect(tasks).toHaveLength(1);
    expect(tasks[0].configKey).toBe('/proj/pkg-a');
  });

  test('configKey unknown → onUnknownConfigKey fires, key still passed through', () => {
    const warnings: Array<{ filePath: string; configKey: string }> = [];
    const tasks = buildPluginLintTasks(
      input([
        { path: '/x.ts', frame: new Uint8Array([1]), configKey: '/nowhere' },
      ]),
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
      input([
        { path: '/x.ts', frame: new Uint8Array([1]), configKey: '/nowhere' },
      ]),
      { configDirSet: new Set() },
    );
    expect(tasks[0].configKey).toBe('/nowhere');
  });

  test('shared rules / collectFixes / suggestionsMode propagate to every task', () => {
    const tasks = buildPluginLintTasks(
      input(
        [
          { path: '/a.ts', frame: new Uint8Array([1]) },
          { path: '/b.ts', frame: new Uint8Array([1]) },
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
      {
        files: [{ path: '/a.ts', frame: new Uint8Array([1]) }],
        collectFixes: false,
      },
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
          frame: new Uint8Array([1]),
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
