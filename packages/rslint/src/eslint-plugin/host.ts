/**
 * Plugin-lint host: owns a {@link WorkerPool} and answers the reverse
 * `pluginLint` requests Go sends. Shared by the CLI engine (framed IPC),
 * the native API's config activation, and the VS Code extension's
 * PluginLintPool (LSP `rslint/pluginLint` requests), so the
 * request→tasks→result boundary lives in exactly one place.
 *
 * Also built as the CLI's private `dist/eslint-plugin/host.js` entry. Keep
 * runtime imports limited to worker coordination and protocol conversion;
 * plugin loading, frame decoding, and rule execution belong in `lint-worker.js`.
 */
import {
  WorkerPool,
  attachWorkerPoolRpc,
  type WorkerPoolOptions,
} from './worker-pool.js';
import {
  buildPluginLintTasks,
  buildPluginLintResult,
  type EslintPluginLintRequest,
  type EslintPluginLintResult,
} from './plugin/plugin-lint-protocol.js';
import type { ConfigDescriptor } from './types.js';
import type { ByteInput, RpcHostChannel } from '../native/types.js';
import { PluginTypeServices } from './type-service-host.js';
import type { TypeServiceHostTransport } from './plugin/type-service-protocol.js';
import { resolvePluginAttachments } from './plugin/attachments.js';

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
    attachments?: readonly ByteInput[],
    typeServices?: TypeServiceHostTransport,
  ): Promise<EslintPluginLintResult>;
  /** Drain in-flight tasks and terminate the worker pool. Idempotent. */
  shutdown(): Promise<void>;
}

/**
 * Build a {@link PluginLintHost} over a freshly-initialized WorkerPool.
 *
 * `configs` empty ⇒ the pool spawns no workers (no-op fast path); a
 * `lint` call then returns empty per-file results. Init rejects if a
 * referenced plugin fails to import. The caller owns failure handling:
 * CLI/API abort; LSP applies its configuration transaction recovery policy.
 */
export async function createPluginLintHost(
  configs: ConfigDescriptor[],
  onLog?: WorkerPoolOptions['onLog'],
  singleThreaded?: boolean,
): Promise<PluginLintHost> {
  return createPluginLintHostWithRpc(configs, onLog, singleThreaded, undefined);
}

export type RpcChannelProvider =
  RpcHostChannel | (() => Promise<RpcHostChannel | undefined>);

/**
 * Internal host factory used by CLI and native API hosts to attach the optional
 * shared-memory RPC channel. Accepts either an eager `RpcHostChannel` or a lazy
 * opener invoked only when the first typed plugin task is dispatched so
 * syntax-only runs never allocate shared memory or start Go's slot scanner.
 * Intentionally not re-exported from the public `@rslint/core/eslint-plugin`
 * entrypoint so `.d.ts` and editor contracts stay unchanged.
 */
export async function createPluginLintHostWithRpc(
  configs: ConfigDescriptor[],
  onLog?: WorkerPoolOptions['onLog'],
  singleThreaded?: boolean,
  rpcChannel?: RpcChannelProvider,
): Promise<PluginLintHost> {
  // --singleThreaded caps both warmup and execution at one worker.
  const pool = new WorkerPool({
    configs,
    onLog,
    workerCount: singleThreaded ? 1 : undefined,
  });
  const rpcOpener = typeof rpcChannel === 'function' ? rpcChannel : undefined;
  if (rpcChannel && typeof rpcChannel !== 'function') {
    attachWorkerPoolRpc(pool, rpcChannel);
  }
  let rpcReady = !rpcOpener || configs.length === 0;
  let rpcOpenPromise: Promise<void> | undefined;
  let closed = false;
  const ensureRpc = async (): Promise<void> => {
    if (rpcReady || closed || !rpcOpener) return;
    rpcOpenPromise ??= Promise.resolve()
      .then(rpcOpener)
      .then(
        (channel) => {
          if (channel && !closed) {
            attachWorkerPoolRpc(pool, channel);
          }
        },
        () => undefined,
      )
      .finally(() => {
        rpcReady = true;
      });
    await rpcOpenPromise;
  };
  const types = new PluginTypeServices();
  await pool.init();
  return {
    async lint(req, signal, attachments, transport) {
      if (req.retiredTypeServices?.length) {
        // Cancel workers before retiring host sessions so in-flight reads
        // observe cancellation rather than a closed-session error.
        pool.retireTypeServices(req.retiredTypeServices);
        types.retire(req.retiredTypeServices);
      }
      if (req.files.length === 0) return { results: [] };
      const resolved = resolvePluginAttachments(req, attachments);
      const tasks = buildPluginLintTasks(resolved);
      if (!transport) {
        // Backwards compatibility: older callers (e.g. an editor host before
        // wiring rslint/typeService) omit transport; strip typeServices so
        // untyped rules and scope analysis stay purely local instead of failing.
        for (const task of tasks) task.typeServices = undefined;
      } else if (
        !rpcReady &&
        !signal?.aborted &&
        tasks.some((task) => task.typeServices && !task.typeServices.syntaxOnly)
      ) {
        await ensureRpc();
      }
      // Cancel the dispatched worker tasks if the caller aborts (a superseding
      // keystroke / document close). cancelTask works whether the task is still
      // queued or already running (cooperative SAB cancel-flag).
      const dispatchedTaskIds: number[] = [];
      const onAbort = () => {
        for (const id of dispatchedTaskIds) pool.cancelTask(id);
      };
      signal?.addEventListener('abort', onAbort);
      try {
        const results = await pool.lintBatch(
          tasks,
          (taskId) => {
            dispatchedTaskIds.push(taskId);
            if (signal?.aborted) pool.cancelTask(taskId);
          },
          transport
            ? async (request, taskSignal) =>
                types.query(request, taskSignal, transport)
            : undefined,
        );
        return buildPluginLintResult(results);
      } finally {
        signal?.removeEventListener('abort', onAbort);
      }
    },
    async shutdown() {
      closed = true;
      try {
        await pool.shutdown();
      } finally {
        types.close();
      }
    },
  };
}
