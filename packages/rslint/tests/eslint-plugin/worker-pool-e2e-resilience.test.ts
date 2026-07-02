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
 * WorkerPool end-to-end — resilience: a non-cloneable task degrades to
 * a clean per-file error (and a flood of them doesn't overflow the
 * stack), a worker exiting while shutdown is in flight doesn't leak its
 * respawn, and a hung listener trips task_timeout → respawn → recovery.
 */

// win32 teardown is gated by SKIP_WIN32_NAPI_TEARDOWN (see that file for the
// nodejs/node#34567 rationale); the flag is false so these run on win32 too.
describe.skipIf(SKIP_WIN32_NAPI_TEARDOWN && process.platform === 'win32')(
  'WorkerPool end-to-end with a local fixture plugin',
  () => {
    // postMessage path: a non-clonable field on the task makes
    // structuredClone throw inside slot.worker.postMessage. The pool must
    // (a) NOT leak the per-task bookkeeping (timer/cancelSlot/inflight)
    // and (b) surface a result-shaped failure rather than poisoning the
    // whole batch via Promise.all rejection.
    test('postMessage failure on unserializable task yields a clean per-file error', async () => {
      const pool = new WorkerPool({
        configs: localConfigs,
        workerCount: 1,
        taskTimeoutMs: 5_000,
      });
      await pool.init();

      // A function on the rule meta — structuredClone refuses to clone
      // functions (DataCloneError). This deterministically triggers the
      // postMessage throw without depending on worker internals.
      const badTask: LintTask = {
        filePath: 'bad.ts',
        text: `const x = null;`,
        rules: {
          'local/no-null': {
            options: [],
            // eslint-disable-next-line @typescript-eslint/no-explicit-any
            meta: { unserializable: () => 'function' } as any,
          },
        },
        collectFixes: false,
        suggestionsMode: 'off',
        configKey: LOCAL_CONFIG_DIR,
      };

      // Sibling task in the same batch — should still succeed even after
      // the bad task's dispatch throws. This is the "Promise.all isn't
      // poisoned" half of the contract.
      const goodTask: LintTask = task('good.ts', `const y = null;`);

      const results = await pool.lintBatch([badTask, goodTask]);

      expect(results).toHaveLength(2);
      expect(results[0].filePath).toBe('bad.ts');
      expect(results[0].parseError).toMatch(/postMessage_failed/);
      expect(results[0].diagnostics).toHaveLength(0);
      expect(results[1].filePath).toBe('good.ts');
      expect(results[1].diagnostics).toHaveLength(1);

      const followUp = await pool.lintBatch([goodTask]);
      expect(followUp[0].diagnostics).toHaveLength(1);

      await pool.shutdown();
    });

    // #2: a batch of MANY non-cloneable (DataCloneError) tasks must each
    // degrade to `postMessage_failed` independently — NOT overflow the stack.
    // Pre-fix the DataCloneError catch called `kickQueue()` synchronously, so
    // consecutive poison tasks recursed (kickQueue → dispatch → catch →
    // kickQueue → …) and the whole batch rejected with `RangeError: Maximum
    // call stack size exceeded` (measured overflow at a few thousand). The fix
    // defers kickQueue via `queueMicrotask`, giving each re-dispatch a fresh
    // stack frame so every file degrades independently.
    test('#2: large batch of non-cloneable tasks all degrade (no stack overflow)', async () => {
      const pool = new WorkerPool({ configs: localConfigs, workerCount: 1 });
      await pool.init();

      const poison = (i: number): LintTask => ({
        filePath: `poison${i}.ts`,
        text: 'const x = null;',
        rules: {
          // eslint-disable-next-line @typescript-eslint/no-explicit-any
          'local/no-null': { options: [], meta: { f: () => i } as any },
        },
        collectFixes: false,
        suggestionsMode: 'off',
        configKey: LOCAL_CONFIG_DIR,
      });

      // 6000 is comfortably past the pre-fix overflow threshold (~2–5k).
      const N = 6000;
      const results = await pool.lintBatch(
        Array.from({ length: N }, (_, i) => poison(i)),
      );

      expect(results).toHaveLength(N);
      for (const r of results) {
        expect(r.parseError).toMatch(/postMessage_failed/);
        expect(r.diagnostics).toHaveLength(0);
      }
      // Pool still healthy after the poison flood.
      const ok = await pool.lintBatch([task('ok.ts', 'const y = null;')]);
      expect(ok[0].diagnostics).toHaveLength(1);

      await pool.shutdown();
    }, 30_000);

    // Directly terminating one worker of a 2-worker pool then shutting down
    // exercises the respawn-during-shutdown race. The terminate of an oxc-napi
    // worker can native-abort on Windows — run it isolated (see
    // ./pool-isolation). Verdict != FAIL means the pool drained without leaking
    // the respawn (or reached terminate then the subprocess aborted on Windows).
    test('worker exit racing shutdown does not leak the respawn', async () => {
      const r = await runPoolScenario('worker-exit-race');
      expect(r.verdict, formatScenarioFailure(r)).not.toBe('FAIL');
    }, 20_000);

    // A hung listener trips the per-task timeout, which terminates the worker
    // and respawns it; the next batch must succeed on the replacement. The
    // timeout-driven terminate of an oxc-napi worker can native-abort on
    // Windows — run it isolated (see ./pool-isolation).
    test('hung listener → task_timeout → respawn → next batch succeeds', async () => {
      const r = await runPoolScenario('task-timeout');
      expect(r.verdict, formatScenarioFailure(r)).not.toBe('FAIL');
    }, 30_000);
  },
);
