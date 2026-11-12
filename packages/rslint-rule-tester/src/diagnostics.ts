import assert from 'node:assert/strict';
import path from 'node:path';
import type { Diagnostic, LintResponse } from '@rslint/core/internal';
import { checkSuggestions } from './suggestions.ts';
import type { ExpectedError } from './types.ts';

export interface CheckContext {
  code: string;
  ruleName: string;
  /** The absolute fixture root. */
  root: string;
  /** Whether the case expects an autofix output. */
  hasOutput: boolean;
}

/**
 * Checks that the case linted one file with one rule, and that the
 * diagnostics match the expected errors in order.
 */
export function checkResult(
  result: Pick<LintResponse, 'fileCount' | 'ruleCount' | 'diagnostics'>,
  errors: readonly ExpectedError[],
  context: CheckContext,
): void {
  assert.equal(result.fileCount, 1, 'Expected exactly one linted file');
  assert.equal(result.ruleCount, 1, 'Expected exactly one executed rule');
  assert.equal(
    result.diagnostics.length,
    errors.length,
    `Should have ${errors.length} errors but had ${result.diagnostics.length}: ${JSON.stringify(result.diagnostics, null, 2)}`,
  );
  for (const [index, expected] of errors.entries()) {
    checkDiagnostic(result.diagnostics[index], expected, index, context);
  }
}

/** Checks one diagnostic against its expected error. */
export function checkDiagnostic(
  diagnostic: Diagnostic,
  expected: ExpectedError,
  index: number,
  context: CheckContext,
): void {
  const at = `errors[${index}]`;
  assert.equal(diagnostic.ruleName, context.ruleName, `${at} rule name`);
  assert.equal(diagnostic.severity, 'error', `${at}.severity`);
  const message =
    typeof expected.message === 'function'
      ? expected.message(context.root.replaceAll('\\', '/'))
      : expected.message;
  assert.equal(diagnostic.message, message, `${at}.message`);
  assert.equal(diagnostic.messageId, expected.messageId, `${at}.messageId`);
  const { start, end } = diagnostic.range;
  assert.equal(start.line, expected.line, `${at}.line`);
  assert.equal(start.column, expected.column, `${at}.column`);
  assert.equal(end.line, expected.endLine, `${at}.endLine`);
  assert.equal(end.column, expected.endColumn, `${at}.endColumn`);
  if (expected.fix) {
    assert.deepEqual(
      diagnostic.fixes,
      [
        {
          startPos: expected.fix.range[0],
          endPos: expected.fix.range[1],
          text: expected.fix.text,
        },
      ],
      `${at}.fix`,
    );
  } else if (!context.hasOutput) {
    assert.equal(
      diagnostic.fixes,
      undefined,
      `${at} has an autofix, but the case has no 'output'`,
    );
  }
  checkSuggestions(context.code, diagnostic, expected.suggestions);
}

/** Checks that autofixing changed only the tested file, into the output. */
export function checkOutput(
  output: LintResponse['output'],
  root: string,
  filename: string,
  expected: string,
): void {
  // Output is keyed by the path relative to the config directory.
  const key = path.relative(root, filename).replaceAll('\\', '/');
  assert.deepEqual(output, { [key]: expected }, 'Fixed output');
}
