import { describe, it, expect } from 'rstack/test';
import { projectSyntax } from './project-frame.js';
import { buildLineStartOffsets } from '../../../src/languages/js/positions.js';

// Frames carry regex/bigint descriptors; JavaScript materializes their values.

interface Node {
  type?: string;
  value?: unknown;
  regex?: { pattern: string; flags: string };
  bigint?: string;
  [k: string]: unknown;
}

function projectAST(code: string): Node {
  const r = projectSyntax('t.js', code, 'module');
  const ast = r.ast as Node;

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

describe('frame decoder: materialize RegExp/BigInt Literal.value', () => {
  it('rebuilds a RegExp value (instanceof RegExp, exact source + flags)', () => {
    const ast = projectAST('const r = /ab+c/gi;');
    const lit = findAll(ast, (n) => n.type === 'Literal' && n.regex != null)[0];
    expect(lit.value).toBeInstanceOf(RegExp);
    expect((lit.value as RegExp).source).toBe('ab+c');
    expect((lit.value as RegExp).flags).toBe('gi');
  });

  it('rebuilds a RegExp with a unicode flag', () => {
    const ast = projectAST('const r = /\\u{1f600}/u;');
    const lit = findAll(ast, (n) => n.type === 'Literal' && n.regex != null)[0];
    expect(lit.value).toBeInstanceOf(RegExp);
    expect((lit.value as RegExp).flags).toBe('u');
  });

  it('rebuilds a decimal BigInt value', () => {
    const ast = projectAST('const b = 123n;');
    const lit = findAll(
      ast,
      (n) => n.type === 'Literal' && n.bigint != null,
    )[0];
    expect(typeof lit.value).toBe('bigint');
    expect(lit.value).toBe(123n);
  });

  it('rebuilds a hex BigInt value', () => {
    const ast = projectAST('const b = 0x1Fn;');
    const lit = findAll(
      ast,
      (n) => n.type === 'Literal' && n.bigint != null,
    )[0];
    expect(lit.value).toBe(31n);
  });

  it('leaves a plain `null` literal untouched (no regex/bigint -> no rebuild)', () => {
    const ast = projectAST('const x = null;');
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

describe('projected literal edge cases', () => {
  it('preserves a lone surrogate from a `\\uXXXX` escape (matches espree)', () => {
    // Go stores the decoded escape as UTF-16 units in the binary frame.
    // JavaScript materialization must preserve U+D800, as espree does.
    const ast = projectAST('const s = "\\uD800";');
    const lit = findAll(
      ast,
      (n) => n.type === 'Literal' && typeof n.value === 'string',
    )[0];
    expect(lit).toBeDefined();
    expect((lit.value as string).length).toBe(1);
    expect((lit.value as string).charCodeAt(0)).toBe(0xd800);
  });

  it('accepts a large comment within the binary frame budget', () => {
    const text = '// ' + 'a'.repeat(16 * 1024 * 1024 + 1);
    const frame = projectSyntax('t.js', text, 'module');
    expect(frame.text).toBe(text);
    expect(frame.comments[0].end).toBe(text.length);
  });
});
