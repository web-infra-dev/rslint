import { describe, expect, test } from 'rstack/test';
import {
  checkExpectedError,
  checkInvalidCase,
  checkValidCase,
} from '../src/cases.ts';

describe('checkValidCase', () => {
  test('accepts a case without expectations', () => {
    expect(() => {
      checkValidCase({ code: 'foo()' });
    }).not.toThrow();
  });

  test.each(['errors', 'output'])('rejects %s', (key) => {
    expect(() => {
      checkValidCase({ code: 'foo()', [key]: undefined });
    }).toThrow(`Valid test case must not have '${key}' property`);
  });
});

describe('checkExpectedError', () => {
  const error = {
    messageId: 'unexpected',
    message: 'Unexpected foo.',
    line: 1,
    column: 1,
    endLine: 1,
    endColumn: 4,
  };

  test('accepts a complete error', () => {
    expect(() => {
      checkExpectedError(error, 0);
    }).not.toThrow();
    expect(() => {
      checkExpectedError({ ...error, message: () => 'Unexpected foo.' }, 0);
    }).not.toThrow();
  });

  test('rejects a message string', () => {
    expect(() => {
      checkExpectedError('Unexpected foo.', 0);
    }).toThrow('errors[0] must be an object');
  });

  test.each(Object.keys(error))('rejects a missing %s', (key) => {
    const rest = Object.fromEntries(
      Object.entries(error).filter(([name]) => name !== key),
    );
    expect(() => {
      checkExpectedError(rest, 2);
    }).toThrow(`errors[2].${key} must be`);
  });
});

describe('checkInvalidCase', () => {
  test('accepts a case with errors', () => {
    expect(() => {
      checkInvalidCase({ errors: [{}] });
    }).not.toThrow();
  });

  test('rejects missing errors', () => {
    expect(() => {
      checkInvalidCase({});
    }).toThrow("must have an 'errors' array");
  });

  test('rejects empty errors', () => {
    expect(() => {
      checkInvalidCase({ errors: [] });
    }).toThrow('at least one error');
  });
});
