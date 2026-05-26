/**
 * ESLint-v10-parity coverage for three SourceCode fixes:
 *
 *   #6 / #7  getTokens(node, before, after) and getTokensBetween(left,
 *            right, padding) accept NUMERIC padding args. ESLint routes
 *            these through a `PaddedTokenCursor`
 *            (`lib/languages/js/source-code/token-store/padded-token-cursor.js`):
 *            the slice is inflated by that many CODE tokens on each
 *            side, clamped to the token array. Pre-fix `normalizeFilterOpts`
 *            had no numeric branch (unlike the sibling skip/count
 *            normalizers), so a numeric padding arg was silently dropped
 *            and the padded call returned the unpadded slice.
 *
 *   #8       getInlineConfigNodes() gates Line comments. ESLint
 *            (`lib/languages/js/source-code/source-code.js`,
 *            getInlineConfigNodes) ends its filter with
 *            `comment.type !== "Line" || /^eslint-disable-(next-)?line$/u
 *            .test(label)` — a Line comment qualifies ONLY when its
 *            directive is `eslint-disable-line` / `eslint-disable-next-line`.
 *            Pre-fix rslint had no Line gate here, so `// eslint-disable foo`
 *            (Line, block-only directive) was returned by
 *            getInlineConfigNodes() while getDisableDirectives() (which
 *            re-applies the gate) dropped it — the two surfaces disagreed.
 *
 * Counts / values below were cross-checked byte-for-byte against
 * eslint@9.32.0's actual `TokenStore` and `SourceCode` for the identical
 * input (the APIs are unchanged through v10.4.0, the review's oracle).
 */

import { describe, test, expect } from '@rstest/core';

import { createSourceCode } from '../../src/source-code/source-code.js';
import type { ESTreeNode } from '../../src/source-code/source-code.js';

// Build a SourceCode over real text so the lexer produces real tokens /
// comments — mirrors `regression-coverage.test.ts`'s `mkSC`. Only a
// `range` is load-bearing on the probe nodes; the token getters resolve
// purely by offset.
function mkSC(text: string) {
  const ast = {
    type: 'Program',
    body: [],
    sourceType: 'module',
    range: [0, text.length] as [number, number],
    loc: {
      start: { line: 1, column: 0 },
      end: { line: 1, column: 0 },
    },
  } as unknown as ESTreeNode;
  return createSourceCode({ text, ast, scopeManagerFactory: () => ({}) });
}

function mkNode(start: number, end: number): ESTreeNode {
  return {
    type: 'Probe',
    range: [start, end],
    loc: {
      start: { line: 1, column: start },
      end: { line: 1, column: end },
    },
    start,
    end,
  } as unknown as ESTreeNode;
}

const vals = (tokens: ReadonlyArray<{ value: string }>): string[] =>
  tokens.map((t) => t.value);

// ────────────────────────────────────────────────────────────────────
// #6 / #7 — numeric token padding
// ────────────────────────────────────────────────────────────────────

describe('getTokens numeric padding (PaddedTokenCursor parity) [#6]', () => {
  // `const x = 1; let y = 2;` tokenizes to exactly 10 code tokens:
  //   [0] 0-5   const   [1] 6-7  x   [2] 8-9  =   [3] 10-11 1   [4] 11-12 ;
  //   [5] 13-16 let     [6] 17-18 y  [7] 19-20 =  [8] 21-22 2   [9] 22-23 ;
  // The `=` of the FIRST statement is token [2], with room on BOTH sides
  // so before/after padding is observable independently.
  const TEXT = 'const x = 1; let y = 2;';
  const EQ = mkNode(8, 9); // the first `=`

  test('no padding → just the in-range token', () => {
    const sc = mkSC(TEXT);
    expect(vals(sc.getTokens(EQ))).toEqual(['=']);
  });

  test('getTokens(node, 1, 1) adds exactly one token before AND one after', () => {
    const sc = mkSC(TEXT);
    const unpadded = sc.getTokens(EQ);
    const padded = sc.getTokens(EQ, 1, 1);
    expect(vals(unpadded)).toEqual(['=']);
    // 1 before (`x`) + the token + 1 after (`1`).
    expect(vals(padded)).toEqual(['x', '=', '1']);
    // The padded slice is strictly larger by 2 (one each side).
    expect(padded.length).toBe(unpadded.length + 2);
  });

  test('before-only and after-only padding are independent', () => {
    const sc = mkSC(TEXT);
    // beforeCount=2, afterCount=0 → two tokens before, none after.
    expect(vals(sc.getTokens(EQ, 2, 0))).toEqual(['const', 'x', '=']);
    // beforeCount=0, afterCount=2 → none before, two after.
    expect(vals(sc.getTokens(EQ, 0, 2))).toEqual(['=', '1', ';']);
  });

  test('omitted afterCount defaults to 0 (matches ESLint afterCount | 0)', () => {
    const sc = mkSC(TEXT);
    // getTokens(node, 2) — afterCount omitted → only before-padding.
    expect(vals(sc.getTokens(EQ, 2))).toEqual(['const', 'x', '=']);
  });

  test('padding clamps to the token array bounds (no overflow / underflow)', () => {
    const sc = mkSC(TEXT);
    // Huge counts saturate at the full 10-token array, never throw.
    expect(vals(sc.getTokens(EQ, 100, 100))).toEqual([
      'const',
      'x',
      '=',
      '1',
      ';',
      'let',
      'y',
      '=',
      '2',
      ';',
    ]);
  });
});

describe('getTokensBetween numeric padding (PaddedTokenCursor parity) [#7]', () => {
  const TEXT = 'const x = 1; let y = 2;';
  // left = `const` (range 0-5), right = the FIRST `;` (range 11-12).
  // Between them sit exactly x, =, 1.
  const LEFT = mkNode(0, 5);
  const RIGHT = mkNode(11, 12);

  test('no padding → just the tokens strictly between', () => {
    const sc = mkSC(TEXT);
    expect(vals(sc.getTokensBetween(LEFT, RIGHT))).toEqual(['x', '=', '1']);
  });

  test('getTokensBetween(left, right, 1) adds one token on EACH side', () => {
    const sc = mkSC(TEXT);
    const unpadded = sc.getTokensBetween(LEFT, RIGHT);
    const padded = sc.getTokensBetween(LEFT, RIGHT, 1);
    expect(vals(unpadded)).toEqual(['x', '=', '1']);
    // padding=1 → before-token `const` + between + after-token `;`.
    expect(vals(padded)).toEqual(['const', 'x', '=', '1', ';']);
    expect(padded.length).toBe(unpadded.length + 2);
  });

  test('padding=2 reaches two tokens on each side (clamped at the low end)', () => {
    const sc = mkSC(TEXT);
    // Low end clamps at index 0 (`const`); high end reaches `let`.
    expect(vals(sc.getTokensBetween(LEFT, RIGHT, 2))).toEqual([
      'const',
      'x',
      '=',
      '1',
      ';',
      'let',
    ]);
  });
});

// ────────────────────────────────────────────────────────────────────
// #8 — getInlineConfigNodes Line gate + getDisableDirectives consistency
// ────────────────────────────────────────────────────────────────────

describe('getInlineConfigNodes Line-comment gate + getDisableDirectives consistency [#8]', () => {
  // Three directive comments:
  //   L1  `// eslint-disable foo`            Line, BLOCK-only directive → EXCLUDED
  //   L2  `// eslint-disable-next-line bar`  Line, line directive       → INCLUDED
  //   L3  `/* eslint-disable baz */`         Block (any label)          → INCLUDED
  const TEXT = [
    '// eslint-disable foo',
    '// eslint-disable-next-line bar',
    '/* eslint-disable baz */',
    'const a = 1;',
  ].join('\n');

  test('getInlineConfigNodes excludes the Line non-line-directive, keeps the other two', () => {
    const sc = mkSC(TEXT);
    const nodes = sc.getInlineConfigNodes();
    // EXACTLY two qualifying nodes — the bad Line comment is gone.
    expect(nodes.length).toBe(2);
    expect(nodes.map((c) => ({ type: c.type, value: c.value }))).toEqual([
      { type: 'Line', value: ' eslint-disable-next-line bar' },
      { type: 'Block', value: ' eslint-disable baz ' },
    ]);
    // The over-reported node must NOT be present.
    expect(nodes.some((c) => c.value === ' eslint-disable foo')).toBe(false);
  });

  test('getDisableDirectives agrees with getInlineConfigNodes', () => {
    const sc = mkSC(TEXT);
    const dd = sc.getDisableDirectives();
    expect(dd.problems).toEqual([]);
    // disable-next-line(bar) + disable(baz) — and crucially NOT a third
    // directive for `// eslint-disable foo`.
    expect(
      dd.directives.map((d) => ({ type: d.type, value: d.value })),
    ).toEqual([
      { type: 'disable-next-line', value: 'bar' },
      { type: 'disable', value: 'baz' },
    ]);

    // The two surfaces describe the SAME set of comment nodes: every
    // directive's carrier comment is among getInlineConfigNodes(), and
    // the counts match (no over- or under-reporting on either side).
    const inlineNodes = sc.getInlineConfigNodes();
    for (const d of dd.directives) {
      expect(inlineNodes).toContain(d.node);
    }
    expect(dd.directives.length).toBe(inlineNodes.length);
  });
});
