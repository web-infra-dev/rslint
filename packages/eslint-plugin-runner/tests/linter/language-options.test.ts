/**
 * Regression tests for A12 (RuleContext.languageOptions) and A13
 * (parser jsx) — added 2026-05-13 after a multi-model code review
 * pointed out that:
 *
 *   - many real ESLint plugins (e.g. eslint-plugin-import,
 *     eslint-plugin-react) branch on
 *     `context.languageOptions.parserOptions.ecmaVersion` or
 *     `context.languageOptions.parserOptions.ecmaFeatures.jsx`.
 *     Before A12, `context.languageOptions` didn't exist, so these
 *     branches went down the wrong path.
 *
 *   - `parserOptions.ecmaFeatures.jsx === true` should enable JSX
 *     parsing even when the file extension is `.js` / `.ts` (a real
 *     React/Next.js convention). Before A13, the flag rode the wire
 *     but `parseSync` ignored it, so `.js` files with JSX content
 *     ParseError'd silently.
 *
 * v10 alignment note: the legacy `context.parserOptions` top-level
 * proxy was removed when rslint moved its surface from v8/v9 mix to
 * v10. Plugins now read `context.languageOptions.parserOptions`.
 */

import { describe, test, expect } from '@rstest/core';

import { lintFile } from '../../src/linter/ecma-language-plugin.js';
import type { LoadedPlugins } from '../../src/plugin/plugin-loader.js';
import type { RuleContext } from '../../src/linter/context.js';

// ── Stub plugin that records the ctx fields it observes ───────────

interface RecordedContext {
  languageOptions?: unknown;
}

function makeStubPlugin(record: RecordedContext) {
  return {
    rules: {
      'record-context': {
        meta: { name: 'record-context' },
        create(ctx: RuleContext) {
          record.languageOptions = ctx.languageOptions;
          return {
            // Trigger once so the listener fires — Program is always present.
            Program() {
              ctx.report({
                node: null as unknown as never,
                loc: { line: 1, column: 0 },
                message: 'recorded',
              });
            },
          };
        },
      },
    },
  };
}

function makeLoadedPlugins(record: RecordedContext): LoadedPlugins {
  const plugin = makeStubPlugin(record);
  return {
    plugins: [],
    rules: new Map<string, unknown>([
      ['stub/record-context', plugin.rules['record-context']],
    ]),
  };
}

// ── A12 — languageOptions exposed on ctx (v10 surface) ────────────

describe('A12: languageOptions exposed on RuleContext', () => {
  test('ctx.languageOptions mirrors the wire payload', () => {
    const recorded: RecordedContext = {};
    const loaded = makeLoadedPlugins(recorded);

    lintFile(
      {
        filePath: 'test.js',
        text: 'const x = 1;\n',
        languageOptions: {
          // ESLint v10: ecmaVersion / sourceType / globals are top-level
          // of languageOptions, not nested under parserOptions.
          ecmaVersion: 2022,
          sourceType: 'module',
          globals: { __DEV__: 'readonly' },
        },
        rules: { 'stub/record-context': { options: [] } },
        collectFixes: false,
        suggestionsMode: 'off',
      },
      loaded,
    );

    expect(recorded.languageOptions).toBeTruthy();
    const lo = recorded.languageOptions as Record<string, unknown>;
    expect(lo.ecmaVersion).toBe(2022);
    expect(lo.sourceType).toBe('module');
    expect(lo.globals).toEqual({ __DEV__: 'readonly' });
  });

  test('ctx.languageOptions.parserOptions carries ecmaFeatures.jsx', () => {
    const recorded: RecordedContext = {};
    const loaded = makeLoadedPlugins(recorded);

    lintFile(
      {
        filePath: 'test.js',
        text: 'const x = 1;\n',
        languageOptions: {
          ecmaVersion: 'latest',
          sourceType: 'script',
          parserOptions: {
            ecmaFeatures: { jsx: true },
          },
        },
        rules: { 'stub/record-context': { options: [] } },
        collectFixes: false,
        suggestionsMode: 'off',
      },
      loaded,
    );

    const lo = recorded.languageOptions as {
      ecmaVersion: string;
      sourceType: string;
      parserOptions: {
        ecmaFeatures: { jsx: boolean };
      };
    };
    expect(lo.ecmaVersion).toBe('latest');
    expect(lo.sourceType).toBe('script');
    expect(lo.parserOptions.ecmaFeatures.jsx).toBe(true);
  });

  test('no user languageOptions → stable empty default (never undefined)', () => {
    const recorded: RecordedContext = {};
    const loaded = makeLoadedPlugins(recorded);

    lintFile(
      {
        filePath: 'test.js',
        text: 'const x = 1;\n',
        // languageOptions intentionally absent
        rules: { 'stub/record-context': { options: [] } },
        collectFixes: false,
        suggestionsMode: 'off',
      },
      loaded,
    );

    // The wrapper must be non-undefined so plugins can safely
    // access nested fields without crashing.
    expect(recorded.languageOptions).toBeTruthy();
    expect(typeof recorded.languageOptions).toBe('object');
  });
});

// ── A13 — JSX parsing flag on .js / .ts files ─────────────────────

describe('A13: parserOptions.ecmaFeatures.jsx promotes parser lang', () => {
  test('.js + ecmaFeatures.jsx=true parses JSX content without parseError', () => {
    const recorded: RecordedContext = {};
    const loaded = makeLoadedPlugins(recorded);

    const result = lintFile(
      {
        // .js extension — without the jsx promotion, oxc-parser would
        // refuse the `<Foo />` and return a parseError.
        filePath: 'component.js',
        text:
          "import React from 'react';\n" +
          "export const App = () => <Foo bar='baz' />;\n",
        languageOptions: {
          sourceType: 'module',
          parserOptions: {
            ecmaFeatures: { jsx: true },
          },
        },
        rules: { 'stub/record-context': { options: [] } },
        collectFixes: false,
        suggestionsMode: 'off',
      },
      loaded,
    );

    expect(result.parseError).toBeUndefined();
  });

  test('.ts + ecmaFeatures.jsx=true parses TSX content without parseError', () => {
    const recorded: RecordedContext = {};
    const loaded = makeLoadedPlugins(recorded);

    const result = lintFile(
      {
        filePath: 'component.ts',
        text:
          'type Props = { name: string };\n' +
          'const App: React.FC<Props> = ({ name }) => <h1>{name}</h1>;\n',
        languageOptions: {
          sourceType: 'module',
          parserOptions: {
            ecmaFeatures: { jsx: true },
          },
        },
        rules: { 'stub/record-context': { options: [] } },
        collectFixes: false,
        suggestionsMode: 'off',
      },
      loaded,
    );

    expect(result.parseError).toBeUndefined();
  });

  test('.jsx files parse JSX regardless of flag', () => {
    // Sanity check that file-extension inference still works when
    // ecmaFeatures.jsx is unset — .jsx is JSX by default.
    const recorded: RecordedContext = {};
    const loaded = makeLoadedPlugins(recorded);

    const result = lintFile(
      {
        filePath: 'component.jsx',
        text: 'const App = () => <Foo />;\n',
        rules: { 'stub/record-context': { options: [] } },
        collectFixes: false,
        suggestionsMode: 'off',
      },
      loaded,
    );

    expect(result.parseError).toBeUndefined();
  });
});

// ── A15 — listener throw produces structured ruleErrors entry ─────

describe('A15: listener throw is attributed to its rule in ruleErrors', () => {
  test('throwing listener writes one ruleErrors entry per rule', () => {
    // A fake plugin whose Identifier listener throws on every node.
    // Previously the throw only reached stderr (dedup'd by selector),
    // and the result.ruleErrors was empty — making it hard to tell
    // which rule actually failed. With A15, the per-file result
    // carries a structured ruleErrors entry tagged with the rule
    // name so the user can see the failure in the diagnostic
    // stream / LSP UI.
    const plugin = {
      rules: {
        'always-throws': {
          meta: { name: 'always-throws' },
          create() {
            return {
              Identifier(_node: unknown) {
                throw new Error('synthetic listener failure');
              },
            };
          },
        },
      },
    };
    const loaded = {
      plugins: [],
      rules: new Map<string, unknown>([
        ['stub/always-throws', plugin.rules['always-throws']],
      ]),
    };

    const result = lintFile(
      {
        filePath: 'test.js',
        text: 'const x = 1;\nconst y = 2;\n',
        rules: { 'stub/always-throws': { options: [] } },
        collectFixes: false,
        suggestionsMode: 'off',
      },
      loaded,
    );

    // The listener fires on every Identifier in the file — `x` and `y`
    // are both Identifiers. We expect at least one ruleErrors entry
    // (the dedup-stderr Set inside the runner doesn't gate ruleErrors,
    // so each throw lands separately).
    expect(result.ruleErrors).toBeDefined();
    const errs = result.ruleErrors ?? [];
    expect(errs.length).toBeGreaterThan(0);
    // Every entry must carry the rule name (NOT an anonymous selector).
    for (const e of errs) {
      expect(e.rule).toBe('stub/always-throws');
      expect(e.message).toContain('synthetic listener failure');
    }
  });

  test('two rules with the same selector get distinct ruleErrors entries', () => {
    // Critical for type-aware-style failures: if `@typescript-eslint/foo`
    // and `@typescript-eslint/bar` both register `Identifier` and BOTH
    // throw because they access `parserServices.program` (undefined),
    // the per-file ruleErrors should show two entries — one per rule
    // — instead of a single selector-tagged stderr line where the
    // user can't tell which one to disable.
    const makePlugin = (name: string) => ({
      meta: { name },
      create() {
        return {
          Identifier() {
            throw new Error(`failure from ${name}`);
          },
        };
      },
    });
    const loaded = {
      plugins: [],
      rules: new Map<string, unknown>([
        ['stub/rule-a', makePlugin('rule-a')],
        ['stub/rule-b', makePlugin('rule-b')],
      ]),
    };

    const result = lintFile(
      {
        filePath: 'test.js',
        text: 'const x = 1;\n',
        rules: {
          'stub/rule-a': { options: [] },
          'stub/rule-b': { options: [] },
        },
        collectFixes: false,
        suggestionsMode: 'off',
      },
      loaded,
    );

    const errs = result.ruleErrors ?? [];
    const seenRules = new Set(errs.map((e) => e.rule));
    expect(seenRules.has('stub/rule-a')).toBe(true);
    expect(seenRules.has('stub/rule-b')).toBe(true);
  });
});

// ── A16: user globals must override ECMA built-in mode flags ──────
//
// `seedEcmaGlobals` layers the built-in set (Array, Object, parseInt, ...)
// into the global scope with mode `readonly`. `seedGlobals` then
// layers the user's `languageOptions.globals` on top — and v9+ ESLint
// honors the user's chosen mode for those names even when they
// collide with built-ins.  e.g. a project that intentionally
// monkey-patches `Array` writes `{ Array: 'writable' }` and expects
// `no-global-assign` to stay quiet.
//
// Pre-fix `ensureGlobal` short-circuited on existing entries without
// updating mode, so user-supplied `'writable'` had no effect on
// built-ins. Empirically confirmed against ESLint v9 (Linter.verify
// returns writeable=true for the same setup).
describe('A16: user globals override built-in mode flags', () => {
  function readArrayMode(
    globalsOpt: Record<string, 'readonly' | 'writable' | 'off'>,
  ): {
    writeable: boolean | undefined;
    eslintImplicitGlobalSetting: string | undefined;
  } {
    let observed: {
      writeable: boolean | undefined;
      eslintImplicitGlobalSetting: string | undefined;
    } = { writeable: undefined, eslintImplicitGlobalSetting: undefined };
    const loaded: LoadedPlugins = {
      plugins: [],
      rules: new Map<string, unknown>([
        [
          'stub/probe',
          {
            meta: { name: 'probe' },
            create(ctx: RuleContext) {
              return {
                Program() {
                  const sm = ctx.sourceCode.scopeManager as {
                    globalScope?: {
                      variables?: Array<{
                        name: string;
                        writeable?: boolean;
                        eslintImplicitGlobalSetting?: string;
                      }>;
                    };
                  } | null;
                  const arr = sm?.globalScope?.variables?.find(
                    (v) => v.name === 'Array',
                  );
                  observed = {
                    writeable: arr?.writeable,
                    eslintImplicitGlobalSetting:
                      arr?.eslintImplicitGlobalSetting,
                  };
                },
              };
            },
          },
        ],
      ]),
    };
    lintFile(
      {
        filePath: 'globals.js',
        text: 'const x = Array;',
        rules: { 'stub/probe': { options: [] } },
        languageOptions: { globals: globalsOpt },
        collectFixes: false,
        suggestionsMode: 'off',
      },
      loaded,
    );
    return observed;
  }

  test('user globals.Array=writable → built-in readonly is upgraded to writable', () => {
    const result = readArrayMode({ Array: 'writable' });
    expect(result.writeable).toBe(true);
    expect(result.eslintImplicitGlobalSetting).toBe('writable');
  });

  test('user globals.Array=readonly → built-in already readonly stays readonly (idempotent)', () => {
    const result = readArrayMode({ Array: 'readonly' });
    expect(result.writeable).toBe(false);
    expect(result.eslintImplicitGlobalSetting).toBe('readonly');
  });
});

// ── A17: ESLint v10 alignment — ecmaVersion/sourceType at top level ──
//
// v10's flat-config spec lifts `ecmaVersion` and `sourceType` from
// `parserOptions` to the top of `languageOptions`. Pre-fix the runner
// only read the legacy `parserOptions.{ecmaVersion,sourceType}`
// positions; a v10-conformant user writing them at the top level got
// silently ignored. These tests pin the v10 positions.
describe('A17: v10 top-level ecmaVersion / sourceType', () => {
  test('languageOptions.sourceType=module accepts ES module imports', () => {
    // `import` statements are module-only syntax. The test pins
    // that the runner's parseSync call gets sourceType='module' from
    // the v10 top-level position — without it oxc would treat the
    // file as script and reject the import.
    const recorded: RecordedContext = {};
    const loaded = makeLoadedPlugins(recorded);
    const result = lintFile(
      {
        filePath: 'a.js',
        text: "import { x } from './x.js';\nx;\n",
        languageOptions: { sourceType: 'module' },
        rules: { 'stub/record-context': { options: [] } },
        collectFixes: false,
        suggestionsMode: 'off',
      },
      loaded,
    );
    expect(result.parseError).toBeUndefined();
  });

  test('languageOptions.sourceType=script reaches ctx.languageOptions', () => {
    // Value pass-through: ctx.languageOptions.sourceType must be
    // 'script' when the user wrote it at the v10 top-level position.
    const recorded: RecordedContext = {};
    const loaded = makeLoadedPlugins(recorded);
    lintFile(
      {
        filePath: 'a.js',
        text: 'var x = 1;\n',
        languageOptions: { sourceType: 'script' },
        rules: { 'stub/record-context': { options: [] } },
        collectFixes: false,
        suggestionsMode: 'off',
      },
      loaded,
    );
    const lo = recorded.languageOptions as { sourceType?: string };
    expect(lo.sourceType).toBe('script');
  });

  test('languageOptions.ecmaVersion flows to scope-factory unchanged', () => {
    // We can't easily observe oxc's accepted ecmaVersion without
    // forcing a syntax that requires a specific version, but the
    // value DOES flow into ctx.languageOptions verbatim — any rule
    // can read `ctx.languageOptions.ecmaVersion` to branch.
    const recorded: RecordedContext = {};
    const loaded = makeLoadedPlugins(recorded);
    lintFile(
      {
        filePath: 'a.js',
        text: 'const x = 1;\n',
        languageOptions: { ecmaVersion: 2024 },
        rules: { 'stub/record-context': { options: [] } },
        collectFixes: false,
        suggestionsMode: 'off',
      },
      loaded,
    );
    const lo = recorded.languageOptions as { ecmaVersion?: number };
    expect(lo.ecmaVersion).toBe(2024);
  });
});
