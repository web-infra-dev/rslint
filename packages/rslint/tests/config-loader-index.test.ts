import { describe, test, expect } from '@rstest/core';
import { normalizeConfig } from '../src/config-loader.js';

// #14 regression: the validation `.map` ran AFTER `.filter`, so its
// `index` was filter-relative. With a non-object entry skipped earlier,
// error messages under-reported the offending entry's ORIGINAL position.
describe('normalizeConfig error index (#14)', () => {
  test('error cites the ORIGINAL index, not the post-filter index', () => {
    // Entry 0 (null) is skipped by the filter; entry 1 has an invalid
    // `files`. The thrown error must say "index 1" (original), not
    // "index 0" (which is what the post-filter index would report).
    expect(() => normalizeConfig([null, { files: 'not-an-array' }])).toThrow(
      /index 1/,
    );
  });

  test('valid config still normalizes (no behavior regression)', () => {
    const out = normalizeConfig([{ rules: { 'no-x': 'error' } }]);
    expect(Array.isArray(out)).toBe(true);
    expect(out).toHaveLength(1);
  });
});
