/**
 * Shared per-plugin conformance suite. Each `<plugin>.test.ts` calls
 * `runConformanceSuite` with its own curated cases; this lints every minimal
 * trigger through BOTH rslint's `eslintPlugins` feature (via the CLI) and
 * ESLint v10 (in-process) and asserts byte-identical diagnostics. Engine
 * fan-out is batched in `harness.ts`, so each plugin spawns rslint only a
 * handful of times.
 *
 * Intentionally EXCLUDED from every plugin's set — documented runner
 * limitations, not test gaps (adding them would assert a divergence):
 *   - Type-aware rules: the runner exposes no type checker
 *     (`parserServices.program` is undefined).
 *   - ESLint code-path analysis: not implemented, so rules built on
 *     `onCodePathStart` / `currentCodePathSegments` never fire (e.g. sonarjs
 *     `no-fallthrough` / `no-dead-store`, promise `always-return`).
 *   - Representation nuances: a bare Position `loc` report gets
 *     endLine/endColumn = null in ESLint but a zero-width end in rslint
 *     (stylistic `eol-last`); a TS member with no accessibility modifier is
 *     `null` (rslint) vs `undefined` (typescript-eslint), tripping unicorn
 *     `no-static-only-class`.
 */
import { describe, test, expect } from '@rstest/core';
import {
  compareCases,
  ALIASES,
  type DiffCase,
  type Verdict,
} from './harness.js';

/**
 * Register the conformance describe blocks for one plugin.
 *
 * @param pkg        plugin package, e.g. `eslint-plugin-unicorn`
 * @param cases      triggers that must report IDENTICALLY on both engines
 * @param cleanCases snippets that must report NOTHING on both engines
 */
export function runConformanceSuite(
  pkg: string,
  cases: DiffCase[],
  cleanCases: DiffCase[] = [],
): void {
  const alias = ALIASES[pkg];
  // Both engines run once per set, lazily on first test + memoized; the
  // per-test `--testTimeout=0` covers the one-off batch cost.
  let cv: Promise<Verdict[]> | undefined;
  const verdicts = () => (cv ??= compareCases(cases));
  let clv: Promise<Verdict[]> | undefined;
  const cleanVerdicts = () => (clv ??= compareCases(cleanCases));

  describe(`${pkg} — eslintPlugins ≡ ESLint v10`, () => {
    test('the curated set is non-empty', () => {
      expect(cases.length).toBeGreaterThan(0);
    });

    cases.forEach((c, i) => {
      test(`${alias}/${c.rule} — identical diagnostics`, async () => {
        const v = (await verdicts())[i];
        // The trigger must actually fire — otherwise two empty results would
        // "match" while testing nothing.
        expect(v.eslint.length).toBeGreaterThan(0);
        // rslint must reproduce ESLint v10 byte-for-byte (toEqual prints the
        // diff on divergence).
        expect(v.rslint).toEqual(v.eslint);
      });
    });
  });

  if (cleanCases.length) {
    describe(`${pkg} — no false positives on clean code`, () => {
      cleanCases.forEach((c, i) => {
        test(`${alias}/${c.rule} — reports nothing`, async () => {
          const v = (await cleanVerdicts())[i];
          expect(v.eslint).toEqual([]);
          expect(v.rslint).toEqual([]);
        });
      });
    });
  }
}
