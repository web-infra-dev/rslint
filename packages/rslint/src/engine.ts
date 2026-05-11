/**
 * Engine: Node-side orchestrator for the IPC CLI handshake.
 *
 * Spawned by `cli.ts` for every JS/TS-config CLI run, regardless of
 * whether `eslintPlugins` is configured — the IPC handshake is now the
 * single user-path through `rslint.cjs`. When no plugin entries are
 * present, the WorkerPool is constructed with `workerCount=0` (no Node
 * worker threads spawned) and the Go side never reaches the compat
 * dispatcher; the IPC channel still carries `init`, `output`, and
 * `shutdown`, which keeps the CLI/Go protocol uniform.
 *
 * Lifecycle:
 *
 *   1. caller invokes runEngine({...})
 *   2. spawn Go (`bin/rslint ...goArgs`)
 *      stdio: ['pipe','pipe','inherit']
 *   3. init WorkerPool (no-op when entries=[]) in parallel with spawn
 *   4. send `init` IPC message to Go; await `response{ok}`
 *   5. dispatch loop:
 *        - inbound `lintEslintPlugin` request → pool.lintBatch → response
 *        - inbound `output` notification     → process.stdout.write
 *        - inbound `cancel`                  → pool.cancelTask (LSP only)
 *      until Go sends `shutdown` (request)   → we ack immediately;
 *      Go then exits with its lint exit code on its own.
 *   6. once `child.exit` fires we drain the pool and resolve with
 *      Go's exit code. (Pool cleanup happens AFTER Go exits, not
 *      inside the shutdown ack handler.)
 *
 * On any failure prior to step 5, returns exit code 2 (runner failure
 * distinct from lint error).
 */

import { spawn, type ChildProcess } from 'node:child_process';
import {
  IpcClient,
  WorkerPool,
  buildCompatTasksByConfigKey,
  buildCompatBatchResult,
  type ConfigDescriptor,
} from '@rslint/eslint-plugin-runner';
import type { EslintPluginEntry } from './config-loader.js';

import type {
  IpcMessage,
  CompatBatchInput,
  CompatBatchResult,
} from '@rslint/eslint-plugin-runner';

// POSIX convention: a process terminated by signal N exits with code
// 128 + N. Node's `child.exit` reports the signal NAME (not the number),
// so map the ones we can actually receive. Collapsing every signal to 130
// (SIGINT) would mislabel a SIGTERM/SIGKILL teardown — including our own
// safeKillGo escalation — as a Ctrl-C.
const SIGNAL_EXIT_CODES: Record<string, number> = {
  SIGHUP: 129,
  SIGINT: 130,
  SIGQUIT: 131,
  SIGTERM: 143,
  SIGKILL: 137,
};

// ─── Public types ────────────────────────────────────────────────────

export interface EngineRunOptions {
  /** Path to the Go rslint binary. */
  binPath: string;
  /** Args to pass to the Go binary (user CLI flags forwarded by cli.ts). */
  goArgs: string[];
  /** Wire-shape plugin entries (`{ prefix, ruleNames }`) sent to Go in
   *  the `init` payload. Empty array is allowed (no plugins → no
   *  placeholder rules). The worker does NOT consume this; workers
   *  load plugins from `workerConfigs[]` directly. */
  eslintPluginEntries: EslintPluginEntry[];
  /** Configs the worker pool will import directly. Empty means no
   *  workers are spawned (workerCount=0 fast path). */
  workerConfigs: ConfigDescriptor[];
  /** The full configs array sent to Go in the `init` payload (each
   *  entry is shaped {configDirectory, configPath?, entries}). Go uses
   *  this for its own config processing; the worker pool does not. */
  configs: unknown[];
  /** Working directory (inherits process.cwd() by default). */
  cwd?: string;
  /**
   * Runtime hints. The IPC-bound subset (forceColor, singleThreaded) is
   * forwarded to the Go binary in the `init` payload's `runtime` block.
   */
  runtime?: {
    forceColor?: boolean;
    singleThreaded?: boolean;
  };
  /** stdout sink (default: process.stdout). Lets tests capture output. */
  stdout?: NodeJS.WritableStream;
  /** stderr sink (default: process.stderr). */
  stderr?: NodeJS.WritableStream;
  /**
   * Extra fields merged into the `init` IPC payload. Used by cli.ts to
   * forward positional args, --fix flag, --format, etc. without making
   * engine.ts itself aware of CLI flag layout. Pass-through.
   */
  extraInit?: Record<string, unknown>;
}

export interface EngineRunResult {
  /** Exit code propagated from Go (or 2 on engine-level failure). */
  exitCode: number;
}

// ─── Implementation ──────────────────────────────────────────────────

export async function runEngine(
  opts: EngineRunOptions,
): Promise<EngineRunResult> {
  const stdout = opts.stdout ?? process.stdout;
  const stderr = opts.stderr ?? process.stderr;

  // ── 1. spawn Go and prep WorkerPool in parallel ─────────────────
  // The Go binary's default entry (no --lsp / --api flag) is `runCLI`,
  // which always waits for an `init` IPC message — true for lint, --init,
  // --help, and JSON-config fallback flows alike.
  const child = spawn(opts.binPath, opts.goArgs, {
    stdio: ['pipe', 'pipe', 'inherit'],
    cwd: opts.cwd ?? process.cwd(),
  });

  // ── Child error / exit watch (registered IMMEDIATELY after spawn) ──
  //
  // Node delivers 'error' (e.g. ENOENT, EACCES on the binary path) and
  // 'exit' asynchronously — typically on a later tick. If they fire
  // before we attach listeners, the runtime treats them as unhandled
  // and kills the whole Node process via 'uncaughtException', bypassing
  // every recovery path below. Attaching here, BEFORE any await, makes
  // the unhappy path (missing binary, immediate-exit Go process)
  // observable to every step that follows.
  //
  // childExit always RESOLVES (never rejects); helpers race it against
  // each await so a child that drops out mid-init unwinds cleanly with
  // a non-zero exitCode instead of leaving Node hanging.
  type ChildExit = { code: number; reason: 'error' | 'exit' };
  const childExit = new Promise<ChildExit>((resolve) => {
    let resolved = false;
    const settle = (state: ChildExit) => {
      if (resolved) return;
      resolved = true;
      resolve(state);
    };
    child.once('error', (err) => {
      // ENOENT / EACCES / EMFILE land here. Surface to stderr so the
      // user has a clue what went wrong; downstream awaits will see
      // the exit via raceWithExit and return cleanly.
      stderr.write(`[engine] Go process spawn/runtime error: ${err.message}\n`);
      settle({ code: 2, reason: 'error' });
    });
    child.once('exit', (code, signal) => {
      const c =
        code != null
          ? code
          : signal != null
            ? (SIGNAL_EXIT_CODES[signal] ?? 128)
            : 1;
      settle({ code: c, reason: 'exit' });
    });
  });

  // Race helper: returns a discriminated union so callers can branch
  // on whether their task completed or the child dropped out first.
  // No throw, no unhandled rejection — childExit only resolves.
  type RaceResult<T> =
    | { kind: 'task'; value: T }
    | { kind: 'exit'; state: ChildExit };
  const raceWithExit = async <T>(task: Promise<T>): Promise<RaceResult<T>> =>
    Promise.race<RaceResult<T>>([
      task.then((value) => ({ kind: 'task' as const, value })),
      childExit.then((state) => ({ kind: 'exit' as const, state })),
    ]);

  // ── Signal forwarding (process-scoped) ──────────────────────────
  // Without our own handler, Node's default SIGINT action is to
  // _exit(130) immediately — the Go child is left to discover the
  // disconnect via its stdin EOF, which a deadlocked child can't
  // observe. By intercepting the signal here and calling safeKillGo,
  // we force the Go child down via SIGTERM (and SIGKILL after grace)
  // BEFORE the Node parent exits.
  //
  // `removeSignalHandlers()` is called open-coded at every return path
  // (not via finally — `runEngine` predates a top-level try/finally
  // refactor) so a long-lived host (test runner, library consumer)
  // doesn't accumulate handlers across calls. See E1 / E4 / E5 paths.
  //
  // Platform notes:
  //   - SIGINT: fired by Ctrl-C on POSIX, Ctrl-C / Ctrl-Break on Windows.
  //   - SIGTERM: standard supervisor shutdown signal on POSIX; on
  //     Windows the listener is allowed but never naturally fires
  //     (registering is therefore a no-op there — not an error).
  //   - SIGHUP: controlling terminal disconnect on POSIX; on Windows
  //     Node fires it when the console window is closed (per Node
  //     docs, the listener IS observable on Windows for this case).
  // Validate child stdio BEFORE installing process-level signal
  // listeners — a thrown error past listener install would leak the
  // listeners on `process` for the life of the Node process (test
  // runners / library hosts accumulate handlers across calls).
  if (!child.stdin || !child.stdout) {
    safeKillGo(child);
    throw new Error('engine: Go child process missing stdin/stdout');
  }
  const ipc = new IpcClient(child.stdout, child.stdin);

  const onSignal = (sig: NodeJS.Signals) => {
    stderr.write(`[engine] received ${sig}, terminating Go child\n`);
    safeKillGo(child);
  };
  process.on('SIGINT', onSignal);
  process.on('SIGTERM', onSignal);
  process.on('SIGHUP', onSignal);
  const removeSignalHandlers = () => {
    process.off('SIGINT', onSignal);
    process.off('SIGTERM', onSignal);
    process.off('SIGHUP', onSignal);
  };

  // workerCount=0 when no config declares any plugin — keeps the IPC
  // handshake uniform across all CLI runs without paying the
  // worker_threads startup cost when no plugin code will ever execute.
  const noPlugins = opts.workerConfigs.length === 0;
  const workerCount = noPlugins
    ? 0
    : opts.runtime?.singleThreaded
      ? 1
      : undefined;
  const pool = new WorkerPool({
    configs: opts.workerConfigs,
    workerCount,
    onLog: (rec) => {
      stderr.write(`[runner-log:${rec.source}:${rec.level}] ${rec.text}\n`);
    },
  });

  // ── 2. wire reverse-RPC handlers BEFORE start (so the first frame
  //      Go writes — which could be lintEslintPlugin or output — is
  //      already routable). ─────────────────────────────────────────
  // Build a `configDirSet` from the worker configs. Each lint batch
  // arrives with `file.configKey` set to the directory of the config
  // that owns the file; the workers themselves use that key to pick
  // the right `LoadedPlugins` from their per-config map. The set here
  // is purely for surfacing "unknown configKey on the wire" warnings
  // before the worker's terser internal-error parseError lands.
  const configDirSet = new Set(
    opts.workerConfigs.map((c) => c.configDirectory),
  );

  ipc.setInboundHandler(async (msg) => {
    switch (msg.kind) {
      case 'lintEslintPlugin':
        return handleLintBatch(pool, configDirSet, msg, stderr);
      case 'shutdown':
        // Acknowledge; teardown happens in step 5 via 'exit' event.
        return { ok: true };
      default:
        throw new Error(`engine: unexpected inbound kind '${msg.kind}'`);
    }
  });

  ipc.registerNotification(
    'output',
    (msg: IpcMessage<{ stream?: string; text?: string }>) => {
      const data = msg.data ?? {};
      if (data.text != null) {
        stdout.write(data.text);
      }
    },
  );

  // ── 3. start ipc reading + init worker pool ─────────────────────
  // ipc.start() is sync (just attaches stream listeners); the slow part
  // is pool.init() spawning worker_threads and importing the user's
  // plugins. Both must succeed before we can send `init` — pool.init()
  // failing means the runner can't service plugin lint requests, so
  // bringing the Go child up just to immediately tear it back down is
  // wasted work.
  ipc.start();
  let poolReady = false;
  {
    let outcome: RaceResult<void>;
    try {
      outcome = await raceWithExit(pool.init());
    } catch (err) {
      // raceWithExit itself doesn't throw, but pool.init() can. The
      // race wrapper rejects with the underlying error in that case.
      stderr.write(
        `[engine] worker pool init failed: ${err instanceof Error ? err.message : String(err)}\n`,
      );
      safeKillGo(child);
      await childExit; // reap so we don't return before Node tears the child down
      removeSignalHandlers();
      return { exitCode: 2 };
    }
    if (outcome.kind === 'exit') {
      // Child died before the pool finished initializing. pool.init()
      // may have already spawned worker_threads that are now orphaned;
      // drain them so the Node process can exit cleanly. The pool may
      // still be partially initialized at this point — shutdown() is
      // tolerant of that (workerCount=0 fast-path or partial drain).
      await pool.shutdown().catch(() => undefined);
      removeSignalHandlers();
      return { exitCode: outcome.state.code };
    }
    poolReady = true;
  }

  // ── 4. send init message to Go and await its response ──────────
  // The Go side only reads `forceColor` and `singleThreaded` from the
  // runtime block — every other runtime knob is JS-local. Filter
  // explicitly so adding a new JS-only knob can never accidentally
  // leak into the wire protocol.
  //
  // Spread ordering: `extraInit` goes FIRST so the four authoritative
  // fields below (configs / eslintPluginEntries / runtime) cannot be
  // overridden by a buggy or malicious caller. Earlier ordering let
  // extraInit clobber `configs`; the call sites today never do this,
  // but the type system can't enforce it, so we enforce it positionally.
  {
    let outcome: RaceResult<IpcMessage<unknown>>;
    try {
      outcome = await raceWithExit(
        ipc.sendRequest<{
          configs: unknown[];
          eslintPluginEntries: EslintPluginEntry[];
          runtime: { forceColor?: boolean; singleThreaded?: boolean };
        }>('init', {
          ...(opts.extraInit ?? {}),
          configs: opts.configs,
          eslintPluginEntries: opts.eslintPluginEntries,
          runtime: {
            forceColor: opts.runtime?.forceColor,
            singleThreaded: opts.runtime?.singleThreaded,
          },
        }),
      );
    } catch (err) {
      stderr.write(
        `[engine] init failed: ${err instanceof Error ? err.message : String(err)}\n`,
      );
      if (poolReady) await pool.shutdown().catch(() => undefined);
      safeKillGo(child);
      await childExit;
      removeSignalHandlers();
      return { exitCode: 2 };
    }
    if (outcome.kind === 'exit') {
      // Go process disappeared between us sending init and receiving
      // the response. Pool is still ours to drain.
      if (poolReady) await pool.shutdown().catch(() => undefined);
      removeSignalHandlers();
      return { exitCode: outcome.state.code };
    }
    const ok = (outcome.value.data as { ok?: boolean })?.ok;
    if (ok !== true) {
      stderr.write(
        `[engine] Go rejected init: ${JSON.stringify(outcome.value.data)}\n`,
      );
      // .catch swallow matches the other pool.shutdown() call sites in
      // this function. Without it, a transient worker.terminate
      // rejection here would surface as a generic "rslint: Error: ..."
      // exit instead of the meaningful exitCode 2 below.
      await pool.shutdown().catch(() => undefined);
      safeKillGo(child);
      await childExit;
      removeSignalHandlers();
      return { exitCode: 2 };
    }
  }

  // ── 5. wait for Go to exit. ipc dispatch + output forwarding
  //      happen in handlers wired above. The error/exit watchers were
  //      attached right after spawn, so this just awaits the same
  //      promise — which may already have resolved if the child
  //      finished while we were processing init. ────────────────────
  const finalExit = await childExit;

  // ── 6. drain workers and resolve ───────────────────────────────
  ipc.close();
  await pool.shutdown().catch(() => undefined);
  removeSignalHandlers();

  return { exitCode: finalExit.code };
}

// ─── inbound handlers ─────────────────────────────────────────────

async function handleLintBatch(
  pool: WorkerPool,
  configDirSet: ReadonlySet<string>,
  msg: IpcMessage<unknown>,
  // stderr is taken from the runEngine caller so injected sinks (test
  // runners that pass a mock stderr) actually capture the unknown-
  // configKey warning. The previous code wrote straight to
  // `process.stderr`, bypassing injection and breaking assertions
  // that snapshot the warning.
  stderr: NodeJS.WritableStream,
): Promise<CompatBatchResult> {
  // Schema (mirrors internal/linter.CompatBatch):
  //   { files: [{path,text,languageOptions?,settings?,configKey?}],
  //     rules: { "<rule>": { options: any[] } },
  //     collectFixes: bool, suggestionsMode: 'off'|'eager' }
  const data = msg.data as Partial<CompatBatchInput> | null | undefined;
  if (!data || !Array.isArray(data.files)) {
    throw new Error('lintEslintPlugin: malformed payload (no files[])');
  }

  // Per-file routing + LintTask building is shared with the LSP host
  // (vscode-extension/src/CompatPool.ts) via `buildCompatTasksByConfigKey`.
  const tasks = buildCompatTasksByConfigKey(data as CompatBatchInput, {
    configDirSet,
    onUnknownConfigKey: (filePath, configKey) => {
      stderr.write(
        `[engine] file ${filePath} carries unknown configKey ${JSON.stringify(
          configKey,
        )}; eslint-plugin rules will not run on it\n`,
      );
    },
  });

  const results = await pool.lintBatch(tasks);
  // Shared result projection so both host paths hand Go an
  // identical, byte-stable wire shape.
  return buildCompatBatchResult(results);
}

// ─── helpers ─────────────────────────────────────────────────────

/**
 * Best-effort, escalating kill: SIGTERM first (gives the binary a
 * window to flush stderr / release file handles), then SIGKILL after
 * a short grace period if the child is still alive.
 *
 * SIGTERM alone is not sufficient on POSIX: a wedged Go binary (stuck
 * in a syscall, deadlocked goroutine ring, or — pathologically — a
 * malicious signal handler) would ignore it and engine.ts would
 * return while the child kept running. The grace period is also long
 * enough that a cooperating binary that's mid-flush won't be forcibly
 * killed.
 *
 * **Windows note**: `child.kill('SIGTERM')` on Windows is equivalent
 * to `TerminateProcess` — the signal argument is ignored and the
 * child is killed forcefully and abruptly. The escalation timer below
 * therefore never fires on Windows (child.exitCode is non-null by the
 * time it would). This means a Go binary on Windows has no chance to
 * flush profiling output or send a clean IPC shutdown — but it's
 * killed reliably, which is the contract this helper guarantees.
 *
 * Grace window: 5 seconds. The CLI's shutdown path normally completes
 * far faster than that (sub-second drain + close), so the SIGKILL
 * fallback only fires when something is wrong.
 */
const KILL_GRACE_MS = 5_000;

function safeKillGo(child: ChildProcess): void {
  try {
    if (child.killed) return;
    child.kill('SIGTERM');
  } catch {
    /* SIGTERM dispatch itself failed (race with exit, OS error) */
    return;
  }

  // Schedule SIGKILL as a backstop. `child.exitCode` flips non-null on
  // exit; if the cooperating SIGTERM landed the process before the
  // grace expires, the timer no-ops.
  const killTimer = setTimeout(() => {
    if (child.exitCode == null && child.signalCode == null) {
      try {
        child.kill('SIGKILL');
      } catch {
        /* ignore — process likely just exited */
      }
    }
  }, KILL_GRACE_MS);

  // Don't keep the Node event loop alive just for this timer.
  if (typeof killTimer.unref === 'function') {
    killTimer.unref();
  }

  // Clean up the timer once the child exits cooperatively.
  child.once('exit', () => {
    clearTimeout(killTimer);
  });
}
