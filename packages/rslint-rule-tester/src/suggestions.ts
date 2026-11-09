import assert from 'node:assert/strict';
import type { Diagnostic } from '@rslint/core/internal';
import { applyEdits } from './edits.ts';
import type { ExpectedSuggestion } from './types.ts';

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
