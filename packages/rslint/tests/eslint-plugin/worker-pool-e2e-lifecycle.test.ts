import { describe, test, expect } from '@rstest/core';

import { WorkerPool } from '../../src/eslint-plugin/worker-pool.js';
import type { LintTask } from '../../src/eslint-plugin/worker-pool.js';

import {
  LOCAL_CONFIG_DIR,
  localConfigs,
  task,
} from './worker-pool-e2e-helpers.js';
import { SKIP_WIN32_NAPI_TEARDOWN } from './win32-napi-teardown.js';
import {
  runPoolScenario,
  formatScenarioFailure,
} from './pool-isolation/harness.js';

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

// win32 teardown is gated by SKIP_WIN32_NAPI_TEARDOWN (see that file for the
// nodejs/node#34567 rationale); the flag is false so these run on win32 too.
describe.skipIf(SKIP_WIN32_NAPI_TEARDOWN && process.platform === 'win32')(
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

    // A plugin listener wedged in a sync infinite loop can't process the
    // inbound shutdown message, so shutdown must escalate to a forced
    // terminate() after the 5s grace. Terminating an oxc-napi worker can
    // native-abort below the JS layer on Windows — run it isolated (see
    // ./pool-isolation/runner.mjs). PASS = clean terminate; TOLERATED-PASS =
    // pool reached terminate then the subprocess aborted; FAIL = hang, failed
    // in-child assertion, or a crash before the terminate point.
    test('shutdown drains a sync-wedged worker via terminate fallback within grace', async () => {
      const r = await runPoolScenario('hang-shutdown');
      expect(r.verdict, formatScenarioFailure(r)).not.toBe('FAIL');
    }, 20_000);

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

    // U11: a plugin with a refed top-level `setInterval` keeps the worker event
    // loop alive, so `pool.shutdown()` can't drain it cooperatively and must
    // escalate to a forced `terminate()`. Terminating a worker that holds the
    // oxc-napi addon can native-abort below the JS layer on Windows — which
    // would crash THIS test process if the pool ran in-process. So the scenario
    // runs in an isolated subprocess (see ./pool-isolation/runner.mjs): a native
    // abort is confined there, and the pool's outcome comes back via milestones.
    //
    //   PASS           = clean terminate (mac/linux): shutdown bounded + drained
    //   TOLERATED-PASS = pool reached terminate, then the subprocess aborted
    //                    (Windows napi teardown) — the pool still did its job
    //   FAIL           = hang, failed in-child assertion, or a crash BEFORE the
    //                    terminate point — a real regression
    test('U11: plugin with a top-level setInterval — shutdown still terminates the worker', async () => {
      const r = await runPoolScenario('u11');
      expect(r.verdict, formatScenarioFailure(r)).not.toBe('FAIL');
    }, 20_000);
  },
);
