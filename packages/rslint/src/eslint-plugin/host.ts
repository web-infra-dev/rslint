/**
 * Plugin-lint host: owns a {@link WorkerPool} and answers the reverse
 * `pluginLint` requests Go sends. Shared by BOTH hosts that own a
 * pool — the CLI engine (`packages/rslint/src/cli/engine.ts`, over the IPC
 * channel) and the VS Code extension's PluginLintPool (over the LSP
 * `rslint/pluginLint` request) — so the request→tasks→result
 * boundary lives in exactly one place.
 */
import { WorkerPool, type WorkerPoolOptions } from './worker-pool.js';
import {
  buildPluginLintTasks,
  buildPluginLintResult,
  type EslintPluginLintRequest,
  type EslintPluginLintResult,
} from './plugin/plugin-lint-protocol.js';
import type { ConfigDescriptor } from './types.js';

export interface PluginLintHost {
  /**
   * Run one reverse batch: build per-file tasks, lint, project results. An
   * optional AbortSignal cancels the dispatched worker tasks — the LSP path
   * wires it to a superseding keystroke / document close so the worker stops
   * instead of running to completion.
   */
  lint(
    req: EslintPluginLintRequest,
    signal?: AbortSignal,
  ): Promise<EslintPluginLintResult>;
  /** Drain in-flight tasks and terminate the worker pool. Idempotent. */
  shutdown(): Promise<void>;
}

/**
 * Build a {@link PluginLintHost} over a freshly-initialized WorkerPool.
 *
 * `configs` empty ⇒ the pool spawns no workers (no-op fast path); a
 * `lint` call then returns empty per-file results. Init rejects if a
 * referenced plugin fails to import — the caller decides how loud to be
 * (CLI fails the run; LSP logs and serves empty).
 */
export async function createPluginLintHost(
  configs: ConfigDescriptor[],
  onLog?: WorkerPoolOptions['onLog'],
  singleThreaded?: boolean,
): Promise<PluginLintHost> {
  // --singleThreaded forces a single worker thread (workerCount=1), the JS
  // analog of the Go pass's no-concurrency mode. Otherwise leave workerCount
  // undefined so the pool keeps its min(cpus, 8) default.
  const pool = new WorkerPool({
    configs,
    onLog,
    workerCount: singleThreaded ? 1 : undefined,
  });
  await pool.init();
  // The worker keys its plugin map on configDirectory verbatim, and the wire
  // configKey Go sends is byte-identical to it (CLI: the raw string Go echoes
  // back; LSP: the same URI) — so no normalization is needed or wanted here.
  const configDirSet = new Set(configs.map((c) => c.configDirectory));
  return {
    async lint(req, signal) {
      const tasks = buildPluginLintTasks(req, {
        configDirSet,
        onUnknownConfigKey: (filePath, configKey) =>
          onLog?.({
            level: 'error',
            source: 'runner',
            text: `eslint-plugin: file ${filePath} carries unknown configKey ${configKey}`,
          }),
      });
      // Cancel the dispatched worker tasks if the caller aborts (a superseding
      // keystroke / document close). cancelTask works whether the task is still
      // queued or already running (cooperative SAB cancel-flag).
      const dispatchedTaskIds: number[] = [];
      const onAbort = () => {
        for (const id of dispatchedTaskIds) pool.cancelTask(id);
      };
      signal?.addEventListener('abort', onAbort);
      try {
        const results = await pool.lintBatch(tasks, (taskId) => {
          dispatchedTaskIds.push(taskId);
          if (signal?.aborted) pool.cancelTask(taskId);
        });
        return buildPluginLintResult(results);
      } finally {
        signal?.removeEventListener('abort', onAbort);
      }
    },
    async shutdown() {
      await pool.shutdown();
    },
  };
}
