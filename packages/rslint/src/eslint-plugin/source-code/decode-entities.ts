// cspell:ignore Ingvar Stepanyan
/**
 * Decode JSX text / string-attribute values exactly as ESLint v10 does.
 *
 * ESLint's default parser (espree) lexes JSX through acorn-jsx, which
 * resolves character references while reading JSX text (`jsx_readText`)
 * and attribute strings (`jsx_readString`). The runner's native parser
 * (`@rslint/native`) leaves both `.value` and `.raw` as the verbatim source, so
 * the runner must reproduce acorn-jsx's decoding to make JSX node /
 * token `.value`s byte-identical to ESLint v10.
 *
 * Ported (bug-for-bug) from acorn-jsx@5.3.2 `index.js`:
 *   - `jsx_readEntity`  → {@link readEntityAt} (named + numeric refs)
 *   - `jsx_readNewLine` → the CRLF→LF branch in {@link decodeJsxText}
 *   https://github.com/acornjs/acorn-jsx/blob/main/index.js
 *   MIT License — Copyright (C) 2012-2017 by Ingvar Stepanyan.
 *
 * Two intentional fidelity quirks, preserved to match espree output:
 *   1. Numeric references use `String.fromCharCode` (NOT
 *      `fromCodePoint`), so astral references (`&#x1F600;`) are
 *      truncated to a single UTF-16 unit, exactly as acorn-jsx does.
 *   2. A reference must terminate with `;` within 10 characters of the
 *      `&`; otherwise the `&` is emitted literally and scanning resumes
 *      at the next character.
 */
import { XHTML_ENTITIES } from './entities.js';

// acorn-jsx: `const hexNumber = /^[\da-fA-F]+$/`, `decimalNumber = /^\d+$/`.
const HEX_NUMBER = /^[\da-fA-F]+$/;
const DECIMAL_NUMBER = /^\d+$/;

/**
 * Resolve the character reference that begins at `raw[ampIndex]` (which
 * must be `&`). Mirrors acorn-jsx `jsx_readEntity`.
 *
 * @returns the decoded replacement and the index to continue scanning
 *   from. When the reference is invalid / unterminated, `decoded` is the
 *   literal `'&'` and `next` is `ampIndex + 1` (acorn-jsx resets `pos`
 *   to just past the ampersand and emits a bare `&`).
 */
function readEntityAt(
  raw: string,
  ampIndex: number,
): { decoded: string; next: number } {
  let str = '';
  let count = 0;
  let entity: string | undefined;
  let pos = ampIndex + 1;
  const len = raw.length;
  // acorn-jsx caps the scan at 10 characters past the `&` (`count++ < 10`).
  while (pos < len && count++ < 10) {
    const ch = raw[pos++];
    if (ch === ';') {
      if (str[0] === '#') {
        if (str[1] === 'x') {
          const hex = str.slice(2);
          if (HEX_NUMBER.test(hex)) {
            entity = String.fromCharCode(parseInt(hex, 16));
          }
        } else {
          const dec = str.slice(1);
          if (DECIMAL_NUMBER.test(dec)) {
            entity = String.fromCharCode(parseInt(dec, 10));
          }
        }
      } else {
        entity = XHTML_ENTITIES[str];
      }
      break;
    }
    str += ch;
  }
  if (entity === undefined) {
    // Invalid / unterminated reference: emit a bare `&` and resume right
    // after it so the accumulated characters are re-scanned as text.
    return { decoded: '&', next: ampIndex + 1 };
  }
  return { decoded: entity, next: pos };
}

/**
 * Decode the value of a JSX **text** node / token (the characters
 * between tags). Resolves character references AND normalizes a source
 * `\r\n` to `\n` — acorn-jsx `jsx_readText` runs `jsx_readNewLine(true)`
 * on line breaks. A lone `\r` (no following `\n`) is left as-is, and a
 * `\r` produced by a numeric reference (`&#13;`) is NOT re-normalized,
 * so both transforms happen in a single source pass.
 */
export function decodeJsxText(raw: string): string {
  // Fast path: nothing to decode and no CRLF to fold.
  if (!raw.includes('&') && !raw.includes('\r')) return raw;
  let out = '';
  let i = 0;
  const len = raw.length;
  while (i < len) {
    const ch = raw[i];
    if (ch === '&') {
      const { decoded, next } = readEntityAt(raw, i);
      out += decoded;
      i = next;
    } else if (ch === '\r' && raw[i + 1] === '\n') {
      out += '\n';
      i += 2;
    } else {
      out += ch;
      i++;
    }
  }
  return out;
}

/**
 * Decode the value of a JSX **string attribute** (`foo="…"`), given the
 * already-unquoted inner text. Resolves character references but does
 * NOT normalize line endings — acorn-jsx `jsx_readString` runs
 * `jsx_readNewLine(false)`, which preserves `\r\n` verbatim.
 */
export function decodeJsxAttributeValue(raw: string): string {
  if (!raw.includes('&')) return raw;
  let out = '';
  let i = 0;
  const len = raw.length;
  while (i < len) {
    const ch = raw[i];
    if (ch === '&') {
      const { decoded, next } = readEntityAt(raw, i);
      out += decoded;
      i = next;
    } else {
      out += ch;
      i++;
    }
  }
  return out;
}
