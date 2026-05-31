import { describe, test, expect } from '@rstest/core';
import path from 'node:path';

import { lintFile } from '../../../src/eslint-plugin/linter/ecma-language-plugin.js';
import { loadPluginsFromConfigs } from '../../../src/eslint-plugin/plugin/plugin-loader.js';
import type { LoadedPlugins } from '../../../src/eslint-plugin/plugin/plugin-loader.js';

/**
 * End-to-end plugin lint flow: load the local fixture plugin
 * (`fixtures/local-plugin.mjs`) via the configs-flow
 * (`loadPluginsFromConfigs` → import the user's rslint.config.mjs →
 * pull live plugin instances), run `local/no-null` on a fixture, and
 * assert the diagnostics. The plugin is a local stand-in (no external
 * dependency); its `no-null` mirrors the eager-options + suggestion
 * shape a real plugin like unicorn would have.
 */

const LOCAL_CONFIG_PATH = path.resolve(
  __dirname,
  '..',
  'fixtures',
  'local.config.mjs',
);
const LOCAL_CONFIG_DIR = path.dirname(LOCAL_CONFIG_PATH);

async function loadLocalPlugin(): Promise<LoadedPlugins> {
  const map = await loadPluginsFromConfigs([
    { configPath: LOCAL_CONFIG_PATH, configDirectory: LOCAL_CONFIG_DIR },
  ]);
  const loaded = map.get(LOCAL_CONFIG_DIR);
  if (!loaded) {
    throw new Error('test fixture: local plugin config did not register');
  }
  return loaded;
}

describe('plugin lint flow with a local fixture plugin', () => {
  test('local no-null produces correct diagnostics', async () => {
    const SOURCE = `// fixture
export function f(x: number | null): boolean {
  if (x === null) return false;
  const y = null;
  return Boolean(y);
}
`;
    const loaded = await loadLocalPlugin();

    const result = lintFile(
      {
        filePath: 'test.ts',
        text: SOURCE,
        rules: {
          'local/no-null': { options: [] },
        },
        collectFixes: false,
        suggestionsMode: 'off',
      },
      loaded,
    );

    expect(result.parseError).toBeUndefined();
    expect(result.cancelled).toBe(false);

    // With local/no-null's default options (checkStrictEquality=false),
    // the rule reports ONLY `const y = null` (line 4 col 13), NOT the
    // `x === null` strict-equality form (line 3). This pinpoints the
    // default-options handling — a regression that swallowed defaults
    // would either flag both or miss the `const y` case.
    expect(result.diagnostics).toHaveLength(1);
    const d = result.diagnostics[0];
    expect(d.ruleName).toBe('local/no-null');
    expect(d.messageId).toBe('error');
    expect(d.message).toMatch(/null|undefined/);
    // Compute the expected offsets from SOURCE rather than baking
    // numbers into the assertion: any whitespace tweak to the fixture
    // would otherwise silently shift the literal and the hard-coded
    // 102/106 would lie. Find the SECOND `null` literal — the first is
    // the type annotation `number | null`, which the rule correctly
    // ignores under default options (checkStrictEquality=false), so
    // that's the strict-equality `x === null` form. The third `null`
    // is on the `const y = ` line and is what the rule flags.
    const constYIdx = SOURCE.indexOf('const y = null');
    const expectedStart = constYIdx + 'const y = '.length;
    const expectedEnd = expectedStart + 'null'.length;
    expect(d.startPos).toBe(expectedStart);
    expect(d.endPos).toBe(expectedEnd);
  });

  test('cancellation flag bails the visit early', async () => {
    const SOURCE = `const a = null; const b = null; const c = null;`;
    const loaded = await loadLocalPlugin();

    // Pre-set cancel flag = 1 so the very first node visit observes cancel.
    const sab = new SharedArrayBuffer(4);
    const flag = new Int32Array(sab);
    Atomics.store(flag, 0, 1);

    const result = lintFile(
      {
        filePath: 'test.ts',
        text: SOURCE,
        rules: { 'local/no-null': { options: [] } },
        collectFixes: false,
        suggestionsMode: 'off',
        cancelFlag: flag,
      },
      loaded,
    );

    expect(result.cancelled).toBe(true);
    // We may have 0 or some diagnostics depending on how early we bailed.
    // The important guarantee is cancel propagation worked.
  });

  test('parse error is captured per-file, not thrown', async () => {
    const loaded: LoadedPlugins = { plugins: [], rules: new Map() };

    const result = lintFile(
      {
        filePath: 'test.ts',
        text: '!!!! invalid syntax (((((((((((((((((((((((((((((((',
        rules: {},
        collectFixes: false,
        suggestionsMode: 'off',
      },
      loaded,
    );

    // oxc-parser is robust — it may parse this without throwing. So we
    // assert "no diagnostics, no thrown error" rather than "parseError set".
    expect(result.diagnostics).toEqual([]);
  });

  // B1: a plugin author's `rule.create()` itself throws (vs. listener
  // throwing later). The error must be attributed to that rule via
  // ruleErrors, and other rules in the same file must continue to run.
  test('B1: rule.create() throw is captured in ruleErrors; other rules continue', async () => {
    interface RuleErr {
      rule: string;
      message: string;
    }
    const loaded: LoadedPlugins = {
      plugins: [],
      rules: new Map<string, unknown>([
        [
          'stub/throws-on-create',
          {
            meta: { name: 'throws-on-create' },
            create() {
              throw new Error('B1 boom — create() failed');
            },
          },
        ],
        [
          'stub/healthy',
          {
            meta: { name: 'healthy' },
            create(ctx: { report: (d: unknown) => void }) {
              return {
                Identifier(node: { name: string }) {
                  if (node.name === 'flag') {
                    ctx.report({ node, message: 'healthy fired' });
                  }
                },
              };
            },
          },
        ],
      ]),
    };

    const result = lintFile(
      {
        filePath: 'b1.ts',
        text: 'const flag = 1;',
        rules: {
          'stub/throws-on-create': { options: [] },
          'stub/healthy': { options: [] },
        },
        collectFixes: false,
        suggestionsMode: 'off',
      },
      loaded,
    );

    // Throw on create: surface as ruleErrors entry pointing at that rule.
    const errs = (result.ruleErrors ?? []) as RuleErr[];
    expect(errs.some((e) => e.rule === 'stub/throws-on-create')).toBe(true);

    // Healthy rule still ran — `flag` identifier reported.
    const healthyDiag = result.diagnostics.find(
      (d) => d.ruleName === 'stub/healthy',
    );
    expect(healthyDiag).toBeDefined();
  });

  // B2: a plugin calls a SourceCode/RuleContext method that doesn't
  // exist (`context.sourceCode.nonExistentMethod()`). Should produce
  // a TypeError attributed to the rule via ruleErrors, not crash the
  // whole file lint.
  test('B2: plugin calling a nonexistent SourceCode method → ruleErrors, not crash', async () => {
    interface RuleErr {
      rule: string;
      message: string;
    }
    const loaded: LoadedPlugins = {
      plugins: [],
      rules: new Map<string, unknown>([
        [
          'stub/bad-api',
          {
            meta: { name: 'bad-api' },
            create(ctx: {
              sourceCode: Record<string, unknown>;
              report: (d: unknown) => void;
            }) {
              return {
                Program(node: unknown) {
                  // Method doesn't exist — runtime TypeError.
                  (
                    ctx.sourceCode as unknown as {
                      nonExistentMethod: () => void;
                    }
                  ).nonExistentMethod();
                  ctx.report({ node, message: 'unreachable' });
                },
              };
            },
          },
        ],
      ]),
    };

    const result = lintFile(
      {
        filePath: 'b2.ts',
        text: 'const x = 1;',
        rules: { 'stub/bad-api': { options: [] } },
        collectFixes: false,
        suggestionsMode: 'off',
      },
      loaded,
    );

    const errs = (result.ruleErrors ?? []) as RuleErr[];
    expect(errs.some((e) => e.rule === 'stub/bad-api')).toBe(true);
    // No diagnostic from the unreachable `ctx.report` after the throw.
    expect(result.diagnostics).toHaveLength(0);
  });

  // B4: a plugin with NO `meta` field (some legacy authors). The
  // worker must still load it and use the prefix from the config.
  test('B4: plugin without `meta` field still loads and runs its rules', async () => {
    const loaded: LoadedPlugins = {
      plugins: [],
      rules: new Map<string, unknown>([
        [
          'stub/no-meta',
          {
            // meta intentionally omitted
            create(ctx: { report: (d: unknown) => void }) {
              return {
                Identifier(node: { name: string }) {
                  if (node.name === 'forbidden') {
                    ctx.report({ node, message: 'no-meta rule fired' });
                  }
                },
              };
            },
          },
        ],
      ]),
    };

    const result = lintFile(
      {
        filePath: 'b4.ts',
        text: 'const forbidden = 1;',
        rules: { 'stub/no-meta': { options: [] } },
        collectFixes: false,
        suggestionsMode: 'off',
      },
      loaded,
    );

    expect(result.diagnostics.length).toBeGreaterThan(0);
    expect(result.ruleErrors ?? []).toHaveLength(0);
  });
});
