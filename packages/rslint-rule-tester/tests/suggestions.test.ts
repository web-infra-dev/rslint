import { describe, expect, test } from 'rstack/test';
import { checkSuggestions } from '../src/suggestions.ts';
import type { ExpectedSuggestion } from '../src/types.ts';

const code = 'let a = 1;';
const message = 'Unexpected let.';
const suggestion = {
  messageId: 'useConst',
  message: 'Use const.',
  fixes: [{ startPos: 0, endPos: 3, text: 'const' }],
};
const expected: ExpectedSuggestion = {
  messageId: 'useConst',
  desc: 'Use const.',
  output: 'const a = 1;',
};

// Defers the check so the assertions can observe a throw.
function check(...args: Parameters<typeof checkSuggestions>) {
  return () => {
    checkSuggestions(...args);
  };
}

describe('checkSuggestions', () => {
  test('accepts no suggestions when none are expected', () => {
    expect(check(code, { message }, undefined)).not.toThrow();
    expect(check(code, { message }, [])).not.toThrow();
  });

  test('rejects unexpected suggestions', () => {
    const diagnostic = { message, suggestions: [suggestion] };
    expect(check(code, diagnostic, undefined)).toThrow(
      'should have no suggestions',
    );
    expect(check(code, diagnostic, [])).toThrow('should have no suggestions');
  });

  test('accepts matching suggestions', () => {
    expect(
      check(code, { message, suggestions: [suggestion] }, [expected]),
    ).not.toThrow();
  });

  test('rejects a missing suggestion', () => {
    expect(check(code, { message }, [expected])).toThrow(
      'should have 1 suggestions',
    );
  });

  test.each([
    ['messageId', { ...expected, messageId: 'other' }],
    ['desc', { ...expected, desc: 'Other.' }],
    ['output', { ...expected, output: 'var a = 1;' }],
  ])('rejects a different %s', (field, want) => {
    expect(check(code, { message, suggestions: [suggestion] }, [want])).toThrow(
      field === 'output' ? 'Output of suggestion 0' : field,
    );
  });

  test('rejects an unknown expectation property', () => {
    expect(
      check(code, { message, suggestions: [suggestion] }, [
        { ...expected, data: {} } as ExpectedSuggestion,
      ]),
    ).toThrow('Invalid property');
  });

  test('rejects a suggestion without edits', () => {
    expect(
      check(code, { message, suggestions: [{ ...suggestion, fixes: [] }] }, [
        expected,
      ]),
    ).toThrow('has no edits');
  });

  test('rejects a suggestion that leaves the code unchanged', () => {
    expect(
      check(
        code,
        {
          message,
          suggestions: [
            { ...suggestion, fixes: [{ startPos: 0, endPos: 3, text: 'let' }] },
          ],
        },
        [{ ...expected, output: code }],
      ),
    ).toThrow('should differ from code');
  });

  test('rejects repeated suggestion messages', () => {
    expect(
      check(code, { message, suggestions: [suggestion, suggestion] }, [
        expected,
        expected,
      ]),
    ).toThrow('repeats');
  });
});
