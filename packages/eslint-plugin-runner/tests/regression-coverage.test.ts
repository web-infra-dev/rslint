/**
 * Behavior-pinning regression tests for ESLint v10 compatibility
 * surfaces across the runner (WorkerPool / SourceCode / tokenizer /
 * scope factory / diagnostic builder / listener merge / options
 * defaults). Each `describe` block fences one specific contract that
 * has broken at least once and is now load-bearing.
 *
 * Layout follows the surface the test targets, not chronology — add
 * new entries grouped with related surface tests rather than at the
 * end. Each test states what behavior is pinned and includes a
 * 1-2 line "before this contract held" note so a future maintainer
 * reading a failure understands the regression risk being guarded.
 */
import { describe, test, expect } from '@rstest/core';
import path from 'node:path';
import * as espree from 'espree';
import { parseSync } from 'oxc-parser';

import { WorkerPool, terminateWorker } from '../src/worker-pool.js';
import { createSourceCode } from '../src/source-code/source-code.js';
import { lintFile } from '../src/linter/ecma-language-plugin.js';
import type { LoadedPlugins } from '../src/plugin/plugin-loader.js';
import type { RuleContext } from '../src/linter/context.js';
import { tokenize } from '../src/lexer/tokenizer.js';
import {
  normalizeAst,
  buildLineStartOffsets,
} from '../src/ast/normalize-ast.js';

// ────────────────────────────────────────────────────────────────────
// WorkerPool: respawn rejection must drain pendingQueue
// ────────────────────────────────────────────────────────────────────

const HANG_CONFIG_PATH = path.resolve(__dirname, 'fixtures', 'hang.config.mjs');
const HANG_CONFIG_DIR = path.dirname(HANG_CONFIG_PATH);

describe('WorkerPool respawn rejection drains pendingQueue', () => {
  test('queued task after respawn-reject does not hang lintBatch', async () => {
    const pool = new WorkerPool({
      configs: [
        { configPath: HANG_CONFIG_PATH, configDirectory: HANG_CONFIG_DIR },
      ],
      workerCount: 1,
      retryCap: 3,
      // Long enough that the in-flight hang task doesn't auto-time-out
      // before the manual terminate below — we want the failure path
      // to be "explicit terminate + respawn reject", not "task_timeout".
      taskTimeoutMs: 30_000,
    });
    await pool.init();

    // 1. In-flight hang task occupies the only worker.
    const hangPromise = pool.lintBatch([
      {
        filePath: 'hang.ts',
        text: 'const x = 1;\n',
        rules: { 'hang/hang': { options: [] } },
        collectFixes: false,
        suggestionsMode: 'off',
        configKey: HANG_CONFIG_DIR,
      },
    ]);

    // Give the worker time to receive + start the hang task.
    await new Promise((r) => setTimeout(r, 200));

    // 2. Sibling task — separate lintBatch call. Worker is busy, so
    //    kickQueue can't dispatch it; it sits in pendingQueue.
    const siblingPromise = pool.lintBatch([
      {
        filePath: 'sib.ts',
        text: 'const TRIGGER = 1;\n',
        rules: { 'hang/noop': { options: [] } },
        collectFixes: false,
        suggestionsMode: 'off',
        configKey: HANG_CONFIG_DIR,
      },
    ]);

    // Give the sibling enqueue + kickQueue a tick to land.
    await new Promise((r) => setTimeout(r, 50));

    // Sanity: sibling actually queued (worker really was busy).
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const internals = pool as any;
    expect(internals.pendingQueue.length).toBe(1);

    // 3. Monkey-patch spawnWorker so the upcoming respawn rejects
    //    rather than producing a fresh worker.
    internals.spawnWorker = () =>
      Promise.reject(new Error('simulated respawn failure'));

    // 4. Crash the worker. Exit handler:
    //      - rejects hangTask via inflight cleanup (worker_crashed)
    //      - calls spawnWorker → patched reject → ONLY LOGS pre-fix.
    //    Pre-fix: pendingQueue stays full forever.
    //    Post-fix: drainQueueIfAllSlotsDegraded() in the reject
    //              callback resolves siblings with pool_degraded.
    await terminateWorker(internals.workers[0].worker);

    // 5. Race sibling against a generous-but-bounded timeout. Pre-fix
    //    the race resolves with 'timeout'; post-fix sibling settles
    //    in well under the cap.
    const TIMEOUT_MS = 3000;
    type Outcome =
      | { tag: 'sibling'; parseError: string | undefined }
      | { tag: 'timeout' };

    const outcome: Outcome = await Promise.race([
      siblingPromise.then(
        (r): Outcome => ({ tag: 'sibling', parseError: r[0].parseError }),
      ),
      new Promise<Outcome>((r) =>
        setTimeout(() => r({ tag: 'timeout' }), TIMEOUT_MS),
      ),
    ]);

    expect(outcome.tag).toBe('sibling');
    if (outcome.tag === 'sibling') {
      // Post-fix contract: drain marks every stranded task with the
      // pool_degraded sentinel so callers can distinguish from a
      // regular per-task crash.
      expect(outcome.parseError).toBe('pool_degraded');
    }

    // Hang batch should also have settled by now — the inflight task
    // gets a worker_crashed result via the exit handler.
    const hangOutcome = await Promise.race([
      hangPromise.then((r) => r[0].parseError),
      new Promise<undefined>((r) => setTimeout(() => r(undefined), 1000)),
    ]);
    expect(hangOutcome).toMatch(/worker_crashed/);
  }, 30_000);
});

// ────────────────────────────────────────────────────────────────────
// SourceCode scope-factory throw must propagate to every caller
// ────────────────────────────────────────────────────────────────────

describe('SourceCode scope-factory throw propagates on every access', () => {
  test('second getScopeManager call re-throws (not silent null)', () => {
    // Minimal Program AST — enough to satisfy SourceCode's shape; no
    // listener walk runs in this direct test.
    const ast = {
      type: 'Program',
      body: [],
      sourceType: 'module',
      range: [0, 0] as [number, number],
      loc: {
        start: { line: 1, column: 0 },
        end: { line: 1, column: 0 },
      },
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
    } as any;

    let calls = 0;
    const sc = createSourceCode({
      text: '',
      ast,
      scopeManagerFactory: () => {
        calls++;
        throw new Error(`scope-factory failure (call ${calls})`);
      },
    });

    // First access: factory throws — this part is the same pre- and
    // post-fix. The lazy access correctly surfaces the throw.
    expect(() => sc.scopeManager).toThrow(/scope-factory failure/);

    // Second access — the regression surface.
    //
    // Pre-fix: `_scopeInit = true` was set BEFORE the factory call, so
    //   after the throw `_scope` stays `null` and subsequent accesses
    //   short-circuit on `_scopeInit && true`, silently returning the
    //   null `_scope`. Downstream `getScopeForNode(null, ...)` returns
    //   null (source-code-helpers.ts:198), so scope-dependent rules
    //   silently degrade with no error recorded.
    //
    // Post-fix: the throw must propagate again so every consumer
    //   (one ruleError per rule, not one for the unlucky first one).
    expect(() => sc.scopeManager).toThrow(/scope-factory failure/);
  });
});

// ────────────────────────────────────────────────────────────────────
// diagnostic-builder: array-fix path validates per element
// ────────────────────────────────────────────────────────────────────

function makeBadArrayFixPlugin(): LoadedPlugins {
  return {
    plugins: [],
    rules: new Map<string, unknown>([
      [
        'stub/bad-array-fix',
        {
          meta: { name: 'bad-array-fix', fixable: 'code' },
          create(ctx: RuleContext) {
            return {
              VariableDeclaration(node: unknown) {
                ctx.report({
                  node: node as never,
                  message: 'bad fix array',
                  // Plugin returns an array whose first entry is NOT
                  // a Fix (missing `range`). Pre-fix the array branch
                  // returns it verbatim and `f.range[0]` throws when
                  // the diagnostic post-processor walks it.
                  fix() {
                    return [
                      // Bogus element — caller forgot `range`.
                      // eslint-disable-next-line @typescript-eslint/no-explicit-any
                      { text: 'broken' } as any,
                      // Valid element — should survive the per-element
                      // filter and reach the wire.
                      { range: [0, 1] as [number, number], text: 'x' },
                    ];
                  },
                });
              },
            };
          },
        },
      ],
    ]),
  };
}

describe('diagnostic-builder array-fix validates per element', () => {
  test('plugin returning [malformed, valid] does not crash lintFile', () => {
    const loaded = makeBadArrayFixPlugin();
    const req = {
      filePath: 'arr.ts',
      text: 'const x = 1;\n',
      rules: { 'stub/bad-array-fix': { options: [] } },
      // Must be true — otherwise descriptor.fix(fixer) is never invoked.
      collectFixes: true,
      suggestionsMode: 'off' as const,
    };

    // Pre-fix: lintFile throws `Cannot read properties of undefined
    //   (reading '0')` from ecma-language-plugin.ts:598 (`f.range[0]`).
    //   The throw escapes lintFile because the diagnostic post-
    //   processing loop has no try/catch.
    //
    // Post-fix: per-element isFix() filter drops the bogus entry,
    //   the valid entry survives, the diagnostic and its fix make
    //   it back cleanly.
    let threw = false;
    let result: ReturnType<typeof lintFile> | undefined;
    try {
      result = lintFile(req, loaded);
    } catch {
      threw = true;
    }

    expect(threw).toBe(false);
    expect(result).toBeDefined();
    expect(result!.parseError).toBeUndefined();
    expect(result!.diagnostics).toHaveLength(1);
    const fixes = result!.diagnostics[0].fixes ?? [];
    // Only the valid fix should survive.
    expect(fixes).toHaveLength(1);
    expect(fixes[0].text).toBe('x');
  });
});

// ────────────────────────────────────────────────────────────────────
// tokenizer: regex inside template ${...} expression
// ────────────────────────────────────────────────────────────────────

describe('tokenizer recognizes regex inside template ${...}', () => {
  test('`${/re/g}` → RegularExpression token, not `/` + ident', () => {
    const { tokens } = tokenize(
      '`${/re/g}`',
      buildLineStartOffsets('`${/re/g}`'),
    );
    const kinds = tokens.map((t) => `${t.type}:${t.value}`);

    // Pre-fix: couldStartRegex returns false for every Template-prev,
    //   so `/re/g` decomposes into `Punctuator(/)`, `Identifier(re)`,
    //   `Punctuator(/)`, `Identifier(g)`. The decomposed form would
    //   match this anti-pattern:
    //     ['Template:`${', 'Punctuator:/', 'Identifier:re',
    //      'Punctuator:/', 'Identifier:g', 'Punctuator:}', 'Template:`']
    //
    // Post-fix: Template heads ending with `${` are expression-prefix,
    //   so `/` starts a regex literal. Expected tokens:
    //     Template(`${`), RegularExpression(/re/g), Punctuator(}),
    //     Template(`)
    const regex = tokens.find((t) => t.type === 'RegularExpression');
    expect(regex).toBeDefined();
    expect(regex?.value).toBe('/re/g');
    expect(regex?.regex).toEqual({ pattern: 're', flags: 'g' });

    // Pin the full sequence so a regression that splits the regex
    // again is loud rather than quiet.
    expect(kinds).toContain('RegularExpression:/re/g');
    expect(kinds).not.toContain('Identifier:re');
  });

  test('control: regex in normal expression context still works', () => {
    // Regression guard: the fix only adds the `${` case to the
    // Template branch; division on a full template literal (no `${`)
    // must STILL be classified as division, not regex.
    const { tokens } = tokenize('`hi` / 2', buildLineStartOffsets('`hi` / 2'));
    // After a full template (no opening `${`), `/` is division.
    expect(tokens.find((t) => t.type === 'RegularExpression')).toBeUndefined();
    expect(tokens.some((t) => t.type === 'Punctuator' && t.value === '/')).toBe(
      true,
    );
  });
});

// ────────────────────────────────────────────────────────────────────
// SourceCode getIndexFromLoc bounds
// ────────────────────────────────────────────────────────────────────

describe('SourceCode getIndexFromLoc throws on invalid / out-of-range input', () => {
  // Shared minimal Program AST + SourceCode factory for these tests.
  const mkSC = (text: string) => {
    const ast = {
      type: 'Program',
      body: [],
      sourceType: 'module',
      range: [0, text.length] as [number, number],
      loc: {
        start: { line: 1, column: 0 },
        end: { line: 1, column: 0 },
      },
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
    } as any;
    return createSourceCode({
      text,
      ast,
      scopeManagerFactory: () => ({}),
    });
  };

  test('line past EOF → RangeError', () => {
    const sc = mkSC('ab\ncd\n');
    // ESLint v10 throws on out-of-range line; pre-fix the runner
    // clamped silently to text.length, hiding the plugin bug.
    expect(() => sc.getIndexFromLoc({ line: 999, column: 0 })).toThrow(
      RangeError,
    );
  });

  test('line < 1 → RangeError', () => {
    const sc = mkSC('ab\ncd\n');
    expect(() => sc.getIndexFromLoc({ line: 0, column: 0 })).toThrow(
      RangeError,
    );
  });

  test('negative column → RangeError', () => {
    const sc = mkSC('ab\ncd\n');
    expect(() => sc.getIndexFromLoc({ line: 2, column: -5 })).toThrow(
      RangeError,
    );
  });

  test('column past line end → RangeError', () => {
    const sc = mkSC('ab\nXY\n');
    // Line 2 has 2 chars; column 9 is way past end.
    expect(() => sc.getIndexFromLoc({ line: 2, column: 9 })).toThrow(
      RangeError,
    );
  });

  test('valid in-range input returns correct offset (regression guard)', () => {
    const sc = mkSC('ab\nfoobar\n');
    // Line 2 starts at offset 3 (`foobar`). Column 2 = offset 5.
    expect(sc.getIndexFromLoc({ line: 2, column: 2 })).toBe(5);
  });

  test('column == lineLen is valid (one-past-last-char insertion point)', () => {
    const sc = mkSC('ab\nfoo\n');
    // Line 2 = `foo`, 3 chars; column 3 points to the newline / start
    // of next line. ESLint accepts this as a valid insertion point.
    expect(sc.getIndexFromLoc({ line: 2, column: 3 })).toBe(6);
  });
});

// ────────────────────────────────────────────────────────────────────
// program.tokens / program.comments identity
// ────────────────────────────────────────────────────────────────────

function makeIdentityProbePlugin(captured: {
  t1?: unknown;
  t2?: unknown;
  c1?: unknown;
  c2?: unknown;
}): LoadedPlugins {
  return {
    plugins: [],
    rules: new Map<string, unknown>([
      [
        'stub/identity-probe',
        {
          meta: { name: 'identity-probe' },
          create(_ctx: RuleContext) {
            return {
              Program(node: unknown) {
                const ast = node as {
                  tokens: unknown;
                  comments: unknown;
                };
                if (captured.t1 === undefined) {
                  captured.t1 = ast.tokens;
                  captured.c1 = ast.comments;
                  captured.t2 = ast.tokens;
                  captured.c2 = ast.comments;
                }
              },
            };
          },
        },
      ],
    ]),
  };
}

describe('program.tokens / program.comments are stable references', () => {
  test('two reads of program.tokens return the same array', () => {
    const captured: {
      t1?: unknown;
      t2?: unknown;
      c1?: unknown;
      c2?: unknown;
    } = {};
    const loaded = makeIdentityProbePlugin(captured);

    const result = lintFile(
      {
        filePath: 'id.ts',
        text: '// hi\nconst x = 1;\n',
        rules: { 'stub/identity-probe': { options: [] } },
        collectFixes: false,
        suggestionsMode: 'off',
      },
      loaded,
    );

    expect(result.parseError).toBeUndefined();
    // Pre-fix: getter rebuilt the array each call → t1 !== t2. Same
    // for comments. Post-fix: cached after first access.
    expect(captured.t1).toBe(captured.t2);
    expect(captured.c1).toBe(captured.c2);
  });
});

// ────────────────────────────────────────────────────────────────────
// diagnostic offset clamp truncates non-integer offsets
// ────────────────────────────────────────────────────────────────────

function makeFractionalLocPlugin(): LoadedPlugins {
  return {
    plugins: [],
    rules: new Map<string, unknown>([
      [
        'stub/fractional-loc',
        {
          meta: { name: 'fractional-loc' },
          create(ctx: RuleContext) {
            return {
              Program(node: unknown) {
                // Fractional loc.column — a buggy plugin or external
                // tooling could produce one. Pre-fix clamp accepted
                // finite decimals, `u16ToByte[2.5] === undefined`
                // leaked onto the wire. Post-fix: Math.trunc keeps
                // offsets integer.
                ctx.report({
                  node: node as never,
                  loc: { line: 1, column: 2.5 } as never,
                  message: 'fractional',
                });
              },
            };
          },
        },
      ],
    ]),
  };
}

describe('diagnostic wire-conversion clamp produces integer byte offsets', () => {
  test('fractional column → integer startPos (not undefined leak)', () => {
    const loaded = makeFractionalLocPlugin();
    const result = lintFile(
      {
        filePath: 'frac.ts',
        text: 'const xyz = 1;\n',
        rules: { 'stub/fractional-loc': { options: [] } },
        collectFixes: false,
        suggestionsMode: 'off',
      },
      loaded,
    );

    expect(result.parseError).toBeUndefined();
    expect(result.diagnostics).toHaveLength(1);
    const d = result.diagnostics[0];
    expect(typeof d.startPos).toBe('number');
    expect(Number.isInteger(d.startPos)).toBe(true);
    expect(d.startPos).toBe(2);
  });
});

// ────────────────────────────────────────────────────────────────────
// getNodeByRangeIndex falls back to getVisitorKeys
// ────────────────────────────────────────────────────────────────────

describe('getNodeByRangeIndex uses getVisitorKeys fallback', () => {
  test('descends into unknown node types via the visitor-keys helper', () => {
    // Construct an AST with a synthetic node type (not in the static
    // RUNNER_VISITOR_KEYS table). With the fallback, getVisitorKeys
    // discovers children via the heuristic helper, and the walk
    // descends into the inner node. Without it, the walk stops at the
    // wrapper and returns the wrapper instead of the deeper node —
    // mismatching listener-merge.ts / normalize-ast.ts behavior.
    const inner = {
      type: 'InnerSynthetic',
      range: [5, 10] as [number, number],
      loc: {
        start: { line: 1, column: 5 },
        end: { line: 1, column: 10 },
      },
    };
    const wrapper = {
      type: 'WrapperSynthetic',
      body: inner,
      range: [0, 15] as [number, number],
      loc: {
        start: { line: 1, column: 0 },
        end: { line: 1, column: 15 },
      },
    };
    const ast = {
      type: 'Program',
      body: [wrapper],
      sourceType: 'module',
      range: [0, 15] as [number, number],
      loc: {
        start: { line: 1, column: 0 },
        end: { line: 1, column: 15 },
      },
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
    } as any;

    const sc = createSourceCode({
      text: 'aaaaa12345xxxxx',
      ast,
      scopeManagerFactory: () => ({}),
    });

    // Index 7 is inside `inner` (5..10). The walk must descend through
    // wrapper → inner via getVisitorKeys (which inspects the wrapper's
    // fields and picks up `.body` as a child). Pre-fix the bare
    // `?? []` skipped the wrapper's children entirely, so the deepest
    // match was the wrapper itself.
    const node = sc.getNodeByRangeIndex(7);
    expect(node).not.toBeNull();
    expect(node?.type).toBe('InnerSynthetic');
  });
});

// ────────────────────────────────────────────────────────────────────
// tokenizer: `/` after identifier-shaped contextual keyword
// ────────────────────────────────────────────────────────────────────

describe('tokenizer treats `/` after identifier `from`/`of`/`as`/`async` as division', () => {
  const cases: Array<[name: string, src: string]> = [
    ['from', 'const k = from / total;'],
    ['of', 'const k = of / 2;'],
    ['as', 'const k = as / 2;'],
    ['async', 'const k = async / 2;'],
  ];
  for (const [name, src] of cases) {
    test(`\`${src}\` — \`/\` after \`${name}\` is Punctuator, not regex start`, () => {
      const { tokens } = tokenize(src, buildLineStartOffsets(src));
      // After ANY identifier — including the contextual keywords
      // `from`/`of`/`as`/`async`/`await` — espree's token layer treats
      // `/` as division (the identifier produces a value). An earlier
      // version put some of these in a "regex-prefix" set, which
      // mis-scanned the rest of the line into one phantom
      // RegularExpression token, shifting all downstream tokens.
      expect(
        tokens.find((t) => t.type === 'RegularExpression'),
      ).toBeUndefined();
      const slashTokens = tokens.filter(
        (t) => t.type === 'Punctuator' && t.value === '/',
      );
      expect(slashTokens).toHaveLength(1);
    });
  }

  test('`await /re/g.test(x)` — `/` after `await` is division (matches espree@11)', () => {
    // espree@11 tokenizes `await /re/g.test(x)` as `await` `/` `re` `/`
    // `g` … in BOTH sloppy and async-module mode: `await` is an
    // Identifier-shaped value and the `/` after it is division, so
    // `/re/g` decomposes into two `/` Punctuators around `re`/`g`. An
    // earlier runner wrongly emitted a single RegularExpression here.
    const src = 'async function f() { return await /re/g.test(x); }';
    const { tokens } = tokenize(src, buildLineStartOffsets(src));
    expect(tokens.find((t) => t.type === 'RegularExpression')).toBeUndefined();
    expect(tokens.find((t) => t.value === 'await')?.type).toBe('Identifier');
    expect(
      tokens.filter((t) => t.type === 'Punctuator' && t.value === '/'),
    ).toHaveLength(2);
  });
});

// ────────────────────────────────────────────────────────────────────
// source-code-helpers: `count: 0` returns empty array
// ────────────────────────────────────────────────────────────────────

function makeTokenProbePlugin(
  captured: Record<string, unknown>,
  probe: (ctx: RuleContext, node: unknown) => void,
): LoadedPlugins {
  // captured is intentionally typed as Record so each test can store
  // arbitrary keys without needing a dedicated interface per probe.
  void captured;
  return {
    plugins: [],
    rules: new Map<string, unknown>([
      [
        'stub/token-probe',
        {
          meta: { name: 'token-probe' },
          create(ctx: RuleContext) {
            return {
              VariableDeclaration(node: unknown) {
                probe(ctx, node);
              },
            };
          },
        },
      ],
    ]),
  };
}

describe('count:0 returns empty token array (ESLint v10 parity)', () => {
  test('getFirstTokens(node, 0) returns [] (not all tokens)', () => {
    const captured: { tokens?: unknown } = {};
    const loaded = makeTokenProbePlugin(captured, (ctx, node) => {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      captured.tokens = (ctx.sourceCode as any).getFirstTokens(node, 0);
    });
    const result = lintFile(
      {
        filePath: 'h3.ts',
        text: 'const x = 1;',
        rules: { 'stub/token-probe': { options: [] } },
        collectFixes: false,
        suggestionsMode: 'off',
      },
      loaded,
    );
    expect(result.parseError).toBeUndefined();
    // Pre-fix: count:0 → "no upper bound" → returns all 5 tokens of
    // `const x = 1;` (const, x, =, 1, ;)
    // Post-fix: count:0 → empty array, matching ESLint v10.
    expect(captured.tokens).toEqual([]);
  });

  test('getFirstTokens(node, {count:0}) returns [] (not all tokens)', () => {
    const captured: { tokens?: unknown } = {};
    const loaded = makeTokenProbePlugin(captured, (ctx, node) => {
      captured.tokens =
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        (ctx.sourceCode as any).getFirstTokens(node, { count: 0 });
    });
    lintFile(
      {
        filePath: 'h3.ts',
        text: 'const x = 1;',
        rules: { 'stub/token-probe': { options: [] } },
        collectFixes: false,
        suggestionsMode: 'off',
      },
      loaded,
    );
    expect(captured.tokens).toEqual([]);
  });

  test('control: omitted opts still returns all tokens', () => {
    const captured: { tokens?: unknown[] } = {};
    const loaded = makeTokenProbePlugin(captured, (ctx, node) => {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      captured.tokens = (ctx.sourceCode as any).getFirstTokens(node);
    });
    lintFile(
      {
        filePath: 'h3.ts',
        text: 'const x = 1;',
        rules: { 'stub/token-probe': { options: [] } },
        collectFixes: false,
        suggestionsMode: 'off',
      },
      loaded,
    );
    // Omitted opts → no cap → all tokens.
    expect((captured.tokens ?? []).length).toBeGreaterThan(0);
  });

  test('getLastTokens(node, 0) returns [] (count:0 parity)', () => {
    const captured: { tokens?: unknown } = {};
    const loaded = makeTokenProbePlugin(captured, (ctx, node) => {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      captured.tokens = (ctx.sourceCode as any).getLastTokens(node, 0);
    });
    lintFile(
      {
        filePath: 'h3.ts',
        text: 'const x = 1;',
        rules: { 'stub/token-probe': { options: [] } },
        collectFixes: false,
        suggestionsMode: 'off',
      },
      loaded,
    );
    expect(captured.tokens).toEqual([]);
  });
});

// ────────────────────────────────────────────────────────────────────
// diagnostic-builder: message interpolation matches ESLint v10
// ────────────────────────────────────────────────────────────────────

function makeInterpolationPlugin(
  message: string,
  data: Record<string, string>,
): LoadedPlugins {
  return {
    plugins: [],
    rules: new Map<string, unknown>([
      [
        'stub/interpolate',
        {
          meta: { name: 'interpolate' },
          create(ctx: RuleContext) {
            return {
              Program(node: unknown) {
                ctx.report({
                  node: node as never,
                  message,
                  data,
                });
              },
            };
          },
        },
      ],
    ]),
  };
}

describe('message interpolation matches ESLint v10 interpolate.js', () => {
  test('`{{ name }}` with spaces is replaced (trim placeholder term)', () => {
    const loaded = makeInterpolationPlugin('hello {{ who }}', { who: 'world' });
    const result = lintFile(
      {
        filePath: 'i.ts',
        text: 'const x = 1;',
        rules: { 'stub/interpolate': { options: [] } },
        collectFixes: false,
        suggestionsMode: 'off',
      },
      loaded,
    );
    expect(result.diagnostics).toHaveLength(1);
    // Pre-fix: literal replaceAll('{{who}}') misses '{{ who }}'.
    // Post-fix: regex /\{\{([^{}]+?)\}\}/gu + trim() matches.
    expect(result.diagnostics[0].message).toBe('hello world');
  });

  test('data value containing `$&` is inserted literally', () => {
    const loaded = makeInterpolationPlugin('got {{val}}', { val: '$&' });
    const result = lintFile(
      {
        filePath: 'i.ts',
        text: 'const x = 1;',
        rules: { 'stub/interpolate': { options: [] } },
        collectFixes: false,
        suggestionsMode: 'off',
      },
      loaded,
    );
    expect(result.diagnostics).toHaveLength(1);
    // Pre-fix: String.prototype.replaceAll interprets `$&` in
    // replacement value as the matched substring → message becomes
    // 'got {{val}}'.
    // Post-fix: function-form replacer inserts the raw value.
    expect(result.diagnostics[0].message).toBe('got $&');
  });

  test('data value containing `$$` is inserted literally', () => {
    const loaded = makeInterpolationPlugin('got {{val}}', { val: '$$' });
    const result = lintFile(
      {
        filePath: 'i.ts',
        text: 'const x = 1;',
        rules: { 'stub/interpolate': { options: [] } },
        collectFixes: false,
        suggestionsMode: 'off',
      },
      loaded,
    );
    // Pre-fix: `$$` becomes `$`. Post-fix: literal `$$`.
    expect(result.diagnostics[0].message).toBe('got $$');
  });

  test('missing data key leaves placeholder verbatim (ESLint behavior)', () => {
    const loaded = makeInterpolationPlugin('hello {{missing}}', {
      other: 'x',
    });
    const result = lintFile(
      {
        filePath: 'i.ts',
        text: 'const x = 1;',
        rules: { 'stub/interpolate': { options: [] } },
        collectFixes: false,
        suggestionsMode: 'off',
      },
      loaded,
    );
    // ESLint v10 interpolate.js leaves `{{missing}}` in place when the
    // data object doesn't have the key.
    expect(result.diagnostics[0].message).toBe('hello {{missing}}');
  });
});

// ────────────────────────────────────────────────────────────────────
// diagnostic-builder: report({ node: syntheticNode }) with no range
// ────────────────────────────────────────────────────────────────────

describe('report({node}) on loc-only synthetic node uses node.loc', () => {
  test('synthetic node without range still produces a diagnostic', () => {
    const loaded: LoadedPlugins = {
      plugins: [],
      rules: new Map<string, unknown>([
        [
          'stub/loc-only',
          {
            meta: { name: 'loc-only' },
            create(ctx: RuleContext) {
              return {
                Program() {
                  // Synthetic node — only `loc`, no `range`. ESLint
                  // accepts this (its diagnostic builder reads
                  // `node.loc` directly). Pre-fix runner destructures
                  // `descriptor.node.range` unconditionally → TypeError
                  // → swallowed by the listener try/catch → diagnostic
                  // silently dropped.
                  ctx.report({
                    node: {
                      type: 'SyntheticIdentifier',
                      loc: {
                        start: { line: 1, column: 0 },
                        end: { line: 1, column: 5 },
                      },
                    } as never,
                    message: 'loc only',
                  });
                },
              };
            },
          },
        ],
      ]),
    };

    const result = lintFile(
      {
        filePath: 'm1.ts',
        text: 'const x = 1;\n',
        rules: { 'stub/loc-only': { options: [] } },
        collectFixes: false,
        suggestionsMode: 'off',
      },
      loaded,
    );
    expect(result.parseError).toBeUndefined();
    // Pre-fix: 0 diagnostics + a ruleErrors entry. Post-fix: 1 diag.
    expect(result.diagnostics).toHaveLength(1);
    expect(result.diagnostics[0].message).toBe('loc only');
  });
});

// ────────────────────────────────────────────────────────────────────
// diagnostic-builder: fix() returning a string is rejected (not iterated)
// ────────────────────────────────────────────────────────────────────

describe('fix() returning a string is rejected, diagnostic survives', () => {
  test('string return from fix() does not crash lintFile and drops the fix', () => {
    const loaded: LoadedPlugins = {
      plugins: [],
      rules: new Map<string, unknown>([
        [
          'stub/bad-string-fix',
          {
            meta: { name: 'bad-string-fix', fixable: 'code' },
            create(ctx: RuleContext) {
              return {
                Program(node: unknown) {
                  ctx.report({
                    node: node as never,
                    message: 'fix returning string',
                    // Plugin contract violation: returns a string. Pre-
                    // fix the iterable branch caught it (strings are
                    // iterable) and `Array.from('foo') = ['f','o','o']`
                    // produced 3 character-shaped pseudo-fixes; the
                    // diagnostic post-processor then crashed on
                    // `f.range[0]`. Post-fix: the string is rejected
                    // with a console.error and the diagnostic still
                    // surfaces with no fix attached.
                    // eslint-disable-next-line @typescript-eslint/no-explicit-any
                    fix: (() => 'foo') as any,
                  });
                },
              };
            },
          },
        ],
      ]),
    };

    let threw = false;
    let result: ReturnType<typeof lintFile> | undefined;
    try {
      result = lintFile(
        {
          filePath: 'm2.ts',
          text: 'const x = 1;\n',
          rules: { 'stub/bad-string-fix': { options: [] } },
          collectFixes: true,
          suggestionsMode: 'off',
        },
        loaded,
      );
    } catch {
      threw = true;
    }
    expect(threw).toBe(false);
    expect(result).toBeDefined();
    expect(result!.parseError).toBeUndefined();
    expect(result!.diagnostics).toHaveLength(1);
    // No fixes (the string was rejected as malformed).
    expect(result!.diagnostics[0].fixes ?? []).toHaveLength(0);
  });
});

// ────────────────────────────────────────────────────────────────────
// tokenizer: `?.` followed by a digit is NOT optional chain
// ────────────────────────────────────────────────────────────────────

describe('`?.` followed by a decimal digit tokenizes as `?` + `.NUM`', () => {
  test('`cond?.4:.2;` — `?.4` is `?` + `.4`, not `?.` + `4`', () => {
    const src = 'cond?.4:.2;';
    const { tokens } = tokenize(src, buildLineStartOffsets(src));
    const kinds = tokens.map((t) => `${t.type}:${t.value}`);
    // Pre-fix: ['Identifier:cond', 'Punctuator:?.', 'Numeric:4',
    //           'Punctuator:;'].split mid-stream  — gets `?.` + `4`.
    // Post-fix per spec: `?.` cannot be followed by DecimalDigit, so
    //   the lexer falls back to `?` + `.4`.
    expect(kinds).toEqual([
      'Identifier:cond',
      'Punctuator:?',
      'Numeric:.4',
      'Punctuator::',
      'Numeric:.2',
      'Punctuator:;',
    ]);
  });

  test('control: `obj?.prop` still optional chain (no digit after)', () => {
    const src = 'obj?.prop';
    const { tokens } = tokenize(src, buildLineStartOffsets(src));
    const kinds = tokens.map((t) => `${t.type}:${t.value}`);
    expect(kinds).toEqual([
      'Identifier:obj',
      'Punctuator:?.',
      'Identifier:prop',
    ]);
  });
});

// ────────────────────────────────────────────────────────────────────
// scope-factory: TS branch honors globalReturn + commonjs
// ────────────────────────────────────────────────────────────────────

describe('scope-factory TS path honors globalReturn / commonjs', () => {
  test('M5: .ts + ecmaFeatures.globalReturn:true wraps Program in function scope', () => {
    interface Observed {
      childScopeTypes?: string[];
    }
    const observed: Observed = {};
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
                    globalScope: { childScopes: Array<{ type: string }> };
                  };
                  observed.childScopeTypes = sm.globalScope.childScopes.map(
                    (c) => c.type,
                  );
                },
              };
            },
          },
        ],
      ]),
    };

    lintFile(
      {
        filePath: 'global-return.ts',
        text: 'var topVar = 1;\nreturn topVar;\n',
        rules: { 'stub/probe': { options: [] } },
        collectFixes: false,
        suggestionsMode: 'off',
        languageOptions: {
          sourceType: 'script',
          parserOptions: {
            ecmaFeatures: { globalReturn: true },
          },
        },
      },
      loaded,
    );

    // Pre-fix: TS branch dropped globalReturn on the floor → no
    // function wrapper. Post-fix: ts-scope-manager analyses with
    // globalReturn:true → function child scope.
    expect(observed.childScopeTypes).toContain('function');
  });

  test('L2: .cts + sourceType:commonjs matches .cjs (function wrapper)', () => {
    interface Observed {
      childScopeTypes?: string[];
    }
    const observed: Observed = {};
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
                    globalScope: { childScopes: Array<{ type: string }> };
                  };
                  observed.childScopeTypes = sm.globalScope.childScopes.map(
                    (c) => c.type,
                  );
                },
              };
            },
          },
        ],
      ]),
    };

    lintFile(
      {
        filePath: 'cjs-style.cts',
        text: 'var topVar = 1;\n',
        rules: { 'stub/probe': { options: [] } },
        collectFixes: false,
        suggestionsMode: 'off',
        languageOptions: {
          sourceType: 'commonjs',
        },
      },
      loaded,
    );

    // Pre-fix: ts-scope-manager treated 'commonjs' as 'script' (no
    // function wrapper); .cjs path through eslint-scope auto-wrapped.
    // Same source got DIFFERENT scope trees by extension.
    // Post-fix: TS path adds globalReturn=true when sourceType is
    // commonjs, matching eslint-scope's behavior.
    expect(observed.childScopeTypes).toContain('function');
  });
});

// ────────────────────────────────────────────────────────────────────
// ecma-language-plugin: .mts / .cts + ecmaFeatures.jsx promote to tsx
// ────────────────────────────────────────────────────────────────────

describe('.mts / .cts + ecmaFeatures.jsx parse as tsx', () => {
  for (const ext of ['mts', 'cts'] as const) {
    test(`.${ext} with JSX content yields a real JSXElement node`, () => {
      let sawJsx = false;
      const loaded: LoadedPlugins = {
        plugins: [],
        rules: new Map<string, unknown>([
          [
            'stub/jsx-probe',
            {
              meta: { name: 'jsx-probe' },
              create() {
                return {
                  JSXElement() {
                    sawJsx = true;
                  },
                };
              },
            },
          ],
        ]),
      };

      lintFile(
        {
          filePath: `Component.${ext}`,
          text: 'const a = <Foo />;\n',
          rules: { 'stub/jsx-probe': { options: [] } },
          collectFixes: false,
          suggestionsMode: 'off',
          languageOptions: {
            sourceType: 'module',
            parserOptions: { ecmaFeatures: { jsx: true } },
          },
        },
        loaded,
      );

      // Pre-fix: .mts / .cts not promoted to 'tsx' → oxc parses as ts
      // (no JSX) → JSXElement never appears in the AST → listener
      // never fires → sawJsx stays false (rule-blind on JSX).
      // Post-fix: promoted to tsx → AST has JSXElement → listener
      // fires.
      expect(sawJsx).toBe(true);
    });
  }
});

// ────────────────────────────────────────────────────────────────────
// WorkerPool.shutdown skips already-exited slots
// ────────────────────────────────────────────────────────────────────

describe('WorkerPool shutdown does not wait for already-exited workers', () => {
  test('shutdown after a worker.terminate() returns promptly', async () => {
    const HANG_PATH = path.resolve(__dirname, 'fixtures', 'hang.config.mjs');
    const HANG_DIR = path.dirname(HANG_PATH);
    const pool = new WorkerPool({
      configs: [{ configPath: HANG_PATH, configDirectory: HANG_DIR }],
      workerCount: 2,
      retryCap: 0, // prevent respawn so terminated slot stays dead
      taskTimeoutMs: 10_000,
    });
    await pool.init();

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const internals = pool as any;
    // Terminate worker[0]; its 'exit' event fires now. With retryCap=0
    // the slot is NOT respawned — it stays in this.workers as dead.
    await terminateWorker(internals.workers[0].worker);
    // Drain pendingQueue cascade.
    await new Promise((r) => setTimeout(r, 100));

    const start = Date.now();
    await pool.shutdown();
    const elapsed = Date.now() - start;

    // Pre-fix: shutdown registers `once('exit')` on the dead slot;
    // 'exit' already fired so the listener never runs → 5_000ms
    // terminate fallback wins. Post-fix: dead slot skipped, shutdown
    // returns in well under 5s.
    expect(elapsed).toBeLessThan(2_000);
  }, 15_000);
});

// ────────────────────────────────────────────────────────────────────
// listener-merge: unified specificity sort + raw dedup
// ────────────────────────────────────────────────────────────────────

describe('bare-type and esquery listeners share one specificity ordering', () => {
  test('cross-rule order on a single Identifier matches ESLint v10', () => {
    const order: string[] = [];
    const loaded: LoadedPlugins = {
      plugins: [],
      rules: new Map<string, unknown>([
        [
          'stub/bare',
          {
            meta: { name: 'bare' },
            create() {
              return {
                Identifier() {
                  order.push('bare/Identifier');
                },
              };
            },
          },
        ],
        [
          'stub/esq',
          {
            meta: { name: 'esq' },
            create() {
              return {
                ':matches(Identifier)'() {
                  order.push('esq/:matches');
                },
              };
            },
          },
        ],
      ]),
    };

    lintFile(
      {
        filePath: 'm4.js',
        text: 'a;\n',
        rules: {
          'stub/bare': { options: [] },
          'stub/esq': { options: [] },
        },
        collectFixes: false,
        suggestionsMode: 'off',
      },
      loaded,
    );

    // Both selectors have identifierCount=1, attributeCount=0; tie-
    // break compares raw strings. `:` (0x3A) < `I` (0x49), so
    // `:matches(Identifier)` sorts BEFORE `Identifier`.
    // Pre-fix: simple.enter fired first (bare always before esquery),
    //   producing ['bare/Identifier', 'esq/:matches'].
    // Post-fix: unified specificity sort → esquery wins the tie-break.
    expect(order).toEqual(['esq/:matches', 'bare/Identifier']);
  });
});

describe('same esquery selector registered by 2 rules deduplicates to 1 bucket entry', () => {
  test('two rules with `:matches(Identifier)` produce one entry with 2 listeners', async () => {
    const { mergeListeners } = await import('../src/linter/listener-merge.js');
    const merged = mergeListeners([
      {
        ruleName: 'A',
        listeners: { ':matches(Identifier)': () => {} },
      },
      {
        ruleName: 'B',
        listeners: { ':matches(Identifier)': () => {} },
      },
    ]);
    const bucket = merged.esqueryByType.enter.get('Identifier') ?? [];
    // Pre-fix: 2 entries (one per rule). Post-fix: 1 deduplicated entry
    // whose `listeners` array preserves registration order.
    expect(bucket).toHaveLength(1);
    expect(bucket[0].listeners.map((l) => l.ruleName)).toEqual(['A', 'B']);
  });
});

// ────────────────────────────────────────────────────────────────────
// getIndexFromLoc: malformed loc shape throws TypeError
// ────────────────────────────────────────────────────────────────────

describe('getIndexFromLoc throws TypeError on malformed loc shape', () => {
  test('missing column → TypeError', () => {
    const text = 'ab\ncd\n';
    const ast = {
      type: 'Program',
      body: [],
      sourceType: 'module',
      range: [0, text.length] as [number, number],
      loc: {
        start: { line: 1, column: 0 },
        end: { line: 3, column: 0 },
      },
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
    } as any;
    const sc = createSourceCode({
      text,
      ast,
      scopeManagerFactory: () => ({}),
    });
    // ESLint v10 throws on missing column; pre-fix the runner produced
    // `lineStart + undefined = NaN` and shipped NaN positions to the
    // wire — silently broken diagnostics from buggy plugins.
    expect(() =>
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      sc.getIndexFromLoc({ line: 2 } as any),
    ).toThrow(TypeError);
  });

  test('null loc → TypeError', () => {
    const text = 'ab\n';
    const ast = {
      type: 'Program',
      body: [],
      sourceType: 'module',
      range: [0, text.length] as [number, number],
      loc: {
        start: { line: 1, column: 0 },
        end: { line: 2, column: 0 },
      },
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
    } as any;
    const sc = createSourceCode({
      text,
      ast,
      scopeManagerFactory: () => ({}),
    });
    expect(() =>
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      sc.getIndexFromLoc(null as any),
    ).toThrow(TypeError);
  });
});

// ────────────────────────────────────────────────────────────────────
// getTokenByRangeStart honors includeComments opts
// ────────────────────────────────────────────────────────────────────

describe('getTokenByRangeStart returns comment when includeComments:true', () => {
  test('comment-starting position returns the Comment node (not null)', () => {
    let captured: unknown = 'sentinel';
    const loaded: LoadedPlugins = {
      plugins: [],
      rules: new Map<string, unknown>([
        [
          'stub/comment-probe',
          {
            meta: { name: 'comment-probe' },
            create(ctx: RuleContext) {
              return {
                Program() {
                  // The block comment starts at byte 0.
                  captured =
                    // eslint-disable-next-line @typescript-eslint/no-explicit-any
                    (ctx.sourceCode as any).getTokenByRangeStart(0, {
                      includeComments: true,
                    });
                },
              };
            },
          },
        ],
      ]),
    };

    lintFile(
      {
        filePath: 'l4.ts',
        text: '/* hi */ const x = 1;\n',
        rules: { 'stub/comment-probe': { options: [] } },
        collectFixes: false,
        suggestionsMode: 'off',
      },
      loaded,
    );

    // Pre-fix: signature was `getTokenByRangeStart(start: number)`,
    // ignored the options object → only code tokens searched → comment
    // at offset 0 not found → null.
    // Post-fix: { includeComments: true } unlocks the comment stream
    // → returns the Block comment.
    expect(captured).not.toBeNull();
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    expect((captured as any)?.type).toBe('Block');
  });
});

// ────────────────────────────────────────────────────────────────────
// tokenizer: unterminated regex at EOL preserves last char in pattern
// ────────────────────────────────────────────────────────────────────

describe('unterminated regex at EOL keeps all body characters', () => {
  test('`/abc` followed by newline preserves `abc` in pattern (not `ab`)', () => {
    // oxc would normally flag this as a parse error; the runner's
    // tokenizer runs anyway in case oxc-parser falls back. Pre-fix
    // the `slashEnd = i` claim (line at the EOL break) caused the
    // pattern slice to drop the final body character.
    const src = '/abc\n';
    const { tokens } = tokenize(src, buildLineStartOffsets(src));
    const regex = tokens.find((t) => t.type === 'RegularExpression');
    expect(regex).toBeDefined();
    // Pre-fix: pattern == 'ab' (last `c` cut off).
    // Post-fix: pattern == 'abc'.
    expect(regex?.regex?.pattern).toBe('abc');
  });
});

// ────────────────────────────────────────────────────────────────────
// getInlineConfigNodes rejects look-alike non-directive prefixes
// ────────────────────────────────────────────────────────────────────

describe('getInlineConfigNodes does NOT match eslint-disable-foo', () => {
  test('`eslint-disable-foo` is not recognised as an inline directive', () => {
    let directives: unknown[] = [];
    const loaded: LoadedPlugins = {
      plugins: [],
      rules: new Map<string, unknown>([
        [
          'stub/directive-probe',
          {
            meta: { name: 'directive-probe' },
            create(ctx: RuleContext) {
              return {
                Program() {
                  // eslint-disable-next-line @typescript-eslint/no-explicit-any
                  directives = (ctx.sourceCode as any).getInlineConfigNodes();
                },
              };
            },
          },
        ],
      ]),
    };

    lintFile(
      {
        filePath: 'l6.ts',
        text: '/* eslint-disable-foo */\nconst x = 1;\n',
        rules: { 'stub/directive-probe': { options: [] } },
        collectFixes: false,
        suggestionsMode: 'off',
      },
      loaded,
    );

    // Pre-fix: `v.startsWith('eslint-disable')` matches
    //   `eslint-disable-foo` (no word boundary). Post-fix: a non-empty
    //   suffix must start with whitespace or end the directive.
    expect(directives).toEqual([]);
  });

  test('control: real `/* eslint-disable foo */` IS recognised', () => {
    let directives: Array<{ value: string }> = [];
    const loaded: LoadedPlugins = {
      plugins: [],
      rules: new Map<string, unknown>([
        [
          'stub/directive-probe',
          {
            meta: { name: 'directive-probe' },
            create(ctx: RuleContext) {
              return {
                Program() {
                  directives =
                    // eslint-disable-next-line @typescript-eslint/no-explicit-any
                    (ctx.sourceCode as any).getInlineConfigNodes() as Array<{
                      value: string;
                    }>;
                },
              };
            },
          },
        ],
      ]),
    };

    lintFile(
      {
        filePath: 'l6c.ts',
        text: '/* eslint-disable no-undef */\nconst x = 1;\n',
        rules: { 'stub/directive-probe': { options: [] } },
        collectFixes: false,
        suggestionsMode: 'off',
      },
      loaded,
    );

    expect(directives).toHaveLength(1);
    expect(directives[0].value.trim()).toBe('eslint-disable no-undef');
  });
});

// ────────────────────────────────────────────────────────────────────
// boundary token getters return null (not undefined) to match ESLint
// ────────────────────────────────────────────────────────────────────

describe('boundary token getters return null (not undefined)', () => {
  test('getTokenBefore at start of source returns null', () => {
    let captured: unknown = 'sentinel';
    const loaded: LoadedPlugins = {
      plugins: [],
      rules: new Map<string, unknown>([
        [
          'stub/edge-probe',
          {
            meta: { name: 'edge-probe' },
            create(ctx: RuleContext) {
              return {
                Program(node: unknown) {
                  // No token before the Program.
                  captured = ctx.sourceCode.getTokenBefore(node as never);
                },
              };
            },
          },
        ],
      ]),
    };

    lintFile(
      {
        filePath: 'l7.ts',
        text: 'const x = 1;\n',
        rules: { 'stub/edge-probe': { options: [] } },
        collectFixes: false,
        suggestionsMode: 'off',
      },
      loaded,
    );

    // Pre-fix: undefined. Post-fix: null (ESLint v10 contract).
    expect(captured).toBeNull();
  });

  test('getTokenAfter past EOF returns null', () => {
    let captured: unknown = 'sentinel';
    const loaded: LoadedPlugins = {
      plugins: [],
      rules: new Map<string, unknown>([
        [
          'stub/edge-probe',
          {
            meta: { name: 'edge-probe' },
            create(ctx: RuleContext) {
              return {
                Program(node: unknown) {
                  captured = ctx.sourceCode.getTokenAfter(node as never);
                },
              };
            },
          },
        ],
      ]),
    };

    lintFile(
      {
        filePath: 'l7.ts',
        text: 'const x = 1;',
        rules: { 'stub/edge-probe': { options: [] } },
        collectFixes: false,
        suggestionsMode: 'off',
      },
      loaded,
    );
    expect(captured).toBeNull();
  });
});

// ────────────────────────────────────────────────────────────────────
// applyOptionDefaults: user option objects must not be shared
// ────────────────────────────────────────────────────────────────────

describe('applyOptionDefaults clones user-supplied option objects', () => {
  test('mutating a returned option object does not leak to the next call', async () => {
    const { applyOptionDefaults } =
      await import('../src/linter/options-defaults.js');
    // Single user-supplied object shared across multiple lint calls
    // (matches the in-process `compat-task-builder` pattern where
    // `sharedRules` lands on every file's task by reference).
    const sharedUserOpt = { allow: ['Array'] };
    const sharedUserOptions = [sharedUserOpt];

    const a = applyOptionDefaults(sharedUserOptions, undefined) as Array<{
      allow: string[];
    }>;
    // Rule mutates its options (push to allow-list).
    a[0].allow.push('Object');

    const b = applyOptionDefaults(sharedUserOptions, undefined) as Array<{
      allow: string[];
    }>;
    // Pre-fix: b[0].allow === sharedUserOpt.allow === ['Array','Object']
    //   (mutation leaked across calls).
    // Post-fix: deep clone → b[0].allow === ['Array'] (untouched).
    expect(b[0].allow).toEqual(['Array']);
    // And the original user object is preserved.
    expect(sharedUserOpt.allow).toEqual(['Array']);
  });

  // `cloneDefault` uses `structuredClone`, which throws `DataCloneError`
  // on a non-cloneable value such as a function (verified against Node
  // 22). ESLint never deep-clones options (passes them by reference), so
  // a function-valued option must NOT make option-defaulting throw — it
  // falls back to returning the value by reference.
  test('function-valued option does not throw; passed through by reference', async () => {
    const { applyOptionDefaults } =
      await import('../src/linter/options-defaults.js');
    const cb = (): number => 1;
    const userOptions = [{ cb }];
    let out: Array<{ cb: () => number }> = [];
    expect(() => {
      out = applyOptionDefaults(userOptions, undefined) as Array<{
        cb: () => number;
      }>;
    }).not.toThrow();
    // By-reference fallback: the function identity is preserved (a
    // function can't be deep-cloned, and ESLint passes options by ref).
    expect(out[0].cb).toBe(cb);
  });

  test('function inside meta.defaultOptions does not throw', async () => {
    const { applyOptionDefaults } =
      await import('../src/linter/options-defaults.js');
    const cb = (): number => 2;
    // No user option for slot 0; default carries a function.
    let out: Array<{ cb: () => number }> = [];
    expect(() => {
      out = applyOptionDefaults([], undefined, [{ cb }]) as Array<{
        cb: () => number;
      }>;
    }).not.toThrow();
    expect(out[0].cb).toBe(cb);
  });

  // End-to-end: the in-process `lintFile` path resolves options inside
  // `createRuleContext`, which runs OUTSIDE the per-rule try/catch. Pre-
  // fix a function-valued option threw `DataCloneError` there and failed
  // the WHOLE file (parseError, zero diagnostics). Post-fix the rule
  // runs and its options reach `create()` intact.
  test('lintFile with a function-valued option does not throw and the rule runs', () => {
    const seen: { cb?: unknown; ran: boolean } = { ran: false };
    const optionFn = (): string => 'hi';
    const stub = {
      meta: { name: 'stub/opt-fn' },
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      create(ctx: any) {
        return {
          // eslint-disable-next-line @typescript-eslint/no-explicit-any
          VariableDeclaration(node: any) {
            seen.ran = true;
            seen.cb = ctx.options[0]?.cb;
            ctx.report({ node, message: 'opt-fn ran' });
          },
        };
      },
    };
    const loaded: LoadedPlugins = {
      plugins: [],
      rules: new Map<string, unknown>([['stub/opt-fn', stub]]),
    };
    let result: ReturnType<typeof lintFile> | undefined;
    expect(() => {
      result = lintFile(
        {
          filePath: 'opt.ts',
          text: 'var x = 1;',
          rules: { 'stub/opt-fn': { options: [{ cb: optionFn }] } },
          collectFixes: false,
          suggestionsMode: 'off',
        },
        loaded,
      );
    }).not.toThrow();
    // The rule ran (no whole-file parseError) and produced its diagnostic.
    expect(seen.ran).toBe(true);
    expect(result?.diagnostics).toHaveLength(1);
    expect(result?.diagnostics[0].message).toBe('opt-fn ran');
    // The function reached the rule by reference.
    expect(seen.cb).toBe(optionFn);
    // No rule-level error was recorded (the file did not fail).
    expect(result?.ruleErrors ?? []).toHaveLength(0);
  });
});

// ────────────────────────────────────────────────────────────────────
// WorkerPool init guards against spawn rejecting with `undefined`
// ────────────────────────────────────────────────────────────────────

describe('WorkerPool init treats any rejection as fatal (no undefined sentinel)', () => {
  test('spawn that rejects with `undefined` still throws from init()', async () => {
    const HANG_PATH = path.resolve(__dirname, 'fixtures', 'hang.config.mjs');
    const HANG_DIR = path.dirname(HANG_PATH);
    const pool = new WorkerPool({
      configs: [{ configPath: HANG_PATH, configDirectory: HANG_DIR }],
      workerCount: 2,
    });
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const internals = pool as any;
    let calls = 0;
    internals.spawnWorker = () => {
      calls++;
      // Mix: one resolves (well, also rejects with undefined),
      // another rejects with undefined. Pre-fix the sentinel
      // `firstFailure === undefined` skipped recording any of them as
      // a failure → init silently returned with `this.workers = []`
      // (under-provisioned pool, every subsequent lintBatch fails
      // mysteriously).
      return Promise.reject(undefined);
    };
    // Pre-fix: init() resolves with no workers. Post-fix: init()
    // rejects (a rejection is a rejection regardless of reason).
    let rejected = false;
    try {
      await pool.init();
    } catch {
      rejected = true;
    }
    expect(rejected).toBe(true);
    expect(calls).toBe(2);
    // Just in case pool resolved (pre-fix bug) — make sure we don't
    // leak workers in the assertions above.
    await pool.shutdown().catch(() => {});
  });
});

// apply-disable-directives — line-comment carriers for `eslint-disable`
// / `eslint-enable` are NOT directives (ESLint v10 only accepts
// `disable-line` / `disable-next-line` in line comments).
describe('line-comment eslint-disable / eslint-enable are NOT directives', () => {
  test('`// eslint-disable no-debugger` does NOT suppress diagnostics on later lines', () => {
    // ESLint v10's linter.js line 445: `if (comment.type === "Line" &&
    //   !lineCommentSupported) return;` where lineCommentSupported is
    //   /^eslint-disable-(next-)?line$/. So `// eslint-disable …` is
    //   silently ignored (NOT a directive); the diagnostic still fires.
    const loaded: LoadedPlugins = {
      plugins: [],
      rules: new Map<string, unknown>([
        [
          'stub/fire-on-debugger',
          {
            meta: { name: 'fire-on-debugger' },
            create(ctx: RuleContext) {
              return {
                DebuggerStatement(node: unknown) {
                  ctx.report({ node: node as never, message: 'debugger' });
                },
              };
            },
          },
        ],
      ]),
    };
    const result = lintFile(
      {
        filePath: 'r1-1.ts',
        text: '// eslint-disable stub/fire-on-debugger\ndebugger;\n',
        rules: { 'stub/fire-on-debugger': { options: [] } },
        collectFixes: false,
        suggestionsMode: 'off',
      },
      loaded,
    );
    expect(result.parseError).toBeUndefined();
    // Pre-fix: line-comment `eslint-disable` treated as block-disable →
    //   suppresses → 0 diagnostics.
    // Post-fix: only block comments carry `eslint-disable`; line carrier
    //   is ignored → 1 diagnostic.
    expect(result.diagnostics).toHaveLength(1);
  });

  test('`// eslint-enable` does NOT re-open a block-disable region', () => {
    const loaded: LoadedPlugins = {
      plugins: [],
      rules: new Map<string, unknown>([
        [
          'stub/fire-on-debugger',
          {
            meta: { name: 'fire-on-debugger' },
            create(ctx: RuleContext) {
              return {
                DebuggerStatement(node: unknown) {
                  ctx.report({ node: node as never, message: 'debugger' });
                },
              };
            },
          },
        ],
      ]),
    };
    const result = lintFile(
      {
        filePath: 'r1-1.ts',
        text:
          '/* eslint-disable stub/fire-on-debugger */\n' +
          '// eslint-enable stub/fire-on-debugger\n' +
          'debugger;\n',
        rules: { 'stub/fire-on-debugger': { options: [] } },
        collectFixes: false,
        suggestionsMode: 'off',
      },
      loaded,
    );
    expect(result.parseError).toBeUndefined();
    // Pre-fix: line `eslint-enable` mis-classified as block-enable, re-
    //   opens the region → 1 diagnostic.
    // Post-fix: line-comment `eslint-enable` ignored; the block-disable
    //   from line 1 stays in effect → 0 diagnostics.
    expect(result.diagnostics).toHaveLength(0);
  });

  test('`/* eslint-disable … */` block comment STILL suppresses (control)', () => {
    const loaded: LoadedPlugins = {
      plugins: [],
      rules: new Map<string, unknown>([
        [
          'stub/fire-on-debugger',
          {
            meta: { name: 'fire-on-debugger' },
            create(ctx: RuleContext) {
              return {
                DebuggerStatement(node: unknown) {
                  ctx.report({ node: node as never, message: 'debugger' });
                },
              };
            },
          },
        ],
      ]),
    };
    const result = lintFile(
      {
        filePath: 'r1-1.ts',
        text: '/* eslint-disable stub/fire-on-debugger */\ndebugger;\n',
        rules: { 'stub/fire-on-debugger': { options: [] } },
        collectFixes: false,
        suggestionsMode: 'off',
      },
      loaded,
    );
    expect(result.diagnostics).toHaveLength(0);
  });
});

// worker-pool.ts:793 — drainQueueIfAllSlotsDegraded mis-fires
//        during multi-worker mid-respawn (treats a transient
//        ready=false slot as permanently dead).
describe('drainQueueIfAllSlotsDegraded does not fire during mid-respawn', () => {
  test('mid-respawn worker prevents drain when another slot just terminally failed', async () => {
    const pool = new WorkerPool({
      configs: [
        { configPath: HANG_CONFIG_PATH, configDirectory: HANG_CONFIG_DIR },
      ],
      workerCount: 2,
      retryCap: 1, // worker A's first respawn-fail is terminal
      taskTimeoutMs: 30_000,
    });
    await pool.init();

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const internals = pool as any;

    // Hang task occupies BOTH workers? No — workerCount=2, single task
    // → only worker 0 gets it. Wedge BOTH with one task each by
    // dispatching 2 hangs first.
    const hangA = pool.lintBatch([
      {
        filePath: 'hangA.ts',
        text: 'const x = 1;\n',
        rules: { 'hang/hang': { options: [] } },
        collectFixes: false,
        suggestionsMode: 'off',
        configKey: HANG_CONFIG_DIR,
      },
    ]);
    const hangB = pool.lintBatch([
      {
        filePath: 'hangB.ts',
        text: 'const x = 1;\n',
        rules: { 'hang/hang': { options: [] } },
        collectFixes: false,
        suggestionsMode: 'off',
        configKey: HANG_CONFIG_DIR,
      },
    ]);
    await new Promise((r) => setTimeout(r, 200));

    // Queue the sibling — both workers busy → goes to pendingQueue.
    const sibling = pool.lintBatch([
      {
        filePath: 'sib.ts',
        text: 'const TRIGGER = 1;\n',
        rules: { 'hang/noop': { options: [] } },
        collectFixes: false,
        suggestionsMode: 'off',
        configKey: HANG_CONFIG_DIR,
      },
    ]);
    await new Promise((r) => setTimeout(r, 50));
    expect(internals.pendingQueue.length).toBe(1);

    // Worker B's respawn will succeed (real spawn). Worker A's respawn
    // is mocked to reject — terminal. So after the dust settles:
    //   - workers[0] dead (crashCount===retryCap, exited=true)
    //   - workers[1] alive (respawned successfully)
    // Pre-fix: when A's respawn rejects, drainQueueIfAllSlotsDegraded
    //   runs while B is mid-respawn (B.ready === false transiently);
    //   sibling gets `pool_degraded` even though B is about to come
    //   back. Post-fix: drain waits for B to settle and skips when B
    //   is alive.
    const originalSpawn = internals.spawnWorker.bind(internals);
    let spawnCallNum = 0;
    internals.spawnWorker = (id: number) => {
      spawnCallNum++;
      // Worker A (id=0): fail on respawn. Worker B (id=1): succeed.
      if (id === 0) {
        return Promise.reject(new Error('A respawn failed'));
      }
      return originalSpawn(id);
    };

    // Terminate both nearly simultaneously to trigger respawn race.
    await Promise.all([
      terminateWorker(internals.workers[0].worker),
      terminateWorker(internals.workers[1].worker),
    ]);
    void spawnCallNum;

    // Wait for sibling outcome.
    const TIMEOUT_MS = 8000;
    type Outcome =
      | { tag: 'sibling'; parseError: string | undefined }
      | { tag: 'timeout' };

    const outcome: Outcome = await Promise.race([
      sibling.then(
        (r): Outcome => ({ tag: 'sibling', parseError: r[0].parseError }),
      ),
      new Promise<Outcome>((r) =>
        setTimeout(() => r({ tag: 'timeout' }), TIMEOUT_MS),
      ),
    ]);

    expect(outcome.tag).toBe('sibling');
    // Post-fix contract: sibling completes normally (B respawned in
    // time and picked it up). Pre-fix: parseError === 'pool_degraded'.
    if (outcome.tag === 'sibling') {
      expect(outcome.parseError).toBeUndefined();
    }

    await Promise.race([hangA, new Promise((r) => setTimeout(r, 1000))]).catch(
      () => {},
    );
    await Promise.race([hangB, new Promise((r) => setTimeout(r, 1000))]).catch(
      () => {},
    );

    await pool.shutdown().catch(() => {});
  }, 30_000);
});

// diagnostic-builder.ts:412 — iterable branch of materializeFixes
//        skips per-element isFix filtering.
describe('materializeFixes iterable branch validates per element', () => {
  test('Set([junk, validFix]) returned by fix() does not crash; valid fix survives', () => {
    const loaded: LoadedPlugins = {
      plugins: [],
      rules: new Map<string, unknown>([
        [
          'stub/bad-set-fix',
          {
            meta: { name: 'bad-set-fix', fixable: 'code' },
            create(ctx: RuleContext) {
              return {
                VariableDeclaration(node: unknown) {
                  ctx.report({
                    node: node as never,
                    message: 'bad set fix',
                    fix() {
                      // eslint-disable-next-line @typescript-eslint/no-explicit-any
                      return new Set<any>([
                        { text: 'broken' }, // bogus — no range
                        { range: [0, 1] as [number, number], text: 'x' },
                      ]);
                    },
                  });
                },
              };
            },
          },
        ],
      ]),
    };

    let threw = false;
    let result: ReturnType<typeof lintFile> | undefined;
    try {
      result = lintFile(
        {
          filePath: 'r1-3-set.ts',
          text: 'const x = 1;\n',
          rules: { 'stub/bad-set-fix': { options: [] } },
          collectFixes: true,
          suggestionsMode: 'off',
        },
        loaded,
      );
    } catch {
      threw = true;
    }

    expect(threw).toBe(false);
    expect(result).toBeDefined();
    expect(result!.parseError).toBeUndefined();
    expect(result!.diagnostics).toHaveLength(1);
    const fixes = result!.diagnostics[0].fixes ?? [];
    expect(fixes).toHaveLength(1);
    expect(fixes[0].text).toBe('x');
  });

  test('generator yielding junk then valid does not crash', () => {
    const loaded: LoadedPlugins = {
      plugins: [],
      rules: new Map<string, unknown>([
        [
          'stub/bad-gen-fix',
          {
            meta: { name: 'bad-gen-fix', fixable: 'code' },
            create(ctx: RuleContext) {
              return {
                VariableDeclaration(node: unknown) {
                  ctx.report({
                    node: node as never,
                    message: 'bad gen fix',
                    // eslint-disable-next-line @typescript-eslint/no-explicit-any
                    fix: function* (): any {
                      yield { text: 'broken' };
                      yield { range: [0, 1] as [number, number], text: 'y' };
                    },
                  });
                },
              };
            },
          },
        ],
      ]),
    };

    let threw = false;
    let result: ReturnType<typeof lintFile> | undefined;
    try {
      result = lintFile(
        {
          filePath: 'r1-3-gen.ts',
          text: 'const x = 1;\n',
          rules: { 'stub/bad-gen-fix': { options: [] } },
          collectFixes: true,
          suggestionsMode: 'off',
        },
        loaded,
      );
    } catch {
      threw = true;
    }

    expect(threw).toBe(false);
    expect(result!.parseError).toBeUndefined();
    expect(result!.diagnostics).toHaveLength(1);
    const fixes = result!.diagnostics[0].fixes ?? [];
    expect(fixes).toHaveLength(1);
    expect(fixes[0].text).toBe('y');
  });
});

// listener-merge.ts:623,668 — byType and wildcard listeners
//        fire in two separate passes; ESLint v10 specificity-sorts them
//        as ONE list per node.
describe('byType and wildcard listeners interleave by specificity', () => {
  test('wildcard `:not(CallExpression)` (lower specificity) fires BEFORE bare Identifier on the same node', () => {
    // Record (rule, nodeType) per fire so we can isolate the events
    // for ONE node (the Identifier) and check their relative order.
    const order: Array<{ rule: string; type: string }> = [];
    const loaded: LoadedPlugins = {
      plugins: [],
      rules: new Map<string, unknown>([
        [
          'stub/bare',
          {
            meta: { name: 'bare' },
            create() {
              return {
                Identifier(node: unknown) {
                  // eslint-disable-next-line @typescript-eslint/no-explicit-any
                  order.push({ rule: 'bare', type: (node as any).type });
                },
              };
            },
          },
        ],
        [
          'stub/wild',
          {
            meta: { name: 'wild' },
            create() {
              return {
                ':not(CallExpression)'(node: unknown) {
                  // eslint-disable-next-line @typescript-eslint/no-explicit-any
                  order.push({ rule: 'wild', type: (node as any).type });
                },
              };
            },
          },
        ],
      ]),
    };
    lintFile(
      {
        filePath: 'r2-4.js',
        text: 'a;\n',
        rules: {
          'stub/bare': { options: [] },
          'stub/wild': { options: [] },
        },
        collectFixes: false,
        suggestionsMode: 'off',
      },
      loaded,
    );

    // Isolate just the Identifier-node firings so other nodes
    // (Program, ExpressionStatement) don't contaminate the order check.
    // `:not(CallExpression)` → attr=0, id=0. Bare `Identifier` →
    // attr=0, id=1. Lower specificity fires first → wildcard first.
    // Pre-fix: byType bucket ran completely before wildcard bucket
    // (order on Identifier: ['bare', 'wild']).
    // Post-fix: merge-sort by specificity (order on Identifier:
    // ['wild', 'bare']).
    const onIdentifier = order
      .filter((e) => e.type === 'Identifier')
      .map((e) => e.rule);
    expect(onIdentifier).toEqual(['wild', 'bare']);
  });
});

// listener-merge.ts:460 — `analyzeSpecificity` recurses through
//        `:has()` and double-counts identifiers; ESLint v10 does NOT
//        recurse into `:has()`.
describe('analyzeSpecificity does NOT recurse into :has()', () => {
  test('`CallExpression:has(Identifier MemberExpression)` has identifierCount=1', async () => {
    const { mergeListeners } = await import('../src/linter/listener-merge.js');
    const merged = mergeListeners([
      {
        ruleName: 'r2-5',
        listeners: {
          'CallExpression:has(Identifier MemberExpression)': () => {},
        },
      },
    ]);
    const entries = merged.esqueryByType.enter.get('CallExpression') ?? [];
    expect(entries).toHaveLength(1);
    // Pre-fix: walk recursed into `:has(...)` adding the inner
    // identifiers → identifierCount=3. Post-fix: `:has` falls to
    // default, no recursion → identifierCount=1 (the outer
    // CallExpression).
    expect(entries[0].identifierCount).toBe(1);
    expect(entries[0].attributeCount).toBe(0);
  });
});

// source-code — `eslint-env` must surface from getInlineConfigNodes,
// and Shebang comments must NOT be classified as directives.
describe('getInlineConfigNodes recognises eslint-env and skips Shebang', () => {
  test('`/* eslint-env browser */` surfaces in getInlineConfigNodes()', () => {
    let directives: Array<{ value: string }> = [];
    const loaded: LoadedPlugins = {
      plugins: [],
      rules: new Map<string, unknown>([
        [
          'stub/directive-probe',
          {
            meta: { name: 'directive-probe' },
            create(ctx: RuleContext) {
              return {
                Program() {
                  directives =
                    // eslint-disable-next-line @typescript-eslint/no-explicit-any
                    (ctx.sourceCode as any).getInlineConfigNodes() as Array<{
                      value: string;
                    }>;
                },
              };
            },
          },
        ],
      ]),
    };
    lintFile(
      {
        filePath: 'r2-10.js',
        text: '/* eslint-env browser */\nconst x = 1;\n',
        rules: { 'stub/directive-probe': { options: [] } },
        collectFixes: false,
        suggestionsMode: 'off',
      },
      loaded,
    );
    // Pre-fix: LABELS list missing `eslint-env` → directive ignored.
    // Post-fix: surfaces.
    expect(directives).toHaveLength(1);
    expect(directives[0].value.trim()).toBe('eslint-env browser');
  });

  test('Shebang comment is NOT classified as a directive', () => {
    let directives: Array<{ value: string }> = [];
    const loaded: LoadedPlugins = {
      plugins: [],
      rules: new Map<string, unknown>([
        [
          'stub/directive-probe',
          {
            meta: { name: 'directive-probe' },
            create(ctx: RuleContext) {
              return {
                Program() {
                  directives =
                    // eslint-disable-next-line @typescript-eslint/no-explicit-any
                    (ctx.sourceCode as any).getInlineConfigNodes() as Array<{
                      value: string;
                    }>;
                },
              };
            },
          },
        ],
      ]),
    };
    lintFile(
      {
        filePath: 'r2-10.js',
        // Real-world shebang: `#!/usr/bin/env node`. Doesn't look like
        // an inline directive, but ESLint explicitly skips Shebang
        // comments regardless of value for v10 parity.
        text: '#!/usr/bin/env node\nconst x = 1;\n',
        rules: { 'stub/directive-probe': { options: [] } },
        collectFixes: false,
        suggestionsMode: 'off',
      },
      loaded,
    );
    expect(directives).toEqual([]);
  });
});

// tokenizer.ts:767 — `couldStartRegex` treats `)` and `}` as
//        always closing an expression (→ next `/` is division).
//        ESLint v10 uses parser state to know `if (x) /re/` is regex.
describe('regex after control-flow `)` and statement `}` is recognised', () => {
  test('`if (x) /re/.test(y)` — `/` after `)` of control-paren starts regex', () => {
    const src = 'if (x) /re/.test(y);\n';
    const { tokens } = tokenize(src, buildLineStartOffsets(src));
    // Pre-fix: `)` was unconditionally treated as expression-end, so
    //   `/re/.test(y)` decomposed into Punctuator(/), Identifier(re),
    //   Punctuator(/), Punctuator(.), …
    // Post-fix: tokenizer tracks which `(` was opened by a control
    //   keyword (if/while/for/switch/catch); its matching `)` then
    //   leaves the lexer in expression-prefix position.
    const regex = tokens.find((t) => t.type === 'RegularExpression');
    expect(regex).toBeDefined();
    expect(regex?.value).toBe('/re/');
  });

  test('control: `f(x) / 2` — `/` after call paren stays division', () => {
    const src = 'const k = f(x) / 2;\n';
    const { tokens } = tokenize(src, buildLineStartOffsets(src));
    // Call-expression close paren is NOT control. `/` after must be
    // division. Regression guard for the fix above.
    expect(tokens.find((t) => t.type === 'RegularExpression')).toBeUndefined();
  });

  test('`if (x) {} /re/.test(y)` — `/` after block `}` starts regex', () => {
    // Block `}` closes a statement, not an expression → next `/` is
    // a regex literal. ESLint / espree both produce a single
    // RegularExpression token. Pre-fix runner saw `}` → division.
    const src = 'if (x) {} /re/.test(y);\n';
    const { tokens } = tokenize(src, buildLineStartOffsets(src));
    const regex = tokens.find((t) => t.type === 'RegularExpression');
    expect(regex).toBeDefined();
    expect(regex?.value).toBe('/re/');
  });

  test('control: `{a: 1} / 2` — `/` after object-literal `}` stays division', () => {
    // In expression position `{a:1}` is an object literal; `/` after
    // it is division. This is the case the existing test at
    // `tokenizer-edge-cases.test.ts:61` accidentally claimed to test.
    const src = 'const k = ({a: 1} / 2);\n';
    const { tokens } = tokenize(src, buildLineStartOffsets(src));
    expect(tokens.find((t) => t.type === 'RegularExpression')).toBeUndefined();
    const slashIdx = tokens.findIndex(
      (t) => t.type === 'Punctuator' && t.value === '/',
    );
    expect(slashIdx).toBeGreaterThan(-1);
  });
});

// tokenizer.ts:502 — Unicode-escaped identifiers (a)
//        fall to the unknown-char skip branch; identifier scan misses
//        the leading `\`. value and range both wrong.
describe('Unicode-escape-prefixed identifier tokenizes correctly', () => {
  test('`var \\u0061 = 1;` — identifier value "a", range starts at backslash', () => {
    const src = 'var \\u0061 = 1;';
    const { tokens } = tokenize(src, buildLineStartOffsets(src));
    const ident = tokens.find(
      (t) => t.type === 'Identifier' || t.type === 'Keyword',
    );
    void ident;
    const identTok = tokens.find((t) => t.value === 'a');
    expect(identTok).toBeDefined();
    expect(identTok?.type).toBe('Identifier');
    // The backslash is at offset 4 (`var ` is 4 chars). The escape
    // span is `a` = 6 chars → range [4, 10].
    expect(identTok?.range).toEqual([4, 10]);
  });
});

// source-code.ts:379 — text retains BOM byte; all downstream
//        offsets are +1 UTF-16 / +3 UTF-8 shifted from ESLint.
describe('leading BOM is stripped before parsing / SourceCode.text', () => {
  test('`<BOM>var x = 1` — diagnostic startPos matches ESLint (BOM-free)', () => {
    let captured: { range?: [number, number]; text?: string } = {};
    const loaded: LoadedPlugins = {
      plugins: [],
      rules: new Map<string, unknown>([
        [
          'stub/probe-id',
          {
            meta: { name: 'probe-id' },
            create(ctx: RuleContext) {
              return {
                Identifier(node: unknown) {
                  // eslint-disable-next-line @typescript-eslint/no-explicit-any
                  const n = node as any;
                  if (n.name === 'x') {
                    captured.range = n.range;
                    captured.text = ctx.sourceCode.getText(n);
                  }
                },
              };
            },
          },
        ],
      ]),
    };
    lintFile(
      {
        filePath: 'r3-8.js',
        text: '﻿var x = 1;\n',
        rules: { 'stub/probe-id': { options: [] } },
        collectFixes: false,
        suggestionsMode: 'off',
      },
      loaded,
    );
    // Pre-fix: BOM kept in text → identifier `x` range [5, 6] (one
    // shifted right). Post-fix: BOM stripped → range [4, 5] (matches
    // ESLint and oxc-on-stripped).
    expect(captured.range).toEqual([4, 5]);
    expect(captured.text).toBe('x');
  });

  test('SourceCode.text does NOT contain the leading BOM character', () => {
    let observedText: string | undefined;
    const loaded: LoadedPlugins = {
      plugins: [],
      rules: new Map<string, unknown>([
        [
          'stub/probe-text',
          {
            meta: { name: 'probe-text' },
            create(ctx: RuleContext) {
              return {
                Program() {
                  observedText = ctx.sourceCode.text;
                },
              };
            },
          },
        ],
      ]),
    };
    lintFile(
      {
        filePath: 'r3-8.js',
        text: '﻿var x = 1;\n',
        rules: { 'stub/probe-text': { options: [] } },
        collectFixes: false,
        suggestionsMode: 'off',
      },
      loaded,
    );
    expect(observedText).toBeDefined();
    expect(observedText!.charCodeAt(0)).not.toBe(0xfeff);
    expect(observedText).toBe('var x = 1;\n');
  });

  test('SourceCode.hasBOM is still true when original file had a BOM', () => {
    let observedHasBOM: boolean | undefined;
    const loaded: LoadedPlugins = {
      plugins: [],
      rules: new Map<string, unknown>([
        [
          'stub/probe-bom',
          {
            meta: { name: 'probe-bom' },
            create(ctx: RuleContext) {
              return {
                Program() {
                  observedHasBOM = ctx.sourceCode.hasBOM;
                },
              };
            },
          },
        ],
      ]),
    };
    lintFile(
      {
        filePath: 'r3-8.js',
        text: '﻿var x = 1;\n',
        rules: { 'stub/probe-bom': { options: [] } },
        collectFixes: false,
        suggestionsMode: 'off',
      },
      loaded,
    );
    expect(observedHasBOM).toBe(true);
  });
});

// source-code-helpers.ts:206 — `sourceCode.getScope()` with no
//        node argument returns globalScope; ESLint v10 throws
//        TypeError. Aligns the strictness so plugin bugs surface
//        loudly instead of silently picking the wrong scope.
describe('sourceCode.getScope() with no node throws TypeError', () => {
  test('plugin calling ctx.sourceCode.getScope() — no args — fails the rule', () => {
    let observed: unknown;
    const loaded: LoadedPlugins = {
      plugins: [],
      rules: new Map<string, unknown>([
        [
          'stub/no-arg-get-scope',
          {
            meta: { name: 'no-arg-get-scope' },
            create(ctx: RuleContext) {
              return {
                Program() {
                  try {
                    // eslint-disable-next-line @typescript-eslint/no-explicit-any
                    (ctx.sourceCode as any).getScope();
                  } catch (e) {
                    observed = e;
                  }
                },
              };
            },
          },
        ],
      ]),
    };
    lintFile(
      {
        filePath: 'r4-9.ts',
        text: 'const x = 1;\n',
        rules: { 'stub/no-arg-get-scope': { options: [] } },
        collectFixes: false,
        suggestionsMode: 'off',
      },
      loaded,
    );
    expect(observed).toBeInstanceOf(TypeError);
    expect((observed as Error).message).toMatch(/Missing required argument/);
  });

  test('control: ctx.sourceCode.getScope(node) still works (no regression)', () => {
    let scope: unknown;
    const loaded: LoadedPlugins = {
      plugins: [],
      rules: new Map<string, unknown>([
        [
          'stub/with-arg',
          {
            meta: { name: 'with-arg' },
            create(ctx: RuleContext) {
              return {
                Program(node: unknown) {
                  scope = ctx.sourceCode.getScope(node as never);
                },
              };
            },
          },
        ],
      ]),
    };
    lintFile(
      {
        filePath: 'r4-9c.ts',
        text: 'const x = 1;\n',
        rules: { 'stub/with-arg': { options: [] } },
        collectFixes: false,
        suggestionsMode: 'off',
      },
      loaded,
    );
    expect(scope).toBeDefined();
  });
});

// ────────────────────────────────────────────────────────────────────
// WorkerPool init race: worker crashing mid-init is not orphaned
// ────────────────────────────────────────────────────────────────────

describe('WorkerPool: worker crashing mid-init is correctly respawned', () => {
  test('crash before all spawns settle → respawn finds original slot → pool ends up at workerCount', async () => {
    const pool = new WorkerPool({
      configs: [
        { configPath: HANG_CONFIG_PATH, configDirectory: HANG_CONFIG_DIR },
      ],
      workerCount: 2,
      retryCap: 3,
      taskTimeoutMs: 5_000,
    });

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const internals = pool as any;
    const originalSpawn = internals.spawnWorker.bind(internals);
    let firstSlot: {
      worker: import('node:worker_threads').Worker;
      id: number;
    } | null = null;
    let callIdx = 0;
    let releaseSecond!: () => void;
    const block = new Promise<void>((r) => {
      releaseSecond = r;
    });

    internals.spawnWorker = async (id: number) => {
      const slot = await originalSpawn(id);
      const myIdx = callIdx++;
      if (myIdx === 0) {
        firstSlot = slot;
      } else {
        // Second worker holds back so the first worker can crash mid-init,
        // exercising the race window between spawn-resolve and the OLD
        // `this.workers = fulfilled` assignment.
        await block;
      }
      return slot;
    };

    const initPromise = pool.init();

    // Wait for the first worker to settle into this.workers.
    let waited = 0;
    while ((internals.workers.length < 1 || !firstSlot) && waited < 5000) {
      await new Promise((r) => setTimeout(r, 50));
      waited += 50;
    }
    expect(firstSlot).not.toBeNull();

    // Terminate slot 0 BEFORE the second spawn resolves. This triggers
    // the exit handler's respawn path while `this.workers` only has
    // slot 0 in it. Pre-fix `this.workers` was empty until the final
    // assignment, so the respawn's findIndex returned -1 and the
    // replacement got terminated as an orphan.
    await terminateWorker(firstSlot!.worker);

    // Give the respawn a moment to complete.
    await new Promise((r) => setTimeout(r, 500));

    // Now let the second worker resolve so init() finishes.
    releaseSecond();
    await initPromise;

    // Post-fix: both slots end up alive. Pre-fix: slot 0's respawn
    // got orphaned, so only slot 1 is ready (readyCount === 1).
    const readyCount = (internals.workers as Array<{ ready: boolean }>).filter(
      (w) => w.ready,
    ).length;
    expect(readyCount).toBe(2);

    await pool.shutdown().catch(() => {
      /* best-effort cleanup */
    });
  }, 30_000);
});

// ────────────────────────────────────────────────────────────────────
// WorkerPool empty-configs fast path (matches `configs` JSDoc contract)
// ────────────────────────────────────────────────────────────────────

describe('WorkerPool: empty configs is a no-worker fast path', () => {
  test('init() with empty configs + default workerCount does NOT throw', async () => {
    // Pre-fix: default workerCount = cpuCount (≠0) → init() spawned
    // workers with configs:[] → lint-worker rejected (`configs[]
    // required`) → init() threw. Post-fix: empty configs forces the
    // effective workerCount to 0, so init() is a clean no-op.
    const pool = new WorkerPool({ configs: [] });
    let threw = false;
    try {
      await pool.init();
    } catch {
      threw = true;
    }
    expect(threw).toBe(false);
    await pool.shutdown();
  });

  test('lintBatch on empty pool returns empty per-file results', async () => {
    const pool = new WorkerPool({ configs: [] });
    await pool.init();
    const results = await pool.lintBatch([
      {
        filePath: 'a.ts',
        text: 'const x = 1;\n',
        rules: {},
        collectFixes: false,
        suggestionsMode: 'off',
      },
      {
        filePath: 'b.ts',
        text: 'const y = 2;\n',
        rules: {},
        collectFixes: false,
        suggestionsMode: 'off',
      },
    ]);
    // No plugins configured ⇒ no plugin diagnostics for any file
    // (consistent with ESLint: zero applicable rules → zero messages).
    // Per-file result is fully populated so downstream merge is
    // unchanged; no parseError (the files aren't broken, there's just
    // nothing to lint).
    expect(results).toHaveLength(2);
    for (const r of results) {
      expect(r.diagnostics).toEqual([]);
      expect(r.fixes).toEqual([]);
      expect(r.suggestionsCount).toBe(0);
      expect(r.cancelled).toBe(false);
      expect(r.parseError).toBeUndefined();
    }
    await pool.shutdown();
  });

  test('empty configs + explicit workerCount>0 still no-op (no work to do)', async () => {
    // Empty configs means no plugins to load — an explicit worker
    // count can't change that, and spawning would crash on init. The
    // pool ignores the explicit count and stays worker-free.
    const pool = new WorkerPool({ configs: [], workerCount: 4 });
    let threw = false;
    try {
      await pool.init();
    } catch {
      threw = true;
    }
    expect(threw).toBe(false);
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    expect(((pool as any).workers as unknown[]).length).toBe(0);
    await pool.shutdown();
  });

  test('shutdown on empty pool is a clean no-op', async () => {
    const pool = new WorkerPool({ configs: [] });
    await pool.init();
    // Should resolve immediately without waiting on any worker exit.
    const start = Date.now();
    await pool.shutdown();
    expect(Date.now() - start).toBeLessThan(1000);
  });
});

// ────────────────────────────────────────────────────────────────────
// tokenizer: string `\`+CRLF line continuation (espree parity)
// ────────────────────────────────────────────────────────────────────

describe('tokenizer: backslash + CRLF line continuation', () => {
  test('`"x\\<CR><LF>y"` is a single String token (not truncated)', () => {
    // \ + CRLF is a LineContinuation — the string spans the line break
    // as one token. Pre-fix `i += 2` skipped \+CR and the LF tripped
    // the line-terminator break, truncating the String and cascading
    // every downstream token on the line.
    const src = 'const s = "x\\\r\ny"; foo(s);';
    const { tokens } = tokenize(src, buildLineStartOffsets(src));
    const strings = tokens.filter((t) => t.type === 'String');
    // Exactly one String token spanning the whole literal.
    expect(strings).toHaveLength(1);
    expect(strings[0].value).toBe('"x\\\r\ny"');
    // Downstream tokens intact: `;` `foo` `(` `s` `)` `;`
    const kinds = tokens.map((t) => `${t.type}:${t.value}`);
    expect(kinds).toContain('Identifier:foo');
    expect(kinds).toContain('Identifier:s');
    // No phantom String swallowing the rest of the line.
    expect(strings.length).toBe(1);
  });

  test('control: `\\`+LF-only line continuation still one token', () => {
    const src = 'const s = "x\\\ny";';
    const { tokens } = tokenize(src, buildLineStartOffsets(src));
    expect(tokens.filter((t) => t.type === 'String')).toHaveLength(1);
  });
});

// ────────────────────────────────────────────────────────────────────
// tokenizer: numeric separator in exponent (espree parity)
// ────────────────────────────────────────────────────────────────────

describe('tokenizer: exponent accepts numeric separator', () => {
  for (const [src, expected] of [
    ['1e1_0', '1e1_0'],
    ['1e-1_0', '1e-1_0'],
    ['6.022e1_4', '6.022e1_4'],
  ] as const) {
    test(`\`${src}\` → single Numeric token`, () => {
      const { tokens } = tokenize(src, buildLineStartOffsets(src));
      const nums = tokens.filter((t) => t.type === 'Numeric');
      expect(nums).toHaveLength(1);
      expect(nums[0].value).toBe(expected);
      // No stray Identifier `_0` from a split.
      expect(tokens.some((t) => t.type === 'Identifier')).toBe(false);
    });
  }
});

// ────────────────────────────────────────────────────────────────────
// applyOptionDefaults: ESLint v10 — slot-level schema `default`s are NOT
// materialized into a positional slot (ajv `useDefaults` only fills
// missing PROPERTIES of an object already present, never creates a slot).
// ────────────────────────────────────────────────────────────────────

describe('applyOptionDefaults: slot-level schema default is not materialized', () => {
  test('schema default only, no user → [] (not materialized)', async () => {
    const { applyOptionDefaults } =
      await import('../src/linter/options-defaults.js');
    // ESLint v10: schema:[{enum,default:'always'}] with no user options
    // yields context.options === [] (schema default is NOT materialized
    // into the positional slot). Pre-fix runner returned ['always'].
    expect(applyOptionDefaults(undefined, [{ default: 'always' }])).toEqual([]);
  });

  test('object schema no default, no user → [] (not [{}])', async () => {
    const { applyOptionDefaults } =
      await import('../src/linter/options-defaults.js');
    expect(applyOptionDefaults(undefined, [{ type: 'object' }])).toEqual([]);
  });

  test('schema default + conflicting defaultOptions → defaultOptions wins', async () => {
    const { applyOptionDefaults } =
      await import('../src/linter/options-defaults.js');
    // ESLint v10: defaultOptions (author's current intent) wins over a
    // stale schema default. Pre-fix Step 1 filled 'always' then Step 2
    // kept it (mergeDefault preserves an already-present user value) →
    // runner wrongly returned ['always'].
    expect(
      applyOptionDefaults(undefined, [{ default: 'always' }], ['never']),
    ).toEqual(['never']);
  });

  test('defaultOptions only, no user → materialized (unchanged)', async () => {
    const { applyOptionDefaults } =
      await import('../src/linter/options-defaults.js');
    expect(
      applyOptionDefaults(undefined, [{ type: 'string' }], ['never']),
    ).toEqual(['never']);
  });

  test('object defaultOptions deep-merges under user value', async () => {
    const { applyOptionDefaults } =
      await import('../src/linter/options-defaults.js');
    // user wins per-key; default fills the rest (unicorn-style).
    expect(
      applyOptionDefaults(
        [{ checkNaN: false }],
        [{ type: 'object' }],
        [{ checkInfinity: false, checkNaN: true }],
      ),
    ).toEqual([{ checkInfinity: false, checkNaN: false }]);
  });
});

// ────────────────────────────────────────────────────────────────────
// applyOptionDefaults: ESLint v10 DOES materialize schema PROPERTY
// `default`s into an object slot that already exists (ajv
// `useDefaults: true` in `shared/ajv.js`, in-place fill that leaks into
// `context.options`). Each expectation below was pinned against the
// eslint@9.32.0 oracle (identical to v10.4.0 for these): a probe rule
// captures `JSON.parse(JSON.stringify(ctx.options))` from `create()`.
// ────────────────────────────────────────────────────────────────────

describe('applyOptionDefaults: schema property defaults fill existing object slots', () => {
  test('property default fills into an existing {} slot', async () => {
    const { applyOptionDefaults } =
      await import('../src/linter/options-defaults.js');
    // oracle: ['error', {}] + properties.mode.default='auto' → [{mode:'auto'}]
    expect(
      applyOptionDefaults(
        [{}],
        [{ type: 'object', properties: { mode: { default: 'auto' } } }],
      ),
    ).toEqual([{ mode: 'auto' }]);
  });

  test('nested property defaults recurse into nested object', async () => {
    const { applyOptionDefaults } =
      await import('../src/linter/options-defaults.js');
    // oracle: ['error', {a:{}}] + properties.a.properties.x.default=1
    //         → [{a:{x:1}}]
    expect(
      applyOptionDefaults(
        [{ a: {} }],
        [
          {
            type: 'object',
            properties: {
              a: { type: 'object', properties: { x: { default: 1 } } },
            },
          },
        ],
      ),
    ).toEqual([{ a: { x: 1 } }]);
  });

  test('user-supplied value wins over schema property default', async () => {
    const { applyOptionDefaults } =
      await import('../src/linter/options-defaults.js');
    // oracle: ['error', {mode:'x'}] + properties.mode.default='auto'
    //         → [{mode:'x'}] (user wins)
    expect(
      applyOptionDefaults(
        [{ mode: 'x' }],
        [{ type: 'object', properties: { mode: { default: 'auto' } } }],
      ),
    ).toEqual([{ mode: 'x' }]);
  });

  test('only the absent key is filled; user keys preserved', async () => {
    const { applyOptionDefaults } =
      await import('../src/linter/options-defaults.js');
    // oracle: ['error', {other:1}] + properties.{mode.default,other}
    //         → [{other:1, mode:'auto'}]
    expect(
      applyOptionDefaults(
        [{ other: 1 }],
        [
          {
            type: 'object',
            properties: { mode: { default: 'auto' }, other: {} },
          },
        ],
      ),
    ).toEqual([{ other: 1, mode: 'auto' }]);
  });

  test('object/array property default is deep-cloned (not shared)', async () => {
    const { applyOptionDefaults } =
      await import('../src/linter/options-defaults.js');
    const schema = [
      { type: 'object', properties: { list: { default: [1, 2] } } },
    ];
    const a = applyOptionDefaults([{}], schema) as Array<{ list: number[] }>;
    const b = applyOptionDefaults([{}], schema) as Array<{ list: number[] }>;
    // oracle: ['error', {}] + properties.list.default=[1,2] → [{list:[1,2]}]
    expect(a).toEqual([{ list: [1, 2] }]);
    // Mutating one call's filled default must not corrupt the schema's
    // default or the next call (the clone in step 2).
    a[0].list.push(99);
    expect(b).toEqual([{ list: [1, 2] }]);
    expect(schema[0].properties!.list.default).toEqual([1, 2]);
  });

  test('NO fabricated slot: severity-only stays [] even with a property default', async () => {
    const { applyOptionDefaults } =
      await import('../src/linter/options-defaults.js');
    // oracle: severity-only ('error') with properties.mode.default='auto'
    //         → [] (ajv fills properties, never creates a positional slot)
    const schema = [
      { type: 'object', properties: { mode: { default: 'auto' } } },
    ];
    expect(applyOptionDefaults(undefined, schema)).toEqual([]);
    expect(applyOptionDefaults([], schema)).toEqual([]);
  });

  test('property fill runs AFTER defaultOptions creates the slot', async () => {
    const { applyOptionDefaults } =
      await import('../src/linter/options-defaults.js');
    // oracle: defaultOptions:[{}] (creates slot 0) + property default
    //         → [{mode:'auto'}]. The fill must run on the merged array.
    expect(
      applyOptionDefaults(
        undefined,
        [{ type: 'object', properties: { mode: { default: 'auto' } } }],
        [{}],
      ),
    ).toEqual([{ mode: 'auto' }]);
  });

  test('defaultOptions value beats schema property default for same key', async () => {
    const { applyOptionDefaults } =
      await import('../src/linter/options-defaults.js');
    // oracle: defaultOptions:[{mode:'fromDefault'}] + property default
    //         'auto' → [{mode:'fromDefault'}] (key already present after
    //         step 1, so step 2 does not overwrite it).
    expect(
      applyOptionDefaults(
        undefined,
        [{ type: 'object', properties: { mode: { default: 'auto' } } }],
        [{ mode: 'fromDefault' }],
      ),
    ).toEqual([{ mode: 'fromDefault' }]);
  });

  test('array-schema items[] map per positional slot', async () => {
    const { applyOptionDefaults } =
      await import('../src/linter/options-defaults.js');
    // oracle: ['error','hi',{}] with schema {items:[str, {props.y.default:9}]}
    //         → ['hi', {y:9}] (slot 1 schema fills slot 1)
    expect(
      applyOptionDefaults(['hi', {}], {
        type: 'array',
        items: [
          { type: 'string' },
          { type: 'object', properties: { y: { default: 9 } } },
        ],
      }),
    ).toEqual(['hi', { y: 9 }]);
  });

  test('single items schema applies to every slot', async () => {
    const { applyOptionDefaults } =
      await import('../src/linter/options-defaults.js');
    // oracle: ['error',{},{}] with schema {items:{props.mode.default:auto}}
    //         → [{mode:'auto'},{mode:'auto'}]
    expect(
      applyOptionDefaults([{}, {}], {
        type: 'array',
        items: { type: 'object', properties: { mode: { default: 'auto' } } },
      }),
    ).toEqual([{ mode: 'auto' }, { mode: 'auto' }]);
  });
});

// ────────────────────────────────────────────────────────────────────
// WorkerPool: spawn-phase hard exit during init rejects promptly
// ────────────────────────────────────────────────────────────────────

describe('WorkerPool: worker hard-exit during init rejects (not 60s hang)', () => {
  test('config that process.exit()s at import → init() rejects fast', async () => {
    const exitPath = path.resolve(
      __dirname,
      'fixtures',
      'exit-on-init.config.mjs',
    );
    const exitDir = path.dirname(exitPath);
    const pool = new WorkerPool({
      configs: [{ configPath: exitPath, configDirectory: exitDir }],
      workerCount: 1,
      // Long init timeout — the test asserts we DON'T wait for it.
      workerInitTimeoutMs: 30_000,
    });

    const start = Date.now();
    let err: Error | undefined;
    try {
      await pool.init();
    } catch (e) {
      err = e as Error;
    }
    const elapsed = Date.now() - start;

    // Rejected (not hung) and well under the 30s init timeout.
    expect(err).toBeDefined();
    expect(elapsed).toBeLessThan(10_000);
    // Message names the exit path, not a misleading "init timed out".
    expect(err!.message).toMatch(/exited during init|init failed/);
    expect(err!.message).not.toMatch(/timed out/);

    await pool.shutdown().catch(() => {});
  }, 35_000);
});

// ────────────────────────────────────────────────────────────────────
// WorkerPool: shutdown awaits an in-flight respawn (no orphan thread)
// ────────────────────────────────────────────────────────────────────

describe('WorkerPool: shutdown awaits in-flight respawn', () => {
  test('crash → respawn-in-flight → shutdown waits for the new worker teardown', async () => {
    const pool = new WorkerPool({
      configs: [
        { configPath: HANG_CONFIG_PATH, configDirectory: HANG_CONFIG_DIR },
      ],
      workerCount: 1,
      retryCap: 3,
      taskTimeoutMs: 30_000,
    });
    await pool.init();

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const internals = pool as any;

    // Make the respawn observably slow so the shutdown clearly has to
    // wait for it. Wrap spawnWorker to delay before delegating.
    const originalSpawn = internals.spawnWorker.bind(internals);
    let respawnStarted = false;
    internals.spawnWorker = async (id: number) => {
      respawnStarted = true;
      await new Promise((r) => setTimeout(r, 300));
      return originalSpawn(id);
    };

    // Crash the only worker → exit handler kicks off the (slow) respawn.
    await terminateWorker(internals.workers[0].worker);
    // Give the exit handler a tick to enter the respawn branch.
    await new Promise((r) => setTimeout(r, 50));
    expect(respawnStarted).toBe(true);
    // Respawn is registered as in-flight.
    expect(internals.respawns.size).toBeGreaterThan(0);

    // Shutdown now — must await the in-flight respawn's teardown.
    await pool.shutdown();

    // After shutdown returns, no respawn is still in-flight (it was
    // awaited, and the closed-branch terminated the new worker).
    expect(internals.respawns.size).toBe(0);
    expect((internals.workers as unknown[]).length).toBe(0);
  }, 30_000);
});

// ════════════════════════════════════════════════════════════════════
// Token-classification & SourceCode/AST alignment with ESLint v10
// (espree@11.2.0 / eslint@10.4.0). Each group pins a contract that
// diverged from ESLint and is now load-bearing.
// ════════════════════════════════════════════════════════════════════

type TT = { type: string; value: string };
function rTok(src: string): TT[] {
  return tokenize(src, buildLineStartOffsets(src), {}).tokens.map((t) => ({
    type: t.type,
    value: t.value,
  }));
}
function eTok(src: string): TT[] {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  return (espree as any)
    .tokenize(src, { ecmaVersion: 'latest', sourceType: 'script' })
    .map((t: TT) => ({ type: t.type, value: t.value }));
}

describe('tokenizer: strict future-reserved words are Identifier (espree@11)', () => {
  for (const [src, word] of [
    ['obj.public;', 'public'],
    ['({ private: 1 });', 'private'],
    ['x.enum;', 'enum'],
    ['var implements = 1;', 'implements'],
    ['obj.interface;', 'interface'],
    ['obj.package;', 'package'],
    ['obj.protected;', 'protected'],
  ] as const) {
    test(`\`${src}\` — \`${word}\` is Identifier, runner matches espree`, () => {
      expect(rTok(src)).toEqual(eTok(src));
      expect(rTok(src).find((t) => t.value === word)?.type).toBe('Identifier');
    });
  }
  test('control: true keywords (class/delete/static/if) stay Keyword', () => {
    for (const [src, word] of [
      ['obj.class;', 'class'],
      ['obj.delete;', 'delete'],
      ['obj.static;', 'static'],
      ['obj.if;', 'if'],
    ] as const) {
      expect(rTok(src)).toEqual(eTok(src));
      expect(rTok(src).find((t) => t.value === word)?.type).toBe('Keyword');
    }
  });
});

describe('tokenizer: `/` after static/let/yield/await is division (espree@11)', () => {
  for (const src of [
    'obj.static / 2;',
    'var let; let / 2;',
    'function f(yield){ return yield / 2; }',
    'function f(await){ return await / 2; }',
  ]) {
    test(JSON.stringify(src), () => {
      expect(rTok(src)).toEqual(eTok(src));
      expect(
        rTok(src).find((t) => t.type === 'RegularExpression'),
      ).toBeUndefined();
    });
  }
});

describe('tokenizer: `/re/` after fn/class declaration body is regex (espree@11)', () => {
  for (const src of [
    'function f() {} /re/.test(x)',
    'class C {} /re/.test(x)',
    'if (x) {} /re/.test(x)', // control — already correct pre-fix
  ]) {
    test(JSON.stringify(src), () => {
      expect(rTok(src)).toEqual(eTok(src));
      expect(rTok(src).find((t) => t.type === 'RegularExpression')?.value).toBe(
        '/re/',
      );
    });
  }
});

const PROGRAM_STUB = (len: number) =>
  ({ type: 'Program', body: [], range: [0, len] }) as never;

describe('SourceCode synthesizes Shebang when oxc omits it (TS family)', () => {
  const SRC = '#!/usr/bin/env node\nconst x = 1;\n';
  test('empty parsedComments (TS family) → synthesized Shebang', () => {
    const sc = createSourceCode({
      text: SRC,
      ast: PROGRAM_STUB(SRC.length),
      parsedComments: [],
    });
    const comments = sc.getAllComments();
    expect(comments).toHaveLength(1);
    expect(comments[0].type).toBe('Shebang');
    expect(comments[0].value).toBe('/usr/bin/env node');
    expect(comments[0].range).toEqual([0, 19]);
  });
  test('oxc offset-0 Line (JS family) → relabeled Shebang, not duplicated', () => {
    const sc = createSourceCode({
      text: SRC,
      ast: PROGRAM_STUB(SRC.length),
      parsedComments: [
        { type: 'Line', value: '/usr/bin/env node', start: 0, end: 19 },
      ],
    });
    const comments = sc.getAllComments();
    expect(comments).toHaveLength(1);
    expect(comments[0].type).toBe('Shebang');
  });
});

describe('SourceCode.getLocFromIndex validates input (eslint@10)', () => {
  const sc = createSourceCode({ text: 'abcde', ast: PROGRAM_STUB(5) });
  test('out-of-range index → RangeError', () => {
    expect(() => sc.getLocFromIndex(999)).toThrow(RangeError);
    expect(() => sc.getLocFromIndex(-1)).toThrow(RangeError);
  });
  test('non-number index → TypeError', () => {
    // @ts-expect-error runtime guard against a wrong-typed argument
    expect(() => sc.getLocFromIndex('5')).toThrow(TypeError);
    // @ts-expect-error runtime guard against a wrong-typed argument
    expect(() => sc.getLocFromIndex(undefined)).toThrow(TypeError);
  });
  test('valid index (incl. EOF one-past-last) returns loc', () => {
    expect(sc.getLocFromIndex(0)).toEqual({ line: 1, column: 0 });
    expect(sc.getLocFromIndex(5)).toEqual({ line: 1, column: 5 });
  });
});

describe('SourceCode.getIndexFromLoc non-last-line bound (eslint@10)', () => {
  const sc = createSourceCode({ text: 'abc\ndef', ast: PROGRAM_STUB(7) });
  test('column at a non-last line terminator → RangeError', () => {
    expect(() => sc.getIndexFromLoc({ line: 1, column: 4 })).toThrow(
      RangeError,
    );
  });
  test('column one-past-last char on non-last line is valid', () => {
    expect(sc.getIndexFromLoc({ line: 1, column: 3 })).toBe(3);
  });
  test('column one-past-last char on the LAST line (EOF) is valid', () => {
    expect(sc.getIndexFromLoc({ line: 2, column: 3 })).toBe(7);
  });
});

describe('normalize-ast: TemplateElement.range includes delimiters (espree@11)', () => {
  function collectTplRanges(root: unknown): Array<[number, number]> {
    const out: Array<[number, number]> = [];
    (function f(n: unknown): void {
      if (!n || typeof n !== 'object') return;
      const node = n as {
        type?: string;
        range?: [number, number];
        [k: string]: unknown;
      };
      if (node.type === 'TemplateElement' && node.range) out.push(node.range);
      for (const k of Object.keys(node)) if (k !== 'parent') f(node[k]);
    })(root);
    return out;
  }
  // The filename extension MATTERS: oxc-parser emits TemplateElement
  // `start`/`end` as COOKED (delimiter-less) offsets for JS extensions
  // but DELIMITER-INCLUSIVE offsets for TS extensions. The normalizer's
  // language-aware expansion must produce the delimiter-inclusive range
  // (matching both espree on JS and @typescript-eslint/parser on TS) in
  // either case. Earlier this helper hardcoded `t.js`, so the TS branch
  // — where the expansion DOUBLE-COUNTED the delimiters — was never
  // exercised. Parameterize the extension so both code paths are tested.
  function runnerTplRanges(
    src: string,
    ext: 'js' | 'ts' | 'tsx' = 'js',
  ): Array<[number, number]> {
    const parsed = parseSync(`t.${ext}`, src, { sourceType: 'module' });
    const ast =
      typeof parsed.program === 'string'
        ? JSON.parse(parsed.program as string)
        : parsed.program;
    normalizeAst(ast as never, buildLineStartOffsets(src), src);
    return collectTplRanges(ast);
  }
  function espreeTplRanges(src: string): Array<[number, number]> {
    return collectTplRanges(
      espree.parse(src, {
        ecmaVersion: 'latest',
        sourceType: 'module',
        range: true,
      }),
    );
  }
  const sortRanges = (rs: Array<[number, number]>): Array<[number, number]> =>
    [...rs].sort((a, b) => a[0] - b[0] || a[1] - b[1]);

  // ── JS: live diff against espree (range:true) ──────────────────────
  for (const src of [
    'const t = `x${a}y`;', // first / tail
    'const u = `abc`;', // single plain element
    'const v = `${a}`;', // two EMPTY elements (raw "")
    'const w = `a\\nb${x}c`;', // escape sequence in raw
    'const n = `o${`i${j}`}p`;', // nested template
  ]) {
    test(`js ${JSON.stringify(src)} matches espree`, () => {
      expect(sortRanges(runnerTplRanges(src, 'js'))).toEqual(
        sortRanges(espreeTplRanges(src)),
      );
    });
  }
  test('js headline pinned: `x${a}y` → [[10,14],[15,18]]', () => {
    expect(runnerTplRanges('const t = `x${a}y`;', 'js')).toEqual([
      [10, 14],
      [15, 18],
    ]);
  });

  // ── TS / TSX: diff against @typescript-eslint/parser ───────────────
  //
  // espree cannot parse TS, so the oracle is @typescript-eslint/parser
  // (parser.parse(src, {range:true, loc:true})). That package is not a
  // dependency of this runner package, so rather than import it here the
  // EXACT delimiter-inclusive ranges it produces are pinned below. They
  // were captured from @typescript-eslint/parser@8.59.4 + oxc-parser
  // @0.132.0 (the exact oxc version this package depends on) — both the
  // TS oracle range AND the runner's normalized range are the
  // delimiter-INCLUSIVE span, e.g. `` `x${ `` = [10,14] and `` }y` `` =
  // [15,18] for `const t = `x${a}y``. Pre-fix the runner expanded
  // unconditionally and DOUBLE-COUNTED on TS (here [9,16]/[14,19]).
  //
  // Ranges are pinned in the runner's DFS collect order (which equals
  // the parser's for flat templates); for the nested case the two
  // traversal orders differ, so it is compared as a sorted set — the
  // value correctness is what the fix is about, not traversal order.
  const tsExpected: Array<{
    label: string;
    src: string;
    ranges: Array<[number, number]>;
    sorted?: boolean;
  }> = [
    {
      label: 'first/tail',
      src: 'const t = `x${a}y`;',
      ranges: [
        [10, 14],
        [15, 18],
      ],
    },
    {
      label: 'single plain element',
      src: 'const u = `abc`;',
      ranges: [[10, 15]],
    },
    {
      label: 'two EMPTY elements (raw "")',
      src: 'const v = `${a}`;',
      ranges: [
        [10, 13],
        [14, 16],
      ],
    },
    {
      label: 'escape sequence in raw',
      src: 'const w = `a\\nb${x}c`;',
      ranges: [
        [10, 17],
        [18, 21],
      ],
    },
    {
      label: 'nested template',
      src: 'const n = `o${`i${j}`}p`;',
      // sorted set: outer-first [10,14], inner-first [14,18],
      // inner-tail [19,21], outer-tail [21,24].
      ranges: [
        [10, 14],
        [14, 18],
        [19, 21],
        [21, 24],
      ],
      sorted: true,
    },
  ];
  for (const { label, src, ranges, sorted } of tsExpected) {
    for (const ext of ['ts', 'tsx'] as const) {
      test(`${ext} ${label}: ${JSON.stringify(src)} → ${JSON.stringify(ranges)}`, () => {
        const got = runnerTplRanges(src, ext);
        if (sorted) {
          expect(sortRanges(got)).toEqual(sortRanges(ranges));
        } else {
          expect(got).toEqual(ranges);
        }
      });
    }
  }

  // Focused double-count regression: the same template under `.js` and
  // `.ts` MUST yield identical delimiter-inclusive ranges. Pre-fix the
  // `.ts` ranges were each 2 wider (start −1, end +1) than the `.js`
  // ranges because oxc's TS offsets already include the delimiters and
  // the normalizer expanded them a second time.
  test('TS does not double-count delimiters: .ts ranges == .js ranges', () => {
    const src = 'const t = `x${a}y`;';
    expect(runnerTplRanges(src, 'ts')).toEqual(runnerTplRanges(src, 'js'));
    expect(runnerTplRanges(src, 'tsx')).toEqual(runnerTplRanges(src, 'js'));
    // And the wrong (double-counted) span is NOT emitted. Pre-fix the
    // `.ts` first element was [9,16] (oxc's delimited [10,14] expanded
    // start−1, end+2) and the tail [14,19] (delimited [15,18] expanded
    // start−1, end+1).
    expect(runnerTplRanges(src, 'ts')).not.toEqual([
      [9, 16],
      [14, 19],
    ]);
  });
});

describe('report() multiple fixes merge into one atomic Fix (eslint@10)', () => {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  function lintWithFix(text: string, fixFn: (f: any, node: any) => unknown) {
    const loaded: LoadedPlugins = {
      plugins: [],
      rules: new Map<string, unknown>([
        [
          't/merge',
          {
            meta: { fixable: 'code' },
            // eslint-disable-next-line @typescript-eslint/no-explicit-any
            create(ctx: any) {
              return {
                // eslint-disable-next-line @typescript-eslint/no-explicit-any
                VariableDeclaration(node: any) {
                  ctx.report({
                    node,
                    message: 'm',
                    // eslint-disable-next-line @typescript-eslint/no-explicit-any
                    fix: (f: any) => fixFn(f, node),
                  });
                },
              };
            },
          },
        ],
      ]),
    };
    return lintFile(
      {
        filePath: 'm.ts',
        text,
        rules: { 't/merge': { options: [] } },
        collectFixes: true,
        suggestionsMode: 'off',
      },
      loaded,
    );
  }
  test('insertBefore + insertAfter → single {range,text}', () => {
    const r = lintWithFix('var x = 1;', (f, node) => [
      f.insertTextBefore(node, 'A'),
      f.insertTextAfter(node, 'B'),
    ]);
    expect(r.diagnostics).toHaveLength(1);
    const fixes = r.diagnostics[0].fixes ?? [];
    expect(fixes).toHaveLength(1);
    expect(fixes[0].range).toEqual([0, 10]);
    expect(fixes[0].text).toBe('Avar x = 1;B');
  });
  test('two replaceTextRange → single fix, original text spliced between', () => {
    const r = lintWithFix('var x = 1;', (f) => [
      f.replaceTextRange([0, 3], 'let'),
      f.replaceTextRange([4, 5], 'y'),
    ]);
    const fixes = r.diagnostics[0].fixes ?? [];
    expect(fixes).toHaveLength(1);
    expect(fixes[0].range).toEqual([0, 5]);
    expect(fixes[0].text).toBe('let y');
  });

  // ESLint v10 THROWS "Fix objects must not be overlapped in a report."
  // when a single report returns overlapping fixes (verified against
  // eslint@10.4.0). Throwing in the runner would let the worker's
  // catch-all wipe every diagnostic on the file, so instead the fix is
  // DROPPED and the diagnostic still fires — same net effect (fix not
  // applied) without losing the diagnostic. Pre-fix `mergeFixes`
  // silently stitched the overlap into a bogus `{range,text}`.
  test('overlapping fixes are dropped; diagnostic still reports', () => {
    const r = lintWithFix('var x = 1;', (f) => [
      // [0,5] and [3,8] overlap (3 < 5).
      f.replaceTextRange([0, 5], 'AAAAA'),
      f.replaceTextRange([3, 8], 'BBBBB'),
    ]);
    expect(r.diagnostics).toHaveLength(1);
    expect(r.diagnostics[0].message).toBe('m');
    // No bogus merged fix emitted.
    const fixes = r.diagnostics[0].fixes;
    expect(fixes == null || fixes.length === 0).toBe(true);
  });

  test('a fully contained fix counts as overlap and is dropped', () => {
    const r = lintWithFix('var x = 1;', (f) => [
      // [3,5] is contained in [0,8] → overlap.
      f.replaceTextRange([0, 8], 'X'),
      f.replaceTextRange([3, 5], 'Y'),
    ]);
    expect(r.diagnostics).toHaveLength(1);
    const fixes = r.diagnostics[0].fixes;
    expect(fixes == null || fixes.length === 0).toBe(true);
  });

  // Control: adjacent (non-overlapping) fixes must NOT be treated as an
  // overlap — they still merge, matching ESLint. Guards the `<` (not
  // `<=`) boundary in the overlap check.
  test('adjacent fixes (touching ends) still merge, NOT dropped', () => {
    const r = lintWithFix('var x = 1;', (f) => [
      f.replaceTextRange([0, 3], 'let'),
      f.replaceTextRange([3, 5], ' Y'),
    ]);
    const fixes = r.diagnostics[0].fixes ?? [];
    expect(fixes).toHaveLength(1);
    expect(fixes[0].range).toEqual([0, 5]);
    expect(fixes[0].text).toBe('let Y');
  });
});

describe('WorkerPool swallows stdout/stderr pipe errors (Windows teardown guard)', () => {
  test('emitting `error` on a worker log pipe does not throw / crash the host', async () => {
    const pool = new WorkerPool({
      configs: [
        { configPath: HANG_CONFIG_PATH, configDirectory: HANG_CONFIG_DIR },
      ],
      workerCount: 1,
    });
    await pool.init();
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const w = (pool as any).workers[0].worker as {
      stdout: import('node:stream').Readable;
      stderr: import('node:stream').Readable;
    };
    // On Windows, `worker.terminate()` destroys these pipes and emits
    // `'error'`. A stream `'error'` with no listener is re-thrown by
    // Node (uncaught) and crashes the host — `emit('error')` on a
    // listener-less stream throws synchronously, so a regression here
    // surfaces as a thrown error. With the spawn-time listener it's a
    // no-op on every platform.
    expect(() =>
      w.stdout.emit('error', new Error('simulated pipe EPIPE')),
    ).not.toThrow();
    expect(() =>
      w.stderr.emit('error', new Error('simulated pipe EPIPE')),
    ).not.toThrow();
    await pool.shutdown();
  });
});
