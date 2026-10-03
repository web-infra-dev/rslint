/* rslint-disable @typescript-eslint/no-unsafe-type-assertion */
/**
 * Bidirectional IPC client over a Node Duplex pair (typically the CLI host
 * process's stdin/stdout, with a Go peer on the other side).
 *
 * Wire format and message shape mirror Go's `internal/ipc.Channel`:
 *
 *   `[4 bytes u32 LE length][JSON payload]`
 *   payload = WireMessage, with optional text or byte attachments
 *
 * Go supplies the shared storage layout through an internal runtime request.
 * Cross-language tests exercise the same contract through the real CLI.
 *
 * Concurrency model:
 *
 *   Node is single-threaded for JS, so no goroutines / mutexes needed —
 *   but the same hazard exists: if an inbound handler calls `sendRequest`
 *   and awaits a reply, the reply must be readable concurrently. Reads
 *   come from the stdin stream's 'data' event, which fires asynchronously
 *   via libuv — so a handler can `await` while data continues to arrive.
 *   That's the asynchrony budget we rely on.
 *
 *   Outbound requests register a Promise resolver in `pending`; the data
 *   handler routes incoming `response`/`error` frames to the matching id.
 *   Inbound requests / notifications are dispatched to user-registered
 *   handlers; for requests, the handler's resolved value (or thrown
 *   error) is wrapped into a `response`/`error` frame.
 *
 * Lifecycle:
 *
 *   - new IpcClient(stdin, stdout)
 *   - .setInboundHandler(...) and/or .registerNotification(...)
 *   - .start()                  — attaches readers / starts pumping
 *   - .sendRequest(...) / .sendNotification(...) anywhere afterward
 *   - .close()                  — stops listening, rejects pending requests
 */

import type { Readable, Writable } from 'node:stream';
import {
  receiveAttachments,
  type ReceivedAttachments,
  type MemoryTransport,
} from './memory-transport.js';
import type {
  IpcMessage,
  MessageKind,
  InboundRequestHandler,
  NotificationHandler,
  ErrorResponseData,
  MemoryBatch,
  MemoryMapping,
  WireMessage,
} from './protocol.js';

// Framing is the fixed wire format, independent of the runtime storage layout.
const HEADER_BYTES = 4;
const MAX_FRAME_BYTES = 256 * 1024 * 1024;
const RESPONSE_KIND = 'response';
const ERROR_KIND = 'error';
const TRANSPORT_CONFIG_KIND = 'transportConfig';
const TRANSPORT_RELEASE_KIND = 'transportRelease';
const RELEASE_POLL_MS = 25;

/** Internal record for a request awaiting its response. */
interface PendingRequest {
  resolve: (msg: IpcMessage) => void;
  reject: (err: Error) => void;
}

type WireFrame<T = unknown> = Omit<WireMessage, 'data'> & { data?: T };

export interface IpcClientOptions {
  /** Native reader storage, owned and closed by this IPC session. */
  readonly memoryTransport?: MemoryTransport;
  /** Child stdio index chosen by spawnIpcProcess for the prepared Unix fd. */
  readonly inheritedMemoryFd?: number;
  /**
   * Initial buffer size for the read accumulator. Frames larger than this
   * will simply grow the buffer; this is just a starting hint for typical
   * traffic. Default 64 KiB.
   */
  readonly initialReadBufferBytes?: number;
}

/**
 * Bidirectional IPC client over a Node Duplex pair.
 */
export class IpcClient {
  // ── streams ──
  private readonly input: Readable;
  private readonly output: Writable;

  // ── routing tables ──
  private readonly pending = new Map<number, PendingRequest>();
  // Only lease release functions are retained, never handler payloads/results.
  // Each owns at least one slot, bounding this queue by the negotiated arena.
  private readonly pendingReleases = new Map<
    number,
    ReceivedAttachments['release']
  >();
  private releaseTimer: NodeJS.Timeout | undefined;
  private readonly notificationHandlers = new Map<
    MessageKind,
    NotificationHandler
  >();
  private inboundHandler: InboundRequestHandler | null = null;

  // ── reader buffer ──
  // Incoming chunks accumulate in a queue and are coalesced into one
  // contiguous Buffer only when a full frame is parseable — see onChunk
  // for why a per-chunk `Buffer.concat` (the previous design) was O(N²).
  private readonly chunks: Buffer[] = [];
  private bufferedBytes = 0;

  // ── state ──
  private nextId = 1;
  private closed = false;
  private started = false;
  private memory: MemoryTransport | undefined;
  private readonly inheritedMemoryFd: number | undefined;
  private bootstrapPromise: Promise<void> | undefined;
  private mappingSent = false;
  private preparingFirstRequest: Promise<void> | undefined;
  private finishClose!: (error: Error) => void;

  /** Resolves after teardown with the first, stable termination reason. */
  readonly done = new Promise<Error>((resolve) => {
    this.finishClose = resolve;
  });

  constructor(input: Readable, output: Writable, opts: IpcClientOptions = {}) {
    this.input = input;
    this.output = output;
    this.memory = opts.memoryTransport;
    this.inheritedMemoryFd = opts.inheritedMemoryFd;
  }

  get isClosed(): boolean {
    return this.closed;
  }

  /**
   * Install the request handler for inbound non-notification frames. Set
   * before `start()`. Calling twice replaces the previous handler.
   */
  setInboundHandler(handler: InboundRequestHandler | null): void {
    this.inboundHandler = handler;
  }

  /**
   * Register a notification handler for a specific kind. Notifications are
   * id=0 frames with no reply. The same kind registered twice overwrites
   * the prior registration.
   */
  registerNotification<TIn = unknown>(
    kind: MessageKind,
    handler: NotificationHandler<TIn>,
  ): void {
    this.notificationHandlers.set(kind, handler as NotificationHandler);
  }

  /**
   * Start listening on the input stream. Idempotent — subsequent calls
   * after the first are no-ops (logged in dev as a warning).
   *
   * Both directions get error listeners. Without one on `output`, a
   * peer-closes-its-stdin (our write target) failure would surface as
   * an unhandled `error` event on the Writable — which Node treats as
   * an uncaught exception or silently drops depending on version, and
   * either way leaves pending `sendRequest`s parked on promises that
   * will never resolve. Mirroring the Go-side fix (writerLoop calls
   * Close on write error), we treat output error as terminal: reject
   * all pending and flip closed.
   */
  start(): void {
    if (this.started || this.closed) return;
    this.started = true;

    this.input.on('data', this.onChunk);
    this.input.on('end', this.onEnd);
    this.input.on('close', this.onEnd);
    this.input.on('error', this.onStreamError);
    this.output.on('error', this.onOutputError);
    // A CLEAN output close (peer ended its read side / pipe EOF /
    // `destroy()` with no error) fires no `'error'` and `write()`
    // doesn't throw — only `'close'` / `'finish'` signal it. Without
    // these, pending requests would hang forever (see onOutputClose).
    this.output.on('close', this.onOutputClose);
    this.output.on('finish', this.onOutputClose);
  }

  /**
   * Stop listening and release session resources. Every terminal path uses
   * this method, so pending requests keep the first termination reason and
   * native storage is closed once even when stream events arrive afterward.
   * The stream and child-process lifetimes remain with the caller.
   */
  close(error = new Error('IpcClient: closed')): void {
    if (this.closed) return;
    this.closed = true;

    this.input.off('data', this.onChunk);
    this.input.off('end', this.onEnd);
    this.input.off('close', this.onEnd);
    this.input.off('error', this.onStreamError);
    this.output.off('error', this.onOutputError);
    this.output.off('close', this.onOutputClose);
    this.output.off('finish', this.onOutputClose);

    this.chunks.length = 0;
    this.bufferedBytes = 0;
    this.inboundHandler = null;
    this.notificationHandlers.clear();
    for (const [, p] of this.pending) p.reject(error);
    this.pending.clear();
    clearTimeout(this.releaseTimer);
    this.releaseTimer = undefined;
    this.pendingReleases.clear();
    this.closeMemory();
    this.finishClose(error);
  }

  private closeMemory(): void {
    const memory = this.memory;
    // Clear ownership before calling native code: bootstrap failure and
    // reentrant close calls must not close the same storage a second time.
    this.memory = undefined;
    try {
      memory?.close();
    } catch (error) {
      // Cleanup failures cannot replace the request's termination reason or
      // interrupt remaining teardown, including when stderr is closing too.
      try {
        process.stderr.write(
          `rslint: memory transport close error: ${safeErrorMessage(error)}\n`,
        );
      } catch {
        // Teardown must complete even when stderr is unavailable.
      }
    }
  }

  /**
   * Send an outbound request and await its response. The returned promise
   * resolves with the response Message on success, or rejects on
   * peer-side error / client close / decode failure.
   *
   * Calls are reqID-multiplexed; multiple in-flight calls are safe.
   */
  async sendRequest<TIn = unknown, TOut = unknown>(
    kind: Exclude<MessageKind, 'response' | 'error'>,
    data: TIn,
    attachments?: readonly (string | Uint8Array)[],
  ): Promise<IpcMessage<TOut>> {
    if (kind === TRANSPORT_CONFIG_KIND) {
      throw new Error('IpcClient: transportConfig is an internal request');
    }
    if (this.closed) {
      throw new Error('IpcClient: cannot sendRequest on closed client');
    }
    if (this.memory && !this.mappingSent) await this.configureMemory();
    // A payload's toJSON may synchronously call sendRequest again. Keep only
    // bootstrap serialization ordered so the mapping stays on the first frame.
    if (this.preparingFirstRequest) await this.preparingFirstRequest;
    if (this.closed) throw await this.done;
    let finishPreparing: (() => void) | undefined;
    if (this.memory && !this.mappingSent) {
      this.preparingFirstRequest = new Promise<void>((resolve) => {
        finishPreparing = resolve;
      });
    }
    let response: Promise<IpcMessage<TOut>>;
    try {
      const mapping = this.mappingSent
        ? undefined
        : this.memory?.descriptor(this.inheritedMemoryFd);
      response = this.writeRequest(kind, data, attachments, mapping);
    } finally {
      if (finishPreparing) {
        this.preparingFirstRequest = undefined;
        finishPreparing();
      }
    }
    return response;
  }

  private async configureMemory(): Promise<void> {
    // Publish the promise before writing: even an in-process peer can reenter.
    // This RPC uses pending/write directly and never waits for its own bootstrap.
    this.bootstrapPromise ??= Promise.resolve().then(async () => {
      const memory = this.memory;
      if (!memory || this.closed) return;
      try {
        const response = await this.writeRequest(
          TRANSPORT_CONFIG_KIND,
          undefined,
        );
        if (this.closed) throw await this.done;
        memory.configure(response.data);
      } catch {
        this.closeMemory();
        if (this.closed) throw await this.done;
        // Unsupported peers, invalid layouts and allocation failures all keep
        // the session usable with complete inline attachments, without retries.
      }
    });
    return this.bootstrapPromise;
  }

  private async writeRequest<TIn = unknown, TOut = unknown>(
    kind: string,
    data: TIn,
    attachments?: readonly (string | Uint8Array)[],
    mapping?: MemoryMapping,
  ): Promise<IpcMessage<TOut>> {
    const id = this.nextId++; // id > 0 always; notifications use 0
    const frame = encodeFrame({
      kind,
      id,
      data,
      attachments: attachments?.map((value) =>
        typeof value === 'string'
          ? { text: value }
          : {
              bytes: Buffer.from(
                value.buffer,
                value.byteOffset,
                value.byteLength,
              ).toString('base64'),
            },
      ),
      transport: mapping ? { mapping } : undefined,
    });
    if (this.closed) throw await this.done;
    // Only a successfully serialized application envelope publishes the mapping.
    if (mapping) this.mappingSent = true;
    const response = new Promise<IpcMessage<TOut>>((resolve, reject) => {
      this.pending.set(id, {
        resolve: resolve as (msg: IpcMessage) => void,
        reject,
      });
    });
    // Register pending BEFORE writing, including synchronous in-process peers.
    this.writeFrameNow(frame);
    return response;
  }

  /**
   * Fire a notification frame (id=0). Returns when the frame has been
   * handed to the underlying stream's write buffer; backpressure on the
   * pipe is handled by Node's stream layer.
   */
  sendNotification<TIn = unknown>(kind: MessageKind, data: TIn): void {
    if (this.closed) {
      throw new Error('IpcClient: cannot sendNotification on closed client');
    }
    const frame = encodeFrame({ kind, id: 0, data });
    this.writeFrameNow(frame);
  }

  /**
   * Send a `response` reply manually. Normally the framework does this
   * after an inbound request handler resolves; this method is exposed so
   * advanced users can reply asynchronously from outside the handler.
   */
  sendResponse<TOut = unknown>(reqId: number, data: TOut): void {
    this.writeResponse(RESPONSE_KIND, reqId, data);
  }

  /** Send a manual error reply; request-owned release remains internal. */
  sendErrorResponse(reqId: number, message: string): void {
    this.writeResponse(ERROR_KIND, reqId, { message });
  }

  private writeResponse(
    kind: string,
    id: number,
    data: unknown,
    released?: MemoryBatch[],
  ): void {
    if (this.closed) return;
    this.writeFrameNow(
      encodeFrame({
        kind,
        id,
        data,
        transport: released ? { released } : undefined,
      }),
    );
  }

  private scheduleReleases(): void {
    if (this.closed || this.releaseTimer || this.pendingReleases.size === 0)
      return;
    // A native reader may outlive a timed-out handler. Do not delay its result
    // or keep an otherwise idle process alive while waiting for reclamation.
    this.releaseTimer = setTimeout(this.pollReleases, RELEASE_POLL_MS).unref();
  }

  private readonly pollReleases = (): void => {
    this.releaseTimer = undefined;
    try {
      for (const [id, release] of this.pendingReleases) {
        if (this.closed) return;
        const batches = release();
        if (batches === false) continue;
        this.pendingReleases.delete(id);
        if (batches)
          this.writeResponse(TRANSPORT_RELEASE_KIND, id, undefined, batches);
      }
    } catch (error) {
      this.close(
        new Error(
          `IpcClient: memory release failed: ${safeErrorMessage(error)}`,
        ),
      );
      return;
    }
    this.scheduleReleases();
  };

  // ─────────────────────────────────────────────────────────────────
  // internals
  // ─────────────────────────────────────────────────────────────────

  private writeFrameNow(frame: Buffer): void {
    // User data can close the session from toJSON during frame serialization.
    if (this.closed) return;
    // Node's stream.write returns false under backpressure but accepts
    // more data; we don't pause here because (a) IPC frames are small,
    // and (b) callers serialize their own logical pacing. If profiling
    // shows backpressure issues, switch to `await once(this.output, 'drain')`.
    //
    // Write itself may throw synchronously if the stream has already
    // errored (e.g. EPIPE after peer's stdin closed). Treat that as a
    // terminal transport failure — same cascade as onOutputError below.
    try {
      this.output.write(frame);
    } catch (err) {
      this.onOutputError(err as Error);
    }
  }

  /**
   * Output stream `error` handler. Terminal: reject all pending
   * outbound requests with a stable error so awaiters return, then
   * trigger the regular close path. Idempotent.
   */
  private readonly onOutputError = (err: Error): void => {
    if (this.closed) return;
    const wrapped = new Error(`IpcClient: output write failed: ${err.message}`);
    this.close(wrapped);
    process.stderr.write(`rslint: output write error: ${err.message}\n`);
  };

  /**
   * Output stream `'close'` / `'finish'` handler. A CLEAN close (peer
   * ended its read side / pipe EOF / `destroy()` with no error) fires
   * no `'error'` and `write()` doesn't throw, so the framed request is
   * silently dropped and its response can never arrive — without this
   * every in-flight + future `sendRequest` would hang forever (there is
   * no per-request timeout). Reject all pending and tear down, mirroring
   * `onOutputError`. Idempotent.
   */
  private readonly onOutputClose = (): void => {
    this.close(
      new Error('IpcClient: output stream closed before response received'),
    );
  };

  /**
   * stream 'data' handler — queues the chunk, then drains every COMPLETE
   * frame currently buffered.
   *
   * Why a queue instead of `this.buf = Buffer.concat([this.buf, chunk])`
   * per chunk: that re-copied the entire accumulator on every 'data'
   * event, so a single large frame delivered in K chunks cost
   * O(frameSize × K) — quadratic when a peer dribbles bytes (the
   * byte-by-byte streaming test below is the pathological case). Here a
   * chunk is only `push`ed (O(1)); the bytes for a frame are coalesced
   * into one contiguous Buffer exactly once, when the whole frame has
   * arrived (`consumeFront`). Total copying is O(total bytes), linear.
   */
  private readonly onChunk = (chunk: Buffer): void => {
    if (this.closed) return;
    this.chunks.push(chunk);
    this.bufferedBytes += chunk.length;

    while (!this.closed && this.bufferedBytes >= HEADER_BYTES) {
      const len = this.peekHeaderLen();
      // Symmetric to Go's `maxFrameSize = 256 MiB` in
      // internal/ipc/frame.go. A frame length that exceeds the
      // cap usually means a stream desync (someone wrote unframed bytes
      // into stdout, header bytes shifted by N). Without this guard we
      // would accumulate N GiB chasing a phantom payload and OOM the
      // worker. Surface as a stream-fatal error so the caller can tear
      // the connection down rather than hang. Fires SYNCHRONOUSLY on the
      // header alone, before any body bytes arrive.
      if (len > MAX_FRAME_BYTES) {
        this.onStreamError(
          new Error(
            `ipc-client: frame length ${len} exceeds cap ${MAX_FRAME_BYTES} ` +
              `(likely stream desync). Connection will be closed.`,
          ),
        );
        return;
      }
      if (this.bufferedBytes < HEADER_BYTES + len) break;

      // Whole frame is buffered: coalesce exactly its bytes into one
      // contiguous Buffer (the single allocating copy per frame). The
      // leftover bytes stay in the queue as their own (possibly sliced)
      // chunks, so no residual slab is pinned.
      const frame = this.consumeFront(HEADER_BYTES + len);
      const body = frame.subarray(HEADER_BYTES);

      let msg: WireMessage;
      try {
        msg = JSON.parse(body.toString('utf8')) as WireMessage;
      } catch (err) {
        // Malformed frame — log and skip; framing is intact (we already
        // consumed the body), so subsequent frames decode normally.
        process.stderr.write(
          `rslint: malformed JSON in frame (len=${len}): ${(err as Error).message}\n`,
        );
        continue;
      }

      if (
        msg === null ||
        typeof msg !== 'object' ||
        Array.isArray(msg) ||
        typeof msg.kind !== 'string' ||
        !Number.isSafeInteger(msg.id) ||
        msg.id < 0
      ) {
        this.onStreamError(new Error('ipc-client: invalid message envelope'));
        return;
      }

      this.dispatch(msg);
    }
  };

  /**
   * Read the 4-byte LE frame-length header at the front of the queue
   * WITHOUT consuming it. The header may straddle chunk boundaries
   * (e.g. a peer that splits the length prefix), so read it byte-by-byte
   * across the leading chunks. Caller guarantees `bufferedBytes >= 4`.
   */
  private peekHeaderLen(): number {
    // Fast path: the whole header lives in the first chunk (the common
    // case — frames usually arrive aligned).
    const first = this.chunks[0];
    if (first.length >= HEADER_BYTES) return first.readUInt32LE(0);
    // Slow path: header split across chunks — assemble the 4 bytes.
    let len = 0;
    let seen = 0;
    for (const c of this.chunks) {
      for (let i = 0; i < c.length && seen < HEADER_BYTES; i++, seen++) {
        len |= c[i] << (8 * seen);
      }
      if (seen >= HEADER_BYTES) break;
    }
    // `>>> 0` reinterprets the (possibly sign-bit-set) result as u32 LE,
    // matching Buffer.readUInt32LE.
    return len >>> 0;
  }

  /**
   * Remove and return the first `n` bytes of the queue as a single
   * contiguous Buffer. Caller guarantees `bufferedBytes >= n`.
   *
   * Slab-pinning guard (carried over from the previous design's
   * `Buffer.from(this.buf.subarray(...))`): when a frame ends mid-chunk,
   * the surviving tail is re-`Buffer.from`'d into its own small slab
   * rather than left as a `subarray` view. Otherwise a 200 MiB chunk
   * carrying one frame + a few trailing bytes would keep its whole slab
   * alive via that tiny residual view.
   */
  private consumeFront(n: number): Buffer {
    this.bufferedBytes -= n;

    // Single-chunk fast path: the front chunk alone covers `n`.
    const first = this.chunks[0];
    if (first.length === n) {
      this.chunks.shift();
      return first;
    }
    if (first.length > n) {
      const frame = first.subarray(0, n);
      // Detach the residual tail from the (possibly huge) source slab.
      this.chunks[0] = Buffer.from(first.subarray(n));
      return frame;
    }

    // Multi-chunk path: gather whole chunks until `n` bytes are covered,
    // splitting the final chunk if it overshoots.
    const parts: Buffer[] = [];
    let need = n;
    while (need > 0) {
      const c = this.chunks[0];
      if (c.length <= need) {
        parts.push(c);
        need -= c.length;
        this.chunks.shift();
      } else {
        parts.push(c.subarray(0, need));
        // Detach the residual tail from the source slab (see above).
        this.chunks[0] = Buffer.from(c.subarray(need));
        need = 0;
      }
    }
    // `Buffer.concat` allocates a fresh contiguous buffer, so the result
    // does not pin any source chunk's slab.
    return Buffer.concat(parts, n);
  }

  private readonly onEnd = (): void => {
    // EOF and clean destruction both make every future response impossible.
    this.close(new Error('IpcClient: peer closed input stream'));
  };

  private readonly onStreamError = (err: Error): void => {
    if (this.closed) return;
    this.close(new Error(`IpcClient: input read failed: ${err.message}`));
    process.stderr.write(`rslint: stream error: ${err.message}\n`);
  };

  /** Route a fully decoded frame. */
  private dispatch(msg: WireMessage): void {
    if (msg.kind === TRANSPORT_RELEASE_KIND) {
      // This endpoint never publishes shared bytes. Control frames must not
      // reach application handlers or settle an unrelated outbound request.
      this.close(new Error('IpcClient: unexpected shared memory release'));
      return;
    }
    if (msg.kind === RESPONSE_KIND || msg.kind === ERROR_KIND) {
      this.routeResponse(msg);
      return;
    }
    if (msg.kind === TRANSPORT_CONFIG_KIND) {
      // Only Go provides the layout. Never expose this control kind to handlers.
      if (msg.id !== 0) {
        this.sendErrorResponse(
          msg.id,
          'transport configuration is provided by Go',
        );
      }
      return;
    }
    if (msg.id === 0) {
      this.dispatchNotification(msg);
      return;
    }
    this.dispatchInboundRequest(msg);
  }

  private routeResponse(msg: WireMessage): void {
    const p = this.pending.get(msg.id);
    if (!p) {
      process.stderr.write(
        `rslint: orphan response id=${msg.id} kind=${msg.kind}\n`,
      );
      return;
    }
    this.pending.delete(msg.id);
    if (msg.kind === ERROR_KIND) {
      const data = msg.data as ErrorResponseData | undefined;
      const message = data?.message;
      p.reject(
        new Error(
          typeof message === 'string' && message.length > 0
            ? message
            : 'request failed',
        ),
      );
      return;
    }
    try {
      // This endpoint only writes inline attachments. Shared response ownership
      // would need a separate caller lifetime, so reject it before admitting a lease.
      if (msg.transport?.batches !== undefined) {
        throw new Error('unexpected shared IPC response attachments');
      }
      const received = receiveAttachments(msg.attachments, undefined);
      p.resolve(this.handlerMessage(msg, received));
    } catch (error) {
      p.reject(error instanceof Error ? error : new Error(String(error)));
    }
  }

  private handlerMessage(
    msg: WireMessage,
    received: ReceivedAttachments,
  ): IpcMessage {
    const result: IpcMessage = { kind: msg.kind, id: msg.id };
    if (msg.data !== undefined) result.data = msg.data;
    if (received.values !== undefined) result.attachments = received.values;
    return result;
  }

  private dispatchNotification(msg: WireMessage): void {
    const handler = this.notificationHandlers.get(msg.kind);
    if (!handler) {
      process.stderr.write(`rslint: unhandled notification kind=${msg.kind}\n`);
      return;
    }
    void runSafely(async () => {
      // Notifications have no acknowledgement and cannot own shared storage.
      if (msg.transport?.batches !== undefined) {
        throw new Error('shared IPC attachments require a request');
      }
      const received = receiveAttachments(msg.attachments, undefined);
      await handler(this.handlerMessage(msg, received));
    }, `notification:${msg.kind}`);
  }

  private dispatchInboundRequest(msg: WireMessage): void {
    // Keep the read loop running while handlers await nested IPC or workers.
    void (async () => {
      let received: ReceivedAttachments | undefined;
      let released: MemoryBatch[] | undefined;
      let kind: string = RESPONSE_KIND;
      let result: unknown;
      try {
        if (
          msg.transport !== undefined &&
          (msg.transport === null ||
            typeof msg.transport !== 'object' ||
            Array.isArray(msg.transport) ||
            msg.transport.mapping !== undefined ||
            msg.transport.released !== undefined)
        ) {
          throw new Error('invalid inbound IPC transport metadata');
        }
        received = receiveAttachments(
          msg.attachments,
          msg.transport?.batches,
          this.memory,
        );
        const handler = this.inboundHandler;
        if (!handler) {
          throw new Error(`no inbound handler registered (kind=${msg.kind})`);
        }
        result = await handler(this.handlerMessage(msg, received));
      } catch (error) {
        kind = ERROR_KIND;
        result = {
          message: safeErrorMessage(error),
        };
      } finally {
        try {
          const batches = received?.release();
          if (batches === false && received && !this.closed) {
            if (this.pendingReleases.has(msg.id)) {
              this.close(new Error('duplicate pending shared memory request'));
            } else {
              this.pendingReleases.set(msg.id, received.release);
              this.scheduleReleases();
            }
          } else if (batches !== false) {
            released = batches;
          }
        } catch (error) {
          kind = ERROR_KIND;
          result = {
            message: safeErrorMessage(error),
          };
        }
      }
      try {
        this.writeResponse(kind, msg.id, result, released);
      } catch (error) {
        // Serialization can fail after the handler returned. The already-proven
        // release still belongs on its error reply, never on the business result.
        this.writeResponse(
          ERROR_KIND,
          msg.id,
          {
            message: safeErrorMessage(error),
          },
          released,
        );
      }
    })();
  }
}

// ────────────────────────────────────────────────────────────────────
// frame encode/decode helpers (exported for tests; not for general use)
// ────────────────────────────────────────────────────────────────────

/**
 * Encode an IPC message into the `[4B u32 LE length][JSON]` wire format.
 */
export function encodeFrame<T = unknown>(msg: WireFrame<T>): Buffer {
  const body = JSON.stringify(msg);
  const length = Buffer.byteLength(body, 'utf8');
  if (length > MAX_FRAME_BYTES) {
    throw new Error(
      `ipc-client: frame length ${length} exceeds cap ${MAX_FRAME_BYTES}`,
    );
  }
  const out = Buffer.allocUnsafe(HEADER_BYTES + length);
  out.writeUInt32LE(length, 0);
  out.write(body, HEADER_BYTES, length, 'utf8');
  return out;
}

/**
 * Decode a single complete frame from `buf` starting at offset 0. Returns
 * the decoded message and the number of bytes consumed, or null if `buf`
 * doesn't yet contain a complete frame.
 *
 * Exposed for the protocol round-trip tests; production code uses the
 * streaming decode inside {@link IpcClient}.
 */
export function decodeFrame(
  buf: Buffer,
): { msg: WireMessage; consumed: number } | null {
  if (buf.length < HEADER_BYTES) return null;
  const len = buf.readUInt32LE(0);
  // Enforce the same cap as the streaming path (ipc-client.ts:290).
  // Without this guard an attacker who can write to a buffer this
  // helper consumes (test harnesses that wire arbitrary streams, or
  // callers that misuse decodeFrame on untrusted input) could pin
  // arbitrary memory via a 4 GiB header.
  if (len > MAX_FRAME_BYTES) {
    throw new Error(
      `ipc-client: frame length ${len} exceeds cap ${MAX_FRAME_BYTES} ` +
        `(possible stream desync or malicious peer)`,
    );
  }
  if (buf.length < HEADER_BYTES + len) return null;
  const body = buf.subarray(HEADER_BYTES, HEADER_BYTES + len);
  const msg = JSON.parse(body.toString('utf8')) as WireMessage;
  return { msg, consumed: HEADER_BYTES + len };
}

/**
 * runSafely invokes `fn` and traps any thrown / rejected error to stderr,
 * tagged with `tag`. Used for notification + handler bodies whose errors
 * cannot be returned to the peer.
 */
async function runSafely(fn: () => unknown, tag: string): Promise<void> {
  try {
    const ret = fn();
    if (ret instanceof Promise) await ret;
  } catch (err) {
    const message = safeErrorMessage(err);
    process.stderr.write(`rslint: handler ${tag} threw: ${message}\n`);
  }
}

/** Error conversion itself must not prevent revocation acknowledgements. */
function safeErrorMessage(error: unknown): string {
  try {
    return error instanceof Error ? String(error.message) : String(error);
  } catch {
    return 'request failed';
  }
}
