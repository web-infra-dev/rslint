/**
 * Regression tests for the post-review ESLint-API alignment fixes
 * (N4-a / N4-b / N4-c / N5). The reference ESLint behavior was
 * verified directly against the upstream source:
 *
 *   - SourceCode#getFirstTokens etc. options shape:
 *       lib/languages/js/source-code/token-store/index.js
 *   - SourceCode#getAncestors / isSpaceBetween:
 *       lib/languages/js/source-code/source-code.js
 *   - RuleContext#report multi-arg form:
 *       lib/linter/file-report.js (normalizeMultiArgReportCall)
 *
 * Each test drives `lintFile` end-to-end so the assertions exercise
 * the same code path real plugin rules hit — not just the helper in
 * isolation.
 */
import { describe, test, expect } from '@rstest/core';

import { lintFile } from '../../../src/eslint-plugin/linter/ecma-language-plugin.js';
import type { LoadedPlugins } from '../../../src/eslint-plugin/plugin/plugin-loader.js';
import type { RuleContext } from '../../../src/eslint-plugin/linter/context.js';

interface Observation<T = unknown> {
  value: T;
}

function loadWithRule(create: (ctx: RuleContext) => unknown): LoadedPlugins {
  return {
    plugins: [],
    rules: new Map<string, unknown>([
      [
        'stub/probe',
        {
          meta: { name: 'probe' },
          create,
        },
      ],
    ]),
  };
}

// ────────────────────────────────────────────────────────────────
// N4-a: token API options overload
// ────────────────────────────────────────────────────────────────

describe('N4-a: token API options overload', () => {
  test('getFirstTokens accepts a number count', () => {
    const obs: Observation = { value: undefined };
    const loaded = loadWithRule((ctx) => ({
      Program(node: unknown) {
        const tokens = ctx.sourceCode.getFirstTokens(node as never, 3);
        obs.value = tokens.map((t) => t.value);
      },
    }));

    lintFile(
      {
        filePath: 'count.js',
        text: 'const a = 1; const b = 2;',
        rules: { 'stub/probe': { options: [] } },
        collectFixes: false,
        suggestionsMode: 'off',
      },
      loaded,
    );
    // First three tokens of `const a = 1; const b = 2;` are
    // `const`, `a`, `=`. The exact rule that matters here is "we
    // honored the count cap" — we don't pin token-stream identity
    // beyond that.
    expect(Array.isArray(obs.value)).toBe(true);
    expect((obs.value as string[]).length).toBe(3);
    expect((obs.value as string[])[0]).toBe('const');
  });

  test('getFirstTokens accepts an options object with filter', () => {
    const obs: Observation = { value: undefined };
    const loaded = loadWithRule((ctx) => ({
      Program(node: unknown) {
        const tokens = ctx.sourceCode.getFirstTokens(node as never, {
          count: 10,
          filter: (t) => t.type === 'Identifier',
        });
        obs.value = tokens.map((t) => t.value);
      },
    }));

    lintFile(
      {
        filePath: 'filter.js',
        text: 'const a = 1; const b = 2; const c = 3;',
        rules: { 'stub/probe': { options: [] } },
        collectFixes: false,
        suggestionsMode: 'off',
      },
      loaded,
    );
    // Only identifiers survive the filter; we expect a, b, c.
    expect(obs.value).toEqual(['a', 'b', 'c']);
  });

  test('getFirstTokens with includeComments interleaves Block comments', () => {
    // Use the WHOLE Program as the node — every code + comment token
    // inside the file is in range. We assert the stream contains at
    // least one Block (the `/* mid */` comment) iff includeComments
    // is on, and zero Blocks otherwise. This pins the
    // includeComments contract without depending on
    // VariableDeclaration's exact range semantics.
    const obs: { withComments?: string[]; withoutComments?: string[] } = {};
    const loaded = loadWithRule((ctx) => ({
      Program(node: unknown) {
        obs.withoutComments = ctx.sourceCode
          .getFirstTokens(node as never, { count: 50 })
          .map((t) => t.type);
        obs.withComments = ctx.sourceCode
          .getFirstTokens(node as never, {
            count: 50,
            includeComments: true,
          })
          .map((t) => t.type);
      },
    }));

    lintFile(
      {
        filePath: 'comments.js',
        text: 'const a = 1; /* mid */ const b = 2;',
        rules: { 'stub/probe': { options: [] } },
        collectFixes: false,
        suggestionsMode: 'off',
      },
      loaded,
    );

    const without = obs.withoutComments ?? [];
    const withC = obs.withComments ?? [];
    expect(without).not.toContain('Block');
    expect(withC).toContain('Block');
    // includeComments adds exactly one entry (the lone block comment).
    expect(withC.length).toBe(without.length + 1);
  });

  // R2: `getTokens(node, { includeComments })` must honor the
  // `includeComments` flag — previously the implementation used a
  // `filterFn` helper that only extracted `.filter` and silently
  // discarded `includeComments`, so callers like `@stylistic` rules
  // that legitimately pass `{ includeComments: true }` got only code
  // tokens back. (The pre-fix test bearing this name actually called
  // `getFirstTokens`, which had always honored the flag — a false
  // coverage signal.)
  test('R2: getTokens(node, { includeComments: true }) splices comment tokens', () => {
    const obs: { withComments?: string[]; withoutComments?: string[] } = {};
    const loaded = loadWithRule((ctx) => ({
      Program(node: unknown) {
        obs.withoutComments = ctx.sourceCode
          .getTokens(node as never)
          .map((t) => t.type);
        obs.withComments = ctx.sourceCode
          .getTokens(node as never, { includeComments: true })
          .map((t) => t.type);
      },
    }));

    lintFile(
      {
        filePath: 'comments.js',
        text: 'const a = 1; /* mid */ const b = 2;',
        rules: { 'stub/probe': { options: [] } },
        collectFixes: false,
        suggestionsMode: 'off',
      },
      loaded,
    );

    const without = obs.withoutComments ?? [];
    const withC = obs.withComments ?? [];
    expect(without).not.toContain('Block');
    expect(withC).toContain('Block');
    expect(withC.length).toBe(without.length + 1);
  });

  test('R2: getTokensBetween(left, right, { includeComments }) splices comments', () => {
    const obs: { withComments?: string[]; withoutComments?: string[] } = {};
    const loaded = loadWithRule((ctx) => ({
      Program(node: unknown) {
        const prog = node as {
          body: ReadonlyArray<{ range: [number, number] }>;
        };
        const [left, right] = [prog.body[0], prog.body[1]];
        obs.withoutComments = ctx.sourceCode
          .getTokensBetween(left as never, right as never)
          .map((t) => t.type);
        obs.withComments = ctx.sourceCode
          .getTokensBetween(left as never, right as never, {
            includeComments: true,
          })
          .map((t) => t.type);
      },
    }));

    lintFile(
      {
        filePath: 'comments.js',
        // Two top-level statements with a block comment strictly
        // between them — no `;` between the comment and the second
        // statement, so the comment is in range (left.range[1] ..
        // right.range[0]).
        text: 'const a = 1; /* mid */ const b = 2;',
        rules: { 'stub/probe': { options: [] } },
        collectFixes: false,
        suggestionsMode: 'off',
      },
      loaded,
    );

    const without = obs.withoutComments ?? [];
    const withC = obs.withComments ?? [];
    expect(without).not.toContain('Block');
    expect(withC).toContain('Block');
    expect(withC.length).toBe(without.length + 1);
  });

  test('R2: getTokens honors `filter` AND `includeComments` together', () => {
    // filter callback receives both code AND comment tokens when
    // includeComments is on. A filter that whitelists only `Block`
    // type returns just the comment(s).
    let observed: string[] | undefined;
    const loaded = loadWithRule((ctx) => ({
      Program(node: unknown) {
        observed = ctx.sourceCode
          .getTokens(node as never, {
            includeComments: true,
            // ESLint's Token union and Comment union are nominally
            // distinct; `Block` is the comment-side type tag, and the
            // SourceCode getter returns the union at runtime. Cast
            // through `unknown` so TS doesn't complain about the
            // overlap check.
            filter: (t) => (t.type as unknown as string) === 'Block',
          })
          .map((t) => t.type);
      },
    }));

    lintFile(
      {
        filePath: 'comments.js',
        text: 'const a = 1; /* mid */ const b = 2;',
        rules: { 'stub/probe': { options: [] } },
        collectFixes: false,
        suggestionsMode: 'off',
      },
      loaded,
    );

    expect(observed).toEqual(['Block']);
  });
});

// ────────────────────────────────────────────────────────────────
// N4-b: sourceCode.getAncestors(node) via node.parent chain
// ────────────────────────────────────────────────────────────────
//
// ESLint v9 removed `context.getAncestors()` entirely; the canonical
// API is now `sourceCode.getAncestors(node)`. rslint aligns with v9
// and only ships that form. Plugins that still call the legacy
// `context.getAncestors()` will (correctly) hit a TypeError —
// matching upstream behavior.

describe('N4-b: sourceCode.getAncestors via node.parent chain', () => {
  test('returns the live ancestor chain when called with a node', () => {
    const obs: Observation = { value: undefined };
    const loaded = loadWithRule((ctx) => ({
      Literal(node: unknown) {
        // The Literal `1` sits inside an Initializer of a
        // VariableDeclarator inside a VariableDeclaration inside
        // Program. v9: `sourceCode.getAncestors(node)`.
        const ancestors = ctx.sourceCode.getAncestors(node as never);
        obs.value = ancestors.map((a: unknown) => (a as { type: string }).type);
      },
    }));

    lintFile(
      {
        filePath: 'ancestors.js',
        text: 'const x = 1;',
        rules: { 'stub/probe': { options: [] } },
        collectFixes: false,
        suggestionsMode: 'off',
      },
      loaded,
    );

    // Order is root-first (matches ESLint), so Program is at index 0.
    const types = obs.value as string[];
    expect(types[0]).toBe('Program');
    expect(types).toContain('VariableDeclaration');
    expect(types).toContain('VariableDeclarator');
    // The Literal itself is NOT in its own ancestor list.
    expect(types).not.toContain('Literal');
  });

  test('context.getAncestors no longer exists (aligned with ESLint v9+)', () => {
    const obs: Observation<string> = { value: 'unset' };
    const loaded = loadWithRule((ctx) => {
      // typeof of a removed property is "undefined". This documents
      // the removal as a positive test, so a future "let's add it
      // back" attempt fails this assertion deliberately.
      obs.value = typeof (ctx as { getAncestors?: unknown }).getAncestors;
      return {};
    });

    lintFile(
      {
        filePath: 'removed.js',
        text: 'const x = 1;',
        rules: { 'stub/probe': { options: [] } },
        collectFixes: false,
        suggestionsMode: 'off',
      },
      loaded,
    );
    expect(obs.value).toBe('undefined');
  });
});

// ────────────────────────────────────────────────────────────────
// N4-c: context.report positional form
// ────────────────────────────────────────────────────────────────

describe('N4-c: context.report accepts legacy positional form', () => {
  test('report(node, message) produces a diagnostic equivalent to the descriptor form', () => {
    // Stub rule that uses the LEGACY positional form. If
    // normalizeReportArgs is missing or broken, the diagnostic
    // would be silently dropped (descriptor.node === undefined →
    // buildDiagnostic returns null).
    const loaded = loadWithRule((ctx) => ({
      Identifier(node: unknown) {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        (ctx.report as any)(node, 'flag identifier');
      },
    }));

    const result = lintFile(
      {
        filePath: 'positional.js',
        text: 'const x = 1;',
        rules: { 'stub/probe': { options: [] } },
        collectFixes: false,
        suggestionsMode: 'off',
      },
      loaded,
    );

    // The file has exactly one Identifier (`x`), so we get one
    // diagnostic at its location.
    expect(result.diagnostics.length).toBeGreaterThan(0);
    expect(result.diagnostics[0].message).toBe('flag identifier');
  });

  test('report(node, message, data) substitutes placeholders', () => {
    const loaded = loadWithRule((ctx) => ({
      Identifier(node: unknown) {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        (ctx.report as any)(node, 'name is {{n}}', { n: 'rslint' });
      },
    }));

    const result = lintFile(
      {
        filePath: 'positional-data.js',
        text: 'const x = 1;',
        rules: { 'stub/probe': { options: [] } },
        collectFixes: false,
        suggestionsMode: 'off',
      },
      loaded,
    );
    expect(result.diagnostics[0].message).toBe('name is rslint');
  });
});

// ────────────────────────────────────────────────────────────────
// N5: isSpaceBetween treats comments as non-space (ESLint parity)
// ────────────────────────────────────────────────────────────────

describe('N5: isSpaceBetween mirrors ESLint token-walk semantics', () => {
  // All fixtures use the same template `let _ = (<LEFT><MID><RIGHT>);`
  // so the source is a valid Program. The probe picks the LEFT/RIGHT
  // identifiers from the token stream by their literal value.
  function checkSpaceBetween(source: string): boolean | null {
    const obs: Observation<boolean | null> = { value: null };
    const loaded = loadWithRule((ctx) => ({
      Program() {
        const tokens = ctx.sourceCode.getTokens(ctx.sourceCode.ast as never);
        const a = tokens.find((t) => t.value === 'aa');
        const b = tokens.find((t) => t.value === 'bb');
        if (a && b) {
          obs.value = ctx.sourceCode.isSpaceBetween(a as never, b as never);
        }
      },
    }));
    lintFile(
      {
        filePath: 'spacing.js',
        text: source,
        rules: { 'stub/probe': { options: [] } },
        collectFixes: false,
        suggestionsMode: 'off',
      },
      loaded,
    );
    return obs.value;
  }

  test('flush comment between two adjacent identifiers → false (no space)', () => {
    // `aa/*c*/bb` is invalid on its own (two adjacent identifiers),
    // so wrap in a function call: `f(aa/*c*/, bb)`. The comment sits
    // flush against `aa` but NOT against the closing paren — what we
    // need is a comment flush against BOTH neighbouring tokens, which
    // requires wrapping like `aa,/*c*/bb` inside an array literal.
    expect(checkSpaceBetween('let _ = [aa,/*c*/bb];')).toBe(false);
  });

  test('comment with surrounding whitespace → true', () => {
    expect(checkSpaceBetween('let _ = [aa, /*c*/ bb];')).toBe(true);
  });

  test('directly adjacent tokens with no whitespace at all → false', () => {
    expect(checkSpaceBetween('let _ = [aa,bb];')).toBe(false);
  });
});

// ────────────────────────────────────────────────────────────────
// context.report({ loc }) — descriptor shape tolerance
// ────────────────────────────────────────────────────────────────
//
// ESLint v9's `lib/linter/file-report.js` accepts `loc` in any of:
//   - `{ line, column }`           — zero-width at point
//   - `{ start, end }`             — full range
//   - `{ start }`                  — partial range; end defaults to start
//
// Pre-fix rslint crashed on the third form ("Cannot read properties
// of undefined (reading 'line')") because it read `loc.end.line`
// without checking. The listener wrapper caught the throw and routed
// it to `ruleErrors`, so the user only saw "rule failed" with no
// information about the intended report — a real plugin-compat bug.
//
// Each test below drives lintFile end-to-end and asserts that the
// expected diagnostic actually surfaces (and that no rule-error
// records show up — those would indicate a thrown exception was
// silently swallowed).

describe('context.report({ loc }) shape tolerance', () => {
  function reportAtLoc<T>(
    loc: T,
    source = 'const x = 1;\nconst y = 2;\n',
  ): {
    diagnostics: Array<{ startPos: number; endPos: number }>;
    ruleErrors: unknown;
  } {
    const loaded = loadWithRule((ctx) => ({
      Program(node: unknown) {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        ctx.report({ loc: loc as any, node: node as any, message: 'point' });
      },
    }));
    const result = lintFile(
      {
        filePath: 'partial.js',
        text: source,
        rules: { 'stub/probe': { options: [] } },
        collectFixes: false,
        suggestionsMode: 'off',
      },
      loaded,
    );
    return {
      diagnostics: result.diagnostics.map((d) => ({
        startPos: d.startPos,
        endPos: d.endPos,
      })),
      ruleErrors: result.ruleErrors,
    };
  }

  // Source is `const x = 1;\nconst y = 2;\n`. Line 1 starts at offset 0,
  // so column 6 → absolute offset 6 (the `x`), column 7 → offset 7.
  // ASCII text means the UTF-8 byte offset on the wire equals the
  // UTF-16 offset. Assert the ABSOLUTE positions (not just the relative
  // width) so a column off-by-one that shifts BOTH ends is caught.
  test('loc { start, end } — full range pinned to absolute offsets', () => {
    const r = reportAtLoc({
      start: { line: 1, column: 6 },
      end: { line: 1, column: 7 },
    });
    expect(r.ruleErrors).toBeUndefined();
    expect(r.diagnostics).toHaveLength(1);
    expect(r.diagnostics[0].startPos).toBe(6);
    expect(r.diagnostics[0].endPos).toBe(7);
  });

  test('loc { start } only — end defaults to start (zero-width) at absolute offset; no crash', () => {
    const r = reportAtLoc({ start: { line: 1, column: 6 } });
    // The critical contract: no rule-error (rule did NOT throw), and
    // the diagnostic surfaced as a zero-width report pinned at offset 6.
    expect(r.ruleErrors).toBeUndefined();
    expect(r.diagnostics).toHaveLength(1);
    expect(r.diagnostics[0].startPos).toBe(6);
    expect(r.diagnostics[0].endPos).toBe(6);
  });

  test('loc { line, column } only — zero-width at absolute offset', () => {
    const r = reportAtLoc({ line: 1, column: 6 });
    expect(r.ruleErrors).toBeUndefined();
    expect(r.diagnostics).toHaveLength(1);
    expect(r.diagnostics[0].startPos).toBe(6);
    expect(r.diagnostics[0].endPos).toBe(6);
  });

  // M17 regression: plugins that return junk from fix(fixer) (not a
  // Fix object, not Fix[], not iterable) used to drop the fix
  // silently. After the fix the value is still dropped (there's
  // nothing meaningful to apply), but a stderr line surfaces so the
  // plugin author can spot the misuse instead of staring at a missing
  // autofix. We check on the diagnostic surface: still emitted, no
  // fixes attached, and the stderr line contains the actionable hint.
  test('M17: fix(fixer) returning a non-Fix value warns and drops', () => {
    const loaded: LoadedPlugins = {
      plugins: [],
      rules: new Map<string, unknown>([
        [
          'stub/bad-fix',
          {
            meta: { name: 'bad-fix' },
            create(ctx: RuleContext) {
              return {
                Identifier(node: { range?: [number, number] }) {
                  ctx.report({
                    node: node as never,
                    message: 'bad',
                    // Plugin author returned the wrong shape — should
                    // be a Fix object or Fix[].
                    fix: () => 42 as unknown as never,
                  });
                },
              };
            },
          },
        ],
      ]),
    };

    // The fix-channel warnings used to go through `process.stderr.write`
    // directly. R4 switched them to `console.error` so the worker-side
    // monkey-patch in lint-worker.ts can forward them to the host's
    // user-visible log channel (otherwise LSP users see them in VS
    // Code's hidden "Window" output, not the rslint output channel).
    // Test infra runs lintFile in-process (no worker), so console.error
    // is unpatched — we monitor it directly.
    const originalErr = console.error;
    let logged = '';
    console.error = (...args: unknown[]) => {
      logged +=
        args.map((a) => (typeof a === 'string' ? a : String(a))).join(' ') +
        '\n';
    };
    try {
      const result = lintFile(
        {
          filePath: 'x.ts',
          text: 'const a = 1;',
          rules: { 'stub/bad-fix': { options: [] } },
          collectFixes: true,
          suggestionsMode: 'off',
        },
        loaded,
      );
      // Diagnostic still surfaces, with NO fix attached.
      expect(result.diagnostics.length).toBeGreaterThan(0);
      const d = result.diagnostics[0];
      expect(d.fixes ?? []).toHaveLength(0);
      // Helpful log line that mentions the fix path.
      expect(logged).toMatch(/unsupported value|Fix dropped/);
    } finally {
      console.error = originalErr;
    }
  });

  // M20 regression: _drainDiagnostics now CLEARS the source after
  // returning a copy. Multiple drains return what arrived BETWEEN
  // drains, not the cumulative list. lintFile only drains once
  // per file in production, but a future API consumer that drains
  // mid-visit would have hit the stale-reference footgun previously.
  test('M20: _drainDiagnostics clears the source array', () => {
    interface Probe {
      drainOnce: unknown[];
      drainTwice: unknown[];
    }
    const probe: Probe = { drainOnce: [], drainTwice: [] };
    const loaded: LoadedPlugins = {
      plugins: [],
      rules: new Map<string, unknown>([
        [
          'stub/probe',
          {
            meta: { name: 'probe' },
            create(ctx: RuleContext) {
              return {
                Identifier(node: { range?: [number, number] }) {
                  ctx.report({ node: node as never, message: 'one' });
                  // First drain: should return ONE diagnostic.
                  probe.drainOnce = (
                    ctx as unknown as {
                      _drainDiagnostics: () => unknown[];
                    }
                  )._drainDiagnostics();
                  // Second drain WITHOUT new report calls between:
                  // must be empty since the first drain cleared.
                  probe.drainTwice = (
                    ctx as unknown as {
                      _drainDiagnostics: () => unknown[];
                    }
                  )._drainDiagnostics();
                },
              };
            },
          },
        ],
      ]),
    };

    lintFile(
      {
        filePath: 'x.ts',
        text: 'const a = 1;',
        rules: { 'stub/probe': { options: [] } },
        collectFixes: false,
        suggestionsMode: 'off',
      },
      loaded,
    );

    expect(probe.drainOnce.length).toBeGreaterThan(0);
    // Critical: after the first drain, the source is empty — second
    // drain must NOT return the same items again.
    expect(probe.drainTwice).toHaveLength(0);
  });
});
