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
import { describe, test, expect } from 'rstack/test';
import fs from 'node:fs';
import path from 'node:path';
import { createRequire } from 'node:module';
import { applyDisableDirectives } from '../../../../rslint/src/eslint-plugin/linter/apply-disable-directives.js';
import { buildLineStartOffsets } from '../../../../rslint/src/languages/js/positions.js';
import type { Comment } from '../../../../rslint/src/languages/js/token-builder.js';

import { lintFile } from '../project-frame.js';
import type { LoadedPlugins } from '../../../../rslint/src/eslint-plugin/plugin/plugin-loader.js';
import type { RuleContext } from '../../../../rslint/src/eslint-plugin/linter/context.js';

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
  for (const prefix of ['eslint', 'rslint']) {
    test(`${prefix}: both APIs expose named and blanket line disable/enable commands`, () => {
      let inspected = false;
      const loaded = makeReportingPlugin((ctx) => {
        const nodes = ctx.sourceCode.getInlineConfigNodes();
        const result = ctx.sourceCode.getDisableDirectives();
        expect(nodes).toHaveLength(4);
        expect(result.problems).toEqual([]);
        expect(
          result.directives.map(({ type, value, justification }) => ({
            type,
            value,
            justification,
          })),
        ).toEqual([
          { type: 'disable', value: 'stub/probe', justification: 'reason' },
          { type: 'enable', value: 'stub/probe', justification: '' },
          { type: 'disable', value: '', justification: '' },
          { type: 'enable', value: '', justification: '' },
        ]);
        expect(result.directives.map(({ node }) => node)).toEqual(nodes);
        inspected = true;
      });
      const source = [
        `// ${prefix}-disable stub/probe -- reason`,
        `// ${prefix}-enable stub/probe`,
        `// ${prefix}-disable`,
        `// ${prefix}-enable`,
        '// global ignored',
        `// ${prefix}-disable-foo`,
        'export {};',
      ].join('\n');
      const result = lintFile(defaultRequest(source), loaded);
      expect(result.parseError).toBeUndefined();
      expect(result.ruleErrors).toBeUndefined();
      expect(inspected).toBe(true);
    });
  }

  test('Unicorn reports blanket line disables like block disables', async () => {
    const { default: unicorn } = await import('eslint-plugin-unicorn');
    const name = 'unicorn/no-abusive-eslint-disable';
    const rule = unicorn.rules?.['no-abusive-eslint-disable'];
    if (!rule) throw new Error(`Missing test rule: ${name}`);
    const loaded: LoadedPlugins = {
      plugins: [],
      rules: new Map([[name, rule]]),
    };
    for (const prefix of ['eslint', 'rslint']) {
      for (const comment of [
        `// ${prefix}-disable`,
        `/* ${prefix}-disable */`,
      ]) {
        const result = lintFile(
          {
            ...defaultRequest(`export {};\n${comment}\n`),
            rules: { [name]: { options: [] } },
          },
          loaded,
        );
        expect(result.parseError).toBeUndefined();
        expect(result.ruleErrors).toBeUndefined();
        expect(result.diagnostics.map(({ messageId }) => messageId)).toEqual([
          'no-abusive-eslint-disable',
        ]);
      }
    }
  });

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

test('returns used entry locations instead of suppressed plugin diagnostics', () => {
  const loaded = makeReportingPlugin((ctx) => {
    ctx.report({ loc: { line: 2, column: 0 }, message: 'reported to Go' });
  });
  const result = lintFile(
    {
      ...defaultRequest('/* eslint-disable stub/probe */\nconst x = 1;\n'),
      collectDirectiveUsage: true,
    },
    loaded,
  );
  expect(result.diagnostics).toHaveLength(0);
  expect(result.usedDisableDirectives).toEqual({
    rules: ['stub/probe'],
    entries: [[0, 0, 0]],
  });
});

test('disable filtering and usage attribution match the shared ESLint corpus', () => {
  const require = createRequire(import.meta.url);
  const espree = require('espree') as {
    tokenize(text: string, options: object): { comments: Comment[] };
  };
  const coreRequire = createRequire(
    new URL('../../../../rslint/package.json', import.meta.url),
  );
  const { ConfigCommentParser } = coreRequire('@eslint/plugin-kit') as {
    ConfigCommentParser: new () => {
      parseDirective(
        value: string,
      ): { label: string; value: string } | undefined;
      parseListConfig(value: string): Record<string, boolean>;
    };
  };
  const parser = new ConfigCommentParser();
  interface Corpus {
    cases: Array<{
      name: string;
      source: string;
      rslintSource: string;
      reports: Array<{ ruleId: string; pos: number; suppressed: boolean }>;
      ignored?: string[];
      invalidRules?: string[];
      variants: Array<{
        severity: string;
        diagnostics: Array<{ line: number; column: number; message: string }>;
      }>;
    }>;
  }
  let reports = 0;
  let entries = 0;
  for (const file of ['source.json', 'upstream.json']) {
    const corpus = JSON.parse(
      fs.readFileSync(
        path.resolve(
          import.meta.dirname,
          '../../../../../internal/rule/testdata/eslint-directives',
          file,
        ),
        'utf8',
      ),
    ) as Corpus;
    expect(corpus.cases.length).toBeGreaterThan(0);
    for (const item of corpus.cases) {
      for (const source of [item.source, item.rslintSource]) {
        const comments = espree.tokenize(source, {
          ecmaVersion: 'latest',
          range: true,
          loc: true,
          comment: true,
        }).comments;
        const lineStarts = buildLineStartOffsets(source);
        for (const reverse of [false, true]) {
          const diagnosticInputs = item.reports.map((report, index) => ({
            ruleName: report.ruleId,
            startPos: report.pos,
            endPos: report.pos,
            message: String(index),
          }));
          if (reverse) diagnosticInputs.reverse();
          const filtered = applyDisableDirectives({
            comments,
            diagnostics: diagnosticInputs,
            lineStartOffsets: lineStarts,
            textLength: source.length,
            collectUsage: true,
          });
          const visible = new Set(
            filtered.diagnostics.map((d) => Number(d.message)),
          );
          item.reports.forEach((report, index) => {
            reports++;
            expect(visible.has(index), item.name).toBe(!report.suppressed);
          });
          const variant = item.variants.find((v) => v.severity !== 'off');
          const usedKeys = new Set(
            (filtered.usedDisableDirectives ?? []).map((usage) =>
              JSON.stringify([
                lineStarts[usage.line] + usage.column,
                usage.ruleId,
              ]),
            ),
          );
          if (variant) {
            for (const comment of comments) {
              if (comment.type !== 'Line' && comment.type !== 'Block') continue;
              const parsed = parser.parseDirective(comment.value);
              const label = parsed?.label.replace(/^rslint-/u, 'eslint-');
              if (
                !parsed ||
                ![
                  'eslint-disable',
                  'eslint-disable-line',
                  'eslint-disable-next-line',
                ].includes(label ?? '')
              )
                continue;
              if (label === 'eslint-disable' && comment.type === 'Line')
                continue;
              if (
                label === 'eslint-disable-line' &&
                comment.loc.start.line !== comment.loc.end.line
              )
                continue;
              const names = Object.keys(parser.parseListConfig(parsed.value));
              for (const name of names.length ? names : [null]) {
                if (
                  name !== null &&
                  (item.invalidRules?.includes(name) ||
                    item.ignored?.includes(name))
                )
                  continue;
                if (name === null && (item.ignored?.length ?? 0) > 0) continue;
                const unused = variant.diagnostics.some(
                  (d) =>
                    d.message.startsWith('Unused eslint-disable') &&
                    lineStarts[d.line - 1] + d.column - 1 ===
                      comment.range[0] &&
                    (name === null || d.message.includes(`'${name}'`)),
                );
                expect(
                  usedKeys.has(JSON.stringify([comment.range[0], name])),
                  item.name,
                ).toBe(!unused);
              }
            }
          }
          for (const usage of filtered.usedDisableDirectives ?? []) {
            entries++;
            const offset = lineStarts[usage.line] + usage.column;
            expect(
              comments.some((c) => c.range[0] === offset),
              item.name,
            ).toBe(true);
            if (variant) {
              expect(
                variant.diagnostics.some(
                  (d) =>
                    d.message.startsWith('Unused eslint-disable') &&
                    lineStarts[d.line - 1] + d.column - 1 === offset &&
                    (usage.ruleId === null ||
                      d.message.includes(`'${usage.ruleId}'`)),
                ),
                item.name,
              ).toBe(false);
            }
          }
          // Every original suppression-only input still works without summary collection.
          const withoutUsage = applyDisableDirectives({
            comments,
            diagnostics: diagnosticInputs,
            lineStartOffsets: lineStarts,
            textLength: source.length,
            collectUsage: false,
          });
          expect(withoutUsage.usedDisableDirectives).toBeUndefined();
          expect(withoutUsage.diagnostics).toEqual(filtered.diagnostics);
        }
      }
    }
  }
  expect(reports).toBe(10012);
  expect(entries).toBeGreaterThan(0);
});

test('suppression normalizes exceptional positions in both collection modes', () => {
  const source = 'x; /* eslint-disable a */\ny();';
  const end = source.indexOf('*/') + 2;
  const comments: Comment[] = [
    {
      type: 'Block',
      value: ' eslint-disable a ',
      range: [3, end],
      loc: { start: { line: 1, column: 3 }, end: { line: 1, column: end } },
    },
  ];
  for (const collectUsage of [false, true]) {
    for (const [startPos, suppressed] of [
      [0, false],
      [-0, false],
      [-1, false],
      [2.9, false],
      [3, true],
      [3.9, true],
      [source.length, true],
      [Number.MAX_SAFE_INTEGER, true],
      [NaN, false],
      [Infinity, false],
      [-Infinity, false],
    ] as const) {
      const result = applyDisableDirectives({
        comments,
        diagnostics: [
          { ruleName: 'a', startPos, endPos: startPos, message: 'test' },
        ],
        lineStartOffsets: buildLineStartOffsets(source),
        textLength: source.length,
        collectUsage,
      });
      expect(result.diagnostics).toHaveLength(suppressed ? 0 : 1);
      if (collectUsage) {
        expect(result.usedDisableDirectives).toHaveLength(suppressed ? 1 : 0);
      } else {
        expect(result.usedDisableDirectives).toBeUndefined();
      }
    }
  }
});

test('line comments support block disable and enable in both collection modes', () => {
  for (const prefix of ['eslint', 'rslint']) {
    for (const names of ['', ' stub/probe']) {
      for (const enabled of [false, true]) {
        const source =
          `// ${prefix}-disable${names}\n` +
          (enabled ? `// ${prefix}-enable${names}\n` : '') +
          'x();';
        const loaded = makeReportingPlugin((ctx) => {
          ctx.report({
            loc: { line: enabled ? 3 : 2, column: 0 },
            message: 'probe',
          });
        });
        for (const collectDirectiveUsage of [false, true]) {
          const result = lintFile(
            { ...defaultRequest(source), collectDirectiveUsage },
            loaded,
          );
          expect(result.parseError).toBeUndefined();
          expect(result.diagnostics).toHaveLength(enabled ? 1 : 0);
          expect(result.usedDisableDirectives).toEqual(
            collectDirectiveUsage
              ? {
                  rules: enabled ? [] : [names ? 'stub/probe' : null],
                  entries: enabled ? [] : [[0, 0, 0]],
                }
              : undefined,
          );
        }
      }
    }
  }
});

test('line suppression expires before the next target line starts', () => {
  for (const prefix of ['eslint', 'rslint']) {
    for (const reversed of [false, true]) {
      const next = `/* ${prefix}-disable-next-line no-debugger, stub/probe */`;
      const same = `/* ${prefix}-disable-line no-debugger, stub/probe */`;
      const source =
        (reversed ? `${same} ${next}` : `${next} ${same}`) +
        '\ndebugger;\ndebugger;';
      const loaded = makeReportingPlugin((ctx) => {
        for (const line of [2, 3])
          ctx.report({ loc: { line, column: 0 }, message: 'probe' });
      });
      for (const collectDirectiveUsage of [false, true]) {
        const result = lintFile(
          { ...defaultRequest(source), collectDirectiveUsage },
          loaded,
        );
        expect(result.parseError).toBeUndefined();
        expect(result.diagnostics).toHaveLength(1);
        expect(result.diagnostics[0].startPos).toBe(
          source.lastIndexOf('debugger;'),
        );
        expect(result.usedDisableDirectives).toEqual(
          collectDirectiveUsage
            ? {
                rules: ['stub/probe'],
                entries: [[0, reversed ? same.length + 1 : 0, 0]],
              }
            : undefined,
        );
      }
    }
  }
});

test('usage summary distinguishes partial, wildcard and repeated directives', () => {
  for (const prefix of ['eslint', 'rslint']) {
    for (const source of [
      `// 😀\r\n/* ${prefix}-disable stub/probe, other/check */\r\nx();`,
      `// 😀\r\n/* ${prefix}-disable */\r\nx();`,
      `/* ${prefix}-disable stub/probe */\n/* ${prefix}-disable stub/probe */\nx();`,
      `// ${prefix}-disable-next-line stub/probe\n`,
      `// 😀\r\n/* ${prefix}-disable stub/probe -- */\r\nx();`,
    ]) {
      const loaded = makeReportingPlugin((ctx) => {
        const lines = buildLineStartOffsets(source);
        const line = lines.length;
        ctx.report({ loc: { line, column: 0 }, message: 'first' });
        ctx.report({ loc: { line, column: 0 }, message: 'second' });
      });
      const result = lintFile(
        { ...defaultRequest(source), collectDirectiveUsage: true },
        loaded,
      );
      expect(result.parseError).toBeUndefined();
      expect(result.diagnostics).toEqual([]);
      expect(result.usedDisableDirectives).toEqual({
        rules: [
          source.includes(`/* ${prefix}-disable */`) ? null : 'stub/probe',
        ],
        entries: [
          [source.startsWith(`// ${prefix}-disable-next-line`) ? 0 : 1, 0, 0],
        ],
      });
    }
  }
});

test('usage locations use UTF-16 columns against BOM-stripped source', () => {
  for (const bom of ['', '\ufeff']) {
    const loaded = makeReportingPlugin((ctx) => {
      ctx.report({ loc: { line: 2, column: 0 }, message: 'hidden' });
    });
    const result = lintFile(
      {
        ...defaultRequest(
          `${bom}"😀"; /* eslint-disable stub/probe */\r\nx();`,
        ),
        collectDirectiveUsage: true,
      },
      loaded,
    );
    expect(result.parseError).toBeUndefined();
    expect(result.diagnostics).toEqual([]);
    expect(result.usedDisableDirectives).toEqual({
      rules: ['stub/probe'],
      entries: [[0, 6, 0]],
    });
  }
});
