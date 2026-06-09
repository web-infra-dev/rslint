import { describe, test, expect } from '@rstest/core';

import { WorkerPool } from '../../src/eslint-plugin/worker-pool.js';
import type { LintTask } from '../../src/eslint-plugin/worker-pool.js';

import {
  LOCAL_CONFIG_DIR,
  localConfigs,
  task,
} from './worker-pool-e2e-helpers.js';
import { SKIP_WIN32_NAPI_TEARDOWN } from './win32-napi-teardown.js';

/**
 * WorkerPool end-to-end — cancellation: cancelTask from inside
 * onTaskDispatched aborts before the worker runs, cancelTask drops a
 * still-queued task without ever posting it, and the shutdown drain
 * honors a queued task's `cancelled` flag (cancelled:true, not
 * parseError:shutdown).
 */

// win32 teardown is gated by SKIP_WIN32_NAPI_TEARDOWN (see that file for the
// nodejs/node#34567 rationale); the flag is false so these run on win32 too.
describe.skipIf(SKIP_WIN32_NAPI_TEARDOWN && process.platform === 'win32')(
  'WorkerPool end-to-end with a local fixture plugin',
  () => {
    test('cancelTask inside onTaskDispatched aborts before worker runs', async () => {
      const pool = new WorkerPool({
        configs: localConfigs,
        workerCount: 1,
        taskTimeoutMs: 10_000,
      });
      await pool.init();

      const tasks: LintTask[] = [task('a.ts', 'const x = null;')];

      const dispatchedIds: number[] = [];
      const results = await pool.lintBatch(tasks, (taskId) => {
        dispatchedIds.push(taskId);
        expect(pool.cancelTask(taskId)).toBe(true);
      });

      expect(dispatchedIds).toHaveLength(1);
      expect(results).toHaveLength(1);
      expect(results[0].cancelled).toBe(true);
      expect(results[0].diagnostics).toHaveLength(0);

      await pool.shutdown();
    }, 20_000);

    test('queue: cancelTask drops a queued (not-yet-dispatched) task without posting to worker', async () => {
      // Drive the queue path: with the worker forced non-idle, the
      // task lands in pendingQueue. cancelTask on its id must mark
      // it cancelled. Releasing the worker (kickQueue) then resolves
      // it as cancelled=true without ever posting to the worker.
      const pool = new WorkerPool({
        configs: localConfigs,
        workerCount: 1,
      });
      await pool.init();
      try {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        const ws = (pool as any).workers as Array<{ ready: boolean }>;
        ws[0].ready = false;

        let capturedId = -1;
        const batchP = pool.lintBatch(
          [
            {
              filePath: 'q.ts',
              text: 'const x = null;\n',
              rules: { 'local/no-null': { options: [] } },
              collectFixes: false,
              suggestionsMode: 'off',
              configKey: LOCAL_CONFIG_DIR,
            },
          ],
          (taskId) => {
            capturedId = taskId;
          },
        );
        expect(capturedId).toBeGreaterThan(0);
        const cancelled = pool.cancelTask(capturedId);
        expect(cancelled).toBe(true);

        // Release the worker — kickQueue should see the cancelled
        // marker and resolve without postMessage.
        ws[0].ready = true;
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        (pool as any).kickQueue();

        const results = await batchP;
        expect(results).toHaveLength(1);
        expect(results[0].cancelled).toBe(true);
        expect(results[0].diagnostics).toEqual([]);
        // Task was never dispatched, so no parseError sentinel either.
        expect(results[0].parseError).toBeUndefined();
      } finally {
        await pool.shutdown();
      }
    }, 15_000);

    test('shutdown drain honors q.cancelled — cancelled-then-shutdown reports cancelled:true, not parseError:shutdown', async () => {
      // Regression for review P2 #15. Pre-fix, the drain loop in
      // `shutdown` blindly stamped every queued task as
      // `cancelled: false, parseError: 'shutdown'` regardless of whether
      // the user had already called `cancelTask(taskId)` on it. The
      // `q.cancelled` flag (set by cancelTask for queued entries) was
      // ignored, so a user-initiated cancel that hadn't been picked up
      // by `kickQueue` yet would be mis-labelled as a shutdown failure
      // — a meaningful distinction for LSP result categorisation and
      // strict-runner error counters.
      const pool = new WorkerPool({
        configs: localConfigs,
        workerCount: 1,
      });
      await pool.init();
      // Same wedge as R1: hold the worker non-ready so kickQueue can't
      // dispatch, leaving every task in pendingQueue.
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      const ws = (pool as any).workers as Array<{ ready: boolean }>;
      ws[0].ready = false;

      // Capture taskIds as they get assigned so we can cancel a
      // specific queued one.
      const taskIds: number[] = [];
      const batchP = pool.lintBatch(
        [1, 2, 3].map((i) => ({
          filePath: `q${i}.ts`,
          text: 'const x = null;\n',
          rules: { 'local/no-null': { options: [] } },
          collectFixes: false,
          suggestionsMode: 'off',
          configKey: LOCAL_CONFIG_DIR,
        })),
        (taskId) => {
          taskIds.push(taskId);
        },
      );

      // Let enqueue finish so onTaskDispatched fires for all 3.
      await new Promise((r) => setTimeout(r, 30));
      expect(taskIds).toHaveLength(3);

      // Cancel the MIDDLE queued task — it must surface as
      // `cancelled: true` from the shutdown drain, while the
      // first/third stay as `parseError: 'shutdown'`.
      pool.cancelTask(taskIds[1]);

      await pool.shutdown();
      const result = await batchP;
      expect(result).toHaveLength(3);

      expect(result[0].parseError).toBe('shutdown');
      expect(result[0].cancelled).toBe(false);

      expect(result[1].cancelled).toBe(true);
      expect(result[1].parseError).toBeUndefined();

      expect(result[2].parseError).toBe('shutdown');
      expect(result[2].cancelled).toBe(false);
    }, 15_000);
  },
);
