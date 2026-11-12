import { describe, expect, test } from 'rstack/test';
import type { Diagnostic } from '@rslint/core/internal';
import {
  checkDiagnostic,
  checkOutput,
  checkResult,
  type CheckContext,
} from '../src/diagnostics.ts';
import type { ExpectedError } from '../src/types.ts';

const context: CheckContext = {
  code: 'var a = 1;',
  ruleName: 'node/no-var',
  root: '/fixture',
  hasOutput: false,
};
const diagnostic: Diagnostic = {
  ruleName: 'node/no-var',
  message: 'Unexpected var.',
  messageId: 'unexpectedVar',
  filePath: 'input.js',
  range: { start: { line: 1, column: 1 }, end: { line: 1, column: 4 } },
};
const expected: ExpectedError = {
  messageId: 'unexpectedVar',
  message: 'Unexpected var.',
  line: 1,
  column: 1,
  endLine: 1,
  endColumn: 4,
};
const fix = { startPos: 0, endPos: 3, text: 'let' };

// Defers the check so the assertions can observe a throw.
function check(...args: Parameters<typeof checkDiagnostic>) {
  return () => {
    checkDiagnostic(...args);
  };
}

describe('checkResult', () => {
  const result = { fileCount: 1, ruleCount: 1, diagnostics: [diagnostic] };

  test('accepts matching diagnostics', () => {
    expect(() => {
      checkResult(result, [expected], context);
    }).not.toThrow();
  });

  test.each([
    ['fileCount', { ...result, fileCount: 0 }, 'one linted file'],
    ['ruleCount', { ...result, ruleCount: 0 }, 'one executed rule'],
    ['diagnostic count', { ...result, diagnostics: [] }, 'Should have 1'],
  ])('rejects a different %s', (_, actual, message) => {
    expect(() => {
      checkResult(actual, [expected], context);
    }).toThrow(message);
  });
});

describe('checkDiagnostic', () => {
  test('accepts a matching diagnostic', () => {
    expect(check(diagnostic, expected, 0, context)).not.toThrow();
  });

  test('passes the fixture root to a message function', () => {
    expect(
      check(
        { ...diagnostic, message: 'Missing /fixture/a.js.' },
        { ...expected, message: (root) => `Missing ${root}/a.js.` },
        0,
        context,
      ),
    ).not.toThrow();
  });

  test.each([
    ['ruleName', { ...diagnostic, ruleName: 'node/other' }, 'rule name'],
    ['message', { ...diagnostic, message: 'Other.' }, 'errors[0].message'],
    ['messageId', { ...diagnostic, messageId: 'other' }, '.messageId'],
    [
      'line',
      {
        ...diagnostic,
        range: { ...diagnostic.range, start: { line: 2, column: 1 } },
      },
      '.line',
    ],
    [
      'column',
      {
        ...diagnostic,
        range: { ...diagnostic.range, start: { line: 1, column: 2 } },
      },
      '.column',
    ],
    [
      'endLine',
      {
        ...diagnostic,
        range: { ...diagnostic.range, end: { line: 2, column: 4 } },
      },
      '.endLine',
    ],
    [
      'endColumn',
      {
        ...diagnostic,
        range: { ...diagnostic.range, end: { line: 1, column: 5 } },
      },
      '.endColumn',
    ],
  ])('rejects a different %s', (_, actual, message) => {
    expect(check(actual, expected, 0, context)).toThrow(message);
  });

  test('rejects an autofix when the case has no output', () => {
    expect(
      check({ ...diagnostic, fixes: [fix] }, expected, 0, context),
    ).toThrow("has an autofix, but the case has no 'output'");
  });

  test('allows an autofix when the case has output', () => {
    expect(
      check({ ...diagnostic, fixes: [fix] }, expected, 0, {
        ...context,
        hasOutput: true,
      }),
    ).not.toThrow();
  });

  test('compares an exact fix', () => {
    const exact: ExpectedError = {
      ...expected,
      fix: { range: [0, 3], text: 'let' },
    };
    expect(
      check({ ...diagnostic, fixes: [fix] }, exact, 0, context),
    ).not.toThrow();
    expect(
      check(
        { ...diagnostic, fixes: [fix, { startPos: 9, endPos: 10, text: '' }] },
        exact,
        0,
        context,
      ),
    ).toThrow('errors[0].fix');
  });

  test('checks suggestions', () => {
    expect(
      check(
        {
          ...diagnostic,
          suggestions: [
            { messageId: 'useLet', message: 'Use let.', fixes: [fix] },
          ],
        },
        expected,
        0,
        context,
      ),
    ).toThrow('should have no suggestions');
  });
});

describe('checkOutput', () => {
  test('accepts output for the tested file only', () => {
    expect(() => {
      checkOutput(
        { 'src/a.js': 'let a;' },
        '/fixture',
        '/fixture/src/a.js',
        'let a;',
      );
    }).not.toThrow();
  });

  test.each([
    ['no output', undefined],
    ['different text', { 'src/a.js': 'const a;' }],
    ['another file', { 'src/a.js': 'let a;', 'src/b.js': 'let b;' }],
  ])('rejects %s', (_, output) => {
    expect(() => {
      checkOutput(output, '/fixture', '/fixture/src/a.js', 'let a;');
    }).toThrow('Fixed output');
  });
});
