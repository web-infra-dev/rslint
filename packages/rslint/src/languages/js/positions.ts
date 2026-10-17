/** Source positions shared by frame decoding, SourceCode, and diagnostics. */
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
 * `source-code.ts`'s `getLocFromIndex` and the frame decoder both
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
 *   `.charCodeAt(i)`, `.slice(a, b)`). the frame also carries offsets
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

// ─────────────────────────────────────────────────────────────────────
// Line-start offset lookup — shared safe accessor for `lso[line - 1]`
// ─────────────────────────────────────────────────────────────────────

/**
 * Return the UTF-16 offset where `line1Based` starts. Clamps both ends so
 * out-of-range inputs (a rule reporting `loc.line` beyond EOF, line < 1)
 * resolve to a safe in-range offset instead of `undefined`.
 *
 *   line < 1                        → 0
 *   line > lso.length               → textLength
 *   otherwise                       → lso[line - 1]
 *
 * Centralising this avoids each call site re-deriving the clamp; a prior
 * inline `lso[Math.max(0, line - 1)]` only clamped the low end and let
 * NaN/undefined propagate into diagnostic ranges.
 */
export function lineStartOffset(
  lso: readonly number[],
  line1Based: number,
  textLength: number,
): number {
  if (line1Based <= 0) return 0;
  if (line1Based - 1 >= lso.length) return textLength;
  return lso[line1Based - 1];
}
