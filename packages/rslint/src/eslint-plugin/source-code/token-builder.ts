/**
 * Rebuild ESLint-shape `Token[]` from the native parser's columnar token stream
 * (`@rslint/native`'s `parse()` returns `tokenTypes`/`tokenStarts`/`tokenEnds`,
 * all UTF-16 offsets). This is the JS half of the token path: the Rust
 * side already mapped each oxc `Kind` to an ESLint token *type* (parser-driven, so regex-vs-
 * division / template / JSX / TS `<` are correct); here we attach `value` (sliced from the
 * ORIGINAL source so lone surrogates survive), `loc` (lazy-line-column), and the
 * `regex` field, applying the few value side-paths that aren't a verbatim slice.
 *
 * Rebuild is lazy (driven by `SourceCode.ensureTokens` on first token-API use), so a file
 * whose rules never touch tokens pays nothing here.
 */
import {
  offsetToLineColumn,
  type SourceLocation,
} from '../ast/normalize-ast.js';
import { decodeJsxText } from './decode-entities.js';

export type TokenType =
  | 'Identifier'
  | 'PrivateIdentifier'
  | 'Keyword'
  | 'Punctuator'
  | 'String'
  | 'Numeric'
  | 'RegularExpression'
  | 'Template'
  | 'Boolean'
  | 'Null'
  | 'JSXIdentifier'
  | 'JSXText';

export interface Token {
  type: TokenType;
  value: string;
  range: [number, number];
  loc: SourceLocation;
  /**
   * Only on `RegularExpression` tokens — espree's shape. Plugins (`eslint-plugin-regexp`,
   * core `no-invalid-regexp` / `prefer-regex-literals`) read `token.regex.{pattern,flags}`.
   */
  regex?: { pattern: string; flags: string };
}

export interface Comment {
  // ESLint v10's three comment kinds. `Shebang` is a leading `#!` line.
  type: 'Line' | 'Block' | 'Shebang';
  value: string;
  range: [number, number];
  loc: SourceLocation;
}

// Indexed by the Rust `token_map::TokenType as u8` code. Codes 11 (JsxText) and 12
// (JsxTextAttr) BOTH surface as the ESLint "JSXText" type but take different value paths
// below, which is the whole reason the Rust side keeps them as distinct codes.
const TYPE_BY_CODE: readonly TokenType[] = [
  'Identifier', // 0
  'Keyword', // 1
  'Punctuator', // 2
  'String', // 3
  'Numeric', // 4
  'RegularExpression', // 5
  'Template', // 6
  'Boolean', // 7
  'Null', // 8
  'PrivateIdentifier', // 9
  'JSXIdentifier', // 10
  'JSXText', // 11  JsxText      → entity-decoded + CRLF-folded value
  'JSXText', // 12  JsxTextAttr  → raw value (incl. quotes), no decode
];

// Value side-paths keyed by code (everything else is a verbatim slice).
const CODE_IDENTIFIER = 0;
const CODE_REGEXP = 5;
const CODE_PRIVATE_IDENTIFIER = 9;
const CODE_JSX_TEXT = 11;

const REGEX_LITERAL_RE = /^\/(.*)\/([a-z]*)$/su;

/**
 * Rebuild the token array from the columnar native output. `text` is the ORIGINAL source
 * (BOM-stripped, same as the parser saw) so `slice(range)` recovers values verbatim,
 * including lone surrogates the Rust side can't represent. `lso` is the line-
 * start-offset table for lazy loc.
 */
export function buildTokens(
  types: Uint8Array,
  starts: Uint32Array,
  ends: Uint32Array,
  text: string,
  lso: number[],
): Token[] {
  const n = types.length;
  const out: Token[] = new Array(n);
  for (let i = 0; i < n; i++) {
    const code = types[i];
    const start = starts[i];
    const end = ends[i];
    const raw = text.slice(start, end);

    let value = raw;
    let regex: { pattern: string; flags: string } | undefined;
    switch (code) {
      case CODE_PRIVATE_IDENTIFIER:
        value = raw.slice(1); // strip leading '#': espree's value excludes it
        break;
      case CODE_IDENTIFIER:
        // espree decodes unicode escapes in identifier values (`abc` → `abc`).
        if (raw.indexOf('\\') !== -1) value = decodeIdentEscapes(raw);
        break;
      case CODE_JSX_TEXT:
        value = decodeJsxText(raw); // entity decode + CRLF→LF (JSX text only, NOT attr string)
        break;
      case CODE_REGEXP: {
        const m = REGEX_LITERAL_RE.exec(raw);
        regex = m ? { pattern: m[1], flags: m[2] } : { pattern: '', flags: '' };
        break;
      }
      // default (incl. JsxTextAttr=12): value is the verbatim slice.
    }

    const tok: Token = {
      type: TYPE_BY_CODE[code],
      value,
      range: [start, end],
      loc: {
        start: offsetToLineColumn(start, lso),
        end: offsetToLineColumn(end, lso),
      },
    };
    if (regex !== undefined) tok.regex = regex;
    out[i] = tok;
  }
  return out;
}

// ── identifier unicode-escape decoding ───────────────────────────────
// Migrated from the deleted hand-written tokenizer. espree resolves `\uXXXX` / `\u{…}` in an
// identifier's value while keeping `range` over the raw source span.

function decodeIdentEscapes(raw: string): string {
  let out = '';
  let i = 0;
  const n = raw.length;
  while (i < n) {
    if (
      raw.charCodeAt(i) === 92 /* \ */ &&
      raw.charCodeAt(i + 1) === 117 /* u */
    ) {
      const dec = decodeOneEscape(raw, i);
      if (dec != null) {
        out += String.fromCodePoint(dec.codepoint);
        i = dec.next;
        continue;
      }
    }
    out += raw[i];
    i++;
  }
  return out;
}

/** Decode a single `\uXXXX` / `\u{…}` escape at `i`, or null if malformed. */
function decodeOneEscape(
  raw: string,
  i: number,
): { codepoint: number; next: number } | null {
  // \u{ABCDE} variable-length form
  if (raw.charCodeAt(i + 2) === 123 /* { */) {
    let j = i + 3;
    while (j < raw.length && isHexDigit(raw.charCodeAt(j))) j++;
    if (raw.charCodeAt(j) !== 125 /* } */) return null;
    const hex = raw.slice(i + 3, j);
    if (hex.length === 0 || hex.length > 6) return null;
    const cp = parseInt(hex, 16);
    if (cp > 0x10ffff) return null;
    return { codepoint: cp, next: j + 1 };
  }
  // \uXXXX fixed 4-hex form
  if (i + 6 > raw.length) return null;
  for (let k = 0; k < 4; k++) {
    if (!isHexDigit(raw.charCodeAt(i + 2 + k))) return null;
  }
  return { codepoint: parseInt(raw.slice(i + 2, i + 6), 16), next: i + 6 };
}

function isHexDigit(ch: number): boolean {
  return (
    (ch >= 48 && ch <= 57) || (ch >= 65 && ch <= 70) || (ch >= 97 && ch <= 102)
  );
}

/**
 * Index of the first token at or after `offset` (binary search on `range[0]`), or -1 if
 * none. Migrated from the deleted tokenizer; used by the SourceCode token getters.
 */
export function tokenIndexAtOrAfter(
  tokens: readonly Token[],
  offset: number,
): number {
  let lo = 0;
  let hi = tokens.length;
  while (lo < hi) {
    const mid = (lo + hi) >> 1;
    if (tokens[mid].range[0] < offset) lo = mid + 1;
    else hi = mid;
  }
  return lo < tokens.length ? lo : -1;
}
