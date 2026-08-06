/**
 * Core-owned editor runtime.
 *
 * stdin/stdout are a transparent, standards-compliant LSP stream between the
 * editor and the Go child. Config evaluation and ESLint-plugin workers live in
 * this Node process and talk to Go over an authenticated private loopback
 * stream using the existing framed protocol.
 */
import { spawn } from 'node:child_process';
import { randomBytes, timingSafeEqual } from 'node:crypto';
import { existsSync, realpathSync } from 'node:fs';
import {
  createServer,
  type AddressInfo,
  type Server,
  type Socket,
} from 'node:net';
import path from 'node:path';
import process from 'node:process';
import { Writable } from 'node:stream';
import { fileURLToPath } from 'node:url';

import {
  CONFIG_DISCOVERY_PROTOCOL_VERSION,
  ConfigModuleHost,
  type ActivateConfigsRequest,
  type LoadConfigsRequest,
} from '../config/config-loader.js';
import { resolveRslintBinary } from '../internal/resolve-binary.js';
import { IpcClient, type IpcMessage } from '../ipc/index.js';
import { EditorPluginPool } from './editor-plugin-pool.js';

interface ActivationRequest extends ActivateConfigsRequest {
  dependencyRevision?: number;
}

interface TransactionControlRequest {
  protocolVersion: typeof CONFIG_DISCOVERY_PROTOCOL_VERSION;
  transactionId: string;
}

const SIGNAL_FORCE_CHILD_EXIT_MS = 500;
const SIGNAL_FORCE_RUNTIME_EXIT_MS = 1_500;
const RUNTIME_CONNECT_TIMEOUT_MS = 10_000;
const RUNTIME_TOKEN_BYTES = 32;
const DEDICATED_RUNTIME_FORCE_EXIT_MS = 1_000;

/**
 * A project config can leave ref'ed timers or other handles behind. Once the
 * native LSP has ended, those handles must not keep a dedicated editor-runtime
 * executable alive (and keep its public stdout transport falsely open).
 * Exported only so the package bin can apply the same executable boundary.
 */
export function scheduleEditorRuntimeProcessExit(code: number): void {
  process.exitCode = code;
  const timer = setTimeout(
    () => process.exit(code),
    DEDICATED_RUNTIME_FORCE_EXIT_MS,
  );
  // A clean runtime with no leaked handles should still exit naturally without
  // waiting for this fallback.
  timer.unref?.();
}

function activePnpPath(cwd: string): string | undefined {
  const injected = process.env.RSLINT_RUNTIME_PNP_PATH;
  if (injected) return injected;
  // The VS Code resolver injects an exact domain. Preserve direct
  // `yarn rslint --lsp` as well: Yarn's require hook marks this process, after
  // which the nearest physical hook is the active domain inherited by Go.
  if (
    (process.versions as Record<string, string | undefined>).pnp === undefined
  )
    return undefined;
  let current = path.resolve(cwd);
  for (;;) {
    for (const name of ['.pnp.cjs', '.pnp.js']) {
      const candidate = path.join(current, name);
      if (existsSync(candidate)) return realpathSync(candidate);
    }
    const parent = path.dirname(current);
    if (parent === current) return undefined;
    current = parent;
  }
}

function asRecord(value: unknown): Record<string, unknown> {
  if (value === null || typeof value !== 'object' || Array.isArray(value)) {
    throw new Error('editor runtime request payload must be an object');
  }
  return value as Record<string, unknown>;
}

function assertControlRequest(value: unknown): TransactionControlRequest {
  const request = asRecord(value);
  if (request.protocolVersion !== CONFIG_DISCOVERY_PROTOCOL_VERSION) {
    throw new Error(
      `unsupported config protocol ${String(request.protocolVersion)}`,
    );
  }
  if (
    typeof request.transactionId !== 'string' ||
    request.transactionId.length === 0
  ) {
    throw new Error('config transactionId must be a non-empty string');
  }
  return request as unknown as TransactionControlRequest;
}

function isLoadRequest(value: unknown): value is LoadConfigsRequest {
  const request = asRecord(value);
  return (
    request.protocolVersion === CONFIG_DISCOVERY_PROTOCOL_VERSION &&
    typeof request.transactionId === 'string' &&
    Array.isArray(request.candidates)
  );
}

function isActivationRequest(value: unknown): value is ActivationRequest {
  const request = asRecord(value);
  return (
    request.protocolVersion === CONFIG_DISCOVERY_PROTOCOL_VERSION &&
    typeof request.transactionId === 'string' &&
    Array.isArray(request.effectiveConfigIds) &&
    (request.dependencyRevision === undefined ||
      (typeof request.dependencyRevision === 'number' &&
        Number.isSafeInteger(request.dependencyRevision) &&
        request.dependencyRevision >= 0))
  );
}

interface RuntimeListener {
  readonly address: string;
  readonly token: string;
  readonly connection: Promise<Socket>;
  close(): Promise<void>;
}

async function closeServer(server: Server): Promise<void> {
  if (!server.listening) return;
  await new Promise<void>((resolve) => {
    server.close(() => resolve());
  });
}

/**
 * Bind a one-shot private channel. Authentication happens before IpcClient sees
 * any bytes, so a local port race cannot inject a framed request or keep the
 * real Go child from connecting.
 */
async function createRuntimeListener(): Promise<RuntimeListener> {
  const token = randomBytes(RUNTIME_TOKEN_BYTES).toString('hex');
  const expectedToken = Buffer.from(token, 'ascii');
  const sockets = new Set<Socket>();
  let accepted = false;
  let resolveConnection!: (socket: Socket) => void;
  let rejectConnection!: (error: Error) => void;
  const connection = new Promise<Socket>((resolve, reject) => {
    resolveConnection = resolve;
    rejectConnection = reject;
  });
  // A rejected accept must always be observed, including a spawn failure that
  // closes the listener before runEditorRuntime reaches its await.
  void connection.catch(() => undefined);

  const server = createServer((socket) => {
    sockets.add(socket);
    socket.once('close', () => sockets.delete(socket));
    socket.on('error', () => undefined);
    if (accepted) {
      socket.destroy();
      return;
    }
    let authentication = Buffer.alloc(0);
    const onData = (chunk: Buffer): void => {
      authentication = Buffer.concat([authentication, chunk]);
      const newline = authentication.indexOf(0x0a);
      if (newline < 0) {
        if (authentication.length > expectedToken.length) socket.destroy();
        return;
      }
      const supplied = authentication.subarray(0, newline);
      const valid =
        supplied.length === expectedToken.length &&
        timingSafeEqual(supplied, expectedToken);
      if (!valid) {
        socket.destroy();
        return;
      }
      accepted = true;
      socket.pause();
      socket.off('data', onData);
      const remainder = authentication.subarray(newline + 1);
      if (remainder.length > 0) socket.unshift(remainder);
      // Idle or failed pre-authentication connections must not remain live for
      // the whole editor session after the real child wins authentication.
      for (const candidate of sockets) {
        if (candidate !== socket) candidate.destroy();
      }
      void closeServer(server);
      resolveConnection(socket);
    };
    socket.on('data', onData);
  });

  await new Promise<void>((resolve, reject) => {
    server.once('error', reject);
    server.listen({ host: '127.0.0.1', port: 0, exclusive: true }, () => {
      server.off('error', reject);
      resolve();
    });
  });
  const bound = server.address() as AddressInfo | null;
  if (!bound) {
    await closeServer(server);
    throw new Error('editor runtime: private listener has no address');
  }
  const timer = setTimeout(() => {
    rejectConnection(
      new Error('editor runtime: Go child did not authenticate in time'),
    );
    for (const socket of sockets) socket.destroy();
    void closeServer(server);
  }, RUNTIME_CONNECT_TIMEOUT_MS);
  timer.unref?.();

  return {
    address: '127.0.0.1:' + String(bound.port),
    token,
    connection: connection.finally(() => clearTimeout(timer)),
    async close() {
      clearTimeout(timer);
      for (const socket of sockets) socket.destroy();
      await closeServer(server);
    },
  };
}

export interface EditorRuntimeOptions {
  readonly cwd?: string;
  readonly binaryPath?: string;
  readonly stdin?: NodeJS.ReadableStream;
  readonly stdout?: NodeJS.WritableStream;
  readonly stderr?: NodeJS.WritableStream;
}

/** Run one editor sidecar until its Go LSP child exits. */
export async function runEditorRuntime(
  options: EditorRuntimeOptions = {},
): Promise<number> {
  const stdin = options.stdin ?? process.stdin;
  const stdout = options.stdout ?? process.stdout;
  const stderr = options.stderr ?? process.stderr;
  const cwd = options.cwd ?? process.cwd();
  // Resolve before opening the listener. A missing optional native package
  // must fail immediately instead of leaving a live server handle that delays
  // sidecar exit until the authentication timeout.
  const binaryPath = options.binaryPath ?? resolveRslintBinary();
  const configHost = new ConfigModuleHost();
  const pluginPool = new EditorPluginPool();
  const transactions = new Set<string>();
  const listener = await createRuntimeListener();

  const childEnv: NodeJS.ProcessEnv = {
    ...process.env,
    RSLINT_RUNTIME_IPC_ADDRESS: listener.address,
    RSLINT_RUNTIME_IPC_TOKEN: listener.token,
  };
  const pnpPath = activePnpPath(cwd);
  if (pnpPath) childEnv.RSLINT_RUNTIME_PNP_PATH = pnpPath;
  else delete childEnv.RSLINT_RUNTIME_PNP_PATH;

  const child = spawn(binaryPath, ['--lsp', '--runtime-ipc'], {
    cwd,
    env: childEnv,
    stdio: ['pipe', 'pipe', 'pipe'],
  });

  if (!child.stdin || !child.stdout || !child.stderr) {
    child.kill('SIGKILL');
    await listener.close();
    await pluginPool.dispose();
    throw new Error('editor runtime: Go child is missing LSP pipes');
  }
  child.stderr.pipe(stderr, { end: false });

  type ChildOutcome = {
    readonly code: number | null;
    readonly signal: NodeJS.Signals | null;
    readonly error?: Error;
  };
  const childExit = new Promise<ChildOutcome>((resolve) => {
    let settled = false;
    const finish = (outcome: ChildOutcome) => {
      if (settled) return;
      settled = true;
      resolve(outcome);
    };
    child.once('error', (error) => finish({ code: null, signal: null, error }));
    child.once('exit', (code, signal) => finish({ code, signal }));
  });

  let signalExitCode: number | undefined;
  let forceChildExitTimer: ReturnType<typeof setTimeout> | undefined;
  let forceRuntimeExitTimer: ReturnType<typeof setTimeout> | undefined;
  const terminateChild = (signal: NodeJS.Signals): void => {
    child.kill(signal);
    forceChildExitTimer ??= setTimeout(() => {
      if (child.exitCode === null && child.signalCode === null) {
        child.kill('SIGKILL');
      }
    }, SIGNAL_FORCE_CHILD_EXIT_MS);
    forceChildExitTimer.unref?.();
  };
  const forwardSignal = (signal: NodeJS.Signals, exitCode: number): void => {
    signalExitCode ??= exitCode;
    // The VS Code process owns this sidecar, while the sidecar owns Go. Every
    // termination path must reap Go before the outer owner may force-kill Node;
    // otherwise a transport failure could strand an uncooperative descendant.
    terminateChild(signal);
    // A final process bound still protects direct `rslint --lsp` users if the
    // OS never reports the child exit after the SIGKILL attempt.
    forceRuntimeExitTimer ??= setTimeout(
      () => process.exit(signalExitCode ?? exitCode),
      SIGNAL_FORCE_RUNTIME_EXIT_MS,
    );
    forceRuntimeExitTimer.unref?.();
  };
  const onSigint = () => forwardSignal('SIGINT', 130);
  const onSigterm = () => forwardSignal('SIGTERM', 143);
  const onSighup = () => forwardSignal('SIGHUP', 129);
  const removeSignalHandlers = (): void => {
    clearTimeout(forceChildExitTimer);
    clearTimeout(forceRuntimeExitTimer);
    process.off('SIGINT', onSigint);
    process.off('SIGTERM', onSigterm);
    process.off('SIGHUP', onSighup);
  };
  // Install these before waiting for private authentication. Otherwise a
  // shutdown in the spawn/connect window can orphan the native child.
  process.on('SIGINT', onSigint);
  process.on('SIGTERM', onSigterm);
  process.on('SIGHUP', onSighup);

  let runtimeSocket: Socket;
  try {
    runtimeSocket = await Promise.race([
      listener.connection,
      childExit.then((outcome) => {
        throw (
          outcome.error ??
          new Error(
            'Go language server exited before private IPC connected ' +
              '(code=' +
              String(outcome.code) +
              ', signal=' +
              String(outcome.signal) +
              ')',
          )
        );
      }),
    ]);
  } catch (error) {
    child.kill('SIGKILL');
    removeSignalHandlers();
    await listener.close();
    await pluginPool.dispose();
    if (signalExitCode !== undefined) return signalExitCode;
    throw error;
  }

  const ipc = new IpcClient(runtimeSocket, runtimeSocket, {
    onTransportFailure(error) {
      const terminateGeneration = (): void => {
        if (child.exitCode !== null || child.signalCode !== null) return;
        stderr.write(
          `rslint: private editor-runtime transport failed: ${error.message}\n`,
        );
        // A logical decoder failure must become physical EOF for Go. Otherwise
        // the authenticated peer can remain alive forever after Node stops
        // reading its requests, and the language client never gets a restart.
        runtimeSocket.destroy();
        terminateChild('SIGTERM');
      };
      if (error.message === 'IpcClient: peer closed input stream') {
        // Go closes its runtime channel immediately before a normal process
        // exit. Give that clean path a short chance to publish childExit so a
        // successful LSP shutdown is not raced into SIGTERM by socket EOF.
        const timer = setTimeout(terminateGeneration, 250);
        timer.unref?.();
        void childExit.then(() => clearTimeout(timer));
        return;
      }
      terminateGeneration();
    },
  });
  ipc.setInboundHandler(async (message, signal) => {
    switch (message.kind) {
      case 'loadConfigs': {
        if (!isLoadRequest(message.data)) {
          throw new Error('editor runtime: invalid loadConfigs request');
        }
        const request = { ...message.data, loadMode: 'fresh' as const };
        transactions.add(request.transactionId);
        try {
          return await configHost.loadConfigs(request, signal);
        } catch (error) {
          transactions.delete(request.transactionId);
          configHost.deleteSession(request.transactionId);
          throw error;
        }
      }
      case 'activateConfigs': {
        if (!isActivationRequest(message.data)) {
          throw new Error('editor runtime: invalid activateConfigs request');
        }
        const request = message.data;
        let pluginHostReady = false;
        try {
          const activation = await configHost.activateConfigs(
            request,
            signal,
            async (plan) => {
              pluginHostReady = await pluginPool.prepare(
                plan,
                request.dependencyRevision ?? 0,
              );
            },
          );
          return {
            transactionId: activation.transactionId,
            eslintPluginEntries: pluginHostReady
              ? activation.eslintPluginEntries
              : [],
            pluginHostReady,
          };
        } catch (error) {
          await pluginPool.abort(request.transactionId).catch(() => undefined);
          transactions.delete(request.transactionId);
          configHost.deleteSession(request.transactionId);
          throw error;
        }
      }
      case 'commitConfigs': {
        const request = assertControlRequest(message.data);
        if (!(await pluginPool.commit(request.transactionId))) {
          throw new Error(
            `editor runtime: unknown staged generation ${JSON.stringify(request.transactionId)}`,
          );
        }
        transactions.delete(request.transactionId);
        configHost.deleteSession(request.transactionId);
        return { transactionId: request.transactionId, committed: true };
      }
      case 'abortConfigs': {
        const request = assertControlRequest(message.data);
        await pluginPool.abort(request.transactionId);
        transactions.delete(request.transactionId);
        configHost.deleteSession(request.transactionId);
        return { transactionId: request.transactionId, aborted: true };
      }
      case 'pluginLint':
        return pluginPool.lint(asRecord(message.data), signal);
      default:
        throw new Error(
          `editor runtime: unexpected private request ${JSON.stringify(message.kind)}`,
        );
    }
  });
  ipc.start();
  // Authentication pauses the socket before handing it to IpcClient so no
  // framed bytes can race ahead of handler installation. Explicitly resume
  // only after every private request handler is installed.
  runtimeSocket.resume();

  // Project config executes in this Node process. Reserve process.stdout for
  // the LSP wire: ordinary config logging (including process.stdout.write)
  // goes to stderr, while Go's bytes use the writer captured before the
  // redirect. A small proxy preserves stream backpressure without letting
  // child.stdout end the editor host's real stdout.
  const protocolWrite = stdout.write;
  const originalProcessStdoutWrite = process.stdout.write;
  const redirectedProcessStdoutWrite = ((...args: unknown[]) =>
    Reflect.apply(stderr.write, stderr, args)) as typeof process.stdout.write;
  process.stdout.write = redirectedProcessStdoutWrite;
  const protocolSink = new Writable({
    write(chunk: Buffer, encoding, callback) {
      try {
        Reflect.apply(protocolWrite, stdout, [chunk, encoding, callback]);
      } catch (error) {
        callback(error instanceof Error ? error : new Error(String(error)));
      }
    },
  });
  protocolSink.on('error', (error) => {
    stderr.write(`rslint: failed to forward LSP output: ${error.message}\n`);
    terminateChild('SIGTERM');
  });

  stdin.pipe(child.stdin);
  child.stdout.pipe(protocolSink);

  try {
    const outcome = await childExit;
    if (outcome.error) {
      stderr.write(
        'rslint: Go language server failed: ' + outcome.error.message + '\n',
      );
    }
    const exitCode =
      signalExitCode ??
      outcome.code ??
      (outcome.signal === 'SIGINT'
        ? 130
        : outcome.signal === 'SIGTERM'
          ? 143
          : 1);

    removeSignalHandlers();
    ipc.close();
    await listener.close();
    for (const transactionId of transactions)
      configHost.deleteSession(transactionId);
    await pluginPool.dispose();
    return exitCode;
  } finally {
    protocolSink.destroy();
    if (process.stdout.write === redirectedProcessStdoutWrite) {
      process.stdout.write = originalProcessStdoutWrite;
    }
  }
}

if (
  process.argv[1] !== undefined &&
  path.resolve(process.argv[1]) === fileURLToPath(import.meta.url)
) {
  runEditorRuntime().then(
    (code) => {
      scheduleEditorRuntimeProcessExit(code);
    },
    (error: unknown) => {
      process.stderr.write(`rslint: editor runtime failed: ${String(error)}\n`);
      scheduleEditorRuntimeProcessExit(1);
    },
  );
}

export type { IpcMessage };
