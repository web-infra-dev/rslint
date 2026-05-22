import { describe, test, expect } from '@rstest/core';
import path from 'node:path';

import { WorkerPool } from '../src/worker-pool.js';
import type { LintTask } from '../src/worker-pool.js';

import {
  LOCAL_CONFIG_DIR,
  localConfigs,
  task,
} from './worker-pool-e2e-helpers.js';

/**
 * WorkerPool end-to-end — resilience: a non-cloneable task degrades to
 * a clean per-file error (and a flood of them doesn't overflow the
 * stack), a worker exiting while shutdown is in flight doesn't leak its
 * respawn, and a hung listener trips task_timeout → respawn → recovery.
 */

describe('WorkerPool end-to-end with a local fixture plugin', () => {
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

  // Respawn-during-shutdown race regression — see the long jsdoc on
  // the failing case in the previous version of this file. Same
  // contract under configs-flow: shutdown completes promptly and the
  // workers array stays empty after a leaked respawn settles.
  test('worker exit racing shutdown does not leak the respawn', async () => {
    const pool = new WorkerPool({
      configs: localConfigs,
      workerCount: 2,
      taskTimeoutMs: 5_000,
    });
    await pool.init();

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const workers = (pool as any).workers as Array<{
      worker: { terminate(): Promise<number> };
    }>;
    expect(workers.length).toBeGreaterThan(0);
    void workers[0].worker.terminate();

    const shutdownStart = Date.now();
    await pool.shutdown();
    const shutdownElapsed = Date.now() - shutdownStart;
    expect(shutdownElapsed).toBeLessThan(10_000);

    await expect(
      pool.lintBatch([
        {
          filePath: 'x.ts',
          text: 'const x = 1;',
          rules: {},
          collectFixes: false,
          suggestionsMode: 'off',
          configKey: LOCAL_CONFIG_DIR,
        },
      ]),
    ).rejects.toThrow(/closed/);

    await new Promise<void>((r) => setTimeout(r, 200));
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const finalWorkers = (pool as any).workers as Array<unknown>;
    expect(finalWorkers.length).toBe(0);
  });

  test('hung listener → task_timeout → respawn → next batch succeeds', async () => {
    const hangConfigPath = path.resolve(
      __dirname,
      'fixtures',
      'hang.config.mjs',
    );
    const hangConfigDir = path.dirname(hangConfigPath);

    const logs: Array<{ level: string; source: string; text: string }> = [];
    const pool = new WorkerPool({
      configs: [
        {
          configPath: hangConfigPath,
          configDirectory: hangConfigDir,
        },
      ],
      workerCount: 1,
      // Aggressive timeout so the test doesn't sit for 30 s. 600 ms
      // is comfortably above worker startup + import latency on CI
      // (the existing happy-path tests in this file complete in ~50–
      // 200 ms once the worker is up), but well below the 60-s test
      // ceiling.
      taskTimeoutMs: 600,
      onLog: (rec) => logs.push(rec),
    });
    await pool.init();

    const hangResult = await pool.lintBatch([
      {
        filePath: 'wedge.ts',
        text: 'const x = 1;\n',
        rules: { 'hang/hang': { options: [] } },
        collectFixes: false,
        suggestionsMode: 'off',
        configKey: hangConfigDir,
      },
    ]);
    expect(hangResult).toHaveLength(1);
    expect(hangResult[0].parseError).toBe('task_timeout');

    await new Promise((r) => setTimeout(r, 500));

    const okResult = await pool.lintBatch([
      {
        filePath: 'ok.ts',
        text: 'const TRIGGER = 1;\n',
        rules: { 'hang/noop': { options: [] } },
        collectFixes: false,
        suggestionsMode: 'off',
        configKey: hangConfigDir,
      },
    ]);
    expect(okResult).toHaveLength(1);
    expect(okResult[0].parseError).toBeUndefined();
    expect(okResult[0].diagnostics).toHaveLength(1);
    expect(okResult[0].diagnostics[0].message).toBe('noop fired');

    const respawnLog = logs.find((l) => l.text.includes('respawning'));
    expect(respawnLog).toBeDefined();

    await pool.shutdown();
  }, 30_000);
});
