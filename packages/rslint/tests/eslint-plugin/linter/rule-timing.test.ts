/**
 * Per-rule timing collection (`collectTiming` → `result.ruleTimes`),
 * driven by Go's `--timing`. Assertions stay
 * magnitude-free (presence / shape / non-negativity only) so the suite
 * can't flake on scheduler noise.
 */
import { describe, test, expect } from '@rstest/core';

import { lintFile } from '../../../src/eslint-plugin/linter/ecma-language-plugin.js';
import type { LoadedPlugins } from '../../../src/eslint-plugin/plugin/plugin-loader.js';
import type { RuleContext } from '../../../src/eslint-plugin/linter/context.js';

function loadWithRules(
  rules: Record<string, (ctx: RuleContext) => unknown>,
): LoadedPlugins {
  return {
    plugins: [],
    rules: new Map<string, unknown>(
      Object.entries(rules).map(([name, create]) => [
        name,
        { meta: { name }, create },
      ]),
    ),
  };
}

describe('collectTiming', () => {
  test('off by default: no ruleTimes on the result', () => {
    const loaded = loadWithRules({
      'stub/probe': () => ({ Program() {} }),
    });
    const result = lintFile(
      {
        filePath: 'a.js',
        text: 'const a = 1;',
        rules: { 'stub/probe': { options: [] } },
        collectFixes: false,
        suggestionsMode: 'off',
      },
      loaded,
    );
    expect(result.ruleTimes).toBeUndefined();
  });

  test('records a non-negative time per executed rule', () => {
    const loaded = loadWithRules({
      'stub/listener': () => ({ Identifier() {} }),
      'stub/empty': () => ({}),
    });
    const result = lintFile(
      {
        filePath: 'a.js',
        text: 'const a = 1; const b = a;',
        rules: {
          'stub/listener': { options: [] },
          'stub/empty': { options: [] },
        },
        collectFixes: false,
        suggestionsMode: 'off',
        collectTiming: true,
      },
      loaded,
    );
    expect(result.parseError).toBeUndefined();
    expect(Object.keys(result.ruleTimes ?? {}).sort()).toEqual([
      'stub/empty',
      'stub/listener',
    ]);
    for (const ms of Object.values(result.ruleTimes ?? {})) {
      expect(typeof ms).toBe('number');
      expect(ms).toBeGreaterThanOrEqual(0);
      expect(Number.isFinite(ms)).toBe(true);
    }
  });

  test('throwing create and throwing listener still get timed', () => {
    const loaded = loadWithRules({
      'stub/create-throws': () => {
        throw new Error('boom');
      },
      'stub/listener-throws': () => ({
        Program() {
          throw new Error('bang');
        },
      }),
    });
    const result = lintFile(
      {
        filePath: 'a.js',
        text: 'const a = 1;',
        rules: {
          'stub/create-throws': { options: [] },
          'stub/listener-throws': { options: [] },
        },
        collectFixes: false,
        suggestionsMode: 'off',
        collectTiming: true,
      },
      loaded,
    );
    expect(Object.keys(result.ruleTimes ?? {}).sort()).toEqual([
      'stub/create-throws',
      'stub/listener-throws',
    ]);
    // Both failures are still reported as rule errors.
    expect((result.ruleErrors ?? []).map((e) => e.rule).sort()).toEqual([
      'stub/create-throws',
      'stub/listener-throws',
    ]);
  });
});
