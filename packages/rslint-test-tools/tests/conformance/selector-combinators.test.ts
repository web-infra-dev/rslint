import { describe, test, expect } from '@rstest/core';
import { pathToFileURL } from 'node:url';
import path from 'node:path';
import { mkdirSync, writeFileSync, rmSync } from 'node:fs';

import {
  runConformance,
  formatReport,
  type ConformanceFixture,
} from '../../src/eslint-conformance.js';

/**
 * Regression pin for esquery-combinator selectors in the runner's
 * listener-dispatch path.
 *
 * Bug: the DFS visit pushed ancestors in root-first order and passed
 * the stack verbatim to `esquery.matches(node, sel, ancestry)`.
 * esquery expects `ancestry[0]` to be the IMMEDIATE PARENT (innermost
 * first); the stack was reversed. Result: every selector using `>`,
 * ` ` (descendant), `+`, or `~` silently never fired. Plugins that
 * write rules as `'FunctionDeclaration > BlockStatement > ReturnStatement'`,
 * `'CallExpression Identifier'`, etc. (a huge fraction of unicorn /
 * react-hooks selectors) were dead under the JS plugin path.
 *
 * Fix: reverse the ancestry stack before each call to `esquery.matches`.
 * This fixture exercises BOTH child- and descendant-combinator forms
 * via a synthetic plugin whose rule reports one diagnostic per match,
 * then byte-compares against ESLint v10.
 *
 * Why a synthetic plugin: the bug is at the **dispatch layer**, not in
 * any specific rule. A real unicorn/react-hooks rule with combinator
 * selectors would have surfaced the same bug, but a focused plugin
 * pins the dispatch behavior directly so a future regression at any
 * unrelated rule still trips this test.
 */

const TMP_DIR = '/tmp/_rslint_selector_combinators';

describe('listener dispatch — esquery combinator selectors conformance', () => {
  test('eslint and rslint dispatch `>` and descendant selectors identically', async () => {
    rmSync(TMP_DIR, { recursive: true, force: true });
    mkdirSync(TMP_DIR, { recursive: true });

    // Synthetic plugin: reports one diagnostic per selector match.
    // Same source on both engines (loaded via `specifier`).
    writeFileSync(
      `${TMP_DIR}/index.mjs`,
      `export default {
  meta: { name: 'sel-probe', version: '0.0.0' },
  rules: {
    probe: {
      meta: { messages: { x: '{{tag}}' } },
      create(context) {
        return {
          // child combinator
          'FunctionDeclaration > BlockStatement > ReturnStatement'(n) {
            context.report({ node: n, messageId: 'x', data: { tag: 'child:' + (n.argument?.type ?? 'void') } });
          },
          // descendant combinator
          'CallExpression Identifier'(n) {
            context.report({ node: n, messageId: 'x', data: { tag: 'desc:' + n.name } });
          },
        };
      },
    },
  },
};
`,
    );
    writeFileSync(
      `${TMP_DIR}/package.json`,
      JSON.stringify({ name: 'sel-probe', type: 'module', main: 'index.mjs' }),
    );

    // Re-require the module so it shows up under both ESLint flat-config
    // (`plugins: { sel: <obj> }`) and rslint (`specifier`).
    const pluginUrl = pathToFileURL(`${TMP_DIR}/index.mjs`).href;
    const { default: plugin } = await import(pluginUrl);

    const baseUrl = pathToFileURL(
      path.resolve(__dirname, '..', '..', 'package.json'),
    ).href;

    const fixtures: ConformanceFixture[] = [
      // Multiple ReturnStatements at different nesting; one bare call.
      {
        filePath: 'a.js',
        text:
          `function outer() {\n` +
          `  function inner() { return 1; }\n` +
          `  return inner();\n` +
          `}\n` +
          `outer();\n`,
        rules: { 'sel/probe': 'error' as const },
      },
      // Clean: no inner/outer functions; only one CallExpression.
      {
        filePath: 'b.js',
        text: `console.log(1);\n`,
        rules: { 'sel/probe': 'error' as const },
      },
      // Deep call chain — every Identifier under each CallExpression
      // should match the descendant selector.
      {
        filePath: 'c.js',
        text: `foo.bar(baz(qux), spam);\n`,
        rules: { 'sel/probe': 'error' as const },
      },
    ];

    try {
      const report = await runConformance({
        plugin: {
          prefix: 'sel',
          plugin: plugin as never,
          // Plugin runner's specifier resolver uses Node's `require.resolve`
          // semantics: bare-name OR path. file:// URLs are not accepted.
          specifier: `${TMP_DIR}/index.mjs`,
          ruleNames: ['probe'],
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
    } finally {
      rmSync(TMP_DIR, { recursive: true, force: true });
    }
  });
});
