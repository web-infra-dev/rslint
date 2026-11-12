import type { ExpectedError } from '@rslint/rule-tester';

// The caller supplies the exact expected reporting span from its source template.
// JavaScript string lengths are UTF-16 code units, matching diagnostic columns.
export function singleLineError(
  messageId: string,
  message: string,
  reportedText: string,
  column = 1,
): ExpectedError {
  return {
    messageId,
    message,
    line: 1,
    column,
    endLine: 1,
    endColumn: column + reportedText.length,
  };
}
