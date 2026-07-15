import { spawn, ChildProcess } from 'child_process';
import { Socket } from 'node:net';
import { RSLintService } from '../service/service.js';
import { resolveRslintBinary } from './resolve-binary.js';
import type {
  RslintServiceInterface,
  RSlintOptions,
  PendingMessage,
  IpcMessage,
  InboundRequestHandler,
  LintOptions,
  LintResponse,
} from '../types.js';

/**
 * Node.js implementation of RslintService using child processes
 */
export class NodeRslintService implements RslintServiceInterface {
  private nextMessageId: number;
  private readonly pendingMessages: Map<number, PendingMessage>;
  private readonly rslintPath: string;
  private readonly process: ChildProcess;
  private chunks: Buffer[];
  private chunkSize: number;
  private expectedSize: number | null;
  // Set once the process can no longer answer (crashed or terminated). Guards
  // sendMessage so requests made after death reject immediately instead of
  // registering a pending that nothing will ever resolve.
  private dead: boolean;
  // True once close() has sent the graceful 'exit' signal: the peer's exit is
  // then EXPECTED (not a crash), so the exit handler resolves in-flight
  // requests instead of rejecting them — otherwise close()'s pending 'exit'
  // request could reject and surface as an unhandledRejection.
  private closing: boolean;
  private inboundHandler: InboundRequestHandler | null;

  constructor(options: RSlintOptions = {}) {
    this.nextMessageId = 1;
    this.pendingMessages = new Map();
    this.rslintPath = options.rslintPath || resolveRslintBinary();
    this.dead = false;
    this.closing = false;
    this.inboundHandler = null;

    this.process = spawn(this.rslintPath, ['--api'], {
      stdio: ['pipe', 'pipe', 'inherit'],
      cwd: options.workingDirectory || process.cwd(),
      env: {
        ...process.env,
      },
    });

    // Start idle: the resident child + its stdio pipes are unref'd so a caller
    // that never calls close() (e.g. a one-off script) still lets the Node
    // process exit; setLoopActive(true) re-refs them around an in-flight request
    // (see sendMessage). The Go peer then sees stdin EOF on parent exit and
    // exits too (Service.Start returns on io.EOF). close()/[Symbol.asyncDispose]
    // remain for prompt teardown in long-lived hosts.
    this.setLoopActive(false);

    // Set up binary message reading
    this.process.stdout!.on('data', (data) => {
      this.handleChunk(data);
    });

    // If the process dies, reject every in-flight request — otherwise their
    // promises hang forever, since the Go peer can no longer answer. Mirrors the
    // wasm browser backend's worker.onerror reject-all-pending.
    this.process.on('error', (err) => {
      this.dead = true;
      this.rejectAllPending(new Error(`rslint process error: ${err.message}`));
    });
    this.process.on('exit', (code, signal) => {
      this.dead = true;
      if (this.closing) {
        // Expected shutdown: close() asked the peer to exit. Settle in-flight
        // requests (the 'exit' request itself) by resolving — the peer may exit
        // before we read its ack frame, and rejecting here would turn close()'s
        // pending into an unhandledRejection.
        this.resolveAllPending();
      } else {
        this.rejectAllPending(
          new Error(
            `rslint process exited unexpectedly (code=${code}, signal=${signal})`,
          ),
        );
      }
    });
    // Once the peer closes, further stdin writes raise EPIPE; swallow it so it
    // doesn't surface as an unhandled 'error' (pending requests are already
    // rejected by the exit/error handlers above).
    this.process.stdin!.on('error', () => {
      /* EPIPE after the peer closed — already handled by the exit/error handlers */
    });

    this.chunks = [];
    this.chunkSize = 0;
    this.expectedSize = null;
  }

  /**
   * Keep the Node event loop alive only while a request is in flight. The
   * resident child and its stdio pipes are unref'd while idle so a caller that
   * never calls close() still lets the process exit; they are ref'd around an
   * in-flight request because a pending promise alone does NOT keep the loop
   * alive — without the ref the loop could drain before the response arrives,
   * leaving the await unsettled (Node would exit with code 13). The piped stdio
   * streams are net.Socket at runtime (with ref/unref); child_process widens
   * them to Readable/Writable, so narrow via `instanceof Socket` to reach those.
   */
  private setLoopActive(active: boolean): void {
    if (active) this.process.ref();
    else this.process.unref();
    for (const stream of [this.process.stdin, this.process.stdout]) {
      if (stream instanceof Socket) {
        if (active) stream.ref();
        else stream.unref();
      }
    }
  }

  private updateLoopActivity(): void {
    // Every valid Go -> Node reverse request is nested inside an unresolved
    // Node -> Go request (currently lint). That outer pending request owns the
    // event-loop ref for the whole exchange. Once Go answers it, any reverse
    // handler still evaluating a non-cancellable JavaScript module is orphaned:
    // Go has stopped waiting for its reply, so it must not pin the caller's
    // process. Unref is not termination; a later sendMessage re-refs the same
    // reusable Go child immediately.
    this.setLoopActive(!this.dead && this.pendingMessages.size > 0);
  }

  /** Install the handler for positive-id requests sent by the Go peer. */
  setInboundHandler(handler: InboundRequestHandler | null): void {
    this.inboundHandler = handler;
  }

  /**
   * Send a message to the rslint process
   */
  async sendMessage(kind: string, data: any): Promise<any> {
    return new Promise((resolve, reject) => {
      // Process already gone — fail fast instead of registering a pending that
      // no exit/error/terminate handler will ever reject (they only sweep
      // pendings that exist at the moment they fire).
      if (this.dead) {
        reject(new Error('rslint service is no longer running'));
        return;
      }
      // 'exit' is the graceful-shutdown signal — from here the peer is expected
      // to exit, so its 'exit' event must not be treated as a crash.
      if (kind === 'exit') {
        this.closing = true;
      }
      const id = this.nextMessageId++;
      const message: IpcMessage = { id, kind, data };

      // Register promise callbacks
      this.pendingMessages.set(id, { resolve, reject });
      // Keep the child + pipes referenced until the response arrives (a
      // pending promise alone does not keep Node's event loop alive).
      this.updateLoopActivity();
      try {
        this.writeMessage(message);
      } catch (error) {
        this.pendingMessages.delete(id);
        this.updateLoopActivity();
        reject(error instanceof Error ? error : new Error(String(error)));
      }
    });
  }

  private writeMessage(message: IpcMessage): void {
    if (this.dead) return;
    const jsonBuffer = Buffer.from(JSON.stringify(message), 'utf8');
    const length = Buffer.alloc(4);
    length.writeUInt32LE(jsonBuffer.length, 0);
    this.process.stdin!.write(Buffer.concat([length, jsonBuffer]));
  }

  /**
   * Handle incoming binary data chunks
   */
  private handleChunk(chunk: Buffer): void {
    this.chunks.push(chunk);
    this.chunkSize += chunk.length;

    // Process complete messages
    while (true) {
      // Read message length if we don't have it yet
      if (this.expectedSize === null) {
        if (this.chunkSize < 4) return;

        // Combine chunks to read the message length
        const combined = Buffer.concat(this.chunks);
        this.expectedSize = combined.readUInt32LE(0);

        // Remove length bytes from buffer
        this.chunks = [combined.subarray(4)];
        this.chunkSize -= 4;
      }

      // Check if we have the full message
      if (this.chunkSize < this.expectedSize) return;

      // Read the message content
      const combined = Buffer.concat(this.chunks);
      const message = combined.subarray(0, this.expectedSize).toString('utf8');

      // Handle the message
      try {
        const parsed: IpcMessage = JSON.parse(message);
        this.handleMessage(parsed);
      } catch (err) {
        console.error('Error parsing message:', err);
      }

      // Reset for next message
      this.chunks = [combined.subarray(this.expectedSize)];
      this.chunkSize = this.chunks[0].length;
      this.expectedSize = null;
    }
  }

  /**
   * Handle a complete message from rslint
   */
  private handleMessage(message: IpcMessage): void {
    const { id, kind, data } = message;
    // IDs are allocated independently in each direction and can collide. Only
    // response/error frames settle an outbound request; every other positive-id
    // frame is a request from Go that must be answered independently.
    if (kind === 'response' || kind === 'error') {
      const pending = this.pendingMessages.get(id);
      if (!pending) return;

      this.pendingMessages.delete(id);
      if (kind === 'error') {
        pending.reject(new Error(data?.message ?? 'rslint request failed'));
      } else {
        pending.resolve(data);
      }
      this.updateLoopActivity();
      return;
    }

    if (id <= 0) return;

    void Promise.resolve()
      .then(async () => {
        if (!this.inboundHandler) {
          throw new Error(
            `no inbound handler registered (kind=${message.kind})`,
          );
        }
        return this.inboundHandler(message);
      })
      .then(
        (result) => {
          try {
            this.writeMessage({ id, kind: 'response', data: result });
          } catch (error) {
            const detail =
              error instanceof Error ? error.message : String(error);
            this.writeMessage({
              id,
              kind: 'error',
              data: { message: `failed to encode inbound response: ${detail}` },
            });
          }
        },
        (error: unknown) => {
          const detail = error instanceof Error ? error.message : String(error);
          this.writeMessage({
            id,
            kind: 'error',
            data: { message: detail },
          });
        },
      )
      .catch(() => {
        // A terminal stdin write failure is reported by the stream/process
        // handlers, which also settle the outer request. Avoid a detached
        // reverse-response chain becoming an unhandled rejection meanwhile.
      });
  }

  /**
   * Reject every in-flight request and clear the queue. Called when the process
   * dies (exit/error) or is terminated, so callers never hang on a process that
   * can no longer reply. No-op when nothing is pending (the normal close path).
   */
  private rejectAllPending(err: Error): void {
    if (this.pendingMessages.size === 0) return;
    for (const [, pending] of this.pendingMessages) {
      pending.reject(err);
    }
    this.pendingMessages.clear();
    this.updateLoopActivity();
  }

  /**
   * Resolve every in-flight request (no payload) and clear the queue. Used on
   * the expected close() shutdown path so the 'exit' request settles cleanly
   * instead of rejecting. No-op when nothing is pending.
   */
  private resolveAllPending(): void {
    if (this.pendingMessages.size === 0) return;
    for (const [, pending] of this.pendingMessages) {
      pending.resolve(null);
    }
    this.pendingMessages.clear();
    this.updateLoopActivity();
  }

  /**
   * Terminate the rslint process
   */
  terminate(): void {
    this.dead = true;
    if (this.process && !this.process.killed) {
      this.process.stdin!.end();
      this.process.kill();
    }
    this.rejectAllPending(new Error('rslint service terminated'));
  }
}

/**
 * One-shot convenience: spin up a Node-backed service, run a single lint
 * request, then tear it down. This is an internal/tooling surface (the
 * rule-tester and the ESLint-plugin conformance harnesses) reached via the
 * `@rslint/core/internal` subpath — the package root deliberately exposes only
 * the high-level `Rslint` class as its linting surface, not this low-level engine.
 */
export async function lint(options: LintOptions): Promise<LintResponse> {
  const service = new RSLintService(
    new NodeRslintService({
      workingDirectory: options.workingDirectory,
    }),
  );
  try {
    return await service.lint(options);
  } finally {
    await service.close();
  }
}

export type { LintOptions, LintResponse, Diagnostic } from '../types.js';
