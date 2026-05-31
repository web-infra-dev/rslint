/**
 * Unit tests for the shared compat-batch boundary helpers used by
 * both the CLI host (engine.ts) and the LSP host (CompatPool.ts).
 *
 * `buildCompatTasksByConfigKey` forwards each file's `configKey`
 * verbatim — the worker picks the right `LoadedPlugins` from its
 * per-config map via that key. The helper here is responsible for:
 *
 *   - emitting `configKey` on every task (empty string when absent),
 *   - firing `onUnknownConfigKey` for hosts that want a clearer log
 *     before the worker's internal-error parseError lands,
 *   - propagating the shared `rules` / `collectFixes` /
 *     `suggestionsMode` block to every task,
 *   - forwarding `languageOptions` / `settings` opaquely to the worker.
 */

import { describe, test, expect } from '@rstest/core';

import {
  buildCompatTasksByConfigKey,
  buildCompatBatchResult,
  type CompatBatchInput,
} from '../../../src/eslint-plugin/plugin/compat-task-builder.js';
import type { LintFileResult } from '../../../src/eslint-plugin/linter/ecma-language-plugin.js';

function input(
  files: CompatBatchInput['files'],
  opts: Partial<CompatBatchInput> = {},
): CompatBatchInput {
  return {
    files,
    rules: opts.rules ?? { 'uc/no-null': { options: [] } },
    collectFixes: opts.collectFixes,
    suggestionsMode: opts.suggestionsMode,
  };
}

describe('buildCompatTasksByConfigKey', () => {
  test('configKey absent on file → empty string on task', () => {
    const tasks = buildCompatTasksByConfigKey(
      input([{ path: '/a.ts', text: 'const x = 1;' }]),
      { configDirSet: new Set() },
    );
    expect(tasks).toHaveLength(1);
    expect(tasks[0].configKey).toBe('');
  });

  test('configKey known in set → passed through verbatim', () => {
    const tasks = buildCompatTasksByConfigKey(
      input([{ path: '/proj/pkg-a/a.ts', text: '', configKey: '/proj/pkg-a' }]),
      { configDirSet: new Set(['/proj/pkg-a']) },
    );
    expect(tasks).toHaveLength(1);
    expect(tasks[0].configKey).toBe('/proj/pkg-a');
  });

  test('configKey unknown → onUnknownConfigKey fires, key still passed through', () => {
    const warnings: Array<{ filePath: string; configKey: string }> = [];
    const tasks = buildCompatTasksByConfigKey(
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
    const tasks = buildCompatTasksByConfigKey(
      input([{ path: '/x.ts', text: '', configKey: '/nowhere' }]),
      { configDirSet: new Set() },
    );
    expect(tasks[0].configKey).toBe('/nowhere');
  });

  test('shared rules / collectFixes / suggestionsMode propagate to every task', () => {
    const tasks = buildCompatTasksByConfigKey(
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
    const tasks = buildCompatTasksByConfigKey(
      { files: [{ path: '/a.ts', text: '' }] },
      { configDirSet: new Set() },
    );
    expect(tasks[0].collectFixes).toBe(false);
    expect(tasks[0].suggestionsMode).toBe('off');
    expect(tasks[0].rules).toEqual({});
  });

  test('languageOptions and settings pass through opaquely', () => {
    const langOpts = { parserOptions: { ecmaVersion: 2024 as const } };
    const settings = { react: { version: '19.0.0' } };
    const tasks = buildCompatTasksByConfigKey(
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

describe('buildCompatBatchResult', () => {
  test('projects to the 5-field wire shape, drops aggregate convenience fields', () => {
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
    const projected = buildCompatBatchResult(results);
    expect(projected.results).toHaveLength(1);
    const r = projected.results[0];
    // Exactly the 5 Go-visible fields, no more.
    expect(Object.keys(r).sort()).toEqual(
      [
        'cancelled',
        'diagnostics',
        'filePath',
        'parseError',
        'ruleErrors',
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
    const projected = buildCompatBatchResult(results);
    expect(projected.results[0].parseError).toBe('parse: unexpected token');
    expect(projected.results[0].ruleErrors).toEqual([
      { rule: 'uc/no-null', message: 'create threw: x undefined' },
    ]);
  });

  test('empty input yields empty results array (not undefined)', () => {
    expect(buildCompatBatchResult([])).toEqual({ results: [] });
  });
});
