import { describe, test, expect } from '@rstest/core';
import { pathToFileURL } from 'node:url';
import path from 'node:path';

// eslint-plugin-import ships CJS; pull the namespace then unwrap default.
// eslint-disable-next-line @typescript-eslint/no-explicit-any
import * as importPluginNS from 'eslint-plugin-import';
import {
  runConformance,
  formatReport,
  type ConformanceFixture,
} from '../../src/eslint-conformance.js';

/**
 * Combined-rules conformance for `eslint-plugin-import`. Enables eight
 * pure-AST rules at once on every fixture so cross-rule interaction
 * matches ESLint v10 exactly.
 *
 * Rule set is the intersection of:
 *   - exists in eslint-plugin-import@2.32
 *   - empirically fires under ESLint v10 (probed before adding)
 *   - doesn't need a module resolver / settings (no-cycle, no-unresolved,
 *     order are out of scope for in-memory fixtures)
 *   - doesn't crash under v10 (no-default-export is excluded: upstream
 *     bug "Cannot use 'in' operator to search for 'sourceType' in
 *     undefined" when loaded against v10).
 *
 * Companion to `import-plugin.test.ts` (single rule per fixture).
 */

// eslint-disable-next-line @typescript-eslint/no-explicit-any
const importPlugin = ((importPluginNS as any).default ?? importPluginNS) as {
  rules: Record<string, unknown>;
  meta?: { name?: string; version?: string };
};

const RULE_NAMES = [
  'first',
  'newline-after-import',
  'no-duplicates',
  'no-mutable-exports',
  'no-webpack-loader-syntax',
  'no-amd',
  'no-commonjs',
  'no-anonymous-default-export',
  'no-named-default',
] as const;

const ALL_RULES = Object.fromEntries(
  RULE_NAMES.map((r) => [`import/${r}`, 'error'] as const),
) as ConformanceFixture['rules'];

describe('eslint-plugin-import combined-rules conformance', () => {
  test('eslint and rslint match when many import rules fire on the same source', async () => {
    const baseUrl = pathToFileURL(
      path.resolve(__dirname, '..', '..', 'package.json'),
    ).href;

    const fixtures: ConformanceFixture[] = [
      // Multi-rule fire on a single file:
      //   - L1, L2 → no-duplicates (same source twice)
      //   - L3    → first (import after a statement)
      //   - L4    → newline-after-import (no blank line before code below)
      {
        filePath: 'multi.js',
        text:
          `import a from './x';\n` +
          `import b from './x';\n` +
          `const sentinel = 1;\n` +
          `import c from './y';\nconst _ = [a,b,c,sentinel];\n`,
        rules: ALL_RULES,
      },
      // no-mutable-exports + no-anonymous-default-export interact in
      // the same file — both fire.
      {
        filePath: 'exports.js',
        text: `export let counter = 0;\n` + `export default function () {}\n`,
        rules: ALL_RULES,
      },
      // Loader-syntax + named-default in one file.
      {
        filePath: 'loaders.js',
        text:
          `import styles from '!style-loader!./x.css';\n` +
          `import { default as foo } from './y';\n` +
          `const _ = [styles, foo];\n`,
        rules: ALL_RULES,
      },
      // AMD + CommonJS (legacy module formats) — both fire.
      {
        filePath: 'legacy.js',
        text:
          `const x = require('./x');\n` +
          `define(['./y'], function (y) { return [x, y]; });\n`,
        rules: ALL_RULES,
      },
      // Clean ES module — nothing should fire.
      {
        filePath: 'clean.js',
        text:
          `import a from './a';\n` +
          `import b from './b';\n` +
          `\n` +
          `export function go() { return [a, b]; }\n`,
        rules: ALL_RULES,
      },
      // Mix: duplicate + newline missing + commonjs in the same file.
      {
        filePath: 'mixed.js',
        text:
          `import a from './a';\n` +
          `import a2 from './a';\n` +
          `const x = require('./x');\n` +
          `const _ = [a, a2, x];\n`,
        rules: ALL_RULES,
      },
    ];

    const report = await runConformance({
      plugin: {
        prefix: 'import',
        plugin: importPlugin as never,
        specifier: 'eslint-plugin-import',
        ruleNames: [...RULE_NAMES],
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
