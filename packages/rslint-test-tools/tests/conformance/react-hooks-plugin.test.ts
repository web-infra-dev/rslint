import { describe, test, expect } from '@rstest/core';
import { pathToFileURL } from 'node:url';
import path from 'node:path';

// eslint-disable-next-line @typescript-eslint/no-explicit-any
import * as reactHooksNS from 'eslint-plugin-react-hooks';
import {
  runConformance,
  formatReport,
  type ConformanceFixture,
} from '../../src/eslint-conformance.js';

/**
 * Conformance harness against `eslint-plugin-react-hooks`.
 *
 * Hook plumbing rules are some of the most-installed in the React
 * ecosystem, so they're a meaningful conformance target. Coverage here
 * is intentionally focused on the high-confidence subset:
 *
 *   - `react-hooks/rules-of-hooks` — pure AST rule: detects hook calls
 *     in conditional / loop / nested-function positions. Doesn't need
 *     type information.
 *
 * Deliberately excluded:
 *
 *   - `react-hooks/exhaustive-deps` — needs `parserServices` /
 *     scope-analysis subtleties around effect dependencies that fall
 *     outside the documented experimental compatibility surface (see
 *     the `@experimental` JSDoc on
 *     `RslintConfigEntry.eslintPlugins` in `packages/rslint/src/define-config.ts`).
 *     When rslint's scope handling catches up, this fixture set can
 *     extend.
 */

// eslint-disable-next-line @typescript-eslint/no-explicit-any
const reactHooks = ((reactHooksNS as any).default ?? reactHooksNS) as {
  rules: Record<string, unknown>;
  meta?: { name?: string; version?: string };
};

describe('eslint-plugin-react-hooks conformance', () => {
  test('eslint and rslint match on rules-of-hooks fixtures', async () => {
    const baseUrl = pathToFileURL(
      path.resolve(__dirname, '..', '..', 'package.json'),
    ).href;

    const fixtures: ConformanceFixture[] = [
      // Violation: hook inside an if branch.
      {
        filePath: 'conditional-hook.jsx',
        text: [
          'function Comp({flag}) {',
          '  if (flag) {',
          '    const [x, setX] = useState(0);',
          '    return x;',
          '  }',
          '  return null;',
          '}',
        ].join('\n'),
        rules: { 'react-hooks/rules-of-hooks': 'error' },
      },
      // Clean: hooks at top-level only.
      {
        filePath: 'clean-hook.jsx',
        text: [
          'function Comp() {',
          '  const [x, setX] = useState(0);',
          '  return x;',
          '}',
        ].join('\n'),
        rules: { 'react-hooks/rules-of-hooks': 'error' },
      },
      // Violation: hook inside a loop.
      {
        filePath: 'loop-hook.jsx',
        text: [
          'function Comp({items}) {',
          '  for (const i of items) {',
          '    useEffect(() => {});',
          '  }',
          '  return null;',
          '}',
        ].join('\n'),
        rules: { 'react-hooks/rules-of-hooks': 'error' },
      },
    ];

    const report = await runConformance({
      plugin: {
        prefix: 'react-hooks',
        plugin: reactHooks as never,
        specifier: 'eslint-plugin-react-hooks',
        ruleNames: ['rules-of-hooks'],
      },
      fixtures,
      resolverBaseUrl: baseUrl,
      workerCount: 1,
    });

    if (report.mismatched > 0) {
      throw new Error(`conformance mismatch:\n${formatReport(report)}`);
    }
    expect(report.mismatched).toBe(0);
    expect(report.matched).toBe(fixtures.length);
  });
});
