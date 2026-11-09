import { describe, expect, test } from 'rstack/test';
import { applyEdits } from '../src/edits.ts';

const code = 'let a = 1;';

describe('applyEdits', () => {
  test('applies several edits as one change', () => {
    expect(
      applyEdits(code, [
        { startPos: 8, endPos: 9, text: '2' },
        { startPos: 0, endPos: 3, text: 'var' },
      ]),
    ).toBe('var a = 2;');
  });

  test('keeps a byte order mark outside the edit offsets', () => {
    expect(
      applyEdits('\uFEFFlet a;', [{ startPos: 0, endPos: 3, text: 'var' }]),
    ).toBe('\uFEFFvar a;');
  });

  test('removes the byte order mark for an edit starting at -1', () => {
    expect(
      applyEdits('\uFEFFlet a;', [{ startPos: -1, endPos: 0, text: '' }]),
    ).toBe('let a;');
  });

  test('rejects overlapping edits', () => {
    expect(() =>
      applyEdits(code, [
        { startPos: 0, endPos: 5, text: '' },
        { startPos: 4, endPos: 6, text: '' },
      ]),
    ).toThrow('overlap');
  });
});
