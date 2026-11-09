import assert from 'node:assert/strict';
import type { Diagnostic } from '@rslint/core/internal';

export interface ExpectedSuggestion {
  messageId: string;
  /** The rendered suggestion message. */
  desc: string;
  /** The source after applying only this suggestion. */
  output: string;
}

type Edit = NonNullable<Diagnostic['fixes']>[number];

const BOM = '\uFEFF';

/**
 * Applies one diagnostic's edits as a single change, like ESLint applies a
 * suggestion. Offsets are UTF-16 and exclude a leading byte order mark; an
 * edit starting at -1 also removes the mark.
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
      `Suggestion edits overlap: ${JSON.stringify(sorted)}`,
    );
  }
  for (const edit of sorted.reverse()) {
    const start = Math.max(0, edit.startPos);
    if (edit.startPos < 0) keepBom = false;
    text = text.slice(0, start) + edit.text + text.slice(edit.endPos);
  }
  return (keepBom ? bom : '') + text;
}

/**
 * Checks a diagnostic's suggestions against the expectation. Omitting the
 * expectation, or passing an empty array, requires that there are none.
 */
export function checkSuggestions(
  code: string,
  diagnostic: Pick<Diagnostic, 'message' | 'suggestions'>,
  expected: readonly ExpectedSuggestion[] | undefined,
): void {
  const actual = diagnostic.suggestions ?? [];
  const context = `on error with message: "${diagnostic.message}"`;
  if (!expected?.length) {
    assert.equal(
      actual.length,
      0,
      `Error should have no suggestions ${context}, but has ${JSON.stringify(actual)}. Specify them in 'suggestions'.`,
    );
    return;
  }
  assert.equal(
    actual.length,
    expected.length,
    `Error should have ${expected.length} suggestions ${context}, but has ${actual.length}`,
  );
  for (const [index, want] of expected.entries()) {
    const unknown = Object.keys(want).filter(
      (key) => !['messageId', 'desc', 'output'].includes(key),
    );
    assert.deepEqual(
      unknown,
      [],
      `Invalid property in expected suggestion ${index}. Expected messageId, desc and output.`,
    );
  }
  const seen = new Set<string>();
  for (const [index, suggestion] of actual.entries()) {
    const want: ExpectedSuggestion = expected[index];
    const at = `suggestion ${index} ${context}`;
    assert.ok(
      !seen.has(suggestion.message),
      `Suggestion message "${suggestion.message}" repeats within one error (${at})`,
    );
    seen.add(suggestion.message);
    assert.equal(suggestion.messageId, want.messageId, `messageId of ${at}`);
    assert.equal(suggestion.message, want.desc, `desc of ${at}`);
    assert.ok(suggestion.fixes?.length, `${at} has no edits`);
    const output = applyEdits(code, suggestion.fixes);
    assert.notEqual(output, code, `Output of ${at} should differ from code`);
    assert.equal(output, want.output, `Output of ${at}`);
  }
}
