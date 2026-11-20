import type { ExpectedError, ExpectedSuggestion } from '@rslint/rule-tester';

/** Keep repeated assertions compact while stating every UTF-16 position. */
export function diagnostic(
  messageId: string,
  message: string,
  range: [line: number, column: number, endLine: number, endColumn: number],
  suggestions?: ExpectedSuggestion[],
): ExpectedError {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId,
    message,
    line,
    column,
    endLine,
    endColumn,
    ...(suggestions === undefined ? {} : { suggestions }),
  };
}
