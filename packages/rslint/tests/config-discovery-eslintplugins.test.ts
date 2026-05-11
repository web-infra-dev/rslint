import { describe, test, expect } from '@rstest/core';
import { isGlobalIgnoreEntry } from '../src/utils/config-discovery.js';

// #6 regression: a config entry that declares `eslintPlugins` is real
// plugin config, NOT a bare global-ignore. The predicate previously
// omitted `eslintPlugins`, so such an entry was misclassified and its
// `ignores` were hoisted to global, silently pruning nested configs.
describe('isGlobalIgnoreEntry — eslintPlugins (#6)', () => {
  test('bare ignores-only entry IS a global ignore', () => {
    expect(isGlobalIgnoreEntry({ ignores: ['**/dist'] } as never)).toBe(true);
  });

  test('entry with eslintPlugins is NOT a global ignore', () => {
    expect(
      isGlobalIgnoreEntry({
        ignores: ['**/dist'],
        eslintPlugins: { uc: {} },
      } as never),
    ).toBe(false);
  });

  test('content fields still disqualify a global ignore (unchanged)', () => {
    expect(isGlobalIgnoreEntry({ ignores: ['x'], rules: {} } as never)).toBe(
      false,
    );
    expect(
      isGlobalIgnoreEntry({ ignores: ['x'], files: ['*.ts'] } as never),
    ).toBe(false);
    expect(isGlobalIgnoreEntry({ ignores: ['x'], plugins: {} } as never)).toBe(
      false,
    );
  });

  test('empty / missing ignores is never a global ignore', () => {
    expect(isGlobalIgnoreEntry({ ignores: [] } as never)).toBe(false);
    expect(isGlobalIgnoreEntry({} as never)).toBe(false);
  });
});
