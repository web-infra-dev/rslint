import { describe, test, expect } from '@rstest/core';
import { pathToFileURL } from 'node:url';
import path from 'node:path';

// eslint-plugin-import ships CJS, no proper ESM default export — pull
// the namespace and read `.default` ourselves.
// eslint-disable-next-line @typescript-eslint/no-explicit-any
import * as importPluginNS from 'eslint-plugin-import';
import {
  runConformance,
  formatReport,
  type ConformanceFixture,
} from '../../src/eslint-conformance.js';

/**
 * Conformance harness against `eslint-plugin-import`.
 *
 * This is one of the most-used ESLint plugins in real codebases
 * (millions of weekly downloads), so getting it onto rslint's
 * compatibility matrix is high-value. We pick a representative slice
 * of rules — NOT the whole rule set, since the plugin includes
 * type-aware / project-aware rules that fall outside rslint's
 * documented support surface (see the `@experimental` JSDoc on
 * `RslintConfigEntry.eslintPlugins` in `packages/rslint/src/define-config.ts`).
 *
 * Rules included:
 *   - `import/no-self-import` — pure AST: a module that imports its
 *     own path. No project / resolver dependence.
 *   - `import/first` — every import must appear before other code.
 *     Purely structural.
 *
 * Deliberately excluded (future work / known-limitation):
 *   - `import/no-cycle` / `import/no-unresolved` — need real module
 *     resolution against the filesystem and the user's `tsconfig`
 *     paths, which the conformance harness's in-memory fixtures don't
 *     model.
 *   - `import/order` — depends on settings + import resolver state.
 *     Adding it requires wiring the harness's per-fixture settings.
 */

// eslint-disable-next-line @typescript-eslint/no-explicit-any
const importPlugin = ((importPluginNS as any).default ?? importPluginNS) as {
  rules: Record<string, unknown>;
  meta?: { name?: string; version?: string };
};

describe('eslint-plugin-import conformance', () => {
  test('eslint and rslint match on a representative slice of import rules', async () => {
    const baseUrl = pathToFileURL(
      path.resolve(__dirname, '..', '..', 'package.json'),
    ).href;

    const fixtures: ConformanceFixture[] = [
      // import/no-self-import — file imports its own basename.
      // (eslint-plugin-import uses the import specifier itself; in
      // sandbox fixtures with no real fs, we use a synthetic spec.)
      {
        filePath: 'self.js',
        text: `import x from './self';\nconst _ = x;`,
        rules: { 'import/no-self-import': 'error' },
      },
      // Clean file: no diagnostic expected.
      {
        filePath: 'clean.js',
        text: `import x from './other';\nconst _ = x;`,
        rules: { 'import/no-self-import': 'error' },
      },
      // import/first — import after a statement violates the rule.
      {
        filePath: 'first-bad.js',
        text: `const x = 1;\nimport y from './y';\nconst _ = y;`,
        rules: { 'import/first': 'error' },
      },
      // import/first — clean: import at top.
      {
        filePath: 'first-good.js',
        text: `import y from './y';\nconst x = 1;\nconst _ = y;`,
        rules: { 'import/first': 'error' },
      },
    ];

    const report = await runConformance({
      plugin: {
        prefix: 'import',
        plugin: importPlugin as never,
        specifier: 'eslint-plugin-import',
        ruleNames: ['no-self-import', 'first'],
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
