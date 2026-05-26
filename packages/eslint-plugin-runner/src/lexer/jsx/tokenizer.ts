/**
 * JSX lexer state machine, factored out of `../tokenizer.ts` for clarity.
 *
 * Two entry points are called from the main tokenize loop:
 *
 *   1. `dispatchJsxStateOnEntry` — runs at the START of every iteration
 *      when the JSX stack is non-empty. Owns JSXText scan and JSX
 *      tag-content scan; falls through (returns `handled: false`) when
 *      the stack top is an `expr` container so the main JS scanner can
 *      take over the inner expression.
 *
 *   2. `tryEnterJsxFromAngle` — called when the main JS scan sees `<`.
 *      Decides whether the `<` opens a JSX element (vs being a JS
 *      comparison or a TSX generic-parameter list) via
 *      `classifyJsxLAngle`. If JSX, pushes the appropriate state entry
 *      and emits the boundary Punctuator(s).
 *
 * The state lifecycle for one element looks like:
 *   `< Tag attrs >` → push `open-tag`     → tag-content scan
 *   `>`              → swap to `text`     → JSXText scan
 *   `</Tag>`         → swap to `close-tag`→ tag-content scan
 *   `>`              → pop
 *   `/>` (self-close on open-tag) skips the text + close-tag stages.
 *
 * Expression containers (`{ … }`) push an `expr` entry with `depth: 0`
 * and let the JS scanner take over. The matching `}` at `depth === 0`
 * pops the entry; nested `{`/`}` adjust `depth` and don't escape JSX.
 */

import {
  couldStartRegex,
  isIdentContinue,
  isIdentStart,
  isLineTerminator,
  isWhitespace,
  locOf,
  makeToken,
  type Token,
} from '../tokenizer.js';
import { decodeJsxText } from './decode-entities.js';

export type JsxContext =
  | { kind: 'open-tag' }
  | { kind: 'close-tag' }
  | { kind: 'text' }
  | { kind: 'expr'; depth: number };

/**
 * Decide whether `<` at offset `at` opens a JSX element. Returns
 *   - `'open'`  for an opening element / fragment (`<Tag …`, `<>` )
 *   - `'close'` if `<` is followed by `/` (`</Tag>`)
 *   - `null`   when the `<` should lex as the comparison operator
 *              (or a TSX generic — `<T,>` / `<T extends X>` /
 *              `<T>(...)` shapes are detected and rejected).
 *
 * The check is conservative: it requires the preceding JS token to
 * be at expression-prefix position (same gate `couldStartRegex` uses)
 * AND the lookahead after `<` to match a JSX-shaped start. TSX
 * disambiguation walks a few chars to spot `,` / `extends` / `>(`
 * follow-ups characteristic of generics.
 */
export function classifyJsxLAngle(
  text: string,
  at: number,
  n: number,
  tokens: readonly Token[],
): 'open' | 'close' | null {
  if (!couldStartRegex(tokens)) return null;
  let j = at + 1;
  while (
    j < n &&
    (isWhitespace(text.charCodeAt(j)) || isLineTerminator(text.charCodeAt(j)))
  ) {
    j++;
  }
  if (j >= n) return null;
  const c = text.charCodeAt(j);
  if (c === 47 /* / */) return 'close'; // </
  if (c === 62 /* > */) return 'open'; // <> fragment
  if (!isIdentStart(c)) return null;
  // Scan ident, then peek the follow-up shape.
  let k = j;
  while (k < n && isIdentContinue(text.charCodeAt(k))) k++;
  while (
    k < n &&
    (isWhitespace(text.charCodeAt(k)) || isLineTerminator(text.charCodeAt(k)))
  ) {
    k++;
  }
  if (k >= n) return null;
  const next = text.charCodeAt(k);
  // Generic shapes seen in TSX:
  //   `<T,`            — generic param separator
  //   `<T extends X>`  — generic constraint
  //   `<T>(`           — generic call / arrow params
  if (next === 44 /* , */) return null;
  if (text.startsWith('extends', k)) {
    const after = text.charCodeAt(k + 7);
    if (Number.isNaN(after) || isWhitespace(after) || isLineTerminator(after)) {
      return null;
    }
  }
  if (next === 62 /* > */) {
    let m = k + 1;
    while (
      m < n &&
      (isWhitespace(text.charCodeAt(m)) || isLineTerminator(text.charCodeAt(m)))
    ) {
      m++;
    }
    if (text.charCodeAt(m) === 40 /* ( */) return null; // <T>(…)
  }
  return 'open';
}

/**
 * Run the JSX state machine at the start of a tokenize iteration.
 * Returns the new `i` and `handled: true` if it consumed input —
 * caller should `continue` its main loop. `handled: false` means the
 * JSX stack top is `expr` and the JS scanner should take over.
 */
export function dispatchJsxStateOnEntry(
  text: string,
  i: number,
  n: number,
  lso: number[],
  tokens: Token[],
  jsxStack: JsxContext[],
): { i: number; handled: boolean } {
  if (jsxStack.length === 0) return { i, handled: false };
  const top = jsxStack[jsxStack.length - 1];

  // ── JSXText scan ──
  if (top.kind === 'text') {
    const start = i;
    while (i < n) {
      const cc = text.charCodeAt(i);
      if (cc === 60 /* < */ || cc === 123 /* { */) break;
      i++;
    }
    if (i > start) {
      tokens.push({
        type: 'JSXText',
        // espree's JSXText token value is entity-decoded with `\r\n`
        // folded to `\n` (acorn-jsx `jsx_readText`); `range` stays the
        // raw source span, so value length may differ from range width.
        value: decodeJsxText(text.slice(start, i)),
        range: [start, i],
        loc: locOf(start, i, lso),
      });
    }
    if (i >= n) return { i, handled: true };
    const c = text.charCodeAt(i);
    if (c === 60 /* < */) {
      // Element boundary: child element opens, OR this element's own
      // closing tag begins.
      //
      // Stack invariant: a CHILD element nests ON TOP of the parent's
      // `text` frame; it does NOT replace it. Only the parent's own
      // `</…>` close tag pops the `text` frame.
      //
      //   - `</…>` (close): this `text` frame's element is closing.
      //     Pop the frame and push `close-tag`; the trailing `>` then
      //     pops `close-tag`, unwinding one level.
      //   - `<child…>` (open): KEEP the parent `text` frame and push
      //     the child's `open-tag` on top. The child's own self-close
      //     (`/>`) or `>`-close pops back to this same `text` frame, so
      //     sibling text/elements after the child resume as JSXText
      //     instead of falling through to the JS scanner.
      if (text.charCodeAt(i + 1) === 47 /* / */) {
        jsxStack.pop();
        tokens.push(makeToken('Punctuator', text, i, i + 1, lso));
        tokens.push(makeToken('Punctuator', text, i + 1, i + 2, lso));
        i += 2;
        jsxStack.push({ kind: 'close-tag' });
      } else {
        tokens.push(makeToken('Punctuator', text, i, i + 1, lso));
        i++;
        jsxStack.push({ kind: 'open-tag' });
      }
      return { i, handled: true };
    }
    if (c === 123 /* { */) {
      // Expression container inside JSX text.
      tokens.push(makeToken('Punctuator', text, i, i + 1, lso));
      i++;
      jsxStack.push({ kind: 'expr', depth: 0 });
      return { i, handled: true };
    }
    return { i, handled: true };
  }

  // ── JSX tag-content scan (open-tag or close-tag) ──
  if (top.kind === 'open-tag' || top.kind === 'close-tag') {
    const c = text.charCodeAt(i);
    // Whitespace inside tag header is non-significant.
    if (isWhitespace(c) || isLineTerminator(c)) {
      return { i: i + 1, handled: true };
    }
    // `>` closes the current tag.
    if (c === 62 /* > */) {
      tokens.push(makeToken('Punctuator', text, i, i + 1, lso));
      if (top.kind === 'open-tag') {
        jsxStack.pop();
        jsxStack.push({ kind: 'text' });
      } else {
        jsxStack.pop();
      }
      return { i: i + 1, handled: true };
    }
    // `/>` self-closing element.
    if (c === 47 /* / */ && text.charCodeAt(i + 1) === 62 /* > */) {
      tokens.push(makeToken('Punctuator', text, i, i + 1, lso));
      tokens.push(makeToken('Punctuator', text, i + 1, i + 2, lso));
      jsxStack.pop();
      return { i: i + 2, handled: true };
    }
    // `.` / `:` / `=` separators in tag header.
    if (c === 46 /* . */ || c === 58 /* : */ || c === 61 /* = */) {
      tokens.push(makeToken('Punctuator', text, i, i + 1, lso));
      return { i: i + 1, handled: true };
    }
    // `{` opens an expression-container attribute value / child
    // expression. Push `expr` so the JS scanner handles the inner.
    if (c === 123 /* { */) {
      tokens.push(makeToken('Punctuator', text, i, i + 1, lso));
      jsxStack.push({ kind: 'expr', depth: 0 });
      return { i: i + 1, handled: true };
    }
    // Quoted attribute value — emitted as `JSXText` (espree contract;
    // quotes are part of `value`).
    //
    // A backslash is an ORDINARY character here: espree / acorn-jsx do
    // NOT process C-style escapes in a JSX attribute string
    // (`jsx_readString` reads raw bytes between the quotes). So
    // `value="C:\"` ends the string at the `"` right after the
    // backslash — treating `\"` as an escaped quote (the old `i += 2`)
    // over-consumed past the real close quote and corrupted every
    // following token.
    if (c === 34 /* " */ || c === 39 /* ' */) {
      const quote = c;
      const start = i;
      i++;
      while (i < n) {
        if (text.charCodeAt(i) === quote) {
          i++;
          break;
        }
        i++;
      }
      // EOF clamp — mirrors the block-comment scanner
      // (`lexer/tokenizer.ts` `Math.min(i + 2, n)`) so a malformed
      // unterminated attribute string can never push the token range
      // past the source end.
      const end = Math.min(i, n);
      tokens.push({
        type: 'JSXText',
        value: text.slice(start, end),
        range: [start, end],
        loc: locOf(start, end, lso),
      });
      return { i: end, handled: true };
    }
    // JSXIdentifier — letters/digits/_/$ plus mid-name `-`.
    if (isIdentStart(c)) {
      const start = i;
      i++;
      while (i < n) {
        const cc = text.charCodeAt(i);
        if (isIdentContinue(cc) || cc === 45 /* - */) i++;
        else break;
      }
      tokens.push({
        type: 'JSXIdentifier',
        value: text.slice(start, i),
        range: [start, i],
        loc: locOf(start, i, lso),
      });
      return { i, handled: true };
    }
    // Unexpected — skip one byte so we don't infinite-loop. The
    // parser will report the syntax error.
    return { i: i + 1, handled: true };
  }

  // `expr` top — caller's JS scanner takes over.
  return { i, handled: false };
}

/**
 * If `text[i] === '<'` and a JSX entry is plausible, push the right
 * state entry and emit the boundary Punctuator(s). Returns
 * `{ entered: true }` when JSX was entered (caller should `continue`).
 */
export function tryEnterJsxFromAngle(
  text: string,
  i: number,
  n: number,
  lso: number[],
  tokens: Token[],
  jsxStack: JsxContext[],
): { i: number; entered: boolean } {
  const kind = classifyJsxLAngle(text, i, n, tokens);
  if (kind === 'open') {
    tokens.push(makeToken('Punctuator', text, i, i + 1, lso));
    jsxStack.push({ kind: 'open-tag' });
    return { i: i + 1, entered: true };
  }
  if (kind === 'close') {
    // Defensive — the JSX-text top-of-loop handler should normally
    // own this path. Keep here for safety on hand-constructed inputs.
    tokens.push(makeToken('Punctuator', text, i, i + 1, lso));
    tokens.push(makeToken('Punctuator', text, i + 1, i + 2, lso));
    jsxStack.push({ kind: 'close-tag' });
    return { i: i + 2, entered: true };
  }
  return { i, entered: false };
}

/**
 * Tracks brace depth on `{`/`}` Punctuators emitted while the top of
 * the JSX stack is an `expr` container. Combined with the depth-0
 * check that pops the `expr` entry, this ensures nested JS braces
 * (object literals, blocks, template `${…}`) don't accidentally
 * escape the JSX expression container.
 */
export function trackExprBraceDepth(jsxStack: JsxContext[], ch: number): void {
  if (jsxStack.length === 0) return;
  const top = jsxStack[jsxStack.length - 1];
  if (top.kind !== 'expr') return;
  if (ch === 123 /* { */) top.depth++;
  else if (ch === 125 /* } */) top.depth--;
}

/**
 * If `text[i] === '}'` AND the JSX stack top is an `expr` at depth 0,
 * the brace closes the expression container — pop the entry and
 * emit `}` Punctuator. Returns `true` if handled (caller continues).
 */
export function tryCloseJsxExprContainer(
  text: string,
  i: number,
  lso: number[],
  tokens: Token[],
  jsxStack: JsxContext[],
): boolean {
  if (jsxStack.length === 0) return false;
  const top = jsxStack[jsxStack.length - 1];
  if (top.kind !== 'expr' || top.depth !== 0) return false;
  jsxStack.pop();
  tokens.push(makeToken('Punctuator', text, i, i + 1, lso));
  return true;
}
