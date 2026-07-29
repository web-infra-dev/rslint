/**
 * Regression coverage for `resolved-options-cache.ts` — restoring
 * cross-file `context.options` identity stability for a static rule
 * config, scoped per `configKey`/rule/options-content. See the
 * module doc for the ESLint-parity rationale.
 */
import { describe, test, expect } from '@rstest/core';

import { lintFile } from '../../../src/eslint-plugin/linter/ecma-language-plugin.js';
import type { LoadedPlugins } from '../../../src/eslint-plugin/plugin/plugin-loader.js';
import type { RuleContext } from '../../../src/eslint-plugin/linter/context.js';

function loadWithRule(create: (ctx: RuleContext) => unknown): LoadedPlugins {
  return {
    plugins: [],
    rules: new Map<string, unknown>([
      ['stub/probe', { meta: { name: 'probe' }, create }],
    ]),
  };
}

describe('resolveOptionsCached: cross-file context.options identity', () => {
  test('same configKey + rule + options content → same context.options reference across files', () => {
    const seen: unknown[] = [];
    const loaded = loadWithRule((ctx) => {
      seen.push(ctx.options[0]);
      return {};
    });

    for (const filePath of ['a.js', 'b.js']) {
      lintFile(
        {
          filePath,
          text: 'const x = 1;',
          configKey: 'cfg-dir-1',
          rules: { 'stub/probe': { options: [{ allow: ['Array'] }] } },
          collectFixes: false,
          suggestionsMode: 'off',
        },
        loaded,
      );
    }

    expect(seen.length).toBe(2);
    expect(seen[0]).toBe(seen[1]);
  });

  test('different configKey → distinct context.options reference', () => {
    const seen: unknown[] = [];
    const loaded = loadWithRule((ctx) => {
      seen.push(ctx.options[0]);
      return {};
    });

    for (const configKey of ['cfg-dir-a', 'cfg-dir-b']) {
      lintFile(
        {
          filePath: 'x.js',
          text: 'const x = 1;',
          configKey,
          rules: { 'stub/probe': { options: [{ allow: ['Array'] }] } },
          collectFixes: false,
          suggestionsMode: 'off',
        },
        loaded,
      );
    }

    expect(seen[0]).not.toBe(seen[1]);
  });

  test('different options content under the same configKey → distinct reference', () => {
    const seen: unknown[] = [];
    const loaded = loadWithRule((ctx) => {
      seen.push(ctx.options[0]);
      return {};
    });

    for (const allow of [['Array'], ['Object']]) {
      lintFile(
        {
          filePath: 'y.js',
          text: 'const x = 1;',
          configKey: 'cfg-dir-2',
          rules: { 'stub/probe': { options: [{ allow }] } },
          collectFixes: false,
          suggestionsMode: 'off',
        },
        loaded,
      );
    }

    expect(seen[0]).not.toBe(seen[1]);
  });

  test('no configKey (ad-hoc caller) → caching opts out, no crash', () => {
    const seen: unknown[] = [];
    const loaded = loadWithRule((ctx) => {
      seen.push(ctx.options[0]);
      return {};
    });

    for (let i = 0; i < 2; i++) {
      lintFile(
        {
          filePath: 'z.js',
          text: 'const x = 1;',
          rules: { 'stub/probe': { options: [{ allow: ['Array'] }] } },
          collectFixes: false,
          suggestionsMode: 'off',
        },
        loaded,
      );
    }

    expect(seen.length).toBe(2);
    // Content is still equal — just not memoized/shared by reference.
    expect(seen[0]).toEqual(seen[1]);
    expect(seen[0]).not.toBe(seen[1]);
  });

  test('mutating a cached options object throws instead of leaking to the next file', () => {
    const loaded = loadWithRule((ctx) => ({
      Program() {
        const opt = ctx.options[0] as { allow: string[] };
        // A rule pushing into its own options array — tolerated by
        // real ESLint only because nothing re-reads the shared
        // default afterward; here the object is reused by the next
        // file hitting the same cache entry, so the frozen array
        // must reject the mutation instead of silently corrupting
        // state for that next file.
        opt.allow.push('Object');
      },
    }));

    const result = lintFile(
      {
        filePath: 'mutate.js',
        text: 'const x = 1;',
        configKey: 'cfg-dir-3',
        rules: { 'stub/probe': { options: [{ allow: ['Array'] }] } },
        collectFixes: false,
        suggestionsMode: 'off',
      },
      loaded,
    );

    const errs = result.ruleErrors ?? [];
    expect(errs.length).toBeGreaterThan(0);
    expect(errs[0]?.rule).toBe('stub/probe');
  });
});
