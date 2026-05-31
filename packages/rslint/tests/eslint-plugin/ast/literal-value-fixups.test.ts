import { describe, it, expect } from '@rstest/core';
import { parse as nativeParse } from '@rslint/native';
import {
  normalizeAst,
  buildLineStartOffsets,
} from '../../../src/eslint-plugin/ast/normalize-ast.js';

// The native parser ships the ESTree as JSON, which cannot carry a RegExp object or a
// BigInt, so RegExp/BigInt literals arrive with `value: null` plus structured
// `regex`/`bigint` fields. `applyLiteralValueFixups` (in normalize-ast) rebuilds the
// value. These tests exercise that exact path (the existing normalize-ast tests feed
// npm oxc-parser objects whose value is already present, so they never trigger it).

interface Node {
  type?: string;
  value?: unknown;
  regex?: { pattern: string; flags: string };
  bigint?: string;
  [k: string]: unknown;
}

function parseAndNormalize(code: string): Node {
  const r = nativeParse('t.js', code, 'module', false);
  const ast = JSON.parse(r.program) as Node;
  normalizeAst(ast as never, buildLineStartOffsets(code), code);
  return ast;
}

function findAll(
  n: Node,
  pred: (x: Node) => boolean,
  out: Node[] = [],
): Node[] {
  if (n && typeof n === 'object') {
    if (pred(n)) out.push(n);
    for (const k of Object.keys(n)) {
      if (k === 'parent') continue;
      const v = (n as Record<string, unknown>)[k];
      if (Array.isArray(v)) v.forEach((c) => findAll(c as Node, pred, out));
      else if (v && typeof v === 'object') findAll(v as Node, pred, out);
    }
  }
  return out;
}

describe('applyLiteralValueFixups: rebuild RegExp/BigInt Literal.value (napi JSON transfer)', () => {
  it('rebuilds a RegExp value (instanceof RegExp, exact source + flags)', () => {
    const ast = parseAndNormalize('const r = /ab+c/gi;');
    const lit = findAll(ast, (n) => n.type === 'Literal' && n.regex != null)[0];
    expect(lit.value).toBeInstanceOf(RegExp);
    expect((lit.value as RegExp).source).toBe('ab+c');
    expect((lit.value as RegExp).flags).toBe('gi');
  });

  it('rebuilds a RegExp with a unicode flag', () => {
    const ast = parseAndNormalize('const r = /\\u{1f600}/u;');
    const lit = findAll(ast, (n) => n.type === 'Literal' && n.regex != null)[0];
    expect(lit.value).toBeInstanceOf(RegExp);
    expect((lit.value as RegExp).flags).toBe('u');
  });

  it('rebuilds a decimal BigInt value', () => {
    const ast = parseAndNormalize('const b = 123n;');
    const lit = findAll(
      ast,
      (n) => n.type === 'Literal' && n.bigint != null,
    )[0];
    expect(typeof lit.value).toBe('bigint');
    expect(lit.value).toBe(123n);
  });

  it('rebuilds a hex BigInt value', () => {
    const ast = parseAndNormalize('const b = 0x1Fn;');
    const lit = findAll(
      ast,
      (n) => n.type === 'Literal' && n.bigint != null,
    )[0];
    expect(lit.value).toBe(31n);
  });

  it('leaves a plain `null` literal untouched (no regex/bigint -> no rebuild)', () => {
    const ast = parseAndNormalize('const x = null;');
    const lit = findAll(
      ast,
      (n) =>
        n.type === 'Literal' &&
        n.regex == null &&
        n.bigint == null &&
        n.value === null,
    )[0];
    expect(lit).toBeDefined();
    expect(lit.value).toBeNull();
  });
});

describe('napi parse edge cases', () => {
  it('preserves a lone surrogate from a `\\uXXXX` escape (matches espree)', () => {
    // Escaped lone surrogate: source is ASCII `\uD800` (NOT a raw surrogate
    // byte, so no String-arg lossy →U+FFFD, unlike DESIGN D3's raw case).
    // oxc decodes to U+D800; the ESTree JSON carries it as `\ud800` and
    // JSON.parse recovers U+D800 — the same value espree produces. Pins the
    // one previously-unverified surrogate corner (review Finding 1).
    const ast = parseAndNormalize('const s = "\\uD800";');
    const lit = findAll(
      ast,
      (n) => n.type === 'Literal' && typeof n.value === 'string',
    )[0];
    expect(lit).toBeDefined();
    expect((lit.value as string).length).toBe(1);
    expect((lit.value as string).charCodeAt(0)).toBe(0xd800);
  });

  it('throws on input exceeding the size guard (MAX_SOURCE_BYTES = 16MB)', () => {
    // The size guard is the ONLY path (besides a caught panic) that yields a
    // parseError; the JS side maps this throw to `result.parseError`. Pins
    // that >16MB input is rejected rather than parsed.
    const tooBig = 'a'.repeat(16 * 1024 * 1024 + 1);
    expect(() => nativeParse('t.js', tooBig, 'module', false)).toThrow();
  });
});
