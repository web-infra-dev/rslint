import { describe, test, expect } from '@rstest/core';
import path from 'node:path';

import { WorkerPool } from '../../src/eslint-plugin/worker-pool.js';

/**
 * WorkerPool end-to-end — multi-config (monorepo) dispatch: a single
 * worker holding several configs routes each task to its own
 * configKey-bound plugin set (same-module/different-prefix AND fully
 * disjoint plugin sets), an unknown configKey surfaces a parseError
 * instead of silently empty-rules, and a rule's fix() throw is
 * forwarded through onLog (LSP-visible).
 */

// Skipped on windows: tearing down a worker that has oxc (a napi addon)
// loaded aborts below the JS layer there (nodejs/node#34567) and crashes
// the rstest worker running this file. These e2e tests spawn real
// workers and tear them down, so they are windows-skipped; they still
// run on linux/macOS.
describe.skipIf(process.platform === 'win32')(
  'WorkerPool end-to-end with a local fixture plugin',
  () => {
    // Multi-config (monorepo) dispatch — a single worker holds TWO
    // configs under different directories with the SAME plugin module
    // under DIFFERENT prefixes (`pkgA` vs `pkgB`). Each task's `configKey`
    // selects which config's `LoadedPlugins` the worker uses to resolve
    // the rule. A bug that mixes them up would either return
    // "rule not found" for one prefix or silently route to the wrong
    // config's plugin instance — both visible as diagnostics that
    // never fire on the affected file.
    test('multi-config dispatch routes each file to its own configKey-bound plugins', async () => {
      const cfgADir = path.resolve(__dirname, 'fixtures', 'cfgA');
      const cfgBDir = path.resolve(__dirname, 'fixtures', 'cfgB');

      const pool = new WorkerPool({
        configs: [
          {
            configPath: path.join(cfgADir, 'rslint.config.mjs'),
            configDirectory: cfgADir,
          },
          {
            configPath: path.join(cfgBDir, 'rslint.config.mjs'),
            configDirectory: cfgBDir,
          },
        ],
        workerCount: 1, // single worker holds both configs, exercises the per-config map
        taskTimeoutMs: 10_000,
      });
      await pool.init();

      // Same file source, different configKey + rule prefix:
      //   - File under cfgA → rule `pkgA/no-null` MUST fire
      //   - File under cfgB → rule `pkgB/no-null` MUST fire
      const NULL_SRC = 'const x = null;';
      const results = await pool.lintBatch([
        {
          filePath: 'a.ts',
          text: NULL_SRC,
          rules: { 'pkgA/no-null': { options: [] } },
          collectFixes: false,
          suggestionsMode: 'off',
          configKey: cfgADir,
        },
        {
          filePath: 'b.ts',
          text: NULL_SRC,
          rules: { 'pkgB/no-null': { options: [] } },
          collectFixes: false,
          suggestionsMode: 'off',
          configKey: cfgBDir,
        },
      ]);

      expect(results).toHaveLength(2);
      // File a (configKey=cfgA): pkgA/no-null fires.
      expect(results[0].parseError).toBeUndefined();
      expect(results[0].diagnostics).toHaveLength(1);
      expect(results[0].diagnostics[0].ruleName).toBe('pkgA/no-null');
      // File b (configKey=cfgB): pkgB/no-null fires.
      expect(results[1].parseError).toBeUndefined();
      expect(results[1].diagnostics).toHaveLength(1);
      expect(results[1].diagnostics[0].ruleName).toBe('pkgB/no-null');

      // Cross-confusion regression: a file under cfgA asking for
      // `pkgB/no-null` must NOT resolve — cfgA's LoadedPlugins doesn't
      // hold pkgB. A worker bug that "falls back" to the union of all
      // configs would produce a hit here.
      const cross = await pool.lintBatch([
        {
          filePath: 'cross.ts',
          text: NULL_SRC,
          rules: { 'pkgB/no-null': { options: [] } },
          collectFixes: false,
          suggestionsMode: 'off',
          configKey: cfgADir,
        },
      ]);
      expect(cross[0].parseError).toBeUndefined();
      expect(cross[0].diagnostics).toHaveLength(0);
      // ruleErrors is the strict signal — cfgA's plugin set doesn't
      // know about pkgB/no-null, so resolveRule returns null and the
      // rule resolver records it.
      expect(cross[0].ruleErrors?.some((e) => e.rule === 'pkgB/no-null')).toBe(
        true,
      );

      await pool.shutdown();
    }, 20_000);

    // Same monorepo-style dispatch as the cfgA/cfgB test above, but
    // BOTH dimensions diverge: cfgX exposes only `plugin-x` (module
    // `./plugin-x.mjs`) under prefix `px`, cfgY exposes only the
    // disjoint `plugin-y` module under prefix `py`. The two plugins
    // share zero identity — different module URL, different rule names,
    // different prefix. The earlier test guards same-module/different-
    // prefix; this one guards the harder claim that the worker's
    // per-config `loadedPluginsByDir` keeps two ENTIRELY DIFFERENT
    // plugin sets isolated, with no shared registry that could leak
    // rules from one config into the other's task.
    test('multi-config dispatch with disjoint plugin sets stays isolated', async () => {
      const cfgXDir = path.resolve(__dirname, 'fixtures', 'cfgX');
      const cfgYDir = path.resolve(__dirname, 'fixtures', 'cfgY');

      const pool = new WorkerPool({
        configs: [
          {
            configPath: path.join(cfgXDir, 'rslint.config.mjs'),
            configDirectory: cfgXDir,
          },
          {
            configPath: path.join(cfgYDir, 'rslint.config.mjs'),
            configDirectory: cfgYDir,
          },
        ],
        workerCount: 1, // single worker holds both → exercises per-config map
        taskTimeoutMs: 10_000,
      });
      await pool.init();

      // Source mentions BOTH banned identifiers. The rule that fires
      // is determined purely by which plugin the file's configKey
      // routes to — `foo` should only flag under cfgX, `bar` only
      // under cfgY.
      const SRC = 'const foo = 1; const bar = 2;';

      const results = await pool.lintBatch([
        {
          filePath: 'x.ts',
          text: SRC,
          rules: { 'px/no-foo': { options: [] } },
          collectFixes: false,
          suggestionsMode: 'off',
          configKey: cfgXDir,
        },
        {
          filePath: 'y.ts',
          text: SRC,
          rules: { 'py/no-bar': { options: [] } },
          collectFixes: false,
          suggestionsMode: 'off',
          configKey: cfgYDir,
        },
      ]);

      expect(results).toHaveLength(2);

      // cfgX file: only `px/no-foo` fires, flagging `foo` exactly once.
      expect(results[0].parseError).toBeUndefined();
      expect(results[0].diagnostics).toHaveLength(1);
      expect(results[0].diagnostics[0].ruleName).toBe('px/no-foo');

      // cfgY file: only `py/no-bar` fires, flagging `bar` exactly once.
      expect(results[1].parseError).toBeUndefined();
      expect(results[1].diagnostics).toHaveLength(1);
      expect(results[1].diagnostics[0].ruleName).toBe('py/no-bar');

      // Cross-confusion both directions: requesting plugin Y's rule on
      // a cfgX file (or X's rule on a cfgY file) must NOT resolve.
      // resolveRule returning null surfaces as a ruleErrors entry; a
      // worker bug that merged the two LoadedPlugins would silently
      // produce diagnostics here instead.
      const cross = await pool.lintBatch([
        {
          filePath: 'cross-x-asks-y.ts',
          text: SRC,
          rules: { 'py/no-bar': { options: [] } },
          collectFixes: false,
          suggestionsMode: 'off',
          configKey: cfgXDir,
        },
        {
          filePath: 'cross-y-asks-x.ts',
          text: SRC,
          rules: { 'px/no-foo': { options: [] } },
          collectFixes: false,
          suggestionsMode: 'off',
          configKey: cfgYDir,
        },
      ]);

      expect(cross[0].diagnostics).toHaveLength(0);
      expect(cross[0].ruleErrors?.some((e) => e.rule === 'py/no-bar')).toBe(
        true,
      );
      expect(cross[1].diagnostics).toHaveLength(0);
      expect(cross[1].ruleErrors?.some((e) => e.rule === 'px/no-foo')).toBe(
        true,
      );

      await pool.shutdown();
    }, 20_000);

    // Unknown configKey on the wire is treated as an internal-invariant
    // violation: the host contract guarantees every configKey was declared
    // in WorkerPoolOptions.configs[]. Silently linting with empty rules
    // would mask the bug — the worker emits a parseError pointing at the
    // missing key so the host can surface it.
    test('unknown configKey produces parseError, not silent empty-rules', async () => {
      const cfgADir = path.resolve(__dirname, 'fixtures', 'cfgA');

      const pool = new WorkerPool({
        configs: [
          {
            configPath: path.join(cfgADir, 'rslint.config.mjs'),
            configDirectory: cfgADir,
          },
        ],
        workerCount: 1,
        taskTimeoutMs: 10_000,
      });
      await pool.init();

      const bogusKey = '/not/in/worker/configs';
      const results = await pool.lintBatch([
        {
          filePath: 'rogue.ts',
          text: 'const x = null;',
          rules: { 'pkgA/no-null': { options: [] } },
          collectFixes: false,
          suggestionsMode: 'off',
          configKey: bogusKey,
        },
      ]);

      expect(results).toHaveLength(1);
      expect(results[0].diagnostics).toHaveLength(0);
      expect(results[0].parseError).toBeDefined();
      // The error message must name BOTH the rogue key and the set of
      // configured directories so an operator can tell whether the host
      // sent the wrong key or the worker was built with the wrong set.
      expect(results[0].parseError).toMatch(/configKey/);
      expect(results[0].parseError).toContain(bogusKey);
      expect(results[0].parseError).toContain(cfgADir);

      await pool.shutdown();
    }, 20_000);

    // R4: rule-side warnings (listener throw, fix() throw, suggest fix
    // throw, invalid fix return) used to go through `process.stderr.write`
    // directly. Under LSP that lands in VS Code's hidden "Window"
    // channel — invisible to a user trying to debug a misbehaving
    // plugin. After R4, the runner uses `console.error`, which
    // `lint-worker.ts` monkey-patches into `parentPort.postMessage(
    // { kind: 'log', level: 'error', text })`. This propagates through
    // the WorkerPool's `onLog` callback to the host, which in turn
    // surfaces it in the user-visible log channel.
    //
    // This test pins the wire path end-to-end: a fixture rule whose
    // fix() throws → onLog sees the forwarded error string. Pre-fix,
    // the rule's stderr never reached `onLog`.
    test('R4: rule fix() throw is forwarded through onLog (LSP-visible)', async () => {
      // Inline a fixture plugin that always errors in fix(). Reuse the
      // local config dir's plugin path so config plumbing matches the
      // pool's expectations, but override the rule resolution via the
      // request's `rules` map. (The pool config still loads the local
      // plugin so worker init succeeds; the task explicitly requests a
      // non-existent rule so we exercise the no-rule path instead.)
      //
      // Cleaner: write a tiny fixture plugin + config.
      const fs = await import('node:fs/promises');
      const fxDir = path.resolve(__dirname, 'fixtures');
      const pluginPath = path.join(fxDir, '_r4-bad-fix-plugin.mjs');
      const cfgPath = path.join(fxDir, '_r4-bad-fix.config.mjs');
      await fs.writeFile(
        pluginPath,
        `export default {
  meta: { name: 'r4', version: '0.0.0' },
  rules: {
    'bad-fix': {
      meta: { type: 'problem', fixable: 'code', schema: [] },
      create(ctx) {
        return {
          Identifier(node) {
            ctx.report({
              node,
              message: 'bad',
              // Throw inside fix(). diagnostic-builder must swallow
              // it (drop the fix, keep the diagnostic) and write a
              // helpful log line that R4 now routes through
              // console.error → parentPort.
              fix: () => { throw new Error('R4_FIX_BOOM'); },
            });
          },
        };
      },
    },
  },
};
`,
        'utf8',
      );
      await fs.writeFile(
        cfgPath,
        `import p from './_r4-bad-fix-plugin.mjs';
export default [{ plugins: { r4: p } }];
`,
        'utf8',
      );

      const logs: Array<{ level: string; source: string; text: string }> = [];
      const pool = new WorkerPool({
        configs: [{ configPath: cfgPath, configDirectory: fxDir }],
        workerCount: 1,
        onLog: (rec) => logs.push(rec),
      });

      try {
        await pool.init();
        await pool.lintBatch([
          {
            filePath: 't.ts',
            text: 'const trigger = 1;',
            rules: { 'r4/bad-fix': { options: [] } },
            collectFixes: true,
            suggestionsMode: 'off',
            configKey: fxDir,
          },
        ]);

        // The forwarded log line must contain the rule name + a hint at
        // the failure path. Pre-R4 the string went to raw stderr and
        // logs would not contain it.
        const errLogs = logs.filter((r) => r.level === 'error');
        const matched = errLogs.find((r) =>
          /r4\/bad-fix.*fix\(\) threw.*R4_FIX_BOOM/.test(r.text),
        );
        expect(matched).toBeDefined();
        // Source must be 'plugin' (runner-side console.error from
        // diagnostic-builder running inside the worker, fed through the
        // patched console).
        expect(matched!.source).toBe('plugin');
      } finally {
        await pool.shutdown();
        await Promise.all([
          fs.rm(pluginPath, { force: true }),
          fs.rm(cfgPath, { force: true }),
        ]);
      }
    }, 15_000);
  },
);
