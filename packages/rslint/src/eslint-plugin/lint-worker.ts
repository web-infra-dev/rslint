/* rslint-disable @typescript-eslint/no-unsafe-type-assertion */
/**
 * Worker-thread entry for the runner.
 *
 * Lifecycle:
 *
 *   parent ──spawn(workerData={configs, cancelSab})──▶
 *     1. await loadPluginsFromConfigs(configs) — each ConfigDescriptor
 *        is imported once, the resulting plugin set is cached per
 *        configDirectory for the worker's entire lifetime
 *     2. send {kind:'ready'} to parent
 *     3. on every {kind:'task'} message:
 *          - select LoadedPlugins via request.configKey
 *          - dispatch to lintFile()
 *          - postMessage {kind:'result', taskId, result}
 *     4. on {kind:'shutdown'} → process.exit(0)
 *
 * Note: plugin `console.*` is intentionally NOT patched in-process
 * (see the longer comment around the init block). The host captures
 * the worker's stdout/stderr streams and forwards them to `onLog`.
 *
 * If init fails (config import error, Node version too old, etc.) the
 * worker sends {kind:'init-error', message} and exits non-zero.
 */

import { parentPort, workerData } from 'node:worker_threads';

import {
  lintFile,
  type LintFileRequest,
  type LintFileResult,
} from './linter/ecma-language-plugin.js';
import {
  loadPluginsFromConfigs,
  type LoadedPlugins,
} from './plugin/plugin-loader.js';
import { viewForSlot } from './cancel-flag.js';
import type { ConfigDescriptor } from './types.js';

// ─── Worker bootstrap data ──────────────────────────────────────────────

interface WorkerInitData {
  /** SAB shared with main thread; viewForSlot picks one Int32 view per task. */
  cancelSab?: SharedArrayBuffer;
  /**
   * List of rslint config files this worker is responsible for. Worker
   * imports each one at init via `loadPluginsFromConfigs`, caches the
   * result keyed by `configDirectory`, and uses `request.configKey` to
   * pick the right `LoadedPlugins` per task.
   */
  configs: ConfigDescriptor[];
}

// ─── Messages ──────────────────────────────────────────────────────────

interface TaskMessage {
  kind: 'task';
  taskId: number;
  /** Cancel-pool slot index for this task; -1 if no cancel support. */
  cancelSlot: number;
  request: Omit<LintFileRequest, 'cancelFlag'>;
}

interface ShutdownMessage {
  kind: 'shutdown';
}

type InboundMessage = TaskMessage | ShutdownMessage;

interface ReadyMessage {
  kind: 'ready';
}
interface InitErrorMessage {
  kind: 'init-error';
  message: string;
}
interface ResultMessage {
  kind: 'result';
  taskId: number;
  result: LintFileResult;
}
// ─── Worker implementation ─────────────────────────────────────────────

if (!parentPort) {
  // Not running as a worker — refuse to bootstrap. (The main thread import
  // path doesn't trigger top-level execution because nothing here runs at
  // import-time outside this guard.)
  throw new Error('lint-worker.ts loaded outside a worker_thread context');
}

// Plugin `console.*` output is NOT monkey-patched here. The main
// thread spawns this worker with `{ stdout: true, stderr: true }`
// (see `worker-pool.ts::spawnWorker`); native console behavior is
// preserved inside the worker, and the captured streams are
// forwarded to the host's `onLog` channel from there. Patching
// console in-process would have altered plugin-visible behavior
// (e.g. `console.assert(true)` becoming non-silent), so we
// intentionally don't.

// ── 1. Init: load plugins ──
//
// configs-flow: `data.configs[]` is a list of rslint config files.
// Each is imported once and the resulting plugin set is stored in
// `loadedPluginsByDir`, keyed by `configDirectory`. The per-task
// dispatcher uses `request.configKey` to pick the right
// `LoadedPlugins`. Each config's plugins are naturally anchored at
// the config file's own location (Node's ESM resolver walks
// node_modules from the config file's directory), so monorepos with
// independent plugin installs per sub-package route correctly.
let loadedPluginsByDir: Map<string, LoadedPlugins> | null = null;

const init = async () => {
  const data = workerData as WorkerInitData;
  if (!data || !Array.isArray(data.configs) || data.configs.length === 0) {
    sendInitError('worker received malformed init data (configs[] required)');
    return;
  }
  try {
    loadedPluginsByDir = await loadPluginsFromConfigs(data.configs);
  } catch (err) {
    sendInitError((err as Error)?.message ?? String(err));
    return;
  }
  const ready: ReadyMessage = { kind: 'ready' };
  parentPort!.postMessage(ready);
};

function sendInitError(message: string): void {
  const m: InitErrorMessage = { kind: 'init-error', message };
  parentPort!.postMessage(m);
  // Setting `process.exitCode` instead of calling `process.exit(1)`
  // lets the current event loop drain naturally before the worker
  // terminates. The previous `setImmediate(() => process.exit(1))`
  // assumed Node's MessagePort flush ordering would deliver the
  // init-error message before the immediate fires; that holds in
  // practice but the spec doesn't guarantee it. Exit-code semantics
  // are identical (non-zero exit signals init failure to the
  // parent's WorkerPool spawn handler) and the parent gets the
  // structured `init-error` message reliably.
  process.exitCode = 1;
  // Remove the inbound message listener so the worker doesn't keep
  // the event loop alive waiting for tasks it will never service.
  parentPort?.removeAllListeners('message');
}

// ── 2. Task dispatch ──
parentPort.on('message', (msg: InboundMessage) => {
  if (loadedPluginsByDir == null) {
    // Could happen if main thread sends a task before observing 'ready'
    // — defensive only; main pool waits for ready by contract.
    if ((msg as TaskMessage).kind === 'task') {
      const t = msg as TaskMessage;
      const r: ResultMessage = {
        kind: 'result',
        taskId: t.taskId,
        result: {
          filePath: t.request.filePath,
          diagnostics: [],
          fixes: [],
          suggestionsCount: 0,
          cancelled: false,
          parseError: 'worker not initialized',
        },
      };
      parentPort!.postMessage(r);
    }
    return;
  }

  if (msg.kind === 'shutdown') {
    // Use `parentPort.close()` instead of `process.exit(0)`. Per Node
    // docs, `MessagePort.close()` delivers any already-queued
    // outbound messages first, THEN closes the channel — the worker
    // then exits cleanly with code 0 once the event loop has no more
    // work. `process.exit(0)` terminates immediately and CAN drop
    // already-posted result messages that hadn't flushed out of the
    // outbound buffer yet — the host then mis-labels a successful
    // lint as `parseError: 'shutdown'` (its drain path rejects the
    // task when no result arrived).
    //
    // Defensive: also remove the message listener so we don't pick
    // up any extra inbound messages that race between the parent's
    // shutdown send and its terminate fallback.
    parentPort?.removeAllListeners('message');
    parentPort?.close();
    return;
  }

  if (msg.kind === 'task') {
    const data = workerData as WorkerInitData;
    let cancelFlag: Int32Array | undefined;
    if (data.cancelSab && msg.cancelSlot >= 0) {
      cancelFlag = viewForSlot(data.cancelSab, msg.cancelSlot);
    }
    const request: LintFileRequest = { ...msg.request, cancelFlag };

    // Pick the right `LoadedPlugins` from the per-config map via
    // `request.configKey`. The CLI host / LSP host produce tasks
    // with `configKey` set to the file's owning configDirectory — the
    // same string Go writes into `CompatLintFile.ConfigKey`. A miss
    // is an internal-invariant violation (the host contract guarantees
    // every configKey on the wire was declared in this worker's
    // `data.configs[]`); silently lint-ing the file with empty rules
    // would mask it.
    const key = request.configKey ?? '';
    const selected = loadedPluginsByDir.get(key);
    if (selected == null) {
      const knownKeys = Array.from(loadedPluginsByDir.keys()).join(', ');
      const result: LintFileResult = {
        filePath: msg.request.filePath,
        diagnostics: [],
        fixes: [],
        suggestionsCount: 0,
        cancelled: false,
        parseError:
          `worker: configKey ${JSON.stringify(key)} not declared in workerData.configs[]; ` +
          `known: [${knownKeys}]`,
      };
      const out: ResultMessage = {
        kind: 'result',
        taskId: msg.taskId,
        result,
      };
      parentPort!.postMessage(out);
      return;
    }

    let result: LintFileResult;
    try {
      result = lintFile(request, selected);
    } catch (err) {
      // lintFile is expected to capture errors internally; this catch
      // is purely belt-and-suspenders so a programming bug in the
      // pipeline doesn't kill the worker.
      result = {
        filePath: msg.request.filePath,
        diagnostics: [],
        fixes: [],
        suggestionsCount: 0,
        cancelled: false,
        parseError: `worker exception: ${(err as Error)?.message ?? err}`,
      };
    }
    const out: ResultMessage = { kind: 'result', taskId: msg.taskId, result };
    parentPort!.postMessage(out);
  }
});

// Kick off init asynchronously.
void init();
