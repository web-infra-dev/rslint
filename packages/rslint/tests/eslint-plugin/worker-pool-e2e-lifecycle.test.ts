import { describe, test, expect } from 'rstack/test';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';

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
  POOL_SCENARIO_OUTER_DEADLOCK_SENTINEL_MS,
} from './pool-isolation/harness.js';

/**
 * WorkerPool end-to-end — lifecycle: init + lintBatch + shutdown happy
 * path, single-threaded mode, fix / suggestion edges, reuse across
 * batches, and the terminate-fallback shutdown drain.
 *
 * Exercises the full happy path inside the runner package (WorkerPool
 * → worker_threads loading the user's rslint config → queued
 * lintBatch → oxc-parser → normalize → scope → context → listeners →
 * plugin-lint-result-shaped data). The plugin here is the local fixture
 * plugin (`fixtures/local-plugin.mjs`), not an external dependency.
 */

// win32 teardown is gated by SKIP_WIN32_NAPI_TEARDOWN (see that file for the
// nodejs/node#34567 rationale); the flag is false so these run on win32 too.
describe.skipIf(SKIP_WIN32_NAPI_TEARDOWN && process.platform === 'win32')(
  'WorkerPool end-to-end with a local fixture plugin',
  () => {
    test('default warmup loads two workers and a larger batch grows to its maximum', async () => {
      const pool = new WorkerPool({ configs: localConfigs, workerCount: 4 });
      const state = pool as any;
      try {
        await pool.init();
        expect(state.workers).toHaveLength(2);
        const batch = pool.lintBatch(
          Array.from({ length: 20 }, (_, i) =>
            task(`growth${i}.ts`, 'const value = null;'),
          ),
        );
        expect(state.startingWorkers.size).toBe(2);
        await Promise.all([...state.startingWorkers]);
        expect(state.workers).toHaveLength(4);
        const results = await batch;
        expect(results).toHaveLength(20);
        for (const result of results) {
          expect(result.parseError).toBeUndefined();
          expect(result.diagnostics).toHaveLength(1);
        }
      } finally {
        await pool.shutdown();
      }
    });

    test('ten thousand files with cancellation grow once and release every task slot', async () => {
      const pool = new WorkerPool({ configs: localConfigs, workerCount: 8 });
      const state = pool as any;
      try {
        await pool.init();
        const files = Array.from({ length: 10_000 }, (_, i) =>
          task(`large${i}.ts`, 'const value = null;'),
        );
        let index = 0;
        const batch = pool.lintBatch(files, (id) => {
          if (index++ % 3 === 0) pool.cancelTask(id);
        });
        expect(state.workers.length + state.startingWorkers.size).toBe(8);
        expect(state.cancelPool.size).toBeGreaterThanOrEqual(files.length);
        const results = await batch;
        await Promise.all([...state.startingWorkers]);
        expect(results).toHaveLength(files.length);
        for (let i = 0; i < results.length; i++) {
          expect(results[i].filePath).toBe(files[i].filePath);
          expect(results[i].parseError).toBeUndefined();
          expect(results[i].cancelled).toBe(i % 3 === 0);
          expect(results[i].diagnostics).toHaveLength(i % 3 === 0 ? 0 : 1);
        }
        expect(state.workers).toHaveLength(8);
        expect(state.workerExits.size).toBe(8);
        expect(state.pendingQueue).toEqual([]);
        expect(
          state.workers.every((slot: any) => slot.inflight.size === 0),
        ).toBe(true);
        expect(state.cancelPool.slotInUse.some((used: number) => used)).toBe(
          false,
        );
        expect(state.cancelPool.freeList).toHaveLength(state.cancelPool.size);

        // Reusing slots after a large cancelled batch must clear old flags.
        const reused = await pool.lintBatch(
          Array.from({ length: 32 }, (_, i) => task(`reuse${i}.ts`, 'null;')),
        );
        expect(
          reused.every(
            (result) => !result.cancelled && result.diagnostics.length === 1,
          ),
        ).toBe(true);
      } finally {
        await pool.shutdown();
      }
      expect(state.workerExits.size).toBe(0);
      expect(state.initializingWorkers.size).toBe(0);
    });

    test('expansion rejects changed config before import and keeps the warm generation usable', async () => {
      const dir = fs.mkdtempSync(path.join(os.tmpdir(), 'rslint-pool-config-'));
      const configPath = path.join(dir, 'rslint.config.mjs');
      const marker = path.join(dir, 'changed-config-executed');
      fs.writeFileSync(
        configPath,
        `export default [{ plugins: { local: { rules: {
        report: { meta: { schema: [] }, create(context) {
          return { Program(node) { context.report({ node, message: 'original config' }); } };
        } }
      } } } }];`,
      );
      const logs: string[] = [];
      const pool = new WorkerPool({
        configs: [{ configPath, configDirectory: dir }],
        workerCount: 3,
        onLog: (record) => logs.push(record.text),
      });
      try {
        await pool.init();
        fs.writeFileSync(
          configPath,
          `import fs from 'node:fs'; fs.writeFileSync(${JSON.stringify(marker)}, 'executed'); export default [];`,
        );
        const batch = pool.lintBatch(
          Array.from({ length: 6 }, (_, i) => ({
            ...task(`changed${i}.ts`, '', 'local/report'),
            configKey: dir,
          })),
        );
        await Promise.allSettled([...(pool as any).startingWorkers]);
        const results = await batch;
        expect(fs.existsSync(marker)).toBe(false);
        expect(
          logs.some((log) =>
            log.includes('plugin config changed since worker initialization'),
          ),
        ).toBe(true);
        expect((pool as any).workers).toHaveLength(2);
        for (const result of results) {
          expect(result.parseError).toBeUndefined();
          expect(result.diagnostics).toHaveLength(1);
          expect(result.diagnostics[0].message).toBe('original config');
        }
      } finally {
        await pool.shutdown();
        fs.rmSync(dir, { recursive: true, force: true });
      }
    });

    test('config mutation during worker import fails initialization', async () => {
      const dir = fs.mkdtempSync(path.join(os.tmpdir(), 'rslint-pool-import-'));
      const configPath = path.join(dir, 'rslint.config.mjs');
      fs.writeFileSync(
        configPath,
        `import fs from 'node:fs';
        fs.appendFileSync(new URL(import.meta.url), '\\n// changed during import');
        export default [];`,
      );
      const pool = new WorkerPool({
        configs: [{ configPath, configDirectory: dir }],
        workerCount: 1,
      });
      try {
        await expect(pool.init()).rejects.toThrow(
          /plugin config changed since worker initialization/,
        );
      } finally {
        await pool.shutdown();
        fs.rmSync(dir, { recursive: true, force: true });
      }
    });

    test('init + lintBatch + shutdown happy path', async () => {
      const logs: Array<{ level: string; source: string; text: string }> = [];

      const pool = new WorkerPool({
        configs: localConfigs,
        workerCount: 2,
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
    // ./pool-isolation/runner.mjs). Only a clean child exit is success; a
    // native abort remains visible in the audit but fails this test.
    test(
      'shutdown drains a sync-wedged worker via terminate fallback',
      async () => {
        const r = await runPoolScenario('hang-shutdown');
        expect(r.verdict, formatScenarioFailure(r)).toBe('PASS');
      },
      POOL_SCENARIO_OUTER_DEADLOCK_SENTINEL_MS,
    );

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
        workerCount: 4,
      });
      const state = pool as any;
      let expandedWorkers: unknown[] | undefined;
      try {
        await pool.init();
        // Model repeated editor requests: concurrent batches, cancellation,
        // empty batches, and reuse of the same worker and cancellation slots.
        for (let wave = 0; wave < 200; wave++) {
          const cancelFirst = wave % 2 === 0;
          const first = pool.lintBatch(
            Array.from({ length: 8 }, (_, i) =>
              task(`a${wave}-${i}.ts`, 'const value = null;'),
            ),
            (id) => {
              if (cancelFirst) pool.cancelTask(id);
            },
          );
          const second = pool.lintBatch(
            Array.from({ length: 8 }, (_, i) =>
              task(`b${wave}-${i}.ts`, 'const value = null;'),
            ),
          );
          const [a, b, empty] = await Promise.all([
            first,
            second,
            pool.lintBatch([]),
          ]);
          expect(empty).toEqual([]);
          for (const [prefix, results, cancelled] of [
            ['a', a, cancelFirst],
            ['b', b, false],
          ] as const) {
            expect(results).toHaveLength(8);
            for (let i = 0; i < results.length; i++) {
              expect(results[i].filePath).toBe(`${prefix}${wave}-${i}.ts`);
              expect(results[i].parseError).toBeUndefined();
              expect(results[i].cancelled).toBe(cancelled);
              expect(results[i].diagnostics).toHaveLength(cancelled ? 0 : 1);
            }
          }
          await Promise.all([...state.startingWorkers]);
          const workers = state.workers.map((slot: any) => slot.worker);
          expandedWorkers ??= workers;
          expect(workers).toEqual(expandedWorkers);
          expect(workers).toHaveLength(4);
          expect(state.respawns.size).toBe(0);
          expect(state.workerExits.size).toBe(4);
          expect(state.pendingQueue).toEqual([]);
          expect(state.cancelPool.slotInUse.some((used: number) => used)).toBe(
            false,
          );
          expect(state.cancelPool.freeList).toHaveLength(state.cancelPool.size);
        }
      } finally {
        await pool.shutdown();
      }
      expect(state.workerExits.size).toBe(0);
    });

    // U11: a plugin with a refed top-level `setInterval` keeps the worker event
    // loop alive, so `pool.shutdown()` can't drain it cooperatively and must
    // escalate to a forced `terminate()`. Terminating a worker that holds the
    // oxc-napi addon can native-abort below the JS layer on Windows — which
    // would crash THIS test process if the pool ran in-process. So the scenario
    // runs in an isolated subprocess (see ./pool-isolation/runner.mjs): a native
    // abort is confined there, and the pool's outcome comes back via milestones.
    //
    // PASS requires exact terminate evidence, pool drain, and a clean child
    // exit. Native aborts remain diagnostic-only and cannot make CI green.
    test(
      'U11: plugin with a top-level setInterval — shutdown still terminates the worker',
      async () => {
        const r = await runPoolScenario('u11');
        expect(r.verdict, formatScenarioFailure(r)).toBe('PASS');
      },
      POOL_SCENARIO_OUTER_DEADLOCK_SENTINEL_MS,
    );
  },
);
