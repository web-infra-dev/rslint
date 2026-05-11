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
 * Combined-rules conformance for `eslint-plugin-react-hooks`. Enables
 * BOTH `rules-of-hooks` and `exhaustive-deps` at once so the harness
 * verifies that the two rules running in the same pass produce
 * diagnostics identical to ESLint v10 — including the negative case
 * where neither fires.
 *
 * `exhaustive-deps` was previously excluded from `react-hooks-plugin.test.ts`
 * pending scope-analysis improvements; this file proves it now works
 * end-to-end against ESLint v10 with rslint's current scope manager.
 */

// eslint-disable-next-line @typescript-eslint/no-explicit-any
const reactHooks = ((reactHooksNS as any).default ?? reactHooksNS) as {
  rules: Record<string, unknown>;
  meta?: { name?: string; version?: string };
};

const ALL_RULES = {
  'react-hooks/rules-of-hooks': 'error',
  'react-hooks/exhaustive-deps': 'error',
} as const satisfies ConformanceFixture['rules'];

describe('eslint-plugin-react-hooks combined-rules conformance', () => {
  test('eslint and rslint agree across rules-of-hooks + exhaustive-deps', async () => {
    const baseUrl = pathToFileURL(
      path.resolve(__dirname, '..', '..', 'package.json'),
    ).href;

    const fixtures: ConformanceFixture[] = [
      // rules-of-hooks fires (hook inside `if`).
      // exhaustive-deps doesn't fire here (no useEffect with empty deps).
      {
        filePath: 'conditional.jsx',
        text: [
          'function Comp({flag}) {',
          '  if (flag) {',
          '    const [x, setX] = useState(0);',
          '    return x;',
          '  }',
          '  return null;',
          '}',
        ].join('\n'),
        rules: ALL_RULES,
      },
      // exhaustive-deps fires (missing dep `id` in the deps array).
      // rules-of-hooks doesn't fire.
      {
        filePath: 'missing-deps.jsx',
        text: [
          'function Comp({id}) {',
          '  useEffect(() => { console.log(id); }, []);',
          '  return null;',
          '}',
        ].join('\n'),
        rules: ALL_RULES,
      },
      // Both fire on the same file:
      //   - rules-of-hooks: hook inside `for`
      //   - exhaustive-deps: missing dep in a separate useEffect
      {
        filePath: 'both.jsx',
        text: [
          'function Comp({items, id}) {',
          '  useEffect(() => { console.log(id); }, []);',
          '  for (const i of items) {',
          '    useEffect(() => {});',
          '  }',
          '  return null;',
          '}',
        ].join('\n'),
        rules: ALL_RULES,
      },
      // Clean: hooks at top-level, deps correctly listed.
      {
        filePath: 'clean.jsx',
        text: [
          'function Comp({id}) {',
          '  const [x, setX] = useState(0);',
          '  useEffect(() => { console.log(id, x); }, [id, x]);',
          '  return x;',
          '}',
        ].join('\n'),
        rules: ALL_RULES,
      },
    ];

    const report = await runConformance({
      plugin: {
        prefix: 'react-hooks',
        plugin: reactHooks as never,
        specifier: 'eslint-plugin-react-hooks',
        ruleNames: ['rules-of-hooks', 'exhaustive-deps'],
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
