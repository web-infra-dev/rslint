import { describe, test, expect, beforeAll, afterAll } from 'rstack/test';
import { once } from 'node:events';
import { PassThrough } from 'node:stream';
import { execFile } from 'node:child_process';
import { createHash } from 'node:crypto';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { promisify } from 'node:util';
import { IpcClient, encodeFrame, decodeFrame } from '../src/ipc/client.js';
import { spawnIpcProcess } from '../src/ipc/process.js';
import type { MessageKind } from '../src/ipc/protocol.js';
import type { IpcClientOptions } from '../src/ipc/client.js';
import {
  createMemoryTransport,
  readAttachmentBytes,
} from '../src/ipc/memory-transport.js';
import {
  getNativeBinding,
  type SharedBytes,
  type MemoryConfiguration,
} from '../src/native/binding.js';
import type {
  WireMessage as IpcMessage,
  MemoryBatch,
} from '../src/ipc/protocol.js';

// Deliberately differs from Go's production layout. Node consumes peer values.
const MEMORY_CONFIG: MemoryConfiguration = {
  version: 1,
  slotCount: 3,
  slotSize: 4096,
  headerSize: 512,
  publicationStride: 32,
};
const MAX_GENERATION = 0xffff_ffff;

/**
 * pairClients wires two IpcClient instances together via two PassThrough
 * streams so they can exchange frames in-process. Returns both clients
 * and a `cleanup` that closes them in the right order.
 *
 * Naming: A is the "Go-equivalent side"; B is the "peer". Tests pick
 * which side acts as the inbound handler.
 */
function pairClients(options: IpcClientOptions = {}): {
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
  const b = new IpcClient(aToB, bToA, options);
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

describe('IPC byte attachments', () => {
  function receiver(
    options: {
      memoryTransport?: ReturnType<typeof createMemoryTransport>;
    } = {},
  ) {
    const memory = options.memoryTransport;
    const pair = pairClients({
      createMemoryTransport: memory ? () => memory : undefined,
    });
    let id = 0;
    pair.b.start();
    if (memory) {
      // Storage/lease tests start after a successful private setup. Real fd
      // transfer is covered below by the Go binary peer.
      memory.transfer = () => ({ version: MEMORY_CONFIG.version });
      const discard = () => undefined;
      pair.streams.bToA.on('data', discard);
      pair.streams.aToB.write(
        encodeFrame({
          kind: 'transportPrepare',
          id: 1000,
          data: { configuration: MEMORY_CONFIG },
        }),
      );
      pair.streams.aToB.write(encodeFrame({ kind: 'transportCommit', id: 0 }));
      pair.streams.bToA.off('data', discard);
    }
    return {
      ...pair,
      async request(message: Partial<IpcMessage>) {
        const response = once(pair.streams.bToA, 'data');
        pair.streams.aToB.write(
          encodeFrame({
            kind: 'testAttachments',
            id: ++id,
            ...message,
          }),
        );
        return decodeFrame((await response)[0])!.msg;
      },
    };
  }

  function emptyBatch(): MemoryBatch {
    return { slot: 0, generation: 1, length: 0 };
  }

  test('delivers complete inline attachments through the ordinary request API', async () => {
    const pair = pairClients();
    pair.b.setInboundHandler((msg) => {
      expect(msg.data).toEqual({ arbitrary: true });
      expect(msg.attachments).toEqual(['', '\ufeff😀\u0000\r\n']);
      expect('transport' in msg).toBe(false);
      return { ok: true };
    });
    pair.a.start();
    pair.b.start();
    try {
      expect(
        (
          await pair.a.sendRequest('anything', { arbitrary: true }, [
            '',
            '\ufeff😀\u0000\r\n',
          ])
        ).data,
      ).toEqual({ ok: true });
    } finally {
      pair.cleanup();
    }
  });

  test('round trips arbitrary binary and typed-array byte ranges without UTF-8 conversion', async () => {
    const binary = Buffer.from(Array.from({ length: 256 }, (_, i) => i));
    const backing = Uint8Array.of(9, 0, 255, 128, 10);
    const pair = pairClients();
    pair.b.setInboundHandler((msg) => {
      expect(msg.data).toEqual({ nested: { format: 'consumer-defined' } });
      expect(msg.attachments).toEqual([
        binary,
        Buffer.from([0, 255, 128]),
        Buffer.alloc(0),
        '',
      ]);
      expect(readAttachmentBytes(msg.attachments![0])).toEqual(binary);
      expect(readAttachmentBytes(msg.attachments![3])).toEqual(Buffer.alloc(0));
      return { size: readAttachmentBytes(msg.attachments![0]).length };
    });
    pair.a.start();
    pair.b.start();
    try {
      const response = await pair.a.sendRequest(
        'customBinaryPayload',
        { nested: { format: 'consumer-defined' } },
        [binary, backing.subarray(1, 4), Buffer.alloc(0), ''],
      );
      expect(response.data).toEqual({ size: 256 });
      expect(backing).toEqual(Uint8Array.of(9, 0, 255, 128, 10));
    } finally {
      pair.cleanup();
    }
  });

  test.each([
    'A',
    'AA',
    'AAA',
    '====',
    'AA=A',
    'A===',
    'AAAA=',
    'AA\n=',
    '-_==',
    'AB==',
    'AAB=',
  ])('rejects malformed or non-canonical base64 bytes: %j', async (bytes) => {
    const pair = receiver();
    let dispatched = false;
    pair.b.setInboundHandler(() => {
      dispatched = true;
    });
    try {
      const result = await pair.request({ attachments: [{ bytes }] });
      expect(result.kind).toBe('error');
      expect(result.data).toEqual({ message: 'invalid IPC attachment bytes' });
      expect(dispatched).toBe(false);
    } finally {
      pair.cleanup();
    }
  });

  test('advertises once without loading native storage or delaying first requests', async () => {
    let created = 0;
    const pair = pairClients({
      createMemoryTransport: () => {
        created++;
        throw new Error('must stay lazy');
      },
    });
    const frames: IpcMessage[] = [];
    pair.streams.bToA.on('data', (chunk: Buffer) => {
      const message = decodeFrame(chunk)!.msg;
      frames.push(message);
      pair.streams.aToB.write(
        encodeFrame({ kind: 'response', id: message.id, data: { ok: true } }),
      );
    });
    pair.b.start();
    try {
      await Promise.all([
        pair.b.sendRequest('init', {}),
        pair.b.sendRequest('other', {}),
      ]);
      expect(frames.map(({ kind }) => kind)).toEqual(['init', 'other']);
      expect(frames[0].transport).toEqual({ sharedMemory: 1 });
      expect(frames[1].transport).toBeUndefined();
      expect(created).toBe(0);
    } finally {
      pair.cleanup();
    }
  });

  test.each(['commit', 'abort'] as const)(
    'prepares storage once and observes %s before admitting readers',
    async (outcome) => {
      let created = 0;
      const memory = createMemoryTransport();
      memory.transfer = () => ({ version: 1 });
      const pair = pairClients({
        createMemoryTransport: () => {
          created++;
          return memory;
        },
      });
      pair.b.setInboundHandler(() => true);
      pair.b.start();
      async function request(message: IpcMessage) {
        const response = once(pair.streams.bToA, 'data');
        pair.streams.aToB.write(encodeFrame(message));
        return decodeFrame((await response)[0])!.msg;
      }
      try {
        expect(
          (
            await request({
              kind: 'transportPrepare',
              id: 1,
              data: { configuration: MEMORY_CONFIG },
            })
          ).data,
        ).toEqual({ version: 1 });
        const attached = {
          kind: 'bytes',
          id: 2,
          attachments: [{ range: { offset: 0, length: 0 } }],
          transport: { batches: [emptyBatch()] },
        };
        expect((await request(attached)).kind).toBe('error');
        pair.streams.aToB.write(
          encodeFrame({
            kind: outcome === 'commit' ? 'transportCommit' : 'transportAbort',
            id: 0,
          }),
        );
        expect((await request({ ...attached, id: 3 })).kind).toBe(
          outcome === 'commit' ? 'response' : 'error',
        );
        expect(
          (
            await request({
              kind: 'transportPrepare',
              id: 4,
              data: { configuration: MEMORY_CONFIG },
            })
          ).kind,
        ).toBe('error');
        expect(created).toBe(1);
        if (outcome === 'abort') expect(memory.configuration()).toBeUndefined();
      } finally {
        pair.cleanup();
      }
    },
  );

  test.each(['load', 'configuration', 'transfer'] as const)(
    'retains inline attachments after native %s failure, without retry',
    async (failure) => {
      let created = 0;
      let closed = 0;
      const memory = createMemoryTransport();
      const close = memory.close;
      memory.close = () => {
        closed++;
        close();
      };
      memory.transfer = () => {
        throw new Error('transfer unavailable');
      };
      const pair = pairClients({
        createMemoryTransport: () => {
          created++;
          if (failure === 'load') {
            memory.close();
            throw new Error('addon unavailable');
          }
          return memory;
        },
      });
      const text = '\ufeffconst complete = "😀";\u0000\r\n';
      pair.b.setInboundHandler((msg) => {
        expect(msg.attachments).toEqual([text]);
        return true;
      });
      pair.b.start();
      async function request(message: IpcMessage) {
        const response = once(pair.streams.bToA, 'data');
        pair.streams.aToB.write(encodeFrame(message));
        return decodeFrame((await response)[0])!.msg;
      }
      try {
        for (const id of [1, 2]) {
          expect(
            (
              await request({
                kind: 'transportPrepare',
                id,
                data: {
                  configuration:
                    failure === 'configuration' ? {} : MEMORY_CONFIG,
                },
              })
            ).kind,
          ).toBe('error');
        }
        expect(
          (await request({ kind: 'inline', id: 3, attachments: [{ text }] }))
            .kind,
        ).toBe('response');
        expect(created).toBe(1);
        expect(closed).toBe(1);
      } finally {
        pair.cleanup();
      }
    },
  );

  test.each(['before prepare', 'create', 'configure', 'transfer'] as const)(
    'close during %s cannot resurrect storage',
    async (when) => {
      let created = 0;
      let closed = 0;
      const memory = createMemoryTransport();
      const close = memory.close;
      memory.close = () => {
        closed++;
        close();
      };
      const reason = new Error('session closed');
      const pair = pairClients({
        createMemoryTransport: () => {
          created++;
          if (when === 'create') pair.b.close(reason);
          return memory;
        },
      });
      const configure = memory.configure;
      memory.configure = (value) => {
        configure(value);
        if (when === 'configure') pair.b.close(reason);
      };
      memory.transfer = () => {
        pair.b.close(reason);
        return { version: 1 };
      };
      const frames: IpcMessage[] = [];
      pair.streams.bToA.on('data', (chunk: Buffer) => {
        frames.push(decodeFrame(chunk)!.msg);
      });
      pair.b.start();
      try {
        if (when === 'before prepare') pair.b.close(reason);
        pair.streams.aToB.write(
          encodeFrame({
            kind: 'transportPrepare',
            id: 1,
            data: { configuration: MEMORY_CONFIG },
          }),
        );
        expect(await pair.b.done).toBe(reason);
        expect(created).toBe(when === 'before prepare' ? 0 : 1);
        expect(closed).toBe(when === 'before prepare' ? 0 : 1);
        expect(frames).toEqual([]);
      } finally {
        pair.cleanup();
        if (!created) memory.close();
      }
    },
  );

  test('does not dispatch the reserved configuration kind to application handlers', async () => {
    const pair = receiver();
    let calls = 0;
    pair.b.setInboundHandler(() => {
      calls++;
    });
    pair.b.registerNotification('transportConfig', () => {
      calls++;
    });
    try {
      const response = await pair.request({ kind: 'transportConfig' });
      expect(response.kind).toBe('error');
      pair.streams.aToB.write(encodeFrame({ kind: 'transportConfig', id: 0 }));
      expect(calls).toBe(0);
    } finally {
      pair.cleanup();
    }
  });

  test.each([false, true])(
    'keeps capability on the first written frame under serialization reentry (outer fails: %s)',
    async (failOuter) => {
      const pair = pairClients({
        createMemoryTransport: () => {
          throw new Error('must stay lazy');
        },
      });
      const frames: IpcMessage[] = [];
      pair.streams.bToA.on('data', (chunk: Buffer) => {
        const message = decodeFrame(chunk)!.msg;
        frames.push(message);
        pair.streams.aToB.write(
          encodeFrame({
            kind: 'response',
            id: message.id,
            data: {},
          }),
        );
      });
      pair.b.start();
      let nested: Promise<unknown> | undefined;
      try {
        const outer = pair.b.sendRequest('outer', {
          toJSON() {
            nested = pair.b.sendRequest('nested', {});
            if (failOuter) throw new Error('cannot encode outer');
            return {};
          },
        });
        if (failOuter)
          await expect(outer).rejects.toThrow('cannot encode outer');
        else await outer;
        await nested;
        expect(frames.map((message) => message.kind)).toEqual(
          failOuter ? ['nested'] : ['outer', 'nested'],
        );
        expect(frames[0].transport).toEqual({ sharedMemory: 1 });
        if (!failOuter) expect(frames[1].transport).toBeUndefined();
      } finally {
        pair.cleanup();
      }
    },
  );

  test.each(['fd', 'configure', 'version', 'descriptor'] as const)(
    'closes native storage when initialization fails: %s',
    (failure) => {
      const binding = getNativeBinding();
      const MemoryArena = binding.MemoryArena;
      let closed = false;
      binding.MemoryArena = class {
        fd() {
          if (failure === 'fd') throw new Error('fd unavailable');
          return undefined;
        }
        configure() {
          if (failure === 'configure') throw new Error('configure unavailable');
        }
        descriptor() {
          if (failure === 'descriptor')
            throw new Error('descriptor unavailable');
          return { version: MEMORY_CONFIG.version + 1 };
        }
        sendFd() {
          throw new Error('not transferred in this test');
        }
        register(): number {
          throw new Error('must not register');
        }
        release(): boolean {
          throw new Error('must not release');
        }
        close() {
          closed = true;
        }
      };
      try {
        expect(() => createMemoryTransport().configure(MEMORY_CONFIG)).toThrow(
          failure === 'version'
            ? 'inconsistent shared memory version'
            : `${failure} unavailable`,
        );
        expect(closed).toBe(true);
      } finally {
        binding.MemoryArena = MemoryArena;
      }
    },
  );

  test.each(['windows', 'unix'] as const)(
    'normalizes nullable native handle fields before publishing a %s descriptor',
    (platform) => {
      const binding = getNativeBinding();
      const MemoryArena = binding.MemoryArena;
      let closed = false;
      binding.MemoryArena = class {
        fd() {
          return platform === 'windows' ? null : 5;
        }
        configure(config: MemoryConfiguration) {
          expect(config).toEqual(MEMORY_CONFIG);
        }
        descriptor() {
          return {
            version: MEMORY_CONFIG.version,
            fd: platform === 'windows' ? null : 5,
            handle: platform === 'windows' ? '42' : null,
            processId: platform === 'windows' ? 123 : null,
          };
        }
        sendFd() {
          throw new Error('not transferred in this test');
        }
        register(): number {
          throw new Error('must not register');
        }
        release(): boolean {
          throw new Error('must not release');
        }
        close() {
          closed = true;
        }
      };
      let memory: ReturnType<typeof createMemoryTransport> | undefined;
      try {
        memory = createMemoryTransport();
        expect(memory.fd()).toBe(platform === 'windows' ? undefined : 5);
        memory.configure(MEMORY_CONFIG);
        expect(JSON.parse(JSON.stringify(memory.descriptor()))).toEqual(
          platform === 'windows'
            ? { version: MEMORY_CONFIG.version, handle: '42', processId: 123 }
            : { version: MEMORY_CONFIG.version, fd: 5 },
        );
      } finally {
        memory?.close();
        binding.MemoryArena = MemoryArena;
      }
      expect(closed).toBe(true);
    },
  );

  test.each(
    [
      undefined,
      [],
      { ...MEMORY_CONFIG, version: 0 },
      { ...MEMORY_CONFIG, slotCount: -1 },
      { ...MEMORY_CONFIG, slotSize: 1.5 },
      { ...MEMORY_CONFIG, headerSize: '512' },
      { ...MEMORY_CONFIG, publicationStride: NaN },
      { ...MEMORY_CONFIG, slotCount: 2 ** 32 + 3 },
      { ...MEMORY_CONFIG, slotSize: 2 ** 32 + 4096 },
    ].map((config) => ({ config })),
  )(
    'rejects an invalid runtime configuration and permanently closes storage: %j',
    ({ config }) => {
      const memory = createMemoryTransport();
      expect(() => memory.configure(config)).toThrow(
        'invalid shared memory configuration',
      );
      expect(memory.fd()).toBeUndefined();
      expect(memory.configuration()).toBeUndefined();
      expect(() => memory.configure(MEMORY_CONFIG)).toThrow('closed');
      memory.close();
    },
  );

  test('accepts the final slot in the peer-provided layout', async () => {
    const pair = receiver({ memoryTransport: createMemoryTransport() });
    const batch = { ...emptyBatch(), slot: MEMORY_CONFIG.slotCount - 1 };
    pair.b.setInboundHandler((msg) => {
      expect(msg.attachments).toEqual([
        { offset: 0, length: 0, lease: expect.any(Number) },
      ]);
      return {};
    });
    try {
      const response = await pair.request({
        attachments: [{ range: { offset: 0, length: 0 } }],
        transport: { batches: [batch] },
      });
      expect(response.kind).toBe('response');
      expect(response.transport).toEqual({ released: [batch] });
    } finally {
      pair.cleanup();
    }
  });

  test.each([
    'success',
    'error',
    'serialization',
    'message-getter',
    'toString',
  ] as const)(
    'revokes native reads before replying and acknowledges release on %s',
    async (outcome) => {
      const pair = receiver({ memoryTransport: createMemoryTransport() });
      const batch = emptyBatch();
      let capability: SharedBytes | undefined;
      pair.b.setInboundHandler((msg) => {
        const attachment = msg.attachments?.[0];
        if (typeof attachment !== 'object' || attachment instanceof Uint8Array)
          throw new Error('missing capability');
        capability = attachment;
        expect(readAttachmentBytes(capability)).toEqual(Buffer.alloc(0));
        expect(msg.data).toEqual({ arbitrary: true });
        expect('transport' in msg).toBe(false);
        if (outcome === 'error') throw new Error('dispatch failed');
        if (outcome === 'message-getter') {
          throw Object.defineProperty(new Error(), 'message', {
            get() {
              throw new Error('bad getter');
            },
          });
        }
        if (outcome === 'toString') {
          throw {
            toString() {
              throw new Error('bad conversion');
            },
          };
        }
        return outcome === 'serialization' ? { value: 1n } : { ok: true };
      });
      try {
        const wire = {
          data: { arbitrary: true },
          attachments: [{ range: { offset: 0, length: 0 } }],
          transport: { batches: [batch] },
        };
        const response = await pair.request(wire);
        expect(response.kind).toBe(
          outcome === 'success' ? 'response' : 'error',
        );
        expect(response.transport).toEqual({ released: [batch] });
        if (outcome === 'message-getter' || outcome === 'toString') {
          expect(response.data).toEqual({ message: 'request failed' });
        }
        expect(() => readAttachmentBytes(capability!)).toThrow('expired');
        const replay = await pair.request(wire);
        expect(replay.kind).toBe('error');
        expect(replay.transport).toBeUndefined();
        expect(replay.data).toEqual({
          message: expect.stringContaining('expired'),
        });
      } finally {
        pair.cleanup();
      }
    },
  );

  test('does not acknowledge a native release that cannot prove reuse', async () => {
    const memory = createMemoryTransport();
    const release = memory.release;
    memory.release = (lease) => {
      release(lease);
      return false;
    };
    const pair = receiver({ memoryTransport: memory });
    pair.b.setInboundHandler(() => ({ ok: true }));
    try {
      const response = await pair.request({
        attachments: [{ range: { offset: 0, length: 0 } }],
        transport: { batches: [emptyBatch()] },
      });
      expect(response.kind).toBe('response');
      expect(response.transport).toBeUndefined();
    } finally {
      pair.cleanup();
    }
  });

  test.each(['response', 'handler error', 'serialization error'] as const)(
    'returns %s before deferred reclamation and acknowledges it exactly once',
    async (outcome) => {
      const memory = createMemoryTransport();
      const release = memory.release;
      let readerFinished = false;
      let attempts = 0;
      // The native worker suite exercises real overlapping reads. Here control
      // the completion boundary to check result/ACK ordering independently.
      memory.release = (lease) => {
        attempts++;
        return readerFinished && release(lease);
      };
      const pair = receiver({ memoryTransport: memory });
      pair.b.setInboundHandler(() => {
        if (outcome === 'handler error') throw new Error('task failed');
        if (outcome === 'serialization error') return 1n;
        return { ok: true };
      });
      const batches = [emptyBatch(), { ...emptyBatch(), slot: 2 }];
      try {
        const response = await pair.request({
          attachments: [{ range: { offset: 0, length: 0 } }],
          transport: { batches },
        });
        expect(response.kind).toBe(
          outcome === 'response' ? 'response' : 'error',
        );
        expect(response.transport).toBeUndefined();
        expect(attempts).toBe(1);
        const acknowledgement = once(pair.streams.bToA, 'data');
        readerFinished = true;
        const frame = decodeFrame((await acknowledgement)[0])!.msg;
        expect(frame).toEqual({
          kind: 'transportRelease',
          id: response.id,
          transport: { released: batches },
        });
        const completedAttempts = attempts;
        const extraFrames: Buffer[] = [];
        pair.streams.bToA.on('data', (frame: Buffer) =>
          extraFrames.push(frame),
        );
        await new Promise((resolve) => setTimeout(resolve, 75));
        expect(attempts).toBe(completedAttempts);
        expect(extraFrames).toEqual([]);
      } finally {
        pair.cleanup();
      }
    },
  );

  test.each(['close', 'native error', 'write error'] as const)(
    'stops deferred release work after %s',
    async (ending) => {
      const memory = createMemoryTransport();
      const release = memory.release;
      let finish = false;
      let attempts = 0;
      memory.release = (lease) => {
        attempts++;
        if (!finish) return false;
        if (ending === 'native error') throw new Error('release failed');
        return release(lease);
      };
      const pair = receiver({ memoryTransport: memory });
      pair.b.setInboundHandler(() => ({ ok: true }));
      try {
        await pair.request({
          attachments: [{ range: { offset: 0, length: 0 } }],
          transport: { batches: [emptyBatch()] },
        });
        if (ending === 'close') pair.b.close();
        if (ending === 'write error') {
          pair.streams.bToA.write = () => {
            throw new Error('release write failed');
          };
        }
        finish = true;
        const reason = await pair.b.done;
        expect(reason.message).toContain(
          ending === 'close' ? 'closed' : 'failed',
        );
        const completedAttempts = attempts;
        await new Promise((resolve) => setTimeout(resolve, 75));
        expect(attempts).toBe(completedAttempts);
        expect(memory.configuration()).toBeUndefined();
      } finally {
        pair.cleanup();
      }
    },
  );

  test('closes on a duplicate request ID instead of losing a pending release', async () => {
    const memory = createMemoryTransport();
    memory.release = () => false;
    const pair = receiver({ memoryTransport: memory });
    pair.b.setInboundHandler(() => undefined);
    const attachments = [{ range: { offset: 0, length: 0 } }];
    try {
      const response = await pair.request({
        attachments,
        transport: { batches: [emptyBatch()] },
      });
      pair.streams.aToB.write(
        encodeFrame({
          kind: 'testAttachments',
          id: response.id,
          attachments,
          transport: { batches: [{ ...emptyBatch(), slot: 1 }] },
        }),
      );
      expect((await pair.b.done).message).toContain('duplicate pending');
      expect(memory.configuration()).toBeUndefined();
    } finally {
      pair.cleanup();
    }
  });

  test('rejects reverse storage ACKs without invoking application handlers', async () => {
    const pair = receiver();
    let calls = 0;
    pair.b.setInboundHandler(() => calls++);
    pair.b.registerNotification('transportRelease', () => calls++);
    try {
      pair.streams.aToB.write(encodeFrame({ kind: 'transportRelease', id: 1 }));
      expect((await pair.b.done).message).toContain(
        'unexpected shared memory release',
      );
      expect(calls).toBe(0);
    } finally {
      pair.cleanup();
    }
  });

  test('registers and acknowledges an ordered batch set as one lease', async () => {
    const memory = createMemoryTransport();
    const register = memory.register;
    const release = memory.release;
    const registrations: MemoryBatch[][] = [];
    const releases: number[] = [];
    memory.register = (batches) => {
      registrations.push(batches);
      return register(batches);
    };
    memory.release = (lease) => {
      releases.push(lease);
      return release(lease);
    };
    const pair = receiver({ memoryTransport: memory });
    const batches = [{ ...emptyBatch(), slot: 2 }, emptyBatch()];
    let capability: SharedBytes | undefined;
    pair.b.setInboundHandler((msg) => {
      const first = msg.attachments![0];
      const second = msg.attachments![1];
      if (typeof first !== 'object' || first instanceof Uint8Array)
        throw new Error('missing capability');
      capability = first;
      expect(second).toEqual(first);
      expect(readAttachmentBytes(first)).toEqual(Buffer.alloc(0));
      return { ok: true };
    });
    try {
      const response = await pair.request({
        attachments: [
          { range: { offset: 0, length: 0 } },
          { range: { offset: 0, length: 0 } },
        ],
        transport: { batches },
      });
      expect(response.kind).toBe('response');
      expect(registrations).toEqual([batches]);
      expect(releases).toEqual([capability!.lease]);
      expect(response.transport).toEqual({ released: batches });
      expect(() => readAttachmentBytes(capability!)).toThrow('expired');
    } finally {
      pair.cleanup();
    }
  });

  test('does not partially register or acknowledge when a later batch fails', async () => {
    const pair = receiver({ memoryTransport: createMemoryTransport() });
    let calls = 0;
    pair.b.setInboundHandler(() => {
      calls++;
      return {};
    });
    const attachments = [{ range: { offset: 0, length: 0 } }];
    const retired = { ...emptyBatch(), slot: 1 };
    try {
      expect(
        (await pair.request({ attachments, transport: { batches: [retired] } }))
          .kind,
      ).toBe('response');
      const failed = await pair.request({
        attachments,
        transport: { batches: [emptyBatch(), retired] },
      });
      expect(failed.kind).toBe('error');
      expect(failed.transport).toBeUndefined();
      expect(calls).toBe(1);
      // The valid first slot was not consumed by the rejected multi-slot lease.
      const valid = await pair.request({
        attachments,
        transport: { batches: [emptyBatch()] },
      });
      expect(valid.kind).toBe('response');
      expect(valid.transport).toEqual({ released: [emptyBatch()] });
      expect(calls).toBe(2);
    } finally {
      pair.cleanup();
    }
  });

  test.each(
    [
      [],
      [emptyBatch(), emptyBatch()],
      [emptyBatch(), { ...emptyBatch(), slot: MEMORY_CONFIG.slotCount }],
    ].map((batches) => ({ batches })),
  )(
    'rejects a malformed batch set before invoking native registration: %j',
    async ({ batches }) => {
      const memory = createMemoryTransport();
      let registrations = 0;
      const register = memory.register;
      memory.register = (values) => {
        registrations++;
        return register(values);
      };
      const pair = receiver({ memoryTransport: memory });
      pair.b.setInboundHandler(() => {
        throw new Error('must not dispatch');
      });
      try {
        const result = await pair.request({
          attachments: [{ range: { offset: 0, length: 0 } }],
          transport: { batches },
        });
        expect(result.kind).toBe('error');
        expect(result.transport).toBeUndefined();
        expect(registrations).toBe(0);
      } finally {
        pair.cleanup();
      }
    },
  );

  test('keeps mixed inline/shared attachments ordered and out of application data', async () => {
    const pair = receiver({ memoryTransport: createMemoryTransport() });
    pair.b.setInboundHandler((msg) => {
      expect(msg.attachments).toEqual([
        'inline',
        { offset: 0, length: 0, lease: expect.any(Number) },
        Buffer.from([0, 255, 128]),
        '',
      ]);
      expect(msg.data).toEqual({ payload: 'unchanged' });
      return {};
    });
    try {
      const response = await pair.request({
        data: { payload: 'unchanged' },
        attachments: [
          { text: 'inline' },
          { range: { offset: 0, length: 0 } },
          { bytes: 'AP+A' },
          { text: '' },
        ],
        transport: { batches: [emptyBatch()] },
      });
      expect(response.transport).toEqual({ released: [emptyBatch()] });
    } finally {
      pair.cleanup();
    }
  });

  test('revokes a valid batch even when no application handler is installed', async () => {
    const pair = receiver({ memoryTransport: createMemoryTransport() });
    try {
      const response = await pair.request({
        attachments: [{ range: { offset: 0, length: 0 } }],
        transport: { batches: [emptyBatch()] },
      });
      expect(response.kind).toBe('error');
      expect(response.transport).toEqual({ released: [emptyBatch()] });
    } finally {
      pair.cleanup();
    }
  });

  test('rejects shared attachments when native storage is unavailable', async () => {
    const pair = receiver();
    pair.b.setInboundHandler(() => {
      throw new Error('must not dispatch');
    });
    try {
      const response = await pair.request({
        attachments: [{ range: { offset: 0, length: 0 } }],
        transport: { batches: [emptyBatch()] },
      });
      expect(response.kind).toBe('error');
      expect(response.data).toEqual({
        message: 'shared IPC attachments are unavailable',
      });
      expect(response.transport).toBeUndefined();
    } finally {
      pair.cleanup();
    }
  });

  test('close revokes a pending handler capability and suppresses a late acknowledgement', async () => {
    const pair = receiver({ memoryTransport: createMemoryTransport() });
    let entered!: () => void;
    let finish!: () => void;
    const started = new Promise<void>((resolve) => {
      entered = resolve;
    });
    const pending = new Promise<void>((resolve) => {
      finish = resolve;
    });
    let capability: SharedBytes | undefined;
    let replies = 0;
    pair.streams.bToA.on('data', () => {
      replies++;
    });
    pair.b.setInboundHandler(async (msg) => {
      const value = msg.attachments?.[0];
      if (typeof value !== 'object' || value instanceof Uint8Array)
        throw new Error('missing capability');
      capability = value;
      entered();
      await pending;
      return { ok: true };
    });
    try {
      pair.streams.aToB.write(
        encodeFrame({
          kind: 'anything',
          id: 1,
          attachments: [{ range: { offset: 0, length: 0 } }],
          transport: { batches: [emptyBatch()] },
        }),
      );
      await started;
      pair.b.close();
      expect(() => readAttachmentBytes(capability!)).toThrow('expired');
      finish();
      await new Promise<void>((resolve) => setImmediate(resolve));
      expect(replies).toBe(0);
    } finally {
      finish();
      pair.cleanup();
    }
  });

  test.each([
    { slot: -1 },
    { slot: MEMORY_CONFIG.slotCount },
    { slot: 0.5 },
    { generation: 0 },
    { generation: MAX_GENERATION + 1 },
    { length: NaN },
    { length: Infinity },
    { length: MEMORY_CONFIG.slotSize + 1 },
  ])(
    'rejects malformed batches before creating a capability: %j',
    async (change) => {
      const pair = receiver({ memoryTransport: createMemoryTransport() });
      let calls = 0;
      pair.b.setInboundHandler(() => {
        calls++;
        return {};
      });
      try {
        const response = await pair.request({
          attachments: [{ range: { offset: 0, length: 0 } }],
          transport: { batches: [{ ...emptyBatch(), ...change }] },
        });
        expect(response.kind).toBe('error');
        expect(response.data).toEqual({
          message: 'invalid IPC attachment batch',
        });
        expect(response.transport).toBeUndefined();
        expect(calls).toBe(0);
        const valid = await pair.request({
          attachments: [{ range: { offset: 0, length: 0 } }],
          transport: { batches: [emptyBatch()] },
        });
        expect(valid.kind).toBe('response');
        expect(calls).toBe(1);
      } finally {
        pair.cleanup();
      }
    },
  );

  test.each([
    { offset: -1, length: 0 },
    { offset: 1, length: 0 },
    { offset: 0, length: 1 },
    { offset: 0, length: 0.5 },
    { offset: 2 ** 32, length: 2 ** 32 },
  ])(
    'rejects malformed ranges before creating a capability: %j',
    async (range) => {
      const pair = receiver({ memoryTransport: createMemoryTransport() });
      let calls = 0;
      pair.b.setInboundHandler(() => {
        calls++;
        return {};
      });
      try {
        const response = await pair.request({
          attachments: [{ range }],
          transport: { batches: [emptyBatch()] },
        });
        expect(response.kind).toBe('error');
        expect(response.transport).toBeUndefined();
        expect(calls).toBe(0);
        const valid = await pair.request({
          attachments: [{ range: { offset: 0, length: 0 } }],
          transport: { batches: [emptyBatch()] },
        });
        expect(valid.kind).toBe('response');
      } finally {
        pair.cleanup();
      }
    },
  );

  test.each([
    { attachments: [{ text: '', bytes: '' }] },
    {
      attachments: [{ bytes: '', range: { offset: 0, length: 0 } }],
      transport: { batches: [emptyBatch()] },
    },
    {
      attachments: [{ text: '', bytes: '', range: { offset: 0, length: 0 } }],
      transport: { batches: [emptyBatch()] },
    },
    { attachments: [{ range: { offset: 0, length: 0 } }] },
    {
      attachments: [{ text: '', range: { offset: 0, length: 0 } }],
      transport: { batches: [emptyBatch()] },
    },
    { attachments: [{ text: '' }], transport: { batches: [emptyBatch()] } },
    {
      attachments: [{ range: { offset: 0, length: 0 } }],
      transport: { batches: [{ ...emptyBatch(), length: 1 }] },
    },
    {
      attachments: [
        { range: { offset: 0, length: 1 } },
        { range: { offset: 0, length: 1 } },
      ],
      transport: { batches: [{ ...emptyBatch(), length: 2 }] },
    },
  ])(
    'rejects missing, conflicting or incomplete attachment metadata: %j',
    async (message) => {
      const pair = receiver({ memoryTransport: createMemoryTransport() });
      pair.b.setInboundHandler(() => {
        throw new Error('must not dispatch');
      });
      try {
        const response = await pair.request(message);
        expect(response.kind).toBe('error');
        expect(response.data).not.toEqual({ message: 'must not dispatch' });
        expect(response.transport).toBeUndefined();
      } finally {
        pair.cleanup();
      }
    },
  );
});

describe('real Go/Node arbitrary byte attachments', () => {
  const repoRoot = fileURLToPath(new URL('../../../', import.meta.url));
  let fixtureRoot: string;
  let binary: string;
  beforeAll(async () => {
    fixtureRoot = fs.mkdtempSync(path.join(os.tmpdir(), 'rslint-binary-peer-'));
    binary = path.join(
      fixtureRoot,
      process.platform === 'win32' ? 'peer.exe' : 'peer',
    );
    await promisify(execFile)(
      'go',
      ['build', '-o', binary, './internal/ipc/testdata/binary-peer'],
      { cwd: repoRoot },
    );
  }, 120_000);
  afterAll(() => {
    if (fixtureRoot) fs.rmSync(fixtureRoot, { recursive: true, force: true });
  });

  test.each([true, false])(
    'preserves binary/text/empty attachments, shared memory: %s',
    async (sharedMemory) => {
      type Payload = { round: number; hashes: string[]; lengths: number[] };
      const { child, client } = spawnIpcProcess({
        binPath: binary,
        goArgs: [],
        sharedMemory,
      });
      const childClosed = new Promise<void>((resolve) =>
        child.once('close', () => resolve()),
      );
      child.on('error', (error) => client.close(error));
      child.once('exit', () => client.close());
      const capabilities: SharedBytes[] = [];
      const owned: Buffer[] = [];
      const hashes: string[] = [];
      let calls = 0;
      const digest = (bytes: Buffer) =>
        createHash('sha256').update(bytes).digest('hex');
      client.setInboundHandler((message) => {
        expect(message.kind).toBe('binaryAttachments');
        const data = message.data as Payload;
        expect(data.round).toBe(calls++);
        expect(message.attachments).toHaveLength(4);
        const buffers = message.attachments!.map(readAttachmentBytes);
        expect(buffers.map(digest)).toEqual(data.hashes);
        expect(buffers.map((value) => value.length)).toEqual(data.lengths);
        expect(buffers[0].toString('utf8')).toBe(
          '\ufeffarbitrary text: café 😀\r\n\u0000',
        );
        expect(buffers[1].length).toBe(0);
        expect(buffers[2].length).toBe(0);
        expect(buffers[3].length).toBeGreaterThan(16 * 1024 * 1024);
        const binaryAttachment = message.attachments![3];
        if (sharedMemory) {
          if (
            typeof binaryAttachment !== 'object' ||
            binaryAttachment instanceof Uint8Array
          )
            throw new Error('real Go peer did not publish shared bytes');
          capabilities.push(binaryAttachment);
          // Changing a returned Buffer cannot mutate Go's published snapshot.
          const changed = readAttachmentBytes(binaryAttachment);
          changed[0] ^= 255;
          expect(digest(readAttachmentBytes(binaryAttachment))).toBe(
            data.hashes[3],
          );
          expect(digest(changed)).not.toBe(data.hashes[3]);
        } else {
          expect(Buffer.isBuffer(binaryAttachment)).toBe(true);
          expect(typeof message.attachments![0]).toBe('string');
        }
        owned.push(buffers[3]);
        hashes.push(data.hashes[3]);
        return {
          round: data.round,
          hashes: buffers.map(digest),
          lengths: buffers.map((value) => value.length),
        };
      });
      client.start();
      try {
        for (const round of [0, 1]) {
          const response = await client.sendRequest<unknown, Payload>(
            'startBinary',
            { round },
          );
          expect(response.data!.round).toBe(round);
          expect(calls).toBe(round + 1); // One application call, even across slots.
          for (const capability of capabilities) {
            expect(() => readAttachmentBytes(capability)).toThrow('expired');
          }
          // Revocation and the next Go publication do not affect owned copies.
          expect(owned.map(digest)).toEqual(hashes);
        }
        expect(hashes[0]).not.toBe(hashes[1]);
      } finally {
        client.close();
        child.kill();
        await childClosed;
      }
      expect(owned.map(digest)).toEqual(hashes);
    },
    120_000,
  );
  test('a real Go peer reuses slots only after Node sends the deferred ACK', async () => {
    const memory = createMemoryTransport();
    const register = memory.register;
    const release = memory.release;
    const publications: MemoryBatch[][] = [];
    let blockedLease: number | undefined;
    let readerFinished = false;
    let reclaimed!: () => void;
    const reclamation = new Promise<void>((resolve) => {
      reclaimed = resolve;
    });
    memory.register = (batches) => {
      publications.push(batches);
      const lease = register(batches);
      blockedLease ??= lease;
      return lease;
    };
    memory.release = (lease) => {
      if (lease === blockedLease && !readerFinished) return false;
      const result = release(lease);
      if (result && lease === blockedLease) queueMicrotask(reclaimed);
      return result;
    };
    const { child, client } = spawnIpcProcess({
      binPath: binary,
      goArgs: [],
      createMemoryTransport: () => memory,
    });
    const childClosed = new Promise<void>((resolve) =>
      child.once('close', () => resolve()),
    );
    child.on('error', (error) => client.close(error));
    child.once('exit', () => client.close());
    client.setInboundHandler((message) => {
      expect(message.kind).toBe('binaryAttachments');
      const data = message.data as {
        round: number;
        hashes: string[];
        lengths: number[];
      };
      const buffers = message.attachments!.map(readAttachmentBytes);
      const hashes = buffers.map((bytes) =>
        createHash('sha256').update(bytes).digest('hex'),
      );
      expect(hashes).toEqual(data.hashes);
      return {
        round: data.round,
        hashes,
        lengths: buffers.map((bytes) => bytes.length),
      };
    });
    client.start();
    try {
      // This test gates native completion; the packaged worker test proves
      // that an actual synchronous native reader creates the same boundary.
      await client.sendRequest('startBinary', { round: 0 });
      await client.sendRequest('startBinary', { round: 1 });
      expect(publications).toHaveLength(2);
      const occupied = new Set(publications[0].map((batch) => batch.slot));
      expect(publications[1].every((batch) => !occupied.has(batch.slot))).toBe(
        true,
      );
      readerFinished = true;
      await reclamation;
      await client.sendRequest('startBinary', { round: 2 });
      expect(publications).toHaveLength(3);
      expect(publications[2].map((batch) => batch.slot)).toEqual([...occupied]);
      expect(
        publications[2].every(
          (batch, index) =>
            batch.generation > publications[0][index].generation,
        ),
      ).toBe(true);
    } finally {
      client.close();
      child.kill();
      await childClosed;
    }
  }, 120_000);
});

describe('encode/decode round-trip', () => {
  test('rejects an oversized outbound frame before writing and keeps the client usable', async () => {
    const pair = pairClients();
    const byteLength = Buffer.byteLength;
    let writes = 0;
    pair.streams.aToB.on('data', () => {
      writes++;
    });
    pair.b.setInboundHandler(() => ({ ok: true }));
    pair.a.start();
    pair.b.start();
    let rejected: Promise<unknown>;
    try {
      // Stub only the size measurement, avoiding a >256 MiB test allocation.
      // The real serializer, pending map, frame writer and recovery all run.
      Buffer.byteLength = (value, encoding) =>
        typeof value === 'string' && value.includes('oversizeProbe')
          ? 256 * 1024 * 1024 + 1
          : byteLength(value, encoding);
      rejected = pair.a.sendRequest('oversizeProbe', {});
    } finally {
      Buffer.byteLength = byteLength;
    }
    try {
      await expect(rejected!).rejects.toThrow('exceeds cap');
      expect(writes).toBe(0);
      expect(
        (pair.a as unknown as { pending: Map<number, unknown> }).pending.size,
      ).toBe(0);
      expect(pair.a.isClosed).toBe(false);
      expect((await pair.a.sendRequest('normal', {})).data).toEqual({
        ok: true,
      });
    } finally {
      pair.cleanup();
    }
  });
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

  test.each([
    ['boom', 'boom'],
    ['', 'request failed'],
    ['peer error: from config', 'peer error: from config'],
  ])('handler error %j → peer receives %j', async (message, expected) => {
    const { a, b, cleanup } = pairClients();
    try {
      b.setInboundHandler(() => {
        throw new Error(message);
      });
      a.start();
      b.start();
      await expect(a.sendRequest('lint', {})).rejects.toMatchObject({
        message: expected,
      });
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
    await expect(pending).rejects.toThrow(/input read failed.*exceeds cap/);

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

describe('IpcClient terminal cleanup', () => {
  test.each([
    null,
    [],
    'text',
    true,
    1,
    {},
    { kind: 'log' },
    { kind: 42, id: 0 },
    { kind: 'log', id: null },
    { kind: 'log', id: -1 },
    { kind: 'log', id: 0.5 },
    { kind: 'log', id: '0' },
    { kind: 'log', id: Number.MAX_SAFE_INTEGER + 1 },
  ])('closes on an invalid envelope before dispatch: %j', async (envelope) => {
    const input = new PassThrough();
    const output = new PassThrough();
    const client = new IpcClient(input, output);
    let dispatched = 0;
    client.setInboundHandler(() => {
      dispatched++;
    });
    client.registerNotification('log', () => {
      dispatched++;
    });
    client.start();
    const pending = Promise.allSettled([client.sendRequest('lint', {})]);
    try {
      const frame = encodeFrame(envelope as IpcMessage);
      const response = encodeFrame({ kind: 'response', id: 1, data: {} });
      // Invalid header data must neither escape the data listener nor allow
      // later frames in the same chunk to resolve a request on this session.
      expect(() => input.write(Buffer.concat([frame, response]))).not.toThrow();
      expect(client.isClosed).toBe(true);
      await client.done;
      expect(dispatched).toBe(0);
      const [result] = await pending;
      expect(result.status).toBe('rejected');
      if (result.status === 'rejected') {
        expect(result.reason.message).toContain('invalid message envelope');
      }
    } finally {
      client.close();
      input.destroy();
      output.destroy();
    }
  });

  test('preserves recovery after malformed JSON with an intact frame boundary', async () => {
    const input = new PassThrough();
    const output = new PassThrough();
    const client = new IpcClient(input, output);
    client.start();
    try {
      const pending = client.sendRequest('lint', {});
      const malformed = Buffer.from([1, 0, 0, 0, 0x7b]); // One-byte JSON body: {
      input.write(
        Buffer.concat([
          malformed,
          encodeFrame({ kind: 'response', id: 1, data: { recovered: true } }),
        ]),
      );
      expect((await pending).data).toEqual({ recovered: true });
      expect(client.isClosed).toBe(false);
    } finally {
      client.close();
      input.destroy();
      output.destroy();
    }
  });

  test.each(['notification', 'response', 'request'] as const)(
    'suppresses a %s when serialization closes the client',
    async (kind) => {
      const input = new PassThrough();
      const output = new PassThrough();
      const client = new IpcClient(input, output);
      const frames: Buffer[] = [];
      output.on('data', (chunk: Buffer) => frames.push(chunk));
      client.start();
      try {
        const data = {
          toJSON() {
            client.close();
            return {};
          },
        };
        if (kind === 'notification') client.sendNotification('log', data);
        else if (kind === 'response') client.sendResponse(1, data);
        else {
          const pending = client.sendRequest('init', data);
          await expect(pending).rejects.toBe(await client.done);
        }
        expect(client.isClosed).toBe(true);
        expect(frames).toEqual([]);
      } finally {
        client.close();
        input.destroy();
        output.destroy();
      }
    },
  );

  const endings = [
    'explicit close',
    'input end',
    'input close',
    'input error',
    'output error',
    'output close',
    'output finish',
    'write throws',
  ] as const;

  test.each(
    endings.flatMap((ending) => [
      { ending, closeThrows: false },
      { ending, closeThrows: true },
    ]),
  )(
    'cleans the session once and preserves its termination reason: %j',
    async ({ ending, closeThrows }) => {
      const input = new PassThrough();
      const output = new PassThrough();
      const memory = createMemoryTransport();
      memory.transfer = () => ({ version: 1 });
      const client = new IpcClient(input, output, {
        createMemoryTransport: () => memory,
      });
      let completed = 0;
      const completedSession = client.done.then((reason) => {
        completed++;
        // Completion observes fully detached streams and released storage.
        expect(closeCalls).toBe(1);
        expect(input.listenerCount('data')).toBe(0);
        expect(memory.fd()).toBeUndefined();
        return reason;
      });
      const reason = new Error('test termination');
      const closeMemory = memory.close;
      const write = output.write;
      let closeCalls = 0;
      memory.close = () => {
        closeCalls++;
        // Teardown is already sealed if native cleanup reenters the client.
        client.close(new Error('reentrant close'));
        closeMemory();
        if (closeThrows) throw new Error('native cleanup failed');
      };
      let requests = 0;
      let sent!: () => void;
      const requestsSent = new Promise<void>((resolve) => {
        sent = resolve;
      });
      output.on('data', (chunk: Buffer) => {
        const frame = decodeFrame(chunk)!.msg;
        if (frame.kind !== 'response' && ++requests === 2) {
          sent();
        }
      });
      client.setInboundHandler(() => undefined);
      client.registerNotification('log', () => undefined);
      client.start();
      input.write(
        encodeFrame({
          kind: 'transportPrepare',
          id: 1000,
          data: { configuration: MEMORY_CONFIG },
        }),
      );
      input.write(encodeFrame({ kind: 'transportCommit', id: 0 }));
      try {
        const pending = Promise.allSettled([
          client.sendRequest('first', {}),
          client.sendRequest('second', {}),
        ]);
        await requestsSent;
        expect(completed).toBe(0);
        // An unfinished frame must stop retaining its buffer after termination.
        input.write(Buffer.from([0, 1]));

        let expected = 'test termination';
        if (ending === 'explicit close') {
          client.close(reason);
        } else if (ending === 'input end') {
          expected = 'IpcClient: peer closed input stream';
          const ended = once(input, 'end');
          input.end();
          await ended;
        } else if (ending === 'input close') {
          expected = 'IpcClient: peer closed input stream';
          const closed = once(input, 'close');
          input.destroy();
          await closed;
        } else if (ending === 'input error' || ending === 'output error') {
          expected =
            ending === 'input error'
              ? 'IpcClient: input read failed: test termination'
              : 'IpcClient: output write failed: test termination';
          const stream = ending === 'input error' ? input : output;
          const errored = once(stream, 'error');
          stream.destroy(reason);
          await errored;
        } else if (ending === 'output close' || ending === 'output finish') {
          expected = 'IpcClient: output stream closed before response received';
          const ended = once(
            output,
            ending === 'output close' ? 'close' : 'finish',
          );
          if (ending === 'output close') output.destroy();
          else output.end();
          await ended;
        } else {
          expected = 'IpcClient: output write failed: test termination';
          output.write = () => {
            throw reason;
          };
          client.sendNotification('log', {});
        }
        client.close(new Error('later close'));
        client.start();

        const closeReason = await completedSession;
        for (const result of await pending) {
          expect(result.status).toBe('rejected');
          if (result.status === 'rejected') {
            expect(result.reason.message).toBe(expected);
            expect(result.reason).toBe(closeReason);
            if (ending === 'explicit close') expect(result.reason).toBe(reason);
          }
        }
        expect(await client.done).toBe(closeReason);
        expect(completed).toBe(1);
        expect(client.isClosed).toBe(true);
        expect(closeCalls).toBe(1);
        expect(memory.fd()).toBeUndefined();
        for (const event of ['data', 'end', 'close', 'error']) {
          expect(input.listenerCount(event)).toBe(0);
        }
        for (const event of ['error', 'close', 'finish']) {
          expect(output.listenerCount(event)).toBe(0);
        }
        const state = client as unknown as {
          chunks: Buffer[];
          bufferedBytes: number;
          pending: Map<number, unknown>;
          inboundHandler: unknown;
          notificationHandlers: Map<string, unknown>;
        };
        expect(state.chunks).toEqual([]);
        expect(state.bufferedBytes).toBe(0);
        expect(state.pending.size).toBe(0);
        expect(state.inboundHandler).toBeNull();
        expect(state.notificationHandlers.size).toBe(0);
        await expect(client.sendRequest('late', {})).rejects.toThrow(/closed/);
      } finally {
        output.write = write;
        client.close();
        closeMemory();
        input.destroy();
        output.destroy();
      }
    },
  );
});
