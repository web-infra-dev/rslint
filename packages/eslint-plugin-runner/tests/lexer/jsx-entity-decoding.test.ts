// cspell:ignore eacute
/**
 * JSX character-reference decoding — differential tests against espree
 * (ESLint v10's default parser, which lexes JSX via acorn-jsx).
 *
 * oxc-parser leaves JSX `.value` as verbatim source; the runner decodes
 * it in `normalize-ast` (AST nodes) and in the JSX tokenizer (tokens) so
 * output is byte-identical to ESLint v10. espree is the oracle here:
 * every case is run through the runner AND through espree, and the
 * resulting `.value`s must agree. A handful of headline cases also pin
 * the exact decoded string so the intended behavior is self-documenting.
 *
 * Surfaces covered:
 *   - AST  `JSXText.value`               (decode + CRLF→LF fold)
 *   - AST  `JSXAttribute` string value   (decode only, CRLF preserved)
 *   - token `JSXText` between tags        (decode + CRLF fold)
 *   - token JSX attribute string          (RAW — espree keeps it raw)
 *   - negative: ordinary JS string literals are NOT entity-decoded
 *   - guard: decoding is applied exactly once (`&amp;amp;` ≠ `&`)
 */
import { describe, test, expect } from '@rstest/core';
import * as espree from 'espree';
import { parseSync } from 'oxc-parser';

import {
  normalizeAst,
  buildLineStartOffsets,
} from '../../src/ast/normalize-ast.js';
import { tokenize } from '../../src/lexer/tokenizer.js';
import {
  decodeJsxText,
  decodeJsxAttributeValue,
} from '../../src/lexer/jsx/decode-entities.js';

// ── AST helpers ──────────────────────────────────────────────────────

interface AnyObj {
  type?: string;
  value?: unknown;
  [k: string]: unknown;
}

/** Collect, in document order, every JSXText value and every string-valued
 *  JSXAttribute value from a parsed program. */
function collectJsxValues(root: unknown): {
  text: string[];
  attr: string[];
} {
  const text: string[] = [];
  const attr: string[] = [];
  (function walk(n: unknown): void {
    if (n == null || typeof n !== 'object') return;
    const node = n as AnyObj;
    if (node.type === 'JSXText' && typeof node.value === 'string') {
      text.push(node.value);
    }
    if (node.type === 'JSXAttribute') {
      const v = node.value as AnyObj | null;
      if (v && v.type === 'Literal' && typeof v.value === 'string') {
        attr.push(v.value);
      }
    }
    for (const k of Object.keys(node)) {
      if (k === 'parent') continue;
      walk(node[k]);
    }
  })(root);
  return { text, attr };
}

/** Collect every string-valued JS `Literal` that is NOT a JSX attribute
 *  value (i.e. ordinary JS strings, including those inside JSX `{…}`). */
function collectJsStringLiterals(root: unknown): string[] {
  const out: string[] = [];
  (function walk(n: unknown, parentType: string | undefined): void {
    if (n == null || typeof n !== 'object') return;
    const node = n as AnyObj;
    if (
      node.type === 'Literal' &&
      typeof node.value === 'string' &&
      parentType !== 'JSXAttribute'
    ) {
      out.push(node.value);
    }
    for (const k of Object.keys(node)) {
      if (k === 'parent') continue;
      walk(node[k], node.type);
    }
  })(root, undefined);
  return out;
}

function runnerAst(src: string): unknown {
  const parsed = parseSync('test.jsx', src, { sourceType: 'module' });
  const ast =
    typeof parsed.program === 'string'
      ? JSON.parse(parsed.program as string)
      : parsed.program;
  normalizeAst(ast as never, buildLineStartOffsets(src), src);
  return ast;
}

function espreeAst(src: string): unknown {
  return espree.parse(src, {
    ecmaVersion: 'latest',
    sourceType: 'module',
    ecmaFeatures: { jsx: true },
  });
}

// ── token helpers ────────────────────────────────────────────────────

interface SimpleToken {
  type: string;
  value: string;
}

function runnerJsxTokens(src: string): SimpleToken[] {
  const { tokens } = tokenize(src, buildLineStartOffsets(src), { jsx: true });
  return tokens.map((t) => ({ type: t.type, value: t.value }));
}

function espreeJsxTokens(src: string): SimpleToken[] {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const tokens = (espree as any).tokenize(src, {
    ecmaVersion: 'latest',
    sourceType: 'module',
    ecmaFeatures: { jsx: true },
  }) as Array<{ type: string; value: string }>;
  return tokens.map((t) => ({ type: t.type, value: t.value }));
}

// ─────────────────────────────────────────────────────────────────────
// AST: JSXText value — decode + CRLF fold (oracle: espree)
// ─────────────────────────────────────────────────────────────────────

describe('AST JSXText.value matches espree (decode + CRLF fold)', () => {
  for (const src of [
    '<a>caf&eacute; &amp; x</a>',
    '<a>&#88; &#x59; &#x1d;</a>', // decimal + hex
    '<a>a & b &amp; c</a>', // bare ampersand kept
    '<a>&foo; &bar;</a>', // unknown names → literal
    '<a>&amp x &lt</a>', // incomplete (no semicolon)
    '<a>&; &#; &#x;</a>', // empty / malformed numeric
    '<a>&#X59;</a>', // capital X is NOT a hex marker in acorn-jsx
    '<a>&nbsp;&copy;&trade;&hearts;</a>',
    '<a>&int;&le;&ge;</a>', // names incl. the reserved-word key `int`
    '<a>x\r\ny</a>', // CRLF folds to LF
    '<a>x\ry\nz</a>', // lone CR kept, lone LF kept
    '<a>&#13;\n</a>', // entity-CR not re-folded against following LF
    '<a>&#x1F600;</a>', // astral ref → fromCharCode truncation quirk
    '<a>&amp;amp;</a>', // decode exactly once (guard)
    '<a>plain text no entities</a>',
  ]) {
    test(JSON.stringify(src), () => {
      const got = collectJsxValues(runnerAst(src)).text;
      const want = collectJsxValues(espreeAst(src)).text;
      expect(got).toEqual(want);
    });
  }

  test('headline values pinned', () => {
    expect(
      collectJsxValues(runnerAst('<a>caf&eacute; &amp; x</a>')).text,
    ).toEqual(['café & x']);
    expect(collectJsxValues(runnerAst('<a>&#88;&#x59;</a>')).text).toEqual([
      'XY',
    ]);
    expect(collectJsxValues(runnerAst('<a>x\r\ny</a>')).text).toEqual(['x\ny']);
    // decoded exactly once
    expect(collectJsxValues(runnerAst('<a>&amp;amp;</a>')).text).toEqual([
      '&amp;',
    ]);
  });
});

// ─────────────────────────────────────────────────────────────────────
// AST: JSXAttribute string value — decode only, CRLF preserved
// ─────────────────────────────────────────────────────────────────────

describe('AST JSXAttribute value matches espree (decode, no CRLF fold)', () => {
  for (const src of [
    '<a b="caf&eacute; &amp; &#88;" />',
    '<a b="x & y" />', // bare ampersand
    '<a b="&foo; &amp;" />', // unknown + known
    '<a b="x\r\ny" />', // CRLF preserved in attributes
    '<a b="&amp;amp;" />', // decode once
    '<a b="plain" />',
  ]) {
    test(JSON.stringify(src), () => {
      const got = collectJsxValues(runnerAst(src)).attr;
      const want = collectJsxValues(espreeAst(src)).attr;
      expect(got).toEqual(want);
    });
  }

  test('headline values pinned', () => {
    expect(
      collectJsxValues(runnerAst('<a b="caf&eacute; &amp; &#88;" />')).attr,
    ).toEqual(['café & X']);
    // CRLF NOT folded in attribute values (unlike text)
    expect(collectJsxValues(runnerAst('<a b="x\r\ny" />')).attr).toEqual([
      'x\r\ny',
    ]);
  });
});

// ─────────────────────────────────────────────────────────────────────
// Negative: ordinary JS string literals are NOT entity-decoded
// ─────────────────────────────────────────────────────────────────────

describe('JS string literals are NOT entity-decoded (matches espree)', () => {
  for (const src of [
    'const s = "x&amp;y";',
    '<a>{"x&amp;y caf&eacute;"}</a>', // JS string inside JSX expression
    'const t = "&#88; &nbsp;";',
  ]) {
    test(JSON.stringify(src), () => {
      const got = collectJsStringLiterals(runnerAst(src));
      const want = collectJsStringLiterals(espreeAst(src));
      expect(got).toEqual(want);
      // and explicitly: the entity text survives verbatim
      expect(got.some((v) => v.includes('&amp;') || v.includes('&#88;'))).toBe(
        true,
      );
    });
  }
});

// ─────────────────────────────────────────────────────────────────────
// Tokens: between-tags JSXText decoded; attribute string stays RAW
// ─────────────────────────────────────────────────────────────────────

describe('JSX tokens match espree', () => {
  for (const src of [
    '<a>caf&eacute; &amp; &#88;</a>',
    '<a>x\r\ny</a>',
    '<a href="x&amp;y caf&eacute;" />', // attr token RAW (with quotes)
    '<a>&amp;amp;</a>',
  ]) {
    test(JSON.stringify(src), () => {
      expect(runnerJsxTokens(src)).toEqual(espreeJsxTokens(src));
    });
  }

  test('headline: between-tags token decoded, attr token raw', () => {
    const textTok = runnerJsxTokens('<a>caf&eacute;</a>').find(
      (t) => t.type === 'JSXText',
    );
    expect(textTok?.value).toBe('café');
    // The attribute string token keeps the source verbatim incl. quotes.
    const attrTok = runnerJsxTokens('<a b="x&amp;y" />').find(
      (t) => t.type === 'JSXText',
    );
    expect(attrTok?.value).toBe('"x&amp;y"');
  });
});

// ─────────────────────────────────────────────────────────────────────
// decode helpers — direct unit edge cases
// ─────────────────────────────────────────────────────────────────────

describe('decodeJsxText / decodeJsxAttributeValue unit edges', () => {
  test('named, decimal, hex references', () => {
    expect(decodeJsxText('&amp;&lt;&gt;&quot;&apos;')).toBe('&<>"\'');
    expect(decodeJsxText('&#65;&#x42;')).toBe('AB');
    expect(decodeJsxText('&nbsp;')).toBe(' ');
  });

  test('invalid / unterminated references emit a literal "&"', () => {
    expect(decodeJsxText('&foo;')).toBe('&foo;');
    expect(decodeJsxText('a & b')).toBe('a & b');
    expect(decodeJsxText('&amp')).toBe('&amp'); // no semicolon
    expect(decodeJsxText('&;')).toBe('&;');
    expect(decodeJsxText('&#;')).toBe('&#;');
    expect(decodeJsxText('&#xZZ;')).toBe('&#xZZ;');
  });

  test('reference must terminate within 10 chars of "&"', () => {
    // 9 chars then ';' → still scanned (the ';' is the 10th char read).
    expect(decodeJsxText('&123456789;')).toBe('&123456789;'); // not numeric (#-less) → literal anyway
    // a `;` past the 10-char window is never reached → literal "&".
    expect(decodeJsxText('&abcdefghijklmnop;')).toBe('&abcdefghijklmnop;');
  });

  test('capital "&#X..;" is not treated as hex (acorn-jsx checks lowercase x)', () => {
    expect(decodeJsxText('&#X41;')).toBe('&#X41;');
  });

  test('astral numeric ref uses fromCharCode (truncation quirk)', () => {
    expect(decodeJsxText('&#x1F600;')).toBe(String.fromCharCode(0x1f600));
  });

  test('decodeJsxText folds CRLF→LF; lone CR/LF preserved', () => {
    expect(decodeJsxText('a\r\nb')).toBe('a\nb');
    expect(decodeJsxText('a\rb')).toBe('a\rb');
    expect(decodeJsxText('a\nb')).toBe('a\nb');
    // a CR produced by a reference is not re-folded against a source LF.
    expect(decodeJsxText('&#13;\n')).toBe('\r\n');
  });

  test('decodeJsxAttributeValue decodes but preserves CRLF', () => {
    expect(decodeJsxAttributeValue('&amp;')).toBe('&');
    expect(decodeJsxAttributeValue('x\r\ny')).toBe('x\r\ny');
  });

  test('fast path returns the same content for entity-free input', () => {
    expect(decodeJsxText('plain text')).toBe('plain text');
    expect(decodeJsxAttributeValue('plain text')).toBe('plain text');
  });
});
