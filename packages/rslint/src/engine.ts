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
import { IpcClient } from './ipc/index.js';
import type { IpcMessage } from './ipc/index.js';

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
  runtime?: { forceColor?: boolean; singleThreaded?: boolean };
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
}

export async function runEngine(opts: EngineRunOptions): Promise<number> {
  const stdout = opts.stdout ?? process.stdout;
  const stderr = opts.stderr ?? process.stderr;

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

  // Without our own SIGINT handler Node's default action _exit(130)s
  // immediately, leaving the Go child to discover the disconnect via stdin EOF
  // — which a wedged child can't. Intercept and force it down first.
  const onSignal = () => {
    // No log: a user Ctrl-C (SIGINT) or a normal SIGTERM/SIGHUP teardown is the
    // expected path, not an error — just forward the kill to the Go child.
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

  // Wire handlers BEFORE start so the first frame Go writes is routable.
  ipc.setInboundHandler((msg) => {
    switch (msg.kind) {
      case 'shutdown':
        // Go signals it's done; teardown happens via the 'exit' event below.
        return { ok: true };
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
          runtime: {
            forceColor: opts.runtime?.forceColor,
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
      removeSignalHandlers();
      if (state.code === 0) return 0;
      stderr.write(`rslint: init failed (Go exited ${state.code})\n`);
      return state.code;
    }
    if (outcome.kind === 'exit') {
      removeSignalHandlers();
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
      removeSignalHandlers();
      return 2;
    }
  }

  // Output forwarding + shutdown ack happen in the handlers above; just wait
  // for the Go child to finish (it exits with its own lint exit code).
  const finalExit = await childExit;
  ipc.close();
  removeSignalHandlers();
  return finalExit.code;
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
