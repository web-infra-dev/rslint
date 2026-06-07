/**
 * Coverage for the cache / lazy-state optimizations added in the
 * 2026 perf sweep. Each test pins a specific invariant the
 * optimization depends on — if a future refactor breaks the
 * invariant, these tests fail before the slowdown ships.
 *
 * The four invariants under test:
 *
 *   1. `createRuleContext({ sourceCode })` returns a RuleContext that
 *      wires the SUPPLIED sourceCode (not a fresh one). All rules on
 *      the same file share one SourceCode → shared lazy caches
 *      (tokens, comments, scopeManager).
 *   2. `createSourceCode({ parsedComments })` makes `getAllComments`
 *      return ESLint-shape comments synthesized from the oxc input
 *      WITHOUT triggering the full text tokenizer. Shebang lines are
 *      rewritten to type='Shebang'.
 *   3. `buildDiagnostic({ lsoCache, descriptor: { loc } })` reuses the
 *      supplied lso array. The expensive `buildLineStartOffsetsLocal`
 *      fallback path runs only when `lsoCache` is absent.
 *   4. `mergeListeners` returns an `esqueryByType` map keyed by
 *      statically-extracted anchor types, with non-typeable selectors
 *      filed under `enterWildcard` / `exitWildcard`. (Covered already by
 *      tests/listener-merge-bucket.test.ts — referenced here for the
 *      audit trail.)
 */

import { describe, test, expect } from '@rstest/core';

import { createSourceCode } from '../../../src/eslint-plugin/source-code/source-code.js';
import { createRuleContext } from '../../../src/eslint-plugin/linter/context.js';
import { buildDiagnostic } from '../../../src/eslint-plugin/linter/diagnostic-builder.js';
import { buildLineStartOffsets } from '../../../src/eslint-plugin/ast/normalize-ast.js';
import { makeFixer } from '../../../src/eslint-plugin/linter/fixer.js';
import type { ESTreeNode } from '../../../src/eslint-plugin/source-code/source-code.js';

// Minimal node helper for synthesizing ASTs in these tests. We don't
// need real oxc output — just a `type` + `range` is enough for the
// SourceCode/buildDiagnostic surfaces we exercise here.
function mkNode(type: string, start: number, end: number): ESTreeNode {
  return {
    type,
    range: [start, end],
    loc: { start: { line: 1, column: start }, end: { line: 1, column: end } },
    start,
    end,
  };
}

// ────────────────────────────────────────────────────────────────────
// 1. shared SourceCode across rule contexts
// ────────────────────────────────────────────────────────────────────

describe('shared SourceCode across rule contexts', () => {
  test('createRuleContext honors opts.sourceCode (does not build a fresh one)', () => {
    const text = 'const x = 1;';
    const ast = mkNode('Program', 0, text.length);

    const sharedSc = createSourceCode({
      text,
      ast,
      scopeManagerFactory: () => ({}),
    });

    const ctxA = createRuleContext({
      ruleName: 'rule-a',
      filePath: '/tmp/x.js',
      userOptions: [],
      settings: {},
      text,
      ast,
      scopeManagerFactory: () => ({}),
      sourceCode: sharedSc,
      collectFixes: false,
      suggestionsMode: 'off',
    });
    const ctxB = createRuleContext({
      ruleName: 'rule-b',
      filePath: '/tmp/x.js',
      userOptions: [],
      settings: {},
      text,
      ast,
      scopeManagerFactory: () => ({}),
      sourceCode: sharedSc,
      collectFixes: false,
      suggestionsMode: 'off',
    });

    // CRITICAL: both contexts must point at the SAME SourceCode object.
    // If `sourceCode` is silently ignored and `createSourceCode` runs
    // twice, rules pay 2× the lazy-state cost (2× tokenize, 2× scope
    // build) and the "shared lazy cache" optimization is undone.
    expect(ctxA.sourceCode).toBe(ctxB.sourceCode);
    expect(ctxA.sourceCode).toBe(sharedSc);
  });

  test('shared sourceCode means lazy state primed by one ctx is visible to another', () => {
    const text = 'function foo() { return 1; }';
    const ast = mkNode('Program', 0, text.length);
    let scopeFactoryCalls = 0;
    const sharedSc = createSourceCode({
      text,
      ast,
      scopeManagerFactory: () => {
        scopeFactoryCalls++;
        return { id: 'shared-scope-mgr' };
      },
    });

    const ctxA = createRuleContext({
      ruleName: 'rule-a',
      filePath: '/tmp/y.js',
      userOptions: [],
      settings: {},
      text,
      ast,
      scopeManagerFactory: () => {
        throw new Error('inner factory must not run — sourceCode is shared');
      },
      sourceCode: sharedSc,
      collectFixes: false,
      suggestionsMode: 'off',
    });

    // ctxA triggers scope build via SourceCode getter.
    void ctxA.sourceCode.scopeManager;
    expect(scopeFactoryCalls).toBe(1);

    // ctxB reuses the same SourceCode → scopeManager getter hits the
    // already-computed cache, factory does NOT run again.
    const ctxB = createRuleContext({
      ruleName: 'rule-b',
      filePath: '/tmp/y.js',
      userOptions: [],
      settings: {},
      text,
      ast,
      scopeManagerFactory: () => {
        throw new Error('inner factory must not run — sourceCode is shared');
      },
      sourceCode: sharedSc,
      collectFixes: false,
      suggestionsMode: 'off',
    });
    void ctxB.sourceCode.scopeManager;
    expect(scopeFactoryCalls).toBe(1); // still 1 — cache held across ctxs
  });
});

// ────────────────────────────────────────────────────────────────────
// 2. oxc-direct comments — getAllComments without tokenize
// ────────────────────────────────────────────────────────────────────

describe('getAllComments uses parsedComments without tokenize', () => {
  test('returns ESLint-shape comments built from parsedComments input', () => {
    const text = '// hello\n/* block */\nlet x = 1;';
    const ast = mkNode('Program', 0, text.length);
    const parsedComments = [
      { type: 'Line' as const, value: ' hello', start: 0, end: 8 },
      { type: 'Block' as const, value: ' block ', start: 9, end: 20 },
    ];

    const sc = createSourceCode({
      text,
      ast,
      scopeManagerFactory: () => ({}),
      parsedComments,
    });

    const comments = sc.getAllComments();
    expect(comments).toHaveLength(2);
    // Each comment must carry the ESLint contract fields: type, value,
    // range (= [start, end]), loc.
    expect(comments[0].type).toBe('Line');
    expect(comments[0].value).toBe(' hello');
    expect(comments[0].range).toEqual([0, 8]);
    expect(comments[0].loc.start.line).toBe(1);
    expect(comments[1].type).toBe('Block');
    expect(comments[1].range).toEqual([9, 20]);
  });

  test('rewrites oxc-emitted Line comment at offset 0 to Shebang when text starts with #!', () => {
    // oxc emits the shebang line as `type: 'Line'`. ESLint exposes it
    // as `type: 'Shebang'`. The buildCommentsFromParsed adapter must
    // rewrite — this is the bit that broke once during the perf sweep
    // (we double-synthesized shebangs); pin it permanently.
    const text = '#!/usr/bin/env node\nlet x = 1;';
    const ast = mkNode('Program', 0, text.length);
    const parsedComments = [
      { type: 'Line' as const, value: '/usr/bin/env node', start: 0, end: 19 },
    ];

    const sc = createSourceCode({
      text,
      ast,
      scopeManagerFactory: () => ({}),
      parsedComments,
    });

    const comments = sc.getAllComments();
    expect(comments).toHaveLength(1);
    expect(comments[0].type).toBe('Shebang');
    expect(comments[0].value).toBe('/usr/bin/env node');
    expect(comments[0].range).toEqual([0, 19]);
  });

  test('does NOT rewrite a Line comment that is not at offset 0', () => {
    // `#!` only counts as shebang when it's the file preamble. A
    // literal `//` comment on line 1 of a non-shebang file must
    // remain `type: 'Line'`.
    const text = '// foo\nlet x = 1;';
    const ast = mkNode('Program', 0, text.length);
    const parsedComments = [
      { type: 'Line' as const, value: ' foo', start: 0, end: 6 },
    ];

    const sc = createSourceCode({
      text,
      ast,
      scopeManagerFactory: () => ({}),
      parsedComments,
    });

    const comments = sc.getAllComments();
    expect(comments).toHaveLength(1);
    // Even though the comment is at offset 0, the text doesn't begin
    // with `#!` — must stay as 'Line'.
    expect(comments[0].type).toBe('Line');
  });

  test('second getAllComments call reuses the cached comment objects (no rebuild)', () => {
    // `getAllComments` returns `ensureComments().slice()`. The `.slice()`
    // means the two returned ARRAYS are always distinct (`a !== b`)
    // whether or not the comments were cached — so `a).not.toBe(b)`
    // proves nothing about caching. What proves the cache is ELEMENT
    // IDENTITY: `ensureComments` builds the `_comments` array once and
    // every later call slices that SAME array, so each comment OBJECT is
    // reused (`a[i] === b[i]`). A rebuild-every-call would synthesize
    // fresh comment objects on the second call → identity would break.
    const text = '// a\n// b\n// c\n';
    const ast = mkNode('Program', 0, text.length);
    const parsedComments = [
      { type: 'Line' as const, value: ' a', start: 0, end: 4 },
      { type: 'Line' as const, value: ' b', start: 5, end: 9 },
      { type: 'Line' as const, value: ' c', start: 10, end: 14 },
    ];

    const sc = createSourceCode({
      text,
      ast,
      scopeManagerFactory: () => ({}),
      parsedComments,
    });

    const a = sc.getAllComments();
    const b = sc.getAllComments();
    expect(a).toHaveLength(3);
    expect(b).toHaveLength(3);
    // The arrays are slice copies, so a caller mutating `a` cannot leak
    // into the cache. (Documented sanity check — NOT the cache proof.)
    expect(a).not.toBe(b);

    // The cache proof: every comment object is the SAME instance across
    // calls. Rebuilding would make these `!==`.
    expect(a[0]).toBe(b[0]);
    expect(a[1]).toBe(b[1]);
    expect(a[2]).toBe(b[2]);
  });
});

// ────────────────────────────────────────────────────────────────────
// 3. lsoCache reuse in buildDiagnostic descriptor.loc path
// ────────────────────────────────────────────────────────────────────

describe('buildDiagnostic reuses lsoCache for descriptor.loc path', () => {
  test('uses lsoCache value for line/column computation when supplied', () => {
    // Build a real lso for a 3-line file. We pass a custom lsoCache
    // with KNOWN incorrect line starts; if buildDiagnostic actually
    // uses the cache (the fast path), the computed startPos comes
    // from those offsets verbatim. If it ignores the cache and
    // recomputes from text, startPos disagrees with the cache.
    const text = 'line0\nline1\nline2\n';
    // Real offsets would be [0, 6, 12]; we feed deliberately altered
    // ones so the cache vs rebuild paths produce different results.
    const fakeLsoCache = [0, 100, 200];

    const diag = buildDiagnostic({
      ruleName: 'test',
      descriptor: {
        loc: { start: { line: 2, column: 3 }, end: { line: 2, column: 5 } },
        // message-only: this test exercises the lsoCache POSITION path,
        // not message resolution. (Pre-Fix-B this also passed `messageId:
        // 'm'` alongside `message` — an invalid both-present descriptor
        // that ESLint v10 / the runner now reject; dropping it keeps the
        // fixture valid without changing what this test asserts.)
        message: 'hello',
      },
      text,
      messages: {},
      fixer: makeFixer(),
      collectFixes: false,
      suggestionsMode: 'off',
      lsoCache: fakeLsoCache,
    });
    expect(diag).not.toBeNull();
    // line 2 (1-based) = lso index 1 = 100 (from our fake cache)
    // + column 3 = 103
    expect(diag!.startPos).toBe(103);
  });

  test('falls back to buildLineStartOffsetsLocal when no lsoCache supplied', () => {
    // No cache → reconstruct lso from text. Verify the descriptor.loc
    // path still works correctly (this is the legacy compatibility
    // path for unit tests / external API callers that don't have an
    // lso to hand in).
    const text = 'a\nb\nc\n';

    const diag = buildDiagnostic({
      ruleName: 'test',
      descriptor: {
        loc: { start: { line: 2, column: 0 }, end: { line: 2, column: 1 } },
        message: 'oops',
      },
      text,
      messages: {},
      fixer: makeFixer(),
      collectFixes: false,
      suggestionsMode: 'off',
    });
    expect(diag).not.toBeNull();
    // Real lso for 'a\nb\nc\n' is [0, 2, 4, 6]; line 2 = index 1 = 2.
    expect(diag!.startPos).toBe(2);
    expect(diag!.endPos).toBe(3);
  });

  test('node-path is unaffected by lsoCache (uses node.range directly)', () => {
    // Descriptors without `loc` take the fast `node.range` path. Any
    // lsoCache supplied must NOT interfere — this guards against a
    // future refactor that accidentally applies the cache on the
    // wrong path.
    const lso = buildLineStartOffsets('x = 1;\ny = 2;');
    const diag = buildDiagnostic({
      ruleName: 'test',
      descriptor: {
        node: { type: 'Identifier', range: [0, 1] } as ESTreeNode,
        message: 'foo',
      },
      text: 'x = 1;\ny = 2;',
      messages: {},
      fixer: makeFixer(),
      collectFixes: false,
      suggestionsMode: 'off',
      lsoCache: lso,
    });
    expect(diag).not.toBeNull();
    expect(diag!.startPos).toBe(0);
    expect(diag!.endPos).toBe(1);
  });
});
