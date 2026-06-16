/**
 * JSX value decoding — AST differential vs espree.
 *
 * `normalize-ast`'s `applyJsxValueDecoding` decodes JSX character references
 * into the ESTree node `.value` (espree / acorn-jsx parity):
 *   - JSXText           → decode + CRLF→LF fold
 *   - JSXAttribute str  → decode only, CRLF preserved
 *   - ordinary JS string Literal → NOT decoded (negative guard)
 *
 * Ported (AST half) from the deleted `lexer/jsx-entity-decoding.test.ts`
 * after the M4 tokenizer removal: the TOKEN half now lives in
 * `token-differential.test.ts`, but the AST path (`normalize-ast.ts:381`
 * `applyJsxValueDecoding`) otherwise lost its coverage. espree is the
 * oracle — every case runs through the native parser AND espree.
 */
import { describe, test, expect } from '@rstest/core';
import * as espree from 'espree';
import { parse as nativeParse } from '../../../src/eslint-plugin/native/load-binding.js';

import {
  normalizeAst,
  buildLineStartOffsets,
} from '../../../src/eslint-plugin/ast/normalize-ast.js';

interface AnyObj {
  type?: string;
  value?: unknown;
  [k: string]: unknown;
}

/** Collect, in document order, every JSXText value and every string-valued
 *  JSXAttribute value from a parsed program. */
function collectJsxValues(root: unknown): { text: string[]; attr: string[] } {
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
  const parsed = nativeParse('test.jsx', src, 'module', true);
  const ast = JSON.parse(parsed.program);
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
      expect(collectJsxValues(runnerAst(src)).text).toEqual(
        collectJsxValues(espreeAst(src)).text,
      );
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
    expect(collectJsxValues(runnerAst('<a>&amp;amp;</a>')).text).toEqual([
      '&amp;',
    ]); // decoded exactly once
  });
});

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
      expect(collectJsxValues(runnerAst(src)).attr).toEqual(
        collectJsxValues(espreeAst(src)).attr,
      );
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

describe('JS string literals are NOT entity-decoded (matches espree)', () => {
  for (const src of [
    'const s = "x&amp;y";',
    '<a>{"x&amp;y caf&eacute;"}</a>', // JS string inside JSX expression
    'const t = "&#88; &nbsp;";',
  ]) {
    test(JSON.stringify(src), () => {
      const got = collectJsStringLiterals(runnerAst(src));
      expect(got).toEqual(collectJsStringLiterals(espreeAst(src)));
      // and explicitly: the entity text survives verbatim
      expect(got.some((v) => v.includes('&amp;') || v.includes('&#88;'))).toBe(
        true,
      );
    });
  }
});
