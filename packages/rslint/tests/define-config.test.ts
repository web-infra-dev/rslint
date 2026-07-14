import { describe, test, expect } from '@rstest/core';
import { globalIgnores } from '../src/config/define-config.js';
import type { RuleEntry, RulesRecord } from '../src/config/define-config.js';

describe('globalIgnores', () => {
  test('returns a config entry containing only the ignores', () => {
    expect(globalIgnores(['dist/**', 'node_modules/**'])).toEqual({
      ignores: ['dist/**', 'node_modules/**'],
    });
  });

  test('produces an entry with no keys other than `ignores`', () => {
    // The global-ignore semantics rely on the entry having ONLY `ignores`
    // (no files/rules/plugins/settings/languageOptions). Lock that in so the
    // go-side `isGlobalIgnoreEntry` keeps recognizing it as a global ignore.
    expect(Object.keys(globalIgnores(['dist/**']))).toEqual(['ignores']);
  });

  test('returns the same patterns array reference', () => {
    const patterns = ['dist/**'];
    expect(globalIgnores(patterns).ignores).toBe(patterns);
  });

  test('throws TypeError when patterns is not an array', () => {
    // @ts-expect-error testing the runtime guard against non-array input
    expect(() => globalIgnores('dist/**')).toThrow(TypeError);
    // @ts-expect-error testing the runtime guard against non-array input
    expect(() => globalIgnores('dist/**')).toThrow(
      'ignorePatterns must be an array',
    );
  });

  test('throws TypeError when patterns is empty', () => {
    expect(() => globalIgnores([])).toThrow(TypeError);
    expect(() => globalIgnores([])).toThrow(
      'ignorePatterns must contain at least one pattern',
    );
  });
});

describe('RuleEntry / RulesRecord typing', () => {
  // Compile-time only — checked by `pnpm typecheck`, not by rstest's
  // (type-stripping) test runner. Locks in that `RuleEntry<Options>`
  // distributes a union-of-tuples `Options` over the severity + rest-args
  // shape, matching a rule-options generator's output
  // (scripts/generate-rule-option-types.mjs).
  test('RuleEntry<Options> distributes options over severities', () => {
    type Options = [] | [{ allow?: string[] }];
    const bareSeverity: RuleEntry<Options> = 'error';
    const severityOnly: RuleEntry<Options> = ['warn'];
    const withOptions: RuleEntry<Options> = ['error', { allow: ['log'] }];
    // @ts-expect-error `nope` isn't a key of Options' option object
    const wrongShape: RuleEntry<Options> = ['error', { nope: true }];
    // @ts-expect-error Options has no third tuple slot
    const tooManyArgs: RuleEntry<Options> = [
      'error',
      { allow: ['log'] },
      'extra',
    ];

    expect(bareSeverity).toBe('error');
    expect(severityOnly).toEqual(['warn']);
    expect(withOptions).toEqual(['error', { allow: ['log'] }]);
    expect(wrongShape).toEqual(['error', { nope: true }]);
    expect(tooManyArgs).toEqual(['error', { allow: ['log'] }, 'extra']);
  });

  test('RulesRecord still accepts arbitrary community rule keys via its fallback index signature', () => {
    const rules: RulesRecord = {
      'some-community-plugin/some-rule': ['error', 1, 2, 3],
      'another-rule': 'warn',
    };
    expect(rules['some-community-plugin/some-rule']).toEqual([
      'error',
      1,
      2,
      3,
    ]);
  });
});
