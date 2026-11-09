import { describe, expect, test } from 'rstack/test';
import { checkValidCase } from '../src/cases.ts';

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
