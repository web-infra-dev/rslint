/* eslint-disable @typescript-eslint/no-unsafe-type-assertion -- AST / parser / scope-manager / plugin-API boundary casts. Each site projects from an `any` / `unknown` peer surface (oxc-parser output, user plugin objects, ESLint v10 wire shapes) into the typed shape this module uses; the contract is runtime-validated at the call boundaries above, not at the cast. Bulk-disabling here instead of per-line keeps the cast sites readable. */
/**
 * AST normalization for oxc-parser output → ESLint-shape ESTree.
 *
 * Spike 2 finding (factual, validated against eslint-plugin-unicorn):
 * oxc-parser already emits ESTree-compatible nodes for the common cases —
 * including `Literal{value: null}` directly (no `NullLiteral` rewrite
 * needed). What `normalizeAst` actually does today is:
 *
 *   1. Add `range: [start, end]` and `loc: {start, end}` to every node
 *      (oxc emits `start`/`end` as UTF-16 code-unit offsets — verified
 *      empirically; not byte offsets, contrary to oxc's own README
 *      wording in some versions).
 *   2. Add `parent` references so rules can walk up the tree.
 *   3. Apply a small set of TS-specific shape fixes that
 *      `@typescript-eslint/scope-manager` and ESLint TS plugins expect.
 *      The set is empirical: drive additions by a fixture failing in the
 *      conformance harness, NOT by guesswork.
 *
 * Single DFS — every transform runs in one tree traversal. Roughly O(N)
 * over the AST.
 *
 * Mutates the AST in place. Callers must not share an AST across
 * files. oxc-parser returns a fresh tree per parse, so the standard
 * per-file lint flow inside the worker is naturally safe — the
 * constraint matters only if some caller decides to cache trees.
 */

import { VISITOR_KEYS, getVisitorKeys } from './visitor-keys.js';
import {
  decodeJsxText,
  decodeJsxAttributeValue,
} from '../lexer/jsx/decode-entities.js';

export interface LocPosition {
  line: number;
  column: number;
}

export interface SourceLocation {
  start: LocPosition;
  end: LocPosition;
}

/**
 * Build a line-start-offset table for fast offset → (line, column)
 * conversion. Index 0 holds the offset of line 1's first character (0).
 *
 * Recognises every ECMAScript LineTerminator (§ 11.3): LF (U+000A),
 * CR (U+000D), `\r\n` (one terminator), LS (U+2028), PS (U+2029).
 * `source-code.ts`'s `getLocFromIndex` and the AST normaliser both
 * call this so a file containing any of those terminators reports
 * identical `(line, column)` regardless of which surface the caller
 * went through — `node.loc.line` matches `sourceCode.getLocFromIndex(
 * node.range[0]).line` byte-for-byte. Pre-fix this function only
 * split on LF and disagreed with the SourceCode side on bare CR / LS
 * / PS files.
 *
 * Offsets are UTF-16 code-unit indices — the same unit JavaScript
 * strings use natively (`.length`, `.charCodeAt`, `.slice`). The
 * runner operates entirely in this unit; conversion to UTF-8 bytes
 * happens at the IPC boundary when handing diagnostics to the Go
 * side.
 *
 * O(N) over the source text. Reused across all nodes in a single file.
 */
export function buildLineStartOffsets(text: string): number[] {
  const offsets = [0];
  for (let i = 0; i < text.length; i++) {
    const ch = text.charCodeAt(i);
    if (ch === 0x0a /* \n */) {
      offsets.push(i + 1);
    } else if (ch === 0x0d /* \r */) {
      // `\r\n` is ONE terminator — skip the trailing LF so the next
      // iteration doesn't count it as a second break.
      if (text.charCodeAt(i + 1) === 0x0a) {
        offsets.push(i + 2);
        i++;
      } else {
        offsets.push(i + 1);
      }
    } else if (ch === 0x2028 /* LS */ || ch === 0x2029 /* PS */) {
      offsets.push(i + 1);
    }
  }
  return offsets;
}

/**
 * Build a UTF-16-code-unit → UTF-8-byte lookup table for the source.
 *
 * Why this exists:
 *
 *   Inside the runner, ranges are UTF-16 code-unit indices — matching
 *   the way JavaScript strings are natively indexed (`.length`,
 *   `.charCodeAt(i)`, `.slice(a, b)`). oxc-parser also emits offsets
 *   in this same UTF-16 unit, so token API / `sourceCode.getText` all
 *   round-trip naturally.
 *
 *   The Go side, however, uses TypeScript's `scanner.GetECMALineAnd
 *   UTF16CharacterOfPosition`, whose `pos` parameter is documented as
 *   a UTF-8 BYTE offset (we verified this empirically — see
 *   `TestScannerPosUnit_UTF8Bytes` in `internal/linter`). Without
 *   converting at the IPC boundary, Go interprets the runner's
 *   utf-16 char offset as a byte position. On ASCII-only files the
 *   two units coincide so nothing's visibly wrong; the moment a file
 *   contains an emoji / CJK / arrow / any multi-byte UTF-8 character,
 *   every diagnostic AFTER that character is reported with a column
 *   shift of (utf8_bytes_consumed_so_far - utf16_units_consumed).
 *
 *   This map provides the conversion: `utf16ToByte[utf16Idx]` is the
 *   UTF-8 byte offset where the character at utf-16 index `utf16Idx`
 *   begins. Callers convert each diagnostic's `startPos`/`endPos`/fix
 *   range right before shipping the result back to Go.
 *
 *   For a surrogate pair (BMP astral chars: emoji etc.), both the high
 *   and low surrogate utf-16 indices map to the SAME byte offset (the
 *   start of the 4-byte UTF-8 sequence). The end-of-string sentinel
 *   `map[text.length]` holds the total byte length so callers can
 *   convert `endPos` (one-past-end) without a bounds check.
 */
export function buildUtf16ToByteMap(text: string): number[] {
  const len = text.length;
  const map = new Array<number>(len + 1);
  let byteIdx = 0;
  let i = 0;
  while (i < len) {
    map[i] = byteIdx;
    const c = text.charCodeAt(i);
    if (c < 0x80) {
      byteIdx += 1;
      i++;
    } else if (c < 0x800) {
      byteIdx += 2;
      i++;
    } else if (c >= 0xd800 && c <= 0xdbff && i + 1 < len) {
      const next = text.charCodeAt(i + 1);
      if (next >= 0xdc00 && next <= 0xdfff) {
        // Surrogate pair → astral plane codepoint → 4 UTF-8 bytes.
        // Both halves of the pair map to the start of the 4-byte run.
        map[i + 1] = byteIdx;
        byteIdx += 4;
        i += 2;
      } else {
        // Unpaired high surrogate. WHATWG-style encoders substitute
        // U+FFFD (3 bytes); match that to stay consistent with how
        // Go's runtime (and most other UTF-8 encoders) handle the case.
        byteIdx += 3;
        i++;
      }
    } else {
      byteIdx += 3;
      i++;
    }
  }
  map[len] = byteIdx;
  return map;
}

/**
 * Convert a UTF-16 code-unit offset into a 1-indexed line / 0-indexed
 * column pair using `lineStartOffsets`. Binary search; O(log N) per
 * lookup.
 *
 * ESLint convention: line is 1-based, column is 0-based. We match
 * that. The `column` field of the returned `LocPosition` is therefore
 * a UTF-16 code-unit index relative to the line start — the same
 * unit ESLint's own `loc.column` uses.
 */
export function offsetToLineColumn(
  offset: number,
  lineStartOffsets: number[],
): LocPosition {
  let lo = 0;
  let hi = lineStartOffsets.length - 1;
  while (lo < hi) {
    const mid = (lo + hi + 1) >> 1;
    if (lineStartOffsets[mid] <= offset) lo = mid;
    else hi = mid - 1;
  }
  return { line: lo + 1, column: offset - lineStartOffsets[lo] };
}

/**
 * Set of node types whose `decorators`/`extends`/`implements` properties
 * @typescript-eslint/scope-manager iterates over with `forEach`. If those
 * fields are absent, scope-manager throws on `undefined.forEach`. We
 * default them to empty arrays.
 */
const NEEDS_DECORATORS = new Set<string>([
  'ClassDeclaration',
  'ClassExpression',
  'PropertyDefinition',
  'MethodDefinition',
  'AccessorProperty',
  'TSAbstractMethodDefinition',
  'TSAbstractPropertyDefinition',
  'TSAbstractAccessorProperty',
]);

/**
 * Stats reported by `normalizeAst` for diagnostics / telemetry. Useful
 * during development of new fixtures: if `transformsApplied` is zero on a
 * fixture that looks like it should hit transforms, the AST shape may
 * already be ESLint-compatible (a good thing).
 */
export interface NormalizeStats {
  nodesVisited: number;
  rangeLocAdded: number;
  parentLinked: number;
  transformsApplied: number;
}

/**
 * Normalize an oxc-parser AST in place. Adds range/loc/parent and applies
 * the empirical set of TS shape fixes. Returns visit / transform counts
 * for inspection by tests.
 *
 * The AST root is typically a `Program` node. Children are recursed via
 * generic key-walking — we don't hardcode a visitor table because oxc's
 * shape may evolve and the parent-link is needed everywhere anyway.
 */
export function normalizeAst(
  root: AnyNode,
  lineStartOffsets: number[],
  text: string,
): NormalizeStats {
  const stats: NormalizeStats = {
    nodesVisited: 0,
    rangeLocAdded: 0,
    parentLinked: 0,
    transformsApplied: 0,
  };
  walk(root, null, lineStartOffsets, text, stats);
  return stats;
}

interface AnyNode {
  type?: string;
  start?: number;
  end?: number;
  range?: [number, number];
  loc?: SourceLocation;
  parent?: AnyNode | null;
  // Anything else — we walk reflectively
  [key: string]: unknown;
}

function walk(
  node: AnyNode | null | undefined,
  parent: AnyNode | null,
  lso: number[],
  text: string,
  stats: NormalizeStats,
): void {
  if (node == null || typeof node !== 'object') return;
  stats.nodesVisited++;

  // ── 1. range / loc ──────────────────────────────────────────────
  //
  // `range` and `loc` are normally absent from the oxc-parser AST
  // and we fill them here. The branches are split so that if a
  // caller passes `range: true` to parseSync (oxc supports it
  // natively — see comment in ecma-language-plugin.ts for why
  // production does NOT) the `range` half is skipped and we only
  // pay for `loc`, which oxc has no equivalent option for. The
  // independent gates make the optimization free to re-enable
  // upstream without touching this file.
  if (typeof node.start === 'number' && typeof node.end === 'number') {
    let added = false;
    // espree (JS) and @typescript-eslint/parser (TS) both include the
    // surrounding template delimiters in a `TemplateElement`'s range —
    // e.g. `` `a${ `` and `` }c` `` for `` `a${b}c` ``. The desired
    // output is always that delimiter-inclusive range. oxc-parser,
    // however, emits `node.start`/`node.end` differently per LANGUAGE
    // (verified against oxc-parser 0.132.0):
    //
    //   • `.js/.jsx/.mjs/.cjs` → COOKED (delimiter-LESS): first elem of
    //     `` `a${b}c` `` is `[9,10]` (just `a`). Here we must expand
    //     back over the 1-char opening delimiter (`` ` `` first elem,
    //     `}` continuation) and forward over the closing delimiter
    //     (`` ` `` if `tail`, 1 char; else `${`, 2 chars) → `[8,12]`.
    //
    //   • `.ts/.tsx/.mts/.cts` → DELIMITER-INCLUSIVE already: the same
    //     element is `[8,12]`. Expanding here would DOUBLE-COUNT the
    //     delimiters → `[7,14]`, which diverges from both oracles.
    //
    // Robust language-agnostic detector: oxc's element span width
    // (`node.end - node.start`) equals `node.value.raw.length` in the
    // cooked mode, but is GREATER (by the delimiter widths: 2 or 3)
    // when the offsets already include delimiters. `.raw` is supplied
    // in BOTH modes and in all element shapes (first/continuation/
    // tail/EMPTY/escaped/nested), so width>rawLen exactly distinguishes
    // the TS delimiter-inclusive case from the JS cooked case. We only
    // expand when oxc's offsets are cooked.
    const isTpl = node.type === 'TemplateElement';
    let rStart = node.start;
    let rEnd = node.end;
    if (isTpl) {
      const rawLen =
        (node.value as { raw?: string } | undefined)?.raw?.length ??
        node.end - node.start;
      const alreadyDelimited = node.end - node.start > rawLen;
      if (!alreadyDelimited) {
        rStart = node.start - 1;
        rEnd = node.end + (node.tail ? 1 : 2);
      }
    }
    // For a TemplateElement, assign the (possibly expanded) range
    // UNCONDITIONALLY — overwriting any range oxc may have supplied.
    // `loc` below always uses the same `rStart`/`rEnd`, so gating the
    // range on `range == null` (as non-TemplateElement nodes are) would
    // risk getting the two out of sync: with oxc's `range: true`, oxc fills a
    // range equal to `[start, end]` (cooked on `.js`, delimiter-
    // inclusive on `.ts` — same per-language distinction the detector
    // reads), the gate would keep it, and `range` would stay un-
    // normalized while `loc` reflects the expansion on JS. Recomputing
    // both from `rStart`/`rEnd` keeps them in lockstep and aligned with
    // the oracles in both languages. Other node types keep the
    // `range == null` gate so a future oxc-provided range is preserved
    // verbatim.
    if (isTpl || node.range == null) {
      node.range = [rStart, rEnd];
      added = true;
    }
    if (node.loc == null) {
      node.loc = {
        start: offsetToLineColumn(rStart, lso),
        end: offsetToLineColumn(rEnd, lso),
      };
      added = true;
    }
    if (added) stats.rangeLocAdded++;
  }

  // ── 2. parent ───────────────────────────────────────────────────
  if (parent != null) {
    node.parent = parent;
    stats.parentLinked++;
  }

  // ── 3. TS-specific shape fixes ──────────────────────────────────
  applyTsShapeTransforms(node, stats);

  // ── 3b. JSX character-reference decoding ────────────────────────
  applyJsxValueDecoding(node, stats);

  // ── 4. Recurse into children ────────────────────────────────────
  //
  // Use the canonical visitor-keys table — same one ESLint v10's
  // Traverser consults — instead of `Object.keys(node)` + a blacklist.
  // That keeps the recursion set identical to what espree /
  // @typescript-eslint/parser declare AND avoids the per-node
  // short-lived `Object.keys` array allocation (a measurable GC
  // pressure on 5000+ file lints — see visitor-keys.ts doc).
  const t = node.type;
  const keys =
    (t !== undefined ? VISITOR_KEYS[t] : undefined) ??
    getVisitorKeys(node as { type?: string });
  for (let i = 0; i < keys.length; i++) {
    const v = node[keys[i]];
    if (v == null) continue;
    if (Array.isArray(v)) {
      for (let j = 0; j < v.length; j++) {
        const item = v[j];
        if (item && typeof item === 'object' && (item as AnyNode).type) {
          walk(item as AnyNode, node, lso, text, stats);
        }
      }
    } else if (typeof v === 'object' && (v as AnyNode).type) {
      walk(v as AnyNode, node, lso, text, stats);
    }
  }
}

/**
 * Decode JSX character references in `.value`, in place, to match
 * ESLint v10 (espree/acorn-jsx). oxc-parser leaves JSX `.value` as the
 * verbatim source; espree resolves entities (and folds `\r\n`→`\n` in
 * text). See {@link decodeJsxText} / {@link decodeJsxAttributeValue}.
 *
 * Two JSX `.value` surfaces, decoded EXACTLY once each:
 *   - `JSXText.value` (text between tags) — decode + CRLF fold.
 *   - the `Literal` value of a `JSXAttribute` (`foo="…"`) — decode only.
 *
 * The attribute case is handled here at the `JSXAttribute` (reaching
 * into its `value`), NOT when the inner `Literal` is later visited:
 * decoding is not idempotent (`&amp;amp;`→`&amp;`→`&`), so the node must
 * be touched once, and a bare `Literal` is indistinguishable from an
 * ordinary JS string literal (which must NOT be entity-decoded).
 *
 * Gated by `node.type`, so non-JSX nodes pay one or two string compares.
 */
function applyJsxValueDecoding(node: AnyNode, stats: NormalizeStats): void {
  const t = node.type;
  if (t === 'JSXText') {
    if (typeof node.raw === 'string') {
      node.value = decodeJsxText(node.raw);
      stats.transformsApplied++;
    }
    return;
  }
  if (t === 'JSXAttribute') {
    const v = node.value as AnyNode | null | undefined;
    if (
      v != null &&
      typeof v === 'object' &&
      v.type === 'Literal' &&
      typeof v.value === 'string'
    ) {
      // oxc's `Literal.value` is the inner (unquoted) attribute text,
      // verbatim; `.raw` keeps the surrounding quotes. Decode the inner
      // text; attribute values do NOT fold `\r\n`.
      v.value = decodeJsxAttributeValue(v.value);
      stats.transformsApplied++;
    }
  }
}

/**
 * In-place TS-shape adjustments. Each branch is gated by `node.type`, so
 * JS files (which never have TS-specific node types) are not affected —
 * the cost on a JS file is one type check per node.
 */
function applyTsShapeTransforms(node: AnyNode, stats: NormalizeStats): void {
  // typeParameters → typeArguments rename for instantiation contexts.
  // ESLint's @typescript-eslint/parser uses `typeArguments` in these
  // positions; oxc may still emit `typeParameters`. The fix is conditional
  // on the inner node type so we don't accidentally rename function-decl
  // typeParameters (which keep their original name).
  if (
    node.typeParameters != null &&
    typeof node.typeParameters === 'object' &&
    (node.typeParameters as AnyNode).type === 'TSTypeParameterInstantiation' &&
    node.typeArguments == null
  ) {
    node.typeArguments = node.typeParameters;
    delete (node as Record<string, unknown>).typeParameters;
    stats.transformsApplied++;
  }
  // superTypeParameters → superTypeArguments (analogous, on classes).
  if (
    node.superTypeParameters != null &&
    typeof node.superTypeParameters === 'object' &&
    node.superTypeArguments == null
  ) {
    node.superTypeArguments = node.superTypeParameters;
    delete (node as Record<string, unknown>).superTypeParameters;
    stats.transformsApplied++;
  }

  // TSMappedType: scope-manager expects `key`, oxc emits typeParameter.name.
  if (
    node.type === 'TSMappedType' &&
    node.typeParameter != null &&
    typeof node.typeParameter === 'object' &&
    (node.typeParameter as AnyNode).name != null &&
    node.key == null
  ) {
    node.key = (node.typeParameter as AnyNode).name;
    stats.transformsApplied++;
  }

  // decorators[] / extends[] / implements[] default arrays — scope-manager
  // calls `.forEach` on these, which throws if the field is missing.
  if (NEEDS_DECORATORS.has(node.type ?? '')) {
    if (node.decorators == null) {
      node.decorators = [];
      stats.transformsApplied++;
    }
  }
  if (node.type === 'ClassDeclaration' || node.type === 'ClassExpression') {
    if (node['extends'] === undefined) {
      // intentionally use bracket access to avoid 'extends' keyword issues
      (node as Record<string, unknown>)['extends'] = null;
      // this is a "set to null" rather than "default array" because most
      // TS class declarations have at most one extends — but scope-manager
      // checks `node.extends?.forEach`, so null is fine.
    }
    if (node['implements'] === undefined) {
      (node as Record<string, unknown>)['implements'] = [];
      stats.transformsApplied++;
    }
  }
}
