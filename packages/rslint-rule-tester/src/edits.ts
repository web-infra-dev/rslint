import assert from 'node:assert/strict';
import type { Diagnostic } from '@rslint/core/internal';

export type Edit = NonNullable<Diagnostic['fixes']>[number];

const BOM = '﻿';

/**
 * Applies edits as a single change. Offsets are UTF-16 and exclude a leading
 * byte order mark; an edit starting at -1 also removes the mark.
 */
export function applyEdits(code: string, edits: readonly Edit[]): string {
  const bom = code.startsWith(BOM) ? BOM : '';
  let text = bom ? code.slice(1) : code;
  let keepBom = true;
  const sorted = [...edits].sort(
    (a, b) => a.startPos - b.startPos || a.endPos - b.endPos,
  );
  for (let i = 1; i < sorted.length; i++) {
    assert.ok(
      sorted[i].startPos >= sorted[i - 1].endPos,
      `Edits overlap: ${JSON.stringify(sorted)}`,
    );
  }
  for (const edit of sorted.reverse()) {
    const start = Math.max(0, edit.startPos);
    if (
      edit.startPos < 0 ||
      (edit.startPos === 0 && edit.text.startsWith(BOM))
    ) {
      keepBom = false;
    }
    text = text.slice(0, start) + edit.text + text.slice(edit.endPos);
  }
  return (keepBom ? bom : '') + text;
}
