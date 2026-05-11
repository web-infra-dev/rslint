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
 * Bulk import-plugin coverage — single-file rules that don't depend on
 * the resolver. Eslint-plugin-import v2 is v8-era and probes
 * `context.parserOptions` (undefined under ESLint v10 flat config) for
 * any rule gated on `isEsmModule(parserOptions)`; those are skipped
 * (no-default-export, prefer-default-export, group-exports,
 * no-named-export). Other in-file rules work fine.
 */

describe('eslint-plugin-import bulk-coverage conformance', () => {
  test('eslint and rslint agree on additional in-file import rules', async () => {
    const baseUrl = pathToFileURL(
      path.resolve(__dirname, '..', '..', 'package.json'),
    ).href;

    const fixtures: ConformanceFixture[] = [
      {
        filePath: 'imports-first.js',
        text: `import a from "./a";\nconst x = 1;\nimport b from "./b";\nvoid x; void a; void b;`,
        rules: { 'import/imports-first': 'error' },
      },
      {
        filePath: 'no-amd-non-define.js',
        text: `define("module-name", function () { return 1; });`,
        rules: { 'import/no-amd': 'error' },
      },
      {
        filePath: 'no-commonjs.js',
        text: `const a = require("./a");\nvoid a;`,
        rules: { 'import/no-commonjs': 'error' },
      },
      {
        filePath: 'no-namespace.js',
        text: `import * as a from "./mod";\nvoid a;`,
        rules: { 'import/no-namespace': 'error' },
      },
      {
        filePath: 'no-anonymous-default-2.js',
        text: `export default () => 1;`,
        rules: { 'import/no-anonymous-default-export': 'error' },
      },
      {
        filePath: 'no-empty-named-blocks.js',
        text: `import {} from "./mod";\n`,
        rules: { 'import/no-empty-named-blocks': 'error' },
      },
      {
        filePath: 'no-webpack-loader-syntax.js',
        text: `import a from "raw-loader!./a";\nvoid a;`,
        rules: { 'import/no-webpack-loader-syntax': 'error' },
      },
      {
        filePath: 'no-import-module-exports.js',
        text: `import x from "./mod";\nmodule.exports = x;`,
        rules: { 'import/no-import-module-exports': 'error' },
      },
    ];

    const ruleNames = Array.from(
      new Set(fixtures.flatMap((f) => Object.keys(f.rules))),
    )
      .map((r) => r.replace(/^import\//, ''))
      .sort();

    const report = await runConformance({
      plugin: {
        prefix: 'import',
        plugin: importPlugin as never,
        specifier: 'eslint-plugin-import',
        ruleNames,
      },
      fixtures,
      resolverBaseUrl: baseUrl,
      workerCount: 1,
    });

    if (report.mismatched > 0) {
      throw new Error('Conformance mismatch:\n' + formatReport(report));
    }
    expect(report.mismatched).toBe(0);
    // Vacuous-pass guard — see unicorn-extended-coverage for rationale.
    expect(
      report.fixtureResults.reduce((n, r) => n + r.eslint.length, 0),
    ).toBeGreaterThan(0);
  });
});
