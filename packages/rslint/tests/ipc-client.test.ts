import { describe, test, expect } from '@rstest/core';
import { once } from 'node:events';
import { PassThrough } from 'node:stream';
import { IpcClient, encodeFrame, decodeFrame } from '../src/ipc/client.js';
import type { IpcMessage, MessageKind } from '../src/ipc/protocol.js';

/**
 * pairClients wires two IpcClient instances together via two PassThrough
 * streams so they can exchange frames in-process. Returns both clients
 * and a `cleanup` that closes them in the right order.
 *
 * Naming: A is the "Go-equivalent side"; B is the "peer". Tests pick
 * which side acts as the inbound handler.
 */
function pairClients(): {
  a: IpcClient;
  b: IpcClient;
  streams: {
    aToB: PassThrough;
    bToA: PassThrough;
  };
  cleanup: () => void;
} {
  // A.write → A→B → B.read
  // B.write → B→A → A.read
  const aToB = new PassThrough();
  const bToA = new PassThrough();
  // IpcClient(input, output) — input is what we read from, output is
  // where we write. So A's input is bToA (what B writes), A's output
  // is aToB (what A writes).
  const a = new IpcClient(bToA, aToB);
  const b = new IpcClient(aToB, bToA);
  return {
    a,
    b,
    streams: { aToB, bToA },
    cleanup: () => {
      a.close();
      b.close();
      // PassThrough streams don't need explicit close for tests, but
      // ending them helps ensure GC promptness in larger suites.
      aToB.end();
      bToA.end();
    },
  };
}

describe('encode/decode round-trip', () => {
  test('encodes a basic message', () => {
    const msg: IpcMessage = { kind: 'init', id: 1, data: { hello: 'world' } };
    const frame = encodeFrame(msg);
    // Header (4B) + body
    expect(frame.length).toBeGreaterThan(4);
    const header = frame.readUInt32LE(0);
    const json = JSON.stringify(msg);
    // The header is the UTF-8 BYTE length of the JSON, matching Go's
    // u32 LE length prefix. Assert against `Buffer.byteLength(...,utf8)`
    // — NOT `json.length` (UTF-16 code units). They happen to be equal
    // for this ASCII payload; the multibyte test below pins the
    // difference so an accidental `json.length`-based framing regression
    // is caught.
    expect(header).toBe(Buffer.byteLength(json, 'utf8'));
    expect(frame.length).toBe(4 + header);
    const decoded = JSON.parse(
      frame.subarray(4).toString('utf8'),
    ) as IpcMessage;
    expect(decoded.kind).toBe('init');
    expect(decoded.id).toBe(1);
    expect((decoded.data as { hello: string }).hello).toBe('world');
  });

  test('frames a multibyte payload by UTF-8 byte length (not UTF-16 .length)', () => {
    // CJK (3 bytes/char) + emoji (4 bytes, surrogate pair = 2 UTF-16
    // units) + accented Latin: every char makes the UTF-8 byte count
    // exceed the UTF-16 `.length`. If the framing used `json.length`
    // the header would under-count and the receiver would slice the
    // body short → stream desync. Pin that the header is the byte
    // length and that the frame round-trips intact.
    const msg: IpcMessage = {
      kind: 'log',
      id: 7,
      data: { text: '日本語 😀 résumé' },
    };
    const json = JSON.stringify(msg);
    const byteLen = Buffer.byteLength(json, 'utf8');

    // Precondition: this payload MUST be multibyte, otherwise the test
    // would silently degrade to the ASCII case and prove nothing.
    expect(byteLen).toBeGreaterThan(json.length);

    const frame = encodeFrame(msg);
    const header = frame.readUInt32LE(0);
    expect(header).toBe(byteLen);
    expect(header).not.toBe(json.length); // the fragile assertion would fail
    expect(frame.length).toBe(4 + byteLen);

    // Round-trips through the real decoder, body intact.
    const result = decodeFrame(frame);
    expect(result).not.toBeNull();
    expect(result!.consumed).toBe(frame.length);
    expect((result!.msg.data as { text: string }).text).toBe(
      '日本語 😀 résumé',
    );
  });

  test('decodes a single complete frame', () => {
    const msg: IpcMessage = { kind: 'log', id: 0, data: { text: 'a' } };
    const frame = encodeFrame(msg);
    const result = decodeFrame(frame);
    expect(result).not.toBeNull();
    expect(result!.consumed).toBe(frame.length);
    expect(result!.msg.kind).toBe('log');
    expect(result!.msg.id).toBe(0);
  });

  test('decodeFrame returns null when buffer is incomplete', () => {
    // Header alone, no body
    const buf = Buffer.alloc(4);
    buf.writeUInt32LE(100, 0);
    expect(decodeFrame(buf)).toBeNull();
  });

  test('decodeFrame returns null when buffer is shorter than header', () => {
    expect(decodeFrame(Buffer.alloc(0))).toBeNull();
    expect(decodeFrame(Buffer.alloc(3))).toBeNull();
  });
});

// The streaming decoder in `IpcClient.onChunk` must reassemble frames
// across arbitrary chunk boundaries: it accumulates into `this.buf` via
// `Buffer.concat` and drains COMPLETE frames in a `while` loop. The
// request/response tests above all deliver one whole frame per write, so
// neither the cross-boundary accumulation nor the multi-frame `while`
// loop is exercised by them. These cases feed deliberately split / fused
// chunks and assert every frame decodes correctly and IN ORDER.
describe('IpcClient streaming reassembly across chunk boundaries', () => {
  // Single reader fed via notification frames (id=0, no reply needed) so
  // we can observe decoded frames in arrival order without a peer.
  function makeReader(): {
    input: PassThrough;
    received: number[];
    waitForCount: (count: number) => Promise<void>;
    cleanup: () => void;
  } {
    const input = new PassThrough();
    const output = new PassThrough();
    const client = new IpcClient(input, output);
    const received: number[] = [];
    const waiters = new Set<{ count: number; resolve: () => void }>();
    client.registerNotification('log', (msg) => {
      received.push((msg.data as { n: number }).n);
      for (const waiter of waiters) {
        if (received.length < waiter.count) continue;
        waiters.delete(waiter);
        waiter.resolve();
      }
    });
    client.start();
    return {
      input,
      received,
      waitForCount(count) {
        if (received.length >= count) return Promise.resolve();
        return new Promise<void>((resolve) => {
          waiters.add({ count, resolve });
        });
      },
      cleanup: () => {
        client.close();
        input.end();
        output.end();
      },
    };
  }

  function logFrame(n: number): Buffer {
    return encodeFrame({ kind: 'log', id: 0, data: { n } });
  }

  function writeChunk(stream: PassThrough, chunk: Buffer): Promise<void> {
    return new Promise((resolve, reject) => {
      stream.write(chunk, (error) => {
        if (error) reject(error);
        else resolve();
      });
    });
  }

  test('(ii) two complete frames in ONE chunk both decode, in order (while loop)', async () => {
    const { input, received, waitForCount, cleanup } = makeReader();
    try {
      // Both frames concatenated into a single 'data' event. A
      // `while`→`if` regression would decode frame 1 and drop frame 2.
      const delivered = waitForCount(2);
      input.write(Buffer.concat([logFrame(1), logFrame(2)]));
      await delivered;
      expect(received).toEqual([1, 2]);
    } finally {
      cleanup();
    }
  });

  test('(i) a frame whose 4-byte header is split across two chunks decodes', async () => {
    const { input, received, waitForCount, cleanup } = makeReader();
    try {
      const frame = logFrame(42);
      // First 2 bytes of the length header only — buf.length (2) <
      // HEADER_BYTES (4), so the while loop must NOT consume anything.
      await writeChunk(input, frame.subarray(0, 2));
      expect(received).toEqual([]);
      // Remainder (rest of header + full body) completes the frame.
      const delivered = waitForCount(1);
      input.write(frame.subarray(2));
      await delivered;
      expect(received).toEqual([42]);
    } finally {
      cleanup();
    }
  });

  test('(iii) a partial frame (header + part of body) then its remainder decodes', async () => {
    const { input, received, waitForCount, cleanup } = makeReader();
    try {
      const frame = logFrame(7);
      // Header complete but body truncated: buf.length <
      // HEADER_BYTES + len, so the `break` arm holds the frame.
      await writeChunk(input, frame.subarray(0, 6));
      expect(received).toEqual([]);
      const delivered = waitForCount(1);
      input.write(frame.subarray(6));
      await delivered;
      expect(received).toEqual([7]);
    } finally {
      cleanup();
    }
  });

  test('mixed: a fused pair followed by a byte-by-byte dribbled frame, all in order', async () => {
    const { input, received, waitForCount, cleanup } = makeReader();
    try {
      // Two frames fused, then a third delivered one byte at a time —
      // stresses both the while loop and repeated partial accumulation.
      const firstPair = waitForCount(2);
      input.write(Buffer.concat([logFrame(10), logFrame(20)]));
      await firstPair;
      expect(received).toEqual([10, 20]);

      const f3 = logFrame(30);
      const third = waitForCount(3);
      for (let i = 0; i < f3.length; i++) {
        input.write(f3.subarray(i, i + 1));
      }
      await third;
      expect(received).toEqual([10, 20, 30]);
    } finally {
      cleanup();
    }
  });

  // ── Linear chunk-queue reassembly (perf fix: no per-chunk concat) ──
  // The decoder queues chunks and coalesces a frame's bytes exactly once
  // when complete, instead of `Buffer.concat`-ing the whole accumulator
  // on every 'data' event. These cases pin that the queue path stays
  // correct: the 4-byte LENGTH HEADER itself split across several
  // single-byte chunks must be reassembled (cross-chunk header peek), and
  // a body spanning many chunks must coalesce in order.

  test('many frames each fully dribbled byte-by-byte decode in order', async () => {
    const { input, received, waitForCount, cleanup } = makeReader();
    try {
      const ns = [1, 2, 3, 4, 5, 6, 7, 8];
      const delivered = waitForCount(ns.length);
      // Every byte of every frame — INCLUDING each frame's 4-byte length
      // header — arrives as its own 'data' event. A regression that read
      // the header via `chunks[0].readUInt32LE(0)` (instead of the
      // cross-chunk peek) would throw RangeError on a 1-byte first chunk.
      for (const n of ns) {
        const f = logFrame(n);
        for (let i = 0; i < f.length; i++) {
          input.write(f.subarray(i, i + 1));
        }
      }
      await delivered;
      expect(received).toEqual(ns);
    } finally {
      cleanup();
    }
  });

  test('frame split into many small chunks with mid-chunk frame boundaries decodes in order', async () => {
    const { input, received, waitForCount, cleanup } = makeReader();
    try {
      // Fuse three frames into one buffer, then re-slice that buffer into
      // fixed 3-byte chunks. The frame boundaries fall in the MIDDLE of
      // chunks, so the decoder must split a chunk at a frame boundary
      // (consumeFront's overshoot branch) and carry the remainder forward.
      const fused = Buffer.concat([
        logFrame(100),
        logFrame(200),
        logFrame(300),
      ]);
      const STEP = 3;
      const delivered = waitForCount(3);
      for (let i = 0; i < fused.length; i += STEP) {
        input.write(fused.subarray(i, Math.min(i + STEP, fused.length)));
      }
      await delivered;
      expect(received).toEqual([100, 200, 300]);
    } finally {
      cleanup();
    }
  });

  test('cross-chunk split header (1+1+2 bytes) before body decodes', async () => {
    const { input, received, waitForCount, cleanup } = makeReader();
    try {
      const frame = logFrame(77);
      // Header delivered as 1 byte, then 1 byte, then 2 bytes — none of
      // these prefixes alone satisfies readUInt32LE(0). Then the body.
      input.write(frame.subarray(0, 1));
      input.write(frame.subarray(1, 2));
      await writeChunk(input, frame.subarray(2, 4));
      expect(received).toEqual([]);
      const delivered = waitForCount(1);
      input.write(frame.subarray(4));
      await delivered;
      expect(received).toEqual([77]);
    } finally {
      cleanup();
    }
  });
});

describe('IpcClient request/response', () => {
  test('basic outbound request → handler reply', async () => {
    const { a, b, cleanup } = pairClients();
    try {
      b.setInboundHandler((msg) => {
        expect(msg.kind).toBe('lint');
        return { ok: true, echo: msg.data };
      });
      a.start();
      b.start();
      const resp = await a.sendRequest('lint', { x: 1 });
      expect(resp.kind).toBe('response');
      expect((resp.data as { ok: boolean }).ok).toBe(true);
    } finally {
      cleanup();
    }
  });

  test('inbound handler can issue reverse sendRequest without deadlock', async () => {
    const { a, b, cleanup } = pairClients();
    try {
      a.setInboundHandler(async (msg) => {
        expect(msg.kind).toBe('cancel');
        return { ack: 'a' };
      });
      b.setInboundHandler(async (msg) => {
        // While handling our own inbound, send a reverse RPC to A.
        const reverseResp = await b.sendRequest('cancel', {
          reverseFrom: msg.id,
        });
        return { reverseGot: (reverseResp.data as { ack: string }).ack };
      });
      a.start();
      b.start();

      const top = await a.sendRequest('init', {});
      expect((top.data as { reverseGot: string }).reverseGot).toBe('a');
    } finally {
      cleanup();
    }
  });

  test('reqID multiplexing: many concurrent requests all resolve correctly', async () => {
    const { a, b, cleanup } = pairClients();
    try {
      b.setInboundHandler(async (msg) => msg.data);
      a.start();
      b.start();

      const N = 50;
      const promises = Array.from({ length: N }, (_, i) =>
        a.sendRequest('lint', { i }).then((r) => (r.data as { i: number }).i),
      );
      const results = await Promise.all(promises);
      expect(results).toEqual(Array.from({ length: N }, (_, i) => i));
    } finally {
      cleanup();
    }
  });

  test('large frame (≥ 64 KiB) round-trips intact', async () => {
    const { a, b, cleanup } = pairClients();
    try {
      b.setInboundHandler(async (msg) => msg.data);
      a.start();
      b.start();

      const big = 'a'.repeat(200 * 1024); // 200 KiB
      const resp = await a.sendRequest('lint', { blob: big });
      expect((resp.data as { blob: string }).blob.length).toBe(big.length);
    } finally {
      cleanup();
    }
  });

  test('notification: no reply expected', async () => {
    const { a, b, cleanup } = pairClients();
    try {
      let received: string | null = null;
      let markReceived!: () => void;
      const notificationReceived = new Promise<void>((resolve) => {
        markReceived = resolve;
      });
      b.registerNotification('log', (msg) => {
        received = (msg.data as { text: string }).text;
        markReceived();
      });
      a.start();
      b.start();
      a.sendNotification('log', { text: 'hello-log' });

      await notificationReceived;
      expect(received).toBe('hello-log');
    } finally {
      cleanup();
    }
  });

  test('inbound request without handler → peer gets error reply', async () => {
    const { a, b, cleanup } = pairClients();
    try {
      // B has no inbound handler set
      a.start();
      b.start();
      await expect(a.sendRequest('lint', {})).rejects.toThrow(
        /no inbound handler registered/,
      );
    } finally {
      cleanup();
    }
  });

  test('handler thrown error → peer receives error reply', async () => {
    const { a, b, cleanup } = pairClients();
    try {
      b.setInboundHandler(() => {
        throw new Error('boom');
      });
      a.start();
      b.start();
      await expect(a.sendRequest('lint', {})).rejects.toThrow(/boom/);
    } finally {
      cleanup();
    }
  });

  test('close() rejects pending requests', async () => {
    const { a, b, cleanup } = pairClients();
    try {
      // B handler hangs forever
      b.setInboundHandler(
        () =>
          new Promise(() => {
            // never resolves: pins close() rejecting the pending request
          }),
      );
      a.start();
      b.start();
      const pending = a.sendRequest('lint', {});
      // sendRequest registers the pending resolver before writing the frame.
      // Closing immediately therefore exercises the real pending-request path
      // without relying on a scheduler turn.
      a.close();
      await expect(pending).rejects.toThrow();
    } finally {
      cleanup();
    }
  });

  test('start() is idempotent — no double listener install', async () => {
    const { a, b, streams, cleanup } = pairClients();
    try {
      a.start();
      a.start(); // must be a no-op, NOT install a second 'data' listener
      expect(streams.bToA.listenerCount('data')).toBe(1);
      b.setInboundHandler(async () => ({ ok: true }));
      b.start();

      // Keep a real round-trip control so the listener-count assertion cannot
      // pass on a client that installed no functional reader at all.
      const got = await a.sendRequest('lint', {});
      expect((got.data as { ok?: boolean }).ok).toBe(true);
    } finally {
      cleanup();
    }
  });

  test('close() is idempotent — second call neither throws nor un-closes', async () => {
    const { a, cleanup } = pairClients();
    try {
      a.start();
      a.close();
      a.close(); // must be a no-op
      // Post-close contract: subsequent sendRequest still rejects.
      // A regression where the second close() reset internal state
      // would let sendRequest hang or succeed instead of rejecting.
      await expect(a.sendRequest('lint', {})).rejects.toThrow(/closed/);
    } finally {
      cleanup();
    }
  });

  test('sendRequest after close throws', async () => {
    const { a, cleanup } = pairClients();
    try {
      a.start();
      a.close();
      await expect(a.sendRequest('lint', {})).rejects.toThrow(/closed/);
    } finally {
      cleanup();
    }
  });
});

describe('Schema parity with Go (smoke)', () => {
  test('all known message kinds are valid strings', () => {
    const kinds: MessageKind[] = [
      'lint',
      'getAstInfo',
      'response',
      'error',
      'handshake',
      'exit',
      'init',
      'cancel',
      'output',
      'log',
      'shutdown',
    ];
    for (const k of kinds) {
      const frame = encodeFrame({ kind: k, id: 0, data: null });
      const decoded = decodeFrame(frame);
      expect(decoded?.msg.kind).toBe(k);
    }
  });

  // Regression: a write failure on the output stream must NOT leave
  // pending sendRequest promises parked forever. Mirrors the Go-side
  // writerLoop EPIPE-cascade fix. Without the JS-side fix:
  //   - output.write throws (or emits error async) when stream is
  //     destroyed.
  //   - Pending sendRequest sits on its respCh promise indefinitely.
  //   - Node sometimes surfaces the unhandled 'error' event as an
  //     uncaught exception.
  // A3 regression — peer-written frame whose declared length exceeds
  // the 256 MiB cap must trigger the OOM-cap guard SYNCHRONOUSLY on the
  // 'data' event and tear the connection down. Without the guard the
  // client sits in the `if (this.buf.length < HEADER_BYTES + len) break;`
  // arm, waiting forever for a body that never comes while any further
  // chunks pile into `this.buf` (unbounded growth → worker OOM).
  //
  // This test pins the guard FIRING, not just "the pending eventually
  // rejected": it asserts (1) the cap-specific diagnostic ("exceeds cap"
  // + "stream desync") reaches stderr — that wording is emitted ONLY by
  // the guard, so a deleted guard leaves stderr empty — and (2) the
  // pending request rejects via the real teardown path, won by the
  // actual rejection rather than a watchdog. The rejection is observed
  // through a fixed microtask barrier after the synchronous guard transition,
  // so a missing rejection fails without parking under the stderr patch.
  test('cap guard fires synchronously on an oversized frame-length header and seals the client', async () => {
    const aToB = new PassThrough();
    const bToA = new PassThrough();
    const a = new IpcClient(bToA, aToB);
    a.start();

    // Capture the cap-specific diagnostic the guard writes to stderr.
    const originalStderrWrite = process.stderr.write;
    let stderr = '';
    (process.stderr as { write: unknown }).write = (
      chunk: string | Uint8Array,
    ): boolean => {
      stderr += typeof chunk === 'string' ? chunk : chunk.toString();
      return true;
    };

    const pending = a.sendRequest('lint', { test: 1 });
    try {
      // Header declaring a body 1 MiB above the 256 MiB cap, then no body.
      const header = Buffer.alloc(4);
      header.writeUInt32LE(257 * 1024 * 1024, 0);
      bToA.write(header);

      // The guard runs in the input data handler. Assert its synchronous state
      // transition before awaiting the pending rejection so a deleted guard
      // fails immediately instead of parking until the outer test watchdog.
      expect(
        (a as unknown as { closed: boolean }).closed,
        'oversized frame must seal the client',
      ).toBe(true);
    } finally {
      (process.stderr as { write: unknown }).write = originalStderrWrite;
    }

    // Await the actual teardown outcome. The suite's finite outer bound is the
    // deadlock sentinel if a mutation seals the client but forgets to reject
    // pending requests; no promise-layer count is part of this contract.
    await expect(pending).rejects.toThrow(/peer closed input stream/);

    // The cap-specific diagnostic — emitted ONLY by the guard — must be
    // present. A deleted/disabled guard leaves stderr empty here.
    expect(stderr).toMatch(/exceeds cap/);
    expect(stderr).toMatch(/stream desync/);

    // The client is sealed: a subsequent sendRequest fails fast.
    await expect(a.sendRequest('lint', {})).rejects.toThrow(
      /cannot sendRequest on closed client/,
    );

    void aToB;
  });

  test('output stream error rejects pending requests and seals client', async () => {
    const aToB = new PassThrough();
    const bToA = new PassThrough();
    const a = new IpcClient(bToA, aToB);
    a.start();

    // Start a request — this enqueues to `pending` and writes one frame
    // before parking on the response promise.
    const pending = a.sendRequest('lint', { test: 1 });

    // Destroy the output stream WITH an error. This drives the
    // 'error' event on the Writable side that `a` writes to.
    const err = new Error('simulated EPIPE');
    const outputErrored = once(aToB, 'error');
    aToB.destroy(err);
    await outputErrored;

    expect(
      (a as unknown as { closed: boolean }).closed,
      'output error must seal the client',
    ).toBe(true);
    let rejected = false;
    let rejectedMessage = '';
    try {
      await pending;
    } catch (e) {
      rejected = true;
      rejectedMessage = (e as Error).message;
    }

    expect(rejected).toBe(true);
    // The rejection should mention the underlying write failure so the
    // caller can distinguish "transport died" from "peer error" / "ctx
    // cancelled".
    expect(rejectedMessage.toLowerCase()).toMatch(/write|epipe|closed/);

    // Cleanup: nothing else to test here, the client is sealed.
    void bToA;
  });

  // After peer closes its write side (we see EOF on input), our
  // IpcClient must seal — future sendRequest calls fail fast.
  // Previously onEnd only rejected pending and left .closed=false,
  // so a sendRequest after onEnd would silently enqueue and wait
  // forever for a response that can never come.
  test('peer-closes-input seals the client (no hang on next sendRequest)', async () => {
    const aToB = new PassThrough();
    const bToA = new PassThrough();
    const a = new IpcClient(bToA, aToB);
    a.start();

    // Peer closes its write side — we see EOF on input (bToA).
    const inputEnded = once(bToA, 'end');
    bToA.end();
    await inputEnded;
    expect(
      (a as unknown as { closed: boolean }).closed,
      'peer input end must seal the client',
    ).toBe(true);

    // sendRequest must throw / reject immediately, not park.
    let rejected = false;
    let msg = '';
    try {
      await a.sendRequest('lint', {});
    } catch (e) {
      rejected = true;
      msg = (e as Error).message;
    }
    expect(rejected).toBe(true);
    expect(msg.toLowerCase()).toMatch(/closed|peer|input/);
  });

  // A future sendRequest after output error must fail fast, not park.
  test('sendRequest after output error rejects immediately', async () => {
    const aToB = new PassThrough();
    const bToA = new PassThrough();
    const a = new IpcClient(bToA, aToB);
    a.start();

    const outputErrored = once(aToB, 'error');
    aToB.destroy(new Error('simulated EPIPE'));
    await outputErrored;

    // sendRequest after the client has been sealed by onOutputError
    // must throw rather than enqueueing into pending. The throw goes
    // out of the sync body before the Promise is even constructed.
    let threw = false;
    let msg = '';
    try {
      // Avoid awaiting — the throw is synchronous, but in case rstest's
      // proxy turns it into a rejection we still want to capture it.
      const p = a.sendRequest('lint', {});
      // If we got here, sendRequest didn't throw — try awaiting in case
      // it returned a rejected Promise instead.
      await p;
    } catch (e) {
      threw = true;
      msg = (e as Error).message;
    }
    expect(threw).toBe(true);
    expect(msg).toMatch(/closed/);
  });
});

describe('IpcClient rejects pending on clean output close (no hang)', () => {
  test('in-flight sendRequest rejects when output is cleanly destroyed', async () => {
    const input = new PassThrough();
    const output = new PassThrough();
    const client = new IpcClient(input, output);
    client.start();
    // No peer responds. A CLEAN close (destroy() with no error) fires
    // no 'error' and write() doesn't throw, so without the 'close'/
    // 'finish' listeners this request would hang forever.
    const pending = client.sendRequest('init', {});
    output.destroy();
    await expect(pending).rejects.toThrow(/output stream closed|closed/);
    client.close();
  });

  test('sendRequest after a clean output close throws (not hang)', async () => {
    const input = new PassThrough();
    const output = new PassThrough();
    const client = new IpcClient(input, output);
    client.start();
    const outputClosed = once(output, 'close');
    output.destroy();
    await outputClosed;
    // `sendRequest` is async, so its top-level closed-guard throw
    // surfaces as a rejected promise, not a synchronous throw.
    await expect(client.sendRequest('init', {})).rejects.toThrow(/closed/);
    client.close();
  });
});
