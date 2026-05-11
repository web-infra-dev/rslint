import { describe, test, expect } from '@rstest/core';
import { pathToFileURL } from 'node:url';
import path from 'node:path';

import importPlugin from 'eslint-plugin-import';
import {
  runConformance,
  formatReport,
  type ConformanceFixture,
} from '../../src/eslint-conformance.js';

/**
 * Extended import-plugin conformance — covers in-file rules that
 * don't rely on the resolver (the conformance harness uses a single
 * synthetic file, so resolver-dependent rules like `no-unresolved`,
 * `no-cycle`, `extensions` would need multi-file fixtures we don't
 * have here).
 *
 * Coverage moves from 2/46 → 7/46 (~15%). Each rule below exercises
 * a distinct shape of AST/visitor pattern: default-export detection,
 * named-as-default, duplicate imports, newline placement, module
 * system bans.
 */

describe('eslint-plugin-import extended-coverage conformance', () => {
  test('eslint and rslint agree on in-file import rules', async () => {
    const baseUrl = pathToFileURL(
      path.resolve(__dirname, '..', '..', 'package.json'),
    ).href;

    const fixtures: ConformanceFixture[] = [
      // no-anonymous-default-export — `export default function() {}`
      // is anonymous; should be flagged.
      {
        filePath: 'no-anonymous.js',
        text: 'export default function() { return 1; }\n',
        rules: { 'import/no-anonymous-default-export': 'error' },
      },
      // no-mutable-exports — `export let` is mutable.
      {
        filePath: 'mutable.js',
        text: 'export let counter = 0;\n',
        rules: { 'import/no-mutable-exports': 'error' },
      },
      // no-duplicates — same module imported twice.
      {
        filePath: 'duplicates.js',
        text: `import { a } from "./mod";\nimport { b } from "./mod";\n`,
        rules: { 'import/no-duplicates': 'error' },
      },
      // newline-after-import — missing blank line between imports and code.
      {
        filePath: 'newline.js',
        text: `import x from "./x";\nconst y = x + 1;\n`,
        rules: { 'import/newline-after-import': 'error' },
      },
      // no-amd — `define(['m'], function(){})` is AMD style.
      {
        filePath: 'amd.js',
        text: `define(['m'], function (m) { return m; });\n`,
        rules: { 'import/no-amd': 'error' },
      },
      // Negative case: a clean file should produce zero diagnostics.
      {
        filePath: 'clean.js',
        text: `import x from "./x";\n\nexport const y = x;\n`,
        rules: {
          'import/no-anonymous-default-export': 'error',
          'import/no-mutable-exports': 'error',
          'import/no-duplicates': 'error',
          'import/newline-after-import': 'error',
          'import/no-amd': 'error',
        },
      },
    ];

    const report = await runConformance({
      plugin: {
        prefix: 'import',
        plugin: importPlugin as never,
        specifier: 'eslint-plugin-import',
        ruleNames: [
          'no-anonymous-default-export',
          'no-mutable-exports',
          'no-duplicates',
          'newline-after-import',
          'no-amd',
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
  });
});
