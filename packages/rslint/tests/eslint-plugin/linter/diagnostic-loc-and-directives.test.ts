/**
 * Regression tests for two diagnostic surfaces:
 *
 *   - `diagnostic-builder` position-form `descriptor.loc` (Position
 *     shape with `line` / `column` directly on `loc`) must default
 *     `column` to 0 — pre-fix `lineStartOffset(...) + undefined === NaN`
 *     shipped NaN positions onto the wire.
 *
 *   - `sourceCode.getDisableDirectives()` must surface `rslint-*`
 *     prefixed directives alongside `eslint-*`. The suppression
 *     engine and `getInlineConfigNodes` already treat the two
 *     prefixes as equivalent first-class kinds; the inspection API
 *     was the missing piece.
 */
import { describe, test, expect } from '@rstest/core';

import { lintFile } from '../../../src/eslint-plugin/linter/ecma-language-plugin.js';
import type { LoadedPlugins } from '../../../src/eslint-plugin/plugin/plugin-loader.js';
import type { RuleContext } from '../../../src/eslint-plugin/linter/context.js';

// ── Test helpers ──────────────────────────────────────────────────

function makeReportingPlugin(fire: (ctx: RuleContext) => void): LoadedPlugins {
  return {
    plugins: [],
    rules: new Map<string, unknown>([
      [
        'stub/probe',
        {
          meta: { name: 'probe' },
          create(ctx: RuleContext) {
            return {
              Program() {
                fire(ctx);
              },
            };
          },
        },
      ],
    ]),
  };
}

function defaultRequest(text: string) {
  return {
    filePath: 'probe.js',
    text,
    rules: { 'stub/probe': { options: [] } },
    collectFixes: false,
    suggestionsMode: 'off' as const,
  };
}

// ── P1 #7 — diagnostic-builder position-form loc ──────────────────

describe('diagnostic-builder: position-form descriptor.loc defaults', () => {
  test('plugin report({ loc: { line: 1 } }) without column → startPos === 0 (not NaN)', () => {
    const loaded = makeReportingPlugin((ctx) => {
      // Position form, intentionally omit `column` — matches the
      // ESLint v10 contract: missing `column` defaults to 0.
      ctx.report({
        loc: { line: 1 } as never,
        message: 'no column',
      });
    });

    const result = lintFile(defaultRequest('const x = 1;\n'), loaded);
    expect(result.parseError).toBeUndefined();
    expect(result.diagnostics).toHaveLength(1);
    const d = result.diagnostics[0];
    // Pre-fix: `lineStartOffset(lso, 1, ...) + undefined === NaN`,
    // then converted through `u16ToByte[NaN] === undefined`. Post-fix:
    // column defaults to 0, lineStartOffset returns 0 (line 1), startPos
    // === 0, endPos === 0.
    expect(Number.isNaN(d.startPos)).toBe(false);
    expect(d.startPos).toBe(0);
    expect(d.endPos).toBe(0);
  });

  test('plugin report({ loc: { line: 2, column: 4 } }) — both fields present, unchanged behavior', () => {
    const text = 'const x = 1;\nconst y = 2;\n';
    const loaded = makeReportingPlugin((ctx) => {
      ctx.report({
        loc: { line: 2, column: 4 } as never,
        message: 'normal position',
      });
    });

    const result = lintFile(defaultRequest(text), loaded);
    expect(result.parseError).toBeUndefined();
    expect(result.diagnostics).toHaveLength(1);
    // Line 2 starts at offset 13 (after the first line + \n). column 4
    // → byte offset 17. After u16→byte conversion (ASCII = identity)
    // startPos === 17. Pin the value to lock the contract.
    expect(result.diagnostics[0].startPos).toBe(17);
  });
});

// ── P1 #6 — out-of-range loc clamps to text length ────────────────

describe('ecma-language-plugin: diagnostic positions clamped to source text length', () => {
  test('report at line beyond EOF yields clamped startPos (not undefined)', () => {
    // Pre-fix: `lineStartOffset` clamped line, but `+ column` had no
    // upper bound and `u16ToByte[超界] === undefined` leaked onto the
    // wire. Post-fix: every u16→byte access goes through `clamp` so
    // the worst case is `startPos === sourceText.length` (one-past-end).
    const text = 'const x = 1;\n';
    const loaded = makeReportingPlugin((ctx) => {
      ctx.report({
        loc: { line: 9999, column: 5 } as never,
        message: 'past EOF',
      });
    });
    const result = lintFile(defaultRequest(text), loaded);
    expect(result.parseError).toBeUndefined();
    expect(result.diagnostics).toHaveLength(1);
    const d = result.diagnostics[0];
    expect(d.startPos).toBeDefined();
    expect(d.endPos).toBeDefined();
    expect(typeof d.startPos).toBe('number');
    expect(typeof d.endPos).toBe('number');
    // Clamped to total byte length — never above.
    expect(d.startPos).toBeLessThanOrEqual(text.length);
    expect(d.endPos).toBeLessThanOrEqual(text.length);
  });

  test('column beyond line length still produces in-range byte offset', () => {
    const text = 'abc\n'; // 4 chars total
    const loaded = makeReportingPlugin((ctx) => {
      ctx.report({
        loc: { line: 1, column: 999 } as never,
        message: 'column past line end',
      });
    });
    const result = lintFile(defaultRequest(text), loaded);
    expect(result.parseError).toBeUndefined();
    const d = result.diagnostics[0];
    // line 1 lineStart = 0, + 999 = 999 → clamped to text.length = 4.
    expect(d.startPos).toBe(4);
    expect(d.endPos).toBe(4);
  });

  test('negative loc.column (defensive) → clamped to 0', () => {
    const text = 'x;\n';
    const loaded = makeReportingPlugin((ctx) => {
      ctx.report({
        loc: { line: 1, column: -5 } as never,
        message: 'negative column',
      });
    });
    const result = lintFile(defaultRequest(text), loaded);
    expect(result.parseError).toBeUndefined();
    expect(result.diagnostics[0].startPos).toBe(0);
  });
});

// ── P1 #10 — getDisableDirectives sees rslint-* prefix ────────────

describe('sourceCode.getDisableDirectives: rslint-* prefix recognition', () => {
  test('returns rslint-disable + rslint-disable-line + rslint-disable-next-line + rslint-enable alongside eslint-*', () => {
    interface DirShape {
      type: 'disable' | 'enable' | 'disable-next-line' | 'disable-line';
      value: string;
    }
    const observed: { directives: DirShape[] } = { directives: [] };

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
                  // Capture from the inspection API, NOT from the
                  // suppression engine. Pre-fix, the switch in
                  // `source-code.ts` only recognised the eslint-*
                  // labels and the rslint-* directives were silently
                  // dropped from this array even though
                  // apply-disable-directives still honored them.
                  const dirs = ctx.sourceCode.getDisableDirectives();
                  observed.directives = dirs.directives.map((d) => ({
                    type: d.type,
                    value: d.value,
                  }));
                },
              };
            },
          },
        ],
      ]),
    };

    const src =
      '/* eslint-disable no-undef */\n' +
      '/* rslint-disable no-console */\n' +
      'foo; // rslint-disable-line no-undef\n' +
      '// rslint-disable-next-line no-debugger\n' +
      'bar;\n' +
      '/* rslint-enable no-console */\n';

    const result = lintFile(defaultRequest(src), loaded);
    expect(result.parseError).toBeUndefined();

    // Order is source order. Pin the full set so a regression that
    // drops one prefix is immediately visible.
    expect(observed.directives).toEqual([
      { type: 'disable', value: 'no-undef' },
      { type: 'disable', value: 'no-console' },
      { type: 'disable-line', value: 'no-undef' },
      { type: 'disable-next-line', value: 'no-debugger' },
      { type: 'enable', value: 'no-console' },
    ]);
  });

  test('mixed eslint-/rslint- with same rule id → both surface as disable directives', () => {
    interface DirShape {
      type: 'disable' | 'enable' | 'disable-next-line' | 'disable-line';
      value: string;
    }
    const observed: { directives: DirShape[] } = { directives: [] };
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
                  const d = ctx.sourceCode.getDisableDirectives();
                  observed.directives = d.directives.map((x) => ({
                    type: x.type,
                    value: x.value,
                  }));
                },
              };
            },
          },
        ],
      ]),
    };

    const src =
      '/* eslint-disable foo */\n' + '/* rslint-disable foo */\nconst x = 1;\n';
    const result = lintFile(defaultRequest(src), loaded);
    expect(result.parseError).toBeUndefined();
    // Both surface (not deduplicated by prefix — they're distinct
    // source nodes; the suppression engine independently merges them).
    expect(observed.directives).toEqual([
      { type: 'disable', value: 'foo' },
      { type: 'disable', value: 'foo' },
    ]);
  });
});
