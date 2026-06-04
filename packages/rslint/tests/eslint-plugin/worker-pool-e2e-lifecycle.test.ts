import { describe, test, expect } from '@rstest/core';
import path from 'node:path';

import { WorkerPool } from '../../src/eslint-plugin/worker-pool.js';
import type { LintTask } from '../../src/eslint-plugin/worker-pool.js';

import {
  LOCAL_CONFIG_DIR,
  localConfigs,
  task,
} from './worker-pool-e2e-helpers.js';

/**
 * WorkerPool end-to-end — lifecycle: init + lintBatch + shutdown happy
 * path, single-threaded mode, fix / suggestion edges, reuse across
 * batches, and the terminate-fallback shutdown drain.
 *
 * Exercises the full happy path inside the runner package (WorkerPool
 * → worker_threads loading the user's rslint config → round-robin
 * lintBatch → oxc-parser → normalize → scope → context → listeners →
 * plugin-lint-result-shaped data). The plugin here is the local fixture
 * plugin (`fixtures/local-plugin.mjs`), not an external dependency.
 */

// Skipped on windows: tearing down a worker that has oxc (a napi addon)
// loaded aborts below the JS layer there (nodejs/node#34567) and crashes
// the rstest worker running this file. These e2e tests spawn real
// workers and tear them down, so they are windows-skipped; they still
// run on linux/macOS.
describe.skipIf(process.platform === 'win32')(
  'WorkerPool end-to-end with a local fixture plugin',
  () => {
    test('init + lintBatch + shutdown happy path', async () => {
      const logs: Array<{ level: string; source: string; text: string }> = [];

      const pool = new WorkerPool({
        configs: localConfigs,
        workerCount: 2,
        taskTimeoutMs: 10_000,
        onLog: (rec) => logs.push(rec),
      });

      await pool.init();

      const tasks: LintTask[] = [
        task('a.ts', `const x = null;`),
        task('b.ts', `const y = "ok"; const z = null;`),
        task('c.ts', `// no nulls here\nconst v = 42;`),
      ];

      const results = await pool.lintBatch(tasks);
      expect(results).toHaveLength(3);

      expect(results[0].diagnostics).toHaveLength(1);
      expect(results[0].diagnostics[0].ruleName).toBe('local/no-null');

      expect(results[1].diagnostics).toHaveLength(1);

      expect(results[2].diagnostics).toHaveLength(0);

      await pool.shutdown();
    });

    test('--singleThreaded honors workerCount=1', async () => {
      const pool = new WorkerPool({
        configs: localConfigs,
        workerCount: 1,
      });

      await pool.init();

      const tasks: LintTask[] = Array.from({ length: 5 }, (_, i) =>
        task(`f${i}.ts`, `const v${i} = null;`),
      );

      const results = await pool.lintBatch(tasks);
      expect(results).toHaveLength(5);
      for (const r of results) {
        expect(r.diagnostics).toHaveLength(1);
      }

      await pool.shutdown();
    });

    test('--fix flag plumbs through; fixes are returned in results', async () => {
      const pool = new WorkerPool({
        configs: localConfigs,
        workerCount: 1,
      });
      await pool.init();

      // local/prefer-array-some HAS an autofix (replaces the entire
      // .filter(...).length > 0 chain with .some(...)). Picking this
      // rule rather than no-null (which is suggestion-only) is what
      // actually exercises the collectFixes:true → result.fixes path
      // end-to-end. A regression that silently drops collectFixes
      // would leave result.fixes empty even though the diagnostic
      // still fires.
      const SRC = `const arr = [1, 2]; const r = arr.filter(x => x > 0).length > 0;`;
      const results = await pool.lintBatch([
        {
          filePath: 'fix.ts',
          text: SRC,
          rules: { 'local/prefer-array-some': { options: [] } },
          collectFixes: true,
          suggestionsMode: 'off',
          configKey: LOCAL_CONFIG_DIR,
        },
      ]);

      expect(results[0].diagnostics).toHaveLength(1);
      // Aggregated fixes (mirror of diagnostics[].fixes) must contain
      // the rule's autofix bytes.
      const fixes = results[0].fixes ?? [];
      expect(fixes.length).toBeGreaterThanOrEqual(1);
      // The autofix's text is the literal replacement bytes — for this
      // rule it's the identifier `some` (rule rewrites just the
      // `filter` member name, not the whole chain). Match on exact
      // contents so a regression that emits empty / wrong bytes is
      // caught.
      expect(fixes[0].text).toBe('some');

      await pool.shutdown();
    });

    test('suggestionsMode=eager produces resolved fix edges', async () => {
      const pool = new WorkerPool({
        configs: localConfigs,
        workerCount: 1,
      });
      await pool.init();

      const results = await pool.lintBatch([
        {
          filePath: 'sug.ts',
          text: `const x = null;`,
          rules: { 'local/no-null': { options: [] } },
          collectFixes: false,
          suggestionsMode: 'eager',
          configKey: LOCAL_CONFIG_DIR,
        },
      ]);

      expect(results[0].diagnostics).toHaveLength(1);
      // suggestionsMode='eager' must materialize suggestion fix edges.
      // local/no-null offers a suggestion that rewrites `null` to
      // `undefined`. A regression that silently drops 'eager' (treating
      // it as 'off') would leave `suggestions[*].fixes` as null/empty
      // even though the diagnostic itself still fires. The previous
      // version of this test only checked `suggestionsCount >= 0`, a
      // tautology — fixed to assert the actual resolved fix bytes.
      const suggestions = results[0].diagnostics[0].suggestions ?? [];
      expect(suggestions.length).toBeGreaterThanOrEqual(1);
      const fixesArr = suggestions[0].fixes ?? [];
      expect(fixesArr.length).toBeGreaterThanOrEqual(1);
      expect(fixesArr[0].text).toBe('undefined');

      await pool.shutdown();
    });

    // pool.shutdown() posts {kind:'shutdown'} to every worker and waits
    // up to 5 s for each to exit cooperatively before falling back to
    // `worker.terminate()`. This invariant matters when the host (CLI
    // parent on Ctrl-C, VS Code on extension dispose, …) tears the pool
    // down while a plugin listener is wedged in a sync infinite loop —
    // the worker can't even process the inbound shutdown message, so
    // graceful exit is impossible and the pool must escalate to a forced
    // terminate. Test fires a never-completing task, calls shutdown
    // without a per-task timeout, and asserts the pool fully drains
    // within grace + slack.
    test('shutdown drains a sync-wedged worker via terminate fallback within grace', async () => {
      const hangConfigPath = path.resolve(
        __dirname,
        'fixtures',
        'hang.config.mjs',
      );
      const hangConfigDir = path.dirname(hangConfigPath);

      const pool = new WorkerPool({
        configs: [
          {
            configPath: hangConfigPath,
            configDirectory: hangConfigDir,
          },
        ],
        workerCount: 1,
        // 60 s — longer than the test ceiling, so a task_timeout-driven
        // respawn can't preempt the shutdown path we're trying to test.
        taskTimeoutMs: 60_000,
      });
      await pool.init();

      // Fire-and-forget: the hang listener spins forever on Program, so
      // the task never resolves naturally. shutdown is the only force
      // that can free the pool.
      const wedgeP = pool.lintBatch([
        {
          filePath: 'wedge.ts',
          text: 'const x = 1;\n',
          rules: { 'hang/hang': { options: [] } },
          collectFixes: false,
          suggestionsMode: 'off',
          configKey: hangConfigDir,
        },
      ]);

      // Brief grace so the worker is actually inside the sync hang
      // when shutdown lands. Without this the worker might still be
      // sitting in its inbound queue and could process the shutdown
      // message before entering the loop — a different (faster) path
      // that doesn't exercise the terminate fallback we're testing.
      await new Promise((r) => setTimeout(r, 200));

      const shutdownStart = Date.now();
      await pool.shutdown();
      const elapsed = Date.now() - shutdownStart;

      // worker-pool.ts grace = 5 s; total elapsed = grace + small
      // overhead for terminate() to actually kill the worker thread.
      // Lower bound 4.5 s catches a regression where shutdown returns
      // immediately (without waiting for the worker). Upper bound 8 s
      // catches a regression where the terminate fallback never fires.
      expect(elapsed).toBeGreaterThanOrEqual(4_500);
      expect(elapsed).toBeLessThan(8_000);

      // In-flight tasks resolve with a sentinel parseError instead of
      // hanging the caller forever. worker-pool.ts:539-548 maps the
      // shutdown-kind rejection back to a result-shaped failure so
      // callers (internal/linter) see one file as plugin-lint-failed rather
      // than the whole batch throwing.
      const wedgeResults = await wedgeP;
      expect(wedgeResults).toHaveLength(1);
      expect(wedgeResults[0].parseError).toBe('shutdown');

      // Pool is fully drained — no leaked worker slot.
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      const finalWorkers = (pool as any).workers as Array<unknown>;
      expect(finalWorkers.length).toBe(0);

      // Idempotent: a second shutdown returns instantly without throwing.
      const secondStart = Date.now();
      await pool.shutdown();
      expect(Date.now() - secondStart).toBeLessThan(50);
    }, 15_000);

    // U12: a single WorkerPool instance must support N lintBatch calls
    // in sequence (the CLI's fix-loop and the LSP's continuous edit
    // stream both do this). State across batches must NOT leak:
    //   - plugin instances stay cached (no re-import per batch)
    //   - diagnostic state is per-batch, not accumulating
    //   - per-task timers are released after each batch
    // The third invariant is the easiest to regress on — a leaked timer
    // would keep the Node event loop alive past `pool.shutdown()`.
    test('U12: WorkerPool reuse across many lintBatch invocations stays stable', async () => {
      const pool = new WorkerPool({
        configs: localConfigs,
        workerCount: 1,
      });
      await pool.init();

      // Run 5 lint batches in sequence. The same worker handles all of
      // them; no respawn / re-init expected.
      const counts: number[] = [];
      for (let i = 0; i < 5; i++) {
        const r = await pool.lintBatch([
          {
            filePath: `iter${i}.ts`,
            text: 'const x = null;\n',
            rules: { 'local/no-null': { options: [] } },
            collectFixes: false,
            suggestionsMode: 'off',
            configKey: LOCAL_CONFIG_DIR,
          },
        ]);
        expect(r).toHaveLength(1);
        expect(r[0].parseError).toBeUndefined();
        counts.push(r[0].diagnostics.length);
      }

      // Every iteration produced the SAME diagnostic count (1, for the
      // single `null` literal). If state leaked (e.g. diagnostics
      // accumulated across batches), counts would grow.
      for (const c of counts) {
        expect(c).toBe(1);
      }

      // Only one worker was used the whole time.
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      const workers = (pool as any).workers as Array<{ ready: boolean }>;
      expect(workers).toHaveLength(1);
      expect(workers[0].ready).toBe(true);

      await pool.shutdown();
    }, 30_000);

    // U11: a plugin that installs a `setInterval` (or any unref-able
    // timer) during its module top-level must NOT prevent the worker
    // from exiting on shutdown. The worker process keeps running until
    // the event loop drains; an unref'd timer is fine, but a refed
    // setInterval would keep the worker alive past `pool.shutdown()`,
    // hanging the parent's Wait.
    //
    // The test uses a fixture plugin that creates a setInterval in its
    // top-level, then verifies pool.shutdown completes in bounded time.
    // If the timer kept the worker alive, shutdown would either time
    // out (caught by Jest's per-test timeout) or rely on terminate()
    // fallback (visible via a longer-than-expected elapsed time).
    test('U11: plugin with a top-level setInterval — shutdown still terminates the worker', async () => {
      const cfgDir = path.resolve(__dirname, 'fixtures');
      const cfgPath = path.join(cfgDir, '_u11.config.mjs');
      const pluginPath = path.join(cfgDir, '_u11-plugin.mjs');
      await import('node:fs/promises').then((fs) =>
        Promise.all([
          fs.writeFile(
            pluginPath,
            `// Intentionally schedules a REFED interval that would
// otherwise keep the worker alive forever. The test pins that
// pool.shutdown() still tears the worker down within a bounded
// elapsed window (5 s graceful + terminate fallback).
const _interval = setInterval(() => { /* never fires within test */ }, 60_000);
// eslint-disable-next-line no-unused-vars
export default {
  meta: { name: 'u11-interval-plugin' },
  rules: { noop: { meta: {}, create() { return {}; } } },
};
`,
            'utf8',
          ),
          fs.writeFile(
            cfgPath,
            `import plugin from './_u11-plugin.mjs';
export default [{ plugins: { u11: plugin } }];
`,
            'utf8',
          ),
        ]),
      );

      try {
        const pool = new WorkerPool({
          configs: [{ configPath: cfgPath, configDirectory: cfgDir }],
          workerCount: 1,
        });
        await pool.init();

        const start = Date.now();
        await pool.shutdown();
        const elapsed = Date.now() - start;

        // Worker pool's shutdown gives each worker up to 5s graceful
        // before terminate(). A worker keeping its event loop alive
        // via setInterval would hit terminate fallback — bounded by
        // grace + small overhead. Test fails if shutdown is unbounded
        // (i.e. hangs past 10s) or if terminate fallback itself fails
        // (worker_threads.terminate is OS-level and very reliable).
        expect(elapsed).toBeLessThan(8_000);

        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        const workers = (pool as any).workers as Array<unknown>;
        expect(workers).toHaveLength(0);
      } finally {
        await import('node:fs/promises').then((fs) =>
          Promise.all([
            fs.rm(cfgPath, { force: true }),
            fs.rm(pluginPath, { force: true }),
          ]),
        );
      }
    }, 15_000);
  },
);
