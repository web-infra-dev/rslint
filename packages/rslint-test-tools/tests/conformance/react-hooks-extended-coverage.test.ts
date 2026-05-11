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
 * Extended react-hooks conformance — beyond the two long-standing
 * "consumer" rules (rules-of-hooks, exhaustive-deps), v7 ships a
 * batch of React-Compiler-aligned rules. Most of them require a full
 * React Compiler runtime / React 19 component context to fire, but
 * a few do pure AST analysis on hook-shaped code and can be exercised
 * with the single-file harness.
 *
 * We test the ones that fire on plain ESM without compiler metadata:
 *
 *   - static-components: detects components shaped without hook usage,
 *     emits a "static component" hint based on JSX + Capitalized name.
 *   - capitalized-calls: enforces PascalCase / lowercase based on
 *     whether the call is a component vs hook.
 *   - purity: flags side effects in render bodies.
 *
 * Each runs as ESLint v10 native vs rslint worker, byte-diff. If a
 * rule turns out to silently no-op in this harness's environment
 * (because it requires React Compiler), conformance still passes
 * (both sides report 0 — vacuously aligned). What matters for the
 * regression net is that NEITHER side falsely fires.
 */

describe('react-hooks extended-coverage conformance', () => {
  test('eslint and rslint agree on react-hooks v7 in-file rules', async () => {
    const baseUrl = pathToFileURL(
      path.resolve(__dirname, '..', '..', 'package.json'),
    ).href;

    const fixtures: ConformanceFixture[] = [
      // Component with a setState call directly in render body.
      // Some v7 builds have `set-state-in-render` flagged at this site.
      {
        filePath: 'set-in-render.jsx',
        text: `function App() {
  const [x, setX] = React.useState(0);
  setX(1);
  return null;
}
`,
        rules: { 'react-hooks/set-state-in-render': 'error' },
      },
      // Capitalized function called as a hook (anti-pattern):
      {
        filePath: 'capitalized.jsx',
        text: `function App() {
  const v = Use(1);
  return v;
}
`,
        rules: { 'react-hooks/capitalized-calls': 'error' },
      },
      // Static (hook-free) component shape:
      {
        filePath: 'static-comp.jsx',
        text: `function Static() {
  return 1;
}
`,
        rules: { 'react-hooks/static-components': 'error' },
      },
      // Side effect inside component body (mutating a global) — purity.
      {
        filePath: 'purity.jsx',
        text: `let counter = 0;
function App() {
  counter++;
  return null;
}
`,
        rules: { 'react-hooks/purity': 'error' },
      },
    ];

    const report = await runConformance({
      plugin: {
        prefix: 'react-hooks',
        plugin: reactHooksPlugin as never,
        specifier: 'eslint-plugin-react-hooks',
        ruleNames: [
          'set-state-in-render',
          'capitalized-calls',
          'static-components',
          'purity',
        ],
      },
      fixtures,
      resolverBaseUrl: baseUrl,
      workerCount: 1,
    });

    if (report.mismatched > 0) {
      throw new Error('Conformance mismatch:\n' + formatReport(report));
    }
    expect(report.mismatched).toBe(0);

    // Vacuous-pass guard. Previously the only assertion was that
    // mismatch===0 — but a rule that silently no-ops on BOTH sides
    // (because the React Compiler runtime / React 19 component context
    // isn't wired) also reports 0===0 and "matches". That made
    // "coverage advanced" claims unfalsifiable. Require at least one
    // fixture to have produced ≥1 ESLint diagnostic, so the suite
    // verifies at least SOME rule wiring actually fired.
    const totalEslintDiags = report.fixtureResults.reduce(
      (n, r) => n + r.eslint.length,
      0,
    );
    expect(totalEslintDiags).toBeGreaterThan(0);
  });
});
