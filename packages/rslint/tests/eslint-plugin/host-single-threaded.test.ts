import { rs, describe, test, expect, afterEach } from '@rstest/core';
import * as workerPoolMod from '../../src/eslint-plugin/worker-pool.js';
import { createPluginLintHost } from '../../src/eslint-plugin/host.js';

// Pins the only production wiring of --singleThreaded into the plugin worker
// pool: createPluginLintHost(_, _, singleThreaded) must construct the WorkerPool
// with workerCount:1 when singleThreaded is set, and leave it undefined (so the
// pool keeps its min(cpus, 8) default) otherwise. Spying the constructor avoids
// spawning real worker threads and asserts the exact option handed to the pool.
describe('createPluginLintHost — singleThreaded → workerCount wiring', () => {
  afterEach(() => {
    rs.restoreAllMocks();
  });

  function spyWorkerPool(): Array<Record<string, unknown>> {
    const calls: Array<Record<string, unknown>> = [];
    rs.spyOn(workerPoolMod, 'WorkerPool').mockImplementation(
      (opts: Record<string, unknown>) => {
        calls.push(opts);
        return {
          init: async () => {},
          lintBatch: async () => [],
          shutdown: async () => {},
        } as unknown as InstanceType<typeof workerPoolMod.WorkerPool>;
      },
    );
    return calls;
  }

  test('singleThreaded=true constructs WorkerPool with workerCount:1', async () => {
    const calls = spyWorkerPool();
    const host = await createPluginLintHost([], undefined, true);
    expect(calls).toHaveLength(1);
    expect(calls[0].workerCount).toBe(1);
    await host.shutdown();
  });

  test('singleThreaded omitted leaves workerCount undefined (pool default)', async () => {
    const calls = spyWorkerPool();
    const host = await createPluginLintHost([]);
    expect(calls).toHaveLength(1);
    expect(calls[0].workerCount).toBeUndefined();
    await host.shutdown();
  });
});
