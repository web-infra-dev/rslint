/**
 * Unit edges for JSX entity decoding (`decodeJsxText` / `decodeJsxAttributeValue`), ported from
 * the deleted `lexer/jsx-entity-decoding` suite when the hand-written tokenizer was removed (M4).
 * These acorn-jsx bug-for-bug ports are now consumed by the token rebuild (JSXText value) and
 * by normalize-ast (JSX node values), so their edge cases — especially `decodeJsxAttributeValue`,
 * which only the AST path uses — stay pinned directly here.
 */
import { describe, test, expect } from '@rstest/core';

import {
  decodeJsxText,
  decodeJsxAttributeValue,
} from '../../../src/eslint-plugin/source-code/decode-entities.js';

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
    expect(decodeJsxText('&123456789;')).toBe('&123456789;'); // not numeric (#-less) → literal
    expect(decodeJsxText('&abcdefghijklmnop;')).toBe('&abcdefghijklmnop;'); // ';' past the 10-char window
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
    expect(decodeJsxText('&#13;\n')).toBe('\r\n'); // a CR from a ref is not re-folded against a source LF
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
