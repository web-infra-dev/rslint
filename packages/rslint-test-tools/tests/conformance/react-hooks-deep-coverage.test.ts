import { describe, test, expect } from '@rstest/core';
import { pathToFileURL } from 'node:url';
import path from 'node:path';

import reactHooksPlugin from 'eslint-plugin-react-hooks';
import {
  runConformance,
  formatReport,
  type ConformanceFixture,
} from '../../src/eslint-conformance.js';

/**
 * Deep react-hooks coverage. Adds three further v7 rules that align
 * cleanly with rslint's per-file pipeline:
 *   - immutability: state mutation patterns
 *   - refs: hook ref usage shape
 *   - component-hook-factories: function-returning-component factories
 *
 * Coverage advances 6/29 → 9/29 (~31%). `error-boundaries` and
 * `syntax` were investigated but produced ESLint=1 / rslint=0 on
 * representative fixtures — they likely require React Compiler
 * metadata that the worker JS plugin path doesn't synthesise. Left
 * out of the conformance set rather than silenced; revisit when the
 * compiler-runtime integration story stabilises.
 */

describe('react-hooks deep-coverage conformance', () => {
  test('eslint and rslint agree on further react-hooks v7 rules', async () => {
    const baseUrl = pathToFileURL(
      path.resolve(__dirname, '..', '..', 'package.json'),
    ).href;

    const fixtures: ConformanceFixture[] = [
      {
        filePath: 'immutability.jsx',
        text: `const obj = { x: 0 };\nfunction Comp() {\n  obj.x++;\n  return obj.x;\n}\n`,
        rules: { 'react-hooks/immutability': 'error' },
      },
      {
        filePath: 'refs.jsx',
        text: `function Comp() {\n  const ref = useRef(null);\n  return ref.current;\n}\n`,
        rules: { 'react-hooks/refs': 'error' },
      },
      {
        filePath: 'component-hook-factories.jsx',
        text: `function makeComp() {\n  return function() { return null; };\n}\n`,
        rules: { 'react-hooks/component-hook-factories': 'error' },
      },
    ];

    const report = await runConformance({
      plugin: {
        prefix: 'react-hooks',
        plugin: reactHooksPlugin as never,
        specifier: 'eslint-plugin-react-hooks',
        ruleNames: ['immutability', 'refs', 'component-hook-factories'],
      },
      fixtures,
      resolverBaseUrl: baseUrl,
      workerCount: 1,
    });

    if (report.mismatched > 0) {
      throw new Error('Conformance mismatch:\n' + formatReport(report));
    }
    expect(report.mismatched).toBe(0);

    // Vacuous-pass guard — see react-hooks-extended-coverage.test.ts
    // for rationale. Without this, three rules that silently no-op
    // on both sides (because they need React Compiler runtime) would
    // report 0===0 and falsely advance the coverage claim.
    const totalEslintDiags = report.fixtureResults.reduce(
      (n, r) => n + r.eslint.length,
      0,
    );
    expect(totalEslintDiags).toBeGreaterThan(0);
  });
});
