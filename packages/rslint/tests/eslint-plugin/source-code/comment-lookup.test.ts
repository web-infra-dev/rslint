/**
 * Differential coverage for the offset-indexed lookups that back
 * `getCommentsBefore` / `getCommentsAfter` / `getCommentsInside` and the
 * numeric-padding token slice.
 *
 * All four used to compute their bound with a full linear scan of the
 * file's comment / token array on EVERY call. A rule that asks for the
 * comments attached to each declaration (`local/jsdoc-format` in the
 * TypeScript repo, and every JSDoc-driven rule) therefore paid
 * O(declarations x comments) per file. They now binary-search to the
 * first candidate and walk forward only across the ones that can still
 * satisfy the bound.
 *
 * The tests below pin that rewrite to the ORIGINAL predicates: each
 * reference implementation here is the pre-change filter, verbatim, and
 * the assertions sweep every interesting offset in a comment-dense
 * fixture (comment/token boundaries and their +-1 neighbourhoods, plus
 * both ends of the file) so boundary-off-by-one is not representable.
 */

import { describe, test, expect } from '@rstest/core';

import { parse as nativeParse } from '../../../src/eslint-plugin/native/load-binding.js';

import { createSourceCode } from '../../../src/eslint-plugin/source-code/source-code.js';
import type { ESTreeNode } from '../../../src/eslint-plugin/source-code/source-code.js';
import {
  spanIndexEndingAtOrBefore,
  type Comment,
  type Token,
} from '../../../src/eslint-plugin/source-code/token-builder.js';
import { paddedTokenSlice } from '../../../src/eslint-plugin/source-code/source-code-helpers.js';

// Same shape as `token-padding.test.ts`'s helper: a real native parse so
// tokens and comments carry real offsets; only `range` is load-bearing on
// the probe nodes, since every getter under test resolves purely by offset.
function mkSC(text: string) {
  const ast = {
    type: 'Program',
    body: [],
    sourceType: 'module',
    range: [0, text.length] as [number, number],
    loc: { start: { line: 1, column: 0 }, end: { line: 1, column: 0 } },
  } as unknown as ESTreeNode;
  const parsed = nativeParse('test.ts', text, 'module', false);
  return createSourceCode({
    text,
    ast,
    scopeManagerFactory: () => ({}),
    parsedTokens: {
      types: parsed.tokenTypes,
      starts: parsed.tokenStarts,
      ends: parsed.tokenEnds,
    },
    parsedComments: parsed.comments as ReadonlyArray<{
      type: 'Line' | 'Block';
      value: string;
      start: number;
      end: number;
    }>,
  });
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

// A comment-dense fixture: leading comments, trailing comments on the
// same line, back-to-back comments with no code between them, a comment
// flush against a token (`a/*x*/b`), comments inside a block, and a
// comment as the very last thing in the file.
const TEXT = [
  '// leading line',
  '/** jsdoc for f */',
  '/* extra */',
  'function f(a /* flush */, b) {',
  '  // inside',
  '  return a /*g*/ + b; // trailing',
  '}',
  '/* between */',
  '/** jsdoc for g */',
  'function g() {}',
  '// last line of the file',
].join('\n');

const sc = mkSC(TEXT);
const PROGRAM = mkNode(0, TEXT.length);
const ALL_TOKENS = sc.getTokens(PROGRAM) as unknown as Token[];
const ALL_COMMENTS = sc.getAllComments() as unknown as Comment[];

// ────────────────────────────────────────────────────────────────────
// Reference implementations — the pre-change bodies, unchanged.
// ────────────────────────────────────────────────────────────────────

// `spanIndexStartingAtOrAfter` as a scan: first index whose start is at/after
// `offset`, or -1.
function refIndexAtOrAfter(
  stream: ReadonlyArray<{ range: [number, number] }>,
  offset: number,
): number {
  for (let i = 0; i < stream.length; i++) {
    if (stream[i].range[0] >= offset) return i;
  }
  return -1;
}

function refCommentsBefore(range: [number, number]): Comment[] {
  let prevTokenEnd = 0;
  const idxAfter = refIndexAtOrAfter(ALL_TOKENS, range[0]);
  const prevIdx = idxAfter < 0 ? ALL_TOKENS.length - 1 : idxAfter - 1;
  if (prevIdx >= 0) prevTokenEnd = ALL_TOKENS[prevIdx].range[1];
  return ALL_COMMENTS.filter(
    (c) => c.range[0] >= prevTokenEnd && c.range[1] <= range[0],
  );
}

function refCommentsAfter(range: [number, number]): Comment[] {
  const idx = refIndexAtOrAfter(ALL_TOKENS, range[1]);
  const nextTokenStart = idx < 0 ? TEXT.length : ALL_TOKENS[idx].range[0];
  return ALL_COMMENTS.filter(
    (c) => c.range[0] >= range[1] && c.range[1] <= nextTokenStart,
  );
}

function refCommentsInside(range: [number, number]): Comment[] {
  return ALL_COMMENTS.filter(
    (c) => c.range[0] >= range[0] && c.range[1] <= range[1],
  );
}

function refPaddedSlice(
  tokens: readonly Token[],
  startLoc: number,
  endLoc: number,
  beforeCount: number,
  afterCount: number,
): Token[] {
  const len = tokens.length;
  if (len === 0) return [];
  let firstIdx = len;
  for (let i = 0; i < len; i++) {
    if (tokens[i].range[0] >= startLoc) {
      firstIdx = i;
      break;
    }
  }
  let lastIdx = -1;
  for (let i = len - 1; i >= 0; i--) {
    if (tokens[i].range[1] <= endLoc) {
      lastIdx = i;
      break;
    }
  }
  const start = Math.max(0, firstIdx - beforeCount);
  const end = Math.min(len - 1, lastIdx + afterCount);
  if (start > end) return [];
  return tokens.slice(start, end + 1);
}

// ────────────────────────────────────────────────────────────────────
// Probe offsets: every token / comment boundary and its +-1
// neighbourhood, plus both ends of the file.
// ────────────────────────────────────────────────────────────────────

const OFFSETS: number[] = (() => {
  const set = new Set<number>([0, TEXT.length]);
  for (const t of [...ALL_TOKENS, ...ALL_COMMENTS]) {
    for (const base of t.range) {
      for (const d of [-1, 0, 1]) {
        const o = base + d;
        if (o >= 0 && o <= TEXT.length) set.add(o);
      }
    }
  }
  return [...set].sort((a, b) => a - b);
})();

const ranges = (): Array<[number, number]> => {
  const out: Array<[number, number]> = [];
  for (const a of OFFSETS) {
    for (const b of OFFSETS) {
      if (b >= a) out.push([a, b]);
    }
  }
  return out;
};

const RANGES = ranges();

describe('comment lookups match the pre-binary-search predicates', () => {
  test('the fixture actually exercises the scan (many comments)', () => {
    // Guards the whole suite: with a one-comment fixture a broken
    // binary search would still pass every assertion below.
    expect(ALL_COMMENTS.length).toBeGreaterThanOrEqual(9);
    expect(ALL_TOKENS.length).toBeGreaterThanOrEqual(20);
    expect(RANGES.length).toBeGreaterThan(1000);
  });

  test('getCommentsBefore over every boundary offset', () => {
    for (const r of RANGES) {
      expect(sc.getCommentsBefore(mkNode(r[0], r[1]))).toEqual(
        refCommentsBefore(r),
      );
    }
  });

  test('getCommentsAfter over every boundary offset', () => {
    for (const r of RANGES) {
      expect(sc.getCommentsAfter(mkNode(r[0], r[1]))).toEqual(
        refCommentsAfter(r),
      );
    }
  });

  test('getCommentsInside over every boundary offset', () => {
    for (const r of RANGES) {
      expect(sc.getCommentsInside(mkNode(r[0], r[1]))).toEqual(
        refCommentsInside(r),
      );
    }
  });

  test('known-good spot checks (ESLint semantics)', () => {
    // ESLint returns only the comments between the node and the nearest
    // neighbouring CODE token — not every earlier comment in the file.
    const fnKeyword = ALL_TOKENS.find((t) => t.value === 'function')!;
    const before = sc.getCommentsBefore(
      mkNode(fnKeyword.range[0], fnKeyword.range[1]),
    );
    expect(before.map((c) => c.value.trim())).toEqual([
      'leading line',
      '* jsdoc for f',
      'extra',
    ]);

    // `function f(a /* flush */, b)` — the comment sits in the gap
    // between `a` and `,`, so it attaches to the `,` and NOT to `b`
    // (whose nearest preceding code token is the `,` itself).
    const flushComment = ALL_COMMENTS.find((c) => c.value === ' flush ')!;
    const comma = ALL_TOKENS.find(
      (t) => t.value === ',' && t.range[0] >= flushComment.range[1],
    )!;
    expect(
      sc
        .getCommentsBefore(mkNode(comma.range[0], comma.range[1]))
        .map((c) => c.value),
    ).toEqual([' flush ']);
    const b = ALL_TOKENS.find(
      (t) => t.value === 'b' && t.range[0] > comma.range[0],
    )!;
    expect(sc.getCommentsBefore(mkNode(b.range[0], b.range[1]))).toEqual([]);

    // The trailing `// last line of the file` comment has no code token
    // after it — the bound falls back to EOF.
    const closeBrace = ALL_TOKENS[ALL_TOKENS.length - 1];
    expect(
      sc
        .getCommentsAfter(mkNode(closeBrace.range[0], closeBrace.range[1]))
        .map((c) => c.value.trim()),
    ).toEqual(['last line of the file']);
  });
});

describe('spanIndexEndingAtOrBefore', () => {
  test('matches a backward scan for every boundary offset', () => {
    for (const offset of OFFSETS) {
      let expected = -1;
      for (let i = ALL_TOKENS.length - 1; i >= 0; i--) {
        if (ALL_TOKENS[i].range[1] <= offset) {
          expected = i;
          break;
        }
      }
      expect(spanIndexEndingAtOrBefore(ALL_TOKENS, offset)).toBe(expected);
    }
  });

  test('-1 before the first token end, last index past EOF', () => {
    expect(spanIndexEndingAtOrBefore(ALL_TOKENS, 0)).toBe(-1);
    expect(spanIndexEndingAtOrBefore(ALL_TOKENS, TEXT.length)).toBe(
      ALL_TOKENS.length - 1,
    );
    expect(spanIndexEndingAtOrBefore([], 42)).toBe(-1);
  });
});

describe('paddedTokenSlice matches the pre-binary-search scans', () => {
  test('over every boundary offset pair and several padding widths', () => {
    for (const r of RANGES) {
      for (const [before, after] of [
        [0, 0],
        [1, 1],
        [2, 0],
        [0, 3],
        [100, 100],
      ]) {
        expect(paddedTokenSlice(ALL_TOKENS, r[0], r[1], before, after)).toEqual(
          refPaddedSlice(ALL_TOKENS, r[0], r[1], before, after),
        );
      }
    }
  });
});

describe('isSpaceBetween is unaffected by the walk bound', () => {
  test('comments do not count as space; real whitespace does', () => {
    const aTok = ALL_TOKENS.find((t) => t.value === 'a')!;
    const flushComment = ALL_COMMENTS.find((c) => c.value === ' flush ')!;
    // The `,` sits flush against the comment's end.
    const comma = ALL_TOKENS.find(
      (t) => t.value === ',' && t.range[0] >= flushComment.range[1],
    )!;

    // `a /* flush */,` — whitespace between `a` and the comment.
    expect(
      sc.isSpaceBetween(
        aTok as unknown as ESTreeNode,
        comma as unknown as ESTreeNode,
      ),
    ).toBe(true);

    // `a/*x*/(b)` — a comment flush against both sides is NOT space.
    const TIGHT = 'a/*x*/(b)';
    const tight = mkSC(TIGHT);
    const tightTokens = tight.getTokens(
      mkNode(0, TIGHT.length),
    ) as unknown as Token[];
    expect(
      tight.isSpaceBetween(
        tightTokens[0] as unknown as ESTreeNode,
        tightTokens[1] as unknown as ESTreeNode,
      ),
    ).toBe(false);

    // Same shape, with a space on one side → space.
    const LOOSE = 'a /*x*/(b)';
    const loose = mkSC(LOOSE);
    const looseTokens = loose.getTokens(
      mkNode(0, LOOSE.length),
    ) as unknown as Token[];
    expect(
      loose.isSpaceBetween(
        looseTokens[0] as unknown as ESTreeNode,
        looseTokens[1] as unknown as ESTreeNode,
      ),
    ).toBe(true);
  });
});
