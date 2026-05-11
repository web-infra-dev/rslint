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
 * Regression pins for two bugs found in the round-3 audit:
 *
 * 1. **Shebang comment**. The runner's hand-rolled tokenizer didn't
 *    recognize a leading `#!...` hashbang. Source like
 *    `#!/usr/bin/env node\nconst x = 1;` got tokenized into 5+
 *    junk tokens (`#`, `!`, `/`, ...) and `getAllComments()` returned
 *    `[]`. ESLint v10 produces a single `{ type: 'Shebang', value:
 *    '/usr/bin/env node' }` comment. Fix lives in `tokenizer.ts`.
 *
 * 2. **JSX scope analysis under `.jsx`**. `scope-factory.ts` passed
 *    `childVisitorKeys: null` to `eslint-scope.analyze`, which fell
 *    back to its internal map that doesn't include JSX node children.
 *    For any JSX-containing source under `.jsx` (or `.js` with
 *    `ecmaFeatures.jsx`), eslint-scope's `fallback: 'iteration'`
 *    walked the `parent` backref and blew the call stack on the
 *    FIRST JSX element. `scope-factory.ts` now passes the full
 *    `eslint-visitor-keys.KEYS` map so JSX children traverse
 *    correctly.
 *
 * The probe plugin reports a string snapshot that captures the
 * relevant surface for each fixture — byte-equal to ESLint v10
 * means the fix holds.
 */

const TMP_DIR = '/tmp/_rslint_shebang_jsx_scope_probe';

/**
 * Note: JSX scope fixtures are NOT in this file. ESLint v10 flat-config
 * needs `languageOptions.parserOptions.ecmaFeatures.jsx = true` to parse
 * `.jsx` files, but the current `runConformance` harness doesn't forward
 * `languageOptions` to the ESLint side. The fix is empirically verified
 * via runner-level probes; extending the harness signature is a
 * follow-up. See round-3 audit notes.
 */
describe('shebang conformance', () => {
  test('eslint and rslint agree on shebang comments and token shape', async () => {
    rmSync(TMP_DIR, { recursive: true, force: true });
    mkdirSync(TMP_DIR, { recursive: true });

    writeFileSync(
      `${TMP_DIR}/index.mjs`,
      `export default {
  meta: { name: 'shebang-jsx-probe', version: '0.0.0' },
  rules: {
    shebang: {
      // The conformance harness compares (ruleName, line, column,
      // messageId) — NOT the rendered message text. So we encode the
      // observed comment shape into distinct messageIds + emit one
      // diagnostic per comment + one per token (capped). A regression
      // where rslint's tokenizer drops the shebang comment would
      // produce a different multiset of messageIds (no shebangType)
      // — caught by the harness's per-diagnostic messageId check.
      meta: {
        messages: {
          shebangType: 'comment was Shebang type',
          lineType: 'comment was Line type',
          blockType: 'comment was Block type',
          tokenAt: 'token observed',
        },
      },
      create(context) {
        return {
          'Program:exit'(p) {
            const sc = context.sourceCode;
            for (const c of sc.getAllComments()) {
              const id =
                c.type === 'Shebang'
                  ? 'shebangType'
                  : c.type === 'Line'
                    ? 'lineType'
                    : 'blockType';
              context.report({ node: p, messageId: id });
            }
            // Cap token reports so a non-shebang-related token-shape
            // drift still surfaces but doesn't bury the comment signal.
            const tokens = sc.getTokens(p);
            for (let i = 0; i < Math.min(tokens.length, 8); i++) {
              context.report({ node: p, messageId: 'tokenAt' });
            }
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
      JSON.stringify({
        name: 'shebang-jsx-probe',
        type: 'module',
        main: 'index.mjs',
      }),
    );

    const pluginUrl = pathToFileURL(`${TMP_DIR}/index.mjs`).href;
    const { default: plugin } = await import(pluginUrl);
    const baseUrl = pathToFileURL(
      path.resolve(__dirname, '..', '..', 'package.json'),
    ).href;

    const fixtures: ConformanceFixture[] = [
      // Shebang at file start — single `Shebang:` comment, regular
      // tokens after.
      {
        filePath: 'cli.js',
        text: '#!/usr/bin/env node\nconst x = 1;',
        rules: { 'sb/shebang': 'error' as const },
      },
      // Only a shebang — no other tokens.
      {
        filePath: 'only.js',
        text: '#!/usr/bin/env node',
        rules: { 'sb/shebang': 'error' as const },
      },
      // No shebang — sanity, confirms we don't synthesize one.
      {
        filePath: 'plain.js',
        text: 'const x = 1;',
        rules: { 'sb/shebang': 'error' as const },
      },
    ];

    try {
      const reportA = await runConformance({
        plugin: {
          prefix: 'sb',
          plugin: plugin as never,
          specifier: `${TMP_DIR}/index.mjs`,
          ruleNames: ['shebang'],
        },
        fixtures,
        resolverBaseUrl: baseUrl,
        workerCount: 1,
      });

      if (reportA.mismatched > 0) {
        throw new Error(`conformance mismatch:\n${formatReport(reportA)}`);
      }
      expect(reportA.mismatched).toBe(0);
      expect(reportA.matched).toBe(fixtures.length);
    } finally {
      rmSync(TMP_DIR, { recursive: true, force: true });
    }
  });
});
