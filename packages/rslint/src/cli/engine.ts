/**
 * Engine: Node-side host for the Go↔Node IPC CLI handshake.
 *
 * Spawns the Go binary (which defaults to runCLI / IPC mode), drives the
 * `init` handshake, forwards the Go child's `output` frames to stdout, and
 * acks `shutdown`. The Go child does all the linting; this host just owns the
 * pipe and the protocol.
 *
 * This is the comm host for native lint — it owns the pipe and the protocol
 * while the Go child does all the linting.
 *
 * Exit codes propagate from the Go child (or 2 on a host-level failure).
 */
import { spawn, type ChildProcess } from 'node:child_process';
import { IpcClient } from '../ipc/index.js';
import type { IpcMessage } from '../ipc/index.js';
import {
  ConfigModuleHost,
  type ActivateConfigsRequest,
  type LoadConfigsRequest,
} from '../config/config-loader.js';

interface PluginLintHost {
  lint(req: unknown): Promise<unknown>;
  shutdown(): Promise<void>;
}

type CreatePluginLintHost = (
  configs: Array<{ configPath: string; configDirectory: string }>,
  onLog?: (rec: { level: string; source: string; text: string }) => void,
  singleThreaded?: boolean,
) => Promise<PluginLintHost>;

// POSIX: a process killed by signal N exits 128+N. Node reports the signal
// NAME, so map the ones we can receive; collapsing all to 130 would mislabel
// a SIGTERM/SIGKILL teardown as a Ctrl-C.
const SIGNAL_EXIT_CODES: Record<string, number> = {
  SIGHUP: 129,
  SIGINT: 130,
  SIGQUIT: 131,
  SIGTERM: 143,
  SIGKILL: 137,
};

export interface EngineRunOptions {
  /** Path to the Go rslint binary. */
  binPath: string;
  /** Args forwarded to the Go binary (user CLI flags, --start-time, files). */
  goArgs: string[];
  /**
   * Configs sent in the `init` payload (JS/TS-config entries, each shaped
   * `{configDirectory, configPath?, entries}`). Empty means "no JS config —
   * let Go load JSON config from disk itself".
   */
  configs: unknown[];
  /** Working directory (defaults to process.cwd()). */
  cwd?: string;
  /** Runtime hints forwarded in the `init` payload's `runtime` block. */
  runtime?: { singleThreaded?: boolean };
  /**
   * Extra fields merged into the `init` payload (positional files, --format,
   * --fix, workingDirectory). Pass-through so the engine stays unaware of CLI
   * flag layout.
   */
  extraInit?: Record<string, unknown>;
  /** stdout sink (default process.stdout). Lets tests capture output. */
  stdout?: NodeJS.WritableStream;
  /** stderr sink (default process.stderr). */
  stderr?: NodeJS.WritableStream;
  /**
   * ESLint-plugin {prefix, ruleNames} metadata forwarded to Go in the
   * `init` payload so it registers placeholder rules for plugin rule names.
   */
  eslintPluginEntries?: Array<{ prefix: string; ruleNames: string[] }>;
  /**
   * Per-config descriptors (configPath + configDirectory) for configs that
   * mount ESLint plugins; used to init the worker pool that answers Go's
   * reverse `pluginLint` requests. Empty ⇒ no pool, zero overhead.
   */
  pluginConfigs?: Array<{ configPath: string; configDirectory: string }>;
  /** @internal Dependency seam for activation-race and cleanup tests. */
  createPluginLintHost?: CreatePluginLintHost;
  /** @internal Dependency seam for post-prepare lifecycle tests. */
  configModuleHost?: ConfigModuleHost;
}

export async function runEngine(opts: EngineRunOptions): Promise<number> {
  const stdout = opts.stdout ?? process.stdout;
  const stderr = opts.stderr ?? process.stderr;
  // TTY fact for the Go side's color decision, probed on the actual sink the
  // forwarded output frames land on. Only this process can observe it — the
  // Go child's own stdout is the IPC pipe. Non-TTY sinks (pipes, files, test
  // streams) have no isTTY property, which coerces to false here.
  // rslint-disable-next-line @typescript-eslint/no-unsafe-type-assertion
  const stdoutIsTTY = (stdout as Partial<NodeJS.WriteStream>).isTTY === true;

  const child = spawn(opts.binPath, opts.goArgs, {
    stdio: ['pipe', 'pipe', 'inherit'],
    cwd: opts.cwd ?? process.cwd(),
  });

  // childExit always RESOLVES (never rejects); awaits race against it so a
  // child that drops out mid-handshake unwinds cleanly instead of hanging.
  // Attached BEFORE any await so an immediate spawn error (ENOENT/EACCES) or
  // instant exit is observed rather than crashing the process as 'unhandled'.
  type ChildExit = { code: number };
  const childExit = new Promise<ChildExit>((resolve) => {
    let resolved = false;
    const settle = (code: number) => {
      if (!resolved) {
        resolved = true;
        resolve({ code });
      }
    };
    child.once('error', (err) => {
      stderr.write(`rslint: Go spawn/runtime error: ${err.message}\n`);
      settle(2);
    });
    child.once('exit', (code, signal) => {
      settle(
        code != null
          ? code
          : signal != null
            ? (SIGNAL_EXIT_CODES[signal] ?? 128)
            : 1,
      );
    });
  });

  type RaceResult<T> =
    | { kind: 'task'; value: T }
    | { kind: 'exit'; state: ChildExit };
  const raceWithExit = async <T>(task: Promise<T>): Promise<RaceResult<T>> =>
    Promise.race<RaceResult<T>>([
      task.then((value) => ({ kind: 'task' as const, value })),
      childExit.then((state) => ({ kind: 'exit' as const, state })),
    ]);

  // Validate stdio BEFORE installing process-level signal listeners, so a throw
  // here can't leak listeners on `process` for a long-lived host.
  if (!child.stdin || !child.stdout) {
    safeKillGo(child);
    throw new Error('engine: Go child process missing stdin/stdout');
  }
  const ipc = new IpcClient(child.stdout, child.stdin);

  // Host the ESLint-plugin worker pool that answers Go's reverse
  // `pluginLint` requests. Loaded via a runtime dynamic import: the
  // `: string` path type stops the library dts build (which excludes
  // src/eslint-plugin/**) from type-checking the worker module, and
  // `webpackIgnore` keeps rslib from bundling it into the engine chunk — it
  // must stay a sibling so the worker's `import.meta.url` resolution finds
  // lint-worker.js. Resolves at runtime to dist/eslint-plugin/index.js.
  let pluginHost: PluginLintHost | null = null;
  const configModuleHost = opts.configModuleHost ?? new ConfigModuleHost();
  const configTransactions = new Set<string>();
  let pluginHostInitialization: Promise<void> | null = null;
  let shuttingDown = false;
  const pendingPluginHostBuilds = new Set<Promise<PluginLintHost | null>>();
  const stagedPluginHosts = new Set<PluginLintHost>();
  const pluginHostShutdowns = new WeakMap<PluginLintHost, Promise<void>>();

  const shutdownPluginHost = (host: PluginLintHost | null): Promise<void> => {
    if (!host) return Promise.resolve();
    stagedPluginHosts.delete(host);
    let shutdown = pluginHostShutdowns.get(host);
    if (!shutdown) {
      shutdown = host.shutdown();
      pluginHostShutdowns.set(host, shutdown);
    }
    return shutdown;
  };

  const publishPluginHost = async (
    host: PluginLintHost | null,
  ): Promise<void> => {
    if (!host) return;
    if (shuttingDown) {
      await shutdownPluginHost(host).catch(() => undefined);
      return;
    }
    stagedPluginHosts.delete(host);
    pluginHost = host;
  };

  const buildPluginHost = (
    pluginConfigs: Array<{ configPath: string; configDirectory: string }>,
  ): Promise<PluginLintHost | null> => {
    const build = (async () => {
      if (pluginConfigs.length === 0 || shuttingDown) return null;
      let createPluginLintHost = opts.createPluginLintHost;
      if (!createPluginLintHost) {
        const pluginEntry: string = './eslint-plugin/index.js';
        const mod: { createPluginLintHost: CreatePluginLintHost } =
          await import(/* webpackIgnore: true */ pluginEntry);
        createPluginLintHost = mod.createPluginLintHost;
      }
      const host = await createPluginLintHost(
        pluginConfigs,
        (rec) => stderr.write(`[rslint:plugin] ${rec.text}\n`),
        opts.runtime?.singleThreaded,
      );
      if (shuttingDown) {
        await shutdownPluginHost(host).catch(() => undefined);
        return null;
      }
      stagedPluginHosts.add(host);
      return host;
    })();
    pendingPluginHostBuilds.add(build);
    void build.then(
      () => pendingPluginHostBuilds.delete(build),
      () => pendingPluginHostBuilds.delete(build),
    );
    return build;
  };

  const ensurePluginHost = async (): Promise<void> => {
    if (pluginHost) return;
    pluginHostInitialization ??= (async () => {
      const pluginConfigs = [...(opts.pluginConfigs ?? [])];
      const createdHost = await buildPluginHost(pluginConfigs);
      await publishPluginHost(createdHost);
    })();
    try {
      await pluginHostInitialization;
    } catch (err) {
      pluginHostInitialization = null;
      throw err;
    }
  };

  // Preserve the legacy preloaded-config path while the reverse-discovery
  // protocol rolls out. New runs initialize lazily after Go has selected its
  // effective catalog, avoiding plugin workers for pruned configs.
  if ((opts.pluginConfigs?.length ?? 0) > 0) {
    try {
      await ensurePluginHost();
    } catch (err) {
      const msg = err instanceof Error ? err.message : String(err);
      stderr.write(`rslint: failed to start ESLint-plugin worker: ${msg}\n`);
      safeKillGo(child);
      return 2;
    }
  }

  // Without our own SIGINT handler Node's default action _exit(130)s
  // immediately, leaving the Go child to discover the disconnect via stdin EOF
  // — which a wedged child can't. Intercept and force it down first.
  const onSignal = () => {
    // No log: a user Ctrl-C (SIGINT) or a normal SIGTERM/SIGHUP teardown is the
    // expected path, not an error — forward the kill to the Go child and tear
    // the worker pool down so its threads don't outlive us.
    shuttingDown = true;
    safeKillGo(child);
    void Promise.allSettled([
      shutdownPluginHost(pluginHost),
      ...[...stagedPluginHosts].map(shutdownPluginHost),
    ]);
  };
  process.on('SIGINT', onSignal);
  process.on('SIGTERM', onSignal);
  process.on('SIGHUP', onSignal);
  const removeSignalHandlers = () => {
    process.off('SIGINT', onSignal);
    process.off('SIGTERM', onSignal);
    process.off('SIGHUP', onSignal);
  };

  // try/finally so the worker pool is always drained — every exit path
  // (init failure as well as normal completion) runs the `finally`, which
  // tears down the signal handlers and the plugin worker pool. Without this
  // the init-failure returns below leaked a graceful drain (the pool's
  // threads were only reaped by the outer process.exit), whereas the normal
  // and signal paths drained it cleanly. `pluginHost?.shutdown()` is
  // null-safe (no pool ⇒ no-op) and idempotent (WorkerPool.shutdown guards
  // on `closed`), so the signal handler firing shutdown first is harmless.
  try {
    // Wire handlers BEFORE start so the first frame Go writes is routable.
    ipc.setInboundHandler(async (msg) => {
      switch (msg.kind) {
        case 'loadConfigs': {
          const request = msg.data as LoadConfigsRequest;
          const response = await configModuleHost.loadConfigs(request);
          configTransactions.add(request.transactionId);
          return response;
        }
        case 'activateConfigs': {
          const request = msg.data as ActivateConfigsRequest;
          let stagedPluginHost: PluginLintHost | null = null;
          try {
            const response = await configModuleHost.activateConfigs(
              request,
              undefined,
              async (activation) => {
                if (pluginHost) return;
                const createdHost = await buildPluginHost(
                  activation.pluginConfigs,
                );
                if (shuttingDown) {
                  await shutdownPluginHost(createdHost).catch(() => undefined);
                  return;
                }
                stagedPluginHost = createdHost;
              },
            );
            // Do not expose a worker that re-imported the config until the
            // post-prepare fingerprint check has accepted the activation.
            const preparedHost = stagedPluginHost as PluginLintHost | null;
            await publishPluginHost(preparedHost);
            stagedPluginHost = null;
            return response;
          } catch (error) {
            await shutdownPluginHost(
              stagedPluginHost as PluginLintHost | null,
            ).catch(() => undefined);
            throw error;
          }
        }
        case 'shutdown':
          // Go signals it's done; teardown happens via the 'exit' event below.
          return { ok: true };
        case 'pluginLint':
          // Go dispatched a plugin-lint batch in parallel with native linting.
          // Answer from the worker pool; an absent pool (no plugins configured)
          // yields empty results. Never throws — per-file/-rule failures travel
          // inside the result payload.
          await ensurePluginHost();
          return pluginHost ? pluginHost.lint(msg.data) : { results: [] };
        default:
          throw new Error(`engine: unexpected inbound kind '${msg.kind}'`);
      }
    });
    ipc.registerNotification(
      'output',
      (msg: IpcMessage<{ stream?: string; text?: string }>) => {
        const text = msg.data?.text;
        if (text != null) stdout.write(text);
      },
    );
    ipc.start();

    // Send `init` and await Go's ack, racing the child dropping out.
    {
      let outcome: RaceResult<IpcMessage<unknown>>;
      try {
        outcome = await raceWithExit(
          ipc.sendRequest('init', {
            ...(opts.extraInit ?? {}),
            configs: opts.configs,
            eslintPlugins: opts.eslintPluginEntries,
            runtime: {
              stdoutIsTTY,
              singleThreaded: opts.runtime?.singleThreaded,
            },
          }),
        );
      } catch {
        // The init request can reject when the Go child exits cleanly before we
        // read its init ack: a fast path (--help / --init) closes the pipe the
        // moment its work is done, and the stdout-EOF seal that rejects the
        // pending request can beat the child 'exit' event into the race above
        // (observed on Linux, where stdout 'end' tends to precede 'exit'). The
        // child's exit code is the source of truth — a clean (0) exit means Go
        // finished its job, so honor it; only a non-zero exit is a real failure.
        safeKillGo(child);
        const state = await childExit;
        if (state.code === 0) return 0;
        stderr.write(`rslint: init failed (Go exited ${state.code})\n`);
        return state.code;
      }
      if (outcome.kind === 'exit') {
        return outcome.state.code;
      }
      const data = outcome.value.data;
      const ok =
        typeof data === 'object' &&
        data !== null &&
        'ok' in data &&
        data.ok === true;
      if (!ok) {
        stderr.write(
          `rslint: Go rejected init: ${JSON.stringify(outcome.value.data)}\n`,
        );
        safeKillGo(child);
        await childExit;
        return 2;
      }
    }

    // Output forwarding + shutdown ack happen in the handlers above; just wait
    // for the Go child to finish (it exits with its own lint exit code).
    const finalExit = await childExit;
    ipc.close();
    return finalExit.code;
  } finally {
    // Single cleanup site for every return above: drop the process-level
    // signal listeners and drain the plugin worker pool. Both are no-ops
    // when already done (removeSignalHandlers off()s detached listeners;
    // pluginHost?.shutdown() is null-safe + idempotent), so this is safe
    // even on the normal path and after the signal handler already fired.
    removeSignalHandlers();
    shuttingDown = true;
    await Promise.allSettled([...pendingPluginHostBuilds]);
    await Promise.allSettled([
      shutdownPluginHost(pluginHost as PluginLintHost | null),
      ...[...stagedPluginHosts].map(shutdownPluginHost),
    ]);
    for (const transactionId of configTransactions) {
      configModuleHost.deleteSession(transactionId);
    }
  }
}

// ─── helpers ─────────────────────────────────────────────────────────

const KILL_GRACE_MS = 5_000;

/**
 * Escalating kill: SIGTERM first (lets the binary flush stderr / send a clean
 * IPC shutdown), then SIGKILL after a grace period if it's still alive. The
 * timer is unref'd so it never keeps the Node event loop alive on its own.
 */
function safeKillGo(child: ChildProcess): void {
  try {
    if (child.killed) return;
    child.kill('SIGTERM');
  } catch {
    return;
  }
  const killTimer = setTimeout(() => {
    if (child.exitCode == null && child.signalCode == null) {
      try {
        child.kill('SIGKILL');
      } catch {
        /* ignore — process likely just exited */
      }
    }
  }, KILL_GRACE_MS);
  if (typeof killTimer.unref === 'function') killTimer.unref();
  child.once('exit', () => {
    clearTimeout(killTimer);
  });
}
