/* eslint-disable @typescript-eslint/no-unsafe-type-assertion -- IPC wire boundary casts. Each site projects a decoded frame's `unknown` / `any` payload (JSON.parse output, Node stream error) into the typed message shape this module uses; the framing + JSON contract is validated at the read boundary, not at the cast. Bulk-disabling here instead of per-line keeps the parse sites readable. */
/**
 * Bidirectional IPC client over a Node Duplex pair (typically the CLI host
 * process's stdin/stdout, with a Go peer on the other side).
 *
 * Wire format and message shape mirror Go's `internal/ipc.Channel`:
 *
 *   `[4 bytes u32 LE length][JSON payload]`
 *   payload  = { kind: MessageKind, id: number, data?: unknown }
 *
 * This is the Node-side counterpart to Go's `internal/ipc.Channel`. The
 * two are deliberately not import-coupled — keeping the packages
 * independent lets the IPC contract evolve via tests rather than via a
 * shared type module. Cross-language wire compatibility is exercised
 * end-to-end through the real CLI path (rslint-test-tools spawns the Go
 * binary over this transport); both ends must agree on the framing above.
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
import type {
  IpcMessage,
  MessageKind,
  InboundRequestHandler,
  NotificationHandler,
  ErrorResponseData,
} from './protocol.js';

/** Internal record for a request awaiting its response. */
interface PendingRequest {
  resolve: (msg: IpcMessage) => void;
  reject: (err: Error) => void;
}

/** Symbol-like sentinel kinds we route specially. */
const RESPONSE_KIND: MessageKind = 'response';
const ERROR_KIND: MessageKind = 'error';

/** Header is 4 bytes u32 LE. */
const HEADER_BYTES = 4;
/**
 * Per-frame body cap, matched to Go's `maxFrameSize` (internal/ipc/frame.go).
 * A peer that writes a frame larger than this is treated as a stream
 * desync and the connection is torn down — without this guard a
 * malformed length header would let the read queue grow unboundedly
 * until the worker OOMs.
 */
const MAX_FRAME_BYTES = 256 * 1024 * 1024;

export interface IpcClientOptions {
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

  constructor(input: Readable, output: Writable, opts: IpcClientOptions = {}) {
    this.input = input;
    this.output = output;
    void opts; // reserved for future use; signature stable
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
    if (this.started) return;
    this.started = true;

    this.input.on('data', this.onChunk);
    this.input.on('end', this.onEnd);
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
   * Stop listening. Pending outbound requests are rejected with a
   * stable error so callers' `await sendRequest(...)` resolves rather
   * than hanging forever. Idempotent.
   */
  close(): void {
    if (this.closed) return;
    this.closed = true;

    this.input.off('data', this.onChunk);
    this.input.off('end', this.onEnd);
    this.input.off('error', this.onStreamError);
    this.output.off('error', this.onOutputError);
    this.output.off('close', this.onOutputClose);
    this.output.off('finish', this.onOutputClose);

    const err = new Error('IpcClient: closed');
    for (const [, p] of this.pending) p.reject(err);
    this.pending.clear();
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
  ): Promise<IpcMessage<TOut>> {
    if (this.closed) {
      throw new Error('IpcClient: cannot sendRequest on closed client');
    }
    const id = this.nextId++; // id > 0 always; notifications use 0
    const frame = encodeFrame({ kind, id, data });

    const promise = new Promise<IpcMessage<TOut>>((resolve, reject) => {
      this.pending.set(id, {
        resolve: resolve as (msg: IpcMessage) => void,
        reject,
      });
    });

    // Order matters: register pending BEFORE writing, otherwise a fast
    // peer could respond before the resolver is in the map.
    this.writeFrameNow(frame);
    return promise;
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
    if (this.closed) return;
    this.writeFrameNow(encodeFrame({ kind: 'response', id: reqId, data }));
  }

  /**
   * Send an `error` reply. Same caveat as {@link sendResponse}.
   */
  sendErrorResponse(reqId: number, message: string): void {
    if (this.closed) return;
    this.writeFrameNow(
      encodeFrame<ErrorResponseData>({
        kind: 'error',
        id: reqId,
        data: { message },
      }),
    );
  }

  // ─────────────────────────────────────────────────────────────────
  // internals
  // ─────────────────────────────────────────────────────────────────

  private writeFrameNow(frame: Buffer): void {
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
    process.stderr.write(`rslint: output write error: ${err.message}\n`);
    const wrapped = new Error(`IpcClient: output write failed: ${err.message}`);
    for (const [, p] of this.pending) p.reject(wrapped);
    this.pending.clear();
    this.closed = true;
    // Detach listeners to mirror close() — close() itself is a no-op
    // now (closed=true) but we should not leave the input listeners
    // dangling.
    this.input.off('data', this.onChunk);
    this.input.off('end', this.onEnd);
    this.input.off('error', this.onStreamError);
    this.output.off('error', this.onOutputError);
    this.output.off('close', this.onOutputClose);
    this.output.off('finish', this.onOutputClose);
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
    if (this.closed) return;
    const err = new Error(
      'IpcClient: output stream closed before response received',
    );
    for (const [, p] of this.pending) p.reject(err);
    this.pending.clear();
    this.closed = true;
    this.input.off('data', this.onChunk);
    this.input.off('end', this.onEnd);
    this.input.off('error', this.onStreamError);
    this.output.off('error', this.onOutputError);
    this.output.off('close', this.onOutputClose);
    this.output.off('finish', this.onOutputClose);
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
    this.chunks.push(chunk);
    this.bufferedBytes += chunk.length;

    while (this.bufferedBytes >= HEADER_BYTES) {
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

      let msg: IpcMessage;
      try {
        msg = JSON.parse(body.toString('utf8')) as IpcMessage;
      } catch (err) {
        // Malformed frame — log and skip; framing is intact (we already
        // consumed the body), so subsequent frames decode normally.
        process.stderr.write(
          `rslint: malformed JSON in frame (len=${len}): ${(err as Error).message}\n`,
        );
        continue;
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
    // Peer closed the input stream. We can't get any more responses,
    // so seal the client: reject pending requests + set closed=true
    // + detach listeners. Without setting closed=true, a subsequent
    // sendRequest would silently enqueue into pending and wait
    // forever for a response that can never arrive (the peer's write
    // end is gone). The output side is left as-is — the underlying
    // stream may still be writable from our perspective, but we
    // surface "closed" so callers get a deterministic error
    // immediately instead of an indefinite hang.
    if (this.closed) return;
    this.closed = true;
    const err = new Error('IpcClient: peer closed input stream');
    for (const [, p] of this.pending) p.reject(err);
    this.pending.clear();
    this.input.off('data', this.onChunk);
    this.input.off('end', this.onEnd);
    this.input.off('error', this.onStreamError);
    this.output.off('error', this.onOutputError);
    // Mirror close(): also detach the output 'close'/'finish' listeners
    // (added for the Windows clean-close fix). onStreamError delegates
    // here, so without this both teardown paths would leak the
    // onOutputClose listener on the output stream.
    this.output.off('close', this.onOutputClose);
    this.output.off('finish', this.onOutputClose);
  };

  private readonly onStreamError = (err: Error): void => {
    process.stderr.write(`rslint: stream error: ${err.message}\n`);
    this.onEnd();
  };

  /** Route a fully decoded frame. */
  private dispatch(msg: IpcMessage): void {
    if (msg.kind === RESPONSE_KIND || msg.kind === ERROR_KIND) {
      this.routeResponse(msg);
      return;
    }
    if (msg.id === 0) {
      this.dispatchNotification(msg);
      return;
    }
    this.dispatchInboundRequest(msg);
  }

  private routeResponse(msg: IpcMessage): void {
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
      p.reject(new Error(`peer error: ${data?.message ?? '(no message)'}`));
      return;
    }
    p.resolve(msg);
  }

  private dispatchNotification(msg: IpcMessage): void {
    const handler = this.notificationHandlers.get(msg.kind);
    if (!handler) {
      // A notification with no registered handler is unexpected — the peer
      // only emits kinds we register. Surface it to stderr as a diagnostic
      // rather than erroring; the frame body is already fully consumed.
      process.stderr.write(`rslint: unhandled notification kind=${msg.kind}\n`);
      return;
    }
    void runSafely(async () => handler(msg), `notification:${msg.kind}`);
  }

  private dispatchInboundRequest(msg: IpcMessage): void {
    const handler = this.inboundHandler;
    if (!handler) {
      this.sendErrorResponse(
        msg.id,
        `no inbound handler registered (kind=${msg.kind})`,
      );
      return;
    }
    // Spawn the handler async so the data loop continues to consume frames
    // even while a handler awaits. This is what enables in-handler
    // sendRequest to receive its reply.
    void (async () => {
      try {
        const result = await handler(msg);
        this.sendResponse(msg.id, result);
      } catch (err) {
        const message = err instanceof Error ? err.message : String(err);
        this.sendErrorResponse(msg.id, message);
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
export function encodeFrame<T = unknown>(msg: IpcMessage<T>): Buffer {
  const body = Buffer.from(JSON.stringify(msg), 'utf8');
  const out = Buffer.allocUnsafe(HEADER_BYTES + body.length);
  out.writeUInt32LE(body.length, 0);
  body.copy(out, HEADER_BYTES);
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
): { msg: IpcMessage; consumed: number } | null {
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
  const msg = JSON.parse(body.toString('utf8')) as IpcMessage;
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
    const message = err instanceof Error ? err.message : String(err);
    process.stderr.write(`rslint: handler ${tag} threw: ${message}\n`);
  }
}
